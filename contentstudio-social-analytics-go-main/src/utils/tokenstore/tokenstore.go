package tokenstore

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	mongodriver "go.mongodb.org/mongo-driver/mongo"
)

const (
	// PlatformFacebook constant for Facebook platform
	PlatformFacebook = "facebook"
	// PlatformInstagram constant for Instagram platform
	PlatformInstagram = "instagram"
	// redisInsertBatchSize controls how many tokens are inserted per Redis SAdd call.
	redisInsertBatchSize = 50
)

// TokenData represents the structure of token data stored in Redis
type TokenData struct {
	PlatformID string `json:"platform_id"`
	Token      string `json:"token"`
}

// TokenStore handles token management between MongoDB and Redis
type TokenStore struct {
	platform    string
	queueName   string
	redisClient *redis.Client
	mongoDB     *mongodriver.Database
	log         *logger.Logger
}

// NewTokenStore creates a new TokenStore instance
func NewTokenStore(
	platform string,
	redisClient *redis.Client,
	mongoDB *mongodriver.Database,
	log *logger.Logger,
) *TokenStore {
	return &TokenStore{
		platform:    platform,
		queueName:   fmt.Sprintf("%s_valid_token_set", platform),
		redisClient: redisClient,
		mongoDB:     mongoDB,
		log:         log,
	}
}

// GetValidAccounts fetches valid platform accounts from MongoDB collection
func (ts *TokenStore) GetValidAccounts(ctx context.Context) ([]mongo.SocialIntegration, error) {
	ts.log.Info().
		Str("platform", ts.platform).
		Msg("Fetching valid accounts from MongoDB")

	collection := ts.mongoDB.Collection("social_integrations")

	var filter bson.M

	switch ts.platform {
	case PlatformFacebook:
		filter = bson.M{
			"platform_type": "facebook",
			"validity":      "valid",
			"type":          "Page",
		}
	case PlatformInstagram:
		filter = bson.M{
			"platform_type": "instagram",
			"validity":      "valid",
			"facebook_page_id": bson.M{
				"$exists": true,
				"$ne":     nil,
			},
		}
	default:
		return nil, fmt.Errorf("TokenStore.GetValidAccounts: unsupported platform: %s", ts.platform)
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		ts.log.Warn().
			Err(err).
			Str("platform", ts.platform).
			Msg("Error fetching valid accounts from MongoDB")
		return nil, fmt.Errorf("TokenStore.GetValidAccounts: failed to fetch valid accounts: %w", err)
	}
	defer cursor.Close(ctx)

	var dbAccounts []mongo.DBSocialIntegration
	decodedCount := 0
	skippedCount := 0

	// Decode documents one by one to catch individual failures
	for cursor.Next(ctx) {
		var dbAccount mongo.DBSocialIntegration
		if err := cursor.Decode(&dbAccount); err != nil {
			skippedCount++
			ts.log.Warn().
				Err(err).
				Int("skipped_count", skippedCount).
				Msg("Failed to decode account document, skipping")
			continue
		}
		dbAccounts = append(dbAccounts, dbAccount)
		decodedCount++
	}

	if err := cursor.Err(); err != nil {
		ts.log.Warn().
			Err(err).
			Str("platform", ts.platform).
			Msg("Error iterating MongoDB cursor")
		return nil, fmt.Errorf("TokenStore.GetValidAccounts: failed to iterate cursor: %w", err)
	}

	ts.log.Info().
		Int("decoded_count", decodedCount).
		Int("skipped_count", skippedCount).
		Str("platform", ts.platform).
		Msg("Finished decoding MongoDB accounts")

	// Convert DB models to SocialIntegration models
	accounts := make([]mongo.SocialIntegration, 0, len(dbAccounts))
	for _, dbAccount := range dbAccounts {
		accounts = append(accounts, mongo.ConvertDBToSocialIntegration(dbAccount))
	}

	ts.log.Info().
		Int("account_count", len(accounts)).
		Str("platform", ts.platform).
		Msg("Successfully fetched valid accounts from MongoDB")

	return accounts, nil
}

// PopulateRedisSet loads the current valid token set from MongoDB.
// MongoDB validity is the only source of truth while debug-token validation is disabled.
func (ts *TokenStore) PopulateRedisSet(ctx context.Context) (map[string]TokenData, error) {
	ts.log.Info().
		Str("platform", ts.platform).
		Msg("Building token set from MongoDB valid accounts")

	tokenSet := make(map[string]TokenData)

	// Get valid accounts from MongoDB
	accounts, err := ts.GetValidAccounts(ctx)
	if err != nil {
		return nil, err
	}

	// Add MongoDB accounts to token set
	for _, account := range accounts {
		// Use platform_identifier if available, otherwise fall back to legacy field
		platformID := account.PlatformIdentifier
		if platformID == "" {
			switch ts.platform {
			case PlatformFacebook:
				platformID = account.FacebookID
			case PlatformInstagram:
				platformID = account.InstagramID
			}
		}

		if platformID == "" {
			ts.log.Warn().
				Str("account_id", account.ID.Hex()).
				Str("platform", ts.platform).
				Msg("Skipping account with no platform identifier")
			continue
		}

		tokenData := TokenData{
			PlatformID: platformID,
			Token:      account.AccessToken,
		}

		tokenJSON, err := json.Marshal(tokenData)
		if err != nil {
			ts.log.Warn().
				Err(err).
				Str("platform_id", platformID).
				Msg("Error marshaling token data")
			continue
		}

		tokenSet[string(tokenJSON)] = tokenData
	}

	ts.log.Info().
		Int("total_tokens", len(tokenSet)).
		Int("mongo_accounts", len(accounts)).
		Msg("Finished populating token set")

	return tokenSet, nil
}

// Validate refreshes Redis from the MongoDB-derived token set.
// Runtime token validation via the debug-token endpoint is intentionally disabled.
func (ts *TokenStore) Validate(ctx context.Context, tokenSet map[string]TokenData) error {
	ts.log.Info().
		Int("token_count", len(tokenSet)).
		Msg("Starting token sync to Redis")

	if err := ts.redisClient.Del(ctx, ts.queueName).Err(); err != nil {
		ts.log.Warn().
			Err(err).
			Str("queue_name", ts.queueName).
			Msg("Error clearing Redis token set before refresh")
		return err
	}

	insertedCount := 0
	batchJSON := make([]string, 0, redisInsertBatchSize)
	batchData := make([]TokenData, 0, redisInsertBatchSize)

	for tokenJSON, tokenData := range tokenSet {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		batchJSON = append(batchJSON, tokenJSON)
		batchData = append(batchData, tokenData)

		if len(batchJSON) == redisInsertBatchSize {
			insertedCount = ts.insertBatch(ctx, batchJSON, batchData, insertedCount)
			batchJSON = batchJSON[:0]
			batchData = batchData[:0]
		}
	}

	if len(batchJSON) > 0 {
		insertedCount = ts.insertBatch(ctx, batchJSON, batchData, insertedCount)
	}

	ts.log.Info().
		Int("inserted_tokens", insertedCount).
		Int("total_processed", len(tokenSet)).
		Msg("Token sync completed")

	return nil
}

func (ts *TokenStore) insertBatch(ctx context.Context, batchJSON []string, batchData []TokenData, insertedSoFar int) int {
	if len(batchJSON) == 0 {
		return insertedSoFar
	}

	members := make([]interface{}, 0, len(batchJSON))
	for _, tokenJSON := range batchJSON {
		members = append(members, tokenJSON)
	}

	if err := ts.redisClient.SAdd(ctx, ts.queueName, members...).Err(); err == nil {
		return ts.logProgress(insertedSoFar, len(batchJSON))
	}

	ts.log.Warn().
		Int("batch_size", len(batchJSON)).
		Str("queue_name", ts.queueName).
		Msg("Batch insert failed, falling back to per-token Redis writes")

	insertedCount := insertedSoFar
	for i, tokenJSON := range batchJSON {
		if err := ts.redisClient.SAdd(ctx, ts.queueName, tokenJSON).Err(); err != nil {
			ts.log.Warn().
				Err(err).
				Str("platform_id", batchData[i].PlatformID).
				Msg("Error adding valid token to Redis")
			continue
		}
		insertedCount++
	}

	return ts.logProgress(insertedSoFar, insertedCount-insertedSoFar)
}

func (ts *TokenStore) logProgress(insertedSoFar, batchInserted int) int {
	insertedCount := insertedSoFar + batchInserted
	for progress := ((insertedSoFar / redisInsertBatchSize) + 1) * redisInsertBatchSize; progress <= insertedCount; progress += redisInsertBatchSize {
		ts.log.Info().
			Int("inserted_tokens", progress).
			Str("platform", ts.platform).
			Msg("Token sync progress")
	}
	return insertedCount
}

// ProcessJob executes the token sync job in sequence.
// 1. Load valid accounts from MongoDB
// 2. Refresh Redis with the MongoDB-derived token set
func (ts *TokenStore) ProcessJob(ctx context.Context) error {
	startTime := time.Now()

	ts.log.Info().
		Str("platform", ts.platform).
		Msg("========== Starting token sync job ==========")

	// Step 1: Populate token set
	tokenSet, err := ts.PopulateRedisSet(ctx)
	if err != nil {
		elapsed := time.Since(startTime)
		ts.log.Warn().
			Err(err).
			Str("platform", ts.platform).
			Dur("elapsed", elapsed).
			Msg("Token sync job FAILED during populate")
		return fmt.Errorf("TokenStore.ProcessJob: token sync failed during populate: %w", err)
	}

	// Step 2: Sync tokens to Redis
	if err := ts.Validate(ctx, tokenSet); err != nil {
		elapsed := time.Since(startTime)
		ts.log.Warn().
			Err(err).
			Str("platform", ts.platform).
			Dur("elapsed", elapsed).
			Msg("Token sync job FAILED during validation")
		return fmt.Errorf("TokenStore.ProcessJob: token sync failed during validation: %w", err)
	}

	elapsed := time.Since(startTime)
	ts.log.Info().
		Str("platform", ts.platform).
		Dur("elapsed", elapsed).
		Msg("========== Token sync job completed successfully ==========")

	return nil
}
