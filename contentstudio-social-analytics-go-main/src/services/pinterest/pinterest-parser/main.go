package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
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
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	parserWorkers    = 6
	publisherWorkers = 6

	parseChanSize   = 500
	publishChanSize = 1000

	metricsEvery = 10 * time.Second

	topicRawUsers        = "raw-pinterest-users"
	topicRawBoards       = "raw-pinterest-boards"
	topicRawPins         = "raw-pinterest-pins"
	topicRawPinInsights  = "raw-pinterest-pin-insights"
	topicRawUserInsights = "raw-pinterest-user-insights"

	topicParsedUsers        = "parsed-pinterest-users"
	topicParsedBoards       = "parsed-pinterest-boards"
	topicParsedPins         = "parsed-pinterest-pins"
	topicParsedPinInsights  = "parsed-pinterest-pin-insights"
	topicParsedUserInsights = "parsed-pinterest-user-insights"

	consumerGroup = "pinterest-parser-group"
)

type ParseJob struct {
	JobType     string
	Key         []byte
	Value       []byte
	OutputTopic string
}

type PublishJob struct {
	Topic string
	Key   string
	Data  []byte
}

func generateRecordID(id string, date time.Time) string {
	hash := md5.Sum([]byte(id + "_" + date.Format("20060102")))
	return hex.EncodeToString(hash[:])
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic(err)
	}
	telemetry.ConfigureSentry(cfg)
	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("parser_workers", parserWorkers).
		Int("publisher_workers", publisherWorkers).
		Str("consumer_group", consumerGroup).
		Msg("Starting Pinterest Parser service")

	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

	usersConsumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create users consumer")
	}
	defer usersConsumer.Close()

	boardsConsumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create boards consumer")
	}
	defer boardsConsumer.Close()

	pinsConsumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create pins consumer")
	}
	defer pinsConsumer.Close()

	pinInsightsConsumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create pin insights consumer")
	}
	defer pinInsightsConsumer.Close()

	userInsightsConsumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create user insights consumer")
	}
	defer userInsightsConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGTERM, syscall.SIGINT)

	parseJobs := make(chan ParseJob, parseChanSize)
	publishJobs := make(chan PublishJob, publishChanSize)

	var pickedCount, pubCount uint64

	var wgParsers sync.WaitGroup
	for i := 0; i < parserWorkers; i++ {
		wgParsers.Add(1)
		go parserWorker(ctx, &wgParsers, i, parseJobs, publishJobs, log)
	}

	var wgPublishers sync.WaitGroup
	for i := 0; i < publisherWorkers; i++ {
		wgPublishers.Add(1)
		go publisherWorker(ctx, &wgPublishers, i, publishJobs, producer, &pubCount, log)
	}

	stopMetrics := make(chan struct{})
	go func() {
		t := time.NewTicker(metricsEvery)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info().
					Int("queue_parse", len(parseJobs)).
					Int("queue_publish", len(publishJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedCount)).
					Uint64("published", atomic.LoadUint64(&pubCount)).
					Msg("metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	var wgConsumers sync.WaitGroup
	wgConsumers.Add(5)

	go func() {
		defer wgConsumers.Done()
		topics := []string{topicRawUsers}
		log.Info().Strs("topics", topics).Msg("Starting users consumer")
		if err := usersConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case parseJobs <- ParseJob{JobType: "user", Key: key, Value: value, OutputTopic: topicParsedUsers}:
				atomic.AddUint64(&pickedCount, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_users").Msg("Users consumer error")
		}
	}()

	go func() {
		defer wgConsumers.Done()
		topics := []string{topicRawBoards}
		log.Info().Strs("topics", topics).Msg("Starting boards consumer")
		if err := boardsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case parseJobs <- ParseJob{JobType: "board", Key: key, Value: value, OutputTopic: topicParsedBoards}:
				atomic.AddUint64(&pickedCount, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_boards").Msg("Boards consumer error")
		}
	}()

	go func() {
		defer wgConsumers.Done()
		topics := []string{topicRawPins}
		log.Info().Strs("topics", topics).Msg("Starting pins consumer")
		if err := pinsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case parseJobs <- ParseJob{JobType: "pin", Key: key, Value: value, OutputTopic: topicParsedPins}:
				atomic.AddUint64(&pickedCount, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_pins").Msg("Pins consumer error")
		}
	}()

	go func() {
		defer wgConsumers.Done()
		topics := []string{topicRawPinInsights}
		log.Info().Strs("topics", topics).Msg("Starting pin insights consumer")
		if err := pinInsightsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case parseJobs <- ParseJob{JobType: "pin_insight", Key: key, Value: value, OutputTopic: topicParsedPinInsights}:
				atomic.AddUint64(&pickedCount, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_pin_insights").Msg("Pin insights consumer error")
		}
	}()

	go func() {
		defer wgConsumers.Done()
		topics := []string{topicRawUserInsights}
		log.Info().Strs("topics", topics).Msg("Starting user insights consumer")
		if err := userInsightsConsumer.Consume(ctx, topics, func(ctx context.Context, topic string, key, value []byte) error {
			select {
			case parseJobs <- ParseJob{JobType: "user_insight", Key: key, Value: value, OutputTopic: topicParsedUserInsights}:
				atomic.AddUint64(&pickedCount, 1)
				return nil
			case <-ctx.Done():
				return ctx.Err()
			}
		}); err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consume_user_insights").Msg("User insights consumer error")
		}
	}()

	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()

	wgConsumers.Wait()

	close(parseJobs)
	wgParsers.Wait()

	close(publishJobs)
	wgPublishers.Wait()

	close(stopMetrics)

	log.Info().Msg("Pinterest Parser service stopped")
}

func parserWorker(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan ParseJob, out chan<- PublishJob, log *logger.Logger) {
	defer wg.Done()
	workerLog := &logger.Logger{Logger: log.With().Int("worker_id", id).Logger()}
	workerLog.Info().Msg("Parser worker started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Parser worker stopped (context cancelled)")
			return
		case job, ok := <-in:
			if !ok {
				workerLog.Info().Msg("Parser worker stopped (channel closed)")
				return
			}
			parseAndQueue(ctx, job, out, workerLog)
		}
	}
}

func publisherWorker(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan PublishJob, producer kafka2.Producer, counter *uint64, log *logger.Logger) {
	defer wg.Done()
	workerLog := &logger.Logger{Logger: log.With().Int("worker_id", id).Logger()}
	workerLog.Info().Msg("Publisher worker started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Publisher worker stopped (context cancelled)")
			return
		case job, ok := <-in:
			if !ok {
				workerLog.Info().Msg("Publisher worker stopped (channel closed)")
				return
			}
			if err := producer.Produce(ctx, job.Topic, []byte(job.Key), job.Data); err != nil {
				workerLog.Error().Err(err).Str("error_message", err.Error()).Str("topic", job.Topic).Str("key", job.Key).Str("function", "publisherWorker").Str("stage", "produce_kafka").Msg("Failed to produce message")
				continue
			}
			atomic.AddUint64(counter, 1)
		}
	}
}

func parseAndQueue(ctx context.Context, job ParseJob, out chan<- PublishJob, log *logger.Logger) {
	switch job.JobType {
	case "user":
		parseUser(ctx, job, out, log)
	case "board":
		parseBoard(ctx, job, out, log)
	case "pin":
		parsePin(ctx, job, out, log)
	case "pin_insight":
		parsePinInsight(ctx, job, out, log)
	case "user_insight":
		parseUserInsight(ctx, job, out, log)
	default:
		log.Warn().Str("job_type", job.JobType).Msg("Unknown job type")
	}
}

func parseUser(ctx context.Context, job ParseJob, out chan<- PublishJob, log *logger.Logger) {
	var raw kafkamodels.RawPinterestUser
	if err := json.Unmarshal(job.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseUser").Str("stage", "unmarshal").Msg("Failed to unmarshal raw user")
		return
	}

	now := time.Now().UTC()
	parsed := kafkamodels.ParsedPinterestUser{
		RecordID:       generateRecordID(raw.UserID, now),
		UserID:         raw.UserID,
		Username:       raw.Username,
		About:          raw.About,
		ProfileImage:   raw.ProfileImage,
		WebsiteURL:     raw.WebsiteURL,
		BusinessName:   raw.BusinessName,
		BoardCount:     raw.BoardCount,
		PinCount:       raw.PinCount,
		AccountType:    raw.AccountType,
		FollowerCount:  raw.FollowerCount,
		FollowingCount: raw.FollowingCount,
		MonthlyViews:   raw.MonthlyViews,
		InsertedAt:     now,
	}

	data, _ := json.Marshal(parsed)
	select {
	case out <- PublishJob{Topic: job.OutputTopic, Key: parsed.RecordID, Data: data}:
	case <-ctx.Done():
		return
	}
}

func parseBoard(ctx context.Context, job ParseJob, out chan<- PublishJob, log *logger.Logger) {
	var raw kafkamodels.RawPinterestBoard
	if err := json.Unmarshal(job.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseBoard").Str("stage", "unmarshal").Msg("Failed to unmarshal raw board")
		return
	}

	now := time.Now().UTC()
	parsed := kafkamodels.ParsedPinterestBoard{
		RecordID:          generateRecordID(raw.BoardID, now),
		BoardID:           raw.BoardID,
		UserID:            raw.UserID,
		Name:              raw.Name,
		Description:       raw.Description,
		Privacy:           raw.Privacy,
		PinCount:          fmt.Sprintf("%d", raw.PinCount),
		FollowerCount:     fmt.Sprintf("%d", raw.FollowerCount),
		CollaboratorCount: fmt.Sprintf("%d", raw.CollaboratorCount),
		Owner:             raw.Owner,
		ImageCoverURL:     raw.ImageCoverURL,
		PinThumbnailURLs:  raw.PinThumbnailURLs,
		CreatedAt:         raw.CreatedAt,
		InsertedAt:        now,
	}

	data, _ := json.Marshal(parsed)
	select {
	case out <- PublishJob{Topic: job.OutputTopic, Key: parsed.RecordID, Data: data}:
	case <-ctx.Done():
		return
	}
}

func parsePin(ctx context.Context, job ParseJob, out chan<- PublishJob, log *logger.Logger) {
	var raw kafkamodels.RawPinterestPin
	if err := json.Unmarshal(job.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parsePin").Str("stage", "unmarshal").Msg("Failed to unmarshal raw pin")
		return
	}

	now := time.Now().UTC()
	isStandard := "0"
	if raw.IsStandard {
		isStandard = "1"
	}
	isOwner := "0"
	if raw.IsOwner {
		isOwner = "1"
	}
	hasBeenPromoted := "0"
	if raw.HasBeenPromoted {
		hasBeenPromoted = "1"
	}

	parsed := kafkamodels.ParsedPinterestPin{
		RecordID:        generateRecordID(raw.PinID, now),
		PinID:           raw.PinID,
		UserID:          raw.UserID,
		BoardID:         raw.BoardID,
		BoardSectionID:  raw.BoardSectionID,
		ParentPinID:     raw.ParentPinID,
		Title:           raw.Title,
		Note:            raw.Note,
		Description:     raw.Description,
		Link:            raw.Link,
		DominantColor:   raw.DominantColor,
		CreativeType:    raw.CreativeType,
		MediaType:       raw.MediaType,
		CoverImageURL:   raw.CoverImageURL,
		VideoURL:        raw.VideoURL,
		Duration:        raw.Duration,
		Height:          raw.Height,
		Width:           raw.Width,
		IsStandard:      isStandard,
		IsOwner:         isOwner,
		HasBeenPromoted: hasBeenPromoted,
		BoardOwner:      raw.BoardOwner,
		ProductTags:     raw.ProductTags,
		CreatedAt:       raw.CreatedAt,
		DayOfWeek:       raw.DayOfWeek,
		HourOfDay:       raw.HourOfDay,
		InsertedAt:      now,
	}

	data, _ := json.Marshal(parsed)
	select {
	case out <- PublishJob{Topic: job.OutputTopic, Key: parsed.RecordID, Data: data}:
	case <-ctx.Done():
		return
	}
}

func parsePinInsight(ctx context.Context, job ParseJob, out chan<- PublishJob, log *logger.Logger) {
	var raw kafkamodels.RawPinterestPinInsight
	if err := json.Unmarshal(job.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parsePinInsight").Str("stage", "unmarshal").Msg("Failed to unmarshal raw pin insight")
		return
	}

	now := time.Now().UTC()
	parsed := kafkamodels.ParsedPinterestPinInsight{
		RecordID:           generateRecordID(raw.PinID, raw.Date),
		PinID:              raw.PinID,
		UserID:             raw.UserID,
		BoardID:            raw.BoardID,
		Date:               raw.Date,
		DataStatus:         raw.DataStatus,
		Impression:         raw.Impression,
		PinClicks:          raw.PinClicks,
		OutboundClicks:     raw.OutboundClicks,
		Saves:              raw.Saves,
		SaveRate:           raw.SaveRate,
		Clickthrough:       raw.Clickthrough,
		ClickthroughRate:   raw.ClickthroughRate,
		Engagement:         raw.Engagement,
		EngagementRate:     raw.EngagementRate,
		VideoMRCView:       raw.VideoMRCView,
		VideoStart:         raw.VideoStart,
		Video10sView:       raw.Video10sView,
		VideoAvgWatchTime:  raw.VideoAvgWatchTime,
		VideoV50WatchTime:  raw.VideoV50WatchTime,
		FullScreenPlay:     raw.FullScreenPlay,
		FullScreenPlaytime: raw.FullScreenPlaytime,
		ProfileVisit:       raw.ProfileVisit,
		Closeup:            raw.Closeup,
		Quartile95sPercent: raw.Quartile95sPercent,
		UserFollow:         raw.UserFollow,
		DayOfWeek:          raw.Date.Weekday().String(),
		HourOfDay:          raw.Date.Hour(),
		InsertedAt:         now,
	}

	data, _ := json.Marshal(parsed)
	select {
	case out <- PublishJob{Topic: job.OutputTopic, Key: parsed.RecordID, Data: data}:
	case <-ctx.Done():
		return
	}
}

func parseUserInsight(ctx context.Context, job ParseJob, out chan<- PublishJob, log *logger.Logger) {
	var raw kafkamodels.RawPinterestUserInsight
	if err := json.Unmarshal(job.Value, &raw); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "parseUserInsight").Str("stage", "unmarshal").Msg("Failed to unmarshal raw user insight")
		return
	}

	now := time.Now().UTC()
	parsed := kafkamodels.ParsedPinterestUserInsight{
		RecordID:           generateRecordID(raw.UserID, raw.Date),
		UserID:             raw.UserID,
		Date:               raw.Date,
		DataStatus:         raw.DataStatus,
		Impression:         raw.Impression,
		PinClicks:          raw.PinClicks,
		PinClickRate:       raw.PinClickRate,
		OutboundClicks:     raw.OutboundClicks,
		Saves:              raw.Saves,
		SaveRate:           raw.SaveRate,
		Clickthrough:       raw.Clickthrough,
		ClickthroughRate:   raw.ClickthroughRate,
		Engagement:         raw.Engagement,
		EngagementRate:     raw.EngagementRate,
		VideoMRCView:       raw.VideoMRCView,
		VideoStart:         raw.VideoStart,
		Video10sView:       raw.Video10sView,
		VideoAvgWatchTime:  raw.VideoAvgWatchTime,
		VideoV50WatchTime:  raw.VideoV50WatchTime,
		FullScreenPlay:     raw.FullScreenPlay,
		FullScreenPlaytime: raw.FullScreenPlaytime,
		ProfileVisit:       raw.ProfileVisit,
		Closeup:            raw.Closeup,
		Quartile95sPercent: raw.Quartile95sPercent,
		InsertedAt:         now,
	}

	data, _ := json.Marshal(parsed)
	select {
	case out <- PublishJob{Topic: job.OutputTopic, Key: parsed.RecordID, Data: data}:
	case <-ctx.Done():
		return
	}
}
