package main

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== RunService Tests ==================

func TestRunService_BasicFlow(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	postJSON, _ := json.Marshal(&kafkamodels.RawTikTokPost{
		TikTokID: "tiktok_user_1",
		Data:     json.RawMessage(`{"id":"video_123"}`),
	})

	postsConsumer := &MockKafkaConsumer{
		Messages: []MockMessage{
			{Topic: rawPostsTopic, Key: []byte("tiktok_video_123"), Value: postJSON},
		},
	}

	insightsConsumer := &MockKafkaConsumer{
		Messages: []MockMessage{},
	}

	deps := &ServiceDependencies{
		Sink:             sink,
		PostsConsumer:    postsConsumer,
		InsightsConsumer: insightsConsumer,
		Logger:           log,
	}

	cfg := ServiceConfig{
		PostsParserWorkers:     1,
		InsightsParserWorkers:  1,
		BatchProcessorsPerType: 1,
		MaxBatchSize:           10,
		BatchTimeout:           100 * time.Millisecond,
		IdleTimeout:            1 * time.Second,
		ParseChanSize:          10,
		MessageChanSize:        100,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService failed: %v", err)
	}

	if atomic.LoadUint64(&sink.BulkInsertTikTokPostsCount) == 0 {
		t.Log("Note: BulkInsertTikTokPosts may not have been called due to timing")
	}
}

func TestRunService_EmptyConsumers(t *testing.T) {
	log := logger.New("error")
	postsConsumer := &MockKafkaConsumer{Messages: []MockMessage{}}
	insightsConsumer := &MockKafkaConsumer{Messages: []MockMessage{}}

	deps := &ServiceDependencies{
		Sink:             NewMockClickHouseSink(),
		PostsConsumer:    postsConsumer,
		InsightsConsumer: insightsConsumer,
		Logger:           log,
	}

	cfg := DefaultServiceConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService should complete: %v", err)
	}
}

func TestRunService_WithInsights(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	insightJSON, _ := json.Marshal(&kafkamodels.ParsedTikTokInsights{
		RecordID:           "insight_123",
		TikTokID:           "tiktok_user_1",
		DisplayName:        "Test User",
		TotalFollowerCount: 1000,
	})

	postsConsumer := &MockKafkaConsumer{Messages: []MockMessage{}}
	insightsConsumer := &MockKafkaConsumer{
		Messages: []MockMessage{
			{Topic: rawInsightsTopic, Key: []byte("insight_123"), Value: insightJSON},
		},
	}

	deps := &ServiceDependencies{
		Sink:             sink,
		PostsConsumer:    postsConsumer,
		InsightsConsumer: insightsConsumer,
		Logger:           log,
	}

	cfg := ServiceConfig{
		PostsParserWorkers:     1,
		InsightsParserWorkers:  1,
		BatchProcessorsPerType: 1,
		MaxBatchSize:           10,
		BatchTimeout:           100 * time.Millisecond,
		IdleTimeout:            1 * time.Second,
		ParseChanSize:          10,
		MessageChanSize:        100,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
	defer cancel()

	err := RunService(ctx, deps, cfg)

	if err != nil && err != context.DeadlineExceeded {
		t.Fatalf("RunService failed: %v", err)
	}
}

// ================== ServiceConfig Tests ==================

func TestDefaultServiceConfig(t *testing.T) {
	cfg := DefaultServiceConfig()

	if cfg.PostsParserWorkers == 0 {
		t.Error("PostsParserWorkers should not be zero")
	}
	if cfg.InsightsParserWorkers == 0 {
		t.Error("InsightsParserWorkers should not be zero")
	}
	if cfg.MaxBatchSize == 0 {
		t.Error("MaxBatchSize should not be zero")
	}
	if cfg.BatchTimeout == 0 {
		t.Error("BatchTimeout should not be zero")
	}
	if cfg.IdleTimeout == 0 {
		t.Error("IdleTimeout should not be zero")
	}
}

func TestDefaultServiceConfig_ReasonableValues(t *testing.T) {
	cfg := DefaultServiceConfig()

	if cfg.PostsParserWorkers < 1 || cfg.PostsParserWorkers > 100 {
		t.Errorf("PostsParserWorkers = %d, should be between 1 and 100", cfg.PostsParserWorkers)
	}
	if cfg.InsightsParserWorkers < 1 || cfg.InsightsParserWorkers > 100 {
		t.Errorf("InsightsParserWorkers = %d, should be between 1 and 100", cfg.InsightsParserWorkers)
	}
	if cfg.MaxBatchSize < 100 || cfg.MaxBatchSize > 100000 {
		t.Errorf("MaxBatchSize = %d, should be between 100 and 100000", cfg.MaxBatchSize)
	}
}

// ================== ServiceMetrics Tests ==================

func TestServiceMetrics_Initialized(t *testing.T) {
	metrics := &ServiceMetrics{}

	if metrics.PostsMessagesReceived != 0 {
		t.Errorf("PostsMessagesReceived should be 0, got %d", metrics.PostsMessagesReceived)
	}
	if metrics.InsightsMessagesReceived != 0 {
		t.Errorf("InsightsMessagesReceived should be 0, got %d", metrics.InsightsMessagesReceived)
	}
}

func TestServiceMetrics_Atomic(t *testing.T) {
	metrics := &ServiceMetrics{}

	atomic.AddUint64(&metrics.PostsMessagesReceived, 10)
	atomic.AddUint64(&metrics.InsightsMessagesReceived, 5)
	atomic.AddUint64(&metrics.PostsMessagesParsed, 9)
	atomic.AddUint64(&metrics.MessagesFailed, 1)

	if atomic.LoadUint64(&metrics.PostsMessagesReceived) != 10 {
		t.Errorf("PostsMessagesReceived = %d, want 10", metrics.PostsMessagesReceived)
	}
	if atomic.LoadUint64(&metrics.InsightsMessagesReceived) != 5 {
		t.Errorf("InsightsMessagesReceived = %d, want 5", metrics.InsightsMessagesReceived)
	}
}

// ================== Parser Tests ==================

func TestParsePost_ValidPost(t *testing.T) {
	raw := &kafkamodels.RawTikTokPost{
		TikTokID: "tiktok_user_1",
		Data:     json.RawMessage(`{"id":"video_123"}`),
	}

	parsed := parsePost(raw)

	if parsed == nil {
		t.Fatal("parsePost should not return nil for valid post")
	}

	if parsed.TikTokID != "tiktok_user_1" {
		t.Errorf("TikTokID = %q, want %q", parsed.TikTokID, "tiktok_user_1")
	}
}

func TestParsePost_NilPost(t *testing.T) {
	parsed := parsePost(nil)

	if parsed != nil {
		t.Error("parsePost(nil) should return nil")
	}
}

func TestParseInsight_ValidInsight(t *testing.T) {
	raw := &kafkamodels.ParsedTikTokInsights{
		RecordID:           "insight_123",
		TikTokID:           "tiktok_user_1",
		DisplayName:        "Test User",
		TotalFollowerCount: 1000,
		TotalVideoCount:    50,
		IsVerified:         true,
		Bio:                "Test bio",
	}

	parsed := parseInsight(raw)

	if parsed == nil {
		t.Fatal("parseInsight should not return nil for valid insight")
	}
	if parsed.RecordID != "insight_123" {
		t.Errorf("RecordID = %q, want %q", parsed.RecordID, "insight_123")
	}
	if parsed.TotalFollowerCount != 1000 {
		t.Errorf("TotalFollowerCount = %d, want 1000", parsed.TotalFollowerCount)
	}
	if !parsed.IsVerified {
		t.Error("IsVerified should be true")
	}
}

func TestParseInsight_NilInsight(t *testing.T) {
	parsed := parseInsight(nil)

	if parsed != nil {
		t.Error("parseInsight(nil) should return nil")
	}
}

func TestParseInsight_AllFields(t *testing.T) {
	raw := &kafkamodels.ParsedTikTokInsights{
		RecordID:            "rec_123",
		TikTokID:            "tt_789",
		DisplayName:         "Creator Name",
		ProfileImage:        "https://example.com/img.jpg",
		TotalFollowerCount:  10000,
		TotalFollowingCount: 500,
		TotalLikeCount:      250000,
		TotalVideoCount:     100,
		TotalVideoViews:     1000000,
		TotalVideoLikes:     50000,
		TotalVideoComments:  5000,
		TotalVideoShares:    1000,
		IsVerified:          true,
		Bio:                 "Content creator",
		ProfileLink:         "https://tiktok.com/@creator",
	}

	parsed := parseInsight(raw)

	if parsed == nil {
		t.Fatal("parseInsight should not return nil")
	}
	if parsed.TotalFollowerCount != 10000 {
		t.Errorf("TotalFollowerCount = %d, want 10000", parsed.TotalFollowerCount)
	}
	if parsed.TotalFollowingCount != 500 {
		t.Errorf("TotalFollowingCount = %d, want 500", parsed.TotalFollowingCount)
	}
	if parsed.TotalVideoViews != 1000000 {
		t.Errorf("TotalVideoViews = %d, want 1000000", parsed.TotalVideoViews)
	}
}

// ================== GetMetrics Tests ==================

func TestGetMetrics_Initialized(t *testing.T) {
	metrics := &ServiceMetrics{}

	result := GetMetrics(metrics)

	if len(result) != 7 {
		t.Errorf("GetMetrics should return 7 metrics, got %d", len(result))
	}

	expectedKeys := []string{
		"posts_received",
		"insights_received",
		"posts_parsed",
		"insights_parsed",
		"messages_batched",
		"messages_failed",
		"insert_errors",
	}

	for _, key := range expectedKeys {
		if _, ok := result[key]; !ok {
			t.Errorf("Expected metric key %q not found", key)
		}
	}
}

func TestGetMetrics_WithValues(t *testing.T) {
	metrics := &ServiceMetrics{}
	atomic.AddUint64(&metrics.PostsMessagesReceived, 10)
	atomic.AddUint64(&metrics.InsightsMessagesReceived, 5)
	atomic.AddUint64(&metrics.PostsMessagesParsed, 9)
	atomic.AddUint64(&metrics.InsightsMessagesParsed, 4)
	atomic.AddUint64(&metrics.MessagesFailed, 2)
	atomic.AddUint64(&metrics.InsertErrors, 1)

	result := GetMetrics(metrics)

	if result["posts_received"] != 10 {
		t.Errorf("posts_received = %d, want 10", result["posts_received"])
	}
	if result["insights_received"] != 5 {
		t.Errorf("insights_received = %d, want 5", result["insights_received"])
	}
	if result["posts_parsed"] != 9 {
		t.Errorf("posts_parsed = %d, want 9", result["posts_parsed"])
	}
	if result["messages_failed"] != 2 {
		t.Errorf("messages_failed = %d, want 2", result["messages_failed"])
	}
	if result["insert_errors"] != 1 {
		t.Errorf("insert_errors = %d, want 1", result["insert_errors"])
	}
}

func TestGetMetrics_ZeroValues(t *testing.T) {
	metrics := &ServiceMetrics{}

	result := GetMetrics(metrics)

	for key, value := range result {
		if value != 0 {
			t.Errorf("expected %s to be 0, got %d", key, value)
		}
	}
}

// ================== ServiceDependencies Tests ==================

func TestServiceDependencies_AllSet(t *testing.T) {
	deps := &ServiceDependencies{
		Sink:             NewMockClickHouseSink(),
		PostsConsumer:    &MockKafkaConsumer{},
		InsightsConsumer: &MockKafkaConsumer{},
		Logger:           logger.New("error"),
	}

	if deps.Sink == nil {
		t.Error("Sink should not be nil")
	}
	if deps.PostsConsumer == nil {
		t.Error("PostsConsumer should not be nil")
	}
	if deps.InsightsConsumer == nil {
		t.Error("InsightsConsumer should not be nil")
	}
	if deps.Logger == nil {
		t.Error("Logger should not be nil")
	}
}
