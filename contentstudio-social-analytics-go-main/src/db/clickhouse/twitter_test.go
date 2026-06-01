package clickhouse

import (
	"context"
	"errors"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// Test_BulkInsertTwitterPosts_EmptySlice tests that empty slice returns no error
func Test_BulkInsertTwitterPosts_EmptySlice(t *testing.T) {
	conn := &mockConn{}
	client := newTestClient(conn)
	err := client.BulkInsertTwitterPosts(context.Background(), []*clickhousemodels.TwitterPosts{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

// Test_BulkInsertTwitterPosts_Table tests various scenarios for BulkInsertTwitterPosts
func Test_BulkInsertTwitterPosts_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		posts     []*clickhousemodels.TwitterPosts
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty posts",
			posts:     []*clickhousemodels.TwitterPosts{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single post success",
			posts: []*clickhousemodels.TwitterPosts{
				{
					TwitterID:      "twitter_123",
					TweetID:        "tweet_456",
					Username:       "testuser",
					Name:           "Test User",
					FollowersCount: 1000,
					TweetedAt:      now,
					SavingTime:     now,
					TweetText:      "Hello Twitter",
					LikeCount:      100,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "multiple posts success",
			posts: []*clickhousemodels.TwitterPosts{
				{
					TwitterID:  "twitter_123",
					TweetID:    "tweet_456",
					Username:   "testuser",
					TweetedAt:  now,
					SavingTime: now,
				},
				{
					TwitterID:  "twitter_124",
					TweetID:    "tweet_457",
					Username:   "testuser2",
					TweetedAt:  now,
					SavingTime: now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			posts: []*clickhousemodels.TwitterPosts{
				{TwitterID: "twitter_123", TweetID: "tweet_456", TweetedAt: now, SavingTime: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			posts: []*clickhousemodels.TwitterPosts{
				{TwitterID: "twitter_123", TweetID: "tweet_456", TweetedAt: now, SavingTime: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			posts: []*clickhousemodels.TwitterPosts{
				{TwitterID: "twitter_123", TweetID: "tweet_456", TweetedAt: now, SavingTime: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertTwitterPosts(context.Background(), tc.posts)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// Test_BulkInsertTwitterInsights_EmptySlice tests that empty slice returns no error
func Test_BulkInsertTwitterInsights_EmptySlice(t *testing.T) {
	conn := &mockConn{}
	client := newTestClient(conn)
	err := client.BulkInsertTwitterInsights(context.Background(), []*clickhousemodels.TwitterInsights{})
	if err != nil {
		t.Fatalf("expected nil error for empty slice, got %v", err)
	}
}

// Test_BulkInsertTwitterInsights_Table tests various scenarios for BulkInsertTwitterInsights
func Test_BulkInsertTwitterInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.TwitterInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.TwitterInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.TwitterInsights{
				{
					TwitterID:      "twitter_123",
					RecordID:       "record_456",
					Username:       "testuser",
					Name:           "Test User",
					FollowersCount: 1000,
					Verified:       "true",
					SavingTime:     now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "multiple insights success",
			insights: []*clickhousemodels.TwitterInsights{
				{
					TwitterID:      "twitter_123",
					RecordID:       "record_456",
					Username:       "testuser",
					FollowersCount: 1000,
					Verified:       "true",
					SavingTime:     now,
				},
				{
					TwitterID:      "twitter_124",
					RecordID:       "record_457",
					Username:       "testuser2",
					FollowersCount: 5000,
					Verified:       "false",
					SavingTime:     now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.TwitterInsights{
				{TwitterID: "twitter_123", RecordID: "record_456", SavingTime: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.TwitterInsights{
				{TwitterID: "twitter_123", RecordID: "record_456", SavingTime: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.TwitterInsights{
				{TwitterID: "twitter_123", RecordID: "record_456", SavingTime: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertTwitterInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

// Test_BulkInsertTwitterPosts_WithAllFields tests bulk insert with all fields populated
func Test_BulkInsertTwitterPosts_WithAllFields(t *testing.T) {
	now := time.Now()
	posts := []*clickhousemodels.TwitterPosts{
		{
			TwitterID:           "twitter_123",
			Name:                "John Doe",
			Username:            "johndoe",
			ProfileImageURL:     "https://example.com/profile.jpg",
			FollowersCount:      5000,
			FollowingCount:      500,
			TweetCount:          250,
			ListedCount:         10,
			TweetID:             "tweet_001",
			EditHistoryTweetIDs: []string{"tweet_001", "tweet_001_v1"},
			AuthorID:            "author_123",
			AuthorUsername:      "author_johndoe",
			IDCreatedAt:         now.AddDate(-5, 0, 0),
			AuthorIDCreated:     now.AddDate(-5, 0, 0),
			TweetedAt:           now,
			SavingTime:          now,
			Hashtags:            []string{"golang", "testing"},
			Permalink:           "https://twitter.com/johndoe/status/tweet_001",
			TweetType:           "original",
			URLs:                []string{"https://example.com"},
			MediaURL:            []string{"https://example.com/image.jpg"},
			UsernameMentioned:   []string{"user1", "user2"},
			UseridMentioned:     []string{"id1", "id2"},
			Lang:                "en",
			TweetText:           "Hello, this is a test tweet!",
			ImpressionCount:     10000,
			RetweetCount:        100,
			ReplyCount:          50,
			LikeCount:           500,
			BookmarkCount:       75,
			QuoteCount:          10,
			TotalEngagement:     735,
			DayOfWeek:           int64(now.Weekday()),
			HourOfDay:           int64(now.Hour()),
		},
	}

	conn := &mockConn{}
	client := newTestClient(conn)
	err := client.BulkInsertTwitterPosts(context.Background(), posts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test_BulkInsertTwitterInsights_WithAllFields tests bulk insert with all fields populated
func Test_BulkInsertTwitterInsights_WithAllFields(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.TwitterInsights{
		{
			TwitterID:          "twitter_123",
			RecordID:           "record_001",
			Name:               "John Doe",
			Username:           "johndoe",
			ProfileImageURL:    "https://example.com/profile.jpg",
			Description:        "Software developer and tech enthusiast",
			Verified:           "true",
			AccountCreatedDate: now.AddDate(-5, 0, 0),
			FollowersCount:     50000,
			FollowingCount:     1000,
			TweetCount:         5000,
			ListedCount:        150,
			LikeCount:          100000,
			DayOfWeek:          int64(now.Weekday()),
			SavingTime:         now,
		},
	}

	conn := &mockConn{}
	client := newTestClient(conn)
	err := client.BulkInsertTwitterInsights(context.Background(), insights)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test_BulkInsertTwitterPosts_MultipleItems tests bulk insert with multiple items
func Test_BulkInsertTwitterPosts_MultipleItems(t *testing.T) {
	now := time.Now()
	posts := []*clickhousemodels.TwitterPosts{
		{
			TwitterID:      "twitter_001",
			TweetID:        "tweet_001",
			Username:       "user1",
			TweetedAt:      now,
			SavingTime:     now,
			FollowersCount: 1000,
		},
		{
			TwitterID:      "twitter_002",
			TweetID:        "tweet_002",
			Username:       "user2",
			TweetedAt:      now,
			SavingTime:     now,
			FollowersCount: 2000,
		},
		{
			TwitterID:      "twitter_003",
			TweetID:        "tweet_003",
			Username:       "user3",
			TweetedAt:      now,
			SavingTime:     now,
			FollowersCount: 3000,
		},
	}

	conn := &mockConn{}
	client := newTestClient(conn)
	err := client.BulkInsertTwitterPosts(context.Background(), posts)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test_BulkInsertTwitterInsights_MultipleItems tests bulk insert with multiple insight items
func Test_BulkInsertTwitterInsights_MultipleItems(t *testing.T) {
	now := time.Now()
	insights := []*clickhousemodels.TwitterInsights{
		{
			TwitterID:      "twitter_001",
			RecordID:       "record_001",
			FollowersCount: 1000,
			SavingTime:     now,
		},
		{
			TwitterID:      "twitter_002",
			RecordID:       "record_002",
			FollowersCount: 2000,
			SavingTime:     now,
		},
		{
			TwitterID:      "twitter_003",
			RecordID:       "record_003",
			FollowersCount: 3000,
			SavingTime:     now,
		},
	}

	conn := &mockConn{}
	client := newTestClient(conn)
	err := client.BulkInsertTwitterInsights(context.Background(), insights)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// Test_BulkInsertTwitterPosts_ContextCancellation tests cancellation
func Test_BulkInsertTwitterPosts_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel the context

	conn := &mockConn{}
	client := newTestClient(conn)

	posts := []*clickhousemodels.TwitterPosts{
		{TwitterID: "twitter_123", TweetID: "tweet_456"},
	}

	// This might error depending on implementation, but shouldn't panic
	_ = client.BulkInsertTwitterPosts(ctx, posts)
}

// Test_BulkInsertTwitterInsights_ContextCancellation tests cancellation
func Test_BulkInsertTwitterInsights_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Immediately cancel the context

	conn := &mockConn{}
	client := newTestClient(conn)

	insights := []*clickhousemodels.TwitterInsights{
		{TwitterID: "twitter_123", RecordID: "record_456"},
	}

	// This might error depending on implementation, but shouldn't panic
	_ = client.BulkInsertTwitterInsights(ctx, insights)
}
