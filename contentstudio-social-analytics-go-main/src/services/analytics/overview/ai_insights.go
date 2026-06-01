package overview

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/overview"
	"golang.org/x/sync/errgroup"
)

const aiCacheTTL = 24 * time.Hour

var validInsightTypes = map[string]string{
	"reach_across_platforms":         "analytics/overview/reach-performance",
	"engagement_across_platforms":    "analytics/overview/engagement-performance",
	"impressions_across_platforms":   "analytics/overview/impressions-performance",
	"platform_performance_comparison": "analytics/overview/performance-comparison",
	"accounts_statistics":            "analytics/overview/account-statistics",
	"top_posts":                      "analytics/overview/top-posts",
}

type agentRequester interface {
	Request(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error)
}

// AIInsightsService orchestrates overview analytics data retrieval, AI agent requests,
// and Redis caching for the AI-powered overview insights endpoint.
type AIInsightsService struct {
	analyticsService Service
	agentClient      agentRequester
	cache            redis.Client
}

// NewAIInsightsService constructs an AIInsightsService.
func NewAIInsightsService(svc Service, agentClient agentRequester, cache redis.Client) *AIInsightsService {
	return &AIInsightsService{
		analyticsService: svc,
		agentClient:      agentClient,
		cache:            cache,
	}
}

// GetAIInsights validates the insight type, checks Redis cache, fetches analytics data,
// sends it to the AI agent, caches and returns the response.
func (s *AIInsightsService) GetAIInsights(ctx context.Context, req *types.AIInsightsRequest) (map[string]interface{}, error) {
	agentEndpoint, ok := validInsightTypes[req.Type]
	if !ok {
		return nil, httputil.NewValidationError("invalid insight type: " + req.Type)
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

	baseReq := &req.OverviewRequest

	var (
		payload map[string]interface{}
		err     error
	)
	switch req.Type {
	case "reach_across_platforms", "engagement_across_platforms", "impressions_across_platforms":
		payload, err = s.buildTopPerformingGraphPayload(ctx, baseReq)
	case "platform_performance_comparison":
		payload, err = s.buildPlatformComparisonPayload(ctx, baseReq)
	case "accounts_statistics":
		payload, err = s.buildAccountStatisticsPayload(ctx, baseReq)
	case "top_posts":
		limit := req.Limit
		if limit <= 0 {
			limit = 5
		}
		payload, err = s.buildTopPostsPayload(ctx, baseReq, limit)
	}
	if err != nil {
		return nil, err
	}
	if payload == nil {
		return map[string]interface{}{"success": false, "message": "insufficient data"}, nil
	}

	payload["language"] = languageOrDefault(req.Language)
	payload["timezone"] = req.Timezone

	response, err := s.agentClient.Request(ctx, agentEndpoint, payload)
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

func (s *AIInsightsService) buildTopPerformingGraphPayload(ctx context.Context, req *types.OverviewRequest) (map[string]interface{}, error) {
	data, err := s.analyticsService.GetTopPerformingGraph(ctx, req)
	if err != nil {
		return nil, err
	}
	if data == nil {
		return nil, nil
	}
	return map[string]interface{}{"dataset": data}, nil
}

func (s *AIInsightsService) buildPlatformComparisonPayload(ctx context.Context, req *types.OverviewRequest) (map[string]interface{}, error) {
	var platformData []*types.AccountDataRow
	var accountData []*types.AccountDataDetailedRow
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.analyticsService.GetPlatformDataIndividual(egCtx, req)
		if err != nil {
			return err
		}
		platformData = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.analyticsService.GetPlatformDataDetailed(egCtx, req)
		if err != nil {
			return err
		}
		accountData = r
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}
	if len(platformData) == 0 && len(accountData) == 0 {
		return nil, nil
	}
	return map[string]interface{}{
		"dataset": map[string]interface{}{
			"platform_data": platformData,
			"account_data":  accountData,
		},
	}, nil
}

func (s *AIInsightsService) buildAccountStatisticsPayload(ctx context.Context, req *types.OverviewRequest) (map[string]interface{}, error) {
	data, err := s.analyticsService.GetPlatformDataDetailed(ctx, req)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return map[string]interface{}{"dataset": data}, nil
}

func (s *AIInsightsService) buildTopPostsPayload(ctx context.Context, req *types.OverviewRequest, limit int) (map[string]interface{}, error) {
	topReq := &types.TopPostsRequest{
		OverviewRequest: *req,
		Type:            "total_engagement",
		Limit:           limit,
	}
	data, err := s.analyticsService.GetTopPosts(ctx, topReq)
	if err != nil {
		return nil, err
	}
	if len(data) == 0 {
		return nil, nil
	}
	return map[string]interface{}{"dataset": data}, nil
}

func (s *AIInsightsService) buildCacheKey(req *types.AIInsightsRequest) string {
	accounts := fmt.Sprintf("%v|%v|%v|%v|%v|%v",
		req.FacebookAccounts, req.InstagramAccounts, req.LinkedInAccounts,
		req.TiktokAccounts, req.YouTubeAccounts, req.PinterestAccounts,
	)
	h := md5.Sum([]byte(accounts))
	return fmt.Sprintf("overview_AI:%s:%s:%s,%s:%s:%x",
		req.Type,
		req.WorkspaceID,
		req.StartDate,
		req.EndDate,
		languageOrDefault(req.Language),
		h,
	)
}

func languageOrDefault(language string) string {
	if language == "" {
		return "en"
	}
	return language
}
