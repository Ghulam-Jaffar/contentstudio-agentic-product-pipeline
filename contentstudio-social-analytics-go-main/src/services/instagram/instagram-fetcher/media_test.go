package main

import (
	"context"
	"fmt"
	"sync"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestEnrichedMedia_Struct(t *testing.T) {
	media := EnrichedMedia{
		RawInstagramMedia: &kafkamodels.RawInstagramMedia{
			ID:        "12345",
			MediaType: "IMAGE",
			MediaURL:  "https://example.com/image.jpg",
			Permalink: "https://instagram.com/p/12345",
			Caption:   "Test caption",
			Timestamp: "2024-01-15T10:30:00+0000",
		},
		Insights: &kafkamodels.RawInstagramMediaInsights{
			Data: []struct {
				Name   string `json:"name"`
				Period string `json:"period"`
				Values []struct {
					Value interface{} `json:"value"`
				} `json:"values"`
				Title       string `json:"title"`
				Description string `json:"description"`
				ID          string `json:"id"`
			}{
				{Name: "reach", Period: "lifetime", Title: "Reach"},
			},
		},
		UserInfo: map[string]interface{}{
			"username":   "testuser",
			"followers":  10000,
			"is_private": false,
		},
	}

	if media.ID != "12345" {
		t.Fatalf("expected ID '12345', got '%s'", media.ID)
	}
	if media.MediaType != "IMAGE" {
		t.Fatalf("expected MediaType 'IMAGE', got '%s'", media.MediaType)
	}
	if media.Insights == nil {
		t.Fatal("expected Insights to be set")
	}
	if len(media.Insights.Data) != 1 {
		t.Fatalf("expected 1 insight data, got %d", len(media.Insights.Data))
	}
	if media.UserInfo["username"] != "testuser" {
		t.Fatalf("expected username 'testuser', got '%v'", media.UserInfo["username"])
	}
}

func TestMediaPublishJob_Struct(t *testing.T) {
	job := MediaPublishJob{
		Media: EnrichedMedia{
			RawInstagramMedia: &kafkamodels.RawInstagramMedia{
				ID: "media123",
			},
		},
		MessageKey: "workspace_instagram_media123",
	}

	if job.Media.ID != "media123" {
		t.Fatalf("expected Media.ID 'media123', got '%s'", job.Media.ID)
	}
	if job.MessageKey != "workspace_instagram_media123" {
		t.Fatalf("expected MessageKey 'workspace_instagram_media123', got '%s'", job.MessageKey)
	}
}

func TestMediaPublishResult_Struct(t *testing.T) {
	cases := []struct {
		name     string
		result   MediaPublishResult
		expected MediaPublishResult
	}{
		{
			name: "successful result",
			result: MediaPublishResult{
				JobIndex: 0,
				Success:  true,
				Error:    nil,
			},
			expected: MediaPublishResult{
				JobIndex: 0,
				Success:  true,
				Error:    nil,
			},
		},
		{
			name: "failed result",
			result: MediaPublishResult{
				JobIndex: 5,
				Success:  false,
				Error:    nil,
			},
			expected: MediaPublishResult{
				JobIndex: 5,
				Success:  false,
				Error:    nil,
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.result.JobIndex != tc.expected.JobIndex {
				t.Fatalf("expected JobIndex %d, got %d", tc.expected.JobIndex, tc.result.JobIndex)
			}
			if tc.result.Success != tc.expected.Success {
				t.Fatalf("expected Success %v, got %v", tc.expected.Success, tc.result.Success)
			}
		})
	}
}

func TestMaxMediaPublishWorkers(t *testing.T) {
	if maxMediaPublishWorkers != 10 {
		t.Fatalf("expected maxMediaPublishWorkers to be 10, got %d", maxMediaPublishWorkers)
	}
}

func TestEnrichedMedia_NilFields(t *testing.T) {
	media := EnrichedMedia{
		RawInstagramMedia: nil,
		Insights:          nil,
		UserInfo:          nil,
	}

	if media.RawInstagramMedia != nil {
		t.Fatal("expected RawInstagramMedia to be nil")
	}
	if media.Insights != nil {
		t.Fatal("expected Insights to be nil")
	}
	if media.UserInfo != nil {
		t.Fatal("expected UserInfo to be nil")
	}
}

func TestEnrichedMedia_WithOnlyRawMedia(t *testing.T) {
	media := EnrichedMedia{
		RawInstagramMedia: &kafkamodels.RawInstagramMedia{
			ID:        "test123",
			MediaType: "VIDEO",
		},
	}

	if media.ID != "test123" {
		t.Fatalf("expected ID 'test123', got '%s'", media.ID)
	}
	if media.Insights != nil {
		t.Fatal("expected Insights to be nil")
	}
}

// ================== publishMediaParallel Tests ==================

func TestPublishMediaParallel_Success(t *testing.T) {
	log := logger.New("error")

	var produceCount int
	var mu sync.Mutex
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			mu.Lock()
			produceCount++
			mu.Unlock()
			return nil
		},
	}

	mediaItems := []EnrichedMedia{
		{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m1", MediaType: "IMAGE"}},
		{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m2", MediaType: "VIDEO"}},
		{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m3", MediaType: "IMAGE"}},
	}

	ctx := context.Background()
	publishMediaParallel(ctx, mediaItems, mockProducer, "ig123", "ws456", 0, log)

	mu.Lock()
	defer mu.Unlock()
	if produceCount != 3 {
		t.Errorf("Produce called %d times, want 3", produceCount)
	}
}

func TestPublishMediaParallel_Empty(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			t.Fatal("Produce should not be called for empty media")
			return nil
		},
	}

	ctx := context.Background()
	publishMediaParallel(ctx, nil, mockProducer, "ig123", "", 0, log)
}

func TestPublishMediaParallel_WithoutWorkspaceID(t *testing.T) {
	log := logger.New("error")

	var capturedKeys []string
	var mu sync.Mutex
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			mu.Lock()
			capturedKeys = append(capturedKeys, string(key))
			mu.Unlock()
			return nil
		},
	}

	mediaItems := []EnrichedMedia{
		{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m1", MediaType: "IMAGE"}},
	}

	ctx := context.Background()
	publishMediaParallel(ctx, mediaItems, mockProducer, "ig123", "", 0, log)

	mu.Lock()
	defer mu.Unlock()
	if len(capturedKeys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(capturedKeys))
	}
	if capturedKeys[0] != "ig123_m1" {
		t.Errorf("key = %q, want %q", capturedKeys[0], "ig123_m1")
	}
}

func TestPublishMediaParallel_WithWorkspaceID(t *testing.T) {
	log := logger.New("error")

	var capturedKeys []string
	var mu sync.Mutex
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			mu.Lock()
			capturedKeys = append(capturedKeys, string(key))
			mu.Unlock()
			return nil
		},
	}

	mediaItems := []EnrichedMedia{
		{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m1", MediaType: "IMAGE"}},
	}

	ctx := context.Background()
	publishMediaParallel(ctx, mediaItems, mockProducer, "ig123", "ws456", 0, log)

	mu.Lock()
	defer mu.Unlock()
	if len(capturedKeys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(capturedKeys))
	}
	if capturedKeys[0] != "ws456_ig123_m1" {
		t.Errorf("key = %q, want %q", capturedKeys[0], "ws456_ig123_m1")
	}
}

func TestPublishMediaParallel_ProduceError(t *testing.T) {
	log := logger.New("error")

	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return fmt.Errorf("kafka unavailable")
		},
	}

	mediaItems := []EnrichedMedia{
		{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m1", MediaType: "IMAGE"}},
		{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m2", MediaType: "VIDEO"}},
	}

	ctx := context.Background()
	publishMediaParallel(ctx, mediaItems, mockProducer, "ig123", "", 0, log)
	// Should complete without panic; errors are logged
}

func TestPublishMediaParallel_ContextCancel(t *testing.T) {
	log := logger.New("error")

	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	// Create many items to increase chance of context cancellation during dispatch
	mediaItems := make([]EnrichedMedia, 50)
	for i := range mediaItems {
		mediaItems[i] = EnrichedMedia{
			RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: fmt.Sprintf("m%d", i), MediaType: "IMAGE"},
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	publishMediaParallel(ctx, mediaItems, mockProducer, "ig123", "", 0, log)
	// Should complete without panic
}

// ================== mediaPublisher Tests ==================

func TestMediaPublisher_Success(t *testing.T) {
	log := logger.New("error")

	var produceCount int
	var mu sync.Mutex
	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			mu.Lock()
			produceCount++
			mu.Unlock()
			return nil
		},
	}

	jobChan := make(chan MediaPublishJob, 5)
	resultChan := make(chan MediaPublishResult, 5)

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaPublisher(context.Background(), &wg, 0, jobChan, resultChan, mockProducer, log)

	jobChan <- MediaPublishJob{
		Media:      EnrichedMedia{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m1", MediaType: "IMAGE"}},
		MessageKey: "key1",
	}
	jobChan <- MediaPublishJob{
		Media:      EnrichedMedia{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m2", MediaType: "VIDEO"}},
		MessageKey: "key2",
	}
	close(jobChan)
	wg.Wait()

	mu.Lock()
	defer mu.Unlock()
	if produceCount != 2 {
		t.Errorf("Produce called %d times, want 2", produceCount)
	}

	// Collect results
	close(resultChan)
	var successCount int
	for r := range resultChan {
		if r.Success {
			successCount++
		}
	}
	if successCount != 2 {
		t.Errorf("successCount = %d, want 2", successCount)
	}
}

func TestMediaPublisher_ProduceError(t *testing.T) {
	log := logger.New("error")

	mockProducer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return fmt.Errorf("kafka error")
		},
	}

	jobChan := make(chan MediaPublishJob, 5)
	resultChan := make(chan MediaPublishResult, 5)

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaPublisher(context.Background(), &wg, 0, jobChan, resultChan, mockProducer, log)

	jobChan <- MediaPublishJob{
		Media:      EnrichedMedia{RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m1", MediaType: "IMAGE"}},
		MessageKey: "key1",
	}
	close(jobChan)
	wg.Wait()

	close(resultChan)
	result := <-resultChan
	if result.Success {
		t.Error("expected failure result")
	}
	if result.Error == nil {
		t.Error("expected non-nil error")
	}
}

func TestMediaPublisher_ContextCancel(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	jobChan := make(chan MediaPublishJob, 5)
	resultChan := make(chan MediaPublishResult, 5)

	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaPublisher(ctx, &wg, 0, jobChan, resultChan, mockProducer, log)

	cancel()
	wg.Wait()
	// Should exit cleanly
}

func TestMediaPublisher_MarshalError(t *testing.T) {
	log := logger.New("error")
	mockProducer := &kafka.MockProducer{}

	jobChan := make(chan MediaPublishJob, 5)
	resultChan := make(chan MediaPublishResult, 5)

	var wg sync.WaitGroup
	wg.Add(1)
	go mediaPublisher(context.Background(), &wg, 0, jobChan, resultChan, mockProducer, log)

	// json.Marshal fails on channels
	jobChan <- MediaPublishJob{
		Media: EnrichedMedia{
			RawInstagramMedia: &kafkamodels.RawInstagramMedia{ID: "m1"},
			UserInfo:          map[string]interface{}{"bad": make(chan int)},
		},
		MessageKey: "key1",
	}
	close(jobChan)
	wg.Wait()

	close(resultChan)
	result := <-resultChan
	if result.Success {
		t.Error("expected failure on marshal error")
	}
	if result.Error == nil {
		t.Error("expected non-nil error for marshal failure")
	}
}
