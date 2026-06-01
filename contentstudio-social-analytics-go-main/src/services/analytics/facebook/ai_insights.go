package facebook

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/facebook"
	"golang.org/x/sync/errgroup"
)

// aiCacheTTL is the Redis TTL for cached AI insight responses.
const aiCacheTTL = 24 * time.Hour

// validInsightTypes maps the API insight type key to its agent endpoint path segment.
var validInsightTypes = map[string]string{
	"page_impressions":                 "page-impressions",
	"page_engagement":                  "page-engagement",
	"publishing_behaviour_impressions": "publishing-impressions",
	"publishing_behaviour_engagements": "publishing-engagement",
	"publishing_behaviour_reach":       "publishing-reach",
	"audience_growth":                  "audience-growth",
	"video_views":                      "video-views",
	"video_watch_time":                 "video-watch-time",
	"video_engagements":                "video-engagement",
	"reels_initial_plays":              "reels-plays",
	"reels_watch_time":                 "reels-watch-time",
	"reels_engagement":                 "reels-engagement",
	"top_posts":                        "top-posts",
	"insights_summary":                 "insights-summary",
}

// AIInsightsService orchestrates AI-powered analytics insights for Facebook.
// It fetches analytics data, packages it as a dataset payload, sends to an AI agent
// endpoint, and caches the result in Redis for 24 hours.
type AIInsightsService struct {
	analyticsService Service
	agentClient      agentRequester
	cache            redis.Client
}

// agentRequester is the interface for making requests to the AI agent HTTP service.
type agentRequester interface {
	Request(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error)
}

// NewAIInsightsService creates a new AIInsightsService with the given dependencies.
func NewAIInsightsService(svc Service, agentClient agentRequester, cache redis.Client) *AIInsightsService {
	return &AIInsightsService{
		analyticsService: svc,
		agentClient:      agentClient,
		cache:            cache,
	}
}

// GetAIInsights validates the insight type, checks Redis cache, fetches analytics data,
// builds the dataset payload, calls the AI agent, caches the response, and returns it.
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

	fbReq, err := s.toFacebookRequest(req)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	switch req.Type {
	case "page_impressions":
		payload, err = s.buildImpressionsPayload(ctx, fbReq)
	case "page_engagement":
		payload, err = s.buildEngagementPayload(ctx, fbReq)
	case "publishing_behaviour_impressions", "publishing_behaviour_engagements", "publishing_behaviour_reach":
		payload, err = s.buildPublishingPayload(ctx, fbReq, req.Type)
	case "audience_growth":
		payload, err = s.buildAudienceGrowthPayload(ctx, fbReq)
	case "video_views", "video_watch_time", "video_engagements":
		payload, err = s.buildVideoPayload(ctx, fbReq, req.Type)
	case "reels_initial_plays", "reels_watch_time", "reels_engagement":
		payload, err = s.buildReelsPayload(ctx, fbReq, req.Type)
	case "top_posts":
		payload, err = s.buildTopPostsPayload(ctx, fbReq)
	case "insights_summary":
		payload, err = s.buildSummaryPayload(ctx, fbReq)
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

	response, err := s.agentClient.Request(ctx, "facebook/"+agentEndpoint, payload)
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

// toFacebookRequest converts an AIInsightsRequest to a FacebookRequest, parsing the
// flexible date field and applying timezone/language defaults.
func (s *AIInsightsService) toFacebookRequest(req *types.AIInsightsRequest) (*types.FacebookRequest, error) {
	startDate, endDate, err := parseDateRange(req.Date)
	if err != nil {
		return nil, httputil.NewValidationError(err.Error())
	}

	return &types.FacebookRequest{
		WorkspaceID: req.WorkspaceID,
		FacebookIDs: []string{req.FacebookID},
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    timezoneOrDefault(req.Timezone),
		Limit:       req.Limit,
	}, nil
}

// buildImpressionsPayload fetches page impression data and wraps it in a dataset payload.
// Returns nil when there is no impression data (show_data == 0).
func (s *AIInsightsService) buildImpressionsPayload(ctx context.Context, req *types.FacebookRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetImpressions(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Impressions == nil || sumInt32(resp.Impressions.PageImpressions) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"page_impressions":  resp.Impressions.PageImpressions,
		"buckets":           resp.Impressions.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}
	return buildDatasetPayload(data), nil
}

// buildEngagementPayload fetches page engagement data and wraps it in a dataset payload.
// Returns nil when there is no engagement data.
func (s *AIInsightsService) buildEngagementPayload(ctx context.Context, req *types.FacebookRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetEngagement(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Engagement == nil || resp.Engagement.Engagement == nil || sumInt32(resp.Engagement.Engagement.PageEngagements) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"page_engagements":  resp.Engagement.Engagement.PageEngagements,
		"buckets":           resp.Engagement.Engagement.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}
	return buildDatasetPayload(data), nil
}

// buildPublishingPayload fetches publishing behaviour data and selects the subset of fields
// relevant to the requested insightType (impressions, engagements, or reach).
func (s *AIInsightsService) buildPublishingPayload(ctx context.Context, req *types.FacebookRequest, insightType string) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetPublishingBehaviour(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.PublishingBehaviour == nil || sumInt32(resp.PublishingBehaviour.PostCount) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"buckets":           resp.PublishingBehaviour.Buckets,
		"post_count":        resp.PublishingBehaviour.PostCount,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}

	switch insightType {
	case "publishing_behaviour_impressions":
		data["paid_impressions"] = resp.PublishingBehaviour.PaidImpressions
		data["organic_impressions"] = resp.PublishingBehaviour.OrganicImpressions
		data["viral_impressions"] = resp.PublishingBehaviour.ViralImpressions
	case "publishing_behaviour_engagements":
		data["reactions_engagement"] = resp.PublishingBehaviour.ReactionsEngagement
		data["comments_engagement"] = resp.PublishingBehaviour.CommentsEngagement
		data["shares_engagement"] = resp.PublishingBehaviour.SharesEngagement
	case "publishing_behaviour_reach":
		data["paid_reach"] = resp.PublishingBehaviour.PaidReach
		data["organic_reach"] = resp.PublishingBehaviour.OrganicReach
		data["viral_reach"] = resp.PublishingBehaviour.ViralReach
	}

	return buildDatasetPayload(data), nil
}

// buildAudienceGrowthPayload fetches fan growth time-series data and wraps it in a dataset payload.
// Returns nil when fan_count sums to zero (no fan data available).
func (s *AIInsightsService) buildAudienceGrowthPayload(ctx context.Context, req *types.FacebookRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetAudienceGrowth(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.AudienceGrowth == nil || sumInt32(resp.AudienceGrowth.FanCount) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"show_data":           resp.AudienceGrowth.ShowData,
		"fan_count":           resp.AudienceGrowth.FanCount,
		"page_fans_daily":     resp.AudienceGrowth.PageFansDaily,
		"page_fans_by_like":   resp.AudienceGrowth.PageFansByLike,
		"page_fans_by_unlike": resp.AudienceGrowth.PageFansByUnlike,
		"page_impressions":    resp.AudienceGrowth.PageImpressions,
		"page_engagements":    resp.AudienceGrowth.PageEngagements,
		"buckets":             resp.AudienceGrowth.Buckets,
		"aggregation_level":   aggregationLevel(req.StartDate, req.EndDate),
	}
	return buildDatasetPayload(data), nil
}

// buildVideoPayload fetches video insights and selects the fields relevant to the
// requested insightType (video_views, video_watch_time, or video_engagements).
func (s *AIInsightsService) buildVideoPayload(ctx context.Context, req *types.FacebookRequest, insightType string) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetVideoInsights(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.VideoInsights == nil || sumInt32(resp.VideoInsights.TotalPosts) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"buckets":           resp.VideoInsights.Buckets,
		"total_posts":       resp.VideoInsights.TotalPosts,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}

	switch insightType {
	case "video_views":
		data["total_views"] = resp.VideoInsights.TotalViews
		data["organic_views"] = resp.VideoInsights.OrganicViews
		data["paid_views"] = resp.VideoInsights.PaidViews
	case "video_watch_time":
		data["total_watch_time"] = resp.VideoInsights.TotalViewTime
		data["organic_watch_time"] = resp.VideoInsights.OrganicViewTime
		data["paid_watch_time"] = resp.VideoInsights.PaidViewTime
	case "video_engagements":
		data["comments"] = resp.VideoInsights.Comments
		data["reactions"] = resp.VideoInsights.Reactions
		data["shares"] = resp.VideoInsights.Shares
	}

	return buildDatasetPayload(data), nil
}

// buildReelsPayload fetches reels analytics and selects fields for the requested
// insightType (reels_initial_plays, reels_watch_time, or reels_engagement).
func (s *AIInsightsService) buildReelsPayload(ctx context.Context, req *types.FacebookRequest, insightType string) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetReelsAnalytics(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Reels == nil || sumInt32(resp.Reels.TotalReels) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"buckets":           resp.Reels.Buckets,
		"total_posts":       resp.Reels.TotalReels,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}

	switch insightType {
	case "reels_initial_plays":
		data["total_views"] = resp.Reels.InitialPlays
	case "reels_watch_time":
		data["total_watch_time"] = resp.Reels.TotalSecondsWatched
	case "reels_engagement":
		data["comments"] = resp.Reels.Comments
		data["reactions"] = resp.Reels.Reactions
		data["shares"] = resp.Reels.Shares
	}

	return buildDatasetPayload(data), nil
}

// buildTopPostsPayload fetches the top posts and wraps them in a dataset payload.
// Returns nil when no posts are found in the period.
func (s *AIInsightsService) buildTopPostsPayload(ctx context.Context, req *types.FacebookRequest) (map[string]interface{}, error) {
	topReq := *req
	topReq.Limit = req.GetLimit(15)

	resp, err := s.analyticsService.GetTopPosts(ctx, &topReq)
	if err != nil {
		return nil, err
	}
	if len(resp.TopPosts) == 0 {
		return nil, nil
	}

	return buildDatasetPayload(map[string]interface{}{
		"top_posts": resp.TopPosts,
	}), nil
}

// buildSummaryPayload fetches all major analytics datasets concurrently and packages them
// into a single summary payload for the AI agent. Returns nil when top posts are empty
// (used as the minimum data requirement).
func (s *AIInsightsService) buildSummaryPayload(ctx context.Context, req *types.FacebookRequest) (map[string]interface{}, error) {
	var (
		summary    *types.SummaryResponse
		reels      *types.ReelsAnalyticsResponse
		publishing *types.PublishingBehaviourResponse
		video      *types.VideoInsightsResponse
		topPosts   *types.TopPostsResponse
	)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		summary, err = s.analyticsService.GetSummary(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		reels, err = s.analyticsService.GetReelsAnalytics(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		publishing, err = s.analyticsService.GetPublishingBehaviour(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		video, err = s.analyticsService.GetVideoInsights(egCtx, req)
		return err
	})
	eg.Go(func() error {
		topReq := *req
		topReq.Limit = req.GetLimit(15)
		var err error
		topPosts, err = s.analyticsService.GetTopPosts(egCtx, &topReq)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if topPosts == nil || len(topPosts.TopPosts) == 0 {
		return nil, nil
	}

	return map[string]interface{}{
		"facebook_page_data": summaryOverview(summary),
		"reels_data": map[string]interface{}{
			"current": reelsCurrent(reels),
			"summary": reelsSummary(reels),
		},
		"publishing_behaviour": map[string]interface{}{
			"current": publishingCurrent(publishing),
			"summary": publishingSummary(publishing),
		},
		"facebook_video_data": map[string]interface{}{
			"current": videoCurrent(video),
			"summary": videoSummary(video),
		},
		"top_posts": topPosts.TopPosts,
	}, nil
}

// buildCacheKey constructs the Redis cache key for a given AI insights request.
// Format: fb_AI:{type}:{facebook_id}:{start},{end}:{language}
func (s *AIInsightsService) buildCacheKey(req *types.AIInsightsRequest) string {
	return fmt.Sprintf(
		"fb_AI:%s:%s:%s:%s",
		req.Type,
		req.FacebookID,
		dateCacheValue(req.Date),
		languageOrDefault(req.Language),
	)
}

func buildDatasetPayload(data map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{"dataset": data}
	if aggregationLevel, ok := data["aggregation_level"]; ok {
		payload["insight_type"] = aggregationLevel
	}
	return payload
}

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

func normalizeDateString(value interface{}) string {
	raw := strings.TrimSpace(fmt.Sprintf("%v", value))
	if len(raw) >= 10 {
		return raw[:10]
	}
	return raw
}

func dateCacheValue(raw interface{}) string {
	start, end, err := parseDateRange(raw)
	if err == nil && start != "" && end != "" {
		return start + "," + end
	}
	return fmt.Sprintf("%v", raw)
}

func languageOrDefault(language string) string {
	if language == "" {
		return "en"
	}
	return language
}

func timezoneOrDefault(timezone string) string {
	if timezone == "" {
		return "UTC"
	}
	return timezone
}

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

func sumInt32(values []int32) int32 {
	var total int32
	for _, value := range values {
		total += value
	}
	return total
}

func summaryOverview(resp *types.SummaryResponse) interface{} {
	if resp == nil || resp.Overview == nil {
		return nil
	}
	return resp.Overview["current"]
}

func reelsCurrent(resp *types.ReelsAnalyticsResponse) interface{} {
	if resp == nil {
		return nil
	}
	return resp.Reels
}

func reelsSummary(resp *types.ReelsAnalyticsResponse) interface{} {
	if resp == nil || resp.ReelsRollup == nil {
		return nil
	}
	return resp.ReelsRollup["current"]
}

func publishingCurrent(resp *types.PublishingBehaviourResponse) interface{} {
	if resp == nil {
		return nil
	}
	return resp.PublishingBehaviour
}

func publishingSummary(resp *types.PublishingBehaviourResponse) interface{} {
	if resp == nil || resp.PublishingBehaviourRollup == nil {
		return nil
	}
	return resp.PublishingBehaviourRollup["current"]
}

func videoCurrent(resp *types.VideoInsightsResponse) interface{} {
	if resp == nil {
		return nil
	}
	return resp.VideoInsights
}

func videoSummary(resp *types.VideoInsightsResponse) interface{} {
	if resp == nil || resp.VideoRollup == nil {
		return nil
	}
	return resp.VideoRollup["current"]
}
