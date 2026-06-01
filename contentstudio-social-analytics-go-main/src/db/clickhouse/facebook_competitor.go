package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	models "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

const competitorURLUpdateBatchSize = 50

// InsertCompetitorInsights inserts a batch of Facebook competitor insights into ClickHouse.
func (c *Client) InsertCompetitorInsights(ctx context.Context, insights []*models.FacebookCompetitorInsights) error {
	if len(insights) == 0 {
		c.Logger.Debug().Msg("No Facebook competitor insights to insert")
		return nil
	}

	query := fmt.Sprintf(`INSERT INTO %s (
		record_id, page_id, followers_count, total_fan_count, talking_about_this,
		biography, profile_picture_url, page_name, page_category, emails,
		birthday, were_here_count, cover_photo_url, permalink, metadata, inserted_at
	)`, models.FacebookCompetitorInsights{}.TableName())

	batch, err := c.Conn.PrepareBatch(ctx, query)
	if err != nil {
		return fmt.Errorf("Client.InsertCompetitorInsights: failed to prepare Facebook competitor insights batch: %w", err)
	}

	for _, insight := range insights {
		if err := batch.Append(
			insight.RecordID,
			insight.PageID,
			insight.FollowersCount,
			insight.TotalFanCount,
			insight.TalkingAboutThis,
			insight.Biography,
			insight.ProfilePictureURL,
			insight.PageName,
			insight.PageCategory,
			insight.Emails,
			insight.Birthday,
			insight.WereHereCount,
			insight.CoverPhotoURL,
			insight.Permalink,
			insight.Metadata,
			insight.InsertedAt,
		); err != nil {
			return fmt.Errorf("Client.InsertCompetitorInsights: failed to append Facebook competitor insight: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.InsertCompetitorInsights: failed to send Facebook competitor insights batch: %w", err)
	}

	c.Logger.Info().
		Int("count", len(insights)).
		Str("table", models.FacebookCompetitorInsights{}.TableName()).
		Msg("Inserted Facebook competitor insights successfully")
	return nil
}

// InsertCompetitorPosts inserts a batch of Facebook competitor posts into ClickHouse.
func (c *Client) InsertCompetitorPosts(ctx context.Context, posts []*models.FacebookCompetitorPosts) error {
	if len(posts) == 0 {
		c.Logger.Debug().Msg("No Facebook competitor posts to insert")
		return nil
	}

	tableName := models.FacebookCompetitorPosts{}.TableName()

	batch, err := c.Conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", tableName))
	if err != nil {
		return fmt.Errorf("Client.InsertCompetitorPosts: failed to prepare Facebook competitor posts batch: %w", err)
	}

	for _, post := range posts {
		if err := batch.Append(
			post.FacebookID,
			post.PostID,
			post.FollowersCount,
			post.FanCount,
			post.PageName,
			post.PageCategory,
			post.Biography,
			post.PostEngagement,
			post.Like,
			post.Haha,
			post.Angry,
			post.Sad,
			post.Thankful,
			post.Love,
			post.TotalPostReactions,
			post.Comments,
			post.Shares,
			post.Caption,
			post.MediaType,
			post.StatusType,
			post.SharedFromName,
			post.SharedFromID,
			post.SharedFromPic,
			post.SharedCreatedAt,
			post.Permalink,
			post.Hashtags,
			post.DayOfWeek,
			post.HourOfDay,
			post.CreatedAt,
			post.InsertedAt,
			post.Wow,
		); err != nil {
			return fmt.Errorf("Client.InsertCompetitorPosts: failed to append Facebook competitor post: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.InsertCompetitorPosts: failed to send Facebook competitor posts batch: %w", err)
	}

	c.Logger.Info().
		Int("count", len(posts)).
		Str("table", tableName).
		Msg("Inserted Facebook competitor posts successfully")
	return nil
}

// InsertCompetitorMediaAssets inserts a batch of Facebook competitor media assets into ClickHouse.
func (c *Client) InsertCompetitorMediaAssets(ctx context.Context, assets []*models.FacebookCompetitorMediaAssets) error {
	if len(assets) == 0 {
		c.Logger.Debug().Msg("No Facebook competitor media assets to insert")
		return nil
	}

	tableName := models.FacebookCompetitorMediaAssets{}.TableName()

	batch, err := c.Conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", tableName))
	if err != nil {
		return fmt.Errorf("Client.InsertCompetitorMediaAssets: failed to prepare Facebook competitor media assets batch: %w", err)
	}

	for _, asset := range assets {
		if err := batch.Append(
			asset.MediaID,
			asset.PostID,
			asset.PageID,
			asset.Caption,
			asset.Description,
			asset.Link,
			asset.AssetType,
			asset.CallToAction,
			asset.CTAType,
			asset.CreatedAt,
			asset.InsertedAt,
		); err != nil {
			return fmt.Errorf("Client.InsertCompetitorMediaAssets: failed to append Facebook competitor media asset: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.InsertCompetitorMediaAssets: failed to send Facebook competitor media assets batch: %w", err)
	}

	c.Logger.Info().
		Int("count", len(assets)).
		Str("table", tableName).
		Msg("Inserted Facebook competitor media assets successfully")
	return nil
}

func (c *Client) GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccount(
	ctx context.Context,
	tableName string,
	facebookID string,
	limit, offset int,
) ([]models.FacebookCompetitorMinimalMediaAsset, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_competitor_media_assets"
	}
	if strings.TrimSpace(facebookID) == "" {
		return nil, fmt.Errorf("Client.GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccount: facebookID is required")
	}

	q := fmt.Sprintf(`
		SELECT
			page_id,
			post_id,
			media_id,
			link,
			created_at
		FROM %s
		WHERE page_id = ?
		  AND created_at < now() - INTERVAL 10 DAY
		  AND link != ''
		ORDER BY created_at DESC, media_id DESC
		LIMIT ? OFFSET ?
	`, tableName)

	rows, err := c.Conn.Query(ctx, q, facebookID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.FacebookCompetitorMinimalMediaAsset, 0, 1024)
	for rows.Next() {
		var row models.FacebookCompetitorMinimalMediaAsset
		if err := rows.ScanStruct(&row); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (c *Client) UpdateFacebookCompetitorMediaAssetURLs(
	ctx context.Context,
	tableName, facebookID string,
	rows []models.FacebookCompetitorMinimalMediaAsset,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_competitor_media_assets"
	}
	if strings.TrimSpace(facebookID) == "" {
		return 0, fmt.Errorf("Client.UpdateFacebookCompetitorMediaAssetURLs: facebookID is required")
	}

	mediaIDs := make([]string, 0, len(rows))
	links := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))

	for _, row := range rows {
		if row.PageID != facebookID || strings.TrimSpace(row.MediaID) == "" {
			continue
		}
		if _, ok := seen[row.MediaID]; ok {
			continue
		}
		seen[row.MediaID] = struct{}{}
		mediaIDs = append(mediaIDs, row.MediaID)
		links = append(links, row.Link)
	}
	if len(mediaIDs) == 0 {
		return 0, nil
	}

	q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  link = transform(
    media_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    link
  ),
  inserted_at = toDateTime64(?, 6, 'UTC')
WHERE page_id = ?
  AND has(CAST(? AS Array(String)), media_id)
	`, tableName)

	for start := 0; start < len(mediaIDs); start += competitorURLUpdateBatchSize {
		end := start + competitorURLUpdateBatchSize
		if end > len(mediaIDs) {
			end = len(mediaIDs)
		}

		updatedAt := time.Now().UTC().Format("2006-01-02 15:04:05.000000")
		if err := c.Conn.Exec(ctx, q, mediaIDs[start:end], links[start:end], updatedAt, facebookID, mediaIDs[start:end]); err != nil {
			return 0, fmt.Errorf("Client.UpdateFacebookCompetitorMediaAssetURLs: update failed: %w", err)
		}
	}

	return len(mediaIDs), nil
}

func (c *Client) GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccount(
	ctx context.Context,
	tableName string,
	facebookID string,
	limit, offset int,
) ([]models.FacebookCompetitorMinimalSharedPost, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_competitor_posts"
	}
	if strings.TrimSpace(facebookID) == "" {
		return nil, fmt.Errorf("Client.GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccount: facebookID is required")
	}

	q := fmt.Sprintf(`
		SELECT
			facebook_id,
			post_id,
			shared_from_id,
			shared_from_pic,
			created_at
		FROM %s
		WHERE facebook_id = ?
		  AND created_at < now() - INTERVAL 10 DAY
		  AND length(shared_from_id) > 0
		  AND shared_from_pic != ''
		ORDER BY created_at DESC, post_id DESC
		LIMIT ? OFFSET ?
	`, tableName)

	rows, err := c.Conn.Query(ctx, q, facebookID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.FacebookCompetitorMinimalSharedPost, 0, 1024)
	for rows.Next() {
		var row models.FacebookCompetitorMinimalSharedPost
		if err := rows.ScanStruct(&row); err != nil {
			return nil, err
		}
		out = append(out, row)
	}
	return out, rows.Err()
}

func (c *Client) UpdateFacebookCompetitorSharedPictures(
	ctx context.Context,
	tableName, facebookID string,
	rows []models.FacebookCompetitorMinimalSharedPost,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_competitor_posts"
	}
	if strings.TrimSpace(facebookID) == "" {
		return 0, fmt.Errorf("Client.UpdateFacebookCompetitorSharedPictures: facebookID is required")
	}

	postIDs := make([]string, 0, len(rows))
	pics := make([]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))
	for _, row := range rows {
		if row.FacebookID != facebookID || strings.TrimSpace(row.PostID) == "" || strings.TrimSpace(row.SharedFromPic) == "" {
			continue
		}
		if _, ok := seen[row.PostID]; ok {
			continue
		}
		seen[row.PostID] = struct{}{}
		postIDs = append(postIDs, row.PostID)
		pics = append(pics, row.SharedFromPic)
	}
	if len(postIDs) == 0 {
		return 0, nil
	}

	q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  shared_from_pic = transform(
    post_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    shared_from_pic
  ),
  inserted_at = toDateTime64(?, 6, 'UTC')
WHERE facebook_id = ?
  AND has(CAST(? AS Array(String)), post_id)
	`, tableName)

	for start := 0; start < len(postIDs); start += competitorURLUpdateBatchSize {
		end := start + competitorURLUpdateBatchSize
		if end > len(postIDs) {
			end = len(postIDs)
		}

		updatedAt := time.Now().UTC().Format("2006-01-02 15:04:05.000000")
		if err := c.Conn.Exec(ctx, q, postIDs[start:end], pics[start:end], updatedAt, facebookID, postIDs[start:end]); err != nil {
			return 0, fmt.Errorf("Client.UpdateFacebookCompetitorSharedPictures: update failed: %w", err)
		}
	}

	return len(postIDs), nil
}

// BulkUpdateFacebookCompetitorMediaAssetURLs issues one ALTER TABLE UPDATE per partition
// month. Adding toYYYYMM(created_at) = ? to the WHERE clause tells ClickHouse to only
// rewrite parts in that one partition, so the total I/O is proportional to the stale data
// months (typically 6–18), not to the total table size or account count.
func (c *Client) BulkUpdateFacebookCompetitorMediaAssetURLs(
	ctx context.Context,
	tableName string,
	assets []models.FacebookCompetitorMinimalMediaAsset,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_competitor_media_assets"
	}

	type entry struct {
		link      string
		partition uint32
	}
	seen := make(map[string]entry, len(assets))
	for _, a := range assets {
		if strings.TrimSpace(a.MediaID) == "" || strings.TrimSpace(a.Link) == "" {
			continue
		}
		if _, ok := seen[a.MediaID]; ok {
			continue
		}
		seen[a.MediaID] = entry{
			link:      a.Link,
			partition: uint32(a.CreatedAt.Year())*100 + uint32(a.CreatedAt.Month()),
		}
	}
	if len(seen) == 0 {
		return 0, nil
	}

	type partGroup struct {
		mediaIDs []string
		links    []string
	}
	byPartition := make(map[uint32]*partGroup)
	for mediaID, e := range seen {
		g := byPartition[e.partition]
		if g == nil {
			g = &partGroup{}
			byPartition[e.partition] = g
		}
		g.mediaIDs = append(g.mediaIDs, mediaID)
		g.links = append(g.links, e.link)
	}

	q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  link = transform(
    media_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    link
  ),
  inserted_at = toDateTime64(?, 6, 'UTC')
WHERE toYYYYMM(created_at) = ?
  AND has(CAST(? AS Array(String)), media_id)
	`, tableName)

	updatedAt := time.Now().UTC().Format("2006-01-02 15:04:05.000000")
	total := 0
	for partition, g := range byPartition {
		if err := c.Conn.Exec(ctx, q, g.mediaIDs, g.links, updatedAt, partition, g.mediaIDs); err != nil {
			return total, fmt.Errorf("Client.BulkUpdateFacebookCompetitorMediaAssetURLs: partition %d update failed: %w", partition, err)
		}
		total += len(g.mediaIDs)
	}

	c.Logger.Info().
		Int("total_updated", total).
		Int("partitions", len(byPartition)).
		Msg("Bulk-updated Facebook competitor media asset URLs in ClickHouse")
	return total, nil
}

// BulkUpdateFacebookCompetitorSharedPictures issues one ALTER TABLE UPDATE per partition
// month using partition pruning, keeping each mutation to a single partition's I/O.
func (c *Client) BulkUpdateFacebookCompetitorSharedPictures(
	ctx context.Context,
	tableName string,
	posts []models.FacebookCompetitorMinimalSharedPost,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "facebook_competitor_posts"
	}

	type entry struct {
		pic       string
		partition uint32
	}
	seen := make(map[string]entry, len(posts))
	for _, p := range posts {
		if strings.TrimSpace(p.PostID) == "" || strings.TrimSpace(p.SharedFromPic) == "" {
			continue
		}
		if _, ok := seen[p.PostID]; ok {
			continue
		}
		seen[p.PostID] = entry{
			pic:       p.SharedFromPic,
			partition: uint32(p.CreatedAt.Year())*100 + uint32(p.CreatedAt.Month()),
		}
	}
	if len(seen) == 0 {
		return 0, nil
	}

	type partGroup struct {
		postIDs []string
		pics    []string
	}
	byPartition := make(map[uint32]*partGroup)
	for postID, e := range seen {
		g := byPartition[e.partition]
		if g == nil {
			g = &partGroup{}
			byPartition[e.partition] = g
		}
		g.postIDs = append(g.postIDs, postID)
		g.pics = append(g.pics, e.pic)
	}

	q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  shared_from_pic = transform(
    post_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    shared_from_pic
  ),
  inserted_at = toDateTime64(?, 6, 'UTC')
WHERE toYYYYMM(created_at) = ?
  AND has(CAST(? AS Array(String)), post_id)
	`, tableName)

	updatedAt := time.Now().UTC().Format("2006-01-02 15:04:05.000000")
	total := 0
	for partition, g := range byPartition {
		if err := c.Conn.Exec(ctx, q, g.postIDs, g.pics, updatedAt, partition, g.postIDs); err != nil {
			return total, fmt.Errorf("Client.BulkUpdateFacebookCompetitorSharedPictures: partition %d update failed: %w", partition, err)
		}
		total += len(g.postIDs)
	}

	c.Logger.Info().
		Int("total_updated", total).
		Int("partitions", len(byPartition)).
		Msg("Bulk-updated Facebook competitor shared pictures in ClickHouse")
	return total, nil
}
