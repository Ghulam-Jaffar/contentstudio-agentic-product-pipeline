// Package fb_competitor provides the ClickHouse repository layer for Facebook competitor analytics.
// Tables queried: facebook_competitor_posts, facebook_competitor_insights, facebook_competitor_media_assets.
package fb_competitor

import "time"

// DataTableMetricsRow holds per-competitor metrics for the data table view.
type DataTableMetricsRow struct {
	FacebookID                   string  `ch:"facebook_id"`
	AverageEngagement            float64 `ch:"averageEngagement"`
	AveragePostsPerWeek          float64 `ch:"averagePostsPerWeek"`
	EngagementRate               float64 `ch:"engagementRate"`
	DayOfWeek                    string  `ch:"dayOfWeek"`
	HourOfDay                    string  `ch:"hourOfDay"`
	AveragePostsPerDay           float64 `ch:"averagePostsPerDay"`
	AveragePostsPerDayEngagement float64 `ch:"averagePostsPerDayEngagement"`
	FollowersCount               int64   `ch:"followersCount"`
	FanCount                     int64   `ch:"fanCount"`
}

// PostingActivityByTypeRow holds aggregated posting activity per media type.
type PostingActivityByTypeRow struct {
	MediaType           string  `ch:"mediaType"`
	AvgTotalEngagements float64 `ch:"avgTotalEngagements"`
	TotalPosts          int64   `ch:"totalPosts"`
	TotalEngagement     int32   `ch:"total_engagement"`
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
	FacebookID          string  `ch:"facebook_id"`
	MediaType           string  `ch:"mediaType"`
	AvgTotalEngagements float64 `ch:"avgTotalEngagements"`
	TotalEngagement     int64   `ch:"totalEngagement"`
	AvgCountByWeek      float64 `ch:"avgCountByWeek"`
	AvgCountByDay       float64 `ch:"avgCountByDay"`
	AvgCountByHour      float64 `ch:"avgCountByHour"`
	TotalPosts          int64   `ch:"totalPosts"`
	WeekCount           int32   `ch:"weekCount"`
	DayCount            int32   `ch:"dayCount"`
	HourCount           int32   `ch:"hourCount"`
	AvgEngagementRate   float64 `ch:"avgEngagementRate"`
	FollowersCount      int64   `ch:"followersCount"`
}

// TopPostRow holds a single post with its media assets for top/least performing posts.
type TopPostRow struct {
	PostID         string    `ch:"post_id"`
	FacebookID     string    `ch:"facebook_id"`
	PageName       string    `ch:"page_name"`
	PageCategory   string    `ch:"page_category"`
	Biography      string    `ch:"biography"`
	FollowersCount int64     `ch:"followers_count"`
	FanCount       int64     `ch:"fan_count"`
	PostEngagement int64     `ch:"post_engagement"`
	Like           int64     `ch:"like"`
	Haha           int64     `ch:"haha"`
	Angry          int64     `ch:"angry"`
	Sad            int64     `ch:"sad"`
	Love           int64     `ch:"love"`
	Thankful       int64     `ch:"thankful"`
	TotalReactions int64     `ch:"total_post_reactions"`
	Comments       int64     `ch:"comments"`
	Shares         int64     `ch:"shares"`
	Caption        string    `ch:"caption"`
	MediaType      string    `ch:"media_type"`
	StatusType     string    `ch:"status_type"`
	Permalink      string    `ch:"permalink"`
	SharedFromName string    `ch:"shared_from_name"`
	SharedFromID   string    `ch:"shared_from_id"`
	SharedFromPic  string    `ch:"shared_from_pic"`
	Hashtags       []string  `ch:"hashtags"`
	DayOfWeek      int32     `ch:"day_of_week"`
	HourOfDay      int32     `ch:"hour_of_day"`
	CreatedAt      time.Time `ch:"created_at"`
	InsertedAt     time.Time `ch:"inserted_at"`
	Category       string    `ch:"category"`
	Image          string    `ch:"image"`
	MediaID        string    `ch:"media_id"`
	MediaCaption   string    `ch:"media_assets.caption"`
	MediaLink      string    `ch:"link"`
	AssetType      string    `ch:"asset_type"`
	CallToAction   string    `ch:"call_to_action"`
	MediaCreatedAt time.Time `ch:"media_assets.created_at"`
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
	FacebookID             string  `ch:"facebook_id"`
	Image                  string  `ch:"image"`
	Name                   string  `ch:"name"`
	Slug                   string  `ch:"slug"`
	FollowersCount         int64   `ch:"followersCount"`
}

// BiographyRow holds biography data per competitor.
type BiographyRow struct {
	Biography       string `ch:"biography"`
	BiographyLength int64  `ch:"biography_length"`
	FacebookID      string `ch:"facebook_id"`
	State           string `ch:"state"`
	Slug            string `ch:"slug"`
	Image           string `ch:"image"`
	Name            string `ch:"name"`
	FollowersCount  int64  `ch:"followersCount"`
}

// FollowersGrowthRow holds per-competitor follower growth time-series.
type FollowersGrowthRow struct {
	FacebookID              string   `ch:"facebook_id"`
	Dates                   []string `ch:"dates"`
	FollowersCount          []int64  `ch:"followers_count"`
	DatesWithFollowersCount []string `ch:"dates_with_followers_count"` // Will be handled as Tuple
}

// PostReactDistributionRow holds engagement aggregates for a single competitor.
type PostReactDistributionRow struct {
	FacebookID       string `ch:"facebook_id"`
	ImageUrl         string `ch:"imageUrl"`
	PageName         string `ch:"page_name"`
	TotalEngagements int32  `ch:"TotalEngagements"`
	TotalPosts       int32  `ch:"totalPosts"`
}

// PostReactDistByCompanyRow holds reaction breakdown for a single competitor.
type PostReactDistByCompanyRow struct {
	FacebookID     string `ch:"facebook_id"`
	PageName       string `ch:"page_name"`
	Image          string `ch:"image"`
	TotalLikes     int32  `ch:"total_likes"`
	TotalHahas     int32  `ch:"total_hahas"`
	TotalAngry     int32  `ch:"total_angry"`
	TotalSad       int32  `ch:"total_sad"`
	TotalThankful  int32  `ch:"total_thankful"`
	TotalLove      int32  `ch:"total_love"`
	TotalWow       int32  `ch:"total_wow"`
	TotalReactions int32  `ch:"total_post_reactions"`
	Comments       int32  `ch:"comments"`
	Shares         int32  `ch:"shares"`
	TotalPosts     int32  `ch:"total_posts"`
}

// PostTypeDistributionRow holds posting distribution per media type per competitor.
type PostTypeDistributionRow struct {
	FacebookID       string `ch:"facebook_id"`
	PageName         string `ch:"page_name"`
	Image            string `ch:"image"`
	MediaType        string `ch:"mediaType"`
	TotalEngagements int32  `ch:"TotalEngagements"`
	TotalPosts       int32  `ch:"totalPosts"`
	PostsPerWeek     int32  `ch:"postsPerWeek"`
	PostsPerDay      int32  `ch:"postsPerDay"`
	PostsPerHour     int32  `ch:"postsPerHour"`
	WeekCount        int32  `ch:"weekCount"`
	DayCount         int32  `ch:"dayCount"`
	HourCount        int32  `ch:"hourCount"`
}

// PostEngagementOverTimeRow holds daily engagement totals for a competitor.
type PostEngagementOverTimeRow struct {
	TotalEngagements int32     `ch:"total_engagements"`
	TotalPosts       int32     `ch:"total_posts"`
	Date             time.Time `ch:"date"`
}

// PostEngagementByCompetitorRow holds total engagement per competitor.
type PostEngagementByCompetitorRow struct {
	FacebookID       string `ch:"facebook_id"`
	PageName         string `ch:"page_name"`
	Image            string `ch:"image"`
	TotalPosts       int32  `ch:"total_posts"`
	TotalEngagements int32  `ch:"total_engagements"`
}
