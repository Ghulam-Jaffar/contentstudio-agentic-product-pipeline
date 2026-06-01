package meta_ads

import (
	"context"
	"io"
	"testing"

	"github.com/rs/zerolog"

	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/meta_ads"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/meta_ads"
)

func TestHelpers(t *testing.T) {
	if got := percentChange(120, 100); got != 20 {
		t.Fatalf("expected 20, got %v", got)
	}
	if got := percentChange(0, 0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
	if got := metricValue(12, 10); got.Change != 20 {
		t.Fatalf("unexpected metric value: %+v", got)
	}
	if got := safeRate(10, 0); got != 0 {
		t.Fatalf("expected 0, got %v", got)
	}
	if got := normalizeMetric("bad"); got != "spend" {
		t.Fatalf("expected spend, got %s", got)
	}
	if page, perPage := normalizePagination(0, 500); page != 1 || perPage != 10 {
		t.Fatalf("unexpected pagination: %d %d", page, perPage)
	}
	if got := metricValueForRow(repo.AgeGenderRow{Spend: 4, Impressions: 100, Reach: 20, Clicks: 5, CTR: 1.2, CPM: 2.3, CPC: 3.4, Frequency: 1.5}, "clicks"); got != 5 {
		t.Fatalf("unexpected metric value for row: %v", got)
	}
}

func TestAIInsightsMethodsWithoutService(t *testing.T) {
	svc := &MetaAdsService{logger: zerolog.New(io.Discard)}
	req := &types.MetaAdsRequest{WorkspaceID: "ws1", AccountID: "act_1", StartDate: "2025-01-01", EndDate: "2025-01-31"}

	if _, err := svc.GetAIInsightsSummary(context.Background(), req); err == nil {
		t.Fatal("expected error when AI service is nil")
	}
	if _, err := svc.GetAIInsightsDetailed(context.Background(), req); err == nil {
		t.Fatal("expected error when AI service is nil")
	}
}

func TestGetSummaryValidation(t *testing.T) {
	svc := &MetaAdsService{logger: zerolog.New(io.Discard)}
	if _, err := svc.GetSummary(context.Background(), &types.MetaAdsRequest{}); err == nil {
		t.Fatal("expected validation error")
	}
}
