package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if maxWorkers != 5 {
		t.Errorf("maxWorkers = %d, want 5", maxWorkers)
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
	if batchProcessorsPerType != 3 {
		t.Errorf("batchProcessorsPerType = %d, want 3", batchProcessorsPerType)
	}
	if consumerGroup != "instagram-clickhouse-sink-group" {
		t.Errorf("consumerGroup = %q, want %q", consumerGroup, "instagram-clickhouse-sink-group")
	}
}

// ================== BatchCollectors Tests ==================

func TestBatchCollectors_Struct(t *testing.T) {
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 10),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 10),
	}

	if bc.posts == nil {
		t.Error("posts channel is nil")
	}
	if bc.insights == nil {
		t.Error("insights channel is nil")
	}
}

// ================== handleParsedPost Tests ==================

func TestHandleParsedPost_Success(t *testing.T) {
	log := logger.New("error")
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 10),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 10),
	}

	post := kafkamodels.ParsedInstagramPost{
		MediaID:     "media123",
		InstagramID: "ig456",
	}
	postJSON, _ := json.Marshal(post)

	ctx := context.Background()
	err := handleParsedPost(ctx, []byte("key"), postJSON, bc, log)
	if err != nil {
		t.Errorf("handleParsedPost returned error: %v", err)
	}

	select {
	case received := <-bc.posts:
		if received.MediaID != "media123" {
			t.Errorf("MediaID = %q, want %q", received.MediaID, "media123")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for post")
	}
}

func TestHandleParsedPost_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 10),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 10),
	}

	ctx := context.Background()
	err := handleParsedPost(ctx, []byte("key"), []byte("invalid json"), bc, log)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleParsedPost_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost), // unbuffered - will block
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 10),
	}

	post := kafkamodels.ParsedInstagramPost{MediaID: "media123"}
	postJSON, _ := json.Marshal(post)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	err := handleParsedPost(ctx, []byte("key"), postJSON, bc, log)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ================== handleParsedInsight Tests ==================

func TestHandleParsedInsight_Success(t *testing.T) {
	log := logger.New("error")
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 10),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 10),
	}

	insight := kafkamodels.ParsedInstagramInsight{
		InstagramID: "ig123",
		RecordID:    "record456",
	}
	insightJSON, _ := json.Marshal(insight)

	ctx := context.Background()
	err := handleParsedInsight(ctx, []byte("key"), insightJSON, bc, log)
	if err != nil {
		t.Errorf("handleParsedInsight returned error: %v", err)
	}

	select {
	case received := <-bc.insights:
		if received.InstagramID != "ig123" {
			t.Errorf("InstagramID = %q, want %q", received.InstagramID, "ig123")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for insight")
	}
}

func TestHandleParsedInsight_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 10),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 10),
	}

	ctx := context.Background()
	err := handleParsedInsight(ctx, []byte("key"), []byte("invalid"), bc, log)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleParsedInsight_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 10),
		insights: make(chan *kafkamodels.ParsedInstagramInsight), // unbuffered
	}

	insight := kafkamodels.ParsedInstagramInsight{InstagramID: "ig123"}
	insightJSON, _ := json.Marshal(insight)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := handleParsedInsight(ctx, []byte("key"), insightJSON, bc, log)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ================== startBatchProcessors Tests ==================

func TestStartBatchProcessors(t *testing.T) {
	log := logger.New("error")
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 10),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 10),
	}

	var insertedPosts, insertedInsights uint64
	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	// Start with 1 processor per type for testing
	startBatchProcessors(ctx, bc, nil, log, &wg, 1, &insertedPosts, &insertedInsights)

	// Give processors time to start
	time.Sleep(50 * time.Millisecond)

	// Cancel to stop processors
	cancel()
	close(bc.posts)
	close(bc.insights)

	wg.Wait()
}

// ================== processPostsBatch Tests ==================

func TestProcessPostsBatch_ContextCancel_EmptyBatch(t *testing.T) {
	log := logger.New("error")
	postsChan := make(chan *kafkamodels.ParsedInstagramPost, 10)

	var insertedCounter uint64

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processPostsBatch(ctx, 0, postsChan, nil, log, &insertedCounter)
		close(done)
	}()

	// Cancel context — processor should NOT exit yet, it drains the channel
	cancel()
	// Close channel — processor exits via channel close
	close(postsChan)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("processPostsBatch did not exit after channel close")
	}
}

func TestProcessPostsBatch_ChannelClose_EmptyBatch(t *testing.T) {
	log := logger.New("error")
	postsChan := make(chan *kafkamodels.ParsedInstagramPost, 10)

	var insertedCounter uint64

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPostsBatch(ctx, 0, postsChan, nil, log, &insertedCounter)
		close(done)
	}()

	close(postsChan)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("processPostsBatch did not exit after channel close")
	}
}

// ================== processInsightsBatch Tests ==================

func TestProcessInsightsBatch_ContextCancel_EmptyBatch(t *testing.T) {
	log := logger.New("error")
	insightsChan := make(chan *kafkamodels.ParsedInstagramInsight, 10)

	var insertedCounter uint64

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processInsightsBatch(ctx, 0, insightsChan, nil, log, &insertedCounter)
		close(done)
	}()

	// Cancel context — processor should NOT exit yet, it drains the channel
	cancel()
	// Close channel — processor exits via channel close
	close(insightsChan)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("processInsightsBatch did not exit after channel close")
	}
}

func TestProcessInsightsBatch_ChannelClose_EmptyBatch(t *testing.T) {
	log := logger.New("error")
	insightsChan := make(chan *kafkamodels.ParsedInstagramInsight, 10)

	var insertedCounter uint64

	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processInsightsBatch(ctx, 0, insightsChan, nil, log, &insertedCounter)
		close(done)
	}()

	close(insightsChan)

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("processInsightsBatch did not exit after channel close")
	}
}

// ================== processPosts Tests ==================

func TestProcessPosts_EmptyBatch(t *testing.T) {
	log := logger.New("error")
	processorLog := log.With().Int("processor_id", 0).Str("type", "posts").Logger()

	batch := []*kafkamodels.ParsedInstagramPost{}
	count := processPosts(context.Background(), batch, nil, &processorLog)

	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

// ================== processInsights Tests ==================

func TestProcessInsights_EmptyBatch(t *testing.T) {
	log := logger.New("error")
	processorLog := log.With().Int("processor_id", 0).Str("type", "insights").Logger()

	batch := []*kafkamodels.ParsedInstagramInsight{}
	count := processInsights(context.Background(), batch, nil, &processorLog)

	if count != 0 {
		t.Errorf("count = %d, want 0", count)
	}
}

// ================== Integration-like Tests ==================

func TestHandleParsedPost_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 100),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 10),
	}

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		post := kafkamodels.ParsedInstagramPost{MediaID: "media" + string(rune('0'+i))}
		postJSON, _ := json.Marshal(post)
		if err := handleParsedPost(ctx, []byte("key"), postJSON, bc, log); err != nil {
			t.Errorf("handleParsedPost[%d] failed: %v", i, err)
		}
	}

	if len(bc.posts) != 10 {
		t.Errorf("posts channel len = %d, want 10", len(bc.posts))
	}
}

func TestHandleParsedInsight_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 10),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 100),
	}

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		insight := kafkamodels.ParsedInstagramInsight{InstagramID: "ig" + string(rune('0'+i))}
		insightJSON, _ := json.Marshal(insight)
		if err := handleParsedInsight(ctx, []byte("key"), insightJSON, bc, log); err != nil {
			t.Errorf("handleParsedInsight[%d] failed: %v", i, err)
		}
	}

	if len(bc.insights) != 10 {
		t.Errorf("insights channel len = %d, want 10", len(bc.insights))
	}
}

// ================== Batch Size Trigger Tests ==================

func TestProcessPostsBatch_Timeout(t *testing.T) {
	log := logger.New("error")

	postsChan := make(chan *kafkamodels.ParsedInstagramPost, 100)
	var insertedCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan struct{})
	go func() {
		processPostsBatch(ctx, 0, postsChan, nil, log, &insertedCounter)
		close(done)
	}()

	// Processor exits via channel close, not ctx cancel
	cancel()
	close(postsChan)
	<-done
}

// ================== Concurrent Access Tests ==================

func TestBatchCollectors_ConcurrentAccess(t *testing.T) {
	bc := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, 1000),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, 1000),
	}

	var wg sync.WaitGroup
	numWriters := 10
	messagesPerWriter := 50

	// Concurrent writers
	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func(writerID int) {
			defer wg.Done()
			for j := 0; j < messagesPerWriter; j++ {
				bc.posts <- &kafkamodels.ParsedInstagramPost{MediaID: "media"}
				bc.insights <- &kafkamodels.ParsedInstagramInsight{InstagramID: "ig"}
			}
		}(i)
	}

	wg.Wait()

	expectedCount := numWriters * messagesPerWriter
	if len(bc.posts) != expectedCount {
		t.Errorf("posts count = %d, want %d", len(bc.posts), expectedCount)
	}
	if len(bc.insights) != expectedCount {
		t.Errorf("insights count = %d, want %d", len(bc.insights), expectedCount)
	}
}

// ================== Atomic Counter Tests ==================

func TestAtomicCounters(t *testing.T) {
	var counter uint64

	var wg sync.WaitGroup
	numGoroutines := 100
	incrementsPerGoroutine := 1000

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < incrementsPerGoroutine; j++ {
				atomic.AddUint64(&counter, 1)
			}
		}()
	}

	wg.Wait()

	expected := uint64(numGoroutines * incrementsPerGoroutine)
	if atomic.LoadUint64(&counter) != expected {
		t.Errorf("counter = %d, want %d", counter, expected)
	}
}
