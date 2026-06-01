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

// YouTube batch configuration constants
const (
	// youtubeMongoFetchSize is the number of accounts to fetch per MongoDB query (small to avoid cursor timeout)
	youtubeMongoFetchSize int64 = 50

	// youtubeKafkaBatchSize is the number of accounts per Kafka message (larger for efficient processing)
	youtubeKafkaBatchSize int64 = 200

	youtubeUpdateIntervalHours = 6
	youtubeConsentMaxDays      = 30 // Exclude accounts with consent older than 30 days
	topicYouTubeBatch          = "work-order-youtube"
)

// ProcessYouTubeAccounts fetches YouTube accounts needing update and produces batch work orders.
// Uses small chunk fetching (to avoid cursor timeout) + accumulation pattern for efficient Kafka batches.
// Only produces batches for accounts that haven't been updated in the last 6 hours.
// Excludes accounts where preferences.last_youtube_consent_time is older than 30 days.
func ProcessYouTubeAccounts(ctx context.Context, db *mongo.Database, producer kafka.Producer, logger zerolog.Logger, syncType string) {
	unifiedRepo := mongodb.NewUnifiedSocialRepository(db, logger.With().Str("repository", "unified_social").Logger())

	log := logger.With().Str("platform", "youtube").Str("topic", topicYouTubeBatch).Logger()

	// Get total count for progress logging (with consent filter)
	totalCount, err := unifiedRepo.CountYouTubeAccountsNeedingUpdate(ctx, youtubeUpdateIntervalHours, youtubeConsentMaxDays)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count YouTube accounts needing update")
		return
	}

	if totalCount == 0 {
		log.Info().Msg("No YouTube accounts need update (or all have expired consent)")
		return
	}

	log.Info().
		Int64("total_accounts", totalCount).
		Int64("mongo_fetch_size", youtubeMongoFetchSize).
		Int64("kafka_batch_size", youtubeKafkaBatchSize).
		Int("consent_max_days", youtubeConsentMaxDays).
		Msg("Starting batch processing")

	var (
		lastID          = primitive.NilObjectID // Start from beginning
		batchesCreated  int
		batchesFailed   int
		accountsQueued  int
		accountsSkipped int
	)

	// Accumulator for valid accounts - will be sent to Kafka when reaching kafkaBatchSize
	accumulated := make([]kafkamodels.YouTubeAccountWorkOrder, 0, youtubeKafkaBatchSize)

	// Helper function to produce a batch to Kafka
	produceBatch := func(batch []kafkamodels.YouTubeAccountWorkOrder) {
		if len(batch) == 0 {
			return
		}

		batchWorkOrder := kafkamodels.YouTubeBatchWorkOrder{
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

		if err := producer.Produce(ctx, topicYouTubeBatch, []byte(batchWorkOrder.BatchID), payload); err != nil {
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
		accounts, err := unifiedRepo.GetYouTubeAccountsNeedingUpdateByID(ctx, youtubeUpdateIntervalHours, youtubeConsentMaxDays, lastID, youtubeMongoFetchSize)
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
		validAccounts, skipped := buildYouTubeAccountBatch(accounts, syncType, log)
		accountsSkipped += skipped
		accumulated = append(accumulated, validAccounts...)

		// Produce to Kafka when we have enough accounts (exactly kafkaBatchSize)
		for int64(len(accumulated)) >= youtubeKafkaBatchSize {
			produceBatch(accumulated[:youtubeKafkaBatchSize])
			accumulated = accumulated[youtubeKafkaBatchSize:]
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
		Msg("Completed batch production for YouTube accounts")
}

// buildYouTubeAccountBatch validates accounts and builds the batch payload.
// Returns the valid accounts and count of skipped accounts.
func buildYouTubeAccountBatch(
	accounts []mongomodels.SocialIntegration,
	syncType string,
	log zerolog.Logger,
) ([]kafkamodels.YouTubeAccountWorkOrder, int) {
	batch := make([]kafkamodels.YouTubeAccountWorkOrder, 0, len(accounts))
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

		// Get refresh token for YouTube (needed for token refresh)
		refreshToken := account.RefreshToken
		if refreshToken == "" {
			refreshToken = GetStringFromExtraData(account.ExtraData, "refresh_token")
		}

		// Validate workspace ID
		if account.WorkspaceID.IsZero() {
			log.Debug().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing workspace_id")
			skipped++
			continue
		}

		batch = append(batch, kafkamodels.YouTubeAccountWorkOrder{
			ID:           account.ID.Hex(),
			ChannelID:    account.PlatformIdentifier,
			AccessToken:  accessToken,
			RefreshToken: refreshToken,
			WorkspaceID:  account.WorkspaceID.Hex(),
			SyncType:     syncType,
		})
	}

	return batch, skipped
}
