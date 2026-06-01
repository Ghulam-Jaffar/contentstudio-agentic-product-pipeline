package processor

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ImmediateWorkOrder mirrors TikTokAccountWorkOrder for immediate processing
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

// TikTokVideoData represents the structure of a single video from TikTok API
type TikTokVideoData struct {
	ID               string `json:"id"`
	CreateTime       int64  `json:"create_time"`
	CoverImageURL    string `json:"cover_image_url,omitempty"`
	ShareURL         string `json:"share_url,omitempty"`
	VideoDescription string `json:"video_description,omitempty"`
	Duration         int64  `json:"duration,omitempty"`
	Height           int64  `json:"height,omitempty"`
	Width            int64  `json:"width,omitempty"`
	Title            string `json:"title,omitempty"`
	EmbedHTML        string `json:"embed_html,omitempty"`
	EmbedLink        string `json:"embed_link,omitempty"`
	LikeCount        int64  `json:"like_count"`
	CommentCount     int64  `json:"comment_count"`
	ShareCount       int64  `json:"share_count"`
	ViewCount        int64  `json:"view_count"`
}

// TikTokUserData represents user info from TikTok API
type TikTokUserData struct {
	OpenID          string `json:"open_id"`
	UnionID         string `json:"union_id"`
	AvatarURL       string `json:"avatar_url"`
	AvatarURL100    string `json:"avatar_url_100"`
	AvatarLargeURL  string `json:"avatar_large_url"`
	DisplayName     string `json:"display_name"`
	BioDescription  string `json:"bio_description"`
	ProfileDeepLink string `json:"profile_deep_link"`
	IsVerified      bool   `json:"is_verified"`
	FollowerCount   int64  `json:"follower_count"`
	FollowingCount  int64  `json:"following_count"`
	LikesCount      int64  `json:"likes_count"`
	VideoCount      int64  `json:"video_count"`
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

// Processor handles TikTok immediate account processing.
type Processor struct {
	tkClient     *social.TikTokClient
	sink         *conversions.ClickHouseSink
	mongoRepo    mongodb.UnifiedSocialRepository
	notifier     *notification.Service
	pusherClient *notification.PusherClient
	log          *logger.Logger
	cfg          *config.Config
}

// New creates a new TikTok Processor with all dependencies.
// The TikTok API client is created internally using config credentials.
func New(
	mongoRepo mongodb.UnifiedSocialRepository,
	sink *conversions.ClickHouseSink,
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	log *logger.Logger,
	cfg *config.Config,
) *Processor {
	return &Processor{
		tkClient:     social.NewTikTokClient(cfg.TikTok.ClientKey, cfg.TikTok.ClientSecret),
		sink:         sink,
		mongoRepo:    mongoRepo,
		notifier:     notifier,
		pusherClient: pusherClient,
		log:          log,
		cfg:          cfg,
	}
}

// ProcessAccount implements the TikTokProcessor interface using the concrete dependencies.
func (p *Processor) ProcessAccount(ctx context.Context, wo ImmediateWorkOrder) error {
	return ProcessAccount(ctx, p.tkClient, p.sink, p.mongoRepo, p.notifier, p.pusherClient, wo, p.cfg.DecryptionKey, p.log)
}

// ProcessAccount processes a TikTok account and stores posts and insights in ClickHouse.
// It handles token decryption, user info fetching, video fetching, data parsing, and notifications.
func ProcessAccount(
	ctx context.Context,
	tkClient *social.TikTokClient,
	sink *conversions.ClickHouseSink,
	mongoRepo interface{}, // UnifiedSocialRepository
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	wo ImmediateWorkOrder,
	decryptionKey string,
	log *logger.Logger,
) (err error) {
	op := log.Operation("ProcessTikTokAccount").
		WithField("workspace_id", wo.WorkspaceID).
		WithField("tiktok_id", wo.TikTokID).
		WithField("sync_type", wo.SyncType).
		WithSentryTags(map[string]string{
			"workspace_id": wo.WorkspaceID,
			"tiktok_id":    wo.TikTokID,
			"sync_type":    wo.SyncType,
		})
	op.Start("processing tiktok work order")
	defer op.Complete(nil, "")

	var chPosts []*chmodels.TikTokPosts
	var chInsights *chmodels.TikTokInsights

	// Fetch account from MongoDB to get open_id and original state for notifications.
	var mongoOpenID string
	var accountForNotifications *mongomodels.SocialIntegration
	var originalState string
	var accountID primitive.ObjectID
	hasAccountID := false
	if wo.ID != "" {
		accountID, err = primitive.ObjectIDFromHex(wo.ID)
		if err != nil {
			log.Warn().Err(err).Str("account_id", wo.ID).Msg("Invalid account ID, continuing without MongoDB data")
		} else {
			hasAccountID = true
			// Type assert mongoRepo to UnifiedSocialRepository
			repo, ok := mongoRepo.(interface {
				FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
			})
			if ok {
				account, err := repo.FindByID(ctx, accountID)
				if err != nil {
					log.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to fetch account from MongoDB")
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
					accountForNotifications = account
					originalState = account.State
					// Extract open_id from ExtraData
					if openID, ok := account.ExtraData["open_id"]; ok {
						if openIDStr, ok := openID.(string); ok && openIDStr != "" {
							mongoOpenID = openIDStr
							log.Info().Str("open_id", mongoOpenID).Msg("Fetched open_id from MongoDB")
						}
					}
				}
			}
		}
	}
	defer func() {
		if !hasAccountID || err == nil {
			return
		}
		if repo, ok := mongoRepo.(interface {
			RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error
		}); ok {
			if recordErr := repo.RecordProcessingError(ctx, accountID, err.Error()); recordErr != nil {
				log.Warn().Err(recordErr).Str("account_id", wo.ID).Msg("Failed to record processing error")
			}
		}
	}()

	// Decrypt access token if needed
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

	// Fetch user info first to get profile data and aggregated stats
	log.Info().Msg("Fetching user profile information")
	userRaw, err := tkClient.FetchUserInfo(ctx, accessToken)
	if err != nil {
		if social.IsExpectedCompetitorErrorTikTok(err) {
			log.Warn().Err(err).Msg("Failed to fetch user info (expected token/permission error)")
		} else {
			log.Warn().
				Err(err).
				Str("error_message", err.Error()).
				Str("function", "ProcessAccount").
				Str("stage", "fetch_user_info").
				Msg("Failed to fetch user info (unexpected error, continuing)")
			logger.CaptureException(err, map[string]string{
				"platform":     "tiktok",
				"component":    "immediate-processor",
				"stage":        "fetch_user_info",
				"tiktok_id":    wo.TikTokID,
				"workspace_id": wo.WorkspaceID,
			}, nil)
		}
		// Continue anyway, we can still fetch videos
	} else {
		var userData TikTokUserData
		if err := json.Unmarshal(userRaw, &userData); err != nil {
			log.Warn().Err(err).Str("error_message", err.Error()).Str("function", "ProcessAccount").Str("stage", "unmarshal_user_data").Msg("Failed to unmarshal user data (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "tiktok", "component": "immediate-processor", "stage": "unmarshal_user_data", "tiktok_id": wo.TikTokID}, nil)
		} else {
			log.Info().
				Str("display_name", userData.DisplayName).
				Int64("follower_count", userData.FollowerCount).
				Msg("Fetched user profile")
			// Use open_id from MongoDB if available, otherwise use from API response
			if mongoOpenID == "" && userData.OpenID != "" {
				mongoOpenID = userData.OpenID
				log.Debug().Str("open_id", mongoOpenID).Msg("Using open_id from TikTok API response")
			}
		}
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

	// Fetch videos from TikTok API with constraints:
	// For requested date ranges, only store videos within start/end.
	// When no date range is provided, preserve the legacy 90-day cutoff.
	const maxVideosImmediate = 999
	cutoffTime := time.Now().AddDate(0, 0, -90) // 90 days ago

	cursor := 0
	videosCount := 0
	totalLikes := int64(0)
	totalComments := int64(0)
	totalShares := int64(0)
	totalViews := int64(0)

	for videosCount < maxVideosImmediate {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		log.Debug().Int("cursor", cursor).Int("videos_collected", videosCount).Msg("Fetching videos batch")

		// Fetch a page of videos (TikTok API max is 20)
		rawVideoData, nextCursor, err := tkClient.FetchUserVideos(ctx, wo.TikTokID, accessToken, cursor, 20)
		if err != nil {
			if social.IsExpectedCompetitorErrorTikTok(err) {
				log.Warn().Err(err).Int("cursor", cursor).Msg("Failed to fetch videos (expected token/permission error)")
			} else {
				log.Warn().
					Err(err).
					Str("error_message", err.Error()).
					Str("function", "ProcessAccount").
					Str("stage", "fetch_videos").
					Msg("Failed to fetch videos (unexpected error)")
				logger.CaptureException(err, map[string]string{
					"platform":     "tiktok",
					"component":    "immediate-processor",
					"stage":        "fetch_videos",
					"tiktok_id":    wo.TikTokID,
					"workspace_id": wo.WorkspaceID,
				}, map[string]interface{}{
					"cursor":           cursor,
					"videos_collected": videosCount,
				})
			}
			break
		}

		// Parse video items from raw data
		var videos []TikTokVideoData
		if err := json.Unmarshal(rawVideoData, &videos); err != nil {
			log.Warn().
				Err(err).
				Str("error_message", err.Error()).
				Str("function", "ProcessAccount").
				Str("stage", "unmarshal_videos").
				Msg("Failed to unmarshal videos array")
			logger.CaptureException(err, map[string]string{
				"platform":     "tiktok",
				"component":    "immediate-processor",
				"stage":        "unmarshal_videos",
				"tiktok_id":    wo.TikTokID,
				"workspace_id": wo.WorkspaceID,
			}, nil)
			break
		}

		// Convert parsed videos to ClickHouse model and respect cutoff time/date range.
		reachedCutoff := false
		for _, video := range videos {
			videoTime := time.Unix(video.CreateTime, 0).UTC()
			if hasRequestedRange {
				if videoTime.Before(startTime) {
					reachedCutoff = true
					break
				}
				if !videoTime.Before(endTime) {
					continue
				}
			} else if videoTime.Before(cutoffTime) {
				reachedCutoff = true
				break
			}

			// Extract hashtags from description or title
			text := video.VideoDescription
			if text == "" {
				text = video.Title
			}
			hashtags := extractHashtags(text)

			// Build parsed post - ensure all fields match Python implementation
			engagement := video.LikeCount + video.CommentCount + video.ShareCount
			engagementRate := 0.0
			if video.ViewCount > 0 {
				engagementRate = float64(engagement) / float64(video.ViewCount)
			}

			// Get user info if available from userRaw
			displayName := ""
			profileLink := ""
			if userRaw != nil {
				var userData TikTokUserData
				if err := json.Unmarshal(userRaw, &userData); err == nil {
					displayName = userData.DisplayName
					profileLink = userData.ProfileDeepLink
				}
			}

			// Use open_id from MongoDB if available, otherwise fall back to account ID
			tiktokID := mongoOpenID
			if tiktokID == "" {
				tiktokID = wo.TikTokID
			}

			parsed := &kafkamodels.ParsedTikTokPost{
				ID:              video.ID,
				TikTokID:        tiktokID,
				DisplayName:     displayName,
				ProfileLink:     profileLink,
				PostDescription: video.VideoDescription,
				CoverImageURL:   video.CoverImageURL,
				ShareURL:        video.ShareURL,
				Duration:        video.Duration,
				Height:          video.Height,
				Width:           video.Width,
				Title:           video.Title,
				EmbedHTML:       video.EmbedHTML,
				EmbedLink:       video.EmbedLink,
				LikeCount:       video.LikeCount,
				CommentCount:    video.CommentCount,
				ShareCount:      video.ShareCount,
				ViewCount:       video.ViewCount,
				EngagementCount: engagement,
				EngagementRate:  engagementRate,
				CreateTime:      video.CreateTime,
				CreatedAt:       time.Unix(video.CreateTime, 0),
				Hashtags:        hashtags,
			}

			// Aggregate stats for insights
			totalLikes += video.LikeCount
			totalComments += video.CommentCount
			totalShares += video.ShareCount
			totalViews += video.ViewCount

			// Convert to ClickHouse model
			cp := conversions.ConvertTikTokPost(parsed)
			if cp != nil {
				chPosts = append(chPosts, cp)
				videosCount++
			}

			// Stop if we've reached the max
			if videosCount >= maxVideosImmediate {
				break
			}
		}

		// Check if there are more results or if we've hit the 90-day cutoff
		if reachedCutoff || nextCursor == 0 || videosCount >= maxVideosImmediate {
			break
		}
		cursor = int(nextCursor)

		// Avoid tight loop
		time.Sleep(300 * time.Millisecond)
	}

	log.Info().
		Int("videos_fetched", videosCount).
		Int64("total_likes", totalLikes).
		Int64("total_comments", totalComments).
		Int64("total_shares", totalShares).
		Int64("total_views", totalViews).
		Msg("Completed video fetch with constraints")

	// Skip if no posts were parsed
	if len(chPosts) == 0 {
		log.Info().Msg("No posts parsed")
	}

	// Insert posts into ClickHouse when we have any parsed posts.
	if len(chPosts) > 0 {
		if err := sink.BulkInsertTikTokPosts(ctx, chPosts); err != nil {
			log.Warn().Err(err).Str("error_message", err.Error()).Int("posts_count", len(chPosts)).Str("function", "ProcessAccount").Str("stage", "insert_posts").Msg("Failed to insert posts")
			return err
		}

		log.Info().Int("inserted_posts", len(chPosts)).Msg("Successfully inserted posts")
	}

	// Query database to get total_video_views from all posts for this tiktok_id
	// Use open_id from MongoDB if available, otherwise fall back to account ID
	tiktokIDForQuery := mongoOpenID
	if tiktokIDForQuery == "" {
		tiktokIDForQuery = wo.TikTokID
	}

	databaseTotalViews, err := sink.RawClient.GetTikTokPostsViewSum(ctx, tiktokIDForQuery)
	if err != nil {
		log.Warn().Err(err).Str("error_message", err.Error()).Str("tiktok_id", tiktokIDForQuery).Str("function", "ProcessAccount").Str("stage", "query_total_views").Msg("Failed to query total views from database, using current batch sum")
		logger.CaptureException(err, map[string]string{"platform": "tiktok", "component": "immediate-processor", "stage": "query_total_views", "tiktok_id": tiktokIDForQuery}, nil)
		// Fall back to the sum from current batch if query fails
		databaseTotalViews = totalViews
	} else {
		log.Info().Int64("database_total_views", databaseTotalViews).Msg("Queried total views from database")
	}

	// Generate and insert insights if we have user data
	if userRaw != nil {
		var userData TikTokUserData
		if err := json.Unmarshal(userRaw, &userData); err == nil {
			// Generate insights using the parser to match Python implementation
			parser := parsing.NewTikTokParser()
			userInfo := &parsing.TikTokUserInfo{
				OpenID:          userData.OpenID,
				UnionID:         userData.UnionID,
				AvatarURL:       userData.AvatarURL,
				AvatarURL100:    userData.AvatarURL100,
				AvatarLargeURL:  userData.AvatarLargeURL,
				DisplayName:     userData.DisplayName,
				BioDescription:  userData.BioDescription,
				ProfileDeepLink: userData.ProfileDeepLink,
				IsVerified:      userData.IsVerified,
				FollowerCount:   userData.FollowerCount,
				FollowingCount:  userData.FollowingCount,
				LikesCount:      userData.LikesCount,
				VideoCount:      userData.VideoCount,
			}
			// Use open_id from MongoDB if available, otherwise fall back to account ID
			tiktokID := mongoOpenID
			if tiktokID == "" {
				tiktokID = wo.TikTokID
			}
			// Use database total views instead of the batch total
			parsedInsights := parser.GenerateInsights(userInfo, tiktokID, mongoOpenID, databaseTotalViews, totalLikes, totalComments, totalShares)
			chInsights = conversions.ConvertTikTokInsights(parsedInsights)

			if err := sink.BulkInsertTikTokInsights(ctx, []*chmodels.TikTokInsights{chInsights}); err != nil {
				log.Warn().Err(err).Str("error_message", err.Error()).Str("function", "ProcessAccount").Str("stage", "insert_insights").Msg("Failed to insert insights (continuing)")
				logger.CaptureException(err, map[string]string{"platform": "tiktok", "component": "immediate-processor", "stage": "insert_insights", "tiktok_id": wo.TikTokID}, nil)
			} else {
				log.Info().Msg("Successfully inserted insights")
			}

		}
	}

	// Update MongoDB state to "Processed" after successful completion
	if wo.ID != "" {
		accountID, err := primitive.ObjectIDFromHex(wo.ID)
		if err != nil {
			log.Warn().Err(err).Str("account_id", wo.ID).Msg("Invalid account ID, skipping state update")
		} else {
			// Type assert mongoRepo to UnifiedSocialRepository
			repo, ok := mongoRepo.(interface {
				Update(ctx context.Context, id primitive.ObjectID, updates bson.M) error
				FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
				ClearProcessingError(ctx context.Context, id primitive.ObjectID) error
			})
			if ok {
				updates := bson.M{
					"state":                     mongomodels.StateProcessed,
					"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02 15:04:05"),
				}
				if err := repo.Update(ctx, accountID, updates); err != nil {
					log.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", wo.ID).Str("function", "ProcessAccount").Msg("Failed to update account state to Processed")
				} else {
					log.Info().Str("account_id", wo.ID).Str("state", mongomodels.StateProcessed).Msg("Updated account state to Processed")
				}
				if err := repo.ClearProcessingError(ctx, accountID); err != nil {
					log.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to clear processing error")
				}
			}
		}
	}

	// Send notifications
	if pusherClient != nil {
		SendNotifications(ctx, mongoRepo, notifier, pusherClient, wo, accountForNotifications, originalState, log)
	}
	return nil
}

// extractHashtags extracts hashtags from text
func extractHashtags(text string) []string {
	hashtags := []string{}
	re := regexp.MustCompile(`#\w+`)
	matches := re.FindAllString(text, -1)
	for _, match := range matches {
		tag := strings.TrimPrefix(match, "#")
		if tag != "" {
			hashtags = append(hashtags, tag)
		}
	}
	return hashtags
}

// SendNotifications sends Pusher and email notifications for TikTok analytics completion
func SendNotifications(
	ctx context.Context,
	mongoRepo interface{}, // UnifiedSocialRepository
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	wo ImmediateWorkOrder,
	account *mongomodels.SocialIntegration,
	originalState string,
	log *logger.Logger,
) {
	accountID := wo.TikTokID
	if account != nil && account.PlatformIdentifier != "" {
		accountID = account.PlatformIdentifier
	}
	channel := fmt.Sprintf("tt-analytics-channel-%s-%s", wo.WorkspaceID, accountID)
	event := fmt.Sprintf("syncing-%s-%s", wo.WorkspaceID, accountID)

	data := map[string]interface{}{
		"state":                     "Processed",
		"account":                   accountID,
		"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02"),
	}

	if err := pusherClient.Trigger(channel, event, data); err != nil {
		log.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("channel", channel).
			Str("event", event).
			Str("function", "SendNotifications").
			Msg("Failed to send Pusher notification")
	} else {
		log.Debug().
			Str("channel", channel).
			Str("event", event).
			Msg("Sent Pusher notification")
	}

	// Send email only for newly added accounts (python behavior).
	if notifier == nil || originalState != mongomodels.StateAdded {
		return
	}

	// Fetch account from MongoDB if not already fetched
	if account == nil {
		if wo.ID == "" {
			log.Warn().Msg("No account ID available for email notification")
			return
		}

		accountID, err := primitive.ObjectIDFromHex(wo.ID)
		if err != nil {
			log.Warn().Err(err).Str("account_id", wo.ID).Msg("Invalid account ID for email notification")
			return
		}

		// Type assert mongoRepo to UnifiedSocialRepository
		repo, ok := mongoRepo.(interface {
			FindByID(ctx context.Context, id primitive.ObjectID) (*mongomodels.SocialIntegration, error)
		})
		if !ok {
			log.Warn().Msg("MongoDB repository does not support FindByID")
			return
		}

		var fetchErr error
		account, fetchErr = repo.FindByID(ctx, accountID)
		if fetchErr != nil {
			log.Warn().Err(fetchErr).Str("account_id", wo.ID).Msg("Failed to fetch account for email notification")
			return
		}
	}

	if account == nil {
		log.Warn().Str("account_id", wo.ID).Msg("Account not found for email notification")
		return
	}

	userID := account.GetUserIDHex()
	if userID == "" || userID == "000000000000000000000000" {
		log.Warn().Str("account_id", wo.ID).Msg("Account has no user_id set, skipping email notification")
		return
	}

	// Get account name
	accountName := account.PlatformName
	if accountName == "" {
		accountName = account.PlatformIdentifier
	}

	workspaceID := account.GetWorkspaceIDHex()
	if workspaceID == "" {
		workspaceID = wo.WorkspaceID
	}

	// Send email through notification service
	if err := notifier.SendAnalyticsNotification(userID, workspaceID, "tiktok", account.ID.Hex(), accountName, false); err != nil {
		log.Warn().
			Err(err).
			Str("account_id", wo.ID).
			Str("user_id", userID).
			Msg("Failed to send email notification")
	} else {
		log.Debug().
			Str("account_id", wo.ID).
			Str("user_id", userID).
			Msg("Sent email notification")
	}
}
