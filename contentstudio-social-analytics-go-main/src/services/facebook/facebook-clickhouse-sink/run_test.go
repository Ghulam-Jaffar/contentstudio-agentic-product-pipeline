package main

import (
	"context"
	"encoding/json"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/rs/zerolog"
)

// ================== ServiceConfig Tests ==================

func TestDefaultServiceConfig(t *testing.T) {
	cfg := DefaultServiceConfig()

	if cfg.PostsAssetsWorkers != postsAssetsWorkers {
		t.Errorf("PostsAssetsWorkers = %d, want %d", cfg.PostsAssetsWorkers, postsAssetsWorkers)
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
	if cfg.MessageChanSize != messageChanSize {
		t.Errorf("MessageChanSize = %d, want %d", cfg.MessageChanSize, messageChanSize)
	}
}

// ================== NewService Tests ==================

func TestNewService(t *testing.T) {
	cfg := ServiceConfig{
		PostsAssetsWorkers:     2,
		InsightsWorkers:        2,
		BatchProcessorsPerType: 1,
		MaxBatchSize:           100,
		BatchTimeout:           time.Second,
		MessageChanSize:        50,
	}

	log := logger.New("error")
	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, &conversions.MockClickHouseClient{})

	deps := ServiceDependencies{
		Sink: sink,
		Log:  log,
	}

	svc := NewService(cfg, deps)

	if svc == nil {
		t.Fatal("NewService returned nil")
	}
	if svc.config.PostsAssetsWorkers != 2 {
		t.Errorf("PostsAssetsWorkers = %d, want 2", svc.config.PostsAssetsWorkers)
	}
	if svc.batches == nil {
		t.Fatal("batches is nil")
	}
	if cap(svc.batches.posts) != 500 {
		t.Errorf("posts channel capacity = %d, want 500", cap(svc.batches.posts))
	}
}

// ================== Service.Run Tests ==================

func TestService_Run_ContextCancel(t *testing.T) {
	cfg := ServiceConfig{
		PostsAssetsWorkers:     1,
		InsightsWorkers:        1,
		BatchProcessorsPerType: 1,
		MaxBatchSize:           10,
		BatchTimeout:           100 * time.Millisecond,
		MessageChanSize:        10,
	}

	log := logger.New("error")
	zlog := zerolog.Nop()
	sink := conversions.NewClickHouseSinkWithClient(&zlog, &conversions.MockClickHouseClient{})

	postsConsumer := &kafka.MockConsumer{
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

	deps := ServiceDependencies{
		Sink:             sink,
		PostsConsumer:    postsConsumer,
		InsightsConsumer: insightsConsumer,
		Log:              log,
	}

	svc := NewService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	// Cancel context
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

func TestService_Run_ProcessesMessages(t *testing.T) {
	cfg := ServiceConfig{
		PostsAssetsWorkers:     1,
		InsightsWorkers:        1,
		BatchProcessorsPerType: 1,
		MaxBatchSize:           10,
		BatchTimeout:           100 * time.Millisecond,
		MessageChanSize:        10,
	}

	log := logger.New("error")
	zlog := zerolog.Nop()

	var insertedPosts int32
	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			atomic.AddInt32(&insertedPosts, int32(len(posts)))
			return nil
		},
	}
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	post := kafkamodels.ParsedFacebookPost{PostID: "post123", PageID: "page456"}
	postJSON, _ := json.Marshal(post)

	postsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			// Send one message
			handler(ctx, topicPosts, []byte("key"), postJSON)
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

	deps := ServiceDependencies{
		Sink:             sink,
		PostsConsumer:    postsConsumer,
		InsightsConsumer: insightsConsumer,
		Log:              log,
	}

	svc := NewService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	// Wait for message processing
	time.Sleep(300 * time.Millisecond)

	// Check metrics
	metrics := svc.GetMetrics()
	if metrics.PickedPostsAssets < 1 {
		t.Errorf("PickedPostsAssets = %d, want >= 1", metrics.PickedPostsAssets)
	}

	cancel()
	<-done
}

func TestService_Run_ProcessesAllMessageTypes(t *testing.T) {
	cfg := ServiceConfig{
		PostsAssetsWorkers:     1,
		InsightsWorkers:        1,
		BatchProcessorsPerType: 1,
		MaxBatchSize:           10,
		BatchTimeout:           100 * time.Millisecond,
		MessageChanSize:        10,
	}

	log := logger.New("error")
	zlog := zerolog.Nop()

	var insertedPosts, insertedAssets, insertedPageIns, insertedVideoIns, insertedReelsIns int32
	mockClient := &conversions.MockClickHouseClient{
		BulkInsertPostsFunc: func(ctx context.Context, posts []*clickhousemodels.FacebookPosts) error {
			atomic.AddInt32(&insertedPosts, int32(len(posts)))
			return nil
		},
		BulkInsertMediaAssetsFunc: func(ctx context.Context, assets []*clickhousemodels.FacebookMediaAssets) error {
			atomic.AddInt32(&insertedAssets, int32(len(assets)))
			return nil
		},
		BulkInsertInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookInsights) error {
			atomic.AddInt32(&insertedPageIns, int32(len(insights)))
			return nil
		},
		BulkInsertVideoInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookVideoInsights) error {
			atomic.AddInt32(&insertedVideoIns, int32(len(insights)))
			return nil
		},
		BulkInsertReelsInsightsFunc: func(ctx context.Context, insights []*clickhousemodels.FacebookReelsInsights) error {
			atomic.AddInt32(&insertedReelsIns, int32(len(insights)))
			return nil
		},
	}
	sink := conversions.NewClickHouseSinkWithClient(&zlog, mockClient)

	// Prepare messages
	post := kafkamodels.ParsedFacebookPost{PostID: "post1"}
	postJSON, _ := json.Marshal(post)

	asset := kafkamodels.ParsedFacebookMediaAsset{MediaID: "media1"}
	assetJSON, _ := json.Marshal(asset)

	pageIns := []*kafkamodels.ParsedFacebookInsights{{PageID: "page1"}}
	pageInsJSON, _ := json.Marshal(pageIns)

	videoIns := kafkamodels.ParsedFacebookVideoInsights{VideoID: "video1"}
	videoInsJSON, _ := json.Marshal(videoIns)

	reelsIns := kafkamodels.ParsedFacebookReelsInsights{PostID: "reel1"}
	reelsInsJSON, _ := json.Marshal(reelsIns)

	postsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topicPosts, []byte("k1"), postJSON)
			handler(ctx, topicMediaAssets, []byte("k2"), assetJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}
	insightsConsumer := &kafka.MockConsumer{
		ConsumeFunc: func(ctx context.Context, topics []string, handler kafka.MessageHandler) error {
			handler(ctx, topicInsights, []byte("k3"), pageInsJSON)
			handler(ctx, topicVideoInsights, []byte("k4"), videoInsJSON)
			handler(ctx, topicReelsInsights, []byte("k5"), reelsInsJSON)
			<-ctx.Done()
			return ctx.Err()
		},
	}

	deps := ServiceDependencies{
		Sink:             sink,
		PostsConsumer:    postsConsumer,
		InsightsConsumer: insightsConsumer,
		Log:              log,
	}

	svc := NewService(cfg, deps)

	ctx, cancel := context.WithCancel(context.Background())

	done := make(chan error)
	go func() {
		done <- svc.Run(ctx)
	}()

	// Wait for processing and batch flush
	time.Sleep(400 * time.Millisecond)

	cancel()
	<-done

	// Verify metrics
	metrics := svc.GetMetrics()
	if metrics.PickedPostsAssets < 2 {
		t.Errorf("PickedPostsAssets = %d, want >= 2", metrics.PickedPostsAssets)
	}
	if metrics.PickedInsights < 3 {
		t.Errorf("PickedInsights = %d, want >= 3", metrics.PickedInsights)
	}
}

// ================== GetBatchCollectors Tests ==================

func TestService_GetBatchCollectors(t *testing.T) {
	cfg := ServiceConfig{MaxBatchSize: 10}
	log := logger.New("error")

	deps := ServiceDependencies{Log: log}
	svc := NewService(cfg, deps)

	batches := svc.GetBatchCollectors()
	if batches == nil {
		t.Fatal("GetBatchCollectors returned nil")
	}
	if batches.posts == nil {
		t.Error("posts channel is nil")
	}
}

// ================== GetMetrics Tests ==================

func TestService_GetMetrics_Initial(t *testing.T) {
	cfg := ServiceConfig{MaxBatchSize: 10}
	log := logger.New("error")

	deps := ServiceDependencies{Log: log}
	svc := NewService(cfg, deps)

	metrics := svc.GetMetrics()
	if metrics.PickedPostsAssets != 0 {
		t.Errorf("PickedPostsAssets = %d, want 0", metrics.PickedPostsAssets)
	}
	if metrics.PickedInsights != 0 {
		t.Errorf("PickedInsights = %d, want 0", metrics.PickedInsights)
	}
}
