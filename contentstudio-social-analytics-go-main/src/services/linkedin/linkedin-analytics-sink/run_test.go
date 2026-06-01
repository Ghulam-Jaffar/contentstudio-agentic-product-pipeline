package main

import (
	"context"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== AnalyticsSinkConfig Tests ==================

func TestDefaultAnalyticsSinkConfig(t *testing.T) {
	cfg := DefaultAnalyticsSinkConfig()

	if cfg.PostsParserWorkers != postsParserWorkers {
		t.Errorf("PostsParserWorkers = %d, want %d", cfg.PostsParserWorkers, postsParserWorkers)
	}
	if cfg.InsightsParserWorkers != insightsParserWorkers {
		t.Errorf("InsightsParserWorkers = %d, want %d", cfg.InsightsParserWorkers, insightsParserWorkers)
	}
	if cfg.BatchProcessorsPerType != batchProcessorsPerType {
		t.Errorf("BatchProcessorsPerType = %d, want %d", cfg.BatchProcessorsPerType, batchProcessorsPerType)
	}
	if cfg.MaxBatchSize != maxBatchSize {
		t.Errorf("MaxBatchSize = %d, want %d", cfg.MaxBatchSize, maxBatchSize)
	}
	if cfg.BatchTimeout != batchTimeout {
		t.Errorf("BatchTimeout = %v, want %v", cfg.BatchTimeout, batchTimeout)
	}
	if cfg.MessageChanSize != messageChanSize {
		t.Errorf("MessageChanSize = %d, want %d", cfg.MessageChanSize, messageChanSize)
	}
}

// ================== NewAnalyticsSinkService Tests ==================

func TestNewAnalyticsSinkService(t *testing.T) {
	cfg := AnalyticsSinkConfig{
		PostsParserWorkers:     2,
		InsightsParserWorkers:  2,
		BatchProcessorsPerType: 1,
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
		MessageChanSize:        10,
	}

	log := logger.New("error")
	deps := AnalyticsSinkDependencies{Log: log}

	svc := NewAnalyticsSinkService(cfg, deps)

	if svc == nil {
		t.Fatal("NewAnalyticsSinkService returned nil")
	}
	if svc.config.PostsParserWorkers != 2 {
		t.Errorf("PostsParserWorkers = %d, want 2", svc.config.PostsParserWorkers)
	}
	if svc.batchCollectors == nil {
		t.Error("batchCollectors is nil")
	}
}

// ================== AnalyticsSinkService.Run Tests ==================

func TestAnalyticsSinkService_Run_ContextCancel(t *testing.T) {
	cfg := AnalyticsSinkConfig{
		PostsParserWorkers:     1,
		InsightsParserWorkers:  1,
		BatchProcessorsPerType: 0,
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
		MessageChanSize:        10,
	}

	log := logger.New("error")

	pagePostsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	pageInsightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profilePostsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profileInsightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := AnalyticsSinkDependencies{
		PagePostsConsumer:       pagePostsConsumer,
		PageInsightsConsumer:    pageInsightsConsumer,
		ProfilePostsConsumer:    profilePostsConsumer,
		ProfileInsightsConsumer: profileInsightsConsumer,
		Sink:                    nil,
		Log:                     log,
	}

	svc := NewAnalyticsSinkService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	cancel()

	select {
	case err := <-done:
		if err != nil {
			t.Errorf("Run returned error: %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Run did not exit after context cancel")
	}
}

func TestAnalyticsSinkService_Run_ProcessesPosts(t *testing.T) {
	cfg := AnalyticsSinkConfig{
		PostsParserWorkers:     1,
		InsightsParserWorkers:  1,
		BatchProcessorsPerType: 0,
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
		MessageChanSize:        10,
	}

	log := logger.New("error")

	pagePostsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topicRawPagePosts, []byte("key"), []byte(`{"id":"post123"}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}
	pageInsightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profilePostsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profileInsightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := AnalyticsSinkDependencies{
		PagePostsConsumer:       pagePostsConsumer,
		PageInsightsConsumer:    pageInsightsConsumer,
		ProfilePostsConsumer:    profilePostsConsumer,
		ProfileInsightsConsumer: profileInsightsConsumer,
		Log:                     log,
	}

	svc := NewAnalyticsSinkService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.PickedPosts < 1 {
		t.Errorf("PickedPosts = %d, want >= 1", metrics.PickedPosts)
	}

	cancel()
	<-done
}

func TestAnalyticsSinkService_Run_ProcessesInsights(t *testing.T) {
	cfg := AnalyticsSinkConfig{
		PostsParserWorkers:     1,
		InsightsParserWorkers:  1,
		BatchProcessorsPerType: 0,
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
		MessageChanSize:        10,
	}

	log := logger.New("error")

	pagePostsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	pageInsightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topicRawPageInsights, []byte("key"), []byte(`{"insights":{}}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profilePostsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	profileInsightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := AnalyticsSinkDependencies{
		PagePostsConsumer:       pagePostsConsumer,
		PageInsightsConsumer:    pageInsightsConsumer,
		ProfilePostsConsumer:    profilePostsConsumer,
		ProfileInsightsConsumer: profileInsightsConsumer,
		Log:                     log,
	}

	svc := NewAnalyticsSinkService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.PickedInsights < 1 {
		t.Errorf("PickedInsights = %d, want >= 1", metrics.PickedInsights)
	}

	cancel()
	<-done
}

// ================== GetMetrics Tests ==================

func TestAnalyticsSinkService_GetMetrics_Initial(t *testing.T) {
	cfg := AnalyticsSinkConfig{
		MaxBatchSize: 100,
	}

	log := logger.New("error")
	deps := AnalyticsSinkDependencies{Log: log}

	svc := NewAnalyticsSinkService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.PickedPosts != 0 {
		t.Errorf("PickedPosts = %d, want 0", metrics.PickedPosts)
	}
	if metrics.PickedInsights != 0 {
		t.Errorf("PickedInsights = %d, want 0", metrics.PickedInsights)
	}
}

// ================== AnalyticsSinkMetrics Tests ==================

func TestAnalyticsSinkMetrics_Struct(t *testing.T) {
	m := AnalyticsSinkMetrics{
		PickedPosts:      10,
		PickedInsights:   20,
		ParsedPosts:      8,
		ParsedInsights:   15,
		InsertedPosts:    5,
		InsertedInsights: 12,
	}

	if m.PickedPosts != 10 {
		t.Errorf("PickedPosts = %d, want 10", m.PickedPosts)
	}
	if m.PickedInsights != 20 {
		t.Errorf("PickedInsights = %d, want 20", m.PickedInsights)
	}
}
