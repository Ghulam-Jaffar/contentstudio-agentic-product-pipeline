package kafka

import "time"

// LinkedinAccountType represents the type of LinkedIn account
type LinkedinAccountType string

const (
	LinkedinAccountTypePage    LinkedinAccountType = "Page"
	LinkedinAccountTypeProfile LinkedinAccountType = "Profile"
)

// LinkedinAccountWorkOrder represents a single LinkedIn account to process.
// Used both standalone and within batch messages.
type LinkedinAccountWorkOrder struct {
	ID          string              `json:"id"`           // MongoDB _id (hex)
	WorkspaceID string              `json:"workspace_id"` // Workspace identifier
	LinkedinID  string              `json:"linkedin_id"`  // LinkedIn organisation/page ID or person ID
	AccessToken string              `json:"access_token"` // OAuth / long-lived access token
	SyncType    string              `json:"sync_type"`    // "incremental" | "full_sync"
	AccountType LinkedinAccountType `json:"account_type"` // "Page" | "Profile"
}

// LinkedinBatchWorkOrder represents a batch of LinkedIn accounts to process.
// The scheduler produces batch messages to reduce Kafka overhead.
// The fetcher unpacks batches and distributes accounts to worker pools.
type LinkedinBatchWorkOrder struct {
	BatchID     string                     `json:"batch_id"`     // Unique batch identifier (UUID)
	SyncType    string                     `json:"sync_type"`    // "incremental" | "full_sync"
	AccountType LinkedinAccountType        `json:"account_type"` // "Page" | "Profile"
	Accounts    []LinkedinAccountWorkOrder `json:"accounts"`     // List of accounts in this batch (max 200)
	CreatedAt   time.Time                  `json:"created_at"`   // Batch creation timestamp
}

// ParsedLinkedinPost represents a processed LinkedIn post ready for downstream storage.
// Only a subset of the full field list is included for the initial integration – extend as needed.

type ParsedLinkedinPost struct {
	LinkedinID             string    `json:"linkedin_id"`
	PostID                 string    `json:"post_id"`
	Activity               string    `json:"activity"`
	Comments               int64     `json:"comments"`
	TotalEngagement        float64   `json:"total_engagement"`
	Favorites              int64     `json:"favorites"`
	PollData               string    `json:"poll_data"`
	Reach                  int64     `json:"reach"`
	Repost                 int64     `json:"repost"`
	PostClicks             int64     `json:"post_clicks"`
	Impressions            int64     `json:"impressions"`
	Title                  string    `json:"title"`
	Image                  string    `json:"image"`
	ArticleURL             string    `json:"article_url"`
	ArticleTitle           string    `json:"article_title"`
	Media                  []string  `json:"media"`
	MediaType              string    `json:"media_type"`
	Type                   string    `json:"type"`
	Hashtags               []string  `json:"hashtags"`
	DayOfWeek              string    `json:"day_of_week"`
	HourOfDay              int64     `json:"hour_of_day"`
	CreatedAt              time.Time `json:"created_at"`
	PublishedAt            time.Time `json:"published_at"`
	LastModifiedAt         time.Time `json:"last_modified_at"`
	LifecycleState         string    `json:"lifecycle_state"`
	Visibility             string    `json:"visibility"`
	SavingTime             time.Time `json:"saving_time"`
	IsReshareDisabled      bool      `json:"is_reshare_disabled"`
	FeedDistribution       string    `json:"feed_distribution"`
	ThirdPartyChannels     []string  `json:"third_party_channels"`
}

// ParsedLinkedinInsights represents daily org insights (one record per day).

type ParsedLinkedinInsights struct {
	LinkedinID           string    `json:"linkedin_id"`
	OrganizationName     string    `json:"organization_name"`
	RecordID             string    `json:"record_id"`
	ImpressionCount      int64     `json:"impressionCount"`
	OrganicFollowerCount int64     `json:"organicFollowerCount"`
	TotalFollowerCount   int64     `json:"totalFollowerCount"`
	PaidFollowerCount    int64     `json:"paidFollowerCount"`
	DailyFollowerCount   int64     `json:"daily_follower_count"` // Daily follower count for that specific day
	Reach                int64     `json:"reach"`
	Repost               int64     `json:"repost"`
	Comments             int64     `json:"comments"`
	PostClicks           int64     `json:"post_clicks"`
	Reactions            int64     `json:"reactions"`
	Engagement           float64   `json:"engagement"`
	FollowersBySeniority string    `json:"followers_by_seniority"`
	FollowersByIndustry  string    `json:"followers_by_industry"`
	FollowersByCountry   string    `json:"followers_by_country"`
	FollowersByCity      string    `json:"followers_by_city"`
	InsertedAt           time.Time `json:"inserted_at"`
	CreatedAt            time.Time `json:"created_at"` // Date bucket from API timeRange
	// Page view statistics
	PageViews             int64  `json:"page_views"`
	UniqueVisitors        int64  `json:"unique_visitors"` // Used to populate Reach field in ClickHouse
	DesktopPageViews      int64  `json:"desktop_page_views"`
	MobilePageViews       int64  `json:"mobile_page_views"`
	OverviewPageViews     int64  `json:"overview_page_views"`
	AboutPageViews        int64  `json:"about_page_views"`
	JobsPageViews         int64  `json:"jobs_page_views"`
	PeoplePageViews       int64  `json:"people_page_views"`
	CareersPageViews      int64  `json:"careers_page_views"`
	LifeAtPageViews       int64  `json:"life_at_page_views"`
	InsightsPageViews     int64  `json:"insights_page_views"`
	ProductsPageViews     int64  `json:"products_page_views"`
	PageViewsByCountry    string `json:"page_views_by_country"`
	PageViewsByRegion     string `json:"page_views_by_region"`
	PageViewsByIndustry   string `json:"page_views_by_industry"`
	PageViewsBySeniority  string `json:"page_views_by_seniority"`
	PageViewsByFunction   string `json:"page_views_by_function"`
	PageViewsByStaffCount string `json:"page_views_by_staff_count"`
}

// ParsedLinkedinMediaAsset image or video metadata.

type ParsedLinkedinMediaAsset struct {
	ID          string `json:"id"`
	DownloadURL string `json:"downloadUrl"`
	Thumbnail   string `json:"thumbnail,omitempty"`
	Type        string `json:"type"` // image|video
}

// ParsedLinkedinStat extracted from organizationalEntityShareStatistics API.

type ParsedLinkedinStat struct {
	ActivityID             string `json:"activity_id"`
	CommentCount           int64  `json:"commentCount"`
	LikeCount              int64  `json:"likeCount"`
	UniqueImpressionsCount int64  `json:"uniqueImpressionsCount"`
	ShareCount             int64  `json:"shareCount"`
	ClickCount             int64  `json:"clickCount"`
	ImpressionCount        int64  `json:"impressionCount"`
}
