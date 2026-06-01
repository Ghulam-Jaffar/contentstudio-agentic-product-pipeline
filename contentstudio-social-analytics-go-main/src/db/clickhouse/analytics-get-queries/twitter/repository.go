package twitter

import (
	"context"
	"fmt"
	"strings"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

type Repository struct {
	client *ch.Client
}

func NewRepository(client *ch.Client) *Repository {
	return &Repository{client: client}
}

func twDateFilter(field string, params *ch.QueryParams) string {
	return fmt.Sprintf(
		"toDate(%s, '%s') BETWEEN '%s' AND '%s'",
		field,
		params.Timezone,
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)
}

func twDailyFill(params *ch.QueryParams) string {
	return fmt.Sprintf(
		"WITH FILL FROM toDate('%s') TO toDate('%s') + 1 STEP 1",
		params.DateFrom.Format("2006-01-02"),
		params.DateTo.Format("2006-01-02"),
	)
}

func (r *Repository) GetSummary(ctx context.Context, params *ch.QueryParams) (*SummaryResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	insightsFilter := twDateFilter("saving_time", params)
	postsFilter := twDateFilter("tweeted_at", params)

	query := fmt.Sprintf(`
SELECT
    coalesce(insights_data.twitter_id, '') AS twitter_id,
    coalesce(insights_data.name, '') AS name,
    coalesce(insights_data.profile_image_url, '') AS profile_image_url,
    coalesce(insights_data.followers_count, 0) AS followers_count,
    coalesce(insights_data.following_count, 0) AS following_count,
    coalesce(insights_data.tweet_count, 0) AS tweet_count,
    coalesce(insights_data.listed_count, 0) AS listed_count,
    coalesce(posts_data.impression_count, 0) AS impression_count,
    coalesce(posts_data.total_engagement, 0) AS total_engagement,
    coalesce(posts_data.reply_count, 0) AS reply_count,
    coalesce(posts_data.retweet_count, 0) AS retweet_count,
    coalesce(posts_data.bookmark_count, 0) AS bookmark_count,
    coalesce(posts_data.like_count, 0) AS like_count,
    coalesce(posts_data.quote_count, 0) AS quote_count,
    coalesce(posts_data.tweet_count, 0) AS posts_tweet_count
FROM (
    SELECT
        first_value(twitter_id) AS twitter_id,
        first_value(name) AS name,
        first_value(profile_image_url) AS profile_image_url,
        first_value(followers_count) AS followers_count,
        first_value(following_count) AS following_count,
        first_value(tweet_count) AS tweet_count,
        first_value(listed_count) AS listed_count,
        first_value(like_count) AS like_count
    FROM (
        SELECT
            twitter_id,
            max(name) AS name,
            max(profile_image_url) AS profile_image_url,
            toInt64(argMin(followers_count, saving_time)) AS followers_count,
            toInt64(argMin(following_count, saving_time)) AS following_count,
            toInt64(max(tweet_count)) AS tweet_count,
            toInt64(max(listed_count)) AS listed_count,
            toInt64(max(like_count)) AS like_count,
            max(saving_time) AS record_date
        FROM twitter_insights
        WHERE twitter_id IN %s AND %s
        GROUP BY twitter_id, record_id
        ORDER BY record_date DESC
    )
) AS insights_data
LEFT JOIN (
    SELECT
        twitter_id,
        max(name) AS name,
        max(profile_image_url) AS profile_image_url,
        toInt64(sum(impression_count)) AS impression_count,
        toInt64(sum(total_engagement)) AS total_engagement,
        toInt64(sum(reply_count)) AS reply_count,
        toInt64(sum(retweet_count)) AS retweet_count,
        toInt64(sum(bookmark_count)) AS bookmark_count,
        toInt64(sum(like_count)) AS like_count,
        toInt64(sum(quote_count)) AS quote_count,
        toInt64(count()) AS tweet_count
    FROM (
        SELECT
            twitter_id,
            tweet_id,
            max(name) AS name,
            max(profile_image_url) AS profile_image_url,
            max(impression_count) AS impression_count,
            max(total_engagement) AS total_engagement,
            max(reply_count) AS reply_count,
            max(retweet_count) AS retweet_count,
            max(bookmark_count) AS bookmark_count,
            max(like_count) AS like_count,
            max(quote_count) AS quote_count,
            max(tweeted_at) AS tweeted_at_time
        FROM twitter_posts
        WHERE twitter_id IN %s AND %s
        GROUP BY twitter_id, tweet_id
        ORDER BY tweeted_at_time DESC
    )
    GROUP BY twitter_id
) AS posts_data ON insights_data.twitter_id = posts_data.twitter_id
`, ids, insightsFilter, ids, postsFilter)

	var result SummaryResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TwitterID,
		&result.Name,
		&result.ProfileImageURL,
		&result.FollowersCount,
		&result.FollowingCount,
		&result.TweetCount,
		&result.ListedCount,
		&result.ImpressionCount,
		&result.TotalEngagement,
		&result.ReplyCount,
		&result.RetweetCount,
		&result.BookmarkCount,
		&result.LikeCount,
		&result.QuoteCount,
		&result.PostsTweetCount,
	)
	if err != nil {
		return nil, fmt.Errorf("GetSummary: %w", err)
	}

	return &result, nil
}

func (r *Repository) GetEngagementImpressionData(ctx context.Context, params *ch.QueryParams) (*EngagementImpressionResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := twDateFilter("tweeted_at", params)
	fill := ch.WithFill(ch.FormatDate(params.DateTo))

	query := fmt.Sprintf(`
SELECT
    any(twitter_id) AS twitter_id,
    groupArray(tweet_count) AS tweet_count,
    groupArray(impression_count) AS impression_count,
    groupArray(total_engagement) AS total_engagement,
    groupArray(formatDateTime(tweeted_at_date, '%%Y-%%m-%%d')) AS tweeted_at_date,
    groupArray(retweet_count) AS retweet_count,
    groupArray(reply_count) AS reply_count,
    groupArray(like_count) AS like_count,
    groupArray(bookmark_count) AS bookmark_count,
    groupArray(quote_count) AS quote_count
FROM (
    SELECT
        twitter_id,
        toInt64(count()) AS tweet_count,
        toInt64(sum(impression_count)) AS impression_count,
        toInt64(sum(total_engagement)) AS total_engagement,
        tweeted_at_date,
        toInt64(sum(retweet_count)) AS retweet_count,
        toInt64(sum(reply_count)) AS reply_count,
        toInt64(sum(like_count)) AS like_count,
        toInt64(sum(bookmark_count)) AS bookmark_count,
        toInt64(sum(quote_count)) AS quote_count
    FROM (
        SELECT
            twitter_id,
            tweet_id,
            max(impression_count) AS impression_count,
            max(total_engagement) AS total_engagement,
            max(tweeted_at) AS tweeted_at_time,
            max(retweet_count) AS retweet_count,
            max(reply_count) AS reply_count,
            max(like_count) AS like_count,
            max(bookmark_count) AS bookmark_count,
            max(quote_count) AS quote_count,
            toDate(max(tweeted_at)) AS tweeted_at_date
        FROM twitter_posts
        WHERE twitter_id IN %s AND %s
        GROUP BY twitter_id, tweet_id
        ORDER BY tweeted_at_time ASC
    )
    GROUP BY twitter_id, tweeted_at_date
    ORDER BY tweeted_at_date ASC
    %s
)
`, ids, dateFilter, fill)

	var result EngagementImpressionResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.TwitterID,
		&result.TweetCount,
		&result.ImpressionCount,
		&result.TotalEngagement,
		&result.TweetedAtDate,
		&result.RetweetCount,
		&result.ReplyCount,
		&result.LikeCount,
		&result.BookmarkCount,
		&result.QuoteCount,
	)
	if err != nil {
		return nil, fmt.Errorf("GetEngagementImpressionData: %w", err)
	}

	return &result, nil
}

func (r *Repository) GetFollowersTrend(ctx context.Context, params *ch.QueryParams) (*FollowersTrendResult, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := twDateFilter("saving_time", params)
	fill := twDailyFill(params)

	query := fmt.Sprintf(`
SELECT
    platform_id,
    any(name) AS name,
    any(username) AS username,
    groupArray(follower_count) AS follower_count,
    groupArray(follower_count_daily) AS follower_count_daily,
    groupArray(following_count) AS following_count,
    groupArray(following_count_daily) AS following_count_daily,
    groupArray(saving_date) AS buckets
FROM (
    SELECT
        platform_id,
        name,
        username,
        follower_count,
        greatest(
            follower_count - lagInFrame(follower_count, 1, follower_count) OVER (
                PARTITION BY platform_id
                ORDER BY bucket_date ASC
                ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
            ),
            0
        ) AS follower_count_daily,
        following_count,
        greatest(
            following_count - lagInFrame(following_count, 1, following_count) OVER (
                PARTITION BY platform_id
                ORDER BY bucket_date ASC
                ROWS BETWEEN UNBOUNDED PRECEDING AND CURRENT ROW
            ),
            0
        ) AS following_count_daily,
        bucket_date,
        formatDateTime(bucket_date, '%%Y-%%m-%%d') AS saving_date
    FROM (
        SELECT
            record_id,
            max(twitter_id) AS platform_id,
            max(name) AS name,
            max(username) AS username,
            toInt64(argMax(followers_count, saving_time)) AS follower_count,
            toInt64(argMax(following_count, saving_time)) AS following_count,
            toDate(max(saving_time)) AS bucket_date
        FROM twitter_insights
        WHERE twitter_id IN %s AND %s
        GROUP BY record_id
        ORDER BY bucket_date ASC
        %s
    )
)
GROUP BY platform_id
`, ids, dateFilter, fill)

	var result FollowersTrendResult
	err := r.client.Conn.QueryRow(ctx, query).Scan(
		&result.PlatformID,
		&result.Name,
		&result.Username,
		&result.FollowerCount,
		&result.FollowerCountDaily,
		&result.FollowingCount,
		&result.FollowingCountDaily,
		&result.Buckets,
	)
	if err != nil {
		return nil, fmt.Errorf("GetFollowersTrend: %w", err)
	}

	return &result, nil
}

func (r *Repository) GetTweetsData(ctx context.Context, params *ch.QueryParams, orderBy string, limit int, sort string) ([]TweetRow, error) {
	ids := ch.FormatAccountIDs(params.AccountIDs)
	dateFilter := twDateFilter("twitter_posts.tweeted_at", params)
	sort = strings.ToUpper(sort)

	query := fmt.Sprintf(`
SELECT
    tweet_id AS id,
    formatDateTime(max(tweeted_at), '%%Y-%%m-%%d %%H:%%i:%%S') AS tweeted_at_value,
    argMax(tweet_text, saving_time) AS tweet_text,
    argMax(tweet_type, saving_time) AS tweet_type,
    argMax(permalink, saving_time) AS permalink,
    argMax(media_url, saving_time) AS media_url,
    toInt32(argMax(listed_count, saving_time)) AS listed_count,
    toInt32(argMax(retweet_count, saving_time)) AS retweet_count,
    toInt32(argMax(like_count, saving_time)) AS like_count,
    toInt32(argMax(reply_count, saving_time)) AS reply_count,
    toInt32(argMax(quote_count, saving_time)) AS quote_count,
    toInt32(argMax(bookmark_count, saving_time)) AS bookmark_count,
    toInt32(argMax(impression_count, saving_time)) AS impression_count,
    toInt32(argMax(total_engagement, saving_time)) AS total_engagement
FROM twitter_posts
WHERE twitter_id IN %s AND %s
GROUP BY tweet_id
ORDER BY %s %s
LIMIT %d
`, ids, dateFilter, orderBy, sort, limit)

	rows, err := r.client.Conn.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("GetTweetsData query: %w", err)
	}
	defer rows.Close()

	result := make([]TweetRow, 0)
	for rows.Next() {
		var row TweetRow
		if err := rows.Scan(
			&row.ID,
			&row.TweetedAt,
			&row.TweetText,
			&row.TweetType,
			&row.Permalink,
			&row.MediaURL,
			&row.ListedCount,
			&row.RetweetCount,
			&row.LikeCount,
			&row.ReplyCount,
			&row.QuoteCount,
			&row.BookmarkCount,
			&row.ImpressionCount,
			&row.TotalEngagement,
		); err != nil {
			return nil, fmt.Errorf("GetTweetsData scan: %w", err)
		}
		result = append(result, row)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("GetTweetsData rows: %w", err)
	}

	return result, nil
}
