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
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/instagram/instagram-immediate-processor/processor"
)

const (
	maxImmediateWorkers = 8
	jobQueueSize        = 100
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Instagram Immediate Processor")

	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential))
	if err != nil {
		log.Fatal().Err(err).Msg("MongoDB connection error")
	}
	defer mongoClient.Disconnect(context.Background())

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

	consumer, err := kafka2.NewConsumer(cfg.Kafka, "instagram-immediate-processor-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Kafka consumer error")
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

	jobQueue := make(chan processor.WorkOrder, jobQueueSize)
	var wg sync.WaitGroup
	var inFlight sync.Map

	for i := 0; i < maxImmediateWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			worker(ctx, workerID, jobQueue, proc, log, &inFlight)
		}(i)
	}
	log.Info().Int("workers", maxImmediateWorkers).Int("queue_size", jobQueueSize).Msg("Started worker pool")

	log.Info().Str("topic", "immediate-work-order-instagram").Msg("Starting Kafka consumer")
	err = consumer.Consume(ctx, []string{"immediate-work-order-instagram"}, func(ctx context.Context, _ string, _ []byte, value []byte) error {
		var wo processor.WorkOrder
		if err := json.Unmarshal(value, &wo); err != nil {
			log.Error().
				Err(err).
				Str("error_message", err.Error()).
				Str("function", "worker").
				Str("stage", "unmarshal_work_order").
				Msg("Failed to unmarshal work order")
			return nil
		}

		select {
		case jobQueue <- wo:
			log.Debug().Str("instagram_id", wo.AccountID).Str("workspace_id", wo.WorkspaceID).Msg("Queued work order")
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})
	if err != nil && err != context.Canceled {
		log.Error().Err(err).Msg("Consumer error")
	}

	close(jobQueue)
	wg.Wait()

	log.Info().Msg("Instagram Immediate Processor stopped")
}

func worker(ctx context.Context, workerID int, jobs <-chan processor.WorkOrder, proc *processor.Processor, log *logger.Logger, inFlight *sync.Map) {
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "immediate").Logger()
	workerLog.Info().Msg("Worker started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Worker stopping (context canceled)")
			return
		case wo, ok := <-jobs:
			if !ok {
				workerLog.Info().Msg("Worker stopping (queue closed)")
				return
			}

			if existingStart, loaded := inFlight.LoadOrStore(wo.AccountID, time.Now()); loaded {
				workerLog.Warn().
					Str("instagram_id", wo.AccountID).
					Str("workspace_id", wo.WorkspaceID).
					Time("started_at", existingStart.(time.Time)).
					Msg("Account already being processed, skipping duplicate request")
				continue
			}

			startTime := time.Now()
			workerLog.Info().
				Str("instagram_id", wo.AccountID).
				Str("platform_identifier", wo.AccountID).
				Str("workspace_id", wo.WorkspaceID).
				Str("sync_type", wo.SyncType).
				Str("start_date", wo.StartDate).
				Str("end_date", wo.EndDate).
				Msg("Processing account")

			err := proc.ProcessAccount(ctx, wo)

			inFlight.Delete(wo.AccountID)

			if err != nil {
				workerLog.Error().
					Err(err).
					Str("error_message", err.Error()).
					Str("instagram_id", wo.AccountID).
					Str("workspace_id", wo.WorkspaceID).
					Dur("duration", time.Since(startTime)).
					Str("function", "worker").
					Str("stage", "process_account").
					Msg("Failed to process account")
			} else {
				workerLog.Info().
					Str("instagram_id", wo.AccountID).
					Dur("duration", time.Since(startTime)).
					Msg("Successfully processed account")
			}
		}
	}
}

func mustCreateProducer(cfg config.KafkaConfig, base zerolog.Logger) kafka2.Producer {
	producer, err := kafka2.NewProducer(cfg, base)
	if err != nil {
		panic("failed to create kafka producer: " + err.Error())
	}
	return producer
}
