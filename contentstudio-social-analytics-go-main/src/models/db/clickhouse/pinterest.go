package clickhouse

import "time"

// PinterestUser represents the ClickHouse table schema for Pinterest users.
type PinterestUser struct {
	RecordID       string    `ch:"record_id"` // MD5(user_id + date)
	UserID         string    `ch:"user_id"`
	Username       string    `ch:"username"`
	About          string    `ch:"about"`
	ProfileImage   string    `ch:"profile_image"`
	WebsiteURL     string    `ch:"website_url"`
	BusinessName   string    `ch:"business_name"`
	BoardCount     int       `ch:"board_count"`
	PinCount       int       `ch:"pin_count"`
	AccountType    string    `ch:"account_type"`
	FollowerCount  int64     `ch:"follower_count"`
	FollowingCount int64     `ch:"following_count"`
	MonthlyViews   int64     `ch:"monthly_views"`
	InsertedAt     time.Time `ch:"inserted_at"`
}

// PinterestBoard represents the ClickHouse table schema for Pinterest boards.
type PinterestBoard struct {
	RecordID          string    `ch:"record_id"` // MD5(board_id + date)
	BoardID           string    `ch:"board_id"`
	UserID            string    `ch:"user_id"`
	Name              string    `ch:"name"`
	Description       string    `ch:"description"`
	Privacy           string    `ch:"privacy"`
	PinCount          string    `ch:"pin_count"`
	FollowerCount     string    `ch:"follower_count"`
	CollaboratorCount string    `ch:"collaborator_count"`
	Owner             string    `ch:"owner"`
	ImageCoverURL     string    `ch:"image_cover_url"`
	PinThumbnailURLs  []string  `ch:"pin_thumbnail_urls"`
	CreatedAt         time.Time `ch:"created_at"`
	InsertedAt        time.Time `ch:"inserted_at"`
}

// PinterestPin represents the ClickHouse table schema for Pinterest pins.
type PinterestPin struct {
	RecordID        string    `ch:"record_id"` // MD5(pin_id + date)
	PinID           string    `ch:"pin_id"`
	UserID          string    `ch:"user_id"`
	BoardID         string    `ch:"board_id"`
	BoardSectionID  string    `ch:"board_section_id"`
	ParentPinID     string    `ch:"parent_pin_id"`
	Title           string    `ch:"title"`
	Note            string    `ch:"note"`
	Description     string    `ch:"description"`
	Link            string    `ch:"link"`
	DominantColor   string    `ch:"dominant_color"`
	CreativeType    string    `ch:"creative_type"`
	MediaType       string    `ch:"media_type"`
	CoverImageURL   string    `ch:"cover_image_url"`
	VideoURL        string    `ch:"video_url"`
	Duration        string    `ch:"duration"`
	Height          string    `ch:"height"`
	Width           string    `ch:"width"`
	IsStandard      string    `ch:"is_standard"`
	IsOwner         string    `ch:"is_owner"`
	HasBeenPromoted string    `ch:"has_been_promoted"`
	BoardOwner      string    `ch:"board_owner"`
	ProductTags     []string  `ch:"product_tags"`
	CreatedAt       time.Time `ch:"created_at"`
	DayOfWeek       string    `ch:"day_of_week"`
	HourOfDay       int       `ch:"hour_of_day"`
	InsertedAt      time.Time `ch:"inserted_at"`
}

// PinterestPinInsight represents the ClickHouse table schema for Pinterest pin insights.
type PinterestPinInsight struct {
	RecordID           string    `ch:"record_id"` // MD5(pin_id + date)
	PinID              string    `ch:"pin_id"`
	UserID             string    `ch:"user_id"`
	BoardID            string    `ch:"board_id"`
	Date               time.Time `ch:"date"`
	DataStatus         string    `ch:"data_status"`
	Impression         int64     `ch:"impression"`
	PinClicks          int64     `ch:"pin_clicks"`
	OutboundClicks     int64     `ch:"outbound_click"`
	Saves              int64     `ch:"saves"`
	SaveRate           float64   `ch:"save_rate"`
	Clickthrough       int64     `ch:"clickthrough"`
	ClickthroughRate   float64   `ch:"clickthrough_rate"`
	Engagement         int64     `ch:"engagement"`
	EngagementRate     float64   `ch:"engagement_rate"`
	VideoMRCView       int64     `ch:"video_mrc_view"`
	VideoStart         int64     `ch:"video_start"`
	Video10sView       int64     `ch:"video_10s_view"`
	VideoAvgWatchTime  int64     `ch:"video_avg_watch_time"`
	VideoV50WatchTime  int64     `ch:"video_v50_watch_time"`
	FullScreenPlay     int64     `ch:"full_screen_play"`
	FullScreenPlaytime int64     `ch:"full_screen_playtime"`
	ProfileVisit       int64     `ch:"profile_visit"`
	Closeup            int64     `ch:"closeup"`
	Quartile95sPercent int64     `ch:"quartile_95s_percent_view"`
	UserFollow         int64     `ch:"user_follow"`
	DayOfWeek          string    `ch:"day_of_week"`
	HourOfDay          int       `ch:"hour_of_day"`
	InsertedAt         time.Time `ch:"inserted_at"`
}

// PinterestUserInsight represents the ClickHouse table schema for Pinterest user insights.
type PinterestUserInsight struct {
	RecordID           string    `ch:"record_id"` // MD5(user_id + date)
	UserID             string    `ch:"user_id"`
	Date               time.Time `ch:"date"`
	DataStatus         string    `ch:"data_status"`
	Impression         int64     `ch:"impression"`
	PinClicks          int64     `ch:"pin_clicks"`
	PinClickRate       float64   `ch:"pin_click_rate"`
	OutboundClicks     int64     `ch:"outbound_click"`
	Saves              int64     `ch:"saves"`
	SaveRate           float64   `ch:"save_rate"`
	Clickthrough       int64     `ch:"clickthrough"`
	ClickthroughRate   float64   `ch:"clickthrough_rate"`
	Engagement         int64     `ch:"engagement"`
	EngagementRate     float64   `ch:"engagement_rate"`
	VideoMRCView       int64     `ch:"video_mrc_view"`
	VideoStart         int64     `ch:"video_start"`
	Video10sView       int64     `ch:"video_10s_view"`
	VideoAvgWatchTime  int64     `ch:"video_avg_watch_time"`
	VideoV50WatchTime  int64     `ch:"video_v50_watch_time"`
	FullScreenPlay     int64     `ch:"full_screen_play"`
	FullScreenPlaytime int64     `ch:"full_screen_playtime"`
	ProfileVisit       int64     `ch:"profile_visit"`
	Closeup            int64     `ch:"closeup"`
	Quartile95sPercent int64     `ch:"quartile_95s_percent_view"`
	InsertedAt         time.Time `ch:"inserted_at"`
}

// TableName returns the ClickHouse table name for PinterestUser.
func (PinterestUser) TableName() string {
	return "pinterest_users"
}

// TableName returns the ClickHouse table name for PinterestBoard.
func (PinterestBoard) TableName() string {
	return "pinterest_boards"
}

// TableName returns the ClickHouse table name for PinterestPin.
func (PinterestPin) TableName() string {
	return "pinterest_pins"
}

// TableName returns the ClickHouse table name for PinterestPinInsight.
func (PinterestPinInsight) TableName() string {
	return "pinterest_pin_insights"
}

// TableName returns the ClickHouse table name for PinterestUserInsight.
func (PinterestUserInsight) TableName() string {
	return "pinterest_user_insights"
}

// PinterestUserColumns returns the column names for batch insert.
func PinterestUserColumns() []string {
	return []string{
		"record_id",
		"user_id",
		"username",
		"about",
		"profile_image",
		"website_url",
		"business_name",
		"board_count",
		"pin_count",
		"account_type",
		"follower_count",
		"following_count",
		"monthly_views",
		"inserted_at",
	}
}

// PinterestBoardColumns returns the column names for batch insert.
func PinterestBoardColumns() []string {
	return []string{
		"record_id",
		"board_id",
		"user_id",
		"name",
		"description",
		"privacy",
		"pin_count",
		"follower_count",
		"collaborator_count",
		"owner",
		"image_cover_url",
		"pin_thumbnail_urls",
		"created_at",
		"inserted_at",
	}
}

// PinterestPinColumns returns the column names for batch insert.
func PinterestPinColumns() []string {
	return []string{
		"record_id",
		"pin_id",
		"user_id",
		"board_id",
		"board_section_id",
		"parent_pin_id",
		"title",
		"note",
		"description",
		"link",
		"dominant_color",
		"creative_type",
		"media_type",
		"cover_image_url",
		"video_url",
		"duration",
		"height",
		"width",
		"is_standard",
		"is_owner",
		"has_been_promoted",
		"board_owner",
		"product_tags",
		"created_at",
		"day_of_week",
		"hour_of_day",
		"inserted_at",
	}
}

// PinterestPinInsightColumns returns the column names for batch insert.
func PinterestPinInsightColumns() []string {
	return []string{
		"record_id",
		"pin_id",
		"user_id",
		"board_id",
		"date",
		"data_status",
		"impression",
		"pin_clicks",
		"outbound_click",
		"saves",
		"save_rate",
		"clickthrough",
		"clickthrough_rate",
		"engagement",
		"engagement_rate",
		"video_mrc_view",
		"video_start",
		"video_10s_view",
		"video_avg_watch_time",
		"video_v50_watch_time",
		"full_screen_play",
		"full_screen_playtime",
		"profile_visit",
		"closeup",
		"quartile_95s_percent_view",
		"user_follow",
		"day_of_week",
		"hour_of_day",
		"inserted_at",
	}
}

// PinterestUserInsightColumns returns the column names for batch insert.
func PinterestUserInsightColumns() []string {
	return []string{
		"record_id",
		"user_id",
		"date",
		"data_status",
		"impression",
		"pin_clicks",
		"pin_click_rate",
		"outbound_click",
		"saves",
		"save_rate",
		"clickthrough",
		"clickthrough_rate",
		"engagement",
		"engagement_rate",
		"video_mrc_view",
		"video_start",
		"video_10s_view",
		"video_avg_watch_time",
		"video_v50_watch_time",
		"full_screen_play",
		"full_screen_playtime",
		"profile_visit",
		"closeup",
		"quartile_95s_percent_view",
		"inserted_at",
	}
}
