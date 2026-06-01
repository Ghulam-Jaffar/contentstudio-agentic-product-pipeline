// Package ig_competitor provides the ClickHouse repository layer for Instagram competitor analytics.
// Tables queried: instagram_competitor_posts, instagram_competitor_insights.
package ig_competitor

import "time"

// DataTableMetricsRow holds per-competitor metrics for the data table view.
type DataTableMetricsRow struct {
	BusinessAccountID            string  `ch:"business_account_id"`
	AverageEngagement            float64 `ch:"averageEngagement"`
	AveragePostsPerWeek          float64 `ch:"averagePostsPerWeek"`
	EngagementRate               float64 `ch:"engagementRate"`
	DayOfWeek                    string  `ch:"dayOfWeek"`
	HourOfDay                    string  `ch:"hourOfDay"`
	AveragePostsPerDay           float64 `ch:"averagePostsPerDay"`
	AveragePostsPerDayEngagement float64 `ch:"averagePostsPerDayEngagement"`
	FollowingCount               int64   `ch:"followingCount"`
	FollowersCount               int64   `ch:"followersCount"`
}

// PostingActivityByTypeRow holds aggregated posting activity per media type.
type PostingActivityByTypeRow struct {
	MediaType           string  `ch:"mediaType"`
	MediaProductType    string  `ch:"mediaProductType"`
	AvgTotalEngagements float64 `ch:"avgTotalEngagements"`
	TotalPosts          int64   `ch:"totalPosts"`
	AvgEngagementRate   float64 `ch:"avgEngagementRate"`
	PostsPerWeek        float64 `ch:"postsPerWeek"`
	PostsPerDay         float64 `ch:"postsPerDay"`
	PostsPerHour        float64 `ch:"postsPerHour"`
	WeekCount           int32   `ch:"weekCount"`
	DayCount            int32   `ch:"dayCount"`
	HourCount           int32   `ch:"hourCount"`
}

// PostingActivityBySpecificTypeRow holds per-competitor metrics for a specific media type.
type PostingActivityBySpecificTypeRow struct {
	BusinessAccountID   string  `ch:"businessAccountId"`
	MediaType           string  `ch:"mediaType"`
	MediaProductType    string  `ch:"mediaProductType"`
	AvgTotalEngagements float64 `ch:"avgTotalEngagements"`
	AvgCountByWeek      float64 `ch:"avgCountByWeek"`
	AvgCountByDay       float64 `ch:"avgCountByDay"`
	AvgCountByHour      float64 `ch:"avgCountByHour"`
	TotalPosts          int64   `ch:"totalPosts"`
	WeekCount           int32   `ch:"weekCount"`
	DayCount            int32   `ch:"dayCount"`
	HourCount           int32   `ch:"hourCount"`
	AvgEngagementRate   float64 `ch:"avgEngagementRate"`
	FollowingCount      int64   `ch:"followingCount"`
	FollowersCount      int64   `ch:"followersCount"`
}

// PostingActivityTableRow holds table view metrics for a specific media type per competitor.
type PostingActivityTableRow struct {
	BusinessAccountID string  `ch:"businessAccountId"`
	Count             int64   `ch:"count"`
	TotalEngagement   int64   `ch:"totalEngagement"`
	EngagementRate    float64 `ch:"engagementRate"`
	FollowingCount    int64   `ch:"followingCount"`
	FollowersCount    int64   `ch:"followersCount"`
	MediaType         string  `ch:"mediaType"`
	MediaProductType  string  `ch:"mediaProductType"`
}

// TopPostRow represents a top/least performing Instagram competitor post.
type TopPostRow struct {
	BusinessAccountID string    `ch:"business_account_id"`
	Engagement        int64     `ch:"engagement"`
	PostID            string    `ch:"post_id"`
	Caption           string    `ch:"caption"`
	MediaType         string    `ch:"media_type"`
	MediaURL          string    `ch:"media_url"`
	Permalink         string    `ch:"permalink"`
	CreatedAt         time.Time `ch:"created_at"`
	LikeCount         int64     `ch:"like_count"`
	CommentsCount     int64     `ch:"comments_count"`
	MediaCount        int64     `ch:"media_count"`
	MediaProductType  string    `ch:"media_product_type"`
}

// TopHashtagRow holds aggregated hashtag metrics.
type TopHashtagRow struct {
	CompaniesUsing         []string `ch:"companies_using"`
	CompaniesName          []string `ch:"companies_name"`
	Count                  int64    `ch:"count"`
	TotalEngagement        int64    `ch:"total_engagement"`
	TotalFollowers         []int64  `ch:"total_followers"`
	EngagementPerFollower  float64  `ch:"engagement_per_follower"`
	EngagementRateByFollow float64  `ch:"engagement_rate_by_follower"`
	EngagementPerPost      float64  `ch:"engagement_per_post"`
	Tag                    string   `ch:"tag"`
}

// IndividualHashtagRow holds per-competitor data for a single hashtag.
type IndividualHashtagRow struct {
	Tag                    string  `ch:"tag"`
	Count                  int64   `ch:"count"`
	TotalEngagement        int64   `ch:"total_engagement"`
	TotalFollowers         int64   `ch:"total_followers"`
	EngagementPerPost      float64 `ch:"engagement_per_post"`
	EngagementPerFollower  float64 `ch:"engagement_per_follower"`
	EngagementRateByFollow float64 `ch:"engagement_rate_by_follower"`
	BusinessAccountID      string  `ch:"business_account_id"`
}

// BiographyRow holds biography data per competitor.
type BiographyRow struct {
	Biography         string `ch:"biography"`
	BiographyLength   int64  `ch:"biography_length"`
	BusinessAccountID string `ch:"business_account_id"`
	State             string `ch:"state"`
	Slug              string `ch:"slug"`
}

// FollowersGrowthRow holds per-competitor follower growth time-series.
type FollowersGrowthRow struct {
	BusinessAccountID    string   `ch:"business_account_id"`
	Dates                []string `ch:"dates"`
	TotalFollowingCount  []int64  `ch:"total_following_count"`
	TotalFollowedByCount []int64  `ch:"total_followed_by_count"`
}
