package main

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
)

// ================== Shared Mock Aliases ==================
// These aliases allow tests to use the shared mocks from their respective packages.
// Actual implementations are in:
//   - clients/social/mock.go
//   - db/mongodb/mock.go
//   - kafka/mock.go

// MockPinterestClient is an alias to the shared mock in clients/social package.
// This allows services to use the common mock for testing Pinterest API operations.
type MockPinterestClient = social.MockPinterestClient

// Verify mock implements interface at compile time
var _ social.PinterestAPI = (*MockPinterestClient)(nil)

// MockUnifiedSocialRepository is an alias to the shared mock in db/mongodb package.
// This allows services to use the common mock for testing MongoDB operations.
type MockUnifiedSocialRepository = mongodb.MockUnifiedSocialRepository

// MockKafkaConsumer is an alias to the shared mock in kafka package.
// This allows services to use the common mock for testing Kafka consumer operations.
type MockKafkaConsumer = kafka.MockConsumerWithMessages

// MockKafkaProducer is an alias to the shared mock in kafka package.
// This allows services to use the common mock for testing Kafka producer operations.
type MockKafkaProducer = kafka.MockProducer

// MockMessage represents a mock Kafka message for testing
type MockMessage = kafka.MockMessage

// ================== Helper Functions ==================

// NewMockKafkaConsumerWithMessages creates a new mock consumer with pre-loaded messages
func NewMockKafkaConsumerWithMessages(messages []MockMessage) *MockKafkaConsumer {
	return &MockKafkaConsumer{
		Messages: messages,
	}
}

// NewMockKafkaProducer creates a new mock Kafka producer for testing
func NewMockKafkaProducer() *MockKafkaProducer {
	return &MockKafkaProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}
}

// NewMockUnifiedSocialRepository creates a new mock MongoDB repository for testing
func NewMockUnifiedSocialRepository() *MockUnifiedSocialRepository {
	return &MockUnifiedSocialRepository{}
}

// NewMockPinterestClient creates a new mock Pinterest client for testing
func NewMockPinterestClient() *MockPinterestClient {
	return &MockPinterestClient{}
}
