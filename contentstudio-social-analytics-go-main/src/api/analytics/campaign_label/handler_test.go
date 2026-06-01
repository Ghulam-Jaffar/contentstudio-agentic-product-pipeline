package campaign_label

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/campaign_label"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/campaign_label"
	"github.com/rs/zerolog"
)

// ---------------------------------------------------------------------------
// Mock service
// ---------------------------------------------------------------------------

type mockService struct {
	setPostIdsFn   func(ctx context.Context, req *campaign_label.CampaignLabelRequest) (*campaign_label.SetPostIdsResponse, error)
	getSummaryFn   func(ctx context.Context, req *campaign_label.CampaignLabelRequest) (*campaign_label.SummaryResponse, error)
	getBreakdownFn func(ctx context.Context, req *campaign_label.CampaignLabelRequest) (map[string]interface{}, error)
	getInsightsFn  func(ctx context.Context, req *campaign_label.CampaignLabelRequest) (map[string]interface{}, error)
	getPlannerFn   func(ctx context.Context, req *campaign_label.PlannerAnalyticsRequest) (map[string]interface{}, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) SetPostIds(ctx context.Context, req *campaign_label.CampaignLabelRequest) (*campaign_label.SetPostIdsResponse, error) {
	if m.setPostIdsFn != nil {
		return m.setPostIdsFn(ctx, req)
	}
	return &campaign_label.SetPostIdsResponse{MatchedPostedIds: []string{"post1", "post2"}}, nil
}

func (m *mockService) GetSummaryAnalytics(ctx context.Context, req *campaign_label.CampaignLabelRequest) (*campaign_label.SummaryResponse, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, req)
	}
	return &campaign_label.SummaryResponse{
		Current:    map[string]interface{}{"total_posts": int32(10)},
		Previous:   map[string]interface{}{"total_posts": int32(5)},
		Difference: map[string]interface{}{"total_posts": int64(5)},
		Percentage: map[string]interface{}{"total_posts": float64(100)},
	}, nil
}

func (m *mockService) GetBreakdownData(ctx context.Context, req *campaign_label.CampaignLabelRequest) (map[string]interface{}, error) {
	if m.getBreakdownFn != nil {
		return m.getBreakdownFn(ctx, req)
	}
	return map[string]interface{}{
		"campaign1": map[string]interface{}{
			"current": []campaign_label.BreakdownRow{},
		},
	}, nil
}

func (m *mockService) GetInsightsBreakdown(ctx context.Context, req *campaign_label.CampaignLabelRequest) (map[string]interface{}, error) {
	if m.getInsightsFn != nil {
		return m.getInsightsFn(ctx, req)
	}
	return map[string]interface{}{
		"campaign1": []campaign_label.InsightsRow{},
	}, nil
}

func (m *mockService) GetPlannerAnalytics(ctx context.Context, req *campaign_label.PlannerAnalyticsRequest) (map[string]interface{}, error) {
	if m.getPlannerFn != nil {
		return m.getPlannerFn(ctx, req)
	}
	return map[string]interface{}{"engagement": 100}, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

func newTestHandler() *Handler {
	return NewHandler(&mockService{}, zerolog.New(io.Discard))
}

// validBody returns a valid CampaignLabelRequest JSON body.
func validBody() []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"workspace_id":      "ws1",
		"start_date":        "2025-01-01",
		"end_date":          "2025-01-31",
		"campaigns":         []string{"camp1"},
		"labels":            []string{"label1"},
		"facebook_accounts": []string{"fb_123"},
	})
	return body
}

// validPlannerBody returns a valid PlannerAnalyticsRequest JSON body.
func validPlannerBody() []byte {
	body, _ := json.Marshal(map[string]interface{}{
		"workspace_id": "ws1",
		"id":           "plan1",
		"all_post_ids": []string{"post1"},
		"platforms":    "facebook",
	})
	return body
}

// ---------------------------------------------------------------------------
// parseCampaignLabelBody tests
// ---------------------------------------------------------------------------

func TestParseCampaignLabelBody(t *testing.T) {
	tests := []struct {
		name      string
		body      interface{}
		expectErr bool
	}{
		{
			name: "valid request",
			body: map[string]interface{}{
				"workspace_id": "ws1",
				"start_date":   "2025-01-01",
				"end_date":     "2025-01-31",
			},
		},
		{
			name: "valid with date fallback",
			body: map[string]interface{}{
				"workspace_id": "ws1",
				"date":         "2025-01-01 - 2025-01-31",
			},
		},
		{
			name: "missing workspace_id",
			body: map[string]interface{}{
				"start_date": "2025-01-01",
				"end_date":   "2025-01-31",
			},
			expectErr: true,
		},
		{
			name: "missing dates",
			body: map[string]interface{}{
				"workspace_id": "ws1",
			},
			expectErr: true,
		},
		{
			name:      "invalid body",
			body:      "not json",
			expectErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			var bodyBytes []byte
			switch v := tc.body.(type) {
			case string:
				bodyBytes = []byte(v)
			default:
				bodyBytes, _ = json.Marshal(v)
			}

			req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(bodyBytes))
			req.Header.Set("Content-Type", "application/json")

			parsed, err := parseCampaignLabelBody(req)
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

// ---------------------------------------------------------------------------
// Handler endpoint tests
// ---------------------------------------------------------------------------

func TestHandleSetPostIds(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/analytics/campaignLabelAnalytics/setPostIdsForCampaignsAndLabels", bytes.NewReader(validBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleSetPostIds(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var body campaign_label.SetPostIdsResponse
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if len(body.MatchedPostedIds) != 2 {
		t.Fatalf("expected 2 post IDs, got %d", len(body.MatchedPostedIds))
	}
}

func TestHandleSummaryAnalytics(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/analytics/campaignLabelAnalytics/getSummaryAnalytics", bytes.NewReader(validBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleSummaryAnalytics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
	if _, ok := body["current"]; !ok {
		t.Fatal("expected 'current' key in response")
	}
}

func TestHandleBreakdownData(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/analytics/campaignLabelAnalytics/getCampaignLabelBreakdownData", bytes.NewReader(validBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleBreakdownData(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleInsightsBreakdown(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/analytics/campaignLabelAnalytics/getCampaignLabelInsightsBreakdown", bytes.NewReader(validBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleInsightsBreakdown(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandlePlannerAnalytics(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodPost, "/analytics/campaignLabelAnalytics/getPlannerAnalytics", bytes.NewReader(validPlannerBody()))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandlePlannerAnalytics(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", w.Code, w.Body.String())
	}
	var body map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to decode json: %v", err)
	}
}

// ---------------------------------------------------------------------------
// Bad-request tests
// ---------------------------------------------------------------------------

func TestHandleSetPostIds_BadRequest(t *testing.T) {
	h := newTestHandler()
	// Missing workspace_id
	badBody, _ := json.Marshal(map[string]interface{}{"campaigns": []string{"c1"}})
	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(badBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandleSetPostIds(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandlePlannerAnalytics_BadRequest(t *testing.T) {
	h := newTestHandler()
	// Missing workspace_id
	badBody, _ := json.Marshal(map[string]interface{}{"platforms": "facebook"})
	req := httptest.NewRequest(http.MethodPost, "/x", bytes.NewReader(badBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	h.HandlePlannerAnalytics(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}
