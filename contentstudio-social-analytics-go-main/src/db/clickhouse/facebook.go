package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// BulkInsertPosts inserts Facebook posts into ClickHouse using batch insert
func (c *Client) BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
	if len(posts) == 0 {
		return nil
	}

	c.Logger.Info().
		Str("table", "facebook_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO facebook_posts (
			page_name, page_id, media_type, post_id, permalink, status_type, video_id,
			category, published_by, published_by_url, shared_from_name, shared_from_id, shared_from_link,
			like, love, haha, wow, sad, angry, thankful, total,
			shares, comments, post_clicks, total_engagement, post_engaged_users,
			day_of_week, hour_of_day, created_time, updated_time, saving_time,
			message_tags, post_metadata, caption, description, full_picture, link,
			post_impressions, post_impressions_unique, post_impressions_paid, post_impressions_paid_unique,
			post_impressions_organic, post_impressions_organic_unique, post_impressions_viral, post_impressions_viral_unique,
			post_video_views, total_impressions, url_refreshed_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertPosts: failed to prepare batch: %w", err)
	}

	now := time.Now()
	for _, post := range posts {
		err = batch.Append(
			post.PageName, post.PageID, post.MediaType, post.PostID, post.Permalink, post.StatusType, post.VideoID,
			post.Category, post.PublishedBy, post.PublishedByURL, post.SharedFromName, post.SharedFromID, post.SharedFromLink,
			post.Like, post.Love, post.Haha, post.Wow, post.Sad, post.Angry, post.Thankful, post.Total,
			post.Shares, post.Comments, post.PostClicks, post.TotalEngagement, post.PostEngagedUsers,
			post.DayOfWeek, post.HourOfDay, post.CreatedTime, post.UpdatedTime, post.SavingTime,
			post.MessageTags, post.PostMetadata, post.Caption, post.Description, post.FullPicture, post.Link,
			post.PostImpressions, post.PostImpressionsUnique, post.PostImpressionsPaid, post.PostImpressionsPaidUnique,
			post.PostImpressionsOrganic, post.PostImpressionsOrganicUnique, post.PostImpressionsViral, post.PostImpressionsViralUnique,
			post.PostVideoViews, post.TotalImpressions, now,
		)
		if err != nil {
			return fmt.Errorf("Client.BulkInsertPosts: failed to append post to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertPosts: failed to send batch: %w", err)
	}

	c.Logger.Info().
		Str("table", "facebook_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")

	return nil
}

// BulkInsertMediaAssets inserts Facebook media assets into ClickHouse
func (c *Client) BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
	if len(assets) == 0 {
		return nil
	}

	c.Logger.Info().
		Str("table", "facebook_media_assets").
		Int("batch_size", len(assets)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO facebook_media_assets (
			page_id, media_id, post_id, asset_type, link, call_to_action, CTA_type, caption,
			description, created_at, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertMediaAssets: failed to prepare batch: %w", err)
	}

	for _, asset := range assets {
		if err := batch.Append(
			asset.PageID, asset.MediaID, asset.PostID, asset.AssetType,
			asset.Link, asset.CallToAction, asset.CTAType, asset.Caption,
			asset.Description, asset.CreatedAt, asset.InsertedAt,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertMediaAssets: failed to append media asset to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertMediaAssets: failed to send batch: %w", err)
	}

	c.Logger.Info().
		Str("table", "facebook_media_assets").
		Int("batch_size", len(assets)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")

	return nil
}

// BulkInsertVideoInsights inserts Facebook video insights into ClickHouse
func (c *Client) BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
	if len(insights) == 0 {
		return nil
	}

	c.Logger.Info().
		Str("table", "facebook_video_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO facebook_video_insights (
			page_id, post_id, video_id, created_time, updated_time,
			total_video_views, total_video_impressions, total_video_complete_views,
			total_video_10s_views, total_video_15s_views, total_video_30s_views,
			total_video_60s_excludes_shorter_views, total_video_avg_time_watched,
			total_video_impressions_unique, total_video_view_total_time,
			total_video_views_unique, total_video_stories_by_action_type, total_video_view_total_time_organic
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertVideoInsights: failed to prepare batch: %w", err)
	}

	for _, insight := range insights {
		if err := batch.Append(
			insight.PageID, insight.PostID, insight.VideoID,
			insight.CreatedTime, insight.UpdatedTime,
			insight.TotalVideoViews, insight.TotalVideoImpressions, insight.TotalVideoCompleteViews,
			insight.TotalVideo10sViews, insight.TotalVideo15sViews, insight.TotalVideo30sViews,
			insight.TotalVideo60sExcludesShorterViews, insight.TotalVideoAvgTimeWatched,
			insight.TotalVideoImpressionsUnique, insight.TotalVideoViewTotalTime,
			insight.TotalVideoViewsUnique, insight.TotalVideoStoriesByActionType, insight.TotalVideoViewTotalTimeOrganic,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertVideoInsights: failed to append video insight to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertVideoInsights: failed to send batch: %w", err)
	}

	c.Logger.Info().
		Str("table", "facebook_video_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")

	return nil
}

// BulkInsertReelsInsights inserts Facebook reels insights into ClickHouse
func (c *Client) BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
	if len(insights) == 0 {
		return nil
	}

	c.Logger.Info().
		Str("table", "facebook_reels_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO facebook_reels_insights (
			page_id, post_id, average_time_watched, total_time_watched_in_ms,
			play_count, impressions_unique, reel_followers, created_at, saving_time
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertReelsInsights: failed to prepare batch: %w", err)
	}

	for _, insight := range insights {
		if err := batch.Append(
			insight.PageID, insight.PostID, insight.AverageTimeWatched, insight.TotalTimeWatchedInMs,
			insight.PlayCount, insight.ImpressionsUnique, insight.ReelFollowers, insight.CreatedAt, insight.SavingTime,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertReelsInsights: failed to append reels insight to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertReelsInsights: failed to send batch: %w", err)
	}

	c.Logger.Info().
		Str("table", "facebook_reels_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")

	return nil
}

// BulkInsertInsights inserts Facebook page insights into ClickHouse
func (c *Client) BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
	if len(insights) == 0 {
		return nil
	}

	c.Logger.Info().
		Str("table", "facebook_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO facebook_insights (
			hash_id, page_id, page_category, day_of_week, year, month, created_time, saving_time,
			page_fans, page_fans_city, page_fans_country, page_fans_locale, page_fans_age, page_fans_gender, page_fans_gender_age,
			page_follows, page_views, page_fan_adds_by_paid_non_paid_unique,
			page_fan_adds_unique, page_fan_removes_unique, page_fans_by_like_source_unique, page_fans_by_unlike_source_unique,
			page_fans_by_like, page_fans_by_unlike, page_total_actions, page_post_engagements, page_impressions,
			page_impressions_organic, page_impressions_paid, page_video_views_paid, page_video_views, page_video_views_organic,
			page_video_views_autoplayed, page_video_views_click_to_play, page_video_repeat_views,
			page_negative_feedback, page_positive_feedback, page_negative_feedback_by_type, page_positive_feedback_by_type,
			page_fans_online, active_users, positive_sentiment, negative_sentiment,
			posts_count, likes_count, talking_about_count, type_count, message_count, prime_time, page_impressions_unique
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertInsights: failed to prepare batch: %w", err)
	}

	for _, insight := range insights {
		if err := batch.Append(
			insight.HashID, insight.PageID, insight.PageCategory, insight.DayOfWeek, insight.Year, insight.Month, insight.CreatedTime, insight.SavingTime,
			insight.PageFans, insight.PageFansCity, insight.PageFansCountry, insight.PageFansLocale, insight.PageFansAge, insight.PageFansGender, insight.PageFansGenderAge,
			insight.PageFollows, insight.PageViews, insight.PageFanAddsByPaidNonPaidUnique,
			insight.PageFanAddsUnique, insight.PageFanRemovesUnique, insight.PageFansByLikeSourceUnique, insight.PageFansByUnlikeSourceUnique,
			insight.PageFansByLike, insight.PageFansByUnlike, insight.PageTotalActions, insight.PagePostEngagements, insight.PageImpressions,
			insight.PageImpressionsOrganic, insight.PageImpressionsPaid, insight.PageVideoViewsPaid, insight.PageVideoViews, insight.PageVideoViewsOrganic,
			insight.PageVideoViewsAutoplayed, insight.PageVideoViewsClickToPlay, insight.PageVideoRepeatViews,
			insight.PageNegativeFeedback, insight.PagePositiveFeedback, insight.PageNegativeFeedbackByType, insight.PagePositiveFeedbackByType,
			insight.PageFansOnline, insight.ActiveUsers, insight.PositiveSentiment, insight.NegativeSentiment,
			insight.PostsCount, insight.LikesCount, insight.TalkingAboutCount, insight.TypeCount, insight.MessageCount, insight.PrimeTime, insight.PageImpressionsUnique,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertInsights: failed to append insight to batch: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertInsights: failed to send batch: %w", err)
	}

	c.Logger.Info().
		Str("table", "facebook_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")

	return nil
}

// Requires: go get github.com/ClickHouse/clickhouse-go/v2

// GetMinimalOlderThan20DaysByPage returns (page_id, post_id, full_picture)
// for posts older than 10 days that have a non-empty full_picture.
func (c *Client) GetMinimalOlderThan20DaysByPage(
	ctx context.Context,
	tableName string, // defaults to "facebook_posts"
	pageID string, // required
	limit, offset int,
) ([]clickhousemodels.MinimalPost, error) {

	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_posts"
	}
	if strings.TrimSpace(pageID) == "" {
		return nil, fmt.Errorf("Client.GetMinimalOlderThan20DaysByPage: pageID is required")
	}

	q := fmt.Sprintf(`
		SELECT
			page_id,
			post_id,
			full_picture
		FROM %s
		WHERE page_id = ?
		  AND full_picture != ''
		  AND url_refreshed_at < now() - INTERVAL 10 DAY
		ORDER BY created_time DESC, post_id DESC
		LIMIT ? OFFSET ?
	`, tableName)

	rows, err := c.Conn.Query(ctx, q, pageID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]clickhousemodels.MinimalPost, 0, 1024)
	for rows.Next() {
		var m clickhousemodels.MinimalPost
		if err := rows.ScanStruct(&m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

// Updates full_picture for a single page_id using transform(from[], to[], default)
// No temp table. No IN (...) needed.
func (c *Client) UpdateFullPictures(
	ctx context.Context,
	tableName string, // e.g., "facebook_posts"
	pageID string, // the page to update
	rows []clickhousemodels.MinimalPost,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_posts"
	}
	if strings.TrimSpace(pageID) == "" {
		return 0, fmt.Errorf("Client.UpdateFullPictures: pageID is required")
	}

	// Build aligned arrays: postIDs[i] ↔ urls[i]
	postIDs := make([]string, 0, len(rows))
	urls := make([]string, 0, len(rows))

	seen := make(map[string]struct{}, len(rows))
	for _, r := range rows {
		if r.PageID != pageID {
			continue
		}
		if r.PostID == "" || strings.TrimSpace(r.FullPicture) == "" {
			continue
		}
		if _, ok := seen[r.PostID]; ok {
			continue
		}
		seen[r.PostID] = struct{}{}
		postIDs = append(postIDs, r.PostID)
		urls = append(urls, r.FullPicture)
	}
	if len(postIDs) == 0 {
		return 0, nil
	}
	if len(postIDs) != len(urls) {
		return 0, fmt.Errorf("Client.UpdateFullPictures: postIDs and urls length mismatch")
	}

	// Lightweight UPDATE using transform:
	// - transform(post_id, fromArr, toArr, default) returns the mapped URL
	// - If post_id not in fromArr, it returns default (the current full_picture), so no-op
	q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  full_picture = transform(
    post_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    full_picture
  ),
  metadata = mapUpdate(
    ifNull(metadata, CAST([], 'Map(String, String)')),
    CAST(map('message', 'full picture link is updated'), 'Map(String, String)')
  ),
  updated_time = toDateTime64(?, 6, 'UTC')
WHERE page_id = ?
  AND has(CAST(? AS Array(String)), post_id)
`, tableName)

	updatedTime := time.Now().UTC().Format("2006-01-02 15:04:05.000000")

	if err := c.Conn.Exec(ctx, q, postIDs, urls, updatedTime, pageID, postIDs); err != nil {
		return 0, fmt.Errorf("Client.UpdateFullPictures: lightweight update(transform) failed: %w", err)
	}

	c.Logger.Info().Msg("Successfully updated Facebook post in clickhouse")
	return len(postIDs), nil
}

func (c *Client) GetDistinctFacebookPageIDsWithStaleURLs(ctx context.Context, tableName string, validPageIDs []string) ([]string, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_posts"
	}
	if len(validPageIDs) == 0 {
		return nil, nil
	}

	
	
	q := fmt.Sprintf(`
		SELECT DISTINCT page_id
		FROM %s
		WHERE page_id IN %s
		  AND full_picture != ''
		  AND url_refreshed_at < now() - INTERVAL 10 DAY
	`, tableName, FormatAccountIDs(validPageIDs))

	rows, err := c.Conn.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("Client.GetDistinctFacebookPageIDsWithStaleURLs: query failed: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("Client.GetDistinctFacebookPageIDsWithStaleURLs: scan failed: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (c *Client) MarkFacebookPostsRefreshed(ctx context.Context, tableName, pageID string) error {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_posts"
	}
	if strings.TrimSpace(pageID) == "" {
		return nil
	}
	now := time.Now()
	ts := now.Unix()
	threshold := now.AddDate(0, 0, -10).Unix()
	q := fmt.Sprintf(`
		ALTER TABLE %s
		UPDATE url_refreshed_at = toDateTime(%d)
		WHERE page_id = ?
		  AND url_refreshed_at < toDateTime(%d)
	`, tableName, ts, threshold)
	if err := c.Conn.Exec(ctx, q, pageID); err != nil {
		return fmt.Errorf("Client.MarkFacebookPostsRefreshed: %w", err)
	}
	return nil
}

func (c *Client) BulkMarkFacebookPostsRefreshed(ctx context.Context, tableName string, pageIDs []string) error {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_posts"
	}
	if len(pageIDs) == 0 {
		return nil
	}
	ts := time.Now().Unix()
	q := fmt.Sprintf(`
		ALTER TABLE %s
		UPDATE url_refreshed_at = toDateTime(%d)
		WHERE has(CAST(? AS Array(String)), page_id)
		  AND url_refreshed_at < now() - INTERVAL 10 DAY
	`, tableName, ts)
	if err := c.Conn.Exec(ctx, q, pageIDs); err != nil {
		return fmt.Errorf("Client.BulkMarkFacebookPostsRefreshed: %w", err)
	}
	return nil
}

const bulkMutationChunkSize = 5000

func (c *Client) BulkUpdateFullPictures(
	ctx context.Context,
	tableName string,
	rows []clickhousemodels.MinimalPost,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_posts"
	}

	type entry struct {
		postID string
		pageID string
		url    string
	}

	ts := time.Now().Unix()
	updateMap := make(map[string]entry, len(rows))
	clearMap := make(map[string]entry, 16)
	for _, r := range rows {
		if r.PostID == "" {
			continue
		}
		if strings.TrimSpace(r.FullPicture) == "" {
			if _, ok := clearMap[r.PostID]; !ok {
				clearMap[r.PostID] = entry{postID: r.PostID, pageID: r.PageID}
			}
		} else {
			if _, ok := updateMap[r.PostID]; !ok {
				updateMap[r.PostID] = entry{postID: r.PostID, pageID: r.PageID, url: r.FullPicture}
			}
		}
	}

	updatedTime := time.Now().UTC().Format("2006-01-02 15:04:05.000000")
	total := 0

	if len(updateMap) > 0 {
		type pair struct {
			postID string
			pageID string
			url    string
		}
		entries := make([]pair, 0, len(updateMap))
		for _, e := range updateMap {
			entries = append(entries, pair{e.postID, e.pageID, e.url})
		}

		for i := 0; i < len(entries); i += bulkMutationChunkSize {
			end := i + bulkMutationChunkSize
			if end > len(entries) {
				end = len(entries)
			}
			chunk := entries[i:end]

			postIDs := make([]string, 0, len(chunk))
			urls := make([]string, 0, len(chunk))
			pageIDSet := make(map[string]struct{}, 16)
			for _, e := range chunk {
				postIDs = append(postIDs, e.postID)
				urls = append(urls, e.url)
				if e.pageID != "" {
					pageIDSet[e.pageID] = struct{}{}
				}
			}
			pageIDs := make([]string, 0, len(pageIDSet))
			for pid := range pageIDSet {
				pageIDs = append(pageIDs, pid)
			}

			q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  full_picture = transform(
    post_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    full_picture
  ),
  url_refreshed_at = toDateTime(%d),
  metadata = mapUpdate(
    ifNull(metadata, CAST([], 'Map(String, String)')),
    CAST(map('message', 'full picture link is updated'), 'Map(String, String)')
  ),
  updated_time = toDateTime64(?, 6, 'UTC')
WHERE has(CAST(? AS Array(String)), page_id)
  AND has(CAST(? AS Array(String)), post_id)
`, tableName, ts)

			if err := c.Conn.Exec(ctx, q, postIDs, urls, updatedTime, pageIDs, postIDs); err != nil {
				return total, fmt.Errorf("Client.BulkUpdateFullPictures: update failed: %w", err)
			}
			total += len(postIDs)
			c.Logger.Info().Int("updated", len(postIDs)).Int("pages", len(pageIDs)).Msg("Bulk-updated Facebook post full_picture URLs")
		}
	}

	if len(clearMap) > 0 {
		type pair struct {
			postID string
			pageID string
		}
		entries := make([]pair, 0, len(clearMap))
		for _, e := range clearMap {
			entries = append(entries, pair{e.postID, e.pageID})
		}

		for i := 0; i < len(entries); i += bulkMutationChunkSize {
			end := i + bulkMutationChunkSize
			if end > len(entries) {
				end = len(entries)
			}
			chunk := entries[i:end]

			clearIDs := make([]string, 0, len(chunk))
			pageIDSet := make(map[string]struct{}, 16)
			for _, e := range chunk {
				clearIDs = append(clearIDs, e.postID)
				if e.pageID != "" {
					pageIDSet[e.pageID] = struct{}{}
				}
			}
			pageIDs := make([]string, 0, len(pageIDSet))
			for pid := range pageIDSet {
				pageIDs = append(pageIDs, pid)
			}

			q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  full_picture = '',
  url_refreshed_at = toDateTime(%d)
WHERE has(CAST(? AS Array(String)), page_id)
  AND has(CAST(? AS Array(String)), post_id)
`, tableName, ts)

			if err := c.Conn.Exec(ctx, q, pageIDs, clearIDs); err != nil {
				return total, fmt.Errorf("Client.BulkUpdateFullPictures: clear failed: %w", err)
			}
			c.Logger.Info().Int("cleared", len(clearIDs)).Msg("Cleared full_picture for permanently inaccessible Facebook posts")
		}
	}

	return total, nil
}
