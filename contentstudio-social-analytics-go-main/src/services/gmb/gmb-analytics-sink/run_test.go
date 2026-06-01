package main

import (
	"sync/atomic"
	"testing"
)

func TestDefaultServiceConfig(t *testing.T) {
	cfg := DefaultServiceConfig()
	if cfg.DataParserWorkers != 5 {
		t.Fatalf("expected 5 data parser workers, got %d", cfg.DataParserWorkers)
	}
	if cfg.BatchProcessorsPerType != batchProcessorsPerType {
		t.Fatalf("expected %d batch processors per type, got %d", batchProcessorsPerType, cfg.BatchProcessorsPerType)
	}
	if cfg.MaxBatchSize != maxBatchSize {
		t.Fatalf("expected %d max batch size, got %d", maxBatchSize, cfg.MaxBatchSize)
	}
	if cfg.IdleTimeoutMinutes != int(idleTimeout.Minutes()) {
		t.Fatalf("expected %d idle timeout minutes, got %d", int(idleTimeout.Minutes()), cfg.IdleTimeoutMinutes)
	}
	if cfg.MessageChanSize != messageChanSize {
		t.Fatalf("expected %d message chan size, got %d", messageChanSize, cfg.MessageChanSize)
	}
}

func TestServiceMetrics_AtomicOperations(t *testing.T) {
	m := &ServiceMetrics{}
	atomic.AddUint64(&m.PickedCount, 10)
	atomic.AddUint64(&m.ParsedDailyMetrics, 5)
	atomic.AddUint64(&m.ParsedMediaAssets, 3)
	atomic.AddUint64(&m.ParsedSearchKeywords, 2)
	atomic.AddUint64(&m.ParsedLocalPosts, 4)
	atomic.AddUint64(&m.ParsedReviews, 6)
	atomic.AddUint64(&m.InsertedDailyMetrics, 5)
	atomic.AddUint64(&m.InsertedMediaAssets, 3)
	atomic.AddUint64(&m.InsertedSearchKeywords, 2)
	atomic.AddUint64(&m.InsertedLocalPosts, 4)
	atomic.AddUint64(&m.InsertedReviews, 6)

	if atomic.LoadUint64(&m.PickedCount) != 10 {
		t.Fatal("unexpected PickedCount")
	}
	if atomic.LoadUint64(&m.ParsedDailyMetrics) != 5 {
		t.Fatal("unexpected ParsedDailyMetrics")
	}
}

func TestGetMetrics(t *testing.T) {
	m := &ServiceMetrics{}
	atomic.StoreUint64(&m.PickedCount, 100)
	atomic.StoreUint64(&m.ParsedDailyMetrics, 50)
	atomic.StoreUint64(&m.ParsedMediaAssets, 30)
	atomic.StoreUint64(&m.ParsedSearchKeywords, 20)
	atomic.StoreUint64(&m.ParsedLocalPosts, 40)
	atomic.StoreUint64(&m.ParsedReviews, 60)
	atomic.StoreUint64(&m.InsertedDailyMetrics, 45)
	atomic.StoreUint64(&m.InsertedMediaAssets, 25)
	atomic.StoreUint64(&m.InsertedSearchKeywords, 15)
	atomic.StoreUint64(&m.InsertedLocalPosts, 35)
	atomic.StoreUint64(&m.InsertedReviews, 55)

	result := GetMetrics(m)

	if result["picked"] != 100 {
		t.Fatalf("expected picked=100, got %d", result["picked"])
	}
	if result["parsed_daily_metrics"] != 50 {
		t.Fatalf("expected parsed_daily_metrics=50, got %d", result["parsed_daily_metrics"])
	}
	if result["parsed_media_assets"] != 30 {
		t.Fatalf("expected parsed_media_assets=30, got %d", result["parsed_media_assets"])
	}
	if result["parsed_search_keywords"] != 20 {
		t.Fatalf("expected parsed_search_keywords=20, got %d", result["parsed_search_keywords"])
	}
	if result["parsed_local_posts"] != 40 {
		t.Fatalf("expected parsed_local_posts=40, got %d", result["parsed_local_posts"])
	}
	if result["parsed_reviews"] != 60 {
		t.Fatalf("expected parsed_reviews=60, got %d", result["parsed_reviews"])
	}
	if result["inserted_daily_metrics"] != 45 {
		t.Fatalf("expected inserted_daily_metrics=45, got %d", result["inserted_daily_metrics"])
	}
	if result["inserted_media_assets"] != 25 {
		t.Fatalf("expected inserted_media_assets=25, got %d", result["inserted_media_assets"])
	}
	if result["inserted_search_keywords"] != 15 {
		t.Fatalf("expected inserted_search_keywords=15, got %d", result["inserted_search_keywords"])
	}
	if result["inserted_local_posts"] != 35 {
		t.Fatalf("expected inserted_local_posts=35, got %d", result["inserted_local_posts"])
	}
	if result["inserted_reviews"] != 55 {
		t.Fatalf("expected inserted_reviews=55, got %d", result["inserted_reviews"])
	}

	// Should have exactly 11 keys
	if len(result) != 11 {
		t.Fatalf("expected 11 metrics, got %d", len(result))
	}
}

func TestServiceConfig_CustomValues(t *testing.T) {
	cfg := ServiceConfig{
		DataParserWorkers:      10,
		BatchProcessorsPerType: 5,
		MaxBatchSize:           5000,
		BatchTimeoutSeconds:    30,
		IdleTimeoutMinutes:     10,
		MessageChanSize:        100000,
	}
	if cfg.DataParserWorkers != 10 {
		t.Fatalf("expected 10, got %d", cfg.DataParserWorkers)
	}
	if cfg.BatchProcessorsPerType != 5 {
		t.Fatalf("expected 5, got %d", cfg.BatchProcessorsPerType)
	}
}
