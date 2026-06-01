package main

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/instagram/instagram-immediate-processor/processor"
)

// ================== ImmediateProcessorConfig Tests ==================

func TestDefaultImmediateProcessorConfig(t *testing.T) {
	cfg := DefaultImmediateProcessorConfig()

	if cfg.MaxWorkers != maxImmediateWorkers {
		t.Errorf("MaxWorkers = %d, want %d", cfg.MaxWorkers, maxImmediateWorkers)
	}
	if cfg.JobQueueSize != jobQueueSize {
		t.Errorf("JobQueueSize = %d, want %d", cfg.JobQueueSize, jobQueueSize)
	}
}

// ================== NewImmediateProcessorService Tests ==================

func TestNewImmediateProcessorService(t *testing.T) {
	cfg := ImmediateProcessorConfig{
		MaxWorkers:   5,
		JobQueueSize: 50,
	}

	log := logger.New("error")
	deps := ImmediateProcessorDependencies{Log: log}

	svc := NewImmediateProcessorService(cfg, deps)

	if svc == nil {
		t.Fatal("NewImmediateProcessorService returned nil")
	}
	if svc.config.MaxWorkers != 5 {
		t.Errorf("MaxWorkers = %d, want 5", svc.config.MaxWorkers)
	}
	if svc.config.JobQueueSize != 50 {
		t.Errorf("JobQueueSize = %d, want 50", svc.config.JobQueueSize)
	}
}

// ================== ImmediateProcessorService.Run Tests ==================

func TestImmediateProcessorService_Run_ContextCancel(t *testing.T) {
	cfg := ImmediateProcessorConfig{
		MaxWorkers:   2,
		JobQueueSize: 10,
	}

	log := logger.New("error")

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := ImmediateProcessorDependencies{
		Consumer:  mockConsumer,
		Processor: nil,
		Log:       log,
	}

	svc := NewImmediateProcessorService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancel")
	}
}

func TestImmediateProcessorService_Run_ReceivesJobs(t *testing.T) {
	cfg := ImmediateProcessorConfig{
		MaxWorkers:   1,
		JobQueueSize: 10,
	}

	log := logger.New("error")

	wo := processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "ig456",
		WorkspaceID: "ws789",
		SyncType:    "immediate",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "immediate-work-order-instagram", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := ImmediateProcessorDependencies{
		Consumer:  mockConsumer,
		Processor: &processor.Processor{}, // Will fail but that's OK
		Log:       log,
	}

	svc := NewImmediateProcessorService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.ReceivedJobs < 1 {
		t.Errorf("ReceivedJobs = %d, want >= 1", metrics.ReceivedJobs)
	}

	cancel()
	<-done
}

func TestImmediateProcessorService_Run_InvalidJSON(t *testing.T) {
	cfg := ImmediateProcessorConfig{
		MaxWorkers:   1,
		JobQueueSize: 10,
	}

	log := logger.New("error")

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "immediate-work-order-instagram", []byte("key"), []byte("invalid json"))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := ImmediateProcessorDependencies{
		Consumer:  mockConsumer,
		Processor: nil,
		Log:       log,
	}

	svc := NewImmediateProcessorService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit")
	}

	// Invalid JSON shouldn't increment received count
	metrics := svc.GetMetrics()
	if metrics.ReceivedJobs != 0 {
		t.Errorf("ReceivedJobs = %d, want 0 (invalid JSON)", metrics.ReceivedJobs)
	}
}

// ================== GetMetrics Tests ==================

func TestImmediateProcessorService_GetMetrics_Initial(t *testing.T) {
	cfg := ImmediateProcessorConfig{}
	log := logger.New("error")
	deps := ImmediateProcessorDependencies{Log: log}

	svc := NewImmediateProcessorService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.ReceivedJobs != 0 {
		t.Errorf("ReceivedJobs = %d, want 0", metrics.ReceivedJobs)
	}
	if metrics.ProcessedJobs != 0 {
		t.Errorf("ProcessedJobs = %d, want 0", metrics.ProcessedJobs)
	}
	if metrics.SkippedJobs != 0 {
		t.Errorf("SkippedJobs = %d, want 0", metrics.SkippedJobs)
	}
	if metrics.FailedJobs != 0 {
		t.Errorf("FailedJobs = %d, want 0", metrics.FailedJobs)
	}
}

// ================== ImmediateProcessorMetrics Tests ==================

func TestImmediateProcessorMetrics_Struct(t *testing.T) {
	m := ImmediateProcessorMetrics{
		ReceivedJobs:  100,
		ProcessedJobs: 90,
		SkippedJobs:   5,
		FailedJobs:    5,
	}

	if m.ReceivedJobs != 100 {
		t.Errorf("ReceivedJobs = %d, want 100", m.ReceivedJobs)
	}
	if m.ProcessedJobs != 90 {
		t.Errorf("ProcessedJobs = %d, want 90", m.ProcessedJobs)
	}
	if m.SkippedJobs != 5 {
		t.Errorf("SkippedJobs = %d, want 5", m.SkippedJobs)
	}
	if m.FailedJobs != 5 {
		t.Errorf("FailedJobs = %d, want 5", m.FailedJobs)
	}
}

// ================== Worker Tests ==================

func TestImmediateProcessorService_Worker_SkipsDuplicate(t *testing.T) {
	cfg := ImmediateProcessorConfig{
		MaxWorkers:   2,
		JobQueueSize: 10,
	}

	log := logger.New("error")

	var jobsSent int32
	wo := processor.WorkOrder{AccountID: "ig123"}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send same job twice
			handler(ctx, "immediate-work-order-instagram", []byte("key"), woJSON)
			atomic.AddInt32(&jobsSent, 1)
			handler(ctx, "immediate-work-order-instagram", []byte("key"), woJSON)
			atomic.AddInt32(&jobsSent, 1)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := ImmediateProcessorDependencies{
		Consumer:  mockConsumer,
		Processor: &processor.Processor{},
		Log:       log,
	}

	svc := NewImmediateProcessorService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(200 * time.Millisecond)
	cancel()
	<-done

	metrics := svc.GetMetrics()
	// Both jobs should be received
	if metrics.ReceivedJobs != 2 {
		t.Errorf("ReceivedJobs = %d, want 2", metrics.ReceivedJobs)
	}
}
