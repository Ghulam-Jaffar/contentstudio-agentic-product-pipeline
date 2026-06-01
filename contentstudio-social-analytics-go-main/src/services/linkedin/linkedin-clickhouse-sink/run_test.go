package main

import (
	"context"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== ClickHouseSinkConfig Tests ==================

func TestDefaultClickHouseSinkConfig(t *testing.T) {
	cfg := DefaultClickHouseSinkConfig()

	if cfg.PostsWorkers != postsWorkers {
		t.Errorf("PostsWorkers = %d, want %d", cfg.PostsWorkers, postsWorkers)
	}
	if cfg.InsightsWorkers != insightsWorkers {
		t.Errorf("InsightsWorkers = %d, want %d", cfg.InsightsWorkers, insightsWorkers)
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
}

// ================== NewClickHouseSinkService Tests ==================

func TestNewClickHouseSinkService(t *testing.T) {
	cfg := ClickHouseSinkConfig{
		PostsWorkers:    2,
		InsightsWorkers: 2,
		MaxBatchSize:    100,
	}

	log := logger.New("error")
	deps := ClickHouseSinkDependencies{Log: log}

	svc := NewClickHouseSinkService(cfg, deps)

	if svc == nil {
		t.Fatal("NewClickHouseSinkService returned nil")
	}
	if svc.config.PostsWorkers != 2 {
		t.Errorf("PostsWorkers = %d, want 2", svc.config.PostsWorkers)
	}
	if svc.batchCollectors == nil {
		t.Error("batchCollectors is nil")
	}
}

// ================== ClickHouseSinkService.Run Tests ==================

func TestClickHouseSinkService_Run_ContextCancel(t *testing.T) {
	cfg := ClickHouseSinkConfig{
		PostsWorkers:           1,
		InsightsWorkers:        1,
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

	deps := ClickHouseSinkDependencies{
		PagePostsConsumer:       pagePostsConsumer,
		PageInsightsConsumer:    pageInsightsConsumer,
		ProfilePostsConsumer:    profilePostsConsumer,
		ProfileInsightsConsumer: profileInsightsConsumer,
		Sink:                    nil,
		Log:                     log,
	}

	svc := NewClickHouseSinkService(cfg, deps)

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

func TestClickHouseSinkService_Run_ProcessesPosts(t *testing.T) {
	cfg := ClickHouseSinkConfig{
		PostsWorkers:           1,
		InsightsWorkers:        1,
		BatchProcessorsPerType: 0,
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
		MessageChanSize:        10,
	}

	log := logger.New("error")

	pagePostsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topicPagePosts, []byte("key"), []byte(`{"post_id":"post123"}`))
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

	deps := ClickHouseSinkDependencies{
		PagePostsConsumer:       pagePostsConsumer,
		PageInsightsConsumer:    pageInsightsConsumer,
		ProfilePostsConsumer:    profilePostsConsumer,
		ProfileInsightsConsumer: profileInsightsConsumer,
		Log:                     log,
	}

	svc := NewClickHouseSinkService(cfg, deps)

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

// ================== GetMetrics Tests ==================

func TestClickHouseSinkService_GetMetrics_Initial(t *testing.T) {
	cfg := ClickHouseSinkConfig{
		MaxBatchSize: 100,
	}

	log := logger.New("error")
	deps := ClickHouseSinkDependencies{Log: log}

	svc := NewClickHouseSinkService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.PickedPosts != 0 {
		t.Errorf("PickedPosts = %d, want 0", metrics.PickedPosts)
	}
	if metrics.PickedInsights != 0 {
		t.Errorf("PickedInsights = %d, want 0", metrics.PickedInsights)
	}
}

// ================== ClickHouseSinkMetrics Tests ==================

func TestClickHouseSinkMetrics_Struct(t *testing.T) {
	m := ClickHouseSinkMetrics{
		PickedPosts:      10,
		PickedInsights:   20,
		InsertedPosts:    8,
		InsertedInsights: 15,
	}

	if m.PickedPosts != 10 {
		t.Errorf("PickedPosts = %d, want 10", m.PickedPosts)
	}
	if m.PickedInsights != 20 {
		t.Errorf("PickedInsights = %d, want 20", m.PickedInsights)
	}
}
