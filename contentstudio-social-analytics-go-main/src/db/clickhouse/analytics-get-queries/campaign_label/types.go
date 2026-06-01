// Package campaign_label provides ClickHouse query builders for campaign/label analytics.
// This file contains query result types returned by repository methods.
package campaign_label

// SummaryResult holds the aggregated summary from the cross-platform ClickHouse query.
type SummaryResult struct {
	TotalPosts                       int32   `json:"total_posts"`
	TotalEngagement                  int32   `json:"total_engagement"`
	TotalImpressions                 int32   `json:"total_impressions"`
	TotalEngagementRatePerImpression float64 `json:"total_engagement_rate_per_impression"`
}

// BreakdownResult holds breakdown data per campaign/label for a specific period.
type BreakdownResult struct {
	ID               string
	Era              string
	TotalPosts       int32
	TotalEngagement  int32
	TotalImpressions int32
}

// InsightsResult holds the time-series insights data per campaign/label.
type InsightsResult struct {
	ID               string
	TotalEngagement  []int32
	TotalImpressions []int32
	TotalPosts       []int32
	CreatedAt        []string
}
