package clickhouse

import (
	"context"
	"errors"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func Test_GetTikTokPostsViewSum_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{values: []any{int64(50000)}},
	}
	client := newTestClient(conn)

	total, err := client.GetTikTokPostsViewSum(context.Background(), "tiktok_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 50000 {
		t.Fatalf("expected 50000, got %d", total)
	}
}

func Test_GetTikTokPostsViewSum_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	client := newTestClient(conn)

	_, err := client.GetTikTokPostsViewSum(context.Background(), "tiktok_123")
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetTikTokPostsViewSum_ZeroViews(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{values: []any{int64(0)}},
	}
	client := newTestClient(conn)

	total, err := client.GetTikTokPostsViewSum(context.Background(), "tiktok_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if total != 0 {
		t.Fatalf("expected 0, got %d", total)
	}
}

func Test_BulkInsertTikTokPosts_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertTikTokPosts(context.Background(), []*clickhousemodels.TikTokPosts{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertTikTokPosts_WithAllFields(t *testing.T) {
	now := time.Now()
	posts := []*clickhousemodels.TikTokPosts{
		{
			TikTokID:        "tt_123",
			DisplayName:     "Test User",
			ProfileLink:     "https://tiktok.com/@testuser",
			PostID:          "post_123",
			CoverImageURL:   "https://example.com/cover.jpg",
			ShareURL:        "https://tiktok.com/@testuser/video/123",
			PostDescription: "Test post description #test",
			Hashtags:        []string{"test", "golang"},
			Duration:        30,
			Height:          1920,
			Width:           1080,
			Title:           "Test Video",
			EmbedHTML:       "<iframe></iframe>",
			EmbedLink:       "https://tiktok.com/embed/123",
			LikeCount:       1000,
			CommentCount:    50,
			ShareCount:      100,
			ViewCount:       50000,
			EngagementCount: 1150,
			EngagementRate:  2.3,
			CreatedAt:       now,
			InsertedAt:      now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertTikTokPosts(context.Background(), posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertTikTokInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertTikTokInsights(context.Background(), []*clickhousemodels.TikTokInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertTikTokInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.TikTokInsights{
		{
			RecordID:            "rec_1",
			TikTokID:            "tt_123",
			DisplayName:         "Test User",
			ProfileImage:        "https://example.com/profile.jpg",
			TotalFollowerCount:  100000,
			TotalFollowingCount: 500,
			TotalLikeCount:      5000000,
			TotalVideoCount:     200,
			TotalVideoViews:     10000000,
			TotalVideoLikes:     500000,
			TotalVideoComments:  50000,
			TotalVideoShares:    25000,
			IsVerified:          true,
			Bio:                 "Test user bio",
			ProfileLink:         "https://tiktok.com/@testuser",
			InsertedAt:          now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertTikTokInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertTikTokPosts_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	posts := []*clickhousemodels.TikTokPosts{
		{TikTokID: "tt_123", PostID: "post_456"},
	}
	_ = client.BulkInsertTikTokPosts(ctx, posts)
}

func Test_BulkInsertTikTokInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []*clickhousemodels.TikTokInsights{
		{RecordID: "rec_1", TikTokID: "tt_123"},
	}
	_ = client.BulkInsertTikTokInsights(ctx, insights)
}

func Test_BulkInsertTikTokPosts_MultipleItems(t *testing.T) {
	now := time.Now()
	posts := []*clickhousemodels.TikTokPosts{
		{TikTokID: "tt_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
		{TikTokID: "tt_1", PostID: "post_2", CreatedAt: now, InsertedAt: now},
		{TikTokID: "tt_1", PostID: "post_3", CreatedAt: now, InsertedAt: now},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertTikTokPosts(context.Background(), posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertTikTokInsights_MultipleItems(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.TikTokInsights{
		{RecordID: "rec_1", TikTokID: "tt_1", InsertedAt: now},
		{RecordID: "rec_2", TikTokID: "tt_2", InsertedAt: now},
		{RecordID: "rec_3", TikTokID: "tt_3", InsertedAt: now},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertTikTokInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
