package main

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	kafka2 "github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"
)

// Token/Auth error types - used to stop processing early when token is invalid
var (
	// ErrTokenInvalid indicates the access token is invalid, expired, or revoked
	ErrTokenInvalid = errors.New("linkedin token invalid or expired")
	// ErrTokenPermissionDenied indicates the token lacks required permissions
	ErrTokenPermissionDenied = errors.New("linkedin token permission denied")
)

// isTokenError checks if an error message indicates a token/auth problem.
// LinkedIn returns these errors for invalid, expired, or permission-denied tokens.
func isTokenError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "401") ||
		strings.Contains(errStr, "403") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "expired") ||
		strings.Contains(errStr, "invalid_token") ||
		strings.Contains(errStr, "invalid access token") ||
		strings.Contains(errStr, "access denied") ||
		strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "not authorized")
}

// isExpectedError checks if an error is expected (permissions/auth) and should not be sent to Sentry
func isExpectedError(err error) bool {
	if err == nil {
		return false
	}
	return social.IsExpectedCompetitorErrorLI(err)
}

// wrapTokenError wraps an error as a token error if it appears to be auth-related.
// This allows callers to check with errors.Is(err, ErrTokenInvalid).
func wrapTokenError(err error) error {
	if err == nil {
		return nil
	}
	if isTokenError(err) {
		errStr := strings.ToLower(err.Error())
		if strings.Contains(errStr, "403") || strings.Contains(errStr, "permission") {
			return errors.Join(ErrTokenPermissionDenied, err)
		}
		return errors.Join(ErrTokenInvalid, err)
	}
	return err
}

// chunk splits a slice into smaller slices of maximum size n.
// Used for batching API calls to respect LinkedIn's request limits.
// Returns nil if input is empty or n <= 0.
func chunk[T any](in []T, n int) [][]T {
	if n <= 0 || len(in) == 0 {
		return nil
	}
	var out [][]T
	for len(in) > n {
		out = append(out, in[:n])
		in = in[n:]
	}
	return append(out, in)
}

// mapKeysToSlice converts a set (map[string]struct{}) to a slice of strings.
// Used after deduplicating IDs to prepare them for API calls.
func mapKeysToSlice(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}

// semForAccount returns or creates a per-account semaphore.
// This prevents duplicate processing of the same LinkedIn account ID.
// Uses sync.Map for thread-safe access across workers.
func semForAccount(id string) *semaphore.Weighted {
	if v, ok := accountSemaphores.Load(id); ok {
		return v.(*semaphore.Weighted)
	}
	sem := semaphore.NewWeighted(perAccountConcurrency)
	// LoadOrStore ensures only one semaphore is created even with concurrent access
	if old, loaded := accountSemaphores.LoadOrStore(id, sem); loaded {
		return old.(*semaphore.Weighted)
	}
	return sem
}

// decryptToken attempts to decrypt an encrypted access token.
// Returns the original token if decryption fails (token may already be plain text).
// Returns empty string if input token is empty.
func decryptToken(token, decryptionKey string) string {
	if token == "" {
		return ""
	}
	if dec, err := crypto.DecryptToken(token, decryptionKey); err == nil && dec != "" {
		return dec
	}
	return token
}

// calculateDateRanges calculates the date ranges for data fetching based on sync type.
// LinkedIn data is typically available with a 2-day lag (today - 2).
//
// For incremental sync: fetches last 10 days of data
// For full sync: fetches last 365 days of data
//
// Returns:
//   - cutoffTime: posts older than this are excluded
//   - startDate: start of date range for insights
//   - endDate: end of date range for insights
func calculateDateRanges(syncType string) (cutoffTime, startDate, endDate time.Time) {
	now := time.Now().UTC()
	// LinkedIn data has ~2 day lag, so last available data is today - 2
	lastAvailableDate := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -2)

	if strings.ToLower(syncType) == "incremental" {
		// Incremental: 10 days of data (e.g., Dec 5 - Dec 14 if today is Dec 17)
		cutoffTime = lastAvailableDate.AddDate(0, 0, -9)
		startDate = time.Date(lastAvailableDate.Year(), lastAvailableDate.Month(), lastAvailableDate.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -9)
	} else {
		// Full sync: 365 days of data
		cutoffTime = lastAvailableDate.AddDate(0, 0, -364)
		startDate = time.Date(lastAvailableDate.Year(), lastAvailableDate.Month(), lastAvailableDate.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -364)
	}

	// End date is the day after last available date (exclusive end)
	endDate = time.Date(lastAvailableDate.Year(), lastAvailableDate.Month(), lastAvailableDate.Day(), 23, 59, 59, 0, time.UTC).AddDate(0, 0, 1)
	return
}

// createOperation creates a logger operation for tracking work order processing.
// Operations provide timing info and are automatically captured in Sentry.
// Includes full context: account_id, linkedin_id, workspace_id, entity_type, sync_type.
func createOperation(log *logger.Logger, order LinkedInAccountWorkOrder, entityType string) *logger.Operation {
	return log.Operation("LinkedInProcessWorkOrder").
		WithFields(map[string]interface{}{
			"account_id":   order.ID,
			"linkedin_id":  order.LinkedinID,
			"workspace_id": order.WorkspaceID,
			"sync_type":    order.SyncType,
			"entity_type":  entityType,
			"account_type": string(order.AccountType),
		}).
		WithSentryTags(map[string]string{
			"platform":     "linkedin",
			"component":    "fetcher",
			"account_id":   order.ID,
			"linkedin_id":  order.LinkedinID,
			"workspace_id": order.WorkspaceID,
			"entity_type":  entityType,
			"sync_type":    order.SyncType,
		})
}

// createLoggerWithContext creates a logger with full account context.
// Use this for consistent logging throughout the processing pipeline.
func createLoggerWithContext(order LinkedInAccountWorkOrder, entityType, funcName string) *logger.Logger {
	log := logger.New("info")
	return &logger.Logger{Logger: log.With().
		Str("fn", funcName).
		Str("account_id", order.ID).
		Str("linkedin_id", order.LinkedinID).
		Str("workspace_id", order.WorkspaceID).
		Str("entity_type", entityType).
		Str("sync_type", order.SyncType).
		Logger()}
}

// parseStatsBatch parses LinkedIn share statistics API response.
// Maps activity IDs (ugcPost or share URN) to their totalShareStatistics.
// Used for enriching posts with engagement metrics.
func parseStatsBatch(body []byte) assetMap {
	type elem struct {
		UGCPost string         `json:"ugcPost"` // URN for UGC posts
		Share   string         `json:"share"`   // URN for share posts
		Total   map[string]any `json:"totalShareStatistics"`
	}
	var resp struct {
		Elements []elem `json:"elements"`
	}
	_ = json.Unmarshal(body, &resp)

	out := make(assetMap, len(resp.Elements))
	for _, e := range resp.Elements {
		// Activity can be either ugcPost or share type
		id := e.UGCPost
		if id == "" {
			id = e.Share
		}
		if id != "" {
			out[id] = e.Total
		}
	}
	return out
}

// parseAssetBatch parses LinkedIn asset (images/videos/documents) API response.
// Maps asset IDs to their full data objects.
// Handles both "id" and "asset" field naming conventions.
func parseAssetBatch(body []byte) assetMap {
	var resp struct {
		Results map[string]map[string]any `json:"results"`
	}
	_ = json.Unmarshal(body, &resp)

	out := make(assetMap, len(resp.Results))
	for _, m := range resp.Results {
		// LinkedIn API uses "id" for most assets
		if id, ok := m["id"].(string); ok && id != "" {
			out[id] = m
			continue
		}
		// Some payloads use "asset" instead of "id"
		if id, ok := m["asset"].(string); ok && id != "" {
			out[id] = m
		}
	}
	return out
}

// fetchAssetsInChunks is a generic function to fetch assets in parallel chunks.
// It handles semaphore-based concurrency limiting, chunking, and result aggregation.
// This reduces code duplication across fetchImages, fetchVideos, fetchDocuments, etc.
func fetchAssetsInChunks(ctx context.Context, eg *errgroup.Group, cfg *assetFetchConfig) {
	for _, ids := range chunk(cfg.ids, cfg.chunkSize) {
		ids := ids // Capture for goroutine
		eg.Go(func() error {
			// Acquire semaphore to limit concurrent API calls
			if err := cfg.semaphore.Acquire(ctx, 1); err != nil {
				return nil
			}
			defer cfg.semaphore.Release(1)

			body, err := cfg.fetchFunc(ctx, ids)
			if err != nil {
				cfg.log.Error().Err(err).
					Str("linkedin_id", cfg.linkedinID).
					Int(cfg.name+"_count", len(ids)).
					Msgf("failed to fetch linkedin %s", cfg.name)
				return nil
			}

			if body != nil {
				parsed := cfg.parseFunc(body)
				cfg.mu.Lock()
				for k, v := range parsed {
					cfg.resultMap[k] = v
				}
				cfg.mu.Unlock()
			}
			return nil
		})
	}
}

// fetchAndPublishOrgDetails fetches organization details and publishes to Kafka.
// Only called for page/organization accounts, not profiles.
func fetchAndPublishOrgDetails(
	ctx context.Context,
	li *social.LinkedInClient,
	producer kafka2.Producer,
	linkedinID, token, outputTopic string,
	log *logger.Logger,
) error {
	body, err := li.FetchOrganizationDetailsRaw(ctx, linkedinID, token)
	if err != nil {
		log.Error().Err(err).Str("linkedin_id", linkedinID).Msg("failed to fetch linkedin organization details")
		return nil
	}
	if body != nil {
		_ = producer.Produce(ctx, outputTopic, []byte(linkedinID), body)
	}
	return nil
}
