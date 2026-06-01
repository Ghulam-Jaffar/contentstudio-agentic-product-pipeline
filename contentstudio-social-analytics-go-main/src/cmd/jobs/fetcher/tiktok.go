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

// TikTok batch configuration constants
const (
	// tiktokMongoFetchSize is the number of accounts to fetch per MongoDB query (small to avoid cursor timeout)
	tiktokMongoFetchSize int64 = 50

	// tiktokKafkaBatchSize is the number of accounts per Kafka message (larger for efficient processing)
	tiktokKafkaBatchSize int64 = 200

	tiktokUpdateIntervalHours = 6
	topicTikTokBatch          = "work-order-tiktok-batch"
)

// ProcessTikTokAccounts fetches TikTok accounts needing update and produces batch work orders.
// Uses pagination to avoid loading all accounts into memory.
// Only produces batches for accounts that haven't been updated in the last 6 hours.
func ProcessTikTokAccounts(ctx context.Context, db *mongo.Database, producer kafka.Producer, logger zerolog.Logger, accountTypes []string, syncType string) {
	log := logger.With().Str("social_network", "tiktok").Logger()
	log.Info().Str("sync_type", syncType).Msg("Processing TikTok accounts")

	// Use unified repository
	unifiedRepo := mongodb.NewUnifiedSocialRepository(db, log.With().Str("repository", "unified_social").Logger())

	processTikTokBatches(ctx, unifiedRepo, producer, log, syncType)

	log.Info().Msg("Completed processing TikTok accounts")
}

// processTikTokBatches fetches accounts in small chunks (to avoid cursor timeout),
// accumulates valid accounts, and produces Kafka messages in larger batches for efficiency.
func processTikTokBatches(
	ctx context.Context,
	repo mongodb.UnifiedSocialRepository,
	producer kafka.Producer,
	logger zerolog.Logger,
	syncType string,
) {
	log := logger.With().Str("topic", topicTikTokBatch).Logger()

	// Get total count for progress logging
	accountTypes := []string{mongomodels.TypeProfile, "profile"} // TikTok only has profile type (accept both cases)
	totalCount, err := repo.CountAccountsNeedingUpdate(ctx, mongomodels.PlatformTikTok, accountTypes, tiktokUpdateIntervalHours)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count TikTok accounts needing update")
		return
	}

	if totalCount == 0 {
		log.Info().Msg("No TikTok accounts need update")
		return
	}

	log.Info().
		Int64("total_accounts", totalCount).
		Int64("mongo_fetch_size", tiktokMongoFetchSize).
		Int64("kafka_batch_size", tiktokKafkaBatchSize).
		Msg("Starting batch processing")

	var (
		lastID          = primitive.NilObjectID // Start from beginning
		batchesCreated  int
		batchesFailed   int
		accountsQueued  int
		accountsSkipped int
	)

	// Accumulator for valid accounts - will be sent to Kafka when reaching kafkaBatchSize
	accumulated := make([]kafkamodels.TikTokAccountWorkOrder, 0, tiktokKafkaBatchSize)

	// Helper function to produce a batch to Kafka
	produceBatch := func(batch []kafkamodels.TikTokAccountWorkOrder) {
		if len(batch) == 0 {
			return
		}

		batchWorkOrder := kafkamodels.TikTokBatchWorkOrder{
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

		if err := producer.Produce(ctx, topicTikTokBatch, []byte(batchWorkOrder.BatchID), payload); err != nil {
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
		accounts, err := repo.GetAccountsNeedingUpdateByID(ctx, mongomodels.PlatformTikTok, accountTypes, tiktokUpdateIntervalHours, lastID, tiktokMongoFetchSize)
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
		validAccounts, skipped := buildTikTokAccountBatch(accounts, syncType, log)
		accountsSkipped += skipped
		accumulated = append(accumulated, validAccounts...)

		// Produce to Kafka when we have enough accounts (exactly kafkaBatchSize)
		for int64(len(accumulated)) >= tiktokKafkaBatchSize {
			produceBatch(accumulated[:tiktokKafkaBatchSize])
			accumulated = accumulated[tiktokKafkaBatchSize:]
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
		Msg("Completed batch production for TikTok accounts")
}

// buildTikTokAccountBatch validates accounts and builds the batch payload.
// Returns the valid accounts and count of skipped accounts.
func buildTikTokAccountBatch(
	accounts []mongomodels.SocialIntegration,
	syncType string,
	log zerolog.Logger,
) ([]kafkamodels.TikTokAccountWorkOrder, int) {
	batch := make([]kafkamodels.TikTokAccountWorkOrder, 0, len(accounts))
	skipped := 0

	for _, account := range accounts {
		// For TikTok, use direct fields first
		accessToken := account.AccessToken
		if accessToken == "" {
			accessToken = account.GetAccessToken()
		}
		if accessToken == "" {
			log.Warn().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing access token")
			skipped++
			continue
		}

		// Get refresh token directly
		refreshToken := account.RefreshToken
		if refreshToken == "" {
			log.Warn().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing refresh token")
			skipped++
			continue
		}

		// Get scopes directly
		scope := account.Scope
		if scope == "" {
			log.Warn().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing scope")
			skipped++
			continue
		}

		// Validate workspace ID
		if account.WorkspaceID.IsZero() {
			log.Warn().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing workspace_id")
			skipped++
			continue
		}

		batch = append(batch, kafkamodels.TikTokAccountWorkOrder{
			ID:           account.ID.Hex(),
			WorkspaceID:  account.WorkspaceID.Hex(),
			TikTokID:     account.PlatformIdentifier,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			Scope:        scope,
			SyncType:     syncType,
		})
	}

	return batch, skipped
}
