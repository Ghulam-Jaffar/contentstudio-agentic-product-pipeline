package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/semaphore"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if maxPageWorkers != 15 {
		t.Errorf("maxPageWorkers = %d, want 15", maxPageWorkers)
	}
	if maxProfileWorkers != 15 {
		t.Errorf("maxProfileWorkers = %d, want 15", maxProfileWorkers)
	}
	if workOrderChanSize != 500 {
		t.Errorf("workOrderChanSize = %d, want 500", workOrderChanSize)
	}
	if statsConcsPerWorker != 5 {
		t.Errorf("statsConcsPerWorker = %d, want 5", statsConcsPerWorker)
	}
	if mediaConcPerWorker != 5 {
		t.Errorf("mediaConcPerWorker = %d, want 5", mediaConcPerWorker)
	}
	if geoConcPerWorker != 3 {
		t.Errorf("geoConcPerWorker = %d, want 3", geoConcPerWorker)
	}
	if perAccountConcurrency != 1 {
		t.Errorf("perAccountConcurrency = %d, want 1", perAccountConcurrency)
	}
	if timestampUpdateChanSize != 1000 {
		t.Errorf("timestampUpdateChanSize = %d, want 1000", timestampUpdateChanSize)
	}
	if idleTimeout != 5*time.Minute {
		t.Errorf("idleTimeout = %v, want 5m", idleTimeout)
	}
	if idleCheckInterval != 30*time.Second {
		t.Errorf("idleCheckInterval = %v, want 30s", idleCheckInterval)
	}
	if topicWorkOrderPageBatch != "work-order-linkedin-page-batch" {
		t.Errorf("topicWorkOrderPageBatch = %q, want %q", topicWorkOrderPageBatch, "work-order-linkedin-page-batch")
	}
	if topicWorkOrderProfileBatch != "work-order-linkedin-profile-batch" {
		t.Errorf("topicWorkOrderProfileBatch = %q, want %q", topicWorkOrderProfileBatch, "work-order-linkedin-profile-batch")
	}
}

// ================== Struct Tests ==================

func TestWorkOrderMessage_Struct(t *testing.T) {
	wo := WorkOrderMessage{
		AccountID:  "acc123",
		LinkedinID: "li456",
		Value:      []byte(`{"test": true}`),
	}

	if wo.AccountID != "acc123" {
		t.Errorf("AccountID = %q, want %q", wo.AccountID, "acc123")
	}
	if wo.LinkedinID != "li456" {
		t.Errorf("LinkedinID = %q, want %q", wo.LinkedinID, "li456")
	}
}

func TestTimestampUpdateRequest_Struct(t *testing.T) {
	req := TimestampUpdateRequest{
		AccountID:  "acc123",
		LinkedinID: "li456",
	}

	if req.AccountID != "acc123" {
		t.Errorf("AccountID = %q, want %q", req.AccountID, "acc123")
	}
	if req.LinkedinID != "li456" {
		t.Errorf("LinkedinID = %q, want %q", req.LinkedinID, "li456")
	}
}

func TestLinkedInBatchWorkOrder_Struct(t *testing.T) {
	batch := LinkedInBatchWorkOrder{
		BatchID:  "batch123",
		SyncType: "incremental",
		Accounts: []kafkamodels.LinkedinAccountWorkOrder{
			{ID: "acc1", LinkedinID: "li1"},
			{ID: "acc2", LinkedinID: "li2"},
		},
	}

	if batch.BatchID != "batch123" {
		t.Errorf("BatchID = %q, want %q", batch.BatchID, "batch123")
	}
	if len(batch.Accounts) != 2 {
		t.Errorf("len(Accounts) = %d, want 2", len(batch.Accounts))
	}
}

// ================== consumeBatchWorkOrders Tests ==================

func TestConsumeBatchWorkOrders_ProcessesBatch(t *testing.T) {
	log := logger.New("error")

	batch := LinkedInBatchWorkOrder{
		BatchID:  "batch123",
		SyncType: "incremental",
		Accounts: []kafkamodels.LinkedinAccountWorkOrder{
			{ID: "acc1", LinkedinID: "li1", WorkspaceID: "ws1"},
			{ID: "acc2", LinkedinID: "li2", WorkspaceID: "ws1"},
		},
	}
	batchJSON, _ := json.Marshal(batch)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), batchJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	sem := semaphore.NewWeighted(50)
	var dispatchWg sync.WaitGroup
	var processedCount int32
	processAccount := func(ctx context.Context, msg WorkOrderMessage) error {
		atomic.AddInt32(&processedCount, 1)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumeBatchWorkOrders(ctx, log, mockConsumer, "test-group", []string{"test-topic"}, sem, &dispatchWg, "page", nil, nil, nil, processAccount)

	time.Sleep(100 * time.Millisecond)
	dispatchWg.Wait()

	if atomic.LoadInt32(&processedCount) != 2 {
		t.Errorf("processAccount called %d times, want 2", processedCount)
	}
}

func TestConsumeBatchWorkOrders_InvalidJSON(t *testing.T) {
	log := logger.New("error")

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), []byte("invalid json"))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	sem := semaphore.NewWeighted(50)
	var dispatchWg sync.WaitGroup
	var processedCount int32
	processAccount := func(ctx context.Context, msg WorkOrderMessage) error {
		atomic.AddInt32(&processedCount, 1)
		return nil
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumeBatchWorkOrders(ctx, log, mockConsumer, "test-group", []string{"test-topic"}, sem, &dispatchWg, "page", nil, nil, nil, processAccount)

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&processedCount) != 0 {
		t.Errorf("processAccount should not be called for invalid JSON, called %d times", processedCount)
	}
}

func TestConsumeBatchWorkOrders_TracksLastMessageTime(t *testing.T) {
	log := logger.New("error")

	batch := LinkedInBatchWorkOrder{
		BatchID:  "batch123",
		Accounts: []kafkamodels.LinkedinAccountWorkOrder{},
	}
	batchJSON, _ := json.Marshal(batch)

	var lastMessageTime int64 = 0

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), batchJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	sem := semaphore.NewWeighted(50)
	var dispatchWg sync.WaitGroup
	processAccount := func(ctx context.Context, msg WorkOrderMessage) error { return nil }

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go consumeBatchWorkOrders(ctx, log, mockConsumer, "test-group", []string{"test-topic"}, sem, &dispatchWg, "page", &lastMessageTime, nil, nil, processAccount)

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&lastMessageTime) == 0 {
		t.Error("lastMessageTime should be updated")
	}
}

func TestConsumeBatchWorkOrders_ContextCancel(t *testing.T) {
	log := logger.New("error")

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	sem := semaphore.NewWeighted(50)
	var dispatchWg sync.WaitGroup
	processAccount := func(ctx context.Context, msg WorkOrderMessage) error { return nil }

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		consumeBatchWorkOrders(ctx, log, mockConsumer, "test-group", []string{"test-topic"}, sem, &dispatchWg, "page", nil, nil, nil, processAccount)
		close(done)
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("consumeBatchWorkOrders did not exit after context cancel")
	}
}

// ================== startTimestampUpdater Tests ==================

func TestStartTimestampUpdater_ContextCancel(t *testing.T) {
	log := logger.New("error")
	repo := &mongodb.MockUnifiedSocialRepository{}
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	startTimestampUpdater(ctx, &wg, repo, timestampUpdateChan, log)

	cancel()
	wg.Wait()
}

func TestStartTimestampUpdater_ChannelClose(t *testing.T) {
	log := logger.New("error")
	repo := &mongodb.MockUnifiedSocialRepository{}
	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	ctx := context.Background()
	var wg sync.WaitGroup

	startTimestampUpdater(ctx, &wg, repo, timestampUpdateChan, log)

	close(timestampUpdateChan)
	wg.Wait()
}

func TestStartTimestampUpdater_ProcessesRequest(t *testing.T) {
	log := logger.New("error")

	var updateCalled int32
	repo := &mongodb.MockUnifiedSocialRepository{
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			atomic.AddInt32(&updateCalled, 1)
			return nil
		},
	}

	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	startTimestampUpdater(ctx, &wg, repo, timestampUpdateChan, log)

	// Send update request with valid ObjectID
	validID := primitive.NewObjectID().Hex()
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID:  validID,
		LinkedinID: "li456",
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(timestampUpdateChan)
	wg.Wait()

	if atomic.LoadInt32(&updateCalled) != 1 {
		t.Errorf("UpdateAnalyticsTimestamp called %d times, want 1", updateCalled)
	}
}

func TestStartTimestampUpdater_InvalidObjectID(t *testing.T) {
	log := logger.New("error")

	var updateCalled int32
	repo := &mongodb.MockUnifiedSocialRepository{
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			atomic.AddInt32(&updateCalled, 1)
			return nil
		},
	}

	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	startTimestampUpdater(ctx, &wg, repo, timestampUpdateChan, log)

	// Send update request with invalid ObjectID
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID:  "invalid-id",
		LinkedinID: "li456",
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(timestampUpdateChan)
	wg.Wait()

	if atomic.LoadInt32(&updateCalled) != 0 {
		t.Errorf("UpdateAnalyticsTimestamp should not be called for invalid ObjectID, called %d times", updateCalled)
	}
}

func TestStartTimestampUpdater_UpdateError(t *testing.T) {
	log := logger.New("error")

	repo := &mongodb.MockUnifiedSocialRepository{
		UpdateAnalyticsTimestampFunc: func(ctx context.Context, id primitive.ObjectID, field string, timestamp time.Time) error {
			return fmt.Errorf("update error")
		},
	}

	timestampUpdateChan := make(chan TimestampUpdateRequest, 10)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	startTimestampUpdater(ctx, &wg, repo, timestampUpdateChan, log)

	// Send update request
	validID := primitive.NewObjectID().Hex()
	timestampUpdateChan <- TimestampUpdateRequest{
		AccountID:  validID,
		LinkedinID: "li456",
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(timestampUpdateChan)
	wg.Wait()
	// Test passes if no panic
}

// ================== Batch Work Order Unmarshal Tests ==================

func TestLinkedInBatchWorkOrder_Unmarshal(t *testing.T) {
	jsonData := `{
		"batch_id": "batch123",
		"sync_type": "incremental",
		"accounts": [
			{
				"id": "acc1",
				"linkedin_id": "li1",
				"access_token": "token1",
				"sync_type": "incremental"
			}
		]
	}`

	var batch LinkedInBatchWorkOrder
	err := json.Unmarshal([]byte(jsonData), &batch)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if batch.BatchID != "batch123" {
		t.Errorf("BatchID = %q, want %q", batch.BatchID, "batch123")
	}
	if len(batch.Accounts) != 1 {
		t.Errorf("len(Accounts) = %d, want 1", len(batch.Accounts))
	}
}

// ================== Concurrent Tests ==================

func TestSemForAccount_ConcurrentAccess(t *testing.T) {
	// Reset global map
	accountSemaphores = sync.Map{}

	var wg sync.WaitGroup
	numGoroutines := 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			accountID := fmt.Sprintf("account_%d", id%10)
			sem := semForAccount(accountID)
			if sem == nil {
				t.Error("semForAccount returned nil")
			}
		}(i)
	}

	wg.Wait()
}

// ================== Two-consumer pattern Tests ==================

func TestStartBatchConsumers_StartsConsumers(t *testing.T) {
	log := logger.New("error")

	var pageConsumeStarted, profileConsumeStarted int32

	pageConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			atomic.AddInt32(&pageConsumeStarted, 1)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profileConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			atomic.AddInt32(&profileConsumeStarted, 1)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	pageSem := semaphore.NewWeighted(maxConcurrentAccounts)
	profileSem := semaphore.NewWeighted(maxConcurrentAccounts)
	var dispatchWg sync.WaitGroup
	processAccount := func(ctx context.Context, msg WorkOrderMessage) error { return nil }

	ctx, cancel := context.WithCancel(context.Background())

	go consumeBatchWorkOrders(ctx, log, pageConsumer, pageConsumerGroup, []string{topicWorkOrderPageBatch}, pageSem, &dispatchWg, "page", nil, nil, nil, processAccount)
	go consumeBatchWorkOrders(ctx, log, profileConsumer, profileConsumerGroup, []string{topicWorkOrderProfileBatch}, profileSem, &dispatchWg, "profile", nil, nil, nil, processAccount)

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&pageConsumeStarted) != 1 {
		t.Error("page consumer should be started")
	}
	if atomic.LoadInt32(&profileConsumeStarted) != 1 {
		t.Error("profile consumer should be started")
	}

	cancel()
}

func TestStartBatchConsumersWithTracking_TracksTime(t *testing.T) {
	log := logger.New("error")

	batch := LinkedInBatchWorkOrder{BatchID: "test", Accounts: []kafkamodels.LinkedinAccountWorkOrder{}}
	batchJSON, _ := json.Marshal(batch)

	var lastMessageTime int64 = 0

	pageConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), batchJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profileConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	pageSem := semaphore.NewWeighted(maxConcurrentAccounts)
	profileSem := semaphore.NewWeighted(maxConcurrentAccounts)
	var dispatchWg sync.WaitGroup
	processAccount := func(ctx context.Context, msg WorkOrderMessage) error { return nil }

	ctx, cancel := context.WithCancel(context.Background())

	go consumeBatchWorkOrders(ctx, log, pageConsumer, pageConsumerGroup, []string{topicWorkOrderPageBatch}, pageSem, &dispatchWg, "page", &lastMessageTime, nil, nil, processAccount)
	go consumeBatchWorkOrders(ctx, log, profileConsumer, profileConsumerGroup, []string{topicWorkOrderProfileBatch}, profileSem, &dispatchWg, "profile", &lastMessageTime, nil, nil, processAccount)

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt64(&lastMessageTime) == 0 {
		t.Error("lastMessageTime should be updated")
	}

	cancel()
}

// ================== Logging Contract Tests (Point 4 — Calling service logs errors with context) ==================

func TestLoggingContract_LinkedInFetcher_ErrorHasContextFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "failed to unmarshal batch work order").
		Str("function", "startBatchConsumer").
		Str("stage", "unmarshal_batch_work_order").
		Msg("LinkedIn fetcher error")

	output := buf.String()

	checks := map[string]string{
		"ERR":                        "expected ERR level",
		"error_message":              "expected error_message field",
		"function":                   "expected function field",
		"startBatchConsumer":         "expected startBatchConsumer value",
		"stage":                      "expected stage field",
		"unmarshal_batch_work_order": "expected unmarshal_batch_work_order stage",
	}
	for substr, errMsg := range checks {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

func TestLoggingContract_LinkedInFetcher_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "Failed to update account state").
		Str("function", "postProcessWorkOrder").
		Str("stage", "update_account_state").
		Msg("Failed to update account state to Processed")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestLoggingContract_LinkedInFetcher_SingleSentryEvent(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "batch consumer error").
		Str("function", "startBatchConsumer").
		Str("stage", "consume_batch").
		Msg("Batch consumer error")

	var errorLevelCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorLevelCount++
		}
	}
	if errorLevelCount != 1 {
		t.Fatalf("expected exactly 1 ErrorLevel hook firing, got %d", errorLevelCount)
	}
}

func TestLoggingContract_LinkedInFetcher_ExpectedError_WarnOnly(t *testing.T) {
	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	log.Warn().
		Str("error_message", "Expected LinkedIn API error fetching page statistics").
		Str("function", "fetchPageData").
		Str("linkedin_id", "li-page-123").
		Msg("Expected LinkedIn API error fetching page statistics")

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Fatalf("expected WRN level, got: %s", output)
	}
	if strings.Contains(output, "ERR") {
		t.Fatalf("expected error should NOT produce ERR level: %s", output)
	}

	for _, r := range *hookRecords {
		if r.Level >= zerolog.ErrorLevel {
			t.Fatalf("expected error should not trigger Error+ hook, got level %v", r.Level)
		}
	}
}
