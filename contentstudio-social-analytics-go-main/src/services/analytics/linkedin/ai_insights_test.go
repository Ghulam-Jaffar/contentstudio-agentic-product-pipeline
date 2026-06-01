package linkedin

import (
	"context"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/linkedin"
	"testing"
	"time"
)

type mockCache struct {
	store map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string]string)}
}

func (m *mockCache) Get(_ context.Context, key string) (string, error) {
	return m.store[key], nil
}

func (m *mockCache) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	m.store[key] = value.(string)
	return nil
}

func (m *mockCache) SetNX(_ context.Context, key string, value interface{}, _ time.Duration) (bool, error) {
	if _, ok := m.store[key]; ok {
		return false, nil
	}
	m.store[key] = value.(string)
	return true, nil
}

func (m *mockCache) Del(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(m.store, k)
	}
	return nil
}

func (m *mockCache) CompareAndDelete(_ context.Context, key, expected string) (bool, error) {
	if val, ok := m.store[key]; ok && val == expected {
		delete(m.store, key)
		return true, nil
	}
	return false, nil
}

func (m *mockCache) DecrBy(_ context.Context, _ string, _ int64) (int64, error) { return 0, nil }
func (m *mockCache) DecrByIfPositive(_ context.Context, _ string, _ int64) (int64, bool, error) {
	return 0, false, nil
}
func (m *mockCache) IncrBy(_ context.Context, _ string, _ int64) (int64, error) { return 0, nil }
func (m *mockCache) Expire(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}
func (m *mockCache) Close() error { return nil }

type fakeAgent struct {
	requestFn func(context.Context, string, map[string]interface{}) (map[string]interface{}, error)
}

func (f *fakeAgent) Request(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	return f.requestFn(ctx, endpoint, payload)
}

type mockAnalyticsService struct {
	getSummaryFn               func(context.Context, *types.LinkedInRequest) (*types.SummaryResponse, error)
	getAudienceGrowthFn        func(context.Context, *types.LinkedInRequest) (*types.AudienceGrowthResponse, error)
	getPageViewsFn             func(context.Context, *types.LinkedInRequest) (*types.PageViewsResponse, error)
	getPublishingBehaviourFn   func(context.Context, *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error)
	getTopPostsFn              func(context.Context, *types.TopPostsRequest) (*types.TopPostsResponse, error)
	getPostsPerDayFn           func(context.Context, *types.LinkedInRequest) (*types.PostsPerDayResponse, error)
	getHashtagsFn              func(context.Context, *types.LinkedInRequest) (*types.HashtagsResponse, error)
	getFollowersDemographicsFn func(context.Context, *types.LinkedInRequest) (*types.DemographicsResponse, error)
}

var _ Service = (*mockAnalyticsService)(nil)

func (m *mockAnalyticsService) GetSummary(ctx context.Context, req *types.LinkedInRequest) (*types.SummaryResponse, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, req)
	}
	return &types.SummaryResponse{Status: true, Overview: map[string]*types.SummaryMetrics{"current": {}, "previous": {}}}, nil
}

func (m *mockAnalyticsService) GetAudienceGrowth(ctx context.Context, req *types.LinkedInRequest) (*types.AudienceGrowthResponse, error) {
	if m.getAudienceGrowthFn != nil {
		return m.getAudienceGrowthFn(ctx, req)
	}
	return &types.AudienceGrowthResponse{Status: true, AudienceGrowth: &types.AudienceGrowthData{}}, nil
}

func (m *mockAnalyticsService) GetPageViews(ctx context.Context, req *types.LinkedInRequest) (*types.PageViewsResponse, error) {
	if m.getPageViewsFn != nil {
		return m.getPageViewsFn(ctx, req)
	}
	return &types.PageViewsResponse{Status: true, PageViews: &types.PageViewsData{}}, nil
}

func (m *mockAnalyticsService) GetPublishingBehaviour(ctx context.Context, req *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error) {
	if m.getPublishingBehaviourFn != nil {
		return m.getPublishingBehaviourFn(ctx, req)
	}
	return &types.PublishingBehaviourResponse{Status: true, PublishingBehaviour: &types.PublishingBehaviourData{}}, nil
}

func (m *mockAnalyticsService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
	if m.getTopPostsFn != nil {
		return m.getTopPostsFn(ctx, req)
	}
	return &types.TopPostsResponse{Status: true, TopPosts: []types.TopPost{}}, nil
}

func (m *mockAnalyticsService) GetPostsPerDay(ctx context.Context, req *types.LinkedInRequest) (*types.PostsPerDayResponse, error) {
	if m.getPostsPerDayFn != nil {
		return m.getPostsPerDayFn(ctx, req)
	}
	return &types.PostsPerDayResponse{Status: true, PostsPerDays: &types.PostsPerDayData{}}, nil
}

func (m *mockAnalyticsService) GetHashtags(ctx context.Context, req *types.LinkedInRequest) (*types.HashtagsResponse, error) {
	if m.getHashtagsFn != nil {
		return m.getHashtagsFn(ctx, req)
	}
	return &types.HashtagsResponse{Status: true, TopHashtags: &types.HashtagsData{}}, nil
}

func (m *mockAnalyticsService) GetFollowersDemographics(ctx context.Context, req *types.LinkedInRequest) (*types.DemographicsResponse, error) {
	if m.getFollowersDemographicsFn != nil {
		return m.getFollowersDemographicsFn(ctx, req)
	}
	return &types.DemographicsResponse{Status: true, FollowerDemographics: map[string]*types.DemographicCategory{}}, nil
}

func newTestAIService(t *testing.T, analytics Service, agent agentRequester) *AIInsightsService {
	t.Helper()
	return NewAIInsightsService(analytics, agent, newMockCache())
}

func TestGetAIInsights_PageViews(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{
		getPageViewsFn: func(context.Context, *types.LinkedInRequest) (*types.PageViewsResponse, error) {
			return &types.PageViewsResponse{
				Status: true,
				PageViews: &types.PageViewsData{
					TotalPageViews: []int32{12, 18},
					ShowData:       1,
					Buckets:        []string{"2025-01-01", "2025-01-02"},
				},
			}, nil
		},
	}, &fakeAgent{requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
		if endpoint != "linkedin/page-views" {
			t.Fatalf("unexpected endpoint %q", endpoint)
		}
		if payload["language"] != "en" {
			t.Fatalf("expected default language 'en', got %v", payload["language"])
		}
		dataset := payload["dataset"].(map[string]interface{})
		if _, ok := dataset["total_page_views"]; !ok {
			t.Fatal("expected total_page_views in dataset")
		}
		return map[string]interface{}{"insights": map[string]interface{}{"summary": "ok"}}, nil
	}})

	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		LinkedinID:  "li_1",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "page_views",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true, got %v", resp["success"])
	}
}

func TestGetAIInsights_InsufficientData(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{
		getPageViewsFn: func(context.Context, *types.LinkedInRequest) (*types.PageViewsResponse, error) {
			return &types.PageViewsResponse{
				Status: true,
				PageViews: &types.PageViewsData{
					TotalPageViews: []int32{0, 0},
					ShowData:       0,
					Buckets:        []string{"2025-01-01", "2025-01-02"},
				},
			}, nil
		},
	}, &fakeAgent{requestFn: func(context.Context, string, map[string]interface{}) (map[string]interface{}, error) {
		t.Fatal("AI agent should not be called for insufficient data")
		return nil, nil
	}})

	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		LinkedinID:  "li_1",
		Date:        []interface{}{"2025-01-01", "2025-01-31"},
		Type:        "page_views",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != false {
		t.Fatalf("expected success=false, got %v", resp["success"])
	}
	if resp["message"] != "insufficient data" {
		t.Fatalf("unexpected message: %v", resp["message"])
	}
}
