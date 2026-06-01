package main

import (
	"context"
	"sync"
	"sync/atomic"

	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ParserServiceConfig holds service configuration
type ParserServiceConfig struct {
	PostsParserWorkers         int
	MediaInsightsParserWorkers int
	PostsPublisherWorkers      int
	MediaInsightsPublisherWorkers int
	ParseChanSize              int
	PublishChanSize            int
}

// DefaultParserServiceConfig returns default configuration
func DefaultParserServiceConfig() ParserServiceConfig {
	return ParserServiceConfig{
		PostsParserWorkers:         postsParserWorkers,
		MediaInsightsParserWorkers: mediaInsightsParserWorkers,
		PostsPublisherWorkers:      postsPublisherWorkers,
		MediaInsightsPublisherWorkers: mediaInsightsPublisherWorkers,
		ParseChanSize:              parseChanSize,
		PublishChanSize:            publishChanSize,
	}
}

// ParserServiceDependencies holds external dependencies
type ParserServiceDependencies struct {
	Producer       kafka2.Producer
	PostsConsumer  kafka2.Consumer
	MIConsumer     kafka2.Consumer // Media + Insights consumer
	Log            *logger.Logger
}

// ParserServiceMetrics holds service metrics
type ParserServiceMetrics struct {
	PickedPosts       uint64
	PickedVideos      uint64
	PickedInsights    uint64
	PublishedPosts    uint64
	PublishedAssets   uint64
	PublishedVideoIns uint64
	PublishedReelsIns uint64
	PublishedPageIns  uint64
}

// ParserService represents the posts parser service
type ParserService struct {
	config  ParserServiceConfig
	deps    ParserServiceDependencies
	metrics ParserServiceMetrics
}

// NewParserService creates a new parser service
func NewParserService(cfg ParserServiceConfig, deps ParserServiceDependencies) *ParserService {
	return &ParserService{
		config: cfg,
		deps:   deps,
	}
}

// Run starts the parser service
func (s *ParserService) Run(ctx context.Context) error {
	log := s.deps.Log

	// Channels
	postsParseJobs := make(chan ParseJob, s.config.ParseChanSize)
	miParseJobs := make(chan ParseJob, s.config.ParseChanSize)
	postsPublishJobs := make(chan PublishJob, s.config.PublishChanSize)
	miPublishJobs := make(chan PublishJob, s.config.PublishChanSize)

	// Start parser pools
	var wgParsers sync.WaitGroup
	for i := 0; i < s.config.PostsParserWorkers; i++ {
		wgParsers.Add(1)
		go postsParser(ctx, &wgParsers, i, postsParseJobs, postsPublishJobs, log)
	}
	for i := 0; i < s.config.MediaInsightsParserWorkers; i++ {
		wgParsers.Add(1)
		go mediaInsightsParser(ctx, &wgParsers, i, miParseJobs, miPublishJobs, log)
	}

	// Start publisher pools
	var wgPublishers sync.WaitGroup
	for i := 0; i < s.config.PostsPublisherWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "posts", postsPublishJobs, s.deps.Producer,
			&s.metrics.PublishedPosts, &s.metrics.PublishedAssets,
			&s.metrics.PublishedVideoIns, &s.metrics.PublishedReelsIns,
			&s.metrics.PublishedPageIns, log)
	}
	for i := 0; i < s.config.MediaInsightsPublisherWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "media-insights", miPublishJobs, s.deps.Producer,
			&s.metrics.PublishedPosts, &s.metrics.PublishedAssets,
			&s.metrics.PublishedVideoIns, &s.metrics.PublishedReelsIns,
			&s.metrics.PublishedPageIns, log)
	}

	// Start consumers
	var wgConsumers sync.WaitGroup
	wgConsumers.Add(2)

	// Posts consumer
	go func() {
		defer wgConsumers.Done()
		s.deps.PostsConsumer.Consume(ctx, []string{rawPostsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.AddUint64(&s.metrics.PickedPosts, 1)
			return handleRawPost(ctx, key, value, postsParseJobs, log)
		})
	}()

	// Media + Insights consumer
	go func() {
		defer wgConsumers.Done()
		s.deps.MIConsumer.Consume(ctx, []string{rawVideosTopic, rawInsightsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			switch topic {
			case rawVideosTopic:
				atomic.AddUint64(&s.metrics.PickedVideos, 1)
				return handleRawVideo(ctx, key, value, miParseJobs, log)
			case rawInsightsTopic:
				atomic.AddUint64(&s.metrics.PickedInsights, 1)
				return handleRawInsights(ctx, key, value, miParseJobs, log)
			}
			return nil
		})
	}()

	// Wait for context cancellation
	<-ctx.Done()

	// Graceful shutdown
	wgConsumers.Wait()

	close(postsParseJobs)
	close(miParseJobs)

	wgParsers.Wait()

	close(postsPublishJobs)
	close(miPublishJobs)

	wgPublishers.Wait()

	return nil
}

// GetMetrics returns current service metrics
func (s *ParserService) GetMetrics() ParserServiceMetrics {
	return ParserServiceMetrics{
		PickedPosts:       atomic.LoadUint64(&s.metrics.PickedPosts),
		PickedVideos:      atomic.LoadUint64(&s.metrics.PickedVideos),
		PickedInsights:    atomic.LoadUint64(&s.metrics.PickedInsights),
		PublishedPosts:    atomic.LoadUint64(&s.metrics.PublishedPosts),
		PublishedAssets:   atomic.LoadUint64(&s.metrics.PublishedAssets),
		PublishedVideoIns: atomic.LoadUint64(&s.metrics.PublishedVideoIns),
		PublishedReelsIns: atomic.LoadUint64(&s.metrics.PublishedReelsIns),
		PublishedPageIns:  atomic.LoadUint64(&s.metrics.PublishedPageIns),
	}
}
