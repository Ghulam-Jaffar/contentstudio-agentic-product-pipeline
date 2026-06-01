package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== NewParserService Tests ==================

func TestNewParserService(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}
	svc := NewParserService(producer, log)

	if svc == nil {
		t.Fatal("NewParserService returned nil")
	}
	if svc.parsePost == nil {
		t.Error("parsePost is nil")
	}
	if svc.parseInsights == nil {
		t.Error("parseInsights is nil")
	}
	if svc.producer == nil {
		t.Error("producer is nil")
	}
}

func TestNewParserServiceWithFuncs(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	customParsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{PostID: "custom"}, nil
	}
	customParseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return []*kafkamodels.ParsedLinkedinInsights{{LinkedinID: "custom"}}, nil
	}

	svc := NewParserServiceWithFuncs(customParsePost, customParseInsights, producer, log)

	if svc == nil {
		t.Fatal("NewParserServiceWithFuncs returned nil")
	}

	// Test custom parsePost
	post, err := svc.parsePost(json.RawMessage(`{}`))
	if err != nil {
		t.Fatalf("parsePost error: %v", err)
	}
	if post.PostID != "custom" {
		t.Errorf("PostID = %q, want %q", post.PostID, "custom")
	}
}

// ================== ParseAndPublishPostTestable Tests ==================

func TestParserService_ParseAndPublishPostTestable_Success(t *testing.T) {
	log := logger.New("error")

	var produceCalls int32
	var lastTopic string
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			lastTopic = topic
			return nil
		},
	}

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{
			PostID:     "post123",
			LinkedinID: "li456",
		}, nil
	}

	svc := NewParserServiceWithFuncs(parsePost, nil, producer, log)

	parsed, err := svc.ParseAndPublishPostTestable(context.Background(), "li456", []byte(`{}`), "output-topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed == nil {
		t.Fatal("expected parsed post")
	}
	if parsed.PostID != "post123" {
		t.Errorf("PostID = %q, want %q", parsed.PostID, "post123")
	}
	if atomic.LoadInt32(&produceCalls) != 1 {
		t.Errorf("produceCalls = %d, want 1", produceCalls)
	}
	if lastTopic != "output-topic" {
		t.Errorf("topic = %q, want %q", lastTopic, "output-topic")
	}
}

func TestParserService_ParseAndPublishPostTestable_ParseError(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return nil, errors.New("parse error")
	}

	svc := NewParserServiceWithFuncs(parsePost, nil, producer, log)

	_, err := svc.ParseAndPublishPostTestable(context.Background(), "li456", []byte(`{}`), "output-topic")
	if err == nil {
		t.Error("expected error")
	}
}

func TestParserService_ParseAndPublishPostTestable_NilResult(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return nil, nil
	}

	svc := NewParserServiceWithFuncs(parsePost, nil, producer, log)

	parsed, err := svc.ParseAndPublishPostTestable(context.Background(), "li456", []byte(`{}`), "output-topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsed != nil {
		t.Error("expected nil result")
	}
}

func TestParserService_ParseAndPublishPostTestable_SetsLinkedinID(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{
			PostID:     "post123",
			LinkedinID: "", // Empty - should be filled in
		}, nil
	}

	svc := NewParserServiceWithFuncs(parsePost, nil, producer, log)

	parsed, err := svc.ParseAndPublishPostTestable(context.Background(), "li789", []byte(`{}`), "output-topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parsed.LinkedinID != "li789" {
		t.Errorf("LinkedinID = %q, want %q", parsed.LinkedinID, "li789")
	}
}

func TestParserService_ParseAndPublishPostTestable_ProduceError(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return errors.New("produce error")
		},
	}

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{PostID: "post123"}, nil
	}

	svc := NewParserServiceWithFuncs(parsePost, nil, producer, log)

	_, err := svc.ParseAndPublishPostTestable(context.Background(), "li456", []byte(`{}`), "output-topic")
	if err == nil {
		t.Error("expected error")
	}
}

// ================== ParseAndPublishInsightsTestable Tests ==================

func TestParserService_ParseAndPublishInsightsTestable_Success(t *testing.T) {
	log := logger.New("error")

	var produceCalls int32
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}

	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return []*kafkamodels.ParsedLinkedinInsights{
			{LinkedinID: "li456", CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
			{LinkedinID: "li456", CreatedAt: time.Date(2024, 1, 16, 0, 0, 0, 0, time.UTC)},
		}, nil
	}

	svc := NewParserServiceWithFuncs(nil, parseInsights, producer, log)

	parsedList, err := svc.ParseAndPublishInsightsTestable(context.Background(), "li456", []byte(`{}`), "output-topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(parsedList) != 2 {
		t.Errorf("len(parsedList) = %d, want 2", len(parsedList))
	}
	if atomic.LoadInt32(&produceCalls) != 2 {
		t.Errorf("produceCalls = %d, want 2", produceCalls)
	}
}

func TestParserService_ParseAndPublishInsightsTestable_ParseError(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return nil, errors.New("parse error")
	}

	svc := NewParserServiceWithFuncs(nil, parseInsights, producer, log)

	_, err := svc.ParseAndPublishInsightsTestable(context.Background(), "li456", []byte(`{}`), "output-topic")
	if err == nil {
		t.Error("expected error")
	}
}

func TestParserService_ParseAndPublishInsightsTestable_EmptyResult(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return nil, nil
	}

	svc := NewParserServiceWithFuncs(nil, parseInsights, producer, log)

	parsedList, err := svc.ParseAndPublishInsightsTestable(context.Background(), "li456", []byte(`{}`), "output-topic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if parsedList != nil {
		t.Error("expected nil result")
	}
}

func TestParserService_ParseAndPublishInsightsTestable_SetsRecordID(t *testing.T) {
	log := logger.New("error")

	var lastKey string
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			lastKey = string(key)
			return nil
		},
	}

	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return []*kafkamodels.ParsedLinkedinInsights{
			{LinkedinID: "li456", CreatedAt: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)},
		}, nil
	}

	svc := NewParserServiceWithFuncs(nil, parseInsights, producer, log)

	parsedList, _ := svc.ParseAndPublishInsightsTestable(context.Background(), "li456", []byte(`{}`), "output-topic")

	expectedRecordID := "li456_2024-01-15"
	if parsedList[0].RecordID != expectedRecordID {
		t.Errorf("RecordID = %q, want %q", parsedList[0].RecordID, expectedRecordID)
	}
	if lastKey != expectedRecordID {
		t.Errorf("lastKey = %q, want %q", lastKey, expectedRecordID)
	}
}

// ================== PostParserWorkerTestable Tests ==================

func TestParserService_PostParserWorkerTestable_ProcessesJobs(t *testing.T) {
	log := logger.New("error")

	var produceCalls int32
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{PostID: "post123"}, nil
	}

	svc := NewParserServiceWithFuncs(parsePost, nil, producer, log)

	in := make(chan ParseJob, 10)
	var counter uint64

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.PostParserWorkerTestable(ctx, in, &counter)
		close(done)
	}()

	// Send jobs
	in <- ParseJob{Key: []byte("li1"), Value: []byte(`{}`), OutputTopic: "output-topic"}
	in <- ParseJob{Key: []byte("li2"), Value: []byte(`{}`), OutputTopic: "output-topic"}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(in)
	<-done

	if atomic.LoadUint64(&counter) != 2 {
		t.Errorf("counter = %d, want 2", counter)
	}
	if atomic.LoadInt32(&produceCalls) != 2 {
		t.Errorf("produceCalls = %d, want 2", produceCalls)
	}
}

func TestParserService_PostParserWorkerTestable_ChannelClose(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}
	svc := NewParserService(producer, log)

	in := make(chan ParseJob, 10)

	done := make(chan struct{})
	go func() {
		svc.PostParserWorkerTestable(context.Background(), in, nil)
		close(done)
	}()

	close(in)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after channel close")
	}
}

func TestParserService_PostParserWorkerTestable_DefaultTopic(t *testing.T) {
	log := logger.New("error")

	var lastTopic string
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			lastTopic = topic
			return nil
		},
	}

	parsePost := func(data json.RawMessage) (*kafkamodels.ParsedLinkedinPost, error) {
		return &kafkamodels.ParsedLinkedinPost{PostID: "post123"}, nil
	}

	svc := NewParserServiceWithFuncs(parsePost, nil, producer, log)

	in := make(chan ParseJob, 10)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.PostParserWorkerTestable(ctx, in, nil)
		close(done)
	}()

	// Send job without OutputTopic
	in <- ParseJob{Key: []byte("li1"), Value: []byte(`{}`), OutputTopic: ""}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(in)
	<-done

	if lastTopic != topicParsedPagePosts {
		t.Errorf("topic = %q, want %q", lastTopic, topicParsedPagePosts)
	}
}

// ================== InsightsParserWorkerTestable Tests ==================

func TestParserService_InsightsParserWorkerTestable_ProcessesJobs(t *testing.T) {
	log := logger.New("error")

	var produceCalls int32
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}

	parseInsights := func(data json.RawMessage) ([]*kafkamodels.ParsedLinkedinInsights, error) {
		return []*kafkamodels.ParsedLinkedinInsights{
			{LinkedinID: "li456", CreatedAt: time.Now()},
		}, nil
	}

	svc := NewParserServiceWithFuncs(nil, parseInsights, producer, log)

	in := make(chan ParseJob, 10)
	var counter uint64

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		svc.InsightsParserWorkerTestable(ctx, in, &counter)
		close(done)
	}()

	// Send jobs
	in <- ParseJob{Key: []byte("li1"), Value: []byte(`{}`), OutputTopic: "output-topic"}
	in <- ParseJob{Key: []byte("li2"), Value: []byte(`{}`), OutputTopic: "output-topic"}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(in)
	<-done

	if atomic.LoadUint64(&counter) != 2 {
		t.Errorf("counter = %d, want 2", counter)
	}
}

func TestParserService_InsightsParserWorkerTestable_ChannelClose(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}
	svc := NewParserService(producer, log)

	in := make(chan ParseJob, 10)

	done := make(chan struct{})
	go func() {
		svc.InsightsParserWorkerTestable(context.Background(), in, nil)
		close(done)
	}()

	close(in)

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("worker did not exit after channel close")
	}
}

// ================== postParser Tests ==================

func TestPostParser_ContextCancelled(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go postParser(ctx, &wg, 1, in, out, log)

	cancel()
	wg.Wait()
}

func TestPostParser_ChannelClosed(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go postParser(ctx, &wg, 1, in, out, log)

	close(in)
	wg.Wait()
}

// ================== insightsParser Tests ==================

func TestInsightsParser_ContextCancelled(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 1, in, out, log)

	cancel()
	wg.Wait()
}

func TestInsightsParser_ChannelClosed(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 1, in, out, log)

	close(in)
	wg.Wait()
}

// ================== publisher Tests ==================

func TestPublisher_Success(t *testing.T) {
	log := logger.New("error")

	var produceCalls int32
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalls, 1)
			return nil
		},
	}

	in := make(chan PublishJob, 10)
	var counter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 1, "posts", in, producer, &counter, log)

	in <- PublishJob{Topic: "test-topic", Key: "key1", Data: []byte(`{}`)}
	in <- PublishJob{Topic: "test-topic", Key: "key2", Data: []byte(`{}`)}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	if atomic.LoadUint64(&counter) != 2 {
		t.Errorf("counter = %d, want 2", counter)
	}
	if atomic.LoadInt32(&produceCalls) != 2 {
		t.Errorf("produceCalls = %d, want 2", produceCalls)
	}
}

func TestPublisher_ProduceError(t *testing.T) {
	log := logger.New("error")

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return errors.New("produce error")
		},
	}

	in := make(chan PublishJob, 10)
	var counter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 1, "posts", in, producer, &counter, log)

	in <- PublishJob{Topic: "test-topic", Key: "key1", Data: []byte(`{}`)}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	// Counter should not be incremented on error
	if atomic.LoadUint64(&counter) != 0 {
		t.Errorf("counter = %d, want 0", counter)
	}
}

func TestPublisher_ContextCancelled(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	in := make(chan PublishJob, 10)
	var counter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 1, "posts", in, producer, &counter, log)

	cancel()
	close(in)
	wg.Wait()
}

func TestPublisher_ChannelClosed(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	in := make(chan PublishJob, 10)
	var counter uint64

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 1, "posts", in, producer, &counter, log)

	close(in)
	wg.Wait()
}
