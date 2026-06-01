package main

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if postParserWorkers != 6 {
		t.Errorf("postParserWorkers = %d, want 6", postParserWorkers)
	}
	if insightsParserWorkers != 6 {
		t.Errorf("insightsParserWorkers = %d, want 6", insightsParserWorkers)
	}
	if postPublisherWorkers != 6 {
		t.Errorf("postPublisherWorkers = %d, want 6", postPublisherWorkers)
	}
	if insightsPublisherWorkers != 6 {
		t.Errorf("insightsPublisherWorkers = %d, want 6", insightsPublisherWorkers)
	}
	if parseChanSize != 500 {
		t.Errorf("parseChanSize = %d, want 500", parseChanSize)
	}
	if publishChanSize != 1000 {
		t.Errorf("publishChanSize = %d, want 1000", publishChanSize)
	}
	if metricsEvery != 10*time.Second {
		t.Errorf("metricsEvery = %v, want 10s", metricsEvery)
	}
	if topicRawPagePosts != "raw-linkedin-page-posts" {
		t.Errorf("topicRawPagePosts = %q, want %q", topicRawPagePosts, "raw-linkedin-page-posts")
	}
	if topicParsedPagePosts != "parsed-linkedin-page-posts" {
		t.Errorf("topicParsedPagePosts = %q, want %q", topicParsedPagePosts, "parsed-linkedin-page-posts")
	}
}

// ================== Struct Tests ==================

func TestParseJob_Struct(t *testing.T) {
	job := ParseJob{
		JobType:     "posts",
		Key:         []byte("key"),
		Value:       []byte("value"),
		OutputTopic: "output-topic",
	}

	if job.JobType != "posts" {
		t.Errorf("JobType = %q, want %q", job.JobType, "posts")
	}
	if job.OutputTopic != "output-topic" {
		t.Errorf("OutputTopic = %q, want %q", job.OutputTopic, "output-topic")
	}
}

func TestPublishJob_Struct(t *testing.T) {
	job := PublishJob{
		Topic: "test-topic",
		Key:   "test-key",
		Data:  []byte("data"),
	}

	if job.Topic != "test-topic" {
		t.Errorf("Topic = %q, want %q", job.Topic, "test-topic")
	}
	if job.Key != "test-key" {
		t.Errorf("Key = %q, want %q", job.Key, "test-key")
	}
}

// ================== Consumer Handler Tests ==================

func TestPostsConsumerHandler(t *testing.T) {
	var pickedCount uint64
	postsParseJobs := make(chan ParseJob, 10)

	mockConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topics[0], []byte("key"), []byte(`{"test": true}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		mockConsumer.Consume(ctx, []string{topicRawPagePosts}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&pickedCount, 1)
			select {
			case postsParseJobs <- ParseJob{JobType: "posts", Key: key, Value: value, OutputTopic: topicParsedPagePosts}:
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
	insightsParseJobs := make(chan ParseJob, 10)

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
			case insightsParseJobs <- ParseJob{JobType: "insights", Key: key, Value: value, OutputTopic: topicParsedPageInsights}:
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

// ================== Parser Worker Tests ==================

func TestPostParser_ChannelClose(t *testing.T) {
	postsParseJobs := make(chan ParseJob, 10)
	postsPublishJobs := make(chan PublishJob, 10)
	log := logger.New("error")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go postParser(ctx, &wg, 0, postsParseJobs, postsPublishJobs, log)

	close(postsParseJobs)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("postParser did not exit after channel close")
	}
}

func TestInsightsParser_ChannelClose(t *testing.T) {
	insightsParseJobs := make(chan ParseJob, 10)
	insightsPublishJobs := make(chan PublishJob, 10)
	log := logger.New("error")

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, insightsParseJobs, insightsPublishJobs, log)

	close(insightsParseJobs)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("insightsParser did not exit after channel close")
	}
}

// ================== Publisher Worker Tests ==================

func TestPublisher_ChannelClose(t *testing.T) {
	publishJobs := make(chan PublishJob, 10)
	var pubCount uint64
	log := logger.New("error")

	producer := &kafka.MockProducer{}

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 0, "posts", publishJobs, producer, &pubCount, log)

	close(publishJobs)

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Expected
	case <-time.After(2 * time.Second):
		t.Fatal("publisher did not exit after channel close")
	}
}

func TestPublisher_ProcessesJobs(t *testing.T) {
	publishJobs := make(chan PublishJob, 10)
	var pubCount uint64
	log := logger.New("error")

	var produceCalled int32
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&produceCalled, 1)
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 0, "posts", publishJobs, producer, &pubCount, log)

	// Send job
	publishJobs <- PublishJob{Topic: "test-topic", Key: "key", Data: []byte("data")}

	time.Sleep(100 * time.Millisecond)

	if atomic.LoadInt32(&produceCalled) != 1 {
		t.Errorf("Produce called %d times, want 1", produceCalled)
	}

	cancel()
	close(publishJobs)
	wg.Wait()
}

// ================== Concurrent Tests ==================

func TestConcurrentParseJobProcessing(t *testing.T) {
	var counter int64
	jobChan := make(chan ParseJob, 100)

	var wg sync.WaitGroup
	numWorkers := 5

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for range jobChan {
				atomic.AddInt64(&counter, 1)
			}
		}()
	}

	// Send jobs
	for i := 0; i < 50; i++ {
		jobChan <- ParseJob{JobType: "posts", Key: []byte("key"), Value: []byte("value")}
	}
	close(jobChan)

	wg.Wait()

	if atomic.LoadInt64(&counter) != 50 {
		t.Errorf("counter = %d, want 50", counter)
	}
}

// ================== parseAndQueuePost Tests ==================

func TestParseAndQueuePost_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	job := ParseJob{
		Key:         []byte("li123"),
		Value:       []byte("not valid json"),
		OutputTopic: topicParsedPagePosts,
	}

	parseAndQueuePost(context.Background(), job, out, log)

	select {
	case <-out:
		t.Error("expected no output for invalid JSON")
	default:
		// Expected - no output
	}
}

func TestParseAndQueuePost_NilParsedResult(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	// Empty JSON object should result in nil or empty parsed result
	job := ParseJob{
		Key:         []byte("li123"),
		Value:       []byte("{}"),
		OutputTopic: topicParsedPagePosts,
	}

	parseAndQueuePost(context.Background(), job, out, log)

	// This may or may not produce output depending on parser behavior
	// The test just ensures no panic occurs
}

func TestParseAndQueuePost_EmptyOutputTopic(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	// Valid post data with empty output topic
	job := ParseJob{
		Key:         []byte("li123"),
		Value:       []byte(`{"id": "urn:li:ugcPost:123", "created": {"time": 1700000000000}}`),
		OutputTopic: "", // Empty - should use fallback
	}

	parseAndQueuePost(context.Background(), job, out, log)

	// Test verifies no panic when output topic is empty
}

func TestParseAndQueuePost_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob) // unbuffered to block

	// Valid post data
	job := ParseJob{
		Key:         []byte("li123"),
		Value:       []byte(`{"id": "urn:li:ugcPost:123", "created": {"time": 1700000000000}}`),
		OutputTopic: topicParsedPagePosts,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parseAndQueuePost(ctx, job, out, log)

	// Should exit due to canceled context
}

// ================== parseAndQueueInsights Tests ==================

func TestParseAndQueueInsights_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	job := ParseJob{
		Key:         []byte("li123"),
		Value:       []byte("not valid json"),
		OutputTopic: topicParsedPageInsights,
	}

	parseAndQueueInsights(context.Background(), job, out, log)

	select {
	case <-out:
		t.Error("expected no output for invalid JSON")
	default:
		// Expected - no output
	}
}

func TestParseAndQueueInsights_EmptyParsedList(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	// Empty JSON object should result in empty parsed list
	job := ParseJob{
		Key:         []byte("li123"),
		Value:       []byte("{}"),
		OutputTopic: topicParsedPageInsights,
	}

	parseAndQueueInsights(context.Background(), job, out, log)

	// Should not produce output for empty data
}

func TestParseAndQueueInsights_EmptyOutputTopic(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	// Insights data with empty output topic
	job := ParseJob{
		Key:         []byte("li123"),
		Value:       []byte(`{"followerData": {"elements": []}}`),
		OutputTopic: "", // Empty - should use fallback
	}

	parseAndQueueInsights(context.Background(), job, out, log)

	// Test verifies no panic when output topic is empty
}

func TestParseAndQueueInsights_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob) // unbuffered to block

	// Insights data that will produce output
	job := ParseJob{
		Key:         []byte("li123"),
		Value:       []byte(`{"followerData": {"elements": []}, "pageStatistics": {"elements": []}}`),
		OutputTopic: topicParsedPageInsights,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parseAndQueueInsights(ctx, job, out, log)

	// Should exit due to canceled context
}

// ================== postParser Flow Tests ==================

func TestPostParser_ProcessesValidJob(t *testing.T) {
	postsParseJobs := make(chan ParseJob, 10)
	postsPublishJobs := make(chan PublishJob, 10)
	log := logger.New("error")

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go postParser(ctx, &wg, 0, postsParseJobs, postsPublishJobs, log)

	// Send a valid post job
	postsParseJobs <- ParseJob{
		Key:         []byte("li123"),
		Value:       []byte(`{"id": "urn:li:ugcPost:123"}`),
		OutputTopic: topicParsedPagePosts,
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(postsParseJobs)
	wg.Wait()
}

func TestPostParser_ContextCancel(t *testing.T) {
	postsParseJobs := make(chan ParseJob, 10)
	postsPublishJobs := make(chan PublishJob, 10)
	log := logger.New("error")

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go postParser(ctx, &wg, 0, postsParseJobs, postsPublishJobs, log)

	cancel()
	wg.Wait()
}

// ================== insightsParser Flow Tests ==================

func TestInsightsParser_ProcessesValidJob(t *testing.T) {
	insightsParseJobs := make(chan ParseJob, 10)
	insightsPublishJobs := make(chan PublishJob, 10)
	log := logger.New("error")

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, insightsParseJobs, insightsPublishJobs, log)

	// Send a valid insights job
	insightsParseJobs <- ParseJob{
		Key:         []byte("li123"),
		Value:       []byte(`{"followerData": {}}`),
		OutputTopic: topicParsedPageInsights,
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(insightsParseJobs)
	wg.Wait()
}

func TestInsightsParser_ContextCancel(t *testing.T) {
	insightsParseJobs := make(chan ParseJob, 10)
	insightsPublishJobs := make(chan PublishJob, 10)
	log := logger.New("error")

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, insightsParseJobs, insightsPublishJobs, log)

	cancel()
	wg.Wait()
}

// ================== Publisher Error Handling Tests ==================

func TestPublisher_ProduceErrorWithContinue(t *testing.T) {
	publishJobs := make(chan PublishJob, 10)
	var pubCount uint64
	log := logger.New("error")

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return context.DeadlineExceeded // Simulate error
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 0, "posts", publishJobs, producer, &pubCount, log)

	// Send job that will fail
	publishJobs <- PublishJob{Topic: "test-topic", Key: "key", Data: []byte("data")}

	time.Sleep(100 * time.Millisecond)

	// Counter should not be incremented on error
	if atomic.LoadUint64(&pubCount) != 0 {
		t.Errorf("pubCount = %d, want 0 on error", pubCount)
	}

	cancel()
	close(publishJobs)
	wg.Wait()
}

func TestPublisher_ContextCancelFlow(t *testing.T) {
	publishJobs := make(chan PublishJob, 10)
	var pubCount uint64
	log := logger.New("error")

	producer := &kafka.MockProducer{}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 0, "posts", publishJobs, producer, &pubCount, log)

	cancel()
	wg.Wait()
}

// ================== Additional Topic Constants Tests ==================

func TestProfileTopicConstants(t *testing.T) {
	if topicRawProfilePosts != "raw-linkedin-profile-posts" {
		t.Errorf("topicRawProfilePosts = %q, want %q", topicRawProfilePosts, "raw-linkedin-profile-posts")
	}
	if topicRawProfileInsights != "raw-linkedin-profile-insights" {
		t.Errorf("topicRawProfileInsights = %q, want %q", topicRawProfileInsights, "raw-linkedin-profile-insights")
	}
	if topicParsedProfilePosts != "parsed-linkedin-profile-posts" {
		t.Errorf("topicParsedProfilePosts = %q, want %q", topicParsedProfilePosts, "parsed-linkedin-profile-posts")
	}
	if topicParsedProfileInsights != "parsed-linkedin-profile-insights" {
		t.Errorf("topicParsedProfileInsights = %q, want %q", topicParsedProfileInsights, "parsed-linkedin-profile-insights")
	}
	if pageParserConsumerGroup != "linkedin-page-parser-group" {
		t.Errorf("pageParserConsumerGroup = %q, want %q", pageParserConsumerGroup, "linkedin-page-parser-group")
	}
	if profileParserConsumerGroup != "linkedin-profile-parser-group" {
		t.Errorf("profileParserConsumerGroup = %q, want %q", profileParserConsumerGroup, "linkedin-profile-parser-group")
	}
}
