package tiktok

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/tiktok"
	"golang.org/x/sync/errgroup"
)

const aiCacheTTL = 24 * time.Hour

var validInsightTypes = map[string]string{
	"audience_growth":             "cumulative-followers",
	"top_posts":                   "top-performing-posts",
	"daily_engagement":            "daily-engagement",
	"cumulative_engagement":       "cumulative-engagement",
	"daily_video_views":           "daily-views",
	"cumulative_video_views":      "cumulative-video-views",
	"engagement_vs_daily_posting": "posting-engagement-correlation",
	"insights_summary":            "insights-overview",
}

type agentRequester interface {
	Request(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error)
}

type AIInsightsService struct {
	analyticsService Service
	agentClient      agentRequester
	cache            redis.Client
}

func NewAIInsightsService(svc Service, agentClient agentRequester, cache redis.Client) *AIInsightsService {
	return &AIInsightsService{
		analyticsService: svc,
		agentClient:      agentClient,
		cache:            cache,
	}
}

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

	baseReq, err := s.toTiktokRequest(req)
	if err != nil {
		return nil, err
	}

	var payload map[string]interface{}
	switch req.Type {
	case "audience_growth":
		payload, err = s.buildAudienceGrowthPayload(ctx, baseReq)
	case "top_posts":
		payload, err = s.buildTopPostsPayload(ctx, baseReq)
	case "daily_engagement":
		payload, err = s.buildDailyEngagementPayload(ctx, baseReq)
	case "cumulative_engagement":
		payload, err = s.buildCumulativeEngagementPayload(ctx, baseReq)
	case "daily_video_views":
		payload, err = s.buildDailyViewsPayload(ctx, baseReq)
	case "cumulative_video_views":
		payload, err = s.buildCumulativeViewsPayload(ctx, baseReq)
	case "engagement_vs_daily_posting":
		payload, err = s.buildEngagementVsPostingPayload(ctx, baseReq)
	case "insights_summary":
		payload, err = s.buildSummaryPayload(ctx, baseReq)
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

	response, err := s.agentClient.Request(ctx, "tiktok/"+agentEndpoint, payload)
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

func (s *AIInsightsService) toTiktokRequest(req *types.AIInsightsRequest) (*types.TiktokRequest, error) {
	startDate, endDate, err := parseDateRange(req.Date)
	if err != nil {
		return nil, httputil.NewValidationError(err.Error())
	}
	return &types.TiktokRequest{
		WorkspaceID: req.WorkspaceID,
		TiktokID:    req.TiktokID,
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    timezoneOrDefault(req.Timezone),
	}, nil
}

func (s *AIInsightsService) buildAudienceGrowthPayload(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	data, err := s.analyticsService.GetDynamicPageFollowersAndViews(ctx, req)
	if err != nil {
		return nil, err
	}
	if !hasPositiveSeries(data, "followers_count_diff") {
		return nil, nil
	}
	return map[string]interface{}{"dataset": data}, nil
}

func (s *AIInsightsService) buildTopPostsPayload(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	data, err := s.analyticsService.GetTopAndLeastPerformingPosts(ctx, req)
	if err != nil {
		return nil, err
	}
	top := nestedArray(data, "data", "top_posts")
	if len(top) == 0 {
		return nil, nil
	}
	return map[string]interface{}{"dataset": data}, nil
}

func (s *AIInsightsService) buildDailyEngagementPayload(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	data, err := s.analyticsService.GetDynamicDailyEngagementsData(ctx, req)
	if err != nil {
		return nil, err
	}
	if !hasNonEmptySeries(data, "daily_engagement") {
		return nil, nil
	}
	return map[string]interface{}{"dataset": data}, nil
}

func (s *AIInsightsService) buildCumulativeEngagementPayload(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	return s.buildDailyEngagementPayload(ctx, req)
}

func (s *AIInsightsService) buildDailyViewsPayload(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	data, err := s.analyticsService.GetDynamicPageFollowersAndViews(ctx, req)
	if err != nil {
		return nil, err
	}
	if !hasPositiveSeries(data, "followers_count_diff") {
		return nil, nil
	}
	return map[string]interface{}{"dataset": data}, nil
}

func (s *AIInsightsService) buildCumulativeViewsPayload(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	return s.buildDailyViewsPayload(ctx, req)
}

func (s *AIInsightsService) buildEngagementVsPostingPayload(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	data, err := s.analyticsService.GetDynamicPageFollowersAndViews(ctx, req)
	if err != nil {
		return nil, err
	}
	if !hasPositiveSeries(data, "followers_count_diff") {
		return nil, nil
	}
	return map[string]interface{}{"dataset": data}, nil
}

func (s *AIInsightsService) buildSummaryPayload(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	var (
		publishingData      map[string]interface{}
		audienceGrowthData  map[string]interface{}
		topPostsData        map[string]interface{}
		dailyEngagementData map[string]interface{}
	)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		publishingData, err = s.analyticsService.GetPostsAndEngagements(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		audienceGrowthData, err = s.analyticsService.GetPageFollowersAndViews(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		topPostsData, err = s.analyticsService.GetTopAndLeastPerformingPosts(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		dailyEngagementData, err = s.analyticsService.GetDailyEngagementsData(egCtx, req)
		return err
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if len(nestedArray(topPostsData, "data", "top_posts")) == 0 {
		return nil, nil
	}

	return map[string]interface{}{
		"dataset": map[string]interface{}{
			"publishing_data":  publishingData,
			"audience_growth":  audienceGrowthData,
			"top_posts":        topPostsData,
			"daily_engagement": dailyEngagementData,
		},
	}, nil
}

func (s *AIInsightsService) buildCacheKey(req *types.AIInsightsRequest) string {
	return fmt.Sprintf(
		"tt_AI:%s:%s:%s:%s",
		req.Type,
		req.TiktokID,
		dateCacheValue(req.Date),
		languageOrDefault(req.Language),
	)
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

func nestedArray(data map[string]interface{}, topKey, nestedKey string) []interface{} {
	top, ok := data[topKey].(map[string]interface{})
	if !ok {
		return nil
	}
	values, ok := top[nestedKey].([]interface{})
	if ok {
		return values
	}
	typedValues, ok := top[nestedKey].([]map[string]interface{})
	if !ok {
		return nil
	}
	result := make([]interface{}, 0, len(typedValues))
	for _, value := range typedValues {
		result = append(result, value)
	}
	return result
}

func firstDataRow(data map[string]interface{}) map[string]interface{} {
	rows, ok := data["data"].([]map[string]interface{})
	if ok && len(rows) > 0 {
		return rows[0]
	}

	anyRows, ok := data["data"].([]interface{})
	if !ok || len(anyRows) == 0 {
		return nil
	}
	row, _ := anyRows[0].(map[string]interface{})
	return row
}

func hasNonEmptySeries(data map[string]interface{}, key string) bool {
	row := firstDataRow(data)
	if row == nil {
		return false
	}
	series, ok := row[key].([]interface{})
	if ok {
		return len(series) > 0
	}
	typedSeries, ok := row[key].([]int64)
	if ok {
		return len(typedSeries) > 0
	}
	return false
}

func hasPositiveSeries(data map[string]interface{}, key string) bool {
	row := firstDataRow(data)
	if row == nil {
		return false
	}
	if ints, ok := row[key].([]int64); ok {
		var total int64
		for _, v := range ints {
			total += v
		}
		return total > 0
	}
	if values, ok := row[key].([]interface{}); ok {
		var total float64
		for _, item := range values {
			switch n := item.(type) {
			case int:
				total += float64(n)
			case int64:
				total += float64(n)
			case float64:
				total += n
			}
		}
		return total > 0
	}
	return false
}
