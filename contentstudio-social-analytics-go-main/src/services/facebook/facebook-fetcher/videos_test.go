package main

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestPublishVideosParallel_EmptyVideos(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	publishVideosParallel(context.Background(), []kafkamodels.RawFacebookVideo{}, producer, "page123", "workspace123", 1, log)
	// Should return immediately without error
}

func TestPublishVideosParallel_Success(t *testing.T) {
	log := logger.New("error")
	var publishCount int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&publishCount, 1)
			return nil
		},
	}

	videos := []kafkamodels.RawFacebookVideo{
		{ID: "video1"},
		{ID: "video2"},
		{ID: "video3"},
	}

	publishVideosParallel(context.Background(), videos, producer, "page123", "workspace123", 1, log)

	if atomic.LoadInt32(&publishCount) != 3 {
		t.Fatalf("expected 3 videos published, got %d", publishCount)
	}
}

func TestPublishVideosParallel_WithoutWorkspaceID(t *testing.T) {
	log := logger.New("error")
	var keys []string

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			keys = append(keys, string(key))
			return nil
		},
	}

	videos := []kafkamodels.RawFacebookVideo{
		{ID: "video1"},
	}

	publishVideosParallel(context.Background(), videos, producer, "page123", "", 1, log)

	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	if keys[0] != "page123_video1" {
		t.Fatalf("expected key 'page123_video1', got '%s'", keys[0])
	}
}

func TestPublishVideosParallel_WithWorkspaceID(t *testing.T) {
	log := logger.New("error")
	var keys []string

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			keys = append(keys, string(key))
			return nil
		},
	}

	videos := []kafkamodels.RawFacebookVideo{
		{ID: "video1"},
	}

	publishVideosParallel(context.Background(), videos, producer, "page123", "workspace456", 1, log)

	if len(keys) != 1 {
		t.Fatalf("expected 1 key, got %d", len(keys))
	}
	if keys[0] != "workspace456_page123_video1" {
		t.Fatalf("expected key 'workspace456_page123_video1', got '%s'", keys[0])
	}
}

func TestPublishVideosParallel_ProducerError(t *testing.T) {
	log := logger.New("error")
	var successCount int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			if string(key) == "workspace123_page123_video2" {
				return errors.New("kafka error")
			}
			atomic.AddInt32(&successCount, 1)
			return nil
		},
	}

	videos := []kafkamodels.RawFacebookVideo{
		{ID: "video1"},
		{ID: "video2"},
		{ID: "video3"},
	}

	publishVideosParallel(context.Background(), videos, producer, "page123", "workspace123", 1, log)

	if atomic.LoadInt32(&successCount) != 2 {
		t.Fatalf("expected 2 videos published successfully, got %d", successCount)
	}
}

func TestPublishVideosParallel_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	ctx, cancel := context.WithCancel(context.Background())

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	}

	videos := make([]kafkamodels.RawFacebookVideo, 100)
	for i := 0; i < 100; i++ {
		videos[i] = kafkamodels.RawFacebookVideo{ID: "video" + string(rune(i))}
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	publishVideosParallel(ctx, videos, producer, "page123", "workspace123", 1, log)
}

func TestConvertVideosToInterface(t *testing.T) {
	videos := []kafkamodels.RawFacebookVideo{
		{ID: "video1"},
		{ID: "video2"},
	}

	result := convertVideosToInterface(videos)

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	video1, ok := result[0].(kafkamodels.RawFacebookVideo)
	if !ok {
		t.Fatal("expected RawFacebookVideo type")
	}
	if video1.ID != "video1" {
		t.Fatalf("expected ID 'video1', got '%s'", video1.ID)
	}
}

func TestConvertVideosToInterface_Empty(t *testing.T) {
	result := convertVideosToInterface([]kafkamodels.RawFacebookVideo{})

	if len(result) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result))
	}
}
