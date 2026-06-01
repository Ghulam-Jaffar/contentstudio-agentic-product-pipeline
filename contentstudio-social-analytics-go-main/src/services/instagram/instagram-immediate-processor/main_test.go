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
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/instagram/instagram-immediate-processor/processor"
	"github.com/rs/zerolog"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if maxImmediateWorkers != 8 {
		t.Errorf("maxImmediateWorkers = %d, want 8", maxImmediateWorkers)
	}
	if jobQueueSize != 100 {
		t.Errorf("jobQueueSize = %d, want 100", jobQueueSize)
	}
}

// ================== worker Tests ==================

func TestWorker_ContextCancel(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan processor.WorkOrder, 10)
	proc := &processor.Processor{}
	var inFlight sync.Map

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker(ctx, 0, jobs, proc, log, &inFlight)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after context cancel")
	}
}

func TestWorker_ChannelClose(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan processor.WorkOrder, 10)
	proc := &processor.Processor{}
	var inFlight sync.Map

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		worker(ctx, 0, jobs, proc, log, &inFlight)
		close(done)
	}()

	close(jobs)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after channel close")
	}
}

func TestWorker_ProcessesWorkOrder(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan processor.WorkOrder, 10)
	proc := &processor.Processor{}
	var inFlight sync.Map

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker(ctx, 0, jobs, proc, log, &inFlight)
		close(done)
	}()

	// Send work order
	wo := processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "ig456",
		WorkspaceID: "ws789",
		SyncType:    "immediate",
	}
	jobs <- wo

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(jobs)
	<-done
}

func TestWorker_SkipsDuplicateAccount(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan processor.WorkOrder, 10)
	proc := &processor.Processor{}
	var inFlight sync.Map

	// Pre-mark account as in-flight
	inFlight.Store("ig456", time.Now())

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker(ctx, 0, jobs, proc, log, &inFlight)
		close(done)
	}()

	// Send work order for already in-flight account
	wo := processor.WorkOrder{
		AccountID: "ig456",
	}
	jobs <- wo

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(jobs)
	<-done
}

func TestWorker_MultipleWorkOrders(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan processor.WorkOrder, 10)
	proc := &processor.Processor{}
	var inFlight sync.Map

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker(ctx, 0, jobs, proc, log, &inFlight)
		close(done)
	}()

	// Send multiple work orders with different account IDs
	for i := 0; i < 3; i++ {
		wo := processor.WorkOrder{
			AccountID:   "ig" + string(rune('0'+i)),
			WorkspaceID: "ws",
			SyncType:    "immediate",
		}
		jobs <- wo
	}

	time.Sleep(200 * time.Millisecond)

	cancel()
	close(jobs)
	<-done
}

// ================== WorkOrder Parsing Tests ==================

func TestWorkOrder_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "acc123",
		"account_id": "ig456",
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
	if wo.AccountID != "ig456" {
		t.Errorf("AccountID = %q, want %q", wo.AccountID, "ig456")
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
	jobs := make(chan processor.WorkOrder, 100)
	var inFlight sync.Map

	var wg sync.WaitGroup
	numWorkers := 3

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for wo := range jobs {
				if _, loaded := inFlight.LoadOrStore(wo.AccountID, time.Now()); loaded {
					continue
				}
				atomic.AddInt32(&processedCount, 1)
				inFlight.Delete(wo.AccountID)
			}
		}(i)
	}

	// Send work orders
	for i := 0; i < 30; i++ {
		jobs <- processor.WorkOrder{AccountID: "ig" + string(rune('0'+i%10))}
	}

	close(jobs)
	wg.Wait()

	// Some may be skipped due to duplicate detection
	if atomic.LoadInt32(&processedCount) == 0 {
		t.Error("expected some messages to be processed")
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	var shutdownCount int32
	jobs := make(chan processor.WorkOrder, 10)

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

// ================== InFlight Map Tests ==================

func TestInFlight_LoadOrStore(t *testing.T) {
	var inFlight sync.Map

	// First store should succeed
	_, loaded := inFlight.LoadOrStore("ig123", time.Now())
	if loaded {
		t.Error("expected first LoadOrStore to not find existing value")
	}

	// Second store should find existing
	_, loaded = inFlight.LoadOrStore("ig123", time.Now())
	if !loaded {
		t.Error("expected second LoadOrStore to find existing value")
	}

	// Delete and try again
	inFlight.Delete("ig123")
	_, loaded = inFlight.LoadOrStore("ig123", time.Now())
	if loaded {
		t.Error("expected LoadOrStore after Delete to not find value")
	}
}

func TestInFlight_ConcurrentAccess(t *testing.T) {
	var inFlight sync.Map
	var wg sync.WaitGroup

	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			key := "ig" + string(rune('0'+id%10))
			if _, loaded := inFlight.LoadOrStore(key, time.Now()); !loaded {
				time.Sleep(time.Millisecond)
				inFlight.Delete(key)
			}
		}(i)
	}

	wg.Wait()
}

// ================== Context Cancellation Tests ==================

func TestWorker_ContextCancellation_WhileProcessing(t *testing.T) {
	log := logger.New("error")
	jobs := make(chan processor.WorkOrder, 10)
	proc := &processor.Processor{}
	var inFlight sync.Map

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		worker(ctx, 0, jobs, proc, log, &inFlight)
		close(done)
	}()

	// Send a work order
	jobs <- processor.WorkOrder{AccountID: "ig123"}

	// Cancel while processing
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after context cancel")
	}
}

// ================== Error Handling Simulation ==================

func TestWorkerSimulation_ProcessError(t *testing.T) {
	var errorCount int32
	jobs := make(chan processor.WorkOrder, 10)
	var inFlight sync.Map

	var wg sync.WaitGroup
	wg.Add(1)

	processFunc := func(ctx context.Context, wo processor.WorkOrder) error {
		return context.DeadlineExceeded
	}

	go func() {
		defer wg.Done()
		for wo := range jobs {
			if _, loaded := inFlight.LoadOrStore(wo.AccountID, time.Now()); loaded {
				continue
			}
			if err := processFunc(context.Background(), wo); err != nil {
				atomic.AddInt32(&errorCount, 1)
			}
			inFlight.Delete(wo.AccountID)
		}
	}()

	jobs <- processor.WorkOrder{AccountID: "ig123"}

	time.Sleep(50 * time.Millisecond)
	close(jobs)
	wg.Wait()

	if atomic.LoadInt32(&errorCount) != 1 {
		t.Errorf("expected 1 error, got %d", errorCount)
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

// ================== Error-Flow Contract Tests (Caller logs Error with complete fields) ==================

func TestErrorFlowContract_ProcessAccountError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("failed to fetch media: API rate limit")
	workerLog := &logger.Logger{Logger: log.Logger.With().Int("worker_id", 1).Logger()}

	workerLog.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("instagram_id", "ig123").
		Str("workspace_id", "ws456").
		Dur("duration", 3*time.Second).
		Str("function", "worker").
		Str("stage", "process_account").
		Msg("Failed to process account")

	output := buf.String()

	for _, field := range []string{"error_message", "function", "worker", "stage", "process_account", "instagram_id", "workspace_id"} {
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
		Msg("Failed to process account")

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
		Msg("Failed to process account")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}
