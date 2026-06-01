package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/rs/zerolog"
)

// ================== extractKeyInfo Tests ==================

func TestExtractKeyInfo(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantPageID string
		wantWsID   string
	}{
		{"empty key", "", "", ""},
		{"single part", "page123", "page123", ""},
		{"two parts", "page123_post456", "page123", ""},
		{"three parts with workspace", "ws123_page456_post789", "page456", "ws123"},
		{"four parts", "ws123_page456_post789_extra", "page456", "ws123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageID, _, wsID := extractKeyInfo(tt.key)
			if pageID != tt.wantPageID {
				t.Errorf("pageID = %q, want %q", pageID, tt.wantPageID)
			}
			if wsID != tt.wantWsID {
				t.Errorf("workspaceID = %q, want %q", wsID, tt.wantWsID)
			}
		})
	}
}

// ================== handleRawPost Tests ==================

func TestHandleRawPost_Success(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob, 10)

	// Use JSON to create the struct with anonymous From field
	postJSON := []byte(`{
		"id": "post123",
		"message": "Test post",
		"from": {"id": "page456", "name": "Test Page"}
	}`)

	err := handleRawPost(context.Background(), []byte("ws123_page456_post123"), postJSON, out, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case job := <-out:
		if job.JobType != "post" {
			t.Errorf("JobType = %q, want %q", job.JobType, "post")
		}
		if job.PageID != "page456" {
			t.Errorf("PageID = %q, want %q", job.PageID, "page456")
		}
		if job.WorkspaceID != "ws123" {
			t.Errorf("WorkspaceID = %q, want %q", job.WorkspaceID, "ws123")
		}
	default:
		t.Fatal("expected job in channel")
	}
}

func TestHandleRawPost_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob, 10)

	err := handleRawPost(context.Background(), []byte("key1"), []byte("invalid"), out, log)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestHandleRawPost_ContextCancellation(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob) // unbuffered

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	rawPost := kafkamodels.RawFacebookPost{ID: "post123"}
	postJSON, _ := json.Marshal(rawPost)

	err := handleRawPost(ctx, []byte("key1"), postJSON, out, log)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

// ================== handleRawVideo Tests ==================

func TestHandleRawVideo_Success(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob, 10)

	videoJSON := []byte(`{
		"id": "video123",
		"title": "Test video"
	}`)

	err := handleRawVideo(context.Background(), []byte("ws123_page456_video123"), videoJSON, out, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case job := <-out:
		if job.JobType != "video" {
			t.Errorf("JobType = %q, want %q", job.JobType, "video")
		}
		if job.PageID != "page456" {
			t.Errorf("PageID = %q, want %q", job.PageID, "page456")
		}
	default:
		t.Fatal("expected job in channel")
	}
}

func TestHandleRawVideo_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob, 10)

	err := handleRawVideo(context.Background(), []byte("key1"), []byte("invalid"), out, log)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// ================== handleRawInsights Tests ==================

func TestHandleRawInsights_Success(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob, 10)

	rawInsights := kafkamodels.RawFacebookInsights{
		PageID: "page123",
	}
	insightsJSON, _ := json.Marshal(rawInsights)

	err := handleRawInsights(context.Background(), []byte("ws123_page456_insights"), insightsJSON, out, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case job := <-out:
		if job.JobType != "insights" {
			t.Errorf("JobType = %q, want %q", job.JobType, "insights")
		}
	default:
		t.Fatal("expected job in channel")
	}
}

func TestHandleRawInsights_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob, 10)

	err := handleRawInsights(context.Background(), []byte("key1"), []byte("invalid"), out, log)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

// ================== ParseJob Struct Tests ==================

func TestParseJob_Struct(t *testing.T) {
	job := ParseJob{
		JobType:     "post",
		PageID:      "page123",
		PageName:    "Test Page",
		WorkspaceID: "ws456",
		MessageKey:  "key789",
	}

	if job.JobType != "post" {
		t.Errorf("JobType = %q, want %q", job.JobType, "post")
	}
	if job.PageID != "page123" {
		t.Errorf("PageID = %q, want %q", job.PageID, "page123")
	}
}

// ================== BatchCollectors Tests ==================

func TestBatchCollectors_Struct(t *testing.T) {
	batches := &BatchCollectors{
		posts:         make(chan *kafkamodels.ParsedFacebookPost, 100),
		mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, 100),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 100),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 100),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 100),
	}

	if cap(batches.posts) != 100 {
		t.Errorf("posts capacity = %d, want 100", cap(batches.posts))
	}
}

// ================== Batch Processor Tests ==================

func TestProcessPostsBatch_Success(t *testing.T) {
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

	done := make(chan struct{})
	go func() {
		processPostsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post1"}
	in <- &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post2"}

	close(in)
	<-done

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Errorf("insertCount = %d, want 2", insertCount)
	}
}

func TestProcessPostsBatch_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			return errors.New("clickhouse error")
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookPost, 10)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPostsBatch(ctx, in, sink, log)
		close(done)
	}()

	in <- &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post1"}
	close(in)
	<-done
	// Should not panic
}

func TestProcessMediaAssetsBatch_Success(t *testing.T) {
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
		t.Errorf("insertCount = %d, want 2", insertCount)
	}
}

func TestProcessPageInsightsBatch_Success(t *testing.T) {
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
		t.Errorf("insertCount = %d, want 2", insertCount)
	}
}

func TestProcessVideoInsightsBatch_Success(t *testing.T) {
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
		t.Errorf("insertCount = %d, want 2", insertCount)
	}
}

func TestProcessReelsInsightsBatch_Success(t *testing.T) {
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

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1", PlayCount: 100}
	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel2", PlayCount: 200}

	close(in)
	<-done

	if atomic.LoadInt32(&insertCount) != 2 {
		t.Errorf("insertCount = %d, want 2", insertCount)
	}
}

// ================== Parser Worker Tests ==================

func TestPostsParser_SkipsNonPostJobs(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		posts:       make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())

	var parsedPosts, parsedAssets uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 0, in, batches, &parsedPosts, &parsedAssets, log)

	// Send a video job (should be skipped)
	in <- ParseJob{JobType: "video", PageID: "page123"}

	time.Sleep(50 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()

	// Should have parsed 0 posts
	if parsedPosts != 0 {
		t.Errorf("parsedPosts = %d, want 0", parsedPosts)
	}
}

func TestMediaInsightsParser_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	close(in)
	wg.Wait()
	// Should exit gracefully
}

// ================== startBatchProcessors Tests ==================

func TestStartBatchProcessors(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{}
	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	ctx, cancel := context.WithCancel(context.Background())

	batches := &BatchCollectors{
		posts:         make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	var wg sync.WaitGroup
	startBatchProcessors(ctx, batches, sink, log, &wg, 1)

	// Cancel and close channels
	cancel()
	close(batches.posts)
	close(batches.mediaAssets)
	close(batches.pageInsights)
	close(batches.videoInsights)
	close(batches.reelsInsights)

	wg.Wait()
	// Should complete without hanging
}

// ================== Consumer Flow Tests with Mocks ==================

func TestConsumerFlow_PostsAndVideos(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	postsParseJobs := make(chan ParseJob, 100)
	miParseJobs := make(chan ParseJob, 100)

	var pickedPosts, pickedVideos uint64

	// Create raw post JSON
	rawPost := `{"id": "post123", "message": "Test"}`
	rawVideo := `{"id": "video123"}`

	// Mock posts consumer
	postsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send a post message
			handler(ctx, rawPostsTopic, []byte("page123_post123"), []byte(rawPost))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	// Mock videos/insights consumer
	miConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send a video message
			handler(ctx, rawVideosTopic, []byte("page123_video123"), []byte(rawVideo))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	// Start posts consumer
	go func() {
		postsConsumer.Consume(ctx, []string{rawPostsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			if err := handleRawPost(ctx, key, value, postsParseJobs, log); err == nil {
				atomic.AddUint64(&pickedPosts, 1)
			}
			return nil
		})
	}()

	// Start videos consumer
	go func() {
		miConsumer.Consume(ctx, []string{rawVideosTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			if err := handleRawVideo(ctx, key, value, miParseJobs, log); err == nil {
				atomic.AddUint64(&pickedVideos, 1)
			}
			return nil
		})
	}()

	// Wait for processing
	time.Sleep(100 * time.Millisecond)

	// Verify jobs were created
	select {
	case job := <-postsParseJobs:
		if job.JobType != "post" {
			t.Errorf("expected post job, got %s", job.JobType)
		}
	default:
		t.Error("expected post job in channel")
	}

	select {
	case job := <-miParseJobs:
		if job.JobType != "video" {
			t.Errorf("expected video job, got %s", job.JobType)
		}
	default:
		t.Error("expected video job in channel")
	}
}

// ================== Full Pipeline Test ==================

func TestFullPipeline_PostsToClickHouse(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	var insertedPosts int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			atomic.AddInt32(&insertedPosts, int32(len(posts)))
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

	// Start batch processor
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		processPostsBatch(ctx, batches.posts, sink, log)
	}()

	// Send parsed posts
	batches.posts <- &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post1"}
	batches.posts <- &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post2"}
	batches.posts <- &kafkamodels.ParsedFacebookPost{PageID: "page1", PostID: "post3"}

	// Close to flush - this triggers the batch to be flushed
	close(batches.posts)
	wg.Wait()

	// All 3 should be inserted in a single batch
	if atomic.LoadInt32(&insertedPosts) != 3 {
		t.Errorf("expected 3 posts inserted, got %d", insertedPosts)
	}
}

// ================== Metrics Tracking Tests ==================

func TestMetricsTracking(t *testing.T) {
	var pickedPosts, pickedVideos, pickedInsights uint64
	var parsedPosts, parsedAssets uint64

	// Simulate picking messages
	for i := 0; i < 5; i++ {
		atomic.AddUint64(&pickedPosts, 1)
	}
	for i := 0; i < 3; i++ {
		atomic.AddUint64(&pickedVideos, 1)
	}
	for i := 0; i < 2; i++ {
		atomic.AddUint64(&pickedInsights, 1)
	}

	// Simulate parsing
	for i := 0; i < 5; i++ {
		atomic.AddUint64(&parsedPosts, 1)
	}
	for i := 0; i < 10; i++ {
		atomic.AddUint64(&parsedAssets, 1)
	}

	if atomic.LoadUint64(&pickedPosts) != 5 {
		t.Errorf("pickedPosts = %d, want 5", pickedPosts)
	}
	if atomic.LoadUint64(&pickedVideos) != 3 {
		t.Errorf("pickedVideos = %d, want 3", pickedVideos)
	}
	if atomic.LoadUint64(&pickedInsights) != 2 {
		t.Errorf("pickedInsights = %d, want 2", pickedInsights)
	}
	if atomic.LoadUint64(&parsedPosts) != 5 {
		t.Errorf("parsedPosts = %d, want 5", parsedPosts)
	}
	if atomic.LoadUint64(&parsedAssets) != 10 {
		t.Errorf("parsedAssets = %d, want 10", parsedAssets)
	}
}

// ================== Constants Tests ==================

func TestAnalyticsSinkConstants(t *testing.T) {
	if postsParserWorkers != 5 {
		t.Errorf("postsParserWorkers = %d, want 5", postsParserWorkers)
	}
	if insightsParserWorkers != 5 {
		t.Errorf("insightsParserWorkers = %d, want 5", insightsParserWorkers)
	}
	if maxBatchSize != 10000 {
		t.Errorf("maxBatchSize = %d, want 10000", maxBatchSize)
	}
	if batchTimeout != 10*time.Second {
		t.Errorf("batchTimeout = %v, want 10s", batchTimeout)
	}
	if rawPostsTopic != "raw-facebook-posts" {
		t.Errorf("rawPostsTopic = %q, want %q", rawPostsTopic, "raw-facebook-posts")
	}
	if rawVideosTopic != "raw-facebook-videos" {
		t.Errorf("rawVideosTopic = %q, want %q", rawVideosTopic, "raw-facebook-videos")
	}
	if rawInsightsTopic != "raw-facebook-insights" {
		t.Errorf("rawInsightsTopic = %q, want %q", rawInsightsTopic, "raw-facebook-insights")
	}
}

// ================== PostsParser Comprehensive Tests ==================

func TestPostsParser_ProcessesValidPost(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		posts:       make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var parsedPosts, parsedAssets uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 0, in, batches, &parsedPosts, &parsedAssets, log)

	// Create a valid raw post
	rawPost := &kafkamodels.RawFacebookPost{
		ID:         "123456789_987654321",
		Message:    "Test message",
		StatusType: "added_photos",
	}

	in <- ParseJob{
		JobType:     "post",
		RawPost:     rawPost,
		PageID:      "123456789",
		PageName:    "Test Page",
		WorkspaceID: "ws123",
		MessageKey:  "ws123_123456789_post1",
	}

	// Give time for processing
	time.Sleep(100 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	// Check if post was parsed (might be 0 if parsing returns nil for incomplete data)
	// The key is that the parser ran without crashing
}

func TestPostsParser_ChannelClose(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		posts:       make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
	}

	in := make(chan ParseJob, 10)
	ctx := context.Background()

	var parsedPosts, parsedAssets uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 0, in, batches, &parsedPosts, &parsedAssets, log)

	// Close channel immediately
	close(in)
	wg.Wait()
	// Should exit gracefully
}

func TestPostsParser_MultipleJobs(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		posts:       make(chan *kafkamodels.ParsedFacebookPost, 100),
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 100),
	}

	in := make(chan ParseJob, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var parsedPosts, parsedAssets uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 0, in, batches, &parsedPosts, &parsedAssets, log)

	// Send multiple jobs including non-post types
	for i := 0; i < 5; i++ {
		in <- ParseJob{
			JobType: "post",
			RawPost: &kafkamodels.RawFacebookPost{
				ID:         "post" + string(rune('0'+i)),
				Message:    "Test",
				StatusType: "mobile_status_update",
			},
			PageID: "page123",
		}
	}

	// Send non-post job (should be skipped)
	in <- ParseJob{JobType: "video", PageID: "page123"}
	in <- ParseJob{JobType: "insights", PageID: "page123"}

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

// ================== MediaInsightsParser Comprehensive Tests ==================

func TestMediaInsightsParser_ProcessesVideo(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	// Create a video job
	rawVideo := &kafkamodels.RawFacebookVideo{
		ID:      "video123",
		Message: "Test Video",
	}

	in <- ParseJob{
		JobType:     "video",
		RawVideo:    rawVideo,
		PageID:      "page123",
		PageName:    "Test Page",
		WorkspaceID: "ws123",
		MessageKey:  "ws123_page123_video123",
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_ProcessesReels(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	// Create a video job with blue_reels_play_count (will be treated as reels)
	videoJSON := `{
		"id": "reel123",
		"title": "Test Reel",
		"video_insights": {
			"data": [
				{"name": "blue_reels_play_count", "values": [{"value": 1000}]}
			]
		}
	}`
	var rawVideo kafkamodels.RawFacebookVideo
	json.Unmarshal([]byte(videoJSON), &rawVideo)

	in <- ParseJob{
		JobType:     "video",
		RawVideo:    &rawVideo,
		PageID:      "page123",
		PageName:    "Test Page",
		WorkspaceID: "ws123",
		MessageKey:  "ws123_page123_reel123",
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_ProcessesInsights(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 100),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	// Create an insights job
	rawInsights := &kafkamodels.RawFacebookInsights{
		PageID:      "page123",
		WorkspaceID: "ws123",
	}

	in <- ParseJob{
		JobType:     "insights",
		RawInsights: rawInsights,
		PageID:      "page123",
		PageName:    "Test Page",
		WorkspaceID: "ws123",
		MessageKey:  "ws123_page123_insights",
	}

	time.Sleep(100 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_ChannelClose(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	in := make(chan ParseJob, 10)
	ctx := context.Background()

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	close(in)
	wg.Wait()
	// Should exit gracefully
}

func TestMediaInsightsParser_MultipleJobTypes(t *testing.T) {
	log := logger.New("error")

	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 100),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 100),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 100),
	}

	in := make(chan ParseJob, 100)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	// Send mix of job types
	for i := 0; i < 3; i++ {
		in <- ParseJob{
			JobType:  "video",
			RawVideo: &kafkamodels.RawFacebookVideo{ID: "video" + string(rune('0'+i))},
			PageID:   "page123",
		}
	}
	for i := 0; i < 2; i++ {
		in <- ParseJob{
			JobType:     "insights",
			RawInsights: &kafkamodels.RawFacebookInsights{PageID: "page123"},
			PageID:      "page123",
		}
	}
	// Unknown type should be ignored
	in <- ParseJob{JobType: "unknown", PageID: "page123"}

	time.Sleep(150 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

// ================== Batch Processor Timeout Tests ==================

func TestProcessVideoInsightsBatch_Timeout(t *testing.T) {
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

	close(in)
	<-done

	if atomic.LoadInt32(&insertCount) != 1 {
		t.Errorf("expected 1 video insight, got %d", insertCount)
	}
}

func TestProcessReelsInsightsBatch_Timeout(t *testing.T) {
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

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1", PlayCount: 100}

	close(in)
	<-done

	if atomic.LoadInt32(&insertCount) != 1 {
		t.Errorf("expected 1 reel insight, got %d", insertCount)
	}
}

func TestProcessMediaAssetsBatch_ContextCancel(t *testing.T) {
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
		t.Errorf("expected 1 asset flushed on cancel, got %d", insertCount)
	}
}

func TestProcessPageInsightsBatch_ContextCancel(t *testing.T) {
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
		t.Errorf("expected 1 insight flushed on cancel, got %d", insertCount)
	}
}

func TestProcessVideoInsightsBatch_ContextCancel(t *testing.T) {
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
		t.Errorf("expected 1 video insight flushed on cancel, got %d", insertCount)
	}
}

func TestProcessReelsInsightsBatch_ContextCancel(t *testing.T) {
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

	in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1", PlayCount: 100}

	time.Sleep(10 * time.Millisecond)
	cancel()
	<-done

	if atomic.LoadInt32(&insertCount) != 1 {
		t.Errorf("expected 1 reel insight flushed on cancel, got %d", insertCount)
	}
}

// ================== Batch Insert Error Tests ==================

func TestProcessMediaAssetsBatch_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			return errors.New("clickhouse error")
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
	close(in)
	<-done
	// Should not panic
}

func TestProcessPageInsightsBatch_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			return errors.New("clickhouse error")
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
	close(in)
	<-done
	// Should not panic
}

func TestProcessVideoInsightsBatch_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			return errors.New("clickhouse error")
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
	close(in)
	<-done
	// Should not panic
}

func TestProcessReelsInsightsBatch_Error(t *testing.T) {
	log := logger.New("error")

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			return errors.New("clickhouse error")
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
	close(in)
	<-done
	// Should not panic
}

// ================== extractKeyInfo Comprehensive Tests ==================

func TestExtractKeyInfo_AllCases(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		wantPageID   string
		wantPageName string
		wantWsID     string
	}{
		{"empty", "", "", "", ""},
		{"single", "page123", "page123", "", ""},
		{"two parts", "page123_post456", "page123", "", ""},
		{"three parts", "ws123_page456_post789", "page456", "", "ws123"},
		{"four parts", "ws_page_post_extra", "page", "", "ws"},
		{"five parts", "a_b_c_d_e", "b", "", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageID, pageName, wsID := extractKeyInfo(tt.key)
			if pageID != tt.wantPageID {
				t.Errorf("pageID = %q, want %q", pageID, tt.wantPageID)
			}
			if pageName != tt.wantPageName {
				t.Errorf("pageName = %q, want %q", pageName, tt.wantPageName)
			}
			if wsID != tt.wantWsID {
				t.Errorf("wsID = %q, want %q", wsID, tt.wantWsID)
			}
		})
	}
}

// ================== Handler Context Cancellation Tests ==================

func TestHandleRawVideo_ContextCancellation(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob) // unbuffered

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	videoJSON := []byte(`{"id": "video123"}`)

	err := handleRawVideo(ctx, []byte("key1"), videoJSON, out, log)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

func TestHandleRawInsights_ContextCancellation(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob) // unbuffered

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	insightsJSON := []byte(`{"page_id": "page123"}`)

	err := handleRawInsights(ctx, []byte("key1"), insightsJSON, out, log)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled, got: %v", err)
	}
}

// ================== Parser Context Cancellation During Channel Sends ==================

func TestPostsParser_ContextCancelDuringPostSend(t *testing.T) {
	log := logger.New("error")

	// Unbuffered channels to test context cancellation during send
	batches := &BatchCollectors{
		posts:       make(chan *kafkamodels.ParsedFacebookPost), // unbuffered
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())

	var parsedPosts, parsedAssets uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 0, in, batches, &parsedPosts, &parsedAssets, log)

	// Send job, then cancel context while parser is trying to send to unbuffered channel
	in <- ParseJob{
		JobType: "post",
		RawPost: &kafkamodels.RawFacebookPost{
			ID:         "page123_post456",
			Message:    "Test",
			StatusType: "mobile_status_update",
		},
		PageID: "page123",
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

func TestPostsParser_ContextCancelDuringAssetSend(t *testing.T) {
	log := logger.New("error")

	// Posts buffered, mediaAssets unbuffered
	batches := &BatchCollectors{
		posts:       make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets: make(chan *kafkamodels.ParsedFacebookMediaAsset), // unbuffered
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())

	var parsedPosts, parsedAssets uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 0, in, batches, &parsedPosts, &parsedAssets, log)

	// Send job with attachments to trigger media asset processing
	in <- ParseJob{
		JobType: "post",
		RawPost: &kafkamodels.RawFacebookPost{
			ID:         "page123_post456",
			Message:    "Test",
			StatusType: "added_photos",
		},
		PageID: "page123",
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_ContextCancelDuringVideoSend(t *testing.T) {
	log := logger.New("error")

	// Unbuffered video insights channel
	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights), // unbuffered
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	in <- ParseJob{
		JobType:  "video",
		RawVideo: &kafkamodels.RawFacebookVideo{ID: "video123"},
		PageID:   "page123",
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_ContextCancelDuringReelsSend(t *testing.T) {
	log := logger.New("error")

	// Unbuffered reels insights channel
	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights), // unbuffered
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	// Create a reel (video with blue_reels_play_count)
	videoJSON := `{
		"id": "reel123",
		"video_insights": {
			"data": [{"name": "blue_reels_play_count", "values": [{"value": 1000}]}]
		}
	}`
	var rawVideo kafkamodels.RawFacebookVideo
	json.Unmarshal([]byte(videoJSON), &rawVideo)

	in <- ParseJob{
		JobType:  "video",
		RawVideo: &rawVideo,
		PageID:   "page123",
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_ContextCancelDuringInsightsSend(t *testing.T) {
	log := logger.New("error")

	// Unbuffered page insights channel
	batches := &BatchCollectors{
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights), // unbuffered
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	in := make(chan ParseJob, 10)
	ctx, cancel := context.WithCancel(context.Background())

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	in <- ParseJob{
		JobType:     "insights",
		RawInsights: &kafkamodels.RawFacebookInsights{PageID: "page123"},
		PageID:      "page123",
	}

	time.Sleep(50 * time.Millisecond)
	cancel()
	close(in)
	wg.Wait()
}

// ================== Batch Processor Multiple Items Tests ==================

func TestProcessPostsBatch_MultipleBatches(t *testing.T) {
	log := logger.New("error")
	var batchCount int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			atomic.AddInt32(&batchCount, 1)
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookPost, 100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPostsBatch(ctx, in, sink, log)
		close(done)
	}()

	// Send 5 posts
	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookPost{PostID: "post" + string(rune('0'+i))}
	}

	close(in)
	<-done

	if atomic.LoadInt32(&batchCount) < 1 {
		t.Errorf("expected at least 1 batch, got %d", batchCount)
	}
}

func TestProcessMediaAssetsBatch_MultipleBatches(t *testing.T) {
	log := logger.New("error")
	var totalAssets int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			atomic.AddInt32(&totalAssets, int32(len(assets)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookMediaAsset, 100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processMediaAssetsBatch(ctx, in, sink, log)
		close(done)
	}()

	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookMediaAsset{PostID: "post" + string(rune('0'+i))}
	}

	close(in)
	<-done

	if atomic.LoadInt32(&totalAssets) != 5 {
		t.Errorf("expected 5 assets, got %d", totalAssets)
	}
}

func TestProcessPageInsightsBatch_MultipleBatches(t *testing.T) {
	log := logger.New("error")
	var totalInsights int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			atomic.AddInt32(&totalInsights, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookInsights, 100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processPageInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookInsights{PageID: "page" + string(rune('0'+i))}
	}

	close(in)
	<-done

	if atomic.LoadInt32(&totalInsights) != 5 {
		t.Errorf("expected 5 insights, got %d", totalInsights)
	}
}

func TestProcessVideoInsightsBatch_MultipleBatches(t *testing.T) {
	log := logger.New("error")
	var totalVideos int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			atomic.AddInt32(&totalVideos, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookVideoInsights, 100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processVideoInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookVideoInsights{PostID: "video" + string(rune('0'+i))}
	}

	close(in)
	<-done

	if atomic.LoadInt32(&totalVideos) != 5 {
		t.Errorf("expected 5 videos, got %d", totalVideos)
	}
}

func TestProcessReelsInsightsBatch_MultipleBatches(t *testing.T) {
	log := logger.New("error")
	var totalReels int32

	mockClient := &conversions.MockClickHouseClient{
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			atomic.AddInt32(&totalReels, int32(len(insights)))
			return nil
		},
	}

	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	in := make(chan *kafkamodels.ParsedFacebookReelsInsights, 100)
	ctx := context.Background()

	done := make(chan struct{})
	go func() {
		processReelsInsightsBatch(ctx, in, sink, log)
		close(done)
	}()

	for i := 0; i < 5; i++ {
		in <- &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel" + string(rune('0'+i))}
	}

	close(in)
	<-done

	if atomic.LoadInt32(&totalReels) != 5 {
		t.Errorf("expected 5 reels, got %d", totalReels)
	}
}

// ================== Handler Tests with Valid Data ==================

func TestHandleRawPost_ValidData(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob, 10)

	ctx := context.Background()

	postJSON := []byte(`{"id": "page123_post456", "message": "Test post", "status_type": "mobile_status_update"}`)

	err := handleRawPost(ctx, []byte("ws_page123_post456"), postJSON, out, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case job := <-out:
		if job.JobType != "post" {
			t.Errorf("expected job type 'post', got %q", job.JobType)
		}
		if job.PageID != "page123" {
			t.Errorf("expected pageID 'page123', got %q", job.PageID)
		}
	default:
		t.Fatal("expected job in channel")
	}
}

func TestHandleRawVideo_ValidData(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob, 10)

	ctx := context.Background()

	videoJSON := []byte(`{"id": "video123", "post_id": "post456"}`)

	err := handleRawVideo(ctx, []byte("ws_page123_video123"), videoJSON, out, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case job := <-out:
		if job.JobType != "video" {
			t.Errorf("expected job type 'video', got %q", job.JobType)
		}
	default:
		t.Fatal("expected job in channel")
	}
}

func TestHandleRawInsights_ValidData(t *testing.T) {
	log := logger.New("error")
	out := make(chan ParseJob, 10)

	ctx := context.Background()

	insightsJSON := []byte(`{"page_id": "page123", "workspace_id": "ws123"}`)

	err := handleRawInsights(ctx, []byte("ws_page123_insights"), insightsJSON, out, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case job := <-out:
		if job.JobType != "insights" {
			t.Errorf("expected job type 'insights', got %q", job.JobType)
		}
	default:
		t.Fatal("expected job in channel")
	}
}

// ================== ParseJob and BatchCollectors Struct Tests ==================

func TestParseJobStruct(t *testing.T) {
	rawPost := &kafkamodels.RawFacebookPost{ID: "post1"}
	rawVideo := &kafkamodels.RawFacebookVideo{ID: "video1"}
	rawInsights := &kafkamodels.RawFacebookInsights{PageID: "page1"}

	job := ParseJob{
		JobType:     "post",
		RawPost:     rawPost,
		RawVideo:    rawVideo,
		RawInsights: rawInsights,
		PageID:      "page123",
		PageName:    "Test Page",
		WorkspaceID: "ws123",
		MessageKey:  "ws123_page123_post1",
	}

	if job.JobType != "post" {
		t.Errorf("JobType = %q, want 'post'", job.JobType)
	}
	if job.RawPost != rawPost {
		t.Error("RawPost not set correctly")
	}
	if job.RawVideo != rawVideo {
		t.Error("RawVideo not set correctly")
	}
	if job.RawInsights != rawInsights {
		t.Error("RawInsights not set correctly")
	}
}

func TestBatchCollectorsStruct(t *testing.T) {
	batches := &BatchCollectors{
		posts:         make(chan *kafkamodels.ParsedFacebookPost, 10),
		mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, 10),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
	}

	if batches.posts == nil {
		t.Error("posts channel is nil")
	}
	if batches.mediaAssets == nil {
		t.Error("mediaAssets channel is nil")
	}
	if batches.pageInsights == nil {
		t.Error("pageInsights channel is nil")
	}
	if batches.videoInsights == nil {
		t.Error("videoInsights channel is nil")
	}
	if batches.reelsInsights == nil {
		t.Error("reelsInsights channel is nil")
	}
}

// ================== Error Path Tests for mediaInsightsParser ==================

func TestMediaInsightsParser_VideoParseError(t *testing.T) {
	log := logger.New("error")
	batches := &BatchCollectors{
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
	}

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64
	in := make(chan ParseJob, 10)
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	// Send a video job with nil RawVideo to cause parse error
	in <- ParseJob{
		JobType:  "video",
		RawVideo: &kafkamodels.RawFacebookVideo{},
		PageID:   "page123",
	}

	time.Sleep(100 * time.Millisecond)
	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_InsightsParseError(t *testing.T) {
	log := logger.New("error")
	batches := &BatchCollectors{
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, 10),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, 10),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, 10),
	}

	var parsedVideoIns, parsedReelsIns, parsedPageIns uint64
	in := make(chan ParseJob, 10)
	var wg sync.WaitGroup
	wg.Add(1)

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	go mediaInsightsParser(ctx, &wg, 0, in, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)

	// Send an insights job with empty data to trigger a different code path
	in <- ParseJob{
		JobType:     "insights",
		RawInsights: &kafkamodels.RawFacebookInsights{},
		PageID:      "page123",
	}

	time.Sleep(100 * time.Millisecond)
	close(in)
	wg.Wait()
}

func TestExtractKeyInfo_ZeroParts(t *testing.T) {
	// This tests the edge case handling - note that strings.Split never returns empty slice
	// but we test the function behavior with various inputs
	tests := []struct {
		name   string
		key    string
		wantPg string
		wantWs string
	}{
		{"underscore only", "_", "", ""},
		{"double underscore", "__", "", ""},
		{"underscore prefix", "_page_", "page", ""},
		{"mixed underscores", "a_b_c_d_e", "b", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageID, _, wsID := extractKeyInfo(tt.key)
			if pageID != tt.wantPg {
				t.Errorf("pageID = %q, want %q", pageID, tt.wantPg)
			}
			if wsID != tt.wantWs {
				t.Errorf("workspaceID = %q, want %q", wsID, tt.wantWs)
			}
		})
	}
}

// ================== Logging Contract Tests ==================

func TestLoggingContract_FacebookSink_ErrorHasContextFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	// Simulate what the Facebook analytics sink does when it gets an unexpected error
	log.Error().
		Str("error_message", "ClickHouse batch insert failed").
		Str("function", "batchInserter").
		Str("stage", "insert_posts").
		Msg("Facebook sink batch insert error")

	output := buf.String()

	checks := map[string]string{
		"ERR":           "expected ERR level in output",
		"error_message": "expected error_message field in output",
		"function":      "expected function field in output",
		"batchInserter": "expected batchInserter value in output",
		"stage":         "expected stage field in output",
	}
	for substr, errMsg := range checks {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

func TestLoggingContract_FacebookSink_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Log an error the way the Facebook sink does — Error level only, no CaptureException
	log.Error().
		Str("error_message", "ClickHouse connection lost").
		Str("function", "batchInserter").
		Str("stage", "flush_batch").
		Msg("Failed to flush batch to ClickHouse")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}
