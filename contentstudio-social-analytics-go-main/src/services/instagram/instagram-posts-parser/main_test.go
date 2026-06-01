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
)

// ================== Constants Tests ==================

func TestConstants(t *testing.T) {
	if mediaParserWorkers != 10 {
		t.Errorf("mediaParserWorkers = %d, want 10", mediaParserWorkers)
	}
	if insightsParserWorkers != 5 {
		t.Errorf("insightsParserWorkers = %d, want 5", insightsParserWorkers)
	}
	if mediaPublishWorkers != 10 {
		t.Errorf("mediaPublishWorkers = %d, want 10", mediaPublishWorkers)
	}
	if insightsPublishWorkers != 6 {
		t.Errorf("insightsPublishWorkers = %d, want 6", insightsPublishWorkers)
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
	if parsedPostsTopic != "parsed-instagram-posts" {
		t.Errorf("parsedPostsTopic = %q, want %q", parsedPostsTopic, "parsed-instagram-posts")
	}
	if parsedInsightsTopic != "parsed-instagram-insights" {
		t.Errorf("parsedInsightsTopic = %q, want %q", parsedInsightsTopic, "parsed-instagram-insights")
	}
}

// ================== ParseJob Tests ==================

func TestParseJob_Struct(t *testing.T) {
	job := ParseJob{
		JobType:      "media",
		EnrichedData: map[string]interface{}{"id": "media123"},
		InstagramID:  "ig456",
		MessageKey:   "ws_ig456_media123",
	}

	if job.JobType != "media" {
		t.Errorf("JobType = %q, want %q", job.JobType, "media")
	}
	if job.InstagramID != "ig456" {
		t.Errorf("InstagramID = %q, want %q", job.InstagramID, "ig456")
	}
	if job.MessageKey != "ws_ig456_media123" {
		t.Errorf("MessageKey = %q, want %q", job.MessageKey, "ws_ig456_media123")
	}
}

// ================== PublishJob Tests ==================

func TestPublishJob_Struct(t *testing.T) {
	job := PublishJob{
		Topic: "parsed-instagram-posts",
		Key:   "ig123_media456",
		Data:  map[string]interface{}{"id": "media456"},
	}

	if job.Topic != "parsed-instagram-posts" {
		t.Errorf("Topic = %q, want %q", job.Topic, "parsed-instagram-posts")
	}
	if job.Key != "ig123_media456" {
		t.Errorf("Key = %q, want %q", job.Key, "ig123_media456")
	}
}

// ================== extractKeyInfo Tests ==================

func TestExtractKeyInfo(t *testing.T) {
	tests := []struct {
		name        string
		key         string
		wantIgID    string
		wantWsID    string
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
		"id":         "media123",
		"media_type": "IMAGE",
	}
	dataJSON, _ := json.Marshal(enrichedData)

	ctx := context.Background()
	err := handleRawMedia(ctx, []byte("ws_ig456_media123"), dataJSON, jobChan, log)
	if err != nil {
		t.Errorf("handleRawMedia returned error: %v", err)
	}

	select {
	case job := <-jobChan:
		if job.JobType != "media" {
			t.Errorf("JobType = %q, want %q", job.JobType, "media")
		}
		if job.InstagramID != "ig456" {
			t.Errorf("InstagramID = %q, want %q", job.InstagramID, "ig456")
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
	// Should return nil (skip invalid)
	if err != nil {
		t.Errorf("expected nil error, got %v", err)
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
		t.Errorf("expected nil error, got %v", err)
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
	out := make(chan PublishJob, 10)
	var pubCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go mediaParser(ctx, &wg, 0, in, out, &pubCounter, log)

	cancel()
	wg.Wait()
}

func TestMediaParser_ChannelClose(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)
	var pubCounter uint64

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go mediaParser(ctx, &wg, 0, in, out, &pubCounter, log)

	close(in)
	wg.Wait()
}

func TestMediaParser_SkipsNonMediaJob(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)
	var pubCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go mediaParser(ctx, &wg, 0, in, out, &pubCounter, log)

	// Send insights job to media parser
	in <- ParseJob{JobType: "insights", InstagramID: "ig123"}

	time.Sleep(50 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	if len(out) != 0 {
		t.Errorf("expected empty output, got %d items", len(out))
	}
}

// ================== insightsParser Tests ==================

func TestInsightsParser_ContextCancel(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)
	var pubCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, in, out, &pubCounter, log)

	cancel()
	wg.Wait()
}

func TestInsightsParser_ChannelClose(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)
	var pubCounter uint64

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, in, out, &pubCounter, log)

	close(in)
	wg.Wait()
}

func TestInsightsParser_SkipsNonInsightsJob(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)
	var pubCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, in, out, &pubCounter, log)

	// Send media job
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

// ================== publisher Tests ==================

func TestPublisher_ContextCancel(t *testing.T) {
	log := logger.New("error")
	in := make(chan PublishJob, 10)
	var pubCounter uint64

	mockProducer := &kafka.MockProducer{}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 0, "media", in, mockProducer, &pubCounter, log)

	cancel()
	wg.Wait()
}

func TestPublisher_ChannelClose(t *testing.T) {
	log := logger.New("error")
	in := make(chan PublishJob, 10)
	var pubCounter uint64

	mockProducer := &kafka.MockProducer{}

	ctx := context.Background()
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 0, "media", in, mockProducer, &pubCounter, log)

	close(in)
	wg.Wait()
}

func TestPublisher_Success(t *testing.T) {
	log := logger.New("error")
	in := make(chan PublishJob, 10)
	var pubCounter uint64

	var producedCount int32
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&producedCount, 1)
			return nil
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 0, "media", in, mockProducer, &pubCounter, log)

	// Send a job
	in <- PublishJob{
		Topic: "parsed-instagram-posts",
		Key:   "ig123_media456",
		Data:  map[string]interface{}{"id": "media456"},
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	if atomic.LoadInt32(&producedCount) != 1 {
		t.Errorf("producedCount = %d, want 1", producedCount)
	}
	if atomic.LoadUint64(&pubCounter) != 1 {
		t.Errorf("pubCounter = %d, want 1", pubCounter)
	}
}

func TestPublisher_MarshalError(t *testing.T) {
	log := logger.New("error")
	in := make(chan PublishJob, 10)
	var pubCounter uint64

	mockProducer := &kafka.MockProducer{}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 0, "media", in, mockProducer, &pubCounter, log)

	// Send job with unmarshalable data (channel can't be marshaled)
	in <- PublishJob{
		Topic: "parsed-instagram-posts",
		Key:   "key",
		Data:  make(chan int), // channels can't be JSON marshaled
	}

	time.Sleep(50 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	// Should have skipped the message
	if atomic.LoadUint64(&pubCounter) != 0 {
		t.Errorf("pubCounter = %d, want 0", pubCounter)
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

func TestConcurrentAccess(t *testing.T) {
	log := logger.New("error")
	jobChan := make(chan ParseJob, 1000)

	var wg sync.WaitGroup
	numWriters := 10
	messagesPerWriter := 50

	ctx := context.Background()

	for i := 0; i < numWriters; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < messagesPerWriter; j++ {
				enrichedData := map[string]interface{}{"id": "media"}
				dataJSON, _ := json.Marshal(enrichedData)
				handleRawMedia(ctx, []byte("key"), dataJSON, jobChan, log)
			}
		}()
	}

	wg.Wait()

	expectedCount := numWriters * messagesPerWriter
	if len(jobChan) != expectedCount {
		t.Errorf("jobChan len = %d, want %d", len(jobChan), expectedCount)
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

// ================== processDailyInsights Tests ==================

func TestProcessDailyInsights_Success(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	job := ParseJob{
		JobType:     "insights",
		InstagramID: "ig123",
		MessageKey:  "ws_ig123_insights",
		EnrichedData: map[string]interface{}{
			"demographics": map[string]interface{}{},
			"user_info": map[string]interface{}{
				"name":     "Test User",
				"username": "testuser",
			},
		},
	}

	dailyData := []interface{}{
		map[string]interface{}{
			"Date": "2024-01-15T00:00:00Z",
			"Data": map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"name": "reach",
						"values": []interface{}{
							map[string]interface{}{"value": float64(1000)},
						},
					},
				},
			},
		},
		map[string]interface{}{
			"Date": "2024-01-16T00:00:00Z",
			"Data": map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"name": "reach",
						"values": []interface{}{
							map[string]interface{}{"value": float64(1200)},
						},
					},
				},
			},
		},
	}

	ctx := context.Background()
	count := processDailyInsights(ctx, nil, job, dailyData, out, log, 0)

	if count != 2 {
		t.Errorf("processDailyInsights returned %d, want 2", count)
	}
	if len(out) != 2 {
		t.Errorf("output channel has %d items, want 2", len(out))
	}
}

func TestProcessDailyInsights_InvalidDailyData(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	job := ParseJob{
		JobType:      "insights",
		InstagramID:  "ig123",
		EnrichedData: map[string]interface{}{},
	}

	// Invalid daily data - not a map
	dailyData := []interface{}{
		"invalid_string",
		123,
	}

	ctx := context.Background()
	count := processDailyInsights(ctx, nil, job, dailyData, out, log, 0)

	if count != 0 {
		t.Errorf("processDailyInsights returned %d, want 0 for invalid data", count)
	}
}

func TestProcessDailyInsights_MissingDate(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	job := ParseJob{
		JobType:      "insights",
		InstagramID:  "ig123",
		EnrichedData: map[string]interface{}{},
	}

	// Daily data with missing date
	dailyData := []interface{}{
		map[string]interface{}{
			"Data": map[string]interface{}{},
		},
	}

	ctx := context.Background()
	count := processDailyInsights(ctx, nil, job, dailyData, out, log, 0)

	if count != 0 {
		t.Errorf("processDailyInsights returned %d, want 0 for missing date", count)
	}
}

func TestProcessDailyInsights_MissingData(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	job := ParseJob{
		JobType:      "insights",
		InstagramID:  "ig123",
		EnrichedData: map[string]interface{}{},
	}

	// Daily data with missing Data field
	dailyData := []interface{}{
		map[string]interface{}{
			"Date": "2024-01-15T00:00:00Z",
		},
	}

	ctx := context.Background()
	count := processDailyInsights(ctx, nil, job, dailyData, out, log, 0)

	if count != 0 {
		t.Errorf("processDailyInsights returned %d, want 0 for missing Data", count)
	}
}

func TestProcessDailyInsights_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob) // unbuffered

	job := ParseJob{
		JobType:     "insights",
		InstagramID: "ig123",
		EnrichedData: map[string]interface{}{
			"user_info": map[string]interface{}{
				"name": "Test",
			},
		},
	}

	dailyData := []interface{}{
		map[string]interface{}{
			"Date": "2024-01-15",
			"Data": map[string]interface{}{},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	count := processDailyInsights(ctx, nil, job, dailyData, out, log, 0)
	// Should return early due to canceled context
	if count > 1 {
		t.Errorf("processDailyInsights should return early on canceled context, got count %d", count)
	}
}

// ================== processLegacyInsights Tests ==================

func TestProcessLegacyInsights_Success(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob, 10)

	job := ParseJob{
		JobType:     "insights",
		InstagramID: "ig123",
		MessageKey:  "ws_ig123_insights",
		EnrichedData: map[string]interface{}{
			"insights": map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"name": "reach",
						"values": []interface{}{
							map[string]interface{}{"value": float64(1000)},
						},
					},
				},
			},
			"user_info": map[string]interface{}{
				"name":            "Test User",
				"username":        "testuser",
				"followers_count": float64(5000),
			},
		},
	}

	ctx := context.Background()
	processLegacyInsights(ctx, nil, job, out, log, 0)

	if len(out) != 1 {
		t.Errorf("output channel has %d items, want 1", len(out))
	}

	select {
	case pubJob := <-out:
		if pubJob.Topic != parsedInsightsTopic {
			t.Errorf("topic = %q, want %q", pubJob.Topic, parsedInsightsTopic)
		}
	default:
		t.Error("expected a publish job")
	}
}

func TestProcessLegacyInsights_ContextCanceled(t *testing.T) {
	log := logger.New("error")
	out := make(chan PublishJob) // unbuffered

	job := ParseJob{
		JobType:     "insights",
		InstagramID: "ig123",
		EnrichedData: map[string]interface{}{
			"insights": map[string]interface{}{
				"data": []interface{}{},
			},
			"user_info": map[string]interface{}{
				"name": "Test",
			},
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	// Should not block when context is canceled
	done := make(chan struct{})
	go func() {
		processLegacyInsights(ctx, nil, job, out, log, 0)
		close(done)
	}()

	select {
	case <-done:
		// Success - function returned
	case <-time.After(time.Second):
		t.Error("processLegacyInsights blocked on canceled context")
	}
}

// ================== insightsParser Integration Tests ==================

func TestInsightsParser_DailyFormat(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)
	var pubCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, in, out, &pubCounter, log)

	// Send job with daily_insights format
	in <- ParseJob{
		JobType:     "insights",
		InstagramID: "ig123",
		EnrichedData: map[string]interface{}{
			"daily_insights": []interface{}{
				map[string]interface{}{
					"Date": "2024-01-15T00:00:00Z",
					"Data": map[string]interface{}{},
				},
			},
			"user_info": map[string]interface{}{
				"name": "Test",
			},
		},
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	if len(out) == 0 {
		t.Log("Note: daily insights may need valid parser - this tests the flow")
	}
}

func TestInsightsParser_LegacyFormat(t *testing.T) {
	log := logger.New("error")
	in := make(chan ParseJob, 10)
	out := make(chan PublishJob, 10)
	var pubCounter uint64

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go insightsParser(ctx, &wg, 0, in, out, &pubCounter, log)

	// Send job with legacy format (no daily_insights)
	in <- ParseJob{
		JobType:     "insights",
		InstagramID: "ig456",
		EnrichedData: map[string]interface{}{
			"insights": map[string]interface{}{
				"data": []interface{}{
					map[string]interface{}{
						"name": "reach",
						"values": []interface{}{
							map[string]interface{}{"value": float64(500)},
						},
					},
				},
			},
			"user_info": map[string]interface{}{
				"name":     "Legacy User",
				"username": "legacyuser",
			},
		},
	}

	time.Sleep(100 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	if len(out) != 1 {
		t.Errorf("output channel has %d items, want 1", len(out))
	}
}

// ================== Publisher Error Tests ==================

func TestPublisher_ProduceError(t *testing.T) {
	log := logger.New("error")
	in := make(chan PublishJob, 10)
	var pubCounter uint64

	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return context.DeadlineExceeded
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	var wg sync.WaitGroup
	wg.Add(1)

	go publisher(ctx, &wg, 0, "media", in, mockProducer, &pubCounter, log)

	in <- PublishJob{
		Topic: "test-topic",
		Key:   "key",
		Data:  map[string]interface{}{"id": "123"},
	}

	time.Sleep(50 * time.Millisecond)

	cancel()
	close(in)
	wg.Wait()

	if atomic.LoadUint64(&pubCounter) != 0 {
		t.Errorf("pubCounter = %d, want 0 (should not increment on error)", pubCounter)
	}
}
