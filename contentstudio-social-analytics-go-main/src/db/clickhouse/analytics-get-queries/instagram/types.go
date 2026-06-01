// Package instagram provides the ClickHouse repository layer for Instagram analytics.
// It contains query result types and repository methods that execute ClickHouse SQL queries
// migrated from the PHP Laravel InstagramAnalyticsBuilder (contentstudio-backend).
//
// Tables queried: instagram_posts, instagram_insights
package instagram

import "time"

// PostsSummaryResult holds aggregated post-level metrics from instagram_posts.
type PostsSummaryResult struct {
	DocCount        int64   `ch:"doc_count"`
	TotalEngagement int64   `ch:"total_engagement"`
	Likes           int64   `ch:"likes"`
	Comments        int64   `ch:"comments"`
	Saved           int64   `ch:"saved"`
	Reach           int64   `ch:"reach"`
	Impressions     int64   `ch:"impressions"`
	Views           int64   `ch:"views"`
	Stories         int64   `ch:"stories"`
	TotalPosts      int64   `ch:"total_posts"`
	EngagementRate  float64 `ch:"engagement_rate"`
}

// InsightsSummaryResult holds aggregated account-level metrics from instagram_insights.
type InsightsSummaryResult struct {
	ProfileViews    int64 `ch:"profile_views"`
	FollowersCount  int64 `ch:"followers_count"`
	FollowsCount    int64 `ch:"follows_count"`
	AccountsEngaged int64 `ch:"accounts_engaged"`
	Engagement      int64 `ch:"engagement"`
	Impressions     int64 `ch:"impressions"`
	Reach           int64 `ch:"reach"`
}

// AudienceResult holds time-series follower data with daily deltas.
type AudienceResult struct {
	ShowData       uint8       `ch:"show_data"`
	Followers      []int32     `ch:"followers"`
	FollowersDaily []int32     `ch:"followers_daily"`
	Buckets        []time.Time `ch:"buckets"`
}

// FollowerCount holds the most recent non-zero follower count.
// Used as a fallback when the current period starts with no follower data.
type FollowerCount struct {
	FollowersCount int32 `ch:"followers_count"`
}

// AudienceRollupResult holds aggregated follower metrics for current/previous period comparison.
type AudienceRollupResult struct {
	FollowerCount  int32 `ch:"follower_count"`
	FollowerGained int32 `ch:"follower_gained"`
}

// PublishingResult holds time-series publishing behaviour data.
type PublishingResult struct {
	Likes       []int32     `ch:"likes"`
	Comments    []int32     `ch:"comments"`
	Saved       []int32     `ch:"saved"`
	Engagement  []int32     `ch:"engagement"`
	Reach       []int32     `ch:"reach"`
	Impressions []int32     `ch:"impressions"`
	Views       []int32     `ch:"views"`
	TotalPosts  []int32     `ch:"total_posts"`
	Buckets     []time.Time `ch:"buckets"`
}

// PublishingRollupRow holds aggregated publishing metrics broken down by media type.
type PublishingRollupRow struct {
	MediaType  string `ch:"media_type"`
	TotalPosts int32  `ch:"total_posts"`
	Likes      int32  `ch:"likes"`
	Comments   int32  `ch:"comments"`
	Saved      int32  `ch:"saved"`
	Engagement int32  `ch:"engagement"`
	Reach      int32  `ch:"reach"`
	Views      int32  `ch:"views"`
}

// TopPostResult holds a single post's data from the top posts query.
type TopPostResult struct {
	InstagramID         string    `ch:"instagram_id"`
	MediaID             string    `ch:"media_id"`
	Caption             string    `ch:"caption"`
	MediaType           string    `ch:"media_type"`
	EntityType          string    `ch:"entity_type"`
	MediaURL            []string  `ch:"media_url"`
	VideoURL            []string  `ch:"video_url"`
	Permalink           string    `ch:"permalink"`
	LikeCount           int64     `ch:"like_count"`
	CommentsCount       int64     `ch:"comments_count"`
	Saved               int64     `ch:"saved"`
	Engagement          int64     `ch:"engagement"`
	Reach               int64     `ch:"reach"`
	Impressions         int64     `ch:"impressions"`
	Views               int64     `ch:"views"`
	Shares              int64     `ch:"shares"`
	ReelsAvgWatchTime   int64     `ch:"reels_avg_watch_time"`
	ReelsTotalWatchTime int64     `ch:"reels_total_watch_time"`
	Exits               int64     `ch:"exits"`
	Replies             int64     `ch:"replies"`
	Hashtags            []string  `ch:"hashtags"`
	DayOfWeek           string    `ch:"day_of_week"`
	HourOfDay           int64     `ch:"hour_of_day"`
	PostCreatedAt       time.Time `ch:"post_created_at"`
	StoredEventAt       time.Time `ch:"stored_event_at"`
}

// ActiveUsersHoursResult holds hourly online follower distribution.
type ActiveUsersHoursResult struct {
	Buckets      []int32 `ch:"buckets"`
	Values       []int32 `ch:"values"`
	HighestValue int32   `ch:"highest_value"`
	HighestHour  int32   `ch:"highest_hour"`
}

// ActiveUsersDaysResult holds day-of-week follower activity distribution.
type ActiveUsersDaysResult struct {
	Buckets      []string `ch:"buckets"`
	Values       []int32  `ch:"values"`
	HighestValue int32    `ch:"highest_value"`
	HighestDay   string   `ch:"highest_day"`
}

// ImpressionsResult holds time-series impressions data.
type ImpressionsResult struct {
	ShowData    uint8       `ch:"show_data"`
	Buckets     []time.Time `ch:"buckets"`
	Impressions []int32     `ch:"impressions"`
}

// ImpressionsRollupResult holds aggregated impressions totals for period comparison.
type ImpressionsRollupResult struct {
	TotalImpressions int64   `ch:"total_impressions"`
	AvgImpressions   float64 `ch:"avg_impressions"`
}

// EngagementResult holds time-series engagement data.
type EngagementResult struct {
	ShowData   uint8       `ch:"show_data"`
	Buckets    []time.Time `ch:"buckets"`
	Engagement []int32     `ch:"engagement"`
	Comments   []int32     `ch:"comments"`
	Reactions  []int32     `ch:"reactions"`
	DocCount   []int32     `ch:"doc_count"`
}

// EngagementRollupResult holds aggregated engagement totals for period comparison.
type EngagementRollupResult struct {
	Engagement    int64   `ch:"engagement"`
	AvgEngagement float64 `ch:"avg_engagement"`
	Comments      int64   `ch:"comments"`
	Reactions     int64   `ch:"reactions"`
	Saved         int64   `ch:"saved"`
	Count         int64   `ch:"count"`
}

// HashtagsResult holds the top 30 hashtags with their engagement metrics.
type HashtagsResult struct {
	Name       []string `ch:"name"`
	Engagement []int32  `ch:"engagement"`
	Likes      []int32  `ch:"likes"`
	Comments   []int32  `ch:"comments"`
	Saved      []int32  `ch:"saved"`
	Posts      []int32  `ch:"posts"`
}

// HashtagsRollupResult holds aggregated hashtag totals for period comparison.
type HashtagsRollupResult struct {
	TotalEngagement     int32 `ch:"total_engagement"`
	TotalLikes          int32 `ch:"total_likes"`
	TotalComments       int32 `ch:"total_comments"`
	TotalSaves          int32 `ch:"total_saves"`
	TotalUniqueHashtags int32 `ch:"total_unique_hashtags"`
	TotalHashtagUses    int32 `ch:"total_hashtag_uses"`
}

// StoriesResult holds time-series stories performance data.
type StoriesResult struct {
	ShowData            uint8       `ch:"show_data"`
	Buckets             []time.Time `ch:"buckets"`
	AvgStoryImpressions []float64   `ch:"avg_story_impressions"`
	StoryImpressions    []int32     `ch:"story_impressions"`
	StoryReach          []int32     `ch:"story_reach"`
	StoryReply          []int32     `ch:"story_reply"`
	StoryExits          []int32     `ch:"story_exits"`
	StoryTapsForward    []int32     `ch:"story_taps_forward"`
	StoryTapsBack       []int32     `ch:"story_taps_back"`
	PublishedStories    []int32     `ch:"published_stories"`
}

// StoriesRollupResult holds aggregated stories totals for period comparison.
type StoriesRollupResult struct {
	StoryImpressions    int64   `ch:"story_impressions"`
	AvgStoryImpressions float64 `ch:"avg_story_impressions"`
	StoryReach          int64   `ch:"story_reach"`
	StoryReply          int64   `ch:"story_reply"`
	StoryExits          int64   `ch:"story_exits"`
	StoryTapsForward    int64   `ch:"story_taps_forward"`
	StoryTapsBack       int64   `ch:"story_taps_back"`
	PublishedStories    int64   `ch:"published_stories"`
}

// ReelsResult holds time-series reels performance data.
type ReelsResult struct {
	ShowData       uint8       `ch:"show_data"`
	Buckets        []time.Time `ch:"buckets"`
	TotalPosts     []int32     `ch:"total_posts"`
	Engagement     []int32     `ch:"engagement"`
	Likes          []int32     `ch:"likes"`
	Comments       []int32     `ch:"comments"`
	Saves          []int32     `ch:"saves"`
	Shares         []int32     `ch:"shares"`
	AvgWatchTime   []float64   `ch:"avg_watch_time"`
	TotalWatchTime []int64     `ch:"total_watch_time"`
}

// ReelsRollupResult holds aggregated reels totals for period comparison.
type ReelsRollupResult struct {
	Engagement     int64   `ch:"engagement"`
	Likes          int64   `ch:"likes"`
	Comments       int64   `ch:"comments"`
	Saves          int64   `ch:"saves"`
	TotalPosts     int64   `ch:"total_posts"`
	Shares         int64   `ch:"shares"`
	AvgWatchTime   float64 `ch:"avg_watch_time"`
	TotalWatchTime int64   `ch:"total_watch_time"`
}

// DemographicsResult holds the latest audience demographic arrays from instagram_insights.
// Each []string stores encoded demographic entries (e.g. JSON or "key:value" pairs).
type DemographicsResult struct {
	AudienceAge       []string `ch:"audience_age"`
	AudienceGender    []string `ch:"audience_gender"`
	AudienceGenderAge []string `ch:"audience_gender_age"`
}

// LocationResult holds the latest audience location arrays from instagram_insights.
type LocationResult struct {
	AudienceCity    []string `ch:"audience_city"`
	AudienceCountry []string `ch:"audience_country"`
}
