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
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/rs/zerolog"
)

func TestHandleParsedPost_Success(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		posts: make(chan *kafkamodels.ParsedFacebookPost, 10),
	}

	post := kafkamodels.ParsedFacebookPost{
		PageID: "page123",
		PostID: "post456",
	}
	postJSON, _ := json.Marshal(post)

	err := handleParsedPost(context.Background(), []byte("key1"), postJSON, batches, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case received := <-batches.posts:
		if received.PageID != "page123" {
			t.Fatalf("expected PageID 'page123', got '%s'", received.PageID)
		}
		if received.PostID != "post456" {
			t.Fatalf("expected PostID 'post456', got '%s'", received.PostID)
		}
	default:
		t.Fatal("expected post in channel")
	}
}

func TestHandleParsedPost_InvalidJSON(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		posts: make(chan *kafkamodels.ParsedFacebookPost, 10),
	}

	err := handleParsedPost(context.Background(), []byte("key1"), []byte("invalid json"), batches, log)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleParsedPost_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		posts: make(chan *kafkamodels.ParsedFacebookPost), // unbuffered, will block
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	post := kafkamodels.ParsedFacebookPost{PageID: "page123"}
	postJSON, _ := json.Marshal(post)

	err := handleParsedPost(ctx, []byte("key1"), postJSON, batches, log)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}

func TestHandleParsedMediaAsset_Success(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
	}

	asset := kafkamodels.ParsedFacebookMediaAsset{
		PostID:    "post123",
		AssetType: "photo",
	}
	assetJSON, _ := json.Marshal(asset)

	err := handleParsedMediaAsset(context.Background(), []byte("key1"), assetJSON, batches, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case received := <-batches.mediaAssets:
		if received.PostID != "post123" {
			t.Fatalf("expected PostID 'post123', got '%s'", received.PostID)
		}
	default:
		t.Fatal("expected asset in channel")
	}
}

func TestHandleParsedMediaAsset_InvalidJSON(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
	}

	err := handleParsedMediaAsset(context.Background(), []byte("key1"), []byte("invalid"), batches, log)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleParsedPageInsights_Success(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		pageInsights: make(chan *kafkamodels.ParsedFacebookInsights, 10),
	}

	// Page insights come as a batch (slice)
	insights := []*kafkamodels.ParsedFacebookInsights{
		{PageID: "page123"},
		{PageID: "page123"},
	}
	insightsJSON, _ := json.Marshal(insights)

	err := handleParsedPageInsights(context.Background(), []byte("key1"), insightsJSON, batches, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should receive 2 insights
	count := 0
	for i := 0; i < 2; i++ {
		select {
		case received := <-batches.pageInsights:
			if received.PageID != "page123" {
				t.Fatalf("expected PageID 'page123', got '%s'", received.PageID)
			}
			count++
		default:
			t.Fatalf("expected insight %d in channel", i+1)
		}
	}
	if count != 2 {
		t.Fatalf("expected 2 insights, got %d", count)
	}
}

func TestHandleParsedPageInsights_InvalidJSON(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		pageInsights: make(chan *kafkamodels.ParsedFacebookInsights, 10),
	}

	err := handleParsedPageInsights(context.Background(), []byte("key1"), []byte("invalid"), batches, log)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleParsedVideoInsights_Success(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
	}

	insights := kafkamodels.ParsedFacebookVideoInsights{
		PostID:  "video123",
		VideoID: "v456",
	}
	insightsJSON, _ := json.Marshal(insights)

	err := handleParsedVideoInsights(context.Background(), []byte("key1"), insightsJSON, batches, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case received := <-batches.videoInsights:
		if received.PostID != "video123" {
			t.Fatalf("expected PostID 'video123', got '%s'", received.PostID)
		}
	default:
		t.Fatal("expected video insight in channel")
	}
}

func TestHandleParsedVideoInsights_InvalidJSON(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
	}

	err := handleParsedVideoInsights(context.Background(), []byte("key1"), []byte("invalid"), batches, log)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleParsedReelsInsights_Success(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	insights := kafkamodels.ParsedFacebookReelsInsights{
		PostID:    "reel123",
		PlayCount: 1000,
	}
	insightsJSON, _ := json.Marshal(insights)

	err := handleParsedReelsInsights(context.Background(), []byte("key1"), insightsJSON, batches, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case received := <-batches.reelsInsights:
		if received.PostID != "reel123" {
			t.Fatalf("expected PostID 'reel123', got '%s'", received.PostID)
		}
		if received.PlayCount != 1000 {
			t.Fatalf("expected PlayCount 1000, got %d", received.PlayCount)
		}
	default:
		t.Fatal("expected reel insight in channel")
	}
}

func TestHandleParsedReelsInsights_InvalidJSON(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	err := handleParsedReelsInsights(context.Background(), []byte("key1"), []byte("invalid"), batches, log)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestBatchCollectors_Struct(t *testing.T) {
	batches := &BatchCollectors{
		posts:         make(chan *kafkamodels.ParsedFacebookPost, 100),
		mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, 100),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 100),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 100),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 100),
	}

	if cap(batches.posts) != 100 {
		t.Fatalf("expected posts channel capacity 100, got %d", cap(batches.posts))
	}
	if cap(batches.mediaAssets) != 100 {
		t.Fatalf("expected mediaAssets channel capacity 100, got %d", cap(batches.mediaAssets))
	}
	if cap(batches.pageInsights) != 100 {
		t.Fatalf("expected pageInsights channel capacity 100, got %d", cap(batches.pageInsights))
	}
	if cap(batches.videoInsights) != 100 {
		t.Fatalf("expected videoInsights channel capacity 100, got %d", cap(batches.videoInsights))
	}
	if cap(batches.reelsInsights) != 100 {
		t.Fatalf("expected reelsInsights channel capacity 100, got %d", cap(batches.reelsInsights))
	}
}

func TestMessage_Struct(t *testing.T) {
	msg := Message{
		Topic: "test-topic",
		Key:   []byte("key1"),
		Value: []byte("value1"),
	}

	if msg.Topic != "test-topic" {
		t.Fatalf("expected Topic 'test-topic', got '%s'", msg.Topic)
	}
	if string(msg.Key) != "key1" {
		t.Fatalf("expected Key 'key1', got '%s'", string(msg.Key))
	}
	if string(msg.Value) != "value1" {
		t.Fatalf("expected Value 'value1', got '%s'", string(msg.Value))
	}
}

// ================== Worker Tests ==================

func TestPostsAssetsWorker_Posts(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		posts:       make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
	}

	msgChan := make(chan Message, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go postsAssetsWorker(ctx, 1, msgChan, batches, log, &wg)

	// Send a post message
	post := kafkamodels.ParsedFacebookPost{PageID: "page123", PostID: "post456"}
	postJSON, _ := json.Marshal(post)
	msgChan <- Message{Topic: topicPosts, Key: []byte("key1"), Value: postJSON}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	select {
	case received := <-batches.posts:
		if received.PageID != "page123" {
			t.Fatalf("expected PageID 'page123', got '%s'", received.PageID)
		}
	default:
		t.Fatal("expected post in channel")
	}

	cancel()
	close(msgChan)
	wg.Wait()
}

func TestPostsAssetsWorker_MediaAssets(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		posts:       make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
	}

	msgChan := make(chan Message, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go postsAssetsWorker(ctx, 1, msgChan, batches, log, &wg)

	// Send a media asset message
	asset := kafkamodels.ParsedFacebookMediaAsset{PostID: "post123", AssetType: "photo"}
	assetJSON, _ := json.Marshal(asset)
	msgChan <- Message{Topic: topicMediaAssets, Key: []byte("key1"), Value: assetJSON}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	select {
	case received := <-batches.mediaAssets:
		if received.PostID != "post123" {
			t.Fatalf("expected PostID 'post123', got '%s'", received.PostID)
		}
	default:
		t.Fatal("expected media asset in channel")
	}

	cancel()
	close(msgChan)
	wg.Wait()
}

func TestInsightsWorker_PageInsights(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	msgChan := make(chan Message, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go insightsWorker(ctx, 1, msgChan, batches, log, &wg)

	// Send a page insights message (batch format)
	insights := []*kafkamodels.ParsedFacebookInsights{{PageID: "page123"}}
	insightsJSON, _ := json.Marshal(insights)
	msgChan <- Message{Topic: topicInsights, Key: []byte("key1"), Value: insightsJSON}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	select {
	case received := <-batches.pageInsights:
		if received.PageID != "page123" {
			t.Fatalf("expected PageID 'page123', got '%s'", received.PageID)
		}
	default:
		t.Fatal("expected page insight in channel")
	}

	cancel()
	close(msgChan)
	wg.Wait()
}

func TestInsightsWorker_VideoInsights(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	msgChan := make(chan Message, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go insightsWorker(ctx, 1, msgChan, batches, log, &wg)

	// Send a video insights message
	insights := kafkamodels.ParsedFacebookVideoInsights{PostID: "video123"}
	insightsJSON, _ := json.Marshal(insights)
	msgChan <- Message{Topic: topicVideoInsights, Key: []byte("key1"), Value: insightsJSON}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	select {
	case received := <-batches.videoInsights:
		if received.PostID != "video123" {
			t.Fatalf("expected PostID 'video123', got '%s'", received.PostID)
		}
	default:
		t.Fatal("expected video insight in channel")
	}

	cancel()
	close(msgChan)
	wg.Wait()
}

func TestInsightsWorker_ReelsInsights(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	msgChan := make(chan Message, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go insightsWorker(ctx, 1, msgChan, batches, log, &wg)

	// Send a reels insights message
	insights := kafkamodels.ParsedFacebookReelsInsights{PostID: "reel123", PlayCount: 500}
	insightsJSON, _ := json.Marshal(insights)
	msgChan <- Message{Topic: topicReelsInsights, Key: []byte("key1"), Value: insightsJSON}

	// Wait for processing
	time.Sleep(50 * time.Millisecond)

	select {
	case received := <-batches.reelsInsights:
		if received.PostID != "reel123" {
			t.Fatalf("expected PostID 'reel123', got '%s'", received.PostID)
		}
	default:
		t.Fatal("expected reels insight in channel")
	}

	cancel()
	close(msgChan)
	wg.Wait()
}

// ================== Process Functions Tests ==================

func TestProcessPosts_Success(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			atomic.AddInt32(&insertCount, int32(len(posts)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookPost{
		{PageID: "page1", PostID: "post1"},
		{PageID: "page1", PostID: "post2"},
		{PageID: "page1", PostID: "post3"},
	}

	processPosts(context.Background(), batch, sink, log)

	if atomic.LoadInt32(&insertCount) != 3 {
		t.Fatalf("expected 3 posts inserted, got %d", insertCount)
	}
}

func TestProcessPosts_Empty(t *testing.T) {
	log := logger.New("error")
	var insertCalled bool

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			insertCalled = true
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	processPosts(context.Background(), []*kafkamodels.ParsedFacebookPost{}, sink, log)

	if insertCalled {
		t.Fatal("expected no insert call for empty batch")
	}
}

func TestProcessPosts_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			return errors.New("clickhouse error")
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookPost{
		{PageID: "page1", PostID: "post1"},
	}

	// Should not panic, just log error
	processPosts(context.Background(), batch, sink, log)
}

func TestProcessMediaAssets_Success(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			atomic.AddInt32(&insertCount, int32(len(assets)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookMediaAsset{
		{PostID: "post1", AssetType: "photo"},
		{PostID: "post2", AssetType: "video"},
	}

	processMediaAssets(context.Background(), batch, sink, log)

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Fatalf("expected 2 assets inserted, got %d", insertCount)
	}
}

func TestProcessPageInsights_Success(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			atomic.AddInt32(&insertCount, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookInsights{
		{PageID: "page1"},
		{PageID: "page2"},
	}

	processPageInsights(context.Background(), batch, sink, log)

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Fatalf("expected 2 insights inserted, got %d", insertCount)
	}
}

func TestProcessVideoInsights_Success(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			atomic.AddInt32(&insertCount, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookVideoInsights{
		{PostID: "video1"},
		{PostID: "video2"},
	}

	processVideoInsights(context.Background(), batch, sink, log)

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Fatalf("expected 2 video insights inserted, got %d", insertCount)
	}
}

func TestProcessReelsInsights_Success(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			atomic.AddInt32(&insertCount, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookReelsInsights{
		{PostID: "reel1", PlayCount: 100},
		{PostID: "reel2", PlayCount: 200},
	}

	processReelsInsights(context.Background(), batch, sink, log)

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Fatalf("expected 2 reels insights inserted, got %d", insertCount)
	}
}

// ================== Batch Processor Tests ==================

func TestProcessPostsBatch_FlushOnClose(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			atomic.AddInt32(&insertCount, int32(len(posts)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookPost, 10)
	ctx := context.Background()

	// Start batch processor
	done := make(chan struct{})
	go func() {
		processPostsBatch(ctx, in, sink, log)
		close(done)
	}()

	// Send some posts
	in <- &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post1"}
	in <- &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post2"}

	// Close channel to trigger flush
	close(in)
	<-done

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Fatalf("expected 2 posts flushed on close, got %d", insertCount)
	}
}

func TestProcessPostsBatch_FlushOnContextCancel(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			atomic.AddInt32(&insertCount, int32(len(posts)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookPost, 10)
	ctx, cancel := context.WithCancel(context.Background())

	// Start batch processor
	done := make(chan struct{})
	go func() {
		processPostsBatch(ctx, in, sink, log)
		close(done)
	}()

	// Send some posts
	in <- &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post1"}

	// Small delay to ensure post is received
	time.Sleep(10 * time.Millisecond)

	// Cancel context to trigger flush
	cancel()
	<-done

	if atomic.LoadInt32(&insertCount) != 1 {
		t.Fatalf("expected 1 post flushed on cancel, got %d", insertCount)
	}
}

// ================== Additional Batch Processor Tests ==================

func TestProcessMediaAssetsBatch_FlushOnClose(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			atomic.AddInt32(&insertCount, int32(len(assets)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookMediaAsset, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processMediaAssetsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post1"}
	in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post2"}
	close(in)
	<-done

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Fatalf("expected 2 assets flushed, got %d", insertCount)
	}
}

func TestProcessMediaAssetsBatch_FlushOnContextCancel(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			atomic.AddInt32(&insertCount, int32(len(assets)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookMediaAsset, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processMediaAssetsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post1"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&insertCount) != 1 {
		t.Fatalf("expected 1 asset flushed on cancel, got %d", insertCount)
	}
}

func TestProcessPageInsightsBatch_FlushOnClose(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			atomic.AddInt32(&insertCount, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPageInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookInsights{PageID: "page1"}
	in <- &kafkamodels.ParsedFacebookInsights{PageID: "page2"}
	close(in)
	<-done

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Fatalf("expected 2 insights flushed, got %d", insertCount)
	}
}

func TestProcessPageInsightsBatch_FlushOnContextCancel(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			atomic.AddInt32(&insertCount, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookInsights, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processPageInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookInsights{PageID: "page1"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&insertCount) != 1 {
		t.Fatalf("expected 1 insight flushed on cancel, got %d", insertCount)
	}
}

func TestProcessVideoInsightsBatch_FlushOnClose(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			atomic.AddInt32(&insertCount, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookVideoInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processVideoInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video1"}
	in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video2"}
	close(in)
	<-done

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Fatalf("expected 2 video insights flushed, got %d", insertCount)
	}
}

func TestProcessVideoInsightsBatch_FlushOnContextCancel(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			atomic.AddInt32(&insertCount, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookVideoInsights, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processVideoInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video1"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&insertCount) != 1 {
		t.Fatalf("expected 1 video insight flushed on cancel, got %d", insertCount)
	}
}

func TestProcessReelsInsightsBatch_FlushOnClose(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			atomic.AddInt32(&insertCount, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookReelsInsights, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processReelsInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}
	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel2"}
	close(in)
	<-done

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Fatalf("expected 2 reels insights flushed, got %d", insertCount)
	}
}

func TestProcessReelsInsightsBatch_FlushOnContextCancel(t *testing.T) {
	log := logger.New("error")
	var insertCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			atomic.AddInt32(&insertCount, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookReelsInsights, 10)
	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan struct{})
	go func() {
		processReelsInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}
	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&insertCount) != 1 {
		t.Fatalf("expected 1 reel insight flushed on cancel, got %d", insertCount)
	}
}

// ================== startBatchProcessors Tests ==================

func TestStartBatchProcessors(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return nil
		},
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return nil
		},
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return nil
		},
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batches := &BatchCollectors{
		posts:         make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup

	startBatchProcessors(ctx, batches, sink, log, &wg, 1)

	// Send one item to each channel
	batches.posts <- &kafkamodels.ParsedFacebookPost{PostID: "post1"}
	batches.mediaAssets <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post1"}
	batches.pageInsights <- &kafkamodels.ParsedFacebookInsights{PageID: "page1"}
	batches.videoInsights <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video1"}
	batches.reelsInsights <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}

	time.Sleep(50 * time.Millisecond)

	// Close all channels and cancel context
	close(batches.posts)
	close(batches.mediaAssets)
	close(batches.pageInsights)
	close(batches.videoInsights)
	close(batches.reelsInsights)
	cancel()

	wg.Wait()
}

// ================== Process Functions Error Tests ==================

func TestProcessMediaAssets_Empty(t *testing.T) {
	log := logger.New("error")
	var insertCalled bool

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			insertCalled = true
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	processMediaAssets(context.Background(), []*kafkamodels.ParsedFacebookMediaAsset{}, sink, log)

	if insertCalled {
		t.Fatal("expected no insert call for empty batch")
	}
}

func TestProcessMediaAssets_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return errors.New("clickhouse error")
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookMediaAsset{{PostID: "post1"}}
	processMediaAssets(context.Background(), batch, sink, log)
	// Should not panic
}

func TestProcessPageInsights_Empty(t *testing.T) {
	log := logger.New("error")
	var insertCalled bool

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			insertCalled = true
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	processPageInsights(context.Background(), []*kafkamodels.ParsedFacebookInsights{}, sink, log)

	if insertCalled {
		t.Fatal("expected no insert call for empty batch")
	}
}

func TestProcessPageInsights_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return errors.New("clickhouse error")
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookInsights{{PageID: "page1"}}
	processPageInsights(context.Background(), batch, sink, log)
	// Should not panic
}

func TestProcessVideoInsights_Empty(t *testing.T) {
	log := logger.New("error")
	var insertCalled bool

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			insertCalled = true
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	processVideoInsights(context.Background(), []*kafkamodels.ParsedFacebookVideoInsights{}, sink, log)

	if insertCalled {
		t.Fatal("expected no insert call for empty batch")
	}
}

func TestProcessVideoInsights_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return errors.New("clickhouse error")
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookVideoInsights{{PostID: "video1"}}
	processVideoInsights(context.Background(), batch, sink, log)
	// Should not panic
}

func TestProcessReelsInsights_Empty(t *testing.T) {
	log := logger.New("error")
	var insertCalled bool

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			insertCalled = true
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	processReelsInsights(context.Background(), []*kafkamodels.ParsedFacebookReelsInsights{}, sink, log)

	if insertCalled {
		t.Fatal("expected no insert call for empty batch")
	}
}

func TestProcessReelsInsights_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return errors.New("clickhouse error")
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	batch := []*kafkamodels.ParsedFacebookReelsInsights{{PostID: "reel1"}}
	processReelsInsights(context.Background(), batch, sink, log)
	// Should not panic
}

// ================== Batch Size Trigger Tests ==================

func TestProcessPostsBatch_MaxBatchSize(t *testing.T) {
	log := logger.New("error")
	var insertCounts []int

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			insertCounts = append(insertCounts, len(posts))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookPost, maxBatchSize+100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPostsBatch(ctx, in, sink, log)
		close(done)
	}()

	// Fill up to maxBatchSize to trigger a flush
	for i := 0; i < maxBatchSize; i++ {
		in <- &kafkamodels.ParsedFacebookPost{PostID: "post" + string(rune(i))}
	}

	// Wait for the batch to be processed
	time.Sleep(50 * time.Millisecond)

	// Add a few more and close
	in <- &kafkamodels.ParsedFacebookPost{PostID: "extra1"}
	in <- &kafkamodels.ParsedFacebookPost{PostID: "extra2"}
	close(in)
	<-done

	if len(insertCounts) < 2 {
		t.Fatalf("expected at least 2 insert calls (batch + remainder), got %d", len(insertCounts))
	}
	if insertCounts[0] != maxBatchSize {
		t.Fatalf("expected first batch size %d, got %d", maxBatchSize, insertCounts[0])
	}
}

func TestProcessMediaAssetsBatch_MaxBatchSize(t *testing.T) {
	log := logger.New("error")
	var insertCounts []int

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			insertCounts = append(insertCounts, len(assets))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookMediaAsset, maxBatchSize+100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processMediaAssetsBatch(ctx, in, sink, log)
		close(done)
	}()

	for i := 0; i < maxBatchSize; i++ {
		in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post" + string(rune(i))}
	}

	time.Sleep(50 * time.Millisecond)

	in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "extra1"}
	close(in)
	<-done

	if len(insertCounts) < 2 {
		t.Fatalf("expected at least 2 insert calls, got %d", len(insertCounts))
	}
	if insertCounts[0] != maxBatchSize {
		t.Fatalf("expected first batch size %d, got %d", maxBatchSize, insertCounts[0])
	}
}

func TestProcessPageInsightsBatch_MaxBatchSize(t *testing.T) {
	log := logger.New("error")
	var insertCounts []int

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			insertCounts = append(insertCounts, len(insights))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookInsights, maxBatchSize+100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPageInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	for i := 0; i < maxBatchSize; i++ {
		in <- &kafkamodels.ParsedFacebookInsights{PageID: "page" + string(rune(i))}
	}

	time.Sleep(50 * time.Millisecond)

	in <- &kafkamodels.ParsedFacebookInsights{PageID: "extra1"}
	close(in)
	<-done

	if len(insertCounts) < 2 {
		t.Fatalf("expected at least 2 insert calls, got %d", len(insertCounts))
	}
}

func TestProcessVideoInsightsBatch_MaxBatchSize(t *testing.T) {
	log := logger.New("error")
	var insertCounts []int

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			insertCounts = append(insertCounts, len(insights))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookVideoInsights, maxBatchSize+100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processVideoInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	for i := 0; i < maxBatchSize; i++ {
		in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video" + string(rune(i))}
	}

	time.Sleep(50 * time.Millisecond)

	in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "extra1"}
	close(in)
	<-done

	if len(insertCounts) < 2 {
		t.Fatalf("expected at least 2 insert calls, got %d", len(insertCounts))
	}
}

func TestProcessReelsInsightsBatch_MaxBatchSize(t *testing.T) {
	log := logger.New("error")
	var insertCounts []int

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			insertCounts = append(insertCounts, len(insights))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookReelsInsights, maxBatchSize+100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processReelsInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	for i := 0; i < maxBatchSize; i++ {
		in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel" + string(rune(i))}
	}

	time.Sleep(50 * time.Millisecond)

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "extra1"}
	close(in)
	<-done

	if len(insertCounts) < 2 {
		t.Fatalf("expected at least 2 insert calls, got %d", len(insertCounts))
	}
}

// ================== Worker Unknown Topic Tests ==================

func TestPostsAssetsWorker_UnknownTopic(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		posts:       make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
	}

	msgChan := make(chan Message, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go postsAssetsWorker(ctx, 1, msgChan, batches, log, &wg)

	// Send message with unknown topic
	msgChan <- Message{Topic: "unknown-topic", Key: []byte("key1"), Value: []byte("{}")}

	time.Sleep(50 * time.Millisecond)

	// No item should be in any channel
	select {
	case <-batches.posts:
		t.Fatal("unexpected post in channel")
	case <-batches.mediaAssets:
		t.Fatal("unexpected media asset in channel")
	default:
		// Expected - nothing in channels
	}

	cancel()
	close(msgChan)
	wg.Wait()
}

func TestInsightsWorker_UnknownTopic(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	msgChan := make(chan Message, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go insightsWorker(ctx, 1, msgChan, batches, log, &wg)

	// Send message with unknown topic
	msgChan <- Message{Topic: "unknown-topic", Key: []byte("key1"), Value: []byte("{}")}

	time.Sleep(50 * time.Millisecond)

	select {
	case <-batches.pageInsights:
		t.Fatal("unexpected page insight in channel")
	case <-batches.videoInsights:
		t.Fatal("unexpected video insight in channel")
	case <-batches.reelsInsights:
		t.Fatal("unexpected reel insight in channel")
	default:
		// Expected - nothing in channels
	}

	cancel()
	close(msgChan)
	wg.Wait()
}

// ================== Handler Context Cancellation Tests ==================

func TestHandleParsedMediaAsset_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset), // unbuffered
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	asset := kafkamodels.ParsedFacebookMediaAsset{PostID: "post123"}
	assetJSON, _ := json.Marshal(asset)

	err := handleParsedMediaAsset(ctx, []byte("key1"), assetJSON, batches, log)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestHandleParsedPageInsights_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		pageInsights: make(chan *kafkamodels.ParsedFacebookInsights), // unbuffered
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	insights := []*kafkamodels.ParsedFacebookInsights{{PageID: "page123"}}
	insightsJSON, _ := json.Marshal(insights)

	err := handleParsedPageInsights(ctx, []byte("key1"), insightsJSON, batches, log)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestHandleParsedVideoInsights_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights), // unbuffered
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	insights := kafkamodels.ParsedFacebookVideoInsights{PostID: "video123"}
	insightsJSON, _ := json.Marshal(insights)

	err := handleParsedVideoInsights(ctx, []byte("key1"), insightsJSON, batches, log)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

func TestHandleParsedReelsInsights_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights), // unbuffered
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	insights := kafkamodels.ParsedFacebookReelsInsights{PostID: "reel123"}
	insightsJSON, _ := json.Marshal(insights)

	err := handleParsedReelsInsights(ctx, []byte("key1"), insightsJSON, batches, log)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got %v", err)
	}
}

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if topicPosts != "parsed-facebook-posts" {
		t.Errorf("topicPosts = %q, want %q", topicPosts, "parsed-facebook-posts")
	}
	if topicMediaAssets != "parsed-facebook-media-assets" {
		t.Errorf("topicMediaAssets = %q, want %q", topicMediaAssets, "parsed-facebook-media-assets")
	}
	if topicInsights != "parsed-facebook-insights" {
		t.Errorf("topicInsights = %q, want %q", topicInsights, "parsed-facebook-insights")
	}
	if topicVideoInsights != "parsed-facebook-video-insights" {
		t.Errorf("topicVideoInsights = %q, want %q", topicVideoInsights, "parsed-facebook-video-insights")
	}
	if topicReelsInsights != "parsed-facebook-reels-insights" {
		t.Errorf("topicReelsInsights = %q, want %q", topicReelsInsights, "parsed-facebook-reels-insights")
	}
	if maxBatchSize != 10000 {
		t.Errorf("maxBatchSize = %d, want %d", maxBatchSize, 10000)
	}
	if batchTimeout != 10*time.Second {
		t.Errorf("batchTimeout = %v, want %v", batchTimeout, 10*time.Second)
	}
}
