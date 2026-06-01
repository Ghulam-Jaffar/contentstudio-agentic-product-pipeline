package clickhouse

import "time"

// FacebookCompetitorInsights represents competitor page insights
type FacebookCompetitorInsights struct {
	RecordID          string            `ch:"record_id"`
	PageID            string            `ch:"page_id"`
	FollowersCount    int64             `ch:"followers_count"`
	TotalFanCount     int64             `ch:"total_fan_count"`
	TalkingAboutThis  int64             `ch:"talking_about_this"`
	Biography         string            `ch:"biography"`
	ProfilePictureURL string            `ch:"profile_picture_url"`
	PageName          string            `ch:"page_name"`
	PageCategory      string            `ch:"page_category"`
	Emails            []string          `ch:"emails"`
	Birthday          string            `ch:"birthday"`
	WereHereCount     int64             `ch:"were_here_count"`
	CoverPhotoURL     string            `ch:"cover_photo_url"`
	Permalink         string            `ch:"permalink"`
	Metadata          map[string]string `ch:"metadata"`
	InsertedAt        time.Time         `ch:"inserted_at"`
}

// FacebookCompetitorPosts represents competitor posts
type FacebookCompetitorPosts struct {
	FacebookID         string    `ch:"facebook_id"`
	PostID             string    `ch:"post_id"`
	FollowersCount     int64     `ch:"followers_count"`
	FanCount           int64     `ch:"fan_count"`
	PageName           string    `ch:"page_name"`
	PageCategory       string    `ch:"page_category"`
	Biography          string    `ch:"biography"`
	PostEngagement     int64     `ch:"post_engagement"`
	Like               int64     `ch:"like"`
	Haha               int64     `ch:"haha"`
	Angry              int64     `ch:"angry"`
	Sad                int64     `ch:"sad"`
	Thankful           int64     `ch:"thankful"`
	Love               int64     `ch:"love"`
	TotalPostReactions int64     `ch:"total_post_reactions"`
	Comments           int64     `ch:"comments"`
	Shares             int64     `ch:"shares"`
	Caption            string    `ch:"caption"`
	MediaType          string    `ch:"media_type"`
	StatusType         string    `ch:"status_type"`
	SharedFromName     string    `ch:"shared_from_name"`
	SharedFromID       string    `ch:"shared_from_id"`
	SharedFromPic      string    `ch:"shared_from_pic"`
	SharedCreatedAt    time.Time `ch:"shared_created_at"`
	Permalink          string    `ch:"permalink"`
	Hashtags           []string  `ch:"hashtags"`
	DayOfWeek          string    `ch:"day_of_week"`
	HourOfDay          int64     `ch:"hour_of_day"`
	CreatedAt          time.Time `ch:"created_at"`
	InsertedAt         time.Time `ch:"inserted_at"`
	Wow                int64     `ch:"wow"`
}

// FacebookCompetitorMediaAssets represents competitor media assets
type FacebookCompetitorMediaAssets struct {
	MediaID      string    `ch:"media_id"`
	PostID       string    `ch:"post_id"`
	PageID       string    `ch:"page_id"`
	Caption      string    `ch:"caption"`
	Description  string    `ch:"description"`
	Link         string    `ch:"link"`
	AssetType    string    `ch:"asset_type"`
	CallToAction string    `ch:"call_to_action"`
	CTAType      string    `ch:"cta_type"`
	CreatedAt    time.Time `ch:"created_at"`
	InsertedAt   time.Time `ch:"inserted_at"`
}

// FacebookCompetitorMinimalMediaAsset stores only the asset URL fields needed by
// the URL refresher job. CreatedAt is required for partition-pruned mutations.
type FacebookCompetitorMinimalMediaAsset struct {
	PageID    string    `ch:"page_id"`
	PostID    string    `ch:"post_id"`
	MediaID   string    `ch:"media_id"`
	Link      string    `ch:"link"`
	CreatedAt time.Time `ch:"created_at"`
}

// FacebookCompetitorMinimalSharedPost stores only the shared-source fields needed
// to refresh shared_from_pic in facebook_competitor_posts. CreatedAt is required
// for partition-pruned mutations.
type FacebookCompetitorMinimalSharedPost struct {
	FacebookID    string    `ch:"facebook_id"`
	PostID        string    `ch:"post_id"`
	SharedFromID  string    `ch:"shared_from_id"`
	SharedFromPic string    `ch:"shared_from_pic"`
	CreatedAt     time.Time `ch:"created_at"`
}

// TableName returns the ClickHouse table name for insights
func (FacebookCompetitorInsights) TableName() string {
	return "facebook_competitor_insights"
}

// TableName returns the ClickHouse table name for posts
func (FacebookCompetitorPosts) TableName() string {
	return "facebook_competitor_posts"
}

// TableName returns the ClickHouse table name for media assets
func (FacebookCompetitorMediaAssets) TableName() string {
	return "facebook_competitor_media_assets"
}
