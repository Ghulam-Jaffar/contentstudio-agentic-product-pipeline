package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/facebook/facebook-immediate-processor/processor"
)

const (
	WorkerPoolSize    = 10
	WorkChannelBuffer = 100
)

type workMessage struct {
	ctx   context.Context
	value []byte
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Facebook Immediate Processor service")

	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	clientOpts := options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential)

	mongoClient, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(context.Background())
	log.Info().Msg("Connected to MongoDB")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)

	proc := processor.New(
		mongodb.NewUnifiedSocialRepository(mongoClient.Database(cfg.Mongo.Database), log.Logger),
		sink,
		mustCreateProducer(cfg.Kafka, log.Logger),
		notification.NewService(cfg.Email, log.Logger, cfg.Email.BackendURL),
		notification.NewPusherClient(cfg.Pusher, log.Logger),
		log,
		cfg,
	)

	consumer, err := kafka2.NewConsumer(cfg.Kafka, "immediate-processor-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()

	workChan := make(chan workMessage, WorkChannelBuffer)
	var wg sync.WaitGroup

	log.Info().
		Int("worker_count", WorkerPoolSize).
		Int("channel_buffer", WorkChannelBuffer).
		Msg("Starting worker pool")

	for i := 0; i < WorkerPoolSize; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			worker(workerID, workChan, proc, log)
		}(i)
	}

	go func() {
		topics := []string{"immediate-work-order-facebook"}
		log.Info().
			Strs("topics", topics).
			Str("consumer_group", "immediate-processor-group").
			Msg("Starting Kafka consumer")

		err = consumer.Consume(ctx, topics,
			func(ctx context.Context, topic string, key, value []byte) error {
				select {
				case workChan <- workMessage{ctx: ctx, value: value}:
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			})

		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Consumer error")
		}
		close(workChan)
	}()

	<-ctx.Done()
	log.Info().Msg("Shutting down, waiting for workers to finish")
	wg.Wait()
	log.Info().Msg("Facebook Immediate Processor service stopped")
}

func worker(workerID int, workChan <-chan workMessage, proc *processor.Processor, log *logger.Logger) {
	log.Info().Int("worker_id", workerID).Msg("Worker started")
	defer log.Info().Int("worker_id", workerID).Msg("Worker stopped")

	for msg := range workChan {
		log.Debug().Int("worker_id", workerID).Msg("Processing work order")

		var workOrder processor.WorkOrder
		if err := json.Unmarshal(msg.value, &workOrder); err != nil {
			log.Error().
				Err(err).
				Str("error_message", err.Error()).
				Str("function", "worker").
				Str("stage", "unmarshal_work_order").
				Msg("Failed to unmarshal work order")
			continue
		}

		log.Info().
			Str("account_id", workOrder.ID).
			Str("facebook_id", workOrder.AccountID).
			Str("platform_identifier", workOrder.AccountID).
			Str("workspace_id", workOrder.WorkspaceID).
			Str("sync_type", workOrder.SyncType).
			Str("start_date", workOrder.StartDate).
			Str("end_date", workOrder.EndDate).
			Msg("Starting immediate processing for Facebook account")

		startTime := time.Now()
		err := proc.ProcessAccount(msg.ctx, workOrder)
		duration := time.Since(startTime)

		if err != nil {
			log.Error().
				Err(err).
				Str("error_message", err.Error()).
				Str("account_id", workOrder.ID).
				Str("facebook_id", workOrder.AccountID).
				Str("workspace_id", workOrder.WorkspaceID).
				Dur("duration", duration).
				Str("function", "worker").
				Str("stage", "process_account").
				Msg("Failed to process Facebook account")
		} else {
			log.Info().
				Str("account_id", workOrder.ID).
				Str("facebook_id", workOrder.AccountID).
				Str("workspace_id", workOrder.WorkspaceID).
				Dur("duration", duration).
				Msg("Successfully processed Facebook account")
		}
	}
}

func mustCreateProducer(cfg config.KafkaConfig, logger zerolog.Logger) kafka2.Producer {
	producer, err := kafka2.NewProducer(cfg, logger)
	if err != nil {
		panic(fmt.Sprintf("Failed to create Kafka producer: %v", err))
	}
	return producer
}
