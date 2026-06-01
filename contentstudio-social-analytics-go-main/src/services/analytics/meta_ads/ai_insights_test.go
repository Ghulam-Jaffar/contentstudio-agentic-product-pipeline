package meta_ads

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/meta_ads"
)

type mockAnalyticsService struct{}

func (m *mockAnalyticsService) GetSummary(context.Context, *types.MetaAdsRequest) (*types.SummaryResponse, error) {
	return &types.SummaryResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetResultsByObjective(context.Context, *types.MetaAdsRequest) (*types.ResultsByObjectiveResponse, error) {
	return &types.ResultsByObjectiveResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetImpressionsVsSpend(context.Context, *types.MetaAdsRequest) (*types.ImpressionsVsSpendResponse, error) {
	return &types.ImpressionsVsSpendResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetClicksVsCTR(context.Context, *types.MetaAdsRequest) (*types.ClicksVsCTRResponse, error) {
	return &types.ClicksVsCTRResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetTopCampaigns(context.Context, *types.MetaAdsRequest) (*types.TopCampaignsResponse, error) {
	return &types.TopCampaignsResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetPerformanceTrend(context.Context, *types.MetaAdsRequest) (*types.PerformanceTrendResponse, error) {
	return &types.PerformanceTrendResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetPerformanceByLevel(context.Context, *types.MetaAdsRequest) (*types.PerformanceByLevelResponse, error) {
	return &types.PerformanceByLevelResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetPerformanceByPlatform(context.Context, *types.MetaAdsRequest) (*types.PerformanceByPlatformResponse, error) {
	return &types.PerformanceByPlatformResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetCampaignsList(context.Context, *types.MetaAdsRequest) (*types.TableResponse, error) {
	return &types.TableResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetAdSetsList(context.Context, *types.MetaAdsRequest) (*types.TableResponse, error) {
	return &types.TableResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetAdsList(context.Context, *types.MetaAdsRequest) (*types.TableResponse, error) {
	return &types.TableResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetDemographicsAgeGender(context.Context, *types.MetaAdsRequest) (*types.DemographicsAgeGenderResponse, error) {
	return &types.DemographicsAgeGenderResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetDemographicsRegionCountry(context.Context, *types.MetaAdsRequest) (*types.DemographicsRegionCountryResponse, error) {
	return &types.DemographicsRegionCountryResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetAIInsightsSummary(context.Context, *types.MetaAdsRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"success": true}, nil
}

func (m *mockAnalyticsService) GetAIInsightsDetailed(context.Context, *types.MetaAdsRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"success": true}, nil
}

func (m *mockAnalyticsService) GetAccountCurrency(context.Context, *types.MetaAdsRequest) (string, error) {
	return "EUR", nil
}

type mockCache struct {
	value string
}

func (m *mockCache) Get(_ context.Context, _ string) (string, error) { return m.value, nil }
func (m *mockCache) Set(_ context.Context, _ string, value interface{}, _ time.Duration) error {
	m.value = value.(string)
	return nil
}
func (m *mockCache) Close() error { return nil }

type mockAgent struct {
	payload map[string]interface{}
}

func (m *mockAgent) Request(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	m.payload = payload
	return map[string]interface{}{"insights": map[string]interface{}{"endpoint": endpoint}}, nil
}

func TestAIHelpers(t *testing.T) {
	if got := formatAIAnalysisWindow(&types.MetaAdsRequest{StartDate: "2025-01-01", EndDate: "2025-01-31"}); got != "Jan 1, 2025 - Jan 31, 2025" {
		t.Fatalf("unexpected window: %s", got)
	}
	if got := normalizeAgentResponse(map[string]interface{}{"insights": map[string]interface{}{"a": 1}}); got["a"] != 1 {
		t.Fatalf("unexpected normalized response: %+v", got)
	}
}

func TestGetAIInsightsInvalidType(t *testing.T) {
	svc := NewAIInsightsService(nil, nil, nil)
	if _, err := svc.GetAIInsights(context.Background(), "invalid", &types.MetaAdsRequest{}); err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestGetAIInsightsCacheHit(t *testing.T) {
	cached, _ := json.Marshal(map[string]interface{}{"cached": true})
	svc := NewAIInsightsService(nil, &mockAgent{}, &mockCache{value: string(cached)})
	resp, err := svc.GetAIInsights(context.Background(), "aiInsightsSummary", &types.MetaAdsRequest{
		WorkspaceID: "ws1",
		AccountID:   "act_1",
		StartDate:   "2025-01-01",
		EndDate:     "2025-01-31",
		Timezone:    "UTC",
		Language:    "en",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["cached"] != true {
		t.Fatalf("expected cached response, got %+v", resp)
	}
}

func TestGetAIInsightsPassesCurrency(t *testing.T) {
	agent := &mockAgent{}
	svc := NewAIInsightsService(&mockAnalyticsService{}, agent, nil)

	_, err := svc.GetAIInsights(context.Background(), "aiInsightsDetailed", &types.MetaAdsRequest{
		WorkspaceID: "ws1",
		AccountID:   "act_1",
		StartDate:   "2025-01-01",
		EndDate:     "2025-01-31",
		Timezone:    "UTC",
		Language:    "en",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if agent.payload["currency"] != "EUR" {
		t.Fatalf("expected payload currency EUR, got %+v", agent.payload["currency"])
	}
}
