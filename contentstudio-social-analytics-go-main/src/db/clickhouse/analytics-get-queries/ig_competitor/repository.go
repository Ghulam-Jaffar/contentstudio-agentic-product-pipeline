// Package ig_competitor provides ClickHouse queries for Instagram competitor analytics.
// Migrated from PHP InstagramCompetitorBuilder (contentstudio-backend).
// Tables: instagram_competitor_posts, instagram_competitor_insights.
package ig_competitor

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// Repository provides Instagram competitor analytics queries against ClickHouse.
type Repository struct {
	client *ch.Client
}

// NewRepository creates a new Instagram competitor analytics repository.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// CompetitorQueryParams holds Instagram competitor query parameters.
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

// AccountInfo holds metadata fetched from the competitor report.
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

// accountFilter formats page IDs for WHERE clauses.
func accountFilter(field string, ids []string) string {
	return fmt.Sprintf("%s IN %s", field, ch.FormatAccountIDs(ids))
}

// accountValues formats page IDs as VALUES for subquery.
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

// dateFilter returns a BETWEEN clause for posts (created_at) or insights (inserted_at).
func dateFilter(startDate, endDate string, insightsFlag bool) string {
	field := "created_at"
	if insightsFlag {
		field = "inserted_at"
	}
	return fmt.Sprintf("%s BETWEEN toDateTime('%s',0) AND toDateTime('%s',0)", field, startDate, endDate)
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

// sanitizeOrderBy strips dangerous characters from ORDER BY fields.
func sanitizeOrderBy(orderBy string) string {
	replacer := strings.NewReplacer("`", "", "\"", "", "'", "")
	return replacer.Replace(orderBy)
}

// sanitizeString escapes single quotes to prevent SQL injection.
func sanitizeString(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}

// queryRows executes a query and returns results as []map[string]interface{}.
func (r *Repository) queryRows(ctx context.Context, query string) ([]map[string]interface{}, error) {
	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("ig_competitor query failed: %w", err)
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
			return nil, fmt.Errorf("ig_competitor scan failed: %w", err)
		}
		row := make(map[string]interface{}, len(colNames))
		for i, name := range colNames {
			row[name] = reflect.ValueOf(valuePtrs[i]).Elem().Interface()
		}
		results = append(results, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("ig_competitor rows error: %w", err)
	}

	if results == nil {
		results = []map[string]interface{}{}
	}
	return results, nil
}

// GetDataTableMetrics returns per-competitor table metrics for the date range.
func (r *Repository) GetDataTableMetrics(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	accountIds := accountValues(params.PageIDs)
	pageFilter := accountFilter("business_account_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)
	insightsDateFilt := dateFilter(params.StartDate, params.EndDate, true)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    %s as image,
    %s as name,
    %s as state,
    %s as slug,
    * FROM (
    SELECT
        pages_ids.business_account_id as business_account_id,
        week_metrics.averageEngagement as averageEngagement,
        week_metrics.averagePostsPerWeek as averagePostsPerWeek,
        week_metrics.engagementRate as engagementRate,
        days_metrics.dayOfWeek as dayOfWeek,
        days_metrics.hourOfDay as hourOfDay,
        days_metrics.averagePostsPerDay as averagePostsPerDay,
        days_metrics.averagePostsPerDayEngagement as averagePostsPerDayEngagement
    FROM (
        SELECT CAST(c1 AS String) as business_account_id FROM VALUES %s
    ) as pages_ids
    LEFT JOIN (
        SELECT
            business_account_id,
            ceil(avg(total_engagement)) as averageEngagement,
            round(sum(posts_in_a_week)/dateDiff('week', toDate('%s'), toDate('%s')),2) as averagePostsPerWeek,
            round((sum(total_engagement)/sum(posts_in_a_week)) / max(total_followed_by_count) * 100, 2) as engagementRate
        FROM (
            SELECT
                business_account_id,
                sum(engagement) as total_engagement,
                toStartOfWeek(created_at) as week,
                count() as posts_in_a_week,
                argMax(total_followed_by_count, created_at) as total_followed_by_count
            FROM instagram_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            GROUP BY business_account_id, week
            ORDER BY week ASC
            WITH FILL FROM toStartOfWeek(toDate('%s')) TO toStartOfWeek(toDate('%s')) STEP INTERVAL 1 WEEK
        )
        GROUP BY business_account_id
    ) as week_metrics ON week_metrics.business_account_id = pages_ids.business_account_id
    LEFT JOIN (
        SELECT
            business_account_id,
            max(username) as slug,
            argMax(%s, total_posts) as dayOfWeek,
            argMax(%s, total_posts) as hourOfDay,
            ceil(avg(total_posts)) as averagePostsPerDay,
            ceil(avg(total_engagement)) as averagePostsPerDayEngagement,
            argMax(hour_of_day, total_posts) as maximumPostsHour,
            argMax(hour_of_day, total_engagement) as mostEngagementHour
        FROM (
            SELECT
                business_account_id,
                max(username) as username,
                sum(engagement) as total_engagement,
                toDayOfWeek(created_at) as day_of_week,
                toHour(created_at) as hour_of_day,
                uniq(post_id) as total_posts,
                max(total_followed_by_count) as total_followed_by_count,
                toDate(created_at) as date_c
            FROM instagram_competitor_posts
            WHERE (post_id, inserted_at) IN (posts)
            GROUP BY business_account_id, day_of_week, hour_of_day, date_c
            ORDER BY date_c ASC WITH FILL FROM (toDate('%s')) TO (toDate('%s')) STEP 1
        )
        GROUP BY business_account_id
    ) as days_metrics ON days_metrics.business_account_id = pages_ids.business_account_id
) as page_metrics
LEFT JOIN (
    SELECT
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    WHERE %s
    GROUP BY instagram_account_id
) as page_metadata ON page_metrics.business_account_id = page_metadata.instagram_account_id
ORDER BY %s DESC`,
		pageFilter, dateFilt,
		constantConditions(params.Accounts, "image", "business_account_id"),
		constantConditions(params.Accounts, "name", "business_account_id"),
		constantConditions(params.Accounts, "state", "business_account_id"),
		constantConditions(params.Accounts, "slug", "business_account_id"),
		accountIds,
		params.StartDate[:10], params.EndDate[:10],
		params.StartDate[:10], params.EndDate[:10],
		dayMapping("day_of_week"),
		hourMapping("hour_of_day"),
		params.StartDate[:10], params.EndDate[:10],
		insightsDateFilt,
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}

// GetPostingActivityGraphByTypes returns aggregated posting activity per media type.
func (r *Repository) GetPostingActivityGraphByTypes(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	pageFilter := accountFilter("business_account_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    post_type as mediaType,
    media_product_type as mediaProductType,
    round(sum(avg_engagement),2) as avgTotalEngagements,
    sum(page_post_count) as totalPosts,
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
        business_account_id, name,
        post_type,
        media_product_type,
        sum(count) as page_post_count,
        groupArray(total_engagement),
        if(page_post_count <= 0, 0, round(sum(total_engagement)/page_post_count,2)) as avg_engagement,
        if(page_post_count <= 0, 0, round(((sum(er)/page_post_count)/max(followers_count))*100,2)) as er
    FROM (
        SELECT
            business_account_id, name,
            if(media_product_type == 'REELS', 'VIDEO REEL', if(media_type == 'CAROUSEL_ALBUM', 'CAROUSEL ALBUM', media_type)) as post_type,
            media_product_type,
            count() as count,
            sum(engagement) as total_engagement,
            total_engagement as er,
            argMax(total_followed_by_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY business_account_id, name, post_type, media_product_type, week
        ORDER BY week ASC
        WITH FILL FROM toStartOfWeek(toDate('%s')) TO toStartOfWeek(toDate('%s')) STEP INTERVAL 1 WEEK
    )
    WHERE business_account_id != ''
    GROUP BY business_account_id, name, post_type, media_product_type
)
GROUP BY post_type, media_product_type
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
func (r *Repository) GetPostingActivityBySpecificType(ctx context.Context, params *CompetitorQueryParams, mediaType, mediaProductType, sortOrder string) ([]map[string]interface{}, error) {
	accountIds := accountValues(params.PageIDs)
	pageFilter := accountFilter("business_account_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    pages_ids.business_account_id as businessAccountId,
    %s as name,
    %s as image,
    %s as state,
    %s as slug,
    page_insights.media_type as mediaType,
    page_insights.media_product_type as mediaProductType,
    page_insights.avgTotalEngagements as avgTotalEngagements,
    page_insights.avgCountByWeek as avgCountByWeek,
    page_insights.avgCountByDay as avgCountByDay,
    page_insights.avgCountByHour as avgCountByHour,
    page_insights.totalPosts as totalPosts,
    page_insights.weekCount as weekCount,
    page_insights.dayCount as dayCount,
    page_insights.hourCount as hourCount,
    page_insights.avgEngagementRate as avgEngagementRate,
    page_constants.followingCount as followingCount,
    page_constants.followersCount as followersCount
FROM (
    SELECT CAST(c1 AS String) as business_account_id FROM VALUES %s
) as pages_ids
LEFT JOIN (
    SELECT
        if(media_product_type == 'REELS', 'VIDEO REEL', if(media_type == 'CAROUSEL_ALBUM', 'CAROUSEL ALBUM', media_type)) as media_type,
        media_product_type,
        business_account_id,
        if(totalPosts <= 0, 0, round(sum(total_engagement)/totalPosts, 2)) as avgTotalEngagements,
        if(dateDiff('week', toDate('%s'), toDate('%s')) = 0, 0, round(sum(count)/dateDiff('week', toDate('%s'), toDate('%s')),2)) as avgCountByWeek,
        if(dateDiff('day', toDate('%s'), toDate('%s')) = 0, 0, round(sum(count)/dateDiff('day', toDate('%s'), toDate('%s')),2)) as avgCountByDay,
        if(dateDiff('hour', toDate('%s'), toDate('%s')) = 0, 0, round(sum(count)/dateDiff('hour', toDate('%s'), toDate('%s')),2)) as avgCountByHour,
        sum(count) as totalPosts,
        dateDiff('week', toDate('%s'), toDate('%s')) as weekCount,
        dateDiff('day', toDate('%s'), toDate('%s')) as dayCount,
        dateDiff('hour', toDate('%s'), toDate('%s')) as hourCount,
        if(totalPosts <= 0, 0, round(((sum(total_engagement)/totalPosts/max(followers_count))*100), 2)) as avgEngagementRate
    FROM (
        SELECT
            business_account_id, media_type, media_product_type,
            count() as count,
            sum(engagement) as total_engagement,
            argMax(total_followed_by_count, created_at) as followers_count,
            toStartOfWeek(created_at) as week
        FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY business_account_id, media_type, media_product_type, week
        ORDER BY week ASC
        WITH FILL FROM toStartOfWeek(toDate('%s')) TO toStartOfWeek(toDate('%s')) STEP INTERVAL 1 WEEK
    ) as page_metrics
    WHERE media_type = '%s' AND media_product_type = '%s'
    GROUP BY business_account_id, media_type, media_product_type
) as page_insights USING business_account_id
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    GROUP BY instagram_account_id
) as page_constants ON page_constants.instagram_account_id = pages_ids.business_account_id
ORDER BY %s DESC`,
		pageFilter, dateFilt,
		constantConditions(params.Accounts, "name", "pages_ids.business_account_id"),
		constantConditions(params.Accounts, "image", "pages_ids.business_account_id"),
		constantConditions(params.Accounts, "state", "pages_ids.business_account_id"),
		constantConditions(params.Accounts, "slug", "pages_ids.business_account_id"),
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
		sanitizeString(mediaProductType),
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}

// GetPostingActivityTableByType returns table view metrics for a specific media type per competitor.
func (r *Repository) GetPostingActivityTableByType(ctx context.Context, params *CompetitorQueryParams, mediaType, mediaProductType, sortOrder string) ([]map[string]interface{}, error) {
	accountIds := accountValues(params.PageIDs)
	pageFilter := accountFilter("business_account_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT
    pages_ids.business_account_id as businessAccountId,
    %s as name,
    %s as image,
    %s as state,
    %s as slug,
    coalesce(page_insights.count, 0) as count,
    coalesce(page_insights.total_engagement, 0) as totalEngagement,
    round(if(coalesce(page_insights.count, 0) <= 0 OR coalesce(page_metadata.followersCount, 0) <= 0, 0,
        ((coalesce(page_insights.total_engagement, 0) / coalesce(page_insights.count, 1)) / coalesce(page_metadata.followersCount, 1)) * 100), 2) as engagementRate,
    coalesce(page_metadata.followingCount, 0) as followingCount,
    coalesce(page_metadata.followersCount, 0) as followersCount,
    page_insights.media_type as mediaType,
    page_insights.media_product_type as mediaProductType
FROM (
    SELECT CAST(c1 AS String) as business_account_id FROM VALUES %s
) as pages_ids
LEFT JOIN (
    SELECT
        business_account_id,
        if(media_product_type == 'REELS', 'VIDEO REEL', if(media_type == 'CAROUSEL_ALBUM', 'CAROUSEL ALBUM', media_type)) as media_type,
        media_product_type,
        max(count) as count,
        max(total_engagement) as total_engagement
    FROM (
        SELECT
            business_account_id, media_type, media_product_type,
            count() as count,
            sum(engagement) as total_engagement
        FROM instagram_competitor_posts
        WHERE (post_id, inserted_at) IN (posts)
        GROUP BY business_account_id, media_type, media_product_type
    ) as page_metrics
    WHERE media_type = '%s' AND media_product_type = '%s'
    GROUP BY business_account_id, media_type, media_product_type
) as page_insights USING business_account_id
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    GROUP BY instagram_account_id
) as page_metadata ON page_metadata.instagram_account_id = pages_ids.business_account_id
ORDER BY %s DESC`,
		pageFilter, dateFilt,
		constantConditions(params.Accounts, "name", "pages_ids.business_account_id"),
		constantConditions(params.Accounts, "image", "pages_ids.business_account_id"),
		constantConditions(params.Accounts, "state", "pages_ids.business_account_id"),
		constantConditions(params.Accounts, "slug", "pages_ids.business_account_id"),
		accountIds,
		sanitizeString(mediaType),
		sanitizeString(mediaProductType),
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}

// GetFollowersGrowthComparison returns per-competitor follower growth time-series.
func (r *Repository) GetFollowersGrowthComparison(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	accountIds := accountValues(params.PageIDs)
	insightsPageFilter := accountFilter("instagram_account_id", params.PageIDs)
	insightsDateFilt := dateFilter(params.StartDate, params.EndDate, true)

	query := fmt.Sprintf(`SELECT
    %s as name,
    %s as image,
    %s as state,
    %s as slug,
    * FROM (
    SELECT
        c1 as business_account_id
    FROM VALUES %s
) as page_ids
LEFT JOIN (
    SELECT
        instagram_account_id,
        groupArray(date) as dates,
        groupArray(total_following_count) as total_following_count,
        groupArray(total_followed_by_count) as total_followed_by_count,
        arrayZip(dates, total_following_count) as dates_with_following_count,
        arrayZip(dates, total_followed_by_count) as dates_with_followed_by_count
    FROM (
        SELECT
            page_name, profile_picture_url, instagram_account_id,
            toDate(inserted_at) as date,
            total_following_count, total_followed_by_count
        FROM instagram_competitor_insights
        WHERE %s AND %s
        GROUP BY inserted_at, instagram_account_id, profile_picture_url, page_name, total_following_count, total_followed_by_count
        ORDER BY inserted_at
    )
    GROUP BY instagram_account_id, page_name
) as page_insights ON page_ids.business_account_id = page_insights.instagram_account_id
ORDER BY %s DESC`,
		constantConditions(params.Accounts, "name", "page_ids.business_account_id"),
		constantConditions(params.Accounts, "image", "page_ids.business_account_id"),
		constantConditions(params.Accounts, "state", "page_ids.business_account_id"),
		constantConditions(params.Accounts, "slug", "page_ids.business_account_id"),
		accountIds,
		insightsPageFilter, insightsDateFilt,
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}

// GetTopAndLeastPerformingPosts returns top 5 and least 5 performing posts per competitor.
func (r *Repository) GetTopAndLeastPerformingPosts(ctx context.Context, params *CompetitorQueryParams) ([]map[string]interface{}, error) {
	accountIds := accountValues(params.PageIDs)
	pageFilter := accountFilter("business_account_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT * FROM (
    SELECT * FROM (
        SELECT
            c1 as business_account_id,
            %s as state,
            %s as slug
        FROM VALUES %s
    ) as pages_ids
    LEFT JOIN (
        SELECT * FROM (
            SELECT
                business_account_id,
                arraySlice(arraySort(x -> -x.1, groupedPosts), 1, 5) as top_5_posts,
                arraySlice(
                    arrayFilter(
                        x -> NOT has(arrayMap(y -> y.2, arraySlice(arraySort(z -> -z.1, groupedPosts), 1, 5)), x.2),
                        arraySort(x -> x.1, groupedPosts)
                    ),
                    1, 5
                ) as least_5_posts
            FROM (
                SELECT
                    business_account_id,
                    groupArray(
                        tuple(engagement, post_id, business_account_id,
                            caption, media_type, media_url, permalink,
                            created_at, like_count, comments_count,
                            media_count, media_product_type)
                    ) as groupedPosts
                FROM instagram_competitor_posts
                WHERE (post_id, inserted_at) IN (posts)
                GROUP BY business_account_id
            )
        )
    ) as posts_insights USING business_account_id
) as posts_information
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    GROUP BY instagram_account_id
) as page_constants ON posts_information.business_account_id = page_constants.instagram_account_id`,
		pageFilter, dateFilt,
		constantConditions(params.Accounts, "state", "business_account_id"),
		constantConditions(params.Accounts, "slug", "business_account_id"),
		accountIds,
	)

	return r.queryRows(ctx, query)
}

// GetTopHashtags returns top hashtags across all competitors.
func (r *Repository) GetTopHashtags(ctx context.Context, params *CompetitorQueryParams, limit int) ([]map[string]interface{}, error) {
	accountIds := accountValues(params.PageIDs)
	pageFilter := accountFilter("business_account_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
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
        hashtags.business_account_id as hashtags_business_ids,
        hashtags.name as hashtags_name,
        hashtags.count as hashtag_count,
        hashtags.total_engagement as hashtags_total_engagement,
        round((hashtags.total_engagement / hashtags.count),2) as engagement_per_post,
        followers.total_followers as followers_total_count,
        round((hashtags.total_engagement / followers.total_followers),2) as engagement_per_follower,
        round((hashtags.total_engagement / followers.total_followers / hashtags.count) * 100,2) as engagement_rate_by_follower,
        hashtags.tag as tag
    FROM (
        SELECT
            total_followers, instagram_account_id
        FROM (
            SELECT
                max(total_followed_by_count) as total_followers,
                instagram_account_id
            FROM instagram_competitor_insights
            WHERE instagram_account_id IN %s
            GROUP BY instagram_account_id
        )
    ) as followers
    LEFT JOIN (
        SELECT
            name as name,
            arrayJoin(hashtags) as tag,
            uniq(post_id) as count,
            business_account_id as business_account_id,
            sum(engagement) as total_engagement
        FROM instagram_competitor_posts
        WHERE length(hashtags) > 0 AND (post_id, inserted_at) IN (posts)
        GROUP BY tag, business_account_id, name
        ORDER BY count DESC
    ) as hashtags ON followers.instagram_account_id = hashtags.business_account_id
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
	pageFilter := accountFilter("business_account_id", params.PageIDs)
	dateFilt := dateFilter(params.StartDate, params.EndDate, false)

	query := fmt.Sprintf(`WITH posts as (
    SELECT post_id, max(inserted_at)
    FROM instagram_competitor_posts
    WHERE %s AND %s
    GROUP BY post_id
)
SELECT * FROM (
    SELECT
        arrayJoin(hashtags) as tag,
        uniq(post_id) as count,
        sum(engagement) as total_engagement,
        max(total_followed_by_count) as total_followers,
        round((total_engagement / count), 2) as engagement_per_post,
        round((total_engagement / total_followers), 2) as engagement_per_follower,
        round((total_engagement / total_followers / count) * 100, 2) as engagement_rate_by_follower,
        business_account_id
    FROM instagram_competitor_posts
    WHERE length(hashtags) > 0 AND tag = '%s' AND (post_id, inserted_at) IN (posts)
    GROUP BY tag, business_account_id
    ORDER BY total_engagement DESC
) as hashtags_statistics
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        %s as slug,
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    GROUP BY instagram_account_id
) as page_constants ON instagram_account_id = hashtags_statistics.business_account_id`,
		pageFilter, dateFilt,
		sanitizeString(hashtag),
		constantConditions(params.Accounts, "slug", "instagram_account_id"),
	)

	return r.queryRows(ctx, query)
}

// GetBiographyData returns biography data per competitor.
func (r *Repository) GetBiographyData(ctx context.Context, params *CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error) {
	pageFilter := accountFilter("business_account_id", params.PageIDs)

	query := fmt.Sprintf(`SELECT * FROM (
    SELECT
        argMax(biography, inserted_at) as biography,
        lengthUTF8(biography) as biography_length,
        business_account_id,
        %s as state,
        %s as slug
    FROM instagram_competitor_posts
    WHERE %s
    GROUP BY business_account_id
) as biography_statistics
LEFT JOIN (
    SELECT
        argMax(profile_picture_url, inserted_at) as image,
        argMax(page_name, inserted_at) as name,
        argMax(total_following_count, inserted_at) as followingCount,
        argMax(total_followed_by_count, inserted_at) as followersCount,
        instagram_account_id
    FROM instagram_competitor_insights
    GROUP BY instagram_account_id
) as page_constants ON instagram_account_id = biography_statistics.business_account_id
ORDER BY %s DESC`,
		constantConditions(params.Accounts, "state", "business_account_id"),
		constantConditions(params.Accounts, "slug", "business_account_id"),
		pageFilter,
		sanitizeOrderBy(sortOrder),
	)

	return r.queryRows(ctx, query)
}
