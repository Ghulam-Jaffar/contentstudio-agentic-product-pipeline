package clickhouse

import (
	"time"
)

// InstagramPost represents the Instagram posts analytics data for ClickHouse
// Corresponds to Python InstagramPosts
type InstagramPost struct {
	InstagramID         string    `ch:"instagram_id" json:"instagram_id"`
	MediaID             string    `ch:"media_id" json:"media_id"`
	Username            string    `ch:"username" json:"username"`
	Name                string    `ch:"name" json:"name"`
	ProfilePictureURL   string    `ch:"profile_picture_url" json:"profile_picture_url"`
	Permalink           string    `ch:"permalink" json:"permalink"`
	LikeCount           int64     `ch:"like_count" json:"like_count"`
	CommentsCount       int64     `ch:"comments_count" json:"comments_count"`
	Engagement          int64     `ch:"engagement" json:"engagement"`
	Impressions         int64     `ch:"impressions" json:"impressions"`
	Views               int64     `ch:"views" json:"views"`
	Reach               int64     `ch:"reach" json:"reach"`
	Saved               int64     `ch:"saved" json:"saved"`
	VideoViews          int64     `ch:"video_views" json:"video_views"`
	Shares              int64     `ch:"shares" json:"shares"`
	ReelsAvgWatchTime   int64     `ch:"reels_avg_watch_time" json:"reels_avg_watch_time"`
	ReelsTotalWatchTime int64     `ch:"reels_total_watch_time" json:"reels_total_watch_time"`
	Exits               int64     `ch:"exits" json:"exits"`
	Replies             int64     `ch:"replies" json:"replies"`
	TapsForward         int64     `ch:"taps_forward" json:"taps_forward"`
	TapsBack            int64     `ch:"taps_back" json:"taps_back"`
	ChildAssetsType     []string  `ch:"child_assets_type" json:"child_assets_type"`
	Caption             string    `ch:"caption" json:"caption"`
	MediaType           string    `ch:"media_type" json:"media_type"`
	EntityType          string    `ch:"entity_type" json:"entity_type"`
	MediaURL            []string  `ch:"media_url" json:"media_url"`
	VideoURL            []string  `ch:"video_url" json:"video_url"`
	Hashtags            []string  `ch:"hashtags" json:"hashtags"`
	DayOfWeek           string    `ch:"day_of_week" json:"day_of_week"`
	HourOfDay           int64     `ch:"hour_of_day" json:"hour_of_day"`
	Year                int64     `ch:"year" json:"year"`
	Month               int64     `ch:"month" json:"month"`
	Timestamp           int64     `ch:"timestamp" json:"timestamp"`
	StoredEventAt       time.Time `ch:"stored_event_at" json:"stored_event_at"`
	PostCreatedAt       time.Time `ch:"post_created_at" json:"post_created_at"`
}

type InstagramMinimalPost struct {
	InstagramID string   `ch:"instagram_id" json:"instagram_id"`
	MediaID     string   `ch:"media_id" json:"media_id"`
	MediaURL    []string `ch:"media_url" json:"media_url"`
	VideoURL    []string `ch:"video_url" json:"video_url"`
}

// InstagramInsight represents the Instagram account insights data for ClickHouse
// Corresponds to Python InstagramInsights
type InstagramInsight struct {
	InstagramID                   string            `ch:"instagram_id" json:"instagram_id"`
	RecordID                      string            `ch:"record_id" json:"record_id"`
	Name                          string            `ch:"name" json:"name"`
	Username                      string            `ch:"username" json:"username"`
	ProfilePictureURL             string            `ch:"profile_picture_url" json:"profile_picture_url"`
	FollowsCount                  int64             `ch:"follows_count" json:"follows_count"`
	FollowersCount                int64             `ch:"followers_count" json:"followers_count"`
	MediaCount                    int64             `ch:"media_count" json:"media_count"`
	Tags                          int64             `ch:"tags" json:"tags"`
	Impressions                   int64             `ch:"impressions" json:"impressions"`
	ProfileViews                  int64             `ch:"profile_views" json:"profile_views"`
	Shares                        int64             `ch:"shares" json:"shares"`
	AccountsEngaged               int64             `ch:"accounts_engaged" json:"accounts_engaged"`
	Engagement                    int64             `ch:"engagement" json:"engagement"`
	Reach                         int64             `ch:"reach" json:"reach"`
	Views                         int64             `ch:"views" json:"views"`
	Saves                         int64             `ch:"saves" json:"saves"`
	Likes                         int64             `ch:"likes" json:"likes"`
	Comments                      int64             `ch:"comments" json:"comments"`
	AudienceAge                   []string          `ch:"audience_age" json:"audience_age"`
	AudienceGender                []string          `ch:"audience_gender" json:"audience_gender"`
	AudienceGenderAge             []string          `ch:"audience_gender_age" json:"audience_gender_age"`
	AudienceLocale                []string          `ch:"audience_locale" json:"audience_locale"`
	AudienceCity                  []string          `ch:"audience_city" json:"audience_city"`
	AudienceCountry               []string          `ch:"audience_country" json:"audience_country"`
	AudienceAgeByEngagement       []string          `ch:"audience_age_by_engagement" json:"audience_age_by_engagement"`
	AudienceGenderByEngagement    []string          `ch:"audience_gender_by_engagement" json:"audience_gender_by_engagement"`
	AudienceGenderAgeByEngagement []string          `ch:"audience_gender_age_by_engagement" json:"audience_gender_age_by_engagement"`
	AudienceCityByEngagement      []string          `ch:"audience_city_by_engagement" json:"audience_city_by_engagement"`
	AudienceCountryByEngagement   []string          `ch:"audience_country_by_engagement" json:"audience_country_by_engagement"`
	AudienceAgeByReach            []string          `ch:"audience_age_by_reach" json:"audience_age_by_reach"`
	AudienceGenderByReach         []string          `ch:"audience_gender_by_reach" json:"audience_gender_by_reach"`
	AudienceGenderAgeByReach      []string          `ch:"audience_gender_age_by_reach" json:"audience_gender_age_by_reach"`
	AudienceCityByReach           []string          `ch:"audience_city_by_reach" json:"audience_city_by_reach"`
	AudienceCountryByReach        []string          `ch:"audience_country_by_reach" json:"audience_country_by_reach"`
	OnlineFollowers               []string          `ch:"online_followers" json:"online_followers"`
	AudienceDatetime              time.Time         `ch:"audience_datetime" json:"audience_datetime"`
	OnlineUsersDatetime           time.Time         `ch:"online_users_datetime" json:"online_users_datetime"`
	DayOfWeek                     string            `ch:"day_of_week" json:"day_of_week"`
	Year                          int64             `ch:"year" json:"year"`
	Month                         int64             `ch:"month" json:"month"`
	CreatedTime                   time.Time         `ch:"created_time" json:"created_time"`
	UpdatedTime                   time.Time         `ch:"updated_time" json:"updated_time"`
	Metadata                      map[string]string `ch:"metadata" json:"metadata"`
	StoredEventAt                 time.Time         `ch:"stored_event_at" json:"stored_event_at"`
}
