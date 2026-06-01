package main

import (
	"context"
	"encoding/json"
	"fmt"
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
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

const (
	postParserWorkers     = 6
	insightsParserWorkers = 6

	postPublisherWorkers     = 6
	insightsPublisherWorkers = 6

	parseChanSize   = 500
	publishChanSize = 1000

	metricsEvery = 10 * time.Second

	// Input topics for pages
	topicRawPagePosts    = "raw-linkedin-page-posts"
	topicRawPageInsights = "raw-linkedin-page-insights"

	// Input topics for profiles
	topicRawProfilePosts    = "raw-linkedin-profile-posts"
	topicRawProfileInsights = "raw-linkedin-profile-insights"

	// Output topics for pages
	topicParsedPagePosts    = "parsed-linkedin-page-posts"
	topicParsedPageInsights = "parsed-linkedin-page-insights"

	// Output topics for profiles
	topicParsedProfilePosts    = "parsed-linkedin-profile-posts"
	topicParsedProfileInsights = "parsed-linkedin-profile-insights"

	// Consumer groups (separate for independent scaling)
	pageParserConsumerGroup    = "linkedin-page-parser-group"
	profileParserConsumerGroup = "linkedin-profile-parser-group"
)

type ParseJob struct {
	JobType     string
	Key         []byte
	Value       []byte
	OutputTopic string // Target topic for parsed output (page or profile specific)
}

type PublishJob struct {
	Topic string
	Key   string
	Data  []byte
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	telemetry.ConfigureSentry(cfg)
	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("post_parser_workers", postParserWorkers).
		Int("insights_parser_workers", insightsParserWorkers).
		Int("post_publisher_workers", postPublisherWorkers).
		Int("insights_publisher_workers", insightsPublisherWorkers).
		Str("page_consumer_group", pageParserConsumerGroup).
		Str("profile_consumer_group", profileParserConsumerGroup).
		Msg("Starting LinkedIn Parser service with separate page/profile pipelines")

	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	// Page consumers
	pagePostsConsumer, err := kafka2.NewConsumer(cfg.Kafka, pageParserConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create page posts consumer")
	}
	defer pagePostsConsumer.Close()

	pageInsightsConsumer, err := kafka2.NewConsumer(cfg.Kafka, pageParserConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create page insights consumer")
	}
	defer pageInsightsConsumer.Close()

	// Profile consumers
	profilePostsConsumer, err := kafka2.NewConsumer(cfg.Kafka, profileParserConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create profile posts consumer")
	}
	defer profilePostsConsumer.Close()

	profileInsightsConsumer, err := kafka2.NewConsumer(cfg.Kafka, profileParserConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create profile insights consumer")
	}
	defer profileInsightsConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	postsParseJobs := make(chan ParseJob, parseChanSize)
	insightsParseJobs := make(chan ParseJob, parseChanSize)

	postsPublishJobs := make(chan PublishJob, publishChanSize)
	insightsPublishJobs := make(chan PublishJob, publishChanSize)

	var pickedPosts, pickedInsights uint64
	var pubPosts, pubInsights uint64

	var wgParsers sync.WaitGroup
	for i := 0; i < postParserWorkers; i++ {
		wgParsers.Add(1)
		go postParser(ctx, &wgParsers, i, postsParseJobs, postsPublishJobs, log)
	}
	for i := 0; i < insightsParserWorkers; i++ {
		wgParsers.Add(1)
		go insightsParser(ctx, &wgParsers, i, insightsParseJobs, insightsPublishJobs, log)
	}

	var wgPublishers sync.WaitGroup
	for i := 0; i < postPublisherWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "posts", postsPublishJobs, producer, &pubPosts, log)
	}
	for i := 0; i < insightsPublisherWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "insights", insightsPublishJobs, producer, &pubInsights, log)
	}

	stopMetrics := make(chan struct{})
	go func() {
		t := time.NewTicker(metricsEvery)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info().
					Str("pipeline", "posts").
					Int("queue_parse", len(postsParseJobs)).
					Int("queue_publish", len(postsPublishJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedPosts)).
					Uint64("published", atomic.LoadUint64(&pubPosts)).
					Msg("metrics")

				log.Info().
					Str("pipeline", "insights").
					Int("queue_parse", len(insightsParseJobs)).
					Int("queue_publish", len(insightsPublishJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedInsights)).
					Uint64("published", atomic.LoadUint64(&pubInsights)).
					Msg("metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	var wgConsumers sync.WaitGroup
	wgConsumers.Add(4) // 2 for pages (posts + insights) + 2 for profiles (posts + insights)

	// Page posts consumer
	go func() {
		defer wgConsumers.Done()
		topics := []string{topicRawPagePosts}
		log.Info().
			Strs("topics", topics).
			Str("consumer_group", pageParserConsumerGroup).
			Msg("Starting page posts consumer")
		if err := pagePostsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case postsParseJobs <- ParseJob{JobType: "post", Key: key, Value: value, OutputTopic: topicParsedPagePosts}:
				atomic.AddUint64(&pickedPosts, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("pool", "page").Str("function", "main").Str("stage", "consume_page_posts").Msg("Page posts consumer error")
		}
	}()

	// Page insights consumer
	go func() {
		defer wgConsumers.Done()
		topics := []string{topicRawPageInsights}
		log.Info().
			Strs("topics", topics).
			Str("consumer_group", pageParserConsumerGroup).
			Msg("Starting page insights consumer")
		if err := pageInsightsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case insightsParseJobs <- ParseJob{JobType: "insights", Key: key, Value: value, OutputTopic: topicParsedPageInsights}:
				atomic.AddUint64(&pickedInsights, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("pool", "page").Str("function", "main").Str("stage", "consume_page_insights").Msg("Page insights consumer error")
		}
	}()

	// Profile posts consumer
	go func() {
		defer wgConsumers.Done()
		topics := []string{topicRawProfilePosts}
		log.Info().
			Strs("topics", topics).
			Str("consumer_group", profileParserConsumerGroup).
			Msg("Starting profile posts consumer")
		if err := profilePostsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case postsParseJobs <- ParseJob{JobType: "post", Key: key, Value: value, OutputTopic: topicParsedProfilePosts}:
				atomic.AddUint64(&pickedPosts, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("pool", "profile").Str("function", "main").Str("stage", "consume_profile_posts").Msg("Profile posts consumer error")
		}
	}()

	// Profile insights consumer
	go func() {
		defer wgConsumers.Done()
		topics := []string{topicRawProfileInsights}
		log.Info().
			Strs("topics", topics).
			Str("consumer_group", profileParserConsumerGroup).
			Msg("Starting profile insights consumer")
		if err := profileInsightsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case insightsParseJobs <- ParseJob{JobType: "insights", Key: key, Value: value, OutputTopic: topicParsedProfileInsights}:
				atomic.AddUint64(&pickedInsights, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("pool", "profile").Str("function", "main").Str("stage", "consume_profile_insights").Msg("Profile insights consumer error")
		}
	}()

	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()

	wgConsumers.Wait()

	close(postsParseJobs)
	close(insightsParseJobs)
	wgParsers.Wait()

	close(postsPublishJobs)
	close(insightsPublishJobs)
	wgPublishers.Wait()

	close(stopMetrics)

	log.Info().Msg("LinkedIn Parser service stopped")
}

func postParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, out chan<- PublishJob, log *logger.Logger) {
	defer wg.Done()
	workerLog := &logger.Logger{Logger: log.With().Str("pool", "posts").Int("worker_id", id).Logger()}
	workerLog.Info().Msg("Parser started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Parser stopped (context cancelled)")
			return
		case job, ok := <-in:
			if !ok {
				workerLog.Info().Msg("Parser stopped (channel closed)")
				return
			}
			parseAndQueuePost(ctx, job, out, workerLog)
		}
	}
}

func insightsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, out chan<- PublishJob, log *logger.Logger) {
	defer wg.Done()
	workerLog := &logger.Logger{Logger: log.With().Str("pool", "insights").Int("worker_id", id).Logger()}
	workerLog.Info().Msg("Parser started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Parser stopped (context cancelled)")
			return
		case job, ok := <-in:
			if !ok {
				workerLog.Info().Msg("Parser stopped (channel closed)")
				return
			}
			parseAndQueueInsights(ctx, job, out, workerLog)
		}
	}
}

func publisher(ctx context.Context, wg *sync.WaitGroup, id int, pool string, in <-chan PublishJob, producer kafka2.Producer, counter *uint64, log *logger.Logger) {
	defer wg.Done()
	workerLog := &logger.Logger{Logger: log.With().Str("pool", pool).Int("worker_id", id).Logger()}
	workerLog.Info().Msg("Publisher started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Publisher stopped (context cancelled)")
			return
		case job, ok := <-in:
			if !ok {
				workerLog.Info().Msg("Publisher stopped (channel closed)")
				return
			}
			if err := producer.Produce(ctx, job.Topic, []byte(job.Key), job.Data); err != nil {
				workerLog.Error().
					Err(err).
					Str("error_message", err.Error()).
					Str("topic", job.Topic).
					Str("key", job.Key).
					Str("function", "publisher").
					Str("stage", "produce_kafka").
					Msg("Failed to produce message")
				continue
			}
			atomic.AddUint64(counter, 1)
		}
	}
}

func parseAndQueuePost(ctx context.Context, job ParseJob, out chan<- PublishJob, log *logger.Logger) {
	linkedinID := string(job.Key)
	startTime := time.Now()

	parsed, err := parsing.ParsePost(json.RawMessage(job.Value))
	if err != nil {
		log.Error().
			Err(err).
			Str("error_message", err.Error()).
			Str("linkedin_id", linkedinID).
			Dur("duration", time.Since(startTime)).
			Str("function", "parseAndQueuePost").
			Str("stage", "parse_post").
			Msg("Failed to parse linkedin post")
		return
	}
	if parsed == nil {
		return
	}

	if parsed.LinkedinID == "" {
		parsed.LinkedinID = linkedinID
	}

	// Use the output topic from the job (page or profile specific)
	outputTopic := job.OutputTopic
	if outputTopic == "" {
		outputTopic = "parsed-linkedin-posts" // fallback for backward compatibility
	}

	data, _ := json.Marshal(parsed)
	select {
	case out <- PublishJob{
		Topic: outputTopic,
		Key:   parsed.PostID,
		Data:  data,
	}:
		log.Debug().
			Str("linkedin_id", linkedinID).
			Str("post_id", parsed.PostID).
			Dur("duration", time.Since(startTime)).
			Msg("Queued parsed linkedin post")
	case <-ctx.Done():
		return
	}
}

func parseAndQueueInsights(ctx context.Context, job ParseJob, out chan<- PublishJob, log *logger.Logger) {
	linkedinID := string(job.Key)
	startTime := time.Now()

	parsedList, err := parsing.ParseInsightsDaily(json.RawMessage(job.Value))
	if err != nil {
		log.Error().
			Err(err).
			Str("error_message", err.Error()).
			Str("linkedin_id", linkedinID).
			Dur("duration", time.Since(startTime)).
			Str("function", "parseAndQueueInsights").
			Str("stage", "parse_insights").
			Msg("Failed to parse linkedin insights")
		return
	}
	if len(parsedList) == 0 {
		return
	}

	// Use the output topic from the job (page or profile specific)
	outputTopic := job.OutputTopic
	if outputTopic == "" {
		outputTopic = "parsed-linkedin-insights" // fallback for backward compatibility
	}

	log.Info().
		Str("linkedin_id", linkedinID).
		Str("topic", outputTopic).
		Int("daily_buckets", len(parsedList)).
		Msg("Parsed linkedin insights into daily buckets")

	for _, parsed := range parsedList {
		if parsed.LinkedinID == "" {
			parsed.LinkedinID = linkedinID
		}

		parsed.RecordID = fmt.Sprintf("%s_%s", parsed.LinkedinID, parsed.CreatedAt.Format("2006-01-02"))

		data, _ := json.Marshal(parsed)
		select {
		case out <- PublishJob{
			Topic: outputTopic,
			Key:   parsed.RecordID,
			Data:  data,
		}:
		case <-ctx.Done():
			return
		}
	}

	log.Debug().
		Str("linkedin_id", linkedinID).
		Int("daily_buckets", len(parsedList)).
		Dur("duration", time.Since(startTime)).
		Msg("Queued parsed linkedin insights")
}
