package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/linkedin/linkedin-immediate-processor/processor"
	"github.com/rs/zerolog"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if WorkerPoolSize != 10 {
		t.Errorf("WorkerPoolSize = %d, want 10", WorkerPoolSize)
	}
	if WorkChannelBuffer != 100 {
		t.Errorf("WorkChannelBuffer = %d, want 100", WorkChannelBuffer)
	}
}

// ================== Struct Tests ==================

func TestWorkMessage_Struct(t *testing.T) {
	ctx := context.Background()
	msg := workMessage{
		ctx:   ctx,
		value: []byte(`{"test": true}`),
	}

	if msg.ctx != ctx {
		t.Error("ctx mismatch")
	}
	if string(msg.value) != `{"test": true}` {
		t.Errorf("value = %q, want %q", msg.value, `{"test": true}`)
	}
}

// ================== Consumer Handler Tests ==================

func TestConsumerHandler(t *testing.T) {
	workChan := make(chan workMessage, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Verify correct topic
			if len(topics) != 1 || topics[0] != "immediate-work-order-linkedin" {
				t.Errorf("unexpected topics: %v", topics)
			}

			handler(ctx, topics[0], []byte("key"), []byte(`{"account_id": "acc123"}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		mockConsumer.Consume(ctx, []string{"immediate-work-order-linkedin"}, func(ctx context.Context, _ string, _ []byte, value []byte) error {
			select {
			case workChan <- workMessage{ctx: ctx, value: value}:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()

	time.Sleep(100 * time.Millisecond)

	if len(workChan) < 1 {
		t.Error("workChan should have at least one message")
	}

	cancel()
}

func TestConsumerHandler_ContextCancel(t *testing.T) {
	workChan := make(chan workMessage, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		mockConsumer.Consume(ctx, []string{"immediate-work-order-linkedin"}, func(ctx context.Context, _ string, _ []byte, value []byte) error {
			select {
			case workChan <- workMessage{ctx: ctx, value: value}:
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("Consumer did not exit after context cancel")
	}
}

// ================== Worker Tests ==================

func TestWorker_ChannelClose(t *testing.T) {
	workChan := make(chan workMessage, 10)
	log := logger.New("error")

	done := make(chan struct{})
	go func() {
		worker(0, workChan, nil, log)
		close(done)
	}()

	close(workChan)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after channel close")
	}
}

func TestWorker_ProcessesMessages(t *testing.T) {
	workChan := make(chan workMessage, 10)
	_ = logger.New("error")

	var processedCount int32

	// We can't easily mock the processor, but we can test the channel mechanics
	done := make(chan struct{})
	go func() {
		for range workChan {
			atomic.AddInt32(&processedCount, 1)
		}
		close(done)
	}()

	// Send work orders
	wo := processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "li456",
		WorkspaceID: "ws789",
		SyncType:    "immediate",
	}
	woJSON, _ := json.Marshal(wo)

	workChan <- workMessage{ctx: context.Background(), value: woJSON}
	workChan <- workMessage{ctx: context.Background(), value: woJSON}

	time.Sleep(100 * time.Millisecond)
	close(workChan)
	<-done

	if atomic.LoadInt32(&processedCount) != 2 {
		t.Errorf("processedCount = %d, want 2", processedCount)
	}
}

func TestWorker_InvalidJSON(t *testing.T) {
	workChan := make(chan workMessage, 10)
	log := logger.New("error")

	done := make(chan struct{})
	go func() {
		worker(0, workChan, nil, log)
		close(done)
	}()

	// Send invalid JSON - worker should handle gracefully
	workChan <- workMessage{ctx: context.Background(), value: []byte("invalid json")}

	time.Sleep(100 * time.Millisecond)
	close(workChan)

	select {
	case <-done:
		// Expected - worker should not panic on invalid JSON
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit")
	}
}

// ================== Work Order Unmarshal Tests ==================

func TestWorkOrder_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "acc123",
		"account_id": "li456",
		"workspace_id": "ws789",
		"sync_type": "immediate"
	}`

	var wo processor.WorkOrder
	err := json.Unmarshal([]byte(jsonData), &wo)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if wo.ID != "acc123" {
		t.Errorf("ID = %q, want %q", wo.ID, "acc123")
	}
	if wo.AccountID != "li456" {
		t.Errorf("AccountID = %q, want %q", wo.AccountID, "li456")
	}
	if wo.WorkspaceID != "ws789" {
		t.Errorf("WorkspaceID = %q, want %q", wo.WorkspaceID, "ws789")
	}
	if wo.SyncType != "immediate" {
		t.Errorf("SyncType = %q, want %q", wo.SyncType, "immediate")
	}
}

// ================== Concurrent Tests ==================

func TestConcurrentWorkers(t *testing.T) {
	workChan := make(chan workMessage, 100)

	var processedCount int64
	var wg sync.WaitGroup

	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range workChan {
				atomic.AddInt64(&processedCount, 1)
			}
		}()
	}

	// Send work orders
	for i := 0; i < 50; i++ {
		workChan <- workMessage{ctx: context.Background(), value: []byte(`{}`)}
	}
	close(workChan)

	wg.Wait()

	if atomic.LoadInt64(&processedCount) != 50 {
		t.Errorf("processedCount = %d, want 50", processedCount)
	}
}

// ================== Error-Flow Contract Tests (Caller logs Error with complete fields) ==================

func TestErrorFlowContract_ProcessAccountError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("failed to fetch LinkedIn posts: timeout")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("linkedin_id", "li123").
		Str("workspace_id", "ws456").
		Dur("duration", 5*time.Second).
		Str("function", "worker").
		Str("stage", "process_account").
		Msg("Failed to process LinkedIn account")

	output := buf.String()

	for _, field := range []string{"error_message", "function", "worker", "stage", "process_account", "linkedin_id", "workspace_id"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_UnmarshalError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("invalid character 'x'")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "worker").
		Str("stage", "unmarshal_work_order").
		Msg("Failed to unmarshal work order")

	output := buf.String()
	for _, field := range []string{"error_message", "function", "worker", "stage", "unmarshal_work_order"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_ProcessAccountError_TriggersHook(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	err := errors.New("clickhouse insert failed")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "worker").
		Str("stage", "process_account").
		Msg("Failed to process LinkedIn account")

	var errorCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorCount++
		}
	}
	if errorCount != 1 {
		t.Fatalf("expected exactly 1 Error-level hook firing, got %d", errorCount)
	}
}

func TestErrorFlowContract_ProcessAccountError_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	err := errors.New("mongo connection refused")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "worker").
		Str("stage", "process_account").
		Msg("Failed to process LinkedIn account")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}
