package tiktok

import (
	"fmt"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

const maxLimit = 100

// TiktokRequest is the base request for TikTok analytics endpoints.
type TiktokRequest struct {
	WorkspaceID string `json:"workspace_id"`
	TiktokID    string `json:"tiktok_id"`
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Timezone    string `json:"timezone"`
}

func (r *TiktokRequest) Validate() error {
	if r.WorkspaceID == "" {
		return httputil.NewValidationError("workspace_id is required")
	}
	if r.TiktokID == "" {
		return httputil.NewValidationError("tiktok_id is required")
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

func (r *TiktokRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

func (r *TiktokRequest) ToQueryParams() (*clickhouse.QueryParams, error) {
	params, err := clickhouse.ParseStartEndDate(r.StartDate, r.EndDate, r.GetTimezone())
	if err != nil {
		return nil, err
	}
	params.AccountIDs = []string{r.TiktokID}
	return params, nil
}

type PostsRequest struct {
	TiktokRequest
	Limit     int    `json:"limit"`
	Offset    int    `json:"offset"`
	SortOrder string `json:"sort_order"`
}

func (r *PostsRequest) GetLimit() int {
	if r.Limit <= 0 {
		return 5
	}
	if r.Limit > maxLimit {
		return maxLimit
	}
	return r.Limit
}

func (r *PostsRequest) GetOffset() int {
	if r.Offset < 0 {
		return 0
	}
	return r.Offset
}

var validSortFields = map[string]bool{
	"total_engagement":  true,
	"engagements_count": true,
	"likes_count":       true,
	"comments_count":    true,
	"shares_count":      true,
	"views_count":       true,
	"created_time":      true,
}

func (r *PostsRequest) GetSortOrder() string {
	if r.SortOrder == "" || !validSortFields[r.SortOrder] {
		return "total_engagement"
	}
	return r.SortOrder
}

type AIInsightsRequest struct {
	WorkspaceID string      `json:"workspace_id"`
	TiktokID    string      `json:"tiktok_id"`
	Date        interface{} `json:"date"`
	Timezone    string      `json:"timezone"`
	Type        string      `json:"type"`
	Limit       int         `json:"limit"`
	Language    string      `json:"language,omitempty"`
}

func BuildDateTimeRange(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid start_date: %w", err)
	}
	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("invalid end_date: %w", err)
	}
	return start.UTC(), end.Add(24*time.Hour - time.Second).UTC(), nil
}
