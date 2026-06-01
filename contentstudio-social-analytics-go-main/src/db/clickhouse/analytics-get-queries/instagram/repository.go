package instagram

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// Repository executes ClickHouse queries for Instagram analytics.
type Repository struct {
	client *ch.Client
}

// NewRepository returns a new Repository backed by the given ClickHouse client.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// igPostDedupCTE mirrors PHP InstagramBuilder's posts CTE: it selects the latest
// stored_event_at per media_id so that the main query can use
// WHERE (media_id, stored_event_at) IN (SELECT media_id, max_event FROM posts)
// to read exactly the latest snapshot row per post directly from instagram_posts.
// This avoids the argMax-alias-in-WHERE and nested-aggregate errors that occur
// when all columns are pre-aggregated inside the CTE.
func igPostDedupCTE(ids, dateFilter, partitionFilter string, extraFilters ...string) string {
	extra := ""
	if len(extraFilters) > 0 {
		extra = " AND " + strings.Join(extraFilters, " AND ")
	}
	return fmt.Sprintf(`posts AS (
		SELECT media_id, max(stored_event_at) AS max_event
		FROM instagram_posts
		WHERE instagram_id IN %s
		  AND %s
		  AND %s%s
		GROUP BY media_id
	)`, ids, dateFilter, partitionFilter, extra)
}

// igDeduped is the WHERE clause fragment that restricts to the latest snapshot
// row per post, matched against the posts CTE produced by igPostDedupCTE.
const igDeduped = "(media_id, stored_event_at) IN (SELECT media_id, max_event FROM posts)"

// GetPostsSummary returns aggregated post-level metrics from instagram_posts for the given period.
// Run concurrently with GetInsightsSummary for maximum throughput.
func (r *Repository) GetPostsSummary(ctx context.Context, params *ch.QueryParams) (*PostsSummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			toInt64(count())                                                              AS doc_count,
			toInt64(sum(engagement))                                                       AS total_engagement,
			toInt64(sum(like_count))                                                       AS likes,
			toInt64(sum(comments_count))                                                   AS comments,
			toInt64(sum(saved))                                                            AS saved,
			toInt64(sum(reach))                                                            AS reach,
			toInt64(sum(impressions))                                                      AS impressions,
			toInt64(sum(views))                                                            AS views,
			toInt64(countIf(entity_type = 'STORY' OR media_type = 'STORY'))               AS stories,
			toInt64(countIf(entity_type != 'STORY' AND media_type != 'STORY'))            AS total_posts,
			%s                                                                             AS engagement_rate
		FROM instagram_posts
		WHERE %s AND instagram_id IN %s AND %s`,
		igPostDedupCTE(ids, dateFilter, partFilter),
		ch.SafeRate("sum(engagement)", "count()", 2),
		igDeduped, ids, partFilter,
	)

	var result PostsSummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.DocCount,
		&result.TotalEngagement,
		&result.Likes,
		&result.Comments,
		&result.Saved,
		&result.Reach,
		&result.Impressions,
		&result.Views,
		&result.Stories,
		&result.TotalPosts,
		&result.EngagementRate,
	)
	return &result, err
}

// GetInsightsSummary returns aggregated account-level metrics from instagram_insights for the given period.
// Run concurrently with GetPostsSummary for maximum throughput.
func (r *Repository) GetInsightsSummary(ctx context.Context, params *ch.QueryParams) (*InsightsSummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_time", params)
	partFilter := ch.PartitionMonthFilter("created_time", params)

	query := fmt.Sprintf(`
		SELECT
			toInt64(sum(daily_profile_views))                                           AS profile_views,
			toInt64(argMaxIf(daily_followers_count, date_bucket, daily_followers_count > 0)) AS followers_count,
			toInt64(argMaxIf(daily_follows_count, date_bucket, daily_follows_count > 0))   AS follows_count,
			toInt64(sum(daily_accounts_engaged))                                         AS accounts_engaged,
			toInt64(sum(daily_engagement))                                               AS engagement,
			toInt64(sum(daily_impressions))                                              AS impressions,
			toInt64(sum(daily_reach))                                                    AS reach
		FROM (
			SELECT
				toDate(created_time) AS date_bucket,
				max(profile_views)  AS daily_profile_views,
				argMin(followers_count, stored_event_at) AS daily_followers_count,
				argMin(follows_count, stored_event_at)  AS daily_follows_count,
				max(accounts_engaged) AS daily_accounts_engaged,
				max(engagement)     AS daily_engagement,
				max(impressions)    AS daily_impressions,
				max(reach)          AS daily_reach
			FROM instagram_insights
			WHERE instagram_id IN %s
			  AND %s
			  AND %s
			GROUP BY date_bucket
		)`,
		ids, dateFilter, partFilter,
	)

	var result InsightsSummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ProfileViews,
		&result.FollowersCount,
		&result.FollowsCount,
		&result.AccountsEngaged,
		&result.Engagement,
		&result.Impressions,
		&result.Reach,
	)
	return &result, err
}

// GetAudienceGrowth returns time-series follower data for the given period.
func (r *Repository) GetAudienceGrowth(ctx context.Context, params *ch.QueryParams) (*AudienceResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_time", params)
	partFilter := ch.PartitionMonthFilter("created_time", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0)                                          AS show_data,
			groupArray(followers_count)                                     AS followers,
			groupArray(0)                                                   AS followers_daily,
			groupArray(bucket)                                              AS buckets
		FROM (
			SELECT
				toDate(created_time, '%s')                                  AS bucket,
				toInt32(argMax(followers_count, stored_event_at))           AS followers_count
			FROM instagram_insights
			WHERE instagram_id IN %s
			  AND %s
			  AND %s
			GROUP BY bucket
			ORDER BY bucket ASC %s
		)`,
		params.Timezone, ids, dateFilter, partFilter, fill,
	)

	var result AudienceResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.Followers,
		&result.FollowersDaily,
		&result.Buckets,
	)
	return &result, err
}

// GetLastFollowerCount returns the most recent non-zero follower count within the given date window.
func (r *Repository) GetLastFollowerCount(ctx context.Context, params *ch.QueryParams) (*FollowerCount, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_time", params)
	partFilter := ch.PartitionMonthFilter("created_time", params)

	query := fmt.Sprintf(`
		SELECT toInt32(followers_count) AS followers_count
		FROM instagram_insights
		WHERE instagram_id IN %s
		  AND %s
		  AND %s
		  AND followers_count > 0
		ORDER BY created_time DESC
		LIMIT 1`,
		ids, dateFilter, partFilter,
	)

	var result FollowerCount
	err := r.client.Conn.QueryRow(ctx, query).Scan(&result.FollowersCount)
	return &result, err
}

// GetAudienceRollup returns aggregated follower metrics for current or previous period.
func (r *Repository) GetAudienceRollup(ctx context.Context, params *ch.QueryParams) (*AudienceRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_time", params)
	partFilter := ch.PartitionMonthFilter("created_time", params)

	query := fmt.Sprintf(`
		SELECT
			toInt32(argMax(followers_count, stored_event_at))       AS follower_count,
			toInt32(max(followers_count) - min(followers_count))    AS follower_gained
		FROM instagram_insights
		WHERE instagram_id IN %s
		  AND %s
		  AND %s`,
		ids, dateFilter, partFilter,
	)

	var result AudienceRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.FollowerCount,
		&result.FollowerGained,
	)
	return &result, err
}

// GetPublishingBehaviour returns time-series publishing metrics filtered by media types.
func (r *Repository) GetPublishingBehaviour(ctx context.Context, params *ch.QueryParams, mediaTypes []string) (*PublishingResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)
	mediaFilter := fmt.Sprintf("media_type IN (%s)", ch.FormatStringList(mediaTypes))
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			groupArray(likes)        AS likes,
			groupArray(comments)     AS comments,
			groupArray(saved)        AS saved,
			groupArray(engagement)   AS engagement,
			groupArray(reach)        AS reach,
			groupArray(impressions)  AS impressions,
			groupArray(views)        AS views,
			groupArray(total_posts)  AS total_posts,
			groupArray(bucket)       AS buckets
		FROM (
			SELECT
				toDate(post_created_at, '%s')   AS bucket,
				toInt32(sum(like_count))        AS likes,
				toInt32(sum(comments_count))    AS comments,
				toInt32(sum(saved))             AS saved,
				toInt32(sum(engagement))        AS engagement,
				toInt32(sum(reach))             AS reach,
				toInt32(sum(impressions))       AS impressions,
				toInt32(sum(views))             AS views,
				toInt32(count())                AS total_posts
			FROM instagram_posts
			WHERE %s AND instagram_id IN %s AND %s
			GROUP BY bucket
			ORDER BY bucket ASC %s
		)`,
		igPostDedupCTE(ids, dateFilter, partFilter, mediaFilter),
		params.Timezone,
		igDeduped, ids, partFilter,
		fill,
	)

	var result PublishingResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Likes,
		&result.Comments,
		&result.Saved,
		&result.Engagement,
		&result.Reach,
		&result.Impressions,
		&result.Views,
		&result.TotalPosts,
		&result.Buckets,
	)
	return &result, err
}

// GetPublishingBehaviourRollup returns aggregated publishing metrics broken down by media type.
func (r *Repository) GetPublishingBehaviourRollup(ctx context.Context, params *ch.QueryParams, mediaTypes []string) ([]PublishingRollupRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)
	mediaFilter := fmt.Sprintf("media_type IN (%s)", ch.FormatStringList(mediaTypes))

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			media_type,
			toInt32(count())             AS total_posts,
			toInt32(sum(like_count))     AS likes,
			toInt32(sum(comments_count)) AS comments,
			toInt32(sum(saved))          AS saved,
			toInt32(sum(engagement))     AS engagement,
			toInt32(sum(reach))          AS reach,
			toInt32(sum(views))          AS views
		FROM instagram_posts
		WHERE %s AND instagram_id IN %s AND %s
		GROUP BY media_type
		ORDER BY total_posts DESC`,
		igPostDedupCTE(ids, dateFilter, partFilter, mediaFilter),
		igDeduped, ids, partFilter,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []PublishingRollupRow
	for rows.Next() {
		var row PublishingRollupRow
		if err := rows.Scan(
			&row.MediaType,
			&row.TotalPosts,
			&row.Likes,
			&row.Comments,
			&row.Saved,
			&row.Engagement,
			&row.Reach,
			&row.Views,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// GetTopPosts returns posts sorted by the given metric, limited to limit rows.
func (r *Repository) GetTopPosts(ctx context.Context, params *ch.QueryParams, orderBy string, limit int, hashtags []string) ([]TopPostResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)

	extraFilters := []string{"entity_type != 'STORY'"}
	if len(hashtags) > 0 {
		extraFilters = append(extraFilters, fmt.Sprintf("hasAny(hashtags, [%s])", ch.FormatStringList(hashtags)))
	}

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			instagram_id,
			media_id,
			caption,
			media_type,
			entity_type,
			media_url,
			video_url,
			permalink,
			like_count,
			comments_count,
			saved,
			like_count + comments_count + saved  AS total_engagement,
			reach,
			impressions,
			views,
			shares,
			reels_avg_watch_time,
			reels_total_watch_time,
			exits,
			replies,
			hashtags,
			day_of_week,
			hour_of_day,
			post_created_at,
			stored_event_at
		FROM instagram_posts
		WHERE %s AND instagram_id IN %s AND %s
		ORDER BY %s DESC
		LIMIT %d`,
		igPostDedupCTE(ids, dateFilter, partFilter, extraFilters...),
		igDeduped, ids, partFilter,
		orderBy, limit,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TopPostResult
	for rows.Next() {
		var row TopPostResult
		if err := rows.Scan(
			&row.InstagramID,
			&row.MediaID,
			&row.Caption,
			&row.MediaType,
			&row.EntityType,
			&row.MediaURL,
			&row.VideoURL,
			&row.Permalink,
			&row.LikeCount,
			&row.CommentsCount,
			&row.Saved,
			&row.Engagement,
			&row.Reach,
			&row.Impressions,
			&row.Views,
			&row.Shares,
			&row.ReelsAvgWatchTime,
			&row.ReelsTotalWatchTime,
			&row.Exits,
			&row.Replies,
			&row.Hashtags,
			&row.DayOfWeek,
			&row.HourOfDay,
			&row.PostCreatedAt,
			&row.StoredEventAt,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// GetActiveUsersHours returns the hourly distribution of online followers.
func (r *Repository) GetActiveUsersHours(ctx context.Context, params *ch.QueryParams) (*ActiveUsersHoursResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("online_users_datetime", params)
	partFilter := ch.PartitionMonthFilter("online_users_datetime", params)

	query := fmt.Sprintf(`
		SELECT
			groupArray(toInt32(number))        AS buckets,
			groupArray(toInt32(avg_value))     AS values,
			toInt32(max(avg_value))            AS highest_value,
			toInt32(argMax(number, avg_value)) AS highest_hour
		FROM (
			SELECT
				number,
				toInt32(avg(toInt32(splitByChar(':', json_str)[2]))) AS avg_value
			FROM numbers(24) AS n
			CROSS JOIN (
				SELECT arrayJoin(max_online_followers) AS json_str
				FROM (
					SELECT max(online_followers) AS max_online_followers
					FROM instagram_insights
					WHERE instagram_id IN %s
					  AND %s
					  AND %s
					  AND notEmpty(online_followers)
					GROUP BY record_id
				)
			) AS followers_data
			WHERE toInt32(splitByChar(':', json_str)[1]) = toInt32(number)
			GROUP BY number
			ORDER BY number ASC
		)`,
		ids, dateFilter, partFilter,
	)

	var result ActiveUsersHoursResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Buckets,
		&result.Values,
		&result.HighestValue,
		&result.HighestHour,
	)
	return &result, err
}

// GetActiveUsersDays returns the day-of-week distribution of online followers.
func (r *Repository) GetActiveUsersDays(ctx context.Context, params *ch.QueryParams) (*ActiveUsersDaysResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_time", params)
	partFilter := ch.PartitionMonthFilter("created_time", params)

	query := fmt.Sprintf(`
		SELECT
			groupArray(dow)                AS buckets,
			groupArray(toInt32(max_value)) AS values,
			toInt32(max(max_value))        AS highest_value,
			argMax(dow, max_value)         AS highest_day
		FROM (
			SELECT
				dow,
				max(toInt32(arraySum(arrayMap((x) -> toInt32(splitByChar(':', x)[2]), followers_data)))) AS max_value
			FROM (
				SELECT
					max(day_of_week)      AS dow,
					max(online_followers) AS followers_data
				FROM instagram_insights
				WHERE instagram_id IN %s
				  AND %s
				  AND %s
				  AND notEmpty(online_followers)
				GROUP BY record_id
			)
			WHERE dow != ''
			GROUP BY dow
			ORDER BY
				CASE
					WHEN dow = 'Monday'    THEN 1
					WHEN dow = 'Tuesday'   THEN 2
					WHEN dow = 'Wednesday' THEN 3
					WHEN dow = 'Thursday'  THEN 4
					WHEN dow = 'Friday'    THEN 5
					WHEN dow = 'Saturday'  THEN 6
					WHEN dow = 'Sunday'    THEN 7
				END
		)`,
		ids, dateFilter, partFilter,
	)

	var result ActiveUsersDaysResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Buckets,
		&result.Values,
		&result.HighestValue,
		&result.HighestDay,
	)
	return &result, err
}

// GetImpressions returns time-series impressions data from instagram_insights.
func (r *Repository) GetImpressions(ctx context.Context, params *ch.QueryParams) (*ImpressionsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_time", params)
	partFilter := ch.PartitionMonthFilter("created_time", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
		SELECT
			toUInt8(if(count() > 0, 1, 0))           AS show_data,
			groupArray(bucket)                        AS buckets,
			groupArray(toInt32(impressions))          AS impressions
		FROM (
			SELECT
				toDate(created_time, '%s')                         AS bucket,
				argMax(impressions, stored_event_at)               AS impressions
			FROM instagram_insights
			WHERE instagram_id IN %s
			  AND %s
			  AND %s
			GROUP BY bucket
			ORDER BY bucket ASC %s
		)`,
		params.Timezone, ids, dateFilter, partFilter, fill,
	)

	var result ImpressionsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.Buckets,
		&result.Impressions,
	)
	return &result, err
}

// GetImpressionsRollup returns aggregated impressions totals.
func (r *Repository) GetImpressionsRollup(ctx context.Context, params *ch.QueryParams) (*ImpressionsRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_time", params)
	partFilter := ch.PartitionMonthFilter("created_time", params)

	query := fmt.Sprintf(`
		SELECT
			sum(impressions)   AS total_impressions,
			%s                 AS avg_impressions
		FROM (
			SELECT
				toDate(created_time, '%s')               AS bucket,
				argMax(impressions, stored_event_at)     AS impressions
			FROM instagram_insights
			WHERE instagram_id IN %s
			  AND %s
			  AND %s
			GROUP BY bucket
		)`,
		ch.SafeRate("sum(impressions)", "count()", 2),
		params.Timezone, ids, dateFilter, partFilter,
	)

	var result ImpressionsRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalImpressions,
		&result.AvgImpressions,
	)
	return &result, err
}

// GetEngagement returns time-series engagement data from instagram_posts.
func (r *Repository) GetEngagement(ctx context.Context, params *ch.QueryParams) (*EngagementResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			toUInt8(if(count() > 0, 1, 0))   AS show_data,
			groupArray(bucket)               AS buckets,
			groupArray(toInt32(engagement))  AS engagement,
			groupArray(toInt32(comments))    AS comments,
			groupArray(toInt32(likes))       AS reactions,
			groupArray(toInt32(doc_count))   AS doc_count
		FROM (
			SELECT
				toDate(post_created_at, '%s')   AS bucket,
				sum(engagement)                 AS engagement,
				sum(comments_count)             AS comments,
				sum(like_count)                 AS likes,
				count()                         AS doc_count
			FROM instagram_posts
			WHERE %s AND instagram_id IN %s AND %s AND entity_type != 'STORY'
			GROUP BY bucket
			ORDER BY bucket ASC %s
		)`,
		igPostDedupCTE(ids, dateFilter, partFilter),
		params.Timezone,
		igDeduped, ids, partFilter,
		fill,
	)

	var result EngagementResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.Buckets,
		&result.Engagement,
		&result.Comments,
		&result.Reactions,
		&result.DocCount,
	)
	return &result, err
}

// GetEngagementRollup returns aggregated engagement totals.
func (r *Repository) GetEngagementRollup(ctx context.Context, params *ch.QueryParams) (*EngagementRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			sum(engagement)     AS engagement,
			%s                  AS avg_engagement,
			sum(comments_count) AS comments,
			sum(like_count)     AS reactions,
			sum(saved)          AS saved,
			count()             AS count
		FROM instagram_posts
		WHERE %s AND instagram_id IN %s AND %s AND entity_type != 'STORY'`,
		igPostDedupCTE(ids, dateFilter, partFilter),
		ch.SafeRate("sum(engagement)", "count()", 2),
		igDeduped, ids, partFilter,
	)

	var result EngagementRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Engagement,
		&result.AvgEngagement,
		&result.Comments,
		&result.Reactions,
		&result.Saved,
		&result.Count,
	)
	return &result, err
}

// GetTopHashtags returns the top 30 hashtags with their engagement metrics.
func (r *Repository) GetTopHashtags(ctx context.Context, params *ch.QueryParams) (*HashtagsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			groupArray(hashtag)           AS name,
			groupArray(toInt32(eng))      AS engagement,
			groupArray(toInt32(lk))       AS likes,
			groupArray(toInt32(cmt))      AS comments,
			groupArray(toInt32(sv))       AS saved,
			groupArray(toInt32(cnt))      AS posts
		FROM (
			SELECT
				hashtag,
				sum(engagement)      AS eng,
				sum(like_count)      AS lk,
				sum(comments_count)  AS cmt,
				sum(saved)           AS sv,
				count()              AS cnt
			FROM instagram_posts
			ARRAY JOIN hashtags AS hashtag
			WHERE %s AND instagram_id IN %s AND %s AND entity_type != 'STORY'
			GROUP BY hashtag
			ORDER BY eng DESC
			LIMIT 30
		)`,
		igPostDedupCTE(ids, dateFilter, partFilter),
		igDeduped, ids, partFilter,
	)

	var result HashtagsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Name,
		&result.Engagement,
		&result.Likes,
		&result.Comments,
		&result.Saved,
		&result.Posts,
	)
	return &result, err
}

// GetTopHashtagsRollup returns aggregated hashtag totals.
func (r *Repository) GetTopHashtagsRollup(ctx context.Context, params *ch.QueryParams) (*HashtagsRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)

	query := fmt.Sprintf(`
		WITH %s,
		hashtag_stats AS (
			SELECT
				hashtag,
				sum(engagement)      AS eng,
				sum(like_count)      AS lk,
				sum(comments_count)  AS cmt,
				sum(saved)           AS sv,
				count()              AS cnt
			FROM instagram_posts
			ARRAY JOIN hashtags AS hashtag
			WHERE %s AND instagram_id IN %s AND %s AND entity_type != 'STORY'
			GROUP BY hashtag
		)
		SELECT
			toInt32(sum(eng))   AS total_engagement,
			toInt32(sum(lk))    AS total_likes,
			toInt32(sum(cmt))   AS total_comments,
			toInt32(sum(sv))    AS total_saves,
			toInt32(count())    AS total_unique_hashtags,
			toInt32(sum(cnt))   AS total_hashtag_uses
		FROM hashtag_stats`,
		igPostDedupCTE(ids, dateFilter, partFilter),
		igDeduped, ids, partFilter,
	)

	var result HashtagsRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalEngagement,
		&result.TotalLikes,
		&result.TotalComments,
		&result.TotalSaves,
		&result.TotalUniqueHashtags,
		&result.TotalHashtagUses,
	)
	return &result, err
}

// GetStoriesPerformance returns time-series stories metrics.
func (r *Repository) GetStoriesPerformance(ctx context.Context, params *ch.QueryParams) (*StoriesResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			toUInt8(if(count() > 0, 1, 0))             AS show_data,
			groupArray(bucket)                          AS buckets,
			groupArray(avg_impressions)                 AS avg_story_impressions,
			groupArray(toInt32(story_impressions))      AS story_impressions,
			groupArray(toInt32(story_reach))            AS story_reach,
			groupArray(toInt32(story_reply))            AS story_reply,
			groupArray(toInt32(story_exits))            AS story_exits,
			groupArray(toInt32(story_taps_forward))     AS story_taps_forward,
			groupArray(toInt32(story_taps_back))        AS story_taps_back,
			groupArray(toInt32(published_stories))      AS published_stories
		FROM (
			SELECT
				toDate(post_created_at, '%s')           AS bucket,
				avg(impressions)                        AS avg_impressions,
				sum(impressions)                        AS story_impressions,
				sum(reach)                              AS story_reach,
				sum(replies)                            AS story_reply,
				sum(exits)                              AS story_exits,
				sum(taps_forward)                       AS story_taps_forward,
				sum(taps_back)                          AS story_taps_back,
				count()                                 AS published_stories
			FROM instagram_posts
			WHERE %s AND instagram_id IN %s AND %s AND entity_type = 'STORY'
			GROUP BY bucket
			ORDER BY bucket ASC %s
		)`,
		igPostDedupCTE(ids, dateFilter, partFilter),
		params.Timezone,
		igDeduped, ids, partFilter,
		fill,
	)

	var result StoriesResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.Buckets,
		&result.AvgStoryImpressions,
		&result.StoryImpressions,
		&result.StoryReach,
		&result.StoryReply,
		&result.StoryExits,
		&result.StoryTapsForward,
		&result.StoryTapsBack,
		&result.PublishedStories,
	)
	return &result, err
}

// GetStoriesRollup returns aggregated stories totals.
func (r *Repository) GetStoriesRollup(ctx context.Context, params *ch.QueryParams) (*StoriesRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			sum(impressions)              AS story_impressions,
			%s                            AS avg_story_impressions,
			sum(reach)                    AS story_reach,
			sum(replies)                  AS story_reply,
			sum(exits)                    AS story_exits,
			sum(taps_forward)             AS story_taps_forward,
			sum(taps_back)                AS story_taps_back,
			toInt64(count())               AS published_stories
		FROM instagram_posts
		WHERE %s AND instagram_id IN %s AND %s AND entity_type = 'STORY'`,
		igPostDedupCTE(ids, dateFilter, partFilter),
		ch.SafeRate("sum(impressions)", "count()", 2),
		igDeduped, ids, partFilter,
	)

	var result StoriesRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.StoryImpressions,
		&result.AvgStoryImpressions,
		&result.StoryReach,
		&result.StoryReply,
		&result.StoryExits,
		&result.StoryTapsForward,
		&result.StoryTapsBack,
		&result.PublishedStories,
	)
	return &result, err
}

// GetReelsPerformance returns time-series reels metrics.
func (r *Repository) GetReelsPerformance(ctx context.Context, params *ch.QueryParams) (*ReelsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			toUInt8(if(count() > 0, 1, 0))          AS show_data,
			groupArray(bucket)                       AS buckets,
			groupArray(toInt32(total_posts))         AS total_posts,
			groupArray(toInt32(engagement))          AS engagement,
			groupArray(toInt32(likes))               AS likes,
			groupArray(toInt32(comments))            AS comments,
			groupArray(toInt32(saves))               AS saves,
			groupArray(toInt32(shares))              AS shares,
			groupArray(avg_watch_time)               AS avg_watch_time,
			groupArray(total_watch_time)             AS total_watch_time
		FROM (
			SELECT
				toDate(post_created_at, '%s')        AS bucket,
				count()                              AS total_posts,
				sum(engagement)                      AS engagement,
				sum(like_count)                      AS likes,
				sum(comments_count)                  AS comments,
				sum(saved)                           AS saves,
				sum(shares)                          AS shares,
				avg(reels_avg_watch_time)            AS avg_watch_time,
				sum(reels_total_watch_time)          AS total_watch_time
			FROM instagram_posts
			WHERE %s AND instagram_id IN %s AND %s AND media_type = 'REELS'
			GROUP BY bucket
			ORDER BY bucket ASC %s
		)`,
		igPostDedupCTE(ids, dateFilter, partFilter, "media_type = 'REELS'"),
		params.Timezone,
		igDeduped, ids, partFilter,
		fill,
	)

	var result ReelsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.Buckets,
		&result.TotalPosts,
		&result.Engagement,
		&result.Likes,
		&result.Comments,
		&result.Saves,
		&result.Shares,
		&result.AvgWatchTime,
		&result.TotalWatchTime,
	)
	return &result, err
}

// GetReelsRollup returns aggregated reels totals.
func (r *Repository) GetReelsRollup(ctx context.Context, params *ch.QueryParams) (*ReelsRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("post_created_at", params)
	partFilter := ch.PartitionMonthFilter("post_created_at", params)

	query := fmt.Sprintf(`
		WITH %s
		SELECT
			sum(engagement)              AS engagement,
			sum(like_count)              AS likes,
			sum(comments_count)          AS comments,
			sum(saved)                   AS saves,
			toInt64(count())              AS total_posts,
			sum(shares)                  AS shares,
			avg(reels_avg_watch_time)    AS avg_watch_time,
			sum(reels_total_watch_time)  AS total_watch_time
		FROM instagram_posts
		WHERE %s AND instagram_id IN %s AND %s AND media_type = 'REELS'`,
		igPostDedupCTE(ids, dateFilter, partFilter, "media_type = 'REELS'"),
		igDeduped, ids, partFilter,
	)

	var result ReelsRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.Engagement,
		&result.Likes,
		&result.Comments,
		&result.Saves,
		&result.TotalPosts,
		&result.Shares,
		&result.AvgWatchTime,
		&result.TotalWatchTime,
	)
	return &result, err
}

// GetDemographics returns the latest audience demographic arrays.
func (r *Repository) GetDemographics(ctx context.Context, params *ch.QueryParams) (*DemographicsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)

	query := fmt.Sprintf(`
		SELECT
			audience_age,
			audience_gender,
			audience_gender_age
		FROM instagram_insights
		WHERE instagram_id IN %s
		  AND notEmpty(audience_age)
		ORDER BY stored_event_at DESC
		LIMIT 1`,
		ids,
	)

	var result DemographicsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.AudienceAge,
		&result.AudienceGender,
		&result.AudienceGenderAge,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return &result, nil
	}
	return &result, err
}

// GetLocations returns the latest audience location arrays.
func (r *Repository) GetLocations(ctx context.Context, params *ch.QueryParams) (*LocationResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)

	query := fmt.Sprintf(`
		SELECT
			audience_city,
			audience_country
		FROM instagram_insights
		WHERE instagram_id IN %s
		  AND notEmpty(audience_city)
		ORDER BY stored_event_at DESC
		LIMIT 1`,
		ids,
	)

	var result LocationResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.AudienceCity,
		&result.AudienceCountry,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return &result, nil
	}
	return &result, err
}
