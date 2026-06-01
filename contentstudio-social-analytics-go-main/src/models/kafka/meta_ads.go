package kafka

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Work Orders
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsWorkOrder is the unit of work produced to the
// "work-order-meta-ads" Kafka topic by the account fetcher and to
// "immediate-work-order-meta-ads" by the API server for on-demand jobs.
type MetaAdsWorkOrder struct {
	// MongoID is the _id of the social_integration document.
	MongoID string `json:"id"`
	// PlatformIdentifier is the act_XXXX identifier (used as ad account ID).
	PlatformIdentifier string `json:"platform_identifier"`
	// AccountID is the numeric ad account ID without the "act_" prefix.
	AccountID string `json:"account_id"`
	// AccessToken is the (optionally encrypted) user access token.
	AccessToken string `json:"access_token"`
	// LongAccessToken is the (optionally encrypted) long-lived user access token.
	LongAccessToken string `json:"long_access_token,omitempty"`
	// WorkspaceID is the ContentStudio workspace that owns the account.
	WorkspaceID string `json:"workspace_id"`
	// UserID is the ContentStudio user who added the account.
	UserID string `json:"user_id"`
	// SyncType is "immediate" or "scheduled".
	SyncType string `json:"sync_type"`
	// StartDate and EndDate are optional overrides for the date range (YYYY-MM-DD).
	StartDate string `json:"start_date,omitempty"`
	EndDate   string `json:"end_date,omitempty"`
}

// MetaAdsBatchWorkOrder wraps a slice of MetaAdsWorkOrder for batch delivery
// (work-order-meta-ads topic).
type MetaAdsBatchWorkOrder struct {
	BatchID  string             `json:"batch_id"`
	Accounts []MetaAdsWorkOrder `json:"accounts"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Kafka payloads produced by the fetcher (one per endpoint)
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsAccountInfoPayload is published to "raw-meta-ads-account-info".
type MetaAdsAccountInfoPayload struct {
	WorkOrder   MetaAdsWorkOrder      `json:"work_order"`
	AccountInfo RawMetaAdsAccountInfo `json:"account_info"`
}

// MetaAdsCampaignsPayload is published (in batches of 500) to "raw-meta-ads-campaigns".
type MetaAdsCampaignsPayload struct {
	WorkOrder MetaAdsWorkOrder     `json:"work_order"`
	Campaigns []RawMetaAdsCampaign `json:"campaigns"`
}

// MetaAdsAdsetsPayload is published (in batches of 500) to "raw-meta-ads-adsets".
type MetaAdsAdsetsPayload struct {
	WorkOrder MetaAdsWorkOrder  `json:"work_order"`
	Adsets    []RawMetaAdsAdset `json:"adsets"`
}

// MetaAdsAdsPayload is published (in batches of 500) to "raw-meta-ads-ads".
type MetaAdsAdsPayload struct {
	WorkOrder MetaAdsWorkOrder `json:"work_order"`
	Ads       []RawMetaAdsAd   `json:"ads"`
}

// MetaAdsCampaignInsightsPayload is published (in batches of 500) to "raw-meta-ads-campaign-insights".
type MetaAdsCampaignInsightsPayload struct {
	WorkOrder MetaAdsWorkOrder       `json:"work_order"`
	Insights  []RawMetaAdsInsightRow `json:"insights"`
}

// MetaAdsAdsetInsightsPayload is published (in batches of 500) to "raw-meta-ads-adset-insights".
type MetaAdsAdsetInsightsPayload struct {
	WorkOrder MetaAdsWorkOrder       `json:"work_order"`
	Insights  []RawMetaAdsInsightRow `json:"insights"`
}

// MetaAdsAdInsightsPayload is published (in batches of 500) to "raw-meta-ads-ad-insights".
type MetaAdsAdInsightsPayload struct {
	WorkOrder MetaAdsWorkOrder       `json:"work_order"`
	Insights  []RawMetaAdsInsightRow `json:"insights"`
}

// MetaAdsDemographicsAgeGenderPayload is published (in batches of 500) to
// "raw-meta-ads-demographics-age-gender".
type MetaAdsDemographicsAgeGenderPayload struct {
	WorkOrder MetaAdsWorkOrder            `json:"work_order"`
	Rows      []RawMetaAdsDemographicsRow `json:"rows"`
}

// MetaAdsDemographicsDevicePlatformPayload is published (in batches of 500) to
// "raw-meta-ads-demographics-device-platform".
type MetaAdsDemographicsDevicePlatformPayload struct {
	WorkOrder MetaAdsWorkOrder            `json:"work_order"`
	Rows      []RawMetaAdsDemographicsRow `json:"rows"`
}

// MetaAdsDemographicsRegionCountryPayload is published (in batches of 500) to
// "raw-meta-ads-demographics-region-country".
type MetaAdsDemographicsRegionCountryPayload struct {
	WorkOrder MetaAdsWorkOrder            `json:"work_order"`
	Rows      []RawMetaAdsDemographicsRow `json:"rows"`
}

// ─────────────────────────────────────────────────────────────────────────────
// Raw API response models
// ─────────────────────────────────────────────────────────────────────────────

// RawMetaAdsAccountInfo maps the /act_{id} response.
type RawMetaAdsAccountInfo struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	Currency      string `json:"currency"`
	AccountStatus int32  `json:"account_status"`
	TimezoneName  string `json:"timezone_name"`
	Business      *struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"business,omitempty"`
	AmountSpent string         `json:"amount_spent"`
	Balance     string         `json:"balance"`
	SpendCap    string         `json:"spend_cap"`
	CreatedTime MetaAdsAPITime `json:"created_time"`
}

// RawMetaAdsCampaign maps one element of the /act_{id}/campaigns response.
type RawMetaAdsCampaign struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	Status          string         `json:"status"`
	EffectiveStatus string         `json:"effective_status"`
	Objective       string         `json:"objective"`
	DailyBudget     string         `json:"daily_budget"`
	LifetimeBudget  string         `json:"lifetime_budget"`
	BudgetRemaining string         `json:"budget_remaining"`
	StartTime       MetaAdsAPITime `json:"start_time"`
	StopTime        MetaAdsAPITime `json:"stop_time"`
	CreatedTime     MetaAdsAPITime `json:"created_time"`
	UpdatedTime     MetaAdsAPITime `json:"updated_time"`
}

// RawMetaAdsAdsetTargeting captures only the fields we need from the targeting object.
type RawMetaAdsAdsetTargeting struct {
	AgeMin       int32 `json:"age_min"`
	AgeMax       int32 `json:"age_max"`
	GeoLocations *struct {
		Countries []string `json:"countries"`
	} `json:"geo_locations,omitempty"`
}

// RawMetaAdsAdset maps one element of the /act_{id}/adsets response.
type RawMetaAdsAdset struct {
	ID               string                    `json:"id"`
	Name             string                    `json:"name"`
	CampaignID       string                    `json:"campaign_id"`
	Status           string                    `json:"status"`
	EffectiveStatus  string                    `json:"effective_status"`
	DailyBudget      string                    `json:"daily_budget"`
	LifetimeBudget   string                    `json:"lifetime_budget"`
	BudgetRemaining  string                    `json:"budget_remaining"`
	BillingEvent     string                    `json:"billing_event"`
	OptimizationGoal string                    `json:"optimization_goal"`
	BidStrategy      string                    `json:"bid_strategy"`
	Targeting        *RawMetaAdsAdsetTargeting `json:"targeting,omitempty"`
	// RawTargeting holds the full JSON of the targeting object for storage.
	RawTargeting json.RawMessage `json:"targeting_raw,omitempty"`
	StartTime    MetaAdsAPITime  `json:"start_time"`
	StopTime     MetaAdsAPITime  `json:"stop_time"`
	EndTime      MetaAdsAPITime  `json:"end_time"`
	CreatedTime  MetaAdsAPITime  `json:"created_time"`
}

// MarshalJSON for RawMetaAdsAdset — captures raw targeting for storage.
func (a *RawMetaAdsAdset) UnmarshalJSON(data []byte) error {
	type Alias RawMetaAdsAdset
	aux := &struct {
		Targeting json.RawMessage `json:"targeting"`
		*Alias
	}{
		Alias: (*Alias)(a),
	}
	if err := json.Unmarshal(data, aux); err != nil {
		return err
	}
	a.RawTargeting = aux.Targeting
	if len(aux.Targeting) > 0 {
		var t RawMetaAdsAdsetTargeting
		if err := json.Unmarshal(aux.Targeting, &t); err == nil {
			a.Targeting = &t
		}
	}
	return nil
}

// RawMetaAdsAd maps one element of the /act_{id}/ads response.
type RawMetaAdsAd struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	AdsetID         string         `json:"adset_id"`
	CampaignID      string         `json:"campaign_id"`
	Status          string         `json:"status"`
	EffectiveStatus string         `json:"effective_status"`
	DailyBudget     string         `json:"daily_budget"`
	LifetimeBudget  string         `json:"lifetime_budget"`
	BudgetRemaining string         `json:"budget_remaining"`
	CreatedTime     MetaAdsAPITime `json:"created_time"`
	UpdatedTime     MetaAdsAPITime `json:"updated_time"`

	// Expanded edge objects
	Adset *struct {
		Name string `json:"name"`
	} `json:"adset,omitempty"`
	Campaign *struct {
		Name      string `json:"name"`
		Objective string `json:"objective"`
	} `json:"campaign,omitempty"`
	Creative *struct {
		ID                     string `json:"id"`
		Name                   string `json:"name"`
		Title                  string `json:"title"`
		Body                   string `json:"body"`
		ImageURL               string `json:"image_url"`
		ThumbnailURL           string `json:"thumbnail_url"`
		ObjectType             string `json:"object_type"`
		EffectiveObjectStoryID string `json:"effective_object_story_id"`
	} `json:"creative,omitempty"`
}

// RawMetaAdsAction represents one element of the actions[] array in insights responses.
type RawMetaAdsAction struct {
	ActionType string `json:"action_type"`
	Value      string `json:"value"`
}

// RawMetaAdsInsightRow maps one row of the /act_{id}/insights response
// (all levels share the same row structure; level-specific fields will be
// populated when present).
type RawMetaAdsInsightRow struct {
	// Level-specific IDs/names (populated depending on level= parameter)
	CampaignID   string `json:"campaign_id,omitempty"`
	CampaignName string `json:"campaign_name,omitempty"`
	Objective    string `json:"objective,omitempty"`
	AdsetID      string `json:"adset_id,omitempty"`
	AdsetName    string `json:"adset_name,omitempty"`
	AdID         string `json:"ad_id,omitempty"`
	AdName       string `json:"ad_name,omitempty"`

	// Core delivery metrics (returned as strings by the API)
	Spend        string `json:"spend"`
	Impressions  string `json:"impressions"`
	Reach        string `json:"reach"`
	Clicks       string `json:"clicks"`
	UniqueClicks string `json:"unique_clicks"`
	CTR          string `json:"ctr"`
	UniqueCTR    string `json:"unique_ctr"`
	CPC          string `json:"cpc"`
	CPM          string `json:"cpm"`
	CPP          string `json:"cpp"`
	Frequency    string `json:"frequency"`

	// Date (same when time_increment=1)
	DateStart string `json:"date_start"`
	DateStop  string `json:"date_stop"`

	// Actions array
	Actions []RawMetaAdsAction `json:"actions,omitempty"`
}

// RawMetaAdsDemographicsRow maps one row of the /act_{id}/insights demographics response.
type RawMetaAdsDemographicsRow struct {
	// Breakdown dimensions (only populated for the relevant breakdown)
	Age               string `json:"age,omitempty"`
	Gender            string `json:"gender,omitempty"`
	ImpressionDevice  string `json:"impression_device,omitempty"`
	PublisherPlatform string `json:"publisher_platform,omitempty"`
	PlatformPosition  string `json:"platform_position,omitempty"`
	Country           string `json:"country,omitempty"`
	Region            string `json:"region,omitempty"`

	// Metrics
	Impressions string `json:"impressions"`
	Reach       string `json:"reach"`
	Clicks      string `json:"clicks"`
	Spend       string `json:"spend"`
	CTR         string `json:"ctr"`
	CPM         string `json:"cpm"`
	CPC         string `json:"cpc"`
	CPP         string `json:"cpp"`
	Frequency   string `json:"frequency"`

	// Date
	DateStart string `json:"date_start"`
	DateStop  string `json:"date_stop"`
}

// PaginatedResponse is a generic wrapper for paginated Graph API list responses.
type PaginatedResponse[T any] struct {
	Data   []T `json:"data"`
	Paging struct {
		Cursors struct {
			Before string `json:"before"`
			After  string `json:"after"`
		} `json:"cursors"`
		Next string `json:"next"`
	} `json:"paging"`
}

// ─────────────────────────────────────────────────────────────────────────────
// MetaAdsAPITime — custom time type for Meta Ads API timestamps.
// Meta returns timestamps like "2025-07-21T14:59:17+0500".
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsAPITime handles Meta's non-standard "+0500" timezone offset format.
type MetaAdsAPITime struct {
	time.Time
}

// IsZero returns true if the underlying time is zero.
func (t MetaAdsAPITime) IsZero() bool { return t.Time.IsZero() }

// UnmarshalJSON implements json.Unmarshaler for MetaAdsAPITime.
func (t *MetaAdsAPITime) UnmarshalJSON(data []byte) error {
	s := strings.Trim(string(data), `"`)
	if s == "" || s == "null" {
		return nil
	}
	// Normalise "+HHMM" → "+HH:MM" for time.Parse.
	if len(s) >= 5 {
		suffix := s[len(s)-5:]
		if (suffix[0] == '+' || suffix[0] == '-') && strings.IndexByte(suffix, ':') == -1 {
			s = s[:len(s)-5] + string(suffix[0]) + suffix[1:3] + ":" + suffix[3:5]
		}
	}
	for _, layout := range []string{
		"2006-01-02T15:04:05-07:00",
		time.RFC3339,
		"2006-01-02",
	} {
		if parsed, err := time.Parse(layout, s); err == nil {
			t.Time = parsed.UTC()
			return nil
		}
	}
	return fmt.Errorf("MetaAdsAPITime.UnmarshalJSON: cannot parse %q", string(data))
}

// MarshalJSON implements json.Marshaler for MetaAdsAPITime.
func (t MetaAdsAPITime) MarshalJSON() ([]byte, error) {
	if t.Time.IsZero() {
		return []byte("null"), nil
	}
	return json.Marshal(t.Time.Format(time.RFC3339))
}
