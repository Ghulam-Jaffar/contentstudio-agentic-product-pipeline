package gmb

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/ai"
	"golang.org/x/sync/errgroup"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/gmb"
)

const aiCacheTTL = 24 * time.Hour

var validInsightTypes = map[string]string{
	"impressions_overview": "impressions-overview",
	"actions_overview":     "actions-overview",
	"search_keywords":      "search-keywords",
	"publishing_behavior":  "publishing-behavior",
	"insights_summary":     "insights-summary",
}

// AIInsightsService prepares GMB analytics datasets and forwards them to the AI agent service.
type AIInsightsService struct {
	analyticsService *GMBAnalyticsService
	agentClient      *ai.AgentClient
	cache            redis.Client
}

// NewAIInsightsService constructs a GMB AI insights service with analytics, agent, and cache dependencies.
func NewAIInsightsService(svc *GMBAnalyticsService, agentClient *ai.AgentClient, cache redis.Client) *AIInsightsService {
	return &AIInsightsService{
		analyticsService: svc,
		agentClient:      agentClient,
		cache:            cache,
	}
}

// GetAIInsights validates the requested insight type, builds the required dataset,
// and delegates insight generation to the external AI agent.
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

	gmbReq := s.toGMBRequest(req)

	var payload map[string]interface{}
	var err error

	switch req.Type {
	case "impressions_overview":
		payload, err = s.buildImpressionsPayload(ctx, gmbReq)
	case "actions_overview":
		payload, err = s.buildActionsPayload(ctx, gmbReq)
	case "search_keywords":
		payload, err = s.buildSearchKeywordsPayload(ctx, gmbReq)
	case "publishing_behavior":
		payload, err = s.buildPublishingPayload(ctx, gmbReq)
	case "insights_summary":
		payload, err = s.buildSummaryPayload(ctx, gmbReq)
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

	language := req.Language
	if language == "" {
		language = "en"
	}
	payload["language"] = language
	payload["timezone"] = req.Timezone

	response, err := s.agentClient.Request(ctx, "gmb/"+agentEndpoint, payload)
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

// toGMBRequest converts the flexible AI date input into the standard GMB analytics request shape.
func (s *AIInsightsService) toGMBRequest(req *types.AIInsightsRequest) *types.GMBRequest {
	var startDate, endDate string

	switch v := req.Date.(type) {
	case []interface{}:
		if len(v) >= 2 {
			startDate = strings.TrimSpace(fmt.Sprintf("%v", v[0]))
			endDate = strings.TrimSpace(fmt.Sprintf("%v", v[1]))
		}
	case string:
		parts := strings.SplitN(v, " - ", 2)
		if len(parts) == 2 {
			startDate = strings.TrimSpace(parts[0])
			endDate = strings.TrimSpace(parts[1])
		}
	}

	if startDate == "" {
		startDate = strings.TrimSpace(req.StartDate)
	}
	if endDate == "" {
		endDate = strings.TrimSpace(req.EndDate)
	}

	return &types.GMBRequest{
		WorkspaceID: req.WorkspaceID,
		GmbID:       req.GmbID,
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    req.Timezone,
	}
}

// buildImpressionsPayload prepares the impressions dataset expected by the GMB AI endpoint.
func (s *AIInsightsService) buildImpressionsPayload(ctx context.Context, req *types.GMBRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetImpressions(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Impressions == nil || len(resp.Impressions.Buckets) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"buckets":        resp.Impressions.Buckets,
		"desktop_maps":   resp.Impressions.DesktopMapsDaily,
		"desktop_search": resp.Impressions.DesktopSearchDaily,
		"mobile_maps":    resp.Impressions.MobileMapsDaily,
		"mobile_search":  resp.Impressions.MobileSearchDaily,
	}

	return map[string]interface{}{"dataset": data}, nil
}

// buildActionsPayload prepares the actions-over-time dataset for the AI agent.
func (s *AIInsightsService) buildActionsPayload(ctx context.Context, req *types.GMBRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetActions(ctx, req)
	if err != nil {
		return nil, err
	}

	if resp.Actions == nil || len(resp.Actions.Buckets) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"buckets":            resp.Actions.Buckets,
		"call_clicks":        resp.Actions.CallClicksDaily,
		"website_clicks":     resp.Actions.WebsiteClicksDaily,
		"direction_requests": resp.Actions.DirectionRequestsDaily,
		"other_actions":      resp.Actions.OtherActionsDaily,
	}

	return map[string]interface{}{"dataset": data}, nil
}

// buildSearchKeywordsPayload converts keyword rows into a compact agent-friendly dataset.
func (s *AIInsightsService) buildSearchKeywordsPayload(ctx context.Context, req *types.GMBRequest) (map[string]interface{}, error) {
	kwReq := &types.SearchKeywordsRequest{GMBRequest: *req}
	resp, err := s.analyticsService.GetSearchKeywords(ctx, kwReq)
	if err != nil {
		return nil, err
	}

	if len(resp.Keywords) == 0 {
		return nil, nil
	}

	keywords := make([]map[string]interface{}, len(resp.Keywords))
	for i, kw := range resp.Keywords {
		keywords[i] = map[string]interface{}{
			"keyword":               kw.Keyword,
			"impressions_value":     kw.ImpressionsValue,
			"impressions_threshold": kw.ImpressionsThreshold,
			"keyword_month":         kw.KeywordMonth,
		}
	}

	return map[string]interface{}{"dataset": map[string]interface{}{"keywords": keywords}}, nil
}

// buildPublishingPayload prepares the publishing activity dataset, including topic type rollups.
func (s *AIInsightsService) buildPublishingPayload(ctx context.Context, req *types.GMBRequest) (map[string]interface{}, error) {
	resp, err := s.analyticsService.GetPublishingBehavior(ctx, req)
	if err != nil {
		return nil, err
	}

	pub := resp.PublishingBehaviour
	if pub == nil || len(pub.PostCount) == 0 {
		return nil, nil
	}

	data := map[string]interface{}{
		"buckets":     pub.Buckets,
		"post_count":  pub.PostCount,
		"topic_types": pub.TopicTypes,
	}

	return map[string]interface{}{"dataset": data}, nil
}

// buildSummaryPayload assembles the combined GMB dataset used for the summary insight prompt.
func (s *AIInsightsService) buildSummaryPayload(ctx context.Context, req *types.GMBRequest) (map[string]interface{}, error) {
	var (
		impressions *types.ImpressionsResponse
		actions     *types.ActionsResponse
		keywords    *types.SearchKeywordsResponse
		publishing  *types.PublishingBehaviorResponse
		topPosts    *types.TopPostsResponse
	)

	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var err error
		impressions, err = s.analyticsService.GetImpressions(egCtx, req)
		return err
	})
	eg.Go(func() error {
		var err error
		actions, err = s.analyticsService.GetActions(egCtx, req)
		return err
	})
	eg.Go(func() error {
		kwReq := &types.SearchKeywordsRequest{GMBRequest: *req}
		resp, err := s.analyticsService.GetSearchKeywords(egCtx, kwReq)
		if err == nil {
			keywords = resp
		}
		return nil
	})
	eg.Go(func() error {
		resp, err := s.analyticsService.GetPublishingBehavior(egCtx, req)
		if err == nil {
			publishing = resp
		}
		return nil
	})
	eg.Go(func() error {
		topReq := &types.TopPostsRequest{GMBRequest: *req, Limit: 5}
		resp, err := s.analyticsService.GetTopPosts(egCtx, topReq)
		if err == nil {
			topPosts = resp
		}
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	if (impressions.Impressions == nil || len(impressions.Impressions.Buckets) == 0) &&
		(actions.Actions == nil || len(actions.Actions.Buckets) == 0) {
		return nil, nil
	}

	payload := map[string]interface{}{
		"gmb_impressions_data": s.impressionsToMap(impressions),
		"gmb_actions_data":     s.actionsToMap(actions),
	}

	if keywords != nil {
		payload["search_keywords"] = keywords.Keywords
	}
	if publishing != nil && publishing.PublishingBehaviour != nil {
		payload["publishing_behaviour"] = publishing.PublishingBehaviour
	}
	if topPosts != nil {
		payload["top_posts"] = topPosts.Posts
	}

	return payload, nil
}

func (s *AIInsightsService) impressionsToMap(resp *types.ImpressionsResponse) map[string]interface{} {
	if resp.Impressions == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"buckets":        resp.Impressions.Buckets,
		"desktop_maps":   resp.Impressions.DesktopMapsDaily,
		"desktop_search": resp.Impressions.DesktopSearchDaily,
		"mobile_maps":    resp.Impressions.MobileMapsDaily,
		"mobile_search":  resp.Impressions.MobileSearchDaily,
	}
}

func (s *AIInsightsService) actionsToMap(resp *types.ActionsResponse) map[string]interface{} {
	if resp.Actions == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"buckets":            resp.Actions.Buckets,
		"call_clicks":        resp.Actions.CallClicksDaily,
		"website_clicks":     resp.Actions.WebsiteClicksDaily,
		"direction_requests": resp.Actions.DirectionRequestsDaily,
		"other_actions":      resp.Actions.OtherActionsDaily,
	}
}

func (s *AIInsightsService) buildCacheKey(req *types.AIInsightsRequest) string {
	dateStr := fmt.Sprintf("%v", req.Date)
	lang := req.Language
	if lang == "" {
		lang = "en"
	}
	return fmt.Sprintf("gmb_AI:%s:%s:%s:%s", req.Type, req.GmbID, dateStr, lang)
}
