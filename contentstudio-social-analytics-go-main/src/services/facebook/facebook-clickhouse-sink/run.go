package main

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ServiceConfig holds service configuration
type ServiceConfig struct {
	PostsAssetsWorkers     int
	InsightsWorkers        int
	BatchProcessorsPerType int
	MaxBatchSize           int
	BatchTimeout           time.Duration
	MessageChanSize        int
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		PostsAssetsWorkers:     postsAssetsWorkers,
		InsightsWorkers:        insightsWorkers,
		BatchProcessorsPerType: batchProcessorsPerType,
		MaxBatchSize:           maxBatchSize,
		BatchTimeout:           batchTimeout,
		MessageChanSize:        messageChanSize,
	}
}

// ServiceDependencies holds external dependencies for the service
type ServiceDependencies struct {
	Sink             *conversions.ClickHouseSink
	PostsConsumer    kafka.Consumer
	InsightsConsumer kafka.Consumer
	Log              *logger.Logger
}

// ServiceMetrics holds service metrics
type ServiceMetrics struct {
	PickedPostsAssets uint64
	PickedInsights    uint64
}

// Service represents the clickhouse sink service
type Service struct {
	config  ServiceConfig
	deps    ServiceDependencies
	metrics ServiceMetrics
	batches *BatchCollectors
}

// NewService creates a new service instance
func NewService(cfg ServiceConfig, deps ServiceDependencies) *Service {
	return &Service{
		config: cfg,
		deps:   deps,
		batches: &BatchCollectors{
			posts:         make(chan *kafkamodels.ParsedFacebookPost, cfg.MaxBatchSize*5),
			mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, cfg.MaxBatchSize*5),
			pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, cfg.MaxBatchSize*5),
			videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, cfg.MaxBatchSize*5),
			reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, cfg.MaxBatchSize*5),
		},
	}
}

// Run starts the service and blocks until context is cancelled
func (s *Service) Run(ctx context.Context) error {
	log := s.deps.Log

	// Start batch processors
	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, s.batches, s.deps.Sink, log, &batchWg, s.config.BatchProcessorsPerType)

	// Message channels
	postsAssetsMsgChan := make(chan Message, s.config.MessageChanSize)
	insightsMsgChan := make(chan Message, s.config.MessageChanSize)

	// Start workers
	var postsAssetsWg sync.WaitGroup
	for i := 0; i < s.config.PostsAssetsWorkers; i++ {
		postsAssetsWg.Add(1)
		go postsAssetsWorker(ctx, i+1, postsAssetsMsgChan, s.batches, log, &postsAssetsWg)
	}

	var insightsWg sync.WaitGroup
	for i := 0; i < s.config.InsightsWorkers; i++ {
		insightsWg.Add(1)
		go insightsWorker(ctx, i+1, insightsMsgChan, s.batches, log, &insightsWg)
	}

	// Start consumers
	postsAssetsTopics := []string{topicPosts, topicMediaAssets}
	insightsTopics := []string{topicInsights, topicVideoInsights, topicReelsInsights}

	var consumersWg sync.WaitGroup
	consumersWg.Add(2)

	// Posts/assets consumer
	go func() {
		defer consumersWg.Done()
		s.deps.PostsConsumer.Consume(ctx, postsAssetsTopics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case postsAssetsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&s.metrics.PickedPostsAssets, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()

	// Insights consumer
	go func() {
		defer consumersWg.Done()
		s.deps.InsightsConsumer.Consume(ctx, insightsTopics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case insightsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&s.metrics.PickedInsights, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	consumersWg.Wait()

	close(postsAssetsMsgChan)
	close(insightsMsgChan)

	postsAssetsWg.Wait()
	insightsWg.Wait()

	close(s.batches.posts)
	close(s.batches.mediaAssets)
	close(s.batches.pageInsights)
	close(s.batches.videoInsights)
	close(s.batches.reelsInsights)

	batchWg.Wait()

	return nil
}

// GetMetrics returns current service metrics
func (s *Service) GetMetrics() ServiceMetrics {
	return ServiceMetrics{
		PickedPostsAssets: atomic.LoadUint64(&s.metrics.PickedPostsAssets),
		PickedInsights:    atomic.LoadUint64(&s.metrics.PickedInsights),
	}
}

// GetBatchCollectors returns the batch collectors (for testing)
func (s *Service) GetBatchCollectors() *BatchCollectors {
	return s.batches
}
