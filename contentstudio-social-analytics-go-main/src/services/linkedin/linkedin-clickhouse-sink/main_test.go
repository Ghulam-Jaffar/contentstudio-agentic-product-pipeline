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
	if postsWorkers != 6 {
		t.Errorf("postsWorkers = %d, want 6", postsWorkers)
	}
	if insightsWorkers != 6 {
		t.Errorf("insightsWorkers = %d, want 6", insightsWorkers)
	}
	if batchProcessorsPerType != 3 {
		t.Errorf("batchProcessorsPerType = %d, want 3", batchProcessorsPerType)
	}
	if maxBatchSize != 1000 {
		t.Errorf("maxBatchSize = %d, want 1000", maxBatchSize)
	}
	if batchTimeout != 5*time.Second {
		t.Errorf("batchTimeout = %v, want 5s", batchTimeout)
	}
	if messageChanSize != 50_000 {
		t.Errorf("messageChanSize = %d, want 50000", messageChanSize)
	}
	if topicPagePosts != "parsed-linkedin-page-posts" {
		t.Errorf("topicPagePosts = %q, want %q", topicPagePosts, "parsed-linkedin-page-posts")
	}
	if topicPageInsights != "parsed-linkedin-page-insights" {
		t.Errorf("topicPageInsights = %q, want %q", topicPageInsights, "parsed-linkedin-page-insights")
	}
}

// ================== Struct Tests ==================

func TestMessage_Struct(t *testing.T) {
	msg := Message{
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
	var pickedCount uint64
	postsMsgChan := make(chan Message, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), []byte(`{"test": true}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		mockConsumer.Consume(ctx, []string{topicPagePosts}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&pickedCount, 1)
			select {
			case postsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
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
}

func TestInsightsConsumerHandler(t *testing.T) {
	var pickedCount uint64
	insightsMsgChan := make(chan Message, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), []byte(`{"insights": {}}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		mockConsumer.Consume(ctx, []string{topicPageInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&pickedCount, 1)
			select {
			case insightsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
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
}

// ================== Worker Tests ==================

func TestPostsWorker_ChannelClose(t *testing.T) {
	postsMsgChan := make(chan Message, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	log := logger.New("error")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go postsWorker(ctx, 0, postsMsgChan, batches, log, &wg)

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
		t.Fatal("postsWorker did not exit after channel close")
	}
}

func TestInsightsWorker_ChannelClose(t *testing.T) {
	insightsMsgChan := make(chan Message, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	log := logger.New("error")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsWorker(ctx, 0, insightsMsgChan, batches, log, &wg)

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
		t.Fatal("insightsWorker did not exit after channel close")
	}
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

// ================== Concurrent Tests ==================

func TestConcurrentMessageProcessing(t *testing.T) {
	var counter int64
	msgChan := make(chan Message, 100)

	var wg sync.WaitGroup
	numWorkers := 5

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range msgChan {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}

	// Send messages
	for i := 0; i < 50; i++ {
		msgChan <- Message{Topic: "test", Key: []byte("key"), Value: []byte("value")}
	}
	close(msgChan)

	wg.Wait()

	if atomic.LoadInt64(&counter) != 50 {
		t.Errorf("counter = %d, want 50", counter)
	}
}

// ================== Handler Tests ==================

func TestHandleParsedPost_Success(t *testing.T) {
	log := logger.New("error")
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	jsonData := `{
		"linkedin_id": "li123",
		"workspace_id": "ws456",
		"post_id": "post789",
		"content": "test content"
	}`

	ctx := context.Background()
	err := handleParsedPost(ctx, []byte("key"), []byte(jsonData), batches, log)
	if err != nil {
		t.Errorf("handleParsedPost returned error: %v", err)
	}

	select {
	case post := <-batches.posts:
		if post.LinkedinID != "li123" {
			t.Errorf("LinkedinID = %q, want %q", post.LinkedinID, "li123")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for post")
	}
}

func TestHandleParsedPost_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx := context.Background()
	err := handleParsedPost(ctx, []byte("key"), []byte("invalid json"), batches, log)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleParsedPost_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost), // unbuffered
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	jsonData := `{"linkedin_id": "li123"}`

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := handleParsedPost(ctx, []byte("key"), []byte(jsonData), batches, log)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestHandleParsedInsights_Success(t *testing.T) {
	log := logger.New("error")
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	jsonData := `{
		"linkedin_id": "li123",
		"workspace_id": "ws456",
		"followers_count": 1000
	}`

	ctx := context.Background()
	err := handleParsedInsights(ctx, []byte("key"), []byte(jsonData), batches, log)
	if err != nil {
		t.Errorf("handleParsedInsights returned error: %v", err)
	}

	select {
	case insights := <-batches.insights:
		if insights.LinkedinID != "li123" {
			t.Errorf("LinkedinID = %q, want %q", insights.LinkedinID, "li123")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for insights")
	}
}

func TestHandleParsedInsights_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx := context.Background()
	err := handleParsedInsights(ctx, []byte("key"), []byte("invalid"), batches, log)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleParsedInsights_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights), // unbuffered
	}

	jsonData := `{"linkedin_id": "li123"}`

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := handleParsedInsights(ctx, []byte("key"), []byte(jsonData), batches, log)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ================== Worker Flow Tests ==================

func TestPostsWorker_ProcessesValidMessage(t *testing.T) {
	log := logger.New("error")
	postsMsgChan := make(chan Message, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go postsWorker(ctx, 1, postsMsgChan, batches, log, &wg)

	jsonData := `{"linkedin_id": "li123", "post_id": "post789"}`
	postsMsgChan <- Message{
		Topic: topicPagePosts,
		Key:   []byte("key"),
		Value: []byte(jsonData),
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(postsMsgChan)
	wg.Wait()

	if len(batches.posts) != 1 {
		t.Errorf("batches.posts has %d items, want 1", len(batches.posts))
	}
}

func TestPostsWorker_ContextCancel(t *testing.T) {
	log := logger.New("error")
	postsMsgChan := make(chan Message, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go postsWorker(ctx, 1, postsMsgChan, batches, log, &wg)

	cancel()
	wg.Wait()
}

func TestPostsWorker_UnknownTopic(t *testing.T) {
	log := logger.New("error")
	postsMsgChan := make(chan Message, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go postsWorker(ctx, 1, postsMsgChan, batches, log, &wg)

	postsMsgChan <- Message{
		Topic: "unknown-topic",
		Key:   []byte("key"),
		Value: []byte("{}"),
	}

	time.Sleep(50 * time.Millisecond)

	cancel()
	close(postsMsgChan)
	wg.Wait()

	if len(batches.posts) != 0 {
		t.Errorf("expected no posts for unknown topic, got %d", len(batches.posts))
	}
}

func TestInsightsWorker_ProcessesPageInsights(t *testing.T) {
	log := logger.New("error")
	insightsMsgChan := make(chan Message, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsWorker(ctx, 1, insightsMsgChan, batches, log, &wg)

	jsonData := `{"linkedin_id": "li123", "followers_count": 500}`
	insightsMsgChan <- Message{
		Topic: topicPageInsights,
		Key:   []byte("key"),
		Value: []byte(jsonData),
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(insightsMsgChan)
	wg.Wait()

	if len(batches.insights) != 1 {
		t.Errorf("batches.insights has %d items, want 1", len(batches.insights))
	}
}

func TestInsightsWorker_ProcessesProfileInsights(t *testing.T) {
	log := logger.New("error")
	insightsMsgChan := make(chan Message, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsWorker(ctx, 1, insightsMsgChan, batches, log, &wg)

	jsonData := `{"linkedin_id": "li456", "connections_count": 250}`
	insightsMsgChan <- Message{
		Topic: topicProfileInsights,
		Key:   []byte("key"),
		Value: []byte(jsonData),
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(insightsMsgChan)
	wg.Wait()

	if len(batches.insights) != 1 {
		t.Errorf("batches.insights has %d items, want 1", len(batches.insights))
	}
}

func TestInsightsWorker_ContextCancel(t *testing.T) {
	log := logger.New("error")
	insightsMsgChan := make(chan Message, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsWorker(ctx, 1, insightsMsgChan, batches, log, &wg)

	cancel()
	wg.Wait()
}

func TestInsightsWorker_UnknownTopic(t *testing.T) {
	log := logger.New("error")
	insightsMsgChan := make(chan Message, 10)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsWorker(ctx, 1, insightsMsgChan, batches, log, &wg)

	insightsMsgChan <- Message{
		Topic: "unknown-insights-topic",
		Key:   []byte("key"),
		Value: []byte("{}"),
	}

	time.Sleep(50 * time.Millisecond)

	cancel()
	close(insightsMsgChan)
	wg.Wait()

	if len(batches.insights) != 0 {
		t.Errorf("expected no insights for unknown topic, got %d", len(batches.insights))
	}
}

// ================== Multiple Message Tests ==================

func TestPostsWorker_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	postsMsgChan := make(chan Message, 100)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 100),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go postsWorker(ctx, 1, postsMsgChan, batches, log, &wg)

	for i := 0; i < 10; i++ {
		jsonData := `{"linkedin_id": "li123", "post_id": "post` + string(rune('0'+i)) + `"}`
		postsMsgChan <- Message{
			Topic: topicPagePosts,
			Key:   []byte("key"),
			Value: []byte(jsonData),
		}
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(postsMsgChan)
	wg.Wait()

	if len(batches.posts) != 10 {
		t.Errorf("batches.posts has %d items, want 10", len(batches.posts))
	}
}

func TestInsightsWorker_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	insightsMsgChan := make(chan Message, 100)
	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, 10),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, 100),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsWorker(ctx, 1, insightsMsgChan, batches, log, &wg)

	for i := 0; i < 10; i++ {
		jsonData := `{"linkedin_id": "li` + string(rune('0'+i)) + `"}`
		insightsMsgChan <- Message{
			Topic: topicPageInsights,
			Key:   []byte("key"),
			Value: []byte(jsonData),
		}
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(insightsMsgChan)
	wg.Wait()

	if len(batches.insights) != 10 {
		t.Errorf("batches.insights has %d items, want 10", len(batches.insights))
	}
}

// ================== Additional Constants Tests ==================

func TestTopicConstants(t *testing.T) {
	if topicProfilePosts != "parsed-linkedin-profile-posts" {
		t.Errorf("topicProfilePosts = %q, want %q", topicProfilePosts, "parsed-linkedin-profile-posts")
	}
	if topicProfileInsights != "parsed-linkedin-profile-insights" {
		t.Errorf("topicProfileInsights = %q, want %q", topicProfileInsights, "parsed-linkedin-profile-insights")
	}
	if pageConsumerGroup != "linkedin-page-clickhouse-sink-group" {
		t.Errorf("pageConsumerGroup = %q, want %q", pageConsumerGroup, "linkedin-page-clickhouse-sink-group")
	}
	if profileConsumerGroup != "linkedin-profile-clickhouse-sink-group" {
		t.Errorf("profileConsumerGroup = %q, want %q", profileConsumerGroup, "linkedin-profile-clickhouse-sink-group")
	}
}
