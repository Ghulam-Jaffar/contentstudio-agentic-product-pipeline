package campaign_label

import (
	"testing"
	"time"

	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/campaign_label"
	"go.mongodb.org/mongo-driver/bson"
)

func TestFilterMappingsByAccounts(t *testing.T) {
	docs := []mongoPostMapping{
		{CampaignID: "camp-1", PlatformID: FlexString("123"), PostedIDs: FlexStringSlice{"post-1"}},
		{CampaignID: "camp-2", PlatformID: FlexString("456"), PostedIDs: FlexStringSlice{"post-2"}},
	}

	filtered := filterMappingsByAccounts(docs, []string{"456"}, false)
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered doc, got %d", len(filtered))
	}
	if filtered[0].CampaignID != "camp-2" {
		t.Fatalf("expected camp-2, got %q", filtered[0].CampaignID)
	}
}

func TestFilterMappingsByAccounts_IncludeAll(t *testing.T) {
	docs := []mongoPostMapping{
		{CampaignID: "camp-1", PlatformID: FlexString("123")},
		{CampaignID: "camp-2", PlatformID: FlexString("456")},
	}

	filtered := filterMappingsByAccounts(docs, []string{"456"}, true)
	if len(filtered) != 2 {
		t.Fatalf("expected all docs when includeAll is true, got %d", len(filtered))
	}
}

// TestComputeDiffAndPct checks the difference/percentage helper for integer values.
func TestComputeDiffAndPct(t *testing.T) {
	tests := []struct {
		name        string
		current     int64
		previous    int64
		expectedDif int64
		expectedPct float64
	}{
		{name: "growth", current: 15, previous: 10, expectedDif: 5, expectedPct: 50.0},
		{name: "decline", current: 5, previous: 10, expectedDif: -5, expectedPct: -50.0},
		{name: "zero previous non-zero current", current: 10, previous: 0, expectedDif: 10, expectedPct: 100.0},
		{name: "both zero", current: 0, previous: 0, expectedDif: 0, expectedPct: 0.0},
		{name: "equal", current: 10, previous: 10, expectedDif: 0, expectedPct: 0.0},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp := &types.SummaryResponse{
				Difference: make(map[string]interface{}),
				Percentage: make(map[string]interface{}),
			}
			computeDiffAndPct(resp, "metric", tc.current, tc.previous)

			dif, ok := resp.Difference["metric"].(int64)
			if !ok {
				t.Fatalf("expected int64 diff, got %T", resp.Difference["metric"])
			}
			if dif != tc.expectedDif {
				t.Fatalf("expected diff=%d, got %d", tc.expectedDif, dif)
			}

			pct, ok := resp.Percentage["metric"].(float64)
			if !ok {
				t.Fatalf("expected float64 pct, got %T", resp.Percentage["metric"])
			}
			if pct != tc.expectedPct {
				t.Fatalf("expected pct=%f, got %f", tc.expectedPct, pct)
			}
		})
	}
}

// TestComputeDiffAndPctFloat checks the difference/percentage helper for float values.
func TestComputeDiffAndPctFloat(t *testing.T) {
	resp := &types.SummaryResponse{
		Difference: make(map[string]interface{}),
		Percentage: make(map[string]interface{}),
	}
	computeDiffAndPctFloat(resp, "rate", 0.5, 0.25)

	dif, ok := resp.Difference["rate"].(float64)
	if !ok {
		t.Fatalf("expected float64 diff, got %T", resp.Difference["rate"])
	}
	if dif != 0.25 {
		t.Fatalf("expected diff=0.25, got %f", dif)
	}
}

func TestRoundFloatValues(t *testing.T) {
	input := map[string]interface{}{
		"engagement":      0.3333333333333333,
		"engagement_rate": 0.1111111111111111,
		"impressions":     int32(3),
		"nested": []interface{}{
			0.6666666666666666,
			map[string]interface{}{"value": 1.2345},
		},
	}

	out, ok := roundFloatValues(input).(map[string]interface{})
	if !ok {
		t.Fatalf("expected map[string]interface{}, got %T", roundFloatValues(input))
	}

	if got := out["engagement"].(float64); got != 0.33 {
		t.Fatalf("expected engagement=0.33, got %v", got)
	}
	if got := out["engagement_rate"].(float64); got != 0.11 {
		t.Fatalf("expected engagement_rate=0.11, got %v", got)
	}
	if got := out["impressions"].(int32); got != 3 {
		t.Fatalf("expected impressions to remain 3, got %v", got)
	}

	nested := out["nested"].([]interface{})
	if got := nested[0].(float64); got != 0.67 {
		t.Fatalf("expected nested[0]=0.67, got %v", got)
	}
	if got := nested[1].(map[string]interface{})["value"].(float64); got != 1.23 {
		t.Fatalf("expected nested[1].value=1.23, got %v", got)
	}
}

// TestPrevDate checks the previous date range calculation.
func TestPrevDate(t *testing.T) {
	start := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 3, 1, 0, 0, 0, 0, time.UTC)
	prevStart, prevEnd := prevDate(start, end)

	expectedPrevStart := "2025-01-02"
	expectedPrevEnd := "2025-01-31"

	if prevStart.Format("2006-01-02") != expectedPrevStart {
		t.Fatalf("expected prevStart=%s, got %s", expectedPrevStart, prevStart.Format("2006-01-02"))
	}
	if prevEnd.Format("2006-01-02") != expectedPrevEnd {
		t.Fatalf("expected prevEnd=%s, got %s", expectedPrevEnd, prevEnd.Format("2006-01-02"))
	}
}

func TestMongoPostMappingDecodesNumericPlatformID(t *testing.T) {
	raw, err := bson.Marshal(bson.M{
		"campaign_id":   "campaign-1",
		"platform_id":   int64(123456789),
		"platform":      "instagram",
		"platform_type": "social",
		"posted_ids":    bson.A{int64(987654321), "abc123"},
	})
	if err != nil {
		t.Fatalf("marshal mongoPostMapping fixture: %v", err)
	}

	var decoded mongoPostMapping
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal mongoPostMapping: %v", err)
	}

	if got, want := string(decoded.PlatformID), "123456789"; got != want {
		t.Fatalf("platform_id mismatch: got %q want %q", got, want)
	}

	if len(decoded.PostedIDs) != 2 {
		t.Fatalf("posted_ids length mismatch: got %d want 2", len(decoded.PostedIDs))
	}

	if got, want := decoded.PostedIDs[0], "987654321"; got != want {
		t.Fatalf("posted_ids[0] mismatch: got %q want %q", got, want)
	}
}

func TestPostingFallbackDecodesNumericIDs(t *testing.T) {
	raw, err := bson.Marshal(bson.M{
		"platform_id":   int64(42),
		"platform":      "facebook",
		"platform_type": "social",
		"posted_id":     int64(99),
	})
	if err != nil {
		t.Fatalf("marshal posting fixture: %v", err)
	}

	var decoded struct {
		PlatformID   FlexString `bson:"platform_id"`
		Platform     string     `bson:"platform"`
		PlatformType string     `bson:"platform_type"`
		PostedID     FlexString `bson:"posted_id"`
	}
	if err := bson.Unmarshal(raw, &decoded); err != nil {
		t.Fatalf("unmarshal posting fixture: %v", err)
	}

	if got, want := string(decoded.PlatformID), "42"; got != want {
		t.Fatalf("platform_id mismatch: got %q want %q", got, want)
	}

	if got, want := string(decoded.PostedID), "99"; got != want {
		t.Fatalf("posted_id mismatch: got %q want %q", got, want)
	}
}
