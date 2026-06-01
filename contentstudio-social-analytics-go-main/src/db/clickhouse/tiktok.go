package clickhouse

import (
	"context"
	"fmt"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// GetTikTokPostsViewSum queries the sum of all view_count for a given tiktok_id
// Uses FINAL modifier on ReplacingMergeTree to automatically get the latest version of each post_id
func (c *Client) GetTikTokPostsViewSum(ctx context.Context, tiktokID string) (int64, error) {
	var totalViews int64

	err := c.Conn.QueryRow(ctx, `
		SELECT SUM(view_count) as total_views
		FROM tiktok_posts FINAL
		WHERE tiktok_id = ?
	`, tiktokID).Scan(&totalViews)

	if err != nil {
		c.Logger.Warn().Err(err).
			Str("tiktok_id", tiktokID).
			Msg("Failed to query total views from posts")
		return 0, fmt.Errorf("Client.GetTikTokPostsViewSum: query total views: %w", err)
	}

	return totalViews, nil
}

// BulkInsertTikTokPosts inserts TikTok posts into ClickHouse in batches.
func (c *Client) BulkInsertTikTokPosts(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error {
	if len(posts) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "tiktok_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
        INSERT INTO tiktok_posts (
            tiktok_id, display_name, profile_link, post_id,
            cover_image_url, share_url, post_description, hashtags, duration,
            height, width, title, embed_html, embed_link,
            like_count, comments_count, share_count, view_count,
            engagement_count, engagement_rate, created_at, inserted_at
        )
    `)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertTikTokPosts: prepare batch: %w", err)
	}
	for _, p := range posts {
		if err := batch.Append(
			p.TikTokID, p.DisplayName, p.ProfileLink, p.PostID,
			p.CoverImageURL, p.ShareURL, p.PostDescription, p.Hashtags, p.Duration,
			p.Height, p.Width, p.Title, p.EmbedHTML, p.EmbedLink,
			p.LikeCount, p.CommentCount, p.ShareCount, p.ViewCount,
			p.EngagementCount, p.EngagementRate, p.CreatedAt, p.InsertedAt,
		); err != nil {
			return fmt.Errorf("append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertTikTokPosts: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "tiktok_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertTikTokInsights inserts account-level insights.
func (c *Client) BulkInsertTikTokInsights(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error {
	if len(insights) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "tiktok_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
        INSERT INTO tiktok_insights (
            record_id, tiktok_id, display_name, profile_image,
            total_follower_count, total_following_count, total_like_count,
            total_video_count, total_video_views, total_video_likes,
            total_video_comments, total_video_shares, is_verified, bio,
            profile_link, inserted_at
        )
    `)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertTikTokInsights: prepare batch insights: %w", err)
	}
	for _, in := range insights {
		if err := batch.Append(
			in.RecordID, in.TikTokID, in.DisplayName, in.ProfileImage,
			in.TotalFollowerCount, in.TotalFollowingCount, in.TotalLikeCount,
			in.TotalVideoCount, in.TotalVideoViews, in.TotalVideoLikes,
			in.TotalVideoComments, in.TotalVideoShares, in.IsVerified, in.Bio,
			in.ProfileLink, in.InsertedAt,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertTikTokInsights: append insights: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertTikTokInsights: send insights: %w", err)
	}
	c.Logger.Info().
		Str("table", "tiktok_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")
	return nil
}
