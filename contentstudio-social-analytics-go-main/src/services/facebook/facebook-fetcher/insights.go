package main

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	maxInsightWorkers = 10 // Number of parallel insight publishers
)

// InsightPublishJob represents a job to publish insights to Kafka
type InsightPublishJob struct {
	Insights   kafkamodels.RawFacebookInsights
	MessageKey string
	WorkerID   int
}

// InsightPublishResult represents the result of publishing insights
type InsightPublishResult struct {
	Success bool
	Error   error
	JobID   string
}

// fetchInsights retrieves Facebook page insights for a given page ID and access token.
// For incremental sync: fetches last 14 days
// For full_sync: fetches last 90 days
func fetchInsights(ctx context.Context, facebookClient *social.FacebookClient, pageID, accessToken, syncType string, log *logger.Logger) (*kafkamodels.RawFacebookInsights, error) {
	// Calculate date range based on sync type
	now := time.Now()
	until := now.AddDate(0, 0, -1) // Yesterday

	var since time.Time
	switch syncType {
	case "full_sync":
		since = until.AddDate(0, 0, -90) // 90 days ago for full sync
	default:
		since = until.AddDate(0, 0, -14) // 14 days ago for incremental
	}

	log.Info().
		Str("page_id", pageID).
		Str("since", since.Format("2006-01-02")).
		Str("until", until.Format("2006-01-02")).
		Msg("Fetching Facebook page insights")

	// Use the FacebookClient to fetch insights
	rawInsights, err := facebookClient.FetchInsights(ctx, pageID, accessToken, since, until)
	if err != nil {
		if isExpectedFacebookError(err) {
			log.Warn().
				Err(err).
				Str("page_id", pageID).
				Msg("Failed to fetch Facebook page insights")
		} else {
			log.Error().
				Err(err).
				Str("page_id", pageID).
				Msg("Failed to fetch Facebook page insights")
		}
		return nil, fmt.Errorf("fetchInsights: failed to fetch insights for page %s: %w", pageID, err)
	}

	log.Info().
		Str("page_id", pageID).
		Int("insights_count", len(rawInsights.Data)).
		Msg("Successfully fetched Facebook page insights")

	return rawInsights, nil
}

// publishInsightsParallel publishes insights to Kafka using parallel workers
func publishInsightsParallel(ctx context.Context, insights []kafkamodels.RawFacebookInsights, producer kafka.Producer, pageID, workspaceID string, workerID int, log *logger.Logger) {
	if len(insights) == 0 {
		log.Info().
			Str("page_id", pageID).
			Int("worker_id", workerID).
			Msg("No insights to publish")
		return
	}

	log.Info().
		Str("page_id", pageID).
		Int("worker_id", workerID).
		Int("insights_count", len(insights)).
		Int("workers", maxInsightWorkers).
		Msg("Starting parallel insights publishing")

	// Create jobs for each insight
	jobs := make([]InsightPublishJob, len(insights))
	for i, insight := range insights {
		// Set workspace ID for insights
		insight.WorkspaceID = workspaceID

		jobs[i] = InsightPublishJob{
			Insights:   insight,
			MessageKey: fmt.Sprintf("%s_%s", pageID, generateInsightID(insight.PageID, insight.SavingTime)),
			WorkerID:   workerID,
		}
	}

	// Create channels for job distribution and results
	jobChan := make(chan InsightPublishJob, len(jobs))
	resultChan := make(chan InsightPublishResult, len(jobs))

	// Start worker goroutines
	var wg sync.WaitGroup
	for i := 0; i < maxInsightWorkers; i++ {
		wg.Add(1)
		go func(workerNum int) {
			defer wg.Done()
			insightPublisher(ctx, jobChan, resultChan, producer, workerNum, log)
		}(i)
	}

	// Send jobs to workers
	go func() {
		defer close(jobChan)
		for _, job := range jobs {
			select {
			case jobChan <- job:
			case <-ctx.Done():
				return
			}
		}
	}()

	// Wait for all workers to complete
	wg.Wait()
	close(resultChan)

	// Collect results
	var successCount, errorCount int
	for result := range resultChan {
		if result.Success {
			successCount++
		} else {
			errorCount++
			log.Error().
				Err(result.Error).
				Str("job_id", result.JobID).
				Str("page_id", pageID).
				Int("worker_id", workerID).
				Msg("Failed to publish insight")
		}
	}

	log.Info().
		Str("page_id", pageID).
		Int("worker_id", workerID).
		Int("success_count", successCount).
		Int("error_count", errorCount).
		Msg("Completed parallel insights publishing")
}

// insightPublisher is a worker that publishes individual insights to Kafka
func insightPublisher(ctx context.Context, jobChan <-chan InsightPublishJob, resultChan chan<- InsightPublishResult, producer kafka.Producer, workerID int, log *logger.Logger) {
	for {
		select {
		case job, ok := <-jobChan:
			if !ok {
				return // Channel closed, worker should exit
			}

			jobID := fmt.Sprintf("%s_%d", job.MessageKey, workerID)

			insightJSON, err := json.Marshal(job.Insights)
			if err != nil {
				result := InsightPublishResult{
					Success: false,
					Error:   err,
					JobID:   jobID,
				}
				select {
				case resultChan <- result:
				case <-ctx.Done():
					return
				}
				continue
			}

			err = producer.Produce(ctx, "raw-facebook-insights", []byte(job.MessageKey), insightJSON)

			result := InsightPublishResult{
				Success: err == nil,
				Error:   err,
				JobID:   jobID,
			}

			if err != nil {
				log.Error().
					Err(err).
					Str("job_id", jobID).
					Str("page_id", job.Insights.PageID).
					Int("publisher_worker_id", workerID).
					Msg("Failed to publish insight to Kafka")
			}

			select {
			case resultChan <- result:
			case <-ctx.Done():
				return
			}

		case <-ctx.Done():
			return
		}
	}
}

// generateInsightID creates a unique ID for insights based on page ID and date
func generateInsightID(pageID string, savingTime time.Time) string {
	date := savingTime.Format("2006-01-02")
	nameID := fmt.Sprintf("%s_%s", pageID, date)
	hash := md5.Sum([]byte(nameID))
	return hex.EncodeToString(hash[:])
}

// joinStrings joins slice of strings with a separator (fetcher function)
func joinStrings(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	if len(slice) == 1 {
		return slice[0]
	}

	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}

// convertInsightsToInterface converts insights slice to interface{} slice for publishing
func convertInsightsToInterface(insights []kafkamodels.RawFacebookInsights) []interface{} {
	result := make([]interface{}, len(insights))
	for i, insight := range insights {
		result[i] = insight
	}
	return result
}
