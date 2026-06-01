package main

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	maxVideoWorkers = 10 // Maximum concurrent video publishers per work order
)

// VideoPublishJob represents a job to publish a single raw video
type VideoPublishJob struct {
	Video      kafkamodels.RawFacebookVideo
	MessageKey string
}

// VideoPublishResult represents the result of publishing a single raw video
type VideoPublishResult struct {
	JobIndex int
	Success  bool
	Error    error
}

// publishVideosParallel publishes videos to Kafka in parallel using worker pool
func publishVideosParallel(ctx context.Context, videos []kafkamodels.RawFacebookVideo, producer kafka.Producer, pageID, workspaceID string, workerID int, log *logger.Logger) {
	if len(videos) == 0 {
		return
	}

	log.Info().
		Int("worker_id", workerID).
		Str("page_id", pageID).
		Int("video_count", len(videos)).
		Msg("Starting parallel video publishing")

	// Create jobs for each video
	jobs := make([]VideoPublishJob, len(videos))
	for i, video := range videos {
		messageKey := fmt.Sprintf("%s_%s", pageID, video.ID)
		if workspaceID != "" {
			messageKey = fmt.Sprintf("%s_%s_%s", workspaceID, pageID, video.ID)
		}
		jobs[i] = VideoPublishJob{
			Video:      video,
			MessageKey: messageKey,
		}
	}

	// Create channels for job distribution and result collection
	jobChan := make(chan VideoPublishJob, len(jobs))
	resultChan := make(chan VideoPublishResult, len(jobs))

	// Start video publisher workers
	var wg sync.WaitGroup
	for i := 0; i < maxVideoWorkers; i++ {
		wg.Add(1)
		go videoPublisher(ctx, &wg, i, jobChan, resultChan, producer, log)
	}

	// Send jobs to workers
	go func() {
		defer close(jobChan)
		for i, job := range jobs {
			select {
			case jobChan <- job:
				log.Debug().
					Int("worker_id", workerID).
					Int("job_index", i).
					Str("video_id", job.Video.ID).
					Msg("Queued video for publishing")
			case <-ctx.Done():
				return
			}
		}
	}()

	// Collect results
	var successCount, errorCount int
	for i := 0; i < len(jobs); i++ {
		select {
		case result := <-resultChan:
			if result.Success {
				successCount++
			} else {
				errorCount++
				log.Error().Err(result.Error).
					Int("worker_id", workerID).
					Int("job_index", result.JobIndex).
					Str("page_id", pageID).
					Msg("Failed to publish video")
			}
		case <-ctx.Done():
			log.Warn().
				Int("worker_id", workerID).
				Str("page_id", pageID).
				Msg("Video publishing cancelled due to context cancellation")
			return
		}
	}

	// Wait for all workers to finish
	wg.Wait()

	log.Info().
		Int("worker_id", workerID).
		Str("page_id", pageID).
		Int("success_count", successCount).
		Int("error_count", errorCount).
		Int("total_videos", len(videos)).
		Msg("Completed parallel video publishing")
}

// videoPublisher is a worker that publishes videos to Kafka
func videoPublisher(ctx context.Context, wg *sync.WaitGroup, workerID int, jobChan <-chan VideoPublishJob, resultChan chan<- VideoPublishResult, producer kafka.Producer, log *logger.Logger) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobChan:
			if !ok {
				return
			}

			result := VideoPublishResult{
				JobIndex: workerID,
			}

			// Serialize video to JSON
			videoData, err := json.Marshal(job.Video)
			if err != nil {
				result.Error = err
				resultChan <- result
				continue
			}

			// Publish to Kafka
			if err := producer.Produce(ctx, "raw-facebook-videos", []byte(job.MessageKey), videoData); err != nil {
				result.Error = err
				resultChan <- result
				continue
			}

			result.Success = true
			resultChan <- result

		case <-ctx.Done():
			return
		}
	}
}

// convertVideosToInterface converts []kafkamodels.RawFacebookVideo to []interface{}
func convertVideosToInterface(videos []kafkamodels.RawFacebookVideo) []interface{} {
	result := make([]interface{}, len(videos))
	for i, video := range videos {
		result[i] = video
	}
	return result
}
