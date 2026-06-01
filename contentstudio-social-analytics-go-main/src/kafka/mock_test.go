package kafka

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMockProducer_Produce(t *testing.T) {
	mock := &MockProducer{}

	// Test with nil function
	err := mock.Produce(context.Background(), "test_topic", []byte("key"), []byte("value"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	var receivedTopic string
	var receivedKey, receivedValue []byte
	mock.ProduceFunc = func(ctx context.Context, topic string, key, value []byte) error {
		receivedTopic = topic
		receivedKey = key
		receivedValue = value
		return nil
	}
	err = mock.Produce(context.Background(), "my_topic", []byte("my_key"), []byte("my_value"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if receivedTopic != "my_topic" {
		t.Errorf("expected topic 'my_topic', got '%s'", receivedTopic)
	}
	if string(receivedKey) != "my_key" {
		t.Errorf("expected key 'my_key', got '%s'", string(receivedKey))
	}
	if string(receivedValue) != "my_value" {
		t.Errorf("expected value 'my_value', got '%s'", string(receivedValue))
	}

	// Test with error
	mock.ProduceFunc = func(ctx context.Context, topic string, key, value []byte) error {
		return errors.New("produce failed")
	}
	err = mock.Produce(context.Background(), "topic", []byte("key"), []byte("value"))
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockProducer_Close(t *testing.T) {
	mock := &MockProducer{}

	// Test with nil function
	err := mock.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	called := false
	mock.CloseFunc = func() error {
		called = true
		return nil
	}
	err = mock.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected CloseFunc to be called")
	}

	// Test with error
	mock.CloseFunc = func() error {
		return errors.New("close failed")
	}
	err = mock.Close()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockConsumer_Consume(t *testing.T) {
	mock := &MockConsumer{}

	handler := func(ctx context.Context, topic string, key, value []byte) error {
		return nil
	}

	// Test with nil function
	err := mock.Consume(context.Background(), []string{"topic"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	var receivedTopics []string
	mock.ConsumeFunc = func(ctx context.Context, topics []string, handler MessageHandler) error {
		receivedTopics = topics
		return nil
	}
	err = mock.Consume(context.Background(), []string{"topic1", "topic2"}, handler)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(receivedTopics) != 2 {
		t.Errorf("expected 2 topics, got %d", len(receivedTopics))
	}

	// Test with error
	mock.ConsumeFunc = func(ctx context.Context, topics []string, handler MessageHandler) error {
		return errors.New("consume failed")
	}
	err = mock.Consume(context.Background(), []string{"topic"}, handler)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockConsumer_Close(t *testing.T) {
	mock := &MockConsumer{}

	// Test with nil function
	err := mock.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	called := false
	mock.CloseFunc = func() error {
		called = true
		return nil
	}
	err = mock.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected CloseFunc to be called")
	}

	// Test with error
	mock.CloseFunc = func() error {
		return errors.New("close failed")
	}
	err = mock.Close()
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockConsumerWithMessages_Consume(t *testing.T) {
	messages := []MockMessage{
		{Topic: "topic1", Key: []byte("key1"), Value: []byte("value1")},
		{Topic: "topic2", Key: []byte("key2"), Value: []byte("value2")},
	}

	mock := &MockConsumerWithMessages{Messages: messages}

	var received []MockMessage
	handler := func(ctx context.Context, topic string, key, value []byte) error {
		received = append(received, MockMessage{Topic: topic, Key: key, Value: value})
		return nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := mock.Consume(ctx, []string{"topic1", "topic2"}, handler)
	if err != context.DeadlineExceeded {
		t.Fatalf("expected context.DeadlineExceeded, got %v", err)
	}

	if len(received) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(received))
	}
}

func TestMockConsumerWithMessages_Consume_ContextCancelled(t *testing.T) {
	messages := []MockMessage{
		{Topic: "topic1", Key: []byte("key1"), Value: []byte("value1")},
	}

	mock := &MockConsumerWithMessages{Messages: messages}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	handler := func(ctx context.Context, topic string, key, value []byte) error {
		return nil
	}

	err := mock.Consume(ctx, []string{"topic1"}, handler)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestMockConsumerWithMessages_Consume_HandlerError(t *testing.T) {
	messages := []MockMessage{
		{Topic: "topic1", Key: []byte("key1"), Value: []byte("value1")},
	}

	mock := &MockConsumerWithMessages{Messages: messages}

	expectedErr := errors.New("handler error")
	handler := func(ctx context.Context, topic string, key, value []byte) error {
		return expectedErr
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	err := mock.Consume(ctx, []string{"topic1"}, handler)
	if err != expectedErr {
		t.Fatalf("expected handler error, got %v", err)
	}
}

func TestMockConsumerWithMessages_Close(t *testing.T) {
	mock := &MockConsumerWithMessages{}

	if mock.IsClosed() {
		t.Fatal("expected IsClosed to be false initially")
	}

	err := mock.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !mock.IsClosed() {
		t.Fatal("expected IsClosed to be true after Close")
	}
}

func TestMockConsumerWithMessages_IsClosed(t *testing.T) {
	mock := &MockConsumerWithMessages{}

	if mock.IsClosed() {
		t.Fatal("expected false before close")
	}

	mock.closed = true
	if !mock.IsClosed() {
		t.Fatal("expected true after setting closed")
	}
}

func TestMockMessage_Struct(t *testing.T) {
	msg := MockMessage{
		Topic: "test_topic",
		Key:   []byte("test_key"),
		Value: []byte("test_value"),
	}

	if msg.Topic != "test_topic" {
		t.Errorf("expected topic 'test_topic', got '%s'", msg.Topic)
	}
	if string(msg.Key) != "test_key" {
		t.Errorf("expected key 'test_key', got '%s'", string(msg.Key))
	}
	if string(msg.Value) != "test_value" {
		t.Errorf("expected value 'test_value', got '%s'", string(msg.Value))
	}
}

func TestMockProducer_ImplementsInterface(t *testing.T) {
	var _ Producer = (*MockProducer)(nil)
}

func TestMockConsumer_ImplementsInterface(t *testing.T) {
	var _ Consumer = (*MockConsumer)(nil)
}

func TestMockConsumerWithMessages_ImplementsInterface(t *testing.T) {
	var _ Consumer = (*MockConsumerWithMessages)(nil)
}
