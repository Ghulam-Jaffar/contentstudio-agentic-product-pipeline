// Package campaign_label provides ClickHouse query builders for campaign/label analytics.
// Queries aggregate cross-platform analytics data (Facebook, Instagram, LinkedIn, TikTok,
// YouTube, Pinterest) for posts identified by their post IDs.
//
// Migrated from PHP: CampaignLabelAnalyticsBuilder (contentstudio-backend).
package campaign_label

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// Repository provides ClickHouse query methods for campaign/label analytics.
type Repository struct {
	client *ch.Client
}

// NewRepository creates a new campaign/label analytics repository.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// dateFilter returns a ClickHouse date filter clause for the given field and period.
// Uses the raw column (no toDateTime wrapper) so ClickHouse can leverage partition pruning.
func dateFilter(field string, startDate, endDate time.Time) string {
	return fmt.Sprintf(
		"%s BETWEEN toDateTime('%s', 0) AND toDateTime('%s', 0)",
		field,
		startDate.Format("2006-01-02 15:04:05"),
		endDate.Format("2006-01-02 15:04:05"),
	)
}

// formatPostIDs formats post IDs for use in ClickHouse arrayJoin.
// Returns a string like: ['id1','id2','id3']
func formatPostIDs(ids []string) string {
	if len(ids) == 0 {
		return "['']"
	}
	quoted := make([]string, len(ids))
	for i, id := range ids {
		quoted[i] = "'" + strings.ReplaceAll(id, "'", "\\'") + "'"
	}
	return "[" + strings.Join(quoted, ",") + "]"
}

func facebookPermalinkIDExpr(column string) string {
	return fmt.Sprintf(
		"if(position(toString(%[1]s), '_') > 0, arrayElement(splitByChar('_', toString(%[1]s)), length(splitByChar('_', toString(%[1]s)))), toString(%[1]s))",
		column,
	)
}

func fbPageFilterClause(pageIDs []string, alias string) string {
	if len(pageIDs) == 0 {
		return ""
	}
	return fmt.Sprintf("AND %s.page_id IN %s", alias, ch.FormatAccountIDs(pageIDs))
}

// fbDateFilterClause builds an AND clause to limit Facebook scans by created_time.
// Returns empty string if startDate/endDate are zero, enabling callers without date ranges.
func fbDateFilterClause(startDate, endDate time.Time, alias string) string {
	if startDate.IsZero() && endDate.IsZero() {
		return ""
	}
	return fmt.Sprintf(
		"AND %s.created_time >= toDateTime('%s', 0) AND %s.created_time <= toDateTime('%s', 0)",
		alias,
		startDate.Format("2006-01-02 15:04:05"),
		alias,
		endDate.Format("2006-01-02 15:04:05"),
	)
}

// buildFacebookSummaryCTEs builds the Facebook post ID resolution CTEs.
// dateFilterClause is an optional AND clause (e.g. "AND created_time >= '...' AND created_time <= '...'")
// used to limit the CROSS JOIN permalink scan to relevant partitions.
func buildFacebookSummaryCTEs(dateFilterClause, pageFilterClause, videoPageFilterClause, videoDateFilterClause string) string {
	permalinkID := facebookPermalinkIDExpr("post_id")

	return fmt.Sprintf(`facebook_input_ids AS (
			SELECT DISTINCT
				post_id AS raw_post_id,
				%s AS permalink_post_id
			FROM postIds
		),
		facebook_permalink_ids AS (
			SELECT groupArray(permalink_post_id) AS ids
			FROM facebook_input_ids
		),
		facebook_reels AS (
			SELECT DISTINCT fvi.post_id AS post_id
			FROM contentstudiobackend.facebook_video_insights AS fvi
			WHERE fvi.video_id IN (SELECT raw_post_id FROM facebook_input_ids)
				%s
				%s
		),
		facebook_post_ids AS (
			SELECT post_id
			FROM (
				SELECT DISTINCT raw_post_id AS post_id
				FROM facebook_input_ids
				UNION ALL
				SELECT DISTINCT fp.post_id AS post_id
				FROM contentstudiobackend.facebook_posts AS fp
				WHERE (fp.post_id IN (SELECT raw_post_id FROM facebook_input_ids)
					OR fp.video_id IN (SELECT raw_post_id FROM facebook_input_ids))
					%s
					%s
				UNION ALL
				SELECT DISTINCT fp.post_id AS post_id
				FROM contentstudiobackend.facebook_posts AS fp
				CROSS JOIN facebook_permalink_ids
				WHERE arrayExists(id -> endsWith(toString(fp.permalink), id), ids)
					%s
					%s
				UNION ALL
				SELECT DISTINCT post_id
				FROM facebook_reels
			)
			GROUP BY post_id
		)`, permalinkID, videoPageFilterClause, videoDateFilterClause, pageFilterClause, dateFilterClause, pageFilterClause, dateFilterClause)
}

// buildFacebookGroupedCTEs builds the Facebook grouped post ID resolution CTEs for breakdown/insights queries.
// dateFilterClause is an optional AND clause to limit the CROSS JOIN permalink scan to relevant partitions.
func buildFacebookGroupedCTEs(dateFilterClause, pageFilterClause, videoPageFilterClause, videoDateFilterClause string) string {
	permalinkID := facebookPermalinkIDExpr("post_id")

	return fmt.Sprintf(`facebook_input_ids AS (
			SELECT DISTINCT
				id,
				post_id AS raw_post_id,
				%s AS permalink_post_id
			FROM postIds
		),
		facebook_permalink_match_arrays AS (
			SELECT groupArray((id, permalink_post_id)) AS id_pairs
			FROM facebook_input_ids
		),
		facebook_reels AS (
			SELECT DISTINCT
				fi.id AS id,
				fvi.post_id AS post_id
			FROM facebook_input_ids AS fi
			INNER JOIN contentstudiobackend.facebook_video_insights AS fvi
				ON fvi.video_id = fi.raw_post_id
			WHERE 1=1 %s
				%s
		),
		facebook_post_ids AS (
			SELECT id, post_id
			FROM (
				SELECT DISTINCT
					id,
					raw_post_id AS post_id
				FROM facebook_input_ids
				UNION ALL
				SELECT DISTINCT
					fi.id AS id,
					fp.post_id AS post_id
				FROM facebook_input_ids AS fi
				INNER JOIN contentstudiobackend.facebook_posts AS fp
					ON (fp.post_id = fi.raw_post_id
					OR fp.video_id = fi.raw_post_id)
				WHERE 1=1 %s
					%s
				UNION ALL
				SELECT DISTINCT
					tupleElement(match_pair, 1) AS id,
					fp.post_id AS post_id
				FROM contentstudiobackend.facebook_posts AS fp
				CROSS JOIN facebook_permalink_match_arrays
				ARRAY JOIN arrayFilter(
					pair -> endsWith(toString(fp.permalink), tupleElement(pair, 2)),
					id_pairs
				) AS match_pair
				WHERE 1=1 %s
					%s
				UNION ALL
				SELECT DISTINCT id, post_id
				FROM facebook_reels
			)
			GROUP BY id, post_id
		),
		matched_post_ids AS (
			SELECT id, post_id
			FROM (
				SELECT id, post_id
				FROM postIds
				UNION ALL
				SELECT id, post_id
				FROM facebook_post_ids
			)
			GROUP BY id, post_id
		)`, permalinkID, videoPageFilterClause, videoDateFilterClause, pageFilterClause, dateFilterClause, pageFilterClause, dateFilterClause)
}

// GetSummary executes the cross-platform summary query for the given post IDs and date range.
// The query aggregates data across Facebook, Instagram, LinkedIn, TikTok, YouTube, and Pinterest
// ClickHouse tables, combining total posts, engagement, and impressions.
func (r *Repository) GetSummary(ctx context.Context, postIDs []string, params *ch.QueryParams, flagSetup map[string]bool) (*SummaryResult, error) {
	if len(postIDs) == 0 {
		return &SummaryResult{}, nil
	}

	formattedIDs := formatPostIDs(postIDs)
	currentFilter := func(field string) string {
		return dateFilter(field, params.DateFrom, params.DateTo.Add(24*time.Hour))
	}

	// Build the postIds CTE
	postIdCTE := fmt.Sprintf("postIds AS (SELECT arrayJoin(%s) AS post_id)", formattedIDs)

	// Build per-platform CTEs (same structure as PHP CampaignLabelAnalyticsBuilder.getSummaryQuery)
	var platformCTEs []string

	if flagSetup["facebook"] {
		postIdCTE += ", " + buildFacebookSummaryCTEs(
			fbDateFilterClause(params.DateFrom, params.DateTo.Add(24*time.Hour), "fp"),
			fbPageFilterClause(params.FacebookIDs, "fp"),
			fbPageFilterClause(params.FacebookIDs, "fvi"),
			fbDateFilterClause(params.DateFrom, params.DateTo.Add(24*time.Hour), "fvi"),
		)

		platformCTEs = append(platformCTEs, fmt.Sprintf(`SELECT
			toInt32(count()) as total_posts,
			toInt32(sum(total_engagement)) as total_engagements,
			toInt32(sum(total_impressions)) as total_impression
			FROM (
				SELECT argMax(total_engagement, saving_time) as total_engagement,
						argMax(post_impressions, saving_time) as total_impressions
				FROM contentstudiobackend.facebook_posts
				WHERE post_id IN (SELECT post_id FROM facebook_post_ids)
					AND %s
				GROUP BY post_id
			)`, currentFilter("created_time")))
	}

	if flagSetup["instagram"] {
		platformCTEs = append(platformCTEs, fmt.Sprintf(`SELECT
			toInt32(count()) as total_posts,
			toInt32(sum(engagement)) as total_engagements,
			toInt32(sum(views)) as total_impression
			FROM (
				SELECT argMax(engagement, stored_event_at) as engagement,
						argMax(views, stored_event_at) as views
				FROM contentstudiobackend.instagram_posts
				WHERE media_id IN (SELECT post_id FROM postIds) AND %s
				GROUP BY media_id
			)`, currentFilter("post_created_at")))
	}

	if flagSetup["linkedin"] {
		platformCTEs = append(platformCTEs, fmt.Sprintf(`SELECT
			toInt32(count()) as total_posts,
			toInt32(sum(total_engagement)) as total_engagements,
			toInt32(sum(impressions)) as total_impression
			FROM (
				SELECT argMax(total_engagement, saving_time) as total_engagement,
						argMax(impressions, saving_time) as impressions
				FROM contentstudiobackend.linkedin_posts
				WHERE activity IN (SELECT post_id FROM postIds) AND %s
				GROUP BY activity
			)`, currentFilter("created_at")))
	}

	if flagSetup["tiktok"] {
		platformCTEs = append(platformCTEs, fmt.Sprintf(`SELECT
			toInt32(count()) as total_posts,
			toInt32(sum(total_engagement)) as total_engagements,
			toInt32(sum(view_count)) as total_impression
			FROM (
				SELECT argMax(engagement_count, inserted_at) as total_engagement,
						argMax(view_count, inserted_at) as view_count
				FROM contentstudiobackend.tiktok_posts
				WHERE post_id IN (SELECT post_id FROM postIds) AND %s
				GROUP BY post_id
			)`, currentFilter("created_at")))
	}

	if flagSetup["youtube"] {
		platformCTEs = append(platformCTEs, fmt.Sprintf(`SELECT
			toInt32(count()) as total_posts,
			toInt32(sum(total_engagement)) as total_engagements,
			toInt32(sum(views)) as total_impression
			FROM (
				SELECT argMax(likes, inserted_at) + argMax(comments, inserted_at) + argMax(shares, inserted_at) + argMax(dislikes, inserted_at) as total_engagement,
						argMax(views, inserted_at) as views
				FROM contentstudiobackend.youtube_videos
				WHERE video_id IN (SELECT post_id FROM postIds) AND %s
				GROUP BY video_id
			)`, currentFilter("published_at")))
	}

	if flagSetup["pinterest"] {
		platformCTEs = append(platformCTEs, fmt.Sprintf(`WITH
			pins AS (
				SELECT pin_id
				FROM contentstudiobackend.pinterest_pins
				WHERE pin_id IN (SELECT post_id FROM postIds)
					AND %s
				GROUP BY pin_id
			),
			pinterest_insights AS (
				SELECT pin_id,
					toInt32(SUM(engagement)) AS engagement,
					toInt32(SUM(impression)) AS impression
				FROM (
					SELECT pin_id,
						argMax(engagement, inserted_at) as engagement,
						argMax(impression, inserted_at) as impression
					FROM contentstudiobackend.pinterest_pin_insights
					WHERE pin_id in (SELECT post_id FROM postIds)
					GROUP BY record_id, pin_id
				)
				GROUP BY pin_id
			)
			SELECT
				toInt32(count()) as total_posts,
				toInt32(sum(total_engagement)) as total_engagements,
				toInt32(sum(total_impressions)) as total_impression
			FROM (
				SELECT
					toInt32(SUM(pinterest_insights.engagement)) AS total_engagement,
					toInt32(SUM(pinterest_insights.impression)) AS total_impressions
				FROM pins
				LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
				GROUP BY pins.pin_id
			)`, currentFilter("created_at")))
	}

	if len(platformCTEs) == 0 {
		return &SummaryResult{}, nil
	}

	// Combine all platform CTEs with UNION ALL
	unionQuery := strings.Join(platformCTEs, " UNION ALL ")

	query := fmt.Sprintf(`
WITH %s
SELECT toInt32(sum(total_posts)) as total_posts,
	toInt32(sum(total_engagements)) as total_engagement,
	toInt32(sum(total_impression)) as total_impressions,
	if(total_impressions != 0, round(total_engagement / total_impressions, 2), 0) as total_engagement_rate_per_impression
FROM (%s)`, postIdCTE, unionQuery)

	var result SummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalPosts,
		&result.TotalEngagement,
		&result.TotalImpressions,
		&result.TotalEngagementRatePerImpression,
	)
	if err != nil {
		return nil, fmt.Errorf("GetSummary: %w", err)
	}
	return &result, nil
}

// buildPostIdPairsCTE builds the CTE for campaign/label → post_id mapping used by breakdown queries.
// Returns the WITH clause and a flat list of all post IDs for filtering.
func buildPostIdPairsCTE(campaignLabelObjects map[string][]string) (string, []string) {
	var pairs []string
	var allPostIDs []string
	for id, postIDs := range campaignLabelObjects {
		for _, postID := range postIDs {
			pairs = append(pairs, fmt.Sprintf("('%s','%s')",
				strings.ReplaceAll(id, "'", "\\'"),
				strings.ReplaceAll(postID, "'", "\\'"),
			))
			allPostIDs = append(allPostIDs, postID)
		}
	}
	if len(pairs) == 0 {
		pairs = append(pairs, "('','')")
	}

	cte := fmt.Sprintf(`WITH pairs AS (
		SELECT [%s] AS pairs_array
	),
	postIds AS (
		SELECT pair.1 AS id, pair.2 AS post_id
		FROM pairs
		ARRAY JOIN pairs_array AS pair
	)`, strings.Join(pairs, ","))

	return cte, allPostIDs
}

// GetBreakdownData executes the campaign/label breakdown query for the given period.
// Returns per-campaign/label aggregated totals for posts, engagement, and impressions.
func (r *Repository) GetBreakdownData(ctx context.Context, campaignLabelObjects map[string][]string, params *ch.QueryParams, period string) ([]BreakdownResult, error) {
	if len(campaignLabelObjects) == 0 {
		return []BreakdownResult{}, nil
	}

	pairsCTE, _ := buildPostIdPairsCTE(campaignLabelObjects)

	var startDate, endDate time.Time
	if period == "current" {
		startDate = params.DateFrom
		endDate = params.DateTo.Add(24 * time.Hour)
	} else {
		startDate = params.PrevDateFrom
		endDate = params.PrevDateTo.Add(24 * time.Hour)
	}

	pairsCTE += ", " + buildFacebookGroupedCTEs(
		fbDateFilterClause(startDate, endDate, "fp"),
		fbPageFilterClause(params.FacebookIDs, "fp"),
		fbPageFilterClause(params.FacebookIDs, "fvi"),
		fbDateFilterClause(startDate, endDate, "fvi"),
	)

	currentFilter := func(field string) string {
		return dateFilter(field, startDate, endDate)
	}

	query := fmt.Sprintf(`%s,
			pinterest_insights AS (
				SELECT pin_id,
					toInt32(SUM(engagement)) AS engagement,
				toInt32(SUM(impression)) AS impression
			FROM (
				SELECT pin_id,
					argMax(engagement, inserted_at) as engagement,
					argMax(impression, inserted_at) as impression
				FROM contentstudiobackend.pinterest_pin_insights
				WHERE pin_id IN (SELECT post_id FROM postIds)
				GROUP BY record_id, pin_id
			)
			GROUP BY pin_id
		)
			SELECT id, COALESCE('%s', '%s') as era,
				COALESCE(toInt32(count()), 0) as total_posts,
				COALESCE(toInt32(sum(total_engagement)), 0) as total_engagement,
				COALESCE(toInt32(sum(total_impressions)), 0) as total_impressions
			FROM (
				SELECT post_id,
					toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
					toFloat64(argMax(post_impressions, saving_time)) AS total_impressions
				FROM contentstudiobackend.facebook_posts
				WHERE post_id IN (SELECT post_id FROM facebook_post_ids) AND %s
				GROUP BY post_id
			UNION ALL
				SELECT media_id AS post_id,
				toFloat64(argMax(engagement, stored_event_at)) AS total_engagement,
				toFloat64(argMax(views, stored_event_at)) AS total_impressions
			FROM contentstudiobackend.instagram_posts
			WHERE media_id IN (SELECT post_id FROM postIds) AND %s
			GROUP BY media_id
		UNION ALL
			SELECT activity AS post_id,
				toFloat64(argMax(total_engagement, saving_time)) AS total_engagement,
				toFloat64(argMax(impressions, saving_time)) AS total_impressions
			FROM contentstudiobackend.linkedin_posts
			WHERE activity IN (SELECT post_id FROM postIds) AND %s
			GROUP BY activity
		UNION ALL
			SELECT post_id,
				toFloat64(argMax(engagement_count, inserted_at)) AS total_engagement,
				toFloat64(argMax(view_count, inserted_at)) AS total_impressions
			FROM contentstudiobackend.tiktok_posts
			WHERE post_id IN (SELECT post_id FROM postIds) AND %s
			GROUP BY post_id
		UNION ALL
			SELECT video_id AS post_id,
				toFloat64(argMax(likes, inserted_at) + argMax(comments, inserted_at) + argMax(shares, inserted_at) + argMax(dislikes, inserted_at)) AS total_engagement,
				toFloat64(argMax(views, inserted_at)) AS total_impressions
			FROM contentstudiobackend.youtube_videos
			WHERE video_id IN (SELECT post_id FROM postIds) AND %s
			GROUP BY video_id
		UNION ALL
			SELECT pins.pin_id as post_id,
				toFloat64(SUM(pinterest_insights.engagement)) AS total_engagement,
				toFloat64(SUM(pinterest_insights.impression)) AS total_impressions
			FROM (
				SELECT pin_id
				FROM contentstudiobackend.pinterest_pins
				WHERE pin_id IN (SELECT post_id FROM postIds) AND %s
					GROUP BY pin_id
				) AS pins
				LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
				GROUP BY pins.pin_id
			) AS all_posts
			LEFT JOIN matched_post_ids ON matched_post_ids.post_id = all_posts.post_id
			GROUP BY matched_post_ids.id`,
		pairsCTE, period, period,
		currentFilter("created_time"),
		currentFilter("post_created_at"),
		currentFilter("created_at"),
		currentFilter("created_at"),
		currentFilter("published_at"),
		currentFilter("created_at"),
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetBreakdownData: %w", err)
	}
	defer rows.Close()

	var results []BreakdownResult
	for rows.Next() {
		var row BreakdownResult
		if err := rows.Scan(&row.ID, &row.Era, &row.TotalPosts, &row.TotalEngagement, &row.TotalImpressions); err != nil {
			return nil, fmt.Errorf("GetBreakdownData scan: %w", err)
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetBreakdownData rows: %w", err)
	}
	return results, nil
}

// GetInsightsData executes the campaign/label time-series insights query.
// Returns daily aggregated engagement, impressions, and post counts per campaign/label.
func (r *Repository) GetInsightsData(ctx context.Context, campaignLabelObjects map[string][]string, params *ch.QueryParams) ([]InsightsResult, error) {
	if len(campaignLabelObjects) == 0 {
		return []InsightsResult{}, nil
	}

	pairsCTE, _ := buildPostIdPairsCTE(campaignLabelObjects)
	startDate := params.DateFrom
	endDate := params.DateTo.Add(24 * time.Hour)
	pairsCTE += ", " + buildFacebookGroupedCTEs(
		fbDateFilterClause(startDate, endDate, "fp"),
		fbPageFilterClause(params.FacebookIDs, "fp"),
		fbPageFilterClause(params.FacebookIDs, "fvi"),
		fbDateFilterClause(startDate, endDate, "fvi"),
	)
	currentFilter := func(field string) string {
		return dateFilter(field, startDate, endDate)
	}

	query := fmt.Sprintf(`%s,
			pinterest_insights AS (
				SELECT pin_id,
					toInt32(SUM(engagement)) AS engagement,
				toInt32(SUM(impression)) AS impression
			FROM (
				SELECT pin_id,
					argMax(engagement, inserted_at) as engagement,
					argMax(impression, inserted_at) as impression
				FROM contentstudiobackend.pinterest_pin_insights
				WHERE pin_id IN (SELECT post_id FROM postIds)
				GROUP BY record_id, pin_id
			)
			GROUP BY pin_id
		)
		SELECT id,
			groupArray(total_engagement) as total_engagement,
			groupArray(total_impressions) as total_impressions,
			groupArray(total_posts) as total_posts,
			groupArray(toString(created_at)) as created_at
			FROM (
				SELECT matched_post_ids.id as id,
					toInt32(sum(total_engagement)) as total_engagement,
					toInt32(sum(total_impressions)) as total_impressions,
					toInt32(count()) as total_posts,
					toDate(created_at) as created_at
				FROM (
					SELECT post_id,
						toFloat64(argMax(total_engagement, saving_time)) as total_engagement,
						toFloat64(argMax(post_impressions, saving_time)) as total_impressions,
						created_time as created_at
					FROM contentstudiobackend.facebook_posts
					WHERE post_id IN (SELECT post_id FROM facebook_post_ids) AND %s
					GROUP BY post_id, created_time
				UNION ALL
					SELECT media_id as post_id,
					toFloat64(argMax(engagement, stored_event_at)) as total_engagement,
					toFloat64(argMax(views, stored_event_at)) as total_impressions,
					post_created_at as created_at
				FROM contentstudiobackend.instagram_posts
				WHERE media_id IN (SELECT post_id FROM postIds) AND %s
				GROUP BY media_id, post_created_at
			UNION ALL
				SELECT activity as post_id,
					toFloat64(argMax(total_engagement, saving_time)) as total_engagement,
					toFloat64(argMax(impressions, saving_time)) as total_impressions,
					created_at
				FROM contentstudiobackend.linkedin_posts
				WHERE activity IN (SELECT post_id FROM postIds) AND %s
				GROUP BY activity, created_at
			UNION ALL
				SELECT post_id,
					toFloat64(argMax(engagement_count, inserted_at)) as total_engagement,
					toFloat64(argMax(view_count, inserted_at)) as total_impressions,
					created_at
				FROM contentstudiobackend.tiktok_posts
				WHERE post_id IN (SELECT post_id FROM postIds) AND %s
				GROUP BY post_id, created_at
			UNION ALL
				SELECT video_id as post_id,
					toFloat64(argMax(likes, inserted_at) + argMax(comments, inserted_at) + argMax(shares, inserted_at) + argMax(dislikes, inserted_at)) as total_engagement,
					toFloat64(argMax(views, inserted_at)) as total_impressions,
					published_at as created_at
				FROM contentstudiobackend.youtube_videos
				WHERE video_id IN (SELECT post_id FROM postIds) AND %s
				GROUP BY video_id, created_at
			UNION ALL
				SELECT pins.pin_id as post_id,
					toFloat64(SUM(pinterest_insights.engagement)) AS total_engagement,
					toFloat64(SUM(pinterest_insights.impression)) AS total_impressions,
					pins.created_at as created_at
				FROM (
					SELECT pin_id, created_at
					FROM contentstudiobackend.pinterest_pins
					WHERE pin_id IN (SELECT post_id FROM postIds) AND %s
						GROUP BY pin_id, created_at
					) AS pins
					LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
					GROUP BY pins.pin_id, pins.created_at
				) AS all_posts
				LEFT JOIN matched_post_ids ON all_posts.post_id = matched_post_ids.post_id
				GROUP BY created_at, matched_post_ids.id
				ORDER BY created_at DESC
			)
			GROUP BY id`,
		pairsCTE,
		currentFilter("created_time"),
		currentFilter("post_created_at"),
		currentFilter("created_at"),
		currentFilter("created_at"),
		currentFilter("published_at"),
		currentFilter("created_at"),
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetInsightsData: %w", err)
	}
	defer rows.Close()

	var results []InsightsResult
	for rows.Next() {
		var row InsightsResult
		if err := rows.Scan(&row.ID, &row.TotalEngagement, &row.TotalImpressions, &row.TotalPosts, &row.CreatedAt); err != nil {
			return nil, fmt.Errorf("GetInsightsData scan: %w", err)
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetInsightsData rows: %w", err)
	}
	return results, nil
}

// GetPlannerAnalytics executes the per-post planner analytics query for a single platform.
// Returns detailed engagement metrics for the specified post IDs.
func (r *Repository) GetPlannerAnalytics(ctx context.Context, postIDs []string, platform string) (map[string]interface{}, error) {
	if len(postIDs) == 0 {
		return map[string]interface{}{}, nil
	}

	formattedIDs := formatPostIDs(postIDs)
	postIdCTE := fmt.Sprintf("postIds AS (SELECT arrayJoin(%s) AS post_id)", formattedIDs)

	var query string
	platformLower := strings.ToLower(platform)

	switch platformLower {
	case "facebook":
		query = fmt.Sprintf(`WITH %s, %s
				SELECT
					last_value(total_engagement) as engagement,
				'Total interactions with your post i.e likes, comments and shares' as engagement_tooltip,
				if(last_value(post_impressions) > 0, round(last_value(total_engagement) / last_value(post_impressions), 2), 0) as engagement_rate,
				'Engagement received over time for this post.' as engagement_rate_tooltip,
				last_value(post_impressions) as impressions,
				'Times your post was displayed to users' as impressions_tooltip,
				last_value(post_impressions_unique) as reach,
				'Times your post was displayed to unique users' as reach_tooltip,
				last_value(comments) as comments,
				'Number of comments on your post' as comments_tooltip,
				last_value(shares) as repost,
				'Number of times your post was shared' as repost_tooltip,
				last_value(post_clicks) as post_clicks,
				'Number of clicks on your post' as post_clicks_tooltip,
				last_value(total) as reactions,
				'Total reactions on your post' as reactions_tooltip,
				last_value(like) as likes,
				'Number of times your post was liked' as likes_tooltip,
				last_value(love) as love,
				'Number of love reactions received on your post' as love_tooltip,
				last_value(wow) as wow,
				'Number of wow reactions received on your post' as wow_tooltip,
				last_value(haha) as haha,
				'Number of haha reactions received on your post' as haha_tooltip,
				last_value(sad) as sad,
				'Number of sad reactions received on your post' as sad_tooltip,
				last_value(angry) as anger,
				'Number of anger reactions received on your post' as anger_tooltip,
					concat(upper(substring(last_value(media_type), 1, 1)), lower(substring(last_value(media_type), 2))) as media_type,
					'Type of media used in your post' as media_type_tooltip
				FROM contentstudiobackend.facebook_posts
				WHERE post_id IN (SELECT post_id FROM facebook_post_ids)
				GROUP BY post_id`, postIdCTE, buildFacebookSummaryCTEs("", "", "", ""))

	case "instagram":
		query = fmt.Sprintf(`WITH %s,
			instagram_agg AS (
				SELECT
					media_id,
					last_value(engagement) as engagement,
					last_value(impressions) as impressions,
					last_value(reach) as reach,
					last_value(like_count) as likes,
					last_value(comments_count) as comments,
					last_value(saved) as saves,
					last_value(media_type) as media_type
				FROM contentstudiobackend.instagram_posts
				WHERE media_id IN (SELECT post_id FROM postIds)
				GROUP BY media_id
			)
			SELECT
				engagement,
				'Total interactions with your post i.e likes, comments and saves' as engagement_tooltip,
				if(impressions > 0, round(engagement / impressions, 2), 0) as engagement_rate,
				'Engagement received over time for this post.' as engagement_rate_tooltip,
				impressions,
				'Total number of times your post was seen' as impressions_tooltip,
				reach,
				'Total number of unique users who saw your post' as reach_tooltip,
				likes,
				'Total number of likes on your post' as likes_tooltip,
				comments,
				'Total number of comments on your post' as comments_tooltip,
				saves,
				'Total number of saves on your post' as saves_tooltip,
				CASE
					WHEN media_type = 'CAROUSEL_ALBUM' THEN 'Carousel'
					ELSE CONCAT(UPPER(SUBSTRING(media_type, 1, 1)), LOWER(SUBSTRING(media_type, 2)))
				END as media_type,
				'Type of media used in your post' as media_type_tooltip
			FROM instagram_agg`, postIdCTE)

	case "linkedin":
		query = fmt.Sprintf(`WITH %s,
			linkedin_agg AS (
				SELECT
					activity,
					last_value(total_engagement) as total_engagement,
					last_value(impressions) as impressions,
					last_value(reach) as reach,
					last_value(comments) as comments,
					last_value(favorites) as reactions,
					last_value(repost) as reposts,
					last_value(post_clicks) as post_clicks,
					last_value(media_type) as media_type
				FROM contentstudiobackend.linkedin_posts
				WHERE activity IN (SELECT post_id FROM postIds)
				GROUP BY activity
			)
			SELECT
				total_engagement as engagement,
				'Total interactions with your post i.e likes, comments and shares' as engagement_tooltip,
				if(impressions > 0, round(total_engagement / impressions, 2), 0) as engagement_rate,
				'Engagement received over time for this post.' as engagement_rate_tooltip,
				impressions,
				'Total number of times your post was seen' as impressions_tooltip,
				reach,
				'Total number of unique users who saw your post' as reach_tooltip,
				comments,
				'Total number of comments on your post' as comments_tooltip,
				reactions,
				'Total number of reactions on your post' as reactions_tooltip,
				reposts,
				'Total number of shares on your post' as reposts_tooltip,
				post_clicks,
				'Total number of clicks on your post' as post_clicks_tooltip,
				concat(upper(substring(media_type, 1, 1)), lower(substring(media_type, 2))) as media_type,
				'Type of media used in your post' as media_type_tooltip
			FROM linkedin_agg`, postIdCTE)

	case "tiktok":
		query = fmt.Sprintf(`WITH %s
			SELECT
				last_value(engagement_count) as engagement,
				'Total interactions with your post i.e likes, comments and shares' as engagement_tooltip,
				last_value(view_count) as views,
				'Number of times your post was viewed' as views_tooltip,
				last_value(like_count) as likes,
				'Total number of reactions on your post' as likes_tooltip,
				last_value(comments_count) as comments,
				'Total number of comments on your post' as comments_tooltip,
				last_value(share_count) as shares,
				'Total number of shares on your post' as shares_tooltip,
				round(last_value(engagement_rate), 2) as engagement_rate,
				'Engagement received over time for this post.' as engagement_rate_tooltip
			FROM contentstudiobackend.tiktok_posts
			WHERE post_id IN (SELECT post_id FROM postIds)
			GROUP BY post_id`, postIdCTE)

	case "youtube":
		query = fmt.Sprintf(`WITH %s,
			youtube_agg AS (
				SELECT
					video_id,
					last_value(likes) as likes,
					last_value(dislikes) as dislikes,
					last_value(comments) as comments,
					last_value(shares) as shares,
					last_value(impressions) as impressions,
					last_value(views) as views,
					last_value(red_views) as red_views,
					last_value(subscribers_gained) as subscribers_gained,
					last_value(minutes_watched) as minutes_watched,
					last_value(red_minutes_watched) as red_minutes_watched,
					last_value(average_view_duration) as average_view_duration,
					last_value(media_type) as media_type
				FROM contentstudiobackend.youtube_videos
				WHERE video_id IN (SELECT post_id FROM postIds)
				GROUP BY video_id
			)
			SELECT
				likes + comments + shares + dislikes as engagement,
				'Total interactions with your post i.e likes, dislikes, comments and shares' as engagement_tooltip,
				if(impressions > 0, round((likes + comments + shares + dislikes) / impressions, 2), 0) as engagement_rate,
				'Engagement received over time for this post.' as engagement_rate_tooltip,
				views,
				'Total views of your post' as views_tooltip,
				likes,
				'Total likes of your post' as likes_tooltip,
				dislikes,
				'Total dislikes of your post' as dislikes_tooltip,
				comments,
				'Total comments of your post' as comments_tooltip,
				shares,
				'Total shares of your post' as shares_tooltip,
				red_views,
				'Total views by subscribers on your post' as red_views_tooltip,
				subscribers_gained,
				'Number of people who have clicked on subscribe while watching this video' as subscribers_gained_tooltip,
				minutes_watched,
				'Minutes for which this video was watched by all users' as minutes_watched_tooltip,
				red_minutes_watched,
				'Minutes for which this video was watched by subscribers' as red_minutes_watched_tooltip,
				average_view_duration,
				'Average view duration for which this video got played' as average_view_duration_tooltip,
				concat(upper(substring(media_type, 1, 1)), lower(substring(media_type, 2))) as media_type,
				'Type of media used in your post' as media_type_tooltip
			FROM youtube_agg`, postIdCTE)

	case "pinterest":
		query = fmt.Sprintf(`WITH %s,
			pins AS (
				SELECT pin_id
				FROM contentstudiobackend.pinterest_pins
				WHERE pin_id IN (SELECT post_id FROM postIds)
				GROUP BY pin_id
			),
			pinterest_insights AS (
				SELECT pin_id,
					last_value(engagement) AS engagement,
					last_value(impression) AS impression,
					last_value(pin_clicks) AS pin_clicks,
					last_value(outbound_click) AS outbound_click,
					last_value(saves) AS saves,
					last_value(engagement_rate) AS engagement_rate
				FROM (
					SELECT pin_id,
						last_value(engagement) as engagement,
						last_value(impression) as impression,
						last_value(pin_clicks) as pin_clicks,
						last_value(outbound_click) as outbound_click,
						last_value(saves) as saves,
						last_value(engagement_rate) as engagement_rate
					FROM contentstudiobackend.pinterest_pin_insights
					WHERE pin_id in (SELECT post_id FROM postIds)
					GROUP BY record_id, pin_id
				)
				GROUP BY pin_id
			)
			SELECT
				toInt32(last_value(pinterest_insights.engagement)) AS engagement,
				'Total interactions with your post i.e pin clicks, outbound clicks and saves' as engagement_tooltip,
				toInt32(last_value(pinterest_insights.impression)) AS impressions,
				'Total number of times your post was seen' as impressions_tooltip,
				toInt32(last_value(pinterest_insights.pin_clicks)) AS pin_clicks,
				'Total number of times your post was clicked to enlarge it on your screen' as pin_clicks_tooltip,
				toInt32(last_value(pinterest_insights.outbound_click)) AS outbound_clicks,
				'Total number of times your post was clicked to follow the URL linked to it' as outbound_clicks_tooltip,
				toInt32(last_value(pinterest_insights.saves)) AS saves,
				'Total number of times your post was saved' as saves_tooltip,
				round(last_value(pinterest_insights.engagement_rate), 2) as engagement_rate,
				'Engagement received over time for this post.' as engagement_rate_tooltip
			FROM pins
			LEFT JOIN pinterest_insights ON pins.pin_id = pinterest_insights.pin_id
			GROUP BY pins.pin_id`, postIdCTE)

	default:
		return map[string]interface{}{}, nil
	}

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetPlannerAnalytics: %w", err)
	}
	defer rows.Close()

	cols := rows.ColumnTypes()
	colNames := make([]string, len(cols))
	scanDests := make([]interface{}, len(cols))
	for i, c := range cols {
		colNames[i] = c.Name()
		scanType := c.ScanType()
		if scanType == nil {
			var v interface{}
			scanDests[i] = &v
			continue
		}
		scanDests[i] = reflect.New(scanType).Interface()
	}

	result := make(map[string]interface{})
	if rows.Next() {
		if err := rows.Scan(scanDests...); err != nil {
			return nil, fmt.Errorf("GetPlannerAnalytics scan: %w", err)
		}
		for i, name := range colNames {
			v := reflect.ValueOf(scanDests[i])
			if v.Kind() == reflect.Ptr && !v.IsNil() {
				result[name] = v.Elem().Interface()
				continue
			}
			result[name] = scanDests[i]
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPlannerAnalytics rows: %w", err)
	}

	return result, nil
}
