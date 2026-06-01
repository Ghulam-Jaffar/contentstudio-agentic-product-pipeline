package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
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
	RawPostsTopic          string
	RawInsightsTopic       string
}

// DefaultServiceConfig returns the default service configuration
func DefaultServiceConfig() ServiceConfig {
	return ServiceConfig{
		PostsParserWorkers:     5,
		InsightsParserWorkers:  5,
		BatchProcessorsPerType: 3,
		MaxBatchSize:           10000,
		BatchTimeout:           10 * time.Second,
		IdleTimeout:            15 * time.Minute,
		ParseChanSize:          1000,
		MessageChanSize:        50000,
		RawPostsTopic:          "raw-twitter-posts",
		RawInsightsTopic:       "raw-twitter-insights",
	}
}

// ServiceDependencies holds all external dependencies
type ServiceDependencies struct {
	Sink             ClickHouseSinkInterface
	PostsConsumer    KafkaConsumerInterface
	InsightsConsumer KafkaConsumerInterface
	Logger           *logger.Logger
}

// ServiceMetrics holds runtime metrics
type ServiceMetrics struct {
	PostsMessagesReceived    uint64
	InsightsMessagesReceived uint64
	PostsMessagesParsed      uint64
	InsightsMessagesParsed   uint64
	MessagesBatched          uint64
	MessagesFailed           uint64
	InsertErrors             uint64
}

// RawMessage represents a raw Kafka message
type RawMessage struct {
	Topic string
	Key   []byte
	Value []byte
}

// BatchCollectors holds channels for batching
type BatchCollectors struct {
	posts    chan *clickhousemodels.TwitterPosts
	insights chan *clickhousemodels.TwitterInsights
}

// RunService starts the Twitter analytics sink service
func RunService(ctx context.Context, deps *ServiceDependencies, cfg ServiceConfig) error {
	log := deps.Logger

	metrics := &ServiceMetrics{}

	log.Info().
		Int("posts_parser_workers", cfg.PostsParserWorkers).
		Int("insights_parser_workers", cfg.InsightsParserWorkers).
		Int("max_batch_size", cfg.MaxBatchSize).
		Dur("batch_timeout", cfg.BatchTimeout).
		Msg("Starting Twitter Analytics Sink")

	var wg sync.WaitGroup

	// Posts parser and ClickHouse sink
	wg.Add(1)
	go processPosts(ctx, deps, &wg, cfg, metrics, log)

	// Insights parser and ClickHouse sink
	wg.Add(1)
	go processInsights(ctx, deps, &wg, cfg, metrics, log)

	wg.Wait()

	log.Info().
		Uint64("posts_received", atomic.LoadUint64(&metrics.PostsMessagesReceived)).
		Uint64("insights_received", atomic.LoadUint64(&metrics.InsightsMessagesReceived)).
		Uint64("parsed_total", atomic.LoadUint64(&metrics.PostsMessagesParsed)+atomic.LoadUint64(&metrics.InsightsMessagesParsed)).
		Uint64("insert_errors", atomic.LoadUint64(&metrics.InsertErrors)).
		Msg("Twitter Analytics Sink stopped")

	return nil
}

// processPosts handles Twitter posts: parse, batch, and insert to ClickHouse
func processPosts(ctx context.Context, deps *ServiceDependencies, wg *sync.WaitGroup, cfg ServiceConfig, metrics *ServiceMetrics, log *logger.Logger) {
	defer wg.Done()

	postsLog := &logger.Logger{Logger: log.Logger.With().Str("type", "posts").Logger()}

	postsJobs := make(chan RawMessage, cfg.MessageChanSize)
	batches := make(chan *clickhousemodels.TwitterPosts, cfg.MaxBatchSize*5)

	var batchWg sync.WaitGroup

	// Start batch processors
	for i := 0; i < cfg.BatchProcessorsPerType; i++ {
		batchWg.Add(1)
		go postsBatchProcessor(ctx, &batchWg, i, batches, deps.Sink, metrics, postsLog)
	}

	// Start parser workers
	var parseWg sync.WaitGroup
	for i := 0; i < cfg.PostsParserWorkers; i++ {
		parseWg.Add(1)
		go postsParser(ctx, &parseWg, i, postsJobs, batches, metrics, postsLog)
	}

	// Consume from Kafka
	err := deps.PostsConsumer.Consume(ctx, []string{cfg.RawPostsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
		received := atomic.AddUint64(&metrics.PostsMessagesReceived, 1)
		postsJobs <- RawMessage{Topic: topic, Key: key, Value: value}
		postsLog.Info().
			Uint64("posts_messages_received_total", received).
			Int("posts_jobs_queue_depth", len(postsJobs)).
			Str("topic", topic).
			Str("key", string(key)).
			Msg("Consumed raw Twitter post message and queued for parsing")
		return nil
	})

	if err != nil && err != context.Canceled {
		log.Error().
			Err(err).
			Str("error_message", err.Error()).
			Str("function", "processPosts").
			Str("stage", "kafka_consume").
			Msg("Posts consumer error")
	}

	close(postsJobs)
	parseWg.Wait()
	close(batches)
	batchWg.Wait()
}

// processInsights handles Twitter insights: parse, batch, and insert to ClickHouse
func processInsights(ctx context.Context, deps *ServiceDependencies, wg *sync.WaitGroup, cfg ServiceConfig, metrics *ServiceMetrics, log *logger.Logger) {
	defer wg.Done()

	insightsLog := &logger.Logger{Logger: log.Logger.With().Str("type", "insights").Logger()}

	insightsJobs := make(chan RawMessage, cfg.MessageChanSize)
	batches := make(chan *clickhousemodels.TwitterInsights, cfg.MaxBatchSize*5)

	var batchWg sync.WaitGroup

	// Start batch processors
	for i := 0; i < cfg.BatchProcessorsPerType; i++ {
		batchWg.Add(1)
		go insightsBatchProcessor(ctx, &batchWg, i, batches, deps.Sink, metrics, insightsLog)
	}

	// Start parser workers
	var parseWg sync.WaitGroup
	for i := 0; i < cfg.InsightsParserWorkers; i++ {
		parseWg.Add(1)
		go insightsParser(ctx, &parseWg, i, insightsJobs, batches, metrics, insightsLog)
	}

	// Consume from Kafka
	err := deps.InsightsConsumer.Consume(ctx, []string{cfg.RawInsightsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
		received := atomic.AddUint64(&metrics.InsightsMessagesReceived, 1)
		insightsJobs <- RawMessage{Topic: topic, Key: key, Value: value}
		insightsLog.Info().
			Uint64("insights_messages_received_total", received).
			Int("insights_jobs_queue_depth", len(insightsJobs)).
			Str("topic", topic).
			Str("key", string(key)).
			Msg("Consumed raw Twitter insight message and queued for parsing")
		return nil
	})

	if err != nil && err != context.Canceled {
		log.Error().
			Err(err).
			Str("error_message", err.Error()).
			Str("function", "processInsights").
			Str("stage", "kafka_consume").
			Msg("Insights consumer error")
	}

	close(insightsJobs)
	parseWg.Wait()
	close(batches)
	batchWg.Wait()
}

// postsParser parses Twitter posts and sends them to the batch channel
func postsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, out chan<- *clickhousemodels.TwitterPosts, metrics *ServiceMetrics, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker_id", id).Str("pool", "posts").Msg("Posts parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			var rawPost kafkamodels.RawTwitterPost
			if err := json.Unmarshal(msg.Value, &rawPost); err != nil {
				log.Error().
					Err(err).
					Str("error_message", err.Error()).
					Str("key", string(msg.Key)).
					Str("function", "postsParser").
					Str("stage", "unmarshal_raw_post").
					Msg("Failed to unmarshal raw post")
				continue
			}

			// Data field contains the ParsedTwitterPost from the fetcher
			var parsedPost kafkamodels.ParsedTwitterPost
			if err := json.Unmarshal(rawPost.Data, &parsedPost); err != nil {
				log.Error().
					Err(err).
					Str("error_message", err.Error()).
					Str("twitter_id", rawPost.TwitterID).
					Str("function", "postsParser").
					Str("stage", "unmarshal_parsed_post").
					Msg("Failed to unmarshal parsed post")
				continue
			}

			if parsedPost.TweetID == "" {
				log.Debug().Str("twitter_id", rawPost.TwitterID).Msg("Parsed post has empty TweetID, skipping")
				continue
			}

			chPost := conversions.ConvertTwitterPost(&parsedPost)
			if chPost != nil {
				select {
				case out <- chPost:
					parsed := atomic.AddUint64(&metrics.PostsMessagesParsed, 1)
					log.Info().
						Int("worker_id", id).
						Str("tweet_id", parsedPost.TweetID).
						Uint64("posts_messages_parsed_total", parsed).
						Int("posts_batch_queue_depth", len(out)).
						Msg("Parsed Twitter post and queued for batch insert")
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

// insightsParser parses Twitter insights and sends them to the batch channel
func insightsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, out chan<- *clickhousemodels.TwitterInsights, metrics *ServiceMetrics, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker_id", id).Str("pool", "insights").Msg("Insights parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			var parsedInsight kafkamodels.ParsedTwitterInsights
			if err := json.Unmarshal(msg.Value, &parsedInsight); err != nil {
				log.Error().
					Err(err).
					Str("error_message", err.Error()).
					Str("key", string(msg.Key)).
					Str("function", "insightsParser").
					Str("stage", "unmarshal_insight").
					Msg("Failed to unmarshal parsed insight")
				continue
			}

			if parsedInsight.RecordID == "" {
				log.Debug().Str("key", string(msg.Key)).Msg("Parsed insight has empty RecordID, skipping")
				continue
			}

			chInsight := conversions.ConvertTwitterInsights(&parsedInsight)
			if chInsight == nil {
				continue
			}

			select {
			case out <- chInsight:
				parsed := atomic.AddUint64(&metrics.InsightsMessagesParsed, 1)
				log.Info().
					Int("worker_id", id).
					Str("record_id", parsedInsight.RecordID).
					Uint64("insights_messages_parsed_total", parsed).
					Int("insights_batch_queue_depth", len(out)).
					Msg("Parsed Twitter insight and queued for batch insert")
			case <-ctx.Done():
				return
			}
		}
	}
}

// postsBatchProcessor batches and inserts Twitter posts into ClickHouse
func postsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *clickhousemodels.TwitterPosts, sink ClickHouseSinkInterface, metrics *ServiceMetrics, log *logger.Logger) {
	defer wg.Done()
	batch := make([]*clickhousemodels.TwitterPosts, 0, 10000)
	flushTimer := time.NewTimer(10 * time.Second)
	defer flushTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				if err := sink.BulkInsertTwitterPosts(ctx, batch); err != nil {
					log.Error().
						Err(err).
						Str("error_message", err.Error()).
						Int("batch_size", len(batch)).
						Str("function", "postsBatchProcessor").
						Str("stage", "flush_final_batch").
						Msg("Failed to flush final posts batch")
					atomic.AddUint64(&metrics.InsertErrors, 1)
				}
			}
			return
		case <-flushTimer.C:
			if len(batch) > 0 {
				log.Info().Int("worker_id", id).Int("batch_size", len(batch)).Msg("Flushing Twitter posts batch on timer")
				if err := sink.BulkInsertTwitterPosts(ctx, batch); err != nil {
					log.Error().
						Err(err).
						Str("error_message", err.Error()).
						Int("batch_size", len(batch)).
						Str("function", "postsBatchProcessor").
						Str("stage", "timer_flush").
						Msg("Failed to insert posts batch")
					atomic.AddUint64(&metrics.InsertErrors, 1)
				}
				batch = make([]*clickhousemodels.TwitterPosts, 0, 10000)
			}
			flushTimer.Reset(10 * time.Second)
		case item, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					if err := sink.BulkInsertTwitterPosts(ctx, batch); err != nil {
						log.Error().
							Err(err).
							Str("error_message", err.Error()).
							Int("batch_size", len(batch)).
							Str("function", "postsBatchProcessor").
							Str("stage", "flush_final_batch").
							Msg("Failed to flush final posts batch")
						atomic.AddUint64(&metrics.InsertErrors, 1)
					}
				}
				return
			}
			batch = append(batch, item)
			if len(batch) >= 10000 {
				log.Info().Int("worker_id", id).Int("batch_size", len(batch)).Msg("Flushing Twitter posts batch on size threshold")
				if err := sink.BulkInsertTwitterPosts(ctx, batch); err != nil {
					log.Error().
						Err(err).
						Str("error_message", err.Error()).
						Int("batch_size", len(batch)).
						Str("function", "postsBatchProcessor").
						Str("stage", "size_threshold_flush").
						Msg("Failed to insert posts batch")
					atomic.AddUint64(&metrics.InsertErrors, 1)
				}
				batch = make([]*clickhousemodels.TwitterPosts, 0, 10000)
				flushTimer.Reset(10 * time.Second)
			}
		}
	}
}

// insightsBatchProcessor batches and inserts Twitter insights into ClickHouse
func insightsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *clickhousemodels.TwitterInsights, sink ClickHouseSinkInterface, metrics *ServiceMetrics, log *logger.Logger) {
	defer wg.Done()
	batch := make([]*clickhousemodels.TwitterInsights, 0, 10000)
	flushTimer := time.NewTimer(10 * time.Second)
	defer flushTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				if err := sink.BulkInsertTwitterInsights(ctx, batch); err != nil {
					log.Error().
						Err(err).
						Str("error_message", err.Error()).
						Int("batch_size", len(batch)).
						Str("function", "insightsBatchProcessor").
						Str("stage", "flush_final_batch").
						Msg("Failed to flush final insights batch")
					atomic.AddUint64(&metrics.InsertErrors, 1)
				}
			}
			return
		case <-flushTimer.C:
			if len(batch) > 0 {
				log.Info().Int("worker_id", id).Int("batch_size", len(batch)).Msg("Flushing Twitter insights batch on timer")
				if err := sink.BulkInsertTwitterInsights(ctx, batch); err != nil {
					log.Error().
						Err(err).
						Str("error_message", err.Error()).
						Int("batch_size", len(batch)).
						Str("function", "insightsBatchProcessor").
						Str("stage", "timer_flush").
						Msg("Failed to insert insights batch")
					atomic.AddUint64(&metrics.InsertErrors, 1)
				}
				batch = make([]*clickhousemodels.TwitterInsights, 0, 10000)
			}
			flushTimer.Reset(10 * time.Second)
		case item, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					if err := sink.BulkInsertTwitterInsights(ctx, batch); err != nil {
						log.Error().
							Err(err).
							Str("error_message", err.Error()).
							Int("batch_size", len(batch)).
							Str("function", "insightsBatchProcessor").
							Str("stage", "flush_final_batch").
							Msg("Failed to flush final insights batch")
						atomic.AddUint64(&metrics.InsertErrors, 1)
					}
				}
				return
			}
			batch = append(batch, item)
			if len(batch) >= 10000 {
				log.Info().Int("worker_id", id).Int("batch_size", len(batch)).Msg("Flushing Twitter insights batch on size threshold")
				if err := sink.BulkInsertTwitterInsights(ctx, batch); err != nil {
					log.Error().
						Err(err).
						Str("error_message", err.Error()).
						Int("batch_size", len(batch)).
						Str("function", "insightsBatchProcessor").
						Str("stage", "size_threshold_flush").
						Msg("Failed to insert insights batch")
					atomic.AddUint64(&metrics.InsertErrors, 1)
				}
				batch = make([]*clickhousemodels.TwitterInsights, 0, 10000)
				flushTimer.Reset(10 * time.Second)
			}
		}
	}
}
