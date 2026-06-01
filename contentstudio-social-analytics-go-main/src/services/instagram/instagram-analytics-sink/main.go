package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

const (
	mediaParserWorkers    = 10
	insightsParserWorkers = 5

	maxBatchSize           = 10000
	batchTimeout           = 10 * time.Second
	batchProcessorsPerType = 3

	messageChanSize = 500

	mediaTopic    = "raw-instagram-media"
	insightsTopic = "raw-instagram-insights"

	consumerGroup = "instagram-posts-parser-group"
)

type ParseJob struct {
	JobType      string
	EnrichedData map[string]interface{}
	InstagramID  string
	MessageKey   string
}

type BatchCollectors struct {
	posts    chan *kafkamodels.ParsedInstagramPost
	insights chan *kafkamodels.ParsedInstagramInsight
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Instagram Analytics Sink (merged parser+sink)")

	clickhouseSink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := clickhouseSink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed - continuing anyway")
	}

	consumerMedia, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create media Kafka consumer")
	}
	defer consumerMedia.Close()

	consumerInsights, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create insights Kafka consumer")
	}
	defer consumerInsights.Close()

	ctx, cancel := context.WithCancel(context.Background())

	mediaJobs := make(chan ParseJob, messageChanSize)
	insightJobs := make(chan ParseJob, messageChanSize)

	batchCollectors := &BatchCollectors{
		posts:    make(chan *kafkamodels.ParsedInstagramPost, maxBatchSize*5),
		insights: make(chan *kafkamodels.ParsedInstagramInsight, maxBatchSize*5),
	}

	var pickedMedia, pickedInsights uint64
	var parsedPosts, parsedInsights uint64
	var insertedPosts, insertedInsights uint64

	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, batchCollectors, clickhouseSink, log, &batchWg, batchProcessorsPerType, &insertedPosts, &insertedInsights)

	var wgParsers sync.WaitGroup
	for i := 0; i < mediaParserWorkers; i++ {
		wgParsers.Add(1)
		go mediaParser(ctx, &wgParsers, i, mediaJobs, batchCollectors.posts, &parsedPosts, log)
	}
	for i := 0; i < insightsParserWorkers; i++ {
		wgParsers.Add(1)
		go insightsParser(ctx, &wgParsers, i, insightJobs, batchCollectors.insights, &parsedInsights, log)
	}

	stopMetrics := make(chan struct{})
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info().
					Str("pipeline", "media").
					Int("parse_queue", len(mediaJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedMedia)).
					Uint64("parsed", atomic.LoadUint64(&parsedPosts)).
					Uint64("inserted", atomic.LoadUint64(&insertedPosts)).
					Int("batch_queue", len(batchCollectors.posts)).
					Msg("pipeline metrics")
				log.Info().
					Str("pipeline", "insights").
					Int("parse_queue", len(insightJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedInsights)).
					Uint64("parsed", atomic.LoadUint64(&parsedInsights)).
					Uint64("inserted", atomic.LoadUint64(&insertedInsights)).
					Int("batch_queue", len(batchCollectors.insights)).
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
		log.Info().Str("topic", mediaTopic).Str("group", consumerGroup).Msg("Consuming media topic...")
		err := consumerMedia.Consume(ctx, []string{mediaTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			err := handleRawMedia(ctx, key, value, mediaJobs, log)
			if err == nil {
				atomic.AddUint64(&pickedMedia, 1)
			}
			return err
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_media").Msg("Media consumer error")
			cancel()
		}
	}()

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", insightsTopic).Str("group", consumerGroup).Msg("Consuming insights topic...")
		err := consumerInsights.Consume(ctx, []string{insightsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			err := handleRawInsights(ctx, key, value, insightJobs, log)
			if err == nil {
				atomic.AddUint64(&pickedInsights, 1)
			}
			return err
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_insights").Msg("Insights consumer error")
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Int("media_parser_workers", mediaParserWorkers).
		Int("insights_parser_workers", insightsParserWorkers).
		Int("batch_processors_per_type", batchProcessorsPerType).
		Int("max_batch_size", maxBatchSize).
		Dur("batch_timeout", batchTimeout).
		Msg("Instagram Analytics Sink started successfully")

	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping...")
	cancel()

	wgConsumers.Wait()

	close(mediaJobs)
	close(insightJobs)

	wgParsers.Wait()

	close(batchCollectors.posts)
	close(batchCollectors.insights)

	batchWg.Wait()

	close(stopMetrics)

	log.Info().
		Uint64("total_picked_media", atomic.LoadUint64(&pickedMedia)).
		Uint64("total_picked_insights", atomic.LoadUint64(&pickedInsights)).
		Uint64("total_parsed_posts", atomic.LoadUint64(&parsedPosts)).
		Uint64("total_parsed_insights", atomic.LoadUint64(&parsedInsights)).
		Uint64("total_inserted_posts", atomic.LoadUint64(&insertedPosts)).
		Uint64("total_inserted_insights", atomic.LoadUint64(&insertedInsights)).
		Msg("Instagram Analytics Sink stopped")
}

func extractKeyInfo(key string) (instagramID, workspaceID string) {
	parts := strings.Split(key, "_")
	switch len(parts) {
	case 0:
		return "", ""
	case 1:
		return parts[0], ""
	case 2:
		return parts[0], ""
	default:
		return parts[1], parts[0]
	}
}

func handleRawMedia(ctx context.Context, key, value []byte, jobChan chan<- ParseJob, log *logger.Logger) error {
	var enrichedData map[string]interface{}
	if err := json.Unmarshal(value, &enrichedData); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleRawMedia").Str("stage", "unmarshal_raw_media").Str("key", string(key)).Msg("Failed to unmarshal raw Instagram media")
		return nil
	}

	instagramID, _ := extractKeyInfo(string(key))

	job := ParseJob{
		JobType:      "media",
		EnrichedData: enrichedData,
		InstagramID:  instagramID,
		MessageKey:   string(key),
	}

	select {
	case jobChan <- job:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func handleRawInsights(ctx context.Context, key, value []byte, jobChan chan<- ParseJob, log *logger.Logger) error {
	var enrichedData map[string]interface{}
	if err := json.Unmarshal(value, &enrichedData); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleRawInsights").Str("stage", "unmarshal_raw_insights").Str("key", string(key)).Msg("Failed to unmarshal raw Instagram insights")
		return nil
	}

	instagramID, _ := extractKeyInfo(string(key))

	job := ParseJob{
		JobType:      "insights",
		EnrichedData: enrichedData,
		InstagramID:  instagramID,
		MessageKey:   string(key),
	}

	select {
	case jobChan <- job:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func mediaParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, out chan<- *kafkamodels.ParsedInstagramPost, parsedCounter *uint64, log *logger.Logger) {
	defer wg.Done()
	parser := parsing.NewInstagramParser()
	log.Info().Int("worker_id", id).Str("pool", "media").Msg("Media parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-in:
			if !ok {
				return
			}
			if job.JobType != "media" {
				continue
			}

			parsedPost, err := parser.ParseMediaWithInsights(job.EnrichedData, job.InstagramID)
			if err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "mediaParser").Str("stage", "parse_media").Int("worker_id", id).Str("instagram_id", job.InstagramID).Msg("Failed to parse media")
				continue
			}
			if parsedPost == nil {
				continue
			}

			select {
			case out <- parsedPost:
				atomic.AddUint64(parsedCounter, 1)
			case <-ctx.Done():
				return
			}
		}
	}
}

func insightsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, out chan<- *kafkamodels.ParsedInstagramInsight, parsedCounter *uint64, log *logger.Logger) {
	defer wg.Done()
	parser := parsing.NewInstagramParser()
	log.Info().Int("worker_id", id).Str("pool", "insights").Msg("Insights parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-in:
			if !ok {
				return
			}
			if job.JobType != "insights" {
				continue
			}

			if dailyData, ok := job.EnrichedData["daily_insights"].([]interface{}); ok {
				count := processDailyInsights(ctx, parser, job, dailyData, out, log, id)
				atomic.AddUint64(parsedCounter, uint64(count))
				log.Debug().Int("worker_id", id).Str("instagram_id", job.InstagramID).Int("daily_records", count).Msg("Parsed daily insights")
				continue
			}

			processLegacyInsights(ctx, parser, job, out, parsedCounter, log, id)
		}
	}
}

func processDailyInsights(ctx context.Context, parser *parsing.InstagramParser, job ParseJob, dailyData []interface{}, out chan<- *kafkamodels.ParsedInstagramInsight, log *logger.Logger, workerID int) int {
	publishedCount := 0

	for _, di := range dailyData {
		diMap, ok := di.(map[string]interface{})
		if !ok {
			continue
		}

		dateStr := extractDateString(diMap)
		if dateStr == "" {
			continue
		}

		data, ok := diMap["Data"].(map[string]interface{})
		if !ok || data == nil {
			continue
		}

		dayEnriched := map[string]interface{}{
			"insights":     data,
			"demographics": job.EnrichedData["demographics"],
			"user_info":    job.EnrichedData["user_info"],
		}

		recordID := fmt.Sprintf("%s_%s", job.InstagramID, dateStr)
		parsedInsight, err := parser.ParseInsightsWithDemographics(dayEnriched, job.InstagramID, recordID)
		if err != nil || parsedInsight == nil {
			continue
		}

		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			parsedInsight.CreatedTime = t
		}
		parsedInsight.Metadata = map[string]string{"source": "fetcher_daily"}

		select {
		case out <- parsedInsight:
			publishedCount++
		case <-ctx.Done():
			return publishedCount
		}
	}

	return publishedCount
}

func extractDateString(diMap map[string]interface{}) string {
	date, ok := diMap["Date"].(string)
	if !ok || date == "" {
		return ""
	}

	if t, err := time.Parse(time.RFC3339, date); err == nil {
		return t.Format("2006-01-02")
	}

	if len(date) >= 10 {
		return date[:10]
	}

	return ""
}

func processLegacyInsights(ctx context.Context, parser *parsing.InstagramParser, job ParseJob, out chan<- *kafkamodels.ParsedInstagramInsight, parsedCounter *uint64, log *logger.Logger, workerID int) {
	today := time.Now().UTC().Format("2006-01-02")
	recordID := fmt.Sprintf("%s_%s", job.InstagramID, today)

	parsedInsight, err := parser.ParseInsightsWithDemographics(job.EnrichedData, job.InstagramID, recordID)
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processLegacyInsights").Str("stage", "parse_insights").Int("worker_id", workerID).Str("instagram_id", job.InstagramID).Msg("Failed to parse insights")
		return
	}

	if parsedInsight == nil {
		return
	}

	select {
	case out <- parsedInsight:
		atomic.AddUint64(parsedCounter, 1)
	case <-ctx.Done():
	}
}

func startBatchProcessors(ctx context.Context, batchCollectors *BatchCollectors, sink *conversions.ClickHouseSink, log *logger.Logger, wg *sync.WaitGroup, processorsPerType int, insertedPosts, insertedInsights *uint64) {
	for i := 0; i < processorsPerType; i++ {
		wg.Add(1)
		go func(processorID int) {
			defer wg.Done()
			processPostsBatch(ctx, processorID, batchCollectors.posts, sink, log, insertedPosts)
		}(i)
	}

	for i := 0; i < processorsPerType; i++ {
		wg.Add(1)
		go func(processorID int) {
			defer wg.Done()
			processInsightsBatch(ctx, processorID, batchCollectors.insights, sink, log, insertedInsights)
		}(i)
	}

	log.Info().
		Int("max_batch_size", maxBatchSize).
		Dur("batch_timeout", batchTimeout).
		Int("processors_per_type", processorsPerType).
		Msg("Started all batch processors")
}

func processPostsBatch(ctx context.Context, processorID int, postsChan <-chan *kafkamodels.ParsedInstagramPost, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	var batch []*kafkamodels.ParsedInstagramPost
	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()

	processorLog := log.With().Int("processor_id", processorID).Str("type", "posts").Logger()
	processorLog.Info().Msg("Batch processor started")

	flush := func() {
		if len(batch) == 0 {
			return
		}
		count := processPosts(context.Background(), batch, sink, &processorLog)
		atomic.AddUint64(insertedCounter, uint64(count))
		batch = nil
	}

	for {
		select {
		case post, ok := <-postsChan:
			if !ok {
				flush()
				return
			}
			batch = append(batch, post)
			if len(batch) >= maxBatchSize {
				flush()
				ticker.Reset(batchTimeout)
			}
		case <-ticker.C:
			flush()
		}
	}
}

func processInsightsBatch(ctx context.Context, processorID int, insightsChan <-chan *kafkamodels.ParsedInstagramInsight, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	var batch []*kafkamodels.ParsedInstagramInsight
	ticker := time.NewTicker(batchTimeout)
	defer ticker.Stop()

	processorLog := log.With().Int("processor_id", processorID).Str("type", "insights").Logger()
	processorLog.Info().Msg("Batch processor started")

	flush := func() {
		if len(batch) == 0 {
			return
		}
		count := processInsights(context.Background(), batch, sink, &processorLog)
		atomic.AddUint64(insertedCounter, uint64(count))
		batch = nil
	}

	for {
		select {
		case insight, ok := <-insightsChan:
			if !ok {
				flush()
				return
			}
			batch = append(batch, insight)
			if len(batch) >= maxBatchSize {
				flush()
				ticker.Reset(batchTimeout)
			}
		case <-ticker.C:
			flush()
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
