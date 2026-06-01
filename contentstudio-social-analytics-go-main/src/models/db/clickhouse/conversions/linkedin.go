package conversions

import (
	"context"
	"time"

	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	parsed "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ConvertLinkedInPost converts parsed Kafka model to ClickHouse model.
func ConvertLinkedInPost(p *parsed.ParsedLinkedinPost) *chmodels.LinkedInPosts {
	if p == nil {
		return nil
	}
	return &chmodels.LinkedInPosts{
		LinkedinID:         p.LinkedinID,
		PostID:             p.PostID,
		Activity:           p.Activity,
		Comments:           p.Comments,
		TotalEngagement:    p.TotalEngagement,
		Favorites:          p.Favorites,
		PollData:           p.PollData,
		Reach:              p.Reach,
		Repost:             p.Repost,
		PostClicks:         p.PostClicks,
		Impressions:        p.Impressions,
		Title:              p.Title,
		Image:              p.Image,
		ArticleURL:         p.ArticleURL,
		ArticleTitle:       p.ArticleTitle,
		Media:              p.Media,
		MediaType:          p.MediaType,
		Type:               p.Type,
		Hashtags:           p.Hashtags,
		DayOfWeek:          p.DayOfWeek,
		HourOfDay:          int(p.HourOfDay),
		CreatedAt:          p.CreatedAt,
		PublishedAt:        p.PublishedAt,
		LastModifiedAt:     p.LastModifiedAt,
		LifecycleState:     p.LifecycleState,
		Visibility:         p.Visibility,
		SavingTime:         time.Now().UTC(),
		IsReshareDisabled:  p.IsReshareDisabled,
		FeedDistribution:   p.FeedDistribution,
		ThirdPartyChannels: p.ThirdPartyChannels,
	}
}

// ConvertLinkedInMediaAsset converts a parsed media asset to ClickHouse model.
func ConvertLinkedInMediaAsset(pa *parsed.ParsedLinkedinMediaAsset) *chmodels.LinkedInMediaAsset {
	if pa == nil {
		return nil
	}
	return &chmodels.LinkedInMediaAsset{
		ID:          pa.ID,
		DownloadURL: pa.DownloadURL,
		Thumbnail:   pa.Thumbnail,
		Type:        pa.Type,
		SavingTime:  time.Now().UTC(),
	}
}

// ConvertLinkedInStat converts a parsed stat to ClickHouse model.
func ConvertLinkedInStat(ps *parsed.ParsedLinkedinStat) *chmodels.LinkedInStat {
	if ps == nil {
		return nil
	}
	return &chmodels.LinkedInStat{
		ActivityID:             ps.ActivityID,
		CommentCount:           ps.CommentCount,
		LikeCount:              ps.LikeCount,
		UniqueImpressionsCount: ps.UniqueImpressionsCount,
		ShareCount:             ps.ShareCount,
		ClickCount:             ps.ClickCount,
		ImpressionCount:        ps.ImpressionCount,
		SavingTime:             time.Now().UTC(),
	}
}

// ConvertLinkedInInsights converts a parsed insights object to ClickHouse model.
func ConvertLinkedInInsights(pi *parsed.ParsedLinkedinInsights) *chmodels.LinkedInInsights {
	if pi == nil {
		return nil
	}
	return &chmodels.LinkedInInsights{
		LinkedinID:            pi.LinkedinID,
		OrganizationName:      pi.OrganizationName,
		RecordID:              pi.RecordID,
		ImpressionCount:       pi.ImpressionCount,
		OrganicFollowerCount:  pi.OrganicFollowerCount,
		TotalFollowerCount:    pi.TotalFollowerCount,
		PaidFollowerCount:     pi.PaidFollowerCount,
		DailyFollowerCount:    pi.DailyFollowerCount,
		Reach:                 pi.Reach,
		Repost:                pi.Repost,
		Comments:              pi.Comments,
		PostClicks:            pi.PostClicks,
		Reactions:             pi.Reactions,
		Engagement:            pi.Engagement,
		FollowersBySeniority:  pi.FollowersBySeniority,
		FollowersByIndustry:   pi.FollowersByIndustry,
		FollowersByCountry:    pi.FollowersByCountry,
		FollowersByCity:       pi.FollowersByCity,
		InsertedAt:            time.Now().UTC(),
		CreatedAt:             pi.CreatedAt,
		PageViews:             pi.PageViews,
		UniqueVisitors:        pi.UniqueVisitors,
		DesktopPageViews:      pi.DesktopPageViews,
		MobilePageViews:       pi.MobilePageViews,
		OverviewPageViews:     pi.OverviewPageViews,
		AboutPageViews:        pi.AboutPageViews,
		JobsPageViews:         pi.JobsPageViews,
		PeoplePageViews:       pi.PeoplePageViews,
		CareersPageViews:      pi.CareersPageViews,
		LifeAtPageViews:       pi.LifeAtPageViews,
		InsightsPageViews:     pi.InsightsPageViews,
		ProductsPageViews:     pi.ProductsPageViews,
		PageViewsByCountry:    pi.PageViewsByCountry,
		PageViewsByRegion:     pi.PageViewsByRegion,
		PageViewsByIndustry:   pi.PageViewsByIndustry,
		PageViewsBySeniority:  pi.PageViewsBySeniority,
		PageViewsByFunction:   pi.PageViewsByFunction,
		PageViewsByStaffCount: pi.PageViewsByStaffCount,
	}
}

// BulkInsertLinkedInPosts inserts posts batch via ClickHouse client.
func (s *ClickHouseSink) BulkInsertLinkedInPosts(ctx context.Context, posts []*chmodels.LinkedInPosts) error {
	if len(posts) == 0 {
		return nil
	}

	return s.ClickhouseClient.BulkInsertLinkedInPosts(ctx, posts)
}

// BulkInsertLinkedInMediaAssets inserts media assets batch via ClickHouse client.
func (s *ClickHouseSink) BulkInsertLinkedInMediaAssets(ctx context.Context, assets []*chmodels.LinkedInMediaAsset) error {
	if len(assets) == 0 {
		return nil
	}
	if inserter, ok := interface{}(s.ClickhouseClient).(interface {
		BulkInsertLinkedInMediaAssets(context.Context, []*chmodels.LinkedInMediaAsset) error
	}); ok {
		return inserter.BulkInsertLinkedInMediaAssets(ctx, assets)
	}
	return nil
}

// BulkInsertLinkedInStats inserts stats batch via ClickHouse client.
func (s *ClickHouseSink) BulkInsertLinkedInStats(ctx context.Context, stats []*chmodels.LinkedInStat) error {
	if len(stats) == 0 {
		return nil
	}
	if inserter, ok := interface{}(s.ClickhouseClient).(interface {
		BulkInsertLinkedInStats(context.Context, []*chmodels.LinkedInStat) error
	}); ok {
		return inserter.BulkInsertLinkedInStats(ctx, stats)
	}
	return nil
}

// BulkInsertLinkedInInsights inserts insights batch via ClickHouse client.
func (s *ClickHouseSink) BulkInsertLinkedInInsights(ctx context.Context, insights []*chmodels.LinkedInInsights) error {
	if len(insights) == 0 {
		return nil
	}

	return s.ClickhouseClient.BulkInsertLinkedInInsights(ctx, insights)
}
