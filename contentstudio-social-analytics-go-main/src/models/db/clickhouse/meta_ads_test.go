package clickhouse

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMetaAdsAccountInfoJSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	model := MetaAdsAccountInfo{
		AccountID:     "act_1",
		Name:          "Account",
		Currency:      "USD",
		AccountStatus: 1,
		TimezoneName:  "UTC",
		BusinessID:    "biz-1",
		BusinessName:  "Business",
		AmountSpent:   "123.45",
		Balance:       "10",
		SpendCap:      "1000",
		CreatedTime:   now,
		InsertedAt:    now,
	}
	data, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded MetaAdsAccountInfo
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.AccountID != model.AccountID || decoded.BusinessName != "Business" {
		t.Fatalf("unexpected decoded model: %+v", decoded)
	}
}

func TestMetaAdsCampaignJSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	model := MetaAdsCampaign{
		AccountID:       "act_1",
		CampaignID:      "camp-1",
		Name:            "Campaign",
		Status:          "ACTIVE",
		EffectiveStatus: "ACTIVE",
		Objective:       "OUTCOME_TRAFFIC",
		DailyBudget:     "10",
		LifetimeBudget:  "20",
		BudgetRemaining: "5",
		StartTime:       now,
		StopTime:        now,
		CreatedTime:     now,
		UpdatedTime:     now,
		InsertedAt:      now,
	}
	data, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded MetaAdsCampaign
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.CampaignID != model.CampaignID {
		t.Fatalf("unexpected decoded campaign: %+v", decoded)
	}
}

func TestMetaAdsCampaignInsightsJSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	model := MetaAdsCampaignInsights{
		AccountID:    "act_1",
		CampaignID:   "camp-1",
		CampaignName: "Campaign",
		Objective:    "OUTCOME_TRAFFIC",
		InsightsDate: now,
		Spend:        10.5,
		Impressions:  100,
		Clicks:       5,
		InsertedAt:   now,
	}
	data, err := json.Marshal(model)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded MetaAdsCampaignInsights
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.Spend != 10.5 || decoded.Clicks != 5 {
		t.Fatalf("unexpected decoded insight: %+v", decoded)
	}
}
