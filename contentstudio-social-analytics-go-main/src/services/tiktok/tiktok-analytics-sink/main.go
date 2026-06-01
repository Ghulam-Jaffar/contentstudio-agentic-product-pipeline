package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	maxBatchSize           = 10000
	batchTimeout           = 10 * time.Second
	batchProcessorsPerType = 3
	messageChanSize        = 50000

	rawPostsTopic    = "raw-tiktok-posts"
	rawInsightsTopic = "raw-tiktok-insights"

	consumerGroup = "tiktok-analytics-sink-group"

	// idleTimeout is the duration after which the service will shutdown
	// if no new messages are received. This allows the service to exit
	// gracefully after batch processing is complete.
	idleTimeout       = 5 * time.Minute
	idleCheckInterval = 30 * time.Second
)

type RawMessage struct {
	Topic string
	Key   []byte
	Value []byte
}

type BatchCollectors struct {
	posts    chan *chmodels.TikTokPosts
	insights chan *chmodels.TikTokInsights
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting TikTok Analytics Sink (merged parser+sink)")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed - continuing anyway")
	}

	postsConsumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create posts consumer")
	}
	defer postsConsumer.Close()

	insightsConsumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create insights consumer")
	}
	defer insightsConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())

	postsJobs := make(chan RawMessage, messageChanSize)
	insightsJobs := make(chan RawMessage, messageChanSize)

	batches := &BatchCollectors{
		posts:    make(chan *chmodels.TikTokPosts, maxBatchSize*5),
		insights: make(chan *chmodels.TikTokInsights, maxBatchSize*5),
	}

	var pickedPosts, pickedInsights uint64
	var parsedPosts, parsedInsights uint64
	var insertedPosts, insertedInsights uint64

	var lastMessageTime int64 = time.Now().UnixNano()

	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, batches, sink, log, &batchWg, batchProcessorsPerType, &insertedPosts, &insertedInsights)

	var wgParsers sync.WaitGroup
	for i := 0; i < 5; i++ {
		wgParsers.Add(1)
		go postsParser(ctx, &wgParsers, i, postsJobs, batches.posts, &parsedPosts, log)
	}
	for i := 0; i < 5; i++ {
		wgParsers.Add(1)
		go insightsParser(ctx, &wgParsers, i, insightsJobs, batches.insights, &parsedInsights, log)
	}

	stopMetrics := make(chan struct{})
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info().
					Str("pipeline", "posts").
					Int("parse_queue", len(postsJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedPosts)).
					Uint64("parsed", atomic.LoadUint64(&parsedPosts)).
					Uint64("inserted", atomic.LoadUint64(&insertedPosts)).
					Int("batch_queue", len(batches.posts)).
					Msg("pipeline metrics")
				log.Info().
					Str("pipeline", "insights").
					Int("parse_queue", len(insightsJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedInsights)).
					Uint64("parsed", atomic.LoadUint64(&parsedInsights)).
					Uint64("inserted", atomic.LoadUint64(&insertedInsights)).
					Int("batch_queue", len(batches.insights)).
					Msg("pipeline metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	var wgConsumers sync.WaitGroup
	wgConsumers.Add(2)

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", rawPostsTopic).Str("group", consumerGroup).Msg("Consuming posts topic...")
		err := postsConsumer.Consume(ctx, []string{rawPostsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			atomic.AddUint64(&pickedPosts, 1)
			postsJobs <- RawMessage{Topic: topic, Key: key, Value: value}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_posts").Msg("Posts consumer error")
			cancel()
		}
	}()

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", rawInsightsTopic).Str("group", consumerGroup).Msg("Consuming insights topic...")
		err := insightsConsumer.Consume(ctx, []string{rawInsightsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			atomic.AddUint64(&pickedInsights, 1)
			insightsJobs <- RawMessage{Topic: topic, Key: key, Value: value}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_insights").Msg("Insights consumer error")
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Int("posts_parser_workers", 5).
		Int("insights_parser_workers", 5).
		Int("batch_processors_per_type", batchProcessorsPerType).
		Int("max_batch_size", maxBatchSize).
		Dur("batch_timeout", batchTimeout).
		Msg("TikTok Analytics Sink started successfully")

	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping...")
	cancel()

	wgConsumers.Wait()

	close(postsJobs)
	close(insightsJobs)

	wgParsers.Wait()

	close(batches.posts)
	close(batches.insights)

	batchWg.Wait()

	close(stopMetrics)

	log.Info().
		Uint64("total_picked_posts", atomic.LoadUint64(&pickedPosts)).
		Uint64("total_picked_insights", atomic.LoadUint64(&pickedInsights)).
		Uint64("total_parsed_posts", atomic.LoadUint64(&parsedPosts)).
		Uint64("total_parsed_insights", atomic.LoadUint64(&parsedInsights)).
		Uint64("total_inserted_posts", atomic.LoadUint64(&insertedPosts)).
		Uint64("total_inserted_insights", atomic.LoadUint64(&insertedInsights)).
		Msg("TikTok Analytics Sink stopped")
}

func postsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, out chan<- *chmodels.TikTokPosts, parsedCounter *uint64, log *logger.Logger) {
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

			var rawPost kafkamodels.RawTikTokPost
			if err := json.Unmarshal(msg.Value, &rawPost); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "postsParser").Str("stage", "unmarshal_raw_post").Str("key", string(msg.Key)).Msg("Failed to unmarshal raw post")
				continue
			}

			// rawPost.Data contains the already-parsed ParsedTikTokPost from the fetcher
			// Unmarshal it directly instead of parsing the raw video data again
			var parsedPost kafkamodels.ParsedTikTokPost
			if err := json.Unmarshal(rawPost.Data, &parsedPost); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "postsParser").Str("stage", "unmarshal_parsed_post").Str("key", string(msg.Key)).Msg("Failed to unmarshal parsed post from data")
				continue
			}

			// Skip if parsing produced no data
			if parsedPost.ID == "" {
				log.Debug().Str("tiktok_id", rawPost.TikTokID).Msg("Parsed post has empty ID, skipping")
				continue
			}

			chPost := conversions.ConvertTikTokPost(&parsedPost)
			if chPost != nil {
				select {
				case out <- chPost:
					atomic.AddUint64(parsedCounter, 1)
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func insightsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, out chan<- *chmodels.TikTokInsights, parsedCounter *uint64, log *logger.Logger) {
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

			// Insights topic publishes ParsedTikTokInsights directly from the fetcher
			// No need to regenerate, just convert to ClickHouse format
			var parsedInsight kafkamodels.ParsedTikTokInsights
			if err := json.Unmarshal(msg.Value, &parsedInsight); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "insightsParser").Str("stage", "unmarshal_parsed_insight").Str("key", string(msg.Key)).Msg("Failed to unmarshal parsed insight")
				continue
			}

			// Skip if insight is empty or missing required fields
			if parsedInsight.RecordID == "" {
				log.Debug().Str("key", string(msg.Key)).Msg("Parsed insight has empty RecordID, skipping")
				continue
			}

			chInsight := conversions.ConvertTikTokInsights(&parsedInsight)
			if chInsight == nil {
				log.Debug().Str("record_id", parsedInsight.RecordID).Msg("Failed to convert insight to ClickHouse format, skipping")
				continue
			}

			select {
			case out <- chInsight:
				atomic.AddUint64(parsedCounter, 1)
			case <-ctx.Done():
				return
			}
		}
	}
}

func startBatchProcessors(ctx context.Context, batches *BatchCollectors, sink *conversions.ClickHouseSink, log *logger.Logger, wg *sync.WaitGroup, numProcessors int, insertedPosts, insertedInsights *uint64) {
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go postsBatchProcessor(ctx, wg, i, batches.posts, sink, log, insertedPosts)
	}
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go insightsBatchProcessor(ctx, wg, i, batches.insights, sink, log, insertedInsights)
	}
}

func postsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.TikTokPosts, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	log.Info().Int("processor_id", id).Str("batch_type", "posts").Msg("Batch processor started")

	batch := make([]*chmodels.TikTokPosts, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				if err := sink.BulkInsertTikTokPosts(ctx, batch); err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("function", "postsBatchProcessor").Str("stage", "flush_final_posts").Msg("Failed to flush final batch of TikTok posts")
				} else {
					atomic.AddUint64(insertedCounter, uint64(len(batch)))
				}
			}
			return

		case <-flushTimer.C:
			if len(batch) > 0 {
				if err := sink.BulkInsertTikTokPosts(ctx, batch); err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("function", "postsBatchProcessor").Str("stage", "flush_timer_posts").Int("processor_id", id).Int("batch_size", len(batch)).Msg("Failed to insert posts batch")
				} else {
					atomic.AddUint64(insertedCounter, uint64(len(batch)))
				}
				batch = make([]*chmodels.TikTokPosts, 0, maxBatchSize)
			}
			flushTimer.Reset(batchTimeout)

		case item, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					if err := sink.BulkInsertTikTokPosts(ctx, batch); err != nil {
						log.Error().Err(err).Str("error_message", err.Error()).Str("function", "postsBatchProcessor").Str("stage", "flush_final_posts_close").Msg("Failed to flush final batch of posts")
					} else {
						atomic.AddUint64(insertedCounter, uint64(len(batch)))
					}
				}
				return
			}

			batch = append(batch, item)
			if len(batch) >= maxBatchSize {
				if err := sink.BulkInsertTikTokPosts(ctx, batch); err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("function", "postsBatchProcessor").Str("stage", "bulk_insert_posts").Int("processor_id", id).Int("batch_size", len(batch)).Msg("Failed to insert posts batch")
				} else {
					atomic.AddUint64(insertedCounter, uint64(len(batch)))
				}
				batch = make([]*chmodels.TikTokPosts, 0, maxBatchSize)
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

func insightsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.TikTokInsights, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	log.Info().Int("processor_id", id).Str("batch_type", "insights").Msg("Batch processor started")

	batch := make([]*chmodels.TikTokInsights, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	// Helper function to update insights with database total views
	updateTotalViewsFromDatabase := func(insights *chmodels.TikTokInsights) {
		databaseTotalViews, err := sink.RawClient.GetTikTokPostsViewSum(ctx, insights.TikTokID)
		if err != nil {
			log.Warn().Err(err).Str("tiktok_id", insights.TikTokID).Msg("Failed to query total views from database, keeping insight value")
			return
		}
		log.Debug().Str("tiktok_id", insights.TikTokID).Int64("database_total_views", databaseTotalViews).Msg("Updated insight with database total views")
		insights.TotalVideoViews = databaseTotalViews
	}

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				// Update each insight with database totals before final flush
				for _, insight := range batch {
					updateTotalViewsFromDatabase(insight)
				}
				if err := sink.BulkInsertTikTokInsights(ctx, batch); err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("function", "insightsBatchProcessor").Str("stage", "flush_final_insights").Msg("Failed to flush final batch of TikTok insights")
				} else {
					atomic.AddUint64(insertedCounter, uint64(len(batch)))
				}
			}
			return

		case <-flushTimer.C:
			if len(batch) > 0 {
				// Update each insight with database totals before flush
				for _, insight := range batch {
					updateTotalViewsFromDatabase(insight)
				}
				if err := sink.BulkInsertTikTokInsights(ctx, batch); err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("function", "insightsBatchProcessor").Str("stage", "flush_timer_insights").Int("processor_id", id).Int("batch_size", len(batch)).Msg("Failed to insert insights batch")
				} else {
					atomic.AddUint64(insertedCounter, uint64(len(batch)))
				}
				batch = make([]*chmodels.TikTokInsights, 0, maxBatchSize)
			}
			flushTimer.Reset(batchTimeout)

		case item, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					// Update each insight with database totals before final flush
					for _, insight := range batch {
						updateTotalViewsFromDatabase(insight)
					}
					if err := sink.BulkInsertTikTokInsights(ctx, batch); err != nil {
						log.Error().Err(err).Str("error_message", err.Error()).Str("function", "insightsBatchProcessor").Str("stage", "flush_final_insights_close").Msg("Failed to flush final batch of insights")
					} else {
						atomic.AddUint64(insertedCounter, uint64(len(batch)))
					}
				}
				return
			}

			batch = append(batch, item)
			if len(batch) >= maxBatchSize {
				// Update each insight with database totals before flush
				for _, insight := range batch {
					updateTotalViewsFromDatabase(insight)
				}
				if err := sink.BulkInsertTikTokInsights(ctx, batch); err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("function", "insightsBatchProcessor").Str("stage", "bulk_insert_insights").Int("processor_id", id).Int("batch_size", len(batch)).Msg("Failed to insert insights batch")
				} else {
					atomic.AddUint64(insertedCounter, uint64(len(batch)))
				}
				batch = make([]*chmodels.TikTokInsights, 0, maxBatchSize)
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}
