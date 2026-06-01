package mongodb

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	mongo3 "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// UnifiedSocialRepository defines the interface for unified social account operations
type UnifiedSocialRepository interface {
	// Core methods
	FindByID(ctx context.Context, id primitive.ObjectID) (*mongo3.SocialIntegration, error)
	GetByPlatformID(ctx context.Context, platformType, platformID string) (*mongo3.SocialIntegration, error)
	GetValidAccounts(ctx context.Context, platformType string, accountTypes []string) ([]mongo3.SocialIntegration, error)
	GetAccountsByWorkspace(ctx context.Context, workspaceID primitive.ObjectID, platforms []string) ([]mongo3.SocialIntegration, error)
	GetAccountsNeedingUpdate(ctx context.Context, platformType string, lastUpdateField string, hours int) ([]mongo3.SocialIntegration, error)

	// Paginated fetch with filtering - for batch processing
	GetAccountsNeedingUpdatePaginated(ctx context.Context, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongo3.SocialIntegration, error)
	CountAccountsNeedingUpdate(ctx context.Context, platformType string, accountTypes []string, hours int) (int64, error)

	// ID-based pagination (cursor-free, avoids timeout issues)
	GetAccountsNeedingUpdateByID(ctx context.Context, platformType string, accountTypes []string, hours int, lastID primitive.ObjectID, limit int64) ([]mongo3.SocialIntegration, error)

	// Valid account pagination — same filter as GetValidAccounts but cursor-based
	GetValidAccountsByID(ctx context.Context, platformType string, accountTypes []string, lastID primitive.ObjectID, limit int64) ([]mongo3.SocialIntegration, error)
	CountValidAccounts(ctx context.Context, platformType string, accountTypes []string) (int64, error)
	GetAccountsByPlatformIDs(ctx context.Context, platformType string, platformIDs []string) ([]mongo3.SocialIntegration, error)

	// YouTube-specific methods (with consent time filter)
	GetYouTubeAccountsNeedingUpdatePaginated(ctx context.Context, hours int, consentDays int, skip, limit int64) ([]mongo3.SocialIntegration, error)
	GetYouTubeAccountsNeedingUpdateByID(ctx context.Context, hours int, consentDays int, lastID primitive.ObjectID, limit int64) ([]mongo3.SocialIntegration, error)
	CountYouTubeAccountsNeedingUpdate(ctx context.Context, hours int, consentDays int) (int64, error)

	// Update methods
	Update(ctx context.Context, id primitive.ObjectID, updates primitive.M) error
	UpdateAnalyticsTimestamp(ctx context.Context, id primitive.ObjectID, timestampType string, timestamp time.Time) error
	UpdateTokens(ctx context.Context, id primitive.ObjectID, tokens map[string]string) error
	UpdateState(ctx context.Context, id primitive.ObjectID, newState string) error
	UpdateValidity(ctx context.Context, id primitive.ObjectID, newValidity string) error

	// Processing error tracking
	RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error
	ClearProcessingError(ctx context.Context, id primitive.ObjectID) error

	// Create and Delete
	Create(ctx context.Context, account *mongo3.SocialIntegration) (primitive.ObjectID, error)
	Delete(ctx context.Context, id primitive.ObjectID) error

	// Twitter job metadata
	InsertTwitterJobMetadata(ctx context.Context, payload TwitterJobMetadataPayload) error
}

// unifiedSocialRepository implements the UnifiedSocialRepository interface
type unifiedSocialRepository struct {
	collection *mongo.Collection
	logger     zerolog.Logger
}

// TwitterJobMetadataPayload defines fields inserted into twitter_jobs_metadata.
type TwitterJobMetadataPayload struct {
	PlatformID  string
	WorkspaceID string
	CreditsUsed int
	ExecutedBy  string
	AppID       string
	AppName     string
}

// NewUnifiedSocialRepository creates a new UnifiedSocialRepository
func NewUnifiedSocialRepository(db *mongo.Database, logger zerolog.Logger) UnifiedSocialRepository {
	return &unifiedSocialRepository{
		collection: db.Collection("social_integrations"),
		logger:     logger.With().Str("repository", "UnifiedSocial").Logger(),
	}
}

// FindByID retrieves a social account by its MongoDB _id
func (r *unifiedSocialRepository) FindByID(ctx context.Context, id primitive.ObjectID) (*mongo3.SocialIntegration, error) {
	var account mongo3.DBSocialIntegration
	err := r.collection.FindOne(ctx, bson.M{"_id": id}).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			r.logger.Warn().Str("id", id.Hex()).Msg("Social account not found by ID")
			return nil, nil
		}
		r.logger.Warn().Err(err).Str("id", id.Hex()).Msg("Error finding social account by ID")
		return nil, err
	}
	parseAccount := mongo3.ConvertDBToSocialIntegration(account)
	return &parseAccount, nil
}

// GetByPlatformID retrieves a social account by platform type and platform-specific ID
func (r *unifiedSocialRepository) GetByPlatformID(ctx context.Context, platformType, platformID string) (*mongo3.SocialIntegration, error) {
	filter := bson.M{
		"platform_type":       platformType,
		"platform_identifier": platformID,
	}

	var account mongo3.DBSocialIntegration
	err := r.collection.FindOne(ctx, filter).Decode(&account)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			// Try legacy field as fallback for backward compatibility
			legacyFilter := r.buildLegacyFilter(platformType, platformID)
			if legacyFilter != nil {
				err = r.collection.FindOne(ctx, legacyFilter).Decode(&account)
				if err == nil {
					parseAccount := mongo3.ConvertDBToSocialIntegration(account)
					return &parseAccount, nil
				}
			}

			r.logger.Warn().
				Str("platform_type", platformType).
				Str("platform_id", platformID).
				Msg("Social account not found by platform ID")
			return nil, nil
		}
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Str("platform_id", platformID).
			Msg("Error finding social account by platform ID")
		return nil, err
	}

	parseAccount := mongo3.ConvertDBToSocialIntegration(account)
	return &parseAccount, nil
}

// GetValidAccounts retrieves all valid accounts for a specific platform and types
func (r *unifiedSocialRepository) GetValidAccounts(ctx context.Context, platformType string, accountTypes []string) ([]mongo3.SocialIntegration, error) {
	filter := bson.M{
		"platform_type": platformType,
		"validity":      mongo3.ValidityValid,
		"state":         bson.M{"$in": []string{mongo3.StateAdded, mongo3.StateSyncing, mongo3.StateProcessed, mongo3.StateFailed}},
		"super_admin_state": bson.M{"$in": []string{
			mongo3.SuperAdminStateActive,
			mongo3.SuperAdminStatePastDue,
		}},
	}

	// Add account type filter if specified
	if len(accountTypes) == 1 {
		filter["type"] = accountTypes[0]
	} else if len(accountTypes) > 1 {
		filter["type"] = bson.M{"$in": accountTypes}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Strs("account_types", accountTypes).
			Msg("Error finding valid social accounts")
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []mongo3.DBSocialIntegration
	if err = cursor.All(ctx, &accounts); err != nil {
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Strs("account_types", accountTypes).
			Msg("Error decoding valid social accounts")
		return nil, err
	}

	r.logger.Info().
		Int("count", len(accounts)).
		Str("platform_type", platformType).
		Strs("account_types", accountTypes).
		Msg("Retrieved valid social accounts")

	parsAccounts := make([]mongo3.SocialIntegration, 0, len(accounts))
	for _, account := range accounts {
		parsAccounts = append(parsAccounts, mongo3.ConvertDBToSocialIntegration(account))
	}

	return parsAccounts, nil
}

func (r *unifiedSocialRepository) buildValidAccountsFilter(platformType string, accountTypes []string) bson.M {
	filter := bson.M{
		"platform_type": platformType,
		"validity":      mongo3.ValidityValid,
		"state":         bson.M{"$in": []string{mongo3.StateAdded, mongo3.StateSyncing, mongo3.StateProcessed, mongo3.StateFailed}},
		"super_admin_state": bson.M{"$in": []string{
			mongo3.SuperAdminStateActive,
			mongo3.SuperAdminStatePastDue,
		}},
	}
	if len(accountTypes) == 1 {
		filter["type"] = accountTypes[0]
	} else if len(accountTypes) > 1 {
		filter["type"] = bson.M{"$in": accountTypes}
	}
	return filter
}

func (r *unifiedSocialRepository) GetValidAccountsByID(ctx context.Context, platformType string, accountTypes []string, lastID primitive.ObjectID, limit int64) ([]mongo3.SocialIntegration, error) {
	filter := r.buildValidAccountsFilter(platformType, accountTypes)
	if lastID != primitive.NilObjectID {
		filter["_id"] = bson.M{"$gt": lastID}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}).
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []mongo3.DBSocialIntegration
	if err = cursor.All(ctx, &accounts); err != nil {
		return nil, err
	}

	out := make([]mongo3.SocialIntegration, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, mongo3.ConvertDBToSocialIntegration(a))
	}
	return out, nil
}

func (r *unifiedSocialRepository) CountValidAccounts(ctx context.Context, platformType string, accountTypes []string) (int64, error) {
	filter := r.buildValidAccountsFilter(platformType, accountTypes)
	return r.collection.CountDocuments(ctx, filter)
}

func (r *unifiedSocialRepository) GetAccountsByPlatformIDs(ctx context.Context, platformType string, platformIDs []string) ([]mongo3.SocialIntegration, error) {
	if len(platformIDs) == 0 {
		return nil, nil
	}
	filter := r.buildValidAccountsFilter(platformType, nil)
	filter["platform_identifier"] = bson.M{"$in": platformIDs}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []mongo3.DBSocialIntegration
	if err = cursor.All(ctx, &accounts); err != nil {
		return nil, err
	}

	out := make([]mongo3.SocialIntegration, 0, len(accounts))
	for _, a := range accounts {
		out = append(out, mongo3.ConvertDBToSocialIntegration(a))
	}
	return out, nil
}

// GetAccountsByWorkspace retrieves all accounts for a workspace, optionally filtered by platforms
func (r *unifiedSocialRepository) GetAccountsByWorkspace(ctx context.Context, workspaceID primitive.ObjectID, platforms []string) ([]mongo3.SocialIntegration, error) {
	filter := bson.M{
		"workspace_id": workspaceID,
		"state":        bson.M{"$ne": mongo3.StateDeleted},
	}

	// Add platform filter if specified
	if len(platforms) > 0 {
		filter["platform_type"] = bson.M{"$in": platforms}
	}

	cursor, err := r.collection.Find(ctx, filter)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Str("workspace_id", workspaceID.Hex()).
			Interface("platforms", platforms).
			Msg("Error finding workspace social accounts")
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []mongo3.DBSocialIntegration
	if err = cursor.All(ctx, &accounts); err != nil {
		r.logger.Warn().
			Err(err).
			Str("workspace_id", workspaceID.Hex()).
			Msg("Error decoding workspace social accounts")
		return nil, err
	}

	parsAccounts := make([]mongo3.SocialIntegration, 0, len(accounts))
	for _, account := range accounts {
		parsAccounts = append(parsAccounts, mongo3.ConvertDBToSocialIntegration(account))
	}

	return parsAccounts, nil
}

// GetAccountsNeedingUpdate retrieves accounts that haven't been updated recently
func (r *unifiedSocialRepository) GetAccountsNeedingUpdate(ctx context.Context, platformType string, lastUpdateField string, hours int) ([]mongo3.SocialIntegration, error) {
	cutoffTime := time.Now().UTC().Add(-time.Duration(hours) * time.Hour)

	filter := bson.M{
		"platform_type": platformType,
		"validity":      mongo3.ValidityValid,
		"state":         mongo3.StateAdded,
		"$or": []bson.M{
			{lastUpdateField: bson.M{"$lt": cutoffTime}},
			{lastUpdateField: nil},
		},
	}

	// Add sorting to process oldest first
	opts := options.Find().SetSort(bson.D{{Key: lastUpdateField, Value: 1}})

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Str("update_field", lastUpdateField).
			Int("hours", hours).
			Msg("Error finding accounts needing update")
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []mongo3.DBSocialIntegration
	if err = cursor.All(ctx, &accounts); err != nil {
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Msg("Error decoding accounts needing update")
		return nil, err
	}

	r.logger.Info().
		Int("count", len(accounts)).
		Str("platform_type", platformType).
		Str("update_field", lastUpdateField).
		Int("hours", hours).
		Msg("Found accounts needing update")

	parsAccounts := make([]mongo3.SocialIntegration, 0, len(accounts))
	for _, account := range accounts {
		parsAccounts = append(parsAccounts, mongo3.ConvertDBToSocialIntegration(account))
	}

	r.clearRetryableProcessingErrors(ctx, parsAccounts)

	return parsAccounts, nil
}

// buildNeedingUpdateFilter creates the filter for accounts needing analytics update
func (r *unifiedSocialRepository) buildNeedingUpdateFilter(platformType string, accountTypes []string, hours int) bson.M {
	filter := bson.M{
		"platform_type": platformType,
		"validity":      mongo3.ValidityValid,
		"state":         bson.M{"$in": []string{mongo3.StateAdded, mongo3.StateSyncing, mongo3.StateProcessed, mongo3.StateFailed}},
		"super_admin_state": bson.M{"$in": []string{
			mongo3.SuperAdminStateActive,
			mongo3.SuperAdminStatePastDue,
		}},
	}

	if len(accountTypes) == 1 {
		filter["type"] = accountTypes[0]
	} else if len(accountTypes) > 1 {
		filter["type"] = bson.M{"$in": accountTypes}
	}

	return filter
}

// GetAccountsNeedingUpdatePaginated retrieves accounts needing update with pagination.
// Used by the scheduler to fetch accounts in batches for batch work order production.
// Filters by last_analytics_updated_at older than `hours` or null.
// Includes retry logic for transient cursor errors.
func (r *unifiedSocialRepository) GetAccountsNeedingUpdatePaginated(ctx context.Context, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongo3.SocialIntegration, error) {
	filter := r.buildNeedingUpdateFilter(platformType, accountTypes, hours)

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetNoCursorTimeout(true)

	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		accounts, err := r.executeAccountsQuery(ctx, filter, opts, platformType, accountTypes, hours, skip, limit)
		if err == nil {
			return accounts, nil
		}

		lastErr = err

		if !isCursorError(err) {
			return nil, err
		}

		r.logger.Warn().
			Err(err).
			Int("attempt", attempt).
			Int("max_retries", maxRetries).
			Str("platform_type", platformType).
			Int64("skip", skip).
			Msg("Cursor error, retrying...")

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
	}

	r.logger.Warn().
		Err(lastErr).
		Str("platform_type", platformType).
		Int64("skip", skip).
		Msg("All retry attempts failed for paginated query")
	return nil, lastErr
}

func (r *unifiedSocialRepository) executeAccountsQuery(ctx context.Context, filter bson.M, opts *options.FindOptions, platformType string, accountTypes []string, hours int, skip, limit int64) ([]mongo3.SocialIntegration, error) {
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Strs("account_types", accountTypes).
			Int("hours", hours).
			Int64("skip", skip).
			Int64("limit", limit).
			Msg("Error finding accounts needing update (paginated)")
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []mongo3.DBSocialIntegration
	if err = cursor.All(ctx, &accounts); err != nil {
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Msg("Error decoding accounts needing update (paginated)")
		return nil, err
	}

	r.logger.Debug().
		Int("count", len(accounts)).
		Str("platform_type", platformType).
		Strs("account_types", accountTypes).
		Int64("skip", skip).
		Int64("limit", limit).
		Msg("Fetched accounts needing update (paginated)")

	parsAccounts := make([]mongo3.SocialIntegration, 0, len(accounts))
	for _, account := range accounts {
		parsAccounts = append(parsAccounts, mongo3.ConvertDBToSocialIntegration(account))
	}

	r.clearRetryableProcessingErrors(ctx, parsAccounts)

	return parsAccounts, nil
}

func isCursorError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "CursorNotFound") ||
		strings.Contains(errStr, "cursor") ||
		strings.Contains(errStr, "Cursor")
}

// CountAccountsNeedingUpdate returns the total count of accounts needing update.
// Used by the scheduler to determine pagination bounds.
func (r *unifiedSocialRepository) CountAccountsNeedingUpdate(ctx context.Context, platformType string, accountTypes []string, hours int) (int64, error) {
	filter := r.buildNeedingUpdateFilter(platformType, accountTypes, hours)

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Strs("account_types", accountTypes).
			Int("hours", hours).
			Msg("Error counting accounts needing update")
		return 0, err
	}

	r.logger.Info().
		Int64("count", count).
		Str("platform_type", platformType).
		Strs("account_types", accountTypes).
		Int("hours", hours).
		Msg("Counted accounts needing update")

	return count, nil
}

// GetAccountsNeedingUpdateByID retrieves accounts needing update using ID-based pagination.
// This avoids cursor timeout issues by using _id > lastID filter instead of skip/limit.
// Pass primitive.NilObjectID as lastID for the first batch.
// Use small limit values (e.g., 50) to ensure fast query completion.
func (r *unifiedSocialRepository) GetAccountsNeedingUpdateByID(ctx context.Context, platformType string, accountTypes []string, hours int, lastID primitive.ObjectID, limit int64) ([]mongo3.SocialIntegration, error) {
	filter := r.buildNeedingUpdateFilter(platformType, accountTypes, hours)

	// Add _id filter for pagination (skip documents we've already seen)
	if lastID != primitive.NilObjectID {
		filter["_id"] = bson.M{"$gt": lastID}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}). // Sort by _id ascending for consistent pagination
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Strs("account_types", accountTypes).
			Str("last_id", lastID.Hex()).
			Int64("limit", limit).
			Msg("Error finding accounts needing update (ID-based)")
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []mongo3.DBSocialIntegration
	if err = cursor.All(ctx, &accounts); err != nil {
		r.logger.Warn().
			Err(err).
			Str("platform_type", platformType).
			Str("last_id", lastID.Hex()).
			Msg("Error decoding accounts needing update (ID-based)")
		return nil, err
	}

	r.logger.Debug().
		Int("count", len(accounts)).
		Str("platform_type", platformType).
		Strs("account_types", accountTypes).
		Str("last_id", lastID.Hex()).
		Int64("limit", limit).
		Msg("Fetched accounts needing update (ID-based)")

	parsAccounts := make([]mongo3.SocialIntegration, 0, len(accounts))
	for _, account := range accounts {
		parsAccounts = append(parsAccounts, mongo3.ConvertDBToSocialIntegration(account))
	}

	r.clearRetryableProcessingErrors(ctx, parsAccounts)

	return parsAccounts, nil
}

func hasProcessingErrorMeta(meta interface{}) bool {
	switch v := meta.(type) {
	case nil:
		return false
	case bson.M:
		return hasProcessingErrorMetaValue(v["last_processing_error"])
	case map[string]interface{}:
		return hasProcessingErrorMetaValue(v["last_processing_error"])
	case map[string]string:
		return strings.TrimSpace(v["last_processing_error"]) != ""
	case primitive.D:
		return hasProcessingErrorMeta(v.Map())
	default:
		return false
	}
}

// HasProcessingErrorMeta reports whether the account metadata contains a retryable processing error.
func HasProcessingErrorMeta(meta interface{}) bool {
	return hasProcessingErrorMeta(meta)
}

func hasProcessingErrorMetaValue(value interface{}) bool {
	switch v := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(v) != ""
	default:
		return true
	}
}

func processingErrorAccountIDs(accounts []mongo3.SocialIntegration) []primitive.ObjectID {
	ids := make([]primitive.ObjectID, 0, len(accounts))
	for _, account := range accounts {
		if account.ID.IsZero() {
			continue
		}
		if hasProcessingErrorMeta(account.MetaData) {
			ids = append(ids, account.ID)
		}
	}
	return ids
}

func (r *unifiedSocialRepository) clearRetryableProcessingErrors(ctx context.Context, accounts []mongo3.SocialIntegration) {
	ids := processingErrorAccountIDs(accounts)
	if len(ids) == 0 {
		return
	}

	now := time.Now().UTC()
	update := bson.M{
		"$unset": bson.M{
			"meta_data.last_processing_error":    "",
			"meta_data.last_processing_error_at": "",
			"meta_data.consecutive_failures":     "",
		},
		"$set": bson.M{
			"updated_at": mongo3.MongoTime{Time: now},
		},
	}

	result, err := r.collection.UpdateMany(ctx, bson.M{"_id": bson.M{"$in": ids}}, update)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Int("accounts", len(ids)).
			Msg("Failed to clear retryable processing errors before scheduling")
		return
	}

	r.logger.Debug().
		Int("accounts", len(ids)).
		Int64("matched", result.MatchedCount).
		Int64("modified", result.ModifiedCount).
		Msg("Cleared retryable processing errors before scheduling")
}

// buildYouTubeFilter creates the filter for YouTube accounts with consent time validation.
// Excludes accounts where preferences.last_youtube_consent_time is older than consentDays.
// Note: last_youtube_consent_time is stored as ISO8601 string in DB, so we compare as string.
func (r *unifiedSocialRepository) buildYouTubeFilter(consentDays int) bson.M {
	consentCutoff := time.Now().UTC().AddDate(0, 0, -consentDays)
	consentCutoffStr := consentCutoff.Format(time.RFC3339)

	filter := bson.M{
		"platform_type": mongo3.PlatformYouTube,
		"validity":      mongo3.ValidityValid,
		"state":         bson.M{"$in": []string{mongo3.StateAdded, mongo3.StateSyncing, mongo3.StateProcessed, mongo3.StateFailed}},
		"super_admin_state": bson.M{"$in": []string{
			mongo3.SuperAdminStateActive,
			mongo3.SuperAdminStatePastDue,
		}},
		"preferences.last_youtube_consent_time": bson.M{"$gte": consentCutoffStr},
	}

	return filter
}

// GetYouTubeAccountsNeedingUpdatePaginated retrieves YouTube accounts needing update with pagination.
// Includes YouTube-specific consent time filter to exclude accounts with expired consent (> consentDays).
// Includes retry logic for transient cursor errors.
func (r *unifiedSocialRepository) GetYouTubeAccountsNeedingUpdatePaginated(ctx context.Context, hours int, consentDays int, skip, limit int64) ([]mongo3.SocialIntegration, error) {
	filter := r.buildYouTubeFilter(consentDays)

	opts := options.Find().
		SetSkip(skip).
		SetLimit(limit).
		SetNoCursorTimeout(true)

	const maxRetries = 3
	var lastErr error

	for attempt := 1; attempt <= maxRetries; attempt++ {
		accounts, err := r.executeYouTubeAccountsQuery(ctx, filter, opts, hours, consentDays, skip, limit)
		if err == nil {
			return accounts, nil
		}

		lastErr = err

		if !isCursorError(err) {
			return nil, err
		}

		r.logger.Warn().
			Err(err).
			Int("attempt", attempt).
			Int("max_retries", maxRetries).
			Str("platform", "youtube").
			Int64("skip", skip).
			Msg("Cursor error, retrying...")

		if attempt < maxRetries {
			time.Sleep(time.Duration(attempt) * 500 * time.Millisecond)
		}
	}

	r.logger.Warn().
		Err(lastErr).
		Str("platform", "youtube").
		Int64("skip", skip).
		Msg("All retry attempts failed for YouTube paginated query")
	return nil, lastErr
}

func (r *unifiedSocialRepository) executeYouTubeAccountsQuery(ctx context.Context, filter bson.M, opts *options.FindOptions, hours int, consentDays int, skip, limit int64) ([]mongo3.SocialIntegration, error) {
	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Int("hours", hours).
			Int("consent_days", consentDays).
			Int64("skip", skip).
			Int64("limit", limit).
			Msg("Error finding YouTube accounts needing update (paginated)")
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []mongo3.DBSocialIntegration
	if err = cursor.All(ctx, &accounts); err != nil {
		r.logger.Warn().
			Err(err).
			Msg("Error decoding YouTube accounts needing update (paginated)")
		return nil, err
	}

	r.logger.Debug().
		Int("count", len(accounts)).
		Int("consent_days", consentDays).
		Int64("skip", skip).
		Int64("limit", limit).
		Msg("Fetched YouTube accounts needing update (paginated)")

	parsAccounts := make([]mongo3.SocialIntegration, 0, len(accounts))
	for _, account := range accounts {
		parsAccounts = append(parsAccounts, mongo3.ConvertDBToSocialIntegration(account))
	}

	r.clearRetryableProcessingErrors(ctx, parsAccounts)

	return parsAccounts, nil
}

// GetYouTubeAccountsNeedingUpdateByID retrieves YouTube accounts needing update using ID-based pagination.
// This avoids cursor timeout issues by using _id > lastID filter instead of skip/limit.
// Pass primitive.NilObjectID as lastID for the first batch.
// Use small limit values (e.g., 50) to ensure fast query completion.
func (r *unifiedSocialRepository) GetYouTubeAccountsNeedingUpdateByID(ctx context.Context, hours int, consentDays int, lastID primitive.ObjectID, limit int64) ([]mongo3.SocialIntegration, error) {
	filter := r.buildYouTubeFilter(consentDays)

	// Add _id filter for pagination (skip documents we've already seen)
	if lastID != primitive.NilObjectID {
		filter["_id"] = bson.M{"$gt": lastID}
	}

	opts := options.Find().
		SetSort(bson.D{{Key: "_id", Value: 1}}). // Sort by _id ascending for consistent pagination
		SetLimit(limit)

	cursor, err := r.collection.Find(ctx, filter, opts)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Int("hours", hours).
			Int("consent_days", consentDays).
			Str("last_id", lastID.Hex()).
			Int64("limit", limit).
			Msg("Error finding YouTube accounts needing update (ID-based)")
		return nil, err
	}
	defer cursor.Close(ctx)

	var accounts []mongo3.DBSocialIntegration
	if err = cursor.All(ctx, &accounts); err != nil {
		r.logger.Warn().
			Err(err).
			Str("last_id", lastID.Hex()).
			Msg("Error decoding YouTube accounts needing update (ID-based)")
		return nil, err
	}

	r.logger.Debug().
		Int("count", len(accounts)).
		Int("consent_days", consentDays).
		Str("last_id", lastID.Hex()).
		Int64("limit", limit).
		Msg("Fetched YouTube accounts needing update (ID-based)")

	parsAccounts := make([]mongo3.SocialIntegration, 0, len(accounts))
	for _, account := range accounts {
		parsAccounts = append(parsAccounts, mongo3.ConvertDBToSocialIntegration(account))
	}

	r.clearRetryableProcessingErrors(ctx, parsAccounts)

	return parsAccounts, nil
}

// CountYouTubeAccountsNeedingUpdate returns the total count of YouTube accounts needing update.
// Includes YouTube-specific consent time filter to exclude accounts with expired consent (> consentDays).
func (r *unifiedSocialRepository) CountYouTubeAccountsNeedingUpdate(ctx context.Context, hours int, consentDays int) (int64, error) {
	filter := r.buildYouTubeFilter(consentDays)

	count, err := r.collection.CountDocuments(ctx, filter)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Int("hours", hours).
			Int("consent_days", consentDays).
			Msg("Error counting YouTube accounts needing update")
		return 0, err
	}

	r.logger.Info().
		Int64("count", count).
		Int("consent_days", consentDays).
		Msg("Counted YouTube accounts needing update")

	return count, nil
}

// Update modifies specific fields of a social account
func (r *unifiedSocialRepository) Update(ctx context.Context, id primitive.ObjectID, updates primitive.M) error {
	if len(updates) == 0 {
		r.logger.Warn().Str("id", id.Hex()).Msg("Update called with no fields to update")
		return nil
	}

	// Add updated_at timestamp
	updates["updated_at"] = mongo3.MongoTime{Time: time.Now().UTC()}

	updateDoc := bson.M{"$set": updates}

	result, err := r.collection.UpdateByID(ctx, id, updateDoc)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Str("id", id.Hex()).
			Interface("updates", updates).
			Msg("Error updating social account")
		return err
	}

	if result.MatchedCount == 0 {
		r.logger.Warn().Str("id", id.Hex()).Msg("No social account found with the given ID")
		return mongo.ErrNoDocuments
	}

	r.logger.Info().
		Str("id", id.Hex()).
		Int64("modified", result.ModifiedCount).
		Msg("Social account updated")

	return nil
}

// UpdateAnalyticsTimestamp updates a specific analytics timestamp field
func (r *unifiedSocialRepository) UpdateAnalyticsTimestamp(ctx context.Context, id primitive.ObjectID, timestampType string, timestamp time.Time) error {
	fieldMap := map[string]string{
		"analytics":    "last_analytics_updated_at",
		"insights":     "last_insights_analytics_updated_at",
		"fans":         "last_fans_analytics_updated_at",
		"video":        "last_video_analytics_updated_at",
		"group":        "last_group_analytics_updated_at",
		"link_preview": "last_link_preview_updated_at",
	}

	field, exists := fieldMap[timestampType]
	if !exists {
		return errors.New("invalid timestamp type")
	}

	updates := primitive.M{
		field: mongo3.MongoTime{Time: timestamp},
	}

	r.logger.Info().
		Str("id", id.Hex()).
		Str("timestamp_type", timestampType).
		Time("timestamp", timestamp).
		Msg("Updating analytics timestamp")

	return r.Update(ctx, id, updates)
}

// UpdateTokens updates token fields for a social account
func (r *unifiedSocialRepository) UpdateTokens(ctx context.Context, id primitive.ObjectID, tokens map[string]string) error {
	updates := primitive.M{}

	for key, value := range tokens {
		switch key {
		case "access_token", "refresh_token", "long_access_token",
			"oauth_token", "oauth_token_secret":
			updates[key] = value
		case "expires_at":
			// Parse and set token_expires_at
			// This would need proper time parsing
			continue
		default:
			r.logger.Warn().
				Str("key", key).
				Msg("Unknown token field, skipping")
		}
	}

	if len(updates) == 0 {
		return errors.New("no valid token fields to update")
	}

	r.logger.Info().
		Str("id", id.Hex()).
		Int("token_count", len(updates)).
		Msg("Updating tokens")

	return r.Update(ctx, id, updates)
}

// UpdateState updates the state field of a social account
func (r *unifiedSocialRepository) UpdateState(ctx context.Context, id primitive.ObjectID, newState string) error {
	updates := primitive.M{"state": newState}

	r.logger.Info().
		Str("id", id.Hex()).
		Str("new_state", newState).
		Msg("Updating social account state")

	return r.Update(ctx, id, updates)
}

// UpdateValidity updates the validity field of a social account
func (r *unifiedSocialRepository) UpdateValidity(ctx context.Context, id primitive.ObjectID, newValidity string) error {
	updates := primitive.M{"validity": newValidity}

	r.logger.Info().
		Str("id", id.Hex()).
		Str("new_validity", newValidity).
		Msg("Updating social account validity")

	return r.Update(ctx, id, updates)
}

// cleanErrorMessage extracts a human-readable error message from verbose internal errors.
func cleanErrorMessage(raw string) string {
	if raw == "" {
		return raw
	}

	// Handle multi-line errors (e.g., LinkedIn: "token expired\nLinkedInClient.Fetch...: status 401 body {json}")
	if strings.Contains(raw, "\n") {
		for _, line := range strings.Split(raw, "\n") {
			if msg := extractJSONMessage(line); msg != "" {
				return capLength(simplifyErrorMessage(msg), 300)
			}
		}
		raw = strings.TrimSpace(strings.SplitN(raw, "\n", 2)[0])
	}

	// Try JSON message extraction (handles LinkedIn, Pinterest JSON error bodies)
	if msg := extractJSONMessage(raw); msg != "" {
		return capLength(simplifyErrorMessage(msg), 300)
	}

	// Strip "XxxClient.MethodName: " prefix chains
	raw = stripClientPrefixes(raw)

	// Strip platform API error wrappers to expose the actual error message
	raw = stripAPIErrorPrefix(raw)

	// After prefix stripping, JSON body may be exposed
	if msg := extractJSONMessage(raw); msg != "" {
		return capLength(simplifyErrorMessage(msg), 300)
	}

	// Strip (OAuthException/NNN) and (Type: ..., Code: NNN) suffixes
	raw = stripAuthSuffixes(raw)

	// TikTok: "error_code - actual message" → just the message
	raw = stripErrorCodePrefix(raw)

	// Simplify common verbose API errors into short user-friendly messages
	raw = simplifyErrorMessage(raw)

	return capLength(strings.TrimSpace(raw), 300)
}

func extractJSONMessage(s string) string {
	start := strings.Index(s, "{")
	if start < 0 {
		return ""
	}
	end := strings.LastIndex(s, "}")
	if end <= start {
		return ""
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(s[start:end+1]), &obj); err != nil {
		return ""
	}

	if msg, ok := obj["message"].(string); ok && msg != "" {
		return msg
	}
	if msg, ok := obj["error_message"].(string); ok && msg != "" {
		return msg
	}
	if errObj, ok := obj["error"].(map[string]interface{}); ok {
		if msg, ok := errObj["message"].(string); ok && msg != "" {
			return msg
		}
	}
	return ""
}

func stripClientPrefixes(s string) string {
	for {
		idx := strings.Index(s, "Client.")
		if idx < 0 {
			return s
		}
		colonIdx := strings.Index(s[idx:], ": ")
		if colonIdx < 0 {
			return s
		}
		s = s[idx+colonIdx+2:]
	}
}

func stripAPIErrorPrefix(s string) string {
	lower := strings.ToLower(s)

	// Match patterns like:
	// "instagram API error: msg"
	// "linkedin posts error: status 401 body {json}"
	// "tiktok api error (status 200): code - msg"
	// "twitter api unauthorized (401): msg"
	// "pinterest API unauthorized (status 401): {json}"
	// "API error (status 401): msg"  (GMB)
	for _, marker := range []string{
		" api error: ",
		" api-server error: ",
		" error: ",
	} {
		if idx := strings.Index(lower, marker); idx >= 0 {
			rest := s[idx+len(marker):]
			restLower := strings.ToLower(rest)
			// Handle "status NNN body {json}" (LinkedIn pattern)
			if strings.HasPrefix(restLower, "status ") {
				if bodyIdx := strings.Index(restLower, " body "); bodyIdx >= 0 {
					return rest[bodyIdx+6:]
				}
			}
			// Handle "(status NNN): msg" that follows "api error "
			if strings.HasPrefix(rest, "(") {
				if closeIdx := strings.Index(rest, "): "); closeIdx >= 0 {
					return rest[closeIdx+3:]
				}
			}
			return rest
		}
	}

	// Handle "(status NNN): msg" or "(NNN): msg" without prior "error" keyword
	// e.g., "twitter api unauthorized (401): msg", "pinterest API unauthorized (status 401): body"
	if parenIdx := strings.Index(s, "("); parenIdx >= 0 {
		if closeIdx := strings.Index(s[parenIdx:], "): "); closeIdx >= 0 {
			return s[parenIdx+closeIdx+3:]
		}
	}

	// Handle "http NNN: msg"
	if strings.HasPrefix(lower, "http ") {
		if idx := strings.Index(s, ": "); idx >= 0 && idx < 12 {
			return s[idx+2:]
		}
	}

	return s
}

func stripAuthSuffixes(s string) string {
	if idx := strings.LastIndex(s, " (OAuthException"); idx > 0 {
		s = s[:idx]
	}
	if idx := strings.LastIndex(s, " (Type: "); idx > 0 {
		s = s[:idx]
	}
	return s
}

func stripErrorCodePrefix(s string) string {
	parts := strings.SplitN(s, " - ", 2)
	if len(parts) == 2 && !strings.Contains(parts[0], " ") && len(parts[0]) > 0 {
		return parts[1]
	}
	return s
}

func simplifyErrorMessage(s string) string {
	lower := strings.ToLower(s)

	// Token/session expiry patterns
	if strings.Contains(lower, "error validating access token") ||
		strings.Contains(lower, "invalid or expired token") ||
		strings.Contains(lower, "the token used in the request has expired") ||
		strings.Contains(lower, "token has been expired or revoked") {
		reason := ""
		switch {
		case strings.Contains(lower, "password"):
			reason = "user changed their password"
		case strings.Contains(lower, "session has expired"):
			reason = "session expired"
		case strings.Contains(lower, "session has been invalidated"):
			reason = "session invalidated"
		case strings.Contains(lower, "has not authorized application"):
			reason = "user has not authorized the application"
		case strings.Contains(lower, "expired"):
			reason = "token expired"
		default:
			reason = "invalid access token"
		}
		return "Access token expired: " + reason
	}

	// Permission errors
	if strings.Contains(lower, "permission") && (strings.Contains(lower, "not granted") || strings.Contains(lower, "is needed")) {
		return "Insufficient permissions: required permissions not granted"
	}

	// Rate limiting
	if strings.Contains(lower, "rate limit") || strings.Contains(lower, "too many requests") || strings.Contains(lower, "quota") {
		return "Rate limit exceeded: too many API requests"
	}

	// Account/page not found
	if strings.Contains(lower, "does not exist") || strings.Contains(lower, "page not found") ||
		(strings.Contains(lower, "nonexist") && strings.Contains(lower, "page")) {
		return "Account or page not found"
	}

	// Unauthorized / authentication failed
	if lower == "unauthorized" || lower == "request failed with status 401: unauthorized" ||
		strings.Contains(lower, "authentication failed") ||
		strings.Contains(lower, "invalid authentication credentials") {
		return "Unauthorized: invalid credentials"
	}

	// Token invalid (TikTok, generic)
	if strings.Contains(lower, "access_token is invalid") || strings.Contains(lower, "access token is invalid") {
		return "Access token expired: token expired"
	}

	return s
}

func capLength(s string, max int) string {
	s = strings.TrimSpace(s)
	if len(s) > max {
		s = s[:max]
	}
	return s
}

// RecordProcessingError records a processing error on the exact account document identified by _id.
// Uses $set for error details and $inc for consecutive failure count.
func (r *unifiedSocialRepository) RecordProcessingError(ctx context.Context, id primitive.ObjectID, errorMessage string) error {
	errorMessage = cleanErrorMessage(errorMessage)

	update := bson.M{
		"$set": bson.M{
			"meta_data.last_processing_error":    errorMessage,
			"meta_data.last_processing_error_at": time.Now().UTC().Format(time.RFC3339),
			"state":                              mongo3.StateFailed,
			"updated_at":                         mongo3.MongoTime{Time: time.Now().UTC()},
		},
		"$inc": bson.M{
			"meta_data.consecutive_failures": 1,
		},
	}

	result, err := r.collection.UpdateByID(ctx, id, update)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Str("id", id.Hex()).
			Msg("Failed to record processing error")
		return err
	}

	if result.MatchedCount == 0 {
		r.logger.Warn().Str("id", id.Hex()).Msg("No social account found to record processing error")
		return mongo.ErrNoDocuments
	}

	r.logger.Info().
		Str("id", id.Hex()).
		Str("error_message", errorMessage).
		Msg("Recorded processing error with state=Failed")

	return nil
}

// ClearProcessingError removes processing error fields from the account's meta_data.
// Called after successful processing to clear stale error state.
func (r *unifiedSocialRepository) ClearProcessingError(ctx context.Context, id primitive.ObjectID) error {
	update := bson.M{
		"$unset": bson.M{
			"meta_data.last_processing_error":    "",
			"meta_data.last_processing_error_at": "",
			"meta_data.consecutive_failures":     "",
		},
		"$set": bson.M{
			"updated_at": mongo3.MongoTime{Time: time.Now().UTC()},
		},
	}

	result, err := r.collection.UpdateByID(ctx, id, update)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Str("id", id.Hex()).
			Msg("Failed to clear processing error")
		return err
	}

	if result.MatchedCount == 0 {
		return nil
	}

	r.logger.Debug().
		Str("id", id.Hex()).
		Msg("Cleared processing error")

	return nil
}

// Create inserts a new social account into the collection
func (r *unifiedSocialRepository) Create(ctx context.Context, account *mongo3.SocialIntegration) (primitive.ObjectID, error) {
	if account.ID.IsZero() {
		account.ID = primitive.NewObjectID()
	}

	if account.CreatedAt == nil {
		now := time.Now().UTC()
		account.CreatedAt = &mongo3.MongoTime{Time: now}
	}

	if account.UpdatedAt == nil {
		now := time.Now().UTC()
		account.UpdatedAt = &mongo3.MongoTime{Time: now}
	}

	// Ensure platform_identifier is set from legacy fields if needed
	if account.PlatformIdentifier == "" {
		account.PlatformIdentifier = account.GetPlatformID()
	}

	result, err := r.collection.InsertOne(ctx, account)
	if err != nil {
		r.logger.Warn().
			Err(err).
			Interface("account", account).
			Msg("Error creating social account")
		return primitive.NilObjectID, err
	}

	insertedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		r.logger.Warn().
			Interface("inserted_id", result.InsertedID).
			Msg("Failed to cast inserted ID to ObjectID")
		return primitive.NilObjectID, errors.New("failed to cast inserted ID")
	}

	r.logger.Info().
		Str("id", insertedID.Hex()).
		Str("platform_type", account.PlatformType).
		Str("platform_id", account.PlatformIdentifier).
		Msg("Social account created successfully")

	return insertedID, nil
}

// Delete soft deletes a social account by updating its state
func (r *unifiedSocialRepository) Delete(ctx context.Context, id primitive.ObjectID) error {
	return r.UpdateState(ctx, id, mongo3.StateDeleted)
}

// InsertTwitterJobMetadata inserts a twitter job execution record into twitter_jobs_metadata.
func (r *unifiedSocialRepository) InsertTwitterJobMetadata(ctx context.Context, payload TwitterJobMetadataPayload) error {
	now := time.Now().UTC()

	doc := bson.M{
		"platform_id":     payload.PlatformID,
		"workspace_id":    payload.WorkspaceID,
		"platform_type":   "twitter",
		"job_type":        "posts",
		"credits_used":    payload.CreditsUsed,
		"executed_by":     firstNonEmptyString(payload.ExecutedBy, "internal"),
		"app_name":        payload.AppName,
		"job_executed_at": now,
		"day_of_week":     weekdayToOneBased(now.Weekday()),
		"hour_of_day":     now.Hour(),
		"updated_at":      now,
		"created_at":      now,
	}

	if appObjectID, err := primitive.ObjectIDFromHex(strings.TrimSpace(payload.AppID)); err == nil {
		doc["app_id"] = appObjectID
	} else if strings.TrimSpace(payload.AppID) != "" {
		doc["app_id"] = payload.AppID
	}

	if _, err := r.collection.Database().Collection("twitter_jobs_metadata").InsertOne(ctx, doc); err != nil {
		return err
	}
	return nil
}

// buildLegacyFilter creates a filter using legacy field names for backward compatibility
func (r *unifiedSocialRepository) buildLegacyFilter(platformType, platformID string) bson.M {
	switch platformType {
	case mongo3.PlatformFacebook:
		return bson.M{
			"platform_type": platformType,
			"facebook_id":   platformID,
		}
	case mongo3.PlatformInstagram:
		return bson.M{
			"platform_type": platformType,
			"instagram_id":  platformID,
		}
	case mongo3.PlatformLinkedIn:
		return bson.M{
			"platform_type": platformType,
			"linkedin_id":   platformID,
		}
	case mongo3.PlatformTwitter:
		return bson.M{
			"platform_type": platformType,
			"twitter_id":    platformID,
		}
	case mongo3.PlatformGMB:
		return bson.M{
			"platform_type": platformType,
			"location_id":   platformID,
		}
	case mongo3.PlatformPinterest:
		return bson.M{
			"platform_type": platformType,
			"pinterest_id":  platformID,
		}
	default:
		return nil
	}
}

// TwitterDeveloperApp represents an analytics-enabled twitter developer app.
type TwitterDeveloperApp struct {
	ID        primitive.ObjectID
	APIKey    string
	APISecret string
	AppName   string
}

// TwitterJobSetting represents settings in twitter_job_settings.
type TwitterJobSetting struct {
	PlatformID string
	JobType    string
	TriggerDay int
	PostCount  int
}

// TwitterRepository reads twitter-specific configuration from MongoDB.
type TwitterRepository struct {
	developerAppsCollection *mongo.Collection
	jobSettingsCollection   *mongo.Collection
}

// NewTwitterRepository creates a new TwitterRepository.
func NewTwitterRepository(db *mongo.Database) *TwitterRepository {
	return &TwitterRepository{
		developerAppsCollection: db.Collection("developer_apps"),
		jobSettingsCollection:   db.Collection("twitter_job_settings"),
	}
}

// GetAnalyticsEnabledDeveloperApps returns twitter developer apps where analytics_enabled=true.
func (r *TwitterRepository) GetAnalyticsEnabledDeveloperApps(ctx context.Context) (map[string]TwitterDeveloperApp, error) {
	cursor, err := r.developerAppsCollection.Find(ctx, bson.M{
		"analytics_enabled": true,
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var raw []bson.M
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, err
	}

	apps := make(map[string]TwitterDeveloperApp, len(raw))
	for _, doc := range raw {
		id, ok := doc["_id"].(primitive.ObjectID)
		if !ok {
			continue
		}
		apps[id.Hex()] = TwitterDeveloperApp{
			ID:        id,
			APIKey:    getStringFromMap(doc, "api_key"),
			APISecret: getStringFromMap(doc, "api_secret"),
			AppName:   getStringFromMap(doc, "app_name"),
		}
	}

	return apps, nil
}

// GetAnalyticsEnabledDeveloperAppByID returns one analytics-enabled developer app by string ObjectID.
func (r *TwitterRepository) GetAnalyticsEnabledDeveloperAppByID(ctx context.Context, developerAppID string) (*TwitterDeveloperApp, error) {
	if developerAppID == "" {
		return nil, nil
	}
	appObjectID, err := primitive.ObjectIDFromHex(developerAppID)
	if err != nil {
		return nil, nil
	}

	var appDoc bson.M
	err = r.developerAppsCollection.FindOne(ctx, bson.M{
		"_id":               appObjectID,
		"analytics_enabled": true,
	}).Decode(&appDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &TwitterDeveloperApp{
		ID:        appObjectID,
		APIKey:    getStringFromMap(appDoc, "api_key"),
		APISecret: getStringFromMap(appDoc, "api_secret"),
		AppName:   getStringFromMap(appDoc, "app_name"),
	}, nil
}

// GetJobSettingsByPlatformIDs fetches job settings keyed by platform_id.
func (r *TwitterRepository) GetJobSettingsByPlatformIDs(ctx context.Context, platformIDs []string) (map[string]TwitterJobSetting, error) {
	settings := make(map[string]TwitterJobSetting)
	if len(platformIDs) == 0 {
		return settings, nil
	}

	cursor, err := r.jobSettingsCollection.Find(ctx, bson.M{
		"platform_id": bson.M{"$in": platformIDs},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var raw []bson.M
	if err := cursor.All(ctx, &raw); err != nil {
		return nil, err
	}

	for _, doc := range raw {
		platformID := getStringFromMap(doc, "platform_id")
		if platformID == "" {
			continue
		}
		settings[platformID] = TwitterJobSetting{
			PlatformID: platformID,
			JobType:    getStringFromMap(doc, "job_type"),
			TriggerDay: getIntFromMap(doc, "trigger_day"),
			PostCount:  getIntFromMap(doc, "post_count"),
		}
	}

	return settings, nil
}

// GetJobSettingByPlatformID fetches one twitter job setting by platform_id.
func (r *TwitterRepository) GetJobSettingByPlatformID(ctx context.Context, platformID string) (*TwitterJobSetting, error) {
	if platformID == "" {
		return nil, nil
	}

	var settingDoc bson.M
	err := r.jobSettingsCollection.FindOne(ctx, bson.M{
		"platform_id": platformID,
	}).Decode(&settingDoc)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, err
	}

	return &TwitterJobSetting{
		PlatformID: platformID,
		JobType:    getStringFromMap(settingDoc, "job_type"),
		TriggerDay: getIntFromMap(settingDoc, "trigger_day"),
		PostCount:  getIntFromMap(settingDoc, "post_count"),
	}, nil
}

func getStringFromMap(doc bson.M, key string) string {
	value, exists := doc[key]
	if !exists {
		return ""
	}
	if stringValue, ok := value.(string); ok {
		return stringValue
	}
	return fmt.Sprint(value)
}

func getIntFromMap(doc bson.M, key string) int {
	value, exists := doc[key]
	if !exists {
		return 0
	}

	switch v := value.(type) {
	case int:
		return v
	case int32:
		return int(v)
	case int64:
		return int(v)
	case float64:
		return int(v)
	case string:
		i, err := strconv.Atoi(v)
		if err == nil {
			return i
		}
	}

	return 0
}

func firstNonEmptyString(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return value
		}
	}
	return ""
}

func weekdayToOneBased(day time.Weekday) int {
	if day == time.Sunday {
		return 7
	}
	return int(day)
}
