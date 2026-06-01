package main

import (
	"context"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== ServiceConfig Tests ==================

func TestDefaultServiceConfig(t *testing.T) {
	cfg := DefaultServiceConfig()

	if cfg.BatchProcessorsPerType != batchProcessorsPerType {
		t.Errorf("BatchProcessorsPerType = %d, want %d", cfg.BatchProcessorsPerType, batchProcessorsPerType)
	}
	if cfg.MaxBatchSize != maxBatchSize {
		t.Errorf("MaxBatchSize = %d, want %d", cfg.MaxBatchSize, maxBatchSize)
	}
	if cfg.BatchTimeout != batchTimeout {
		t.Errorf("BatchTimeout = %v, want %v", cfg.BatchTimeout, batchTimeout)
	}
}

// ================== NewService Tests ==================

func TestNewService(t *testing.T) {
	cfg := ServiceConfig{
		BatchProcessorsPerType: 2,
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
	}

	log := logger.New("error")
	deps := ServiceDependencies{Log: log}

	svc := NewService(cfg, deps)

	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.config.BatchProcessorsPerType != 2 {
		t.Errorf("BatchProcessorsPerType = %d, want 2", svc.config.BatchProcessorsPerType)
	}
	if svc.batchCollectors == nil {
		t.Error("batchCollectors is nil")
	}
	if svc.batchCollectors.posts == nil {
		t.Error("posts channel is nil")
	}
	if svc.batchCollectors.insights == nil {
		t.Error("insights channel is nil")
	}
}

// ================== Service.Run Tests ==================

func TestService_Run_ContextCancel(t *testing.T) {
	cfg := ServiceConfig{
		BatchProcessorsPerType: 0, // Don't start batch processors to avoid nil sink panic
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
	}

	log := logger.New("error")

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := ServiceDependencies{
		Consumer: mockConsumer,
		Sink:     nil,
		Log:      log,
	}

	svc := NewService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	// Cancel immediately
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

func TestService_Run_UnknownTopic(t *testing.T) {
	cfg := ServiceConfig{
		BatchProcessorsPerType: 0, // Don't start batch processors to avoid nil sink panic
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
	}

	log := logger.New("error")

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send message to unknown topic - should be ignored
			handler(ctx, "unknown-topic", []byte("key"), []byte("{}"))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := ServiceDependencies{
		Consumer: mockConsumer,
		Sink:     nil,
		Log:      log,
	}

	svc := NewService(cfg, deps)

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

	// No messages should have been picked
	metrics := svc.GetMetrics()
	if metrics.PickedPosts != 0 {
		t.Errorf("PickedPosts = %d, want 0", metrics.PickedPosts)
	}
	if metrics.PickedInsights != 0 {
		t.Errorf("PickedInsights = %d, want 0", metrics.PickedInsights)
	}
}

// ================== GetMetrics Tests ==================

func TestService_GetMetrics_Initial(t *testing.T) {
	cfg := ServiceConfig{
		BatchProcessorsPerType: 1,
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
	}

	log := logger.New("error")
	deps := ServiceDependencies{Log: log}

	svc := NewService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.PickedPosts != 0 {
		t.Errorf("PickedPosts = %d, want 0", metrics.PickedPosts)
	}
	if metrics.PickedInsights != 0 {
		t.Errorf("PickedInsights = %d, want 0", metrics.PickedInsights)
	}
	if metrics.InsertedPosts != 0 {
		t.Errorf("InsertedPosts = %d, want 0", metrics.InsertedPosts)
	}
	if metrics.InsertedInsights != 0 {
		t.Errorf("InsertedInsights = %d, want 0", metrics.InsertedInsights)
	}
}

// ================== GetBatchCollectors Tests ==================

func TestService_GetBatchCollectors(t *testing.T) {
	cfg := ServiceConfig{
		BatchProcessorsPerType: 1,
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
	}

	log := logger.New("error")
	deps := ServiceDependencies{Log: log}

	svc := NewService(cfg, deps)

	bc := svc.GetBatchCollectors()
	if bc == nil {
		t.Fatal("GetBatchCollectors returned nil")
	}
	if bc.posts == nil {
		t.Error("posts channel is nil")
	}
	if bc.insights == nil {
		t.Error("insights channel is nil")
	}
}

// ================== ServiceMetrics Tests ==================

func TestServiceMetrics_Struct(t *testing.T) {
	m := ServiceMetrics{
		PickedPosts:      10,
		PickedInsights:   20,
		InsertedPosts:    5,
		InsertedInsights: 15,
	}

	if m.PickedPosts != 10 {
		t.Errorf("PickedPosts = %d, want 10", m.PickedPosts)
	}
	if m.PickedInsights != 20 {
		t.Errorf("PickedInsights = %d, want 20", m.PickedInsights)
	}
	if m.InsertedPosts != 5 {
		t.Errorf("InsertedPosts = %d, want 5", m.InsertedPosts)
	}
	if m.InsertedInsights != 15 {
		t.Errorf("InsertedInsights = %d, want 15", m.InsertedInsights)
	}
}
