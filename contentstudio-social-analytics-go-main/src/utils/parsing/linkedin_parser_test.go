package parsing

import (
	"encoding/json"
	"testing"
	"time"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestParsePost(t *testing.T) {
	cases := []struct {
		name      string
		rawJSON   string
		checkFunc func(*testing.T, interface{})
		expectErr bool
	}{
		{
			name: "parses basic post with commentary",
			rawJSON: `{
				"id": "urn:li:share:123456789",
				"commentary": "Hello world #linkedin #test",
				"createdAt": 1704067200000,
				"publishedAt": 1704067200000,
				"lifecycleState": "PUBLISHED",
				"visibility": "PUBLIC"
			}`,
			checkFunc: func(t *testing.T, result interface{}) {
				post := result.(*interface{})
				if post == nil {
					t.Fatal("expected non-nil post")
				}
			},
			expectErr: false,
		},
		{
			name: "parses post with meta stats",
			rawJSON: `{
				"id": "urn:li:share:123456789",
				"createdAt": 1704067200000,
				"meta": {
					"stats": {
						"clickCount": 100,
						"commentCount": 50,
						"engagement": 0.05,
						"impressionCount": 5000,
						"uniqueImpressionsCount": 4000,
						"shareCount": 25,
						"likeCount": 200
					}
				}
			}`,
			expectErr: false,
		},
		{
			name: "parses post with images content",
			rawJSON: `{
				"id": "urn:li:share:123456789",
				"createdAt": 1704067200000,
				"content": {
					"multiImage": {
						"images": [
							{"id": "img1"}
						]
					}
				}
			}`,
			expectErr: false,
		},
		{
			name: "parses post with video content",
			rawJSON: `{
				"id": "urn:li:share:123456789",
				"createdAt": 1704067200000,
				"content": {
					"media": {
						"id": "urn:li:video:123"
					}
				}
			}`,
			expectErr: false,
		},
		{
			name: "parses post with article content",
			rawJSON: `{
				"id": "urn:li:share:123456789",
				"createdAt": 1704067200000,
				"content": {
					"article": {
						"title": "My Article",
						"source": "https://example.com"
					}
				}
			}`,
			expectErr: false,
		},
		{
			name: "parses post with poll content",
			rawJSON: `{
				"id": "urn:li:share:123456789",
				"createdAt": 1704067200000,
				"content": {
					"poll": {
						"question": "What do you think?",
						"options": ["Option A", "Option B"]
					}
				}
			}`,
			expectErr: false,
		},
		{
			name:      "invalid JSON returns error",
			rawJSON:   `{invalid json}`,
			expectErr: true,
		},
		{
			name: "parses text-only post",
			rawJSON: `{
				"id": "urn:li:share:123456789",
				"createdAt": 1704067200000,
				"commentary": "Just text, no media"
			}`,
			expectErr: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParsePost(json.RawMessage(tc.rawJSON))

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result == nil {
				t.Fatal("expected non-nil result")
			}
		})
	}
}

func TestParsePost_FieldValues(t *testing.T) {
	rawJSON := `{
		"id": "urn:li:share:123456789",
		"commentary": "Hello #linkedin world",
		"createdAt": 1704067200000,
		"publishedAt": 1704153600000,
		"lastModifiedAt": 1704240000000,
		"lifecycleState": "PUBLISHED",
		"visibility": "PUBLIC",
		"isReshareDisabledByAuthor": true,
		"distribution": {
			"feedDistribution": "MAIN_FEED",
			"thirdPartyDistributionChannels": ["TWITTER", "FACEBOOK"]
		},
		"meta": {
			"stats": {
				"clickCount": 100,
				"commentCount": 50,
				"engagement": 0.05,
				"impressionCount": 5000,
				"uniqueImpressionsCount": 4000,
				"shareCount": 25,
				"likeCount": 200
			},
			"assets": {
				"images": [{"downloadUrl": "https://example.com/img.jpg"}],
				"videos": [{"thumbnail": "https://example.com/thumb.jpg"}]
			}
		}
	}`

	result, err := ParsePost(json.RawMessage(rawJSON))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.PostID != "123456789" {
		t.Fatalf("expected PostID '123456789', got %q", result.PostID)
	}

	if result.Activity != "urn:li:share:123456789" {
		t.Fatalf("expected Activity 'urn:li:share:123456789', got %q", result.Activity)
	}

	if result.LifecycleState != "PUBLISHED" {
		t.Fatalf("expected LifecycleState 'PUBLISHED', got %q", result.LifecycleState)
	}

	if result.Visibility != "PUBLIC" {
		t.Fatalf("expected Visibility 'PUBLIC', got %q", result.Visibility)
	}

	if !result.IsReshareDisabled {
		t.Fatal("expected IsReshareDisabled to be true")
	}

	if result.PostClicks != 100 {
		t.Fatalf("expected PostClicks 100, got %d", result.PostClicks)
	}

	if result.Comments != 50 {
		t.Fatalf("expected Comments 50, got %d", result.Comments)
	}

	if result.Impressions != 5000 {
		t.Fatalf("expected Impressions 5000, got %d", result.Impressions)
	}

	if result.Reach != 4000 {
		t.Fatalf("expected Reach 4000, got %d", result.Reach)
	}

	if result.Repost != 25 {
		t.Fatalf("expected Repost 25, got %d", result.Repost)
	}

	if result.Favorites != 200 {
		t.Fatalf("expected Favorites 200, got %d", result.Favorites)
	}

	if result.FeedDistribution != "MAIN_FEED" {
		t.Fatalf("expected FeedDistribution 'MAIN_FEED', got %q", result.FeedDistribution)
	}

	if len(result.ThirdPartyChannels) != 2 {
		t.Fatalf("expected 2 third party channels, got %d", len(result.ThirdPartyChannels))
	}

	if len(result.Hashtags) != 1 || result.Hashtags[0] != "linkedin" {
		t.Fatalf("expected hashtags [linkedin], got %v", result.Hashtags)
	}
}

func TestParsePost_MediaTypes(t *testing.T) {
	cases := []struct {
		name             string
		rawJSON          string
		expectedType     string
		expectedHasMedia bool
	}{
		{
			name: "multiImage is images",
			rawJSON: `{
				"id": "urn:li:share:123",
				"createdAt": 1704067200000,
				"content": {"multiImage": {"images": []}}
			}`,
			expectedType: "images",
		},
		{
			name: "video media is videos",
			rawJSON: `{
				"id": "urn:li:share:123",
				"createdAt": 1704067200000,
				"content": {"media": {"id": "urn:li:video:123"}}
			}`,
			expectedType: "videos",
		},
		{
			name: "document media is carousel",
			rawJSON: `{
				"id": "urn:li:share:123",
				"createdAt": 1704067200000,
				"content": {"media": {"id": "urn:li:document:123"}}
			}`,
			expectedType: "carousel",
		},
		{
			name: "article is link",
			rawJSON: `{
				"id": "urn:li:share:123",
				"createdAt": 1704067200000,
				"content": {"article": {"title": "Test"}}
			}`,
			expectedType: "link",
		},
		{
			name: "poll is poll",
			rawJSON: `{
				"id": "urn:li:share:123",
				"createdAt": 1704067200000,
				"content": {"poll": {"question": "Test?"}}
			}`,
			expectedType: "poll",
		},
		{
			name: "no content is text",
			rawJSON: `{
				"id": "urn:li:share:123",
				"createdAt": 1704067200000
			}`,
			expectedType: "text",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParsePost(json.RawMessage(tc.rawJSON))

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.MediaType != tc.expectedType {
				t.Fatalf("expected MediaType %q, got %q", tc.expectedType, result.MediaType)
			}
		})
	}
}

func TestParseMediaAsset(t *testing.T) {
	cases := []struct {
		name         string
		rawJSON      string
		expectedType string
		expectErr    bool
	}{
		{
			name:         "parses image asset",
			rawJSON:      `{"id": "urn:li:image:123", "downloadUrl": "https://example.com/img.jpg"}`,
			expectedType: "image",
			expectErr:    false,
		},
		{
			name:         "parses video asset",
			rawJSON:      `{"id": "urn:li:video:123", "downloadUrl": "https://example.com/video.mp4", "thumbnail": "https://example.com/thumb.jpg"}`,
			expectedType: "video",
			expectErr:    false,
		},
		{
			name:      "invalid JSON returns error",
			rawJSON:   `{invalid}`,
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseMediaAsset(json.RawMessage(tc.rawJSON))

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if result.Type != tc.expectedType {
				t.Fatalf("expected Type %q, got %q", tc.expectedType, result.Type)
			}
		})
	}
}

func TestParseStat(t *testing.T) {
	cases := []struct {
		name        string
		rawJSON     string
		expectedLen int
		expectErr   bool
	}{
		{
			name: "parses batch stats",
			rawJSON: `{
				"elements": [
					{
						"ugcPost": "urn:li:ugcPost:123",
						"totalShareStatistics": {
							"commentCount": 10,
							"likeCount": 100,
							"uniqueImpressionsCount": 5000,
							"shareCount": 25,
							"clickCount": 50,
							"impressionCount": 6000
						}
					},
					{
						"share": "urn:li:share:456",
						"totalShareStatistics": {
							"commentCount": 5,
							"likeCount": 50
						}
					}
				]
			}`,
			expectedLen: 2,
			expectErr:   false,
		},
		{
			name: "parses single stat",
			rawJSON: `{
				"ugcPost": "urn:li:ugcPost:123",
				"totalShareStatistics": {
					"commentCount": 10,
					"likeCount": 100
				}
			}`,
			expectedLen: 1,
			expectErr:   false,
		},
		{
			name:        "empty elements returns nil",
			rawJSON:     `{"elements": []}`,
			expectedLen: 0,
			expectErr:   false,
		},
		{
			name:      "invalid JSON returns error",
			rawJSON:   `{invalid}`,
			expectErr: true,
		},
		{
			name: "no ID returns nil",
			rawJSON: `{
				"totalShareStatistics": {
					"commentCount": 10
				}
			}`,
			expectedLen: 0,
			expectErr:   false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result, err := ParseStat(json.RawMessage(tc.rawJSON))

			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}

			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(result) != tc.expectedLen {
				t.Fatalf("expected %d stats, got %d", tc.expectedLen, len(result))
			}
		})
	}
}

func TestParseStat_Values(t *testing.T) {
	rawJSON := `{
		"ugcPost": "urn:li:ugcPost:123",
		"totalShareStatistics": {
			"commentCount": 10,
			"likeCount": 100,
			"uniqueImpressionsCount": 5000,
			"shareCount": 25,
			"clickCount": 50,
			"impressionCount": 6000
		}
	}`

	result, err := ParseStat(json.RawMessage(rawJSON))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 stat, got %d", len(result))
	}

	stat := result[0]

	if stat.ActivityID != "urn:li:ugcPost:123" {
		t.Fatalf("expected ActivityID 'urn:li:ugcPost:123', got %q", stat.ActivityID)
	}

	if stat.CommentCount != 10 {
		t.Fatalf("expected CommentCount 10, got %d", stat.CommentCount)
	}

	if stat.LikeCount != 100 {
		t.Fatalf("expected LikeCount 100, got %d", stat.LikeCount)
	}

	if stat.UniqueImpressionsCount != 5000 {
		t.Fatalf("expected UniqueImpressionsCount 5000, got %d", stat.UniqueImpressionsCount)
	}

	if stat.ShareCount != 25 {
		t.Fatalf("expected ShareCount 25, got %d", stat.ShareCount)
	}

	if stat.ClickCount != 50 {
		t.Fatalf("expected ClickCount 50, got %d", stat.ClickCount)
	}

	if stat.ImpressionCount != 6000 {
		t.Fatalf("expected ImpressionCount 6000, got %d", stat.ImpressionCount)
	}
}

func TestEnrichPostWithStats(t *testing.T) {
	t.Run("nil post does nothing", func(t *testing.T) {
		EnrichPostWithStats(nil, nil)
	})

	t.Run("nil stats does nothing", func(t *testing.T) {
		post, _ := ParsePost(json.RawMessage(`{"id": "urn:li:share:123", "createdAt": 1704067200000}`))
		EnrichPostWithStats(post, nil)
	})

	t.Run("enriches post with matching stat", func(t *testing.T) {
		post, _ := ParsePost(json.RawMessage(`{"id": "urn:li:share:123", "createdAt": 1704067200000}`))

		stats, _ := ParseStat(json.RawMessage(`{
			"ugcPost": "urn:li:share:123",
			"totalShareStatistics": {
				"commentCount": 10,
				"likeCount": 100,
				"uniqueImpressionsCount": 5000,
				"shareCount": 25,
				"clickCount": 50,
				"impressionCount": 6000
			}
		}`))

		statsMap := map[string]*kafkamodels.ParsedLinkedinStat{}
		for _, s := range stats {
			statsMap[s.ActivityID] = s
		}

		EnrichPostWithStats(post, statsMap)

		if post.Comments != 10 {
			t.Fatalf("expected Comments 10, got %d", post.Comments)
		}

		if post.Favorites != 100 {
			t.Fatalf("expected Favorites 100, got %d", post.Favorites)
		}
	})
}

func TestEnrichPostWithMedia(t *testing.T) {
	t.Run("nil post does nothing", func(t *testing.T) {
		EnrichPostWithMedia(nil, nil)
	})

	t.Run("nil media does nothing", func(t *testing.T) {
		post, _ := ParsePost(json.RawMessage(`{"id": "urn:li:share:123", "createdAt": 1704067200000}`))
		EnrichPostWithMedia(post, nil)
	})
}

func TestParseInsights(t *testing.T) {
	t.Run("parses legacy follower data", func(t *testing.T) {
		rawJSON := `{
			"firstDegreeSize": 10000,
			"elements": [
				{
					"followerCountsBySeniority": [
						{"seniority": "urn:li:seniority:5", "followerCounts": {"organicFollowerCount": 100, "paidFollowerCount": 10}}
					],
					"followerCountsByIndustry": [
						{"industry": "urn:li:industry:4", "followerCounts": {"organicFollowerCount": 500, "paidFollowerCount": 50}}
					],
					"followerCountsByGeoCountry": [
						{"geo": "urn:li:geo:103644278", "followerCounts": {"organicFollowerCount": 1000, "paidFollowerCount": 100}}
					]
				}
			]
		}`

		result, err := ParseInsights(json.RawMessage(rawJSON))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result == nil {
			t.Fatal("expected non-nil result")
		}

		if result.TotalFollowerCount != 10000 {
			t.Fatalf("expected TotalFollowerCount 10000, got %d", result.TotalFollowerCount)
		}
	})

	t.Run("returns nil for empty elements", func(t *testing.T) {
		rawJSON := `{"firstDegreeSize": 0, "elements": []}`

		result, err := ParseInsights(json.RawMessage(rawJSON))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if result != nil {
			t.Fatal("expected nil result")
		}
	})

	t.Run("invalid JSON returns error", func(t *testing.T) {
		_, err := ParseInsights(json.RawMessage(`{invalid}`))

		if err == nil {
			t.Fatal("expected error, got nil")
		}
	})
}

func TestParseInsightsDaily(t *testing.T) {
	t.Run("parses merged page and share statistics", func(t *testing.T) {
		rawJSON := `{
			"followerData": {
				"firstDegreeSize": 10000,
				"elements": [
					{
						"followerCountsByGeoCountry": [
							{"geo": "urn:li:geo:103644278", "followerCounts": {"organicFollowerCount": 1000, "paidFollowerCount": 100}}
						]
					}
				]
			},
			"pageStatistics": {
				"elements": [
					{
						"timeRange": {"start": 1704067200000, "end": 1704153600000},
						"totalPageStatistics": {
							"views": {
								"allPageViews": {"pageViews": 500, "uniquePageViews": 400}
							}
						}
					}
				]
			},
			"shareStatistics": {
				"elements": [
					{
						"timeRange": {"start": 1704067200000, "end": 1704153600000},
						"totalShareStatistics": {
							"impressionCount": 5000,
							"uniqueImpressionsCount": 4000,
							"clickCount": 100,
							"likeCount": 200,
							"commentCount": 50,
							"shareCount": 25
						}
					}
				]
			}
		}`

		results, err := ParseInsightsDaily(json.RawMessage(rawJSON))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}

		result := results[0]

		if result.TotalFollowerCount != 10000 {
			t.Fatalf("expected TotalFollowerCount 10000, got %d", result.TotalFollowerCount)
		}

		if result.PageViews != 500 {
			t.Fatalf("expected PageViews 500, got %d", result.PageViews)
		}

		if result.ImpressionCount != 5000 {
			t.Fatalf("expected ImpressionCount 5000, got %d", result.ImpressionCount)
		}
	})

	t.Run("parses profile insights", func(t *testing.T) {
		rawJSON := `{
			"entityType": "profile",
			"impressionData": {
				"elements": [
					{"dateRange": {"start": {"year": 2024, "month": 1, "day": 15}}, "count": 1000}
				]
			},
			"commentData": {
				"elements": [
					{"dateRange": {"start": {"year": 2024, "month": 1, "day": 15}}, "count": 50}
				]
			},
			"reactionData": {
				"elements": [
					{"dateRange": {"start": {"year": 2024, "month": 1, "day": 15}}, "count": 200}
				]
			},
			"reshareData": {
				"elements": [
					{"dateRange": {"start": {"year": 2024, "month": 1, "day": 15}}, "count": 25}
				]
			},
			"followerData": {
				"elements": [
					{"dateRange": {"start": {"year": 2024, "month": 1, "day": 15}}, "memberFollowersCount": 5000}
				]
			}
		}`

		results, err := ParseInsightsDaily(json.RawMessage(rawJSON))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) == 0 {
			t.Fatal("expected at least one result")
		}

		result := results[0]

		if result.ImpressionCount != 1000 {
			t.Fatalf("expected ImpressionCount 1000, got %d", result.ImpressionCount)
		}

		if result.Comments != 50 {
			t.Fatalf("expected Comments 50, got %d", result.Comments)
		}

		if result.Reactions != 200 {
			t.Fatalf("expected Reactions 200, got %d", result.Reactions)
		}

		if result.Repost != 25 {
			t.Fatalf("expected Repost 25, got %d", result.Repost)
		}

		if result.DailyFollowerCount != 5000 {
			t.Fatalf("expected DailyFollowerCount 5000, got %d", result.DailyFollowerCount)
		}
	})

	t.Run("handles empty merged data", func(t *testing.T) {
		rawJSON := `{
			"followerData": null,
			"pageStatistics": null,
			"shareStatistics": null
		}`

		results, err := ParseInsightsDaily(json.RawMessage(rawJSON))

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(results) != 1 {
			t.Fatalf("expected 1 result with follower data only, got %d", len(results))
		}
	})
}

func TestParseProfileAnalyticsDaily(t *testing.T) {
	t.Run("parses daily analytics", func(t *testing.T) {
		rawJSON := `{
			"elements": [
				{"dateRange": {"start": {"year": 2024, "month": 1, "day": 15}}, "count": 100},
				{"dateRange": {"start": {"year": 2024, "month": 1, "day": 16}}, "count": 150}
			]
		}`

		result := parseProfileAnalyticsDaily(json.RawMessage(rawJSON))

		if len(result) != 2 {
			t.Fatalf("expected 2 results, got %d", len(result))
		}

		if result["2024-01-15"].Count != 100 {
			t.Fatalf("expected count 100 for 2024-01-15, got %d", result["2024-01-15"].Count)
		}

		if result["2024-01-16"].Count != 150 {
			t.Fatalf("expected count 150 for 2024-01-16, got %d", result["2024-01-16"].Count)
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		result := parseProfileAnalyticsDaily(nil)

		if result != nil {
			t.Fatal("expected nil result for nil input")
		}
	})

	t.Run("handles null JSON", func(t *testing.T) {
		result := parseProfileAnalyticsDaily(json.RawMessage(`null`))

		if result != nil {
			t.Fatal("expected nil result for null JSON")
		}
	})
}

func TestParseProfileFollowerDaily(t *testing.T) {
	t.Run("parses daily follower counts", func(t *testing.T) {
		rawJSON := `{
			"elements": [
				{"dateRange": {"start": {"year": 2024, "month": 1, "day": 15}}, "memberFollowersCount": 5000},
				{"dateRange": {"start": {"year": 2024, "month": 1, "day": 16}}, "memberFollowersCount": 5050}
			]
		}`

		result := parseProfileFollowerDaily(json.RawMessage(rawJSON))

		if len(result) != 2 {
			t.Fatalf("expected 2 results, got %d", len(result))
		}

		if result["2024-01-15"].Count != 5000 {
			t.Fatalf("expected count 5000 for 2024-01-15, got %d", result["2024-01-15"].Count)
		}
	})

	t.Run("handles empty data", func(t *testing.T) {
		result := parseProfileFollowerDaily(nil)

		if result != nil {
			t.Fatal("expected nil result for nil input")
		}
	})
}

func TestParseFollowerDataSnapshot(t *testing.T) {
	rawJSON := `{
		"firstDegreeSize": 10000,
		"geoNames": {"103644278": "United States"},
		"elements": [
			{
				"followerCountsBySeniority": [
					{"seniority": "urn:li:seniority:5", "followerCounts": {"organicFollowerCount": 100, "paidFollowerCount": 10}}
				],
				"followerCountsByIndustry": [
					{"industry": "urn:li:industry:4", "followerCounts": {"organicFollowerCount": 500, "paidFollowerCount": 50}}
				],
				"followerCountsByGeoCountry": [
					{"geo": "urn:li:geo:103644278", "followerCounts": {"organicFollowerCount": 1000, "paidFollowerCount": 100}}
				],
				"followerCountsByGeo": [
					{"geo": "urn:li:geo:12345", "followerCounts": {"organicFollowerCount": 200, "paidFollowerCount": 20}}
				]
			}
		]
	}`

	result := parseFollowerDataSnapshot(json.RawMessage(rawJSON))

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.TotalFollowerCount != 10000 {
		t.Fatalf("expected TotalFollowerCount 10000, got %d", result.TotalFollowerCount)
	}

	if result.FollowersBySeniority == "" {
		t.Fatal("expected FollowersBySeniority to be populated")
	}

	if result.FollowersByIndustry == "" {
		t.Fatal("expected FollowersByIndustry to be populated")
	}

	if result.FollowersByCountry == "" {
		t.Fatal("expected FollowersByCountry to be populated")
	}

	if result.FollowersByCity == "" {
		t.Fatal("expected FollowersByCity to be populated")
	}
}

func TestParseShareStatisticsDaily(t *testing.T) {
	rawJSON := `{
		"elements": [
			{
				"timeRange": {"start": 1704067200000, "end": 1704153600000},
				"totalShareStatistics": {
					"impressionCount": 5000,
					"uniqueImpressionsCount": 4000,
					"clickCount": 100,
					"likeCount": 200,
					"commentCount": 50,
					"shareCount": 25,
					"engagement": 0.05
				}
			},
			{
				"timeRange": {"start": 1704153600000, "end": 1704240000000},
				"totalShareStatistics": {
					"impressionCount": 6000,
					"uniqueImpressionsCount": 4500
				}
			}
		]
	}`

	result := parseShareStatisticsDaily(json.RawMessage(rawJSON))

	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}

	stat := result[1704067200000]
	if stat.ImpressionCount != 5000 {
		t.Fatalf("expected ImpressionCount 5000, got %d", stat.ImpressionCount)
	}

	if stat.UniqueImpressionsCount != 4000 {
		t.Fatalf("expected UniqueImpressionsCount 4000, got %d", stat.UniqueImpressionsCount)
	}

	if stat.ClickCount != 100 {
		t.Fatalf("expected ClickCount 100, got %d", stat.ClickCount)
	}
}

func TestParsePageStatisticsDaily(t *testing.T) {
	rawJSON := `{
		"elements": [
			{
				"timeRange": {"start": 1704067200000, "end": 1704153600000},
				"totalPageStatistics": {
					"views": {
						"allPageViews": {"pageViews": 500, "uniquePageViews": 400},
						"allDesktopPageViews": {"pageViews": 300},
						"allMobilePageViews": {"pageViews": 200},
						"overviewPageViews": {"pageViews": 100},
						"aboutPageViews": {"pageViews": 50}
					}
				},
				"pageStatisticsByGeoCountry": [
					{"geo": "urn:li:geo:103644278", "pageStatistics": {"views": {"allPageViews": {"pageViews": 100}}}}
				]
			}
		]
	}`

	results := parsePageStatisticsDaily(json.RawMessage(rawJSON))

	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}

	result := results[0]

	if result.PageViews != 500 {
		t.Fatalf("expected PageViews 500, got %d", result.PageViews)
	}

	if result.UniqueVisitors != 400 {
		t.Fatalf("expected UniqueVisitors 400, got %d", result.UniqueVisitors)
	}

	if result.DesktopPageViews != 300 {
		t.Fatalf("expected DesktopPageViews 300, got %d", result.DesktopPageViews)
	}

	if result.MobilePageViews != 200 {
		t.Fatalf("expected MobilePageViews 200, got %d", result.MobilePageViews)
	}

	if result.OverviewPageViews != 100 {
		t.Fatalf("expected OverviewPageViews 100, got %d", result.OverviewPageViews)
	}

	expectedDate := time.UnixMilli(1704067200000).UTC()
	if !result.CreatedAt.Equal(expectedDate) {
		t.Fatalf("expected CreatedAt %v, got %v", expectedDate, result.CreatedAt)
	}
}

func TestParseFollowerDataIntoInsights(t *testing.T) {
	rawJSON := `{
		"firstDegreeSize": 10000,
		"geoNames": {
			"103644278": "United States"
		},
		"elements": [{
			"followerCountsBySeniority": [{
				"seniority": "urn:li:seniority:3",
				"followerCounts": {
					"organicFollowerCount": 500,
					"paidFollowerCount": 100
				}
			}],
			"followerCountsByIndustry": [{
				"industry": "urn:li:industry:4",
				"followerCounts": {
					"organicFollowerCount": 800,
					"paidFollowerCount": 200
				}
			}],
			"followerCountsByGeoCountry": [{
				"geo": "urn:li:geo:103644278",
				"followerCounts": {
					"organicFollowerCount": 3000,
					"paidFollowerCount": 500
				}
			}],
			"followerCountsByGeo": [{
				"geo": "urn:li:geo:90000084",
				"followerCounts": {
					"organicFollowerCount": 1500,
					"paidFollowerCount": 250
				}
			}]
		}]
	}`

	ins := &kafkamodels.ParsedLinkedinInsights{}
	parseFollowerDataIntoInsights(json.RawMessage(rawJSON), ins)

	if ins.TotalFollowerCount != 10000 {
		t.Fatalf("expected TotalFollowerCount 10000, got %d", ins.TotalFollowerCount)
	}

	if ins.FollowersBySeniority == "" {
		t.Fatal("expected FollowersBySeniority to be populated")
	}

	if ins.FollowersByIndustry == "" {
		t.Fatal("expected FollowersByIndustry to be populated")
	}

	if ins.FollowersByCountry == "" {
		t.Fatal("expected FollowersByCountry to be populated")
	}

	if ins.FollowersByCity == "" {
		t.Fatal("expected FollowersByCity to be populated")
	}
}

func TestParseFollowerDataIntoInsights_EmptyElements(t *testing.T) {
	rawJSON := `{"firstDegreeSize": 100, "elements": []}`

	ins := &kafkamodels.ParsedLinkedinInsights{}
	parseFollowerDataIntoInsights(json.RawMessage(rawJSON), ins)

	if ins.TotalFollowerCount != 0 {
		t.Fatalf("expected TotalFollowerCount 0 for empty elements, got %d", ins.TotalFollowerCount)
	}
}

func TestParseFollowerDataIntoInsights_InvalidJSON(t *testing.T) {
	ins := &kafkamodels.ParsedLinkedinInsights{}
	parseFollowerDataIntoInsights(json.RawMessage(`invalid json`), ins)

	if ins.TotalFollowerCount != 0 {
		t.Fatalf("expected TotalFollowerCount 0 for invalid JSON, got %d", ins.TotalFollowerCount)
	}
}

func TestParsePageStatisticsIntoInsights(t *testing.T) {
	rawJSON := `{
		"elements": [{
			"totalPageStatistics": {
				"views": {
					"allPageViews": {"pageViews": 1000, "uniquePageViews": 800},
					"allDesktopPageViews": {"pageViews": 600, "uniquePageViews": 500},
					"allMobilePageViews": {"pageViews": 400, "uniquePageViews": 300},
					"overviewPageViews": {"pageViews": 200, "uniquePageViews": 150},
					"aboutPageViews": {"pageViews": 100, "uniquePageViews": 80},
					"jobsPageViews": {"pageViews": 50, "uniquePageViews": 40},
					"peoplePageViews": {"pageViews": 30, "uniquePageViews": 25},
					"careersPageViews": {"pageViews": 20, "uniquePageViews": 15},
					"lifeAtPageViews": {"pageViews": 10, "uniquePageViews": 8},
					"insightsPageViews": {"pageViews": 5, "uniquePageViews": 4},
					"productsPageViews": {"pageViews": 3, "uniquePageViews": 2}
				}
			},
			"pageStatisticsByGeoCountry": [{
				"geo": "urn:li:geo:103644278",
				"pageStatistics": {
					"views": {
						"allPageViews": {"pageViews": 500, "uniquePageViews": 400}
					}
				}
			}],
			"pageStatisticsByGeo": [{
				"geo": "urn:li:geo:90000084",
				"pageStatistics": {
					"views": {
						"allPageViews": {"pageViews": 200, "uniquePageViews": 150}
					}
				}
			}],
			"pageStatisticsByIndustryV2": [{
				"industryV2": "urn:li:industry:4",
				"pageStatistics": {
					"views": {
						"allPageViews": {"pageViews": 300, "uniquePageViews": 250}
					}
				}
			}],
			"pageStatisticsBySeniority": [{
				"seniority": "urn:li:seniority:3",
				"pageStatistics": {
					"views": {
						"allPageViews": {"pageViews": 100, "uniquePageViews": 80}
					}
				}
			}],
			"pageStatisticsByFunction": [{
				"function": "urn:li:function:1",
				"pageStatistics": {
					"views": {
						"allPageViews": {"pageViews": 50, "uniquePageViews": 40}
					}
				}
			}],
			"pageStatisticsByStaffCountRange": [{
				"staffCountRange": "SIZE_51_200",
				"pageStatistics": {
					"views": {
						"allPageViews": {"pageViews": 25, "uniquePageViews": 20}
					}
				}
			}]
		}]
	}`

	ins := &kafkamodels.ParsedLinkedinInsights{}
	parsePageStatisticsIntoInsights(json.RawMessage(rawJSON), ins)

	if ins.PageViews != 1000 {
		t.Fatalf("expected PageViews 1000, got %d", ins.PageViews)
	}

	if ins.UniqueVisitors != 800 {
		t.Fatalf("expected UniqueVisitors 800, got %d", ins.UniqueVisitors)
	}

	if ins.DesktopPageViews != 600 {
		t.Fatalf("expected DesktopPageViews 600, got %d", ins.DesktopPageViews)
	}

	if ins.MobilePageViews != 400 {
		t.Fatalf("expected MobilePageViews 400, got %d", ins.MobilePageViews)
	}

	if ins.OverviewPageViews != 200 {
		t.Fatalf("expected OverviewPageViews 200, got %d", ins.OverviewPageViews)
	}

	if ins.AboutPageViews != 100 {
		t.Fatalf("expected AboutPageViews 100, got %d", ins.AboutPageViews)
	}

	if ins.JobsPageViews != 50 {
		t.Fatalf("expected JobsPageViews 50, got %d", ins.JobsPageViews)
	}

	if ins.PeoplePageViews != 30 {
		t.Fatalf("expected PeoplePageViews 30, got %d", ins.PeoplePageViews)
	}

	if ins.CareersPageViews != 20 {
		t.Fatalf("expected CareersPageViews 20, got %d", ins.CareersPageViews)
	}

	if ins.LifeAtPageViews != 10 {
		t.Fatalf("expected LifeAtPageViews 10, got %d", ins.LifeAtPageViews)
	}

	if ins.InsightsPageViews != 5 {
		t.Fatalf("expected InsightsPageViews 5, got %d", ins.InsightsPageViews)
	}

	if ins.ProductsPageViews != 3 {
		t.Fatalf("expected ProductsPageViews 3, got %d", ins.ProductsPageViews)
	}
}

func TestParsePageStatisticsIntoInsights_EmptyElements(t *testing.T) {
	rawJSON := `{"elements": []}`

	ins := &kafkamodels.ParsedLinkedinInsights{}
	parsePageStatisticsIntoInsights(json.RawMessage(rawJSON), ins)

	if ins.PageViews != 0 {
		t.Fatalf("expected PageViews 0 for empty elements, got %d", ins.PageViews)
	}
}

func TestParsePageStatisticsIntoInsights_InvalidJSON(t *testing.T) {
	ins := &kafkamodels.ParsedLinkedinInsights{}
	parsePageStatisticsIntoInsights(json.RawMessage(`invalid json`), ins)

	if ins.PageViews != 0 {
		t.Fatalf("expected PageViews 0 for invalid JSON, got %d", ins.PageViews)
	}
}
