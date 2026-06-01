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

func TestPublishInsightsParallel_EmptyInsights(t *testing.T) {
	log := logger.New("error")
	producer := &kafka.MockProducer{}

	publishInsightsParallel(context.Background(), []kafkamodels.RawFacebookInsights{}, producer, "page123", "workspace123", 1, log)
	// Should return immediately without error
}

func TestPublishInsightsParallel_Success(t *testing.T) {
	log := logger.New("error")
	var publishCount int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddInt32(&publishCount, 1)
			return nil
		},
	}

	insights := []kafkamodels.RawFacebookInsights{
		{PageID: "page123", SavingTime: time.Now()},
		{PageID: "page123", SavingTime: time.Now().Add(-24 * time.Hour)},
		{PageID: "page123", SavingTime: time.Now().Add(-48 * time.Hour)},
	}

	publishInsightsParallel(context.Background(), insights, producer, "page123", "workspace123", 1, log)

	if atomic.LoadInt32(&publishCount) != 3 {
		t.Fatalf("expected 3 insights published, got %d", publishCount)
	}
}

func TestPublishInsightsParallel_ProducerError(t *testing.T) {
	log := logger.New("error")
	var callCount int32

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			count := atomic.AddInt32(&callCount, 1)
			if count == 2 {
				return errors.New("kafka error")
			}
			return nil
		},
	}

	insights := []kafkamodels.RawFacebookInsights{
		{PageID: "page123", SavingTime: time.Now()},
		{PageID: "page123", SavingTime: time.Now().Add(-24 * time.Hour)},
		{PageID: "page123", SavingTime: time.Now().Add(-48 * time.Hour)},
	}

	publishInsightsParallel(context.Background(), insights, producer, "page123", "workspace123", 1, log)

	// All 3 should be attempted
	if atomic.LoadInt32(&callCount) != 3 {
		t.Fatalf("expected 3 produce calls, got %d", callCount)
	}
}

func TestPublishInsightsParallel_ContextCancellation(t *testing.T) {
	log := logger.New("error")

	ctx, cancel := context.WithCancel(context.Background())

	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			time.Sleep(100 * time.Millisecond)
			return nil
		},
	}

	insights := make([]kafkamodels.RawFacebookInsights, 100)
	for i := 0; i < 100; i++ {
		insights[i] = kafkamodels.RawFacebookInsights{PageID: "page123", SavingTime: time.Now()}
	}

	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	publishInsightsParallel(ctx, insights, producer, "page123", "workspace123", 1, log)
}

func TestGenerateInsightID(t *testing.T) {
	savingTime := time.Date(2024, 1, 15, 12, 0, 0, 0, time.UTC)

	id1 := generateInsightID("page123", savingTime)
	id2 := generateInsightID("page123", savingTime)
	id3 := generateInsightID("page456", savingTime)
	id4 := generateInsightID("page123", savingTime.Add(24*time.Hour))

	// Same inputs should produce same ID
	if id1 != id2 {
		t.Fatalf("expected same ID for same inputs, got '%s' and '%s'", id1, id2)
	}

	// Different page should produce different ID
	if id1 == id3 {
		t.Fatal("expected different ID for different page")
	}

	// Different date should produce different ID
	if id1 == id4 {
		t.Fatal("expected different ID for different date")
	}

	// ID should be a valid hex string (32 chars for MD5)
	if len(id1) != 32 {
		t.Fatalf("expected 32 char hex string, got %d chars", len(id1))
	}
}

func TestJoinStrings(t *testing.T) {
	tests := []struct {
		name     string
		slice    []string
		sep      string
		expected string
	}{
		{
			name:     "empty slice",
			slice:    []string{},
			sep:      ",",
			expected: "",
		},
		{
			name:     "single element",
			slice:    []string{"a"},
			sep:      ",",
			expected: "a",
		},
		{
			name:     "multiple elements",
			slice:    []string{"a", "b", "c"},
			sep:      ",",
			expected: "a,b,c",
		},
		{
			name:     "different separator",
			slice:    []string{"a", "b", "c"},
			sep:      "-",
			expected: "a-b-c",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := joinStrings(tt.slice, tt.sep)
			if result != tt.expected {
				t.Fatalf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestConvertInsightsToInterface(t *testing.T) {
	insights := []kafkamodels.RawFacebookInsights{
		{PageID: "page1"},
		{PageID: "page2"},
	}

	result := convertInsightsToInterface(insights)

	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}

	insight1, ok := result[0].(kafkamodels.RawFacebookInsights)
	if !ok {
		t.Fatal("expected RawFacebookInsights type")
	}
	if insight1.PageID != "page1" {
		t.Fatalf("expected PageID 'page1', got '%s'", insight1.PageID)
	}
}

func TestConvertInsightsToInterface_Empty(t *testing.T) {
	result := convertInsightsToInterface([]kafkamodels.RawFacebookInsights{})

	if len(result) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result))
	}
}
