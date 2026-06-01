// Package social contains clients for interacting with social media platform APIs.
package social

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	// igAPIVersion defines the API version for Instagram Graph API.
	igAPIVersion = "v19.0"
	// igBaseURL is the default base URL when accessing via Facebook Graph API domain.
	igBaseURL = "https://graph.facebook.com/" // business accounts are accessed via FB domain
	// igMaxPagesToFetch is a safety brake to prevent infinite loops during pagination.
	igMaxPagesToFetch = 100
	// igMediaFields is the list of fields to request for each media item (mirrors Python code).
	igMediaFields = "id,comments_count,thumbnail_url,caption,username,like_count,hashtags,media_type,media_product_type,media_url,timestamp,children{media_type,media_url,thumbnail_url},permalink"
	// igStoryFields is the list of fields to request for each story item.
	igStoryFields = "id,comments_count,thumbnail_url,caption,username,like_count,hashtags,media_type,media_product_type,media_url,timestamp,children{media_type,media_url,thumbnail_url},permalink"

	// Retry configuration
	igMaxAttempts   = 5
	igRefreshFields = "id,media_type,media_product_type,media_url,thumbnail_url,children{media_type,media_url,thumbnail_url}"
)

// InstagramAuthError represents an authentication/authorization error from Instagram API
type InstagramAuthError struct {
	Message    string
	StatusCode int
	ErrorCode  int
}

func (e *InstagramAuthError) Error() string {
	return fmt.Sprintf("instagram auth error (status=%d, code=%d): %s", e.StatusCode, e.ErrorCode, e.Message)
}

// IsAuthError checks if an error is an authentication/authorization error
// These include: expired tokens, invalid tokens, permissions issues
func IsAuthError(err error) bool {
	if err == nil {
		return false
	}
	// Check if it's our custom auth error type
	if _, ok := err.(*InstagramAuthError); ok {
		return true
	}
	// Check error message for common auth-related patterns
	errStr := strings.ToLower(err.Error())

	// Exclude known non-auth errors that have OAuthException type
	// "(#10) Not enough viewers for the media to show insights" - story/media just needs more views
	// Note: "(#10) Application does not have permission" IS an auth error, handled below
	if strings.Contains(errStr, "(#10)") && strings.Contains(errStr, "not enough viewers") {
		return false
	}

	authPatterns := []string{
		"invalid oauth",
		"access token",
		"token has expired",
		"session has expired",
		"status 401",
		"status 403",
		"unauthorized",
		"(#190)",                               // Access token expired
		"(#102)",                               // Session invalid
		"(#100)",                               // Invalid parameter (often token related)
		"(#4)",                                 // Application request limit
		"(#200)",                               // Permission error
		"application does not have permission", // (#10) Permission error
		"does not have permission for this action", // (#10) Permission error variant
	}
	for _, pattern := range authPatterns {
		if strings.Contains(errStr, pattern) {
			return true
		}
	}
	return false
}

// IsExpectedCompetitorError checks if an error is an expected competitor API error that should not be sent to Sentry
// These are permission/auth errors (400-404) that are expected when fetching competitor data
func IsExpectedCompetitorError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()

	// Expected permission/auth errors from Instagram API for competitors (produce 400, 401, 403, 404)
	expectedPatterns := []string{
		"Application does not have permission for this action",
		"OAuthException/10",
		"Invalid OAuth access token",
		"Cannot parse access token",
		"OAuthException/190",
		"user must be an administrator, editor, or moderator",
		"does not exist, cannot be loaded due to missing permissions",
		"GraphMethodException/100",
		"This Page access token belongs to a Page that is not accessible",
	}

	for _, pattern := range expectedPatterns {
		if strings.Contains(errMsg, pattern) {
			return true
		}
	}

	return false
}

// igApiError represents the structure of an error response from the Instagram Graph API.
type igApiError struct {
	Error struct {
		Message   string `json:"message"`
		Type      string `json:"type"`
		Code      int    `json:"code"`
		FBTraceID string `json:"fbtrace_id"`
	} `json:"error"`
}

// InstagramClient is a client for interacting with the Instagram Graph API.
// It focuses on fetching media data similar to the legacy Python implementation.
// Additional endpoints can be added as needed.
// NOTE: Insights per media are fetched separately (not yet implemented).
type InstagramClient struct {
	httpClient *http.Client
	baseURL    string
	appSecret  string
	log        *logger.Logger
	rate       *RateManager
}

// NewInstagramClient returns a new InstagramClient with default rate limits.
// appSecret is required for generating appsecret_proof (HMAC token security).
// If you want to use the instagram.com domain (connected via IG login), override baseURL after construction.
func NewInstagramClient(appSecret string) *InstagramClient {
	return NewInstagramClientWithRates(appSecret, NewRateManager(RateLimits{}))
}

// NewInstagramClientWithRates creates a client wired to a shared RateManager.
func NewInstagramClientWithRates(appSecret string, rm *RateManager) *InstagramClient {
	if rm == nil {
		rm = NewRateManager(RateLimits{})
	}
	return &InstagramClient{
		httpClient: &http.Client{Timeout: 45 * time.Second},
		baseURL:    igBaseURL,
		appSecret:  appSecret,
		log:        logger.New("info"),
		rate:       rm,
	}
}

// WithBaseURL allows overriding the default baseURL (e.g., "https://graph.instagram.com").
func (c *InstagramClient) WithBaseURL(url string) *InstagramClient {
	c.baseURL = url
	return c
}

// waitRate applies global + per-token throttling in one place.
func (c *InstagramClient) waitRate(ctx context.Context, token string) error {
	if c.rate == nil {
		// extremely defensive; should not happen
		c.rate = NewRateManager(RateLimits{})
	}
	return c.rate.Wait(ctx, token)
}

// generateAppSecretProof produces the SHA256 HMAC of the access token using the app secret.
func (c *InstagramClient) generateAppSecretProof(accessToken string) string {
	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(accessToken))
	return hex.EncodeToString(h.Sum(nil))
}

// doWithRetry executes the HTTP request with rate-limit waits and exponential backoff + jitter.
func (c *InstagramClient) doWithRetry(
	ctx context.Context,
	instagramID string,
	req *http.Request,
	caller string,
) (body []byte, status int, err error) {

	for attempt := 1; attempt <= igMaxAttempts; attempt++ {
		if err = c.waitRate(ctx, req.URL.Query().Get("access_token")); err != nil {
			c.log.Warn().
				Str("instagram_id", instagramID).
				Str("caller", caller).
				Err(err).
				Msg("Instagram API: rate limit wait failed")
			return nil, 0, fmt.Errorf("InstagramClient.doWithRetry: rate limit wait failed: %w", err)
		}

		resp, httpErr := c.httpClient.Do(req)
		if httpErr != nil {
			if attempt == igMaxAttempts {
				c.log.Warn().
					Str("instagram_id", instagramID).
					Str("caller", caller).
					Int("attempt", attempt).
					Err(httpErr).
					Msg("Instagram API: HTTP request failed; max attempts reached")
				return nil, 0, fmt.Errorf("InstagramClient.doWithRetry: request failed: %w", httpErr)
			}
			delay := computeBackoff(attempt)
			c.log.Warn().
				Str("instagram_id", instagramID).
				Str("caller", caller).
				Int("attempt", attempt).
				Dur("backoff", delay).
				Err(httpErr).
				Msg("Instagram API: HTTP request failed; retrying with backoff")
			time.Sleep(delay)
			continue
		}

		body, _ = io.ReadAll(resp.Body)
		resp.Body.Close()
		status = resp.StatusCode

		if status == http.StatusOK {
			c.log.Debug().
				Str("instagram_id", instagramID).
				Msg("Instagram API: got 200 OK")
			return body, status, nil
		}

		// Non-200: parse Instagram error (if present), then honor Retry-After or backoff.
		var igErr igApiError
		_ = json.Unmarshal(body, &igErr)

		// Construct error for checking if it's expected
		var currentErr error
		if igErr.Error.Message != "" {
			currentErr = fmt.Errorf("InstagramClient.doWithRetry: instagram API error: %s (%s/%d)", igErr.Error.Message, igErr.Error.Type, igErr.Error.Code)
		} else {
			currentErr = fmt.Errorf("InstagramClient.doWithRetry: http %d: %s", status, string(body))
		}

		// Check if this is an expected client error (4xx permission/auth issues).
		// Code 3006 ("Not enough users") means the account has too few followers for
		// Instagram to expose audience insights — permanent, never worth retrying.
		isExpected := status >= 400 && status < 500 && (IsExpectedCompetitorError(currentErr) || igErr.Error.Code == 3006)

		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if s, _ := strconv.Atoi(strings.TrimSpace(ra)); s > 0 {
				c.log.Warn().
					Str("instagram_id", instagramID).
					Str("caller", caller).
					Int("attempt", attempt).
					Int("retry_after_sec", s).
					Int("status_code", status).
					Str("ig_error_type", igErr.Error.Type).
					Int("ig_error_code", igErr.Error.Code).
					Str("ig_error_message", igErr.Error.Message).
					Msg("Instagram API: non-200; respecting Retry-After")
				time.Sleep(time.Duration(s) * time.Second)
			}
		} else {
			// For expected errors on first attempt, don't retry
			if isExpected && attempt == 1 {
				c.log.Warn().
					Str("instagram_id", instagramID).
					Str("caller", caller).
					Int("status_code", status).
					Str("ig_error_type", igErr.Error.Type).
					Int("ig_error_code", igErr.Error.Code).
					Str("ig_error_message", igErr.Error.Message).
					Msg("Instagram API: expected client error (permission/auth); not retrying")
				return nil, status, currentErr
			}

			delay := computeBackoff(attempt)
			c.log.Warn().
				Str("instagram_id", instagramID).
				Str("caller", caller).
				Int("attempt", attempt).
				Int("status_code", status).
				Dur("backoff", delay).
				Str("ig_error_type", igErr.Error.Type).
				Int("ig_error_code", igErr.Error.Code).
				Str("ig_error_message", igErr.Error.Message).
				Msg("Instagram API: non-200; retrying with backoff")
			time.Sleep(delay)
		}

		if attempt == igMaxAttempts {
			// Final error with payload details if available
			if igErr.Error.Message != "" {
				// Use Warn for expected errors, Error for unexpected
				if isExpected {
					c.log.Warn().
						Str("instagram_id", instagramID).
						Str("caller", caller).
						Int("status_code", status).
						Str("ig_error_type", igErr.Error.Type).
						Int("ig_error_code", igErr.Error.Code).
						Str("ig_error_message", igErr.Error.Message).
						Msg("Instagram API: expected client error after max attempts")
				} else {
					c.log.Warn().
						Str("instagram_id", instagramID).
						Str("caller", caller).
						Int("status_code", status).
						Str("ig_error_type", igErr.Error.Type).
						Int("ig_error_code", igErr.Error.Code).
						Str("ig_error_message", igErr.Error.Message).
						Msg("Instagram API: giving up after max attempts (Instagram error)")
				}
				return nil, status, currentErr
			}
			if isExpected {
				c.log.Warn().
					Str("instagram_id", instagramID).
					Str("caller", caller).
					Int("status_code", status).
					Msg("Instagram API: expected client error after max attempts (HTTP)")
			} else {
				c.log.Warn().
					Str("instagram_id", instagramID).
					Str("caller", caller).
					Int("status_code", status).
					Msg("Instagram API: giving up after max attempts (HTTP)")
			}
			return nil, status, currentErr
		}
	}
	return nil, 0, fmt.Errorf("InstagramClient.doWithRetry: unreachable")
}

// FetchAccountMedia retrieves profile info and media for the specified Instagram ID.
// The returned struct mirrors the RawInstagramAccountResponse model (name, profile pic, media list).
// "limit" controls the media count requested in one call (max 100 according to API).
// The Python implementation first attempts an extended fields set that includes stories; if that fails, it retries with fewer fields.
// Here we mimic the same behaviour.
func (c *InstagramClient) FetchAccountMedia(ctx context.Context, instagramID, accessToken string, limit int) (*kafkamodels.RawInstagramAccountResponse, error) {
	startTime := time.Now()
	c.log.Info().
		Str("instagram_id", instagramID).
		Int("limit", limit).
		Str("module", "instagram_client").
		Msg("Starting FetchAccountMedia")

	if limit <= 0 {
		limit = 100
	}

	// two field sets similar to Python logic
	// Fields are URL-encoded to avoid issues when we embed them later in query param.
	firstFields := fmt.Sprintf("name,profile_picture_url,media.limit(%d){id,comments_count,thumbnail_url,caption,username,like_count,hashtags,media_type,media_product_type,media_url,timestamp,children{media_type,media_url,thumbnail_url},permalink},stories{id,media_type,media_url,owner,timestamp,username,permalink}", limit)
	secondFields := fmt.Sprintf("name,profile_picture_url,media.limit(%d){id,comments_count,caption,username,like_count,media_product_type,media_type,media_url,timestamp,children,permalink}", limit)

	// try first request
	resp, err := c.doAccountRequest(ctx, instagramID, accessToken, firstFields)
	if err != nil {
		// If we received handled error (400) but error type matched, try second fields set
		c.log.Warn().Err(err).Str("instagram_id", instagramID).Msg("First media request failed, retrying with reduced field set")
		resp, err = c.doAccountRequest(ctx, instagramID, accessToken, secondFields)
		if err != nil {
			return nil, err
		}
	}

	elapsed := time.Since(startTime)
	c.log.Info().
		Str("instagram_id", instagramID).
		Dur("elapsed_time", elapsed).
		Int("media_count", len(resp.Media.Data)).
		Str("module", "instagram_client").
		Msg("Completed FetchAccountMedia")

	return resp, nil
}

// doAccountRequest builds and performs the underlying HTTP request for FetchAccountMedia.
func (c *InstagramClient) doAccountRequest(ctx context.Context, instagramID, accessToken, fields string) (*kafkamodels.RawInstagramAccountResponse, error) {
	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s", c.baseURL, igAPIVersion, instagramID))
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.doAccountRequest: failed to parse base URL: %w", err)
	}

	q := apiURL.Query()
	q.Set("fields", fields)
	q.Set("access_token", accessToken)
	apiURL.RawQuery = q.Encode()

	// Build request with appsecret_proof
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.doAccountRequest: failed to create request: %w", err)
	}

	proof := c.generateAppSecretProof(accessToken)
	req.URL.RawQuery = req.URL.Query().Encode() + "&appsecret_proof=" + proof

	c.log.Info().Str("instagram_id", instagramID).Msg("Fetching Instagram account media")

	body, status, err := c.doWithRetry(ctx, instagramID, req, "FetchAccountMedia")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("InstagramClient.doAccountRequest: instagram API returned status %d: %s", status, string(body))
	}

	var apiResp kafkamodels.RawInstagramAccountResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("InstagramClient.doAccountRequest: failed to decode response: %w", err)
	}

	return &apiResp, nil
}

// FetchMedia retrieves all media items for the given Instagram account with a default page limit.
func (c *InstagramClient) FetchMedia(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
	return c.FetchMediaWithLimit(ctx, instagramID, accessToken, igMaxPagesToFetch)
}

// FetchMediaSince retrieves media items created after the 'since' time.
// Uses early termination - stops fetching when it encounters media older than 'since'
// since Instagram returns media in reverse chronological order (newest first).
// This is more efficient than fetching all pages and filtering afterwards.
func (c *InstagramClient) FetchMediaSince(ctx context.Context, instagramID, accessToken string, since time.Time) ([]kafkamodels.RawInstagramMedia, error) {
	var allMedia []kafkamodels.RawInstagramMedia

	c.log.Info().
		Str("instagram_id", instagramID).
		Time("since", since).
		Str("module", "instagram_client").
		Msg("Starting to fetch Instagram media with date filter")

	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/media", c.baseURL, igAPIVersion, instagramID))
	if err != nil {
		c.log.Warn().Err(err).Str("instagram_id", instagramID).Msg("Failed to parse media base URL")
		return nil, fmt.Errorf("InstagramClient.FetchMediaSince: failed to parse base URL: %w", err)
	}

	q := apiURL.Query()
	q.Set("fields", igMediaFields)
	q.Set("access_token", accessToken)
	q.Set("limit", "100")
	apiURL.RawQuery = q.Encode()

	nextURL := apiURL.String()
	pagesFetched := 0
	startTime := time.Now()
	reachedOldMedia := false

	for nextURL != "" && pagesFetched < igMaxPagesToFetch && !reachedOldMedia {
		select {
		case <-ctx.Done():
			c.log.Warn().Str("instagram_id", instagramID).Msg("Context cancelled during media fetch")
			return nil, ctx.Err()
		default:
		}

		c.log.Debug().Str("instagram_id", instagramID).Int("current_page", pagesFetched+1).Msg("Fetching media page")

		req, err := http.NewRequestWithContext(ctx, "GET", nextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("InstagramClient.FetchMediaSince: failed to create request: %w", err)
		}

		proof := c.generateAppSecretProof(accessToken)
		rq := req.URL.Query()
		rq.Set("appsecret_proof", proof)
		req.URL.RawQuery = rq.Encode()

		body, status, err := c.doWithRetry(ctx, instagramID, req, "FetchMediaSince")
		if err != nil {
			return nil, fmt.Errorf("InstagramClient.FetchMediaSince: failed HTTP request: %w", err)
		}

		if status != http.StatusOK {
			return nil, fmt.Errorf("InstagramClient.FetchMediaSince: instagram API returned %d: %s", status, string(body))
		}

		var apiResp struct {
			Data   []kafkamodels.RawInstagramMedia `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("InstagramClient.FetchMediaSince: failed to decode JSON: %w", err)
		}

		// Filter media by timestamp and check for early termination
		for _, media := range apiResp.Data {
			// Instagram returns timestamps in format: 2025-04-16T11:27:51+0000
			// Try multiple formats to handle variations
			var mediaTime time.Time
			var parseErr error
			for _, layout := range []string{
				"2006-01-02T15:04:05-0700", // Instagram format without colon
				time.RFC3339,               // Standard format with colon
			} {
				mediaTime, parseErr = time.Parse(layout, media.Timestamp)
				if parseErr == nil {
					break
				}
			}
			if parseErr != nil {
				// If we can't parse timestamp, include the media anyway
				c.log.Warn().Str("media_id", media.ID).Str("timestamp", media.Timestamp).Msg("Failed to parse media timestamp, including anyway")
				allMedia = append(allMedia, media)
				continue
			}

			// Check if media is older than 'since' - early termination
			if mediaTime.Before(since) {
				c.log.Info().
					Str("instagram_id", instagramID).
					Time("media_time", mediaTime).
					Time("since", since).
					Int("media_collected", len(allMedia)).
					Msg("Reached media older than 'since', stopping pagination")
				reachedOldMedia = true
				break
			}

			allMedia = append(allMedia, media)
		}

		c.log.Info().
			Str("instagram_id", instagramID).
			Int("page", pagesFetched+1).
			Int("media_in_page", len(apiResp.Data)).
			Int("total_media", len(allMedia)).
			Bool("reached_old_media", reachedOldMedia).
			Msg("Fetched media page")

		nextURL = apiResp.Paging.Next
		pagesFetched++
	}

	elapsed := time.Since(startTime)

	c.log.Info().
		Str("instagram_id", instagramID).
		Int("pages_fetched", pagesFetched).
		Int("total_media", len(allMedia)).
		Time("since", since).
		Bool("early_termination", reachedOldMedia).
		Dur("elapsed", elapsed).
		Msg("Completed fetching Instagram media with date filter")

	return allMedia, nil
}

// FetchMediaWithLimit retrieves media items with pagination, similar to FacebookClient.FetchPostsWithLimit.
func (c *InstagramClient) FetchMediaWithLimit(ctx context.Context, instagramID, accessToken string, maxPages int) ([]kafkamodels.RawInstagramMedia, error) {
	var allMedia []kafkamodels.RawInstagramMedia

	c.log.Info().
		Str("instagram_id", instagramID).
		Int("max_pages", maxPages).
		Str("module", "instagram_client").
		Msg("Starting to fetch Instagram media")

	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/media", c.baseURL, igAPIVersion, instagramID))
	if err != nil {
		c.log.Warn().Err(err).Str("instagram_id", instagramID).Msg("Failed to parse media base URL")
		return nil, fmt.Errorf("InstagramClient.FetchMediaWithLimit: failed to parse base URL: %w", err)
	}

	q := apiURL.Query()
	q.Set("fields", igMediaFields)
	q.Set("access_token", accessToken)
	q.Set("limit", "100")
	apiURL.RawQuery = q.Encode()

	nextURL := apiURL.String()
	pagesFetched := 0
	startTime := time.Now()

	for nextURL != "" && pagesFetched < maxPages {
		select {
		case <-ctx.Done():
			c.log.Warn().Str("instagram_id", instagramID).Msg("Context cancelled during media fetch")
			return nil, ctx.Err()
		default:
		}

		c.log.Debug().Str("instagram_id", instagramID).Int("current_page", pagesFetched+1).Msg("Fetching media page")

		req, err := http.NewRequestWithContext(ctx, "GET", nextURL, nil)
		if err != nil {
			return nil, fmt.Errorf("InstagramClient.FetchMediaWithLimit: failed to create request: %w", err)
		}

		body, status, err := c.doWithRetry(ctx, instagramID, req, "FetchMedia")
		if err != nil {
			return nil, err
		}

		if status != http.StatusOK {
			return nil, fmt.Errorf("InstagramClient.FetchMediaWithLimit: instagram API returned %d: %s", status, string(body))
		}

		var apiResp struct {
			Data   []kafkamodels.RawInstagramMedia `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}

		if err := json.Unmarshal(body, &apiResp); err != nil {
			return nil, fmt.Errorf("InstagramClient.FetchMediaWithLimit: failed to decode JSON: %w", err)
		}

		allMedia = append(allMedia, apiResp.Data...)

		c.log.Info().Str("instagram_id", instagramID).Int("page", pagesFetched+1).Int("media_in_page", len(apiResp.Data)).Int("total_media", len(allMedia)).Msg("Fetched media page")

		nextURL = apiResp.Paging.Next
		pagesFetched++
	}

	elapsed := time.Since(startTime)
	if pagesFetched >= maxPages {
		c.log.Warn().Str("instagram_id", instagramID).Msg("Reached maximum page limit while fetching media")
	}

	c.log.Info().Str("instagram_id", instagramID).Int("pages_fetched", pagesFetched).Int("total_media", len(allMedia)).Dur("elapsed", elapsed).Msg("Completed fetching Instagram media")

	return allMedia, nil
}

// FetchAllMedia retrieves ALL media items for the given Instagram account.
//
// This paginates through all pages until there's no next page.
// WARNING: Use with caution - accounts with many posts will result in
// many API calls and large data transfers.
//
// For most use cases, prefer FetchMediaSince or FetchMediaWithLimit.
func (c *InstagramClient) FetchAllMedia(
	ctx context.Context,
	instagramID, accessToken string,
) ([]kafkamodels.RawInstagramMedia, error) {
	c.log.Info().
		Str("instagram_id", instagramID).
		Msg("Starting to fetch ALL Instagram media (no page limit)")

	// Build initial URL
	firstURL, err := c.buildMediaURL(instagramID, accessToken)
	if err != nil {
		return nil, err
	}

	// Paginate through all pages
	var allMedia []kafkamodels.RawInstagramMedia
	nextURL := firstURL
	pagesFetched := 0
	startTime := time.Now()

	for nextURL != "" {
		// Check for context cancellation
		select {
		case <-ctx.Done():
			c.log.Warn().
				Str("instagram_id", instagramID).
				Int("pages_so_far", pagesFetched).
				Int("media_so_far", len(allMedia)).
				Msg("Context cancelled during media fetch")
			return nil, ctx.Err()
		default:
		}

		// Fetch single page
		media, next, err := c.fetchMediaPage(ctx, nextURL, accessToken)
		if err != nil {
			return nil, err
		}

		allMedia = append(allMedia, media...)
		pagesFetched++

		c.log.Info().
			Str("instagram_id", instagramID).
			Int("page", pagesFetched).
			Int("media_in_page", len(media)).
			Int("total_media", len(allMedia)).
			Bool("has_next", next != "").
			Msg("Fetched media page")

		nextURL = next
	}

	c.log.Info().
		Str("instagram_id", instagramID).
		Int("pages_fetched", pagesFetched).
		Int("total_media", len(allMedia)).
		Dur("elapsed", time.Since(startTime)).
		Msg("Completed fetching ALL Instagram media")

	return allMedia, nil
}

// buildMediaURL constructs the initial media API URL with required parameters.
func (c *InstagramClient) buildMediaURL(instagramID, accessToken string) (string, error) {
	apiURL, err := url.Parse(
		fmt.Sprintf("%s%s/%s/media", c.baseURL, igAPIVersion, instagramID),
	)
	if err != nil {
		c.log.Warn().Err(err).Str("instagram_id", instagramID).Msg("Failed to parse media URL")
		return "", fmt.Errorf("InstagramClient.buildMediaURL: failed to parse base URL: %w", err)
	}

	q := apiURL.Query()
	q.Set("fields", igMediaFields)
	q.Set("access_token", accessToken)
	q.Set("limit", "100")
	apiURL.RawQuery = q.Encode()

	return apiURL.String(), nil
}

// fetchMediaPage fetches a single page of media from the given URL.
// Returns the media items, the next page URL (empty if no more pages), and any error.
func (c *InstagramClient) fetchMediaPage(
	ctx context.Context,
	pageURL, accessToken string,
) ([]kafkamodels.RawInstagramMedia, string, error) {
	// Create request with app secret proof
	req, err := http.NewRequestWithContext(ctx, "GET", pageURL, nil)
	if err != nil {
		return nil, "", fmt.Errorf("InstagramClient.fetchMediaPage: failed to create request: %w", err)
	}

	rq := req.URL.Query()
	rq.Set("appsecret_proof", c.generateAppSecretProof(accessToken))
	req.URL.RawQuery = rq.Encode()

	// Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("InstagramClient.fetchMediaPage: failed HTTP request: %w", err)
	}
	defer resp.Body.Close()

	// Handle errors
	if resp.StatusCode != http.StatusOK {
		return nil, "", c.handleMediaAPIError(resp)
	}

	// Decode response
	var apiResp struct {
		Data   []kafkamodels.RawInstagramMedia `json:"data"`
		Paging struct {
			Next string `json:"next"`
		} `json:"paging"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return nil, "", fmt.Errorf("InstagramClient.fetchMediaPage: failed to decode JSON: %w", err)
	}

	return apiResp.Data, apiResp.Paging.Next, nil
}

func (c *InstagramClient) GetMediaURLs(
	ctx context.Context,
	instagramID, accessToken string,
	posts []clickhousemodels.InstagramMinimalPost,
) ([]clickhousemodels.InstagramMinimalPost, error) {
	if strings.TrimSpace(accessToken) == "" {
		return nil, fmt.Errorf("InstagramClient.GetMediaURLs: no valid access token")
	}
	if len(posts) == 0 {
		return nil, nil
	}

	const perAccountConcurrency = 5

	// Deduplicate media IDs first
	seen := make(map[string]struct{}, len(posts))
	uniquePosts := make([]clickhousemodels.InstagramMinimalPost, 0, len(posts))
	for _, post := range posts {
		if strings.TrimSpace(post.MediaID) == "" {
			continue
		}
		if _, ok := seen[post.MediaID]; ok {
			continue
		}
		seen[post.MediaID] = struct{}{}
		uniquePosts = append(uniquePosts, post)
	}

	innerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		out = make([]clickhousemodels.InstagramMinimalPost, 0, len(uniquePosts))
		sem = semaphore.NewWeighted(perAccountConcurrency)
	)
	var fetchErr error

	for _, post := range uniquePosts {
		if err := sem.Acquire(innerCtx, 1); err != nil {
			break
		}
		wg.Add(1)
		p := post
		go func() {
			defer wg.Done()
			defer sem.Release(1)

			refreshed, err := c.fetchMediaURLsForID(innerCtx, instagramID, accessToken, p.MediaID)
			if err != nil {
				if shouldSkipInstagramMediaRefreshError(err) {
					c.log.Warn().
						Str("instagram_id", instagramID).
						Str("media_id", p.MediaID).
						Err(err).
						Msg("Skipping media URL refresh for inaccessible Instagram media ID")
					// Clear URLs so ClickHouse excludes this post from future refresh runs.
					mu.Lock()
					out = append(out, clickhousemodels.InstagramMinimalPost{
						InstagramID: instagramID,
						MediaID:     p.MediaID,
						MediaURL:    []string{},
						VideoURL:    []string{},
					})
					mu.Unlock()
					return
				}
				mu.Lock()
				if fetchErr == nil {
					fetchErr = err
				}
				mu.Unlock()
				cancel() // stop remaining goroutines immediately on fatal error
				return
			}
			if len(refreshed.MediaURL) == 0 && len(refreshed.VideoURL) == 0 {
				return
			}
			mu.Lock()
			out = append(out, clickhousemodels.InstagramMinimalPost{
				InstagramID: instagramID,
				MediaID:     p.MediaID,
				MediaURL:    refreshed.MediaURL,
				VideoURL:    refreshed.VideoURL,
			})
			mu.Unlock()
		}()
	}

	wg.Wait()
	if fetchErr != nil {
		return nil, fetchErr
	}
	sort.Slice(out, func(i, j int) bool { return out[i].MediaID < out[j].MediaID })
	return out, nil
}

func shouldSkipInstagramMediaRefreshError(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())

	hasUnsupported := strings.Contains(msg, "unsupported get request")
	hasMissingPerm := strings.Contains(msg, "does not exist, cannot be loaded due to missing permissions")
	hasMethodCode := strings.Contains(msg, "graphmethodexception/100") || strings.Contains(msg, "igapiexception/100")

	return hasMethodCode && (hasUnsupported || hasMissingPerm)
}

func (c *InstagramClient) fetchMediaURLsForID(
	ctx context.Context,
	instagramID, accessToken, mediaID string,
) (*kafkamodels.ParsedInstagramPost, error) {
	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s", c.baseURL, igAPIVersion, mediaID))
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.fetchMediaURLsForID: parse URL: %w", err)
	}

	q := apiURL.Query()
	q.Set("fields", igRefreshFields)
	q.Set("access_token", accessToken)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.fetchMediaURLsForID: create request: %w", err)
	}

	if c.appSecret != "" {
		rq := req.URL.Query()
		rq.Set("appsecret_proof", c.generateAppSecretProof(accessToken))
		req.URL.RawQuery = rq.Encode()
	}

	body, status, err := c.doWithRetry(ctx, instagramID, req, "GetMediaURLs")
	if err != nil {
		return nil, err
	}
	if status != http.StatusOK {
		return nil, fmt.Errorf("InstagramClient.fetchMediaURLsForID: instagram API returned %d: %s", status, string(body))
	}

	var media kafkamodels.RawInstagramMedia
	if err := json.Unmarshal(body, &media); err != nil {
		return nil, fmt.Errorf("InstagramClient.fetchMediaURLsForID: decode JSON: %w", err)
	}

	return buildMinimalInstagramURLs(media), nil
}

func buildMinimalInstagramURLs(media kafkamodels.RawInstagramMedia) *kafkamodels.ParsedInstagramPost {
	parsed := &kafkamodels.ParsedInstagramPost{
		MediaID: media.ID,
	}

	if strings.EqualFold(media.MediaType, "VIDEO") {
		if strings.TrimSpace(media.MediaURL) != "" {
			parsed.VideoURL = append(parsed.VideoURL, media.MediaURL)
		}
		if strings.TrimSpace(media.ThumbnailURL) != "" {
			parsed.MediaURL = append(parsed.MediaURL, media.ThumbnailURL)
		}
	} else if strings.TrimSpace(media.MediaURL) != "" {
		parsed.MediaURL = append(parsed.MediaURL, media.MediaURL)
	}

	for _, child := range media.Children.Data {
		if strings.EqualFold(child.MediaType, "VIDEO") {
			if strings.TrimSpace(child.MediaURL) != "" {
				parsed.VideoURL = append(parsed.VideoURL, child.MediaURL)
			}
			if strings.TrimSpace(child.ThumbnailURL) != "" {
				parsed.MediaURL = append(parsed.MediaURL, child.ThumbnailURL)
			}
			continue
		}
		if strings.TrimSpace(child.MediaURL) != "" {
			parsed.MediaURL = append(parsed.MediaURL, child.MediaURL)
		}
	}

	return parsed
}

// handleMediaAPIError processes non-OK API responses and returns appropriate errors.
func (c *InstagramClient) handleMediaAPIError(resp *http.Response) error {
	body, _ := io.ReadAll(resp.Body)
	errStr := string(body)

	// Check for authentication errors
	if resp.StatusCode == 401 || resp.StatusCode == 403 ||
		strings.Contains(errStr, "OAuthException") ||
		strings.Contains(errStr, "access token") {
		return &InstagramAuthError{
			Message:    errStr,
			StatusCode: resp.StatusCode,
		}
	}

	return fmt.Errorf("InstagramClient.handleMediaAPIError: instagram API returned %d: %s", resp.StatusCode, errStr)
}

// FetchStories retrieves all active stories for the given Instagram account.
// Stories are ephemeral content that expires after 24 hours.
// Returns a slice of RawInstagramMedia with media_product_type = "STORY".
func (c *InstagramClient) FetchStories(ctx context.Context, instagramID, accessToken string) ([]kafkamodels.RawInstagramMedia, error) {
	startTime := time.Now()

	c.log.Info().
		Str("instagram_id", instagramID).
		Str("module", "instagram_client").
		Msg("Starting FetchStories")

	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/stories", c.baseURL, igAPIVersion, instagramID))
	if err != nil {
		c.log.Warn().Err(err).Str("instagram_id", instagramID).Msg("Failed to parse stories URL")
		return nil, fmt.Errorf("InstagramClient.FetchStories: failed to parse stories URL: %w", err)
	}

	q := apiURL.Query()
	q.Set("fields", igStoryFields)
	q.Set("access_token", accessToken)
	q.Set("limit", "100")
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.FetchStories: failed to create request: %w", err)
	}

	proof := c.generateAppSecretProof(accessToken)
	rq := req.URL.Query()
	rq.Set("appsecret_proof", proof)
	req.URL.RawQuery = rq.Encode()

	body, status, err := c.doWithRetry(ctx, instagramID, req, "FetchStories")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("InstagramClient.FetchStories: instagram stories API returned %d: %s", status, string(body))
	}

	var apiResp struct {
		Data   []kafkamodels.RawInstagramMedia `json:"data"`
		Paging struct {
			Cursors struct {
				Before string `json:"before"`
				After  string `json:"after"`
			} `json:"cursors"`
		} `json:"paging"`
	}

	if err := json.Unmarshal(body, &apiResp); err != nil {
		return nil, fmt.Errorf("InstagramClient.FetchStories: failed to decode stories response: %w", err)
	}

	elapsed := time.Since(startTime)
	c.log.Info().
		Str("instagram_id", instagramID).
		Int("stories_count", len(apiResp.Data)).
		Dur("elapsed", elapsed).
		Msg("Completed FetchStories")

	return apiResp.Data, nil
}

// FetchInsights retrieves Instagram account insights for the given date range.
// Based on Instagram API documentation:
// - Time series metrics (reach): Support metric_type=time_series, return daily values in "values" array
// - Total value metrics (likes, comments, etc.): Only support metric_type=total_value, return aggregated total
// See: https://developers.facebook.com/docs/instagram-platform/api-reference/instagram-user/insights
func (c *InstagramClient) FetchInsights(ctx context.Context, instagramID, accessToken string, since, until time.Time) (*kafkamodels.RawInstagramInsightsResponse, error) {
	startTime := time.Now()

	c.log.Info().Str("instagram_id", instagramID).
		Str("since", since.Format("2006-01-02")).
		Str("until", until.Format("2006-01-02")).
		Str("module", "instagram_client").
		Msg("Starting FetchInsights")

	combined := &kafkamodels.RawInstagramInsightsResponse{}

	// Helper function to fetch metrics
	fetch := func(metrics []string, extraParams map[string]string) error {
		apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/insights", c.baseURL, igAPIVersion, instagramID))
		if err != nil {
			return fmt.Errorf("InstagramClient.FetchInsights: failed to parse insights URL: %w", err)
		}

		q := apiURL.Query()
		q.Set("metric", strings.Join(metrics, ","))
		q.Set("period", "day")
		q.Set("since", fmt.Sprintf("%d", since.Unix()))
		q.Set("until", fmt.Sprintf("%d", until.Unix()))
		q.Set("access_token", accessToken)
		for k, v := range extraParams {
			q.Set(k, v)
		}
		proof := c.generateAppSecretProof(accessToken)
		q.Set("appsecret_proof", proof)
		apiURL.RawQuery = q.Encode()

		req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)

		if err != nil {
			return fmt.Errorf("InstagramClient.FetchInsights: failed to create request: %w", err)
		}

		body, status, err := c.doWithRetry(ctx, instagramID, req, "FetchMediaInsights")
		if err != nil {
			return err
		}

		if status != http.StatusOK {
			return fmt.Errorf("InstagramClient.FetchInsights: instagram insights API status %d: %s", status, string(body))
		}

		var r kafkamodels.RawInstagramInsightsResponse
		if err := json.Unmarshal(body, &r); err != nil {
			return fmt.Errorf("InstagramClient.FetchInsights: decode JSON: %w", err)
		}

		combined.Data = append(combined.Data, r.Data...)
		return nil
	}

	// All metrics as total_value (aggregated for the date range)
	// reach is included here - we don't need daily breakdown since other metrics are also aggregated
	totalValueMetrics := []string{"reach", "accounts_engaged", "likes", "comments", "saves", "shares", "total_interactions", "views", "replies", "profile_views"}
	if err := fetch(totalValueMetrics, map[string]string{"metric_type": "total_value"}); err != nil {
		c.log.Warn().Err(err).Msg("Failed to fetch total value metrics")
	}

	elapsed := time.Since(startTime)
	c.log.Info().Str("instagram_id", instagramID).Int("metrics", len(combined.Data)).Dur("elapsed", elapsed).Msg("Completed FetchInsights")

	return combined, nil
}

// DailyInsight holds insights data for a single day.
// Used by FetchInsightsDaily to return per-day metrics.
type DailyInsight struct {
	Date time.Time                                 // The date this insight represents
	Data *kafkamodels.RawInstagramInsightsResponse // Raw API response for this day
}

// igDailyMetrics defines the metrics fetched for daily insights.
// These are total_value metrics that get aggregated per day.
var igDailyMetrics = []string{
	"reach", "accounts_engaged", "likes", "comments", "saves",
	"shares", "total_interactions", "views", "replies", "profile_views",
}

// FetchInsightsDaily retrieves Instagram account insights for each day separately.
//
// Unlike FetchInsights which returns aggregated totals for a date range,
// this method makes one API call per day to get daily granularity for ALL metrics.
//
// Parameters:
//   - days: Number of days to fetch (starting from today, going backwards)
//   - concurrency: Max parallel API calls (recommended: 5-10)
//
// Returns a slice of DailyInsight, one for each day. Days with API errors
// will have nil Data field.
func (c *InstagramClient) FetchInsightsDaily(
	ctx context.Context,
	instagramID, accessToken string,
	days, concurrency int,
) ([]DailyInsight, error) {
	startTime := time.Now()
	c.log.Info().
		Str("instagram_id", instagramID).
		Int("days", days).
		Int("concurrency", concurrency).
		Msg("Starting FetchInsightsDaily")

	// Generate list of days to fetch (today - 2 days, going backwards)
	// Exclude today and yesterday's incomplete data, start from day before yesterday
	today := time.Now().UTC().Truncate(24 * time.Hour)
	startDay := today.AddDate(0, 0, -2) // Day before yesterday
	dates := make([]time.Time, days)
	for i := 0; i < days; i++ {
		dates[i] = startDay.AddDate(0, 0, -i)
	}

	// Fetch each day in parallel with semaphore for concurrency control
	results := make([]DailyInsight, days)
	var wg sync.WaitGroup
	var fetchErrors int32
	sem := make(chan struct{}, concurrency)

	for i, date := range dates {
		wg.Add(1)
		go func(idx int, day time.Time) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := c.fetchSingleDayInsights(ctx, instagramID, accessToken, day)
			if result == nil {
				atomic.AddInt32(&fetchErrors, 1)
			} else {
				results[idx] = *result
			}
		}(i, date)
	}
	wg.Wait()

	// Log completion stats
	elapsed := time.Since(startTime)
	successCount := days - int(atomic.LoadInt32(&fetchErrors))
	c.log.Info().
		Str("instagram_id", instagramID).
		Int("days_requested", days).
		Int("days_success", successCount).
		Int32("days_failed", atomic.LoadInt32(&fetchErrors)).
		Dur("elapsed", elapsed).
		Msg("Completed FetchInsightsDaily")

	return results, nil
}

// FetchInsightsDailyBetween retrieves Instagram account insights for each day in the requested range.
// The range is normalized to UTC day boundaries and is inclusive of both start and end dates.
func (c *InstagramClient) FetchInsightsDailyBetween(
	ctx context.Context,
	instagramID, accessToken string,
	since, until time.Time,
	concurrency int,
) ([]DailyInsight, error) {
	startTime := time.Now()

	since = time.Date(since.UTC().Year(), since.UTC().Month(), since.UTC().Day(), 0, 0, 0, 0, time.UTC)
	until = time.Date(until.UTC().Year(), until.UTC().Month(), until.UTC().Day(), 0, 0, 0, 0, time.UTC)
	if until.Before(since) {
		return nil, fmt.Errorf("InstagramClient.FetchInsightsDailyBetween: until must be on or after since")
	}
	if concurrency <= 0 {
		concurrency = 10
	}

	days := int(until.Sub(since).Hours()/24) + 1
	c.log.Info().
		Str("instagram_id", instagramID).
		Time("since", since).
		Time("until", until).
		Int("days", days).
		Int("concurrency", concurrency).
		Msg("Starting FetchInsightsDailyBetween")

	dates := make([]time.Time, days)
	for i := 0; i < days; i++ {
		dates[i] = since.AddDate(0, 0, i)
	}

	results := make([]DailyInsight, days)
	var wg sync.WaitGroup
	var fetchErrors int32
	sem := make(chan struct{}, concurrency)

	for i, date := range dates {
		wg.Add(1)
		go func(idx int, day time.Time) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			result := c.fetchSingleDayInsights(ctx, instagramID, accessToken, day)
			if result == nil {
				atomic.AddInt32(&fetchErrors, 1)
			} else {
				results[idx] = *result
			}
		}(i, date)
	}
	wg.Wait()

	elapsed := time.Since(startTime)
	successCount := days - int(atomic.LoadInt32(&fetchErrors))
	c.log.Info().
		Str("instagram_id", instagramID).
		Int("days_requested", days).
		Int("days_success", successCount).
		Int32("days_failed", atomic.LoadInt32(&fetchErrors)).
		Dur("elapsed", elapsed).
		Msg("Completed FetchInsightsDailyBetween")

	return results, nil
}

// fetchSingleDayInsights fetches insights for a single day.
// Returns nil if the API call fails.
func (c *InstagramClient) fetchSingleDayInsights(
	ctx context.Context,
	instagramID, accessToken string,
	day time.Time,
) *DailyInsight {
	// Build API URL for this day's insights
	// Use midnight to 23:59 UTC boundaries
	since := day                                    // Day at 00:00 UTC
	until := day.Add(23*time.Hour + 59*time.Minute) // Day at 23:59 UTC

	apiURL, err := url.Parse(
		fmt.Sprintf("%s%s/%s/insights", c.baseURL, igAPIVersion, instagramID),
	)
	if err != nil {
		c.log.Warn().Err(err).Time("date", day).Msg("Failed to parse insights URL")
		return nil
	}

	// Set query parameters
	q := apiURL.Query()
	q.Set("metric", strings.Join(igDailyMetrics, ","))
	q.Set("period", "day")
	q.Set("metric_type", "total_value")
	q.Set("since", fmt.Sprintf("%d", since.Unix()))
	q.Set("until", fmt.Sprintf("%d", until.Unix()))
	q.Set("access_token", accessToken)
	q.Set("appsecret_proof", c.generateAppSecretProof(accessToken))
	apiURL.RawQuery = q.Encode()

	// Execute request
	req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if err != nil {
		c.log.Warn().Err(err).Time("date", day).Msg("Failed to create request")
		return nil
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		c.log.Warn().Err(err).Time("date", day).Msg("HTTP request failed")
		return nil
	}
	defer resp.Body.Close()

	// Read response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		c.log.Warn().Err(err).Time("date", day).Msg("Failed to read response body")
		return nil
	}

	// Handle non-OK responses
	if resp.StatusCode != http.StatusOK {
		c.log.Warn().
			Int("status", resp.StatusCode).
			Time("date", day).
			Str("body", string(body)).
			Msg("Insights API error")
		return nil
	}

	// Decode response
	var r kafkamodels.RawInstagramInsightsResponse
	if err := json.Unmarshal(body, &r); err != nil {
		c.log.Warn().Err(err).Time("date", day).Msg("Failed to decode response")
		return nil
	}

	return &DailyInsight{Date: day, Data: &r}
}

// FetchMediaInsights retrieves insights for a specific media item (post, video, reel, story).
// Different metrics are available based on media type.
func (c *InstagramClient) FetchMediaInsights(ctx context.Context, mediaID, accessToken, mediaType, mediaProductType string) (*kafkamodels.RawInstagramMediaInsights, error) {
	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/insights", c.baseURL, igAPIVersion, mediaID))
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.FetchMediaInsights: failed to parse media insights URL: %w", err)
	}

	// Select metrics based on media product type (per Instagram API documentation)
	// https://developers.facebook.com/docs/instagram-platform/reference/instagram-media/insights
	var metrics []string
	switch mediaProductType {
	case "REELS":
		// REELS metrics: comments, ig_reels_avg_watch_time, ig_reels_video_view_total_time,
		// likes, reach, saved, shares, total_interactions, views
		metrics = []string{
			"comments", "likes", "reach", "saved", "shares",
			"total_interactions", "views",
			"ig_reels_avg_watch_time", "ig_reels_video_view_total_time",
		}
	case "STORY":
		// STORY metrics: reach, replies, shares, total_interactions, views
		metrics = []string{
			"reach", "replies", "shares", "total_interactions", "views",
		}
	default:
		// FEED (posts) metrics: comments, likes, reach, saved, shares, total_interactions, views
		metrics = []string{
			"comments", "likes", "reach", "saved", "shares",
			"total_interactions", "views",
		}
	}

	q := apiURL.Query()
	q.Set("metric", strings.Join(metrics, ","))
	q.Set("access_token", accessToken)
	proof := c.generateAppSecretProof(accessToken)
	q.Set("appsecret_proof", proof)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.FetchMediaInsights: failed to create request: %w", err)
	}

	body, status, err := c.doWithRetry(ctx, mediaID, req, "FetchMediaChildren")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		// Return nil for 400 errors (insights not available for this media type)
		if status == http.StatusBadRequest {
			c.log.Debug().Str("media_id", mediaID).Msg("Insights not available for this media")
			return nil, nil
		}
		return nil, fmt.Errorf("InstagramClient.FetchMediaInsights: media insights API status %d: %s", status, string(body))
	}

	var r kafkamodels.RawInstagramMediaInsights
	if err := json.Unmarshal(body, &r); err != nil {
		return nil, fmt.Errorf("InstagramClient.FetchMediaInsights: decode JSON: %w", err)
	}

	return &r, nil
}

// FetchAccountDemographics retrieves audience demographic insights for the Instagram account.
// This includes audience by city, country, age, gender, locale, and online followers.
func (c *InstagramClient) FetchAccountDemographics(ctx context.Context, instagramID, accessToken string) (*kafkamodels.RawInstagramDemographics, error) {
	combined := &kafkamodels.RawInstagramDemographics{Data: []struct {
		Name       string `json:"name"`
		Period     string `json:"period"`
		TotalValue struct {
			Value      interface{} `json:"value"`
			Breakdowns []struct {
				Results []struct {
					DimensionValues []string `json:"dimension_values"`
					Value           int      `json:"value"`
				} `json:"results"`
			} `json:"breakdowns,omitempty"`
		} `json:"total_value"`
	}{}}

	// Fetch different demographic breakdowns
	demographicMetrics := []struct {
		metrics    []string
		breakdowns []string
	}{
		{
			metrics:    []string{"follower_demographics"},
			breakdowns: []string{"city", "country", "age", "gender", "age,gender"},
		},
		{
			metrics:    []string{"engaged_audience_demographics"},
			breakdowns: []string{"city", "country", "age", "gender", "age,gender"},
		},
		{
			metrics:    []string{"reached_audience_demographics"},
			breakdowns: []string{"city", "country", "age", "gender", "age,gender"},
		},
	}

	for _, dm := range demographicMetrics {
		for _, breakdown := range dm.breakdowns {
			apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/insights", c.baseURL, igAPIVersion, instagramID))
			if err != nil {
				return nil, fmt.Errorf("InstagramClient.FetchAccountDemographics: failed to parse demographics URL: %w", err)
			}

			q := apiURL.Query()
			q.Set("metric", strings.Join(dm.metrics, ","))
			q.Set("breakdown", breakdown)
			q.Set("metric_type", "total_value")
			q.Set("period", "lifetime")
			q.Set("timeframe", "this_month")
			q.Set("access_token", accessToken)
			proof := c.generateAppSecretProof(accessToken)
			q.Set("appsecret_proof", proof)
			apiURL.RawQuery = q.Encode()

			req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
			if err != nil {
				continue
			}

			body, status, err := c.doWithRetry(ctx, instagramID, req, "FetchMediaInsights")
			if err != nil {
				continue
			}

			if status == http.StatusOK {
				var r kafkamodels.RawInstagramDemographics
				if err := json.Unmarshal(body, &r); err == nil && len(r.Data) > 0 {
					combined.Data = append(combined.Data, r.Data...)
				}
			}
		}
	}

	// Fetch online followers
	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s/insights", c.baseURL, igAPIVersion, instagramID))
	if err != nil {
		return combined, nil
	}

	q := apiURL.Query()
	q.Set("metric", "online_followers")
	q.Set("period", "lifetime")
	q.Set("until", fmt.Sprintf("%d", time.Now().UTC().Add(-48*time.Hour).Unix()))
	q.Set("access_token", accessToken)
	proof := c.generateAppSecretProof(accessToken)
	q.Set("appsecret_proof", proof)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if err == nil {
		body, status, err := c.doWithRetry(ctx, instagramID, req, "FetchInsights")
		if err == nil && status == http.StatusOK {
			var r struct {
				Data []struct {
					Name   string `json:"name"`
					Period string `json:"period"`
					Values []struct {
						Value   map[string]int `json:"value"`
						EndTime string         `json:"end_time"`
					} `json:"values"`
				} `json:"data"`
			}
			if err := json.Unmarshal(body, &r); err == nil && len(r.Data) > 0 {
				// Convert online followers to demographic format
				for _, d := range r.Data {
					demData := struct {
						Name       string `json:"name"`
						Period     string `json:"period"`
						TotalValue struct {
							Value      interface{} `json:"value"`
							Breakdowns []struct {
								Results []struct {
									DimensionValues []string `json:"dimension_values"`
									Value           int      `json:"value"`
								} `json:"results"`
							} `json:"breakdowns,omitempty"`
						} `json:"total_value"`
					}{
						Name:   d.Name,
						Period: d.Period,
					}
					if len(d.Values) > 0 {
						demData.TotalValue.Value = d.Values[0].Value
					}
					combined.Data = append(combined.Data, demData)
				}
			}
		}
	}

	c.log.Info().Str("instagram_id", instagramID).Int("demographics", len(combined.Data)).Msg("Completed FetchAccountDemographics")
	return combined, nil
}

// FetchUserInfo retrieves basic user information including followers count, follows count, etc.
func (c *InstagramClient) FetchUserInfo(ctx context.Context, instagramID, accessToken string) (map[string]interface{}, error) {
	apiURL, err := url.Parse(fmt.Sprintf("%s%s/%s", c.baseURL, igAPIVersion, instagramID))
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.FetchUserInfo: failed to parse user info URL: %w", err)
	}

	fields := []string{"followers_count", "follows_count", "media_count", "name", "profile_picture_url", "username"}

	q := apiURL.Query()
	q.Set("fields", strings.Join(fields, ","))
	q.Set("access_token", accessToken)
	proof := c.generateAppSecretProof(accessToken)
	q.Set("appsecret_proof", proof)
	apiURL.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, "GET", apiURL.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("InstagramClient.FetchUserInfo: failed to create request: %w", err)
	}

	body, status, err := c.doWithRetry(ctx, instagramID, req, "FetchInsights")
	if err != nil {
		return nil, err
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("InstagramClient.FetchUserInfo: user info API status %d: %s", status, string(body))
	}

	var result map[string]interface{}
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("InstagramClient.FetchUserInfo: decode JSON: %w", err)
	}

	return result, nil
}
