package kafka

import (
	"time"
)

// Pinterest Work Order Models

type PinterestAccountWorkOrder struct {
	ID          string `json:"id"`
	AccountID   string `json:"account_id"`
	AccessToken string `json:"access_token"`
	AccountType string `json:"account_type"` // "board" or "profile"
	BoardID     string `json:"board_id,omitempty"`
	WorkspaceID string `json:"workspace_id"`
	SyncType    string `json:"sync_type"` // "incremental" | "full_sync"
	StartDate   string `json:"start_date,omitempty"`
	EndDate     string `json:"end_date,omitempty"`
}

type PinterestBatchWorkOrder struct {
	BatchID   string                      `json:"batch_id"`
	SyncType  string                      `json:"sync_type"`
	Accounts  []PinterestAccountWorkOrder `json:"accounts"`
	CreatedAt time.Time                   `json:"created_at"`
}

// Pinterest Raw Data Models (from API responses)

type RawPinterestUser struct {
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	About          string    `json:"about"`
	ProfileImage   string    `json:"profile_image"`
	WebsiteURL     string    `json:"website_url"`
	BusinessName   string    `json:"business_name"`
	BoardCount     int       `json:"board_count"`
	PinCount       int       `json:"pin_count"`
	AccountType    string    `json:"account_type"`
	FollowerCount  int64     `json:"follower_count"`
	FollowingCount int64     `json:"following_count"`
	MonthlyViews   int64     `json:"monthly_views"`
	WorkspaceID    string    `json:"workspace_id"`
	SavingTime     time.Time `json:"saving_time"`
}

type RawPinterestBoard struct {
	BoardID           string    `json:"board_id"`
	UserID            string    `json:"user_id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Privacy           string    `json:"privacy"`
	PinCount          int       `json:"pin_count"`
	FollowerCount     int       `json:"follower_count"`
	CollaboratorCount int       `json:"collaborator_count"`
	CreatedAt         time.Time `json:"created_at"`
	Owner             string    `json:"owner"`
	ImageCoverURL     string    `json:"image_cover_url"`
	PinThumbnailURLs  []string  `json:"pin_thumbnail_urls"`
	WorkspaceID       string    `json:"workspace_id"`
	SavingTime        time.Time `json:"saving_time"`
}

type RawPinterestPin struct {
	PinID           string    `json:"pin_id"`
	UserID          string    `json:"user_id"`
	BoardID         string    `json:"board_id"`
	BoardSectionID  string    `json:"board_section_id"`
	ParentPinID     string    `json:"parent_pin_id"`
	Title           string    `json:"title"`
	Note            string    `json:"note"`
	Description     string    `json:"description"`
	Link            string    `json:"link"`
	DominantColor   string    `json:"dominant_color"`
	CreativeType    string    `json:"creative_type"`
	MediaType       string    `json:"media_type"`
	CoverImageURL   string    `json:"cover_image_url"`
	VideoURL        string    `json:"video_url"`
	Duration        string    `json:"duration"`
	Height          string    `json:"height"`
	Width           string    `json:"width"`
	IsStandard      bool      `json:"is_standard"`
	IsOwner         bool      `json:"is_owner"`
	HasBeenPromoted bool      `json:"has_been_promoted"`
	BoardOwner      string    `json:"board_owner"`
	ProductTags     []string  `json:"product_tags"`
	CreatedAt       time.Time `json:"created_at"`
	DayOfWeek       string    `json:"day_of_week"`
	HourOfDay       int       `json:"hour_of_day"`
	WorkspaceID     string    `json:"workspace_id"`
	SavingTime      time.Time `json:"saving_time"`
}

type RawPinterestPinInsight struct {
	PinID              string    `json:"pin_id"`
	UserID             string    `json:"user_id"`
	BoardID            string    `json:"board_id"`
	Date               time.Time `json:"date"`
	DataStatus         string    `json:"data_status"`
	Impression         int64     `json:"impression"`
	PinClicks          int64     `json:"pin_clicks"`
	OutboundClicks     int64     `json:"outbound_clicks"`
	Saves              int64     `json:"saves"`
	SaveRate           float64   `json:"save_rate"`
	Clickthrough       int64     `json:"clickthrough"`
	ClickthroughRate   float64   `json:"clickthrough_rate"`
	Engagement         int64     `json:"engagement"`
	EngagementRate     float64   `json:"engagement_rate"`
	VideoMRCView       int64     `json:"video_mrc_view"`
	VideoStart         int64     `json:"video_start"`
	Video10sView       int64     `json:"video_10s_view"`
	VideoAvgWatchTime  int64     `json:"video_avg_watch_time"`
	VideoV50WatchTime  int64     `json:"video_v50_watch_time"`
	FullScreenPlay     int64     `json:"full_screen_play"`
	FullScreenPlaytime int64     `json:"full_screen_playtime"`
	ProfileVisit       int64     `json:"profile_visit"`
	Closeup            int64     `json:"closeup"`
	Quartile95sPercent int64     `json:"quartile_95s_percent_view"`
	UserFollow         int64     `json:"user_follow"`
	WorkspaceID        string    `json:"workspace_id"`
	SavingTime         time.Time `json:"saving_time"`
}

type RawPinterestUserInsight struct {
	UserID             string    `json:"user_id"`
	Date               time.Time `json:"date"`
	DataStatus         string    `json:"data_status"`
	Impression         int64     `json:"impression"`
	PinClicks          int64     `json:"pin_clicks"`
	PinClickRate       float64   `json:"pin_click_rate"`
	OutboundClicks     int64     `json:"outbound_clicks"`
	Saves              int64     `json:"saves"`
	SaveRate           float64   `json:"save_rate"`
	Clickthrough       int64     `json:"clickthrough"`
	ClickthroughRate   float64   `json:"clickthrough_rate"`
	Engagement         int64     `json:"engagement"`
	EngagementRate     float64   `json:"engagement_rate"`
	VideoMRCView       int64     `json:"video_mrc_view"`
	VideoStart         int64     `json:"video_start"`
	Video10sView       int64     `json:"video_10s_view"`
	VideoAvgWatchTime  int64     `json:"video_avg_watch_time"`
	VideoV50WatchTime  int64     `json:"video_v50_watch_time"`
	FullScreenPlay     int64     `json:"full_screen_play"`
	FullScreenPlaytime int64     `json:"full_screen_playtime"`
	ProfileVisit       int64     `json:"profile_visit"`
	Closeup            int64     `json:"closeup"`
	Quartile95sPercent int64     `json:"quartile_95s_percent_view"`
	WorkspaceID        string    `json:"workspace_id"`
	SavingTime         time.Time `json:"saving_time"`
}

// Pinterest Parsed Data Models (analytics-ready for ClickHouse)

type ParsedPinterestUser struct {
	RecordID       string    `json:"record_id"` // MD5(user_id + date)
	UserID         string    `json:"user_id"`
	Username       string    `json:"username"`
	About          string    `json:"about"`
	ProfileImage   string    `json:"profile_image"`
	WebsiteURL     string    `json:"website_url"`
	BusinessName   string    `json:"business_name"`
	BoardCount     int       `json:"board_count"`
	PinCount       int       `json:"pin_count"`
	AccountType    string    `json:"account_type"`
	FollowerCount  int64     `json:"follower_count"`
	FollowingCount int64     `json:"following_count"`
	MonthlyViews   int64     `json:"monthly_views"`
	InsertedAt     time.Time `json:"inserted_at"`
}

type ParsedPinterestBoard struct {
	RecordID          string    `json:"record_id"` // MD5(board_id + date)
	BoardID           string    `json:"board_id"`
	UserID            string    `json:"user_id"`
	Name              string    `json:"name"`
	Description       string    `json:"description"`
	Privacy           string    `json:"privacy"`
	PinCount          string    `json:"pin_count"`
	FollowerCount     string    `json:"follower_count"`
	CollaboratorCount string    `json:"collaborator_count"`
	Owner             string    `json:"owner"`
	ImageCoverURL     string    `json:"image_cover_url"`
	PinThumbnailURLs  []string  `json:"pin_thumbnail_urls"`
	CreatedAt         time.Time `json:"created_at"`
	InsertedAt        time.Time `json:"inserted_at"`
}

type ParsedPinterestPin struct {
	RecordID        string    `json:"record_id"` // MD5(pin_id + date)
	PinID           string    `json:"pin_id"`
	UserID          string    `json:"user_id"`
	BoardID         string    `json:"board_id"`
	BoardSectionID  string    `json:"board_section_id"`
	ParentPinID     string    `json:"parent_pin_id"`
	Title           string    `json:"title"`
	Note            string    `json:"note"`
	Description     string    `json:"description"`
	Link            string    `json:"link"`
	DominantColor   string    `json:"dominant_color"`
	CreativeType    string    `json:"creative_type"`
	MediaType       string    `json:"media_type"`
	CoverImageURL   string    `json:"cover_image_url"`
	VideoURL        string    `json:"video_url"`
	Duration        string    `json:"duration"`
	Height          string    `json:"height"`
	Width           string    `json:"width"`
	IsStandard      string    `json:"is_standard"`
	IsOwner         string    `json:"is_owner"`
	HasBeenPromoted string    `json:"has_been_promoted"`
	BoardOwner      string    `json:"board_owner"`
	ProductTags     []string  `json:"product_tags"`
	CreatedAt       time.Time `json:"created_at"`
	DayOfWeek       string    `json:"day_of_week"`
	HourOfDay       int       `json:"hour_of_day"`
	InsertedAt      time.Time `json:"inserted_at"`
}

type ParsedPinterestPinInsight struct {
	RecordID           string    `json:"record_id"` // MD5(pin_id + date)
	PinID              string    `json:"pin_id"`
	UserID             string    `json:"user_id"`
	BoardID            string    `json:"board_id"`
	Date               time.Time `json:"date"`
	DataStatus         string    `json:"data_status"`
	Impression         int64     `json:"impression"`
	PinClicks          int64     `json:"pin_clicks"`
	OutboundClicks     int64     `json:"outbound_clicks"`
	Saves              int64     `json:"saves"`
	SaveRate           float64   `json:"save_rate"`
	Clickthrough       int64     `json:"clickthrough"`
	ClickthroughRate   float64   `json:"clickthrough_rate"`
	Engagement         int64     `json:"engagement"`
	EngagementRate     float64   `json:"engagement_rate"`
	VideoMRCView       int64     `json:"video_mrc_view"`
	VideoStart         int64     `json:"video_start"`
	Video10sView       int64     `json:"video_10s_view"`
	VideoAvgWatchTime  int64     `json:"video_avg_watch_time"`
	VideoV50WatchTime  int64     `json:"video_v50_watch_time"`
	FullScreenPlay     int64     `json:"full_screen_play"`
	FullScreenPlaytime int64     `json:"full_screen_playtime"`
	ProfileVisit       int64     `json:"profile_visit"`
	Closeup            int64     `json:"closeup"`
	Quartile95sPercent int64     `json:"quartile_95s_percent_view"`
	UserFollow         int64     `json:"user_follow"`
	DayOfWeek          string    `json:"day_of_week"`
	HourOfDay          int       `json:"hour_of_day"`
	InsertedAt         time.Time `json:"inserted_at"`
}

type ParsedPinterestUserInsight struct {
	RecordID           string    `json:"record_id"` // MD5(user_id + date)
	UserID             string    `json:"user_id"`
	Date               time.Time `json:"date"`
	DataStatus         string    `json:"data_status"`
	Impression         int64     `json:"impression"`
	PinClicks          int64     `json:"pin_clicks"`
	PinClickRate       float64   `json:"pin_click_rate"`
	OutboundClicks     int64     `json:"outbound_clicks"`
	Saves              int64     `json:"saves"`
	SaveRate           float64   `json:"save_rate"`
	Clickthrough       int64     `json:"clickthrough"`
	ClickthroughRate   float64   `json:"clickthrough_rate"`
	Engagement         int64     `json:"engagement"`
	EngagementRate     float64   `json:"engagement_rate"`
	VideoMRCView       int64     `json:"video_mrc_view"`
	VideoStart         int64     `json:"video_start"`
	Video10sView       int64     `json:"video_10s_view"`
	VideoAvgWatchTime  int64     `json:"video_avg_watch_time"`
	VideoV50WatchTime  int64     `json:"video_v50_watch_time"`
	FullScreenPlay     int64     `json:"full_screen_play"`
	FullScreenPlaytime int64     `json:"full_screen_playtime"`
	ProfileVisit       int64     `json:"profile_visit"`
	Closeup            int64     `json:"closeup"`
	Quartile95sPercent int64     `json:"quartile_95s_percent_view"`
	InsertedAt         time.Time `json:"inserted_at"`
}

// Pinterest Account Type constants
const (
	PinterestAccountTypeBoard   = "board"
	PinterestAccountTypeProfile = "profile"
)

// Pinterest Sync Type constants
const (
	PinterestSyncTypeIncremental = "incremental"
	PinterestSyncTypeImmediate   = "immediate"
	PinterestSyncTypeFullSync    = "full_sync"
)

// Pinterest Data Status constants
const (
	PinterestDataStatusReady                 = "READY"
	PinterestDataStatusEstimate              = "ESTIMATE"
	PinterestDataStatusProcessing            = "PROCESSING"
	PinterestDataStatusBeforePinCreated      = "BEFORE_PIN_CREATED"
	PinterestDataStatusBeforeBusinessCreated = "BEFORE_BUSINESS_CREATED"
)

// PinterestKafkaTopics defines the Kafka topics for Pinterest data.
var PinterestKafkaTopics = struct {
	WorkOrder          string
	ImmediateWorkOrder string
	RawUsers           string
	RawBoards          string
	RawPins            string
	RawPinInsights     string
	RawUserInsights    string
	ParsedUsers        string
	ParsedBoards       string
	ParsedPins         string
	ParsedPinInsights  string
	ParsedUserInsights string
}{
	WorkOrder:          "work-order-pinterest",
	ImmediateWorkOrder: "immediate-work-order-pinterest",
	RawUsers:           "raw-pinterest-users",
	RawBoards:          "raw-pinterest-boards",
	RawPins:            "raw-pinterest-pins",
	RawPinInsights:     "raw-pinterest-pin-insights",
	RawUserInsights:    "raw-pinterest-user-insights",
	ParsedUsers:        "parsed-pinterest-users",
	ParsedBoards:       "parsed-pinterest-boards",
	ParsedPins:         "parsed-pinterest-pins",
	ParsedPinInsights:  "parsed-pinterest-pin-insights",
	ParsedUserInsights: "parsed-pinterest-user-insights",
}
