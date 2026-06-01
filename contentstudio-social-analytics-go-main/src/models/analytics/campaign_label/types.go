// Package campaign_label defines request/response types for Campaign & Label analytics endpoints.
// These endpoints aggregate cross-platform analytics data for posts belonging to specific
// campaigns (folders) and labels, querying across Facebook, Instagram, LinkedIn, TikTok,
// YouTube, and Pinterest ClickHouse tables.
//
// Migrated from PHP: CampaignLabelAnalyticsController (contentstudio-backend).
package campaign_label

import (
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// SupportedPlatforms lists the social platforms supported for campaign/label analytics.
// GMB is excluded as it doesn't participate in campaign/label post tracking.
var SupportedPlatforms = []string{
	"facebook", "instagram", "linkedin", "pinterest", "youtube", "tiktok",
}

// CampaignLabelRequest is the base request for campaign/label analytics POST endpoints.
// Uses POST because the payload contains variable-length arrays of campaigns, labels,
// and per-platform account IDs that would be impractical as query parameters.
type CampaignLabelRequest struct {
	WorkspaceID        string   `json:"workspace_id"`
	StartDate          string   `json:"start_date"`
	EndDate            string   `json:"end_date"`
	Date               string   `json:"date"`
	Timezone           string   `json:"timezone"`
	Campaigns          []string `json:"campaigns"`
	Labels             []string `json:"labels"`
	IncludeAllAccounts bool     `json:"include_all_accounts"`
	FacebookAccounts   []string `json:"facebook_accounts"`
	InstagramAccounts  []string `json:"instagram_accounts"`
	TwitterAccounts    []string `json:"twitter_accounts"`
	LinkedinAccounts   []string `json:"linkedin_accounts"`
	PinterestAccounts  []string `json:"pinterest_accounts"`
	YoutubeAccounts    []string `json:"youtube_accounts"`
	TiktokAccounts     []string `json:"tiktok_accounts"`
	GmbAccounts        []string `json:"gmb_accounts"`
	TumblrAccounts     []string `json:"tumblr_accounts"`
	SortBy             string   `json:"sort_by"`
}

// Validate checks that required fields are present and dates are valid.
func (r *CampaignLabelRequest) Validate() error {
	if r.WorkspaceID == "" {
		return httputil.NewValidationError("workspace_id is required")
	}

	// Resolve start_date/end_date with fallback to date param
	r.resolveDates()

	if r.StartDate == "" || r.EndDate == "" {
		return httputil.NewValidationError("start_date and end_date (or date) are required")
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

// resolveDates populates StartDate/EndDate from the "date" field if they are empty.
// The "date" field uses the format "YYYY-MM-DD - YYYY-MM-DD".
// start_date and end_date take priority over date.
func (r *CampaignLabelRequest) resolveDates() {
	if r.StartDate == "" || r.EndDate == "" {
		if r.Date != "" {
			parts := strings.SplitN(strings.TrimSpace(r.Date), " - ", 2)
			if len(parts) == 2 {
				if r.StartDate == "" {
					r.StartDate = strings.TrimSpace(parts[0])
				}
				if r.EndDate == "" {
					r.EndDate = strings.TrimSpace(parts[1])
				}
			}
		}
	}
}

// GetTimezone returns the timezone, defaulting to UTC.
func (r *CampaignLabelRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

// GetDateString returns the date as "YYYY-MM-DD - YYYY-MM-DD" format.
func (r *CampaignLabelRequest) GetDateString() string {
	if r.Date != "" {
		return r.Date
	}
	return r.StartDate + " - " + r.EndDate
}

// ToQueryParams converts the request into ClickHouse query parameters.
func (r *CampaignLabelRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	params, err := clickhouse.ParseStartEndDate(r.StartDate, r.EndDate, r.GetTimezone())
	if err != nil {
		return nil, err
	}
	params.FacebookIDs = append([]string{}, r.FacebookAccounts...)
	return params, nil
}

// GetAllAccountIDs returns all non-GMB account IDs across all platforms.
// GMB is excluded from campaign/label analytics.
func (r *CampaignLabelRequest) GetAllAccountIDs() []string {
	var all []string
	all = append(all, r.FacebookAccounts...)
	all = append(all, r.InstagramAccounts...)
	all = append(all, r.LinkedinAccounts...)
	all = append(all, r.PinterestAccounts...)
	all = append(all, r.YoutubeAccounts...)
	all = append(all, r.TiktokAccounts...)
	return all
}

// GetFlagSetup returns the platform query flags for campaign/label analytics.
// When no accounts are selected, the report should still consider all supported
// platforms instead of collapsing to a single default platform.
func (r *CampaignLabelRequest) GetFlagSetup() map[string]bool {
	hasAnyAccounts := len(r.GetAllAccountIDs()) > 0

	// If no platform accounts are selected, default to all supported platforms.
	// The campaign/label report is driven by campaign/label membership first, so
	// an empty account selection should not collapse the analytics to YouTube only.
	if !hasAnyAccounts {
		return map[string]bool{
			"facebook":  true,
			"instagram": true,
			"linkedin":  true,
			"pinterest": true,
			"youtube":   true,
			"tiktok":    true,
		}
	}

	return map[string]bool{
		"facebook":  len(r.FacebookAccounts) > 0,
		"instagram": len(r.InstagramAccounts) > 0,
		"linkedin":  len(r.LinkedinAccounts) > 0,
		"pinterest": len(r.PinterestAccounts) > 0,
		"youtube":   true, // YouTube always included per PHP logic
		"tiktok":    len(r.TiktokAccounts) > 0,
	}
}

// PlannerAnalyticsRequest is the request for the planner post analytics endpoint.
// Uses POST because it accepts an array of post IDs in the payload.
type PlannerAnalyticsRequest struct {
	WorkspaceID string   `json:"workspace_id"`
	ID          string   `json:"id"`
	AllPostIDs  []string `json:"all_post_ids"`
	Platforms   string   `json:"platforms"`
}

// Validate checks that required fields are present.
func (r *PlannerAnalyticsRequest) Validate() error {
	if r.WorkspaceID == "" {
		return httputil.NewValidationError("workspace_id is required")
	}
	if len(r.AllPostIDs) == 0 {
		return httputil.NewValidationError("all_post_ids is required")
	}
	return nil
}

// SetPostIdsResponse represents the response from the setPostIds endpoint.
type SetPostIdsResponse struct {
	MatchedPostedIds []string `json:"matchedPostedIds"`
}

// SummaryResponse represents the response from getSummaryAnalytics.
// Contains current/previous period data plus computed differences and percentages.
type SummaryResponse struct {
	Current    map[string]interface{} `json:"current"`
	Previous   map[string]interface{} `json:"previous"`
	Difference map[string]interface{} `json:"difference"`
	Percentage map[string]interface{} `json:"percentage"`
}

// BreakdownRow represents a single row in the breakdown data response.
type BreakdownRow struct {
	ID               string `json:"id"`
	Era              string `json:"era"`
	TotalPosts       int32  `json:"total_posts"`
	TotalEngagement  int32  `json:"total_engagement"`
	TotalImpressions int32  `json:"total_impressions"`
}

// InsightsRow represents a single row in the insights breakdown response.
type InsightsRow struct {
	ID               string   `json:"id"`
	TotalEngagement  []int32  `json:"total_engagement"`
	TotalImpressions []int32  `json:"total_impressions"`
	TotalPosts       []int32  `json:"total_posts"`
	CreatedAt        []string `json:"created_at"`
}

// CampaignLabelPostMapping represents cached post IDs for a campaign or label on a platform.
type CampaignLabelPostMapping struct {
	CampaignID   string   `bson:"campaign_id,omitempty"`
	LabelID      string   `bson:"label_id,omitempty"`
	PlatformID   string   `bson:"platform_id"`
	Platform     string   `bson:"platform"`
	PlatformType string   `bson:"platform_type"`
	PostedIDs    []string `bson:"posted_ids"`
}
