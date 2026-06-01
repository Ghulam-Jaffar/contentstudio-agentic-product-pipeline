package kafka

import (
	"encoding/json"
	"testing"
	"time"
)

func TestMetaAdsWorkOrderJSONRoundTrip(t *testing.T) {
	wo := MetaAdsWorkOrder{
		MongoID:            "mongo-1",
		PlatformIdentifier: "act_123",
		AccountID:          "123",
		AccessToken:        "access",
		LongAccessToken:    "long",
		WorkspaceID:        "ws-1",
		UserID:             "user-1",
		SyncType:           "scheduled",
		StartDate:          "2025-01-01",
		EndDate:            "2025-01-31",
	}

	data, err := json.Marshal(wo)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded MetaAdsWorkOrder
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.PlatformIdentifier != wo.PlatformIdentifier {
		t.Fatalf("unexpected platform identifier: %q", decoded.PlatformIdentifier)
	}
}

func TestMetaAdsBatchWorkOrderJSONRoundTrip(t *testing.T) {
	batch := MetaAdsBatchWorkOrder{
		BatchID: "batch-1",
		Accounts: []MetaAdsWorkOrder{
			{MongoID: "1", PlatformIdentifier: "act_1"},
		},
	}

	data, err := json.Marshal(batch)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	var decoded MetaAdsBatchWorkOrder
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if decoded.BatchID != batch.BatchID || len(decoded.Accounts) != 1 {
		t.Fatalf("unexpected decoded batch: %+v", decoded)
	}
}

func TestMetaAdsAPITimeJSON(t *testing.T) {
	var ts MetaAdsAPITime
	if err := json.Unmarshal([]byte(`"2025-07-21T14:59:17+0500"`), &ts); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ts.Time.Location() != time.UTC {
		t.Fatalf("expected UTC, got %v", ts.Time.Location())
	}
	data, err := json.Marshal(ts)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if string(data) == "null" {
		t.Fatal("expected non-null timestamp")
	}
}

func TestRawMetaAdsAdsetUnmarshalJSON(t *testing.T) {
	raw := []byte(`{
		"id":"adset-1",
		"name":"Adset",
		"campaign_id":"camp-1",
		"status":"ACTIVE",
		"effective_status":"ACTIVE",
		"targeting":{
			"age_min":18,
			"age_max":34,
			"geo_locations":{"countries":["US","CA"]}
		}
	}`)

	var adset RawMetaAdsAdset
	if err := json.Unmarshal(raw, &adset); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if adset.Targeting == nil {
		t.Fatal("expected targeting to be populated")
	}
	if len(adset.Targeting.GeoLocations.Countries) != 2 {
		t.Fatalf("unexpected countries: %+v", adset.Targeting.GeoLocations)
	}
	if len(adset.RawTargeting) == 0 {
		t.Fatal("expected raw targeting to be captured")
	}
}
