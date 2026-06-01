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
		posts:    make(chan *chmodels.TwitterPosts, 10),
		insights: make(chan *chmodels.TwitterInsights, 10),
	}

	if bc.posts == nil {
		t.Error("posts channel is nil")
	}
	if bc.insights == nil {
		t.Error("insights channel is nil")
	}
}

func TestBatchCollectors_Creation(t *testing.T) {
	cfg := DefaultServiceConfig()
	bc := &BatchCollectors{
		posts:    make(chan *chmodels.TwitterPosts, cfg.MaxBatchSize*5),
		insights: make(chan *chmodels.TwitterInsights, cfg.MaxBatchSize*5),
	}

	if cap(bc.posts) != cfg.MaxBatchSize*5 {
		t.Errorf("posts capacity = %d, want %d", cap(bc.posts), cfg.MaxBatchSize*5)
	}
	if cap(bc.insights) != cfg.MaxBatchSize*5 {
		t.Errorf("insights capacity = %d, want %d", cap(bc.insights), cfg.MaxBatchSize*5)
	}
}

// ================== ServiceMetrics Tests ==================

func TestServiceMetrics_Initialization(t *testing.T) {
	metrics := &ServiceMetrics{}

	if atomic.LoadUint64(&metrics.PostsMessagesReceived) != 0 {
		t.Error("PostsMessagesReceived should be 0 initially")
	}
	if atomic.LoadUint64(&metrics.InsightsMessagesReceived) != 0 {
		t.Error("InsightsMessagesReceived should be 0 initially")
	}
	if atomic.LoadUint64(&metrics.PostsMessagesParsed) != 0 {
		t.Error("PostsMessagesParsed should be 0 initially")
	}
	if atomic.LoadUint64(&metrics.InsightsMessagesParsed) != 0 {
		t.Error("InsightsMessagesParsed should be 0 initially")
	}
}

func TestServiceMetrics_AtomicOperations(t *testing.T) {
	metrics := &ServiceMetrics{}

	atomic.AddUint64(&metrics.PostsMessagesReceived, 5)
	if atomic.LoadUint64(&metrics.PostsMessagesReceived) != 5 {
		t.Errorf("PostsMessagesReceived = %d, want 5", atomic.LoadUint64(&metrics.PostsMessagesReceived))
	}

	atomic.AddUint64(&metrics.InsightsMessagesReceived, 3)
	if atomic.LoadUint64(&metrics.InsightsMessagesReceived) != 3 {
		t.Errorf("InsightsMessagesReceived = %d, want 3", atomic.LoadUint64(&metrics.InsightsMessagesReceived))
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
			handler(ctx, topics[0], []byte("key"), []byte(`{"data": {}, "twitter_id": "test123"}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Simulate consumer behavior
	go func() {
		mockConsumer.Consume(ctx, []string{"raw-twitter-posts"}, func(ctx context.Context, topic string, key, value []byte) error {
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
			handler(ctx, topics[0], []byte("key"), []byte(`{"data": {}, "twitter_id": "test456"}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		mockConsumer.Consume(ctx, []string{"raw-twitter-insights"}, func(ctx context.Context, topic string, key, value []byte) error {
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

// ================== ServiceDependencies Tests ==================

func TestServiceDependencies_Creation(t *testing.T) {
	log := logger.New("error")
	mockSink := &MockClickHouseSink{}
	mockConsumer := &kafka.MockConsumer{}

	deps := &ServiceDependencies{
		Sink:             mockSink,
		PostsConsumer:    mockConsumer,
		InsightsConsumer: mockConsumer,
		Logger:           log,
	}

	if deps.Sink == nil {
		t.Error("Sink should not be nil")
	}
	if deps.PostsConsumer == nil {
		t.Error("PostsConsumer should not be nil")
	}
	if deps.InsightsConsumer == nil {
		t.Error("InsightsConsumer should not be nil")
	}
	if deps.Logger == nil {
		t.Error("Logger should not be nil")
	}
}

// ================== Interface Compliance Tests ==================

func TestMockClickHouseSink_Implements_ClickHouseSinkInterface(t *testing.T) {
	var _ ClickHouseSinkInterface = (*MockClickHouseSink)(nil)
}

func TestMockConsumer_Implements_KafkaConsumerInterface(t *testing.T) {
	var _ KafkaConsumerInterface = (*kafka.MockConsumer)(nil)
}
