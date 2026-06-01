package main

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== RunService Tests ==================

func TestRunService_BasicFlow(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	postJSON, _ := json.Marshal(&kafkamodels.RawFacebookPost{
		ID:         "page123_post456",
		Message:    "Test post",
		StatusType: "mobile_status_update",
	})

	postsConsumer := &MockKafkaConsumer{
		Messages: []MockMessage{
			{Topic: rawPostsTopic, Key: []byte("ws_page123_post456"), Value: postJSON},
		},
	}

	miConsumer := &MockKafkaConsumer{
		Messages: []MockMessage{},
	}

	deps := ServiceDependencies{
		Sink:          sink,
		PostsConsumer: postsConsumer,
		MIConsumer:    miConsumer,
		Logger:        log,
	}

	cfg := ServiceConfig{
		PostsParserWorkers:    1,
		InsightsParserWorkers: 1,
		BatchProcessorsPerType: 1,
		MaxBatchSize:          10,
		BatchTimeout:          100 * time.Millisecond,
		IdleTimeout:           1 * time.Second,
		ParseChanSize:         10,
		MessageChanSize:       100,
	}

	ctx, cancel := context.WithCancel(context.Background())
	shutdownSignal := make(chan struct{})

	go func() {
		time.Sleep(200 * time.Millisecond)
		close(shutdownSignal)
	}()

	metrics, err := RunService(ctx, deps, cfg, shutdownSignal)
	cancel()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics == nil {
		t.Fatal("metrics should not be nil")
	}
}

func TestRunService_WithVideoAndInsights(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	videoJSON, _ := json.Marshal(&kafkamodels.RawFacebookVideo{
		ID:     "video123",
		PostID: "post456",
	})

	insightsJSON, _ := json.Marshal(&kafkamodels.RawFacebookInsights{
		PageID:      "page123",
		WorkspaceID: "ws123",
	})

	postsConsumer := &MockKafkaConsumer{Messages: []MockMessage{}}

	miConsumer := &MockKafkaConsumer{
		Messages: []MockMessage{
			{Topic: rawVideosTopic, Key: []byte("ws_page123_video123"), Value: videoJSON},
			{Topic: rawInsightsTopic, Key: []byte("ws_page123_insights"), Value: insightsJSON},
		},
	}

	deps := ServiceDependencies{
		Sink:          sink,
		PostsConsumer: postsConsumer,
		MIConsumer:    miConsumer,
		Logger:        log,
	}

	cfg := ServiceConfig{
		PostsParserWorkers:    1,
		InsightsParserWorkers: 1,
		BatchProcessorsPerType: 1,
		MaxBatchSize:          10,
		BatchTimeout:          100 * time.Millisecond,
		IdleTimeout:           1 * time.Second,
		ParseChanSize:         10,
		MessageChanSize:       100,
	}

	ctx, cancel := context.WithCancel(context.Background())
	shutdownSignal := make(chan struct{})

	go func() {
		time.Sleep(200 * time.Millisecond)
		close(shutdownSignal)
	}()

	metrics, err := RunService(ctx, deps, cfg, shutdownSignal)
	cancel()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if metrics == nil {
		t.Fatal("metrics should not be nil")
	}
}

func TestRunService_HealthCheckFails(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()
	sink.HealthFunc = func() error {
		return errors.New("clickhouse unavailable")
	}

	postsConsumer := &MockKafkaConsumer{Messages: []MockMessage{}}
	miConsumer := &MockKafkaConsumer{Messages: []MockMessage{}}

	deps := ServiceDependencies{
		Sink:          sink,
		PostsConsumer: postsConsumer,
		MIConsumer:    miConsumer,
		Logger:        log,
	}

	cfg := ServiceConfig{
		PostsParserWorkers:    1,
		InsightsParserWorkers: 1,
		BatchProcessorsPerType: 1,
		MaxBatchSize:          10,
		BatchTimeout:          100 * time.Millisecond,
		IdleTimeout:           1 * time.Second,
		ParseChanSize:         10,
		MessageChanSize:       100,
	}

	ctx, cancel := context.WithCancel(context.Background())
	shutdownSignal := make(chan struct{})

	go func() {
		time.Sleep(100 * time.Millisecond)
		close(shutdownSignal)
	}()

	// Should not fail - health check failure is logged but service continues
	_, err := RunService(ctx, deps, cfg, shutdownSignal)
	cancel()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRunService_ContextCancellation(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	postsConsumer := &MockKafkaConsumer{Messages: []MockMessage{}}
	miConsumer := &MockKafkaConsumer{Messages: []MockMessage{}}

	deps := ServiceDependencies{
		Sink:          sink,
		PostsConsumer: postsConsumer,
		MIConsumer:    miConsumer,
		Logger:        log,
	}

	cfg := ServiceConfig{
		PostsParserWorkers:    1,
		InsightsParserWorkers: 1,
		BatchProcessorsPerType: 1,
		MaxBatchSize:          10,
		BatchTimeout:          100 * time.Millisecond,
		IdleTimeout:           1 * time.Second,
		ParseChanSize:         10,
		MessageChanSize:       100,
	}

	ctx, cancel := context.WithCancel(context.Background())
	shutdownSignal := make(chan struct{})

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := RunService(ctx, deps, cfg, shutdownSignal)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// ================== ServiceConfig Tests ==================

func TestDefaultServiceConfig(t *testing.T) {
	cfg := DefaultServiceConfig()

	if cfg.PostsParserWorkers != postsParserWorkers {
		t.Errorf("PostsParserWorkers = %d, want %d", cfg.PostsParserWorkers, postsParserWorkers)
	}
	if cfg.InsightsParserWorkers != insightsParserWorkers {
		t.Errorf("InsightsParserWorkers = %d, want %d", cfg.InsightsParserWorkers, insightsParserWorkers)
	}
	if cfg.BatchProcessorsPerType != batchProcessorsPerType {
		t.Errorf("BatchProcessorsPerType = %d, want %d", cfg.BatchProcessorsPerType, batchProcessorsPerType)
	}
	if cfg.MaxBatchSize != maxBatchSize {
		t.Errorf("MaxBatchSize = %d, want %d", cfg.MaxBatchSize, maxBatchSize)
	}
	if cfg.BatchTimeout != batchTimeout {
		t.Errorf("BatchTimeout = %v, want %v", cfg.BatchTimeout, batchTimeout)
	}
	if cfg.IdleTimeout != idleTimeout {
		t.Errorf("IdleTimeout = %v, want %v", cfg.IdleTimeout, idleTimeout)
	}
}

// ================== Batch Processor Interface Tests ==================

func TestProcessPostsBatchWithInterface_EmptyBatch(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookPost, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPostsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	close(in)
	<-done

	if atomic.LoadInt32(&sink.PostsInserted) != 0 {
		t.Errorf("expected 0 posts inserted for empty batch, got %d", sink.PostsInserted)
	}
}

func TestProcessPostsBatchWithInterface_SingleItem(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookPost, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPostsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookPost{PostID: "post1", PageID: "page1"}
	close(in)
	<-done

	if atomic.LoadInt32(&sink.PostsInserted) != 1 {
		t.Errorf("expected 1 post inserted, got %d", sink.PostsInserted)
	}
}

func TestProcessPostsBatchWithInterface_BatchFull(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookPost, 20)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPostsBatchWithInterface(ctx, in, sink, log, 5, 10*time.Second)
		close(done)
	}()

	// Send exactly maxBatchSize items
	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookPost{PostID: "post" + string(rune('0'+i))}
	}

	time.Sleep(50 * time.Millisecond)
	close(in)
	<-done

	if atomic.LoadInt32(&sink.PostsInserted) != 5 {
		t.Errorf("expected 5 posts inserted, got %d", sink.PostsInserted)
	}
}

func TestProcessPostsBatchWithInterface_Timeout(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookPost, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPostsBatchWithInterface(ctx, in, sink, log, 100, 50*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookPost{PostID: "post1"}
	time.Sleep(100 * time.Millisecond) // Wait for timeout
	close(in)
	<-done

	if atomic.LoadInt32(&sink.PostsInserted) != 1 {
		t.Errorf("expected 1 post inserted after timeout, got %d", sink.PostsInserted)
	}
}

func TestProcessPostsBatchWithInterface_ContextCancel(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookPost, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processPostsBatchWithInterface(ctx, in, sink, log, 100, 10*time.Second)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookPost{PostID: "post1"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&sink.PostsInserted) != 1 {
		t.Errorf("expected 1 post flushed on cancel, got %d", sink.PostsInserted)
	}
}

func TestProcessPostsBatchWithInterface_Error(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()
	sink.BulkInsertPostsFunc = func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
		return errors.New("insert failed")
	}

	in := make(chan *kafkamodels.ParsedFacebookPost, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPostsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookPost{PostID: "post1"}
	close(in)
	<-done
	// Should not panic
}

func TestProcessMediaAssetsBatchWithInterface_SingleItem(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookMediaAsset, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processMediaAssetsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post1"}
	close(in)
	<-done

	if atomic.LoadInt32(&sink.MediaAssetsInserted) != 1 {
		t.Errorf("expected 1 asset inserted, got %d", sink.MediaAssetsInserted)
	}
}

func TestProcessMediaAssetsBatchWithInterface_BatchFull(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookMediaAsset, 20)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processMediaAssetsBatchWithInterface(ctx, in, sink, log, 5, 10*time.Second)
		close(done)
	}()

	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post" + string(rune('0'+i))}
	}

	time.Sleep(50 * time.Millisecond)
	close(in)
	<-done

	if atomic.LoadInt32(&sink.MediaAssetsInserted) != 5 {
		t.Errorf("expected 5 assets inserted, got %d", sink.MediaAssetsInserted)
	}
}

func TestProcessPageInsightsBatchWithInterface_SingleItem(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPageInsightsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookInsights{PageID: "page1"}
	close(in)
	<-done

	if atomic.LoadInt32(&sink.InsightsInserted) != 1 {
		t.Errorf("expected 1 insight inserted, got %d", sink.InsightsInserted)
	}
}

func TestProcessPageInsightsBatchWithInterface_BatchFull(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookInsights, 20)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPageInsightsBatchWithInterface(ctx, in, sink, log, 5, 10*time.Second)
		close(done)
	}()

	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookInsights{PageID: "page" + string(rune('0'+i))}
	}

	time.Sleep(50 * time.Millisecond)
	close(in)
	<-done

	if atomic.LoadInt32(&sink.InsightsInserted) != 5 {
		t.Errorf("expected 5 insights inserted, got %d", sink.InsightsInserted)
	}
}

func TestProcessVideoInsightsBatchWithInterface_SingleItem(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookVideoInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processVideoInsightsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video1"}
	close(in)
	<-done

	if atomic.LoadInt32(&sink.VideoInsightsInserted) != 1 {
		t.Errorf("expected 1 video insight inserted, got %d", sink.VideoInsightsInserted)
	}
}

func TestProcessVideoInsightsBatchWithInterface_BatchFull(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookVideoInsights, 20)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processVideoInsightsBatchWithInterface(ctx, in, sink, log, 5, 10*time.Second)
		close(done)
	}()

	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video" + string(rune('0'+i))}
	}

	time.Sleep(50 * time.Millisecond)
	close(in)
	<-done

	if atomic.LoadInt32(&sink.VideoInsightsInserted) != 5 {
		t.Errorf("expected 5 video insights inserted, got %d", sink.VideoInsightsInserted)
	}
}

func TestProcessReelsInsightsBatchWithInterface_SingleItem(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookReelsInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processReelsInsightsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}
	close(in)
	<-done

	if atomic.LoadInt32(&sink.ReelsInsightsInserted) != 1 {
		t.Errorf("expected 1 reel insight inserted, got %d", sink.ReelsInsightsInserted)
	}
}

func TestProcessReelsInsightsBatchWithInterface_BatchFull(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookReelsInsights, 20)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processReelsInsightsBatchWithInterface(ctx, in, sink, log, 5, 10*time.Second)
		close(done)
	}()

	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel" + string(rune('0'+i))}
	}

	time.Sleep(50 * time.Millisecond)
	close(in)
	<-done

	if atomic.LoadInt32(&sink.ReelsInsightsInserted) != 5 {
		t.Errorf("expected 5 reel insights inserted, got %d", sink.ReelsInsightsInserted)
	}
}

// ================== startBatchProcessorsWithInterface Tests ==================

func TestStartBatchProcessorsWithInterface(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	batches := &BatchCollectors{
		posts:         make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	startBatchProcessorsWithInterface(ctx, batches, sink, log, &wg, 2, 10, 100*time.Millisecond)

	// Send items to each channel
	batches.posts <- &kafkamodels.ParsedFacebookPost{PostID: "post1"}
	batches.mediaAssets <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post1"}
	batches.pageInsights <- &kafkamodels.ParsedFacebookInsights{PageID: "page1"}
	batches.videoInsights <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video1"}
	batches.reelsInsights <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}

	time.Sleep(50 * time.Millisecond)

	close(batches.posts)
	close(batches.mediaAssets)
	close(batches.pageInsights)
	close(batches.videoInsights)
	close(batches.reelsInsights)

	cancel()
	wg.Wait()

	if atomic.LoadInt32(&sink.PostsInserted) < 1 {
		t.Error("expected at least 1 post inserted")
	}
	if atomic.LoadInt32(&sink.MediaAssetsInserted) < 1 {
		t.Error("expected at least 1 media asset inserted")
	}
	if atomic.LoadInt32(&sink.InsightsInserted) < 1 {
		t.Error("expected at least 1 insight inserted")
	}
	if atomic.LoadInt32(&sink.VideoInsightsInserted) < 1 {
		t.Error("expected at least 1 video insight inserted")
	}
	if atomic.LoadInt32(&sink.ReelsInsightsInserted) < 1 {
		t.Error("expected at least 1 reel insight inserted")
	}
}

// ================== MockKafkaConsumer Tests ==================

func TestMockKafkaConsumer_Consume(t *testing.T) {
	consumer := &MockKafkaConsumer{
		Messages: []MockMessage{
			{Topic: "test-topic", Key: []byte("key1"), Value: []byte("value1")},
			{Topic: "test-topic", Key: []byte("key2"), Value: []byte("value2")},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	var received int

	go func() {
		consumer.Consume(ctx, []string{"test-topic"}, func(ctx context.Context, topic string, key, value []byte) error {
			received++
			return nil
		})
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	if received != 2 {
		t.Errorf("expected 2 messages received, got %d", received)
	}
}

func TestMockKafkaConsumer_Close(t *testing.T) {
	consumer := &MockKafkaConsumer{}

	if consumer.Closed {
		t.Error("consumer should not be closed initially")
	}

	consumer.Close()

	if !consumer.Closed {
		t.Error("consumer should be closed after Close()")
	}
}

// ================== MockClickHouseSink Tests ==================

func TestMockClickHouseSink_DefaultBehavior(t *testing.T) {
	sink := NewMockClickHouseSink()

	// Test Health
	if err := sink.Health(); err != nil {
		t.Errorf("Health() should return nil by default, got %v", err)
	}

	// Test Convert methods
	post := sink.ConvertFacebookPost(&kafkamodels.ParsedFacebookPost{PostID: "p1", PageID: "page1"})
	if post.PostID != "p1" {
		t.Errorf("ConvertFacebookPost PostID = %q, want 'p1'", post.PostID)
	}

	asset := sink.ConvertFacebookMediaAssets(&kafkamodels.ParsedFacebookMediaAsset{PostID: "p1"})
	if asset.PostID != "p1" {
		t.Errorf("ConvertFacebookMediaAssets PostID = %q, want 'p1'", asset.PostID)
	}

	insight := sink.ConvertFacebookInsights(&kafkamodels.ParsedFacebookInsights{PageID: "page1"})
	if insight.PageID != "page1" {
		t.Errorf("ConvertFacebookInsights PageID = %q, want 'page1'", insight.PageID)
	}

	videoIns := sink.ConvertFacebookVideoInsights(&kafkamodels.ParsedFacebookVideoInsights{PostID: "v1"})
	if videoIns.PostID != "v1" {
		t.Errorf("ConvertFacebookVideoInsights PostID = %q, want 'v1'", videoIns.PostID)
	}

	reelIns := sink.ConvertFacebookReelsInsights(&kafkamodels.ParsedFacebookReelsInsights{PostID: "r1"})
	if reelIns.PostID != "r1" {
		t.Errorf("ConvertFacebookReelsInsights PostID = %q, want 'r1'", reelIns.PostID)
	}
}

// ================== Error Path Tests ==================

func TestProcessMediaAssetsBatchWithInterface_Error(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()
	sink.BulkInsertMediaAssetsFunc = func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
		return errors.New("insert failed")
	}

	in := make(chan *kafkamodels.ParsedFacebookMediaAsset, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processMediaAssetsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post1"}
	close(in)
	<-done
	// Should not panic
}

func TestProcessPageInsightsBatchWithInterface_Error(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()
	sink.BulkInsertInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
		return errors.New("insert failed")
	}

	in := make(chan *kafkamodels.ParsedFacebookInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPageInsightsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookInsights{PageID: "page1"}
	close(in)
	<-done
	// Should not panic
}

func TestProcessVideoInsightsBatchWithInterface_Error(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()
	sink.BulkInsertVideoInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
		return errors.New("insert failed")
	}

	in := make(chan *kafkamodels.ParsedFacebookVideoInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processVideoInsightsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video1"}
	close(in)
	<-done
	// Should not panic
}

func TestProcessReelsInsightsBatchWithInterface_Error(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()
	sink.BulkInsertReelsInsightsFunc = func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
		return errors.New("insert failed")
	}

	in := make(chan *kafkamodels.ParsedFacebookReelsInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processReelsInsightsBatchWithInterface(ctx, in, sink, log, 10, 100*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}
	close(in)
	<-done
	// Should not panic
}

// ================== Context Cancellation Tests ==================

func TestProcessMediaAssetsBatchWithInterface_ContextCancel(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookMediaAsset, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processMediaAssetsBatchWithInterface(ctx, in, sink, log, 100, 10*time.Second)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post1"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&sink.MediaAssetsInserted) != 1 {
		t.Errorf("expected 1 asset flushed on cancel, got %d", sink.MediaAssetsInserted)
	}
}

func TestProcessPageInsightsBatchWithInterface_ContextCancel(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookInsights, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processPageInsightsBatchWithInterface(ctx, in, sink, log, 100, 10*time.Second)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookInsights{PageID: "page1"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&sink.InsightsInserted) != 1 {
		t.Errorf("expected 1 insight flushed on cancel, got %d", sink.InsightsInserted)
	}
}

func TestProcessVideoInsightsBatchWithInterface_ContextCancel(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookVideoInsights, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processVideoInsightsBatchWithInterface(ctx, in, sink, log, 100, 10*time.Second)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video1"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&sink.VideoInsightsInserted) != 1 {
		t.Errorf("expected 1 video insight flushed on cancel, got %d", sink.VideoInsightsInserted)
	}
}

func TestProcessReelsInsightsBatchWithInterface_ContextCancel(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookReelsInsights, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processReelsInsightsBatchWithInterface(ctx, in, sink, log, 100, 10*time.Second)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&sink.ReelsInsightsInserted) != 1 {
		t.Errorf("expected 1 reel insight flushed on cancel, got %d", sink.ReelsInsightsInserted)
	}
}

// ================== Timeout Flush Tests ==================

func TestProcessMediaAssetsBatchWithInterface_Timeout(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookMediaAsset, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processMediaAssetsBatchWithInterface(ctx, in, sink, log, 100, 50*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post1"}
	time.Sleep(100 * time.Millisecond)
	close(in)
	<-done

	if atomic.LoadInt32(&sink.MediaAssetsInserted) != 1 {
		t.Errorf("expected 1 asset inserted after timeout, got %d", sink.MediaAssetsInserted)
	}
}

func TestProcessPageInsightsBatchWithInterface_Timeout(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPageInsightsBatchWithInterface(ctx, in, sink, log, 100, 50*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookInsights{PageID: "page1"}
	time.Sleep(100 * time.Millisecond)
	close(in)
	<-done

	if atomic.LoadInt32(&sink.InsightsInserted) != 1 {
		t.Errorf("expected 1 insight inserted after timeout, got %d", sink.InsightsInserted)
	}
}

func TestProcessVideoInsightsBatchWithInterface_Timeout(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookVideoInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processVideoInsightsBatchWithInterface(ctx, in, sink, log, 100, 50*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video1"}
	time.Sleep(100 * time.Millisecond)
	close(in)
	<-done

	if atomic.LoadInt32(&sink.VideoInsightsInserted) != 1 {
		t.Errorf("expected 1 video insight inserted after timeout, got %d", sink.VideoInsightsInserted)
	}
}

func TestProcessReelsInsightsBatchWithInterface_Timeout(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	in := make(chan *kafkamodels.ParsedFacebookReelsInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processReelsInsightsBatchWithInterface(ctx, in, sink, log, 100, 50*time.Millisecond)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}
	time.Sleep(100 * time.Millisecond)
	close(in)
	<-done

	if atomic.LoadInt32(&sink.ReelsInsightsInserted) != 1 {
		t.Errorf("expected 1 reel insight inserted after timeout, got %d", sink.ReelsInsightsInserted)
	}
}
