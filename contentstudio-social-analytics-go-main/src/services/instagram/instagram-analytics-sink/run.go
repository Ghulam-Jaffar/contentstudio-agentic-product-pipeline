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
	MediaParserWorkers     int
	InsightsParserWorkers  int
	BatchProcessorsPerType int
	MaxBatchSize           int
	BatchTimeout           time.Duration
	MessageChanSize        int
}

// DefaultAnalyticsSinkConfig returns default configuration
func DefaultAnalyticsSinkConfig() AnalyticsSinkConfig {
	return AnalyticsSinkConfig{
		MediaParserWorkers:     mediaParserWorkers,
		InsightsParserWorkers:  insightsParserWorkers,
		BatchProcessorsPerType: batchProcessorsPerType,
		MaxBatchSize:           maxBatchSize,
		BatchTimeout:           batchTimeout,
		MessageChanSize:        messageChanSize,
	}
}

// AnalyticsSinkDependencies holds external dependencies
type AnalyticsSinkDependencies struct {
	MediaConsumer    kafka.Consumer
	InsightsConsumer kafka.Consumer
	Sink             *conversions.ClickHouseSink
	Log              *logger.Logger
}

// AnalyticsSinkMetrics holds service metrics
type AnalyticsSinkMetrics struct {
	PickedMedia      uint64
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
			posts:    make(chan *kafkamodels.ParsedInstagramPost, cfg.MaxBatchSize*5),
			insights: make(chan *kafkamodels.ParsedInstagramInsight, cfg.MaxBatchSize*5),
		},
	}
}

// Run starts the analytics sink service
func (s *AnalyticsSinkService) Run(ctx context.Context) error {
	log := s.deps.Log

	mediaJobs := make(chan ParseJob, s.config.MessageChanSize)
	insightJobs := make(chan ParseJob, s.config.MessageChanSize)

	// Start batch processors
	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, s.batchCollectors, s.deps.Sink, log, &batchWg, s.config.BatchProcessorsPerType, &s.metrics.InsertedPosts, &s.metrics.InsertedInsights)

	// Start parser pools
	var wgParsers sync.WaitGroup
	for i := 0; i < s.config.MediaParserWorkers; i++ {
		wgParsers.Add(1)
		go mediaParser(ctx, &wgParsers, i, mediaJobs, s.batchCollectors.posts, &s.metrics.ParsedPosts, log)
	}
	for i := 0; i < s.config.InsightsParserWorkers; i++ {
		wgParsers.Add(1)
		go insightsParser(ctx, &wgParsers, i, insightJobs, s.batchCollectors.insights, &s.metrics.ParsedInsights, log)
	}

	// Start consumers
	var wgConsumers sync.WaitGroup
	wgConsumers.Add(2)

	go func() {
		defer wgConsumers.Done()
		s.deps.MediaConsumer.Consume(ctx, []string{mediaTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			err := handleRawMedia(ctx, key, value, mediaJobs, log)
			if err == nil {
				atomic.AddUint64(&s.metrics.PickedMedia, 1)
			}
			return err
		})
	}()

	go func() {
		defer wgConsumers.Done()
		s.deps.InsightsConsumer.Consume(ctx, []string{insightsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			err := handleRawInsights(ctx, key, value, insightJobs, log)
			if err == nil {
				atomic.AddUint64(&s.metrics.PickedInsights, 1)
			}
			return err
		})
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	wgConsumers.Wait()

	close(mediaJobs)
	close(insightJobs)

	wgParsers.Wait()

	close(s.batchCollectors.posts)
	close(s.batchCollectors.insights)

	batchWg.Wait()

	return nil
}

// GetMetrics returns current service metrics
func (s *AnalyticsSinkService) GetMetrics() AnalyticsSinkMetrics {
	return AnalyticsSinkMetrics{
		PickedMedia:      atomic.LoadUint64(&s.metrics.PickedMedia),
		PickedInsights:   atomic.LoadUint64(&s.metrics.PickedInsights),
		ParsedPosts:      atomic.LoadUint64(&s.metrics.ParsedPosts),
		ParsedInsights:   atomic.LoadUint64(&s.metrics.ParsedInsights),
		InsertedPosts:    atomic.LoadUint64(&s.metrics.InsertedPosts),
		InsertedInsights: atomic.LoadUint64(&s.metrics.InsertedInsights),
	}
}

// GetBatchCollectors returns the batch collectors
func (s *AnalyticsSinkService) GetBatchCollectors() *BatchCollectors {
	return s.batchCollectors
}
