package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/instagram/instagram-immediate-processor/processor"
)

// ImmediateProcessorConfig holds service configuration
type ImmediateProcessorConfig struct {
	MaxWorkers   int
	JobQueueSize int
}

// DefaultImmediateProcessorConfig returns default configuration
func DefaultImmediateProcessorConfig() ImmediateProcessorConfig {
	return ImmediateProcessorConfig{
		MaxWorkers:   maxImmediateWorkers,
		JobQueueSize: jobQueueSize,
	}
}

// ImmediateProcessorDependencies holds external dependencies
type ImmediateProcessorDependencies struct {
	Consumer  kafka.Consumer
	Processor *processor.Processor
	Log       *logger.Logger
}

// ImmediateProcessorMetrics holds service metrics
type ImmediateProcessorMetrics struct {
	ReceivedJobs   uint64
	ProcessedJobs  uint64
	SkippedJobs    uint64
	FailedJobs     uint64
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

	jobQueue := make(chan processor.WorkOrder, s.config.JobQueueSize)

	var wg sync.WaitGroup

	for i := 0; i < s.config.MaxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.runWorker(ctx, workerID, jobQueue)
		}(i)
	}

	// Consumer
	go func() {
		s.deps.Consumer.Consume(ctx, []string{"immediate-work-order-instagram"}, func(ctx context.Context, topic string, key, value []byte) error {
			var wo processor.WorkOrder
			if err := json.Unmarshal(value, &wo); err != nil {
				log.Error().Err(err).Msg("Failed to unmarshal work order")
				return nil
			}

			atomic.AddUint64(&s.metrics.ReceivedJobs, 1)

			select {
			case jobQueue <- wo:
			case <-ctx.Done():
				return ctx.Err()
			}
			return nil
		})
	}()

	<-ctx.Done()

	close(jobQueue)
	wg.Wait()

	return nil
}

func (s *ImmediateProcessorService) runWorker(ctx context.Context, workerID int, jobs <-chan processor.WorkOrder) {
	log := s.deps.Log
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "immediate").Logger()

	for {
		select {
		case <-ctx.Done():
			return
		case wo, ok := <-jobs:
			if !ok {
				return
			}

			if existingStart, loaded := s.inFlight.LoadOrStore(wo.AccountID, time.Now()); loaded {
				workerLog.Warn().
					Str("instagram_id", wo.AccountID).
					Time("started_at", existingStart.(time.Time)).
					Msg("Account already being processed, skipping")
				atomic.AddUint64(&s.metrics.SkippedJobs, 1)
				continue
			}

			startTime := time.Now()
			err := s.deps.Processor.ProcessAccount(ctx, wo)
			s.inFlight.Delete(wo.AccountID)

			if err != nil {
				atomic.AddUint64(&s.metrics.FailedJobs, 1)
				workerLog.Error().
					Err(err).
					Str("instagram_id", wo.AccountID).
					Dur("duration", time.Since(startTime)).
					Msg("Failed to process account")
			} else {
				atomic.AddUint64(&s.metrics.ProcessedJobs, 1)
				workerLog.Info().
					Str("instagram_id", wo.AccountID).
					Dur("duration", time.Since(startTime)).
					Msg("Successfully processed account")
			}
		}
	}
}

// GetMetrics returns current service metrics
func (s *ImmediateProcessorService) GetMetrics() ImmediateProcessorMetrics {
	return ImmediateProcessorMetrics{
		ReceivedJobs:  atomic.LoadUint64(&s.metrics.ReceivedJobs),
		ProcessedJobs: atomic.LoadUint64(&s.metrics.ProcessedJobs),
		SkippedJobs:   atomic.LoadUint64(&s.metrics.SkippedJobs),
		FailedJobs:    atomic.LoadUint64(&s.metrics.FailedJobs),
	}
}
