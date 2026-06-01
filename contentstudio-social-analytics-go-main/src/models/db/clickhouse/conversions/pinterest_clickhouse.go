package conversions

import (
	"context"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ConvertPinterestUser converts parsed Kafka model to ClickHouse model.
func ConvertPinterestUser(p *kafkamodels.ParsedPinterestUser) *chmodels.PinterestUser {
	if p == nil {
		return nil
	}
	return &chmodels.PinterestUser{
		UserID:         p.UserID,
		Username:       p.Username,
		About:          p.About,
		ProfileImage:   p.ProfileImage,
		WebsiteURL:     p.WebsiteURL,
		BusinessName:   p.BusinessName,
		BoardCount:     p.BoardCount,
		PinCount:       p.PinCount,
		AccountType:    p.AccountType,
		FollowerCount:  p.FollowerCount,
		FollowingCount: p.FollowingCount,
		MonthlyViews:   p.MonthlyViews,
		InsertedAt:     time.Now().UTC(),
	}
}

// BulkInsertPinterestUsers delegates to the ClickHouse client.
func (s *ClickHouseSink) BulkInsertPinterestUsers(ctx context.Context, users []chmodels.PinterestUser) error {
	return s.ClickhouseClient.BulkInsertPinterestUsers(ctx, users)
}

// ConvertPinterestBoard converts parsed Kafka model to ClickHouse model.
func ConvertPinterestBoard(p *kafkamodels.ParsedPinterestBoard) *chmodels.PinterestBoard {
	if p == nil {
		return nil
	}
	return &chmodels.PinterestBoard{
		RecordID:          p.RecordID,
		BoardID:           p.BoardID,
		UserID:            p.UserID,
		Name:              p.Name,
		Description:       p.Description,
		Privacy:           p.Privacy,
		PinCount:          p.PinCount,
		FollowerCount:     p.FollowerCount,
		CollaboratorCount: p.CollaboratorCount,
		Owner:             p.Owner,
		ImageCoverURL:     p.ImageCoverURL,
		PinThumbnailURLs:  p.PinThumbnailURLs,
		CreatedAt:         p.CreatedAt,
		InsertedAt:        time.Now().UTC(),
	}
}

// BulkInsertPinterestBoards delegates to the ClickHouse client.
func (s *ClickHouseSink) BulkInsertPinterestBoards(ctx context.Context, boards []chmodels.PinterestBoard) error {
	return s.ClickhouseClient.BulkInsertPinterestBoards(ctx, boards)
}

// ConvertPinterestPin converts parsed Kafka model to ClickHouse model.
func ConvertPinterestPin(p *kafkamodels.ParsedPinterestPin) *chmodels.PinterestPin {
	if p == nil {
		return nil
	}
	return &chmodels.PinterestPin{
		RecordID:        p.RecordID,
		PinID:           p.PinID,
		UserID:          p.UserID,
		BoardID:         p.BoardID,
		BoardSectionID:  p.BoardSectionID,
		ParentPinID:     p.ParentPinID,
		Title:           p.Title,
		Note:            p.Note,
		Description:     p.Description,
		Link:            p.Link,
		DominantColor:   p.DominantColor,
		CreativeType:    p.CreativeType,
		MediaType:       p.MediaType,
		CoverImageURL:   p.CoverImageURL,
		VideoURL:        p.VideoURL,
		Duration:        p.Duration,
		Height:          p.Height,
		Width:           p.Width,
		IsStandard:      p.IsStandard,
		IsOwner:         p.IsOwner,
		HasBeenPromoted: p.HasBeenPromoted,
		BoardOwner:      p.BoardOwner,
		ProductTags:     p.ProductTags,
		CreatedAt:       p.CreatedAt,
		DayOfWeek:       p.DayOfWeek,
		HourOfDay:       p.HourOfDay,
		InsertedAt:      time.Now().UTC(),
	}
}

// BulkInsertPinterestPins delegates to the ClickHouse client.
func (s *ClickHouseSink) BulkInsertPinterestPins(ctx context.Context, pins []chmodels.PinterestPin) error {
	return s.ClickhouseClient.BulkInsertPinterestPins(ctx, pins)
}

// ConvertPinterestPinInsight converts parsed Kafka model to ClickHouse model.
func ConvertPinterestPinInsight(p *kafkamodels.ParsedPinterestPinInsight) *chmodels.PinterestPinInsight {
	if p == nil {
		return nil
	}
	return &chmodels.PinterestPinInsight{
		RecordID:           p.RecordID,
		PinID:              p.PinID,
		UserID:             p.UserID,
		BoardID:            p.BoardID,
		Date:               p.Date,
		DataStatus:         p.DataStatus,
		Impression:         p.Impression,
		PinClicks:          p.PinClicks,
		OutboundClicks:     p.OutboundClicks,
		Saves:              p.Saves,
		SaveRate:           p.SaveRate,
		Clickthrough:       p.Clickthrough,
		ClickthroughRate:   p.ClickthroughRate,
		Engagement:         p.Engagement,
		EngagementRate:     p.EngagementRate,
		VideoMRCView:       p.VideoMRCView,
		VideoStart:         p.VideoStart,
		Video10sView:       p.Video10sView,
		VideoAvgWatchTime:  p.VideoAvgWatchTime,
		VideoV50WatchTime:  p.VideoV50WatchTime,
		FullScreenPlay:     p.FullScreenPlay,
		FullScreenPlaytime: p.FullScreenPlaytime,
		ProfileVisit:       p.ProfileVisit,
		Closeup:            p.Closeup,
		Quartile95sPercent: p.Quartile95sPercent,
		UserFollow:         p.UserFollow,
		DayOfWeek:          p.DayOfWeek,
		HourOfDay:          p.HourOfDay,
		InsertedAt:         time.Now().UTC(),
	}
}

// BulkInsertPinterestPinInsights delegates to the ClickHouse client.
func (s *ClickHouseSink) BulkInsertPinterestPinInsights(ctx context.Context, insights []chmodels.PinterestPinInsight) error {
	return s.ClickhouseClient.BulkInsertPinterestPinInsights(ctx, insights)
}

// ConvertPinterestUserInsight converts parsed Kafka model to ClickHouse model.
func ConvertPinterestUserInsight(p *kafkamodels.ParsedPinterestUserInsight) *chmodels.PinterestUserInsight {
	if p == nil {
		return nil
	}
	return &chmodels.PinterestUserInsight{
		RecordID:           p.RecordID,
		UserID:             p.UserID,
		Date:               p.Date,
		DataStatus:         p.DataStatus,
		Impression:         p.Impression,
		PinClicks:          p.PinClicks,
		PinClickRate:       p.PinClickRate,
		OutboundClicks:     p.OutboundClicks,
		Saves:              p.Saves,
		SaveRate:           p.SaveRate,
		Clickthrough:       p.Clickthrough,
		ClickthroughRate:   p.ClickthroughRate,
		Engagement:         p.Engagement,
		EngagementRate:     p.EngagementRate,
		VideoMRCView:       p.VideoMRCView,
		VideoStart:         p.VideoStart,
		Video10sView:       p.Video10sView,
		VideoAvgWatchTime:  p.VideoAvgWatchTime,
		VideoV50WatchTime:  p.VideoV50WatchTime,
		FullScreenPlay:     p.FullScreenPlay,
		FullScreenPlaytime: p.FullScreenPlaytime,
		ProfileVisit:       p.ProfileVisit,
		Closeup:            p.Closeup,
		Quartile95sPercent: p.Quartile95sPercent,
		InsertedAt:         time.Now().UTC(),
	}
}

// BulkInsertPinterestUserInsights delegates to the ClickHouse client.
func (s *ClickHouseSink) BulkInsertPinterestUserInsights(ctx context.Context, insights []chmodels.PinterestUserInsight) error {
	return s.ClickhouseClient.BulkInsertPinterestUserInsights(ctx, insights)
}
