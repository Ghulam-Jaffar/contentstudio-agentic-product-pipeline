package conversions

import (
	"context"
	"fmt"
	"strings"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ConvertFacebookPost converts a ParsedFacebookPost to ClickHouse FacebookPosts model
func (s *ClickHouseSink) ConvertFacebookPost(parsed *kafkamodels.ParsedFacebookPost) *clickhousemodels.FacebookPosts {
	if parsed == nil {
		return nil
	}

	return &clickhousemodels.FacebookPosts{
		PageName:                     parsed.PageName,
		PageID:                       parsed.PageID,
		MediaType:                    parsed.MediaType,
		PostID:                       parsed.PostID,
		Permalink:                    parsed.Permalink,
		StatusType:                   parsed.StatusType,
		VideoID:                      parsed.VideoID,
		Category:                     parsed.Category,
		PublishedBy:                  parsed.PublishedBy,
		PublishedByURL:               parsed.PublishedByURL,
		SharedFromName:               parsed.SharedFromName,
		SharedFromID:                 parsed.SharedFromID,
		SharedFromLink:               parsed.SharedFromLink,
		Like:                         parsed.Like,
		Love:                         parsed.Love,
		Haha:                         parsed.Haha,
		Wow:                          parsed.Wow,
		Sad:                          parsed.Sad,
		Angry:                        parsed.Angry,
		Thankful:                     parsed.Thankful,
		Total:                        parsed.Total,
		Shares:                       parsed.Shares,
		Comments:                     parsed.Comments,
		PostClicks:                   parsed.PostClicks,
		TotalEngagement:              parsed.TotalEngagement,
		PostEngagedUsers:             parsed.PostEngagedUsers,
		DayOfWeek:                    parsed.DayOfWeek,
		HourOfDay:                    parsed.HourOfDay,
		CreatedTime:                  parsed.CreatedTime,
		UpdatedTime:                  parsed.UpdatedTime,
		SavingTime:                   parsed.SavingTime,
		MessageTags:                  parsed.MessageTags,
		PostMetadata:                 parsed.PostMetadata,
		Caption:                      parsed.Caption,
		Description:                  parsed.Description,
		FullPicture:                  parsed.FullPicture,
		Link:                         parsed.Link,
		PostImpressions:              parsed.PostImpressions,
		PostImpressionsUnique:        parsed.PostImpressionsUnique,
		PostImpressionsPaid:          parsed.PostImpressionsPaid,
		PostImpressionsPaidUnique:    parsed.PostImpressionsPaidUnique,
		PostImpressionsOrganic:       parsed.PostImpressionsOrganic,
		PostImpressionsOrganicUnique: parsed.PostImpressionsOrganicUnique,
		PostImpressionsViral:         parsed.PostImpressionsViral,
		PostImpressionsViralUnique:   parsed.PostImpressionsViralUnique,
		PostVideoViews:               parsed.PostVideoViews,
		TotalImpressions:             parsed.TotalImpressions,
	}
}

// ConvertFacebookMediaAssets converts a ParsedFacebookMediaAsset to ClickHouse FacebookMediaAssets model
func (s *ClickHouseSink) ConvertFacebookMediaAssets(parsed *kafkamodels.ParsedFacebookMediaAsset) *clickhousemodels.FacebookMediaAssets {
	if parsed == nil {
		return nil
	}

	return &clickhousemodels.FacebookMediaAssets{
		PageID:       parsed.PageID,
		MediaID:      parsed.MediaID,
		PostID:       parsed.PostID,
		AssetType:    parsed.AssetType,
		Link:         parsed.Link,
		CallToAction: parsed.CallToAction,
		CTAType:      parsed.CTAType,
		Caption:      parsed.Caption,
		Description:  parsed.Description,
		CreatedAt:    parsed.CreatedAt,
		InsertedAt:   parsed.InsertedAt,
	}
}

// ConvertFacebookVideoInsights converts a ParsedFacebookVideoInsights to ClickHouse FacebookVideoInsights model
func (s *ClickHouseSink) ConvertFacebookVideoInsights(parsed *kafkamodels.ParsedFacebookVideoInsights) *clickhousemodels.FacebookVideoInsights {
	if parsed == nil {
		return nil
	}

	return &clickhousemodels.FacebookVideoInsights{
		PostID:      parsed.PostID,
		PageID:      parsed.PageID,
		VideoID:     parsed.VideoID,
		CreatedTime: parsed.CreatedTime,
		UpdatedTime: parsed.UpdatedTime,

		TotalVideoFollowers:                             parsed.TotalVideoFollowers,
		TotalVideoViews:                                 parsed.TotalVideoViews,
		TotalVideoViewsUnique:                           parsed.TotalVideoViewsUnique,
		TotalVideoViewsAutoplayed:                       parsed.TotalVideoViewsAutoplayed,
		TotalVideoViewsOrganic:                          parsed.TotalVideoViewsOrganic,
		TotalVideoViewsOrganicUnique:                    parsed.TotalVideoViewsOrganicUnique,
		TotalVideoViewsPaid:                             parsed.TotalVideoViewsPaid,
		TotalVideoViewsPaidUnique:                       parsed.TotalVideoViewsPaidUnique,
		TotalVideoViewsSoundOn:                          parsed.TotalVideoViewsSoundOn,
		TotalVideoViewsByDistributionType:               convertStringArrayFields(parsed.TotalVideoViewsByDistributionType),
		TotalVideoViewTimeByDistributionType:            convertStringArrayFields(parsed.TotalVideoViewTimeByDistributionType),
		TotalVideoViewTimeByCountryID:                   convertStringArrayFields(parsed.TotalVideoViewTimeByCountryID),
		TotalVideoViewTimeByRegionID:                    convertStringArrayFields(parsed.TotalVideoViewTimeByRegionID),
		TotalVideoViewTimeByAgeBucketAndGender:          convertStringArrayFields(parsed.TotalVideoViewTimeByAgeBucketAndGender),
		TotalVideoPlayCount:                             parsed.TotalVideoPlayCount,
		TotalVideoConsumptionRate:                       parsed.TotalVideoConsumptionRate,
		TotalVideoCompleteViews:                         parsed.TotalVideoCompleteViews,
		TotalVideoCompleteViewsUnique:                   parsed.TotalVideoCompleteViewsUnique,
		TotalVideoCompleteViewsAutoplayed:               parsed.TotalVideoCompleteViewsAutoplayed,
		TotalVideoCompleteViewsClickedToPlay:            parsed.TotalVideoCompleteViewsClickedToPlay,
		TotalVideoCompleteViewsOrganic:                  parsed.TotalVideoCompleteViewsOrganic,
		TotalVideoCompleteViewsOrganicUnique:            parsed.TotalVideoCompleteViewsOrganicUnique,
		TotalVideoCompleteViewsPaid:                     parsed.TotalVideoCompleteViewsPaid,
		TotalVideoCompleteViewsPaidUnique:               parsed.TotalVideoCompleteViewsPaidUnique,
		VideoAsset60sVideoViewTotalCountByIsMonetizable: convertStringArrayFields(parsed.VideoAsset60sVideoViewTotalCountByIsMonetizable),
		TotalVideo15minExcludesShorterViews:             parsed.TotalVideo15minExcludesShorterViews,
		TotalVideo15minExcludesShorterViewsUnique:       parsed.TotalVideo15minExcludesShorterViewsUnique,
		TotalVideo60sExcludesShorterViews:               parsed.TotalVideo60sExcludesShorterViews,
		TotalVideo30sViews:                              parsed.TotalVideo30sViews,
		TotalVideo30sViewsUnique:                        parsed.TotalVideo30sViewsUnique,
		TotalVideo30sViewsAutoplayed:                    parsed.TotalVideo30sViewsAutoplayed,
		TotalVideo30sViewsClickedToPlay:                 parsed.TotalVideo30sViewsClickedToPlay,
		TotalVideo30sViewsOrganic:                       parsed.TotalVideo30sViewsOrganic,
		TotalVideo30sViewsPaid:                          parsed.TotalVideo30sViewsPaid,
		TotalVideo30sViewsSoundOn:                       parsed.TotalVideo30sViewsSoundOn,
		TotalVideo10sViews:                              parsed.TotalVideo10sViews,
		TotalVideo10sViewsUnique:                        parsed.TotalVideo10sViewsUnique,
		TotalVideo10sViewsAutoplayed:                    parsed.TotalVideo10sViewsAutoplayed,
		TotalVideo10sViewsClickedToPlay:                 parsed.TotalVideo10sViewsClickedToPlay,
		TotalVideo10sViewsOrganic:                       parsed.TotalVideo10sViewsOrganic,
		TotalVideo10sViewsPaid:                          parsed.TotalVideo10sViewsPaid,
		TotalVideo10sViewsSoundOn:                       parsed.TotalVideo10sViewsSoundOn,
		TotalVideo15sViews:                              parsed.TotalVideo15sViews,
		TotalVideoAvgTimeWatched:                        parsed.TotalVideoAvgTimeWatched,
		TotalVideoViewTotalTime:                         parsed.TotalVideoViewTotalTime,
		TotalVideoViewTotalTimeOrganic:                  parsed.TotalVideoViewTotalTimeOrganic,
		TotalVideoViewTotalTimePaid:                     parsed.TotalVideoViewTotalTimePaid,
		TotalVideoRetentionGraphAutoplayed:              convertStringArrayFields(parsed.TotalVideoRetentionGraphAutoplayed),
		TotalVideoRetentionGraphClickedToPlay:           convertStringArrayFields(parsed.TotalVideoRetentionGraphClickedToPlay),
		TotalVideoRetentionGraphGenderMale:              convertStringArrayFields(parsed.TotalVideoRetentionGraphGenderMale),
		TotalVideoRetentionGraphGenderFemale:            convertStringArrayFields(parsed.TotalVideoRetentionGraphGenderFemale),
		TotalVideoImpressions:                           parsed.TotalVideoImpressions,
		TotalVideoImpressionsUnique:                     parsed.TotalVideoImpressionsUnique,
		TotalVideoImpressionsPaidUnique:                 parsed.TotalVideoImpressionsPaidUnique,
		TotalVideoImpressionsPaid:                       parsed.TotalVideoImpressionsPaid,
		TotalVideoImpressionsOrganicUnique:              parsed.TotalVideoImpressionsOrganicUnique,
		TotalVideoImpressionsOrganic:                    parsed.TotalVideoImpressionsOrganic,
		TotalVideoImpressionsViralUnique:                parsed.TotalVideoImpressionsViralUnique,
		TotalVideoImpressionsViral:                      parsed.TotalVideoImpressionsViral,
		TotalVideoImpressionsFanUnique:                  parsed.TotalVideoImpressionsFanUnique,
		TotalVideoImpressionsFan:                        parsed.TotalVideoImpressionsFan,
		TotalVideoImpressionsFanPaidUnique:              parsed.TotalVideoImpressionsFanPaidUnique,
		TotalVideoImpressionsFanPaid:                    parsed.TotalVideoImpressionsFanPaid,
		TotalVideoStoriesByActionType:                   convertStringArrayFields(parsed.TotalVideoStoriesByActionType),
		TotalVideoReactionsByTypeTotal:                  convertStringArrayFields(parsed.TotalVideoReactionsByTypeTotal),
		TotalEngagement:                                 parsed.TotalEngagement,
		TotalVideoAdBreakEarnings:                       parsed.TotalVideoAdBreakEarnings,
		TotalVideoAdBreakAdImpressions:                  parsed.TotalVideoAdBreakAdImpressions,
		TotalVideoAdBreakAdCPM:                          int64(parsed.TotalVideoAdBreakAdCPM), // Fix type conversion from float64 to int64
	}
}

// ConvertFacebookReelsInsights converts a ParsedFacebookReelsInsights to ClickHouse FacebookReelsInsights model
func (s *ClickHouseSink) ConvertFacebookReelsInsights(parsed *kafkamodels.ParsedFacebookReelsInsights) *clickhousemodels.FacebookReelsInsights {
	if parsed == nil {
		return nil
	}

	return &clickhousemodels.FacebookReelsInsights{
		PageID:               parsed.PageID,
		PostID:               parsed.PostID,
		AverageTimeWatched:   parsed.AverageTimeWatched,
		TotalTimeWatchedInMs: parsed.TotalTimeWatchedInMs,
		PlayCount:            parsed.PlayCount,
		ImpressionsUnique:    parsed.ImpressionsUnique,
		ReelFollowers:        parsed.ReelFollowers,
		CreatedAt:            parsed.CreatedAt,
		SavingTime:           parsed.SavingTime,
	}
}

// ConvertFacebookInsights converts a ParsedFacebookInsights to ClickHouse FacebookInsights model
func (s *ClickHouseSink) ConvertFacebookInsights(parsed *kafkamodels.ParsedFacebookInsights) *clickhousemodels.FacebookInsights {
	if parsed == nil {
		return nil
	}

	return &clickhousemodels.FacebookInsights{
		HashID:       parsed.HashID,
		PageID:       parsed.PageID,
		PageCategory: parsed.PageCategory,
		DayOfWeek:    parsed.DayOfWeek,
		Year:         int64(parsed.Year),
		Month:        int64(parsed.Month),
		CreatedTime:  parsed.CreatedTime,
		SavingTime:   parsed.SavingTime,

		// Fans info and demographics
		PageFans:          parsed.PageFans,
		PageFansCity:      convertStringArrayFields(parsed.PageFansCity),
		PageFansCountry:   convertStringArrayFields(parsed.PageFansCountry),
		PageFansLocale:    convertStringArrayFields(parsed.PageFansLocale),
		PageFansAge:       convertStringArrayFields(parsed.PageFansAge),
		PageFansGender:    convertStringArrayFields(parsed.PageFansGender),
		PageFansGenderAge: convertStringArrayFields(parsed.PageFansGenderAge),
		PageFollows:       parsed.PageFollows,
		PageViews:         parsed.PageViews,

		// Paid/unpaid fans and total
		PageFanAddsByPaidNonPaidUnique: convertStringArrayFields(parsed.PageFanAddsByPaidNonPaidUnique),

		// New added and removed fans
		PageFanAddsUnique:    parsed.PageFanAddsUnique,
		PageFanRemovesUnique: parsed.PageFanRemovesUnique,

		// Liked pages by source
		PageFansByLikeSourceUnique:   convertStringArrayFields(parsed.PageFansByLikeSourceUnique),
		PageFansByUnlikeSourceUnique: convertStringArrayFields(parsed.PageFansByUnlikeSourceUnique),

		// Total fans likes and unlikes
		PageFansByLike:   parsed.PageFansByLike,
		PageFansByUnlike: parsed.PageFansByUnlike,

		// Total clicks
		PageTotalActions: parsed.PageTotalActions,

		// Total engagements
		PagePostEngagements: parsed.PagePostEngagements,

		// Total impressions
		PageImpressions:       parsed.PageImpressions,
		PageImpressionsUnique: parsed.PageImpressionsUnique,

		// Impressions paid or organic
		PageImpressionsOrganic: parsed.PageImpressionsOrganic,
		PageImpressionsPaid:    parsed.PageImpressionsPaid,

		// Video views paid, total and organic
		PageVideoViewsPaid:    parsed.PageVideoViewsPaid,
		PageVideoViews:        parsed.PageVideoViews,
		PageVideoViewsOrganic: parsed.PageVideoViewsOrganic,

		// Metrics for video play
		PageVideoViewsAutoplayed:  parsed.PageVideoViewsAutoplayed,
		PageVideoViewsClickToPlay: parsed.PageVideoViewsClickToPlay,
		PageVideoRepeatViews:      parsed.PageVideoRepeatViews,

		// Total feedback count
		PageNegativeFeedback: parsed.PageNegativeFeedback,
		PagePositiveFeedback: parsed.PagePositiveFeedback,

		// Feedback types
		PageNegativeFeedbackByType: convertStringArrayFields(parsed.PageNegativeFeedbackByType),
		PagePositiveFeedbackByType: convertStringArrayFields(parsed.PagePositiveFeedbackByType),

		// Total fans online at what hour
		PageFansOnline: convertStringArrayFields(parsed.PageFansOnline),
		ActiveUsers:    parsed.ActiveUsers,

		// Sentiments towards the page
		PositiveSentiment: parsed.PositiveSentiment,
		NegativeSentiment: parsed.NegativeSentiment,

		// Posts count
		PostsCount:        parsed.PostsCount,
		LikesCount:        parsed.LikesCount,
		TalkingAboutCount: parsed.TalkingAboutCount,

		// Types of posts: links, images, videos
		TypeCount: convertStringArrayFields(parsed.TypeCount),

		// Posts sent and received
		MessageCount: convertStringArrayFields(parsed.MessageCount),

		// Prime time
		PrimeTime: parsed.PrimeTime,
	}
}

// BulkInsertPosts inserts multiple Facebook posts into ClickHouse
func (s *ClickHouseSink) BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
	if len(posts) == 0 {
		return nil
	}

	s.logger.Info().
		Int("count", len(posts)).
		Msg("Bulk inserting Facebook posts to ClickHouse")

	return s.ClickhouseClient.BulkInsertPosts(ctx, posts)
}

// BulkInsertMediaAssets inserts multiple Facebook media assets into ClickHouse
func (s *ClickHouseSink) BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
	if len(assets) == 0 {
		return nil
	}

	s.logger.Info().
		Int("count", len(assets)).
		Msg("Bulk inserting Facebook media assets to ClickHouse")

	return s.ClickhouseClient.BulkInsertMediaAssets(ctx, assets)
}

// BulkInsertVideoInsights inserts multiple Facebook video insights into ClickHouse
func (s *ClickHouseSink) BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
	if len(insights) == 0 {
		return nil
	}

	s.logger.Info().
		Int("count", len(insights)).
		Msg("Bulk inserting Facebook video insights to ClickHouse")

	return s.ClickhouseClient.BulkInsertVideoInsights(ctx, insights)
}

// BulkInsertReelsInsights inserts multiple Facebook reels insights into ClickHouse
func (s *ClickHouseSink) BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
	if len(insights) == 0 {
		return nil
	}

	s.logger.Info().
		Int("count", len(insights)).
		Msg("Bulk inserting Facebook reels insights to ClickHouse")

	return s.ClickhouseClient.BulkInsertReelsInsights(ctx, insights)
}

// BulkInsertInsights inserts multiple Facebook page insights into ClickHouse
func (s *ClickHouseSink) BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
	if len(insights) == 0 {
		return nil
	}

	s.logger.Info().
		Int("count", len(insights)).
		Msg("Bulk inserting Facebook page insights to ClickHouse")

	return s.ClickhouseClient.BulkInsertInsights(ctx, insights)
}

// ProcessParsedData processes a batch of parsed Facebook data and converts it to ClickHouse models
func (s *ClickHouseSink) ProcessParsedData(ctx context.Context, data interface{}) error {
	switch parsedData := data.(type) {
	case *kafkamodels.ParsedFacebookPost:
		clickhousePost := s.ConvertFacebookPost(parsedData)
		return s.BulkInsertPosts(ctx, []*clickhousemodels.FacebookPosts{clickhousePost})

	case []*kafkamodels.ParsedFacebookPost:
		clickhousePosts := make([]*clickhousemodels.FacebookPosts, 0, len(parsedData))
		for _, post := range parsedData {
			if converted := s.ConvertFacebookPost(post); converted != nil {
				clickhousePosts = append(clickhousePosts, converted)
			}
		}
		return s.BulkInsertPosts(ctx, clickhousePosts)

	case *kafkamodels.ParsedFacebookVideoInsights:
		clickhouseInsights := s.ConvertFacebookVideoInsights(parsedData)
		return s.BulkInsertVideoInsights(ctx, []*clickhousemodels.FacebookVideoInsights{clickhouseInsights})

	case []*kafkamodels.ParsedFacebookVideoInsights:
		clickhouseInsights := make([]*clickhousemodels.FacebookVideoInsights, 0, len(parsedData))
		for _, insights := range parsedData {
			if converted := s.ConvertFacebookVideoInsights(insights); converted != nil {
				clickhouseInsights = append(clickhouseInsights, converted)
			}
		}
		return s.BulkInsertVideoInsights(ctx, clickhouseInsights)

	case *kafkamodels.ParsedFacebookReelsInsights:
		clickhouseInsights := s.ConvertFacebookReelsInsights(parsedData)
		return s.BulkInsertReelsInsights(ctx, []*clickhousemodels.FacebookReelsInsights{clickhouseInsights})

	case []*kafkamodels.ParsedFacebookReelsInsights:
		clickhouseInsights := make([]*clickhousemodels.FacebookReelsInsights, 0, len(parsedData))
		for _, insights := range parsedData {
			if converted := s.ConvertFacebookReelsInsights(insights); converted != nil {
				clickhouseInsights = append(clickhouseInsights, converted)
			}
		}
		return s.BulkInsertReelsInsights(ctx, clickhouseInsights)

	case *kafkamodels.ParsedFacebookInsights:
		clickhouseInsights := s.ConvertFacebookInsights(parsedData)
		return s.BulkInsertInsights(ctx, []*clickhousemodels.FacebookInsights{clickhouseInsights})

	case []*kafkamodels.ParsedFacebookInsights:
		clickhouseInsights := make([]*clickhousemodels.FacebookInsights, 0, len(parsedData))
		for _, insights := range parsedData {
			if converted := s.ConvertFacebookInsights(insights); converted != nil {
				clickhouseInsights = append(clickhouseInsights, converted)
			}
		}
		return s.BulkInsertInsights(ctx, clickhouseInsights)

	case *kafkamodels.ParsedFacebookMediaAsset:
		clickhouseAsset := s.ConvertFacebookMediaAssets(parsedData)
		return s.BulkInsertMediaAssets(ctx, []*clickhousemodels.FacebookMediaAssets{clickhouseAsset})

	case []*kafkamodels.ParsedFacebookMediaAsset:
		clickhouseAssets := make([]*clickhousemodels.FacebookMediaAssets, 0, len(parsedData))
		for _, asset := range parsedData {
			if converted := s.ConvertFacebookMediaAssets(asset); converted != nil {
				clickhouseAssets = append(clickhouseAssets, converted)
			}
		}
		return s.BulkInsertMediaAssets(ctx, clickhouseAssets)

	default:
		return fmt.Errorf("ClickHouseSink.ProcessParsedData: unsupported data type: %T", data)
	}
}

// convertStringArrayFields handles conversion of array fields, filtering out empty strings
func convertStringArrayFields(input []string) []string {
	if len(input) == 0 {
		return []string{}
	}

	result := make([]string, 0, len(input))
	for _, str := range input {
		if strings.TrimSpace(str) != "" {
			result = append(result, str)
		}
	}

	return result
}

// HandleFailedInsert handles failed inserts by pushing to Redis queue for retry
func (s *ClickHouseSink) HandleFailedInsert(ctx context.Context, data interface{}, err error) {
	dataType := fmt.Sprintf("%T", data)

	s.logger.Error().
		Err(err).
		Str("data_type", dataType).
		Str("error_message", err.Error()).
		Str("function", "HandleFailedInsert").
		Str("stage", "bulk_insert").
		Msg("Failed to insert data to ClickHouse")

	// TODO: Implement Redis queue push for failed inserts
	// Example:
	// s.redisClient.RPush(ctx, "failed_facebook_inserts", data)
}

// Health checks the health of the ClickHouse connection
func (s *ClickHouseSink) Health() error {
	return s.ClickhouseClient.Health()
}
