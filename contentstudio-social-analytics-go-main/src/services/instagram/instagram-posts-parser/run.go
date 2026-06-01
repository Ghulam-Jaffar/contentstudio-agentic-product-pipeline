package main

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// ParserConfig holds service configuration
type ParserConfig struct {
	MediaParserWorkers     int
	InsightsParserWorkers  int
	MediaPublishWorkers    int
	InsightsPublishWorkers int
	MessageChanSize        int
}

// DefaultParserConfig returns default configuration
func DefaultParserConfig() ParserConfig {
	return ParserConfig{
		MediaParserWorkers:     mediaParserWorkers,
		InsightsParserWorkers:  insightsParserWorkers,
		MediaPublishWorkers:    mediaPublishWorkers,
		InsightsPublishWorkers: insightsPublishWorkers,
		MessageChanSize:        messageChanSize,
	}
}

// ParserDependencies holds external dependencies
type ParserDependencies struct {
	MediaConsumer    kafka.Consumer
	InsightsConsumer kafka.Consumer
	Producer         kafka.Producer
	Log              *logger.Logger
}

// ParserMetrics holds service metrics
type ParserMetrics struct {
	PickedMedia     uint64
	PickedInsights  uint64
	PublishedMedia  uint64
	PublishedInsights uint64
}

// ParserService represents the posts parser service
type ParserService struct {
	config  ParserConfig
	deps    ParserDependencies
	metrics ParserMetrics
}

// NewParserService creates a new parser service
func NewParserService(cfg ParserConfig, deps ParserDependencies) *ParserService {
	return &ParserService{
		config: cfg,
		deps:   deps,
	}
}

// Run starts the parser service
func (s *ParserService) Run(ctx context.Context) error {
	log := s.deps.Log

	mediaJobs := make(chan ParseJob, s.config.MessageChanSize)
	insightJobs := make(chan ParseJob, s.config.MessageChanSize)
	mediaPubJobs := make(chan PublishJob, s.config.MessageChanSize)
	insightPubJobs := make(chan PublishJob, s.config.MessageChanSize)

	// Start parser pools
	var wgParsers sync.WaitGroup
	for i := 0; i < s.config.MediaParserWorkers; i++ {
		wgParsers.Add(1)
		go mediaParser(ctx, &wgParsers, i, mediaJobs, mediaPubJobs, &s.metrics.PublishedMedia, log)
	}
	for i := 0; i < s.config.InsightsParserWorkers; i++ {
		wgParsers.Add(1)
		go insightsParser(ctx, &wgParsers, i, insightJobs, insightPubJobs, &s.metrics.PublishedInsights, log)
	}

	// Start publisher pools
	var wgPublishers sync.WaitGroup
	for i := 0; i < s.config.MediaPublishWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "media", mediaPubJobs, s.deps.Producer, &s.metrics.PublishedMedia, log)
	}
	for i := 0; i < s.config.InsightsPublishWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "insights", insightPubJobs, s.deps.Producer, &s.metrics.PublishedInsights, log)
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

	close(mediaPubJobs)
	close(insightPubJobs)

	wgPublishers.Wait()

	return nil
}

// GetMetrics returns current service metrics
func (s *ParserService) GetMetrics() ParserMetrics {
	return ParserMetrics{
		PickedMedia:      atomic.LoadUint64(&s.metrics.PickedMedia),
		PickedInsights:   atomic.LoadUint64(&s.metrics.PickedInsights),
		PublishedMedia:   atomic.LoadUint64(&s.metrics.PublishedMedia),
		PublishedInsights: atomic.LoadUint64(&s.metrics.PublishedInsights),
	}
}
