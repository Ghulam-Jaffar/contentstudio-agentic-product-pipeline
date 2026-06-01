package tiktok

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/tiktok"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/tiktok"
)

type mockService struct{}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetPageAndPostsInsights(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": map[string]interface{}{"tiktok_id": "tt_123"}}, nil
}
func (m *mockService) GetPageFollowersAndViews(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockService) GetDynamicPageFollowersAndViews(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockService) GetPostsAndEngagements(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockService) GetDailyEngagementsData(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockService) GetDynamicDailyEngagementsData(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}
func (m *mockService) GetTopAndLeastPerformingPosts(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": map[string]interface{}{"top_posts": []map[string]interface{}{}, "least_posts": []map[string]interface{}{}}}, nil
}
func (m *mockService) GetPostsData(_ context.Context, _ *types.PostsRequest) (map[string]interface{}, error) {
	return map[string]interface{}{"data": []map[string]interface{}{}}, nil
}

func newTestHandler() *Handler {
	return NewHandler(&mockService{}, zerolog.New(io.Discard))
}

const validQueryStr = "workspace_id=ws1&tiktok_id=tt_123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"

func TestParseBaseRequest(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		expectErr bool
	}{
		{name: "valid", query: validQueryStr},
		{name: "valid with date range", query: "workspace_id=ws1&tiktok_id=tt_123&date=2025-01-01+-+2025-01-31&timezone=UTC"},
		{name: "missing workspace", query: "tiktok_id=tt_123&start_date=2025-01-01&end_date=2025-01-31", expectErr: true},
		{name: "missing tiktok id", query: "workspace_id=ws1&start_date=2025-01-01&end_date=2025-01-31", expectErr: true},
		{name: "invalid timezone", query: "workspace_id=ws1&tiktok_id=tt_123&start_date=2025-01-01&end_date=2025-01-31&timezone=Invalid/Zone", expectErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/x?"+tc.query, nil)
			parsed, err := parseBaseRequest(req)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if parsed.WorkspaceID != "ws1" {
				t.Fatalf("expected workspace_id=ws1, got %q", parsed.WorkspaceID)
			}
		})
	}
}

func TestParsePostsRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/x?"+validQueryStr+"&limit=abc", nil)
	if _, err := parsePostsRequest(req); err == nil {
		t.Fatal("expected invalid limit error")
	}

	req = httptest.NewRequest(http.MethodGet, "/x?"+validQueryStr+"&offset=abc", nil)
	if _, err := parsePostsRequest(req); err == nil {
		t.Fatal("expected invalid offset error")
	}
}

func TestHandlePageAndPostsInsights(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/tiktok/getPageAndPostsInsights?"+validQueryStr, nil)
	w := httptest.NewRecorder()

	h.HandlePageAndPostsInsights(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	data := body["data"].(map[string]interface{})
	if data["tiktok_id"] != "tt_123" {
		t.Fatalf("expected tiktok_id tt_123, got %v", data["tiktok_id"])
	}
}

func TestHandlePostsData_BadRequest(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/tiktok/getPostsData?"+validQueryStr+"&limit=abc", nil)
	w := httptest.NewRecorder()

	h.HandlePostsData(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleAIInsights_ServiceUnavailable(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/tiktok/ai_insights?workspace_id=ws1&tiktok_id=tt_123&date=2025-01-01+-+2025-01-31&type=insights_summary", nil)
	w := httptest.NewRecorder()

	h.HandleAIInsights(w, req)
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected status 503, got %d", w.Code)
	}
}
