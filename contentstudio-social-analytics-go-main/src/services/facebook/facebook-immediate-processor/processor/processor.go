package processor

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

type WorkOrder struct {
	ID              string `json:"id"`
	AccountID       string `json:"account_id"`
	Type            string `json:"type"`
	AccessToken     string `json:"access_token"`
	WorkspaceID     string `json:"workspace_id"`
	LongAccessToken string `json:"long_access_token"`
	SyncType        string `json:"sync_type"`
	StartDate       string `json:"start_date,omitempty"`
	EndDate         string `json:"end_date,omitempty"`
}

type ParsedData struct {
	Posts         []kafkamodels.ParsedFacebookPost
	MediaAssets   []kafkamodels.ParsedFacebookMediaAsset
	VideoInsights []kafkamodels.ParsedFacebookVideoInsights
	ReelsInsights []kafkamodels.ParsedFacebookReelsInsights
	Insights      []*kafkamodels.ParsedFacebookInsights
}

// isExpectedFacebookError returns true for expected auth/permission errors that should not go to Sentry.
func isExpectedFacebookError(err error) bool {
	if err == nil {
		return false
	}
	return social.IsExpectedCompetitorErrorFB(err)
}

type Processor struct {
	MongoRepo      mongodb.UnifiedSocialRepository
	FacebookClient FacebookClientInterface
	Parser         *parsing.FacebookParser
	Sink           ClickHouseSinkInterface
	Notifier       NotifierInterface
	PusherClient   PusherClientInterface
	Producer       kafka2.Producer
	Logger         *logger.Logger
	Config         *config.Config
}

func New(
	mongoRepo mongodb.UnifiedSocialRepository,
	sink *conversions.ClickHouseSink,
	producer kafka2.Producer,
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	log *logger.Logger,
	cfg *config.Config,
) *Processor {
	return &Processor{
		MongoRepo:      mongoRepo,
		FacebookClient: social.NewFacebookClient(cfg.Facebook.AppSecret),
		Parser:         parsing.NewFacebookParser(),
		Sink:           sink,
		Notifier:       notifier,
		PusherClient:   pusherClient,
		Producer:       producer,
		Logger:         log,
		Config:         cfg,
	}
}

func (p *Processor) ProcessAccount(ctx context.Context, workOrder WorkOrder) (err error) {
	baseTags := map[string]string{
		"platform":     "facebook",
		"component":    "immediate-processor",
		"account_id":   workOrder.ID,
		"facebook_id":  workOrder.AccountID,
		"workspace_id": workOrder.WorkspaceID,
		"sync_type":    workOrder.SyncType,
	}
	baseExtras := map[string]interface{}{
		"account_id":   workOrder.ID,
		"facebook_id":  workOrder.AccountID,
		"workspace_id": workOrder.WorkspaceID,
		"sync_type":    workOrder.SyncType,
	}

	capture := func(stage string, e error, extra map[string]interface{}) {
		if e == nil {
			return
		}
		tags := make(map[string]string, len(baseTags)+1)
		for k, v := range baseTags {
			tags[k] = v
		}
		tags["stage"] = stage

		extras := make(map[string]interface{}, len(baseExtras)+len(extra))
		for k, v := range baseExtras {
			extras[k] = v
		}
		for k, v := range extra {
			extras[k] = v
		}
		logger.CaptureException(e, tags, extras)
	}

	accountID, err := primitive.ObjectIDFromHex(workOrder.ID)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: invalid account ID: %w", err)
	}
	defer func() {
		if err == nil {
			return
		}
		if recordErr := p.MongoRepo.RecordProcessingError(ctx, accountID, err.Error()); recordErr != nil {
			p.Logger.Warn().Err(recordErr).Str("account_id", workOrder.ID).Msg("Failed to record processing error")
		}
	}()

	account, err := p.MongoRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to fetch account from MongoDB: %w", err)
	}
	if account == nil {
		return fmt.Errorf("Processor.ProcessAccount: account not found: %s", workOrder.ID)
	}
	if mongodb.HasProcessingErrorMeta(account.MetaData) {
		if err := p.MongoRepo.ClearProcessingError(ctx, accountID); err != nil {
			p.Logger.Warn().Err(err).Str("account_id", workOrder.ID).Msg("Failed to clear stale processing error before retry")
		}
	}

	accessToken, err := crypto.DecryptToken(account.GetAccessToken(), p.Config.DecryptionKey)
	if err != nil {
		accessToken = getStringFromExtraData(account.ExtraData, "access_token")
		if accessToken == "" {
			return fmt.Errorf("Processor.ProcessAccount: no valid access token available")
		}
	}

	since, until, err := resolveFacebookDateRange(workOrder.StartDate, workOrder.EndDate)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: invalid date range: %w", err)
	}

	originalState := account.State
	if err := p.MongoRepo.UpdateState(ctx, accountID, "Processing"); err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", workOrder.ID).Str("function", "ProcessAccount").Str("stage", "set_state_processing").Msg("Failed to set account state to Processing")
	}

	posts, videos, insights, err := p.fetchAllData(ctx, workOrder, account.PlatformIdentifier, accessToken, since, until, capture)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to fetch data: %w", err)
	}

	p.Logger.Info().
		Str("account_id", workOrder.ID).
		Int("posts_fetched", len(posts)).
		Int("videos_fetched", len(videos)).
		Bool("has_insights", insights != nil).
		Msg("Completed fetching data from Facebook")

	parsedData, err := p.parseAllData(workOrder, account, posts, videos, insights, capture)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to parse data: %w", err)
	}

	if err := p.storeInClickHouse(ctx, workOrder, parsedData); err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to store data: %w", err)
	}

	hasFetchedData := len(posts) > 0 || len(videos) > 0 || insights != nil
	now := time.Now()
	if err := p.MongoRepo.UpdateState(ctx, accountID, "Processed"); err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", workOrder.ID).Str("function", "ProcessAccount").Str("stage", "set_state_processed").Msg("Failed to set account state to Processed")
	}
	if err := p.MongoRepo.ClearProcessingError(ctx, accountID); err != nil {
		p.Logger.Warn().Err(err).Str("account_id", workOrder.ID).Msg("Failed to clear processing error")
	}
	if err := p.MongoRepo.UpdateAnalyticsTimestamp(ctx, accountID, "analytics", now); err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", workOrder.ID).Str("function", "ProcessAccount").Str("stage", "update_analytics_timestamp").Msg("Failed to update analytics timestamp")
	}
	if err := p.MongoRepo.UpdateAnalyticsTimestamp(ctx, accountID, "insights", now); err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", workOrder.ID).Str("function", "ProcessAccount").Str("stage", "update_insights_timestamp").Msg("Failed to update insights timestamp")
	}
	if err := p.MongoRepo.UpdateAnalyticsTimestamp(ctx, accountID, "video", now); err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", workOrder.ID).Str("function", "ProcessAccount").Str("stage", "update_video_timestamp").Msg("Failed to update video timestamp")
	}

	userID := account.GetUserIDHex()
	workspaceID := account.GetWorkspaceIDHex()
	if workspaceID == "" {
		workspaceID = workOrder.WorkspaceID
	}

	if !hasFetchedData {
		p.Logger.Warn().
			Str("account_id", workOrder.ID).
			Str("facebook_id", account.FacebookID).
			Msg("Skipping notifications because no Facebook data was fetched")
		return nil
	}

	p.sendPusherNotification(account, workspaceID, originalState)

	if originalState == "Added" {
		p.sendEmailNotification(userID, workspaceID, account.PlatformIdentifier, account.PlatformName)
	}

	return nil
}

func (p *Processor) fetchAllData(
	ctx context.Context,
	workOrder WorkOrder,
	pageID, accessToken string,
	since, until time.Time,
	capture func(string, error, map[string]interface{}),
) (
	posts []kafkamodels.RawFacebookPost,
	videos []kafkamodels.RawFacebookVideo,
	insights *kafkamodels.RawFacebookInsights,
	err error,
) {
	var (
		postsMu    sync.Mutex
		videosMu   sync.Mutex
		insightsMu sync.Mutex
	)

	eg, egctx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		fetchedPosts, fetchErr := p.FacebookClient.FetchPostsSince(egctx, pageID, accessToken, since, until)
		if fetchErr != nil {
			if isExpectedFacebookError(fetchErr) {
				p.Logger.Warn().
					Err(fetchErr).
					Str("facebook_id", pageID).
					Msg("Failed to fetch Facebook posts (expected API error)")
			} else {
				p.Logger.Warn().
					Err(fetchErr).
					Str("error_message", fetchErr.Error()).
					Str("facebook_id", pageID).
					Str("function", "fetchAllData").
					Str("stage", "fetch_posts").
					Msg("Failed to fetch Facebook posts (unexpected error, continuing)")
				capture("fetch_posts", fetchErr, nil)
			}
			return nil
		}
		postsMu.Lock()
		posts = fetchedPosts
		postsMu.Unlock()
		return nil
	})

	eg.Go(func() error {
		fetchedVideos, fetchErr := p.FacebookClient.FetchVideosSince(egctx, pageID, accessToken, since, until)
		if fetchErr != nil {
			if isExpectedFacebookError(fetchErr) {
				p.Logger.Warn().
					Err(fetchErr).
					Str("facebook_id", pageID).
					Msg("Failed to fetch Facebook videos (expected API error)")
			} else {
				p.Logger.Warn().
					Err(fetchErr).
					Str("error_message", fetchErr.Error()).
					Str("facebook_id", pageID).
					Str("function", "fetchAllData").
					Str("stage", "fetch_videos").
					Msg("Failed to fetch Facebook videos (unexpected error, continuing)")
				capture("fetch_videos", fetchErr, nil)
			}
			return nil
		}
		videosMu.Lock()
		videos = fetchedVideos
		videosMu.Unlock()
		return nil
	})

	eg.Go(func() error {
		fetchedInsights, fetchErr := p.FacebookClient.FetchInsights(egctx, pageID, accessToken, since, until)
		if fetchErr != nil {
			if isExpectedFacebookError(fetchErr) {
				p.Logger.Warn().
					Err(fetchErr).
					Str("facebook_id", pageID).
					Msg("Failed to fetch Facebook insights (expected API error)")
			} else {
				p.Logger.Warn().
					Err(fetchErr).
					Str("error_message", fetchErr.Error()).
					Str("facebook_id", pageID).
					Str("function", "fetchAllData").
					Str("stage", "fetch_insights").
					Msg("Failed to fetch Facebook insights (unexpected error, continuing)")
				capture("fetch_insights", fetchErr, nil)
			}
			return nil
		}
		insightsMu.Lock()
		insights = fetchedInsights
		insightsMu.Unlock()
		return nil
	})

	if err = eg.Wait(); err != nil {
		return
	}

	return
}

func resolveFacebookDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	now := time.Now().UTC()
	defaultEnd := time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 0, time.UTC)
	defaultStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -90)

	startDate = strings.TrimSpace(startDate)
	endDate = strings.TrimSpace(endDate)
	if startDate == "" && endDate == "" {
		return defaultStart, defaultEnd, nil
	}
	if startDate == "" || endDate == "" {
		return time.Time{}, time.Time{}, fmt.Errorf("start_date and end_date are both required")
	}

	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date %q: %w", startDate, err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date %q: %w", endDate, err)
	}
	start = time.Date(start.Year(), start.Month(), start.Day(), 0, 0, 0, 0, time.UTC)
	end = time.Date(end.Year(), end.Month(), end.Day(), 23, 59, 59, 0, time.UTC)
	if end.Before(start) {
		return time.Time{}, time.Time{}, fmt.Errorf("end_date must not be before start_date")
	}
	return start, end, nil
}

func (p *Processor) parseAllData(
	workOrder WorkOrder,
	account *mongomodels.SocialIntegration,
	posts []kafkamodels.RawFacebookPost,
	videos []kafkamodels.RawFacebookVideo,
	insights *kafkamodels.RawFacebookInsights,
	capture func(string, error, map[string]interface{}),
) (*ParsedData, error) {
	result := &ParsedData{}
	workspaceID := getStringFromExtraData(account.ExtraData, "workspace_id")
	pageName := getStringFromExtraData(account.ExtraData, "name")

	for _, post := range posts {
		parsed, assets, parseErr := p.Parser.ParsePost(post, account.FacebookID, pageName, workspaceID)
		if parseErr != nil {
			capture("parse_post", parseErr, map[string]interface{}{"post_id": post.ID})
			continue
		}
		result.Posts = append(result.Posts, *parsed)
		result.MediaAssets = append(result.MediaAssets, assets...)
	}

	filteredVideos, skippedVideos := parsing.FilterFacebookVideos(account.FacebookID, posts, videos)
	if skippedVideos > 0 {
		p.Logger.Warn().
			Str("account_id", workOrder.ID).
			Str("facebook_id", account.FacebookID).
			Int("videos_fetched", len(videos)).
			Int("videos_skipped", skippedVideos).
			Int("videos_allowed", len(filteredVideos)).
			Msg("Skipped raw Facebook videos that did not match a video or reel post")
	}

	for _, video := range filteredVideos {
		parsed, parseErr := p.Parser.ParseVideo(video, account.FacebookID, pageName)
		if parseErr != nil {
			capture("parse_video", parseErr, map[string]interface{}{"video_id": video.ID})
			continue
		}

		if parsed.BlueReelsPlayCount > 0 {
			reels := kafkamodels.ParsedFacebookReelsInsights{
				PageID:               parsed.PageID,
				PostID:               parsed.PostID,
				AverageTimeWatched:   int64(parsed.PostVideoAvgTimeWatched),
				TotalTimeWatchedInMs: parsed.PostVideoViewTime,
				PlayCount:            parsed.BlueReelsPlayCount,
				ImpressionsUnique:    parsed.PostImpressionsUnique,
				ReelFollowers:        0,
				CreatedAt:            parsed.CreatedTime,
				SavingTime:           parsed.SavingTime,
			}
			result.ReelsInsights = append(result.ReelsInsights, reels)
		} else {
			result.VideoInsights = append(result.VideoInsights, parsed)
		}
	}

	if insights != nil {
		parsedList, parseErr := p.Parser.ParseInsightsDaily(*insights, account.FacebookID, workspaceID)
		if parseErr != nil {
			capture("parse_insights", parseErr, nil)
		} else {
			result.Insights = parsedList
		}
	}

	return result, nil
}

func (p *Processor) storeInClickHouse(
	ctx context.Context,
	workOrder WorkOrder,
	data *ParsedData,
) error {
	if data == nil {
		return errors.New("parsed data is nil")
	}

	if len(data.Posts) > 0 {
		clickhousePosts := make([]*clickhousemodels.FacebookPosts, 0, len(data.Posts))
		for _, post := range data.Posts {
			if converted := p.Sink.ConvertFacebookPost(&post); converted != nil {
				clickhousePosts = append(clickhousePosts, converted)
			}
		}
		if len(clickhousePosts) > 0 {
			if err := p.Sink.BulkInsertPosts(ctx, clickhousePosts); err != nil {
				return fmt.Errorf("Processor.storeInClickHouse: failed to insert posts: %w", err)
			}
		}
	}

	if len(data.MediaAssets) > 0 {
		clickhouseAssets := make([]*clickhousemodels.FacebookMediaAssets, 0, len(data.MediaAssets))
		for _, asset := range data.MediaAssets {
			if converted := p.Sink.ConvertFacebookMediaAssets(&asset); converted != nil {
				clickhouseAssets = append(clickhouseAssets, converted)
			}
		}
		if len(clickhouseAssets) > 0 {
			if err := p.Sink.BulkInsertMediaAssets(ctx, clickhouseAssets); err != nil {
				return fmt.Errorf("Processor.storeInClickHouse: failed to insert media assets: %w", err)
			}
		}
	}

	if len(data.VideoInsights) > 0 {
		clickhouseVideos := make([]*clickhousemodels.FacebookVideoInsights, 0, len(data.VideoInsights))
		for _, video := range data.VideoInsights {
			if converted := p.Sink.ConvertFacebookVideoInsights(&video); converted != nil {
				clickhouseVideos = append(clickhouseVideos, converted)
			}
		}
		if len(clickhouseVideos) > 0 {
			if err := p.Sink.BulkInsertVideoInsights(ctx, clickhouseVideos); err != nil {
				return fmt.Errorf("Processor.storeInClickHouse: failed to insert video insights: %w", err)
			}
		}
	}

	if len(data.ReelsInsights) > 0 {
		clickhouseReels := make([]*clickhousemodels.FacebookReelsInsights, 0, len(data.ReelsInsights))
		for _, reel := range data.ReelsInsights {
			if converted := p.Sink.ConvertFacebookReelsInsights(&reel); converted != nil {
				clickhouseReels = append(clickhouseReels, converted)
			}
		}
		if len(clickhouseReels) > 0 {
			if err := p.Sink.BulkInsertReelsInsights(ctx, clickhouseReels); err != nil {
				return fmt.Errorf("Processor.storeInClickHouse: failed to insert reels insights: %w", err)
			}
		}
	}

	if len(data.Insights) > 0 {
		clickhouseInsights := make([]*clickhousemodels.FacebookInsights, 0, len(data.Insights))
		for _, insight := range data.Insights {
			if converted := p.Sink.ConvertFacebookInsights(insight); converted != nil {
				clickhouseInsights = append(clickhouseInsights, converted)
			}
		}
		if len(clickhouseInsights) > 0 {
			if err := p.Sink.BulkInsertInsights(ctx, clickhouseInsights); err != nil {
				return fmt.Errorf("Processor.storeInClickHouse: failed to insert page insights: %w", err)
			}
		}
	}

	return nil
}

func (p *Processor) sendPusherNotification(account *mongomodels.SocialIntegration, workspaceID, originalState string) {
	if p.PusherClient == nil {
		return
	}

	accountID := account.FacebookID
	if accountID == "" {
		accountID = account.PlatformIdentifier
	}
	if accountID == "" {
		return
	}

	// Frontend subscribes to fb-analytics-channel-{workspace_id}-{facebook_id}.
	channel := fmt.Sprintf("fb-analytics-channel-%s-%s", workspaceID, accountID)
	event := fmt.Sprintf("syncing-%s-%s", workspaceID, accountID)

	data := map[string]interface{}{
		"state":                     "Processed",
		"account":                   accountID,
		"last_analytics_updated_at": time.Now().UTC().Format("2006-01-02"),
	}

	if err := p.PusherClient.Trigger(channel, event, data); err != nil {
		p.Logger.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("channel", channel).
			Str("event", event).
			Str("function", "sendPusherNotification").
			Msg("Failed to send Pusher notification")
	}
}

func (p *Processor) sendEmailNotification(userID, workspaceID, accountID, accountName string) {
	if p.Notifier == nil {
		return
	}
	if err := p.Notifier.SendAnalyticsNotification(userID, workspaceID, "facebook", accountID, accountName, false); err != nil {
		p.Logger.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("user_id", userID).
			Str("workspace_id", workspaceID).
			Str("account_id", accountID).
			Str("function", "sendEmailNotification").
			Msg("Failed to send analytics notification to backend")
	}
}

func (p *Processor) SendToDeadLetterQueue(workOrder WorkOrder, processingError error) {
	// Implementation for DLQ
}

func getStringFromExtraData(extraData map[string]interface{}, key string) string {
	if val, ok := extraData[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return ""
}
