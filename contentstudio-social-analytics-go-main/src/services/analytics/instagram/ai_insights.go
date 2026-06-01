package instagram

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/instagram"
	"golang.org/x/sync/errgroup"
)

// aiCacheTTL is the duration for which AI insight responses are cached in Redis.
const aiCacheTTL = 24 * time.Hour

// validInsightTypes maps the API-facing insight type key to the AI agent sub-path segment.
// Only keys present here are accepted; all others return a validation error.
var validInsightTypes = map[string]string{
	"impressions":                      "impressions",
	"engagement":                       "engagement",
	"publishing_behaviour_impressions": "publishing-impressions",
	"publishing_behaviour_engagements": "publishing-engagement",
	"publishing_behaviour_reach":       "publishing-reach",
	"audience_growth":                  "audience-growth",
	"reels_engagement":                 "reels-engagement",
	"reels_watch_time":                 "reels-watch-time",
	"reels_shares":                     "reels-shares",
	"stories_interactions":             "stories-interactions",
	"stories_impressions":              "stories-impressions",
	"stories_reach":                    "stories-reach",
	"top_posts":                        "top-posts",
	"top_hashtags":                     "hashtags",
	"insights_summary":                 "insights-summary",
}

// AIInsightsService orchestrates Instagram analytics data retrieval, AI agent requests,
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

	baseReq, err := s.toInstagramRequest(req)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	switch req.Type {
	case "impressions":
		payload, err = s.buildImpressionsPayload(ctx, baseReq)
	case "engagement":
		payload, err = s.buildEngagementPayload(ctx, baseReq)
	case "publishing_behaviour_impressions", "publishing_behaviour_engagements", "publishing_behaviour_reach":
		payload, err = s.buildPublishingPayload(ctx, baseReq, req.Type)
	case "audience_growth":
		payload, err = s.buildAudienceGrowthPayload(ctx, baseReq)
	case "reels_engagement", "reels_watch_time", "reels_shares":
		payload, err = s.buildReelsPayload(ctx, baseReq, req.Type)
	case "stories_interactions", "stories_impressions", "stories_reach":
		payload, err = s.buildStoriesPayload(ctx, baseReq, req.Type)
	case "top_posts":
		payload, err = s.buildTopPostsPayload(ctx, baseReq, req.Limit)
	case "top_hashtags":
		payload, err = s.buildTopHashtagsPayload(ctx, baseReq)
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

	response, err := s.agentClient.Request(ctx, "instagram/"+agentEndpoint, payload)
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

// toInstagramRequest parses the raw date field and constructs a typed InstagramRequest
// suitable for passing to the underlying analytics service methods.
func (s *AIInsightsService) toInstagramRequest(req *types.AIInsightsRequest) (*types.InstagramRequest, error) {
	startDate, endDate, err := parseDateRange(req.Date)
	if err != nil {
		return nil, httputil.NewValidationError(err.Error())
	}
	return &types.InstagramRequest{
		WorkspaceID: req.WorkspaceID,
		InstagramID: req.InstagramID,
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    timezoneOrDefault(req.Timezone),
	}, nil
}

// buildImpressionsPayload fetches page impression data from instagram_insights and builds
// a dataset payload; returns nil when ShowData is 0 indicating no data for the period.
func (s *AIInsightsService) buildImpressionsPayload(ctx context.Context, req *types.InstagramRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetImpressions(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Impressions == nil || resp.Impressions.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":         resp.Impressions.ShowData,
		"buckets":           resp.Impressions.Buckets,
		"impressions":       resp.Impressions.Impressions,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildEngagementPayload fetches engagement time-series data and builds a dataset payload
// containing total engagement, comments, and reactions; returns nil when ShowData is 0.
func (s *AIInsightsService) buildEngagementPayload(ctx context.Context, req *types.InstagramRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetEngagement(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Engagements == nil || resp.Engagements.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":         resp.Engagements.ShowData,
		"buckets":           resp.Engagements.Buckets,
		"engagement":        resp.Engagements.Engagement,
		"comments":          resp.Engagements.Comments,
		"reactions":         resp.Engagements.Reactions,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildPublishingPayload fetches publishing behaviour data and builds a dataset payload whose
// metric fields vary by insightType: impressions, engagement fields, or reach.
func (s *AIInsightsService) buildPublishingPayload(ctx context.Context, req *types.InstagramRequest, insightType string) (map[string]interface{}, error) {
	pbReq := &types.PublishingBehaviourRequest{InstagramRequest: *req}
	resp, err := s.analyticsService.GetPublishingBehaviour(ctx, pbReq)
	if err != nil {
		return nil, err
	}
	if resp.PublishingBehaviour == nil || sumInt32(resp.PublishingBehaviour.TotalPosts) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"total_posts":       resp.PublishingBehaviour.TotalPosts,
		"buckets":           resp.PublishingBehaviour.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}
	switch insightType {
	case "publishing_behaviour_impressions":
		data["impressions"] = resp.PublishingBehaviour.Impressions
	case "publishing_behaviour_engagements":
		data["likes"] = resp.PublishingBehaviour.Likes
		data["comments"] = resp.PublishingBehaviour.Comments
		data["saved"] = resp.PublishingBehaviour.Saved
		data["engagement"] = resp.PublishingBehaviour.Engagement
	case "publishing_behaviour_reach":
		data["reach"] = resp.PublishingBehaviour.Reach
	}
	return buildDatasetPayload(data), nil
}

// buildAudienceGrowthPayload fetches audience growth data and builds a dataset payload
// containing cumulative follower counts and daily deltas; returns nil when ShowData is 0.
func (s *AIInsightsService) buildAudienceGrowthPayload(ctx context.Context, req *types.InstagramRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetAudienceGrowth(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.AudienceGrowth == nil || resp.AudienceGrowth.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":         resp.AudienceGrowth.ShowData,
		"followers":         resp.AudienceGrowth.Followers,
		"followers_daily":   resp.AudienceGrowth.FollowersDaily,
		"buckets":           resp.AudienceGrowth.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildReelsPayload fetches reels performance data and builds a dataset payload whose
// metric fields vary by insightType: engagement, watch time, or shares.
func (s *AIInsightsService) buildReelsPayload(ctx context.Context, req *types.InstagramRequest, insightType string) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetReelsPerformance(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Reels == nil || resp.Reels.ShowData == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"total_posts":       resp.Reels.TotalPosts,
		"buckets":           resp.Reels.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}
	switch insightType {
	case "reels_engagement":
		data["engagement"] = resp.Reels.Engagement
		data["likes"] = resp.Reels.Likes
		data["comments"] = resp.Reels.Comments
		data["saves"] = resp.Reels.Saves
	case "reels_watch_time":
		data["avg_watch_time"] = resp.Reels.AvgWatchTime
		data["total_watch_time"] = resp.Reels.TotalWatchTime
	case "reels_shares":
		data["shares"] = resp.Reels.Shares
	}
	return buildDatasetPayload(data), nil
}

// buildStoriesPayload fetches stories performance data and builds a dataset payload whose
// metric fields vary by insightType: interaction counts, impressions, or reach.
func (s *AIInsightsService) buildStoriesPayload(ctx context.Context, req *types.InstagramRequest, insightType string) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetStoriesPerformance(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.StoriesPerformance == nil || resp.StoriesPerformance.ShowData == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"buckets":           resp.StoriesPerformance.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}
	switch insightType {
	case "stories_interactions":
		data["story_reply"] = resp.StoriesPerformance.StoryReply
		data["story_exits"] = resp.StoriesPerformance.StoryExits
		data["story_taps_forward"] = resp.StoriesPerformance.StoryTapsForward
		data["story_taps_back"] = resp.StoriesPerformance.StoryTapsBack
		data["published_stories"] = resp.StoriesPerformance.PublishedStories
	case "stories_impressions":
		data["story_impressions"] = resp.StoriesPerformance.StoryImpressions
		data["avg_story_impressions"] = resp.StoriesPerformance.AvgStoryImpressions
	case "stories_reach":
		data["story_reach"] = resp.StoriesPerformance.StoryReach
	}
	return buildDatasetPayload(data), nil
}

// buildTopPostsPayload fetches the top posts list (defaulting to 15 when limit is unset)
// and wraps the result in a dataset payload; returns nil when no posts are found.
func (s *AIInsightsService) buildTopPostsPayload(ctx context.Context, req *types.InstagramRequest, limit int) (map[string]interface{}, error) {
	topReq := &types.TopPostsRequest{
		InstagramRequest: *req,
		Limit:            limit,
	}
	if topReq.Limit <= 0 {
		topReq.Limit = 15
	}

	resp, err := s.analyticsService.GetTopPosts(ctx, topReq)
	if err != nil {
		return nil, err
	}
	if len(resp.TopPosts) == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"status":    true,
		"top_posts": resp.TopPosts,
	}), nil
}

// buildTopHashtagsPayload fetches hashtag analytics and builds a dataset payload containing
// the top-hashtags list and its rollup; returns nil when no hashtag posts are found.
func (s *AIInsightsService) buildTopHashtagsPayload(ctx context.Context, req *types.InstagramRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetHashtags(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.TopHashtags == nil || sumInt32(resp.TopHashtags.Posts) == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"status":              true,
		"top_hashtags":        resp.TopHashtags,
		"top_hashtags_rollup": resp.TopHashtagsRollup,
	}), nil
}

// buildSummaryPayload fetches seven analytics datasets concurrently via errgroup and assembles
// a combined summary payload; returns nil when no top posts exist (used as the data-available guard).
func (s *AIInsightsService) buildSummaryPayload(ctx context.Context, req *types.InstagramRequest, limit int) (map[string]interface{}, error) {
	var (
		summary    *types.SummaryResponse
		publishing *types.PublishingBehaviourResponse
		audience   *types.AudienceGrowthResponse
		topPosts   *types.TopPostsResponse
		hashtags   *types.HashtagsResponse
		reels      *types.ReelsPerformanceResponse
		stories    *types.StoriesPerformanceResponse
	)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		summary, err = s.analyticsService.GetSummary(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		publishing, err = s.analyticsService.GetPublishingBehaviour(egCtx, &types.PublishingBehaviourRequest{InstagramRequest: *req})
		return err
	})
	eg.Go(func() error {
		var err error
		audience, err = s.analyticsService.GetAudienceGrowth(egCtx, req)
		return err
	})
	eg.Go(func() error {
		topReq := &types.TopPostsRequest{InstagramRequest: *req, Limit: limit}
		if topReq.Limit <= 0 {
			topReq.Limit = 15
		}
		var err error
		topPosts, err = s.analyticsService.GetTopPosts(egCtx, topReq)
		return err
	})
	eg.Go(func() error {
		var err error
		hashtags, err = s.analyticsService.GetHashtags(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		reels, err = s.analyticsService.GetReelsPerformance(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		stories, err = s.analyticsService.GetStoriesPerformance(egCtx, req)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if topPosts == nil || len(topPosts.TopPosts) == 0 {
		return nil, nil
	}

	return buildDatasetPayload(map[string]interface{}{
		"account_data":         summary,
		"publishing_behaviour": publishing,
		"audience_growth":      audience,
		"top_posts":            topPosts,
		"top_hashtags":         hashtags,
		"reels":                reels,
		"stories":              stories,
	}), nil
}

// buildCacheKey returns the Redis key for caching an AI insight response.
// Format: ig_AI:{type}:{instagram_id}:{start},{end}:{language}
func (s *AIInsightsService) buildCacheKey(req *types.AIInsightsRequest) string {
	return fmt.Sprintf(
		"ig_AI:%s:%s:%s:%s",
		req.Type,
		req.InstagramID,
		dateCacheValue(req.Date),
		languageOrDefault(req.Language),
	)
}

// buildDatasetPayload wraps data under a "dataset" key and promotes the "aggregation_level"
// field to the top-level "insight_type" key when present, as expected by the AI agent.
func buildDatasetPayload(data map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{"dataset": data}
	if aggregationLevel, ok := data["aggregation_level"]; ok {
		payload["insight_type"] = aggregationLevel
	}
	return payload
}

// parseDateRange accepts a date value in multiple formats ([]interface{}, []string, or a
// comma/pipe/" - " delimited string) and returns the ISO-8601 start and end date strings.
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

// dateCacheValue formats the raw date value as "start,end" for use in a Redis cache key,
// falling back to the raw string representation if parsing fails.
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
// This is forwarded to the AI agent as insight_type to guide response granularity.
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

// sumInt32 returns the sum of all values in the slice; used to detect empty time-series data.
func sumInt32(values []int32) int32 {
	var total int32
	for _, v := range values {
		total += v
	}
	return total
}
