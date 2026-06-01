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
	maxPostWorkers = 10 // Maximum concurrent post publishers per work order
)

// PostPublishJob represents a job to publish a single raw post
type PostPublishJob struct {
	Post       kafkamodels.RawFacebookPost
	MessageKey string
}

// PostPublishResult represents the result of publishing a single raw post
type PostPublishResult struct {
	JobIndex int
	Success  bool
	Error    error
}

// publishPostsParallel publishes posts to Kafka in parallel using worker pool
func publishPostsParallel(
	ctx context.Context,
	posts []kafkamodels.RawFacebookPost,
	producer kafka.Producer,
	pageID, workspaceID string,
	workerID int,
	log *logger.Logger,
) {
	if len(posts) == 0 {
		return
	}

	log.Info().
		Int("worker_id", workerID).
		Str("page_id", pageID).
		Int("post_count", len(posts)).
		Msg("Starting parallel post publishing")

	// Create jobs for each post
	jobs := make([]PostPublishJob, len(posts))
	for i, post := range posts {
		//if post.StatusType == "added_video" {
		//	continue
		//}
		messageKey := fmt.Sprintf("%s_%s", pageID, post.ID)
		if workspaceID != "" {
			messageKey = fmt.Sprintf("%s_%s_%s", workspaceID, pageID, post.ID)
		}
		jobs[i] = PostPublishJob{
			Post:       post,
			MessageKey: messageKey,
		}
	}

	// Create channels for job distribution and result collection
	jobChan := make(chan PostPublishJob, len(jobs))
	resultChan := make(chan PostPublishResult, len(jobs))

	// Start post publisher workers
	var wg sync.WaitGroup
	for i := 0; i < maxPostWorkers; i++ {
		wg.Add(1)
		go postPublisher(ctx, &wg, i, jobChan, resultChan, producer, log)
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
					Str("post_id", job.Post.ID).
					Msg("Queued post for publishing")
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
					Msg("Failed to publish post")
			}
		case <-ctx.Done():
			log.Warn().
				Int("worker_id", workerID).
				Str("page_id", pageID).
				Msg("Post publishing cancelled due to context cancellation")
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
		Int("total_posts", len(posts)).
		Msg("Completed parallel post publishing")
}

// postPublisher is a worker that publishes posts to Kafka
func postPublisher(ctx context.Context, wg *sync.WaitGroup, workerID int, jobChan <-chan PostPublishJob, resultChan chan<- PostPublishResult, producer kafka.Producer, log *logger.Logger) {
	defer wg.Done()

	for {
		select {
		case job, ok := <-jobChan:
			if !ok {
				return
			}

			result := PostPublishResult{
				JobIndex: workerID,
			}

			// Serialize post to JSON
			postData, err := json.Marshal(job.Post)
			if err != nil {
				result.Error = err
				resultChan <- result
				continue
			}

			// Publish to Kafka
			if err := producer.Produce(ctx, "raw-facebook-posts", []byte(job.MessageKey), postData); err != nil {
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

// convertPostsToInterface converts []kafkamodels.RawFacebookPost to []interface{}
func convertPostsToInterface(posts []kafkamodels.RawFacebookPost) []interface{} {
	result := make([]interface{}, len(posts))
	for i, post := range posts {
		result[i] = post
	}
	return result
}
