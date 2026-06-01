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

// Instagram batch configuration constants
const (
	// instagramMongoFetchSize is the number of accounts to fetch per MongoDB query (small to avoid cursor timeout)
	instagramMongoFetchSize int64 = 50

	// instagramKafkaBatchSize is the number of accounts per Kafka message (larger for efficient processing)
	instagramKafkaBatchSize int64 = 200

	// instagramUpdateIntervalHours is how old last_analytics_updated_at must be before re-fetching
	instagramUpdateIntervalHours = 6

	// Kafka topic for work orders
	topicInstagram = "work-order-instagram"
)

// ProcessInstagramAccounts fetches Instagram accounts needing update and produces batch work orders.
// Uses pagination to avoid loading all accounts into memory.
// Only produces batches for accounts that haven't been updated in the last 6 hours.
func ProcessInstagramAccounts(
	ctx context.Context,
	db *mongo.Database,
	producer kafka.Producer,
	logger zerolog.Logger,
	accountTypes []string,
	syncType string,
) {
	unifiedRepo := mongodb.NewUnifiedSocialRepository(
		db,
		logger.With().Str("repository", "unified_social").Logger(),
	)

	logger.Info().Msg("Starting Instagram account processing with batch pattern")

	processInstagramBatches(ctx, unifiedRepo, producer, logger, syncType)

	logger.Info().Msg("Completed processing Instagram accounts")
}

// processInstagramBatches fetches accounts in small chunks (to avoid cursor timeout),
// accumulates valid accounts, and produces Kafka messages in larger batches for efficiency.
// Deduplicates by Instagram ID to prevent duplicate insights when same account is in multiple workspaces.
func processInstagramBatches(
	ctx context.Context,
	repo mongodb.UnifiedSocialRepository,
	producer kafka.Producer,
	logger zerolog.Logger,
	syncType string,
) {
	log := logger.With().Str("platform", "instagram").Str("topic", topicInstagram).Logger()

	// Get total count for progress logging
	totalCount, err := repo.CountAccountsNeedingUpdate(
		ctx,
		mongomodels.PlatformInstagram,
		nil, // Instagram doesn't filter by account type
		instagramUpdateIntervalHours,
	)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count Instagram accounts needing update")
		return
	}

	if totalCount == 0 {
		log.Info().Msg("No Instagram accounts need update")
		return
	}

	log.Info().
		Int64("total_accounts", totalCount).
		Int64("mongo_fetch_size", instagramMongoFetchSize).
		Int64("kafka_batch_size", instagramKafkaBatchSize).
		Msg("Starting batch processing")

	var (
		lastID            = primitive.NilObjectID
		batchesCreated    int
		batchesFailed     int
		accountsQueued    int
		accountsSkipped   int
		duplicatesSkipped int
	)

	// Track seen Instagram IDs to prevent duplicates across workspaces
	seenInstagramIDs := make(map[string]bool)

	// Accumulator for valid accounts - will be sent to Kafka when reaching kafkaBatchSize
	accumulated := make([]kafkamodels.InstagramAccountWorkOrder, 0, instagramKafkaBatchSize)

	produceBatch := func(batch []kafkamodels.InstagramAccountWorkOrder) {
		if len(batch) == 0 {
			return
		}

		batchWorkOrder := kafkamodels.InstagramBatchWorkOrder{
			BatchID:   uuid.New().String(),
			SyncType:  syncType,
			Accounts:  batch,
			CreatedAt: time.Now().UTC(),
		}

		payload, err := json.Marshal(batchWorkOrder)
		if err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Msg("Failed to marshal batch; skipping batch")
			batchesFailed++
			return
		}

		if err := producer.Produce(ctx, topicInstagram, []byte(batchWorkOrder.BatchID), payload); err != nil {
			log.Error().Err(err).Str("batch_id", batchWorkOrder.BatchID).Int("accounts_in_batch", len(batch)).Msg("Failed to produce batch; skipping batch")
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
		accounts, err := repo.GetAccountsNeedingUpdateByID(
			ctx,
			mongomodels.PlatformInstagram,
			nil,
			instagramUpdateIntervalHours,
			lastID,
			instagramMongoFetchSize,
		)
		if err != nil {
			log.Error().Err(err).Str("last_id", lastID.Hex()).Msg("Failed to fetch accounts batch")
			break
		}

		if len(accounts) == 0 {
			break
		}

		lastID = accounts[len(accounts)-1].ID

		validAccounts, skipped, duplicates := buildInstagramAccountBatch(accounts, syncType, seenInstagramIDs, log)
		accountsSkipped += skipped
		duplicatesSkipped += duplicates
		accumulated = append(accumulated, validAccounts...)

		for int64(len(accumulated)) >= instagramKafkaBatchSize {
			produceBatch(accumulated[:instagramKafkaBatchSize])
			accumulated = accumulated[instagramKafkaBatchSize:]
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
		Int("duplicates_skipped", duplicatesSkipped).
		Int("unique_instagram_ids", len(seenInstagramIDs)).
		Int64("total_eligible", totalCount).
		Msg("Completed batch production for Instagram accounts")
}

// buildInstagramAccountBatch validates accounts and builds the batch payload.
// Deduplicates by Instagram ID - same account in multiple workspaces only fetched once.
// Returns the valid accounts, count of skipped accounts, and count of duplicates.
func buildInstagramAccountBatch(
	accounts []mongomodels.SocialIntegration,
	syncType string,
	seenInstagramIDs map[string]bool,
	log zerolog.Logger,
) ([]kafkamodels.InstagramAccountWorkOrder, int, int) {
	batch := make([]kafkamodels.InstagramAccountWorkOrder, 0, len(accounts))
	skipped := 0
	duplicates := 0

	for _, account := range accounts {
		// Skip if we've already seen this Instagram ID (same account in different workspace)
		if seenInstagramIDs[account.PlatformIdentifier] {
			duplicates++
			continue
		}

		connectedViaIG := GetBoolFromExtraData(account.ExtraData, "connected_via_instagram")

		// Get access token - check multiple possible locations
		accessToken := getInstagramAccessToken(account, connectedViaIG)
		if accessToken == "" {
			log.Debug().
				Str("platform_identifier", account.PlatformIdentifier).
				Msg("Skipping account: missing access token")
			skipped++
			continue
		}

		// Validate workspace ID
		if account.WorkspaceID.IsZero() {
			log.Debug().
				Str("platform_identifier", account.PlatformIdentifier).
				Msg("Skipping account: missing workspace_id")
			skipped++
			continue
		}

		// Mark this Instagram ID as seen
		seenInstagramIDs[account.PlatformIdentifier] = true

		batch = append(batch, kafkamodels.InstagramAccountWorkOrder{
			ID:                    account.ID.Hex(),
			WorkspaceID:           account.WorkspaceID.Hex(),
			InstagramID:           account.PlatformIdentifier,
			AccessToken:           accessToken,
			ConnectedViaInstagram: connectedViaIG,
			SyncType:              syncType,
		})
	}

	return batch, skipped, duplicates
}

// getInstagramAccessToken extracts the access token from an Instagram account.
// Handles both direct Instagram connections and Facebook-linked accounts.
func getInstagramAccessToken(account mongomodels.SocialIntegration, connectedViaIG bool) string {
	if connectedViaIG {
		// Direct Instagram connection
		if account.AccessToken != "" {
			return account.AccessToken
		}
		return GetStringFromExtraData(account.ExtraData, "access_token")
	}

	// Facebook-linked Instagram account
	// Check nested user_details first
	if account.UserDetails != nil {
		if details, ok := account.UserDetails.(map[string]interface{}); ok {
			if token, exists := details["access_token"].(string); exists && token != "" {
				return token
			}
		}
	}

	// Fallback to main access token field
	return account.GetAccessToken()
}
