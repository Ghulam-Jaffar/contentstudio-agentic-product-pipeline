// Package twitter provides the ClickHouse repository layer for Twitter/X analytics.
// This file contains query result types returned by repository methods.
package twitter

// SummaryResult holds the aggregated Twitter account summary for a date range.
type SummaryResult struct {
	TwitterID       string
	Name            string
	ProfileImageURL string
	FollowersCount  int64
	FollowingCount  int64
	TweetCount      int64
	ListedCount     int64
	ImpressionCount int64
	TotalEngagement int64
	ReplyCount      int64
	RetweetCount    int64
	BookmarkCount   int64
	LikeCount       int64
	QuoteCount      int64
	PostsTweetCount int64
}

// EngagementImpressionResult holds time-series engagement and impression data.
type EngagementImpressionResult struct {
	TwitterID       string
	TweetCount      []int64
	ImpressionCount []int64
	TotalEngagement []int64
	TweetedAtDate   []string
	RetweetCount    []int64
	ReplyCount      []int64
	LikeCount       []int64
	BookmarkCount   []int64
	QuoteCount      []int64
}

// FollowersTrendResult holds time-series follower trend data.
type FollowersTrendResult struct {
	PlatformID          string
	Name                string
	Username            string
	FollowerCount       []int64
	FollowerCountDaily  []int64
	FollowingCount      []int64
	FollowingCountDaily []int64
	Buckets             []string
}

// TweetRow holds a single tweet's data for the posts list and top/least posts queries.
type TweetRow struct {
	ID              string
	TweetedAt       string
	TweetText       string
	TweetType       string
	Permalink       string
	MediaURL        []string
	ListedCount     int32
	RetweetCount    int32
	LikeCount       int32
	ReplyCount      int32
	QuoteCount      int32
	BookmarkCount   int32
	ImpressionCount int32
	TotalEngagement int32
}
