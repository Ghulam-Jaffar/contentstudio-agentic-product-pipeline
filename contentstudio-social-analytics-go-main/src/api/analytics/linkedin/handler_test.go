package linkedin

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/linkedin"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/linkedin"
)

// --- Mock service ---

type mockService struct {
	getSummaryFn              func(ctx context.Context, req *types.LinkedInRequest) (*types.SummaryResponse, error)
	getAudienceGrowthFn       func(ctx context.Context, req *types.LinkedInRequest) (*types.AudienceGrowthResponse, error)
	getPageViewsFn            func(ctx context.Context, req *types.LinkedInRequest) (*types.PageViewsResponse, error)
	getPublishingBehaviourFn  func(ctx context.Context, req *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error)
	getTopPostsFn             func(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error)
	getPostsPerDayFn          func(ctx context.Context, req *types.LinkedInRequest) (*types.PostsPerDayResponse, error)
	getHashtagsFn             func(ctx context.Context, req *types.LinkedInRequest) (*types.HashtagsResponse, error)
	getFollowersDemographicFn func(ctx context.Context, req *types.LinkedInRequest) (*types.DemographicsResponse, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetSummary(ctx context.Context, req *types.LinkedInRequest) (*types.SummaryResponse, error) {
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

func (m *mockService) GetAudienceGrowth(ctx context.Context, req *types.LinkedInRequest) (*types.AudienceGrowthResponse, error) {
	if m.getAudienceGrowthFn != nil {
		return m.getAudienceGrowthFn(ctx, req)
	}
	return &types.AudienceGrowthResponse{
		Status:         true,
		AudienceGrowth: &types.AudienceGrowthData{},
		AudienceGrowthRollup: map[string]*types.AudienceGrowthRollup{
			"current":  {},
			"previous": {},
		},
	}, nil
}

func (m *mockService) GetPageViews(ctx context.Context, req *types.LinkedInRequest) (*types.PageViewsResponse, error) {
	if m.getPageViewsFn != nil {
		return m.getPageViewsFn(ctx, req)
	}
	return &types.PageViewsResponse{
		Status:    true,
		PageViews: &types.PageViewsData{},
		PageViewsRollup: map[string]*types.PageViewsRollup{
			"current":  {},
			"previous": {},
		},
	}, nil
}

func (m *mockService) GetPublishingBehaviour(ctx context.Context, req *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error) {
	if m.getPublishingBehaviourFn != nil {
		return m.getPublishingBehaviourFn(ctx, req)
	}
	return &types.PublishingBehaviourResponse{
		Status:              true,
		PublishingBehaviour: &types.PublishingBehaviourData{},
		PublishingBehaviourRollup: map[string][]types.PublishingBehaviourMediaType{
			"current":  {},
			"previous": {},
		},
	}, nil
}

func (m *mockService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
	if m.getTopPostsFn != nil {
		return m.getTopPostsFn(ctx, req)
	}
	return &types.TopPostsResponse{Status: true, TopPosts: []types.TopPost{}}, nil
}

func (m *mockService) GetPostsPerDay(ctx context.Context, req *types.LinkedInRequest) (*types.PostsPerDayResponse, error) {
	if m.getPostsPerDayFn != nil {
		return m.getPostsPerDayFn(ctx, req)
	}
	return &types.PostsPerDayResponse{
		Status: true,
		PostsPerDays: &types.PostsPerDayData{
			Data: types.PostsPerDayInner{
				Days: map[string]int32{},
			},
		},
	}, nil
}

func (m *mockService) GetHashtags(ctx context.Context, req *types.LinkedInRequest) (*types.HashtagsResponse, error) {
	if m.getHashtagsFn != nil {
		return m.getHashtagsFn(ctx, req)
	}
	return &types.HashtagsResponse{
		Status:      true,
		TopHashtags: &types.HashtagsData{},
		TopHashtagsRollup: map[string]*types.HashtagsRollup{
			"current":  {},
			"previous": {},
		},
	}, nil
}

func (m *mockService) GetFollowersDemographics(ctx context.Context, req *types.LinkedInRequest) (*types.DemographicsResponse, error) {
	if m.getFollowersDemographicFn != nil {
		return m.getFollowersDemographicFn(ctx, req)
	}
	return &types.DemographicsResponse{
		Status:               true,
		FollowerDemographics: map[string]*types.DemographicCategory{},
	}, nil
}

// --- Test helpers ---

func newTestHandler() *LinkedInHandler {
	return NewLinkedInHandler(&mockService{}, zerolog.New(io.Discard))
}

func newTestHandlerWithService(svc service.Service) *LinkedInHandler {
	return NewLinkedInHandler(svc, zerolog.New(io.Discard))
}

const validQueryStr = "workspace_id=ws1&linkedin_id=li_123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"
const validDateRangeQueryStr = "workspace_id=ws1&linkedin_id=li_123&date=2025-01-01+-+2025-01-31&timezone=UTC"

// --- Tests ---

func TestNewLinkedInHandler(t *testing.T) {
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
			name:  "valid date range request",
			query: validDateRangeQueryStr,
		},
		{
			name:      "missing workspace_id",
			query:     "linkedin_id=li_123&start_date=2025-01-01&end_date=2025-01-31",
			expectErr: true,
		},
		{
			name:      "missing linkedin_id",
			query:     "workspace_id=ws1&start_date=2025-01-01&end_date=2025-01-31",
			expectErr: true,
		},
		{
			name:      "missing start_date",
			query:     "workspace_id=ws1&linkedin_id=li_123&end_date=2025-01-31",
			expectErr: true,
		},
		{
			name:      "missing end_date",
			query:     "workspace_id=ws1&linkedin_id=li_123&start_date=2025-01-01",
			expectErr: true,
		},
		{
			name:      "invalid date format",
			query:     "workspace_id=ws1&linkedin_id=li_123&start_date=01-01-2025&end_date=2025-01-31",
			expectErr: true,
		},
		{
			name:      "end before start",
			query:     "workspace_id=ws1&linkedin_id=li_123&start_date=2025-02-01&end_date=2025-01-01",
			expectErr: true,
		},
		{
			name:      "invalid timezone",
			query:     "workspace_id=ws1&linkedin_id=li_123&start_date=2025-01-01&end_date=2025-01-31&timezone=Invalid/Zone",
			expectErr: true,
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
			if parsed.LinkedinID != "li_123" {
				t.Fatalf("expected li_123, got %q", parsed.LinkedinID)
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

func TestHandleAudienceGrowth(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
		{name: "success with date range", query: validDateRangeQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/audienceGrowth?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleAudienceGrowth(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandlePageViews(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
		{name: "success with date range", query: validDateRangeQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/pageViews?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandlePageViews(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandlePublishingBehaviour(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
		{name: "success with date range", query: validDateRangeQueryStr, expectedStatus: http.StatusOK},
		{name: "with media_type filter", query: validQueryStr + "&media_type=images,videos", expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/publishingBehaviour?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandlePublishingBehaviour(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
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
		{name: "success with date range", query: validDateRangeQueryStr, expectedStatus: http.StatusOK},
		{name: "with limit and order_by", query: validQueryStr + "&limit=5&order_by=impressions", expectedStatus: http.StatusOK},
		{name: "with hashtags", query: validQueryStr + "&hashtags=tech,marketing", expectedStatus: http.StatusOK},
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

func TestHandleGetTopPosts(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
		{name: "success with date range", query: validDateRangeQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/getTopPosts?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleGetTopPosts(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandlePostsPerDay(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
		{name: "success with date range", query: validDateRangeQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/postsPerDays?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandlePostsPerDay(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleHashtags(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
		{name: "success with date range", query: validDateRangeQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/hashtags?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleHashtags(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleFollowersDemographics(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		expectedStatus int
	}{
		{name: "missing params", query: "", expectedStatus: http.StatusBadRequest},
		{name: "success", query: validQueryStr, expectedStatus: http.StatusOK},
		{name: "success with date range", query: validDateRangeQueryStr, expectedStatus: http.StatusOK},
	}

	h := newTestHandler()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/followersDemographics?"+tc.query, nil)
			w := httptest.NewRecorder()
			h.HandleFollowersDemographics(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestHandleAIInsights_NotConfigured(t *testing.T) {
	h := newTestHandler()

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/overview/linkedin/ai_insights?workspace_id=ws1&linkedin_id=li_123&date=2025-01-01+-+2025-01-31&type=page_views",
		nil,
	)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d", w.Code)
	}
}

func TestHandleAIInsights_MissingFields(t *testing.T) {
	h := newTestHandler()
	h.SetAIInsightsService(service.NewAIInsightsService(nil, nil, nil))

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/overview/linkedin/ai_insights?workspace_id=ws1&linkedin_id=li_123",
		nil,
	)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestHandleAIInsights_InvalidLimit(t *testing.T) {
	h := newTestHandler()
	h.SetAIInsightsService(service.NewAIInsightsService(nil, nil, nil))

	req := httptest.NewRequest(
		http.MethodGet,
		"/analytics/overview/linkedin/ai_insights?workspace_id=ws1&linkedin_id=li_123&date=2025-01-01+-+2025-01-31&type=page_views&limit=abc",
		nil,
	)
	w := httptest.NewRecorder()
	h.HandleAIInsights(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
