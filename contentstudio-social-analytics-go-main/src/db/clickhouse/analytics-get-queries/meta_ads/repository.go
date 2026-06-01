// Package meta_ads provides ClickHouse repository methods for Meta Ads analytics.
// All queries use FINAL to get deduplicated rows from ReplacingMergeTree tables.
// Partition pruning is applied via toYYYYMM(insights_date) filters.
package meta_ads

import (
	"context"
	"fmt"
	"strings"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// Repository executes ClickHouse queries for Meta Ads analytics.
type Repository struct {
	client *ch.Client
}

const (
	metricSpend       = "spend"
	metricImpressions = "impressions"
	metricReach       = "reach"
	metricClicks      = "clicks"
	metricCPM         = "cpm"
	metricCPC         = "cpc"
	metricCTR         = "ctr"
	metricFrequency   = "frequency"

	objectiveAwareness    = "OUTCOME_AWARENESS"
	objectiveTraffic      = "OUTCOME_TRAFFIC"
	objectiveEngagement   = "OUTCOME_ENGAGEMENT"
	objectiveLeads        = "OUTCOME_LEADS"
	objectiveAppPromotion = "OUTCOME_APP_PROMOTION"
	objectiveSales        = "OUTCOME_SALES"

	legacyObjectiveBrandAwareness = "BRAND_AWARENESS"
	legacyObjectiveReach          = "REACH"
	legacyObjectiveTraffic        = "TRAFFIC"
	legacyObjectiveEngagement     = "ENGAGEMENT"
	legacyObjectiveVideoViews     = "VIDEO_VIEWS"
	legacyObjectiveMessages       = "MESSAGES"
	legacyObjectiveLeadGeneration = "LEAD_GENERATION"
	legacyObjectiveAppInstalls    = "APP_INSTALLS"
	legacyObjectiveConversions    = "CONVERSIONS"
	legacyObjectiveCatalogSales   = "CATALOG_SALES"
	legacyObjectiveStoreVisits    = "STORE_VISITS"
)

var metaAdsObjectives = []string{
	objectiveAwareness,
	objectiveTraffic,
	objectiveEngagement,
	objectiveLeads,
	objectiveAppPromotion,
	objectiveSales,
}

// NewRepository returns a new Repository backed by the given ClickHouse client.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// insightsDateFilter returns a WHERE clause for insights_date with partition pruning.
func insightsDateFilter(params *ch.QueryParams) string {
	return fmt.Sprintf(
		"toYYYYMM(insights_date) >= %s AND toYYYYMM(insights_date) <= %s AND insights_date >= '%s' AND insights_date <= '%s'",
		params.DateFrom.Format("200601"),
		params.DateTo.Format("200601"),
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)
}

// prevInsightsDateFilter returns the same filter for the previous period.
func prevInsightsDateFilter(params *ch.QueryParams) string {
	return fmt.Sprintf(
		"toYYYYMM(insights_date) >= %s AND toYYYYMM(insights_date) <= %s AND insights_date >= '%s' AND insights_date <= '%s'",
		params.PrevDateFrom.Format("200601"),
		params.PrevDateTo.Format("200601"),
		params.PrevDateFrom.Format("2006-01-02"),
		params.PrevDateTo.Format("2006-01-02"),
	)
}

// safeDiv returns a SQL expression for safe division (avoids div-by-zero).
func safeDiv(numerator, denominator string) string {
	return fmt.Sprintf("if(%s > 0, %s / %s, 0)", denominator, numerator, denominator)
}

func sqlString(value string) string {
	return "'" + strings.ReplaceAll(value, "'", "\\'") + "'"
}

func searchHaving(columnExpr, search string) string {
	if search == "" {
		return ""
	}
	return fmt.Sprintf("positionCaseInsensitive(%s, %s) > 0", columnExpr, sqlString(search))
}

// GetAccountCurrency returns the stored currency code for a Meta Ads account.
func (r *Repository) GetAccountCurrency(ctx context.Context, accountID string) (string, error) {
	if accountID == "" {
		return "", fmt.Errorf("GetAccountCurrency: accountID is required")
	}

	query := fmt.Sprintf(`
		SELECT currency
		FROM meta_ads_account_info FINAL
		WHERE account_id = %s
		LIMIT 1
	`, sqlString(accountID))

	var currency string
	if err := r.client.Conn.QueryRow(ctx, query).Scan(&currency); err != nil {
		return "", fmt.Errorf("GetAccountCurrency: %w", err)
	}
	return currency, nil
}

// resultsExpr returns the SQL expression for "results" given a campaign objective.
// It maps both new and legacy objective values to the correct metric column.
func resultsExprForObjective(objective string) string {
	switch objective {
	case objectiveAwareness:
		return "sum(reach)"
	case objectiveTraffic:
		return "sum(clicks)"
	case objectiveEngagement:
		return "sum(actions_post_engagement)"
	case objectiveLeads:
		return "if(sum(actions_lead) > 0, sum(actions_lead), sum(actions_offsite_conversion_fb_pixel_lead))"
	case objectiveAppPromotion:
		return "sum(actions_mobile_app_install)"
	case objectiveSales:
		return "if(sum(actions_purchase) > 0, sum(actions_purchase), sum(actions_offsite_conversion_fb_pixel_purchase))"
	default:
		return "0"
	}
}

// metricExpr returns a SQL column expression for a given metric name.
// Used for performance trend, by-level, and by-platform queries.
func metricExpr(metric string) string {
	switch metric {
	case metricImpressions:
		return "toFloat64(sum(impressions))"
	case metricReach:
		return "toFloat64(sum(reach))"
	case metricClicks:
		return "toFloat64(sum(clicks))"
	case metricCPM:
		return safeDiv("sum(spend) * 1000", "sum(impressions)")
	case metricCPC:
		return safeDiv("sum(spend)", "sum(clicks)")
	case metricCTR:
		return safeDiv("sum(clicks) * 100", "sum(impressions)")
	case metricFrequency:
		return safeDiv("sum(impressions)", "sum(reach)")
	default: // spend
		return "sum(spend)"
	}
}

// ─────────────────────────────────────────────
// Summary
// ─────────────────────────────────────────────

// GetSummary returns aggregated metrics for campaign insights in the given period.
func (r *Repository) GetSummary(ctx context.Context, params *ch.QueryParams) (*SummaryResult, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	query := fmt.Sprintf(`
		SELECT
			sum(spend)       AS spend,
			sum(reach)       AS reach,
			sum(impressions) AS impressions,
			sum(clicks)      AS clicks
		FROM meta_ads_campaign_insights FINAL
		WHERE account_id = %s AND %s`,
		id, dateFilter,
	)

	var result SummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Spend,
		&result.Reach,
		&result.Impressions,
		&result.Clicks,
	)
	if err != nil {
		return nil, fmt.Errorf("GetSummary: %w", err)
	}
	return &result, nil
}

// ─────────────────────────────────────────────
// Results by Objective
// ─────────────────────────────────────────────

// GetResultsByObjective returns results grouped by campaign objective.
// The objectives queried are the 6 standard Meta Ads objectives.
func (r *Repository) GetResultsByObjective(ctx context.Context, params *ch.QueryParams) ([]ObjectiveResultRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	var rows []ObjectiveResultRow
	for _, obj := range metaAdsObjectives {
		resultsExpr := resultsExprForObjective(obj)

		// Legacy objective mapping - we also include legacy names
		var legacyFilter string
		switch obj {
		case objectiveAwareness:
			legacyFilter = fmt.Sprintf(
				"objective IN ('%s','%s','%s')",
				objectiveAwareness,
				legacyObjectiveBrandAwareness,
				legacyObjectiveReach,
			)
		case objectiveTraffic:
			legacyFilter = fmt.Sprintf("objective IN ('%s','%s')", objectiveTraffic, legacyObjectiveTraffic)
		case objectiveEngagement:
			legacyFilter = fmt.Sprintf(
				"objective IN ('%s','%s','%s','%s')",
				objectiveEngagement,
				legacyObjectiveEngagement,
				legacyObjectiveVideoViews,
				legacyObjectiveMessages,
			)
		case objectiveLeads:
			legacyFilter = fmt.Sprintf("objective IN ('%s','%s')", objectiveLeads, legacyObjectiveLeadGeneration)
		case objectiveAppPromotion:
			legacyFilter = fmt.Sprintf("objective IN ('%s','%s')", objectiveAppPromotion, legacyObjectiveAppInstalls)
		case objectiveSales:
			legacyFilter = fmt.Sprintf(
				"objective IN ('%s','%s','%s','%s')",
				objectiveSales,
				legacyObjectiveConversions,
				legacyObjectiveCatalogSales,
				legacyObjectiveStoreVisits,
			)
		}

		query := fmt.Sprintf(`
			SELECT
				'%s'                    AS objective_key,
				%s                      AS results,
				sum(spend)              AS spend,
				uniqExact(campaign_id)  AS campaign_count
			FROM meta_ads_campaign_insights FINAL
			WHERE account_id = %s AND %s AND %s`,
			obj, resultsExpr, id, dateFilter, legacyFilter,
		)

		var row ObjectiveResultRow
		err := r.client.Conn.QueryRow(ctx, query).Scan(
			&row.Objective,
			&row.Results,
			&row.Spend,
			&row.CampaignCount,
		)
		if err != nil {
			return nil, fmt.Errorf("GetResultsByObjective[%s]: %w", obj, err)
		}
		rows = append(rows, row)
	}
	return rows, nil
}

// GetResultsByObjectivePrev returns previous period results (results only) for a single objective.
func (r *Repository) GetResultsByObjectivePrevBatch(ctx context.Context, params *ch.QueryParams) (map[string]int64, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := prevInsightsDateFilter(params)

	objectives := []struct {
		key    string
		legacy string
		expr   string
	}{
		{objectiveAwareness, fmt.Sprintf("objective IN ('%s','%s','%s')", objectiveAwareness, legacyObjectiveBrandAwareness, legacyObjectiveReach), "sum(reach)"},
		{objectiveTraffic, fmt.Sprintf("objective IN ('%s','%s')", objectiveTraffic, legacyObjectiveTraffic), "sum(clicks)"},
		{objectiveEngagement, fmt.Sprintf("objective IN ('%s','%s','%s','%s')", objectiveEngagement, legacyObjectiveEngagement, legacyObjectiveVideoViews, legacyObjectiveMessages), "sum(actions_post_engagement)"},
		{objectiveLeads, fmt.Sprintf("objective IN ('%s','%s')", objectiveLeads, legacyObjectiveLeadGeneration), "if(sum(actions_lead) > 0, sum(actions_lead), sum(actions_offsite_conversion_fb_pixel_lead))"},
		{objectiveAppPromotion, fmt.Sprintf("objective IN ('%s','%s')", objectiveAppPromotion, legacyObjectiveAppInstalls), "sum(actions_mobile_app_install)"},
		{objectiveSales, fmt.Sprintf("objective IN ('%s','%s','%s','%s')", objectiveSales, legacyObjectiveConversions, legacyObjectiveCatalogSales, legacyObjectiveStoreVisits), "if(sum(actions_purchase) > 0, sum(actions_purchase), sum(actions_offsite_conversion_fb_pixel_purchase))"},
	}

	result := make(map[string]int64)
	for _, obj := range objectives {
		query := fmt.Sprintf(`
			SELECT %s AS results
			FROM meta_ads_campaign_insights FINAL
			WHERE account_id = %s AND %s AND %s`,
			obj.expr, id, dateFilter, obj.legacy,
		)
		var val int64
		if err := r.client.Conn.QueryRow(ctx, query).Scan(&val); err != nil {
			return nil, fmt.Errorf("GetResultsByObjectivePrev[%s]: %w", obj.key, err)
		}
		result[obj.key] = val
	}
	return result, nil
}

// ─────────────────────────────────────────────
// Daily metrics chart data
// ─────────────────────────────────────────────

// GetDailyMetrics returns daily impressions, spend, clicks, and CTR for charting.
func (r *Repository) GetDailyMetrics(ctx context.Context, params *ch.QueryParams) ([]DailyMetricsRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	query := fmt.Sprintf(`
		SELECT
			toString(insights_date)              AS dt,
			sum(spend)                           AS spend,
			sum(impressions)                     AS total_impressions,
			sum(clicks)                          AS total_clicks,
			%s                                   AS ctr
		FROM meta_ads_campaign_insights FINAL
		WHERE account_id = %s AND %s
		GROUP BY insights_date
		ORDER BY insights_date ASC
		WITH FILL FROM toDate('%s') TO toDate('%s') STEP 1`,
		safeDiv("sum(clicks) * 100.0", "sum(impressions)"),
		id, dateFilter,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetDailyMetrics: %w", err)
	}
	defer rows.Close()

	var result []DailyMetricsRow
	for rows.Next() {
		var row DailyMetricsRow
		if err := rows.Scan(&row.InsightsDate, &row.Spend, &row.Impressions, &row.Clicks, &row.CTR); err != nil {
			return nil, fmt.Errorf("GetDailyMetrics scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Top 5 Campaigns
// ─────────────────────────────────────────────

// GetTopCampaigns returns the top 5 campaigns sorted by the given metric (spend | impressions | ctr).
func (r *Repository) GetTopCampaigns(ctx context.Context, params *ch.QueryParams, sortBy string) ([]TopCampaignRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	var orderExpr string
	switch sortBy {
	case "impressions":
		orderExpr = "total_impressions"
	case "ctr":
		orderExpr = "ctr"
	default:
		orderExpr = "spend"
	}

	query := fmt.Sprintf(`
		SELECT
			campaign_id,
			any(campaign_name) AS campaign_name,
			sum(spend)         AS spend,
			sum(impressions)   AS total_impressions,
			%s                 AS ctr
		FROM meta_ads_campaign_insights FINAL
		WHERE account_id = %s AND %s
		GROUP BY campaign_id
		ORDER BY %s DESC
		LIMIT 5`,
		safeDiv("sum(clicks) * 100.0", "sum(impressions)"),
		id, dateFilter,
		orderExpr,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetTopCampaigns: %w", err)
	}
	defer rows.Close()

	var result []TopCampaignRow
	for rows.Next() {
		var row TopCampaignRow
		if err := rows.Scan(&row.CampaignID, &row.CampaignName, &row.Spend, &row.Impressions, &row.CTR); err != nil {
			return nil, fmt.Errorf("GetTopCampaigns scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Performance Trend
// ─────────────────────────────────────────────

// GetPerformanceTrend returns a daily time series for the selected metric.
func (r *Repository) GetPerformanceTrend(ctx context.Context, params *ch.QueryParams, metric string) ([]TrendRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)
	valueExpr := metricExpr(metric)

	query := fmt.Sprintf(`
		SELECT
			toString(insights_date) AS dt,
			%s                      AS value
		FROM meta_ads_campaign_insights FINAL
		WHERE account_id = %s AND %s
		GROUP BY insights_date
		ORDER BY insights_date ASC
		WITH FILL FROM toDate('%s') TO toDate('%s') STEP 1`,
		valueExpr, id, dateFilter,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetPerformanceTrend: %w", err)
	}
	defer rows.Close()

	var result []TrendRow
	for rows.Next() {
		var row TrendRow
		if err := rows.Scan(&row.InsightsDate, &row.Value); err != nil {
			return nil, fmt.Errorf("GetPerformanceTrend scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Performance by Campaign
// ─────────────────────────────────────────────

// GetPerformanceByCampaign returns top campaigns by selected metric (max 20 rows).
func (r *Repository) GetPerformanceByCampaign(ctx context.Context, params *ch.QueryParams, metric string) ([]LevelBreakdownRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)
	valueExpr := metricExpr(metric)

	query := fmt.Sprintf(`
		SELECT
			campaign_id        AS id,
			any(campaign_name) AS name,
			%s                 AS value
		FROM meta_ads_campaign_insights FINAL
		WHERE account_id = %s AND %s
		GROUP BY campaign_id
		ORDER BY value DESC
		LIMIT 20`,
		valueExpr, id, dateFilter,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetPerformanceByCampaign: %w", err)
	}
	defer rows.Close()

	var result []LevelBreakdownRow
	for rows.Next() {
		var row LevelBreakdownRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Value); err != nil {
			return nil, fmt.Errorf("GetPerformanceByCampaign scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Performance by Ad Set
// ─────────────────────────────────────────────

// GetPerformanceByAdSet returns top ad sets by selected metric (max 20 rows).
func (r *Repository) GetPerformanceByAdSet(ctx context.Context, params *ch.QueryParams, metric string) ([]LevelBreakdownRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)
	valueExpr := metricExpr(metric)

	query := fmt.Sprintf(`
		SELECT
			adset_id        AS id,
			any(adset_name) AS name,
			%s              AS value
		FROM meta_ads_adset_insights FINAL
		WHERE account_id = %s AND %s
		GROUP BY adset_id
		ORDER BY value DESC
		LIMIT 20`,
		valueExpr, id, dateFilter,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetPerformanceByAdSet: %w", err)
	}
	defer rows.Close()

	var result []LevelBreakdownRow
	for rows.Next() {
		var row LevelBreakdownRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Value); err != nil {
			return nil, fmt.Errorf("GetPerformanceByAdSet scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Performance by Ad
// ─────────────────────────────────────────────

// GetPerformanceByAd returns top ads by selected metric (max 20 rows).
func (r *Repository) GetPerformanceByAd(ctx context.Context, params *ch.QueryParams, metric string) ([]LevelBreakdownRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)
	valueExpr := metricExpr(metric)

	query := fmt.Sprintf(`
		SELECT
			ad_id        AS id,
			any(ad_name) AS name,
			%s           AS value
		FROM meta_ads_ad_insights FINAL
		WHERE account_id = %s AND %s
		GROUP BY ad_id
		ORDER BY value DESC
		LIMIT 20`,
		valueExpr, id, dateFilter,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetPerformanceByAd: %w", err)
	}
	defer rows.Close()

	var result []LevelBreakdownRow
	for rows.Next() {
		var row LevelBreakdownRow
		if err := rows.Scan(&row.ID, &row.Name, &row.Value); err != nil {
			return nil, fmt.Errorf("GetPerformanceByAd scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Performance by Platform
// ─────────────────────────────────────────────

// GetPerformanceByPlatform returns metric grouped by publisher_platform.
func (r *Repository) GetPerformanceByPlatform(ctx context.Context, params *ch.QueryParams, metric string) ([]PlatformBreakdownRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)
	valueExpr := metricExpr(metric)

	query := fmt.Sprintf(`
		SELECT
			publisher_platform AS platform,
			%s                 AS value
		FROM meta_ads_demographics_device_platform FINAL
		WHERE account_id = %s AND %s
		GROUP BY publisher_platform
		ORDER BY value DESC`,
		valueExpr, id, dateFilter,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetPerformanceByPlatform: %w", err)
	}
	defer rows.Close()

	var result []PlatformBreakdownRow
	for rows.Next() {
		var row PlatformBreakdownRow
		if err := rows.Scan(&row.Platform, &row.Value); err != nil {
			return nil, fmt.Errorf("GetPerformanceByPlatform scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Campaigns Table
// ─────────────────────────────────────────────

// resultsExprByObjective returns a SQL CASE expression that computes "results"
// based on the objective value of each row.
const resultsExprByCampaignObjective = `multiIf(
	any(ci.objective) IN ('OUTCOME_AWARENESS','BRAND_AWARENESS','REACH'), sum(ci.reach),
	any(ci.objective) IN ('OUTCOME_TRAFFIC','TRAFFIC'), sum(ci.clicks),
	any(ci.objective) IN ('OUTCOME_ENGAGEMENT','ENGAGEMENT','VIDEO_VIEWS','MESSAGES'), sum(ci.actions_post_engagement),
	any(ci.objective) IN ('OUTCOME_LEADS','LEAD_GENERATION'), if(sum(ci.actions_lead) > 0, sum(ci.actions_lead), sum(ci.actions_offsite_conversion_fb_pixel_lead)),
	any(ci.objective) IN ('OUTCOME_APP_PROMOTION','APP_INSTALLS'), sum(ci.actions_mobile_app_install),
	any(ci.objective) IN ('OUTCOME_SALES','CONVERSIONS','CATALOG_SALES','STORE_VISITS'), if(sum(ci.actions_purchase) > 0, sum(ci.actions_purchase), sum(ci.actions_offsite_conversion_fb_pixel_purchase)),
	0
)`

// GetCampaignsList returns paginated campaign rows.
// Status comes from the meta_ads_campaigns FINAL list table via a LEFT JOIN.
func (r *Repository) GetCampaignsList(ctx context.Context, params *ch.QueryParams, status, objective, search, orderBy, orderDir string, page, perPage int) ([]CampaignTableRow, int64, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	// Build having filters (applied after GROUP BY)
	var havingClauses []string
	if status != "" {
		havingClauses = append(havingClauses, fmt.Sprintf("any(c.status) = '%s'", strings.ReplaceAll(status, "'", "\\'")))
	}
	if objective != "" {
		havingClauses = append(havingClauses, "any(ci.objective) IN "+buildObjectiveInList(objective))
	}
	if clause := searchHaving("any(ci.campaign_name)", search); clause != "" {
		havingClauses = append(havingClauses, clause)
	}
	havingClause := ""
	if len(havingClauses) > 0 {
		havingClause = "HAVING " + strings.Join(havingClauses, " AND ")
	}

	validOrderBy := map[string]bool{"spend": true, "impressions": true, "reach": true, "clicks": true, "cpm": true, "cpc": true, "ctr": true, "frequency": true}
	if !validOrderBy[orderBy] {
		orderBy = "spend"
	}
	orderExpr := map[string]string{
		"spend":       "total_spend",
		"impressions": "total_impressions",
		"reach":       "total_reach",
		"clicks":      "total_clicks",
		"cpm":         "avg_cpm",
		"cpc":         "avg_cpc",
		"ctr":         "avg_ctr",
		"frequency":   "avg_frequency",
	}[orderBy]
	if orderDir != "asc" {
		orderDir = "desc"
	}

	offset := (page - 1) * perPage

	// Count distinct campaigns matching filters
	countQuery := fmt.Sprintf(`
		SELECT uniqExact(ci.campaign_id) AS total
		FROM meta_ads_campaign_insights AS ci FINAL
		LEFT ANY JOIN (
			SELECT campaign_id, argMax(status, inserted_at) AS status
			FROM meta_ads_campaigns FINAL
			WHERE account_id = %s
			GROUP BY campaign_id
		) AS c ON ci.campaign_id = c.campaign_id
		WHERE ci.account_id = %s AND %s`,
		id, id, dateFilter)
	if len(havingClauses) > 0 {
		// For count we approximate by running the full grouped query
		countQuery = fmt.Sprintf(`
			SELECT count() FROM (
				SELECT ci.campaign_id
				FROM meta_ads_campaign_insights AS ci FINAL
				LEFT ANY JOIN (
					SELECT campaign_id, argMax(status, inserted_at) AS status
					FROM meta_ads_campaigns FINAL
					WHERE account_id = %s
					GROUP BY campaign_id
				) AS c ON ci.campaign_id = c.campaign_id
				WHERE ci.account_id = %s AND %s
				GROUP BY ci.campaign_id
				%s
			)`, id, id, dateFilter, havingClause)
	}

	var total uint64
	if err := r.client.Conn.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetCampaignsList count: %w", err)
	}

	query := fmt.Sprintf(`
		SELECT
			ci.campaign_id,
			any(ci.campaign_name)  AS campaign_name,
			any(c.status)          AS status,
			any(ci.objective)      AS campaign_objective,
			%s                     AS results,
			sum(ci.spend)          AS total_spend,
			sum(ci.reach)          AS total_reach,
			sum(ci.impressions)    AS total_impressions,
			%s                     AS avg_frequency,
			sum(ci.clicks)         AS total_clicks,
			%s                     AS avg_cpm,
			%s                     AS avg_cpc,
			%s                     AS avg_ctr
		FROM meta_ads_campaign_insights AS ci FINAL
		LEFT ANY JOIN (
			SELECT campaign_id, argMax(status, inserted_at) AS status
			FROM meta_ads_campaigns FINAL
			WHERE account_id = %s
			GROUP BY campaign_id
		) AS c ON ci.campaign_id = c.campaign_id
		WHERE ci.account_id = %s AND %s
		GROUP BY ci.campaign_id
		%s
		ORDER BY %s %s
		LIMIT %d OFFSET %d`,
		resultsExprByCampaignObjective,
		safeDiv("sum(ci.impressions)", "sum(ci.reach)"),
		safeDiv("sum(ci.spend) * 1000", "sum(ci.impressions)"),
		safeDiv("sum(ci.spend)", "sum(ci.clicks)"),
		safeDiv("sum(ci.clicks) * 100.0", "sum(ci.impressions)"),
		id,
		id, dateFilter,
		havingClause,
		orderExpr, orderDir,
		perPage, offset,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("GetCampaignsList: %w", err)
	}
	defer rows.Close()

	var result []CampaignTableRow
	for rows.Next() {
		var row CampaignTableRow
		if err := rows.Scan(
			&row.CampaignID, &row.CampaignName, &row.Status, &row.Objective,
			&row.Results, &row.Spend, &row.Reach, &row.Impressions,
			&row.Frequency, &row.Clicks, &row.CPM, &row.CPC, &row.CTR,
		); err != nil {
			return nil, 0, fmt.Errorf("GetCampaignsList scan: %w", err)
		}
		result = append(result, row)
	}
	return result, int64(total), rows.Err()
}

// GetCampaignFilterOptions returns distinct status and objective values for filter dropdowns.
func (r *Repository) GetCampaignFilterOptions(ctx context.Context, params *ch.QueryParams) ([]string, []string, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	statusQuery := fmt.Sprintf(`
		SELECT DISTINCT status
		FROM (
			SELECT argMax(status, inserted_at) AS status
			FROM meta_ads_campaigns FINAL
			WHERE account_id = %s
			GROUP BY campaign_id
		)
		WHERE status != ''
		ORDER BY status`, id)

	statusRows, err := r.client.Conn.Query(ctx, statusQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("GetCampaignFilterOptions status: %w", err)
	}
	defer statusRows.Close()
	statusSet := make(map[string]struct{})
	var statuses []string
	for statusRows.Next() {
		var s string
		if err := statusRows.Scan(&s); err != nil {
			return nil, nil, err
		}
		if _, seen := statusSet[s]; !seen {
			statusSet[s] = struct{}{}
			statuses = append(statuses, s)
		}
	}

	objQuery := fmt.Sprintf(`
		SELECT DISTINCT objective FROM meta_ads_campaign_insights FINAL
		WHERE account_id = %s AND %s AND objective != ''
		ORDER BY objective`, id, dateFilter)

	objRows, err := r.client.Conn.Query(ctx, objQuery)
	if err != nil {
		return nil, nil, fmt.Errorf("GetCampaignFilterOptions objectives: %w", err)
	}
	defer objRows.Close()
	var objectives []string
	for objRows.Next() {
		var o string
		if err := objRows.Scan(&o); err != nil {
			return nil, nil, err
		}
		objectives = append(objectives, o)
	}

	return statuses, objectives, nil
}

// ─────────────────────────────────────────────
// Ad Sets Table
// ─────────────────────────────────────────────

// GetAdSetsList returns paginated ad set rows.
// Status comes from meta_ads_adsets FINAL list table via a LEFT JOIN.
func (r *Repository) GetAdSetsList(ctx context.Context, params *ch.QueryParams, status, objective, search, orderBy, orderDir string, page, perPage int) ([]AdSetTableRow, int64, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	var havingClauses []string
	if status != "" {
		havingClauses = append(havingClauses, fmt.Sprintf("any(a.status) = '%s'", strings.ReplaceAll(status, "'", "\\'")))
	}
	if objective != "" {
		havingClauses = append(havingClauses, "any(ci.objective) IN "+buildObjectiveInList(objective))
	}
	if clause := searchHaving("any(ai.adset_name)", search); clause != "" {
		havingClauses = append(havingClauses, clause)
	}
	havingClause := ""
	if len(havingClauses) > 0 {
		havingClause = "HAVING " + strings.Join(havingClauses, " AND ")
	}

	validOrderBy := map[string]bool{"spend": true, "impressions": true, "reach": true, "clicks": true, "cpm": true, "cpc": true, "ctr": true, "frequency": true}
	if !validOrderBy[orderBy] {
		orderBy = "spend"
	}
	orderExpr := map[string]string{
		"spend":       "total_spend",
		"impressions": "total_impressions",
		"reach":       "total_reach",
		"clicks":      "total_clicks",
		"cpm":         "avg_cpm",
		"cpc":         "avg_cpc",
		"ctr":         "avg_ctr",
		"frequency":   "avg_frequency",
	}[orderBy]
	if orderDir != "asc" {
		orderDir = "desc"
	}

	offset := (page - 1) * perPage

	countQuery := fmt.Sprintf(`
		SELECT count() FROM (
			SELECT ai.adset_id
			FROM meta_ads_adset_insights AS ai FINAL
			LEFT ANY JOIN (
				SELECT adset_id, argMax(status, inserted_at) AS status
				FROM meta_ads_adsets FINAL WHERE account_id = %s GROUP BY adset_id
			) AS a ON ai.adset_id = a.adset_id
			LEFT ANY JOIN (
				SELECT campaign_id, argMax(objective, inserted_at) AS objective
				FROM meta_ads_campaigns FINAL WHERE account_id = %s GROUP BY campaign_id
			) AS ci ON ai.campaign_id = ci.campaign_id
			WHERE ai.account_id = %s AND %s
			GROUP BY ai.adset_id
			%s
		)`, id, id, id, dateFilter, havingClause)

	var total uint64
	if err := r.client.Conn.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetAdSetsList count: %w", err)
	}

	// Results expression uses the campaign objective mapped to adset_insights columns
	adsetResultsExpr := `multiIf(
		any(ci.objective) IN ('OUTCOME_AWARENESS','BRAND_AWARENESS','REACH'), sum(ai.reach),
		any(ci.objective) IN ('OUTCOME_TRAFFIC','TRAFFIC'), sum(ai.clicks),
		any(ci.objective) IN ('OUTCOME_ENGAGEMENT','ENGAGEMENT','VIDEO_VIEWS','MESSAGES'), sum(ai.actions_post_engagement),
		any(ci.objective) IN ('OUTCOME_LEADS','LEAD_GENERATION'), if(sum(ai.actions_lead) > 0, sum(ai.actions_lead), sum(ai.actions_offsite_conversion_fb_pixel_lead)),
		any(ci.objective) IN ('OUTCOME_APP_PROMOTION','APP_INSTALLS'), sum(ai.actions_mobile_app_install),
		any(ci.objective) IN ('OUTCOME_SALES','CONVERSIONS','CATALOG_SALES','STORE_VISITS'), if(sum(ai.actions_purchase) > 0, sum(ai.actions_purchase), sum(ai.actions_offsite_conversion_fb_pixel_purchase)),
		0
	)`

	query := fmt.Sprintf(`
		SELECT
			ai.adset_id,
			any(ai.adset_name)    AS adset_name,
			any(ai.campaign_id)   AS campaign_id,
			any(ai.campaign_name) AS campaign_name,
			any(a.status)         AS status,
			any(ci.objective)     AS campaign_objective,
			%s                    AS results,
			sum(ai.spend)         AS total_spend,
			sum(ai.reach)         AS total_reach,
			sum(ai.impressions)   AS total_impressions,
			%s                    AS avg_frequency,
			sum(ai.clicks)        AS total_clicks,
			%s                    AS avg_cpm,
			%s                    AS avg_cpc,
			%s                    AS avg_ctr
		FROM meta_ads_adset_insights AS ai FINAL
		LEFT ANY JOIN (
			SELECT adset_id, argMax(status, inserted_at) AS status
			FROM meta_ads_adsets FINAL WHERE account_id = %s GROUP BY adset_id
		) AS a ON ai.adset_id = a.adset_id
		LEFT ANY JOIN (
			SELECT campaign_id, argMax(objective, inserted_at) AS objective
			FROM meta_ads_campaigns FINAL WHERE account_id = %s GROUP BY campaign_id
		) AS ci ON ai.campaign_id = ci.campaign_id
		WHERE ai.account_id = %s AND %s
		GROUP BY ai.adset_id
		%s
		ORDER BY %s %s
		LIMIT %d OFFSET %d`,
		adsetResultsExpr,
		safeDiv("sum(ai.impressions)", "sum(ai.reach)"),
		safeDiv("sum(ai.spend) * 1000", "sum(ai.impressions)"),
		safeDiv("sum(ai.spend)", "sum(ai.clicks)"),
		safeDiv("sum(ai.clicks) * 100.0", "sum(ai.impressions)"),
		id, id, id, dateFilter,
		havingClause,
		orderExpr, orderDir,
		perPage, offset,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("GetAdSetsList: %w", err)
	}
	defer rows.Close()

	var result []AdSetTableRow
	for rows.Next() {
		var row AdSetTableRow
		if err := rows.Scan(
			&row.AdSetID, &row.AdSetName, &row.CampaignID, &row.CampaignName,
			&row.Status, &row.Objective,
			&row.Results, &row.Spend, &row.Reach, &row.Impressions,
			&row.Frequency, &row.Clicks, &row.CPM, &row.CPC, &row.CTR,
		); err != nil {
			return nil, 0, fmt.Errorf("GetAdSetsList scan: %w", err)
		}
		result = append(result, row)
	}
	return result, int64(total), rows.Err()
}

// GetAdSetFilterOptions returns distinct statuses for the ad set filter dropdown.
func (r *Repository) GetAdSetFilterOptions(ctx context.Context, params *ch.QueryParams) ([]string, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"

	query := fmt.Sprintf(`
		SELECT DISTINCT status
		FROM (
			SELECT argMax(status, inserted_at) AS status
			FROM meta_ads_adsets FINAL
			WHERE account_id = %s
			GROUP BY adset_id
		)
		WHERE status != ''
		ORDER BY status`, id)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetAdSetFilterOptions: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	var result []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Ads Table
// ─────────────────────────────────────────────

// GetAdsList returns paginated ad rows.
// Status comes from meta_ads_ads FINAL list table via a LEFT JOIN.
func (r *Repository) GetAdsList(ctx context.Context, params *ch.QueryParams, status, objective, search, orderBy, orderDir string, page, perPage int) ([]AdTableRow, int64, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	var havingClauses []string
	if status != "" {
		havingClauses = append(havingClauses, fmt.Sprintf("any(ads.status) = '%s'", strings.ReplaceAll(status, "'", "\\'")))
	}
	if objective != "" {
		havingClauses = append(havingClauses, "any(ci.objective) IN "+buildObjectiveInList(objective))
	}
	if clause := searchHaving("any(adi.ad_name)", search); clause != "" {
		havingClauses = append(havingClauses, clause)
	}
	havingClause := ""
	if len(havingClauses) > 0 {
		havingClause = "HAVING " + strings.Join(havingClauses, " AND ")
	}

	validOrderBy := map[string]bool{"spend": true, "impressions": true, "reach": true, "clicks": true, "cpm": true, "cpc": true, "ctr": true, "frequency": true}
	if !validOrderBy[orderBy] {
		orderBy = "spend"
	}
	orderExpr := map[string]string{
		"spend":       "total_spend",
		"impressions": "total_impressions",
		"reach":       "total_reach",
		"clicks":      "total_clicks",
		"cpm":         "avg_cpm",
		"cpc":         "avg_cpc",
		"ctr":         "avg_ctr",
		"frequency":   "avg_frequency",
	}[orderBy]
	if orderDir != "asc" {
		orderDir = "desc"
	}

	offset := (page - 1) * perPage

	countQuery := fmt.Sprintf(`
		SELECT count() FROM (
			SELECT adi.ad_id
			FROM meta_ads_ad_insights AS adi FINAL
			LEFT ANY JOIN (
				SELECT ad_id, argMax(status, inserted_at) AS status, argMax(adset_id, inserted_at) AS adset_id
				FROM meta_ads_ads FINAL WHERE account_id = %s GROUP BY ad_id
			) AS ads ON adi.ad_id = ads.ad_id
			LEFT ANY JOIN (
				SELECT campaign_id, argMax(objective, inserted_at) AS objective
				FROM meta_ads_campaigns FINAL WHERE account_id = %s GROUP BY campaign_id
			) AS ci ON adi.campaign_id = ci.campaign_id
			WHERE adi.account_id = %s AND %s
			GROUP BY adi.ad_id
			%s
		)`, id, id, id, dateFilter, havingClause)

	var total uint64
	if err := r.client.Conn.QueryRow(ctx, countQuery).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("GetAdsList count: %w", err)
	}

	adResultsExpr := `multiIf(
		any(ci.objective) IN ('OUTCOME_AWARENESS','BRAND_AWARENESS','REACH'), sum(adi.reach),
		any(ci.objective) IN ('OUTCOME_TRAFFIC','TRAFFIC'), sum(adi.clicks),
		any(ci.objective) IN ('OUTCOME_ENGAGEMENT','ENGAGEMENT','VIDEO_VIEWS','MESSAGES'), sum(adi.actions_post_engagement),
		any(ci.objective) IN ('OUTCOME_LEADS','LEAD_GENERATION'), if(sum(adi.actions_lead) > 0, sum(adi.actions_lead), sum(adi.actions_offsite_conversion_fb_pixel_lead)),
		any(ci.objective) IN ('OUTCOME_APP_PROMOTION','APP_INSTALLS'), sum(adi.actions_mobile_app_install),
		any(ci.objective) IN ('OUTCOME_SALES','CONVERSIONS','CATALOG_SALES','STORE_VISITS'), if(sum(adi.actions_purchase) > 0, sum(adi.actions_purchase), sum(adi.actions_offsite_conversion_fb_pixel_purchase)),
		0
	)`

	query := fmt.Sprintf(`
		SELECT
			adi.ad_id,
			any(adi.ad_name)     AS ad_name,
			any(ads.adset_id)    AS adset_id,
			any(ads.adset_name)  AS adset_name,
			any(adi.campaign_id) AS campaign_id,
			any(adi.campaign_name) AS campaign_name,
			any(ads.status)      AS status,
			any(ci.objective)    AS campaign_objective,
			any(ads.creative_name) AS creative_name,
			any(ads.creative_title) AS creative_title,
			any(ads.creative_body) AS creative_body,
			any(ads.creative_thumbnail_url) AS creative_thumbnail_url,
			any(ads.creative_effective_object_story_id) AS creative_effective_object_story_id,
			%s                   AS results,
			sum(adi.spend)       AS total_spend,
			sum(adi.reach)       AS total_reach,
			sum(adi.impressions) AS total_impressions,
			%s                   AS avg_frequency,
			sum(adi.clicks)      AS total_clicks,
			%s                   AS avg_cpm,
			%s                   AS avg_cpc,
			%s                   AS avg_ctr
		FROM meta_ads_ad_insights AS adi FINAL
		LEFT ANY JOIN (
			SELECT ad_id, argMax(status, inserted_at) AS status, argMax(adset_id, inserted_at) AS adset_id, argMax(adset_name, inserted_at) AS adset_name, argMax(creative_name, inserted_at) AS creative_name, argMax(creative_title, inserted_at) AS creative_title, argMax(creative_body, inserted_at) AS creative_body, argMax(creative_thumbnail_url, inserted_at) AS creative_thumbnail_url, argMax(creative_effective_object_story_id, inserted_at) AS creative_effective_object_story_id
			FROM meta_ads_ads FINAL WHERE account_id = %s GROUP BY ad_id
		) AS ads ON adi.ad_id = ads.ad_id
		LEFT ANY JOIN (
			SELECT campaign_id, argMax(objective, inserted_at) AS objective
			FROM meta_ads_campaigns FINAL WHERE account_id = %s GROUP BY campaign_id
		) AS ci ON adi.campaign_id = ci.campaign_id
		WHERE adi.account_id = %s AND %s
		GROUP BY adi.ad_id
		%s
		ORDER BY %s %s
		LIMIT %d OFFSET %d`,
		adResultsExpr,
		safeDiv("sum(adi.impressions)", "sum(adi.reach)"),
		safeDiv("sum(adi.spend) * 1000", "sum(adi.impressions)"),
		safeDiv("sum(adi.spend)", "sum(adi.clicks)"),
		safeDiv("sum(adi.clicks) * 100.0", "sum(adi.impressions)"),
		id, id, id, dateFilter,
		havingClause,
		orderExpr, orderDir,
		perPage, offset,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, 0, fmt.Errorf("GetAdsList: %w", err)
	}
	defer rows.Close()

	var result []AdTableRow
	for rows.Next() {
		var row AdTableRow
		if err := rows.Scan(
			&row.AdID, &row.AdName, &row.AdSetID, &row.AdSetName, &row.CampaignID, &row.CampaignName,
			&row.Status, &row.Objective, &row.CreativeName, &row.CreativeTitle, &row.CreativeBody, &row.CreativeThumbnailURL, &row.CreativeEffectiveObjectStoryID,
			&row.Results, &row.Spend, &row.Reach, &row.Impressions,
			&row.Frequency, &row.Clicks, &row.CPM, &row.CPC, &row.CTR,
		); err != nil {
			return nil, 0, fmt.Errorf("GetAdsList scan: %w", err)
		}
		result = append(result, row)
	}
	return result, int64(total), rows.Err()
}

// GetAdFilterOptions returns distinct statuses for the ads filter dropdown.
func (r *Repository) GetAdFilterOptions(ctx context.Context, params *ch.QueryParams) ([]string, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"

	query := fmt.Sprintf(`
		SELECT DISTINCT status
		FROM (
			SELECT argMax(status, inserted_at) AS status
			FROM meta_ads_ads FINAL
			WHERE account_id = %s
			GROUP BY ad_id
		)
		WHERE status != ''
		ORDER BY status`, id)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetAdFilterOptions: %w", err)
	}
	defer rows.Close()

	seen := make(map[string]struct{})
	var result []string
	for rows.Next() {
		var s string
		if err := rows.Scan(&s); err != nil {
			return nil, err
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			result = append(result, s)
		}
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Demographics: Age & Gender
// ─────────────────────────────────────────────

// GetDemographicsAgeGender returns age/gender breakdown rows.
func (r *Repository) GetDemographicsAgeGender(ctx context.Context, params *ch.QueryParams) ([]AgeGenderRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	query := fmt.Sprintf(`
		SELECT
			age,
			gender,
			sum(spend)       AS total_spend,
			sum(impressions) AS total_impressions,
			sum(reach)       AS total_reach,
			sum(clicks)      AS total_clicks,
			%s               AS avg_ctr,
			%s               AS avg_cpm,
			%s               AS avg_cpc,
			%s               AS avg_frequency
		FROM meta_ads_demographics_age_gender FINAL
		WHERE account_id = %s AND %s
		GROUP BY age, gender
		ORDER BY age ASC, gender ASC`,
		safeDiv("sum(clicks) * 100.0", "sum(impressions)"),
		safeDiv("sum(spend) * 1000", "sum(impressions)"),
		safeDiv("sum(spend)", "sum(clicks)"),
		safeDiv("sum(impressions)", "sum(reach)"),
		id, dateFilter,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetDemographicsAgeGender: %w", err)
	}
	defer rows.Close()

	var result []AgeGenderRow
	for rows.Next() {
		var row AgeGenderRow
		if err := rows.Scan(
			&row.Age, &row.Gender,
			&row.Spend, &row.Impressions, &row.Reach, &row.Clicks,
			&row.CTR, &row.CPM, &row.CPC, &row.Frequency,
		); err != nil {
			return nil, fmt.Errorf("GetDemographicsAgeGender scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Demographics: Region / Country
// ─────────────────────────────────────────────

// GetDemographicsRegionCountry returns region or country breakdown.
// If breakdown = "country", region is omitted and grouped only by country.
// If breakdown = "region", an optional countryFilter scopes the results.
func (r *Repository) GetDemographicsRegionCountry(ctx context.Context, params *ch.QueryParams, breakdown, countryFilter, orderBy, orderDir string) ([]RegionCountryRow, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	validOrderBy := map[string]bool{"spend": true, "impressions": true, "clicks": true, "ctr": true}
	if !validOrderBy[orderBy] {
		orderBy = "spend"
	}
	orderExpr := map[string]string{
		"spend":       "total_spend",
		"impressions": "total_impressions",
		"clicks":      "total_clicks",
		"ctr":         "avg_ctr",
	}[orderBy]
	if orderDir != "asc" {
		orderDir = "desc"
	}

	var selectFields, groupByFields string
	var extraFilter string

	if breakdown == "country" {
		selectFields = "country, '' AS region"
		groupByFields = "country"
	} else {
		selectFields = "country, region"
		groupByFields = "country, region"
		if countryFilter != "" {
			extraFilter = fmt.Sprintf(" AND country = '%s'", strings.ReplaceAll(countryFilter, "'", "\\'"))
		}
	}

	query := fmt.Sprintf(`
		SELECT
			%s,
			sum(spend)       AS total_spend,
			sum(impressions) AS total_impressions,
			sum(clicks)      AS total_clicks,
			%s               AS avg_ctr
		FROM meta_ads_demographics_region_country FINAL
		WHERE account_id = %s AND %s%s
		GROUP BY %s
		ORDER BY %s %s
		LIMIT 20`,
		selectFields,
		safeDiv("sum(clicks) * 100.0", "sum(impressions)"),
		id, dateFilter, extraFilter,
		groupByFields,
		orderExpr, orderDir,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetDemographicsRegionCountry: %w", err)
	}
	defer rows.Close()

	var result []RegionCountryRow
	for rows.Next() {
		var row RegionCountryRow
		if err := rows.Scan(&row.Country, &row.Region, &row.Spend, &row.Impressions, &row.Clicks, &row.CTR); err != nil {
			return nil, fmt.Errorf("GetDemographicsRegionCountry scan: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

// GetAvailableCountries returns distinct countries for the region filter dropdown.
func (r *Repository) GetAvailableCountries(ctx context.Context, params *ch.QueryParams) ([]string, error) {
	id := "'" + strings.ReplaceAll(params.AccountIDs[0], "'", "\\'") + "'"
	dateFilter := insightsDateFilter(params)

	query := fmt.Sprintf(`
		SELECT DISTINCT country FROM meta_ads_demographics_region_country FINAL
		WHERE account_id = %s AND %s AND country != ''
		ORDER BY country`, id, dateFilter)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetAvailableCountries: %w", err)
	}
	defer rows.Close()

	var result []string
	for rows.Next() {
		var c string
		if err := rows.Scan(&c); err != nil {
			return nil, err
		}
		result = append(result, c)
	}
	return result, rows.Err()
}

// ─────────────────────────────────────────────
// Helper: Objective filter clause
// ─────────────────────────────────────────────

func buildObjectiveFilter(objective string) string {
	switch objective {
	case objectiveAwareness:
		return fmt.Sprintf("objective IN ('%s','%s','%s')", objectiveAwareness, legacyObjectiveBrandAwareness, legacyObjectiveReach)
	case objectiveTraffic:
		return fmt.Sprintf("objective IN ('%s','%s')", objectiveTraffic, legacyObjectiveTraffic)
	case objectiveEngagement:
		return fmt.Sprintf("objective IN ('%s','%s','%s','%s')", objectiveEngagement, legacyObjectiveEngagement, legacyObjectiveVideoViews, legacyObjectiveMessages)
	case objectiveLeads:
		return fmt.Sprintf("objective IN ('%s','%s')", objectiveLeads, legacyObjectiveLeadGeneration)
	case objectiveAppPromotion:
		return fmt.Sprintf("objective IN ('%s','%s')", objectiveAppPromotion, legacyObjectiveAppInstalls)
	case objectiveSales:
		return fmt.Sprintf("objective IN ('%s','%s','%s','%s')", objectiveSales, legacyObjectiveConversions, legacyObjectiveCatalogSales, legacyObjectiveStoreVisits)
	default:
		return fmt.Sprintf("objective = '%s'", strings.ReplaceAll(objective, "'", "\\'"))
	}
}

// buildObjectiveInList returns the IN (...) list portion (without the column name) for use in HAVING clauses.
func buildObjectiveInList(objective string) string {
	switch objective {
	case objectiveAwareness:
		return fmt.Sprintf("('%s','%s','%s')", objectiveAwareness, legacyObjectiveBrandAwareness, legacyObjectiveReach)
	case objectiveTraffic:
		return fmt.Sprintf("('%s','%s')", objectiveTraffic, legacyObjectiveTraffic)
	case objectiveEngagement:
		return fmt.Sprintf("('%s','%s','%s','%s')", objectiveEngagement, legacyObjectiveEngagement, legacyObjectiveVideoViews, legacyObjectiveMessages)
	case objectiveLeads:
		return fmt.Sprintf("('%s','%s')", objectiveLeads, legacyObjectiveLeadGeneration)
	case objectiveAppPromotion:
		return fmt.Sprintf("('%s','%s')", objectiveAppPromotion, legacyObjectiveAppInstalls)
	case objectiveSales:
		return fmt.Sprintf("('%s','%s','%s','%s')", objectiveSales, legacyObjectiveConversions, legacyObjectiveCatalogSales, legacyObjectiveStoreVisits)
	default:
		return "('" + strings.ReplaceAll(objective, "'", "\\'") + "')"
	}
}
