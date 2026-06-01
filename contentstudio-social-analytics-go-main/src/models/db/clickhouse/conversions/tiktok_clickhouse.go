package conversions

import (
	"context"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ConvertTikTokPost converts the parsed Kafka model to ClickHouse model
func ConvertTikTokPost(p *kafkamodels.ParsedTikTokPost) *chmodels.TikTokPosts {
	if p == nil {
		return nil
	}
	return &chmodels.TikTokPosts{

		TikTokID:        p.TikTokID, // May contain OpenID from MongoDB or TikTok account ID
		DisplayName:     p.DisplayName,
		ProfileLink:     p.ProfileLink,
		PostID:          p.ID, // Video ID from TikTok API
		CoverImageURL:   p.CoverImageURL,
		ShareURL:        p.ShareURL,
		PostDescription: p.PostDescription,
		Hashtags:        p.Hashtags,
		Duration:        p.Duration,
		Height:          p.Height,
		Width:           p.Width,
		Title:           p.Title,
		EmbedHTML:       p.EmbedHTML,
		EmbedLink:       p.EmbedLink,
		LikeCount:       p.LikeCount,
		CommentCount:    p.CommentCount,
		ShareCount:      p.ShareCount,
		ViewCount:       p.ViewCount,
		EngagementCount: p.EngagementCount,
		EngagementRate:  p.EngagementRate,
		CreatedAt:       time.Unix(p.CreateTime, 0),
		InsertedAt:      time.Now().UTC(),
	}
}

// BulkInsertTikTokPosts is a thin wrapper delegating to the ClickHouse client
func (s *ClickHouseSink) BulkInsertTikTokPosts(ctx context.Context, posts []*chmodels.TikTokPosts) error {
	return s.ClickhouseClient.BulkInsertTikTokPosts(ctx, posts)
}

// ConvertTikTokInsights converts parsed insights to ClickHouse model.
func ConvertTikTokInsights(p *kafkamodels.ParsedTikTokInsights) *chmodels.TikTokInsights {
	if p == nil {
		return nil
	}
	return &chmodels.TikTokInsights{
		RecordID:            p.RecordID,
		TikTokID:            p.TikTokID,
		DisplayName:         p.DisplayName,
		ProfileImage:        p.ProfileImage,
		TotalFollowerCount:  p.TotalFollowerCount,
		TotalFollowingCount: p.TotalFollowingCount,
		TotalLikeCount:      p.TotalLikeCount,
		TotalVideoCount:     p.TotalVideoCount,
		TotalVideoViews:     p.TotalVideoViews,
		TotalVideoLikes:     p.TotalVideoLikes,
		TotalVideoComments:  p.TotalVideoComments,
		TotalVideoShares:    p.TotalVideoShares,
		IsVerified:          p.IsVerified,
		Bio:                 p.Bio,
		ProfileLink:         p.ProfileLink,
		InsertedAt:          time.Unix(p.InsertedAt, 0),
	}
}

// BulkInsertTikTokInsights delegates to client.
func (s *ClickHouseSink) BulkInsertTikTokInsights(ctx context.Context, insights []*chmodels.TikTokInsights) error {
	return s.ClickhouseClient.BulkInsertTikTokInsights(ctx, insights)
}
