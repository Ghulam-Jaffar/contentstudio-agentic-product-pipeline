package main

import (
	"context"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/sync/semaphore"
)

const (
	// Worker pool sizes - kept for backwards compatibility with existing tests
	maxMediaWorkers    = 10
	maxInsightsWorkers = 5

	// Queue sizes - kept for backwards compatibility with existing tests
	mediaQueueSize    = 500
	insightsQueueSize = 500

	// Concurrency limits - controls API call parallelism
	// Instagram API: ~200 calls/user/hour, be conservative
	mediaInsightsConcPerWorker   = 3 // Max concurrent media insights API calls per worker (total: 30)
	accountInsightsConcPerWorker = 2 // Max concurrent account insights API calls per worker (total: 10)

	// maxConcurrentAccounts is the max number of Instagram accounts processed simultaneously.
	// At ~60s/account avg: 50 accounts × 60s = ~3.5h for 10,509 accounts (within 4h target).
	maxConcurrentAccounts = 50

	// Per-account concurrency limit prevents duplicate processing
	perAccountConcurrency int64 = 1

	// Timestamp update channel size
	timestampUpdateChanSize = 1000
)

// Service-level concurrency semaphores
var (
	// accountSemaphores prevents duplicate processing of the same Instagram account
	accountSemaphores sync.Map

	mediaInsightsConc = semaphore.NewWeighted(int64(mediaInsightsConcPerWorker * maxMediaWorkers))
)

// TimestampUpdateRequest represents a request to update last_analytics_updated_at
type TimestampUpdateRequest struct {
	AccountID   string
	InstagramID string
}

// semForAccount returns a semaphore for the given account ID
func semForAccount(id string, capacity int64) *semaphore.Weighted {
	if v, ok := accountSemaphores.Load(id); ok {
		return v.(*semaphore.Weighted)
	}
	sem := semaphore.NewWeighted(capacity)
	if old, loaded := accountSemaphores.LoadOrStore(id, sem); loaded {
		return old.(*semaphore.Weighted)
	}
	return sem
}

// isExpectedInstagramError returns true for expected auth/permission/viewer errors that should not go to Sentry.
func isExpectedInstagramError(err error) bool {
	if err == nil {
		return false
	}
	if social.IsAuthError(err) || social.IsExpectedCompetitorError(err) {
		return true
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "not enough viewers") || strings.Contains(errStr, "(#10)")
}

type ResolvedOrder struct {
	AccountID             string // MongoDB _id for timestamp updates
	InstagramID           string
	WorkspaceID           string
	AccessTokenPlaintext  string
	ConnectedViaInstagram bool
	AppSecret             string
}

type MediaJob struct {
	Order    ResolvedOrder
	SyncType string     // "incremental", "full_sync", "immediate"
	Since    *time.Time // nil for full_sync/immediate (fetch all), set for incremental
	Ack      func()     // called when fully processed; may be nil
}

type InsightsJob struct {
	Order    ResolvedOrder
	SyncType string
	Since    time.Time
	Until    time.Time
	Ack      func() // called when fully processed; may be nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("media_workers", maxMediaWorkers).
		Int("insights_workers", maxInsightsWorkers).
		Msg("Starting Instagram Fetcher")

	// MongoDB
	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(context.Background())

	pingCtx, pingCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer pingCancel()
	if err := mongoClient.Ping(pingCtx, readpref.Primary()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}
	log.Info().Msg("MongoDB connected")

	db := mongoClient.Database(cfg.Mongo.Database)
	mongoRepo := mongodb.NewUnifiedSocialRepository(db, log.Logger)

	// Kafka
	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	consumer, err := kafka2.NewConsumer(cfg.Kafka, "instagram-fetcher-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	// Signal handling
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received, stopping dispatch (in-progress jobs will complete)...")
		cancel()
	}()

	svc := NewFetcherService(
		FetcherConfig{
			MaxMediaWorkers:         maxMediaWorkers,
			MaxInsightsWorkers:      maxInsightsWorkers,
			MediaQueueSize:          mediaQueueSize,
			InsightsQueueSize:       insightsQueueSize,
			MaxConcurrentAccounts:   maxConcurrentAccounts,
			TimestampUpdateChanSize: timestampUpdateChanSize,
			DecryptionKey:           cfg.DecryptionKey,
			AppSecret:               cfg.Facebook.AppSecret,
		},
		FetcherDependencies{
			Producer:      producer,
			Consumer:      consumer,
			MongoRepo:     mongoRepo,
			ClientFactory: DefaultInstagramClientFactory,
			Log:           log,
		},
	)

	startTime := time.Now()
	if err := svc.Run(ctx); err != nil {
		log.Error().Err(err).Msg("Instagram Fetcher failed")
	}

	log.Info().
		Dur("total_duration", time.Since(startTime)).
		Msg("Instagram Fetcher completed — all jobs and goroutines finished")
}

// startTimestampUpdater starts a goroutine that updates last_analytics_updated_at after successful processing.
func startTimestampUpdater(
	wg *sync.WaitGroup,
	repo mongodb.UnifiedSocialRepository,
	timestampUpdateChan <-chan TimestampUpdateRequest,
	log *logger.Logger,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().Msg("Timestamp updater started")

		for req := range timestampUpdateChan {
			account, err := repo.GetByPlatformID(context.Background(), "instagram", req.InstagramID)
			if err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("instagram_id", req.InstagramID).Str("function", "startTimestampUpdater").Str("stage", "find_account").Msg("Failed to find account for timestamp update")
				continue
			}
			if account == nil {
				log.Warn().Str("instagram_id", req.InstagramID).Msg("Account not found for timestamp update")
				continue
			}

			now := time.Now().UTC()
			if err := repo.UpdateState(context.Background(), account.ID, mongomodels.StateProcessed); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("instagram_id", account.PlatformIdentifier).Str("function", "startTimestampUpdater").Str("stage", "update_account_state").Msg("Failed to update account state to Processed")
				continue
			}
			if err := repo.UpdateAnalyticsTimestamp(context.Background(), account.ID, "analytics", now); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("instagram_id", account.PlatformIdentifier).Str("function", "startTimestampUpdater").Str("stage", "update_analytics_timestamp").Msg("Failed to update analytics timestamp")
			} else {
				repo.ClearProcessingError(context.Background(), account.ID)
				log.Debug().Str("instagram_id", account.PlatformIdentifier).Msg("Updated account state and analytics timestamp")
			}
		}

		log.Info().Msg("Timestamp update channel closed, updater stopping")
	}()
}

// ---------- helpers ----------

func processAccount(
	ctx context.Context,
	order kafkamodels.InstagramAccountWorkOrder,
	decryptionKey string,
	appSecret string,
	mediaJobs chan<- MediaJob,
	insightsJobs chan<- InsightsJob,
	log *logger.Logger,
	mongoRepo mongodb.UnifiedSocialRepository,
	mediaAck func(),
	insightsAck func(),
) error {
	token := resolveAccessToken(order.AccessToken, decryptionKey, order.InstagramID, log)

	if token == "" {
		log.Error().Str("instagram_id", order.InstagramID).Str("function", "processAccount").Str("stage", "resolve_token").Msg("Empty access token after resolution; skipping")
		if mongoRepo != nil {
			if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
				mongoRepo.RecordProcessingError(ctx, accountID, "Access token is empty or decryption failed")
			}
		}
		// No jobs dispatched — release both pre-added WaitGroup counts.
		if mediaAck != nil {
			mediaAck()
		}
		if insightsAck != nil {
			insightsAck()
		}
		return nil
	}

	ro := ResolvedOrder{
		AccountID:             order.ID,
		InstagramID:           order.InstagramID,
		WorkspaceID:           order.WorkspaceID,
		AccessTokenPlaintext:  token,
		ConnectedViaInstagram: order.ConnectedViaInstagram,
		AppSecret:             appSecret,
	}

	// Calculate date range based on sync type
	// - incremental: Last 14 days
	// - full_sync/immediate: All data (nil since)
	var mediaSince *time.Time
	switch order.SyncType {
	case "incremental":
		t := time.Now().UTC().AddDate(0, 0, -14)
		mediaSince = &t
	}

	// media job
	select {
	case mediaJobs <- MediaJob{Order: ro, SyncType: order.SyncType, Since: mediaSince, Ack: mediaAck}:
	case <-ctx.Done():
		// Neither job dispatched.
		if mediaAck != nil {
			mediaAck()
		}
		if insightsAck != nil {
			insightsAck()
		}
		return ctx.Err()
	}

	// Calculate insights date range based on sync type:
	// - incremental: Last 14 days
	// - immediate: Last 30 days
	// - full_sync: Last 89 days
	// Instagram uses 08:00 UTC as day boundary (midnight Pacific Time)
	// Formula to match Instagram dashboard:
	// - Since: (Today - N days) at 08:00 UTC
	// - Until: Yesterday at 05:00 UTC
	today := time.Now().UTC().Truncate(24 * time.Hour)
	until := today.AddDate(0, 0, -1).Add(5 * time.Hour) // Yesterday at 05:00 UTC
	var insightsSince time.Time
	switch order.SyncType {
	case "incremental":
		insightsSince = today.AddDate(0, 0, -15).Add(8 * time.Hour) // 14 days ago at 08:00 UTC
	case "immediate":
		insightsSince = today.AddDate(0, 0, -30).Add(8 * time.Hour) // 30 days ago at 08:00 UTC
	default:
		// full_sync: 89 days
		insightsSince = today.AddDate(0, 0, -89).Add(8 * time.Hour) // 89 days ago at 08:00 UTC
	}

	select {
	case insightsJobs <- InsightsJob{Order: ro, SyncType: order.SyncType, Since: insightsSince, Until: until, Ack: insightsAck}:
	case <-ctx.Done():
		// Media job was dispatched (its worker will call mediaAck); insights was not.
		if insightsAck != nil {
			insightsAck()
		}
		return ctx.Err()
	}

	return nil
}

func resolveAccessToken(accessToken, decryptionKey, igID string, log *logger.Logger) string {
	if accessToken == "" {
		return ""
	}
	// plain tokens: Instagram starts w/ IGAA..., Facebook starts w/ EAA...
	if strings.HasPrefix(accessToken, "IGAA") || strings.HasPrefix(accessToken, "EAA") {
		return accessToken
	}
	if dec, err := crypto.DecryptToken(accessToken, decryptionKey); err == nil && dec != "" {
		return dec
	}
	log.Error().Str("instagram_id", igID).Msg("Token decryption failed; account will be skipped")
	return ""
}
