package clickhouse

import "time"

// LinkedInPosts ClickHouse model (flat struct for bulk insert)
// Column order must match table definition in ClickHouse.

type LinkedInPosts struct {
	LinkedinID         string    `ch:"linkedin_id"`
	PostID             string    `ch:"post_id"`
	Activity           string    `ch:"activity"`
	Comments           int64     `ch:"comments"`
	TotalEngagement    float64   `ch:"total_engagement"`
	Favorites          int64     `ch:"favorites"`
	PollData           string    `ch:"poll_data"`
	Reach              int64     `ch:"reach"`
	Repost             int64     `ch:"repost"`
	PostClicks         int64     `ch:"post_clicks"`
	Impressions        int64     `ch:"impressions"`
	Title              string    `ch:"title"`
	Image              string    `ch:"image"`
	ArticleURL         string    `ch:"article_url"`
	ArticleTitle       string    `ch:"article_title"`
	Media              []string  `ch:"media"`
	MediaType          string    `ch:"media_type"`
	Type               string    `ch:"type"`
	Hashtags           []string  `ch:"hashtags"`
	DayOfWeek          string    `ch:"day_of_week"`
	HourOfDay          int       `ch:"hour_of_day"`
	CreatedAt          time.Time `ch:"created_at"`
	PublishedAt        time.Time `ch:"published_at"`
	LastModifiedAt     time.Time `ch:"last_modified_at"`
	LifecycleState     string    `ch:"lifecycle_state"`
	Visibility         string    `ch:"visibility"`
	SavingTime         time.Time `ch:"saving_time"`
	IsReshareDisabled  bool      `ch:"is_reshare_disabled"`
	FeedDistribution   string    `ch:"feed_distribution"`
	ThirdPartyChannels []string  `ch:"third_party_channels"`
}

type LinkedInMinimalPost struct {
	LinkedinID string   `ch:"linkedin_id" json:"linkedin_id"`
	PostID     string   `ch:"post_id" json:"post_id"`
	Activity   string   `ch:"activity" json:"activity"`
	Image      string   `ch:"image" json:"image"`
	Media      []string `ch:"media" json:"media"`
}

// LinkedInInsights ClickHouse struct (daily buckets)

type LinkedInInsights struct {
	LinkedinID           string    `ch:"linkedin_id"`
	OrganizationName     string    `ch:"organization_name"`
	RecordID             string    `ch:"record_id"`
	ImpressionCount      int64     `ch:"impressionCount"`
	OrganicFollowerCount int64     `ch:"organicFollowerCount"`
	TotalFollowerCount   int64     `ch:"totalFollowerCount"`
	PaidFollowerCount    int64     `ch:"paidFollowerCount"`
	DailyFollowerCount   int64     `ch:"daily_follower_count"` // Daily follower count for that specific day
	Reach                int64     `ch:"reach"`
	Repost               int64     `ch:"repost"`
	Comments             int64     `ch:"comments"`
	PostClicks           int64     `ch:"post_clicks"`
	Reactions            int64     `ch:"reactions"`
	Engagement           float64   `ch:"engagement"`
	FollowersBySeniority string    `ch:"followers_by_seniority"`
	FollowersByIndustry  string    `ch:"followers_by_industry"`
	FollowersByCountry   string    `ch:"followers_by_country"`
	FollowersByCity      string    `ch:"followers_by_city"`
	InsertedAt           time.Time `ch:"inserted_at"`
	CreatedAt            time.Time `ch:"created_at"` // Date bucket from API timeRange
	// Page view statistics
	PageViews             int64  `ch:"page_views"`
	UniqueVisitors        int64  `ch:"unique_visitors"` // UniquePageViews from page stats
	DesktopPageViews      int64  `ch:"desktop_page_views"`
	MobilePageViews       int64  `ch:"mobile_page_views"`
	OverviewPageViews     int64  `ch:"overview_page_views"`
	AboutPageViews        int64  `ch:"about_page_views"`
	JobsPageViews         int64  `ch:"jobs_page_views"`
	PeoplePageViews       int64  `ch:"people_page_views"`
	CareersPageViews      int64  `ch:"careers_page_views"`
	LifeAtPageViews       int64  `ch:"life_at_page_views"`
	InsightsPageViews     int64  `ch:"insights_page_views"`
	ProductsPageViews     int64  `ch:"products_page_views"`
	PageViewsByCountry    string `ch:"page_views_by_country"`
	PageViewsByRegion     string `ch:"page_views_by_region"`
	PageViewsByIndustry   string `ch:"page_views_by_industry"`
	PageViewsBySeniority  string `ch:"page_views_by_seniority"`
	PageViewsByFunction   string `ch:"page_views_by_function"`
	PageViewsByStaffCount string `ch:"page_views_by_staff_count"`
}

// LinkedInMediaAsset ClickHouse model

type LinkedInMediaAsset struct {
	ID          string    `ch:"id"`
	DownloadURL string    `ch:"download_url"`
	Thumbnail   string    `ch:"thumbnail"`
	Type        string    `ch:"type"`
	SavingTime  time.Time `ch:"saving_time"`
}

// LinkedInStat ClickHouse model

type LinkedInStat struct {
	ActivityID             string    `ch:"activity_id"`
	CommentCount           int64     `ch:"comment_count"`
	LikeCount              int64     `ch:"like_count"`
	UniqueImpressionsCount int64     `ch:"unique_impressions_count"`
	ShareCount             int64     `ch:"share_count"`
	ClickCount             int64     `ch:"click_count"`
	ImpressionCount        int64     `ch:"impression_count"`
	SavingTime             time.Time `ch:"saving_time"`
}
