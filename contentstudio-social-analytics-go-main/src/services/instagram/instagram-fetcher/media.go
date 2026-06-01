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
	maxMediaPublishWorkers = 10
)

type EnrichedMedia struct {
	*kafkamodels.RawInstagramMedia
	Insights *kafkamodels.RawInstagramMediaInsights `json:"insights,omitempty"`
	UserInfo map[string]interface{}                 `json:"user_info,omitempty"`
}

type MediaPublishJob struct {
	Media      EnrichedMedia
	MessageKey string
}

type MediaPublishResult struct {
	JobIndex int
	Success  bool
	Error    error
}

func publishMediaParallel(
	ctx context.Context,
	mediaItems []EnrichedMedia,
	producer kafka.Producer,
	instagramID, workspaceID string,
	workerID int,
	log *logger.Logger,
) {
	if len(mediaItems) == 0 {
		return
	}

	log.Info().
		Int("worker_id", workerID).
		Str("instagram_id", instagramID).
		Int("media_count", len(mediaItems)).
		Msg("Starting parallel media publishing")

	jobs := make([]MediaPublishJob, len(mediaItems))
	for i, media := range mediaItems {
		messageKey := fmt.Sprintf("%s_%s", instagramID, media.ID)
		if workspaceID != "" {
			messageKey = fmt.Sprintf("%s_%s_%s", workspaceID, instagramID, media.ID)
		}
		jobs[i] = MediaPublishJob{
			Media:      media,
			MessageKey: messageKey,
		}
	}

	jobChan := make(chan MediaPublishJob, len(jobs))
	resultChan := make(chan MediaPublishResult, len(jobs))

	var wg sync.WaitGroup
	for i := 0; i < maxMediaPublishWorkers; i++ {
		wg.Add(1)
		go mediaPublisher(ctx, &wg, i, jobChan, resultChan, producer, log)
	}

	go func() {
		defer close(jobChan)
		for i, job := range jobs {
			select {
			case jobChan <- job:
				log.Debug().
					Int("worker_id", workerID).
					Int("job_index", i).
					Str("media_id", job.Media.ID).
					Msg("Queued media for publishing")
			case <-ctx.Done():
				return
			}
		}
	}()

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
					Str("instagram_id", instagramID).
					Msg("Failed to publish media")
			}
		case <-ctx.Done():
			log.Warn().
				Int("worker_id", workerID).
				Str("instagram_id", instagramID).
				Msg("Media publishing cancelled due to context cancellation")
			return
		}
	}

	wg.Wait()

	log.Info().
		Int("worker_id", workerID).
		Str("instagram_id", instagramID).
		Int("success_count", successCount).
		Int("error_count", errorCount).
		Int("total_media", len(mediaItems)).
		Msg("Completed parallel media publishing")
}

func mediaPublisher(ctx context.Context, wg *sync.WaitGroup, workerID int, jobChan <-chan MediaPublishJob, resultChan chan<- MediaPublishResult, producer kafka.Producer, log *logger.Logger) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobChan:
			if !ok {
				return
			}

			result := MediaPublishResult{
				JobIndex: workerID,
			}

			mediaData, err := json.Marshal(job.Media)
			if err != nil {
				result.Error = err
				resultChan <- result
				continue
			}

			if err := producer.Produce(ctx, "raw-instagram-media", []byte(job.MessageKey), mediaData); err != nil {
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
