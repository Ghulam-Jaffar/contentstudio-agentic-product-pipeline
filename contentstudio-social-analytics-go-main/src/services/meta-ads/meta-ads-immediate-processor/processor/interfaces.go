package processor

import (
	"context"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// MetaAdsAPIInterface defines the Meta Ads Graph API operations needed by this processor.
type MetaAdsAPIInterface interface {
	DebugToken(ctx context.Context, inputToken, appAccessToken string) (*social.DebugTokenResult, error)
	FetchAccountInfo(ctx context.Context, accountID, accessToken string) (*kafkamodels.RawMetaAdsAccountInfo, error)
	FetchCampaigns(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsCampaign, error)
	FetchAdsets(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsAdset, error)
	FetchAds(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsAd, error)
	FetchCampaignInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsInsightRow, error)
	FetchAdsetInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsInsightRow, error)
	FetchAdInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsInsightRow, error)
	FetchAgeGenderInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsDemographicsRow, error)
	FetchDevicePlatformInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsDemographicsRow, error)
	FetchRegionCountryInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsDemographicsRow, error)
}

// PusherClientInterface abstracts the Pusher client for testing.
type PusherClientInterface interface {
	Trigger(channel string, event string, data interface{}) error
}

// NotifierInterface abstracts the notification service for testing.
type NotifierInterface interface {
	SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error
}

// ClickHouseSinkInterface defines the ClickHouse operations needed by this processor.
type ClickHouseSinkInterface interface {
	// Account info
	ConvertMetaAdsAccountInfo(accountID string, raw kafkamodels.RawMetaAdsAccountInfo) *clickhousemodels.MetaAdsAccountInfo
	BulkInsertMetaAdsAccountInfo(ctx context.Context, rows []*clickhousemodels.MetaAdsAccountInfo) error
	// Campaigns
	ConvertMetaAdsCampaign(accountID string, raw kafkamodels.RawMetaAdsCampaign) *clickhousemodels.MetaAdsCampaign
	BulkInsertMetaAdsCampaigns(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaign) error
	// Ad sets
	ConvertMetaAdsAdset(accountID string, raw kafkamodels.RawMetaAdsAdset) *clickhousemodels.MetaAdsAdset
	BulkInsertMetaAdsAdsets(ctx context.Context, rows []*clickhousemodels.MetaAdsAdset) error
	// Ads
	ConvertMetaAdsAd(accountID string, raw kafkamodels.RawMetaAdsAd) *clickhousemodels.MetaAdsAd
	BulkInsertMetaAdsAds(ctx context.Context, rows []*clickhousemodels.MetaAdsAd) error
	// Campaign insights
	ConvertMetaAdsCampaignInsight(accountID string, raw kafkamodels.RawMetaAdsInsightRow) *clickhousemodels.MetaAdsCampaignInsights
	BulkInsertMetaAdsCampaignInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsCampaignInsights) error
	// Adset insights
	ConvertMetaAdsAdsetInsight(accountID string, raw kafkamodels.RawMetaAdsInsightRow) *clickhousemodels.MetaAdsAdsetInsights
	BulkInsertMetaAdsAdsetInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdsetInsights) error
	// Ad insights
	ConvertMetaAdsAdInsight(accountID string, raw kafkamodels.RawMetaAdsInsightRow) *clickhousemodels.MetaAdsAdInsights
	BulkInsertMetaAdsAdInsights(ctx context.Context, rows []*clickhousemodels.MetaAdsAdInsights) error
	// Demographics age/gender
	ConvertMetaAdsDemographicsAgeGender(accountID string, raw kafkamodels.RawMetaAdsDemographicsRow) *clickhousemodels.MetaAdsDemographicsAgeGender
	BulkInsertMetaAdsDemographicsAgeGender(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsAgeGender) error
	// Demographics device/platform
	ConvertMetaAdsDemographicsDevicePlatform(accountID string, raw kafkamodels.RawMetaAdsDemographicsRow) *clickhousemodels.MetaAdsDemographicsDevicePlatform
	BulkInsertMetaAdsDemographicsDevicePlatform(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsDevicePlatform) error
	// Demographics region/country
	ConvertMetaAdsDemographicsRegionCountry(accountID string, raw kafkamodels.RawMetaAdsDemographicsRow) *clickhousemodels.MetaAdsDemographicsRegionCountry
	BulkInsertMetaAdsDemographicsRegionCountry(ctx context.Context, rows []*clickhousemodels.MetaAdsDemographicsRegionCountry) error
}
