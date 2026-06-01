package main

import (
	"context"
	"sync/atomic"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

// ================== Mock ClickHouse Sink ==================

// MockClickHouseSink implements ClickHouseSinkInterface for testing
type MockClickHouseSink struct {
	BulkInsertTwitterPostsCount    uint64
	BulkInsertTwitterInsightsCount uint64
	BulkInsertPostsFunc            func(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error
	BulkInsertInsightsFunc         func(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error
}

func (m *MockClickHouseSink) BulkInsertTwitterPosts(ctx context.Context, posts []*clickhousemodels.TwitterPosts) error {
	atomic.AddUint64(&m.BulkInsertTwitterPostsCount, 1)
	if m.BulkInsertPostsFunc != nil {
		return m.BulkInsertPostsFunc(ctx, posts)
	}
	return nil
}

func (m *MockClickHouseSink) BulkInsertTwitterInsights(ctx context.Context, insights []*clickhousemodels.TwitterInsights) error {
	atomic.AddUint64(&m.BulkInsertTwitterInsightsCount, 1)
	if m.BulkInsertInsightsFunc != nil {
		return m.BulkInsertInsightsFunc(ctx, insights)
	}
	return nil
}

func (m *MockClickHouseSink) Health() error {
	return nil
}

// NewMockClickHouseSink creates a new mock sink with default implementations
func NewMockClickHouseSink() *MockClickHouseSink {
	return &MockClickHouseSink{}
}

// ================== Mock Kafka Consumer ==================

// MockKafkaConsumer implements KafkaConsumerInterface for testing
type MockKafkaConsumer struct {
	ConsumeFunc func(ctx context.Context, topics []string, handler kafka.MessageHandler) error
	CloseFunc   func() error
	Messages    []MockMessage
	Closed      bool
}

// MockMessage represents a mock Kafka message for testing
type MockMessage struct {
	Topic string
	Key   []byte
	Value []byte
}

func (m *MockKafkaConsumer) Consume(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
	if m.ConsumeFunc != nil {
		return m.ConsumeFunc(ctx, topics, handler)
	}
	for _, msg := range m.Messages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			if err := handler(ctx, msg.Topic, msg.Key, msg.Value); err != nil {
				return err
			}
		}
	}
	<-ctx.Done()
	return ctx.Err()
}

func (m *MockKafkaConsumer) ConsumeWithAck(ctx context.Context, topics []string, handler kafka.AcknowledgingMessageHandler) error {
	return nil
}

func (m *MockKafkaConsumer) Close() error {
	m.Closed = true
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// NewMockKafkaConsumerWithMessages creates a new mock consumer with pre-loaded messages
func NewMockKafkaConsumerWithMessages(messages []MockMessage) *MockKafkaConsumer {
	return &MockKafkaConsumer{
		Messages: messages,
	}
}
