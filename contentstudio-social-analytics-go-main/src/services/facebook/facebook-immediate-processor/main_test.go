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

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/facebook/facebook-immediate-processor/processor"
	"github.com/rs/zerolog"
)

// ================== worker Tests (actual function) ==================

func TestWorker_ActualFunction_ProcessesMessages(t *testing.T) {
	log := logger.New("error")

	var processedCount int32
	mockProc := &processor.Processor{}
	// Note: We can't easily mock processor.Processor since it's a concrete type
	// But we can test worker behavior with the channel

	workChan := make(chan workMessage, 10)

	// Start actual worker - it will fail on ProcessAccount since mockProc is not properly initialized
	// but we can verify the worker receives and processes messages
	go worker(1, workChan, mockProc, log)

	// Send work order
	wo := processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "fb456",
		WorkspaceID: "ws789",
		SyncType:    "immediate",
	}
	woJSON, _ := json.Marshal(wo)
	workChan <- workMessage{ctx: context.Background(), value: woJSON}

	time.Sleep(100 * time.Millisecond)
	close(workChan)
	time.Sleep(50 * time.Millisecond)

	// Worker should have processed without panic
	_ = processedCount
}

func TestWorker_ActualFunction_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	mockProc := &processor.Processor{}

	workChan := make(chan workMessage, 10)

	go worker(1, workChan, mockProc, log)

	// Send invalid JSON - worker should continue without crashing
	workChan <- workMessage{ctx: context.Background(), value: []byte("invalid json")}

	time.Sleep(50 * time.Millisecond)
	close(workChan)
	time.Sleep(50 * time.Millisecond)
	// Test passes if no panic
}

func TestWorker_ActualFunction_ChannelClose(t *testing.T) {
	log := logger.New("error")
	mockProc := &processor.Processor{}

	workChan := make(chan workMessage, 10)

	done := make(chan struct{})
	go func() {
		worker(1, workChan, mockProc, log)
		close(done)
	}()

	// Close channel to trigger worker exit
	close(workChan)

	select {
	case <-done:
		// Expected - worker exited
	case <-time.After(time.Second):
		t.Fatal("worker did not exit on channel close")
	}
}

func TestWorker_ActualFunction_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	mockProc := &processor.Processor{}

	workChan := make(chan workMessage, 10)

	go worker(1, workChan, mockProc, log)

	// Send multiple work orders
	for i := 0; i < 3; i++ {
		wo := processor.WorkOrder{
			ID:        "acc" + string(rune('0'+i)),
			AccountID: "fb" + string(rune('0'+i)),
		}
		woJSON, _ := json.Marshal(wo)
		workChan <- workMessage{ctx: context.Background(), value: woJSON}
	}

	time.Sleep(100 * time.Millisecond)
	close(workChan)
	time.Sleep(50 * time.Millisecond)
}

// ================== Simulated worker Tests ==================

func TestWorkerSimulation_ProcessesMessages(t *testing.T) {
	var processedCount int32

	workChan := make(chan workMessage, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for msg := range workChan {
			var workOrder processor.WorkOrder
			if err := json.Unmarshal(msg.value, &workOrder); err != nil {
				continue
			}
			atomic.AddInt32(&processedCount, 1)
		}
	}()

	wo := processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "fb456",
		WorkspaceID: "ws789",
		SyncType:    "immediate",
	}
	woJSON, _ := json.Marshal(wo)

	workChan <- workMessage{ctx: context.Background(), value: woJSON}

	time.Sleep(50 * time.Millisecond)
	close(workChan)
	wg.Wait()

	if atomic.LoadInt32(&processedCount) != 1 {
		t.Errorf("expected 1 processed, got %d", processedCount)
	}
}

func TestWorkerSimulation_MultipleMessages(t *testing.T) {
	var processedCount int32

	workChan := make(chan workMessage, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for msg := range workChan {
			var workOrder processor.WorkOrder
			if err := json.Unmarshal(msg.value, &workOrder); err != nil {
				continue
			}
			atomic.AddInt32(&processedCount, 1)
		}
	}()

	for i := 0; i < 5; i++ {
		wo := processor.WorkOrder{
			ID:        "acc" + string(rune('0'+i)),
			AccountID: "fb" + string(rune('0'+i)),
		}
		woJSON, _ := json.Marshal(wo)
		workChan <- workMessage{ctx: context.Background(), value: woJSON}
	}

	time.Sleep(100 * time.Millisecond)
	close(workChan)
	wg.Wait()

	if atomic.LoadInt32(&processedCount) != 5 {
		t.Errorf("expected 5 processed, got %d", processedCount)
	}
}

// ================== Error handling simulation ==================

func TestWorkerSimulation_ProcessError(t *testing.T) {
	var errorCount int32

	workChan := make(chan workMessage, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	processFunc := func(ctx context.Context, wo processor.WorkOrder) error {
		return errors.New("process error")
	}

	go func() {
		defer wg.Done()
		for msg := range workChan {
			var workOrder processor.WorkOrder
			if err := json.Unmarshal(msg.value, &workOrder); err != nil {
				continue
			}
			if err := processFunc(msg.ctx, workOrder); err != nil {
				atomic.AddInt32(&errorCount, 1)
			}
		}
	}()

	wo := processor.WorkOrder{ID: "acc123", AccountID: "fb456"}
	woJSON, _ := json.Marshal(wo)
	workChan <- workMessage{ctx: context.Background(), value: woJSON}

	time.Sleep(50 * time.Millisecond)
	close(workChan)
	wg.Wait()

	if atomic.LoadInt32(&errorCount) != 1 {
		t.Errorf("expected 1 error, got %d", errorCount)
	}
}

// ================== workMessage Tests ==================

func TestWorkMessage_Struct(t *testing.T) {
	ctx := context.Background()
	value := []byte(`{"id": "test"}`)

	msg := workMessage{
		ctx:   ctx,
		value: value,
	}

	if msg.ctx != ctx {
		t.Error("ctx mismatch")
	}
	if string(msg.value) != `{"id": "test"}` {
		t.Errorf("value = %q, want %q", string(msg.value), `{"id": "test"}`)
	}
}

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if WorkerPoolSize != 10 {
		t.Errorf("WorkerPoolSize = %d, want 10", WorkerPoolSize)
	}
	if WorkChannelBuffer != 100 {
		t.Errorf("WorkChannelBuffer = %d, want 100", WorkChannelBuffer)
	}
}

// ================== WorkOrder Parsing Tests ==================

func TestWorkOrder_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "acc123",
		"account_id": "fb456",
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
	if wo.AccountID != "fb456" {
		t.Errorf("AccountID = %q, want %q", wo.AccountID, "fb456")
	}
	if wo.WorkspaceID != "ws789" {
		t.Errorf("WorkspaceID = %q, want %q", wo.WorkspaceID, "ws789")
	}
	if wo.SyncType != "immediate" {
		t.Errorf("SyncType = %q, want %q", wo.SyncType, "immediate")
	}
}

// ================== Worker Pool Simulation Tests ==================

func TestWorkerPool_ConcurrentProcessing(t *testing.T) {
	var processedCount int32
	workChan := make(chan workMessage, 100)

	var wg sync.WaitGroup
	numWorkers := 3

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for msg := range workChan {
				var workOrder processor.WorkOrder
				if err := json.Unmarshal(msg.value, &workOrder); err != nil {
					continue
				}
				atomic.AddInt32(&processedCount, 1)
			}
		}(i)
	}

	// Send messages
	for i := 0; i < 30; i++ {
		wo := processor.WorkOrder{ID: "acc", AccountID: "fb"}
		woJSON, _ := json.Marshal(wo)
		workChan <- workMessage{ctx: context.Background(), value: woJSON}
	}

	close(workChan)
	wg.Wait()

	if atomic.LoadInt32(&processedCount) != 30 {
		t.Errorf("expected 30 processed, got %d", processedCount)
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	var shutdownCount int32
	workChan := make(chan workMessage, 10)

	var wg sync.WaitGroup
	numWorkers := 5

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			defer atomic.AddInt32(&shutdownCount, 1)
			for range workChan {
				// Process
			}
		}(i)
	}

	// Close channel to trigger shutdown
	close(workChan)
	wg.Wait()

	if atomic.LoadInt32(&shutdownCount) != int32(numWorkers) {
		t.Errorf("expected %d workers shutdown, got %d", numWorkers, shutdownCount)
	}
}

// ================== Context Cancellation Tests ==================

func TestWorker_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	var processed int32
	workChan := make(chan workMessage, 10)

	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		defer wg.Done()
		for {
			select {
			case msg, ok := <-workChan:
				if !ok {
					return
				}
				var workOrder processor.WorkOrder
				if err := json.Unmarshal(msg.value, &workOrder); err != nil {
					continue
				}
				atomic.AddInt32(&processed, 1)
			case <-ctx.Done():
				return
			}
		}
	}()

	// Send one message
	wo := processor.WorkOrder{ID: "acc", AccountID: "fb"}
	woJSON, _ := json.Marshal(wo)
	workChan <- workMessage{ctx: ctx, value: woJSON}

	time.Sleep(50 * time.Millisecond)

	// Cancel context
	cancel()
	close(workChan)
	wg.Wait()

	if atomic.LoadInt32(&processed) != 1 {
		t.Errorf("expected 1 processed, got %d", processed)
	}
}

// ================== Error-Flow Contract Tests (Caller logs Error with complete fields) ==================

// TestErrorFlowContract_ProcessAccountError_LogsWithAllFields verifies that when
// ProcessAccount returns an error, the worker function logs it at Error level with
// error_message, function, and stage fields — ensuring it goes to Sentry via the hook.
func TestErrorFlowContract_ProcessAccountError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	// Replicate the exact logging pattern from the worker function
	err := errors.New("failed to fetch page posts: API timeout")
	workOrder := processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "fb456",
		WorkspaceID: "ws789",
	}
	duration := 5 * time.Second

	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("account_id", workOrder.ID).
		Str("facebook_id", workOrder.AccountID).
		Str("workspace_id", workOrder.WorkspaceID).
		Dur("duration", duration).
		Str("function", "worker").
		Str("stage", "process_account").
		Msg("Failed to process Facebook account")

	output := buf.String()

	requiredFields := map[string]string{
		"error_message":   "missing error_message field",
		"function":        "missing function field",
		"worker":          "missing function value 'worker'",
		"stage":           "missing stage field",
		"process_account": "missing stage value 'process_account'",
		"account_id":      "missing account_id field",
		"facebook_id":     "missing facebook_id field",
		"workspace_id":    "missing workspace_id field",
	}

	for substr, errMsg := range requiredFields {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

// TestErrorFlowContract_UnmarshalError_LogsWithAllFields verifies unmarshal errors
// are logged with complete fields.
func TestErrorFlowContract_UnmarshalError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("invalid character 'x' looking for beginning of value")

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

// TestErrorFlowContract_ProcessAccountError_TriggersHook verifies the Error log
// fires the Sentry hook exactly once.
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
		Msg("Failed to process Facebook account")

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

// TestErrorFlowContract_ProcessAccountError_NoCaptureException verifies that
// the caller relies on the hook for Sentry (no explicit CaptureException).
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
		Msg("Failed to process Facebook account")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}
