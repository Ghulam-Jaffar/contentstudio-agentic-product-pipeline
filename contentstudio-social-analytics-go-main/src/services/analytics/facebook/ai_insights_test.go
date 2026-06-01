package facebook

import (
	"context"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/facebook"
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
	getSummaryFn              func(context.Context, *types.FacebookRequest) (*types.SummaryResponse, error)
	getAudienceGrowthFn       func(context.Context, *types.FacebookRequest) (*types.AudienceGrowthResponse, error)
	getPublishingBehaviourFn  func(context.Context, *types.FacebookRequest) (*types.PublishingBehaviourResponse, error)
	getTopPostsFn             func(context.Context, *types.FacebookRequest) (*types.TopPostsResponse, error)
	getActiveUsersFn          func(context.Context, *types.FacebookRequest) (*types.ActiveUsersResponse, error)
	getImpressionsFn          func(context.Context, *types.FacebookRequest) (*types.ImpressionsResponse, error)
	getEngagementFn           func(context.Context, *types.FacebookRequest) (*types.EngagementResponse, error)
	getReelsAnalyticsFn       func(context.Context, *types.FacebookRequest) (*types.ReelsAnalyticsResponse, error)
	getVideoInsightsFn        func(context.Context, *types.FacebookRequest) (*types.VideoInsightsResponse, error)
	getDemographicsFn         func(context.Context, *types.FacebookRequest) (*types.DemographicsResponse, error)
	getOverviewDemographicsFn func(context.Context, *types.FacebookRequest) (*types.DemographicsResponse, error)
	getAudienceLocationFn     func(context.Context, *types.FacebookRequest) (*types.DemographicsResponse, error)
}

var _ Service = (*mockAnalyticsService)(nil)

func (m *mockAnalyticsService) GetSummary(ctx context.Context, req *types.FacebookRequest) (*types.SummaryResponse, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, req)
	}
	return &types.SummaryResponse{Status: true, Overview: map[string]*types.SummaryMetrics{"current": {}, "previous": {}}}, nil
}

func (m *mockAnalyticsService) GetAudienceGrowth(ctx context.Context, req *types.FacebookRequest) (*types.AudienceGrowthResponse, error) {
	if m.getAudienceGrowthFn != nil {
		return m.getAudienceGrowthFn(ctx, req)
	}
	return &types.AudienceGrowthResponse{Status: true, AudienceGrowth: &types.AudienceGrowthData{}}, nil
}

func (m *mockAnalyticsService) GetPublishingBehaviour(ctx context.Context, req *types.FacebookRequest) (*types.PublishingBehaviourResponse, error) {
	if m.getPublishingBehaviourFn != nil {
		return m.getPublishingBehaviourFn(ctx, req)
	}
	return &types.PublishingBehaviourResponse{Status: true, PublishingBehaviour: &types.PublishingBehaviourData{}}, nil
}

func (m *mockAnalyticsService) GetTopPosts(ctx context.Context, req *types.FacebookRequest) (*types.TopPostsResponse, error) {
	if m.getTopPostsFn != nil {
		return m.getTopPostsFn(ctx, req)
	}
	return &types.TopPostsResponse{Status: true, TopPosts: []types.TopPost{}}, nil
}

func (m *mockAnalyticsService) GetActiveUsers(ctx context.Context, req *types.FacebookRequest) (*types.ActiveUsersResponse, error) {
	if m.getActiveUsersFn != nil {
		return m.getActiveUsersFn(ctx, req)
	}
	return &types.ActiveUsersResponse{Status: true, ActiveUsers: &types.ActiveUsersData{}}, nil
}

func (m *mockAnalyticsService) GetImpressions(ctx context.Context, req *types.FacebookRequest) (*types.ImpressionsResponse, error) {
	if m.getImpressionsFn != nil {
		return m.getImpressionsFn(ctx, req)
	}
	return &types.ImpressionsResponse{Status: true, Impressions: &types.ImpressionsData{}}, nil
}

func (m *mockAnalyticsService) GetEngagement(ctx context.Context, req *types.FacebookRequest) (*types.EngagementResponse, error) {
	if m.getEngagementFn != nil {
		return m.getEngagementFn(ctx, req)
	}
	return &types.EngagementResponse{Status: true, Engagement: &types.EngagementContainer{Engagement: &types.EngagementData{}}}, nil
}

func (m *mockAnalyticsService) GetReelsAnalytics(ctx context.Context, req *types.FacebookRequest) (*types.ReelsAnalyticsResponse, error) {
	if m.getReelsAnalyticsFn != nil {
		return m.getReelsAnalyticsFn(ctx, req)
	}
	return &types.ReelsAnalyticsResponse{Status: true, Reels: &types.ReelsData{}}, nil
}

func (m *mockAnalyticsService) GetVideoInsights(ctx context.Context, req *types.FacebookRequest) (*types.VideoInsightsResponse, error) {
	if m.getVideoInsightsFn != nil {
		return m.getVideoInsightsFn(ctx, req)
	}
	return &types.VideoInsightsResponse{Status: true, VideoInsights: &types.VideoInsightsData{}}, nil
}

func (m *mockAnalyticsService) GetDemographics(ctx context.Context, req *types.FacebookRequest) (*types.DemographicsResponse, error) {
	if m.getDemographicsFn != nil {
		return m.getDemographicsFn(ctx, req)
	}
	return &types.DemographicsResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetOverviewDemographics(ctx context.Context, req *types.FacebookRequest) (*types.DemographicsResponse, error) {
	if m.getOverviewDemographicsFn != nil {
		return m.getOverviewDemographicsFn(ctx, req)
	}
	return &types.DemographicsResponse{Status: true}, nil
}

func (m *mockAnalyticsService) GetAudienceLocation(ctx context.Context, req *types.FacebookRequest) (*types.DemographicsResponse, error) {
	if m.getAudienceLocationFn != nil {
		return m.getAudienceLocationFn(ctx, req)
	}
	return &types.DemographicsResponse{Status: true}, nil
}

func newTestAIService(t *testing.T, analytics Service, agent agentRequester) *AIInsightsService {
	t.Helper()
	return NewAIInsightsService(analytics, agent, newMockCache())
}

func TestGetAIInsights_PageImpressions(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{
		getImpressionsFn: func(context.Context, *types.FacebookRequest) (*types.ImpressionsResponse, error) {
			return &types.ImpressionsResponse{
				Status: true,
				Impressions: &types.ImpressionsData{
					PageImpressions: []int32{10, 20},
					Buckets:         []string{"2025-01-01", "2025-01-02"},
				},
			}, nil
		},
	}, &fakeAgent{requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
		if endpoint != "facebook/page-impressions" {
			t.Fatalf("unexpected endpoint %q", endpoint)
		}
		if payload["language"] != "en" {
			t.Fatalf("expected default language 'en', got %v", payload["language"])
		}
		dataset := payload["dataset"].(map[string]interface{})
		if _, ok := dataset["page_impressions"]; !ok {
			t.Fatal("expected page_impressions in dataset")
		}
		return map[string]interface{}{"insights": map[string]interface{}{"summary": "ok"}}, nil
	}})

	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		FacebookID:  "fb_1",
		Date:        []interface{}{"2025-01-01", "2025-01-31"},
		Type:        "page_impressions",
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
		getImpressionsFn: func(context.Context, *types.FacebookRequest) (*types.ImpressionsResponse, error) {
			return &types.ImpressionsResponse{
				Status: true,
				Impressions: &types.ImpressionsData{
					PageImpressions: []int32{0, 0},
					Buckets:         []string{"2025-01-01", "2025-01-02"},
				},
			}, nil
		},
	}, &fakeAgent{requestFn: func(context.Context, string, map[string]interface{}) (map[string]interface{}, error) {
		t.Fatal("AI agent should not be called for insufficient data")
		return nil, nil
	}})

	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		FacebookID:  "fb_1",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "page_impressions",
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
