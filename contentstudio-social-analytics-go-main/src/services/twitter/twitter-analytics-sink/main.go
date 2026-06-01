package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Twitter Analytics Sink (merged parser+sink)")

	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	if err := sink.Health(); err != nil {
		log.Warn().Err(err).Msg("ClickHouse health check failed - continuing anyway")
	}

	postsConsumer, err := kafka2.NewConsumer(cfg.Kafka, "twitter-analytics-sink-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create posts consumer")
	}
	defer postsConsumer.Close()

	insightsConsumer, err := kafka2.NewConsumer(cfg.Kafka, "twitter-analytics-sink-group", log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create insights consumer")
	}
	defer insightsConsumer.Close()

	ctx, cancel := context.WithCancel(context.Background())

	// Prepare dependencies
	deps := &ServiceDependencies{
		Sink:             sink,
		PostsConsumer:    postsConsumer,
		InsightsConsumer: insightsConsumer,
		Logger:           log,
	}

	// Prepare configuration
	serviceCfg := DefaultServiceConfig()

	// Start service in goroutine
	serviceDone := make(chan error, 1)
	go func() {
		serviceDone <- RunService(ctx, deps, serviceCfg)
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Int("posts_parser_workers", serviceCfg.PostsParserWorkers).
		Int("insights_parser_workers", serviceCfg.InsightsParserWorkers).
		Int("batch_processors_per_type", serviceCfg.BatchProcessorsPerType).
		Int("max_batch_size", serviceCfg.MaxBatchSize).
		Dur("batch_timeout", serviceCfg.BatchTimeout).
		Msg("Twitter Analytics Sink started successfully")

	select {
	case <-sigChan:
		log.Info().Msg("Shutdown signal received, stopping...")
	case err := <-serviceDone:
		if err != nil {
			log.Error().Err(err).Str("error_message", err.Error()).Str("function", "main").Str("stage", "service_run").Msg("Service error")
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

	log.Info().Msg("Twitter Analytics Sink stopped")
}
