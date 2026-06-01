package kafka

import (
	"context"
)

// MockProducer is a mock implementation of Producer for testing.
// It can be used by any service that needs to mock Kafka producer operations.
type MockProducer struct {
	ProduceFunc func(ctx context.Context, topic string, key, value []byte) error
	CloseFunc   func() error
}

func (m *MockProducer) Produce(ctx context.Context, topic string, key, value []byte) error {
	if m.ProduceFunc != nil {
		return m.ProduceFunc(ctx, topic, key, value)
	}
	return nil
}

func (m *MockProducer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Verify mock implements interface at compile time
var _ Producer = (*MockProducer)(nil)

// MockConsumer is a mock implementation of Consumer for testing.
// It can be used by any service that needs to mock Kafka consumer operations.
type MockConsumer struct {
	ConsumeFunc func(ctx context.Context, topics []string, handler MessageHandler) error
	CloseFunc   func() error
}

func (m *MockConsumer) Consume(ctx context.Context, topics []string, handler MessageHandler) error {
	if m.ConsumeFunc != nil {
		return m.ConsumeFunc(ctx, topics, handler)
	}
	return nil
}

func (m *MockConsumer) ConsumeWithAck(ctx context.Context, topics []string, handler AcknowledgingMessageHandler) error {
	if m.ConsumeFunc != nil {
		return m.ConsumeFunc(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			return handler(ctx, topic, key, value, func() {})
		})
	}
	return nil
}

func (m *MockConsumer) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}

// Verify mock implements interface at compile time
var _ Consumer = (*MockConsumer)(nil)

// MockMessage represents a mock Kafka message for testing
type MockMessage struct {
	Topic string
	Key   []byte
	Value []byte
}

// MockConsumerWithMessages is a mock consumer that delivers predefined messages
type MockConsumerWithMessages struct {
	Messages []MockMessage
	closed   bool
}

func (m *MockConsumerWithMessages) Consume(ctx context.Context, topics []string, handler MessageHandler) error {
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

func (m *MockConsumerWithMessages) ConsumeWithAck(ctx context.Context, topics []string, handler AcknowledgingMessageHandler) error {
	for _, msg := range m.Messages {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			ack := func() {}
			if err := handler(ctx, msg.Topic, msg.Key, msg.Value, ack); err != nil {
				return err
			}
		}
	}
	<-ctx.Done()
	return ctx.Err()
}

func (m *MockConsumerWithMessages) Close() error {
	m.closed = true
	return nil
}

// IsClosed returns whether the consumer has been closed
func (m *MockConsumerWithMessages) IsClosed() bool {
	return m.closed
}

// Verify mock implements interface at compile time
var _ Consumer = (*MockConsumerWithMessages)(nil)
