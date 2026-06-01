// Package pinterest provides the ClickHouse repository layer for Pinterest analytics.
// It contains query result types and repository methods that execute ClickHouse SQL queries
// migrated from the PHP Laravel PinterestAnalyticsBuilder (contentstudio-backend).
//
// Tables queried: pinterest_pins, pinterest_pin_insights, pinterest_user_insights,
// pinterest_users, pinterest_boards
package pinterest

import "time"

// SummaryResult holds aggregated account-level metrics for the Pinterest summary widget.
// Supports both user-mode (pinterest_user_insights) and board-mode (pinterest_pin_insights).
type SummaryResult struct {
	FollowerCount   int64 `ch:"follower_count"`
	Impressions     int64 `ch:"impressions"`
	PinClicks       int64 `ch:"pin_clicks"`
	OutboundClicks  int64 `ch:"outbound_clicks"`
	Saves           int64 `ch:"saves"`
	TotalEngagement int64 `ch:"total_engagement"`
}

// FollowerTrendResult holds time-series follower data with daily deltas and cumulative totals.
type FollowerTrendResult struct {
	ShowData        uint8       `ch:"show_data"`
	FollowersDaily  []int32     `ch:"followers_daily"`
	FollowersGained []int32     `ch:"followers_gained"`
	Buckets         []time.Time `ch:"buckets"`
}

// ImpressionsTrendResult holds time-series impressions data with daily and cumulative totals.
type ImpressionsTrendResult struct {
	ShowData         uint8       `ch:"show_data"`
	ImpressionsTotal []int32     `ch:"impressions_total"`
	ImpressionsDaily []int32     `ch:"impressions_daily"`
	Buckets          []time.Time `ch:"buckets"`
}

// EngagementTrendResult holds time-series engagement metrics broken down by type
// with both daily and cumulative (running total) arrays for each metric.
type EngagementTrendResult struct {
	ShowData            uint8       `ch:"show_data"`
	SavesDaily          []int32     `ch:"saves_daily"`
	SavesTotal          []int32     `ch:"saves_total"`
	OutboundClicksDaily []int32     `ch:"outbound_clicks_daily"`
	OutboundClicksTotal []int32     `ch:"outbound_clicks_total"`
	PinClicksDaily      []int32     `ch:"pin_clicks_daily"`
	PinClicksTotal      []int32     `ch:"pin_clicks_total"`
	EngagementDaily     []int32     `ch:"engagement_daily"`
	EngagementTotal     []int32     `ch:"engagement_total"`
	Buckets             []time.Time `ch:"buckets"`
}

// PinPostingResult holds time-series pin publication counts, optionally filtered by media type.
type PinPostingResult struct {
	ShowData  uint8       `ch:"show_data"`
	PinsCount []int32     `ch:"pins_count"`
	Buckets   []time.Time `ch:"buckets"`
}

// PinRollupResult holds aggregated pin-level metrics for the pin rollup widget.
// Combines data from pinterest_pins and pinterest_pin_insights.
type PinRollupResult struct {
	TotalPins        int64   `ch:"total_pins"`
	Impressions      int64   `ch:"impressions"`
	PinClicks        int64   `ch:"pin_clicks"`
	OutboundClicks   int64   `ch:"outbound_clicks"`
	Saves            int64   `ch:"saves"`
	QuartilePercView float64 `ch:"quartile_95s_percent_view"`
	VideoViews       int64   `ch:"video_views"`
	Video10sViews    int64   `ch:"video_10s_views"`
	AvgWatchTime     float64 `ch:"avg_watch_time"`
}

// PinRow holds per-pin metrics for the top/least pins table widget.
// Joins pinterest_pins, pinterest_pin_insights, and pinterest_boards.
type PinRow struct {
	PinID           string    `ch:"pin_id"`
	BoardName       string    `ch:"board_name"`
	Permalink       string    `ch:"permalink"`
	EmbedLink       string    `ch:"embed_link"`
	Title           string    `ch:"title"`
	Description     string    `ch:"description"`
	BoardOwner      string    `ch:"board_owner"`
	MediaType       string    `ch:"media_type"`
	CoverImageURL   string    `ch:"cover_image_url"`
	DominantColor   string    `ch:"dominant_color"`
	CreativeType    string    `ch:"creative_type"`
	ProductTags     []string  `ch:"product_tags"`
	Height          int64     `ch:"height"`
	Width           int64     `ch:"width"`
	CreatedAt       time.Time `ch:"created_at"`
	Impressions     int64     `ch:"impressions"`
	PinClicks       int64     `ch:"pin_clicks"`
	OutboundClicks  int64     `ch:"outbound_clicks"`
	Saves           int64     `ch:"saves"`
	TotalEngagement int64     `ch:"total_engagement"`
	EngagementRate  float64   `ch:"engagement_rate"`
}

// PinPerformanceResult holds daily time-series metrics across all pins for the performance widget.
// Each array position corresponds to a day bucket, showing aggregated metrics across all pins published that day.
type PinPerformanceResult struct {
	ShowData       uint8       `ch:"show_data"`
	PinsCount      []int32     `ch:"pins_count"`
	PinClicks      []int32     `ch:"pin_clicks"`
	OutboundClicks []int32     `ch:"outbound_clicks"`
	Saves          []int32     `ch:"saves"`
	Engagements    []int32     `ch:"engagements"`
	Impressions    []int32     `ch:"impressions"`
	Buckets        []time.Time `ch:"buckets"`
}
