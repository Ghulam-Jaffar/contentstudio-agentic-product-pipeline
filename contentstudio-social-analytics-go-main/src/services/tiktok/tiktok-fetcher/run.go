package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// FetcherConfig holds service configuration
type FetcherConfig struct {
	MaxVideoWorkers         int
	MaxInsightsWorkers      int
	VideoQueueSize          int
	InsightsQueueSize       int
	TimestampUpdateChanSize int
	DecryptionKey           string
	ClientKey               string
	ClientSecret            string
}

// DefaultFetcherConfig returns default configuration
func DefaultFetcherConfig() FetcherConfig {
	return FetcherConfig{
		MaxVideoWorkers:         maxWorkers,
		MaxInsightsWorkers:      maxWorkers,
		VideoQueueSize:          workChanSize,
		InsightsQueueSize:       workChanSize,
		TimestampUpdateChanSize: 100,
	}
}

// FetcherDependencies holds external dependencies
type FetcherDependencies struct {
	VideoConsumer    kafka2.Consumer
	InsightsConsumer kafka2.Consumer
	Producer         kafka2.Producer
	MongoRepo        mongodb.UnifiedSocialRepository
	TikTokClient     *social.TikTokClient
	Log              *logger.Logger
}

// FetcherMetrics holds service metrics
type FetcherMetrics struct {
	VideoBatchesReceived      uint64
	InsightsBatchesReceived   uint64
	VideoAccountsProcessed    uint64
	InsightsAccountsProcessed uint64
	TotalBatchesProduced      uint64
	ProducerErrors            uint64
}

// RunService starts the TikTok fetcher service with the given dependencies
func RunService(ctx context.Context, deps *FetcherDependencies, cfg FetcherConfig) error {
	log := deps.Log

	metrics := &FetcherMetrics{}

	log.Info().
		Int("max_video_workers", cfg.MaxVideoWorkers).
		Int("max_insights_workers", cfg.MaxInsightsWorkers).
		Msg("Starting TikTok Fetcher Service")

	var wg sync.WaitGroup

	// Start video batch processor
	wg.Add(1)
	go processVideoBatches(ctx, deps, &wg, metrics, log)

	// Start insights batch processor
	wg.Add(1)
	go processInsightsBatches(ctx, deps, &wg, metrics, log)

	// Wait for all processors to complete
	wg.Wait()

	log.Info().
		Uint64("total_batches_produced", atomic.LoadUint64(&metrics.TotalBatchesProduced)).
		Uint64("producer_errors", atomic.LoadUint64(&metrics.ProducerErrors)).
		Msg("TikTok Fetcher Service stopped")

	return nil
}

// processVideoBatches processes video batch work orders
func processVideoBatches(ctx context.Context, deps *FetcherDependencies, wg *sync.WaitGroup, metrics *FetcherMetrics, log *logger.Logger) {
	defer wg.Done()

	batchLog := log.Logger.With().Str("processor", "video_batches").Logger()

	err := deps.VideoConsumer.Consume(ctx, []string{workOrdersTopic}, func(ctx context.Context, topic string, key, value []byte) error {
		var batch kafkamodels.TikTokBatchWorkOrder
		if err := json.Unmarshal(value, &batch); err != nil {
			batchLog.Error().Err(err).Msg("Failed to unmarshal video batch")
			return nil
		}

		atomic.AddUint64(&metrics.VideoBatchesReceived, 1)
		atomic.AddUint64(&metrics.VideoAccountsProcessed, uint64(len(batch.Accounts)))

		// Process batch (implementation in main.go)
		// For now, just track metrics
		atomic.AddUint64(&metrics.TotalBatchesProduced, 1)

		return nil
	})

	if err != nil && err != context.Canceled {
		batchLog.Error().Err(err).Msg("Video batch consumer error")
	}
}

// processInsightsBatches processes insights batch work orders
func processInsightsBatches(ctx context.Context, deps *FetcherDependencies, wg *sync.WaitGroup, metrics *FetcherMetrics, log *logger.Logger) {
	defer wg.Done()

	batchLog := log.Logger.With().Str("processor", "insights_batches").Logger()

	err := deps.InsightsConsumer.Consume(ctx, []string{workOrdersTopic}, func(ctx context.Context, topic string, key, value []byte) error {
		var batch kafkamodels.TikTokBatchWorkOrder
		if err := json.Unmarshal(value, &batch); err != nil {
			batchLog.Error().Err(err).Msg("Failed to unmarshal insights batch")
			return nil
		}

		atomic.AddUint64(&metrics.InsightsBatchesReceived, 1)
		atomic.AddUint64(&metrics.InsightsAccountsProcessed, uint64(len(batch.Accounts)))

		// Process batch (implementation in main.go)
		// For now, just track metrics
		atomic.AddUint64(&metrics.TotalBatchesProduced, 1)

		return nil
	})

	if err != nil && err != context.Canceled {
		batchLog.Error().Err(err).Msg("Insights batch consumer error")
	}
}

// GetMetrics returns current service metrics
func GetMetrics(metrics *FetcherMetrics) map[string]uint64 {
	return map[string]uint64{
		"video_batches_received":      atomic.LoadUint64(&metrics.VideoBatchesReceived),
		"insights_batches_received":   atomic.LoadUint64(&metrics.InsightsBatchesReceived),
		"video_accounts_processed":    atomic.LoadUint64(&metrics.VideoAccountsProcessed),
		"insights_accounts_processed": atomic.LoadUint64(&metrics.InsightsAccountsProcessed),
		"total_batches_produced":      atomic.LoadUint64(&metrics.TotalBatchesProduced),
		"producer_errors":             atomic.LoadUint64(&metrics.ProducerErrors),
	}
}
