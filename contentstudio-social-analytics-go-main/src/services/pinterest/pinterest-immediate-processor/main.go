package main

import (
	"context"
	"encoding/json"
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
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/pinterest/pinterest-immediate-processor/processor"
)

const (
	WorkerPoolSize    = 5
	WorkChannelBuffer = 50
)

type workMessage struct {
	ctx   context.Context
	value []byte
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Pinterest Immediate Processor")

	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(context.Background())
	log.Info().Msg("Connected to MongoDB")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)

	proc := processor.New(
		mongodb.NewUnifiedSocialRepository(mongoClient.Database(cfg.Mongo.Database), log.Logger),
		sink,
		notification.NewService(cfg.Email, log.Logger, cfg.Email.BackendURL),
		notification.NewPusherClient(cfg.Pusher, log.Logger),
		log,
		cfg,
	)

	consumer, err := kafka2.NewConsumer(cfg.Kafka, "pinterest-immediate-processor-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigc
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
		log.Info().Msg("Consuming immediate-work-order-pinterest topic")
		err = consumer.Consume(ctx, []string{"immediate-work-order-pinterest"}, func(ctx context.Context, _ string, _ []byte, value []byte) error {
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
	log.Info().Msg("Pinterest Immediate Processor stopped")
}

func worker(workerID int, workChan <-chan workMessage, proc *processor.Processor, log *logger.Logger) {
	log.Info().Int("worker_id", workerID).Msg("Worker started")
	defer log.Info().Int("worker_id", workerID).Msg("Worker stopped")

	for msg := range workChan {
		log.Debug().Int("worker_id", workerID).Msg("Processing work order")

		var wo processor.WorkOrder
		if err := json.Unmarshal(msg.value, &wo); err != nil {
			log.Error().Err(err).Msg("Failed to unmarshal work order")
			continue
		}

		log.Info().
			Str("account_id", wo.AccountID).
			Str("platform_identifier", wo.AccountID).
			Str("workspace_id", wo.WorkspaceID).
			Str("sync_type", wo.SyncType).
			Str("start_date", wo.StartDate).
			Str("end_date", wo.EndDate).
			Msg("Starting immediate processing for Pinterest account")

		startTime := time.Now()
		err := proc.ProcessAccount(msg.ctx, wo)
		duration := time.Since(startTime)

		if err != nil {
			log.Error().
				Err(err).
				Str("account_id", wo.AccountID).
				Str("workspace_id", wo.WorkspaceID).
				Dur("duration", duration).
				Msg("Failed to process Pinterest account")
		} else {
			log.Info().
				Str("account_id", wo.AccountID).
				Str("workspace_id", wo.WorkspaceID).
				Dur("duration", duration).
				Msg("Successfully processed Pinterest account")
		}
	}
}

func mustCreateProducer(cfg config.KafkaConfig, base zerolog.Logger) kafka2.Producer {
	producer, err := kafka2.NewProducer(cfg, base)
	if err != nil {
		base.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	return producer
}
