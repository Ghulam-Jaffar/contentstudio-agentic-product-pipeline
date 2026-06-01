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

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	postsWorkers    = 6
	insightsWorkers = 6

	batchProcessorsPerType = 3

	maxBatchSize   = 1000
	batchTimeout   = 5 * time.Second
	idleFlushDelay = 100 * time.Millisecond

	messageChanSize = 50_000

	// Consumer groups (separate for pages and profiles)
	pageConsumerGroup    = "linkedin-page-clickhouse-sink-group"
	profileConsumerGroup = "linkedin-profile-clickhouse-sink-group"

	// Input topics for pages
	topicPagePosts    = "parsed-linkedin-page-posts"
	topicPageInsights = "parsed-linkedin-page-insights"

	// Input topics for profiles
	topicProfilePosts    = "parsed-linkedin-profile-posts"
	topicProfileInsights = "parsed-linkedin-profile-insights"

	// Deprecated: kept for backward compatibility
	consumerGroup = pageConsumerGroup
	topicPosts    = topicPagePosts
	topicInsights = topicPageInsights

	// idleTimeout is the duration after which the service will shutdown
	// if no new messages are received. This allows the service to exit
	// gracefully after batch processing is complete.
	idleTimeout       = 5 * time.Minute
	idleCheckInterval = 30 * time.Second
)

type Message struct {
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
		println("Failed to load config:", err.Error())
		os.Exit(1)
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().
		Str("page_consumer_group", pageConsumerGroup).
		Str("profile_consumer_group", profileConsumerGroup).
		Msg("Starting LinkedIn ClickHouse Sink Service with separate page/profile pipelines")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed (continuing)")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Page consumers
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

	// Profile consumers
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

	postsMsgChan := make(chan Message, messageChanSize)
	insightsMsgChan := make(chan Message, messageChanSize)

	var pickedPosts, pickedInsights uint64

	// Track last message time for idle timeout detection
	var lastMessageTime int64 = time.Now().UnixNano()

	var postsWg sync.WaitGroup
	for i := 0; i < postsWorkers; i++ {
		postsWg.Add(1)
		go postsWorker(ctx, i+1, postsMsgChan, batches, log, &postsWg)
	}

	var insightsWg sync.WaitGroup
	for i := 0; i < insightsWorkers; i++ {
		insightsWg.Add(1)
		go insightsWorker(ctx, i+1, insightsMsgChan, batches, log, &insightsWg)
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
					Int("queue_len", len(postsMsgChan)).
					Uint64("picked", atomic.LoadUint64(&pickedPosts)).
					Int("b_posts", len(batches.posts)).
					Msg("pipeline metrics")

				log.Info().
					Str("pipeline", "insights").
					Int("queue_len", len(insightsMsgChan)).
					Uint64("picked", atomic.LoadUint64(&pickedInsights)).
					Int("b_insights", len(batches.insights)).
					Msg("pipeline metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	var consumersWg sync.WaitGroup
	consumersWg.Add(4) // 2 for pages (posts + insights) + 2 for profiles (posts + insights)

	// Page posts consumer
	go func() {
		defer consumersWg.Done()
		topics := []string{topicPagePosts}
		log.Info().Strs("topics", topics).Str("consumer_group", pageConsumerGroup).Msg("Consuming page posts topics...")
		err := pagePostsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case postsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedPosts, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_page_posts").Str("pool", "page").Msg("Page posts consumer error")
			cancel()
		}
	}()

	// Page insights consumer
	go func() {
		defer consumersWg.Done()
		topics := []string{topicPageInsights}
		log.Info().Strs("topics", topics).Str("consumer_group", pageConsumerGroup).Msg("Consuming page insights topics...")
		err := pageInsightsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case insightsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedInsights, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_page_insights").Str("pool", "page").Msg("Page insights consumer error")
			cancel()
		}
	}()

	// Profile posts consumer
	go func() {
		defer consumersWg.Done()
		topics := []string{topicProfilePosts}
		log.Info().Strs("topics", topics).Str("consumer_group", profileConsumerGroup).Msg("Consuming profile posts topics...")
		err := profilePostsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case postsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedPosts, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_profile_posts").Str("pool", "profile").Msg("Profile posts consumer error")
			cancel()
		}
	}()

	// Profile insights consumer
	go func() {
		defer consumersWg.Done()
		topics := []string{topicProfileInsights}
		log.Info().Strs("topics", topics).Str("consumer_group", profileConsumerGroup).Msg("Consuming profile insights topics...")
		err := profileInsightsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case insightsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedInsights, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_profile_insights").Str("pool", "profile").Msg("Profile insights consumer error")
			cancel()
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Msg("LinkedIn ClickHouse Sink Service started successfully")

	<-sig
	log.Info().Msg("Shutdown signal received")

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

	log.Info().Msg("LinkedIn ClickHouse Sink Service stopped")
}

func postsWorker(ctx context.Context, workerID int, msgCh <-chan Message, b *BatchCollectors, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Info().Int("worker_id", workerID).Str("pool", "posts").Msg("worker started")

	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-msgCh:
			if !ok {
				return
			}
			if strings.HasSuffix(m.Topic, topicPosts) {
				_ = handleParsedPost(ctx, m.Key, m.Value, b, log)
			} else {
				log.Warn().Str("topic", m.Topic).Msg("posts worker: unknown topic")
			}
		}
	}
}

func insightsWorker(ctx context.Context, workerID int, msgCh <-chan Message, b *BatchCollectors, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Info().Int("worker_id", workerID).Str("pool", "insights").Msg("worker started")

	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-msgCh:
			if !ok {
				return
			}
			if m.Topic == topicPageInsights || m.Topic == topicProfileInsights {
				_ = handleParsedInsights(ctx, m.Key, m.Value, b, log)
			} else {
				log.Warn().Str("topic", m.Topic).Msg("insights worker: unknown topic")
			}
		}
	}
}

func handleParsedPost(ctx context.Context, key, value []byte, b *BatchCollectors, log *logger.Logger) error {
	var parsed kafkamodels.ParsedLinkedinPost
	if err := json.Unmarshal(value, &parsed); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleParsedPost").Str("stage", "unmarshal_post").Str("key", string(key)).Msg("unmarshal post failed")
		return err
	}
	select {
	case b.posts <- &parsed:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func handleParsedInsights(ctx context.Context, key, value []byte, b *BatchCollectors, log *logger.Logger) error {
	var parsed kafkamodels.ParsedLinkedinInsights
	if err := json.Unmarshal(value, &parsed); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleParsedInsights").Str("stage", "unmarshal_insights").Str("key", string(key)).Msg("unmarshal insights failed")
		return err
	}
	select {
	case b.insights <- &parsed:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func startBatchProcessors(ctx context.Context, b *BatchCollectors, sink *conversions.ClickHouseSink, log *logger.Logger, wg *sync.WaitGroup, n int) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processPostsBatch(ctx, b.posts, sink, log) }()
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processInsightsBatch(ctx, b.insights, sink, log) }()
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				log.Info().
					Int("q_posts", len(b.posts)).
					Int("q_insights", len(b.insights)).
					Msg("batch queues depth")
			}
		}
	}()
}

func processPostsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedLinkedinPost, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedLinkedinPost
	maxTimer := time.NewTimer(batchTimeout)
	idleTimer := time.NewTimer(idleFlushDelay)
	idleTimer.Stop()

	defer maxTimer.Stop()
	defer idleTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				processPosts(context.Background(), batch, sink, log)
			}
			return
		case x, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					processPosts(context.Background(), batch, sink, log)
				}
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatchSize {
				processPosts(ctx, batch, sink, log)
				batch = nil
				maxTimer.Reset(batchTimeout)
				idleTimer.Stop()
			} else {
				idleTimer.Reset(idleFlushDelay)
			}
		case <-idleTimer.C:
			if len(batch) > 0 {
				processPosts(ctx, batch, sink, log)
				batch = nil
				maxTimer.Reset(batchTimeout)
			}
		case <-maxTimer.C:
			if len(batch) > 0 {
				processPosts(ctx, batch, sink, log)
				batch = nil
			}
			maxTimer.Reset(batchTimeout)
		}
	}
}

func processInsightsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedLinkedinInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedLinkedinInsights
	maxTimer := time.NewTimer(batchTimeout)
	idleTimer := time.NewTimer(idleFlushDelay)
	idleTimer.Stop()

	defer maxTimer.Stop()
	defer idleTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				processInsights(context.Background(), batch, sink, log)
			}
			return
		case x, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					processInsights(context.Background(), batch, sink, log)
				}
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatchSize {
				processInsights(ctx, batch, sink, log)
				batch = nil
				maxTimer.Reset(batchTimeout)
				idleTimer.Stop()
			} else {
				idleTimer.Reset(idleFlushDelay)
			}
		case <-idleTimer.C:
			if len(batch) > 0 {
				processInsights(ctx, batch, sink, log)
				batch = nil
				maxTimer.Reset(batchTimeout)
			}
		case <-maxTimer.C:
			if len(batch) > 0 {
				processInsights(ctx, batch, sink, log)
				batch = nil
			}
			maxTimer.Reset(batchTimeout)
		}
	}
}

func processPosts(ctx context.Context, batch []*kafkamodels.ParsedLinkedinPost, sink *conversions.ClickHouseSink, log *logger.Logger) {
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
	if len(rows) == 0 {
		return
	}
	startTime := time.Now()
	if err := sink.BulkInsertLinkedInPosts(ctx, rows); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processPosts").Str("stage", "bulk_insert_posts").Int("count", len(rows)).Dur("duration", time.Since(startTime)).Msg("bulk insert posts failed")
	} else {
		log.Info().Int("count", len(rows)).Dur("duration", time.Since(startTime)).Msg("bulk inserted linkedin posts")
	}
}

func processInsights(ctx context.Context, batch []*kafkamodels.ParsedLinkedinInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
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
	if len(rows) == 0 {
		return
	}
	startTime := time.Now()
	if err := sink.BulkInsertLinkedInInsights(ctx, rows); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processInsights").Str("stage", "bulk_insert_insights").Int("count", len(rows)).Dur("duration", time.Since(startTime)).Msg("bulk insert insights failed")
	} else {
		log.Info().Int("count", len(rows)).Dur("duration", time.Since(startTime)).Msg("bulk inserted linkedin insights")
	}
}
