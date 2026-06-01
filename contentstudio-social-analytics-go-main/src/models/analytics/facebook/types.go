package facebook

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

const maxLimit = 100

var defaultMediaTypes = []string{"text", "link", "images", "videos", "carousel", "share", "reels", "others"}

var validMediaTypes = map[string]bool{
	"text":     true,
	"link":     true,
	"images":   true,
	"videos":   true,
	"carousel": true,
	"share":    true,
	"reels":    true,
	"others":   true,
}

var validOrderByFields = map[string]bool{
	"total_engagement":        true,
	"created_time":            true,
	"comments":                true,
	"shares":                  true,
	"post_clicks":             true,
	"post_impressions":        true,
	"post_impressions_unique": true,
	"post_video_views":        true,
	"total":                   true,
}

type StringList []string

func (s *StringList) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		*s = nil
		return nil
	}

	var single string
	if err := json.Unmarshal(data, &single); err == nil {
		*s = normalizeStringList([]string{single})
		return nil
	}

	var many []string
	if err := json.Unmarshal(data, &many); err == nil {
		*s = normalizeStringList(many)
		return nil
	}

	return fmt.Errorf("expected string or []string")
}

type FacebookRequest struct {
	WorkspaceID string   `json:"workspace_id"`
	FacebookIDs []string `json:"facebook_id"`
	Date        string   `json:"date,omitempty"`
	StartDate   string   `json:"start_date,omitempty"`
	EndDate     string   `json:"end_date,omitempty"`
	Timezone    string   `json:"timezone,omitempty"`
	MediaType   []string `json:"media_type,omitempty"`
	Limit       int      `json:"limit,omitempty"`
	OrderBy     string   `json:"order_by,omitempty"`
}

type SummaryResponse struct {
	Status   bool                       `json:"status"`
	Overview map[string]*SummaryMetrics `json:"overview"`
}

type SummaryMetrics struct {
	DocCount               int32 `json:"doc_count"`
	TotalEngagement        int32 `json:"total_engagement"`
	Reactions              int32 `json:"reactions"`
	Comments               int32 `json:"comments"`
	PostsClicks            int32 `json:"posts_clicks"`
	Impressions            int32 `json:"impressions"`
	Reach                  int32 `json:"reach"`
	Repost                 int32 `json:"repost"`
	PositiveSentiment      int32 `json:"positive_sentiment"`
	NegativeSentiment      int32 `json:"negative_sentiment"`
	PageImpressions        int32 `json:"page_impressions"`
	PageImpressionsPaid    int32 `json:"page_impressions_paid"`
	PageImpressionsOrganic int32 `json:"page_impressions_organic"`
	PageEngagements        int32 `json:"page_engagements"`
	PagePositiveFeedback   int32 `json:"page_positive_feedback"`
	PageNegativeFeedback   int32 `json:"page_negative_feedback"`
	FanCount               int32 `json:"fan_count"`
	TalkingAboutCount      int32 `json:"talking_about_count"`
	PageFollows            int32 `json:"page_follows"`
}

type AudienceGrowthResponse struct {
	Status               bool                             `json:"status"`
	AudienceGrowth       *AudienceGrowthData              `json:"audience_growth"`
	AudienceGrowthRollup map[string]*AudienceGrowthRollup `json:"audience_growth_rollup"`
}

type AudienceGrowthData struct {
	ShowData         int32    `json:"show_data"`
	FanCount         []int32  `json:"fan_count"`
	PageFansDaily    []int32  `json:"page_fans_daily"`
	PageFansByLike   []int32  `json:"page_fans_by_like"`
	PageFansByUnlike []int32  `json:"page_fans_by_unlike"`
	PageImpressions  []int32  `json:"page_impressions"`
	PageEngagements  []int32  `json:"page_engagements"`
	Buckets          []string `json:"buckets"`
}

type AudienceGrowthRollup struct {
	AvgPageFansByLike   float64 `json:"avg_page_fans_by_like"`
	AvgPageFansByUnlike float64 `json:"avg_page_fans_by_unlike"`
	FanCount            int32   `json:"fan_count"`
	TalkingAboutCount   int32   `json:"talking_about_count"`
	DocCount            int32   `json:"doc_count"`
}

type PublishingBehaviourResponse struct {
	Status                    bool                         `json:"status"`
	PublishingBehaviour       *PublishingBehaviourData     `json:"publishing_behaviour"`
	PublishingBehaviourRollup map[string]*PublishingRollup `json:"publishing_behaviour_rollup"`
}

type PublishingBehaviourData struct {
	ReactionsEngagement []int32  `json:"reactions_engagement"`
	CommentsEngagement  []int32  `json:"comments_engagement"`
	SharesEngagement    []int32  `json:"shares_engagement"`
	PaidImpressions     []int32  `json:"paid_impressions"`
	OrganicImpressions  []int32  `json:"organic_impressions"`
	ViralImpressions    []int32  `json:"viral_impressions"`
	PaidReach           []int32  `json:"paid_reach"`
	OrganicReach        []int32  `json:"organic_reach"`
	ViralReach          []int32  `json:"viral_reach"`
	Buckets             []string `json:"buckets"`
	PostCount           []int32  `json:"post_count"`
}

type PublishingRollup struct {
	DocCount        int32 `json:"doc_count"`
	TotalEngagement int32 `json:"total_engagement"`
	Reactions       int32 `json:"reactions"`
	Comments        int32 `json:"comments"`
	PostClicks      int32 `json:"post_clicks"`
	Impressions     int32 `json:"impressions"`
	Shares          int32 `json:"shares"`
}

type TopPostsResponse struct {
	Status   bool      `json:"status"`
	TopPosts []TopPost `json:"top_posts"`
}

type TopPost struct {
	PageName                     string       `json:"page_name"`
	PageID                       string       `json:"page_id"`
	PostID                       string       `json:"post_id"`
	Permalink                    string       `json:"permalink"`
	StatusType                   string       `json:"status_type"`
	MediaType                    string       `json:"media_type"`
	VideoID                      string       `json:"video_id"`
	Category                     string       `json:"category"`
	PublishedBy                  string       `json:"published_by"`
	PublishedByURL               string       `json:"published_by_url"`
	SharedFromName               string       `json:"shared_from_name"`
	SharedFromID                 string       `json:"shared_from_id"`
	SharedFromLink               string       `json:"shared_from_link"`
	Like                         int64        `json:"like"`
	Love                         int64        `json:"love"`
	Haha                         int64        `json:"haha"`
	Wow                          int64        `json:"wow"`
	Sad                          int64        `json:"sad"`
	Angry                        int64        `json:"angry"`
	Total                        int64        `json:"total"`
	Shares                       int64        `json:"shares"`
	Comments                     int64        `json:"comments"`
	PostClicks                   int64        `json:"post_clicks"`
	TotalEngagement              float64      `json:"total_engagement"`
	PostEngagedUsers             int64        `json:"post_engaged_users"`
	DayOfWeek                    string       `json:"day_of_week"`
	HourOfDay                    int64        `json:"hour_of_day"`
	CreatedTime                  string       `json:"created_time"`
	UpdatedTime                  string       `json:"updated_time"`
	SavingTime                   string       `json:"saving_time"`
	MessageTags                  string       `json:"message_tags"`
	PostMetadata                 string       `json:"post_metadata"`
	Caption                      string       `json:"caption"`
	Description                  string       `json:"description"`
	FullPicture                  string       `json:"full_picture"`
	Link                         string       `json:"link"`
	PostImpressions              int64        `json:"post_impressions"`
	PostImpressionsUnique        int64        `json:"post_impressions_unique"`
	PostImpressionsPaid          int64        `json:"post_impressions_paid"`
	PostImpressionsPaidUnique    int64        `json:"post_impressions_paid_unique"`
	PostImpressionsOrganic       int64        `json:"post_impressions_organic"`
	PostImpressionsOrganicUnique int64        `json:"post_impressions_organic_unique"`
	PostImpressionsViral         int64        `json:"post_impressions_viral"`
	PostImpressionsViralUnique   int64        `json:"post_impressions_viral_unique"`
	PostVideoViews               int64        `json:"post_video_views"`
	TotalImpressions             int64        `json:"total_impressions"`
	MediaAssets                  []MediaAsset `json:"media_assets"`
}

type MediaAsset struct {
	MediaID      string `json:"media_id"`
	Caption      string `json:"caption"`
	Link         string `json:"link"`
	AssetType    string `json:"assetType"`
	CallToAction string `json:"callToAction"`
	CreatedAt    string `json:"createdAt"`
}

type ActiveUsersResponse struct {
	Status      bool             `json:"status"`
	ActiveUsers *ActiveUsersData `json:"active_users"`
}

type ActiveUsersData struct {
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
	PageImpressions []int32  `json:"page_impressions"`
	Buckets         []string `json:"buckets"`
}

type ImpressionsRollup struct {
	TotalImpressions      int32   `json:"total_impressions"`
	AvgImpressionsPerDay  float64 `json:"avg_impressions_per_day"`
	AvgImpressionsPerWeek float64 `json:"avg_impressions_per_week"`
}

type EngagementResponse struct {
	Status     bool                 `json:"status"`
	Engagement *EngagementContainer `json:"engagement"`
}

type EngagementContainer struct {
	Engagement       *EngagementData              `json:"engagement"`
	EngagementRollup map[string]*EngagementRollup `json:"engagement_rollup"`
}

type EngagementData struct {
	PageEngagements []int32  `json:"page_engagements"`
	Buckets         []string `json:"buckets"`
}

type EngagementRollup struct {
	PageEngagements       int32   `json:"page_engagements"`
	AvgEngagementsPerDay  float64 `json:"avg_engagements_per_day"`
	AvgEngagementsPerWeek float64 `json:"avg_engagements_per_week"`
}

type ReelsAnalyticsResponse struct {
	Status      bool                    `json:"status"`
	Reels       *ReelsData              `json:"reels"`
	ReelsRollup map[string]*ReelsRollup `json:"reels_rollup"`
}

type ReelsData struct {
	Buckets             []string  `json:"buckets"`
	TotalReels          []int32   `json:"total_reels"`
	TotalSecondsWatched []float64 `json:"total_seconds_watched"`
	InitialPlays        []int32   `json:"initial_plays"`
	Engagement          []int32   `json:"engagement"`
	Reactions           []int32   `json:"reactions"`
	Comments            []int32   `json:"comments"`
	Shares              []int32   `json:"shares"`
	ShowData            int32     `json:"show_data"`
}

type ReelsRollup struct {
	TotalReels            int32   `json:"total_reels"`
	AverageSecondsWatched float64 `json:"average_seconds_watched"`
	TotalSecondsWatched   int32   `json:"total_seconds_watched"`
	InitialPlays          int32   `json:"initial_plays"`
	Reach                 int32   `json:"reach"`
	Engagement            int32   `json:"engagement"`
	Reactions             int32   `json:"reactions"`
	Comments              int32   `json:"comments"`
	Shares                int32   `json:"shares"`
}

type VideoInsightsResponse struct {
	Status        bool                    `json:"status"`
	VideoInsights *VideoInsightsData      `json:"video_insights"`
	VideoRollup   map[string]*VideoRollup `json:"video_rollup"`
}

type VideoInsightsData struct {
	Buckets         []string  `json:"buckets"`
	TotalViewTime   []float64 `json:"total_view_time"`
	OrganicViewTime []float64 `json:"organic_view_time"`
	PaidViewTime    []float64 `json:"paid_view_time"`
	TotalViews      []int32   `json:"total_views"`
	OrganicViews    []int32   `json:"organic_views"`
	PaidViews       []int32   `json:"paid_views"`
	Comments        []int32   `json:"comments"`
	Reactions       []int32   `json:"reactions"`
	Shares          []int32   `json:"shares"`
	TotalPosts      []int32   `json:"total_posts"`
}

type VideoRollup struct {
	TotalViewTime   float64 `json:"total_view_time"`
	OrganicViewTime float64 `json:"organic_view_time"`
	PaidViewTime    float64 `json:"paid_view_time"`
	TotalViews      int32   `json:"total_views"`
	OrganicViews    int32   `json:"organic_views"`
	PaidViews       int32   `json:"paid_views"`
	Comments        int32   `json:"comments"`
	Reactions       int32   `json:"reactions"`
	Shares          int32   `json:"shares"`
	TotalPosts      int32   `json:"total_posts"`
}

type DemographicsResponse struct {
	Status          bool             `json:"status"`
	AudienceGender  map[string]int32 `json:"audience_gender,omitempty"`
	Fans            int32            `json:"fans,omitempty"`
	AudienceAge     *AudienceAgeData `json:"audience_age,omitempty"`
	MaxGenderAge    *MaxGenderAge    `json:"max_gender_age,omitempty"`
	AudienceCountry map[string]int32 `json:"audience_country,omitempty"`
	AudienceCity    map[string]int32 `json:"audience_city,omitempty"`
}

type AIInsightsRequest struct {
	WorkspaceID string      `json:"workspace_id"`
	FacebookID  string      `json:"facebook_id"`
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

type AudienceAgeData struct {
	FansAge *AgeBreakdown `json:"fans_age"`
	MaxAge  int32         `json:"max_age"`
}

type AgeBreakdown struct {
	Age65Plus int32 `json:"65+"`
	Age55To64 int32 `json:"55-64"`
	Age45To54 int32 `json:"45-54"`
	Age35To44 int32 `json:"35-44"`
	Age25To34 int32 `json:"25-34"`
	Age18To34 int32 `json:"18-34"`
	Age13To17 int32 `json:"13-17"`
}

type MaxGenderAge struct {
	MaxValue int32  `json:"max_value"`
	Age      string `json:"age"`
	Gender   string `json:"gender"`
}

func (r *FacebookRequest) Validate() error {
	if r.WorkspaceID == "" {
		return httputil.NewValidationError("workspace_id is required")
	}
	if len(r.GetAccountIDs()) == 0 {
		return httputil.NewValidationError("facebook_id is required")
	}

	start, end, err := r.normalizedDates()
	if err != nil {
		return httputil.NewValidationError(err.Error())
	}
	r.StartDate = start
	r.EndDate = end
	if r.Timezone != "" {
		if _, err := time.LoadLocation(r.Timezone); err != nil {
			return httputil.NewValidationError("invalid timezone: " + r.Timezone)
		}
	}
	return nil
}

func (r *FacebookRequest) GetAccountIDs() []string {
	r.FacebookIDs = normalizeStringList(r.FacebookIDs)
	return r.FacebookIDs
}

func (r *FacebookRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

func (r *FacebookRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	start, end, err := r.normalizedDates()
	if err != nil {
		return nil, err
	}

	params, err := clickhouse.ParseStartEndDate(start, end, r.GetTimezone())
	if err != nil {
		return nil, err
	}
	params.AccountIDs = r.GetAccountIDs()
	return params, nil
}

func (r *FacebookRequest) GetMediaTypes() []string {
	if len(r.MediaType) == 0 {
		return defaultMediaTypes
	}
	valid := make([]string, 0, len(r.MediaType))
	for _, mt := range normalizeStringList(r.MediaType) {
		if validMediaTypes[mt] {
			valid = append(valid, mt)
		}
	}
	if len(valid) == 0 {
		return defaultMediaTypes
	}
	return valid
}

func (r *FacebookRequest) GetOrderBy() string {
	if !validOrderByFields[r.OrderBy] {
		return "total_engagement"
	}
	return r.OrderBy
}

func (r *FacebookRequest) GetLimit(defaultValue int) int {
	if r.Limit <= 0 {
		return defaultValue
	}
	if r.Limit > maxLimit {
		return maxLimit
	}
	return r.Limit
}

func (r *FacebookRequest) normalizedDates() (string, string, error) {
	if r.Date != "" {
		parts := strings.SplitN(r.Date, " - ", 2)
		if len(parts) != 2 {
			return "", "", fmt.Errorf("date must be in 'YYYY-MM-DD - YYYY-MM-DD' format")
		}
		start, err := normalizeDate(parts[0])
		if err != nil {
			return "", "", fmt.Errorf("invalid start date")
		}
		end, err := normalizeDate(parts[1])
		if err != nil {
			return "", "", fmt.Errorf("invalid end date")
		}
		if end < start {
			return "", "", fmt.Errorf("end date cannot be before start date")
		}
		return start, end, nil
	}

	start, err := normalizeDate(r.StartDate)
	if err != nil {
		return "", "", fmt.Errorf("start_date must be in YYYY-MM-DD format")
	}
	end, err := normalizeDate(r.EndDate)
	if err != nil {
		return "", "", fmt.Errorf("end_date must be in YYYY-MM-DD format")
	}
	if end < start {
		return "", "", fmt.Errorf("end_date cannot be before start_date")
	}
	return start, end, nil
}

func normalizeDate(value string) (string, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return "", fmt.Errorf("empty date")
	}
	if len(value) >= 10 {
		value = value[:10]
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return "", err
	}
	return t.Format("2006-01-02"), nil
}

func normalizeStringList(values []string) []string {
	var out []string
	for _, value := range values {
		for _, part := range strings.Split(value, ",") {
			part = strings.TrimSpace(part)
			part = strings.Trim(part, "()[]\"'")
			if part == "" {
				continue
			}
			out = append(out, part)
		}
	}
	return out
}
