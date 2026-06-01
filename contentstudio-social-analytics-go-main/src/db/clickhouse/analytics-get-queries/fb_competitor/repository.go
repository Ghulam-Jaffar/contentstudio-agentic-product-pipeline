// Package fb_competitor provides ClickHouse queries for Facebook competitor analytics.
// Migrated from PHP FacebookCompetitorBuilder (contentstudio-backend).
// Tables: facebook_competitor_posts, facebook_competitor_insights, facebook_competitor_media_assets.
package fb_competitor

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// Repository provides Facebook competitor analytics queries against ClickHouse.
type Repository struct {
	client *ch.Client
}

// NewRepository creates a new Facebook competitor analytics repository.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// accountFilter formats page IDs for WHERE clauses.
func accountFilter(field string, ids []string) string {
	return fmt.Sprintf("%s IN %s", field, ch.FormatAccountIDs(ids))
}

// accountValues formats page IDs as VALUES for subquery: (('id1'),('id2'))
func accountValues(ids []string) string {
	if len(ids) == 0 {
		return "((''))"
	}
	parts := make([]string, len(ids))
	for i, id := range ids {
		parts[i] = "('" + strings.ReplaceAll(id, "'", "\\'") + "')"
	}
	return "(" + strings.Join(parts, ",") + ")"
}

// dateFilter returns a BETWEEN clause for posts (created_at) or insights (inserted_at) in UTC.
func dateFilter(startDate, endDate string, insightsFlag bool) string {
	field := "created_at"
	if insightsFlag {
		field = "inserted_at"
	}
	return fmt.Sprintf("%s BETWEEN toDateTime('%s',0) AND toDateTime('%s',0)", field, startDate, endDate)
}

// dateFilterField returns a BETWEEN clause for an arbitrary date column.
func dateFilterField(dateColumn, startDate, endDate string) string {
	return fmt.Sprintf("%s BETWEEN toDateTime('%s',0) AND toDateTime('%s',0)", dateColumn, startDate, endDate)
}

// constantConditions builds a multiIf expression to map IDs to account metadata.
func constantConditions(accounts map[string]AccountInfo, field, idField string) string {
	q := "multiIf("
	for id, info := range accounts {
		val := strings.ReplaceAll(info.fieldValue(field), "'", "\\'")
		q += fmt.Sprintf("%s='%s','%s',", idField, id, val)
	}
	q += "'Random')"
	return q
}

// dayMapping generates a multiIf mapping day-of-week numbers to day names.
func dayMapping(field string) string {
	days := []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"}
	m := "multiIf("
	for i, d := range days {
		m += fmt.Sprintf("%s = %d, '%s',\n", field, i+1, d)
	}
	m += "NULL)"
	return m
}

// hourMapping generates a multiIf mapping hour numbers to formatted time strings.
func hourMapping(field string) string {
	hours := []string{
		"12:00 AM", "01:00 AM", "02:00 AM", "03:00 AM", "04:00 AM", "05:00 AM",
		"06:00 AM", "07:00 AM", "08:00 AM", "09:00 AM", "10:00 AM", "11:00 AM",
		"12:00 PM", "01:00 PM", "02:00 PM", "03:00 PM", "04:00 PM", "05:00 PM",
		"06:00 PM", "07:00 PM", "08:00 PM", "09:00 PM", "10:00 PM", "11:00 PM",
	}
	m := "multiIf("
	for i, h := range hours {
		m += fmt.Sprintf("%s = %d, '%s',\n", field, i, h)
	}
	m += "'Random')"
	return m
}

// AccountInfo holds metadata fetched from the competitor report for building query constants.
type AccountInfo struct {
	Image string
	Name  string
	State string
	Slug  string
}

func (a AccountInfo) fieldValue(field string) string {
	switch field {
	case "image":
		return a.Image
	case "name":
		return a.Name
	case "state":
		return a.State
	case "slug":
		return a.Slug
	}
	return ""
}

// CompetitorQueryParams holds Facebook competitor query parameters.
type CompetitorQueryParams struct {
	PageIDs   []string
	Accounts  map[string]AccountInfo
	StartDate string // UTC formatted: "2025-01-01 00:00:00"
	EndDate   string // UTC formatted: "2025-01-31 23:59:59"
	DaysDiff  int
}

// PrevPeriod returns a new CompetitorQueryParams shifted back by DaysDiff days.
func (p *CompetitorQueryParams) PrevPeriod() *CompetitorQueryParams {
	startDate := mustParseShift(p.StartDate, -p.DaysDiff)
	endDate := mustParseShift(p.EndDate, -p.DaysDiff)
	return &CompetitorQueryParams{
		PageIDs:   p.PageIDs,
		Accounts:  p.Accounts,
		StartDate: startDate,
		EndDate:   endDate,
		DaysDiff:  p.DaysDiff,
	}
}

func mustParseShift(dateStr string, days int) string {
	layouts := []string{"2006-01-02 15:04:05", "2006-01-02"}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, dateStr); err == nil {
			return t.AddDate(0, 0, days).Format(layout)
		}
	}
	return dateStr
}

// GetDataTableMetrics returns per-competitor table metrics for the date range.
func (r *Repository) GetDataTableMetrics(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	accountIds := accountValues(params.PageIDs)
	pageFilter := accountFilter("facebook_id", params.PageIDs)
	insightsPageFilter := accountFilter("page_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)
	insightsDateFilt := dateFilter(params.StartDate, params.EndDate, true)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    %s as image,
    %s as name,
    %s as state,
    * FROM (
    SELECT
        pages_ids.facebook_id as facebook_id,
        week_metrics.averageEngagement as averageEngagement,
        week_metrics.averagePostsPerWeek as averagePostsPerWeek,
        week_metrics.engagementRate as engagementRate,
        days_metrics.dayOfWeek as dayOfWeek,
        days_metrics.hourOfDay as hourOfDay,
        days_metrics.averagePostsPerDay as averagePostsPerDay,
        days_metrics.averagePostsPerDayEngagement as averagePostsPerDayEngagement
    FROM (
        SELECT CAST(c1 AS String) as facebook_id FROM VALUES %s
    ) as pages_ids
    LEFT JOIN (
        SELECT
            facebook_id,
            round(avg(total_engagement), 2) as averageEngagement,
            if(dateDiff('week', toDate('%s'), toDate('%s'))!=0,
                round(sum(posts_in_a_week)/dateDiff('week', toDate('%s'), toDate('%s')),2), 0) as averagePostsPerWeek,
            round((sum(total_engagement)/sum(posts_in_a_week)) / max(followers_count) * 100, 2) as engagementRate
        FROM (
            SELECT
                CAST(facebook_id AS String) AS facebook_id,
                sum(post_engagement) as total_engagement,
                toStartOfWeek(created_at) as week,
                count() as posts_in_a_week,
                followers_count
            FROM facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            GROUP BY facebook_id, week, followers_count
            ORDER BY week ASC
            WITH FILL FROM toStartOfWeek(toDate('%s')) TO toStartOfWeek(toDate('%s')) STEP INTERVAL 1 WEEK
        )
        GROUP BY facebook_id
    ) as week_metrics ON CAST(week_metrics.facebook_id AS String) = CAST(pages_ids.facebook_id AS String)
    LEFT JOIN (
        SELECT
            facebook_id,
            max(page_name) as slug,
            argMax(%s, total_posts) as dayOfWeek,
            argMax(%s, total_posts) as hourOfDay,
            if(dateDiff('day', toDate('%s'), toDate('%s'))!=0,
                round(sum(total_posts)/dateDiff('day', toDate('%s'), toDate('%s')),2), 0) as averagePostsPerDay,
            avg(total_engagement) as averagePostsPerDayEngagement,
            argMax(hour_of_day, total_posts) as maximumPostsHour,
            argMax(hour_of_day, total_engagement) as mostEngagementHour
        FROM (
            SELECT
                facebook_id,
                max(page_name) as page_name,
                sum(post_engagement) as total_engagement,
                toDayOfWeek(created_at) as day_of_week,
                toHour(created_at) as hour_of_day,
                uniq(post_id) as total_posts,
                toDate(created_at) as date_c
            FROM facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            GROUP BY facebook_id, day_of_week, hour_of_day, date_c
            ORDER BY date_c ASC WITH FILL FROM (toDate('%s')) TO (toDate('%s')) STEP 1
        )
        GROUP BY facebook_id
    ) as days_metrics ON CAST(days_metrics.facebook_id AS String) = CAST(pages_ids.facebook_id AS String)
) as page_metrics
LEFT JOIN (
    SELECT
        max(followers_count) as followersCount,
        max(total_fan_count) as fanCount,
        page_id
    FROM facebook_competitor_insights
    WHERE %s AND %s
    GROUP BY page_id
) as page_metadata ON page_metrics.facebook_id = page_metadata.page_id
ORDER BY %s DESC`,
		pageFilter, dateFilt,
		constantConditions(params.Accounts, "image", "facebook_id"),
		constantConditions(params.Accounts, "name", "facebook_id"),
		constantConditions(params.Accounts, "state", "facebook_id"),
		accountIds,
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		dayMapping("day_of_week"),
		hourMapping("hour_of_day"),
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		insightsPageFilter, insightsDateFilt,
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}

// GetPostingActivityGraphByTypes returns aggregated posting activity per media type.
func (r *Repository) GetPostingActivityGraphByTypes(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	pageFilter := accountFilter("facebook_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    media_type as mediaType,
    round(sum(avg_engagement),2) as avgTotalEngagements,
    sum(page_post_count) as totalPosts,
    toInt32(sum(total_engagement)) as total_engagement,
    round(sum(er),2) as avgEngagementRate,
    if(dateDiff('week', toDate('%s'), toDate('%s')) = 0, 0,
        round(totalPosts/dateDiff('week', toDate('%s'), toDate('%s')),2)) as postsPerWeek,
    if(dateDiff('day', toDate('%s'), toDate('%s')) = 0, 0,
        round(totalPosts/dateDiff('day', toDate('%s'), toDate('%s')),2)) as postsPerDay,
    if(dateDiff('hour', toDate('%s'), toDate('%s')) = 0, 0,
        round(totalPosts/dateDiff('hour', toDate('%s'), toDate('%s')),2)) as postsPerHour,
    dateDiff('week', toDate('%s'), toDate('%s')) as weekCount,
    dateDiff('day', toDate('%s'), toDate('%s')) as dayCount,
    dateDiff('hour', toDate('%s'), toDate('%s')) as hourCount
FROM (
    SELECT
        facebook_id, page_name, media_type,
        sum(count) as page_post_count,
        groupArray(total_post_engagement),
        if(page_post_count <= 0, 0, round(sum(total_post_engagement)/page_post_count,2)) as avg_engagement,
        if(page_post_count <= 0,0,round(((sum(er)/page_post_count)/max(followers_count))*100,2)) as er,
        sum(total_post_engagement) as total_engagement
    FROM (
        SELECT
            facebook_id, page_name, media_type,
            count() as count,
            sum(post_engagement) as total_post_engagement,
            total_post_engagement as er,
            argMax(followers_count,created_at) as followers_count,
            toStartOfWeek(created_at) as week
        FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY facebook_id, page_name, media_type, week
        ORDER BY week ASC
        WITH FILL FROM toStartOfWeek(toDate('%s')) TO toStartOfWeek(toDate('%s')) STEP INTERVAL 1 WEEK
    )
    WHERE media_type!=''
    GROUP BY facebook_id, page_name, media_type
)
GROUP BY media_type
ORDER BY %s DESC`,
		pageFilter, dateFilt,
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}

// GetPostingActivityBySpecificType returns per-competitor metrics for a specific media type.
func (r *Repository) GetPostingActivityBySpecificType(ctx context.Context, params *CompetitorQueryParams, mediaType, sortOrder string) ([]map[string]interface{}, error) {
	accountIds := accountValues(params.PageIDs)
	pageFilter := accountFilter("facebook_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    pages_ids.facebook_id as facebook_id,
    %s as name,
    %s as image,
    %s as state,
    %s as slug,
    page_insights.media_type as mediaType,
    page_insights.avgTotalEngagements as avgTotalEngagements,
    page_insights.total_engagement as totalEngagement,
    page_insights.avgCountByWeek as avgCountByWeek,
    page_insights.avgCountByDay as avgCountByDay,
    page_insights.avgCountByHour as avgCountByHour,
    page_insights.totalPosts as totalPosts,
    page_insights.weekCount as weekCount,
    page_insights.dayCount as dayCount,
    page_insights.hourCount as hourCount,
    page_insights.avgEngagementRate as avgEngagementRate,
    page_constants.followersCount as followersCount
FROM (
    SELECT c1 as facebook_id FROM VALUES %s
) as pages_ids
LEFT JOIN (
    SELECT
        media_type, facebook_id,
        sum(total_engagement) as total_engagement,
        if(totalPosts <= 0, 0, round(total_engagement/totalPosts, 2)) as avgTotalEngagements,
        if(dateDiff('week', toDate('%s'), toDate('%s')) = 0, 0,
            round(sum(count)/dateDiff('week', toDate('%s'), toDate('%s')),2)) as avgCountByWeek,
        if(dateDiff('day', toDate('%s'), toDate('%s')) = 0, 0,
            round(sum(count)/dateDiff('day', toDate('%s'), toDate('%s')),2)) as avgCountByDay,
        if(dateDiff('hour', toDate('%s'), toDate('%s')) = 0, 0,
            round(sum(count)/dateDiff('hour', toDate('%s'), toDate('%s')),2)) as avgCountByHour,
        sum(count) as totalPosts,
        dateDiff('week', toDate('%s'), toDate('%s')) as weekCount,
        dateDiff('day', toDate('%s'), toDate('%s')) as dayCount,
        dateDiff('hour', toDate('%s'), toDate('%s')) as hourCount,
        if(totalPosts <= 0, 0, round(((total_engagement/totalPosts/max(followers_count))*100), 2)) as avgEngagementRate
    FROM (
        SELECT
            facebook_id, media_type,
            count() as count,
            sum(post_engagement) as total_engagement,
            argMax(followers_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY facebook_id, media_type, week
        ORDER BY week ASC
        WITH FILL FROM toStartOfWeek(toDate('%s')) TO toStartOfWeek(toDate('%s')) STEP INTERVAL 1 WEEK
    ) as page_metrics
    WHERE media_type = '%s'
    GROUP BY facebook_id, media_type
) as page_insights USING facebook_id
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        argMax(followers_count, inserted_at) as followersCount,
        page_id
    FROM facebook_competitor_insights
    GROUP BY page_id
) as page_constants ON page_constants.page_id = pages_ids.facebook_id
ORDER BY %s DESC`,
		pageFilter, dateFilt,
		constantConditions(params.Accounts, "name", "pages_ids.facebook_id"),
		constantConditions(params.Accounts, "image", "pages_ids.facebook_id"),
		constantConditions(params.Accounts, "state", "pages_ids.facebook_id"),
		constantConditions(params.Accounts, "slug", "pages_ids.facebook_id"),
		accountIds,
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		sanitizeString(mediaType),
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}

// GetTopAndLeastPerformingPosts returns top 5 and least 5 posts per competitor with media assets.
func (r *Repository) GetTopAndLeastPerformingPosts(ctx context.Context, params *CompetitorQueryParams) ([]map[string]interface{}, error) {
	pageFilter := accountFilter("facebook_id", params.PageIDs)
	insightsPageFilter := accountFilter("page_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
),
top_posts_ids AS (
    SELECT facebook_id, groupArray(post_id) AS top_post_ids
    FROM (
        SELECT facebook_id, post_id
        FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        ORDER BY post_engagement DESC
        LIMIT 5 BY facebook_id
    )
    GROUP BY facebook_id
),
media_assets AS (
    SELECT *
    FROM facebook_competitor_media_assets
    WHERE %s AND %s
    ORDER BY inserted_at DESC
    LIMIT 1 BY post_id
)
SELECT * FROM (
    SELECT *
    FROM (
        SELECT *,
            'top_5_posts' as category,
            'https://graph.facebook.com/' || facebook_id || '/picture?type=large' as image
        FROM (
            SELECT *
            FROM facebook_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            ORDER BY post_engagement DESC
            LIMIT 5 BY facebook_id
        ) as post_data
        LEFT JOIN (
            SELECT * FROM media_assets
        ) as media_assets USING post_id
    ) as top_posts
    UNION ALL (
        SELECT *,
            'least_5_posts' as category,
            'https://graph.facebook.com/' || facebook_id || '/picture?type=large' as image
        FROM (
            SELECT p.*
            FROM facebook_competitor_posts AS p
            LEFT JOIN top_posts_ids USING (facebook_id)
            WHERE (p.post_id, p.inserted_at) IN (posts)
              AND NOT has(ifNull(top_post_ids, []), p.post_id)
            ORDER BY post_engagement ASC
            LIMIT 5 BY facebook_id
        ) as post_data
        LEFT JOIN (
            SELECT * FROM media_assets
        ) as media_assets USING post_id
    )
)`,
		pageFilter, dateFilt,
		insightsPageFilter, dateFilterField("created_at", params.StartDate, params.EndDate),
	)

	return r.queryRows(ctx, query)
}

// GetTopHashtags returns the top N hashtags across all competitors.
func (r *Repository) GetTopHashtags(ctx context.Context, params *CompetitorQueryParams, limit int) ([]map[string]interface{}, error) {
	pageFilter := accountFilter("facebook_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)
	accountIds := accountValues(params.PageIDs)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    groupUniqArray(hashtags_business_ids) as companies_using,
    groupUniqArray(hashtags_name) as companies_name,
    sum(hashtag_count) as count,
    sum(hashtags_total_engagement) as total_engagement,
    groupUniqArray(followers_total_count) as total_followers,
    round(sum(engagement_per_follower),2) as engagement_per_follower,
    round(sum(engagement_rate_by_follower),2) as engagement_rate_by_follower,
    round(sum(engagement_per_post),2) as engagement_per_post,
    tag
FROM (
    SELECT
        hashtags.facebook_id as hashtags_business_ids,
        hashtags.name as hashtags_name,
        hashtags.count as hashtag_count,
        hashtags.total_engagement as hashtags_total_engagement,
        round((hashtags.total_engagement / hashtags.count),2) as engagement_per_post,
        followers.total_followers as followers_total_count,
        round((hashtags.total_engagement / followers.total_followers),2) as engagement_per_follower,
        round((hashtags.total_engagement / followers.total_followers / hashtags.count) * 100,2) as engagement_rate_by_follower,
        hashtags.tag as tag
    FROM (
        SELECT total_followers, page_id
        FROM (
            SELECT max(followers_count) as total_followers, page_id
            FROM facebook_competitor_insights
            WHERE page_id IN %s
            GROUP BY page_id
        )
    ) as followers
    LEFT JOIN (
        SELECT
            page_name as name,
            arrayJoin(hashtags) as tag,
            uniq(post_id) as count,
            facebook_id as facebook_id,
            sum(post_engagement) as total_engagement
        FROM facebook_competitor_posts
        WHERE length(hashtags) > 0 AND (post_id, inserted_at) IN (posts)
        GROUP BY tag, facebook_id, page_name
        ORDER BY count DESC
    ) as hashtags ON followers.page_id = hashtags.facebook_id
)
WHERE tag != ''
GROUP BY tag
ORDER BY count DESC
LIMIT %d`,
		pageFilter, dateFilt,
		accountIds,
		limit,
	)

	return r.queryRows(ctx, query)
}

// GetIndividualHashtagData returns per-competitor data for a specific hashtag.
func (r *Repository) GetIndividualHashtagData(ctx context.Context, params *CompetitorQueryParams, hashtag string) ([]map[string]interface{}, error) {
	pageFilter := accountFilter("facebook_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT *
FROM (
    SELECT
        arrayJoin(hashtags) as tag,
        uniq(post_id) as count,
        sum(post_engagement) as total_engagement,
        max(followers_count) as total_followers,
        round((total_engagement / count), 2) as engagement_per_post,
        round((total_engagement / total_followers), 2) as engagement_per_follower,
        round((total_engagement / total_followers / count) * 100, 2) as engagement_rate_by_follower,
        facebook_id
    FROM facebook_competitor_posts
    WHERE length(hashtags) > 0 AND tag = '%s' AND (post_id, inserted_at) IN (posts)
    GROUP BY tag, facebook_id
    ORDER BY total_engagement DESC
) as hashtags_statistics
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        %s as slug,
        argMax(followers_count, inserted_at) as followersCount,
        page_id
    FROM facebook_competitor_insights
    GROUP BY page_id
) as page_constants ON page_id = hashtags_statistics.facebook_id`,
		pageFilter, dateFilt,
		sanitizeString(hashtag),
		constantConditions(params.Accounts, "slug", "page_id"),
	)

	return r.queryRows(ctx, query)
}

// GetBiographyData returns biography data per competitor.
func (r *Repository) GetBiographyData(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	pageFilter := accountFilter("facebook_id", params.PageIDs)

	query := fmt.Sprintf(`SELECT * FROM (
    SELECT
        argMax(biography, inserted_at) as biography,
        lengthUTF8(biography) as biography_length,
        facebook_id,
        %s as state,
        %s as slug
    FROM facebook_competitor_posts
    WHERE %s
    GROUP BY facebook_id
) as biography_statistics
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        argMax(followers_count, inserted_at) as followersCount,
        page_id
    FROM facebook_competitor_insights
    GROUP BY page_id
) as page_constants ON page_id = biography_statistics.facebook_id
ORDER BY %s DESC`,
		constantConditions(params.Accounts, "state", "facebook_id"),
		constantConditions(params.Accounts, "slug", "facebook_id"),
		pageFilter,
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}

// GetFollowersGrowthComparison returns per-competitor follower growth time-series.
func (r *Repository) GetFollowersGrowthComparison(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	accountIds := accountValues(params.PageIDs)
	insightsPageFilter := accountFilter("page_id", params.PageIDs)
	insightsDateFilt := dateFilter(params.StartDate, params.EndDate, true)

	query := fmt.Sprintf(`SELECT
    facebook_id,
    %s as name,
    %s as image,
    %s as state,
    %s as slug,
    dates, followers_count, dates_with_followers_count
FROM (
    SELECT c1 as facebook_id FROM VALUES %s
) as page_ids
LEFT JOIN (
    SELECT
        facebook_id,
        groupArray(date) as dates,
        groupArray(followers_count) as followers_count,
        arrayZip(dates, followers_count) as dates_with_followers_count
    FROM (
        SELECT
            page_name, profile_picture_url,
            page_id as facebook_id,
            toDate(inserted_at) as date,
            toInt32(followers_count) as followers_count
        FROM facebook_competitor_insights
        WHERE %s AND %s
        ORDER BY date ASC
    )
    GROUP BY facebook_id
) as page_insights ON page_ids.facebook_id = page_insights.facebook_id
ORDER BY %s ASC`,
		constantConditions(params.Accounts, "name", "page_ids.facebook_id"),
		constantConditions(params.Accounts, "image", "page_ids.facebook_id"),
		constantConditions(params.Accounts, "state", "page_ids.facebook_id"),
		constantConditions(params.Accounts, "slug", "page_ids.facebook_id"),
		accountIds,
		insightsPageFilter, insightsDateFilt,
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}

// GetPostReactDistribution returns engagement aggregates for a single competitor.
func (r *Repository) GetPostReactDistribution(ctx context.Context, params *CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error) {
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE facebook_id = '%s' AND %s
    GROUP BY post_id
)
SELECT
    facebook_id as facebook_id,
    concat('https://facebook.com/',facebook_id,'/picture') as imageUrl,
    page_name as page_name,
    toInt32(round(sum(total_engagement), 2)) as TotalEngagements,
    toInt32(sum(page_post_count)) as totalPosts
FROM (
    SELECT
        facebook_id, page_name,
        sum(count) as page_post_count,
        groupArray(total_post_engagement),
        if(page_post_count <= 0, 0, round(sum(total_post_engagement), 2)) as total_engagement,
        if(page_post_count <= 0, 0, round(((sum(er)/page_post_count)/max(followers_count))*100, 2)) as er
    FROM (
        SELECT
            facebook_id, page_name,
            count() as count,
            sum(post_engagement) as total_post_engagement,
            total_post_engagement as er,
            argMax(followers_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY facebook_id, page_name, week
        ORDER BY week ASC
        WITH FILL FROM toStartOfWeek(toDate('%s')) TO toStartOfWeek(toDate('%s')) STEP INTERVAL 1 WEEK
    )
    GROUP BY facebook_id, page_name
)
WHERE facebook_id IS NOT NULL AND facebook_id!=''
GROUP BY facebook_id, page_name`,
		sanitizeString(facebookID), dateFilt,
		params.StartDate[:10], params.EndDate[:10],
	)

	return r.queryRows(ctx, query)
}

// GetPostReactDistributionByCompany returns reaction breakdown for a single competitor.
func (r *Repository) GetPostReactDistributionByCompany(ctx context.Context, params *CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error) {
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts AS (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE facebook_id = '%s' AND %s
    GROUP BY post_id
)
SELECT
    facebook_id,
    page_name,
    'https://graph.facebook.com/' || facebook_id || '/picture' as image,
    toInt32(sum(like)) as total_likes,
    toInt32(sum(haha)) as total_hahas,
    toInt32(sum(angry)) as total_angry,
    toInt32(sum(sad)) as total_sad,
    toInt32(sum(thankful)) as total_thankful,
    toInt32(sum(love)) as total_love,
    toInt32(sum(wow)) as total_wow,
    toInt32(sum(total_post_reactions)) as total_post_reactions,
    toInt32(sum(comments)) as comments,
    toInt32(sum(shares)) as shares,
    toInt32(count()) as total_posts
FROM facebook_competitor_posts
WHERE (post_id, inserted_at) IN (posts)
GROUP BY facebook_id, page_name`,
		sanitizeString(facebookID), dateFilt,
	)

	return r.queryRows(ctx, query)
}

// GetPostTypeDistribution returns post type distribution per competitor.
func (r *Repository) GetPostTypeDistribution(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	pageFilter := accountFilter("facebook_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    facebook_id,
    page_name as page_name,
    'https://graph.facebook.com/' || facebook_id || '/picture' as image,
    media_type as mediaType,
    toInt32(round(sum(total_engagement), 2)) as TotalEngagements,
    toInt32(sum(page_post_count)) as totalPosts,
    if(dateDiff('week', toDate('%s'), toDate('%s')) = 0, 0,
        toInt32(round(totalPosts/dateDiff('week', toDate('%s'), toDate('%s')), 2))) as postsPerWeek,
    if(dateDiff('day', toDate('%s'), toDate('%s')) = 0, 0,
        toInt32(round(totalPosts/dateDiff('day', toDate('%s'), toDate('%s')), 2))) as postsPerDay,
    if(dateDiff('hour', toDate('%s'), toDate('%s')) = 0, 0,
        toInt32(round(totalPosts/dateDiff('hour', toDate('%s'), toDate('%s')), 2))) as postsPerHour,
    toInt32(dateDiff('week', toDate('%s'), toDate('%s'))) as weekCount,
    toInt32(dateDiff('day', toDate('%s'), toDate('%s'))) as dayCount,
    toInt32(dateDiff('hour', toDate('%s'), toDate('%s'))) as hourCount
FROM (
    SELECT
        facebook_id, page_name, media_type,
        sum(count) as page_post_count,
        groupArray(total_post_engagement),
        if(page_post_count <= 0, 0, round(sum(total_post_engagement), 2)) as total_engagement
    FROM (
        SELECT
            facebook_id, page_name, media_type,
            count() as count,
            sum(post_engagement) as total_post_engagement,
            total_post_engagement as er,
            argMax(followers_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        FROM facebook_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY facebook_id, page_name, media_type, week
        ORDER BY week ASC
        WITH FILL FROM toStartOfWeek(toDate('%s')) TO toStartOfWeek(toDate('%s')) STEP INTERVAL 1 WEEK
    )
    WHERE media_type!=''
    GROUP BY facebook_id, page_name, media_type
)
GROUP BY mediaType, facebook_id, page_name`,
		pageFilter, dateFilt,
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
	)

	return r.queryRows(ctx, query)
}

// GetPostEngagementOverTime returns daily engagement for a single competitor.
func (r *Repository) GetPostEngagementOverTime(ctx context.Context, params *CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error) {
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts AS (
    SELECT post_id, max(inserted_at)
    FROM facebook_competitor_posts
    WHERE facebook_id = '%s' AND %s
    GROUP BY post_id
)
SELECT
    toInt32(sum(post_engagement)) as total_engagements,
    toInt32(count()) as total_posts,
    toDate(created_at) as date
FROM facebook_competitor_posts
WHERE (post_id, inserted_at) IN (posts)
GROUP BY toDate(created_at)
ORDER BY toDate(created_at) ASC WITH FILL FROM toDate('%s') TO toDate('%s') STEP 1`,
		sanitizeString(facebookID), dateFilt,
		params.StartDate[:10], params.EndDate[:10],
	)

	return r.queryRows(ctx, query)
}

// GetPostEngagementByCompetitor returns total engagement per competitor.
func (r *Repository) GetPostEngagementByCompetitor(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	pageFilter := accountFilter("facebook_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`SELECT
    facebook_id,
    page_name,
    'https://graph.facebook.com/' || facebook_id || '/picture' as image,
    toInt32(count()) as total_posts,
    toInt32(sum(post_engagement)) as total_engagements
FROM (
    SELECT
        facebook_id, page_name, post_id,
        max(post_engagement) as post_engagement
    FROM facebook_competitor_posts
    WHERE %s AND %s
    GROUP BY facebook_id, page_name, post_id
)
GROUP BY facebook_id, page_name
ORDER BY total_engagements DESC`,
		pageFilter, dateFilt,
	)

	return r.queryRows(ctx, query)
}

// queryRows executes a query and returns results as []map[string]interface{}.
// This matches the PHP pattern of returning raw associative arrays.
func (r *Repository) queryRows(ctx context.Context, query string) ([]map[string]interface{}, error) {
	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("fb_competitor query failed: %w", err)
	}
	defer rows.Close()

	cols := rows.ColumnTypes()
	colNames := make([]string, len(cols))
	scanTypes := make([]reflect.Type, len(cols))
	for i, c := range cols {
		colNames[i] = c.Name()
		scanTypes[i] = c.ScanType()
	}

	var results []map[string]interface{}
	for rows.Next() {
		valuePtrs := make([]interface{}, len(colNames))
		for i, st := range scanTypes {
			valuePtrs[i] = reflect.New(st).Interface()
		}
		if err := rows.Scan(valuePtrs...); err != nil {
			return nil, fmt.Errorf("fb_competitor scan failed: %w", err)
		}
		row := make(map[string]interface{}, len(colNames))
		for i, name := range colNames {
			row[name] = reflect.ValueOf(valuePtrs[i]).Elem().Interface()
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("fb_competitor rows error: %w", err)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
}

// sanitizeString removes single quotes from a string to prevent SQL injection.
func sanitizeString(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

// sanitizeOrderBy strips backtick, double-quote, and single-quote from ORDER BY fields.
func sanitizeOrderBy(orderBy string) string {
	replacer := strings.NewReplacer("`", "", "\"", "", "'", "")
	return replacer.Replace(orderBy)
}
