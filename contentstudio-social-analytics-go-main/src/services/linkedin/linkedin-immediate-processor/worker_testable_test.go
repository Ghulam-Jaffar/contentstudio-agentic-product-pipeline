package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/linkedin/linkedin-immediate-processor/processor"
)

// MockProcessor is a mock implementation of ProcessorAPI
type MockProcessor struct {
	ProcessAccountFunc func(ctx context.Context, wo processor.WorkOrder) error
}

func (m *MockProcessor) ProcessAccount(ctx context.Context, wo processor.WorkOrder) error {
	if m.ProcessAccountFunc != nil {
		return m.ProcessAccountFunc(ctx, wo)
	}
	return nil
}

// ================== NewWorkerService Tests ==================

func TestNewWorkerService(t *testing.T) {
	log := logger.New("error")
	proc := &MockProcessor{}

	svc := NewWorkerService(proc, log)

	if svc == nil {
		t.Fatal("NewWorkerService returned nil")
	}
	if svc.proc == nil {
		t.Error("proc is nil")
	}
	if svc.log == nil {
		t.Error("log is nil")
	}
}

// ================== ProcessWorkOrderTestable Tests ==================

func TestWorkerService_ProcessWorkOrderTestable_Success(t *testing.T) {
	log := logger.New("error")

	var processCalls int32
	proc := &MockProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo processor.WorkOrder) error {
			atomic.AddInt32(&processCalls, 1)
			if wo.AccountID != "li456" {
				t.Errorf("AccountID = %q, want %q", wo.AccountID, "li456")
			}
			return nil
		},
	}

	svc := NewWorkerService(proc, log)

	wo := processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "li456",
		WorkspaceID: "ws789",
		SyncType:    "immediate",
	}
	value, _ := json.Marshal(wo)

	err := svc.ProcessWorkOrderTestable(context.Background(), value)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if atomic.LoadInt32(&processCalls) != 1 {
		t.Errorf("processCalls = %d, want 1", processCalls)
	}
}

func TestWorkerService_ProcessWorkOrderTestable_UnmarshalError(t *testing.T) {
	log := logger.New("error")
	proc := &MockProcessor{}

	svc := NewWorkerService(proc, log)

	err := svc.ProcessWorkOrderTestable(context.Background(), []byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestWorkerService_ProcessWorkOrderTestable_ProcessError(t *testing.T) {
	log := logger.New("error")

	proc := &MockProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo processor.WorkOrder) error {
			return errors.New("process error")
		},
	}

	svc := NewWorkerService(proc, log)

	wo := processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "li456",
		WorkspaceID: "ws789",
	}
	value, _ := json.Marshal(wo)

	err := svc.ProcessWorkOrderTestable(context.Background(), value)
	if err == nil {
		t.Error("expected error")
	}
}

// ================== WorkerLoopTestable Tests ==================

func TestWorkerService_WorkerLoopTestable_ProcessesMessages(t *testing.T) {
	log := logger.New("error")

	var processCalls int32
	proc := &MockProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo processor.WorkOrder) error {
			atomic.AddInt32(&processCalls, 1)
			return nil
		},
	}

	svc := NewWorkerService(proc, log)

	workChan := make(chan workMessage, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.WorkerLoopTestable(ctx, workChan)
		close(done)
	}()

	// Send work orders
	wo := processor.WorkOrder{ID: "acc123", AccountID: "li456", WorkspaceID: "ws789"}
	value, _ := json.Marshal(wo)

	workChan <- workMessage{ctx: context.Background(), value: value}
	workChan <- workMessage{ctx: context.Background(), value: value}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(workChan)
	<-done

	if atomic.LoadInt32(&processCalls) != 2 {
		t.Errorf("processCalls = %d, want 2", processCalls)
	}
}

func TestWorkerService_WorkerLoopTestable_ChannelClose(t *testing.T) {
	log := logger.New("error")
	proc := &MockProcessor{}

	svc := NewWorkerService(proc, log)

	workChan := make(chan workMessage, 10)

	done := make(chan struct{})
	go func() {
		svc.WorkerLoopTestable(context.Background(), workChan)
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

func TestWorkerService_WorkerLoopTestable_ContextCancel(t *testing.T) {
	log := logger.New("error")
	proc := &MockProcessor{}

	svc := NewWorkerService(proc, log)

	workChan := make(chan workMessage, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.WorkerLoopTestable(ctx, workChan)
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

// ================== ParseWorkOrderTestable Tests ==================

func TestParseWorkOrderTestable_Success(t *testing.T) {
	wo := processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "li456",
		WorkspaceID: "ws789",
		SyncType:    "immediate",
	}
	value, _ := json.Marshal(wo)

	parsed, err := ParseWorkOrderTestable(value)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.ID != "acc123" {
		t.Errorf("ID = %q, want %q", parsed.ID, "acc123")
	}
	if parsed.AccountID != "li456" {
		t.Errorf("AccountID = %q, want %q", parsed.AccountID, "li456")
	}
	if parsed.WorkspaceID != "ws789" {
		t.Errorf("WorkspaceID = %q, want %q", parsed.WorkspaceID, "ws789")
	}
	if parsed.SyncType != "immediate" {
		t.Errorf("SyncType = %q, want %q", parsed.SyncType, "immediate")
	}
}

func TestParseWorkOrderTestable_InvalidJSON(t *testing.T) {
	_, err := ParseWorkOrderTestable([]byte("invalid json"))
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ================== ValidateWorkOrderTestable Tests ==================

func TestValidateWorkOrderTestable_Valid(t *testing.T) {
	wo := &processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "li456",
		WorkspaceID: "ws789",
	}

	if !ValidateWorkOrderTestable(wo) {
		t.Error("expected valid work order")
	}
}

func TestValidateWorkOrderTestable_MissingID(t *testing.T) {
	wo := &processor.WorkOrder{
		ID:          "",
		AccountID:   "li456",
		WorkspaceID: "ws789",
	}

	if ValidateWorkOrderTestable(wo) {
		t.Error("expected invalid work order (missing ID)")
	}
}

func TestValidateWorkOrderTestable_MissingAccountID(t *testing.T) {
	wo := &processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "",
		WorkspaceID: "ws789",
	}

	if ValidateWorkOrderTestable(wo) {
		t.Error("expected invalid work order (missing AccountID)")
	}
}

func TestValidateWorkOrderTestable_MissingWorkspaceID(t *testing.T) {
	wo := &processor.WorkOrder{
		ID:          "acc123",
		AccountID:   "li456",
		WorkspaceID: "",
	}

	if ValidateWorkOrderTestable(wo) {
		t.Error("expected invalid work order (missing WorkspaceID)")
	}
}

// ================== Concurrent Tests ==================

func TestWorkerService_ConcurrentProcessing(t *testing.T) {
	log := logger.New("error")

	var processCalls int64
	proc := &MockProcessor{
		ProcessAccountFunc: func(ctx context.Context, wo processor.WorkOrder) error {
			atomic.AddInt64(&processCalls, 1)
			return nil
		},
	}

	svc := NewWorkerService(proc, log)

	workChan := make(chan workMessage, 100)

	ctx, cancel := context.WithCancel(context.Background())

	// Start multiple workers
	for i := 0; i < 3; i++ {
		go svc.WorkerLoopTestable(ctx, workChan)
	}

	// Send many work orders
	wo := processor.WorkOrder{ID: "acc123", AccountID: "li456", WorkspaceID: "ws789"}
	value, _ := json.Marshal(wo)

	for i := 0; i < 50; i++ {
		workChan <- workMessage{ctx: context.Background(), value: value}
	}

	time.Sleep(200 * time.Millisecond)

	cancel()
	close(workChan)

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&processCalls) != 50 {
		t.Errorf("processCalls = %d, want 50", processCalls)
	}
}
