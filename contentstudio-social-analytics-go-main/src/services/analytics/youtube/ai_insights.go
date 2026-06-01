package youtube

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/youtube"
	"golang.org/x/sync/errgroup"
)

// aiCacheTTL is the duration for which AI insight responses are cached in Redis.
const aiCacheTTL = 24 * time.Hour

// validInsightTypes maps the API-facing insight type key to the AI agent sub-path segment.
// Only keys present here are accepted; all others return a validation error.
var validInsightTypes = map[string]string{
	"subscribers_trend":          "cumulative-subscribers-trend",
	"daily_views":                "daily-views",
	"daily_engagement":           "daily-engagement",
	"daily_watch_time":           "daily-watch-time",
	"viewers_find_videos":        "traffic-sources",
	"engagement_vs_posting_pattern": "posting-patterns",
	"sharing_services":           "sharing-trends",
	"top_and_least_posts":        "top-least-performing-posts",
	"insights_summary":           "overview-summary",
}

// AIInsightsService orchestrates YouTube analytics data retrieval, AI agent requests,
// and Redis caching for the AI-powered insights endpoints.
type AIInsightsService struct {
	analyticsService Service
	agentClient      agentRequester
	cache            redis.Client
}

// agentRequester is the interface for sending a dataset payload to an AI agent endpoint
// and receiving the generated insight response.
type agentRequester interface {
	Request(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error)
}

// NewAIInsightsService constructs an AIInsightsService wiring together the analytics service,
// the AI agent client, and the Redis cache used for 24-hour response caching.
func NewAIInsightsService(svc Service, agentClient agentRequester, cache redis.Client) *AIInsightsService {
	return &AIInsightsService{
		analyticsService: svc,
		agentClient:      agentClient,
		cache:            cache,
	}
}

// GetAIInsights validates the insight type, checks Redis for a cached result, fetches the
// appropriate analytics data, sends it to the AI agent, and caches and returns the response.
// Returns {"success": false, "message": "insufficient data"} when the dataset is empty.
func (s *AIInsightsService) GetAIInsights(ctx context.Context, req *types.AIInsightsRequest) (map[string]interface{}, error) {
	agentEndpoint, ok := validInsightTypes[req.Type]
	if !ok {
		return nil, httputil.NewValidationError("invalid insight type: " + req.Type)
	}
	if s.analyticsService == nil || s.agentClient == nil {
		return nil, httputil.NewInternalError("ai insights service not configured")
	}

	cacheKey := s.buildCacheKey(req)
	if s.cache != nil {
		if cached, err := s.cache.Get(ctx, cacheKey); err == nil && cached != "" {
			var result map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &result); err == nil {
				result["success"] = true
				return result, nil
			}
		}
	}

	baseReq, err := s.toYoutubeRequest(req)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	switch req.Type {
	case "subscribers_trend":
		payload, err = s.buildSubscriberTrendPayload(ctx, baseReq)
	case "daily_views":
		payload, err = s.buildViewsTrendPayload(ctx, baseReq)
	case "daily_engagement":
		payload, err = s.buildEngagementTrendPayload(ctx, baseReq)
	case "daily_watch_time":
		payload, err = s.buildWatchTimeTrendPayload(ctx, baseReq)
	case "viewers_find_videos":
		payload, err = s.buildFindVideoPayload(ctx, baseReq)
	case "engagement_vs_posting_pattern":
		payload, err = s.buildPerformanceSchedulePayload(ctx, baseReq)
	case "sharing_services":
		payload, err = s.buildVideoSharingPayload(ctx, baseReq)
	case "top_and_least_posts":
		payload, err = s.buildTopVideosPayload(ctx, baseReq, req.Limit)
	case "insights_summary":
		payload, err = s.buildSummaryPayload(ctx, baseReq, req.Limit)
	}
	if err != nil {
		return nil, err
	}
	if payload == nil {
		return map[string]interface{}{
			"success": false,
			"message": "insufficient data",
		}, nil
	}

	payload["language"] = languageOrDefault(req.Language)
	payload["timezone"] = timezoneOrDefault(req.Timezone)

	response, err := s.agentClient.Request(ctx, "youtube/"+agentEndpoint, payload)
	if err != nil {
		return nil, fmt.Errorf("ai agent request: %w", err)
	}

	if s.cache != nil {
		if data, err := json.Marshal(response); err == nil {
			_ = s.cache.Set(ctx, cacheKey, string(data), aiCacheTTL)
		}
	}

	response["success"] = true
	return response, nil
}

// toYoutubeRequest parses the raw date field and constructs a typed YoutubeRequest.
func (s *AIInsightsService) toYoutubeRequest(req *types.AIInsightsRequest) (*types.YoutubeRequest, error) {
	startDate, endDate, err := parseDateRange(req.Date)
	if err != nil {
		return nil, httputil.NewValidationError(err.Error())
	}
	return &types.YoutubeRequest{
		WorkspaceID: req.WorkspaceID,
		YoutubeID:   req.YoutubeID,
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    timezoneOrDefault(req.Timezone),
	}, nil
}

// buildSubscriberTrendPayload fetches subscriber trend data and builds a dataset payload;
// returns nil when ShowData is 0.
func (s *AIInsightsService) buildSubscriberTrendPayload(ctx context.Context, req *types.YoutubeRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetDynamicSubscriberTrend(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":               resp.ShowData,
		"subscribers_gained_daily": resp.SubscribersGainedDaily,
		"subscribers_total":       resp.SubscribersTotal,
		"buckets":                 resp.Buckets,
		"aggregation_level":       aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildEngagementTrendPayload fetches engagement trend data and builds a dataset payload;
// returns nil when ShowData is 0.
func (s *AIInsightsService) buildEngagementTrendPayload(ctx context.Context, req *types.YoutubeRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetDynamicEngagementTrend(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":        resp.ShowData,
		"like_daily":       resp.LikeDaily,
		"like_total":       resp.LikeTotal,
		"dislike_daily":    resp.DislikeDaily,
		"dislike_total":    resp.DislikeTotal,
		"share_daily":      resp.ShareDaily,
		"share_total":      resp.ShareTotal,
		"comment_daily":    resp.CommentDaily,
		"comment_total":    resp.CommentTotal,
		"engagement_daily": resp.EngagementDaily,
		"engagement_total": resp.EngagementTotal,
		"buckets":          resp.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildViewsTrendPayload fetches views trend data and builds a dataset payload;
// returns nil when ShowData is 0.
func (s *AIInsightsService) buildViewsTrendPayload(ctx context.Context, req *types.YoutubeRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetDynamicViewsTrend(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":                  resp.ShowData,
		"subscriber_views_daily":     resp.SubscriberViewsDaily,
		"subscriber_views_total":     resp.SubscriberViewsTotal,
		"non_subscriber_views_daily": resp.NonSubscriberViewsDaily,
		"non_subscriber_views_total": resp.NonSubscriberViewsTotal,
		"video_views_daily":          resp.VideoViewsDaily,
		"video_views_total":          resp.VideoViewsTotal,
		"buckets":                    resp.Buckets,
		"aggregation_level":          aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildWatchTimeTrendPayload fetches watch time trend data and builds a dataset payload;
// returns nil when ShowData is 0.
func (s *AIInsightsService) buildWatchTimeTrendPayload(ctx context.Context, req *types.YoutubeRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetDynamicWatchTimeTrend(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":                       resp.ShowData,
		"subscriber_watch_time_daily":     resp.SubscriberWatchTimeDaily,
		"subscriber_watch_time_total":     resp.SubscriberWatchTimeTotal,
		"non_subscriber_watch_time_daily": resp.NonSubscriberWatchTimeDaily,
		"non_subscriber_watch_time_total": resp.NonSubscriberWatchTimeTotal,
		"average_watch_time":              resp.AverageWatchTime,
		"buckets":                         resp.Buckets,
		"aggregation_level":               aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildTopVideosPayload fetches top videos and wraps in a dataset payload;
// returns nil when no videos are found.
func (s *AIInsightsService) buildTopVideosPayload(ctx context.Context, req *types.YoutubeRequest, limit int) (map[string]interface{}, error) {
	topReq := &types.TopVideosRequest{
		YoutubeRequest: *req,
		Limit:          limit,
	}
	if topReq.Limit <= 0 {
		topReq.Limit = 15
	}
	resp, err := s.analyticsService.GetSortedTopVideos(ctx, topReq)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"status": true,
		"data":   resp.Data,
	}), nil
}

// buildPerformanceSchedulePayload fetches performance and schedule data and builds a payload;
// returns nil when both engagement and views have no data.
func (s *AIInsightsService) buildPerformanceSchedulePayload(ctx context.Context, req *types.YoutubeRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetPerformanceAndSchedule(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Engagement == nil || (resp.Engagement.ShowData == 0 && resp.VideoViews.ShowData == 0) {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"engagement":  resp.Engagement,
		"video_views": resp.VideoViews,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildSummaryPayload fetches multiple datasets concurrently and assembles a combined payload;
// returns nil when no top videos exist.
func (s *AIInsightsService) buildSummaryPayload(ctx context.Context, req *types.YoutubeRequest, limit int) (map[string]interface{}, error) {
	var (
		summary     *types.SummaryResponse
		subsTrend   *types.SubscriberTrendResponse
		engTrend    *types.EngagementTrendResponse
		viewsTrend  *types.ViewsTrendResponse
		topVideos   *types.SortedTopVideosResponse
		performance *types.PerformanceScheduleResponse
	)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		summary, err = s.analyticsService.GetSummary(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		subsTrend, err = s.analyticsService.GetDynamicSubscriberTrend(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		engTrend, err = s.analyticsService.GetDynamicEngagementTrend(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		viewsTrend, err = s.analyticsService.GetDynamicViewsTrend(egCtx, req)
		return err
	})
	eg.Go(func() error {
		topReq := &types.TopVideosRequest{YoutubeRequest: *req, Limit: limit}
		if topReq.Limit <= 0 {
			topReq.Limit = 15
		}
		var err error
		topVideos, err = s.analyticsService.GetSortedTopVideos(egCtx, topReq)
		return err
	})
	eg.Go(func() error {
		var err error
		performance, err = s.analyticsService.GetPerformanceAndSchedule(egCtx, req)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if topVideos == nil || len(topVideos.Data) == 0 {
		return nil, nil
	}

	return buildDatasetPayload(map[string]interface{}{
		"summary":           summary,
		"subscriber_trend":  subsTrend,
		"engagement_trend":  engTrend,
		"views_trend":       viewsTrend,
		"top_videos":        topVideos,
		"performance":       performance,
	}), nil
}

// buildFindVideoPayload fetches traffic source data and builds a dataset payload;
// returns nil when no data is found.
func (s *AIInsightsService) buildFindVideoPayload(ctx context.Context, req *types.YoutubeRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetFindVideo(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"status": resp.Status,
		"data":   resp.Data,
	}), nil
}

// buildVideoSharingPayload fetches video sharing breakdown data and builds a dataset payload;
// returns nil when no data is found.
func (s *AIInsightsService) buildVideoSharingPayload(ctx context.Context, req *types.YoutubeRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetVideoSharing(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"status": resp.Status,
		"data":   resp.Data,
	}), nil
}

// buildCacheKey returns the Redis key for caching an AI insight response.
// Format: yt_AI:{type}:{youtube_id}:{start},{end}:{language}
func (s *AIInsightsService) buildCacheKey(req *types.AIInsightsRequest) string {
	return fmt.Sprintf(
		"yt_AI:%s:%s:%s:%s",
		req.Type,
		req.YoutubeID,
		dateCacheValue(req.Date),
		languageOrDefault(req.Language),
	)
}

// buildDatasetPayload wraps data under a "dataset" key and promotes the "aggregation_level"
// field to the top-level "insight_type" key when present.
func buildDatasetPayload(data map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{"dataset": data}
	if level, ok := data["aggregation_level"]; ok {
		payload["insight_type"] = level
	}
	return payload
}

// parseDateRange accepts a date value in multiple formats and returns ISO-8601 start and end date strings.
func parseDateRange(raw interface{}) (string, string, error) {
	switch value := raw.(type) {
	case []interface{}:
		if len(value) < 2 {
			return "", "", fmt.Errorf("date must include start and end values")
		}
		return normalizeDateString(value[0]), normalizeDateString(value[1]), nil
	case []string:
		if len(value) < 2 {
			return "", "", fmt.Errorf("date must include start and end values")
		}
		return normalizeDateString(value[0]), normalizeDateString(value[1]), nil
	case string:
		parts := strings.FieldsFunc(value, func(r rune) bool {
			return r == ',' || r == '|'
		})
		if len(parts) < 2 && strings.Contains(value, " - ") {
			parts = strings.SplitN(value, " - ", 2)
		}
		if len(parts) < 2 {
			return "", "", fmt.Errorf("date must be in a supported range format")
		}
		return normalizeDateString(parts[0]), normalizeDateString(parts[1]), nil
	default:
		return "", "", fmt.Errorf("unsupported date format")
	}
}

// normalizeDateString trims whitespace and truncates to the first 10 characters (YYYY-MM-DD).
func normalizeDateString(value interface{}) string {
	raw := strings.TrimSpace(fmt.Sprintf("%v", value))
	if len(raw) >= 10 {
		return raw[:10]
	}
	return raw
}

// dateCacheValue formats the raw date value as "start,end" for use in a Redis cache key.
func dateCacheValue(raw interface{}) string {
	start, end, err := parseDateRange(raw)
	if err == nil && start != "" && end != "" {
		return start + "," + end
	}
	return fmt.Sprintf("%v", raw)
}

// languageOrDefault returns the given language or "en" when the field is empty.
func languageOrDefault(language string) string {
	if language == "" {
		return "en"
	}
	return language
}

// timezoneOrDefault returns the given timezone or "UTC" when the field is empty.
func timezoneOrDefault(timezone string) string {
	if timezone == "" {
		return "UTC"
	}
	return timezone
}

// aggregationLevel returns "monthly" when the date range exceeds 60 days, otherwise "daily".
func aggregationLevel(startDate, endDate string) string {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return "daily"
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return "daily"
	}
	if end.Sub(start).Hours()/24 > 60 {
		return "monthly"
	}
	return "daily"
}
