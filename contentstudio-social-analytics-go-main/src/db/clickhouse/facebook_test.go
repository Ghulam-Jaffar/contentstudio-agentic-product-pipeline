package clickhouse

import (
	"context"
	"errors"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func Test_BulkInsertPosts_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertPosts(context.Background(), []*clickhousemodels.FacebookPosts{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertPosts_WithAllFields(t *testing.T) {
	now := time.Now()
	posts := []*clickhousemodels.FacebookPosts{
		{
			PageName: "Test Page", PageID: "page_123", MediaType: "photo",
			PostID: "post_123", Permalink: "https://fb.com/post_123", StatusType: "added_photos",
			VideoID: "vid_1", Category: "entertainment", PublishedBy: "admin",
			PublishedByURL: "https://fb.com/admin", SharedFromName: "Other Page",
			SharedFromID: "page_other", SharedFromLink: "https://fb.com/other",
			Like: 100, Love: 50, Haha: 20, Wow: 10, Sad: 5, Angry: 2, Thankful: 1, Total: 188,
			Shares: 30, Comments: 45, PostClicks: 200, TotalEngagement: 463,
			PostEngagedUsers: 300, DayOfWeek: "Monday", HourOfDay: 14,
			CreatedTime: now, UpdatedTime: now, SavingTime: now,
			MessageTags: []string{"tag1", "tag2"}, PostMetadata: "{}", Caption: "Test caption",
			Description: "Test description", FullPicture: "https://example.com/pic.jpg",
			Link: "https://example.com", PostImpressions: 5000, PostImpressionsUnique: 4000,
			PostImpressionsPaid: 1000, PostImpressionsPaidUnique: 800,
			PostImpressionsOrganic: 4000, PostImpressionsOrganicUnique: 3200,
			PostImpressionsViral: 500, PostImpressionsViralUnique: 400,
			PostVideoViews: 1500, TotalImpressions: 5500,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertPosts(context.Background(), posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertPosts_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	posts := []*clickhousemodels.FacebookPosts{{PageID: "page_1", PostID: "post_1"}}
	_ = client.BulkInsertPosts(ctx, posts)
}

func Test_BulkInsertMediaAssets_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertMediaAssets(context.Background(), []*clickhousemodels.FacebookMediaAssets{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertMediaAssets_WithAllFields(t *testing.T) {
	now := time.Now()
	assets := []*clickhousemodels.FacebookMediaAssets{
		{
			PageID: "page_123", MediaID: "media_123", PostID: "post_123",
			AssetType: "photo", Link: "https://example.com/photo.jpg",
			CallToAction: "Learn More", CTAType: "LEARN_MORE",
			Caption: "Photo caption", Description: "Photo description",
			CreatedAt: now, InsertedAt: now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertMediaAssets(context.Background(), assets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertMediaAssets_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	assets := []*clickhousemodels.FacebookMediaAssets{{PageID: "page_1", MediaID: "media_1"}}
	_ = client.BulkInsertMediaAssets(ctx, assets)
}

func Test_BulkInsertVideoInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertVideoInsights(context.Background(), []*clickhousemodels.FacebookVideoInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertVideoInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.FacebookVideoInsights{
		{
			PageID: "page_123", PostID: "post_123", VideoID: "vid_123",
			CreatedTime: now, UpdatedTime: now,
			TotalVideoViews: 10000, TotalVideoImpressions: 50000, TotalVideoCompleteViews: 5000,
			TotalVideo10sViews: 8000, TotalVideo15sViews: 7000, TotalVideo30sViews: 4000,
			TotalVideo60sExcludesShorterViews: 2000, TotalVideoAvgTimeWatched: 45,
			TotalVideoImpressionsUnique: 40000, TotalVideoViewTotalTime: 300000,
			TotalVideoViewsUnique: 9000,
			TotalVideoStoriesByActionType: []string{"share:100", "like:500"},
			TotalVideoViewTotalTimeOrganic: 250000,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertVideoInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertVideoInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []*clickhousemodels.FacebookVideoInsights{{PageID: "page_1", PostID: "post_1"}}
	_ = client.BulkInsertVideoInsights(ctx, insights)
}

func Test_BulkInsertReelsInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertReelsInsights(context.Background(), []*clickhousemodels.FacebookReelsInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertReelsInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.FacebookReelsInsights{
		{
			PageID: "page_123", PostID: "post_123",
			AverageTimeWatched: 15, TotalTimeWatchedInMs: 300000,
			PlayCount: 5000, ImpressionsUnique: 4000, ReelFollowers: 200,
			CreatedAt: now, SavingTime: now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertReelsInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertReelsInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []*clickhousemodels.FacebookReelsInsights{{PageID: "page_1", PostID: "post_1"}}
	_ = client.BulkInsertReelsInsights(ctx, insights)
}

func Test_BulkInsertInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertInsights(context.Background(), []*clickhousemodels.FacebookInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.FacebookInsights{
		{
			HashID: "hash_123", PageID: "page_123", PageCategory: "Entertainment",
			DayOfWeek: "Monday", Year: 2025, Month: 6, CreatedTime: now, SavingTime: now,
			PageFans: 50000, PageFansCity: []string{"NYC:1000", "LA:500"},
			PageFansCountry: []string{"US:3000", "UK:1000"}, PageFansLocale: []string{"en_US:4000"},
			PageFansAge: []string{"18-24:2000", "25-34:3000"}, PageFansGender: []string{"M:2500", "F:2500"},
			PageFansGenderAge: []string{"M.18-24:1000"}, PageFollows: 48000, PageViews: 10000,
			PageFanAddsByPaidNonPaidUnique: []string{"paid:100", "unpaid:200"},
			PageFanAddsUnique: 300, PageFanRemovesUnique: 50,
			PageFansByLikeSourceUnique: []string{"search:100"}, PageFansByUnlikeSourceUnique: []string{"news_feed:20"},
			PageFansByLike: 350, PageFansByUnlike: 50, PageTotalActions: 500, PagePostEngagements: 2000,
			PageImpressions: 100000, PageImpressionsOrganic: 80000, PageImpressionsPaid: 20000,
			PageVideoViewsPaid: 5000, PageVideoViews: 15000, PageVideoViewsOrganic: 10000,
			PageVideoViewsAutoplayed: 8000, PageVideoViewsClickToPlay: 7000, PageVideoRepeatViews: 3000,
			PageNegativeFeedback: 10, PagePositiveFeedback: 500,
			PageNegativeFeedbackByType: []string{"hide_all_clicks:5"}, PagePositiveFeedbackByType: []string{"like:300"},
			PageFansOnline: []string{"0:100", "1:200"}, ActiveUsers: 5000,
			PositiveSentiment: 800, NegativeSentiment: 20,
			PostsCount: 50, LikesCount: 10000, TalkingAboutCount: 3000,
			TypeCount: []string{"photo:20", "video:15", "link:15"},
			MessageCount: []string{"sent:100", "received:200"},
			PrimeTime: now, PageImpressionsUnique: 70000,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []*clickhousemodels.FacebookInsights{{HashID: "hash_1", PageID: "page_1"}}
	_ = client.BulkInsertInsights(ctx, insights)
}

func Test_GetMinimalOlderThan20DaysByPage_EmptyPageID(t *testing.T) {
	client := newTestClient(&mockConn{})
	_, err := client.GetMinimalOlderThan20DaysByPage(context.Background(), "facebook_posts", "", 500, 0)
	if err == nil {
		t.Fatalf("expected error for empty pageID, got nil")
	}
}

func Test_GetMinimalOlderThan20DaysByPage_DefaultTable(t *testing.T) {
	client := newTestClient(&mockConn{queryRows: &mockRows{nextCount: 0}})
	posts, err := client.GetMinimalOlderThan20DaysByPage(context.Background(), "", "page_123", 500, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 0 {
		t.Fatalf("expected empty result, got %d", len(posts))
	}
}

func Test_GetMinimalOlderThan20DaysByPage_QueryError(t *testing.T) {
	client := newTestClient(&mockConn{queryErr: errors.New("query failed")})
	_, err := client.GetMinimalOlderThan20DaysByPage(context.Background(), "facebook_posts", "page_123", 500, 0)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetMinimalOlderThan20DaysByPage_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{queryRows: &mockRows{nextCount: 0}})
	_, _ = client.GetMinimalOlderThan20DaysByPage(ctx, "facebook_posts", "page_123", 500, 0)
}

func Test_UpdateFullPictures_EmptyPageID(t *testing.T) {
	client := newTestClient(&mockConn{})
	_, err := client.UpdateFullPictures(context.Background(), "facebook_posts", "", nil)
	if err == nil {
		t.Fatalf("expected error for empty pageID, got nil")
	}
}

func Test_UpdateFullPictures_EmptyRows(t *testing.T) {
	client := newTestClient(&mockConn{})
	count, err := client.UpdateFullPictures(context.Background(), "facebook_posts", "page_123", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 count for empty rows, got %d", count)
	}
}

func Test_UpdateFullPictures_DefaultTable(t *testing.T) {
	client := newTestClient(&mockConn{})
	rows := []clickhousemodels.MinimalPost{
		{PageID: "page_123", PostID: "post_1", FullPicture: "https://example.com/pic1.jpg"},
	}
	count, err := client.UpdateFullPictures(context.Background(), "", "page_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1, got %d", count)
	}
}

func Test_UpdateFullPictures_ExecError(t *testing.T) {
	client := newTestClient(&mockConn{execErr: errors.New("exec failed")})
	rows := []clickhousemodels.MinimalPost{
		{PageID: "page_123", PostID: "post_1", FullPicture: "https://example.com/pic1.jpg"},
	}
	_, err := client.UpdateFullPictures(context.Background(), "facebook_posts", "page_123", rows)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_UpdateFullPictures_SkipsMismatchedPageID(t *testing.T) {
	client := newTestClient(&mockConn{})
	rows := []clickhousemodels.MinimalPost{
		{PageID: "other_page", PostID: "post_1", FullPicture: "https://example.com/pic1.jpg"},
	}
	count, err := client.UpdateFullPictures(context.Background(), "facebook_posts", "page_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 for mismatched page, got %d", count)
	}
}

func Test_UpdateFullPictures_SkipsEmptyPostIDAndPicture(t *testing.T) {
	client := newTestClient(&mockConn{})
	rows := []clickhousemodels.MinimalPost{
		{PageID: "page_123", PostID: "", FullPicture: "https://example.com/pic1.jpg"},
		{PageID: "page_123", PostID: "post_1", FullPicture: ""},
		{PageID: "page_123", PostID: "post_1", FullPicture: "   "},
	}
	count, err := client.UpdateFullPictures(context.Background(), "facebook_posts", "page_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0 for invalid rows, got %d", count)
	}
}

func Test_UpdateFullPictures_DeduplicatesByPostID(t *testing.T) {
	client := newTestClient(&mockConn{})
	rows := []clickhousemodels.MinimalPost{
		{PageID: "page_123", PostID: "post_1", FullPicture: "https://example.com/pic1.jpg"},
		{PageID: "page_123", PostID: "post_1", FullPicture: "https://example.com/pic2.jpg"},
	}
	count, err := client.UpdateFullPictures(context.Background(), "facebook_posts", "page_123", rows)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected 1 after dedup, got %d", count)
	}
}
