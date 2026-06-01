package linkedin

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// Repository provides LinkedIn analytics queries against ClickHouse.
// All query patterns are migrated from the PHP LinkedInAnalyticsBuilder
// (contentstudio-backend: app/Http/AnalyticsV2/Builders/LinkedInAnalyticsBuilder.php).
type Repository struct {
	client *ch.Client
}

// NewRepository creates a new LinkedIn analytics repository.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// GetSummary returns aggregated post and page-level metrics for a date range.
// Joins linkedin_posts (post engagement) with linkedin_insights (page metrics) using a CTE for post deduplication.
// Mirrors PHP: LinkedInAnalyticsBuilder::getSummaryData()
func (r *Repository) GetSummary(ctx context.Context, params *ch.QueryParams) (*SummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	postDateFilter := ch.DateFilter("published_at", params)
	insightsDateFilter := ch.DateFilter("created_at", params)
	createdAtFilter := insightsDateFilter
	partFilter := ch.PartitionMonthFilter("published_at", params)
	postEngRate := ch.SafeRate("sum(total_engagement * impressions)", "sum(impressions)", 4)
	insightsEngRate := ch.SafeRate("sum(daily_engagement_impressions)", "sum(daily_impressions)", 4)

	query := fmt.Sprintf(`%s
SELECT
    toInt32(coalesce(post.post_comments, 0))        AS post_comments,
    toInt32(coalesce(post.post_favorites, 0))        AS post_likes,
    toInt32(coalesce(post.total_engagement, 0))      AS total_engagement,
    toInt32(coalesce(post.total_posts, 0))           AS total_posts,
    toInt32(coalesce(post.post_shares, 0))           AS post_shares,
    toInt32(coalesce(post.post_clicks, 0))           AS post_clicks,
    toInt32(coalesce(insights.total_follower_count, 0)) AS followers,
    toInt32(coalesce(insights.page_views, 0))           AS page_views,
    toInt32(coalesce(insights.page_reach, 0))           AS page_reach,
    toInt32(coalesce(insights.page_shares, 0))          AS page_shares,
    toInt32(coalesce(insights.page_comments, 0))        AS page_comments,
    toInt32(coalesce(insights.page_reactions, 0))       AS page_reactions,
    toInt32(coalesce(insights.total_impressions, 0))    AS page_impressions,
    toInt32(coalesce(insights.unique_visitors, 0))      AS page_unique_visitors,
    round(coalesce(insights.engagement_rate, 0), 2)     AS engagement_rate,
    round(coalesce(post.post_engagement_rate, 0), 2)    AS post_engagement_rate
FROM
(
    SELECT toString(c1) AS linkedin_id
    FROM VALUES %s
) AS page_ids
LEFT JOIN
(
    SELECT
        linkedin_id,
        post_comments, post_favorites, post_clicks, post_shares,
        post_comments + post_favorites + post_shares AS total_engagement,
        post_engagement_rate, total_posts
    FROM (
        SELECT
            linkedin_id,
            SUM(comments) AS post_comments, SUM(favorites) AS post_favorites,
            SUM(post_clicks) AS post_clicks, SUM(repost) AS post_shares,
            %s AS post_engagement_rate, COUNT(*) AS total_posts
        FROM linkedin_posts
        WHERE linkedin_id IN %s AND %s AND (post_id, saving_time) IN posts
        GROUP BY linkedin_id
    )
) AS post USING linkedin_id
LEFT JOIN
(
    SELECT
        linkedin_id AS platform_id,
        argMax(latest_follower_count, latest_date) AS total_follower_count,
        sum(daily_page_views) AS page_views, sum(daily_reach) AS page_reach,
        sum(daily_repost) AS page_shares, sum(daily_comments) AS page_comments,
        sum(daily_reactions) AS page_reactions, sum(daily_impressions) AS total_impressions,
        %s AS engagement_rate, sum(daily_unique_visitors) AS unique_visitors
    FROM (
        SELECT
            linkedin_id, toDate(created_at) AS latest_date,
            argMax(totalFollowerCount, inserted_at) AS latest_follower_count,
            argMax(page_views, inserted_at) AS daily_page_views,
            argMax(reach, inserted_at) AS daily_reach,
            argMax(repost, inserted_at) AS daily_repost,
            argMax(comments, inserted_at) AS daily_comments,
            argMax(reactions, inserted_at) AS daily_reactions,
            argMax(impressionCount, inserted_at) AS daily_impressions,
            argMax(engagement, inserted_at) * argMax(impressionCount, inserted_at) AS daily_engagement_impressions,
            argMax(unique_visitors, inserted_at) AS daily_unique_visitors
        FROM linkedin_insights
        WHERE linkedin_id IN %s AND %s
        GROUP BY record_id, linkedin_id, toDate(created_at)
    )
    GROUP BY linkedin_id
) AS insights ON page_ids.linkedin_id = insights.platform_id
`, ch.PostDedupCTE(ids, postDateFilter, partFilter), ids, postEngRate, ids, createdAtFilter, insightsEngRate, ids, insightsDateFilter)

	var result SummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PostComments, &result.PostLikes, &result.TotalEngagement,
		&result.TotalPosts, &result.PostShares, &result.PostClicks,
		&result.Followers, &result.PageViews, &result.PageReach,
		&result.PageShares, &result.PageComments, &result.PageReactions,
		&result.PageImpressions, &result.PageUniqueVisitors,
		&result.EngagementRate, &result.PostEngagementRate,
	)
	if err != nil {
		return nil, fmt.Errorf("GetSummary: %w", err)
	}
	return &result, nil
}

// GetPostsSummary returns aggregated post-level metrics from linkedin_posts for the given period.
// Run concurrently with GetInsightsSummary for maximum throughput.
func (r *Repository) GetPostsSummary(ctx context.Context, params *ch.QueryParams) (*PostsSummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("published_at", params)
	createdAtFilter := ch.DateFilter("created_at", params)
	partFilter := ch.PartitionMonthFilter("published_at", params)
	postEngRate := ch.SafeRate("sum(total_engagement * impressions)", "sum(impressions)", 4)

	query := fmt.Sprintf(`%s
SELECT
    toInt32(SUM(comments))                                  AS post_comments,
    toInt32(SUM(favorites))                                 AS post_likes,
    toInt32(SUM(comments) + SUM(favorites) + SUM(repost))   AS total_engagement_count,
    toInt32(COUNT(*))                                       AS total_posts,
    toInt32(SUM(repost))                                    AS post_shares,
    toInt32(SUM(post_clicks))                               AS post_clicks,
    %s                                                      AS post_engagement_rate
FROM linkedin_posts
WHERE linkedin_id IN %s AND %s AND (post_id, saving_time) IN posts
`, ch.PostDedupCTE(ids, dateFilter, partFilter), postEngRate, ids, createdAtFilter)

	var result PostsSummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PostComments, &result.PostLikes, &result.TotalEngagement,
		&result.TotalPosts, &result.PostShares, &result.PostClicks,
		&result.PostEngagementRate,
	)
	if err != nil {
		return nil, fmt.Errorf("GetPostsSummary: %w", err)
	}
	return &result, nil
}

// GetInsightsSummary returns aggregated page-level metrics from linkedin_insights for the given period.
// Run concurrently with GetPostsSummary for maximum throughput.
func (r *Repository) GetInsightsSummary(ctx context.Context, params *ch.QueryParams) (*InsightsSummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)
	engRate := ch.SafeRate("sum(daily_engagement_impressions)", "sum(daily_impressions)", 4)

	query := fmt.Sprintf(`
SELECT
    toInt32(any(flw.followers))                         AS followers,
    toInt32(SUM(daily_page_views))                      AS page_views,
    toInt32(SUM(daily_reach))                           AS page_reach,
    toInt32(SUM(daily_repost))                          AS page_shares,
    toInt32(SUM(daily_comments))                        AS page_comments,
    toInt32(SUM(daily_reactions))                       AS page_reactions,
    toInt32(SUM(daily_impressions))                     AS page_impressions,
    toInt32(SUM(daily_unique_visitors))                 AS page_unique_visitors,
    %s                                                  AS engagement_rate
FROM (
    SELECT
        argMax(page_views, inserted_at)                                                 AS daily_page_views,
        argMax(reach, inserted_at)                                                      AS daily_reach,
        argMax(repost, inserted_at)                                                     AS daily_repost,
        argMax(comments, inserted_at)                                                   AS daily_comments,
        argMax(reactions, inserted_at)                                                  AS daily_reactions,
        argMax(impressionCount, inserted_at)                                            AS daily_impressions,
        argMax(engagement, inserted_at) * argMax(impressionCount, inserted_at)         AS daily_engagement_impressions,
        argMax(unique_visitors, inserted_at)                                            AS daily_unique_visitors
    FROM linkedin_insights
    WHERE linkedin_id IN %s AND %s
    GROUP BY record_id, toDate(created_at)
) AS di
CROSS JOIN (
    -- Use argMin(inserted_at) per date (no record_id grouping) to pick the original daily
    -- snapshot and ignore backfilled rows, consistent with GetAudienceRollup.
    SELECT toInt32(argMax(fc, d)) AS followers
    FROM (
        SELECT
            argMin(totalFollowerCount, inserted_at) AS fc,
            toDate(created_at)                      AS d
        FROM linkedin_insights
        WHERE linkedin_id IN %s AND %s
        GROUP BY d
    )
) AS flw
`, engRate, ids, dateFilter, ids, dateFilter)

	var result InsightsSummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Followers, &result.PageViews, &result.PageReach,
		&result.PageShares, &result.PageComments, &result.PageReactions,
		&result.PageImpressions, &result.PageUniqueVisitors,
		&result.EngagementRate,
	)
	if err != nil {
		return nil, fmt.Errorf("GetInsightsSummary: %w", err)
	}
	return &result, nil
}

// GetAudienceGrowth returns time-series follower data (organic, paid, total) with daily deltas.
// Uses arrayFill to forward-fill zero gaps, arrayDifference for daily change, and WITH FILL for date continuity.
// Mirrors PHP: LinkedInAnalyticsBuilder::getAudienceGrowthData()
func (r *Repository) GetAudienceGrowth(ctx context.Context, params *ch.QueryParams) (*AudienceResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
SELECT
    notEmpty(total_follower_count_temp) AS show_data,
    arrayFill(x -> not x == 0, organic_follower_count_temp) AS organic_follower_count,
    arrayMap(x -> toInt32(x),
        arrayPushFront(arrayDifference(arrayFill(x -> not x == 0, organic_follower_count_temp)), toInt32(0))
    ) AS organic_followers_daily,
    arrayFill(x -> not x == 0, paid_follower_count_temp) AS paid_follower_count,
    arrayMap(x -> toInt32(x),
        arrayPushFront(arrayDifference(arrayFill(x -> not x == 0, paid_follower_count_temp)), toInt32(0))
    ) AS paid_followers_daily,
    arrayFill(x -> not x == 0, total_follower_count_temp) AS total_follower_count,
    arrayMap(x -> toInt32(x),
        arrayPushFront(arrayDifference(arrayFill(x -> not x == 0, total_follower_count_temp)), toInt32(0))
    ) AS total_followers_daily,
    buckets
FROM (
    SELECT
        groupArray(dates) AS buckets,
        groupArray(organicFollowerCount) AS organic_follower_count_temp,
        groupArray(paidFollowerCount) AS paid_follower_count_temp,
        groupArray(totalFollowerCount) AS total_follower_count_temp
    FROM (
        -- For each date, pick the earliest-inserted row's follower counts (original daily snapshot).
        -- argMin(inserted_at) per date deterministically ignores later backfills, consistent
        -- with the GetAudienceRollup deduplication approach.
        SELECT
            toDate(created_at)                                 AS dates,
            toInt32(argMin(organicFollowerCount, inserted_at)) AS organicFollowerCount,
            toInt32(argMin(paidFollowerCount,    inserted_at)) AS paidFollowerCount,
            toInt32(argMin(totalFollowerCount,   inserted_at)) AS totalFollowerCount
        FROM linkedin_insights
        WHERE linkedin_id IN %s AND %s
        GROUP BY dates
        ORDER BY dates ASC
        %s
    )
)
`, ids, dateFilter, fill)

	var result AudienceResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.OrganicFollowerCount, &result.OrganicFollowersDaily,
		&result.PaidFollowerCount, &result.PaidFollowersDaily,
		&result.TotalFollowerCount, &result.TotalFollowersDaily,
		&result.Buckets,
	)
	if err != nil {
		return nil, fmt.Errorf("GetAudienceGrowth: %w", err)
	}
	return &result, nil
}

// GetLastFollowerCounts returns the most recent non-zero follower counts from linkedin_insights.
// Used as a fallback when the current period starts with zero follower data.
func (r *Repository) GetLastFollowerCounts(ctx context.Context, params *ch.QueryParams) (*FollowerCounts, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    arrayFirst(x -> x != 0, groupArray(totalFollowerCount)) AS total_follower_count,
    arrayFirst(x -> x != 0, groupArray(organicFollowerCount)) AS organic_follower_count,
    arrayFirst(x -> x != 0, groupArray(paidFollowerCount)) AS paid_follower_count
FROM
(
    SELECT
        argMin(created_at, inserted_at) AS inserted_time,
        toInt32(argMin(totalFollowerCount, inserted_at)) AS totalFollowerCount,
        toInt32(argMin(organicFollowerCount, inserted_at)) AS organicFollowerCount,
        toInt32(argMin(paidFollowerCount, inserted_at)) AS paidFollowerCount
    FROM linkedin_insights
    WHERE linkedin_id IN %s AND %s
    GROUP BY record_id
    ORDER BY inserted_time DESC
)
`, ids, dateFilter)

	var result FollowerCounts
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalFollowerCount,
		&result.OrganicFollowerCount,
		&result.PaidFollowerCount,
	)
	if err != nil {
		return nil, fmt.Errorf("GetLastFollowerCounts: %w", err)
	}
	return &result, nil
}

// GetAudienceRollup returns aggregated follower counts and averages for a date range.
// Used for current vs previous period comparison in the audience growth widget.
func (r *Repository) GetAudienceRollup(ctx context.Context, params *ch.QueryParams) (*AudienceRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    toInt32(last_value(organicFollowerCount)) AS organic_follower_count,
    toInt32(last_value(paidFollowerCount)) AS paid_follower_count,
    toInt32(last_value(totalFollowerCount)) AS total_follower_count,
    ifNotFinite(round(AVG(totalFollowerCount), 2), 0) AS avg_follower_count
FROM
(
    SELECT
        argMin(organicFollowerCount, inserted_at) AS organicFollowerCount,
        argMin(paidFollowerCount, inserted_at) AS paidFollowerCount,
        argMin(totalFollowerCount, inserted_at) AS totalFollowerCount,
        toDate(created_at) AS dates
    FROM linkedin_insights
    WHERE linkedin_id IN %s AND %s
    GROUP BY dates
    ORDER BY dates ASC
)
`, ids, dateFilter)

	var result AudienceRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.OrganicFollowerCount,
		&result.PaidFollowerCount,
		&result.TotalFollowerCount,
		&result.AvgFollowerCount,
	)
	if err != nil {
		return nil, fmt.Errorf("GetAudienceRollup: %w", err)
	}
	return &result, nil
}

// GetPageViews returns time-series page view data split by desktop and mobile.
// Uses arrayCumSum for cumulative totals and WITH FILL for date gap filling.
// Mirrors PHP: LinkedInAnalyticsBuilder::getPageViewsData()
func (r *Repository) GetPageViews(ctx context.Context, params *ch.QueryParams) (*PageViewsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
SELECT
    arrayMap(x -> toInt32(x), arrayCumSum(desktop_views_daily)) AS desktop_page_views,
    arrayMap(x -> toInt32(x), arrayCumSum(mobile_views_daily)) AS mobile_page_views,
    arrayMap(x -> toInt32(x), arrayCumSum(total_views_daily)) AS total_page_views,
    arrayMap(x -> toInt32(x), desktop_views_daily) AS desktop_page_views_daily,
    arrayMap(x -> toInt32(x), mobile_views_daily) AS mobile_page_views_daily,
    arrayMap(x -> toInt32(x), total_views_daily) AS total_page_views_daily,
    toInt32(arraySum(total_views_daily)) AS show_data,
    buckets
FROM (
    SELECT
        groupArray(dates) AS buckets,
        groupArray(desktop_views) AS desktop_views_daily,
        groupArray(mobile_views) AS mobile_views_daily,
        groupArray(total_views) AS total_views_daily
    FROM (
        SELECT
            toDate(created_at) AS dates,
            toInt32(SUM(desktop_page_views)) AS desktop_views,
            toInt32(SUM(mobile_page_views)) AS mobile_views,
            toInt32(SUM(page_views)) AS total_views
        FROM linkedin_insights
        WHERE linkedin_id IN %s AND %s
        GROUP BY dates
        ORDER BY dates ASC
        %s
    )
)
`, ids, dateFilter, fill)

	var result PageViewsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.DesktopPageViews, &result.MobilePageViews, &result.TotalPageViews,
		&result.DesktopPageViewsDaily, &result.MobilePageViewsDaily, &result.TotalPageViewsDaily,
		&result.ShowData, &result.Buckets,
	)
	if err != nil {
		return nil, fmt.Errorf("GetPageViews: %w", err)
	}
	return &result, nil
}

// GetPageViewsRollup returns aggregated page view totals and averages for a date range.
// Used for current vs previous period comparison in the page views widget.
func (r *Repository) GetPageViewsRollup(ctx context.Context, params *ch.QueryParams) (*PageViewsRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    toInt32(SUM(total_views)) AS total_page_views,
    toInt32(SUM(desktop_views)) AS desktop_page_views,
    toInt32(SUM(mobile_views)) AS mobile_page_views,
    ifNotFinite(round(AVG(total_views), 2), 0) AS avg_page_views
FROM
(
    SELECT
        toInt32(SUM(page_views)) AS total_views,
        toInt32(SUM(desktop_page_views)) AS desktop_views,
        toInt32(SUM(mobile_page_views)) AS mobile_views
    FROM linkedin_insights
    WHERE linkedin_id IN %s AND %s
    GROUP BY toDate(created_at)
    ORDER BY toDate(created_at) ASC
)
`, ids, dateFilter)

	var result PageViewsRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalPageViews, &result.DesktopPageViews,
		&result.MobilePageViews, &result.AvgPageViews,
	)
	if err != nil {
		return nil, fmt.Errorf("GetPageViewsRollup: %w", err)
	}
	return &result, nil
}

// GetPublishingBehaviour returns time-series engagement metrics per day filtered by media types.
// Uses the post deduplication CTE with last_value() for per-post metric snapshots.
// Mirrors PHP: LinkedInAnalyticsBuilder::getPublishingBehaviourData()
func (r *Repository) GetPublishingBehaviour(ctx context.Context, params *ch.QueryParams, mediaTypes []string) (*PublishingResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("published_at", params)
	createdAtFilter := ch.DateFilter("created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))
	partFilter := ch.PartitionMonthFilter("published_at", params)
	mediaFilter := fmt.Sprintf("media_type IN (%s)", ch.FormatStringList(mediaTypes))

	query := fmt.Sprintf(`%s
SELECT
    groupArray(likes) AS likes, groupArray(comments) AS comments,
    groupArray(shares) AS shares, groupArray(clicks) AS clicks,
    groupArray(engagement_rate) AS engagement_rate,
    groupArray(impression) AS impressions,
    groupArray(total_posts) AS total_posts,
    groupArray(engagements) AS engagement,
    groupArray(reach) AS reach, groupArray(created_at) AS buckets
FROM (
    SELECT
        toInt32(SUM(like_count)) AS likes, toInt32(SUM(comments_count)) AS comments,
        toInt32(SUM(shares_count)) AS shares, toInt32(SUM(clicks_count)) AS clicks,
        toInt32(SUM(impressions)) AS impression, toInt32(COUNT(*)) AS total_posts,
        toInt32(likes + comments + shares) AS engagements,
        if(SUM(impressions) > 0, toFloat32(round(100 * engagements / impression, 2)), 0) AS engagement_rate,
        toInt32(SUM(reach)) AS reach, toDate(created_at) AS created_at
    FROM (
        SELECT post_id,
            last_value(favorites) AS like_count, last_value(comments) AS comments_count,
            last_value(repost) AS shares_count, last_value(post_clicks) AS clicks_count,
            last_value(impressions) AS impressions, last_value(reach) AS reach, created_at
        FROM linkedin_posts
        WHERE linkedin_id IN %s AND %s AND (post_id, saving_time) IN posts
        GROUP BY post_id, created_at
    )
    GROUP BY created_at
    ORDER BY created_at ASC
    %s
)
`, ch.PostDedupCTE(ids, dateFilter, partFilter, mediaFilter), ids, createdAtFilter, fill)

	var result PublishingResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Likes, &result.Comments, &result.Shares, &result.Clicks,
		&result.EngagementRate, &result.Impressions, &result.TotalPosts,
		&result.Engagement, &result.Reach, &result.Buckets,
	)
	if err != nil {
		return nil, fmt.Errorf("GetPublishingBehaviour: %w", err)
	}
	return &result, nil
}

// GetPublishingBehaviourRollup returns engagement metrics grouped by media type (text, images, videos, link, carousel)
// plus a "total" row. Uses UNION ALL to combine per-type and total rows.
func (r *Repository) GetPublishingBehaviourRollup(ctx context.Context, params *ch.QueryParams) ([]PublishingRollupRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("published_at", params)
	createdAtFilter := ch.DateFilter("created_at", params)
	partFilter := ch.PartitionMonthFilter("published_at", params)

	query := fmt.Sprintf(`%s,
media_types AS (
    SELECT arrayJoin(['text', 'images', 'videos', 'link', 'carousel']) AS media_type
),
metrics AS (
    SELECT
        mt.media_type,
        toInt32(COALESCE(total_posts, 0)) AS total_posts,
        toInt32(COALESCE(likes, 0)) AS likes,
        toInt32(COALESCE(comments, 0)) AS comments,
        toInt32(COALESCE(shares, 0)) AS shares,
        toInt32(COALESCE(clicks, 0)) AS clicks,
        toInt32(COALESCE(engagements, 0)) AS engagements,
        toInt32(COALESCE(impressions, 0)) AS impressions,
        toInt32(COALESCE(reach, 0)) AS reach
    FROM media_types mt
    LEFT JOIN
    (
        SELECT
            media_type,
            toInt32(COUNT(*)) AS total_posts,
            toInt32(SUM(like_count)) AS likes,
            toInt32(SUM(comments_count)) AS comments,
            toInt32(SUM(shares_count)) AS shares,
            toInt32(SUM(clicks_count)) AS clicks,
            toInt32(likes + comments + shares) AS engagements,
            toInt32(SUM(impressions)) AS impressions,
            toInt32(SUM(reach)) AS reach
        FROM
        (
            SELECT
                post_id,
                last_value(media_type) AS media_type,
                last_value(favorites) AS like_count,
                last_value(comments) AS comments_count,
                last_value(repost) AS shares_count,
                last_value(post_clicks) AS clicks_count,
                last_value(total_engagement) AS engagement,
                last_value(impressions) AS impressions,
                last_value(reach) AS reach,
                created_at
            FROM linkedin_posts
            WHERE linkedin_id IN %s AND %s AND (post_id, saving_time) IN posts
            GROUP BY post_id, created_at
        )
        GROUP BY media_type
    ) t ON mt.media_type = t.media_type
)
SELECT * FROM (
    SELECT * FROM metrics
    UNION ALL
    SELECT
        'total' AS media_type,
        toInt32(SUM(total_posts)) AS total_posts,
        toInt32(SUM(likes)) AS likes,
        toInt32(SUM(comments)) AS comments,
        toInt32(SUM(shares)) AS shares,
        toInt32(SUM(clicks)) AS clicks,
        toInt32(SUM(engagements)) AS engagements,
        toInt32(SUM(impressions)) AS impressions,
        toInt32(SUM(reach)) AS reach
    FROM metrics
)
ORDER BY CASE WHEN media_type = 'total' THEN 1 ELSE 0 END, media_type
`, ch.PostDedupCTE(ids, dateFilter, partFilter, "media_type != ''"), ids, createdAtFilter)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetPublishingBehaviourRollup: %w", err)
	}
	defer rows.Close()

	var results []PublishingRollupRow
	for rows.Next() {
		var row PublishingRollupRow
		if err := rows.Scan(
			&row.MediaType, &row.TotalPosts, &row.Likes, &row.Comments,
			&row.Shares, &row.Clicks, &row.Engagements, &row.Impressions, &row.Reach,
		); err != nil {
			return nil, fmt.Errorf("GetPublishingBehaviourRollup scan: %w", err)
		}
		results = append(results, row)
	}
	return results, nil
}

// GetTopPosts returns the top N posts ordered by the given metric (e.g. total_engagement, impressions).
// Supports optional hashtag filtering with OR logic across has(hashtags, tag) conditions.
// Mirrors PHP: LinkedInAnalyticsBuilder::getTopPostsData()
func (r *Repository) GetTopPosts(ctx context.Context, params *ch.QueryParams, limit int, orderBy string, hashtags []string) ([]TopPostResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("published_at", params)
	createdAtFilter := ch.DateFilter("created_at", params)
	partFilter := ch.PartitionMonthFilter("published_at", params)

	var extraFilters []string
	if len(hashtags) > 0 {
		conditions := make([]string, len(hashtags))
		for i, h := range hashtags {
			conditions[i] = fmt.Sprintf("has(hashtags, '%s')", strings.ReplaceAll(h, "'", "\\'"))
		}
		extraFilters = append(extraFilters, "("+strings.Join(conditions, " OR ")+")")
	}

	query := fmt.Sprintf(`%s
SELECT
    linkedin_id, post_id, activity, media_type, article_url, article_title,
    post_data, image, media, type, hashtags, comments, total_engagement,
    favorites, title, day_of_week, hour_of_day, created_at, saving_time,
    poll_data, reach, repost, post_clicks, impressions, published_at
FROM linkedin_posts
WHERE linkedin_id IN %s AND %s AND (post_id, saving_time) IN posts
ORDER BY %s DESC
LIMIT %d
`, ch.PostDedupCTE(ids, dateFilter, partFilter, extraFilters...), ids, createdAtFilter, orderBy, limit)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetTopPosts: %w", err)
	}
	defer rows.Close()

	var results []TopPostResult
	for rows.Next() {
		var row TopPostResult
		if err := rows.Scan(
			&row.LinkedinID, &row.PostID, &row.Activity, &row.MediaType,
			&row.ArticleURL, &row.ArticleTitle, &row.PostData, &row.Image,
			&row.Media, &row.Type, &row.Hashtags, &row.Comments,
			&row.TotalEngagement, &row.Favorites, &row.Title, &row.DayOfWeek,
			&row.HourOfDay, &row.CreatedAt, &row.SavingTime, &row.PollData,
			&row.Reach, &row.Repost, &row.PostClicks, &row.Impressions, &row.PublishedAt,
		); err != nil {
			return nil, fmt.Errorf("GetTopPosts scan: %w", err)
		}
		results = append(results, row)
	}
	return results, nil
}

// GetPostsPerDay returns the count of posts published on each day of the week.
// Mirrors PHP: LinkedInAnalyticsBuilder::getPostsPerDaysData()
func (r *Repository) GetPostsPerDay(ctx context.Context, params *ch.QueryParams) (*PostsPerDayResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("published_at", params)
	createdAtFilter := ch.DateFilter("created_at", params)
	partFilter := ch.PartitionMonthFilter("published_at", params)

	query := fmt.Sprintf(`%s
SELECT
    toInt32(countIf(day_of_week = 'Monday')) AS Monday,
    toInt32(countIf(day_of_week = 'Tuesday')) AS Tuesday,
    toInt32(countIf(day_of_week = 'Wednesday')) AS Wednesday,
    toInt32(countIf(day_of_week = 'Thursday')) AS Thursday,
    toInt32(countIf(day_of_week = 'Friday')) AS Friday,
    toInt32(countIf(day_of_week = 'Saturday')) AS Saturday,
    toInt32(countIf(day_of_week = 'Sunday')) AS Sunday
FROM linkedin_posts
WHERE linkedin_id IN %s AND %s AND (post_id, saving_time) IN posts
`, ch.PostDedupCTE(ids, dateFilter, partFilter), ids, createdAtFilter)

	var result PostsPerDayResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Monday, &result.Tuesday, &result.Wednesday,
		&result.Thursday, &result.Friday, &result.Saturday, &result.Sunday,
	)
	if err != nil {
		return nil, fmt.Errorf("GetPostsPerDay: %w", err)
	}
	return &result, nil
}

// GetTopHashtags returns the top 30 hashtags with their engagement breakdown.
// Uses arrayJoin to explode hashtag arrays and aggregates per hashtag.
// Mirrors PHP: LinkedInAnalyticsBuilder::getHashtagsData()
func (r *Repository) GetTopHashtags(ctx context.Context, params *ch.QueryParams) (*HashtagsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("published_at", params)
	createdAtFilter := ch.DateFilter("created_at", params)
	partFilter := ch.PartitionMonthFilter("published_at", params)

	query := fmt.Sprintf(`%s
SELECT
    groupArray(name) AS name, groupArray(engagements) AS engagements,
    groupArray(likes) AS likes, groupArray(comments) AS comments,
    groupArray(shares) AS shares, groupArray(posts) AS posts
FROM (
    SELECT
        arrayJoin(lp.hashtags) AS name,
        toInt32(SUM(lp.favorites)) AS likes, toInt32(SUM(lp.comments)) AS comments,
        toInt32(SUM(lp.repost)) AS shares,
        toInt32(likes + comments + shares) AS engagements,
        toInt32(count(DISTINCT lp.post_id)) AS posts
    FROM linkedin_posts AS lp
    INNER JOIN posts AS p ON lp.post_id = p.post_id AND lp.saving_time = p.saving_time
    WHERE lp.linkedin_id IN %s AND %s AND length(lp.hashtags) > 0
    GROUP BY name
    ORDER BY engagements DESC
    LIMIT 30
)
`, ch.PostDedupCTE(ids, dateFilter, partFilter), ids, createdAtFilter)

	var result HashtagsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Name, &result.Engagements, &result.Likes,
		&result.Comments, &result.Shares, &result.Posts,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTopHashtags: %w", err)
	}
	return &result, nil
}

// GetTopHashtagsRollup returns aggregated hashtag totals for a date range.
// Used for current vs previous period comparison in the hashtags widget.
func (r *Repository) GetTopHashtagsRollup(ctx context.Context, params *ch.QueryParams) (*HashtagsRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("published_at", params)
	createdAtFilter := ch.DateFilter("created_at", params)
	partFilter := ch.PartitionMonthFilter("published_at", params)

	query := fmt.Sprintf(`%s,
post_data AS (
    SELECT lp.hashtags, lp.favorites, lp.comments, lp.repost, lp.impressions, lp.reach
    FROM linkedin_posts AS lp
    INNER JOIN posts AS p ON lp.post_id = p.post_id AND lp.saving_time = p.saving_time
    WHERE lp.linkedin_id IN %s AND %s AND length(lp.hashtags) > 0
)
SELECT
    toInt32((SELECT COUNT(DISTINCT arrayJoin(hashtags)) FROM post_data)) AS total_hashtags,
    toInt32(SUM(length(hashtags))) AS total_times_used,
    toInt32(SUM(favorites)) AS total_likes,
    toInt32(SUM(comments)) AS total_comments,
    toInt32(SUM(repost)) AS total_shares,
    toInt32(total_likes + total_comments + total_shares) AS total_engagement,
    toInt32(SUM(impressions)) AS total_impressions,
    toInt32(SUM(reach)) AS total_reach
FROM post_data
`, ch.PostDedupCTE(ids, dateFilter, partFilter), ids, createdAtFilter)

	var result HashtagsRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalHashtags, &result.TotalTimesUsed,
		&result.TotalLikes, &result.TotalComments, &result.TotalShares,
		&result.TotalEngagement, &result.TotalImpressions, &result.TotalReach,
	)
	if err != nil {
		return nil, fmt.Errorf("GetTopHashtagsRollup: %w", err)
	}
	return &result, nil
}

// GetFollowersDemographics returns the most recent follower demographic breakdown
// (seniority, industry, country, city) from linkedin_insights.
// Mirrors PHP: LinkedInAnalyticsBuilder::getFollowersDemographicsData()
func (r *Repository) GetFollowersDemographics(ctx context.Context, params *ch.QueryParams) (*DemographicsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    followers_by_seniority,
    followers_by_industry,
    followers_by_country,
    followers_by_city,
    totalFollowerCount
FROM linkedin_insights
WHERE linkedin_id IN %s AND %s
ORDER BY created_at DESC
LIMIT 1
`, ids, dateFilter)

	var result DemographicsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.FollowersBySeniority,
		&result.FollowersByIndustry,
		&result.FollowersByCountry,
		&result.FollowersByCity,
		&result.TotalFollowerCount,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return &result, nil
	}
	if err != nil {
		return nil, fmt.Errorf("GetFollowersDemographics: %w", err)
	}
	return &result, nil
}
