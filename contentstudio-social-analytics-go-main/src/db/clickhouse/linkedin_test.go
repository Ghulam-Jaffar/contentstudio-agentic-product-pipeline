package clickhouse

import (
	"context"
	"errors"
	"testing"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func Test_BulkInsertLinkedInPosts_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertLinkedInPosts(context.Background(), []*chmodels.LinkedInPosts{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertLinkedInPosts_WithAllFields(t *testing.T) {
	now := time.Now()
	posts := []*chmodels.LinkedInPosts{
		{
			LinkedinID: "li_123", PostID: "urn:li:share:123456",
			Activity: "urn:li:activity:123456", Comments: 50, TotalEngagement: 500.5,
			Favorites: 200, PollData: "{}", Reach: 10000, Repost: 30,
			PostClicks: 150, Impressions: 20000, Title: "Test Post",
			Image: "https://example.com/img.jpg", ArticleURL: "https://example.com/article",
			ArticleTitle: "Test Article", Media: []string{"https://example.com/media1.jpg"},
			MediaType: "images", Type: "share", Hashtags: []string{"#linkedin", "#test"},
			DayOfWeek: "Wednesday", HourOfDay: 14,
			CreatedAt: now, PublishedAt: now, LastModifiedAt: now,
			LifecycleState: "PUBLISHED", Visibility: "PUBLIC",
			SavingTime: now, IsReshareDisabled: false,
			FeedDistribution: "MAIN_FEED", ThirdPartyChannels: []string{"twitter"},
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertLinkedInPosts(context.Background(), posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertLinkedInPosts_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	posts := []*chmodels.LinkedInPosts{{LinkedinID: "li_1", PostID: "post_1"}}
	_ = client.BulkInsertLinkedInPosts(ctx, posts)
}

func Test_BulkInsertLinkedInInsights_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.BulkInsertLinkedInInsights(context.Background(), []*chmodels.LinkedInInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_BulkInsertLinkedInInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*chmodels.LinkedInInsights{
		{
			LinkedinID: "li_123", OrganizationName: "Test Org",
			RecordID: "rec_123", ImpressionCount: 50000,
			OrganicFollowerCount: 10000, TotalFollowerCount: 12000, PaidFollowerCount: 2000,
			DailyFollowerCount: 50, Reach: 30000, Repost: 100,
			Comments: 200, PostClicks: 500, Reactions: 1000, Engagement: 5.5,
			FollowersBySeniority: "{\"senior\":100}", FollowersByIndustry: "{\"tech\":500}",
			FollowersByCountry: "{\"US\":3000}", FollowersByCity: "{\"SF\":500}",
			InsertedAt: now, CreatedAt: now,
			PageViews: 5000, UniqueVisitors: 3000,
			DesktopPageViews: 3000, MobilePageViews: 2000,
			OverviewPageViews: 2000, AboutPageViews: 500,
			JobsPageViews: 300, PeoplePageViews: 400,
			CareersPageViews: 200, LifeAtPageViews: 100,
			InsightsPageViews: 50, ProductsPageViews: 150,
			PageViewsByCountry: "{\"US\":2000}", PageViewsByRegion: "{\"CA\":500}",
			PageViewsByIndustry: "{\"tech\":1000}", PageViewsBySeniority: "{\"senior\":300}",
			PageViewsByFunction: "{\"eng\":400}", PageViewsByStaffCount: "{\"1000+\":600}",
		},
	}

	client := newTestClient(&mockConn{})
	err := client.BulkInsertLinkedInInsights(context.Background(), insights)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func Test_BulkInsertLinkedInInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	insights := []*chmodels.LinkedInInsights{{LinkedinID: "li_1", RecordID: "rec_1"}}
	_ = client.BulkInsertLinkedInInsights(ctx, insights)
}

func Test_GetGeoMappings_EmptyIDs(t *testing.T) {
	client := newTestClient(&mockConn{})
	result, err := client.GetGeoMappings(context.Background(), []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(result))
	}
}

func Test_GetGeoMappings_QueryError(t *testing.T) {
	client := newTestClient(&mockConn{queryErr: errors.New("query failed")})
	_, err := client.GetGeoMappings(context.Background(), []string{"geo_1"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetGeoMappings_Success(t *testing.T) {
	client := newTestClient(&mockConn{
		queryRows: &mockRows{
			nextCount: 2,
			scanValues: [][]any{
				{"geo_1", "United States"},
				{"geo_2", "Canada"},
			},
		},
	})
	result, err := client.GetGeoMappings(context.Background(), []string{"geo_1", "geo_2"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(result))
	}
	if result["geo_1"] != "United States" {
		t.Fatalf("expected 'United States', got %q", result["geo_1"])
	}
	if result["geo_2"] != "Canada" {
		t.Fatalf("expected 'Canada', got %q", result["geo_2"])
	}
}

func Test_GetGeoMappings_ScanError(t *testing.T) {
	client := newTestClient(&mockConn{
		queryRows: &mockRows{
			nextCount: 1,
			scanErr:   errors.New("scan failed"),
		},
	})
	_, err := client.GetGeoMappings(context.Background(), []string{"geo_1"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetGeoMappings_RowsError(t *testing.T) {
	client := newTestClient(&mockConn{
		queryRows: &mockRows{
			nextCount: 0,
			errVal:    errors.New("rows error"),
		},
	})
	_, err := client.GetGeoMappings(context.Background(), []string{"geo_1"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_InsertGeoMappings_EmptyMap(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.InsertGeoMappings(context.Background(), map[string]string{})
	if err != nil {
		t.Fatalf("expected nil error for empty map, got %v", err)
	}
}

func Test_InsertGeoMappings_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	_ = client.InsertGeoMappings(ctx, map[string]string{"geo_1": "United States"})
}

func Test_InsertGeoMappingsWithType_EmptySlice(t *testing.T) {
	client := newTestClient(&mockConn{})
	err := client.InsertGeoMappingsWithType(context.Background(), []GeoMappingWithType{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func Test_InsertGeoMappingsWithType_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := newTestClient(&mockConn{})
	mappings := []GeoMappingWithType{{ID: "geo_1", Name: "US", Type: "country"}}
	_ = client.InsertGeoMappingsWithType(ctx, mappings)
}
