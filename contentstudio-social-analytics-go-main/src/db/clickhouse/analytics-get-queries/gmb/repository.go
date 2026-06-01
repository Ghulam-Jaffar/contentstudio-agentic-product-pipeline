package gmb

import (
	"context"
	"fmt"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

type Repository struct {
	client *ch.Client
}

func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

// gmbDateFilter returns a SQL WHERE clause for gmb_id IN (...) AND date filter.
func gmbDateFilter(ids, dateFilter string) string {
	return fmt.Sprintf("gmb_id IN %s AND %s", ids, dateFilter)
}

// GetSummary returns aggregated metrics from gmb_daily_metrics, gmb_reviews, and gmb_local_posts.
func (r *Repository) GetSummary(ctx context.Context, params *ch.QueryParams) (*SummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	metricsFilter := ch.DateFilter("created_at", params)
	reviewsFilter := ch.DateFilter("created_at", params)
	postsFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    coalesce(m.total_impressions, 0) AS total_impressions,
    coalesce(m.search_impressions, 0) AS search_impressions,
    coalesce(m.maps_impressions, 0) AS maps_impressions,
    coalesce(m.website_clicks, 0) AS website_clicks,
    coalesce(m.call_clicks, 0) AS call_clicks,
    coalesce(m.direction_requests, 0) AS direction_requests,
    coalesce(m.other_actions, 0) AS other_actions,
    coalesce(rv.total_reviews, 0) AS total_reviews,
    coalesce(rv.avg_rating, 0) AS average_rating,
    coalesce(p.total_posts, 0) AS total_posts
FROM (SELECT 1 AS join_key) AS base
LEFT JOIN (
    SELECT
        1 AS join_key,
        sum(business_impressions_desktop_maps + business_impressions_desktop_search +
            business_impressions_mobile_maps + business_impressions_mobile_search) AS total_impressions,
        sum(business_impressions_desktop_search + business_impressions_mobile_search) AS search_impressions,
        sum(business_impressions_desktop_maps + business_impressions_mobile_maps) AS maps_impressions,
        sum(website_clicks) AS website_clicks,
        sum(call_clicks) AS call_clicks,
        sum(business_direction_requests) AS direction_requests,
        sum(business_conversations + business_bookings + business_food_orders + business_food_menu_clicks) AS other_actions
    FROM gmb_daily_metrics FINAL
    WHERE %s
) AS m ON base.join_key = m.join_key
LEFT JOIN (
    SELECT
        1 AS join_key,
        toInt64(count()) AS total_reviews,
        ifNotFinite(round(avg(star_rating), 2), 0) AS avg_rating
    FROM gmb_reviews FINAL
    WHERE %s
) AS rv ON base.join_key = rv.join_key
LEFT JOIN (
    SELECT
        1 AS join_key,
        toInt64(count()) AS total_posts
    FROM gmb_local_posts FINAL
    WHERE %s
) AS p ON base.join_key = p.join_key
`, gmbDateFilter(ids, metricsFilter), gmbDateFilter(ids, reviewsFilter), gmbDateFilter(ids, postsFilter))

	var result SummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalImpressions, &result.SearchImpressions, &result.MapsImpressions,
		&result.WebsiteClicks, &result.CallClicks, &result.DirectionRequests,
		&result.OtherActions, &result.TotalReviews, &result.AverageRating, &result.TotalPosts,
	)
	if err != nil {
		return nil, fmt.Errorf("GetSummary: %w", err)
	}
	return &result, nil
}

// GetImpressions returns time-series impression data grouped by date with daily values.
func (r *Repository) GetImpressions(ctx context.Context, params *ch.QueryParams) (*ImpressionsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
SELECT
    groupArray(desktop_maps) AS desktop_maps_daily,
    groupArray(desktop_search) AS desktop_search_daily,
    groupArray(mobile_maps) AS mobile_maps_daily,
    groupArray(mobile_search) AS mobile_search_daily,
    groupArray(desktop_maps + desktop_search + mobile_maps + mobile_search) AS total_impressions_daily,
    arraySum(groupArray(desktop_maps + desktop_search + mobile_maps + mobile_search)) AS show_data,
    groupArray(dates) AS buckets
FROM (
    SELECT
        toDate(created_at) AS dates,
        toInt64(sum(business_impressions_desktop_maps)) AS desktop_maps,
        toInt64(sum(business_impressions_desktop_search)) AS desktop_search,
        toInt64(sum(business_impressions_mobile_maps)) AS mobile_maps,
        toInt64(sum(business_impressions_mobile_search)) AS mobile_search
    FROM gmb_daily_metrics FINAL
    WHERE %s
    GROUP BY dates
    ORDER BY dates ASC
    %s
)
`, gmbDateFilter(ids, dateFilter), fill)

	var result ImpressionsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.DesktopMapsDaily, &result.DesktopSearchDaily,
		&result.MobileMapsDaily, &result.MobileSearchDaily,
		&result.TotalImpressionsDaily, &result.ShowData, &result.Buckets,
	)
	if err != nil {
		return nil, fmt.Errorf("GetImpressions: %w", err)
	}
	return &result, nil
}

// GetImpressionsRollup returns aggregated impression totals for a date range.
func (r *Repository) GetImpressionsRollup(ctx context.Context, params *ch.QueryParams) (*ImpressionsRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    toInt64(sum(total)) AS total_impressions,
    toInt64(sum(dm)) AS desktop_maps,
    toInt64(sum(ds)) AS desktop_search,
    toInt64(sum(mm)) AS mobile_maps,
    toInt64(sum(ms)) AS mobile_search,
    ifNotFinite(round(avg(total), 2), 0) AS avg_impressions
FROM (
    SELECT
        toDate(created_at) AS dates,
        sum(business_impressions_desktop_maps) AS dm,
        sum(business_impressions_desktop_search) AS ds,
        sum(business_impressions_mobile_maps) AS mm,
        sum(business_impressions_mobile_search) AS ms,
        dm + ds + mm + ms AS total
    FROM gmb_daily_metrics FINAL
    WHERE %s
    GROUP BY dates
)
`, gmbDateFilter(ids, dateFilter))

	var result ImpressionsRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalImpressions, &result.DesktopMaps, &result.DesktopSearch,
		&result.MobileMaps, &result.MobileSearch, &result.AvgImpressions,
	)
	if err != nil {
		return nil, fmt.Errorf("GetImpressionsRollup: %w", err)
	}
	return &result, nil
}

// GetActions returns time-series customer action data grouped by date.
func (r *Repository) GetActions(ctx context.Context, params *ch.QueryParams) (*ActionsResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
SELECT
    groupArray(calls) AS call_clicks_daily,
    groupArray(website) AS website_clicks_daily,
    groupArray(directions) AS direction_requests_daily,
    groupArray(others) AS other_actions_daily,
    arraySum(groupArray(calls + website + directions + others)) AS show_data,
    groupArray(dates) AS buckets
FROM (
    SELECT
        toDate(created_at) AS dates,
        toInt64(sum(call_clicks)) AS calls,
        toInt64(sum(website_clicks)) AS website,
        toInt64(sum(business_direction_requests)) AS directions,
        toInt64(sum(business_conversations + business_bookings + business_food_orders + business_food_menu_clicks)) AS others
    FROM gmb_daily_metrics FINAL
    WHERE %s
    GROUP BY dates
    ORDER BY dates ASC
    %s
)
`, gmbDateFilter(ids, dateFilter), fill)

	var result ActionsResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.CallClicksDaily, &result.WebsiteClicksDaily,
		&result.DirectionRequestsDaily, &result.OtherActionsDaily,
		&result.ShowData, &result.Buckets,
	)
	if err != nil {
		return nil, fmt.Errorf("GetActions: %w", err)
	}
	return &result, nil
}

// GetActionsRollup returns aggregated action totals for a date range.
func (r *Repository) GetActionsRollup(ctx context.Context, params *ch.QueryParams) (*ActionsRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    toInt64(sum(calls)) AS total_call_clicks,
    toInt64(sum(website)) AS total_website_clicks,
    toInt64(sum(directions)) AS total_direction_requests,
    toInt64(sum(others)) AS total_other_actions,
    ifNotFinite(round(avg(calls + website + directions + others), 2), 0) AS avg_actions
FROM (
    SELECT
        toDate(created_at) AS dates,
        sum(call_clicks) AS calls,
        sum(website_clicks) AS website,
        sum(business_direction_requests) AS directions,
        sum(business_conversations + business_bookings + business_food_orders + business_food_menu_clicks) AS others
    FROM gmb_daily_metrics FINAL
    WHERE %s
    GROUP BY dates
)
`, gmbDateFilter(ids, dateFilter))

	var result ActionsRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TotalCallClicks, &result.TotalWebsiteClicks,
		&result.TotalDirectionRequests, &result.TotalOtherActions, &result.AvgActions,
	)
	if err != nil {
		return nil, fmt.Errorf("GetActionsRollup: %w", err)
	}
	return &result, nil
}

// GetSearchKeywords returns search keywords ordered by impression value.
func (r *Repository) GetSearchKeywords(ctx context.Context, params *ch.QueryParams, limit int) ([]SearchKeywordRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("keyword_month", params)

	query := fmt.Sprintf(`
SELECT
    keyword,
    toInt64(sum(impressions_value)) AS impressions_value,
    toInt64(max(impressions_threshold)) AS impressions_threshold,
    max(keyword_month) AS latest_keyword_month
FROM gmb_search_keywords_monthly FINAL
WHERE %s
GROUP BY keyword
ORDER BY impressions_value DESC
LIMIT %d
`, gmbDateFilter(ids, dateFilter), limit)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetSearchKeywords: %w", err)
	}
	defer rows.Close()

	var results []SearchKeywordRow
	for rows.Next() {
		var row SearchKeywordRow
		if err := rows.Scan(&row.Keyword, &row.ImpressionsValue, &row.ImpressionsThreshold, &row.KeywordMonth); err != nil {
			return nil, fmt.Errorf("GetSearchKeywords scan: %w", err)
		}
		results = append(results, row)
	}
	return results, nil
}

// GetTopPosts returns the latest posts ordered by the given field.
func (r *Repository) GetTopPosts(ctx context.Context, params *ch.QueryParams, limit int, orderBy string) ([]TopPostRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    post_name, summary, state, topic_type, search_url,
    media_names, media_formats, media_google_urls, created_at
FROM gmb_local_posts FINAL
WHERE %s
ORDER BY %s DESC
LIMIT %d
`, gmbDateFilter(ids, dateFilter), orderBy, limit)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetTopPosts: %w", err)
	}
	defer rows.Close()

	var results []TopPostRow
	for rows.Next() {
		var row TopPostRow
		if err := rows.Scan(
			&row.PostName, &row.Summary, &row.State, &row.TopicType, &row.SearchURL,
			&row.MediaNames, &row.MediaFormats, &row.MediaGoogleURLs, &row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetTopPosts scan: %w", err)
		}
		results = append(results, row)
	}
	return results, nil
}

// GetPublishingBehavior returns time-series post counts grouped by date.
func (r *Repository) GetPublishingBehavior(ctx context.Context, params *ch.QueryParams) (*PublishingResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
SELECT
    groupArray(post_count) AS post_count,
    groupArray(dates) AS buckets
FROM (
    SELECT
        toDate(created_at) AS dates,
        toInt64(count()) AS post_count
    FROM gmb_local_posts FINAL
    WHERE %s
    GROUP BY dates
    ORDER BY dates ASC
    %s
)
`, gmbDateFilter(ids, dateFilter), fill)

	var result PublishingResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(&result.PostCount, &result.Buckets)
	if err != nil {
		return nil, fmt.Errorf("GetPublishingBehavior: %w", err)
	}
	return &result, nil
}

// GetTopicTypes returns post counts grouped by topic type.
func (r *Repository) GetTopicTypes(ctx context.Context, params *ch.QueryParams) ([]TopicTypeRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    topic_type,
    toInt64(count()) AS count
FROM gmb_local_posts FINAL
WHERE %s
GROUP BY topic_type
ORDER BY count DESC
`, gmbDateFilter(ids, dateFilter))

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetTopicTypes: %w", err)
	}
	defer rows.Close()

	var results []TopicTypeRow
	for rows.Next() {
		var row TopicTypeRow
		if err := rows.Scan(&row.TopicType, &row.Count); err != nil {
			return nil, fmt.Errorf("GetTopicTypes scan: %w", err)
		}
		results = append(results, row)
	}
	return results, nil
}

// GetReviewsSummary returns aggregated review statistics.
func (r *Repository) GetReviewsSummary(ctx context.Context, params *ch.QueryParams) (*ReviewsSummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    ifNotFinite(round(avg(star_rating), 2), 0) AS avg_rating,
    toInt64(count()) AS total_reviews,
    toInt64(countIf(star_rating = 1)) AS star_1,
    toInt64(countIf(star_rating = 2)) AS star_2,
    toInt64(countIf(star_rating = 3)) AS star_3,
    toInt64(countIf(star_rating = 4)) AS star_4,
    toInt64(countIf(star_rating = 5)) AS star_5
FROM gmb_reviews FINAL
WHERE %s
`, gmbDateFilter(ids, dateFilter))

	var result ReviewsSummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.AvgRating, &result.TotalReviews,
		&result.Star1, &result.Star2, &result.Star3, &result.Star4, &result.Star5,
	)
	if err != nil {
		return nil, fmt.Errorf("GetReviewsSummary: %w", err)
	}
	return &result, nil
}

// GetReviewsTimeSeries returns daily review counts.
func (r *Repository) GetReviewsTimeSeries(ctx context.Context, params *ch.QueryParams) (*ReviewsTimeSeriesResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
SELECT
    groupArray(review_count) AS daily_reviews,
    groupArray(dates) AS buckets
FROM (
    SELECT
        toDate(created_at) AS dates,
        toInt64(count()) AS review_count
    FROM gmb_reviews FINAL
    WHERE %s
    GROUP BY dates
    ORDER BY dates ASC
    %s
)
`, gmbDateFilter(ids, dateFilter), fill)

	var result ReviewsTimeSeriesResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(&result.DailyReviews, &result.Buckets)
	if err != nil {
		return nil, fmt.Errorf("GetReviewsTimeSeries: %w", err)
	}
	return &result, nil
}

// GetReviewsList returns individual review items ordered by creation date.
func (r *Repository) GetReviewsList(ctx context.Context, params *ch.QueryParams, limit int) ([]ReviewRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    review_id, reviewer_display_name, reviewer_profile_photo_url,
    toInt64(star_rating) AS star_rating, comment, reply_comment, created_at
FROM gmb_reviews FINAL
WHERE %s
ORDER BY created_at DESC
LIMIT %d
`, gmbDateFilter(ids, dateFilter), limit)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetReviewsList: %w", err)
	}
	defer rows.Close()

	var results []ReviewRow
	for rows.Next() {
		var row ReviewRow
		if err := rows.Scan(
			&row.ReviewID, &row.ReviewerDisplayName, &row.ReviewerProfilePhotoURL,
			&row.StarRating, &row.Comment, &row.ReplyComment, &row.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("GetReviewsList scan: %w", err)
		}
		results = append(results, row)
	}
	return results, nil
}

// GetReviewsRollup returns aggregated review totals for period comparison.
func (r *Repository) GetReviewsRollup(ctx context.Context, params *ch.QueryParams) (*ReviewsRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    toInt64(count()) AS total_reviews,
    ifNotFinite(round(avg(star_rating), 2), 0) AS avg_rating
FROM gmb_reviews FINAL
WHERE %s
`, gmbDateFilter(ids, dateFilter))

	var result ReviewsRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(&result.TotalReviews, &result.AvgRating)
	if err != nil {
		return nil, fmt.Errorf("GetReviewsRollup: %w", err)
	}
	return &result, nil
}

// GetMediaActivity returns time-series media counts split by photo/video.
func (r *Repository) GetMediaActivity(ctx context.Context, params *ch.QueryParams) (*MediaResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
SELECT
    groupArray(photos) AS photo_count_daily,
    groupArray(videos) AS video_count_daily,
    arraySum(groupArray(photos + videos)) AS show_data,
    groupArray(dates) AS buckets
FROM (
    SELECT
        toDate(created_at) AS dates,
        toInt64(countIf(lower(media_format) != 'video')) AS photos,
        toInt64(countIf(lower(media_format) = 'video')) AS videos
    FROM gmb_media_assets FINAL
    WHERE %s
    GROUP BY dates
    ORDER BY dates ASC
    %s
)
`, gmbDateFilter(ids, dateFilter), fill)

	var result MediaResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PhotoCountDaily, &result.VideoCountDaily, &result.ShowData, &result.Buckets,
	)
	if err != nil {
		return nil, fmt.Errorf("GetMediaActivity: %w", err)
	}
	return &result, nil
}

// GetMediaActivityRollup returns aggregated media totals for period comparison.
func (r *Repository) GetMediaActivityRollup(ctx context.Context, params *ch.QueryParams) (*MediaRollupResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := ch.DateFilter("created_at", params)

	query := fmt.Sprintf(`
SELECT
    toInt64(countIf(lower(media_format) != 'video')) AS total_photos,
    toInt64(countIf(lower(media_format) = 'video')) AS total_videos,
    ifNotFinite(round(toFloat64(count()) / greatest(toFloat64(%d), 1), 2), 0) AS avg_media
FROM gmb_media_assets FINAL
WHERE %s
`, params.DayCount, gmbDateFilter(ids, dateFilter))

	var result MediaRollupResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(&result.TotalPhotos, &result.TotalVideos, &result.AvgMedia)
	if err != nil {
		return nil, fmt.Errorf("GetMediaActivityRollup: %w", err)
	}
	return &result, nil
}
