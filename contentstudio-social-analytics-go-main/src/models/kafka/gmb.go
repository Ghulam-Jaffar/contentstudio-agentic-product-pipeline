package kafka

import "time"

// GMBBatchWorkOrder represents a batch of GMB accounts dispatched by the unified account fetcher.
type GMBBatchWorkOrder struct {
	BatchID   string                `json:"batch_id"`
	SyncType  string                `json:"sync_type"`
	Accounts  []GMBAccountWorkOrder `json:"accounts"`
	CreatedAt time.Time             `json:"created_at"`
}

// GMBAccountWorkOrder represents a single GMB account to process.
type GMBAccountWorkOrder struct {
	ID           string `json:"id"`            // MongoDB _id (hex)
	WorkspaceID  string `json:"workspace_id"`  // Workspace identifier
	AccountID    string `json:"account_id"`    // GMB account ID
	LocationID   string `json:"location_id"`   // GMB location ID
	AccessToken  string `json:"access_token"`  // OAuth access token
	RefreshToken string `json:"refresh_token"` // OAuth refresh token
	AccountName  string `json:"account_name"`  // Display name of the account
	LocationName string `json:"location_name"` // Display name of the location
	LanguageCode string `json:"language_code"` // Language code (e.g., "en")
	SyncType     string `json:"sync_type"`     // "incremental" | "full_sync"
	StartDate    string `json:"start_date,omitempty"`
	EndDate      string `json:"end_date,omitempty"`
}

// ParsedGMBDailyMetrics represents parsed daily performance metrics from the GMB API.
type ParsedGMBDailyMetrics struct {
	AccountID                        string `json:"account_id"`
	LocationID                       string `json:"location_id"`
	AccountName                      string `json:"account_name"`
	LocationName                     string `json:"location_name"`
	Date                             string `json:"date"` // YYYY-MM-DD
	BusinessImpressionsDesktopMaps   uint64 `json:"business_impressions_desktop_maps"`
	BusinessImpressionsDesktopSearch uint64 `json:"business_impressions_desktop_search"`
	BusinessImpressionsMobileMaps    uint64 `json:"business_impressions_mobile_maps"`
	BusinessImpressionsMobileSearch  uint64 `json:"business_impressions_mobile_search"`
	CallClicks                       uint64 `json:"call_clicks"`
	WebsiteClicks                    uint64 `json:"website_clicks"`
	BusinessDirectionRequests        uint64 `json:"business_direction_requests"`
	BusinessConversations            uint64 `json:"business_conversations"`
	BusinessBookings                 uint64 `json:"business_bookings"`
	BusinessFoodOrders               uint64 `json:"business_food_orders"`
	BusinessFoodMenuClicks           uint64 `json:"business_food_menu_clicks"`
}

// ParsedGMBSearchKeyword represents a single parsed search keyword from GMB API.
type ParsedGMBSearchKeyword struct {
	AccountID            string `json:"account_id"`
	LocationID           string `json:"location_id"`
	AccountName          string `json:"account_name"`
	LocationName         string `json:"location_name"`
	KeywordMonth         string `json:"keyword_month"` // YYYY-MM
	Keyword              string `json:"keyword"`
	ImpressionsValue     uint64 `json:"impressions_value"`
	ImpressionsThreshold uint64 `json:"impressions_threshold"`
}

// ParsedGMBLocalPost represents a parsed local post from GMB API.
type ParsedGMBLocalPost struct {
	AccountID       string   `json:"account_id"`
	LocationID      string   `json:"location_id"`
	AccountName     string   `json:"account_name"`
	LocationName    string   `json:"location_name"`
	LanguageCode    string   `json:"language_code"`
	PostName        string   `json:"post_name"`
	Summary         string   `json:"summary"`
	State           string   `json:"state"`
	TopicType       string   `json:"topic_type"`
	SearchURL       string   `json:"search_url"`
	CreateTime      string   `json:"create_time"`
	UpdateTime      string   `json:"update_time"`
	MediaNames      []string `json:"media_names"`
	MediaFormats    []string `json:"media_formats"`
	MediaGoogleURLs []string `json:"media_google_urls"`
}

// ParsedGMBReview represents a parsed review from GMB API.
type ParsedGMBReview struct {
	AccountID               string `json:"account_id"`
	LocationID              string `json:"location_id"`
	AccountName             string `json:"account_name"`
	LocationName            string `json:"location_name"`
	ReviewID                string `json:"review_id"`
	ReviewName              string `json:"review_name"`
	ReviewerDisplayName     string `json:"reviewer_display_name"`
	ReviewerProfilePhotoURL string `json:"reviewer_profile_photo_url"`
	StarRating              string `json:"star_rating"` // e.g., "FOUR", "FIVE"
	Comment                 string `json:"comment"`
	CreateTime              string `json:"create_time"`
	UpdateTime              string `json:"update_time"`
	ReplyComment            string `json:"reply_comment"`
	ReplyUpdateTime         string `json:"reply_update_time"`
}

// ParsedGMBMediaAsset represents a parsed media asset from GMB API.
type ParsedGMBMediaAsset struct {
	AccountID                   string `json:"account_id"`
	LocationID                  string `json:"location_id"`
	AccountName                 string `json:"account_name"`
	LocationName                string `json:"location_name"`
	LanguageCode                string `json:"language_code"`
	MediaName                   string `json:"media_name"`
	SourceURL                   string `json:"source_url"`
	MediaFormat                 string `json:"media_format"`
	LocationAssociationCategory string `json:"location_association_category"`
	GoogleURL                   string `json:"google_url"`
	ThumbnailURL                string `json:"thumbnail_url"`
	WidthPixels                 uint64 `json:"width_pixels"`
	HeightPixels                uint64 `json:"height_pixels"`
	CreateTime                  string `json:"create_time"`
}

// RawGMBData wraps raw JSON data from GMB APIs for Kafka transport.
type RawGMBData struct {
	WorkspaceID  string      `json:"workspace_id"`
	AccountID    string      `json:"account_id"`
	LocationID   string      `json:"location_id"`
	AccountName  string      `json:"account_name"`
	LocationName string      `json:"location_name"`
	LanguageCode string      `json:"language_code"`
	DataType     string      `json:"data_type"`     // "performance_metrics" | "search_keywords" | "local_posts" | "reviews" | "media_assets"
	KeywordMonth string      `json:"keyword_month"` // YYYY-MM, only set for search_keywords
	Data         interface{} `json:"data"`
}
