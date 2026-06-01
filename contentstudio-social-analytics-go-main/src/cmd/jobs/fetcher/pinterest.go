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

// Pinterest batch configuration constants
const (
	// pinterestMongoFetchSize is the number of accounts to fetch per MongoDB query (small to avoid cursor timeout)
	pinterestMongoFetchSize int64 = 50

	// pinterestKafkaBatchSize is the number of accounts per Kafka message (larger for efficient processing)
	pinterestKafkaBatchSize int64 = 200

	pinterestUpdateIntervalHours = 6
	topicPinterestBatch          = "work-order-pinterest"
)

// ProcessPinterestAccounts fetches Pinterest accounts needing update and produces batch work orders.
// Uses small chunk fetching (to avoid cursor timeout) + accumulation pattern for efficient Kafka batches.
// Only produces batches for accounts that haven't been updated in the last 6 hours.
func ProcessPinterestAccounts(ctx context.Context, db *mongo.Database, producer kafka.Producer, logger zerolog.Logger, accountTypes []string, syncType string) {
	unifiedRepo := mongodb.NewUnifiedSocialRepository(db, logger.With().Str("repository", "unified_social").Logger())

	log := logger.With().Str("platform", "pinterest").Str("topic", topicPinterestBatch).Logger()

	// Get total count for progress logging
	totalCount, err := unifiedRepo.CountAccountsNeedingUpdate(ctx, mongomodels.PlatformPinterest, accountTypes, pinterestUpdateIntervalHours)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count Pinterest accounts needing update")
		return
	}

	if totalCount == 0 {
		log.Info().Msg("No Pinterest accounts need update")
		return
	}

	log.Info().
		Int64("total_accounts", totalCount).
		Int64("mongo_fetch_size", pinterestMongoFetchSize).
		Int64("kafka_batch_size", pinterestKafkaBatchSize).
		Msg("Starting batch processing")

	var (
		lastID          = primitive.NilObjectID // Start from beginning
		batchesCreated  int
		batchesFailed   int
		accountsQueued  int
		accountsSkipped int
	)

	// Accumulator for valid accounts - will be sent to Kafka when reaching kafkaBatchSize
	accumulated := make([]kafkamodels.PinterestAccountWorkOrder, 0, pinterestKafkaBatchSize)

	// Helper function to produce a batch to Kafka
	produceBatch := func(batch []kafkamodels.PinterestAccountWorkOrder) {
		if len(batch) == 0 {
			return
		}

		batchWorkOrder := kafkamodels.PinterestBatchWorkOrder{
			BatchID:   uuid.New().String(),
			SyncType:  syncType,
			Accounts:  batch,
			CreatedAt: time.Now().UTC(),
		}

		payload, err := json.Marshal(batchWorkOrder)
		if err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Msg("Failed to marshal batch work order")
			batchesFailed++
			return
		}

		if err := producer.Produce(ctx, topicPinterestBatch, []byte(batchWorkOrder.BatchID), payload); err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Msg("Failed to produce batch work order to Kafka")
			batchesFailed++
			return
		}

		batchesCreated++
		accountsQueued += len(batch)

		log.Info().
			Str("batch_id", batchWorkOrder.BatchID).
			Int("accounts_in_batch", len(batch)).
			Int("batches_created", batchesCreated).
			Int("total_queued", accountsQueued).
			Msg("Produced batch work order")
	}

	// Fetch small chunks from MongoDB using ID-based pagination
	for {
		accounts, err := unifiedRepo.GetAccountsNeedingUpdateByID(ctx, mongomodels.PlatformPinterest, accountTypes, pinterestUpdateIntervalHours, lastID, pinterestMongoFetchSize)
		if err != nil {
			log.Error().Err(err).Str("last_id", lastID.Hex()).Msg("Failed to fetch accounts batch")
			break
		}

		if len(accounts) == 0 {
			break
		}

		// Update lastID for next iteration
		lastID = accounts[len(accounts)-1].ID

		// Validate and accumulate accounts
		validAccounts, skipped := buildPinterestAccountBatch(accounts, syncType, log)
		accountsSkipped += skipped
		accumulated = append(accumulated, validAccounts...)

		// Produce to Kafka when we have enough accounts (exactly kafkaBatchSize)
		for int64(len(accumulated)) >= pinterestKafkaBatchSize {
			produceBatch(accumulated[:pinterestKafkaBatchSize])
			accumulated = accumulated[pinterestKafkaBatchSize:]
		}
	}

	// Produce any remaining accounts
	if len(accumulated) > 0 {
		produceBatch(accumulated)
	}

	log.Info().
		Int("batches_created", batchesCreated).
		Int("batches_failed", batchesFailed).
		Int("accounts_queued", accountsQueued).
		Int("accounts_skipped", accountsSkipped).
		Int64("total_eligible", totalCount).
		Msg("Completed batch production for Pinterest accounts")
}

// buildPinterestAccountBatch validates accounts and builds the batch payload.
// Returns the valid accounts and count of skipped accounts.
func buildPinterestAccountBatch(
	accounts []mongomodels.SocialIntegration,
	syncType string,
	log zerolog.Logger,
) ([]kafkamodels.PinterestAccountWorkOrder, int) {
	batch := make([]kafkamodels.PinterestAccountWorkOrder, 0, len(accounts))
	skipped := 0

	for _, account := range accounts {
		// Get access token
		accessToken := account.GetAccessToken()
		if accessToken == "" {
			accessToken = GetStringFromExtraData(account.ExtraData, "access_token")
		}
		if accessToken == "" {
			log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing access token")
			skipped++
			continue
		}

		// Validate workspace ID
		if account.WorkspaceID.IsZero() {
			log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing workspace_id")
			skipped++
			continue
		}

		// Get account type (board or profile) - default to profile if not specified
		accountType := account.Type
		if accountType == "" {
			accountType = kafkamodels.PinterestAccountTypeProfile
		}

		// Get board ID for board accounts
		boardID := ""
		if accountType == kafkamodels.PinterestAccountTypeBoard {
			boardID = GetStringFromExtraData(account.ExtraData, "board_id")
			if boardID == "" {
				log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping board account: missing board_id")
				skipped++
				continue
			}
		}

		batch = append(batch, kafkamodels.PinterestAccountWorkOrder{
			ID:          account.ID.Hex(),
			AccountID:   account.PlatformIdentifier,
			AccessToken: accessToken,
			AccountType: accountType,
			BoardID:     boardID,
			WorkspaceID: account.WorkspaceID.Hex(),
			SyncType:    syncType,
		})
	}

	return batch, skipped
}
