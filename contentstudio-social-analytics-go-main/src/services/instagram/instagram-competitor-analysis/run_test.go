package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== CompetitorAnalysisConfig Tests ==================

func TestDefaultCompetitorAnalysisConfig(t *testing.T) {
	cfg := DefaultCompetitorAnalysisConfig()

	if cfg.WorkersPerPool != WorkersPerPool {
		t.Errorf("WorkersPerPool = %d, want %d", cfg.WorkersPerPool, WorkersPerPool)
	}
}

// ================== NewCompetitorAnalysisService Tests ==================

func TestNewCompetitorAnalysisService(t *testing.T) {
	cfg := CompetitorAnalysisConfig{
		WorkersPerPool: 5,
	}

	log := logger.New("error")
	deps := CompetitorAnalysisDependencies{Log: log}

	svc := NewCompetitorAnalysisService(cfg, deps)

	if svc == nil {
		t.Fatal("NewCompetitorAnalysisService returned nil")
	}
	if svc.config.WorkersPerPool != 5 {
		t.Errorf("WorkersPerPool = %d, want 5", svc.config.WorkersPerPool)
	}
}

// ================== CompetitorAnalysisService.Run Tests ==================

func TestCompetitorAnalysisService_Run_ContextCancel(t *testing.T) {
	cfg := CompetitorAnalysisConfig{
		WorkersPerPool: 2,
	}

	log := logger.New("error")

	realtimeConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	batchConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := CompetitorAnalysisDependencies{
		RealtimeConsumer: realtimeConsumer,
		BatchConsumer:    batchConsumer,
		Log:              log,
	}

	svc := NewCompetitorAnalysisService(cfg, deps)

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

func TestCompetitorAnalysisService_Run_ProcessesRealtimeJobs(t *testing.T) {
	cfg := CompetitorAnalysisConfig{
		WorkersPerPool: 2,
	}

	log := logger.New("error")

	wo := kafkaModels.CompetitorWorkOrder{
		ReportID: "report123",
		PageID:   "page456",
		Channel:  "instagram",
		Mode:     "incremental",
	}
	woJSON, _ := json.Marshal(wo)

	realtimeConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-instagram", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	batchConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := CompetitorAnalysisDependencies{
		RealtimeConsumer: realtimeConsumer,
		BatchConsumer:    batchConsumer,
		Log:              log,
	}

	svc := NewCompetitorAnalysisService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.RealtimeJobsReceived < 1 {
		t.Errorf("RealtimeJobsReceived = %d, want >= 1", metrics.RealtimeJobsReceived)
	}

	cancel()
	<-done
}

func TestCompetitorAnalysisService_Run_ProcessesBatchJobs(t *testing.T) {
	cfg := CompetitorAnalysisConfig{
		WorkersPerPool: 2,
	}

	log := logger.New("error")

	wo := kafkaModels.CompetitorWorkOrder{
		ReportID: "report789",
		PageID:   "page012",
		Channel:  "instagram",
		Mode:     "full_sync",
	}
	woJSON, _ := json.Marshal(wo)

	realtimeConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	batchConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-instagram-batch", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := CompetitorAnalysisDependencies{
		RealtimeConsumer: realtimeConsumer,
		BatchConsumer:    batchConsumer,
		Log:              log,
	}

	svc := NewCompetitorAnalysisService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.BatchJobsReceived < 1 {
		t.Errorf("BatchJobsReceived = %d, want >= 1", metrics.BatchJobsReceived)
	}

	cancel()
	<-done
}

func TestCompetitorAnalysisService_Run_IgnoresWrongChannel(t *testing.T) {
	cfg := CompetitorAnalysisConfig{
		WorkersPerPool: 2,
	}

	log := logger.New("error")

	wo := kafkaModels.CompetitorWorkOrder{
		ReportID: "report123",
		PageID:   "page456",
		Channel:  "facebook", // Wrong channel
		Mode:     "incremental",
	}
	woJSON, _ := json.Marshal(wo)

	realtimeConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-instagram", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	batchConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := CompetitorAnalysisDependencies{
		RealtimeConsumer: realtimeConsumer,
		BatchConsumer:    batchConsumer,
		Log:              log,
	}

	svc := NewCompetitorAnalysisService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.RealtimeJobsReceived != 0 {
		t.Errorf("RealtimeJobsReceived = %d, want 0 (wrong channel)", metrics.RealtimeJobsReceived)
	}

	cancel()
	<-done
}

// ================== GetMetrics Tests ==================

func TestCompetitorAnalysisService_GetMetrics_Initial(t *testing.T) {
	cfg := CompetitorAnalysisConfig{}
	log := logger.New("error")
	deps := CompetitorAnalysisDependencies{Log: log}

	svc := NewCompetitorAnalysisService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.RealtimeJobsReceived != 0 {
		t.Errorf("RealtimeJobsReceived = %d, want 0", metrics.RealtimeJobsReceived)
	}
	if metrics.BatchJobsReceived != 0 {
		t.Errorf("BatchJobsReceived = %d, want 0", metrics.BatchJobsReceived)
	}
	if metrics.RealtimeJobsProcessed != 0 {
		t.Errorf("RealtimeJobsProcessed = %d, want 0", metrics.RealtimeJobsProcessed)
	}
	if metrics.BatchJobsProcessed != 0 {
		t.Errorf("BatchJobsProcessed = %d, want 0", metrics.BatchJobsProcessed)
	}
}

// ================== CompetitorAnalysisMetrics Tests ==================

func TestCompetitorAnalysisMetrics_Struct(t *testing.T) {
	m := CompetitorAnalysisMetrics{
		RealtimeJobsReceived:  100,
		BatchJobsReceived:     50,
		RealtimeJobsProcessed: 95,
		BatchJobsProcessed:    48,
	}

	if m.RealtimeJobsReceived != 100 {
		t.Errorf("RealtimeJobsReceived = %d, want 100", m.RealtimeJobsReceived)
	}
	if m.BatchJobsReceived != 50 {
		t.Errorf("BatchJobsReceived = %d, want 50", m.BatchJobsReceived)
	}
	if m.RealtimeJobsProcessed != 95 {
		t.Errorf("RealtimeJobsProcessed = %d, want 95", m.RealtimeJobsProcessed)
	}
	if m.BatchJobsProcessed != 48 {
		t.Errorf("BatchJobsProcessed = %d, want 48", m.BatchJobsProcessed)
	}
}
