package main

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
)

// ================== Shared Mock Aliases ==================

// MockKafkaConsumer is an alias to the shared mock in kafka package.
type MockKafkaConsumer = kafka.MockConsumerWithMessages

// MockKafkaProducer is an alias to the shared mock in kafka package.
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
