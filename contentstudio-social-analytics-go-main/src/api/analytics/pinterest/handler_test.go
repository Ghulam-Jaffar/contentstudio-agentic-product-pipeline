package pinterest

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/pinterest"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/pinterest"
)

type mockService struct {
	getTopPinsFn func(context.Context, *types.TopPinsRequest) (*types.TopPinsResponse, error)
	getPinPostingFn func(context.Context, *types.FilteredPinRequest) (*types.PinPostingResponse, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetSummary(_ context.Context, _ *types.PinterestRequest) (*types.SummaryResponse, error) {
	return &types.SummaryResponse{Status: true}, nil
}
func (m *mockService) GetFollowerTrend(_ context.Context, _ *types.PinterestRequest) (*types.FollowerTrendResponse, error) {
	return &types.FollowerTrendResponse{Status: true}, nil
}
func (m *mockService) GetDynamicFollowerTrend(_ context.Context, _ *types.PinterestRequest) (*types.FollowerTrendResponse, error) {
	return &types.FollowerTrendResponse{Status: true}, nil
}
func (m *mockService) GetImpressionsTrend(_ context.Context, _ *types.PinterestRequest) (*types.ImpressionsTrendResponse, error) {
	return &types.ImpressionsTrendResponse{Status: true}, nil
}
func (m *mockService) GetDynamicImpressionsTrend(_ context.Context, _ *types.PinterestRequest) (*types.ImpressionsTrendResponse, error) {
	return &types.ImpressionsTrendResponse{Status: true}, nil
}
func (m *mockService) GetEngagementTrend(_ context.Context, _ *types.PinterestRequest) (*types.EngagementTrendResponse, error) {
	return &types.EngagementTrendResponse{Status: true}, nil
}
func (m *mockService) GetDynamicEngagementTrend(_ context.Context, _ *types.PinterestRequest) (*types.EngagementTrendResponse, error) {
	return &types.EngagementTrendResponse{Status: true}, nil
}
func (m *mockService) GetPinPosting(ctx context.Context, req *types.FilteredPinRequest) (*types.PinPostingResponse, error) {
	if m.getPinPostingFn != nil {
		return m.getPinPostingFn(ctx, req)
	}
	return &types.PinPostingResponse{Status: true}, nil
}
func (m *mockService) GetDynamicPinPosting(_ context.Context, _ *types.FilteredPinRequest) (*types.PinPostingResponse, error) {
	return &types.PinPostingResponse{Status: true}, nil
}
func (m *mockService) GetPinRollup(_ context.Context, _ *types.PinterestRequest) (*types.PinRollupResponse, error) {
	return &types.PinRollupResponse{Status: true}, nil
}
func (m *mockService) GetTopPins(ctx context.Context, req *types.TopPinsRequest) (*types.TopPinsResponse, error) {
	if m.getTopPinsFn != nil {
		return m.getTopPinsFn(ctx, req)
	}
	return &types.TopPinsResponse{Status: true}, nil
}
func (m *mockService) GetPinPerformance(_ context.Context, _ *types.PinterestRequest) (*types.PinPerformanceResponse, error) {
	return &types.PinPerformanceResponse{Status: true}, nil
}

func newTestHandler(svc service.Service) *Handler {
	return NewHandler(svc, zerolog.New(io.Discard))
}

const validQuery = "pinterest_id=pin_1&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"

func get(url string) *http.Request {
	return httptest.NewRequest(http.MethodGet, url, nil)
}

func TestParseBaseRequest_FromStartEndDate(t *testing.T) {
	req := get("/?" + validQuery)
	parsed, err := parseBaseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.PinterestID != "pin_1" {
		t.Fatalf("expected pin_1, got %q", parsed.PinterestID)
	}
	if parsed.StartDate != "2025-01-01" || parsed.EndDate != "2025-01-31" {
		t.Fatalf("unexpected dates: %s / %s", parsed.StartDate, parsed.EndDate)
	}
}

func TestParseBaseRequest_DateRangeFallback(t *testing.T) {
	req := get("/?pinterest_id=pin_1&date=2025-01-01+-+2025-01-31&timezone=UTC")
	parsed, err := parseBaseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.StartDate != "2025-01-01" || parsed.EndDate != "2025-01-31" {
		t.Fatalf("unexpected dates: %s / %s", parsed.StartDate, parsed.EndDate)
	}
}

func TestParseBaseRequest_MissingPinterestID(t *testing.T) {
	req := get("/?start_date=2025-01-01&end_date=2025-01-31")
	_, err := parseBaseRequest(req)
	if err == nil {
		t.Fatal("expected error for missing pinterest_id")
	}
}

func TestHandleSummary(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleSummary(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleSummary_MissingParams(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleSummary(w, get("/?pinterest_id=pin_1"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleFollowerTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleFollowerTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDynamicFollowerTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleDynamicFollowerTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleImpressionsTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleImpressionsTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDynamicImpressionsTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleDynamicImpressionsTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleEngagementTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleEngagementTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDynamicEngagementTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleDynamicEngagementTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlePinPosting_NoFilter(t *testing.T) {
	var gotFilter string
	h := newTestHandler(&mockService{
		getPinPostingFn: func(_ context.Context, req *types.FilteredPinRequest) (*types.PinPostingResponse, error) {
			gotFilter = req.FilterBy
			return &types.PinPostingResponse{Status: true}, nil
		},
	})
	w := httptest.NewRecorder()
	h.HandlePinPosting(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotFilter != "" {
		t.Fatalf("expected empty filter_by, got %q", gotFilter)
	}
}

func TestHandlePinPosting_WithFilter(t *testing.T) {
	var gotFilter string
	h := newTestHandler(&mockService{
		getPinPostingFn: func(_ context.Context, req *types.FilteredPinRequest) (*types.PinPostingResponse, error) {
			gotFilter = req.FilterBy
			return &types.PinPostingResponse{Status: true}, nil
		},
	})
	w := httptest.NewRecorder()
	h.HandlePinPosting(w, get("/?"+validQuery+"&filter_by=video"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotFilter != "video" {
		t.Fatalf("expected filter_by=video, got %q", gotFilter)
	}
}

func TestHandleDynamicPinPosting(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleDynamicPinPosting(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlePinRollup(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandlePinRollup(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleTopPins_DefaultLimit(t *testing.T) {
	var gotLimit int
	h := newTestHandler(&mockService{
		getTopPinsFn: func(_ context.Context, req *types.TopPinsRequest) (*types.TopPinsResponse, error) {
			gotLimit = req.Limit
			return &types.TopPinsResponse{Status: true}, nil
		},
	})
	w := httptest.NewRecorder()
	h.HandleTopPins(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotLimit != 5 {
		t.Fatalf("expected default limit 5, got %d", gotLimit)
	}
}

func TestHandleTopPins_CustomLimit(t *testing.T) {
	var gotLimit int
	h := newTestHandler(&mockService{
		getTopPinsFn: func(_ context.Context, req *types.TopPinsRequest) (*types.TopPinsResponse, error) {
			gotLimit = req.Limit
			return &types.TopPinsResponse{Status: true}, nil
		},
	})
	w := httptest.NewRecorder()
	h.HandleTopPins(w, get("/?"+validQuery+"&limit=20"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotLimit != 20 {
		t.Fatalf("expected limit 20, got %d", gotLimit)
	}
}

func TestHandleTopPins_InvalidLimit(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleTopPins(w, get("/?"+validQuery+"&limit=abc"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandlePinPerformance(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandlePinPerformance(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleAIInsights_NotConfigured(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, get("/?"+validQuery+"&type=impressions"))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHandleAIInsights_MissingFields(t *testing.T) {
	h := newTestHandler(&mockService{})
	h.SetAIInsightsService(service.NewAIInsightsService(nil, nil, nil))
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, get("/?pinterest_id=pin_1"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
