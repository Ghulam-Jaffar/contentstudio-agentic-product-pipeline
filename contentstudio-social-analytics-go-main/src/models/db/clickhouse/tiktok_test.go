package clickhouse

import (
	"encoding/json"
	"testing"
	"time"
)

func TestTikTokPosts_Struct(t *testing.T) {
	now := time.Now()
	post := TikTokPosts{
		TikTokID:        "tt456",
		DisplayName:     "Test Creator",
		ProfileLink:     "https://tiktok.com/@testcreator",
		PostID:          "post789",
		CoverImageURL:   "https://example.com/cover.jpg",
		ShareURL:        "https://tiktok.com/@testcreator/video/123",
		PostDescription: "Test video description #viral #tiktok",
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
		CreatedAt:       now,
		InsertedAt:      now,
	}

	if post.TikTokID != "tt456" {
		t.Fatalf("expected TikTokID 'tt456', got %s", post.TikTokID)
	}
	if post.ViewCount != 100000 {
		t.Fatalf("expected ViewCount 100000, got %d", post.ViewCount)
	}
	if post.EngagementRate != 11.5 {
		t.Fatalf("expected EngagementRate 11.5, got %f", post.EngagementRate)
	}
	if len(post.Hashtags) != 3 {
		t.Fatalf("expected 3 hashtags, got %d", len(post.Hashtags))
	}
}

func TestTikTokPosts_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	post := TikTokPosts{
		TikTokID:   "tt123",
		PostID:     "post456",
		ViewCount:  100000,
		LikeCount:  10000,
		InsertedAt: now,
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result TikTokPosts
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.ViewCount != post.ViewCount {
		t.Fatalf("expected ViewCount %d, got %d", post.ViewCount, result.ViewCount)
	}
}

func TestTikTokPosts_VideoDimensions(t *testing.T) {
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
		post := TikTokPosts{
			TikTokID: "tt123",
			Width:    tc.width,
			Height:   tc.height,
			Duration: tc.duration,
		}

		if post.Width != tc.width {
			t.Fatalf("expected Width %d, got %d", tc.width, post.Width)
		}
		if post.Height != tc.height {
			t.Fatalf("expected Height %d, got %d", tc.height, post.Height)
		}
	}
}

func TestTikTokPosts_EngagementCalculation(t *testing.T) {
	post := TikTokPosts{
		TikTokID:     "tt123",
		LikeCount:    10000,
		CommentCount: 500,
		ShareCount:   1000,
		ViewCount:    100000,
	}

	totalEngagement := post.LikeCount + post.CommentCount + post.ShareCount
	if totalEngagement != 11500 {
		t.Fatalf("expected total engagement 11500, got %d", totalEngagement)
	}

	engagementRate := float64(totalEngagement) / float64(post.ViewCount) * 100
	expectedRate := 11.5
	if engagementRate != expectedRate {
		t.Fatalf("expected engagement rate %f, got %f", expectedRate, engagementRate)
	}
}

func TestTikTokInsights_Struct(t *testing.T) {
	now := time.Now()
	insights := TikTokInsights{
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
		Bio:                 "Content creator | Lifestyle | Travel",
		ProfileLink:         "https://tiktok.com/@testcreator",
		InsertedAt:          now,
	}

	if insights.TikTokID != "tt456" {
		t.Fatalf("expected TikTokID 'tt456', got %s", insights.TikTokID)
	}
	if insights.TotalFollowerCount != 500000 {
		t.Fatalf("expected TotalFollowerCount 500000, got %d", insights.TotalFollowerCount)
	}
	if !insights.IsVerified {
		t.Fatal("expected IsVerified to be true")
	}
}

func TestTikTokInsights_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	insights := TikTokInsights{
		TikTokID:           "tt123",
		TotalFollowerCount: 500000,
		TotalVideoViews:    50000000,
		IsVerified:         true,
		InsertedAt:         now,
	}

	data, err := json.Marshal(insights)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result TikTokInsights
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.TotalFollowerCount != insights.TotalFollowerCount {
		t.Fatalf("expected TotalFollowerCount %d, got %d", insights.TotalFollowerCount, result.TotalFollowerCount)
	}
	if result.IsVerified != insights.IsVerified {
		t.Fatalf("expected IsVerified %t, got %t", insights.IsVerified, result.IsVerified)
	}
}

func TestTikTokInsights_VerifiedStatus(t *testing.T) {
	cases := []struct {
		isVerified bool
	}{
		{true},
		{false},
	}

	for _, tc := range cases {
		insights := TikTokInsights{
			TikTokID:   "tt123",
			IsVerified: tc.isVerified,
		}

		if insights.IsVerified != tc.isVerified {
			t.Fatalf("expected IsVerified %t, got %t", tc.isVerified, insights.IsVerified)
		}
	}
}

func TestTikTokInsights_AccountMetrics(t *testing.T) {
	insights := TikTokInsights{
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

	avgViewsPerVideo := insights.TotalVideoViews / insights.TotalVideoCount
	if avgViewsPerVideo != 200000 {
		t.Fatalf("expected avg views per video 200000, got %d", avgViewsPerVideo)
	}

	avgLikesPerVideo := insights.TotalVideoLikes / insights.TotalVideoCount
	if avgLikesPerVideo != 18000 {
		t.Fatalf("expected avg likes per video 18000, got %d", avgLikesPerVideo)
	}
}

func TestTikTokPosts_Hashtags(t *testing.T) {
	post := TikTokPosts{
		TikTokID:        "tt123",
		PostDescription: "Check out this #viral video #fyp #tiktok #trending",
		Hashtags:        []string{"viral", "fyp", "tiktok", "trending"},
	}

	if len(post.Hashtags) != 4 {
		t.Fatalf("expected 4 hashtags, got %d", len(post.Hashtags))
	}

	expectedHashtags := map[string]bool{"viral": true, "fyp": true, "tiktok": true, "trending": true}
	for _, hashtag := range post.Hashtags {
		if !expectedHashtags[hashtag] {
			t.Fatalf("unexpected hashtag: %s", hashtag)
		}
	}
}
