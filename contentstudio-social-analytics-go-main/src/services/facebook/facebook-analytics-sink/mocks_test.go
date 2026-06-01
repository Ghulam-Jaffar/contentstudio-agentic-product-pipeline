package main

import (
	"context"
	"sync/atomic"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// MockClickHouseSink wraps conversions.MockClickHouseSink with atomic counters for verification
type MockClickHouseSink struct {
	*conversions.MockClickHouseSink
	PostsInserted         int32
	MediaAssetsInserted   int32
	InsightsInserted      int32
	VideoInsightsInserted int32
	ReelsInsightsInserted int32
}

func (m *MockClickHouseSink) BulkInsertPosts(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
	atomic.AddInt32(&m.PostsInserted, int32(len(posts)))
	return m.MockClickHouseSink.BulkInsertPosts(ctx, posts)
}

func (m *MockClickHouseSink) BulkInsertMediaAssets(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
	atomic.AddInt32(&m.MediaAssetsInserted, int32(len(assets)))
	return m.MockClickHouseSink.BulkInsertMediaAssets(ctx, assets)
}

func (m *MockClickHouseSink) BulkInsertInsights(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
	atomic.AddInt32(&m.InsightsInserted, int32(len(insights)))
	return m.MockClickHouseSink.BulkInsertInsights(ctx, insights)
}

func (m *MockClickHouseSink) BulkInsertVideoInsights(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
	atomic.AddInt32(&m.VideoInsightsInserted, int32(len(insights)))
	return m.MockClickHouseSink.BulkInsertVideoInsights(ctx, insights)
}

func (m *MockClickHouseSink) BulkInsertReelsInsights(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
	atomic.AddInt32(&m.ReelsInsightsInserted, int32(len(insights)))
	return m.MockClickHouseSink.BulkInsertReelsInsights(ctx, insights)
}

// NewMockClickHouseSink creates a new mock sink with default implementations
func NewMockClickHouseSink() *MockClickHouseSink {
	return &MockClickHouseSink{
		MockClickHouseSink: conversions.NewMockClickHouseSink(),
	}
}

// MockMessage represents a mock Kafka message for testing
type MockMessage = kafka.MockMessage

// MockKafkaConsumer implements KafkaConsumerInterface for testing
// This wraps kafka.MockConsumerWithMessages but uses the local MessageHandler type
type MockKafkaConsumer struct {
	ConsumeFunc func(ctx context.Context, topics []string, handler kafka.MessageHandler) error
	CloseFunc   func() error
	Messages    []MockMessage
	Closed      bool
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

// NewMockKafkaConsumer creates a new mock consumer
func NewMockKafkaConsumer() *MockKafkaConsumer {
	return &MockKafkaConsumer{}
}

// ClickHouseSinkInterface verification - ensure our mock implements the interface
var _ ClickHouseSinkInterface = (*MockClickHouseSink)(nil)

// KafkaConsumerInterface verification
var _ KafkaConsumerInterface = (*MockKafkaConsumer)(nil)

// Helper function to convert kafka models for tests
func newParsedPost(postID, pageID string) *kafkamodels.ParsedFacebookPost {
	return &kafkamodels.ParsedFacebookPost{PostID: postID, PageID: pageID}
}

func newParsedMediaAsset(postID string) *kafkamodels.ParsedFacebookMediaAsset {
	return &kafkamodels.ParsedFacebookMediaAsset{PostID: postID}
}

func newParsedInsights(pageID string) *kafkamodels.ParsedFacebookInsights {
	return &kafkamodels.ParsedFacebookInsights{PageID: pageID}
}

func newParsedVideoInsights(postID string) *kafkamodels.ParsedFacebookVideoInsights {
	return &kafkamodels.ParsedFacebookVideoInsights{PostID: postID}
}

func newParsedReelsInsights(postID string) *kafkamodels.ParsedFacebookReelsInsights {
	return &kafkamodels.ParsedFacebookReelsInsights{PostID: postID}
}
