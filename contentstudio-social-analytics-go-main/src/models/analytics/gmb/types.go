package gmb

import (
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// --- Request types ---

// GMBRequest is the base request for all GMB analytics endpoints.
// Query params: workspace_id, gmb_id, start_date, end_date, timezone.
type GMBRequest struct {
	WorkspaceID string `json:"workspace_id"`
	GmbID       string `json:"gmb_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Timezone    string `json:"timezone"`
}

// TopPostsRequest extends GMBRequest with pagination options.
type TopPostsRequest struct {
	GMBRequest
	Limit   int    `json:"limit"`
	OrderBy string `json:"order_by"`
}

// SearchKeywordsRequest extends GMBRequest with pagination options.
type SearchKeywordsRequest struct {
	GMBRequest
	Limit int `json:"limit"`
}

// --- Response types ---

type SummaryResponse struct {
	Status   bool                       `json:"status"`
	Overview map[string]*SummaryMetrics `json:"overview"`
}

type SummaryMetrics struct {
	TotalImpressions  int64   `json:"total_impressions"`
	SearchImpressions int64   `json:"search_impressions"`
	MapsImpressions   int64   `json:"maps_impressions"`
	WebsiteClicks     int64   `json:"website_clicks"`
	CallClicks        int64   `json:"call_clicks"`
	DirectionRequests int64   `json:"direction_requests"`
	OtherActions      int64   `json:"other_actions"`
	TotalReviews      int64   `json:"total_reviews"`
	AverageRating     float64 `json:"average_rating"`
	TotalPosts        int64   `json:"total_posts"`
}

type ImpressionsResponse struct {
	Status           bool                              `json:"status"`
	Impressions      *ImpressionsData                  `json:"impressions"`
	ImpressionsRolup map[string]*ImpressionsRollupData `json:"impressions_rollup"`
}

type ImpressionsData struct {
	DesktopMaps          []int64  `json:"desktop_maps"`
	DesktopSearch        []int64  `json:"desktop_search"`
	MobileMaps           []int64  `json:"mobile_maps"`
	MobileSearch         []int64  `json:"mobile_search"`
	TotalImpressions     []int64  `json:"total_impressions"`
	DesktopMapsDaily     []int64  `json:"desktop_maps_daily"`
	DesktopSearchDaily   []int64  `json:"desktop_search_daily"`
	MobileMapsDaily      []int64  `json:"mobile_maps_daily"`
	MobileSearchDaily    []int64  `json:"mobile_search_daily"`
	TotalImpressionsDaily []int64 `json:"total_impressions_daily"`
	ShowData             int64    `json:"show_data"`
	Buckets              []string `json:"buckets"`
}

type ImpressionsRollupData struct {
	TotalImpressions  int64   `json:"total_impressions"`
	DesktopMaps       int64   `json:"desktop_maps"`
	DesktopSearch     int64   `json:"desktop_search"`
	MobileMaps        int64   `json:"mobile_maps"`
	MobileSearch      int64   `json:"mobile_search"`
	AvgImpressions    float64 `json:"avg_impressions"`
}

type ActionsResponse struct {
	Status        bool                           `json:"status"`
	Actions       *ActionsData                   `json:"actions"`
	ActionsRollup map[string]*ActionsRollupData  `json:"actions_rollup"`
}

type ActionsData struct {
	CallClicks             []int64  `json:"call_clicks"`
	WebsiteClicks          []int64  `json:"website_clicks"`
	DirectionRequests      []int64  `json:"direction_requests"`
	OtherActions           []int64  `json:"other_actions"`
	CallClicksDaily        []int64  `json:"call_clicks_daily"`
	WebsiteClicksDaily     []int64  `json:"website_clicks_daily"`
	DirectionRequestsDaily []int64  `json:"direction_requests_daily"`
	OtherActionsDaily      []int64  `json:"other_actions_daily"`
	ShowData               int64    `json:"show_data"`
	Buckets                []string `json:"buckets"`
}

type ActionsRollupData struct {
	TotalCallClicks        int64   `json:"total_call_clicks"`
	TotalWebsiteClicks     int64   `json:"total_website_clicks"`
	TotalDirectionRequests int64   `json:"total_direction_requests"`
	TotalOtherActions      int64   `json:"total_other_actions"`
	AvgActions             float64 `json:"avg_actions"`
}

type SearchKeywordsResponse struct {
	Status   bool             `json:"status"`
	Keywords []SearchKeyword  `json:"keywords"`
}

type SearchKeyword struct {
	Keyword              string `json:"keyword"`
	ImpressionsValue     int64  `json:"impressions_value"`
	ImpressionsThreshold int64  `json:"impressions_threshold"`
	KeywordMonth         string `json:"keyword_month"`
}

type TopPostsResponse struct {
	Status bool      `json:"status"`
	Posts  []TopPost `json:"posts"`
}

type TopPost struct {
	PostName        string   `json:"post_name"`
	Summary         string   `json:"summary"`
	State           string   `json:"state"`
	TopicType       string   `json:"topic_type"`
	SearchURL       string   `json:"search_url"`
	MediaNames      []string `json:"media_names"`
	MediaFormats    []string `json:"media_formats"`
	MediaGoogleURLs []string `json:"media_google_urls"`
	CreatedAt       string   `json:"created_at"`
}

type PublishingBehaviorResponse struct {
	Status              bool                    `json:"status"`
	PublishingBehaviour *PublishingBehaviorData  `json:"publishing_behaviour"`
}

type PublishingBehaviorData struct {
	Buckets    []string          `json:"buckets"`
	PostCount  []int64           `json:"post_count"`
	TopicTypes []TopicTypeCount  `json:"topic_types"`
}

type TopicTypeCount struct {
	Name  string `json:"name"`
	Count int64  `json:"count"`
}

type ReviewsResponse struct {
	Status        bool                          `json:"status"`
	Reviews       *ReviewsData                  `json:"reviews"`
	ReviewsRollup map[string]*ReviewsRollupData `json:"reviews_rollup"`
}

type ReviewsData struct {
	AvgRating        float64          `json:"avg_rating"`
	TotalReviews     int64            `json:"total_reviews"`
	StarDistribution map[string]int64 `json:"star_distribution"`
	Buckets          []string         `json:"buckets"`
	DailyReviews     []int64          `json:"daily_reviews"`
	ReviewsList      []ReviewItem     `json:"reviews_list"`
}

type ReviewItem struct {
	ReviewID                string `json:"review_id"`
	ReviewerDisplayName     string `json:"reviewer_display_name"`
	ReviewerProfilePhotoURL string `json:"reviewer_profile_photo_url"`
	StarRating              int64  `json:"star_rating"`
	Comment                 string `json:"comment"`
	ReplyComment            string `json:"reply_comment"`
	CreatedAt               string `json:"created_at"`
}

type ReviewsRollupData struct {
	TotalReviews int64   `json:"total_reviews"`
	AvgRating    float64 `json:"avg_rating"`
}

type MediaActivityResponse struct {
	Status             bool                                `json:"status"`
	MediaActivity      *MediaActivityData                  `json:"media_activity"`
	MediaActivityRollup map[string]*MediaActivityRollupData `json:"media_activity_rollup"`
}

type MediaActivityData struct {
	PhotoCount      []int64  `json:"photo_count"`
	VideoCount      []int64  `json:"video_count"`
	PhotoCountDaily []int64  `json:"photo_count_daily"`
	VideoCountDaily []int64  `json:"video_count_daily"`
	ShowData        int64    `json:"show_data"`
	Buckets         []string `json:"buckets"`
}

type MediaActivityRollupData struct {
	TotalPhotos int64   `json:"total_photos"`
	TotalVideos int64   `json:"total_videos"`
	AvgMedia    float64 `json:"avg_media"`
}

type ErrorResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

// AIInsightsRequest is the POST body for the AI insights endpoint.
type AIInsightsRequest struct {
	WorkspaceID string `json:"workspace_id"`
	GmbID       string `json:"gmb_id"`
	StartDate   string      `json:"start_date,omitempty"`
	EndDate     string      `json:"end_date,omitempty"`
	Date        interface{} `json:"date"`
	Timezone    string      `json:"timezone"`
	Type        string      `json:"type"`
	Limit       int         `json:"limit"`
	Language    string      `json:"language,omitempty"`
}

// AIInsightsResponse wraps the AI agent response.
type AIInsightsResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// --- Request helpers ---

const maxLimit = 100

func (r *GMBRequest) Validate() error {
	if r.WorkspaceID == "" {
		return httputil.NewValidationError("workspace_id is required")
	}
	if r.GmbID == "" {
		return httputil.NewValidationError("gmb_id is required")
	}
	if r.StartDate == "" {
		return httputil.NewValidationError("start_date is required")
	}
	if r.EndDate == "" {
		return httputil.NewValidationError("end_date is required")
	}
	startDate, err := time.Parse("2006-01-02", r.StartDate)
	if err != nil {
		return httputil.NewValidationError("start_date must be in YYYY-MM-DD format")
	}
	endDate, err := time.Parse("2006-01-02", r.EndDate)
	if err != nil {
		return httputil.NewValidationError("end_date must be in YYYY-MM-DD format")
	}
	if endDate.Before(startDate) {
		return httputil.NewValidationError("end_date cannot be before start_date")
	}
	if r.Timezone != "" {
		if _, err := time.LoadLocation(r.Timezone); err != nil {
			return httputil.NewValidationError("invalid timezone: " + r.Timezone)
		}
	}
	return nil
}

func (r *GMBRequest) GetAccountIDs() []string {
	if r.GmbID != "" {
		return []string{r.GmbID}
	}
	return nil
}

func (r *GMBRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

func (r *GMBRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	params, err := clickhouse.ParseStartEndDate(r.StartDate, r.EndDate, r.GetTimezone())
	if err != nil {
		return nil, err
	}
	params.AccountIDs = r.GetAccountIDs()
	return params, nil
}

func (r *TopPostsRequest) GetLimit() int {
	if r.Limit <= 0 {
		return 15
	}
	if r.Limit > maxLimit {
		return maxLimit
	}
	return r.Limit
}

var validTopPostOrderFields = map[string]bool{
	"created_at": true,
}

func (r *TopPostsRequest) GetOrderBy() string {
	if r.OrderBy == "" || !validTopPostOrderFields[r.OrderBy] {
		return "created_at"
	}
	return r.OrderBy
}

func (r *SearchKeywordsRequest) GetLimit() int {
	if r.Limit <= 0 {
		return 50
	}
	if r.Limit > maxLimit {
		return maxLimit
	}
	return r.Limit
}
