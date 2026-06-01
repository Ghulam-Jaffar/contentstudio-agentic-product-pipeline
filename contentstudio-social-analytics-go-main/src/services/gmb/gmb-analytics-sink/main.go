package main

import (
	"context"
	"encoding/json"
	"os"
	"os/signal"
	"strconv"
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

	rawDataTopic  = "raw-gmb-data"
	consumerGroup = "gmb-analytics-sink-group"

	idleTimeout       = 5 * time.Minute
	idleCheckInterval = 30 * time.Second
)

type RawMessage struct {
	Topic string
	Key   []byte
	Value []byte
}

type BatchCollectors struct {
	dailyMetrics   chan *chmodels.GMBDailyMetrics
	mediaAssets    chan *chmodels.GMBMediaAssets
	searchKeywords chan *chmodels.GMBSearchKeywordsMonthly
	localPosts     chan *chmodels.GMBLocalPosts
	reviews        chan *chmodels.GMBReviews
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting GMB Analytics Sink (merged parser+sink)")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed - continuing anyway")
	}

	consumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create consumer")
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())

	jobs := make(chan RawMessage, messageChanSize)

	batches := &BatchCollectors{
		dailyMetrics:   make(chan *chmodels.GMBDailyMetrics, maxBatchSize*5),
		mediaAssets:    make(chan *chmodels.GMBMediaAssets, maxBatchSize*5),
		searchKeywords: make(chan *chmodels.GMBSearchKeywordsMonthly, maxBatchSize*5),
		localPosts:     make(chan *chmodels.GMBLocalPosts, maxBatchSize*5),
		reviews:        make(chan *chmodels.GMBReviews, maxBatchSize*5),
	}

	var pickedCount uint64
	var parsedDailyMetrics, parsedMediaAssets, parsedSearchKeywords, parsedLocalPosts, parsedReviews uint64
	var insertedDailyMetrics, insertedMediaAssets, insertedSearchKeywords, insertedLocalPosts, insertedReviews uint64

	var lastMessageTime int64 = time.Now().UnixNano()

	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, batches, sink, log, &batchWg, batchProcessorsPerType,
		&insertedDailyMetrics, &insertedMediaAssets, &insertedSearchKeywords,
		&insertedLocalPosts, &insertedReviews)

	var wgParsers sync.WaitGroup
	for i := 0; i < 5; i++ {
		wgParsers.Add(1)
		go dataParser(ctx, &wgParsers, i, jobs, batches, log,
			&parsedDailyMetrics, &parsedMediaAssets, &parsedSearchKeywords,
			&parsedLocalPosts, &parsedReviews)
	}

	stopMetrics := make(chan struct{})
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info().
					Int("parse_queue", len(jobs)).
					Uint64("picked", atomic.LoadUint64(&pickedCount)).
					Uint64("parsed_daily_metrics", atomic.LoadUint64(&parsedDailyMetrics)).
					Uint64("parsed_media_assets", atomic.LoadUint64(&parsedMediaAssets)).
					Uint64("parsed_search_keywords", atomic.LoadUint64(&parsedSearchKeywords)).
					Uint64("parsed_local_posts", atomic.LoadUint64(&parsedLocalPosts)).
					Uint64("parsed_reviews", atomic.LoadUint64(&parsedReviews)).
					Uint64("inserted_daily_metrics", atomic.LoadUint64(&insertedDailyMetrics)).
					Uint64("inserted_media_assets", atomic.LoadUint64(&insertedMediaAssets)).
					Uint64("inserted_search_keywords", atomic.LoadUint64(&insertedSearchKeywords)).
					Uint64("inserted_local_posts", atomic.LoadUint64(&insertedLocalPosts)).
					Uint64("inserted_reviews", atomic.LoadUint64(&insertedReviews)).
					Msg("pipeline metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	var wgConsumer sync.WaitGroup
	wgConsumer.Add(1)
	go func() {
		defer wgConsumer.Done()
		log.Info().Str("topic", rawDataTopic).Str("group", consumerGroup).Msg("Consuming GMB data topic...")
		err := consumer.Consume(ctx, []string{rawDataTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			atomic.AddUint64(&pickedCount, 1)
			jobs <- RawMessage{Topic: topic, Key: key, Value: value}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("function", "main").Str("stage", "consumer").Msg("Consumer error")
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Int("parser_workers", 5).
		Int("batch_processors_per_type", batchProcessorsPerType).
		Int("max_batch_size", maxBatchSize).
		Dur("batch_timeout", batchTimeout).
		Msg("GMB Analytics Sink started successfully")

	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping...")
	cancel()

	wgConsumer.Wait()
	close(jobs)
	wgParsers.Wait()
	close(batches.dailyMetrics)
	close(batches.mediaAssets)
	close(batches.searchKeywords)
	close(batches.localPosts)
	close(batches.reviews)
	batchWg.Wait()
	close(stopMetrics)

	log.Info().
		Uint64("total_picked", atomic.LoadUint64(&pickedCount)).
		Uint64("total_parsed_daily_metrics", atomic.LoadUint64(&parsedDailyMetrics)).
		Uint64("total_parsed_media_assets", atomic.LoadUint64(&parsedMediaAssets)).
		Uint64("total_parsed_search_keywords", atomic.LoadUint64(&parsedSearchKeywords)).
		Uint64("total_parsed_local_posts", atomic.LoadUint64(&parsedLocalPosts)).
		Uint64("total_parsed_reviews", atomic.LoadUint64(&parsedReviews)).
		Uint64("total_inserted_daily_metrics", atomic.LoadUint64(&insertedDailyMetrics)).
		Uint64("total_inserted_media_assets", atomic.LoadUint64(&insertedMediaAssets)).
		Uint64("total_inserted_search_keywords", atomic.LoadUint64(&insertedSearchKeywords)).
		Uint64("total_inserted_local_posts", atomic.LoadUint64(&insertedLocalPosts)).
		Uint64("total_inserted_reviews", atomic.LoadUint64(&insertedReviews)).
		Msg("GMB Analytics Sink stopped")
}

// dataParser parses raw GMB data and routes to appropriate batch collectors
func dataParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, batches *BatchCollectors, log *logger.Logger,
	parsedDM, parsedMA, parsedSK, parsedLP, parsedRV *uint64) {
	defer wg.Done()
	log.Info().Int("worker_id", id).Msg("GMB data parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			var rawData kafkamodels.RawGMBData
			if err := json.Unmarshal(msg.Value, &rawData); err != nil {
				log.Error().Err(err).Str("function", "dataParser").Str("stage", "unmarshal_raw_data").Msg("Failed to unmarshal raw GMB data")
				continue
			}

			// Re-marshal the Data field to get JSON bytes for further parsing
			dataBytes, err := json.Marshal(rawData.Data)
			if err != nil {
				log.Error().Err(err).Str("function", "dataParser").Str("stage", "marshal_data").Msg("Failed to marshal data field")
				continue
			}

			switch rawData.DataType {
			case "performance_metrics":
				parsePerformanceMetrics(ctx, dataBytes, rawData, batches.dailyMetrics, parsedDM, log)
			case "search_keywords":
				parseSearchKeywords(ctx, dataBytes, rawData, batches.searchKeywords, parsedSK, log)
			case "local_posts":
				parseLocalPosts(ctx, dataBytes, rawData, batches.localPosts, parsedLP, log)
			case "reviews":
				parseReviews(ctx, dataBytes, rawData, batches.reviews, parsedRV, log)
			case "media_assets":
				parseMediaAssets(ctx, dataBytes, rawData, batches.mediaAssets, parsedMA, log)
			default:
				log.Warn().Str("data_type", rawData.DataType).Msg("Unknown GMB data type")
			}
		}
	}
}

func parsePerformanceMetrics(ctx context.Context, dataBytes []byte, raw kafkamodels.RawGMBData, out chan<- *chmodels.GMBDailyMetrics, counter *uint64, log *logger.Logger) {
	var resp struct {
		MultiDailyMetricTimeSeries []struct {
			DailyMetricTimeSeries []struct {
				DailyMetric string `json:"dailyMetric"`
				TimeSeries  struct {
					DatedValues []struct {
						Date struct {
							Year  int `json:"year"`
							Month int `json:"month"`
							Day   int `json:"day"`
						} `json:"date"`
						Value string `json:"value"`
					} `json:"datedValues"`
				} `json:"timeSeries"`
			} `json:"dailyMetricTimeSeries"`
		} `json:"multiDailyMetricTimeSeries"`
	}
	if err := json.Unmarshal(dataBytes, &resp); err != nil {
		log.Warn().Err(err).Str("location_id", raw.LocationID).Msg("Failed to unmarshal performance metrics response")
		return
	}

	log.Info().
		Str("location_id", raw.LocationID).
		Int("multi_series_count", len(resp.MultiDailyMetricTimeSeries)).
		Msg("Parsing performance metrics")

	// Group by date then convert
	dateMap := make(map[string]*conversions.GMBDailyMetricsBuilder)
	for _, multi := range resp.MultiDailyMetricTimeSeries {
		for _, ts := range multi.DailyMetricTimeSeries {
			metric := ts.DailyMetric
			for _, dv := range ts.TimeSeries.DatedValues {
				dateStr := time.Date(dv.Date.Year, time.Month(dv.Date.Month), dv.Date.Day, 0, 0, 0, 0, time.UTC).Format("2006-01-02")
				builder, ok := dateMap[dateStr]
				if !ok {
					builder = &conversions.GMBDailyMetricsBuilder{
						AccountID:    raw.AccountID,
						LocationID:   raw.LocationID,
						AccountName:  raw.AccountName,
						LocationName: raw.LocationName,
						Date:         dateStr,
					}
					dateMap[dateStr] = builder
				}
				builder.SetMetric(metric, dv.Value)
			}
		}
	}

	log.Info().
		Str("location_id", raw.LocationID).
		Int("daily_rows", len(dateMap)).
		Msg("Performance metrics parsed, sending to batch processor")

	for _, builder := range dateMap {
		chMetric := builder.Build()
		if chMetric != nil {
			select {
			case out <- chMetric:
				atomic.AddUint64(counter, 1)
			case <-ctx.Done():
				return
			}
		}
	}
}

func parseSearchKeywords(ctx context.Context, dataBytes []byte, raw kafkamodels.RawGMBData, out chan<- *chmodels.GMBSearchKeywordsMonthly, counter *uint64, log *logger.Logger) {
	var resp struct {
		SearchKeywordsCounts []struct {
			SearchKeyword string `json:"searchKeyword"`
			InsightsValue struct {
				Value     string `json:"value"`
				Threshold string `json:"threshold"`
			} `json:"insightsValue"`
		} `json:"searchKeywordsCounts"`
	}
	if err := json.Unmarshal(dataBytes, &resp); err != nil {
		log.Warn().Err(err).Msg("Failed to unmarshal search keywords response")
		return
	}

	for _, kw := range resp.SearchKeywordsCounts {
		impVal, _ := strconv.ParseUint(kw.InsightsValue.Value, 10, 64)
		impThresh, _ := strconv.ParseUint(kw.InsightsValue.Threshold, 10, 64)

		parsed := &kafkamodels.ParsedGMBSearchKeyword{
			AccountID:            raw.AccountID,
			LocationID:           raw.LocationID,
			AccountName:          raw.AccountName,
			LocationName:         raw.LocationName,
			KeywordMonth:         raw.KeywordMonth,
			Keyword:              kw.SearchKeyword,
			ImpressionsValue:     impVal,
			ImpressionsThreshold: impThresh,
		}
		chKeyword := conversions.ConvertGMBSearchKeyword(parsed)
		if chKeyword != nil {
			select {
			case out <- chKeyword:
				atomic.AddUint64(counter, 1)
			case <-ctx.Done():
				return
			}
		}
	}
}

func parseLocalPosts(ctx context.Context, dataBytes []byte, raw kafkamodels.RawGMBData, out chan<- *chmodels.GMBLocalPosts, counter *uint64, log *logger.Logger) {
	var resp struct {
		LocalPosts []struct {
			Name       string `json:"name"`
			Summary    string `json:"summary"`
			State      string `json:"state"`
			TopicType  string `json:"topicType"`
			SearchURL  string `json:"searchUrl"`
			CreateTime string `json:"createTime"`
			UpdateTime string `json:"updateTime"`
			Media      []struct {
				Name      string `json:"name"`
				Format    string `json:"mediaFormat"`
				GoogleURL string `json:"googleUrl"`
			} `json:"media"`
		} `json:"localPosts"`
	}
	if err := json.Unmarshal(dataBytes, &resp); err != nil {
		log.Warn().Err(err).Msg("Failed to unmarshal local posts response")
		return
	}

	for _, post := range resp.LocalPosts {
		mediaNames := make([]string, len(post.Media))
		mediaFormats := make([]string, len(post.Media))
		mediaGoogleURLs := make([]string, len(post.Media))
		for j, m := range post.Media {
			mediaNames[j] = m.Name
			mediaFormats[j] = m.Format
			mediaGoogleURLs[j] = m.GoogleURL
		}

		parsed := &kafkamodels.ParsedGMBLocalPost{
			AccountID:       raw.AccountID,
			LocationID:      raw.LocationID,
			AccountName:     raw.AccountName,
			LocationName:    raw.LocationName,
			LanguageCode:    raw.LanguageCode,
			PostName:        post.Name,
			Summary:         post.Summary,
			State:           post.State,
			TopicType:       post.TopicType,
			SearchURL:       post.SearchURL,
			CreateTime:      post.CreateTime,
			UpdateTime:      post.UpdateTime,
			MediaNames:      mediaNames,
			MediaFormats:    mediaFormats,
			MediaGoogleURLs: mediaGoogleURLs,
		}
		chPost := conversions.ConvertGMBLocalPost(parsed)
		if chPost != nil {
			select {
			case out <- chPost:
				atomic.AddUint64(counter, 1)
			case <-ctx.Done():
				return
			}
		}
	}
}

func parseReviews(ctx context.Context, dataBytes []byte, raw kafkamodels.RawGMBData, out chan<- *chmodels.GMBReviews, counter *uint64, log *logger.Logger) {
	var resp struct {
		Reviews []struct {
			Name     string `json:"name"`
			ReviewID string `json:"reviewId"`
			Reviewer struct {
				DisplayName     string `json:"displayName"`
				ProfilePhotoURL string `json:"profilePhotoUrl"`
			} `json:"reviewer"`
			StarRating  string `json:"starRating"`
			Comment     string `json:"comment"`
			CreateTime  string `json:"createTime"`
			UpdateTime  string `json:"updateTime"`
			ReviewReply *struct {
				Comment    string `json:"comment"`
				UpdateTime string `json:"updateTime"`
			} `json:"reviewReply,omitempty"`
		} `json:"reviews"`
	}
	if err := json.Unmarshal(dataBytes, &resp); err != nil {
		log.Warn().Err(err).Msg("Failed to unmarshal reviews response")
		return
	}

	for _, review := range resp.Reviews {
		var replyComment, replyUpdateTime string
		if review.ReviewReply != nil {
			replyComment = review.ReviewReply.Comment
			replyUpdateTime = review.ReviewReply.UpdateTime
		}
		parsed := &kafkamodels.ParsedGMBReview{
			AccountID:               raw.AccountID,
			LocationID:              raw.LocationID,
			AccountName:             raw.AccountName,
			LocationName:            raw.LocationName,
			ReviewID:                review.ReviewID,
			ReviewName:              review.Name,
			ReviewerDisplayName:     review.Reviewer.DisplayName,
			ReviewerProfilePhotoURL: review.Reviewer.ProfilePhotoURL,
			StarRating:              review.StarRating,
			Comment:                 review.Comment,
			CreateTime:              review.CreateTime,
			UpdateTime:              review.UpdateTime,
			ReplyComment:            replyComment,
			ReplyUpdateTime:         replyUpdateTime,
		}
		chReview := conversions.ConvertGMBReview(parsed)
		if chReview != nil {
			select {
			case out <- chReview:
				atomic.AddUint64(counter, 1)
			case <-ctx.Done():
				return
			}
		}
	}
}

func parseMediaAssets(ctx context.Context, dataBytes []byte, raw kafkamodels.RawGMBData, out chan<- *chmodels.GMBMediaAssets, counter *uint64, log *logger.Logger) {
	var resp struct {
		MediaItems []struct {
			Name                string `json:"name"`
			MediaFormat         string `json:"mediaFormat"`
			LocationAssociation struct {
				Category string `json:"category"`
			} `json:"locationAssociation"`
			GoogleURL    string `json:"googleUrl"`
			ThumbnailURL string `json:"thumbnailUrl"`
			SourceURL    string `json:"sourceUrl"`
			DataRef      struct {
				ResourceName string `json:"resourceName"`
			} `json:"dataRef"`
			Dimensions struct {
				WidthPixels  uint64 `json:"widthPixels"`
				HeightPixels uint64 `json:"heightPixels"`
			} `json:"dimensions"`
			CreateTime string `json:"createTime"`
		} `json:"mediaItems"`
	}
	if err := json.Unmarshal(dataBytes, &resp); err != nil {
		log.Warn().Err(err).Msg("Failed to unmarshal media assets response")
		return
	}

	for _, media := range resp.MediaItems {
		parsed := &kafkamodels.ParsedGMBMediaAsset{
			AccountID:                   raw.AccountID,
			LocationID:                  raw.LocationID,
			AccountName:                 raw.AccountName,
			LocationName:                raw.LocationName,
			LanguageCode:                raw.LanguageCode,
			MediaName:                   media.Name,
			SourceURL:                   media.SourceURL,
			MediaFormat:                 media.MediaFormat,
			LocationAssociationCategory: media.LocationAssociation.Category,
			GoogleURL:                   media.GoogleURL,
			ThumbnailURL:                media.ThumbnailURL,
			WidthPixels:                 media.Dimensions.WidthPixels,
			HeightPixels:                media.Dimensions.HeightPixels,
			CreateTime:                  media.CreateTime,
		}
		chAsset := conversions.ConvertGMBMediaAsset(parsed)
		if chAsset != nil {
			select {
			case out <- chAsset:
				atomic.AddUint64(counter, 1)
			case <-ctx.Done():
				return
			}
		}
	}
}

// startBatchProcessors starts batch processors for all 5 GMB data types
func startBatchProcessors(ctx context.Context, batches *BatchCollectors, sink ClickHouseSinkInterface, log *logger.Logger, wg *sync.WaitGroup, numProcessors int,
	insertedDM, insertedMA, insertedSK, insertedLP, insertedRV *uint64) {
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go dailyMetricsBatchProcessor(ctx, wg, i, batches.dailyMetrics, sink, log, insertedDM)
	}
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go mediaAssetsBatchProcessor(ctx, wg, i, batches.mediaAssets, sink, log, insertedMA)
	}
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go searchKeywordsBatchProcessor(ctx, wg, i, batches.searchKeywords, sink, log, insertedSK)
	}
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go localPostsBatchProcessor(ctx, wg, i, batches.localPosts, sink, log, insertedLP)
	}
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go reviewsBatchProcessor(ctx, wg, i, batches.reviews, sink, log, insertedRV)
	}
}

func dailyMetricsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.GMBDailyMetrics, sink ClickHouseSinkInterface, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	batch := make([]*chmodels.GMBDailyMetrics, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flush := func(stage string) {
		if len(batch) == 0 {
			return
		}
		log.Info().Str("stage", stage).Int("batch_size", len(batch)).Msg("Flushing daily metrics batch to ClickHouse")
		if err := sink.BulkInsertGMBDailyMetrics(ctx, batch); err != nil {
			log.Error().Err(err).Str("function", "dailyMetricsBatchProcessor").Str("stage", stage).Int("batch_size", len(batch)).Msg("Failed to insert daily metrics batch")
		} else {
			inserted := uint64(len(batch))
			atomic.AddUint64(insertedCounter, inserted)
			log.Info().Uint64("inserted", inserted).Uint64("total_inserted", atomic.LoadUint64(insertedCounter)).Str("stage", stage).Msg("Daily metrics batch inserted successfully")
		}
		batch = make([]*chmodels.GMBDailyMetrics, 0, maxBatchSize)
	}

	for {
		select {
		case <-ctx.Done():
			flush("flush_final")
			return
		case <-flushTimer.C:
			flush("flush_timer")
			flushTimer.Reset(batchTimeout)
		case item, ok := <-in:
			if !ok {
				flush("flush_close")
				return
			}
			batch = append(batch, item)
			if len(batch) >= maxBatchSize {
				flush("bulk_insert")
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

func mediaAssetsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.GMBMediaAssets, sink ClickHouseSinkInterface, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	batch := make([]*chmodels.GMBMediaAssets, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flush := func(stage string) {
		if len(batch) == 0 {
			return
		}
		if err := sink.BulkInsertGMBMediaAssets(ctx, batch); err != nil {
			log.Error().Err(err).Str("function", "mediaAssetsBatchProcessor").Str("stage", stage).Int("batch_size", len(batch)).Msg("Failed to insert media assets batch")
		} else {
			atomic.AddUint64(insertedCounter, uint64(len(batch)))
		}
		batch = make([]*chmodels.GMBMediaAssets, 0, maxBatchSize)
	}

	for {
		select {
		case <-ctx.Done():
			flush("flush_final")
			return
		case <-flushTimer.C:
			flush("flush_timer")
			flushTimer.Reset(batchTimeout)
		case item, ok := <-in:
			if !ok {
				flush("flush_close")
				return
			}
			batch = append(batch, item)
			if len(batch) >= maxBatchSize {
				flush("bulk_insert")
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

func searchKeywordsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.GMBSearchKeywordsMonthly, sink ClickHouseSinkInterface, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	batch := make([]*chmodels.GMBSearchKeywordsMonthly, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flush := func(stage string) {
		if len(batch) == 0 {
			return
		}
		if err := sink.BulkInsertGMBSearchKeywordsMonthly(ctx, batch); err != nil {
			log.Error().Err(err).Str("function", "searchKeywordsBatchProcessor").Str("stage", stage).Int("batch_size", len(batch)).Msg("Failed to insert search keywords batch")
		} else {
			atomic.AddUint64(insertedCounter, uint64(len(batch)))
		}
		batch = make([]*chmodels.GMBSearchKeywordsMonthly, 0, maxBatchSize)
	}

	for {
		select {
		case <-ctx.Done():
			flush("flush_final")
			return
		case <-flushTimer.C:
			flush("flush_timer")
			flushTimer.Reset(batchTimeout)
		case item, ok := <-in:
			if !ok {
				flush("flush_close")
				return
			}
			batch = append(batch, item)
			if len(batch) >= maxBatchSize {
				flush("bulk_insert")
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

func localPostsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.GMBLocalPosts, sink ClickHouseSinkInterface, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	batch := make([]*chmodels.GMBLocalPosts, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flush := func(stage string) {
		if len(batch) == 0 {
			return
		}
		if err := sink.BulkInsertGMBLocalPosts(ctx, batch); err != nil {
			log.Error().Err(err).Str("function", "localPostsBatchProcessor").Str("stage", stage).Int("batch_size", len(batch)).Msg("Failed to insert local posts batch")
		} else {
			atomic.AddUint64(insertedCounter, uint64(len(batch)))
		}
		batch = make([]*chmodels.GMBLocalPosts, 0, maxBatchSize)
	}

	for {
		select {
		case <-ctx.Done():
			flush("flush_final")
			return
		case <-flushTimer.C:
			flush("flush_timer")
			flushTimer.Reset(batchTimeout)
		case item, ok := <-in:
			if !ok {
				flush("flush_close")
				return
			}
			batch = append(batch, item)
			if len(batch) >= maxBatchSize {
				flush("bulk_insert")
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

func reviewsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.GMBReviews, sink ClickHouseSinkInterface, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	batch := make([]*chmodels.GMBReviews, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flush := func(stage string) {
		if len(batch) == 0 {
			return
		}
		if err := sink.BulkInsertGMBReviews(ctx, batch); err != nil {
			log.Error().Err(err).Str("function", "reviewsBatchProcessor").Str("stage", stage).Int("batch_size", len(batch)).Msg("Failed to insert reviews batch")
		} else {
			atomic.AddUint64(insertedCounter, uint64(len(batch)))
		}
		batch = make([]*chmodels.GMBReviews, 0, maxBatchSize)
	}

	for {
		select {
		case <-ctx.Done():
			flush("flush_final")
			return
		case <-flushTimer.C:
			flush("flush_timer")
			flushTimer.Reset(batchTimeout)
		case item, ok := <-in:
			if !ok {
				flush("flush_close")
				return
			}
			batch = append(batch, item)
			if len(batch) >= maxBatchSize {
				flush("bulk_insert")
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}
