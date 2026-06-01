package gmb

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/gmb"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/gmb"
)

// --- Mock service ---

type mockService struct {
	getSummaryFn            func(ctx context.Context, req *types.GMBRequest) (*types.SummaryResponse, error)
	getImpressionsFn        func(ctx context.Context, req *types.GMBRequest) (*types.ImpressionsResponse, error)
	getActionsFn            func(ctx context.Context, req *types.GMBRequest) (*types.ActionsResponse, error)
	getSearchKeywordsFn     func(ctx context.Context, req *types.SearchKeywordsRequest) (*types.SearchKeywordsResponse, error)
	getTopPostsFn           func(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error)
	getPublishingBehaviorFn func(ctx context.Context, req *types.GMBRequest) (*types.PublishingBehaviorResponse, error)
	getReviewsFn            func(ctx context.Context, req *types.GMBRequest) (*types.ReviewsResponse, error)
	getMediaActivityFn      func(ctx context.Context, req *types.GMBRequest) (*types.MediaActivityResponse, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetSummary(ctx context.Context, req *types.GMBRequest) (*types.SummaryResponse, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, req)
	}
	return &types.SummaryResponse{
		Status: true,
		Overview: map[string]*types.SummaryMetrics{
			"current":  {},
			"previous": {},
		},
	}, nil
}

func (m *mockService) GetImpressions(ctx context.Context, req *types.GMBRequest) (*types.ImpressionsResponse, error) {
	if m.getImpressionsFn != nil {
		return m.getImpressionsFn(ctx, req)
	}
	return &types.ImpressionsResponse{
		Status:      true,
		Impressions: &types.ImpressionsData{},
		ImpressionsRolup: map[string]*types.ImpressionsRollupData{
			"current":  {},
			"previous": {},
		},
	}, nil
}

func (m *mockService) GetActions(ctx context.Context, req *types.GMBRequest) (*types.ActionsResponse, error) {
	if m.getActionsFn != nil {
		return m.getActionsFn(ctx, req)
	}
	return &types.ActionsResponse{
		Status:  true,
		Actions: &types.ActionsData{},
		ActionsRollup: map[string]*types.ActionsRollupData{
			"current":  {},
			"previous": {},
		},
	}, nil
}

func (m *mockService) GetSearchKeywords(ctx context.Context, req *types.SearchKeywordsRequest) (*types.SearchKeywordsResponse, error) {
	if m.getSearchKeywordsFn != nil {
		return m.getSearchKeywordsFn(ctx, req)
	}
	return &types.SearchKeywordsResponse{Status: true, Keywords: []types.SearchKeyword{}}, nil
}

func (m *mockService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
	if m.getTopPostsFn != nil {
		return m.getTopPostsFn(ctx, req)
	}
	return &types.TopPostsResponse{Status: true, Posts: []types.TopPost{}}, nil
}

func (m *mockService) GetPublishingBehavior(ctx context.Context, req *types.GMBRequest) (*types.PublishingBehaviorResponse, error) {
	if m.getPublishingBehaviorFn != nil {
		return m.getPublishingBehaviorFn(ctx, req)
	}
	return &types.PublishingBehaviorResponse{
		Status:              true,
		PublishingBehaviour: &types.PublishingBehaviorData{},
	}, nil
}

func (m *mockService) GetReviews(ctx context.Context, req *types.GMBRequest) (*types.ReviewsResponse, error) {
	if m.getReviewsFn != nil {
		return m.getReviewsFn(ctx, req)
	}
	return &types.ReviewsResponse{
		Status:  true,
		Reviews: &types.ReviewsData{},
		ReviewsRollup: map[string]*types.ReviewsRollupData{
			"current":  {},
			"previous": {},
		},
	}, nil
}

func (m *mockService) GetMediaActivity(ctx context.Context, req *types.GMBRequest) (*types.MediaActivityResponse, error) {
	if m.getMediaActivityFn != nil {
		return m.getMediaActivityFn(ctx, req)
	}
	return &types.MediaActivityResponse{
		Status:        true,
		MediaActivity: &types.MediaActivityData{},
		MediaActivityRollup: map[string]*types.MediaActivityRollupData{
			"current":  {},
			"previous": {},
		},
	}, nil
}

// --- Test helpers ---

func newTestHandler() *GMBHandler {
	return NewGMBHandler(&mockService{}, zerolog.New(io.Discard))
}

func newTestHandlerWithService(svc service.Service) *GMBHandler {
	return NewGMBHandler(svc, zerolog.New(io.Discard))
}

const validQueryStr = "workspace_id=ws1&gmb_id=loc_123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"

// --- Tests ---

func TestNewGMBHandler(t *testing.T) {
	h := newTestHandler()
	if h == nil {
		t.Fatal("expected non-nil handler")
	}
}

func TestParseBaseRequest(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		expectErr bool
	}{
		{
			name:  "valid request",
			query: validQueryStr,
		},
		{
			name:      "missing workspace_id",
			query:     "gmb_id=loc_123&start_date=2025-01-01&end_date=2025-01-31",
			expectErr: true,
		},
		{
			name:      "missing gmb_id",
			query:     "workspace_id=ws1&start_date=2025-01-01&end_date=2025-01-31",
			expectErr: true,
		},
		{
			name:      "missing start_date",
			query:     "workspace_id=ws1&gmb_id=loc_123&end_date=2025-01-31",
			expectErr: true,
		},
		{
			name:      "missing end_date",
			query:     "workspace_id=ws1&gmb_id=loc_123&start_date=2025-01-01",
			expectErr: true,
		},
		{
			name:      "invalid date format",
			query:     "workspace_id=ws1&gmb_id=loc_123&start_date=01-01-2025&end_date=2025-01-31",
			expectErr: true,
		},
		{
			name:      "end before start",
			query:     "workspace_id=ws1&gmb_id=loc_123&start_date=2025-02-01&end_date=2025-01-01",
			expectErr: true,
		},
		{
			name:      "invalid timezone",
			query:     "workspace_id=ws1&gmb_id=loc_123&start_date=2025-01-01&end_date=2025-01-31&timezone=Invalid/Zone",
			expectErr: true,
		},
		{
			name:  "empty timezone defaults to UTC",
			query: "workspace_id=ws1&gmb_id=loc_123&start_date=2025-01-01&end_date=2025-01-31",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test?"+tc.query, nil)
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
				t.Fatalf("expected ws1, got %q", parsed.WorkspaceID)
			}
			if parsed.GmbID != "loc_123" {
				t.Fatalf("expected loc_123, got %q", parsed.GmbID)
			}
		})
	}
}

func TestHandleSummary(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
		checkResp      func(t *testing.T, body []byte)
	}{
		{
			name:           "missing params",
			query:          "",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "success",
			query:          validQueryStr,
			expectedStatus: http.StatusOK,
			checkResp: func(t *testing.T, body []byte) {
				var resp map[string]interface{}
				json.Unmarshal(body, &resp)
				if resp["status"] != true {
					t.Fatalf("expected status true, got %v", resp["status"])
				}
				overview := resp["overview"].(map[string]interface{})
				if _, ok := overview["current"]; !ok {
					t.Fatal("expected 'current' in overview")
				}
				if _, ok := overview["previous"]; !ok {
					t.Fatal("expected 'previous' in overview")
				}
			},
		},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/summary?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleSummary(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
			if tc.checkResp != nil {
				tc.checkResp(t, w.Body.Bytes())
			}
		})
	}
}

func TestHandleSummary_ServiceError(t *testing.T) {
	svc := &mockService{
		getSummaryFn: func(ctx context.Context, req *types.GMBRequest) (*types.SummaryResponse, error) {
			return nil, errors.New("db connection failed")
		},
	}
	h := newTestHandlerWithService(svc)

	req := httptest.NewRequest("GET", "/summary?"+validQueryStr, nil)
	w := httptest.NewRecorder()
	h.HandleSummary(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
	}
}

func TestHandleImpressions(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/impressions?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleImpressions(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleImpressions_ResponseShape(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/impressions?"+validQueryStr, nil)
	w := httptest.NewRecorder()
	h.HandleImpressions(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != true {
		t.Fatalf("expected status true, got %v", resp["status"])
	}
	if _, ok := resp["impressions"]; !ok {
		t.Fatal("expected 'impressions' in response")
	}
	if _, ok := resp["impressions_rollup"]; !ok {
		t.Fatal("expected 'impressions_rollup' in response")
	}
}

func TestHandleActions(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/actions?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleActions(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleActions_ResponseShape(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/actions?"+validQueryStr, nil)
	w := httptest.NewRecorder()
	h.HandleActions(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != true {
		t.Fatalf("expected status true, got %v", resp["status"])
	}
	if _, ok := resp["actions"]; !ok {
		t.Fatal("expected 'actions' in response")
	}
	if _, ok := resp["actions_rollup"]; !ok {
		t.Fatal("expected 'actions_rollup' in response")
	}
}

func TestHandleSearchKeywords(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success with defaults", query: validQueryStr, expectedStatus: http.StatusOK},
		{name: "with limit", query: validQueryStr + "&limit=10", expectedStatus: http.StatusOK},
		{name: "invalid limit returns bad request", query: validQueryStr + "&limit=abc", expectedStatus: http.StatusBadRequest},
		{name: "zero limit defaults gracefully", query: validQueryStr + "&limit=0", expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/searchKeywords?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleSearchKeywords(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleSearchKeywords_LimitParsing(t *testing.T) {
	var capturedReq *types.SearchKeywordsRequest
	svc := &mockService{
		getSearchKeywordsFn: func(ctx context.Context, req *types.SearchKeywordsRequest) (*types.SearchKeywordsResponse, error) {
			capturedReq = req
			return &types.SearchKeywordsResponse{Status: true, Keywords: []types.SearchKeyword{}}, nil
		},
	}
	h := newTestHandlerWithService(svc)

	req := httptest.NewRequest("GET", "/searchKeywords?"+validQueryStr+"&limit=15", nil)
	w := httptest.NewRecorder()
	h.HandleSearchKeywords(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
	if capturedReq == nil {
		t.Fatal("expected request to be captured")
	}
	if capturedReq.Limit != 15 {
		t.Fatalf("expected limit 15, got %d", capturedReq.Limit)
	}
}

func TestHandleTopPosts(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success with defaults", query: validQueryStr, expectedStatus: http.StatusOK},
		{name: "with limit and order_by", query: validQueryStr + "&limit=5&order_by=created_at", expectedStatus: http.StatusOK},
		{name: "invalid limit returns bad request", query: validQueryStr + "&limit=abc", expectedStatus: http.StatusBadRequest},
		{name: "negative limit defaults gracefully", query: validQueryStr + "&limit=-1", expectedStatus: http.StatusOK},
		{name: "zero limit defaults gracefully", query: validQueryStr + "&limit=0", expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/topPosts?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleTopPosts(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleTopPosts_ParamParsing(t *testing.T) {
	var capturedReq *types.TopPostsRequest
	svc := &mockService{
		getTopPostsFn: func(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
			capturedReq = req
			return &types.TopPostsResponse{Status: true, Posts: []types.TopPost{}}, nil
		},
	}
	h := newTestHandlerWithService(svc)

	req := httptest.NewRequest("GET", "/topPosts?"+validQueryStr+"&limit=10&order_by=created_at", nil)
	w := httptest.NewRecorder()
	h.HandleTopPosts(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, w.Code)
	}
	if capturedReq == nil {
		t.Fatal("expected request to be captured")
	}
	if capturedReq.Limit != 10 {
		t.Fatalf("expected limit 10, got %d", capturedReq.Limit)
	}
	if capturedReq.OrderBy != "created_at" {
		t.Fatalf("expected order_by created_at, got %q", capturedReq.OrderBy)
	}
}

func TestHandlePublishingBehavior(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/publishingBehavior?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandlePublishingBehavior(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleReviews(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/reviews?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleReviews(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleReviews_ResponseShape(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/reviews?"+validQueryStr, nil)
	w := httptest.NewRecorder()
	h.HandleReviews(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != true {
		t.Fatalf("expected status true, got %v", resp["status"])
	}
	if _, ok := resp["reviews"]; !ok {
		t.Fatal("expected 'reviews' in response")
	}
	if _, ok := resp["reviews_rollup"]; !ok {
		t.Fatal("expected 'reviews_rollup' in response")
	}
}

func TestHandleMediaActivity(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/mediaActivity?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleMediaActivity(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleMediaActivity_ResponseShape(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest("GET", "/mediaActivity?"+validQueryStr, nil)
	w := httptest.NewRecorder()
	h.HandleMediaActivity(w, req)

	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)

	if resp["status"] != true {
		t.Fatalf("expected status true, got %v", resp["status"])
	}
	if _, ok := resp["media_activity"]; !ok {
		t.Fatal("expected 'media_activity' in response")
	}
	if _, ok := resp["media_activity_rollup"]; !ok {
		t.Fatal("expected 'media_activity_rollup' in response")
	}
}

// --- AI Insights handler tests ---

func TestHandleAIInsights_NotConfigured(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/gmb/ai_insights?workspace_id=ws1&gmb_id=loc_123&date=2025-01-01+-+2025-01-31&type=impressions_overview", nil)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}

	var resp map[string]interface{}
	json.NewDecoder(w.Body).Decode(&resp)
	if resp["status"] != false {
		t.Fatal("expected status=false")
	}
}

func TestHandleAIInsights_MissingRequiredFields(t *testing.T) {
	h := newTestHandler()
	h.SetAIInsightsService(service.NewAIInsightsService(nil, nil, nil))

	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/gmb/ai_insights?workspace_id=ws1", nil)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAIInsights_MissingFields(t *testing.T) {
	h := newTestHandler()
	h.SetAIInsightsService(service.NewAIInsightsService(nil, nil, nil))

	tests := []struct {
		name  string
		query string
	}{
		{"missing workspace_id", "gmb_id=loc&type=test"},
		{"missing gmb_id", "workspace_id=ws&type=test"},
		{"missing type", "workspace_id=ws&gmb_id=loc"},
		{"all empty", ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/analytics/overview/gmb/ai_insights?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleAIInsights(w, req)

			if w.Code != http.StatusBadRequest {
				t.Fatalf("expected 400, got %d", w.Code)
			}
		})
	}
}

func TestHandleAIInsights_ServiceError(t *testing.T) {
	// Use a real AIInsightsService with nil dependencies — will fail on GetAIInsights
	// because agentClient is nil. This tests the 500 error path.
	aiSvc := service.NewAIInsightsService(nil, nil, nil)

	h := newTestHandler()
	h.SetAIInsightsService(aiSvc)
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/gmb/ai_insights?workspace_id=ws1&gmb_id=loc_123&date=2025-01-01+-+2025-01-31&type=impressions_overview", nil)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestHandleAIInsights_ServiceErrorWithLimit(t *testing.T) {
	aiSvc := service.NewAIInsightsService(nil, nil, nil)

	h := newTestHandler()
	h.SetAIInsightsService(aiSvc)
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/gmb/ai_insights?workspace_id=ws1&gmb_id=loc_123&date=2025-01-01+-+2025-01-31&timezone=UTC&type=impressions_overview&limit=5", nil)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", w.Code)
	}
}

func TestAllHandlers_ServiceError(t *testing.T) {
	dbErr := errors.New("db connection failed")
	svc := &mockService{
		getSummaryFn: func(ctx context.Context, req *types.GMBRequest) (*types.SummaryResponse, error) {
			return nil, dbErr
		},
		getImpressionsFn: func(ctx context.Context, req *types.GMBRequest) (*types.ImpressionsResponse, error) {
			return nil, dbErr
		},
		getActionsFn: func(ctx context.Context, req *types.GMBRequest) (*types.ActionsResponse, error) {
			return nil, dbErr
		},
		getSearchKeywordsFn: func(ctx context.Context, req *types.SearchKeywordsRequest) (*types.SearchKeywordsResponse, error) {
			return nil, dbErr
		},
		getTopPostsFn: func(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
			return nil, dbErr
		},
		getPublishingBehaviorFn: func(ctx context.Context, req *types.GMBRequest) (*types.PublishingBehaviorResponse, error) {
			return nil, dbErr
		},
		getReviewsFn: func(ctx context.Context, req *types.GMBRequest) (*types.ReviewsResponse, error) {
			return nil, dbErr
		},
		getMediaActivityFn: func(ctx context.Context, req *types.GMBRequest) (*types.MediaActivityResponse, error) {
			return nil, dbErr
		},
	}
	h := newTestHandlerWithService(svc)

	handlers := []struct {
		name    string
		path    string
		handler func(w http.ResponseWriter, r *http.Request)
	}{
		{name: "summary", path: "/summary", handler: h.HandleSummary},
		{name: "impressions", path: "/impressions", handler: h.HandleImpressions},
		{name: "actions", path: "/actions", handler: h.HandleActions},
		{name: "searchKeywords", path: "/searchKeywords", handler: h.HandleSearchKeywords},
		{name: "topPosts", path: "/topPosts", handler: h.HandleTopPosts},
		{name: "publishingBehavior", path: "/publishingBehavior", handler: h.HandlePublishingBehavior},
		{name: "reviews", path: "/reviews", handler: h.HandleReviews},
		{name: "mediaActivity", path: "/mediaActivity", handler: h.HandleMediaActivity},
	}

	for _, tc := range handlers {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tc.path+"?"+validQueryStr, nil)
			w := httptest.NewRecorder()
			tc.handler(w, req)

			if w.Code != http.StatusInternalServerError {
				t.Fatalf("expected %d, got %d", http.StatusInternalServerError, w.Code)
			}

			var resp map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &resp)
			if resp["status"] != false {
				t.Fatalf("expected status false, got %v", resp["status"])
			}
		})
	}
}
