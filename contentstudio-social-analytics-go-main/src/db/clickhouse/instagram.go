package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// BulkInsertInstagramPosts inserts Instagram posts into ClickHouse using batch insert
func (c *Client) BulkInsertInstagramPosts(ctx context.Context, posts []*clickhousemodels.InstagramPost) error {
	if len(posts) == 0 {
		return nil
	}

	c.Logger.Info().
		Str("table", "instagram_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	b, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO instagram_posts (
			instagram_id, media_id, like_count, comments_count, engagement, impressions, views, reach, saved, video_views,
			shares, reels_avg_watch_time, reels_total_watch_time, exits, replies, taps_forward, taps_back,
			child_assets_type, caption, media_type, entity_type, media_url, video_url,
			username, name, profile_picture_url, permalink, hashtags,
			day_of_week, hour_of_day, year, month, timestamp, stored_event_at, post_created_at, url_refreshed_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertInstagramPosts: failed to prepare batch: %w", err)
	}

	now := time.Now()
	for _, post := range posts {
		err = b.Append(
			post.InstagramID, post.MediaID, post.LikeCount, post.CommentsCount, post.Engagement, post.Impressions, post.Views, post.Reach, post.Saved, post.VideoViews,
			post.Shares, post.ReelsAvgWatchTime, post.ReelsTotalWatchTime, post.Exits, post.Replies, post.TapsForward, post.TapsBack,
			post.ChildAssetsType, post.Caption, post.MediaType, post.EntityType, post.MediaURL, post.VideoURL,
			post.Username, post.Name, post.ProfilePictureURL, post.Permalink, post.Hashtags,
			post.DayOfWeek, post.HourOfDay, post.Year, post.Month, post.Timestamp, post.StoredEventAt, post.PostCreatedAt, now,
		)

		if err != nil {
			return fmt.Errorf("Client.BulkInsertInstagramPosts: failed to append post to batch: %w", err)
		}
	}

	if err := b.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertInstagramPosts: failed to send batch: %w", err)
	}

	c.Logger.Info().
		Str("table", "instagram_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")

	return nil
}

// BulkInsertInstagramInsights inserts Instagram insights into ClickHouse
func (c *Client) BulkInsertInstagramInsights(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error {
	if len(insights) == 0 {
		return nil
	}

	c.Logger.Info().
		Str("table", "instagram_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO instagram_insights (
			instagram_id, record_id, name, username, profile_picture_url,
			follows_count, followers_count, media_count, tags,
			impressions, profile_views, shares, accounts_engaged, engagement, reach, views, saves, likes, comments,
			audience_age, audience_gender, audience_gender_age, audience_locale, audience_city, audience_country,
			audience_age_by_engagement, audience_gender_by_engagement, audience_gender_age_by_engagement, audience_city_by_engagement, audience_country_by_engagement,
			audience_age_by_reach, audience_gender_by_reach, audience_gender_age_by_reach, audience_city_by_reach, audience_country_by_reach,
			online_followers, audience_datetime, online_users_datetime,
			day_of_week, year, month, created_time, updated_time, metadata, stored_event_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertInstagramInsights: failed to prepare batch: %w", err)
	}

	for _, insight := range insights {
		if err := batch.Append(
			insight.InstagramID, insight.RecordID, insight.Name, insight.Username, insight.ProfilePictureURL,
			insight.FollowsCount, insight.FollowersCount, insight.MediaCount, insight.Tags,
			insight.Impressions, insight.ProfileViews, insight.Shares, insight.AccountsEngaged, insight.Engagement, insight.Reach, insight.Views, insight.Saves, insight.Likes, insight.Comments,
			insight.AudienceAge, insight.AudienceGender, insight.AudienceGenderAge, insight.AudienceLocale, insight.AudienceCity, insight.AudienceCountry,
			insight.AudienceAgeByEngagement, insight.AudienceGenderByEngagement, insight.AudienceGenderAgeByEngagement, insight.AudienceCityByEngagement, insight.AudienceCountryByEngagement,
			insight.AudienceAgeByReach, insight.AudienceGenderByReach, insight.AudienceGenderAgeByReach, insight.AudienceCityByReach, insight.AudienceCountryByReach,
			insight.OnlineFollowers, insight.AudienceDatetime, insight.OnlineUsersDatetime,
			insight.DayOfWeek, insight.Year, insight.Month, insight.CreatedTime, insight.UpdatedTime, insight.Metadata, insight.StoredEventAt,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertInstagramInsights: failed to append insight to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertInstagramInsights: failed to send batch: %w", err)
	}

	c.Logger.Info().
		Str("table", "instagram_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")

	return nil
}

func (c *Client) GetMinimalInstagramOlderThan20DaysByAccount(
	ctx context.Context,
	tableName string,
	instagramID string,
	limit, offset int,
) ([]clickhousemodels.InstagramMinimalPost, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "instagram_posts"
	}
	if strings.TrimSpace(instagramID) == "" {
		return nil, fmt.Errorf("Client.GetMinimalInstagramOlderThan20DaysByAccount: instagramID is required")
	}

	q := fmt.Sprintf(`
		SELECT
			instagram_id,
			media_id,
			media_url,
			video_url
		FROM %s
		WHERE instagram_id = ?
		  AND (length(media_url) > 0 OR length(video_url) > 0)
		  AND url_refreshed_at < now() - INTERVAL 10 DAY
		ORDER BY post_created_at DESC, media_id DESC
		LIMIT ? OFFSET ?
	`, tableName)

	rows, err := c.Conn.Query(ctx, q, instagramID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]clickhousemodels.InstagramMinimalPost, 0, 1024)
	for rows.Next() {
		var m clickhousemodels.InstagramMinimalPost
		if err := rows.ScanStruct(&m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (c *Client) GetDistinctInstagramIDsWithStaleURLs(ctx context.Context, tableName string, validIDs []string) ([]string, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "instagram_posts"
	}
	if len(validIDs) == 0 {
		return nil, nil
	}

	q := fmt.Sprintf(`
		SELECT DISTINCT instagram_id
		FROM %s
		WHERE instagram_id IN %s
		  AND (length(media_url) > 0 OR length(video_url) > 0)
		  AND url_refreshed_at < now() - INTERVAL 10 DAY
	`, tableName, FormatAccountIDs(validIDs))

	rows, err := c.Conn.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("Client.GetDistinctInstagramIDsWithStaleURLs: query failed: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("Client.GetDistinctInstagramIDsWithStaleURLs: scan failed: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (c *Client) MarkInstagramPostsRefreshed(ctx context.Context, tableName, instagramID string) error {
	if strings.TrimSpace(tableName) == "" {
		tableName = "instagram_posts"
	}
	if strings.TrimSpace(instagramID) == "" {
		return nil
	}
	now := time.Now()
	ts := now.Unix()
	threshold := now.AddDate(0, 0, -10).Unix()
	q := fmt.Sprintf(`
		ALTER TABLE %s
		UPDATE url_refreshed_at = toDateTime(%d)
		WHERE instagram_id = ?
		  AND url_refreshed_at < toDateTime(%d)
	`, tableName, ts, threshold)
	if err := c.Conn.Exec(ctx, q, instagramID); err != nil {
		return fmt.Errorf("Client.MarkInstagramPostsRefreshed: %w", err)
	}
	return nil
}

func (c *Client) BulkMarkInstagramPostsRefreshed(ctx context.Context, tableName string, instagramIDs []string) error {
	if strings.TrimSpace(tableName) == "" {
		tableName = "instagram_posts"
	}
	if len(instagramIDs) == 0 {
		return nil
	}
	ts := time.Now().Unix()
	q := fmt.Sprintf(`
		ALTER TABLE %s
		UPDATE url_refreshed_at = toDateTime(%d)
		WHERE has(CAST(? AS Array(String)), instagram_id)
		  AND url_refreshed_at < now() - INTERVAL 10 DAY
	`, tableName, ts)
	if err := c.Conn.Exec(ctx, q, instagramIDs); err != nil {
		return fmt.Errorf("Client.BulkMarkInstagramPostsRefreshed: %w", err)
	}
	return nil
}

func (c *Client) BulkUpdateInstagramMediaURLs(
	ctx context.Context,
	tableName string,
	rows []clickhousemodels.InstagramMinimalPost,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "instagram_posts"
	}

	type entry struct {
		id       string
		mediaURL []string
		videoURL []string
	}

	ts := time.Now().Unix()
	seen := make(map[string]struct{}, len(rows))
	entries := make([]entry, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.MediaID) == "" {
			continue
		}
		if _, ok := seen[row.MediaID]; ok {
			continue
		}
		seen[row.MediaID] = struct{}{}
		entries = append(entries, entry{row.MediaID, row.MediaURL, row.VideoURL})
	}

	if len(entries) == 0 {
		return 0, nil
	}

	const chunkSize = 5000
	total := 0

	for i := 0; i < len(entries); i += chunkSize {
		end := i + chunkSize
		if end > len(entries) {
			end = len(entries)
		}
		chunk := entries[i:end]

		ids := make([]string, 0, len(chunk))
		mediaURLs := make([][]string, 0, len(chunk))
		videoURLs := make([][]string, 0, len(chunk))
		for _, e := range chunk {
			ids = append(ids, e.id)
			mediaURLs = append(mediaURLs, e.mediaURL)
			videoURLs = append(videoURLs, e.videoURL)
		}

		q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  media_url = if(
    indexOf(CAST(? AS Array(String)), media_id) > 0,
    arrayElement(CAST(? AS Array(Array(String))), indexOf(CAST(? AS Array(String)), media_id)),
    media_url
  ),
  video_url = if(
    indexOf(CAST(? AS Array(String)), media_id) > 0,
    arrayElement(CAST(? AS Array(Array(String))), indexOf(CAST(? AS Array(String)), media_id)),
    video_url
  ),
  url_refreshed_at = toDateTime(%d)
WHERE has(CAST(? AS Array(String)), media_id)
`, tableName, ts)

		if err := c.Conn.Exec(ctx, q,
			ids, mediaURLs, ids,
			ids, videoURLs, ids,
			ids,
		); err != nil {
			return total, fmt.Errorf("Client.BulkUpdateInstagramMediaURLs: update failed: %w", err)
		}
		total += len(ids)
	}

	c.Logger.Info().
		Str("table", tableName).
		Int("media_ids", total).
		Msg("Bulk-updated Instagram media URLs")

	return total, nil
}

func (c *Client) UpdateInstagramMediaURLs(
	ctx context.Context,
	tableName string,
	instagramID string,
	rows []clickhousemodels.InstagramMinimalPost,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "instagram_posts"
	}
	if strings.TrimSpace(instagramID) == "" {
		return 0, fmt.Errorf("Client.UpdateInstagramMediaURLs: instagramID is required")
	}

	ids := make([]string, 0, len(rows))
	mediaURLs := make([][]string, 0, len(rows))
	videoURLs := make([][]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))

	for _, row := range rows {
		if row.InstagramID != instagramID {
			continue
		}
		if strings.TrimSpace(row.MediaID) == "" {
			continue
		}
		if _, ok := seen[row.MediaID]; ok {
			continue
		}
		seen[row.MediaID] = struct{}{}
		ids = append(ids, row.MediaID)
		mediaURLs = append(mediaURLs, row.MediaURL)
		videoURLs = append(videoURLs, row.VideoURL)
	}

	if len(ids) == 0 {
		return 0, nil
	}

	q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  media_url = if(
    indexOf(CAST(? AS Array(String)), media_id) > 0,
    arrayElement(CAST(? AS Array(Array(String))), indexOf(CAST(? AS Array(String)), media_id)),
    media_url
  ),
  video_url = if(
    indexOf(CAST(? AS Array(String)), media_id) > 0,
    arrayElement(CAST(? AS Array(Array(String))), indexOf(CAST(? AS Array(String)), media_id)),
    video_url
  )
WHERE instagram_id = ?
  AND has(CAST(? AS Array(String)), media_id)
`, tableName)

	if err := c.Conn.Exec(
		ctx,
		q,
		ids, mediaURLs, ids,
		ids, videoURLs, ids,
		instagramID,
		ids,
	); err != nil {
		return 0, fmt.Errorf("Client.UpdateInstagramMediaURLs: update failed: %w", err)
	}

	return len(ids), nil
}
