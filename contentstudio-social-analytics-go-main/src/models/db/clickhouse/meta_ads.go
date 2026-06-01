package clickhouse

import "time"

// ─────────────────────────────────────────────────────────────────────────────
// 1. Ad Account Info
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsAccountInfo represents the ClickHouse table schema for a Meta Ads
// account. Corresponds to the /act_{id} endpoint response.
type MetaAdsAccountInfo struct {
	AccountID     string    `ch:"account_id"    json:"account_id"`
	Name          string    `ch:"name"          json:"name"`
	Currency      string    `ch:"currency"      json:"currency"`
	AccountStatus int32     `ch:"account_status" json:"account_status"`
	TimezoneName  string    `ch:"timezone_name" json:"timezone_name"`
	BusinessID    string    `ch:"business_id"   json:"business_id"`
	BusinessName  string    `ch:"business_name" json:"business_name"`
	AmountSpent   string    `ch:"amount_spent"  json:"amount_spent"`
	Balance       string    `ch:"balance"       json:"balance"`
	SpendCap      string    `ch:"spend_cap"     json:"spend_cap"`
	CreatedTime   time.Time `ch:"created_time"  json:"created_time"`
	InsertedAt    time.Time `ch:"inserted_at"   json:"inserted_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 2. Campaigns
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsCampaign represents the ClickHouse table schema for Meta Ads campaigns.
// Corresponds to the /act_{id}/campaigns list endpoint.
type MetaAdsCampaign struct {
	AccountID       string    `ch:"account_id"       json:"account_id"`
	CampaignID      string    `ch:"campaign_id"      json:"campaign_id"`
	Name            string    `ch:"name"             json:"name"`
	Status          string    `ch:"status"           json:"status"`
	EffectiveStatus string    `ch:"effective_status" json:"effective_status"`
	Objective       string    `ch:"objective"        json:"objective"`
	DailyBudget     string    `ch:"daily_budget"     json:"daily_budget"`
	LifetimeBudget  string    `ch:"lifetime_budget"  json:"lifetime_budget"`
	BudgetRemaining string    `ch:"budget_remaining" json:"budget_remaining"`
	StartTime       time.Time `ch:"start_time"       json:"start_time"`
	StopTime        time.Time `ch:"stop_time"        json:"stop_time"`
	CreatedTime     time.Time `ch:"created_time"     json:"created_time"`
	UpdatedTime     time.Time `ch:"updated_time"     json:"updated_time"`
	InsertedAt      time.Time `ch:"inserted_at"      json:"inserted_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 3. Ad Sets
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsAdset represents the ClickHouse table schema for Meta Ads ad sets.
// Corresponds to the /act_{id}/adsets list endpoint.
// targeting_countries holds targeting.geo_locations.countries.
// targeting_json holds the full raw targeting object as JSON for future use.
type MetaAdsAdset struct {
	AccountID          string    `ch:"account_id"           json:"account_id"`
	AdsetID            string    `ch:"adset_id"             json:"adset_id"`
	Name               string    `ch:"name"                 json:"name"`
	CampaignID         string    `ch:"campaign_id"          json:"campaign_id"`
	Status             string    `ch:"status"               json:"status"`
	EffectiveStatus    string    `ch:"effective_status"     json:"effective_status"`
	DailyBudget        string    `ch:"daily_budget"         json:"daily_budget"`
	LifetimeBudget     string    `ch:"lifetime_budget"      json:"lifetime_budget"`
	BudgetRemaining    string    `ch:"budget_remaining"     json:"budget_remaining"`
	BillingEvent       string    `ch:"billing_event"        json:"billing_event"`
	OptimizationGoal   string    `ch:"optimization_goal"    json:"optimization_goal"`
	BidStrategy        string    `ch:"bid_strategy"         json:"bid_strategy"`
	AgeMin             int32     `ch:"age_min"              json:"age_min"`
	AgeMax             int32     `ch:"age_max"              json:"age_max"`
	TargetingCountries []string  `ch:"targeting_countries"  json:"targeting_countries"`
	TargetingJSON      string    `ch:"targeting_json"       json:"targeting_json"`
	StartTime          time.Time `ch:"start_time"           json:"start_time"`
	StopTime           time.Time `ch:"stop_time"            json:"stop_time"`
	EndTime            time.Time `ch:"end_time"             json:"end_time"`
	CreatedTime        time.Time `ch:"created_time"         json:"created_time"`
	InsertedAt         time.Time `ch:"inserted_at"          json:"inserted_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 4. Ads
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsAd represents the ClickHouse table schema for Meta Ads ads.
// Corresponds to the /act_{id}/ads list endpoint.
// Creative fields are flattened from the nested creative{} object.
type MetaAdsAd struct {
	AccountID                      string    `ch:"account_id"              json:"account_id"`
	AdID                           string    `ch:"ad_id"                   json:"ad_id"`
	Name                           string    `ch:"name"                    json:"name"`
	AdsetID                        string    `ch:"adset_id"                json:"adset_id"`
	AdsetName                      string    `ch:"adset_name"              json:"adset_name"`
	CampaignID                     string    `ch:"campaign_id"             json:"campaign_id"`
	CampaignName                   string    `ch:"campaign_name"           json:"campaign_name"`
	Status                         string    `ch:"status"                  json:"status"`
	EffectiveStatus                string    `ch:"effective_status"        json:"effective_status"`
	Objective                      string    `ch:"objective"               json:"objective"`
	CreativeID                     string    `ch:"creative_id"             json:"creative_id"`
	CreativeName                   string    `ch:"creative_name"           json:"creative_name"`
	CreativeTitle                  string    `ch:"creative_title"          json:"creative_title"`
	CreativeBody                   string    `ch:"creative_body"           json:"creative_body"`
	CreativeImageURL               string    `ch:"creative_image_url"      json:"creative_image_url"`
	CreativeThumbnailURL           string    `ch:"creative_thumbnail_url"  json:"creative_thumbnail_url"`
	CreativeObjectType             string    `ch:"creative_object_type"    json:"creative_object_type"`
	CreativeEffectiveObjectStoryID string    `ch:"creative_effective_object_story_id" json:"creative_effective_object_story_id"`
	DailyBudget                    string    `ch:"daily_budget"            json:"daily_budget"`
	LifetimeBudget                 string    `ch:"lifetime_budget"         json:"lifetime_budget"`
	BudgetRemaining                string    `ch:"budget_remaining"        json:"budget_remaining"`
	CreatedTime                    time.Time `ch:"created_time"            json:"created_time"`
	UpdatedTime                    time.Time `ch:"updated_time"            json:"updated_time"`
	InsertedAt                     time.Time `ch:"inserted_at"             json:"inserted_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 5. Campaign Insights  (level=campaign, time_increment=1)
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsCampaignInsights represents daily campaign-level performance metrics.
// insights_date corresponds to date_start = date_stop (time_increment=1).
// Action fields are flattened from the actions[] array keyed by action_type.
// Outbound / video metrics are flattened from their respective response arrays.
type MetaAdsCampaignInsights struct {
	AccountID    string    `ch:"account_id"    json:"account_id"`
	CampaignID   string    `ch:"campaign_id"   json:"campaign_id"`
	CampaignName string    `ch:"campaign_name" json:"campaign_name"`
	Objective    string    `ch:"objective"     json:"objective"`
	InsightsDate time.Time `ch:"insights_date" json:"insights_date"`

	// Core delivery metrics
	Spend        float64 `ch:"spend"         json:"spend"`
	Impressions  int64   `ch:"impressions"   json:"impressions"`
	Reach        int64   `ch:"reach"         json:"reach"`
	Clicks       int64   `ch:"clicks"        json:"clicks"`
	UniqueClicks int64   `ch:"unique_clicks" json:"unique_clicks"`
	CTR          float64 `ch:"ctr"           json:"ctr"`
	UniqueCTR    float64 `ch:"unique_ctr"    json:"unique_ctr"`
	CPC          float64 `ch:"cpc"           json:"cpc"`
	CPM          float64 `ch:"cpm"           json:"cpm"`
	CPP          float64 `ch:"cpp"           json:"cpp"`
	Frequency    float64 `ch:"frequency"     json:"frequency"`

	// Flattened actions[] by action_type
	ActionsPurchase                         int64 `ch:"actions_purchase"                               json:"actions_purchase"`
	ActionsPostEngagement                   int64 `ch:"actions_post_engagement"                        json:"actions_post_engagement"`
	ActionsOffsiteConversionFbPixelPurchase int64 `ch:"actions_offsite_conversion_fb_pixel_purchase"   json:"actions_offsite_conversion_fb_pixel_purchase"`
	ActionsLinkClick                        int64 `ch:"actions_link_click"                             json:"actions_link_click"`
	ActionsLead                             int64 `ch:"actions_lead"                                   json:"actions_lead"`
	ActionsOffsiteConversionFbPixelLead     int64 `ch:"actions_offsite_conversion_fb_pixel_lead"       json:"actions_offsite_conversion_fb_pixel_lead"`
	ActionsMobileAppInstall                 int64 `ch:"actions_mobile_app_install"                     json:"actions_mobile_app_install"`

	InsertedAt time.Time `ch:"inserted_at" json:"inserted_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 6. Adset Insights  (level=adset, time_increment=1)
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsAdsetInsights represents daily ad-set-level performance metrics.
type MetaAdsAdsetInsights struct {
	AccountID    string    `ch:"account_id"    json:"account_id"`
	AdsetID      string    `ch:"adset_id"      json:"adset_id"`
	AdsetName    string    `ch:"adset_name"    json:"adset_name"`
	CampaignID   string    `ch:"campaign_id"   json:"campaign_id"`
	CampaignName string    `ch:"campaign_name" json:"campaign_name"`
	InsightsDate time.Time `ch:"insights_date" json:"insights_date"`

	Spend        float64 `ch:"spend"         json:"spend"`
	Impressions  int64   `ch:"impressions"   json:"impressions"`
	Reach        int64   `ch:"reach"         json:"reach"`
	Clicks       int64   `ch:"clicks"        json:"clicks"`
	UniqueClicks int64   `ch:"unique_clicks" json:"unique_clicks"`
	CTR          float64 `ch:"ctr"           json:"ctr"`
	UniqueCTR    float64 `ch:"unique_ctr"    json:"unique_ctr"`
	CPC          float64 `ch:"cpc"           json:"cpc"`
	CPM          float64 `ch:"cpm"           json:"cpm"`
	CPP          float64 `ch:"cpp"           json:"cpp"`
	Frequency    float64 `ch:"frequency"     json:"frequency"`

	ActionsPurchase                         int64 `ch:"actions_purchase"                               json:"actions_purchase"`
	ActionsPostEngagement                   int64 `ch:"actions_post_engagement"                        json:"actions_post_engagement"`
	ActionsOffsiteConversionFbPixelPurchase int64 `ch:"actions_offsite_conversion_fb_pixel_purchase"   json:"actions_offsite_conversion_fb_pixel_purchase"`
	ActionsLinkClick                        int64 `ch:"actions_link_click"                             json:"actions_link_click"`
	ActionsLead                             int64 `ch:"actions_lead"                                   json:"actions_lead"`
	ActionsOffsiteConversionFbPixelLead     int64 `ch:"actions_offsite_conversion_fb_pixel_lead"       json:"actions_offsite_conversion_fb_pixel_lead"`
	ActionsMobileAppInstall                 int64 `ch:"actions_mobile_app_install"                     json:"actions_mobile_app_install"`

	InsertedAt time.Time `ch:"inserted_at" json:"inserted_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 7. Ad Insights  (level=ad, time_increment=1)
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsAdInsights represents daily ad-level performance metrics.
type MetaAdsAdInsights struct {
	AccountID    string    `ch:"account_id"    json:"account_id"`
	AdID         string    `ch:"ad_id"         json:"ad_id"`
	AdName       string    `ch:"ad_name"       json:"ad_name"`
	AdsetID      string    `ch:"adset_id"      json:"adset_id"`
	CampaignID   string    `ch:"campaign_id"   json:"campaign_id"`
	CampaignName string    `ch:"campaign_name" json:"campaign_name"`
	InsightsDate time.Time `ch:"insights_date" json:"insights_date"`

	Spend        float64 `ch:"spend"         json:"spend"`
	Impressions  int64   `ch:"impressions"   json:"impressions"`
	Reach        int64   `ch:"reach"         json:"reach"`
	Clicks       int64   `ch:"clicks"        json:"clicks"`
	UniqueClicks int64   `ch:"unique_clicks" json:"unique_clicks"`
	CTR          float64 `ch:"ctr"           json:"ctr"`
	UniqueCTR    float64 `ch:"unique_ctr"    json:"unique_ctr"`
	CPC          float64 `ch:"cpc"           json:"cpc"`
	CPM          float64 `ch:"cpm"           json:"cpm"`
	CPP          float64 `ch:"cpp"           json:"cpp"`
	Frequency    float64 `ch:"frequency"     json:"frequency"`

	ActionsPurchase                         int64 `ch:"actions_purchase"                               json:"actions_purchase"`
	ActionsPostEngagement                   int64 `ch:"actions_post_engagement"                        json:"actions_post_engagement"`
	ActionsOffsiteConversionFbPixelPurchase int64 `ch:"actions_offsite_conversion_fb_pixel_purchase"   json:"actions_offsite_conversion_fb_pixel_purchase"`
	ActionsLinkClick                        int64 `ch:"actions_link_click"                             json:"actions_link_click"`
	ActionsLead                             int64 `ch:"actions_lead"                                   json:"actions_lead"`
	ActionsOffsiteConversionFbPixelLead     int64 `ch:"actions_offsite_conversion_fb_pixel_lead"       json:"actions_offsite_conversion_fb_pixel_lead"`
	ActionsMobileAppInstall                 int64 `ch:"actions_mobile_app_install"                     json:"actions_mobile_app_install"`

	InsertedAt time.Time `ch:"inserted_at" json:"inserted_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 8. Demographics: Age & Gender  (breakdowns=age,gender, time_increment=1)
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsDemographicsAgeGender represents daily age/gender breakdown metrics.
type MetaAdsDemographicsAgeGender struct {
	AccountID    string    `ch:"account_id"    json:"account_id"`
	InsightsDate time.Time `ch:"insights_date" json:"insights_date"`
	Age          string    `ch:"age"           json:"age"`
	Gender       string    `ch:"gender"        json:"gender"`
	Impressions  int64     `ch:"impressions"   json:"impressions"`
	Reach        int64     `ch:"reach"         json:"reach"`
	Clicks       int64     `ch:"clicks"        json:"clicks"`
	Spend        float64   `ch:"spend"         json:"spend"`
	CTR          float64   `ch:"ctr"           json:"ctr"`
	CPM          float64   `ch:"cpm"           json:"cpm"`
	CPC          float64   `ch:"cpc"           json:"cpc"`
	CPP          float64   `ch:"cpp"           json:"cpp"`
	Frequency    float64   `ch:"frequency"     json:"frequency"`
	InsertedAt   time.Time `ch:"inserted_at"   json:"inserted_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 9. Demographics: Device & Platform
//    (breakdowns=impression_device,publisher_platform,platform_position, time_increment=1)
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsDemographicsDevicePlatform represents daily device/platform breakdown metrics.
type MetaAdsDemographicsDevicePlatform struct {
	AccountID         string    `ch:"account_id"          json:"account_id"`
	InsightsDate      time.Time `ch:"insights_date"       json:"insights_date"`
	ImpressionDevice  string    `ch:"impression_device"   json:"impression_device"`
	PublisherPlatform string    `ch:"publisher_platform"  json:"publisher_platform"`
	PlatformPosition  string    `ch:"platform_position"   json:"platform_position"`
	Impressions       int64     `ch:"impressions"         json:"impressions"`
	Reach             int64     `ch:"reach"               json:"reach"`
	Clicks            int64     `ch:"clicks"              json:"clicks"`
	Spend             float64   `ch:"spend"               json:"spend"`
	CTR               float64   `ch:"ctr"                 json:"ctr"`
	CPM               float64   `ch:"cpm"                 json:"cpm"`
	CPC               float64   `ch:"cpc"                 json:"cpc"`
	CPP               float64   `ch:"cpp"                 json:"cpp"`
	Frequency         float64   `ch:"frequency"           json:"frequency"`
	InsertedAt        time.Time `ch:"inserted_at"         json:"inserted_at"`
}

// ─────────────────────────────────────────────────────────────────────────────
// 10. Demographics: Region & Country  (breakdowns=region,country, time_increment=1)
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsDemographicsRegionCountry represents daily region/country breakdown metrics.
type MetaAdsDemographicsRegionCountry struct {
	AccountID    string    `ch:"account_id"    json:"account_id"`
	InsightsDate time.Time `ch:"insights_date" json:"insights_date"`
	Country      string    `ch:"country"       json:"country"`
	Region       string    `ch:"region"        json:"region"`
	Impressions  int64     `ch:"impressions"   json:"impressions"`
	Reach        int64     `ch:"reach"         json:"reach"`
	Clicks       int64     `ch:"clicks"        json:"clicks"`
	Spend        float64   `ch:"spend"         json:"spend"`
	CTR          float64   `ch:"ctr"           json:"ctr"`
	CPM          float64   `ch:"cpm"           json:"cpm"`
	CPC          float64   `ch:"cpc"           json:"cpc"`
	CPP          float64   `ch:"cpp"           json:"cpp"`
	Frequency    float64   `ch:"frequency"     json:"frequency"`
	InsertedAt   time.Time `ch:"inserted_at"   json:"inserted_at"`
}
