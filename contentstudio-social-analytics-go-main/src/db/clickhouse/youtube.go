package clickhouse

import (
	"context"
	"fmt"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// BulkInsertYouTubeChannels inserts YouTube channels into ClickHouse in batches.
// Stores daily snapshots of channel data for historical tracking.
func (c *Client) BulkInsertYouTubeChannels(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error {
	if len(channels) == 0 {
		return nil
	}

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO youtube_channels (
			record_id, channel_id, title, description, custom_url,
			thumbnail_url, external_banner_url, country,
			subscriber_count, video_count, view_count,
			published_at, created_at, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeChannels: prepare batch youtube channels: %w", err)
	}
	for _, ch := range channels {
		if err := batch.Append(
			ch.RecordID, ch.ChannelID, ch.Title, ch.Description, ch.CustomURL,
			ch.ThumbnailURL, ch.ExternalBanner, ch.Country,
			ch.SubscriberCount, ch.VideoCount, ch.ViewCount,
			ch.PublishedAt, ch.CreatedAt, ch.InsertedAt,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertYouTubeChannels: append youtube channel: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeChannels: send youtube channels batch: %w", err)
	}
	return nil
}

// BulkInsertYouTubeVideos inserts YouTube videos into ClickHouse in batches.
// Stores daily snapshots of video data for historical tracking.
func (c *Client) BulkInsertYouTubeVideos(ctx context.Context, videos []*clickhousemodels.YouTubeVideo) error {
	if len(videos) == 0 {
		return nil
	}

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO youtube_videos (
			video_id, channel_id, title, description, duration,
			thumbnail_url, iframe_embed_html,
			likes, dislikes, views, comments, shares, favorites, saved,
			subscribers_gained, red_views, minutes_watched, red_minutes_watched,
			average_view_duration, average_view_percentage,
			impressions, impressions_click_through_rate,
			published_at, created_at, inserted_at, media_type
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeVideos: prepare batch youtube videos: %w", err)
	}
	for _, v := range videos {
		if err := batch.Append(
			v.VideoID, v.ChannelID, v.Title, v.Description, v.Duration,
			v.ThumbnailURL, v.IframeEmbedHTML,
			v.Likes, v.Dislikes, v.Views, v.Comments, v.Shares, v.Favorites, v.Saved,
			v.SubscribersGained, v.RedViews, v.MinutesWatched, v.RedMinutesWatched,
			v.AvgViewDuration, v.AvgViewPercentage,
			v.Impressions, v.ImpressionsClickThroughRate,
			v.PublishedAt, v.CreatedAt, v.InsertedAt, v.MediaType,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertYouTubeVideos: append youtube video: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeVideos: send youtube videos batch: %w", err)
	}
	return nil
}

// BulkInsertYouTubeActivityInsights inserts YouTube activity insights into ClickHouse in batches.
func (c *Client) BulkInsertYouTubeActivityInsights(ctx context.Context, insights []*clickhousemodels.YouTubeActivityInsights) error {
	if len(insights) == 0 {
		return nil
	}

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO youtube_activity_insights (
			record_id, channel_id, red_views, views, likes, dislikes,
			comments, shares, subscribers_gained,
			estimated_minutes_watched, estimated_red_minutes_watched,
			average_view_duration, average_view_percentage, created_at, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeActivityInsights: prepare batch youtube activity insights: %w", err)
	}
	for _, i := range insights {
		if err := batch.Append(
			i.RecordID, i.ChannelID, i.RedViews, i.Views, i.Likes, i.Dislikes,
			i.Comments, i.Shares, i.SubscribersGained,
			i.EstimatedMinutesWatched, i.EstimatedRedMinutesWatched,
			i.AvgViewDuration, i.AvgViewPercentage, i.CreatedAt, i.InsertedAt,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertYouTubeActivityInsights: append youtube activity insight: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeActivityInsights: send youtube activity insights batch: %w", err)
	}
	return nil
}

// BulkInsertYouTubeTrafficInsights inserts YouTube traffic insights into ClickHouse in batches.
func (c *Client) BulkInsertYouTubeTrafficInsights(ctx context.Context, insights []*clickhousemodels.YouTubeTrafficInsights) error {
	if len(insights) == 0 {
		return nil
	}

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO youtube_traffic_insights (
			record_id, channel_id, paid_views, annotation_views, end_screen_views,
			campaign_card_view, subscriber_views, no_link_other_views,
			yt_channel_views, yt_search_views, related_video_views, yt_other_page_views,
			ext_url_views, playlist_views, notification_views,
			subscriber_watch_time, non_subsciber_watch_time, created_at, shorts_views
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeTrafficInsights: prepare batch youtube traffic insights: %w", err)
	}
	for _, i := range insights {
		if err := batch.Append(
			i.RecordID, i.ChannelID, i.PaidViews, i.AnnotationViews, i.EndScreenViews,
			i.CampaignCardViews, i.SubscriberViews, i.NoLinkOtherViews,
			i.YTChannelViews, i.YTSearchViews, i.RelatedVideoViews, i.YTOtherPageViews,
			i.ExtURLViews, i.PlaylistViews, i.NotificationViews,
			i.SubscriberWatchTime, i.NonSubscriberWatchTime, i.CreatedAt, i.ShortsViews,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertYouTubeTrafficInsights: append youtube traffic insight: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeTrafficInsights: send youtube traffic insights batch: %w", err)
	}
	return nil
}

// BulkInsertYouTubeSharedInsights inserts YouTube shared insights into ClickHouse in batches.
func (c *Client) BulkInsertYouTubeSharedInsights(ctx context.Context, insights []*clickhousemodels.YouTubeSharedInsights) error {
	if len(insights) == 0 {
		return nil
	}

	batch, err := c.Conn.PrepareBatch(ctx, `
		INSERT INTO youtube_shared_insights (
			record_id, channel_id, ameba, blogger, copy_paste, cyworld, digg, dropbox,
			embed, mail, whats_app, other, facebook_messenger, facebook_pages, facebook,
			fotka, vkontakte, discord, google_plus, goo, hangouts, linkedin, pinterest,
			myspace, reddit, skype, telegram, twitter, tumblr, viber, weibo, wechat,
			youtube, inserted_at
		)
	`)
	if err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeSharedInsights: prepare batch youtube shared insights: %w", err)
	}
	for _, i := range insights {
		if err := batch.Append(
			i.RecordID, i.ChannelID, i.Ameba, i.Blogger, i.CopyPaste, i.Cyworld, i.Digg, i.Dropbox,
			i.Embed, i.Mail, i.WhatsApp, i.Other, i.FacebookMsgr, i.FacebookPages, i.Facebook,
			i.Fotka, i.VKontakte, i.Discord, i.GooglePlus, i.Goo, i.Hangouts, i.LinkedIn, i.Pinterest,
			i.Myspace, i.Reddit, i.Skype, i.Telegram, i.Twitter, i.Tumblr, i.Viber, i.Weibo, i.WeChat,
			i.YouTube, i.InsertedAt,
		); err != nil {
			return fmt.Errorf("Client.BulkInsertYouTubeSharedInsights: append youtube shared insight: %w", err)
		}
	}
	if err := batch.Send(); err != nil {
		return fmt.Errorf("Client.BulkInsertYouTubeSharedInsights: send youtube shared insights batch: %w", err)
	}
	return nil
}
