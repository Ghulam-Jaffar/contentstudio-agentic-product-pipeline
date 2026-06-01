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
	BatchProcessorsPerType int
	MaxBatchSize           int
	BatchTimeout           time.Duration
}

// DefaultServiceConfig returns default configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		BatchProcessorsPerType: batchProcessorsPerType,
		MaxBatchSize:           maxBatchSize,
		BatchTimeout:           batchTimeout,
	}
}

// ServiceDependencies holds external dependencies
type ServiceDependencies struct {
	Consumer kafka.Consumer
	Sink     *conversions.ClickHouseSink
	Log      *logger.Logger
}

// ServiceMetrics holds service metrics
type ServiceMetrics struct {
	PickedPosts      uint64
	PickedInsights   uint64
	InsertedPosts    uint64
	InsertedInsights uint64
}

// Service represents the clickhouse sink service
type Service struct {
	config          ServiceConfig
	deps            ServiceDependencies
	metrics         ServiceMetrics
	batchCollectors *BatchCollectors
}

// NewService creates a new service
func NewService(cfg ServiceConfig, deps ServiceDependencies) *Service {
	return &Service{
		config: cfg,
		deps:   deps,
		batchCollectors: &BatchCollectors{
			posts:    make(chan *kafkamodels.ParsedInstagramPost, cfg.MaxBatchSize*5),
			insights: make(chan *kafkamodels.ParsedInstagramInsight, cfg.MaxBatchSize*5),
		},
	}
}

// Run starts the service
func (s *Service) Run(ctx context.Context) error {
	log := s.deps.Log

	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, s.batchCollectors, s.deps.Sink, log, &batchWg, s.config.BatchProcessorsPerType, &s.metrics.InsertedPosts, &s.metrics.InsertedInsights)

	topics := []string{
		"parsed-instagram-posts",
		"parsed-instagram-insights",
	}

	var consumerWg sync.WaitGroup
	consumerWg.Add(1)

	go func() {
		defer consumerWg.Done()
		s.deps.Consumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			if topic == "parsed-instagram-posts" {
				atomic.AddUint64(&s.metrics.PickedPosts, 1)
				return handleParsedPost(ctx, key, value, s.batchCollectors, log)
			} else if topic == "parsed-instagram-insights" {
				atomic.AddUint64(&s.metrics.PickedInsights, 1)
				return handleParsedInsight(ctx, key, value, s.batchCollectors, log)
			}
			return nil
		})
	}()

	<-ctx.Done()

	consumerWg.Wait()

	close(s.batchCollectors.posts)
	close(s.batchCollectors.insights)

	batchWg.Wait()

	return nil
}

// GetMetrics returns current service metrics
func (s *Service) GetMetrics() ServiceMetrics {
	return ServiceMetrics{
		PickedPosts:      atomic.LoadUint64(&s.metrics.PickedPosts),
		PickedInsights:   atomic.LoadUint64(&s.metrics.PickedInsights),
		InsertedPosts:    atomic.LoadUint64(&s.metrics.InsertedPosts),
		InsertedInsights: atomic.LoadUint64(&s.metrics.InsertedInsights),
	}
}

// GetBatchCollectors returns the batch collectors
func (s *Service) GetBatchCollectors() *BatchCollectors {
	return s.batchCollectors
}
