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
	BulkInsertDailyMetricsCount   uint64
	BulkInsertMediaAssetsCount    uint64
	BulkInsertSearchKeywordsCount uint64
	BulkInsertLocalPostsCount     uint64
	BulkInsertReviewsCount        uint64

	BulkInsertDailyMetricsFunc   func(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error
	BulkInsertMediaAssetsFunc    func(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error
	BulkInsertSearchKeywordsFunc func(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error
	BulkInsertLocalPostsFunc     func(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error
	BulkInsertReviewsFunc        func(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error
}

func (m *MockClickHouseSink) BulkInsertGMBDailyMetrics(ctx context.Context, metrics []*clickhousemodels.GMBDailyMetrics) error {
	atomic.AddUint64(&m.BulkInsertDailyMetricsCount, 1)
	if m.BulkInsertDailyMetricsFunc != nil {
		return m.BulkInsertDailyMetricsFunc(ctx, metrics)
	}
	return nil
}

func (m *MockClickHouseSink) BulkInsertGMBMediaAssets(ctx context.Context, assets []*clickhousemodels.GMBMediaAssets) error {
	atomic.AddUint64(&m.BulkInsertMediaAssetsCount, 1)
	if m.BulkInsertMediaAssetsFunc != nil {
		return m.BulkInsertMediaAssetsFunc(ctx, assets)
	}
	return nil
}

func (m *MockClickHouseSink) BulkInsertGMBSearchKeywordsMonthly(ctx context.Context, keywords []*clickhousemodels.GMBSearchKeywordsMonthly) error {
	atomic.AddUint64(&m.BulkInsertSearchKeywordsCount, 1)
	if m.BulkInsertSearchKeywordsFunc != nil {
		return m.BulkInsertSearchKeywordsFunc(ctx, keywords)
	}
	return nil
}

func (m *MockClickHouseSink) BulkInsertGMBLocalPosts(ctx context.Context, posts []*clickhousemodels.GMBLocalPosts) error {
	atomic.AddUint64(&m.BulkInsertLocalPostsCount, 1)
	if m.BulkInsertLocalPostsFunc != nil {
		return m.BulkInsertLocalPostsFunc(ctx, posts)
	}
	return nil
}

func (m *MockClickHouseSink) BulkInsertGMBReviews(ctx context.Context, reviews []*clickhousemodels.GMBReviews) error {
	atomic.AddUint64(&m.BulkInsertReviewsCount, 1)
	if m.BulkInsertReviewsFunc != nil {
		return m.BulkInsertReviewsFunc(ctx, reviews)
	}
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
