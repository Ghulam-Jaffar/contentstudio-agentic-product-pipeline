package main

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// ErrUnauthorized is returned when the API returns a 401 status
var ErrUnauthorized = errors.New("unauthorized: invalid or expired token")

// Per-account concurrency guard - ensures we don't run multiple pipelines for the same channel at once
var accountSemaphores sync.Map // map[string]*semaphore.Weighted

func semForAccount(id string, capacity int64) *semaphore.Weighted {
	if v, ok := accountSemaphores.Load(id); ok {
		return v.(*semaphore.Weighted)
	}
	sem := semaphore.NewWeighted(capacity)
	if old, loaded := accountSemaphores.LoadOrStore(id, sem); loaded {
		return old.(*semaphore.Weighted)
	}
	return sem
}

// isUnauthorizedError checks if the error indicates a 401 Unauthorized response
func isUnauthorizedError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "status 401") || strings.Contains(errStr, "request failed with status 401")
}

const (
	maxWorkers          = 10
	workOrderChanSize   = 200
	timestampChanSize   = 500
	consumerGroup       = "youtube-fetcher-group"
	topicWorkOrderBatch = "work-order-youtube"

	// maxConcurrentAccounts is the max number of YouTube channels processed simultaneously.
	maxConcurrentAccounts = 50

	// Output topics for raw data
	topicRawChannels         = "raw-youtube-channels"
	topicRawVideos           = "raw-youtube-videos"
	topicRawActivityInsights = "raw-youtube-activity-insights"
	topicRawTrafficInsights  = "raw-youtube-traffic-insights"
	topicRawSharedInsights   = "raw-youtube-shared-insights"

	idleTimeout = 15 * time.Minute

	// Date ranges for different sync types
	incrementalVideosDays   = 14 // Daily scheduler
	incrementalInsightsDays = 14
	immediateVideosDays     = 90 // Immediate processor
	immediateInsightsDays   = 90
	fullSyncVideosDays      = 365 // Full sync
	fullSyncInsightsDays    = 365
)

type WorkOrderMessage struct {
	AccountID   string
	ChannelID   string
	Value       []byte
	AccessToken string
}

type TimestampUpdateRequest struct {
	AccountID string
	ChannelID string
}

// Service holds all dependencies for the YouTube fetcher service.
// This allows for dependency injection in tests.
type Service struct {
	totalProcessed        int64
	totalFailed           int64
	YTClient              social.YouTubeAPI
	Producer              kafka2.Producer
	Consumer              kafka2.Consumer
	MongoRepo             mongodb.UnifiedSocialRepository
	Logger                *logger.Logger
	DecryptionKey         string
	MaxWorkers            int
	MaxConcurrentAccounts int
	IdleTimeout           time.Duration
	IdleCheckPeriod       time.Duration
}

// NewService creates a new Service with the given dependencies.
func NewService(
	ytClient social.YouTubeAPI,
	producer kafka2.Producer,
	consumer kafka2.Consumer,
	mongoRepo mongodb.UnifiedSocialRepository,
	log *logger.Logger,
	decryptionKey string,
) *Service {
	return &Service{
		YTClient:              ytClient,
		Producer:              producer,
		Consumer:              consumer,
		MongoRepo:             mongoRepo,
		Logger:                log,
		DecryptionKey:         decryptionKey,
		MaxWorkers:            maxWorkers,
		MaxConcurrentAccounts: maxConcurrentAccounts,
		IdleTimeout:           idleTimeout,
		IdleCheckPeriod:       30 * time.Second,
	}
}

// Run starts the service and blocks until shutdown.
func (s *Service) Run(ctx context.Context) error {
	timestampUpdateChan := make(chan TimestampUpdateRequest, timestampChanSize)

	var lastMessageTime int64 = time.Now().UnixNano()

	maxConc := s.MaxConcurrentAccounts
	if maxConc <= 0 {
		maxConc = maxConcurrentAccounts
	}
	accountSem := semaphore.NewWeighted(int64(maxConc))
	var dispatchWg sync.WaitGroup

	var wg sync.WaitGroup
	s.startTimestampUpdater(ctx, &wg, timestampUpdateChan)
	s.startBatchConsumer(ctx, accountSem, &dispatchWg, timestampUpdateChan, &lastMessageTime)

	s.Logger.Info().Int("max_concurrent_accounts", maxConc).Msg("YouTube Fetcher service is running")

	<-ctx.Done()
	s.Logger.Info().Msg("Context cancelled, stopping service...")

	dispatchWg.Wait()

	s.Logger.Info().
		Int64("total_processed", atomic.LoadInt64(&s.totalProcessed)).
		Int64("total_failed", atomic.LoadInt64(&s.totalFailed)).
		Msg("YouTube Fetcher service stopped")

	close(timestampUpdateChan)
	wg.Wait()

	return nil
}

func (s *Service) startWorkerPool(
	ctx context.Context,
	wg *sync.WaitGroup,
	workOrderChan chan WorkOrderMessage,
	timestampUpdateChan chan TimestampUpdateRequest,
) {
	for i := 0; i < s.MaxWorkers; i++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			s.workOrderProcessor(ctx, workerID, workOrderChan, timestampUpdateChan)
		}(i)
	}
	s.Logger.Info().Int("workers", s.MaxWorkers).Msg("Started YouTube fetcher workers")
}

func (s *Service) startBatchConsumer(
	ctx context.Context,
	accountSem *semaphore.Weighted,
	dispatchWg *sync.WaitGroup,
	timestampUpdateChan chan TimestampUpdateRequest,
	lastMessageTime *int64,
) {
	go func() {
		s.Logger.Info().Str("topic", topicWorkOrderBatch).Msg("Starting batch Kafka consumer")
		err := s.Consumer.Consume(ctx, []string{topicWorkOrderBatch}, func(ctx context.Context, topic string, key, value []byte) error {
			if lastMessageTime != nil {
				atomic.StoreInt64(lastMessageTime, time.Now().UnixNano())
			}

			var batch kafkamodels.YouTubeBatchWorkOrder
			if err := json.Unmarshal(value, &batch); err != nil {
				s.Logger.Error().Err(err).Str("function", "startBatchConsumer").Str("stage", "unmarshal_batch_work_order").Msg("Failed to unmarshal batch work order")
				return nil
			}

			total := len(batch.Accounts)
			s.Logger.Info().
				Str("batch_id", batch.BatchID).
				Int("accounts", total).
				Msg("Received batch work order, dispatching goroutines")

			var batchWg sync.WaitGroup
			var batchProcessed, batchFailed int64

			for _, account := range batch.Accounts {
				acc := account
				accountPayload, err := json.Marshal(acc)
				if err != nil {
					s.Logger.Error().Err(err).Str("channel_id", acc.ChannelID).Str("function", "startBatchConsumer").Str("stage", "marshal_account_work_order").Msg("Failed to marshal account work order")
					atomic.AddInt64(&batchFailed, 1)
					continue
				}
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
					msg := WorkOrderMessage{
						AccountID:   acc.ID,
						ChannelID:   acc.ChannelID,
						Value:       accountPayload,
						AccessToken: acc.AccessToken,
					}
					processWorkOrder(ctx, msg, s.YTClient, s.Producer, s.MongoRepo, s.DecryptionKey, s.Logger, timestampUpdateChan)
					atomic.AddInt64(&batchProcessed, 1)
				}()
			}

			batchID := batch.BatchID
			go func() {
				batchWg.Wait()
				p := atomic.LoadInt64(&batchProcessed)
				f := atomic.LoadInt64(&batchFailed)
				atomic.AddInt64(&s.totalProcessed, p)
				atomic.AddInt64(&s.totalFailed, f)
				s.Logger.Info().
					Str("batch_id", batchID).
					Int("total", total).
					Int64("processed", p).
					Int64("failed", f).
					Msg("Batch processing complete")
			}()

			return nil
		})

		if err != nil && err != context.Canceled {
			s.Logger.Error().Err(err).Str("function", "startBatchConsumer").Str("stage", "consume_batch").Msg("Batch consumer error")
		}
	}()
}

func (s *Service) startTimestampUpdater(
	ctx context.Context,
	wg *sync.WaitGroup,
	timestampUpdateChan <-chan TimestampUpdateRequest,
) {
	wg.Add(1)
	go func() {
		defer wg.Done()
		s.Logger.Info().Msg("Timestamp updater started")

		for {
			select {
			case req, ok := <-timestampUpdateChan:
				if !ok {
					s.Logger.Info().Msg("Timestamp update channel closed")
					return
				}

				objectID, err := primitive.ObjectIDFromHex(req.AccountID)
				if err != nil {
					s.Logger.Error().Err(err).Str("error_message", err.Error()).Str("channel_id", req.ChannelID).Str("function", "startTimestampUpdater").Str("stage", "parse_object_id").Msg("Invalid ObjectID for timestamp update")
					continue
				}

				now := time.Now().UTC()
				if err := s.MongoRepo.UpdateState(context.Background(), objectID, mongomodels.StateProcessed); err != nil {
					s.Logger.Error().Err(err).Str("error_message", err.Error()).Str("channel_id", req.ChannelID).Str("function", "startTimestampUpdater").Str("stage", "update_account_state").Msg("Failed to update account state to Processed")
					continue
				}
				if err := s.MongoRepo.UpdateAnalyticsTimestamp(context.Background(), objectID, "analytics", now); err != nil {
					s.Logger.Error().Err(err).Str("error_message", err.Error()).Str("channel_id", req.ChannelID).Str("function", "startTimestampUpdater").Str("stage", "update_analytics_timestamp").Msg("Failed to update analytics timestamp")
				} else {
					s.MongoRepo.ClearProcessingError(context.Background(), objectID)
					s.Logger.Debug().Str("channel_id", req.ChannelID).Msg("Updated account state and analytics timestamp")
				}

			case <-ctx.Done():
				s.Logger.Info().Msg("Timestamp updater stopping")
				return
			}
		}
	}()
}

func (s *Service) workOrderProcessor(
	ctx context.Context,
	workerID int,
	workOrderChan <-chan WorkOrderMessage,
	timestampUpdateChan chan<- TimestampUpdateRequest,
) {
	workerLog := s.Logger.With().Int("worker_id", workerID).Logger()
	workerLog.Info().Msg("Worker started")

	for {
		select {
		case <-ctx.Done():
			workerLog.Info().Msg("Worker stopped (context cancelled)")
			return
		case msg, ok := <-workOrderChan:
			if !ok {
				workerLog.Info().Msg("Worker stopped (channel closed)")
				return
			}
			processWorkOrder(ctx, msg, s.YTClient, s.Producer, s.MongoRepo, s.DecryptionKey, &logger.Logger{Logger: workerLog}, timestampUpdateChan)
		}
	}
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("workers", maxWorkers).
		Str("consumer_group", consumerGroup).
		Msg("Starting YouTube Fetcher service")

	ytClient := social.NewYouTubeClient(cfg.YouTube.ClientID, cfg.YouTube.ClientSecret)

	mongoClient, mongoRepo := initMongoDB(cfg, log)
	defer mongoClient.Disconnect(context.Background())

	consumer, producer := initKafkaClients(cfg, log)
	defer consumer.Close()
	defer producer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	svc := NewService(ytClient, producer, consumer, mongoRepo, log, cfg.DecryptionKey)

	done := make(chan struct{})
	go func() {
		defer close(done)
		svc.Run(ctx)
	}()

	select {
	case <-sigChan:
		log.Info().Msg("Shutdown signal received, stopping service...")
	case <-done:
		log.Info().Msg("Service stopped")
	}

	cancel()
	<-done

	log.Info().Msg("YouTube Fetcher service stopped")
}

func initMongoDB(cfg *config.Config, log *logger.Logger) (*mongo.Client, mongodb.UnifiedSocialRepository) {
	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	clientOpts := options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential)

	mongoClient, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}

	db := mongoClient.Database(cfg.Mongo.Database)
	repo := mongodb.NewUnifiedSocialRepository(db, log.Logger)

	log.Info().Msg("MongoDB connected for timestamp updates")
	return mongoClient, repo
}

func initKafkaClients(cfg *config.Config, log *logger.Logger) (kafka2.Consumer, kafka2.Producer) {
	consumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}

	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}

	return consumer, producer
}

func processWorkOrder(
	ctx context.Context,
	msg WorkOrderMessage,
	ytClient social.YouTubeAPI,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	decryptionKey string,
	log *logger.Logger,
	timestampUpdateChan chan<- TimestampUpdateRequest,
) {
	var wo kafkamodels.YouTubeAccountWorkOrder
	if err := json.Unmarshal(msg.Value, &wo); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("function", "processWorkOrder").Str("stage", "unmarshal_work_order").Msg("Failed to unmarshal work order")
		return
	}

	if wo.ChannelID == "" || wo.AccessToken == "" {
		log.Warn().Str("channel_id", wo.ChannelID).Msg("Skipping work order with missing channel_id or access_token")
		if wo.ID != "" {
			if accountID, parseErr := primitive.ObjectIDFromHex(wo.ID); parseErr == nil {
				mongoRepo.RecordProcessingError(ctx, accountID, "Access token is empty or channel ID is missing")
			}
		}
		return
	}

	// Per-account concurrency gate - prevent multiple simultaneous fetches for same channel
	sem := semForAccount(wo.ChannelID, 1)
	if err := sem.Acquire(ctx, 1); err != nil {
		log.Warn().Err(err).Str("channel_id", wo.ChannelID).Msg("Failed to acquire semaphore")
		return
	}
	defer sem.Release(1)

	accessToken := wo.AccessToken
	if decrypted, err := crypto.DecryptToken(accessToken, decryptionKey); err == nil {
		accessToken = decrypted
	}

	// Refresh token if needed - YouTube tokens expire in 1 hour
	if wo.RefreshToken != "" {
		refreshToken := wo.RefreshToken
		if decrypted, err := crypto.DecryptToken(refreshToken, decryptionKey); err == nil {
			refreshToken = decrypted
		}
		if tokenResp, err := ytClient.RefreshToken(ctx, refreshToken); err == nil {
			accessToken = tokenResp.AccessToken
			log.Info().Str("channel_id", wo.ChannelID).Msg("Successfully refreshed access token")
		} else {
			log.Warn().Err(err).Str("channel_id", wo.ChannelID).Msg("Token refresh failed, using existing token")
		}
	}

	log.Info().Str("channel_id", wo.ChannelID).Str("sync_type", wo.SyncType).Msg("Processing YouTube account")

	now := time.Now().UTC()
	// YouTube Analytics API has a 2-3 day data delay, so use 3 days ago as end date
	// This ensures we get complete data for the requested period
	endDate := now.AddDate(0, 0, -3)

	// Determine date ranges based on sync type for both videos and insights
	var videosSince time.Time
	var insightsDays int
	switch wo.SyncType {
	case kafkamodels.YouTubeSyncTypeFullSync:
		videosSince = endDate.AddDate(0, 0, -fullSyncVideosDays)
		insightsDays = fullSyncInsightsDays
	case kafkamodels.YouTubeSyncTypeImmediate:
		videosSince = endDate.AddDate(0, 0, -immediateVideosDays)
		insightsDays = immediateInsightsDays
	default:
		videosSince = endDate.AddDate(0, 0, -incrementalVideosDays)
		insightsDays = incrementalInsightsDays
	}
	// Calculate insights start date based on sync type (e.g., -13 for 14 days inclusive)
	insightsStartDate := endDate.AddDate(0, 0, -(insightsDays - 1))

	log.Info().
		Str("channel_id", wo.ChannelID).
		Str("sync_type", wo.SyncType).
		Time("insights_start_date", insightsStartDate).
		Time("insights_end_date", endDate).
		Time("videos_since", videosSince).
		Msg("Date range for YouTube analytics")

	// ==================== PARALLEL DATA FETCHING ====================
	// Phase 1: Fetch channel + insights in parallel.
	// Phase 2: Fetch videos using uploads playlist ID from channel data.
	// If any API returns 401, cancel all other requests immediately.
	// ================================================================

	var (
		channelData      *social.YouTubeChannelItem
		activities       []social.YouTubeActivityItem
		videoDetailsMap  map[string]*social.YouTubeVideoItem
		activityInsights *social.YouTubeAnalyticsResponse
		trafficInsights  *social.YouTubeAnalyticsResponse
		sharedInsights   *social.YouTubeAnalyticsResponse
		activityMu       sync.Mutex
		trafficMu        sync.Mutex
		sharedMu         sync.Mutex
	)

	// All API calls run in a single errgroup. Videos wait for channel data
	// via a signal channel, so they run in parallel with insights.
	eg, egCtx := errgroup.WithContext(ctx)
	channelReady := make(chan struct{})

	// Fetch channel data (signals when done so video goroutine can start)
	eg.Go(func() error {
		defer close(channelReady)
		channelResp, err := ytClient.FetchChannels(egCtx, accessToken)
		if err != nil {
			if isUnauthorizedError(err) {
				log.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "fetch_channels").Msg("Unauthorized - stopping all API calls")
				return ErrUnauthorized
			}
			log.Error().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "fetch_channels").Msg("Failed to fetch channel data")
			return nil
		}
		if len(channelResp.Items) > 0 {
			channelData = &channelResp.Items[0]
		}
		return nil
	})

	// Fetch videos + details (waits for channel data, runs parallel with insights)
	eg.Go(func() error {
		select {
		case <-channelReady:
		case <-egCtx.Done():
			return egCtx.Err()
		}
		if channelData == nil {
			return nil
		}
		uploadsPlaylistID := channelData.ContentDetails.RelatedPlaylists.Uploads
		if uploadsPlaylistID == "" {
			log.Warn().Str("channel_id", wo.ChannelID).Msg("No uploads playlist ID found, skipping video fetch")
			return nil
		}

		var fetchErr error
		activities, fetchErr = ytClient.FetchVideos(egCtx, accessToken, uploadsPlaylistID, videosSince)
		if fetchErr != nil {
			if isUnauthorizedError(fetchErr) {
				log.Warn().Err(fetchErr).Str("channel_id", wo.ChannelID).Str("stage", "fetch_videos").Msg("Unauthorized - aborting")
				return ErrUnauthorized
			}
			log.Error().Err(fetchErr).Str("channel_id", wo.ChannelID).Str("stage", "fetch_videos").Msg("Failed to fetch videos")
			return nil
		}

		var videoIDs []string
		for _, a := range activities {
			if vid := a.ContentDetails.Upload.VideoID; vid != "" {
				videoIDs = append(videoIDs, vid)
			}
		}

		if len(videoIDs) > 0 {
			videoDetails, detailsErr := ytClient.FetchVideoDetails(egCtx, accessToken, videoIDs)
			if detailsErr != nil {
				if isUnauthorizedError(detailsErr) {
					log.Warn().Err(detailsErr).Str("channel_id", wo.ChannelID).Str("stage", "fetch_video_details").Msg("Unauthorized - aborting")
					return ErrUnauthorized
				}
				log.Error().Err(detailsErr).Str("channel_id", wo.ChannelID).Str("stage", "fetch_video_details").Msg("Failed to fetch video details")
			} else {
				videoDetailsMap = make(map[string]*social.YouTubeVideoItem)
				for i := range videoDetails {
					videoDetailsMap[videoDetails[i].ID] = &videoDetails[i]
				}
			}
		}

		log.Info().
			Str("channel_id", wo.ChannelID).
			Int("videos_count", len(activities)).
			Int("video_details_count", len(videoDetailsMap)).
			Msg("Fetched videos and details")
		return nil
	})

	// Fetch activity insights (independent)
	eg.Go(func() error {
		resp, err := ytClient.FetchActivityInsights(egCtx, accessToken, insightsStartDate, endDate)
		if err != nil {
			if isUnauthorizedError(err) {
				log.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "fetch_activity_insights").Msg("Unauthorized - stopping all API calls")
				return ErrUnauthorized
			}
			log.Error().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "fetch_activity_insights").Msg("Failed to fetch activity insights")
			return nil
		}
		activityMu.Lock()
		activityInsights = resp
		activityMu.Unlock()
		return nil
	})

	// Fetch traffic insights (independent)
	eg.Go(func() error {
		resp, err := ytClient.FetchTrafficInsights(egCtx, accessToken, insightsStartDate, endDate)
		if err != nil {
			if isUnauthorizedError(err) {
				log.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "fetch_traffic_insights").Msg("Unauthorized - stopping all API calls")
				return ErrUnauthorized
			}
			log.Error().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "fetch_traffic_insights").Msg("Failed to fetch traffic insights")
			return nil
		}
		trafficMu.Lock()
		trafficInsights = resp
		trafficMu.Unlock()
		return nil
	})

	// Fetch shared insights (independent)
	eg.Go(func() error {
		resp, err := ytClient.FetchSharedInsights(egCtx, accessToken, insightsStartDate, endDate)
		if err != nil {
			if isUnauthorizedError(err) {
				log.Warn().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "fetch_shared_insights").Msg("Unauthorized - stopping all API calls")
				return ErrUnauthorized
			}
			log.Error().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "fetch_shared_insights").Msg("Failed to fetch shared insights")
			return nil
		}
		sharedMu.Lock()
		sharedInsights = resp
		sharedMu.Unlock()
		return nil
	})

	// Wait for all goroutines (channel + videos + insights)
	if err := eg.Wait(); err != nil {
		if errors.Is(err, ErrUnauthorized) {
			log.Warn().Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "parallel_fetch").Msg("Aborting due to unauthorized error")
			if accountID, parseErr := primitive.ObjectIDFromHex(wo.ID); parseErr == nil {
				mongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
			}
			return
		}
		log.Error().Err(err).Str("error_message", err.Error()).Str("channel_id", wo.ChannelID).Str("function", "processWorkOrder").Str("stage", "parallel_fetch").Msg("Error during parallel fetch")
	}

	// ==================== PRODUCE MESSAGES ====================

	// Produce channel data
	if channelData != nil {
		subscriberCount, _ := strconv.ParseInt(channelData.Statistics.SubscriberCount, 10, 64)
		videoCount, _ := strconv.ParseInt(channelData.Statistics.VideoCount, 10, 64)
		viewCount, _ := strconv.ParseInt(channelData.Statistics.ViewCount, 10, 64)

		raw := kafkamodels.RawYouTubeChannel{
			ChannelID:       channelData.ID,
			Title:           channelData.Snippet.Title,
			Description:     channelData.Snippet.Description,
			CustomURL:       channelData.Snippet.CustomURL,
			ThumbnailURL:    channelData.Snippet.Thumbnails.High.URL,
			BannerURL:       channelData.BrandingSettings.Image.BannerExternalURL,
			Country:         channelData.Snippet.Country,
			SubscriberCount: subscriberCount,
			VideoCount:      videoCount,
			ViewCount:       viewCount,
			WorkspaceID:     wo.WorkspaceID,
			SavingTime:      now,
		}
		if pubAt, err := time.Parse(time.RFC3339, channelData.Snippet.PublishedAt); err == nil {
			raw.PublishedAt = pubAt
		}
		produceMessage(ctx, producer, topicRawChannels, wo.ChannelID, raw, log)
	}

	// Produce video messages
	if len(activities) > 0 {
		// Detect media type for each video using efficient duration-based detection
		// This uses duration from video details (fast) with HTTP fallback (parallel)
		var videosForDetection []social.YouTubeVideoItem
		for _, activity := range activities {
			videoID := activity.ContentDetails.Upload.VideoID
			if videoID == "" {
				continue
			}
			if details, ok := videoDetailsMap[videoID]; ok {
				videosForDetection = append(videosForDetection, *details)
			}
		}

		// Efficient detection: duration-based first, then parallel HTTP for unknowns
		videoMediaTypes := ytClient.DetectMediaTypes(ctx, videosForDetection)

		log.Info().
			Str("channel_id", wo.ChannelID).
			Int("videos_detected", len(videoMediaTypes)).
			Msg("Detected video media types")

		today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
		totalVideoMessages := 0

		for _, activity := range activities {
			videoID := activity.ContentDetails.Upload.VideoID
			if videoID == "" {
				continue
			}

			details, hasDetails := videoDetailsMap[videoID]
			if !hasDetails {
				log.Debug().Str("video_id", videoID).Msg("No details found for video, skipping")
				continue
			}

			views, _ := strconv.ParseInt(details.Statistics.ViewCount, 10, 64)
			likes, _ := strconv.ParseInt(details.Statistics.LikeCount, 10, 64)
			dislikes, _ := strconv.ParseInt(details.Statistics.DislikeCount, 10, 64)
			comments, _ := strconv.ParseInt(details.Statistics.CommentCount, 10, 64)
			favorites, _ := strconv.ParseInt(details.Statistics.FavoriteCount, 10, 64)

			raw := kafkamodels.RawYouTubeVideo{
				VideoID:       videoID,
				ChannelID:     wo.ChannelID,
				Title:         details.Snippet.Title,
				Description:   details.Snippet.Description,
				ThumbnailURL:  details.Snippet.Thumbnails.High.URL,
				Duration:      details.ContentDetails.Duration,
				WorkspaceID:   wo.WorkspaceID,
				SavingTime:    now,
				AnalyticsDate: today,
				MediaType:     videoMediaTypes[videoID],
				Views:         views,
				Likes:         likes,
				Dislikes:      dislikes,
				Comments:      comments,
				Favorites:     favorites,
			}

			if pubAt, err := time.Parse(time.RFC3339, details.Snippet.PublishedAt); err == nil {
				raw.PublishedAt = pubAt
			}
			if raw.ThumbnailURL == "" {
				raw.ThumbnailURL = details.Snippet.Thumbnails.Default.URL
			}

			produceMessage(ctx, producer, topicRawVideos, raw.VideoID, raw, log)
			totalVideoMessages++
		}

		log.Info().
			Str("channel_id", wo.ChannelID).
			Int("total_video_messages", totalVideoMessages).
			Msg("Produced video messages with lifetime statistics")
	}

	// Produce activity insights
	if activityInsights != nil {
		raw := struct {
			ChannelID   string                           `json:"channel_id"`
			Response    *social.YouTubeAnalyticsResponse `json:"response"`
			WorkspaceID string                           `json:"workspace_id"`
			SavingTime  time.Time                        `json:"saving_time"`
		}{
			ChannelID:   wo.ChannelID,
			Response:    activityInsights,
			WorkspaceID: wo.WorkspaceID,
			SavingTime:  now,
		}
		produceMessage(ctx, producer, topicRawActivityInsights, wo.ChannelID, raw, log)
		log.Info().Str("channel_id", wo.ChannelID).Int("rows", len(activityInsights.Rows)).Msg("Produced activity insights")
	}

	// Produce traffic insights
	if trafficInsights != nil {
		raw := struct {
			ChannelID   string                           `json:"channel_id"`
			Response    *social.YouTubeAnalyticsResponse `json:"response"`
			WorkspaceID string                           `json:"workspace_id"`
			SavingTime  time.Time                        `json:"saving_time"`
		}{
			ChannelID:   wo.ChannelID,
			Response:    trafficInsights,
			WorkspaceID: wo.WorkspaceID,
			SavingTime:  now,
		}
		produceMessage(ctx, producer, topicRawTrafficInsights, wo.ChannelID, raw, log)
	}

	// Produce shared insights
	if sharedInsights != nil {
		raw := struct {
			ChannelID   string                           `json:"channel_id"`
			Response    *social.YouTubeAnalyticsResponse `json:"response"`
			WorkspaceID string                           `json:"workspace_id"`
			SavingTime  time.Time                        `json:"saving_time"`
		}{
			ChannelID:   wo.ChannelID,
			Response:    sharedInsights,
			WorkspaceID: wo.WorkspaceID,
			SavingTime:  now,
		}
		produceMessage(ctx, producer, topicRawSharedInsights, wo.ChannelID, raw, log)
	}

	// Update timestamp after successful fetch
	select {
	case timestampUpdateChan <- TimestampUpdateRequest{AccountID: wo.ID, ChannelID: wo.ChannelID}:
	default:
		log.Warn().Str("channel_id", wo.ChannelID).Msg("Timestamp update channel full, skipping update")
	}

	log.Info().Str("channel_id", wo.ChannelID).Msg("Completed YouTube account fetch")
}

func produceMessage(ctx context.Context, producer kafka2.Producer, topic, key string, value interface{}, log *logger.Logger) {
	payload, err := json.Marshal(value)
	if err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("topic", topic).Str("function", "produceMessage").Str("stage", "marshal_message").Msg("Failed to marshal message")
		return
	}

	if err := producer.Produce(ctx, topic, []byte(key), payload); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("topic", topic).Str("key", key).Str("function", "produceMessage").Str("stage", "produce_kafka").Msg("Failed to produce message")
	}
}
