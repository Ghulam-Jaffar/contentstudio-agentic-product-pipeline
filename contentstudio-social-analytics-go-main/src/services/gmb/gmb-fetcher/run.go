package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"golang.org/x/sync/semaphore"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// FetcherConfig holds service configuration
type FetcherConfig struct {
	MaxWorkers            int
	QueueSize             int
	MaxConcurrentAccounts int
	DecryptionKey         string
}

// DefaultFetcherConfig returns default configuration
func DefaultFetcherConfig() FetcherConfig {
	return FetcherConfig{
		MaxWorkers:            maxWorkers,
		QueueSize:             workChanSize,
		MaxConcurrentAccounts: maxConcurrentAccounts,
	}
}

// FetcherDependencies holds external dependencies
type FetcherDependencies struct {
	Producer  kafka2.Producer
	Consumer  kafka2.Consumer
	MongoRepo mongodb.UnifiedSocialRepository
	GMBClient *social.GMBClient
	Log       *logger.Logger
}

// FetcherMetrics holds service metrics
type FetcherMetrics struct {
	AccountsDispatched uint64
	AccountsProcessed  uint64
	ProcessingErrors   uint64
}

// RunService starts the GMB fetcher service as a long-running Kafka consumer.
// It consumes work orders from the work-order-gmb topic and dispatches one goroutine per account.
// ack is not used (plain Consume); goroutines are bounded by a semaphore.
func RunService(ctx context.Context, deps *FetcherDependencies, cfg FetcherConfig) error {
	log := deps.Log

	metrics := &FetcherMetrics{}

	maxConc := cfg.MaxConcurrentAccounts
	if maxConc <= 0 {
		maxConc = maxConcurrentAccounts
	}
	accountSem := semaphore.NewWeighted(int64(maxConc))
	var dispatchWg sync.WaitGroup

	log.Info().
		Int("max_concurrent_accounts", maxConc).
		Msg("Starting GMB Fetcher Service")

	go func() {
		_ = deps.Consumer.Consume(ctx, []string{workOrderTopic}, func(ctx context.Context, topic string, key, value []byte) error {
			var batch kafkamodels.GMBBatchWorkOrder
			if err := json.Unmarshal(value, &batch); err != nil {
				log.Error().Err(err).Str("topic", topic).Msg("failed to unmarshal GMB batch work order")
				return nil
			}
			total := len(batch.Accounts)
			log.Info().
				Str("batch_id", batch.BatchID).
				Str("sync_type", batch.SyncType).
				Int("account_count", total).
				Msg("received GMB batch work order, dispatching goroutines")

			var batchWg sync.WaitGroup
			var batchProcessed, batchFailed int64

			for _, wo := range batch.Accounts {
				w := wo
				dispatchWg.Add(1)
				batchWg.Add(1)
				go func() {
					defer dispatchWg.Done()
					defer batchWg.Done()
					if err := accountSem.Acquire(ctx, 1); err != nil {
						atomic.AddInt64(&batchFailed, 1)
						return
					}
					defer accountSem.Release(1)

					atomic.AddUint64(&metrics.AccountsDispatched, 1)
					if err := HandleWorkOrder(ctx, w, deps.GMBClient, deps.Producer, deps.MongoRepo, cfg.DecryptionKey, log); err != nil {
						log.Error().Err(err).Str("location_id", w.LocationID).Msg("work order failed")
						atomic.AddUint64(&metrics.ProcessingErrors, 1)
						atomic.AddInt64(&batchFailed, 1)
					} else {
						atomic.AddUint64(&metrics.AccountsProcessed, 1)
						atomic.AddInt64(&batchProcessed, 1)
					}
				}()
			}

			batchID := batch.BatchID
			syncType := batch.SyncType
			go func() {
				batchWg.Wait()
				log.Info().
					Str("batch_id", batchID).
					Str("sync_type", syncType).
					Int("total", total).
					Int64("processed", atomic.LoadInt64(&batchProcessed)).
					Int64("failed", atomic.LoadInt64(&batchFailed)).
					Msg("Batch processing complete")
			}()
			return nil
		})
	}()

	<-ctx.Done()
	dispatchWg.Wait()

	log.Info().
		Uint64("accounts_dispatched", atomic.LoadUint64(&metrics.AccountsDispatched)).
		Uint64("accounts_processed", atomic.LoadUint64(&metrics.AccountsProcessed)).
		Uint64("processing_errors", atomic.LoadUint64(&metrics.ProcessingErrors)).
		Msg("GMB Fetcher Service stopped")

	return nil
}

// GetMetrics returns current service metrics
func GetMetrics(metrics *FetcherMetrics) map[string]uint64 {
	return map[string]uint64{
		"accounts_dispatched": atomic.LoadUint64(&metrics.AccountsDispatched),
		"accounts_processed":  atomic.LoadUint64(&metrics.AccountsProcessed),
		"processing_errors":   atomic.LoadUint64(&metrics.ProcessingErrors),
	}
}
