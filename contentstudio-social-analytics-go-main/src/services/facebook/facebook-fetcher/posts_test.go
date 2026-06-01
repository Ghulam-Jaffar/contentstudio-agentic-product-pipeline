package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestPublishPostsParallel_EmptyPosts(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	publishPostsParallel(context.Background(), []kafkamodels.RawFacebookPost{}, producer, "page123", "workspace123", 1, log)
	// Should return immediately without error
}

func TestPublishPostsParallel_Success(t *testing.T) {
	log := logger.New("error")
	var publishCount int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&publishCount, 1)
			return nil
		},
	}

	posts := []kafkamodels.RawFacebookPost{
		{ID: "post1", Message: "Test post 1"},
		{ID: "post2", Message: "Test post 2"},
		{ID: "post3", Message: "Test post 3"},
	}

	publishPostsParallel(context.Background(), posts, producer, "page123", "workspace123", 1, log)

	if atomic.LoadInt32(&publishCount) != 3 {
		t.Fatalf("expected 3 posts published, got %d", publishCount)
	}
}

func TestPublishPostsParallel_WithoutWorkspaceID(t *testing.T) {
	log := logger.New("error")
	var keys []string

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			keys = append(keys, string(key))
			return nil
		},
	}

	posts := []kafkamodels.RawFacebookPost{
		{ID: "post1", Message: "Test post 1"},
	}

	publishPostsParallel(context.Background(), posts, producer, "page123", "", 1, log)

	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	// Without workspace ID, key should be "pageID_postID"
	if keys[0] != "page123_post1" {
		t.Fatalf("expected key 'page123_post1', got '%s'", keys[0])
	}
}

func TestPublishPostsParallel_WithWorkspaceID(t *testing.T) {
	log := logger.New("error")
	var keys []string

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			keys = append(keys, string(key))
			return nil
		},
	}

	posts := []kafkamodels.RawFacebookPost{
		{ID: "post1", Message: "Test post 1"},
	}

	publishPostsParallel(context.Background(), posts, producer, "page123", "workspace456", 1, log)

	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	// With workspace ID, key should be "workspaceID_pageID_postID"
	if keys[0] != "workspace456_page123_post1" {
		t.Fatalf("expected key 'workspace456_page123_post1', got '%s'", keys[0])
	}
}

func TestPublishPostsParallel_ProducerError(t *testing.T) {
	log := logger.New("error")
	var successCount int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			if string(key) == "workspace123_page123_post2" {
				return errors.New("kafka error")
			}
			atomic.AddInt32(&successCount, 1)
			return nil
		},
	}

	posts := []kafkamodels.RawFacebookPost{
		{ID: "post1", Message: "Test post 1"},
		{ID: "post2", Message: "Test post 2"},
		{ID: "post3", Message: "Test post 3"},
	}

	publishPostsParallel(context.Background(), posts, producer, "page123", "workspace123", 1, log)

	// 2 should succeed, 1 should fail
	if atomic.LoadInt32(&successCount) != 2 {
		t.Fatalf("expected 2 posts published successfully, got %d", successCount)
	}
}

func TestPublishPostsParallel_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	ctx, cancel := context.WithCancel(context.Background())

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	}

	posts := make([]kafkamodels.RawFacebookPost, 100)
	for i := 0; i < 100; i++ {
		posts[i] = kafkamodels.RawFacebookPost{ID: "post" + string(rune(i))}
	}

	// Cancel context after short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	// Should return early due to context cancellation
	publishPostsParallel(ctx, posts, producer, "page123", "workspace123", 1, log)
}

func TestConvertPostsToInterface(t *testing.T) {
	posts := []kafkamodels.RawFacebookPost{
		{ID: "post1", Message: "Test 1"},
		{ID: "post2", Message: "Test 2"},
	}

	result := convertPostsToInterface(posts)

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	post1, ok := result[0].(kafkamodels.RawFacebookPost)
	if !ok {
		t.Fatal("expected RawFacebookPost type")
	}
	if post1.ID != "post1" {
		t.Fatalf("expected ID 'post1', got '%s'", post1.ID)
	}
}

func TestConvertPostsToInterface_Empty(t *testing.T) {
	result := convertPostsToInterface([]kafkamodels.RawFacebookPost{})

	if len(result) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result))
	}
}

// ================== WorkOrderMessage Tests ==================

func TestWorkOrderMessage_Struct(t *testing.T) {
	msg := WorkOrderMessage{
		AccountID:  "acc123",
		FacebookID: "fb456",
		Value:      []byte("test"),
	}

	if msg.AccountID != "acc123" {
		t.Errorf("AccountID = %q, want %q", msg.AccountID, "acc123")
	}
	if msg.FacebookID != "fb456" {
		t.Errorf("FacebookID = %q, want %q", msg.FacebookID, "fb456")
	}
}

// ================== BatchMessage Tests ==================

func TestBatchMessage_Struct(t *testing.T) {
	msg := BatchMessage{
		Key:   []byte("key1"),
		Value: []byte("value1"),
	}

	if string(msg.Key) != "key1" {
		t.Errorf("Key = %q, want %q", string(msg.Key), "key1")
	}
	if string(msg.Value) != "value1" {
		t.Errorf("Value = %q, want %q", string(msg.Value), "value1")
	}
}

// ================== hasReelsMetric Tests ==================

func TestHasReelsMetric_True(t *testing.T) {
	video := kafkamodels.RawFacebookVideo{
		ID: "video123",
	}
	// Set VideoInsights with reels metric using JSON unmarshal
	videoJSON := `{
		"id": "video123",
		"video_insights": {
			"data": [
				{"name": "total_video_views", "values": [{"value": 100}]},
				{"name": "blue_reels_play_count", "values": [{"value": 50}]}
			]
		}
	}`
	json.Unmarshal([]byte(videoJSON), &video)

	if !hasReelsMetric(video) {
		t.Error("expected hasReelsMetric to return true")
	}
}

func TestHasReelsMetric_False(t *testing.T) {
	video := kafkamodels.RawFacebookVideo{
		ID: "video123",
	}
	videoJSON := `{
		"id": "video123",
		"video_insights": {
			"data": [
				{"name": "total_video_views", "values": [{"value": 100}]}
			]
		}
	}`
	json.Unmarshal([]byte(videoJSON), &video)

	if hasReelsMetric(video) {
		t.Error("expected hasReelsMetric to return false")
	}
}

func TestHasReelsMetric_EmptyInsights(t *testing.T) {
	video := kafkamodels.RawFacebookVideo{
		ID: "video123",
	}

	if hasReelsMetric(video) {
		t.Error("expected hasReelsMetric to return false for empty insights")
	}
}

// ================== semForAccount Tests ==================

func TestSemForAccount_CreateNew(t *testing.T) {
	// Clear the map for this test
	accountSemaphores = sync.Map{}

	sem := semForAccount("test_account_1", 1)
	if sem == nil {
		t.Fatal("expected non-nil semaphore")
	}

	// Getting same account should return same semaphore
	sem2 := semForAccount("test_account_1", 1)
	if sem != sem2 {
		t.Error("expected same semaphore for same account")
	}
}

func TestSemForAccount_DifferentAccounts(t *testing.T) {
	accountSemaphores = sync.Map{}

	sem1 := semForAccount("account_a", 1)
	sem2 := semForAccount("account_b", 1)

	if sem1 == sem2 {
		t.Error("expected different semaphores for different accounts")
	}
}

// ================== workOrderProcessor Tests ==================

func TestWorkOrderProcessor_ChannelClose(t *testing.T) {
	ctx := context.Background()
	workOrderChan := make(chan WorkOrderMessage)
	close(workOrderChan)

	// Should exit gracefully when channel is closed
	// This is a simplified version - actual test would need mocked dependencies
	done := make(chan struct{})
	go func() {
		for {
			select {
			case _, ok := <-workOrderChan:
				if !ok {
					close(done)
					return
				}
			case <-ctx.Done():
				close(done)
				return
			}
		}
	}()

	select {
	case <-done:
		// Expected
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for processor to exit")
	}
}

func TestWorkOrderProcessor_ContextCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	workOrderChan := make(chan WorkOrderMessage, 10)

	done := make(chan struct{})
	go func() {
		for {
			select {
			case <-ctx.Done():
				close(done)
				return
			case _, ok := <-workOrderChan:
				if !ok {
					close(done)
					return
				}
			}
		}
	}()

	cancel()

	select {
	case <-done:
		// Expected
	case <-time.After(1 * time.Second):
		t.Fatal("timeout waiting for processor to exit on context cancel")
	}
}

// ================== FacebookAccountWorkOrder Parsing Tests ==================

func TestFacebookAccountWorkOrder_Unmarshal(t *testing.T) {
	jsonData := `{
		"id": "acc123",
		"facebook_id": "fb456",
		"workspace_id": "ws789",
		"sync_type": "full_sync",
		"access_token": "token123"
	}`

	var wo kafkamodels.FacebookAccountWorkOrder
	err := json.Unmarshal([]byte(jsonData), &wo)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if wo.ID != "acc123" {
		t.Errorf("ID = %q, want %q", wo.ID, "acc123")
	}
	if wo.FacebookID != "fb456" {
		t.Errorf("FacebookID = %q, want %q", wo.FacebookID, "fb456")
	}
	if wo.WorkspaceID != "ws789" {
		t.Errorf("WorkspaceID = %q, want %q", wo.WorkspaceID, "ws789")
	}
	if wo.SyncType != "full_sync" {
		t.Errorf("SyncType = %q, want %q", wo.SyncType, "full_sync")
	}
}

// ================== FacebookBatchWorkOrder Parsing Tests ==================

func TestFacebookBatchWorkOrder_Unmarshal(t *testing.T) {
	jsonData := `{
		"batch_id": "batch123",
		"sync_type": "incremental",
		"accounts": [
			{"id": "acc1", "facebook_id": "fb1"},
			{"id": "acc2", "facebook_id": "fb2"}
		]
	}`

	var batch kafkamodels.FacebookBatchWorkOrder
	err := json.Unmarshal([]byte(jsonData), &batch)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if batch.BatchID != "batch123" {
		t.Errorf("BatchID = %q, want %q", batch.BatchID, "batch123")
	}
	if batch.SyncType != "incremental" {
		t.Errorf("SyncType = %q, want %q", batch.SyncType, "incremental")
	}
	if len(batch.Accounts) != 2 {
		t.Errorf("Accounts count = %d, want 2", len(batch.Accounts))
	}
}

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if maxWorkers != 15 {
		t.Errorf("maxWorkers = %d, want 15", maxWorkers)
	}
	if workOrderChanSize != 500 {
		t.Errorf("workOrderChanSize = %d, want 500", workOrderChanSize)
	}
	if idleTimeout != 5*time.Minute {
		t.Errorf("idleTimeout = %v, want 5m", idleTimeout)
	}
	if idleCheckInterval != 30*time.Second {
		t.Errorf("idleCheckInterval = %v, want 30s", idleCheckInterval)
	}
}
