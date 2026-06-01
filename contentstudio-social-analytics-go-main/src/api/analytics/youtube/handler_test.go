package youtube

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/youtube"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/youtube"
)

type mockService struct {
	getSortedTopVideosFn func(context.Context, *types.TopVideosRequest) (*types.SortedTopVideosResponse, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetSummary(_ context.Context, _ *types.YoutubeRequest) (*types.SummaryResponse, error) {
	return &types.SummaryResponse{Status: true}, nil
}
func (m *mockService) GetSubscriberTrend(_ context.Context, _ *types.YoutubeRequest) (*types.SubscriberTrendResponse, error) {
	return &types.SubscriberTrendResponse{Status: true}, nil
}
func (m *mockService) GetDynamicSubscriberTrend(_ context.Context, _ *types.YoutubeRequest) (*types.SubscriberTrendResponse, error) {
	return &types.SubscriberTrendResponse{Status: true}, nil
}
func (m *mockService) GetEngagementTrend(_ context.Context, _ *types.YoutubeRequest) (*types.EngagementTrendResponse, error) {
	return &types.EngagementTrendResponse{Status: true}, nil
}
func (m *mockService) GetDynamicEngagementTrend(_ context.Context, _ *types.YoutubeRequest) (*types.EngagementTrendResponse, error) {
	return &types.EngagementTrendResponse{Status: true}, nil
}
func (m *mockService) GetViewsTrend(_ context.Context, _ *types.YoutubeRequest) (*types.ViewsTrendResponse, error) {
	return &types.ViewsTrendResponse{Status: true}, nil
}
func (m *mockService) GetDynamicViewsTrend(_ context.Context, _ *types.YoutubeRequest) (*types.ViewsTrendResponse, error) {
	return &types.ViewsTrendResponse{Status: true}, nil
}
func (m *mockService) GetWatchTimeTrend(_ context.Context, _ *types.YoutubeRequest) (*types.WatchTimeTrendResponse, error) {
	return &types.WatchTimeTrendResponse{Status: true}, nil
}
func (m *mockService) GetDynamicWatchTimeTrend(_ context.Context, _ *types.YoutubeRequest) (*types.WatchTimeTrendResponse, error) {
	return &types.WatchTimeTrendResponse{Status: true}, nil
}
func (m *mockService) GetFindVideo(_ context.Context, _ *types.YoutubeRequest) (*types.FindVideoResponse, error) {
	return &types.FindVideoResponse{Status: true}, nil
}
func (m *mockService) GetVideoSharing(_ context.Context, _ *types.YoutubeRequest) (*types.VideoSharingResponse, error) {
	return &types.VideoSharingResponse{Status: true}, nil
}
func (m *mockService) GetTopVideos(_ context.Context, _ *types.YoutubeRequest) (*types.TopVideosResponse, error) {
	return &types.TopVideosResponse{Status: true}, nil
}
func (m *mockService) GetLeastVideos(_ context.Context, _ *types.YoutubeRequest) (*types.LeastVideosResponse, error) {
	return &types.LeastVideosResponse{Status: true}, nil
}
func (m *mockService) GetSortedTopVideos(ctx context.Context, req *types.TopVideosRequest) (*types.SortedTopVideosResponse, error) {
	if m.getSortedTopVideosFn != nil {
		return m.getSortedTopVideosFn(ctx, req)
	}
	return &types.SortedTopVideosResponse{Status: true}, nil
}
func (m *mockService) GetPerformanceAndSchedule(_ context.Context, _ *types.YoutubeRequest) (*types.PerformanceScheduleResponse, error) {
	return &types.PerformanceScheduleResponse{Status: true}, nil
}

func newTestHandler(svc service.Service) *Handler {
	return NewHandler(svc, zerolog.New(io.Discard))
}

const validQuery = "youtube_id=yt_1&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"

func get(url string) *http.Request {
	return httptest.NewRequest(http.MethodGet, url, nil)
}

func TestParseBaseRequest_FromStartEndDate(t *testing.T) {
	req := get("/?" + validQuery)
	parsed, err := parseBaseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.YoutubeID != "yt_1" {
		t.Fatalf("expected yt_1, got %q", parsed.YoutubeID)
	}
	if parsed.StartDate != "2025-01-01" || parsed.EndDate != "2025-01-31" {
		t.Fatalf("unexpected dates: %s / %s", parsed.StartDate, parsed.EndDate)
	}
}

func TestParseBaseRequest_DateRangeFallback(t *testing.T) {
	req := get("/?youtube_id=yt_1&date=2025-01-01+-+2025-01-31&timezone=UTC")
	parsed, err := parseBaseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.StartDate != "2025-01-01" || parsed.EndDate != "2025-01-31" {
		t.Fatalf("unexpected dates: %s / %s", parsed.StartDate, parsed.EndDate)
	}
}

func TestParseBaseRequest_MissingYoutubeID(t *testing.T) {
	req := get("/?start_date=2025-01-01&end_date=2025-01-31")
	_, err := parseBaseRequest(req)
	if err == nil {
		t.Fatal("expected error for missing youtube_id")
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
	h.HandleSummary(w, get("/?youtube_id=yt_1"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleSubscriberTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleSubscriberTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDynamicSubscriberTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleDynamicSubscriberTrend(w, get("/?"+validQuery))
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

func TestHandleViewsTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleViewsTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDynamicViewsTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleDynamicViewsTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleWatchTimeTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleWatchTimeTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleDynamicWatchTimeTrend(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleDynamicWatchTimeTrend(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleFindVideo(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleFindVideo(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleVideoSharing(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleVideoSharing(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleTopPosts(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleTopPosts(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleLeastPosts(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleLeastPosts(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleGetTopPosts_DefaultLimit(t *testing.T) {
	var gotLimit int
	h := newTestHandler(&mockService{
		getSortedTopVideosFn: func(_ context.Context, req *types.TopVideosRequest) (*types.SortedTopVideosResponse, error) {
			gotLimit = req.Limit
			return &types.SortedTopVideosResponse{Status: true}, nil
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

func TestHandleGetTopPosts_CustomLimit(t *testing.T) {
	var gotLimit int
	h := newTestHandler(&mockService{
		getSortedTopVideosFn: func(_ context.Context, req *types.TopVideosRequest) (*types.SortedTopVideosResponse, error) {
			gotLimit = req.Limit
			return &types.SortedTopVideosResponse{Status: true}, nil
		},
	})
	w := httptest.NewRecorder()
	h.HandleGetTopPosts(w, get("/?"+validQuery+"&limit=30"))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotLimit != 30 {
		t.Fatalf("expected limit 30, got %d", gotLimit)
	}
}

func TestHandleGetTopPosts_InvalidLimit(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleGetTopPosts(w, get("/?"+validQuery+"&limit=abc"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandlePerformanceAndSchedule(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandlePerformanceAndSchedule(w, get("/?"+validQuery))
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleAIInsights_NotConfigured(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, get("/?"+validQuery+"&type=summary"))
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHandleAIInsights_MissingFields(t *testing.T) {
	h := newTestHandler(&mockService{})
	h.SetAIInsightsService(service.NewAIInsightsService(nil, nil, nil))
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, get("/?youtube_id=yt_1"))
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
