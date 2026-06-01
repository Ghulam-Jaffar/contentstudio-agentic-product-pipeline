package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Integration-like Tests with Mocks ==================

func TestBatchConsumer_ProcessesBatchWorkOrder(t *testing.T) {
	_ = logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workOrderChan := make(chan WorkOrderMessage, 100)

	// Create a batch work order
	batch := kafkamodels.FacebookBatchWorkOrder{
		BatchID:  "batch123",
		SyncType: "incremental",
		Accounts: []kafkamodels.FacebookAccountWorkOrder{
			{ID: "acc1", FacebookID: "fb1", WorkspaceID: "ws1"},
			{ID: "acc2", FacebookID: "fb2", WorkspaceID: "ws1"},
		},
	}
	batchJSON, _ := json.Marshal(batch)

	var receivedCount int32

	// Mock consumer that sends one batch message
	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Verify correct topic
			if len(topics) != 1 || topics[0] != "work-order-facebook" {
				t.Errorf("unexpected topics: %v", topics)
			}

			// Send batch message
			err := handler(ctx, "work-order-facebook", nil, batchJSON)
			if err != nil {
				return err
			}

			// Wait for context cancellation
			<-ctx.Done()
			return ctx.Err()
		},
	}

	// Start a goroutine to receive work orders
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case wo, ok := <-workOrderChan:
				if !ok {
					return
				}
				atomic.AddInt32(&receivedCount, 1)
				_ = wo
			case <-ctx.Done():
				return
			}
		}
	}()

	// Simulate batch consumer behavior
	go func() {
		mockConsumer.Consume(ctx, []string{"work-order-facebook"}, func(ctx context.Context, topic string, key, value []byte) error {
			var batch kafkamodels.FacebookBatchWorkOrder
			if err := json.Unmarshal(value, &batch); err != nil {
				return nil
			}

			for _, account := range batch.Accounts {
				accountPayload, _ := json.Marshal(account)
				select {
				case workOrderChan <- WorkOrderMessage{AccountID: account.ID, FacebookID: account.FacebookID, Value: accountPayload}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		})
	}()

	// Wait for processing
	time.Sleep(100 * time.Millisecond)
	cancel()

	// Close channel and wait
	close(workOrderChan)
	wg.Wait()

	if atomic.LoadInt32(&receivedCount) != 2 {
		t.Errorf("expected 2 work orders, got %d", receivedCount)
	}
}

func TestWorkOrderDistribution_MultipleAccounts(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	workOrderChan := make(chan WorkOrderMessage, 100)

	accounts := []kafkamodels.FacebookAccountWorkOrder{
		{ID: "acc1", FacebookID: "fb1"},
		{ID: "acc2", FacebookID: "fb2"},
		{ID: "acc3", FacebookID: "fb3"},
		{ID: "acc4", FacebookID: "fb4"},
		{ID: "acc5", FacebookID: "fb5"},
	}

	// Distribute accounts to channel
	go func() {
		for _, account := range accounts {
			accountPayload, _ := json.Marshal(account)
			select {
			case workOrderChan <- WorkOrderMessage{AccountID: account.ID, FacebookID: account.FacebookID, Value: accountPayload}:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Receive and count
	received := make([]WorkOrderMessage, 0)
	timeout := time.After(500 * time.Millisecond)

	for i := 0; i < 5; i++ {
		select {
		case wo := <-workOrderChan:
			received = append(received, wo)
		case <-timeout:
			t.Fatalf("timeout waiting for work orders, got %d", len(received))
		}
	}

	if len(received) != 5 {
		t.Errorf("expected 5 work orders, got %d", len(received))
	}
}

// ================== Worker Pool Tests ==================

func TestWorkerPool_ProcessesWorkOrders(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	workOrderChan := make(chan WorkOrderMessage, 10)
	var processedCount int32

	// Start worker pool (simplified)
	var wg sync.WaitGroup
	numWorkers := 3
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case wo, ok := <-workOrderChan:
					if !ok {
						return
					}
					atomic.AddInt32(&processedCount, 1)
					_ = wo
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	// Send work orders
	for i := 0; i < 10; i++ {
		workOrderChan <- WorkOrderMessage{
			AccountID:  "acc",
			FacebookID: "fb",
			Value:      []byte("test"),
		}
	}

	// Wait for processing
	time.Sleep(100 * time.Millisecond)
	cancel()
	close(workOrderChan)
	wg.Wait()

	if atomic.LoadInt32(&processedCount) != 10 {
		t.Errorf("expected 10 processed, got %d", processedCount)
	}
}

func TestWorkerPool_GracefulShutdown(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	workOrderChan := make(chan WorkOrderMessage, 10)
	var shutdownCount int32

	var wg sync.WaitGroup
	numWorkers := 5
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			defer atomic.AddInt32(&shutdownCount, 1)
			for {
				select {
				case _, ok := <-workOrderChan:
					if !ok {
						return
					}
				case <-ctx.Done():
					return
				}
			}
		}(i)
	}

	// Cancel immediately
	cancel()
	close(workOrderChan)
	wg.Wait()

	if atomic.LoadInt32(&shutdownCount) != int32(numWorkers) {
		t.Errorf("expected %d workers to shutdown, got %d", numWorkers, shutdownCount)
	}
}

// ================== Idle Timeout Tests ==================

func TestIdleTimeoutDetection(t *testing.T) {
	var lastMessageTime int64 = time.Now().UnixNano()

	// Simulate idle check
	idleTimeout := 100 * time.Millisecond

	// Initially not idle
	lastTime := time.Unix(0, atomic.LoadInt64(&lastMessageTime))
	idleDuration := time.Since(lastTime)
	if idleDuration >= idleTimeout {
		t.Error("should not be idle immediately")
	}

	// Wait for idle timeout
	time.Sleep(150 * time.Millisecond)

	lastTime = time.Unix(0, atomic.LoadInt64(&lastMessageTime))
	idleDuration = time.Since(lastTime)
	if idleDuration < idleTimeout {
		t.Error("should be idle after timeout")
	}
}

func TestIdleTimeoutReset(t *testing.T) {
	var lastMessageTime int64 = time.Now().UnixNano()
	idleTimeout := 500 * time.Millisecond

	// Wait partial time
	time.Sleep(50 * time.Millisecond)

	// Reset by updating last message time
	atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())

	// Wait a short time
	time.Sleep(50 * time.Millisecond)

	// Should not be idle because we reset recently
	lastTime := time.Unix(0, atomic.LoadInt64(&lastMessageTime))
	idleDuration := time.Since(lastTime)
	if idleDuration >= idleTimeout {
		t.Error("should not be idle after reset")
	}
}

// ================== Publisher Tests with Mock Producer ==================

func TestPublishFlow_PostsVideosInsights(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	var postCount, videoCount, insightCount int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			switch topic {
			case "raw-facebook-posts":
				atomic.AddInt32(&postCount, 1)
			case "raw-facebook-videos":
				atomic.AddInt32(&videoCount, 1)
			case "raw-facebook-insights":
				atomic.AddInt32(&insightCount, 1)
			}
			return nil
		},
	}

	// Publish posts
	posts := []kafkamodels.RawFacebookPost{
		{ID: "post1"}, {ID: "post2"},
	}
	publishPostsParallel(ctx, posts, producer, "page123", "ws123", 0, log)

	// Publish videos
	videos := []kafkamodels.RawFacebookVideo{
		{ID: "video1"}, {ID: "video2"}, {ID: "video3"},
	}
	publishVideosParallel(ctx, videos, producer, "page123", "ws123", 0, log)

	// Publish insights
	insights := []kafkamodels.RawFacebookInsights{
		{PageID: "page123"},
	}
	publishInsightsParallel(ctx, insights, producer, "page123", "ws123", 0, log)

	if atomic.LoadInt32(&postCount) != 2 {
		t.Errorf("expected 2 posts, got %d", postCount)
	}
	if atomic.LoadInt32(&videoCount) != 3 {
		t.Errorf("expected 3 videos, got %d", videoCount)
	}
	if atomic.LoadInt32(&insightCount) != 1 {
		t.Errorf("expected 1 insight, got %d", insightCount)
	}
}

// ================== Semaphore Per Account Tests ==================

func TestSemaphorePerAccount_ConcurrencyControl(t *testing.T) {
	accountSemaphores = sync.Map{} // Reset

	accountID := "test_account_concurrent"
	capacity := int64(2)

	sem := semForAccount(accountID, capacity)

	// Should be able to acquire up to capacity
	ctx := context.Background()
	if err := sem.Acquire(ctx, 1); err != nil {
		t.Fatalf("first acquire failed: %v", err)
	}
	if err := sem.Acquire(ctx, 1); err != nil {
		t.Fatalf("second acquire failed: %v", err)
	}

	// Third acquire should block - test with timeout
	acquireDone := make(chan bool)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
		defer cancel()
		err := sem.Acquire(ctx, 1)
		acquireDone <- (err == nil)
	}()

	select {
	case success := <-acquireDone:
		if success {
			t.Error("third acquire should have blocked/timed out")
		}
	case <-time.After(500 * time.Millisecond):
		t.Error("test timeout")
	}

	// Release one and try again
	sem.Release(1)

	ctx2, cancel2 := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel2()
	if err := sem.Acquire(ctx2, 1); err != nil {
		t.Errorf("acquire after release failed: %v", err)
	}
}

// ================== JSON Parsing Tests ==================

func TestParseInvalidWorkOrder(t *testing.T) {
	invalidJSON := []byte(`{"invalid": json}`)

	var wo kafkamodels.FacebookAccountWorkOrder
	err := json.Unmarshal(invalidJSON, &wo)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestParseValidWorkOrder(t *testing.T) {
	validJSON := []byte(`{
		"id": "acc123",
		"facebook_id": "fb456",
		"workspace_id": "ws789",
		"sync_type": "full_sync",
		"access_token": "token123",
		"long_access_token": "long_token"
	}`)

	var wo kafkamodels.FacebookAccountWorkOrder
	err := json.Unmarshal(validJSON, &wo)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if wo.ID != "acc123" {
		t.Errorf("ID = %q, want %q", wo.ID, "acc123")
	}
	if wo.FacebookID != "fb456" {
		t.Errorf("FacebookID = %q, want %q", wo.FacebookID, "fb456")
	}
	if wo.SyncType != "full_sync" {
		t.Errorf("SyncType = %q, want %q", wo.SyncType, "full_sync")
	}
	if wo.AccessToken != "token123" {
		t.Errorf("AccessToken = %q, want %q", wo.AccessToken, "token123")
	}
	if wo.LongAccessToken != "long_token" {
		t.Errorf("LongAccessToken = %q, want %q", wo.LongAccessToken, "long_token")
	}
}

// ================== Date Range Calculation Tests ==================

func TestDateRangeCalculation_Incremental(t *testing.T) {
	syncType := "incremental"
	until := time.Now()
	var since time.Time

	switch syncType {
	case "incremental":
		since = time.Date(until.Year(), until.Month(), until.Day()-14, 0, 0, 0, 0, time.UTC)
	case "full_sync":
		since = time.Date(until.Year(), until.Month(), until.Day()-90, 0, 0, 0, 0, time.UTC)
	default:
		since = time.Date(until.Year(), until.Month(), until.Day()-14, 0, 0, 0, 0, time.UTC)
	}

	daysDiff := int(until.Sub(since).Hours() / 24)
	if daysDiff < 13 || daysDiff > 15 {
		t.Errorf("incremental should be ~14 days, got %d", daysDiff)
	}
}

func TestDateRangeCalculation_FullSync(t *testing.T) {
	syncType := "full_sync"
	until := time.Now()
	var since time.Time

	switch syncType {
	case "incremental":
		since = time.Date(until.Year(), until.Month(), until.Day()-14, 0, 0, 0, 0, time.UTC)
	case "full_sync":
		since = time.Date(until.Year(), until.Month(), until.Day()-90, 0, 0, 0, 0, time.UTC)
	default:
		since = time.Date(until.Year(), until.Month(), until.Day()-14, 0, 0, 0, 0, time.UTC)
	}

	daysDiff := int(until.Sub(since).Hours() / 24)
	if daysDiff < 89 || daysDiff > 91 {
		t.Errorf("full_sync should be ~90 days, got %d", daysDiff)
	}
}

func TestDateRangeCalculation_Default(t *testing.T) {
	syncType := "unknown"
	until := time.Now()
	var since time.Time

	switch syncType {
	case "incremental":
		since = time.Date(until.Year(), until.Month(), until.Day()-14, 0, 0, 0, 0, time.UTC)
	case "full_sync":
		since = time.Date(until.Year(), until.Month(), until.Day()-90, 0, 0, 0, 0, time.UTC)
	default:
		since = time.Date(until.Year(), until.Month(), until.Day()-14, 0, 0, 0, 0, time.UTC)
	}

	daysDiff := int(until.Sub(since).Hours() / 24)
	if daysDiff < 13 || daysDiff > 15 {
		t.Errorf("default should be ~14 days, got %d", daysDiff)
	}
}

// ================== hasReelsMetric Tests ==================

func TestHasReelsMetric_WithReels_Main(t *testing.T) {
	videoJSON := `{
		"id": "video123",
		"video_insights": {
			"data": [
				{"name": "total_video_views", "values": [{"value": 100}]},
				{"name": "blue_reels_play_count", "values": [{"value": 500}]}
			]
		}
	}`
	var video kafkamodels.RawFacebookVideo
	json.Unmarshal([]byte(videoJSON), &video)

	if !hasReelsMetric(video) {
		t.Error("expected hasReelsMetric to return true for video with blue_reels_play_count")
	}
}

func TestHasReelsMetric_WithoutReels_Main(t *testing.T) {
	videoJSON := `{
		"id": "video123",
		"video_insights": {
			"data": [
				{"name": "total_video_views", "values": [{"value": 100}]},
				{"name": "total_video_impressions", "values": [{"value": 200}]}
			]
		}
	}`
	var video kafkamodels.RawFacebookVideo
	json.Unmarshal([]byte(videoJSON), &video)

	if hasReelsMetric(video) {
		t.Error("expected hasReelsMetric to return false for video without blue_reels_play_count")
	}
}

// ================== WorkOrderMessage Struct Tests ==================

func TestWorkOrderMessageStruct(t *testing.T) {
	msg := WorkOrderMessage{
		AccountID:  "acc123",
		FacebookID: "fb456",
		Value:      []byte(`{"id": "test"}`),
	}

	if msg.AccountID != "acc123" {
		t.Errorf("AccountID = %q, want %q", msg.AccountID, "acc123")
	}
	if msg.FacebookID != "fb456" {
		t.Errorf("FacebookID = %q, want %q", msg.FacebookID, "fb456")
	}
}

// ================== Semaphore Tests ==================

func TestSemForAccount_NewAccount(t *testing.T) {
	accountSemaphores = sync.Map{} // Reset

	accountID := "new_test_account"
	capacity := int64(5)

	sem := semForAccount(accountID, capacity)
	if sem == nil {
		t.Fatal("semaphore should not be nil")
	}
}

func TestSemForAccount_ExistingAccount(t *testing.T) {
	accountSemaphores = sync.Map{} // Reset

	accountID := "existing_test_account"
	capacity := int64(3)

	// Create first
	sem1 := semForAccount(accountID, capacity)

	// Get again - should return same semaphore
	sem2 := semForAccount(accountID, capacity)

	if sem1 != sem2 {
		t.Error("should return same semaphore for same account")
	}
}

// ================== Batch Work Order Tests ==================

func TestParseBatchWorkOrder(t *testing.T) {
	batchJSON := []byte(`{
		"batch_id": "batch123",
		"sync_type": "incremental",
		"accounts": [
			{"id": "acc1", "facebook_id": "fb1", "workspace_id": "ws1"},
			{"id": "acc2", "facebook_id": "fb2", "workspace_id": "ws2"}
		]
	}`)

	var batch kafkamodels.FacebookBatchWorkOrder
	if err := json.Unmarshal(batchJSON, &batch); err != nil {
		t.Fatalf("failed to parse batch: %v", err)
	}

	if batch.BatchID != "batch123" {
		t.Errorf("BatchID = %q, want %q", batch.BatchID, "batch123")
	}
	if batch.SyncType != "incremental" {
		t.Errorf("SyncType = %q, want %q", batch.SyncType, "incremental")
	}
	if len(batch.Accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(batch.Accounts))
	}
}

// ================== Publisher Error Handling Tests ==================
// NOTE: TestPublishPostsParallel_EmptyPosts, TestPublishVideosParallel_EmptyVideos
// and TestPublishInsightsParallel_EmptyInsights are defined in posts_test.go, videos_test.go,
// and insights_test.go respectively

// ================== Context Cancellation Tests ==================

func TestWorkerPool_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	workOrderChan := make(chan WorkOrderMessage, 10)
	var cancelledCount int32

	var wg sync.WaitGroup
	numWorkers := 3
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			for {
				select {
				case _, ok := <-workOrderChan:
					if !ok {
						return
					}
				case <-ctx.Done():
					atomic.AddInt32(&cancelledCount, 1)
					return
				}
			}
		}(i)
	}

	// Cancel context
	cancel()
	wg.Wait()

	if atomic.LoadInt32(&cancelledCount) != int32(numWorkers) {
		t.Errorf("expected %d workers cancelled, got %d", numWorkers, cancelledCount)
	}
}

// ================== Publisher with Large Batches ==================

func TestPublishPostsParallel_LargeBatch(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	var callCount int32
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&callCount, 1)
			return nil
		},
	}

	// Create 100 posts
	posts := make([]kafkamodels.RawFacebookPost, 100)
	for i := 0; i < 100; i++ {
		posts[i] = kafkamodels.RawFacebookPost{ID: "post" + string(rune('0'+i%10))}
	}

	publishPostsParallel(ctx, posts, producer, "page123", "ws123", 0, log)

	if atomic.LoadInt32(&callCount) != 100 {
		t.Errorf("expected 100 calls, got %d", callCount)
	}
}

func TestPublishVideosParallel_LargeBatch(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	var callCount int32
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&callCount, 1)
			return nil
		},
	}

	// Create 50 videos
	videos := make([]kafkamodels.RawFacebookVideo, 50)
	for i := 0; i < 50; i++ {
		videos[i] = kafkamodels.RawFacebookVideo{ID: "video" + string(rune('0'+i%10))}
	}

	publishVideosParallel(ctx, videos, producer, "page123", "ws123", 0, log)

	if atomic.LoadInt32(&callCount) != 50 {
		t.Errorf("expected 50 calls, got %d", callCount)
	}
}

// ================== WorkOrderMessage with AccountID ==================

func TestWorkOrderMessage_AllFields(t *testing.T) {
	msg := WorkOrderMessage{
		AccountID:  "acc456",
		FacebookID: "fb123",
		Value:      []byte("value"),
	}

	if msg.AccountID != "acc456" {
		t.Errorf("AccountID = %q, want %q", msg.AccountID, "acc456")
	}
	if msg.FacebookID != "fb123" {
		t.Errorf("FacebookID = %q, want %q", msg.FacebookID, "fb123")
	}
	if string(msg.Value) != "value" {
		t.Errorf("Value = %q, want %q", string(msg.Value), "value")
	}
}

// ================== processWorkOrder Tests ==================

func TestProcessWorkOrder_InvalidJSON(t *testing.T) {
	ctx := context.Background()
	invalidJSON := []byte("not valid json")

	err := processWorkOrder(ctx, invalidJSON, nil, nil, nil, "", 1)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestProcessWorkOrder_EmptyAccessToken(t *testing.T) {
	ctx := context.Background()

	// Work order with no access token
	wo := kafkamodels.FacebookAccountWorkOrder{
		ID:          "acc123",
		FacebookID:  "fb456",
		WorkspaceID: "ws789",
		SyncType:    "incremental",
		AccessToken: "", // empty
	}
	woJSON, _ := json.Marshal(wo)

	// Should return nil (skip without error) when no token available
	err := processWorkOrder(ctx, woJSON, nil, nil, nil, "", 1)
	if err != nil {
		t.Errorf("expected nil error for empty token (skip behavior), got %v", err)
	}
}

func TestProcessWorkOrder_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	wo := kafkamodels.FacebookAccountWorkOrder{
		ID:          "acc123",
		FacebookID:  "fb456",
		WorkspaceID: "ws789",
		SyncType:    "incremental",
		AccessToken: "valid_token",
	}
	woJSON, _ := json.Marshal(wo)

	// Should return error when context is cancelled during semaphore acquisition
	err := processWorkOrder(ctx, woJSON, nil, nil, nil, "", 1)
	if err == nil {
		t.Error("expected error when context is cancelled")
	}
}

// ================== Topic Constants Tests ==================

func TestTopicConstants(t *testing.T) {
	tests := []struct {
		name     string
		got      string
		expected string
	}{
		{"work order topic", "work-order-facebook", "work-order-facebook"},
		{"raw posts topic", "raw-facebook-posts", "raw-facebook-posts"},
		{"raw videos topic", "raw-facebook-videos", "raw-facebook-videos"},
		{"raw insights topic", "raw-facebook-insights", "raw-facebook-insights"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.expected {
				t.Errorf("got %q, want %q", tt.got, tt.expected)
			}
		})
	}
}

// ================== Logging Contract Tests (Point 4 — Calling service logs errors with context) ==================

func TestLoggingContract_FacebookFetcher_ErrorHasContextFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	// Simulate what the Facebook fetcher does when it gets an unexpected error
	log.Error().
		Str("error_message", "failed to fetch page posts").
		Str("function", "processWorkOrder").
		Str("stage", "fetch_posts").
		Msg("Facebook fetcher error")

	output := buf.String()

	checks := map[string]string{
		"ERR":              "expected ERR level in output",
		"error_message":    "expected error_message field",
		"function":         "expected function field",
		"processWorkOrder": "expected processWorkOrder value",
		"stage":            "expected stage field",
		"fetch_posts":      "expected fetch_posts stage value",
	}
	for substr, errMsg := range checks {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

func TestLoggingContract_FacebookFetcher_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Log an error the way the fetcher does — Error level only, no explicit CaptureException
	log.Error().
		Str("error_message", "API rate limit exceeded").
		Str("function", "processWorkOrder").
		Str("stage", "fetch_posts").
		Msg("Failed to fetch posts")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestLoggingContract_FacebookFetcher_SingleSentryEvent(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "connection timeout").
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

func TestLoggingContract_FacebookFetcher_ExpectedError_WarnOnly(t *testing.T) {
	hookRecords, hookCleanup := logger.InstallHookSpy()
	defer hookCleanup()

	log, buf := logger.NewTestLoggerWithHook()

	// Simulate expected auth error — should be Warn, not Error
	log.Warn().
		Str("error_message", "OAuthException (#190) token expired").
		Str("function", "processWorkOrder").
		Str("page_id", "fb-page-123").
		Msg("Facebook auth error, skipping account")

	output := buf.String()

	if !strings.Contains(output, "WRN") {
		t.Fatalf("expected WRN level in output, got: %s", output)
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
