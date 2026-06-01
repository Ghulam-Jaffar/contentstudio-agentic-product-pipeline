package main

import (
	"context"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ================== ParserConfig Tests ==================

func TestDefaultParserConfig(t *testing.T) {
	cfg := DefaultParserConfig()

	if cfg.MediaParserWorkers != mediaParserWorkers {
		t.Errorf("MediaParserWorkers = %d, want %d", cfg.MediaParserWorkers, mediaParserWorkers)
	}
	if cfg.InsightsParserWorkers != insightsParserWorkers {
		t.Errorf("InsightsParserWorkers = %d, want %d", cfg.InsightsParserWorkers, insightsParserWorkers)
	}
	if cfg.MediaPublishWorkers != mediaPublishWorkers {
		t.Errorf("MediaPublishWorkers = %d, want %d", cfg.MediaPublishWorkers, mediaPublishWorkers)
	}
	if cfg.InsightsPublishWorkers != insightsPublishWorkers {
		t.Errorf("InsightsPublishWorkers = %d, want %d", cfg.InsightsPublishWorkers, insightsPublishWorkers)
	}
	if cfg.MessageChanSize != messageChanSize {
		t.Errorf("MessageChanSize = %d, want %d", cfg.MessageChanSize, messageChanSize)
	}
}

// ================== NewParserService Tests ==================

func TestNewParserService(t *testing.T) {
	cfg := ParserConfig{
		MediaParserWorkers:     2,
		InsightsParserWorkers:  2,
		MediaPublishWorkers:    2,
		InsightsPublishWorkers: 2,
		MessageChanSize:        10,
	}

	log := logger.New("error")
	deps := ParserDependencies{Log: log}

	svc := NewParserService(cfg, deps)

	if svc == nil {
		t.Fatal("NewParserService returned nil")
	}
	if svc.config.MediaParserWorkers != 2 {
		t.Errorf("MediaParserWorkers = %d, want 2", svc.config.MediaParserWorkers)
	}
}

// ================== ParserService.Run Tests ==================

func TestParserService_Run_ContextCancel(t *testing.T) {
	cfg := ParserConfig{
		MediaParserWorkers:     1,
		InsightsParserWorkers:  1,
		MediaPublishWorkers:    1,
		InsightsPublishWorkers: 1,
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
	producer := &kafka.MockProducer{}

	deps := ParserDependencies{
		MediaConsumer:    mediaConsumer,
		InsightsConsumer: insightsConsumer,
		Producer:         producer,
		Log:              log,
	}

	svc := NewParserService(cfg, deps)

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

func TestParserService_Run_ProcessesMedia(t *testing.T) {
	cfg := ParserConfig{
		MediaParserWorkers:     1,
		InsightsParserWorkers:  1,
		MediaPublishWorkers:    1,
		InsightsPublishWorkers: 1,
		MessageChanSize:        10,
	}

	log := logger.New("error")

	mediaConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
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
	producer := &kafka.MockProducer{}

	deps := ParserDependencies{
		MediaConsumer:    mediaConsumer,
		InsightsConsumer: insightsConsumer,
		Producer:         producer,
		Log:              log,
	}

	svc := NewParserService(cfg, deps)

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

func TestParserService_Run_ProcessesInsights(t *testing.T) {
	cfg := ParserConfig{
		MediaParserWorkers:     1,
		InsightsParserWorkers:  1,
		MediaPublishWorkers:    1,
		InsightsPublishWorkers: 1,
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
			handler(ctx, insightsTopic, []byte("ws_ig123_insights"), []byte(`{"insights":{}}`))
			<-ctx.Done()
			return ctx.Err()
		},
	}
	producer := &kafka.MockProducer{}

	deps := ParserDependencies{
		MediaConsumer:    mediaConsumer,
		InsightsConsumer: insightsConsumer,
		Producer:         producer,
		Log:              log,
	}

	svc := NewParserService(cfg, deps)

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

func TestParserService_GetMetrics_Initial(t *testing.T) {
	cfg := ParserConfig{}
	log := logger.New("error")
	deps := ParserDependencies{Log: log}

	svc := NewParserService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.PickedMedia != 0 {
		t.Errorf("PickedMedia = %d, want 0", metrics.PickedMedia)
	}
	if metrics.PickedInsights != 0 {
		t.Errorf("PickedInsights = %d, want 0", metrics.PickedInsights)
	}
	if metrics.PublishedMedia != 0 {
		t.Errorf("PublishedMedia = %d, want 0", metrics.PublishedMedia)
	}
	if metrics.PublishedInsights != 0 {
		t.Errorf("PublishedInsights = %d, want 0", metrics.PublishedInsights)
	}
}

// ================== ParserMetrics Tests ==================

func TestParserMetrics_Struct(t *testing.T) {
	m := ParserMetrics{
		PickedMedia:      10,
		PickedInsights:   20,
		PublishedMedia:   8,
		PublishedInsights: 15,
	}

	if m.PickedMedia != 10 {
		t.Errorf("PickedMedia = %d, want 10", m.PickedMedia)
	}
	if m.PickedInsights != 20 {
		t.Errorf("PickedInsights = %d, want 20", m.PickedInsights)
	}
	if m.PublishedMedia != 8 {
		t.Errorf("PublishedMedia = %d, want 8", m.PublishedMedia)
	}
	if m.PublishedInsights != 15 {
		t.Errorf("PublishedInsights = %d, want 15", m.PublishedInsights)
	}
}
