package main

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ================== ParserServiceConfig Tests ==================

func TestDefaultParserServiceConfig(t *testing.T) {
	cfg := DefaultParserServiceConfig()

	if cfg.PostsParserWorkers != postsParserWorkers {
		t.Errorf("PostsParserWorkers = %d, want %d", cfg.PostsParserWorkers, postsParserWorkers)
	}
	if cfg.MediaInsightsParserWorkers != mediaInsightsParserWorkers {
		t.Errorf("MediaInsightsParserWorkers = %d, want %d", cfg.MediaInsightsParserWorkers, mediaInsightsParserWorkers)
	}
	if cfg.PostsPublisherWorkers != postsPublisherWorkers {
		t.Errorf("PostsPublisherWorkers = %d, want %d", cfg.PostsPublisherWorkers, postsPublisherWorkers)
	}
	if cfg.MediaInsightsPublisherWorkers != mediaInsightsPublisherWorkers {
		t.Errorf("MediaInsightsPublisherWorkers = %d, want %d", cfg.MediaInsightsPublisherWorkers, mediaInsightsPublisherWorkers)
	}
	if cfg.ParseChanSize != parseChanSize {
		t.Errorf("ParseChanSize = %d, want %d", cfg.ParseChanSize, parseChanSize)
	}
	if cfg.PublishChanSize != publishChanSize {
		t.Errorf("PublishChanSize = %d, want %d", cfg.PublishChanSize, publishChanSize)
	}
}

// ================== NewParserService Tests ==================

func TestNewParserService(t *testing.T) {
	cfg := ParserServiceConfig{
		PostsParserWorkers:         2,
		MediaInsightsParserWorkers: 2,
		PostsPublisherWorkers:      2,
		MediaInsightsPublisherWorkers: 2,
		ParseChanSize:              10,
		PublishChanSize:            10,
	}

	log := logger.New("error")
	deps := ParserServiceDependencies{Log: log}

	svc := NewParserService(cfg, deps)

	if svc == nil {
		t.Fatal("NewParserService returned nil")
	}
	if svc.config.PostsParserWorkers != 2 {
		t.Errorf("PostsParserWorkers = %d, want 2", svc.config.PostsParserWorkers)
	}
}

// ================== ParserService.Run Tests ==================

func TestParserService_Run_ContextCancel(t *testing.T) {
	cfg := ParserServiceConfig{
		PostsParserWorkers:         1,
		MediaInsightsParserWorkers: 1,
		PostsPublisherWorkers:      1,
		MediaInsightsPublisherWorkers: 1,
		ParseChanSize:              10,
		PublishChanSize:            10,
	}

	log := logger.New("error")

	postsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	miConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	producer := &kafka.MockProducer{}

	deps := ParserServiceDependencies{
		Producer:      producer,
		PostsConsumer: postsConsumer,
		MIConsumer:    miConsumer,
		Log:           log,
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

func TestParserService_Run_ProcessesPosts(t *testing.T) {
	cfg := ParserServiceConfig{
		PostsParserWorkers:         1,
		MediaInsightsParserWorkers: 1,
		PostsPublisherWorkers:      1,
		MediaInsightsPublisherWorkers: 1,
		ParseChanSize:              10,
		PublishChanSize:            10,
	}

	log := logger.New("error")

	rawPost := kafkamodels.RawFacebookPost{ID: "post123"}
	postJSON, _ := json.Marshal(rawPost)

	postsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, rawPostsTopic, []byte("ws_page_key"), postJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	miConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	deps := ParserServiceDependencies{
		Producer:      producer,
		PostsConsumer: postsConsumer,
		MIConsumer:    miConsumer,
		Log:           log,
	}

	svc := NewParserService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.PickedPosts < 1 {
		t.Errorf("PickedPosts = %d, want >= 1", metrics.PickedPosts)
	}

	cancel()
	<-done
}

func TestParserService_Run_ProcessesVideosAndInsights(t *testing.T) {
	cfg := ParserServiceConfig{
		PostsParserWorkers:         1,
		MediaInsightsParserWorkers: 1,
		PostsPublisherWorkers:      1,
		MediaInsightsPublisherWorkers: 1,
		ParseChanSize:              10,
		PublishChanSize:            10,
	}

	log := logger.New("error")

	rawVideo := kafkamodels.RawFacebookVideo{ID: "video123"}
	videoJSON, _ := json.Marshal(rawVideo)

	rawInsights := kafkamodels.RawFacebookInsights{PageID: "page123"}
	insightsJSON, _ := json.Marshal(rawInsights)

	postsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			<-ctx.Done()
			return ctx.Err()
		},
	}
	miConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, rawVideosTopic, []byte("key1"), videoJSON)
			handler(ctx, rawInsightsTopic, []byte("key2"), insightsJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	producer := &kafka.MockProducer{
		ProduceFunc: func(ctx context.Context, topic string, key, value []byte) error {
			return nil
		},
	}

	deps := ParserServiceDependencies{
		Producer:      producer,
		PostsConsumer: postsConsumer,
		MIConsumer:    miConsumer,
		Log:           log,
	}

	svc := NewParserService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	time.Sleep(200 * time.Millisecond)

	metrics := svc.GetMetrics()
	if metrics.PickedVideos < 1 {
		t.Errorf("PickedVideos = %d, want >= 1", metrics.PickedVideos)
	}
	if metrics.PickedInsights < 1 {
		t.Errorf("PickedInsights = %d, want >= 1", metrics.PickedInsights)
	}

	cancel()
	<-done
}

// ================== GetMetrics Tests ==================

func TestParserService_GetMetrics_Initial(t *testing.T) {
	cfg := ParserServiceConfig{}
	log := logger.New("error")
	deps := ParserServiceDependencies{Log: log}

	svc := NewParserService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.PickedPosts != 0 {
		t.Errorf("PickedPosts = %d, want 0", metrics.PickedPosts)
	}
	if metrics.PickedVideos != 0 {
		t.Errorf("PickedVideos = %d, want 0", metrics.PickedVideos)
	}
	if metrics.PickedInsights != 0 {
		t.Errorf("PickedInsights = %d, want 0", metrics.PickedInsights)
	}
	if metrics.PublishedPosts != 0 {
		t.Errorf("PublishedPosts = %d, want 0", metrics.PublishedPosts)
	}
}

// ================== ParserServiceMetrics Tests ==================

func TestParserServiceMetrics_Struct(t *testing.T) {
	m := ParserServiceMetrics{
		PickedPosts:       10,
		PickedVideos:      20,
		PickedInsights:    30,
		PublishedPosts:    5,
		PublishedAssets:   15,
		PublishedVideoIns: 10,
		PublishedReelsIns: 5,
		PublishedPageIns:  25,
	}

	if m.PickedPosts != 10 {
		t.Errorf("PickedPosts = %d, want 10", m.PickedPosts)
	}
	if m.PublishedPageIns != 25 {
		t.Errorf("PublishedPageIns = %d, want 25", m.PublishedPageIns)
	}
}
