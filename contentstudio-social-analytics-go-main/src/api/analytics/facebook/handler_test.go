package facebook

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/facebook"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/facebook"
)

type mockService struct {
	getSummaryFn func(context.Context, *types.FacebookRequest) (*types.SummaryResponse, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetSummary(ctx context.Context, req *types.FacebookRequest) (*types.SummaryResponse, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, req)
	}
	return &types.SummaryResponse{Status: true, Overview: map[string]*types.SummaryMetrics{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetAudienceGrowth(context.Context, *types.FacebookRequest) (*types.AudienceGrowthResponse, error) {
	return &types.AudienceGrowthResponse{Status: true, AudienceGrowth: &types.AudienceGrowthData{}, AudienceGrowthRollup: map[string]*types.AudienceGrowthRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetPublishingBehaviour(context.Context, *types.FacebookRequest) (*types.PublishingBehaviourResponse, error) {
	return &types.PublishingBehaviourResponse{Status: true, PublishingBehaviour: &types.PublishingBehaviourData{}, PublishingBehaviourRollup: map[string]*types.PublishingRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetTopPosts(context.Context, *types.FacebookRequest) (*types.TopPostsResponse, error) {
	return &types.TopPostsResponse{Status: true, TopPosts: []types.TopPost{}}, nil
}
func (m *mockService) GetActiveUsers(context.Context, *types.FacebookRequest) (*types.ActiveUsersResponse, error) {
	return &types.ActiveUsersResponse{Status: true, ActiveUsers: &types.ActiveUsersData{}}, nil
}
func (m *mockService) GetImpressions(context.Context, *types.FacebookRequest) (*types.ImpressionsResponse, error) {
	return &types.ImpressionsResponse{Status: true, Impressions: &types.ImpressionsData{}, ImpressionsRollup: map[string]*types.ImpressionsRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetEngagement(context.Context, *types.FacebookRequest) (*types.EngagementResponse, error) {
	return &types.EngagementResponse{Status: true, Engagement: &types.EngagementContainer{Engagement: &types.EngagementData{}, EngagementRollup: map[string]*types.EngagementRollup{"current": {}, "previous": {}}}}, nil
}
func (m *mockService) GetReelsAnalytics(context.Context, *types.FacebookRequest) (*types.ReelsAnalyticsResponse, error) {
	return &types.ReelsAnalyticsResponse{Status: true, Reels: &types.ReelsData{}, ReelsRollup: map[string]*types.ReelsRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetVideoInsights(context.Context, *types.FacebookRequest) (*types.VideoInsightsResponse, error) {
	return &types.VideoInsightsResponse{Status: true, VideoInsights: &types.VideoInsightsData{}, VideoRollup: map[string]*types.VideoRollup{"current": {}, "previous": {}}}, nil
}
func (m *mockService) GetDemographics(context.Context, *types.FacebookRequest) (*types.DemographicsResponse, error) {
	return &types.DemographicsResponse{Status: true}, nil
}
func (m *mockService) GetOverviewDemographics(context.Context, *types.FacebookRequest) (*types.DemographicsResponse, error) {
	return &types.DemographicsResponse{Status: true}, nil
}
func (m *mockService) GetAudienceLocation(context.Context, *types.FacebookRequest) (*types.DemographicsResponse, error) {
	return &types.DemographicsResponse{Status: true}, nil
}

func newTestHandler(svc service.Service) *Handler {
	return NewHandler(svc, zerolog.New(io.Discard))
}

func TestParseRequest_FromQuery(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/facebook/summary?workspace_id=ws1&facebook_id=fb_1,fb_2&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC", nil)
	parsed, err := parseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(parsed.FacebookIDs) != 2 {
		t.Fatalf("expected 2 facebook ids, got %d", len(parsed.FacebookIDs))
	}
}

func TestHandleSummary(t *testing.T) {
	h := newTestHandler(&mockService{
		getSummaryFn: func(_ context.Context, req *types.FacebookRequest) (*types.SummaryResponse, error) {
			if req.WorkspaceID != "ws1" {
				t.Fatalf("expected workspace_id ws1, got %q", req.WorkspaceID)
			}
			return &types.SummaryResponse{Status: true, Overview: map[string]*types.SummaryMetrics{"current": {}, "previous": {}}}, nil
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/facebook/summary?workspace_id=ws1&facebook_id=fb_1&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC", nil)
	w := httptest.NewRecorder()
	h.HandleSummary(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var resp types.SummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true in response")
	}
}

func TestHandleAIInsights_NotConfigured(t *testing.T) {
	h := newTestHandler(&mockService{})

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/overview/facebook/ai_insights?workspace_id=ws1&facebook_id=fb_1&date=2025-01-01+-+2025-01-31&type=page_impressions",
		nil,
	)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHandleAIInsights_MissingFields(t *testing.T) {
	h := newTestHandler(&mockService{})
	h.SetAIInsightsService(service.NewAIInsightsService(nil, nil, nil))

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/overview/facebook/ai_insights?workspace_id=ws1&facebook_id=fb_1",
		nil,
	)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAIInsights_InvalidLimit(t *testing.T) {
	h := newTestHandler(&mockService{})
	h.SetAIInsightsService(service.NewAIInsightsService(nil, nil, nil))

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/overview/facebook/ai_insights?workspace_id=ws1&facebook_id=fb_1&date=2025-01-01+-+2025-01-31&type=page_impressions&limit=abc",
		nil,
	)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
