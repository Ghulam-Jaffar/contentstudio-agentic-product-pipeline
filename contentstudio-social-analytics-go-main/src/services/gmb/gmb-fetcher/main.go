package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/sync/semaphore"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
)

const (
	rawDataTopic   = "raw-gmb-data"
	workOrderTopic = "work-order-gmb"
	consumerGroup  = "gmb-fetcher-group"
	maxWorkers     = 10
	workChanSize   = 200

	// maxConcurrentAccounts is the max number of GMB locations processed simultaneously.
	maxConcurrentAccounts = 50
)

// Per-account concurrency guard
var accountSemaphores sync.Map

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

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("failed to load config: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)
	log := logger.New(cfg.LogLevel)
	log.Info().
		Int("workers", maxWorkers).
		Msg("Starting GMB Fetcher service")

	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka producer")
	}
	defer producer.Close()

	consumer, err := kafka2.NewConsumer(cfg.Kafka, consumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("failed to create kafka consumer")
	}
	defer consumer.Close()

	// MongoDB connection
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
	log.Info().Msg("Connected to MongoDB")

	mongoRepo := mongodb.NewUnifiedSocialRepository(mongoClient.Database(cfg.Mongo.Database), log.Logger)

	// ---- Build a RateManager for GMB API rate limiting ----
	// Google Business Profile Performance API: 300 requests/min per GCP project (shared across all endpoints).
	// Default: 4 RPS global (240/min, safely under 300/min), 2 RPS per token.
	rm := social.NewRateManager(social.RateLimits{
		GlobalRPS:     4.0,
		GlobalBurst:   5,
		PerTokenRPS:   2.0,
		PerTokenBurst: 3,
	})
	gmbClient := social.NewGMBClientWithRates(cfg.GMB.ClientID, cfg.GMB.ClientSecret, rm)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received")
		cancel()
	}()

	deps := &FetcherDependencies{
		Producer:  producer,
		Consumer:  consumer,
		MongoRepo: mongoRepo,
		GMBClient: gmbClient,
		Log:       log,
	}

	fetcherCfg := DefaultFetcherConfig()
	fetcherCfg.DecryptionKey = cfg.DecryptionKey

	if runErr := RunService(ctx, deps, fetcherCfg); runErr != nil {
		log.Error().Err(runErr).Msg("GMB Fetcher service error")
	}
	log.Info().Msg("GMB Fetcher stopped")
}

// HandleWorkOrder processes a single GMB account work order
func HandleWorkOrder(ctx context.Context, workOrder kafkamodels.GMBAccountWorkOrder, gmbClient social.GMBAPI, producer kafka2.Producer, mongoRepo mongodb.UnifiedSocialRepository, decryptionKey string, log *logger.Logger) error {
	locationID := workOrder.LocationID
	accountID := workOrder.AccountID
	workspaceID := workOrder.WorkspaceID

	log.Info().
		Str("location_id", locationID).
		Str("account_id", accountID).
		Str("workspace_id", workspaceID).
		Msg("Processing GMB work order")

	// Per-account concurrency guard — block until the previous run completes
	sem := semForAccount(locationID, 1)
	if err := sem.Acquire(ctx, 1); err != nil {
		log.Warn().Err(err).Str("location_id", locationID).Msg("Failed to acquire location semaphore")
		return nil
	}
	defer sem.Release(1)

	// Lookup account in MongoDB
	objID, err := primitive.ObjectIDFromHex(workOrder.ID)
	if err != nil {
		return fmt.Errorf("HandleWorkOrder: invalid social account ID: %w", err)
	}

	account, err := mongoRepo.FindByID(ctx, objID)
	if err != nil {
		mongoRepo.RecordProcessingError(ctx, objID, err.Error())
		return fmt.Errorf("HandleWorkOrder: find account: %w", err)
	}
	if account == nil {
		log.Warn().
			Str("social_account_id", workOrder.ID).
			Msg("Account not found in MongoDB, skipping")
		return nil
	}

	// Get access credentials
	accessToken := account.AccessToken
	refreshToken := account.RefreshToken

	// Decrypt tokens if needed
	if decryptionKey != "" && accessToken != "" {
		decrypted, err := crypto.DecryptToken(accessToken, decryptionKey)
		if err != nil {
			log.Warn().Err(err).Str("location_id", locationID).Msg("Failed to decrypt access token, using raw value")
		} else {
			accessToken = decrypted
		}
	}
	if decryptionKey != "" && refreshToken != "" {
		decrypted, err := crypto.DecryptToken(refreshToken, decryptionKey)
		if err != nil {
			log.Warn().Err(err).Str("location_id", locationID).Msg("Failed to decrypt refresh token, using raw value")
		} else {
			refreshToken = decrypted
		}
	}

	// Refresh access token
	tokenResp, err := gmbClient.RefreshToken(ctx, refreshToken)
	if err != nil {
		if social.IsExpectedCompetitorErrorGMB(err) {
			mongoRepo.RecordProcessingError(ctx, objID, err.Error())
			log.Warn().Err(err).Str("location_id", locationID).Msg("Expected token error, skipping account")
			return nil
		}
		mongoRepo.RecordProcessingError(ctx, objID, err.Error())
		return fmt.Errorf("HandleWorkOrder: refresh token: %w", err)
	}
	accessToken = tokenResp.AccessToken

	now := time.Now().UTC()

	// 1. Fetch Voice of Merchant status
	hasVoiceOfMerchant := false
	voMResp, voMErr := gmbClient.FetchVoiceOfMerchant(ctx, workOrder.LocationID, accessToken)
	if voMErr != nil {
		log.Warn().Err(voMErr).Str("location_id", locationID).Msg("Failed to fetch VoM, defaulting hasVoiceOfMerchant to false")
		if social.IsAuthError(voMErr) {
			mongoRepo.RecordProcessingError(ctx, objID, voMErr.Error())
			return voMErr
		}
	}
	if voMResp != nil {
		hasVoiceOfMerchant = voMResp.HasVoiceOfMerchant
	}

	// 2. Fetch performance metrics and search keywords only if hasVoiceOfMerchant
	if hasVoiceOfMerchant {
		fetchPerformanceMetrics(ctx, gmbClient, producer, workOrder, accessToken, now, log)
		fetchSearchKeywords(ctx, gmbClient, producer, workOrder, accessToken, now, log)
	} else {
		log.Info().Str("location_id", locationID).Msg("Skipping performance metrics and search keywords (hasVoiceOfMerchant=false)")
	}

	// 3. Fetch local posts
	fetchLocalPosts(ctx, gmbClient, producer, workOrder, accessToken, log)

	// 4. Fetch reviews (up to 2 pages)
	fetchReviews(ctx, gmbClient, producer, workOrder, accessToken, log)

	// 5. Fetch media assets
	fetchMediaAssets(ctx, gmbClient, producer, workOrder, accessToken, log)

	// Update MongoDB: last fetched timestamp, state, and has_voice_of_merchant
	if err := mongoRepo.Update(ctx, objID, bson.M{
		"last_analytics_fetched_at": now,
		"last_analytics_updated_at": now.Format("2006-01-02 15:04:05"),
		"state":                     mongomodels.StateProcessed,
		"has_voice_of_merchant":     hasVoiceOfMerchant,
	}); err != nil {
		log.Warn().Err(err).Str("location_id", locationID).Msg("Failed to update last fetched timestamp")
	}
	mongoRepo.ClearProcessingError(ctx, objID)

	log.Info().
		Str("location_id", locationID).
		Msg("GMB work order completed")

	return nil
}

func fetchPerformanceMetrics(ctx context.Context, gmbClient social.GMBAPI, producer kafka2.Producer, wo kafkamodels.GMBAccountWorkOrder, accessToken string, now time.Time, log *logger.Logger) {
	for i := 0; i < 3; i++ {
		endDate := now.AddDate(0, -i, 0)
		startDate := endDate.AddDate(0, -1, 0)

		resp, err := gmbClient.FetchPerformanceMetrics(ctx, wo.LocationID, accessToken, startDate, endDate)
		if err != nil {
			log.Warn().Err(err).
				Str("location_id", wo.LocationID).
				Int("month_offset", i).
				Msg("Failed to fetch performance metrics")
			continue
		}

		log.Info().
			Str("location_id", wo.LocationID).
			Int("month_offset", i).
			Str("start_date", startDate.Format("2006-01-02")).
			Str("end_date", endDate.Format("2006-01-02")).
			Int("multi_series_count", len(resp.MultiDailyMetricTimeSeries)).
			Msg("Fetched performance metrics, producing to Kafka")

		data, _ := json.Marshal(kafkamodels.RawGMBData{
			WorkspaceID:  wo.WorkspaceID,
			AccountID:    wo.AccountID,
			LocationID:   wo.LocationID,
			AccountName:  wo.AccountName,
			LocationName: wo.LocationName,
			LanguageCode: wo.LanguageCode,
			DataType:     "performance_metrics",
			Data:         resp,
		})

		if err := producer.Produce(ctx, rawDataTopic, []byte(wo.LocationID), data); err != nil {
			log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Failed to produce performance metrics to Kafka")
		}
	}
}

func fetchSearchKeywords(ctx context.Context, gmbClient social.GMBAPI, producer kafka2.Producer, wo kafkamodels.GMBAccountWorkOrder, accessToken string, now time.Time, log *logger.Logger) {
	for i := 0; i < 3; i++ {
		endMonth := now.AddDate(0, -i, 0)
		startMonth := endMonth.AddDate(0, -1, 0)

		resp, err := gmbClient.FetchSearchKeywords(ctx, wo.LocationID, accessToken, startMonth, endMonth)
		if err != nil {
			log.Warn().Err(err).
				Str("location_id", wo.LocationID).
				Int("month_offset", i).
				Msg("Failed to fetch search keywords")
			continue
		}

		data, _ := json.Marshal(kafkamodels.RawGMBData{
			WorkspaceID:  wo.WorkspaceID,
			AccountID:    wo.AccountID,
			LocationID:   wo.LocationID,
			AccountName:  wo.AccountName,
			LocationName: wo.LocationName,
			LanguageCode: wo.LanguageCode,
			DataType:     "search_keywords",
			KeywordMonth: startMonth.Format("2006-01"),
			Data:         resp,
		})

		if err := producer.Produce(ctx, rawDataTopic, []byte(wo.LocationID), data); err != nil {
			log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Failed to produce search keywords to Kafka")
		}
	}
}

func fetchLocalPosts(ctx context.Context, gmbClient social.GMBAPI, producer kafka2.Producer, wo kafkamodels.GMBAccountWorkOrder, accessToken string, log *logger.Logger) {
	resp, err := gmbClient.FetchLocalPosts(ctx, wo.AccountID, wo.LocationID, accessToken, "")
	if err != nil {
		log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Failed to fetch local posts")
		return
	}

	data, _ := json.Marshal(kafkamodels.RawGMBData{
		WorkspaceID:  wo.WorkspaceID,
		AccountID:    wo.AccountID,
		LocationID:   wo.LocationID,
		AccountName:  wo.AccountName,
		LocationName: wo.LocationName,
		LanguageCode: wo.LanguageCode,
		DataType:     "local_posts",
		Data:         resp,
	})

	if err := producer.Produce(ctx, rawDataTopic, []byte(wo.LocationID), data); err != nil {
		log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Failed to produce local posts to Kafka")
	}
}

func fetchReviews(ctx context.Context, gmbClient social.GMBAPI, producer kafka2.Producer, wo kafkamodels.GMBAccountWorkOrder, accessToken string, log *logger.Logger) {
	pageToken := ""
	for page := 0; page < 2; page++ {
		resp, err := gmbClient.FetchReviews(ctx, wo.AccountID, wo.LocationID, accessToken, pageToken)
		if err != nil {
			log.Warn().Err(err).Str("location_id", wo.LocationID).Int("page", page).Msg("Failed to fetch reviews")
			return
		}

		data, _ := json.Marshal(kafkamodels.RawGMBData{
			WorkspaceID:  wo.WorkspaceID,
			AccountID:    wo.AccountID,
			LocationID:   wo.LocationID,
			AccountName:  wo.AccountName,
			LocationName: wo.LocationName,
			LanguageCode: wo.LanguageCode,
			DataType:     "reviews",
			Data:         resp,
		})

		if err := producer.Produce(ctx, rawDataTopic, []byte(wo.LocationID), data); err != nil {
			log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Failed to produce reviews to Kafka")
		}

		if resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}
}

func fetchMediaAssets(ctx context.Context, gmbClient social.GMBAPI, producer kafka2.Producer, wo kafkamodels.GMBAccountWorkOrder, accessToken string, log *logger.Logger) {
	resp, err := gmbClient.FetchMediaAssets(ctx, wo.AccountID, wo.LocationID, accessToken, "")
	if err != nil {
		log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Failed to fetch media assets")
		return
	}

	data, _ := json.Marshal(kafkamodels.RawGMBData{
		WorkspaceID:  wo.WorkspaceID,
		AccountID:    wo.AccountID,
		LocationID:   wo.LocationID,
		AccountName:  wo.AccountName,
		LocationName: wo.LocationName,
		LanguageCode: wo.LanguageCode,
		DataType:     "media_assets",
		Data:         resp,
	})

	if err := producer.Produce(ctx, rawDataTopic, []byte(wo.LocationID), data); err != nil {
		log.Warn().Err(err).Str("location_id", wo.LocationID).Msg("Failed to produce media assets to Kafka")
	}
}
