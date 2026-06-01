package clickhouse

import (
	"context"
	"fmt"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// BulkInsertGMBDailyMetrics inserts GMB daily metrics into ClickHouse in batches.
func (c *Client) BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error {
	if len(metrics) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "gmb_daily_metrics").
		Int("batch_size", len(metrics)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO gmb_daily_metrics (
			gmb_id, account_id, location_id, account_name, location_name,
			platform_name, inserted_at, created_at,
			business_impressions_desktop_maps,
			business_impressions_desktop_search,
			business_impressions_mobile_maps,
			business_impressions_mobile_search,
			call_clicks, website_clicks, business_direction_requests,
			business_conversations, business_bookings,
			business_food_orders, business_food_menu_clicks
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertGMBDailyMetrics: prepare batch: %w", err)
	}
	for _, m := range metrics {
		if err := batch.Append(
			m.GmbID, m.AccountID, m.LocationID, m.AccountName, m.LocationName,
			m.PlatformName, m.InsertedAt, m.CreatedAt,
			m.BusinessImpressionsDesktopMaps,
			m.BusinessImpressionsDesktopSearch,
			m.BusinessImpressionsMobileMaps,
			m.BusinessImpressionsMobileSearch,
			m.CallClicks, m.WebsiteClicks, m.BusinessDirectionRequests,
			m.BusinessConversations, m.BusinessBookings,
			m.BusinessFoodOrders, m.BusinessFoodMenuClicks,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertGMBDailyMetrics: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertGMBDailyMetrics: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "gmb_daily_metrics").
		Int("batch_size", len(metrics)).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertGMBMediaAssets inserts GMB media assets into ClickHouse in batches.
func (c *Client) BulkInsertGMBMediaAssets(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error {
	if len(assets) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "gmb_media_assets").
		Int("batch_size", len(assets)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO gmb_media_assets (
			gmb_id, account_id, location_id, account_name, location_name,
			platform_name, language_code, inserted_at, created_at,
			media_name, source_url, media_format, location_association_category,
			google_url, thumbnail_url, width_pixels, height_pixels
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertGMBMediaAssets: prepare batch: %w", err)
	}
	for _, a := range assets {
		if err := batch.Append(
			a.GmbID, a.AccountID, a.LocationID, a.AccountName, a.LocationName,
			a.PlatformName, a.LanguageCode, a.InsertedAt, a.CreatedAt,
			a.MediaName, a.SourceURL, a.MediaFormat, a.LocationAssociationCategory,
			a.GoogleURL, a.ThumbnailURL, a.WidthPixels, a.HeightPixels,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertGMBMediaAssets: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertGMBMediaAssets: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "gmb_media_assets").
		Int("batch_size", len(assets)).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertGMBSearchKeywordsMonthly inserts GMB search keywords into ClickHouse in batches.
func (c *Client) BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error {
	if len(keywords) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "gmb_search_keywords_monthly").
		Int("batch_size", len(keywords)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO gmb_search_keywords_monthly (
			gmb_id, account_id, location_id, account_name, location_name,
			platform_name, inserted_at, keyword_month,
			keyword, impressions_value, impressions_threshold
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertGMBSearchKeywordsMonthly: prepare batch: %w", err)
	}
	for _, k := range keywords {
		if err := batch.Append(
			k.GmbID, k.AccountID, k.LocationID, k.AccountName, k.LocationName,
			k.PlatformName, k.InsertedAt, k.KeywordMonth,
			k.Keyword, k.ImpressionsValue, k.ImpressionsThreshold,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertGMBSearchKeywordsMonthly: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertGMBSearchKeywordsMonthly: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "gmb_search_keywords_monthly").
		Int("batch_size", len(keywords)).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertGMBLocalPosts inserts GMB local posts into ClickHouse in batches.
func (c *Client) BulkInsertGMBLocalPosts(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error {
	if len(posts) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "gmb_local_posts").
		Int("batch_size", len(posts)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO gmb_local_posts (
			gmb_id, account_id, location_id, account_name, location_name,
			platform_name, language_code, inserted_at,
			created_at, updated_at, post_name, summary, state, topic_type,
			search_url, media_names, media_formats, media_google_urls
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertGMBLocalPosts: prepare batch: %w", err)
	}
	for _, p := range posts {
		if err := batch.Append(
			p.GmbID, p.AccountID, p.LocationID, p.AccountName, p.LocationName,
			p.PlatformName, p.LanguageCode, p.InsertedAt,
			p.CreatedAt, p.UpdatedAt, p.PostName, p.Summary, p.State, p.TopicType,
			p.SearchURL, p.MediaNames, p.MediaFormats, p.MediaGoogleURLs,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertGMBLocalPosts: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertGMBLocalPosts: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "gmb_local_posts").
		Int("batch_size", len(posts)).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertGMBReviews inserts GMB reviews into ClickHouse in batches.
func (c *Client) BulkInsertGMBReviews(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error {
	if len(reviews) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "gmb_reviews").
		Int("batch_size", len(reviews)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO gmb_reviews (
			gmb_id, account_id, location_id, account_name, location_name,
			platform_name, inserted_at, created_at, updated_at,
			review_id, review_name, reviewer_display_name, reviewer_profile_photo_url,
			star_rating, comment, reply_comment, reply_update_time
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertGMBReviews: prepare batch: %w", err)
	}
	for _, r := range reviews {
		if err := batch.Append(
			r.GmbID, r.AccountID, r.LocationID, r.AccountName, r.LocationName,
			r.PlatformName, r.InsertedAt, r.CreatedAt, r.UpdatedAt,
			r.ReviewID, r.ReviewName, r.ReviewerDisplayName, r.ReviewerProfilePhotoURL,
			r.StarRating, r.Comment, r.ReplyComment, r.ReplyUpdateTime,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertGMBReviews: append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertGMBReviews: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "gmb_reviews").
		Int("batch_size", len(reviews)).
		Msg("Batch insert completed successfully")
	return nil
}
