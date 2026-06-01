package clickhouse

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInstagramPost_Struct(t *testing.T) {
	now := time.Now()
	post := InstagramPost{
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
		Caption:             "Test caption #test #instagram",
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

	if post.InstagramID != "ig123" {
		t.Fatalf("expected InstagramID 'ig123', got %s", post.InstagramID)
	}
	if post.Engagement != 1100 {
		t.Fatalf("expected Engagement 1100, got %d", post.Engagement)
	}
	if len(post.Hashtags) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(post.Hashtags))
	}
}

func TestInstagramPost_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	post := InstagramPost{
		InstagramID:   "ig123",
		MediaID:       "media456",
		LikeCount:     1000,
		StoredEventAt: now,
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result InstagramPost
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.LikeCount != post.LikeCount {
		t.Fatalf("expected LikeCount %d, got %d", post.LikeCount, result.LikeCount)
	}
}

func TestInstagramPost_MediaTypes(t *testing.T) {
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
		post := InstagramPost{
			InstagramID: "ig123",
			MediaType:   tc.mediaType,
			EntityType:  tc.entityType,
		}

		if post.MediaType != tc.mediaType {
			t.Fatalf("expected MediaType %s, got %s", tc.mediaType, post.MediaType)
		}
		if post.EntityType != tc.entityType {
			t.Fatalf("expected EntityType %s, got %s", tc.entityType, post.EntityType)
		}
	}
}

func TestInstagramPost_StoryMetrics(t *testing.T) {
	post := InstagramPost{
		InstagramID: "ig123",
		MediaType:   "IMAGE",
		EntityType:  "STORY",
		Exits:       100,
		Replies:     25,
		TapsForward: 150,
		TapsBack:    30,
		Reach:       5000,
		Impressions: 6000,
	}

	if post.Exits != 100 {
		t.Fatalf("expected Exits 100, got %d", post.Exits)
	}
	if post.TapsForward != 150 {
		t.Fatalf("expected TapsForward 150, got %d", post.TapsForward)
	}
}

func TestInstagramInsight_Struct(t *testing.T) {
	now := time.Now()
	insight := InstagramInsight{
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
		AudienceAge:       []string{"18-24:2000", "25-34:4000", "35-44:2500"},
		AudienceGender:    []string{"M:5500", "F:4500"},
		AudienceGenderAge: []string{"M.18-24:1000", "F.18-24:1000", "M.25-34:2000", "F.25-34:2000"},
		AudienceLocale:    []string{"en_US:6000", "en_GB:2000"},
		AudienceCity:      []string{"New York:1000", "Los Angeles:800"},
		AudienceCountry:   []string{"US:6000", "UK:2000"},
		OnlineFollowers:   []string{"0:1000", "1:1200", "2:1500", "3:1800"},
		DayOfWeek:         "Monday",
		Year:              2025,
		Month:             1,
		CreatedTime:       now,
		UpdatedTime:       now,
		StoredEventAt:     now,
	}

	if insight.InstagramID != "ig123" {
		t.Fatalf("expected InstagramID 'ig123', got %s", insight.InstagramID)
	}
	if insight.FollowersCount != 10000 {
		t.Fatalf("expected FollowersCount 10000, got %d", insight.FollowersCount)
	}
	if len(insight.AudienceAge) != 3 {
		t.Fatalf("expected 3 age groups, got %d", len(insight.AudienceAge))
	}
}

func TestInstagramInsight_JSON_Marshal(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	insight := InstagramInsight{
		InstagramID:    "ig123",
		FollowersCount: 10000,
		CreatedTime:    now,
	}

	data, err := json.Marshal(insight)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result InstagramInsight
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if result.FollowersCount != insight.FollowersCount {
		t.Fatalf("expected FollowersCount %d, got %d", insight.FollowersCount, result.FollowersCount)
	}
}

func TestInstagramInsight_AudienceDemographics(t *testing.T) {
	insight := InstagramInsight{
		InstagramID:                  "ig123",
		AudienceAge:                  []string{"13-17:500", "18-24:2000", "25-34:4000", "35-44:2500", "45-54:800", "55-64:150", "65+:50"},
		AudienceGender:               []string{"M:5500", "F:4300", "U:200"},
		AudienceCity:                 []string{"New York, US:1000", "Los Angeles, US:800", "London, UK:600"},
		AudienceCountry:              []string{"US:6000", "UK:2000", "CA:1000", "AU:500"},
		AudienceAgeByEngagement:      []string{"18-24:500", "25-34:1000"},
		AudienceGenderByEngagement:   []string{"M:800", "F:700"},
		AudienceCityByEngagement:     []string{"New York, US:200", "LA, US:150"},
		AudienceCountryByEngagement:  []string{"US:1200", "UK:300"},
		AudienceAgeByReach:           []string{"18-24:10000", "25-34:20000"},
		AudienceGenderByReach:        []string{"M:25000", "F:23000"},
		AudienceCityByReach:          []string{"NYC:5000", "LA:4000"},
		AudienceCountryByReach:       []string{"US:40000", "UK:15000"},
	}

	if len(insight.AudienceAge) != 7 {
		t.Fatalf("expected 7 age groups, got %d", len(insight.AudienceAge))
	}
	if len(insight.AudienceCountry) != 4 {
		t.Fatalf("expected 4 countries, got %d", len(insight.AudienceCountry))
	}
}

func TestInstagramInsight_OnlineFollowers(t *testing.T) {
	insight := InstagramInsight{
		InstagramID: "ig123",
		OnlineFollowers: []string{
			"0:1000", "1:1100", "2:1200", "3:1300", "4:1400", "5:1500",
			"6:1600", "7:1700", "8:1800", "9:2000", "10:2200", "11:2400",
			"12:2500", "13:2600", "14:2700", "15:2800", "16:2900", "17:3000",
			"18:2800", "19:2500", "20:2200", "21:1900", "22:1500", "23:1200",
		},
	}

	if len(insight.OnlineFollowers) != 24 {
		t.Fatalf("expected 24 hours of online followers data, got %d", len(insight.OnlineFollowers))
	}
}

func TestInstagramInsight_Metadata(t *testing.T) {
	insight := InstagramInsight{
		InstagramID: "ig123",
		Metadata: map[string]string{
			"sync_type":     "incremental",
			"last_sync":     "2025-01-15T12:00:00Z",
			"data_version":  "v2",
		},
	}

	if insight.Metadata["sync_type"] != "incremental" {
		t.Fatalf("expected sync_type 'incremental', got %s", insight.Metadata["sync_type"])
	}
	if len(insight.Metadata) != 3 {
		t.Fatalf("expected 3 metadata entries, got %d", len(insight.Metadata))
	}
}

func TestInstagramCompetitorInsights_TableName(t *testing.T) {
	insights := InstagramCompetitorInsights{}
	expected := "instagram_competitor_insights"
	if got := insights.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

func TestInstagramCompetitorPosts_TableName(t *testing.T) {
	posts := InstagramCompetitorPosts{}
	expected := "instagram_competitor_posts"
	if got := posts.TableName(); got != expected {
		t.Errorf("TableName() = %v, want %v", got, expected)
	}
}

func TestInstagramCompetitorInsights_Struct(t *testing.T) {
	now := time.Now()
	insights := InstagramCompetitorInsights{
		RecordID:             "record123",
		InstagramAccountID:   "ig456",
		TotalFollowedByCount: 10000,
		TotalFollowingCount:  500,
		ProfilePictureURL:    "https://example.com/profile.jpg",
		PageName:             "Test Page",
		Metadata:             map[string]string{"key": "value"},
		InsertedAt:           now,
	}

	if insights.RecordID != "record123" {
		t.Fatalf("expected RecordID 'record123', got %s", insights.RecordID)
	}
	if insights.TotalFollowedByCount != 10000 {
		t.Fatalf("expected TotalFollowedByCount 10000, got %d", insights.TotalFollowedByCount)
	}
}

func TestInstagramCompetitorPosts_Struct(t *testing.T) {
	now := time.Now()
	post := InstagramCompetitorPosts{
		InstagramID:          123456,
		PostID:               "post789",
		BusinessAccountID:    "ba123",
		TotalFollowedByCount: 10000,
		TotalFollowingCount:  500,
		Username:             "testuser",
		Name:                 "Test User",
		PageCategory:         "Business",
		ProfilePictureURL:    "https://example.com/profile.jpg",
		Biography:            "Test bio",
		Engagement:           1100,
		LikeCount:            1000,
		CommentsCount:        100,
		MediaCount:           200,
		Caption:              "Test caption",
		MediaType:            "IMAGE",
		MediaProductType:     "FEED",
		MediaURL:             "https://example.com/media.jpg",
		Permalink:            "https://instagram.com/p/abc123",
		Hashtags:             []string{"test", "instagram"},
		CreatedAt:            now,
		InsertedAt:           now,
	}

	if post.PostID != "post789" {
		t.Fatalf("expected PostID 'post789', got %s", post.PostID)
	}
	if post.Engagement != 1100 {
		t.Fatalf("expected Engagement 1100, got %d", post.Engagement)
	}
	if len(post.Hashtags) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(post.Hashtags))
	}
}
