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

// ClickHouseSinkConfig holds service configuration
type ClickHouseSinkConfig struct {
	PostsWorkers           int
	InsightsWorkers        int
	BatchProcessorsPerType int
	MaxBatchSize           int
	BatchTimeout           time.Duration
	MessageChanSize        int
}

// DefaultClickHouseSinkConfig returns default configuration
func DefaultClickHouseSinkConfig() ClickHouseSinkConfig {
	return ClickHouseSinkConfig{
		PostsWorkers:           postsWorkers,
		InsightsWorkers:        insightsWorkers,
		BatchProcessorsPerType: batchProcessorsPerType,
		MaxBatchSize:           maxBatchSize,
		BatchTimeout:           batchTimeout,
		MessageChanSize:        messageChanSize,
	}
}

// ClickHouseSinkDependencies holds external dependencies
type ClickHouseSinkDependencies struct {
	PagePostsConsumer       kafka.Consumer
	PageInsightsConsumer    kafka.Consumer
	ProfilePostsConsumer    kafka.Consumer
	ProfileInsightsConsumer kafka.Consumer
	Sink                    *conversions.ClickHouseSink
	Log                     *logger.Logger
}

// ClickHouseSinkMetrics holds service metrics
type ClickHouseSinkMetrics struct {
	PickedPosts      uint64
	PickedInsights   uint64
	InsertedPosts    uint64
	InsertedInsights uint64
}

// ClickHouseSinkService represents the ClickHouse sink service
type ClickHouseSinkService struct {
	config          ClickHouseSinkConfig
	deps            ClickHouseSinkDependencies
	metrics         ClickHouseSinkMetrics
	batchCollectors *BatchCollectors
}

// NewClickHouseSinkService creates a new ClickHouse sink service
func NewClickHouseSinkService(cfg ClickHouseSinkConfig, deps ClickHouseSinkDependencies) *ClickHouseSinkService {
	return &ClickHouseSinkService{
		config: cfg,
		deps:   deps,
		batchCollectors: &BatchCollectors{
			posts:    make(chan *kafkamodels.ParsedLinkedinPost, cfg.MaxBatchSize*5),
			insights: make(chan *kafkamodels.ParsedLinkedinInsights, cfg.MaxBatchSize*5),
		},
	}
}

// Run starts the ClickHouse sink service
func (s *ClickHouseSinkService) Run(ctx context.Context) error {
	log := s.deps.Log

	postsMsgChan := make(chan Message, s.config.MessageChanSize)
	insightsMsgChan := make(chan Message, s.config.MessageChanSize)

	// Start batch processors
	var batchWg sync.WaitGroup
	if s.deps.Sink != nil && s.config.BatchProcessorsPerType > 0 {
		startBatchProcessors(ctx, s.batchCollectors, s.deps.Sink, log, &batchWg, s.config.BatchProcessorsPerType)
	}

	// Start workers
	var postsWg sync.WaitGroup
	for i := 0; i < s.config.PostsWorkers; i++ {
		postsWg.Add(1)
		go postsWorker(ctx, i+1, postsMsgChan, s.batchCollectors, log, &postsWg)
	}

	var insightsWg sync.WaitGroup
	for i := 0; i < s.config.InsightsWorkers; i++ {
		insightsWg.Add(1)
		go insightsWorker(ctx, i+1, insightsMsgChan, s.batchCollectors, log, &insightsWg)
	}

	// Start consumers
	var consumerWg sync.WaitGroup
	consumerWg.Add(4)

	go func() {
		defer consumerWg.Done()
		s.deps.PagePostsConsumer.Consume(ctx, []string{topicPagePosts}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&s.metrics.PickedPosts, 1)
			select {
			case postsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	go func() {
		defer consumerWg.Done()
		s.deps.PageInsightsConsumer.Consume(ctx, []string{topicPageInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&s.metrics.PickedInsights, 1)
			select {
			case insightsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	go func() {
		defer consumerWg.Done()
		s.deps.ProfilePostsConsumer.Consume(ctx, []string{topicProfilePosts}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&s.metrics.PickedPosts, 1)
			select {
			case postsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	go func() {
		defer consumerWg.Done()
		s.deps.ProfileInsightsConsumer.Consume(ctx, []string{topicProfileInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&s.metrics.PickedInsights, 1)
			select {
			case insightsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
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
func (s *ClickHouseSinkService) GetMetrics() ClickHouseSinkMetrics {
	return ClickHouseSinkMetrics{
		PickedPosts:      atomic.LoadUint64(&s.metrics.PickedPosts),
		PickedInsights:   atomic.LoadUint64(&s.metrics.PickedInsights),
		InsertedPosts:    atomic.LoadUint64(&s.metrics.InsertedPosts),
		InsertedInsights: atomic.LoadUint64(&s.metrics.InsertedInsights),
	}
}
