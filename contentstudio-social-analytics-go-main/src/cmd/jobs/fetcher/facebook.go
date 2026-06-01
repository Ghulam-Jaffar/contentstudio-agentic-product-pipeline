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

// Facebook batch configuration constants
const (
	// facebookMongoFetchSize is the number of accounts to fetch per MongoDB query (small to avoid cursor timeout)
	facebookMongoFetchSize int64 = 50

	// facebookKafkaBatchSize is the number of accounts per Kafka message (larger for efficient processing)
	facebookKafkaBatchSize int64 = 200

	facebookUpdateIntervalHours = 6
	topicFacebookBatch          = "work-order-facebook"
)

// ProcessFacebookAccounts fetches Facebook accounts needing update and produces batch work orders.
// Uses pagination to avoid loading all accounts into memory.
// Only produces batches for accounts that haven't been updated in the last 6 hours.
func ProcessFacebookAccounts(db *mongo.Database, producer kafka.Producer, log zerolog.Logger, accountTypes []string, syncType string) {
	log.Info().Strs("accountTypes", accountTypes).Str("syncType", syncType).Msg("Processing Facebook accounts")

	// Use unified repository
	unifiedRepo := mongodb.NewUnifiedSocialRepository(db, log.With().Str("repository", "unified_social").Logger())

	// Adjust account types to match database casing if necessary
	queryAccountTypes := make([]string, 0, len(accountTypes))
	for _, accountType := range accountTypes {
		switch accountType {
		case "page":
			queryAccountTypes = append(queryAccountTypes, "Page")
		case "group":
			queryAccountTypes = append(queryAccountTypes, "Group")
		default:
			queryAccountTypes = append(queryAccountTypes, accountType)
		}
	}

	if len(queryAccountTypes) != len(accountTypes) {
		log.Info().Strs("original_account_types", accountTypes).Strs("query_account_types", queryAccountTypes).Msg("Adjusted accountTypes for DB query")
	}

	processFacebookBatches(context.Background(), unifiedRepo, producer, log, syncType, queryAccountTypes)

	log.Info().Msg("Completed processing Facebook accounts")
}

// processFacebookBatches fetches accounts in small chunks (to avoid cursor timeout),
// accumulates valid accounts, and produces Kafka messages in larger batches for efficiency.
func processFacebookBatches(
	ctx context.Context,
	repo mongodb.UnifiedSocialRepository,
	producer kafka.Producer,
	logger zerolog.Logger,
	syncType string,
	accountTypes []string,
) {
	log := logger.With().Strs("account_types", accountTypes).Str("topic", topicFacebookBatch).Logger()

	// Get total count for progress logging
	totalCount, err := repo.CountAccountsNeedingUpdate(ctx, mongomodels.PlatformFacebook, accountTypes, facebookUpdateIntervalHours)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count Facebook accounts needing update")
		return
	}

	if totalCount == 0 {
		log.Info().Msg("No Facebook accounts need update")
		return
	}

	log.Info().
		Int64("total_accounts", totalCount).
		Int64("mongo_fetch_size", facebookMongoFetchSize).
		Int64("kafka_batch_size", facebookKafkaBatchSize).
		Msg("Starting batch processing")

	var (
		lastID           = primitive.NilObjectID
		batchesCreated   int
		batchesFailed    int
		accountsQueued   int
		accountsSkipped  int
	)

	accumulated := make([]kafkamodels.FacebookAccountWorkOrder, 0, facebookKafkaBatchSize)

	produceBatch := func(batch []kafkamodels.FacebookAccountWorkOrder) {
		if len(batch) == 0 {
			return
		}

		batchWorkOrder := kafkamodels.FacebookBatchWorkOrder{
			BatchID:   uuid.New().String(),
			SyncType:  syncType,
			Accounts:  batch,
			CreatedAt: time.Now().UTC(),
		}

		payload, err := json.Marshal(batchWorkOrder)
		if err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Msg("Failed to marshal batch work order; skipping batch")
			batchesFailed++
			return
		}

		if err := producer.Produce(ctx, topicFacebookBatch, []byte(batchWorkOrder.BatchID), payload); err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Int("accounts_in_batch", len(batch)).Msg("Failed to produce batch work order; skipping batch")
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

	for {
		accounts, err := repo.GetAccountsNeedingUpdateByID(ctx, mongomodels.PlatformFacebook, accountTypes, facebookUpdateIntervalHours, lastID, facebookMongoFetchSize)
		if err != nil {
			log.Error().Err(err).Str("last_id", lastID.Hex()).Msg("Failed to fetch accounts batch")
			break
		}

		if len(accounts) == 0 {
			break
		}

		lastID = accounts[len(accounts)-1].ID

		validAccounts, skipped := buildFacebookAccountBatch(accounts, syncType, log)
		accountsSkipped += skipped
		accumulated = append(accumulated, validAccounts...)

		for int64(len(accumulated)) >= facebookKafkaBatchSize {
			produceBatch(accumulated[:facebookKafkaBatchSize])
			accumulated = accumulated[facebookKafkaBatchSize:]
		}
	}

	if len(accumulated) > 0 {
		produceBatch(accumulated)
	}

	log.Info().
		Int("batches_created", batchesCreated).
		Int("accounts_queued", accountsQueued).
		Int("accounts_skipped", accountsSkipped).
		Int64("total_eligible", totalCount).
		Msg("Completed batch production for Facebook accounts")
}

// buildFacebookAccountBatch validates accounts and builds the batch payload.
// Returns the valid accounts and count of skipped accounts.
func buildFacebookAccountBatch(
	accounts []mongomodels.SocialIntegration,
	syncType string,
	log zerolog.Logger,
) ([]kafkamodels.FacebookAccountWorkOrder, int) {
	batch := make([]kafkamodels.FacebookAccountWorkOrder, 0, len(accounts))
	skipped := 0

	for _, account := range accounts {
		// Get access token - try long access token first, then regular
		accessToken := GetStringFromExtraData(account.ExtraData, "access_token")
		longAccessToken := account.GetAccessToken()

		if accessToken == "" && longAccessToken == "" {
			log.Debug().Str("facebook_id", account.PlatformIdentifier).Msg("Skipping account: missing access token")
			skipped++
			continue
		}

		// Validate workspace ID
		if account.WorkspaceID.IsZero() {
			log.Debug().Str("facebook_id", account.PlatformIdentifier).Msg("Skipping account: missing workspace_id")
			skipped++
			continue
		}

		batch = append(batch, kafkamodels.FacebookAccountWorkOrder{
			ID:              account.ID.Hex(),
			FacebookID:      account.PlatformIdentifier,
			Type:            account.Type,
			AccessToken:     accessToken,
			WorkspaceID:     account.WorkspaceID.Hex(),
			LongAccessToken: longAccessToken,
			SyncType:        syncType,
		})
	}

	return batch, skipped
}
