package tiktok

import (
	"context"
	"fmt"
	"strings"
	"time"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

type Repository struct {
	client *ch.Client
}

func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

func ttDateFilter(field string, params *ch.QueryParams) string {
	return fmt.Sprintf(
		"toDate(%s, '%s') BETWEEN '%s' AND '%s'",
		field,
		params.Timezone,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)
}

func ttFormatIDList(ids []string) string {
	quoted := make([]string, 0, len(ids))
	for _, id := range ids {
		quoted = append(quoted, "'"+strings.ReplaceAll(id, "'", "\\'")+"'")
	}
	return strings.Join(quoted, ",")
}

func (r *Repository) GetSummary(ctx context.Context, params *ch.QueryParams) (*SummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	idList := ttFormatIDList(params.AccountIDs)
	postsFilter := ttDateFilter("created_at", params)
	insightsFilter := ttDateFilter("inserted_at", params)

	query := fmt.Sprintf(`
SELECT
    platform_id.tiktok_id AS tiktok_id,
    if(posts_summary.display_name = '', '', posts_summary.display_name) AS page_name,
    '' AS logo,
    toInt64(coalesce(posts_summary.total_likes, 0)) AS total_likes,
    toInt64(coalesce(posts_summary.total_comments, 0)) AS total_comments,
    toInt64(coalesce(posts_summary.total_shares, 0)) AS total_shares,
    toInt64(coalesce(posts_summary.total_engagements, 0)) AS total_engagements,
    toInt64(coalesce(posts_summary.total_posts, 0)) AS total_posts,
    toInt64(coalesce(insights_summary.total_follower_count, 0)) AS total_follower_count,
    toInt64(coalesce(insights_summary.total_following_count, 0)) AS total_following_count,
    toInt64(coalesce(posts_summary.total_video_views, 0)) AS total_video_views
FROM (
    SELECT arrayJoin([%s]) AS tiktok_id
) AS platform_id
LEFT JOIN (
    SELECT
        tiktok_id,
        display_name,
        total_likes,
        total_comments,
        total_shares,
        total_engagements,
        total_video_views,
        total_posts
    FROM (
        SELECT
            tiktok_id,
            any(display_name) AS display_name,
            toInt64(sum(like_count)) AS total_likes,
            toInt64(sum(comments_count)) AS total_comments,
            toInt64(sum(share_count)) AS total_shares,
            toInt64(sum(engagement_count)) AS total_engagements,
            toInt64(sum(view_count)) AS total_video_views,
            toInt64(count()) AS total_posts
        FROM (
            SELECT
                tiktok_id,
                post_id,
                max(like_count) AS like_count,
                max(comments_count) AS comments_count,
                max(share_count) AS share_count,
                max(engagement_count) AS engagement_count,
                max(view_count) AS view_count,
                any(display_name) AS display_name
            FROM tiktok_posts FINAL
            WHERE tiktok_id IN %s AND %s
            GROUP BY tiktok_id, post_id
        )
        GROUP BY tiktok_id
    )
) AS posts_summary ON platform_id.tiktok_id = posts_summary.tiktok_id
LEFT JOIN (
    SELECT
        tiktok_id,
        toInt64(max(total_follower_count)) AS total_follower_count,
        toInt64(max(total_following_count)) AS total_following_count
    FROM tiktok_insights FINAL
    WHERE tiktok_id IN %s AND %s
    GROUP BY tiktok_id
) AS insights_summary ON platform_id.tiktok_id = insights_summary.tiktok_id
`, idList, ids, postsFilter, ids, insightsFilter)

	var result SummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TiktokID,
		&result.PageName,
		&result.Logo,
		&result.TotalLikes,
		&result.TotalComments,
		&result.TotalShares,
		&result.TotalEngagements,
		&result.TotalPosts,
		&result.TotalFollowerCount,
		&result.TotalFollowingCount,
		&result.TotalVideoViews,
	)
	if err != nil {
		return nil, fmt.Errorf("GetSummary: %w", err)
	}

	return &result, nil
}

func (r *Repository) GetFollowersAndViews(ctx context.Context, params *ch.QueryParams, dynamic bool) (*FollowersViewsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ttDateFilter("inserted_at", params)

	periodExpr := "toDateTime(inserted_at)"
	fillFrom := fmt.Sprintf("toStartOfDay(toDate('%s'))", params.DateFrom.Format("2006-01-02"))
	fillTo := fmt.Sprintf("toStartOfDay(toDate('%s'))", params.DateTo.Add(24*time.Hour).Format("2006-01-02"))
	fillStep := "INTERVAL 1 DAY"
	if dynamic && params.DayCount > 180 {
		periodExpr = "toStartOfMonth(toDate(inserted_at))"
		fillFrom = fmt.Sprintf("toStartOfMonth(toDate('%s'))", params.DateFrom.Format("2006-01-02"))
		fillTo = fmt.Sprintf("toStartOfMonth(toDate('%s'))", params.DateTo.Add(24*time.Hour).Format("2006-01-02"))
		fillStep = "INTERVAL 1 MONTH"
	}

	query := fmt.Sprintf(`
SELECT
    platform_id,
    any(display_name) AS display_name,
    '' AS logo,
    groupArray(follower_count) AS followers_count,
    groupArray(views_per_day) AS views_per_day,
    groupArray(follower_count_diff) AS followers_count_diff,
    groupArray(views_per_day_diff) AS views_per_day_diff,
    groupArray(bucket_day) AS day_bucket
FROM (
    SELECT
        platform_id,
        display_name,
        follower_count,
        greatest(
            follower_count - lagInFrame(follower_count, 1, 0)
                OVER (PARTITION BY platform_id ORDER BY bucket_time ASC),
            0
        ) AS follower_count_diff,
        views_per_day,
        greatest(
            views_per_day - lagInFrame(views_per_day, 1, 0)
                OVER (PARTITION BY platform_id ORDER BY bucket_time ASC),
            0
        ) AS views_per_day_diff,
        formatDateTime(toDateTime(bucket_time), '%%Y-%%m-%%d %%H:%%i:%%S') AS bucket_day,
        bucket_time
    FROM (
        SELECT
            max(tiktok_id) AS platform_id,
            max(display_name) AS display_name,
            toInt64(max(total_follower_count)) AS follower_count,
            toInt64(max(total_video_views)) AS views_per_day,
            %s AS bucket_time
        FROM tiktok_insights FINAL
        WHERE tiktok_id IN %s AND %s
        GROUP BY bucket_time
        ORDER BY bucket_time
        WITH FILL
            FROM %s
            TO %s
            STEP %s
            INTERPOLATE (
                platform_id AS '%s',
                display_name AS '',
                follower_count AS -1,
                views_per_day AS -1
            )
    )
)
GROUP BY platform_id
`, periodExpr, ids, dateFilter, fillFrom, fillTo, fillStep, strings.ReplaceAll(params.AccountIDs[0], "'", "\\'"))

	var result FollowersViewsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PlatformID,
		&result.DisplayName,
		&result.Logo,
		&result.FollowersCount,
		&result.ViewsPerDay,
		&result.FollowersCountDiff,
		&result.ViewsPerDayDiff,
		&result.DayBucket,
	)
	if err != nil {
		return nil, fmt.Errorf("GetFollowersAndViews: %w", err)
	}

	return &result, nil
}

func (r *Repository) GetPostsAndEngagements(ctx context.Context, params *ch.QueryParams) (*PostsEngagementResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ttDateFilter("created_at", params)
	fillFrom := params.DateFrom.Format("2006-01-02")
	fillTo := params.DateTo.Add(24 * time.Hour).Format("2006-01-02")

	query := fmt.Sprintf(`
SELECT
    tiktok_id,
    any(display_name) AS page_name,
    '' AS logo,
    groupArray(formatDateTime(posting_day, '%%Y-%%m-%%d')) AS days_bucket,
    groupArray(sum_view_count) AS sum_view_count,
    groupArray(sum_like_count) AS sum_like_count,
    groupArray(sum_comments_count) AS sum_comments_count,
    groupArray(sum_share_count) AS sum_share_count,
    groupArray(sum_engagement_count) AS sum_engagement_count,
    groupArray(avg_engagement_rate) AS avg_engagement_rate,
    groupArray(post_count) AS post_count
FROM (
    SELECT
        tiktok_id,
        any(display_name) AS display_name,
        posting_day,
        toInt64(sum(view_count)) AS sum_view_count,
        toInt64(sum(like_count)) AS sum_like_count,
        toInt64(sum(comments_count)) AS sum_comments_count,
        toInt64(sum(share_count)) AS sum_share_count,
        toInt64(sum(engagement_count)) AS sum_engagement_count,
        toInt64(round(avg(engagement_rate), 2)) AS avg_engagement_rate,
        toInt64(count()) AS post_count
    FROM (
        SELECT
            tiktok_id,
            post_id,
            max(view_count) AS view_count,
            max(like_count) AS like_count,
            max(comments_count) AS comments_count,
            max(share_count) AS share_count,
            max(engagement_count) AS engagement_count,
            max(engagement_rate) AS engagement_rate,
            toDate(max(created_at)) AS posting_day,
            argMax(display_name, inserted_at) AS display_name
        FROM tiktok_posts FINAL
        WHERE tiktok_id IN %s AND %s
        GROUP BY tiktok_id, post_id
    )
    GROUP BY tiktok_id, posting_day
    ORDER BY posting_day ASC
    WITH FILL FROM toDate('%s') TO toDate('%s') STEP 1
    INTERPOLATE (tiktok_id AS '%s')
)
GROUP BY tiktok_id
`, ids, dateFilter, fillFrom, fillTo, strings.ReplaceAll(params.AccountIDs[0], "'", "\\'"))

	var result PostsEngagementResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TiktokID,
		&result.PageName,
		&result.Logo,
		&result.DaysBucket,
		&result.SumViewCount,
		&result.SumLikeCount,
		&result.SumCommentsCount,
		&result.SumShareCount,
		&result.SumEngagementCount,
		&result.AvgEngagementRate,
		&result.PostCount,
	)
	if err != nil {
		return nil, fmt.Errorf("GetPostsAndEngagements: %w", err)
	}

	return &result, nil
}

func (r *Repository) GetDailyEngagementsData(ctx context.Context, params *ch.QueryParams, dynamic bool) (*DailyEngagementResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ttDateFilter("created_at", params)

	periodExpr := "toDate(post_created_at)"
	fillFrom := fmt.Sprintf("toDate('%s')", params.DateFrom.Format("2006-01-02"))
	fillTo := fmt.Sprintf("toDate('%s')", params.DateTo.Add(24*time.Hour).Format("2006-01-02"))
	fillStep := "INTERVAL 1 DAY"
	if dynamic && params.DayCount > 180 {
		periodExpr = "toStartOfMonth(toDate(post_created_at))"
		fillFrom = fmt.Sprintf("toStartOfMonth(toDate('%s'))", params.DateFrom.Format("2006-01-02"))
		fillTo = fmt.Sprintf("toStartOfMonth(toDate('%s'))", params.DateTo.Add(24*time.Hour).Format("2006-01-02"))
		fillStep = "INTERVAL 1 MONTH"
	}

	query := fmt.Sprintf(`
SELECT
    platform_id AS tiktok_id,
    '' AS page_name,
    '' AS logo,
    groupArray(total_video_likes) AS total_video_likes,
    groupArray(total_video_comments) AS total_video_comments,
    groupArray(total_video_shares) AS total_video_shares,
    groupArray(daily_video_likes) AS daily_video_likes,
    groupArray(daily_video_comments) AS daily_video_comments,
    groupArray(daily_video_shares) AS daily_video_shares,
    groupArray(total_engagement) AS total_engagement,
    groupArray(daily_engagement) AS daily_engagement,
    groupArray(metric_day) AS days_bucket
FROM (
    SELECT
        platform_id,
        sum(daily_video_likes) OVER w AS total_video_likes,
        sum(daily_video_comments) OVER w AS total_video_comments,
        sum(daily_video_shares) OVER w AS total_video_shares,
        toInt64(greatest(daily_video_likes, 0)) AS daily_video_likes,
        toInt64(greatest(daily_video_comments, 0)) AS daily_video_comments,
        toInt64(greatest(daily_video_shares, 0)) AS daily_video_shares,
        (sum(daily_video_likes) OVER w + sum(daily_video_comments) OVER w + sum(daily_video_shares) OVER w) AS total_engagement,
        toInt64(greatest(daily_video_likes, 0) + greatest(daily_video_comments, 0) + greatest(daily_video_shares, 0)) AS daily_engagement,
        formatDateTime(toDateTime(metric_day), '%%Y-%%m-%%d %%H:%%i:%%S') AS metric_day
    FROM (
        SELECT
            toInt64(ifNull(daily_video_likes, 0)) AS daily_video_likes,
            toInt64(ifNull(daily_video_comments, 0)) AS daily_video_comments,
            toInt64(ifNull(daily_video_shares, 0)) AS daily_video_shares,
            platform_id,
            metric_day
        FROM (
            SELECT
                toInt64(sum(like_count)) AS daily_video_likes,
                toInt64(sum(comments_count)) AS daily_video_comments,
                toInt64(sum(share_count)) AS daily_video_shares,
                max(tiktok_id) AS platform_id,
                %s AS metric_day
            FROM (
                SELECT
                    tiktok_id,
                    post_id,
                    max(like_count) AS like_count,
                    max(comments_count) AS comments_count,
                    max(share_count) AS share_count,
                    max(created_at) AS post_created_at
                FROM tiktok_posts FINAL
                WHERE tiktok_id IN %s AND %s
                GROUP BY tiktok_id, post_id
            )
            GROUP BY metric_day
            ORDER BY metric_day ASC
            WITH FILL FROM %s TO %s STEP %s
            INTERPOLATE (platform_id AS '%s')
        )
    )
    WINDOW w AS (PARTITION BY platform_id ORDER BY metric_day ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW)
)
GROUP BY platform_id
`, periodExpr, ids, dateFilter, fillFrom, fillTo, fillStep, strings.ReplaceAll(params.AccountIDs[0], "'", "\\'"))

	var result DailyEngagementResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TiktokID,
		&result.PageName,
		&result.Logo,
		&result.TotalVideoLikes,
		&result.TotalVideoComments,
		&result.TotalVideoShares,
		&result.DailyVideoLikes,
		&result.DailyVideoComments,
		&result.DailyVideoShares,
		&result.TotalEngagement,
		&result.DailyEngagement,
		&result.DaysBucket,
	)
	if err != nil {
		return nil, fmt.Errorf("GetDailyEngagementsData: %w", err)
	}

	return &result, nil
}

func (r *Repository) basePostsQuery(params *ch.QueryParams) string {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ttDateFilter("created_at", params)
	return fmt.Sprintf(`
FROM tiktok_posts FINAL
WHERE tiktok_id IN %s AND %s
GROUP BY tiktok_id, post_id
`, ids, dateFilter)
}

func (r *Repository) GetTopAndLeastPerformingPosts(ctx context.Context, params *ch.QueryParams) (top []PostRow, least []PostRow, err error) {
	base := r.basePostsQuery(params)
	query := fmt.Sprintf(`
WITH latest_insights AS (
    SELECT
        tiktok_id,
        toInt64(argMax(total_follower_count, inserted_at)) AS total_follower_count
    FROM tiktok_insights FINAL
    WHERE tiktok_id IN %s
    GROUP BY tiktok_id
)
SELECT
    category,
    p.tiktok_id,
    p.page_name,
    '' AS logo,
    p.profile_link,
    p.post_id,
    p.cover_image_url,
    p.share_url,
    p.post_description,
    p.hashtags,
    p.duration,
    p.height,
    p.width,
    p.title,
    p.embed_html,
    p.embed_link,
    p.likes_count,
    p.comments_count,
    p.shares_count,
    p.views_count,
    p.engagements_count,
    p.total_engagement,
    p.engagement_rate,
    p.inserted_at_value,
    p.created_time_value,
    coalesce(i.total_follower_count, 0) AS total_follower_count
FROM (
    SELECT
        'top_posts' AS category,
        tiktok_id,
        argMax(display_name, inserted_at) AS page_name,
        argMax(profile_link, inserted_at) AS profile_link,
        post_id,
        argMax(cover_image_url, inserted_at) AS cover_image_url,
        argMax(share_url, inserted_at) AS share_url,
        argMax(post_description, inserted_at) AS post_description,
        argMax(hashtags, inserted_at) AS hashtags,
        toInt64(argMax(duration, inserted_at)) AS duration,
        toInt64(argMax(height, inserted_at)) AS height,
        toInt64(argMax(width, inserted_at)) AS width,
        argMax(title, inserted_at) AS title,
        argMax(embed_html, inserted_at) AS embed_html,
        argMax(embed_link, inserted_at) AS embed_link,
        toInt64(max(like_count)) AS likes_count,
        toInt64(max(comments_count)) AS comments_count,
        toInt64(max(share_count)) AS shares_count,
        toInt64(max(view_count)) AS views_count,
        toInt64(max(engagement_count)) AS engagements_count,
        toInt64(max(engagement_count)) AS total_engagement,
        round(max(engagement_rate), 2) AS engagement_rate,
        formatDateTime(max(inserted_at), '%%Y-%%m-%%d %%H:%%i:%%S') AS inserted_at_value,
        formatDateTime(max(created_at), '%%Y-%%m-%%d %%H:%%i:%%S') AS created_time_value
    %s
    ORDER BY engagements_count DESC
    LIMIT 5
    UNION ALL
    SELECT
        'least_posts' AS category,
        tiktok_id,
        argMax(display_name, inserted_at) AS page_name,
        argMax(profile_link, inserted_at) AS profile_link,
        post_id,
        argMax(cover_image_url, inserted_at) AS cover_image_url,
        argMax(share_url, inserted_at) AS share_url,
        argMax(post_description, inserted_at) AS post_description,
        argMax(hashtags, inserted_at) AS hashtags,
        toInt64(argMax(duration, inserted_at)) AS duration,
        toInt64(argMax(height, inserted_at)) AS height,
        toInt64(argMax(width, inserted_at)) AS width,
        argMax(title, inserted_at) AS title,
        argMax(embed_html, inserted_at) AS embed_html,
        argMax(embed_link, inserted_at) AS embed_link,
        toInt64(max(like_count)) AS likes_count,
        toInt64(max(comments_count)) AS comments_count,
        toInt64(max(share_count)) AS shares_count,
        toInt64(max(view_count)) AS views_count,
        toInt64(max(engagement_count)) AS engagements_count,
        toInt64(max(engagement_count)) AS total_engagement,
        round(max(engagement_rate), 2) AS engagement_rate,
        formatDateTime(max(inserted_at), '%%Y-%%m-%%d %%H:%%i:%%S') AS inserted_at_value,
        formatDateTime(max(created_at), '%%Y-%%m-%%d %%H:%%i:%%S') AS created_time_value
    %s
    ORDER BY engagements_count ASC
    LIMIT 5
) p
LEFT JOIN latest_insights i ON i.tiktok_id = p.tiktok_id
`, ch.FormatAccountIDs(params.AccountIDs), base, base)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, nil, fmt.Errorf("GetTopAndLeastPerformingPosts query: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var row PostRow
		if err := rows.Scan(
			&row.Category,
			&row.TiktokID,
			&row.PageName,
			&row.Logo,
			&row.ProfileLink,
			&row.PostID,
			&row.CoverImageURL,
			&row.ShareURL,
			&row.PostDescription,
			&row.Hashtags,
			&row.Duration,
			&row.Height,
			&row.Width,
			&row.Title,
			&row.EmbedHTML,
			&row.EmbedLink,
			&row.LikesCount,
			&row.CommentsCount,
			&row.SharesCount,
			&row.ViewsCount,
			&row.EngagementsCount,
			&row.TotalEngagement,
			&row.EngagementRate,
			&row.InsertedAt,
			&row.CreatedTime,
			&row.TotalFollowerCount,
		); err != nil {
			return nil, nil, fmt.Errorf("GetTopAndLeastPerformingPosts scan: %w", err)
		}
		if row.Category == "top_posts" {
			top = append(top, row)
		} else {
			least = append(least, row)
		}
	}
	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("GetTopAndLeastPerformingPosts rows: %w", err)
	}
	return top, least, nil
}

func sanitizeSortField(sort string) string {
	allowed := map[string]bool{
		"total_engagement":  true,
		"engagements_count": true,
		"likes_count":       true,
		"comments_count":    true,
		"shares_count":      true,
		"views_count":       true,
		"created_time":      true,
	}
	if allowed[sort] {
		return sort
	}
	return "total_engagement"
}

func (r *Repository) GetPostsData(ctx context.Context, params *ch.QueryParams, sortOrder string, limit, offset int) ([]PostRow, error) {
	sortField := sanitizeSortField(strings.TrimSpace(sortOrder))
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ttDateFilter("created_at", params)

	query := fmt.Sprintf(`
WITH post_aggs AS (
    SELECT
        tiktok_id,
        post_id,
        argMax(display_name, inserted_at)    AS page_name,
        argMax(profile_link, inserted_at)    AS profile_link,
        argMax(cover_image_url, inserted_at) AS cover_image_url,
        argMax(share_url, inserted_at)       AS share_url,
        argMax(post_description, inserted_at) AS post_description,
        argMax(hashtags, inserted_at)        AS hashtags,
        toInt64(argMax(duration, inserted_at)) AS duration,
        toInt64(argMax(height, inserted_at)) AS height,
        toInt64(argMax(width, inserted_at))  AS width,
        argMax(title, inserted_at)           AS title,
        argMax(embed_html, inserted_at)      AS embed_html,
        argMax(embed_link, inserted_at)      AS embed_link,
        toInt64(max(like_count))             AS likes_count,
        toInt64(max(comments_count))         AS comments_count,
        toInt64(max(share_count))            AS shares_count,
        toInt64(max(view_count))             AS views_count,
        toInt64(max(engagement_count))       AS engagements_count,
        toInt64(max(engagement_count))       AS total_engagement,
        round(max(engagement_rate), 2)       AS engagement_rate,
        formatDateTime(max(inserted_at), '%%Y-%%m-%%d %%H:%%i:%%S') AS inserted_at_value,
        formatDateTime(max(created_at),  '%%Y-%%m-%%d %%H:%%i:%%S') AS created_time_value
    FROM tiktok_posts FINAL
    WHERE tiktok_id IN %s AND %s
    GROUP BY tiktok_id, post_id
),
total_counts AS (
    SELECT tiktok_id, toInt64(count()) AS total
    FROM post_aggs
    GROUP BY tiktok_id
),
latest_insights AS (
    SELECT
        tiktok_id,
        toInt64(argMax(total_follower_count, inserted_at)) AS total_follower_count
    FROM tiktok_insights FINAL
    WHERE tiktok_id IN %s
    GROUP BY tiktok_id
)
SELECT
    p.tiktok_id,
    p.page_name,
    '' AS logo,
    p.profile_link,
    p.post_id,
    p.cover_image_url,
    p.share_url,
    p.post_description,
    p.hashtags,
    p.duration,
    p.height,
    p.width,
    p.title,
    p.embed_html,
    p.embed_link,
    p.likes_count,
    p.comments_count,
    p.shares_count,
    p.views_count,
    p.engagements_count,
    p.total_engagement,
    p.engagement_rate,
    p.inserted_at_value,
    p.created_time_value,
    coalesce(i.total_follower_count, 0) AS total_follower_count,
    coalesce(c.total, 0)                AS total
FROM (
    SELECT * FROM post_aggs
    ORDER BY %s DESC
    LIMIT %d OFFSET %d
) p
LEFT JOIN total_counts c    ON c.tiktok_id = p.tiktok_id
LEFT JOIN latest_insights i ON i.tiktok_id = p.tiktok_id
`, ids, dateFilter, ids, sortField, limit, offset)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetPostsData query: %w", err)
	}
	defer rows.Close()

	result := make([]PostRow, 0)
	for rows.Next() {
		var row PostRow
		if err := rows.Scan(
			&row.TiktokID,
			&row.PageName,
			&row.Logo,
			&row.ProfileLink,
			&row.PostID,
			&row.CoverImageURL,
			&row.ShareURL,
			&row.PostDescription,
			&row.Hashtags,
			&row.Duration,
			&row.Height,
			&row.Width,
			&row.Title,
			&row.EmbedHTML,
			&row.EmbedLink,
			&row.LikesCount,
			&row.CommentsCount,
			&row.SharesCount,
			&row.ViewsCount,
			&row.EngagementsCount,
			&row.TotalEngagement,
			&row.EngagementRate,
			&row.InsertedAt,
			&row.CreatedTime,
			&row.TotalFollowerCount,
			&row.Total,
		); err != nil {
			return nil, fmt.Errorf("GetPostsData scan: %w", err)
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetPostsData rows: %w", err)
	}

	return result, nil
}
