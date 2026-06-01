package pinterest

import (
	"context"
	"fmt"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// Repository executes ClickHouse queries for Pinterest analytics.
type Repository struct {
	client *ch.Client
}

// NewRepository returns a new Repository backed by the given ClickHouse client.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// ptDateFilter returns a WHERE clause fragment using toDate() comparisons.
// Pinterest tables store date fields without timezone, so toDate() is used directly.
func ptDateFilter(field string, params *ch.QueryParams) string {
	return fmt.Sprintf(
		"toDate(%s) BETWEEN toDate('%s') AND toDate('%s')",
		field,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)
}

// ptDateFill returns a WITH FILL clause for time-series gap filling.
// When daily is true, fills by day; otherwise fills by calendar month.
func ptDateFill(params *ch.QueryParams, monthly bool) string {
	if monthly {
		return fmt.Sprintf(
			"WITH FILL FROM toStartOfMonth(toDate('%s')) TO toStartOfMonth(toDate('%s')) STEP toIntervalMonth(1)",
			params.DateFrom.Format("2006-01-02"),
			params.DateTo.AddDate(0, 1, 0).Format("2006-01-02"),
		)
	}
	return fmt.Sprintf(
		"WITH FILL FROM toDate('%s') TO toDate('%s') STEP 1",
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.AddDate(0, 0, 1).Format("2006-01-02"),
	)
}

// periodFn returns the SQL expression to bucket a date field by day or month.
func periodFn(field string, monthly bool) string {
	if monthly {
		return fmt.Sprintf("toStartOfMonth(toDate(%s))", field)
	}
	return fmt.Sprintf("toDate(%s)", field)
}

// GetSummaryForUser returns aggregated user-level metrics from pinterest_user_insights
// joined with the latest follower count from pinterest_users.
// Insights are first deduplicated per (user_id, date) then summed, matching the PHP builder pattern.
func (r *Repository) GetSummaryForUser(ctx context.Context, params *ch.QueryParams) (*SummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ptDateFilter("created_at", params)

	query := fmt.Sprintf(`
		SELECT
			toInt64(sum(pu.follower_count))                                                        AS follower_count,
			toInt64(sum(pui.impressions))                                                          AS impressions,
			toInt64(sum(pui.pin_clicks))                                                           AS pin_clicks,
			toInt64(sum(pui.outbound_clicks))                                                      AS outbound_clicks,
			toInt64(sum(pui.saves))                                                                AS saves,
			toInt64(sum(pui.engagement))                                                           AS total_engagement
		FROM (
			SELECT
				user_id,
				sum(impression)     AS impressions,
				sum(pin_clicks)     AS pin_clicks,
				sum(outbound_click) AS outbound_clicks,
				sum(saves)          AS saves,
				sum(engagement)     AS engagement
			FROM (
				SELECT
					user_id,
					max(impression)     AS impression,
					max(pin_clicks)     AS pin_clicks,
					max(outbound_click) AS outbound_click,
					max(saves)          AS saves,
					max(engagement)     AS engagement
				FROM pinterest_user_insights
				WHERE user_id IN %s
				  AND %s
				GROUP BY user_id, toDate(created_at)
			)
			GROUP BY user_id
		) AS pui
		LEFT JOIN (
			SELECT
				user_id,
				argMax(follower_count, inserted_at) AS follower_count
			FROM pinterest_users
			WHERE user_id IN %s
			GROUP BY user_id
		) AS pu USING user_id`,
		ids, dateFilter, ids,
	)

	var result SummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.FollowerCount,
		&result.Impressions,
		&result.PinClicks,
		&result.OutboundClicks,
		&result.Saves,
		&result.TotalEngagement,
	)
	return &result, err
}

// GetSummaryForBoard returns aggregated board-level metrics from pinterest_pin_insights
// for pins belonging to the specified boards.
// boardIDs is a pre-formatted SQL IN clause string like ('id1','id2').
// Matches PHP summaryQueryForBoard: board follower_count from pinterest_boards (argMax),
// pin metrics deduplicated per (pin_id, date) then summed, joined via CROSS JOIN.
func (r *Repository) GetSummaryForBoard(ctx context.Context, params *ch.QueryParams, boardIDs string) (*SummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	boardDateFilter := ptDateFilter("inserted_at", params)
	insightDateFilter := ptDateFilter("created_at", params)

	query := fmt.Sprintf(`
		WITH pins AS (
			SELECT pin_id FROM pinterest_pins WHERE board_id IN %s GROUP BY pin_id
		)
		SELECT
			toInt64(any(pb.follower_count))                                                        AS follower_count,
			toInt64(sum(ppi.impressions))                                                          AS impressions,
			toInt64(sum(ppi.pin_clicks))                                                           AS pin_clicks,
			toInt64(sum(ppi.outbound_clicks))                                                      AS outbound_clicks,
			toInt64(sum(ppi.saves))                                                                AS saves,
			toInt64(sum(ppi.engagement))                                                           AS total_engagement
		FROM (
			SELECT
				board_id,
				argMax(follower_count, inserted_at) AS follower_count
			FROM pinterest_boards
			WHERE board_id IN %s
			  AND %s
			GROUP BY board_id
		) pb
		CROSS JOIN (
			SELECT
				sum(impression)     AS impressions,
				sum(pin_clicks)     AS pin_clicks,
				sum(outbound_click) AS outbound_clicks,
				sum(saves)          AS saves,
				sum(engagement)     AS engagement
			FROM (
				SELECT
					max(impression)     AS impression,
					max(pin_clicks)     AS pin_clicks,
					max(outbound_click) AS outbound_click,
					max(saves)          AS saves,
					max(engagement)     AS engagement
				FROM pinterest_pin_insights
				WHERE pin_id IN (SELECT pin_id FROM pins)
				  AND %s
				GROUP BY pin_id, toDate(created_at)
			)
		) ppi`,
		boardIDs,
		boardIDs, boardDateFilter,
		insightDateFilter,
	)
	_ = ids

	var result SummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.FollowerCount,
		&result.Impressions,
		&result.PinClicks,
		&result.OutboundClicks,
		&result.Saves,
		&result.TotalEngagement,
	)
	return &result, err
}

// GetFollowerTrendForUser returns time-series follower counts from pinterest_users.
// When daily is false, data is aggregated by calendar month.
func (r *Repository) GetFollowerTrendForUser(ctx context.Context, params *ch.QueryParams, daily bool) (*FollowerTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	fill := ptDateFill(params, !daily)
	bucket := periodFn("inserted_at", !daily)

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0)                                                                       AS show_data,
			arrayDifference(arrayFill(x -> not x == 0, groupArray(follower_count)))                     AS followers_daily,
			arrayFill(x -> not x == 0, groupArray(follower_count))                                      AS followers_gained,
			groupArray(bucket)                                                                           AS buckets
		FROM (
			SELECT
				%s                                                                                       AS bucket,
				toInt32(argMin(follower_count, inserted_at))                                             AS follower_count
			FROM pinterest_users
			WHERE user_id IN %s
			  AND toDate(inserted_at) BETWEEN toDate('%s') AND toDate('%s')
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucket, ids,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
		fill,
	)

	var result FollowerTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.FollowersDaily,
		&result.FollowersGained,
		&result.Buckets,
	)
	return &result, err
}

// GetFollowerTrendForBoard returns time-series follower counts from pinterest_boards.
// boardIDs is a pre-formatted SQL IN clause string.
func (r *Repository) GetFollowerTrendForBoard(ctx context.Context, params *ch.QueryParams, boardIDs string, daily bool) (*FollowerTrendResult, error) {
	fill := ptDateFill(params, !daily)
	bucket := periodFn("inserted_at", !daily)

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0)                                                                       AS show_data,
			arrayDifference(arrayFill(x -> not x == 0, groupArray(follower_count)))                     AS followers_daily,
			arrayFill(x -> not x == 0, groupArray(follower_count))                                      AS followers_gained,
			groupArray(bucket)                                                                           AS buckets
		FROM (
			SELECT
				%s                                                                                       AS bucket,
				toInt32(argMin(follower_count, inserted_at))                                             AS follower_count
			FROM pinterest_boards
			WHERE board_id IN %s
			  AND toDate(inserted_at) BETWEEN toDate('%s') AND toDate('%s')
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucket, boardIDs,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
		fill,
	)

	var result FollowerTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.FollowersDaily,
		&result.FollowersGained,
		&result.Buckets,
	)
	return &result, err
}

// GetImpressionsTrendForUser returns time-series impression data from pinterest_user_insights.
// When daily is false, data is aggregated by calendar month.
func (r *Repository) GetImpressionsTrendForUser(ctx context.Context, params *ch.QueryParams, daily bool) (*ImpressionsTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ptDateFilter("created_at", params)
	fill := ptDateFill(params, !daily)
	bucket := periodFn("created_at", !daily)

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0)                                                                                       AS show_data,
			groupArray(toInt32(imp_daily))                                                                               AS impressions_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(imp_daily))))                                     AS impressions_total,
			groupArray(bucket)                                                                                           AS buckets
		FROM (
			SELECT
				%s                                                                                                       AS bucket,
				toInt32(sum(impression))                                                                                 AS imp_daily
			FROM pinterest_user_insights
			WHERE user_id IN %s
			  AND %s
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucket, ids, dateFilter, fill,
	)

	var result ImpressionsTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.ImpressionsDaily,
		&result.ImpressionsTotal,
		&result.Buckets,
	)
	return &result, err
}

// GetImpressionsTrendForBoard returns time-series impression data from pinterest_pin_insights
// for pins belonging to the specified boards.
func (r *Repository) GetImpressionsTrendForBoard(ctx context.Context, params *ch.QueryParams, boardIDs string, daily bool) (*ImpressionsTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ptDateFilter("created_at", params)
	insightDateFilter := ptDateFilter("created_at", params)
	fill := ptDateFill(params, !daily)
	bucket := periodFn("created_at", !daily)

	query := fmt.Sprintf(`
		WITH pins AS (
			SELECT pin_id FROM pinterest_pins WHERE board_id IN %s AND user_id IN %s
		)
		SELECT
			if(count() > 0, 1, 0)                                                                                       AS show_data,
			groupArray(toInt32(imp_daily))                                                                               AS impressions_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(imp_daily))))                                     AS impressions_total,
			groupArray(bucket)                                                                                           AS buckets
		FROM (
			SELECT
				%s                                                                                                       AS bucket,
				toInt32(sum(impression))                                                                                 AS imp_daily
			FROM pinterest_pin_insights
			WHERE pin_id IN (SELECT pin_id FROM pins)
			  AND %s
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		boardIDs, ids,
		bucket, insightDateFilter, fill,
	)
	_ = dateFilter

	var result ImpressionsTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.ImpressionsDaily,
		&result.ImpressionsTotal,
		&result.Buckets,
	)
	return &result, err
}

// GetEngagementTrendForUser returns time-series engagement data from pinterest_user_insights
// broken down by saves, outbound clicks, pin clicks, and total engagement.
func (r *Repository) GetEngagementTrendForUser(ctx context.Context, params *ch.QueryParams, daily bool) (*EngagementTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ptDateFilter("created_at", params)
	fill := ptDateFill(params, !daily)
	bucket := periodFn("created_at", !daily)

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0)                                                                                               AS show_data,
			groupArray(toInt32(saves_d))                                                                                         AS saves_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(saves_d))))                                                AS saves_total,
			groupArray(toInt32(outbound_clicks_d))                                                                               AS outbound_clicks_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(outbound_clicks_d))))                                      AS outbound_clicks_total,
			groupArray(toInt32(pin_clicks_d))                                                                                    AS pin_clicks_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(pin_clicks_d))))                                           AS pin_clicks_total,
			groupArray(toInt32(engagement_d))                                                                                    AS engagement_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(engagement_d))))                                           AS engagement_total,
			groupArray(bucket)                                                                                                   AS buckets
		FROM (
			SELECT
				%s                                                                                                               AS bucket,
				toInt32(sum(saves))          AS saves_d,
				toInt32(sum(outbound_click)) AS outbound_clicks_d,
				toInt32(sum(pin_clicks))     AS pin_clicks_d,
				toInt32(sum(engagement))     AS engagement_d
			FROM pinterest_user_insights
			WHERE user_id IN %s
			  AND %s
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucket, ids, dateFilter, fill,
	)

	var result EngagementTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.SavesDaily,
		&result.SavesTotal,
		&result.OutboundClicksDaily,
		&result.OutboundClicksTotal,
		&result.PinClicksDaily,
		&result.PinClicksTotal,
		&result.EngagementDaily,
		&result.EngagementTotal,
		&result.Buckets,
	)
	return &result, err
}

// GetEngagementTrendForBoard returns time-series engagement data from pinterest_pin_insights
// for pins belonging to the specified boards.
func (r *Repository) GetEngagementTrendForBoard(ctx context.Context, params *ch.QueryParams, boardIDs string, daily bool) (*EngagementTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ptDateFilter("created_at", params)
	fill := ptDateFill(params, !daily)
	bucket := periodFn("created_at", !daily)

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0)                                                                                               AS show_data,
			groupArray(toInt32(saves_d))                                                                                         AS saves_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(saves_d))))                                                AS saves_total,
			groupArray(toInt32(outbound_clicks_d))                                                                               AS outbound_clicks_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(outbound_clicks_d))))                                      AS outbound_clicks_total,
			groupArray(toInt32(pin_clicks_d))                                                                                    AS pin_clicks_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(pin_clicks_d))))                                           AS pin_clicks_total,
			groupArray(toInt32(engagement_d))                                                                                    AS engagement_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(engagement_d))))                                           AS engagement_total,
			groupArray(bucket)                                                                                                   AS buckets
		FROM (
			SELECT
				%s                                                                                                               AS bucket,
				toInt32(sum(saves))          AS saves_d,
				toInt32(sum(outbound_click)) AS outbound_clicks_d,
				toInt32(sum(pin_clicks))     AS pin_clicks_d,
				toInt32(sum(engagement))     AS engagement_d
			FROM pinterest_pin_insights
			WHERE pin_id IN (
				SELECT pin_id FROM pinterest_pins WHERE board_id IN %s AND user_id IN %s
			)
			  AND %s
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucket, boardIDs, ids, dateFilter, fill,
	)

	var result EngagementTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.SavesDaily,
		&result.SavesTotal,
		&result.OutboundClicksDaily,
		&result.OutboundClicksTotal,
		&result.PinClicksDaily,
		&result.PinClicksTotal,
		&result.EngagementDaily,
		&result.EngagementTotal,
		&result.Buckets,
	)
	return &result, err
}

// GetPinPostingForUser returns time-series pin publication counts from pinterest_pins for owner pins.
// filterBy optionally restricts to a specific media_type ('video' or 'image'); empty means all types.
func (r *Repository) GetPinPostingForUser(ctx context.Context, params *ch.QueryParams, filterBy string, daily bool) (*PinPostingResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ptDateFilter("created_at", params)
	fill := ptDateFill(params, !daily)
	bucket := periodFn("created_at", !daily)

	mediaFilter := ""
	if filterBy != "" {
		mediaFilter = fmt.Sprintf(" AND media_type = '%s'", filterBy)
	}

	query := fmt.Sprintf(`
		WITH pins AS (
			SELECT
				pin_id,
				%s AS bucket
			FROM pinterest_pins
			WHERE user_id IN %s
			  AND is_owner = 1
			  %s
			  AND %s
			GROUP BY pin_id, bucket
		)
		SELECT
			if(count() > 0, 1, 0)         AS show_data,
			groupArray(toInt32(pin_count)) AS pins_count,
			groupArray(bucket)             AS buckets
		FROM (
			SELECT
				bucket,
				toInt32(count())          AS pin_count
			FROM pins
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucket, ids, mediaFilter, dateFilter, fill,
	)

	var result PinPostingResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.PinsCount,
		&result.Buckets,
	)
	return &result, err
}

// GetPinPostingForBoard returns time-series pin publication counts for pins in the specified boards.
func (r *Repository) GetPinPostingForBoard(ctx context.Context, params *ch.QueryParams, boardIDs string, filterBy string, daily bool) (*PinPostingResult, error) {
	dateFilter := ptDateFilter("created_at", params)
	fill := ptDateFill(params, !daily)
	bucket := periodFn("created_at", !daily)

	mediaFilter := ""
	if filterBy != "" {
		mediaFilter = fmt.Sprintf(" AND media_type = '%s'", filterBy)
	}

	query := fmt.Sprintf(`
		WITH pins AS (
			SELECT
				pin_id,
				%s AS bucket
			FROM pinterest_pins
			WHERE board_id IN %s
			  AND is_owner = 1
			  %s
			  AND %s
			GROUP BY pin_id, bucket
		)
		SELECT
			if(count() > 0, 1, 0)         AS show_data,
			groupArray(toInt32(pin_count)) AS pins_count,
			groupArray(bucket)             AS buckets
		FROM (
			SELECT
				bucket,
				toInt32(count())          AS pin_count
			FROM pins
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucket, boardIDs, mediaFilter, dateFilter, fill,
	)

	var result PinPostingResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.PinsCount,
		&result.Buckets,
	)
	return &result, err
}

// GetPinRollupForUser returns aggregated pin-level metrics for all user pins in the date range.
// Insights are not date-filtered — all historical metrics for pins posted in the period are summed,
// matching PHP's pinPostingRollupQueryForUser which has no date filter on pinterest_pin_insights.
func (r *Repository) GetPinRollupForUser(ctx context.Context, params *ch.QueryParams) (*PinRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	pinDateFilter := ptDateFilter("created_at", params)

	query := fmt.Sprintf(`
		WITH
			pins AS (
				SELECT pin_id FROM pinterest_pins WHERE user_id IN %s AND %s
			),
			pin_count AS (
				SELECT toInt64(count()) AS total_pins FROM pins
			),
			insights AS (
				SELECT
					toInt64(sum(impression))                    AS impressions,
					toInt64(sum(pin_clicks))                    AS pin_clicks,
					toInt64(sum(outbound_click))                AS outbound_clicks,
					toInt64(sum(saves))                         AS saves,
					round(avg(quartile_95s_percent_view), 2)    AS quartile_95s_percent_view,
					toInt64(sum(video_start))                   AS video_views,
					toInt64(sum(video_10s_view))                AS video_10s_views,
					round(avg(video_avg_watch_time), 2)         AS avg_watch_time
				FROM pinterest_pin_insights
				WHERE pin_id IN (SELECT pin_id FROM pins)
			)
		SELECT
			pc.total_pins,
			i.impressions,
			i.pin_clicks,
			i.outbound_clicks,
			i.saves,
			i.quartile_95s_percent_view,
			i.video_views,
			i.video_10s_views,
			i.avg_watch_time
		FROM insights i
		CROSS JOIN pin_count pc`,
		ids, pinDateFilter,
	)

	var result PinRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalPins,
		&result.Impressions,
		&result.PinClicks,
		&result.OutboundClicks,
		&result.Saves,
		&result.QuartilePercView,
		&result.VideoViews,
		&result.Video10sViews,
		&result.AvgWatchTime,
	)
	return &result, err
}

// GetPinRollupForBoard returns aggregated pin-level metrics for pins in the specified boards.
// Insights are not date-filtered — all historical metrics for pins posted in the period are summed,
// matching PHP's pinPostingRollupQueryForBoard which has no date filter on pinterest_pin_insights.
func (r *Repository) GetPinRollupForBoard(ctx context.Context, params *ch.QueryParams, boardIDs string) (*PinRollupResult, error) {
	pinDateFilter := ptDateFilter("created_at", params)

	query := fmt.Sprintf(`
		WITH
			pins AS (
				SELECT pin_id FROM pinterest_pins WHERE board_id IN %s AND %s
			),
			pin_count AS (
				SELECT toInt64(count()) AS total_pins FROM pins
			),
			insights AS (
				SELECT
					toInt64(sum(impression))                    AS impressions,
					toInt64(sum(pin_clicks))                    AS pin_clicks,
					toInt64(sum(outbound_click))                AS outbound_clicks,
					toInt64(sum(saves))                         AS saves,
					round(avg(quartile_95s_percent_view), 2)    AS quartile_95s_percent_view,
					toInt64(sum(video_start))                   AS video_views,
					toInt64(sum(video_10s_view))                AS video_10s_views,
					round(avg(video_avg_watch_time), 2)         AS avg_watch_time
				FROM pinterest_pin_insights
				WHERE pin_id IN (SELECT pin_id FROM pins)
			)
		SELECT
			pc.total_pins,
			i.impressions,
			i.pin_clicks,
			i.outbound_clicks,
			i.saves,
			i.quartile_95s_percent_view,
			i.video_views,
			i.video_10s_views,
			i.avg_watch_time
		FROM insights i
		CROSS JOIN pin_count pc`,
		boardIDs, pinDateFilter,
	)

	var result PinRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalPins,
		&result.Impressions,
		&result.PinClicks,
		&result.OutboundClicks,
		&result.Saves,
		&result.QuartilePercView,
		&result.VideoViews,
		&result.Video10sViews,
		&result.AvgWatchTime,
	)
	return &result, err
}

// validPinOrderByFields is the whitelist of allowed ORDER BY columns to prevent SQL injection.
var validPinOrderByFields = map[string]bool{
	"impressions":      true,
	"pin_clicks":       true,
	"outbound_clicks":  true,
	"saves":            true,
	"total_engagement": true,
	"engagement_rate":  true,
}

// safePinOrderBy validates and returns the order-by field, defaulting to impressions.
func safePinOrderBy(orderBy string) string {
	if validPinOrderByFields[orderBy] {
		return orderBy
	}
	return "impressions"
}

// GetPinsForUser returns per-pin metrics for all user pins sorted by the specified metric.
// orderBy must be one of the valid fields; ascending controls sort direction; limit caps the result set.
func (r *Repository) GetPinsForUser(ctx context.Context, params *ch.QueryParams, orderBy string, limit int, ascending bool) ([]PinRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ptDateFilter("created_at", params)
	direction := "DESC"
	if ascending {
		direction = "ASC"
	}
	safeOrder := safePinOrderBy(orderBy)

	query := fmt.Sprintf(`
		SELECT
			pins.pin_id,
			last_value(pb.name)                                                                                             AS board_name,
			format('https://www.pinterest.com/pin/{0}/', pins.pin_id)                                                      AS permalink,
			format('https://www.pinterest.com/pin/{0}/', pins.pin_id)                                                      AS embed_link,
			last_value(pins.title)                                                                                          AS title,
			last_value(pins.description)                                                                                    AS description,
			last_value(pins.board_owner)                                                                                    AS board_owner,
			replaceOne(last_value(pins.media_type), '_', ' ')                                                               AS media_type,
			last_value(pins.cover_image_url)                                                                                AS cover_image_url,
			last_value(pins.dominant_color)                                                                                 AS dominant_color,
			last_value(pins.creative_type)                                                                                  AS creative_type,
			last_value(pins.product_tags)                                                                                   AS product_tags,
			toInt64OrZero(last_value(pins.height))                                                                          AS height,
			toInt64OrZero(last_value(pins.width))                                                                           AS width,
			last_value(pins.created_at)                                                                                     AS created_at,
			toInt64(sum(ppi.impression))                                                                                    AS impressions,
			toInt64(sum(ppi.pin_clicks))                                                                                    AS pin_clicks,
			toInt64(sum(ppi.outbound_click))                                                                                AS outbound_clicks,
			toInt64(sum(ppi.saves))                                                                                         AS saves,
			toInt64(sum(ppi.engagement))                                                                                    AS total_engagement,
			round(if(count() > 0, toFloat64(sum(ppi.pin_clicks) + sum(ppi.outbound_click) + sum(ppi.saves)) / count(), 0), 2) AS engagement_rate
		FROM (
			SELECT pin_id, board_id,
				last_value(title)           AS title,
				last_value(description)     AS description,
				last_value(board_owner)     AS board_owner,
				last_value(media_type)      AS media_type,
				last_value(cover_image_url) AS cover_image_url,
				last_value(dominant_color)  AS dominant_color,
				last_value(creative_type)   AS creative_type,
				last_value(product_tags)    AS product_tags,
				last_value(height)          AS height,
				last_value(width)           AS width,
				created_at
			FROM pinterest_pins
			WHERE user_id IN %s
			  AND %s
			  AND is_owner = 1
			GROUP BY pin_id, board_id, created_at
		) AS pins
		LEFT JOIN (
			SELECT board_id, last_value(name) AS name
			FROM pinterest_boards
			GROUP BY board_id
		) AS pb ON pins.board_id = pb.board_id
		LEFT JOIN (
			SELECT pin_id,
				max(impression)    AS impression,
				max(pin_clicks)    AS pin_clicks,
				max(outbound_click) AS outbound_click,
				max(saves)         AS saves,
				max(engagement)    AS engagement
			FROM pinterest_pin_insights
			WHERE user_id IN %s
			GROUP BY pin_id, created_at
			ORDER BY created_at DESC
		) AS ppi USING pin_id
		GROUP BY pins.pin_id, pins.board_id
		ORDER BY %s %s
		LIMIT %d`,
		ids, dateFilter, ids, safeOrder, direction, limit,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PinRow
	for rows.Next() {
		var row PinRow
		if err := rows.Scan(
			&row.PinID,
			&row.BoardName,
			&row.Permalink,
			&row.EmbedLink,
			&row.Title,
			&row.Description,
			&row.BoardOwner,
			&row.MediaType,
			&row.CoverImageURL,
			&row.DominantColor,
			&row.CreativeType,
			&row.ProductTags,
			&row.Height,
			&row.Width,
			&row.CreatedAt,
			&row.Impressions,
			&row.PinClicks,
			&row.OutboundClicks,
			&row.Saves,
			&row.TotalEngagement,
			&row.EngagementRate,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// GetPinsForBoard returns per-pin metrics for pins in the specified boards sorted by the specified metric.
func (r *Repository) GetPinsForBoard(ctx context.Context, params *ch.QueryParams, boardIDs string, orderBy string, limit int, ascending bool) ([]PinRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ptDateFilter("created_at", params)
	direction := "DESC"
	if ascending {
		direction = "ASC"
	}
	safeOrder := safePinOrderBy(orderBy)

	query := fmt.Sprintf(`
		SELECT
			pins.pin_id,
			last_value(pb.name)                                                                                             AS board_name,
			format('https://www.pinterest.com/pin/{0}/', pins.pin_id)                                                      AS permalink,
			format('https://www.pinterest.com/pin/{0}/', pins.pin_id)                                                      AS embed_link,
			last_value(pins.title)                                                                                          AS title,
			last_value(pins.description)                                                                                    AS description,
			last_value(pins.board_owner)                                                                                    AS board_owner,
			replaceOne(last_value(pins.media_type), '_', ' ')                                                               AS media_type,
			last_value(pins.cover_image_url)                                                                                AS cover_image_url,
			last_value(pins.dominant_color)                                                                                 AS dominant_color,
			last_value(pins.creative_type)                                                                                  AS creative_type,
			last_value(pins.product_tags)                                                                                   AS product_tags,
			toInt64OrZero(last_value(pins.height))                                                                          AS height,
			toInt64OrZero(last_value(pins.width))                                                                           AS width,
			last_value(pins.created_at)                                                                                     AS created_at,
			toInt64(sum(ppi.impression))                                                                                    AS impressions,
			toInt64(sum(ppi.pin_clicks))                                                                                    AS pin_clicks,
			toInt64(sum(ppi.outbound_click))                                                                                AS outbound_clicks,
			toInt64(sum(ppi.saves))                                                                                         AS saves,
			toInt64(sum(ppi.engagement))                                                                                    AS total_engagement,
			round(if(count() > 0, toFloat64(sum(ppi.pin_clicks) + sum(ppi.outbound_click) + sum(ppi.saves)) / count(), 0), 2) AS engagement_rate
		FROM (
			SELECT pin_id, board_id,
				last_value(title)           AS title,
				last_value(description)     AS description,
				last_value(board_owner)     AS board_owner,
				last_value(media_type)      AS media_type,
				last_value(cover_image_url) AS cover_image_url,
				last_value(dominant_color)  AS dominant_color,
				last_value(creative_type)   AS creative_type,
				last_value(product_tags)    AS product_tags,
				last_value(height)          AS height,
				last_value(width)           AS width,
				created_at
			FROM pinterest_pins
			WHERE board_id IN %s
			  AND %s
			  AND is_owner = 1
			GROUP BY pin_id, board_id, created_at
		) AS pins
		LEFT JOIN (
			SELECT board_id, last_value(name) AS name
			FROM pinterest_boards
			GROUP BY board_id
		) AS pb ON pins.board_id = pb.board_id
		LEFT JOIN (
			SELECT pin_id,
				max(impression)     AS impression,
				max(pin_clicks)     AS pin_clicks,
				max(outbound_click) AS outbound_click,
				max(saves)          AS saves,
				max(engagement)     AS engagement
			FROM pinterest_pin_insights
			WHERE user_id IN %s
			GROUP BY pin_id, created_at
			ORDER BY created_at DESC
		) AS ppi USING pin_id
		GROUP BY pins.pin_id, pins.board_id
		ORDER BY %s %s
		LIMIT %d`,
		boardIDs, dateFilter, ids, safeOrder, direction, limit,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PinRow
	for rows.Next() {
		var row PinRow
		if err := rows.Scan(
			&row.PinID,
			&row.BoardName,
			&row.Permalink,
			&row.EmbedLink,
			&row.Title,
			&row.Description,
			&row.BoardOwner,
			&row.MediaType,
			&row.CoverImageURL,
			&row.DominantColor,
			&row.CreativeType,
			&row.ProductTags,
			&row.Height,
			&row.Width,
			&row.CreatedAt,
			&row.Impressions,
			&row.PinClicks,
			&row.OutboundClicks,
			&row.Saves,
			&row.TotalEngagement,
			&row.EngagementRate,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// GetPinPerformanceForUser returns daily time-series performance metrics grouped by pin creation date.
// Each bucket represents a day; metrics are the sum across all pins created on that day.
// Insights are not date-filtered — all historical metrics for pins posted in the period are summed,
// matching PHP's pinPerformanceQuery which has no date filter on pinterest_pin_insights.
func (r *Repository) GetPinPerformanceForUser(ctx context.Context, params *ch.QueryParams) (*PinPerformanceResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	pinDateFilter := ptDateFilter("pp.created_at", params)

	query := fmt.Sprintf(`
		WITH
			pin_data AS (
				SELECT pp.pin_id, toDate(pp.created_at) AS pin_date
				FROM pinterest_pins pp
				WHERE pp.user_id IN %s AND %s
				GROUP BY pp.pin_id, pin_date
			),
			pin_metrics AS (
				SELECT
					ppi.pin_id,
					toInt32(sum(ppi.pin_clicks))      AS pin_clicks,
					toInt32(sum(ppi.outbound_click))  AS outbound_clicks,
					toInt32(sum(ppi.saves))           AS saves,
					toInt32(sum(ppi.engagement))      AS engagement,
					toInt32(sum(ppi.impression))      AS impressions
				FROM pinterest_pin_insights ppi
				WHERE ppi.pin_id IN (SELECT pin_id FROM pin_data)
				GROUP BY ppi.pin_id
			),
			daily_stats AS (
				SELECT
					pd.pin_date                      AS bucket,
					toInt32(count())                 AS pins_count,
					toInt32(sum(pm.pin_clicks))      AS pin_clicks,
					toInt32(sum(pm.outbound_clicks)) AS outbound_clicks,
					toInt32(sum(pm.saves))           AS saves,
					toInt32(sum(pm.engagement))      AS engagements,
					toInt32(sum(pm.impressions))     AS impressions
				FROM pin_data pd
				LEFT JOIN pin_metrics pm ON pd.pin_id = pm.pin_id
				GROUP BY bucket
				ORDER BY bucket ASC
				WITH FILL FROM toDate('%s') TO toDate('%s') STEP 1
			)
		SELECT
			if(count() > 0, 1, 0)                AS show_data,
			groupArray(toInt32(pins_count))       AS pins_count,
			groupArray(toInt32(pin_clicks))       AS pin_clicks,
			groupArray(toInt32(outbound_clicks))  AS outbound_clicks,
			groupArray(toInt32(saves))            AS saves,
			groupArray(toInt32(engagements))      AS engagements,
			groupArray(toInt32(impressions))      AS impressions,
			groupArray(bucket)                    AS buckets
		FROM daily_stats`,
		ids, pinDateFilter,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.AddDate(0, 0, 1).Format("2006-01-02"),
	)

	var result PinPerformanceResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.PinsCount,
		&result.PinClicks,
		&result.OutboundClicks,
		&result.Saves,
		&result.Engagements,
		&result.Impressions,
		&result.Buckets,
	)
	return &result, err
}

// GetPinPerformanceForBoard returns daily time-series performance metrics for pins in the specified boards.
// Insights are not date-filtered — all historical metrics for pins posted in the period are summed,
// matching PHP's pinPerformanceQueryForBoard which has no date filter on pinterest_pin_insights.
func (r *Repository) GetPinPerformanceForBoard(ctx context.Context, params *ch.QueryParams, boardIDs string) (*PinPerformanceResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	pinDateFilter := ptDateFilter("pp.created_at", params)

	query := fmt.Sprintf(`
		WITH
			pin_data AS (
				SELECT pp.pin_id, toDate(pp.created_at) AS pin_date
				FROM pinterest_pins pp
				WHERE pp.user_id IN %s AND pp.board_id IN %s AND %s
				GROUP BY pp.pin_id, pin_date
			),
			pin_metrics AS (
				SELECT
					ppi.pin_id,
					toInt32(sum(ppi.pin_clicks))      AS pin_clicks,
					toInt32(sum(ppi.outbound_click))  AS outbound_clicks,
					toInt32(sum(ppi.saves))           AS saves,
					toInt32(sum(ppi.engagement))      AS engagement,
					toInt32(sum(ppi.impression))      AS impressions
				FROM pinterest_pin_insights ppi
				WHERE ppi.pin_id IN (SELECT pin_id FROM pin_data)
				GROUP BY ppi.pin_id
			),
			daily_stats AS (
				SELECT
					pd.pin_date                      AS bucket,
					toInt32(count())                 AS pins_count,
					toInt32(sum(pm.pin_clicks))      AS pin_clicks,
					toInt32(sum(pm.outbound_clicks)) AS outbound_clicks,
					toInt32(sum(pm.saves))           AS saves,
					toInt32(sum(pm.engagement))      AS engagements,
					toInt32(sum(pm.impressions))     AS impressions
				FROM pin_data pd
				LEFT JOIN pin_metrics pm ON pd.pin_id = pm.pin_id
				GROUP BY bucket
				ORDER BY bucket ASC
				WITH FILL FROM toDate('%s') TO toDate('%s') STEP 1
			)
		SELECT
			if(count() > 0, 1, 0)                AS show_data,
			groupArray(toInt32(pins_count))       AS pins_count,
			groupArray(toInt32(pin_clicks))       AS pin_clicks,
			groupArray(toInt32(outbound_clicks))  AS outbound_clicks,
			groupArray(toInt32(saves))            AS saves,
			groupArray(toInt32(engagements))      AS engagements,
			groupArray(toInt32(impressions))      AS impressions,
			groupArray(bucket)                    AS buckets
		FROM daily_stats`,
		ids, boardIDs, pinDateFilter,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.AddDate(0, 0, 1).Format("2006-01-02"),
	)

	var result PinPerformanceResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.PinsCount,
		&result.PinClicks,
		&result.OutboundClicks,
		&result.Saves,
		&result.Engagements,
		&result.Impressions,
		&result.Buckets,
	)
	return &result, err
}
