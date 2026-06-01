package clickhouse

import (
	"time"
)

// FacebookInsights represents the Facebook page insights data for ClickHouse
// Corresponds to Python FacebookInsightsModel
type FacebookInsights struct {
	// IDs
	HashID       string `ch:"hash_id" json:"hash_id"`
	PageID       string `ch:"page_id" json:"page_id"`
	PageCategory string `ch:"page_category" json:"page_category"`

	// Date
	DayOfWeek   string    `ch:"day_of_week" json:"day_of_week"`
	Year        int64     `ch:"year" json:"year"`
	Month       int64     `ch:"month" json:"month"`
	CreatedTime time.Time `ch:"created_time" json:"created_time"` // The actual date the data belongs to
	SavingTime  time.Time `ch:"saving_time" json:"saving_time"`   // When the record was saved

	// Fans info and demographics
	PageFans          int64    `ch:"page_fans" json:"page_fans"`
	PageFansCity      []string `ch:"page_fans_city" json:"page_fans_city"`
	PageFansCountry   []string `ch:"page_fans_country" json:"page_fans_country"`
	PageFansLocale    []string `ch:"page_fans_locale" json:"page_fans_locale"`
	PageFansAge       []string `ch:"page_fans_age" json:"page_fans_age"`
	PageFansGender    []string `ch:"page_fans_gender" json:"page_fans_gender"`
	PageFansGenderAge []string `ch:"page_fans_gender_age" json:"page_fans_gender_age"`

	PageFollows int64 `ch:"page_follows" json:"page_follows"`
	PageViews   int64 `ch:"page_views" json:"page_views"`

	// Paid/unpaid fans and total
	PageFanAddsByPaidNonPaidUnique []string `ch:"page_fan_adds_by_paid_non_paid_unique" json:"page_fan_adds_by_paid_non_paid_unique"`

	// New added and removed fans
	PageFanAddsUnique    int64 `ch:"page_fan_adds_unique" json:"page_fan_adds_unique"`
	PageFanRemovesUnique int64 `ch:"page_fan_removes_unique" json:"page_fan_removes_unique"`

	// Liked pages by source
	PageFansByLikeSourceUnique   []string `ch:"page_fans_by_like_source_unique" json:"page_fans_by_like_source_unique"`
	PageFansByUnlikeSourceUnique []string `ch:"page_fans_by_unlike_source_unique" json:"page_fans_by_unlike_source_unique"`

	// Total fans likes and unlikes
	PageFansByLike   int64 `ch:"page_fans_by_like" json:"page_fans_by_like"`
	PageFansByUnlike int64 `ch:"page_fans_by_unlike" json:"page_fans_by_unlike"`

	// Total clicks
	PageTotalActions int64 `ch:"page_total_actions" json:"page_total_actions"`

	// Total engagements
	PagePostEngagements int64 `ch:"page_post_engagements" json:"page_post_engagements"`

	// Total impressions
	PageImpressions       int64 `ch:"page_impressions" json:"page_impressions"`
	PageImpressionsUnique int64 `ch:"page_impressions_unique" json:"page_impressions_unique"`

	// Impressions paid or organic
	PageImpressionsOrganic int64 `ch:"page_impressions_organic" json:"page_impressions_organic"`
	PageImpressionsPaid    int64 `ch:"page_impressions_paid" json:"page_impressions_paid"`

	// Video views paid, total and organic
	PageVideoViewsPaid    int64 `ch:"page_video_views_paid" json:"page_video_views_paid"`
	PageVideoViews        int64 `ch:"page_video_views" json:"page_video_views"`
	PageVideoViewsOrganic int64 `ch:"page_video_views_organic" json:"page_video_views_organic"`

	// Metrics for video play
	PageVideoViewsAutoplayed  int64 `ch:"page_video_views_autoplayed" json:"page_video_views_autoplayed"`
	PageVideoViewsClickToPlay int64 `ch:"page_video_views_click_to_play" json:"page_video_views_click_to_play"`
	PageVideoRepeatViews      int64 `ch:"page_video_repeat_views" json:"page_video_repeat_views"`

	// Total feedback count
	PageNegativeFeedback int64 `ch:"page_negative_feedback" json:"page_negative_feedback"`
	PagePositiveFeedback int64 `ch:"page_positive_feedback" json:"page_positive_feedback"`

	// Feedback types
	PageNegativeFeedbackByType []string `ch:"page_negative_feedback_by_type" json:"page_negative_feedback_by_type"`
	PagePositiveFeedbackByType []string `ch:"page_positive_feedback_by_type" json:"page_positive_feedback_by_type"`

	// Total fans online at what hour
	PageFansOnline []string `ch:"page_fans_online" json:"page_fans_online"`
	ActiveUsers    int64    `ch:"active_users" json:"active_users"`

	// Sentiments towards the page
	PositiveSentiment int64 `ch:"positive_sentiment" json:"positive_sentiment"`
	NegativeSentiment int64 `ch:"negative_sentiment" json:"negative_sentiment"`

	// Posts count
	PostsCount        int64 `ch:"posts_count" json:"posts_count"`
	LikesCount        int64 `ch:"likes_count" json:"likes_count"`
	TalkingAboutCount int64 `ch:"talking_about_count" json:"talking_about_count"`

	// Types of posts: links, images, videos
	TypeCount []string `ch:"type_count" json:"type_count"`

	// Posts sent and received
	MessageCount []string `ch:"message_count" json:"message_count"`

	// Prime time
	PrimeTime time.Time `ch:"prime_time" json:"prime_time"`
}

// FacebookPosts represents the Facebook posts analytics data for ClickHouse
// Corresponds to Python FacebookAnalyticsModel
type FacebookPosts struct {
	PageName                     string    `ch:"page_name" json:"page_name"`
	PageID                       string    `ch:"page_id" json:"page_id"`
	MediaType                    string    `ch:"media_type" json:"media_type"`
	PostID                       string    `ch:"post_id" json:"post_id"`
	Permalink                    string    `ch:"permalink" json:"permalink"`
	StatusType                   string    `ch:"status_type" json:"status_type"`
	VideoID                      string    `ch:"video_id" json:"video_id"`
	Category                     string    `ch:"category" json:"category"`
	PublishedBy                  string    `ch:"published_by" json:"published_by"`
	PublishedByURL               string    `ch:"published_by_url" json:"published_by_url"`
	SharedFromName               string    `ch:"shared_from_name" json:"shared_from_name"`
	SharedFromID                 string    `ch:"shared_from_id" json:"shared_from_id"`
	SharedFromLink               string    `ch:"shared_from_link" json:"shared_from_link"`
	Like                         int32     `ch:"like" json:"like"`
	Love                         int32     `ch:"love" json:"love"`
	Haha                         int32     `ch:"haha" json:"haha"`
	Wow                          int32     `ch:"wow" json:"wow"`
	Sad                          int32     `ch:"sad" json:"sad"`
	Angry                        int32     `ch:"angry" json:"angry"`
	Thankful                     int32     `ch:"thankful" json:"thankful"`
	Total                        int64     `ch:"total" json:"total"`
	Shares                       int32     `ch:"shares" json:"shares"`
	Comments                     int32     `ch:"comments" json:"comments"`
	PostClicks                   int64     `ch:"post_clicks" json:"post_clicks"`
	TotalEngagement              int64     `ch:"total_engagement" json:"total_engagement"`
	PostEngagedUsers             int64     `ch:"post_engaged_users" json:"post_engaged_users"`
	DayOfWeek                    string    `ch:"day_of_week" json:"day_of_week"`
	HourOfDay                    int32     `ch:"hour_of_day" json:"hour_of_day"`
	CreatedTime                  time.Time `ch:"created_time" json:"created_time"`
	UpdatedTime                  time.Time `ch:"updated_time" json:"updated_time"`
	SavingTime                   time.Time `ch:"saving_time" json:"saving_time"`
	MessageTags                  []string  `ch:"message_tags" json:"message_tags"`
	PostMetadata                 string    `ch:"post_metadata" json:"post_metadata"`
	Caption                      string    `ch:"caption" json:"caption"`
	Description                  string    `ch:"description" json:"description"`
	FullPicture                  string    `ch:"full_picture" json:"full_picture"`
	Link                         string    `ch:"link" json:"link"`
	PostImpressions              int64     `ch:"post_impressions" json:"post_impressions"`
	PostImpressionsUnique        int64     `ch:"post_impressions_unique" json:"post_impressions_unique"`
	PostImpressionsPaid          int64     `ch:"post_impressions_paid" json:"post_impressions_paid"`
	PostImpressionsPaidUnique    int64     `ch:"post_impressions_paid_unique" json:"post_impressions_paid_unique"`
	PostImpressionsOrganic       int64     `ch:"post_impressions_organic" json:"post_impressions_organic"`
	PostImpressionsOrganicUnique int64     `ch:"post_impressions_organic_unique" json:"post_impressions_organic_unique"`
	PostImpressionsViral         int64     `ch:"post_impressions_viral" json:"post_impressions_viral"`
	PostImpressionsViralUnique   int64     `ch:"post_impressions_viral_unique" json:"post_impressions_viral_unique"`
	PostVideoViews               int64     `ch:"post_video_views" json:"post_video_views"`
	TotalImpressions             int64     `ch:"total_impressions" json:"total_impressions"`
}

// FacebookMediaAssets represents the Facebook media assets data for ClickHouse
// Corresponds to Python FacebookMediaAssests
type FacebookMediaAssets struct {
	PageID       string    `ch:"page_id" json:"page_id"`
	MediaID      string    `ch:"media_id" json:"media_id"`
	PostID       string    `ch:"post_id" json:"post_id"`
	AssetType    string    `ch:"asset_type" json:"asset_type"`
	Link         string    `ch:"link" json:"link"`
	CallToAction string    `ch:"call_to_action" json:"call_to_action"`
	CTAType      string    `ch:"CTA_type" json:"CTA_type"`
	Caption      string    `ch:"caption" json:"caption"`
	Description  string    `ch:"description" json:"description"`
	CreatedAt    time.Time `ch:"created_at" json:"created_at"`
	InsertedAt   time.Time `ch:"inserted_at" json:"inserted_at"`
}

// FacebookReelsInsights represents the Facebook reels insights data for ClickHouse
// Corresponds to Python FacebookReelsInsights
type FacebookReelsInsights struct {
	PageID               string    `ch:"page_id" json:"page_id"`
	PostID               string    `ch:"post_id" json:"post_id"`
	AverageTimeWatched   int64     `ch:"average_time_watched" json:"average_time_watched"`
	TotalTimeWatchedInMs int64     `ch:"total_time_watched_in_ms" json:"total_time_watched_in_ms"`
	PlayCount            int64     `ch:"play_count" json:"play_count"`
	ImpressionsUnique    int64     `ch:"impressions_unique" json:"impressions_unique"`
	ReelFollowers        int64     `ch:"reel_followers" json:"reel_followers"`
	CreatedAt            time.Time `ch:"created_at" json:"created_at"`
	SavingTime           time.Time `ch:"saving_time" json:"saving_time"`
}

// FacebookVideoInsights represents the Facebook video insights data for ClickHouse
// Corresponds to Python FacebookVideoInsights
type FacebookVideoInsights struct {
	// ID and time
	PostID      string    `ch:"post_id" json:"post_id"`
	PageID      string    `ch:"page_id" json:"page_id"`
	VideoID     string    `ch:"video_id" json:"video_id"`
	CreatedTime time.Time `ch:"created_time" json:"created_time"`
	UpdatedTime time.Time `ch:"updated_time" json:"updated_time"`

	// Followers and Views stats
	TotalVideoFollowers               int64    `ch:"total_video_followers" json:"total_video_followers"`
	TotalVideoViews                   int64    `ch:"total_video_views" json:"total_video_views"`
	TotalVideoViewsUnique             int64    `ch:"total_video_views_unique" json:"total_video_views_unique"`
	TotalVideoViewsAutoplayed         int64    `ch:"total_video_views_autoplayed" json:"total_video_views_autoplayed"`
	TotalVideoViewsOrganic            int64    `ch:"total_video_views_organic" json:"total_video_views_organic"`
	TotalVideoViewsOrganicUnique      int64    `ch:"total_video_views_organic_unique" json:"total_video_views_organic_unique"`
	TotalVideoViewsPaid               int64    `ch:"total_video_views_paid" json:"total_video_views_paid"`
	TotalVideoViewsPaidUnique         int64    `ch:"total_video_views_paid_unique" json:"total_video_views_paid_unique"`
	TotalVideoViewsSoundOn            int64    `ch:"total_video_views_sound_on" json:"total_video_views_sound_on"`
	TotalVideoViewsByDistributionType []string `ch:"total_video_views_by_distribution_type" json:"total_video_views_by_distribution_type"`

	// View time demographics
	TotalVideoViewTimeByDistributionType   []string `ch:"total_video_view_time_by_distribution_type" json:"total_video_view_time_by_distribution_type"`
	TotalVideoViewTimeByCountryID          []string `ch:"total_video_view_time_by_country_id" json:"total_video_view_time_by_country_id"`
	TotalVideoViewTimeByRegionID           []string `ch:"total_video_view_time_by_region_id" json:"total_video_view_time_by_region_id"`
	TotalVideoViewTimeByAgeBucketAndGender []string `ch:"total_video_view_time_by_age_bucket_and_gender" json:"total_video_view_time_by_age_bucket_and_gender"`

	// Views stats by watch time
	TotalVideoPlayCount                             int64    `ch:"total_video_play_count" json:"total_video_play_count"`
	TotalVideoConsumptionRate                       float64  `ch:"total_video_consumption_rate" json:"total_video_consumption_rate"`
	TotalVideoCompleteViews                         int64    `ch:"total_video_complete_views" json:"total_video_complete_views"`
	TotalVideoCompleteViewsUnique                   int64    `ch:"total_video_complete_views_unique" json:"total_video_complete_views_unique"`
	TotalVideoCompleteViewsAutoplayed               int64    `ch:"total_video_complete_views_autoplayed" json:"total_video_complete_views_autoplayed"`
	TotalVideoCompleteViewsClickedToPlay            int64    `ch:"total_video_complete_views_clicked_to_play" json:"total_video_complete_views_clicked_to_play"`
	TotalVideoCompleteViewsOrganic                  int64    `ch:"total_video_complete_views_organic" json:"total_video_complete_views_organic"`
	TotalVideoCompleteViewsOrganicUnique            int64    `ch:"total_video_complete_views_organic_unique" json:"total_video_complete_views_organic_unique"`
	TotalVideoCompleteViewsPaid                     int64    `ch:"total_video_complete_views_paid" json:"total_video_complete_views_paid"`
	TotalVideoCompleteViewsPaidUnique               int64    `ch:"total_video_complete_views_paid_unique" json:"total_video_complete_views_paid_unique"`
	VideoAsset60sVideoViewTotalCountByIsMonetizable []string `ch:"video_asset_60s_video_view_total_count_by_is_monetizable" json:"video_asset_60s_video_view_total_count_by_is_monetizable"`
	TotalVideo15minExcludesShorterViews             int64    `ch:"total_video_15min_excludes_shorter_views" json:"total_video_15min_excludes_shorter_views"`
	TotalVideo15minExcludesShorterViewsUnique       int64    `ch:"total_video_15min_excludes_shorter_views_unique" json:"total_video_15min_excludes_shorter_views_unique"`
	TotalVideo60sExcludesShorterViews               int64    `ch:"total_video_60s_excludes_shorter_views" json:"total_video_60s_excludes_shorter_views"`
	TotalVideo30sViews                              int64    `ch:"total_video_30s_views" json:"total_video_30s_views"`
	TotalVideo30sViewsUnique                        int64    `ch:"total_video_30s_views_unique" json:"total_video_30s_views_unique"`
	TotalVideo30sViewsAutoplayed                    int64    `ch:"total_video_30s_views_autoplayed" json:"total_video_30s_views_autoplayed"`
	TotalVideo30sViewsClickedToPlay                 int64    `ch:"total_video_30s_views_clicked_to_play" json:"total_video_30s_views_clicked_to_play"`
	TotalVideo30sViewsOrganic                       int64    `ch:"total_video_30s_views_organic" json:"total_video_30s_views_organic"`
	TotalVideo30sViewsPaid                          int64    `ch:"total_video_30s_views_paid" json:"total_video_30s_views_paid"`
	TotalVideo30sViewsSoundOn                       int64    `ch:"total_video_30s_views_sound_on" json:"total_video_30s_views_sound_on"`
	TotalVideo10sViews                              int64    `ch:"total_video_10s_views" json:"total_video_10s_views"`
	TotalVideo10sViewsUnique                        int64    `ch:"total_video_10s_views_unique" json:"total_video_10s_views_unique"`
	TotalVideo10sViewsAutoplayed                    int64    `ch:"total_video_10s_views_autoplayed" json:"total_video_10s_views_autoplayed"`
	TotalVideo10sViewsClickedToPlay                 int64    `ch:"total_video_10s_views_clicked_to_play" json:"total_video_10s_views_clicked_to_play"`
	TotalVideo10sViewsOrganic                       int64    `ch:"total_video_10s_views_organic" json:"total_video_10s_views_organic"`
	TotalVideo10sViewsPaid                          int64    `ch:"total_video_10s_views_paid" json:"total_video_10s_views_paid"`
	TotalVideo10sViewsSoundOn                       int64    `ch:"total_video_10s_views_sound_on" json:"total_video_10s_views_sound_on"`
	TotalVideo15sViews                              int64    `ch:"total_video_15s_views" json:"total_video_15s_views"`
	TotalVideoAvgTimeWatched                        int64    `ch:"total_video_avg_time_watched" json:"total_video_avg_time_watched"`
	TotalVideoViewTotalTime                         int64    `ch:"total_video_view_total_time" json:"total_video_view_total_time"`
	TotalVideoViewTotalTimeOrganic                  int64    `ch:"total_video_view_total_time_organic" json:"total_video_view_total_time_organic"`
	TotalVideoViewTotalTimePaid                     int64    `ch:"total_video_view_total_time_paid" json:"total_video_view_total_time_paid"`

	// Audience retention info
	TotalVideoRetentionGraphAutoplayed    []string `ch:"total_video_retention_graph_autoplayed" json:"total_video_retention_graph_autoplayed"`
	TotalVideoRetentionGraphClickedToPlay []string `ch:"total_video_retention_graph_clicked_to_play" json:"total_video_retention_graph_clicked_to_play"`
	TotalVideoRetentionGraphGenderMale    []string `ch:"total_video_retention_graph_gender_male" json:"total_video_retention_graph_gender_male"`
	TotalVideoRetentionGraphGenderFemale  []string `ch:"total_video_retention_graph_gender_female" json:"total_video_retention_graph_gender_female"`

	// Impressions
	TotalVideoImpressions              int64    `ch:"total_video_impressions" json:"total_video_impressions"`
	TotalVideoImpressionsUnique        int64    `ch:"total_video_impressions_unique" json:"total_video_impressions_unique"`
	TotalVideoImpressionsPaidUnique    int64    `ch:"total_video_impressions_paid_unique" json:"total_video_impressions_paid_unique"`
	TotalVideoImpressionsPaid          int64    `ch:"total_video_impressions_paid" json:"total_video_impressions_paid"`
	TotalVideoImpressionsOrganicUnique int64    `ch:"total_video_impressions_organic_unique" json:"total_video_impressions_organic_unique"`
	TotalVideoImpressionsOrganic       int64    `ch:"total_video_impressions_organic" json:"total_video_impressions_organic"`
	TotalVideoImpressionsViralUnique   int64    `ch:"total_video_impressions_viral_unique" json:"total_video_impressions_viral_unique"`
	TotalVideoImpressionsViral         int64    `ch:"total_video_impressions_viral" json:"total_video_impressions_viral"`
	TotalVideoImpressionsFanUnique     int64    `ch:"total_video_impressions_fan_unique" json:"total_video_impressions_fan_unique"`
	TotalVideoImpressionsFan           int64    `ch:"total_video_impressions_fan" json:"total_video_impressions_fan"`
	TotalVideoImpressionsFanPaidUnique int64    `ch:"total_video_impressions_fan_paid_unique" json:"total_video_impressions_fan_paid_unique"`
	TotalVideoImpressionsFanPaid       int64    `ch:"total_video_impressions_fan_paid" json:"total_video_impressions_fan_paid"`
	TotalVideoStoriesByActionType      []string `ch:"total_video_stories_by_action_type" json:"total_video_stories_by_action_type"`
	TotalVideoReactionsByTypeTotal     []string `ch:"total_video_reactions_by_type_total" json:"total_video_reactions_by_type_total"`
	TotalEngagement                    int64    `ch:"total_engagement" json:"total_engagement"`

	// Ads information
	TotalVideoAdBreakEarnings      float64 `ch:"total_video_ad_break_earnings" json:"total_video_ad_break_earnings"`
	TotalVideoAdBreakAdImpressions int64   `ch:"total_video_ad_break_ad_impressions" json:"total_video_ad_break_ad_impressions"`
	TotalVideoAdBreakAdCPM         int64   `ch:"total_video_ad_break_ad_cpm" json:"total_video_ad_break_ad_cpm"`
}

type MinimalPost struct {
	PageID      string `ch:"page_id" json:"page_id"`
	PostID      string `ch:"post_id" json:"post_id"`
	FullPicture string `ch:"full_picture" json:"full_picture"`
}

// ---------- New: shared payload type ----------
type FbItem struct {
	ID          string `json:"id"`
	FullPicture string `json:"full_picture"`
	Error       *struct {
		Code int    `json:"code"`
		Type string `json:"type"`
	} `json:"error"`
	Attachments *struct {
		Data []struct {
			MediaType string `json:"media_type"`
			Type      string `json:"type"`
			Target    *struct {
				ID string `json:"id"`
			} `json:"target"`
			Media *struct {
				Image *struct {
					Src           string `json:"src"`
					Width, Height int
				} `json:"image"`
			} `json:"media"`
			Subattachments *struct {
				Data []struct {
					MediaType string `json:"media_type"`
					Type      string `json:"type"`
					Target    *struct {
						ID string `json:"id"`
					} `json:"target"`
					Media *struct {
						Image *struct {
							Src           string `json:"src"`
							Width, Height int
						} `json:"image"`
					} `json:"media"`
				} `json:"data"`
			} `json:"subattachments"`
		} `json:"data"`
	} `json:"attachments"`
}
