package instagram

import (
	"context"
	"testing"
	"time"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/instagram"
)

// --- Mock cache ---

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

// --- Mock agent ---

type fakeAgent struct {
	requestFn func(context.Context, string, map[string]interface{}) (map[string]interface{}, error)
}

func (f *fakeAgent) Request(ctx context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	return f.requestFn(ctx, endpoint, payload)
}

// --- Mock analytics service ---

type mockAnalyticsService struct {
	getSummaryFn             func(context.Context, *types.InstagramRequest) (*types.SummaryResponse, error)
	getAudienceGrowthFn      func(context.Context, *types.InstagramRequest) (*types.AudienceGrowthResponse, error)
	getPublishingBehaviourFn func(context.Context, *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error)
	getTopPostsFn            func(context.Context, *types.TopPostsRequest) (*types.TopPostsResponse, error)
	getActiveUsersFn         func(context.Context, *types.InstagramRequest) (*types.ActiveUsersResponse, error)
	getImpressionsFn         func(context.Context, *types.InstagramRequest) (*types.ImpressionsResponse, error)
	getEngagementFn          func(context.Context, *types.InstagramRequest) (*types.EngagementResponse, error)
	getHashtagsFn            func(context.Context, *types.InstagramRequest) (*types.HashtagsResponse, error)
	getStoriesPerformanceFn  func(context.Context, *types.InstagramRequest) (*types.StoriesPerformanceResponse, error)
	getReelsPerformanceFn    func(context.Context, *types.InstagramRequest) (*types.ReelsPerformanceResponse, error)
	getDemographicsAgeFn     func(context.Context, *types.InstagramRequest) (*types.DemographicsAgeResponse, error)
	getCountryCityFn         func(context.Context, *types.InstagramRequest) (*types.CountryCityResponse, error)
}

var _ Service = (*mockAnalyticsService)(nil)

func (m *mockAnalyticsService) GetSummary(ctx context.Context, req *types.InstagramRequest) (*types.SummaryResponse, error) {
	if m.getSummaryFn != nil {
		return m.getSummaryFn(ctx, req)
	}
	return &types.SummaryResponse{Status: true, Overview: map[string]*types.SummaryMetrics{"current": {TotalPosts: 10}, "previous": {}}}, nil
}

func (m *mockAnalyticsService) GetAudienceGrowth(ctx context.Context, req *types.InstagramRequest) (*types.AudienceGrowthResponse, error) {
	if m.getAudienceGrowthFn != nil {
		return m.getAudienceGrowthFn(ctx, req)
	}
	return &types.AudienceGrowthResponse{
		Status: true,
		AudienceGrowth: &types.AudienceGrowthData{
			ShowData:  1,
			Followers: []int32{1000, 1010},
			Buckets:   []string{"2025-01-01", "2025-01-02"},
		},
	}, nil
}

func (m *mockAnalyticsService) GetPublishingBehaviour(ctx context.Context, req *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error) {
	if m.getPublishingBehaviourFn != nil {
		return m.getPublishingBehaviourFn(ctx, req)
	}
	return &types.PublishingBehaviourResponse{
		Status: true,
		PublishingBehaviour: &types.PublishingBehaviourData{
			TotalPosts: []int32{5, 3},
			Buckets:    []string{"2025-01-01", "2025-01-02"},
		},
	}, nil
}

func (m *mockAnalyticsService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
	if m.getTopPostsFn != nil {
		return m.getTopPostsFn(ctx, req)
	}
	return &types.TopPostsResponse{
		Status:   true,
		TopPosts: []types.TopPost{{MediaID: "media_1", LikeCount: 100}},
	}, nil
}

func (m *mockAnalyticsService) GetActiveUsers(ctx context.Context, req *types.InstagramRequest) (*types.ActiveUsersResponse, error) {
	if m.getActiveUsersFn != nil {
		return m.getActiveUsersFn(ctx, req)
	}
	return &types.ActiveUsersResponse{Status: true, ActiveUsersHours: &types.ActiveUsersHours{}, ActiveUsersDays: &types.ActiveUsersDays{}}, nil
}

func (m *mockAnalyticsService) GetImpressions(ctx context.Context, req *types.InstagramRequest) (*types.ImpressionsResponse, error) {
	if m.getImpressionsFn != nil {
		return m.getImpressionsFn(ctx, req)
	}
	return &types.ImpressionsResponse{
		Status: true,
		Impressions: &types.ImpressionsData{
			ShowData:    1,
			Impressions: []int32{500, 600},
			Buckets:     []string{"2025-01-01", "2025-01-02"},
		},
	}, nil
}

func (m *mockAnalyticsService) GetEngagement(ctx context.Context, req *types.InstagramRequest) (*types.EngagementResponse, error) {
	if m.getEngagementFn != nil {
		return m.getEngagementFn(ctx, req)
	}
	return &types.EngagementResponse{
		Status: true,
		Engagements: &types.EngagementData{
			ShowData:   1,
			Engagement: []int32{100, 200},
			Buckets:    []string{"2025-01-01", "2025-01-02"},
		},
	}, nil
}

func (m *mockAnalyticsService) GetHashtags(ctx context.Context, req *types.InstagramRequest) (*types.HashtagsResponse, error) {
	if m.getHashtagsFn != nil {
		return m.getHashtagsFn(ctx, req)
	}
	return &types.HashtagsResponse{
		Status: true,
		TopHashtags: &types.HashtagsData{
			Name:  []string{"#go"},
			Posts: []int32{10},
		},
	}, nil
}

func (m *mockAnalyticsService) GetStoriesPerformance(ctx context.Context, req *types.InstagramRequest) (*types.StoriesPerformanceResponse, error) {
	if m.getStoriesPerformanceFn != nil {
		return m.getStoriesPerformanceFn(ctx, req)
	}
	return &types.StoriesPerformanceResponse{
		Status: true,
		StoriesPerformance: &types.StoriesData{
			ShowData: 1,
			Buckets:  []string{"2025-01-01"},
		},
	}, nil
}

func (m *mockAnalyticsService) GetReelsPerformance(ctx context.Context, req *types.InstagramRequest) (*types.ReelsPerformanceResponse, error) {
	if m.getReelsPerformanceFn != nil {
		return m.getReelsPerformanceFn(ctx, req)
	}
	return &types.ReelsPerformanceResponse{
		Status: true,
		Reels: &types.ReelsData{
			ShowData:   1,
			TotalPosts: []int32{5},
			Buckets:    []string{"2025-01-01"},
		},
	}, nil
}

func (m *mockAnalyticsService) GetDemographicsAge(ctx context.Context, req *types.InstagramRequest) (*types.DemographicsAgeResponse, error) {
	if m.getDemographicsAgeFn != nil {
		return m.getDemographicsAgeFn(ctx, req)
	}
	return &types.DemographicsAgeResponse{AudienceAge: map[string]int64{}, AudienceGender: map[string]int64{}}, nil
}

func (m *mockAnalyticsService) GetCountryCity(ctx context.Context, req *types.InstagramRequest) (*types.CountryCityResponse, error) {
	if m.getCountryCityFn != nil {
		return m.getCountryCityFn(ctx, req)
	}
	return &types.CountryCityResponse{AudienceCity: map[string]int64{}, AudienceCountry: map[string]int64{}}, nil
}

func newTestAIService(t *testing.T, analytics Service, agent agentRequester) *AIInsightsService {
	t.Helper()
	return NewAIInsightsService(analytics, agent, newMockCache())
}

func defaultAgent() *fakeAgent {
	return &fakeAgent{requestFn: func(_ context.Context, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"insights": "ok"}, nil
	}}
}

// --- Invalid insight type ---

func TestGetAIInsights_InvalidType(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, defaultAgent())
	_, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "not_a_real_type",
	})
	if err == nil {
		t.Fatal("expected error for invalid type")
	}
}

// --- Impressions ---

func TestGetAIInsights_Impressions(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/impressions" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			if payload["language"] != "en" {
				t.Fatalf("expected language=en, got %v", payload["language"])
			}
			dataset, ok := payload["dataset"].(map[string]interface{})
			if !ok {
				t.Fatal("expected dataset in payload")
			}
			if _, ok := dataset["impressions"]; !ok {
				t.Fatal("expected impressions in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})

	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "impressions",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true, got %v", resp["success"])
	}
}

// --- Impressions insufficient data ---

func TestGetAIInsights_Impressions_InsufficientData(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{
		getImpressionsFn: func(_ context.Context, _ *types.InstagramRequest) (*types.ImpressionsResponse, error) {
			return &types.ImpressionsResponse{
				Status: true,
				Impressions: &types.ImpressionsData{
					ShowData: 0,
				},
			}, nil
		},
	}, &fakeAgent{requestFn: func(_ context.Context, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
		t.Fatal("agent should not be called for insufficient data")
		return nil, nil
	}})

	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "impressions",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != false {
		t.Fatalf("expected success=false, got %v", resp["success"])
	}
	if resp["message"] != "insufficient data" {
		t.Fatalf("expected message='insufficient data', got %v", resp["message"])
	}
}

// --- Engagement ---

func TestGetAIInsights_Engagement(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, _ map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/engagement" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "engagement",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true, got %v", resp["success"])
	}
}

// --- Publishing behaviour types ---

func TestGetAIInsights_PublishingBehaviourImpressions(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/publishing-impressions" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["impressions"]; !ok {
				t.Fatal("expected impressions in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "publishing_behaviour_impressions",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true, got %v", resp["success"])
	}
}

func TestGetAIInsights_PublishingBehaviourEngagements(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/publishing-engagement" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["engagement"]; !ok {
				t.Fatal("expected engagement in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "publishing_behaviour_engagements",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

func TestGetAIInsights_PublishingBehaviourReach(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/publishing-reach" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["reach"]; !ok {
				t.Fatal("expected reach in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "publishing_behaviour_reach",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

// --- Audience growth ---

func TestGetAIInsights_AudienceGrowth(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/audience-growth" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["followers"]; !ok {
				t.Fatal("expected followers in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "audience_growth",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true, got %v", resp["success"])
	}
}

// --- Reels ---

func TestGetAIInsights_ReelsEngagement(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/reels-engagement" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["engagement"]; !ok {
				t.Fatal("expected engagement in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "reels_engagement",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

func TestGetAIInsights_ReelsWatchTime(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/reels-watch-time" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["avg_watch_time"]; !ok {
				t.Fatal("expected avg_watch_time in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "reels_watch_time",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

func TestGetAIInsights_ReelsShares(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/reels-shares" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["shares"]; !ok {
				t.Fatal("expected shares in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "reels_shares",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

// --- Stories ---

func TestGetAIInsights_StoriesInteractions(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/stories-interactions" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["story_reply"]; !ok {
				t.Fatal("expected story_reply in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "stories_interactions",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

func TestGetAIInsights_StoriesImpressions(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/stories-impressions" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["story_impressions"]; !ok {
				t.Fatal("expected story_impressions in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "stories_impressions",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

func TestGetAIInsights_StoriesReach(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/stories-reach" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["story_reach"]; !ok {
				t.Fatal("expected story_reach in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "stories_reach",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

// --- Top posts ---

func TestGetAIInsights_TopPosts(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/top-posts" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["top_posts"]; !ok {
				t.Fatal("expected top_posts in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "top_posts",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true, got %v", resp["success"])
	}
}

func TestGetAIInsights_TopPosts_InsufficientData(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{
		getTopPostsFn: func(_ context.Context, _ *types.TopPostsRequest) (*types.TopPostsResponse, error) {
			return &types.TopPostsResponse{Status: true, TopPosts: []types.TopPost{}}, nil
		},
	}, &fakeAgent{requestFn: func(_ context.Context, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
		t.Fatal("agent should not be called for insufficient data")
		return nil, nil
	}})

	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "top_posts",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != false {
		t.Fatalf("expected success=false for empty posts")
	}
}

// --- Top hashtags ---

func TestGetAIInsights_TopHashtags(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/hashtags" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["top_hashtags"]; !ok {
				t.Fatal("expected top_hashtags in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "top_hashtags",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

// --- Insights summary ---

func TestGetAIInsights_InsightsSummary(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
			if endpoint != "instagram/insights-summary" {
				t.Fatalf("unexpected endpoint %q", endpoint)
			}
			dataset := payload["dataset"].(map[string]interface{})
			if _, ok := dataset["account_data"]; !ok {
				t.Fatal("expected account_data in dataset")
			}
			if _, ok := dataset["top_posts"]; !ok {
				t.Fatal("expected top_posts in dataset")
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "insights_summary",
		Limit:       10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true, got %v", resp["success"])
	}
}

func TestGetAIInsights_InsightsSummary_NoTopPosts(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{
		getTopPostsFn: func(_ context.Context, _ *types.TopPostsRequest) (*types.TopPostsResponse, error) {
			return &types.TopPostsResponse{Status: true, TopPosts: []types.TopPost{}}, nil
		},
	}, &fakeAgent{requestFn: func(_ context.Context, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
		t.Fatal("agent should not be called when no top posts")
		return nil, nil
	}})

	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "insights_summary",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != false {
		t.Fatalf("expected success=false when no top posts")
	}
}

// --- Cache ---

func TestGetAIInsights_CacheHit(t *testing.T) {
	cache := newMockCache()
	cache.store["ig_AI:impressions:ig_123:2025-01-01,2025-01-31:en"] = `{"cached_result":"yes"}`

	agentCalled := false
	svc := NewAIInsightsService(&mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, _ string, _ map[string]interface{}) (map[string]interface{}, error) {
			agentCalled = true
			return map[string]interface{}{"insights": "ok"}, nil
		},
	}, cache)

	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "impressions",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agentCalled {
		t.Fatal("agent should not be called on cache hit")
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true from cache, got %v", resp["success"])
	}
	if resp["cached_result"] != "yes" {
		t.Fatalf("expected cached_result=yes, got %v", resp["cached_result"])
	}
}

// --- Language / date format variations ---

func TestGetAIInsights_CustomLanguage(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, &fakeAgent{
		requestFn: func(_ context.Context, _ string, payload map[string]interface{}) (map[string]interface{}, error) {
			if payload["language"] != "fr" {
				t.Fatalf("expected language=fr, got %v", payload["language"])
			}
			return map[string]interface{}{"insights": "ok"}, nil
		},
	})
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "impressions",
		Language:    "fr",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

func TestGetAIInsights_DateAsSlice(t *testing.T) {
	svc := newTestAIService(t, &mockAnalyticsService{}, defaultAgent())
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		InstagramID: "ig_123",
		Date:        []interface{}{"2025-01-01", "2025-01-31"},
		Type:        "impressions",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp["success"] != true {
		t.Fatalf("expected success=true")
	}
}

// --- aggregationLevel ---

func TestAggregationLevel_Daily(t *testing.T) {
	level := aggregationLevel("2025-01-01", "2025-01-31")
	if level != "daily" {
		t.Fatalf("expected daily, got %q", level)
	}
}

func TestAggregationLevel_Monthly(t *testing.T) {
	level := aggregationLevel("2025-01-01", "2025-03-15")
	if level != "monthly" {
		t.Fatalf("expected monthly for >60 days, got %q", level)
	}
}

func TestAggregationLevel_InvalidDate(t *testing.T) {
	level := aggregationLevel("not-a-date", "2025-01-31")
	if level != "daily" {
		t.Fatalf("expected daily fallback for invalid date, got %q", level)
	}
}

// --- parseDateRange ---

func TestParseDateRange_StringFormat(t *testing.T) {
	start, end, err := parseDateRange("2025-01-01 - 2025-01-31")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != "2025-01-01" {
		t.Fatalf("expected start=2025-01-01, got %q", start)
	}
	if end != "2025-01-31" {
		t.Fatalf("expected end=2025-01-31, got %q", end)
	}
}

func TestParseDateRange_SliceFormat(t *testing.T) {
	start, end, err := parseDateRange([]interface{}{"2025-01-01", "2025-01-31"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != "2025-01-01" {
		t.Fatalf("expected start=2025-01-01, got %q", start)
	}
	if end != "2025-01-31" {
		t.Fatalf("expected end=2025-01-31, got %q", end)
	}
}

func TestParseDateRange_StringSliceFormat(t *testing.T) {
	start, end, err := parseDateRange([]string{"2025-06-01", "2025-06-30"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if start != "2025-06-01" {
		t.Fatalf("expected start=2025-06-01, got %q", start)
	}
	if end != "2025-06-30" {
		t.Fatalf("expected end=2025-06-30, got %q", end)
	}
}

func TestParseDateRange_InvalidFormat(t *testing.T) {
	_, _, err := parseDateRange(12345)
	if err == nil {
		t.Fatal("expected error for unsupported date format")
	}
}
