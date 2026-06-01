package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== RunService Tests ==================

func TestRunService_BasicFlow(t *testing.T) {
	log := logger.New("error")

	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	consumer := NewMockKafkaConsumer()

	deps := &FetcherDependencies{
		Producer:  producer,
		Consumer:  consumer,
		MongoRepo: mongoRepo,
		GMBClient: nil,
		Log:       log,
	}

	cfg := FetcherConfig{
		MaxWorkers: 1,
		QueueSize:  10,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)
	if err != nil {
		t.Fatalf("RunService failed: %v", err)
	}
}

// ================== FetcherConfig Tests ==================

func TestDefaultFetcherConfig(t *testing.T) {
	cfg := DefaultFetcherConfig()

	if cfg.MaxWorkers != maxWorkers {
		t.Fatalf("expected MaxWorkers %d, got %d", maxWorkers, cfg.MaxWorkers)
	}
	if cfg.QueueSize != workChanSize {
		t.Fatalf("expected QueueSize %d, got %d", workChanSize, cfg.QueueSize)
	}
	if cfg.MaxConcurrentAccounts != maxConcurrentAccounts {
		t.Fatalf("expected MaxConcurrentAccounts %d, got %d", maxConcurrentAccounts, cfg.MaxConcurrentAccounts)
	}
}

// ================== FetcherMetrics Tests ==================

func TestGetMetrics(t *testing.T) {
	metrics := &FetcherMetrics{}

	atomic.StoreUint64(&metrics.AccountsDispatched, 10)
	atomic.StoreUint64(&metrics.AccountsProcessed, 8)
	atomic.StoreUint64(&metrics.ProcessingErrors, 2)

	result := GetMetrics(metrics)

	if result["accounts_dispatched"] != 10 {
		t.Fatalf("expected accounts_dispatched=10, got %d", result["accounts_dispatched"])
	}
	if result["accounts_processed"] != 8 {
		t.Fatalf("expected accounts_processed=8, got %d", result["accounts_processed"])
	}
	if result["processing_errors"] != 2 {
		t.Fatalf("expected processing_errors=2, got %d", result["processing_errors"])
	}
}

func TestGetMetrics_ZeroValues(t *testing.T) {
	metrics := &FetcherMetrics{}
	result := GetMetrics(metrics)

	for key, val := range result {
		if val != 0 {
			t.Fatalf("expected %s=0, got %d", key, val)
		}
	}
}

// ================== FetcherDependencies Tests ==================

func TestFetcherDependencies_Creation(t *testing.T) {
	log := logger.New("error")

	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()
	consumer := NewMockKafkaConsumer()

	deps := &FetcherDependencies{
		Producer:  producer,
		Consumer:  consumer,
		MongoRepo: mongoRepo,
		Log:       log,
	}

	if deps.Producer == nil {
		t.Fatal("expected Producer to be set")
	}
	if deps.Consumer == nil {
		t.Fatal("expected Consumer to be set")
	}
	if deps.MongoRepo == nil {
		t.Fatal("expected MongoRepo to be set")
	}
	if deps.Log == nil {
		t.Fatal("expected Log to be set")
	}
}
