package meta_ads

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/meta_ads"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/meta_ads"
)

type mockService struct {
	getSummaryFn                   func(context.Context, *types.MetaAdsRequest) (*types.SummaryResponse, error)
	getResultsByObjectiveFn        func(context.Context, *types.MetaAdsRequest) (*types.ResultsByObjectiveResponse, error)
	getImpressionsVsSpendFn        func(context.Context, *types.MetaAdsRequest) (*types.ImpressionsVsSpendResponse, error)
	getClicksVsCTRFn               func(context.Context, *types.MetaAdsRequest) (*types.ClicksVsCTRResponse, error)
	getTopCampaignsFn              func(context.Context, *types.MetaAdsRequest) (*types.TopCampaignsResponse, error)
	getPerformanceTrendFn          func(context.Context, *types.MetaAdsRequest) (*types.PerformanceTrendResponse, error)
	getPerformanceByLevelFn        func(context.Context, *types.MetaAdsRequest) (*types.PerformanceByLevelResponse, error)
	getPerformanceByPlatformFn     func(context.Context, *types.MetaAdsRequest) (*types.PerformanceByPlatformResponse, error)
	getCampaignsListFn             func(context.Context, *types.MetaAdsRequest) (*types.TableResponse, error)
	getAdSetsListFn                func(context.Context, *types.MetaAdsRequest) (*types.TableResponse, error)
	getAdsListFn                   func(context.Context, *types.MetaAdsRequest) (*types.TableResponse, error)
	getDemographicsAgeGenderFn     func(context.Context, *types.MetaAdsRequest) (*types.DemographicsAgeGenderResponse, error)
	getDemographicsRegionCountryFn func(context.Context, *types.MetaAdsRequest) (*types.DemographicsRegionCountryResponse, error)
	getAIInsightsSummaryFn         func(context.Context, *types.MetaAdsRequest) (map[string]interface{}, error)
	getAIInsightsDetailedFn        func(context.Context, *types.MetaAdsRequest) (map[string]interface{}, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetSummary(ctx context.Context, req *types.MetaAdsRequest) (*types.SummaryResponse, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, req)
	}
	return &types.SummaryResponse{Status: true}, nil
}
func (m *mockService) GetResultsByObjective(ctx context.Context, req *types.MetaAdsRequest) (*types.ResultsByObjectiveResponse, error) {
	if m.getResultsByObjectiveFn != nil {
		return m.getResultsByObjectiveFn(ctx, req)
	}
	return &types.ResultsByObjectiveResponse{Status: true}, nil
}
func (m *mockService) GetImpressionsVsSpend(ctx context.Context, req *types.MetaAdsRequest) (*types.ImpressionsVsSpendResponse, error) {
	if m.getImpressionsVsSpendFn != nil {
		return m.getImpressionsVsSpendFn(ctx, req)
	}
	return &types.ImpressionsVsSpendResponse{Status: true}, nil
}
func (m *mockService) GetClicksVsCTR(ctx context.Context, req *types.MetaAdsRequest) (*types.ClicksVsCTRResponse, error) {
	if m.getClicksVsCTRFn != nil {
		return m.getClicksVsCTRFn(ctx, req)
	}
	return &types.ClicksVsCTRResponse{Status: true}, nil
}
func (m *mockService) GetTopCampaigns(ctx context.Context, req *types.MetaAdsRequest) (*types.TopCampaignsResponse, error) {
	if m.getTopCampaignsFn != nil {
		return m.getTopCampaignsFn(ctx, req)
	}
	return &types.TopCampaignsResponse{Status: true}, nil
}
func (m *mockService) GetPerformanceTrend(ctx context.Context, req *types.MetaAdsRequest) (*types.PerformanceTrendResponse, error) {
	if m.getPerformanceTrendFn != nil {
		return m.getPerformanceTrendFn(ctx, req)
	}
	return &types.PerformanceTrendResponse{Status: true}, nil
}
func (m *mockService) GetPerformanceByLevel(ctx context.Context, req *types.MetaAdsRequest) (*types.PerformanceByLevelResponse, error) {
	if m.getPerformanceByLevelFn != nil {
		return m.getPerformanceByLevelFn(ctx, req)
	}
	return &types.PerformanceByLevelResponse{Status: true}, nil
}
func (m *mockService) GetPerformanceByPlatform(ctx context.Context, req *types.MetaAdsRequest) (*types.PerformanceByPlatformResponse, error) {
	if m.getPerformanceByPlatformFn != nil {
		return m.getPerformanceByPlatformFn(ctx, req)
	}
	return &types.PerformanceByPlatformResponse{Status: true}, nil
}
func (m *mockService) GetCampaignsList(ctx context.Context, req *types.MetaAdsRequest) (*types.TableResponse, error) {
	if m.getCampaignsListFn != nil {
		return m.getCampaignsListFn(ctx, req)
	}
	return &types.TableResponse{Status: true}, nil
}
func (m *mockService) GetAdSetsList(ctx context.Context, req *types.MetaAdsRequest) (*types.TableResponse, error) {
	if m.getAdSetsListFn != nil {
		return m.getAdSetsListFn(ctx, req)
	}
	return &types.TableResponse{Status: true}, nil
}
func (m *mockService) GetAdsList(ctx context.Context, req *types.MetaAdsRequest) (*types.TableResponse, error) {
	if m.getAdsListFn != nil {
		return m.getAdsListFn(ctx, req)
	}
	return &types.TableResponse{Status: true}, nil
}
func (m *mockService) GetDemographicsAgeGender(ctx context.Context, req *types.MetaAdsRequest) (*types.DemographicsAgeGenderResponse, error) {
	if m.getDemographicsAgeGenderFn != nil {
		return m.getDemographicsAgeGenderFn(ctx, req)
	}
	return &types.DemographicsAgeGenderResponse{Status: true}, nil
}
func (m *mockService) GetDemographicsRegionCountry(ctx context.Context, req *types.MetaAdsRequest) (*types.DemographicsRegionCountryResponse, error) {
	if m.getDemographicsRegionCountryFn != nil {
		return m.getDemographicsRegionCountryFn(ctx, req)
	}
	return &types.DemographicsRegionCountryResponse{Status: true}, nil
}
func (m *mockService) GetAIInsightsSummary(ctx context.Context, req *types.MetaAdsRequest) (map[string]interface{}, error) {
	if m.getAIInsightsSummaryFn != nil {
		return m.getAIInsightsSummaryFn(ctx, req)
	}
	return map[string]interface{}{"success": true}, nil
}
func (m *mockService) GetAIInsightsDetailed(ctx context.Context, req *types.MetaAdsRequest) (map[string]interface{}, error) {
	if m.getAIInsightsDetailedFn != nil {
		return m.getAIInsightsDetailedFn(ctx, req)
	}
	return map[string]interface{}{"success": true}, nil
}

func newTestHandler() *Handler {
	return NewHandler(&mockService{}, zerolog.New(io.Discard))
}

func TestParseRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?workspace_id=ws1&account_id=act_1&start_date=2025-01-01&end_date=2025-01-31&page=2&per_page=25", nil)
	req.Header.Set("X-LOCALE", "fr")

	parsed, err := parseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.Language != "fr" {
		t.Fatalf("expected language fr, got %q", parsed.Language)
	}
	if parsed.Page != 2 || parsed.PerPage != 25 {
		t.Fatalf("unexpected pagination: %+v", parsed)
	}
}

func TestParseRequest_InvalidPage(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/?workspace_id=ws1&account_id=act_1&start_date=2025-01-01&end_date=2025-01-31&page=x", nil)
	if _, err := parseRequest(req); err == nil {
		t.Fatal("expected error for invalid page")
	}
}

func TestHandleSummary(t *testing.T) {
	h := NewHandler(&mockService{
		getSummaryFn: func(_ context.Context, req *types.MetaAdsRequest) (*types.SummaryResponse, error) {
			if req.WorkspaceID != "ws1" || req.AccountID != "act_1" {
				t.Fatalf("unexpected request: %+v", req)
			}
			return &types.SummaryResponse{Status: true}, nil
		},
	}, zerolog.New(io.Discard))

	req := httptest.NewRequest(http.MethodGet, "/?workspace_id=ws1&account_id=act_1&start_date=2025-01-01&end_date=2025-01-31", nil)
	w := httptest.NewRecorder()
	h.HandleSummary(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp types.SummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status true")
	}
}

func TestHandleAIInsightsDetailed(t *testing.T) {
	h := NewHandler(&mockService{
		getAIInsightsDetailedFn: func(_ context.Context, req *types.MetaAdsRequest) (map[string]interface{}, error) {
			if req.Language != "en" {
				t.Fatalf("expected language en, got %q", req.Language)
			}
			return map[string]interface{}{"success": true, "insights": map[string]interface{}{"foo": "bar"}}, nil
		},
	}, zerolog.New(io.Discard))

	req := httptest.NewRequest(http.MethodGet, "/?workspace_id=ws1&account_id=act_1&start_date=2025-01-01&end_date=2025-01-31", nil)
	req.Header.Set("X-LOCALE", "en")
	w := httptest.NewRecorder()
	h.HandleAIInsightsDetailed(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
