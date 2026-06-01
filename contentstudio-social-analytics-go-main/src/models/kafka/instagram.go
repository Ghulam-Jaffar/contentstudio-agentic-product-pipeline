package kafka

import (
	"time"
)

// InstagramChild represents a child asset for an Instagram media item (e.g., images in a carousel)
// Only the subset of fields required for analytics are included. Extend as needed.
type InstagramChild struct {
	MediaType    string `json:"media_type"`
	MediaURL     string `json:"media_url"`
	ThumbnailURL string `json:"thumbnail_url"`
}

// RawInstagramMedia mirrors the structure returned by the Instagram Graph API for a media item.
// This struct is intentionally kept minimal; additional fields can be added as the pipeline evolves.
type RawInstagramMedia struct {
	ID               string `json:"id"`
	CommentsCount    int    `json:"comments_count"`
	ThumbnailURL     string `json:"thumbnail_url"`
	Caption          string `json:"caption"`
	Username         string `json:"username"`
	LikeCount        int    `json:"like_count"`
	MediaType        string `json:"media_type"`
	MediaProductType string `json:"media_product_type"`
	MediaURL         string `json:"media_url"`
	Timestamp        string `json:"timestamp"`
	Children         struct {
		Data []InstagramChild `json:"data"`
	} `json:"children"`
	Permalink string `json:"permalink"`
}

// RawInstagramAccountResponse is the top-level response when requesting profile information along with media/stories.
type RawInstagramAccountResponse struct {
	Name              string `json:"name"`
	ProfilePictureURL string `json:"profile_picture_url"`
	Media             struct {
		Data   []RawInstagramMedia `json:"data"`
		Paging struct {
			Next string `json:"next"`
		} `json:"paging"`
	} `json:"media"`
}

// InstagramInsightValue represents a single value entry for an Instagram insight metric.
// Note: Both simple daily metrics ("values") and lifetime metrics ("total_value") are supported.
type InstagramInsightValue struct {
	Value   int    `json:"value"`
	EndTime string `json:"end_time,omitempty"`
}

// InstagramInsightBreakdownResult mirrors the structure in demographic insights breakdowns.
type InstagramInsightBreakdownResult struct {
	DimensionValues []string `json:"dimension_values"`
	Value           int      `json:"value"`
}

type InstagramInsightBreakdown struct {
	Results []InstagramInsightBreakdownResult `json:"results"`
}

// RawInstagramInsightData is a single metric block returned by the Graph API.
type RawInstagramInsightData struct {
	Name       string                  `json:"name"`
	Period     string                  `json:"period"`
	Values     []InstagramInsightValue `json:"values,omitempty"`
	TotalValue struct {
		Value      int                         `json:"value"`
		Breakdowns []InstagramInsightBreakdown `json:"breakdowns,omitempty"`
	} `json:"total_value,omitempty"`
}

// RawInstagramInsightsResponse wraps the list of insight data entries.
// It corresponds directly to the Graph API response.
type RawInstagramInsightsResponse struct {
	Data []RawInstagramInsightData `json:"data"`
}

// RawInstagramMediaInsights represents insights data for a specific media item.
type RawInstagramMediaInsights struct {
	Data []struct {
		Name   string `json:"name"`
		Period string `json:"period"`
		Values []struct {
			Value interface{} `json:"value"`
		} `json:"values"`
		Title       string `json:"title"`
		Description string `json:"description"`
		ID          string `json:"id"`
	} `json:"data"`
}

//type RawInstagramMediaInsights struct {
//	Data []struct {
//		Name   string `json:"name"`
//		Period string `json:"period,omitempty"`
//		Values []struct {
//			Value interface{} `json:"value"`
//		} `json:"values,omitempty"`
//		TotalValue struct {
//			Breakdowns []struct {
//				Results []struct {
//					DimensionValues []string `json:"dimension_values"`
//					Value           int      `json:"value"`
//				} `json:"results"`
//			} `json:"breakdowns,omitempty"`
//		} `json:"total_value,omitempty"`
//	} `json:"data"`
//}

// RawInstagramDemographics represents demographic insights from the API.
type RawInstagramDemographics struct {
	Data []struct {
		Name       string `json:"name"`
		Period     string `json:"period"`
		TotalValue struct {
			Value      interface{} `json:"value"`
			Breakdowns []struct {
				Results []struct {
					DimensionValues []string `json:"dimension_values"`
					Value           int      `json:"value"`
				} `json:"results"`
			} `json:"breakdowns,omitempty"`
		} `json:"total_value"`
	} `json:"data"`
}

// ParsedInstagramPost represents the parsed Instagram post data for analytics.
// This mirrors the structure of the InstagramPosts model in the Python implementation.
type ParsedInstagramPost struct {
	InstagramID         string    `json:"instagram_id"`
	MediaID             string    `json:"media_id"`
	Username            string    `json:"username"`
	Name                string    `json:"name"`
	ProfilePictureURL   string    `json:"profile_picture_url"`
	Permalink           string    `json:"permalink"`
	LikeCount           int64     `json:"like_count"`
	CommentsCount       int64     `json:"comments_count"`
	Engagement          int64     `json:"engagement"`
	Impressions         int64     `json:"impressions"`
	Views               int64     `json:"views"`
	Reach               int64     `json:"reach"`
	Saved               int64     `json:"saved"`
	VideoViews          int64     `json:"video_views"`
	Shares              int64     `json:"shares"`
	ReelsAvgWatchTime   int64     `json:"reels_avg_watch_time"`
	ReelsTotalWatchTime int64     `json:"reels_total_watch_time"`
	Exits               int64     `json:"exits"`
	Replies             int64     `json:"replies"`
	TapsForward         int64     `json:"taps_forward"`
	TapsBack            int64     `json:"taps_back"`
	ChildAssetsType     []string  `json:"child_assets_type"`
	Caption             string    `json:"caption"`
	MediaType           string    `json:"media_type"`
	EntityType          string    `json:"entity_type"`
	MediaURL            []string  `json:"media_url"`
	VideoURL            []string  `json:"video_url"`
	Hashtags            []string  `json:"hashtags"`
	DayOfWeek           string    `json:"day_of_week"`
	HourOfDay           int64     `json:"hour_of_day"`
	Year                int64     `json:"year"`
	Month               int64     `json:"month"`
	Timestamp           int64     `json:"timestamp"`
	StoredEventAt       time.Time `json:"stored_event_at"`
	PostCreatedAt       time.Time `json:"post_created_at"`
}

// ParsedInstagramInsight represents the parsed Instagram insights data for analytics.
// This mirrors the structure of the InstagramInsights model in the Python implementation.
type ParsedInstagramInsight struct {
	InstagramID                   string            `json:"instagram_id"`
	RecordID                      string            `json:"record_id"`
	Name                          string            `json:"name"`
	Username                      string            `json:"username"`
	ProfilePictureURL             string            `json:"profile_picture_url"`
	FollowsCount                  int64             `json:"follows_count"`
	FollowersCount                int64             `json:"followers_count"`
	FollowerCount                 int64             `json:"follower_count"`
	MediaCount                    int64             `json:"media_count"`
	Tags                          int64             `json:"tags"`
	Impressions                   int64             `json:"impressions"`
	ProfileViews                  int64             `json:"profile_views"`
	Shares                        int64             `json:"shares"`
	AccountsEngaged               int64             `json:"accounts_engaged"`
	Engagement                    int64             `json:"engagement"`
	Reach                         int64             `json:"reach"`
	Views                         int64             `json:"views"`
	Saves                         int64             `json:"saves"`
	Likes                         int64             `json:"likes"`
	Comments                      int64             `json:"comments"`
	AudienceAge                   []string          `json:"audience_age"`
	AudienceGender                []string          `json:"audience_gender"`
	AudienceGenderAge             []string          `json:"audience_gender_age"`
	AudienceLocale                []string          `json:"audience_locale"`
	AudienceCity                  []string          `json:"audience_city"`
	AudienceCountry               []string          `json:"audience_country"`
	AudienceAgeByEngagement       []string          `json:"audience_age_by_engagement"`
	AudienceGenderByEngagement    []string          `json:"audience_gender_by_engagement"`
	AudienceGenderAgeByEngagement []string          `json:"audience_gender_age_by_engagement"`
	AudienceCityByEngagement      []string          `json:"audience_city_by_engagement"`
	AudienceCountryByEngagement   []string          `json:"audience_country_by_engagement"`
	AudienceAgeByReach            []string          `json:"audience_age_by_reach"`
	AudienceGenderByReach         []string          `json:"audience_gender_by_reach"`
	AudienceGenderAgeByReach      []string          `json:"audience_gender_age_by_reach"`
	AudienceCityByReach           []string          `json:"audience_city_by_reach"`
	AudienceCountryByReach        []string          `json:"audience_country_by_reach"`
	OnlineFollowers               []string          `json:"online_followers"`
	AudienceDatetime              time.Time         `json:"audience_datetime"`
	OnlineUsersDatetime           time.Time         `json:"online_users_datetime"`
	DayOfWeek                     string            `json:"day_of_week"`
	Year                          int64             `json:"year"`
	Month                         int64             `json:"month"`
	CreatedTime                   time.Time         `json:"created_time"`    // The actual date the data belongs to
	UpdatedTime                   time.Time         `json:"updated_time"`    // When the record was last modified
	Metadata                      map[string]string `json:"metadata"`        // Key-value pairs for additional info (source, method, etc.)
	StoredEventAt                 time.Time         `json:"stored_event_at"` // When the record was first saved
}

// InstagramAccountWorkOrder mirrors the structure expected from work-order-instagram topic
// It matches the Facebook variant but adapted for Instagram.
type InstagramAccountWorkOrder struct {
	ID                    string `json:"id"`
	InstagramID           string `json:"instagram_id"`
	AccessToken           string `json:"access_token"`
	WorkspaceID           string `json:"workspace_id"`
	SyncType              string `json:"sync_type"` // e.g., "incremental", "full_sync"
	ConnectedViaInstagram bool   `json:"connected_via_instagram"`
}

// InstagramBatchWorkOrder represents a batch of Instagram accounts to process.
// The scheduler produces batch messages to reduce Kafka overhead.
// The fetcher unpacks batches and distributes accounts to worker pools.
type InstagramBatchWorkOrder struct {
	BatchID   string                      `json:"batch_id"`   // Unique batch identifier (UUID)
	SyncType  string                      `json:"sync_type"`  // "incremental" | "full_sync"
	Accounts  []InstagramAccountWorkOrder `json:"accounts"`   // List of accounts in this batch (max 200)
	CreatedAt time.Time                   `json:"created_at"` // Batch creation timestamp
}

// ImmediateWorkOrder represents the structure of messages from the Instagram immediate work order Kafka topic
// It mirrors Facebook version but with Instagram specific IDs.
