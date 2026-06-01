package processor

import (
	"context"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestInterfacesCompile(t *testing.T) {
	var _ MetaAdsAPIInterface = (*social.MetaAdsClient)(nil)
	var _ ClickHouseSinkInterface = (*conversions.ClickHouseSink)(nil)
	var _ PusherClientInterface = (*noopPusher)(nil)
	var _ NotifierInterface = (*noopNotifier)(nil)
}

type noopPusher struct{}

func (n *noopPusher) Trigger(channel string, event string, data interface{}) error { return nil }

type noopNotifier struct{}

func (n *noopNotifier) SendAnalyticsNotification(userID, workspaceID, platform, accountID, accountName string, isCompetitor bool) error {
	return nil
}

type noopAPI struct{}

func (n *noopAPI) DebugToken(ctx context.Context, inputToken, appAccessToken string) (*social.DebugTokenResult, error) {
	return nil, nil
}
func (n *noopAPI) FetchAccountInfo(ctx context.Context, accountID, accessToken string) (*kafkamodels.RawMetaAdsAccountInfo, error) {
	return nil, nil
}
func (n *noopAPI) FetchCampaigns(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsCampaign, error) {
	return nil, nil
}
func (n *noopAPI) FetchAdsets(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsAdset, error) {
	return nil, nil
}
func (n *noopAPI) FetchAds(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsAd, error) {
	return nil, nil
}
func (n *noopAPI) FetchCampaignInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsInsightRow, error) {
	return nil, nil
}
func (n *noopAPI) FetchAdsetInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsInsightRow, error) {
	return nil, nil
}
func (n *noopAPI) FetchAdInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsInsightRow, error) {
	return nil, nil
}
func (n *noopAPI) FetchAgeGenderInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsDemographicsRow, error) {
	return nil, nil
}
func (n *noopAPI) FetchDevicePlatformInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsDemographicsRow, error) {
	return nil, nil
}
func (n *noopAPI) FetchRegionCountryInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]kafkamodels.RawMetaAdsDemographicsRow, error) {
	return nil, nil
}
