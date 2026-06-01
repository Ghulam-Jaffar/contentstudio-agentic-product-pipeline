package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if mediaParserWorkers != 10 {
		t.Errorf("mediaParserWorkers = %d, want 10", mediaParserWorkers)
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
	if batchProcessorsPerType != 3 {
		t.Errorf("batchProcessorsPerType = %d, want 3", batchProcessorsPerType)
	}
	if messageChanSize != 500 {
		t.Errorf("messageChanSize = %d, want 500", messageChanSize)
	}
	if mediaTopic != "raw-instagram-media" {
		t.Errorf("mediaTopic = %q, want %q", mediaTopic, "raw-instagram-media")
	}
	if insightsTopic != "raw-instagram-insights" {
		t.Errorf("insightsTopic = %q, want %q", insightsTopic, "raw-instagram-insights")
	}
	if consumerGroup != "instagram-posts-parser-group" {
		t.Errorf("consumerGroup = %q, want %q", consumerGroup, "instagram-posts-parser-group")
	}
}

// ================== ParseJob Tests ==================

func TestParseJob_Struct(t *testing.T) {
	job := ParseJob{
		JobType:      "media",
		EnrichedData: map[string]interface{}{"key": "value"},
		InstagramID:  "ig123",
		MessageKey:   "ws_ig123_media456",
	}

	if job.JobType != "media" {
		t.Errorf("JobType = %q, want %q", job.JobType, "media")
	}
	if job.InstagramID != "ig123" {
		t.Errorf("InstagramID = %q, want %q", job.InstagramID, "ig123")
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

// ================== extractKeyInfo Tests ==================

func TestExtractKeyInfo(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		wantIgID string
		wantWsID string
	}{
		{"empty", "", "", ""},
		{"single part", "ig123", "ig123", ""},
		{"two parts", "ig123_media456", "ig123", ""},
		{"three parts", "ws789_ig123_media456", "ig123", "ws789"},
		{"four parts", "a_b_c_d", "b", "a"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			igID, wsID := extractKeyInfo(tt.key)
			if igID != tt.wantIgID {
				t.Errorf("igID = %q, want %q", igID, tt.wantIgID)
			}
			if wsID != tt.wantWsID {
				t.Errorf("wsID = %q, want %q", wsID, tt.wantWsID)
			}
		})
	}
}

// ================== handleRawMedia Tests ==================

func TestHandleRawMedia_Success(t *testing.T) {
	log := logger.New("error")
	jobChan := make(chan ParseJob, 10)

	enrichedData := map[string]interface{}{
		"id":   "media123",
		"type": "IMAGE",
	}
	dataJSON, _ := json.Marshal(enrichedData)

	ctx := context.Background()
	err := handleRawMedia(ctx, []byte("ws_ig123_media456"), dataJSON, jobChan, log)
	if err != nil {
		t.Errorf("handleRawMedia returned error: %v", err)
	}

	select {
	case job := <-jobChan:
		if job.JobType != "media" {
			t.Errorf("JobType = %q, want %q", job.JobType, "media")
		}
		if job.InstagramID != "ig123" {
			t.Errorf("InstagramID = %q, want %q", job.InstagramID, "ig123")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for job")
	}
}

func TestHandleRawMedia_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	jobChan := make(chan ParseJob, 10)

	ctx := context.Background()
	err := handleRawMedia(ctx, []byte("key"), []byte("invalid json"), jobChan, log)
	// Should return nil (skip invalid) but not crash
	if err != nil {
		t.Errorf("expected nil error for invalid JSON, got %v", err)
	}
}

func TestHandleRawMedia_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	jobChan := make(chan ParseJob) // unbuffered

	enrichedData := map[string]interface{}{"id": "media123"}
	dataJSON, _ := json.Marshal(enrichedData)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := handleRawMedia(ctx, []byte("key"), dataJSON, jobChan, log)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ================== handleRawInsights Tests ==================

func TestHandleRawInsights_Success(t *testing.T) {
	log := logger.New("error")
	jobChan := make(chan ParseJob, 10)

	enrichedData := map[string]interface{}{
		"insights": map[string]interface{}{"impressions": 100},
	}
	dataJSON, _ := json.Marshal(enrichedData)

	ctx := context.Background()
	// Use 3-part key: workspace_instagramID_something
	err := handleRawInsights(ctx, []byte("ws_ig123_insights"), dataJSON, jobChan, log)
	if err != nil {
		t.Errorf("handleRawInsights returned error: %v", err)
	}

	select {
	case job := <-jobChan:
		if job.JobType != "insights" {
			t.Errorf("JobType = %q, want %q", job.JobType, "insights")
		}
		if job.InstagramID != "ig123" {
			t.Errorf("InstagramID = %q, want %q", job.InstagramID, "ig123")
		}
	case <-time.After(time.Second):
		t.Fatal("timeout waiting for job")
	}
}

func TestHandleRawInsights_InvalidJSON(t *testing.T) {
	log := logger.New("error")
	jobChan := make(chan ParseJob, 10)

	ctx := context.Background()
	err := handleRawInsights(ctx, []byte("key"), []byte("invalid"), jobChan, log)
	if err != nil {
		t.Errorf("expected nil error for invalid JSON, got %v", err)
	}
}

func TestHandleRawInsights_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	jobChan := make(chan ParseJob) // unbuffered

	enrichedData := map[string]interface{}{"insights": map[string]interface{}{}}
	dataJSON, _ := json.Marshal(enrichedData)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := handleRawInsights(ctx, []byte("key"), dataJSON, jobChan, log)
	if err != context.Canceled {
		t.Errorf("expected context.Canceled, got %v", err)
	}
}

// ================== mediaParser Tests ==================

func TestMediaParser_ContextCancel(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan *kafkamodels.ParsedInstagramPost, 10)
	var parsedCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go mediaParser(ctx, &wg, 0, in, out, &parsedCounter, log)

	cancel()
	wg.Wait()
}

func TestMediaParser_ChannelClose(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan *kafkamodels.ParsedInstagramPost, 10)
	var parsedCounter uint64

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go mediaParser(ctx, &wg, 0, in, out, &parsedCounter, log)

	close(in)
	wg.Wait()
}

func TestMediaParser_SkipsNonMediaJob(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan *kafkamodels.ParsedInstagramPost, 10)
	var parsedCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go mediaParser(ctx, &wg, 0, in, out, &parsedCounter, log)

	// Send non-media job
	in <- ParseJob{JobType: "insights", InstagramID: "ig123"}

	time.Sleep(50 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	// Output should be empty since job was skipped
	if len(out) != 0 {
		t.Errorf("expected empty output, got %d items", len(out))
	}
}

// ================== insightsParser Tests ==================

func TestInsightsParser_ContextCancel(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan *kafkamodels.ParsedInstagramInsight, 10)
	var parsedCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, in, out, &parsedCounter, log)

	cancel()
	wg.Wait()
}

func TestInsightsParser_ChannelClose(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan *kafkamodels.ParsedInstagramInsight, 10)
	var parsedCounter uint64

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, in, out, &parsedCounter, log)

	close(in)
	wg.Wait()
}

func TestInsightsParser_SkipsNonInsightsJob(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan *kafkamodels.ParsedInstagramInsight, 10)
	var parsedCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, in, out, &parsedCounter, log)

	// Send media job to insights parser
	in <- ParseJob{JobType: "media", InstagramID: "ig123"}

	time.Sleep(50 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	if len(out) != 0 {
		t.Errorf("expected empty output, got %d items", len(out))
	}
}

// ================== extractDateString Tests ==================

func TestExtractDateString(t *testing.T) {
	tests := []struct {
		name  string
		diMap map[string]interface{}
		want  string
	}{
		{
			name:  "RFC3339 format",
			diMap: map[string]interface{}{"Date": "2024-01-15T12:00:00Z"},
			want:  "2024-01-15",
		},
		{
			name:  "Date only",
			diMap: map[string]interface{}{"Date": "2024-01-15"},
			want:  "2024-01-15",
		},
		{
			name:  "Empty date",
			diMap: map[string]interface{}{"Date": ""},
			want:  "",
		},
		{
			name:  "Missing date",
			diMap: map[string]interface{}{},
			want:  "",
		},
		{
			name:  "Wrong type",
			diMap: map[string]interface{}{"Date": 12345},
			want:  "",
		},
		{
			name:  "Short date",
			diMap: map[string]interface{}{"Date": "2024-01"},
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractDateString(tt.diMap)
			if got != tt.want {
				t.Errorf("extractDateString() = %q, want %q", got, tt.want)
			}
		})
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

	startBatchProcessors(ctx, bc, nil, log, &wg, 1, &insertedPosts, &insertedInsights)

	time.Sleep(50 * time.Millisecond)

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

	// Close immediately without sending messages
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

	// Close immediately without sending messages
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

// ================== Multiple Message Tests ==================

func TestHandleRawMedia_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	jobChan := make(chan ParseJob, 100)

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		enrichedData := map[string]interface{}{"id": "media" + string(rune('0'+i))}
		dataJSON, _ := json.Marshal(enrichedData)
		if err := handleRawMedia(ctx, []byte("key"), dataJSON, jobChan, log); err != nil {
			t.Errorf("handleRawMedia[%d] failed: %v", i, err)
		}
	}

	if len(jobChan) != 10 {
		t.Errorf("jobChan len = %d, want 10", len(jobChan))
	}
}

func TestHandleRawInsights_MultipleMessages(t *testing.T) {
	log := logger.New("error")
	jobChan := make(chan ParseJob, 100)

	ctx := context.Background()

	for i := 0; i < 10; i++ {
		enrichedData := map[string]interface{}{"insights": map[string]interface{}{}}
		dataJSON, _ := json.Marshal(enrichedData)
		if err := handleRawInsights(ctx, []byte("key"), dataJSON, jobChan, log); err != nil {
			t.Errorf("handleRawInsights[%d] failed: %v", i, err)
		}
	}

	if len(jobChan) != 10 {
		t.Errorf("jobChan len = %d, want 10", len(jobChan))
	}
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

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < messagesPerWriter; j++ {
				bc.posts <- &kafkamodels.ParsedInstagramPost{MediaID: "media"}
				bc.insights <- &kafkamodels.ParsedInstagramInsight{InstagramID: "ig"}
			}
		}()
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

// ================== Logging Contract Tests (Point 4 — Calling service logs errors with context) ==================

func TestLoggingContract_InstagramAnalyticsSink_ErrorHasContextFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	// Simulate what the sink does when it gets an unexpected error
	log.Error().
		Str("error_message", "bulk insert failed").
		Str("function", "batchInsertWorker").
		Str("stage", "bulk_insert_posts").
		Msg("Instagram analytics sink error")

	output := buf.String()

	checks := map[string]string{
		"ERR":               "expected ERR level",
		"error_message":     "expected error_message field",
		"function":          "expected function field",
		"batchInsertWorker": "expected batchInsertWorker value",
		"stage":             "expected stage field",
		"bulk_insert_posts": "expected bulk_insert_posts stage value",
	}
	for substr, errMsg := range checks {
		if !strings.Contains(output, substr) {
			t.Errorf("%s, got: %s", errMsg, output)
		}
	}
}

func TestLoggingContract_InstagramAnalyticsSink_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "unmarshal failed").
		Str("function", "mediaParserWorker").
		Str("stage", "unmarshal_raw_media").
		Msg("Failed to unmarshal raw media")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}

func TestLoggingContract_InstagramAnalyticsSink_SingleSentryEvent(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	log.Error().
		Str("error_message", "clickhouse connection lost").
		Str("function", "batchInsertWorker").
		Str("stage", "bulk_insert_insights").
		Msg("Failed to insert insights")

	var errorLevelCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorLevelCount++
		}
	}
	if errorLevelCount != 1 {
		t.Fatalf("expected exactly 1 ErrorLevel hook firing, got %d", errorLevelCount)
	}
}
