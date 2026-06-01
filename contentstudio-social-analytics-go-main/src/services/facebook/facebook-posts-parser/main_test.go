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

// ================== extractKeyInfo Tests ==================

func TestExtractKeyInfo(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		wantPageID string
		wantName   string
		wantWSID   string
	}{
		{"empty key", "", "", "", ""},
		{"single part", "page123", "page123", "", ""},
		{"two parts", "page123_extra", "page123", "", ""},
		{"three parts", "ws789_page123_extra", "page123", "", "ws789"},
		{"four parts", "ws789_page123_extra_more", "page123", "", "ws789"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageID, pageName, wsID := extractKeyInfo(tt.key)
			if pageID != tt.wantPageID {
				t.Errorf("pageID = %q, want %q", pageID, tt.wantPageID)
			}
			if pageName != tt.wantName {
				t.Errorf("pageName = %q, want %q", pageName, tt.wantName)
			}
			if wsID != tt.wantWSID {
				t.Errorf("workspaceID = %q, want %q", wsID, tt.wantWSID)
			}
		})
	}
}

// ================== postsParser Tests ==================

func TestPostsParser_ContextCancel(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 1, in, out, log)

	cancel()
	wg.Wait()
}

func TestPostsParser_ChannelClose(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 1, in, out, log)

	close(in)
	wg.Wait()
}

func TestPostsParser_SkipsNonPostJobs(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 1, in, out, log)

	// Send non-post job
	in <- ParseJob{JobType: "video"}
	in <- ParseJob{JobType: "insights"}

	time.Sleep(50 * time.Millisecond)

	// Should not produce any output
	select {
	case <-out:
		t.Error("should not produce output for non-post jobs")
	default:
		// Expected
	}

	close(in)
	wg.Wait()
}

func TestPostsParser_ProcessesPostJob(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 100)

	var wg sync.WaitGroup
	wg.Add(1)
	go postsParser(ctx, &wg, 1, in, out, log)

	// Send post job
	rawPost := &kafkamodels.RawFacebookPost{
		ID:      "post123",
		Message: "Test post",
	}
	in <- ParseJob{
		JobType:     "post",
		RawPost:     rawPost,
		PageID:      "page456",
		WorkspaceID: "ws789",
	}

	time.Sleep(100 * time.Millisecond)

	close(in)
	wg.Wait()
	close(out)

	// May or may not produce output depending on parsing success
	// The test validates the parser doesn't crash
}

// ================== mediaInsightsParser Tests ==================

func TestMediaInsightsParser_ContextCancel(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 1, in, out, log)

	cancel()
	wg.Wait()
}

func TestMediaInsightsParser_ChannelClose(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 1, in, out, log)

	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_SkipsPostJobs(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 1, in, out, log)

	// Send post job (should be skipped)
	in <- ParseJob{JobType: "post"}

	time.Sleep(50 * time.Millisecond)

	select {
	case <-out:
		t.Error("should not produce output for post jobs")
	default:
		// Expected
	}

	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_ProcessesVideoJob(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 100)

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 1, in, out, log)

	// Send video job
	rawVideo := &kafkamodels.RawFacebookVideo{
		ID:      "video123",
		PostID:  "post456",
		Message: "Test video",
	}
	in <- ParseJob{
		JobType:     "video",
		RawVideo:    rawVideo,
		PageID:      "page789",
		WorkspaceID: "ws123",
	}

	time.Sleep(100 * time.Millisecond)

	close(in)
	wg.Wait()
}

func TestMediaInsightsParser_ProcessesInsightsJob(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 100)

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaInsightsParser(ctx, &wg, 1, in, out, log)

	// Send insights job
	rawInsights := &kafkamodels.RawFacebookInsights{
		PageID: "page123",
	}
	in <- ParseJob{
		JobType:     "insights",
		RawInsights: rawInsights,
		PageID:      "page123",
		WorkspaceID: "ws456",
	}

	time.Sleep(100 * time.Millisecond)

	close(in)
	wg.Wait()
}

// ================== publisher Tests ==================

func TestPublisher_ContextCancel(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())

	in := make(chan PublishJob, 10)
	producer := &kafka.MockProducer{}

	var pubPosts, pubAssets, pubVid, pubReels, pubIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go publisher(ctx, &wg, 1, "test", in, producer, &pubPosts, &pubAssets, &pubVid, &pubReels, &pubIns, log)

	cancel()
	wg.Wait()
}

func TestPublisher_ChannelClose(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()

	in := make(chan PublishJob, 10)
	producer := &kafka.MockProducer{}

	var pubPosts, pubAssets, pubVid, pubReels, pubIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go publisher(ctx, &wg, 1, "test", in, producer, &pubPosts, &pubAssets, &pubVid, &pubReels, &pubIns, log)

	close(in)
	wg.Wait()
}

func TestPublisher_PublishesPosts(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan PublishJob, 10)

	var producedCount int32
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&producedCount, 1)
			return nil
		},
	}

	var pubPosts, pubAssets, pubVid, pubReels, pubIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go publisher(ctx, &wg, 1, "test", in, producer, &pubPosts, &pubAssets, &pubVid, &pubReels, &pubIns, log)

	// Send post job
	in <- PublishJob{
		Topic: parsedPostsTopic,
		Key:   "page123_post456",
		Data:  &kafkamodels.ParsedFacebookPost{PostID: "post456"},
	}

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadInt32(&producedCount) != 1 {
		t.Errorf("expected 1 produce call, got %d", producedCount)
	}
	if atomic.LoadUint64(&pubPosts) != 1 {
		t.Errorf("expected pubPosts=1, got %d", pubPosts)
	}

	close(in)
	wg.Wait()
}

func TestPublisher_PublishesMediaAssets(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan PublishJob, 10)

	var producedCount int32
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&producedCount, 1)
			return nil
		},
	}

	var pubPosts, pubAssets, pubVid, pubReels, pubIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go publisher(ctx, &wg, 1, "test", in, producer, &pubPosts, &pubAssets, &pubVid, &pubReels, &pubIns, log)

	in <- PublishJob{
		Topic: parsedMediaAssetsTopic,
		Key:   "page123_media456",
		Data:  kafkamodels.ParsedFacebookMediaAsset{MediaID: "media456"},
	}

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadUint64(&pubAssets) != 1 {
		t.Errorf("expected pubAssets=1, got %d", pubAssets)
	}

	close(in)
	wg.Wait()
}

func TestPublisher_PublishesVideoInsights(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan PublishJob, 10)

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	var pubPosts, pubAssets, pubVid, pubReels, pubIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go publisher(ctx, &wg, 1, "test", in, producer, &pubPosts, &pubAssets, &pubVid, &pubReels, &pubIns, log)

	in <- PublishJob{
		Topic: parsedVideoInsightsTopic,
		Key:   "page123_video456",
		Data:  &kafkamodels.ParsedFacebookVideoInsights{VideoID: "video456"},
	}

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadUint64(&pubVid) != 1 {
		t.Errorf("expected pubVid=1, got %d", pubVid)
	}

	close(in)
	wg.Wait()
}

func TestPublisher_PublishesReelsInsights(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan PublishJob, 10)

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	var pubPosts, pubAssets, pubVid, pubReels, pubIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go publisher(ctx, &wg, 1, "test", in, producer, &pubPosts, &pubAssets, &pubVid, &pubReels, &pubIns, log)

	in <- PublishJob{
		Topic: parsedReelsInsightsTopic,
		Key:   "page123_reel456",
		Data:  &kafkamodels.ParsedFacebookReelsInsights{PostID: "reel456"},
	}

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadUint64(&pubReels) != 1 {
		t.Errorf("expected pubReels=1, got %d", pubReels)
	}

	close(in)
	wg.Wait()
}

func TestPublisher_PublishesPageInsights(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan PublishJob, 10)

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	var pubPosts, pubAssets, pubVid, pubReels, pubIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go publisher(ctx, &wg, 1, "test", in, producer, &pubPosts, &pubAssets, &pubVid, &pubReels, &pubIns, log)

	in <- PublishJob{
		Topic: parsedInsightsTopic,
		Key:   "page123",
		Data:  []*kafkamodels.ParsedFacebookInsights{{PageID: "page123"}},
	}

	time.Sleep(50 * time.Millisecond)

	if atomic.LoadUint64(&pubIns) != 1 {
		t.Errorf("expected pubIns=1, got %d", pubIns)
	}

	close(in)
	wg.Wait()
}

func TestPublisher_ProduceError(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	in := make(chan PublishJob, 10)

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return errors.New("kafka error")
		},
	}

	var pubPosts, pubAssets, pubVid, pubReels, pubIns uint64

	var wg sync.WaitGroup
	wg.Add(1)
	go publisher(ctx, &wg, 1, "test", in, producer, &pubPosts, &pubAssets, &pubVid, &pubReels, &pubIns, log)

	in <- PublishJob{
		Topic: parsedPostsTopic,
		Key:   "page123_post456",
		Data:  &kafkamodels.ParsedFacebookPost{PostID: "post456"},
	}

	time.Sleep(50 * time.Millisecond)

	// Should not increment counter on error
	if atomic.LoadUint64(&pubPosts) != 0 {
		t.Errorf("expected pubPosts=0 on error, got %d", pubPosts)
	}

	close(in)
	wg.Wait()
}

// ================== parseRawFacebookPostOrVideoOrInsights Tests ==================

func TestParseRawFacebookPostOrVideoOrInsights_UnknownJobType(t *testing.T) {
	job := ParseJob{JobType: "unknown"}
	result := parseRawFacebookPostOrVideoOrInsights(job)
	if result.Error == nil {
		t.Error("expected error for unknown job type")
	}
	if result.Error.Error() != "parseRawFacebookPostOrVideoOrInsights: unknown job type: unknown" {
		t.Errorf("unexpected error message: %v", result.Error)
	}
}

func TestParseRawFacebookPostOrVideoOrInsights_PostJob(t *testing.T) {
	rawPost := kafkamodels.RawFacebookPost{
		ID:      "post123",
		Message: "Test post",
	}
	job := ParseJob{
		JobType: "post",
		RawPost: &rawPost,
		PageID:  "page456",
	}
	result := parseRawFacebookPostOrVideoOrInsights(job)
	// Result depends on parsing library - may succeed or fail
	// Test validates no panic
	_ = result
}

func TestParseRawFacebookPostOrVideoOrInsights_VideoJob(t *testing.T) {
	rawVideo := kafkamodels.RawFacebookVideo{
		ID:      "video123",
		PostID:  "post456",
		Message: "Test video",
	}
	job := ParseJob{
		JobType:  "video",
		RawVideo: &rawVideo,
		PageID:   "page789",
		PageName: "Test Page",
	}
	result := parseRawFacebookPostOrVideoOrInsights(job)
	// May succeed or fail depending on video data
	_ = result
}

func TestParseRawFacebookPostOrVideoOrInsights_InsightsJob(t *testing.T) {
	rawInsights := kafkamodels.RawFacebookInsights{
		PageID: "page123",
	}
	job := ParseJob{
		JobType:     "insights",
		RawInsights: &rawInsights,
		PageID:      "page123",
		WorkspaceID: "ws456",
	}
	result := parseRawFacebookPostOrVideoOrInsights(job)
	// May succeed or fail depending on insights data
	_ = result
}

// ================== ParseJob and ParseResult Struct Tests ==================

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

func TestParseResult_Struct(t *testing.T) {
	result := ParseResult{
		ParsedPost:    &kafkamodels.ParsedFacebookPost{PostID: "post123"},
		MediaAssets:   []kafkamodels.ParsedFacebookMediaAsset{{MediaID: "m1"}},
		VideoInsights: &kafkamodels.ParsedFacebookVideoInsights{VideoID: "v1"},
		ReelsInsights: &kafkamodels.ParsedFacebookReelsInsights{PostID: "r1"},
		Insights:      []*kafkamodels.ParsedFacebookInsights{{PageID: "p1"}},
	}

	if result.ParsedPost.PostID != "post123" {
		t.Errorf("PostID = %q, want %q", result.ParsedPost.PostID, "post123")
	}
	if len(result.MediaAssets) != 1 {
		t.Errorf("MediaAssets length = %d, want 1", len(result.MediaAssets))
	}
}

func TestPublishJob_Struct(t *testing.T) {
	job := PublishJob{
		Topic: "test-topic",
		Key:   "key123",
		Data:  map[string]string{"test": "data"},
	}

	if job.Topic != "test-topic" {
		t.Errorf("Topic = %q, want %q", job.Topic, "test-topic")
	}
	if job.Key != "key123" {
		t.Errorf("Key = %q, want %q", job.Key, "key123")
	}
}

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if rawPostsTopic != "raw-facebook-posts" {
		t.Errorf("rawPostsTopic = %q, want %q", rawPostsTopic, "raw-facebook-posts")
	}
	if rawVideosTopic != "raw-facebook-videos" {
		t.Errorf("rawVideosTopic = %q, want %q", rawVideosTopic, "raw-facebook-videos")
	}
	if rawInsightsTopic != "raw-facebook-insights" {
		t.Errorf("rawInsightsTopic = %q, want %q", rawInsightsTopic, "raw-facebook-insights")
	}
	if parsedPostsTopic != "parsed-facebook-posts" {
		t.Errorf("parsedPostsTopic = %q, want %q", parsedPostsTopic, "parsed-facebook-posts")
	}
	if parsedMediaAssetsTopic != "parsed-facebook-media-assets" {
		t.Errorf("parsedMediaAssetsTopic = %q, want %q", parsedMediaAssetsTopic, "parsed-facebook-media-assets")
	}
	if parsedVideoInsightsTopic != "parsed-facebook-video-insights" {
		t.Errorf("parsedVideoInsightsTopic = %q, want %q", parsedVideoInsightsTopic, "parsed-facebook-video-insights")
	}
	if parsedReelsInsightsTopic != "parsed-facebook-reels-insights" {
		t.Errorf("parsedReelsInsightsTopic = %q, want %q", parsedReelsInsightsTopic, "parsed-facebook-reels-insights")
	}
	if parsedInsightsTopic != "parsed-facebook-insights" {
		t.Errorf("parsedInsightsTopic = %q, want %q", parsedInsightsTopic, "parsed-facebook-insights")
	}
}

// ================== Handler Tests ==================

func TestHandleRawPost_Success(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	parseJobs := make(chan ParseJob, 10)

	rawPost := kafkamodels.RawFacebookPost{
		ID: "post123",
	}
	postJSON, _ := json.Marshal(rawPost)

	err := handleRawPost(ctx, []byte("ws789_page456_extra"), postJSON, parseJobs, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case job := <-parseJobs:
		if job.JobType != "post" {
			t.Errorf("JobType = %q, want %q", job.JobType, "post")
		}
		if job.PageID != "page456" {
			t.Errorf("PageID = %q, want %q", job.PageID, "page456")
		}
	default:
		t.Fatal("expected job in channel")
	}
}

func TestHandleRawPost_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	parseJobs := make(chan ParseJob, 10)

	err := handleRawPost(ctx, []byte("key"), []byte("invalid json"), parseJobs, log)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleRawVideo_Success(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	parseJobs := make(chan ParseJob, 10)

	rawVideo := kafkamodels.RawFacebookVideo{
		ID:     "video123",
		PostID: "post456",
	}
	videoJSON, _ := json.Marshal(rawVideo)

	err := handleRawVideo(ctx, []byte("ws789_page456_extra"), videoJSON, parseJobs, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case job := <-parseJobs:
		if job.JobType != "video" {
			t.Errorf("JobType = %q, want %q", job.JobType, "video")
		}
	default:
		t.Fatal("expected job in channel")
	}
}

func TestHandleRawVideo_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	parseJobs := make(chan ParseJob, 10)

	err := handleRawVideo(ctx, []byte("key"), []byte("invalid json"), parseJobs, log)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestHandleRawInsights_Success(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	parseJobs := make(chan ParseJob, 10)

	rawInsights := kafkamodels.RawFacebookInsights{
		PageID: "page123",
	}
	insightsJSON, _ := json.Marshal(rawInsights)

	err := handleRawInsights(ctx, []byte("ws789_page456_extra"), insightsJSON, parseJobs, log)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	select {
	case job := <-parseJobs:
		if job.JobType != "insights" {
			t.Errorf("JobType = %q, want %q", job.JobType, "insights")
		}
	default:
		t.Fatal("expected job in channel")
	}
}

func TestHandleRawInsights_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	ctx := context.Background()
	parseJobs := make(chan ParseJob, 10)

	err := handleRawInsights(ctx, []byte("key"), []byte("invalid json"), parseJobs, log)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

// ================== Context Cancellation Handler Tests ==================

func TestHandleRawPost_ContextCancel(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parseJobs := make(chan ParseJob) // unbuffered

	rawPost := kafkamodels.RawFacebookPost{ID: "post123"}
	postJSON, _ := json.Marshal(rawPost)

	err := handleRawPost(ctx, []byte("key"), postJSON, parseJobs, log)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestHandleRawVideo_ContextCancel(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parseJobs := make(chan ParseJob) // unbuffered

	rawVideo := kafkamodels.RawFacebookVideo{ID: "video123"}
	videoJSON, _ := json.Marshal(rawVideo)

	err := handleRawVideo(ctx, []byte("key"), videoJSON, parseJobs, log)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

func TestHandleRawInsights_ContextCancel(t *testing.T) {
	log := logger.New("error")
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	parseJobs := make(chan ParseJob) // unbuffered

	rawInsights := kafkamodels.RawFacebookInsights{PageID: "page123"}
	insightsJSON, _ := json.Marshal(rawInsights)

	err := handleRawInsights(ctx, []byte("key"), insightsJSON, parseJobs, log)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}
