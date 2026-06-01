// Package overview defines request and response types for the cross-platform Overview analytics API.
// These types map to the JSON contracts expected by the ContentStudio frontend, preserving the
// same field names and structure as the PHP OverviewV2Controller responses.
package overview

// OverviewRequest is the base request for all overview analytics endpoints.
// Query params: workspace_id, start_date, end_date ("YYYY-MM-DD"), timezone, and per-platform account ID arrays.
type OverviewRequest struct {
	WorkspaceID      string   `json:"workspace_id"`
	StartDate        string   `json:"start_date"`
	EndDate          string   `json:"end_date"`
	Timezone         string   `json:"timezone"`
	FacebookAccounts []string `json:"facebook_accounts"`
	InstagramAccounts []string `json:"instagram_accounts"`
	LinkedInAccounts []string `json:"linkedin_accounts"`
	TiktokAccounts   []string `json:"tiktok_accounts"`
	YouTubeAccounts  []string `json:"youtube_accounts"`
	PinterestAccounts []string `json:"pinterest_accounts"`
}

// TopPostsRequest extends OverviewRequest with sort type and result limit for getTopPosts.
type TopPostsRequest struct {
	OverviewRequest
	Type  string `json:"type"`
	Limit int    `json:"limit"`
}

// --- Response types ---

// SummaryData holds the aggregated cross-platform metrics matching the PHP getSummary response shape.
type SummaryData struct {
	Followers              int64   `json:"followers"`
	Posts                  int64   `json:"posts"`
	Engagement             int64   `json:"engagement"`
	Impressions            int64   `json:"impressions"`
	Reach                  int64   `json:"reach"`
	EngagementRate         float64 `json:"engagement_rate"`
	SecondaryFollowers     int64   `json:"secondary_followers"`
	SecondaryPosts         int64   `json:"secondary_posts"`
	SecondaryEngagement    int64   `json:"secondary_engagement"`
	SecondaryImpressions   int64   `json:"secondary_impressions"`
	SecondaryReach         int64   `json:"secondary_reach"`
	SecondaryEngagementRate float64 `json:"secondary_engagement_rate"`
	DiffFollowers          int64   `json:"diff_followers"`
	DiffPosts              int64   `json:"diff_posts"`
	DiffEngagement         int64   `json:"diff_engagement"`
	DiffImpressions        int64   `json:"diff_impressions"`
	DiffReach              int64   `json:"diff_reach"`
	DiffEngagementRate     float64 `json:"diff_engagement_rate"`
	FollowersChangePct     float64 `json:"followers_change_pct"`
	PostsChangePct         float64 `json:"posts_change_pct"`
	EngagementChangePct    float64 `json:"engagement_change_pct"`
	ImpressionsChangePct   float64 `json:"impressions_change_pct"`
	ReachChangePct         float64 `json:"reach_change_pct"`
	EngagementRateChangePct float64 `json:"engagement_rate_change_pct"`
}

// SummaryResponse wraps SummaryData under a "summary" key, matching the PHP getSummary envelope.
type SummaryResponse struct {
	Summary *SummaryData `json:"summary"`
}

// TopPerformingGraphResponse holds per-platform daily time-series arrays for the top performing graph.
// Buckets are formatted as "YYYY-MM-DD" strings.
type TopPerformingGraphResponse struct {
	Buckets                  []string  `json:"buckets"`
	FacebookPostCount        []float64 `json:"facebook_post_count"`
	InstagramPostCount       []float64 `json:"instagram_post_count"`
	LinkedInPostCount        []float64 `json:"linkedin_post_count"`
	TiktokPostCount          []float64 `json:"tiktok_post_count"`
	YouTubePostCount         []float64 `json:"youtube_post_count"`
	PinterestPostCount       []float64 `json:"pinterest_post_count"`
	FacebookEngagementCount  []float64 `json:"facebook_engagement_count"`
	InstagramEngagementCount []float64 `json:"instagram_engagement_count"`
	LinkedInEngagementCount  []float64 `json:"linkedin_engagement_count"`
	TiktokEngagementCount    []float64 `json:"tiktok_engagement_count"`
	YouTubeEngagementCount   []float64 `json:"youtube_engagement_count"`
	PinterestEngagementCount []float64 `json:"pinterest_engagement_count"`
	FacebookImpressionCount  []float64 `json:"facebook_impression_count"`
	InstagramImpressionCount []float64 `json:"instagram_impression_count"`
	LinkedInImpressionCount  []float64 `json:"linkedin_impression_count"`
	TiktokImpressionCount    []float64 `json:"tiktok_impression_count"`
	YouTubeImpressionCount   []float64 `json:"youtube_impression_count"`
	PinterestImpressionCount []float64 `json:"pinterest_impression_count"`
	FacebookReachCount       []float64 `json:"facebook_reach_count"`
	InstagramReachCount      []float64 `json:"instagram_reach_count"`
	LinkedInReachCount       []float64 `json:"linkedin_reach_count"`
	TiktokReachCount         []float64 `json:"tiktok_reach_count"`
	YouTubeReachCount        []float64 `json:"youtube_reach_count"`
	PinterestReachCount      []float64 `json:"pinterest_reach_count"`
}

// PlatformDataRow represents a single platform's aggregated data for getPlatformData with type="grouped".
type PlatformDataRow struct {
	Followers    int32  `json:"followers"`
	TotalPosts   int32  `json:"total_posts"`
	Engagement   int32  `json:"engagement"`
	Impressions  int32  `json:"impressions"`
	Reach        int32  `json:"reach"`
	Reactions    int32  `json:"reactions"`
	Comments     int32  `json:"comments"`
	Shares       int32  `json:"shares"`
	PlatformType string `json:"platform_type"`
}

// AccountDataRow represents a single account's aggregated data for getPlatformData with type="individual".
type AccountDataRow struct {
	Followers    int32  `json:"followers"`
	TotalPosts   int32  `json:"total_posts"`
	Engagement   int32  `json:"engagement"`
	Impressions  int32  `json:"impressions"`
	Reach        int32  `json:"reach"`
	Reactions    int32  `json:"reactions"`
	Comments     int32  `json:"comments"`
	Shares       int32  `json:"shares"`
	PlatformType string `json:"platform_type"`
	AccountID    string `json:"account_id"`
}

// AccountDataDetailedRow represents current/previous period metrics with pct changes for a single account.
type AccountDataDetailedRow struct {
	PlatformType         string  `json:"platform_type"`
	AccountID            string  `json:"account_id"`
	AccountName          string  `json:"account_name"`
	CurrentFollowers     int32   `json:"current_followers"`
	OldFollowers         int32   `json:"old_followers"`
	CurrentPosts         int32   `json:"current_posts"`
	OldPosts             int32   `json:"old_posts"`
	CurrentEngagement    int32   `json:"current_engagement"`
	OldEngagement        int32   `json:"old_engagement"`
	CurrentImpressions   int32   `json:"current_impressions"`
	OldImpressions       int32   `json:"old_impressions"`
	CurrentReach         int32   `json:"current_reach"`
	OldReach             int32   `json:"old_reach"`
	FollowersChangePct   float64 `json:"followers_change_pct"`
	PostsChangePct       float64 `json:"posts_change_pct"`
	EngagementChangePct  float64 `json:"engagement_change_pct"`
	ImpressionsChangePct float64 `json:"impressions_change_pct"`
	ReachChangePct       float64 `json:"reach_change_pct"`
}

// AccountDataGraphsRow represents per-account time-series data.
// Buckets are formatted as "YYYY-MM-DD" strings.
type AccountDataGraphsRow struct {
	AccountID   string    `json:"account_id"`
	Engagement  []float64 `json:"engagement"`
	Reach       []float64 `json:"reach"`
	Impressions []float64 `json:"impressions"`
	Posts       []float64 `json:"posts"`
	Buckets     []string  `json:"buckets"`
}

// AIInsightsRequest is the request type for the overviewV2 AI insights endpoint.
type AIInsightsRequest struct {
	OverviewRequest
	Type     string `json:"type"`
	Limit    int    `json:"limit"`
	Language string `json:"language"`
}

// TopPostRow represents a single top-performing post across any platform.
type TopPostRow struct {
	PlatformType    string `json:"platform_type"`
	AccountID       string `json:"account_id"`
	PostID          string `json:"post_id"`
	Likes           int32  `json:"likes"`
	Comments        int32  `json:"comments"`
	Shares          int32  `json:"shares"`
	Saves           int32  `json:"saves"`
	PinClicks       int32  `json:"pin_clicks"`
	OutboundClicks  int32  `json:"outbound_clicks"`
	DislikesCount   int32  `json:"dislikes_count"`
	Permalink       string `json:"permalink"`
	MediaType       string `json:"media_type"`
	Thumbnail       string `json:"thumbnail"`
	Category        string `json:"category"`
	CreatedTime     string `json:"created_time"`
	TotalEngagement int32  `json:"total_engagement"`
	Views           int32  `json:"views"`
	Reach           int32  `json:"reach"`
}
