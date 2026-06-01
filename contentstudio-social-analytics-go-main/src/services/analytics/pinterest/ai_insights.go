package pinterest

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/pinterest"
	"golang.org/x/sync/errgroup"
)

// aiCacheTTL is the duration for which AI insight responses are cached in Redis.
const aiCacheTTL = 24 * time.Hour

// validInsightTypes maps the API-facing insight type key to the AI agent sub-path segment.
// Only keys present here are accepted; all others return a validation error.
var validInsightTypes = map[string]string{
	"daily_engagement":               "daily-engagement",
	"daily_impressions":              "daily-impressions",
	"impressions_vs_posting_pattern": "impressions-posting-pattern",
	"engagement_vs_posting_pattern":  "engagement-posting-pattern",
	"daily_pin_posting":              "daily-pin-posting",
	"daily_followers_trend":          "daily-followers",
	"cumulative_followers_trend":     "cumulative-followers",
	"top_and_least_posts":            "top-performing-posts",
	"insights_summary":               "overview-insights",
}

// AIInsightsService orchestrates Pinterest analytics data retrieval, AI agent requests,
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

	baseReq, err := s.toPinterestRequest(req)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	switch req.Type {
	case "daily_engagement":
		payload, err = s.buildEngagementTrendPayload(ctx, baseReq)
	case "daily_impressions":
		payload, err = s.buildImpressionsTrendPayload(ctx, baseReq)
	case "impressions_vs_posting_pattern":
		payload, err = s.buildImpressionsVsPostingPayload(ctx, baseReq)
	case "engagement_vs_posting_pattern":
		payload, err = s.buildEngagementVsPostingPayload(ctx, baseReq)
	case "daily_pin_posting":
		payload, err = s.buildPinPostingPayload(ctx, baseReq)
	case "daily_followers_trend", "cumulative_followers_trend":
		payload, err = s.buildFollowerTrendPayload(ctx, baseReq)
	case "top_and_least_posts":
		payload, err = s.buildTopPinsPayload(ctx, baseReq, req.Limit)
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

	response, err := s.agentClient.Request(ctx, "pinterest/"+agentEndpoint, payload)
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

// toPinterestRequest parses the raw date field and constructs a typed PinterestRequest
// suitable for passing to the underlying analytics service methods.
func (s *AIInsightsService) toPinterestRequest(req *types.AIInsightsRequest) (*types.PinterestRequest, error) {
	startDate, endDate, err := parseDateRange(req.Date)
	if err != nil {
		return nil, httputil.NewValidationError(err.Error())
	}
	return &types.PinterestRequest{
		WorkspaceID: req.WorkspaceID,
		PinterestID: req.PinterestID,
		BoardID:     req.BoardID,
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    timezoneOrDefault(req.Timezone),
	}, nil
}

// buildFollowerTrendPayload fetches follower trend data and builds a dataset payload;
// returns nil when ShowData is 0 indicating no data for the period.
func (s *AIInsightsService) buildFollowerTrendPayload(ctx context.Context, req *types.PinterestRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetDynamicFollowerTrend(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":         resp.ShowData,
		"followers_daily":   resp.FollowersDaily,
		"followers_gained":  resp.FollowersGained,
		"buckets":           resp.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildImpressionsTrendPayload fetches impressions trend data and builds a dataset payload;
// returns nil when ShowData is 0.
func (s *AIInsightsService) buildImpressionsTrendPayload(ctx context.Context, req *types.PinterestRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetDynamicImpressionsTrend(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":         resp.ShowData,
		"impressions_daily": resp.ImpressionsDaily,
		"impressions_total": resp.ImpressionsTotal,
		"buckets":           resp.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildEngagementTrendPayload fetches engagement trend data and builds a dataset payload;
// returns nil when ShowData is 0.
func (s *AIInsightsService) buildEngagementTrendPayload(ctx context.Context, req *types.PinterestRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetDynamicEngagementTrend(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":             resp.ShowData,
		"saves_daily":           resp.SavesDaily,
		"saves_total":           resp.SavesTotal,
		"outbound_clicks_daily": resp.OutboundClicksDaily,
		"outbound_clicks_total": resp.OutboundClicksTotal,
		"pin_clicks_daily":      resp.PinClicksDaily,
		"pin_clicks_total":      resp.PinClicksTotal,
		"engagement_daily":      resp.EngagementDaily,
		"engagement_total":      resp.EngagementTotal,
		"buckets":               resp.Buckets,
		"aggregation_level":     aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildPinPostingPayload fetches pin posting data and builds a dataset payload;
// returns nil when ShowData is 0.
func (s *AIInsightsService) buildPinPostingPayload(ctx context.Context, req *types.PinterestRequest) (map[string]interface{}, error) {
	filteredReq := &types.FilteredPinRequest{PinterestRequest: *req}
	resp, err := s.analyticsService.GetDynamicPinPosting(ctx, filteredReq)
	if err != nil {
		return nil, err
	}
	if resp.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"show_data":         resp.ShowData,
		"pins_count":        resp.PinsCount,
		"buckets":           resp.Buckets,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildPinRollupPayload fetches pin rollup data and builds a dataset payload;
// returns nil when no pins exist for the period.
func (s *AIInsightsService) buildPinRollupPayload(ctx context.Context, req *types.PinterestRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetPinRollup(ctx, req)
	if err != nil {
		return nil, err
	}
	if resp.Overview == nil || resp.Overview.Current == nil || resp.Overview.Current.TotalPins == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"status":   true,
		"overview": resp.Overview,
	}), nil
}

// buildTopPinsPayload fetches the top pins list and wraps it in a dataset payload;
// returns nil when no pins are found.
func (s *AIInsightsService) buildTopPinsPayload(ctx context.Context, req *types.PinterestRequest, limit int) (map[string]interface{}, error) {
	topReq := &types.TopPinsRequest{
		PinterestRequest: *req,
		Limit:            limit,
	}
	if topReq.Limit <= 0 {
		topReq.Limit = 15
	}

	resp, err := s.analyticsService.GetTopPins(ctx, topReq)
	if err != nil {
		return nil, err
	}
	if len(resp.Top) == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"status": true,
		"top":    resp.Top,
		"least":  resp.Least,
	}), nil
}

// buildSummaryPayload fetches multiple analytics datasets concurrently via errgroup and assembles
// a combined summary payload; returns nil when no top pins exist.
func (s *AIInsightsService) buildSummaryPayload(ctx context.Context, req *types.PinterestRequest, limit int) (map[string]interface{}, error) {
	var (
		summary     *types.SummaryResponse
		follower    *types.FollowerTrendResponse
		impressions *types.ImpressionsTrendResponse
		engagement  *types.EngagementTrendResponse
		pinRollup   *types.PinRollupResponse
		topPins     *types.TopPinsResponse
	)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		summary, err = s.analyticsService.GetSummary(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		follower, err = s.analyticsService.GetDynamicFollowerTrend(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		impressions, err = s.analyticsService.GetDynamicImpressionsTrend(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		engagement, err = s.analyticsService.GetDynamicEngagementTrend(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		pinRollup, err = s.analyticsService.GetPinRollup(egCtx, req)
		return err
	})
	eg.Go(func() error {
		topReq := &types.TopPinsRequest{PinterestRequest: *req, Limit: limit}
		if topReq.Limit <= 0 {
			topReq.Limit = 15
		}
		var err error
		topPins, err = s.analyticsService.GetTopPins(egCtx, topReq)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if topPins == nil || len(topPins.Top) == 0 {
		return nil, nil
	}

	return buildDatasetPayload(map[string]interface{}{
		"account_data":      summary,
		"follower_trend":    follower,
		"impressions_trend": impressions,
		"engagement_trend":  engagement,
		"pin_rollup":        pinRollup,
		"top_pins":          topPins,
	}), nil
}

// buildImpressionsVsPostingPayload concurrently fetches impressions and pin posting data,
// combining them for the impressions-vs-posting-pattern agent endpoint.
func (s *AIInsightsService) buildImpressionsVsPostingPayload(ctx context.Context, req *types.PinterestRequest) (map[string]interface{}, error) {
	filteredReq := &types.FilteredPinRequest{PinterestRequest: *req}
	var impressions *types.ImpressionsTrendResponse
	var pinPosting *types.PinPostingResponse
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		impressions, err = s.analyticsService.GetDynamicImpressionsTrend(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		pinPosting, err = s.analyticsService.GetDynamicPinPosting(egCtx, filteredReq)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if impressions.ShowData == 0 || pinPosting.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"impressions":       impressions,
		"pin_posting":       pinPosting,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildEngagementVsPostingPayload concurrently fetches engagement and pin posting data,
// combining them for the engagement-vs-posting-pattern agent endpoint.
func (s *AIInsightsService) buildEngagementVsPostingPayload(ctx context.Context, req *types.PinterestRequest) (map[string]interface{}, error) {
	filteredReq := &types.FilteredPinRequest{PinterestRequest: *req}
	var engagement *types.EngagementTrendResponse
	var pinPosting *types.PinPostingResponse
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		engagement, err = s.analyticsService.GetDynamicEngagementTrend(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		pinPosting, err = s.analyticsService.GetDynamicPinPosting(egCtx, filteredReq)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if engagement.ShowData == 0 {
		return nil, nil
	}
	return buildDatasetPayload(map[string]interface{}{
		"engagement":        engagement,
		"pin_posting":       pinPosting,
		"aggregation_level": aggregationLevel(req.StartDate, req.EndDate),
	}), nil
}

// buildCacheKey returns the Redis key for caching an AI insight response.
// Format: pt_AI:{type}:{pinterest_id}:{board_id or 'user'}:{start},{end}:{language}
func (s *AIInsightsService) buildCacheKey(req *types.AIInsightsRequest) string {
	boardPart := "user"
	if req.BoardID != "" {
		boardPart = req.BoardID
	}
	return fmt.Sprintf(
		"pt_AI:%s:%s:%s:%s:%s",
		req.Type,
		req.PinterestID,
		boardPart,
		dateCacheValue(req.Date),
		languageOrDefault(req.Language),
	)
}

// buildDatasetPayload wraps data under a "dataset" key and promotes the "aggregation_level"
// field to the top-level "insight_type" key when present, as expected by the AI agent.
func buildDatasetPayload(data map[string]interface{}) map[string]interface{} {
	payload := map[string]interface{}{"dataset": data}
	if aggLevel, ok := data["aggregation_level"]; ok {
		payload["insight_type"] = aggLevel
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
	if strings.EqualFold(timezone, "Europe/Kyiv") {
		return "Europe/Riga"
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
