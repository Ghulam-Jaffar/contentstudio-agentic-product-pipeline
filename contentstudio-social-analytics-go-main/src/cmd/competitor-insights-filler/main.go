package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	ch "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

const (
	AppName = "competitor-gap-filler"

	PlatformFacebook  = "facebook"
	PlatformInstagram = "instagram"
)

func main() {
	// --------------------
	// Parse CLI arguments
	// --------------------
	platform := flag.String(
		"platform",
		PlatformFacebook,
		"Target platform: facebook | instagram",
	)

	workers := flag.Int(
		"workers",
		8,
		"Number of concurrent workers",
	)

	lookbackDays := flag.Int(
		"lookbackDays",
		10,
		"Number of days to look back for records",
	)

	flag.Parse()

	// --------------------
	// Load configuration
	// --------------------
	cfg, err := config.LoadConfig()
	if err != nil {
		logFatal("failed to load configuration", err)
	}

	telemetry.ConfigureSentry(cfg)

	appLogger := logger.New(cfg.LogLevel)

	op := appLogger.
		Operation("competitor_gap_filler").
		WithSentryTags(map[string]string{
			"app":      AppName,
			"platform": *platform,
		}).
		WithSentryExtras(map[string]interface{}{
			"workers":       *workers,
			"lookback_days": *lookbackDays,
		})

	defer func() {
		op.Complete(nil, "")
		logger.FlushSentry(5 * time.Second)
	}()

	appLogger.Info().
		Str("app", AppName).
		Str("platform", *platform).
		Int("workers", *workers).
		Int("lookback_days", *lookbackDays).
		Msg("Starting gap filler job")

	// --------------------
	// Validate CLI input
	// --------------------
	if !isValidPlatform(*platform) {
		appLogger.Fatal().
			Str("platform", *platform).
			Msg("Invalid platform value. Must be 'facebook' or 'instagram'")
	}

	// --------------------
	// Initialize ClickHouse
	// --------------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	chConn, err := initClickHouse(cfg)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to initialize ClickHouse")
	}
	defer chConn.Close()

	appLogger.Info().Msg("ClickHouse connection established")

	// --------------------
	// Setup graceful shutdown
	// --------------------
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		appLogger.Warn().Msg("Received shutdown signal, canceling job...")
		cancel()
	}()

	// --------------------
	// Run gap fill job
	// --------------------
	jobConfig := jobConfig{
		Platform:     *platform,
		Workers:      *workers,
		LookbackDays: *lookbackDays,
		From:         time.Now().AddDate(0, 0, -*lookbackDays),
		To:           time.Now(),
	}

	err = RunGapFillJob(ctx, chConn, appLogger, jobConfig)
	if err != nil {
		if ctx.Err() != nil {
			appLogger.Warn().Msg("Job cancelled by user")
			os.Exit(0)
		}
		appLogger.Fatal().Err(err).Msg("Gap fill job failed")
	}

	appLogger.Info().Msg("Gap fill job completed successfully")
}

// isValidPlatform checks if the platform is supported
func isValidPlatform(platform string) bool {
	return platform == PlatformFacebook || platform == PlatformInstagram
}

// initClickHouse initializes and returns a ClickHouse connection
func initClickHouse(cfg *config.Config) (ch.Conn, error) {

	conn, err := ch.Open(&ch.Options{
		Addr: []string{fmt.Sprintf("%s:%d", cfg.ClickHouse.Host, cfg.ClickHouse.Port)},
		Auth: ch.Auth{
			Database: cfg.ClickHouse.Database,
			Username: cfg.ClickHouse.Username,
			Password: cfg.ClickHouse.Password,
		},
	})
	if err != nil {
		return nil, err
	}
	if err := conn.Ping(context.Background()); err != nil {
		return nil, err
	}
	return conn, nil
}

// logFatal logs a fatal error and exits
func logFatal(msg string, err error) {
	log := logger.New("error")

	op := log.
		Operation("fatal_startup_error").
		WithSentryExtras(map[string]interface{}{
			"message": msg,
		})

	op.Complete(err, "")
	log.Fatal().Err(err).Msg(msg)

	logger.FlushSentry(5 * time.Second)
}

// RunGapFillJob orchestrates the entire gap filling process
func RunGapFillJob(
	ctx context.Context,
	conn ch.Conn,
	log *logger.Logger,
	cfg jobConfig,
) error {
	op := log.
		Operation("run_gap_fill_job").
		WithSentryTags(map[string]string{
			"platform": cfg.Platform,
		}).
		WithSentryExtras(map[string]interface{}{
			"from":     cfg.From,
			"to":       cfg.To,
			"workers":  cfg.Workers,
			"lookback": cfg.LookbackDays,
		})

	defer func() {
		op.Complete(nil, "")
	}()

	log.Info().
		Str("platform", cfg.Platform).
		Time("from", cfg.From).
		Time("to", cfg.To).
		Msg("Starting gap fill process")

	// Statistics tracking
	stats := &gapStats{}

	// Channels for pipeline
	competitorCh := make(chan competitorKey, cfg.Workers*2)
	resultCh := make(chan competitorGapResult, cfg.Workers*2)
	insertCh := make(chan interface{}, 1000)

	var wg sync.WaitGroup
	var collectorWg sync.WaitGroup
	var inserterWg sync.WaitGroup

	// Start the inserter goroutine
	inserterWg.Add(1)
	go func() {
		defer inserterWg.Done()
		runInserter(ctx, cfg, conn, log, insertCh, stats)
	}()

	// Start worker pool
	for i := 0; i < cfg.Workers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			runWorker(ctx, workerID, cfg, conn, log, competitorCh, resultCh)
		}(i)
	}

	// Start result collector
	collectorWg.Add(1)
	go func() {
		defer collectorWg.Done()
		collectResults(ctx, log, resultCh, insertCh, stats)
	}()

	// Produce competitors
	log.Info().Msg("Fetching competitors to process")
	err := produceCompetitors(ctx, cfg, conn, log, competitorCh)
	close(competitorCh)

	if err != nil {
		log.Error().Err(err).Msg("Failed to produce competitors")
		close(resultCh)
		collectorWg.Wait()
		close(insertCh)
		inserterWg.Wait()
		return err
	}

	// Wait for all workers to finish
	wg.Wait()
	close(resultCh)

	log.Info().Msg("All workers completed, waiting for collector")

	// Wait for result collector to finish
	collectorWg.Wait()

	log.Info().Msg("All workers completed, waiting for inserter")

	// Wait for inserter to finish
	close(insertCh)
	inserterWg.Wait()

	// Log final statistics
	log.Info().
		Int("competitors_processed", int(stats.CompetitorsProcessed)).
		Int("gap_segments", int(stats.GapSegments)).
		Int("records_generated", int(stats.RecordsGenerated)).
		Int("records_inserted", int(stats.RecordsInserted)).
		Int("errors", int(stats.Errors)).
		Msg("Gap fill job completed")

	if stats.Errors > 0 {
		err := fmt.Errorf("RunGapFillJob: job completed with %d errors", stats.Errors)
		op.Complete(err, "partial_failure")
		return err
	}

	return nil
}
