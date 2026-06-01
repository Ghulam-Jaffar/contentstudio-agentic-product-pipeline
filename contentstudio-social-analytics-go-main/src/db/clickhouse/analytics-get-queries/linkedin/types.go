// Package linkedin provides the ClickHouse repository layer for LinkedIn analytics.
// It contains query result types and repository methods that execute ClickHouse SQL queries
// migrated from the PHP Laravel LinkedInAnalyticsBuilder (contentstudio-backend).
//
// Tables queried: linkedin_posts, linkedin_insights
package linkedin

import "time"

// SummaryResult holds aggregated summary metrics for a LinkedIn page over a date range.
// Combines post-level metrics (from linkedin_posts) and page-level insights (from linkedin_insights).
type SummaryResult struct {
	PostComments       int32   `ch:"post_comments"`
	PostLikes          int32   `ch:"post_likes"`
	TotalEngagement    int32   `ch:"total_engagement"`
	TotalPosts         int32   `ch:"total_posts"`
	PostShares         int32   `ch:"post_shares"`
	PostClicks         int32   `ch:"post_clicks"`
	Followers          int32   `ch:"followers"`
	PageViews          int32   `ch:"page_views"`
	PageReach          int32   `ch:"page_reach"`
	PageShares         int32   `ch:"page_shares"`
	PageComments       int32   `ch:"page_comments"`
	PageReactions      int32   `ch:"page_reactions"`
	PageImpressions    int32   `ch:"page_impressions"`
	PageUniqueVisitors int32   `ch:"page_unique_visitors"`
	EngagementRate     float64 `ch:"engagement_rate"`
	PostEngagementRate float64 `ch:"post_engagement_rate"`
}

// PostsSummaryResult holds aggregated post-level metrics from linkedin_posts.
// Run concurrently with InsightsSummaryResult for maximum throughput.
type PostsSummaryResult struct {
	PostComments       int32   `ch:"post_comments"`
	PostLikes          int32   `ch:"post_likes"`
	TotalEngagement    int32   `ch:"total_engagement"`
	TotalPosts         int32   `ch:"total_posts"`
	PostShares         int32   `ch:"post_shares"`
	PostClicks         int32   `ch:"post_clicks"`
	PostEngagementRate float64 `ch:"post_engagement_rate"`
}

// InsightsSummaryResult holds aggregated page-level metrics from linkedin_insights.
// Run concurrently with PostsSummaryResult for maximum throughput.
type InsightsSummaryResult struct {
	Followers          int32   `ch:"followers"`
	PageViews          int32   `ch:"page_views"`
	PageReach          int32   `ch:"page_reach"`
	PageShares         int32   `ch:"page_shares"`
	PageComments       int32   `ch:"page_comments"`
	PageReactions      int32   `ch:"page_reactions"`
	PageImpressions    int32   `ch:"page_impressions"`
	PageUniqueVisitors int32   `ch:"page_unique_visitors"`
	EngagementRate     float64 `ch:"engagement_rate"`
}

// AudienceResult holds time-series audience growth data with organic, paid, and total follower counts.
// Daily values are computed as differences between consecutive cumulative counts.
type AudienceResult struct {
	ShowData              uint8       `ch:"show_data"`
	OrganicFollowerCount  []int32     `ch:"organic_follower_count"`
	OrganicFollowersDaily []int32     `ch:"organic_followers_daily"`
	PaidFollowerCount     []int32     `ch:"paid_follower_count"`
	PaidFollowersDaily    []int32     `ch:"paid_followers_daily"`
	TotalFollowerCount    []int32     `ch:"total_follower_count"`
	TotalFollowersDaily   []int32     `ch:"total_followers_daily"`
	Buckets               []time.Time `ch:"buckets"`
}

// FollowerCounts holds the most recent non-zero follower counts.
// Used as a fallback when the current period has no follower data.
type FollowerCounts struct {
	TotalFollowerCount   int32 `ch:"total_follower_count"`
	OrganicFollowerCount int32 `ch:"organic_follower_count"`
	PaidFollowerCount    int32 `ch:"paid_follower_count"`
}

// AudienceRollupResult holds aggregated audience metrics for current/previous period comparison.
type AudienceRollupResult struct {
	OrganicFollowerCount int32   `ch:"organic_follower_count"`
	PaidFollowerCount    int32   `ch:"paid_follower_count"`
	TotalFollowerCount   int32   `ch:"total_follower_count"`
	AvgFollowerCount     float64 `ch:"avg_follower_count"`
}

// PageViewsResult holds time-series page view data split by device type (desktop/mobile).
// Both cumulative and daily values are provided along with date buckets.
type PageViewsResult struct {
	DesktopPageViews      []int32     `ch:"desktop_page_views"`
	MobilePageViews       []int32     `ch:"mobile_page_views"`
	TotalPageViews        []int32     `ch:"total_page_views"`
	DesktopPageViewsDaily []int32     `ch:"desktop_page_views_daily"`
	MobilePageViewsDaily  []int32     `ch:"mobile_page_views_daily"`
	TotalPageViewsDaily   []int32     `ch:"total_page_views_daily"`
	ShowData              int32       `ch:"show_data"`
	Buckets               []time.Time `ch:"buckets"`
}

// PageViewsRollupResult holds aggregated page view totals for current/previous period comparison.
type PageViewsRollupResult struct {
	TotalPageViews   int32   `ch:"total_page_views"`
	DesktopPageViews int32   `ch:"desktop_page_views"`
	MobilePageViews  int32   `ch:"mobile_page_views"`
	AvgPageViews     float64 `ch:"avg_page_views"`
}

// PublishingResult holds time-series publishing behaviour data including engagement metrics per day.
type PublishingResult struct {
	Likes          []int32     `ch:"likes"`
	Comments       []int32     `ch:"comments"`
	Shares         []int32     `ch:"shares"`
	Clicks         []int32     `ch:"clicks"`
	EngagementRate []float32   `ch:"engagement_rate"`
	Impressions    []int32     `ch:"impressions"`
	TotalPosts     []int32     `ch:"total_posts"`
	Engagement     []int32     `ch:"engagement"`
	Reach          []int32     `ch:"reach"`
	Buckets        []time.Time `ch:"buckets"`
}

// PublishingRollupRow holds aggregated publishing metrics broken down by media type.
// Includes a "total" row with summed values across all media types.
type PublishingRollupRow struct {
	MediaType   string `ch:"media_type"`
	TotalPosts  int32  `ch:"total_posts"`
	Likes       int32  `ch:"likes"`
	Comments    int32  `ch:"comments"`
	Shares      int32  `ch:"shares"`
	Clicks      int32  `ch:"clicks"`
	Engagements int32  `ch:"engagements"`
	Impressions int32  `ch:"impressions"`
	Reach       int32  `ch:"reach"`
}

// TopPostResult holds a single post's data returned from the top posts query.
// Posts are deduplicated using max(saving_time) per post_id and ordered by the requested metric.
type TopPostResult struct {
	LinkedinID      string    `ch:"linkedin_id"`
	PostID          string    `ch:"post_id"`
	Activity        string    `ch:"activity"`
	MediaType       string    `ch:"media_type"`
	ArticleURL      string    `ch:"article_url"`
	ArticleTitle    string    `ch:"article_title"`
	PostData        string    `ch:"post_data"`
	Image           string    `ch:"image"`
	Media           []string  `ch:"media"`
	Type            string    `ch:"type"`
	Hashtags        []string  `ch:"hashtags"`
	Comments        int64     `ch:"comments"`
	TotalEngagement float64   `ch:"total_engagement"`
	Favorites       int64     `ch:"favorites"`
	Title           string    `ch:"title"`
	DayOfWeek       string    `ch:"day_of_week"`
	HourOfDay       int64     `ch:"hour_of_day"`
	CreatedAt       time.Time `ch:"created_at"`
	SavingTime      time.Time `ch:"saving_time"`
	PollData        string    `ch:"poll_data"`
	Reach           int64     `ch:"reach"`
	Repost          int64     `ch:"repost"`
	PostClicks      int64     `ch:"post_clicks"`
	Impressions     int64     `ch:"impressions"`
	PublishedAt     time.Time `ch:"published_at"`
}

// PostsPerDayResult holds the count of posts published on each day of the week.
type PostsPerDayResult struct {
	Monday    int32 `ch:"Monday"`
	Tuesday   int32 `ch:"Tuesday"`
	Wednesday int32 `ch:"Wednesday"`
	Thursday  int32 `ch:"Thursday"`
	Friday    int32 `ch:"Friday"`
	Saturday  int32 `ch:"Saturday"`
	Sunday    int32 `ch:"Sunday"`
}

// HashtagsResult holds the top 30 hashtags with their engagement metrics as parallel arrays.
type HashtagsResult struct {
	Name        []string `ch:"name"`
	Engagements []int32  `ch:"engagements"`
	Likes       []int32  `ch:"likes"`
	Comments    []int32  `ch:"comments"`
	Shares      []int32  `ch:"shares"`
	Posts       []int32  `ch:"posts"`
}

// HashtagsRollupResult holds aggregated hashtag totals for current/previous period comparison.
type HashtagsRollupResult struct {
	TotalHashtags    int32 `ch:"total_hashtags"`
	TotalTimesUsed   int32 `ch:"total_times_used"`
	TotalLikes       int32 `ch:"total_likes"`
	TotalComments    int32 `ch:"total_comments"`
	TotalShares      int32 `ch:"total_shares"`
	TotalEngagement  int32 `ch:"total_engagement"`
	TotalImpressions int32 `ch:"total_impressions"`
	TotalReach       int32 `ch:"total_reach"`
}

// DemographicsResult holds follower demographics data stored as JSON strings in ClickHouse.
// Each JSON field maps category names to follower counts (e.g., {"Engineering": 150, "Marketing": 80}).
type DemographicsResult struct {
	FollowersBySeniority string `ch:"followers_by_seniority"`
	FollowersByIndustry  string `ch:"followers_by_industry"`
	FollowersByCountry   string `ch:"followers_by_country"`
	FollowersByCity      string `ch:"followers_by_city"`
	TotalFollowerCount   int64  `ch:"totalFollowerCount"`
}
