package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// CompetitorAnalysisConfig holds service configuration
type CompetitorAnalysisConfig struct {
	WorkersPerPool int
}

// DefaultCompetitorAnalysisConfig returns default configuration
func DefaultCompetitorAnalysisConfig() CompetitorAnalysisConfig {
	return CompetitorAnalysisConfig{
		WorkersPerPool: WorkersPerPool,
	}
}

// CompetitorAnalysisDependencies holds external dependencies
type CompetitorAnalysisDependencies struct {
	RealtimeConsumer kafka.Consumer
	BatchConsumer    kafka.Consumer
	Log              *logger.Logger
}

// CompetitorAnalysisMetrics holds service metrics
type CompetitorAnalysisMetrics struct {
	RealtimeJobsReceived uint64
	BatchJobsReceived    uint64
	RealtimeJobsProcessed uint64
	BatchJobsProcessed   uint64
}

// CompetitorAnalysisService represents the competitor analysis service
type CompetitorAnalysisService struct {
	config  CompetitorAnalysisConfig
	deps    CompetitorAnalysisDependencies
	metrics CompetitorAnalysisMetrics
}

// NewCompetitorAnalysisService creates a new competitor analysis service
func NewCompetitorAnalysisService(cfg CompetitorAnalysisConfig, deps CompetitorAnalysisDependencies) *CompetitorAnalysisService {
	return &CompetitorAnalysisService{
		config: cfg,
		deps:   deps,
	}
}

// Run starts the competitor analysis service
func (s *CompetitorAnalysisService) Run(ctx context.Context) error {
	log := s.deps.Log

	realtimeJobs := make(chan CompetitorJob, 100)
	batchJobs := make(chan CompetitorJob, 100)

	var wg sync.WaitGroup

	// Start realtime workers
	for i := 0; i < s.config.WorkersPerPool; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.processJobs(ctx, workerID, "realtime", realtimeJobs, &s.metrics.RealtimeJobsProcessed, log)
		}(i)
	}

	// Start batch workers
	for i := 0; i < s.config.WorkersPerPool; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.processJobs(ctx, workerID, "batch", batchJobs, &s.metrics.BatchJobsProcessed, log)
		}(i)
	}

	// Start consumers
	var consumerWg sync.WaitGroup
	consumerWg.Add(2)

	go func() {
		defer consumerWg.Done()
		s.consumeRealtime(ctx, realtimeJobs, log)
	}()

	go func() {
		defer consumerWg.Done()
		s.consumeBatch(ctx, batchJobs, log)
	}()

	<-ctx.Done()

	close(realtimeJobs)
	close(batchJobs)

	wg.Wait()
	consumerWg.Wait()

	return nil
}

func (s *CompetitorAnalysisService) consumeRealtime(ctx context.Context, jobs chan<- CompetitorJob, log *logger.Logger) {
	s.deps.RealtimeConsumer.Consume(ctx, []string{"competitor-work-order-instagram"}, func(_ context.Context, _ string, _, value []byte) error {
		var wo kafkaModels.CompetitorWorkOrder
		if err := json.Unmarshal(value, &wo); err != nil {
			log.Warn().Err(err).Msg("Failed to unmarshal realtime work order")
			return nil
		}
		if wo.Channel != "instagram" {
			return nil
		}

		atomic.AddUint64(&s.metrics.RealtimeJobsReceived, 1)

		select {
		case jobs <- CompetitorJob{ReportID: wo.ReportID, PageID: wo.PageID, CompID: wo.PageID, Mode: apiModels.SyncMode(wo.Mode)}:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})
}

func (s *CompetitorAnalysisService) consumeBatch(ctx context.Context, jobs chan<- CompetitorJob, log *logger.Logger) {
	s.deps.BatchConsumer.Consume(ctx, []string{"competitor-work-order-instagram-batch"}, func(_ context.Context, _ string, _, value []byte) error {
		var wo kafkaModels.CompetitorWorkOrder
		if err := json.Unmarshal(value, &wo); err != nil {
			log.Warn().Err(err).Msg("Failed to unmarshal batch work order")
			return nil
		}
		if wo.Channel != "instagram" {
			return nil
		}

		atomic.AddUint64(&s.metrics.BatchJobsReceived, 1)

		select {
		case jobs <- CompetitorJob{ReportID: wo.ReportID, PageID: wo.PageID, CompID: wo.PageID, Mode: apiModels.SyncMode(wo.Mode)}:
		case <-ctx.Done():
			return ctx.Err()
		}
		return nil
	})
}

func (s *CompetitorAnalysisService) processJobs(ctx context.Context, workerID int, pool string, jobs <-chan CompetitorJob, counter *uint64, log *logger.Logger) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-jobs:
			if !ok {
				return
			}
			// Process job (actual processing would happen here)
			atomic.AddUint64(counter, 1)
			log.Debug().
				Int("worker_id", workerID).
				Str("pool", pool).
				Str("report_id", job.ReportID).
				Str("page_id", job.PageID).
				Msg("Processed job")
		}
	}
}

// GetMetrics returns current service metrics
func (s *CompetitorAnalysisService) GetMetrics() CompetitorAnalysisMetrics {
	return CompetitorAnalysisMetrics{
		RealtimeJobsReceived:  atomic.LoadUint64(&s.metrics.RealtimeJobsReceived),
		BatchJobsReceived:     atomic.LoadUint64(&s.metrics.BatchJobsReceived),
		RealtimeJobsProcessed: atomic.LoadUint64(&s.metrics.RealtimeJobsProcessed),
		BatchJobsProcessed:    atomic.LoadUint64(&s.metrics.BatchJobsProcessed),
	}
}
