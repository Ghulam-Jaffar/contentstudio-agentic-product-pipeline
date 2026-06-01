// helpers.go provides shared SQL building utilities for analytics queries.
// These helpers are used across all platform repositories to generate
// date filters, account ID formatting, post deduplication CTEs, and more.
// Patterns are migrated from the PHP AnalyticsBuilder base class.

package clickhouse

import (
	"fmt"
	"strings"
	"time"
)

// QueryParams holds common analytics query parameters shared across all platforms.
type QueryParams struct {
	AccountIDs   []string
	FacebookIDs  []string
	DateFrom     time.Time
	DateTo       time.Time
	PrevDateFrom time.Time
	PrevDateTo   time.Time
	Timezone     string
	DayCount     int
}

// ParseDateRange parses a "YYYY-MM-DD - YYYY-MM-DD" date string and computes the previous period.
// The previous period is calculated by shifting the date range backwards by the same number of days.
// Example: "2025-01-15 - 2025-02-14" (31 days) → previous: "2024-12-15 - 2025-01-15"
func ParseDateRange(dateStr, timezone string) (*QueryParams, error) {
	if dateStr == "" {
		return nil, fmt.Errorf("date parameter is required")
	}

	parts := strings.SplitN(dateStr, " - ", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid date format, expected 'YYYY-MM-DD - YYYY-MM-DD', got: %s", dateStr)
	}

	dateFrom, err := time.Parse("2006-01-02", strings.TrimSpace(parts[0]))
	if err != nil {
		return nil, fmt.Errorf("invalid start date: %w", err)
	}

	dateTo, err := time.Parse("2006-01-02", strings.TrimSpace(parts[1]))
	if err != nil {
		return nil, fmt.Errorf("invalid end date: %w", err)
	}

	if dateTo.Before(dateFrom) {
		return nil, fmt.Errorf("end date cannot be before start date")
	}

	if timezone == "" {
		timezone = "UTC"
	}

	dayCount := int(dateTo.Sub(dateFrom).Hours()/24) + 1
	diff := dateTo.Sub(dateFrom)
	prevDateTo := dateFrom.AddDate(0, 0, -1)
	prevDateFrom := dateFrom.Add(-diff)

	return &QueryParams{
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		PrevDateFrom: prevDateFrom,
		PrevDateTo:   prevDateTo,
		Timezone:     timezone,
		DayCount:     dayCount,
	}, nil
}

// ParseStartEndDate parses separate start_date and end_date strings and computes the previous period.
func ParseStartEndDate(startDate, endDate, timezone string) (*QueryParams, error) {
	if startDate == "" || endDate == "" {
		return nil, fmt.Errorf("start_date and end_date are required")
	}

	dateFrom, err := time.Parse("2006-01-02", strings.TrimSpace(startDate))
	if err != nil {
		return nil, fmt.Errorf("invalid start_date: %w", err)
	}

	dateTo, err := time.Parse("2006-01-02", strings.TrimSpace(endDate))
	if err != nil {
		return nil, fmt.Errorf("invalid end_date: %w", err)
	}

	if dateTo.Before(dateFrom) {
		return nil, fmt.Errorf("end_date cannot be before start_date")
	}

	if timezone == "" {
		timezone = "UTC"
	}

	dayCount := int(dateTo.Sub(dateFrom).Hours()/24) + 1
	diff := dateTo.Sub(dateFrom)
	prevDateTo := dateFrom.AddDate(0, 0, -1)
	prevDateFrom := dateFrom.Add(-diff)

	return &QueryParams{
		DateFrom:     dateFrom,
		DateTo:       dateTo,
		PrevDateFrom: prevDateFrom,
		PrevDateTo:   prevDateTo,
		Timezone:     timezone,
		DayCount:     dayCount,
	}, nil
}

// FormatAccountIDs formats account IDs for use in SQL IN clauses.
// Returns a string like: ('id1','id2','id3')
func FormatAccountIDs(ids []string) string {
	if len(ids) == 0 {
		return "('')"
	}
	quoted := make([]string, len(ids))
	for i, id := range ids {
		quoted[i] = "'" + strings.ReplaceAll(id, "'", "\\'") + "'"
	}
	return "(" + strings.Join(quoted, ",") + ")"
}

// DateFilter returns a SQL WHERE clause fragment for date filtering with timezone support.
// Converts the field to a Date in the request timezone so the boundary strings are
// compared in the same reference frame — avoids the UTC-vs-local mismatch that
// occurred when the field was a local DateTime but boundaries were parsed as UTC.
func DateFilter(field string, params *QueryParams) string {
	return fmt.Sprintf(
		"toDate(%s, '%s') >= '%s' AND toDate(%s, '%s') <= '%s'",
		field, params.Timezone, params.DateFrom.Format("2006-01-02"),
		field, params.Timezone, params.DateTo.Format("2006-01-02"),
	)
}

// PartitionMonthFilter returns a WHERE condition on the partition key field that enables
// ClickHouse to prune monthly partitions before applying the semantic date filter.
// Use this alongside DateFilter when the table is partitioned by a different column
// than the one used for semantic filtering (e.g. saving_time vs created_time).
// Since both columns are set at insert time their month values are identical, so this
// filter never excludes valid rows — it only helps ClickHouse skip irrelevant partitions.
func PartitionMonthFilter(partitionField string, params *QueryParams) string {
	return fmt.Sprintf(
		"toYYYYMM(%s) >= %s AND toYYYYMM(%s) <= %s",
		partitionField, params.DateFrom.Format("200601"),
		partitionField, params.DateTo.Format("200601"),
	)
}

// PrevDateFilter returns the same date filter but for the previous period.
func PrevDateFilter(field string, params *QueryParams) string {
	return fmt.Sprintf(
		"toDate(%s, '%s') >= '%s' AND toDate(%s, '%s') <= '%s'",
		field, params.Timezone, params.PrevDateFrom.Format("2006-01-02"),
		field, params.Timezone, params.PrevDateTo.Format("2006-01-02"),
	)
}

// IsDailyGranularity returns true if the date range is 60 days or fewer.
// Used to determine whether to aggregate by day or by month.
func IsDailyGranularity(params *QueryParams) bool {
	return params.DayCount <= 60
}

// FormatDate returns a date formatted as YYYY-MM-DD.
func FormatDate(t time.Time) string {
	return t.Format("2006-01-02")
}

// PostDedupCTE returns the common post deduplication CTE used across LinkedIn queries.
// It selects the latest snapshot (max saving_time) for each post_id.
// partitionFilter should be PartitionMonthFilter("created_at", params) to enable partition
// pruning on linkedin_posts (PARTITION BY created_at) alongside the published_at date filter.
func PostDedupCTE(ids, dateFilter, partitionFilter string, extraFilters ...string) string {
	extra := ""
	for _, f := range extraFilters {
		extra += " AND " + f
	}
	return fmt.Sprintf(`WITH posts AS (
    SELECT post_id, max(saving_time) AS saving_time
    FROM linkedin_posts
    WHERE linkedin_id IN %s AND %s AND %s%s
    GROUP BY post_id
)`, ids, dateFilter, partitionFilter, extra)
}

// WithFill returns a ClickHouse WITH FILL clause for time-series gap filling.
func WithFill(endDate string) string {
	return fmt.Sprintf("WITH FILL TO toDate('%s') STEP 1", endDate)
}

// SafeRate returns a SQL expression for a safe division that handles NaN/Inf.
func SafeRate(numerator, denominator string, precision int) string {
	return fmt.Sprintf("ifNotFinite(round(%s / nullIf(%s, 0), %d), 0)", numerator, denominator, precision)
}

// FormatStringList formats a list of strings for use in SQL IN clauses.
func FormatStringList(items []string) string {
	quoted := make([]string, len(items))
	for i, item := range items {
		quoted[i] = "'" + strings.ReplaceAll(item, "'", "\\'") + "'"
	}
	return strings.Join(quoted, ",")
}
