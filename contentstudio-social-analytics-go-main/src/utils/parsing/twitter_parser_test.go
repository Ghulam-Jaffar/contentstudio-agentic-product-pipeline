package parsing

import (
	"strings"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
)

func TestNewTwitterParser(t *testing.T) {
	p := NewTwitterParser()
	if p == nil {
		t.Fatal("NewTwitterParser returned nil")
	}
}

func TestParseTweet_Basic(t *testing.T) {
	p := NewTwitterParser()
	user := &social.TwitterUser{
		ID:              "12345",
		Name:            "Test User",
		Username:        "testuser",
		ProfileImageURL: "https://pbs.twimg.com/img.jpg",
		CreatedAt:       "2020-01-15T10:30:00.000Z",
	}
	user.PublicMetrics.FollowersCount = 5000
	user.PublicMetrics.FollowingCount = 200
	user.PublicMetrics.TweetCount = 1500
	user.PublicMetrics.ListedCount = 50

	tweet := social.TwitterTweet{
		ID:        "tweet_001",
		Text:      "Hello #golang world @mention1",
		AuthorID:  "12345",
		CreatedAt: "2024-01-15T12:00:00.000Z",
		Lang:      "en",
		PublicMetrics: social.TwitterTweetMetrics{
			ImpressionCount: 1000,
			RetweetCount:    10,
			ReplyCount:      5,
			LikeCount:       50,
			BookmarkCount:   3,
			QuoteCount:      2,
		},
		Entities: &social.TwitterEntities{
			Hashtags: []struct {
				Tag string `json:"tag"`
			}{
				{Tag: "golang"},
			},
			Mentions: []struct {
				Username string `json:"username"`
				ID       string `json:"id"`
			}{
				{Username: "mention1", ID: "99999"},
			},
			URLs: []struct {
				URL         string `json:"url"`
				ExpandedURL string `json:"expanded_url"`
				DisplayURL  string `json:"display_url"`
			}{
				{URL: "https://t.co/abc", ExpandedURL: "https://example.com"},
			},
		},
		EditHistoryTweetIDs: []string{"tweet_001"},
	}

	post := p.ParseTweet(tweet, user, nil)
	if post == nil {
		t.Fatal("ParseTweet returned nil")
	}

	if post.TweetID != "tweet_001" {
		t.Errorf("TweetID = %q, want %q", post.TweetID, "tweet_001")
	}
	if post.TwitterID != "12345" {
		t.Errorf("TwitterID = %q, want %q", post.TwitterID, "12345")
	}
	if post.Name != "Test User" {
		t.Errorf("Name = %q, want %q", post.Name, "Test User")
	}
	if post.Username != "testuser" {
		t.Errorf("Username = %q, want %q", post.Username, "testuser")
	}
	if post.FollowersCount != 5000 {
		t.Errorf("FollowersCount = %d, want 5000", post.FollowersCount)
	}
	if post.TweetText != "Hello #golang world @mention1" {
		t.Errorf("TweetText = %q", post.TweetText)
	}
	if post.Lang != "en" {
		t.Errorf("Lang = %q, want %q", post.Lang, "en")
	}
	if post.ImpressionCount != 1000 {
		t.Errorf("ImpressionCount = %d, want 1000", post.ImpressionCount)
	}
	// TotalEngagement = retweet(10) + reply(5) + like(50) + bookmark(3) + quote(2) = 70
	if post.TotalEngagement != 70 {
		t.Errorf("TotalEngagement = %d, want 70", post.TotalEngagement)
	}
	if post.TweetType != "tweet" {
		t.Errorf("TweetType = %q, want %q", post.TweetType, "tweet")
	}
	if len(post.Hashtags) != 1 || post.Hashtags[0] != "golang" {
		t.Errorf("Hashtags = %v, want [golang]", post.Hashtags)
	}
	if len(post.UsernameMentioned) != 1 || post.UsernameMentioned[0] != "mention1" {
		t.Errorf("UsernameMentioned = %v", post.UsernameMentioned)
	}
	if len(post.UseridMentioned) != 1 || post.UseridMentioned[0] != "99999" {
		t.Errorf("UseridMentioned = %v", post.UseridMentioned)
	}
	if len(post.URLs) != 1 || post.URLs[0] != "https://example.com" {
		t.Errorf("URLs = %v", post.URLs)
	}
	if !strings.Contains(post.Permalink, "testuser") || !strings.Contains(post.Permalink, "tweet_001") {
		t.Errorf("Permalink = %q", post.Permalink)
	}
}

func TestParseTweet_EmptyID(t *testing.T) {
	p := NewTwitterParser()
	tweet := social.TwitterTweet{ID: ""}
	post := p.ParseTweet(tweet, nil, nil)
	if post != nil {
		t.Error("Expected nil for empty tweet ID")
	}
}

func TestParseTweet_NilUserInfo(t *testing.T) {
	p := NewTwitterParser()
	tweet := social.TwitterTweet{
		ID:        "tweet_002",
		AuthorID:  "12345",
		CreatedAt: "2024-01-15T12:00:00.000Z",
		PublicMetrics: social.TwitterTweetMetrics{
			LikeCount: 10,
		},
	}

	post := p.ParseTweet(tweet, nil, nil)
	if post == nil {
		t.Fatal("ParseTweet returned nil")
	}
	if post.TwitterID != "" {
		t.Errorf("TwitterID should be empty, got %q", post.TwitterID)
	}
	if post.TotalEngagement != 10 {
		t.Errorf("TotalEngagement = %d, want 10", post.TotalEngagement)
	}
}

func TestParseTweet_TweetTypes(t *testing.T) {
	p := NewTwitterParser()

	tests := []struct {
		name string
		refs []social.TwitterReferenceTweet
		want string
	}{
		{"original", nil, "tweet"},
		{"retweet", []social.TwitterReferenceTweet{{Type: "retweeted", ID: "1"}}, "retweet"},
		{"quote", []social.TwitterReferenceTweet{{Type: "quoted", ID: "2"}}, "quote"},
		{"reply", []social.TwitterReferenceTweet{{Type: "replied_to", ID: "3"}}, "reply"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tweet := social.TwitterTweet{
				ID:               "t_" + tc.name,
				ReferencedTweets: tc.refs,
			}
			post := p.ParseTweet(tweet, nil, nil)
			if post.TweetType != tc.want {
				t.Errorf("TweetType = %q, want %q", post.TweetType, tc.want)
			}
		})
	}
}

func TestParseTweet_MediaExpansion(t *testing.T) {
	p := NewTwitterParser()

	includes := &social.TwitterIncludes{
		Media: []social.TwitterMedia{
			{MediaKey: "mk_1", Type: "photo", URL: "https://pbs.twimg.com/photo1.jpg"},
			{MediaKey: "mk_2", Type: "video", PreviewImageURL: "https://pbs.twimg.com/preview2.jpg"},
		},
	}

	tweet := social.TwitterTweet{
		ID: "tweet_media",
		Attachments: &social.TwitterAttachments{
			MediaKeys: []string{"mk_1", "mk_2"},
		},
	}

	post := p.ParseTweet(tweet, nil, includes)
	if post == nil {
		t.Fatal("ParseTweet returned nil")
	}
	if len(post.MediaURL) != 2 {
		t.Fatalf("MediaURL length = %d, want 2", len(post.MediaURL))
	}
	if post.MediaURL[0] != "https://pbs.twimg.com/photo1.jpg" {
		t.Errorf("MediaURL[0] = %q", post.MediaURL[0])
	}
	if post.MediaURL[1] != "https://pbs.twimg.com/preview2.jpg" {
		t.Errorf("MediaURL[1] = %q", post.MediaURL[1])
	}
}

func TestGenerateInsights(t *testing.T) {
	p := NewTwitterParser()

	user := &social.TwitterUser{
		ID:              "12345",
		Name:            "Test User",
		Username:        "testuser",
		Description:     "Test bio",
		ProfileImageURL: "https://example.com/img.jpg",
		Verified:        true,
		CreatedAt:       "2020-01-15T10:30:00.000Z",
	}
	user.PublicMetrics.FollowersCount = 10000
	user.PublicMetrics.FollowingCount = 500
	user.PublicMetrics.TweetCount = 5000
	user.PublicMetrics.ListedCount = 100
	user.PublicMetrics.LikeCount = 2000

	insights := p.GenerateInsights(user)
	if insights == nil {
		t.Fatal("GenerateInsights returned nil")
	}

	if insights.TwitterID != "12345" {
		t.Errorf("TwitterID = %q, want %q", insights.TwitterID, "12345")
	}
	if insights.RecordID == "" {
		t.Error("RecordID should not be empty")
	}
	if insights.Name != "Test User" {
		t.Errorf("Name = %q, want %q", insights.Name, "Test User")
	}
	if insights.FollowersCount != 10000 {
		t.Errorf("FollowersCount = %d, want 10000", insights.FollowersCount)
	}
	if insights.LikeCount != 2000 {
		t.Errorf("LikeCount = %d, want 2000", insights.LikeCount)
	}
	if !insights.Verified {
		t.Error("Verified should be true")
	}
	if insights.InsertedAt == 0 {
		t.Error("InsertedAt should not be 0")
	}
}

func TestGenerateInsights_NilUser(t *testing.T) {
	p := NewTwitterParser()
	insights := p.GenerateInsights(nil)
	if insights != nil {
		t.Error("Expected nil for nil user")
	}
}

func TestDetermineTweetType(t *testing.T) {
	tests := []struct {
		name string
		refs []social.TwitterReferenceTweet
		want string
	}{
		{"empty", nil, "tweet"},
		{"retweet", []social.TwitterReferenceTweet{{Type: "retweeted"}}, "retweet"},
		{"quoted", []social.TwitterReferenceTweet{{Type: "quoted"}}, "quote"},
		{"replied", []social.TwitterReferenceTweet{{Type: "replied_to"}}, "reply"},
		{"multiple_retweet_first", []social.TwitterReferenceTweet{{Type: "retweeted"}, {Type: "quoted"}}, "retweet"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := determineTweetType(tc.refs)
			if got != tc.want {
				t.Errorf("determineTweetType() = %q, want %q", got, tc.want)
			}
		})
	}
}

func TestBuildMediaKeyMap(t *testing.T) {
	media := []social.TwitterMedia{
		{MediaKey: "mk1", Type: "photo", URL: "url1"},
		{MediaKey: "mk2", Type: "video", URL: "url2"},
	}
	m := buildMediaKeyMap(media)
	if len(m) != 2 {
		t.Fatalf("map length = %d, want 2", len(m))
	}
	if m["mk1"].URL != "url1" {
		t.Errorf("mk1 URL = %q", m["mk1"].URL)
	}
	if m["mk2"].Type != "video" {
		t.Errorf("mk2 Type = %q", m["mk2"].Type)
	}
}
