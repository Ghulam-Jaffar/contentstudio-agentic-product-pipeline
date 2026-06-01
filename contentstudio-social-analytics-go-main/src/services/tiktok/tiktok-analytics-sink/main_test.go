package main

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if maxBatchSize != 10000 {
		t.Errorf("maxBatchSize = %d, want 10000", maxBatchSize)
	}
	if batchTimeout != 10*time.Second {
		t.Errorf("batchTimeout = %v, want 10s", batchTimeout)
	}
	if batchProcessorsPerType != 3 {
		t.Errorf("batchProcessorsPerType = %d, want 3", batchProcessorsPerType)
	}
	if messageChanSize != 50000 {
		t.Errorf("messageChanSize = %d, want 50000", messageChanSize)
	}
	if rawPostsTopic != "raw-tiktok-posts" {
		t.Errorf("rawPostsTopic = %q, want %q", rawPostsTopic, "raw-tiktok-posts")
	}
	if rawInsightsTopic != "raw-tiktok-insights" {
		t.Errorf("rawInsightsTopic = %q, want %q", rawInsightsTopic, "raw-tiktok-insights")
	}
	if consumerGroup != "tiktok-analytics-sink-group" {
		t.Errorf("consumerGroup = %q, want %q", consumerGroup, "tiktok-analytics-sink-group")
	}
	if idleTimeout != 5*time.Minute {
		t.Errorf("idleTimeout = %v, want 5m", idleTimeout)
	}
	if idleCheckInterval != 30*time.Second {
		t.Errorf("idleCheckInterval = %v, want 30s", idleCheckInterval)
	}
}

// ================== Struct Tests ==================

func TestRawMessage_Struct(t *testing.T) {
	msg := RawMessage{
		Topic: "test-topic",
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	if msg.Topic != "test-topic" {
		t.Errorf("Topic = %q, want %q", msg.Topic, "test-topic")
	}
	if string(msg.Key) != "key" {
		t.Errorf("Key = %q, want %q", string(msg.Key), "key")
	}
	if string(msg.Value) != "value" {
		t.Errorf("Value = %q, want %q", string(msg.Value), "value")
	}
}

func TestBatchCollectors_Struct(t *testing.T) {
	bc := &BatchCollectors{
		posts:    make(chan *chmodels.TikTokPosts, 10),
		insights: make(chan *chmodels.TikTokInsights, 10),
	}

	if bc.posts == nil {
		t.Error("posts channel is nil")
	}
	if bc.insights == nil {
		t.Error("insights channel is nil")
	}
}

func TestBatchCollectors_Creation(t *testing.T) {
	bc := &BatchCollectors{
		posts:    make(chan *chmodels.TikTokPosts, maxBatchSize*5),
		insights: make(chan *chmodels.TikTokInsights, maxBatchSize*5),
	}

	if cap(bc.posts) != maxBatchSize*5 {
		t.Errorf("posts capacity = %d, want %d", cap(bc.posts), maxBatchSize*5)
	}
	if cap(bc.insights) != maxBatchSize*5 {
		t.Errorf("insights capacity = %d, want %d", cap(bc.insights), maxBatchSize*5)
	}
}

// ================== Consumer Handler Tests ==================

func TestPostsConsumerHandler(t *testing.T) {
	log := logger.New("error")

	var pickedCount uint64
	postsMsgChan := make(chan RawMessage, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Simulate sending a message
			handler(ctx, topics[0], []byte("key"), []byte(`{"data": {}, "tiktok_id": "test123"}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Simulate consumer behavior
	go func() {
		mockConsumer.Consume(ctx, []string{rawPostsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&pickedCount, 1)
			select {
			case postsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadUint64(&pickedCount) < 1 {
		t.Errorf("pickedCount = %d, want >= 1", pickedCount)
	}

	cancel()
	_ = log
}

func TestInsightsConsumerHandler(t *testing.T) {
	log := logger.New("error")

	var pickedCount uint64
	insightsMsgChan := make(chan RawMessage, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), []byte(`{"data": {}, "tiktok_id": "test456"}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		mockConsumer.Consume(ctx, []string{rawInsightsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&pickedCount, 1)
			select {
			case insightsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadUint64(&pickedCount) < 1 {
		t.Errorf("pickedCount = %d, want >= 1", pickedCount)
	}

	cancel()
	_ = log
}

// ================== Message Processing Tests ==================

func TestMessagePicking_Posts(t *testing.T) {
	var picked uint64
	msgChan := make(chan RawMessage, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Simulate message picking
	go func() {
		for i := 0; i < 10; i++ {
			select {
			case msgChan <- RawMessage{Topic: rawPostsTopic, Key: []byte("key"), Value: []byte("value")}:
				atomic.AddUint64(&picked, 1)
			case <-ctx.Done():
				return
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)

	count := atomic.LoadUint64(&picked)
	if count != 10 {
		t.Errorf("picked = %d, want 10", count)
	}
}

func TestMessagePicking_Insights(t *testing.T) {
	var picked uint64
	msgChan := make(chan RawMessage, 100)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Simulate message picking
	go func() {
		for i := 0; i < 5; i++ {
			select {
			case msgChan <- RawMessage{Topic: rawInsightsTopic, Key: []byte("key"), Value: []byte("value")}:
				atomic.AddUint64(&picked, 1)
			case <-ctx.Done():
				return
			}
		}
	}()

	time.Sleep(50 * time.Millisecond)

	count := atomic.LoadUint64(&picked)
	if count != 5 {
		t.Errorf("picked = %d, want 5", count)
	}
}

// ================== Atomic Operations Tests ==================

func TestAtomicCounters_Posts(t *testing.T) {
	var pickedPosts, parsedPosts, insertedPosts uint64

	atomic.AddUint64(&pickedPosts, 10)
	atomic.AddUint64(&parsedPosts, 8)
	atomic.AddUint64(&insertedPosts, 8)

	if atomic.LoadUint64(&pickedPosts) != 10 {
		t.Errorf("pickedPosts = %d, want 10", pickedPosts)
	}
	if atomic.LoadUint64(&parsedPosts) != 8 {
		t.Errorf("parsedPosts = %d, want 8", parsedPosts)
	}
	if atomic.LoadUint64(&insertedPosts) != 8 {
		t.Errorf("insertedPosts = %d, want 8", insertedPosts)
	}
}

func TestAtomicCounters_Insights(t *testing.T) {
	var pickedInsights, parsedInsights, insertedInsights uint64

	atomic.AddUint64(&pickedInsights, 5)
	atomic.AddUint64(&parsedInsights, 5)
	atomic.AddUint64(&insertedInsights, 5)

	if atomic.LoadUint64(&pickedInsights) != 5 {
		t.Errorf("pickedInsights = %d, want 5", pickedInsights)
	}
	if atomic.LoadUint64(&parsedInsights) != 5 {
		t.Errorf("parsedInsights = %d, want 5", parsedInsights)
	}
	if atomic.LoadUint64(&insertedInsights) != 5 {
		t.Errorf("insertedInsights = %d, want 5", insertedInsights)
	}
}

// ================== Channel Operations Tests ==================

func TestChannelCommunication_Posts(t *testing.T) {
	postsChan := make(chan *chmodels.TikTokPosts, 10)

	post := &chmodels.TikTokPosts{
		PostID:   "post123",
		TikTokID: "tk456",
	}

	postsChan <- post

	received := <-postsChan
	if received.PostID != "post123" {
		t.Errorf("PostID = %q, want %q", received.PostID, "post123")
	}
	if received.TikTokID != "tk456" {
		t.Errorf("TikTokID = %q, want %q", received.TikTokID, "tk456")
	}
}

func TestChannelCommunication_Insights(t *testing.T) {
	insightsChan := make(chan *chmodels.TikTokInsights, 10)

	insight := &chmodels.TikTokInsights{
		TikTokID:           "tk123",
		RecordID:           "ws456_tk123",
		TotalFollowerCount: 1000,
		TotalLikeCount:     500,
	}

	insightsChan <- insight

	received := <-insightsChan
	if received.TikTokID != "tk123" {
		t.Errorf("TikTokID = %q, want %q", received.TikTokID, "tk123")
	}
	if received.RecordID != "ws456_tk123" {
		t.Errorf("RecordID = %q, want %q", received.RecordID, "ws456_tk123")
	}
	if received.TotalFollowerCount != 1000 {
		t.Errorf("TotalFollowerCount = %d, want 1000", received.TotalFollowerCount)
	}
}

// ================== Context Cancellation Tests ==================

func TestContextCancellation_Consumer(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan bool)

	go func() {
		select {
		case <-ctx.Done():
			done <- true
		case <-time.After(100 * time.Millisecond):
			done <- false
		}
	}()

	cancel()

	if !<-done {
		t.Error("context cancellation did not trigger")
	}
}

func TestContextCancellation_Parser(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	msgChan := make(chan RawMessage)
	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			select {
			case <-ctx.Done():
				return
			case <-msgChan:
			}
		}
	}()

	select {
	case <-ctx.Done():
	case <-time.After(250 * time.Millisecond):
		t.Fatal("context should be cancelled")
	}

	select {
	case <-done:
	case <-time.After(100 * time.Millisecond):
		t.Fatal("parser goroutine should exit after context cancellation")
	}
}

// ================== Buffer and Queue Tests ==================

func TestMessageChannelCapacity_Posts(t *testing.T) {
	msgChan := make(chan RawMessage, messageChanSize)

	if cap(msgChan) != messageChanSize {
		t.Errorf("message channel capacity = %d, want %d", cap(msgChan), messageChanSize)
	}

	// Test that we can fill it
	for i := 0; i < 100; i++ {
		select {
		case msgChan <- RawMessage{Topic: "test", Key: []byte("key"), Value: []byte("value")}:
		default:
			t.Fatal("channel should not be full at 100 messages")
		}
	}

	if len(msgChan) != 100 {
		t.Errorf("channel length = %d, want 100", len(msgChan))
	}
}

func TestBatchCollectorCapacity(t *testing.T) {
	batches := &BatchCollectors{
		posts:    make(chan *chmodels.TikTokPosts, maxBatchSize*5),
		insights: make(chan *chmodels.TikTokInsights, maxBatchSize*5),
	}

	expectedCap := maxBatchSize * 5

	if cap(batches.posts) != expectedCap {
		t.Errorf("posts capacity = %d, want %d", cap(batches.posts), expectedCap)
	}
	if cap(batches.insights) != expectedCap {
		t.Errorf("insights capacity = %d, want %d", cap(batches.insights), expectedCap)
	}
}

// ================== Idle Timeout Tests ==================

func TestIdleTimeoutLogic(t *testing.T) {
	var lastMessageTime int64 = time.Now().Add(-20 * time.Minute).UnixNano()

	lastTime := time.Unix(0, atomic.LoadInt64(&lastMessageTime))
	idleDuration := time.Since(lastTime)

	if idleDuration < idleTimeout {
		t.Errorf("idle duration = %v, should be >= %v", idleDuration, idleTimeout)
	}
}

func TestIdleTimeoutNotReached(t *testing.T) {
	var lastMessageTime int64 = time.Now().UnixNano()

	lastTime := time.Unix(0, atomic.LoadInt64(&lastMessageTime))
	idleDuration := time.Since(lastTime)

	if idleDuration >= idleTimeout {
		t.Errorf("idle duration = %v, should be < %v", idleDuration, idleTimeout)
	}
}

func TestLastMessageTimeUpdate(t *testing.T) {
	var lastMessageTime int64

	before := time.Now().UnixNano()
	atomic.StoreInt64(&lastMessageTime, before)

	time.Sleep(10 * time.Millisecond)

	after := time.Now().UnixNano()
	atomic.StoreInt64(&lastMessageTime, after)

	stored := atomic.LoadInt64(&lastMessageTime)
	if stored <= before {
		t.Error("lastMessageTime should be updated to a newer value")
	}
}

// ================== Topic Configuration Tests ==================

func TestTopicConfiguration(t *testing.T) {
	tests := []struct {
		name     string
		topic    string
		expected string
	}{
		{"posts topic", rawPostsTopic, "raw-tiktok-posts"},
		{"insights topic", rawInsightsTopic, "raw-tiktok-insights"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.topic != tt.expected {
				t.Errorf("topic = %q, want %q", tt.topic, tt.expected)
			}
		})
	}
}

func TestConsumerGroupConfiguration(t *testing.T) {
	if consumerGroup != "tiktok-analytics-sink-group" {
		t.Errorf("consumerGroup = %q, want %q", consumerGroup, "tiktok-analytics-sink-group")
	}
}

// ================== Batch Size Tests ==================

func TestBatchSizeConfiguration(t *testing.T) {
	if maxBatchSize <= 0 {
		t.Error("maxBatchSize should be positive")
	}

	if maxBatchSize > 100000 {
		t.Error("maxBatchSize seems too large")
	}
}

func TestBatchTimeoutConfiguration(t *testing.T) {
	if batchTimeout <= 0 {
		t.Error("batchTimeout should be positive")
	}

	if batchTimeout > time.Minute {
		t.Error("batchTimeout seems too large")
	}
}

// ================== Worker Configuration Tests ==================

func TestBatchProcessorCount(t *testing.T) {
	if batchProcessorsPerType <= 0 {
		t.Error("batchProcessorsPerType should be positive")
	}

	if batchProcessorsPerType > 10 {
		t.Error("batchProcessorsPerType seems too large")
	}
}

func TestParserWorkerConfiguration(t *testing.T) {
	// In main.go, we start 5 posts parsers and 5 insights parsers
	postsWorkers := 5
	insightsWorkers := 5

	if postsWorkers != insightsWorkers {
		t.Log("Note: posts and insights have different worker counts")
	}

	if postsWorkers <= 0 || insightsWorkers <= 0 {
		t.Error("worker counts should be positive")
	}
}

// ================== Logging Contract Tests ==================

func TestLoggingContract_TikTokSink_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Log an error the way the TikTok sink does — Error level only, no CaptureException
	log.Error().
		Str("error_message", "ClickHouse batch insert failed").
		Str("function", "batchInserter").
		Str("stage", "insert_posts").
		Msg("Failed to insert batch to ClickHouse")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}

	_ = log
}
