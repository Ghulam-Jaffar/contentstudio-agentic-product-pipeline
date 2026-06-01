package conversions

import (
	"context"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ConvertInstagramPost converts a ParsedInstagramPost to ClickHouse InstagramPost model
func (s *ClickHouseSink) ConvertInstagramPost(parsed *kafkamodels.ParsedInstagramPost) *clickhousemodels.InstagramPost {
	if parsed == nil {
		return nil
	}

	return &clickhousemodels.InstagramPost{
		InstagramID:         parsed.InstagramID,
		MediaID:             parsed.MediaID,
		Username:            parsed.Username,
		Name:                parsed.Name,
		ProfilePictureURL:   parsed.ProfilePictureURL,
		Permalink:           parsed.Permalink,
		LikeCount:           parsed.LikeCount,
		CommentsCount:       parsed.CommentsCount,
		Engagement:          parsed.Engagement,
		Impressions:         parsed.Impressions,
		Views:               parsed.Views,
		Reach:               parsed.Reach,
		Saved:               parsed.Saved,
		VideoViews:          parsed.VideoViews,
		Shares:              parsed.Shares,
		ReelsAvgWatchTime:   parsed.ReelsAvgWatchTime,
		ReelsTotalWatchTime: parsed.ReelsTotalWatchTime,
		Exits:               parsed.Exits,
		Replies:             parsed.Replies,
		TapsForward:         parsed.TapsForward,
		TapsBack:            parsed.TapsBack,
		ChildAssetsType:     parsed.ChildAssetsType,
		Caption:             parsed.Caption,
		MediaType:           parsed.MediaType,
		EntityType:          parsed.EntityType,
		MediaURL:            parsed.MediaURL,
		VideoURL:            parsed.VideoURL,
		Hashtags:            parsed.Hashtags,
		DayOfWeek:           parsed.DayOfWeek,
		HourOfDay:           parsed.HourOfDay,
		Year:                parsed.Year,
		Month:               parsed.Month,
		Timestamp:           parsed.Timestamp,
		StoredEventAt:       parsed.StoredEventAt,
		PostCreatedAt:       parsed.PostCreatedAt,
	}
}

// ConvertInstagramInsight converts a ParsedInstagramInsight to ClickHouse InstagramInsight model
func (s *ClickHouseSink) ConvertInstagramInsight(parsed *kafkamodels.ParsedInstagramInsight) *clickhousemodels.InstagramInsight {
	if parsed == nil {
		return nil
	}

	return &clickhousemodels.InstagramInsight{
		InstagramID:                   parsed.InstagramID,
		RecordID:                      parsed.RecordID,
		Name:                          parsed.Name,
		Username:                      parsed.Username,
		ProfilePictureURL:             parsed.ProfilePictureURL,
		FollowsCount:                  parsed.FollowsCount,
		FollowersCount:                parsed.FollowersCount,
		MediaCount:                    parsed.MediaCount,
		Tags:                          parsed.Tags,
		Impressions:                   parsed.Impressions,
		ProfileViews:                  parsed.ProfileViews,
		Shares:                        parsed.Shares,
		AccountsEngaged:               parsed.AccountsEngaged,
		Engagement:                    parsed.Engagement,
		Reach:                         parsed.Reach,
		Views:                         parsed.Views,
		Saves:                         parsed.Saves,
		Likes:                         parsed.Likes,
		Comments:                      parsed.Comments,
		AudienceAge:                   parsed.AudienceAge,
		AudienceGender:                parsed.AudienceGender,
		AudienceGenderAge:             parsed.AudienceGenderAge,
		AudienceLocale:                parsed.AudienceLocale,
		AudienceCity:                  parsed.AudienceCity,
		AudienceCountry:               parsed.AudienceCountry,
		AudienceAgeByEngagement:       parsed.AudienceAgeByEngagement,
		AudienceGenderByEngagement:    parsed.AudienceGenderByEngagement,
		AudienceGenderAgeByEngagement: parsed.AudienceGenderAgeByEngagement,
		AudienceCityByEngagement:      parsed.AudienceCityByEngagement,
		AudienceCountryByEngagement:   parsed.AudienceCountryByEngagement,
		AudienceAgeByReach:            parsed.AudienceAgeByReach,
		AudienceGenderByReach:         parsed.AudienceGenderByReach,
		AudienceGenderAgeByReach:      parsed.AudienceGenderAgeByReach,
		AudienceCityByReach:           parsed.AudienceCityByReach,
		AudienceCountryByReach:        parsed.AudienceCountryByReach,
		OnlineFollowers:               parsed.OnlineFollowers,
		AudienceDatetime:              parsed.AudienceDatetime,
		OnlineUsersDatetime:           parsed.OnlineUsersDatetime,
		DayOfWeek:                     parsed.DayOfWeek,
		Year:                          parsed.Year,
		Month:                         parsed.Month,
		CreatedTime:                   parsed.CreatedTime,
		UpdatedTime:                   parsed.UpdatedTime,
		Metadata:                      parsed.Metadata,
		StoredEventAt:                 parsed.StoredEventAt,
	}
}

// BulkInsertInstagramPosts inserts multiple Instagram posts into ClickHouse
func (s *ClickHouseSink) BulkInsertInstagramPosts(ctx context.Context, posts []*clickhousemodels.InstagramPost) error {
	if len(posts) == 0 {
		return nil
	}

	s.logger.Info().
		Int("count", len(posts)).
		Msg("Bulk inserting Instagram posts to ClickHouse")
	// This will be implemented in the ClickHouse client package
	return s.ClickhouseClient.BulkInsertInstagramPosts(ctx, posts)
}

// BulkInsertInstagramInsights inserts multiple Instagram insights into ClickHouse
func (s *ClickHouseSink) BulkInsertInstagramInsights(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error {
	if len(insights) == 0 {
		return nil
	}

	s.logger.Info().
		Int("count", len(insights)).
		Msg("Bulk inserting Instagram insights to ClickHouse")

	// This will be implemented in the ClickHouse client package
	return s.ClickhouseClient.BulkInsertInstagramInsights(ctx, insights)
}
