package kafka

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// FacebookTime is a custom time type that can handle Facebook's timestamp format
type FacebookTime struct {
	time.Time
}

// UnmarshalJSON implements the json.Unmarshaler interface to handle Facebook's timestamp format
func (ft *FacebookTime) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	timeStr := strings.Trim(string(data), "\"")
	if timeStr == "" || timeStr == "null" {
		return nil
	}

	// Facebook returns timestamps like "2025-03-13T06:37:33+0000"
	// We need to convert the "+0000" format to "+00:00" for Go to parse it
	if strings.HasSuffix(timeStr, "+0000") {
		timeStr = strings.Replace(timeStr, "+0000", "+00:00", 1)
	} else if strings.HasSuffix(timeStr, "-0000") {
		timeStr = strings.Replace(timeStr, "-0000", "+00:00", 1)
	}

	// Try parsing with the modified format
	parsedTime, err := time.Parse("2006-01-02T15:04:05-07:00", timeStr)
	if err != nil {
		// If that fails, try with standard RFC3339 format
		parsedTime, err = time.Parse(time.RFC3339, timeStr)
		if err != nil {
			return fmt.Errorf("FacebookTime.UnmarshalJSON: failed to parse Facebook timestamp '%s': %w", string(data), err)
		}
	}

	ft.Time = parsedTime
	return nil
}

// MarshalJSON implements the json.Marshaler interface
func (ft FacebookTime) MarshalJSON() ([]byte, error) {
	if ft.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(ft.Time.Format(time.RFC3339))
}

// RawFacebookPost represents the raw Facebook post data returned from the Graph API
// This structure matches the comprehensive fields defined in the Facebook client
type RawFacebookPost struct {
	ID      string `json:"id"`
	Message string `json:"message"`

	From *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"from"`
	CreatedTime    FacebookTime `json:"created_time"`
	UpdatedTime    FacebookTime `json:"updated_time"`
	PermalinkURL   string       `json:"permalink_url"`
	FullPicture    string       `json:"full_picture"`
	StatusType     string       `json:"status_type"`
	PostClicks     int64        `json:"post_clicks"`
	PostLinkClicks int64        `json:"post_link_clicks"`
	AdminCreator   *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"admin_creator"`
	ParentID string `json:"parent_id"`

	// Message tags for mentions and hashtags
	MessageTags []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Offset int    `json:"offset"`
		Length int    `json:"length"`
	} `json:"message_tags"`

	// Attachments with comprehensive media info
	Attachments *struct {
		Data []struct {
			Type        string `json:"type"`
			MediaType   string `json:"media_type"`
			Caption     string `json:"caption"`
			Description string `json:"description"`
			Link        string `json:"link"`
			Target      *struct {
				ID string `json:"id"`
			} `json:"target"`
			Media *struct {
				Src    string `json:"src,omitempty"`
				Source string `json:"source"`
				Image  *struct {
					Height int    `json:"height"`
					Width  int    `json:"width"`
					Src    string `json:"src"`
					Source string `json:"source"`
				} `json:"image"`
			} `json:"media"`
			Subattachments *struct {
				Data []struct {
					Type      string `json:"type"`
					MediaType string `json:"media_type"`
					Media     *struct {
						Src    string `json:"src"`
						Source string `json:"source"`
						Image  *struct {
							Height int    `json:"height"`
							Width  int    `json:"width"`
							Src    string `json:"src"`
							Source string `json:"source"`
						} `json:"image"`
					} `json:"media"`
				} `json:"data"`
			} `json:"subattachments"`
		} `json:"data"`
	} `json:"attachments"`

	// Child attachments for carousel posts
	ChildAttachments []struct {
		Type        string `json:"type"`
		MediaType   string `json:"media_type"`
		Description string `json:"description"`
		Media       *struct {
			Source string `json:"source"`
			Image  *struct {
				Height int    `json:"height"`
				Width  int    `json:"width"`
				Source string `json:"source"`
			} `json:"image"`
		} `json:"media"`
	} `json:"child_attachments"`

	// Shares count
	Shares *struct {
		Count int `json:"count"`
	} `json:"shares"`

	// Reaction counts by type
	Total *struct {
		Summary *struct {
			TotalCount int `json:"total_count"`
		} `json:"summary"`
	} `json:"total"`

	Like *struct {
		Summary *struct {
			TotalCount int `json:"total_count"`
		} `json:"summary"`
	} `json:"like"`

	Love *struct {
		Summary *struct {
			TotalCount int `json:"total_count"`
		} `json:"summary"`
	} `json:"love"`

	Haha *struct {
		Summary *struct {
			TotalCount int `json:"total_count"`
		} `json:"summary"`
	} `json:"haha"`

	Wow *struct {
		Summary *struct {
			TotalCount int `json:"total_count"`
		} `json:"summary"`
	} `json:"wow"`

	Sad *struct {
		Summary *struct {
			TotalCount int `json:"total_count"`
		} `json:"summary"`
	} `json:"sad"`

	Angry *struct {
		Summary *struct {
			TotalCount int `json:"total_count"`
		} `json:"summary"`
	} `json:"angry"`

	Thankful *struct {
		Summary *struct {
			TotalCount int `json:"total_count"`
		} `json:"summary"`
	} `json:"thankful"`

	// Comments with summary
	Comments *struct {
		Summary *struct {
			TotalCount int  `json:"total_count"`
			CanComment bool `json:"can_comment"`
		} `json:"summary"`
	} `json:"comments"`

	// Post insights/metrics
	Insights *struct {
		Data []struct {
			Name        string `json:"name"`
			Period      string `json:"period"`
			Title       string `json:"title"`
			Description string `json:"description"`
			Values      []struct {
				Value   int    `json:"value"`
				EndTime string `json:"end_time"`
			} `json:"values"`
		} `json:"data"`
	} `json:"insights"`

	// ---- NEW FIELDS ----

	PostMediaViewByAdd *struct {
		Data []struct {
			Name        string `json:"name"`
			Period      string `json:"period"`
			Title       string `json:"title"`
			Description string `json:"description"`
			ID          string `json:"id"`
			Values      []struct {
				Value     int    `json:"value"`
				IsFromAds string `json:"is_from_ads,omitempty"`
			} `json:"values"`
		} `json:"data"`
		Paging *struct {
			Previous string `json:"previous"`
			Next     string `json:"next"`
		} `json:"paging"`
	} `json:"post_media_view_by_add"`

	PostMediaViewByFollowers *struct {
		Data []struct {
			Name        string `json:"name"`
			Period      string `json:"period"`
			Title       string `json:"title"`
			Description string `json:"description"`
			ID          string `json:"id"`
			Values      []struct {
				Value     int    `json:"value"`
				IsFromAds string `json:"is_from_ads,omitempty"`
			} `json:"values"`
		} `json:"data"`
		Paging *struct {
			Previous string `json:"previous"`
			Next     string `json:"next"`
		} `json:"paging"`
	} `json:"post_media_view_by_followers"`
}

// ParsedFacebookPost represents the parsed Facebook post data for analytics
type ParsedFacebookPost struct {
	PageName                     string    `json:"page_name"`
	PageID                       string    `json:"page_id"`
	MediaType                    string    `json:"media_type"`
	PostID                       string    `json:"post_id"`
	Permalink                    string    `json:"permalink"`
	StatusType                   string    `json:"status_type"`
	VideoID                      string    `json:"video_id"`
	Category                     string    `json:"category"`
	PublishedBy                  string    `json:"published_by"`
	PublishedByURL               string    `json:"published_by_url"`
	SharedFromName               string    `json:"shared_from_name"`
	SharedFromID                 string    `json:"shared_from_id"`
	SharedFromLink               string    `json:"shared_from_link"`
	Like                         int32     `json:"like"`
	Love                         int32     `json:"love"`
	Haha                         int32     `json:"haha"`
	Wow                          int32     `json:"wow"`
	Sad                          int32     `json:"sad"`
	Angry                        int32     `json:"angry"`
	Thankful                     int32     `json:"thankful"`
	Total                        int64     `json:"total"`
	Shares                       int32     `json:"shares"`
	Comments                     int32     `json:"comments"`
	PostClicks                   int64     `json:"post_clicks"`
	PostClicksUnique             int64     `json:"post_clicks_unique"`
	TotalEngagement              int64     `json:"total_engagement"`
	PostEngaged                  int64     `json:"post_engaged"`
	PostEngagedUsers             int64     `json:"post_engaged_users"`
	DayOfWeek                    string    `json:"day_of_week"`
	HourOfDay                    int32     `json:"hour_of_day"`
	CreatedTime                  time.Time `json:"created_time"`
	UpdatedTime                  time.Time `json:"updated_time"`
	SavingTime                   time.Time `json:"saving_time"`
	MessageTags                  []string  `json:"message_tags"`
	PostMetadata                 string    `json:"post_metadata"`
	Caption                      string    `json:"caption"`
	Description                  string    `json:"description"`
	FullPicture                  string    `json:"full_picture"`
	Link                         string    `json:"link"`
	PostImpressions              int64     `json:"post_impressions"`
	PostMediaView                int64     `json:"post_media_view"`
	PostMediaViewAds             int64     `json:"post_media_view_ads"`
	PostImpressionsUnique        int64     `json:"post_impressions_unique"`
	PostImpressionsPaid          int64     `json:"post_impressions_paid"`
	PostImpressionsPaidUnique    int64     `json:"post_impressions_paid_unique"`
	PostImpressionsOrganic       int64     `json:"post_impressions_organic"`
	PostImpressionsOrganicUnique int64     `json:"post_impressions_organic_unique"`
	PostImpressionsViral         int64     `json:"post_impressions_viral"`
	PostImpressionsViralUnique   int64     `json:"post_impressions_viral_unique"`
	PostVideoViews               int64     `json:"post_video_views"`
	TotalImpressions             int64     `json:"total_impressions"`
	PostVideoViewTime            int64     `json:"post_video_view_time"`
	PostVideoPlayTime            int64     `json:"post_video_play_time"`
	PostNegativeFeedback         int64     `json:"post_negative_feedback"`
	PostNegativeFeedbackUnique   int64     `json:"post_negative_feedback_unique"`
	PostEngagementType           string    `json:"post_engagement_type"`
}

// ParsedFacebookMediaAsset represents the parsed Facebook media asset data
type ParsedFacebookMediaAsset struct {
	PageID       string    `json:"page_id"`
	MediaID      string    `json:"media_id"`
	PostID       string    `json:"post_id"`
	AssetType    string    `json:"asset_type"`
	Link         string    `json:"link"`
	CallToAction string    `json:"call_to_action"`
	CTAType      string    `json:"cta_type"`
	Caption      string    `json:"caption"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	InsertedAt   time.Time `json:"inserted_at"`
}

// RawFacebookVideo represents the raw Facebook video data from the Graph API
type RawFacebookVideo struct {
	ID          string `json:"id"`
	PostID      string `json:"post_id"`
	Message     string `json:"message"`
	MessageTags []struct {
		ID     string `json:"id"`
		Name   string `json:"name"`
		Type   string `json:"type"`
		Offset int    `json:"offset"`
		Length int    `json:"length"`
	} `json:"message_tags"`
	Picture       string         `json:"picture"` // quick single-size thumbnail
	Thumbnails    ThumbnailsData `json:"thumbnails"`
	Description   string         `json:"description"`
	CreatedTime   FacebookTime   `json:"created_time"`
	UpdatedTime   FacebookTime   `json:"updated_time"`
	PermalinkURL  string         `json:"permalink_url"`
	VideoInsights struct {
		Data []struct {
			Name   string `json:"name"`
			Period string `json:"period"`
			Values []struct {
				Value   interface{} `json:"value"`
				EndTime string      `json:"end_time"`
			} `json:"values"`
			Title       string `json:"title"`
			Description string `json:"description"`
		} `json:"data"`
		Paging struct {
			Previous string `json:"previous"`
			Next     string `json:"next"`
		} `json:"paging"`
	} `json:"video_insights"`
	//FaceBookVideosPostInsights *RawFacebookPost `json:"facebook_video_post_insights"`
}

type Thumbnail struct {
	URI         string `json:"uri"`
	IsPreferred bool   `json:"is_preferred"`
	ID          string `json:"id"`
}

type ThumbnailsData struct {
	Data []Thumbnail `json:"data"`
}

// ParsedFacebookVideoInsights represents the parsed Facebook video insights data
type ParsedFacebookVideoInsights struct {
	PageID          string    `json:"page_id"`
	PostID          string    `json:"post_id"`
	VideoID         string    `json:"video_id"`
	CreatedTime     time.Time `json:"created_time"`
	UpdatedTime     time.Time `json:"updated_time"`
	SavingTime      time.Time `json:"saving_time"`
	TotalEngagement int64     `json:"total_engagement"`

	// Existing fields
	PostVideoAvgTimeWatched int64 `json:"post_video_avg_time_watched"`
	BlueReelsPlayCount      int64 `json:"blue_reels_play_count"`
	PostImpressionsUnique   int64 `json:"post_impressions_unique"`
	PostVideoViewTime       int64 `json:"post_video_view_time"`

	// Followers and Views stats
	TotalVideoFollowers               int64    `json:"total_video_followers"` // NEWLY ADDED - matches ClickHouse model
	TotalVideoViews                   int64    `json:"total_video_views"`
	TotalVideoViewsUnique             int64    `json:"total_video_views_unique"`
	TotalVideoViewsAutoplayed         int64    `json:"total_video_views_autoplayed"`
	TotalVideoViewsClickedToPlay      int64    `json:"total_video_views_clicked_to_play"` // NEWLY ADDED - matches ClickHouse model
	TotalVideoViewsOrganic            int64    `json:"total_video_views_organic"`
	TotalVideoViewsOrganicUnique      int64    `json:"total_video_views_organic_unique"`
	TotalVideoViewsPaid               int64    `json:"total_video_views_paid"`
	TotalVideoViewsPaidUnique         int64    `json:"total_video_views_paid_unique"`
	TotalVideoViewsSoundOn            int64    `json:"total_video_views_sound_on"`
	TotalVideoViewsByDistributionType []string `json:"total_video_views_by_distribution_type"` // FIXED TYPE - changed from string to []string

	// View time demographics - NEWLY ADDED for ClickHouse compatibility
	TotalVideoViewTimeByDistributionType   []string `json:"total_video_view_time_by_distribution_type"`     // FIXED TYPE - changed from string to []string
	TotalVideoViewTimeByCountryID          []string `json:"total_video_view_time_by_country_id"`            // NEWLY ADDED - matches ClickHouse model
	TotalVideoViewTimeByRegionID           []string `json:"total_video_view_time_by_region_id"`             // FIXED TYPE - changed from string to []string
	TotalVideoViewTimeByAgeBucketAndGender []string `json:"total_video_view_time_by_age_bucket_and_gender"` // FIXED TYPE - changed from string to []string

	// Views stats by watch time
	TotalVideoPlayCount                             int64    `json:"total_video_play_count"`       // NEWLY ADDED - matches ClickHouse model
	TotalVideoConsumptionRate                       float64  `json:"total_video_consumption_rate"` // NEWLY ADDED - matches ClickHouse model
	TotalVideoCompleteViews                         int64    `json:"total_video_complete_views"`
	TotalVideoCompleteViewsUnique                   int64    `json:"total_video_complete_views_unique"`
	TotalVideoCompleteViewsAutoplayed               int64    `json:"total_video_complete_views_autoplayed"` // FIXED NAME - removed underscore to match ClickHouse
	TotalVideoCompleteViewsClickedToPlay            int64    `json:"total_video_complete_views_clicked_to_play"`
	TotalVideoCompleteViewsOrganic                  int64    `json:"total_video_complete_views_organic"`
	TotalVideoCompleteViewsOrganicUnique            int64    `json:"total_video_complete_views_organic_unique"`
	TotalVideoCompleteViewsPaid                     int64    `json:"total_video_complete_views_paid"`
	TotalVideoCompleteViewsPaidUnique               int64    `json:"total_video_complete_views_paid_unique"`
	VideoAsset60sVideoViewTotalCountByIsMonetizable []string `json:"video_asset_60s_video_view_total_count_by_is_monetizable"` // NEWLY ADDED - matches ClickHouse model
	TotalVideo15minExcludesShorterViews             int64    `json:"total_video_15min_excludes_shorter_views"`                 // NEWLY ADDED - matches ClickHouse model
	TotalVideo15minExcludesShorterViewsUnique       int64    `json:"total_video_15min_excludes_shorter_views_unique"`          // NEWLY ADDED - matches ClickHouse model
	TotalVideo60sExcludesShorterViews               int64    `json:"total_video_60s_excludes_shorter_views"`
	TotalVideo30sViews                              int64    `json:"total_video_30s_views"`                 // NEWLY ADDED - matches ClickHouse model
	TotalVideo30sViewsUnique                        int64    `json:"total_video_30s_views_unique"`          // NEWLY ADDED - matches ClickHouse model
	TotalVideo30sViewsAutoplayed                    int64    `json:"total_video_30s_views_autoplayed"`      // NEWLY ADDED - matches ClickHouse model
	TotalVideo30sViewsClickedToPlay                 int64    `json:"total_video_30s_views_clicked_to_play"` // NEWLY ADDED - matches ClickHouse model
	TotalVideo30sViewsOrganic                       int64    `json:"total_video_30s_views_organic"`         // NEWLY ADDED - matches ClickHouse model
	TotalVideo30sViewsPaid                          int64    `json:"total_video_30s_views_paid"`            // NEWLY ADDED - matches ClickHouse model
	TotalVideo30sViewsSoundOn                       int64    `json:"total_video_30s_views_sound_on"`        // NEWLY ADDED - matches ClickHouse model
	TotalVideo10sViews                              int64    `json:"total_video_10s_views"`
	TotalVideo10sViewsUnique                        int64    `json:"total_video_10s_views_unique"`
	TotalVideo10sViewsAutoplayed                    int64    `json:"total_video_10s_views_autoplayed"` // FIXED NAME - removed underscore to match ClickHouse
	TotalVideo10sViewsClickedToPlay                 int64    `json:"total_video_10s_views_clicked_to_play"`
	TotalVideo10sViewsOrganic                       int64    `json:"total_video_10s_views_organic"`
	TotalVideo10sViewsPaid                          int64    `json:"total_video_10s_views_paid"`
	TotalVideo10sViewsSoundOn                       int64    `json:"total_video_10s_views_sound_on"`
	TotalVideo15sViews                              int64    `json:"total_video_15s_views"`
	TotalVideoAvgTimeWatched                        int64    `json:"total_video_avg_time_watched"` // FIXED TYPE - changed from float64 to int64 to match ClickHouse
	TotalVideoViewTotalTime                         int64    `json:"total_video_view_total_time"`
	TotalVideoViewTotalTimeOrganic                  int64    `json:"total_video_view_total_time_organic"`
	TotalVideoViewTotalTimePaid                     int64    `json:"total_video_view_total_time_paid"`

	// Audience retention info - NEWLY ADDED for ClickHouse compatibility
	TotalVideoRetentionGraphAutoplayed    []string `json:"total_video_retention_graph_autoplayed"`      // FIXED TYPE - changed from string to []string
	TotalVideoRetentionGraphClickedToPlay []string `json:"total_video_retention_graph_clicked_to_play"` // FIXED TYPE - changed from string to []string
	TotalVideoRetentionGraphGenderMale    []string `json:"total_video_retention_graph_gender_male"`     // NEWLY ADDED - matches ClickHouse model
	TotalVideoRetentionGraphGenderFemale  []string `json:"total_video_retention_graph_gender_female"`   // NEWLY ADDED - matches ClickHouse model

	// Impressions
	TotalVideoImpressions              int64    `json:"total_video_impressions"`
	TotalVideoImpressionsUnique        int64    `json:"total_video_impressions_unique"`
	TotalVideoImpressionsPaidUnique    int64    `json:"total_video_impressions_paid_unique"`
	TotalVideoImpressionsPaid          int64    `json:"total_video_impressions_paid"`
	TotalVideoImpressionsOrganicUnique int64    `json:"total_video_impressions_organic_unique"`
	TotalVideoImpressionsOrganic       int64    `json:"total_video_impressions_organic"`
	TotalVideoImpressionsViralUnique   int64    `json:"total_video_impressions_viral_unique"`
	TotalVideoImpressionsViral         int64    `json:"total_video_impressions_viral"`
	TotalVideoImpressionsFanUnique     int64    `json:"total_video_impressions_fan_unique"`
	TotalVideoImpressionsFan           int64    `json:"total_video_impressions_fan"`
	TotalVideoImpressionsFanPaidUnique int64    `json:"total_video_impressions_fan_paid_unique"`
	TotalVideoImpressionsFanPaid       int64    `json:"total_video_impressions_fan_paid"`
	TotalVideoStoriesByActionType      []string `json:"total_video_stories_by_action_type"` // FIXED TYPE - changed from string to []string

	TotalVideoReactionsByTypeTotal []string `json:"total_video_reactions_by_type_total"` // FIXED TYPE - changed from string to []string

	// Legacy fields maintained for backward compatibility
	TotalVideoRetentionGraph    string `json:"total_video_retention_graph"` // DEPRECATED - keeping for backward compatibility
	TotalVideoViewTotalTimeLive int64  `json:"total_video_view_total_time_live"`
	TotalVideoViewsLive         int64  `json:"total_video_views_live"`

	// Ads information - NEWLY ADDED for ClickHouse compatibility
	TotalVideoAdBreakEarnings      float64 `json:"total_video_ad_break_earnings"`       // NEWLY ADDED - matches ClickHouse model
	TotalVideoAdBreakAdImpressions int64   `json:"total_video_ad_break_ad_impressions"` // NEWLY ADDED - matches ClickHouse model
	TotalVideoAdBreakAdCPM         float64 `json:"total_video_ad_break_ad_cpm"`         // NEWLY ADDED - matches ClickHouse model
}

// ParsedFacebookReelsInsights represents focused reels metrics extracted from video insights
// This matches the Python FacebookReelsInsights model with only essential reels fields
type ParsedFacebookReelsInsights struct {
	PageID               string    `json:"page_id"`
	PostID               string    `json:"post_id"`
	AverageTimeWatched   int64     `json:"average_time_watched"`     // post_video_avg_time_watched
	TotalTimeWatchedInMs int64     `json:"total_time_watched_in_ms"` // post_video_view_time
	PlayCount            int64     `json:"play_count"`               // blue_reels_play_count
	ImpressionsUnique    int64     `json:"impressions_unique"`       // post_impressions_unique
	ReelFollowers        int64     `json:"reel_followers"`           // post_video_followers
	CreatedAt            time.Time `json:"created_at"`
	SavingTime           time.Time `json:"saving_time"`
}

// RawFacebookInsights represents the raw insights data from Facebook Graph API
type RawFacebookInsights struct {
	PageID      string                 `json:"page_id"`
	WorkspaceID string                 `json:"workspace_id"`
	Data        []FacebookInsightData  `json:"data"`
	Paging      map[string]interface{} `json:"paging,omitempty"`
	SavingTime  time.Time              `json:"saving_time"`
}

// FacebookInsightData represents individual insight metric from Facebook API
type FacebookInsightData struct {
	Name        string                 `json:"name"`
	Period      string                 `json:"period"`
	Values      []FacebookInsightValue `json:"values"`
	Title       string                 `json:"title,omitempty"`
	Description string                 `json:"description,omitempty"`
	ID          string                 `json:"id,omitempty"`
}

// FacebookInsightValue represents a single insight value with date range
type FacebookInsightValue struct {
	Value   interface{} `json:"value"`
	EndTime string      `json:"end_time"`
}

// ParsedFacebookInsights represents processed Facebook page insights matching Python FacebookInsightsModel
type ParsedFacebookInsights struct {
	// IDs and basic info
	HashID       string `json:"hash_id"`
	PageID       string `json:"page_id"`
	WorkspaceID  string `json:"workspace_id"`
	PageCategory string `json:"page_category"`

	// Date information
	DayOfWeek   string    `json:"day_of_week"`
	Year        int       `json:"year"`
	Month       int       `json:"month"`
	CreatedTime time.Time `json:"created_time"` // The actual date the data belongs to (from API end_time)
	SavingTime  time.Time `json:"saving_time"`  // When the record was saved (today's date)

	// Fan information and demographics
	PageFans          int64    `json:"page_fans"`
	PageFansCity      []string `json:"page_fans_city"`
	PageFansCountry   []string `json:"page_fans_country"`
	PageFansLocale    []string `json:"page_fans_locale"`
	PageFansAge       []string `json:"page_fans_age"`
	PageFansGender    []string `json:"page_fans_gender"`
	PageFansGenderAge []string `json:"page_fans_gender_age"`

	// Page metrics
	PageFollows int64 `json:"page_follows"`
	PageViews   int64 `json:"page_views"`

	// Fan adds/removes
	PageFanAddsUnique              int64    `json:"page_fan_adds_unique"`
	PageFanRemovesUnique           int64    `json:"page_fan_removes_unique"`
	PageFanAddsByPaidNonPaidUnique []string `json:"page_fan_adds_by_paid_non_paid_unique"`
	PageFansByLikeSourceUnique     []string `json:"page_fans_by_like_source_unique"`
	PageFansByUnlikeSourceUnique   []string `json:"page_fans_by_unlike_source_unique"`
	PageFansByLike                 int64    `json:"page_fans_by_like"`
	PageFansByUnlike               int64    `json:"page_fans_by_unlike"`

	// Engagement metrics
	PageTotalActions    int64 `json:"page_total_actions"`
	PagePostEngagements int64 `json:"page_post_engagements"`

	// Impression metrics
	PageImpressions        int64 `json:"page_impressions"`
	PageImpressionsUnique  int64 `json:"page_impressions_unique"`
	PageMediaView          int64 `json:"page_media_view"`
	PageImpressionsOrganic int64 `json:"page_impressions_organic"`
	PageImpressionsPaid    int64 `json:"page_impressions_paid"`

	// Video metrics
	PageVideoViews            int64 `json:"page_video_views"`
	PageVideoViewsPaid        int64 `json:"page_video_views_paid"`
	PageVideoViewsOrganic     int64 `json:"page_video_views_organic"`
	PageVideoViewsAutoplayed  int64 `json:"page_video_views_autoplayed"`
	PageVideoViewsClickToPlay int64 `json:"page_video_views_click_to_play"`
	PageVideoRepeatViews      int64 `json:"page_video_repeat_views"`

	// Feedback metrics
	PageNegativeFeedback       int64    `json:"page_negative_feedback"`
	PagePositiveFeedback       int64    `json:"page_positive_feedback"`
	PageNegativeFeedbackByType []string `json:"page_negative_feedback_by_type"`
	PagePositiveFeedbackByType []string `json:"page_positive_feedback_by_type"`

	// Activity metrics
	PageFansOnline []string  `json:"page_fans_online"`
	ActiveUsers    int64     `json:"active_users"`
	PrimeTime      time.Time `json:"prime_time"`

	// Sentiment and engagement
	PositiveSentiment int64 `json:"positive_sentiment"`
	NegativeSentiment int64 `json:"negative_sentiment"`

	// Content metrics
	PostsCount        int64    `json:"posts_count"`
	LikesCount        int64    `json:"likes_count"`
	TalkingAboutCount int64    `json:"talking_about_count"`
	TypeCount         []string `json:"type_count"`    // link, photo, video counts
	MessageCount      []string `json:"message_count"` // sent, received counts

	// Reaction metrics (positive sentiment components)
	PageActionsPostReactionsLikeTotal  int64 `json:"page_actions_post_reactions_like_total"`
	PageActionsPostReactionsLoveTotal  int64 `json:"page_actions_post_reactions_love_total"`
	PageActionsPostReactionsAngerTotal int64 `json:"page_actions_post_reactions_anger_total"`
}

// FacebookAccountWorkOrder represents the structure of messages from the Kafka topic
type FacebookAccountWorkOrder struct {
	ID              string `json:"id"`
	FacebookID      string `json:"facebook_id"`
	Type            string `json:"type"`              // e.g., "Page", "Group"
	AccessToken     string `json:"access_token"`      // From ExtraData
	WorkspaceID     string `json:"workspace_id"`      // From ExtraData
	LongAccessToken string `json:"long_access_token"` // The page's long-lived access token
	SyncType        string `json:"sync_type"`         // e.g., "incremental", "full_sync"
}

// FacebookBatchWorkOrder represents a batch of Facebook accounts to process.
// The scheduler produces batch messages to reduce Kafka overhead.
// The fetcher unpacks batches and distributes accounts to worker pools.
type FacebookBatchWorkOrder struct {
	BatchID   string                     `json:"batch_id"`   // Unique batch identifier (UUID)
	SyncType  string                     `json:"sync_type"`  // "incremental" | "full_sync"
	Accounts  []FacebookAccountWorkOrder `json:"accounts"`   // List of accounts in this batch (max 200)
	CreatedAt time.Time                  `json:"created_at"` // Batch creation timestamp
}
