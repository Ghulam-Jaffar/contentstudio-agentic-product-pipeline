package parsing

import (
	"strings"
	"testing"
	"time"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestNewFacebookParser(t *testing.T) {
	parser := NewFacebookParser()

	if parser == nil {
		t.Fatal("expected non-nil parser")
	}

	if parser.log == nil {
		t.Fatal("expected logger to be initialized")
	}

	if len(parser.MediaTypeMapping) == 0 {
		t.Fatal("expected MediaTypeMapping to be initialized")
	}

	expectedMappings := map[string]string{
		"multi_share_no_end_card": "carousel",
		"photo":                   "images",
		"album":                   "images",
		"video_inline":            "videos",
		"link":                    "link",
		"share":                   "link",
	}

	for key, expected := range expectedMappings {
		if parser.MediaTypeMapping[key] != expected {
			t.Fatalf("expected MediaTypeMapping[%s] = %s, got %s", key, expected, parser.MediaTypeMapping[key])
		}
	}
}

func TestFacebookParser_getMediaTypeFromStatus(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name       string
		statusType string
		expected   string
	}{
		{
			name:       "added_photos returns images",
			statusType: "added_photos",
			expected:   "images",
		},
		{
			name:       "added_video returns videos",
			statusType: "added_video",
			expected:   "videos",
		},
		{
			name:       "shared_story returns link",
			statusType: "shared_story",
			expected:   "link",
		},
		{
			name:       "published_story returns link",
			statusType: "published_story",
			expected:   "link",
		},
		{
			name:       "mobile_status_update returns text",
			statusType: "mobile_status_update",
			expected:   "text",
		},
		{
			name:       "reels returns reels",
			statusType: "reels",
			expected:   "reels",
		},
		{
			name:       "unknown status returns others",
			statusType: "unknown_status",
			expected:   "others",
		},
		{
			name:       "empty status returns others",
			statusType: "",
			expected:   "others",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.getMediaTypeFromStatus(tc.statusType)
			if result != tc.expected {
				t.Fatalf("expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func TestFacebookParser_getInt64Value(t *testing.T) {
	parser := NewFacebookParser()

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
			name:     "int returns int64",
			input:    42,
			expected: 42,
		},
		{
			name:     "int64 returns itself",
			input:    int64(100),
			expected: 100,
		},
		{
			name:     "float64 returns truncated int64",
			input:    123.456,
			expected: 123,
		},
		{
			name:     "empty string returns 0",
			input:    "",
			expected: 0,
		},
		{
			name:     "non-empty string returns 0",
			input:    "123",
			expected: 0,
		},
		{
			name:     "negative int",
			input:    -50,
			expected: -50,
		},
		{
			name:     "zero",
			input:    0,
			expected: 0,
		},
		{
			name:     "large float64",
			input:    1e10,
			expected: 10000000000,
		},
		{
			name:     "bool returns 0",
			input:    true,
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

func TestFacebookParser_getFloat64Value(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name     string
		input    interface{}
		expected float64
	}{
		{
			name:     "nil returns 0",
			input:    nil,
			expected: 0.0,
		},
		{
			name:     "float64 returns itself",
			input:    3.14,
			expected: 3.14,
		},
		{
			name:     "int returns float64",
			input:    42,
			expected: 42.0,
		},
		{
			name:     "int64 returns float64",
			input:    int64(100),
			expected: 100.0,
		},
		{
			name:     "empty string returns 0",
			input:    "",
			expected: 0.0,
		},
		{
			name:     "bool returns 0",
			input:    true,
			expected: 0.0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.getFloat64Value(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %f, got %f", tc.expected, result)
			}
		})
	}
}

func TestFacebookParser_getStringValue(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{
			name:     "nil returns empty string",
			input:    nil,
			expected: "",
		},
		{
			name:     "string returns itself",
			input:    "hello",
			expected: "hello",
		},
		{
			name:     "empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "int returns empty string",
			input:    42,
			expected: "",
		},
		{
			name:     "bool returns empty string",
			input:    true,
			expected: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.getStringValue(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestFacebookParser_getStringArrayValue(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name     string
		input    interface{}
		expected []string
	}{
		{
			name:     "nil returns empty slice",
			input:    nil,
			expected: []string{},
		},
		{
			name:     "string slice returns itself",
			input:    []string{"a", "b", "c"},
			expected: []string{"a", "b", "c"},
		},
		{
			name:     "single string returns slice with one element",
			input:    "single",
			expected: []string{"single"},
		},
		{
			name:     "int returns empty slice",
			input:    42,
			expected: []string{},
		},
		{
			name:     "empty string slice",
			input:    []string{},
			expected: []string{},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.getStringArrayValue(tc.input)
			if len(result) != len(tc.expected) {
				t.Fatalf("expected length %d, got %d", len(tc.expected), len(result))
			}
			for i, v := range tc.expected {
				if result[i] != v {
					t.Fatalf("expected result[%d] = %q, got %q", i, v, result[i])
				}
			}
		})
	}
}

func TestFacebookParser_extractVideoID(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name     string
		link     string
		expected string
	}{
		{
			name:     "extracts video ID from standard URL",
			link:     "https://www.facebook.com/watch/videos/123456789/",
			expected: "123456789",
		},
		{
			name:     "extracts video ID without trailing slash",
			link:     "https://www.facebook.com/videos/987654321",
			expected: "987654321",
		},
		{
			name:     "returns empty for non-video URL",
			link:     "https://www.facebook.com/page/posts/123",
			expected: "",
		},
		{
			name:     "returns empty for empty link",
			link:     "",
			expected: "",
		},
		{
			name:     "handles video ID with extra path segments",
			link:     "https://www.facebook.com/videos/111222333/some-title",
			expected: "111222333",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.extractVideoID(tc.link)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestGenerateMediaID(t *testing.T) {
	cases := []struct {
		name   string
		postID string
		index  int
	}{
		{
			name:   "generates MD5 hash for post_0",
			postID: "123456_789012",
			index:  0,
		},
		{
			name:   "generates different hash for different index",
			postID: "123456_789012",
			index:  1,
		},
		{
			name:   "generates hash for empty post ID",
			postID: "",
			index:  0,
		},
		{
			name:   "generates hash for large index",
			postID: "test",
			index:  999999,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := GenerateMediaID(tc.postID, tc.index)

			if len(result) != 32 {
				t.Fatalf("expected 32 character MD5 hash, got %d characters", len(result))
			}

			for _, c := range result {
				if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f')) {
					t.Fatalf("expected lowercase hex characters, found %c", c)
				}
			}
		})
	}

	result1 := GenerateMediaID("post1", 0)
	result2 := GenerateMediaID("post1", 1)
	if result1 == result2 {
		t.Fatal("expected different hashes for different indices")
	}

	result3 := GenerateMediaID("post1", 0)
	result4 := GenerateMediaID("post2", 0)
	if result3 == result4 {
		t.Fatal("expected different hashes for different post IDs")
	}
}

func TestFacebookParser_extractMessageTags(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name     string
		tags     []struct {
			ID     string `json:"id"`
			Name   string `json:"name"`
			Type   string `json:"type"`
			Offset int    `json:"offset"`
			Length int    `json:"length"`
		}
		expected []string
	}{
		{
			name:     "empty tags returns nil",
			tags:     nil,
			expected: nil,
		},
		{
			name: "extracts single tag name",
			tags: []struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Type   string `json:"type"`
				Offset int    `json:"offset"`
				Length int    `json:"length"`
			}{
				{ID: "123", Name: "John Doe", Type: "user", Offset: 0, Length: 8},
			},
			expected: []string{"John Doe"},
		},
		{
			name: "extracts multiple tag names",
			tags: []struct {
				ID     string `json:"id"`
				Name   string `json:"name"`
				Type   string `json:"type"`
				Offset int    `json:"offset"`
				Length int    `json:"length"`
			}{
				{ID: "1", Name: "User One", Type: "user", Offset: 0, Length: 8},
				{ID: "2", Name: "User Two", Type: "user", Offset: 10, Length: 8},
			},
			expected: []string{"User One", "User Two"},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.extractMessageTags(tc.tags)

			if len(result) != len(tc.expected) {
				t.Fatalf("expected %d tags, got %d", len(tc.expected), len(result))
			}

			for i, v := range tc.expected {
				if result[i] != v {
					t.Fatalf("expected tag[%d] = %q, got %q", i, v, result[i])
				}
			}
		})
	}
}

func TestFacebookParser_convertMapToStringSlice(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name     string
		input    interface{}
		checkLen int
	}{
		{
			name:     "nil returns empty slice",
			input:    nil,
			checkLen: 0,
		},
		{
			name:     "map returns key:value pairs",
			input:    map[string]interface{}{"key1": 100, "key2": 200},
			checkLen: 2,
		},
		{
			name:     "non-map returns single element",
			input:    "single value",
			checkLen: 1,
		},
		{
			name:     "empty map returns empty slice",
			input:    map[string]interface{}{},
			checkLen: 0,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.convertMapToStringSlice(tc.input)

			if len(result) != tc.checkLen {
				t.Fatalf("expected %d elements, got %d", tc.checkLen, len(result))
			}
		})
	}
}

func TestFacebookParser_sumMapValues(t *testing.T) {
	parser := NewFacebookParser()

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
			name:     "map sums all values",
			input:    map[string]interface{}{"a": 10, "b": 20, "c": 30},
			expected: 60,
		},
		{
			name:     "non-map returns value",
			input:    42,
			expected: 42,
		},
		{
			name:     "empty map returns 0",
			input:    map[string]interface{}{},
			expected: 0,
		},
		{
			name:     "map with float values",
			input:    map[string]interface{}{"a": 10.5, "b": 20.9},
			expected: 30,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.sumMapValues(tc.input)
			if result != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestFacebookParser_getOnlineFans(t *testing.T) {
	parser := NewFacebookParser()

	t.Run("initializes all 24 hours", func(t *testing.T) {
		result := parser.getOnlineFans(nil)

		if len(result) != 24 {
			t.Fatalf("expected 24 hours, got %d", len(result))
		}
	})

	t.Run("fills in actual data", func(t *testing.T) {
		input := map[string]interface{}{
			"0":  100,
			"12": 500,
			"23": 200,
		}

		result := parser.getOnlineFans(input)

		if result["0"] != 100 {
			t.Fatalf("expected hour 0 = 100, got %v", result["0"])
		}
		if result["12"] != 500 {
			t.Fatalf("expected hour 12 = 500, got %v", result["12"])
		}
		if result["23"] != 200 {
			t.Fatalf("expected hour 23 = 200, got %v", result["23"])
		}
	})
}

func TestFacebookParser_getPrimeTime(t *testing.T) {
	parser := NewFacebookParser()

	baseDate := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)

	cases := []struct {
		name         string
		activity     map[string]interface{}
		expectedHour int
	}{
		{
			name:         "empty activity returns start of day",
			activity:     map[string]interface{}{},
			expectedHour: 0,
		},
		{
			name:         "single hour with activity",
			activity:     map[string]interface{}{"10": 500},
			expectedHour: 10,
		},
		{
			name:         "multiple hours returns highest",
			activity:     map[string]interface{}{"8": 100, "12": 500, "18": 300},
			expectedHour: 12,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.getPrimeTime(tc.activity, baseDate)
			if result.Hour() != tc.expectedHour {
				t.Fatalf("expected hour %d, got %d", tc.expectedHour, result.Hour())
			}
		})
	}
}

func TestFacebookParser_calculateAverageOnlineFans(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name     string
		activity map[string]interface{}
		expected int64
	}{
		{
			name:     "empty activity returns 0",
			activity: map[string]interface{}{},
			expected: 0,
		},
		{
			name:     "calculates average correctly",
			activity: map[string]interface{}{"0": 100, "1": 200, "2": 300},
			expected: 200,
		},
		{
			name:     "single value",
			activity: map[string]interface{}{"12": 1000},
			expected: 1000,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.calculateAverageOnlineFans(tc.activity)
			if result != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, result)
			}
		})
	}
}

func TestFacebookParser_extractGenderFromGenderAge(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name          string
		input         interface{}
		expectedEmpty bool
	}{
		{
			name:          "nil input",
			input:         nil,
			expectedEmpty: false,
		},
		{
			name: "valid gender age data",
			input: map[string]interface{}{
				"M.18-24": 100,
				"M.25-34": 200,
				"F.18-24": 150,
				"F.25-34": 175,
			},
			expectedEmpty: false,
		},
		{
			name:          "empty map",
			input:         map[string]interface{}{},
			expectedEmpty: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.extractGenderFromGenderAge(tc.input)
			if tc.expectedEmpty && len(result) > 0 {
				t.Fatal("expected empty result")
			}
		})
	}
}

func TestFacebookParser_extractAgeFromGenderAge(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name  string
		input interface{}
	}{
		{
			name:  "nil input",
			input: nil,
		},
		{
			name: "valid gender age data with underscore",
			input: map[string]interface{}{
				"M_18-24": 100,
				"M_25-34": 200,
				"F_18-24": 150,
			},
		},
		{
			name: "valid gender age data with dot",
			input: map[string]interface{}{
				"M.18-24": 100,
				"M.25-34": 200,
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.extractAgeFromGenderAge(tc.input)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
		})
	}
}

func TestFacebookParser_ParsePost_BasicFields(t *testing.T) {
	parser := NewFacebookParser()

	now := time.Now().UTC()
	rawPost := kafkamodels.RawFacebookPost{
		ID:           "123456_789012",
		Message:      "Test post message",
		PermalinkURL: "https://facebook.com/post/123",
		StatusType:   "added_photos",
		CreatedTime:  kafkamodels.FacebookTime{Time: now},
		UpdatedTime:  kafkamodels.FacebookTime{Time: now.Add(time.Hour)},
	}

	post, assets, err := parser.ParsePost(rawPost, "123456", "Test Page", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post == nil {
		t.Fatal("expected non-nil post")
	}

	if post.PageID != "123456" {
		t.Fatalf("expected PageID %q, got %q", "123456", post.PageID)
	}

	if post.PageName != "Test Page" {
		t.Fatalf("expected PageName %q, got %q", "Test Page", post.PageName)
	}

	if post.PostID != "123456_789012" {
		t.Fatalf("expected PostID %q, got %q", "123456_789012", post.PostID)
	}

	if post.Caption != "Test post message" {
		t.Fatalf("expected Caption %q, got %q", "Test post message", post.Caption)
	}

	if post.Permalink != "https://facebook.com/post/123" {
		t.Fatalf("expected Permalink %q, got %q", "https://facebook.com/post/123", post.Permalink)
	}

	if post.MediaType != "images" {
		t.Fatalf("expected MediaType %q, got %q", "images", post.MediaType)
	}

	if assets == nil {
		t.Fatal("expected non-nil assets")
	}
}

func TestFacebookParser_ParsePost_WithReactions(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID:          "123456_789012",
		StatusType:  "added_photos",
		CreatedTime: kafkamodels.FacebookTime{Time: time.Now()},
		Total: &struct {
			Summary *struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		}{
			Summary: &struct {
				TotalCount int `json:"total_count"`
			}{
				TotalCount: 100,
			},
		},
		Like: &struct {
			Summary *struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		}{
			Summary: &struct {
				TotalCount int `json:"total_count"`
			}{
				TotalCount: 50,
			},
		},
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
				TotalCount: 25,
				CanComment: true,
			},
		},
	}

	post, _, err := parser.ParsePost(rawPost, "123456", "Test Page", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.Total != 100 {
		t.Fatalf("expected Total 100, got %d", post.Total)
	}

	if post.Like != 50 {
		t.Fatalf("expected Like 50, got %d", post.Like)
	}

	if post.Comments != 25 {
		t.Fatalf("expected Comments 25, got %d", post.Comments)
	}
}

func TestFacebookParser_ParsePost_ReelDetection(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID:           "123456_789012",
		StatusType:   "added_video",
		PermalinkURL: "https://www.facebook.com/reel/123456789",
		CreatedTime:  kafkamodels.FacebookTime{Time: time.Now()},
	}

	post, _, err := parser.ParsePost(rawPost, "123456", "Test Page", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.MediaType != "reels" {
		t.Fatalf("expected MediaType %q, got %q", "reels", post.MediaType)
	}
}

func TestFacebookParser_ParsePost_TextPostWithNoAttachments(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID:          "123456_789012",
		Message:     "Just a text update",
		StatusType:  "shared_story",
		Attachments: nil,
		CreatedTime: kafkamodels.FacebookTime{Time: time.Now()},
	}

	post, _, err := parser.ParsePost(rawPost, "123456", "Test Page", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.MediaType != "link" {
		t.Fatalf("expected MediaType %q, got %q", "link", post.MediaType)
	}
}

func TestFacebookParser_ParsePost_WithAttachments(t *testing.T) {
	parser := NewFacebookParser()
	now := time.Now()

	rawPost := kafkamodels.RawFacebookPost{
		ID:          "123456_789012",
		Message:     "Post with attachment",
		StatusType:  "added_photos",
		CreatedTime: kafkamodels.FacebookTime{Time: now},
		UpdatedTime: kafkamodels.FacebookTime{Time: now.Add(time.Hour)},
		Attachments: &struct {
			Data []struct {
				Type        string `json:"type"`
				MediaType   string `json:"media_type"`
				Caption     string `json:"caption"`
				Description string `json:"description"`
				Link        string `json:"link"`
				Target      *struct {
					ID string `json:"id"`
				} `json:"target"`
				Media *struct {
					Src    string `json:"src,omitempty"`
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Src    string `json:"src"`
						Source string `json:"source"`
					} `json:"image"`
				} `json:"media"`
				Subattachments *struct {
					Data []struct {
						Type      string `json:"type"`
						MediaType string `json:"media_type"`
						Media     *struct {
							Src    string `json:"src"`
							Source string `json:"source"`
							Image  *struct {
								Height int    `json:"height"`
								Width  int    `json:"width"`
								Src    string `json:"src"`
								Source string `json:"source"`
							} `json:"image"`
						} `json:"media"`
					} `json:"data"`
				} `json:"subattachments"`
			} `json:"data"`
		}{
			Data: []struct {
				Type        string `json:"type"`
				MediaType   string `json:"media_type"`
				Caption     string `json:"caption"`
				Description string `json:"description"`
				Link        string `json:"link"`
				Target      *struct {
					ID string `json:"id"`
				} `json:"target"`
				Media *struct {
					Src    string `json:"src,omitempty"`
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Src    string `json:"src"`
						Source string `json:"source"`
					} `json:"image"`
				} `json:"media"`
				Subattachments *struct {
					Data []struct {
						Type      string `json:"type"`
						MediaType string `json:"media_type"`
						Media     *struct {
							Src    string `json:"src"`
							Source string `json:"source"`
							Image  *struct {
								Height int    `json:"height"`
								Width  int    `json:"width"`
								Src    string `json:"src"`
								Source string `json:"source"`
							} `json:"image"`
						} `json:"media"`
					} `json:"data"`
				} `json:"subattachments"`
			}{
				{
					Type:        "photo",
					Caption:     "Test Caption",
					Description: "Test Description",
					Link:        "https://example.com/photo",
					Media: &struct {
						Src    string `json:"src,omitempty"`
						Source string `json:"source"`
						Image  *struct {
							Height int    `json:"height"`
							Width  int    `json:"width"`
							Src    string `json:"src"`
							Source string `json:"source"`
						} `json:"image"`
					}{
						Image: &struct {
							Height int    `json:"height"`
							Width  int    `json:"width"`
							Src    string `json:"src"`
							Source string `json:"source"`
						}{
							Src: "https://example.com/image.jpg",
						},
					},
				},
			},
		},
	}

	post, assets, err := parser.ParsePost(rawPost, "123456", "Test Page", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post == nil {
		t.Fatal("expected non-nil post")
	}

	if post.FullPicture != "https://example.com/image.jpg" {
		t.Fatalf("expected FullPicture %q, got %q", "https://example.com/image.jpg", post.FullPicture)
	}

	if post.Link != "https://example.com/photo" {
		t.Fatalf("expected Link %q, got %q", "https://example.com/photo", post.Link)
	}

	if post.Description != "Test Description" {
		t.Fatalf("expected Description %q, got %q", "Test Description", post.Description)
	}

	if len(assets) == 0 {
		t.Fatal("expected at least one media asset")
	}
}

func TestFacebookParser_ParsePost_WithShares(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID:          "123456_789012",
		Message:     "Post with shares",
		CreatedTime: kafkamodels.FacebookTime{Time: time.Now()},
		Shares: &struct {
			Count int `json:"count"`
		}{
			Count: 50,
		},
	}

	post, _, err := parser.ParsePost(rawPost, "123456", "Test Page", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.Shares != 50 {
		t.Fatalf("expected Shares 50, got %d", post.Shares)
	}
}

func TestFacebookParser_ParsePost_VideoPost(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID:          "123456_789012",
		Message:     "Video post",
		StatusType:  "added_video",
		CreatedTime: kafkamodels.FacebookTime{Time: time.Now()},
		Attachments: &struct {
			Data []struct {
				Type        string `json:"type"`
				MediaType   string `json:"media_type"`
				Caption     string `json:"caption"`
				Description string `json:"description"`
				Link        string `json:"link"`
				Target      *struct {
					ID string `json:"id"`
				} `json:"target"`
				Media *struct {
					Src    string `json:"src,omitempty"`
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Src    string `json:"src"`
						Source string `json:"source"`
					} `json:"image"`
				} `json:"media"`
				Subattachments *struct {
					Data []struct {
						Type      string `json:"type"`
						MediaType string `json:"media_type"`
						Media     *struct {
							Src    string `json:"src"`
							Source string `json:"source"`
							Image  *struct {
								Height int    `json:"height"`
								Width  int    `json:"width"`
								Src    string `json:"src"`
								Source string `json:"source"`
							} `json:"image"`
						} `json:"media"`
					} `json:"data"`
				} `json:"subattachments"`
			} `json:"data"`
		}{
			Data: []struct {
				Type        string `json:"type"`
				MediaType   string `json:"media_type"`
				Caption     string `json:"caption"`
				Description string `json:"description"`
				Link        string `json:"link"`
				Target      *struct {
					ID string `json:"id"`
				} `json:"target"`
				Media *struct {
					Src    string `json:"src,omitempty"`
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Src    string `json:"src"`
						Source string `json:"source"`
					} `json:"image"`
				} `json:"media"`
				Subattachments *struct {
					Data []struct {
						Type      string `json:"type"`
						MediaType string `json:"media_type"`
						Media     *struct {
							Src    string `json:"src"`
							Source string `json:"source"`
							Image  *struct {
								Height int    `json:"height"`
								Width  int    `json:"width"`
								Src    string `json:"src"`
								Source string `json:"source"`
							} `json:"image"`
						} `json:"media"`
					} `json:"data"`
				} `json:"subattachments"`
			}{
				{
					Type: "video_inline",
					Link: "https://www.facebook.com/watch/?v=987654321",
				},
			},
		},
	}

	post, _, err := parser.ParsePost(rawPost, "123456", "Test Page", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.MediaType != "videos" {
		t.Fatalf("expected MediaType %q, got %q", "videos", post.MediaType)
	}
}

func TestFacebookParser_ParsePost_NoCreatedTime(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID:      "123456_789012",
		Message: "Post without created time",
	}

	post, _, err := parser.ParsePost(rawPost, "123456", "Test Page", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.DayOfWeek != "" {
		t.Fatalf("expected empty DayOfWeek, got %q", post.DayOfWeek)
	}
}

func TestFacebookParser_ParsePost_SharePost(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID:          "123456_789012",
		Message:     "Share post",
		StatusType:  "shared_story",
		CreatedTime: kafkamodels.FacebookTime{Time: time.Now()},
		Attachments: nil,
	}

	post, _, err := parser.ParsePost(rawPost, "123456", "Test Page", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post.MediaType != "link" {
		t.Fatalf("expected MediaType %q, got %q", "link", post.MediaType)
	}
}

func TestFacebookParser_ParseVideo(t *testing.T) {
	parser := NewFacebookParser()

	rawVideo := kafkamodels.RawFacebookVideo{
		ID:          "video123",
		PostID:      "post456",
		CreatedTime: kafkamodels.FacebookTime{Time: time.Now()},
		UpdatedTime: kafkamodels.FacebookTime{Time: time.Now().Add(time.Hour)},
	}

	result, err := parser.ParseVideo(rawVideo, "page123", "Test Page")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.VideoID != "video123" {
		t.Fatalf("expected VideoID %q, got %q", "video123", result.VideoID)
	}

	if result.PageID != "page123" {
		t.Fatalf("expected PageID %q, got %q", "page123", result.PageID)
	}

	expectedPostID := "page123_post456"
	if result.PostID != expectedPostID {
		t.Fatalf("expected PostID %q, got %q", expectedPostID, result.PostID)
	}
}

func TestFacebookParser_ParseVideoPostInsights(t *testing.T) {
	parser := NewFacebookParser()

	now := time.Now().UTC()
	videoAsPost := kafkamodels.RawFacebookPost{
		ID:          "123456_789012",
		StatusType:  "added_video",
		CreatedTime: kafkamodels.FacebookTime{Time: now},
		UpdatedTime: kafkamodels.FacebookTime{Time: now.Add(time.Hour)},
	}

	result := parser.ParseVideoPostInsights(videoAsPost)

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.CreatedTime.IsZero() {
		t.Fatal("expected CreatedTime to be set")
	}

	if result.DayOfWeek == "" {
		t.Fatal("expected DayOfWeek to be set")
	}
}

func TestSplitComposite(t *testing.T) {
	cases := []struct {
		name           string
		id             string
		expectedPageID string
		expectedPostID string
	}{
		{
			name:           "splits standard composite ID",
			id:             "123456_789012",
			expectedPageID: "123456",
			expectedPostID: "789012",
		},
		{
			name:           "handles no underscore",
			id:             "123456789012",
			expectedPageID: "",
			expectedPostID: "123456789012",
		},
		{
			name:           "handles multiple underscores",
			id:             "123_456_789",
			expectedPageID: "123",
			expectedPostID: "456_789",
		},
		{
			name:           "handles empty string",
			id:             "",
			expectedPageID: "",
			expectedPostID: "",
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			pageID, postID := splitComposite(tc.id)

			if pageID != tc.expectedPageID {
				t.Fatalf("expected pageID %q, got %q", tc.expectedPageID, pageID)
			}

			if postID != tc.expectedPostID {
				t.Fatalf("expected postID %q, got %q", tc.expectedPostID, postID)
			}
		})
	}
}

func TestParseRawFacebookPost(t *testing.T) {
	rawPost := kafkamodels.RawFacebookPost{
		ID:         "123456_789012",
		Message:    "Test message",
		StatusType: "added_photos",
		From: &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{
			ID:   "123456",
			Name: "Test Page",
		},
		CreatedTime: kafkamodels.FacebookTime{Time: time.Now()},
	}

	post, assets, err := ParseRawFacebookPost(rawPost)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if post == nil {
		t.Fatal("expected non-nil post")
	}

	if post.PageID != "123456" {
		t.Fatalf("expected PageID %q, got %q", "123456", post.PageID)
	}

	if post.PageName != "Test Page" {
		t.Fatalf("expected PageName %q, got %q", "Test Page", post.PageName)
	}

	if assets == nil {
		t.Fatal("expected non-nil assets")
	}
}

func TestParseVideoFromJSON(t *testing.T) {
	rawVideo := kafkamodels.RawFacebookVideo{
		ID:          "video123",
		PostID:      "post456",
		CreatedTime: kafkamodels.FacebookTime{Time: time.Now()},
	}

	result, err := ParseVideoFromJSON(rawVideo, "page123", "Test Page")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.VideoID != "video123" {
		t.Fatalf("expected VideoID %q, got %q", "video123", result.VideoID)
	}
}

func TestFacebookParser_ParseInsights(t *testing.T) {
	parser := NewFacebookParser()

	rawInsights := kafkamodels.RawFacebookInsights{
		Data: []kafkamodels.FacebookInsightData{
			{
				Name:   "page_follows",
				Period: "day",
				Values: []kafkamodels.FacebookInsightValue{
					{
						Value:   int64(1000),
						EndTime: "2024-01-15T08:00:00+0000",
					},
				},
			},
			{
				Name:   "page_fans",
				Period: "day",
				Values: []kafkamodels.FacebookInsightValue{
					{
						Value:   int64(5000),
						EndTime: "2024-01-15T08:00:00+0000",
					},
				},
			},
		},
	}

	result, err := parser.ParseInsights(rawInsights, "page123", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}

	if result.PageID != "page123" {
		t.Fatalf("expected PageID %q, got %q", "page123", result.PageID)
	}

	if result.WorkspaceID != "workspace1" {
		t.Fatalf("expected WorkspaceID %q, got %q", "workspace1", result.WorkspaceID)
	}
}

func TestFacebookParser_ParseInsightsDaily(t *testing.T) {
	parser := NewFacebookParser()

	rawInsights := kafkamodels.RawFacebookInsights{
		Data: []kafkamodels.FacebookInsightData{
			{
				Name:   "page_follows",
				Period: "day",
				Values: []kafkamodels.FacebookInsightValue{
					{
						Value:   int64(1000),
						EndTime: "2024-01-15T08:00:00+0000",
					},
					{
						Value:   int64(1100),
						EndTime: "2024-01-16T08:00:00+0000",
					},
				},
			},
		},
	}

	results, err := parser.ParseInsightsDaily(rawInsights, "page123", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(results) != 2 {
		t.Fatalf("expected 2 daily records, got %d", len(results))
	}

	for _, result := range results {
		if result.PageID != "page123" {
			t.Fatalf("expected PageID %q, got %q", "page123", result.PageID)
		}
	}
}

func TestFacebookParser_ParseInsights_EmptyData(t *testing.T) {
	parser := NewFacebookParser()

	rawInsights := kafkamodels.RawFacebookInsights{
		Data: []kafkamodels.FacebookInsightData{},
	}

	result, err := parser.ParseInsights(rawInsights, "page123", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Fatal("expected nil result for empty data")
	}
}

func TestParseInsightsFromJSON(t *testing.T) {
	rawInsights := kafkamodels.RawFacebookInsights{
		Data: []kafkamodels.FacebookInsightData{
			{
				Name:   "page_follows",
				Period: "day",
				Values: []kafkamodels.FacebookInsightValue{
					{
						Value:   int64(1000),
						EndTime: "2024-01-15T08:00:00+0000",
					},
				},
			},
		},
	}

	result, err := ParseInsightsFromJSON(rawInsights, "page123", "workspace1")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestFacebookParser_calculateDerivedMetrics(t *testing.T) {
	parser := NewFacebookParser()

	parsed := &kafkamodels.ParsedFacebookInsights{
		ActiveUsers: 2400,
	}

	parser.calculateDerivedMetrics(parsed)

	if parsed.ActiveUsers != 100 {
		t.Fatalf("expected ActiveUsers 100 (2400/24), got %d", parsed.ActiveUsers)
	}

	parsedZero := &kafkamodels.ParsedFacebookInsights{
		ActiveUsers: 0,
	}

	parser.calculateDerivedMetrics(parsedZero)

	if parsedZero.ActiveUsers != 0 {
		t.Fatalf("expected ActiveUsers 0, got %d", parsedZero.ActiveUsers)
	}
}

func TestFacebookParser_getValueForDate(t *testing.T) {
	parser := NewFacebookParser()

	targetDate := time.Date(2024, 1, 15, 8, 0, 0, 0, time.UTC)

	cases := []struct {
		name     string
		values   []kafkamodels.FacebookInsightValue
		expected interface{}
	}{
		{
			name:     "empty values returns nil",
			values:   []kafkamodels.FacebookInsightValue{},
			expected: nil,
		},
		{
			name: "finds matching date RFC3339",
			values: []kafkamodels.FacebookInsightValue{
				{Value: int64(100), EndTime: "2024-01-15T08:00:00Z"},
			},
			expected: int64(100),
		},
		{
			name: "finds matching date Facebook format",
			values: []kafkamodels.FacebookInsightValue{
				{Value: int64(200), EndTime: "2024-01-15T08:00:00+0000"},
			},
			expected: int64(200),
		},
		{
			name: "no matching date returns nil",
			values: []kafkamodels.FacebookInsightValue{
				{Value: int64(100), EndTime: "2024-01-16T08:00:00+0000"},
			},
			expected: nil,
		},
		{
			name: "invalid date format skipped",
			values: []kafkamodels.FacebookInsightValue{
				{Value: int64(100), EndTime: "invalid-date"},
			},
			expected: nil,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parser.getValueForDate(tc.values, targetDate)
			if result != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestFacebookParser_applyInsightValue(t *testing.T) {
	parser := NewFacebookParser()

	cases := []struct {
		name       string
		metricName string
		value      interface{}
		checkFunc  func(*kafkamodels.ParsedFacebookInsights) bool
	}{
		{
			name:       "page_follows",
			metricName: "page_follows",
			value:      int64(1000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageFollows == 1000 },
		},
		{
			name:       "page_fans",
			metricName: "page_fans",
			value:      int64(5000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageFans == 5000 },
		},
		{
			name:       "page_views_total",
			metricName: "page_views_total",
			value:      int64(2000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageViews == 2000 },
		},
		{
			name:       "page_post_engagements",
			metricName: "page_post_engagements",
			value:      int64(500),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PagePostEngagements == 500 },
		},
		{
			name:       "page_video_views",
			metricName: "page_video_views",
			value:      int64(10000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageVideoViews == 10000 },
		},
		{
			name:       "page_total_actions",
			metricName: "page_total_actions",
			value:      int64(1500),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageTotalActions == 1500 },
		},
		{
			name:       "page_impressions_unique",
			metricName: "page_impressions_unique",
			value:      int64(8000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageImpressionsUnique == 8000 },
		},
		{
			name:       "page_media_view",
			metricName: "page_media_view",
			value:      int64(3000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageMediaView == 3000 && p.PageImpressions == 3000 },
		},
		{
			name:       "page_impressions_organic",
			metricName: "page_impressions_organic",
			value:      int64(6000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageImpressionsOrganic == 6000 },
		},
		{
			name:       "page_impressions_paid",
			metricName: "page_impressions_paid",
			value:      int64(2000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageImpressionsPaid == 2000 },
		},
		{
			name:       "page_fan_adds_unique",
			metricName: "page_fan_adds_unique",
			value:      int64(100),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageFanAddsUnique == 100 },
		},
		{
			name:       "page_fan_removes_unique",
			metricName: "page_fan_removes_unique",
			value:      int64(50),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageFanRemovesUnique == 50 },
		},
		{
			name:       "page_video_views_paid",
			metricName: "page_video_views_paid",
			value:      int64(1000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageVideoViewsPaid == 1000 },
		},
		{
			name:       "page_video_views_organic",
			metricName: "page_video_views_organic",
			value:      int64(9000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageVideoViewsOrganic == 9000 },
		},
		{
			name:       "page_video_views_autoplayed",
			metricName: "page_video_views_autoplayed",
			value:      int64(7000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageVideoViewsAutoplayed == 7000 },
		},
		{
			name:       "page_video_views_click_to_play",
			metricName: "page_video_views_click_to_play",
			value:      int64(3000),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageVideoViewsClickToPlay == 3000 },
		},
		{
			name:       "page_video_repeat_views",
			metricName: "page_video_repeat_views",
			value:      int64(500),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageVideoRepeatViews == 500 },
		},
		{
			name:       "page_actions_post_reactions_like_total",
			metricName: "page_actions_post_reactions_like_total",
			value:      int64(1200),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageActionsPostReactionsLikeTotal == 1200 && p.PositiveSentiment >= 1200 },
		},
		{
			name:       "page_actions_post_reactions_love_total",
			metricName: "page_actions_post_reactions_love_total",
			value:      int64(300),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageActionsPostReactionsLoveTotal == 300 },
		},
		{
			name:       "page_actions_post_reactions_anger_total",
			metricName: "page_actions_post_reactions_anger_total",
			value:      int64(10),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageActionsPostReactionsAngerTotal == 10 && p.NegativeSentiment == 10 },
		},
		{
			name:       "page_negative_feedback",
			metricName: "page_negative_feedback",
			value:      int64(25),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageNegativeFeedback == 25 },
		},
		{
			name:       "talking_about_count",
			metricName: "talking_about_count",
			value:      int64(500),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.TalkingAboutCount == 500 },
		},
		{
			name:       "page_category",
			metricName: "page_category",
			value:      "Technology",
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageCategory == "Technology" },
		},
		{
			name:       "page_impressions_organic_v2",
			metricName: "page_impressions_organic_v2",
			value:      int64(5500),
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return p.PageImpressionsOrganic == 5500 },
		},
		{
			name:       "page_fan_adds_by_paid_non_paid_unique",
			metricName: "page_fan_adds_by_paid_non_paid_unique",
			value:      map[string]interface{}{"paid": 100, "non_paid": 200},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PageFanAddsByPaidNonPaidUnique) == 2 && p.PageFansByLike == 300 },
		},
		{
			name:       "page_negative_feedback_by_type",
			metricName: "page_negative_feedback_by_type",
			value:      map[string]interface{}{"hide_all": 10, "hide_clicks": 20},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PageNegativeFeedbackByType) == 2 },
		},
		{
			name:       "page_positive_feedback_by_type",
			metricName: "page_positive_feedback_by_type",
			value:      map[string]interface{}{"like": 100, "comment": 50},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PagePositiveFeedbackByType) == 2 && p.PagePositiveFeedback == 150 },
		},
		{
			name:       "page_fans_online",
			metricName: "page_fans_online",
			value:      map[string]interface{}{"0": 100, "12": 200, "18": 150},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PageFansOnline) > 0 && p.ActiveUsers > 0 },
		},
		{
			name:       "page_fans_locale",
			metricName: "page_fans_locale",
			value:      map[string]interface{}{"en_US": 500, "es_ES": 200},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PageFansLocale) == 2 },
		},
		{
			name:       "page_fans_country",
			metricName: "page_fans_country",
			value:      map[string]interface{}{"US": 1000, "GB": 500},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PageFansCountry) == 2 },
		},
		{
			name:       "page_fans_city",
			metricName: "page_fans_city",
			value:      map[string]interface{}{"New York": 300, "Los Angeles": 200},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PageFansCity) == 2 },
		},
		{
			name:       "page_fans_gender_age",
			metricName: "page_fans_gender_age",
			value:      map[string]interface{}{"M.25-34": 200, "F.25-34": 150},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PageFansGenderAge) == 2 },
		},
		{
			name:       "page_fans_by_like_source_unique",
			metricName: "page_fans_by_like_source_unique",
			value:      map[string]interface{}{"page": 100, "ads": 50},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PageFansByLikeSourceUnique) == 2 && p.PageFansByLike == 150 },
		},
		{
			name:       "page_fans_by_unlike_source_unique",
			metricName: "page_fans_by_unlike_source_unique",
			value:      map[string]interface{}{"page": 10, "other": 5},
			checkFunc:  func(p *kafkamodels.ParsedFacebookInsights) bool { return len(p.PageFansByUnlikeSourceUnique) == 2 && p.PageFansByUnlike == 15 },
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			parsed := &kafkamodels.ParsedFacebookInsights{}
			parser.applyInsightValue(parsed, tc.metricName, tc.value)

			if !tc.checkFunc(parsed) {
				t.Fatalf("metric %s not applied correctly", tc.metricName)
			}
		})
	}
}

func TestFacebookParser_parsePostInsights(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID: "123456_789012",
		Insights: &struct {
			Data []struct {
				Name        string `json:"name"`
				Period      string `json:"period"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Values      []struct {
					Value   int    `json:"value"`
					EndTime string `json:"end_time"`
				} `json:"values"`
			} `json:"data"`
		}{
			Data: []struct {
				Name        string `json:"name"`
				Period      string `json:"period"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Values      []struct {
					Value   int    `json:"value"`
					EndTime string `json:"end_time"`
				} `json:"values"`
			}{
				{
					Name:   "post_impressions",
					Period: "lifetime",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 1000},
					},
				},
				{
					Name:   "post_impressions_unique",
					Period: "lifetime",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 800},
					},
				},
			},
		},
	}

	post := &kafkamodels.ParsedFacebookPost{}
	parser.parsePostInsights(rawPost, post)

	if post.PostImpressions != 1000 {
		t.Fatalf("expected PostImpressions 1000, got %d", post.PostImpressions)
	}

	if post.PostImpressionsUnique != 800 {
		t.Fatalf("expected PostImpressionsUnique 800, got %d", post.PostImpressionsUnique)
	}

	if post.TotalImpressions != 1800 {
		t.Fatalf("expected TotalImpressions 1800, got %d", post.TotalImpressions)
	}
}

func TestFacebookParser_ParsePostReactions(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID: "123456_789012",
		Total: &struct {
			Summary *struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		}{
			Summary: &struct {
				TotalCount int `json:"total_count"`
			}{
				TotalCount: 100,
			},
		},
		Like: &struct {
			Summary *struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		}{
			Summary: &struct {
				TotalCount int `json:"total_count"`
			}{
				TotalCount: 50,
			},
		},
		Love: &struct {
			Summary *struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		}{
			Summary: &struct {
				TotalCount int `json:"total_count"`
			}{
				TotalCount: 20,
			},
		},
		Haha: &struct {
			Summary *struct {
				TotalCount int `json:"total_count"`
			} `json:"summary"`
		}{
			Summary: &struct {
				TotalCount int `json:"total_count"`
			}{
				TotalCount: 10,
			},
		},
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
				TotalCount: 25,
				CanComment: true,
			},
		},
	}

	post := &kafkamodels.ParsedFacebookPost{}
	parser.ParsePostReactions(rawPost, post)

	if post.Total != 100 {
		t.Fatalf("expected Total 100, got %d", post.Total)
	}

	if post.Like != 50 {
		t.Fatalf("expected Like 50, got %d", post.Like)
	}

	if post.Love != 20 {
		t.Fatalf("expected Love 20, got %d", post.Love)
	}

	if post.Haha != 10 {
		t.Fatalf("expected Haha 10, got %d", post.Haha)
	}

	if post.Comments != 25 {
		t.Fatalf("expected Comments 25, got %d", post.Comments)
	}

	expectedEngagement := int64(100 + 25)
	if post.TotalEngagement != expectedEngagement {
		t.Fatalf("expected TotalEngagement %d, got %d", expectedEngagement, post.TotalEngagement)
	}
}

func TestFacebookParser_parseAttachments_Empty(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID:          "123456_789012",
		Attachments: nil,
	}

	assets := parser.parseAttachments(rawPost, "123456")

	if len(assets) != 0 {
		t.Fatalf("expected 0 assets, got %d", len(assets))
	}
}

func TestFacebookParser_parseChildAttachments(t *testing.T) {
	parser := NewFacebookParser()

	now := time.Now().UTC()
	rawPost := kafkamodels.RawFacebookPost{
		ID:          "123456_789012",
		Message:     "Carousel post",
		CreatedTime: kafkamodels.FacebookTime{Time: now},
		ChildAttachments: []struct {
			Type        string `json:"type"`
			MediaType   string `json:"media_type"`
			Description string `json:"description"`
			Media       *struct {
				Source string `json:"source"`
				Image  *struct {
					Height int    `json:"height"`
					Width  int    `json:"width"`
					Source string `json:"source"`
				} `json:"image"`
			} `json:"media"`
		}{
			{
				Type:        "photo",
				Description: "First image",
				Media: &struct {
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Source string `json:"source"`
					} `json:"image"`
				}{
					Image: &struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Source string `json:"source"`
					}{
						Source: "https://example.com/image1.jpg",
					},
				},
			},
			{
				Type:        "video",
				Description: "Second video",
				Media: &struct {
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Source string `json:"source"`
					} `json:"image"`
				}{
					Source: "https://example.com/video1.mp4",
				},
			},
		},
	}

	assets := parser.parseChildAttachments(rawPost, "123456")

	if len(assets) != 2 {
		t.Fatalf("expected 2 assets, got %d", len(assets))
	}

	if assets[0].AssetType != "photo" {
		t.Fatalf("expected first asset type 'photo', got %q", assets[0].AssetType)
	}

	if assets[1].AssetType != "video" {
		t.Fatalf("expected second asset type 'video', got %q", assets[1].AssetType)
	}

	if assets[0].PostID != "123456_789012" {
		t.Fatalf("expected PostID %q, got %q", "123456_789012", assets[0].PostID)
	}

	if !strings.Contains(assets[0].Link, "image1.jpg") {
		t.Fatalf("expected Link to contain 'image1.jpg', got %q", assets[0].Link)
	}
}

func TestFacebookParser_ParseVideo_WithVideoInsights(t *testing.T) {
	parser := NewFacebookParser()

	rawVideo := kafkamodels.RawFacebookVideo{
		ID:     "video123",
		PostID: "post456",
		CreatedTime: kafkamodels.FacebookTime{
			Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
		UpdatedTime: kafkamodels.FacebookTime{
			Time: time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC),
		},
	}

	// Populate VideoInsights using the correct inline struct structure
	rawVideo.VideoInsights.Data = []struct {
		Name   string `json:"name"`
		Period string `json:"period"`
		Values []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		} `json:"values"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}{
		{
			Name:   "total_video_views",
			Period: "lifetime",
			Values: []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			}{
				{Value: 5000, EndTime: "2024-01-15T12:00:00+0000"},
			},
		},
		{
			Name:   "total_video_views_unique",
			Period: "lifetime",
			Values: []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			}{
				{Value: 3000, EndTime: "2024-01-15T12:00:00+0000"},
			},
		},
		{
			Name:   "total_video_views_autoplayed",
			Period: "lifetime",
			Values: []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			}{
				{Value: 2000, EndTime: "2024-01-15T12:00:00+0000"},
			},
		},
		{
			Name:   "total_video_views_clicked_to_play",
			Period: "lifetime",
			Values: []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			}{
				{Value: 1000, EndTime: "2024-01-15T12:00:00+0000"},
			},
		},
		{
			Name:   "total_video_complete_views",
			Period: "lifetime",
			Values: []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			}{
				{Value: 800, EndTime: "2024-01-15T12:00:00+0000"},
			},
		},
		{
			Name:   "total_video_10s_views",
			Period: "lifetime",
			Values: []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			}{
				{Value: 1500, EndTime: "2024-01-15T12:00:00+0000"},
			},
		},
		{
			Name:   "total_video_avg_time_watched",
			Period: "lifetime",
			Values: []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			}{
				{Value: 45000, EndTime: "2024-01-15T12:00:00+0000"},
			},
		},
		{
			Name:   "total_video_impressions",
			Period: "lifetime",
			Values: []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			}{
				{Value: 10000, EndTime: "2024-01-15T12:00:00+0000"},
			},
		},
		{
			Name:   "blue_reels_play_count",
			Period: "lifetime",
			Values: []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			}{
				{Value: 7500, EndTime: "2024-01-15T12:00:00+0000"},
			},
		},
	}

	parsed, err := parser.ParseVideo(rawVideo, "page123", "Test Page")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.VideoID != "video123" {
		t.Fatalf("expected VideoID 'video123', got %q", parsed.VideoID)
	}

	if parsed.PostID != "page123_post456" {
		t.Fatalf("expected PostID 'page123_post456', got %q", parsed.PostID)
	}

	if parsed.TotalVideoViews != 5000 {
		t.Fatalf("expected TotalVideoViews 5000, got %d", parsed.TotalVideoViews)
	}

	if parsed.TotalVideoViewsUnique != 3000 {
		t.Fatalf("expected TotalVideoViewsUnique 3000, got %d", parsed.TotalVideoViewsUnique)
	}

	if parsed.TotalVideoCompleteViews != 800 {
		t.Fatalf("expected TotalVideoCompleteViews 800, got %d", parsed.TotalVideoCompleteViews)
	}

	if parsed.TotalVideo10sViews != 1500 {
		t.Fatalf("expected TotalVideo10sViews 1500, got %d", parsed.TotalVideo10sViews)
	}

	if parsed.BlueReelsPlayCount != 7500 {
		t.Fatalf("expected BlueReelsPlayCount 7500, got %d", parsed.BlueReelsPlayCount)
	}
}

func TestFacebookParser_ParseVideo_AllInsightMetrics(t *testing.T) {
	parser := NewFacebookParser()

	rawVideo := kafkamodels.RawFacebookVideo{
		ID:     "video789",
		PostID: "post101",
		CreatedTime: kafkamodels.FacebookTime{
			Time: time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		},
	}

	// Test all insight metric parsing
	rawVideo.VideoInsights.Data = []struct {
		Name   string `json:"name"`
		Period string `json:"period"`
		Values []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		} `json:"values"`
		Title       string `json:"title"`
		Description string `json:"description"`
	}{
		{Name: "total_video_views_organic", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 100}}},
		{Name: "total_video_views_organic_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 80}}},
		{Name: "total_video_views_paid", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 50}}},
		{Name: "total_video_views_paid_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 40}}},
		{Name: "total_video_views_sound_on", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 60}}},
		{Name: "total_video_complete_views_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 30}}},
		{Name: "total_video_complete_views_auto_played", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 25}}},
		{Name: "total_video_complete_views_clicked_to_play", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 20}}},
		{Name: "total_video_complete_views_organic", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 15}}},
		{Name: "total_video_complete_views_organic_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 12}}},
		{Name: "total_video_complete_views_paid", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 10}}},
		{Name: "total_video_complete_views_paid_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 8}}},
		{Name: "total_video_10s_views_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 70}}},
		{Name: "total_video_10s_views_auto_played", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 35}}},
		{Name: "total_video_10s_views_clicked_to_play", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 28}}},
		{Name: "total_video_10s_views_organic", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 22}}},
		{Name: "total_video_10s_views_paid", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 18}}},
		{Name: "total_video_10s_views_sound_on", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 14}}},
		{Name: "total_video_15s_views", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 65}}},
		{Name: "total_video_60s_excludes_shorter_views", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 55}}},
		{Name: "post_video_avg_time_watched", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 30000}}},
		{Name: "total_video_view_total_time", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 500000}}},
		{Name: "total_video_view_total_time_organic", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 400000}}},
		{Name: "total_video_view_total_time_paid", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 100000}}},
		{Name: "post_video_view_time", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 250000}}},
		{Name: "total_video_impressions_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 8000}}},
		{Name: "total_video_impressions_paid", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 2000}}},
		{Name: "total_video_impressions_paid_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 1500}}},
		{Name: "total_video_impressions_organic", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 6000}}},
		{Name: "total_video_impressions_organic_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 5500}}},
		{Name: "total_video_impressions_viral", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 1000}}},
		{Name: "total_video_impressions_viral_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 800}}},
		{Name: "total_video_impressions_fan", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 3000}}},
		{Name: "total_video_impressions_fan_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 2500}}},
		{Name: "total_video_impressions_fan_paid", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 500}}},
		{Name: "total_video_impressions_fan_paid_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 400}}},
		{Name: "post_impressions_unique", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 9000}}},
		{Name: "total_video_retention_graph", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: "retention_data"}}},
		{Name: "total_video_view_total_time_live", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 50000}}},
		{Name: "total_video_views_live", Values: []struct {
			Value   interface{} `json:"value"`
			EndTime string      `json:"end_time"`
		}{{Value: 200}}},
	}

	parsed, err := parser.ParseVideo(rawVideo, "page123", "Test Page")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.TotalVideoViewsOrganic != 100 {
		t.Fatalf("expected TotalVideoViewsOrganic 100, got %d", parsed.TotalVideoViewsOrganic)
	}

	if parsed.TotalVideoViewsPaid != 50 {
		t.Fatalf("expected TotalVideoViewsPaid 50, got %d", parsed.TotalVideoViewsPaid)
	}

	if parsed.TotalVideo15sViews != 65 {
		t.Fatalf("expected TotalVideo15sViews 65, got %d", parsed.TotalVideo15sViews)
	}

	if parsed.TotalVideoViewTotalTime != 500000 {
		t.Fatalf("expected TotalVideoViewTotalTime 500000, got %d", parsed.TotalVideoViewTotalTime)
	}

	if parsed.TotalVideoImpressionsViral != 1000 {
		t.Fatalf("expected TotalVideoImpressionsViral 1000, got %d", parsed.TotalVideoImpressionsViral)
	}

	if parsed.TotalVideoViewTotalTimeLive != 50000 {
		t.Fatalf("expected TotalVideoViewTotalTimeLive 50000, got %d", parsed.TotalVideoViewTotalTimeLive)
	}

	if parsed.TotalVideoViewsLive != 200 {
		t.Fatalf("expected TotalVideoViewsLive 200, got %d", parsed.TotalVideoViewsLive)
	}
}

func TestFacebookParser_parseAttachments_WithPhoto(t *testing.T) {
	parser := NewFacebookParser()

	createdTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	rawPost := kafkamodels.RawFacebookPost{
		ID:      "123456_789012",
		Message: "Test post message",
		CreatedTime: kafkamodels.FacebookTime{
			Time: createdTime,
		},
	}

	// Set attachments using the correct struct type
	rawPost.Attachments = &struct {
		Data []struct {
			Type        string `json:"type"`
			MediaType   string `json:"media_type"`
			Caption     string `json:"caption"`
			Description string `json:"description"`
			Link        string `json:"link"`
			Target      *struct {
				ID string `json:"id"`
			} `json:"target"`
			Media *struct {
				Src    string `json:"src,omitempty"`
				Source string `json:"source"`
				Image  *struct {
					Height int    `json:"height"`
					Width  int    `json:"width"`
					Src    string `json:"src"`
					Source string `json:"source"`
				} `json:"image"`
			} `json:"media"`
			Subattachments *struct {
				Data []struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type"`
					Media     *struct {
						Src    string `json:"src"`
						Source string `json:"source"`
						Image  *struct {
							Height int    `json:"height"`
							Width  int    `json:"width"`
							Src    string `json:"src"`
							Source string `json:"source"`
						} `json:"image"`
					} `json:"media"`
				} `json:"data"`
			} `json:"subattachments"`
		} `json:"data"`
	}{
		Data: []struct {
			Type        string `json:"type"`
			MediaType   string `json:"media_type"`
			Caption     string `json:"caption"`
			Description string `json:"description"`
			Link        string `json:"link"`
			Target      *struct {
				ID string `json:"id"`
			} `json:"target"`
			Media *struct {
				Src    string `json:"src,omitempty"`
				Source string `json:"source"`
				Image  *struct {
					Height int    `json:"height"`
					Width  int    `json:"width"`
					Src    string `json:"src"`
					Source string `json:"source"`
				} `json:"image"`
			} `json:"media"`
			Subattachments *struct {
				Data []struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type"`
					Media     *struct {
						Src    string `json:"src"`
						Source string `json:"source"`
						Image  *struct {
							Height int    `json:"height"`
							Width  int    `json:"width"`
							Src    string `json:"src"`
							Source string `json:"source"`
						} `json:"image"`
					} `json:"media"`
				} `json:"data"`
			} `json:"subattachments"`
		}{
			{
				Type:        "photo",
				MediaType:   "photo",
				Description: "Photo attachment",
				Link:        "https://example.com/photo",
				Target: &struct {
					ID string `json:"id"`
				}{
					ID: "media123",
				},
				Media: &struct {
					Src    string `json:"src,omitempty"`
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Src    string `json:"src"`
						Source string `json:"source"`
					} `json:"image"`
				}{
					Image: &struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Src    string `json:"src"`
						Source string `json:"source"`
					}{
						Source: "https://example.com/image.jpg",
					},
				},
			},
		},
	}

	assets := parser.parseAttachments(rawPost, "123456")

	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}

	if assets[0].PostID != "123456_789012" {
		t.Fatalf("expected PostID '123456_789012', got %q", assets[0].PostID)
	}

	if assets[0].MediaID != "media123" {
		t.Fatalf("expected MediaID 'media123', got %q", assets[0].MediaID)
	}

	if assets[0].AssetType != "photo" {
		t.Fatalf("expected AssetType 'photo', got %q", assets[0].AssetType)
	}

	if assets[0].CallToAction != "https://example.com/photo" {
		t.Fatalf("expected CallToAction 'https://example.com/photo', got %q", assets[0].CallToAction)
	}
}

func TestFacebookParser_parseAttachments_WithVideo(t *testing.T) {
	parser := NewFacebookParser()

	createdTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	rawPost := kafkamodels.RawFacebookPost{
		ID:      "123456_789012",
		Message: "Video post",
		CreatedTime: kafkamodels.FacebookTime{
			Time: createdTime,
		},
	}

	rawPost.Attachments = &struct {
		Data []struct {
			Type        string `json:"type"`
			MediaType   string `json:"media_type"`
			Caption     string `json:"caption"`
			Description string `json:"description"`
			Link        string `json:"link"`
			Target      *struct {
				ID string `json:"id"`
			} `json:"target"`
			Media *struct {
				Src    string `json:"src,omitempty"`
				Source string `json:"source"`
				Image  *struct {
					Height int    `json:"height"`
					Width  int    `json:"width"`
					Src    string `json:"src"`
					Source string `json:"source"`
				} `json:"image"`
			} `json:"media"`
			Subattachments *struct {
				Data []struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type"`
					Media     *struct {
						Src    string `json:"src"`
						Source string `json:"source"`
						Image  *struct {
							Height int    `json:"height"`
							Width  int    `json:"width"`
							Src    string `json:"src"`
							Source string `json:"source"`
						} `json:"image"`
					} `json:"media"`
				} `json:"data"`
			} `json:"subattachments"`
		} `json:"data"`
	}{
		Data: []struct {
			Type        string `json:"type"`
			MediaType   string `json:"media_type"`
			Caption     string `json:"caption"`
			Description string `json:"description"`
			Link        string `json:"link"`
			Target      *struct {
				ID string `json:"id"`
			} `json:"target"`
			Media *struct {
				Src    string `json:"src,omitempty"`
				Source string `json:"source"`
				Image  *struct {
					Height int    `json:"height"`
					Width  int    `json:"width"`
					Src    string `json:"src"`
					Source string `json:"source"`
				} `json:"image"`
			} `json:"media"`
			Subattachments *struct {
				Data []struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type"`
					Media     *struct {
						Src    string `json:"src"`
						Source string `json:"source"`
						Image  *struct {
							Height int    `json:"height"`
							Width  int    `json:"width"`
							Src    string `json:"src"`
							Source string `json:"source"`
						} `json:"image"`
					} `json:"media"`
				} `json:"data"`
			} `json:"subattachments"`
		}{
			{
				Type:      "video_inline",
				MediaType: "video",
				Media: &struct {
					Src    string `json:"src,omitempty"`
					Source string `json:"source"`
					Image  *struct {
						Height int    `json:"height"`
						Width  int    `json:"width"`
						Src    string `json:"src"`
						Source string `json:"source"`
					} `json:"image"`
				}{
					Source: "https://example.com/video.mp4",
				},
			},
		},
	}

	assets := parser.parseAttachments(rawPost, "123456")

	if len(assets) != 1 {
		t.Fatalf("expected 1 asset, got %d", len(assets))
	}

	if assets[0].AssetType != "video" {
		t.Fatalf("expected AssetType 'video', got %q", assets[0].AssetType)
	}

	if assets[0].Link != "https://example.com/video.mp4" {
		t.Fatalf("expected video Link, got %q", assets[0].Link)
	}
}

func TestFacebookParser_parseAttachments_WithSubattachments(t *testing.T) {
	parser := NewFacebookParser()

	createdTime := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	rawPost := kafkamodels.RawFacebookPost{
		ID:      "123456_789012",
		Message: "Carousel post",
		CreatedTime: kafkamodels.FacebookTime{
			Time: createdTime,
		},
	}

	rawPost.Attachments = &struct {
		Data []struct {
			Type        string `json:"type"`
			MediaType   string `json:"media_type"`
			Caption     string `json:"caption"`
			Description string `json:"description"`
			Link        string `json:"link"`
			Target      *struct {
				ID string `json:"id"`
			} `json:"target"`
			Media *struct {
				Src    string `json:"src,omitempty"`
				Source string `json:"source"`
				Image  *struct {
					Height int    `json:"height"`
					Width  int    `json:"width"`
					Src    string `json:"src"`
					Source string `json:"source"`
				} `json:"image"`
			} `json:"media"`
			Subattachments *struct {
				Data []struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type"`
					Media     *struct {
						Src    string `json:"src"`
						Source string `json:"source"`
						Image  *struct {
							Height int    `json:"height"`
							Width  int    `json:"width"`
							Src    string `json:"src"`
							Source string `json:"source"`
						} `json:"image"`
					} `json:"media"`
				} `json:"data"`
			} `json:"subattachments"`
		} `json:"data"`
	}{
		Data: []struct {
			Type        string `json:"type"`
			MediaType   string `json:"media_type"`
			Caption     string `json:"caption"`
			Description string `json:"description"`
			Link        string `json:"link"`
			Target      *struct {
				ID string `json:"id"`
			} `json:"target"`
			Media *struct {
				Src    string `json:"src,omitempty"`
				Source string `json:"source"`
				Image  *struct {
					Height int    `json:"height"`
					Width  int    `json:"width"`
					Src    string `json:"src"`
					Source string `json:"source"`
				} `json:"image"`
			} `json:"media"`
			Subattachments *struct {
				Data []struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type"`
					Media     *struct {
						Src    string `json:"src"`
						Source string `json:"source"`
						Image  *struct {
							Height int    `json:"height"`
							Width  int    `json:"width"`
							Src    string `json:"src"`
							Source string `json:"source"`
						} `json:"image"`
					} `json:"media"`
				} `json:"data"`
			} `json:"subattachments"`
		}{
			{
				Type: "multi_share_no_end_card",
				Subattachments: &struct {
					Data []struct {
						Type      string `json:"type"`
						MediaType string `json:"media_type"`
						Media     *struct {
							Src    string `json:"src"`
							Source string `json:"source"`
							Image  *struct {
								Height int    `json:"height"`
								Width  int    `json:"width"`
								Src    string `json:"src"`
								Source string `json:"source"`
							} `json:"image"`
						} `json:"media"`
					} `json:"data"`
				}{
					Data: []struct {
						Type      string `json:"type"`
						MediaType string `json:"media_type"`
						Media     *struct {
							Src    string `json:"src"`
							Source string `json:"source"`
							Image  *struct {
								Height int    `json:"height"`
								Width  int    `json:"width"`
								Src    string `json:"src"`
								Source string `json:"source"`
							} `json:"image"`
						} `json:"media"`
					}{
						{
							Type: "photo",
							Media: &struct {
								Src    string `json:"src"`
								Source string `json:"source"`
								Image  *struct {
									Height int    `json:"height"`
									Width  int    `json:"width"`
									Src    string `json:"src"`
									Source string `json:"source"`
								} `json:"image"`
							}{
								Image: &struct {
									Height int    `json:"height"`
									Width  int    `json:"width"`
									Src    string `json:"src"`
									Source string `json:"source"`
								}{
									Source: "https://example.com/sub1.jpg",
								},
							},
						},
						{
							Type: "video",
							Media: &struct {
								Src    string `json:"src"`
								Source string `json:"source"`
								Image  *struct {
									Height int    `json:"height"`
									Width  int    `json:"width"`
									Src    string `json:"src"`
									Source string `json:"source"`
								} `json:"image"`
							}{
								Source: "https://example.com/sub2.mp4",
							},
						},
					},
				},
			},
		},
	}

	assets := parser.parseAttachments(rawPost, "123456")

	if len(assets) != 3 {
		t.Fatalf("expected 3 assets (1 main + 2 subattachments), got %d", len(assets))
	}

	if assets[0].AssetType != "photo" {
		t.Fatalf("expected main AssetType 'photo', got %q", assets[0].AssetType)
	}

	if assets[1].AssetType != "photo" {
		t.Fatalf("expected first subattachment AssetType 'photo', got %q", assets[1].AssetType)
	}

	if assets[2].AssetType != "video" {
		t.Fatalf("expected second subattachment AssetType 'video', got %q", assets[2].AssetType)
	}
}

func TestFacebookParser_parsePostInsights_AllMetrics(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID: "123456_789012",
		Insights: &struct {
			Data []struct {
				Name        string `json:"name"`
				Period      string `json:"period"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Values      []struct {
					Value   int    `json:"value"`
					EndTime string `json:"end_time"`
				} `json:"values"`
			} `json:"data"`
		}{
			Data: []struct {
				Name        string `json:"name"`
				Period      string `json:"period"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Values      []struct {
					Value   int    `json:"value"`
					EndTime string `json:"end_time"`
				} `json:"values"`
			}{
				{
					Name: "post_impressions",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 1000},
					},
				},
				{
					Name: "post_impressions_unique",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 800},
					},
				},
				{
					Name: "post_impressions_paid",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 200},
					},
				},
				{
					Name: "post_impressions_paid_unique",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 150},
					},
				},
				{
					Name: "post_impressions_organic",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 700},
					},
				},
				{
					Name: "post_impressions_organic_unique",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 600},
					},
				},
				{
					Name: "post_impressions_viral",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 100},
					},
				},
				{
					Name: "post_impressions_viral_unique",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 80},
					},
				},
				{
					Name: "post_clicks",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 50},
					},
				},
				{
					Name: "post_video_views",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 500},
					},
				},
				{
					Name: "post_media_view",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{
						{Value: 300},
					},
				},
			},
		},
	}

	post := &kafkamodels.ParsedFacebookPost{}
	parser.parsePostInsights(rawPost, post)

	if post.PostImpressions != 300 {
		t.Fatalf("expected PostImpressions 300 (last assigned from post_media_view), got %d", post.PostImpressions)
	}
	if post.PostImpressionsUnique != 800 {
		t.Fatalf("expected PostImpressionsUnique 800, got %d", post.PostImpressionsUnique)
	}
	if post.PostImpressionsPaid != 200 {
		t.Fatalf("expected PostImpressionsPaid 200, got %d", post.PostImpressionsPaid)
	}
	if post.PostImpressionsPaidUnique != 150 {
		t.Fatalf("expected PostImpressionsPaidUnique 150, got %d", post.PostImpressionsPaidUnique)
	}
	if post.PostImpressionsOrganic != 700 {
		t.Fatalf("expected PostImpressionsOrganic 700, got %d", post.PostImpressionsOrganic)
	}
	if post.PostImpressionsOrganicUnique != 600 {
		t.Fatalf("expected PostImpressionsOrganicUnique 600, got %d", post.PostImpressionsOrganicUnique)
	}
	if post.PostImpressionsViral != 100 {
		t.Fatalf("expected PostImpressionsViral 100, got %d", post.PostImpressionsViral)
	}
	if post.PostImpressionsViralUnique != 80 {
		t.Fatalf("expected PostImpressionsViralUnique 80, got %d", post.PostImpressionsViralUnique)
	}
	if post.PostClicks != 50 {
		t.Fatalf("expected PostClicks 50, got %d", post.PostClicks)
	}
	if post.PostVideoViews != 500 {
		t.Fatalf("expected PostVideoViews 500, got %d", post.PostVideoViews)
	}
}

func TestFacebookParser_parsePostInsights_EmptyInsights(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID:       "123456_789012",
		Insights: nil,
	}

	post := &kafkamodels.ParsedFacebookPost{}
	parser.parsePostInsights(rawPost, post)

	if post.PostImpressions != 0 {
		t.Fatalf("expected PostImpressions 0, got %d", post.PostImpressions)
	}
}

func TestFacebookParser_parsePostInsights_EmptyValues(t *testing.T) {
	parser := NewFacebookParser()

	rawPost := kafkamodels.RawFacebookPost{
		ID: "123456_789012",
		Insights: &struct {
			Data []struct {
				Name        string `json:"name"`
				Period      string `json:"period"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Values      []struct {
					Value   int    `json:"value"`
					EndTime string `json:"end_time"`
				} `json:"values"`
			} `json:"data"`
		}{
			Data: []struct {
				Name        string `json:"name"`
				Period      string `json:"period"`
				Title       string `json:"title"`
				Description string `json:"description"`
				Values      []struct {
					Value   int    `json:"value"`
					EndTime string `json:"end_time"`
				} `json:"values"`
			}{
				{
					Name: "post_impressions",
					Values: []struct {
						Value   int    `json:"value"`
						EndTime string `json:"end_time"`
					}{},
				},
			},
		},
	}

	post := &kafkamodels.ParsedFacebookPost{}
	parser.parsePostInsights(rawPost, post)

	if post.PostImpressions != 0 {
		t.Fatalf("expected PostImpressions 0, got %d", post.PostImpressions)
	}
}

func TestFacebookParser_processInsightMetric(t *testing.T) {
	parser := NewFacebookParser()

	insight := kafkamodels.FacebookInsightData{
		Name:   "page_media_view",
		Period: "day",
		Values: []kafkamodels.FacebookInsightValue{
			{
				Value:   1000,
				EndTime: "2024-01-15T08:00:00Z",
			},
		},
	}

	targetDate := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)

	parsed := &kafkamodels.ParsedFacebookInsights{}
	parser.processInsightMetric(parsed, insight, targetDate)

	if parsed.PageImpressions != 1000 {
		t.Fatalf("expected PageImpressions 1000, got %d", parsed.PageImpressions)
	}
}
