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
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

const (
	postsParserWorkers    = 5
	insightsParserWorkers = 5

	batchProcessorsPerType = 3

	maxBatchSize   = 10000
	batchTimeout   = 10 * time.Second
	idleFlushDelay = 100 * time.Millisecond

	messageChanSize = 50000

	pageConsumerGroup    = "linkedin-page-parser-group"
	profileConsumerGroup = "linkedin-profile-parser-group"

	// Raw input topics for pages
	topicRawPagePosts    = "raw-linkedin-page-posts"
	topicRawPageInsights = "raw-linkedin-page-insights"

	// Raw input topics for profiles
	topicRawProfilePosts    = "raw-linkedin-profile-posts"
	topicRawProfileInsights = "raw-linkedin-profile-insights"

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
	posts    chan *kafkamodels.ParsedLinkedinPost
	insights chan *kafkamodels.ParsedLinkedinInsights
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("posts_parser_workers", postsParserWorkers).
		Int("insights_parser_workers", insightsParserWorkers).
		Int("batch_processors", batchProcessorsPerType).
		Str("page_consumer_group", pageConsumerGroup).
		Str("profile_consumer_group", profileConsumerGroup).
		Msg("Starting LinkedIn Analytics Sink (merged parser+sink)")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed (continuing)")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Consumers for raw topics - use page group for page topics
	pagePostsConsumer, err := kafka.NewConsumer(cfg.Kafka, pageConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create page posts consumer")
	}
	defer pagePostsConsumer.Close()

	pageInsightsConsumer, err := kafka.NewConsumer(cfg.Kafka, pageConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create page insights consumer")
	}
	defer pageInsightsConsumer.Close()

	// Use profile group for profile topics
	profilePostsConsumer, err := kafka.NewConsumer(cfg.Kafka, profileConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create profile posts consumer")
	}
	defer profilePostsConsumer.Close()

	profileInsightsConsumer, err := kafka.NewConsumer(cfg.Kafka, profileConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create profile insights consumer")
	}
	defer profileInsightsConsumer.Close()

	batches := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedLinkedinPost, maxBatchSize*5),
		insights: make(chan *kafkamodels.ParsedLinkedinInsights, maxBatchSize*5),
	}

	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, batches, sink, log, &batchWg, batchProcessorsPerType)

	postsMsgChan := make(chan RawMessage, messageChanSize)
	insightsMsgChan := make(chan RawMessage, messageChanSize)

	var pickedPosts, pickedInsights uint64
	var parsedPosts, parsedInsights uint64

	// Track last message time for idle timeout detection
	var lastMessageTime int64 = time.Now().UnixNano()

	// Start parser workers
	var postsWg sync.WaitGroup
	for i := 0; i < postsParserWorkers; i++ {
		postsWg.Add(1)
		go postsParserWorker(ctx, i+1, postsMsgChan, batches, &parsedPosts, log, &postsWg)
	}

	var insightsWg sync.WaitGroup
	for i := 0; i < insightsParserWorkers; i++ {
		insightsWg.Add(1)
		go insightsParserWorker(ctx, i+1, insightsMsgChan, batches, &parsedInsights, log, &insightsWg)
	}

	// Metrics logging
	stopMetrics := make(chan struct{})
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info().
					Str("pipeline", "posts").
					Int("queue_raw", len(postsMsgChan)).
					Uint64("picked", atomic.LoadUint64(&pickedPosts)).
					Uint64("parsed", atomic.LoadUint64(&parsedPosts)).
					Int("batch_queue", len(batches.posts)).
					Msg("pipeline metrics")

				log.Info().
					Str("pipeline", "insights").
					Int("queue_raw", len(insightsMsgChan)).
					Uint64("picked", atomic.LoadUint64(&pickedInsights)).
					Uint64("parsed", atomic.LoadUint64(&parsedInsights)).
					Int("batch_queue", len(batches.insights)).
					Msg("pipeline metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	var consumersWg sync.WaitGroup
	consumersWg.Add(4)

	// Page posts consumer
	go func() {
		defer consumersWg.Done()
		topics := []string{topicRawPagePosts}
		log.Info().Strs("topics", topics).Msg("Starting page posts consumer")
		err := pagePostsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case postsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedPosts, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_page_posts").Msg("Page posts consumer error")
		}
	}()

	// Page insights consumer
	go func() {
		defer consumersWg.Done()
		topics := []string{topicRawPageInsights}
		log.Info().Strs("topics", topics).Msg("Starting page insights consumer")
		err := pageInsightsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case insightsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedInsights, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_page_insights").Msg("Page insights consumer error")
		}
	}()

	// Profile posts consumer
	go func() {
		defer consumersWg.Done()
		topics := []string{topicRawProfilePosts}
		log.Info().Strs("topics", topics).Msg("Starting profile posts consumer")
		err := profilePostsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case postsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedPosts, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_profile_posts").Msg("Profile posts consumer error")
		}
	}()

	// Profile insights consumer
	go func() {
		defer consumersWg.Done()
		topics := []string{topicRawProfileInsights}
		log.Info().Strs("topics", topics).Msg("Starting profile insights consumer")
		err := profileInsightsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case insightsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedInsights, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_profile_insights").Msg("Profile insights consumer error")
		}
	}()

	// Handle shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Int("posts_parser_workers", postsParserWorkers).
		Int("insights_parser_workers", insightsParserWorkers).
		Int("batch_processors", batchProcessorsPerType).
		Msg("LinkedIn Analytics Sink started successfully")

	<-sig
	log.Info().Msg("Shutdown signal received, stopping...")

	cancel()
	consumersWg.Wait()

	close(postsMsgChan)
	close(insightsMsgChan)

	postsWg.Wait()
	insightsWg.Wait()

	close(batches.posts)
	close(batches.insights)

	batchWg.Wait()
	close(stopMetrics)

	log.Info().Msg("LinkedIn Analytics Sink stopped")
}

func postsParserWorker(ctx context.Context, workerID int, msgCh <-chan RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	workerLog := log.With().Str("pool", "posts").Int("worker_id", workerID).Logger()
	workerLog.Info().Msg("Parser worker started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Parser worker stopped (context cancelled)")
			return
		case m, ok := <-msgCh:
			if !ok {
				workerLog.Info().Msg("Parser worker stopped (channel closed)")
				return
			}
			parseAndQueuePost(ctx, m, b, counter, &logger.Logger{Logger: workerLog})
		}
	}
}

func insightsParserWorker(ctx context.Context, workerID int, msgCh <-chan RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	workerLog := log.With().Str("pool", "insights").Int("worker_id", workerID).Logger()
	workerLog.Info().Msg("Parser worker started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Parser worker stopped (context cancelled)")
			return
		case m, ok := <-msgCh:
			if !ok {
				workerLog.Info().Msg("Parser worker stopped (channel closed)")
				return
			}
			parseAndQueueInsights(ctx, m, b, counter, &logger.Logger{Logger: workerLog})
		}
	}
}

func parseAndQueuePost(ctx context.Context, msg RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger) {
	linkedinID := string(msg.Key)

	parsed, err := parsing.ParsePost(json.RawMessage(msg.Value))
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseAndQueuePost").Str("stage", "parse_post").Str("linkedin_id", linkedinID).Msg("Failed to parse post")
		return
	}
	if parsed == nil {
		return
	}

	if parsed.LinkedinID == "" {
		parsed.LinkedinID = linkedinID
	}

	select {
	case b.posts <- parsed:
		atomic.AddUint64(counter, 1)
	case <-ctx.Done():
		return
	}
}

func parseAndQueueInsights(ctx context.Context, msg RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger) {
	linkedinID := string(msg.Key)

	parsedList, err := parsing.ParseInsightsDaily(json.RawMessage(msg.Value))
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseAndQueueInsights").Str("stage", "parse_insights").Str("linkedin_id", linkedinID).Msg("Failed to parse insights")
		return
	}
	if len(parsedList) == 0 {
		return
	}

	for _, parsed := range parsedList {
		if parsed.LinkedinID == "" {
			parsed.LinkedinID = linkedinID
		}
		parsed.RecordID = parsed.LinkedinID + "_" + parsed.CreatedAt.Format("2006-01-02")

		select {
		case b.insights <- parsed:
			atomic.AddUint64(counter, 1)
		case <-ctx.Done():
			return
		}
	}
}

func startBatchProcessors(ctx context.Context, b *BatchCollectors, sink *conversions.ClickHouseSink, log *logger.Logger, wg *sync.WaitGroup, n int) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processPostsBatch(ctx, id, b.posts, sink, log)
		}(i)
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processInsightsBatch(ctx, id, b.insights, sink, log)
		}(i)
	}
}

func processPostsBatch(ctx context.Context, id int, in <-chan *kafkamodels.ParsedLinkedinPost, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedLinkedinPost
	maxTimer := time.NewTimer(batchTimeout)
	idleTimer := time.NewTimer(idleFlushDelay)
	idleTimer.Stop()

	defer maxTimer.Stop()
	defer idleTimer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		rows := make([]*chmodels.LinkedInPosts, 0, len(batch))
		for _, x := range batch {
			row := conversions.ConvertLinkedInPost(x)
			if row != nil {
				rows = append(rows, row)
			}
		}
		if len(rows) > 0 {
			startTime := time.Now()
			if err := sink.BulkInsertLinkedInPosts(ctx, rows); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processPostsBatch").Str("stage", "bulk_insert_posts").Int("count", len(rows)).Msg("bulk insert posts failed")
			} else {
				log.Info().Int("processor", id).Int("count", len(rows)).Dur("duration", time.Since(startTime)).Msg("inserted linkedin posts")
			}
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
			if len(batch) >= maxBatchSize {
				flush()
				maxTimer.Reset(batchTimeout)
				idleTimer.Stop()
			} else {
				idleTimer.Reset(idleFlushDelay)
			}
		case <-idleTimer.C:
			flush()
			maxTimer.Reset(batchTimeout)
		case <-maxTimer.C:
			flush()
			maxTimer.Reset(batchTimeout)
		}
	}
}

func processInsightsBatch(ctx context.Context, id int, in <-chan *kafkamodels.ParsedLinkedinInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedLinkedinInsights
	maxTimer := time.NewTimer(batchTimeout)
	idleTimer := time.NewTimer(idleFlushDelay)
	idleTimer.Stop()

	defer maxTimer.Stop()
	defer idleTimer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		rows := make([]*chmodels.LinkedInInsights, 0, len(batch))
		for _, x := range batch {
			row := conversions.ConvertLinkedInInsights(x)
			if row != nil {
				rows = append(rows, row)
			}
		}
		if len(rows) > 0 {
			startTime := time.Now()
			if err := sink.BulkInsertLinkedInInsights(ctx, rows); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processInsightsBatch").Str("stage", "bulk_insert_insights").Int("count", len(rows)).Msg("bulk insert insights failed")
			} else {
				log.Info().Int("processor", id).Int("count", len(rows)).Dur("duration", time.Since(startTime)).Msg("inserted linkedin insights")
			}
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
			if len(batch) >= maxBatchSize {
				flush()
				maxTimer.Reset(batchTimeout)
				idleTimer.Stop()
			} else {
				idleTimer.Reset(idleFlushDelay)
			}
		case <-idleTimer.C:
			flush()
			maxTimer.Reset(batchTimeout)
		case <-maxTimer.C:
			flush()
			maxTimer.Reset(batchTimeout)
		}
	}
}
