package kafka

import (
	"encoding/json"
	"time"
)

// TwitterAccountWorkOrder represents a single Twitter account to process.
// Used both standalone and within batch messages.
type TwitterAccountWorkOrder struct {
	ID               string `json:"id"`                 // MongoDB _id (hex)
	WorkspaceID      string `json:"workspace_id"`       // Workspace identifier
	TwitterID        string `json:"twitter_id"`         // Twitter user ID
	OAuthToken       string `json:"oauth_token"`        // OAuth 1.0a token (per-user)
	OAuthTokenSecret string `json:"oauth_token_secret"` // OAuth 1.0a token secret (per-user)
	PostCount        int    `json:"post_count"`         // Number of posts to fetch
	NTweets          int    `json:"n_tweets,omitempty"` // Requested tweet count from sync API
	APIKey           string `json:"api_key"`            // Developer app key
	APISecret        string `json:"api_secret"`         // Developer app secret
	AppName          string `json:"app_name"`           // Twitter app name for job logs
	AppID            string `json:"app_id"`             // Developer app ObjectID (hex)
	ExecutedBy       string `json:"executed_by"`        // Job initiator (Python parity: internal)
	SyncType         string `json:"sync_type"`          // "incremental" | "full_sync"
}

// TwitterBatchWorkOrder represents a batch of Twitter accounts to process.
// The scheduler produces batch messages to reduce Kafka overhead.
// The fetcher unpacks batches and distributes accounts to worker pools.
type TwitterBatchWorkOrder struct {
	BatchID   string                    `json:"batch_id"`   // Unique batch identifier (UUID)
	SyncType  string                    `json:"sync_type"`  // "incremental" | "full_sync"
	Accounts  []TwitterAccountWorkOrder `json:"accounts"`   // List of accounts in this batch (max 200)
	CreatedAt time.Time                 `json:"created_at"` // Batch creation timestamp
}

// RawTwitterPost is the exact JSON blob returned by Twitter APIs. We wrap it in a
// struct to keep type-safety when sending through Kafka.
type RawTwitterPost struct {
	WorkspaceID string          `json:"workspace_id"`
	TwitterID   string          `json:"twitter_id"`
	Data        json.RawMessage `json:"data"`
}

// ParsedTwitterPost is the normalised structure produced by the parser service
// and consumed downstream by the ClickHouse sink.
// All fields from the Twitter API v2 users/:id/tweets endpoint are included.
type ParsedTwitterPost struct {
	TwitterID           string   `json:"twitter_id"`
	Name                string   `json:"name"`
	Username            string   `json:"username"`
	ProfileImageURL     string   `json:"profile_image_url,omitempty"`
	FollowersCount      int64    `json:"followers_count"`
	FollowingCount      int64    `json:"following_count"`
	TweetCount          int64    `json:"tweet_count"`
	ListedCount         int64    `json:"listed_count"`
	TweetID             string   `json:"tweet_id"`
	EditHistoryTweetIDs []string `json:"edit_history_tweet_ids,omitempty"`
	AuthorID            string   `json:"author_id"`
	AuthorUsername      string   `json:"author_username,omitempty"`
	IDCreatedAt         string   `json:"id_created_at,omitempty"`     // Twitter user's account creation date (RFC3339)
	AuthorIDCreated     string   `json:"author_id_created,omitempty"` // Author account creation date
	TweetedAt           string   `json:"tweeted_at"`                  // Tweet creation time (RFC3339)
	Hashtags            []string `json:"hashtags,omitempty"`
	Permalink           string   `json:"permalink"`
	TweetType           string   `json:"tweet_type"`
	URLs                []string `json:"urls,omitempty"`
	MediaURL            []string `json:"media_url,omitempty"`
	UsernameMentioned   []string `json:"username_mentioned,omitempty"`
	UseridMentioned     []string `json:"userid_mentioned,omitempty"`
	Lang                string   `json:"lang,omitempty"`
	TweetText           string   `json:"tweet_text"`
	ImpressionCount     int64    `json:"impression_count"`
	RetweetCount        int64    `json:"retweet_count"`
	ReplyCount          int64    `json:"reply_count"`
	LikeCount           int64    `json:"like_count"`
	BookmarkCount       int64    `json:"bookmark_count"`
	QuoteCount          int64    `json:"quote_count"`
	TotalEngagement     int64    `json:"total_engagement"`
}

// ParsedTwitterInsights represents aggregated account-level metrics.
type ParsedTwitterInsights struct {
	TwitterID          string `json:"twitter_id"`
	RecordID           string `json:"record_id"`
	Name               string `json:"name"`
	Username           string `json:"username"`
	ProfileImageURL    string `json:"profile_image_url,omitempty"`
	Description        string `json:"description,omitempty"`
	Verified           bool   `json:"verified"`
	AccountCreatedDate string `json:"account_created_date,omitempty"` // RFC3339
	FollowersCount     int64  `json:"followers_count"`
	FollowingCount     int64  `json:"following_count"`
	TweetCount         int64  `json:"tweet_count"`
	ListedCount        int64  `json:"listed_count"`
	LikeCount          int64  `json:"like_count"`
	InsertedAt         int64  `json:"inserted_at"` // Unix timestamp
}
