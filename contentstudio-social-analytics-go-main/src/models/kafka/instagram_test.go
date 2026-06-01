package kafka

import (
	"encoding/json"
	"testing"
	"time"
)

func TestInstagramChild_Struct(t *testing.T) {
	child := InstagramChild{
		MediaType:    "IMAGE",
		MediaURL:     "https://example.com/image.jpg",
		ThumbnailURL: "https://example.com/thumb.jpg",
	}

	if child.MediaType != "IMAGE" {
		t.Fatalf("expected MediaType 'IMAGE', got %s", child.MediaType)
	}
	if child.MediaURL != "https://example.com/image.jpg" {
		t.Fatalf("expected MediaURL, got %s", child.MediaURL)
	}
}

func TestRawInstagramMedia_Struct(t *testing.T) {
	media := RawInstagramMedia{
		ID:               "media123",
		CommentsCount:    50,
		ThumbnailURL:     "https://example.com/thumb.jpg",
		Caption:          "Test caption #hashtag",
		Username:         "testuser",
		LikeCount:        100,
		MediaType:        "IMAGE",
		MediaProductType: "FEED",
		MediaURL:         "https://example.com/image.jpg",
		Timestamp:        "2025-01-15T12:00:00+0000",
		Permalink:        "https://instagram.com/p/abc123",
	}

	if media.ID != "media123" {
		t.Fatalf("expected ID 'media123', got %s", media.ID)
	}
	if media.LikeCount != 100 {
		t.Fatalf("expected LikeCount 100, got %d", media.LikeCount)
	}
	if media.MediaType != "IMAGE" {
		t.Fatalf("expected MediaType 'IMAGE', got %s", media.MediaType)
	}
}

func TestRawInstagramMedia_WithChildren(t *testing.T) {
	media := RawInstagramMedia{
		ID:        "carousel123",
		MediaType: "CAROUSEL_ALBUM",
	}
	media.Children.Data = []InstagramChild{
		{MediaType: "IMAGE", MediaURL: "img1.jpg"},
		{MediaType: "IMAGE", MediaURL: "img2.jpg"},
		{MediaType: "VIDEO", MediaURL: "vid1.mp4"},
	}

	if len(media.Children.Data) != 3 {
		t.Fatalf("expected 3 children, got %d", len(media.Children.Data))
	}
}

func TestRawInstagramAccountResponse_Struct(t *testing.T) {
	response := RawInstagramAccountResponse{
		Name:              "Test Account",
		ProfilePictureURL: "https://example.com/profile.jpg",
	}
	response.Media.Data = []RawInstagramMedia{
		{ID: "media1", MediaType: "IMAGE"},
		{ID: "media2", MediaType: "VIDEO"},
	}
	response.Media.Paging.Next = "https://api.instagram.com/next"

	if response.Name != "Test Account" {
		t.Fatalf("expected Name 'Test Account', got %s", response.Name)
	}
	if len(response.Media.Data) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(response.Media.Data))
	}
}

func TestInstagramInsightValue_Struct(t *testing.T) {
	value := InstagramInsightValue{
		Value:   1000,
		EndTime: "2025-01-15T00:00:00+0000",
	}

	if value.Value != 1000 {
		t.Fatalf("expected Value 1000, got %d", value.Value)
	}
}

func TestInstagramInsightBreakdownResult_Struct(t *testing.T) {
	result := InstagramInsightBreakdownResult{
		DimensionValues: []string{"US", "18-24"},
		Value:           500,
	}

	if len(result.DimensionValues) != 2 {
		t.Fatalf("expected 2 dimension values, got %d", len(result.DimensionValues))
	}
	if result.Value != 500 {
		t.Fatalf("expected Value 500, got %d", result.Value)
	}
}

func TestInstagramInsightBreakdown_Struct(t *testing.T) {
	breakdown := InstagramInsightBreakdown{
		Results: []InstagramInsightBreakdownResult{
			{DimensionValues: []string{"US"}, Value: 1000},
			{DimensionValues: []string{"UK"}, Value: 500},
		},
	}

	if len(breakdown.Results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(breakdown.Results))
	}
}

func TestRawInstagramInsightData_Struct(t *testing.T) {
	data := RawInstagramInsightData{
		Name:   "impressions",
		Period: "day",
		Values: []InstagramInsightValue{
			{Value: 1000, EndTime: "2025-01-15"},
			{Value: 1200, EndTime: "2025-01-16"},
		},
	}

	if data.Name != "impressions" {
		t.Fatalf("expected Name 'impressions', got %s", data.Name)
	}
	if len(data.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(data.Values))
	}
}

func TestRawInstagramInsightData_WithTotalValue(t *testing.T) {
	data := RawInstagramInsightData{
		Name:   "audience_country",
		Period: "lifetime",
	}
	data.TotalValue.Value = 5000
	data.TotalValue.Breakdowns = []InstagramInsightBreakdown{
		{
			Results: []InstagramInsightBreakdownResult{
				{DimensionValues: []string{"US"}, Value: 3000},
				{DimensionValues: []string{"UK"}, Value: 2000},
			},
		},
	}

	if data.TotalValue.Value != 5000 {
		t.Fatalf("expected TotalValue 5000, got %d", data.TotalValue.Value)
	}
}

func TestRawInstagramInsightsResponse_Struct(t *testing.T) {
	response := RawInstagramInsightsResponse{
		Data: []RawInstagramInsightData{
			{Name: "impressions", Period: "day"},
			{Name: "reach", Period: "day"},
		},
	}

	if len(response.Data) != 2 {
		t.Fatalf("expected 2 data items, got %d", len(response.Data))
	}
}

func TestRawInstagramMediaInsights_Struct(t *testing.T) {
	insights := RawInstagramMediaInsights{}
	insights.Data = []struct {
		Name   string `json:"name"`
		Period string `json:"period"`
		Values []struct {
			Value interface{} `json:"value"`
		} `json:"values"`
		Title       string `json:"title"`
		Description string `json:"description"`
		ID          string `json:"id"`
	}{
		{
			Name:   "impressions",
			Period: "lifetime",
			Values: []struct {
				Value interface{} `json:"value"`
			}{{Value: 1000}},
			Title: "Impressions",
		},
	}

	if len(insights.Data) != 1 {
		t.Fatalf("expected 1 data item, got %d", len(insights.Data))
	}
}

func TestRawInstagramDemographics_Struct(t *testing.T) {
	demo := RawInstagramDemographics{}
	demo.Data = []struct {
		Name       string `json:"name"`
		Period     string `json:"period"`
		TotalValue struct {
			Value      interface{} `json:"value"`
			Breakdowns []struct {
				Results []struct {
					DimensionValues []string `json:"dimension_values"`
					Value           int      `json:"value"`
				} `json:"results"`
			} `json:"breakdowns,omitempty"`
		} `json:"total_value"`
	}{
		{
			Name:   "audience_country",
			Period: "lifetime",
		},
	}

	if len(demo.Data) != 1 {
		t.Fatalf("expected 1 demographic data item, got %d", len(demo.Data))
	}
}

func TestParsedInstagramPost_Struct(t *testing.T) {
	now := time.Now()
	post := ParsedInstagramPost{
		InstagramID:         "ig123",
		MediaID:             "media456",
		Username:            "testuser",
		Name:                "Test User",
		ProfilePictureURL:   "https://example.com/profile.jpg",
		Permalink:           "https://instagram.com/p/abc",
		LikeCount:           100,
		CommentsCount:       25,
		Engagement:          125,
		Impressions:         1000,
		Views:               500,
		Reach:               800,
		Saved:               50,
		VideoViews:          500,
		Shares:              10,
		ReelsAvgWatchTime:   5000,
		ReelsTotalWatchTime: 25000,
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

	if post.InstagramID != "ig123" {
		t.Fatalf("expected InstagramID 'ig123', got %s", post.InstagramID)
	}
	if post.Engagement != 125 {
		t.Fatalf("expected Engagement 125, got %d", post.Engagement)
	}
	if len(post.Hashtags) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(post.Hashtags))
	}
}

func TestParsedInstagramInsight_Struct(t *testing.T) {
	now := time.Now()
	insight := ParsedInstagramInsight{
		InstagramID:       "ig123",
		RecordID:          "rec456",
		Name:              "Test Account",
		Username:          "testuser",
		ProfilePictureURL: "https://example.com/profile.jpg",
		FollowsCount:      100,
		FollowersCount:    10000,
		FollowerCount:     10000,
		MediaCount:        500,
		Tags:              50,
		Impressions:       50000,
		ProfileViews:      5000,
		Shares:            200,
		AccountsEngaged:   3000,
		Engagement:        8000,
		Reach:             40000,
		Views:             20000,
		Saves:             1000,
		Likes:             15000,
		Comments:          2000,
		AudienceAge:       []string{"18-24:2000", "25-34:4000"},
		AudienceGender:    []string{"M:5500", "F:4500"},
		AudienceCity:      []string{"New York:1000", "Los Angeles:800"},
		AudienceCountry:   []string{"US:6000", "UK:2000"},
		OnlineFollowers:   []string{"0:100", "1:150", "2:200"},
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
	if len(insight.AudienceAge) != 2 {
		t.Fatalf("expected 2 audience age entries, got %d", len(insight.AudienceAge))
	}
}

func TestInstagramAccountWorkOrder_Struct(t *testing.T) {
	wo := InstagramAccountWorkOrder{
		ID:                    "acc123",
		InstagramID:           "ig456",
		AccessToken:           "token789",
		WorkspaceID:           "ws012",
		SyncType:              "incremental",
		ConnectedViaInstagram: true,
	}

	if wo.InstagramID != "ig456" {
		t.Fatalf("expected InstagramID 'ig456', got %s", wo.InstagramID)
	}
	if !wo.ConnectedViaInstagram {
		t.Fatal("expected ConnectedViaInstagram to be true")
	}
}

func TestInstagramBatchWorkOrder_Struct(t *testing.T) {
	now := time.Now()
	batch := InstagramBatchWorkOrder{
		BatchID:   "batch123",
		SyncType:  "full_sync",
		CreatedAt: now,
		Accounts: []InstagramAccountWorkOrder{
			{ID: "acc1", InstagramID: "ig1"},
			{ID: "acc2", InstagramID: "ig2"},
			{ID: "acc3", InstagramID: "ig3"},
		},
	}

	if batch.BatchID != "batch123" {
		t.Fatalf("expected BatchID 'batch123', got %s", batch.BatchID)
	}
	if len(batch.Accounts) != 3 {
		t.Fatalf("expected 3 accounts, got %d", len(batch.Accounts))
	}
}

func TestRawInstagramMedia_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "media123",
		"username": "testuser",
		"caption": "Test post #hashtag",
		"media_type": "IMAGE",
		"media_url": "https://example.com/image.jpg",
		"timestamp": "2025-01-15T12:00:00+0000",
		"permalink": "https://instagram.com/p/abc123",
		"like_count": 100,
		"comments_count": 25
	}`

	var media RawInstagramMedia
	err := json.Unmarshal([]byte(jsonData), &media)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if media.ID != "media123" {
		t.Fatalf("expected ID 'media123', got %s", media.ID)
	}
	if media.LikeCount != 100 {
		t.Fatalf("expected LikeCount 100, got %d", media.LikeCount)
	}
}

func TestInstagramAccountWorkOrder_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "acc123",
		"instagram_id": "ig456",
		"access_token": "token789",
		"workspace_id": "ws012",
		"sync_type": "incremental",
		"connected_via_instagram": true
	}`

	var wo InstagramAccountWorkOrder
	err := json.Unmarshal([]byte(jsonData), &wo)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if wo.ID != "acc123" {
		t.Fatalf("expected ID 'acc123', got %s", wo.ID)
	}
	if !wo.ConnectedViaInstagram {
		t.Fatal("expected ConnectedViaInstagram to be true")
	}
}
