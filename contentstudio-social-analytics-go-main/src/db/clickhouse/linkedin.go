package clickhouse

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// BulkInsertLinkedInPosts inserts a batch of LinkedIn posts into ClickHouse.
func (c *Client) BulkInsertLinkedInPosts(ctx context.Context, posts []*chmodels.LinkedInPosts) error {
	if len(posts) == 0 {
		return nil
	}

	c.Logger.Info().
		Str("table", "linkedin_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `INSERT INTO linkedin_posts (
        linkedin_id, post_id, activity, comments, total_engagement, favorites, poll_data, reach, repost, post_clicks, impressions, title, image, article_url, article_title, media, media_type, type, hashtags, day_of_week, hour_of_day, created_at, published_at, last_modified_at, lifecycle_state, visibility, saving_time, is_reshare_disabled, feed_distribution, third_party_channels, url_refreshed_at
    ) VALUES`)
	if err != nil {
		return err
	}

	now := time.Now()
	for _, p := range posts {
		if err := batch.Append(
			p.LinkedinID,
			p.PostID,
			p.Activity,
			p.Comments,
			p.TotalEngagement,
			p.Favorites,
			p.PollData,
			p.Reach,
			p.Repost,
			p.PostClicks,
			p.Impressions,
			p.Title,
			p.Image,
			p.ArticleURL,
			p.ArticleTitle,
			p.Media,
			p.MediaType,
			p.Type,
			p.Hashtags,
			p.DayOfWeek,
			p.HourOfDay,
			p.CreatedAt,
			p.PublishedAt,
			p.LastModifiedAt,
			p.LifecycleState,
			p.Visibility,
			p.SavingTime,
			p.IsReshareDisabled,
			p.FeedDistribution,
			p.ThirdPartyChannels,
			now,
		); err != nil {
			return err
		}
	}

	if err := batch.Send(); err != nil {
		return err
	}

	c.Logger.Info().
		Str("table", "linkedin_posts").
		Int("batch_size", len(posts)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")

	return nil
}

// BulkInsertLinkedInInsights inserts LinkedIn insights batch.
func (c *Client) BulkInsertLinkedInInsights(ctx context.Context, insights []*chmodels.LinkedInInsights) error {
	if len(insights) == 0 {
		return nil
	}

	c.Logger.Info().
		Str("table", "linkedin_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Starting batch insert")

	// Log each insight being inserted for debugging
	for i, in := range insights {
		c.Logger.Debug().
			Int("index", i).
			Str("linkedin_id", in.LinkedinID).
			Str("record_id", in.RecordID).
			Time("created_at", in.CreatedAt).
			Int64("impressions", in.ImpressionCount).
			Int64("page_views", in.PageViews).
			Int64("followers", in.TotalFollowerCount).
			Msg("Inserting insight row")
	}

	// Prevent silent block-level dedup; optionally wait for async-insert if you use it.
	settings := clickhouse.Settings{
		"insert_deduplicate": 0,
		// "async_insert": 1,
		// "wait_for_async_insert": 1,
		// "wait_for_async_insert_timeout": 30_000,
	}
	ctx = clickhouse.Context(ctx, clickhouse.WithSettings(settings))

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO linkedin_insights (
			linkedin_id,
			organization_name,
			record_id,
			impressionCount,
			organicFollowerCount,
			totalFollowerCount,
			paidFollowerCount,
			daily_follower_count,
			reach,
			repost,
			comments,
			post_clicks,
			reactions,
			engagement,
			followers_by_seniority,
			followers_by_industry,
			followers_by_country,
			followers_by_city,
			inserted_at,
			created_at,
			page_views,
			unique_visitors,
			desktop_page_views,
			mobile_page_views,
			overview_page_views,
			about_page_views,
			jobs_page_views,
			people_page_views,
			careers_page_views,
			life_at_page_views,
			insights_page_views,
			products_page_views,
			page_views_by_country,
			page_views_by_region,
			page_views_by_industry,
			page_views_by_seniority,
			page_views_by_function,
			page_views_by_staff_count
		) SETTINGS insert_deduplicate=0 VALUES
	`)
	if err != nil {
		return err
	}

	for _, in := range insights {
		if err := batch.Append(
			in.LinkedinID,
			in.OrganizationName,
			in.RecordID,
			in.ImpressionCount,
			in.OrganicFollowerCount,
			in.TotalFollowerCount,
			in.PaidFollowerCount,
			in.DailyFollowerCount,
			in.Reach,
			in.Repost,
			in.Comments,
			in.PostClicks,
			in.Reactions,
			in.Engagement,
			in.FollowersBySeniority,
			in.FollowersByIndustry,
			in.FollowersByCountry,
			in.FollowersByCity,
			in.InsertedAt,
			in.CreatedAt,
			in.PageViews,
			in.UniqueVisitors,
			in.DesktopPageViews,
			in.MobilePageViews,
			in.OverviewPageViews,
			in.AboutPageViews,
			in.JobsPageViews,
			in.PeoplePageViews,
			in.CareersPageViews,
			in.LifeAtPageViews,
			in.InsightsPageViews,
			in.ProductsPageViews,
			in.PageViewsByCountry,
			in.PageViewsByRegion,
			in.PageViewsByIndustry,
			in.PageViewsBySeniority,
			in.PageViewsByFunction,
			in.PageViewsByStaffCount,
		); err != nil {
			return err
		}
	}

	if err := batch.Send(); err != nil {
		return err
	}

	c.Logger.Info().
		Str("table", "linkedin_insights").
		Int("batch_size", len(insights)).
		Int("batch_count", 1).
		Msg("Batch insert completed successfully")

	return nil
}

func (c *Client) GetMinimalLinkedInOlderThan7DaysByAccount(
	ctx context.Context,
	tableName string,
	linkedinID string,
	limit, offset int,
) ([]chmodels.LinkedInMinimalPost, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "linkedin_posts"
	}
	if strings.TrimSpace(linkedinID) == "" {
		return nil, fmt.Errorf("Client.GetMinimalLinkedInOlderThan7DaysByAccount: linkedinID is required")
	}

	q := fmt.Sprintf(`
		SELECT
			linkedin_id,
			post_id,
			activity,
			image,
			media
		FROM %s
		WHERE linkedin_id = ?
		  AND (image != '' OR length(media) > 0)
		  AND url_refreshed_at < now() - INTERVAL 10 DAY
		ORDER BY published_at DESC, post_id DESC
		LIMIT ? OFFSET ?
	`, tableName)

	rows, err := c.Conn.Query(ctx, q, linkedinID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]chmodels.LinkedInMinimalPost, 0, 1024)
	for rows.Next() {
		var m chmodels.LinkedInMinimalPost
		if err := rows.ScanStruct(&m); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

func (c *Client) GetDistinctLinkedInIDsWithStaleURLs(ctx context.Context, tableName string, validIDs []string) ([]string, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "linkedin_posts"
	}
	if len(validIDs) == 0 {
		return nil, nil
	}

q := fmt.Sprintf(`
		SELECT DISTINCT linkedin_id
		FROM %s
		WHERE linkedin_id IN %s
		  AND (image != '' OR length(media) > 0)
		  AND url_refreshed_at < now() - INTERVAL 10 DAY
	`, tableName, FormatAccountIDs(validIDs))

	rows, err := c.Conn.Query(ctx, q)
	if err != nil {
		return nil, fmt.Errorf("Client.GetDistinctLinkedInIDsWithStaleURLs: query failed: %w", err)
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("Client.GetDistinctLinkedInIDsWithStaleURLs: scan failed: %w", err)
		}
		ids = append(ids, id)
	}
	return ids, rows.Err()
}

func (c *Client) MarkLinkedInPostsRefreshed(ctx context.Context, tableName, linkedinID string) error {
	if strings.TrimSpace(tableName) == "" {
		tableName = "linkedin_posts"
	}
	if strings.TrimSpace(linkedinID) == "" {
		return nil
	}
	now := time.Now()
	ts := now.Unix()
	threshold := now.AddDate(0, 0, -10).Unix()
	q := fmt.Sprintf(`
		ALTER TABLE %s
		UPDATE url_refreshed_at = toDateTime(%d)
		WHERE linkedin_id = ?
		  AND url_refreshed_at < toDateTime(%d)
	`, tableName, ts, threshold)
	if err := c.Conn.Exec(ctx, q, linkedinID); err != nil {
		return fmt.Errorf("Client.MarkLinkedInPostsRefreshed: %w", err)
	}
	return nil
}

func (c *Client) BulkMarkLinkedInPostsRefreshed(ctx context.Context, tableName string, linkedinIDs []string) error {
	if strings.TrimSpace(tableName) == "" {
		tableName = "linkedin_posts"
	}
	if len(linkedinIDs) == 0 {
		return nil
	}
	ts := time.Now().Unix()
	q := fmt.Sprintf(`
		ALTER TABLE %s
		UPDATE url_refreshed_at = toDateTime(%d)
		WHERE has(CAST(? AS Array(String)), linkedin_id)
		  AND url_refreshed_at < now() - INTERVAL 10 DAY
	`, tableName, ts)
	if err := c.Conn.Exec(ctx, q, linkedinIDs); err != nil {
		return fmt.Errorf("Client.BulkMarkLinkedInPostsRefreshed: %w", err)
	}
	return nil
}

func (c *Client) BulkUpdateLinkedInPostURLs(
	ctx context.Context,
	tableName string,
	rows []chmodels.LinkedInMinimalPost,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "linkedin_posts"
	}

	type entry struct {
		postID string
		image  string
		media  []string
	}

	ts := time.Now().Unix()
	seen := make(map[string]struct{}, len(rows))
	entries := make([]entry, 0, len(rows))
	for _, row := range rows {
		if strings.TrimSpace(row.PostID) == "" {
			continue
		}
		if _, ok := seen[row.PostID]; ok {
			continue
		}
		seen[row.PostID] = struct{}{}
		entries = append(entries, entry{row.PostID, row.Image, row.Media})
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

		postIDs := make([]string, 0, len(chunk))
		images := make([]string, 0, len(chunk))
		media := make([][]string, 0, len(chunk))
		for _, e := range chunk {
			postIDs = append(postIDs, e.postID)
			images = append(images, e.image)
			media = append(media, e.media)
		}

		q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  image = transform(
    post_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    image
  ),
  media = if(
    indexOf(CAST(? AS Array(String)), post_id) > 0,
    arrayElement(CAST(? AS Array(Array(String))), indexOf(CAST(? AS Array(String)), post_id)),
    media
  ),
  url_refreshed_at = toDateTime(%d)
WHERE has(CAST(? AS Array(String)), post_id)
`, tableName, ts)

		if err := c.Conn.Exec(ctx, q,
			postIDs, images,
			postIDs, media, postIDs,
			postIDs,
		); err != nil {
			return total, fmt.Errorf("Client.BulkUpdateLinkedInPostURLs: update failed: %w", err)
		}
		total += len(postIDs)
	}

	c.Logger.Info().
		Str("table", tableName).
		Int("post_ids", total).
		Msg("Bulk-updated LinkedIn post URLs")

	return total, nil
}

func (c *Client) UpdateLinkedInPostURLs(
	ctx context.Context,
	tableName string,
	linkedinID string,
	rows []chmodels.LinkedInMinimalPost,
) (int, error) {
	if strings.TrimSpace(tableName) == "" {
		tableName = "linkedin_posts"
	}
	if strings.TrimSpace(linkedinID) == "" {
		return 0, fmt.Errorf("Client.UpdateLinkedInPostURLs: linkedinID is required")
	}

	postIDs := make([]string, 0, len(rows))
	images := make([]string, 0, len(rows))
	media := make([][]string, 0, len(rows))
	seen := make(map[string]struct{}, len(rows))

	for _, row := range rows {
		if row.LinkedinID != linkedinID {
			continue
		}
		if strings.TrimSpace(row.PostID) == "" {
			continue
		}
		if _, ok := seen[row.PostID]; ok {
			continue
		}
		seen[row.PostID] = struct{}{}
		postIDs = append(postIDs, row.PostID)
		images = append(images, row.Image)
		media = append(media, row.Media)
	}

	if len(postIDs) == 0 {
		return 0, nil
	}

	q := fmt.Sprintf(`
ALTER TABLE %s
UPDATE
  image = transform(
    post_id,
    CAST(? AS Array(String)),
    CAST(? AS Array(String)),
    image
  ),
  media = if(
    indexOf(CAST(? AS Array(String)), post_id) > 0,
    arrayElement(CAST(? AS Array(Array(String))), indexOf(CAST(? AS Array(String)), post_id)),
    media
  )
WHERE linkedin_id = ?
  AND has(CAST(? AS Array(String)), post_id)
`, tableName)

	if err := c.Conn.Exec(
		ctx,
		q,
		postIDs, images,
		postIDs, media, postIDs,
		linkedinID,
		postIDs,
	); err != nil {
		return 0, fmt.Errorf("Client.UpdateLinkedInPostURLs: update failed: %w", err)
	}

	return len(postIDs), nil
}

// GetGeoMappings retrieves geo ID to name mappings from ClickHouse cache.
// Returns a map of geo_id -> geo_name for the given IDs.
// IDs not found in the cache will not be present in the returned map.
func (c *Client) GetGeoMappings(ctx context.Context, geoIDs []string) (map[string]string, error) {
	if len(geoIDs) == 0 {
		return map[string]string{}, nil
	}

	// Build query with IN clause
	query := `SELECT geo_id, geo_name FROM linkedin_geo_mapping WHERE geo_id IN (?)`

	rows, err := c.Conn.Query(ctx, query, geoIDs)
	if err != nil {
		return nil, fmt.Errorf("Client.GetGeoMappings: failed to query geo mappings: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var geoID, geoName string
		if err := rows.Scan(&geoID, &geoName); err != nil {
			return nil, fmt.Errorf("Client.GetGeoMappings: failed to scan geo mapping row: %w", err)
		}
		result[geoID] = geoName
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("Client.GetGeoMappings: error iterating geo mapping rows: %w", err)
	}

	c.Logger.Debug().
		Int("requested", len(geoIDs)).
		Int("found", len(result)).
		Msg("Retrieved geo mappings from cache")

	return result, nil
}

// InsertGeoMappings inserts new geo ID to name mappings into ClickHouse cache.
// Uses ReplacingMergeTree so duplicates will be handled automatically.
func (c *Client) InsertGeoMappings(ctx context.Context, mappings map[string]string) error {
	if len(mappings) == 0 {
		return nil
	}

	batch, err := c.Conn.PrepareBatch(ctx, `INSERT INTO linkedin_geo_mapping (geo_id, geo_name, geo_type) VALUES`)
	if err != nil {
		return fmt.Errorf("Client.InsertGeoMappings: failed to prepare geo mapping batch: %w", err)
	}

	for geoID, geoName := range mappings {
		if err := batch.Append(geoID, geoName, ""); err != nil {
			return fmt.Errorf("Client.InsertGeoMappings: failed to append geo mapping: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.InsertGeoMappings: failed to send geo mapping batch: %w", err)
	}

	c.Logger.Info().
		Int("count", len(mappings)).
		Msg("Inserted geo mappings to cache")

	return nil
}

// GeoMappingWithType represents a geo mapping with type information.
type GeoMappingWithType struct {
	ID   string
	Name string
	Type string
}

// InsertGeoMappingsWithType inserts new geo ID to name mappings with type into ClickHouse cache.
// Uses ReplacingMergeTree so duplicates will be handled automatically.
func (c *Client) InsertGeoMappingsWithType(ctx context.Context, mappings []GeoMappingWithType) error {
	if len(mappings) == 0 {
		return nil
	}

	batch, err := c.Conn.PrepareBatch(ctx, `INSERT INTO linkedin_geo_mapping (geo_id, geo_name, geo_type) VALUES`)
	if err != nil {
		return fmt.Errorf("Client.InsertGeoMappingsWithType: failed to prepare geo mapping batch: %w", err)
	}

	for _, m := range mappings {
		if err := batch.Append(m.ID, m.Name, m.Type); err != nil {
			return fmt.Errorf("Client.InsertGeoMappingsWithType: failed to append geo mapping: %w", err)
		}
	}

	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.InsertGeoMappingsWithType: failed to send geo mapping batch: %w", err)
	}

	c.Logger.Info().
		Int("count", len(mappings)).
		Msg("Inserted geo mappings with type to cache")

	return nil
}
