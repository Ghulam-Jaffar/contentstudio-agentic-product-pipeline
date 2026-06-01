// Package tiktok provides the ClickHouse repository layer for TikTok analytics.
// This file contains query result types returned by repository methods.
package tiktok

// SummaryResult holds the aggregated TikTok account summary for a date range.
type SummaryResult struct {
	TiktokID            string
	PageName            string
	Logo                string
	TotalLikes          int64
	TotalComments       int64
	TotalShares         int64
	TotalEngagements    int64
	TotalPosts          int64
	TotalFollowerCount  int64
	TotalFollowingCount int64
	TotalVideoViews     int64
}

// FollowersViewsResult holds time-series followers and views data.
type FollowersViewsResult struct {
	PlatformID         string
	DisplayName        string
	Logo               string
	FollowersCount     []int64
	ViewsPerDay        []int64
	FollowersCountDiff []int64
	ViewsPerDayDiff    []int64
	DayBucket          []string
}

// PostsEngagementResult holds time-series posts and engagement data.
type PostsEngagementResult struct {
	TiktokID           string
	PageName           string
	Logo               string
	DaysBucket         []string
	SumViewCount       []int64
	SumLikeCount       []int64
	SumCommentsCount   []int64
	SumShareCount      []int64
	SumEngagementCount []int64
	AvgEngagementRate  []int64
	PostCount          []int64
}

// DailyEngagementResult holds time-series daily engagement breakdown data.
type DailyEngagementResult struct {
	TiktokID           string
	PageName           string
	Logo               string
	TotalVideoLikes    []int64
	TotalVideoComments []int64
	TotalVideoShares   []int64
	DailyVideoLikes    []int64
	DailyVideoComments []int64
	DailyVideoShares   []int64
	TotalEngagement    []int64
	DailyEngagement    []int64
	DaysBucket         []string
}

// PostRow holds a single TikTok post's data for the top/least posts and posts list queries.
type PostRow struct {
	Category           string
	TiktokID           string
	PageName           string
	Logo               string
	ProfileLink        string
	PostID             string
	CoverImageURL      string
	ShareURL           string
	PostDescription    string
	Hashtags           []string
	Duration           int64
	Height             int64
	Width              int64
	Title              string
	EmbedHTML          string
	EmbedLink          string
	LikesCount         int64
	CommentsCount      int64
	SharesCount        int64
	ViewsCount         int64
	EngagementsCount   int64
	TotalEngagement    int64
	EngagementRate     float64
	InsertedAt         string
	CreatedTime        string
	TotalFollowerCount int64
	Total              int64
}
