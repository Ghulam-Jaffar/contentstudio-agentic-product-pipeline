package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/semaphore"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

// FetcherConfig holds service configuration
type FetcherConfig struct {
	MaxWorkers               int
	WorkQueueSize            int
	TweetsPerPage            int
	DecryptionKey            string
	IdleTimeout              time.Duration
	IdleCheckPeriod          time.Duration
	WorkOrdersTopic          string
	RawPostsTopic            string
	RawInsightsTopic         string
	AccountSemaphoreCapacity int64
	MaxConcurrentAccounts    int
}

// DefaultFetcherConfig returns default configuration
func DefaultFetcherConfig() FetcherConfig {
	return FetcherConfig{
		MaxWorkers:               10,
		WorkQueueSize:            200,
		TweetsPerPage:            100,
		IdleTimeout:              15 * time.Minute,
		IdleCheckPeriod:          30 * time.Second,
		WorkOrdersTopic:          "work-order-twitter-batch",
		RawPostsTopic:            "raw-twitter-posts",
		RawInsightsTopic:         "raw-twitter-insights",
		AccountSemaphoreCapacity: 1,
		MaxConcurrentAccounts:    50,
	}
}

// FetcherDependencies holds external dependencies
type FetcherDependencies struct {
	Consumer      kafka2.Consumer
	Producer      kafka2.Producer
	MongoRepo     mongodb.UnifiedSocialRepository
	TwitterClient social.TwitterAPI
	Log           *logger.Logger
}

// FetcherMetrics holds service metrics
type FetcherMetrics struct {
	BatchesReceived    uint64
	WorkOrdersConsumed uint64
	AccountsProcessed  uint64
	TweetsProduced     uint64
	InsightsProduced   uint64
	ProducerErrors     uint64
	WorkOrderErrors    uint64
	UserInfoRequests   uint64
	TweetAPIRequests   uint64
}

// WorkOrderMessage represents a work order message
type WorkOrderMessage struct {
	Key   []byte
	Value []byte
	Ack   func() // called when fully processed; may be nil
}

// accountSemaphores Per-account concurrency guard - ensures we don't run multiple pipelines for the same Twitter account at once
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

// RunService starts the Twitter fetcher service with the given dependencies
func RunService(ctx context.Context, deps *FetcherDependencies, cfg FetcherConfig) error {
	log := deps.Log

	metrics := &FetcherMetrics{}

	log.Info().
		Int("max_workers", cfg.MaxWorkers).
		Dur("idle_timeout", cfg.IdleTimeout).
		Msg("Starting Twitter Fetcher Service")

	var wg sync.WaitGroup

	// Start batch processor
	wg.Add(1)
	go processBatches(ctx, deps, &wg, cfg, metrics, log)

	// Wait for all processors to complete
	wg.Wait()

	log.Info().
		Uint64("batches_received", atomic.LoadUint64(&metrics.BatchesReceived)).
		Uint64("work_orders_consumed", atomic.LoadUint64(&metrics.WorkOrdersConsumed)).
		Uint64("accounts_processed", atomic.LoadUint64(&metrics.AccountsProcessed)).
		Uint64("tweets_produced", atomic.LoadUint64(&metrics.TweetsProduced)).
		Uint64("insights_produced", atomic.LoadUint64(&metrics.InsightsProduced)).
		Uint64("user_info_requests", atomic.LoadUint64(&metrics.UserInfoRequests)).
		Uint64("tweet_api_requests", atomic.LoadUint64(&metrics.TweetAPIRequests)).
		Uint64("producer_errors", atomic.LoadUint64(&metrics.ProducerErrors)).
		Msg("Twitter Fetcher Service stopped")

	return nil
}

// processBatches processes batch work orders using goroutine-per-account dispatch.
// ack() is called immediately after launching goroutines so Kafka consumer is not blocked.
func processBatches(ctx context.Context, deps *FetcherDependencies, wg *sync.WaitGroup, cfg FetcherConfig, metrics *FetcherMetrics, log *logger.Logger) {
	defer wg.Done()

	maxConc := cfg.MaxConcurrentAccounts
	if maxConc <= 0 {
		maxConc = 50
	}
	accountSem := semaphore.NewWeighted(int64(maxConc))
	var dispatchWg sync.WaitGroup

	err := deps.Consumer.ConsumeWithAck(ctx, []string{cfg.WorkOrdersTopic}, func(ctx context.Context, topic string, key, value []byte, ack func()) error {
		var batch kafkamodels.TwitterBatchWorkOrder
		if err := json.Unmarshal(value, &batch); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processBatches").Str("stage", "unmarshal_batch_work_order").Msg("failed to unmarshal batch work order")
			atomic.AddUint64(&metrics.WorkOrderErrors, 1)
			ack()
			return nil
		}

		atomic.AddUint64(&metrics.BatchesReceived, 1)
		total := len(batch.Accounts)
		log.Info().
			Str("batch_id", batch.BatchID).
			Int("account_count", total).
			Msg("processing batch work order, dispatching goroutines")

		var batchWg sync.WaitGroup
		var batchProcessed, batchFailed int64

		for _, account := range batch.Accounts {
			acc := account
			accountData, err := json.Marshal(acc)
			if err != nil {
				log.Error().Err(err).Str("twitter_id", acc.TwitterID).Msg("Failed to marshal account work order; skipping")
				atomic.AddInt64(&batchFailed, 1)
				continue
			}
			atomic.AddUint64(&metrics.WorkOrdersConsumed, 1)
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

				if err := handleWorkOrder(ctx, []byte(acc.TwitterID), accountData, deps, cfg, metrics, log); err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processBatches").Str("stage", "handle_work_order").Msg("work order failed")
					atomic.AddUint64(&metrics.WorkOrderErrors, 1)
					atomic.AddInt64(&batchFailed, 1)
				} else {
					atomic.AddUint64(&metrics.AccountsProcessed, 1)
					atomic.AddInt64(&batchProcessed, 1)
				}
			}()
		}

		batchID := batch.BatchID
		go func() {
			batchWg.Wait()
			log.Info().
				Str("batch_id", batchID).
				Int("total", total).
				Int64("processed", atomic.LoadInt64(&batchProcessed)).
				Int64("failed", atomic.LoadInt64(&batchFailed)).
				Msg("Batch processing complete")
		}()
		ack()
		return nil
	})

	if err != nil && err != context.Canceled {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processBatches").Str("stage", "consume_work_orders").Msg("consumer error")
	}

	dispatchWg.Wait()
}

// handleWorkOrder processes a single Twitter work order
func handleWorkOrder(ctx context.Context, key, value []byte, deps *FetcherDependencies, cfg FetcherConfig, metrics *FetcherMetrics, log *logger.Logger) (err error) {
	var order kafkamodels.TwitterAccountWorkOrder
	if err := json.Unmarshal(value, &order); err != nil {
		return err
	}

	sem := semForAccount(order.TwitterID, cfg.AccountSemaphoreCapacity)
	if err := sem.Acquire(ctx, 1); err != nil {
		log.Warn().Err(err).Str("twitter_id", order.TwitterID).Msg("Failed to acquire semaphore")
		return nil
	}
	defer sem.Release(1)

	log.Info().Str("twitter_id", order.TwitterID).Str("sync_type", order.SyncType).Msg("Processing Twitter account")
	log.Info().
		Str("twitter_id", order.TwitterID).
		Str("platform_identifier", order.TwitterID).
		Int("n_tweets", order.NTweets).
		Str("sync_type", order.SyncType).
		Str("workspace_id", order.WorkspaceID).
		Msg("Starting Twitter fetcher work order processing")

	// Track success/failure for MongoDB update
	var processingError error
	defer func() {
		if order.ID == "" {
			return
		}

		accountID, parseErr := primitive.ObjectIDFromHex(order.ID)
		if parseErr != nil {
			log.Warn().Err(parseErr).Str("account_id", order.ID).Msg("Invalid account ID, skipping state update")
			return
		}

		if processingError != nil {
			if recordErr := deps.MongoRepo.RecordProcessingError(ctx, accountID, processingError.Error()); recordErr != nil {
				log.Error().Err(recordErr).Str("error_message", recordErr.Error()).Str("account_id", order.ID).Str("function", "handleWorkOrder").Str("stage", "record_processing_error").Msg("Failed to record processing error")
			}
			return
		}

		// Update MongoDB only on successful processing (python parity).
		updates := bson.M{
			"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02 15:04:05"),
		}

		updates["state"] = mongomodels.StateProcessed

		if updateErr := deps.MongoRepo.Update(ctx, accountID, updates); updateErr != nil {
			log.Error().Err(updateErr).Str("error_message", updateErr.Error()).Str("account_id", order.ID).Str("function", "handleWorkOrder").Str("stage", "update_account_state").Msg("Failed to update account state")
		} else {
			log.Info().
				Str("account_id", order.ID).
				Str("state", updates["state"].(string)).
				Str("last_analytics_updated_at", updates["last_analytics_updated_at"].(string)).
				Msg("Updated account state and timestamp")
		}

		if clearErr := deps.MongoRepo.ClearProcessingError(ctx, accountID); clearErr != nil {
			log.Error().Err(clearErr).Str("error_message", clearErr.Error()).Str("account_id", order.ID).Str("function", "handleWorkOrder").Str("stage", "clear_processing_error").Msg("Failed to clear processing error")
		}
	}()

	// Decrypt tokens
	oauthToken := order.OAuthToken
	if dec, err := crypto.DecryptToken(oauthToken, cfg.DecryptionKey); err == nil {
		oauthToken = dec
	}
	oauthTokenSecret := order.OAuthTokenSecret
	if dec, err := crypto.DecryptToken(oauthTokenSecret, cfg.DecryptionKey); err == nil {
		oauthTokenSecret = dec
	}

	// Fetch user info
	userInfoFetched := false
	rawTweetsFetched := 0
	twClient := deps.TwitterClient
	if order.APIKey != "" && order.APISecret != "" {
		twClient = social.NewTwitterClient(order.APIKey, order.APISecret)
	}
	atomic.AddUint64(&metrics.UserInfoRequests, 1)
	log.Info().Str("twitter_id", order.TwitterID).Msg("Calling Twitter user info endpoint")
	userResp, err := twClient.FetchUserInfo(ctx, []string{order.TwitterID}, oauthToken, oauthTokenSecret)
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("twitter_id", order.TwitterID).Str("function", "handleWorkOrder").Str("stage", "fetch_user_info").Msg("Failed to fetch user info")
		processingError = err
		return err
	}

	var userInfo *social.TwitterUser
	userRecords := 0
	if userResp != nil && len(userResp.Data) > 0 {
		userInfo = &userResp.Data[0]
		userInfoFetched = true
		userRecords = len(userResp.Data)
	} else if userResp != nil {
		userRecords = len(userResp.Data)
	}
	log.Info().
		Str("twitter_id", order.TwitterID).
		Int("user_records", userRecords).
		Bool("user_info_fetched", userInfoFetched).
		Msg("Completed Twitter user info endpoint")

	parser := parsing.NewTwitterParser()

	// Fetch tweets with pagination
	var paginationToken string
	tweetsProcessed := 0
	requestedPostCount := order.NTweets
	remainingTweets := order.NTweets
	if remainingTweets <= 0 {
		requestedPostCount = 30
		remainingTweets = 30
	}
	usePagination := requestedPostCount > 100
	log.Info().
		Str("twitter_id", order.TwitterID).
		Int("requested_post_count", requestedPostCount).
		Bool("use_pagination", usePagination).
		Msg("Prepared tweet fetch plan")

	pageNumber := 0
	for remainingTweets > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		pageSize := remainingTweets
		if usePagination {
			pageSize = 50
		}
		if pageSize > 100 {
			pageSize = 100
		}
		pageNumber++
		atomic.AddUint64(&metrics.TweetAPIRequests, 1)
		log.Info().
			Str("twitter_id", order.TwitterID).
			Int("page_number", pageNumber).
			Int("page_size", pageSize).
			Int("remaining_tweets_before_call", remainingTweets).
			Bool("has_pagination_token", paginationToken != "").
			Msg("Calling Twitter tweets endpoint")
		resp, err := twClient.FetchUserTweets(ctx, order.TwitterID, oauthToken, oauthTokenSecret, pageSize, paginationToken)
		if err != nil {
			if social.IsExpectedTwitterError(err) {
				log.Warn().Err(err).Str("twitter_id", order.TwitterID).Msg("Expected Twitter API error")
			} else {
				log.Error().Err(err).Str("error_message", err.Error()).Str("twitter_id", order.TwitterID).Str("function", "handleWorkOrder").Str("stage", "fetch_tweets").Msg("Failed to fetch tweets")
			}
			if social.IsAuthError(err) {
				processingError = err
			}
			break
		}

		if resp == nil || len(resp.Data) == 0 {
			log.Info().
				Str("twitter_id", order.TwitterID).
				Int("page_number", pageNumber).
				Msg("Twitter tweets endpoint returned empty page")
			break
		}
		rawTweetsFetched += len(resp.Data)
		log.Info().
			Str("twitter_id", order.TwitterID).
			Int("page_number", pageNumber).
			Int("raw_tweets_in_page", len(resp.Data)).
			Int("raw_tweets_fetched_total", rawTweetsFetched).
			Msg("Processed Twitter tweets endpoint page")

		for _, tweet := range resp.Data {
			parsedPost := parser.ParseTweet(tweet, userInfo, resp.Includes)
			if parsedPost == nil {
				continue
			}

			rawPost := kafkamodels.RawTwitterPost{
				WorkspaceID: order.WorkspaceID,
				TwitterID:   order.TwitterID,
			}
			postJSON, err := json.Marshal(parsedPost)
			if err != nil {
				log.Error().Err(err).Str("tweet_id", parsedPost.TweetID).Msg("Failed to marshal parsed tweet; skipping")
				continue
			}
			rawPost.Data = postJSON
			postData, err := json.Marshal(rawPost)
			if err != nil {
				log.Error().Err(err).Str("tweet_id", parsedPost.TweetID).Msg("Failed to marshal raw post; skipping")
				continue
			}

			if err := deps.Producer.Produce(ctx, cfg.RawPostsTopic, []byte(parsedPost.TweetID), postData); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("tweet_id", parsedPost.TweetID).Str("function", "handleWorkOrder").Str("stage", "produce_post").Msg("Failed to publish post")
				atomic.AddUint64(&metrics.ProducerErrors, 1)
			} else {
				atomic.AddUint64(&metrics.TweetsProduced, 1)
			}

			tweetsProcessed++
			remainingTweets--
			if remainingTweets <= 0 {
				break
			}
		}

		if remainingTweets <= 0 {
			break
		}
		if len(resp.Data) == 0 {
			break
		}
		if !usePagination {
			break
		}
		if resp.Meta == nil || resp.Meta.NextToken == "" {
			log.Info().
				Str("twitter_id", order.TwitterID).
				Int("page_number", pageNumber).
				Msg("No next pagination token, stopping tweet fetch")
			break
		}
		if resp.Meta != nil {
			paginationToken = resp.Meta.NextToken
		}

		time.Sleep(300 * time.Millisecond)
	}

	log.Info().
		Str("twitter_id", order.TwitterID).
		Int("tweets_processed", tweetsProcessed).
		Msg("Completed fetching tweets")

	creditsUsed := rawTweetsFetched
	if userInfoFetched {
		creditsUsed++
	}
	if err := deps.MongoRepo.InsertTwitterJobMetadata(ctx, mongodb.TwitterJobMetadataPayload{
		PlatformID:  order.TwitterID,
		WorkspaceID: order.WorkspaceID,
		CreditsUsed: creditsUsed,
		ExecutedBy:  firstNonEmpty(order.ExecutedBy, "internal"),
		AppID:       order.AppID,
		AppName:     order.AppName,
	}); err != nil {
		log.Warn().Err(err).Str("twitter_id", order.TwitterID).Msg("Failed to insert twitter job metadata")
	}

	// Generate and publish insights
	if userInfo != nil {
		insights := parser.GenerateInsights(userInfo)
		insightsData, err := json.Marshal(insights)
		if err != nil {
			log.Error().Err(err).Str("twitter_id", order.TwitterID).Msg("Failed to marshal insights; skipping insights publish")
		} else if err := deps.Producer.Produce(ctx, cfg.RawInsightsTopic, []byte(insights.RecordID), insightsData); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleWorkOrder").Str("stage", "produce_insights").Msg("Failed to publish insights")
			atomic.AddUint64(&metrics.ProducerErrors, 1)
		} else {
			log.Info().
				Str("twitter_id", order.TwitterID).
				Str("record_id", insights.RecordID).
				Msg("Published Twitter insights")
			atomic.AddUint64(&metrics.InsightsProduced, 1)
		}
	}

	// MongoDB update handled by deferred function
	return nil
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}
