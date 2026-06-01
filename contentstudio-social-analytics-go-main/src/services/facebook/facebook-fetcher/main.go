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
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// WorkOrderMessage represents a work order with its Kafka metadata
type WorkOrderMessage struct {
	AccountID  string // MongoDB _id for logging
	FacebookID string // Facebook ID for logging
	Value      []byte
}

// BatchMessage represents a batch message from Kafka
type BatchMessage struct {
	Key   []byte
	Value []byte
}

const (
	maxWorkers        = 15  // Maximum concurrent work order processors
	workOrderChanSize = 500 // Channel buffer size for work orders

	// maxConcurrentAccounts is the max number of Facebook accounts processed simultaneously.
	maxConcurrentAccounts = 50

	// idleTimeout is the duration after which the service will shutdown
	// if no new messages are received. This allows the service to exit
	// gracefully after batch processing is complete.
	idleTimeout       = 5 * time.Minute
	idleCheckInterval = 30 * time.Second
)

// isExpectedFacebookError returns true for expected auth/permission errors that should not go to Sentry.
func isExpectedFacebookError(err error) bool {
	if err == nil {
		return false
	}
	return social.IsExpectedCompetitorErrorFB(err)
}

// ----- per-account concurrency guard -----
// ensures we don't run multiple full pipelines for the same FacebookID at once
var accountSemaphores sync.Map // map[string]*semaphore.Weighted

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

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	// Initialize logger
	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Facebook Fetcher service")

	// ---- Build a RateManager from config (with sane defaults if zero) ----
	// Assumes cfg.Facebook has these fields; if not, wire them accordingly.
	perTokenRPS := cfg.Facebook.PerTokenRPS
	if perTokenRPS <= 0 {
		perTokenRPS = 4.0
	}
	perTokenBurst := cfg.Facebook.PerTokenBurst
	if perTokenBurst <= 0 {
		perTokenBurst = 4
	}
	globalRPS := cfg.Facebook.GlobalRPS
	if globalRPS <= 0 {
		globalRPS = 12.0
	}
	globalBurst := cfg.Facebook.GlobalBurst
	if globalBurst <= 0 {
		globalBurst = 12
	}
	// Optional: cap how many concurrent work orders for the same FB page (1-2 recommended)
	perAccountConcurrency := cfg.Facebook.PerAccountConcurrency
	if perAccountConcurrency <= 0 {
		perAccountConcurrency = 1
	}

	rm := social.NewRateManager(social.RateLimits{
		PerTokenRPS:   perTokenRPS,
		PerTokenBurst: perTokenBurst,
		GlobalRPS:     globalRPS,
		GlobalBurst:   globalBurst,
	})

	// Create Facebook client with rate manager
	facebookClient := social.NewFacebookClientWithRates(cfg.Facebook.AppSecret, rm)

	// Create Kafka consumer for batch topic
	consumer, err := kafka2.NewConsumer(cfg.Kafka, "facebook-fetcher-batch-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	// Create Kafka producer for publishing raw posts
	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	// Initialize MongoDB for state/timestamp updates
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
	if err := mongoClient.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}
	mongoRepo := mongodb.NewUnifiedSocialRepository(mongoClient.Database(cfg.Mongo.Database), log.Logger)

	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	accountSem := semaphore.NewWeighted(maxConcurrentAccounts)
	var dispatchWg sync.WaitGroup
	var totalProcessed, totalFailed int64

	// Track last message time for idle timeout detection
	var lastMessageTime int64 = time.Now().UnixNano()

	// Start batch consumer - unpacks batches and dispatches one goroutine per account
	go func() {
		topics := []string{"work-order-facebook"}
		log.Info().Strs("topics", topics).Msg("Starting batch consumer")

		if err := consumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			var batch kafkamodels.FacebookBatchWorkOrder
			if err := json.Unmarshal(value, &batch); err != nil {
				log.Error().Err(err).Str("function", "main").Str("stage", "unmarshal_batch_work_order").Msg("Failed to unmarshal batch work order")
				return nil
			}

			total := len(batch.Accounts)
			log.Info().
				Str("batch_id", batch.BatchID).
				Int("accounts", total).
				Str("sync_type", batch.SyncType).
				Msg("Received batch work order, dispatching goroutines")

			var batchWg sync.WaitGroup
			var batchProcessed, batchFailed int64

			for _, account := range batch.Accounts {
				acc := account
				accountPayload, err := json.Marshal(acc)
				if err != nil {
					log.Error().Err(err).Str("facebook_id", acc.FacebookID).Str("function", "main").Str("stage", "marshal_account_work_order").Msg("Failed to marshal account work order")
					atomic.AddInt64(&batchFailed, 1)
					continue
				}
				dispatchWg.Add(1)
				batchWg.Add(1)
				go func() {
					defer dispatchWg.Done()
					defer batchWg.Done()
					if err := accountSem.Acquire(ctx, 1); err != nil {
						atomic.AddInt64(&batchFailed, 1)
						return
					}
					defer accountSem.Release(1)
					if err := processWorkOrder(ctx, accountPayload, facebookClient, producer, mongoRepo, cfg.DecryptionKey, int64(perAccountConcurrency)); err != nil {
						log.Error().Err(err).Str("facebook_id", acc.FacebookID).Str("function", "main").Str("stage", "process_work_order").Msg("Failed to process work order")
						atomic.AddInt64(&batchFailed, 1)
					} else {
						atomic.AddInt64(&batchProcessed, 1)
					}
				}()
			}

			batchID := batch.BatchID
			syncType := batch.SyncType
			go func() {
				batchWg.Wait()
				p := atomic.LoadInt64(&batchProcessed)
				f := atomic.LoadInt64(&batchFailed)
				atomic.AddInt64(&totalProcessed, p)
				atomic.AddInt64(&totalFailed, f)
				log.Info().
					Str("batch_id", batchID).
					Str("sync_type", syncType).
					Int("total", total).
					Int64("processed", p).
					Int64("failed", f).
					Msg("Batch processing complete")
			}()

			return nil
		}); err != nil {
			if err != context.Canceled {
				log.Error().Err(err).Str("function", "main").Str("stage", "consume_batch").Msg("Batch consumer error")
			}
		}
	}()

	log.Info().
		Int("max_concurrent_accounts", maxConcurrentAccounts).
		Msg("Facebook Fetcher service is running")

	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping service...")

	cancel()
	dispatchWg.Wait()

	log.Info().
		Int64("total_processed", atomic.LoadInt64(&totalProcessed)).
		Int64("total_failed", atomic.LoadInt64(&totalFailed)).
		Msg("Facebook Fetcher service stopped")
}

// workOrderProcessor processes work orders from the channel
func workOrderProcessor(
	ctx context.Context,
	workerID int,
	workOrderChan <-chan WorkOrderMessage,
	facebookClient *social.FacebookClient,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	decryptionKey string,
	perAccountConcurrency int64, // injected capacity for per-account semaphore
) {
	log := logger.New("info").With().Int("worker_id", workerID).Str("component", "work-order-processor").Logger()
	log.Info().Msg("Work order processor started")

	for {
		select {
		case workOrder, ok := <-workOrderChan:
			if !ok {
				log.Info().Msg("Work order channel closed, stopping processor")
				return
			}

			if err := processWorkOrder(ctx, workOrder.Value, facebookClient, producer, mongoRepo, decryptionKey, perAccountConcurrency); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("facebook_id", workOrder.FacebookID).Str("function", "workOrderProcessor").Str("stage", "process_work_order").Msg("Failed to process work order")
			}

		case <-ctx.Done():
			log.Info().Msg("Context cancelled, stopping processor")
			return
		}
	}
}

// workOrderProcessorWithTracking processes work orders with active job tracking for graceful shutdown.
// Tracks activeJobs count and updates lastMessageTime on job completion.
func workOrderProcessorWithTracking(
	ctx context.Context,
	workerID int,
	workOrderChan <-chan WorkOrderMessage,
	facebookClient *social.FacebookClient,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	decryptionKey string,
	perAccountConcurrency int64,
	activeJobs *int64,
	lastMessageTime *int64,
) {
	log := logger.New("info").With().Int("worker_id", workerID).Str("component", "work-order-processor").Logger()
	log.Info().Msg("Work order processor started")

	for {
		select {
		case workOrder, ok := <-workOrderChan:
			if !ok {
				log.Info().Msg("Work order channel closed, stopping processor")
				return
			}

			// Track active job
			atomic.AddInt64(activeJobs, 1)

			if err := processWorkOrder(ctx, workOrder.Value, facebookClient, producer, mongoRepo, decryptionKey, perAccountConcurrency); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("facebook_id", workOrder.FacebookID).Str("function", "workOrderProcessorWithTracking").Str("stage", "process_work_order").Msg("Failed to process work order")
			}

			// Job completed - update tracking
			atomic.AddInt64(activeJobs, -1)
			atomic.StoreInt64(lastMessageTime, time.Now().UnixNano())

		case <-ctx.Done():
			log.Info().Msg("Context cancelled, stopping processor")
			return
		}
	}
}

// --------- helpers to derive reels ---------

// hasReelsMetric checks a video's insights for "blue_reels_play_count"
func hasReelsMetric(v kafkamodels.RawFacebookVideo) bool {
	return parsing.HasFacebookReelsMetric(v)
}

// processWorkOrder processes individual FacebookAccountWorkOrder messages
func processWorkOrder(
	ctx context.Context,
	value []byte,
	facebookClient *social.FacebookClient,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	decryptionKey string,
	perAccountConcurrency int64,
) (err error) {
	log := logger.New("info")
	log = &logger.Logger{Logger: log.With().Str("function", "processWorkOrder").Logger()}

	log.Info().
		Int("value_size", len(value)).
		Msg("Received work order message")

	// Track processing errors for the defer block to decide success vs failure
	var processingErr error

	// Parse work order
	var workOrder kafkamodels.FacebookAccountWorkOrder
	if err := json.Unmarshal(value, &workOrder); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("raw_value", string(value)).Str("function", "processWorkOrder").Str("stage", "unmarshal_work_order").Msg("Failed to unmarshal work order")
		return err
	}

	baseTags := map[string]string{
		"function":     "processWorkOrder",
		"account_id":   workOrder.ID,
		"workspace_id": workOrder.WorkspaceID,
		"facebook_id":  workOrder.FacebookID,
	}

	op := log.Operation("processWorkOrder").
		WithFields(map[string]interface{}{
			"account_id":   workOrder.ID,
			"facebook_id":  workOrder.FacebookID,
			"workspace_id": workOrder.WorkspaceID,
			"sync_type":    workOrder.SyncType,
		}).
		WithSentryTags(baseTags)
	op.Start("processing facebook account work order")
	defer func() {
		op.Complete(err, "")
	}()

	// Update Mongo state/timestamp after processing attempt.
	// On success: mark as processed and clear any previous processing error.
	// On error: record the processing error instead of marking as processed.
	defer func() {
		if mongoRepo == nil {
			return
		}
		if workOrder.ID == "" {
			return
		}
		accountID, parseErr := primitive.ObjectIDFromHex(workOrder.ID)
		if parseErr != nil {
			log.Warn().Err(parseErr).Str("account_id", workOrder.ID).Msg("Invalid account ID, skipping state update")
			return
		}

		if processingErr != nil {
			if recordErr := mongoRepo.RecordProcessingError(ctx, accountID, processingErr.Error()); recordErr != nil {
				log.Error().Err(recordErr).Str("error_message", recordErr.Error()).Str("account_id", workOrder.ID).Str("function", "processWorkOrder").Str("stage", "record_processing_error").Msg("Failed to record processing error")
			}
			return
		}

		updates := bson.M{
			"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02 15:04:05"),
			"state":                     mongomodels.StateProcessed,
		}

		if updateErr := mongoRepo.Update(ctx, accountID, updates); updateErr != nil {
			log.Error().Err(updateErr).Str("error_message", updateErr.Error()).Str("account_id", workOrder.ID).Str("function", "processWorkOrder").Str("stage", "update_account_state").Msg("Failed to update account state and timestamp")
		}

		if clearErr := mongoRepo.ClearProcessingError(ctx, accountID); clearErr != nil {
			log.Error().Err(clearErr).Str("error_message", clearErr.Error()).Str("account_id", workOrder.ID).Str("function", "processWorkOrder").Str("stage", "clear_processing_error").Msg("Failed to clear processing error")
		}
	}()

	log.Info().
		Str("id", workOrder.ID).
		Str("facebook_id", workOrder.FacebookID).
		Str("type", workOrder.Type).
		Str("workspace_id", workOrder.WorkspaceID).
		Str("sync_type", workOrder.SyncType).
		Bool("has_access_token", workOrder.AccessToken != "").
		Bool("has_long_access_token", workOrder.LongAccessToken != "").
		Msg("Processing Facebook account work order")

	// ---- per-account concurrency gate ----
	sem := semForAccount(workOrder.FacebookID, perAccountConcurrency)
	if err := sem.Acquire(ctx, 1); err != nil {
		processingErr = err
		return err
	}
	defer sem.Release(1)

	// Resolve access token (prefer long token when decryptable)
	accessToken := workOrder.AccessToken
	if workOrder.LongAccessToken != "" {
		if decrypted, err := crypto.DecryptToken(workOrder.LongAccessToken, decryptionKey); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("facebook_id", workOrder.FacebookID).Str("function", "processWorkOrder").Str("stage", "decrypt_long_token").Msg("Failed to decrypt long access token; falling back to regular token")
		} else {
			accessToken = decrypted
			log.Info().Str("facebook_id", workOrder.FacebookID).Msg("Successfully decrypted long access token")
		}
	}
	if accessToken == "" {
		log.Error().Str("facebook_id", workOrder.FacebookID).Str("function", "processWorkOrder").Str("stage", "resolve_access_token").Msg("No valid access token available")
		if accountID, parseErr := primitive.ObjectIDFromHex(workOrder.ID); parseErr == nil {
			mongoRepo.RecordProcessingError(ctx, accountID, "Access token is empty or decryption failed")
		}
		return nil // skip without error to avoid retries
	}

	// ==================== DATE RANGE CALCULATION ====================
	// Facebook Graph API uses Unix timestamps for 'since' and 'until' parameters.
	//
	// Date Range Strategy:
	// - 'until' is set to current time (now) to include the most recent posts
	// - 'since' is set to midnight UTC on the target day to ensure full day coverage
	//   (using time.Date with day offset ensures we get complete days, not partial)
	//
	// Sync Types:
	// - incremental: Last 14 days - for regular scheduled syncs to catch recent activity
	// - full_sync:   Last 90 days - for initial setup or recovery to backfill historical data
	// - default:     Last 14 days - fallback behavior matches incremental
	//
	// maxPages controls pagination limit (Facebook returns ~50 posts per page for feed)
	// ================================================================
	maxPages := 5
	until := time.Now()
	var since time.Time
	switch workOrder.SyncType {
	case "incremental":
		maxPages = 5
		// 14 days of data starting from midnight UTC
		since = time.Date(until.Year(), until.Month(), until.Day()-14, 0, 0, 0, 0, time.UTC)
	case "full_sync":
		maxPages = 100
		// 90 days of historical data starting from midnight UTC
		since = time.Date(until.Year(), until.Month(), until.Day()-90, 0, 0, 0, 0, time.UTC)
	default:
		maxPages = 5
		// Default to 14 days (same as incremental)
		since = time.Date(until.Year(), until.Month(), until.Day()-14, 0, 0, 0, 0, time.UTC)
	}
	log.Info().
		Str("facebook_id", workOrder.FacebookID).
		Str("sync_type", workOrder.SyncType).
		Int("max_pages", maxPages).
		Time("since", since).
		Time("until", until).
		Msg("Derived fetch parameters")

	overallStart := time.Now()

	// ==================== CONCURRENT DATA FETCHING ====================
	// Fetch posts, videos, and insights in parallel using errgroup for efficiency.
	//
	// Error Handling Strategy:
	// - POSTS (critical):    Errors cancel the entire group - posts are required
	// - VIDEOS (best-effort): Errors are logged but don't cancel - uses separate timeout
	// - INSIGHTS (best-effort): Errors are logged but don't cancel - uses separate timeout
	//
	// This ensures we always get posts (or fail fast), while videos/insights
	// failures don't block the entire sync operation.
	// ================================================================
	var (
		posts            []kafkamodels.RawFacebookPost
		videos           []kafkamodels.RawFacebookVideo
		ins              *kafkamodels.RawFacebookInsights
		postsErr         error
		postsErrExpected bool
	)

	// errgroup with context: allows posts to cancel the group on failure
	g, gctx := errgroup.WithContext(ctx)

	// POSTS (critical): returning error cancels gctx and fails the run
	g.Go(func() error {
		span := time.Now()
		var ps []kafkamodels.RawFacebookPost
		var err error

		log.Info().Str("facebook_id", workOrder.FacebookID).Time("since", since).Time("until", until).Str("sync_type", workOrder.SyncType).Msg("Fetching Facebook posts with date filter")
		ps, err = facebookClient.FetchPostsSince(gctx, workOrder.FacebookID, accessToken, since, until)

		if err != nil {
			postsErrExpected = isExpectedFacebookError(err)
			if postsErrExpected {
				log.Warn().Err(err).Str("facebook_id", workOrder.FacebookID).Msg("Failed to fetch Facebook posts")
			} else {
				log.Error().Err(err).Str("error_message", err.Error()).Str("facebook_id", workOrder.FacebookID).Str("function", "processWorkOrder").Str("stage", "fetch_posts").Msg("Failed to fetch Facebook posts")
			}
			postsErr = err
			return err // cancel errgroup
		}
		posts = ps
		log.Info().
			Str("facebook_id", workOrder.FacebookID).
			Int("posts_count", len(posts)).
			Str("sync_type", workOrder.SyncType).
			Dur("fetch_dur", time.Since(span)).
			Msg("Fetched Facebook posts")
		return nil
	})

	// VIDEOS (best-effort): never cancel the group; use its own timeout
	g.Go(func() error {
		span := time.Now()

		// separate child timeout; IMPORTANT: base ctx (not gctx), so it won’t be canceled by errgroup
		windows := (maxPages*50 + 19) / 20 // rough upper bound (50 items/page, 20 IDs/window for multi-GET)
		budget := time.Duration(windows)*1200*time.Millisecond + time.Duration(maxPages)*2*time.Second
		if budget < 2*time.Minute {
			budget = 2 * time.Minute
		}
		if budget > 10*time.Minute {
			budget = 10 * time.Minute
		}
		vctx, cancel := context.WithTimeout(ctx, budget)
		defer cancel()

		log.Info().Str("facebook_id", workOrder.FacebookID).Time("since", since).Time("until", until).Str("sync_type", workOrder.SyncType).Msg("Fetching Facebook videos with date filter")
		vs, err := facebookClient.FetchVideosSince(vctx, workOrder.FacebookID, accessToken, since, until)

		if err != nil {
			// Best-effort: log and move on
			if isExpectedFacebookError(err) {
				log.Warn().Err(err).Str("facebook_id", workOrder.FacebookID).Msg("Failed to fetch Facebook videos (best-effort, continuing)")
			} else {
				log.Error().Err(err).Str("error_message", err.Error()).Str("facebook_id", workOrder.FacebookID).Str("function", "processWorkOrder").Str("stage", "fetch_videos").Msg("Failed to fetch Facebook videos (best-effort, continuing)")
			}
			return nil // do NOT cancel group
		}
		videos = vs
		log.Info().
			Str("facebook_id", workOrder.FacebookID).
			Int("videos_count", len(videos)).
			Str("sync_type", workOrder.SyncType).
			Dur("fetch_dur", time.Since(span)).
			Msg("Fetched Facebook videos")
		return nil
	})

	// INSIGHTS (best-effort): never cancel the group; use its own timeout
	g.Go(func() error {
		span := time.Now()

		ictx, cancel := context.WithTimeout(ctx, 75*time.Second)
		defer cancel()

		log.Info().Str("facebook_id", workOrder.FacebookID).Str("sync_type", workOrder.SyncType).Msg("Fetching Facebook page insights")
		insights, err := fetchInsights(ictx, facebookClient, workOrder.FacebookID, accessToken, workOrder.SyncType, log)
		if err != nil {
			if isExpectedFacebookError(err) {
				log.Warn().Err(err).Str("facebook_id", workOrder.FacebookID).Msg("Failed to fetch Facebook page insights (best-effort, continuing)")
			} else {
				log.Error().Err(err).Str("error_message", err.Error()).Str("facebook_id", workOrder.FacebookID).Str("function", "processWorkOrder").Str("stage", "fetch_insights").Msg("Failed to fetch Facebook page insights (best-effort, continuing)")
			}
			return nil // do NOT cancel group
		}
		ins = insights
		if ins != nil {
			log.Info().
				Str("facebook_id", workOrder.FacebookID).
				Int("insights_count", len(ins.Data)).
				Dur("fetch_dur", time.Since(span)).
				Msg("Fetched Facebook page insights")
		} else {
			log.Warn().Str("facebook_id", workOrder.FacebookID).Msg("No insights returned")
		}
		return nil
	})

	// wait for all; only posts can produce a group error
	_ = g.Wait()

	// Decide how fatal postsErr is:
	if postsErr != nil && len(posts) == 0 {
		if social.IsAuthError(postsErr) {
			processingErr = postsErr
			return nil
		}
		// Non-auth error: still publish videos/insights if present
		if postsErrExpected {
			log.Warn().Err(postsErr).Msg("Posts fetch failed, but will still publish videos/insights if present")
		} else {
			log.Error().Err(postsErr).Str("error_message", postsErr.Error()).Str("function", "processWorkOrder").Str("stage", "fetch_posts").Msg("Posts fetch failed, but will still publish videos/insights if present")
		}
		// don't return here
	}

	// --- publish posts ---
	if len(posts) > 0 {
		pubStart := time.Now()
		log.Info().Str("facebook_id", workOrder.FacebookID).Int("posts_count", len(posts)).Msg("Publishing posts to raw-facebook-posts")
		publishPostsParallel(ctx, posts, producer, workOrder.FacebookID, workOrder.WorkspaceID, 0, log)
		log.Info().
			Str("facebook_id", workOrder.FacebookID).
			Int("total_posts", len(posts)).
			Dur("publish_dur", time.Since(pubStart)).
			Msg("Completed publishing posts")
	}

	// --- publish videos ---
	if len(videos) > 0 {
		filteredVideos, skippedVideos := parsing.FilterFacebookVideos(workOrder.FacebookID, posts, videos)
		if skippedVideos > 0 {
			log.Warn().
				Str("facebook_id", workOrder.FacebookID).
				Int("videos_fetched", len(videos)).
				Int("videos_skipped", skippedVideos).
				Int("videos_allowed", len(filteredVideos)).
				Msg("Skipped raw Facebook videos that did not match a video or reel post")
		}
		videos = filteredVideos
	}

	if len(videos) > 0 {
		pubStart := time.Now()
		log.Info().Str("facebook_id", workOrder.FacebookID).Int("videos_count", len(videos)).Msg("Publishing videos to raw-facebook-videos")
		publishVideosParallel(ctx, videos, producer, workOrder.FacebookID, workOrder.WorkspaceID, 0, log)
		log.Info().
			Str("facebook_id", workOrder.FacebookID).
			Int("total_videos", len(videos)).
			Dur("publish_dur", time.Since(pubStart)).
			Msg("Completed publishing videos")
	}

	// --- publish insights (optional) ---
	if ins != nil && len(ins.Data) > 0 {
		pubStart := time.Now()
		log.Info().Str("facebook_id", workOrder.FacebookID).Int("insights_count", len(ins.Data)).Msg("Publishing insights to raw-facebook-insights")
		publishInsightsParallel(ctx, []kafkamodels.RawFacebookInsights{*ins}, producer, workOrder.FacebookID, workOrder.WorkspaceID, 0, log)
		log.Info().
			Str("facebook_id", workOrder.FacebookID).
			Int("total_insights", len(ins.Data)).
			Dur("publish_dur", time.Since(pubStart)).
			Msg("Completed publishing insights")
	}

	// Summary
	log.Info().
		Str("id", workOrder.ID).
		Str("facebook_id", workOrder.FacebookID).
		Str("sync_type", workOrder.SyncType).
		Int("posts_processed", len(posts)).
		Int("videos_processed", len(videos)).
		Int("insights_metrics", func() int {
			if ins == nil {
				return 0
			}
			return len(ins.Data)
		}()).
		Dur("total_parallel_time", time.Since(overallStart)).
		Msg("Work order completed (parallel)")

	return nil
}
