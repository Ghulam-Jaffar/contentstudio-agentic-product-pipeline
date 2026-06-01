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
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/parsing"
)

/* ======================= Config ======================= */

const (
	// Consumers (separate, so you can scale/tune independently)
	postsConsumerGroup = "facebook-posts-parser-group"

	// Input topics
	rawPostsTopic    = "raw-facebook-posts"
	rawVideosTopic   = "raw-facebook-videos"
	rawInsightsTopic = "raw-facebook-insights"

	// Output topics
	parsedPostsTopic         = "parsed-facebook-posts"
	parsedMediaAssetsTopic   = "parsed-facebook-media-assets"
	parsedVideoInsightsTopic = "parsed-facebook-video-insights"
	parsedReelsInsightsTopic = "parsed-facebook-reels-insights"
	parsedInsightsTopic      = "parsed-facebook-insights"

	// Parser pools (independent)
	postsParserWorkers         = 5
	mediaInsightsParserWorkers = 5

	// Publisher pools (independent)
	postsPublisherWorkers         = 6
	mediaInsightsPublisherWorkers = 6

	// Channel sizes
	parseChanSize   = 100
	publishChanSize = 200

	metricsEvery = 10 * time.Second
)

/* ======================= Models ======================= */

type ParseJob struct {
	JobType     string // "post" | "video" | "insights"
	RawPost     *kafkamodels.RawFacebookPost
	RawVideo    *kafkamodels.RawFacebookVideo
	RawInsights *kafkamodels.RawFacebookInsights
	PageID      string
	PageName    string
	WorkspaceID string
	MessageKey  string
}

type ParseResult struct {
	ParsedPost    *kafkamodels.ParsedFacebookPost
	MediaAssets   []kafkamodels.ParsedFacebookMediaAsset
	VideoInsights *kafkamodels.ParsedFacebookVideoInsights
	ReelsInsights *kafkamodels.ParsedFacebookReelsInsights
	Insights      []*kafkamodels.ParsedFacebookInsights // Changed to slice for daily records
	Error         error
}

type PublishJob struct {
	Topic string
	Key   string
	Data  interface{}
}

/* ======================= main ======================= */

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Facebook Parser (posts + videos/insights)")

	// Shared producer (thread-safe)
	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	// Two independent consumers
	postsConsumer, err := kafka2.NewConsumer(cfg.Kafka, postsConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Posts consumer")
	}
	defer postsConsumer.Close()

	miConsumer, err := kafka2.NewConsumer(cfg.Kafka, postsConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Media+Insights consumer")
	}
	defer miConsumer.Close()

	// Context & signals
	ctx, cancel := context.WithCancel(context.Background())
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	// Channels (independent pipelines)
	postsParseJobs := make(chan ParseJob, parseChanSize)
	miParseJobs := make(chan ParseJob, parseChanSize)

	postsPublishJobs := make(chan PublishJob, publishChanSize)
	miPublishJobs := make(chan PublishJob, publishChanSize)

	// Metrics
	var pickedPosts, pickedVideos, pickedInsights uint64
	var pubPosts, pubMediaAssets, pubVideoInsights, pubReelsInsights, pubInsights uint64

	// Start parser pools
	var wgParsers sync.WaitGroup
	for i := 0; i < postsParserWorkers; i++ {
		wgParsers.Add(1)
		go postsParser(ctx, &wgParsers, i, postsParseJobs, postsPublishJobs, log)
	}
	for i := 0; i < mediaInsightsParserWorkers; i++ {
		wgParsers.Add(1)
		go mediaInsightsParser(ctx, &wgParsers, i, miParseJobs, miPublishJobs, log)
	}

	// Start publisher pools
	var wgPublishers sync.WaitGroup
	for i := 0; i < postsPublisherWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "posts", postsPublishJobs, producer, &pubPosts, &pubMediaAssets, &pubVideoInsights, &pubReelsInsights, &pubInsights, log)
	}
	for i := 0; i < mediaInsightsPublisherWorkers; i++ {
		wgPublishers.Add(1)
		go publisher(ctx, &wgPublishers, i, "media+insights", miPublishJobs, producer, &pubPosts, &pubMediaAssets, &pubVideoInsights, &pubReelsInsights, &pubInsights, log)
	}

	// Metrics ticker
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
					Uint64("published_posts", atomic.LoadUint64(&pubPosts)).
					Uint64("published_assets", atomic.LoadUint64(&pubMediaAssets)).
					Msg("metrics")

				log.Info().
					Str("pipeline", "videos+insights").
					Int("queue_parse", len(miParseJobs)).
					Int("queue_publish", len(miPublishJobs)).
					Uint64("picked_videos", atomic.LoadUint64(&pickedVideos)).
					Uint64("picked_insights", atomic.LoadUint64(&pickedInsights)).
					Uint64("published_video_insights", atomic.LoadUint64(&pubVideoInsights)).
					Uint64("published_reels_insights", atomic.LoadUint64(&pubReelsInsights)).
					Uint64("published_insights", atomic.LoadUint64(&pubInsights)).
					Msg("metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	// Start consumers (independent)
	var wgConsumers sync.WaitGroup
	wgConsumers.Add(2)

	// Posts consumer
	go func() {
		defer wgConsumers.Done()
		log.Info().Str("group", postsConsumerGroup).Strs("topics", []string{rawPostsTopic}).Msg("Consuming posts...")
		err := postsConsumer.Consume(ctx, []string{rawPostsTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			if !strings.HasSuffix(topic, rawPostsTopic) {
				return nil
			}
			if err := handleRawPost(ctx, key, value, postsParseJobs, log); err == nil {
				atomic.AddUint64(&pickedPosts, 1)
			}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_posts").Msg("Posts consumer error")
			cancel()
		}
	}()

	// Media + Insights consumer
	go func() {
		defer wgConsumers.Done()
		topics := []string{rawVideosTopic, rawInsightsTopic}
		log.Info().Str("group", postsConsumerGroup).Strs("topics", topics).Msg("Consuming videos+insights...")
		err := miConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			switch {
			case strings.HasSuffix(topic, rawVideosTopic):
				if err := handleRawVideo(ctx, key, value, miParseJobs, log); err == nil {
					atomic.AddUint64(&pickedVideos, 1)
				}
			case strings.HasSuffix(topic, rawInsightsTopic):
				if err := handleRawInsights(ctx, key, value, miParseJobs, log); err == nil {
					atomic.AddUint64(&pickedInsights, 1)
				}
			default:
				// ignore
			}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_media_insights").Msg("Media+Insights consumer error")
			cancel()
		}
	}()

	// Shutdown
	go func() {
		<-sig
		log.Info().Msg("Shutdown signal received, stopping...")
		cancel()
	}()

	// Wait for consumers to stop
	wgConsumers.Wait()

	// Close parse channels and wait parsers
	close(postsParseJobs)
	close(miParseJobs)
	wgParsers.Wait()

	// Close publish channels and wait publishers
	close(postsPublishJobs)
	close(miPublishJobs)
	wgPublishers.Wait()

	close(stopMetrics)

	log.Info().Msg("Facebook Parser stopped")
}

/* ======================= Consumer Handlers ======================= */

func handleRawPost(ctx context.Context, key, value []byte, out chan<- ParseJob, log *logger.Logger) error {
	var raw kafkamodels.RawFacebookPost
	if err := json.Unmarshal(value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleRawPost").Str("stage", "unmarshal").Msg("unmarshal raw post")
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
		log.Debug().Str("post_id", raw.ID).Msg("queued post")
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func handleRawVideo(ctx context.Context, key, value []byte, out chan<- ParseJob, log *logger.Logger) error {
	var raw kafkamodels.RawFacebookVideo
	if err := json.Unmarshal(value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleRawVideo").Str("stage", "unmarshal").Msg("unmarshal raw video")
		return err
	}

	pageID, _, workspaceID := extractKeyInfo(string(key))

	job := ParseJob{
		JobType:     "video",
		RawVideo:    &raw,
		PageID:      pageID,
		PageName:    "",
		WorkspaceID: workspaceID,
		MessageKey:  string(key),
	}

	select {
	case out <- job:
		log.Debug().Str("video_id", raw.ID).Msg("queued video")
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func handleRawInsights(ctx context.Context, key, value []byte, out chan<- ParseJob, log *logger.Logger) error {
	var raw kafkamodels.RawFacebookInsights
	if err := json.Unmarshal(value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "handleRawInsights").Str("stage", "unmarshal").Msg("unmarshal raw insights")
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
		log.Debug().Str("page_id", raw.PageID).Msg("queued insights")
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
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
		// workspaceID_pageID_...
		return parts[1], "", parts[0]
	}
}

/* ======================= Parser Workers ======================= */

func postsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, out chan<- PublishJob, log *logger.Logger) {
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
				log.Warn().Int("worker", id).Str("pool", "posts").Str("job", job.JobType).Msg("skip non-post")
				continue
			}
			res := parseRawFacebookPostOrVideoOrInsights(job)
			if res.Error != nil {
				log.Error().Err(res.Error).Str("error_message", res.Error.Error()).Int("worker", id).Str("pool", "posts").Str("function", "postsParser").Str("stage", "parse_post").Msg("parse error")
				continue
			}

			// publish parsed post
			if res.ParsedPost != nil {
				select {
				case out <- PublishJob{
					Topic: parsedPostsTopic,
					Key:   fmt.Sprintf("%s_%s", job.PageID, res.ParsedPost.PostID),
					Data:  res.ParsedPost,
				}:
				case <-ctx.Done():
					return
				}
			}

			// publish media assets
			for _, a := range res.MediaAssets {
				a := a
				select {
				case out <- PublishJob{
					Topic: parsedMediaAssetsTopic,
					Key:   fmt.Sprintf("%s_%s", a.PageID, a.MediaID),
					Data:  a,
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func mediaInsightsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, out chan<- PublishJob, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker", id).Str("pool", "media+insights").Msg("media+insights parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-in:
			if !ok {
				return
			}
			if job.JobType != "video" && job.JobType != "insights" {
				log.Warn().Int("worker", id).Str("pool", "media+insights").Str("job", job.JobType).Msg("skip non-video/insights")
				continue
			}

			res := parseRawFacebookPostOrVideoOrInsights(job)
			if res.Error != nil {
				log.Error().Err(res.Error).Str("error_message", res.Error.Error()).Int("worker", id).Str("pool", "media+insights").Str("function", "mediaInsightsParser").Str("stage", "parse_media_insights").Msg("parse error")
				continue
			}

			//For "video": we may get Post (normalized), MediaAssets, VideoInsights, ReelsInsights
			if res.ParsedPost != nil {
				select {
				case out <- PublishJob{
					Topic: parsedPostsTopic,
					Key:   fmt.Sprintf("%s_%s", job.PageID, res.ParsedPost.PostID),
					Data:  res.ParsedPost,
				}:
				case <-ctx.Done():
					return
				}
			}
			for _, a := range res.MediaAssets {
				a := a
				select {
				case out <- PublishJob{
					Topic: parsedMediaAssetsTopic,
					Key:   fmt.Sprintf("%s_%s", a.PageID, a.MediaID),
					Data:  a,
				}:
				case <-ctx.Done():
					return
				}
			}
			if res.VideoInsights != nil {
				select {
				case out <- PublishJob{
					Topic: parsedVideoInsightsTopic,
					Key:   fmt.Sprintf("%s_%s", job.PageID, res.VideoInsights.VideoID),
					Data:  res.VideoInsights,
				}:
				case <-ctx.Done():
					return
				}
			}
			if res.ReelsInsights != nil {
				select {
				case out <- PublishJob{
					Topic: parsedReelsInsightsTopic,
					Key:   fmt.Sprintf("%s_%s", job.PageID, res.ReelsInsights.PostID),
					Data:  res.ReelsInsights,
				}:
				case <-ctx.Done():
					return
				}
			}

			// For "insights": page insights (batch of daily records)
			if len(res.Insights) > 0 {
				select {
				case out <- PublishJob{
					Topic: parsedInsightsTopic,
					Key:   job.PageID,
					Data:  res.Insights, // Publish as batch (slice of daily records)
				}:
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

/* ======================= Publisher Workers ======================= */

func publisher(
	ctx context.Context,
	wg *sync.WaitGroup,
	id int,
	pool string,
	in <-chan PublishJob,
	producer kafka2.Producer,
	pubPosts, pubAssets, pubVid, pubReels, pubPageInsights *uint64,
	log *logger.Logger,
) {
	defer wg.Done()
	log.Info().Int("worker", id).Str("pool", pool).Msg("publisher started")

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
				log.Error().Err(err).Str("error_message", err.Error()).Int("worker", id).Str("pool", pool).Str("function", "publisher").Str("stage", "marshal").Msg("marshal failed")
				continue
			}
			if err := producer.Produce(ctx, job.Topic, []byte(job.Key), data); err != nil {
				log.Error().
					Err(err).
					Str("error_message", err.Error()).
					Int("worker", id).
					Str("pool", pool).
					Str("topic", job.Topic).
					Str("key", job.Key).
					Str("function", "publisher").
					Str("stage", "produce_kafka").
					Msg("publish failed")
				continue
			}

			// classify counters
			switch job.Topic {
			case parsedPostsTopic:
				atomic.AddUint64(pubPosts, 1)
			case parsedMediaAssetsTopic:
				atomic.AddUint64(pubAssets, 1)
			case parsedVideoInsightsTopic:
				atomic.AddUint64(pubVid, 1)
			case parsedReelsInsightsTopic:
				atomic.AddUint64(pubReels, 1)
			case parsedInsightsTopic:
				atomic.AddUint64(pubPageInsights, 1)
			}
		}
	}
}

/* ======================= Parser (shared) ======================= */

func parseRawFacebookPostOrVideoOrInsights(job ParseJob) ParseResult {
	switch job.JobType {
	case "post":
		parsedPost, mediaAssets, err := parsing.ParseRawFacebookPost(*job.RawPost)
		if err != nil {
			return ParseResult{Error: err}
		}
		return ParseResult{ParsedPost: parsedPost, MediaAssets: mediaAssets}

	case "video":
		parser := parsing.NewFacebookParser()
		vi, err := parser.ParseVideo(*job.RawVideo, job.PageID, job.PageName)
		if err != nil {
			return ParseResult{Error: fmt.Errorf("parseRawFacebookPostOrVideoOrInsights: failed to parse video insights: %w", err)}
		}

		// it is used to process the videos and reels as post related insights
		//parsePost := parser.ParseVideoPostInsights(*job.RawVideo.FaceBookVideosPostInsights)

		mediaType := "videos"
		if vi.BlueReelsPlayCount > 0 {
			mediaType = "reels"
		}

		// setting the same time as video post because we are getting unknown time
		//vi.CreatedTime = parsePost.CreatedTime
		//vi.UpdatedTime = parsePost.UpdatedTime

		// Normalize to a post-like record
		//post := &kafkamodels.ParsedFacebookPost{
		//	PageID:                       vi.PageID,
		//	PageName:                     job.PageName,
		//	PostID:                       job.RawVideo.PostID,
		//	VideoID:                      job.RawVideo.ID,
		//	MediaType:                    mediaType,
		//	CreatedTime:                  vi.CreatedTime,
		//	UpdatedTime:                  vi.UpdatedTime,
		//	SavingTime:                   vi.SavingTime,
		//	PostVideoViews:               vi.TotalVideoViews,
		//	PostImpressions:              vi.TotalVideoImpressions,
		//	PostImpressionsUnique:        vi.TotalVideoImpressionsUnique,
		//	TotalImpressions:             vi.TotalVideoImpressions,
		//	TotalEngagement:              vi.TotalEngagement,
		//	Description:                  job.RawVideo.Description,
		//	Caption:                      job.RawVideo.Message,
		//	Permalink:                    job.RawVideo.FaceBookVideosPostInsights.PermalinkURL,
		//	FullPicture:                  job.RawVideo.FaceBookVideosPostInsights.FullPicture,
		//	Like:                         parsePost.Like,
		//	Comments:                     parsePost.Comments,
		//	Wow:                          parsePost.Wow,
		//	Thankful:                     parsePost.Thankful,
		//	Love:                         parsePost.Love,
		//	Haha:                         parsePost.Haha,
		//	Sad:                          parsePost.Sad,
		//	Angry:                        parsePost.Angry,
		//	Total:                        parsePost.Total,
		//	Shares:                       parsePost.Shares,
		//	PostClicks:                   parsePost.PostClicks,
		//	PostClicksUnique:             parsePost.PostClicksUnique,
		//	PostEngaged:                  parsePost.PostEngaged,
		//	PostEngagedUsers:             parsePost.PostEngagedUsers,
		//	PostEngagementType:           parsePost.PostEngagementType,
		//	PostImpressionsOrganic:       parsePost.PostImpressionsOrganic,
		//	PostImpressionsOrganicUnique: parsePost.PostImpressionsOrganicUnique,
		//	PostImpressionsViral:         parsePost.PostImpressionsViral,
		//	PostImpressionsViralUnique:   parsePost.PostImpressionsViralUnique,
		//	PostImpressionsPaid:          parsePost.PostImpressionsPaid,
		//	PostImpressionsPaidUnique:    parsePost.PostImpressionsPaidUnique,
		//	PostMetadata:                 parsePost.PostMetadata,
		//	PostNegativeFeedback:         parsePost.PostNegativeFeedback,
		//	PostNegativeFeedbackUnique:   parsePost.PostNegativeFeedbackUnique,
		//	StatusType:                   "added_video", // for reels and videos status type is added_video, we hardcoded
		//	// that because at this moment we don't have the access of status type
		//
		//}

		//var assets []kafkamodels.ParsedFacebookMediaAsset
		//if job.RawVideo.ID != "" {
		//	assets = append(assets, kafkamodels.ParsedFacebookMediaAsset{
		//		PostID:       job.RawVideo.PostID,
		//		PageID:       job.PageID,
		//		MediaID:      job.RawVideo.ID,
		//		AssetType:    mediaType,
		//		CallToAction: post.Permalink,
		//		Link:         post.FullPicture,
		//		Caption:      job.RawVideo.Message,
		//		Description:  job.RawVideo.Description,
		//		CreatedAt:    post.CreatedTime,
		//		InsertedAt:   vi.SavingTime,
		//	})
		//}

		var reels *kafkamodels.ParsedFacebookReelsInsights
		if mediaType == "reels" {
			reels = &kafkamodels.ParsedFacebookReelsInsights{
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
		}

		if reels != nil {
			return ParseResult{
				//ParsedPost:    post,
				//MediaAssets:   assets,
				ReelsInsights: reels,
			}
		}

		return ParseResult{
			//ParsedPost:    post,
			//MediaAssets:   assets,
			VideoInsights: &vi,
		}

	case "insights":
		parser := parsing.NewFacebookParser()
		insList, err := parser.ParseInsightsDaily(*job.RawInsights, job.PageID, job.WorkspaceID)
		if err != nil {
			return ParseResult{Error: fmt.Errorf("parseRawFacebookPostOrVideoOrInsights: failed to parse insights: %w", err)}
		}
		return ParseResult{Insights: insList}
	}

	return ParseResult{Error: fmt.Errorf("parseRawFacebookPostOrVideoOrInsights: unknown job type: %s", job.JobType)}
}
