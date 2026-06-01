package ig_competitor

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/ig_competitor"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/ig_competitor"
)

type mockService struct {
	result interface{}
	err    error
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetDataTableMetrics(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	return m.result, m.err
}
func (m *mockService) GetPostingActivityGraphByTypes(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	return m.result, m.err
}
func (m *mockService) GetPostingActivityBySpecificType(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	return m.result, m.err
}
func (m *mockService) GetPostingActivityTableByType(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	return m.result, m.err
}
func (m *mockService) GetFollowersGrowthComparison(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	return m.result, m.err
}
func (m *mockService) GetTopAndLeastPerformingPosts(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	return m.result, m.err
}
func (m *mockService) GetTopHashtags(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	return m.result, m.err
}
func (m *mockService) GetIndividualHashtagData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	return m.result, m.err
}
func (m *mockService) GetBiographyData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	return m.result, m.err
}

func newTestHandler(result interface{}, err error) *Handler {
	logger := zerolog.New(io.Discard)
	return NewHandler(&mockService{result: result, err: err}, logger)
}

func TestParseRequest_Valid(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?_id=abc123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC", nil)
	req, err := parseRequest(r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.ReportID != "abc123" {
		t.Errorf("expected report_id=abc123, got %s", req.ReportID)
	}
	if req.StartDate != "2025-01-01" {
		t.Errorf("expected start_date=2025-01-01, got %s", req.StartDate)
	}
}

func TestParseRequest_MissingDates(t *testing.T) {
	// Both dates absent is valid (e.g. biographyData)
	r := httptest.NewRequest("GET", "/test?_id=abc123", nil)
	req, err := parseRequest(r)
	if err != nil {
		t.Fatalf("expected no error for missing dates, got %v", err)
	}
	if req.StartDate != "" || req.EndDate != "" {
		t.Fatalf("expected empty dates, got start=%q end=%q", req.StartDate, req.EndDate)
	}
}

func TestParseRequest_PartialDates(t *testing.T) {
	// Only one date provided should error
	r := httptest.NewRequest("GET", "/test?_id=abc123&start_date=2025-01-01", nil)
	_, err := parseRequest(r)
	if err == nil {
		t.Fatal("expected validation error for partial dates")
	}
}

func TestParseRequest_InvalidLimit(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?_id=abc123&start_date=2025-01-01&end_date=2025-01-31&limit=notanumber", nil)
	_, err := parseRequest(r)
	if err == nil {
		t.Fatal("expected error for invalid limit")
	}
}

func TestParseRequest_WithMediaProductType(t *testing.T) {
	r := httptest.NewRequest("GET", "/test?_id=abc123&start_date=2025-01-01&end_date=2025-01-31&media_type=VIDEO&media_product_type=REELS", nil)
	req, err := parseRequest(r)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if req.MediaType != "VIDEO" {
		t.Errorf("expected media_type=VIDEO, got %s", req.MediaType)
	}
	if req.MediaProductType != "REELS" {
		t.Errorf("expected media_product_type=REELS, got %s", req.MediaProductType)
	}
}

func TestHandleDataTableMetrics(t *testing.T) {
	expected := map[string]interface{}{"data": []interface{}{}}
	h := newTestHandler(expected, nil)
	r := httptest.NewRequest("GET", "/test?_id=abc123&start_date=2025-01-01&end_date=2025-01-31", nil)
	w := httptest.NewRecorder()
	h.HandleDataTableMetrics(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlePostingActivityGraphByTypes(t *testing.T) {
	expected := map[string]interface{}{"data": []interface{}{}}
	h := newTestHandler(expected, nil)
	r := httptest.NewRequest("GET", "/test?_id=abc123&start_date=2025-01-01&end_date=2025-01-31", nil)
	w := httptest.NewRecorder()
	h.HandlePostingActivityGraphByTypes(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandlePostingActivityTableByType(t *testing.T) {
	expected := map[string]interface{}{"data": []interface{}{}}
	h := newTestHandler(expected, nil)
	r := httptest.NewRequest("GET", "/test?_id=abc123&start_date=2025-01-01&end_date=2025-01-31&media_type=VIDEO&media_product_type=REELS", nil)
	w := httptest.NewRecorder()
	h.HandlePostingActivityTableByType(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}

func TestHandleFollowersGrowthComparison(t *testing.T) {
	expected := map[string]interface{}{"data": []interface{}{}}
	h := newTestHandler(expected, nil)
	r := httptest.NewRequest("GET", "/test?_id=abc123&start_date=2025-01-01&end_date=2025-01-31", nil)
	w := httptest.NewRecorder()
	h.HandleFollowersGrowthComparison(w, r)
	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}
}
