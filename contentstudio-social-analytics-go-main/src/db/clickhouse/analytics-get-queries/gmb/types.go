package gmb

import "time"

// SummaryResult holds aggregated summary metrics for a GMB location over a date range.
type SummaryResult struct {
	TotalImpressions  int64   `ch:"total_impressions"`
	SearchImpressions int64   `ch:"search_impressions"`
	MapsImpressions   int64   `ch:"maps_impressions"`
	WebsiteClicks     int64   `ch:"website_clicks"`
	CallClicks        int64   `ch:"call_clicks"`
	DirectionRequests int64   `ch:"direction_requests"`
	OtherActions      int64   `ch:"other_actions"`
	TotalReviews      int64   `ch:"total_reviews"`
	AverageRating     float64 `ch:"average_rating"`
	TotalPosts        int64   `ch:"total_posts"`
}

// ImpressionsResult holds time-series impression data split by platform/device.
type ImpressionsResult struct {
	DesktopMapsDaily     []int64     `ch:"desktop_maps_daily"`
	DesktopSearchDaily   []int64     `ch:"desktop_search_daily"`
	MobileMapsDaily      []int64     `ch:"mobile_maps_daily"`
	MobileSearchDaily    []int64     `ch:"mobile_search_daily"`
	TotalImpressionsDaily []int64    `ch:"total_impressions_daily"`
	ShowData             int64       `ch:"show_data"`
	Buckets              []time.Time `ch:"buckets"`
}

// ImpressionsRollupResult holds aggregated impression totals for period comparison.
type ImpressionsRollupResult struct {
	TotalImpressions int64   `ch:"total_impressions"`
	DesktopMaps      int64   `ch:"desktop_maps"`
	DesktopSearch    int64   `ch:"desktop_search"`
	MobileMaps       int64   `ch:"mobile_maps"`
	MobileSearch     int64   `ch:"mobile_search"`
	AvgImpressions   float64 `ch:"avg_impressions"`
}

// ActionsResult holds time-series customer action data.
type ActionsResult struct {
	CallClicksDaily        []int64     `ch:"call_clicks_daily"`
	WebsiteClicksDaily     []int64     `ch:"website_clicks_daily"`
	DirectionRequestsDaily []int64     `ch:"direction_requests_daily"`
	OtherActionsDaily      []int64     `ch:"other_actions_daily"`
	ShowData               int64       `ch:"show_data"`
	Buckets                []time.Time `ch:"buckets"`
}

// ActionsRollupResult holds aggregated action totals for period comparison.
type ActionsRollupResult struct {
	TotalCallClicks        int64   `ch:"total_call_clicks"`
	TotalWebsiteClicks     int64   `ch:"total_website_clicks"`
	TotalDirectionRequests int64   `ch:"total_direction_requests"`
	TotalOtherActions      int64   `ch:"total_other_actions"`
	AvgActions             float64 `ch:"avg_actions"`
}

// SearchKeywordRow holds a single keyword result from search keywords query.
type SearchKeywordRow struct {
	Keyword              string    `ch:"keyword"`
	ImpressionsValue     int64     `ch:"impressions_value"`
	ImpressionsThreshold int64     `ch:"impressions_threshold"`
	KeywordMonth         time.Time `ch:"keyword_month"`
}

// TopPostRow holds a single post result from the top posts query.
type TopPostRow struct {
	PostName        string    `ch:"post_name"`
	Summary         string    `ch:"summary"`
	State           string    `ch:"state"`
	TopicType       string    `ch:"topic_type"`
	SearchURL       string    `ch:"search_url"`
	MediaNames      []string  `ch:"media_names"`
	MediaFormats    []string  `ch:"media_formats"`
	MediaGoogleURLs []string  `ch:"media_google_urls"`
	CreatedAt       time.Time `ch:"created_at"`
}

// PublishingResult holds time-series post count data.
type PublishingResult struct {
	PostCount []int64     `ch:"post_count"`
	Buckets   []time.Time `ch:"buckets"`
}

// TopicTypeRow holds a topic type with its post count.
type TopicTypeRow struct {
	TopicType string `ch:"topic_type"`
	Count     int64  `ch:"count"`
}

// ReviewsSummaryResult holds aggregated review stats.
type ReviewsSummaryResult struct {
	AvgRating    float64 `ch:"avg_rating"`
	TotalReviews int64   `ch:"total_reviews"`
	Star1        int64   `ch:"star_1"`
	Star2        int64   `ch:"star_2"`
	Star3        int64   `ch:"star_3"`
	Star4        int64   `ch:"star_4"`
	Star5        int64   `ch:"star_5"`
}

// ReviewsTimeSeriesResult holds daily review counts.
type ReviewsTimeSeriesResult struct {
	DailyReviews []int64     `ch:"daily_reviews"`
	Buckets      []time.Time `ch:"buckets"`
}

// ReviewRow holds a single review item.
type ReviewRow struct {
	ReviewID                string    `ch:"review_id"`
	ReviewerDisplayName     string    `ch:"reviewer_display_name"`
	ReviewerProfilePhotoURL string    `ch:"reviewer_profile_photo_url"`
	StarRating              int64     `ch:"star_rating"`
	Comment                 string    `ch:"comment"`
	ReplyComment            string    `ch:"reply_comment"`
	CreatedAt               time.Time `ch:"created_at"`
}

// ReviewsRollupResult holds aggregated review totals for period comparison.
type ReviewsRollupResult struct {
	TotalReviews int64   `ch:"total_reviews"`
	AvgRating    float64 `ch:"avg_rating"`
}

// MediaResult holds time-series media activity data.
type MediaResult struct {
	PhotoCountDaily []int64     `ch:"photo_count_daily"`
	VideoCountDaily []int64     `ch:"video_count_daily"`
	ShowData        int64       `ch:"show_data"`
	Buckets         []time.Time `ch:"buckets"`
}

// MediaRollupResult holds aggregated media totals for period comparison.
type MediaRollupResult struct {
	TotalPhotos int64   `ch:"total_photos"`
	TotalVideos int64   `ch:"total_videos"`
	AvgMedia    float64 `ch:"avg_media"`
}
