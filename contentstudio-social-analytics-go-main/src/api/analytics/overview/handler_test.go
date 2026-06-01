package overview

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/overview"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/overview"
)

type mockService struct {
	getPlatformDataGroupedFn    func(context.Context, *types.OverviewRequest) ([]*types.PlatformDataRow, error)
	getPlatformDataIndividualFn func(context.Context, *types.OverviewRequest) ([]*types.AccountDataRow, error)
	getTopPostsFn               func(context.Context, *types.TopPostsRequest) ([]*types.TopPostRow, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetSummary(_ context.Context, _ *types.OverviewRequest) (*types.SummaryResponse, error) {
	return &types.SummaryResponse{Summary: &types.SummaryData{}}, nil
}
func (m *mockService) GetTopPerformingGraph(_ context.Context, _ *types.OverviewRequest) (*types.TopPerformingGraphResponse, error) {
	return &types.TopPerformingGraphResponse{}, nil
}
func (m *mockService) GetPlatformDataGrouped(ctx context.Context, req *types.OverviewRequest) ([]*types.PlatformDataRow, error) {
	if m.getPlatformDataGroupedFn != nil {
		return m.getPlatformDataGroupedFn(ctx, req)
	}
	return nil, nil
}
func (m *mockService) GetPlatformDataIndividual(ctx context.Context, req *types.OverviewRequest) ([]*types.AccountDataRow, error) {
	if m.getPlatformDataIndividualFn != nil {
		return m.getPlatformDataIndividualFn(ctx, req)
	}
	return nil, nil
}
func (m *mockService) GetPlatformDataDetailed(_ context.Context, _ *types.OverviewRequest) ([]*types.AccountDataDetailedRow, error) {
	return nil, nil
}
func (m *mockService) GetPlatformDataGraphs(_ context.Context, _ *types.OverviewRequest) ([]*types.AccountDataGraphsRow, error) {
	return nil, nil
}
func (m *mockService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) ([]*types.TopPostRow, error) {
	if m.getTopPostsFn != nil {
		return m.getTopPostsFn(ctx, req)
	}
	return nil, nil
}

func newTestHandler(svc service.Service) *Handler {
	return NewHandler(svc, zerolog.New(io.Discard))
}

func postJSON(t *testing.T, body interface{}, handlerFn http.HandlerFunc) *httptest.ResponseRecorder {
	t.Helper()
	b, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("failed to marshal body: %v", err)
	}
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	handlerFn(w, req)
	return w
}

var validBody = map[string]interface{}{
	"workspace_id":       "ws1",
	"start_date":         "2025-01-01",
	"end_date":           "2025-01-31",
	"timezone":           "UTC",
	"facebook_accounts":  []string{"fb_1"},
	"instagram_accounts": []string{},
	"linkedin_accounts":  []string{},
	"tiktok_accounts":    []string{},
	"youtube_accounts":   []string{},
	"pinterest_accounts": []string{},
}

func TestParseOverviewBody_Valid(t *testing.T) {
	b, _ := json.Marshal(validBody)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	body, err := parseOverviewBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body.StartDate != "2025-01-01" {
		t.Fatalf("expected start_date 2025-01-01, got %q", body.StartDate)
	}
	if body.EndDate != "2025-01-31" {
		t.Fatalf("expected end_date 2025-01-31, got %q", body.EndDate)
	}
	if body.WorkspaceID != "ws1" {
		t.Fatalf("expected workspace_id ws1, got %q", body.WorkspaceID)
	}
	if len(body.FacebookAccounts) != 1 || body.FacebookAccounts[0] != "fb_1" {
		t.Fatalf("expected facebook_accounts [fb_1], got %v", body.FacebookAccounts)
	}
}

func TestParseOverviewBody_MissingStartDate(t *testing.T) {
	b, _ := json.Marshal(map[string]interface{}{"end_date": "2025-01-31"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	_, err := parseOverviewBody(req)
	if err == nil {
		t.Fatal("expected error for missing start_date")
	}
}

func TestParseOverviewBody_MissingEndDate(t *testing.T) {
	b, _ := json.Marshal(map[string]interface{}{"start_date": "2025-01-01"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	_, err := parseOverviewBody(req)
	if err == nil {
		t.Fatal("expected error for missing end_date")
	}
}

func TestParseOverviewBody_InvalidJSON(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte("not json")))
	_, err := parseOverviewBody(req)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestParseOverviewBody_TypeAndLimitFields(t *testing.T) {
	b, _ := json.Marshal(map[string]interface{}{
		"start_date": "2025-01-01",
		"end_date":   "2025-01-31",
		"type":       "likes",
		"limit":      10,
	})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	body, err := parseOverviewBody(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body.Type != "likes" {
		t.Fatalf("expected type \"likes\", got %q", body.Type)
	}
	if body.Limit != 10 {
		t.Fatalf("expected limit 10, got %d", body.Limit)
	}
}

func TestHandleSummary_Post(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := postJSON(t, validBody, h.HandleSummary)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp types.SummaryResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if resp.Summary == nil {
		t.Fatal("expected non-nil summary in response")
	}
}

func TestHandleSummary_MissingDates(t *testing.T) {
	h := newTestHandler(&mockService{})
	b, _ := json.Marshal(map[string]interface{}{"workspace_id": "ws1"})
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(b))
	w := httptest.NewRecorder()
	h.HandleSummary(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleTopPerformingGraph_Post(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := postJSON(t, validBody, h.HandleTopPerformingGraph)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlePlatformData_Grouped(t *testing.T) {
	var called bool
	svc := &mockService{
		getPlatformDataGroupedFn: func(_ context.Context, _ *types.OverviewRequest) ([]*types.PlatformDataRow, error) {
			called = true
			return nil, nil
		},
	}
	body := map[string]interface{}{
		"start_date": "2025-01-01",
		"end_date":   "2025-01-31",
		"type":       "grouped",
	}
	h := newTestHandler(svc)
	w := postJSON(t, body, h.HandlePlatformData)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected GetPlatformDataGrouped to be called")
	}
}

func TestHandlePlatformData_Individual(t *testing.T) {
	var called bool
	svc := &mockService{
		getPlatformDataIndividualFn: func(_ context.Context, _ *types.OverviewRequest) ([]*types.AccountDataRow, error) {
			called = true
			return nil, nil
		},
	}
	body := map[string]interface{}{
		"start_date": "2025-01-01",
		"end_date":   "2025-01-31",
		"type":       "individual",
	}
	h := newTestHandler(svc)
	w := postJSON(t, body, h.HandlePlatformData)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Fatal("expected GetPlatformDataIndividual to be called")
	}
}

func TestHandlePlatformDataDetailed_Post(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := postJSON(t, validBody, h.HandlePlatformDataDetailed)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandlePlatformDataGraphs_Post(t *testing.T) {
	h := newTestHandler(&mockService{})
	w := postJSON(t, validBody, h.HandlePlatformDataGraphs)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestHandleTopPosts_DefaultLimit(t *testing.T) {
	var gotLimit int
	svc := &mockService{
		getTopPostsFn: func(_ context.Context, req *types.TopPostsRequest) ([]*types.TopPostRow, error) {
			gotLimit = req.Limit
			return nil, nil
		},
	}
	body := map[string]interface{}{
		"start_date": "2025-01-01",
		"end_date":   "2025-01-31",
	}
	h := newTestHandler(svc)
	w := postJSON(t, body, h.HandleTopPosts)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotLimit != 20 {
		t.Fatalf("expected default limit 20, got %d", gotLimit)
	}
}

func TestHandleTopPosts_CustomLimitAndType(t *testing.T) {
	var gotLimit int
	var gotType string
	svc := &mockService{
		getTopPostsFn: func(_ context.Context, req *types.TopPostsRequest) ([]*types.TopPostRow, error) {
			gotLimit = req.Limit
			gotType = req.Type
			return nil, nil
		},
	}
	body := map[string]interface{}{
		"start_date": "2025-01-01",
		"end_date":   "2025-01-31",
		"type":       "likes",
		"limit":      10,
	}
	h := newTestHandler(svc)
	w := postJSON(t, body, h.HandleTopPosts)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if gotLimit != 10 {
		t.Fatalf("expected limit 10, got %d", gotLimit)
	}
	if gotType != "likes" {
		t.Fatalf("expected type \"likes\", got %q", gotType)
	}
}
