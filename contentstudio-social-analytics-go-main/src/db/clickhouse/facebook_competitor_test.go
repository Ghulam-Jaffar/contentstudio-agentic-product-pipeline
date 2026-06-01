package clickhouse

import (
	"context"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func Test_InsertCompetitorInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.InsertCompetitorInsights(context.Background(), []*clickhousemodels.FacebookCompetitorInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_InsertCompetitorInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.FacebookCompetitorInsights{
		{
			RecordID: "rec_123", PageID: "page_123",
			FollowersCount: 50000, TotalFanCount: 55000, TalkingAboutThis: 3000,
			Biography: "Test page bio", ProfilePictureURL: "https://example.com/pic.jpg",
			PageName: "Test Competitor", PageCategory: "Entertainment",
			Emails: []string{"test@example.com"}, Birthday: "2010-01-01",
			WereHereCount: 1000, CoverPhotoURL: "https://example.com/cover.jpg",
			Permalink: "https://fb.com/test",
			Metadata:  map[string]string{"source": "api"},
			InsertedAt: now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.InsertCompetitorInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_InsertCompetitorInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []*clickhousemodels.FacebookCompetitorInsights{{RecordID: "rec_1", PageID: "page_1"}}
	_ = client.InsertCompetitorInsights(ctx, insights)
}

func Test_InsertCompetitorPosts_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.InsertCompetitorPosts(context.Background(), []*clickhousemodels.FacebookCompetitorPosts{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_InsertCompetitorPosts_WithAllFields(t *testing.T) {
	now := time.Now()
	posts := []*clickhousemodels.FacebookCompetitorPosts{
		{
			FacebookID: "fb_123", PostID: "post_123",
			FollowersCount: 50000, FanCount: 55000,
			PageName: "Competitor Page", PageCategory: "Entertainment",
			Biography: "Competitor bio", PostEngagement: 500,
			Like: 200, Haha: 30, Angry: 5, Sad: 10, Thankful: 2, Love: 80, Wow: 15,
			TotalPostReactions: 342, Comments: 50, Shares: 108,
			Caption: "Test post caption", MediaType: "photo", StatusType: "added_photos",
			SharedFromName: "Original Page", SharedFromID: "page_orig",
			SharedFromPic: "https://example.com/orig.jpg", SharedCreatedAt: now.AddDate(0, 0, -1),
			Permalink: "https://fb.com/post_123", Hashtags: []string{"#test", "#social"},
			DayOfWeek: "Tuesday", HourOfDay: 10,
			CreatedAt: now, InsertedAt: now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.InsertCompetitorPosts(context.Background(), posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_InsertCompetitorPosts_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	posts := []*clickhousemodels.FacebookCompetitorPosts{{FacebookID: "fb_1", PostID: "post_1"}}
	_ = client.InsertCompetitorPosts(ctx, posts)
}

func Test_InsertCompetitorMediaAssets_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.InsertCompetitorMediaAssets(context.Background(), []*clickhousemodels.FacebookCompetitorMediaAssets{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_InsertCompetitorMediaAssets_WithAllFields(t *testing.T) {
	now := time.Now()
	assets := []*clickhousemodels.FacebookCompetitorMediaAssets{
		{
			MediaID: "media_123", PostID: "post_123", PageID: "page_123",
			Caption: "Media caption", Description: "Media description",
			Link: "https://example.com/media.jpg", AssetType: "photo",
			CallToAction: "Shop Now", CTAType: "SHOP_NOW",
			CreatedAt: now, InsertedAt: now,
		},
	}

	client := newTestClient(&mockConn{})
	err := client.InsertCompetitorMediaAssets(context.Background(), assets)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_InsertCompetitorMediaAssets_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	assets := []*clickhousemodels.FacebookCompetitorMediaAssets{{MediaID: "media_1", PostID: "post_1"}}
	_ = client.InsertCompetitorMediaAssets(ctx, assets)
}
