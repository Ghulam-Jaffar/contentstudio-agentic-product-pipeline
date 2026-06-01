package parsing

import (
	"testing"
	"time"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestNewInstagramParser(t *testing.T) {
	parser := NewInstagramParser()

	if parser == nil {
		t.Fatal("expected non-nil parser")
	}

	if parser.hashtagRegex == nil {
		t.Fatal("expected hashtagRegex to be initialized")
	}
}

func TestInstagramParser_getInt64Value(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{
			name:     "nil returns 0",
			input:    nil,
			expected: 0,
		},
		{
			name:     "int64 returns itself",
			input:    int64(100),
			expected: 100,
		},
		{
			name:     "int returns int64",
			input:    42,
			expected: 42,
		},
		{
			name:     "float64 returns truncated int64",
			input:    123.456,
			expected: 123,
		},
		{
			name:     "negative number",
			input:    int64(-50),
			expected: -50,
		},
		{
			name:     "zero",
			input:    0,
			expected: 0,
		},
		{
			name:     "large number",
			input:    int64(9223372036854775807),
			expected: 9223372036854775807,
		},
		{
			name:     "unsupported type returns 0",
			input:    "string",
			expected: 0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.getInt64Value(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestInstagramParser_ParseMedia(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name              string
		media             kafkamodels.RawInstagramMedia
		instagramID       string
		accountName       string
		username          string
		profilePictureURL string
		checkFunc         func(*kafkamodels.ParsedInstagramPost) bool
	}{
		{
			name: "parses basic image post",
			media: kafkamodels.RawInstagramMedia{
				ID:               "media123",
				Username:         "testuser",
				MediaType:        "IMAGE",
				MediaProductType: "FEED",
				Caption:          "Test caption",
				Permalink:        "https://instagram.com/p/abc123",
				LikeCount:        100,
				CommentsCount:    50,
				Timestamp:        "2024-01-15T10:30:00+0000",
			},
			instagramID:       "insta123",
			accountName:       "Test Account",
			username:          "testuser",
			profilePictureURL: "https://example.com/pic.jpg",
			checkFunc: func(p *kafkamodels.ParsedInstagramPost) bool {
				return p.MediaID == "media123" &&
					p.InstagramID == "insta123" &&
					p.Username == "testuser" &&
					p.Name == "Test Account" &&
					p.MediaType == "IMAGE" &&
					p.LikeCount == 100 &&
					p.CommentsCount == 50
			},
		},
		{
			name: "parses video post",
			media: kafkamodels.RawInstagramMedia{
				ID:               "video123",
				MediaType:        "VIDEO",
				MediaProductType: "FEED",
				Timestamp:        "2024-01-15T10:30:00+0000",
			},
			instagramID: "insta123",
			checkFunc: func(p *kafkamodels.ParsedInstagramPost) bool {
				return p.MediaType == "VIDEO"
			},
		},
		{
			name: "parses reels post",
			media: kafkamodels.RawInstagramMedia{
				ID:               "reel123",
				MediaType:        "VIDEO",
				MediaProductType: "REELS",
				Timestamp:        "2024-01-15T10:30:00+0000",
			},
			instagramID: "insta123",
			checkFunc: func(p *kafkamodels.ParsedInstagramPost) bool {
				return p.MediaType == "REELS"
			},
		},
		{
			name: "parses carousel album",
			media: kafkamodels.RawInstagramMedia{
				ID:               "carousel123",
				MediaType:        "CAROUSEL_ALBUM",
				MediaProductType: "FEED",
				Timestamp:        "2024-01-15T10:30:00+0000",
			},
			instagramID: "insta123",
			checkFunc: func(p *kafkamodels.ParsedInstagramPost) bool {
				return p.MediaType == "CAROUSEL_ALBUM"
			},
		},
		{
			name: "extracts hashtags from caption",
			media: kafkamodels.RawInstagramMedia{
				ID:        "media123",
				MediaType: "IMAGE",
				Caption:   "Check out #summer #vacation #travel",
				Timestamp: "2024-01-15T10:30:00+0000",
			},
			instagramID: "insta123",
			checkFunc: func(p *kafkamodels.ParsedInstagramPost) bool {
				return len(p.Hashtags) == 3 &&
					contains(p.Hashtags, "summer") &&
					contains(p.Hashtags, "vacation") &&
					contains(p.Hashtags, "travel")
			},
		},
		{
			name: "handles RFC3339 timestamp",
			media: kafkamodels.RawInstagramMedia{
				ID:        "media123",
				MediaType: "IMAGE",
				Timestamp: "2024-01-15T10:30:00Z",
			},
			instagramID: "insta123",
			checkFunc: func(p *kafkamodels.ParsedInstagramPost) bool {
				return !p.PostCreatedAt.IsZero()
			},
		},
		{
			name: "handles invalid timestamp",
			media: kafkamodels.RawInstagramMedia{
				ID:        "media123",
				MediaType: "IMAGE",
				Timestamp: "invalid-timestamp",
			},
			instagramID: "insta123",
			checkFunc: func(p *kafkamodels.ParsedInstagramPost) bool {
				return p.PostCreatedAt.IsZero()
			},
		},
		{
			name: "calculates engagement",
			media: kafkamodels.RawInstagramMedia{
				ID:            "media123",
				MediaType:     "IMAGE",
				LikeCount:     100,
				CommentsCount: 50,
				Timestamp:     "2024-01-15T10:30:00+0000",
			},
			instagramID: "insta123",
			checkFunc: func(p *kafkamodels.ParsedInstagramPost) bool {
				return p.Engagement == 150
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.ParseMedia(tc.media, tc.instagramID, tc.accountName, tc.username, tc.profilePictureURL)

			if result == nil {
				t.Fatal("expected non-nil result")
			}

			if !tc.checkFunc(result) {
				t.Fatal("check function failed")
			}
		})
	}
}

func TestInstagramParser_ParseMedia_ChildAssets(t *testing.T) {
	parser := NewInstagramParser()

	media := kafkamodels.RawInstagramMedia{
		ID:        "carousel123",
		MediaType: "CAROUSEL_ALBUM",
		MediaURL:  "https://example.com/main.jpg",
		Timestamp: "2024-01-15T10:30:00+0000",
		Children: struct {
			Data []kafkamodels.InstagramChild `json:"data"`
		}{
			Data: []kafkamodels.InstagramChild{
				{MediaType: "IMAGE", MediaURL: "https://example.com/child1.jpg"},
				{MediaType: "VIDEO", MediaURL: "https://example.com/child2.mp4", ThumbnailURL: "https://example.com/thumb2.jpg"},
			},
		},
	}

	result := parser.ParseMedia(media, "insta123", "Account", "user", "https://pic.jpg")

	if len(result.ChildAssetsType) != 2 {
		t.Fatalf("expected 2 child assets, got %d", len(result.ChildAssetsType))
	}

	if result.ChildAssetsType[0] != "IMAGE" {
		t.Fatalf("expected first child type IMAGE, got %s", result.ChildAssetsType[0])
	}

	if result.ChildAssetsType[1] != "VIDEO" {
		t.Fatalf("expected second child type VIDEO, got %s", result.ChildAssetsType[1])
	}
}

func TestInstagramParser_ParseMediaWithInsights(t *testing.T) {
	parser := NewInstagramParser()

	enrichedData := map[string]interface{}{
		"id":            "media123",
		"media_type":    "IMAGE",
		"timestamp":     "2024-01-15T10:30:00+0000",
		"like_count":    float64(100),
		"comments_count": float64(50),
		"user_info": map[string]interface{}{
			"name":                "Test Account",
			"username":            "testuser",
			"profile_picture_url": "https://example.com/pic.jpg",
		},
		"insights": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"name": "impressions",
					"values": []interface{}{
						map[string]interface{}{"value": float64(1000)},
					},
				},
				map[string]interface{}{
					"name": "reach",
					"values": []interface{}{
						map[string]interface{}{"value": float64(800)},
					},
				},
				map[string]interface{}{
					"name": "saved",
					"values": []interface{}{
						map[string]interface{}{"value": float64(25)},
					},
				},
			},
		},
	}

	result, err := parser.ParseMediaWithInsights(enrichedData, "insta123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.Impressions != 1000 {
		t.Fatalf("expected Impressions 1000, got %d", result.Impressions)
	}

	if result.Reach != 800 {
		t.Fatalf("expected Reach 800, got %d", result.Reach)
	}

	if result.Saved != 25 {
		t.Fatalf("expected Saved 25, got %d", result.Saved)
	}
}

func TestInstagramParser_ParseMediaWithInsights_Navigation(t *testing.T) {
	parser := NewInstagramParser()

	enrichedData := map[string]interface{}{
		"id":         "story123",
		"media_type": "IMAGE",
		"timestamp":  "2024-01-15T10:30:00+0000",
		"insights": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"name": "navigation",
					"total_value": map[string]interface{}{
						"breakdowns": []interface{}{
							map[string]interface{}{
								"results": []interface{}{
									map[string]interface{}{
										"dimension_values": []interface{}{"tap_back"},
										"value":            float64(10),
									},
									map[string]interface{}{
										"dimension_values": []interface{}{"tap_forward"},
										"value":            float64(50),
									},
									map[string]interface{}{
										"dimension_values": []interface{}{"tap_exit"},
										"value":            float64(5),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := parser.ParseMediaWithInsights(enrichedData, "insta123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.TapsBack != 10 {
		t.Fatalf("expected TapsBack 10, got %d", result.TapsBack)
	}

	if result.TapsForward != 50 {
		t.Fatalf("expected TapsForward 50, got %d", result.TapsForward)
	}

	if result.Exits != 5 {
		t.Fatalf("expected Exits 5, got %d", result.Exits)
	}
}

func TestInstagramParser_applyMediaMetric(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name       string
		metricName string
		value      int64
		checkFunc  func(*kafkamodels.ParsedInstagramPost) bool
	}{
		{
			name:       "impressions",
			metricName: "impressions",
			value:      1000,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.Impressions == 1000 },
		},
		{
			name:       "reach",
			metricName: "reach",
			value:      800,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.Reach == 800 },
		},
		{
			name:       "saved",
			metricName: "saved",
			value:      100,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.Saved == 100 },
		},
		{
			name:       "video_views",
			metricName: "video_views",
			value:      5000,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.VideoViews == 5000 && p.Views == 5000 },
		},
		{
			name:       "views",
			metricName: "views",
			value:      3000,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.Views == 3000 },
		},
		{
			name:       "exits",
			metricName: "exits",
			value:      50,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.Exits == 50 },
		},
		{
			name:       "replies",
			metricName: "replies",
			value:      25,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.Replies == 25 },
		},
		{
			name:       "taps_forward",
			metricName: "taps_forward",
			value:      75,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.TapsForward == 75 },
		},
		{
			name:       "taps_back",
			metricName: "taps_back",
			value:      30,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.TapsBack == 30 },
		},
		{
			name:       "shares",
			metricName: "shares",
			value:      200,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.Shares == 200 },
		},
		{
			name:       "ig_reels_avg_watch_time",
			metricName: "ig_reels_avg_watch_time",
			value:      15000,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.ReelsAvgWatchTime == 15000 },
		},
		{
			name:       "ig_reels_video_view_total_time",
			metricName: "ig_reels_video_view_total_time",
			value:      500000,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.ReelsTotalWatchTime == 500000 },
		},
		{
			name:       "total_interactions",
			metricName: "total_interactions",
			value:      250,
			checkFunc:  func(p *kafkamodels.ParsedInstagramPost) bool { return p.Engagement == 250 },
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parsed := &kafkamodels.ParsedInstagramPost{}
			parser.applyMediaMetric(parsed, tc.metricName, tc.value)

			if !tc.checkFunc(parsed) {
				t.Fatal("check function failed")
			}
		})
	}
}

func TestInstagramParser_ParseInsights(t *testing.T) {
	parser := NewInstagramParser()

	raw := &kafkamodels.RawInstagramInsightsResponse{
		Data: []kafkamodels.RawInstagramInsightData{
			{
				Name: "impressions",
				Values: []kafkamodels.InstagramInsightValue{
					{Value: 1000},
				},
			},
			{
				Name: "reach",
				Values: []kafkamodels.InstagramInsightValue{
					{Value: 800},
				},
			},
			{
				Name: "profile_views",
				Values: []kafkamodels.InstagramInsightValue{
					{Value: 50},
				},
			},
		},
	}

	result, err := parser.ParseInsights(raw, "insta123", "testuser", "Test Account", "https://pic.jpg", "record123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.InstagramID != "insta123" {
		t.Fatalf("expected InstagramID 'insta123', got %q", result.InstagramID)
	}

	if result.Impressions != 1000 {
		t.Fatalf("expected Impressions 1000, got %d", result.Impressions)
	}

	if result.Reach != 800 {
		t.Fatalf("expected Reach 800, got %d", result.Reach)
	}

	if result.ProfileViews != 50 {
		t.Fatalf("expected ProfileViews 50, got %d", result.ProfileViews)
	}
}

func TestInstagramParser_ParseInsights_NilInput(t *testing.T) {
	parser := NewInstagramParser()

	result, err := parser.ParseInsights(nil, "insta123", "user", "name", "pic", "record")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Fatal("expected nil result for nil input")
	}
}

func TestInstagramParser_ParseInsights_EmptyData(t *testing.T) {
	parser := NewInstagramParser()

	raw := &kafkamodels.RawInstagramInsightsResponse{
		Data: []kafkamodels.RawInstagramInsightData{},
	}

	result, err := parser.ParseInsights(raw, "insta123", "user", "name", "pic", "record")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Fatal("expected nil result for empty data")
	}
}

func TestInstagramParser_ParseInsights_TotalValue(t *testing.T) {
	parser := NewInstagramParser()

	raw := &kafkamodels.RawInstagramInsightsResponse{
		Data: []kafkamodels.RawInstagramInsightData{
			{
				Name: "impressions",
				TotalValue: struct {
					Value      int                                  `json:"value"`
					Breakdowns []kafkamodels.InstagramInsightBreakdown `json:"breakdowns,omitempty"`
				}{
					Value: 5000,
				},
			},
		},
	}

	result, err := parser.ParseInsights(raw, "insta123", "user", "name", "pic", "record")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Impressions != 5000 {
		t.Fatalf("expected Impressions 5000, got %d", result.Impressions)
	}
}

func TestInstagramParser_applyMetric(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name       string
		metricName string
		value      interface{}
		checkFunc  func(*kafkamodels.ParsedInstagramInsight) bool
	}{
		{
			name:       "follows",
			metricName: "follows",
			value:      int64(1000),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.FollowsCount == 1000 },
		},
		{
			name:       "follower_count",
			metricName: "follower_count",
			value:      int64(5000),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.FollowerCount == 5000 },
		},
		{
			name:       "media_count",
			metricName: "media_count",
			value:      int64(100),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.MediaCount == 100 },
		},
		{
			name:       "impressions",
			metricName: "impressions",
			value:      int64(10000),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.Impressions == 10000 },
		},
		{
			name:       "profile_views",
			metricName: "profile_views",
			value:      int64(500),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.ProfileViews == 500 },
		},
		{
			name:       "reach",
			metricName: "reach",
			value:      int64(8000),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.Reach == 8000 },
		},
		{
			name:       "likes",
			metricName: "likes",
			value:      int64(200),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.Likes == 200 },
		},
		{
			name:       "comments",
			metricName: "comments",
			value:      int64(50),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.Comments == 50 },
		},
		{
			name:       "saves",
			metricName: "saves",
			value:      int64(75),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.Saves == 75 },
		},
		{
			name:       "shares",
			metricName: "shares",
			value:      int64(30),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.Shares == 30 },
		},
		{
			name:       "views",
			metricName: "views",
			value:      int64(15000),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.Views == 15000 },
		},
		{
			name:       "accounts_engaged",
			metricName: "accounts_engaged",
			value:      int64(300),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.AccountsEngaged == 300 },
		},
		{
			name:       "tags",
			metricName: "tags",
			value:      int64(10),
			checkFunc:  func(p *kafkamodels.ParsedInstagramInsight) bool { return p.Tags == 10 },
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parsed := &kafkamodels.ParsedInstagramInsight{}
			parser.applyMetric(parsed, tc.metricName, tc.value)

			if !tc.checkFunc(parsed) {
				t.Fatal("check function failed")
			}
		})
	}
}

func TestInstagramParser_ParseInsightsWithDemographics(t *testing.T) {
	parser := NewInstagramParser()

	enrichedData := map[string]interface{}{
		"user_info": map[string]interface{}{
			"name":                "Test Account",
			"username":            "testuser",
			"profile_picture_url": "https://example.com/pic.jpg",
			"followers_count":     float64(10000),
			"follows_count":       float64(500),
			"media_count":         float64(100),
		},
		"insights": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"name": "impressions",
					"values": []interface{}{
						map[string]interface{}{"value": float64(5000)},
					},
				},
			},
		},
	}

	result, err := parser.ParseInsightsWithDemographics(enrichedData, "insta123", "record123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.InstagramID != "insta123" {
		t.Fatalf("expected InstagramID 'insta123', got %q", result.InstagramID)
	}

	if result.Name != "Test Account" {
		t.Fatalf("expected Name 'Test Account', got %q", result.Name)
	}

	if result.FollowersCount != 10000 {
		t.Fatalf("expected FollowersCount 10000, got %d", result.FollowersCount)
	}
}

func TestInstagramParser_ParseInsightsWithDemographics_OnlineFollowers(t *testing.T) {
	parser := NewInstagramParser()

	enrichedData := map[string]interface{}{
		"demographics": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"name": "online_followers",
					"total_value": map[string]interface{}{
						"value": map[string]interface{}{
							"0":  float64(100),
							"12": float64(500),
							"23": float64(200),
						},
					},
				},
			},
		},
	}

	result, err := parser.ParseInsightsWithDemographics(enrichedData, "insta123", "record123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.OnlineFollowers) == 0 {
		t.Fatal("expected OnlineFollowers to be populated")
	}
}

func TestInstagramParser_parseEndTime(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name      string
		endTime   string
		expectErr bool
	}{
		{
			name:      "parses Facebook format",
			endTime:   "2024-01-15T08:00:00+0000",
			expectErr: false,
		},
		{
			name:      "parses RFC3339",
			endTime:   "2024-01-15T08:00:00Z",
			expectErr: false,
		},
		{
			name:      "fails on invalid format",
			endTime:   "invalid-date",
			expectErr: true,
		},
		{
			name:      "empty string fails",
			endTime:   "",
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := parser.parseEndTime(tc.endTime)

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.IsZero() {
				t.Fatal("expected non-zero time")
			}
		})
	}
}

func TestInstagramParser_getValueForDate(t *testing.T) {
	parser := NewInstagramParser()

	targetDate := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)

	cases := []struct {
		name     string
		values   []interface{}
		expected int64
	}{
		{
			name:     "empty values returns 0",
			values:   []interface{}{},
			expected: 0,
		},
		{
			name: "finds matching date",
			values: []interface{}{
				map[string]interface{}{
					"end_time": "2024-01-15T08:00:00+0000",
					"value":    float64(100),
				},
			},
			expected: 100,
		},
		{
			name: "no matching date returns 0",
			values: []interface{}{
				map[string]interface{}{
					"end_time": "2024-01-16T08:00:00+0000",
					"value":    float64(100),
				},
			},
			expected: 0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.getValueForDate(tc.values, targetDate)
			if result != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestInstagramParser_ParseInsightsDaily(t *testing.T) {
	parser := NewInstagramParser()

	enrichedData := map[string]interface{}{
		"user_info": map[string]interface{}{
			"name":     "Test Account",
			"username": "testuser",
		},
		"insights": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"name": "reach",
					"values": []interface{}{
						map[string]interface{}{
							"end_time": "2024-01-15T08:00:00+0000",
							"value":    float64(1000),
						},
						map[string]interface{}{
							"end_time": "2024-01-16T08:00:00+0000",
							"value":    float64(1100),
						},
					},
				},
			},
		},
	}

	results, err := parser.ParseInsightsDaily(enrichedData, "insta123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 daily records, got %d", len(results))
	}
}

func TestInstagramParser_ParseInsightsDaily_NoData(t *testing.T) {
	parser := NewInstagramParser()

	enrichedData := map[string]interface{}{}

	results, err := parser.ParseInsightsDaily(enrichedData, "insta123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if results != nil {
		t.Fatal("expected nil result for empty data")
	}
}

func TestInstagramParser_setDemographicField(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name      string
		fieldName string
		values    []string
		checkFunc func(*kafkamodels.ParsedInstagramInsight) bool
	}{
		{
			name:      "AudienceAge",
			fieldName: "AudienceAge",
			values:    []string{"18-24:100", "25-34:200"},
			checkFunc: func(p *kafkamodels.ParsedInstagramInsight) bool { return len(p.AudienceAge) == 2 },
		},
		{
			name:      "AudienceGender",
			fieldName: "AudienceGender",
			values:    []string{"M:300", "F:400"},
			checkFunc: func(p *kafkamodels.ParsedInstagramInsight) bool { return len(p.AudienceGender) == 2 },
		},
		{
			name:      "AudienceCity",
			fieldName: "AudienceCity",
			values:    []string{"New York:100", "Los Angeles:80"},
			checkFunc: func(p *kafkamodels.ParsedInstagramInsight) bool { return len(p.AudienceCity) == 2 },
		},
		{
			name:      "AudienceCountry",
			fieldName: "AudienceCountry",
			values:    []string{"US:500", "UK:200"},
			checkFunc: func(p *kafkamodels.ParsedInstagramInsight) bool { return len(p.AudienceCountry) == 2 },
		},
		{
			name:      "AudienceAgeByEngagement",
			fieldName: "AudienceAge_by_engagement",
			values:    []string{"18-24:50"},
			checkFunc: func(p *kafkamodels.ParsedInstagramInsight) bool { return len(p.AudienceAgeByEngagement) == 1 },
		},
		{
			name:      "AudienceGenderByReach",
			fieldName: "AudienceGender_by_reach",
			values:    []string{"M:150"},
			checkFunc: func(p *kafkamodels.ParsedInstagramInsight) bool { return len(p.AudienceGenderByReach) == 1 },
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parsed := &kafkamodels.ParsedInstagramInsight{}
			parser.setDemographicField(parsed, tc.fieldName, tc.values)

			if !tc.checkFunc(parsed) {
				t.Fatal("check function failed")
			}
		})
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}


func TestInstagramParser_ParseInsightsWithDemographics_EmptyInsights(t *testing.T) {
	parser := NewInstagramParser()

	enrichedData := map[string]interface{}{
		"user_info": map[string]interface{}{
			"name":     "Test Account",
			"username": "testuser",
		},
	}

	result, err := parser.ParseInsightsWithDemographics(enrichedData, "insta123", "record123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramParser_ParseInsightsWithDemographics_WithDemographics(t *testing.T) {
	parser := NewInstagramParser()

	enrichedData := map[string]interface{}{
		"user_info": map[string]interface{}{
			"name":            "Test Account",
			"username":        "testuser",
			"followers_count": float64(10000),
		},
		"insights": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"name": "impressions",
					"values": []interface{}{
						map[string]interface{}{"value": float64(5000)},
					},
				},
			},
		},
		"demographics": map[string]interface{}{
			"data": []interface{}{
				map[string]interface{}{
					"name": "engaged_audience_demographics",
					"total_value": map[string]interface{}{
						"breakdowns": []interface{}{
							map[string]interface{}{
								"dimension_keys": []interface{}{"age"},
								"results": []interface{}{
									map[string]interface{}{
										"dimension_values": []interface{}{"18-24"},
										"value":            float64(100),
									},
								},
							},
						},
					},
				},
			},
		},
	}

	result, err := parser.ParseInsightsWithDemographics(enrichedData, "insta123", "record123")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.FollowersCount != 10000 {
		t.Fatalf("expected FollowersCount 10000, got %d", result.FollowersCount)
	}
}

func TestInstagramParser_ParseMedia_AllMediaTypes(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name            string
		mediaType       string
		expectedType    string
		mediaProductType string
	}{
		{
			name:            "IMAGE type",
			mediaType:       "IMAGE",
			expectedType:    "IMAGE",
		},
		{
			name:            "VIDEO type",
			mediaType:       "VIDEO",
			expectedType:    "VIDEO",
		},
		{
			name:            "CAROUSEL_ALBUM type",
			mediaType:       "CAROUSEL_ALBUM",
			expectedType:    "CAROUSEL_ALBUM",
		},
		{
			name:            "REELS product type",
			mediaType:       "VIDEO",
			mediaProductType: "REELS",
			expectedType:    "REELS",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			media := kafkamodels.RawInstagramMedia{
				ID:               "media123",
				MediaType:        tc.mediaType,
				MediaProductType: tc.mediaProductType,
				Timestamp:        "2024-01-15T10:30:00+0000",
			}

			result := parser.ParseMedia(media, "insta123", "Account", "user", "https://pic.jpg")

			if result.MediaType != tc.expectedType {
				t.Fatalf("expected MediaType %q, got %q", tc.expectedType, result.MediaType)
			}
		})
	}
}

func TestInstagramParser_getInt64Value_EdgeCases(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name     string
		input    interface{}
		expected int64
	}{
		{
			name:     "json.Number string",
			input:    "12345",
			expected: 0,
		},
		{
			name:     "map returns 0",
			input:    map[string]interface{}{},
			expected: 0,
		},
		{
			name:     "slice returns 0",
			input:    []interface{}{},
			expected: 0,
		},
		{
			name:     "int32 returns 0",
			input:    int32(100),
			expected: 0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.getInt64Value(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestInstagramParser_parseDemographicBreakdown(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name       string
		metricName string
		results    []interface{}
		checkField string
		expectLen  int
	}{
		{
			name:       "empty results",
			metricName: "follower_demographics",
			results:    []interface{}{},
			checkField: "AudienceAge",
			expectLen:  0,
		},
		{
			name:       "age demographics",
			metricName: "follower_demographics",
			results: []interface{}{
				map[string]interface{}{
					"dimension_values": []interface{}{"18-24"},
					"value":            100,
				},
				map[string]interface{}{
					"dimension_values": []interface{}{"25-34"},
					"value":            200,
				},
			},
			checkField: "AudienceAge",
			expectLen:  2,
		},
		{
			name:       "gender demographics",
			metricName: "follower_demographics",
			results: []interface{}{
				map[string]interface{}{
					"dimension_values": []interface{}{"M"},
					"value":            150,
				},
				map[string]interface{}{
					"dimension_values": []interface{}{"F"},
					"value":            250,
				},
			},
			checkField: "AudienceGender",
			expectLen:  2,
		},
		{
			name:       "country demographics",
			metricName: "follower_demographics",
			results: []interface{}{
				map[string]interface{}{
					"dimension_values": []interface{}{"US"},
					"value":            500,
				},
				map[string]interface{}{
					"dimension_values": []interface{}{"GB"},
					"value":            300,
				},
			},
			checkField: "AudienceCountry",
			expectLen:  2,
		},
		{
			name:       "city demographics",
			metricName: "follower_demographics",
			results: []interface{}{
				map[string]interface{}{
					"dimension_values": []interface{}{"New York"},
					"value":            400,
				},
			},
			checkField: "AudienceCity",
			expectLen:  1,
		},
		{
			name:       "gender-age demographics",
			metricName: "follower_demographics",
			results: []interface{}{
				map[string]interface{}{
					"dimension_values": []interface{}{"18-24", "M"},
					"value":            100,
				},
				map[string]interface{}{
					"dimension_values": []interface{}{"25-34", "F"},
					"value":            200,
				},
			},
			checkField: "AudienceGenderAge",
			expectLen:  2,
		},
		{
			name:       "engaged audience age",
			metricName: "engaged_audience_demographics",
			results: []interface{}{
				map[string]interface{}{
					"dimension_values": []interface{}{"18-24"},
					"value":            50,
				},
			},
			checkField: "AudienceAgeByEngagement",
			expectLen:  1,
		},
		{
			name:       "reached audience age",
			metricName: "reached_audience_demographics",
			results: []interface{}{
				map[string]interface{}{
					"dimension_values": []interface{}{"18-24"},
					"value":            75,
				},
			},
			checkField: "AudienceAgeByReach",
			expectLen:  1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parsed := &kafkamodels.ParsedInstagramInsight{}
			parser.parseDemographicBreakdown(parsed, tc.metricName, tc.results)

			var fieldValues []string
			switch tc.checkField {
			case "AudienceAge":
				fieldValues = parsed.AudienceAge
			case "AudienceGender":
				fieldValues = parsed.AudienceGender
			case "AudienceCountry":
				fieldValues = parsed.AudienceCountry
			case "AudienceCity":
				fieldValues = parsed.AudienceCity
			case "AudienceGenderAge":
				fieldValues = parsed.AudienceGenderAge
			case "AudienceAgeByEngagement":
				fieldValues = parsed.AudienceAgeByEngagement
			case "AudienceAgeByReach":
				fieldValues = parsed.AudienceAgeByReach
			}

			if len(fieldValues) != tc.expectLen {
				t.Fatalf("expected %d items in %s, got %d", tc.expectLen, tc.checkField, len(fieldValues))
			}
		})
	}
}

func TestInstagramParser_parseDemographics(t *testing.T) {
	parser := NewInstagramParser()

	cases := []struct {
		name         string
		enrichedData map[string]interface{}
		expectOnline bool
	}{
		{
			name:         "nil demographics",
			enrichedData: map[string]interface{}{},
			expectOnline: false,
		},
		{
			name: "online followers metric",
			enrichedData: map[string]interface{}{
				"demographics": map[string]interface{}{
					"data": []interface{}{
						map[string]interface{}{
							"name": "online_followers",
							"total_value": map[string]interface{}{
								"value": map[string]interface{}{
									"0":  100,
									"12": 200,
								},
							},
						},
					},
				},
			},
			expectOnline: true,
		},
		{
			name: "engaged audience demographics with gender",
			enrichedData: map[string]interface{}{
				"demographics": map[string]interface{}{
					"data": []interface{}{
						map[string]interface{}{
							"name": "engaged_audience_demographics",
							"total_value": map[string]interface{}{
								"breakdowns": []interface{}{
									map[string]interface{}{
										"results": []interface{}{
											map[string]interface{}{
												"dimension_values": []interface{}{"M"},
												"value":            100,
											},
											map[string]interface{}{
												"dimension_values": []interface{}{"F"},
												"value":            150,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectOnline: false,
		},
		{
			name: "reached audience demographics with age",
			enrichedData: map[string]interface{}{
				"demographics": map[string]interface{}{
					"data": []interface{}{
						map[string]interface{}{
							"name": "reached_audience_demographics",
							"total_value": map[string]interface{}{
								"breakdowns": []interface{}{
									map[string]interface{}{
										"results": []interface{}{
											map[string]interface{}{
												"dimension_values": []interface{}{"18-24"},
												"value":            200,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectOnline: false,
		},
		{
			name: "follower demographics with gender",
			enrichedData: map[string]interface{}{
				"demographics": map[string]interface{}{
					"data": []interface{}{
						map[string]interface{}{
							"name": "follower_demographics",
							"total_value": map[string]interface{}{
								"breakdowns": []interface{}{
									map[string]interface{}{
										"results": []interface{}{
											map[string]interface{}{
												"dimension_values": []interface{}{"M"},
												"value":            300,
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectOnline: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parsed := &kafkamodels.ParsedInstagramInsight{}
			parser.parseDemographics(parsed, tc.enrichedData)

			if tc.expectOnline && len(parsed.OnlineFollowers) == 0 {
				t.Fatal("expected OnlineFollowers to be populated")
			}
		})
	}
}
