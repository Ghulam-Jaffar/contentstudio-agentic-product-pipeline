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

// AnalyticsSinkConfig holds service configuration
type AnalyticsSinkConfig struct {
	PostsParserWorkers     int
	InsightsParserWorkers  int
	BatchProcessorsPerType int
	MaxBatchSize           int
	BatchTimeout           time.Duration
	MessageChanSize        int
}

// DefaultAnalyticsSinkConfig returns default configuration
func DefaultAnalyticsSinkConfig() AnalyticsSinkConfig {
	return AnalyticsSinkConfig{
		PostsParserWorkers:     postsParserWorkers,
		InsightsParserWorkers:  insightsParserWorkers,
		BatchProcessorsPerType: batchProcessorsPerType,
		MaxBatchSize:           maxBatchSize,
		BatchTimeout:           batchTimeout,
		MessageChanSize:        messageChanSize,
	}
}

// AnalyticsSinkDependencies holds external dependencies
type AnalyticsSinkDependencies struct {
	PagePostsConsumer       kafka.Consumer
	PageInsightsConsumer    kafka.Consumer
	ProfilePostsConsumer    kafka.Consumer
	ProfileInsightsConsumer kafka.Consumer
	Sink                    *conversions.ClickHouseSink
	Log                     *logger.Logger
}

// AnalyticsSinkMetrics holds service metrics
type AnalyticsSinkMetrics struct {
	PickedPosts      uint64
	PickedInsights   uint64
	ParsedPosts      uint64
	ParsedInsights   uint64
	InsertedPosts    uint64
	InsertedInsights uint64
}

// AnalyticsSinkService represents the analytics sink service
type AnalyticsSinkService struct {
	config          AnalyticsSinkConfig
	deps            AnalyticsSinkDependencies
	metrics         AnalyticsSinkMetrics
	batchCollectors *BatchCollectors
}

// NewAnalyticsSinkService creates a new analytics sink service
func NewAnalyticsSinkService(cfg AnalyticsSinkConfig, deps AnalyticsSinkDependencies) *AnalyticsSinkService {
	return &AnalyticsSinkService{
		config: cfg,
		deps:   deps,
		batchCollectors: &BatchCollectors{
			posts:    make(chan *kafkamodels.ParsedLinkedinPost, cfg.MaxBatchSize*5),
			insights: make(chan *kafkamodels.ParsedLinkedinInsights, cfg.MaxBatchSize*5),
		},
	}
}

// Run starts the analytics sink service
func (s *AnalyticsSinkService) Run(ctx context.Context) error {
	log := s.deps.Log

	postsMsgChan := make(chan RawMessage, s.config.MessageChanSize)
	insightsMsgChan := make(chan RawMessage, s.config.MessageChanSize)

	// Start batch processors
	var batchWg sync.WaitGroup
	if s.deps.Sink != nil && s.config.BatchProcessorsPerType > 0 {
		startBatchProcessors(ctx, s.batchCollectors, s.deps.Sink, log, &batchWg, s.config.BatchProcessorsPerType)
	}

	// Start parser workers
	var postsWg sync.WaitGroup
	for i := 0; i < s.config.PostsParserWorkers; i++ {
		postsWg.Add(1)
		go postsParserWorker(ctx, i+1, postsMsgChan, s.batchCollectors, &s.metrics.ParsedPosts, log, &postsWg)
	}

	var insightsWg sync.WaitGroup
	for i := 0; i < s.config.InsightsParserWorkers; i++ {
		insightsWg.Add(1)
		go insightsParserWorker(ctx, i+1, insightsMsgChan, s.batchCollectors, &s.metrics.ParsedInsights, log, &insightsWg)
	}

	// Start consumers
	var consumerWg sync.WaitGroup
	consumerWg.Add(4)

	go func() {
		defer consumerWg.Done()
		s.deps.PagePostsConsumer.Consume(ctx, []string{topicRawPagePosts}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&s.metrics.PickedPosts, 1)
			select {
			case postsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	go func() {
		defer consumerWg.Done()
		s.deps.PageInsightsConsumer.Consume(ctx, []string{topicRawPageInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&s.metrics.PickedInsights, 1)
			select {
			case insightsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	go func() {
		defer consumerWg.Done()
		s.deps.ProfilePostsConsumer.Consume(ctx, []string{topicRawProfilePosts}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&s.metrics.PickedPosts, 1)
			select {
			case postsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	go func() {
		defer consumerWg.Done()
		s.deps.ProfileInsightsConsumer.Consume(ctx, []string{topicRawProfileInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&s.metrics.PickedInsights, 1)
			select {
			case insightsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	<-ctx.Done()

	consumerWg.Wait()
	close(postsMsgChan)
	close(insightsMsgChan)

	postsWg.Wait()
	insightsWg.Wait()

	close(s.batchCollectors.posts)
	close(s.batchCollectors.insights)

	batchWg.Wait()

	return nil
}

// GetMetrics returns current service metrics
func (s *AnalyticsSinkService) GetMetrics() AnalyticsSinkMetrics {
	return AnalyticsSinkMetrics{
		PickedPosts:      atomic.LoadUint64(&s.metrics.PickedPosts),
		PickedInsights:   atomic.LoadUint64(&s.metrics.PickedInsights),
		ParsedPosts:      atomic.LoadUint64(&s.metrics.ParsedPosts),
		ParsedInsights:   atomic.LoadUint64(&s.metrics.ParsedInsights),
		InsertedPosts:    atomic.LoadUint64(&s.metrics.InsertedPosts),
		InsertedInsights: atomic.LoadUint64(&s.metrics.InsertedInsights),
	}
}
