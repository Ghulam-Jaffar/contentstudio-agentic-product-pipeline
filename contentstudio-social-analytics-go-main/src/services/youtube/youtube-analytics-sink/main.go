package main

import (
	"context"
	"encoding/json"
	"math"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	channelParserWorkers  = 3
	videoParserWorkers    = 5
	activityParserWorkers = 3
	trafficParserWorkers  = 3
	sharedParserWorkers   = 3

	batchProcessorsPerType = 3

	maxBatchSize   = 5000
	batchTimeout   = 10 * time.Second
	idleFlushDelay = 100 * time.Millisecond

	messageChanSize = 50000

	consumerGroup = "youtube-clickhouse-sink-group"

	topicRawChannels         = "raw-youtube-channels"
	topicRawVideos           = "raw-youtube-videos"
	topicRawActivityInsights = "raw-youtube-activity-insights"
	topicRawTrafficInsights  = "raw-youtube-traffic-insights"
	topicRawSharedInsights   = "raw-youtube-shared-insights"

	idleTimeout = 15 * time.Minute
)

type RawMessage struct {
	Topic string
	Key   []byte
	Value []byte
}

type BatchCollectors struct {
	channels         chan *chmodels.YouTubeChannel
	videos           chan *chmodels.YouTubeVideo
	activityInsights chan *chmodels.YouTubeActivityInsights
	trafficInsights  chan *chmodels.YouTubeTrafficInsights
	sharedInsights   chan *chmodels.YouTubeSharedInsights
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load config: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("channel_workers", channelParserWorkers).
		Int("video_workers", videoParserWorkers).
		Int("activity_workers", activityParserWorkers).
		Int("traffic_workers", trafficParserWorkers).
		Int("shared_workers", sharedParserWorkers).
		Int("batch_processors", batchProcessorsPerType).
		Str("consumer_group", consumerGroup).
		Msg("Starting YouTube Analytics Sink (merged parser+sink)")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed (continuing)")
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Create consumers for each raw topic
	channelsConsumer, err := kafka.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create channels consumer")
	}
	defer channelsConsumer.Close()

	videosConsumer, err := kafka.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create videos consumer")
	}
	defer videosConsumer.Close()

	activityConsumer, err := kafka.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create activity insights consumer")
	}
	defer activityConsumer.Close()

	trafficConsumer, err := kafka.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create traffic insights consumer")
	}
	defer trafficConsumer.Close()

	sharedConsumer, err := kafka.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create shared insights consumer")
	}
	defer sharedConsumer.Close()

	batches := &BatchCollectors{
		channels:         make(chan *chmodels.YouTubeChannel, maxBatchSize*3),
		videos:           make(chan *chmodels.YouTubeVideo, maxBatchSize*5),
		activityInsights: make(chan *chmodels.YouTubeActivityInsights, maxBatchSize*3),
		trafficInsights:  make(chan *chmodels.YouTubeTrafficInsights, maxBatchSize*3),
		sharedInsights:   make(chan *chmodels.YouTubeSharedInsights, maxBatchSize*3),
	}

	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, batches, sink, log, &batchWg, batchProcessorsPerType)

	channelsMsgChan := make(chan RawMessage, messageChanSize)
	videosMsgChan := make(chan RawMessage, messageChanSize)
	activityMsgChan := make(chan RawMessage, messageChanSize)
	trafficMsgChan := make(chan RawMessage, messageChanSize)
	sharedMsgChan := make(chan RawMessage, messageChanSize)

	var pickedChannels, pickedVideos, pickedActivity, pickedTraffic, pickedShared uint64
	var parsedChannels, parsedVideos, parsedActivity, parsedTraffic, parsedShared uint64

	var lastMessageTime int64 = time.Now().UnixNano()

	// Start parser workers
	var channelsWg, videosWg, activityWg, trafficWg, sharedWg sync.WaitGroup

	for i := 0; i < channelParserWorkers; i++ {
		channelsWg.Add(1)
		go channelParserWorker(ctx, i+1, channelsMsgChan, batches, &parsedChannels, log, &channelsWg)
	}

	for i := 0; i < videoParserWorkers; i++ {
		videosWg.Add(1)
		go videoParserWorker(ctx, i+1, videosMsgChan, batches, &parsedVideos, log, &videosWg)
	}

	for i := 0; i < activityParserWorkers; i++ {
		activityWg.Add(1)
		go activityParserWorker(ctx, i+1, activityMsgChan, batches, &parsedActivity, log, &activityWg)
	}

	for i := 0; i < trafficParserWorkers; i++ {
		trafficWg.Add(1)
		go trafficParserWorker(ctx, i+1, trafficMsgChan, batches, &parsedTraffic, log, &trafficWg)
	}

	for i := 0; i < sharedParserWorkers; i++ {
		sharedWg.Add(1)
		go sharedParserWorker(ctx, i+1, sharedMsgChan, batches, &parsedShared, log, &sharedWg)
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
					Uint64("picked_channels", atomic.LoadUint64(&pickedChannels)).
					Uint64("parsed_channels", atomic.LoadUint64(&parsedChannels)).
					Uint64("picked_videos", atomic.LoadUint64(&pickedVideos)).
					Uint64("parsed_videos", atomic.LoadUint64(&parsedVideos)).
					Uint64("picked_activity", atomic.LoadUint64(&pickedActivity)).
					Uint64("parsed_activity", atomic.LoadUint64(&parsedActivity)).
					Uint64("picked_traffic", atomic.LoadUint64(&pickedTraffic)).
					Uint64("parsed_traffic", atomic.LoadUint64(&parsedTraffic)).
					Uint64("picked_shared", atomic.LoadUint64(&pickedShared)).
					Uint64("parsed_shared", atomic.LoadUint64(&parsedShared)).
					Msg("pipeline metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	var consumersWg sync.WaitGroup
	consumersWg.Add(5)

	// Channels consumer
	go func() {
		defer consumersWg.Done()
		log.Info().Str("topic", topicRawChannels).Msg("Starting channels consumer")
		err := channelsConsumer.Consume(ctx, []string{topicRawChannels}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case channelsMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedChannels, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_channels").Msg("Channels consumer error")
		}
	}()

	// Videos consumer
	go func() {
		defer consumersWg.Done()
		log.Info().Str("topic", topicRawVideos).Msg("Starting videos consumer")
		err := videosConsumer.Consume(ctx, []string{topicRawVideos}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case videosMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedVideos, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_videos").Msg("Videos consumer error")
		}
	}()

	// Activity insights consumer
	go func() {
		defer consumersWg.Done()
		log.Info().Str("topic", topicRawActivityInsights).Msg("Starting activity insights consumer")
		err := activityConsumer.Consume(ctx, []string{topicRawActivityInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case activityMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedActivity, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_activity_insights").Msg("Activity insights consumer error")
		}
	}()

	// Traffic insights consumer
	go func() {
		defer consumersWg.Done()
		log.Info().Str("topic", topicRawTrafficInsights).Msg("Starting traffic insights consumer")
		err := trafficConsumer.Consume(ctx, []string{topicRawTrafficInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case trafficMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedTraffic, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_traffic_insights").Msg("Traffic insights consumer error")
		}
	}()

	// Shared insights consumer
	go func() {
		defer consumersWg.Done()
		log.Info().Str("topic", topicRawSharedInsights).Msg("Starting shared insights consumer")
		err := sharedConsumer.Consume(ctx, []string{topicRawSharedInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			select {
			case sharedMsgChan <- RawMessage{Topic: topic, Key: key, Value: value}:
				atomic.AddUint64(&pickedShared, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_shared_insights").Msg("Shared insights consumer error")
		}
	}()

	// Handle shutdown
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)

	log.Info().Msg("YouTube Analytics Sink started successfully")

	<-sig
	log.Info().Msg("Shutdown signal received, stopping...")

	cancel()
	consumersWg.Wait()

	close(channelsMsgChan)
	close(videosMsgChan)
	close(activityMsgChan)
	close(trafficMsgChan)
	close(sharedMsgChan)

	channelsWg.Wait()
	videosWg.Wait()
	activityWg.Wait()
	trafficWg.Wait()
	sharedWg.Wait()

	close(batches.channels)
	close(batches.videos)
	close(batches.activityInsights)
	close(batches.trafficInsights)
	close(batches.sharedInsights)

	batchWg.Wait()
	close(stopMetrics)

	log.Info().Msg("YouTube Analytics Sink stopped")
}

func channelParserWorker(ctx context.Context, workerID int, msgCh <-chan RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	workerLog := log.With().Str("pool", "channels").Int("worker_id", workerID).Logger()
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
			parseAndQueueChannel(ctx, m, b, counter, &logger.Logger{Logger: workerLog})
		}
	}
}

func videoParserWorker(ctx context.Context, workerID int, msgCh <-chan RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	workerLog := log.With().Str("pool", "videos").Int("worker_id", workerID).Logger()
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
			parseAndQueueVideo(ctx, m, b, counter, &logger.Logger{Logger: workerLog})
		}
	}
}

func activityParserWorker(ctx context.Context, workerID int, msgCh <-chan RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	workerLog := log.With().Str("pool", "activity").Int("worker_id", workerID).Logger()
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
			parseAndQueueActivityInsights(ctx, m, b, counter, &logger.Logger{Logger: workerLog})
		}
	}
}

func trafficParserWorker(ctx context.Context, workerID int, msgCh <-chan RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	workerLog := log.With().Str("pool", "traffic").Int("worker_id", workerID).Logger()
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
			parseAndQueueTrafficInsights(ctx, m, b, counter, &logger.Logger{Logger: workerLog})
		}
	}
}

func sharedParserWorker(ctx context.Context, workerID int, msgCh <-chan RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger, wg *sync.WaitGroup) {
	defer wg.Done()
	workerLog := log.With().Str("pool", "shared").Int("worker_id", workerID).Logger()
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
			parseAndQueueSharedInsights(ctx, m, b, counter, &logger.Logger{Logger: workerLog})
		}
	}
}

func parseAndQueueChannel(ctx context.Context, msg RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger) {
	var raw kafkamodels.RawYouTubeChannel
	if err := json.Unmarshal(msg.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseAndQueueChannel").Str("stage", "unmarshal_channel").Msg("Failed to parse channel")
		return
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	ch := &chmodels.YouTubeChannel{
		RecordID:        conversions.GenerateYouTubeRecordID(raw.ChannelID, today),
		ChannelID:       raw.ChannelID,
		Title:           raw.Title,
		Description:     raw.Description,
		CustomURL:       raw.CustomURL,
		ThumbnailURL:    raw.ThumbnailURL,
		ExternalBanner:  raw.BannerURL,
		Country:         raw.Country,
		SubscriberCount: raw.SubscriberCount,
		VideoCount:      raw.VideoCount,
		ViewCount:       raw.ViewCount,
		PublishedAt:     raw.PublishedAt,
		CreatedAt:       today,
		InsertedAt:      now,
	}

	select {
	case b.channels <- ch:
		atomic.AddUint64(counter, 1)
	case <-ctx.Done():
	}
}

func parseAndQueueVideo(ctx context.Context, msg RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger) {
	var raw kafkamodels.RawYouTubeVideo
	if err := json.Unmarshal(msg.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseAndQueueVideo").Str("stage", "unmarshal_video").Msg("Failed to parse video")
		return
	}

	now := time.Now().UTC()
	// Use AnalyticsDate if set, otherwise fall back to today
	createdAt := raw.AnalyticsDate
	if createdAt.IsZero() {
		createdAt = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	}

	video := &chmodels.YouTubeVideo{
		VideoID:                     raw.VideoID,
		ChannelID:                   raw.ChannelID,
		Title:                       raw.Title,
		Description:                 raw.Description,
		Duration:                    raw.Duration,
		ThumbnailURL:                raw.ThumbnailURL,
		IframeEmbedHTML:             social.GenerateEmbedHTML(raw.VideoID),
		Likes:                       raw.Likes,
		Dislikes:                    raw.Dislikes,
		Views:                       raw.Views,
		Comments:                    raw.Comments,
		Shares:                      raw.Shares,
		Favorites:                   raw.Favorites,
		Saved:                       raw.Saved,
		SubscribersGained:           raw.SubscribersGained,
		RedViews:                    raw.RedViews,
		MinutesWatched:              raw.MinutesWatched,
		RedMinutesWatched:           raw.RedMinutesWatched,
		AvgViewDuration:             raw.AvgViewDuration,
		AvgViewPercentage:           raw.AvgViewPercentage,
		Impressions:                 raw.Impressions,
		ImpressionsClickThroughRate: raw.ImpressionsClickThroughRate,
		PublishedAt:                 raw.PublishedAt,
		CreatedAt:                   createdAt,
		InsertedAt:                  now,
		MediaType:                   raw.MediaType,
	}

	select {
	case b.videos <- video:
		atomic.AddUint64(counter, 1)
	case <-ctx.Done():
	}
}

func parseAndQueueActivityInsights(ctx context.Context, msg RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger) {
	var raw struct {
		ChannelID string                           `json:"channel_id"`
		Response  *social.YouTubeAnalyticsResponse `json:"response"`
	}
	if err := json.Unmarshal(msg.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseAndQueueActivityInsights").Str("stage", "unmarshal_activity_insights").Msg("Failed to parse activity insights")
		return
	}

	if raw.Response == nil || len(raw.Response.Rows) == 0 {
		return
	}

	colIndex := make(map[string]int)
	for i, col := range raw.Response.ColumnHeaders {
		colIndex[col.Name] = i
	}

	// Group by date to deduplicate (similar to traffic insights)
	dayData := make(map[string]*chmodels.YouTubeActivityInsights)
	now := time.Now().UTC()

	for _, row := range raw.Response.Rows {
		dateStr, ok := row[colIndex["day"]].(string)
		if !ok {
			continue
		}
		date, _ := time.Parse("2006-01-02", dateStr)

		if dayData[dateStr] == nil {
			dayData[dateStr] = &chmodels.YouTubeActivityInsights{
				RecordID:   conversions.GenerateYouTubeRecordID(raw.ChannelID, date),
				ChannelID:  raw.ChannelID,
				CreatedAt:  date,
				InsertedAt: now,
			}
		}

		// Accumulate metrics for the same date
		dayData[dateStr].RedViews += getInt64FromRow(row, colIndex, "redViews")
		dayData[dateStr].Views += getInt64FromRow(row, colIndex, "views")
		dayData[dateStr].Likes += getInt64FromRow(row, colIndex, "likes")
		dayData[dateStr].Dislikes += getInt64FromRow(row, colIndex, "dislikes")
		dayData[dateStr].Comments += getInt64FromRow(row, colIndex, "comments")
		dayData[dateStr].Shares += getInt64FromRow(row, colIndex, "shares")
		dayData[dateStr].SubscribersGained += getInt64FromRow(row, colIndex, "subscribersGained")
		dayData[dateStr].EstimatedMinutesWatched += getInt64FromRow(row, colIndex, "estimatedMinutesWatched")
		dayData[dateStr].EstimatedRedMinutesWatched += getInt64FromRow(row, colIndex, "estimatedRedMinutesWatched")
		dayData[dateStr].AvgViewDuration = getInt64FromRow(row, colIndex, "averageViewDuration")
		dayData[dateStr].AvgViewPercentage = getFloat64FromRow(row, colIndex, "averageViewPercentage")
	}

	for _, insight := range dayData {
		select {
		case b.activityInsights <- insight:
			atomic.AddUint64(counter, 1)
		case <-ctx.Done():
			return
		}
	}
}

func parseAndQueueTrafficInsights(ctx context.Context, msg RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger) {
	var raw struct {
		ChannelID string                           `json:"channel_id"`
		Response  *social.YouTubeAnalyticsResponse `json:"response"`
	}
	if err := json.Unmarshal(msg.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseAndQueueTrafficInsights").Str("stage", "unmarshal_traffic_insights").Msg("Failed to parse traffic insights")
		return
	}

	if raw.Response == nil || len(raw.Response.Rows) == 0 {
		return
	}

	colIndex := make(map[string]int)
	for i, col := range raw.Response.ColumnHeaders {
		colIndex[col.Name] = i
	}

	// Group by date
	dayData := make(map[string]*chmodels.YouTubeTrafficInsights)
	subscriberWatchTimeRaw := make(map[string]float64)
	nonSubscriberWatchTimeRaw := make(map[string]float64)

	for _, row := range raw.Response.Rows {
		dateStr, ok := row[colIndex["day"]].(string)
		if !ok {
			continue
		}
		date, _ := time.Parse("2006-01-02", dateStr)

		if dayData[dateStr] == nil {
			dayData[dateStr] = &chmodels.YouTubeTrafficInsights{
				RecordID:  conversions.GenerateYouTubeRecordID(raw.ChannelID, date),
				ChannelID: raw.ChannelID,
				CreatedAt: date,
			}
		}

		trafficSource, _ := row[colIndex["insightTrafficSourceType"]].(string)
		views := getInt64FromRow(row, colIndex, "views")
		rawWatchTime := getFloat64FromRow(row, colIndex, "estimatedMinutesWatched")

		switch strings.ToUpper(trafficSource) {
		case kafkamodels.TrafficSourcePaid:
			dayData[dateStr].PaidViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceAnnotation:
			dayData[dateStr].AnnotationViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceEndScreen:
			dayData[dateStr].EndScreenViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceCampaignCard:
			dayData[dateStr].CampaignCardViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceSubscriber:
			dayData[dateStr].SubscriberViews = views
			subscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceNoLinkOther:
			dayData[dateStr].NoLinkOtherViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceYTChannel:
			dayData[dateStr].YTChannelViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceYTSearch:
			dayData[dateStr].YTSearchViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceRelatedVideo:
			dayData[dateStr].RelatedVideoViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceYTOtherPage:
			dayData[dateStr].YTOtherPageViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceExtURL:
			dayData[dateStr].ExtURLViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourcePlaylist:
			dayData[dateStr].PlaylistViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceNotification:
			dayData[dateStr].NotificationViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		case kafkamodels.TrafficSourceShorts:
			dayData[dateStr].ShortsViews = views
			nonSubscriberWatchTimeRaw[dateStr] += rawWatchTime
		}
	}

	for date, insight := range dayData {
		insight.SubscriberWatchTime = int64(math.Round(subscriberWatchTimeRaw[date]))
		insight.NonSubscriberWatchTime = int64(math.Round(nonSubscriberWatchTimeRaw[date]))
		select {
		case b.trafficInsights <- insight:
			atomic.AddUint64(counter, 1)
		case <-ctx.Done():
			return
		}
	}
}

func parseAndQueueSharedInsights(ctx context.Context, msg RawMessage, b *BatchCollectors, counter *uint64, log *logger.Logger) {
	var raw struct {
		ChannelID string                           `json:"channel_id"`
		Response  *social.YouTubeAnalyticsResponse `json:"response"`
	}
	if err := json.Unmarshal(msg.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseAndQueueSharedInsights").Str("stage", "unmarshal_shared_insights").Msg("Failed to parse shared insights")
		return
	}

	if raw.Response == nil || len(raw.Response.Rows) == 0 {
		return
	}

	now := time.Now().UTC()
	insight := &chmodels.YouTubeSharedInsights{
		RecordID:   conversions.GenerateYouTubeRecordID(raw.ChannelID, now),
		ChannelID:  raw.ChannelID,
		InsertedAt: now,
	}

	colIndex := make(map[string]int)
	for i, col := range raw.Response.ColumnHeaders {
		colIndex[col.Name] = i
	}

	for _, row := range raw.Response.Rows {
		service, _ := row[colIndex["sharingService"]].(string)
		shares := getInt64FromRow(row, colIndex, "shares")

		switch strings.ToUpper(service) {
		case kafkamodels.SharingServiceAmeba:
			insight.Ameba = shares
		case kafkamodels.SharingServiceBlogger:
			insight.Blogger = shares
		case kafkamodels.SharingServiceCopyPaste:
			insight.CopyPaste = shares
		case kafkamodels.SharingServiceCyworld:
			insight.Cyworld = shares
		case kafkamodels.SharingServiceDigg:
			insight.Digg = shares
		case kafkamodels.SharingServiceDropbox:
			insight.Dropbox = shares
		case kafkamodels.SharingServiceEmbed:
			insight.Embed = shares
		case kafkamodels.SharingServiceMail:
			insight.Mail = shares
		case kafkamodels.SharingServiceWhatsApp:
			insight.WhatsApp = shares
		case kafkamodels.SharingServiceOther:
			insight.Other = shares
		case kafkamodels.SharingServiceFacebookMsgr:
			insight.FacebookMsgr = shares
		case kafkamodels.SharingServiceFacebookPages:
			insight.FacebookPages = shares
		case kafkamodels.SharingServiceFacebook:
			insight.Facebook = shares
		case kafkamodels.SharingServiceFotka:
			insight.Fotka = shares
		case kafkamodels.SharingServiceVKontakte:
			insight.VKontakte = shares
		case kafkamodels.SharingServiceDiscord:
			insight.Discord = shares
		case kafkamodels.SharingServiceGooglePlus:
			insight.GooglePlus = shares
		case kafkamodels.SharingServiceGoo:
			insight.Goo = shares
		case kafkamodels.SharingServiceHangouts:
			insight.Hangouts = shares
		case kafkamodels.SharingServiceLinkedIn:
			insight.LinkedIn = shares
		case kafkamodels.SharingServicePinterest:
			insight.Pinterest = shares
		case kafkamodels.SharingServiceMyspace:
			insight.Myspace = shares
		case kafkamodels.SharingServiceReddit:
			insight.Reddit = shares
		case kafkamodels.SharingServiceSkype:
			insight.Skype = shares
		case kafkamodels.SharingServiceTelegram:
			insight.Telegram = shares
		case kafkamodels.SharingServiceTwitter:
			insight.Twitter = shares
		case kafkamodels.SharingServiceTumblr:
			insight.Tumblr = shares
		case kafkamodels.SharingServiceViber:
			insight.Viber = shares
		case kafkamodels.SharingServiceWeibo:
			insight.Weibo = shares
		case kafkamodels.SharingServiceWeChat:
			insight.WeChat = shares
		case kafkamodels.SharingServiceYouTube, kafkamodels.SharingServiceYouTubeGaming,
			kafkamodels.SharingServiceYouTubeKids, kafkamodels.SharingServiceYouTubeMusic,
			kafkamodels.SharingServiceYouTubeTV:
			insight.YouTube += shares
		}
	}

	select {
	case b.sharedInsights <- insight:
		atomic.AddUint64(counter, 1)
	case <-ctx.Done():
	}
}

func getInt64FromRow(row []interface{}, colIndex map[string]int, colName string) int64 {
	idx, ok := colIndex[colName]
	if !ok || idx >= len(row) {
		return 0
	}

	switch v := row[idx].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	case string:
		if i, err := strconv.ParseInt(v, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

func getFloat64FromRow(row []interface{}, colIndex map[string]int, colName string) float64 {
	idx, ok := colIndex[colName]
	if !ok || idx >= len(row) {
		return 0
	}

	switch v := row[idx].(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0
}

func startBatchProcessors(ctx context.Context, b *BatchCollectors, sink *conversions.ClickHouseSink, log *logger.Logger, wg *sync.WaitGroup, n int) {
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processChannelsBatch(ctx, id, b.channels, sink, log)
		}(i)
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processVideosBatch(ctx, id, b.videos, sink, log)
		}(i)
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processActivityBatch(ctx, id, b.activityInsights, sink, log)
		}(i)
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processTrafficBatch(ctx, id, b.trafficInsights, sink, log)
		}(i)
	}

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			processSharedBatch(ctx, id, b.sharedInsights, sink, log)
		}(i)
	}
}

func processChannelsBatch(ctx context.Context, id int, in <-chan *chmodels.YouTubeChannel, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*chmodels.YouTubeChannel
	maxTimer := time.NewTimer(batchTimeout)
	idleTimer := time.NewTimer(idleFlushDelay)
	idleTimer.Stop()

	defer maxTimer.Stop()
	defer idleTimer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		startTime := time.Now()
		if err := sink.BulkInsertYouTubeChannels(ctx, batch); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processChannelsBatch").Str("stage", "bulk_insert_channels").Int("count", len(batch)).Msg("bulk insert channels failed")
		} else {
			log.Info().Int("processor", id).Int("count", len(batch)).Dur("duration", time.Since(startTime)).Msg("inserted youtube channels")
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

func processVideosBatch(ctx context.Context, id int, in <-chan *chmodels.YouTubeVideo, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*chmodels.YouTubeVideo
	maxTimer := time.NewTimer(batchTimeout)
	idleTimer := time.NewTimer(idleFlushDelay)
	idleTimer.Stop()

	defer maxTimer.Stop()
	defer idleTimer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		startTime := time.Now()
		if err := sink.BulkInsertYouTubeVideos(ctx, batch); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processVideosBatch").Str("stage", "bulk_insert_videos").Int("count", len(batch)).Msg("bulk insert videos failed")
		} else {
			log.Info().Int("processor", id).Int("count", len(batch)).Dur("duration", time.Since(startTime)).Msg("inserted youtube videos")
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

func processActivityBatch(ctx context.Context, id int, in <-chan *chmodels.YouTubeActivityInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*chmodels.YouTubeActivityInsights
	maxTimer := time.NewTimer(batchTimeout)
	idleTimer := time.NewTimer(idleFlushDelay)
	idleTimer.Stop()

	defer maxTimer.Stop()
	defer idleTimer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		startTime := time.Now()
		if err := sink.BulkInsertYouTubeActivityInsights(ctx, batch); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processActivityBatch").Str("stage", "bulk_insert_activity_insights").Int("count", len(batch)).Msg("bulk insert activity insights failed")
		} else {
			log.Info().Int("processor", id).Int("count", len(batch)).Dur("duration", time.Since(startTime)).Msg("inserted youtube activity insights")
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

func processTrafficBatch(ctx context.Context, id int, in <-chan *chmodels.YouTubeTrafficInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*chmodels.YouTubeTrafficInsights
	maxTimer := time.NewTimer(batchTimeout)
	idleTimer := time.NewTimer(idleFlushDelay)
	idleTimer.Stop()

	defer maxTimer.Stop()
	defer idleTimer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		startTime := time.Now()
		if err := sink.BulkInsertYouTubeTrafficInsights(ctx, batch); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processTrafficBatch").Str("stage", "bulk_insert_traffic_insights").Int("count", len(batch)).Msg("bulk insert traffic insights failed")
		} else {
			log.Info().Int("processor", id).Int("count", len(batch)).Dur("duration", time.Since(startTime)).Msg("inserted youtube traffic insights")
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

func processSharedBatch(ctx context.Context, id int, in <-chan *chmodels.YouTubeSharedInsights, sink *conversions.ClickHouseSink, log *logger.Logger) {
	var batch []*chmodels.YouTubeSharedInsights
	maxTimer := time.NewTimer(batchTimeout)
	idleTimer := time.NewTimer(idleFlushDelay)
	idleTimer.Stop()

	defer maxTimer.Stop()
	defer idleTimer.Stop()

	flush := func() {
		if len(batch) == 0 {
			return
		}
		startTime := time.Now()
		if err := sink.BulkInsertYouTubeSharedInsights(ctx, batch); err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processSharedBatch").Str("stage", "bulk_insert_shared_insights").Int("count", len(batch)).Msg("bulk insert shared insights failed")
		} else {
			log.Info().Int("processor", id).Int("count", len(batch)).Dur("duration", time.Since(startTime)).Msg("inserted youtube shared insights")
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
