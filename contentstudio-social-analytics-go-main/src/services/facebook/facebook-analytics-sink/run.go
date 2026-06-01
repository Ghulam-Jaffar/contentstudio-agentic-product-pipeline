package main

import (
	"context"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ServiceConfig holds configuration for the analytics sink service
type ServiceConfig struct {
	PostsParserWorkers     int
	InsightsParserWorkers  int
	BatchProcessorsPerType int
	MaxBatchSize           int
	BatchTimeout           time.Duration
	IdleTimeout            time.Duration
	ParseChanSize          int
	MessageChanSize        int
}

// DefaultServiceConfig returns the default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		PostsParserWorkers:     postsParserWorkers,
		InsightsParserWorkers:  insightsParserWorkers,
		BatchProcessorsPerType: batchProcessorsPerType,
		MaxBatchSize:           maxBatchSize,
		BatchTimeout:           batchTimeout,
		IdleTimeout:            idleTimeout,
		ParseChanSize:          parseChanSize,
		MessageChanSize:        messageChanSize,
	}
}

// ServiceDependencies holds all external dependencies
type ServiceDependencies struct {
	Sink          ClickHouseSinkInterface
	PostsConsumer KafkaConsumerInterface
	MIConsumer    KafkaConsumerInterface
	Logger        *logger.Logger
}

// ServiceMetrics holds runtime metrics
type ServiceMetrics struct {
	PickedPosts     uint64
	PickedVideos    uint64
	PickedInsights  uint64
	ParsedPosts     uint64
	ParsedAssets    uint64
	ParsedVideoIns  uint64
	ParsedReelsIns  uint64
	ParsedPageIns   uint64
	LastMessageTime int64
}

// RunService runs the analytics sink service with the given dependencies
func RunService(ctx context.Context, deps ServiceDependencies, cfg ServiceConfig, shutdownSignal <-chan struct{}) (*ServiceMetrics, error) {
	log := deps.Logger
	sink := deps.Sink

	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed - continuing anyway")
	}

	postsParseJobs := make(chan ParseJob, cfg.ParseChanSize)
	miParseJobs := make(chan ParseJob, cfg.ParseChanSize)

	batches := &BatchCollectors{
		posts:         make(chan *kafkamodels.ParsedFacebookPost, cfg.MaxBatchSize*5),
		mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, cfg.MaxBatchSize*5),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, cfg.MaxBatchSize*5),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, cfg.MaxBatchSize*5),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, cfg.MaxBatchSize*5),
	}

	metrics := &ServiceMetrics{
		LastMessageTime: time.Now().UnixNano(),
	}

	var batchWg sync.WaitGroup
	startBatchProcessorsWithInterface(ctx, batches, sink, log, &batchWg, cfg.BatchProcessorsPerType, cfg.MaxBatchSize, cfg.BatchTimeout)

	var wgParsers sync.WaitGroup
	for i := 0; i < cfg.PostsParserWorkers; i++ {
		wgParsers.Add(1)
		go postsParser(ctx, &wgParsers, i, postsParseJobs, batches, &metrics.ParsedPosts, &metrics.ParsedAssets, log)
	}
	for i := 0; i < cfg.InsightsParserWorkers; i++ {
		wgParsers.Add(1)
		go mediaInsightsParser(ctx, &wgParsers, i, miParseJobs, batches, &metrics.ParsedVideoIns, &metrics.ParsedReelsIns, &metrics.ParsedPageIns, log)
	}

	var wgConsumers sync.WaitGroup
	wgConsumers.Add(2)

	consumerCtx, cancelConsumers := context.WithCancel(ctx)

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", rawPostsTopic).Msg("Consuming posts...")
		err := deps.PostsConsumer.Consume(consumerCtx, []string{rawPostsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&metrics.LastMessageTime, time.Now().UnixNano())
			if err := handleRawPost(ctx, key, value, postsParseJobs, log); err == nil {
				atomic.AddUint64(&metrics.PickedPosts, 1)
			}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Posts consumer error")
		}
	}()

	go func() {
		defer wgConsumers.Done()
		topics := []string{rawVideosTopic, rawInsightsTopic}
		log.Info().Strs("topics", topics).Msg("Consuming videos+insights...")
		err := deps.MIConsumer.Consume(consumerCtx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&metrics.LastMessageTime, time.Now().UnixNano())
			switch {
			case strings.HasSuffix(topic, rawVideosTopic):
				if err := handleRawVideo(ctx, key, value, miParseJobs, log); err == nil {
					atomic.AddUint64(&metrics.PickedVideos, 1)
				}
			case strings.HasSuffix(topic, rawInsightsTopic):
				if err := handleRawInsights(ctx, key, value, miParseJobs, log); err == nil {
					atomic.AddUint64(&metrics.PickedInsights, 1)
				}
			}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Media+Insights consumer error")
		}
	}()

	// Wait for shutdown signal or context cancellation
	select {
	case <-ctx.Done():
	case <-shutdownSignal:
	}

	cancelConsumers()
	wgConsumers.Wait()

	close(postsParseJobs)
	close(miParseJobs)
	wgParsers.Wait()

	close(batches.posts)
	close(batches.mediaAssets)
	close(batches.pageInsights)
	close(batches.videoInsights)
	close(batches.reelsInsights)
	batchWg.Wait()

	return metrics, nil
}

// startBatchProcessorsWithInterface starts batch processors using the interface
func startBatchProcessorsWithInterface(ctx context.Context, b *BatchCollectors, sink ClickHouseSinkInterface, log *logger.Logger, wg *sync.WaitGroup, n, maxBatch int, timeout time.Duration) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processPostsBatchWithInterface(ctx, b.posts, sink, log, maxBatch, timeout)
		}()
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processMediaAssetsBatchWithInterface(ctx, b.mediaAssets, sink, log, maxBatch, timeout)
		}()
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processPageInsightsBatchWithInterface(ctx, b.pageInsights, sink, log, maxBatch, timeout)
		}()
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processVideoInsightsBatchWithInterface(ctx, b.videoInsights, sink, log, maxBatch, timeout)
		}()
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			processReelsInsightsBatchWithInterface(ctx, b.reelsInsights, sink, log, maxBatch, timeout)
		}()
	}

	log.Info().
		Int("max_batch_size", maxBatch).
		Dur("batch_timeout", timeout).
		Int("processors_per_type", n).
		Msg("Started all batch processors")
}

func processPostsBatchWithInterface(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookPost, sink ClickHouseSinkInterface, log *logger.Logger, maxBatch int, timeout time.Duration) {
	var batch []*kafkamodels.ParsedFacebookPost
	t := time.NewTicker(timeout)
	defer t.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		rows := make([]*clickhousemodels.FacebookPosts, len(batch))
		for i, x := range batch {
			rows[i] = sink.ConvertFacebookPost(x)
		}
		if err := sink.BulkInsertPosts(context.Background(), rows); err != nil {
			log.Error().Err(err).Int("count", len(batch)).Msg("bulk insert posts failed")
		} else {
			log.Info().Int("batch_size", len(batch)).Msg("Inserted posts batch")
		}
		batch = nil
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case x, ok := <-in:
			if !ok {
				flush()
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatch {
				flush()
				t.Reset(timeout)
			}
		case <-t.C:
			flush()
		}
	}
}

func processMediaAssetsBatchWithInterface(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookMediaAsset, sink ClickHouseSinkInterface, log *logger.Logger, maxBatch int, timeout time.Duration) {
	var batch []*kafkamodels.ParsedFacebookMediaAsset
	t := time.NewTicker(timeout)
	defer t.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		rows := make([]*clickhousemodels.FacebookMediaAssets, len(batch))
		for i, x := range batch {
			rows[i] = sink.ConvertFacebookMediaAssets(x)
		}
		if err := sink.BulkInsertMediaAssets(context.Background(), rows); err != nil {
			log.Error().Err(err).Int("count", len(batch)).Msg("bulk insert media assets failed")
		} else {
			log.Info().Int("batch_size", len(batch)).Msg("Inserted media assets batch")
		}
		batch = nil
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case x, ok := <-in:
			if !ok {
				flush()
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatch {
				flush()
				t.Reset(timeout)
			}
		case <-t.C:
			flush()
		}
	}
}

func processPageInsightsBatchWithInterface(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookInsights, sink ClickHouseSinkInterface, log *logger.Logger, maxBatch int, timeout time.Duration) {
	var batch []*kafkamodels.ParsedFacebookInsights
	t := time.NewTicker(timeout)
	defer t.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		rows := make([]*clickhousemodels.FacebookInsights, len(batch))
		for i, x := range batch {
			rows[i] = sink.ConvertFacebookInsights(x)
		}
		if err := sink.BulkInsertInsights(context.Background(), rows); err != nil {
			log.Error().Err(err).Int("count", len(batch)).Msg("bulk insert insights failed")
		} else {
			log.Info().Int("batch_size", len(batch)).Msg("Inserted insights batch")
		}
		batch = nil
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case x, ok := <-in:
			if !ok {
				flush()
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatch {
				flush()
				t.Reset(timeout)
			}
		case <-t.C:
			flush()
		}
	}
}

func processVideoInsightsBatchWithInterface(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookVideoInsights, sink ClickHouseSinkInterface, log *logger.Logger, maxBatch int, timeout time.Duration) {
	var batch []*kafkamodels.ParsedFacebookVideoInsights
	t := time.NewTicker(timeout)
	defer t.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		rows := make([]*clickhousemodels.FacebookVideoInsights, len(batch))
		for i, x := range batch {
			rows[i] = sink.ConvertFacebookVideoInsights(x)
		}
		if err := sink.BulkInsertVideoInsights(context.Background(), rows); err != nil {
			log.Error().Err(err).Int("count", len(batch)).Msg("bulk insert video insights failed")
		} else {
			log.Info().Int("batch_size", len(batch)).Msg("Inserted video insights batch")
		}
		batch = nil
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case x, ok := <-in:
			if !ok {
				flush()
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatch {
				flush()
				t.Reset(timeout)
			}
		case <-t.C:
			flush()
		}
	}
}

func processReelsInsightsBatchWithInterface(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookReelsInsights, sink ClickHouseSinkInterface, log *logger.Logger, maxBatch int, timeout time.Duration) {
	var batch []*kafkamodels.ParsedFacebookReelsInsights
	t := time.NewTicker(timeout)
	defer t.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		rows := make([]*clickhousemodels.FacebookReelsInsights, len(batch))
		for i, x := range batch {
			rows[i] = sink.ConvertFacebookReelsInsights(x)
		}
		if err := sink.BulkInsertReelsInsights(context.Background(), rows); err != nil {
			log.Error().Err(err).Int("count", len(batch)).Msg("bulk insert reels insights failed")
		} else {
			log.Info().Int("batch_size", len(batch)).Msg("Inserted reels insights batch")
		}
		batch = nil
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return
		case x, ok := <-in:
			if !ok {
				flush()
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatch {
				flush()
				t.Reset(timeout)
			}
		case <-t.C:
			flush()
		}
	}
}
