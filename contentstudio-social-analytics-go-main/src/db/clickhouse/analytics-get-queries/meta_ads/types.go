package meta_ads

// ─────────────────────────────────────────────
// Summary (cards)
// ─────────────────────────────────────────────

type SummaryResult struct {
	Spend       float64 `ch:"spend"`
	Reach       int64   `ch:"reach"`
	Impressions int64   `ch:"impressions"`
	Clicks      int64   `ch:"clicks"`
}

// ─────────────────────────────────────────────
// Results by objective
// ─────────────────────────────────────────────

type ObjectiveResultRow struct {
	Objective     string  `ch:"objective"`
	Results       int64   `ch:"results"`
	Spend         float64 `ch:"spend"`
	CampaignCount uint64  `ch:"campaign_count"`
}

// ─────────────────────────────────────────────
// Chart: impressions vs spend / clicks vs ctr
// ─────────────────────────────────────────────

type DailyMetricsRow struct {
	InsightsDate string  `ch:"insights_date"`
	Spend        float64 `ch:"spend"`
	Impressions  int64   `ch:"impressions"`
	Clicks       int64   `ch:"clicks"`
	CTR          float64 `ch:"ctr"`
}

// ─────────────────────────────────────────────
// Top campaigns
// ─────────────────────────────────────────────

type TopCampaignRow struct {
	CampaignID   string  `ch:"campaign_id"`
	CampaignName string  `ch:"campaign_name"`
	Spend        float64 `ch:"spend"`
	Impressions  int64   `ch:"impressions"`
	CTR          float64 `ch:"ctr"`
}

// ─────────────────────────────────────────────
// Performance trend
// ─────────────────────────────────────────────

type TrendRow struct {
	InsightsDate string  `ch:"insights_date"`
	Value        float64 `ch:"value"`
}

// ─────────────────────────────────────────────
// Performance breakdown by level
// ─────────────────────────────────────────────

type LevelBreakdownRow struct {
	ID    string  `ch:"id"`
	Name  string  `ch:"name"`
	Value float64 `ch:"value"`
}

// ─────────────────────────────────────────────
// Performance by publisher platform
// ─────────────────────────────────────────────

type PlatformBreakdownRow struct {
	Platform string  `ch:"platform"`
	Value    float64 `ch:"value"`
}

// ─────────────────────────────────────────────
// Campaign table row
// ─────────────────────────────────────────────

type CampaignTableRow struct {
	CampaignID   string  `ch:"campaign_id"`
	CampaignName string  `ch:"campaign_name"`
	Status       string  `ch:"status"`
	Objective    string  `ch:"objective"`
	Results      int64   `ch:"results"`
	Spend        float64 `ch:"spend"`
	Reach        int64   `ch:"reach"`
	Impressions  int64   `ch:"impressions"`
	Frequency    float64 `ch:"frequency"`
	Clicks       int64   `ch:"clicks"`
	CPM          float64 `ch:"cpm"`
	CPC          float64 `ch:"cpc"`
	CTR          float64 `ch:"ctr"`
}

// ─────────────────────────────────────────────
// AdSet table row
// ─────────────────────────────────────────────

type AdSetTableRow struct {
	AdSetID      string  `ch:"adset_id"`
	AdSetName    string  `ch:"adset_name"`
	CampaignID   string  `ch:"campaign_id"`
	CampaignName string  `ch:"campaign_name"`
	Status       string  `ch:"status"`
	Objective    string  `ch:"objective"`
	Results      int64   `ch:"results"`
	Spend        float64 `ch:"spend"`
	Reach        int64   `ch:"reach"`
	Impressions  int64   `ch:"impressions"`
	Frequency    float64 `ch:"frequency"`
	Clicks       int64   `ch:"clicks"`
	CPM          float64 `ch:"cpm"`
	CPC          float64 `ch:"cpc"`
	CTR          float64 `ch:"ctr"`
}

// ─────────────────────────────────────────────
// Ad table row
// ─────────────────────────────────────────────

type AdTableRow struct {
	AdID                           string  `ch:"ad_id"`
	AdName                         string  `ch:"ad_name"`
	AdSetID                        string  `ch:"adset_id"`
	AdSetName                      string  `ch:"adset_name"`
	CampaignID                     string  `ch:"campaign_id"`
	CampaignName                   string  `ch:"campaign_name"`
	Status                         string  `ch:"status"`
	Objective                      string  `ch:"objective"`
	CreativeName                   string  `ch:"creative_name"`
	CreativeTitle                  string  `ch:"creative_title"`
	CreativeBody                   string  `ch:"creative_body"`
	CreativeThumbnailURL           string  `ch:"creative_thumbnail_url"`
	CreativeEffectiveObjectStoryID string  `ch:"creative_effective_object_story_id"`
	Results                        int64   `ch:"results"`
	Spend                          float64 `ch:"spend"`
	Reach                          int64   `ch:"reach"`
	Impressions                    int64   `ch:"impressions"`
	Frequency                      float64 `ch:"frequency"`
	Clicks                         int64   `ch:"clicks"`
	CPM                            float64 `ch:"cpm"`
	CPC                            float64 `ch:"cpc"`
	CTR                            float64 `ch:"ctr"`
}

// ─────────────────────────────────────────────
// Demographics: age/gender
// ─────────────────────────────────────────────

type AgeGenderRow struct {
	Age         string  `ch:"age"`
	Gender      string  `ch:"gender"`
	Spend       float64 `ch:"spend"`
	Impressions int64   `ch:"impressions"`
	Reach       int64   `ch:"reach"`
	Clicks      int64   `ch:"clicks"`
	CTR         float64 `ch:"ctr"`
	CPM         float64 `ch:"cpm"`
	CPC         float64 `ch:"cpc"`
	Frequency   float64 `ch:"frequency"`
}

// ─────────────────────────────────────────────
// Demographics: region/country
// ─────────────────────────────────────────────

type RegionCountryRow struct {
	Country     string  `ch:"country"`
	Region      string  `ch:"region"`
	Spend       float64 `ch:"spend"`
	Impressions int64   `ch:"impressions"`
	Clicks      int64   `ch:"clicks"`
	CTR         float64 `ch:"ctr"`
}

// CountTotal is used for paginated COUNT queries
type CountTotal struct {
	Total int64 `ch:"total"`
}
