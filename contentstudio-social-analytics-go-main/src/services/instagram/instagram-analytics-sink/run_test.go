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

	if cfg.MediaParserWorkers != mediaParserWorkers {
		t.Errorf("MediaParserWorkers = %d, want %d", cfg.MediaParserWorkers, mediaParserWorkers)
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
		MediaParserWorkers:     2,
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
	if svc.config.MediaParserWorkers != 2 {
		t.Errorf("MediaParserWorkers = %d, want 2", svc.config.MediaParserWorkers)
	}
	if svc.batchCollectors == nil {
		t.Error("batchCollectors is nil")
	}
	if svc.batchCollectors.posts == nil {
		t.Error("posts channel is nil")
	}
	if svc.batchCollectors.insights == nil {
		t.Error("insights channel is nil")
	}
}

// ================== AnalyticsSinkService.Run Tests ==================

func TestAnalyticsSinkService_Run_ContextCancel(t *testing.T) {
	cfg := AnalyticsSinkConfig{
		MediaParserWorkers:     1,
		InsightsParserWorkers:  1,
		BatchProcessorsPerType: 0, // Don't start batch processors
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
		MessageChanSize:        10,
	}

	log := logger.New("error")

	mediaConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	insightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := AnalyticsSinkDependencies{
		MediaConsumer:    mediaConsumer,
		InsightsConsumer: insightsConsumer,
		Sink:             nil,
		Log:              log,
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

func TestAnalyticsSinkService_Run_ProcessesMedia(t *testing.T) {
	cfg := AnalyticsSinkConfig{
		MediaParserWorkers:     1,
		InsightsParserWorkers:  1,
		BatchProcessorsPerType: 0, // Don't start batch processors
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
		MessageChanSize:        10,
	}

	log := logger.New("error")

	mediaConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send valid media message
			handler(ctx, mediaTopic, []byte("ws_ig123_media456"), []byte(`{"id":"media123"}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}
	insightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := AnalyticsSinkDependencies{
		MediaConsumer:    mediaConsumer,
		InsightsConsumer: insightsConsumer,
		Sink:             nil,
		Log:              log,
	}

	svc := NewAnalyticsSinkService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(100 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.PickedMedia < 1 {
		t.Errorf("PickedMedia = %d, want >= 1", metrics.PickedMedia)
	}

	cancel()
	<-done
}

func TestAnalyticsSinkService_Run_ProcessesInsights(t *testing.T) {
	cfg := AnalyticsSinkConfig{
		MediaParserWorkers:     1,
		InsightsParserWorkers:  1,
		BatchProcessorsPerType: 0, // Don't start batch processors
		MaxBatchSize:           100,
		BatchTimeout:           5 * time.Second,
		MessageChanSize:        10,
	}

	log := logger.New("error")

	mediaConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	insightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send valid insights message
			handler(ctx, insightsTopic, []byte("ws_ig123_insights"), []byte(`{"insights":{}}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := AnalyticsSinkDependencies{
		MediaConsumer:    mediaConsumer,
		InsightsConsumer: insightsConsumer,
		Sink:             nil,
		Log:              log,
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
	if metrics.PickedMedia != 0 {
		t.Errorf("PickedMedia = %d, want 0", metrics.PickedMedia)
	}
	if metrics.PickedInsights != 0 {
		t.Errorf("PickedInsights = %d, want 0", metrics.PickedInsights)
	}
	if metrics.ParsedPosts != 0 {
		t.Errorf("ParsedPosts = %d, want 0", metrics.ParsedPosts)
	}
	if metrics.ParsedInsights != 0 {
		t.Errorf("ParsedInsights = %d, want 0", metrics.ParsedInsights)
	}
	if metrics.InsertedPosts != 0 {
		t.Errorf("InsertedPosts = %d, want 0", metrics.InsertedPosts)
	}
	if metrics.InsertedInsights != 0 {
		t.Errorf("InsertedInsights = %d, want 0", metrics.InsertedInsights)
	}
}

// ================== GetBatchCollectors Tests ==================

func TestAnalyticsSinkService_GetBatchCollectors(t *testing.T) {
	cfg := AnalyticsSinkConfig{
		MaxBatchSize: 100,
	}

	log := logger.New("error")
	deps := AnalyticsSinkDependencies{Log: log}

	svc := NewAnalyticsSinkService(cfg, deps)

	bc := svc.GetBatchCollectors()
	if bc == nil {
		t.Fatal("GetBatchCollectors returned nil")
	}
	if bc.posts == nil {
		t.Error("posts channel is nil")
	}
	if bc.insights == nil {
		t.Error("insights channel is nil")
	}
}

// ================== AnalyticsSinkMetrics Tests ==================

func TestAnalyticsSinkMetrics_Struct(t *testing.T) {
	m := AnalyticsSinkMetrics{
		PickedMedia:      10,
		PickedInsights:   20,
		ParsedPosts:      8,
		ParsedInsights:   15,
		InsertedPosts:    5,
		InsertedInsights: 12,
	}

	if m.PickedMedia != 10 {
		t.Errorf("PickedMedia = %d, want 10", m.PickedMedia)
	}
	if m.PickedInsights != 20 {
		t.Errorf("PickedInsights = %d, want 20", m.PickedInsights)
	}
	if m.ParsedPosts != 8 {
		t.Errorf("ParsedPosts = %d, want 8", m.ParsedPosts)
	}
	if m.ParsedInsights != 15 {
		t.Errorf("ParsedInsights = %d, want 15", m.ParsedInsights)
	}
	if m.InsertedPosts != 5 {
		t.Errorf("InsertedPosts = %d, want 5", m.InsertedPosts)
	}
	if m.InsertedInsights != 12 {
		t.Errorf("InsertedInsights = %d, want 12", m.InsertedInsights)
	}
}
