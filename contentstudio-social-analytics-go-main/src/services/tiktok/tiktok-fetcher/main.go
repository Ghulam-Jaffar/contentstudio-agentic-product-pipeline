package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/semaphore"

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
)

const (
	workOrdersTopic  = "work-order-tiktok-batch"
	rawPostsTopic    = "raw-tiktok-posts"
	rawInsightsTopic = "raw-tiktok-insights"
	maxWorkers       = 10
	workChanSize     = 200

	// maxConcurrentAccounts is the max number of TikTok accounts processed simultaneously.
	maxConcurrentAccounts = 50

	// idleTimeout is the duration after which the service will shutdown
	// if no new messages are received. This allows the service to exit
	// gracefully after batch processing is complete.
	idleTimeout       = 5 * time.Minute
	idleCheckInterval = 30 * time.Second
)

// Per-account concurrency guard - ensures we don't run multiple pipelines for the same TikTok account at once
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

// ExpiresDateObject represents the MongoDB object format for expires_in/refresh_expires_in
// This matches the serializeUsing() format from the Laravel backend
type ExpiresDateObject struct {
	Date     string `bson:"date" json:"date"`
	Timezone string `bson:"timezone,omitempty" json:"timezone,omitempty"`
}

// convertExpiresInToObject converts expires_in seconds to MongoDB object format
// If the value is already an object, it returns it as-is
func convertExpiresInToObject(expiresIn interface{}) interface{} {
	switch v := expiresIn.(type) {
	case int:
		// Convert seconds to future timestamp
		expiresAt := time.Now().Add(time.Duration(v) * time.Second)
		return ExpiresDateObject{
			Date:     expiresAt.Format("2006-01-02 15:04:05"),
			Timezone: "UTC",
		}
	case float64:
		// Convert seconds to future timestamp
		expiresAt := time.Now().Add(time.Duration(int64(v)) * time.Second)
		return ExpiresDateObject{
			Date:     expiresAt.Format("2006-01-02 15:04:05"),
			Timezone: "UTC",
		}
	default:
		// Already an object or nil, return as-is
		return expiresIn
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)
	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("max_concurrent_accounts", maxConcurrentAccounts).
		Dur("idle_timeout", idleTimeout).
		Msg("Starting TikTok Fetcher service")

	consumer, err := kafka2.NewConsumer(cfg.Kafka, "tiktok-fetcher-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka consumer")
	}
	defer consumer.Close()

	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka producer")
	}
	defer producer.Close()

	// MongoDB connection for updating account state
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
	log.Info().Msg("Connected to MongoDB")

	mongoRepo := mongodb.NewUnifiedSocialRepository(mongoClient.Database(cfg.Mongo.Database), log.Logger)

	// TikTok client needs client credentials for token refresh
	tkClient := social.NewTikTokClient(cfg.TikTok.ClientKey, cfg.TikTok.ClientSecret)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	accountSem := semaphore.NewWeighted(maxConcurrentAccounts)
	var dispatchWg sync.WaitGroup
	var totalProcessed, totalFailed int64

	// Track last message time for idle timeout
	var lastMessageTime int64 = time.Now().UnixNano()

	// Consume batch work orders
	go func() {
		topics := []string{workOrdersTopic}

		if err := consumer.ConsumeWithAck(ctx, topics, func(ctx context.Context, topic string, key, value []byte, ack func()) error {
			// Update last message time
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())

			var batch kafkamodels.TikTokBatchWorkOrder
			if err := json.Unmarshal(value, &batch); err != nil {
				log.Error().Err(err).Str("function", "main").Str("stage", "unmarshal_batch_work_order").Msg("failed to unmarshal batch work order")
				ack() // unrecoverable bad message — advance offset so it's not redelivered
				return nil
			}

			log.Info().
				Str("batch_id", batch.BatchID).
				Int("account_count", len(batch.Accounts)).
				Msg("processing batch work order")

			// One ack guards the whole batch message; use a WaitGroup so it fires
			// only after every account in the batch has been fully processed.
			var batchWg sync.WaitGroup
			var batchProcessed, batchFailed int64
			total := len(batch.Accounts)
			batchID := batch.BatchID
			for _, account := range batch.Accounts {
				acc := account
				accountData, err := json.Marshal(acc)
				if err != nil {
					log.Error().Err(err).Str("tiktok_id", acc.TikTokID).Msg("Failed to marshal account work order; skipping")
					atomic.AddInt64(&batchFailed, 1)
					continue
				}
				batchWg.Add(1)
				dispatchWg.Add(1)
				go func() {
					defer dispatchWg.Done()
					defer batchWg.Done()
					if err := accountSem.Acquire(ctx, 1); err != nil {
						return
					}
					defer accountSem.Release(1)
					if err := HandleWorkOrder(ctx, []byte(acc.TikTokID), accountData, tkClient, producer, mongoRepo, cfg.DecryptionKey, log); err != nil {
						log.Error().Err(err).Str("tiktok_id", acc.TikTokID).Str("function", "main").Str("stage", "handle_work_order").Msg("work order failed")
						atomic.AddInt64(&batchFailed, 1)
					} else {
						atomic.AddInt64(&batchProcessed, 1)
					}
				}()
			}
			go func() {
				batchWg.Wait()
				p := atomic.LoadInt64(&batchProcessed)
				f := atomic.LoadInt64(&batchFailed)
				atomic.AddInt64(&totalProcessed, p)
				atomic.AddInt64(&totalFailed, f)
				log.Info().
					Str("batch_id", batchID).
					Int("total", total).
					Int64("processed", p).
					Int64("failed", f).
					Msg("Batch processing complete")
				ack()
			}()
			return nil
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_work_orders").Msg("consumer error")
		}
	}()

	log.Info().
		Int("max_concurrent_accounts", maxConcurrentAccounts).
		Msg("TikTok Fetcher service is running")

	<-sigChan
	log.Info().Msg("Shutdown signal received")

	cancel()
	dispatchWg.Wait()
	log.Info().
		Int64("total_processed", atomic.LoadInt64(&totalProcessed)).
		Int64("total_failed", atomic.LoadInt64(&totalFailed)).
		Msg("TikTok Fetcher service stopped")
}

// HandleWorkOrder processes a single TikTok work order
func HandleWorkOrder(ctx context.Context, key, value []byte, tkClient social.TikTokAPI, producer kafka2.Producer, mongoRepo mongodb.UnifiedSocialRepository, decryptionKey string, log *logger.Logger) (err error) {
	var order kafkamodels.TikTokAccountWorkOrder
	if err := json.Unmarshal(value, &order); err != nil {
		return err
	}

	// Per-account concurrency gate - prevent multiple simultaneous fetches for same TikTok account
	sem := semForAccount(order.TikTokID, 1)
	if err := sem.Acquire(ctx, 1); err != nil {
		log.Warn().Err(err).Str("tiktok_id", order.TikTokID).Msg("Failed to acquire semaphore")
		return nil
	}
	defer sem.Release(1)

	log.Info().Str("tiktok_id", order.TikTokID).Str("sync_type", order.SyncType).Msg("Processing TikTok account")

	// Validate scopes
	if !parsing.ValidateScopes(order.Scope) {
		validationErr := errors.New("invalid scopes for TikTok analytics")
		log.Error().Err(validationErr).Str("error_message", validationErr.Error()).Str("tiktok_id", order.TikTokID).Str("scope", order.Scope).Str("function", "HandleWorkOrder").Str("stage", "validate_scopes").Msg(validationErr.Error())
		if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
			mongoRepo.RecordProcessingError(context.Background(), accountID, validationErr.Error())
		}
		return nil
	}

	// Decrypt access token if needed
	accessToken := order.AccessToken
	if dec, err := crypto.DecryptToken(accessToken, decryptionKey); err == nil {
		accessToken = dec
	}

	refreshToken := order.RefreshToken
	if dec, err := crypto.DecryptToken(refreshToken, decryptionKey); err == nil {
		refreshToken = dec
	}

	// Try to refresh token first (like Python implementation)
	if refreshToken != "" {
		tokenResp, err := tkClient.RefreshToken(ctx, refreshToken)
		if err == nil && tokenResp != nil && tokenResp.AccessToken != "" {
			log.Info().Str("tiktok_id", order.TikTokID).Msg("Successfully refreshed TikTok token")
			accessToken = tokenResp.AccessToken
			refreshToken = tokenResp.RefreshToken

			// Convert expires_in and refresh_expires_in to MongoDB object format
			expiresInObj := convertExpiresInToObject(tokenResp.ExpiresIn)
			refreshExpiresInObj := convertExpiresInToObject(tokenResp.RefreshExpiresIn)

			// Log the refreshed token details
			log.Debug().
				Interface("expires_in", expiresInObj).
				Interface("refresh_expires_in", refreshExpiresInObj).
				Str("tiktok_id", order.TikTokID).
				Msg("Token refresh details - should be persisted to MongoDB by backend")
		} else {
			log.Warn().Err(err).Str("tiktok_id", order.TikTokID).Msg("Failed to refresh token, using existing")
		}
	}

	// Fetch user info first
	userInfoRaw, err := tkClient.FetchUserInfo(ctx, accessToken)
	if err != nil {
		if social.IsExpectedCompetitorErrorTikTok(err) {
			log.Warn().Err(err).Str("tiktok_id", order.TikTokID).Msg("Failed to fetch user info (expected token/permission error)")
		} else {
			log.Error().Err(err).Str("error_message", err.Error()).Str("tiktok_id", order.TikTokID).Str("function", "HandleWorkOrder").Str("stage", "fetch_user_info").Msg("Failed to fetch user info")
		}
		if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
			mongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
		}
		return err
	}

	parser := parsing.NewTikTokParser()
	userInfo, err := parser.ParseUserInfo(userInfoRaw)
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "HandleWorkOrder").Str("stage", "parse_user_info").Msg("Failed to parse user info")
		if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
			mongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
		}
		return err
	}

	// Determine fetch limit and cutoff time based on sync type
	// For incremental sync (cron): last 14 days, max 999 videos
	// For full sync: no time limit, no video limit
	// Note: These limits match the fallback option in immediate processor
	const maxVideosCron = 999
	const daysBackIncremental = 14

	var maxVideos int
	var cutoffTime time.Time
	if strings.ToLower(order.SyncType) == "incremental" {
		maxVideos = maxVideosCron
		cutoffTime = time.Now().AddDate(0, 0, -daysBackIncremental) // 14 days ago
	} else {
		maxVideos = -1           // No limit for full sync
		cutoffTime = time.Time{} // No cutoff for full sync
	}

	// Fetch videos and aggregate stats
	var cursor int64 = 0
	videosProcessed := 0
	var totalViews, totalLikes, totalComments, totalShares int64

	for {
		// Check context before each API call
		select {
		case <-ctx.Done():
			log.Warn().Str("tiktok_id", order.TikTokID).Msg("Context cancelled during video fetch")
			return ctx.Err()
		default:
		}

		// Fetch batch of videos
		videosRaw, nextCursor, hasMore, err := tkClient.FetchVideoList(ctx, accessToken, cursor, 20)
		if err != nil {
			if social.IsExpectedCompetitorErrorTikTok(err) {
				log.Warn().Err(err).Int64("cursor", cursor).Msg("Failed to fetch videos (expected token/permission error)")
			} else {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "HandleWorkOrder").Str("stage", "fetch_videos").Msg("Failed to fetch videos")
			}
			if social.IsAuthError(err) {
				if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
					mongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
				}
				return err
			}
			break
		}

		// Parse videos array
		var videos []json.RawMessage
		if err := json.Unmarshal(videosRaw, &videos); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "HandleWorkOrder").Str("stage", "unmarshal_videos").Msg("Failed to unmarshal videos array")
			break
		}

		// Process each video with open_id from userInfo (which comes from TikTok API)
		for _, videoRaw := range videos {
			parsedPost, err := parser.ParseVideo(videoRaw, userInfo, userInfo.OpenID)
			if err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "HandleWorkOrder").Str("stage", "parse_video").Msg("Failed to parse video")
				continue
			}

			// Set CreatedAt field
			parsedPost.CreatedAt = time.Unix(parsedPost.CreateTime, 0)

			// Check cutoff time for incremental sync
			if !cutoffTime.IsZero() && parsedPost.CreatedAt.Before(cutoffTime) {
				log.Debug().
					Time("video_created", parsedPost.CreatedAt).
					Time("cutoff", cutoffTime).
					Str("tiktok_id", order.TikTokID).
					Msg("Video before cutoff time, stopping fetch")
				goto endVideoLoop
			}

			// Aggregate stats for insights
			totalViews += parsedPost.ViewCount
			totalLikes += parsedPost.LikeCount
			totalComments += parsedPost.CommentCount
			totalShares += parsedPost.ShareCount

			// Wrap in RawTikTokPost format to match analytics sink expectations
			// Note: We need to store the parsed post as JSON in the Data field
			parsedPostJSON, err := json.Marshal(parsedPost)
			if err != nil {
				log.Error().Err(err).Str("post_id", parsedPost.ID).Msg("Failed to marshal parsed post; skipping")
				continue
			}
			rawPost := kafkamodels.RawTikTokPost{
				WorkspaceID: order.WorkspaceID,
				TikTokID:    order.TikTokID,
				Data:        parsedPostJSON,
			}
			postData, err := json.Marshal(rawPost)
			if err != nil {
				log.Error().Err(err).Str("post_id", parsedPost.ID).Msg("Failed to marshal raw post; skipping")
				continue
			}

			if err := producer.Produce(ctx, rawPostsTopic, []byte(parsedPost.ID), postData); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("post_id", parsedPost.ID).Str("function", "HandleWorkOrder").Str("stage", "produce_post").Msg("Failed to publish post")
			}

			videosProcessed++
			if maxVideos > 0 && videosProcessed >= maxVideos {
				log.Debug().Int("videos_processed", videosProcessed).Int("max_videos", maxVideos).Msg("Reached max video limit")
				break
			}
		}
	endVideoLoop:

		// Check if we should continue
		if !hasMore || nextCursor == 0 {
			break
		}
		if maxVideos > 0 && videosProcessed >= maxVideos {
			break
		}
		cursor = nextCursor
	}

	log.Info().
		Str("tiktok_id", order.TikTokID).
		Int("videos_processed", videosProcessed).
		Msg("Completed fetching videos")

	// Generate and publish insights with open_id from userInfo
	insights := parser.GenerateInsights(userInfo, userInfo.OpenID, userInfo.OpenID, totalViews, totalLikes, totalComments, totalShares)

	insightsData, err := json.Marshal(insights)
	if err != nil {
		log.Error().Err(err).Str("tiktok_id", order.TikTokID).Msg("Failed to marshal insights; skipping insights publish")
	} else if err := producer.Produce(ctx, rawInsightsTopic, []byte(insights.RecordID), insightsData); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "HandleWorkOrder").Str("stage", "produce_insights").Msg("Failed to publish insights")
	} else {
		log.Info().
			Str("tiktok_id", order.TikTokID).
			Str("record_id", insights.RecordID).
			Msg("Published TikTok insights")
	}

	// Update MongoDB state to "Processed" after successful completion
	if order.ID != "" {
		accountID, err := primitive.ObjectIDFromHex(order.ID)
		if err != nil {
			log.Warn().Err(err).Str("account_id", order.ID).Msg("Invalid account ID, skipping state update")
		} else {
			updates := bson.M{
				"state":                     mongomodels.StateProcessed,
				"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02 15:04:05"),
			}
			if err := mongoRepo.Update(ctx, accountID, updates); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("account_id", order.ID).Str("function", "HandleWorkOrder").Str("stage", "update_account_state").Msg("Failed to update account state to Processed")
			} else {
				mongoRepo.ClearProcessingError(context.Background(), accountID)
				log.Info().Str("account_id", order.ID).Str("state", mongomodels.StateProcessed).Msg("Updated account state to Processed")
			}
		}
	}

	return nil
}
