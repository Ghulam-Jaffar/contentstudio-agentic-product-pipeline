package conversions

import (
	"context"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestBulkInsertInstagramPosts_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertInstagramPosts(context.Background(), []*clickhousemodels.InstagramPost{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertInstagramPosts_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertInstagramPostsFunc: func(ctx context.Context, posts []*clickhousemodels.InstagramPost) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	posts := []*clickhousemodels.InstagramPost{{InstagramID: "ig123"}}
	err := sink.BulkInsertInstagramPosts(context.Background(), posts)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertInstagramPosts to be called")
	}
}

func TestBulkInsertInstagramInsights_EmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	err := sink.BulkInsertInstagramInsights(context.Background(), []*clickhousemodels.InstagramInsight{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

func TestBulkInsertInstagramInsights_Success(t *testing.T) {
	called := false
	mock := &mockClickHouseClient{
		bulkInsertInstagramInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error {
			called = true
			return nil
		},
	}
	sink := newTestSink(mock)

	insights := []*clickhousemodels.InstagramInsight{{InstagramID: "ig123"}}
	err := sink.BulkInsertInstagramInsights(context.Background(), insights)

	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if !called {
		t.Fatal("expected BulkInsertInstagramInsights to be called")
	}
}

func TestConvertInstagramPost_NilInput(t *testing.T) {
	sink := &ClickHouseSink{}
	result := sink.ConvertInstagramPost(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertInstagramPost_ValidInput(t *testing.T) {
	sink := &ClickHouseSink{}
	now := time.Now()

	input := &kafkamodels.ParsedInstagramPost{
		InstagramID:         "ig123",
		MediaID:             "media456",
		Username:            "testuser",
		Name:                "Test User",
		ProfilePictureURL:   "https://example.com/profile.jpg",
		Permalink:           "https://instagram.com/p/abc123",
		LikeCount:           1000,
		CommentsCount:       100,
		Engagement:          1100,
		Impressions:         5000,
		Views:               3000,
		Reach:               4000,
		Saved:               200,
		VideoViews:          2500,
		Shares:              50,
		ReelsAvgWatchTime:   5000,
		ReelsTotalWatchTime: 250000,
		Exits:               100,
		Replies:             25,
		TapsForward:         150,
		TapsBack:            30,
		ChildAssetsType:     []string{"IMAGE", "IMAGE", "VIDEO"},
		Caption:             "Test caption #test",
		MediaType:           "CAROUSEL_ALBUM",
		EntityType:          "FEED",
		MediaURL:            []string{"img1.jpg", "img2.jpg"},
		VideoURL:            []string{"vid1.mp4"},
		Hashtags:            []string{"test", "instagram"},
		DayOfWeek:           "Monday",
		HourOfDay:           14,
		Year:                2025,
		Month:               1,
		Timestamp:           now.Unix(),
		StoredEventAt:       now,
		PostCreatedAt:       now,
	}

	result := sink.ConvertInstagramPost(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.InstagramID != "ig123" {
		t.Fatalf("expected InstagramID 'ig123', got %s", result.InstagramID)
	}
	if result.MediaID != "media456" {
		t.Fatalf("expected MediaID 'media456', got %s", result.MediaID)
	}
	if result.LikeCount != 1000 {
		t.Fatalf("expected LikeCount 1000, got %d", result.LikeCount)
	}
	if result.Engagement != 1100 {
		t.Fatalf("expected Engagement 1100, got %d", result.Engagement)
	}
	if result.MediaType != "CAROUSEL_ALBUM" {
		t.Fatalf("expected MediaType 'CAROUSEL_ALBUM', got %s", result.MediaType)
	}
	if len(result.Hashtags) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(result.Hashtags))
	}
}

func TestConvertInstagramPost_AllMediaTypes(t *testing.T) {
	sink := &ClickHouseSink{}

	cases := []struct {
		mediaType  string
		entityType string
	}{
		{"IMAGE", "FEED"},
		{"VIDEO", "FEED"},
		{"CAROUSEL_ALBUM", "FEED"},
		{"VIDEO", "REELS"},
		{"IMAGE", "STORY"},
	}

	for _, tc := range cases {
		input := &kafkamodels.ParsedInstagramPost{
			InstagramID: "ig123",
			MediaType:   tc.mediaType,
			EntityType:  tc.entityType,
		}

		result := sink.ConvertInstagramPost(input)

		if result.MediaType != tc.mediaType {
			t.Fatalf("expected MediaType %s, got %s", tc.mediaType, result.MediaType)
		}
		if result.EntityType != tc.entityType {
			t.Fatalf("expected EntityType %s, got %s", tc.entityType, result.EntityType)
		}
	}
}

func TestConvertInstagramInsight_NilInput(t *testing.T) {
	sink := &ClickHouseSink{}
	result := sink.ConvertInstagramInsight(nil)
	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestConvertInstagramInsight_ValidInput(t *testing.T) {
	sink := &ClickHouseSink{}
	now := time.Now()

	input := &kafkamodels.ParsedInstagramInsight{
		InstagramID:       "ig123",
		RecordID:          "rec456",
		Name:              "Test Account",
		Username:          "testuser",
		ProfilePictureURL: "https://example.com/profile.jpg",
		FollowsCount:      500,
		FollowersCount:    10000,
		MediaCount:        200,
		Tags:              50,
		Impressions:       100000,
		ProfileViews:      5000,
		Shares:            500,
		AccountsEngaged:   3000,
		Engagement:        15000,
		Reach:             80000,
		Views:             50000,
		Saves:             2000,
		Likes:             30000,
		Comments:          5000,
		AudienceAge:       []string{"18-24:2000", "25-34:4000"},
		AudienceGender:    []string{"M:5500", "F:4500"},
		AudienceCity:      []string{"NYC:1000", "LA:800"},
		AudienceCountry:   []string{"US:6000", "UK:2000"},
		OnlineFollowers:   []string{"0:1000", "1:1200"},
		DayOfWeek:         "Monday",
		Year:              2025,
		Month:             1,
		CreatedTime:       now,
		UpdatedTime:       now,
		StoredEventAt:     now,
	}

	result := sink.ConvertInstagramInsight(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.InstagramID != "ig123" {
		t.Fatalf("expected InstagramID 'ig123', got %s", result.InstagramID)
	}
	if result.FollowersCount != 10000 {
		t.Fatalf("expected FollowersCount 10000, got %d", result.FollowersCount)
	}
	if result.Engagement != 15000 {
		t.Fatalf("expected Engagement 15000, got %d", result.Engagement)
	}
	if len(result.AudienceAge) != 2 {
		t.Fatalf("expected 2 age groups, got %d", len(result.AudienceAge))
	}
}

func TestConvertInstagramInsight_WithMetadata(t *testing.T) {
	sink := &ClickHouseSink{}

	input := &kafkamodels.ParsedInstagramInsight{
		InstagramID: "ig123",
		Metadata: map[string]string{
			"sync_type":    "incremental",
			"data_version": "v2",
		},
	}

	result := sink.ConvertInstagramInsight(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.Metadata["sync_type"] != "incremental" {
		t.Fatalf("expected sync_type 'incremental', got %s", result.Metadata["sync_type"])
	}
}

func TestConvertInstagramInsight_AllAudienceFields(t *testing.T) {
	sink := &ClickHouseSink{}

	input := &kafkamodels.ParsedInstagramInsight{
		InstagramID:                   "ig123",
		AudienceAge:                   []string{"18-24:2000"},
		AudienceGender:                []string{"M:5500"},
		AudienceGenderAge:             []string{"M.18-24:1000"},
		AudienceLocale:                []string{"en_US:6000"},
		AudienceCity:                  []string{"NYC:1000"},
		AudienceCountry:               []string{"US:6000"},
		AudienceAgeByEngagement:       []string{"18-24:500"},
		AudienceGenderByEngagement:    []string{"M:800"},
		AudienceGenderAgeByEngagement: []string{"M.18-24:400"},
		AudienceCityByEngagement:      []string{"NYC:200"},
		AudienceCountryByEngagement:   []string{"US:1200"},
		AudienceAgeByReach:            []string{"18-24:10000"},
		AudienceGenderByReach:         []string{"M:25000"},
		AudienceGenderAgeByReach:      []string{"M.18-24:12000"},
		AudienceCityByReach:           []string{"NYC:5000"},
		AudienceCountryByReach:        []string{"US:40000"},
	}

	result := sink.ConvertInstagramInsight(input)

	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(result.AudienceAge) != 1 {
		t.Fatalf("expected 1 age entry, got %d", len(result.AudienceAge))
	}
	if len(result.AudienceAgeByEngagement) != 1 {
		t.Fatalf("expected 1 age by engagement entry, got %d", len(result.AudienceAgeByEngagement))
	}
	if len(result.AudienceCountryByReach) != 1 {
		t.Fatalf("expected 1 country by reach entry, got %d", len(result.AudienceCountryByReach))
	}
}
