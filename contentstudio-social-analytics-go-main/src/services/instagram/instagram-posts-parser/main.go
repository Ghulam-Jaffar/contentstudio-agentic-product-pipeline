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

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

const (
	// Separate parser worker pools
	mediaParserWorkers    = 10
	insightsParserWorkers = 5

	// Separate publisher worker pools
	mediaPublishWorkers    = 10
	insightsPublishWorkers = 6

	// Buffers - increased to match fetcher throughput
	messageChanSize = 500 // per channel (was 100)

	// Topics
	mediaTopic    = "raw-instagram-media"
	insightsTopic = "raw-instagram-insights"

	parsedPostsTopic    = "parsed-instagram-posts"
	parsedInsightsTopic = "parsed-instagram-insights"
)

type ParseJob struct {
	JobType      string                 // "media" or "insights"
	EnrichedData map[string]interface{} // Enriched data with insights/demographics
	InstagramID  string
	MessageKey   string
}

type PublishJob struct {
	Topic string
	Key   string
	Data  interface{}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Instagram Parser service")

	// Independent consumers per topic
	consumerMedia, err := kafka2.NewConsumer(cfg.Kafka, "instagram-posts-parser-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create media Kafka consumer")
	}
	defer consumerMedia.Close()

	consumerInsights, err := kafka2.NewConsumer(cfg.Kafka, "instagram-posts-parser-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create insights Kafka consumer")
	}
	defer consumerInsights.Close()

	// Independent producers per pipeline (to avoid backpressure coupling)
	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Dedicated channels for parsing and publishing
	mediaJobs := make(chan ParseJob, messageChanSize)
	insightJobs := make(chan ParseJob, messageChanSize)
	mediaPubJobs := make(chan PublishJob, messageChanSize)
	insightPubJobs := make(chan PublishJob, messageChanSize)

	// Metrics
	var pickedMedia, pickedInsights uint64
	var pubMedia, pubInsights uint64

	// Start parser pools
	var wgParsers sync.WaitGroup
	for i := 0; i < mediaParserWorkers; i++ {
		wgParsers.Add(1)
		go mediaParser(ctx, &wgParsers, i, mediaJobs, mediaPubJobs, &pubMedia, log)
	}
	for i := 0; i < insightsParserWorkers; i++ {
		wgParsers.Add(1)
		go insightsParser(ctx, &wgParsers, i, insightJobs, insightPubJobs, &pubInsights, log)
	}

	// Start publisher pools (independent)
	var wgPublishers sync.WaitGroup
	for i := 0; i < mediaPublishWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "media", mediaPubJobs, producer, &pubMedia, log)
	}
	for i := 0; i < insightsPublishWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "insights", insightPubJobs, producer, &pubInsights, log)
	}

	// Periodic metrics
	stopMetrics := make(chan struct{})
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info().
					Str("pipeline", "media").
					Int("queue_len", len(mediaJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedMedia)).
					Uint64("published", atomic.LoadUint64(&pubMedia)).
					Msg("pipeline metrics")
				log.Info().
					Str("pipeline", "insights").
					Int("queue_len", len(insightJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedInsights)).
					Uint64("published", atomic.LoadUint64(&pubInsights)).
					Msg("pipeline metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	// Start consumers independently (each can block without affecting the other)
	var wgConsumers sync.WaitGroup
	wgConsumers.Add(2)

	go func() {
		defer wgConsumers.Done()
		log.Info().Strs("topics", []string{mediaTopic}).Str("group", "instagram-posts-parser-group").Msg("Consuming media topic...")
		err := consumerMedia.Consume(ctx, []string{mediaTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			// Note: only mediaTopic reaches here
			err := handleRawMedia(ctx, key, value, mediaJobs, log)
			if err == nil {
				atomic.AddUint64(&pickedMedia, 1)
			}
			return err
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_media").Msg("Media consumer error")
			cancel()
		}
	}()

	go func() {
		defer wgConsumers.Done()
		log.Info().Strs("topics", []string{insightsTopic}).Str("group", "instagram-posts-parser-group").Msg("Consuming insights topic...")
		err := consumerInsights.Consume(ctx, []string{insightsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			// Note: only insightsTopic reaches here
			err := handleRawInsights(ctx, key, value, insightJobs, log)
			if err == nil {
				atomic.AddUint64(&pickedInsights, 1)
			}
			return err
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_insights").Msg("Insights consumer error")
			cancel()
		}
	}()

	// Handle shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan
	log.Info().Msg("Shutdown signal received, canceling context...")
	cancel()

	// Wait for consumers to stop
	wgConsumers.Wait()

	// Close parse job channels (no more inputs)
	close(mediaJobs)
	close(insightJobs)

	// Wait parsers to drain -> they will flush publish jobs
	wgParsers.Wait()

	// Close publish job channels (no more outputs)
	close(mediaPubJobs)
	close(insightPubJobs)

	// Wait publishers to finish outstanding sends
	wgPublishers.Wait()

	close(stopMetrics)

	log.Info().Msg("Instagram Parser service stopped")
}

// ------------------ Consumer Handlers ------------------

// extractKeyInfo extracts Instagram ID and workspace ID from the message key.
// Key formats:
// - "instagramID_mediaID" (2 parts) -> instagramID is parts[0]
// - "workspaceID_instagramID_mediaID" (3+ parts) -> instagramID is parts[1], workspaceID is parts[0]
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
		// workspaceID_instagramID_mediaID
		return parts[1], parts[0]
	}
}

func handleRawMedia(ctx context.Context, key, value []byte, jobChan chan<- ParseJob, log *logger.Logger) error {
	var enrichedData map[string]interface{}
	if err := json.Unmarshal(value, &enrichedData); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("key", string(key)).Str("function", "handleRawMedia").Str("stage", "unmarshal").Msg("Failed to unmarshal raw Instagram media")
		return nil // skip invalid
	}

	instagramID, workspaceID := extractKeyInfo(string(key))

	mediaID := ""
	if id, ok := enrichedData["id"].(string); ok {
		mediaID = id
	}

	job := ParseJob{
		JobType:      "media",
		EnrichedData: enrichedData,
		InstagramID:  instagramID,
		MessageKey:   string(key),
	}

	select {
	case jobChan <- job:
		log.Debug().
			Str("media_id", mediaID).
			Str("instagram_id", instagramID).
			Str("workspace_id", workspaceID).
			Msg("Queued raw media for processing")
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func handleRawInsights(ctx context.Context, key, value []byte, jobChan chan<- ParseJob, log *logger.Logger) error {
	var enrichedData map[string]interface{}
	if err := json.Unmarshal(value, &enrichedData); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("key", string(key)).Str("function", "handleRawInsights").Str("stage", "unmarshal").Msg("Failed to unmarshal raw Instagram insights")
		return nil // skip invalid
	}

	instagramID, workspaceID := extractKeyInfo(string(key))

	job := ParseJob{
		JobType:      "insights",
		EnrichedData: enrichedData,
		InstagramID:  instagramID,
		MessageKey:   string(key),
	}

	select {
	case jobChan <- job:
		log.Debug().
			Str("instagram_id", instagramID).
			Str("workspace_id", workspaceID).
			Msg("Queued raw insights for processing")
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

// ------------------ Parser Workers ------------------

func mediaParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, out chan<- PublishJob, pubCounter *uint64, log *logger.Logger) {
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
				log.Warn().Int("worker_id", id).Str("pool", "media").Str("job_type", job.JobType).Msg("Skipping non-media job")
				continue
			}

			parsedPost, err := parser.ParseMediaWithInsights(job.EnrichedData, job.InstagramID)
			if err != nil {
				log.Error().Err(err).
					Str("error_message", err.Error()).
					Int("worker_id", id).
					Str("instagram_id", job.InstagramID).
					Str("function", "mediaParser").
					Str("stage", "parse_media").
					Msg("Failed to parse media")
				continue
			}
			if parsedPost == nil {
				log.Debug().Int("worker_id", id).Str("pool", "media").Msg("Parsed post was nil, skipping")
				continue
			}

			select {
			case out <- PublishJob{
				Topic: parsedPostsTopic,
				Key:   fmt.Sprintf("%s_%s", job.InstagramID, parsedPost.MediaID),
				Data:  parsedPost,
			}:
				log.Debug().
					Int("worker_id", id).
					Str("pool", "media").
					Str("media_id", parsedPost.MediaID).
					Msg("Queued post for publish")
			case <-ctx.Done():
				return
			}
		}
	}
}

// insightsParser processes insights jobs and publishes parsed insights to Kafka.
// Supports both daily insights format (multiple records) and legacy single-record format.
func insightsParser(
	ctx context.Context,
	wg *sync.WaitGroup,
	id int,
	in <-chan ParseJob,
	out chan<- PublishJob,
	pubCounter *uint64,
	log *logger.Logger,
) {
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

			// Check for new daily_insights format (from fetcher with FetchInsightsDaily)
			if dailyData, ok := job.EnrichedData["daily_insights"].([]interface{}); ok {
				count := processDailyInsights(ctx, parser, job, dailyData, out, log, id)
				log.Info().
					Int("worker_id", id).
					Str("instagram_id", job.InstagramID).
					Int("daily_records", count).
					Msg("Parsed daily insights")
				continue
			}

			// Legacy format: single aggregated insights
			processLegacyInsights(ctx, parser, job, out, log, id)
		}
	}
}

// processDailyInsights handles the new daily insights format.
// Parses each day's insights and publishes them as separate records.
// Returns the number of successfully published records.
func processDailyInsights(
	ctx context.Context,
	parser *parsing.InstagramParser,
	job ParseJob,
	dailyData []interface{},
	out chan<- PublishJob,
	log *logger.Logger,
	workerID int,
) int {
	publishedCount := 0

	for _, di := range dailyData {
		diMap, ok := di.(map[string]interface{})
		if !ok {
			continue
		}

		// Extract and validate date
		dateStr := extractDateString(diMap)
		if dateStr == "" {
			continue
		}

		// Extract insights data
		data, ok := diMap["Data"].(map[string]interface{})
		if !ok || data == nil {
			continue
		}

		// Build enriched data combining day's insights with shared demographics
		dayEnriched := map[string]interface{}{
			"insights":     data,
			"demographics": job.EnrichedData["demographics"],
			"user_info":    job.EnrichedData["user_info"],
		}

		// Parse and publish
		recordID := fmt.Sprintf("%s_%s", job.InstagramID, dateStr)
		parsedInsight, err := parser.ParseInsightsWithDemographics(
			dayEnriched, job.InstagramID, recordID,
		)
		if err != nil || parsedInsight == nil {
			continue
		}

		// Set metadata for daily records
		if t, err := time.Parse("2006-01-02", dateStr); err == nil {
			parsedInsight.CreatedTime = t
		}
		parsedInsight.Metadata = map[string]string{"source": "fetcher_daily"}

		// Publish to output channel
		select {
		case out <- PublishJob{
			Topic: parsedInsightsTopic,
			Key:   parsedInsight.RecordID,
			Data:  parsedInsight,
		}:
			publishedCount++
		case <-ctx.Done():
			return publishedCount
		}
	}

	return publishedCount
}

// extractDateString extracts and formats the date from a daily insight map.
// Returns empty string if date cannot be extracted.
func extractDateString(diMap map[string]interface{}) string {
	date, ok := diMap["Date"].(string)
	if !ok || date == "" {
		return ""
	}

	// Try RFC3339 format first
	if t, err := time.Parse(time.RFC3339, date); err == nil {
		return t.Format("2006-01-02")
	}

	// Fallback: take first 10 characters (YYYY-MM-DD)
	if len(date) >= 10 {
		return date[:10]
	}

	return ""
}

// processLegacyInsights handles the legacy single-record insights format.
func processLegacyInsights(
	ctx context.Context,
	parser *parsing.InstagramParser,
	job ParseJob,
	out chan<- PublishJob,
	log *logger.Logger,
	workerID int,
) {
	today := time.Now().UTC().Format("2006-01-02")
	recordID := fmt.Sprintf("%s_%s", job.InstagramID, today)

	parsedInsight, err := parser.ParseInsightsWithDemographics(
		job.EnrichedData, job.InstagramID, recordID,
	)
	if err != nil {
		log.Error().Err(err).
			Str("error_message", err.Error()).
			Int("worker_id", workerID).
			Str("instagram_id", job.InstagramID).
			Str("function", "processLegacyInsights").
			Str("stage", "parse_insights").
			Msg("Failed to parse insights")
		return
	}

	if parsedInsight == nil {
		return
	}

	log.Info().
		Int("worker_id", workerID).
		Str("instagram_id", job.InstagramID).
		Str("record_id", recordID).
		Msg("Parsed insights (legacy format)")

	select {
	case out <- PublishJob{
		Topic: parsedInsightsTopic,
		Key:   parsedInsight.RecordID,
		Data:  parsedInsight,
	}:
	case <-ctx.Done():
	}
}

// ------------------ Publisher Workers ------------------

func publisher(ctx context.Context, wg *sync.WaitGroup, id int, pool string, in <-chan PublishJob, producer kafka2.Producer, pubCounter *uint64, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker_id", id).Str("pool", pool).Msg("Publisher started")

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-in:
			if !ok {
				return
			}

			data, err := json.Marshal(job.Data)
			if err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Int("worker_id", id).Str("pool", pool).Str("key", job.Key).Str("function", "publisher").Str("stage", "marshal").Msg("Marshal failed")
				continue
			}

			if err := producer.Produce(ctx, job.Topic, []byte(job.Key), data); err != nil {
				log.Error().
					Err(err).
					Str("error_message", err.Error()).
					Int("worker_id", id).
					Str("pool", pool).
					Str("topic", job.Topic).
					Str("key", job.Key).
					Str("function", "publisher").
					Str("stage", "produce_kafka").
					Msg("Publish failed")
				continue
			}

			atomic.AddUint64(pubCounter, 1)

			log.Debug().
				Int("worker_id", id).
				Str("pool", pool).
				Str("topic", job.Topic).
				Str("key", job.Key).
				Msg("Published")
		}
	}
}
