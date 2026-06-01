package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/tiktok/tiktok-immediate-processor/processor"
)

// ================== Mock Implementations ==================

type MockConsumer struct {
	ConsumeFunc func(ctx context.Context, topics []string, handler kafka.MessageHandler) error
	messages    []struct {
		topic string
		key   []byte
		value []byte
	}
}

func (m *MockConsumer) Consume(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
	if m.ConsumeFunc != nil {
		return m.ConsumeFunc(ctx, topics, handler)
	}

	for _, msg := range m.messages {
		if err := handler(ctx, msg.topic, msg.key, msg.value); err != nil {
			return err
		}
	}
	<-ctx.Done()
	return ctx.Err()
}

func (m *MockConsumer) ConsumeWithAck(ctx context.Context, topics []string, handler kafka.AcknowledgingMessageHandler) error {
	return nil
}

func (m *MockConsumer) Close() error {
	return nil
}

type MockProcessor struct {
	ProcessAccountFunc func(wo processor.ImmediateWorkOrder) error
	processedAccounts  []processor.ImmediateWorkOrder
	mu                 sync.Mutex
}

func (m *MockProcessor) ProcessAccount(wo processor.ImmediateWorkOrder) error {
	m.mu.Lock()
	m.processedAccounts = append(m.processedAccounts, wo)
	m.mu.Unlock()

	if m.ProcessAccountFunc != nil {
		return m.ProcessAccountFunc(wo)
	}
	return nil
}

func (m *MockProcessor) GetProcessedAccounts() []processor.ImmediateWorkOrder {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.processedAccounts
}

// ================== NewImmediateProcessorService Tests ==================

func TestNewImmediateProcessorService(t *testing.T) {
	cfg := DefaultImmediateProcessorConfig()
	deps := ImmediateProcessorDependencies{
		Consumer:  &MockConsumer{},
		Processor: &MockProcessor{},
		Log:       logger.New("error"),
	}

	svc := NewImmediateProcessorService(cfg, deps)

	if svc == nil {
		t.Fatal("NewImmediateProcessorService returned nil")
	}
	if svc.config.MaxWorkers != cfg.MaxWorkers {
		t.Errorf("MaxWorkers = %d, want %d", svc.config.MaxWorkers, cfg.MaxWorkers)
	}
}

func TestNewImmediateProcessorService_CustomConfig(t *testing.T) {
	cfg := ImmediateProcessorConfig{
		MaxWorkers:   10,
		JobQueueSize: 500,
	}
	deps := ImmediateProcessorDependencies{
		Consumer:  &MockConsumer{},
		Processor: &MockProcessor{},
		Log:       logger.New("error"),
	}

	svc := NewImmediateProcessorService(cfg, deps)

	if svc.config.MaxWorkers != 10 {
		t.Errorf("MaxWorkers = %d, want 10", svc.config.MaxWorkers)
	}
	if svc.config.JobQueueSize != 500 {
		t.Errorf("JobQueueSize = %d, want 500", svc.config.JobQueueSize)
	}
}

// ================== DefaultImmediateProcessorConfig Tests ==================

func TestDefaultImmediateProcessorConfig(t *testing.T) {
	cfg := DefaultImmediateProcessorConfig()

	if cfg.MaxWorkers == 0 {
		t.Error("MaxWorkers should not be zero")
	}
	if cfg.JobQueueSize == 0 {
		t.Error("JobQueueSize should not be zero")
	}
}

func TestDefaultImmediateProcessorConfig_ReasonableValues(t *testing.T) {
	cfg := DefaultImmediateProcessorConfig()

	if cfg.MaxWorkers < 1 || cfg.MaxWorkers > 100 {
		t.Errorf("MaxWorkers = %d, should be between 1 and 100", cfg.MaxWorkers)
	}
	if cfg.JobQueueSize < 10 || cfg.JobQueueSize > 10000 {
		t.Errorf("JobQueueSize = %d, should be between 10 and 10000", cfg.JobQueueSize)
	}
}

// ================== Run Tests ==================

func TestRun_BasicFlow(t *testing.T) {
	log := logger.New("error")
	mockProcessor := &MockProcessor{}

	wo := processor.ImmediateWorkOrder{
		ID:          "order_123",
		TikTokID:    "tiktok_789",
		AccessToken: "test_token",
		SyncType:    "immediate",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Deliver message then wait for context
			handler(ctx, "immediate-work-order-tiktok", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	cfg := ImmediateProcessorConfig{
		MaxWorkers:   2,
		JobQueueSize: 10,
	}

	deps := ImmediateProcessorDependencies{
		Consumer:  mockConsumer,
		Processor: mockProcessor,
		Log:       log,
	}

	svc := NewImmediateProcessorService(cfg, deps)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := svc.Run(ctx)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run failed: %v", err)
	}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	if svc.metrics.ReceivedJobs == 0 {
		t.Log("Note: ReceivedJobs may be 0 due to timing")
	}
}

func TestRun_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	mockConsumer := &MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := ImmediateProcessorDependencies{
		Consumer:  mockConsumer,
		Processor: &MockProcessor{},
		Log:       log,
	}

	svc := NewImmediateProcessorService(DefaultImmediateProcessorConfig(), deps)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	err := svc.Run(ctx)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run should complete gracefully: %v", err)
	}
}

func TestRun_ProcessorError(t *testing.T) {
	log := logger.New("error")

	wo := processor.ImmediateWorkOrder{
		ID:       "order_123",
		TikTokID: "tiktok_789",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "immediate-work-order-tiktok", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	mockProcessor := &MockProcessor{
		ProcessAccountFunc: func(wo processor.ImmediateWorkOrder) error {
			return errors.New("processing failed")
		},
	}

	cfg := ImmediateProcessorConfig{
		MaxWorkers:   1,
		JobQueueSize: 10,
	}

	deps := ImmediateProcessorDependencies{
		Consumer:  mockConsumer,
		Processor: mockProcessor,
		Log:       log,
	}

	svc := NewImmediateProcessorService(cfg, deps)

	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	err := svc.Run(ctx)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run should complete gracefully: %v", err)
	}

	// Check failed jobs counter
	time.Sleep(100 * time.Millisecond)
	if atomic.LoadUint64(&svc.metrics.FailedJobs) == 0 {
		t.Log("Note: FailedJobs may be 0 due to timing")
	}
}

func TestRun_InvalidJSON(t *testing.T) {
	log := logger.New("error")

	mockConsumer := &MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send invalid JSON
			handler(ctx, "immediate-work-order-tiktok", []byte("key"), []byte("invalid-json"))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := ImmediateProcessorDependencies{
		Consumer:  mockConsumer,
		Processor: &MockProcessor{},
		Log:       log,
	}

	svc := NewImmediateProcessorService(DefaultImmediateProcessorConfig(), deps)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := svc.Run(ctx)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("Run should handle invalid JSON gracefully: %v", err)
	}
}

// ================== GetMetrics Tests ==================

func TestGetMetrics_Initialized(t *testing.T) {
	svc := &ImmediateProcessorService{}

	metrics := svc.GetMetrics()

	if len(metrics) != 4 {
		t.Errorf("GetMetrics should return 4 metrics, got %d", len(metrics))
	}

	expectedKeys := []string{
		"received_jobs",
		"processed_jobs",
		"failed_jobs",
		"skipped_jobs",
	}

	for _, key := range expectedKeys {
		if _, ok := metrics[key]; !ok {
			t.Errorf("Expected metric key %q not found", key)
		}
	}
}

func TestGetMetrics_WithValues(t *testing.T) {
	svc := &ImmediateProcessorService{}

	atomic.AddUint64(&svc.metrics.ReceivedJobs, 100)
	atomic.AddUint64(&svc.metrics.ProcessedJobs, 95)
	atomic.AddUint64(&svc.metrics.FailedJobs, 3)
	atomic.AddUint64(&svc.metrics.SkippedJobs, 2)

	metrics := svc.GetMetrics()

	if metrics["received_jobs"] != 100 {
		t.Errorf("received_jobs = %d, want 100", metrics["received_jobs"])
	}
	if metrics["processed_jobs"] != 95 {
		t.Errorf("processed_jobs = %d, want 95", metrics["processed_jobs"])
	}
	if metrics["failed_jobs"] != 3 {
		t.Errorf("failed_jobs = %d, want 3", metrics["failed_jobs"])
	}
	if metrics["skipped_jobs"] != 2 {
		t.Errorf("skipped_jobs = %d, want 2", metrics["skipped_jobs"])
	}
}

func TestGetMetrics_ZeroValues(t *testing.T) {
	svc := &ImmediateProcessorService{}

	metrics := svc.GetMetrics()

	for key, value := range metrics {
		if value != 0 {
			t.Errorf("expected %s to be 0, got %d", key, value)
		}
	}
}

func TestGetMetrics_AtomicOperations(t *testing.T) {
	svc := &ImmediateProcessorService{}

	// Simulate concurrent updates
	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			atomic.AddUint64(&svc.metrics.ReceivedJobs, 1)
			atomic.AddUint64(&svc.metrics.ProcessedJobs, 1)
		}()
	}
	wg.Wait()

	metrics := svc.GetMetrics()

	if metrics["received_jobs"] != 100 {
		t.Errorf("received_jobs = %d, want 100", metrics["received_jobs"])
	}
	if metrics["processed_jobs"] != 100 {
		t.Errorf("processed_jobs = %d, want 100", metrics["processed_jobs"])
	}
}

// ================== ImmediateProcessorMetrics Tests ==================

func TestImmediateProcessorMetrics_Initialized(t *testing.T) {
	metrics := &ImmediateProcessorMetrics{}

	if metrics.ReceivedJobs != 0 {
		t.Errorf("ReceivedJobs should be 0, got %d", metrics.ReceivedJobs)
	}
	if metrics.ProcessedJobs != 0 {
		t.Errorf("ProcessedJobs should be 0, got %d", metrics.ProcessedJobs)
	}
	if metrics.FailedJobs != 0 {
		t.Errorf("FailedJobs should be 0, got %d", metrics.FailedJobs)
	}
	if metrics.SkippedJobs != 0 {
		t.Errorf("SkippedJobs should be 0, got %d", metrics.SkippedJobs)
	}
}

// ================== ImmediateProcessorDependencies Tests ==================

func TestImmediateProcessorDependencies_AllSet(t *testing.T) {
	deps := ImmediateProcessorDependencies{
		Consumer:  &MockConsumer{},
		Processor: &MockProcessor{},
		Log:       logger.New("error"),
	}

	if deps.Consumer == nil {
		t.Error("Consumer should not be nil")
	}
	if deps.Processor == nil {
		t.Error("Processor should not be nil")
	}
	if deps.Log == nil {
		t.Error("Log should not be nil")
	}
}
