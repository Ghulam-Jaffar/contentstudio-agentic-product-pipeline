package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// InsertInstagramCompetitorInsights batch inserts Instagram competitor insights into ClickHouse.
func (c *Client) InsertInstagramCompetitorInsights(ctx context.Context, insights []*models.InstagramCompetitorInsights) error {
	if len(insights) == 0 {
		c.Logger.Debug().Msg("No Instagram competitor insights to insert")
		return nil
	}

	tableName := models.InstagramCompetitorInsights{}.TableName()

	query := fmt.Sprintf(`INSERT INTO %s (
		record_id, instagram_account_id, total_followed_by_count, total_following_count,
		profile_picture_url, page_name, metadata, inserted_at
	)`, tableName)

	batch, err := c.Conn.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("Client.InsertInstagramCompetitorInsights: failed to prepare Instagram competitor insights batch: %w", err)
	}

	for _, insight := range insights {
		if err := batch.Append(
			insight.RecordID,
			insight.InstagramAccountID,
			insight.TotalFollowedByCount,
			insight.TotalFollowingCount,
			insight.ProfilePictureURL,
			insight.PageName,
			insight.Metadata,
			insight.InsertedAt,
		); err != nil {
			return fmt.Errorf("Client.InsertInstagramCompetitorInsights: failed to append Instagram competitor insight: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.InsertInstagramCompetitorInsights: failed to send Instagram competitor insights batch: %w", err)
	}

	c.Logger.Info().
		Int("count", len(insights)).
		Str("table", tableName).
		Msg("Successfully inserted Instagram competitor insights")
	return nil
}

// InsertInstagramCompetitorPosts batch inserts Instagram competitor posts into ClickHouse.
func (c *Client) InsertInstagramCompetitorPosts(ctx context.Context, posts []*models.InstagramCompetitorPosts) error {
	if len(posts) == 0 {
		c.Logger.Debug().Msg("No Instagram competitor posts to insert")
		return nil
	}

	tableName := models.InstagramCompetitorPosts{}.TableName()

	batch, err := c.Conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", tableName))
	if err != nil {
		return fmt.Errorf("Client.InsertInstagramCompetitorPosts: failed to prepare Instagram competitor posts batch: %w", err)
	}

	for _, post := range posts {
		if err := batch.Append(
			post.InstagramID,
			post.PostID,
			post.BusinessAccountID,
			post.TotalFollowedByCount,
			post.TotalFollowingCount,
			post.Username,
			post.Name,
			post.PageCategory,
			post.ProfilePictureURL,
			post.Biography,
			post.Engagement,
			post.LikeCount,
			post.CommentsCount,
			post.MediaCount,
			post.Caption,
			post.MediaType,
			post.MediaProductType,
			post.MediaURL,
			post.Permalink,
			post.Hashtags,
			post.CreatedAt,
			post.InsertedAt,
		); err != nil {
			return fmt.Errorf("Client.InsertInstagramCompetitorPosts: failed to append Instagram competitor post: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.InsertInstagramCompetitorPosts: failed to send Instagram competitor posts batch: %w", err)
	}

	c.Logger.Info().
		Int("count", len(posts)).
		Str("table", tableName).
		Msg("Successfully inserted Instagram competitor posts")
	return nil
}

func (c *Client) GetMinimalInstagramCompetitorOlderThan7DaysByAccount(
	ctx context.Context,
	tableName string,
	instagramID int64,
	limit, offset int,
) ([]models.InstagramCompetitorMinimalPost, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "instagram_competitor_posts"
	}
	if instagramID == 0 {
		return nil, fmt.Errorf("Client.GetMinimalInstagramCompetitorOlderThan7DaysByAccount: instagramID is required")
	}

	q := fmt.Sprintf(`
		SELECT
			instagram_id,
			post_id,
			media_url,
			profile_picture_url,
			created_at
		FROM %s
		WHERE (instagram_id = ? OR business_account_id = toString(?))
		  AND created_at < now() - INTERVAL 10 DAY
		  AND (media_url != '' OR profile_picture_url != '')
		ORDER BY created_at DESC, post_id DESC
		LIMIT ? OFFSET ?
	`, tableName)

	rows, err := c.Conn.Query(ctx, q, instagramID, instagramID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.InstagramCompetitorMinimalPost, 0, 1024)
	for rows.Next() {
		var row models.InstagramCompetitorMinimalPost
		if err := rows.ScanStruct(&row); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (c *Client) UpdateInstagramCompetitorMediaURLs(
	ctx context.Context,
	tableName string,
	instagramID int64,
	profilePictureURL string,
	rows []models.InstagramCompetitorMinimalPost,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "instagram_competitor_posts"
	}
	if instagramID == 0 {
		return 0, fmt.Errorf("Client.UpdateInstagramCompetitorMediaURLs: instagramID is required")
	}

	postIDs := make([]string, 0, len(rows))
	mediaURLs := make([]string, 0, len(rows))
	type selectedRow struct {
		mediaURL string
		prefer   bool
	}
	selected := make(map[string]selectedRow, len(rows))
	order := make([]string, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.PostID) == "" {
			continue
		}

		prefer := row.InstagramID == instagramID
		if current, ok := selected[row.PostID]; ok {
			// If we have duplicate post IDs, prefer the row that matches the target
			// instagram_id. This avoids selecting an unrelated duplicate when the
			// query returns rows joined via business_account_id.
			if !current.prefer && prefer {
				selected[row.PostID] = selectedRow{mediaURL: row.MediaURL, prefer: true}
			}
			continue
		}

		selected[row.PostID] = selectedRow{mediaURL: row.MediaURL, prefer: prefer}
		order = append(order, row.PostID)
	}

	for _, postID := range order {
		postIDs = append(postIDs, postID)
		mediaURLs = append(mediaURLs, selected[postID].mediaURL)
	}
	if len(postIDs) == 0 && strings.TrimSpace(profilePictureURL) == "" {
		return 0, nil
	}

	// Statement 1: refresh media_url + bump inserted_at only for the resolved batch.
	// Using has() with no OR condition ensures other rows' staleness timestamps are
	// NOT touched — they must remain stale for future refresh runs.
	if len(postIDs) > 0 {
		qMedia := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  media_url = transform(
    post_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    media_url
  ),
  inserted_at = toDateTime64(?, 6, 'UTC')
WHERE (instagram_id = ? OR business_account_id = toString(?))
  AND has(CAST(? AS Array(String)), post_id)
		`, tableName)

		for start := 0; start < len(postIDs); start += competitorURLUpdateBatchSize {
			end := start + competitorURLUpdateBatchSize
			if end > len(postIDs) {
				end = len(postIDs)
			}
			updatedAt := time.Now().UTC().Format("2006-01-02 15:04:05.000000")
			if err := c.Conn.Exec(ctx, qMedia, postIDs[start:end], mediaURLs[start:end], updatedAt, instagramID, instagramID, postIDs[start:end]); err != nil {
				return 0, fmt.Errorf("Client.UpdateInstagramCompetitorMediaURLs: media update failed: %w", err)
			}
		}
	}

	// Statement 2: propagate the account's profile picture to all its posts.
	// Intentionally does NOT change inserted_at — profile_picture_url is page-level
	// metadata and updating it must not reset staleness for post media_urls.
	if strings.TrimSpace(profilePictureURL) != "" {
		qPic := fmt.Sprintf(`
ALTER TABLE %s
UPDATE profile_picture_url = ?
WHERE (instagram_id = ? OR business_account_id = toString(?))
		`, tableName)
		if err := c.Conn.Exec(ctx, qPic, profilePictureURL, instagramID, instagramID); err != nil {
			return 0, fmt.Errorf("Client.UpdateInstagramCompetitorMediaURLs: profile picture update failed: %w", err)
		}
	}

	return len(postIDs), nil
}

// BulkUpdateInstagramCompetitorMediaURLs issues one ALTER TABLE UPDATE per partition month
// for media_url updates (partition-pruned), plus two partition-spanning mutations for
// profile_picture_url (account-level metadata that must touch all partitions for an account).
// profilePics maps instagram_id → profile_picture_url for accounts that returned a new URL.
func (c *Client) BulkUpdateInstagramCompetitorMediaURLs(
	ctx context.Context,
	tableName string,
	posts []models.InstagramCompetitorMinimalPost,
	profilePics map[int64]string,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "instagram_competitor_posts"
	}

	updatedAt := time.Now().UTC().Format("2006-01-02 15:04:05.000000")

	// --- Media URL: partition-aware, one mutation per month ---
	type mediaEntry struct {
		mediaURL  string
		partition uint32
	}
	seenPosts := make(map[string]mediaEntry, len(posts))
	for _, p := range posts {
		if strings.TrimSpace(p.PostID) == "" {
			continue
		}
		if _, ok := seenPosts[p.PostID]; ok {
			continue
		}
		seenPosts[p.PostID] = mediaEntry{
			mediaURL:  p.MediaURL,
			partition: uint32(p.CreatedAt.Year())*100 + uint32(p.CreatedAt.Month()),
		}
	}

	if len(seenPosts) > 0 {
		type partGroup struct {
			postIDs   []string
			mediaURLs []string
		}
		byPartition := make(map[uint32]*partGroup)
		for postID, e := range seenPosts {
			g := byPartition[e.partition]
			if g == nil {
				g = &partGroup{}
				byPartition[e.partition] = g
			}
			g.postIDs = append(g.postIDs, postID)
			g.mediaURLs = append(g.mediaURLs, e.mediaURL)
		}

		qMedia := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  media_url = transform(
    post_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    media_url
  ),
  inserted_at = toDateTime64(?, 6, 'UTC')
WHERE toYYYYMM(created_at) = ?
  AND has(CAST(? AS Array(String)), post_id)
		`, tableName)

		for partition, g := range byPartition {
			if err := c.Conn.Exec(ctx, qMedia, g.postIDs, g.mediaURLs, updatedAt, partition, g.postIDs); err != nil {
				return 0, fmt.Errorf("Client.BulkUpdateInstagramCompetitorMediaURLs: media update partition %d failed: %w", partition, err)
			}
		}
	}

	// --- Profile picture: account-level metadata spans all partitions.
	// Two mutations cover both storage patterns (numeric instagram_id vs string business_account_id).
	// Not partition-prunable, but the array is bounded by the number of accounts processed
	// and profile pictures change rarely, keeping this lightweight in practice.
	if len(profilePics) > 0 {
		igIDs := make([]string, 0, len(profilePics))
		picURLs := make([]string, 0, len(profilePics))
		for id, pic := range profilePics {
			if strings.TrimSpace(pic) == "" {
				continue
			}
			igIDs = append(igIDs, fmt.Sprintf("%d", id))
			picURLs = append(picURLs, pic)
		}
		if len(igIDs) > 0 {
			qByID := fmt.Sprintf(`
ALTER TABLE %s
UPDATE profile_picture_url = transform(
  toString(instagram_id),
  CAST(? AS Array(String)),
  CAST(? AS Array(String)),
  profile_picture_url
)
WHERE has(CAST(? AS Array(String)), toString(instagram_id))
			`, tableName)
			if err := c.Conn.Exec(ctx, qByID, igIDs, picURLs, igIDs); err != nil {
				return 0, fmt.Errorf("Client.BulkUpdateInstagramCompetitorMediaURLs: profile picture by instagram_id failed: %w", err)
			}

			qByBAID := fmt.Sprintf(`
ALTER TABLE %s
UPDATE profile_picture_url = transform(
  business_account_id,
  CAST(? AS Array(String)),
  CAST(? AS Array(String)),
  profile_picture_url
)
WHERE has(CAST(? AS Array(String)), business_account_id)
			`, tableName)
			if err := c.Conn.Exec(ctx, qByBAID, igIDs, picURLs, igIDs); err != nil {
				return 0, fmt.Errorf("Client.BulkUpdateInstagramCompetitorMediaURLs: profile picture by business_account_id failed: %w", err)
			}
		}
	}

	c.Logger.Info().
		Int("posts_updated", len(seenPosts)).
		Int("accounts_with_profile_pics", len(profilePics)).
		Msg("Bulk-updated Instagram competitor media URLs in ClickHouse")
	return len(seenPosts), nil
}
