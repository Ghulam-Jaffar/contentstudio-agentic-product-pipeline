// Package overview provides the ClickHouse repository layer for cross-platform Overview analytics.
// It contains query result types and repository methods executing ClickHouse SQL queries
// migrated from the PHP Laravel OverviewV2Builder + OverviewV2Controller (contentstudio-backend).
//
// Tables queried: facebook_posts, facebook_insights, instagram_posts, instagram_insights,
// linkedin_posts, linkedin_insights, tiktok_posts, tiktok_insights,
// pinterest_pins, pinterest_pin_insights, pinterest_boards,
// youtube_videos, youtube_channels, mv_social_daily_metrics
package overview

import "time"

// PlatformDataRow holds a single platform's aggregated analytics data.
// Returned by GetPlatformData (grouped by platform_type).
type PlatformDataRow struct {
	Followers    int32
	TotalPosts   int32
	Engagement   int32
	Impressions  int32
	Reach        int32
	Reactions    int32
	Comments     int32
	Shares       int32
	PlatformType string
}

// AccountDataRow holds a single account's aggregated analytics data.
// Returned by GetAccountData (grouped by account_id and platform_type).
type AccountDataRow struct {
	Followers    int32
	TotalPosts   int32
	Engagement   int32
	Impressions  int32
	Reach        int32
	Reactions    int32
	Comments     int32
	Shares       int32
	PlatformType string
	AccountID    string
}

// TopPerformingGraphResult holds per-platform daily time-series from mv_social_daily_metrics.
// Returned by GetTopPerformingGraph as a single aggregated row with array fields.
type TopPerformingGraphResult struct {
	Buckets                  []time.Time `json:"buckets"`
	FacebookPostCount        []float64   `json:"facebook_post_count"`
	InstagramPostCount       []float64   `json:"instagram_post_count"`
	LinkedInPostCount        []float64   `json:"linkedin_post_count"`
	TiktokPostCount          []float64   `json:"tiktok_post_count"`
	YouTubePostCount         []float64   `json:"youtube_post_count"`
	PinterestPostCount       []float64   `json:"pinterest_post_count"`
	FacebookEngagementCount  []float64   `json:"facebook_engagement_count"`
	InstagramEngagementCount []float64   `json:"instagram_engagement_count"`
	LinkedInEngagementCount  []float64   `json:"linkedin_engagement_count"`
	TiktokEngagementCount    []float64   `json:"tiktok_engagement_count"`
	YouTubeEngagementCount   []float64   `json:"youtube_engagement_count"`
	PinterestEngagementCount []float64   `json:"pinterest_engagement_count"`
	FacebookImpressionCount  []float64   `json:"facebook_impression_count"`
	InstagramImpressionCount []float64   `json:"instagram_impression_count"`
	LinkedInImpressionCount  []float64   `json:"linkedin_impression_count"`
	TiktokImpressionCount    []float64   `json:"tiktok_impression_count"`
	YouTubeImpressionCount   []float64   `json:"youtube_impression_count"`
	PinterestImpressionCount []float64   `json:"pinterest_impression_count"`
	FacebookReachCount       []float64   `json:"facebook_reach_count"`
	InstagramReachCount      []float64   `json:"instagram_reach_count"`
	LinkedInReachCount       []float64   `json:"linkedin_reach_count"`
	TiktokReachCount         []float64   `json:"tiktok_reach_count"`
	YouTubeReachCount        []float64   `json:"youtube_reach_count"`
	PinterestReachCount      []float64   `json:"pinterest_reach_count"`
}

// AccountDataDetailedRow holds current/previous period data with pct changes for a single account.
// Returned by GetAccountDataDetailed.
type AccountDataDetailedRow struct {
	PlatformType         string
	AccountID            string
	AccountName          string
	CurrentFollowers     int32
	OldFollowers         int32
	CurrentPosts         int32
	OldPosts             int32
	CurrentEngagement    int32
	OldEngagement        int32
	CurrentImpressions   int32
	OldImpressions       int32
	CurrentReach         int32
	OldReach             int32
	FollowersChangePct   float64
	PostsChangePct       float64
	EngagementChangePct  float64
	ImpressionsChangePct float64
	ReachChangePct       float64
}

// AccountDataGraphsRow holds time-series per-account arrays for graphs.
// Returned by GetAccountDataGraphs (one row per account_id).
type AccountDataGraphsRow struct {
	AccountID   string
	Engagement  []float64
	Reach       []float64
	Impressions []float64
	Posts       []float64
	Buckets     []time.Time
}

// TopPostRow holds data for a single top-performing post across any platform.
// Returned by GetTopPosts.
type TopPostRow struct {
	PlatformType    string
	AccountID       string
	PostID          string
	Likes           int32
	Comments        int32
	Shares          int32
	Saves           int32
	PinClicks       int32
	OutboundClicks  int32
	DislikesCount   int32
	Permalink       string
	MediaType       string
	Thumbnail       string
	Category        string
	CreatedTime     time.Time
	TotalEngagement int32
	Views           int32
	Reach           int32
}
