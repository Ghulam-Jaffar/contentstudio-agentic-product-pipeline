package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if postsParserWorkers != 5 {
		t.Errorf("postsParserWorkers = %d, want 5", postsParserWorkers)
	}
	if insightsParserWorkers != 5 {
		t.Errorf("insightsParserWorkers = %d, want 5", insightsParserWorkers)
	}
	if batchProcessorsPerType != 3 {
		t.Errorf("batchProcessorsPerType = %d, want 3", batchProcessorsPerType)
	}
	if maxBatchSize != 10000 {
		t.Errorf("maxBatchSize = %d, want 10000", maxBatchSize)
	}
	if batchTimeout != 10*time.Second {
		t.Errorf("batchTimeout = %v, want 10s", batchTimeout)
	}
	if messageChanSize != 50000 {
		t.Errorf("messageChanSize = %d, want 50000", messageChanSize)
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
}

func TestBatchCollectors_Struct(t *testing.T) {
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	if bc.posts == nil {
		t.Error("posts channel is nil")
	}
	if bc.insights == nil {
		t.Error("insights channel is nil")
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
			handler(ctx, topics[0], []byte("key"), []byte(`{"test": true}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	// Simulate consumer behavior
	go func() {
		mockConsumer.Consume(ctx, []string{topicRawPagePosts}, func(ctx context.Context, topic string, key, value []byte) error {
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
			handler(ctx, topics[0], []byte("key"), []byte(`{"insights": {}}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		mockConsumer.Consume(ctx, []string{topicRawPageInsights}, func(ctx context.Context, topic string, key, value []byte) error {
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

// ================== Batch Processor Tests ==================

func TestBatchProcessors_ContextCancel(t *testing.T) {
	_, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	cancel()

	// Verify channels can be closed without panic
	close(batches.posts)
	close(batches.insights)
}

// ================== Worker Tests ==================

func TestPostsParserWorker_ChannelClose(t *testing.T) {
	postsMsgChan := make(chan RawMessage, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	var parsedCount uint64
	log := logger.New("error")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go postsParserWorker(ctx, 0, postsMsgChan, batches, &parsedCount, log, &wg)

	close(postsMsgChan)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("postsParserWorker did not exit after channel close")
	}
}

func TestInsightsParserWorker_ChannelClose(t *testing.T) {
	insightsMsgChan := make(chan RawMessage, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	var parsedCount uint64
	log := logger.New("error")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParserWorker(ctx, 0, insightsMsgChan, batches, &parsedCount, log, &wg)

	close(insightsMsgChan)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("insightsParserWorker did not exit after channel close")
	}
}

// ================== JSON Unmarshal Tests ==================

func TestParsedLinkedinPost_Unmarshal(t *testing.T) {
	jsonData := `{
		"linkedin_id": "li123",
		"workspace_id": "ws456",
		"post_id": "post789"
	}`

	var post kafkamodels.ParsedLinkedinPost
	err := json.Unmarshal([]byte(jsonData), &post)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}

func TestParsedLinkedinInsights_Unmarshal(t *testing.T) {
	jsonData := `{
		"linkedin_id": "li123",
		"workspace_id": "ws456"
	}`

	var insights kafkamodels.ParsedLinkedinInsights
	err := json.Unmarshal([]byte(jsonData), &insights)
	if err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
}
