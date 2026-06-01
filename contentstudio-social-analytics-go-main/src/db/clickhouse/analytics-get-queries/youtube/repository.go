package youtube

import (
	"context"
	"fmt"
	"math"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// Repository executes ClickHouse queries for YouTube analytics.
type Repository struct {
	client *ch.Client
}

// NewRepository returns a new Repository backed by the given ClickHouse client.
func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// ytDateFilter returns a WHERE clause fragment using toDate() comparisons.
// YouTube tables store dates without time components, so toDate() is used instead of toDateTime().
func ytDateFilter(field string, params *ch.QueryParams) string {
	return fmt.Sprintf(
		"toDate(%s, '%s') BETWEEN toDate('%s') AND toDate('%s')",
		field,
		params.Timezone,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)
}

// ytDateFill returns a WITH FILL clause for daily time-series gap filling.
// The TO date is end+1 day so the end date is included in the fill range.
func ytDateFill(params *ch.QueryParams) string {
	return fmt.Sprintf(
		"WITH FILL FROM toDate('%s') TO toDate('%s') STEP 1",
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.AddDate(0, 0, 1).Format("2006-01-02"),
	)
}

// ytMonthFill returns a WITH FILL clause for monthly time-series gap filling.
func ytMonthFill(params *ch.QueryParams) string {
	return fmt.Sprintf(
		"WITH FILL FROM toStartOfMonth(toDate('%s')) TO toStartOfMonth(toDate('%s')) STEP toIntervalMonth(1)",
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.AddDate(0, 1, 0).Format("2006-01-02"),
	)
}

// periodFnDaily returns the SQL expression to bucket a timestamp by day.
func periodFnDaily(field, tz string) string {
	return fmt.Sprintf("toDate(%s, '%s')", field, tz)
}

// periodFnMonthly returns the SQL expression to bucket a timestamp by month.
func periodFnMonthly(field, tz string) string {
	return fmt.Sprintf("toStartOfMonth(toDate(%s, '%s'))", field, tz)
}

// GetActivitySummary returns aggregated metrics from youtube_activity_insights for the given period.
func (r *Repository) GetActivitySummary(ctx context.Context, params *ch.QueryParams) (*ActivitySummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("created_at", params)

	query := fmt.Sprintf(`
		SELECT
			watch_time, avg_view_duration, likes, dislikes, comments, shares,
			likes + dislikes + comments + shares AS engagement,
			views
		FROM (
			SELECT
				toInt64(sum(estimated_minutes_watched)) AS watch_time,
				round(ifNotFinite(avg(average_view_duration), 0), 2) AS avg_view_duration,
				toInt64(sum(likes))                     AS likes,
				toInt64(sum(dislikes))                  AS dislikes,
				toInt64(sum(comments))                  AS comments,
				toInt64(sum(shares))                    AS shares,
				toInt64(sum(views))                     AS views
			FROM (
				SELECT channel_id, created_at,
					argMax(estimated_minutes_watched, created_at) AS estimated_minutes_watched,
					argMax(average_view_duration, created_at)     AS average_view_duration,
					argMax(views, created_at)                     AS views,
					argMax(likes, created_at)                     AS likes,
					argMax(dislikes, created_at)                  AS dislikes,
					argMax(comments, created_at)                  AS comments,
					argMax(shares, created_at)                    AS shares
				FROM youtube_activity_insights
				WHERE channel_id IN %s
				  AND %s
				GROUP BY record_id, channel_id, created_at
			)
		)`,
		ids, dateFilter,
	)

	var result ActivitySummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.WatchTime,
		&result.AvgViewDuration,
		&result.Likes,
		&result.Dislikes,
		&result.Comments,
		&result.Shares,
		&result.Engagement,
		&result.Views,
	)
	return &result, err
}

// GetSubscriberSummary returns the latest subscriber count from youtube_channels.
func (r *Repository) GetSubscriberSummary(ctx context.Context, params *ch.QueryParams) (*SubscriberSummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("inserted_at", params)

	query := fmt.Sprintf(`
		SELECT toInt64(argMax(subscriber_count, inserted_at)) AS subscribers
		FROM youtube_channels
		WHERE channel_id IN %s
		  AND %s`,
		ids, dateFilter,
	)

	var result SubscriberSummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(&result.Subscribers)
	return &result, err
}

// GetVideoCount returns the number of distinct videos published in the given period.
func (r *Repository) GetVideoCount(ctx context.Context, params *ch.QueryParams) (*VideoCountResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("published_at", params)

	query := fmt.Sprintf(`
		SELECT toInt64(count(DISTINCT video_id)) AS video_count
		FROM youtube_videos
		WHERE channel_id IN %s
		  AND %s`,
		ids, dateFilter,
	)

	var result VideoCountResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(&result.VideoCount)
	return &result, err
}

// GetSubscriberTrend returns time-series subscriber data.
// When daily is true the buckets are per-day; when false they are per-month.
func (r *Repository) GetSubscriberTrend(ctx context.Context, params *ch.QueryParams, daily bool) (*SubscriberTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)

	var bucketExpr, fillClause string
	if daily {
		bucketExpr = periodFnDaily("inserted_at", params.Timezone)
		fillClause = ytDateFill(params)
	} else {
		bucketExpr = periodFnMonthly("inserted_at", params.Timezone)
		fillClause = ytMonthFill(params)
	}

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0) AS show_data,
			arrayDifference(arrayFill(x -> not x == 0, groupArray(subscriber_count))) AS subscribers_gained_daily,
			arrayFill(x -> not x == 0, groupArray(subscriber_count))                 AS subscribers_total,
			groupArray(bucket)                                                         AS buckets
		FROM (
			SELECT
				%s AS bucket,
				toInt32(argMin(subscriber_count, inserted_at)) AS subscriber_count
			FROM youtube_channels
			WHERE channel_id IN %s
			  AND toDate(inserted_at, '%s') BETWEEN toDate('%s') AND toDate('%s')
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucketExpr,
		ids,
		params.Timezone,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
		fillClause,
	)

	var result SubscriberTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.SubscribersGainedDaily,
		&result.SubscribersTotal,
		&result.Buckets,
	)
	return &result, err
}

// GetLatestSubscriberCount returns the single most-recent subscriber_count row.
// Used as a fallback to back-fill leading zeros in the subscriber trend.
func (r *Repository) GetLatestSubscriberCount(ctx context.Context, params *ch.QueryParams) (*LatestSubscriberResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)

	query := fmt.Sprintf(`
		SELECT toInt32(subscriber_count) AS subscriber_count
		FROM youtube_channels
		WHERE channel_id IN %s
		ORDER BY inserted_at DESC
		LIMIT 1`,
		ids,
	)

	var result LatestSubscriberResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(&result.SubscriberCount)
	return &result, err
}

// GetEngagementTrend returns time-series engagement data with daily and cumulative totals.
// When daily is true the buckets are per-day; when false they are per-month.
func (r *Repository) GetEngagementTrend(ctx context.Context, params *ch.QueryParams, daily bool) (*EngagementTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("created_at", params)

	var bucketExpr, fillClause string
	if daily {
		bucketExpr = periodFnDaily("created_at", params.Timezone)
		fillClause = ytDateFill(params)
	} else {
		bucketExpr = periodFnMonthly("created_at", params.Timezone)
		fillClause = ytMonthFill(params)
	}

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0) AS show_data,
			groupArray(toInt32(likes_daily))    AS like_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(likes_daily))))    AS like_total,
			groupArray(toInt32(dislikes_daily)) AS dislike_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(dislikes_daily)))) AS dislike_total,
			groupArray(toInt32(shares_daily))   AS share_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(shares_daily))))   AS share_total,
			groupArray(toInt32(comments_daily)) AS comment_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(comments_daily)))) AS comment_total,
			groupArray(toInt32(likes_daily + dislikes_daily + comments_daily + shares_daily)) AS engagement_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(likes_daily + dislikes_daily + comments_daily + shares_daily)))) AS engagement_total,
			groupArray(bucket) AS buckets
		FROM (
			SELECT
				%s AS bucket,
				sum(likes)    AS likes_daily,
				sum(dislikes) AS dislikes_daily,
				sum(comments) AS comments_daily,
				sum(shares)   AS shares_daily
			FROM (
				SELECT channel_id, created_at,
					argMax(likes, created_at)    AS likes,
					argMax(dislikes, created_at) AS dislikes,
					argMax(comments, created_at) AS comments,
					argMax(shares, created_at)   AS shares
				FROM youtube_activity_insights
				WHERE channel_id IN %s
				  AND %s
				GROUP BY record_id, channel_id, created_at
			)
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucketExpr,
		ids,
		dateFilter,
		fillClause,
	)

	var result EngagementTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.LikeDaily,
		&result.LikeTotal,
		&result.DislikeDaily,
		&result.DislikeTotal,
		&result.ShareDaily,
		&result.ShareTotal,
		&result.CommentDaily,
		&result.CommentTotal,
		&result.EngagementDaily,
		&result.EngagementTotal,
		&result.Buckets,
	)
	return &result, err
}

// GetViewsTrend returns time-series view data split by subscriber and non-subscriber traffic sources.
// When daily is true the buckets are per-day; when false they are per-month.
func (r *Repository) GetViewsTrend(ctx context.Context, params *ch.QueryParams, daily bool) (*ViewsTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("created_at", params)

	var bucketExpr, fillClause string
	if daily {
		bucketExpr = periodFnDaily("created_at", params.Timezone)
		fillClause = ytDateFill(params)
	} else {
		bucketExpr = periodFnMonthly("created_at", params.Timezone)
		fillClause = ytMonthFill(params)
	}

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0) AS show_data,
			groupArray(toInt32(subscriber_views))    AS subscriber_views_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(subscriber_views))))    AS subscriber_views_total,
			groupArray(toInt32(non_subscriber_views)) AS non_subscriber_views_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(non_subscriber_views)))) AS non_subscriber_views_total,
			groupArray(toInt32(subscriber_views + non_subscriber_views)) AS video_views_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(subscriber_views + non_subscriber_views)))) AS video_views_total,
			groupArray(bucket) AS buckets
		FROM (
			SELECT
				%s AS bucket,
				sum(subscriber_views) AS subscriber_views,
				sum(paid_views + annotation_views + end_screen_views + campaign_card_view +
					no_link_other_views + yt_channel_views + yt_search_views + related_video_views +
					yt_other_page_views + ext_url_views + playlist_views + notification_views + shorts_views
				) AS non_subscriber_views
			FROM (
				SELECT channel_id, created_at,
					argMax(subscriber_views, created_at)    AS subscriber_views,
					argMax(paid_views, created_at)          AS paid_views,
					argMax(annotation_views, created_at)    AS annotation_views,
					argMax(end_screen_views, created_at)    AS end_screen_views,
					argMax(campaign_card_view, created_at)  AS campaign_card_view,
					argMax(no_link_other_views, created_at) AS no_link_other_views,
					argMax(yt_channel_views, created_at)    AS yt_channel_views,
					argMax(yt_search_views, created_at)     AS yt_search_views,
					argMax(related_video_views, created_at) AS related_video_views,
					argMax(yt_other_page_views, created_at) AS yt_other_page_views,
					argMax(ext_url_views, created_at)       AS ext_url_views,
					argMax(playlist_views, created_at)      AS playlist_views,
					argMax(notification_views, created_at)  AS notification_views,
					argMax(shorts_views, created_at)        AS shorts_views
				FROM youtube_traffic_insights
				WHERE channel_id IN %s
				  AND %s
				GROUP BY record_id, channel_id, created_at
			)
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucketExpr,
		ids,
		dateFilter,
		fillClause,
	)

	var result ViewsTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.SubscriberViewsDaily,
		&result.SubscriberViewsTotal,
		&result.NonSubscriberViewsDaily,
		&result.NonSubscriberViewsTotal,
		&result.VideoViewsDaily,
		&result.VideoViewsTotal,
		&result.Buckets,
	)
	return &result, err
}

// GetWatchTimeTrend returns time-series watch time data split by subscriber and non-subscriber.
// When daily is true the buckets are per-day; when false they are per-month.
func (r *Repository) GetWatchTimeTrend(ctx context.Context, params *ch.QueryParams, daily bool) (*WatchTimeTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("created_at", params)

	var bucketExpr, fillClause string
	if daily {
		bucketExpr = periodFnDaily("created_at", params.Timezone)
		fillClause = ytDateFill(params)
	} else {
		bucketExpr = periodFnMonthly("created_at", params.Timezone)
		fillClause = ytMonthFill(params)
	}

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0) AS show_data,
			groupArray(toInt32(subscriber_watch_time))     AS subscriber_watch_time_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(subscriber_watch_time))))     AS subscriber_watch_time_total,
			groupArray(toInt32(non_subscriber_watch_time)) AS non_subscriber_watch_time_daily,
			arrayMap(x -> toInt32(x), arrayCumSum(groupArray(toInt32(non_subscriber_watch_time)))) AS non_subscriber_watch_time_total,
			groupArray(round(toFloat64(subscriber_watch_time + non_subscriber_watch_time), 2))     AS average_watch_time,
			groupArray(bucket) AS buckets
		FROM (
			SELECT
				%s AS bucket,
				sum(subscriber_watch_time)    AS subscriber_watch_time,
				sum(non_subsciber_watch_time) AS non_subscriber_watch_time
			FROM (
				SELECT channel_id, created_at,
					argMax(subscriber_watch_time, created_at)    AS subscriber_watch_time,
					argMax(non_subsciber_watch_time, created_at) AS non_subsciber_watch_time
				FROM youtube_traffic_insights
				WHERE channel_id IN %s
				  AND %s
				GROUP BY record_id, channel_id, created_at
			)
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		bucketExpr,
		ids,
		dateFilter,
		fillClause,
	)

	var result WatchTimeTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.SubscriberWatchTimeDaily,
		&result.SubscriberWatchTimeTotal,
		&result.NonSubscriberWatchTimeDaily,
		&result.NonSubscriberWatchTimeTotal,
		&result.AverageWatchTime,
		&result.Buckets,
	)
	return &result, err
}

// GetFindVideo returns the 13 traffic sources with their view counts and percentage shares.
// Sources are ordered by view count descending; zero-count sources are excluded.
func (r *Repository) GetFindVideo(ctx context.Context, params *ch.QueryParams) ([]TrafficSourceRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("created_at", params)

	query := fmt.Sprintf(`
		WITH dedupTraffic AS (
			SELECT channel_id, created_at,
				argMax(subscriber_views, created_at)    AS subscriber_views,
				argMax(paid_views, created_at)          AS paid_views,
				argMax(annotation_views, created_at)    AS annotation_views,
				argMax(end_screen_views, created_at)    AS end_screen_views,
				argMax(campaign_card_view, created_at)  AS campaign_card_view,
				argMax(no_link_other_views, created_at) AS no_link_other_views,
				argMax(yt_channel_views, created_at)    AS yt_channel_views,
				argMax(yt_search_views, created_at)     AS yt_search_views,
				argMax(related_video_views, created_at) AS related_video_views,
				argMax(yt_other_page_views, created_at) AS yt_other_page_views,
				argMax(ext_url_views, created_at)       AS ext_url_views,
				argMax(playlist_views, created_at)      AS playlist_views,
				argMax(notification_views, created_at)  AS notification_views
			FROM youtube_traffic_insights
			WHERE channel_id IN %s
			  AND %s
			GROUP BY record_id, channel_id, created_at
		)
		SELECT name, value, round(if(total > 0, value * 100.0 / total, 0), 2) AS perc_value
		FROM (
			SELECT 'paid_views'          AS name, toInt64(sum(paid_views))          AS value FROM dedupTraffic
			UNION ALL SELECT 'annotation_views',    toInt64(sum(annotation_views))    FROM dedupTraffic
			UNION ALL SELECT 'end_screen_views',    toInt64(sum(end_screen_views))    FROM dedupTraffic
			UNION ALL SELECT 'campaign_card_view',  toInt64(sum(campaign_card_view))  FROM dedupTraffic
			UNION ALL SELECT 'subscriber_views',    toInt64(sum(subscriber_views))    FROM dedupTraffic
			UNION ALL SELECT 'no_link_other_views', toInt64(sum(no_link_other_views)) FROM dedupTraffic
			UNION ALL SELECT 'yt_channel_views',    toInt64(sum(yt_channel_views))    FROM dedupTraffic
			UNION ALL SELECT 'yt_search_views',     toInt64(sum(yt_search_views))     FROM dedupTraffic
			UNION ALL SELECT 'related_video_views', toInt64(sum(related_video_views)) FROM dedupTraffic
			UNION ALL SELECT 'yt_other_page_views', toInt64(sum(yt_other_page_views)) FROM dedupTraffic
			UNION ALL SELECT 'ext_url_views',       toInt64(sum(ext_url_views))       FROM dedupTraffic
			UNION ALL SELECT 'playlist_views',      toInt64(sum(playlist_views))      FROM dedupTraffic
			UNION ALL SELECT 'notification_views',  toInt64(sum(notification_views))  FROM dedupTraffic
		) src
		CROSS JOIN (
			SELECT toInt64(sum(
				paid_views + annotation_views + end_screen_views + campaign_card_view +
				subscriber_views + no_link_other_views + yt_channel_views + yt_search_views +
				related_video_views + yt_other_page_views + ext_url_views + playlist_views + notification_views
			)) AS total FROM dedupTraffic
		) t
		WHERE value > 0
		ORDER BY value DESC`,
		ids, dateFilter,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []TrafficSourceRow
	for rows.Next() {
		var row TrafficSourceRow
		if err := rows.Scan(&row.Name, &row.Value, &row.PercValue); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// GetVideoSharing queries the latest row from youtube_shared_insights within the date range
// and returns all 31 sharing platforms as a slice of SharingRow with percentage shares computed in Go.
// The date filter on inserted_at matches PHP's videoSharingQuery which uses DateFilter('inserted_at').
func (r *Repository) GetVideoSharing(ctx context.Context, params *ch.QueryParams) ([]SharingRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("inserted_at", params)

	query := fmt.Sprintf(`
		SELECT
			ameba, blogger, copy_paste, cyworld, digg, dropbox, embed, mail, whats_app,
			other, facebook_messenger, facebook_pages, facebook, fotka, vkontakte, google_plus,
			discord, linkedin, goo, hangouts, pinterest, myspace, reddit, skype, telegram,
			tumblr, twitter, viber, weibo, wechat, youtube
		FROM youtube_shared_insights
		WHERE channel_id IN %s
		  AND %s
		ORDER BY inserted_at DESC
		LIMIT 1`,
		ids, dateFilter,
	)

	var raw SharedInsightsRow
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&raw.Ameba, &raw.Blogger, &raw.CopyPaste, &raw.Cyworld, &raw.Digg, &raw.Dropbox,
		&raw.Embed, &raw.Mail, &raw.Whatsapp, &raw.Other, &raw.FacebookMsgr,
		&raw.FacebookPages, &raw.Facebook, &raw.Fotka, &raw.Vkontakte, &raw.GooglePlus,
		&raw.Discord, &raw.Linkedin, &raw.Goo, &raw.Hangouts, &raw.Pinterest, &raw.Myspace,
		&raw.Reddit, &raw.Skype, &raw.Telegram, &raw.Tumblr, &raw.Twitter, &raw.Viber,
		&raw.Weibo, &raw.Wechat, &raw.Youtube,
	)
	if err != nil {
		return nil, err
	}

	type entry struct {
		name  string
		value int64
	}
	entries := []entry{
		{"Ameba", raw.Ameba}, {"Blogger", raw.Blogger}, {"CopyPaste", raw.CopyPaste},
		{"Cyworld", raw.Cyworld}, {"Digg", raw.Digg}, {"Dropbox", raw.Dropbox},
		{"Embed", raw.Embed}, {"Mail", raw.Mail}, {"Whatsapp", raw.Whatsapp},
		{"Other", raw.Other}, {"FacebookMessenger", raw.FacebookMsgr},
		{"FacebookPages", raw.FacebookPages}, {"Facebook", raw.Facebook},
		{"Fotka", raw.Fotka}, {"Vkontakte", raw.Vkontakte}, {"GooglePlus", raw.GooglePlus},
		{"Discord", raw.Discord}, {"Linkedin", raw.Linkedin}, {"Goo", raw.Goo},
		{"Hangouts", raw.Hangouts}, {"Pinterest", raw.Pinterest}, {"Myspace", raw.Myspace},
		{"Reddit", raw.Reddit}, {"Skype", raw.Skype}, {"Telegram", raw.Telegram},
		{"Tumblr", raw.Tumblr}, {"Twitter", raw.Twitter}, {"Viber", raw.Viber},
		{"Weibo", raw.Weibo}, {"Wechat", raw.Wechat}, {"Youtube", raw.Youtube},
	}

	var total int64
	for _, e := range entries {
		total += e.value
	}

	results := make([]SharingRow, 0, len(entries))
	for _, e := range entries {
		var perc float64
		if total > 0 {
			perc = math.Round(float64(e.value)*100.0/float64(total)*100) / 100
		}
		results = append(results, SharingRow{
			Name:      e.name,
			Value:     e.value,
			PercValue: perc,
		})
	}
	return results, nil
}

// orderByAliasMap maps frontend order_by values (including PHP-style singular forms)
// to the SQL aliases used in the GetTopVideos SELECT clause.
var orderByAliasMap = map[string]string{
	"views":                 "views",
	"likes":                 "likes",
	"like":                  "likes",
	"dislikes":              "dislikes",
	"dislike":               "dislikes",
	"engagement":            "engagement",
	"comments":              "comments",
	"comment":               "comments",
	"shares":                "shares",
	"share":                 "shares",
	"engagement_rate":       "engagement_rate",
	"minutes_watched":       "minutes_watched",
	"average_view_duration": "avg_view_duration",
	"subscribers_gained":    "subscribers_gained",
	"published_at":          "pub_at",
}

// GetTopVideos returns videos ordered by the given metric.
// orderBy is validated against a whitelist; defaults to "views" if empty or invalid.
// ascending controls sort direction.
func (r *Repository) GetTopVideos(ctx context.Context, params *ch.QueryParams, orderBy string, limit int, ascending bool) ([]VideoRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("published_at", params)

	if alias, ok := orderByAliasMap[orderBy]; ok {
		orderBy = alias
	} else {
		orderBy = "views"
	}
	direction := "DESC"
	if ascending {
		direction = "ASC"
	}

	// Inner query uses short non-column-name aliases to avoid ClickHouse resolving
	// argMax(likes, ...) as the 'likes' aggregate alias (which would cause error 184).
	query := fmt.Sprintf(`
		SELECT
			video_id,
			_title                                                                AS title,
			_description                                                          AS description,
			_duration                                                             AS duration,
			_thumbnail_url                                                        AS thumbnail_url,
			_media_type                                                           AS media_type,
			_iframe_embed_url                                                     AS iframe_embed_url,
			_share_url                                                            AS share_url,
			toInt64(_lk + _dlk + _cm + _sh)                                      AS engagement,
			_lk                                                                   AS likes,
			_dlk                                                                  AS dislikes,
			_v                                                                    AS views,
			_rv                                                                   AS red_views,
			_fv                                                                   AS favorites,
			_cm                                                                   AS comments,
			_sg                                                                   AS subscribers_gained,
			_sh                                                                   AS shares,
			_mw                                                                   AS minutes_watched,
			_rmw                                                                  AS red_minutes_watched,
			_avd                                                                  AS avg_view_duration,
			_avp                                                                  AS avg_view_percentage,
			round(if(_rc != 0, toFloat64(_lk + _dlk + _cm + _sh) / toFloat64(_rc), 0), 2) AS engagement_rate,
			_pa                                                                   AS pub_at
		FROM (
			SELECT
				video_id,
				any(title)                                                        AS _title,
				any(description)                                                  AS _description,
				toInt64OrZero(any(duration))                                      AS _duration,
				any(thumbnail_url)                                                AS _thumbnail_url,
				any(media_type)                                                   AS _media_type,
				any(iframe_embed_html)                                            AS _iframe_embed_url,
				concat('https://www.youtube.com/watch?v=', video_id)             AS _share_url,
				toInt64(argMax(likes, inserted_at))                               AS _lk,
				toInt64(argMax(dislikes, inserted_at))                            AS _dlk,
				toInt64(argMax(views, inserted_at))                               AS _v,
				toInt64(argMax(red_views, inserted_at))                           AS _rv,
				toInt64(argMax(favorites, inserted_at))                           AS _fv,
				toInt64(argMax(comments, inserted_at))                            AS _cm,
				toInt64(argMax(subscribers_gained, inserted_at))                  AS _sg,
				toInt64(argMax(shares, inserted_at))                              AS _sh,
				toInt64(argMax(minutes_watched, inserted_at))                     AS _mw,
				toInt64(argMax(red_minutes_watched, inserted_at))                 AS _rmw,
				round(toFloat64(argMax(average_view_duration, created_at)), 2)   AS _avd,
				round(argMax(average_view_percentage, inserted_at), 2)            AS _avp,
				toInt64(count())                                                   AS _rc,
				argMin(published_at, inserted_at)                                  AS _pa
			FROM youtube_videos
			WHERE channel_id IN %s
			  AND %s
			GROUP BY video_id
		)
		ORDER BY %s %s
		LIMIT %d`,
		ids, dateFilter, orderBy, direction, limit,
	)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []VideoRow
	for rows.Next() {
		var row VideoRow
		if err := rows.Scan(
			&row.VideoID, &row.Title, &row.Description, &row.Duration,
			&row.ThumbnailURL, &row.MediaType, &row.IframeEmbedURL, &row.ShareURL,
			&row.Engagement, &row.Likes, &row.Dislikes, &row.Views, &row.RedViews,
			&row.Favorites, &row.Comments, &row.SubscribersGained, &row.Shares,
			&row.MinutesWatched, &row.RedMinutesWatched, &row.AvgViewDuration,
			&row.AvgViewPercentage, &row.EngagementRate, &row.PublishedAt,
		); err != nil {
			return nil, err
		}
		results = append(results, row)
	}
	return results, rows.Err()
}

// GetPerformanceEngagement returns time-series video engagement metrics grouped by publish date.
func (r *Repository) GetPerformanceEngagement(ctx context.Context, params *ch.QueryParams) (*PerformanceEngagementResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("published_at", params)
	fill := ytDateFill(params)

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0)              AS show_data,
			groupArray(bucket)                  AS buckets,
			groupArray(toInt32(cnt))            AS count,
			groupArray(toInt32(likes_sum))      AS likes,
			groupArray(toInt32(dislikes_sum))   AS dislikes,
			groupArray(toInt32(shares_sum))     AS shares,
			groupArray(toInt32(comments_sum))   AS comments,
			groupArray(toInt32(engagement_sum)) AS engagement
		FROM (
			SELECT
				bucket,
				count()                                                      AS cnt,
				sum(likes)                                                   AS likes_sum,
				sum(dislikes)                                                AS dislikes_sum,
				sum(shares)                                                  AS shares_sum,
				sum(comments)                                                AS comments_sum,
				sum(likes) + sum(dislikes) + sum(shares) + sum(comments)    AS engagement_sum
			FROM (
				SELECT
					toDate(published_at, '%s')    AS bucket,
					argMax(likes, inserted_at)    AS likes,
					argMax(dislikes, inserted_at) AS dislikes,
					argMax(shares, inserted_at)   AS shares,
					argMax(comments, inserted_at) AS comments
				FROM youtube_videos
				WHERE channel_id IN %s
				  AND %s
				GROUP BY video_id, channel_id, published_at
			)
			GROUP BY bucket
			ORDER BY bucket ASC
			%s
		)`,
		params.Timezone, ids, dateFilter, fill,
	)

	var result PerformanceEngagementResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.Buckets,
		&result.Count,
		&result.Likes,
		&result.Dislikes,
		&result.Shares,
		&result.Comments,
		&result.Engagement,
	)
	return &result, err
}

// GetPerformanceViews returns time-series video view metrics grouped by publish date,
// including subscriber vs non-subscriber view breakdown from youtube_traffic_insights.
func (r *Repository) GetPerformanceViews(ctx context.Context, params *ch.QueryParams) (*PerformanceViewsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ytDateFilter("published_at", params)
	trafficDateFilter := ytDateFilter("created_at", params)
	fill := ytDateFill(params)

	query := fmt.Sprintf(`
		SELECT
			if(count() > 0, 1, 0)                                       AS show_data,
			groupArray(toInt32(video_count))                             AS count,
			groupArray(toInt32(coalesce(subscriber_views, 0)))           AS subscriber_views,
			groupArray(toInt32(coalesce(non_subscriber_views, 0)))       AS non_subscriber_views,
			groupArray(bucket)                                           AS buckets
		FROM (
			SELECT
				videos.bucket            AS bucket,
				videos.video_count       AS video_count,
				traffic.subscriber_views AS subscriber_views,
				traffic.non_subscriber_views AS non_subscriber_views
			FROM (
				SELECT
					toDate(published_at, '%s') AS bucket,
					count()                    AS video_count
				FROM youtube_videos
				WHERE channel_id IN %s
				  AND %s
				GROUP BY bucket
				ORDER BY bucket ASC
				%s
			) AS videos
			LEFT JOIN (
				SELECT
					toDate(created_at, '%s')                 AS bucket,
					toInt32(sum(subscriber_views))           AS subscriber_views,
					toInt32(sum(
						paid_views + annotation_views + end_screen_views + campaign_card_view +
						no_link_other_views + yt_channel_views + yt_search_views + related_video_views +
						yt_other_page_views + ext_url_views + playlist_views + notification_views + shorts_views
					))                                       AS non_subscriber_views
				FROM (
					SELECT channel_id, created_at,
						argMax(subscriber_views, created_at)    AS subscriber_views,
						argMax(paid_views, created_at)          AS paid_views,
						argMax(annotation_views, created_at)    AS annotation_views,
						argMax(end_screen_views, created_at)    AS end_screen_views,
						argMax(campaign_card_view, created_at)  AS campaign_card_view,
						argMax(no_link_other_views, created_at) AS no_link_other_views,
						argMax(yt_channel_views, created_at)    AS yt_channel_views,
						argMax(yt_search_views, created_at)     AS yt_search_views,
						argMax(related_video_views, created_at) AS related_video_views,
						argMax(yt_other_page_views, created_at) AS yt_other_page_views,
						argMax(ext_url_views, created_at)       AS ext_url_views,
						argMax(playlist_views, created_at)      AS playlist_views,
						argMax(notification_views, created_at)  AS notification_views,
						argMax(shorts_views, created_at)        AS shorts_views
					FROM youtube_traffic_insights
					WHERE channel_id IN %s
					  AND %s
					GROUP BY record_id, channel_id, created_at
				)
				GROUP BY bucket
			) AS traffic ON videos.bucket = traffic.bucket
		)`,
		params.Timezone, ids, dateFilter, fill,
		params.Timezone, ids, trafficDateFilter,
	)

	var result PerformanceViewsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.ShowData,
		&result.Count,
		&result.SubscriberViews,
		&result.NonSubscriberViews,
		&result.Buckets,
	)
	return &result, err
}
