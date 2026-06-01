package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)
	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Twitter Fetcher service")

	consumer, err := kafka2.NewConsumer(cfg.Kafka, "twitter-fetcher-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka consumer")
	}
	defer consumer.Close()

	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka producer")
	}
	defer producer.Close()

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

	mongoRepo := mongodb.NewUnifiedSocialRepository(mongoClient.Database(cfg.Mongo.Database), log.Logger)

	twClient := social.NewTwitterClient(cfg.Twitter.ConsumerKey, cfg.Twitter.ConsumerSecret)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Prepare dependencies
	deps := &FetcherDependencies{
		Consumer:      consumer,
		Producer:      producer,
		MongoRepo:     mongoRepo,
		TwitterClient: twClient,
		Log:           log,
	}

	// Prepare configuration
	serviceCfg := DefaultFetcherConfig()
	serviceCfg.DecryptionKey = cfg.DecryptionKey

	// Start service in goroutine
	serviceDone := make(chan error, 1)
	go func() {
		serviceDone <- RunService(ctx, deps, serviceCfg)
	}()

	select {
	case <-sigChan:
		log.Info().Msg("Shutdown signal received")
	case err := <-serviceDone:
		if err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "run_service").Msg("Service error")
		}
	}

	cancel()

	// Wait for service to complete with timeout
	select {
	case <-serviceDone:
		log.Info().Msg("Service stopped gracefully")
	case <-time.After(10 * time.Second):
		log.Warn().Msg("Service shutdown timeout")
	}

	log.Info().Msg("Twitter Fetcher stopped")
}
