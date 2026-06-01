package fetcher

import (
	"context"
	"encoding/json"
	"strings"
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

// GMB batch configuration constants
const (
	// gmbMongoFetchSize is the number of accounts to fetch per MongoDB query
	gmbMongoFetchSize int64 = 50

	// gmbKafkaBatchSize is the number of accounts per Kafka message
	gmbKafkaBatchSize int64 = 50

	gmbUpdateIntervalHours = 6
	topicGMBBatch          = "work-order-gmb"
)

// ProcessGMBAccounts fetches GMB accounts needing update and produces batch work orders.
// Uses small chunk fetching (to avoid cursor timeout) + accumulation pattern for efficient Kafka batches.
// Only produces batches for accounts that haven't been updated in the last 6 hours.
func ProcessGMBAccounts(ctx context.Context, db *mongo.Database, producer kafka.Producer, logger zerolog.Logger, syncType string) {
	unifiedRepo := mongodb.NewUnifiedSocialRepository(db, logger.With().Str("repository", "unified_social").Logger())

	log := logger.With().Str("platform", "gmb").Str("topic", topicGMBBatch).Logger()

	// Get total count for progress logging
	totalCount, err := unifiedRepo.CountAccountsNeedingUpdate(ctx, mongomodels.PlatformGMB, nil, gmbUpdateIntervalHours)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count GMB accounts needing update")
		return
	}

	if totalCount == 0 {
		log.Info().Msg("No GMB accounts need update")
		return
	}

	log.Info().
		Int64("total_accounts", totalCount).
		Int64("mongo_fetch_size", gmbMongoFetchSize).
		Int64("kafka_batch_size", gmbKafkaBatchSize).
		Msg("Starting batch processing")

	var (
		lastID          = primitive.NilObjectID
		batchesCreated  int
		batchesFailed   int
		accountsQueued  int
		accountsSkipped int
	)

	// Accumulator for valid accounts — will be sent to Kafka when reaching gmbKafkaBatchSize
	accumulated := make([]kafkamodels.GMBAccountWorkOrder, 0, gmbKafkaBatchSize)

	// Helper function to produce a batch to Kafka
	produceBatch := func(batch []kafkamodels.GMBAccountWorkOrder) {
		if len(batch) == 0 {
			return
		}

		batchWorkOrder := kafkamodels.GMBBatchWorkOrder{
			BatchID:   uuid.New().String(),
			SyncType:  syncType,
			Accounts:  batch,
			CreatedAt: time.Now().UTC(),
		}

		payload, err := json.Marshal(batchWorkOrder)
		if err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Msg("Failed to marshal GMB batch work order")
			batchesFailed++
			return
		}

		if err := producer.Produce(ctx, topicGMBBatch, []byte(batchWorkOrder.BatchID), payload); err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Msg("Failed to produce GMB batch work order to Kafka")
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
			Msg("Produced GMB batch work order")
	}

	// Fetch small chunks from MongoDB using ID-based pagination
	for {
		accounts, err := unifiedRepo.GetAccountsNeedingUpdateByID(ctx, mongomodels.PlatformGMB, nil, gmbUpdateIntervalHours, lastID, gmbMongoFetchSize)
		if err != nil {
			log.Error().Err(err).Str("last_id", lastID.Hex()).Msg("Failed to fetch GMB accounts batch")
			break
		}

		if len(accounts) == 0 {
			break
		}

		// Update lastID for next iteration
		lastID = accounts[len(accounts)-1].ID

		// Validate and accumulate accounts
		validAccounts, skipped := buildGMBAccountBatch(accounts, syncType, log)
		accountsSkipped += skipped
		accumulated = append(accumulated, validAccounts...)

		// Produce to Kafka when we have enough accounts
		for int64(len(accumulated)) >= gmbKafkaBatchSize {
			produceBatch(accumulated[:gmbKafkaBatchSize])
			accumulated = accumulated[gmbKafkaBatchSize:]
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
		Msg("Completed batch production for GMB accounts")
}

// buildGMBAccountBatch validates accounts and builds the batch payload.
// Returns the valid accounts and count of skipped accounts.
func buildGMBAccountBatch(
	accounts []mongomodels.SocialIntegration,
	syncType string,
	log zerolog.Logger,
) ([]kafkamodels.GMBAccountWorkOrder, int) {
	batch := make([]kafkamodels.GMBAccountWorkOrder, 0, len(accounts))
	skipped := 0

	for _, account := range accounts {
		// Get access token
		accessToken := account.GetAccessToken()
		if accessToken == "" {
			accessToken = GetStringFromExtraData(account.ExtraData, "access_token")
		}
		if accessToken == "" {
			log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping GMB account: missing access token")
			skipped++
			continue
		}

		// Get refresh token
		refreshToken := account.RefreshToken
		if refreshToken == "" {
			log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping GMB account: missing refresh token")
			skipped++
			continue
		}

		// Validate workspace ID
		if account.WorkspaceID.IsZero() {
			log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping GMB account: missing workspace_id")
			skipped++
			continue
		}

		// Parse GMB account + location IDs from platform_identifier
		gmbAccountID, locationID, ok := parseGMBPlatformIdentifier(account.PlatformIdentifier)
		if !ok {
			log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping GMB account: invalid platform_identifier format")
			skipped++
			continue
		}

		// Resolve display names
		accountName := account.PlatformName
		locationName := GetStringFromExtraData(account.ExtraData, "location_name")
		if locationName == "" {
			locationName = accountName
		}

		// Language code (default to "en")
		languageCode := account.LanguageCode
		if languageCode == "" {
			languageCode = GetStringFromExtraData(account.ExtraData, "language_code")
		}
		if languageCode == "" {
			languageCode = "en"
		}

		batch = append(batch, kafkamodels.GMBAccountWorkOrder{
			ID:           account.ID.Hex(),
			WorkspaceID:  account.WorkspaceID.Hex(),
			AccountID:    gmbAccountID,
			LocationID:   locationID,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			AccountName:  accountName,
			LocationName: locationName,
			LanguageCode: languageCode,
			SyncType:     syncType,
		})
	}

	return batch, skipped
}

// parseGMBPlatformIdentifier parses a GMB platform identifier in the form
// "accounts/{accountID}/locations/{locationID}" and returns the account ID,
// location ID, and whether parsing succeeded.
func parseGMBPlatformIdentifier(platformIdentifier string) (accountID, locationID string, ok bool) {
	parts := strings.Split(platformIdentifier, "/")
	if len(parts) != 4 || parts[0] != "accounts" || parts[2] != "locations" {
		return "", "", false
	}
	if parts[1] == "" || parts[3] == "" {
		return "", "", false
	}
	return parts[1], parts[3], true
}
