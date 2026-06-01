package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"strings"
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
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/youtube/youtube-immediate-processor/processor"
)

const (
	WorkerPoolSize    = 10
	WorkChannelBuffer = 100
	ConsumerGroup     = "youtube-immediate-processor-group"
	Topic             = "immediate-work-order-youtube"
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
	log.Info().Msg("Starting YouTube Immediate Processor")

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

	consumer, err := kafka2.NewConsumer(cfg.Kafka, ConsumerGroup, log.Logger)
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
		log.Info().Str("topic", Topic).Msg("Consuming YouTube work orders")
		err = consumer.Consume(ctx, []string{Topic}, func(ctx context.Context, _ string, _ []byte, value []byte) error {
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
	log.Info().Msg("YouTube Immediate Processor stopped")
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
			Str("channel_id", wo.ChannelID).
			Str("platform_identifier", wo.ChannelID).
			Str("workspace_id", wo.WorkspaceID).
			Str("sync_type", wo.SyncType).
			Str("start_date", wo.StartDate).
			Str("end_date", wo.EndDate).
			Msg("Starting immediate processing for YouTube account")

		startTime := time.Now()
		err := proc.ProcessAccount(msg.ctx, wo)
		duration := time.Since(startTime)

		if err != nil {
			if isYouTubeExpectedError(err) {
				// Expected auth/permission errors - just skip this account, no logging needed
			} else {
				log.Error().
					Err(err).
					Str("error_message", err.Error()).
					Str("channel_id", wo.ChannelID).
					Str("workspace_id", wo.WorkspaceID).
					Dur("duration", duration).
					Str("function", "worker").
					Str("stage", "process_account").
					Msg("Failed to process YouTube account")
			}
		} else {
			log.Info().
				Str("channel_id", wo.ChannelID).
				Str("workspace_id", wo.WorkspaceID).
				Dur("duration", duration).
				Msg("Successfully processed YouTube account")
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

// isYouTubeExpectedError checks if an error is an expected/operational error (auth, permission, etc)
// These errors do not warrant logging or Sentry alerting as they're expected when:
// - Access tokens expire
// - User revokes permissions
// - Account credentials become invalid
func isYouTubeExpectedError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific sentinel errors
	if errors.Is(err, processor.ErrUnauthorized) {
		return true
	}

	// Check error message for common auth/permission patterns
	errStr := strings.ToLower(err.Error())
	authPatterns := []string{
		"unauthorized",
		"invalid.*credential",
		"invalid.*token",
		"access token",
		"token.*expired",
		"expired token",
		"permission",
		"auth",
		"401",
		"403",
		"forbidden",
		"unauthenticated",
		"revoked",
	}

	for _, pattern := range authPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
