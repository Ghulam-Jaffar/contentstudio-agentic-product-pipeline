package processor

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestResolveAccessToken(t *testing.T) {
	p := &Processor{Config: &config.Config{DecryptionKey: ""}}
	wo := kafkamodels.MetaAdsWorkOrder{LongAccessToken: "long-token", AccessToken: "short-token"}
	account := &mongomodels.SocialIntegration{AccessToken: "account-token"}

	if got, err := p.resolveAccessToken(wo, account); err != nil || got != "long-token" {
		t.Fatalf("unexpected token: %q err=%v", got, err)
	}

	wo = kafkamodels.MetaAdsWorkOrder{AccessToken: "short-token"}
	if got, err := p.resolveAccessToken(wo, account); err != nil || got != "account-token" {
		t.Fatalf("unexpected token: %q err=%v", got, err)
	}

	account = &mongomodels.SocialIntegration{}
	if got, err := p.resolveAccessToken(wo, account); err != nil || got != "short-token" {
		t.Fatalf("unexpected token: %q err=%v", got, err)
	}

	wo = kafkamodels.MetaAdsWorkOrder{}
	if got, err := p.resolveAccessToken(wo, account); err == nil || got != "" {
		t.Fatalf("expected error for missing token, got %q err=%v", got, err)
	}
}

func TestResolveDateRange(t *testing.T) {
	p := &Processor{}
	start, end := p.resolveDateRange("2025-01-10", "2025-02-05")
	if start.Format("2006-01-02") != "2025-01-10" || end.Format("2006-01-02") != "2025-02-05" {
		t.Fatalf("unexpected parsed range: %v - %v", start, end)
	}

	start, end = p.resolveDateRange("", "")
	if end.Hour() != 23 || end.Minute() != 59 || end.Second() != 59 {
		t.Fatalf("unexpected default end: %v", end)
	}
	if end.Sub(start) != metaAdsImmediateMaxDateRange {
		t.Fatalf("unexpected default range duration: %v", end.Sub(start))
	}
}

func TestBatchInsert(t *testing.T) {
	var batches [][]int
	batchInsert(context.Background(), []int{1, 2, 3, 4, 5}, 2, func(rows []int) error {
		batches = append(batches, append([]int(nil), rows...))
		return nil
	}, logger.NewNop(), "endpoint")

	if len(batches) != 3 {
		t.Fatalf("expected 3 batches, got %d", len(batches))
	}
	if len(batches[0]) != 2 || len(batches[2]) != 1 {
		t.Fatalf("unexpected batches: %+v", batches)
	}
}

type concurrentMetaAdsAPI struct {
	start   chan string
	release <-chan struct{}
}

func (m *concurrentMetaAdsAPI) DebugToken(context.Context, string, string) (*social.DebugTokenResult, error) {
	return nil, nil
}
func (m *concurrentMetaAdsAPI) FetchAccountInfo(ctx context.Context, accountID, accessToken string) (*kafkamodels.RawMetaAdsAccountInfo, error) {
	m.start <- "account_info"
	<-m.release
	return &kafkamodels.RawMetaAdsAccountInfo{}, nil
}
func (m *concurrentMetaAdsAPI) FetchCampaigns(context.Context, string, string, time.Time, time.Time) ([]kafkamodels.RawMetaAdsCampaign, error) {
	m.start <- "campaigns"
	<-m.release
	return nil, nil
}
func (m *concurrentMetaAdsAPI) FetchAdsets(context.Context, string, string, time.Time, time.Time) ([]kafkamodels.RawMetaAdsAdset, error) {
	return nil, nil
}
func (m *concurrentMetaAdsAPI) FetchAds(context.Context, string, string, time.Time, time.Time) ([]kafkamodels.RawMetaAdsAd, error) {
	return nil, nil
}
func (m *concurrentMetaAdsAPI) FetchCampaignInsights(context.Context, string, string, time.Time, time.Time) ([]kafkamodels.RawMetaAdsInsightRow, error) {
	return nil, nil
}
func (m *concurrentMetaAdsAPI) FetchAdsetInsights(context.Context, string, string, time.Time, time.Time) ([]kafkamodels.RawMetaAdsInsightRow, error) {
	return nil, nil
}
func (m *concurrentMetaAdsAPI) FetchAdInsights(context.Context, string, string, time.Time, time.Time) ([]kafkamodels.RawMetaAdsInsightRow, error) {
	return nil, nil
}
func (m *concurrentMetaAdsAPI) FetchAgeGenderInsights(context.Context, string, string, time.Time, time.Time) ([]kafkamodels.RawMetaAdsDemographicsRow, error) {
	return nil, nil
}
func (m *concurrentMetaAdsAPI) FetchDevicePlatformInsights(context.Context, string, string, time.Time, time.Time) ([]kafkamodels.RawMetaAdsDemographicsRow, error) {
	return nil, nil
}
func (m *concurrentMetaAdsAPI) FetchRegionCountryInsights(context.Context, string, string, time.Time, time.Time) ([]kafkamodels.RawMetaAdsDemographicsRow, error) {
	return nil, nil
}

func TestFetchAndStoreRunsConcurrentFetches(t *testing.T) {
	started := make(chan string, 2)
	release := make(chan struct{})
	api := &concurrentMetaAdsAPI{
		start:   started,
		release: release,
	}
	zlog := zerolog.New(io.Discard)
	sink := conversions.NewClickHouseSinkWithClient(&zlog, &conversions.MockClickHouseClient{})
	p := &Processor{
		MetaAdsAPI: api,
		Sink:       sink,
		Logger:     logger.NewNop(),
	}

	wo := kafkamodels.MetaAdsWorkOrder{PlatformIdentifier: "act_123"}
	account := &mongomodels.SocialIntegration{}
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		done <- p.fetchAndStore(ctx, wo, account, "access-token", time.Now().Add(-24*time.Hour), time.Now())
	}()

	first := <-started
	second := <-started
	close(release)

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	case <-ctx.Done():
		t.Fatal("timed out waiting for concurrent fetches to complete")
	}

	if first == second {
		t.Fatal("expected two distinct fetches to start concurrently")
	}
}
