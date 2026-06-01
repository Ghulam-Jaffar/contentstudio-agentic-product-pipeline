package instagram

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/instagram"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/instagram"
)

type mockService struct {
	getTopPostsFn func(context.Context, *types.TopPostsRequest) (*types.TopPostsResponse, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetSummary(_ context.Context, _ *types.InstagramRequest) (*types.SummaryResponse, error) {
	return &types.SummaryResponse{Status: true}, nil
}
func (m *mockService) GetAudienceGrowth(_ context.Context, _ *types.InstagramRequest) (*types.AudienceGrowthResponse, error) {
	return &types.AudienceGrowthResponse{Status: true}, nil
}
func (m *mockService) GetPublishingBehaviour(_ context.Context, _ *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error) {
	return &types.PublishingBehaviourResponse{Status: true}, nil
}
func (m *mockService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
	if m.getTopPostsFn != nil {
		return m.getTopPostsFn(ctx, req)
	}
	return &types.TopPostsResponse{Status: true}, nil
}
func (m *mockService) GetActiveUsers(_ context.Context, _ *types.InstagramRequest) (*types.ActiveUsersResponse, error) {
	return &types.ActiveUsersResponse{Status: true}, nil
}
func (m *mockService) GetImpressions(_ context.Context, _ *types.InstagramRequest) (*types.ImpressionsResponse, error) {
	return &types.ImpressionsResponse{Status: true}, nil
}
func (m *mockService) GetEngagement(_ context.Context, _ *types.InstagramRequest) (*types.EngagementResponse, error) {
	return &types.EngagementResponse{Status: true}, nil
}
func (m *mockService) GetHashtags(_ context.Context, _ *types.InstagramRequest) (*types.HashtagsResponse, error) {
	return &types.HashtagsResponse{Status: true}, nil
}
func (m *mockService) GetStoriesPerformance(_ context.Context, _ *types.InstagramRequest) (*types.StoriesPerformanceResponse, error) {
	return &types.StoriesPerformanceResponse{Status: true}, nil
}
func (m *mockService) GetReelsPerformance(_ context.Context, _ *types.InstagramRequest) (*types.ReelsPerformanceResponse, error) {
	return &types.ReelsPerformanceResponse{Status: true}, nil
}
func (m *mockService) GetDemographicsAge(_ context.Context, _ *types.InstagramRequest) (*types.DemographicsAgeResponse, error) {
	return &types.DemographicsAgeResponse{}, nil
}
func (m *mockService) GetCountryCity(_ context.Context, _ *types.InstagramRequest) (*types.CountryCityResponse, error) {
	return &types.CountryCityResponse{}, nil
}

func newTestHandler(svc service.Service) *InstagramHandler {
	return NewInstagramHandler(svc, zerolog.New(io.Discard))
}

const validQuery = "workspace_id=ws1&instagram_id=ig_1&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"

func get(url string) *http.Request {
	return httptest.NewRequest(http.MethodGet, url, nil)
}

func TestParseBaseRequest_FromStartEndDate(t *testing.T) {
	req := get("/?" + validQuery)
	parsed, err := parseBaseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.InstagramID != "ig_1" {
		t.Fatalf("expected ig_1, got %q", parsed.InstagramID)
	}
	if parsed.StartDate != "2025-01-01" || parsed.EndDate != "2025-01-31" {
		t.Fatalf("unexpected dates: %s / %s", parsed.StartDate, parsed.EndDate)
	}
}

func TestParseBaseRequest_DateRangeFallback(t *testing.T) {
	req := get("/?workspace_id=ws1&instagram_id=ig_1&date=2025-01-01+-+2025-01-31&timezone=UTC")
	parsed, err := parseBaseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.StartDate != "2025-01-01" || parsed.EndDate != "2025-01-31" {
		t.Fatalf("unexpected dates: %s / %s", parsed.StartDate, parsed.EndDate)
	}
}

func TestParseBaseRequest_MissingInstagramID(t *testing.T) {
	req := get("/?workspace_id=ws1&start_date=2025-01-01&end_date=2025-01-31")
	_, err := parseBaseRequest(req)
	if err == nil {
		t.Fatal("expected error for missing instagram_id")
	}
}

func TestParseBaseRequest_MissingWorkspaceID(t *testing.T) {
	req := get("/?instagram_id=ig_1&start_date=2025-01-01&end_date=2025-01-31")
	_, err := parseBaseRequest(req)
	if err == nil {
		t.Fatal("expected error for missing workspace_id")
	}
}

func TestHandleSummary(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleSummary(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp types.SummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status=true")
	}
}

func TestHandleSummary_MissingParams(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleSummary(w, get("/?workspace_id=ws1"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAudienceGrowth(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleAudienceGrowth(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlePublishingBehaviour(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandlePublishingBehaviour(w, get("/?"+validQuery+"&media_type=IMAGE,VIDEO"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleTopPosts_DefaultLimit(t *testing.T) {
	var gotLimit int
	h := newTestHandler(&mockService{
		getTopPostsFn: func(_ context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
			gotLimit = req.Limit
			return &types.TopPostsResponse{Status: true}, nil
		},
	})
	w := httptest.NewRecorder()
	h.HandleTopPosts(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotLimit != 5 {
		t.Fatalf("expected default limit 5, got %d", gotLimit)
	}
}

func TestHandleGetTopPosts_DefaultLimit(t *testing.T) {
	var gotLimit int
	h := newTestHandler(&mockService{
		getTopPostsFn: func(_ context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
			gotLimit = req.Limit
			return &types.TopPostsResponse{Status: true}, nil
		},
	})
	w := httptest.NewRecorder()
	h.HandleGetTopPosts(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotLimit != 15 {
		t.Fatalf("expected default limit 15, got %d", gotLimit)
	}
}

func TestHandleTopPosts_CustomLimit(t *testing.T) {
	var gotLimit int
	h := newTestHandler(&mockService{
		getTopPostsFn: func(_ context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
			gotLimit = req.Limit
			return &types.TopPostsResponse{Status: true}, nil
		},
	})
	w := httptest.NewRecorder()
	h.HandleTopPosts(w, get("/?"+validQuery+"&limit=10"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotLimit != 10 {
		t.Fatalf("expected limit 10, got %d", gotLimit)
	}
}

func TestHandleTopPosts_InvalidLimit(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleTopPosts(w, get("/?"+validQuery+"&limit=abc"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleActiveUsers(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleActiveUsers(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleImpressions(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleImpressions(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleEngagement(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleEngagement(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleHashtags(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleHashtags(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleStoriesPerformance(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleStoriesPerformance(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleReelsPerformance(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleReelsPerformance(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDemographicsAge(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleDemographicsAge(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleCountryCity(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleCountryCity(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleAIInsights_NotConfigured(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, get("/?"+validQuery+"&type=engagement"))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHandleAIInsights_MissingFields(t *testing.T) {
	h := newTestHandler(&mockService{})
	h.SetAIInsightsService(service.NewAIInsightsService(nil, nil, nil))
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, get("/?workspace_id=ws1&instagram_id=ig_1"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
