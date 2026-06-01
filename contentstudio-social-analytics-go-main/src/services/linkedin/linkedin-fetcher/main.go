package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/sync/semaphore"
)

const (
	maxPageWorkers    = 15
	maxProfileWorkers = 15
	workOrderChanSize = 500

	statsConcsPerWorker = 5
	mediaConcPerWorker  = 5
	geoConcPerWorker    = 3

	entityTypeOrganization = "organization"

	topicWorkOrderPageBatch    = "work-order-linkedin-page-batch"
	topicWorkOrderProfileBatch = "work-order-linkedin-profile-batch"

	pagePostsTopic        = "raw-linkedin-page-posts"
	pageInsightsTopic     = "raw-linkedin-page-insights"
	pageOrganizationTopic = "raw-linkedin-page-organization"
	profileInsightsTopic  = "raw-linkedin-profile-insights"

	pageConsumerGroup    = "linkedin-page-fetcher-group"
	profileConsumerGroup = "linkedin-profile-fetcher-group"

	maxConcurrentAccounts = 50

	perAccountConcurrency int64 = 1

	timestampUpdateChanSize = 1000

	idleTimeout       = 5 * time.Minute
	idleCheckInterval = 30 * time.Second
)

var (
	statsConc = semaphore.NewWeighted(int64(statsConcsPerWorker * maxConcurrentAccounts))
	mediaConc = semaphore.NewWeighted(int64(mediaConcPerWorker * maxConcurrentAccounts))
	geoConc   = semaphore.NewWeighted(int64(geoConcPerWorker * maxConcurrentAccounts))

	accountSemaphores sync.Map
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("max_concurrent_accounts", maxConcurrentAccounts).
		Msg("Starting LinkedIn Fetcher service")

	liClient := social.NewLinkedInClient()

	chClient, err := clickhouse.NewClient(cfg.ClickHouse, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create ClickHouse client")
	}

	geoResolver := social.NewGeoResolver(liClient, chClient)

	mongoClient, mongoRepo := initMongoDB(cfg, log)
	defer mongoClient.Disconnect(context.Background())

	pageConsumer, profileConsumer, producer := initKafkaClients(cfg, log)
	defer pageConsumer.Close()
	defer profileConsumer.Close()
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	timestampUpdateChan := make(chan TimestampUpdateRequest, timestampUpdateChanSize)
	var lastMessageTime int64 = time.Now().UnixNano()

	pageSem := semaphore.NewWeighted(maxConcurrentAccounts)
	profileSem := semaphore.NewWeighted(maxConcurrentAccounts)
	var dispatchWg sync.WaitGroup
	var totalProcessed, totalFailed int64

	pageProcessor := func(ctx context.Context, msg WorkOrderMessage) error {
		return processPageWorkOrder(ctx, msg, liClient, geoResolver, producer, mongoRepo, cfg.DecryptionKey, timestampUpdateChan)
	}
	profileProcessor := func(ctx context.Context, msg WorkOrderMessage) error {
		return processProfileWorkOrder(ctx, msg, liClient, producer, mongoRepo, cfg.DecryptionKey, timestampUpdateChan)
	}

	go consumeBatchWorkOrders(ctx, log, pageConsumer, pageConsumerGroup, []string{topicWorkOrderPageBatch}, pageSem, &dispatchWg, "page", &lastMessageTime, &totalProcessed, &totalFailed, pageProcessor)
	go consumeBatchWorkOrders(ctx, log, profileConsumer, profileConsumerGroup, []string{topicWorkOrderProfileBatch}, profileSem, &dispatchWg, "profile", &lastMessageTime, &totalProcessed, &totalFailed, profileProcessor)

	var wg sync.WaitGroup
	startTimestampUpdater(ctx, &wg, mongoRepo, timestampUpdateChan, log)

	log.Info().Msg("LinkedIn Fetcher service is running")

	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping service...")

	cancel()
	dispatchWg.Wait()

	log.Info().
		Int64("total_processed", atomic.LoadInt64(&totalProcessed)).
		Int64("total_failed", atomic.LoadInt64(&totalFailed)).
		Msg("LinkedIn Fetcher service stopped")

	close(timestampUpdateChan)
	wg.Wait()
}

func initMongoDB(cfg *config.Config, log *logger.Logger) (*mongo.Client, mongodb.UnifiedSocialRepository) {
	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	clientOpts := options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential)

	mongoClient, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}

	db := mongoClient.Database(cfg.Mongo.Database)
	repo := mongodb.NewUnifiedSocialRepository(db, log.Logger)

	log.Info().Msg("MongoDB connected for timestamp updates")
	return mongoClient, repo
}

func initKafkaClients(cfg *config.Config, log *logger.Logger) (kafka2.Consumer, kafka2.Consumer, kafka2.Producer) {
	pageConsumer, err := kafka2.NewConsumer(cfg.Kafka, pageConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create page Kafka consumer")
	}

	profileConsumer, err := kafka2.NewConsumer(cfg.Kafka, profileConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create profile Kafka consumer")
	}

	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}

	return pageConsumer, profileConsumer, producer
}

func consumeBatchWorkOrders(
	ctx context.Context,
	log *logger.Logger,
	consumer kafka2.Consumer,
	consumerGroup string,
	topics []string,
	sem *semaphore.Weighted,
	dispatchWg *sync.WaitGroup,
	poolName string,
	lastMessageTime *int64,
	totalProcessed *int64,
	totalFailed *int64,
	processAccount func(ctx context.Context, msg WorkOrderMessage) error,
) {
	log.Info().
		Strs("topics", topics).
		Str("consumer_group", consumerGroup).
		Msgf("Starting %s batch Kafka consumer", poolName)

	err := consumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
		if lastMessageTime != nil {
			atomic.StoreInt64(lastMessageTime, time.Now().UnixNano())
		}

		var batch LinkedInBatchWorkOrder
		if err := json.Unmarshal(value, &batch); err != nil {
			log.Error().Err(err).Str("pool", poolName).Str("function", "consumeBatchWorkOrders").Str("stage", "unmarshal_batch_work_order").Msg("Failed to unmarshal batch work order")
			return nil
		}

		total := len(batch.Accounts)
		log.Info().
			Str("batch_id", batch.BatchID).
			Int("accounts", total).
			Str("pool", poolName).
			Msg("Received batch work order, dispatching goroutines")

		var batchWg sync.WaitGroup
		var batchProcessed, batchFailed int64

		for _, account := range batch.Accounts {
			acc := account
			accountPayload, err := json.Marshal(acc)
			if err != nil {
				log.Error().Err(err).Str("linkedin_id", acc.LinkedinID).Str("function", "consumeBatchWorkOrders").Str("stage", "marshal_account_work_order").Msg("Failed to marshal account work order")
				atomic.AddInt64(&batchFailed, 1)
				continue
			}

			dispatchWg.Add(1)
			batchWg.Add(1)
			go func() {
				defer dispatchWg.Done()
				defer batchWg.Done()
				if err := sem.Acquire(ctx, 1); err != nil {
					atomic.AddInt64(&batchFailed, 1)
					return
				}
				defer sem.Release(1)

				perAccSem := semForAccount(acc.LinkedinID)
				if err := perAccSem.Acquire(ctx, 1); err != nil {
					atomic.AddInt64(&batchFailed, 1)
					return
				}
				defer perAccSem.Release(1)

				msg := WorkOrderMessage{AccountID: acc.ID, LinkedinID: acc.LinkedinID, Value: accountPayload}
				if err := processAccount(ctx, msg); err != nil {
					log.Error().Err(err).Str("linkedin_id", acc.LinkedinID).Str("pool", poolName).Str("function", "consumeBatchWorkOrders").Str("stage", "process_account").Msg("Failed to process account")
					atomic.AddInt64(&batchFailed, 1)
				} else {
					atomic.AddInt64(&batchProcessed, 1)
				}
			}()
		}

		batchID := batch.BatchID
		go func() {
			batchWg.Wait()
			p := atomic.LoadInt64(&batchProcessed)
			f := atomic.LoadInt64(&batchFailed)
			if totalProcessed != nil {
				atomic.AddInt64(totalProcessed, p)
			}
			if totalFailed != nil {
				atomic.AddInt64(totalFailed, f)
			}
			log.Info().
				Str("batch_id", batchID).
				Str("pool", poolName).
				Int("total", total).
				Int64("processed", p).
				Int64("failed", f).
				Msg("Batch processing complete")
		}()

		return nil
	})

	if err != nil && err != context.Canceled {
		log.Error().Err(err).Str("pool", poolName).Str("function", "consumeBatchWorkOrders").Str("stage", "consume_batch").Msgf("%s batch consumer error", poolName)
	}
}

func startTimestampUpdater(
	ctx context.Context,
	wg *sync.WaitGroup,
	repo mongodb.UnifiedSocialRepository,
	timestampUpdateChan <-chan TimestampUpdateRequest,
	log *logger.Logger,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("Timestamp updater started")

		for {
			select {
			case req, ok := <-timestampUpdateChan:
				if !ok {
					log.Info().Msg("Timestamp update channel closed")
					return
				}

				objectID, err := primitive.ObjectIDFromHex(req.AccountID)
				if err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("linkedin_id", req.LinkedinID).Str("function", "startTimestampUpdater").Str("stage", "parse_object_id").Msg("Invalid ObjectID for timestamp update")
					continue
				}

				now := time.Now().UTC()
				if err := repo.UpdateState(context.Background(), objectID, mongomodels.StateProcessed); err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("linkedin_id", req.LinkedinID).Str("function", "startTimestampUpdater").Str("stage", "update_account_state").Msg("Failed to update account state to Processed")
					continue
				}
				if err := repo.UpdateAnalyticsTimestamp(context.Background(), objectID, "analytics", now); err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("linkedin_id", req.LinkedinID).Str("function", "startTimestampUpdater").Str("stage", "update_analytics_timestamp").Msg("Failed to update analytics timestamp")
				} else {
					repo.ClearProcessingError(context.Background(), objectID)
					log.Debug().Str("linkedin_id", req.LinkedinID).Msg("Updated account state and analytics timestamp")
				}

			case <-ctx.Done():
				log.Info().Msg("Timestamp updater stopping")
				return
			}
		}
	}()
}
