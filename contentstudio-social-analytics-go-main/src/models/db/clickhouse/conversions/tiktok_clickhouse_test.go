package conversions

import (
	"context"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestBulkInsertTikTokPosts_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertTikTokPostsFunc: func(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	posts := []*clickhousemodels.TikTokPosts{{TikTokID: "tt123"}}
	err := sink.BulkInsertTikTokPosts(context.Background(), posts)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertTikTokPosts to be called")
	}
}

func TestBulkInsertTikTokInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertTikTokInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := []*clickhousemodels.TikTokInsights{{TikTokID: "tt123"}}
	err := sink.BulkInsertTikTokInsights(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertTikTokInsights to be called")
	}
}

func TestConvertTikTokPost_NilInput(t *testing.T) {
	result := ConvertTikTokPost(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertTikTokPost_ValidInput(t *testing.T) {
	createTime := time.Now().Unix()

	input := &kafkamodels.ParsedTikTokPost{
		WorkspaceID:     "ws123",
		ID:              "tt456",
		TikTokID:        "tt456",
		DisplayName:     "Test Creator",
		ProfileLink:     "https://tiktok.com/@testcreator",
		CoverImageURL:   "https://example.com/cover.jpg",
		ShareURL:        "https://tiktok.com/@testcreator/video/123",
		PostDescription: "Test video #viral #tiktok",
		Hashtags:        []string{"viral", "tiktok", "fyp"},
		Duration:        30,
		Height:          1920,
		Width:           1080,
		Title:           "Test Video Title",
		EmbedHTML:       "<iframe src='...'></iframe>",
		EmbedLink:       "https://tiktok.com/embed/v2/123",
		LikeCount:       10000,
		CommentCount:    500,
		ShareCount:      1000,
		ViewCount:       100000,
		EngagementCount: 11500,
		EngagementRate:  11.5,
		CreateTime:      createTime,
	}

	result := ConvertTikTokPost(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.TikTokID != "tt456" {
		t.Fatalf("expected TikTokID 'tt456', got %s", result.TikTokID)
	}
	if result.PostID != "tt456" {
		t.Fatalf("expected PostID 'tt456', got %s", result.PostID)
	}
	if result.ViewCount != 100000 {
		t.Fatalf("expected ViewCount 100000, got %d", result.ViewCount)
	}
	if result.EngagementRate != 11.5 {
		t.Fatalf("expected EngagementRate 11.5, got %f", result.EngagementRate)
	}
	if len(result.Hashtags) != 3 {
		t.Fatalf("expected 3 hashtags, got %d", len(result.Hashtags))
	}
	if result.CreatedAt.Unix() != createTime {
		t.Fatalf("expected CreatedAt to match input CreateTime")
	}
	if result.InsertedAt.IsZero() {
		t.Fatal("expected InsertedAt to be set")
	}
}

func TestConvertTikTokPost_VideoDimensions(t *testing.T) {
	cases := []struct {
		width    int64
		height   int64
		duration int64
	}{
		{1080, 1920, 15},  // Vertical short
		{1080, 1920, 60},  // Vertical medium
		{1080, 1920, 180}, // Vertical 3min
		{1920, 1080, 30},  // Horizontal
		{1080, 1080, 30},  // Square
	}

	for _, tc := range cases {
		input := &kafkamodels.ParsedTikTokPost{
			ID:       "tt123",
			Width:    tc.width,
			Height:   tc.height,
			Duration: tc.duration,
		}

		result := ConvertTikTokPost(input)

		if result.Width != tc.width {
			t.Fatalf("expected Width %d, got %d", tc.width, result.Width)
		}
		if result.Height != tc.height {
			t.Fatalf("expected Height %d, got %d", tc.height, result.Height)
		}
		if result.Duration != tc.duration {
			t.Fatalf("expected Duration %d, got %d", tc.duration, result.Duration)
		}
	}
}

func TestConvertTikTokPost_Hashtags(t *testing.T) {
	input := &kafkamodels.ParsedTikTokPost{
		ID:              "tt123",
		PostDescription: "Check out this #viral video #fyp #tiktok #trending",
		Hashtags:        []string{"viral", "fyp", "tiktok", "trending"},
	}

	result := ConvertTikTokPost(input)

	if len(result.Hashtags) != 4 {
		t.Fatalf("expected 4 hashtags, got %d", len(result.Hashtags))
	}

	expectedHashtags := map[string]bool{"viral": true, "fyp": true, "tiktok": true, "trending": true}
	for _, hashtag := range result.Hashtags {
		if !expectedHashtags[hashtag] {
			t.Fatalf("unexpected hashtag: %s", hashtag)
		}
	}
}

func TestConvertTikTokInsights_NilInput(t *testing.T) {
	result := ConvertTikTokInsights(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertTikTokInsights_ValidInput(t *testing.T) {
	insertedAt := time.Now().Unix()

	input := &kafkamodels.ParsedTikTokInsights{
		RecordID:            "rec123",
		TikTokID:            "tt456",
		DisplayName:         "Test Creator",
		ProfileImage:        "https://example.com/profile.jpg",
		TotalFollowerCount:  500000,
		TotalFollowingCount: 100,
		TotalLikeCount:      5000000,
		TotalVideoCount:     250,
		TotalVideoViews:     50000000,
		TotalVideoLikes:     4500000,
		TotalVideoComments:  100000,
		TotalVideoShares:    500000,
		IsVerified:          true,
		Bio:                 "Content creator | Lifestyle",
		ProfileLink:         "https://tiktok.com/@testcreator",
		InsertedAt:          insertedAt,
	}

	result := ConvertTikTokInsights(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.TikTokID != "tt456" {
		t.Fatalf("expected TikTokID 'tt456', got %s", result.TikTokID)
	}
	if result.TotalFollowerCount != 500000 {
		t.Fatalf("expected TotalFollowerCount 500000, got %d", result.TotalFollowerCount)
	}
	if result.TotalVideoViews != 50000000 {
		t.Fatalf("expected TotalVideoViews 50000000, got %d", result.TotalVideoViews)
	}
	if !result.IsVerified {
		t.Fatal("expected IsVerified to be true")
	}
	if result.InsertedAt.Unix() != insertedAt {
		t.Fatal("expected InsertedAt to match input")
	}
}

func TestConvertTikTokInsights_VerifiedStatus(t *testing.T) {
	cases := []struct {
		isVerified bool
	}{
		{true},
		{false},
	}

	for _, tc := range cases {
		input := &kafkamodels.ParsedTikTokInsights{
			TikTokID:   "tt123",
			IsVerified: tc.isVerified,
		}

		result := ConvertTikTokInsights(input)

		if result.IsVerified != tc.isVerified {
			t.Fatalf("expected IsVerified %t, got %t", tc.isVerified, result.IsVerified)
		}
	}
}

func TestConvertTikTokInsights_AccountMetrics(t *testing.T) {
	input := &kafkamodels.ParsedTikTokInsights{
		TikTokID:            "tt123",
		TotalFollowerCount:  500000,
		TotalFollowingCount: 100,
		TotalLikeCount:      5000000,
		TotalVideoCount:     250,
		TotalVideoViews:     50000000,
		TotalVideoLikes:     4500000,
		TotalVideoComments:  100000,
		TotalVideoShares:    500000,
	}

	result := ConvertTikTokInsights(input)

	if result.TotalFollowerCount != 500000 {
		t.Fatalf("expected TotalFollowerCount 500000, got %d", result.TotalFollowerCount)
	}

	avgViewsPerVideo := result.TotalVideoViews / result.TotalVideoCount
	if avgViewsPerVideo != 200000 {
		t.Fatalf("expected avg views per video 200000, got %d", avgViewsPerVideo)
	}
}
