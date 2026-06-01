// Package pinterest defines request and response types for the Pinterest analytics API.
// These types map directly to the JSON contracts expected by the ContentStudio frontend,
// preserving the same field names and structure as the PHP Laravel API responses.
//
// Request types include validation logic and conversion to ClickHouse query parameters.
// Response types match the frontend's expected JSON shape for each analytics widget.
//
// Pinterest supports two modes: user mode (no BoardID) and board mode (BoardID present).
// Use HasBoard() to determine which repository methods to call.
package pinterest

import (
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// --- Request types ---

// PinterestRequest is the base request for all Pinterest analytics endpoints.
// Query params: workspace_id, pinterest_id, board_id (optional), start_date, end_date, timezone.
type PinterestRequest struct {
	WorkspaceID string `json:"workspace_id"`
	PinterestID string `json:"pinterest_id"`
	BoardID     string `json:"board_id,omitempty"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Timezone    string `json:"timezone"`
}

// FilteredPinRequest extends PinterestRequest with a media type filter for pin posting queries.
type FilteredPinRequest struct {
	PinterestRequest
	FilterBy string `json:"filter_by,omitempty"`
}

// TopPinsRequest extends PinterestRequest with pagination and sort options.
type TopPinsRequest struct {
	PinterestRequest
	Limit   int    `json:"limit"`
	OrderBy string `json:"order_by"`
}

// AIInsightsRequest is the request for Pinterest AI insights.
type AIInsightsRequest struct {
	WorkspaceID string      `json:"workspace_id"`
	PinterestID string      `json:"pinterest_id"`
	BoardID     string      `json:"board_id,omitempty"`
	Date        interface{} `json:"date"`
	Timezone    string      `json:"timezone"`
	Type        string      `json:"type"`
	Limit       int         `json:"limit"`
	Language    string      `json:"language,omitempty"`
}

// --- Response types ---

// SummaryResponse wraps current/previous/percentage/difference metrics under an "overview" key.
type SummaryResponse struct {
	Status   bool             `json:"status"`
	Overview *SummaryOverview `json:"overview"`
}

// SummaryOverview groups current/previous metrics with percentage changes and absolute differences.
type SummaryOverview struct {
	Current    *SummaryMetrics       `json:"current"`
	Previous   *SummaryMetrics       `json:"previous"`
	Percentage *SummaryChangeMetrics `json:"percentage"`
	Difference *SummaryMetrics       `json:"difference"`
}

// SummaryMetrics holds the aggregate account-level metrics for one period.
type SummaryMetrics struct {
	FollowerCount   int64 `json:"follower_count"`
	Impressions     int64 `json:"impressions"`
	PinClicks       int64 `json:"pin_clicks"`
	OutboundClicks  int64 `json:"outbound_clicks"`
	Saves           int64 `json:"saves"`
	TotalEngagement int64 `json:"total_engagement"`
}

// SummaryChangeMetrics holds percentage changes for each summary metric.
type SummaryChangeMetrics struct {
	FollowerCount   float64 `json:"follower_count"`
	Impressions     float64 `json:"impressions"`
	PinClicks       float64 `json:"pin_clicks"`
	OutboundClicks  float64 `json:"outbound_clicks"`
	Saves           float64 `json:"saves"`
	TotalEngagement float64 `json:"total_engagement"`
}

// FollowerTrendResponse is the response for the follower trend widget.
type FollowerTrendResponse struct {
	Status           bool     `json:"status"`
	ShowData         int32    `json:"show_data"`
	FollowersDaily   []int32  `json:"followers_daily"`
	FollowersGained  []int32  `json:"followers_gained"`
	Buckets          []string `json:"buckets"`
	AggregationLevel string   `json:"aggregation_level,omitempty"`
}

// ImpressionsTrendResponse is the response for the impressions trend widget.
type ImpressionsTrendResponse struct {
	Status           bool     `json:"status"`
	ShowData         int32    `json:"show_data"`
	ImpressionsDaily []int32  `json:"impressions_daily"`
	ImpressionsTotal []int32  `json:"impressions_total"`
	Buckets          []string `json:"buckets"`
	AggregationLevel string   `json:"aggregation_level,omitempty"`
}

// EngagementTrendResponse is the response for the engagement trend widget.
type EngagementTrendResponse struct {
	Status              bool     `json:"status"`
	ShowData            int32    `json:"show_data"`
	SavesDaily          []int32  `json:"saves_daily"`
	SavesTotal          []int32  `json:"saves_total"`
	OutboundClicksDaily []int32  `json:"outbound_clicks_daily"`
	OutboundClicksTotal []int32  `json:"outbound_clicks_total"`
	PinClicksDaily      []int32  `json:"pin_clicks_daily"`
	PinClicksTotal      []int32  `json:"pin_clicks_total"`
	EngagementDaily     []int32  `json:"engagement_daily"`
	EngagementTotal     []int32  `json:"engagement_total"`
	Buckets             []string `json:"buckets"`
	AggregationLevel    string   `json:"aggregation_level,omitempty"`
}

// PinPostingResponse is the response for the pin posting frequency widget.
type PinPostingResponse struct {
	Status           bool     `json:"status"`
	ShowData         int32    `json:"show_data"`
	PinsCount        []int32  `json:"pins_count"`
	Buckets          []string `json:"buckets"`
	AggregationLevel string   `json:"aggregation_level,omitempty"`
}

// PinRollupResponse wraps current/previous/percentage/difference metrics under an "overview" key.
type PinRollupResponse struct {
	Status   bool               `json:"status"`
	Overview *PinRollupOverview `json:"overview"`
}

// PinRollupOverview groups current/previous metrics with percentage changes and absolute differences.
type PinRollupOverview struct {
	Current    *PinRollupMetrics       `json:"current"`
	Previous   *PinRollupMetrics       `json:"previous"`
	Percentage *PinRollupChangeMetrics `json:"percentage"`
	Difference *PinRollupMetrics       `json:"difference"`
}

// PinRollupChangeMetrics holds percentage changes for each pin rollup metric.
type PinRollupChangeMetrics struct {
	TotalPins        float64 `json:"total_pins"`
	Impressions      float64 `json:"impressions"`
	PinClicks        float64 `json:"pin_clicks"`
	OutboundClicks   float64 `json:"outbound_clicks"`
	Saves            float64 `json:"saves"`
	QuartilePercView float64 `json:"quartile_95s_percent_view"`
	VideoViews       float64 `json:"video_views"`
	Video10sViews    float64 `json:"video_10s_view"`
	AvgWatchTime     float64 `json:"avg_watch_time"`
}

// PinRollupMetrics holds the aggregate pin-level metrics for one period.
type PinRollupMetrics struct {
	TotalPins        int64   `json:"total_pins"`
	Impressions      int64   `json:"impressions"`
	PinClicks        int64   `json:"pin_clicks"`
	OutboundClicks   int64   `json:"outbound_clicks"`
	Saves            int64   `json:"saves"`
	QuartilePercView float64 `json:"quartile_95s_percent_view"`
	VideoViews       int64   `json:"video_views"`
	Video10sViews    int64   `json:"video_10s_view"`
	AvgWatchTime     float64 `json:"avg_watch_time"`
}

// TopPinsResponse holds the top and least performing pin lists.
type TopPinsResponse struct {
	Status bool      `json:"status"`
	Top    []PinItem `json:"top"`
	Least  []PinItem `json:"least"`
}

// PinItem holds per-pin display and metrics data for the top/least pins table.
type PinItem struct {
	PinID           string   `json:"pin_id"`
	BoardName       string   `json:"board_name"`
	Permalink       string   `json:"permalink"`
	EmbedLink       string   `json:"embed_link"`
	Title           string   `json:"title"`
	Description     string   `json:"description"`
	BoardOwner      string   `json:"board_owner"`
	MediaType       string   `json:"media_type"`
	CoverImageURL   string   `json:"cover_image_url"`
	DominantColor   string   `json:"dominant_color"`
	CreativeType    string   `json:"creative_type"`
	ProductTags     []string `json:"product_tags"`
	Height          int64    `json:"height"`
	Width           int64    `json:"width"`
	CreatedAt       string   `json:"created_at"`
	Impressions     int64    `json:"impressions"`
	PinClicks       int64    `json:"pin_clicks"`
	OutboundClicks  int64    `json:"outbound_clicks"`
	Saves           int64    `json:"saves"`
	TotalEngagement int64    `json:"total_engagement"`
	EngagementRate  float64  `json:"engagement_rate"`
}

// PinPerformanceResponse is the response for the pin performance over time widget.
type PinPerformanceResponse struct {
	Status         bool     `json:"status"`
	ShowData       int32    `json:"show_data"`
	PinsCount      []int32  `json:"pins_count"`
	PinClicks      []int32  `json:"pin_clicks"`
	OutboundClicks []int32  `json:"outbound_clicks"`
	Saves          []int32  `json:"saves"`
	Engagements    []int32  `json:"engagements"`
	Impressions    []int32  `json:"impressions"`
	Buckets        []string `json:"buckets"`
}

// --- Request helpers ---

// Validate checks all required fields and validates date formats.
func (r *PinterestRequest) Validate() error {
	if r.PinterestID == "" {
		return httputil.NewValidationError("pinterest_id is required")
	}
	if r.StartDate == "" {
		return httputil.NewValidationError("start_date is required")
	}
	if r.EndDate == "" {
		return httputil.NewValidationError("end_date is required")
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

// GetTimezone returns the requested timezone, defaulting to UTC.
func (r *PinterestRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

// ToQueryParams converts the request into ClickHouse query parameters.
func (r *PinterestRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	params, err := clickhouse.ParseStartEndDate(r.StartDate, r.EndDate, r.GetTimezone())
	if err != nil {
		return nil, err
	}
	params.AccountIDs = []string{r.PinterestID}
	return params, nil
}

// HasBoard returns true when the request targets a specific board rather than the whole account.
func (r *PinterestRequest) HasBoard() bool {
	return r.BoardID != ""
}

// FormatBoardIDs returns the BoardID formatted as a SQL IN clause suitable for ClickHouse queries.
// Supports comma-separated board IDs for multi-board queries.
func (r *PinterestRequest) FormatBoardIDs() string {
	if r.BoardID == "" {
		return "('')"
	}
	parts := strings.Split(r.BoardID, ",")
	quoted := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			quoted = append(quoted, fmt.Sprintf("'%s'", strings.ReplaceAll(p, "'", "\\'")))
		}
	}
	if len(quoted) == 0 {
		return "('')"
	}
	return "(" + strings.Join(quoted, ",") + ")"
}
