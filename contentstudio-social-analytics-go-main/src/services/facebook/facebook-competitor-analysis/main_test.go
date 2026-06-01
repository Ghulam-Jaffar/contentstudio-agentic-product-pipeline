package main

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== CompetitorJob Tests ==================

func TestCompetitorJob_Struct(t *testing.T) {
	job := CompetitorJob{
		ReportID: "report123",
		PageID:   "page456",
		CompID:   "comp789",
		Mode:     api.SyncMode("full_sync"),
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
	if job.Mode != api.SyncMode("full_sync") {
		t.Errorf("Mode = %q, want %q", job.Mode, "full_sync")
	}
}

// ================== toString Tests ==================

func TestToString(t *testing.T) {
	log := logger.New("error")

	tests := []struct {
		name     string
		input    interface{}
		expected string
	}{
		{"string input", "test123", "test123"},
		{"int input", 123, "123"},
		{"int32 input", int32(456), "456"},
		{"int64 input", int64(789), "789"},
		{"unsupported type", struct{}{}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := toString(tt.input, log)
			if result != tt.expected {
				t.Errorf("toString(%v) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

// ================== consumeRealtime Tests ==================

func TestConsumeRealtime_ValidMessage(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)
	ctx, cancel := context.WithCancel(context.Background())

	// Create work order
	wo := kafkaModels.CompetitorWorkOrder{
		Channel:  "facebook",
		ReportID: "report123",
		PageID:   "page456",
		Mode:     "full_sync",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Simulate receiving a message
			err := handler(ctx, "competitor-work-order-facebook", nil, woJSON)
			if err != nil {
				return err
			}
			// Wait for context cancellation
			<-ctx.Done()
			return ctx.Err()
		},
	}

	go consumeRealtime(ctx, mockConsumer, jobs, log)

	// Wait for job to be processed
	select {
	case job := <-jobs:
		if job.ReportID != "report123" {
			t.Errorf("ReportID = %q, want %q", job.ReportID, "report123")
		}
		if job.PageID != "page456" {
			t.Errorf("PageID = %q, want %q", job.PageID, "page456")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for job")
	}

	cancel()
}

func TestConsumeRealtime_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send invalid JSON
			handler(ctx, "competitor-work-order-facebook", nil, []byte("invalid json"))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	go consumeRealtime(ctx, mockConsumer, jobs, log)

	// Should not receive any job
	select {
	case <-jobs:
		t.Fatal("should not receive job for invalid JSON")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}
}

func TestConsumeRealtime_WrongChannel(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wo := kafkaModels.CompetitorWorkOrder{
		Channel:  "instagram", // Wrong channel
		ReportID: "report123",
		PageID:   "page456",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-facebook", nil, woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	go consumeRealtime(ctx, mockConsumer, jobs, log)

	// Should not receive job for wrong channel
	select {
	case <-jobs:
		t.Fatal("should not receive job for wrong channel")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}
}

// ================== consumeBatch Tests ==================

func TestConsumeBatch_ValidMessage(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)
	ctx, cancel := context.WithCancel(context.Background())

	wo := kafkaModels.CompetitorWorkOrder{
		Channel:  "facebook",
		ReportID: "report789",
		PageID:   "page012",
		Mode:     "incremental",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-facebook-batch", nil, woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	go consumeBatch(ctx, mockConsumer, jobs, log)

	select {
	case job := <-jobs:
		if job.ReportID != "report789" {
			t.Errorf("ReportID = %q, want %q", job.ReportID, "report789")
		}
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for job")
	}

	cancel()
}

// ================== Worker Pool Tests ==================

func TestWorkerPool_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	jobs := make(chan CompetitorJob, 10)
	close(jobs)

	var wg sync.WaitGroup
	wg.Add(1)

	// Simulate a simple worker that exits on context cancellation
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case _, ok := <-jobs:
				if !ok {
					return
				}
			}
		}
	}()

	wg.Wait()
	// Should complete without hanging
}

// ================== Channel Pipeline Tests ==================

func TestChannelPipeline_Flow(t *testing.T) {
	// Test that data flows through the pipeline correctly
	jobs := make(chan CompetitorJob, 10)
	results := make(chan string, 10)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(1)

	// Simple worker that processes jobs
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case job, ok := <-jobs:
				if !ok {
					return
				}
				results <- job.PageID
			}
		}
	}()

	// Send jobs
	jobs <- CompetitorJob{PageID: "page1"}
	jobs <- CompetitorJob{PageID: "page2"}

	// Check results
	result1 := <-results
	result2 := <-results

	if result1 != "page1" {
		t.Errorf("result1 = %q, want %q", result1, "page1")
	}
	if result2 != "page2" {
		t.Errorf("result2 = %q, want %q", result2, "page2")
	}

	close(jobs)
	wg.Wait()
}

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if PlatformTypeFacebook != "facebook" {
		t.Errorf("PlatformTypeFacebook = %q, want %q", PlatformTypeFacebook, "facebook")
	}
	if TokenQueueFacebook != "facebook_valid_token_set" {
		t.Errorf("TokenQueueFacebook = %q, want %q", TokenQueueFacebook, "facebook_valid_token_set")
	}
	if WorkersPerPool != 10 {
		t.Errorf("WorkersPerPool = %d, want %d", WorkersPerPool, 10)
	}
}

// ================== consumeBatch Additional Tests ==================

func TestConsumeBatch_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-facebook-batch", nil, []byte("invalid json"))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	go consumeBatch(ctx, mockConsumer, jobs, log)

	select {
	case <-jobs:
		t.Fatal("should not receive job for invalid JSON")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}
}

func TestConsumeBatch_WrongChannel(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	wo := kafkaModels.CompetitorWorkOrder{
		Channel:  "instagram",
		ReportID: "report123",
		PageID:   "page456",
	}
	woJSON, _ := json.Marshal(wo)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, "competitor-work-order-facebook-batch", nil, woJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	go consumeBatch(ctx, mockConsumer, jobs, log)

	select {
	case <-jobs:
		t.Fatal("should not receive job for wrong channel")
	case <-time.After(100 * time.Millisecond):
		// Expected
	}
}

func TestConsumeBatch_ContextCancellation(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob) // Unbuffered - will block
	ctx, cancel := context.WithCancel(context.Background())

	wo := kafkaModels.CompetitorWorkOrder{
		Channel:  "facebook",
		ReportID: "report123",
		PageID:   "page456",
	}
	woJSON, _ := json.Marshal(wo)

	consumerDone := make(chan struct{})
	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Try to send but context will be cancelled
			err := handler(ctx, "competitor-work-order-facebook-batch", nil, woJSON)
			close(consumerDone)
			return err
		},
	}

	go consumeBatch(ctx, mockConsumer, jobs, log)

	// Cancel context to trigger cancellation path
	cancel()

	select {
	case <-consumerDone:
		// Expected - handler returned
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for consumer to handle context cancellation")
	}
}

func TestConsumeRealtime_ContextCancellation(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob) // Unbuffered - will block
	ctx, cancel := context.WithCancel(context.Background())

	wo := kafkaModels.CompetitorWorkOrder{
		Channel:  "facebook",
		ReportID: "report123",
		PageID:   "page456",
	}
	woJSON, _ := json.Marshal(wo)

	consumerDone := make(chan struct{})
	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			err := handler(ctx, "competitor-work-order-facebook", nil, woJSON)
			close(consumerDone)
			return err
		},
	}

	go consumeRealtime(ctx, mockConsumer, jobs, log)

	cancel()

	select {
	case <-consumerDone:
		// Expected
	case <-time.After(1 * time.Second):
		t.Fatal("timeout")
	}
}

// ================== Multiple Jobs Tests ==================

func TestConsumeRealtime_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan CompetitorJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	messages := []kafkaModels.CompetitorWorkOrder{
		{Channel: "facebook", ReportID: "r1", PageID: "p1", Mode: "full"},
		{Channel: "facebook", ReportID: "r2", PageID: "p2", Mode: "incremental"},
		{Channel: "instagram", ReportID: "r3", PageID: "p3"}, // Should be skipped
	}

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			for _, msg := range messages {
				msgJSON, _ := json.Marshal(msg)
				handler(ctx, "competitor-work-order-facebook", nil, msgJSON)
			}
			<-ctx.Done()
			return ctx.Err()
		},
	}

	go consumeRealtime(ctx, mockConsumer, jobs, log)

	// Should receive 2 jobs (instagram one skipped)
	receivedCount := 0
	timeout := time.After(500 * time.Millisecond)

loop:
	for {
		select {
		case <-jobs:
			receivedCount++
			if receivedCount == 2 {
				break loop
			}
		case <-timeout:
			break loop
		}
	}

	if receivedCount != 2 {
		t.Errorf("expected 2 jobs, got %d", receivedCount)
	}
}

// ================== toString Edge Cases ==================

func TestToString_FloatType(t *testing.T) {
	log := logger.New("error")
	// Float should return empty string (unsupported)
	result := toString(3.14, log)
	if result != "" {
		t.Errorf("toString(float) = %q, want empty string", result)
	}
}

func TestToString_NilType(t *testing.T) {
	log := logger.New("error")
	result := toString(nil, log)
	if result != "" {
		t.Errorf("toString(nil) = %q, want empty string", result)
	}
}

// ================== Job Mode Tests ==================

func TestCompetitorJob_Modes(t *testing.T) {
	tests := []struct {
		name string
		mode api.SyncMode
	}{
		{"full_sync", api.SyncMode("full_sync")},
		{"incremental", api.SyncMode("incremental")},
		{"historical", api.SyncMode("historical")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := CompetitorJob{
				ReportID: "report123",
				PageID:   "page456",
				Mode:     tt.mode,
			}
			if job.Mode != tt.mode {
				t.Errorf("Mode = %q, want %q", job.Mode, tt.mode)
			}
		})
	}
}
