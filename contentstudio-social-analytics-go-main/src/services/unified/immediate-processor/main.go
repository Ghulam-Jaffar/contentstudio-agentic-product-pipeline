package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"os"
	"os/signal"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/notification"

	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	fbprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/facebook/facebook-immediate-processor/processor"
	gmbprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/gmb/gmb-immediate-processor/processor"
	igprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/instagram/instagram-immediate-processor/processor"
	liprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/linkedin/linkedin-immediate-processor/processor"
	metaadsprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/meta-ads/meta-ads-immediate-processor/processor"
	ptprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/pinterest/pinterest-immediate-processor/processor"
	tkprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/tiktok/tiktok-immediate-processor/processor"
	twprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/twitter/twitter-immediate-processor/processor"
	ytprocessor "github.com/d4interactive/contentstudio-social-analytics-go/src/services/youtube/youtube-immediate-processor/processor"
)

// =============================================================================
// UNIFIED IMMEDIATE PROCESSOR - QUEUE ARCHITECTURE
// =============================================================================
//
// This service processes immediate analytics requests for all social platforms
// (Facebook, Instagram, LinkedIn) using a two-tier queue system:
//
//                              ┌─────────────────────────────────┐
//                              │       GLOBAL QUEUE              │
//                              │   Capacity: 100K                │
//                              │   Admission control for all     │
//                              │   platforms combined            │
//                              │                                 │
//                              │   If full → Reject immediately  │
//                              │   User gets "system busy" msg   │
//                              └───────────────┬─────────────────┘
//                                              │
//                                              │ Admitted requests
//                                              │ routed by platform
//                                              ▼
//                    ┌─────────────────────────┼─────────────────────────┐
//                    │                         │                         │
//                    ▼                         ▼                         ▼
//          ┌─────────────────┐       ┌─────────────────┐       ┌─────────────────┐
//          │  Facebook Queue │       │ Instagram Queue │       │  LinkedIn Queue │
//          │   Capacity: 24K │       │  Capacity: 16K  │       │  Capacity: 8K   │
//          │   Workers: 40   │       │   Workers: 30   │       │   Workers: 15   │
//          └────────┬────────┘       └────────┬────────┘       └────────┬────────┘
//                   │                         │                         │
//                   │ Parallel                │ Parallel                │ Parallel
//                   │ goroutines              │ goroutines              │ goroutines
//                   ▼                         ▼                         ▼
//             40 Workers                 30 Workers                15 Workers
//             (concurrent)              (concurrent)              (concurrent)
//
// KEY DESIGN DECISIONS:
//
// 1. GLOBAL QUEUE (100K capacity):
//    - Acts as admission control for the entire system
//    - Prevents memory exhaustion from unbounded queuing
//    - Allows immediate rejection when system is overloaded
//    - Capacity: 100K to accommodate future platforms (Twitter, TikTok, etc.)
//
// 2. PLATFORM QUEUES (run in parallel):
//    - Each platform has dedicated workers running as goroutines
//    - Facebook processing does NOT block Instagram or LinkedIn
//    - Workers process from their platform queue concurrently
//
// 3. QUEUE SIZES (designed for 3 replicas):
//    - Global: 100K total (33K per replica)
//    - Facebook: 24K (highest volume)
//    - Instagram: 16K (medium volume)
//    - LinkedIn: 8K (lower volume)
//    - Reserved ~52K for future platforms
//
// 4. DATA PERSISTENCE:
//    - Queues are IN-MEMORY (Go channels)
//    - If app crashes, queued work orders are LOST
//    - Kafka offset committed before processing (at-most-once delivery)
//    - ClickHouse ReplacingMergeTree handles duplicates on retry
//
// 5. ADDING NEW PLATFORMS:
//    - Add entry to PlatformConfig map
//    - Import processor package
//    - Add case to platformWorker switch
//    - Add topic to topics slice
//
// =============================================================================

// Consumer groups - reuse existing immediate processor groups
const (
	facebookConsumerGroup  = "immediate-processor-group"
	instagramConsumerGroup = "instagram-immediate-processor-group"
	linkedinConsumerGroup  = "linkedin-immediate-processor-group"
	youtubeConsumerGroup   = "youtube-immediate-processor-group"
	tiktokConsumerGroup    = "tiktok-immediate-processor-group"
	twitterConsumerGroup   = "twitter-immediate-processor-group"
	pinterestConsumerGroup = "pinterest-immediate-processor-group"
	gmbConsumerGroup       = "gmb-immediate-processor-group"
	metaAdsConsumerGroup   = "meta-ads-immediate-processor-group"
)

// GlobalQueueCapacity is the maximum number of work orders that can be queued
// across all platforms combined. Set to 100K to accommodate:
// - Current: Facebook (24K) + Instagram (16K) + LinkedIn (8K) = 48K
// - Future: Twitter, TikTok, Pinterest, YouTube = ~52K reserved
// With 3 replicas, each handles ~33K concurrent requests
const GlobalQueueCapacity = 100000

// PlatformConfig defines worker and queue settings for each social platform.
// To add a new platform:
// 1. Add entry here with appropriate Workers, QueueSize, MaxCapacity
// 2. Import the processor package
// 3. Add processing case in platformWorker()
// 4. Add Kafka topic to topics slice
type PlatformConfig struct {
	Workers     int // Number of concurrent goroutines processing this platform
	QueueSize   int // Channel buffer size (immediate backpressure threshold)
	MaxCapacity int // Maximum accounts this platform can handle
}

var PlatformSettings = map[string]PlatformConfig{
	"facebook":  {Workers: 40, QueueSize: 500, MaxCapacity: 24000},
	"instagram": {Workers: 30, QueueSize: 400, MaxCapacity: 16000},
	"linkedin":  {Workers: 15, QueueSize: 200, MaxCapacity: 8000},
	"tiktok":    {Workers: 10, QueueSize: 150, MaxCapacity: 5000},
	// Future platforms - uncomment when ready:
	"twitter":   {Workers: 25, QueueSize: 300, MaxCapacity: 12000},
	"pinterest": {Workers: 15, QueueSize: 200, MaxCapacity: 8000},
	"youtube":   {Workers: 20, QueueSize: 250, MaxCapacity: 10000},
	"gmb":       {Workers: 10, QueueSize: 150, MaxCapacity: 5000},
	"meta_ads":  {Workers: 15, QueueSize: 200, MaxCapacity: 8000},
}

// ImmediateWorkOrder represents a request to process analytics for a social account.
// Consumed from Kafka topics: immediate-work-order-{platform}
type ImmediateWorkOrder struct {
	ID                    string `json:"id"`
	Platform              string `json:"platform"`
	AccountID             string `json:"account_id"`
	Type                  string `json:"type"`
	AccessToken           string `json:"access_token"`
	LongAccessToken       string `json:"long_access_token"`
	RefreshToken          string `json:"refresh_token"`
	WorkspaceID           string `json:"workspace_id"`
	SyncType              string `json:"sync_type"`
	ConnectedViaInstagram bool   `json:"connected_via_instagram"`
	StartDate             string `json:"start_date,omitempty"`
	EndDate               string `json:"end_date,omitempty"`

	// Twitter-specific payload fields
	TwitterID        string `json:"twitter_id"`
	OAuthToken       string `json:"oauth_token"`
	OAuthTokenSecret string `json:"oauth_token_secret"`
	PostCount        int    `json:"post_count"`
	APIKey           string `json:"api_key"`
	APISecret        string `json:"api_secret"`
	AppName          string `json:"app_name"`
	AppID            string `json:"app_id"`
	ExecutedBy       string `json:"executed_by"`

	// GMB-specific payload fields
	LocationID   string `json:"location_id"`
	AccountName  string `json:"account_name"`
	LocationName string `json:"location_name"`
	LanguageCode string `json:"language_code"`
}

// =============================================================================
// GLOBAL QUEUE - Admission Control
// =============================================================================

// GlobalQueue provides system-wide admission control.
// When the queue is full, new requests are rejected immediately,
// allowing the user to be notified that the system is busy.
type GlobalQueue struct {
	current  int64 // Current number of items across all platform queues
	capacity int64 // Maximum allowed items (GlobalQueueCapacity)
	admitted int64 // Total items admitted (for stats)
	rejected int64 // Total items rejected due to capacity (for stats)
}

// NewGlobalQueue creates a new global queue with the specified capacity
func NewGlobalQueue(capacity int) *GlobalQueue {
	return &GlobalQueue{
		capacity: int64(capacity),
	}
}

// TryAdmit attempts to admit a work order to the global queue.
// Returns true if admitted, false if queue is full (system busy).
// This is the first check before routing to platform-specific queues.
func (gq *GlobalQueue) TryAdmit() bool {
	for {
		current := atomic.LoadInt64(&gq.current)
		if current >= gq.capacity {
			atomic.AddInt64(&gq.rejected, 1)
			return false
		}
		if atomic.CompareAndSwapInt64(&gq.current, current, current+1) {
			atomic.AddInt64(&gq.admitted, 1)
			return true
		}
	}
}

// Release decrements the global queue counter when a work order completes.
// Must be called after processing completes (success or failure).
func (gq *GlobalQueue) Release() {
	atomic.AddInt64(&gq.current, -1)
}

// Stats returns current global queue statistics
func (gq *GlobalQueue) Stats() (current, capacity, admitted, rejected int64) {
	return atomic.LoadInt64(&gq.current),
		gq.capacity,
		atomic.LoadInt64(&gq.admitted),
		atomic.LoadInt64(&gq.rejected)
}

// =============================================================================
// PLATFORM QUEUES - Per-Platform Processing
// =============================================================================

// QueueStats tracks statistics for monitoring and debugging
type QueueStats struct {
	Platform    string
	QueueDepth  int   // Current items in queue
	MaxCapacity int   // Queue channel capacity
	Processed   int64 // Total successfully processed
	Dropped     int64 // Total dropped due to full queue
}

// PlatformJobChannels manages per-platform job queues and statistics.
// Each platform has its own channel and worker pool that run in parallel.
type PlatformJobChannels struct {
	channels  map[string]chan ImmediateWorkOrder
	processed map[string]*int64 // Atomic counters per platform
	dropped   map[string]*int64 // Atomic counters per platform
	mu        sync.RWMutex
}

// NewPlatformJobChannels initializes channels for all configured platforms
func NewPlatformJobChannels() *PlatformJobChannels {
	pjc := &PlatformJobChannels{
		channels:  make(map[string]chan ImmediateWorkOrder),
		processed: make(map[string]*int64),
		dropped:   make(map[string]*int64),
	}
	for platform, cfg := range PlatformSettings {
		pjc.channels[platform] = make(chan ImmediateWorkOrder, cfg.QueueSize)
		var processed, dropped int64
		pjc.processed[platform] = &processed
		pjc.dropped[platform] = &dropped
	}
	return pjc
}

// GetChannel returns the job channel for a platform.
// Creates a default channel if platform not in config (for future extensibility).
func (pjc *PlatformJobChannels) GetChannel(platform string) chan ImmediateWorkOrder {
	pjc.mu.RLock()
	ch, ok := pjc.channels[platform]
	pjc.mu.RUnlock()
	if ok {
		return ch
	}
	// Create default channel for unknown platforms
	pjc.mu.Lock()
	defer pjc.mu.Unlock()
	if ch, ok = pjc.channels[platform]; ok {
		return ch
	}
	ch = make(chan ImmediateWorkOrder, 50)
	pjc.channels[platform] = ch
	var processed, dropped int64
	pjc.processed[platform] = &processed
	pjc.dropped[platform] = &dropped
	return ch
}

// TryEnqueue attempts to add a work order to the platform queue.
// Returns false if the platform queue is full.
// Note: Global queue admission should be checked BEFORE calling this.
func (pjc *PlatformJobChannels) TryEnqueue(platform string, wo ImmediateWorkOrder) bool {
	pjc.mu.RLock()
	ch, ok := pjc.channels[platform]
	dropped := pjc.dropped[platform]
	pjc.mu.RUnlock()

	if !ok {
		return false
	}

	select {
	case ch <- wo:
		return true
	default:
		// Platform queue is full
		if dropped != nil {
			atomic.AddInt64(dropped, 1)
		}
		return false
	}
}

// IncrementProcessed increments the successful processing counter
func (pjc *PlatformJobChannels) IncrementProcessed(platform string) {
	pjc.mu.RLock()
	processed := pjc.processed[platform]
	pjc.mu.RUnlock()
	if processed != nil {
		atomic.AddInt64(processed, 1)
	}
}

// GetStats returns current statistics for all platforms
func (pjc *PlatformJobChannels) GetStats() map[string]QueueStats {
	pjc.mu.RLock()
	defer pjc.mu.RUnlock()

	result := make(map[string]QueueStats)
	for platform, ch := range pjc.channels {
		stats := QueueStats{
			Platform:    platform,
			QueueDepth:  len(ch),
			MaxCapacity: cap(ch),
		}
		if processed := pjc.processed[platform]; processed != nil {
			stats.Processed = atomic.LoadInt64(processed)
		}
		if dropped := pjc.dropped[platform]; dropped != nil {
			stats.Dropped = atomic.LoadInt64(dropped)
		}
		result[platform] = stats
	}
	return result
}

// CloseAll closes all platform channels, signaling workers to stop
func (pjc *PlatformJobChannels) CloseAll() {
	pjc.mu.Lock()
	defer pjc.mu.Unlock()
	for _, ch := range pjc.channels {
		close(ch)
	}
}

// =============================================================================
// UNIFIED PROCESSOR - Main Service
// =============================================================================

// FacebookProcessor interface for testing
type FacebookProcessor interface {
	ProcessAccount(ctx context.Context, wo fbprocessor.WorkOrder) error
}

// InstagramProcessor interface for testing
type InstagramProcessor interface {
	ProcessAccount(ctx context.Context, wo igprocessor.WorkOrder) error
}

// LinkedInProcessor interface for testing
type LinkedInProcessor interface {
	ProcessAccount(ctx context.Context, wo liprocessor.WorkOrder) error
}

// YouTubeProcessor interface for testing
type YouTubeProcessor interface {
	ProcessAccount(ctx context.Context, wo ytprocessor.WorkOrder) error
}

// TikTokProcessor interface for testing
type TikTokProcessor interface {
	ProcessAccount(ctx context.Context, wo tkprocessor.ImmediateWorkOrder) error
}

// TwitterProcessor interface for testing
type TwitterProcessor interface {
	ProcessAccount(ctx context.Context, wo twprocessor.ImmediateWorkOrder) error
}

// PinterestProcessor interface for testing
type PinterestProcessor interface {
	ProcessAccount(ctx context.Context, wo ptprocessor.WorkOrder) error
}

// GMBProcessor interface for testing
type GMBProcessor interface {
	ProcessAccount(ctx context.Context, wo gmbprocessor.ImmediateWorkOrder) error
}

// MetaAdsProcessor interface for testing
type MetaAdsProcessor interface {
	ProcessAccount(ctx context.Context, wo kafkamodels.MetaAdsWorkOrder) error
}

// UnifiedProcessor holds references to all platform-specific processors
type UnifiedProcessor struct {
	instagramProcessor InstagramProcessor
	facebookProcessor  FacebookProcessor
	linkedinProcessor  LinkedInProcessor
	youtubeProcessor   YouTubeProcessor
	tiktokProcessor    TikTokProcessor
	twitterProcessor   TwitterProcessor
	pinterestProcessor PinterestProcessor
	gmbProcessor       GMBProcessor
	metaAdsProcessor   MetaAdsProcessor
	mongoRepo          mongodb.UnifiedSocialRepository
	logger             *logger.Logger
}

func main() {
	// Command-line flags
	workerMultiplier := flag.Float64("workerMultiplier", 1.0, "Multiplier for worker counts (for scaling)")
	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)

	// Calculate total workers for logging
	totalWorkers := 0
	for _, pcfg := range PlatformSettings {
		totalWorkers += int(float64(pcfg.Workers) * *workerMultiplier)
	}

	log.Info().
		Int("total_workers", totalWorkers).
		Int("global_queue_capacity", GlobalQueueCapacity).
		Float64("worker_multiplier", *workerMultiplier).
		Msg("Starting Unified Immediate Processor")

	// =========================================================================
	// DATABASE CONNECTIONS
	// =========================================================================

	// MongoDB connection (for account information)
	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	mongoClient, err := mongo.Connect(context.Background(), options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential))
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}
	cancel()
	log.Info().Msg("Connected to MongoDB")

	// =========================================================================
	// SHARED DEPENDENCIES
	// =========================================================================

	mongoRepo := mongodb.NewUnifiedSocialRepository(mongoClient.Database(cfg.Mongo.Database), log.Logger)
	sink := conversions.NewClickHouseSink(&log.Logger, cfg)
	producer := mustCreateProducer(cfg.Kafka, log.Logger)
	notifier := notification.NewService(cfg.Email, log.Logger, cfg.Email.BackendURL)
	pusherClient := notification.NewPusherClient(cfg.Pusher, log.Logger)

	// =========================================================================
	// INITIALIZE PROCESSORS
	// =========================================================================

	// Each platform processor is initialized with shared dependencies.
	// Processors contain the actual business logic for fetching, parsing,
	// and storing analytics data. API clients are created internally by each processor.

	processor := &UnifiedProcessor{
		instagramProcessor: igprocessor.New(mongoRepo, sink, producer, notifier, pusherClient, log, cfg),
		facebookProcessor:  fbprocessor.New(mongoRepo, sink, producer, notifier, pusherClient, log, cfg),
		linkedinProcessor:  liprocessor.New(mongoRepo, sink, producer, notifier, pusherClient, log, cfg),
		youtubeProcessor:   ytprocessor.New(mongoRepo, sink, notifier, pusherClient, log, cfg),
		tiktokProcessor:    tkprocessor.New(mongoRepo, sink, notifier, pusherClient, log, cfg),
		twitterProcessor:   twprocessor.New(mongoRepo, sink, notifier, pusherClient, log, cfg),
		pinterestProcessor: ptprocessor.New(mongoRepo, sink, notifier, pusherClient, log, cfg),
		gmbProcessor:       gmbprocessor.New(mongoRepo, sink, notifier, pusherClient, log, cfg),
		metaAdsProcessor:   metaadsprocessor.New(mongoRepo, sink, notifier, pusherClient, log, cfg),
		mongoRepo:          mongoRepo,
		logger:             log,
	}

	// =========================================================================
	// KAFKA CONSUMERS - One per platform using existing consumer groups
	// =========================================================================

	facebookConsumer, err := kafka2.NewConsumer(cfg.Kafka, facebookConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Facebook Kafka consumer")
	}
	defer facebookConsumer.Close()

	instagramConsumer, err := kafka2.NewConsumer(cfg.Kafka, instagramConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Instagram Kafka consumer")
	}
	defer instagramConsumer.Close()

	linkedinConsumer, err := kafka2.NewConsumer(cfg.Kafka, linkedinConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create LinkedIn Kafka consumer")
	}
	defer linkedinConsumer.Close()

	youtubeConsumer, err := kafka2.NewConsumer(cfg.Kafka, youtubeConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create YouTube Kafka consumer")
	}
	defer youtubeConsumer.Close()

	tiktokConsumer, err := kafka2.NewConsumer(cfg.Kafka, tiktokConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create TikTok Kafka consumer")
	}
	defer tiktokConsumer.Close()

	twitterConsumer, err := kafka2.NewConsumer(cfg.Kafka, twitterConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Twitter Kafka consumer")
	}
	defer twitterConsumer.Close()

	pinterestConsumer, err := kafka2.NewConsumer(cfg.Kafka, pinterestConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Pinterest Kafka consumer")
	}
	defer pinterestConsumer.Close()

	gmbConsumer, err := kafka2.NewConsumer(cfg.Kafka, gmbConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create GMB Kafka consumer")
	}
	defer gmbConsumer.Close()

	metaAdsConsumer, err := kafka2.NewConsumer(cfg.Kafka, metaAdsConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Meta Ads Kafka consumer")
	}
	defer metaAdsConsumer.Close()

	// =========================================================================
	// GRACEFUL SHUTDOWN
	// =========================================================================

	runCtx, runCancel := context.WithCancel(context.Background())
	defer runCancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received, waiting for workers to finish...")
		runCancel()
	}()

	// =========================================================================
	// INITIALIZE QUEUES
	// =========================================================================

	// Global queue for admission control
	globalQueue := NewGlobalQueue(GlobalQueueCapacity)

	// Per-platform queues for parallel processing
	platformJobs := NewPlatformJobChannels()

	// =========================================================================
	// START WORKER POOLS (one pool per platform, running in parallel)
	// =========================================================================

	var wg sync.WaitGroup

	for platform, pcfg := range PlatformSettings {
		workerCount := int(float64(pcfg.Workers) * *workerMultiplier)
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

		// Start workers as goroutines - each worker processes from the same channel
		// This enables parallel processing within each platform
		for i := 0; i < workerCount; i++ {
			wg.Add(1)
			go func(p string, workerID int, jobs <-chan ImmediateWorkOrder) {
				defer wg.Done()
				processor.platformWorker(runCtx, p, workerID, jobs, platformJobs, globalQueue)
			}(platform, i, jobChan)
		}
	}

	// =========================================================================
	// STATS LOGGER (logs queue statistics every 30 seconds)
	// =========================================================================

	go func() {
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case <-ticker.C:
				// Log global queue stats
				current, capacity, admitted, rejected := globalQueue.Stats()
				log.Info().
					Int64("global_current", current).
					Int64("global_capacity", capacity).
					Int64("global_admitted", admitted).
					Int64("global_rejected", rejected).
					Float64("global_utilization_pct", float64(current)/float64(capacity)*100).
					Msg("Global queue stats")

				// Log per-platform stats
				stats := platformJobs.GetStats()
				for platform, s := range stats {
					log.Info().
						Str("platform", platform).
						Int("queue_depth", s.QueueDepth).
						Int("queue_capacity", s.MaxCapacity).
						Int64("processed", s.Processed).
						Int64("dropped", s.Dropped).
						Float64("utilization_pct", float64(s.QueueDepth)/float64(s.MaxCapacity)*100).
						Msg("Platform queue stats")
				}
			}
		}
	}()

	// =========================================================================
	// KAFKA CONSUMERS - Routes messages through global queue to platform queues
	// Each platform has its own consumer using existing consumer groups
	// =========================================================================

	// Message handler shared by all consumers
	handleMessage := func(topic string, value []byte) {
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

		// STEP 1: Check global queue capacity (admission control)
		if !globalQueue.TryAdmit() {
			current, capacity, _, rejected := globalQueue.Stats()
			log.Warn().
				Str("platform", wo.Platform).
				Str("account_id", wo.AccountID).
				Str("workspace_id", wo.WorkspaceID).
				Int64("global_current", current).
				Int64("global_capacity", capacity).
				Int64("global_rejected", rejected).
				Msg("SYSTEM BUSY - Global queue full, work order rejected. User should retry later.")
			return
		}

		// STEP 2: Route to platform-specific queue
		if !platformJobs.TryEnqueue(wo.Platform, wo) {
			globalQueue.Release()
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
	}

	// Start consumer for each platform
	var consumerWg sync.WaitGroup

	// Facebook consumer
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		topic := "immediate-work-order-facebook"
		log.Info().Str("topic", topic).Str("group", facebookConsumerGroup).Msg("Starting Facebook consumer")
		err := facebookConsumer.Consume(runCtx, []string{topic}, func(ctx context.Context, t string, key, value []byte) error {
			handleMessage(t, value)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Facebook consumer error")
		}
	}()

	// Instagram consumer
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		topic := "immediate-work-order-instagram"
		log.Info().Str("topic", topic).Str("group", instagramConsumerGroup).Msg("Starting Instagram consumer")
		err := instagramConsumer.Consume(runCtx, []string{topic}, func(ctx context.Context, t string, key, value []byte) error {
			handleMessage(t, value)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Instagram consumer error")
		}
	}()

	// LinkedIn consumer
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		topic := "immediate-work-order-linkedin"
		log.Info().Str("topic", topic).Str("group", linkedinConsumerGroup).Msg("Starting LinkedIn consumer")
		err := linkedinConsumer.Consume(runCtx, []string{topic}, func(ctx context.Context, t string, key, value []byte) error {
			handleMessage(t, value)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("LinkedIn consumer error")
		}
	}()

	// YouTube consumer
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		topic := "immediate-work-order-youtube"
		log.Info().Str("topic", topic).Str("group", youtubeConsumerGroup).Msg("Starting YouTube consumer")
		err := youtubeConsumer.Consume(runCtx, []string{topic}, func(ctx context.Context, t string, key, value []byte) error {
			handleMessage(t, value)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("YouTube consumer error")
		}
	}()

	// TikTok consumer
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		topic := "immediate-work-order-tiktok"
		log.Info().Str("topic", topic).Str("group", tiktokConsumerGroup).Msg("Starting TikTok consumer")
		err := tiktokConsumer.Consume(runCtx, []string{topic}, func(ctx context.Context, t string, key, value []byte) error {
			handleMessage(t, value)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("TikTok consumer error")
		}
	}()

	// Twitter consumer
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		topic := "immediate-work-order-twitter"
		log.Info().Str("topic", topic).Str("group", twitterConsumerGroup).Msg("Starting Twitter consumer")
		err := twitterConsumer.Consume(runCtx, []string{topic}, func(ctx context.Context, t string, key, value []byte) error {
			handleMessage(t, value)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Twitter consumer error")
		}
	}()

	// Pinterest consumer
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		topic := "immediate-work-order-pinterest"
		log.Info().Str("topic", topic).Str("group", pinterestConsumerGroup).Msg("Starting Pinterest consumer")
		err := pinterestConsumer.Consume(runCtx, []string{topic}, func(ctx context.Context, t string, key, value []byte) error {
			handleMessage(t, value)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Pinterest consumer error")
		}
	}()

	// GMB consumer
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		topic := "immediate-work-order-gmb"
		log.Info().Str("topic", topic).Str("group", gmbConsumerGroup).Msg("Starting GMB consumer")
		err := gmbConsumer.Consume(runCtx, []string{topic}, func(ctx context.Context, t string, key, value []byte) error {
			handleMessage(t, value)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("GMB consumer error")
		}
	}()

	// Meta Ads consumer
	consumerWg.Add(1)
	go func() {
		defer consumerWg.Done()
		topic := "immediate-work-order-meta-ads"
		log.Info().Str("topic", topic).Str("group", metaAdsConsumerGroup).Msg("Starting Meta Ads consumer")
		err := metaAdsConsumer.Consume(runCtx, []string{topic}, func(ctx context.Context, t string, key, value []byte) error {
			handleMessage(t, value)
			return nil
		})
		if err != nil && err != context.Canceled {
			log.Error().Err(err).Msg("Meta Ads consumer error")
		}
	}()

	// Wait for consumers to finish, then close platform channels
	go func() {
		consumerWg.Wait()
		platformJobs.CloseAll()
	}()

	// Wait for all workers to finish
	wg.Wait()
	log.Info().Msg("Unified Immediate Processor stopped")
}

// =============================================================================
// WORKER - Processes work orders from platform queue
// =============================================================================

// platformWorker processes work orders for a specific platform.
// Multiple workers run concurrently for each platform, enabling parallel processing.
// Workers are long-lived goroutines that read from the platform's job channel.
func (p *UnifiedProcessor) platformWorker(
	ctx context.Context,
	platform string,
	workerID int,
	jobs <-chan ImmediateWorkOrder,
	platformJobs *PlatformJobChannels,
	globalQueue *GlobalQueue,
) {
	log := p.logger.Logger.With().Str("platform", platform).Int("worker_id", workerID).Logger()
	log.Debug().Msg("Worker started")

	processedCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Info().Int("processed", processedCount).Msg("Worker stopped (context cancelled)")
			return

		case wo, ok := <-jobs:
			if !ok {
				log.Info().Int("processed", processedCount).Msg("Worker stopped (channel closed)")
				return
			}

			startTime := time.Now()
			log.Info().
				Str("account_id", wo.AccountID).
				Str("platform_identifier", wo.AccountID).
				Str("workspace_id", wo.WorkspaceID).
				Str("sync_type", wo.SyncType).
				Str("start_date", wo.StartDate).
				Str("end_date", wo.EndDate).
				Int("n_tweets", wo.PostCount).
				Msg("Processing work order")

			// Process based on platform
			var err error
			switch platform {
			case "facebook":
				err = p.facebookProcessor.ProcessAccount(ctx, fbprocessor.WorkOrder{
					ID:              wo.ID,
					AccountID:       wo.AccountID,
					Type:            wo.Type,
					AccessToken:     wo.AccessToken,
					WorkspaceID:     wo.WorkspaceID,
					LongAccessToken: wo.LongAccessToken,
					SyncType:        wo.SyncType,
					StartDate:       wo.StartDate,
					EndDate:         wo.EndDate,
				})
			case "instagram":
				err = p.instagramProcessor.ProcessAccount(ctx, igprocessor.WorkOrder{
					ID:                    wo.ID,
					AccountID:             wo.AccountID,
					Type:                  wo.Type,
					AccessToken:           wo.AccessToken,
					WorkspaceID:           wo.WorkspaceID,
					SyncType:              wo.SyncType,
					ConnectedViaInstagram: wo.ConnectedViaInstagram,
					StartDate:             wo.StartDate,
					EndDate:               wo.EndDate,
				})
			case "linkedin":
				err = p.linkedinProcessor.ProcessAccount(ctx, liprocessor.WorkOrder{
					ID:          wo.ID,
					AccountID:   wo.AccountID,
					AccessToken: wo.AccessToken,
					WorkspaceID: wo.WorkspaceID,
					SyncType:    wo.SyncType,
					StartDate:   wo.StartDate,
					EndDate:     wo.EndDate,
				})
			case "youtube":
				err = p.youtubeProcessor.ProcessAccount(ctx, ytprocessor.WorkOrder{
					ID:           wo.ID,
					ChannelID:    wo.AccountID,
					AccessToken:  wo.AccessToken,
					RefreshToken: wo.RefreshToken,
					WorkspaceID:  wo.WorkspaceID,
					SyncType:     wo.SyncType,
					StartDate:    wo.StartDate,
					EndDate:      wo.EndDate,
				})
			case "tiktok":
				err = p.tiktokProcessor.ProcessAccount(ctx, tkprocessor.ImmediateWorkOrder{
					ID:           wo.ID,
					WorkspaceID:  wo.WorkspaceID,
					TikTokID:     wo.AccountID,
					AccessToken:  wo.AccessToken,
					RefreshToken: wo.RefreshToken,
					SyncType:     wo.SyncType,
					StartDate:    wo.StartDate,
					EndDate:      wo.EndDate,
				})
			case "twitter":
				twitterID := wo.AccountID
				if twitterID == "" {
					twitterID = wo.TwitterID
				}
				oauthToken := wo.AccessToken
				if oauthToken == "" {
					oauthToken = wo.OAuthToken
				}
				oauthTokenSecret := wo.RefreshToken
				if oauthTokenSecret == "" {
					oauthTokenSecret = wo.OAuthTokenSecret
				}
				err = p.twitterProcessor.ProcessAccount(ctx, twprocessor.ImmediateWorkOrder{
					ID:               wo.ID,
					WorkspaceID:      wo.WorkspaceID,
					TwitterID:        twitterID,
					OAuthToken:       oauthToken,
					OAuthTokenSecret: oauthTokenSecret,
					NTweets:          wo.PostCount,
					APIKey:           wo.APIKey,
					APISecret:        wo.APISecret,
					AppName:          wo.AppName,
					AppID:            wo.AppID,
					ExecutedBy:       wo.ExecutedBy,
					SyncType:         wo.SyncType,
				})
			case "pinterest":
				err = p.pinterestProcessor.ProcessAccount(ctx, ptprocessor.WorkOrder{
					ID:          wo.ID,
					AccountID:   wo.AccountID,
					AccessToken: wo.AccessToken,
					AccountType: wo.Type,
					WorkspaceID: wo.WorkspaceID,
					SyncType:    wo.SyncType,
					StartDate:   wo.StartDate,
					EndDate:     wo.EndDate,
				})
			case "gmb":
				err = p.gmbProcessor.ProcessAccount(ctx, gmbprocessor.ImmediateWorkOrder{
					ID:           wo.ID,
					WorkspaceID:  wo.WorkspaceID,
					AccountID:    wo.AccountID,
					LocationID:   wo.LocationID,
					AccessToken:  wo.AccessToken,
					RefreshToken: wo.RefreshToken,
					AccountName:  wo.AccountName,
					LocationName: wo.LocationName,
					LanguageCode: wo.LanguageCode,
					SyncType:     wo.SyncType,
					StartDate:    wo.StartDate,
					EndDate:      wo.EndDate,
				})
			case "meta_ads":
				err = p.metaAdsProcessor.ProcessAccount(ctx, kafkamodels.MetaAdsWorkOrder{
					MongoID:            wo.ID,
					PlatformIdentifier: wo.AccountID,
					AccountID:          wo.AccountID,
					AccessToken:        wo.AccessToken,
					LongAccessToken:    wo.LongAccessToken,
					WorkspaceID:        wo.WorkspaceID,
					UserID:             wo.ID,
					SyncType:           wo.SyncType,
					StartDate:          wo.StartDate,
					EndDate:            wo.EndDate,
				})
			default:
				log.Warn().Msg("No processor implemented for platform")
			}

			// IMPORTANT: Release from global queue after processing completes
			globalQueue.Release()

			if err != nil {
				// Suppress logging for expected/auth errors (they're operational, not exceptional)
				if isExpectedError(err) {
					// Expected auth/permission errors - just skip this account, no logging needed
				} else {
					log.Error().
						Err(err).
						Str("error_message", err.Error()).
						Str("account_id", wo.AccountID).
						Str("workspace_id", wo.WorkspaceID).
						Str("sync_type", wo.SyncType).
						Str("function", "platformWorker").
						Str("stage", "process_account").
						Dur("duration", time.Since(startTime)).
						Msg("Failed to process work order")
				}
			} else {
				processedCount++
				platformJobs.IncrementProcessed(platform)
				log.Info().
					Str("account_id", wo.AccountID).
					Dur("duration", time.Since(startTime)).
					Msg("Work order completed")
			}

		}
	}
}

// =============================================================================
// HELPERS
// =============================================================================

// inferPlatformFromTopic extracts the platform name from Kafka topic
func inferPlatformFromTopic(topic string) string {
	switch {
	case strings.Contains(topic, "facebook"):
		return "facebook"
	case strings.Contains(topic, "instagram"):
		return "instagram"
	case strings.Contains(topic, "linkedin"):
		return "linkedin"
	case strings.Contains(topic, "youtube"):
		return "youtube"
	case strings.Contains(topic, "tiktok"):
		return "tiktok"
	case strings.Contains(topic, "twitter"):
		return "twitter"
	case strings.Contains(topic, "pinterest"):
		return "pinterest"
	case strings.Contains(topic, "gmb"):
		return "gmb"
	case strings.Contains(topic, "meta-ads"), strings.Contains(topic, "meta_ads"):
		return "meta_ads"
	default:
		return ""
	}
}

// mustCreateProducer creates a Kafka producer or panics on failure
func mustCreateProducer(kafkaCfg config.KafkaConfig, log zerolog.Logger) kafka2.Producer {
	producer, err := kafka2.NewProducer(kafkaCfg, log)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	return producer
}

// NewUnifiedProcessor creates a new UnifiedProcessor with the given processors
// This constructor enables dependency injection for testing
func NewUnifiedProcessor(
	facebookProcessor FacebookProcessor,
	instagramProcessor InstagramProcessor,
	linkedinProcessor LinkedInProcessor,
	youtubeProcessor YouTubeProcessor,
	tiktokProcessor TikTokProcessor,
	twitterProcessor TwitterProcessor,
	pinterestProcessor PinterestProcessor,
	gmbProcessor GMBProcessor,
	log *logger.Logger,
) *UnifiedProcessor {
	return &UnifiedProcessor{
		facebookProcessor:  facebookProcessor,
		instagramProcessor: instagramProcessor,
		linkedinProcessor:  linkedinProcessor,
		youtubeProcessor:   youtubeProcessor,
		tiktokProcessor:    tiktokProcessor,
		twitterProcessor:   twitterProcessor,
		pinterestProcessor: pinterestProcessor,
		gmbProcessor:       gmbProcessor,
		logger:             log,
	}
}

// PlatformWorkerTestable is a testable version of platformWorker
func (p *UnifiedProcessor) PlatformWorkerTestable(
	ctx context.Context,
	platform string,
	workerID int,
	jobs <-chan ImmediateWorkOrder,
	platformJobs *PlatformJobChannels,
	globalQueue *GlobalQueue,
) {
	p.platformWorker(ctx, platform, workerID, jobs, platformJobs, globalQueue)
}

// HandleMessage processes a single message from Kafka - extracted for testing
func HandleMessage(topic string, value []byte, globalQueue *GlobalQueue, platformJobs *PlatformJobChannels, log *logger.Logger) {
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

	// STEP 1: Check global queue capacity (admission control)
	if !globalQueue.TryAdmit() {
		_, _, _, rejected := globalQueue.Stats()
		log.Warn().
			Str("platform", wo.Platform).
			Str("account_id", wo.AccountID).
			Int64("global_rejected", rejected).
			Msg("Global queue full, work order rejected")
		return
	}

	// STEP 2: Route to platform-specific queue
	if !platformJobs.TryEnqueue(wo.Platform, wo) {
		globalQueue.Release()
		stats := platformJobs.GetStats()
		s := stats[wo.Platform]
		log.Warn().
			Str("platform", wo.Platform).
			Int("queue_depth", s.QueueDepth).
			Int64("dropped", s.Dropped).
			Msg("Platform queue full, work order dropped")
	}
}

// CalculateUtilization calculates utilization percentage
func CalculateUtilization(current, capacity int64) float64 {
	if capacity == 0 {
		return 0
	}
	return float64(current) / float64(capacity) * 100
}

// CalculateQueueUtilization calculates utilization for a queue
func CalculateQueueUtilization(depth, capacity int) float64 {
	if capacity == 0 {
		return 0
	}
	return float64(depth) / float64(capacity) * 100
}

// GetTotalWorkerCount calculates total workers across all platforms
func GetTotalWorkerCount(multiplier float64) int {
	total := 0
	for _, pcfg := range PlatformSettings {
		workers := int(float64(pcfg.Workers) * multiplier)
		if workers < 1 {
			workers = 1
		}
		total += workers
	}
	return total
}

// GetPlatformWorkerCount calculates workers for a specific platform
func GetPlatformWorkerCount(platform string, multiplier float64) int {
	pcfg, ok := PlatformSettings[platform]
	if !ok {
		return 1
	}
	workers := int(float64(pcfg.Workers) * multiplier)
	if workers < 1 {
		workers = 1
	}
	return workers
}

// ValidatePlatform checks if a platform is supported
func ValidatePlatform(platform string) bool {
	_, ok := PlatformSettings[platform]
	return ok
}

// GetSupportedPlatforms returns list of supported platforms
func GetSupportedPlatforms() []string {
	platforms := make([]string, 0, len(PlatformSettings))
	for p := range PlatformSettings {
		platforms = append(platforms, p)
	}
	return platforms
}

// ParseWorkOrder parses a work order from JSON bytes
func ParseWorkOrder(data []byte) (*ImmediateWorkOrder, error) {
	var wo ImmediateWorkOrder
	if err := json.Unmarshal(data, &wo); err != nil {
		return nil, err
	}
	return &wo, nil
}

// ResolvePlatform determines the platform from work order or topic
func ResolvePlatform(wo *ImmediateWorkOrder, topic string) string {
	if wo.Platform != "" {
		return wo.Platform
	}
	return inferPlatformFromTopic(topic)
}

// isExpectedError checks if an error is an expected/operational error (auth, permission, etc)
// These errors do not warrant logging or Sentry alerting as they're expected when:
// - Access tokens expire
// - User revokes permissions
// - Account credentials become invalid
// - Platform-specific business logic (e.g., insufficient viewers for analytics)
func isExpectedError(err error) bool {
	if err == nil {
		return false
	}

	// Check for specific sentinel errors (YouTube)
	if errors.Is(err, ytprocessor.ErrUnauthorized) {
		return true
	}

	// Check error message for common auth/permission patterns
	errStr := strings.ToLower(err.Error())
	authPatterns := []string{
		"unauthorized",
		"invalid.*credential",
		"invalid.*token",
		"access token",
		"token.*expired",
		"expired token",
		"permission",
		"auth",
		"401",
		"403",
		"forbidden",
		"unauthenticated",
		"revoked",
	}

	for _, pattern := range authPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}

	return false
}
