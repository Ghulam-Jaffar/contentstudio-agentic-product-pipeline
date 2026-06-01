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

	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"golang.org/x/sync/errgroup"

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
	maxWorkers        = 15
	workOrderChanSize = 500

	// fetcherDateRangeDays is the look-back period used by the scheduled fetcher.
	fetcherDateRangeDays = 14

	kafkaBatchTopic    = "work-order-meta-ads"
	kafkaConsumerGroup = "meta-ads-fetcher-group"

	topicAccountInfo      = "raw-meta-ads-account-info"
	topicCampaigns        = "raw-meta-ads-campaigns"
	topicAdsets           = "raw-meta-ads-adsets"
	topicAds              = "raw-meta-ads-ads"
	topicCampaignInsights = "raw-meta-ads-campaign-insights"
	topicAdsetInsights    = "raw-meta-ads-adset-insights"
	topicAdInsights       = "raw-meta-ads-ad-insights"
	topicAgeGender        = "raw-meta-ads-demographics-age-gender"
	topicDevicePlatform   = "raw-meta-ads-demographics-device-platform"
	topicRegionCountry    = "raw-meta-ads-demographics-region-country"

	batchSize = 500
)

type workOrderMsg struct {
	wo kafkamodels.MetaAdsWorkOrder
}

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	telemetry.ConfigureSentry(cfg)

	log := logger.New(cfg.LogLevel)
	log.Info().Msg("Starting Meta Ads Fetcher service")

	metaAdsClient := social.NewMetaAdsClient(cfg.Facebook.AppSecret, log)

	consumer, err := kafka2.NewConsumer(cfg.Kafka, kafkaConsumerGroup, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka consumer")
	}
	defer consumer.Close()

	producer, err := kafka2.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()

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
	if err := mongoClient.Ping(context.Background(), readpref.Primary()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}
	mongoRepo := mongodb.NewUnifiedSocialRepository(mongoClient.Database(cfg.Mongo.Database), log.Logger)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Info().Msg("Shutdown signal received, cancelling context")
		cancel()
	}()

	workChan := make(chan workOrderMsg, workOrderChanSize)

	var wg sync.WaitGroup
	for i := 0; i < maxWorkers; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			fetchWorker(ctx, id, workChan, metaAdsClient, producer, mongoRepo, cfg, log)
		}(i)
	}

	log.Info().Int("workers", maxWorkers).Str("topic", kafkaBatchTopic).Msg("Starting batch consumer")

	if err := consumer.Consume(ctx, []string{kafkaBatchTopic},
		func(ctx context.Context, topic string, key, value []byte) error {
			var batch kafkamodels.MetaAdsBatchWorkOrder
			if err := json.Unmarshal(value, &batch); err != nil {
				log.Error().Err(err).Str("stage", "unmarshal_batch").Msg("Failed to unmarshal MetaAdsBatchWorkOrder")
				return nil
			}
			log.Info().
				Str("batch_id", batch.BatchID).
				Int("accounts", len(batch.Accounts)).
				Msg("Received batch work order")

			for _, wo := range batch.Accounts {
				select {
				case workChan <- workOrderMsg{wo: wo}:
				case <-ctx.Done():
					return ctx.Err()
				}
			}
			return nil
		}); err != nil && err != context.Canceled {
		log.Error().Err(err).Msg("Consumer error")
	}

	close(workChan)
	wg.Wait()
	log.Info().Msg("Meta Ads Fetcher service stopped")
}

func fetchWorker(
	ctx context.Context,
	workerID int,
	workChan <-chan workOrderMsg,
	client *social.MetaAdsClient,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	cfg *config.Config,
	log *logger.Logger,
) {
	log.Info().Int("worker_id", workerID).Msg("Fetcher worker started")
	defer log.Info().Int("worker_id", workerID).Msg("Fetcher worker stopped")

	for msg := range workChan {
		wo := msg.wo
		start := time.Now()
		if err := processAccount(ctx, wo, client, producer, mongoRepo, cfg, log); err != nil {
			log.Error().Err(err).
				Str("account_id", wo.MongoID).
				Str("platform_identifier", wo.PlatformIdentifier).
				Dur("duration", time.Since(start)).
				Msg("Failed to fetch Meta Ads account data")
		} else {
			log.Info().
				Str("account_id", wo.MongoID).
				Str("platform_identifier", wo.PlatformIdentifier).
				Dur("duration", time.Since(start)).
				Msg("Fetched Meta Ads account data successfully")

			// Update Mongo state -> processed and analytics timestamp
			if mongoRepo != nil {
				if objID, err := primitive.ObjectIDFromHex(wo.MongoID); err == nil {
					now := time.Now().UTC()
					if err := mongoRepo.UpdateState(context.Background(), objID, mongomodels.StateProcessed); err != nil {
						log.Error().Err(err).Str("account_id", wo.MongoID).Msg("Failed to update account state to processed")
					}
					if err := mongoRepo.UpdateAnalyticsTimestamp(context.Background(), objID, "analytics", now); err != nil {
						log.Error().Err(err).Str("account_id", wo.MongoID).Msg("Failed to update analytics timestamp")
					} else {
						log.Debug().Str("account_id", wo.MongoID).Msg("Updated account state and analytics timestamp")
					}
				}
			}
		}
	}
}

func processAccount(
	ctx context.Context,
	wo kafkamodels.MetaAdsWorkOrder,
	client *social.MetaAdsClient,
	producer kafka2.Producer,
	mongoRepo mongodb.UnifiedSocialRepository,
	cfg *config.Config,
	log *logger.Logger,
) error {
	accessToken := resolveToken(wo, cfg.DecryptionKey)
	if accessToken == "" {
		return fmt.Errorf("no access token available for account %s", wo.MongoID)
	}

	accountID := wo.PlatformIdentifier
	until := time.Now().UTC()
	since := until.AddDate(0, 0, -fetcherDateRangeDays)

	eg, egCtx := errgroup.WithContext(ctx)

	// 1. Account info
	eg.Go(func() error {
		raw, err := client.FetchAccountInfo(egCtx, accountID, accessToken)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "account_info").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Str("endpoint", "account_info").Str("account_id", accountID).Msg("Fetched account info")
		payload := kafkamodels.MetaAdsAccountInfoPayload{WorkOrder: wo, AccountInfo: *raw}
		return publishJSON(egCtx, producer, topicAccountInfo, wo.MongoID, payload, log)
	})

	// 2. Campaigns
	eg.Go(func() error {
		rows, err := client.FetchCampaigns(egCtx, accountID, accessToken, since, until)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "campaigns").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Int("rows", len(rows)).Str("endpoint", "campaigns").Str("account_id", accountID).Msg("Fetched rows")
		return publishBatched(egCtx, producer, topicCampaigns, wo, rows, batchSize,
			func(batch []kafkamodels.RawMetaAdsCampaign) interface{} {
				return kafkamodels.MetaAdsCampaignsPayload{WorkOrder: wo, Campaigns: batch}
			}, log)
	})

	// 3. Adsets
	eg.Go(func() error {
		rows, err := client.FetchAdsets(egCtx, accountID, accessToken, since, until)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "adsets").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Int("rows", len(rows)).Str("endpoint", "adsets").Str("account_id", accountID).Msg("Fetched rows")
		return publishBatched(egCtx, producer, topicAdsets, wo, rows, batchSize,
			func(batch []kafkamodels.RawMetaAdsAdset) interface{} {
				return kafkamodels.MetaAdsAdsetsPayload{WorkOrder: wo, Adsets: batch}
			}, log)
	})

	// 4. Ads
	eg.Go(func() error {
		rows, err := client.FetchAds(egCtx, accountID, accessToken, since, until)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "ads").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Int("rows", len(rows)).Str("endpoint", "ads").Str("account_id", accountID).Msg("Fetched rows")
		return publishBatched(egCtx, producer, topicAds, wo, rows, batchSize,
			func(batch []kafkamodels.RawMetaAdsAd) interface{} {
				return kafkamodels.MetaAdsAdsPayload{WorkOrder: wo, Ads: batch}
			}, log)
	})

	// 5. Campaign insights
	eg.Go(func() error {
		rows, err := client.FetchCampaignInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "campaign_insights").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Int("rows", len(rows)).Str("endpoint", "campaign_insights").Str("account_id", accountID).Msg("Fetched rows")
		return publishBatched(egCtx, producer, topicCampaignInsights, wo, rows, batchSize,
			func(batch []kafkamodels.RawMetaAdsInsightRow) interface{} {
				return kafkamodels.MetaAdsCampaignInsightsPayload{WorkOrder: wo, Insights: batch}
			}, log)
	})

	// 6. Adset insights
	eg.Go(func() error {
		rows, err := client.FetchAdsetInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "adset_insights").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Int("rows", len(rows)).Str("endpoint", "adset_insights").Str("account_id", accountID).Msg("Fetched rows")
		return publishBatched(egCtx, producer, topicAdsetInsights, wo, rows, batchSize,
			func(batch []kafkamodels.RawMetaAdsInsightRow) interface{} {
				return kafkamodels.MetaAdsAdsetInsightsPayload{WorkOrder: wo, Insights: batch}
			}, log)
	})

	// 7. Ad insights
	eg.Go(func() error {
		rows, err := client.FetchAdInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "ad_insights").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Int("rows", len(rows)).Str("endpoint", "ad_insights").Str("account_id", accountID).Msg("Fetched rows")
		return publishBatched(egCtx, producer, topicAdInsights, wo, rows, batchSize,
			func(batch []kafkamodels.RawMetaAdsInsightRow) interface{} {
				return kafkamodels.MetaAdsAdInsightsPayload{WorkOrder: wo, Insights: batch}
			}, log)
	})

	// 8. Demographics age/gender
	eg.Go(func() error {
		rows, err := client.FetchAgeGenderInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "demographics_age_gender").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Int("rows", len(rows)).Str("endpoint", "demographics_age_gender").Str("account_id", accountID).Msg("Fetched rows")
		return publishBatched(egCtx, producer, topicAgeGender, wo, rows, batchSize,
			func(batch []kafkamodels.RawMetaAdsDemographicsRow) interface{} {
				return kafkamodels.MetaAdsDemographicsAgeGenderPayload{WorkOrder: wo, Rows: batch}
			}, log)
	})

	// 9. Demographics device/platform
	eg.Go(func() error {
		rows, err := client.FetchDevicePlatformInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "demographics_device_platform").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Int("rows", len(rows)).Str("endpoint", "demographics_device_platform").Str("account_id", accountID).Msg("Fetched rows")
		return publishBatched(egCtx, producer, topicDevicePlatform, wo, rows, batchSize,
			func(batch []kafkamodels.RawMetaAdsDemographicsRow) interface{} {
				return kafkamodels.MetaAdsDemographicsDevicePlatformPayload{WorkOrder: wo, Rows: batch}
			}, log)
	})

	// 10. Demographics region/country
	eg.Go(func() error {
		rows, err := client.FetchRegionCountryInsights(egCtx, accountID, accessToken, since, until)
		if err != nil {
			log.Warn().Err(err).Str("endpoint", "demographics_region_country").Str("account_id", accountID).Msg("Fetch failed")
			return nil
		}
		log.Info().Int("rows", len(rows)).Str("endpoint", "demographics_region_country").Str("account_id", accountID).Msg("Fetched rows")
		return publishBatched(egCtx, producer, topicRegionCountry, wo, rows, batchSize,
			func(batch []kafkamodels.RawMetaAdsDemographicsRow) interface{} {
				return kafkamodels.MetaAdsDemographicsRegionCountryPayload{WorkOrder: wo, Rows: batch}
			}, log)
	})

	return eg.Wait()
}

func resolveToken(wo kafkamodels.MetaAdsWorkOrder, decryptionKey string) string {
	if wo.LongAccessToken != "" {
		if dec, err := crypto.DecryptToken(wo.LongAccessToken, decryptionKey); err == nil && dec != "" {
			return dec
		}
		return wo.LongAccessToken
	}
	if wo.AccessToken != "" {
		if dec, err := crypto.DecryptToken(wo.AccessToken, decryptionKey); err == nil && dec != "" {
			return dec
		}
		return wo.AccessToken
	}
	return ""
}

func publishJSON(ctx context.Context, producer kafka2.Producer, topic, key string, payload interface{}, log *logger.Logger) error {
	b, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal payload for %s: %w", topic, err)
	}
	if err := producer.Produce(ctx, topic, []byte(key), b); err != nil {
		log.Error().Err(err).Str("topic", topic).Str("key", key).Msg("Failed to publish to Kafka")
		return fmt.Errorf("produce to %s: %w", topic, err)
	}
	return nil
}

func publishBatched[T any](
	ctx context.Context,
	producer kafka2.Producer,
	topic string,
	wo kafkamodels.MetaAdsWorkOrder,
	rows []T,
	size int,
	makeBatch func([]T) interface{},
	log *logger.Logger,
) error {
	for i := 0; i < len(rows); i += size {
		end := i + size
		if end > len(rows) {
			end = len(rows)
		}
		payload := makeBatch(rows[i:end])
		if err := publishJSON(ctx, producer, topic, wo.MongoID, payload, log); err != nil {
			return err
		}
	}
	return nil
}

// ensure mongomodels import used (for potential future state updates)
var _ = mongomodels.PlatformMetaAds
