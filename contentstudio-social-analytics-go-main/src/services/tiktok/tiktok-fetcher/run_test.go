package main

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== RunService Tests ==================

func TestRunService_BasicFlow(t *testing.T) {
	log := logger.New("error")

	batch := &kafkamodels.TikTokBatchWorkOrder{
		Accounts: []kafkamodels.TikTokAccountWorkOrder{
			{
				TikTokID:    "tiktok_user_1",
				AccessToken: "token_1",
			},
		},
	}
	batchJSON, _ := json.Marshal(batch)

	videoConsumer := NewMockKafkaConsumerWithMessages([]MockMessage{
		{Topic: workOrdersTopic, Key: []byte("batch_1"), Value: batchJSON},
	})
	insightsConsumer := NewMockKafkaConsumerWithMessages([]MockMessage{})
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()

	deps := &FetcherDependencies{
		VideoConsumer:    videoConsumer,
		InsightsConsumer: insightsConsumer,
		Producer:         producer,
		MongoRepo:        mongoRepo,
		TikTokClient:     nil, // Not used in basic flow
		Log:              log,
	}

	cfg := FetcherConfig{
		MaxVideoWorkers:         1,
		MaxInsightsWorkers:      1,
		VideoQueueSize:          10,
		InsightsQueueSize:       10,
		TimestampUpdateChanSize: 10,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService failed: %v", err)
	}
}

func TestRunService_EmptyConsumers(t *testing.T) {
	log := logger.New("error")

	videoConsumer := NewMockKafkaConsumerWithMessages([]MockMessage{})
	insightsConsumer := NewMockKafkaConsumerWithMessages([]MockMessage{})
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()

	deps := &FetcherDependencies{
		VideoConsumer:    videoConsumer,
		InsightsConsumer: insightsConsumer,
		Producer:         producer,
		MongoRepo:        mongoRepo,
		Log:              log,
	}

	cfg := DefaultFetcherConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService should complete: %v", err)
	}
}

func TestRunService_WithInsightsBatch(t *testing.T) {
	log := logger.New("error")

	batch := &kafkamodels.TikTokBatchWorkOrder{
		Accounts: []kafkamodels.TikTokAccountWorkOrder{
			{
				TikTokID:    "tiktok_user_1",
				AccessToken: "token_1",
			},
			{
				TikTokID:    "tiktok_user_2",
				AccessToken: "token_2",
			},
		},
	}
	batchJSON, _ := json.Marshal(batch)

	videoConsumer := NewMockKafkaConsumerWithMessages([]MockMessage{})
	insightsConsumer := NewMockKafkaConsumerWithMessages([]MockMessage{
		{Topic: workOrdersTopic, Key: []byte("batch_1"), Value: batchJSON},
	})
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()

	deps := &FetcherDependencies{
		VideoConsumer:    videoConsumer,
		InsightsConsumer: insightsConsumer,
		Producer:         producer,
		MongoRepo:        mongoRepo,
		Log:              log,
	}

	cfg := FetcherConfig{
		MaxVideoWorkers:         1,
		MaxInsightsWorkers:      1,
		VideoQueueSize:          10,
		InsightsQueueSize:       10,
		TimestampUpdateChanSize: 10,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService failed: %v", err)
	}
}

// ================== FetcherConfig Tests ==================

func TestDefaultFetcherConfig(t *testing.T) {
	cfg := DefaultFetcherConfig()

	if cfg.MaxVideoWorkers == 0 {
		t.Error("MaxVideoWorkers should not be zero")
	}
	if cfg.MaxInsightsWorkers == 0 {
		t.Error("MaxInsightsWorkers should not be zero")
	}
	if cfg.VideoQueueSize == 0 {
		t.Error("VideoQueueSize should not be zero")
	}
	if cfg.InsightsQueueSize == 0 {
		t.Error("InsightsQueueSize should not be zero")
	}
}

func TestDefaultFetcherConfig_ReasonableValues(t *testing.T) {
	cfg := DefaultFetcherConfig()

	if cfg.MaxVideoWorkers < 1 || cfg.MaxVideoWorkers > 100 {
		t.Errorf("MaxVideoWorkers = %d, should be between 1 and 100", cfg.MaxVideoWorkers)
	}
	if cfg.MaxInsightsWorkers < 1 || cfg.MaxInsightsWorkers > 100 {
		t.Errorf("MaxInsightsWorkers = %d, should be between 1 and 100", cfg.MaxInsightsWorkers)
	}
}

// ================== FetcherMetrics Tests ==================

func TestFetcherMetrics_Initialized(t *testing.T) {
	metrics := &FetcherMetrics{}

	if metrics.VideoBatchesReceived != 0 {
		t.Errorf("VideoBatchesReceived should be 0, got %d", metrics.VideoBatchesReceived)
	}
	if metrics.InsightsBatchesReceived != 0 {
		t.Errorf("InsightsBatchesReceived should be 0, got %d", metrics.InsightsBatchesReceived)
	}
	if metrics.VideoAccountsProcessed != 0 {
		t.Errorf("VideoAccountsProcessed should be 0, got %d", metrics.VideoAccountsProcessed)
	}
	if metrics.ProducerErrors != 0 {
		t.Errorf("ProducerErrors should be 0, got %d", metrics.ProducerErrors)
	}
}

func TestFetcherMetrics_Atomic(t *testing.T) {
	metrics := &FetcherMetrics{}

	atomic.AddUint64(&metrics.VideoBatchesReceived, 10)
	atomic.AddUint64(&metrics.InsightsBatchesReceived, 5)
	atomic.AddUint64(&metrics.VideoAccountsProcessed, 100)
	atomic.AddUint64(&metrics.InsightsAccountsProcessed, 50)
	atomic.AddUint64(&metrics.TotalBatchesProduced, 15)
	atomic.AddUint64(&metrics.ProducerErrors, 2)

	if atomic.LoadUint64(&metrics.VideoBatchesReceived) != 10 {
		t.Errorf("VideoBatchesReceived = %d, want 10", metrics.VideoBatchesReceived)
	}
	if atomic.LoadUint64(&metrics.InsightsBatchesReceived) != 5 {
		t.Errorf("InsightsBatchesReceived = %d, want 5", metrics.InsightsBatchesReceived)
	}
	if atomic.LoadUint64(&metrics.VideoAccountsProcessed) != 100 {
		t.Errorf("VideoAccountsProcessed = %d, want 100", metrics.VideoAccountsProcessed)
	}
}

// ================== GetMetrics Tests ==================

func TestGetMetrics_Initialized(t *testing.T) {
	metrics := &FetcherMetrics{}

	result := GetMetrics(metrics)

	if len(result) != 6 {
		t.Errorf("GetMetrics should return 6 metrics, got %d", len(result))
	}

	expectedKeys := []string{
		"video_batches_received",
		"insights_batches_received",
		"video_accounts_processed",
		"insights_accounts_processed",
		"total_batches_produced",
		"producer_errors",
	}

	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("Expected metric key %q not found", key)
		}
	}
}

func TestGetMetrics_WithValues(t *testing.T) {
	metrics := &FetcherMetrics{}
	atomic.AddUint64(&metrics.VideoBatchesReceived, 10)
	atomic.AddUint64(&metrics.InsightsBatchesReceived, 5)
	atomic.AddUint64(&metrics.VideoAccountsProcessed, 100)
	atomic.AddUint64(&metrics.InsightsAccountsProcessed, 50)
	atomic.AddUint64(&metrics.TotalBatchesProduced, 15)
	atomic.AddUint64(&metrics.ProducerErrors, 2)

	result := GetMetrics(metrics)

	if result["video_batches_received"] != 10 {
		t.Errorf("video_batches_received = %d, want 10", result["video_batches_received"])
	}
	if result["insights_batches_received"] != 5 {
		t.Errorf("insights_batches_received = %d, want 5", result["insights_batches_received"])
	}
	if result["video_accounts_processed"] != 100 {
		t.Errorf("video_accounts_processed = %d, want 100", result["video_accounts_processed"])
	}
	if result["producer_errors"] != 2 {
		t.Errorf("producer_errors = %d, want 2", result["producer_errors"])
	}
}

func TestGetMetrics_ZeroValues(t *testing.T) {
	metrics := &FetcherMetrics{}

	result := GetMetrics(metrics)

	for key, value := range result {
		if value != 0 {
			t.Errorf("expected %s to be 0, got %d", key, value)
		}
	}
}

// ================== FetcherDependencies Tests ==================

func TestFetcherDependencies_AllSet(t *testing.T) {
	deps := &FetcherDependencies{
		VideoConsumer:    NewMockKafkaConsumerWithMessages(nil),
		InsightsConsumer: NewMockKafkaConsumerWithMessages(nil),
		Producer:         NewMockKafkaProducer(),
		MongoRepo:        NewMockUnifiedSocialRepository(),
		TikTokClient:     nil,
		Log:              logger.New("error"),
	}

	if deps.VideoConsumer == nil {
		t.Error("VideoConsumer should not be nil")
	}
	if deps.InsightsConsumer == nil {
		t.Error("InsightsConsumer should not be nil")
	}
	if deps.Producer == nil {
		t.Error("Producer should not be nil")
	}
	if deps.MongoRepo == nil {
		t.Error("MongoRepo should not be nil")
	}
	if deps.Log == nil {
		t.Error("Log should not be nil")
	}
}

// ================== Invalid JSON Tests ==================

func TestRunService_InvalidJSON(t *testing.T) {
	log := logger.New("error")

	videoConsumer := &kafka.MockConsumerWithMessages{
		Messages: []kafka.MockMessage{
			{Topic: workOrdersTopic, Key: []byte("batch_1"), Value: []byte("invalid-json")},
		},
	}
	insightsConsumer := NewMockKafkaConsumerWithMessages([]MockMessage{})
	producer := NewMockKafkaProducer()
	mongoRepo := NewMockUnifiedSocialRepository()

	deps := &FetcherDependencies{
		VideoConsumer:    videoConsumer,
		InsightsConsumer: insightsConsumer,
		Producer:         producer,
		MongoRepo:        mongoRepo,
		Log:              log,
	}

	cfg := FetcherConfig{
		MaxVideoWorkers:         1,
		MaxInsightsWorkers:      1,
		VideoQueueSize:          10,
		InsightsQueueSize:       10,
		TimestampUpdateChanSize: 10,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService should handle invalid JSON gracefully: %v", err)
	}
}
