package fetcher

import (
	"context"
	"encoding/json"
	"fmt"
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

// Twitter batch configuration constants
const (
	// twitterMongoFetchSize is the number of accounts to fetch per MongoDB query (small to avoid cursor timeout)
	twitterMongoFetchSize int64 = 50

	// twitterKafkaBatchSize is the number of accounts per Kafka message (larger for efficient processing)
	twitterKafkaBatchSize int64 = 200

	twitterUpdateIntervalHours = 6
	topicTwitterBatch          = "work-order-twitter-batch"
)

// ProcessTwitterAccounts fetches Twitter accounts needing update and produces batch work orders.
// Uses pagination to avoid loading all accounts into memory.
// Only produces batches for accounts that haven't been updated in the last 6 hours.
func ProcessTwitterAccounts(ctx context.Context, db *mongo.Database, producer kafka.Producer, logger zerolog.Logger, accountTypes []string, syncType string) {
	log := logger.With().Str("social_network", "twitter").Logger()
	log.Info().Str("sync_type", "incremental").Msg("Processing Twitter accounts")

	// Use unified repository
	unifiedRepo := mongodb.NewUnifiedSocialRepository(db, log.With().Str("repository", "unified_social").Logger())

	processTwitterBatches(ctx, db, unifiedRepo, producer, log, "incremental")

	log.Info().Msg("Completed processing Twitter accounts")
}

// processTwitterBatches fetches accounts in small chunks (to avoid cursor timeout),
// accumulates valid accounts, and produces Kafka messages in larger batches for efficiency.
func processTwitterBatches(
	ctx context.Context,
	db *mongo.Database,
	repo mongodb.UnifiedSocialRepository,
	producer kafka.Producer,
	logger zerolog.Logger,
	syncType string,
) {
	log := logger.With().Str("topic", topicTwitterBatch).Logger()
	now := time.Now().UTC()
	twitterRepo := mongodb.NewTwitterRepository(db)

	developerApps, err := twitterRepo.GetAnalyticsEnabledDeveloperApps(ctx)
	if err != nil {
		log.Error().Err(err).Msg("Failed to load developer apps")
		return
	}
	if len(developerApps) == 0 {
		log.Warn().Msg("No analytics-enabled developer apps found; accounts with developer_app_id will be skipped")
	}

	// Get total count for progress logging.
	// Intentionally do not filter by account type for Twitter.
	accountTypes := []string{}
	totalCount, err := repo.CountAccountsNeedingUpdate(ctx, mongomodels.PlatformTwitter, accountTypes, twitterUpdateIntervalHours)
	if err != nil {
		log.Error().Err(err).Msg("Failed to count Twitter accounts needing update")
		return
	}

	if totalCount == 0 {
		log.Info().Msg("No Twitter accounts need update")
		return
	}

	log.Info().
		Int64("total_accounts", totalCount).
		Int64("total_twitter_accounts", totalCount).
		Int64("mongo_fetch_size", twitterMongoFetchSize).
		Int64("kafka_batch_size", twitterKafkaBatchSize).
		Msg("Starting Twitter batch account processing")

	var (
		lastID          = primitive.NilObjectID // Start from beginning
		batchesCreated  int
		batchesFailed   int
		accountsQueued  int
		accountsSkipped int
	)

	// Accumulator for valid accounts - will be sent to Kafka when reaching kafkaBatchSize
	accumulated := make([]kafkamodels.TwitterAccountWorkOrder, 0, twitterKafkaBatchSize)

	// Helper function to produce a batch to Kafka
	produceBatch := func(batch []kafkamodels.TwitterAccountWorkOrder) {
		if len(batch) == 0 {
			return
		}

		batchWorkOrder := kafkamodels.TwitterBatchWorkOrder{
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

		if err := producer.Produce(ctx, topicTwitterBatch, []byte(batchWorkOrder.BatchID), payload); err != nil {
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
		accounts, err := repo.GetAccountsNeedingUpdateByID(ctx, mongomodels.PlatformTwitter, accountTypes, twitterUpdateIntervalHours, lastID, twitterMongoFetchSize)
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
		validAccounts, skipped, stats, err := buildTwitterAccountBatch(ctx, twitterRepo, accounts, syncType, log, developerApps, now)
		if err != nil {
			log.Error().Err(err).Msg("Failed to build Twitter account batch")
			break
		}
		log.Info().
			Int("twitter_accounts_total_in_chunk", len(accounts)).
			Int("twitter_accounts_total_input", stats.TotalInput).
			Int("twitter_accounts_missing_oauth_token", stats.MissingOAuthToken).
			Int("twitter_accounts_missing_oauth_token_secret", stats.MissingOAuthTokenSecret).
			Int("twitter_accounts_missing_platform_identifier", stats.MissingPlatformIdentifier).
			Int("twitter_accounts_missing_workspace_id", stats.MissingWorkspaceID).
			Int("twitter_accounts_missing_developer_app_id", stats.MissingDeveloperAppID).
			Int("twitter_accounts_non_eligible_developer_app", stats.DeveloperAppNotEligible).
			Int("twitter_accounts_missing_job_settings", stats.MissingJobSettings).
			Int("twitter_accounts_schedule_skipped", stats.ScheduleSkipped).
			Int("twitter_accounts_produced_workorders", stats.ProducedWorkOrders).
			Int("twitter_accounts_skipped_total", skipped).
			Msg("Twitter account filtering stats for fetched chunk")
		accountsSkipped += skipped
		accumulated = append(accumulated, validAccounts...)
		log.Info().
			Int("twitter_accounts_accumulated_for_kafka", len(accumulated)).
			Msg("Accumulated Twitter work orders pending Kafka production")

		// Produce to Kafka when we have enough accounts (exactly kafkaBatchSize)
		for int64(len(accumulated)) >= twitterKafkaBatchSize {
			produceBatch(accumulated[:twitterKafkaBatchSize])
			accumulated = accumulated[twitterKafkaBatchSize:]
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
		Msg("Completed batch production for Twitter accounts")
}

// buildTwitterAccountBatch validates accounts and builds the batch payload.
// Returns the valid accounts and count of skipped accounts.
func buildTwitterAccountBatch(
	ctx context.Context,
	twitterRepo *mongodb.TwitterRepository,
	accounts []mongomodels.SocialIntegration,
	syncType string,
	log zerolog.Logger,
	developerApps map[string]mongodb.TwitterDeveloperApp,
	now time.Time,
) ([]kafkamodels.TwitterAccountWorkOrder, int, twitterBatchFilterStats, error) {
	batch := make([]kafkamodels.TwitterAccountWorkOrder, 0, len(accounts))
	stats := twitterBatchFilterStats{TotalInput: len(accounts)}
	skipped := 0
	filtered := make([]twitterFilteredAccount, 0, len(accounts))
	platformIDs := make([]string, 0, len(accounts))

	for _, account := range accounts {
		// For Twitter, use OAuth token and secret
		oauthToken := account.OAuthToken
		if oauthToken == "" {
			log.Warn().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing oauth_token")
			stats.MissingOAuthToken++
			skipped++
			continue
		}

		oauthTokenSecret := account.OAuthTokenSecret
		if oauthTokenSecret == "" {
			log.Warn().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing oauth_token_secret")
			stats.MissingOAuthTokenSecret++
			skipped++
			continue
		}

		// Get Twitter ID from PlatformIdentifier or TwitterID field
		twitterID := account.PlatformIdentifier
		if twitterID == "" {
			log.Warn().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing platform identifier")
			stats.MissingPlatformIdentifier++
			skipped++
			continue
		}

		// Validate workspace ID
		if account.WorkspaceID.IsZero() {
			log.Warn().Str("platform_identifier", account.PlatformIdentifier).Msg("Skipping account: missing workspace_id")
			stats.MissingWorkspaceID++
			skipped++
			continue
		}

		if strings.TrimSpace(account.DeveloperAppID) == "" {
			stats.MissingDeveloperAppID++
			skipped++
			continue
		}

		if _, ok := developerApps[account.DeveloperAppID]; !ok {
			log.Warn().
				Str("platform_identifier", account.PlatformIdentifier).
				Str("developer_app_id", account.DeveloperAppID).
				Msg("Skipping account: non-null developer_app_id has no matching analytics-enabled developer app")
			stats.DeveloperAppNotEligible++
			skipped++
			continue
		}

		filtered = append(filtered, twitterFilteredAccount{
			Account:          account,
			OAuthToken:       oauthToken,
			OAuthTokenSecret: oauthTokenSecret,
		})
		platformIDs = append(platformIDs, account.PlatformIdentifier)
	}

	jobSettings, err := twitterRepo.GetJobSettingsByPlatformIDs(ctx, platformIDs)
	if err != nil {
		return nil, skipped, stats, fmt.Errorf("buildTwitterAccountBatch: load twitter job settings: %w", err)
	}

	for _, item := range filtered {
		account := item.Account
		app := developerApps[account.DeveloperAppID]
		jobSetting, ok := jobSettings[account.PlatformIdentifier]
		if !ok {
			log.Warn().
				Str("platform_identifier", account.PlatformIdentifier).
				Msg("Skipping account: missing twitter_job_settings entry for account that passed developer_app check")
			stats.MissingJobSettings++
			skipped++
			continue
		}

		if !shouldScheduleTwitterAccount(jobSetting, now) {
			stats.ScheduleSkipped++
			skipped++
			continue
		}

		batch = append(batch, kafkamodels.TwitterAccountWorkOrder{
			ID:               account.ID.Hex(),
			WorkspaceID:      account.WorkspaceID.Hex(),
			TwitterID:        account.PlatformIdentifier,
			OAuthToken:       item.OAuthToken,
			OAuthTokenSecret: item.OAuthTokenSecret,
			PostCount:        jobSetting.PostCount,
			APIKey:           app.APIKey,
			APISecret:        app.APISecret,
			AppName:          app.AppName,
			AppID:            app.ID.Hex(),
			ExecutedBy:       "internal",
			SyncType:         syncType,
		})
		stats.ProducedWorkOrders++
	}

	return batch, skipped, stats, nil
}

type twitterFilteredAccount struct {
	Account          mongomodels.SocialIntegration
	OAuthToken       string
	OAuthTokenSecret string
}

type twitterBatchFilterStats struct {
	TotalInput                int
	MissingOAuthToken         int
	MissingOAuthTokenSecret   int
	MissingPlatformIdentifier int
	MissingWorkspaceID        int
	MissingDeveloperAppID     int
	DeveloperAppNotEligible   int
	MissingJobSettings        int
	ScheduleSkipped           int
	ProducedWorkOrders        int
}

func shouldScheduleTwitterAccount(setting mongodb.TwitterJobSetting, now time.Time) bool {
	switch strings.ToLower(setting.JobType) {
	case "daily":
		return true
	case "weekly":
		return setting.TriggerDay == weekdayToOneBased(now.Weekday())
	case "monthly":
		return setting.TriggerDay == now.Day()
	case "never":
		return false
	default:
		return false
	}
}

func weekdayToOneBased(day time.Weekday) int {
	if day == time.Sunday {
		return 7
	}
	return int(day)
}
