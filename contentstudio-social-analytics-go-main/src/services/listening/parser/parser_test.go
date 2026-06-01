package parser

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// --- mocks ---

type mockDedup struct {
	seen map[string]bool
	err  error
}

type mockTopicStatusChecker struct {
	topic *mongomodels.ListeningTopic
	err   error
}

func newMockDedup() *mockDedup {
	return &mockDedup{seen: make(map[string]bool)}
}

func (m *mockDedup) SetNX(_ context.Context, key string, _ interface{}, _ time.Duration) (bool, error) {
	if m.err != nil {
		return false, m.err
	}

	if m.seen[key] {
		return false, nil
	}
	m.seen[key] = true
	return true, nil
}

func (m *mockTopicStatusChecker) GetTopicByID(_ context.Context, _ string) (*mongomodels.ListeningTopic, error) {
	return m.topic, m.err
}

func newTestParser(prod *mockProducerRecorder, dedup *mockDedup) *ParserService {
	log, _ := logger.NewTestLogger()
	return NewParserService(prod, dedup, nil, log, 48)
}

func makeRawPayload(platform string, items []map[string]interface{}, order kafkamodels.ListeningWorkOrder) []byte {
	rawItems, _ := json.Marshal(items)
	payload := kafkamodels.ListeningRawPayload{
		TopicID:   order.TopicID,
		Platform:  platform,
		Keyword:   "test",
		RawData:   rawItems,
		WorkOrder: order,
	}
	data, _ := json.Marshal(payload)
	return data
}

// --- normalization tests ---

func TestNormalizeMention(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		platform string
		topicID  string
		raw      string
		check    func(*testing.T, *kafkamodels.ListeningMention)
		wantErr  bool
	}{
		{
			name:     "twitter basic fields",
			platform: "twitter",
			topicID:  "topic-1",
			raw: `{
				"id": "12345",
				"text": "Hello world from Twitter",
				"author_id": "user1",
				"author_name": "John",
				"url": "https://twitter.com/status/12345",
				"created_at": "2025-01-15T10:00:00",
				"like_count": 10,
				"comment_count": 5,
				"share_count": 3,
				"view_count": 100
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.MentionID != "twitter:12345" {
					t.Errorf("mention_id: want %q, got %q", "twitter:12345", m.MentionID)
				}
				if m.Platform != "twitter" {
					t.Errorf("platform: want %q, got %q", "twitter", m.Platform)
				}
				if m.PostText != "Hello world from Twitter" {
					t.Errorf("post_text: got %q", m.PostText)
				}
				if m.TotalEngagement != 118 {
					t.Errorf("total_engagement: want 118, got %d", m.TotalEngagement)
				}
				if m.ContentHash == "" {
					t.Error("content_hash should not be empty")
				}
			},
		},
		{
			name:     "twitter data365 shape with nested author and fallback URL",
			platform: "twitter",
			topicID:  "topic-1",
			raw: `{
				"id": "2029325560820703698",
				"body": "Hello from X",
				"published": "2025-01-15T10:00:00Z",
				"likes_count": 10,
				"reply_count": 5,
				"retweet_count": 3,
				"user_id": "user1",
				"author": { "handle": "korrssk" }
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.AuthorID != "user1" {
					t.Errorf("author_id: want %q, got %q", "user1", m.AuthorID)
				}
				if m.AuthorHandle != "korrssk" {
					t.Errorf("author_handle: want %q, got %q", "korrssk", m.AuthorHandle)
				}
				if m.AuthorURL != "https://x.com/korrssk" {
					t.Errorf("author_url: got %q", m.AuthorURL)
				}
				if m.URL != "https://x.com/korrssk/status/2029325560820703698" {
					t.Errorf("url: got %q", m.URL)
				}
				if m.TotalEngagement != 18 {
					t.Errorf("total_engagement: want 18, got %d", m.TotalEngagement)
				}
			},
		},
		{
			name:     "twitter extracts language and followers",
			platform: "twitter",
			topicID:  "topic-1",
			raw: `{
				"id": "2029325560820703699",
				"body": "Hello from X about #OpenAI and #Golang",
				"published": "2025-01-15T10:00:00Z",
				"lang": "en",
				"author": {
					"handle": "korrssk",
					"followers_count": "4200"
				},
				"hashtags": [{"tag": "OpenAI"}, {"tag": "Golang"}]
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.Language != "en" {
					t.Errorf("language: want %q, got %q", "en", m.Language)
				}
				if m.AuthorFollowers != 4200 {
					t.Errorf("author_followers: want 4200, got %d", m.AuthorFollowers)
				}
				if len(m.AITags) != 0 {
					t.Errorf("ai_tags: want empty, got %v", m.AITags)
				}
			},
		},
		{
			name:     "twitter fallback URL without handle uses author_id",
			platform: "twitter",
			topicID:  "topic-1",
			raw: `{
				"id": "2029325560820703698",
				"body": "Hello from X",
				"user_id": "2029325572719972759"
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.AuthorID != "2029325572719972759" {
					t.Errorf("author_id: got %q", m.AuthorID)
				}
				if m.URL != "https://x.com/2029325572719972759/status/2029325560820703698" {
					t.Errorf("url: got %q", m.URL)
				}
				if m.AuthorURL != "https://x.com/2029325572719972759" {
					t.Errorf("author_url: got %q", m.AuthorURL)
				}
			},
		},
		{
			name:     "instagram basic fields",
			platform: "instagram",
			topicID:  "topic-2",
			raw: `{
				"id": "ig_999",
				"caption": "Beautiful sunset #travel",
				"username": "photographer",
				"permalink": "https://instagram.com/p/ig_999",
				"timestamp": 1705312800,
				"likes": 50,
				"comments": 12,
				"is_video": false
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.PostText != "Beautiful sunset #travel" {
					t.Errorf("post_text: got %q", m.PostText)
				}
				if m.AuthorID != "photographer" {
					t.Errorf("author_id: got %q", m.AuthorID)
				}
				if m.URL != "https://instagram.com/p/ig_999" {
					t.Errorf("url: got %q", m.URL)
				}
				if m.MediaType != "text" {
					t.Errorf("media_type: want %q, got %q", "text", m.MediaType)
				}
			},
		},
		{
			name:     "instagram data365 shape",
			platform: "instagram",
			topicID:  "topic-2",
			raw: `{
				"id": "ig_1000",
				"text": "Beautiful sunset #travel",
				"owner_id": "ig_owner_1",
				"owner_username": "photographer",
				"attached_image_url": "https://cdn.example.com/ig.jpg",
				"created_time": "2025-01-15T10:00:00Z",
				"likes_count": 50,
				"comments_count": 12,
				"is_video": false
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.AuthorID != "ig_owner_1" {
					t.Errorf("author_id: got %q", m.AuthorID)
				}
				if m.AuthorHandle != "photographer" {
					t.Errorf("author_handle: got %q", m.AuthorHandle)
				}
				if m.AuthorURL != "https://www.instagram.com/photographer/" {
					t.Errorf("author_url: got %q", m.AuthorURL)
				}
				if m.LikesCount != 50 {
					t.Errorf("likes_count: want 50, got %d", m.LikesCount)
				}
				if m.CommentsCount != 12 {
					t.Errorf("comments_count: want 12, got %d", m.CommentsCount)
				}
				if m.TotalEngagement != 62 {
					t.Errorf("total_engagement: want 62, got %d", m.TotalEngagement)
				}
				if len(m.MediaURLs) != 1 || m.MediaURLs[0] != "https://cdn.example.com/ig.jpg" {
					t.Errorf("media_urls: got %#v", m.MediaURLs)
				}
			},
		},
		{
			name:     "tiktok is_video sets content_type",
			platform: "tiktok",
			topicID:  "topic-3",
			raw: `{
				"id": "tt_777",
				"description": "Cool dance video",
				"author_name": "dancer",
				"url": "https://tiktok.com/v/tt_777",
				"created_at": "2025-03-01T12:00:00",
				"play_count": 5000,
				"like_count": 200,
				"comment_count": 30,
				"share_count": 50,
				"is_video": true
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.ContentType != "video" {
					t.Errorf("content_type: want %q, got %q", "video", m.ContentType)
				}
				if m.TotalEngagement != 5280 {
					t.Errorf("total_engagement: want 5280, got %d", m.TotalEngagement)
				}
			},
		},
		{
			name:     "tiktok fallback URL with handle",
			platform: "tiktok",
			topicID:  "topic-3",
			raw: `{
				"id": "7483920192019201920",
				"description": "Fallback TikTok URL",
				"username": "korrssk"
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.URL != "https://www.tiktok.com/@korrssk/video/7483920192019201920" {
					t.Errorf("url: got %q", m.URL)
				}
			},
		},
		{
			name:     "tiktok fallback URL without handle uses embed",
			platform: "tiktok",
			topicID:  "topic-3",
			raw: `{
				"id": "7483920192019201920",
				"description": "Fallback TikTok URL"
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.URL != "https://www.tiktok.com/embed/v2/7483920192019201920" {
					t.Errorf("url: got %q", m.URL)
				}
			},
		},
		{
			name:     "reddit text takes priority over title",
			platform: "reddit",
			topicID:  "topic-4",
			raw: `{
				"id": "rd_abc",
				"title": "Discussion about Go",
				"text": "I think Go is great for backend",
				"author_name": "redditor",
				"url": "https://reddit.com/r/golang/rd_abc",
				"created_at": "2025-02-20T08:00:00",
				"score": 42,
				"comments": 15
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.PostText != "I think Go is great for backend" {
					t.Errorf("post_text: got %q", m.PostText)
				}
				if m.TotalEngagement != 57 {
					t.Errorf("total_engagement: want 57, got %d", m.TotalEngagement)
				}
			},
		},
		{
			name:     "reddit attached_image_url as array",
			platform: "reddit",
			topicID:  "topic-4",
			raw: `{
				"id": "rd_img",
				"title": "Discussion with image",
				"text": "Post with reddit image array",
				"author_name": "redditor",
				"attached_image_url": ["https://reddit.example.com/one.jpg", "https://reddit.example.com/two.jpg"],
				"created_at": "2025-02-20T08:00:00Z"
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if len(m.MediaURLs) != 2 {
					t.Fatalf("media_urls: want 2, got %d (%#v)", len(m.MediaURLs), m.MediaURLs)
				}
				if m.MediaURLs[0] != "https://reddit.example.com/one.jpg" || m.MediaURLs[1] != "https://reddit.example.com/two.jpg" {
					t.Errorf("media_urls: got %#v", m.MediaURLs)
				}
			},
		},
		{
			name:     "facebook basic fields",
			platform: "facebook",
			topicID:  "topic-5",
			raw: `{
				"id": "fb_001",
				"text": "Facebook post about Go",
				"author_id": "fb_user",
				"author_name": "FB User",
				"url": "https://facebook.com/post/fb_001",
				"created_at": "2025-01-10T09:00:00",
				"likes": 20,
				"comments": 8,
				"shares": 4
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.MentionID != "facebook:fb_001" {
					t.Errorf("mention_id: want %q, got %q", "facebook:fb_001", m.MentionID)
				}
				if m.TotalEngagement != 32 {
					t.Errorf("total_engagement: want 32, got %d", m.TotalEngagement)
				}
			},
		},
		{
			name:     "facebook data365 shape with reactions",
			platform: "facebook",
			topicID:  "topic-5",
			raw: `{
				"id": "fb_002",
				"text": "Facebook post from Data365",
				"owner_id": "page_123",
				"owner_username": "brand.page",
				"owner_full_name": "Brand Page",
				"attached_image_url": "https://cdn.example.com/post.jpg",
				"created_time": "2025-01-10T09:00:00Z",
				"comments_count": 8,
				"shares_count": 4,
				"reactions_total_count": 20,
				"reactions_like_count": 11
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.AuthorID != "page_123" {
					t.Errorf("author_id: got %q", m.AuthorID)
				}
				if m.AuthorName != "Brand Page" {
					t.Errorf("author_name: got %q", m.AuthorName)
				}
				if m.AuthorHandle != "brand.page" {
					t.Errorf("author_handle: got %q", m.AuthorHandle)
				}
				if m.TotalEngagement != 32 {
					t.Errorf("total_engagement: want 32, got %d", m.TotalEngagement)
				}
				if m.CommentsCount != 8 {
					t.Errorf("comments_count: want 8, got %d", m.CommentsCount)
				}
				if m.SharesCount != 4 {
					t.Errorf("shares_count: want 4, got %d", m.SharesCount)
				}
				if m.LikesCount != 11 {
					t.Errorf("likes_count: want 11, got %d", m.LikesCount)
				}
				if got := m.PostedAt.Format(time.RFC3339); got != "2025-01-10T09:00:00Z" {
					t.Errorf("posted_at: got %s", got)
				}
				if len(m.MediaURLs) != 1 || m.MediaURLs[0] != "https://cdn.example.com/post.jpg" {
					t.Errorf("media_urls: got %#v", m.MediaURLs)
				}
			},
		},
		{
			name:     "threads basic fields",
			platform: "threads",
			topicID:  "topic-6",
			raw: `{
				"id": "th_555",
				"text": "Threads post about tech",
				"username": "threader",
				"url": "https://threads.net/th_555",
				"created_at": "2025-04-01T14:00:00",
				"likes": 15,
				"comments": 3
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.Platform != "threads" {
					t.Errorf("platform: want %q, got %q", "threads", m.Platform)
				}
				if m.AuthorID != "threader" {
					t.Errorf("author_id: got %q", m.AuthorID)
				}
			},
		},
		{
			name:     "threads data365 shape with numeric owner_id",
			platform: "threads",
			topicID:  "topic-6",
			raw: `{
				"id": 123456789,
				"text": "Threads post about tech",
				"created_time": "2025-04-01T14:00:00Z",
				"owner_id": 123456789,
				"owner_username": "threader",
				"owner_full_name": "Thread User",
				"owner_profile_pic_url": "https://cdn.example.com/threader.jpg",
				"likes_count": 15,
				"comments_count": 3,
				"reposts_count": 2,
				"quotes_count": 1
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.AuthorID != "123456789" {
					t.Errorf("author_id: want %q, got %q", "123456789", m.AuthorID)
				}
				if m.AuthorHandle != "threader" {
					t.Errorf("author_handle: got %q", m.AuthorHandle)
				}
				if m.AuthorURL != "https://www.threads.net/@threader" {
					t.Errorf("author_url: got %q", m.AuthorURL)
				}
				if m.AuthorImageURL != "https://cdn.example.com/threader.jpg" {
					t.Errorf("author_image_url: got %q", m.AuthorImageURL)
				}
				if m.LikesCount != 15 {
					t.Errorf("likes_count: want 15, got %d", m.LikesCount)
				}
				if m.CommentsCount != 3 {
					t.Errorf("comments_count: want 3, got %d", m.CommentsCount)
				}
				if m.SharesCount != 2 {
					t.Errorf("shares_count (reposts): want 2, got %d", m.SharesCount)
				}
				if m.TotalEngagement != 20 {
					t.Errorf("total_engagement: want 20, got %d", m.TotalEngagement)
				}
			},
		},
		{
			name:     "threads numeric owner_id coerced to string",
			platform: "threads",
			topicID:  "topic-6",
			raw: `{
				"id": "th_556",
				"text": "Threads post with numeric owner id",
				"owner_id": 123456789,
				"owner_username": "threader",
				"created_time": "2025-04-01T14:00:00Z"
			}`,
			check: func(t *testing.T, m *kafkamodels.ListeningMention) {
				t.Helper()
				if m.AuthorID != "123456789" {
					t.Errorf("author_id: want %q, got %q", "123456789", m.AuthorID)
				}
			},
		},
		{
			name:     "missing ID returns error",
			platform: "twitter",
			topicID:  "topic-1",
			raw:      `{"text": "no id"}`,
			wantErr:  true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			topicID := tc.topicID
			if topicID == "" {
				topicID = "topic-1"
			}

			m, err := normalizeMention(tc.platform, topicID, json.RawMessage(tc.raw))
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.check != nil {
				tc.check(t, m)
			}
		})
	}
}

// --- filter tests ---

func TestPassesFilters(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		mention  kafkamodels.ListeningMention
		order    kafkamodels.ListeningWorkOrder
		wantPass bool
	}{
		{
			name:     "excluded keyword blocks mention",
			mention:  kafkamodels.ListeningMention{PostText: "This is about SPAM and ads"},
			order:    kafkamodels.ListeningWorkOrder{ExcludeKeywords: []string{"spam"}},
			wantPass: false,
		},
		{
			name:     "no excluded keyword match passes",
			mention:  kafkamodels.ListeningMention{PostText: "Clean content here"},
			order:    kafkamodels.ListeningWorkOrder{ExcludeKeywords: []string{"spam"}},
			wantPass: true,
		},
		{
			name:     "included author passes case-insensitively",
			mention:  kafkamodels.ListeningMention{PostText: "hello", AuthorID: "alice"},
			order:    kafkamodels.ListeningWorkOrder{IncludeAuthors: []string{"Alice", "Bob"}},
			wantPass: true,
		},
		{
			name:     "author not in include list is blocked",
			mention:  kafkamodels.ListeningMention{PostText: "hello", AuthorID: "charlie"},
			order:    kafkamodels.ListeningWorkOrder{IncludeAuthors: []string{"Alice", "Bob"}},
			wantPass: false,
		},
		{
			name:     "excluded author is blocked case-insensitively",
			mention:  kafkamodels.ListeningMention{PostText: "hello", AuthorID: "spammer"},
			order:    kafkamodels.ListeningWorkOrder{ExcludeAuthors: []string{"Spammer"}},
			wantPass: false,
		},
		{
			name:     "no filters passes everything",
			mention:  kafkamodels.ListeningMention{PostText: "anything", AuthorID: "anyone"},
			order:    kafkamodels.ListeningWorkOrder{},
			wantPass: true,
		},
		{
			name:     "language match passes",
			mention:  kafkamodels.ListeningMention{PostText: "hello", Language: "en"},
			order:    kafkamodels.ListeningWorkOrder{Languages: []string{"en", "fr"}},
			wantPass: true,
		},
		{
			name:     "language mismatch is blocked",
			mention:  kafkamodels.ListeningMention{PostText: "hola", Language: "es"},
			order:    kafkamodels.ListeningWorkOrder{Languages: []string{"en", "fr"}},
			wantPass: false,
		},
		{
			name:     "language match is case-insensitive",
			mention:  kafkamodels.ListeningMention{PostText: "hello", Language: "EN"},
			order:    kafkamodels.ListeningWorkOrder{Languages: []string{"en"}},
			wantPass: true,
		},
		{
			name:     "empty mention language passes language filter (permissive)",
			mention:  kafkamodels.ListeningMention{PostText: "hello", Language: ""},
			order:    kafkamodels.ListeningWorkOrder{Languages: []string{"en"}},
			wantPass: true,
		},
		{
			name:     "empty Languages in order disables language filter",
			mention:  kafkamodels.ListeningMention{PostText: "hola", Language: "es"},
			order:    kafkamodels.ListeningWorkOrder{Languages: nil},
			wantPass: true,
		},
		{
			name:     "region match passes",
			mention:  kafkamodels.ListeningMention{PostText: "hi", AuthorCountry: "US"},
			order:    kafkamodels.ListeningWorkOrder{Regions: []string{"US", "CA"}},
			wantPass: true,
		},
		{
			name:     "region mismatch is blocked",
			mention:  kafkamodels.ListeningMention{PostText: "hi", AuthorCountry: "DE"},
			order:    kafkamodels.ListeningWorkOrder{Regions: []string{"US", "CA"}},
			wantPass: false,
		},
		{
			name:     "region match is case-insensitive",
			mention:  kafkamodels.ListeningMention{PostText: "hi", AuthorCountry: "us"},
			order:    kafkamodels.ListeningWorkOrder{Regions: []string{"US"}},
			wantPass: true,
		},
		{
			name:     "empty mention country passes regions filter (permissive)",
			mention:  kafkamodels.ListeningMention{PostText: "hi", AuthorCountry: ""},
			order:    kafkamodels.ListeningWorkOrder{Regions: []string{"US"}},
			wantPass: true,
		},
		{
			name:     "empty Regions in order disables region filter",
			mention:  kafkamodels.ListeningMention{PostText: "hi", AuthorCountry: "DE"},
			order:    kafkamodels.ListeningWorkOrder{Regions: nil},
			wantPass: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got := passesFilters(&tc.mention, tc.order)
			if got != tc.wantPass {
				t.Errorf("passesFilters: want %v, got %v", tc.wantPass, got)
			}
		})
	}
}

// --- matched keywords tests ---

func TestComputeMatchedKeywords(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		text          string
		keywords      []string
		caseSensitive bool
		exactMatch    bool
		wantCount     int
	}{
		{
			name:      "matches multiple keywords",
			text:      "I love Golang and Rust programming",
			keywords:  []string{"golang", "python", "rust"},
			wantCount: 2,
		},
		{
			name:      "case-insensitive by default",
			text:      "GOLANG is great",
			keywords:  []string{"golang"},
			wantCount: 1,
		},
		{
			name:      "no matches returns empty",
			text:      "nothing relevant",
			keywords:  []string{"golang", "rust"},
			wantCount: 0,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			matched := computeMatchedKeywords(tc.text, tc.keywords, tc.caseSensitive, tc.exactMatch)
			if len(matched) != tc.wantCount {
				t.Errorf("matched count: want %d, got %d (%v)", tc.wantCount, len(matched), matched)
			}
		})
	}
}

// --- timestamp tests ---

func TestParseTimestamp(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		rfc3339   string
		unixSec   int64
		wantYear  int
		wantMonth time.Month
		wantDay   int
	}{
		{
			name:      "RFC3339 string",
			rfc3339:   "2025-01-15T10:30:00Z",
			wantYear:  2025,
			wantMonth: time.January,
			wantDay:   15,
		},
		{
			name:     "unix timestamp",
			unixSec:  1705312800,
			wantYear: 2024,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ts := parseTimestamp(tc.rfc3339, "", "", "", tc.unixSec)
			if ts.Year() != tc.wantYear {
				t.Errorf("year: want %d, got %d", tc.wantYear, ts.Year())
			}
			if tc.wantMonth != 0 && ts.Month() != tc.wantMonth {
				t.Errorf("month: want %v, got %v", tc.wantMonth, ts.Month())
			}
			if tc.wantDay != 0 && ts.Day() != tc.wantDay {
				t.Errorf("day: want %d, got %d", tc.wantDay, ts.Day())
			}
		})
	}
}

// --- dedup and pipeline tests ---

func TestHandleRawPayload_DedupSkipsDuplicate(t *testing.T) {
	t.Parallel()
	prod := &mockProducerRecorder{}
	dedup := newMockDedup()
	svc := newTestParser(prod, dedup)

	items := []map[string]interface{}{
		{"id": "1", "text": "hello", "author_id": "a1", "created_at": "2025-01-01T00:00:00"},
	}
	order := kafkamodels.ListeningWorkOrder{
		TopicID:         "topic-1",
		IncludeKeywords: []string{"hello"},
	}

	data := makeRawPayload("twitter", items, order)

	if err := svc.HandleRawPayload(context.Background(), "", nil, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	prod.mu.Lock()
	count1 := len(prod.messages)
	prod.mu.Unlock()
	if count1 != 1 {
		t.Fatalf("expected 1 emitted, got %d", count1)
	}

	if err := svc.HandleRawPayload(context.Background(), "", nil, data); err != nil {
		t.Fatalf("unexpected error on second call: %v", err)
	}
	prod.mu.Lock()
	count2 := len(prod.messages)
	prod.mu.Unlock()
	if count2 != 1 {
		t.Errorf("expected still 1 emitted (dedup), got %d", count2)
	}
}

func TestHandleRawPayload_DedupErrorFailOpen(t *testing.T) {
	t.Parallel()
	prod := &mockProducerRecorder{}
	dedup := newMockDedup()
	dedup.err = fmt.Errorf("redis unavailable")
	svc := newTestParser(prod, dedup)

	items := []map[string]interface{}{
		{"id": "1", "text": "hello", "author_id": "a1", "created_at": "2025-01-01T00:00:00"},
	}
	order := kafkamodels.ListeningWorkOrder{
		TopicID:         "topic-1",
		IncludeKeywords: []string{"hello"},
	}

	data := makeRawPayload("twitter", items, order)
	if err := svc.HandleRawPayload(context.Background(), "", nil, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prod.mu.Lock()
	defer prod.mu.Unlock()
	if len(prod.messages) != 1 {
		t.Fatalf("expected mention to be emitted in fail-open mode, got %d", len(prod.messages))
	}
}

func TestHandleRawPayload_DoesNotDropAlreadyFetchedPayloadWhenTopicLimitReached(t *testing.T) {
	t.Parallel()
	prod := &mockProducerRecorder{}
	dedup := newMockDedup()
	log, _ := logger.NewTestLogger()
	svc := NewParserService(prod, dedup, &mockTopicStatusChecker{
		topic: &mongomodels.ListeningTopic{MentionsLimitReached: true},
	}, log, 48)

	order := kafkamodels.ListeningWorkOrder{TopicID: "topic-1"}
	items := []map[string]interface{}{
		{
			"id":          "12345",
			"text":        "Hello world from Twitter",
			"author_id":   "user1",
			"author_name": "John",
			"url":         "https://twitter.com/status/12345",
			"created_at":  "2025-01-15T10:00:00",
		},
	}

	err := svc.HandleRawPayload(context.Background(), "", nil, makeRawPayload("twitter", items, order))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prod.mu.Lock()
	defer prod.mu.Unlock()
	if len(prod.messages) != 1 {
		t.Fatalf("expected fetched payload to continue through parsing, got %d messages", len(prod.messages))
	}
}

func TestHandleRawPayload_FiltersAndEmits(t *testing.T) {
	t.Parallel()
	prod := &mockProducerRecorder{}
	dedup := newMockDedup()
	svc := newTestParser(prod, dedup)

	items := []map[string]interface{}{
		{"id": "1", "text": "I love golang", "author_id": "alice", "created_at": "2025-01-01T00:00:00"},
		{"id": "2", "text": "spam spam spam", "author_id": "bob", "created_at": "2025-01-01T00:00:00"},
		{"id": "3", "text": "rust is cool", "author_id": "charlie", "created_at": "2025-01-01T00:00:00"},
	}
	order := kafkamodels.ListeningWorkOrder{
		TopicID:         "topic-1",
		IncludeKeywords: []string{"golang", "rust"},
		ExcludeKeywords: []string{"spam"},
	}

	data := makeRawPayload("twitter", items, order)
	if err := svc.HandleRawPayload(context.Background(), "", nil, data); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	prod.mu.Lock()
	defer prod.mu.Unlock()
	if len(prod.messages) != 2 {
		t.Errorf("expected 2 emitted (1 filtered), got %d", len(prod.messages))
	}

	for _, msg := range prod.messages {
		var mention kafkamodels.ListeningMention
		if err := json.Unmarshal(msg.Value, &mention); err != nil {
			t.Fatalf("unmarshal mention: %v", err)
		}
		if mention.TopicID != "topic-1" {
			t.Errorf("topic_id: want %q, got %q", "topic-1", mention.TopicID)
		}
		if msg.Topic != kafkamodels.TopicListeningParsed {
			t.Errorf("topic: want %q, got %q", kafkamodels.TopicListeningParsed, msg.Topic)
		}
	}
}

func TestHandleRawPayload_InvalidJSON(t *testing.T) {
	t.Parallel()
	svc := newTestParser(&mockProducerRecorder{}, newMockDedup())
	err := svc.HandleRawPayload(context.Background(), "", nil, []byte("{bad"))
	if err == nil {
		t.Fatal("expected unmarshal error")
	}
}

// --- helper tests ---

func TestTruncateText(t *testing.T) {
	t.Parallel()
	long := string(make([]byte, 20000))
	result := truncateText(long, maxTextBytes)
	if len(result) != maxTextBytes {
		t.Errorf("expected %d bytes, got %d", maxTextBytes, len(result))
	}
}

func TestTruncateText_PreservesUTF8(t *testing.T) {
	t.Parallel()

	input := "hello🙂world"
	result := truncateText(input, len("hello🙂"))
	if result != "hello🙂" {
		t.Fatalf("unexpected truncation: %q", result)
	}
}

func TestContainsTerm_ExactMatchChecksAllOccurrences(t *testing.T) {
	t.Parallel()

	if !containsTerm("foogolang golang", "golang", true, false) {
		t.Fatal("expected later exact-match occurrence to be detected")
	}
}

func TestCoalesce(t *testing.T) {
	t.Parallel()
	if coalesce("", "", "third") != "third" {
		t.Error("expected 'third'")
	}
	if coalesce("first", "second") != "first" {
		t.Error("expected 'first'")
	}
	if coalesce("", "") != "" {
		t.Error("expected empty")
	}
}
