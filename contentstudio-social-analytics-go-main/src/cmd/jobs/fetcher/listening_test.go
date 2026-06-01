package fetcher

import (
	"testing"
	"time"

	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/scheduler"
)

func TestBuildListeningWorkOrder_UsesInitialWindowForNewTopic(t *testing.T) {
	createdAt := time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC)
	now := time.Date(2026, 4, 11, 9, 30, 0, 0, time.UTC)

	topic := &mongomodels.ListeningTopic{
		TopicID:            "topic-1",
		WorkspaceID:        "ws-1",
		IncludeKeywords:    []string{"apple"},
		EnabledPlatforms:   []string{"twitter"},
		LastFetchedCursors: map[string]string{"twitter": "cursor-1"},
		CreatedAt:          createdAt,
	}

	order := scheduler.BuildRecurringWorkOrder(topic, now)

	wantFromDate := createdAt.Add(-30 * 24 * time.Hour)
	if !order.FromDate.Equal(wantFromDate) {
		t.Fatalf("expected initial from_date %s, got %s", wantFromDate, order.FromDate)
	}
	if !order.ToDate.Equal(now.UTC()) {
		t.Fatalf("expected to_date %s, got %s", now.UTC(), order.ToDate)
	}
	if got := order.Cursors["twitter"]; got != "cursor-1" {
		t.Fatalf("expected cursor to be copied, got %q", got)
	}
	if order.SyncType != "incremental" {
		t.Fatalf("expected sync_type incremental, got %q", order.SyncType)
	}
}

func TestBuildListeningWorkOrder_UsesLastFetchedAtForIncrementalRuns(t *testing.T) {
	lastFetchedAt := time.Date(2026, 4, 9, 15, 45, 0, 0, time.UTC)
	now := time.Date(2026, 4, 10, 16, 0, 0, 0, time.UTC)

	topic := &mongomodels.ListeningTopic{
		TopicID:          "topic-1",
		WorkspaceID:      "ws-1",
		IncludeKeywords:  []string{"apple"},
		EnabledPlatforms: []string{"twitter"},
		CreatedAt:        time.Date(2026, 3, 1, 12, 0, 0, 0, time.UTC),
		LastFetchedAt:    lastFetchedAt,
	}

	order := scheduler.BuildRecurringWorkOrder(topic, now)

	if !order.FromDate.Equal(lastFetchedAt) {
		t.Fatalf("expected incremental from_date %s, got %s", lastFetchedAt, order.FromDate)
	}
	if !order.ToDate.Equal(now.UTC()) {
		t.Fatalf("expected to_date %s, got %s", now.UTC(), order.ToDate)
	}
	if order.SyncType != "incremental" {
		t.Fatalf("expected sync_type incremental, got %q", order.SyncType)
	}
}
