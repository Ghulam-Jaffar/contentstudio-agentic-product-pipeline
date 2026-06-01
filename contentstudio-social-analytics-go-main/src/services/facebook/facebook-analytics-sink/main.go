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
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

const (
	postsParserWorkers    = 5
	insightsParserWorkers = 5

	maxBatchSize           = 10000
	batchTimeout           = 10 * time.Second
	batchProcessorsPerType = 3

	parseChanSize   = 100
	messageChanSize = 50000

	rawPostsTopic    = "raw-facebook-posts"
	rawVideosTopic   = "raw-facebook-videos"
	rawInsightsTopic = "raw-facebook-insights"

	consumerGroup = "facebook-posts-parser-group"

	// idleTimeout is the duration after which the service will shutdown
	// if no new messages are received. This allows the service to exit
	// gracefully after batch processing is complete.
	idleTimeout       = 5 * time.Minute
	idleCheckInterval = 30 * time.Second
)

type ParseJob struct {
	JobType     string
	RawPost     *kafkamodels.RawFacebookPost
	RawVideo    *kafkamodels.RawFacebookVideo
	RawInsights *kafkamodels.RawFacebookInsights
	PageID      string
	PageName    string
	WorkspaceID string
	MessageKey  string
}

type BatchCollectors struct {
	posts         chan *kafkamodels.ParsedFacebookPost
	mediaAssets   chan *kafkamodels.ParsedFacebookMediaAsset
	pageInsights  chan *kafkamodels.ParsedFacebookInsights
	videoInsights chan *kafkamodels.ParsedFacebookVideoInsights
	reelsInsights chan *kafkamodels.ParsedFacebookReelsInsights
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Facebook Analytics Sink (merged parser+sink)")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed - continuing anyway")
	}

	postsConsumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create posts consumer")
	}
	defer postsConsumer.Close()

	miConsumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create media+insights consumer")
	}
	defer miConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())

	postsParseJobs := make(chan ParseJob, parseChanSize)
	miParseJobs := make(chan ParseJob, parseChanSize)

	batches := &BatchCollectors{
		posts:         make(chan *kafkamodels.ParsedFacebookPost, maxBatchSize*5),
		mediaAssets:   make(chan *kafkamodels.ParsedFacebookMediaAsset, maxBatchSize*5),
		pageInsights:  make(chan *kafkamodels.ParsedFacebookInsights, maxBatchSize*5),
		videoInsights: make(chan *kafkamodels.ParsedFacebookVideoInsights, maxBatchSize*5),
		reelsInsights: make(chan *kafkamodels.ParsedFacebookReelsInsights, maxBatchSize*5),
	}

	var pickedPosts, pickedVideos, pickedInsights uint64
	var parsedPosts, parsedAssets, parsedVideoIns, parsedReelsIns, parsedPageIns uint64

	// Track last message time for idle timeout detection
	var lastMessageTime int64 = time.Now().UnixNano()

	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, batches, sink, log, &batchWg, batchProcessorsPerType)

	var wgParsers sync.WaitGroup
	for i := 0; i < postsParserWorkers; i++ {
		wgParsers.Add(1)
		go postsParser(ctx, &wgParsers, i, postsParseJobs, batches, &parsedPosts, &parsedAssets, log)
	}
	for i := 0; i < insightsParserWorkers; i++ {
		wgParsers.Add(1)
		go mediaInsightsParser(ctx, &wgParsers, i, miParseJobs, batches, &parsedVideoIns, &parsedReelsIns, &parsedPageIns, log)
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
					Int("parse_queue", len(postsParseJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedPosts)).
					Uint64("parsed_posts", atomic.LoadUint64(&parsedPosts)).
					Uint64("parsed_assets", atomic.LoadUint64(&parsedAssets)).
					Int("batch_posts", len(batches.posts)).
					Int("batch_assets", len(batches.mediaAssets)).
					Msg("metrics")
				log.Info().
					Str("pipeline", "videos+insights").
					Int("parse_queue", len(miParseJobs)).
					Uint64("picked_videos", atomic.LoadUint64(&pickedVideos)).
					Uint64("picked_insights", atomic.LoadUint64(&pickedInsights)).
					Uint64("parsed_video_ins", atomic.LoadUint64(&parsedVideoIns)).
					Uint64("parsed_reels_ins", atomic.LoadUint64(&parsedReelsIns)).
					Uint64("parsed_page_ins", atomic.LoadUint64(&parsedPageIns)).
					Int("batch_page", len(batches.pageInsights)).
					Int("batch_video", len(batches.videoInsights)).
					Int("batch_reels", len(batches.reelsInsights)).
					Msg("metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	var wgConsumers sync.WaitGroup
	wgConsumers.Add(2)

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", rawPostsTopic).Str("group", consumerGroup).Msg("Consuming posts...")
		err := postsConsumer.Consume(ctx, []string{rawPostsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			if err := handleRawPost(ctx, key, value, postsParseJobs, log); err == nil {
				atomic.AddUint64(&pickedPosts, 1)
			}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_posts").Msg("Posts consumer error")
			cancel()
		}
	}()

	go func() {
		defer wgConsumers.Done()
		topics := []string{rawVideosTopic, rawInsightsTopic}
		log.Info().Strs("topics", topics).Str("group", consumerGroup).Msg("Consuming videos+insights...")
		err := miConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			switch {
			case strings.HasSuffix(topic, rawVideosTopic):
				if err := handleRawVideo(ctx, key, value, miParseJobs, log); err == nil {
					atomic.AddUint64(&pickedVideos, 1)
				}
			case strings.HasSuffix(topic, rawInsightsTopic):
				if err := handleRawInsights(ctx, key, value, miParseJobs, log); err == nil {
					atomic.AddUint64(&pickedInsights, 1)
				}
			}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_media_insights").Msg("Media+Insights consumer error")
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Int("posts_parser_workers", postsParserWorkers).
		Int("insights_parser_workers", insightsParserWorkers).
		Int("batch_processors_per_type", batchProcessorsPerType).
		Int("max_batch_size", maxBatchSize).
		Dur("batch_timeout", batchTimeout).
		Msg("Facebook Analytics Sink started successfully")

	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping...")
	cancel()

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

	close(stopMetrics)

	log.Info().
		Uint64("total_picked_posts", atomic.LoadUint64(&pickedPosts)).
		Uint64("total_picked_videos", atomic.LoadUint64(&pickedVideos)).
		Uint64("total_picked_insights", atomic.LoadUint64(&pickedInsights)).
		Uint64("total_parsed_posts", atomic.LoadUint64(&parsedPosts)).
		Uint64("total_parsed_assets", atomic.LoadUint64(&parsedAssets)).
		Uint64("total_parsed_video_ins", atomic.LoadUint64(&parsedVideoIns)).
		Uint64("total_parsed_reels_ins", atomic.LoadUint64(&parsedReelsIns)).
		Uint64("total_parsed_page_ins", atomic.LoadUint64(&parsedPageIns)).
		Msg("Facebook Analytics Sink stopped")
}

func extractKeyInfo(key string) (pageID, pageName, workspaceID string) {
	parts := strings.Split(key, "_")
	switch len(parts) {
	case 0:
		return "", "", ""
	case 1:
		return parts[0], "", ""
	case 2:
		return parts[0], "", ""
	default:
		return parts[1], "", parts[0]
	}
}

func handleRawPost(ctx context.Context, key, value []byte, out chan<- ParseJob, log *logger.Logger) error {
	var raw kafkamodels.RawFacebookPost
	if err := json.Unmarshal(value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleRawPost").Str("stage", "unmarshal_raw_post").Msg("unmarshal raw post")
		return err
	}

	pageID, pageName, workspaceID := extractKeyInfo(string(key))
	if raw.From != nil {
		if pageName == "" {
			pageName = raw.From.Name
		}
		if pageID == "" {
			pageID = raw.From.ID
		}
	}

	job := ParseJob{
		JobType:     "post",
		RawPost:     &raw,
		PageID:      pageID,
		PageName:    pageName,
		WorkspaceID: workspaceID,
		MessageKey:  string(key),
	}

	select {
	case out <- job:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func handleRawVideo(ctx context.Context, key, value []byte, out chan<- ParseJob, log *logger.Logger) error {
	var raw kafkamodels.RawFacebookVideo
	if err := json.Unmarshal(value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleRawVideo").Str("stage", "unmarshal_raw_video").Msg("unmarshal raw video")
		return err
	}

	pageID, _, workspaceID := extractKeyInfo(string(key))

	job := ParseJob{
		JobType:     "video",
		RawVideo:    &raw,
		PageID:      pageID,
		WorkspaceID: workspaceID,
		MessageKey:  string(key),
	}

	select {
	case out <- job:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func handleRawInsights(ctx context.Context, key, value []byte, out chan<- ParseJob, log *logger.Logger) error {
	var raw kafkamodels.RawFacebookInsights
	if err := json.Unmarshal(value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleRawInsights").Str("stage", "unmarshal_raw_insights").Msg("unmarshal raw insights")
		return err
	}

	pageID, pageName, workspaceID := extractKeyInfo(string(key))

	job := ParseJob{
		JobType:     "insights",
		RawInsights: &raw,
		PageID:      pageID,
		PageName:    pageName,
		WorkspaceID: workspaceID,
		MessageKey:  string(key),
	}

	select {
	case out <- job:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func postsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, batches *BatchCollectors, parsedPosts, parsedAssets *uint64, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker", id).Str("pool", "posts").Msg("posts parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-in:
			if !ok {
				return
			}
			if job.JobType != "post" {
				continue
			}

			parsedPost, mediaAssets, err := parsing.ParseRawFacebookPost(*job.RawPost)
			if err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "postsParser").Str("stage", "parse_post").Int("worker", id).Msg("parse error")
				continue
			}

			if parsedPost != nil {
				select {
				case batches.posts <- parsedPost:
					atomic.AddUint64(parsedPosts, 1)
				case <-ctx.Done():
					return
				}
			}

			for _, a := range mediaAssets {
				a := a
				select {
				case batches.mediaAssets <- &a:
					atomic.AddUint64(parsedAssets, 1)
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func mediaInsightsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, batches *BatchCollectors, parsedVideoIns, parsedReelsIns, parsedPageIns *uint64, log *logger.Logger) {
	defer wg.Done()
	parser := parsing.NewFacebookParser()
	log.Info().Int("worker", id).Str("pool", "media+insights").Msg("media+insights parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-in:
			if !ok {
				return
			}

			switch job.JobType {
			case "video":
				vi, err := parser.ParseVideo(*job.RawVideo, job.PageID, job.PageName)
				if err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("function", "mediaInsightsParser").Str("stage", "parse_video").Int("worker", id).Msg("parse video error")
					continue
				}

				mediaType := "videos"
				if vi.BlueReelsPlayCount > 0 {
					mediaType = "reels"
				}

				if mediaType == "reels" {
					reels := &kafkamodels.ParsedFacebookReelsInsights{
						PageID:               vi.PageID,
						PostID:               vi.PostID,
						AverageTimeWatched:   int64(vi.PostVideoAvgTimeWatched),
						TotalTimeWatchedInMs: vi.PostVideoViewTime,
						PlayCount:            vi.BlueReelsPlayCount,
						ImpressionsUnique:    vi.PostImpressionsUnique,
						ReelFollowers:        0,
						CreatedAt:            vi.CreatedTime,
						SavingTime:           vi.SavingTime,
					}
					select {
					case batches.reelsInsights <- reels:
						atomic.AddUint64(parsedReelsIns, 1)
					case <-ctx.Done():
						return
					}
				} else {
					select {
					case batches.videoInsights <- &vi:
						atomic.AddUint64(parsedVideoIns, 1)
					case <-ctx.Done():
						return
					}
				}

			case "insights":
				insList, err := parser.ParseInsightsDaily(*job.RawInsights, job.PageID, job.WorkspaceID)
				if err != nil {
					log.Error().Err(err).Str("error_message", err.Error()).Str("function", "mediaInsightsParser").Str("stage", "parse_insights").Int("worker", id).Msg("parse insights error")
					continue
				}

				for _, ins := range insList {
					select {
					case batches.pageInsights <- ins:
						atomic.AddUint64(parsedPageIns, 1)
					case <-ctx.Done():
						return
					}
				}
			}
		}
	}
}

func startBatchProcessors(ctx context.Context, b *BatchCollectors, sink *conversions.ClickHouseSink, log *logger.Logger, wg *sync.WaitGroup, n int) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processPostsBatch(ctx, b.posts, sink, log) }()
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processMediaAssetsBatch(ctx, b.mediaAssets, sink, log) }()
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processPageInsightsBatch(ctx, b.pageInsights, sink, log) }()
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processVideoInsightsBatch(ctx, b.videoInsights, sink, log) }()
	}
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); processReelsInsightsBatch(ctx, b.reelsInsights, sink, log) }()
	}

	log.Info().
		Int("max_batch_size", maxBatchSize).
		Dur("batch_timeout", batchTimeout).
		Int("processors_per_type", n).
		Msg("Started all batch processors")
}

func processPostsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookPost, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookPost
	t := time.NewTicker(batchTimeout)
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
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processPostsBatch").Str("stage", "bulk_insert_posts").Int("count", len(batch)).Msg("bulk insert posts failed")
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
			if len(batch) >= maxBatchSize {
				flush()
				t.Reset(batchTimeout)
			}
		case <-t.C:
			flush()
		}
	}
}

func processMediaAssetsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookMediaAsset, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookMediaAsset
	t := time.NewTicker(batchTimeout)
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
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processMediaAssetsBatch").Str("stage", "bulk_insert_media_assets").Int("count", len(batch)).Msg("bulk insert media assets failed")
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
			if len(batch) >= maxBatchSize {
				flush()
				t.Reset(batchTimeout)
			}
		case <-t.C:
			flush()
		}
	}
}

func processPageInsightsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookInsights
	t := time.NewTicker(batchTimeout)
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
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processPageInsightsBatch").Str("stage", "bulk_insert_page_insights").Int("count", len(batch)).Msg("bulk insert page insights failed")
		} else {
			log.Info().Int("batch_size", len(batch)).Msg("Inserted page insights batch")
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
				t.Reset(batchTimeout)
			}
		case <-t.C:
			flush()
		}
	}
}

func processVideoInsightsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookVideoInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookVideoInsights
	t := time.NewTicker(batchTimeout)
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
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processVideoInsightsBatch").Str("stage", "bulk_insert_video_insights").Int("count", len(batch)).Msg("bulk insert video insights failed")
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
			if len(batch) >= maxBatchSize {
				flush()
				t.Reset(batchTimeout)
			}
		case <-t.C:
			flush()
		}
	}
}

func processReelsInsightsBatch(ctx context.Context, in <-chan *kafkamodels.ParsedFacebookReelsInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*kafkamodels.ParsedFacebookReelsInsights
	t := time.NewTicker(batchTimeout)
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
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processReelsInsightsBatch").Str("stage", "bulk_insert_reels_insights").Int("count", len(batch)).Msg("bulk insert reels insights failed")
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
			if len(batch) >= maxBatchSize {
				flush()
				t.Reset(batchTimeout)
			}
		case <-t.C:
			flush()
		}
	}
}
