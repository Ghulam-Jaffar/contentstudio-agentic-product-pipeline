package meta_ads

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/meta_ads"
	"golang.org/x/sync/errgroup"
)

// aiCacheTTL is the Redis TTL for cached AI insight responses.
const aiCacheTTL = 24 * time.Hour

// validInsightTypes maps the API insight type key to its agent endpoint path segment.
var validInsightTypes = map[string]string{
	"aiInsightsSummary":  "analytics/meta-ads/overview-summary",
	"aiInsightsDetailed": "analytics/meta-ads/insights-detailed",
}

// AIInsightsService orchestrates AI-powered analytics insights for Meta Ads.
// It fetches analytics data, packages it as a dataset payload, sends to an AI agent
// endpoint, and caches the result in Redis for 24 hours.
type AIInsightsService struct {
	analyticsService currencyAwareAnalyticsService
	agentClient      agentRequester
	cache            redis.Client
}

// currencyAwareAnalyticsService extends the base Meta Ads service with account currency lookup.
type currencyAwareAnalyticsService interface {
	Service
	GetAccountCurrency(ctx context.Context, req *types.MetaAdsRequest) (string, error)
}

// agentRequester is the interface for making requests to the AI agent HTTP service.
type agentRequester interface {
	Request(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error)
}

// NewAIInsightsService creates a new AIInsightsService instance.
func NewAIInsightsService(analyticsService currencyAwareAnalyticsService, agentClient agentRequester, cache redis.Client) *AIInsightsService {
	return &AIInsightsService{
		analyticsService: analyticsService,
		agentClient:      agentClient,
		cache:            cache,
	}
}

// GetAIInsights fetches or generates AI insights for the specified insight type.
func (s *AIInsightsService) GetAIInsights(ctx context.Context, insightType string, req *types.MetaAdsRequest) (map[string]interface{}, error) {
	// Validate insight type
	agentEndpoint, valid := validInsightTypes[insightType]
	if !valid {
		return nil, fmt.Errorf("invalid insight type: %s", insightType)
	}

	// Create cache key
	cacheKey := fmt.Sprintf("meta_ads_ai_insights:%s:%s:%s:%s:%s:%s:%s:%s",
		insightType,
		req.WorkspaceID,
		req.AccountID,
		req.StartDate,
		req.EndDate,
		req.Timezone,
		req.Language,
		req.Currency,
	)

	// Try to get from cache
	if s.cache != nil {
		cached, err := s.cache.Get(ctx, cacheKey)
		if err == nil && cached != "" {
			var cachedResponse map[string]interface{}
			if err := json.Unmarshal([]byte(cached), &cachedResponse); err == nil {
				return cachedResponse, nil
			}
		}
	}

	// Fetch analytics data in parallel
	if s.analyticsService != nil {
		if currency, err := s.analyticsService.GetAccountCurrency(ctx, req); err == nil && currency != "" {
			req.Currency = currency
		} else if req.Currency == "" {
			req.Currency = "USD"
		}
	} else if req.Currency == "" {
		req.Currency = "USD"
	}

	dataset, err := s.fetchAnalyticsDataset(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch analytics dataset: %w", err)
	}

	analysisWindow := formatAIAnalysisWindow(req)
	dataset["metadata"] = map[string]interface{}{
		"start_date":      req.StartDate,
		"end_date":        req.EndDate,
		"timezone":        req.Timezone,
		"currency":        req.Currency,
		"analysis_window": analysisWindow,
	}

	// Prepare payload for AI agent
	payload := map[string]interface{}{
		"dataset":         dataset,
		"language":        req.Language,
		"timezone":        req.Timezone,
		"currency":        req.Currency,
		"analysis_window": analysisWindow,
	}

	// Call AI agent
	agentResponse, err := s.agentClient.Request(ctx, agentEndpoint, payload)
	if err != nil {
		return map[string]interface{}{
			"success": false,
			"error":   fmt.Sprintf("AI agent request failed: %v", err),
		}, nil
	}

	normalized := normalizeAgentResponse(agentResponse)

	// Cache the response
	if s.cache != nil {
		jsonBytes, _ := json.Marshal(normalized)
		_ = s.cache.Set(ctx, cacheKey, string(jsonBytes), aiCacheTTL)
	}

	return normalized, nil
}

func normalizeAgentResponse(agentResponse map[string]interface{}) map[string]interface{} {
	if agentResponse == nil {
		return map[string]interface{}{"success": false}
	}
	if insights, ok := agentResponse["insights"].(map[string]interface{}); ok && insights != nil {
		return insights
	}
	return agentResponse
}

func formatAIAnalysisWindow(req *types.MetaAdsRequest) string {
	if req == nil {
		return "Selected date range"
	}
	start, startErr := time.Parse("2006-01-02", req.StartDate)
	end, endErr := time.Parse("2006-01-02", req.EndDate)
	if startErr == nil && endErr == nil {
		return fmt.Sprintf("%s - %s", start.Format("Jan 2, 2006"), end.Format("Jan 2, 2006"))
	}
	if req.StartDate != "" && req.EndDate != "" {
		return fmt.Sprintf("%s - %s", req.StartDate, req.EndDate)
	}
	return "Selected date range"
}

// fetchAnalyticsDataset collects the required datasets for AI insights generation.
func (s *AIInsightsService) fetchAnalyticsDataset(ctx context.Context, req *types.MetaAdsRequest) (map[string]interface{}, error) {
	var (
		dataset = make(map[string]interface{})
		eg      errgroup.Group
		mu      sync.Mutex
	)

	setDataset := func(key string, value interface{}) {
		mu.Lock()
		defer mu.Unlock()
		dataset[key] = value
	}

	// Fetch summary data
	eg.Go(func() error {
		summaryResp, err := s.analyticsService.GetSummary(ctx, req)
		if err != nil {
			return err
		}
		setDataset("summary", summaryResp)
		return nil
	})

	// Fetch results by objective
	eg.Go(func() error {
		resultsByObjectiveResp, err := s.analyticsService.GetResultsByObjective(ctx, req)
		if err != nil {
			return err
		}
		setDataset("results_by_objective", resultsByObjectiveResp)
		return nil
	})

	// Fetch impressions vs spend
	eg.Go(func() error {
		impressionsVsSpendResp, err := s.analyticsService.GetImpressionsVsSpend(ctx, req)
		if err != nil {
			return err
		}
		setDataset("impressions_vs_spend", impressionsVsSpendResp)
		return nil
	})

	// Fetch clicks vs CTR
	eg.Go(func() error {
		clicksVsCTRResp, err := s.analyticsService.GetClicksVsCTR(ctx, req)
		if err != nil {
			return err
		}
		setDataset("clicks_vs_ctr", clicksVsCTRResp)
		return nil
	})

	// Fetch top campaigns
	eg.Go(func() error {
		topCampaignsResp, err := s.analyticsService.GetTopCampaigns(ctx, req)
		if err != nil {
			return err
		}
		setDataset("top_campaigns", topCampaignsResp)
		return nil
	})

	// Fetch performance trend
	eg.Go(func() error {
		performanceTrendResp, err := s.analyticsService.GetPerformanceTrend(ctx, req)
		if err != nil {
			return err
		}
		setDataset("performance_trend", performanceTrendResp)
		return nil
	})

	// Fetch performance by level
	eg.Go(func() error {
		performanceByLevelResp, err := s.analyticsService.GetPerformanceByLevel(ctx, req)
		if err != nil {
			return err
		}
		setDataset("performance_by_level", performanceByLevelResp)
		return nil
	})

	// Fetch performance by platform
	eg.Go(func() error {
		performanceByPlatformResp, err := s.analyticsService.GetPerformanceByPlatform(ctx, req)
		if err != nil {
			return err
		}
		setDataset("performance_by_platform", performanceByPlatformResp)
		return nil
	})

	// Fetch campaigns list (limited)
	eg.Go(func() error {
		campaignsListResp, err := s.analyticsService.GetCampaignsList(ctx, req)
		if err != nil {
			return err
		}
		// Limit to top 5 for AI insights
		if campaignsListResp != nil && campaignsListResp.Data != nil {
			if len(campaignsListResp.Data) > 5 {
				campaignsListResp.Data = campaignsListResp.Data[:5]
			}
		}
		setDataset("campaigns_list", campaignsListResp)
		return nil
	})

	// Fetch ad sets list (limited)
	eg.Go(func() error {
		adSetsListResp, err := s.analyticsService.GetAdSetsList(ctx, req)
		if err != nil {
			return err
		}
		// Limit to top 5 for AI insights
		if adSetsListResp != nil && adSetsListResp.Data != nil {
			if len(adSetsListResp.Data) > 5 {
				adSetsListResp.Data = adSetsListResp.Data[:5]
			}
		}
		setDataset("ad_sets_list", adSetsListResp)
		return nil
	})

	// Fetch ads list (limited)
	eg.Go(func() error {
		adsListResp, err := s.analyticsService.GetAdsList(ctx, req)
		if err != nil {
			return err
		}
		// Limit to top 5 for AI insights
		if adsListResp != nil && adsListResp.Data != nil {
			if len(adsListResp.Data) > 5 {
				adsListResp.Data = adsListResp.Data[:5]
			}
		}
		setDataset("ads_list", adsListResp)
		return nil
	})

	// Fetch demographics age/gender
	eg.Go(func() error {
		demographicsAgeGenderResp, err := s.analyticsService.GetDemographicsAgeGender(ctx, req)
		if err != nil {
			return err
		}
		setDataset("demographics_age_gender", demographicsAgeGenderResp)
		return nil
	})

	// Fetch demographics region/country
	eg.Go(func() error {
		demographicsRegionCountryResp, err := s.analyticsService.GetDemographicsRegionCountry(ctx, req)
		if err != nil {
			return err
		}
		setDataset("demographics_region_country", demographicsRegionCountryResp)
		return nil
	})

	if err := eg.Wait(); err != nil {
		return nil, err
	}

	attachComputedMetrics(dataset)

	return dataset, nil
}

func attachComputedMetrics(dataset map[string]interface{}) {
	if dataset == nil {
		return
	}

	summary, _ := dataset["summary"].(*types.SummaryResponse)
	resultsByObjective, _ := dataset["results_by_objective"].(*types.ResultsByObjectiveResponse)
	if summary == nil || resultsByObjective == nil {
		return
	}

	var totalResults int64
	for _, row := range resultsByObjective.Data {
		totalResults += row.Results
	}

	currentSpend := summary.Spend.Current
	computed := map[string]interface{}{
		"spend":   currentSpend,
		"results": totalResults,
		"roas":    0.0,
	}
	if currentSpend > 0 {
		// ROAS: guard against integer rounding to zero by keeping higher precision.
		// Note: ideally revenue should be used (revenue / spend). If revenue becomes available,
		// replace `totalResults` with totalRevenue. For now compute results/spend with 4 decimal precision.
		roasRaw := float64(totalResults) / currentSpend
		computed["roas"] = math.Round(roasRaw*10000) / 10000
	}

	dataset["computed_metrics"] = computed
}
