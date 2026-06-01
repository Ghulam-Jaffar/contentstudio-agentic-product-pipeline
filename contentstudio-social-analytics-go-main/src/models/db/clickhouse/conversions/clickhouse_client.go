package conversions

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/rs/zerolog"
)

// ClickHouseClientInterface defines the interface for ClickHouse operations
type ClickHouseClientInterface interface {
	BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error
	BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error
	BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error
	BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error
	BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error
	BulkInsertInstagramPosts(ctx context.Context, posts []*clickhousemodels.InstagramPost) error
	BulkInsertInstagramInsights(ctx context.Context, insights []*clickhousemodels.InstagramInsight) error
	BulkInsertLinkedInPosts(ctx context.Context, posts []*clickhousemodels.LinkedInPosts) error
	BulkInsertLinkedInInsights(ctx context.Context, insights []*clickhousemodels.LinkedInInsights) error
	BulkInsertTikTokPosts(ctx context.Context, posts []*clickhousemodels.TikTokPosts) error
	BulkInsertTikTokInsights(ctx context.Context, insights []*clickhousemodels.TikTokInsights) error
	BulkInsertTwitterPosts(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error
	BulkInsertTwitterInsights(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error
	BulkInsertYouTubeChannels(ctx context.Context, channels []*clickhousemodels.YouTubeChannel) error
	BulkInsertYouTubeVideos(ctx context.Context, videos []*clickhousemodels.YouTubeVideo) error
	BulkInsertYouTubeActivityInsights(ctx context.Context, insights []*clickhousemodels.YouTubeActivityInsights) error
	BulkInsertYouTubeTrafficInsights(ctx context.Context, insights []*clickhousemodels.YouTubeTrafficInsights) error
	BulkInsertYouTubeSharedInsights(ctx context.Context, insights []*clickhousemodels.YouTubeSharedInsights) error
	BulkInsertPinterestUsers(ctx context.Context, users []clickhousemodels.PinterestUser) error
	BulkInsertPinterestBoards(ctx context.Context, boards []clickhousemodels.PinterestBoard) error
	BulkInsertPinterestPins(ctx context.Context, pins []clickhousemodels.PinterestPin) error
	BulkInsertPinterestPinInsights(ctx context.Context, insights []clickhousemodels.PinterestPinInsight) error
	BulkInsertPinterestUserInsights(ctx context.Context, insights []clickhousemodels.PinterestUserInsight) error
	BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error
	BulkInsertGMBMediaAssets(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error
	BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error
	BulkInsertGMBLocalPosts(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error
	BulkInsertGMBReviews(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error
	// Meta Ads
	BulkInsertMetaAdsAccountInfo(ctx context.Context, rows []*clickhousemodels.MetaAdsAccountInfo) error
	BulkInsertMetaAdsCampaigns(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaign) error
	BulkInsertMetaAdsAdsets(ctx context.Context, rows []*clickhousemodels.MetaAdsAdset) error
	BulkInsertMetaAdsAds(ctx context.Context, rows []*clickhousemodels.MetaAdsAd) error
	BulkInsertMetaAdsCampaignInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaignInsights) error
	BulkInsertMetaAdsAdsetInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdsetInsights) error
	BulkInsertMetaAdsAdInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdInsights) error
	BulkInsertMetaAdsDemographicsAgeGender(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsAgeGender) error
	BulkInsertMetaAdsDemographicsDevicePlatform(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsDevicePlatform) error
	BulkInsertMetaAdsDemographicsRegionCountry(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsRegionCountry) error
	Health() error
	GetMinimalOlderThan20DaysByPage(ctx context.Context, tableName, pageID string, limit, offset int) ([]clickhousemodels.MinimalPost, error)
	UpdateFullPictures(ctx context.Context, tableName, pageID string, posts []clickhousemodels.MinimalPost) (int, error)
	BulkUpdateFullPictures(ctx context.Context, tableName string, posts []clickhousemodels.MinimalPost) (int, error)
	GetMinimalInstagramOlderThan20DaysByAccount(ctx context.Context, tableName, instagramID string, limit, offset int) ([]clickhousemodels.InstagramMinimalPost, error)
	UpdateInstagramMediaURLs(ctx context.Context, tableName, instagramID string, posts []clickhousemodels.InstagramMinimalPost) (int, error)
	BulkUpdateInstagramMediaURLs(ctx context.Context, tableName string, posts []clickhousemodels.InstagramMinimalPost) (int, error)
	GetMinimalLinkedInOlderThan7DaysByAccount(ctx context.Context, tableName, linkedinID string, limit, offset int) ([]clickhousemodels.LinkedInMinimalPost, error)
	UpdateLinkedInPostURLs(ctx context.Context, tableName, linkedinID string, posts []clickhousemodels.LinkedInMinimalPost) (int, error)
	BulkUpdateLinkedInPostURLs(ctx context.Context, tableName string, posts []clickhousemodels.LinkedInMinimalPost) (int, error)
	GetDistinctFacebookPageIDsWithStaleURLs(ctx context.Context, tableName string, validPageIDs []string) ([]string, error)
	GetDistinctInstagramIDsWithStaleURLs(ctx context.Context, tableName string, validIDs []string) ([]string, error)
	GetDistinctLinkedInIDsWithStaleURLs(ctx context.Context, tableName string, validIDs []string) ([]string, error)
	MarkFacebookPostsRefreshed(ctx context.Context, tableName, pageID string) error
	BulkMarkFacebookPostsRefreshed(ctx context.Context, tableName string, pageIDs []string) error
	MarkInstagramPostsRefreshed(ctx context.Context, tableName, instagramID string) error
	BulkMarkInstagramPostsRefreshed(ctx context.Context, tableName string, instagramIDs []string) error
	MarkLinkedInPostsRefreshed(ctx context.Context, tableName, linkedinID string) error
	BulkMarkLinkedInPostsRefreshed(ctx context.Context, tableName string, linkedinIDs []string) error
	GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccount(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error)
	UpdateFacebookCompetitorMediaAssetURLs(ctx context.Context, tableName, facebookID string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error)
	BulkUpdateFacebookCompetitorMediaAssetURLs(ctx context.Context, tableName string, assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset) (int, error)
	GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccount(ctx context.Context, tableName, facebookID string, limit, offset int) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error)
	UpdateFacebookCompetitorSharedPictures(ctx context.Context, tableName, facebookID string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error)
	BulkUpdateFacebookCompetitorSharedPictures(ctx context.Context, tableName string, posts []clickhousemodels.FacebookCompetitorMinimalSharedPost) (int, error)
	GetMinimalInstagramCompetitorOlderThan7DaysByAccount(ctx context.Context, tableName string, instagramID int64, limit, offset int) ([]clickhousemodels.InstagramCompetitorMinimalPost, error)
	UpdateInstagramCompetitorMediaURLs(ctx context.Context, tableName string, instagramID int64, profilePictureURL string, posts []clickhousemodels.InstagramCompetitorMinimalPost) (int, error)
	BulkUpdateInstagramCompetitorMediaURLs(ctx context.Context, tableName string, posts []clickhousemodels.InstagramCompetitorMinimalPost, profilePics map[int64]string) (int, error)
}

// ClickHouseSink handles the conversion and storage of parsed Facebook data into ClickHouse
type ClickHouseSink struct {
	logger           *zerolog.Logger
	ClickhouseClient ClickHouseClientInterface
	// RawClient provides access to the underlying *clickhouse.Client for cases that need it
	// (e.g., GeoResolver which uses ClickHouse-specific methods not in the interface)
	RawClient *clickhouse.Client
}

// newClientFunc is a function type for creating ClickHouse clients (allows mocking in tests)
type newClientFunc func(cfg config.ClickHouseConfig, logger zerolog.Logger) (*clickhouse.Client, error)

// newClickHouseClient is the default client creator, can be overridden in tests
var newClickHouseClient newClientFunc = clickhouse.NewClient

// NewClickHouseSink creates a new ClickHouse sink instance
func NewClickHouseSink(logger *zerolog.Logger, cfg *config.Config) *ClickHouseSink {

	// Create ClickHouse client
	clickhouseClient, err := newClickHouseClient(cfg.ClickHouse, *logger)
	if err != nil {
		logger.Fatal().
			Err(err).
			Str("host", cfg.ClickHouse.Host).
			Int("port", cfg.ClickHouse.Port).
			Str("database", cfg.ClickHouse.Database).
			Msg("Failed to create ClickHouse client for sink")
	}

	return &ClickHouseSink{
		logger:           logger,
		ClickhouseClient: clickhouseClient,
		RawClient:        clickhouseClient,
	}
}

// NewClickHouseSinkWithClient creates a new ClickHouse sink with a provided client (useful for testing)
func NewClickHouseSinkWithClient(logger *zerolog.Logger, client ClickHouseClientInterface) *ClickHouseSink {
	// Try to get raw client if it's a *clickhouse.Client
	rawClient, _ := client.(*clickhouse.Client)
	return &ClickHouseSink{
		logger:           logger,
		ClickhouseClient: client,
		RawClient:        rawClient,
	}
}
