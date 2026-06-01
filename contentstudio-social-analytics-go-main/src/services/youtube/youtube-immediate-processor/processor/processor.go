package processor

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
)

// ErrUnauthorized is returned when the API returns a 401 status
var ErrUnauthorized = errors.New("unauthorized: invalid or expired token")

const (
	DefaultInsightsDays     = 90
	FullSyncInsightsDays    = 365
	DefaultVideosDays       = 90
	FullSyncVideosStartYear = 2020
)

type WorkOrder struct {
	ID           string `json:"id"`
	ChannelID    string `json:"channel_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	WorkspaceID  string `json:"workspace_id"`
	SyncType     string `json:"sync_type"`
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
}

type FetchedData struct {
	Channel          *social.YouTubeChannelItem
	Videos           []social.YouTubeActivityItem
	VideoDetails     map[string]*social.YouTubeVideoItem // Video ID -> Video details with lifetime stats
	ActivityInsights *social.YouTubeAnalyticsResponse
	TrafficInsights  *social.YouTubeAnalyticsResponse
	SharedInsights   *social.YouTubeAnalyticsResponse
}

type ParsedData struct {
	Channel          *kafkamodels.ParsedYouTubeChannel
	Videos           []kafkamodels.ParsedYouTubeVideo
	ActivityInsights []kafkamodels.ParsedYouTubeActivityInsights
	TrafficInsights  []kafkamodels.ParsedYouTubeTrafficInsights
	SharedInsights   *kafkamodels.ParsedYouTubeSharedInsights
}

type Processor struct {
	MongoRepo    mongodb.UnifiedSocialRepository
	YTClient     social.YouTubeAPI
	Sink         *conversions.ClickHouseSink
	Notifier     *notification.Service
	PusherClient *notification.PusherClient
	Logger       *logger.Logger
	Cfg          *config.Config
}

func New(
	mongoRepo mongodb.UnifiedSocialRepository,
	sink *conversions.ClickHouseSink,
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	log *logger.Logger,
	cfg *config.Config,
) *Processor {
	return &Processor{
		MongoRepo:    mongoRepo,
		YTClient:     social.NewYouTubeClient(cfg.YouTube.ClientID, cfg.YouTube.ClientSecret),
		Sink:         sink,
		Notifier:     notifier,
		PusherClient: pusherClient,
		Logger:       log,
		Cfg:          cfg,
	}
}

func NewWithClient(
	mongoRepo mongodb.UnifiedSocialRepository,
	ytClient social.YouTubeAPI,
	sink *conversions.ClickHouseSink,
	notifier *notification.Service,
	pusherClient *notification.PusherClient,
	log *logger.Logger,
	cfg *config.Config,
) *Processor {
	return &Processor{
		MongoRepo:    mongoRepo,
		YTClient:     ytClient,
		Sink:         sink,
		Notifier:     notifier,
		PusherClient: pusherClient,
		Logger:       log,
		Cfg:          cfg,
	}
}

func (p *Processor) ProcessAccount(ctx context.Context, wo WorkOrder) (err error) {
	if wo.ChannelID == "" || wo.AccessToken == "" {
		p.Logger.Warn().
			Str("channel_id", wo.ChannelID).
			Bool("has_token", wo.AccessToken != "").
			Msg("Skipping work order with missing channel_id or access_token")
		return nil
	}

	p.Logger.Info().
		Str("channel_id", wo.ChannelID).
		Bool("has_access_token", wo.AccessToken != "").
		Bool("has_refresh_token", wo.RefreshToken != "").
		Int("access_token_len", len(wo.AccessToken)).
		Int("refresh_token_len", len(wo.RefreshToken)).
		Msg("YouTube work order received")

	// Decrypt token if encrypted
	if decrypted, err := crypto.DecryptToken(wo.AccessToken, p.Cfg.DecryptionKey); err == nil {
		wo.AccessToken = decrypted
		p.Logger.Debug().Str("channel_id", wo.ChannelID).Msg("Decrypted access token")
	}

	accountID, err := primitive.ObjectIDFromHex(wo.ID)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: invalid account ID: %w", err)
	}
	defer func() {
		if err == nil {
			return
		}
		if recordErr := p.MongoRepo.RecordProcessingError(ctx, accountID, err.Error()); recordErr != nil {
			p.Logger.Warn().Err(recordErr).Str("account_id", wo.ID).Msg("Failed to record processing error")
		}
	}()

	account, err := p.MongoRepo.FindByID(ctx, accountID)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to fetch account from MongoDB: %w", err)
	}
	if account == nil {
		return fmt.Errorf("Processor.ProcessAccount: account not found: %s", wo.ID)
	}
	if mongodb.HasProcessingErrorMeta(account.MetaData) {
		if err := p.MongoRepo.ClearProcessingError(ctx, accountID); err != nil {
			p.Logger.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to clear stale processing error before retry")
		}
	}

	originalState := account.State

	// Refresh token if needed - YouTube tokens expire in 1 hour
	accessToken, err := p.refreshTokenIfNeeded(ctx, wo, account)
	if err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "ProcessAccount").Str("stage", "refresh_token").Msg("Token refresh failed, will try with existing token")
		logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "refresh_token", "channel_id": wo.ChannelID}, nil)
		accessToken = wo.AccessToken
	}
	wo.AccessToken = accessToken

	// Fetch all data from YouTube APIs
	p.Logger.Info().Str("channel_id", wo.ChannelID).Msg("Starting YouTube data fetch")
	fetchedData, err := p.fetchAllData(ctx, wo)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to fetch YouTube data: %w", err)
	}

	// Parse the fetched data
	p.Logger.Info().Str("channel_id", wo.ChannelID).Msg("Parsing YouTube data")
	parsedData, err := p.parseAllData(wo, fetchedData)
	if err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to parse YouTube data: %w", err)
	}

	// Store in ClickHouse
	p.Logger.Info().
		Str("channel_id", wo.ChannelID).
		Int("videos", len(parsedData.Videos)).
		Int("activity_insights", len(parsedData.ActivityInsights)).
		Int("traffic_insights", len(parsedData.TrafficInsights)).
		Msg("Storing YouTube data in ClickHouse")

	if err := p.storeInClickHouse(ctx, wo, parsedData); err != nil {
		return fmt.Errorf("Processor.ProcessAccount: failed to store YouTube data: %w", err)
	}

	// Persist successful completion state and analytics timestamp in MongoDB.
	now := time.Now().UTC()
	if err := p.MongoRepo.UpdateState(ctx, accountID, mongomodels.StateProcessed); err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", wo.ID).Str("function", "ProcessAccount").Msg("Failed to update account state to Processed")
	}
	if err := p.MongoRepo.ClearProcessingError(ctx, accountID); err != nil {
		p.Logger.Warn().Err(err).Str("account_id", wo.ID).Msg("Failed to clear processing error")
	}
	if err := p.MongoRepo.UpdateAnalyticsTimestamp(ctx, accountID, "analytics", now); err != nil {
		p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("account_id", wo.ID).Str("function", "ProcessAccount").Msg("Failed to update last_analytics_updated_at")
	}

	// Send notifications
	userID := account.GetUserIDHex()
	workspaceID := account.GetWorkspaceIDHex()
	if workspaceID == "" {
		workspaceID = wo.WorkspaceID
	}

	p.sendPusherNotification(account, workspaceID, originalState)

	if originalState == mongomodels.StateAdded {
		p.sendEmailNotification(userID, workspaceID, wo.ChannelID, account.PlatformName)
	}

	p.Logger.Info().Str("channel_id", wo.ChannelID).Msg("YouTube account processing completed")
	return nil
}

func (p *Processor) refreshTokenIfNeeded(ctx context.Context, wo WorkOrder, account *mongomodels.SocialIntegration) (string, error) {
	// Try to get refresh token from work order first, then from account ExtraData
	refreshToken := wo.RefreshToken
	source := "work_order"

	if refreshToken == "" && account.ExtraData != nil {
		if rt, ok := account.ExtraData["refresh_token"].(string); ok && rt != "" {
			refreshToken = rt
			source = "extra_data_refresh_token"
		} else if rt, ok := account.ExtraData["refreshToken"].(string); ok && rt != "" {
			refreshToken = rt
			source = "extra_data_refreshToken"
		}
	}

	p.Logger.Info().
		Str("channel_id", wo.ChannelID).
		Bool("has_refresh_token", refreshToken != "").
		Int("refresh_token_len", len(refreshToken)).
		Str("source", source).
		Msg("Checking refresh token availability")

	if refreshToken == "" {
		p.Logger.Warn().Str("channel_id", wo.ChannelID).Msg("No refresh token available, using access token as-is (may be expired)")
		return wo.AccessToken, nil
	}

	// Decrypt refresh token if encrypted
	originalLen := len(refreshToken)
	if decrypted, err := crypto.DecryptToken(refreshToken, p.Cfg.DecryptionKey); err == nil {
		refreshToken = decrypted
		p.Logger.Debug().
			Str("channel_id", wo.ChannelID).
			Int("original_len", originalLen).
			Int("decrypted_len", len(refreshToken)).
			Msg("Decrypted refresh token")
	} else {
		p.Logger.Debug().
			Str("channel_id", wo.ChannelID).
			Err(err).
			Msg("Refresh token not encrypted or decryption failed, using as-is")
	}

	p.Logger.Info().Str("channel_id", wo.ChannelID).Msg("Attempting to refresh YouTube access token")

	tokenResp, err := p.YTClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		return "", fmt.Errorf("Processor.refreshTokenIfNeeded: token refresh failed: %w", err)
	}

	p.Logger.Info().
		Str("channel_id", wo.ChannelID).
		Int("new_token_len", len(tokenResp.AccessToken)).
		Msg("Successfully refreshed YouTube access token")
	return tokenResp.AccessToken, nil
}

// isUnauthorizedError checks if the error indicates a 401 Unauthorized response
func isUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "status 401") || strings.Contains(errStr, "request failed with status 401")
}

func parseYouTubeDateRange(startDateStr, endDateStr string) (time.Time, time.Time, bool, error) {
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

	return startDate.UTC(), endDate.UTC(), true, nil
}

func (p *Processor) fetchAllData(ctx context.Context, wo WorkOrder) (*FetchedData, error) {
	data := &FetchedData{}

	// Determine date ranges based on requested range or sync type.
	startDate, endDate, hasRequestedRange, err := parseYouTubeDateRange(wo.StartDate, wo.EndDate)
	if err != nil {
		return nil, fmt.Errorf("Processor.fetchAllData: invalid date range: %w", err)
	}

	var videosSince time.Time
	var insightsStartDate time.Time

	if hasRequestedRange {
		videosSince = startDate
		insightsStartDate = startDate
	} else if wo.SyncType == kafkamodels.YouTubeSyncTypeFullSync {
		videosSince = time.Date(FullSyncVideosStartYear, 1, 1, 0, 0, 0, 0, time.UTC)
		insightsStartDate = time.Date(FullSyncVideosStartYear, 1, 1, 0, 0, 0, 0, time.UTC)
	} else {
		now := time.Now().UTC()
		// YouTube Analytics API has a 2-3 day data delay, so use 3 days ago as end date
		// This ensures we get complete data for the requested period.
		endDate = now.AddDate(0, 0, -3)
		videosSince = endDate.AddDate(0, 0, -DefaultVideosDays)
		insightsStartDate = endDate.AddDate(0, 0, -DefaultInsightsDays)
	}

	if hasRequestedRange {
		lagCutoff := time.Now().UTC().AddDate(0, 0, -3)
		if endDate.After(lagCutoff) {
			endDate = lagCutoff
		}
		if endDate.Before(startDate) {
			return nil, fmt.Errorf("Processor.fetchAllData: requested end_date is too recent for available YouTube Analytics data")
		}
	}

	var (
		videosMu           sync.Mutex
		activityInsightsMu sync.Mutex
		trafficInsightsMu  sync.Mutex
		sharedInsightsMu   sync.Mutex
	)

	// All API calls run in a single errgroup. Videos wait for channel data
	// via a signal channel, so they run in parallel with insights.
	eg, egCtx := errgroup.WithContext(ctx)
	channelReady := make(chan struct{})

	// Fetch channel data (signals when done so video goroutine can start)
	eg.Go(func() error {
		defer close(channelReady)
		channelResp, err := p.YTClient.FetchChannels(egCtx, wo.AccessToken)
		if err != nil {
			if isUnauthorizedError(err) {
				p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "fetchAllData").Str("stage", "fetch_channels").Msg("Unauthorized - stopping all API calls (expected auth error)")
				return fmt.Errorf("Processor.fetchAllData: %w: %v", ErrUnauthorized, err)
			}
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "fetchAllData").Str("stage", "fetch_channels").Msg("Failed to fetch channel data (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "fetch_channels", "channel_id": wo.ChannelID}, nil)
			return nil
		}
		if len(channelResp.Items) > 0 {
			data.Channel = &channelResp.Items[0]
		}
		return nil
	})

	// Fetch videos + details (waits for channel data, runs parallel with insights)
	eg.Go(func() error {
		select {
		case <-channelReady:
		case <-egCtx.Done():
			return egCtx.Err()
		}
		if data.Channel == nil {
			return nil
		}
		uploadsPlaylistID := data.Channel.ContentDetails.RelatedPlaylists.Uploads
		if uploadsPlaylistID == "" {
			p.Logger.Warn().Str("channel_id", wo.ChannelID).Msg("No uploads playlist ID found, skipping video fetch")
			return nil
		}

		videos, err := p.YTClient.FetchVideos(egCtx, wo.AccessToken, uploadsPlaylistID, videosSince)
		if err != nil {
			if isUnauthorizedError(err) {
				return fmt.Errorf("Processor.fetchAllData: %w: %v", ErrUnauthorized, err)
			}
			p.Logger.Warn().Err(err).Str("channel_id", wo.ChannelID).Str("stage", "fetch_videos").Msg("Failed to fetch videos (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "fetch_videos", "channel_id": wo.ChannelID}, nil)
			return nil
		}

		if hasRequestedRange {
			endExclusive := endDate.AddDate(0, 0, 1)
			filtered := make([]social.YouTubeActivityItem, 0, len(videos))
			for _, video := range videos {
				publishedStr := video.Snippet.PublishedAt
				publishedAt, parseErr := time.Parse(time.RFC3339, publishedStr)
				if parseErr != nil {
					continue
				}
				if publishedAt.Before(startDate) || !publishedAt.Before(endExclusive) {
					continue
				}
				filtered = append(filtered, video)
			}
			videos = filtered
		}

		videosMu.Lock()
		data.Videos = videos
		videosMu.Unlock()

		var videoIDs []string
		for _, v := range videos {
			if v.ContentDetails.Upload.VideoID != "" {
				videoIDs = append(videoIDs, v.ContentDetails.Upload.VideoID)
			}
		}

		if len(videoIDs) > 0 {
			videoDetails, err := p.YTClient.FetchVideoDetails(egCtx, wo.AccessToken, videoIDs)
			if err != nil {
				if isUnauthorizedError(err) {
					return fmt.Errorf("Processor.fetchAllData: %w: %v", ErrUnauthorized, err)
				}
				p.Logger.Warn().Err(err).Str("channel_id", wo.ChannelID).Str("stage", "fetch_video_details").Msg("Failed to fetch video details")
				logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "fetch_video_details", "channel_id": wo.ChannelID}, nil)
			} else {
				videosMu.Lock()
				data.VideoDetails = make(map[string]*social.YouTubeVideoItem)
				for i := range videoDetails {
					data.VideoDetails[videoDetails[i].ID] = &videoDetails[i]
				}
				videosMu.Unlock()
				p.Logger.Info().
					Str("channel_id", wo.ChannelID).
					Int("video_details_count", len(data.VideoDetails)).
					Msg("Fetched video details with lifetime statistics")
			}
		}
		return nil
	})

	// Fetch activity insights (independent)
	eg.Go(func() error {
		activityInsights, err := p.YTClient.FetchActivityInsights(egCtx, wo.AccessToken, insightsStartDate, endDate)
		if err != nil {
			if isUnauthorizedError(err) {
				p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "fetchAllData").Str("stage", "fetch_activity_insights").Msg("Unauthorized - stopping all API calls (expected auth error)")
				return fmt.Errorf("Processor.fetchAllData: %w: %v", ErrUnauthorized, err)
			}
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "fetchAllData").Str("stage", "fetch_activity_insights").Msg("Failed to fetch activity insights (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "fetch_activity_insights", "channel_id": wo.ChannelID}, nil)
			return nil
		}
		activityInsightsMu.Lock()
		data.ActivityInsights = activityInsights
		activityInsightsMu.Unlock()
		return nil
	})

	// Fetch traffic insights (independent)
	eg.Go(func() error {
		trafficInsights, err := p.YTClient.FetchTrafficInsights(egCtx, wo.AccessToken, insightsStartDate, endDate)
		if err != nil {
			if isUnauthorizedError(err) {
				p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "fetchAllData").Str("stage", "fetch_traffic_insights").Msg("Unauthorized - stopping all API calls (expected auth error)")
				return fmt.Errorf("Processor.fetchAllData: %w: %v", ErrUnauthorized, err)
			}
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "fetchAllData").Str("stage", "fetch_traffic_insights").Msg("Failed to fetch traffic insights (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "fetch_traffic_insights", "channel_id": wo.ChannelID}, nil)
			return nil
		}

		trafficInsightsMu.Lock()
		data.TrafficInsights = trafficInsights
		trafficInsightsMu.Unlock()
		return nil
	})

	// Fetch shared insights (independent)
	eg.Go(func() error {
		sharedInsights, err := p.YTClient.FetchSharedInsights(egCtx, wo.AccessToken, insightsStartDate, endDate)
		if err != nil {
			if isUnauthorizedError(err) {
				p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "fetchAllData").Str("stage", "fetch_shared_insights").Msg("Unauthorized - stopping all API calls (expected auth error)")
				return fmt.Errorf("Processor.fetchAllData: %w: %v", ErrUnauthorized, err)
			}
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "fetchAllData").Str("stage", "fetch_shared_insights").Msg("Failed to fetch shared insights (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "fetch_shared_insights", "channel_id": wo.ChannelID}, nil)
			return nil
		}
		sharedInsightsMu.Lock()
		data.SharedInsights = sharedInsights
		sharedInsightsMu.Unlock()
		return nil
	})

	// Wait for all goroutines (channel + videos + insights)
	if err := eg.Wait(); err != nil {
		if errors.Is(err, ErrUnauthorized) {
			return nil, err
		}
		return nil, fmt.Errorf("Processor.fetchAllData: failed to fetch YouTube data: %w", err)
	}

	p.Logger.Info().
		Str("channel_id", wo.ChannelID).
		Bool("has_channel", data.Channel != nil).
		Int("videos_count", len(data.Videos)).
		Int("video_details_count", len(data.VideoDetails)).
		Bool("has_activity_insights", data.ActivityInsights != nil).
		Bool("has_traffic_insights", data.TrafficInsights != nil).
		Bool("has_shared_insights", data.SharedInsights != nil).
		Msg("Completed YouTube data fetch")

	return data, nil
}

func (p *Processor) parseAllData(wo WorkOrder, data *FetchedData) (*ParsedData, error) {
	parsed := &ParsedData{}
	now := time.Now().UTC()

	// Parse channel
	if data.Channel != nil {
		parsed.Channel = p.parseChannel(wo.ChannelID, data.Channel, now)
	}

	// Detect media types efficiently using duration-based detection with parallel HTTP fallback
	var videosForDetection []social.YouTubeVideoItem
	if data.VideoDetails != nil {
		for _, details := range data.VideoDetails {
			if details != nil {
				videosForDetection = append(videosForDetection, *details)
			}
		}
	}
	mediaTypes := p.YTClient.DetectMediaTypes(context.Background(), videosForDetection)

	p.Logger.Info().
		Str("channel_id", wo.ChannelID).
		Int("videos_detected", len(mediaTypes)).
		Msg("Detected video media types")

	// Parse videos with lifetime statistics from Data API
	for _, video := range data.Videos {
		videoID := video.ContentDetails.Upload.VideoID
		if videoID == "" {
			continue
		}
		// Get video details with lifetime stats
		var videoDetails *social.YouTubeVideoItem
		if data.VideoDetails != nil {
			videoDetails = data.VideoDetails[videoID]
		}
		// Get pre-computed media type
		mediaType := mediaTypes[videoID]
		if mediaType == "" {
			mediaType = kafkamodels.YouTubeMediaTypeVideo // Default to video
		}
		parsedVideo := p.parseVideo(wo.ChannelID, &video, videoDetails, mediaType, now)
		parsed.Videos = append(parsed.Videos, *parsedVideo)
	}

	// Parse activity insights
	if data.ActivityInsights != nil && len(data.ActivityInsights.Rows) > 0 {
		parsed.ActivityInsights = p.parseActivityInsights(wo.ChannelID, data.ActivityInsights)
	}

	// Parse traffic insights
	if data.TrafficInsights != nil && len(data.TrafficInsights.Rows) > 0 {
		parsed.TrafficInsights = p.parseTrafficInsights(wo.ChannelID, data.TrafficInsights)
	}

	// Parse shared insights
	if data.SharedInsights != nil && len(data.SharedInsights.Rows) > 0 {
		parsed.SharedInsights = p.parseSharedInsights(wo.ChannelID, data.SharedInsights, now)
	}

	p.Logger.Info().
		Str("channel_id", wo.ChannelID).
		Int("activity_rows", func() int {
			if data.ActivityInsights == nil {
				return 0
			}
			return len(data.ActivityInsights.Rows)
		}()).
		Int("traffic_rows", func() int {
			if data.TrafficInsights == nil {
				return 0
			}
			return len(data.TrafficInsights.Rows)
		}()).
		Int("shared_rows", func() int {
			if data.SharedInsights == nil {
				return 0
			}
			return len(data.SharedInsights.Rows)
		}()).
		Msg("Parsed YouTube insights row counts")

	return parsed, nil
}

func (p *Processor) parseChannel(channelID string, ch *social.YouTubeChannelItem, now time.Time) *kafkamodels.ParsedYouTubeChannel {
	publishedAt, _ := time.Parse(time.RFC3339, ch.Snippet.PublishedAt)
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	subscriberCount, _ := strconv.ParseInt(ch.Statistics.SubscriberCount, 10, 64)
	videoCount, _ := strconv.ParseInt(ch.Statistics.VideoCount, 10, 64)
	viewCount, _ := strconv.ParseInt(ch.Statistics.ViewCount, 10, 64)

	return &kafkamodels.ParsedYouTubeChannel{
		RecordID:        conversions.GenerateYouTubeRecordID(channelID, today),
		ChannelID:       ch.ID,
		Title:           ch.Snippet.Title,
		Description:     ch.Snippet.Description,
		CustomURL:       ch.Snippet.CustomURL,
		ThumbnailURL:    ch.Snippet.Thumbnails.High.URL,
		BannerURL:       ch.BrandingSettings.Image.BannerExternalURL,
		Country:         ch.Snippet.Country,
		SubscriberCount: subscriberCount,
		VideoCount:      videoCount,
		ViewCount:       viewCount,
		PublishedAt:     publishedAt,
		CreatedAt:       today,
		InsertedAt:      now,
	}
}

func (p *Processor) parseVideo(channelID string, video *social.YouTubeActivityItem, details *social.YouTubeVideoItem, mediaType string, now time.Time) *kafkamodels.ParsedYouTubeVideo {
	videoID := video.ContentDetails.Upload.VideoID
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)

	parsed := &kafkamodels.ParsedYouTubeVideo{
		VideoID:         videoID,
		ChannelID:       channelID,
		IframeEmbedHTML: social.GenerateEmbedHTML(videoID),
		MediaType:       mediaType,
		CreatedAt:       today,
		InsertedAt:      now,
	}

	// Use video details with lifetime stats if available
	if details != nil {
		parsed.Title = details.Snippet.Title
		parsed.Description = details.Snippet.Description
		parsed.Duration = details.ContentDetails.Duration

		// Thumbnail URL
		if details.Snippet.Thumbnails.High.URL != "" {
			parsed.ThumbnailURL = details.Snippet.Thumbnails.High.URL
		} else {
			parsed.ThumbnailURL = details.Snippet.Thumbnails.Default.URL
		}

		// Published date
		if pubAt, err := time.Parse(time.RFC3339, details.Snippet.PublishedAt); err == nil {
			parsed.PublishedAt = pubAt
		}

		// Lifetime statistics from Data API
		parsed.Views, _ = strconv.ParseInt(details.Statistics.ViewCount, 10, 64)
		parsed.Likes, _ = strconv.ParseInt(details.Statistics.LikeCount, 10, 64)
		parsed.Dislikes, _ = strconv.ParseInt(details.Statistics.DislikeCount, 10, 64)
		parsed.Comments, _ = strconv.ParseInt(details.Statistics.CommentCount, 10, 64)
		parsed.Favorites, _ = strconv.ParseInt(details.Statistics.FavoriteCount, 10, 64)
	} else {
		// Fallback to activity data
		parsed.Title = video.Snippet.Title
		parsed.Description = video.Snippet.Description
		if video.Snippet.Thumbnails.High.URL != "" {
			parsed.ThumbnailURL = video.Snippet.Thumbnails.High.URL
		} else {
			parsed.ThumbnailURL = video.Snippet.Thumbnails.Default.URL
		}
		if pubAt, err := time.Parse(time.RFC3339, video.Snippet.PublishedAt); err == nil {
			parsed.PublishedAt = pubAt
		}
	}

	return parsed
}

func (p *Processor) parseActivityInsights(channelID string, resp *social.YouTubeAnalyticsResponse) []kafkamodels.ParsedYouTubeActivityInsights {
	var insights []kafkamodels.ParsedYouTubeActivityInsights

	// Build column index map
	colIndex := make(map[string]int)
	for i, col := range resp.ColumnHeaders {
		colIndex[col.Name] = i
	}

	for _, row := range resp.Rows {
		dateStr, ok := row[colIndex["day"]].(string)
		if !ok {
			continue
		}
		date, _ := time.Parse("2006-01-02", dateStr)

		insight := kafkamodels.ParsedYouTubeActivityInsights{
			RecordID:                   conversions.GenerateYouTubeRecordID(channelID, date),
			ChannelID:                  channelID,
			RedViews:                   getInt64FromRow(row, colIndex, "redViews"),
			Views:                      getInt64FromRow(row, colIndex, "views"),
			Likes:                      getInt64FromRow(row, colIndex, "likes"),
			Dislikes:                   getInt64FromRow(row, colIndex, "dislikes"),
			Comments:                   getInt64FromRow(row, colIndex, "comments"),
			Shares:                     getInt64FromRow(row, colIndex, "shares"),
			SubscribersGained:          getInt64FromRow(row, colIndex, "subscribersGained"),
			EstimatedMinutesWatched:    getInt64FromRow(row, colIndex, "estimatedMinutesWatched"),
			EstimatedRedMinutesWatched: getInt64FromRow(row, colIndex, "estimatedRedMinutesWatched"),
			AvgViewDuration:            getInt64FromRow(row, colIndex, "averageViewDuration"),
			AvgViewPercentage:          getFloat64FromRow(row, colIndex, "averageViewPercentage"),
			CreatedAt:                  date,
		}
		insights = append(insights, insight)
	}

	return insights
}

func (p *Processor) parseTrafficInsights(channelID string, resp *social.YouTubeAnalyticsResponse) []kafkamodels.ParsedYouTubeTrafficInsights {
	// Group by date and aggregate traffic sources
	dayData := make(map[string]*kafkamodels.ParsedYouTubeTrafficInsights)
	// Accumulate raw float watch times per day, round only at the end
	subscriberWatchTimeRaw := make(map[string]float64)
	nonSubscriberWatchTimeRaw := make(map[string]float64)

	colIndex := make(map[string]int)
	for i, col := range resp.ColumnHeaders {
		colIndex[col.Name] = i
	}

	for _, row := range resp.Rows {
		dateStr, ok := row[colIndex["day"]].(string)
		if !ok {
			continue
		}
		date, _ := time.Parse("2006-01-02", dateStr)

		if dayData[dateStr] == nil {
			dayData[dateStr] = &kafkamodels.ParsedYouTubeTrafficInsights{
				RecordID:  conversions.GenerateYouTubeRecordID(channelID, date),
				ChannelID: channelID,
				CreatedAt: date,
			}
		}

		trafficSource, _ := row[colIndex["insightTrafficSourceType"]].(string)
		views := getInt64FromRow(row, colIndex, "views")
		rawWatchTime := getFloat64FromRow(row, colIndex, "estimatedMinutesWatched")

		switch strings.ToUpper(trafficSource) {
		case kafkamodels.TrafficSourcePaid:
			dayData[dateStr].PaidViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceAnnotation:
			dayData[dateStr].AnnotationViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceEndScreen:
			dayData[dateStr].EndScreenViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceCampaignCard:
			dayData[dateStr].CampaignCardViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceSubscriber:
			dayData[dateStr].SubscriberViews = views
			subscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceNoLinkOther:
			dayData[dateStr].NoLinkOtherViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceYTChannel:
			dayData[dateStr].YTChannelViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceYTSearch:
			dayData[dateStr].YTSearchViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceRelatedVideo:
			dayData[dateStr].RelatedVideoViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceYTOtherPage:
			dayData[dateStr].YTOtherPageViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceExtURL:
			dayData[dateStr].ExtURLViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourcePlaylist:
			dayData[dateStr].PlaylistViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceNotification:
			dayData[dateStr].NotificationViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceShorts:
			dayData[dateStr].ShortsViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		}
	}

	// Round accumulated watch times and assign to parsed data
	var insights []kafkamodels.ParsedYouTubeTrafficInsights
	for date, insight := range dayData {
		insight.SubscriberWatchTime = int64(math.Round(subscriberWatchTimeRaw[date]))
		insight.NonSubscriberWatchTime = int64(math.Round(nonSubscriberWatchTimeRaw[date]))

		insights = append(insights, *insight)
	}

	return insights
}

func (p *Processor) parseSharedInsights(channelID string, resp *social.YouTubeAnalyticsResponse, now time.Time) *kafkamodels.ParsedYouTubeSharedInsights {
	insight := &kafkamodels.ParsedYouTubeSharedInsights{
		RecordID:   conversions.GenerateYouTubeRecordID(channelID, now),
		ChannelID:  channelID,
		InsertedAt: now,
	}

	colIndex := make(map[string]int)
	for i, col := range resp.ColumnHeaders {
		colIndex[col.Name] = i
	}

	for _, row := range resp.Rows {
		service, _ := row[colIndex["sharingService"]].(string)
		shares := getInt64FromRow(row, colIndex, "shares")

		switch strings.ToUpper(service) {
		case kafkamodels.SharingServiceAmeba:
			insight.Ameba = shares
		case kafkamodels.SharingServiceBlogger:
			insight.Blogger = shares
		case kafkamodels.SharingServiceCopyPaste:
			insight.CopyPaste = shares
		case kafkamodels.SharingServiceCyworld:
			insight.Cyworld = shares
		case kafkamodels.SharingServiceDigg:
			insight.Digg = shares
		case kafkamodels.SharingServiceDropbox:
			insight.Dropbox = shares
		case kafkamodels.SharingServiceEmbed:
			insight.Embed = shares
		case kafkamodels.SharingServiceMail:
			insight.Mail = shares
		case kafkamodels.SharingServiceWhatsApp:
			insight.WhatsApp = shares
		case kafkamodels.SharingServiceOther:
			insight.Other = shares
		case kafkamodels.SharingServiceFacebookMsgr:
			insight.FacebookMsgr = shares
		case kafkamodels.SharingServiceFacebookPages:
			insight.FacebookPages = shares
		case kafkamodels.SharingServiceFacebook:
			insight.Facebook = shares
		case kafkamodels.SharingServiceFotka:
			insight.Fotka = shares
		case kafkamodels.SharingServiceVKontakte:
			insight.VKontakte = shares
		case kafkamodels.SharingServiceDiscord:
			insight.Discord = shares
		case kafkamodels.SharingServiceGooglePlus:
			insight.GooglePlus = shares
		case kafkamodels.SharingServiceGoo:
			insight.Goo = shares
		case kafkamodels.SharingServiceHangouts:
			insight.Hangouts = shares
		case kafkamodels.SharingServiceLinkedIn:
			insight.LinkedIn = shares
		case kafkamodels.SharingServicePinterest:
			insight.Pinterest = shares
		case kafkamodels.SharingServiceMyspace:
			insight.Myspace = shares
		case kafkamodels.SharingServiceReddit:
			insight.Reddit = shares
		case kafkamodels.SharingServiceSkype:
			insight.Skype = shares
		case kafkamodels.SharingServiceTelegram:
			insight.Telegram = shares
		case kafkamodels.SharingServiceTwitter:
			insight.Twitter = shares
		case kafkamodels.SharingServiceTumblr:
			insight.Tumblr = shares
		case kafkamodels.SharingServiceViber:
			insight.Viber = shares
		case kafkamodels.SharingServiceWeibo:
			insight.Weibo = shares
		case kafkamodels.SharingServiceWeChat:
			insight.WeChat = shares
		case kafkamodels.SharingServiceYouTube, kafkamodels.SharingServiceYouTubeGaming,
			kafkamodels.SharingServiceYouTubeKids, kafkamodels.SharingServiceYouTubeMusic,
			kafkamodels.SharingServiceYouTubeTV:
			insight.YouTube += shares
		}
	}

	return insight
}

func (p *Processor) storeInClickHouse(ctx context.Context, wo WorkOrder, data *ParsedData) error {
	// Store channel
	if data.Channel != nil {
		chChannel := conversions.ConvertYouTubeChannel(data.Channel)
		if err := p.Sink.BulkInsertYouTubeChannels(ctx, []*clickhousemodels.YouTubeChannel{chChannel}); err != nil {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "storeInClickHouse").Str("stage", "insert_channel").Msg("Failed to insert channel (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "insert_channel", "channel_id": wo.ChannelID}, nil)
		}
	}

	// Store videos
	if len(data.Videos) > 0 {
		var chVideos []*clickhousemodels.YouTubeVideo
		for _, v := range data.Videos {
			chVideo := conversions.ConvertYouTubeVideo(&v)
			if chVideo != nil {
				chVideos = append(chVideos, chVideo)
			}
		}
		if err := p.Sink.BulkInsertYouTubeVideos(ctx, chVideos); err != nil {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "storeInClickHouse").Str("stage", "insert_videos").Msg("Failed to insert videos (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "insert_videos", "channel_id": wo.ChannelID}, nil)
		}
		p.Logger.Info().Str("channel_id", wo.ChannelID).Int("count", len(chVideos)).Msg("Inserted YouTube videos")
	}

	// Store activity insights
	if len(data.ActivityInsights) > 0 {
		var chInsights []*clickhousemodels.YouTubeActivityInsights
		for _, ins := range data.ActivityInsights {
			chIns := conversions.ConvertYouTubeActivityInsights(&ins)
			if chIns != nil {
				chInsights = append(chInsights, chIns)
			}
		}
		if err := p.Sink.BulkInsertYouTubeActivityInsights(ctx, chInsights); err != nil {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "storeInClickHouse").Str("stage", "insert_activity_insights").Msg("Failed to insert activity insights (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "insert_activity_insights", "channel_id": wo.ChannelID}, nil)
		}
		p.Logger.Info().Str("channel_id", wo.ChannelID).Int("count", len(chInsights)).Msg("Inserted YouTube activity insights")
	}

	// Store traffic insights
	if len(data.TrafficInsights) > 0 {
		var chInsights []*clickhousemodels.YouTubeTrafficInsights
		for _, ins := range data.TrafficInsights {
			chIns := conversions.ConvertYouTubeTrafficInsights(&ins)
			if chIns != nil {
				chInsights = append(chInsights, chIns)
			}
		}
		if err := p.Sink.BulkInsertYouTubeTrafficInsights(ctx, chInsights); err != nil {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "storeInClickHouse").Str("stage", "insert_traffic_insights").Msg("Failed to insert traffic insights (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "insert_traffic_insights", "channel_id": wo.ChannelID}, nil)
		}
		p.Logger.Info().Str("channel_id", wo.ChannelID).Int("count", len(chInsights)).Msg("Inserted YouTube traffic insights")
	}

	// Store shared insights
	if data.SharedInsights != nil {
		chIns := conversions.ConvertYouTubeSharedInsights(data.SharedInsights)
		if err := p.Sink.BulkInsertYouTubeSharedInsights(ctx, []*clickhousemodels.YouTubeSharedInsights{chIns}); err != nil {
			p.Logger.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "storeInClickHouse").Str("stage", "insert_shared_insights").Msg("Failed to insert shared insights (continuing)")
			logger.CaptureException(err, map[string]string{"platform": "youtube", "component": "immediate-processor", "stage": "insert_shared_insights", "channel_id": wo.ChannelID}, nil)
		}
	}

	return nil
}

func (p *Processor) sendPusherNotification(account *mongomodels.SocialIntegration, workspaceID, originalState string) {
	if p.PusherClient == nil {
		return
	}

	accountID := account.PlatformIdentifier
	if accountID == "" {
		accountID = account.GetPlatformID()
	}
	if accountID == "" {
		return
	}

	// Frontend subscribes to yt-analytics-channel-{workspace_id}-{channel_id}.
	channel := fmt.Sprintf("yt-analytics-channel-%s-%s", workspaceID, accountID)
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

func (p *Processor) sendEmailNotification(userID, workspaceID, channelID, channelName string) {
	if p.Notifier == nil {
		return
	}
	if err := p.Notifier.SendAnalyticsNotification(userID, workspaceID, "youtube", channelID, channelName, false); err != nil {
		p.Logger.Warn().
			Err(err).
			Str("error_message", err.Error()).
			Str("user_id", userID).
			Str("workspace_id", workspaceID).
			Str("channel_id", channelID).
			Str("function", "sendEmailNotification").
			Msg("Failed to send analytics notification to backend")
	}
}

func getInt64FromRow(row []interface{}, colIndex map[string]int, colName string) int64 {
	idx, ok := colIndex[colName]
	if !ok || idx >= len(row) {
		return 0
	}

	switch v := row[idx].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	default:
		return 0
	}
}

func getFloat64FromRow(row []interface{}, colIndex map[string]int, colName string) float64 {
	idx, ok := colIndex[colName]
	if !ok || idx >= len(row) {
		return 0
	}

	switch v := row[idx].(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	default:
		return 0
	}
}
