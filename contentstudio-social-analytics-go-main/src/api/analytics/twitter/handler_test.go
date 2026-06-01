package twitter

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/twitter"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/twitter"
)

type mockService struct {
	getPageAndPostsInsightsFn     func(ctx context.Context, req *types.TwitterRequest) (*types.MetricsResponse, error)
	getEngagementImpressionDataFn func(ctx context.Context, req *types.TwitterRequest) (*types.EngagementImpressionResponse, error)
	getFollowersTrendDataFn       func(ctx context.Context, req *types.TwitterRequest) (*types.FollowersTrendResponse, error)
	getTopTweetsFn                func(ctx context.Context, req *types.TweetsRequest) (*types.TopTweetsResponse, error)
	getLeastTweetsFn              func(ctx context.Context, req *types.TweetsRequest) (*types.LeastTweetsResponse, error)
	getCreditsUsedCountFn         func(ctx context.Context, req *types.TwitterRequest) (*types.CreditsUsedResponse, error)
}

var _ service.Service = (*mockService)(nil)

func (m *mockService) GetPageAndPostsInsights(ctx context.Context, req *types.TwitterRequest) (*types.MetricsResponse, error) {
	if m.getPageAndPostsInsightsFn != nil {
		return m.getPageAndPostsInsightsFn(ctx, req)
	}
	return &types.MetricsResponse{Data: map[string]interface{}{"twitter_id": "tw_123"}}, nil
}

func (m *mockService) GetEngagementImpressionData(ctx context.Context, req *types.TwitterRequest) (*types.EngagementImpressionResponse, error) {
	if m.getEngagementImpressionDataFn != nil {
		return m.getEngagementImpressionDataFn(ctx, req)
	}
	return &types.EngagementImpressionResponse{TwitterID: "tw_123"}, nil
}

func (m *mockService) GetFollowersTrendData(ctx context.Context, req *types.TwitterRequest) (*types.FollowersTrendResponse, error) {
	if m.getFollowersTrendDataFn != nil {
		return m.getFollowersTrendDataFn(ctx, req)
	}
	return &types.FollowersTrendResponse{PlatformID: "tw_123"}, nil
}

func (m *mockService) GetTopTweets(ctx context.Context, req *types.TweetsRequest) (*types.TopTweetsResponse, error) {
	if m.getTopTweetsFn != nil {
		return m.getTopTweetsFn(ctx, req)
	}
	return &types.TopTweetsResponse{TopTweets: []types.Tweet{}}, nil
}

func (m *mockService) GetLeastTweets(ctx context.Context, req *types.TweetsRequest) (*types.LeastTweetsResponse, error) {
	if m.getLeastTweetsFn != nil {
		return m.getLeastTweetsFn(ctx, req)
	}
	return &types.LeastTweetsResponse{LeastTweets: []types.Tweet{}}, nil
}

func (m *mockService) GetCreditsUsedCount(ctx context.Context, req *types.TwitterRequest) (*types.CreditsUsedResponse, error) {
	if m.getCreditsUsedCountFn != nil {
		return m.getCreditsUsedCountFn(ctx, req)
	}
	return &types.CreditsUsedResponse{Data: types.CreditsUsedData{CreditsUsed: 10}}, nil
}

func newTestHandler() *Handler {
	return NewHandler(&mockService{}, zerolog.New(io.Discard))
}

const validQueryStr = "workspace_id=ws1&twitter_id=tw_123&start_date=2025-01-01&end_date=2025-01-31&timezone=UTC"

func TestParseBaseRequest(t *testing.T) {
	tests := []struct {
		name      string
		query     string
		expectErr bool
	}{
		{name: "valid", query: validQueryStr},
		{name: "valid with date range", query: "workspace_id=ws1&twitter_id=tw_123&date=2025-01-01+-+2025-01-31&timezone=UTC"},
		{name: "missing workspace", query: "twitter_id=tw_123&start_date=2025-01-01&end_date=2025-01-31", expectErr: true},
		{name: "missing twitter id", query: "workspace_id=ws1&start_date=2025-01-01&end_date=2025-01-31", expectErr: true},
		{name: "missing start", query: "workspace_id=ws1&twitter_id=tw_123&end_date=2025-01-31", expectErr: true},
		{name: "invalid timezone", query: "workspace_id=ws1&twitter_id=tw_123&start_date=2025-01-01&end_date=2025-01-31&timezone=Invalid/Zone", expectErr: true},
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

func TestParseTweetsRequest(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/x?"+validQueryStr+"&limit=abc", nil)
	_, err := parseTweetsRequest(req)
	if err == nil {
		t.Fatal("expected invalid limit error")
	}
}

func TestHandlePageAndPostsInsights(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/twitter/getPageAndPostsInsights?"+validQueryStr, nil)
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
	if data["twitter_id"] != "tw_123" {
		t.Fatalf("expected twitter_id tw_123, got %v", data["twitter_id"])
	}
}

func TestHandleTopTweets_BadRequest(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/twitter/getTopTweets?"+validQueryStr+"&limit=abc", nil)
	w := httptest.NewRecorder()

	h.HandleTopTweets(w, req)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", w.Code)
	}
}

func TestHandleCreditsUsedCount(t *testing.T) {
	h := newTestHandler()
	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/twitter/getCreditsUsedCount?"+validQueryStr, nil)
	w := httptest.NewRecorder()

	h.HandleCreditsUsedCount(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}
}
