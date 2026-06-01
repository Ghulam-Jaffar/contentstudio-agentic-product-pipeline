package meta_ads

import (
	"fmt"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// MetaAdsRequest is the common request struct for all meta ads analytics endpoints.
type MetaAdsRequest struct {
	WorkspaceID string
	AccountID   string // single ad account ID
	StartDate   string
	EndDate     string
	Timezone    string
	Language    string
	Currency    string
	// Pagination
	Page    int
	PerPage int
	// Filtering & sorting
	Status    string
	Objective string
	Search    string
	OrderBy   string
	OrderDir  string
	// Performance & demographics selectors
	Metric    string // e.g. spend, impressions, clicks, cpm, cpc, ctr, frequency
	Breakdown string // region or country
	Country   string // country filter for region table
	Level     string // campaign | adset | ad (for performance breakdown)
	SortBy    string // top campaigns sort: spend | impressions | ctr
}

// Validate returns a ValidationError if required fields are missing.
func (r *MetaAdsRequest) Validate() error {
	if r.WorkspaceID == "" {
		return httputil.NewValidationError("workspace_id is required")
	}
	if r.AccountID == "" {
		return httputil.NewValidationError("account_id is required")
	}
	if r.StartDate == "" || r.EndDate == "" {
		return httputil.NewValidationError("start_date and end_date are required")
	}
	return nil
}

// ToQueryParams converts the request to ClickHouse QueryParams.
func (r *MetaAdsRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	params, err := clickhouse.ParseStartEndDate(r.StartDate, r.EndDate, r.Timezone)
	if err != nil {
		return nil, fmt.Errorf("%w", httputil.NewValidationError(err.Error()))
	}
	params.AccountIDs = []string{r.AccountID}
	return params, nil
}

// ─────────────────────────────────────────────
// Overview: Summary / Cards
// ─────────────────────────────────────────────

type MetricValue struct {
	Current  float64 `json:"current"`
	Previous float64 `json:"previous"`
	Change   float64 `json:"change"` // percentage change
}

type SummaryResponse struct {
	Status      bool        `json:"status"`
	Spend       MetricValue `json:"spend"`
	Reach       MetricValue `json:"reach"`
	Impressions MetricValue `json:"impressions"`
	Clicks      MetricValue `json:"clicks"`
	CPM         MetricValue `json:"cpm"`
	CPC         MetricValue `json:"cpc"`
	CTR         MetricValue `json:"ctr"`
}

// ─────────────────────────────────────────────
// Overview: Results by Objective
// ─────────────────────────────────────────────

type ObjectiveResult struct {
	Objective     string  `json:"objective"`
	Label         string  `json:"label"`
	Results       int64   `json:"results"`
	ResultsPrev   int64   `json:"results_prev"`
	ResultsChange float64 `json:"results_change"`
	CostPerResult float64 `json:"cost_per_result"`
	CampaignCount uint64  `json:"campaign_count"`
}

type ResultsByObjectiveResponse struct {
	Status bool              `json:"status"`
	Data   []ObjectiveResult `json:"data"`
}

// ─────────────────────────────────────────────
// Overview: Charts
// ─────────────────────────────────────────────

type TimeSeriesPoint struct {
	Date  string  `json:"date"`
	Value float64 `json:"value"`
}

type ImpressionsVsSpendResponse struct {
	Status      bool      `json:"status"`
	Dates       []string  `json:"dates"`
	Impressions []float64 `json:"impressions"`
	Spend       []float64 `json:"spend"`
}

type ClicksVsCTRResponse struct {
	Status bool      `json:"status"`
	Dates  []string  `json:"dates"`
	Clicks []float64 `json:"clicks"`
	CTR    []float64 `json:"ctr"`
}

// ─────────────────────────────────────────────
// Overview: Top Campaigns
// ─────────────────────────────────────────────

type TopCampaignRow struct {
	CampaignID   string  `json:"campaign_id"`
	CampaignName string  `json:"campaign_name"`
	Spend        float64 `json:"spend"`
	Impressions  int64   `json:"impressions"`
	CTR          float64 `json:"ctr"`
}

type TopCampaignsResponse struct {
	Status bool             `json:"status"`
	Data   []TopCampaignRow `json:"data"`
}

// ─────────────────────────────────────────────
// Performance: Trend
// ─────────────────────────────────────────────

type PerformanceTrendResponse struct {
	Status bool      `json:"status"`
	Metric string    `json:"metric"`
	Dates  []string  `json:"dates"`
	Values []float64 `json:"values"`
	Total  float64   `json:"total"`
}

// ─────────────────────────────────────────────
// Performance: By Level (campaign/adset/ad)
// ─────────────────────────────────────────────

type PerformanceLevelRow struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Value float64 `json:"value"`
}

type PerformanceByLevelResponse struct {
	Status  bool                  `json:"status"`
	Level   string                `json:"level"`
	Metric  string                `json:"metric"`
	Data    []PerformanceLevelRow `json:"data"`
	HasMore bool                  `json:"has_more"`
}

// ─────────────────────────────────────────────
// Performance: By Platform
// ─────────────────────────────────────────────

type PlatformBreakdownRow struct {
	Platform string  `json:"platform"`
	Value    float64 `json:"value"`
	Percent  float64 `json:"percent"`
}

type PerformanceByPlatformResponse struct {
	Status bool                   `json:"status"`
	Metric string                 `json:"metric"`
	Total  float64                `json:"total"`
	Data   []PlatformBreakdownRow `json:"data"`
}

// ─────────────────────────────────────────────
// Campaigns / AdSets / Ads table
// ─────────────────────────────────────────────

type TableRow struct {
	ID                             string  `json:"id"`
	Name                           string  `json:"name"`
	ParentID                       string  `json:"parent_id,omitempty"`
	ParentName                     string  `json:"parent_name,omitempty"`
	GrandParentID                  string  `json:"grand_parent_id,omitempty"`
	GrandParentName                string  `json:"grand_parent_name,omitempty"`
	Status                         string  `json:"status"`
	Objective                      string  `json:"objective"`
	CreativeName                   string  `json:"creative_name,omitempty"`
	CreativeTitle                  string  `json:"creative_title,omitempty"`
	CreativeBody                   string  `json:"creative_body,omitempty"`
	CreativeThumbnailURL           string  `json:"creative_thumbnail_url,omitempty"`
	CreativeEffectiveObjectStoryID string  `json:"creative_effective_object_story_id,omitempty"`
	Results                        int64   `json:"results"`
	Spend                          float64 `json:"spend"`
	Reach                          int64   `json:"reach"`
	Impressions                    int64   `json:"impressions"`
	Frequency                      float64 `json:"frequency"`
	Clicks                         int64   `json:"clicks"`
	CPM                            float64 `json:"cpm"`
	CPC                            float64 `json:"cpc"`
	CTR                            float64 `json:"ctr"`
}

type TableResponse struct {
	Status  bool       `json:"status"`
	Data    []TableRow `json:"data"`
	Total   int64      `json:"total"`
	Page    int        `json:"page"`
	PerPage int        `json:"per_page"`
	// Filter options (dynamic)
	AvailableStatuses   []string `json:"available_statuses"`
	AvailableObjectives []string `json:"available_objectives"`
}

// ─────────────────────────────────────────────
// Demographics: Age & Gender
// ─────────────────────────────────────────────

type AgeBreakdownRow struct {
	AgeRange string  `json:"age_range"`
	Value    float64 `json:"value"`
	Percent  float64 `json:"percent"`
}

type GenderBreakdownRow struct {
	Gender  string  `json:"gender"`
	Value   float64 `json:"value"`
	Percent float64 `json:"percent"`
	Count   int64   `json:"count"`
}

type DemographicsAgeGenderResponse struct {
	Status   bool                 `json:"status"`
	Metric   string               `json:"metric"`
	ByAge    []AgeBreakdownRow    `json:"by_age"`
	ByGender []GenderBreakdownRow `json:"by_gender"`
}

// ─────────────────────────────────────────────
// Demographics: Region / Country
// ─────────────────────────────────────────────

type RegionCountryRow struct {
	Country     string  `json:"country"`
	Region      string  `json:"region,omitempty"`
	Spend       float64 `json:"spend"`
	Impressions int64   `json:"impressions"`
	Clicks      int64   `json:"clicks"`
	CTR         float64 `json:"ctr"`
}

type DemographicsRegionCountryResponse struct {
	Status    bool               `json:"status"`
	Breakdown string             `json:"breakdown"` // "region" or "country"
	Data      []RegionCountryRow `json:"data"`
	// Available countries for region filter dropdown
	AvailableCountries []string `json:"available_countries,omitempty"`
}
