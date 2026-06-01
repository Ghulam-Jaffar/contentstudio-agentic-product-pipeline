package main

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
)

// ================== Shared Mock Aliases ==================

// MockGMBClient is an alias to the shared mock in clients/social package.
type MockGMBClient = social.MockGMBClient

// Verify mock implements interface at compile time
var _ GMBAPI = (*MockGMBClient)(nil)

// MockUnifiedSocialRepository is an alias to the shared mock in db/mongodb package.
type MockUnifiedSocialRepository = mongodb.MockUnifiedSocialRepository

// MockKafkaProducer is an alias to the shared mock in kafka package.
type MockKafkaProducer = kafka.MockProducer

// MockKafkaConsumer is an alias to the shared mock in kafka package.
type MockKafkaConsumer = kafka.MockConsumer

// ================== Helper Functions ==================

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

// NewMockGMBClient creates a new mock GMB client for testing
func NewMockGMBClient() *MockGMBClient {
	return &MockGMBClient{}
}

// NewMockKafkaConsumer creates a new mock Kafka consumer for testing
func NewMockKafkaConsumer() *MockKafkaConsumer {
	return &MockKafkaConsumer{}
}
