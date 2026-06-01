// Package youtube defines request and response types for the YouTube analytics API.
// These types map directly to the JSON contracts expected by the ContentStudio frontend,
// preserving the same field names and structure as the PHP Laravel API responses.
//
// Request types include validation logic and conversion to ClickHouse query parameters.
// Response types match the frontend's expected JSON shape for each analytics widget.
package youtube

import (
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// --- Request types ---

// YoutubeRequest is the base request for all YouTube analytics endpoints.
// Query params: workspace_id, youtube_id, start_date, end_date, timezone.
type YoutubeRequest struct {
	WorkspaceID string `json:"workspace_id"`
	YoutubeID   string `json:"youtube_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Timezone    string `json:"timezone"`
}

// TopVideosRequest extends YoutubeRequest with pagination and sort options.
type TopVideosRequest struct {
	YoutubeRequest
	Limit   int    `json:"limit"`
	OrderBy string `json:"order_by"`
}

// AIInsightsRequest is the request for YouTube AI insights.
type AIInsightsRequest struct {
	WorkspaceID string      `json:"workspace_id"`
	YoutubeID   string      `json:"youtube_id"`
	Date        interface{} `json:"date"`
	Timezone    string      `json:"timezone"`
	Type        string      `json:"type"`
	Limit       int         `json:"limit"`
	Language    string      `json:"language,omitempty"`
}

// --- Response types ---

// SummaryMetrics holds aggregated metrics for a single period.
// JSON field names use PHP-compatible singular forms (like, dislike, comment, share).
type SummaryMetrics struct {
	WatchTime       int64   `json:"watch_time"`
	AvgViewDuration float64 `json:"avg_view_duration"`
	Likes           int64   `json:"like"`
	Dislikes        int64   `json:"dislike"`
	Comments        int64   `json:"comment"`
	Shares          int64   `json:"share"`
	Engagement      int64   `json:"engagement"`
	Subscribers     int64   `json:"subscribers"`
	Views           int64   `json:"views"`
	Videos          int64   `json:"videos"`
}

// SummaryChangeMetrics holds percentage changes for each metric (float values).
type SummaryChangeMetrics struct {
	WatchTime       float64 `json:"watch_time"`
	AvgViewDuration float64 `json:"avg_view_duration"`
	Likes           float64 `json:"like"`
	Dislikes        float64 `json:"dislike"`
	Comments        float64 `json:"comment"`
	Shares          float64 `json:"share"`
	Engagement      float64 `json:"engagement"`
	Subscribers     float64 `json:"subscribers"`
	Views           float64 `json:"views"`
	Videos          float64 `json:"videos"`
}

// SummaryOverview groups current/previous metrics with percentage changes and absolute differences.
type SummaryOverview struct {
	Current    *SummaryMetrics       `json:"current"`
	Previous   *SummaryMetrics       `json:"previous"`
	Percentage *SummaryChangeMetrics `json:"percentage"`
	Difference *SummaryMetrics       `json:"difference"`
}

// SummaryResponse holds current and previous period summary metrics.
type SummaryResponse struct {
	Status   bool             `json:"status"`
	Overview *SummaryOverview `json:"overview"`
}

// SubscriberTrendResponse holds time-series subscriber data.
type SubscriberTrendResponse struct {
	Status                 bool     `json:"status"`
	ShowData               int32    `json:"show_data"`
	SubscribersGainedDaily []int32  `json:"subscribers_gained_daily"`
	SubscribersTotal       []int32  `json:"subscribers_total"`
	Buckets                []string `json:"buckets"`
	AggregationLevel       string   `json:"aggregation_level,omitempty"`
}

// EngagementTrendResponse holds time-series engagement metrics.
type EngagementTrendResponse struct {
	Status           bool     `json:"status"`
	ShowData         int32    `json:"show_data"`
	LikeDaily        []int32  `json:"like_daily"`
	LikeTotal        []int32  `json:"like_total"`
	DislikeDaily     []int32  `json:"dislike_daily"`
	DislikeTotal     []int32  `json:"dislike_total"`
	ShareDaily       []int32  `json:"share_daily"`
	ShareTotal       []int32  `json:"share_total"`
	CommentDaily     []int32  `json:"comment_daily"`
	CommentTotal     []int32  `json:"comment_total"`
	EngagementDaily  []int32  `json:"engagement_daily"`
	EngagementTotal  []int32  `json:"engagement_total"`
	Buckets          []string `json:"buckets"`
	AggregationLevel string   `json:"aggregation_level,omitempty"`
}

// ViewsTrendResponse holds time-series view metrics split by subscriber and non-subscriber.
type ViewsTrendResponse struct {
	Status                  bool     `json:"status"`
	ShowData                int32    `json:"show_data"`
	SubscriberViewsDaily    []int32  `json:"subscriber_views_daily"`
	SubscriberViewsTotal    []int32  `json:"subscriber_views_total"`
	NonSubscriberViewsDaily []int32  `json:"non_subscriber_views_daily"`
	NonSubscriberViewsTotal []int32  `json:"non_subscriber_views_total"`
	VideoViewsDaily         []int32  `json:"video_views_daily"`
	VideoViewsTotal         []int32  `json:"video_views_total"`
	Buckets                 []string `json:"buckets"`
	AggregationLevel        string   `json:"aggregation_level,omitempty"`
}

// WatchTimeTrendResponse holds time-series watch time metrics.
type WatchTimeTrendResponse struct {
	Status                      bool      `json:"status"`
	ShowData                    int32     `json:"show_data"`
	SubscriberWatchTimeDaily    []int32   `json:"subscriber_watch_time_daily"`
	SubscriberWatchTimeTotal    []int32   `json:"subscriber_watch_time_total"`
	NonSubscriberWatchTimeDaily []int32   `json:"non_subscriber_watch_time_daily"`
	NonSubscriberWatchTimeTotal []int32   `json:"non_subscriber_watch_time_total"`
	AverageWatchTime            []float64 `json:"average_watch_time"`
	Buckets                     []string  `json:"buckets"`
	AggregationLevel            string    `json:"aggregation_level,omitempty"`
}

// FindVideoResponse holds traffic source breakdown data.
type FindVideoResponse struct {
	Status bool               `json:"status"`
	Data   []TrafficSourceItem `json:"data"`
}

// TrafficSourceItem holds a single traffic source entry.
type TrafficSourceItem struct {
	Name      string  `json:"name"`
	Value     int64   `json:"value"`
	PercValue float64 `json:"perc_value"`
}

// VideoSharingResponse holds video sharing platform breakdown data.
type VideoSharingResponse struct {
	Status bool          `json:"status"`
	Data   []SharingItem `json:"data"`
}

// SharingItem holds a single sharing platform entry.
type SharingItem struct {
	Name      string  `json:"name"`
	Value     int64   `json:"value"`
	PercValue float64 `json:"perc_value"`
}

// TopVideosResponse holds the top videos ordered by views and by engagement.
type TopVideosResponse struct {
	Status                      bool        `json:"status"`
	TopPostsOrderedByViews      []VideoItem `json:"top_posts_ordered_by_views"`
	TopPostsOrderedByEngagement []VideoItem `json:"top_posts_ordered_by_engagement"`
}

// LeastVideosResponse holds the least-performing videos ordered by views and by engagement.
type LeastVideosResponse struct {
	Status                         bool        `json:"status"`
	LeastPostsOrderedByViews       []VideoItem `json:"least_posts_ordered_by_views"`
	LeastPostsOrderedByEngagement  []VideoItem `json:"least_posts_ordered_by_engagement"`
}

// SortedTopVideosResponse holds videos sorted by a configurable metric.
type SortedTopVideosResponse struct {
	Status bool        `json:"status"`
	Data   []VideoItem `json:"top_posts"`
}

// VideoItem holds metrics for a single YouTube video.
// JSON field names use PHP-compatible singular forms for likes/dislikes/comments/shares.
type VideoItem struct {
	VideoID           string  `json:"video_id"`
	Title             string  `json:"title"`
	Description       string  `json:"description"`
	Duration          int64   `json:"duration"`
	ThumbnailURL      string  `json:"thumbnail_url"`
	MediaType         string  `json:"media_type"`
	IframeEmbedURL    string  `json:"iframe_embed_url"`
	ShareURL          string  `json:"share_url"`
	Engagement        int64   `json:"engagement"`
	Likes             int64   `json:"like"`
	Dislikes          int64   `json:"dislike"`
	Views             int64   `json:"views"`
	RedViews          int64   `json:"red_views"`
	Favorites         int64   `json:"favorites"`
	Comments          int64   `json:"comment"`
	SubscribersGained int64   `json:"subscribers_gained"`
	Shares            int64   `json:"share"`
	MinutesWatched    int64   `json:"minutes_watched"`
	RedMinutesWatched int64   `json:"red_minutes_watched"`
	AvgViewDuration   float64 `json:"average_view_duration"`
	AvgViewPercentage float64 `json:"average_view_percentage"`
	EngagementRate    float64 `json:"engagement_rate"`
	PublishedAt       string  `json:"published_at"`
}

// PerformanceScheduleResponse holds video performance data grouped by publish date.
type PerformanceScheduleResponse struct {
	Status     bool                       `json:"status"`
	Engagement *PerformanceEngagementData `json:"engagement"`
	VideoViews *PerformanceViewsData      `json:"video_views"`
}

// PerformanceEngagementData holds time-series engagement metrics by publish date.
type PerformanceEngagementData struct {
	ShowData   int32    `json:"show_data"`
	Buckets    []string `json:"buckets"`
	Count      []int32  `json:"count"`
	Likes      []int32  `json:"likes"`
	Dislikes   []int32  `json:"dislikes"`
	Shares     []int32  `json:"shares"`
	Comments   []int32  `json:"comments"`
	Engagement []int32  `json:"engagement"`
}

// PerformanceViewsData holds time-series view metrics by publish date.
type PerformanceViewsData struct {
	ShowData           int32    `json:"show_data"`
	Buckets            []string `json:"buckets"`
	Count              []int32  `json:"count"`
	SubscriberViews    []int32  `json:"subscriber_views"`
	NonSubscriberViews []int32  `json:"non_subscriber_views"`
}

// --- Request helpers ---

const maxLimit = 100

// Validate checks all required fields and validates date formats and timezone.
func (r *YoutubeRequest) Validate() error {
	if r.YoutubeID == "" {
		return httputil.NewValidationError("youtube_id is required")
	}
	if r.StartDate == "" {
		return httputil.NewValidationError("start_date is required")
	}
	if r.EndDate == "" {
		return httputil.NewValidationError("end_date is required")
	}
	startDate, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return httputil.NewValidationError("start_date must be in YYYY-MM-DD format")
	}
	endDate, err := time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return httputil.NewValidationError("end_date must be in YYYY-MM-DD format")
	}
	if endDate.Before(startDate) {
		return httputil.NewValidationError("end_date cannot be before start_date")
	}
	if r.Timezone != "" {
		if _, err := time.LoadLocation(r.Timezone); err != nil {
			return httputil.NewValidationError("invalid timezone: " + r.Timezone)
		}
	}
	return nil
}

// GetTimezone returns the requested timezone, defaulting to UTC.
func (r *YoutubeRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

// ToQueryParams converts the request into ClickHouse query parameters.
func (r *YoutubeRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	params, err := clickhouse.ParseStartEndDate(r.StartDate, r.EndDate, r.GetTimezone())
	if err != nil {
		return nil, err
	}
	params.AccountIDs = []string{r.YoutubeID}
	return params, nil
}

// validTopVideoOrderBy is the whitelist of allowed ORDER BY values for GetSortedTopVideos.
var validTopVideoOrderBy = map[string]bool{
	"views":           true,
	"likes":           true,
	"dislikes":        true,
	"engagement":      true,
	"comments":        true,
	"shares":          true,
	"engagement_rate": true,
	"minutes_watched": true,
	"published_at":    true,
}

// GetOrderBy returns a validated ORDER BY column. Defaults to "views" if empty or invalid.
func (r *TopVideosRequest) GetOrderBy() string {
	if r.OrderBy == "" || !validTopVideoOrderBy[r.OrderBy] {
		return "views"
	}
	return r.OrderBy
}

// GetLimit returns the video limit, clamped between 1 and maxLimit.
func (r *TopVideosRequest) GetLimit() int {
	if r.Limit <= 0 {
		return 15
	}
	if r.Limit > maxLimit {
		return maxLimit
	}
	return r.Limit
}
