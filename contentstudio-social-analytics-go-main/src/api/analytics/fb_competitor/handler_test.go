package fb_competitor

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/fb_competitor"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/fb_competitor"
)

type mockService struct{}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetDataTableMetrics(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}, "data_prev": []interface{}{}, "data_table_metrics": []interface{}{}}, nil
}
func (m *mockService) GetPostingActivityGraphByTypes(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetPostingActivityBySpecificType(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetTopAndLeastPerformingPosts(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetTopHashtags(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetIndividualHashtagData(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetBiographyData(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetFollowersGrowthComparison(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetPostReactDistribution(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetPostReactDistributionByCompany(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetPostTypeDistribution(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetPostEngagementOverTime(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}
func (m *mockService) GetPostEngagementByCompetitor(context.Context, *types.CompetitorRequest) (interface{}, error) {
	return map[string]interface{}{"data": []interface{}{}}, nil
}

func newTestHandler() *Handler {
	return NewHandler(&mockService{}, zerolog.New(io.Discard))
}

func TestParseRequest_Valid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/analytics/overview/facebook/competitor/dataTableMetrics?_id=abc123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC",
		nil)
	parsed, err := parseRequest(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed.ReportID != "abc123" {
		t.Fatalf("expected report_id abc123, got %q", parsed.ReportID)
	}
	if parsed.StartDate != "2025-01-01" {
		t.Fatalf("expected start_date 2025-01-01, got %q", parsed.StartDate)
	}
}

func TestParseRequest_MissingDates(t *testing.T) {
	// Both dates absent is valid (e.g. biographyData)
	req := httptest.NewRequest(http.MethodGet,
		"/analytics/overview/facebook/competitor/biographyData?_id=abc123&timezone=UTC",
		nil)
	parsed, err := parseRequest(req)
	if err != nil {
		t.Fatalf("expected no error for missing dates, got %v", err)
	}
	if parsed.StartDate != "" || parsed.EndDate != "" {
		t.Fatalf("expected empty dates, got start=%q end=%q", parsed.StartDate, parsed.EndDate)
	}
}

func TestParseRequest_PartialDates(t *testing.T) {
	// Only one date provided should error
	req := httptest.NewRequest(http.MethodGet,
		"/analytics/overview/facebook/competitor/dataTableMetrics?_id=abc123&start_date=2025-01-01&timezone=UTC",
		nil)
	_, err := parseRequest(req)
	if err == nil {
		t.Fatal("expected validation error for partial dates")
	}
}

func TestParseRequest_InvalidLimit(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet,
		"/analytics/overview/facebook/competitor/topHashtags?_id=abc123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC&limit=abc",
		nil)
	_, err := parseRequest(req)
	if err == nil {
		t.Fatal("expected validation error for invalid limit")
	}
}

func TestHandleDataTableMetrics(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet,
		"/analytics/overview/facebook/competitor/dataTableMetrics?_id=abc123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC",
		nil)
	w := httptest.NewRecorder()
	h.HandleDataTableMetrics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if _, ok := resp["data"]; !ok {
		t.Fatal("expected 'data' key in response")
	}
}

func TestHandlePostingActivityGraphByTypes(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet,
		"/analytics/overview/facebook/competitor/postingActivityGraphByTypes?_id=abc123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC",
		nil)
	w := httptest.NewRecorder()
	h.HandlePostingActivityGraphByTypes(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlePostReactDistribution_WithFacebookID(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet,
		"/analytics/overview/facebook/competitor/postReactDistribution?facebook_id=12345&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC",
		nil)
	w := httptest.NewRecorder()
	h.HandlePostReactDistribution(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}
