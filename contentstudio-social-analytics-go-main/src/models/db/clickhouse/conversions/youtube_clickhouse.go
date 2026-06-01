package conversions

import (
	"context"
	"crypto/md5"
	"fmt"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	parsed "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ConvertYouTubeChannel converts parsed Kafka model to ClickHouse model.
func ConvertYouTubeChannel(p *parsed.ParsedYouTubeChannel) *chmodels.YouTubeChannel {
	if p == nil {
		return nil
	}
	return &chmodels.YouTubeChannel{
		RecordID:        p.RecordID,
		ChannelID:       p.ChannelID,
		Title:           p.Title,
		Description:     p.Description,
		CustomURL:       p.CustomURL,
		ThumbnailURL:    p.ThumbnailURL,
		ExternalBanner:  p.BannerURL,
		Country:         p.Country,
		SubscriberCount: p.SubscriberCount,
		VideoCount:      p.VideoCount,
		ViewCount:       p.ViewCount,
		PublishedAt:     p.PublishedAt,
		CreatedAt:       p.CreatedAt,
		InsertedAt:      p.InsertedAt,
	}
}

// ConvertYouTubeVideo converts parsed Kafka model to ClickHouse model.
func ConvertYouTubeVideo(p *parsed.ParsedYouTubeVideo) *chmodels.YouTubeVideo {
	if p == nil {
		return nil
	}
	return &chmodels.YouTubeVideo{
		VideoID:                     p.VideoID,
		ChannelID:                   p.ChannelID,
		Title:                       p.Title,
		Description:                 p.Description,
		Duration:                    p.Duration,
		ThumbnailURL:                p.ThumbnailURL,
		IframeEmbedHTML:             p.IframeEmbedHTML,
		Likes:                       p.Likes,
		Dislikes:                    p.Dislikes,
		Views:                       p.Views,
		Comments:                    p.Comments,
		Shares:                      p.Shares,
		Favorites:                   p.Favorites,
		Saved:                       p.Saved,
		SubscribersGained:           p.SubscribersGained,
		RedViews:                    p.RedViews,
		MinutesWatched:              p.MinutesWatched,
		RedMinutesWatched:           p.RedMinutesWatched,
		AvgViewDuration:             p.AvgViewDuration,
		AvgViewPercentage:           p.AvgViewPercentage,
		Impressions:                 p.Impressions,
		ImpressionsClickThroughRate: p.ImpressionsClickThroughRate,
		PublishedAt:                 p.PublishedAt,
		CreatedAt:                   p.CreatedAt,
		InsertedAt:                  p.InsertedAt,
		MediaType:                   p.MediaType,
	}
}

// ConvertYouTubeActivityInsights converts parsed Kafka model to ClickHouse model.
func ConvertYouTubeActivityInsights(p *parsed.ParsedYouTubeActivityInsights) *chmodels.YouTubeActivityInsights {
	if p == nil {
		return nil
	}
	return &chmodels.YouTubeActivityInsights{
		RecordID:                   p.RecordID,
		ChannelID:                  p.ChannelID,
		RedViews:                   p.RedViews,
		Views:                      p.Views,
		Likes:                      p.Likes,
		Dislikes:                   p.Dislikes,
		Comments:                   p.Comments,
		Shares:                     p.Shares,
		SubscribersGained:          p.SubscribersGained,
		EstimatedMinutesWatched:    p.EstimatedMinutesWatched,
		EstimatedRedMinutesWatched: p.EstimatedRedMinutesWatched,
		AvgViewDuration:            p.AvgViewDuration,
		AvgViewPercentage:          p.AvgViewPercentage,
		CreatedAt:                  p.CreatedAt,
		InsertedAt:                 time.Now().UTC(),
	}
}

// ConvertYouTubeTrafficInsights converts parsed Kafka model to ClickHouse model.
func ConvertYouTubeTrafficInsights(p *parsed.ParsedYouTubeTrafficInsights) *chmodels.YouTubeTrafficInsights {
	if p == nil {
		return nil
	}
	return &chmodels.YouTubeTrafficInsights{
		RecordID:               p.RecordID,
		ChannelID:              p.ChannelID,
		PaidViews:              p.PaidViews,
		AnnotationViews:        p.AnnotationViews,
		EndScreenViews:         p.EndScreenViews,
		CampaignCardViews:      p.CampaignCardViews,
		SubscriberViews:        p.SubscriberViews,
		NoLinkOtherViews:       p.NoLinkOtherViews,
		YTChannelViews:         p.YTChannelViews,
		YTSearchViews:          p.YTSearchViews,
		RelatedVideoViews:      p.RelatedVideoViews,
		YTOtherPageViews:       p.YTOtherPageViews,
		ExtURLViews:            p.ExtURLViews,
		PlaylistViews:          p.PlaylistViews,
		NotificationViews:      p.NotificationViews,
		SubscriberWatchTime:    p.SubscriberWatchTime,
		NonSubscriberWatchTime: p.NonSubscriberWatchTime,
		CreatedAt:              p.CreatedAt,
		ShortsViews:            p.ShortsViews,
	}
}

// ConvertYouTubeSharedInsights converts parsed Kafka model to ClickHouse model.
func ConvertYouTubeSharedInsights(p *parsed.ParsedYouTubeSharedInsights) *chmodels.YouTubeSharedInsights {
	if p == nil {
		return nil
	}
	return &chmodels.YouTubeSharedInsights{
		RecordID:      p.RecordID,
		ChannelID:     p.ChannelID,
		Ameba:         p.Ameba,
		Blogger:       p.Blogger,
		CopyPaste:     p.CopyPaste,
		Cyworld:       p.Cyworld,
		Digg:          p.Digg,
		Dropbox:       p.Dropbox,
		Embed:         p.Embed,
		Mail:          p.Mail,
		WhatsApp:      p.WhatsApp,
		Other:         p.Other,
		FacebookMsgr:  p.FacebookMsgr,
		FacebookPages: p.FacebookPages,
		Facebook:      p.Facebook,
		Fotka:         p.Fotka,
		VKontakte:     p.VKontakte,
		Discord:       p.Discord,
		GooglePlus:    p.GooglePlus,
		Goo:           p.Goo,
		Hangouts:      p.Hangouts,
		LinkedIn:      p.LinkedIn,
		Pinterest:     p.Pinterest,
		Myspace:       p.Myspace,
		Reddit:        p.Reddit,
		Skype:         p.Skype,
		Telegram:      p.Telegram,
		Twitter:       p.Twitter,
		Tumblr:        p.Tumblr,
		Viber:         p.Viber,
		Weibo:         p.Weibo,
		WeChat:        p.WeChat,
		YouTube:       p.YouTube,
		InsertedAt:    p.InsertedAt,
	}
}

// GenerateYouTubeRecordID generates a unique record ID using MD5 hash.
// Used for channels, activity insights, traffic insights, and shared insights.
func GenerateYouTubeRecordID(channelID string, date time.Time) string {
	data := fmt.Sprintf("%s_%s", channelID, date.Format("2006-01-02"))
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// ConvertRawToChannel converts raw channel data to parsed format.
// Creates a daily snapshot with created_at set to today's date.
func ConvertRawToChannel(raw *parsed.RawYouTubeChannel) *parsed.ParsedYouTubeChannel {
	if raw == nil {
		return nil
	}
	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	return &parsed.ParsedYouTubeChannel{
		RecordID:        GenerateYouTubeRecordID(raw.ChannelID, today),
		ChannelID:       raw.ChannelID,
		Title:           raw.Title,
		Description:     raw.Description,
		CustomURL:       raw.CustomURL,
		ThumbnailURL:    raw.ThumbnailURL,
		BannerURL:       raw.BannerURL,
		Country:         raw.Country,
		SubscriberCount: raw.SubscriberCount,
		VideoCount:      raw.VideoCount,
		ViewCount:       raw.ViewCount,
		PublishedAt:     raw.PublishedAt,
		CreatedAt:       today,
		InsertedAt:      now,
	}
}

// ConvertRawToVideo converts raw video data to parsed format.
// Uses AnalyticsDate as created_at if set, otherwise falls back to today's date.
func ConvertRawToVideo(raw *parsed.RawYouTubeVideo) *parsed.ParsedYouTubeVideo {
	if raw == nil {
		return nil
	}
	now := time.Now().UTC()
	// Use AnalyticsDate if set, otherwise fall back to today
	createdAt := raw.AnalyticsDate
	if createdAt.IsZero() {
		createdAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}
	return &parsed.ParsedYouTubeVideo{
		VideoID:                     raw.VideoID,
		ChannelID:                   raw.ChannelID,
		Title:                       raw.Title,
		Description:                 raw.Description,
		Duration:                    raw.Duration,
		ThumbnailURL:                raw.ThumbnailURL,
		IframeEmbedHTML:             raw.IframeEmbedHTML,
		MediaType:                   raw.MediaType,
		Likes:                       raw.Likes,
		Dislikes:                    raw.Dislikes,
		Views:                       raw.Views,
		Comments:                    raw.Comments,
		Shares:                      raw.Shares,
		Favorites:                   raw.Favorites,
		Saved:                       raw.Saved,
		SubscribersGained:           raw.SubscribersGained,
		RedViews:                    raw.RedViews,
		MinutesWatched:              raw.MinutesWatched,
		RedMinutesWatched:           raw.RedMinutesWatched,
		AvgViewDuration:             raw.AvgViewDuration,
		AvgViewPercentage:           raw.AvgViewPercentage,
		Impressions:                 raw.Impressions,
		ImpressionsClickThroughRate: raw.ImpressionsClickThroughRate,
		PublishedAt:                 raw.PublishedAt,
		CreatedAt:                   createdAt,
		InsertedAt:                  now,
	}
}

// ConvertRawToActivityInsights converts raw activity insights to parsed format.
func ConvertRawToActivityInsights(raw *parsed.RawYouTubeActivityInsights) *parsed.ParsedYouTubeActivityInsights {
	if raw == nil {
		return nil
	}
	return &parsed.ParsedYouTubeActivityInsights{
		RecordID:                   GenerateYouTubeRecordID(raw.ChannelID, raw.Date),
		ChannelID:                  raw.ChannelID,
		Views:                      raw.Views,
		RedViews:                   raw.RedViews,
		Likes:                      raw.Likes,
		Dislikes:                   raw.Dislikes,
		Comments:                   raw.Comments,
		Shares:                     raw.Shares,
		SubscribersGained:          raw.SubscribersGained,
		EstimatedMinutesWatched:    raw.EstimatedMinutesWatched,
		EstimatedRedMinutesWatched: raw.EstimatedRedMinutesWatched,
		AvgViewDuration:            raw.AvgViewDuration,
		AvgViewPercentage:          raw.AvgViewPercentage,
		CreatedAt:                  raw.Date,
	}
}

// AggregateTrafficInsights aggregates raw traffic insights by date into parsed format.
// Groups multiple traffic source records for the same channel and date.
func AggregateTrafficInsights(channelID string, date time.Time, rawInsights []*parsed.RawYouTubeTrafficInsights) *parsed.ParsedYouTubeTrafficInsights {
	if len(rawInsights) == 0 {
		return nil
	}

	result := &parsed.ParsedYouTubeTrafficInsights{
		RecordID:  GenerateYouTubeRecordID(channelID, date),
		ChannelID: channelID,
		CreatedAt: date,
	}

	for _, raw := range rawInsights {
		switch raw.TrafficSource {
		case parsed.TrafficSourcePaid:
			result.PaidViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceAnnotation:
			result.AnnotationViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceEndScreen:
			result.EndScreenViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceCampaignCard:
			result.CampaignCardViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceSubscriber:
			result.SubscriberViews = raw.Views
			result.SubscriberWatchTime = raw.MinutesWatched
		case parsed.TrafficSourceNoLinkOther:
			result.NoLinkOtherViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceYTChannel:
			result.YTChannelViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceYTSearch:
			result.YTSearchViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceRelatedVideo:
			result.RelatedVideoViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceYTOtherPage:
			result.YTOtherPageViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceExtURL:
			result.ExtURLViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourcePlaylist:
			result.PlaylistViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceNotification:
			result.NotificationViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		case parsed.TrafficSourceShorts:
			result.ShortsViews = raw.Views
			result.NonSubscriberWatchTime += raw.MinutesWatched
		}
	}

	return result
}

// AggregateSharedInsights aggregates raw shared insights into parsed format.
// Groups multiple sharing service records for the same channel.
func AggregateSharedInsights(channelID string, date time.Time, rawInsights []*parsed.RawYouTubeSharedInsights) *parsed.ParsedYouTubeSharedInsights {
	if len(rawInsights) == 0 {
		return nil
	}

	result := &parsed.ParsedYouTubeSharedInsights{
		RecordID:   GenerateYouTubeRecordID(channelID, date),
		ChannelID:  channelID,
		InsertedAt: time.Now().UTC(),
	}

	for _, raw := range rawInsights {
		switch raw.SharingService {
		case parsed.SharingServiceAmeba:
			result.Ameba = raw.Shares
		case parsed.SharingServiceBlogger:
			result.Blogger = raw.Shares
		case parsed.SharingServiceCopyPaste:
			result.CopyPaste = raw.Shares
		case parsed.SharingServiceCyworld:
			result.Cyworld = raw.Shares
		case parsed.SharingServiceDigg:
			result.Digg = raw.Shares
		case parsed.SharingServiceDropbox:
			result.Dropbox = raw.Shares
		case parsed.SharingServiceEmbed:
			result.Embed = raw.Shares
		case parsed.SharingServiceMail:
			result.Mail = raw.Shares
		case parsed.SharingServiceWhatsApp:
			result.WhatsApp = raw.Shares
		case parsed.SharingServiceOther:
			result.Other = raw.Shares
		case parsed.SharingServiceFacebookMsgr:
			result.FacebookMsgr = raw.Shares
		case parsed.SharingServiceFacebookPages:
			result.FacebookPages = raw.Shares
		case parsed.SharingServiceFacebook:
			result.Facebook = raw.Shares
		case parsed.SharingServiceFotka:
			result.Fotka = raw.Shares
		case parsed.SharingServiceVKontakte:
			result.VKontakte = raw.Shares
		case parsed.SharingServiceDiscord:
			result.Discord = raw.Shares
		case parsed.SharingServiceGooglePlus:
			result.GooglePlus = raw.Shares
		case parsed.SharingServiceGoo:
			result.Goo = raw.Shares
		case parsed.SharingServiceHangouts:
			result.Hangouts = raw.Shares
		case parsed.SharingServiceLinkedIn:
			result.LinkedIn = raw.Shares
		case parsed.SharingServicePinterest:
			result.Pinterest = raw.Shares
		case parsed.SharingServiceMyspace:
			result.Myspace = raw.Shares
		case parsed.SharingServiceReddit:
			result.Reddit = raw.Shares
		case parsed.SharingServiceSkype:
			result.Skype = raw.Shares
		case parsed.SharingServiceTelegram:
			result.Telegram = raw.Shares
		case parsed.SharingServiceTwitter:
			result.Twitter = raw.Shares
		case parsed.SharingServiceTumblr:
			result.Tumblr = raw.Shares
		case parsed.SharingServiceViber:
			result.Viber = raw.Shares
		case parsed.SharingServiceWeibo:
			result.Weibo = raw.Shares
		case parsed.SharingServiceWeChat:
			result.WeChat = raw.Shares
		case parsed.SharingServiceYouTube,
			parsed.SharingServiceYouTubeGaming,
			parsed.SharingServiceYouTubeKids,
			parsed.SharingServiceYouTubeMusic,
			parsed.SharingServiceYouTubeTV:
			// Consolidate all YouTube variants into single field
			result.YouTube += raw.Shares
		}
	}

	return result
}

// BulkInsertYouTubeChannels inserts channels batch via ClickHouse client.
func (s *ClickHouseSink) BulkInsertYouTubeChannels(ctx context.Context, channels []*chmodels.YouTubeChannel) error {
	if len(channels) == 0 {
		return nil
	}
	return s.ClickhouseClient.BulkInsertYouTubeChannels(ctx, channels)
}

// BulkInsertYouTubeVideos inserts videos batch via ClickHouse client.
func (s *ClickHouseSink) BulkInsertYouTubeVideos(ctx context.Context, videos []*chmodels.YouTubeVideo) error {
	if len(videos) == 0 {
		return nil
	}
	return s.ClickhouseClient.BulkInsertYouTubeVideos(ctx, videos)
}

// BulkInsertYouTubeActivityInsights inserts activity insights batch via ClickHouse client.
func (s *ClickHouseSink) BulkInsertYouTubeActivityInsights(ctx context.Context, insights []*chmodels.YouTubeActivityInsights) error {
	if len(insights) == 0 {
		return nil
	}
	return s.ClickhouseClient.BulkInsertYouTubeActivityInsights(ctx, insights)
}

// BulkInsertYouTubeTrafficInsights inserts traffic insights batch via ClickHouse client.
func (s *ClickHouseSink) BulkInsertYouTubeTrafficInsights(ctx context.Context, insights []*chmodels.YouTubeTrafficInsights) error {
	if len(insights) == 0 {
		return nil
	}
	return s.ClickhouseClient.BulkInsertYouTubeTrafficInsights(ctx, insights)
}

// BulkInsertYouTubeSharedInsights inserts shared insights batch via ClickHouse client.
func (s *ClickHouseSink) BulkInsertYouTubeSharedInsights(ctx context.Context, insights []*chmodels.YouTubeSharedInsights) error {
	if len(insights) == 0 {
		return nil
	}
	return s.ClickhouseClient.BulkInsertYouTubeSharedInsights(ctx, insights)
}
