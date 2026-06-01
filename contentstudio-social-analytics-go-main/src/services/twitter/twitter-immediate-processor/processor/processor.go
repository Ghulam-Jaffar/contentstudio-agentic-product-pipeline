package processor

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

// ImmediateWorkOrder represents a Twitter immediate processing work order.
type ImmediateWorkOrder struct {
	ID               string `json:"id"`
	WorkspaceID      string `json:"workspace_id"`
	TwitterID        string `json:"twitter_id"`
	OAuthToken       string `json:"oauth_token"`
	OAuthTokenSecret string `json:"oauth_token_secret"`
	NTweets          int    `json:"n_tweets,omitempty"`
	APIKey           string `json:"api_key"`
	APISecret        string `json:"api_secret"`
	AppName          string `json:"app_name"`
	AppID            string `json:"app_id"`
	ExecutedBy       string `json:"executed_by"`
	SyncType         string `json:"sync_type"`
}

// Processor handles Twitter immediate account processing.
type Processor struct {
	mongoRepo    mongodb.UnifiedSocialRepository
	sink         *conversions.ClickHouseSink
	notifier     *notification.Service
	pusherClient *notification.PusherClient
	log          *logger.Logger
	cfg          *config.Config
}

// New creates a new Twitter Processor with all dependencies.
func New(
	mongoRepo mongodb.UnifiedSocialRepository,
	sink *conversions.ClickHouseSink,
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	log *logger.Logger,
	cfg *config.Config,
) *Processor {
	return &Processor{
		mongoRepo:    mongoRepo,
		sink:         sink,
		notifier:     notifier,
		pusherClient: pusherClient,
		log:          log,
		cfg:          cfg,
	}
}

// ProcessAccount implements ProcessorInterface using the concrete dependencies.
func (p *Processor) ProcessAccount(ctx context.Context, wo ImmediateWorkOrder) error {
	twClient := social.NewTwitterClient(p.cfg.Twitter.ConsumerKey, p.cfg.Twitter.ConsumerSecret)
	if wo.APIKey != "" && wo.APISecret != "" {
		twClient = social.NewTwitterClient(wo.APIKey, wo.APISecret)
	}
	return ProcessAccount(ctx, twClient, p.sink, p.mongoRepo, p.notifier, p.pusherClient, wo, p.cfg.DecryptionKey, p.log)
}

// ProcessAccount processes a Twitter account and stores posts and insights in ClickHouse.
// This is the core function that can be called directly with injected dependencies for testing.
func ProcessAccount(
	ctx context.Context,
	twClient TwitterTweetFetcher,
	sink TwitterPostSink,
	mongoRepo SocialRepository,
	notifier NotifierInterface,
	pusherClient PusherClientInterface,
	wo ImmediateWorkOrder,
	decryptionKey string,
	log *logger.Logger,
) (err error) {
	op := log.Operation("ProcessTwitterAccount").
		WithField("workspace_id", wo.WorkspaceID).
		WithField("twitter_id", wo.TwitterID).
		WithField("sync_type", wo.SyncType).
		WithSentryTags(map[string]string{
			"workspace_id": wo.WorkspaceID,
			"twitter_id":   wo.TwitterID,
			"sync_type":    wo.SyncType,
		})
	op.Start("processing twitter work order")
	log.Info().
		Str("twitter_id", wo.TwitterID).
		Str("platform_identifier", wo.TwitterID).
		Int("n_tweets", wo.NTweets).
		Str("sync_type", wo.SyncType).
		Str("workspace_id", wo.WorkspaceID).
		Msg("Starting Twitter immediate work order processing")
	var chPosts []*chmodels.TwitterPosts
	defer func() {
		op.WithField("parsed_posts", len(chPosts)).
			Complete(nil, "")
	}()

	// Deferred MongoDB success update.
	defer func() {
		if wo.ID != "" {
			accountID, parseErr := primitive.ObjectIDFromHex(wo.ID)
			if parseErr != nil {
				log.Warn().Err(parseErr).Str("account_id", wo.ID).Msg("Invalid account ID for MongoDB update")
				return
			}
			if err != nil {
				if recordErr := mongoRepo.RecordProcessingError(ctx, accountID, err.Error()); recordErr != nil {
					log.Warn().Err(recordErr).Str("account_id", wo.ID).Msg("Failed to record processing error")
				}
				return
			}

			updates := bson.M{
				"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02 15:04:05"),
			}

			updates["state"] = mongomodels.StateProcessed

			if updateErr := mongoRepo.Update(ctx, accountID, updates); updateErr != nil {
				log.Warn().Err(updateErr).Str("error_message", updateErr.Error()).Str("account_id", wo.ID).Str("function", "ProcessAccount").Str("stage", "update_mongo_state").Msg("Failed to update account state and last_analytics_updated_at in MongoDB")
			} else {
				log.Info().
					Str("account_id", wo.ID).
					Str("state", updates["state"].(string)).
					Str("last_analytics_updated_at", updates["last_analytics_updated_at"].(string)).
					Msg("Updated account state and last_analytics_updated_at in MongoDB")
			}
			if clearErr := mongoRepo.ClearProcessingError(ctx, accountID); clearErr != nil {
				log.Warn().Err(clearErr).Str("account_id", wo.ID).Msg("Failed to clear processing error")
			}
		}
	}()

	// 1. Fetch account from MongoDB if ID is provided
	var account *mongomodels.SocialIntegration
	var originalState string
	if wo.ID != "" {
		accountID, parseErr := primitive.ObjectIDFromHex(wo.ID)
		if parseErr != nil {
			log.Warn().Err(parseErr).Str("account_id", wo.ID).Msg("Invalid account ID, continuing without MongoDB data")
		} else {
			account, err = mongoRepo.FindByID(ctx, accountID)
			if err != nil {
				log.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to fetch account from MongoDB, continuing without")
				err = nil // Reset - non-fatal
			} else if account != nil {
				if mongodb.HasProcessingErrorMeta(account.MetaData) {
					if clearErr := mongoRepo.ClearProcessingError(ctx, accountID); clearErr != nil {
						log.Warn().Err(clearErr).Str("account_id", wo.ID).Msg("Failed to clear stale processing error before retry")
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

	// 2. Decrypt tokens
	oauthToken := wo.OAuthToken
	if dec, decErr := crypto.DecryptToken(oauthToken, decryptionKey); decErr == nil {
		oauthToken = dec
	}
	oauthTokenSecret := wo.OAuthTokenSecret
	if dec, decErr := crypto.DecryptToken(oauthTokenSecret, decryptionKey); decErr == nil {
		oauthTokenSecret = dec
	}

	// 3. Fetch user info
	log.Info().Str("twitter_id", wo.TwitterID).Msg("Calling Twitter user info endpoint")
	userResp, err := twClient.FetchUserInfo(ctx, []string{wo.TwitterID}, oauthToken, oauthTokenSecret)
	if err != nil {
		log.Warn().Err(err).Str("error_message", err.Error()).Str("twitter_id", wo.TwitterID).Str("function", "ProcessAccount").Str("stage", "fetch_user_info").Msg("Failed to fetch user info")
		return fmt.Errorf("ProcessAccount: failed to fetch user info: %w", err)
	}
	var userInfo *social.TwitterUser
	userRecords := 0
	if userResp != nil && len(userResp.Data) > 0 {
		userInfo = &userResp.Data[0]
		userRecords = len(userResp.Data)
	} else if userResp != nil {
		userRecords = len(userResp.Data)
	}
	log.Info().
		Str("twitter_id", wo.TwitterID).
		Int("user_records", userRecords).
		Bool("user_info_fetched", userInfo != nil).
		Msg("Completed Twitter user info endpoint")

	parser := parsing.NewTwitterParser()

	// 4. Fetch tweets with pagination
	var paginationToken string
	rawTweetsFetched := 0
	requestedPostCount := wo.NTweets
	remainingTweets := wo.NTweets
	if remainingTweets <= 0 {
		requestedPostCount = 30
		remainingTweets = 30
	}
	usePagination := requestedPostCount > 100
	log.Info().
		Str("twitter_id", wo.TwitterID).
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
		log.Info().
			Str("twitter_id", wo.TwitterID).
			Int("page_number", pageNumber).
			Int("page_size", pageSize).
			Int("remaining_tweets_before_call", remainingTweets).
			Bool("has_pagination_token", paginationToken != "").
			Msg("Calling Twitter tweets endpoint")

		resp, fetchErr := twClient.FetchUserTweets(ctx, wo.TwitterID, oauthToken, oauthTokenSecret, pageSize, paginationToken)
		if fetchErr != nil {
			log.Warn().Err(fetchErr).Str("error_message", fetchErr.Error()).Str("function", "ProcessAccount").Str("stage", "fetch_tweets").Msg("fetch tweets failed (continuing with collected tweets)")
			logger.CaptureException(fetchErr, map[string]string{
				"platform":  "twitter",
				"component": "immediate-processor",
				"stage":     "fetch_tweets",
			}, nil)
			break
		}

		if resp == nil || len(resp.Data) == 0 {
			log.Info().
				Str("twitter_id", wo.TwitterID).
				Int("page_number", pageNumber).
				Msg("Twitter tweets endpoint returned empty page")
			break
		}
		rawTweetsFetched += len(resp.Data)
		log.Info().
			Str("twitter_id", wo.TwitterID).
			Int("page_number", pageNumber).
			Int("raw_tweets_in_page", len(resp.Data)).
			Int("raw_tweets_fetched_total", rawTweetsFetched).
			Msg("Processed Twitter tweets endpoint page")

		for _, tweet := range resp.Data {
			parsed := parser.ParseTweet(tweet, userInfo, resp.Includes)
			if parsed == nil {
				continue
			}
			cp := conversions.ConvertTwitterPost(parsed)
			if cp != nil {
				chPosts = append(chPosts, cp)
			}
			remainingTweets--
			if remainingTweets <= 0 {
				break
			}
		}

		if remainingTweets <= 0 {
			break
		}
		if !usePagination {
			break
		}
		if resp.Meta == nil || resp.Meta.NextToken == "" {
			log.Info().
				Str("twitter_id", wo.TwitterID).
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
		Str("twitter_id", wo.TwitterID).
		Int("tweets_processed", len(chPosts)).
		Int("raw_tweets_fetched", rawTweetsFetched).
		Msg("Completed fetching tweets")

	// 5. Insert posts
	if len(chPosts) > 0 {
		if err = sink.BulkInsertTwitterPosts(ctx, chPosts); err != nil {
			return fmt.Errorf("ProcessAccount: failed to insert posts: %w", err)
		}
		log.Info().Int("inserted_posts", len(chPosts)).Msg("Posts inserted")
	} else {
		log.Info().Msg("No posts parsed")
	}

	// 6. Generate and insert insights
	if userInfo != nil {
		insightsParsed := parser.GenerateInsights(userInfo)
		if insightsParsed != nil {
			chInsight := conversions.ConvertTwitterInsights(insightsParsed)
			if chInsight != nil {
				if insightErr := sink.BulkInsertTwitterInsights(ctx, []*chmodels.TwitterInsights{chInsight}); insightErr != nil {
					log.Warn().Err(insightErr).Str("error_message", insightErr.Error()).Str("function", "ProcessAccount").Str("stage", "insert_insights").Msg("Failed to insert insights (continuing)")
					logger.CaptureException(insightErr, map[string]string{"platform": "twitter", "component": "immediate-processor", "stage": "insert_insights", "twitter_id": wo.TwitterID}, nil)
				} else {
					log.Info().Str("record_id", insightsParsed.RecordID).Msg("Insights inserted")
				}
			}
		}
	}

	creditsUsed := rawTweetsFetched
	if userInfo != nil {
		creditsUsed++
	}
	if metaErr := mongoRepo.InsertTwitterJobMetadata(ctx, mongodb.TwitterJobMetadataPayload{
		PlatformID:  wo.TwitterID,
		WorkspaceID: wo.WorkspaceID,
		CreditsUsed: creditsUsed,
		ExecutedBy:  firstNonEmpty(wo.ExecutedBy, "internal"),
		AppID:       wo.AppID,
		AppName:     wo.AppName,
	}); metaErr != nil {
		log.Warn().Err(metaErr).Str("twitter_id", wo.TwitterID).Msg("Failed to insert twitter job metadata")
	} else {
		log.Info().
			Str("twitter_id", wo.TwitterID).
			Int("credits_used", creditsUsed).
			Msg("Inserted twitter job metadata")
	}

	// 7. Send notifications if account was fetched from MongoDB
	if account != nil {
		userID := account.GetUserIDHex()
		workspaceID := account.GetWorkspaceIDHex()
		if workspaceID == "" {
			workspaceID = wo.WorkspaceID
		}

		SendPusherNotification(pusherClient, account, workspaceID, originalState, log)

		if originalState == "Added" {
			SendEmailNotification(notifier, userID, workspaceID, wo.TwitterID, account.PlatformName, log)
		}

		log.Info().
			Str("twitter_id", wo.TwitterID).
			Str("workspace_id", workspaceID).
			Bool("email_sent", originalState == "Added").
			Msg("Notifications sent")
	}

	// MongoDB state update handled by deferred function

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

// SendPusherNotification sends a real-time notification via Pusher when Twitter analytics are completed.
func SendPusherNotification(pusherClient PusherClientInterface, account *mongomodels.SocialIntegration, workspaceID, originalState string, log *logger.Logger) {
	if pusherClient == nil {
		return
	}

	twitterID := account.PlatformIdentifier
	if twitterID == "" {
		twitterID = account.TwitterID
	}
	if twitterID == "" {
		twitterID = "unknown"
	}

	// Frontend Twitter analytics page subscribes to twitter-analytics-channel-{workspace_id}-{twitter_id}.
	channel := fmt.Sprintf("twitter-analytics-channel-%s-%s", workspaceID, twitterID)
	event := fmt.Sprintf("syncing-%s-%s", workspaceID, twitterID)

	data := map[string]interface{}{
		"state":                     "Processed",
		"account":                   twitterID,
		"name":                      account.PlatformName,
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

// SendEmailNotification sends an analytics completion notification email via the backend API.
func SendEmailNotification(notifier NotifierInterface, userID, workspaceID, accountID, accountName string, log *logger.Logger) {
	if notifier == nil {
		return
	}

	err := notifier.SendAnalyticsNotification(
		userID,
		workspaceID,
		"twitter",
		accountID,
		accountName,
		false,
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
