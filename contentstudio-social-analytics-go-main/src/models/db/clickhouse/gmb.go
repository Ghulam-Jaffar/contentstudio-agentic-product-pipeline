package clickhouse

import "time"

// GMBDailyMetrics represents the ClickHouse table schema for GMB daily performance metrics.
type GMBDailyMetrics struct {
	GmbID                            string    `ch:"gmb_id" json:"gmb_id"`
	AccountID                        string    `ch:"account_id" json:"account_id"`
	LocationID                       string    `ch:"location_id" json:"location_id"`
	AccountName                      string    `ch:"account_name" json:"account_name"`
	LocationName                     string    `ch:"location_name" json:"location_name"`
	PlatformName                     string    `ch:"platform_name" json:"platform_name"`
	InsertedAt                       time.Time `ch:"inserted_at" json:"inserted_at"`
	CreatedAt                        time.Time `ch:"created_at" json:"created_at"`
	BusinessImpressionsDesktopMaps   uint64    `ch:"business_impressions_desktop_maps" json:"business_impressions_desktop_maps"`
	BusinessImpressionsDesktopSearch uint64    `ch:"business_impressions_desktop_search" json:"business_impressions_desktop_search"`
	BusinessImpressionsMobileMaps    uint64    `ch:"business_impressions_mobile_maps" json:"business_impressions_mobile_maps"`
	BusinessImpressionsMobileSearch  uint64    `ch:"business_impressions_mobile_search" json:"business_impressions_mobile_search"`
	CallClicks                       uint64    `ch:"call_clicks" json:"call_clicks"`
	WebsiteClicks                    uint64    `ch:"website_clicks" json:"website_clicks"`
	BusinessDirectionRequests        uint64    `ch:"business_direction_requests" json:"business_direction_requests"`
	BusinessConversations            uint64    `ch:"business_conversations" json:"business_conversations"`
	BusinessBookings                 uint64    `ch:"business_bookings" json:"business_bookings"`
	BusinessFoodOrders               uint64    `ch:"business_food_orders" json:"business_food_orders"`
	BusinessFoodMenuClicks           uint64    `ch:"business_food_menu_clicks" json:"business_food_menu_clicks"`
}

// GMBMediaAssets represents the ClickHouse table schema for GMB media assets.
type GMBMediaAssets struct {
	GmbID                       string    `ch:"gmb_id" json:"gmb_id"`
	AccountID                   string    `ch:"account_id" json:"account_id"`
	LocationID                  string    `ch:"location_id" json:"location_id"`
	AccountName                 string    `ch:"account_name" json:"account_name"`
	LocationName                string    `ch:"location_name" json:"location_name"`
	PlatformName                string    `ch:"platform_name" json:"platform_name"`
	LanguageCode                string    `ch:"language_code" json:"language_code"`
	InsertedAt                  time.Time `ch:"inserted_at" json:"inserted_at"`
	CreatedAt                   time.Time `ch:"created_at" json:"created_at"`
	MediaName                   string    `ch:"media_name" json:"media_name"`
	SourceURL                   string    `ch:"source_url" json:"source_url"`
	MediaFormat                 string    `ch:"media_format" json:"media_format"`
	LocationAssociationCategory string    `ch:"location_association_category" json:"location_association_category"`
	GoogleURL                   string    `ch:"google_url" json:"google_url"`
	ThumbnailURL                string    `ch:"thumbnail_url" json:"thumbnail_url"`
	WidthPixels                 uint64    `ch:"width_pixels" json:"width_pixels"`
	HeightPixels                uint64    `ch:"height_pixels" json:"height_pixels"`
}

// GMBSearchKeywordsMonthly represents the ClickHouse table schema for GMB monthly search keywords.
type GMBSearchKeywordsMonthly struct {
	GmbID                string    `ch:"gmb_id" json:"gmb_id"`
	AccountID            string    `ch:"account_id" json:"account_id"`
	LocationID           string    `ch:"location_id" json:"location_id"`
	AccountName          string    `ch:"account_name" json:"account_name"`
	LocationName         string    `ch:"location_name" json:"location_name"`
	PlatformName         string    `ch:"platform_name" json:"platform_name"`
	InsertedAt           time.Time `ch:"inserted_at" json:"inserted_at"`
	KeywordMonth         time.Time `ch:"keyword_month" json:"keyword_month"`
	Keyword              string    `ch:"keyword" json:"keyword"`
	ImpressionsValue     uint64    `ch:"impressions_value" json:"impressions_value"`
	ImpressionsThreshold uint64    `ch:"impressions_threshold" json:"impressions_threshold"`
}

// GMBLocalPosts represents the ClickHouse table schema for GMB local posts.
type GMBLocalPosts struct {
	GmbID           string    `ch:"gmb_id" json:"gmb_id"`
	AccountID       string    `ch:"account_id" json:"account_id"`
	LocationID      string    `ch:"location_id" json:"location_id"`
	AccountName     string    `ch:"account_name" json:"account_name"`
	LocationName    string    `ch:"location_name" json:"location_name"`
	PlatformName    string    `ch:"platform_name" json:"platform_name"`
	LanguageCode    string    `ch:"language_code" json:"language_code"`
	InsertedAt      time.Time `ch:"inserted_at" json:"inserted_at"`
	CreatedAt       time.Time `ch:"created_at" json:"created_at"`
	UpdatedAt       time.Time `ch:"updated_at" json:"updated_at"`
	PostName        string    `ch:"post_name" json:"post_name"`
	Summary         string    `ch:"summary" json:"summary"`
	State           string    `ch:"state" json:"state"`
	TopicType       string    `ch:"topic_type" json:"topic_type"`
	SearchURL       string    `ch:"search_url" json:"search_url"`
	MediaNames      []string  `ch:"media_names" json:"media_names"`
	MediaFormats    []string  `ch:"media_formats" json:"media_formats"`
	MediaGoogleURLs []string  `ch:"media_google_urls" json:"media_google_urls"`
}

// GMBReviews represents the ClickHouse table schema for GMB reviews.
type GMBReviews struct {
	GmbID                   string    `ch:"gmb_id" json:"gmb_id"`
	AccountID               string    `ch:"account_id" json:"account_id"`
	LocationID              string    `ch:"location_id" json:"location_id"`
	AccountName             string    `ch:"account_name" json:"account_name"`
	LocationName            string    `ch:"location_name" json:"location_name"`
	PlatformName            string    `ch:"platform_name" json:"platform_name"`
	InsertedAt              time.Time `ch:"inserted_at" json:"inserted_at"`
	CreatedAt               time.Time `ch:"created_at" json:"created_at"`
	UpdatedAt               time.Time `ch:"updated_at" json:"updated_at"`
	ReviewID                string    `ch:"review_id" json:"review_id"`
	ReviewName              string    `ch:"review_name" json:"review_name"`
	ReviewerDisplayName     string    `ch:"reviewer_display_name" json:"reviewer_display_name"`
	ReviewerProfilePhotoURL string    `ch:"reviewer_profile_photo_url" json:"reviewer_profile_photo_url"`
	StarRating      uint64    `ch:"star_rating" json:"star_rating"`
	Comment         string    `ch:"comment" json:"comment"`
	ReplyComment    string    `ch:"reply_comment" json:"reply_comment"`
	ReplyUpdateTime time.Time `ch:"reply_update_time" json:"reply_update_time"`
}
