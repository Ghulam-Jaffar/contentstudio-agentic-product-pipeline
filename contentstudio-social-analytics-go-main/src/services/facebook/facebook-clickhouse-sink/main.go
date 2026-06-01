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
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	// worker threads reading from Kafka → routing to batch queues
	postsAssetsWorkers = 6
	insightsWorkers    = 6

	// batch processors per data type (run in parallel; each does timed/size-based flush)
	batchProcessorsPerType = 3

	// batching
	maxBatchSize = 10000
	// batchTimeout controls how long to wait before flushing a partial batch.
	// Reduced from 180s to 10s for faster data visibility in dashboards.
	// Trade-off: More frequent smaller inserts vs fewer larger batches.
	batchTimeout = 10 * time.Second

	// channels
	messageChanSize = 50_000 // buffer per pipeline

	// consumer group
	consumerGroup = "facebook-clickhouse-sink-group"

	// idleTimeout is the duration after which the service will shutdown
	// if no new messages are received. This allows the service to exit
	// gracefully after batch processing is complete.
	idleTimeout       = 5 * time.Minute
	idleCheckInterval = 30 * time.Second

	// topics (exact names)
	topicPosts         = "parsed-facebook-posts"
	topicMediaAssets   = "parsed-facebook-media-assets"
	topicVideoInsights = "parsed-facebook-video-insights"
	topicReelsInsights = "parsed-facebook-reels-insights"
	topicInsights      = "parsed-facebook-insights"
)

type Message struct {
	Topic string
	Key   []byte
	Value []byte
}

// Batch collectors (ClickHouse batch queues)
type BatchCollectors struct {
	// posts + assets pipeline
	posts       chan *kafkamodels.ParsedFacebookPost
	mediaAssets chan *kafkamodels.ParsedFacebookMediaAsset

	// insights pipeline (all insights together)
	pageInsights  chan *kafkamodels.ParsedFacebookInsights
	videoInsights chan *kafkamodels.ParsedFacebookVideoInsights
	reelsInsights chan *kafkamodels.ParsedFacebookReelsInsights
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		println("Failed to load config:", err.Error())
		os.Exit(1)
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Facebook ClickHouse Sink Service")

	// ClickHouse sink
	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed (continuing)")
	}

	// ---- Context & shutdown ----
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ---- Consumers (separated by pipeline) ----
	// Pipeline 1: posts + media assets
	postsAssetsTopics := []string{topicPosts, topicMediaAssets}

	// Pipeline 2: all insights together (page, video, reels)
	insightsTopics := []string{topicInsights, topicVideoInsights, topicReelsInsights}

	postsConsumer, err := kafka.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create posts/assets consumer")
	}
	defer postsConsumer.Close()

	insightsConsumer, err := kafka.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create insights consumer")
	}
	defer insightsConsumer.Close()

	// ---- Batch queues ----
	batches := &BatchCollectors{
		posts:         make(chan *kafkamodels.ParsedFacebookPost, maxBatchSize*5),
		mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, maxBatchSize*5),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, maxBatchSize*5),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, maxBatchSize*5),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, maxBatchSize*5),
	}

	// ---- Start batch processors for each data type ----
	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, batches, sink, log, &batchWg, batchProcessorsPerType)

	// ---- Independent message channels & worker pools ----
	postsAssetsMsgChan := make(chan Message, messageChanSize)
	insightsMsgChan := make(chan Message, messageChanSize)

	// metrics
	var pickedPostsAssets, pickedInsights uint64

	// Track last message time for idle timeout detection
	var lastMessageTime int64 = time.Now().UnixNano()

	// workers: posts + assets
	var postsAssetsWg sync.WaitGroup
	for i := 0; i < postsAssetsWorkers; i++ {
		postsAssetsWg.Add(1)
		go postsAssetsWorker(ctx, i+1, postsAssetsMsgChan, batches, log, &postsAssetsWg)
	}

	// workers: all insights
	var insightsWg sync.WaitGroup
	for i := 0; i < insightsWorkers; i++ {
		insightsWg.Add(1)
		go insightsWorker(ctx, i+1, insightsMsgChan, batches, log, &insightsWg)
	}

	// ---- Metrics ticker ----
	stopMetrics := make(chan struct{})
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info().
					Str("pipeline", "posts_assets").
					Int("queue_len", len(postsAssetsMsgChan)).
					Uint64("picked", atomic.LoadUint64(&pickedPostsAssets)).
					Int("b_posts", len(batches.posts)).
					Int("b_media", len(batches.mediaAssets)).
					Msg("pipeline metrics")

				log.Info().
					Str("pipeline", "insights").
					Int("queue_len", len(insightsMsgChan)).
					Uint64("picked", atomic.LoadUint64(&pickedInsights)).
					Int("b_page", len(batches.pageInsights)).
					Int("b_video", len(batches.videoInsights)).
					Int("b_reels", len(batches.reelsInsights)).
					Msg("pipeline metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	// ---- Start both consumers ----
	var consumersWg sync.WaitGroup
	consumersWg.Add(2)

	// posts + assets consumer
	go func() {
		defer consumersWg.Done()
		log.Info().Strs("topics", postsAssetsTopics).Msg("Consuming posts/assets topics...")
		err := postsConsumer.Consume(ctx, postsAssetsTopics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case postsAssetsMsgChan <- Message{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedPostsAssets, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_posts_assets").Msg("Posts/assets consumer error")
			cancel()
		}
	}()

	// insights consumer (page + video + reels)
	go func() {
		defer consumersWg.Done()
		log.Info().Strs("topics", insightsTopics).Msg("Consuming insights topics...")
		err := insightsConsumer.Consume(ctx, insightsTopics, func(ctx context.Context, topic string, key, value []byte) error {
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
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_insights").Msg("Insights consumer error")
			cancel()
		}
	}()

	// ---- Shutdown handling ----
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Msg("Facebook ClickHouse Sink Service started successfully")

	<-sig
	log.Info().Msg("Shutdown signal received")

	// stop consumers
	cancel()
	consumersWg.Wait()

	// close message channels → stop routing workers
	close(postsAssetsMsgChan)
	close(insightsMsgChan)

	// wait routing workers to finish
	postsAssetsWg.Wait()
	insightsWg.Wait()

	// close batch input channels → stop batch processors
	close(batches.posts)
	close(batches.mediaAssets)
	close(batches.pageInsights)
	close(batches.videoInsights)
	close(batches.reelsInsights)

	// wait batch processors
	batchWg.Wait()

	// stop metrics
	close(stopMetrics)

	log.Info().Msg("Facebook ClickHouse Sink Service stopped")
}

// ===================== Workers: route messages → batch queues =====================

func postsAssetsWorker(ctx context.Context, workerID int, msgCh <-chan Message, b *BatchCollectors, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	log.Info().Int("worker_id", workerID).Str("pool", "posts-assets").Msg("worker started")

	for {
		select {
		case <-ctx.Done():
			return
		case m, ok := <-msgCh:
			if !ok {
				return
			}
			switch {
			case strings.HasSuffix(m.Topic, topicPosts):
				_ = handleParsedPost(ctx, m.Key, m.Value, b, log)
			case strings.HasSuffix(m.Topic, topicMediaAssets):
				_ = handleParsedMediaAsset(ctx, m.Key, m.Value, b, log)
			default:
				log.Warn().Str("topic", m.Topic).Msg("posts-assets worker: unknown topic")
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
			switch {
			case strings.HasSuffix(m.Topic, topicInsights):
				_ = handleParsedPageInsights(ctx, m.Key, m.Value, b, log)
			case strings.HasSuffix(m.Topic, topicVideoInsights):
				_ = handleParsedVideoInsights(ctx, m.Key, m.Value, b, log)
			case strings.HasSuffix(m.Topic, topicReelsInsights):
				_ = handleParsedReelsInsights(ctx, m.Key, m.Value, b, log)
			default:
				log.Warn().Str("topic", m.Topic).Msg("insights worker: unknown topic")
			}
		}
	}
}

// ===================== Consumer Handlers → Batch Queues =====================

func handleParsedPost(ctx context.Context, key, value []byte, b *BatchCollectors, log *logger.Logger) error {
	var parsed kafkamodels.ParsedFacebookPost
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

func handleParsedMediaAsset(ctx context.Context, key, value []byte, b *BatchCollectors, log *logger.Logger) error {
	var parsed kafkamodels.ParsedFacebookMediaAsset
	if err := json.Unmarshal(value, &parsed); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleParsedMediaAsset").Str("stage", "unmarshal_media_asset").Str("key", string(key)).Msg("unmarshal media asset failed")
		return err
	}
	select {
	case b.mediaAssets <- &parsed:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func handleParsedPageInsights(ctx context.Context, key, value []byte, b *BatchCollectors, log *logger.Logger) error {
	// Parse as batch (slice of daily records)
	var parsedBatch []*kafkamodels.ParsedFacebookInsights
	if err := json.Unmarshal(value, &parsedBatch); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleParsedPageInsights").Str("stage", "unmarshal_page_insights").Str("key", string(key)).Msg("unmarshal page insights batch failed")
		return err
	}

	// Send each daily record to the batch channel
	for _, parsed := range parsedBatch {
		select {
		case b.pageInsights <- parsed:
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	log.Debug().
		Str("page_id", string(key)).
		Int("daily_records", len(parsedBatch)).
		Msg("Received page insights batch")

	return nil
}

func handleParsedVideoInsights(ctx context.Context, key, value []byte, b *BatchCollectors, log *logger.Logger) error {
	var parsed kafkamodels.ParsedFacebookVideoInsights
	if err := json.Unmarshal(value, &parsed); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleParsedVideoInsights").Str("stage", "unmarshal_video_insights").Str("key", string(key)).Msg("unmarshal video insights failed")
		return err
	}
	select {
	case b.videoInsights <- &parsed:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func handleParsedReelsInsights(ctx context.Context, key, value []byte, b *BatchCollectors, log *logger.Logger) error {
	var parsed kafkamodels.ParsedFacebookReelsInsights
	if err := json.Unmarshal(value, &parsed); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleParsedReelsInsights").Str("stage", "unmarshal_reels_insights").Str("key", string(key)).Msg("unmarshal reels insights failed")
		return err
	}
	select {
	case b.reelsInsights <- &parsed:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// ===================== Start Batch Processors =====================

func startBatchProcessors(ctx context.Context, b *BatchCollectors, sink *conversions.ClickHouseSink, log *logger.Logger, wg *sync.WaitGroup, n int) {
	// posts
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processPostsBatch(ctx, b.posts, sink, log) }()
	}
	// media assets
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processMediaAssetsBatch(ctx, b.mediaAssets, sink, log) }()
	}
	// page insights
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processPageInsightsBatch(ctx, b.pageInsights, sink, log) }()
	}
	// video insights
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processVideoInsightsBatch(ctx, b.videoInsights, sink, log) }()
	}
	// reels insights
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processReelsInsightsBatch(ctx, b.reelsInsights, sink, log) }()
	}

	// batch-queue visibility
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
					Int("q_media", len(b.mediaAssets)).
					Int("q_ins_page", len(b.pageInsights)).
					Int("q_ins_video", len(b.videoInsights)).
					Int("q_ins_reels", len(b.reelsInsights)).
					Msg("batch queues depth")
			}
		}
	}()
}

// ===================== Batch Loops =====================

func processPostsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookPost, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookPost
	t := time.NewTicker(batchTimeout)
	defer t.Stop()
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
				t.Reset(batchTimeout)
			}
		case <-t.C:
			if len(batch) > 0 {
				processPosts(ctx, batch, sink, log)
				batch = nil
			}
		}
	}
}

func processMediaAssetsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookMediaAsset, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookMediaAsset
	t := time.NewTicker(batchTimeout)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				processMediaAssets(context.Background(), batch, sink, log)
			}
			return
		case x, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					processMediaAssets(context.Background(), batch, sink, log)
				}
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatchSize {
				processMediaAssets(ctx, batch, sink, log)
				batch = nil
				t.Reset(batchTimeout)
			}
		case <-t.C:
			if len(batch) > 0 {
				processMediaAssets(ctx, batch, sink, log)
				batch = nil
			}
		}
	}
}

func processPageInsightsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookInsights
	t := time.NewTicker(batchTimeout)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				processPageInsights(context.Background(), batch, sink, log)
			}
			return
		case x, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					processPageInsights(context.Background(), batch, sink, log)
				}
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatchSize {
				processPageInsights(ctx, batch, sink, log)
				batch = nil
				t.Reset(batchTimeout)
			}
		case <-t.C:
			if len(batch) > 0 {
				processPageInsights(ctx, batch, sink, log)
				batch = nil
			}
		}
	}
}

func processVideoInsightsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookVideoInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookVideoInsights
	t := time.NewTicker(batchTimeout)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				processVideoInsights(context.Background(), batch, sink, log)
			}
			return
		case x, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					processVideoInsights(context.Background(), batch, sink, log)
				}
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatchSize {
				processVideoInsights(ctx, batch, sink, log)
				batch = nil
				t.Reset(batchTimeout)
			}
		case <-t.C:
			if len(batch) > 0 {
				processVideoInsights(ctx, batch, sink, log)
				batch = nil
			}
		}
	}
}

func processReelsInsightsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookReelsInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookReelsInsights
	t := time.NewTicker(batchTimeout)
	defer t.Stop()
	for {
		select {
		case <-ctx.Done():
			if len(batch) > 0 {
				processReelsInsights(context.Background(), batch, sink, log)
			}
			return
		case x, ok := <-in:
			if !ok {
				if len(batch) > 0 {
					processReelsInsights(context.Background(), batch, sink, log)
				}
				return
			}
			batch = append(batch, x)
			if len(batch) >= maxBatchSize {
				processReelsInsights(ctx, batch, sink, log)
				batch = nil
				t.Reset(batchTimeout)
			}
		case <-t.C:
			if len(batch) > 0 {
				processReelsInsights(ctx, batch, sink, log)
				batch = nil
			}
		}
	}
}

// ===================== ClickHouse insert helpers =====================

func processPosts(ctx context.Context, batch []*kafkamodels.ParsedFacebookPost, sink *conversions.ClickHouseSink, log *logger.Logger) {
	if len(batch) == 0 {
		return
	}
	rows := make([]*clickhousemodels.FacebookPosts, len(batch))
	for i, x := range batch {
		rows[i] = sink.ConvertFacebookPost(x)
	}
	if err := sink.BulkInsertPosts(ctx, rows); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processPosts").Str("stage", "bulk_insert_posts").Int("count", len(batch)).Msg("bulk insert posts failed")
		for _, x := range batch {
			sink.HandleFailedInsert(ctx, x, err)
		}
	}
}

func processMediaAssets(ctx context.Context, batch []*kafkamodels.ParsedFacebookMediaAsset, sink *conversions.ClickHouseSink, log *logger.Logger) {
	if len(batch) == 0 {
		return
	}
	rows := make([]*clickhousemodels.FacebookMediaAssets, len(batch))
	for i, x := range batch {
		rows[i] = sink.ConvertFacebookMediaAssets(x)
	}
	if err := sink.BulkInsertMediaAssets(ctx, rows); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processMediaAssets").Str("stage", "bulk_insert_media_assets").Int("count", len(batch)).Msg("bulk insert media assets failed")
		for _, x := range batch {
			sink.HandleFailedInsert(ctx, x, err)
		}
	}
}

func processPageInsights(ctx context.Context, batch []*kafkamodels.ParsedFacebookInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	if len(batch) == 0 {
		return
	}
	rows := make([]*clickhousemodels.FacebookInsights, len(batch))
	for i, x := range batch {
		rows[i] = sink.ConvertFacebookInsights(x)
	}
	if err := sink.BulkInsertInsights(ctx, rows); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processPageInsights").Str("stage", "bulk_insert_page_insights").Int("count", len(batch)).Msg("bulk insert page insights failed")
		for _, x := range batch {
			sink.HandleFailedInsert(ctx, x, err)
		}
	}
}

func processVideoInsights(ctx context.Context, batch []*kafkamodels.ParsedFacebookVideoInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	if len(batch) == 0 {
		return
	}
	rows := make([]*clickhousemodels.FacebookVideoInsights, len(batch))
	for i, x := range batch {
		rows[i] = sink.ConvertFacebookVideoInsights(x)
	}
	if err := sink.BulkInsertVideoInsights(ctx, rows); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processVideoInsights").Str("stage", "bulk_insert_video_insights").Int("count", len(batch)).Msg("bulk insert video insights failed")
		for _, x := range batch {
			sink.HandleFailedInsert(ctx, x, err)
		}
	}
}

func processReelsInsights(ctx context.Context, batch []*kafkamodels.ParsedFacebookReelsInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	if len(batch) == 0 {
		return
	}
	rows := make([]*clickhousemodels.FacebookReelsInsights, len(batch))
	for i, x := range batch {
		rows[i] = sink.ConvertFacebookReelsInsights(x)
	}
	if err := sink.BulkInsertReelsInsights(ctx, rows); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processReelsInsights").Str("stage", "bulk_insert_reels_insights").Int("count", len(batch)).Msg("bulk insert reels insights failed")
		for _, x := range batch {
			sink.HandleFailedInsert(ctx, x, err)
		}
	}
}
