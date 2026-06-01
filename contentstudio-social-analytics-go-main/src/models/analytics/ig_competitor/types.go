// Package ig_competitor provides request/response types for Instagram competitor analytics.
// Migrated from PHP InstagramCompetitorController (contentstudio-backend).
package ig_competitor

import (
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
)

// CompetitorRequest holds the parsed query parameters for IG competitor analytics endpoints.
type CompetitorRequest struct {
	ReportID  string `json:"_id"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	Timezone  string `json:"timezone"`
	SortOrder string `json:"sort_order,omitempty"`
	// Single-competitor endpoints
	BusinessAccountID string `json:"business_account_id,omitempty"`
	// postingActivityBySpecificType, postingActivityTableByType
	MediaType        string `json:"media_type,omitempty"`
	MediaProductType string `json:"media_product_type,omitempty"`
	// individualHashtagData
	Hashtag string `json:"hashtag,omitempty"`
	// topHashtags
	Limit int `json:"limit,omitempty"`
}

func (r *CompetitorRequest) Validate() error {
	if r.ReportID == "" && r.BusinessAccountID == "" {
		return httputil.NewValidationError("_id (report_id) or business_account_id is required")
	}
	start, end, err := r.normalizedDates()
	if err != nil {
		return httputil.NewValidationError(err.Error())
	}
	r.StartDate = start
	r.EndDate = end
	if r.Timezone != "" {
		if _, err := time.LoadLocation(r.Timezone); err != nil {
			return httputil.NewValidationError("invalid timezone: " + r.Timezone)
		}
	}
	return nil
}

func (r *CompetitorRequest) GetTimezone() string {
	if r.Timezone == "" {
		return "UTC"
	}
	return r.Timezone
}

func (r *CompetitorRequest) GetLimit() int {
	if r.Limit <= 0 {
		return 7
	}
	return r.Limit
}

func (r *CompetitorRequest) GetSortOrder(defaultOrder string) string {
	if r.SortOrder == "" {
		return defaultOrder
	}
	return r.SortOrder
}

func (r *CompetitorRequest) normalizedDates() (string, string, error) {
	start := strings.TrimSpace(r.StartDate)
	end := strings.TrimSpace(r.EndDate)
	if start == "" && end == "" {
		return "", "", nil
	}
	if start == "" || end == "" {
		return "", "", fmt.Errorf("both start_date and end_date must be provided together")
	}
	if _, err := time.Parse("2006-01-02", start); err != nil {
		return "", "", fmt.Errorf("start_date must be in YYYY-MM-DD format")
	}
	if _, err := time.Parse("2006-01-02", end); err != nil {
		return "", "", fmt.Errorf("end_date must be in YYYY-MM-DD format")
	}
	if end < start {
		return "", "", fmt.Errorf("end_date cannot be before start_date")
	}
	return start, end, nil
}

// ToUTCDateRange converts local dates + timezone into UTC start/end for ClickHouse queries.
func (r *CompetitorRequest) ToUTCDateRange() (string, string, int, error) {
	if r.StartDate == "" && r.EndDate == "" {
		return "", "", 0, nil
	}

	loc, err := time.LoadLocation(r.GetTimezone())
	if err != nil {
		loc = time.UTC
	}

	startLocal, _ := time.ParseInLocation("2006-01-02", r.StartDate, loc)
	startLocal = startLocal.Add(1 * time.Second)
	startUTC := startLocal.UTC()

	endLocal, _ := time.ParseInLocation("2006-01-02", r.EndDate, loc)
	endLocal = time.Date(endLocal.Year(), endLocal.Month(), endLocal.Day(), 23, 59, 59, 0, loc)
	endUTC := endLocal.UTC()

	daysDiff := int(endUTC.Sub(startUTC).Hours()/24) + 1

	return startUTC.Format("2006-01-02 15:04:05"), endUTC.Format("2006-01-02 15:04:05"), daysDiff, nil
}
