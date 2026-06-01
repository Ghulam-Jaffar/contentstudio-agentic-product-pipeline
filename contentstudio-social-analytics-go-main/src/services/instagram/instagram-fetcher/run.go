package main

import (
	"context"
	"encoding/json"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/sync/semaphore"
)

// InstagramClientFactory creates Instagram API clients
type InstagramClientFactory func(appSecret string, connectedViaInstagram bool) InstagramAPI

// DefaultInstagramClientFactory creates real Instagram clients
func DefaultInstagramClientFactory(appSecret string, connectedViaInstagram bool) InstagramAPI {
	client := social.NewInstagramClient(appSecret)
	if connectedViaInstagram {
		client = client.WithBaseURL("https://graph.instagram.com/")
	}
	return client
}

// FetcherConfig holds service configuration
type FetcherConfig struct {
	MaxMediaWorkers         int
	MaxInsightsWorkers      int
	MediaQueueSize          int
	InsightsQueueSize       int
	MaxConcurrentAccounts   int
	TimestampUpdateChanSize int
	DecryptionKey           string
	AppSecret               string
}

// DefaultFetcherConfig returns default configuration
func DefaultFetcherConfig() FetcherConfig {
	return FetcherConfig{
		MaxMediaWorkers:         maxMediaWorkers,
		MaxInsightsWorkers:      maxInsightsWorkers,
		MediaQueueSize:          mediaQueueSize,
		InsightsQueueSize:       insightsQueueSize,
		MaxConcurrentAccounts:   maxConcurrentAccounts,
		TimestampUpdateChanSize: timestampUpdateChanSize,
	}
}

// FetcherDependencies holds external dependencies
type FetcherDependencies struct {
	Producer      kafka2.Producer
	Consumer      kafka2.Consumer
	MongoRepo     mongodb.UnifiedSocialRepository
	ClientFactory InstagramClientFactory
	Log           *logger.Logger
}

// FetcherMetrics holds service metrics
type FetcherMetrics struct {
	BatchesReceived     uint64
	AccountsProcessed   uint64
	MediaJobsCreated    uint64
	InsightsJobsCreated uint64
	TimestampUpdates    uint64
}

// FetcherService represents the Instagram fetcher service
type FetcherService struct {
	config  FetcherConfig
	deps    FetcherDependencies
	metrics FetcherMetrics
}

// NewFetcherService creates a new fetcher service
func NewFetcherService(cfg FetcherConfig, deps FetcherDependencies) *FetcherService {
	if deps.ClientFactory == nil {
		deps.ClientFactory = DefaultInstagramClientFactory
	}
	return &FetcherService{
		config: cfg,
		deps:   deps,
	}
}

// Run starts the fetcher service
func (s *FetcherService) Run(ctx context.Context) error {
	log := s.deps.Log

	timestampUpdateChan := make(chan TimestampUpdateRequest, s.config.TimestampUpdateChanSize)

	var updaterWg sync.WaitGroup
	if s.deps.MongoRepo != nil {
		startTimestampUpdater(&updaterWg, s.deps.MongoRepo, timestampUpdateChan, log)
	}

	maxConc := s.config.MaxConcurrentAccounts
	if maxConc <= 0 {
		maxConc = maxConcurrentAccounts
	}
	accountSem := semaphore.NewWeighted(int64(maxConc))

	var dispatchWg sync.WaitGroup
	var totalProcessed, totalFailed int64

	if s.deps.Consumer != nil {
		topics := []string{"work-order-instagram"}
		err := s.deps.Consumer.ConsumeWithAck(ctx, topics, func(ctx context.Context, topic string, key, value []byte, ack func()) error {
			var batch kafkamodels.InstagramBatchWorkOrder
			if err := json.Unmarshal(value, &batch); err != nil {
				log.Error().Err(err).Str("function", "Run").Str("stage", "unmarshal_batch_work_order").Msg("Failed to unmarshal batch work order")
				ack()
				return nil
			}

			total := len(batch.Accounts)
			log.Info().
				Str("batch_id", batch.BatchID).
				Int("accounts", total).
				Str("sync_type", batch.SyncType).
				Msg("Received batch work order, dispatching goroutines")

			atomic.AddUint64(&s.metrics.BatchesReceived, 1)

			var batchWg sync.WaitGroup
			var batchProcessed, batchFailed int64

			for _, account := range batch.Accounts {
				order := account

				token := resolveAccessToken(order.AccessToken, s.config.DecryptionKey, order.InstagramID, log)
				if token == "" {
					log.Error().Str("instagram_id", order.InstagramID).Str("function", "Run").Str("stage", "resolve_token").Msg("Empty access token after resolution; skipping")
					if s.deps.MongoRepo != nil {
						if accountID, parseErr := primitive.ObjectIDFromHex(order.ID); parseErr == nil {
							s.deps.MongoRepo.RecordProcessingError(ctx, accountID, "Access token is empty or decryption failed")
						}
					}
					atomic.AddInt64(&batchFailed, 1)
					continue
				}

				ro := ResolvedOrder{
					AccountID:             order.ID,
					InstagramID:           order.InstagramID,
					WorkspaceID:           order.WorkspaceID,
					AccessTokenPlaintext:  token,
					ConnectedViaInstagram: order.ConnectedViaInstagram,
					AppSecret:             s.config.AppSecret,
				}

				var mediaSince *time.Time
				if order.SyncType == "incremental" {
					t := time.Now().UTC().AddDate(0, 0, -14)
					mediaSince = &t
				}

				today := time.Now().UTC().Truncate(24 * time.Hour)
				until := today.AddDate(0, 0, -1).Add(5 * time.Hour)
				var insightsSince time.Time
				switch order.SyncType {
				case "incremental":
					insightsSince = today.AddDate(0, 0, -15).Add(8 * time.Hour)
				case "immediate":
					insightsSince = today.AddDate(0, 0, -30).Add(8 * time.Hour)
				default:
					insightsSince = today.AddDate(0, 0, -89).Add(8 * time.Hour)
				}

				mediaJob := MediaJob{Order: ro, SyncType: order.SyncType, Since: mediaSince}
				insightsJob := InsightsJob{Order: ro, SyncType: order.SyncType, Since: insightsSince, Until: until}

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

					sem := semForAccount(ro.InstagramID, perAccountConcurrency)
					if err := sem.Acquire(ctx, 1); err != nil {
						atomic.AddInt64(&batchFailed, 1)
						return
					}
					defer sem.Release(1)

					atomic.AddUint64(&s.metrics.AccountsProcessed, 1)
					atomic.AddUint64(&s.metrics.MediaJobsCreated, 1)
					atomic.AddUint64(&s.metrics.InsightsJobsCreated, 1)

					s.processMediaJobWithClient(context.Background(), log, mediaJob, timestampUpdateChan)
					s.processInsightsJobWithClient(context.Background(), log, insightsJob)
					atomic.AddInt64(&batchProcessed, 1)
				}()
			}

			batchID := batch.BatchID
			syncType := batch.SyncType
			go func() {
				batchWg.Wait()
				p := atomic.LoadInt64(&batchProcessed)
				f := atomic.LoadInt64(&batchFailed)
				atomic.AddInt64(&totalProcessed, p)
				atomic.AddInt64(&totalFailed, f)
				log.Info().
					Str("batch_id", batchID).
					Str("sync_type", syncType).
					Int("total", total).
					Int64("processed", p).
					Int64("failed", f).
					Msg("Batch processing complete")
			}()

			ack()
			return nil
		})
		if err != nil && err != context.Canceled && err != context.DeadlineExceeded {
			log.Error().Err(err).Str("function", "Run").Str("stage", "consume").Msg("Kafka consumer error")
		}
	}

	dispatchWg.Wait()

	log.Info().
		Int64("total_processed", atomic.LoadInt64(&totalProcessed)).
		Int64("total_failed", atomic.LoadInt64(&totalFailed)).
		Msg("Instagram Fetcher service stopped")

	close(timestampUpdateChan)
	updaterWg.Wait()

	return nil
}

func (s *FetcherService) mediaWorkerLoop(ctx context.Context, workerID int, jobs <-chan MediaJob, timestampUpdateChan chan<- TimestampUpdateRequest) {
	log := s.deps.Log
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "media").Logger()
	workerLog.Info().Msg("Media worker started")

	for job := range jobs {
		sem := semForAccount(job.Order.InstagramID, perAccountConcurrency)
		if err := sem.Acquire(context.Background(), 1); err != nil {
			workerLog.Error().Err(err).Str("instagram_id", job.Order.InstagramID).Msg("Failed to acquire per-account semaphore; skipping media job")
			if job.Ack != nil {
				job.Ack()
			}
			continue
		}

		atomic.AddUint64(&s.metrics.MediaJobsCreated, 1)
		s.processMediaJobWithClient(context.Background(), &logger.Logger{Logger: workerLog}, job, timestampUpdateChan)
		sem.Release(1)
		if job.Ack != nil {
			job.Ack()
		}
	}

	workerLog.Info().Msg("Media queue drained; worker stopping")
}

func (s *FetcherService) insightsWorkerLoop(ctx context.Context, workerID int, jobs <-chan InsightsJob) {
	log := s.deps.Log
	workerLog := log.With().Int("worker_id", workerID).Str("pool", "insights").Logger()
	workerLog.Info().Msg("Insights worker started")

	for job := range jobs {
		sem := semForAccount(job.Order.InstagramID, perAccountConcurrency)
		if err := sem.Acquire(context.Background(), 1); err != nil {
			workerLog.Error().Err(err).Str("instagram_id", job.Order.InstagramID).Msg("Failed to acquire per-account semaphore; skipping insights job")
			if job.Ack != nil {
				job.Ack()
			}
			continue
		}

		atomic.AddUint64(&s.metrics.InsightsJobsCreated, 1)
		s.processInsightsJobWithClient(context.Background(), &logger.Logger{Logger: workerLog}, job)
		sem.Release(1)
		if job.Ack != nil {
			job.Ack()
		}
	}

	workerLog.Info().Msg("Insights queue drained; worker stopping")
}

// processMediaJobWithClient processes a media job using the injected client factory
func (s *FetcherService) processMediaJobWithClient(ctx context.Context, log *logger.Logger, job MediaJob, timestampUpdateChan chan<- TimestampUpdateRequest) {
	ig := s.deps.ClientFactory(job.Order.AppSecret, job.Order.ConnectedViaInstagram)

	// Fetch user info
	userInfo, err := ig.FetchUserInfo(ctx, job.Order.InstagramID, job.Order.AccessTokenPlaintext)
	if err != nil {
		if isExpectedInstagramError(err) {
			log.Warn().Err(err).Str("error_message", err.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_user_info").Msg("FetchUserInfo failed")
		} else {
			log.Error().Err(err).Str("error_message", err.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_user_info").Msg("FetchUserInfo failed")
		}
		if accountID, parseErr := primitive.ObjectIDFromHex(job.Order.AccountID); parseErr == nil {
			s.deps.MongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
		}
		return
	}

	// Fetch media based on sync type
	var media []kafkamodels.RawInstagramMedia
	if job.Since != nil {
		media, err = ig.FetchMediaSince(ctx, job.Order.InstagramID, job.Order.AccessTokenPlaintext, *job.Since)
	} else {
		media, err = ig.FetchMedia(ctx, job.Order.InstagramID, job.Order.AccessTokenPlaintext)
	}

	if err != nil {
		if isExpectedInstagramError(err) {
			log.Warn().Err(err).Str("error_message", err.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_media").Msg("FetchMedia failed")
		} else {
			log.Error().Err(err).Str("error_message", err.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_media").Msg("FetchMedia failed")
		}
		if social.IsAuthError(err) {
			if accountID, parseErr := primitive.ObjectIDFromHex(job.Order.AccountID); parseErr == nil {
				s.deps.MongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
			}
		}
		return
	}

	// Fetch stories
	stories, storiesErr := ig.FetchStories(ctx, job.Order.InstagramID, job.Order.AccessTokenPlaintext)
	if storiesErr != nil {
		if isExpectedInstagramError(storiesErr) {
			log.Warn().Err(storiesErr).Str("error_message", storiesErr.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_stories").Msg("FetchStories failed (continuing)")
		} else {
			log.Error().Err(storiesErr).Str("error_message", storiesErr.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_stories").Msg("FetchStories failed (continuing)")
		}
	}

	// Enrich media with insights - parallel using mediaInsightsConc semaphore
	enrichedMedia := make([]EnrichedMedia, len(media))
	var mediaWg sync.WaitGroup
	for i := range media {
		enrichedMedia[i] = EnrichedMedia{RawInstagramMedia: &media[i]}
		mediaWg.Add(1)
		go func(i int) {
			defer mediaWg.Done()
			if err := mediaInsightsConc.Acquire(ctx, 1); err != nil {
				return
			}
			defer mediaInsightsConc.Release(1)
			if insights, insErr := ig.FetchMediaInsights(ctx, media[i].ID, job.Order.AccessTokenPlaintext, media[i].MediaType, media[i].MediaProductType); insErr == nil {
				enrichedMedia[i].Insights = insights
			} else if social.IsAuthError(insErr) {
				log.Warn().Err(insErr).Str("media_id", media[i].ID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_media_insights").Msg("Auth error on media insights")
			} else if isExpectedInstagramError(insErr) {
				log.Warn().Err(insErr).Str("media_id", media[i].ID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_media_insights").Msg("Expected error fetching media insights")
			} else {
				log.Error().Err(insErr).Str("error_message", insErr.Error()).Str("media_id", media[i].ID).Str("media_product_type", media[i].MediaProductType).Str("function", "processMediaJobWithClient").Str("stage", "fetch_media_insights").Msg("Failed to fetch media insights")
			}
		}(i)
	}
	mediaWg.Wait()

	// Enrich stories with insights - parallel using mediaInsightsConc semaphore
	enrichedStories := make([]EnrichedMedia, len(stories))
	var storyWg sync.WaitGroup
	for i := range stories {
		enrichedStories[i] = EnrichedMedia{RawInstagramMedia: &stories[i]}
		storyWg.Add(1)
		go func(i int) {
			defer storyWg.Done()
			if err := mediaInsightsConc.Acquire(ctx, 1); err != nil {
				return
			}
			defer mediaInsightsConc.Release(1)
			if insights, insErr := ig.FetchMediaInsights(ctx, stories[i].ID, job.Order.AccessTokenPlaintext, stories[i].MediaType, stories[i].MediaProductType); insErr == nil {
				enrichedStories[i].Insights = insights
			} else if social.IsAuthError(insErr) {
				log.Warn().Err(insErr).Str("media_id", stories[i].ID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_story_insights").Msg("Auth error on story insights")
			} else if isExpectedInstagramError(insErr) {
				log.Warn().Err(insErr).Str("media_id", stories[i].ID).Str("function", "processMediaJobWithClient").Str("stage", "fetch_story_insights").Msg("Expected error fetching story insights")
			} else {
				log.Error().Err(insErr).Str("error_message", insErr.Error()).Str("media_id", stories[i].ID).Str("media_product_type", stories[i].MediaProductType).Str("function", "processMediaJobWithClient").Str("stage", "fetch_story_insights").Msg("Failed to fetch story insights")
			}
		}(i)
	}
	storyWg.Wait()
	enrichedMedia = append(enrichedMedia, enrichedStories...)

	// Publish to Kafka in chunks to avoid MESSAGE_TOO_LARGE for accounts with many posts
	const mediaChunkSize = 50
	key := []byte(job.Order.InstagramID)
	for i := 0; i < len(enrichedMedia); i += mediaChunkSize {
		end := i + mediaChunkSize
		if end > len(enrichedMedia) {
			end = len(enrichedMedia)
		}
		chunk := struct {
			Media    []EnrichedMedia        `json:"media"`
			UserInfo map[string]interface{} `json:"user_info,omitempty"`
		}{
			Media:    enrichedMedia[i:end],
			UserInfo: userInfo,
		}
		data, err := json.Marshal(chunk)
		if err != nil {
			log.Error().Err(err).Str("instagram_id", job.Order.InstagramID).Msg("Failed to marshal media payload; skipping Kafka publish")
			continue
		}
		if err := s.deps.Producer.Produce(ctx, "raw-instagram-media", key, data); err != nil {
			log.Error().Err(err).Str("instagram_id", job.Order.InstagramID).Int("chunk_start", i).Int("chunk_end", end).Msg("Failed to produce raw-instagram-media message")
		}
	}

	// Request timestamp update
	select {
	case timestampUpdateChan <- TimestampUpdateRequest{AccountID: job.Order.AccountID, InstagramID: job.Order.InstagramID}:
	default:
		log.Warn().Str("instagram_id", job.Order.InstagramID).Msg("Timestamp update channel full; update dropped")
	}

	log.Info().
		Str("instagram_id", job.Order.InstagramID).
		Int("media_count", len(enrichedMedia)).
		Msg("Processed media job")
}

// processInsightsJobWithClient processes an insights job using the injected client factory
func (s *FetcherService) processInsightsJobWithClient(ctx context.Context, log *logger.Logger, job InsightsJob) {
	ig := s.deps.ClientFactory(job.Order.AppSecret, job.Order.ConnectedViaInstagram)

	// Fetch user info first
	userInfo, err := ig.FetchUserInfo(ctx, job.Order.InstagramID, job.Order.AccessTokenPlaintext)
	if err != nil {
		if isExpectedInstagramError(err) {
			log.Warn().Err(err).Str("error_message", err.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processInsightsJobWithClient").Str("stage", "fetch_user_info").Msg("FetchUserInfo failed")
		} else {
			log.Error().Err(err).Str("error_message", err.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processInsightsJobWithClient").Str("stage", "fetch_user_info").Msg("FetchUserInfo failed")
		}
		if accountID, parseErr := primitive.ObjectIDFromHex(job.Order.AccountID); parseErr == nil {
			s.deps.MongoRepo.RecordProcessingError(context.Background(), accountID, err.Error())
		}
	}

	// Fetch demographics
	demographics, demoErr := ig.FetchAccountDemographics(ctx, job.Order.InstagramID, job.Order.AccessTokenPlaintext)
	if demoErr != nil {
		if isExpectedInstagramError(demoErr) {
			log.Warn().Err(demoErr).Str("error_message", demoErr.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processInsightsJobWithClient").Str("stage", "fetch_demographics").Msg("FetchAccountDemographics failed (continuing)")
		} else {
			log.Error().Err(demoErr).Str("error_message", demoErr.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processInsightsJobWithClient").Str("stage", "fetch_demographics").Msg("FetchAccountDemographics failed (continuing)")
		}
	}

	// Calculate days for insights
	days := int(job.Until.Sub(job.Since).Hours() / 24)
	if days < 1 {
		days = 1
	}
	if days > 89 {
		days = 89
	}

	// Fetch daily insights
	dailyInsights, insErr := ig.FetchInsightsDaily(ctx, job.Order.InstagramID, job.Order.AccessTokenPlaintext, days, 5)
	if insErr != nil {
		if isExpectedInstagramError(insErr) {
			log.Warn().Err(insErr).Str("error_message", insErr.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processInsightsJobWithClient").Str("stage", "fetch_insights_daily").Msg("FetchInsightsDaily failed (continuing)")
		} else {
			log.Error().Err(insErr).Str("error_message", insErr.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processInsightsJobWithClient").Str("stage", "fetch_insights_daily").Msg("FetchInsightsDaily failed (continuing)")
		}
		if social.IsAuthError(insErr) {
			if accountID, parseErr := primitive.ObjectIDFromHex(job.Order.AccountID); parseErr == nil {
				s.deps.MongoRepo.RecordProcessingError(context.Background(), accountID, insErr.Error())
			}
		}
	}

	if userInfo == nil || len(userInfo) == 0 {
		log.Warn().Str("instagram_id", job.Order.InstagramID).Msg("No user info available, skipping insights publish")
		return
	}

	if len(dailyInsights) == 0 && demographics == nil {
		log.Warn().Str("instagram_id", job.Order.InstagramID).Msg("No insights/demographics to publish")
		return
	}

	// Publish combined insights
	combined := struct {
		DailyInsights []social.DailyInsight                 `json:"daily_insights,omitempty"`
		Demographics  *kafkamodels.RawInstagramDemographics `json:"demographics,omitempty"`
		UserInfo      map[string]interface{}                `json:"user_info,omitempty"`
	}{
		DailyInsights: dailyInsights,
		Demographics:  demographics,
		UserInfo:      userInfo,
	}

	key := []byte(job.Order.InstagramID)
	payload, _ := json.Marshal(combined)

	if err := s.deps.Producer.Produce(ctx, "raw-instagram-insights", key, payload); err != nil {
		log.Error().Err(err).Str("error_message", err.Error()).Str("instagram_id", job.Order.InstagramID).Str("function", "processInsightsJobWithClient").Str("stage", "produce_insights").Msg("Failed to publish insights")
	} else {
		log.Info().
			Str("instagram_id", job.Order.InstagramID).
			Int("daily_insights_count", len(dailyInsights)).
			Msg("Published combined insights payload")
	}
}

// GetMetrics returns current service metrics
func (s *FetcherService) GetMetrics() FetcherMetrics {
	return FetcherMetrics{
		BatchesReceived:     atomic.LoadUint64(&s.metrics.BatchesReceived),
		AccountsProcessed:   atomic.LoadUint64(&s.metrics.AccountsProcessed),
		MediaJobsCreated:    atomic.LoadUint64(&s.metrics.MediaJobsCreated),
		InsightsJobsCreated: atomic.LoadUint64(&s.metrics.InsightsJobsCreated),
		TimestampUpdates:    atomic.LoadUint64(&s.metrics.TimestampUpdates),
	}
}
