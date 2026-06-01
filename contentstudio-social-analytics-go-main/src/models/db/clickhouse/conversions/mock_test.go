package conversions

import (
	"context"
	"errors"
	"testing"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestMockClickHouseClient_BulkInsertPosts(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertPosts(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	called := false
	mock.BulkInsertPostsFunc = func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
		called = true
		return nil
	}
	err = mock.BulkInsertPosts(context.Background(), []*clickhousemodels.FacebookPosts{{PostID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected function to be called")
	}

	// Test with error
	mock.BulkInsertPostsFunc = func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertPosts(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_BulkInsertMediaAssets(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertMediaAssets(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertMediaAssetsFunc = func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
		return nil
	}
	err = mock.BulkInsertMediaAssets(context.Background(), []*clickhousemodels.FacebookMediaAssets{{MediaID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertMediaAssetsFunc = func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertMediaAssets(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_BulkInsertVideoInsights(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertVideoInsights(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertVideoInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
		return nil
	}
	err = mock.BulkInsertVideoInsights(context.Background(), []*clickhousemodels.FacebookVideoInsights{{VideoID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertVideoInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertVideoInsights(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_BulkInsertReelsInsights(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertReelsInsights(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertReelsInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
		return nil
	}
	err = mock.BulkInsertReelsInsights(context.Background(), []*clickhousemodels.FacebookReelsInsights{{PostID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertReelsInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertReelsInsights(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_BulkInsertInsights(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertInsights(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
		return nil
	}
	err = mock.BulkInsertInsights(context.Background(), []*clickhousemodels.FacebookInsights{{PageID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertInsights(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_BulkInsertInstagramPosts(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertInstagramPosts(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertInstagramPostsFunc = func(ctx context.Context, posts []*clickhousemodels.InstagramPost) error {
		return nil
	}
	err = mock.BulkInsertInstagramPosts(context.Background(), []*clickhousemodels.InstagramPost{{MediaID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertInstagramPostsFunc = func(ctx context.Context, posts []*clickhousemodels.InstagramPost) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertInstagramPosts(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_BulkInsertInstagramInsights(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertInstagramInsights(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertInstagramInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error {
		return nil
	}
	err = mock.BulkInsertInstagramInsights(context.Background(), []*clickhousemodels.InstagramInsight{{InstagramID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertInstagramInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertInstagramInsights(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_GetMinimalInstagramOlderThan20DaysByAccount(t *testing.T) {
	mock := &MockClickHouseClient{}

	posts, err := mock.GetMinimalInstagramOlderThan20DaysByAccount(context.Background(), "instagram_posts", "ig_123", 25, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if posts != nil {
		t.Fatalf("expected nil posts from default mock, got %#v", posts)
	}

	called := false
	mock.GetMinimalInstagramOlderThan20DaysByAccountFunc = func(ctx context.Context, tableName, instagramID string, limit, offset int) ([]clickhousemodels.InstagramMinimalPost, error) {
		called = true
		return []clickhousemodels.InstagramMinimalPost{{InstagramID: instagramID, MediaID: "media_1"}}, nil
	}

	posts, err = mock.GetMinimalInstagramOlderThan20DaysByAccount(context.Background(), "instagram_posts", "ig_123", 25, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected custom function to be called")
	}
	if len(posts) != 1 || posts[0].MediaID != "media_1" {
		t.Fatalf("unexpected posts: %#v", posts)
	}
}

func TestMockClickHouseClient_UpdateInstagramMediaURLs(t *testing.T) {
	mock := &MockClickHouseClient{}

	count, err := mock.UpdateInstagramMediaURLs(context.Background(), "instagram_posts", "ig_123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected zero count from default mock, got %d", count)
	}

	called := false
	mock.UpdateInstagramMediaURLsFunc = func(ctx context.Context, tableName, instagramID string, posts []clickhousemodels.InstagramMinimalPost) (int, error) {
		called = true
		return len(posts), nil
	}

	rows := []clickhousemodels.InstagramMinimalPost{{InstagramID: "ig_123", MediaID: "media_1"}}
	count, err = mock.UpdateInstagramMediaURLs(context.Background(), "instagram_posts", "ig_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected custom function to be called")
	}
	if count != 1 {
		t.Fatalf("expected count 1, got %d", count)
	}
}

func TestMockClickHouseClient_BulkInsertLinkedInPosts(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertLinkedInPosts(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertLinkedInPostsFunc = func(ctx context.Context, posts []*clickhousemodels.LinkedInPosts) error {
		return nil
	}
	err = mock.BulkInsertLinkedInPosts(context.Background(), []*clickhousemodels.LinkedInPosts{{PostID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertLinkedInPostsFunc = func(ctx context.Context, posts []*clickhousemodels.LinkedInPosts) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertLinkedInPosts(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_BulkInsertLinkedInInsights(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertLinkedInInsights(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertLinkedInInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.LinkedInInsights) error {
		return nil
	}
	err = mock.BulkInsertLinkedInInsights(context.Background(), []*clickhousemodels.LinkedInInsights{{LinkedinID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertLinkedInInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.LinkedInInsights) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertLinkedInInsights(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_BulkInsertTikTokPosts(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertTikTokPosts(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertTikTokPostsFunc = func(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error {
		return nil
	}
	err = mock.BulkInsertTikTokPosts(context.Background(), []*clickhousemodels.TikTokPosts{{PostID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertTikTokPostsFunc = func(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertTikTokPosts(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_BulkInsertTikTokInsights(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.BulkInsertTikTokInsights(context.Background(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.BulkInsertTikTokInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error {
		return nil
	}
	err = mock.BulkInsertTikTokInsights(context.Background(), []*clickhousemodels.TikTokInsights{{TikTokID: "test"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.BulkInsertTikTokInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error {
		return errors.New("insert failed")
	}
	err = mock.BulkInsertTikTokInsights(context.Background(), nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_Health(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	err := mock.Health()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	mock.HealthFunc = func() error {
		return nil
	}
	err = mock.Health()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with error
	mock.HealthFunc = func() error {
		return errors.New("health check failed")
	}
	err = mock.Health()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_GetMinimalOlderThan20DaysByPage(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	result, err := mock.GetMinimalOlderThan20DaysByPage(context.Background(), "table", "page123", 25, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result")
	}

	// Test with custom function
	expectedPosts := []clickhousemodels.MinimalPost{{PostID: "post1"}, {PostID: "post2"}}
	mock.GetMinimalOlderThan20DaysByPageFunc = func(ctx context.Context, tableName, pageID string, limit, offset int) ([]clickhousemodels.MinimalPost, error) {
		return expectedPosts, nil
	}
	result, err = mock.GetMinimalOlderThan20DaysByPage(context.Background(), "table", "page123", 25, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(result))
	}

	// Test with error
	mock.GetMinimalOlderThan20DaysByPageFunc = func(ctx context.Context, tableName, pageID string, limit, offset int) ([]clickhousemodels.MinimalPost, error) {
		return nil, errors.New("query failed")
	}
	_, err = mock.GetMinimalOlderThan20DaysByPage(context.Background(), "table", "page123", 25, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseClient_UpdateFullPictures(t *testing.T) {
	mock := &MockClickHouseClient{}

	// Test with nil function
	count, err := mock.UpdateFullPictures(context.Background(), "table", "page123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	// Test with custom function
	mock.UpdateFullPicturesFunc = func(ctx context.Context, tableName, pageID string, posts []clickhousemodels.MinimalPost) (int, error) {
		return 5, nil
	}
	count, err = mock.UpdateFullPictures(context.Background(), "table", "page123", []clickhousemodels.MinimalPost{{PostID: "post1"}})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 5 {
		t.Fatalf("expected 5, got %d", count)
	}

	// Test with error
	mock.UpdateFullPicturesFunc = func(ctx context.Context, tableName, pageID string, posts []clickhousemodels.MinimalPost) (int, error) {
		return 0, errors.New("update failed")
	}
	_, err = mock.UpdateFullPictures(context.Background(), "table", "page123", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClickHouseSink_ConvertFacebookPost(t *testing.T) {
	mock := &MockClickHouseSink{}
	input := &kafkamodels.ParsedFacebookPost{
		PostID: "post123",
		PageID: "page456",
	}

	// Test with nil function (default implementation)
	result := mock.ConvertFacebookPost(input)
	if result.PostID != "post123" {
		t.Fatalf("expected PostID 'post123', got '%s'", result.PostID)
	}
	if result.PageID != "page456" {
		t.Fatalf("expected PageID 'page456', got '%s'", result.PageID)
	}

	// Test with custom function
	mock.ConvertFacebookPostFunc = func(p *kafkamodels.ParsedFacebookPost) *clickhousemodels.FacebookPosts {
		return &clickhousemodels.FacebookPosts{
			PostID: "custom_" + p.PostID,
			PageID: "custom_" + p.PageID,
		}
	}
	result = mock.ConvertFacebookPost(input)
	if result.PostID != "custom_post123" {
		t.Fatalf("expected PostID 'custom_post123', got '%s'", result.PostID)
	}
}

func TestMockClickHouseSink_ConvertFacebookMediaAssets(t *testing.T) {
	mock := &MockClickHouseSink{}
	input := &kafkamodels.ParsedFacebookMediaAsset{
		PostID: "post123",
	}

	// Test with nil function (default implementation)
	result := mock.ConvertFacebookMediaAssets(input)
	if result.PostID != "post123" {
		t.Fatalf("expected PostID 'post123', got '%s'", result.PostID)
	}

	// Test with custom function
	mock.ConvertFacebookMediaAssetsFunc = func(a *kafkamodels.ParsedFacebookMediaAsset) *clickhousemodels.FacebookMediaAssets {
		return &clickhousemodels.FacebookMediaAssets{
			PostID:  a.PostID,
			MediaID: "media_generated",
		}
	}
	result = mock.ConvertFacebookMediaAssets(input)
	if result.MediaID != "media_generated" {
		t.Fatalf("expected MediaID 'media_generated', got '%s'", result.MediaID)
	}
}

func TestMockClickHouseSink_ConvertFacebookInsights(t *testing.T) {
	mock := &MockClickHouseSink{}
	input := &kafkamodels.ParsedFacebookInsights{
		PageID: "page123",
	}

	// Test with nil function (default implementation)
	result := mock.ConvertFacebookInsights(input)
	if result.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got '%s'", result.PageID)
	}

	// Test with custom function
	mock.ConvertFacebookInsightsFunc = func(ins *kafkamodels.ParsedFacebookInsights) *clickhousemodels.FacebookInsights {
		return &clickhousemodels.FacebookInsights{
			PageID:   ins.PageID,
			PageFans: 10000,
		}
	}
	result = mock.ConvertFacebookInsights(input)
	if result.PageFans != 10000 {
		t.Fatalf("expected PageFans 10000, got %d", result.PageFans)
	}
}

func TestMockClickHouseSink_ConvertFacebookVideoInsights(t *testing.T) {
	mock := &MockClickHouseSink{}
	input := &kafkamodels.ParsedFacebookVideoInsights{
		PostID: "post123",
	}

	// Test with nil function (default implementation)
	result := mock.ConvertFacebookVideoInsights(input)
	if result.PostID != "post123" {
		t.Fatalf("expected PostID 'post123', got '%s'", result.PostID)
	}

	// Test with custom function
	mock.ConvertFacebookVideoInsightsFunc = func(vi *kafkamodels.ParsedFacebookVideoInsights) *clickhousemodels.FacebookVideoInsights {
		return &clickhousemodels.FacebookVideoInsights{
			PostID:          vi.PostID,
			TotalVideoViews: 5000,
		}
	}
	result = mock.ConvertFacebookVideoInsights(input)
	if result.TotalVideoViews != 5000 {
		t.Fatalf("expected TotalVideoViews 5000, got %d", result.TotalVideoViews)
	}
}

func TestMockClickHouseSink_ConvertFacebookReelsInsights(t *testing.T) {
	mock := &MockClickHouseSink{}
	input := &kafkamodels.ParsedFacebookReelsInsights{
		PostID: "post123",
	}

	// Test with nil function (default implementation)
	result := mock.ConvertFacebookReelsInsights(input)
	if result.PostID != "post123" {
		t.Fatalf("expected PostID 'post123', got '%s'", result.PostID)
	}

	// Test with custom function
	mock.ConvertFacebookReelsInsightsFunc = func(ri *kafkamodels.ParsedFacebookReelsInsights) *clickhousemodels.FacebookReelsInsights {
		return &clickhousemodels.FacebookReelsInsights{
			PostID:    ri.PostID,
			PlayCount: 1000,
		}
	}
	result = mock.ConvertFacebookReelsInsights(input)
	if result.PlayCount != 1000 {
		t.Fatalf("expected PlayCount 1000, got %d", result.PlayCount)
	}
}

func TestNewMockClickHouseSink(t *testing.T) {
	mock := NewMockClickHouseSink()
	if mock == nil {
		t.Fatal("expected non-nil mock")
	}

	// Verify default functions work
	input := &kafkamodels.ParsedFacebookPost{PostID: "test", PageID: "page"}
	result := mock.ConvertFacebookPost(input)
	if result.PostID != "test" {
		t.Fatalf("expected PostID 'test', got '%s'", result.PostID)
	}
}

func TestMockClickHouseClient_ImplementsInterface(t *testing.T) {
	var _ ClickHouseClientInterface = (*MockClickHouseClient)(nil)
}
