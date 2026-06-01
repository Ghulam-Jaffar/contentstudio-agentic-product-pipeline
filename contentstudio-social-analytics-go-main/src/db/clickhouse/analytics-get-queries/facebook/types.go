// Package facebook provides the ClickHouse repository layer for Facebook analytics.
// It contains query result types and repository methods that execute ClickHouse SQL queries
// migrated from the PHP Laravel FacebookAnalyticsBuilder (contentstudio-backend).
//
// Tables queried: facebook_posts, facebook_insights, facebook_reels_insights,
// facebook_video_insights, facebook_media_assets.
package facebook

import "time"

// SummaryResult holds the combined post and page-level summary metrics for a date range.
// Post fields come from facebook_posts; page fields come from facebook_insights.
type SummaryResult struct {
	// DocCount is the total number of deduplicated posts in the period.
	DocCount               int32 `ch:"doc_count"`
	// TotalEngagement is reactions + comments + shares + post_clicks across all posts.
	TotalEngagement        int32 `ch:"total_engagement"`
	Reactions              int32 `ch:"reactions"`
	Comments               int32 `ch:"comments"`
	PostsClicks            int32 `ch:"posts_clicks"`
	Impressions            int32 `ch:"impressions"`
	Reach                  int32 `ch:"reach"`
	Repost                 int32 `ch:"repost"`
	PositiveSentiment      int32 `ch:"positive_sentiment"`
	NegativeSentiment      int32 `ch:"negative_sentiment"`
	PageImpressions        int32 `ch:"page_impressions"`
	PageImpressionsPaid    int32 `ch:"page_impressions_paid"`
	PageImpressionsOrganic int32 `ch:"page_impressions_organic"`
	PageEngagements        int32 `ch:"page_engagements"`
	PagePositiveFeedback   int32 `ch:"page_positive_feedback"`
	PageNegativeFeedback   int32 `ch:"page_negative_feedback"`
	// FanCount is the maximum page_fans value observed across the period (point-in-time snapshot).
	FanCount               int32 `ch:"fan_count"`
	TalkingAboutCount      int32 `ch:"talking_about_count"`
	PageFollows            int32 `ch:"page_follows"`
}

// PostsSummaryResult holds aggregated post-level metrics from facebook_posts.
// Run concurrently with InsightsSummaryResult for maximum throughput.
type PostsSummaryResult struct {
	DocCount        int32 `ch:"doc_count"`
	TotalEngagement int32 `ch:"total_engagement"`
	Reactions       int32 `ch:"reactions"`
	Comments        int32 `ch:"comments"`
	PostsClicks     int32 `ch:"posts_clicks"`
	Impressions     int32 `ch:"impressions"`
	Reach           int32 `ch:"reach"`
	Repost          int32 `ch:"repost"`
}

// InsightsSummaryResult holds aggregated page-level metrics from facebook_insights.
// Run concurrently with PostsSummaryResult for maximum throughput.
type InsightsSummaryResult struct {
	PositiveSentiment      int32 `ch:"positive_sentiment"`
	NegativeSentiment      int32 `ch:"negative_sentiment"`
	PageImpressions        int32 `ch:"page_impressions"`
	PageImpressionsPaid    int32 `ch:"page_impressions_paid"`
	PageImpressionsOrganic int32 `ch:"page_impressions_organic"`
	PageEngagements        int32 `ch:"page_engagements"`
	PagePositiveFeedback   int32 `ch:"page_positive_feedback"`
	PageNegativeFeedback   int32 `ch:"page_negative_feedback"`
	// FanCount is the maximum page_fans snapshot across the date range.
	FanCount               int32 `ch:"fan_count"`
	TalkingAboutCount      int32 `ch:"talking_about_count"`
	PageFollows            int32 `ch:"page_follows"`
}

// AudienceGrowthResult holds time-series fan/engagement data from facebook_insights.
// FanCount is forward-filled via arrayFill to bridge zero-value gaps.
// PageFansDaily is the daily delta computed with lagInFrame() over an ordered window.
type AudienceGrowthResult struct {
	// ShowData is 1 if the fan_count array is non-empty, 0 otherwise.
	ShowData         uint8       `ch:"show_data"`
	// FanCount is the cumulative fan count per day, with zeros forward-filled.
	FanCount         []int32     `ch:"fan_count"`
	// PageFansDaily is the day-over-day change in fan count.
	PageFansDaily    []int32     `ch:"page_fans_daily"`
	PageFansByLike   []int32     `ch:"page_fans_by_like"`
	PageFansByUnlike []int32     `ch:"page_fans_by_unlike"`
	PageImpressions  []int32     `ch:"page_impressions"`
	PageEngagements  []int32     `ch:"page_engagements"`
	Buckets          []time.Time `ch:"buckets"`
}

// AudienceGrowthRollupResult holds aggregated fan metrics for current/previous period comparison.
type AudienceGrowthRollupResult struct {
	AvgPageFansByLike   float64 `ch:"avg_page_fans_by_like"`
	AvgPageFansByUnlike float64 `ch:"avg_page_fans_by_unlike"`
	FanCount            int32   `ch:"fan_count"`
	TalkingAboutCount   int32   `ch:"talking_about_count"`
	DocCount            int32   `ch:"doc_count"`
}

// LastFollowerCounts holds the most recent non-zero fan counts from facebook_insights.
// Used as a fallback when the current period starts with zero fan data.
type LastFollowerCounts struct {
	PageFans         int32 `ch:"page_fans"`
	PageFansByLike   int32 `ch:"page_fans_by_like"`
	PageFansByUnlike int32 `ch:"page_fans_by_unlike"`
}

// PublishingBehaviourResult holds time-series engagement and reach data from facebook_posts.
// Each slice entry corresponds to a single date bucket aligned with Buckets.
type PublishingBehaviourResult struct {
	ReactionsEngagement []int32     `ch:"reactions_engagement"`
	CommentsEngagement  []int32     `ch:"comments_engagement"`
	SharesEngagement    []int32     `ch:"shares_engagement"`
	PaidImpressions     []int32     `ch:"paid_impressions"`
	OrganicImpressions  []int32     `ch:"organic_impressions"`
	ViralImpressions    []int32     `ch:"viral_impressions"`
	PaidReach           []int32     `ch:"paid_reach"`
	OrganicReach        []int32     `ch:"organic_reach"`
	ViralReach          []int32     `ch:"viral_reach"`
	Buckets             []time.Time `ch:"buckets"`
	PostCount           []int32     `ch:"post_count"`
}

// PublishingRollupResult holds aggregated post-level metrics for current/previous period comparison.
type PublishingRollupResult struct {
	DocCount        int32 `ch:"doc_count"`
	TotalEngagement int32 `ch:"total_engagement"`
	Reactions       int32 `ch:"reactions"`
	Comments        int32 `ch:"comments"`
	PostClicks      int32 `ch:"post_clicks"`
	Impressions     int32 `ch:"impressions"`
	Shares          int32 `ch:"shares"`
}

// TopPostRow holds a single post's data joined with its media assets from facebook_media_assets.
// Posts are deduplicated via max(saving_time); each asset produces an extra row that is
// collapsed into MediaAssets in the service layer.
type TopPostRow struct {
	PageName                     string
	PageID                       string
	PostID                       string
	Permalink                    string
	StatusType                   string
	MediaType                    string
	VideoID                      string
	Category                     string
	PublishedBy                  string
	PublishedByURL               string
	SharedFromName               string
	SharedFromID                 string
	SharedFromLink               string
	Like                         int64
	Love                         int64
	Haha                         int64
	Wow                          int64
	Sad                          int64
	Angry                        int64
	// Total is the sum of all reaction types (like + love + haha + wow + sad + angry).
	Total                        int64
	Shares                       int64
	Comments                     int64
	PostClicks                   int64
	TotalEngagement              float64
	PostEngagedUsers             int64
	DayOfWeek                    string
	HourOfDay                    int64
	CreatedTime                  time.Time
	UpdatedTime                  time.Time
	SavingTime                   time.Time
	MessageTags                  string
	PostMetadata                 string
	Caption                      string
	Description                  string
	FullPicture                  string
	Link                         string
	PostImpressions              int64
	PostImpressionsUnique        int64
	PostImpressionsPaid          int64
	PostImpressionsPaidUnique    int64
	PostImpressionsOrganic       int64
	PostImpressionsOrganicUnique int64
	PostImpressionsViral         int64
	PostImpressionsViralUnique   int64
	PostVideoViews               int64
	TotalImpressions             int64
	// MediaID through AssetCreatedAt are populated from the LEFT JOIN with facebook_media_assets.
	MediaID                      string
	MediaCaption                 string
	MediaLink                    string
	AssetType                    string
	CallToAction                 string
	AssetCreatedAt               time.Time
}

// ActiveUsersHoursResult holds the average number of fans online per hour of the day.
// Data is sourced from the page_fans_online field in facebook_insights, stored as
// a '$'-delimited array string (e.g. "0$120$1$85$...").
type ActiveUsersHoursResult struct {
	Buckets      []int32 `ch:"buckets"`
	Values       []int32 `ch:"values"`
	HighestValue int32   `ch:"highest_value"`
	HighestHour  int32   `ch:"highest_hour"`
}

// ActiveUsersDaysResult holds the count of posts published on each day of the week.
// The day_of_week value is sourced from facebook_insights and mapped to Monday–Sunday.
type ActiveUsersDaysResult struct {
	Buckets      []string `ch:"buckets"`
	Values       []int32  `ch:"values"`
	HighestValue int32    `ch:"highest_value"`
	HighestDay   string   `ch:"highest_day"`
}

// ImpressionsResult holds time-series page impression data from facebook_insights.
type ImpressionsResult struct {
	PageImpressions []int32     `ch:"page_impressions"`
	Buckets         []time.Time `ch:"buckets"`
}

// ImpressionsRollupResult holds aggregated impression totals with daily and weekly averages.
type ImpressionsRollupResult struct {
	TotalImpressions      int32   `ch:"total_impressions"`
	AvgImpressionsPerDay  float64 `ch:"avg_impressions_per_day"`
	AvgImpressionsPerWeek float64 `ch:"avg_impressions_per_week"`
}

// EngagementResult holds time-series page engagement data from facebook_insights.
type EngagementResult struct {
	PageEngagements []int32     `ch:"page_engagements"`
	Buckets         []time.Time `ch:"buckets"`
}

// EngagementRollupResult holds aggregated engagement totals with daily and weekly averages.
type EngagementRollupResult struct {
	PageEngagements       int32   `ch:"page_engagements"`
	AvgEngagementsPerDay  float64 `ch:"avg_engagements_per_day"`
	AvgEngagementsPerWeek float64 `ch:"avg_engagements_per_week"`
}

// ReelsAnalyticsResult holds time-series reels performance data.
// Data is joined from facebook_reels_insights (views, watch time) and facebook_posts (engagement).
type ReelsAnalyticsResult struct {
	Buckets             []time.Time `ch:"buckets"`
	TotalReels          []int32     `ch:"total_reels"`
	// TotalSecondsWatched is total_time_watched_in_ms / 1000, rounded to 2 decimal places.
	TotalSecondsWatched []float64   `ch:"total_seconds_watched"`
	// InitialPlays maps to the play_count field in facebook_reels_insights.
	InitialPlays        []int32     `ch:"initial_plays"`
	Engagement          []int32     `ch:"engagement"`
	Reactions           []int32     `ch:"reactions"`
	Comments            []int32     `ch:"comments"`
	Shares              []int32     `ch:"shares"`
	// ShowData is the total reel count across the period; 0 means no reels data to display.
	ShowData            int32       `ch:"show_data"`
}

// ReelsRollupResult holds aggregated reels totals for current/previous period comparison.
type ReelsRollupResult struct {
	TotalReels            int32   `ch:"total_reels"`
	// AverageSecondsWatched is total_seconds_watched / initial_plays, protected against division by zero.
	AverageSecondsWatched float64 `ch:"average_seconds_watched"`
	TotalSecondsWatched   int32   `ch:"total_seconds_watched"`
	InitialPlays          int32   `ch:"initial_plays"`
	Reach                 int32   `ch:"reach"`
	Engagement            int32   `ch:"engagement"`
	Reactions             int32   `ch:"reactions"`
	Comments              int32   `ch:"comments"`
	Shares                int32   `ch:"shares"`
}

// VideoInsightsResult holds time-series video performance data.
// Data is joined from facebook_video_insights (view times, view counts) and facebook_posts (engagement).
// View time values are in seconds (total_video_view_total_time / 1000).
type VideoInsightsResult struct {
	Buckets         []time.Time `ch:"buckets"`
	TotalViewTime   []float64   `ch:"total_view_time"`
	OrganicViewTime []float64   `ch:"organic_view_time"`
	PaidViewTime    []float64   `ch:"paid_view_time"`
	TotalViews      []int32     `ch:"total_views"`
	OrganicViews    []int32     `ch:"organic_views"`
	PaidViews       []int32     `ch:"paid_views"`
	Comments        []int32     `ch:"comments"`
	Reactions       []int32     `ch:"reactions"`
	Shares          []int32     `ch:"shares"`
	TotalPosts      []int32     `ch:"total_posts"`
}

// VideoRollupResult holds aggregated video totals for current/previous period comparison.
type VideoRollupResult struct {
	TotalViewTime   float64 `ch:"total_view_time"`
	OrganicViewTime float64 `ch:"organic_view_time"`
	PaidViewTime    float64 `ch:"paid_view_time"`
	TotalViews      int32   `ch:"total_views"`
	OrganicViews    int32   `ch:"organic_views"`
	PaidViews       int32   `ch:"paid_views"`
	Comments        int32   `ch:"comments"`
	Reactions       int32   `ch:"reactions"`
	Shares          int32   `ch:"shares"`
	TotalPosts      int32   `ch:"total_posts"`
}

// AudienceGenderResult holds the fan count split by gender from facebook_insights.
// The page_fans_gender field is a '$'-delimited string that is parsed in ClickHouse.
type AudienceGenderResult struct {
	Fans int32 `ch:"fans"`
	M    int32 `ch:"M"`
	F    int32 `ch:"F"`
	U    int32 `ch:"U"`
}

// MaxGenderAgeResult holds the gender+age bucket with the highest fan count.
// Derived from page_fans_gender_age in facebook_insights.
type MaxGenderAgeResult struct {
	MaxValue int32  `ch:"max_value"`
	Age      string `ch:"age"`
	Gender   string `ch:"gender"`
}

// AudienceAgeResult holds fan counts broken down by age bracket.
// Derived from the page_fans_age '$'-delimited string in facebook_insights.
type AudienceAgeResult struct {
	Age65Plus int32 `ch:"65+"`
	Age55To64 int32 `ch:"55-64"`
	Age45To54 int32 `ch:"45-54"`
	Age35To44 int32 `ch:"35-44"`
	Age25To34 int32 `ch:"25-34"`
	Age18To34 int32 `ch:"18-34"`
	Age13To17 int32 `ch:"13-17"`
}
