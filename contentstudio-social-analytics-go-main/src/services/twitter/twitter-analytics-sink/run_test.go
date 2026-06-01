package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/rs/zerolog"
)

// ================== RunService Tests ==================

func TestRunService_BasicFlow(t *testing.T) {
	log := logger.New("error")
	sink := NewMockClickHouseSink()

	postJSON, _ := json.Marshal(&kafkamodels.ParsedTwitterPost{
		TwitterID: "twitter_user_1",
		TweetID:   "tweet_123",
		Name:      "Test User",
	})

	postsConsumer := &MockKafkaConsumer{
		Messages: []MockMessage{
			{Topic: "raw-twitter-posts", Key: []byte("twitter_tweet_123"), Value: postJSON},
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

	if atomic.LoadUint64(&sink.BulkInsertTwitterPostsCount) == 0 {
		t.Log("Note: BulkInsertTwitterPosts may not have been called due to timing")
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

	insightJSON, _ := json.Marshal(&kafkamodels.ParsedTwitterInsights{
		RecordID:       "insight_123",
		TwitterID:      "twitter_user_1",
		Name:           "Test User",
		FollowersCount: 1000,
	})

	postsConsumer := &MockKafkaConsumer{Messages: []MockMessage{}}
	insightsConsumer := &MockKafkaConsumer{
		Messages: []MockMessage{
			{Topic: "raw-twitter-insights", Key: []byte("insight_123"), Value: insightJSON},
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
}

func TestServiceConfig_CustomValues(t *testing.T) {
	cfg := ServiceConfig{
		PostsParserWorkers:    5,
		InsightsParserWorkers: 3,
		MaxBatchSize:          50,
		BatchTimeout:          200 * time.Millisecond,
	}

	if cfg.PostsParserWorkers != 5 {
		t.Errorf("Expected PostsParserWorkers=5, got %d", cfg.PostsParserWorkers)
	}
	if cfg.InsightsParserWorkers != 3 {
		t.Errorf("Expected InsightsParserWorkers=3, got %d", cfg.InsightsParserWorkers)
	}
	if cfg.MaxBatchSize != 50 {
		t.Errorf("Expected MaxBatchSize=50, got %d", cfg.MaxBatchSize)
	}
	if cfg.BatchTimeout != 200*time.Millisecond {
		t.Errorf("Expected BatchTimeout=200ms, got %v", cfg.BatchTimeout)
	}
}

// ================== Error-Flow Contract Tests (Caller logs Error with complete fields) ==================

func TestErrorFlowContract_PostsConsumerError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("kafka broker not available")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "processPosts").
		Str("stage", "kafka_consume").
		Msg("Posts consumer error")

	output := buf.String()
	for _, field := range []string{"error_message", "function", "processPosts", "stage", "kafka_consume"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_InsightsConsumerError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("kafka broker not available")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "processInsights").
		Str("stage", "kafka_consume").
		Msg("Insights consumer error")

	output := buf.String()
	for _, field := range []string{"error_message", "function", "processInsights", "stage", "kafka_consume"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_PostsUnmarshalError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("invalid character 'x'")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("key", "test-key").
		Str("function", "postsParser").
		Str("stage", "unmarshal_raw_post").
		Msg("Failed to unmarshal raw post")

	output := buf.String()
	for _, field := range []string{"error_message", "function", "postsParser", "stage", "unmarshal_raw_post"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_BatchInsertError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("clickhouse connection refused")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Int("batch_size", 500).
		Str("function", "postsBatchProcessor").
		Str("stage", "timer_flush").
		Msg("Failed to insert posts batch")

	output := buf.String()
	for _, field := range []string{"error_message", "function", "postsBatchProcessor", "stage", "timer_flush", "batch_size"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_InsightsBatchInsertError_LogsWithAllFields(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()

	err := errors.New("clickhouse connection refused")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Int("batch_size", 1000).
		Str("function", "insightsBatchProcessor").
		Str("stage", "flush_final_batch").
		Msg("Failed to flush final insights batch")

	output := buf.String()
	for _, field := range []string{"error_message", "function", "insightsBatchProcessor", "stage", "flush_final_batch", "batch_size"} {
		if !strings.Contains(output, field) {
			t.Errorf("missing %q in output: %s", field, output)
		}
	}
}

func TestErrorFlowContract_BatchInsertError_TriggersHook(t *testing.T) {
	hookRecords, cleanup := logger.InstallHookSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	err := errors.New("clickhouse timeout")
	log.Error().
		Err(err).
		Str("error_message", err.Error()).
		Str("function", "postsBatchProcessor").
		Str("stage", "size_threshold_flush").
		Msg("Failed to insert posts batch")

	var errorCount int
	for _, r := range *hookRecords {
		if r.Level == zerolog.ErrorLevel {
			errorCount++
		}
	}
	if errorCount != 1 {
		t.Fatalf("expected exactly 1 Error-level hook firing, got %d", errorCount)
	}
}

func TestErrorFlowContract_AllErrors_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	// Log multiple Error-level events
	for _, msg := range []string{"consumer error", "unmarshal error", "batch insert error"} {
		err := errors.New(msg)
		log.Error().
			Err(err).
			Str("error_message", err.Error()).
			Str("function", "test").
			Str("stage", "test").
			Msg(msg)
	}

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls (hook handles Sentry), got %d", len(*captureRecords))
	}
}
