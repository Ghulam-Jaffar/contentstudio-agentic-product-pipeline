package kafka

import (
	"encoding/json"
	"testing"
	"time"
)

func TestFacebookTime_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		name        string
		input       string
		expected    time.Time
		expectError bool
	}{
		{
			name:     "Facebook format with +0000",
			input:    `"2025-03-13T06:37:33+0000"`,
			expected: time.Date(2025, 3, 13, 6, 37, 33, 0, time.UTC),
		},
		{
			name:     "Facebook format with -0000",
			input:    `"2025-03-13T06:37:33-0000"`,
			expected: time.Date(2025, 3, 13, 6, 37, 33, 0, time.UTC),
		},
		{
			name:     "RFC3339 format",
			input:    `"2025-03-13T06:37:33+00:00"`,
			expected: time.Date(2025, 3, 13, 6, 37, 33, 0, time.UTC),
		},
		{
			name:     "RFC3339 with timezone",
			input:    `"2025-03-13T06:37:33-05:00"`,
			expected: time.Date(2025, 3, 13, 11, 37, 33, 0, time.UTC),
		},
		{
			name:     "empty string",
			input:    `""`,
			expected: time.Time{},
		},
		{
			name:     "null value",
			input:    `"null"`,
			expected: time.Time{},
		},
		{
			name:        "invalid format",
			input:       `"not-a-date"`,
			expectError: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var ft FacebookTime
			err := json.Unmarshal([]byte(tc.input), &ft)

			if tc.expectError {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if !ft.Time.Equal(tc.expected) {
				t.Fatalf("expected %v, got %v", tc.expected, ft.Time)
			}
		})
	}
}

func TestFacebookTime_MarshalJSON(t *testing.T) {
	cases := []struct {
		name     string
		input    FacebookTime
		expected string
	}{
		{
			name:     "zero time returns null",
			input:    FacebookTime{},
			expected: "null",
		},
		{
			name:     "non-zero time returns RFC3339",
			input:    FacebookTime{Time: time.Date(2025, 3, 13, 6, 37, 33, 0, time.UTC)},
			expected: `"2025-03-13T06:37:33Z"`,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			data, err := json.Marshal(tc.input)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if string(data) != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, string(data))
			}
		})
	}
}

func TestFacebookTime_RoundTrip(t *testing.T) {
	original := FacebookTime{Time: time.Date(2025, 6, 15, 12, 30, 45, 0, time.UTC)}

	data, err := json.Marshal(original)
	if err != nil {
		t.Fatalf("marshal error: %v", err)
	}

	var result FacebookTime
	err = json.Unmarshal(data, &result)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if !result.Time.Equal(original.Time) {
		t.Fatalf("round trip failed: expected %v, got %v", original.Time, result.Time)
	}
}

func TestRawFacebookPost_Struct(t *testing.T) {
	post := RawFacebookPost{
		ID:           "123456789_987654321",
		Message:      "Test post message",
		PermalinkURL: "https://facebook.com/post/123",
		FullPicture:  "https://example.com/image.jpg",
		StatusType:   "added_photos",
	}

	if post.ID != "123456789_987654321" {
		t.Fatalf("expected ID, got %s", post.ID)
	}
	if post.Message != "Test post message" {
		t.Fatalf("expected Message, got %s", post.Message)
	}
}

func TestRawFacebookPost_WithReactions(t *testing.T) {
	post := RawFacebookPost{
		ID: "test_post",
		Like: &struct {
			Summary *struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		}{
			Summary: &struct {
				TotalCount int `json:"total_count"`
			}{TotalCount: 100},
		},
		Love: &struct {
			Summary *struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		}{
			Summary: &struct {
				TotalCount int `json:"total_count"`
			}{TotalCount: 50},
		},
	}

	if post.Like.Summary.TotalCount != 100 {
		t.Fatalf("expected 100 likes, got %d", post.Like.Summary.TotalCount)
	}
	if post.Love.Summary.TotalCount != 50 {
		t.Fatalf("expected 50 loves, got %d", post.Love.Summary.TotalCount)
	}
}

func TestRawFacebookPost_WithShares(t *testing.T) {
	post := RawFacebookPost{
		ID: "test_post",
		Shares: &struct {
			Count int `json:"count"`
		}{Count: 25},
	}

	if post.Shares.Count != 25 {
		t.Fatalf("expected 25 shares, got %d", post.Shares.Count)
	}
}

func TestRawFacebookPost_WithComments(t *testing.T) {
	post := RawFacebookPost{
		ID: "test_post",
		Comments: &struct {
			Summary *struct {
				TotalCount int  `json:"total_count"`
				CanComment bool `json:"can_comment"`
			} `json:"summary"`
		}{
			Summary: &struct {
				TotalCount int  `json:"total_count"`
				CanComment bool `json:"can_comment"`
			}{
				TotalCount: 15,
				CanComment: true,
			},
		},
	}

	if post.Comments.Summary.TotalCount != 15 {
		t.Fatalf("expected 15 comments, got %d", post.Comments.Summary.TotalCount)
	}
	if !post.Comments.Summary.CanComment {
		t.Fatal("expected CanComment to be true")
	}
}

func TestParsedFacebookPost_Struct(t *testing.T) {
	now := time.Now()
	post := ParsedFacebookPost{
		PageID:          "page123",
		PostID:          "post456",
		MediaType:       "photo",
		Like:            100,
		Love:            50,
		Haha:            25,
		Wow:             10,
		Sad:             5,
		Angry:           2,
		Thankful:        1,
		Total:           193,
		Shares:          30,
		Comments:        45,
		TotalEngagement: 268,
		CreatedTime:     now,
		DayOfWeek:       "Monday",
		HourOfDay:       14,
	}

	if post.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %s", post.PageID)
	}
	if post.Total != 193 {
		t.Fatalf("expected Total 193, got %d", post.Total)
	}
	if post.TotalEngagement != 268 {
		t.Fatalf("expected TotalEngagement 268, got %d", post.TotalEngagement)
	}
}

func TestParsedFacebookMediaAsset_Struct(t *testing.T) {
	now := time.Now()
	asset := ParsedFacebookMediaAsset{
		PageID:       "page123",
		MediaID:      "media456",
		PostID:       "post789",
		AssetType:    "image",
		Link:         "https://example.com/image.jpg",
		CallToAction: "Learn More",
		CTAType:      "LEARN_MORE",
		Caption:      "Test caption",
		CreatedAt:    now,
		InsertedAt:   now,
	}

	if asset.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %s", asset.PageID)
	}
	if asset.AssetType != "image" {
		t.Fatalf("expected AssetType 'image', got %s", asset.AssetType)
	}
}

func TestRawFacebookVideo_Struct(t *testing.T) {
	video := RawFacebookVideo{
		ID:           "video123",
		PostID:       "post456",
		Message:      "Video description",
		Description:  "Video description",
		Picture:      "https://example.com/thumb.jpg",
		PermalinkURL: "https://facebook.com/video/123",
	}

	if video.ID != "video123" {
		t.Fatalf("expected ID 'video123', got %s", video.ID)
	}
	if video.PostID != "post456" {
		t.Fatalf("expected PostID 'post456', got %s", video.PostID)
	}
}

func TestThumbnail_Struct(t *testing.T) {
	thumb := Thumbnail{
		URI:         "https://example.com/thumb.jpg",
		IsPreferred: true,
		ID:          "thumb123",
	}

	if thumb.URI != "https://example.com/thumb.jpg" {
		t.Fatalf("expected URI, got %s", thumb.URI)
	}
	if !thumb.IsPreferred {
		t.Fatal("expected IsPreferred to be true")
	}
}

func TestThumbnailsData_Struct(t *testing.T) {
	data := ThumbnailsData{
		Data: []Thumbnail{
			{URI: "thumb1.jpg", IsPreferred: false},
			{URI: "thumb2.jpg", IsPreferred: true},
		},
	}

	if len(data.Data) != 2 {
		t.Fatalf("expected 2 thumbnails, got %d", len(data.Data))
	}
}

func TestParsedFacebookVideoInsights_Struct(t *testing.T) {
	now := time.Now()
	insights := ParsedFacebookVideoInsights{
		PageID:                  "page123",
		PostID:                  "post456",
		VideoID:                 "video789",
		CreatedTime:             now,
		TotalEngagement:         1000,
		TotalVideoViews:         5000,
		TotalVideoViewsUnique:   4500,
		TotalVideoCompleteViews: 2000,
		TotalVideoImpressions:   10000,
	}

	if insights.PageID != "page123" {
		t.Fatalf("expected PageID 'page123', got %s", insights.PageID)
	}
	if insights.TotalVideoViews != 5000 {
		t.Fatalf("expected TotalVideoViews 5000, got %d", insights.TotalVideoViews)
	}
}

func TestParsedFacebookReelsInsights_Struct(t *testing.T) {
	now := time.Now()
	reels := ParsedFacebookReelsInsights{
		PageID:               "page123",
		PostID:               "post456",
		AverageTimeWatched:   15000,
		TotalTimeWatchedInMs: 500000,
		PlayCount:            1000,
		ImpressionsUnique:    800,
		ReelFollowers:        50,
		CreatedAt:            now,
		SavingTime:           now,
	}

	if reels.PlayCount != 1000 {
		t.Fatalf("expected PlayCount 1000, got %d", reels.PlayCount)
	}
}

func TestRawFacebookInsights_Struct(t *testing.T) {
	now := time.Now()
	insights := RawFacebookInsights{
		PageID:      "page123",
		WorkspaceID: "workspace456",
		SavingTime:  now,
		Data: []FacebookInsightData{
			{
				Name:   "page_impressions",
				Period: "day",
				Values: []FacebookInsightValue{
					{Value: 1000, EndTime: "2025-01-01T00:00:00+0000"},
				},
			},
		},
	}

	if len(insights.Data) != 1 {
		t.Fatalf("expected 1 insight data, got %d", len(insights.Data))
	}
	if insights.Data[0].Name != "page_impressions" {
		t.Fatalf("expected metric name 'page_impressions', got %s", insights.Data[0].Name)
	}
}

func TestParsedFacebookInsights_Struct(t *testing.T) {
	now := time.Now()
	insights := ParsedFacebookInsights{
		HashID:            "hash123",
		PageID:            "page456",
		WorkspaceID:       "workspace789",
		PageFans:          10000,
		PageFollows:       9500,
		PageViews:         50000,
		PageImpressions:   100000,
		PageVideoViews:    25000,
		CreatedTime:       now,
		SavingTime:        now,
		Year:              2025,
		Month:             1,
		DayOfWeek:         "Monday",
		PageFansCity:      []string{"New York:1000", "Los Angeles:800"},
		PageFansCountry:   []string{"US:5000", "UK:2000"},
		PageFansAge:       []string{"25-34:3000", "35-44:2500"},
		PageFansGenderAge: []string{"M.25-34:1500", "F.25-34:1500"},
	}

	if insights.PageFans != 10000 {
		t.Fatalf("expected PageFans 10000, got %d", insights.PageFans)
	}
	if len(insights.PageFansCity) != 2 {
		t.Fatalf("expected 2 city entries, got %d", len(insights.PageFansCity))
	}
}

func TestFacebookAccountWorkOrder_Struct(t *testing.T) {
	wo := FacebookAccountWorkOrder{
		ID:              "acc123",
		FacebookID:      "fb456",
		Type:            "Page",
		AccessToken:     "token789",
		WorkspaceID:     "ws012",
		LongAccessToken: "long_token",
		SyncType:        "incremental",
	}

	if wo.FacebookID != "fb456" {
		t.Fatalf("expected FacebookID 'fb456', got %s", wo.FacebookID)
	}
	if wo.SyncType != "incremental" {
		t.Fatalf("expected SyncType 'incremental', got %s", wo.SyncType)
	}
}

func TestFacebookBatchWorkOrder_Struct(t *testing.T) {
	now := time.Now()
	batch := FacebookBatchWorkOrder{
		BatchID:   "batch123",
		SyncType:  "full_sync",
		CreatedAt: now,
		Accounts: []FacebookAccountWorkOrder{
			{ID: "acc1", FacebookID: "fb1"},
			{ID: "acc2", FacebookID: "fb2"},
		},
	}

	if batch.BatchID != "batch123" {
		t.Fatalf("expected BatchID 'batch123', got %s", batch.BatchID)
	}
	if len(batch.Accounts) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(batch.Accounts))
	}
}

func TestFacebookInsightData_Struct(t *testing.T) {
	data := FacebookInsightData{
		Name:        "page_impressions_unique",
		Period:      "day",
		Title:       "Daily Page Reach",
		Description: "The number of unique people who saw any content",
		ID:          "insight123",
		Values: []FacebookInsightValue{
			{Value: 1000, EndTime: "2025-01-01T00:00:00+0000"},
			{Value: 1200, EndTime: "2025-01-02T00:00:00+0000"},
		},
	}

	if data.Name != "page_impressions_unique" {
		t.Fatalf("expected Name 'page_impressions_unique', got %s", data.Name)
	}
	if len(data.Values) != 2 {
		t.Fatalf("expected 2 values, got %d", len(data.Values))
	}
}

func TestFacebookInsightValue_Struct(t *testing.T) {
	value := FacebookInsightValue{
		Value:   5000,
		EndTime: "2025-01-15T00:00:00+0000",
	}

	if value.Value != 5000 {
		t.Fatalf("expected Value 5000, got %v", value.Value)
	}
}

func TestRawFacebookPost_JSON_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "123456789_987654321",
		"message": "Test post",
		"created_time": "2025-03-13T06:37:33+0000",
		"permalink_url": "https://facebook.com/post/123",
		"full_picture": "https://example.com/image.jpg"
	}`

	var post RawFacebookPost
	err := json.Unmarshal([]byte(jsonData), &post)
	if err != nil {
		t.Fatalf("unmarshal error: %v", err)
	}

	if post.ID != "123456789_987654321" {
		t.Fatalf("expected ID, got %s", post.ID)
	}
	if post.CreatedTime.Time.IsZero() {
		t.Fatal("expected non-zero CreatedTime")
	}
}
