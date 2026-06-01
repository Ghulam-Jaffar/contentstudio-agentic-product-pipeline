package main

import (
	"context"
	"errors"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	redisdb "github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/enrichment"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/fetcher"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/parser"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/quota"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/sink"
)

const (
	fetcherGroup    = "listening-fetcher-group"
	parserGroup     = "listening-parser-group"
	sinkGroup       = "listening-sink-group"
	enrichmentGroup = "listening-enrichment-group"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Listening Consumer")

	mongoClient, err := connectMongo(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(context.Background())

	clickhouseClient, err := clickhouse.NewClient(cfg.ClickHouse, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to ClickHouse")
	}
	defer clickhouseClient.Close()

	producer, err := kafka.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	redisClient := initRedis(cfg, log)
	defer redisClient.Close()

	data365Client := social.NewData365Client(cfg.Data365, log)
	if cfg.Data365.BaseURL == "" || cfg.Data365.AccessToken == "" {
		log.Warn().Msg("DATA365 config incomplete; fetch stage will fail until APP_DATA365_BASE_URL and APP_DATA365_ACCESS_TOKEN are set")
	}

	var batchAnalyzer enrichment.AIAnalyzer
	if cfg.AIAgents.BaseURL != "" {
		batchAnalyzer = enrichment.NewAgentAnalyzer(
			cfg.AIAgents.BaseURL,
			cfg.AIAgents.APIKey,
			cfg.AIAgents.Timeout,
			log.Logger,
		)
	} else {
		log.Warn().Msg("AI_AGENTS_BASE_URL not set; enrichment stage will be disabled")
	}

	listeningRepo := mongodb.NewListeningRepository(mongoClient.Database(cfg.Mongo.Database), log)
	workspaceRepo := mongodb.NewListeningWorkspaceRepository(mongoClient.Database(cfg.Mongo.Database), log)
	listeningReader := clickhouse.NewListeningReadRepository(clickhouseClient, log.Logger)
	listeningWriter := clickhouse.NewListeningWriteRepository(clickhouseClient, log.Logger)
	lock := redisdb.NewDistributedLock(redisClient, log.Logger)

	quotaService := quota.NewQuotaService(listeningRepo, workspaceRepo, log)
	distributedQuota := quota.NewDistributedQuotaTracker(redisClient, log)

	// First-batch Pusher client drives the feed-page "collecting your first
	// mentions" animation. The same client wires both terminal events:
	// success (sink) when a mention lands, and empty (fetcher) when an
	// initial sync completes without producing any items. When Pusher
	// config is incomplete both notifiers are omitted and the frontend
	// falls back to its 5min timeout.
	pusherClient := notification.NewPusherClient(cfg.Pusher, log.Logger)

	fetcherService := fetcher.NewFetcherService(
		data365Client,
		producer,
		lock,
		log,
		cfg.Listening.LockTTLMin,
		cfg.Listening.BatchSize,        // incremental cycles
		cfg.Listening.BatchSizeInitial, // first crawl per topic
		quotaService,
	).WithProgressTracker(listeningRepo).WithTopicSyncMarker(listeningRepo).WithSuperAdminResolver(workspaceRepo).WithEmptyBatchNotifier(pusherClient).WithQuotaTopicCounter(listeningRepo)
	if cfg.Redis.Addr != "" {
		fetcherService = fetcherService.WithDistributedQuota(distributedQuota)
	}
	parserService := parser.NewParserService(producer, redisClient, listeningRepo, log, cfg.Listening.DedupTTLHours)

	sinkService := sink.NewSinkService(listeningWriter, listeningRepo, workspaceRepo, producer, log, cfg.Listening.MaxRetries).
		WithFirstBatchNotifier(pusherClient)
	enrichmentService := enrichment.NewEnrichmentService(batchAnalyzer, listeningWriter, listeningRepo, log).
		WithBackfillSource(
			listeningReader,
			time.Duration(cfg.Listening.EnrichmentBackfillIntervalSec)*time.Second,
			time.Duration(cfg.Listening.EnrichmentBackfillLookbackHours)*time.Hour,
		)

	fetchConsumer, err := kafka.NewConsumer(cfg.Kafka, fetcherGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create fetch consumer")
	}
	defer fetchConsumer.Close()

	parserConsumer, err := kafka.NewConsumer(cfg.Kafka, parserGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create parser consumer")
	}
	defer parserConsumer.Close()

	sinkConsumer, err := kafka.NewConsumer(cfg.Kafka, sinkGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create sink consumer")
	}
	defer sinkConsumer.Close()

	var enrichmentConsumer kafka.Consumer
	if batchAnalyzer != nil {
		enrichmentConsumer, err = kafka.NewConsumer(cfg.Kafka, enrichmentGroup, log.Logger)
		if err != nil {
			log.Fatal().Err(err).Msg("Failed to create enrichment consumer")
		}
		defer enrichmentConsumer.Close()
	}

	runCtx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()

	var wg sync.WaitGroup

	startStage(&wg, runCtx, log, "fetcher", fetchConsumer, kafkamodels.TopicListeningWork, fetcherService.HandleWorkOrder)
	startStage(&wg, runCtx, log, "parser", parserConsumer, kafkamodels.TopicListeningRaw, parserService.HandleRawPayload)

	// Sink and enrichment both consume listening-parsed in parallel: sink is the
	// fast path (immediate insert with empty tags/sentiment so the UI sees the
	// mention), enrichment is the slow path (re-inserts the row with AI results).
	// ReplacingMergeTree(updated_at) sorted on (topic_id, platform, mention_id)
	// collapses the two writes into one row on merge, with the enriched copy
	// winning because its updated_at is strictly later. See sink package doc.
	startStage(&wg, runCtx, log, "sink", sinkConsumer, kafkamodels.TopicListeningParsed, sinkService.HandleParsedMention)
	if batchAnalyzer != nil {
		startStage(
			&wg,
			runCtx,
			log,
			"enrichment",
			enrichmentConsumer,
			kafkamodels.TopicListeningParsed,
			enrichmentService.HandleParsedMention,
		)

		wg.Add(1)
		go func() {
			defer wg.Done()
			enrichmentService.StartFlushLoop(runCtx)
		}()

		wg.Add(1)
		go func() {
			defer wg.Done()
			enrichmentService.StartBackfillLoop(runCtx)
		}()
	}

	wg.Wait()
	log.Info().Msg("Listening Consumer stopped")
}

func connectMongo(cfg *config.Config) (*mongo.Client, error) {
	return mongo.Connect(context.Background(), mongoClientOptions(cfg))
}

func mongoClientOptions(cfg *config.Config) *options.ClientOptions {
	clientOpts := options.Client().ApplyURI(cfg.Mongo.URI)
	if cfg.Mongo.Username != "" {
		clientOpts.SetAuth(options.Credential{
			Username:   cfg.Mongo.Username,
			Password:   cfg.Mongo.Password,
			AuthSource: cfg.Mongo.Database,
		})
	}

	return clientOpts
}

// listeningRedis is the union of Redis methods required by the listening-consumer
// pipeline stages (distributed lock, quota tracker, dedup checker) plus Close.
type listeningRedis interface {
	Get(ctx context.Context, key string) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	CompareAndDelete(ctx context.Context, key, expected string) (bool, error)
	DecrByIfPositive(ctx context.Context, key string, amount int64) (int64, bool, error)
	DecrByIfExists(ctx context.Context, key string, amount int64) (int64, bool, error)
	IncrByIfExists(ctx context.Context, key string, amount int64) (int64, bool, error)
	Close() error
}

func initRedis(cfg *config.Config, log *logger.Logger) listeningRedis {
	if cfg.Redis.Addr == "" {
		log.Warn().Msg("APP_REDIS_ADDR not set; using in-memory mock Redis for locks/dedup")
		return &redisdb.MockRedisClient{}
	}

	client, err := redisdb.NewClient(cfg.Redis.Addr, cfg.Redis.Password, cfg.Redis.DB, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Redis connection failed; cannot start without distributed locking")
	}

	return client
}

func startStage(
	wg *sync.WaitGroup,
	ctx context.Context,
	log *logger.Logger,
	name string,
	consumer kafka.Consumer,
	topic string,
	handler kafka.MessageHandler,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()

		log.Info().Str("stage", name).Str("topic", topic).Msg("Starting listening stage")

		err := consumer.Consume(ctx, []string{topic}, handler)
		if err == nil {
			log.Info().Str("stage", name).Msg("Listening stage stopped")
			return
		}

		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			log.Info().Str("stage", name).Msg("Listening stage canceled")
			return
		}

		log.Error().Err(err).Str("stage", name).Msg("Listening stage failed")
	}()
}
