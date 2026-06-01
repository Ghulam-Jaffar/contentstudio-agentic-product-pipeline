package kafka

import (
	"time"
)

// YouTube Work Order Models

// YouTubeAccountWorkOrder represents a single YouTube account work order.
type YouTubeAccountWorkOrder struct {
	ID           string `json:"id"`
	ChannelID    string `json:"channel_id"`
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	WorkspaceID  string `json:"workspace_id"`
	SyncType     string `json:"sync_type"` // "incremental" | "full_sync"
}

// YouTubeBatchWorkOrder represents a batch of YouTube account work orders.
type YouTubeBatchWorkOrder struct {
	BatchID   string                    `json:"batch_id"`
	SyncType  string                    `json:"sync_type"`
	Accounts  []YouTubeAccountWorkOrder `json:"accounts"`
	CreatedAt time.Time                 `json:"created_at"`
}

// YouTube Raw Data Models (from API responses)

// RawYouTubeChannel represents raw channel data from YouTube Data API.
type RawYouTubeChannel struct {
	ChannelID       string    `json:"channel_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	CustomURL       string    `json:"custom_url"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	BannerURL       string    `json:"banner_url"`
	Country         string    `json:"country"`
	SubscriberCount int64     `json:"subscriber_count"`
	VideoCount      int64     `json:"video_count"`
	ViewCount       int64     `json:"view_count"`
	PublishedAt     time.Time `json:"published_at"`
	WorkspaceID     string    `json:"workspace_id"`
	SavingTime      time.Time `json:"saving_time"`
}

// RawYouTubeVideo represents raw video data from YouTube Data API.
type RawYouTubeVideo struct {
	VideoID                     string    `json:"video_id"`
	ChannelID                   string    `json:"channel_id"`
	Title                       string    `json:"title"`
	Description                 string    `json:"description"`
	ThumbnailURL                string    `json:"thumbnail_url"`
	Duration                    string    `json:"duration"` // ISO 8601 duration (e.g., PT1H2M3S)
	IframeEmbedHTML             string    `json:"iframe_embed_html"`
	PublishedAt                 time.Time `json:"published_at"`
	AnalyticsDate               time.Time `json:"analytics_date"` // Date of the analytics snapshot
	MediaType                   string    `json:"media_type"`     // "video" | "short"
	Views                       int64     `json:"views"`
	Likes                       int64     `json:"likes"`
	Dislikes                    int64     `json:"dislikes"`
	Comments                    int64     `json:"comments"`
	Shares                      int64     `json:"shares"`    // Not available with video dimension
	Favorites                   int64     `json:"favorites"` // Deprecated by YouTube (always 0)
	Saved                       int64     `json:"saved"`     // videosAddedToPlaylists
	SubscribersGained           int64     `json:"subscribers_gained"`
	RedViews                    int64     `json:"red_views"` // YouTube Partner only
	MinutesWatched              int64     `json:"minutes_watched"`
	RedMinutesWatched           int64     `json:"red_minutes_watched"` // YouTube Partner only
	AvgViewDuration             int64     `json:"average_view_duration"`
	AvgViewPercentage           float64   `json:"average_view_percentage"`
	Impressions                 int64     `json:"impressions"`
	ImpressionsClickThroughRate float64   `json:"impressions_click_through_rate"`
	WorkspaceID                 string    `json:"workspace_id"`
	SavingTime                  time.Time `json:"saving_time"`
}

// RawYouTubeActivityInsights represents daily activity insights from Analytics API.
type RawYouTubeActivityInsights struct {
	ChannelID                  string    `json:"channel_id"`
	Date                       time.Time `json:"date"`
	Views                      int64     `json:"views"`
	RedViews                   int64     `json:"red_views"`
	Likes                      int64     `json:"likes"`
	Dislikes                   int64     `json:"dislikes"`
	Comments                   int64     `json:"comments"`
	Shares                     int64     `json:"shares"`
	SubscribersGained          int64     `json:"subscribers_gained"`
	EstimatedMinutesWatched    int64     `json:"estimated_minutes_watched"`
	EstimatedRedMinutesWatched int64     `json:"estimated_red_minutes_watched"`
	AvgViewDuration            int64     `json:"average_view_duration"`
	AvgViewPercentage          float64   `json:"average_view_percentage"`
	WorkspaceID                string    `json:"workspace_id"`
	SavingTime                 time.Time `json:"saving_time"`
}

// RawYouTubeTrafficInsights represents traffic source insights from Analytics API.
type RawYouTubeTrafficInsights struct {
	ChannelID      string    `json:"channel_id"`
	Date           time.Time `json:"date"`
	TrafficSource  string    `json:"traffic_source"`
	Views          int64     `json:"views"`
	MinutesWatched int64     `json:"minutes_watched"`
	WorkspaceID    string    `json:"workspace_id"`
	SavingTime     time.Time `json:"saving_time"`
}

// RawYouTubeSharedInsights represents sharing service insights from Analytics API.
type RawYouTubeSharedInsights struct {
	ChannelID      string    `json:"channel_id"`
	SharingService string    `json:"sharing_service"`
	Shares         int64     `json:"shares"`
	WorkspaceID    string    `json:"workspace_id"`
	SavingTime     time.Time `json:"saving_time"`
}

// YouTube Parsed Data Models (analytics-ready for ClickHouse)

// ParsedYouTubeChannel represents parsed channel data ready for ClickHouse.
// Stores daily snapshots of channel data for historical tracking.
type ParsedYouTubeChannel struct {
	RecordID        string    `json:"record_id"` // MD5(channel_id + date) for daily deduplication
	ChannelID       string    `json:"channel_id"`
	Title           string    `json:"title"`
	Description     string    `json:"description"`
	CustomURL       string    `json:"custom_url"`
	ThumbnailURL    string    `json:"thumbnail_url"`
	BannerURL       string    `json:"external_banner_url"`
	Country         string    `json:"country"`
	SubscriberCount int64     `json:"subscriber_count"`
	VideoCount      int64     `json:"video_count"`
	ViewCount       int64     `json:"view_count"`
	PublishedAt     time.Time `json:"published_at"`
	CreatedAt       time.Time `json:"created_at"`  // Date of the snapshot (daily data)
	InsertedAt      time.Time `json:"inserted_at"` // When inserted into DB
}

// ParsedYouTubeVideo represents parsed video data ready for ClickHouse.
// Stores daily snapshots of video data for historical tracking.
type ParsedYouTubeVideo struct {
	VideoID                     string    `json:"video_id"`
	ChannelID                   string    `json:"channel_id"`
	Title                       string    `json:"title"`
	Description                 string    `json:"description"`
	Duration                    string    `json:"duration"`
	ThumbnailURL                string    `json:"thumbnail_url"`
	IframeEmbedHTML             string    `json:"iframe_embed_html"`
	MediaType                   string    `json:"media_type"` // "video" | "short"
	Likes                       int64     `json:"likes"`
	Dislikes                    int64     `json:"dislikes"`
	Views                       int64     `json:"views"`
	Comments                    int64     `json:"comments"`
	Shares                      int64     `json:"shares"`
	Favorites                   int64     `json:"favorites"`
	Saved                       int64     `json:"saved"` // videosAddedToPlaylists
	SubscribersGained           int64     `json:"subscribers_gained"`
	RedViews                    int64     `json:"red_views"`
	MinutesWatched              int64     `json:"minutes_watched"`
	RedMinutesWatched           int64     `json:"red_minutes_watched"`
	AvgViewDuration             int64     `json:"average_view_duration"`
	AvgViewPercentage           float64   `json:"average_view_percentage"`
	Impressions                 int64     `json:"impressions"`
	ImpressionsClickThroughRate float64   `json:"impressions_click_through_rate"`
	PublishedAt                 time.Time `json:"published_at"`
	CreatedAt                   time.Time `json:"created_at"`  // Date of the snapshot (daily data)
	InsertedAt                  time.Time `json:"inserted_at"` // When inserted into DB
}

// ParsedYouTubeActivityInsights represents parsed activity insights for ClickHouse.
type ParsedYouTubeActivityInsights struct {
	RecordID                   string    `json:"record_id"` // MD5(channel_id + date)
	ChannelID                  string    `json:"channel_id"`
	RedViews                   int64     `json:"red_views"`
	Views                      int64     `json:"views"`
	Likes                      int64     `json:"likes"`
	Dislikes                   int64     `json:"dislikes"`
	Comments                   int64     `json:"comments"`
	Shares                     int64     `json:"shares"`
	SubscribersGained          int64     `json:"subscribers_gained"`
	EstimatedMinutesWatched    int64     `json:"estimated_minutes_watched"`
	EstimatedRedMinutesWatched int64     `json:"estimated_red_minutes_watched"`
	AvgViewDuration            int64     `json:"average_view_duration"`
	AvgViewPercentage          float64   `json:"average_view_percentage"`
	CreatedAt                  time.Time `json:"created_at"`
}

// ParsedYouTubeTrafficInsights represents parsed traffic insights for ClickHouse.
// Traffic sources are aggregated by day into separate columns.
type ParsedYouTubeTrafficInsights struct {
	RecordID               string    `json:"record_id"` // MD5(channel_id + date)
	ChannelID              string    `json:"channel_id"`
	PaidViews              int64     `json:"paid_views"`
	AnnotationViews        int64     `json:"annotation_views"`
	EndScreenViews         int64     `json:"end_screen_views"`
	CampaignCardViews      int64     `json:"campaign_card_view"`
	SubscriberViews        int64     `json:"subscriber_views"`
	NoLinkOtherViews       int64     `json:"no_link_other_views"`
	YTChannelViews         int64     `json:"yt_channel_views"`
	YTSearchViews          int64     `json:"yt_search_views"`
	RelatedVideoViews      int64     `json:"related_video_views"`
	YTOtherPageViews       int64     `json:"yt_other_page_views"`
	ExtURLViews            int64     `json:"ext_url_views"`
	PlaylistViews          int64     `json:"playlist_views"`
	NotificationViews      int64     `json:"notification_views"`
	ShortsViews            int64     `json:"shorts_views"`
	SubscriberWatchTime    int64     `json:"subscriber_watch_time"`
	NonSubscriberWatchTime int64     `json:"non_subsciber_watch_time"` // Note: typo matches schema
	CreatedAt              time.Time `json:"created_at"`
}

// ParsedYouTubeSharedInsights represents parsed sharing insights for ClickHouse.
// Sharing services are aggregated into separate columns.
type ParsedYouTubeSharedInsights struct {
	RecordID      string    `json:"record_id"` // MD5(channel_id + date)
	ChannelID     string    `json:"channel_id"`
	Ameba         int64     `json:"ameba"`
	Blogger       int64     `json:"blogger"`
	CopyPaste     int64     `json:"copy_paste"`
	Cyworld       int64     `json:"cyworld"`
	Digg          int64     `json:"digg"`
	Dropbox       int64     `json:"dropbox"`
	Embed         int64     `json:"embed"`
	Mail          int64     `json:"mail"`
	WhatsApp      int64     `json:"whats_app"`
	Other         int64     `json:"other"`
	FacebookMsgr  int64     `json:"facebook_messenger"`
	FacebookPages int64     `json:"facebook_pages"`
	Facebook      int64     `json:"facebook"`
	Fotka         int64     `json:"fotka"`
	VKontakte     int64     `json:"vkontakte"`
	Discord       int64     `json:"discord"`
	GooglePlus    int64     `json:"google_plus"`
	Goo           int64     `json:"goo"`
	Hangouts      int64     `json:"hangouts"`
	LinkedIn      int64     `json:"linkedin"`
	Pinterest     int64     `json:"pinterest"`
	Myspace       int64     `json:"myspace"`
	Reddit        int64     `json:"reddit"`
	Skype         int64     `json:"skype"`
	Telegram      int64     `json:"telegram"`
	Twitter       int64     `json:"twitter"`
	Tumblr        int64     `json:"tumblr"`
	Viber         int64     `json:"viber"`
	Weibo         int64     `json:"weibo"`
	WeChat        int64     `json:"wechat"`
	YouTube       int64     `json:"youtube"`
	InsertedAt    time.Time `json:"inserted_at"`
}

// TrafficSourceType constants for YouTube traffic sources.
const (
	TrafficSourcePaid         = "PAID"
	TrafficSourceAnnotation   = "ANNOTATION"
	TrafficSourceEndScreen    = "END_SCREEN"
	TrafficSourceCampaignCard = "CAMPAIGN_CARD"
	TrafficSourceSubscriber   = "SUBSCRIBER"
	TrafficSourceNoLinkOther  = "NO_LINK_OTHER"
	TrafficSourceYTChannel    = "YT_CHANNEL"
	TrafficSourceYTSearch     = "YT_SEARCH"
	TrafficSourceRelatedVideo = "RELATED_VIDEO"
	TrafficSourceYTOtherPage  = "YT_OTHER_PAGE"
	TrafficSourceExtURL       = "EXT_URL"
	TrafficSourcePlaylist     = "PLAYLIST"
	TrafficSourceNotification = "NOTIFICATION"
	TrafficSourceShorts       = "SHORTS"
)

// SharingServiceType constants for YouTube sharing services.
const (
	SharingServiceAmeba         = "AMEBA"
	SharingServiceBlogger       = "BLOGGER"
	SharingServiceCopyPaste     = "COPY_PASTE"
	SharingServiceCyworld       = "CYWORLD"
	SharingServiceDigg          = "DIGG"
	SharingServiceDropbox       = "DROPBOX"
	SharingServiceEmbed         = "EMBED"
	SharingServiceMail          = "MAIL"
	SharingServiceWhatsApp      = "WHATS_APP"
	SharingServiceOther         = "OTHER"
	SharingServiceFacebookMsgr  = "FACEBOOK_MESSENGER"
	SharingServiceFacebookPages = "FACEBOOK_PAGES"
	SharingServiceFacebook      = "FACEBOOK"
	SharingServiceFotka         = "FOTKA"
	SharingServiceVKontakte     = "VKONTAKTE"
	SharingServiceDiscord       = "DISCORD"
	SharingServiceGooglePlus    = "GOOGLEPLUS"
	SharingServiceGoo           = "GOO"
	SharingServiceHangouts      = "HANGOUTS"
	SharingServiceLinkedIn      = "LINKEDIN"
	SharingServicePinterest     = "PINTEREST"
	SharingServiceMyspace       = "MYSPACE"
	SharingServiceReddit        = "REDDIT"
	SharingServiceSkype         = "SKYPE"
	SharingServiceTelegram      = "TELEGRAM"
	SharingServiceTwitter       = "TWITTER"
	SharingServiceTumblr        = "TUMBLR"
	SharingServiceViber         = "VIBER"
	SharingServiceWeibo         = "WEIBO"
	SharingServiceWeChat        = "WECHAT"
	SharingServiceYouTube       = "YOUTUBE"
	SharingServiceYouTubeGaming = "YOUTUBE_GAMING"
	SharingServiceYouTubeKids   = "YOUTUBE_KIDS"
	SharingServiceYouTubeMusic  = "YOUTUBE_MUSIC"
	SharingServiceYouTubeTV     = "YOUTUBE_TV"
)

// YouTubeVideoMediaType constants.
const (
	YouTubeMediaTypeVideo = "video"
	YouTubeMediaTypeShort = "short"
)

// YouTubeSyncType constants.
const (
	YouTubeSyncTypeIncremental = "incremental"
	YouTubeSyncTypeImmediate   = "immediate"
	YouTubeSyncTypeFullSync    = "full_sync"
)

// YouTubeKafkaTopics defines the Kafka topics for YouTube data.
var YouTubeKafkaTopics = struct {
	WorkOrder              string
	ImmediateWorkOrder     string
	RawChannels            string
	RawVideos              string
	RawActivityInsights    string
	RawTrafficInsights     string
	RawSharedInsights      string
	ParsedChannels         string
	ParsedVideos           string
	ParsedActivityInsights string
	ParsedTrafficInsights  string
	ParsedSharedInsights   string
}{
	WorkOrder:              "work-order-youtube",
	ImmediateWorkOrder:     "immediate-work-order-youtube",
	RawChannels:            "raw-youtube-channels",
	RawVideos:              "raw-youtube-videos",
	RawActivityInsights:    "raw-youtube-activity-insights",
	RawTrafficInsights:     "raw-youtube-traffic-insights",
	RawSharedInsights:      "raw-youtube-shared-insights",
	ParsedChannels:         "parsed-youtube-channels",
	ParsedVideos:           "parsed-youtube-videos",
	ParsedActivityInsights: "parsed-youtube-activity-insights",
	ParsedTrafficInsights:  "parsed-youtube-traffic-insights",
	ParsedSharedInsights:   "parsed-youtube-shared-insights",
}
