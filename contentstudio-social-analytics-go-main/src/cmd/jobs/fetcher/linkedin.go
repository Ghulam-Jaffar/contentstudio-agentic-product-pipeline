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

// Batch configuration constants
const (
	// linkedinMongoFetchSize is the number of accounts to fetch per MongoDB query (small to avoid cursor timeout)
	linkedinMongoFetchSize int64 = 50

	// linkedinKafkaBatchSize is the number of accounts per Kafka message (larger for efficient processing)
	linkedinKafkaBatchSize int64 = 200

	// updateIntervalHours is how old last_analytics_updated_at must be before re-fetching
	updateIntervalHours = 6

	// Kafka topics for batch work orders
	topicLinkedinPageBatch    = "work-order-linkedin-page-batch"
	topicLinkedinProfileBatch = "work-order-linkedin-profile-batch"
)

// accountTypeToKafkaType maps account type strings to Kafka model types and topics
var accountTypeConfig = map[string]struct {
	kafkaType kafkamodels.LinkedinAccountType
	topic     string
}{
	mongomodels.TypeProfile: {kafkamodels.LinkedinAccountTypeProfile, topicLinkedinProfileBatch},
	mongomodels.TypePage:    {kafkamodels.LinkedinAccountTypePage, topicLinkedinPageBatch},
}

// ProcessLinkedinAccounts fetches LinkedIn accounts needing update and produces batch work orders.
// Uses pagination to avoid loading all accounts into memory.
// Only produces batches for accounts that haven't been updated in the last 6 hours.
// accountTypes is a slice of account types to process (e.g., ["profile", "page"]).
// If empty, all supported account types are processed.
func ProcessLinkedinAccounts(ctx context.Context, db *mongo.Database, producer kafka.Producer, logger zerolog.Logger, accountTypes []string, syncType string) {
	unifiedRepo := mongodb.NewUnifiedSocialRepository(db, logger.With().Str("repository", "unified_social").Logger())

	// If no account types specified, process all supported types
	if len(accountTypes) == 0 {
		accountTypes = []string{mongomodels.TypePage}
	}

	// Normalize account types to match database casing (e.g., "page" -> "Page", "profile" -> "Profile")
	normalizedAccountTypes := make([]string, 0, len(accountTypes))
	for _, accountType := range accountTypes {
		normalized := normalizeLinkedinAccountType(accountType)
		normalizedAccountTypes = append(normalizedAccountTypes, normalized)
	}

	if len(normalizedAccountTypes) != len(accountTypes) {
		logger.Info().Strs("original_account_types", accountTypes).Strs("normalized_account_types", normalizedAccountTypes).Msg("Adjusted accountTypes for DB query")
	}

	logger.Info().Strs("account_types", normalizedAccountTypes).Msg("Starting LinkedIn account processing")

	for _, accountType := range normalizedAccountTypes {
		config, ok := accountTypeConfig[accountType]
		if !ok {
			logger.Warn().Str("account_type", accountType).Msg("Unknown account type, skipping")
			continue
		}
		processLinkedinBatches(ctx, unifiedRepo, producer, logger, syncType, accountType, config.topic, config.kafkaType)
	}

	logger.Info().Msg("Completed processing LinkedIn accounts")
}

// processLinkedinBatches fetches accounts in small chunks (to avoid cursor timeout),
// accumulates valid accounts, and produces Kafka messages in larger batches for efficiency.
func processLinkedinBatches(
	ctx context.Context,
	repo mongodb.UnifiedSocialRepository,
	producer kafka.Producer,
	logger zerolog.Logger,
	syncType string,
	accountType string,
	topic string,
	kafkaAccountType kafkamodels.LinkedinAccountType,
) {
	log := logger.With().Str("account_type", accountType).Str("topic", topic).Logger()

	// Get total count for progress logging (single account type as slice)
	accountTypes := []string{accountType}
	totalCount, err := repo.CountAccountsNeedingUpdate(ctx, mongomodels.PlatformLinkedIn, accountTypes, updateIntervalHours)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count LinkedIn accounts needing update")
		return
	}

	if totalCount == 0 {
		log.Info().Msg("No LinkedIn accounts need update")
		return
	}

	log.Info().
		Int64("total_accounts", totalCount).
		Int64("mongo_fetch_size", linkedinMongoFetchSize).
		Int64("kafka_batch_size", linkedinKafkaBatchSize).
		Msg("Starting batch processing")

	var (
		lastID          = primitive.NilObjectID
		batchesCreated  int
		batchesFailed   int
		accountsQueued  int
		accountsSkipped int
	)

	accumulated := make([]kafkamodels.LinkedinAccountWorkOrder, 0, linkedinKafkaBatchSize)

	produceBatch := func(batch []kafkamodels.LinkedinAccountWorkOrder) {
		if len(batch) == 0 {
			return
		}

		batchWorkOrder := kafkamodels.LinkedinBatchWorkOrder{
			BatchID:     uuid.New().String(),
			SyncType:    syncType,
			AccountType: kafkaAccountType,
			Accounts:    batch,
			CreatedAt:   time.Now().UTC(),
		}

		payload, err := json.Marshal(batchWorkOrder)
		if err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Msg("Failed to marshal batch work order; skipping batch")
			batchesFailed++
			return
		}

		if err := producer.Produce(ctx, topic, []byte(batchWorkOrder.BatchID), payload); err != nil {
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
		accounts, err := repo.GetAccountsNeedingUpdateByID(ctx, mongomodels.PlatformLinkedIn, accountTypes, updateIntervalHours, lastID, linkedinMongoFetchSize)
		if err != nil {
			log.Error().Err(err).Str("last_id", lastID.Hex()).Msg("Failed to fetch accounts batch")
			break
		}

		if len(accounts) == 0 {
			break
		}

		lastID = accounts[len(accounts)-1].ID

		validAccounts, skipped := buildAccountBatch(accounts, syncType, kafkaAccountType, log)
		accountsSkipped += skipped
		accumulated = append(accumulated, validAccounts...)

		for int64(len(accumulated)) >= linkedinKafkaBatchSize {
			produceBatch(accumulated[:linkedinKafkaBatchSize])
			accumulated = accumulated[linkedinKafkaBatchSize:]
		}
	}

	if len(accumulated) > 0 {
		produceBatch(accumulated)
	}

	log.Info().
		Int("batches_created", batchesCreated).
		Int("batches_failed", batchesFailed).
		Int("accounts_queued", accountsQueued).
		Int("accounts_skipped", accountsSkipped).
		Int64("total_eligible", totalCount).
		Msg("Completed batch production for LinkedIn accounts")
}

// buildAccountBatch validates accounts and builds the batch payload.
// Returns the valid accounts and count of skipped accounts.
func buildAccountBatch(
	accounts []mongomodels.SocialIntegration,
	syncType string,
	kafkaAccountType kafkamodels.LinkedinAccountType,
	log zerolog.Logger,
) ([]kafkamodels.LinkedinAccountWorkOrder, int) {
	batch := make([]kafkamodels.LinkedinAccountWorkOrder, 0, len(accounts))
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

		batch = append(batch, kafkamodels.LinkedinAccountWorkOrder{
			ID:          account.ID.Hex(),
			WorkspaceID: account.WorkspaceID.Hex(),
			LinkedinID:  account.PlatformIdentifier,
			AccessToken: accessToken,
			SyncType:    syncType,
			AccountType: kafkaAccountType,
		})
	}

	return batch, skipped
}

// normalizeLinkedinAccountType converts account type strings to match database casing.
// Handles case-insensitive input (e.g., "page" -> "Page", "profile" -> "Profile").
func normalizeLinkedinAccountType(accountType string) string {
	switch strings.ToLower(accountType) {
	case "page":
		return mongomodels.TypePage
	case "profile":
		return mongomodels.TypeProfile
	default:
		return accountType
	}
}
