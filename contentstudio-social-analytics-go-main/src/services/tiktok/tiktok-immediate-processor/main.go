// Package main implements the TikTok immediate processor service.
// This service consumes immediate TikTok work orders from Kafka and processes them by fetching
// TikTok account data (videos, insights) and storing the results in ClickHouse for analytics.
// It handles account state management, token decryption, and real-time notifications to clients.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
)

const (
	// immediateTopic is the Kafka topic name for immediate TikTok work orders
	immediateTopic = "immediate-work-order-tiktok"
)

// ImmediateWorkOrder mirrors TikTokAccountWorkOrder but kept separate for clarity
// SyncType can be "full" or "incremental" for future extensibility.
type ImmediateWorkOrder struct {
	ID           string `json:"id"`
	WorkspaceID  string `json:"workspace_id"`
	TikTokID     string `json:"tiktok_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SyncType     string `json:"sync_type"`
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
}

// TikTokVideoFetcher is the interface for fetching TikTok videos from the API
type TikTokVideoFetcher interface {
	// FetchUserVideos fetches videos for a TikTok user with pagination support
	FetchUserVideos(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error)
	RefreshToken(ctx context.Context, refreshToken string) (*social.RefreshTokenResponse, error)
}

// TikTokPostSink is the interface for storing TikTok posts in ClickHouse
type TikTokPostSink interface {
	// BulkInsertTikTokPosts inserts multiple TikTok posts into ClickHouse
	BulkInsertTikTokPosts(ctx context.Context, posts []*chmodels.TikTokPosts) error
}

// SocialRepository is the interface for accessing social account data from MongoDB
type SocialRepository interface {
	// FindByID retrieves a social integration account by its MongoDB ID
	FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
}

// PusherNotifier is the interface for sending real-time notifications via Pusher
type PusherNotifier interface {
	// Trigger sends a real-time event to a Pusher channel
	Trigger(channel, event string, data interface{}) error
}

// EmailNotifier is the interface for sending analytics completion notifications via email
type EmailNotifier interface {
	// SendAnalyticsNotification sends an analytics notification email to a user
	SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
}

func parseTikTokDateRange(startDateStr, endDateStr string) (time.Time, time.Time, bool, error) {
	startDateStr = strings.TrimSpace(startDateStr)
	endDateStr = strings.TrimSpace(endDateStr)
	if startDateStr == "" && endDateStr == "" {
		return time.Time{}, time.Time{}, false, nil
	}
	if startDateStr == "" || endDateStr == "" {
		return time.Time{}, time.Time{}, false, fmt.Errorf("both start_date and end_date are required")
	}

	startDate, err := time.Parse("2006-01-02", startDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid start_date %q: %w", startDateStr, err)
	}
	endDate, err := time.Parse("2006-01-02", endDateStr)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("invalid end_date %q: %w", endDateStr, err)
	}
	if endDate.Before(startDate) {
		return time.Time{}, time.Time{}, false, fmt.Errorf("end_date must be on or after start_date")
	}

	return startDate.UTC(), endDate.AddDate(0, 0, 1).UTC(), true, nil
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)
	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting TikTok Immediate Processor service")

	// MongoDB connection for account information
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
	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	notifier := notification.NewService(cfg.Email, log.Logger, cfg.Email.BackendURL)
	pusherClient := notification.NewPusherClient(cfg.Pusher, log.Logger)
	consumer, err := kafka.NewConsumer(cfg.Kafka, "tiktok-immediate-processor-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka consumer")
	}
	defer consumer.Close()

	tkClient := social.NewTikTokClient(cfg.TikTok.ClientKey, cfg.TikTok.ClientSecret)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("shutdown signal received")
		cancel()
	}()

	handler := func(ctx context.Context, _ string, _ []byte, value []byte) error {
		var wo ImmediateWorkOrder
		if err := json.Unmarshal(value, &wo); err != nil {
			log.Error().
				Err(err).
				Str("error_message", err.Error()).
				Str("function", "handler").
				Str("stage", "unmarshal_work_order").
				Msg("Failed to unmarshal work order")
			return nil // skip message
		}
		if err := processAccount(ctx, tkClient, sink, mongoRepo, notifier, pusherClient, wo, cfg.DecryptionKey, log); err != nil {
			log.Error().
				Err(err).
				Str("error_message", err.Error()).
				Str("account_id", wo.ID).
				Str("tiktok_id", wo.TikTokID).
				Str("workspace_id", wo.WorkspaceID).
				Str("function", "handler").
				Str("stage", "process_account").
				Msg("Failed to process TikTok account")
		}
		return nil
	}

	log.Info().Msg("Consuming " + immediateTopic)
	if err := consumer.Consume(ctx, []string{immediateTopic}, handler); err != nil && err != context.Canceled {
		log.Error().
			Err(err).
			Str("error_message", err.Error()).
			Str("function", "main").
			Str("stage", "kafka_consume").
			Msg("Consumer error")
	}

	log.Info().Msg("TikTok Immediate Processor stopped")
}

// processAccount is the unexported wrapper that adapts concrete types to the generic ProcessAccount function.
// It bridges between the main function's concrete dependencies and the testable ProcessAccount interface.
func processAccount(ctx context.Context, tkClient *social.TikTokClient, sink *conversions.ClickHouseSink, mongoRepo mongodb.UnifiedSocialRepository, notifier *notification.Service, pusherClient *notification.PusherClient, wo ImmediateWorkOrder, decryptionKey string, log *logger.Logger) (err error) {
	return ProcessAccount(ctx, tkClient, sink, mongoRepo, notifier, pusherClient, wo, decryptionKey, log)
}

// ProcessAccount processes a TikTok account and stores posts and insights in ClickHouse.
// It handles account state management, token decryption, user info fetching, video fetching,
// data parsing, ClickHouse insertion, and sending notifications to users and Pusher clients.
// The function is testable because it accepts interfaces rather than concrete types.
func ProcessAccount(ctx context.Context, tkClient TikTokVideoFetcher, sink TikTokPostSink, mongoRepo SocialRepository, notifier EmailNotifier, pusherClient PusherNotifier, wo ImmediateWorkOrder, decryptionKey string, log *logger.Logger) (err error) {
	op := log.Operation("ProcessTikTokAccount").
		WithField("workspace_id", wo.WorkspaceID).
		WithField("tiktok_id", wo.TikTokID).
		WithField("platform_identifier", wo.TikTokID).
		WithField("start_date", wo.StartDate).
		WithField("end_date", wo.EndDate).
		WithField("sync_type", wo.SyncType).
		WithSentryTags(map[string]string{
			"workspace_id": wo.WorkspaceID,
			"tiktok_id":    wo.TikTokID,
			"sync_type":    wo.SyncType,
		})
	op.Start("processing tiktok work order")
	var chPosts []*chmodels.TikTokPosts
	defer func() {
		op.WithField("parsed_posts", len(chPosts)).
			Complete(err, "")
	}()

	// 1. Fetch account from MongoDB if ID is provided
	var account *mongomodels.SocialIntegration
	var originalState string
	if wo.ID != "" {
		log.Info().
			Str("account_id", wo.ID).
			Msg("Fetching account from MongoDB")

		accountID, err := primitive.ObjectIDFromHex(wo.ID)
		if err != nil {
			log.Warn().Err(err).Str("account_id", wo.ID).Msg("Invalid account ID, continuing without MongoDB data")
		} else {
			account, err = mongoRepo.FindByID(ctx, accountID)
			if err != nil {
				log.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to fetch account from MongoDB, continuing without")
			} else if account != nil {
				if mongodb.HasProcessingErrorMeta(account.MetaData) {
					if clearRepo, ok := mongoRepo.(interface {
						ClearProcessingError(ctx context.Context, id primitive.ObjectID) error
					}); ok {
						if clearErr := clearRepo.ClearProcessingError(ctx, accountID); clearErr != nil {
							log.Warn().Err(clearErr).Str("account_id", wo.ID).Msg("Failed to clear stale processing error before retry")
						}
					}
				}
				originalState = account.State
				log.Info().
					Str("account_id", wo.ID).
					Str("platform_identifier", account.PlatformIdentifier).
					Str("state", account.State).
					Msg("Fetched account from MongoDB")
			}
		}
	}

	defer func() {
		if wo.ID == "" {
			return
		}
		accountID, parseErr := primitive.ObjectIDFromHex(wo.ID)
		if parseErr != nil {
			return
		}
		if err != nil {
			if recordRepo, ok := mongoRepo.(interface {
				RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error
			}); ok {
				if recordErr := recordRepo.RecordProcessingError(ctx, accountID, err.Error()); recordErr != nil {
					log.Warn().Err(recordErr).Str("account_id", wo.ID).Msg("Failed to record processing error")
				}
			}
			return
		}
		if clearRepo, ok := mongoRepo.(interface {
			ClearProcessingError(ctx context.Context, id primitive.ObjectID) error
		}); ok {
			if clearErr := clearRepo.ClearProcessingError(ctx, accountID); clearErr != nil {
				log.Warn().Err(clearErr).Str("account_id", wo.ID).Msg("Failed to clear processing error")
			}
		}
	}()

	accessToken := wo.AccessToken
	if dec, err := crypto.DecryptToken(accessToken, decryptionKey); err == nil {
		accessToken = dec
	}
	refreshToken := wo.RefreshToken
	if dec, err := crypto.DecryptToken(refreshToken, decryptionKey); err == nil {
		refreshToken = dec
	}

	// Keep immediate processor behavior consistent with fetcher:
	// attempt refresh first and use refreshed token when available.
	if refreshToken != "" {
		tokenResp, refreshErr := tkClient.RefreshToken(ctx, refreshToken)
		if refreshErr == nil && tokenResp != nil && tokenResp.AccessToken != "" {
			log.Info().
				Str("tiktok_id", wo.TikTokID).
				Msg("Using refreshed TikTok access token")
			accessToken = tokenResp.AccessToken
		} else {
			log.Warn().
				Err(refreshErr).
				Str("tiktok_id", wo.TikTokID).
				Msg("Failed to refresh TikTok token, falling back to existing access token")
		}
	} else {
		log.Info().
			Str("tiktok_id", wo.TikTokID).
			Msg("Refresh token missing, using existing TikTok access token")
	}

	startTime, endTime, hasRequestedRange, dateRangeErr := parseTikTokDateRange(wo.StartDate, wo.EndDate)
	if dateRangeErr != nil {
		log.Warn().
			Err(dateRangeErr).
			Str("start_date", wo.StartDate).
			Str("end_date", wo.EndDate).
			Msg("Invalid TikTok date range")
		return dateRangeErr
	}

	cursor := 0
	reachedCutoff := false
	for {
		rawData, nextCursor, err := tkClient.FetchUserVideos(ctx, wo.TikTokID, accessToken, cursor, 20)
		if err != nil {
			log.Error().Err(err).Msg("fetch videos failed")
			break
		}
		// rawData expected to be JSON array of video items
		var items []struct {
			ID         string `json:"id"`
			Desc       string `json:"desc"`
			CreateTime int64  `json:"create_time"`
			Stats      struct {
				DiggCount    int64 `json:"digg_count"`
				CommentCount int64 `json:"comment_count"`
				ShareCount   int64 `json:"share_count"`
				PlayCount    int64 `json:"play_count"`
			} `json:"stats"`
		}
		if err := json.Unmarshal(rawData, &items); err != nil {
			log.Error().Err(err).Msg("unmarshal items")
			break
		}
		for _, itm := range items {
			videoTime := time.Unix(itm.CreateTime, 0).UTC()
			if hasRequestedRange {
				if videoTime.Before(startTime) {
					reachedCutoff = true
					break
				}
				if !videoTime.Before(endTime) {
					continue
				}
			}
			parsed := kafkamodels.ParsedTikTokPost{
				ID:              itm.ID,
				WorkspaceID:     wo.WorkspaceID,
				PostDescription: itm.Desc,
				LikeCount:       itm.Stats.DiggCount,
				CommentCount:    itm.Stats.CommentCount,
				ShareCount:      itm.Stats.ShareCount,
				ViewCount:       itm.Stats.PlayCount,
				CreateTime:      itm.CreateTime,
			}
			cp := conversions.ConvertTikTokPost(&parsed)
			if cp != nil {
				chPosts = append(chPosts, cp)
			}
		}
		if reachedCutoff || nextCursor == 0 {
			break
		}
		cursor = int(nextCursor)
		// avoid tight loop
		time.Sleep(300 * time.Millisecond)
	}

	if len(chPosts) == 0 {
		log.Info().Msg("no posts parsed")
		return nil
	}
	if err := sink.BulkInsertTikTokPosts(ctx, chPosts); err != nil {
		return err
	}
	log.Info().Int("inserted_posts", len(chPosts)).Msg("account processed")

	// Send notifications if account was fetched from MongoDB
	if account != nil {
		userID := account.GetUserIDHex()
		workspaceID := account.GetWorkspaceIDHex()
		if workspaceID == "" {
			workspaceID = wo.WorkspaceID
		}

		// Send Pusher notification for real-time UI updates
		SendPusherNotification(pusherClient, account, workspaceID, log)

		// Send email notification only for newly added accounts
		if originalState == "Added" {
			SendEmailNotification(notifier, userID, workspaceID, wo.TikTokID, account.PlatformName, log)
		}

		log.Info().
			Str("tiktok_id", wo.TikTokID).
			Str("workspace_id", workspaceID).
			Bool("email_sent", originalState == "Added").
			Msg("Notifications sent")
	}

	return nil
}

// SendPusherNotification sends a real-time notification via Pusher when TikTok analytics are completed.
// It notifies connected clients about the account processing state change for real-time UI updates.
func SendPusherNotification(pusherClient PusherNotifier, account *mongomodels.SocialIntegration, workspaceID string, log *logger.Logger) {
	if pusherClient == nil {
		return
	}

	tiktokID := account.PlatformIdentifier
	if tiktokID == "" {
		// Fallback to the TikTok ID from work order if PlatformIdentifier is empty
		tiktokID = "unknown"
	}

	channel := fmt.Sprintf("tt-analytics-channel-%s-%s", workspaceID, tiktokID)
	event := fmt.Sprintf("syncing-%s-%s", workspaceID, tiktokID)

	data := map[string]interface{}{
		"state":                     "Processed",
		"account":                   tiktokID,
		"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02"),
	}

	if err := pusherClient.Trigger(channel, event, data); err != nil {
		log.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("channel", channel).
			Str("event", event).
			Str("function", "SendPusherNotification").
			Msg("Failed to send Pusher notification")
	} else {
		log.Debug().
			Str("channel", channel).
			Str("event", event).
			Msg("Sent Pusher notification")
	}
}

// sendPusherNotification is the unexported wrapper that adapts concrete Pusher client to the interface.
// It provides backward compatibility and bridges concrete and interface-based implementations.
func sendPusherNotification(pusherClient *notification.PusherClient, account *mongomodels.SocialIntegration, workspaceID string, log *logger.Logger) {
	SendPusherNotification(pusherClient, account, workspaceID, log)
}

// SendEmailNotification sends an analytics completion notification email via the backend API.
// It notifies users via email when TikTok analytics processing is complete, but only for newly added accounts.
func SendEmailNotification(notifier EmailNotifier, userID, workspaceID, accountID, accountName string, log *logger.Logger) {
	if notifier == nil {
		return
	}

	// Send analytics notification to backend API
	err := notifier.SendAnalyticsNotification(
		userID,      // userID
		workspaceID, // workspaceID
		"tiktok",    // platform
		accountID,   // accountID
		accountName, // accountName
		false,       // isCompetitor (false for social accounts)
	)

	if err != nil {
		log.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("user_id", userID).
			Str("workspace_id", workspaceID).
			Str("account_id", accountID).
			Str("function", "SendEmailNotification").
			Msg("Failed to send analytics notification to backend")
	} else {
		log.Info().
			Str("user_id", userID).
			Str("workspace_id", workspaceID).
			Str("account_id", accountID).
			Msg("Analytics notification sent to backend successfully")
	}
}

// sendEmailNotification is the unexported wrapper that adapts concrete notification service to the interface.
// It provides backward compatibility and bridges concrete and interface-based implementations.
func sendEmailNotification(notifier *notification.Service, userID, workspaceID, accountID, accountName string, log *logger.Logger) {
	SendEmailNotification(notifier, userID, workspaceID, accountID, accountName, log)
}
