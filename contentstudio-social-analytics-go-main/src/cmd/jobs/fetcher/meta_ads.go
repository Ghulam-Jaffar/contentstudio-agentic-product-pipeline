package fetcher

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// Meta Ads batch configuration
const (
	metaAdsMongoFetchSize      int64 = 50
	metaAdsKafkaBatchSize      int64 = 200
	metaAdsUpdateIntervalHours       = 6
	topicMetaAdsBatch                = "work-order-meta-ads"
)

// ProcessMetaAdsAccounts fetches Meta Ads accounts needing update and produces
// batch work orders to the "work-order-meta-ads" Kafka topic.
func ProcessMetaAdsAccounts(db *mongo.Database, producer kafka.Producer, log zerolog.Logger, syncType string) {
	log.Info().Str("sync_type", syncType).Msg("Processing Meta Ads accounts")

	repo := mongodb.NewUnifiedSocialRepository(db, log.With().Str("repository", "unified_social").Logger())

	ctx := context.Background()

	totalCount, err := repo.CountAccountsNeedingUpdate(ctx, mongomodels.PlatformMetaAds, nil, metaAdsUpdateIntervalHours)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count Meta Ads accounts needing update")
		return
	}

	if totalCount == 0 {
		log.Info().Msg("No Meta Ads accounts need update")
		return
	}

	log.Info().
		Int64("total_accounts", totalCount).
		Int64("mongo_fetch_size", metaAdsMongoFetchSize).
		Int64("kafka_batch_size", metaAdsKafkaBatchSize).
		Msg("Starting Meta Ads batch processing")

	var (
		lastID          = primitive.NilObjectID
		batchesCreated  int
		accountsQueued  int
		accountsSkipped int
	)

	accumulated := make([]kafkamodels.MetaAdsWorkOrder, 0, metaAdsKafkaBatchSize)
	kafkaHealthy := true

	produceBatch := func(batch []kafkamodels.MetaAdsWorkOrder) bool {
		if len(batch) == 0 {
			return true
		}
		batchWorkOrder := kafkamodels.MetaAdsBatchWorkOrder{
			BatchID:  uuid.New().String(),
			Accounts: batch,
		}
		payload, err := json.Marshal(batchWorkOrder)
		if err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Msg("Failed to marshal Meta Ads batch work order")
			return false
		}
		if err := producer.Produce(ctx, topicMetaAdsBatch, []byte(batchWorkOrder.BatchID), payload); err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Msg("Failed to produce Meta Ads batch work order to Kafka")
			return false
		}
		batchesCreated++
		accountsQueued += len(batch)
		log.Info().
			Str("batch_id", batchWorkOrder.BatchID).
			Int("accounts_in_batch", len(batch)).
			Int("batches_created", batchesCreated).
			Int("total_queued", accountsQueued).
			Msg("Produced Meta Ads batch work order")
		return true
	}

	for kafkaHealthy {
		accounts, err := repo.GetAccountsNeedingUpdateByID(ctx, mongomodels.PlatformMetaAds, nil, metaAdsUpdateIntervalHours, lastID, metaAdsMongoFetchSize)
		if err != nil {
			log.Error().Err(err).Str("last_id", lastID.Hex()).Msg("Failed to fetch Meta Ads accounts batch")
			break
		}
		if len(accounts) == 0 {
			break
		}
		lastID = accounts[len(accounts)-1].ID

		valid, skipped := buildMetaAdsAccountBatch(accounts, syncType, log)
		accountsSkipped += skipped
		accumulated = append(accumulated, valid...)

		for int64(len(accumulated)) >= metaAdsKafkaBatchSize {
			batch := accumulated[:metaAdsKafkaBatchSize]
			if !produceBatch(batch) {
				kafkaHealthy = false
				log.Error().Int("pending_accounts", len(accumulated)).Msg("Kafka unhealthy, stopping Meta Ads accumulation")
				break
			}
			accumulated = accumulated[metaAdsKafkaBatchSize:]
		}
	}

	if kafkaHealthy && len(accumulated) > 0 {
		produceBatch(accumulated)
	}

	log.Info().
		Int("batches_created", batchesCreated).
		Int("accounts_queued", accountsQueued).
		Int("accounts_skipped", accountsSkipped).
		Int64("total_eligible", totalCount).
		Msg("Completed Meta Ads batch production")
}

// buildMetaAdsAccountBatch validates accounts and builds work order slice.
func buildMetaAdsAccountBatch(
	accounts []mongomodels.SocialIntegration,
	syncType string,
	log zerolog.Logger,
) ([]kafkamodels.MetaAdsWorkOrder, int) {
	batch := make([]kafkamodels.MetaAdsWorkOrder, 0, len(accounts))
	skipped := 0
	_ = time.Now() // ensure time import used

	for _, account := range accounts {
		longAccessToken := account.GetAccessToken()
		accessToken := GetStringFromExtraData(account.ExtraData, "access_token")

		if accessToken == "" && longAccessToken == "" {
			log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping Meta Ads account: missing access token")
			skipped++
			continue
		}
		if account.WorkspaceID.IsZero() {
			log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping Meta Ads account: missing workspace_id")
			skipped++
			continue
		}

		// Strip "act_" prefix to get the numeric account ID
		accountID := account.PlatformIdentifier
		if len(accountID) > 4 && accountID[:4] == "act_" {
			accountID = accountID[4:]
		}

		batch = append(batch, kafkamodels.MetaAdsWorkOrder{
			MongoID:            account.ID.Hex(),
			PlatformIdentifier: account.PlatformIdentifier,
			AccountID:          accountID,
			AccessToken:        accessToken,
			LongAccessToken:    longAccessToken,
			WorkspaceID:        account.WorkspaceID.Hex(),
			SyncType:           syncType,
		})
	}

	return batch, skipped
}
