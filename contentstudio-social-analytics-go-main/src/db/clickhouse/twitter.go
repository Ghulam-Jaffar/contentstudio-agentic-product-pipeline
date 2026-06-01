package clickhouse

import (
	"context"
	"fmt"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// BulkInsertTwitterPosts inserts Twitter posts into ClickHouse in batches.
func (c *Client) BulkInsertTwitterPosts(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error {
	if len(posts) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "twitter_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
        INSERT INTO twitter_posts (
            twitter_id, name, username, profile_image_url,
            followers_count, following_count, tweet_count, listed_count,
            tweet_id, edit_history_tweet_ids, author_id, author_username,
            id_created_at, author_id_created, tweeted_at, saving_time,
            hashtags, permalink, tweet_type, urls, media_url,
            username_mentioned, userid_mentioned, lang, tweet_text,
            impression_count, retweet_count, reply_count, like_count,
            bookmark_count, quote_count, total_engagement,
            day_of_week, hour_of_day
        )
    `)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertTwitterPosts: prepare batch: %w", err)
	}
	for _, p := range posts {
		if err := batch.Append(
			p.TwitterID, p.Name, p.Username, p.ProfileImageURL,
			p.FollowersCount, p.FollowingCount, p.TweetCount, p.ListedCount,
			p.TweetID, p.EditHistoryTweetIDs, p.AuthorID, p.AuthorUsername,
			p.IDCreatedAt, p.AuthorIDCreated, p.TweetedAt, p.SavingTime,
			p.Hashtags, p.Permalink, p.TweetType, p.URLs, p.MediaURL,
			p.UsernameMentioned, p.UseridMentioned, p.Lang, p.TweetText,
			p.ImpressionCount, p.RetweetCount, p.ReplyCount, p.LikeCount,
			p.BookmarkCount, p.QuoteCount, p.TotalEngagement,
			p.DayOfWeek, p.HourOfDay,
		); err != nil {
			return fmt.Errorf("append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertTwitterPosts: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "twitter_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertTwitterInsights inserts account-level insights into ClickHouse.
func (c *Client) BulkInsertTwitterInsights(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error {
	if len(insights) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "twitter_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
        INSERT INTO twitter_insights (
            twitter_id, record_id, name, username, profile_image_url,
            description, verified, account_created_date,
            followers_count, following_count, tweet_count, listed_count, like_count,
            day_of_week, saving_time
        )
    `)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertTwitterInsights: prepare batch insights: %w", err)
	}
	for _, in := range insights {
		if err := batch.Append(
			in.TwitterID, in.RecordID, in.Name, in.Username, in.ProfileImageURL,
			in.Description, in.Verified, in.AccountCreatedDate,
			in.FollowersCount, in.FollowingCount, in.TweetCount, in.ListedCount, in.LikeCount,
			in.DayOfWeek, in.SavingTime,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertTwitterInsights: append insights: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertTwitterInsights: send insights: %w", err)
	}
	c.Logger.Info().
		Str("table", "twitter_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")
	return nil
}
