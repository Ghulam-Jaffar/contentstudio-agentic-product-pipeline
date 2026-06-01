package clickhouse

import (
	"context"
	"fmt"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// BulkInsertPinterestUsers inserts Pinterest users into ClickHouse.
func (c *Client) BulkInsertPinterestUsers(ctx context.Context, users []clickhousemodels.PinterestUser) error {
	if len(users) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "pinterest_users").
		Int("batch_size", len(users)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO pinterest_users (
			user_id, profile_image, website_url, username,
			about, business_name, board_count, pin_count,
			account_type, follower_count, following_count,
			monthly_views, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestUsers: prepare batch: %w", err)
	}
	for _, u := range users {
		if err := batch.Append(
			u.UserID, u.ProfileImage, u.WebsiteURL, u.Username,
			u.About, u.BusinessName, int64(u.BoardCount), int64(u.PinCount),
			u.AccountType, u.FollowerCount, u.FollowingCount,
			u.MonthlyViews, u.InsertedAt,
		); err != nil {
			return fmt.Errorf("append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestUsers: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "pinterest_users").
		Int("batch_size", len(users)).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertPinterestBoards inserts Pinterest boards into ClickHouse.
func (c *Client) BulkInsertPinterestBoards(ctx context.Context, boards []clickhousemodels.PinterestBoard) error {
	if len(boards) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "pinterest_boards").
		Int("batch_size", len(boards)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO pinterest_boards (
			record_id, user_id, board_id, name, owner,
			description, privacy, image_cover_url,
			pin_thumbnail_urls, collaborator_count, pin_count,
			follower_count, created_at, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestBoards: prepare batch: %w", err)
	}
	for _, b := range boards {
		if err := batch.Append(
			b.RecordID, b.UserID, b.BoardID, b.Name, b.Owner,
			b.Description, b.Privacy, b.ImageCoverURL,
			b.PinThumbnailURLs, parseInt64(b.CollaboratorCount), parseInt64(b.PinCount),
			parseInt64(b.FollowerCount), b.CreatedAt, b.InsertedAt,
		); err != nil {
			return fmt.Errorf("append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestBoards: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "pinterest_boards").
		Int("batch_size", len(boards)).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertPinterestPins inserts Pinterest pins into ClickHouse.
func (c *Client) BulkInsertPinterestPins(ctx context.Context, pins []clickhousemodels.PinterestPin) error {
	if len(pins) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "pinterest_pins").
		Int("batch_size", len(pins)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO pinterest_pins (
			pin_id, user_id, board_id, title, note,
			parent_pin_id, board_section_id, description,
			board_owner, media_type, cover_image_url, video_url,
			duration, height, width, dominant_color, product_tags,
			creative_type, is_standard, is_owner, has_been_promoted,
			hour_of_day, day_of_week, created_at, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestPins: prepare batch: %w", err)
	}
	for _, p := range pins {
		if err := batch.Append(
			p.PinID, p.UserID, p.BoardID, p.Title, p.Note,
			p.ParentPinID, p.BoardSectionID, p.Description,
			p.BoardOwner, p.MediaType, p.CoverImageURL, p.VideoURL,
			p.Duration, p.Height, p.Width, p.DominantColor, p.ProductTags,
			p.CreativeType, parseInt64(p.IsStandard), parseInt64(p.IsOwner), parseInt64(p.HasBeenPromoted),
			int64(p.HourOfDay), p.DayOfWeek, p.CreatedAt, p.InsertedAt,
		); err != nil {
			return fmt.Errorf("append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestPins: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "pinterest_pins").
		Int("batch_size", len(pins)).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertPinterestPinInsights inserts Pinterest pin insights into ClickHouse.
func (c *Client) BulkInsertPinterestPinInsights(ctx context.Context, insights []clickhousemodels.PinterestPinInsight) error {
	if len(insights) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "pinterest_pin_insights").
		Int("batch_size", len(insights)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO pinterest_pin_insights (
			record_id, user_id, pin_id, board_id,
			pin_clicks, video_mrc_view, full_screen_play,
			outbound_click, video_v50_watch_time, clickthrough,
			clickthrough_rate, engagement, engagement_rate,
			video_start, profile_visit, closeup,
			full_screen_playtime, video_avg_watch_time,
			video_10s_view, quartile_95s_percent_view, user_follow,
			impression, saves, save_rate, data_status,
			day_of_week, hour_of_day, created_at, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestPinInsights: prepare batch: %w", err)
	}
	for _, i := range insights {
		if err := batch.Append(
			i.RecordID, i.UserID, i.PinID, i.BoardID,
			i.PinClicks, i.VideoMRCView, i.FullScreenPlay,
			i.OutboundClicks, i.VideoV50WatchTime, i.Clickthrough,
			int64(i.ClickthroughRate), i.Engagement, int64(i.EngagementRate),
			i.VideoStart, i.ProfileVisit, i.Closeup,
			i.FullScreenPlaytime, i.VideoAvgWatchTime,
			i.Video10sView, i.Quartile95sPercent, i.UserFollow,
			i.Impression, i.Saves, int64(i.SaveRate), i.DataStatus,
			i.DayOfWeek, int64(i.HourOfDay), i.Date, i.InsertedAt,
		); err != nil {
			return fmt.Errorf("append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestPinInsights: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "pinterest_pin_insights").
		Int("batch_size", len(insights)).
		Msg("Batch insert completed successfully")
	return nil
}

// BulkInsertPinterestUserInsights inserts Pinterest user insights into ClickHouse.
func (c *Client) BulkInsertPinterestUserInsights(ctx context.Context, insights []clickhousemodels.PinterestUserInsight) error {
	if len(insights) == 0 {
		return nil
	}
	c.Logger.Info().
		Str("table", "pinterest_user_insights").
		Int("batch_size", len(insights)).
		Msg("Starting batch insert")

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO pinterest_user_insights (
			record_id, user_id, pin_clicks, pin_click_rate,
			video_mrc_view, full_screen_play, outbound_click,
			video_v50_watch_time, clickthrough, clickthrough_rate,
			engagement, engagement_rate, video_start,
			profile_visit, closeup, full_screen_playtime,
			video_avg_watch_time, video_10s_view,
			quartile_95s_percent_view, impression, saves, save_rate,
			data_status, created_at, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestUserInsights: prepare batch: %w", err)
	}
	for _, i := range insights {
		if err := batch.Append(
			i.RecordID, i.UserID, i.PinClicks, int64(i.PinClickRate),
			i.VideoMRCView, i.FullScreenPlay, i.OutboundClicks,
			i.VideoV50WatchTime, i.Clickthrough, int64(i.ClickthroughRate),
			i.Engagement, int64(i.EngagementRate), i.VideoStart,
			i.ProfileVisit, i.Closeup, i.FullScreenPlaytime,
			i.VideoAvgWatchTime, i.Video10sView,
			i.Quartile95sPercent, i.Impression, i.Saves, int64(i.SaveRate),
			i.DataStatus, i.Date, i.InsertedAt,
		); err != nil {
			return fmt.Errorf("append: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertPinterestUserInsights: send batch: %w", err)
	}
	c.Logger.Info().
		Str("table", "pinterest_user_insights").
		Int("batch_size", len(insights)).
		Msg("Batch insert completed successfully")
	return nil
}

func parseInt64(s string) int64 {
	var v int64
	fmt.Sscanf(s, "%d", &v)
	return v
}
