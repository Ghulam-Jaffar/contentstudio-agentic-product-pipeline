// Package youtube provides the ClickHouse repository layer for YouTube analytics.
// It contains query result types and repository methods that execute ClickHouse SQL queries
// migrated from the PHP Laravel YouTubeAnalyticsBuilder (contentstudio-backend).
//
// Tables queried: youtube_videos, youtube_activity_insights, youtube_channels,
// youtube_traffic_insights, youtube_shared_insights
package youtube

import "time"

// ActivitySummaryResult holds aggregated metrics from youtube_activity_insights.
type ActivitySummaryResult struct {
	WatchTime       int64   `ch:"watch_time"`
	AvgViewDuration float64 `ch:"avg_view_duration"`
	Likes           int64   `ch:"likes"`
	Dislikes        int64   `ch:"dislikes"`
	Comments        int64   `ch:"comments"`
	Shares          int64   `ch:"shares"`
	Engagement      int64   `ch:"engagement"`
	Views           int64   `ch:"views"`
}

// SubscriberSummaryResult holds the latest subscriber count from youtube_channels.
type SubscriberSummaryResult struct {
	Subscribers int64 `ch:"subscribers"`
}

// VideoCountResult holds the distinct video count from youtube_videos.
type VideoCountResult struct {
	VideoCount int64 `ch:"video_count"`
}

// SubscriberTrendResult holds time-series subscriber data with daily deltas.
type SubscriberTrendResult struct {
	ShowData               uint8       `ch:"show_data"`
	SubscribersGainedDaily []int32     `ch:"subscribers_gained_daily"`
	SubscribersTotal       []int32     `ch:"subscribers_total"`
	Buckets                []time.Time `ch:"buckets"`
}

// EngagementTrendResult holds time-series engagement metrics with daily and cumulative totals.
type EngagementTrendResult struct {
	ShowData        uint8       `ch:"show_data"`
	LikeDaily       []int32     `ch:"like_daily"`
	LikeTotal       []int32     `ch:"like_total"`
	DislikeDaily    []int32     `ch:"dislike_daily"`
	DislikeTotal    []int32     `ch:"dislike_total"`
	ShareDaily      []int32     `ch:"share_daily"`
	ShareTotal      []int32     `ch:"share_total"`
	CommentDaily    []int32     `ch:"comment_daily"`
	CommentTotal    []int32     `ch:"comment_total"`
	EngagementDaily []int32     `ch:"engagement_daily"`
	EngagementTotal []int32     `ch:"engagement_total"`
	Buckets         []time.Time `ch:"buckets"`
}

// ViewsTrendResult holds time-series view metrics split by subscriber and non-subscriber sources.
type ViewsTrendResult struct {
	ShowData                uint8       `ch:"show_data"`
	SubscriberViewsDaily    []int32     `ch:"subscriber_views_daily"`
	SubscriberViewsTotal    []int32     `ch:"subscriber_views_total"`
	NonSubscriberViewsDaily []int32     `ch:"non_subscriber_views_daily"`
	NonSubscriberViewsTotal []int32     `ch:"non_subscriber_views_total"`
	VideoViewsDaily         []int32     `ch:"video_views_daily"`
	VideoViewsTotal         []int32     `ch:"video_views_total"`
	Buckets                 []time.Time `ch:"buckets"`
}

// WatchTimeTrendResult holds time-series watch time metrics split by subscriber type.
type WatchTimeTrendResult struct {
	ShowData                    uint8       `ch:"show_data"`
	SubscriberWatchTimeDaily    []int32     `ch:"subscriber_watch_time_daily"`
	SubscriberWatchTimeTotal    []int32     `ch:"subscriber_watch_time_total"`
	NonSubscriberWatchTimeDaily []int32     `ch:"non_subscriber_watch_time_daily"`
	NonSubscriberWatchTimeTotal []int32     `ch:"non_subscriber_watch_time_total"`
	AverageWatchTime            []float64   `ch:"average_watch_time"`
	Buckets                     []time.Time `ch:"buckets"`
}

// TrafficSourceRow holds a single traffic source name, its view count, and its share percentage.
type TrafficSourceRow struct {
	Name      string  `ch:"name"`
	Value     int64   `ch:"value"`
	PercValue float64 `ch:"perc_value"`
}

// SharingRow holds a single sharing platform name, its share count, and its share percentage.
type SharingRow struct {
	Name      string  `ch:"name"`
	Value     int64   `ch:"value"`
	PercValue float64 `ch:"perc_value"`
}

// VideoRow holds the metrics for a single video from youtube_videos.
type VideoRow struct {
	VideoID           string    `ch:"video_id"`
	Title             string    `ch:"title"`
	Description       string    `ch:"description"`
	Duration          int64     `ch:"duration"`
	ThumbnailURL      string    `ch:"thumbnail_url"`
	MediaType         string    `ch:"media_type"`
	IframeEmbedURL    string    `ch:"iframe_embed_url"`
	ShareURL          string    `ch:"share_url"`
	Engagement        int64     `ch:"engagement"`
	Likes             int64     `ch:"likes"`
	Dislikes          int64     `ch:"dislikes"`
	Views             int64     `ch:"views"`
	RedViews          int64     `ch:"red_views"`
	Favorites         int64     `ch:"favorites"`
	Comments          int64     `ch:"comments"`
	SubscribersGained int64     `ch:"subscribers_gained"`
	Shares            int64     `ch:"shares"`
	MinutesWatched    int64     `ch:"minutes_watched"`
	RedMinutesWatched int64     `ch:"red_minutes_watched"`
	AvgViewDuration   float64   `ch:"avg_view_duration"`
	AvgViewPercentage float64   `ch:"avg_view_percentage"`
	EngagementRate    float64   `ch:"engagement_rate"`
	PublishedAt       time.Time `ch:"published_at"`
}

// PerformanceEngagementResult holds time-series video engagement data grouped by publish date.
type PerformanceEngagementResult struct {
	ShowData   uint8       `ch:"show_data"`
	Buckets    []time.Time `ch:"buckets"`
	Count      []int32     `ch:"count"`
	Likes      []int32     `ch:"likes"`
	Dislikes   []int32     `ch:"dislikes"`
	Shares     []int32     `ch:"shares"`
	Comments   []int32     `ch:"comments"`
	Engagement []int32     `ch:"engagement"`
}

// PerformanceViewsResult holds time-series video view data grouped by publish date.
type PerformanceViewsResult struct {
	ShowData          uint8       `ch:"show_data"`
	Buckets           []time.Time `ch:"buckets"`
	Count             []int32     `ch:"count"`
	SubscriberViews   []int32     `ch:"subscriber_views"`
	NonSubscriberViews []int32    `ch:"non_subscriber_views"`
}

// LatestSubscriberResult holds the most recent subscriber count from youtube_channels.
type LatestSubscriberResult struct {
	SubscriberCount int32 `ch:"subscriber_count"`
}

// SharedInsightsRow holds the raw row from youtube_shared_insights with one column per platform.
type SharedInsightsRow struct {
	Ameba           int64 `ch:"ameba"`
	Blogger         int64 `ch:"blogger"`
	CopyPaste       int64 `ch:"copy_paste"`
	Cyworld         int64 `ch:"cyworld"`
	Digg            int64 `ch:"digg"`
	Dropbox         int64 `ch:"dropbox"`
	Embed           int64 `ch:"embed"`
	Mail            int64 `ch:"mail"`
	Whatsapp        int64 `ch:"whats_app"`
	Other           int64 `ch:"other"`
	FacebookMsgr    int64 `ch:"facebook_messenger"`
	FacebookPages   int64 `ch:"facebook_pages"`
	Facebook        int64 `ch:"facebook"`
	Fotka           int64 `ch:"fotka"`
	Vkontakte       int64 `ch:"vkontakte"`
	GooglePlus      int64 `ch:"google_plus"`
	Discord         int64 `ch:"discord"`
	Linkedin        int64 `ch:"linkedin"`
	Goo             int64 `ch:"goo"`
	Hangouts        int64 `ch:"hangouts"`
	Pinterest       int64 `ch:"pinterest"`
	Myspace         int64 `ch:"myspace"`
	Reddit          int64 `ch:"reddit"`
	Skype           int64 `ch:"skype"`
	Telegram        int64 `ch:"telegram"`
	Tumblr          int64 `ch:"tumblr"`
	Twitter         int64 `ch:"twitter"`
	Viber           int64 `ch:"viber"`
	Weibo           int64 `ch:"weibo"`
	Wechat          int64 `ch:"wechat"`
	Youtube         int64 `ch:"youtube"`
}
