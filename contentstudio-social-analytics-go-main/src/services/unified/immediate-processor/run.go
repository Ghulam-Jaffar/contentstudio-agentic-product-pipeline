package main

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"

	fbprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/facebook/facebook-immediate-processor/processor"
	igprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/instagram/instagram-immediate-processor/processor"
	liprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/linkedin/linkedin-immediate-processor/processor"
	twprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/twitter/twitter-immediate-processor/processor"
)

// ServiceDependencies holds all injectable dependencies for the unified immediate processor.
// This struct enables dependency injection for testing with mock implementations.
type ServiceDependencies struct {
	FacebookConsumer  kafka.Consumer    // Kafka consumer for Facebook work orders
	InstagramConsumer kafka.Consumer    // Kafka consumer for Instagram work orders
	LinkedInConsumer  kafka.Consumer    // Kafka consumer for LinkedIn work orders
	YouTubeConsumer   kafka.Consumer    // Kafka consumer for YouTube work orders
	TikTokConsumer    kafka.Consumer    // Kafka consumer for TikTok work orders
	TwitterConsumer   kafka.Consumer    // Kafka consumer for Twitter work orders
	PinterestConsumer kafka.Consumer    // Kafka consumer for Pinterest work orders
	GMBConsumer       kafka.Consumer    // Kafka consumer for GMB work orders
	Processor         *UnifiedProcessor // Unified processor with platform handlers
	Logger            *logger.Logger    // Structured logger instance
	WorkerMultiplier  float64           // Multiplier for worker pool sizing (default: 1.0)
}

// RunService starts the unified immediate processor with the given dependencies.
// This is the main testable entry point that orchestrates queues, workers, and consumers.
//
// Architecture:
//   - Global queue provides admission control (100K capacity)
//   - Platform-specific queues route work to dedicated workers
//   - Workers process jobs and release global queue slots
//
// Performance: Handles ~48K accounts with 85 workers (40 FB, 30 IG, 15 LI at 1.0 multiplier)
func RunService(ctx context.Context, deps *ServiceDependencies) error {
	log := deps.Logger

	// Initialize two-tier queue system
	globalQueue := NewGlobalQueue(GlobalQueueCapacity)
	platformJobs := NewPlatformJobChannels()

	// Start worker pools for each platform
	var workerWg sync.WaitGroup
	StartWorkerPools(
		ctx,
		deps.Processor,
		platformJobs,
		globalQueue,
		deps.WorkerMultiplier,
		&workerWg,
		log,
	)

	// Start background stats logger
	go RunStatsLogger(ctx, globalQueue, platformJobs, log)

	// Start Kafka consumers
	var consumerWg sync.WaitGroup
	StartConsumers(ctx, deps, globalQueue, platformJobs, &consumerWg)

	// Cleanup: wait for consumers, then close channels
	go func() {
		consumerWg.Wait()
		platformJobs.CloseAll()
	}()

	// Wait for all workers to complete
	workerWg.Wait()
	log.Info().Msg("Unified Immediate Processor stopped")

	return nil
}

// StartWorkerPools starts worker goroutines for each platform.
// Each platform has a configurable number of workers based on PlatformSettings.
//
// Performance: Workers are scaled by workerMultiplier, minimum 1 per platform.
// Memory: Each worker holds minimal state, processing is sequential per worker.
func StartWorkerPools(
	ctx context.Context,
	processor *UnifiedProcessor,
	platformJobs *PlatformJobChannels,
	globalQueue *GlobalQueue,
	workerMultiplier float64,
	wg *sync.WaitGroup,
	log *logger.Logger,
) {
	for platform, pcfg := range PlatformSettings {
		workerCount := int(float64(pcfg.Workers) * workerMultiplier)
		if workerCount < 1 {
			workerCount = 1
		}

		jobChan := platformJobs.GetChannel(platform)

		log.Info().
			Str("platform", platform).
			Int("workers", workerCount).
			Int("queue_size", pcfg.QueueSize).
			Int("max_capacity", pcfg.MaxCapacity).
			Msg("Starting platform worker pool")

		// Launch workers for this platform
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func(p string, workerID int, jobs <-chan ImmediateWorkOrder) {
				defer wg.Done()
				processor.platformWorker(ctx, p, workerID, jobs, platformJobs, globalQueue)
			}(platform, i, jobChan)
		}
	}
}

// RunStatsLogger logs queue statistics periodically using default 30-second interval.
func RunStatsLogger(
	ctx context.Context,
	globalQueue *GlobalQueue,
	platformJobs *PlatformJobChannels,
	log *logger.Logger,
) {
	RunStatsLoggerWithInterval(ctx, globalQueue, platformJobs, log, 30*time.Second)
}

// RunStatsLoggerWithInterval logs queue statistics with configurable interval.
// Useful for testing with shorter intervals.
func RunStatsLoggerWithInterval(
	ctx context.Context,
	globalQueue *GlobalQueue,
	platformJobs *PlatformJobChannels,
	log *logger.Logger,
	interval time.Duration,
) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			LogQueueStats(globalQueue, platformJobs, log)
		}
	}
}

// LogQueueStats logs current queue statistics for monitoring.
// Logs global queue stats and per-platform queue stats.
func LogQueueStats(
	globalQueue *GlobalQueue,
	platformJobs *PlatformJobChannels,
	log *logger.Logger,
) {
	current, capacity, admitted, rejected := globalQueue.Stats()
	log.Info().
		Int64("global_current", current).
		Int64("global_capacity", capacity).
		Int64("global_admitted", admitted).
		Int64("global_rejected", rejected).
		Float64("global_utilization_pct", CalculateUtilization(current, capacity)).
		Msg("Global queue stats")

	stats := platformJobs.GetStats()
	for platform, s := range stats {
		log.Info().
			Str("platform", platform).
			Int("queue_depth", s.QueueDepth).
			Int("queue_capacity", s.MaxCapacity).
			Int64("processed", s.Processed).
			Int64("dropped", s.Dropped).
			Float64("utilization_pct", CalculateQueueUtilization(s.QueueDepth, s.MaxCapacity)).
			Msg("Platform queue stats")
	}
}

// StartConsumers starts Kafka consumers for all configured platforms.
// Each consumer runs in its own goroutine and processes messages asynchronously.
func StartConsumers(
	ctx context.Context,
	deps *ServiceDependencies,
	globalQueue *GlobalQueue,
	platformJobs *PlatformJobChannels,
	wg *sync.WaitGroup,
) {
	log := deps.Logger

	// Message handler shared by all consumers
	handleMessage := func(topic string, value []byte) {
		ProcessKafkaMessage(topic, value, globalQueue, platformJobs, log)
	}

	// Start platform-specific consumers
	startConsumerIfNotNil(
		ctx, deps.FacebookConsumer, wg, log, handleMessage,
		"immediate-work-order-facebook", facebookConsumerGroup, "Facebook",
	)
	startConsumerIfNotNil(
		ctx, deps.InstagramConsumer, wg, log, handleMessage,
		"immediate-work-order-instagram", instagramConsumerGroup, "Instagram",
	)
	startConsumerIfNotNil(
		ctx, deps.LinkedInConsumer, wg, log, handleMessage,
		"immediate-work-order-linkedin", linkedinConsumerGroup, "LinkedIn",
	)
	startConsumerIfNotNil(
		ctx, deps.YouTubeConsumer, wg, log, handleMessage,
		"immediate-work-order-youtube", youtubeConsumerGroup, "YouTube",
	)
	startConsumerIfNotNil(
		ctx, deps.TikTokConsumer, wg, log, handleMessage,
		"immediate-work-order-tiktok", tiktokConsumerGroup, "TikTok",
	)
	startConsumerIfNotNil(
		ctx, deps.TwitterConsumer, wg, log, handleMessage,
		"immediate-work-order-twitter", twitterConsumerGroup, "Twitter",
	)
	startConsumerIfNotNil(
		ctx, deps.PinterestConsumer, wg, log, handleMessage,
		"immediate-work-order-pinterest", pinterestConsumerGroup, "Pinterest",
	)
	startConsumerIfNotNil(
		ctx, deps.GMBConsumer, wg, log, handleMessage,
		"immediate-work-order-gmb", gmbConsumerGroup, "GMB",
	)
}

// startConsumerIfNotNil starts a consumer goroutine if the consumer is not nil.
// Extracted to reduce code duplication in StartConsumers.
func startConsumerIfNotNil(
	ctx context.Context,
	consumer kafka.Consumer,
	wg *sync.WaitGroup,
	log *logger.Logger,
	handleMessage func(topic string, value []byte),
	topic, group, name string,
) {
	if consumer == nil {
		return
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		log.Info().
			Str("topic", topic).
			Str("group", group).
			Msgf("Starting %s consumer", name)

		err := consumer.Consume(
			ctx,
			[]string{topic},
			func(ctx context.Context, t string, key, value []byte) error {
				handleMessage(t, value)
				return nil
			},
		)
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msgf("%s consumer error", name)
		}
	}()
}

// ProcessKafkaMessage processes a single Kafka message through the queue system.
// Implements two-tier admission control: global queue then platform queue.
//
// Performance: O(1) for admission control, O(1) for queue operations.
// Backpressure: Rejects messages when queues are full to prevent memory exhaustion.
func ProcessKafkaMessage(
	topic string,
	value []byte,
	globalQueue *GlobalQueue,
	platformJobs *PlatformJobChannels,
	log *logger.Logger,
) {
	wo, err := decodeImmediateWorkOrder(topic, value)
	if err != nil {
		log.Error().Err(err).Str("topic", topic).Msg("Failed to unmarshal work order")
		return
	}

	// Determine platform from work order or topic
	if wo.Platform == "" {
		wo.Platform = inferPlatformFromTopic(topic)
	}
	if wo.Platform == "" {
		log.Warn().Str("topic", topic).Msg("Could not determine platform")
		return
	}

	// STEP 1: Global admission control
	if !globalQueue.TryAdmit() {
		logGlobalQueueFull(globalQueue, wo, log)
		return
	}

	// STEP 2: Platform queue routing
	if !platformJobs.TryEnqueue(wo.Platform, wo) {
		globalQueue.Release()
		logPlatformQueueFull(platformJobs, wo, log)
	}
}

func decodeImmediateWorkOrder(topic string, value []byte) (ImmediateWorkOrder, error) {
	var wo ImmediateWorkOrder
	if err := json.Unmarshal(value, &wo); err != nil {
		return ImmediateWorkOrder{}, err
	}

	if strings.Contains(topic, "twitter") {
		var twWO twprocessor.ImmediateWorkOrder
		if err := json.Unmarshal(value, &twWO); err == nil && twWO.TwitterID != "" {
			wo.Platform = "twitter"
			wo.AccountID = twWO.TwitterID
			wo.AccessToken = twWO.OAuthToken
			wo.RefreshToken = twWO.OAuthTokenSecret
			wo.WorkspaceID = twWO.WorkspaceID
			wo.SyncType = twWO.SyncType
			wo.TwitterID = twWO.TwitterID
			wo.OAuthToken = twWO.OAuthToken
			wo.OAuthTokenSecret = twWO.OAuthTokenSecret
			wo.PostCount = twWO.NTweets
			wo.APIKey = twWO.APIKey
			wo.APISecret = twWO.APISecret
			wo.AppName = twWO.AppName
			wo.AppID = twWO.AppID
			wo.ExecutedBy = twWO.ExecutedBy
		}
	}

	return wo, nil
}

// logGlobalQueueFull logs when a work order is rejected due to global queue capacity.
func logGlobalQueueFull(globalQueue *GlobalQueue, wo ImmediateWorkOrder, log *logger.Logger) {
	current, capacity, _, rejected := globalQueue.Stats()
	log.Warn().
		Str("platform", wo.Platform).
		Str("account_id", wo.AccountID).
		Str("workspace_id", wo.WorkspaceID).
		Int64("global_current", current).
		Int64("global_capacity", capacity).
		Int64("global_rejected", rejected).
		Msg("SYSTEM BUSY - Global queue full, work order rejected")
}

// logPlatformQueueFull logs when a work order is dropped due to platform queue capacity.
func logPlatformQueueFull(
	platformJobs *PlatformJobChannels,
	wo ImmediateWorkOrder,
	log *logger.Logger,
) {
	stats := platformJobs.GetStats()
	s := stats[wo.Platform]
	log.Warn().
		Str("platform", wo.Platform).
		Str("account_id", wo.AccountID).
		Str("workspace_id", wo.WorkspaceID).
		Int("queue_depth", s.QueueDepth).
		Int("queue_capacity", s.MaxCapacity).
		Int64("dropped", s.Dropped).
		Msg("Platform queue full, work order dropped")
}

// CreateProcessorFromDeps creates a UnifiedProcessor from processor implementations.
// Used in main() to create the processor with real implementations.
func CreateProcessorFromDeps(
	fbProc FacebookProcessor,
	igProc InstagramProcessor,
	liProc LinkedInProcessor,
	tkProc TikTokProcessor,
	log *logger.Logger,
) *UnifiedProcessor {
	return &UnifiedProcessor{
		facebookProcessor:  fbProc,
		instagramProcessor: igProc,
		linkedinProcessor:  liProc,
		tiktokProcessor:    tkProc,
		logger:             log,
	}
}

// ConvertToFacebookWorkOrder converts ImmediateWorkOrder to Facebook processor format.
func ConvertToFacebookWorkOrder(wo ImmediateWorkOrder) fbprocessor.WorkOrder {
	return fbprocessor.WorkOrder{
		ID:              wo.ID,
		AccountID:       wo.AccountID,
		Type:            wo.Type,
		AccessToken:     wo.AccessToken,
		WorkspaceID:     wo.WorkspaceID,
		LongAccessToken: wo.LongAccessToken,
		SyncType:        wo.SyncType,
	}
}

// ConvertToInstagramWorkOrder converts ImmediateWorkOrder to Instagram processor format.
func ConvertToInstagramWorkOrder(wo ImmediateWorkOrder) igprocessor.WorkOrder {
	return igprocessor.WorkOrder{
		ID:                    wo.ID,
		AccountID:             wo.AccountID,
		Type:                  wo.Type,
		AccessToken:           wo.AccessToken,
		WorkspaceID:           wo.WorkspaceID,
		SyncType:              wo.SyncType,
		ConnectedViaInstagram: wo.ConnectedViaInstagram,
	}
}

// ConvertToLinkedInWorkOrder converts ImmediateWorkOrder to LinkedIn processor format.
func ConvertToLinkedInWorkOrder(wo ImmediateWorkOrder) liprocessor.WorkOrder {
	return liprocessor.WorkOrder{
		ID:          wo.ID,
		AccountID:   wo.AccountID,
		AccessToken: wo.AccessToken,
		WorkspaceID: wo.WorkspaceID,
		SyncType:    wo.SyncType,
		StartDate:   wo.StartDate,
		EndDate:     wo.EndDate,
	}
}

// GetDefaultWorkerMultiplier returns the default worker multiplier (1.0).
func GetDefaultWorkerMultiplier() float64 {
	return 1.0
}

// CalculateWorkerCount calculates the number of workers for a platform.
// Returns minimum of 1 even with zero or negative multiplier.
func CalculateWorkerCount(platform string, multiplier float64) int {
	pcfg, ok := PlatformSettings[platform]
	if !ok {
		return 1
	}
	workerCount := int(float64(pcfg.Workers) * multiplier)
	if workerCount < 1 {
		return 1
	}
	return workerCount
}

// GetPlatformQueueCapacity returns the queue capacity for a platform.
// Returns 50 (default) for unknown platforms.
func GetPlatformQueueCapacity(platform string) int {
	pcfg, ok := PlatformSettings[platform]
	if !ok {
		return 50
	}
	return pcfg.QueueSize
}

// GetPlatformMaxCapacity returns the max capacity for a platform.
// Returns 0 for unknown platforms.
func GetPlatformMaxCapacity(platform string) int {
	pcfg, ok := PlatformSettings[platform]
	if !ok {
		return 0
	}
	return pcfg.MaxCapacity
}

// GetGlobalQueueDefaultCapacity returns the default global queue capacity (100K).
func GetGlobalQueueDefaultCapacity() int64 {
	return GlobalQueueCapacity
}

// ValidateWorkOrder validates a work order has required fields.
func ValidateWorkOrder(wo ImmediateWorkOrder) bool {
	return wo.Platform != "" && wo.AccountID != ""
}

// GetConsumerGroupForPlatform returns the consumer group name for a platform.
// Returns empty string for unknown platforms.
func GetConsumerGroupForPlatform(platform string) string {
	switch platform {
	case "facebook":
		return facebookConsumerGroup
	case "instagram":
		return instagramConsumerGroup
	case "linkedin":
		return linkedinConsumerGroup
	case "youtube":
		return youtubeConsumerGroup
	case "tiktok":
		return tiktokConsumerGroup
	case "twitter":
		return twitterConsumerGroup
	case "pinterest":
		return pinterestConsumerGroup
	case "gmb":
		return gmbConsumerGroup
	default:
		return ""
	}
}

// GetTopicForPlatform returns the Kafka topic for a platform.
// Returns empty string for unknown platforms.
func GetTopicForPlatform(platform string) string {
	switch platform {
	case "facebook":
		return "immediate-work-order-facebook"
	case "instagram":
		return "immediate-work-order-instagram"
	case "linkedin":
		return "immediate-work-order-linkedin"
	case "youtube":
		return "immediate-work-order-youtube"
	case "tiktok":
		return "immediate-work-order-tiktok"
	case "twitter":
		return "immediate-work-order-twitter"
	case "pinterest":
		return "immediate-work-order-pinterest"
	case "gmb":
		return "immediate-work-order-gmb"
	default:
		return ""
	}
}

// IsValidWorkOrderJSON validates JSON can be parsed as a work order.
func IsValidWorkOrderJSON(data []byte) bool {
	var wo ImmediateWorkOrder
	return json.Unmarshal(data, &wo) == nil
}

// GetAllConsumerGroups returns all consumer group names mapped by platform.
func GetAllConsumerGroups() map[string]string {
	return map[string]string{
		"facebook":  facebookConsumerGroup,
		"instagram": instagramConsumerGroup,
		"linkedin":  linkedinConsumerGroup,
		"youtube":   youtubeConsumerGroup,
		"tiktok":    tiktokConsumerGroup,
		"twitter":   twitterConsumerGroup,
		"pinterest": pinterestConsumerGroup,
		"gmb":       gmbConsumerGroup,
	}
}

// CalculateQueueFillPercentage calculates how full a queue is as a percentage.
// Returns 0.0 if capacity is zero to avoid division by zero.
func CalculateQueueFillPercentage(current, capacity int) float64 {
	if capacity == 0 {
		return 0.0
	}
	return float64(current) / float64(capacity) * 100.0
}

// EstimateTotalWorkers estimates total workers across all platforms.
// Performance: O(n) where n is number of platforms.
func EstimateTotalWorkers(multiplier float64) int {
	total := 0
	for platform := range PlatformSettings {
		total += CalculateWorkerCount(platform, multiplier)
	}
	return total
}

// GetAllPlatformTopics returns all Kafka topics for platforms.
func GetAllPlatformTopics() []string {
	return []string{
		"immediate-work-order-facebook",
		"immediate-work-order-instagram",
		"immediate-work-order-linkedin",
		"immediate-work-order-youtube",
		"immediate-work-order-tiktok",
		"immediate-work-order-twitter",
		"immediate-work-order-pinterest",
		"immediate-work-order-gmb",
	}
}
