package conversions

import (
	"context"
	"testing"
	"time"

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestConvertMetaAdsAccountInfo(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	row := sink.ConvertMetaAdsAccountInfo("act_1", kafkamodels.RawMetaAdsAccountInfo{
		Name:          "Account",
		Currency:      "USD",
		AccountStatus: 1,
		TimezoneName:  "UTC",
		Business: &struct {
			ID   string `json:"id"`
			Name string `json:"name"`
		}{ID: "biz-1", Name: "Business"},
		CreatedTime: kafkamodels.MetaAdsAPITime{Time: time.Date(2025, 5, 13, 17, 42, 11, 0, time.UTC)},
	})
	if row.AccountID != "act_1" || row.BusinessName != "Business" {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestConvertMetaAdsCampaign(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	row := sink.ConvertMetaAdsCampaign("act_1", kafkamodels.RawMetaAdsCampaign{ID: "camp-1", Name: "Campaign"})
	if row.CampaignID != "camp-1" || row.AccountID != "act_1" {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestConvertMetaAdsAdset(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	row := sink.ConvertMetaAdsAdset("act_1", kafkamodels.RawMetaAdsAdset{
		ID:   "adset-1",
		Name: "Adset",
		Targeting: &kafkamodels.RawMetaAdsAdsetTargeting{
			AgeMin: 18,
			AgeMax: 34,
			GeoLocations: &struct {
				Countries []string `json:"countries"`
			}{Countries: []string{"US", "CA"}},
		},
	})
	if row.AdsetID != "adset-1" || len(row.TargetingCountries) != 2 {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestConvertMetaAdsCampaignInsight(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	row := sink.ConvertMetaAdsCampaignInsight("act_1", kafkamodels.RawMetaAdsInsightRow{
		CampaignID:   "camp-1",
		CampaignName: "Campaign",
		Objective:    "OUTCOME_TRAFFIC",
		DateStart:    "2025-01-01",
		Spend:        "10.5",
		Impressions:  "100",
		Clicks:       "7",
		Actions: []kafkamodels.RawMetaAdsAction{
			{ActionType: "purchase", Value: "2"},
		},
	})
	if row.Spend != 10.5 || row.ActionsPurchase != 2 || row.Clicks != 7 {
		t.Fatalf("unexpected row: %+v", row)
	}
}

func TestBulkInsertWrappersEmptySlice(t *testing.T) {
	sink := newTestSink(&mockClickHouseClient{})
	if err := sink.BulkInsertMetaAdsCampaigns(context.Background(), nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := sink.BulkInsertMetaAdsAds(context.Background(), []*clickhousemodels.MetaAdsAd{}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
