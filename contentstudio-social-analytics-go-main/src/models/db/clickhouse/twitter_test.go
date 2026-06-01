package clickhouse

import (
	"testing"
	"time"
)

func TestTwitterPosts_Struct(t *testing.T) {
	now := time.Now()
	post := &TwitterPosts{
		TwitterID:           "12345",
		Name:                "Test User",
		Username:            "testuser",
		ProfileImageURL:     "https://example.com/img.jpg",
		FollowersCount:      10000,
		FollowingCount:      500,
		TweetCount:          5000,
		ListedCount:         100,
		TweetID:             "tweet_123",
		EditHistoryTweetIDs: []string{"tweet_123"},
		AuthorID:            "12345",
		AuthorUsername:      "testuser",
		IDCreatedAt:         now,
		AuthorIDCreated:     now,
		TweetedAt:           now,
		SavingTime:          now,
		Hashtags:            []string{"golang", "twitter"},
		Permalink:           "https://twitter.com/testuser/status/tweet_123",
		TweetType:           "tweet",
		URLs:                []string{"https://example.com"},
		MediaURL:            []string{"https://example.com/media.jpg"},
		UsernameMentioned:   []string{"otheruser"},
		UseridMentioned:     []string{"67890"},
		Lang:                "en",
		TweetText:           "Hello World #golang #twitter",
		ImpressionCount:     1000,
		RetweetCount:        10,
		ReplyCount:          5,
		LikeCount:           100,
		BookmarkCount:       3,
		QuoteCount:          2,
		TotalEngagement:     120,
		DayOfWeek:           1,
		HourOfDay:           14,
	}

	if post.TwitterID != "12345" {
		t.Errorf("TwitterID = %q, want %q", post.TwitterID, "12345")
	}
	if post.TweetID != "tweet_123" {
		t.Errorf("TweetID = %q, want %q", post.TweetID, "tweet_123")
	}
	if post.FollowersCount != 10000 {
		t.Errorf("FollowersCount = %d, want 10000", post.FollowersCount)
	}
	if post.TotalEngagement != 120 {
		t.Errorf("TotalEngagement = %d, want 120", post.TotalEngagement)
	}
	if len(post.Hashtags) != 2 {
		t.Errorf("Hashtags length = %d, want 2", len(post.Hashtags))
	}
}

func TestTwitterInsights_Struct(t *testing.T) {
	now := time.Now()
	insight := &TwitterInsights{
		TwitterID:          "12345",
		RecordID:           "record_abc123",
		Name:               "Test User",
		Username:           "testuser",
		ProfileImageURL:    "https://example.com/img.jpg",
		Description:        "Test bio",
		Verified:           "true",
		AccountCreatedDate: now,
		FollowersCount:     10000,
		FollowingCount:     500,
		TweetCount:         5000,
		ListedCount:        100,
		LikeCount:          2000,
		DayOfWeek:          1,
		SavingTime:         now,
	}

	if insight.TwitterID != "12345" {
		t.Errorf("TwitterID = %q, want %q", insight.TwitterID, "12345")
	}
	if insight.RecordID != "record_abc123" {
		t.Errorf("RecordID = %q, want %q", insight.RecordID, "record_abc123")
	}
	if insight.FollowersCount != 10000 {
		t.Errorf("FollowersCount = %d, want 10000", insight.FollowersCount)
	}
	if insight.Verified != "true" {
		t.Error("Verified should be true string")
	}
	if insight.Description != "Test bio" {
		t.Errorf("Description = %q, want %q", insight.Description, "Test bio")
	}
}
