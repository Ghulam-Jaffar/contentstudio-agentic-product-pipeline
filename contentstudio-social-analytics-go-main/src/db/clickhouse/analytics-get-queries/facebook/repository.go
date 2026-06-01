package facebook

import (
	"context"
	"fmt"
	"strings"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// Repository provides Facebook analytics queries against ClickHouse.
// All query patterns are migrated from the PHP FacebookAnalyticsBuilder
// (contentstudio-backend: app/Http/AnalyticsV2/Builders/FacebookAnalyticsBuilder.php).
type Repository struct {
	client *ch.Client
}

// NewRepository creates a new Facebook analytics repository.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// GetSummary returns combined post and page-level summary metrics for a date range.
// Uses CTEs to first deduplicate posts via max(saving_time), then cross-joins post and
// insights CTEs using LEFT JOIN ON 1=1 so a result row is always returned even when
// one side has no data.
// Mirrors PHP: FacebookAnalyticsBuilder::getSummaryData()
func (r *Repository) GetSummary(ctx context.Context, params *ch.QueryParams) (*SummaryResult, error) {
	query := fmt.Sprintf(`
WITH posts AS (
    SELECT post_id, max(saving_time) AS saving_time
    FROM facebook_posts
    WHERE %s AND %s
    GROUP BY post_id
),
post_summary_raw AS (
    SELECT
        toInt32(count()) AS doc_count,
        toInt32(sum(total)) AS reactions,
        toInt32(sum(comments)) AS comments,
        toInt32(sum(post_clicks)) AS posts_clicks,
        toInt32(sum(post_impressions)) AS impressions,
        toInt32(sum(post_impressions_unique)) AS reach,
        toInt32(sum(shares)) AS repost
    FROM facebook_posts
    WHERE (post_id, saving_time) IN (SELECT post_id, saving_time FROM posts)
),
post_summary AS (
    SELECT
        doc_count,
        toInt32(reactions + comments + repost + posts_clicks) AS total_engagement,
        reactions,
        comments,
        posts_clicks,
        impressions,
        reach,
        repost
    FROM post_summary_raw
),
insights_summary AS (
    SELECT
        toInt32(sum(positive_sentiment)) AS positive_sentiment,
        toInt32(sum(negative_sentiment)) AS negative_sentiment,
        toInt32(sum(page_impressions)) AS page_impressions,
        toInt32(sum(page_impressions_paid)) AS page_impressions_paid,
        toInt32(sum(page_impressions_organic)) AS page_impressions_organic,
        toInt32(sum(page_engagements)) AS page_engagements,
        toInt32(sum(page_positive_feedback)) AS page_positive_feedback,
        toInt32(sum(page_negative_feedback)) AS page_negative_feedback,
        toInt32(max(fan_count)) AS fan_count,
        toInt32(sum(talking_about_count)) AS talking_about_count,
        toInt32(max(page_follows)) AS page_follows
    FROM (
        SELECT
            toDate(created_time) AS created_date,
            argMin(page_fans, saving_time) AS fan_count,
            max(page_impressions) AS page_impressions,
            max(page_impressions_paid) AS page_impressions_paid,
            max(page_impressions_organic) AS page_impressions_organic,
            max(page_post_engagements) AS page_engagements,
            max(positive_sentiment) AS positive_sentiment,
            max(negative_sentiment) AS negative_sentiment,
            max(page_positive_feedback) AS page_positive_feedback,
            max(page_negative_feedback) AS page_negative_feedback,
            max(talking_about_count) AS talking_about_count,
            argMin(page_follows, saving_time) AS page_follows
        FROM facebook_insights
        WHERE %s AND %s AND %s
        GROUP BY page_id, created_date
    )
)
SELECT
    toInt32(coalesce(post_summary.doc_count, 0)) AS doc_count,
    toInt32(coalesce(post_summary.total_engagement, 0)) AS total_engagement,
    toInt32(coalesce(post_summary.reactions, 0)) AS reactions,
    toInt32(coalesce(post_summary.comments, 0)) AS comments,
    toInt32(coalesce(post_summary.posts_clicks, 0)) AS posts_clicks,
    toInt32(coalesce(post_summary.impressions, 0)) AS impressions,
    toInt32(coalesce(post_summary.reach, 0)) AS reach,
    toInt32(coalesce(post_summary.repost, 0)) AS repost,
    toInt32(coalesce(insights_summary.positive_sentiment, 0)) AS positive_sentiment,
    toInt32(coalesce(insights_summary.negative_sentiment, 0)) AS negative_sentiment,
    toInt32(coalesce(insights_summary.page_impressions, 0)) AS page_impressions,
    toInt32(coalesce(insights_summary.page_impressions_paid, 0)) AS page_impressions_paid,
    toInt32(coalesce(insights_summary.page_impressions_organic, 0)) AS page_impressions_organic,
    toInt32(coalesce(insights_summary.page_engagements, 0)) AS page_engagements,
    toInt32(coalesce(insights_summary.page_positive_feedback, 0)) AS page_positive_feedback,
    toInt32(coalesce(insights_summary.page_negative_feedback, 0)) AS page_negative_feedback,
    toInt32(coalesce(insights_summary.fan_count, 0)) AS fan_count,
    toInt32(coalesce(insights_summary.talking_about_count, 0)) AS talking_about_count,
    toInt32(coalesce(insights_summary.page_follows, 0)) AS page_follows
FROM (SELECT 1) AS stub
LEFT JOIN post_summary ON 1 = 1
LEFT JOIN insights_summary ON 1 = 1
`, accountFilter("page_id", params), dateTimeFilter("created_time", params), accountFilter("page_id", params), dateTimeFilter("created_time", params), ch.PartitionMonthFilter("created_time", params))

	var result SummaryResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.DocCount,
		&result.TotalEngagement,
		&result.Reactions,
		&result.Comments,
		&result.PostsClicks,
		&result.Impressions,
		&result.Reach,
		&result.Repost,
		&result.PositiveSentiment,
		&result.NegativeSentiment,
		&result.PageImpressions,
		&result.PageImpressionsPaid,
		&result.PageImpressionsOrganic,
		&result.PageEngagements,
		&result.PagePositiveFeedback,
		&result.PageNegativeFeedback,
		&result.FanCount,
		&result.TalkingAboutCount,
		&result.PageFollows,
	); err != nil {
		return nil, fmt.Errorf("GetSummary: %w", err)
	}

	return &result, nil
}

// GetPostsSummary returns aggregated post-level metrics from facebook_posts.
// Deduplicates via max(saving_time) per post_id and sums engagement fields.
// Run concurrently with GetInsightsSummary for maximum throughput.
func (r *Repository) GetPostsSummary(ctx context.Context, params *ch.QueryParams) (*PostsSummaryResult, error) {
	query := fmt.Sprintf(`
WITH posts AS (
    SELECT post_id, max(saving_time) AS saving_time
    FROM facebook_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    toInt32(count()) AS doc_count,
    toInt32(sum(total + comments + shares + post_clicks)) AS total_engagement,
    toInt32(sum(total)) AS reactions,
    toInt32(sum(comments)) AS post_comments,
    toInt32(sum(post_clicks)) AS posts_clicks,
    toInt32(sum(post_impressions)) AS impressions,
    toInt32(sum(post_impressions_unique)) AS reach,
    toInt32(sum(shares)) AS repost
FROM facebook_posts
WHERE %s AND %s AND (post_id, saving_time) IN (SELECT post_id, saving_time FROM posts)
`, accountFilter("page_id", params), dateTimeFilter("created_time", params), accountFilter("page_id", params), dateTimeFilter("created_time", params))

	var result PostsSummaryResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.DocCount,
		&result.TotalEngagement,
		&result.Reactions,
		&result.Comments,
		&result.PostsClicks,
		&result.Impressions,
		&result.Reach,
		&result.Repost,
	); err != nil {
		return nil, fmt.Errorf("GetPostsSummary: %w", err)
	}
	return &result, nil
}

// GetInsightsSummary returns aggregated page-level metrics from facebook_insights.
// Deduplicates by (page_id, date) using argMin/max per row, then sums/maxes across days.
// Run concurrently with GetPostsSummary for maximum throughput.
func (r *Repository) GetInsightsSummary(ctx context.Context, params *ch.QueryParams) (*InsightsSummaryResult, error) {
	query := fmt.Sprintf(`
SELECT
    toInt32(sum(positive_sentiment)) AS positive_sentiment,
    toInt32(sum(negative_sentiment)) AS negative_sentiment,
    toInt32(sum(page_impressions)) AS page_impressions,
    toInt32(sum(page_impressions_paid)) AS page_impressions_paid,
    toInt32(sum(page_impressions_organic)) AS page_impressions_organic,
    toInt32(sum(page_engagements)) AS page_engagements,
    toInt32(sum(page_positive_feedback)) AS page_positive_feedback,
    toInt32(sum(page_negative_feedback)) AS page_negative_feedback,
    toInt32(max(fan_count)) AS fan_count,
    toInt32(sum(talking_about_count)) AS talking_about_count,
    toInt32(max(page_follows)) AS page_follows
FROM (
    SELECT
        toDate(created_time) AS created_date,
        argMin(page_fans, saving_time) AS fan_count,
        max(page_impressions) AS page_impressions,
        max(page_impressions_paid) AS page_impressions_paid,
        max(page_impressions_organic) AS page_impressions_organic,
        max(page_post_engagements) AS page_engagements,
        max(positive_sentiment) AS positive_sentiment,
        max(negative_sentiment) AS negative_sentiment,
        max(page_positive_feedback) AS page_positive_feedback,
        max(page_negative_feedback) AS page_negative_feedback,
        max(talking_about_count) AS talking_about_count,
        argMin(page_follows, saving_time) AS page_follows
    FROM facebook_insights
    WHERE %s AND %s AND %s
    GROUP BY page_id, created_date
)
`, accountFilter("page_id", params), dateTimeFilter("created_time", params), ch.PartitionMonthFilter("created_time", params))

	var result InsightsSummaryResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PositiveSentiment,
		&result.NegativeSentiment,
		&result.PageImpressions,
		&result.PageImpressionsPaid,
		&result.PageImpressionsOrganic,
		&result.PageEngagements,
		&result.PagePositiveFeedback,
		&result.PageNegativeFeedback,
		&result.FanCount,
		&result.TalkingAboutCount,
		&result.PageFollows,
	); err != nil {
		return nil, fmt.Errorf("GetInsightsSummary: %w", err)
	}
	return &result, nil
}

// GetAudienceGrowth returns time-series fan data from facebook_insights with daily deltas.
// Uses lagInFrame() window function to compute the day-over-day fan change and
// arrayFill to forward-fill zero-value entries caused by missing data days.
// WITH FILL FROM/TO ensures a continuous date array for the frontend.
// Mirrors PHP: FacebookAnalyticsBuilder::getAudienceGrowthData()
func (r *Repository) GetAudienceGrowth(ctx context.Context, params *ch.QueryParams) (*AudienceGrowthResult, error) {
	query := fmt.Sprintf(`
SELECT
    notEmpty(fan_count_temp) AS show_data,
    arrayFill(x -> not x == 0, fan_count_temp) AS fan_count,
    page_fans_daily,
    page_fans_by_like_temp AS page_fans_by_like,
    page_fans_by_unlike_temp AS page_fans_by_unlike,
    page_impressions,
    page_engagements,
    date AS buckets
FROM (
    SELECT
        groupArray(page_fans_total) AS fan_count_temp,
        groupArray(page_fans_daily) AS page_fans_daily,
        groupArray(page_fans_by_like) AS page_fans_by_like_temp,
        groupArray(page_fans_by_unlike) AS page_fans_by_unlike_temp,
        groupArray(page_impressions) AS page_impressions,
        groupArray(page_engagements) AS page_engagements,
        groupArray(created_date) AS date
    FROM (
        SELECT
            toInt32(page_fans_total) AS page_fans_total,
            toInt32(if(page_fans_total > 0, page_fans_total - lagInFrame(page_fans_total, 1, page_fans_total) OVER (ORDER BY created_date ASC), 0)) AS page_fans_daily,
            toInt32(page_fans_by_like) AS page_fans_by_like,
            toInt32(page_fans_by_unlike) AS page_fans_by_unlike,
            toInt32(page_impressions) AS page_impressions,
            toInt32(page_engagements) AS page_engagements,
            created_date
        FROM (
            SELECT
                argMin(page_follows, saving_time) AS page_fans_total,
                argMin(page_fans_by_like, saving_time) AS page_fans_by_like,
                argMin(page_fans_by_unlike, saving_time) AS page_fans_by_unlike,
                max(page_impressions) AS page_impressions,
                max(page_post_engagements) AS page_engagements,
                toDate(created_time) AS created_date
            FROM facebook_insights
            WHERE %s AND %s AND %s
            GROUP BY created_date
            ORDER BY created_date ASC
            %s
        )
    )
)`, accountFilter("page_id", params), dateTimeFilter("created_time", params), ch.PartitionMonthFilter("created_time", params), fillFromTo(params))

	var result AudienceGrowthResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.FanCount,
		&result.PageFansDaily,
		&result.PageFansByLike,
		&result.PageFansByUnlike,
		&result.PageImpressions,
		&result.PageEngagements,
		&result.Buckets,
	); err != nil {
		return nil, fmt.Errorf("GetAudienceGrowth: %w", err)
	}

	return &result, nil
}

// GetLastFollowerCounts returns the most recent non-zero fan counts from facebook_insights.
// arrayFirst selects the first non-zero value from the ordered group array, providing
// a stable fallback when the selected period starts with missing fan data.
func (r *Repository) GetLastFollowerCounts(ctx context.Context, params *ch.QueryParams) (*LastFollowerCounts, error) {
	query := fmt.Sprintf(`
SELECT
    arrayFirst(x -> x != 0, groupArray(fans)) AS page_fans,
    arrayFirst(x -> x != 0, groupArray(page_fans_by_like)) AS page_fans_by_like,
    arrayFirst(x -> x != 0, groupArray(page_fans_by_unlike)) AS page_fans_by_unlike
FROM (
    SELECT
        last_value(created_time) AS inserted_time,
        toInt32(last_value(page_fans)) AS fans,
        toInt32(last_value(page_fans_by_like)) AS page_fans_by_like,
        toInt32(last_value(page_fans_by_unlike)) AS page_fans_by_unlike
    FROM facebook_insights
    WHERE %s AND %s AND %s AND page_fans != 0
    GROUP BY hash_id
    ORDER BY inserted_time DESC
)`, accountFilter("page_id", params), dateOnlyFilter("created_time", params), ch.PartitionMonthFilter("created_time", params))

	var result LastFollowerCounts
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PageFans,
		&result.PageFansByLike,
		&result.PageFansByUnlike,
	); err != nil {
		return nil, fmt.Errorf("GetLastFollowerCounts: %w", err)
	}

	return &result, nil
}

// GetAudienceGrowthRollup returns aggregated fan metrics for a date range.
// Used for current vs previous period comparison in the audience growth widget.
func (r *Repository) GetAudienceGrowthRollup(ctx context.Context, params *ch.QueryParams) (*AudienceGrowthRollupResult, error) {
	query := fmt.Sprintf(`
SELECT
    round(avg(page_fans_by_like), 2) AS avg_page_fans_by_like,
    round(avg(page_fans_by_unlike), 2) AS avg_page_fans_by_unlike,
    toInt32(max(page_follows)) AS fan_count,
    toInt32(sum(talking_about_count)) AS talking_about_count,
    toInt32(count()) AS doc_count
FROM (
    SELECT
        argMin(page_fans_by_like, saving_time) AS page_fans_by_like,
        argMin(page_fans_by_unlike, saving_time) AS page_fans_by_unlike,
        argMin(page_follows, saving_time) AS page_follows,
        max(talking_about_count) AS talking_about_count,
        toDate(created_time) AS created_date
    FROM facebook_insights
    WHERE %s AND %s AND %s
    GROUP BY page_id, created_date
    ORDER BY created_date ASC
)`, accountFilter("page_id", params), dateTimeFilter("created_time", params), ch.PartitionMonthFilter("created_time", params))

	var result AudienceGrowthRollupResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.AvgPageFansByLike,
		&result.AvgPageFansByUnlike,
		&result.FanCount,
		&result.TalkingAboutCount,
		&result.DocCount,
	); err != nil {
		return nil, fmt.Errorf("GetAudienceGrowthRollup: %w", err)
	}

	return &result, nil
}

// GetPublishingBehaviour returns time-series post engagement and reach broken down by day.
// Posts are deduplicated via max(saving_time) per post_id, then aggregated by date.
// WITH FILL TO ensures the date array is continuous to the end of the period.
// Mirrors PHP: FacebookAnalyticsBuilder::getPublishingBehaviourData()
func (r *Repository) GetPublishingBehaviour(ctx context.Context, params *ch.QueryParams, mediaTypes []string) (*PublishingBehaviourResult, error) {
	query := fmt.Sprintf(`
WITH posts AS (
    SELECT post_id, max(saving_time) AS saving_time
    FROM facebook_posts
    WHERE %s AND media_type IN (%s) AND %s
    GROUP BY post_id
)
SELECT
    groupArray(reactions) AS reactions_engagement,
    groupArray(comments) AS comments_engagement,
    groupArray(shares) AS shares_engagement,
    groupArray(paid_impressions) AS paid_impressions,
    groupArray(organic_impressions) AS organic_impressions,
    groupArray(viral_impressions) AS viral_impressions,
    groupArray(paid_reach) AS paid_reach,
    groupArray(organic_reach) AS organic_reach,
    groupArray(viral_reach) AS viral_reach,
    groupArray(created_date) AS buckets,
    groupArray(post_count) AS post_count
FROM (
    SELECT
        count() AS post_count,
        sum(total) AS reactions,
        sum(comments) AS comments,
        sum(shares) AS shares,
        sum(post_impressions_paid) AS paid_impressions,
        sum(post_impressions_organic) AS organic_impressions,
        sum(post_impressions_viral) AS viral_impressions,
        sum(post_impressions_paid_unique) AS paid_reach,
        sum(post_impressions_organic_unique) AS organic_reach,
        sum(post_impressions_viral_unique) AS viral_reach,
        toDate(created_time) AS created_date
    FROM facebook_posts
    WHERE %s AND %s AND (post_id, saving_time) IN (SELECT post_id, saving_time FROM posts)
    GROUP BY created_date
    ORDER BY created_date ASC
    %s
)`, accountFilter("page_id", params), ch.FormatStringList(mediaTypes), dateTimeFilter("created_time", params), accountFilter("page_id", params), dateTimeFilter("created_time", params), fillTo(params))

	var result PublishingBehaviourResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ReactionsEngagement,
		&result.CommentsEngagement,
		&result.SharesEngagement,
		&result.PaidImpressions,
		&result.OrganicImpressions,
		&result.ViralImpressions,
		&result.PaidReach,
		&result.OrganicReach,
		&result.ViralReach,
		&result.Buckets,
		&result.PostCount,
	); err != nil {
		return nil, fmt.Errorf("GetPublishingBehaviour: %w", err)
	}

	return &result, nil
}

// GetPublishingBehaviourRollup returns aggregated post metrics for current/previous period comparison.
// Deduplicates posts via max(saving_time), then sums engagement fields over the period.
func (r *Repository) GetPublishingBehaviourRollup(ctx context.Context, params *ch.QueryParams) (*PublishingRollupResult, error) {
	query := fmt.Sprintf(`
WITH posts AS (
    SELECT post_id, max(saving_time) AS saving_time
    FROM facebook_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    doc_count,
    toInt32(reactions + comments + shares + post_clicks) AS total_engagement,
    reactions,
    comments,
    post_clicks,
    impressions,
    shares
FROM (
    SELECT
        toInt32(count()) AS doc_count,
        toInt32(sum(total)) AS reactions,
        toInt32(sum(comments)) AS comments,
        toInt32(sum(post_clicks)) AS post_clicks,
        toInt32(sum(post_impressions)) AS impressions,
        toInt32(sum(shares)) AS shares
    FROM facebook_posts
    WHERE %s AND %s AND (post_id, saving_time) IN (SELECT post_id, saving_time FROM posts)
)
`, accountFilter("page_id", params), dateTimeFilter("created_time", params), accountFilter("page_id", params), dateTimeFilter("created_time", params))

	var result PublishingRollupResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.DocCount,
		&result.TotalEngagement,
		&result.Reactions,
		&result.Comments,
		&result.PostClicks,
		&result.Impressions,
		&result.Shares,
	); err != nil {
		return nil, fmt.Errorf("GetPublishingBehaviourRollup: %w", err)
	}

	return &result, nil
}

// GetTopPosts returns the top N posts ordered by the given metric, joined with media assets.
// Uses LIMIT N BY page_id to respect per-page caps, then LEFT JOINs facebook_media_assets
// so each asset row is returned separately (the service layer collapses them into MediaAssets).
// Mirrors PHP: FacebookAnalyticsBuilder::getTopPostsData()
func (r *Repository) GetTopPosts(ctx context.Context, params *ch.QueryParams, mediaTypes []string, limit int, orderBy string) ([]TopPostRow, error) {
	query := fmt.Sprintf(`
WITH posts AS (
    SELECT post_id, max(saving_time) AS saving_time
    FROM facebook_posts
    WHERE %s AND media_type IN (%s) AND %s
    GROUP BY post_id
),
latest_posts AS (
    SELECT
        toString(page_name) AS page_name,
        toString(page_id) AS page_id,
        toString(post_id) AS post_id,
        toString(permalink) AS permalink,
        toString(status_type) AS status_type,
        toString(media_type) AS media_type,
        toString(video_id) AS video_id,
        toString(category) AS category,
        toString(published_by) AS published_by,
        toString(published_by_url) AS published_by_url,
        toString(shared_from_name) AS shared_from_name,
        toString(shared_from_id) AS shared_from_id,
        toString(shared_from_link) AS shared_from_link,
        toInt64(like) AS like,
        toInt64(love) AS love,
        toInt64(haha) AS haha,
        toInt64(wow) AS wow,
        toInt64(sad) AS sad,
        toInt64(angry) AS angry,
        toInt64(total) AS total,
        toInt64(shares) AS shares,
        toInt64(comments) AS comments,
        toInt64(post_clicks) AS post_clicks,
        toFloat64(total + comments + shares + post_clicks) AS total_engagement,
        toInt64(post_engaged_users) AS post_engaged_users,
        toString(day_of_week) AS day_of_week,
        toInt64(hour_of_day) AS hour_of_day,
        created_time,
        updated_time,
        saving_time,
        toString(message_tags) AS message_tags,
        toString(post_metadata) AS post_metadata,
        toString(caption) AS caption,
        toString(description) AS description,
        toString(full_picture) AS full_picture,
        toString(link) AS link,
        toInt64(post_impressions) AS post_impressions,
        toInt64(post_impressions_unique) AS post_impressions_unique,
        toInt64(post_impressions_paid) AS post_impressions_paid,
        toInt64(post_impressions_paid_unique) AS post_impressions_paid_unique,
        toInt64(post_impressions_organic) AS post_impressions_organic,
        toInt64(post_impressions_organic_unique) AS post_impressions_organic_unique,
        toInt64(post_impressions_viral) AS post_impressions_viral,
        toInt64(post_impressions_viral_unique) AS post_impressions_viral_unique,
        toInt64(post_video_views) AS post_video_views,
        toInt64(total_impressions) AS total_impressions
    FROM facebook_posts
    WHERE (post_id, saving_time) IN (SELECT post_id, saving_time FROM posts)
    ORDER BY %s DESC, created_time
    LIMIT %d BY page_id
),
assets AS (
    SELECT
        toString(post_id) AS post_id,
        toString(media_id) AS media_id,
        toString(caption) AS media_caption,
        toString(link) AS media_link,
        toString(asset_type) AS asset_type,
        toString(call_to_action) AS call_to_action,
        created_at AS asset_created_at
    FROM facebook_media_assets
    WHERE %s AND %s
)
SELECT
    latest_posts.page_name,
    latest_posts.page_id,
    latest_posts.post_id,
    latest_posts.permalink,
    latest_posts.status_type,
    latest_posts.media_type,
    latest_posts.video_id,
    latest_posts.category,
    latest_posts.published_by,
    latest_posts.published_by_url,
    latest_posts.shared_from_name,
    latest_posts.shared_from_id,
    latest_posts.shared_from_link,
    latest_posts.like,
    latest_posts.love,
    latest_posts.haha,
    latest_posts.wow,
    latest_posts.sad,
    latest_posts.angry,
    latest_posts.total,
    latest_posts.shares,
    latest_posts.comments,
    latest_posts.post_clicks,
    latest_posts.total_engagement,
    latest_posts.post_engaged_users,
    latest_posts.day_of_week,
    latest_posts.hour_of_day,
    latest_posts.created_time,
    latest_posts.updated_time,
    latest_posts.saving_time,
    latest_posts.message_tags,
    latest_posts.post_metadata,
    latest_posts.caption,
    latest_posts.description,
    latest_posts.full_picture,
    latest_posts.link,
    latest_posts.post_impressions,
    latest_posts.post_impressions_unique,
    latest_posts.post_impressions_paid,
    latest_posts.post_impressions_paid_unique,
    latest_posts.post_impressions_organic,
    latest_posts.post_impressions_organic_unique,
    latest_posts.post_impressions_viral,
    latest_posts.post_impressions_viral_unique,
    latest_posts.post_video_views,
    latest_posts.total_impressions,
    ifNull(assets.media_id, '') AS media_id,
    ifNull(assets.media_caption, '') AS media_caption,
    ifNull(assets.media_link, '') AS media_link,
    ifNull(assets.asset_type, '') AS asset_type,
    ifNull(assets.call_to_action, '') AS call_to_action,
    ifNull(assets.asset_created_at, toDateTime('1970-01-01 00:00:00')) AS asset_created_at
FROM latest_posts
LEFT JOIN assets USING (post_id)
ORDER BY latest_posts.%s DESC, latest_posts.created_time
`, accountFilter("page_id", params), ch.FormatStringList(mediaTypes), dateTimeFilter("created_time", params), orderBy, limit, accountFilter("page_id", params), dateTimeFilter("created_at", params), orderBy)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetTopPosts: %w", err)
	}
	defer rows.Close()

	var out []TopPostRow
	for rows.Next() {
		var row TopPostRow
		if err := rows.Scan(
			&row.PageName,
			&row.PageID,
			&row.PostID,
			&row.Permalink,
			&row.StatusType,
			&row.MediaType,
			&row.VideoID,
			&row.Category,
			&row.PublishedBy,
			&row.PublishedByURL,
			&row.SharedFromName,
			&row.SharedFromID,
			&row.SharedFromLink,
			&row.Like,
			&row.Love,
			&row.Haha,
			&row.Wow,
			&row.Sad,
			&row.Angry,
			&row.Total,
			&row.Shares,
			&row.Comments,
			&row.PostClicks,
			&row.TotalEngagement,
			&row.PostEngagedUsers,
			&row.DayOfWeek,
			&row.HourOfDay,
			&row.CreatedTime,
			&row.UpdatedTime,
			&row.SavingTime,
			&row.MessageTags,
			&row.PostMetadata,
			&row.Caption,
			&row.Description,
			&row.FullPicture,
			&row.Link,
			&row.PostImpressions,
			&row.PostImpressionsUnique,
			&row.PostImpressionsPaid,
			&row.PostImpressionsPaidUnique,
			&row.PostImpressionsOrganic,
			&row.PostImpressionsOrganicUnique,
			&row.PostImpressionsViral,
			&row.PostImpressionsViralUnique,
			&row.PostVideoViews,
			&row.TotalImpressions,
			&row.MediaID,
			&row.MediaCaption,
			&row.MediaLink,
			&row.AssetType,
			&row.CallToAction,
			&row.AssetCreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetTopPosts: %w", err)
		}
		out = append(out, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetTopPosts: %w", err)
	}

	return out, nil
}

// GetActiveUsersHours returns average fans-online per hour of the day from facebook_insights.
// The page_fans_online field is stored as a '$'-delimited string; splitByChar parses it into
// [hour, value] pairs which are then averaged across days using sumForEach / count.
// Mirrors PHP: FacebookAnalyticsBuilder::getActiveUsersHoursData()
func (r *Repository) GetActiveUsersHours(ctx context.Context, params *ch.QueryParams) (*ActiveUsersHoursResult, error) {
	query := fmt.Sprintf(`
SELECT
    max(buckets) AS buckets,
    max(value) AS values,
    toInt32(arrayMax(values)) AS highest_value,
    toInt32(buckets[indexOf(values, highest_value)]) AS highest_hour
FROM (
    SELECT
        max(buckets) AS buckets,
        arrayMap(x -> toInt32(x / count()), sumForEach(values)) AS value
    FROM (
        SELECT
            arrayMap(x -> toInt32(x[1]), active_users_per_hour) AS buckets,
            arrayMap(x -> toInt32(x[2]), active_users_per_hour) AS values
        FROM (
            SELECT arrayMap(x -> splitByChar('$', x), arr) AS active_users_per_hour
            FROM (
                SELECT max(page_fans_online) AS arr
                FROM facebook_insights
                WHERE %s AND %s
                GROUP BY hash_id
            )
        )
    )
)`, accountFilter("page_id", params), dateOnlyFilter("created_time", params))

	var result ActiveUsersHoursResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Buckets,
		&result.Values,
		&result.HighestValue,
		&result.HighestHour,
	); err != nil {
		return nil, fmt.Errorf("GetActiveUsersHours: %w", err)
	}

	return &result, nil
}

// GetActiveUsersDays returns the count of active-insight records per day of the week.
// WITH FILL FROM 1 TO 8 STEP 1 ensures all seven days are present even if some have no data.
// Mirrors PHP: FacebookAnalyticsBuilder::getActiveUsersDaysData()
func (r *Repository) GetActiveUsersDays(ctx context.Context, params *ch.QueryParams) (*ActiveUsersDaysResult, error) {
	query := fmt.Sprintf(`
SELECT
    groupArray(day_name) AS buckets,
    groupArray(active_users) AS values,
    max(active_users) AS highest_value,
    buckets[indexOf(values, highest_value)] AS highest_day
FROM (
    SELECT
        ['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'][day_num] AS day_name,
        ifNull(active_users, 0) AS active_users
    FROM (
        SELECT
            day_num,
            toInt32(count()) AS active_users
        FROM (
            SELECT
                indexOf(['Monday', 'Tuesday', 'Wednesday', 'Thursday', 'Friday', 'Saturday', 'Sunday'], last_value(day_of_week)) AS day_num
            FROM facebook_insights
            WHERE %s AND %s
            GROUP BY hash_id
        )
        GROUP BY day_num
        ORDER BY day_num ASC
        WITH FILL FROM 1 TO 8 STEP 1
    )
)`, accountFilter("page_id", params), dateOnlyFilter("created_time", params))

	var result ActiveUsersDaysResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Buckets,
		&result.Values,
		&result.HighestValue,
		&result.HighestDay,
	); err != nil {
		return nil, fmt.Errorf("GetActiveUsersDays: %w", err)
	}

	return &result, nil
}

// GetImpressions returns time-series page impressions from facebook_insights.
// Deduplicates by date using max(page_impressions) per day with WITH FILL TO for continuity.
// Mirrors PHP: FacebookAnalyticsBuilder::getImpressionsData()
func (r *Repository) GetImpressions(ctx context.Context, params *ch.QueryParams) (*ImpressionsResult, error) {
	query := fmt.Sprintf(`
SELECT
    groupArray(page_impressions) AS page_impressions,
    groupArray(created_date) AS buckets
FROM (
    SELECT
        toInt32(max(page_impressions)) AS page_impressions,
        toDate(created_time) AS created_date
    FROM facebook_insights
    WHERE %s AND %s
    GROUP BY created_date
    ORDER BY created_date ASC
    %s
)`, accountFilter("page_id", params), dateTimeFilter("created_time", params), fillTo(params))

	var result ImpressionsResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PageImpressions,
		&result.Buckets,
	); err != nil {
		return nil, fmt.Errorf("GetImpressions: %w", err)
	}

	return &result, nil
}

// GetImpressionsRollup returns total and average impressions for a date range.
// Averages are computed by dividing by DayCount (daily) and maxWeeks(DayCount) (weekly).
func (r *Repository) GetImpressionsRollup(ctx context.Context, params *ch.QueryParams) (*ImpressionsRollupResult, error) {
	query := fmt.Sprintf(`
SELECT
    toInt32(sum(page_impressions)) AS total_impressions,
    round(sum(page_impressions) / %d, 2) AS avg_impressions_per_day,
    round(sum(page_impressions) / %d, 2) AS avg_impressions_per_week
FROM (
    SELECT
        max(page_impressions) AS page_impressions,
        toDate(created_time) AS created_date
    FROM facebook_insights
    WHERE %s AND %s
    GROUP BY page_id, created_date
)`, params.DayCount, maxWeeks(params.DayCount), accountFilter("page_id", params), dateTimeFilter("created_time", params))

	var result ImpressionsRollupResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalImpressions,
		&result.AvgImpressionsPerDay,
		&result.AvgImpressionsPerWeek,
	); err != nil {
		return nil, fmt.Errorf("GetImpressionsRollup: %w", err)
	}

	return &result, nil
}

// GetEngagement returns time-series page engagement from facebook_insights.
// Mirrors PHP: FacebookAnalyticsBuilder::getEngagementData()
func (r *Repository) GetEngagement(ctx context.Context, params *ch.QueryParams) (*EngagementResult, error) {
	query := fmt.Sprintf(`
SELECT
    groupArray(page_engagements) AS page_engagements,
    groupArray(created_date) AS buckets
FROM (
    SELECT
        toInt32(max(page_post_engagements)) AS page_engagements,
        toDate(created_time) AS created_date
    FROM facebook_insights
    WHERE %s AND %s
    GROUP BY created_date
    ORDER BY created_date ASC
    %s
)`, accountFilter("page_id", params), dateTimeFilter("created_time", params), fillTo(params))

	var result EngagementResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PageEngagements,
		&result.Buckets,
	); err != nil {
		return nil, fmt.Errorf("GetEngagement: %w", err)
	}

	return &result, nil
}

// GetEngagementRollup returns total and average page engagements for a date range.
// Averages are computed by dividing by DayCount and maxWeeks(DayCount) respectively.
func (r *Repository) GetEngagementRollup(ctx context.Context, params *ch.QueryParams) (*EngagementRollupResult, error) {
	query := fmt.Sprintf(`
SELECT
    toInt32(sum(page_post_engagements)) AS page_engagements,
    round(sum(page_post_engagements) / %d, 2) AS avg_engagements_per_day,
    round(sum(page_post_engagements) / %d, 2) AS avg_engagements_per_week
FROM (
    SELECT
        max(page_post_engagements) AS page_post_engagements,
        toDate(created_time) AS created_date
    FROM facebook_insights
    WHERE %s AND %s
    GROUP BY page_id, created_date
)`, params.DayCount, maxWeeks(params.DayCount), accountFilter("page_id", params), dateTimeFilter("created_time", params))

	var result EngagementRollupResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PageEngagements,
		&result.AvgEngagementsPerDay,
		&result.AvgEngagementsPerWeek,
	); err != nil {
		return nil, fmt.Errorf("GetEngagementRollup: %w", err)
	}

	return &result, nil
}

// GetReelsAnalytics returns time-series reels performance data.
// Joins facebook_reels_insights (watch time, plays) with facebook_posts (engagement) by post_id.
// last_value() deduplicates per-reel metrics before the daily GROUP BY.
// WITH FILL TO provides date continuity.
// Mirrors PHP: FacebookAnalyticsBuilder::getReelsAnalyticsData()
func (r *Repository) GetReelsAnalytics(ctx context.Context, params *ch.QueryParams) (*ReelsAnalyticsResult, error) {
	query := fmt.Sprintf(`
WITH facebook_post_data_raw AS (
    SELECT
        post_id,
        last_value(total) AS reactions,
        last_value(comments) AS comments,
        last_value(shares) AS repost,
        last_value(post_clicks) AS post_clicks,
        toDate(created_time) AS created_at
    FROM facebook_posts
    WHERE %s AND %s
    GROUP BY post_id, created_at
),
facebook_post_data AS (
    SELECT
        post_id,
        reactions,
        comments,
        repost,
        post_clicks,
        toInt32(reactions + comments + repost + post_clicks) AS total_engagement,
        created_at
    FROM facebook_post_data_raw
)
SELECT
    groupArray(created_at) AS buckets,
    groupArray(total_reels_count) AS total_reels,
    groupArray(total_seconds_watched) AS total_seconds_watched,
    groupArray(initial_plays) AS initial_plays,
    groupArray(total_engagement) AS engagement,
    groupArray(reactions) AS reactions,
    groupArray(comments) AS comments,
    groupArray(shares) AS shares,
    toInt32(sum(total_reels_count)) AS show_data
FROM (
    SELECT
        created_at,
        toInt32(count()) AS total_reels_count,
        round(sum(total_time_watched_in_ms) / 1000, 2) AS total_seconds_watched,
        toInt32(sum(play_count)) AS initial_plays,
        toInt32(sum(total_engagement)) AS total_engagement,
        toInt32(sum(reactions)) AS reactions,
        toInt32(sum(comments)) AS comments,
        toInt32(sum(repost)) AS shares
    FROM (
        SELECT
            post_id,
            toDate(created_at) AS created_at,
            last_value(total_time_watched_in_ms) AS total_time_watched_in_ms,
            last_value(play_count) AS play_count,
            last_value(impressions_unique) AS reach
        FROM facebook_reels_insights
        WHERE %s AND %s
        GROUP BY post_id, created_at
    ) AS reels
    LEFT JOIN facebook_post_data ON facebook_post_data.post_id = reels.post_id
    GROUP BY reels.created_at
    ORDER BY created_at ASC
    %s
)`, accountFilter("page_id", params), dateTimeFilter("created_time", params), accountFilter("page_id", params), dateTimeFilter("created_at", params), fillTo(params))

	var result ReelsAnalyticsResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Buckets,
		&result.TotalReels,
		&result.TotalSecondsWatched,
		&result.InitialPlays,
		&result.Engagement,
		&result.Reactions,
		&result.Comments,
		&result.Shares,
		&result.ShowData,
	); err != nil {
		return nil, fmt.Errorf("GetReelsAnalytics: %w", err)
	}

	return &result, nil
}

// GetReelsRollup returns aggregated reels totals for a date range.
// average_seconds_watched is guarded with an if(initial_plays != 0, ...) to prevent division by zero.
func (r *Repository) GetReelsRollup(ctx context.Context, params *ch.QueryParams) (*ReelsRollupResult, error) {
	query := fmt.Sprintf(`
WITH facebook_post_data_raw AS (
    SELECT
        post_id,
        last_value(total) AS reactions,
        last_value(comments) AS comments,
        last_value(shares) AS repost,
        last_value(post_clicks) AS post_clicks,
        toDate(created_time) AS created_at
    FROM facebook_posts
    WHERE %s AND %s
    GROUP BY post_id, created_at
),
facebook_post_data AS (
    SELECT
        post_id,
        reactions,
        comments,
        repost,
        post_clicks,
        toInt32(reactions + comments + repost + post_clicks) AS total_engagement,
        created_at
    FROM facebook_post_data_raw
)
SELECT
    toInt32(count()) AS total_reels,
    if(initial_plays != 0, round(total_seconds_watched / initial_plays, 2), 0) AS average_seconds_watched,
    toInt32(sum(total_time_watched_in_ms) / 1000) AS total_seconds_watched,
    toInt32(sum(play_count)) AS initial_plays,
    toInt32(sum(reach)) AS reach,
    toInt32(sum(total_engagement)) AS engagement,
    toInt32(sum(reactions)) AS reactions,
    toInt32(sum(comments)) AS comments,
    toInt32(sum(repost)) AS shares
FROM (
    SELECT
        post_id,
        last_value(total_time_watched_in_ms) AS total_time_watched_in_ms,
        last_value(play_count) AS play_count,
        last_value(impressions_unique) AS reach
    FROM facebook_reels_insights
    WHERE %s AND %s
    GROUP BY post_id
) AS reels
LEFT JOIN facebook_post_data ON facebook_post_data.post_id = reels.post_id
`, accountFilter("page_id", params), dateTimeFilter("created_time", params), accountFilter("page_id", params), dateTimeFilter("created_at", params))

	var result ReelsRollupResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalReels,
		&result.AverageSecondsWatched,
		&result.TotalSecondsWatched,
		&result.InitialPlays,
		&result.Reach,
		&result.Engagement,
		&result.Reactions,
		&result.Comments,
		&result.Shares,
	); err != nil {
		return nil, fmt.Errorf("GetReelsRollup: %w", err)
	}

	return &result, nil
}

// GetVideoInsights returns time-series video performance data.
// Joins facebook_video_insights with facebook_posts by post_id; last_value() deduplicates
// per-video metrics and divides total_video_view_total_time by 1000 to convert ms to seconds.
// WITH FILL TO provides date continuity.
// Mirrors PHP: FacebookAnalyticsBuilder::getVideoInsightsData()
func (r *Repository) GetVideoInsights(ctx context.Context, params *ch.QueryParams) (*VideoInsightsResult, error) {
	query := fmt.Sprintf(`
WITH posts_raw AS (
    SELECT
        post_id,
        last_value(total) AS reactions,
        last_value(comments) AS comments,
        last_value(shares) AS repost,
        last_value(post_clicks) AS post_clicks,
        toDate(created_time) AS created_at
    FROM facebook_posts
    WHERE %s AND %s
    GROUP BY post_id, created_at
),
posts AS (
    SELECT
        post_id,
        reactions,
        comments,
        repost,
        post_clicks,
        toInt32(reactions + comments + repost + post_clicks) AS total_engagement,
        created_at
    FROM posts_raw
)
SELECT
    groupArray(created_date) AS buckets,
    groupArray(total_view_time) AS total_view_time,
    groupArray(organic_view_time) AS organic_view_time,
    groupArray(paid_view_time) AS paid_view_time,
    groupArray(total_views) AS total_views,
    groupArray(organic_views) AS organic_views,
    groupArray(paid_views) AS paid_views,
    groupArray(comments) AS comments,
    groupArray(reactions) AS reactions,
    groupArray(shares) AS shares,
    groupArray(total_posts) AS total_posts
FROM (
    SELECT
        created_date,
        toInt32(count()) AS total_posts,
        toFloat64(sum(total_video_view_total_time)) AS total_view_time,
        toFloat64(sum(total_video_view_total_time_organic)) AS organic_view_time,
        toFloat64(sum(total_video_view_total_time_paid)) AS paid_view_time,
        toInt32(sum(total_video_views)) AS total_views,
        toInt32(sum(total_video_views_organic)) AS organic_views,
        toInt32(sum(total_video_views_paid)) AS paid_views,
        toInt32(sum(comments)) AS comments,
        toInt32(sum(reactions)) AS reactions,
        toInt32(sum(repost)) AS shares
    FROM (
        SELECT
            post_id,
            last_value(toDate(created_time)) AS created_date,
            last_value(total_video_view_total_time / 1000) AS total_video_view_total_time,
            last_value(total_video_view_total_time_organic / 1000) AS total_video_view_total_time_organic,
            last_value(total_video_view_total_time_paid / 1000) AS total_video_view_total_time_paid,
            last_value(total_video_views) AS total_video_views,
            last_value(total_video_views_organic) AS total_video_views_organic,
            last_value(total_video_views_paid) AS total_video_views_paid
        FROM facebook_video_insights
        WHERE toString(post_id) IN (SELECT toString(post_id) FROM posts)
        GROUP BY post_id
    ) AS videos
    LEFT JOIN posts ON posts.post_id = videos.post_id
    GROUP BY videos.created_date
    ORDER BY created_date ASC
    %s
)`, accountFilter("page_id", params), dateTimeFilter("created_time", params), fillTo(params))

	var result VideoInsightsResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Buckets,
		&result.TotalViewTime,
		&result.OrganicViewTime,
		&result.PaidViewTime,
		&result.TotalViews,
		&result.OrganicViews,
		&result.PaidViews,
		&result.Comments,
		&result.Reactions,
		&result.Shares,
		&result.TotalPosts,
	); err != nil {
		return nil, fmt.Errorf("GetVideoInsights: %w", err)
	}

	return &result, nil
}

// GetVideoRollup returns aggregated video totals for a date range.
// Used for the previous period comparison; the current period rollup is derived in the
// service layer from the time-series result to avoid a redundant DB round-trip.
func (r *Repository) GetVideoRollup(ctx context.Context, params *ch.QueryParams) (*VideoRollupResult, error) {
	query := fmt.Sprintf(`
WITH posts_raw AS (
    SELECT
        post_id,
        last_value(total) AS reactions,
        last_value(comments) AS comments,
        last_value(shares) AS repost,
        last_value(post_clicks) AS post_clicks,
        toDate(created_time) AS created_at
    FROM facebook_posts
    WHERE %s AND %s
    GROUP BY post_id, created_at
),
posts AS (
    SELECT
        post_id,
        reactions,
        comments,
        repost,
        post_clicks,
        toInt32(reactions + comments + repost + post_clicks) AS total_engagement,
        created_at
    FROM posts_raw
)
SELECT
    toFloat64(sum(total_video_view_total_time)) AS total_view_time,
    toFloat64(sum(total_video_view_total_time_organic)) AS organic_view_time,
    toFloat64(sum(total_video_view_total_time_paid)) AS paid_view_time,
    toInt32(sum(total_video_views)) AS total_views,
    toInt32(sum(total_video_views_organic)) AS organic_views,
    toInt32(sum(total_video_views_paid)) AS paid_views,
    toInt32(sum(comments)) AS comments,
    toInt32(sum(reactions)) AS reactions,
    toInt32(sum(repost)) AS shares,
    toInt32(count()) AS total_posts
FROM (
    SELECT
        post_id,
        last_value(total_video_view_total_time / 1000) AS total_video_view_total_time,
        last_value(total_video_view_total_time_organic / 1000) AS total_video_view_total_time_organic,
        last_value(total_video_view_total_time_paid / 1000) AS total_video_view_total_time_paid,
        last_value(total_video_views) AS total_video_views,
        last_value(total_video_views_organic) AS total_video_views_organic,
        last_value(total_video_views_paid) AS total_video_views_paid
    FROM facebook_video_insights
    WHERE toString(post_id) IN (SELECT toString(post_id) FROM posts)
    GROUP BY post_id
) AS videos
LEFT JOIN posts ON posts.post_id = videos.post_id
`, accountFilter("page_id", params), dateTimeFilter("created_time", params))

	var result VideoRollupResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalViewTime,
		&result.OrganicViewTime,
		&result.PaidViewTime,
		&result.TotalViews,
		&result.OrganicViews,
		&result.PaidViews,
		&result.Comments,
		&result.Reactions,
		&result.Shares,
		&result.TotalPosts,
	); err != nil {
		return nil, fmt.Errorf("GetVideoRollup: %w", err)
	}

	return &result, nil
}

// GetAudienceGender returns the fan count split by gender from facebook_insights.
// page_fans_gender is stored as a '$'-delimited string; splitByChar + arrayJoin extracts
// the gender/value pairs before CASE WHEN pivots them into M/F/U columns.
// Mirrors PHP: FacebookAnalyticsBuilder::getAudienceGenderData()
func (r *Repository) GetAudienceGender(ctx context.Context, params *ch.QueryParams) (*AudienceGenderResult, error) {
	query := fmt.Sprintf(`
SELECT
    toInt32(max(page_fans)) AS fans,
    MAX(CASE WHEN gender = 'M' THEN toInt32(value) ELSE 0 END) AS M,
    MAX(CASE WHEN gender = 'U' THEN toInt32(value) ELSE 0 END) AS F,
    MAX(CASE WHEN gender = 'F' THEN toInt32(value) ELSE 0 END) AS U
FROM (
    SELECT
        arrayJoin(arrayMap(x -> x[1], gender_pair)) AS gender,
        arrayJoin(arrayMap(x -> x[2], gender_pair)) AS value,
        page_fans
    FROM (
        SELECT
            arrayMap(x -> splitByChar('$', x), page_fans_gender) AS gender_pair,
            page_fans
        FROM (
            SELECT
                max(page_fans_gender) AS page_fans_gender,
                max(page_fans) AS page_fans
            FROM facebook_insights
            WHERE %s AND %s
            GROUP BY hash_id
        )
    )
)`, accountFilter("page_id", params), dateOnlyFilter("created_time", params))

	var result AudienceGenderResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Fans,
		&result.M,
		&result.F,
		&result.U,
	); err != nil {
		return nil, fmt.Errorf("GetAudienceGender: %w", err)
	}

	return &result, nil
}

// GetMaxGenderAge returns the gender+age bucket with the highest fan count.
// page_fans_gender_age is a '$'-delimited string; arrayMax finds the peak value and
// indexOf locates the corresponding label, which is then split on '_' to get gender/age.
// Mirrors PHP: FacebookAnalyticsBuilder::getMaxGenderAgeData()
func (r *Repository) GetMaxGenderAge(ctx context.Context, params *ch.QueryParams) (*MaxGenderAgeResult, error) {
	query := fmt.Sprintf(`
SELECT
    arrayMax(arrayMap(x -> toInt32(x[2]), gender_age_pair)) AS max_value,
    arrayMap(x -> splitByChar('_', x), arrayMap(x -> x[1], gender_age_pair))[indexOf(arrayMap(x -> toInt32(x[2]), gender_age_pair), max_value)][2] AS age,
    arrayMap(x -> splitByChar('_', x), arrayMap(x -> x[1], gender_age_pair))[indexOf(arrayMap(x -> toInt32(x[2]), gender_age_pair), max_value)][1] AS gender
FROM (
    SELECT arrayMap(x -> splitByChar('$', x), page_fans_gender_age) AS gender_age_pair
    FROM (
        SELECT max(page_fans_gender_age) AS page_fans_gender_age
        FROM facebook_insights
        WHERE %s AND %s
        GROUP BY hash_id
    )
)
ORDER BY max_value DESC
LIMIT 1`, accountFilter("page_id", params), dateOnlyFilter("created_time", params))

	var result MaxGenderAgeResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.MaxValue,
		&result.Age,
		&result.Gender,
	); err != nil {
		return nil, fmt.Errorf("GetMaxGenderAge: %w", err)
	}

	return &result, nil
}

// GetAudienceAge returns fan counts broken down by age bracket from facebook_insights.
// page_fans_age is a '$'-delimited string; ARRAY JOIN + CASE WHEN pivots values into
// age-specific columns. Takes the most recent row (ORDER BY DESC LIMIT 1) for accuracy.
// Mirrors PHP: FacebookAnalyticsBuilder::getAudienceAgeData()
func (r *Repository) GetAudienceAge(ctx context.Context, params *ch.QueryParams) (*AudienceAgeResult, error) {
	query := fmt.Sprintf(`
SELECT
    max(CASE WHEN ages = '65+' THEN ages_count[1] ELSE 0 END) AS "65+",
    max(CASE WHEN ages = '55-64' THEN ages_count[2] ELSE 0 END) AS "55-64",
    max(CASE WHEN ages = '45-54' THEN ages_count[3] ELSE 0 END) AS "45-54",
    max(CASE WHEN ages = '35-44' THEN ages_count[4] ELSE 0 END) AS "35-44",
    max(CASE WHEN ages = '25-34' THEN ages_count[5] ELSE 0 END) AS "25-34",
    max(CASE WHEN ages = '18-24' THEN ages_count[6] ELSE 0 END) AS "18-34",
    max(CASE WHEN ages = '13-17' THEN ages_count[7] ELSE 0 END) AS "13-17"
FROM (
    SELECT
        arrayMap(x -> x[1], age_pair) AS ages,
        arrayMap(x -> toInt32(x[2]), age_pair) AS ages_count
    FROM (
        SELECT arrayMap(x -> splitByChar('$', x), page_fans_age) AS age_pair
        FROM (
            SELECT page_fans_age, saving_time
            FROM (
                SELECT
                    hash_id,
                    last_value(page_fans_age) AS page_fans_age,
                    last_value(created_time) AS saving_time,
                    last_value(page_id) AS page_id
                FROM facebook_insights
                GROUP BY hash_id
            )
            WHERE %s AND %s
            ORDER BY saving_time DESC
            LIMIT 1
        )
    )
) ARRAY JOIN ages`, accountFilter("page_id", params), dateOnlyFilter("saving_time", params))

	var result AudienceAgeResult
	if err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Age65Plus,
		&result.Age55To64,
		&result.Age45To54,
		&result.Age35To44,
		&result.Age25To34,
		&result.Age18To34,
		&result.Age13To17,
	); err != nil {
		return nil, fmt.Errorf("GetAudienceAge: %w", err)
	}

	return &result, nil
}

// GetAudienceCountry returns fan counts by country from facebook_insights.
// page_fans_country is a '$'-delimited string decoded via splitByChar + ARRAY JOIN.
// Uses the most recent available record and returns rows ordered by count descending.
// Mirrors PHP: FacebookAnalyticsBuilder::getAudienceCountryData()
func (r *Repository) GetAudienceCountry(ctx context.Context, params *ch.QueryParams) (map[string]int32, error) {
	query := fmt.Sprintf(`
SELECT
    countries,
    country_values
FROM (
    SELECT
        arrayMap(x -> x[1], page_fans_country_pair) AS countries,
        arrayMap(x -> toInt32(x[2]), page_fans_country_pair) AS country_values
    FROM (
        SELECT arrayMap(x -> splitByChar('$', x), page_fans_country) AS page_fans_country_pair
        FROM (
            SELECT page_fans_country
            FROM (
                SELECT last_value(page_fans_country) AS page_fans_country, created_time
                FROM facebook_insights
                WHERE %s AND %s
                GROUP BY created_time
                ORDER BY created_time DESC
                LIMIT 1
            )
        )
    )
) ARRAY JOIN countries AS countries, country_values AS country_values
ORDER BY country_values DESC`, accountFilter("page_id", params), dateOnlyFilter("created_time", params))

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetAudienceCountry: %w", err)
	}
	defer rows.Close()

	out := make(map[string]int32)
	for rows.Next() {
		var key string
		var value int32
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("GetAudienceCountry: %w", err)
		}
		out[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAudienceCountry: %w", err)
	}
	return out, nil
}

// GetAudienceCity returns fan counts by city from facebook_insights.
// page_fans_city is a '$'-delimited string decoded via splitByChar + ARRAY JOIN.
// Mirrors PHP: FacebookAnalyticsBuilder::getAudienceCityData()
func (r *Repository) GetAudienceCity(ctx context.Context, params *ch.QueryParams) (map[string]int32, error) {
	query := fmt.Sprintf(`
SELECT
    cities,
    city_values
FROM (
    SELECT
        arrayMap(x -> x[1], page_fans_city_pair) AS cities,
        arrayMap(x -> toInt32(x[2]), page_fans_city_pair) AS city_values
    FROM (
        SELECT arrayMap(x -> splitByChar('$', x), page_fans_city) AS page_fans_city_pair
        FROM (
            SELECT page_fans_city
            FROM (
                SELECT max(page_fans_city) AS page_fans_city
                FROM facebook_insights
                WHERE %s AND %s
                GROUP BY hash_id
            )
            ORDER BY page_fans_city DESC
            LIMIT 1
        )
    )
) ARRAY JOIN cities AS cities, city_values AS city_values
ORDER BY city_values DESC`, accountFilter("page_id", params), dateOnlyFilter("created_time", params))

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetAudienceCity: %w", err)
	}
	defer rows.Close()

	out := make(map[string]int32)
	for rows.Next() {
		var key string
		var value int32
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("GetAudienceCity: %w", err)
		}
		out[key] = value
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetAudienceCity: %w", err)
	}
	return out, nil
}

// accountFilter produces an IN clause that casts the field to string before comparing.
// Facebook stores page_id as a numeric type; toString() ensures consistent matching.
func accountFilter(field string, params *ch.QueryParams) string {
	return fmt.Sprintf("toString(%s) IN %s", field, ch.FormatAccountIDs(params.AccountIDs))
}

// dateTimeFilter produces a BETWEEN clause that converts the field to a Date in the
// request timezone before comparing. Using toDate on both sides avoids the
// UTC-vs-local-midnight mismatch that occurred when the field was converted to a
// local DateTime but the boundary strings were parsed as UTC.
func dateTimeFilter(field string, params *ch.QueryParams) string {
	return fmt.Sprintf(
		"toDate(%s, '%s') BETWEEN '%s' AND '%s'",
		field,
		params.Timezone,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)
}

// dateOnlyFilter produces a date-only BETWEEN clause without timezone conversion.
// Used for fields that store dates rather than timestamps.
func dateOnlyFilter(field string, params *ch.QueryParams) string {
	return fmt.Sprintf(
		"toDate(%s) BETWEEN toDate('%s') AND toDate('%s')",
		field,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)
}

// fillTo produces a "WITH FILL TO date+1 STEP 1" clause for ORDER BY date columns.
// The +1 ensures the last day of the range is included.
func fillTo(params *ch.QueryParams) string {
	return fmt.Sprintf("WITH FILL TO toDate('%s') + 1 STEP 1", params.DateTo.Format("2006-01-02"))
}

// fillFromTo produces a "WITH FILL FROM date TO date+1 STEP 1" clause.
// Used in audience growth to guarantee the full date range is present even with sparse data.
func fillFromTo(params *ch.QueryParams) string {
	return fmt.Sprintf(
		"WITH FILL FROM toDate('%s') TO toDate('%s') + 1 STEP 1",
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)
}

// maxWeeks returns the number of complete (or partial) weeks in dayCount.
// Used to compute weekly averages; always returns at least 1 to avoid division by zero.
func maxWeeks(dayCount int) int {
	weeks := dayCount / 7
	if dayCount%7 != 0 {
		weeks++
	}
	if weeks <= 0 {
		return 1
	}
	return weeks
}

// sanitizeOrderBy strips backtick, double-quote, and single-quote characters from an
// ORDER BY column name to prevent SQL injection via user-supplied sort fields.
func sanitizeOrderBy(orderBy string) string {
	replacer := strings.NewReplacer("`", "", "\"", "", "'", "")
	return replacer.Replace(orderBy)
}
