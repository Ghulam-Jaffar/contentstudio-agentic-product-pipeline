package tiktok

import (
	"context"
	"testing"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/tiktok"
)

type mockAgent struct {
	lastEndpoint string
	lastPayload  map[string]interface{}
}

func (m *mockAgent) Request(_ context.Context, endpoint string, payload map[string]interface{}) (map[string]interface{}, error) {
	m.lastEndpoint = endpoint
	m.lastPayload = payload
	return map[string]interface{}{"insights": []string{"ok"}}, nil
}

type mockAnalyticsService struct {
	dynamicFollowers map[string]interface{}
	topPosts         map[string]interface{}
	postsEngagements map[string]interface{}
	followersViews   map[string]interface{}
	dailyEngagements map[string]interface{}
}

func (m *mockAnalyticsService) GetPageAndPostsInsights(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}
func (m *mockAnalyticsService) GetPageFollowersAndViews(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return m.followersViews, nil
}
func (m *mockAnalyticsService) GetDynamicPageFollowersAndViews(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return m.dynamicFollowers, nil
}
func (m *mockAnalyticsService) GetPostsAndEngagements(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return m.postsEngagements, nil
}
func (m *mockAnalyticsService) GetDailyEngagementsData(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return m.dailyEngagements, nil
}
func (m *mockAnalyticsService) GetDynamicDailyEngagementsData(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return m.dailyEngagements, nil
}
func (m *mockAnalyticsService) GetTopAndLeastPerformingPosts(_ context.Context, _ *types.TiktokRequest) (map[string]interface{}, error) {
	return m.topPosts, nil
}
func (m *mockAnalyticsService) GetPostsData(_ context.Context, _ *types.PostsRequest) (map[string]interface{}, error) {
	return map[string]interface{}{}, nil
}

func TestAIInsights_InvalidType(t *testing.T) {
	svc := NewAIInsightsService(&mockAnalyticsService{}, &mockAgent{}, nil)
	_, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		TiktokID:    "tt_1",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "unknown",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestAIInsights_InsufficientData(t *testing.T) {
	analytics := &mockAnalyticsService{
		dynamicFollowers: map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{"followers_count_diff": []interface{}{0.0, 0.0}},
			},
		},
	}
	svc := NewAIInsightsService(analytics, &mockAgent{}, nil)
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		TiktokID:    "tt_1",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "audience_growth",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok, _ := resp["success"].(bool); ok {
		t.Fatalf("expected success=false, got %v", resp["success"])
	}
}

func TestAIInsights_SummarySuccess(t *testing.T) {
	agent := &mockAgent{}
	analytics := &mockAnalyticsService{
		postsEngagements: map[string]interface{}{"data": []interface{}{map[string]interface{}{"post_count": []interface{}{1.0}}}},
		followersViews:   map[string]interface{}{"data": []interface{}{map[string]interface{}{"followers_count": []interface{}{10.0}}}},
		topPosts: map[string]interface{}{
			"data": map[string]interface{}{
				"top_posts":   []interface{}{map[string]interface{}{"post_id": "p1"}},
				"least_posts": []interface{}{},
			},
		},
		dailyEngagements: map[string]interface{}{"data": []interface{}{map[string]interface{}{"daily_engagement": []interface{}{3.0}}}},
		dynamicFollowers: map[string]interface{}{"data": []interface{}{map[string]interface{}{"followers_count_diff": []interface{}{2.0}}}},
	}
	svc := NewAIInsightsService(analytics, agent, nil)
	resp, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		TiktokID:    "tt_1",
		Date:        "2025-01-01 - 2025-01-31",
		Type:        "insights_summary",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if agent.lastEndpoint != "tiktok/insights-overview" {
		t.Fatalf("unexpected endpoint %q", agent.lastEndpoint)
	}
	if ok, _ := resp["success"].(bool); !ok {
		t.Fatalf("expected success=true, got %v", resp["success"])
	}
}
