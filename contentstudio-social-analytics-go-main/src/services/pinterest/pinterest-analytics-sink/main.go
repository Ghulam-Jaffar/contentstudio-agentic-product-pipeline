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
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	maxBatchSize           = 10000
	batchTimeout           = 10 * time.Second
	batchProcessorsPerType = 3
	messageChanSize        = 50000

	consumerGroup = "pinterest-analytics-sink-group"

	idleTimeout       = 5 * time.Minute
	idleCheckInterval = 30 * time.Second
)

type RawMessage struct {
	Topic string
	Key   []byte
	Value []byte
}

type BatchCollectors struct {
	users        chan *chmodels.PinterestUser
	boards       chan *chmodels.PinterestBoard
	pins         chan *chmodels.PinterestPin
	pinInsights  chan *chmodels.PinterestPinInsight
	userInsights chan *chmodels.PinterestUserInsight
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Pinterest Analytics Sink (merged parser+sink)")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed - continuing anyway")
	}

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

	usersJobs := make(chan RawMessage, messageChanSize)
	boardsJobs := make(chan RawMessage, messageChanSize)
	pinsJobs := make(chan RawMessage, messageChanSize)
	pinInsightsJobs := make(chan RawMessage, messageChanSize)
	userInsightsJobs := make(chan RawMessage, messageChanSize)

	batches := &BatchCollectors{
		users:        make(chan *chmodels.PinterestUser, maxBatchSize*5),
		boards:       make(chan *chmodels.PinterestBoard, maxBatchSize*5),
		pins:         make(chan *chmodels.PinterestPin, maxBatchSize*5),
		pinInsights:  make(chan *chmodels.PinterestPinInsight, maxBatchSize*5),
		userInsights: make(chan *chmodels.PinterestUserInsight, maxBatchSize*5),
	}

	var pickedUsers, pickedBoards, pickedPins, pickedPinInsights, pickedUserInsights uint64
	var parsedUsers, parsedBoards, parsedPins, parsedPinInsights, parsedUserInsights uint64
	var insertedUsers, insertedBoards, insertedPins, insertedPinInsights, insertedUserInsights uint64

	var lastMessageTime int64 = time.Now().UnixNano()

	var batchWg sync.WaitGroup
	startBatchProcessors(ctx, batches, sink, log, &batchWg, batchProcessorsPerType,
		&insertedUsers, &insertedBoards, &insertedPins, &insertedPinInsights, &insertedUserInsights)

	var wgParsers sync.WaitGroup
	for i := 0; i < 3; i++ {
		wgParsers.Add(1)
		go usersParser(ctx, &wgParsers, i, usersJobs, batches.users, &parsedUsers, log)
	}
	for i := 0; i < 3; i++ {
		wgParsers.Add(1)
		go boardsParser(ctx, &wgParsers, i, boardsJobs, batches.boards, &parsedBoards, log)
	}
	for i := 0; i < 5; i++ {
		wgParsers.Add(1)
		go pinsParser(ctx, &wgParsers, i, pinsJobs, batches.pins, &parsedPins, log)
	}
	for i := 0; i < 5; i++ {
		wgParsers.Add(1)
		go pinInsightsParser(ctx, &wgParsers, i, pinInsightsJobs, batches.pinInsights, &parsedPinInsights, log)
	}
	for i := 0; i < 3; i++ {
		wgParsers.Add(1)
		go userInsightsParser(ctx, &wgParsers, i, userInsightsJobs, batches.userInsights, &parsedUserInsights, log)
	}

	stopMetrics := make(chan struct{})
	go func() {
		t := time.NewTicker(10 * time.Second)
		defer t.Stop()
		for {
			select {
			case <-t.C:
				log.Info().
					Str("pipeline", "users").
					Int("parse_queue", len(usersJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedUsers)).
					Uint64("parsed", atomic.LoadUint64(&parsedUsers)).
					Uint64("inserted", atomic.LoadUint64(&insertedUsers)).
					Int("batch_queue", len(batches.users)).
					Msg("pipeline metrics")
				log.Info().
					Str("pipeline", "boards").
					Int("parse_queue", len(boardsJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedBoards)).
					Uint64("parsed", atomic.LoadUint64(&parsedBoards)).
					Uint64("inserted", atomic.LoadUint64(&insertedBoards)).
					Int("batch_queue", len(batches.boards)).
					Msg("pipeline metrics")
				log.Info().
					Str("pipeline", "pins").
					Int("parse_queue", len(pinsJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedPins)).
					Uint64("parsed", atomic.LoadUint64(&parsedPins)).
					Uint64("inserted", atomic.LoadUint64(&insertedPins)).
					Int("batch_queue", len(batches.pins)).
					Msg("pipeline metrics")
				log.Info().
					Str("pipeline", "pin_insights").
					Int("parse_queue", len(pinInsightsJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedPinInsights)).
					Uint64("parsed", atomic.LoadUint64(&parsedPinInsights)).
					Uint64("inserted", atomic.LoadUint64(&insertedPinInsights)).
					Int("batch_queue", len(batches.pinInsights)).
					Msg("pipeline metrics")
				log.Info().
					Str("pipeline", "user_insights").
					Int("parse_queue", len(userInsightsJobs)).
					Uint64("picked", atomic.LoadUint64(&pickedUserInsights)).
					Uint64("parsed", atomic.LoadUint64(&parsedUserInsights)).
					Uint64("inserted", atomic.LoadUint64(&insertedUserInsights)).
					Int("batch_queue", len(batches.userInsights)).
					Msg("pipeline metrics")
			case <-stopMetrics:
				return
			}
		}
	}()

	var wgConsumers sync.WaitGroup
	wgConsumers.Add(5)

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", kafkamodels.PinterestKafkaTopics.RawUsers).Str("group", consumerGroup).Msg("Consuming users topic...")
		err := usersConsumer.Consume(ctx, []string{kafkamodels.PinterestKafkaTopics.RawUsers}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			atomic.AddUint64(&pickedUsers, 1)
			usersJobs <- RawMessage{Topic: topic, Key: key, Value: value}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_users").Msg("Users consumer error")
			cancel()
		}
	}()

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", kafkamodels.PinterestKafkaTopics.RawBoards).Str("group", consumerGroup).Msg("Consuming boards topic...")
		err := boardsConsumer.Consume(ctx, []string{kafkamodels.PinterestKafkaTopics.RawBoards}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			atomic.AddUint64(&pickedBoards, 1)
			boardsJobs <- RawMessage{Topic: topic, Key: key, Value: value}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_boards").Msg("Boards consumer error")
			cancel()
		}
	}()

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", kafkamodels.PinterestKafkaTopics.RawPins).Str("group", consumerGroup).Msg("Consuming pins topic...")
		err := pinsConsumer.Consume(ctx, []string{kafkamodels.PinterestKafkaTopics.RawPins}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			atomic.AddUint64(&pickedPins, 1)
			pinsJobs <- RawMessage{Topic: topic, Key: key, Value: value}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_pins").Msg("Pins consumer error")
			cancel()
		}
	}()

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", kafkamodels.PinterestKafkaTopics.RawPinInsights).Str("group", consumerGroup).Msg("Consuming pin insights topic...")
		err := pinInsightsConsumer.Consume(ctx, []string{kafkamodels.PinterestKafkaTopics.RawPinInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			atomic.AddUint64(&pickedPinInsights, 1)
			pinInsightsJobs <- RawMessage{Topic: topic, Key: key, Value: value}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_pin_insights").Msg("Pin insights consumer error")
			cancel()
		}
	}()

	go func() {
		defer wgConsumers.Done()
		log.Info().Str("topic", kafkamodels.PinterestKafkaTopics.RawUserInsights).Str("group", consumerGroup).Msg("Consuming user insights topic...")
		err := userInsightsConsumer.Consume(ctx, []string{kafkamodels.PinterestKafkaTopics.RawUserInsights}, func(ctx context.Context, topic string, key, value []byte) error {
			atomic.StoreInt64(&lastMessageTime, time.Now().UnixNano())
			atomic.AddUint64(&pickedUserInsights, 1)
			userInsightsJobs <- RawMessage{Topic: topic, Key: key, Value: value}
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "consumer_user_insights").Msg("User insights consumer error")
			cancel()
		}
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Int("users_parser_workers", 3).
		Int("boards_parser_workers", 3).
		Int("pins_parser_workers", 5).
		Int("pin_insights_parser_workers", 5).
		Int("user_insights_parser_workers", 3).
		Int("batch_processors_per_type", batchProcessorsPerType).
		Int("max_batch_size", maxBatchSize).
		Dur("batch_timeout", batchTimeout).
		Msg("Pinterest Analytics Sink started successfully")

	<-sigChan
	log.Info().Msg("Shutdown signal received, stopping...")
	cancel()

	wgConsumers.Wait()

	close(usersJobs)
	close(boardsJobs)
	close(pinsJobs)
	close(pinInsightsJobs)
	close(userInsightsJobs)

	wgParsers.Wait()

	close(batches.users)
	close(batches.boards)
	close(batches.pins)
	close(batches.pinInsights)
	close(batches.userInsights)

	batchWg.Wait()

	close(stopMetrics)

	log.Info().
		Uint64("total_picked_users", atomic.LoadUint64(&pickedUsers)).
		Uint64("total_picked_boards", atomic.LoadUint64(&pickedBoards)).
		Uint64("total_picked_pins", atomic.LoadUint64(&pickedPins)).
		Uint64("total_picked_pin_insights", atomic.LoadUint64(&pickedPinInsights)).
		Uint64("total_picked_user_insights", atomic.LoadUint64(&pickedUserInsights)).
		Uint64("total_parsed_users", atomic.LoadUint64(&parsedUsers)).
		Uint64("total_parsed_boards", atomic.LoadUint64(&parsedBoards)).
		Uint64("total_parsed_pins", atomic.LoadUint64(&parsedPins)).
		Uint64("total_parsed_pin_insights", atomic.LoadUint64(&parsedPinInsights)).
		Uint64("total_parsed_user_insights", atomic.LoadUint64(&parsedUserInsights)).
		Uint64("total_inserted_users", atomic.LoadUint64(&insertedUsers)).
		Uint64("total_inserted_boards", atomic.LoadUint64(&insertedBoards)).
		Uint64("total_inserted_pins", atomic.LoadUint64(&insertedPins)).
		Uint64("total_inserted_pin_insights", atomic.LoadUint64(&insertedPinInsights)).
		Uint64("total_inserted_user_insights", atomic.LoadUint64(&insertedUserInsights)).
		Msg("Pinterest Analytics Sink stopped")
}

func usersParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, out chan<- *chmodels.PinterestUser, parsedCounter *uint64, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker_id", id).Str("pool", "users").Msg("Users parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			var rawUser kafkamodels.RawPinterestUser
			if err := json.Unmarshal(msg.Value, &rawUser); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "usersParser").Str("stage", "unmarshal_raw_user").Str("key", string(msg.Key)).Msg("Failed to unmarshal raw user")
				continue
			}

			if rawUser.UserID == "" {
				log.Debug().Str("key", string(msg.Key)).Msg("Raw user has empty UserID, skipping")
				continue
			}

			parsedUser := parseRawUser(&rawUser)
			chUser := conversions.ConvertPinterestUser(parsedUser)
			if chUser != nil {
				select {
				case out <- chUser:
					atomic.AddUint64(parsedCounter, 1)
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func boardsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, out chan<- *chmodels.PinterestBoard, parsedCounter *uint64, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker_id", id).Str("pool", "boards").Msg("Boards parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			var rawBoard kafkamodels.RawPinterestBoard
			if err := json.Unmarshal(msg.Value, &rawBoard); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "boardsParser").Str("stage", "unmarshal_raw_board").Str("key", string(msg.Key)).Msg("Failed to unmarshal raw board")
				continue
			}

			if rawBoard.BoardID == "" {
				log.Debug().Str("key", string(msg.Key)).Msg("Raw board has empty BoardID, skipping")
				continue
			}

			parsedBoard := parseRawBoard(&rawBoard)
			chBoard := conversions.ConvertPinterestBoard(parsedBoard)
			if chBoard != nil {
				select {
				case out <- chBoard:
					atomic.AddUint64(parsedCounter, 1)
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func pinsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, out chan<- *chmodels.PinterestPin, parsedCounter *uint64, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker_id", id).Str("pool", "pins").Msg("Pins parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			var rawPin kafkamodels.RawPinterestPin
			if err := json.Unmarshal(msg.Value, &rawPin); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "pinsParser").Str("stage", "unmarshal_raw_pin").Str("key", string(msg.Key)).Msg("Failed to unmarshal raw pin")
				continue
			}

			if rawPin.PinID == "" {
				log.Debug().Str("key", string(msg.Key)).Msg("Raw pin has empty PinID, skipping")
				continue
			}

			parsedPin := parseRawPin(&rawPin)
			chPin := conversions.ConvertPinterestPin(parsedPin)
			if chPin != nil {
				select {
				case out <- chPin:
					atomic.AddUint64(parsedCounter, 1)
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func pinInsightsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, out chan<- *chmodels.PinterestPinInsight, parsedCounter *uint64, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker_id", id).Str("pool", "pin_insights").Msg("Pin insights parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			var rawInsight kafkamodels.RawPinterestPinInsight
			if err := json.Unmarshal(msg.Value, &rawInsight); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "pinInsightsParser").Str("stage", "unmarshal_raw_pin_insight").Str("key", string(msg.Key)).Msg("Failed to unmarshal raw pin insight")
				continue
			}

			if rawInsight.PinID == "" {
				log.Debug().Str("key", string(msg.Key)).Msg("Raw pin insight has empty PinID, skipping")
				continue
			}

			parsedInsight := parseRawPinInsight(&rawInsight)
			chInsight := conversions.ConvertPinterestPinInsight(parsedInsight)
			if chInsight != nil {
				select {
				case out <- chInsight:
					atomic.AddUint64(parsedCounter, 1)
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func userInsightsParser(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan RawMessage, out chan<- *chmodels.PinterestUserInsight, parsedCounter *uint64, log *logger.Logger) {
	defer wg.Done()
	log.Info().Int("worker_id", id).Str("pool", "user_insights").Msg("User insights parser started")

	for {
		select {
		case <-ctx.Done():
			return
		case msg, ok := <-in:
			if !ok {
				return
			}

			var rawInsight kafkamodels.RawPinterestUserInsight
			if err := json.Unmarshal(msg.Value, &rawInsight); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "userInsightsParser").Str("stage", "unmarshal_raw_user_insight").Str("key", string(msg.Key)).Msg("Failed to unmarshal raw user insight")
				continue
			}

			if rawInsight.UserID == "" {
				log.Debug().Str("key", string(msg.Key)).Msg("Raw user insight has empty UserID, skipping")
				continue
			}

			parsedInsight := parseRawUserInsight(&rawInsight)
			chInsight := conversions.ConvertPinterestUserInsight(parsedInsight)
			if chInsight != nil {
				select {
				case out <- chInsight:
					atomic.AddUint64(parsedCounter, 1)
				case <-ctx.Done():
					return
				}
			}
		}
	}
}

func startBatchProcessors(ctx context.Context, batches *BatchCollectors, sink *conversions.ClickHouseSink, log *logger.Logger, wg *sync.WaitGroup, numProcessors int,
	insertedUsers, insertedBoards, insertedPins, insertedPinInsights, insertedUserInsights *uint64) {
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go usersBatchProcessor(ctx, wg, i, batches.users, sink, log, insertedUsers)
	}
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go boardsBatchProcessor(ctx, wg, i, batches.boards, sink, log, insertedBoards)
	}
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go pinsBatchProcessor(ctx, wg, i, batches.pins, sink, log, insertedPins)
	}
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go pinInsightsBatchProcessor(ctx, wg, i, batches.pinInsights, sink, log, insertedPinInsights)
	}
	for i := 0; i < numProcessors; i++ {
		wg.Add(1)
		go userInsightsBatchProcessor(ctx, wg, i, batches.userInsights, sink, log, insertedUserInsights)
	}
}

func usersBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.PinterestUser, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	log.Info().Int("processor_id", id).Str("batch_type", "users").Msg("Batch processor started")

	batch := make([]chmodels.PinterestUser, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flushBatch := func() {
		if len(batch) > 0 {
			if err := sink.BulkInsertPinterestUsers(ctx, batch); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "usersBatchProcessor").Str("stage", "bulk_insert_users").Int("processor_id", id).Int("batch_size", len(batch)).Msg("Failed to insert users batch")
			} else {
				atomic.AddUint64(insertedCounter, uint64(len(batch)))
			}
			batch = make([]chmodels.PinterestUser, 0, maxBatchSize)
		}
	}

	for {
		select {
		case <-ctx.Done():
			flushBatch()
			return

		case <-flushTimer.C:
			flushBatch()
			flushTimer.Reset(batchTimeout)

		case item, ok := <-in:
			if !ok {
				flushBatch()
				return
			}

			batch = append(batch, *item)
			if len(batch) >= maxBatchSize {
				flushBatch()
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

func boardsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.PinterestBoard, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	log.Info().Int("processor_id", id).Str("batch_type", "boards").Msg("Batch processor started")

	batch := make([]chmodels.PinterestBoard, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flushBatch := func() {
		if len(batch) > 0 {
			if err := sink.BulkInsertPinterestBoards(ctx, batch); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "boardsBatchProcessor").Str("stage", "bulk_insert_boards").Int("processor_id", id).Int("batch_size", len(batch)).Msg("Failed to insert boards batch")
			} else {
				atomic.AddUint64(insertedCounter, uint64(len(batch)))
			}
			batch = make([]chmodels.PinterestBoard, 0, maxBatchSize)
		}
	}

	for {
		select {
		case <-ctx.Done():
			flushBatch()
			return

		case <-flushTimer.C:
			flushBatch()
			flushTimer.Reset(batchTimeout)

		case item, ok := <-in:
			if !ok {
				flushBatch()
				return
			}

			batch = append(batch, *item)
			if len(batch) >= maxBatchSize {
				flushBatch()
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

func pinsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.PinterestPin, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	log.Info().Int("processor_id", id).Str("batch_type", "pins").Msg("Batch processor started")

	batch := make([]chmodels.PinterestPin, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flushBatch := func() {
		if len(batch) > 0 {
			if err := sink.BulkInsertPinterestPins(ctx, batch); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "pinsBatchProcessor").Str("stage", "bulk_insert_pins").Int("processor_id", id).Int("batch_size", len(batch)).Msg("Failed to insert pins batch")
			} else {
				atomic.AddUint64(insertedCounter, uint64(len(batch)))
			}
			batch = make([]chmodels.PinterestPin, 0, maxBatchSize)
		}
	}

	for {
		select {
		case <-ctx.Done():
			flushBatch()
			return

		case <-flushTimer.C:
			flushBatch()
			flushTimer.Reset(batchTimeout)

		case item, ok := <-in:
			if !ok {
				flushBatch()
				return
			}

			batch = append(batch, *item)
			if len(batch) >= maxBatchSize {
				flushBatch()
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

func pinInsightsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.PinterestPinInsight, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	log.Info().Int("processor_id", id).Str("batch_type", "pin_insights").Msg("Batch processor started")

	batch := make([]chmodels.PinterestPinInsight, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flushBatch := func() {
		if len(batch) > 0 {
			if err := sink.BulkInsertPinterestPinInsights(ctx, batch); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "pinInsightsBatchProcessor").Str("stage", "bulk_insert_pin_insights").Int("processor_id", id).Int("batch_size", len(batch)).Msg("Failed to insert pin insights batch")
			} else {
				atomic.AddUint64(insertedCounter, uint64(len(batch)))
			}
			batch = make([]chmodels.PinterestPinInsight, 0, maxBatchSize)
		}
	}

	for {
		select {
		case <-ctx.Done():
			flushBatch()
			return

		case <-flushTimer.C:
			flushBatch()
			flushTimer.Reset(batchTimeout)

		case item, ok := <-in:
			if !ok {
				flushBatch()
				return
			}

			batch = append(batch, *item)
			if len(batch) >= maxBatchSize {
				flushBatch()
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

func userInsightsBatchProcessor(ctx context.Context, wg *sync.WaitGroup, id int, in <-chan *chmodels.PinterestUserInsight, sink *conversions.ClickHouseSink, log *logger.Logger, insertedCounter *uint64) {
	defer wg.Done()
	log.Info().Int("processor_id", id).Str("batch_type", "user_insights").Msg("Batch processor started")

	batch := make([]chmodels.PinterestUserInsight, 0, maxBatchSize)
	flushTimer := time.NewTimer(batchTimeout)
	defer flushTimer.Stop()

	flushBatch := func() {
		if len(batch) > 0 {
			if err := sink.BulkInsertPinterestUserInsights(ctx, batch); err != nil {
				log.Error().Err(err).Str("error_message", err.Error()).Str("function", "userInsightsBatchProcessor").Str("stage", "bulk_insert_user_insights").Int("processor_id", id).Int("batch_size", len(batch)).Msg("Failed to insert user insights batch")
			} else {
				atomic.AddUint64(insertedCounter, uint64(len(batch)))
			}
			batch = make([]chmodels.PinterestUserInsight, 0, maxBatchSize)
		}
	}

	for {
		select {
		case <-ctx.Done():
			flushBatch()
			return

		case <-flushTimer.C:
			flushBatch()
			flushTimer.Reset(batchTimeout)

		case item, ok := <-in:
			if !ok {
				flushBatch()
				return
			}

			batch = append(batch, *item)
			if len(batch) >= maxBatchSize {
				flushBatch()
				flushTimer.Reset(batchTimeout)
			}
		}
	}
}

// Parsing helper functions - convert Raw models to Parsed models

func generateRecordID(id string, date time.Time) string {
	hash := md5.Sum([]byte(id + "_" + date.Format("20060102")))
	return hex.EncodeToString(hash[:])
}

func parseRawUser(raw *kafkamodels.RawPinterestUser) *kafkamodels.ParsedPinterestUser {
	now := time.Now().UTC()
	return &kafkamodels.ParsedPinterestUser{
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
}

func parseRawBoard(raw *kafkamodels.RawPinterestBoard) *kafkamodels.ParsedPinterestBoard {
	now := time.Now().UTC()
	return &kafkamodels.ParsedPinterestBoard{
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
}

func parseRawPin(raw *kafkamodels.RawPinterestPin) *kafkamodels.ParsedPinterestPin {
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

	return &kafkamodels.ParsedPinterestPin{
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
}

func parseRawPinInsight(raw *kafkamodels.RawPinterestPinInsight) *kafkamodels.ParsedPinterestPinInsight {
	now := time.Now().UTC()
	return &kafkamodels.ParsedPinterestPinInsight{
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
}

func parseRawUserInsight(raw *kafkamodels.RawPinterestUserInsight) *kafkamodels.ParsedPinterestUserInsight {
	now := time.Now().UTC()
	return &kafkamodels.ParsedPinterestUserInsight{
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
}
