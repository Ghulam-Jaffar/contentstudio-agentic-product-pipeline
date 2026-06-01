package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	maxWorkers             = 5                // Maximum concurrent message processors
	maxBatchSize           = 10000            // Maximum batch size for bulk inserts
	batchTimeout           = 10 * time.Second // Timeout for batching (matching Facebook for faster data visibility)
	messageChanSize        = 50000            // Increased channel buffer size to handle high throughput
	batchProcessorsPerType = 3                // Number of parallel batch processors per data type
	consumerGroup          = "instagram-clickhouse-sink-group"

	defaultIdleTimeoutMinutes = 60
	idleCheckInterval         = 30 * time.Second
)

// Batch collectors for different data types
type BatchCollectors struct {
	posts    chan *kafkamodels.ParsedInstagramPost
	insights chan *kafkamodels.ParsedInstagramInsight
}

func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logger.New("error").Fatal().Err(err).Msg("Failed to load config")
	}
	telemetry.ConfigureSentry(cfg)

	// Initialize logger
	log := logger.New(cfg.LogLevel)

	idleTimeoutMinutes := cfg.SinkIdleTimeoutMinutes
	if idleTimeoutMinutes <= 0 {
		idleTimeoutMinutes = defaultIdleTimeoutMinutes
	}
	idleTimeout := time.Duration(idleTimeoutMinutes) * time.Minute

	log.Info().Dur("idle_timeout", idleTimeout).Msg("Starting Instagram ClickHouse Sink Service")

	// Initialize ClickHouse sink
	clickhouseSink := conversions.NewClickHouseSink(&log.Logger, cfg)

	// Check ClickHouse health
	if err := clickhouseSink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed - service will continue for development")
	}

	// Topics to consume from
	topics := []string{
		"parsed-instagram-posts",
		"parsed-instagram-insights",
	}

	// Initialize Kafka consumer
	consumer, err := kafka.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	// Context for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize batch collectors
	batchCollectors := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, maxBatchSize*5),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, maxBatchSize*5),
	}

	// Atomic counters for metrics
	var pickedPosts, pickedInsights uint64
	var insertedPosts, insertedInsights uint64

	// Track last message time for idle timeout detection
	var lastMessageTime int64 = time.Now().UnixNano()

	// Start batch processors for each data type
	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, batchCollectors, clickhouseSink, log, &batchWg, batchProcessorsPerType, &insertedPosts, &insertedInsights)

	// Handle shutdown gracefully
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	var consumerWg sync.WaitGroup
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		if err := consumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			if strings.HasSuffix(topic, "parsed-instagram-posts") {
				atomic.AddUint64(&pickedPosts, 1)
				return handleParsedPost(ctx, key, value, batchCollectors, log)
			} else if strings.HasSuffix(topic, "parsed-instagram-insights") {
				atomic.AddUint64(&pickedInsights, 1)
				return handleParsedInsight(ctx, key, value, batchCollectors, log)
			}
			log.Warn().Str("topic", topic).Msg("Unknown topic, skipping message")
			return nil
		}); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer").Msg("Consumer error")
			cancel()
		}
	}()

	// Periodic metrics logging
	stopMetrics := make(chan struct{})
	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				log.Info().
					Uint64("picked_posts", atomic.LoadUint64(&pickedPosts)).
					Uint64("picked_insights", atomic.LoadUint64(&pickedInsights)).
					Uint64("inserted_posts", atomic.LoadUint64(&insertedPosts)).
					Uint64("inserted_insights", atomic.LoadUint64(&insertedInsights)).
					Int("posts_queue", len(batchCollectors.posts)).
					Int("insights_queue", len(batchCollectors.insights)).
					Msg("Metrics snapshot")
			case <-stopMetrics:
				return
			}
		}
	}()

	log.Info().
		Int("workers", maxWorkers).
		Strs("topics", topics).
		Str("consumer_group", consumerGroup).
		Int("batch_size", maxBatchSize).
		Dur("batch_timeout", batchTimeout).
		Msg("Instagram ClickHouse Sink Service started successfully")

	<-sigChan
	log.Info().Msg("Received shutdown signal, stopping Instagram ClickHouse Sink Service...")

	// Cancel context to stop consumer
	cancel()

	// Wait for consumer to finish before closing batch channels
	consumerWg.Wait()

	// Close batch channels so processors drain all remaining items and exit
	close(batchCollectors.posts)
	close(batchCollectors.insights)

	// Wait for batch processors to finish
	batchWg.Wait()

	// Stop metrics
	close(stopMetrics)

	log.Info().
		Uint64("total_picked_posts", atomic.LoadUint64(&pickedPosts)).
		Uint64("total_picked_insights", atomic.LoadUint64(&pickedInsights)).
		Uint64("total_inserted_posts", atomic.LoadUint64(&insertedPosts)).
		Uint64("total_inserted_insights", atomic.LoadUint64(&insertedInsights)).
		Msg("Instagram ClickHouse Sink Service stopped")
}

func handleParsedPost(ctx context.Context, key, value []byte, batchCollectors *BatchCollectors, log *logger.Logger) error {
	var parsedPost kafkamodels.ParsedInstagramPost
	if err := json.Unmarshal(value, &parsedPost); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleParsedPost").Str("stage", "unmarshal_post").Str("key", string(key)).Msg("Failed to unmarshal parsed Instagram post")
		return err
	}

	select {
	case batchCollectors.posts <- &parsedPost:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func handleParsedInsight(ctx context.Context, key, value []byte, batchCollectors *BatchCollectors, log *logger.Logger) error {
	var parsedInsight kafkamodels.ParsedInstagramInsight
	if err := json.Unmarshal(value, &parsedInsight); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleParsedInsight").Str("stage", "unmarshal_insight").Str("key", string(key)).Msg("Failed to unmarshal parsed Instagram insight")
		return err
	}

	select {
	case batchCollectors.insights <- &parsedInsight:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func startBatchProcessors(ctx context.Context, batchCollectors *BatchCollectors, sink *conversions.ClickHouseSink, log *logger.Logger, wg *sync.WaitGroup, batchProcessorsPerType int, insertedPosts, insertedInsights *uint64) {
	for i := 0; i < batchProcessorsPerType; i++ {
		wg.Add(1)
		go func(processorID int) {
			defer wg.Done()
			processPostsBatch(ctx, processorID, batchCollectors.posts, sink, log, insertedPosts)
		}(i)
	}

	for i := 0; i < batchProcessorsPerType; i++ {
		wg.Add(1)
		go func(processorID int) {
			defer wg.Done()
			processInsightsBatch(ctx, processorID, batchCollectors.insights, sink, log, insertedInsights)
		}(i)
	}

	log.Info().
		Int("max_batch_size", maxBatchSize).
		Dur("batch_timeout", batchTimeout).
		Int("processors_per_type", batchProcessorsPerType).
		Msg("Started all batch processors")
}

func processPostsBatch(ctx context.Context, processorID int, postsChan <-chan *kafkamodels.ParsedInstagramPost, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	var batch []*kafkamodels.ParsedInstagramPost
	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()

	processorLog := log.With().Int("processor_id", processorID).Str("type", "posts").Logger()
	processorLog.Info().Msg("Batch processor started")

	var processedBatches, totalInserted uint64

	defer func() {
		processorLog.Info().
			Uint64("batches_processed", processedBatches).
			Uint64("total_inserted", totalInserted).
			Msg("Batch processor stopped")
	}()

	for {
		select {
		case post, ok := <-postsChan:
			if !ok {
				if len(batch) > 0 {
					count := processPosts(context.Background(), batch, sink, &processorLog)
					atomic.AddUint64(insertedCounter, uint64(count))
					totalInserted += uint64(count)
				}
				return
			}
			batch = append(batch, post)
			if len(batch) >= maxBatchSize {
				count := processPosts(context.Background(), batch, sink, &processorLog)
				atomic.AddUint64(insertedCounter, uint64(count))
				totalInserted += uint64(count)
				processedBatches++
				batch = nil
				ticker.Reset(batchTimeout)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				count := processPosts(context.Background(), batch, sink, &processorLog)
				atomic.AddUint64(insertedCounter, uint64(count))
				totalInserted += uint64(count)
				processedBatches++
				batch = nil
			}
		}
	}
}

func processInsightsBatch(ctx context.Context, processorID int, insightsChan <-chan *kafkamodels.ParsedInstagramInsight, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	var batch []*kafkamodels.ParsedInstagramInsight
	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()

	processorLog := log.With().Int("processor_id", processorID).Str("type", "insights").Logger()
	processorLog.Info().Msg("Batch processor started")

	var processedBatches, totalInserted uint64

	defer func() {
		processorLog.Info().
			Uint64("batches_processed", processedBatches).
			Uint64("total_inserted", totalInserted).
			Msg("Batch processor stopped")
	}()

	for {
		select {
		case insight, ok := <-insightsChan:
			if !ok {
				if len(batch) > 0 {
					count := processInsights(context.Background(), batch, sink, &processorLog)
					atomic.AddUint64(insertedCounter, uint64(count))
					totalInserted += uint64(count)
				}
				return
			}
			batch = append(batch, insight)
			if len(batch) >= maxBatchSize {
				count := processInsights(context.Background(), batch, sink, &processorLog)
				atomic.AddUint64(insertedCounter, uint64(count))
				totalInserted += uint64(count)
				processedBatches++
				batch = nil
				ticker.Reset(batchTimeout)
			}

		case <-ticker.C:
			if len(batch) > 0 {
				count := processInsights(context.Background(), batch, sink, &processorLog)
				atomic.AddUint64(insertedCounter, uint64(count))
				totalInserted += uint64(count)
				processedBatches++
				batch = nil
			}
		}
	}
}

func processPosts(ctx context.Context, batch []*kafkamodels.ParsedInstagramPost, sink *conversions.ClickHouseSink, log *zerolog.Logger) int {
	startTime := time.Now()
	batchSize := len(batch)

	convertedPosts := make([]*clickhousemodels.InstagramPost, 0, batchSize)
	for _, p := range batch {
		convertedPosts = append(convertedPosts, sink.ConvertInstagramPost(p))
	}

	if err := sink.BulkInsertInstagramPosts(ctx, convertedPosts); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processPosts").Str("stage", "bulk_insert_posts").Int("batch_size", batchSize).Dur("duration", time.Since(startTime)).Msg("Failed to bulk insert Instagram posts")
		return 0
	}

	log.Info().Int("batch_size", batchSize).Dur("duration", time.Since(startTime)).Msg("Inserted Instagram posts batch")
	return batchSize
}

func processInsights(ctx context.Context, batch []*kafkamodels.ParsedInstagramInsight, sink *conversions.ClickHouseSink, log *zerolog.Logger) int {
	startTime := time.Now()
	batchSize := len(batch)

	convertedInsights := make([]*clickhousemodels.InstagramInsight, 0, batchSize)
	for _, i := range batch {
		convertedInsights = append(convertedInsights, sink.ConvertInstagramInsight(i))
	}

	if err := sink.BulkInsertInstagramInsights(ctx, convertedInsights); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processInsights").Str("stage", "bulk_insert_insights").Int("batch_size", batchSize).Dur("duration", time.Since(startTime)).Msg("Failed to bulk insert Instagram insights")
		return 0
	}

	log.Info().Int("batch_size", batchSize).Dur("duration", time.Since(startTime)).Msg("Inserted Instagram insights batch")
	return batchSize
}
