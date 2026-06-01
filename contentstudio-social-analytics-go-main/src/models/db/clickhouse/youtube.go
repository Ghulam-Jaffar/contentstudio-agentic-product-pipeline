package clickhouse

import (
	"time"
)

// YouTubeChannel represents a row in the youtube_channels table.
// Stores daily snapshots of channel data for historical tracking.
type YouTubeChannel struct {
	RecordID        string    `ch:"record_id"` // MD5(channel_id + date) for daily deduplication
	ChannelID       string    `ch:"channel_id"`
	Title           string    `ch:"title"`
	Description     string    `ch:"description"`
	CustomURL       string    `ch:"custom_url"`
	ThumbnailURL    string    `ch:"thumbnail_url"`
	ExternalBanner  string    `ch:"external_banner_url"`
	Country         string    `ch:"country"`
	SubscriberCount int64     `ch:"subscriber_count"`
	VideoCount      int64     `ch:"video_count"`
	ViewCount       int64     `ch:"view_count"`
	PublishedAt     time.Time `ch:"published_at"`
	CreatedAt       time.Time `ch:"created_at"`  // Date of the snapshot (daily data)
	InsertedAt      time.Time `ch:"inserted_at"` // When inserted into DB
}

// YouTubeVideo represents a row in the youtube_videos table.
// Stores daily snapshots of video data for historical tracking.
type YouTubeVideo struct {
	VideoID                     string    `ch:"video_id"`
	ChannelID                   string    `ch:"channel_id"`
	Title                       string    `ch:"title"`
	Description                 string    `ch:"description"`
	Duration                    string    `ch:"duration"`
	ThumbnailURL                string    `ch:"thumbnail_url"`
	IframeEmbedHTML             string    `ch:"iframe_embed_html"`
	Likes                       int64     `ch:"likes"`
	Dislikes                    int64     `ch:"dislikes"`
	Views                       int64     `ch:"views"`
	Comments                    int64     `ch:"comments"`
	Shares                      int64     `ch:"shares"`
	Favorites                   int64     `ch:"favorites"`
	Saved                       int64     `ch:"saved"`
	SubscribersGained           int64     `ch:"subscribers_gained"`
	RedViews                    int64     `ch:"red_views"`
	MinutesWatched              int64     `ch:"minutes_watched"`
	RedMinutesWatched           int64     `ch:"red_minutes_watched"`
	AvgViewDuration             int64   `ch:"average_view_duration"`
	AvgViewPercentage           float64 `ch:"average_view_percentage"`
	Impressions                 int64   `ch:"impressions"`
	ImpressionsClickThroughRate float64   `ch:"impressions_click_through_rate"`
	PublishedAt                 time.Time `ch:"published_at"`
	CreatedAt                   time.Time `ch:"created_at"`  // Date of the snapshot (daily data)
	InsertedAt                  time.Time `ch:"inserted_at"` // When inserted into DB
	MediaType                   string    `ch:"media_type"`
}

// YouTubeActivityInsights represents a row in the youtube_activity_insights table.
type YouTubeActivityInsights struct {
	RecordID                   string    `ch:"record_id"`
	ChannelID                  string    `ch:"channel_id"`
	RedViews                   int64     `ch:"red_views"`
	Views                      int64     `ch:"views"`
	Likes                      int64     `ch:"likes"`
	Dislikes                   int64     `ch:"dislikes"`
	Comments                   int64     `ch:"comments"`
	Shares                     int64     `ch:"shares"`
	SubscribersGained          int64     `ch:"subscribers_gained"`
	EstimatedMinutesWatched    int64     `ch:"estimated_minutes_watched"`
	EstimatedRedMinutesWatched int64     `ch:"estimated_red_minutes_watched"`
	AvgViewDuration            int64     `ch:"average_view_duration"`
	AvgViewPercentage          float64   `ch:"average_view_percentage"`
	CreatedAt                  time.Time `ch:"created_at"`
	InsertedAt                 time.Time `ch:"inserted_at"`
}

// YouTubeTrafficInsights represents a row in the youtube_traffic_insights table.
type YouTubeTrafficInsights struct {
	RecordID               string    `ch:"record_id"`
	ChannelID              string    `ch:"channel_id"`
	PaidViews              int64     `ch:"paid_views"`
	AnnotationViews        int64     `ch:"annotation_views"`
	EndScreenViews         int64     `ch:"end_screen_views"`
	CampaignCardViews      int64     `ch:"campaign_card_view"`
	SubscriberViews        int64     `ch:"subscriber_views"`
	NoLinkOtherViews       int64     `ch:"no_link_other_views"`
	YTChannelViews         int64     `ch:"yt_channel_views"`
	YTSearchViews          int64     `ch:"yt_search_views"`
	RelatedVideoViews      int64     `ch:"related_video_views"`
	YTOtherPageViews       int64     `ch:"yt_other_page_views"`
	ExtURLViews            int64     `ch:"ext_url_views"`
	PlaylistViews          int64     `ch:"playlist_views"`
	NotificationViews      int64     `ch:"notification_views"`
	SubscriberWatchTime    int64     `ch:"subscriber_watch_time"`
	NonSubscriberWatchTime int64     `ch:"non_subsciber_watch_time"` // typo matches schema
	CreatedAt              time.Time `ch:"created_at"`
	ShortsViews            int64     `ch:"shorts_views"`
}

// YouTubeSharedInsights represents a row in the youtube_shared_insights table.
type YouTubeSharedInsights struct {
	RecordID      string    `ch:"record_id"`
	ChannelID     string    `ch:"channel_id"`
	Ameba         int64     `ch:"ameba"`
	Blogger       int64     `ch:"blogger"`
	CopyPaste     int64     `ch:"copy_paste"`
	Cyworld       int64     `ch:"cyworld"`
	Digg          int64     `ch:"digg"`
	Dropbox       int64     `ch:"dropbox"`
	Embed         int64     `ch:"embed"`
	Mail          int64     `ch:"mail"`
	WhatsApp      int64     `ch:"whats_app"`
	Other         int64     `ch:"other"`
	FacebookMsgr  int64     `ch:"facebook_messenger"`
	FacebookPages int64     `ch:"facebook_pages"`
	Facebook      int64     `ch:"facebook"`
	Fotka         int64     `ch:"fotka"`
	VKontakte     int64     `ch:"vkontakte"`
	Discord       int64     `ch:"discord"`
	GooglePlus    int64     `ch:"google_plus"`
	Goo           int64     `ch:"goo"`
	Hangouts      int64     `ch:"hangouts"`
	LinkedIn      int64     `ch:"linkedin"`
	Pinterest     int64     `ch:"pinterest"`
	Myspace       int64     `ch:"myspace"`
	Reddit        int64     `ch:"reddit"`
	Skype         int64     `ch:"skype"`
	Telegram      int64     `ch:"telegram"`
	Twitter       int64     `ch:"twitter"`
	Tumblr        int64     `ch:"tumblr"`
	Viber         int64     `ch:"viber"`
	Weibo         int64     `ch:"weibo"`
	WeChat        int64     `ch:"wechat"`
	YouTube       int64     `ch:"youtube"`
	InsertedAt    time.Time `ch:"inserted_at"`
}

// TableName returns the ClickHouse table name for YouTubeChannel.
func (YouTubeChannel) TableName() string {
	return "youtube_channels"
}

// TableName returns the ClickHouse table name for YouTubeVideo.
func (YouTubeVideo) TableName() string {
	return "youtube_videos"
}

// TableName returns the ClickHouse table name for YouTubeActivityInsights.
func (YouTubeActivityInsights) TableName() string {
	return "youtube_activity_insights"
}

// TableName returns the ClickHouse table name for YouTubeTrafficInsights.
func (YouTubeTrafficInsights) TableName() string {
	return "youtube_traffic_insights"
}

// TableName returns the ClickHouse table name for YouTubeSharedInsights.
func (YouTubeSharedInsights) TableName() string {
	return "youtube_shared_insights"
}

// YouTubeChannelColumns returns the column names for batch insert.
func YouTubeChannelColumns() []string {
	return []string{
		"record_id",
		"channel_id",
		"title",
		"description",
		"custom_url",
		"thumbnail_url",
		"external_banner_url",
		"country",
		"subscriber_count",
		"video_count",
		"view_count",
		"published_at",
		"created_at",
		"inserted_at",
	}
}

// YouTubeVideoColumns returns the column names for batch insert.
func YouTubeVideoColumns() []string {
	return []string{
		"video_id",
		"channel_id",
		"title",
		"description",
		"duration",
		"thumbnail_url",
		"iframe_embed_html",
		"likes",
		"dislikes",
		"views",
		"comments",
		"shares",
		"favorites",
		"saved",
		"subscribers_gained",
		"red_views",
		"minutes_watched",
		"red_minutes_watched",
		"average_view_duration",
		"average_view_percentage",
		"impressions",
		"impressions_click_through_rate",
		"published_at",
		"created_at",
		"inserted_at",
		"media_type",
	}
}

// YouTubeActivityInsightsColumns returns the column names for batch insert.
func YouTubeActivityInsightsColumns() []string {
	return []string{
		"record_id",
		"channel_id",
		"red_views",
		"views",
		"likes",
		"dislikes",
		"comments",
		"shares",
		"subscribers_gained",
		"estimated_minutes_watched",
		"estimated_red_minutes_watched",
		"average_view_duration",
		"average_view_percentage",
		"created_at",
		"inserted_at",
	}
}

// YouTubeTrafficInsightsColumns returns the column names for batch insert.
func YouTubeTrafficInsightsColumns() []string {
	return []string{
		"record_id",
		"channel_id",
		"paid_views",
		"annotation_views",
		"end_screen_views",
		"campaign_card_view",
		"subscriber_views",
		"no_link_other_views",
		"yt_channel_views",
		"yt_search_views",
		"related_video_views",
		"yt_other_page_views",
		"ext_url_views",
		"playlist_views",
		"notification_views",
		"subscriber_watch_time",
		"non_subsciber_watch_time",
		"created_at",
		"shorts_views",
	}
}

// YouTubeSharedInsightsColumns returns the column names for batch insert.
func YouTubeSharedInsightsColumns() []string {
	return []string{
		"record_id",
		"channel_id",
		"ameba",
		"blogger",
		"copy_paste",
		"cyworld",
		"digg",
		"dropbox",
		"embed",
		"mail",
		"whats_app",
		"other",
		"facebook_messenger",
		"facebook_pages",
		"facebook",
		"fotka",
		"vkontakte",
		"discord",
		"google_plus",
		"goo",
		"hangouts",
		"linkedin",
		"pinterest",
		"myspace",
		"reddit",
		"skype",
		"telegram",
		"twitter",
		"tumblr",
		"viber",
		"weibo",
		"wechat",
		"youtube",
		"inserted_at",
	}
}
