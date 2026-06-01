// Package instagram defines request and response types for the Instagram analytics API.
// These types map directly to the JSON contracts expected by the ContentStudio frontend,
// preserving the same field names and structure as the PHP Laravel API responses.
//
// Request types include validation logic and conversion to ClickHouse query parameters.
// Response types match the frontend's expected JSON shape for each analytics widget.
package instagram

import (
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// --- Request types ---

// InstagramRequest is the base request for all Instagram analytics endpoints.
// Query params: workspace_id, instagram_id, start_date, end_date, timezone.
type InstagramRequest struct {
	WorkspaceID string `json:"workspace_id"`
	InstagramID string `json:"instagram_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Timezone    string `json:"timezone"`
}

// TopPostsRequest extends InstagramRequest with pagination and filtering options.
type TopPostsRequest struct {
	InstagramRequest
	Limit    int      `json:"limit"`
	OrderBy  string   `json:"order_by"`
	Hashtags []string `json:"hashtags"`
}

// PublishingBehaviourRequest extends InstagramRequest with media type filtering.
type PublishingBehaviourRequest struct {
	InstagramRequest
	MediaType []string `json:"media_type"`
}

// AIInsightsRequest is the request for Instagram AI insights.
type AIInsightsRequest struct {
	WorkspaceID string      `json:"workspace_id"`
	InstagramID string      `json:"instagram_id"`
	Date        interface{} `json:"date"`
	Timezone    string      `json:"timezone"`
	Type        string      `json:"type"`
	Limit       int         `json:"limit"`
	Language    string      `json:"language,omitempty"`
}

// --- Response types ---

type SummaryResponse struct {
	Status   bool                       `json:"status"`
	Overview map[string]*SummaryMetrics `json:"overview"`
}

type SummaryMetrics struct {
	TotalPosts         int64   `json:"total_posts"`
	PostEngagement     int64   `json:"post_engagement"`
	PostReactions      int64   `json:"post_reactions"`
	PostComments       int64   `json:"post_comments"`
	PostSaves          int64   `json:"post_saves"`
	PostReach          int64   `json:"post_reach"`
	ProfileImpressions int64   `json:"profile_impressions"`
	PostViews          int64   `json:"post_views"`
	TotalStories       int64   `json:"total_stories"`
	ProfileViews       int64   `json:"profile_views"`
	FollowersCount     int64   `json:"followers_count"`
	FollowsCount       int64   `json:"follows_count"`
	AccountsEngaged    int64   `json:"accounts_engaged"`
	ProfileEngagement  int64   `json:"profile_engagement"`
	ProfileReach       int64   `json:"profile_reach"`
	DocCount           int64   `json:"doc_count"`
	EngRate            float64 `json:"eng_rate"`
}

type AudienceGrowthResponse struct {
	Status               bool                             `json:"status"`
	AudienceGrowth       *AudienceGrowthData              `json:"audience_growth"`
	AudienceGrowthRollup map[string]*AudienceGrowthRollup `json:"audience_growth_rollup"`
}

type AudienceGrowthData struct {
	ShowData       int32    `json:"show_data"`
	Followers      []int32  `json:"followers"`
	FollowersDaily []int32  `json:"followers_daily"`
	Buckets        []string `json:"buckets"`
}

type AudienceGrowthRollup struct {
	FollowerCount  int32 `json:"follower_count"`
	FollowerGained int32 `json:"follower_gained"`
}

type PublishingBehaviourResponse struct {
	Status                    bool                                      `json:"status"`
	PublishingBehaviour       *PublishingBehaviourData                  `json:"publishing_behaviour"`
	PublishingBehaviourRollup map[string][]PublishingBehaviourMediaType `json:"publishing_behaviour_rollup"`
}

type PublishingBehaviourData struct {
	Likes       []int32  `json:"likes"`
	Comments    []int32  `json:"comments"`
	Saved       []int32  `json:"saved"`
	Engagement  []int32  `json:"engagement"`
	Reach       []int32  `json:"reach"`
	Impressions []int32  `json:"impressions"`
	Views       []int32  `json:"views"`
	TotalPosts  []int32  `json:"total_posts"`
	Buckets     []string `json:"buckets"`
}

type PublishingBehaviourMediaType struct {
	MediaType  string `json:"media_type"`
	TotalPosts int32  `json:"total_posts"`
	Likes      int32  `json:"likes"`
	Comments   int32  `json:"comments"`
	Saved      int32  `json:"saved"`
	Engagement int32  `json:"engagement"`
	Reach      int32  `json:"reach"`
	Views      int32  `json:"views"`
}

type TopPostsResponse struct {
	Status   bool      `json:"status"`
	TopPosts []TopPost `json:"top_posts"`
}

type TopPost struct {
	InstagramID         string   `json:"instagram_id"`
	MediaID             string   `json:"media_id"`
	Caption             string   `json:"caption"`
	MediaType           string   `json:"media_type"`
	EntityType          string   `json:"entity_type"`
	MediaURL            []string `json:"media_url"`
	VideoURL            []string `json:"video_url"`
	Permalink           string   `json:"permalink"`
	LikeCount           int64    `json:"like_count"`
	CommentsCount       int64    `json:"comments_count"`
	Saved               int64    `json:"saved"`
	Engagement          int64    `json:"engagement"`
	Reach               int64    `json:"reach"`
	Impressions         int64    `json:"impressions"`
	Views               int64    `json:"views"`
	Shares              int64    `json:"shares"`
	ReelsAvgWatchTime   int64    `json:"reels_avg_watch_time"`
	ReelsTotalWatchTime int64    `json:"reels_total_watch_time"`
	Exits               int64    `json:"exits"`
	Replies             int64    `json:"replies"`
	Hashtags            []string `json:"hashtags"`
	DayOfWeek           string   `json:"day_of_week"`
	HourOfDay           int64    `json:"hour_of_day"`
	PostCreatedAt       string   `json:"post_created_at"`
	StoredEventAt       string   `json:"stored_event_at"`
}

type ActiveUsersResponse struct {
	Status           bool              `json:"status"`
	ActiveUsersHours *ActiveUsersHours `json:"active_users_hours"`
	ActiveUsersDays  *ActiveUsersDays  `json:"active_users_days"`
}

type ActiveUsersHours struct {
	Buckets      []int32 `json:"buckets"`
	Values       []int32 `json:"values"`
	HighestValue int32   `json:"highest_value"`
	HighestHour  int32   `json:"highest_hour"`
}

type ActiveUsersDays struct {
	Buckets      []string `json:"buckets"`
	Values       []int32  `json:"values"`
	HighestValue int32    `json:"highest_value"`
	HighestDay   string   `json:"highest_day"`
}

type ImpressionsResponse struct {
	Status            bool                          `json:"status"`
	Impressions       *ImpressionsData              `json:"impressions"`
	ImpressionsRollup map[string]*ImpressionsRollup `json:"impressions_rollup"`
}

type ImpressionsData struct {
	ShowData    int32    `json:"show_data"`
	Buckets     []string `json:"buckets"`
	Impressions []int32  `json:"impressions"`
}

type ImpressionsRollup struct {
	TotalImpressions int64   `json:"total_impressions"`
	AvgImpressions   float64 `json:"avg_impressions"`
}

type EngagementResponse struct {
	Status            bool                         `json:"status"`
	Engagements       *EngagementData              `json:"engagements"`
	EngagementsRollup map[string]*EngagementRollup `json:"engagements_rollup"`
}

type EngagementData struct {
	ShowData   int32    `json:"show_data"`
	Buckets    []string `json:"buckets"`
	Engagement []int32  `json:"engagement"`
	Comments   []int32  `json:"comments"`
	Reactions  []int32  `json:"reactions"`
	DocCount   []int32  `json:"doc_count"`
}

type EngagementRollup struct {
	Engagement    int64   `json:"engagement"`
	AvgEngagement float64 `json:"avg_engagement"`
	Comments      int64   `json:"comments"`
	Reactions     int64   `json:"reactions"`
	Saved         int64   `json:"saved"`
	Count         int64   `json:"count"`
}

type HashtagsResponse struct {
	Status            bool                       `json:"status"`
	TopHashtags       *HashtagsData              `json:"top_hashtags"`
	TopHashtagsRollup map[string]*HashtagsRollup `json:"top_hashtags_rollup"`
}

type HashtagsData struct {
	Name       []string `json:"name"`
	Engagement []int32  `json:"engagement"`
	Likes      []int32  `json:"likes"`
	Comments   []int32  `json:"comments"`
	Saved      []int32  `json:"saved"`
	Posts      []int32  `json:"posts"`
}

type HashtagsRollup struct {
	TotalEngagement     int32 `json:"total_engagement"`
	TotalLikes          int32 `json:"total_likes"`
	TotalComments       int32 `json:"total_comments"`
	TotalSaves          int32 `json:"total_saves"`
	TotalUniqueHashtags int32 `json:"total_unique_hashtags"`
	TotalHashtagUses    int32 `json:"total_hashtag_uses"`
}

type StoriesPerformanceResponse struct {
	Status             bool                      `json:"status"`
	StoriesPerformance *StoriesData              `json:"stories_performance"`
	StoriesRollup      map[string]*StoriesRollup `json:"stories_rollup"`
}

type StoriesData struct {
	ShowData            int32     `json:"show_data"`
	Buckets             []string  `json:"buckets"`
	AvgStoryImpressions []float64 `json:"avg_story_impressions"`
	StoryImpressions    []int32   `json:"story_impressions"`
	StoryReach          []int32   `json:"story_reach"`
	StoryReply          []int32   `json:"story_reply"`
	StoryExits          []int32   `json:"story_exits"`
	StoryTapsForward    []int32   `json:"story_taps_forward"`
	StoryTapsBack       []int32   `json:"story_taps_back"`
	PublishedStories    []int32   `json:"published_stories"`
}

type StoriesRollup struct {
	StoryImpressions    int64   `json:"story_impressions"`
	AvgStoryImpressions float64 `json:"avg_story_impressions"`
	StoryReach          int64   `json:"story_reach"`
	StoryReply          int64   `json:"story_reply"`
	StoryExits          int64   `json:"story_exits"`
	StoryTapsForward    int64   `json:"story_taps_forward"`
	StoryTapsBack       int64   `json:"story_taps_back"`
	PublishedStories    int64   `json:"published_stories"`
}

type ReelsPerformanceResponse struct {
	Status      bool                    `json:"status"`
	Reels       *ReelsData              `json:"reels"`
	ReelsRollup map[string]*ReelsRollup `json:"reels_rollup"`
}

type ReelsData struct {
	ShowData       int32     `json:"show_data"`
	Buckets        []string  `json:"buckets"`
	TotalPosts     []int32   `json:"total_posts"`
	Engagement     []int32   `json:"engagement"`
	Likes          []int32   `json:"likes"`
	Comments       []int32   `json:"comments"`
	Saves          []int32   `json:"saves"`
	Shares         []int32   `json:"shares"`
	AvgWatchTime   []float64 `json:"avg_watch_time"`
	TotalWatchTime []int64   `json:"total_watch_time"`
}

type ReelsRollup struct {
	Engagement     int64   `json:"engagement"`
	Likes          int64   `json:"likes"`
	Comments       int64   `json:"comments"`
	Saves          int64   `json:"saves"`
	TotalPosts     int64   `json:"total_posts"`
	Shares         int64   `json:"shares"`
	AvgWatchTime   float64 `json:"avg_watch_time"`
	TotalWatchTime int64   `json:"total_watch_time"`
}

type DemographicsAgeResponse struct {
	AudienceAge    map[string]int64 `json:"audience_age"`
	AudienceGender map[string]int64 `json:"audience_gender"`
	MaxAudienceAge *MaxAudienceAge  `json:"max_audience_age"`
}

type MaxAudienceAge struct {
	Gender string `json:"gender"`
	Age    string `json:"age"`
	Value  int64  `json:"value"`
}

type CountryCityResponse struct {
	AudienceCity    map[string]int64 `json:"audience_city"`
	AudienceCountry map[string]int64 `json:"audience_country"`
}

// --- Request helpers ---

const maxLimit = 100

// Validate checks all required fields and validates date formats and timezone.
func (r *InstagramRequest) Validate() error {
	if r.WorkspaceID == "" {
		return httputil.NewValidationError("workspace_id is required")
	}
	if r.InstagramID == "" {
		return httputil.NewValidationError("instagram_id is required")
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

// GetAccountIDs returns the Instagram account ID as a single-element slice.
func (r *InstagramRequest) GetAccountIDs() []string {
	if r.InstagramID != "" {
		return []string{r.InstagramID}
	}
	return nil
}

// GetTimezone returns the requested timezone, defaulting to UTC.
func (r *InstagramRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

// ToQueryParams converts the request into ClickHouse query parameters.
func (r *InstagramRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	params, err := clickhouse.ParseStartEndDate(r.StartDate, r.EndDate, r.GetTimezone())
	if err != nil {
		return nil, err
	}
	params.AccountIDs = r.GetAccountIDs()
	return params, nil
}

// validOrderByFields is the whitelist of allowed ORDER BY columns to prevent SQL injection.
var validOrderByFields = map[string]bool{
	"total_engagement": true,
	"like_count":       true,
	"comments_count":   true,
	"saved":            true,
	"reach":            true,
	"impressions":      true,
	"views":            true,
	"shares":           true,
	"post_created_at":  true,
}

// GetLimit returns the post limit, clamped between 1 and maxLimit (100).
func (r *TopPostsRequest) GetLimit() int {
	if r.Limit <= 0 {
		return 15
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

// validMediaTypes is the whitelist of allowed Instagram media type filter values.
var validMediaTypes = map[string]bool{
	"REELS":          true,
	"IMAGE":          true,
	"VIDEO":          true,
	"CAROUSEL_ALBUM": true,
}

var defaultMediaTypes = []string{"REELS", "IMAGE", "VIDEO", "CAROUSEL_ALBUM"}

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
