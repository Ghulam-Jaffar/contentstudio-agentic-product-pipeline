// Package linkedin defines request and response types for the LinkedIn analytics API.
// These types map directly to the JSON contracts expected by the ContentStudio frontend,
// preserving the same field names and structure as the PHP Laravel API responses.
//
// Request types include validation logic and conversion to ClickHouse query parameters.
// Response types match the frontend's expected JSON shape for each analytics widget.
package linkedin

import (
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// --- Request types ---

// LinkedInRequest is the base request for all LinkedIn analytics endpoints.
// Query params: workspace_id, linkedin_id, start_date, end_date, timezone.
type LinkedInRequest struct {
	WorkspaceID string `json:"workspace_id"`
	LinkedinID  string `json:"linkedin_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Timezone    string `json:"timezone"`
}

// TopPostsRequest extends LinkedInRequest with pagination and filtering options.
type TopPostsRequest struct {
	LinkedInRequest
	Limit    int      `json:"limit"`
	OrderBy  string   `json:"order_by"`
	Hashtags []string `json:"hashtags"`
}

// PublishingBehaviourRequest extends LinkedInRequest with media type filtering.
type PublishingBehaviourRequest struct {
	LinkedInRequest
	MediaType []string `json:"media_type"`
}

// --- Response types ---
// All response types include a Status field and match the PHP API's JSON structure
// so the frontend can consume them without changes.

// SummaryResponse wraps current and previous period summary metrics.
type SummaryResponse struct {
	Status   bool                       `json:"status"`
	Overview map[string]*SummaryMetrics `json:"overview"`
}

// SummaryMetrics holds all summary KPIs for a single period.
type SummaryMetrics struct {
	PostComments       int32   `json:"post_comments"`
	PostLikes          int32   `json:"post_likes"`
	TotalEngagement    int32   `json:"total_engagement"`
	TotalPosts         int32   `json:"total_posts"`
	PostShares         int32   `json:"post_shares"`
	PostClicks         int32   `json:"post_clicks"`
	Followers          int32   `json:"followers"`
	PageViews          int32   `json:"page_views"`
	PageReach          int32   `json:"page_reach"`
	PageShares         int32   `json:"page_shares"`
	PageComments       int32   `json:"page_comments"`
	PageReactions      int32   `json:"page_reactions"`
	PageImpressions    int32   `json:"page_impressions"`
	PageUniqueVisitors int32   `json:"page_unique_visitors"`
	EngagementRate     float64 `json:"engagement_rate"`
	PostEngagementRate float64 `json:"post_engagement_rate"`
}

type AudienceGrowthResponse struct {
	Status               bool                             `json:"status"`
	AudienceGrowth       *AudienceGrowthData              `json:"audience_growth"`
	AudienceGrowthRollup map[string]*AudienceGrowthRollup `json:"audience_growth_rollup"`
}

type AudienceGrowthData struct {
	ShowData              int32    `json:"show_data"`
	OrganicFollowerCount  []int32  `json:"organic_follower_count"`
	OrganicFollowersDaily []int32  `json:"organic_followers_daily"`
	PaidFollowerCount     []int32  `json:"paid_follower_count"`
	PaidFollowersDaily    []int32  `json:"paid_followers_daily"`
	TotalFollowerCount    []int32  `json:"total_follower_count"`
	TotalFollowersDaily   []int32  `json:"total_followers_daily"`
	Buckets               []string `json:"buckets"`
}

type AudienceGrowthRollup struct {
	OrganicFollowerCount int32   `json:"organic_follower_count"`
	PaidFollowerCount    int32   `json:"paid_follower_count"`
	TotalFollowerCount   int32   `json:"total_follower_count"`
	AvgFollowerCount     float64 `json:"avg_follower_count"`
}

type PageViewsResponse struct {
	Status          bool                        `json:"status"`
	PageViews       *PageViewsData              `json:"page_views"`
	PageViewsRollup map[string]*PageViewsRollup `json:"page_views_rollup"`
}

type PageViewsData struct {
	DesktopPageViews      []int32  `json:"desktop_page_views"`
	MobilePageViews       []int32  `json:"mobile_page_views"`
	TotalPageViews        []int32  `json:"total_page_views"`
	DesktopPageViewsDaily []int32  `json:"desktop_page_views_daily"`
	MobilePageViewsDaily  []int32  `json:"mobile_page_views_daily"`
	TotalPageViewsDaily   []int32  `json:"total_page_views_daily"`
	ShowData              int32    `json:"show_data"`
	Buckets               []string `json:"buckets"`
}

type PageViewsRollup struct {
	TotalPageViews   int32   `json:"total_page_views"`
	DesktopPageViews int32   `json:"desktop_page_views"`
	MobilePageViews  int32   `json:"mobile_page_views"`
	AvgPageViews     float64 `json:"avg_page_views"`
}

type PublishingBehaviourResponse struct {
	Status                    bool                                      `json:"status"`
	PublishingBehaviour       *PublishingBehaviourData                  `json:"publishing_behaviour"`
	PublishingBehaviourRollup map[string][]PublishingBehaviourMediaType `json:"publishing_behaviour_rollup"`
}

type PublishingBehaviourData struct {
	Likes          []int32   `json:"likes"`
	Comments       []int32   `json:"comments"`
	Shares         []int32   `json:"shares"`
	Clicks         []int32   `json:"clicks"`
	EngagementRate []float32 `json:"engagement_rate"`
	Impressions    []int32   `json:"impressions"`
	TotalPosts     []int32   `json:"total_posts"`
	Engagement     []int32   `json:"engagement"`
	Reach          []int32   `json:"reach"`
	Buckets        []string  `json:"buckets"`
}

type PublishingBehaviourMediaType struct {
	MediaType   string `json:"media_type"`
	TotalPosts  int32  `json:"total_posts"`
	Likes       int32  `json:"likes"`
	Comments    int32  `json:"comments"`
	Shares      int32  `json:"shares"`
	Clicks      int32  `json:"clicks"`
	Engagements int32  `json:"engagements"`
	Impressions int32  `json:"impressions"`
	Reach       int32  `json:"reach"`
}

type TopPostsResponse struct {
	Status   bool      `json:"status"`
	TopPosts []TopPost `json:"top_posts"`
}

type TopPost struct {
	LinkedinID      string   `json:"linkedin_id"`
	PostID          string   `json:"post_id"`
	Activity        string   `json:"activity"`
	MediaType       string   `json:"media_type"`
	ArticleURL      string   `json:"article_url"`
	ArticleTitle    string   `json:"article_title"`
	PostData        string   `json:"post_data"`
	Image           string   `json:"image"`
	Media           []string `json:"media"`
	Type            string   `json:"type"`
	Hashtags        []string `json:"hashtags"`
	Comments        int64    `json:"comments"`
	TotalEngagement float64  `json:"total_engagement"`
	Favorites       int64    `json:"favorites"`
	Title           string   `json:"title"`
	DayOfWeek       string   `json:"day_of_week"`
	HourOfDay       int64    `json:"hour_of_day"`
	CreatedAt       string   `json:"created_at"`
	SavingTime      string   `json:"saving_time"`
	PollData        string   `json:"poll_data"`
	Reach           int64    `json:"reach"`
	Repost          int64    `json:"repost"`
	PostClicks      int64    `json:"post_clicks"`
	Impressions     int64    `json:"impressions"`
	PublishedAt     string   `json:"published_at"`
}

type PostsPerDayResponse struct {
	Status       bool             `json:"status"`
	PostsPerDays *PostsPerDayData `json:"posts_per_days"`
}

type PostsPerDayData struct {
	Data PostsPerDayInner `json:"data"`
}

type PostsPerDayInner struct {
	Days     map[string]int32 `json:"days"`
	ShowData int32            `json:"show_data"`
}

type HashtagsResponse struct {
	Status            bool                       `json:"status"`
	TopHashtags       *HashtagsData              `json:"top_hashtags"`
	TopHashtagsRollup map[string]*HashtagsRollup `json:"top_hashtags_rollup"`
}

type HashtagsData struct {
	Name        []string `json:"name"`
	Engagements []int32  `json:"engagements"`
	Likes       []int32  `json:"likes"`
	Comments    []int32  `json:"comments"`
	Shares      []int32  `json:"shares"`
	Posts       []int32  `json:"posts"`
}

type HashtagsRollup struct {
	TotalHashtags    int32 `json:"total_hashtags"`
	TotalTimesUsed   int32 `json:"total_times_used"`
	TotalLikes       int32 `json:"total_likes"`
	TotalComments    int32 `json:"total_comments"`
	TotalShares      int32 `json:"total_shares"`
	TotalEngagement  int32 `json:"total_engagement"`
	TotalImpressions int32 `json:"total_impressions"`
	TotalReach       int32 `json:"total_reach"`
}

type DemographicsResponse struct {
	Status               bool                            `json:"status"`
	FollowerDemographics map[string]*DemographicCategory `json:"follower_demographics"`
}

type DemographicCategory struct {
	Buckets []string `json:"buckets"`
	Values  []int32  `json:"values"`
}

type ErrorResponse struct {
	Status  bool   `json:"status"`
	Message string `json:"message"`
}

type AIInsightsRequest struct {
	WorkspaceID string      `json:"workspace_id"`
	LinkedinID  string      `json:"linkedin_id"`
	Date        interface{} `json:"date"`
	Timezone    string      `json:"timezone"`
	Type        string      `json:"type"`
	Limit       int         `json:"limit"`
	Language    string      `json:"language,omitempty"`
}

type AIInsightsResponse struct {
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Message string                 `json:"message,omitempty"`
}

// --- Request helpers ---

const maxLimit = 100

// Validate checks all required fields and validates date formats, ranges, and timezone.
func (r *LinkedInRequest) Validate() error {
	if r.WorkspaceID == "" {
		return httputil.NewValidationError("workspace_id is required")
	}
	if r.LinkedinID == "" {
		return httputil.NewValidationError("linkedin_id is required")
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

// GetAccountIDs returns the LinkedIn page ID as a single-element slice for ClickHouse IN clauses.
func (r *LinkedInRequest) GetAccountIDs() []string {
	if r.LinkedinID != "" {
		return []string{r.LinkedinID}
	}
	return nil
}

// GetTimezone returns the requested timezone, defaulting to UTC.
func (r *LinkedInRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

// ToQueryParams converts the request into ClickHouse query parameters with auto-calculated previous period.
func (r *LinkedInRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	params, err := clickhouse.ParseStartEndDate(r.StartDate, r.EndDate, r.GetTimezone())
	if err != nil {
		return nil, err
	}
	params.AccountIDs = r.GetAccountIDs()
	return params, nil
}

// validOrderByFields is a whitelist of allowed ORDER BY columns to prevent SQL injection.
var validOrderByFields = map[string]bool{
	"total_engagement": true,
	"impressions":      true,
	"reach":            true,
	"comments":         true,
	"favorites":        true,
	"repost":           true,
	"post_clicks":      true,
	"created_at":       true,
}

// GetLimit returns the post limit, clamped between 1 and maxLimit (100). Defaults to 3.
func (r *TopPostsRequest) GetLimit() int {
	if r.Limit <= 0 {
		return 3
	}
	if r.Limit > maxLimit {
		return maxLimit
	}
	return r.Limit
}

// GetOrderBy returns the validated sort field. Defaults to "total_engagement".
func (r *TopPostsRequest) GetOrderBy() string {
	if r.OrderBy == "" || !validOrderByFields[r.OrderBy] {
		return "total_engagement"
	}
	return r.OrderBy
}

// validMediaTypes is a whitelist of allowed media type filter values.
var validMediaTypes = map[string]bool{
	"text":      true,
	"images":    true,
	"videos":    true,
	"carousel":  true,
	"link":      true,
	"documents": true,
}

var defaultMediaTypes = []string{"text", "images", "videos", "carousel", "link"}

// GetMediaTypes returns validated media types, falling back to defaults if none are valid.
func (r *PublishingBehaviourRequest) GetMediaTypes() []string {
	if len(r.MediaType) == 0 {
		return defaultMediaTypes
	}
	valid := make([]string, 0, len(r.MediaType))
	for _, mt := range r.MediaType {
		if validMediaTypes[mt] {
			valid = append(valid, mt)
		}
	}
	if len(valid) == 0 {
		return defaultMediaTypes
	}
	return valid
}
