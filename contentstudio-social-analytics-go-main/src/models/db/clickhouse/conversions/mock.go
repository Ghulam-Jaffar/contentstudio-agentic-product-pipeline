package conversions

import (
	"context"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// MockClickHouseClient is a mock implementation of ClickHouseClientInterface for testing.
// It can be used by any service that needs to mock ClickHouse operations.
type MockClickHouseClient struct {
	BulkInsertPostsFunc                    func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error
	BulkInsertMediaAssetsFunc              func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error
	BulkInsertVideoInsightsFunc            func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error
	BulkInsertReelsInsightsFunc            func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error
	BulkInsertInsightsFunc                 func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error
	BulkInsertInstagramPostsFunc           func(ctx context.Context, posts []*clickhousemodels.InstagramPost) error
	BulkInsertInstagramInsightsFunc        func(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error
	BulkInsertLinkedInPostsFunc            func(ctx context.Context, posts []*clickhousemodels.LinkedInPosts) error
	BulkInsertLinkedInInsightsFunc         func(ctx context.Context, insights []*clickhousemodels.LinkedInInsights) error
	BulkInsertTikTokPostsFunc              func(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error
	BulkInsertTikTokInsightsFunc           func(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error
	BulkInsertTwitterPostsFunc             func(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error
	BulkInsertTwitterInsightsFunc          func(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error
	BulkInsertYouTubeChannelsFunc          func(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error
	BulkInsertYouTubeVideosFunc            func(ctx context.Context, videos []*clickhousemodels.YouTubeVideo) error
	BulkInsertYouTubeActivityInsightsFunc  func(ctx context.Context, insights []*clickhousemodels.YouTubeActivityInsights) error
	BulkInsertYouTubeTrafficInsightsFunc   func(ctx context.Context, insights []*clickhousemodels.YouTubeTrafficInsights) error
	BulkInsertYouTubeSharedInsightsFunc    func(ctx context.Context, insights []*clickhousemodels.YouTubeSharedInsights) error
	BulkInsertPinterestUsersFunc           func(ctx context.Context, users []clickhousemodels.PinterestUser) error
	BulkInsertPinterestBoardsFunc          func(ctx context.Context, boards []clickhousemodels.PinterestBoard) error
	BulkInsertPinterestPinsFunc            func(ctx context.Context, pins []clickhousemodels.PinterestPin) error
	BulkInsertPinterestPinInsightsFunc     func(ctx context.Context, insights []clickhousemodels.PinterestPinInsight) error
	BulkInsertPinterestUserInsightsFunc    func(ctx context.Context, insights []clickhousemodels.PinterestUserInsight) error
	BulkInsertGMBDailyMetricsFunc          func(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error
	BulkInsertGMBMediaAssetsFunc           func(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error
	BulkInsertGMBSearchKeywordsMonthlyFunc func(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error
	BulkInsertGMBLocalPostsFunc            func(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error
	BulkInsertGMBReviewsFunc               func(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error
	// Meta Ads mock funcs
	BulkInsertMetaAdsAccountInfoFunc                                   func(ctx context.Context, rows []*clickhousemodels.MetaAdsAccountInfo) error
	BulkInsertMetaAdsCampaignsFunc                                     func(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaign) error
	BulkInsertMetaAdsAdsetsFunc                                        func(ctx context.Context, rows []*clickhousemodels.MetaAdsAdset) error
	BulkInsertMetaAdsAdsFunc                                           func(ctx context.Context, rows []*clickhousemodels.MetaAdsAd) error
	BulkInsertMetaAdsCampaignInsightsFunc                              func(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaignInsights) error
	BulkInsertMetaAdsAdsetInsightsFunc                                 func(ctx context.Context, rows []*clickhousemodels.MetaAdsAdsetInsights) error
	BulkInsertMetaAdsAdInsightsFunc                                    func(ctx context.Context, rows []*clickhousemodels.MetaAdsAdInsights) error
	BulkInsertMetaAdsDemographicsAgeGenderFunc                         func(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsAgeGender) error
	BulkInsertMetaAdsDemographicsDevicePlatformFunc                    func(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsDevicePlatform) error
	BulkInsertMetaAdsDemographicsRegionCountryFunc                     func(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsRegionCountry) error
	HealthFunc                                                         func() error
	GetMinimalOlderThan20DaysByPageFunc                                func(ctx context.Context, tableName, pageID string, limit, offset int) ([]clickhousemodels.MinimalPost, error)
	UpdateFullPicturesFunc                                             func(ctx context.Context, tableName, pageID string, posts []clickhousemodels.MinimalPost) (int, error)
	BulkUpdateFullPicturesFunc                                         func(ctx context.Context, tableName string, posts []clickhousemodels.MinimalPost) (int, error)
	GetMinimalInstagramOlderThan20DaysByAccountFunc                    func(ctx context.Context, tableName, instagramID string, limit, offset int) ([]clickhousemodels.InstagramMinimalPost, error)
	UpdateInstagramMediaURLsFunc                                       func(ctx context.Context, tableName, instagramID string, posts []clickhousemodels.InstagramMinimalPost) (int, error)
	BulkUpdateInstagramMediaURLsFunc                                   func(ctx context.Context, tableName string, posts []clickhousemodels.InstagramMinimalPost) (int, error)
	GetMinimalLinkedInOlderThan7DaysByAccountFunc                      func(ctx context.Context, tableName, linkedinID string, limit, offset int) ([]clickhousemodels.LinkedInMinimalPost, error)
	UpdateLinkedInPostURLsFunc                                         func(ctx context.Context, tableName, linkedinID string, posts []clickhousemodels.LinkedInMinimalPost) (int, error)
	BulkUpdateLinkedInPostURLsFunc                                     func(ctx context.Context, tableName string, posts []clickhousemodels.LinkedInMinimalPost) (int, error)
	GetDistinctFacebookPageIDsWithStaleURLsFunc                        func(ctx context.Context, tableName string, validPageIDs []string) ([]string, error)
	GetDistinctInstagramIDsWithStaleURLsFunc                           func(ctx context.Context, tableName string, validIDs []string) ([]string, error)
	GetDistinctLinkedInIDsWithStaleURLsFunc                            func(ctx context.Context, tableName string, validIDs []string) ([]string, error)
	MarkFacebookPostsRefreshedFunc                                     func(ctx context.Context, tableName, pageID string) error
	BulkMarkFacebookPostsRefreshedFunc                                 func(ctx context.Context, tableName string, pageIDs []string) error
	MarkInstagramPostsRefreshedFunc                                    func(ctx context.Context, tableName, instagramID string) error
	BulkMarkInstagramPostsRefreshedFunc                                func(ctx context.Context, tableName string, instagramIDs []string) error
	BulkMarkLinkedInPostsRefreshedFunc                                 func(ctx context.Context, tableName string, linkedinIDs []string) error
	MarkLinkedInPostsRefreshedFunc                                     func(ctx context.Context, tableName, linkedinID string) error
	GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccountFunc func(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error)
	UpdateFacebookCompetitorMediaAssetURLsFunc                         func(ctx context.Context, tableName, facebookID string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error)
	BulkUpdateFacebookCompetitorMediaAssetURLsFunc                     func(ctx context.Context, tableName string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error)
	GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccountFunc func(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error)
	UpdateFacebookCompetitorSharedPicturesFunc                         func(ctx context.Context, tableName, facebookID string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error)
	BulkUpdateFacebookCompetitorSharedPicturesFunc                     func(ctx context.Context, tableName string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error)
	GetMinimalInstagramCompetitorOlderThan7DaysByAccountFunc           func(ctx context.Context, tableName string, instagramID int64, limit, offset int) ([]clickhousemodels.InstagramCompetitorMinimalPost, error)
	UpdateInstagramCompetitorMediaURLsFunc                             func(ctx context.Context, tableName string, instagramID int64, profilePictureURL string, posts []clickhousemodels.InstagramCompetitorMinimalPost) (int, error)
	BulkUpdateInstagramCompetitorMediaURLsFunc                         func(ctx context.Context, tableName string, posts []clickhousemodels.InstagramCompetitorMinimalPost, profilePics map[int64]string) (int, error)
}

func (m *MockClickHouseClient) BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
	if m.BulkInsertPostsFunc != nil {
		return m.BulkInsertPostsFunc(ctx, posts)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
	if m.BulkInsertMediaAssetsFunc != nil {
		return m.BulkInsertMediaAssetsFunc(ctx, assets)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
	if m.BulkInsertVideoInsightsFunc != nil {
		return m.BulkInsertVideoInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
	if m.BulkInsertReelsInsightsFunc != nil {
		return m.BulkInsertReelsInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
	if m.BulkInsertInsightsFunc != nil {
		return m.BulkInsertInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertInstagramPosts(ctx context.Context, posts []*clickhousemodels.InstagramPost) error {
	if m.BulkInsertInstagramPostsFunc != nil {
		return m.BulkInsertInstagramPostsFunc(ctx, posts)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertInstagramInsights(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error {
	if m.BulkInsertInstagramInsightsFunc != nil {
		return m.BulkInsertInstagramInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertLinkedInPosts(ctx context.Context, posts []*clickhousemodels.LinkedInPosts) error {
	if m.BulkInsertLinkedInPostsFunc != nil {
		return m.BulkInsertLinkedInPostsFunc(ctx, posts)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertLinkedInInsights(ctx context.Context, insights []*clickhousemodels.LinkedInInsights) error {
	if m.BulkInsertLinkedInInsightsFunc != nil {
		return m.BulkInsertLinkedInInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertTikTokPosts(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error {
	if m.BulkInsertTikTokPostsFunc != nil {
		return m.BulkInsertTikTokPostsFunc(ctx, posts)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertTikTokInsights(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error {
	if m.BulkInsertTikTokInsightsFunc != nil {
		return m.BulkInsertTikTokInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertTwitterPosts(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error {
	if m.BulkInsertTwitterPostsFunc != nil {
		return m.BulkInsertTwitterPostsFunc(ctx, posts)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertTwitterInsights(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error {
	if m.BulkInsertTwitterInsightsFunc != nil {
		return m.BulkInsertTwitterInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertYouTubeChannels(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error {
	if m.BulkInsertYouTubeChannelsFunc != nil {
		return m.BulkInsertYouTubeChannelsFunc(ctx, channels)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertYouTubeVideos(ctx context.Context, videos []*clickhousemodels.YouTubeVideo) error {
	if m.BulkInsertYouTubeVideosFunc != nil {
		return m.BulkInsertYouTubeVideosFunc(ctx, videos)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertYouTubeActivityInsights(ctx context.Context, insights []*clickhousemodels.YouTubeActivityInsights) error {
	if m.BulkInsertYouTubeActivityInsightsFunc != nil {
		return m.BulkInsertYouTubeActivityInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertYouTubeTrafficInsights(ctx context.Context, insights []*clickhousemodels.YouTubeTrafficInsights) error {
	if m.BulkInsertYouTubeTrafficInsightsFunc != nil {
		return m.BulkInsertYouTubeTrafficInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertYouTubeSharedInsights(ctx context.Context, insights []*clickhousemodels.YouTubeSharedInsights) error {
	if m.BulkInsertYouTubeSharedInsightsFunc != nil {
		return m.BulkInsertYouTubeSharedInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertPinterestUsers(ctx context.Context, users []clickhousemodels.PinterestUser) error {
	if m.BulkInsertPinterestUsersFunc != nil {
		return m.BulkInsertPinterestUsersFunc(ctx, users)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertPinterestBoards(ctx context.Context, boards []clickhousemodels.PinterestBoard) error {
	if m.BulkInsertPinterestBoardsFunc != nil {
		return m.BulkInsertPinterestBoardsFunc(ctx, boards)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertPinterestPins(ctx context.Context, pins []clickhousemodels.PinterestPin) error {
	if m.BulkInsertPinterestPinsFunc != nil {
		return m.BulkInsertPinterestPinsFunc(ctx, pins)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertPinterestPinInsights(ctx context.Context, insights []clickhousemodels.PinterestPinInsight) error {
	if m.BulkInsertPinterestPinInsightsFunc != nil {
		return m.BulkInsertPinterestPinInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertPinterestUserInsights(ctx context.Context, insights []clickhousemodels.PinterestUserInsight) error {
	if m.BulkInsertPinterestUserInsightsFunc != nil {
		return m.BulkInsertPinterestUserInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error {
	if m.BulkInsertGMBDailyMetricsFunc != nil {
		return m.BulkInsertGMBDailyMetricsFunc(ctx, metrics)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertGMBMediaAssets(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error {
	if m.BulkInsertGMBMediaAssetsFunc != nil {
		return m.BulkInsertGMBMediaAssetsFunc(ctx, assets)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error {
	if m.BulkInsertGMBSearchKeywordsMonthlyFunc != nil {
		return m.BulkInsertGMBSearchKeywordsMonthlyFunc(ctx, keywords)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertGMBLocalPosts(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error {
	if m.BulkInsertGMBLocalPostsFunc != nil {
		return m.BulkInsertGMBLocalPostsFunc(ctx, posts)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertGMBReviews(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error {
	if m.BulkInsertGMBReviewsFunc != nil {
		return m.BulkInsertGMBReviewsFunc(ctx, reviews)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsAccountInfo(ctx context.Context, rows []*clickhousemodels.MetaAdsAccountInfo) error {
	if m.BulkInsertMetaAdsAccountInfoFunc != nil {
		return m.BulkInsertMetaAdsAccountInfoFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsCampaigns(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaign) error {
	if m.BulkInsertMetaAdsCampaignsFunc != nil {
		return m.BulkInsertMetaAdsCampaignsFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsAdsets(ctx context.Context, rows []*clickhousemodels.MetaAdsAdset) error {
	if m.BulkInsertMetaAdsAdsetsFunc != nil {
		return m.BulkInsertMetaAdsAdsetsFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsAds(ctx context.Context, rows []*clickhousemodels.MetaAdsAd) error {
	if m.BulkInsertMetaAdsAdsFunc != nil {
		return m.BulkInsertMetaAdsAdsFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsCampaignInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaignInsights) error {
	if m.BulkInsertMetaAdsCampaignInsightsFunc != nil {
		return m.BulkInsertMetaAdsCampaignInsightsFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsAdsetInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdsetInsights) error {
	if m.BulkInsertMetaAdsAdsetInsightsFunc != nil {
		return m.BulkInsertMetaAdsAdsetInsightsFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsAdInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdInsights) error {
	if m.BulkInsertMetaAdsAdInsightsFunc != nil {
		return m.BulkInsertMetaAdsAdInsightsFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsDemographicsAgeGender(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsAgeGender) error {
	if m.BulkInsertMetaAdsDemographicsAgeGenderFunc != nil {
		return m.BulkInsertMetaAdsDemographicsAgeGenderFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsDemographicsDevicePlatform(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsDevicePlatform) error {
	if m.BulkInsertMetaAdsDemographicsDevicePlatformFunc != nil {
		return m.BulkInsertMetaAdsDemographicsDevicePlatformFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) BulkInsertMetaAdsDemographicsRegionCountry(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsRegionCountry) error {
	if m.BulkInsertMetaAdsDemographicsRegionCountryFunc != nil {
		return m.BulkInsertMetaAdsDemographicsRegionCountryFunc(ctx, rows)
	}
	return nil
}

func (m *MockClickHouseClient) Health() error {
	if m.HealthFunc != nil {
		return m.HealthFunc()
	}
	return nil
}

func (m *MockClickHouseClient) GetMinimalOlderThan20DaysByPage(ctx context.Context, tableName, pageID string, limit, offset int) ([]clickhousemodels.MinimalPost, error) {
	if m.GetMinimalOlderThan20DaysByPageFunc != nil {
		return m.GetMinimalOlderThan20DaysByPageFunc(ctx, tableName, pageID, limit, offset)
	}
	return nil, nil
}

func (m *MockClickHouseClient) UpdateFullPictures(ctx context.Context, tableName, pageID string, posts []clickhousemodels.MinimalPost) (int, error) {
	if m.UpdateFullPicturesFunc != nil {
		return m.UpdateFullPicturesFunc(ctx, tableName, pageID, posts)
	}
	return 0, nil
}

func (m *MockClickHouseClient) BulkUpdateFullPictures(ctx context.Context, tableName string, posts []clickhousemodels.MinimalPost) (int, error) {
	if m.BulkUpdateFullPicturesFunc != nil {
		return m.BulkUpdateFullPicturesFunc(ctx, tableName, posts)
	}
	return 0, nil
}

func (m *MockClickHouseClient) GetMinimalInstagramOlderThan20DaysByAccount(ctx context.Context, tableName, instagramID string, limit, offset int) ([]clickhousemodels.InstagramMinimalPost, error) {
	if m.GetMinimalInstagramOlderThan20DaysByAccountFunc != nil {
		return m.GetMinimalInstagramOlderThan20DaysByAccountFunc(ctx, tableName, instagramID, limit, offset)
	}
	return nil, nil
}

func (m *MockClickHouseClient) UpdateInstagramMediaURLs(ctx context.Context, tableName, instagramID string, posts []clickhousemodels.InstagramMinimalPost) (int, error) {
	if m.UpdateInstagramMediaURLsFunc != nil {
		return m.UpdateInstagramMediaURLsFunc(ctx, tableName, instagramID, posts)
	}
	return 0, nil
}

func (m *MockClickHouseClient) BulkUpdateInstagramMediaURLs(ctx context.Context, tableName string, posts []clickhousemodels.InstagramMinimalPost) (int, error) {
	if m.BulkUpdateInstagramMediaURLsFunc != nil {
		return m.BulkUpdateInstagramMediaURLsFunc(ctx, tableName, posts)
	}
	return 0, nil
}

func (m *MockClickHouseClient) GetMinimalLinkedInOlderThan7DaysByAccount(ctx context.Context, tableName, linkedinID string, limit, offset int) ([]clickhousemodels.LinkedInMinimalPost, error) {
	if m.GetMinimalLinkedInOlderThan7DaysByAccountFunc != nil {
		return m.GetMinimalLinkedInOlderThan7DaysByAccountFunc(ctx, tableName, linkedinID, limit, offset)
	}
	return nil, nil
}

func (m *MockClickHouseClient) UpdateLinkedInPostURLs(ctx context.Context, tableName, linkedinID string, posts []clickhousemodels.LinkedInMinimalPost) (int, error) {
	if m.UpdateLinkedInPostURLsFunc != nil {
		return m.UpdateLinkedInPostURLsFunc(ctx, tableName, linkedinID, posts)
	}
	return 0, nil
}

func (m *MockClickHouseClient) BulkUpdateLinkedInPostURLs(ctx context.Context, tableName string, posts []clickhousemodels.LinkedInMinimalPost) (int, error) {
	if m.BulkUpdateLinkedInPostURLsFunc != nil {
		return m.BulkUpdateLinkedInPostURLsFunc(ctx, tableName, posts)
	}
	return 0, nil
}

func (m *MockClickHouseClient) GetDistinctFacebookPageIDsWithStaleURLs(ctx context.Context, tableName string, validPageIDs []string) ([]string, error) {
	if m.GetDistinctFacebookPageIDsWithStaleURLsFunc != nil {
		return m.GetDistinctFacebookPageIDsWithStaleURLsFunc(ctx, tableName, validPageIDs)
	}
	return nil, nil
}

func (m *MockClickHouseClient) GetDistinctInstagramIDsWithStaleURLs(ctx context.Context, tableName string, validIDs []string) ([]string, error) {
	if m.GetDistinctInstagramIDsWithStaleURLsFunc != nil {
		return m.GetDistinctInstagramIDsWithStaleURLsFunc(ctx, tableName, validIDs)
	}
	return nil, nil
}

func (m *MockClickHouseClient) GetDistinctLinkedInIDsWithStaleURLs(ctx context.Context, tableName string, validIDs []string) ([]string, error) {
	if m.GetDistinctLinkedInIDsWithStaleURLsFunc != nil {
		return m.GetDistinctLinkedInIDsWithStaleURLsFunc(ctx, tableName, validIDs)
	}
	return nil, nil
}

func (m *MockClickHouseClient) MarkFacebookPostsRefreshed(ctx context.Context, tableName, pageID string) error {
	if m.MarkFacebookPostsRefreshedFunc != nil {
		return m.MarkFacebookPostsRefreshedFunc(ctx, tableName, pageID)
	}
	return nil
}

func (m *MockClickHouseClient) BulkMarkFacebookPostsRefreshed(ctx context.Context, tableName string, pageIDs []string) error {
	if m.BulkMarkFacebookPostsRefreshedFunc != nil {
		return m.BulkMarkFacebookPostsRefreshedFunc(ctx, tableName, pageIDs)
	}
	return nil
}

func (m *MockClickHouseClient) MarkInstagramPostsRefreshed(ctx context.Context, tableName, instagramID string) error {
	if m.MarkInstagramPostsRefreshedFunc != nil {
		return m.MarkInstagramPostsRefreshedFunc(ctx, tableName, instagramID)
	}
	return nil
}

func (m *MockClickHouseClient) BulkMarkInstagramPostsRefreshed(ctx context.Context, tableName string, instagramIDs []string) error {
	if m.BulkMarkInstagramPostsRefreshedFunc != nil {
		return m.BulkMarkInstagramPostsRefreshedFunc(ctx, tableName, instagramIDs)
	}
	return nil
}

func (m *MockClickHouseClient) MarkLinkedInPostsRefreshed(ctx context.Context, tableName, linkedinID string) error {
	if m.MarkLinkedInPostsRefreshedFunc != nil {
		return m.MarkLinkedInPostsRefreshedFunc(ctx, tableName, linkedinID)
	}
	return nil
}

func (m *MockClickHouseClient) BulkMarkLinkedInPostsRefreshed(ctx context.Context, tableName string, linkedinIDs []string) error {
	if m.BulkMarkLinkedInPostsRefreshedFunc != nil {
		return m.BulkMarkLinkedInPostsRefreshedFunc(ctx, tableName, linkedinIDs)
	}
	return nil
}

func (m *MockClickHouseClient) GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccount(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
	if m.GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccountFunc != nil {
		return m.GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccountFunc(ctx, tableName, facebookID, limit, offset)
	}
	return nil, nil
}

func (m *MockClickHouseClient) UpdateFacebookCompetitorMediaAssetURLs(ctx context.Context, tableName, facebookID string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error) {
	if m.UpdateFacebookCompetitorMediaAssetURLsFunc != nil {
		return m.UpdateFacebookCompetitorMediaAssetURLsFunc(ctx, tableName, facebookID, assets)
	}
	return 0, nil
}

func (m *MockClickHouseClient) BulkUpdateFacebookCompetitorMediaAssetURLs(ctx context.Context, tableName string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error) {
	if m.BulkUpdateFacebookCompetitorMediaAssetURLsFunc != nil {
		return m.BulkUpdateFacebookCompetitorMediaAssetURLsFunc(ctx, tableName, assets)
	}
	return 0, nil
}

func (m *MockClickHouseClient) GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccount(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
	if m.GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccountFunc != nil {
		return m.GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccountFunc(ctx, tableName, facebookID, limit, offset)
	}
	return nil, nil
}

func (m *MockClickHouseClient) UpdateFacebookCompetitorSharedPictures(ctx context.Context, tableName, facebookID string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error) {
	if m.UpdateFacebookCompetitorSharedPicturesFunc != nil {
		return m.UpdateFacebookCompetitorSharedPicturesFunc(ctx, tableName, facebookID, posts)
	}
	return 0, nil
}

func (m *MockClickHouseClient) BulkUpdateFacebookCompetitorSharedPictures(ctx context.Context, tableName string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error) {
	if m.BulkUpdateFacebookCompetitorSharedPicturesFunc != nil {
		return m.BulkUpdateFacebookCompetitorSharedPicturesFunc(ctx, tableName, posts)
	}
	return 0, nil
}

func (m *MockClickHouseClient) GetMinimalInstagramCompetitorOlderThan7DaysByAccount(ctx context.Context, tableName string, instagramID int64, limit, offset int) ([]clickhousemodels.InstagramCompetitorMinimalPost, error) {
	if m.GetMinimalInstagramCompetitorOlderThan7DaysByAccountFunc != nil {
		return m.GetMinimalInstagramCompetitorOlderThan7DaysByAccountFunc(ctx, tableName, instagramID, limit, offset)
	}
	return nil, nil
}

func (m *MockClickHouseClient) UpdateInstagramCompetitorMediaURLs(ctx context.Context, tableName string, instagramID int64, profilePictureURL string, posts []clickhousemodels.InstagramCompetitorMinimalPost) (int, error) {
	if m.UpdateInstagramCompetitorMediaURLsFunc != nil {
		return m.UpdateInstagramCompetitorMediaURLsFunc(ctx, tableName, instagramID, profilePictureURL, posts)
	}
	return 0, nil
}

func (m *MockClickHouseClient) BulkUpdateInstagramCompetitorMediaURLs(ctx context.Context, tableName string, posts []clickhousemodels.InstagramCompetitorMinimalPost, profilePics map[int64]string) (int, error) {
	if m.BulkUpdateInstagramCompetitorMediaURLsFunc != nil {
		return m.BulkUpdateInstagramCompetitorMediaURLsFunc(ctx, tableName, posts, profilePics)
	}
	return 0, nil
}

// Verify mock implements interface at compile time
var _ ClickHouseClientInterface = (*MockClickHouseClient)(nil)

// MockClickHouseSink is a mock implementation that includes both client operations
// and conversion methods for testing analytics sink services.
type MockClickHouseSink struct {
	MockClickHouseClient
	ConvertFacebookPostFunc          func(p *kafkamodels.ParsedFacebookPost) *clickhousemodels.FacebookPosts
	ConvertFacebookMediaAssetsFunc   func(a *kafkamodels.ParsedFacebookMediaAsset) *clickhousemodels.FacebookMediaAssets
	ConvertFacebookInsightsFunc      func(ins *kafkamodels.ParsedFacebookInsights) *clickhousemodels.FacebookInsights
	ConvertFacebookVideoInsightsFunc func(vi *kafkamodels.ParsedFacebookVideoInsights) *clickhousemodels.FacebookVideoInsights
	ConvertFacebookReelsInsightsFunc func(ri *kafkamodels.ParsedFacebookReelsInsights) *clickhousemodels.FacebookReelsInsights
}

func (m *MockClickHouseSink) ConvertFacebookPost(p *kafkamodels.ParsedFacebookPost) *clickhousemodels.FacebookPosts {
	if m.ConvertFacebookPostFunc != nil {
		return m.ConvertFacebookPostFunc(p)
	}
	return &clickhousemodels.FacebookPosts{
		PostID: p.PostID,
		PageID: p.PageID,
	}
}

func (m *MockClickHouseSink) ConvertFacebookMediaAssets(a *kafkamodels.ParsedFacebookMediaAsset) *clickhousemodels.FacebookMediaAssets {
	if m.ConvertFacebookMediaAssetsFunc != nil {
		return m.ConvertFacebookMediaAssetsFunc(a)
	}
	return &clickhousemodels.FacebookMediaAssets{
		PostID: a.PostID,
	}
}

func (m *MockClickHouseSink) ConvertFacebookInsights(ins *kafkamodels.ParsedFacebookInsights) *clickhousemodels.FacebookInsights {
	if m.ConvertFacebookInsightsFunc != nil {
		return m.ConvertFacebookInsightsFunc(ins)
	}
	return &clickhousemodels.FacebookInsights{
		PageID: ins.PageID,
	}
}

func (m *MockClickHouseSink) ConvertFacebookVideoInsights(vi *kafkamodels.ParsedFacebookVideoInsights) *clickhousemodels.FacebookVideoInsights {
	if m.ConvertFacebookVideoInsightsFunc != nil {
		return m.ConvertFacebookVideoInsightsFunc(vi)
	}
	return &clickhousemodels.FacebookVideoInsights{
		PostID: vi.PostID,
	}
}

func (m *MockClickHouseSink) ConvertFacebookReelsInsights(ri *kafkamodels.ParsedFacebookReelsInsights) *clickhousemodels.FacebookReelsInsights {
	if m.ConvertFacebookReelsInsightsFunc != nil {
		return m.ConvertFacebookReelsInsightsFunc(ri)
	}
	return &clickhousemodels.FacebookReelsInsights{
		PostID: ri.PostID,
	}
}

// NewMockClickHouseSink creates a new mock sink with default implementations
func NewMockClickHouseSink() *MockClickHouseSink {
	return &MockClickHouseSink{}
}
