package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/tiktok/tiktok-immediate-processor/processor"
)

// ImmediateProcessorConfig holds service configuration
type ImmediateProcessorConfig struct {
	MaxWorkers   int
	JobQueueSize int
}

// DefaultImmediateProcessorConfig returns default configuration
func DefaultImmediateProcessorConfig() ImmediateProcessorConfig {
	return ImmediateProcessorConfig{
		MaxWorkers:   5,   // maxImmediateWorkers from main.go
		JobQueueSize: 100, // jobQueueSize from main.go
	}
}

// ImmediateProcessorDependencies holds external dependencies
type ImmediateProcessorDependencies struct {
	Consumer  kafka.Consumer
	Processor processor.ProcessorInterface
	Log       *logger.Logger
}

// ImmediateProcessorMetrics holds service metrics
type ImmediateProcessorMetrics struct {
	ReceivedJobs  uint64
	ProcessedJobs uint64
	SkippedJobs   uint64
	FailedJobs    uint64
}

// ImmediateProcessorService represents the immediate processor service
type ImmediateProcessorService struct {
	config   ImmediateProcessorConfig
	deps     ImmediateProcessorDependencies
	metrics  ImmediateProcessorMetrics
	inFlight sync.Map
}

// NewImmediateProcessorService creates a new immediate processor service
func NewImmediateProcessorService(cfg ImmediateProcessorConfig, deps ImmediateProcessorDependencies) *ImmediateProcessorService {
	return &ImmediateProcessorService{
		config: cfg,
		deps:   deps,
	}
}

// Run starts the immediate processor service
func (s *ImmediateProcessorService) Run(ctx context.Context) error {
	log := s.deps.Log

	jobQueue := make(chan processor.ImmediateWorkOrder, s.config.JobQueueSize)

	var wg sync.WaitGroup

	for i := 0; i < s.config.MaxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.runWorker(ctx, workerID, jobQueue)
		}(i)
	}

	// Consumer
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.deps.Consumer.Consume(ctx, []string{"immediate-work-order-tiktok"}, func(ctx context.Context, topic string, key, value []byte) error {
			var wo processor.ImmediateWorkOrder
			if err := json.Unmarshal(value, &wo); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal work order")
				return nil
			}

			atomic.AddUint64(&s.metrics.ReceivedJobs, 1)

			select {
			case jobQueue <- wo:
			case <-ctx.Done():
				return ctx.Err()
			default:
				atomic.AddUint64(&s.metrics.SkippedJobs, 1)
				log.Warn().Str("account_id", wo.TikTokID).Msg("Job queue full, skipping")
			}

			return nil
		})
	}()

	wg.Wait()

	log.Info().
		Uint64("received_jobs", atomic.LoadUint64(&s.metrics.ReceivedJobs)).
		Uint64("processed_jobs", atomic.LoadUint64(&s.metrics.ProcessedJobs)).
		Uint64("failed_jobs", atomic.LoadUint64(&s.metrics.FailedJobs)).
		Uint64("skipped_jobs", atomic.LoadUint64(&s.metrics.SkippedJobs)).
		Msg("TikTok Immediate Processor stopped")

	return nil
}

// runWorker processes work orders from the queue
func (s *ImmediateProcessorService) runWorker(ctx context.Context, workerID int, jobQueue <-chan processor.ImmediateWorkOrder) {
	log := s.deps.Log.Logger.With().Int("worker_id", workerID).Logger()

	for {
		select {
		case wo, ok := <-jobQueue:
			if !ok {
				return
			}

			start := time.Now()

			if err := s.deps.Processor.ProcessAccount(wo); err != nil {
				atomic.AddUint64(&s.metrics.FailedJobs, 1)
				log.Error().
					Err(err).
					Str("account_id", wo.TikTokID).
					Dur("duration", time.Since(start)).
					Msg("Failed to process account")
			} else {
				atomic.AddUint64(&s.metrics.ProcessedJobs, 1)
				log.Info().
					Str("account_id", wo.TikTokID).
					Dur("duration", time.Since(start)).
					Msg("Account processed")
			}

		case <-ctx.Done():
			return
		}
	}
}

// GetMetrics returns current service metrics
func (s *ImmediateProcessorService) GetMetrics() map[string]uint64 {
	return map[string]uint64{
		"received_jobs":  atomic.LoadUint64(&s.metrics.ReceivedJobs),
		"processed_jobs": atomic.LoadUint64(&s.metrics.ProcessedJobs),
		"failed_jobs":    atomic.LoadUint64(&s.metrics.FailedJobs),
		"skipped_jobs":   atomic.LoadUint64(&s.metrics.SkippedJobs),
	}
}
