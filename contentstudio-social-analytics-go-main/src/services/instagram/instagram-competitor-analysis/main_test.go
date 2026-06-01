package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if PlatformTypeInstagram != "instagram" {
		t.Errorf("PlatformTypeInstagram = %q, want %q", PlatformTypeInstagram, "instagram")
	}
	if TokenQueueInstagram != "instagram_valid_token_set" {
		t.Errorf("TokenQueueInstagram = %q, want %q", TokenQueueInstagram, "instagram_valid_token_set")
	}
	if WorkersPerPool != 10 {
		t.Errorf("WorkersPerPool = %d, want 10", WorkersPerPool)
	}
}

// ================== CompetitorJob Tests ==================

func TestCompetitorJob_Struct(t *testing.T) {
	job := CompetitorJob{
		ReportID: "report123",
		PageID:   "page456",
		CompID:   "comp789",
		Mode:     apiModels.SyncModeIncremental,
	}

	if job.ReportID != "report123" {
		t.Errorf("ReportID = %q, want %q", job.ReportID, "report123")
	}
	if job.PageID != "page456" {
		t.Errorf("PageID = %q, want %q", job.PageID, "page456")
	}
	if job.CompID != "comp789" {
		t.Errorf("CompID = %q, want %q", job.CompID, "comp789")
	}
}

// ================== toString Tests ==================

func TestToString(t *testing.T) {
	log := logger.New("error")

	tests := []struct {
		name  string
		input interface{}
		want  string
	}{
		{"string", "hello", "hello"},
		{"int", 123, "123"},
		{"int32", int32(456), "456"},
		{"int64", int64(789), "789"},
		{"unsupported", []int{1, 2, 3}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toString(tt.input, log)
			if got != tt.want {
				t.Errorf("toString(%v) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ================== consumeRealtime Tests ==================

func TestConsumeRealtime_ValidMessage(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)

	wo := kafkaModels.CompetitorWorkOrder{
		ReportID: "report123",
		PageID:   "page456",
		Channel:  "instagram",
		Mode:     "incremental",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-instagram", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go consumeRealtime(ctx, mockConsumer, jobs, log)

	select {
	case job := <-jobs:
		if job.ReportID != "report123" {
			t.Errorf("ReportID = %q, want %q", job.ReportID, "report123")
		}
		if job.PageID != "page456" {
			t.Errorf("PageID = %q, want %q", job.PageID, "page456")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for job")
	}

	cancel()
}

func TestConsumeRealtime_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-instagram", []byte("key"), []byte("invalid json"))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go consumeRealtime(ctx, mockConsumer, jobs, log)

	time.Sleep(100 * time.Millisecond)
	cancel()

	// Should not have added any jobs
	if len(jobs) != 0 {
		t.Errorf("expected empty jobs channel, got %d", len(jobs))
	}
}

func TestConsumeRealtime_WrongChannel(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)

	wo := kafkaModels.CompetitorWorkOrder{
		ReportID: "report123",
		PageID:   "page456",
		Channel:  "facebook", // Wrong channel
		Mode:     "incremental",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-instagram", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go consumeRealtime(ctx, mockConsumer, jobs, log)

	time.Sleep(100 * time.Millisecond)
	cancel()

	// Should not have added any jobs (wrong channel)
	if len(jobs) != 0 {
		t.Errorf("expected empty jobs channel, got %d", len(jobs))
	}
}

// ================== consumeBatch Tests ==================

func TestConsumeBatch_ValidMessage(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)

	wo := kafkaModels.CompetitorWorkOrder{
		ReportID: "report789",
		PageID:   "page012",
		Channel:  "instagram",
		Mode:     "full_sync",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-instagram-batch", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go consumeBatch(ctx, mockConsumer, jobs, log)

	select {
	case job := <-jobs:
		if job.ReportID != "report789" {
			t.Errorf("ReportID = %q, want %q", job.ReportID, "report789")
		}
		if job.PageID != "page012" {
			t.Errorf("PageID = %q, want %q", job.PageID, "page012")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for job")
	}

	cancel()
}

func TestConsumeBatch_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-instagram-batch", []byte("key"), []byte("invalid"))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go consumeBatch(ctx, mockConsumer, jobs, log)

	time.Sleep(100 * time.Millisecond)
	cancel()

	if len(jobs) != 0 {
		t.Errorf("expected empty jobs channel, got %d", len(jobs))
	}
}

func TestConsumeBatch_WrongChannel(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)

	wo := kafkaModels.CompetitorWorkOrder{
		ReportID: "report123",
		PageID:   "page456",
		Channel:  "linkedin", // Wrong channel
		Mode:     "incremental",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-instagram-batch", []byte("key"), woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go consumeBatch(ctx, mockConsumer, jobs, log)

	time.Sleep(100 * time.Millisecond)
	cancel()

	if len(jobs) != 0 {
		t.Errorf("expected empty jobs channel, got %d", len(jobs))
	}
}

// ================== Context Cancellation Tests ==================

func TestConsumeRealtime_ContextCancel(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		consumeRealtime(ctx, mockConsumer, jobs, log)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("consumeRealtime did not exit after context cancel")
	}
}

func TestConsumeBatch_ContextCancel(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		consumeBatch(ctx, mockConsumer, jobs, log)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("consumeBatch did not exit after context cancel")
	}
}

// ================== Multiple Message Tests ==================

func TestConsumeRealtime_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 100)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			for i := 0; i < 5; i++ {
				wo := kafkaModels.CompetitorWorkOrder{
					ReportID: "report" + string(rune('0'+i)),
					PageID:   "page" + string(rune('0'+i)),
					Channel:  "instagram",
					Mode:     "incremental",
				}
				woJSON, _ := json.Marshal(wo)
				handler(ctx, "competitor-work-order-instagram", []byte("key"), woJSON)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go consumeRealtime(ctx, mockConsumer, jobs, log)

	time.Sleep(200 * time.Millisecond)
	cancel()

	if len(jobs) != 5 {
		t.Errorf("expected 5 jobs, got %d", len(jobs))
	}
}

func TestConsumeBatch_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 100)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			for i := 0; i < 5; i++ {
				wo := kafkaModels.CompetitorWorkOrder{
					ReportID: "report" + string(rune('0'+i)),
					PageID:   "page" + string(rune('0'+i)),
					Channel:  "instagram",
					Mode:     "full_sync",
				}
				woJSON, _ := json.Marshal(wo)
				handler(ctx, "competitor-work-order-instagram-batch", []byte("key"), woJSON)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go consumeBatch(ctx, mockConsumer, jobs, log)

	time.Sleep(200 * time.Millisecond)
	cancel()

	if len(jobs) != 5 {
		t.Errorf("expected 5 jobs, got %d", len(jobs))
	}
}

// ================== Worker Pool Simulation Tests ==================

func TestWorkerPool_ConcurrentProcessing(t *testing.T) {
	var processedCount int32
	jobs := make(chan CompetitorJob, 100)

	var wg sync.WaitGroup
	numWorkers := 3

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobs {
				atomic.AddInt32(&processedCount, 1)
			}
		}()
	}

	// Send jobs
	for i := 0; i < 30; i++ {
		jobs <- CompetitorJob{ReportID: "report", PageID: "page"}
	}

	close(jobs)
	wg.Wait()

	if atomic.LoadInt32(&processedCount) != 30 {
		t.Errorf("expected 30 processed, got %d", processedCount)
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	var shutdownCount int32
	jobs := make(chan CompetitorJob, 10)

	var wg sync.WaitGroup
	numWorkers := 5

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			defer atomic.AddInt32(&shutdownCount, 1)
			for range jobs {
				// Process
			}
		}()
	}

	close(jobs)
	wg.Wait()

	if atomic.LoadInt32(&shutdownCount) != int32(numWorkers) {
		t.Errorf("expected %d workers shutdown, got %d", numWorkers, shutdownCount)
	}
}

// ================== Atomic Counter Tests ==================

func TestAtomicCounters(t *testing.T) {
	var counter int32

	var wg sync.WaitGroup
	numGoroutines := 100
	incrementsPerGoroutine := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				atomic.AddInt32(&counter, 1)
			}
		}()
	}

	wg.Wait()

	expected := int32(numGoroutines * incrementsPerGoroutine)
	if atomic.LoadInt32(&counter) != expected {
		t.Errorf("counter = %d, want %d", counter, expected)
	}
}
