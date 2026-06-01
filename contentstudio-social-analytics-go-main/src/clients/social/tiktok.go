// Package social provides API clients for social networks.
package social

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

const (
	tiktokBaseURL  = "https://open.tiktokapis.com/v2/"
	tiktokTokenURL = "https://open.tiktokapis.com/v2/oauth/token/"
)

// TikTokClient is a lightweight client that supports fetching user info and videos.
// Only the endpoints required for analytics pipeline are implemented.
// Authentication expects a valid user access token obtained by OAuth2.

type TikTokClient struct {
	httpClient   *http.Client
	baseURL      string
	clientKey    string
	clientSecret string
	log          *logger.Logger
}

// NewTikTokClient returns a new client with sane defaults.
func NewTikTokClient(clientKey, clientSecret string) *TikTokClient {
	return &TikTokClient{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		baseURL:      tiktokBaseURL,
		clientKey:    clientKey,
		clientSecret: clientSecret,
		log:          logger.New("info"),
	}
}

// doWithRetry executes a request built by makeReq with up to 3 attempts and exponential backoff.
// makeReq is called fresh on every attempt so POST body readers are never double-consumed.
func (c *TikTokClient) doWithRetry(ctx context.Context, caller string, makeReq func() (*http.Request, error)) ([]byte, int, error) {
	const maxAttempts = 3

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		req, err := makeReq()
		if err != nil {
			return nil, 0, err
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == maxAttempts {
				return nil, 0, fmt.Errorf("TikTokClient.%s: request failed: %w", caller, err)
			}
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
			continue
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		// Never retry auth errors
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
			return body, resp.StatusCode, nil
		}
		if resp.StatusCode == http.StatusOK {
			return body, resp.StatusCode, nil
		}
		// Retry server errors
		if resp.StatusCode >= 500 && attempt < maxAttempts {
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(time.Duration(attempt) * 500 * time.Millisecond):
			}
			continue
		}

		return body, resp.StatusCode, nil
	}

	return nil, 0, fmt.Errorf("TikTokClient.%s: max retries exceeded", caller)
}

// FetchUserVideos returns raw JSON blobs for videos of the specified user.
// "cursor" and "maxCount" implement TikTok pagination; pass 0 for cursor to start.
// For simplicity we fetch only one page – caller handles pagination loop.
func (c *TikTokClient) FetchUserVideos(ctx context.Context, userID, accessToken string, cursor, maxCount int) (json.RawMessage, int64, error) {
	if maxCount <= 0 || maxCount > 20 {
		maxCount = 20 // TikTok API limit is 20
	}
	endpoint := fmt.Sprintf("%svideo/list/", c.baseURL)

	// Fields to request for video list endpoint
	fields := []string{
		"id", "create_time", "cover_image_url", "share_url", "video_description",
		"duration", "height", "width", "title", "embed_html", "embed_link",
		"like_count", "comment_count", "share_count", "view_count",
	}

	values := url.Values{}
	values.Set("fields", strings.Join(fields, ","))

	// Build request body with pagination params
	reqBody := map[string]interface{}{
		"max_count": maxCount,
	}
	if cursor > 0 {
		reqBody["cursor"] = cursor
	}

	bodyBytes, _ := json.Marshal(reqBody)
	start := time.Now()

	respBody, status, err := c.doWithRetry(ctx, "FetchUserVideos", func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint+"?"+values.Encode(), bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	})
	if err != nil {
		return nil, 0, err
	}

	var parsed struct {
		Data struct {
			Videos  json.RawMessage `json:"videos"`
			Cursor  int64           `json:"cursor"`
			HasMore bool            `json:"has_more"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &parsed); err != nil {
		return nil, 0, fmt.Errorf("TikTokClient.FetchUserVideos: failed to decode response (status %d): %w", status, err)
	}

	if parsed.Error.Code != "" && parsed.Error.Code != "ok" {
		return nil, 0, fmt.Errorf("TikTokClient.FetchUserVideos: tiktok api error (status %d): %s - %s", status, parsed.Error.Code, parsed.Error.Message)
	}

	if status != http.StatusOK {
		return nil, 0, fmt.Errorf("TikTokClient.FetchUserVideos: tiktok api-server non-200: %d", status)
	}

	elapsed := time.Since(start)
	c.log.Debug().Str("user", userID).Dur("elapsed", elapsed).Msg("FetchUserVideos page")
	nextCursor := parsed.Data.Cursor
	if !parsed.Data.HasMore {
		nextCursor = 0
	}
	return parsed.Data.Videos, nextCursor, nil
}

// FetchUserInfo fetches user profile information including follower counts.
func (c *TikTokClient) FetchUserInfo(ctx context.Context, accessToken string) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("%suser/info/", c.baseURL)

	// Fields to request
	fields := []string{
		"open_id", "union_id", "avatar_url", "avatar_url_100", "avatar_large_url",
		"display_name", "bio_description", "profile_deep_link", "is_verified",
		"follower_count", "following_count", "likes_count", "video_count",
	}

	values := url.Values{}
	values.Set("fields", strings.Join(fields, ","))

	reqURL := endpoint + "?" + values.Encode()
	respBody, status, err := c.doWithRetry(ctx, "FetchUserInfo", func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			User json.RawMessage `json:"user"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("TikTokClient.FetchUserInfo: failed to decode user info response (status %d): %w", status, err)
	}

	if result.Error.Code != "" && result.Error.Code != "ok" {
		return nil, fmt.Errorf("TikTokClient.FetchUserInfo: tiktok user info api error (status %d): %s - %s", status, result.Error.Code, result.Error.Message)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("TikTokClient.FetchUserInfo: tiktok user info api-server error: %d", status)
	}

	return result.Data.User, nil
}

// FetchVideoList fetches a list of videos with full metadata.
func (c *TikTokClient) FetchVideoList(ctx context.Context, accessToken string, cursor int64, maxCount int) (json.RawMessage, int64, bool, error) {
	endpoint := fmt.Sprintf("%svideo/list/", c.baseURL)

	// Fields to request
	fields := []string{
		"id", "create_time", "cover_image_url", "share_url", "video_description",
		"duration", "height", "width", "title", "embed_html", "embed_link",
		"like_count", "comment_count", "share_count", "view_count",
	}

	// Build request body
	reqBody := map[string]interface{}{
		"max_count": maxCount,
	}
	if cursor > 0 {
		reqBody["cursor"] = cursor
	}

	bodyBytes, _ := json.Marshal(reqBody)
	reqURL := endpoint + "?fields=" + url.QueryEscape(strings.Join(fields, ","))

	respBody, status, err := c.doWithRetry(ctx, "FetchVideoList", func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	})
	if err != nil {
		return nil, 0, false, err
	}

	var result struct {
		Data struct {
			Videos  json.RawMessage `json:"videos"`
			Cursor  int64           `json:"cursor"`
			HasMore bool            `json:"has_more"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, 0, false, fmt.Errorf("TikTokClient.FetchVideoList: failed to decode video list response (status %d): %w", status, err)
	}

	if result.Error.Code != "" && result.Error.Code != "ok" {
		return nil, 0, false, fmt.Errorf("TikTokClient.FetchVideoList: tiktok video list api error (status %d): %s - %s", status, result.Error.Code, result.Error.Message)
	}

	if status != http.StatusOK {
		return nil, 0, false, fmt.Errorf("TikTokClient.FetchVideoList: tiktok video list api-server error: %d", status)
	}

	return result.Data.Videos, result.Data.Cursor, result.Data.HasMore, nil
}

// FetchVideoDetails fetches details for specific videos using the video query endpoint.
func (c *TikTokClient) FetchVideoDetails(ctx context.Context, accessToken string, videoIDs []string) (json.RawMessage, error) {
	endpoint := fmt.Sprintf("%svideo/query/", c.baseURL)

	// Fields to request for video details
	fields := []string{"id", "title", "share_url"}

	values := url.Values{}
	values.Set("fields", strings.Join(fields, ","))

	// Build request body with video filters
	reqBody := map[string]interface{}{
		"filters": map[string]interface{}{
			"video_ids": videoIDs,
		},
	}

	bodyBytes, _ := json.Marshal(reqBody)
	reqURL := endpoint + "?" + values.Encode()

	respBody, status, err := c.doWithRetry(ctx, "FetchVideoDetails", func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")
		return req, nil
	})
	if err != nil {
		return nil, err
	}

	var result struct {
		Data struct {
			Videos json.RawMessage `json:"videos"`
		} `json:"data"`
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("TikTokClient.FetchVideoDetails: failed to decode video query response (status %d): %w", status, err)
	}

	if result.Error.Code != "" && result.Error.Code != "ok" {
		return nil, fmt.Errorf("TikTokClient.FetchVideoDetails: tiktok video query api error (status %d): %s - %s", status, result.Error.Code, result.Error.Message)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("TikTokClient.FetchVideoDetails: tiktok video query api-server error: %d", status)
	}

	return result.Data.Videos, nil
}

// RefreshToken refreshes an expired access token using the refresh token.
// This follows the OAuth2 refresh_token flow per TikTok API documentation.
// Returns a RefreshTokenResponse containing the new access token, refresh token,
// and expiration information (expires_in and refresh_expires_in).
func (c *TikTokClient) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	if c.clientKey == "" || c.clientSecret == "" {
		return nil, fmt.Errorf("TikTokClient.RefreshToken: client credentials not configured")
	}

	formValues := url.Values{}
	formValues.Set("client_key", c.clientKey)
	formValues.Set("client_secret", c.clientSecret)
	formValues.Set("grant_type", "refresh_token")
	formValues.Set("refresh_token", refreshToken)
	formBody := []byte(formValues.Encode())

	start := time.Now()
	respBody, status, err := c.doWithRetry(ctx, "RefreshToken", func() (*http.Request, error) {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, tiktokTokenURL, bytes.NewReader(formBody))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		return req, nil
	})
	if err != nil {
		return nil, err
	}

	elapsed := time.Since(start)
	c.log.Debug().Dur("elapsed", elapsed).Msg("RefreshToken API call completed")

	if status != http.StatusOK {
		return nil, fmt.Errorf("TikTokClient.RefreshToken: token refresh failed (status %d): %s", status, string(respBody))
	}

	var result struct {
		AccessToken      string `json:"access_token"`
		RefreshToken     string `json:"refresh_token"`
		ExpiresIn        int    `json:"expires_in"`
		RefreshExpiresIn int    `json:"refresh_expires_in"`
		Scope            string `json:"scope"`
	}

	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, err
	}

	if result.AccessToken == "" {
		return nil, fmt.Errorf("TikTokClient.RefreshToken: no access token in response")
	}

	// Return the response with expires_in values as integers (from TikTok API)
	// The caller will handle converting to object format if needed for MongoDB
	return &RefreshTokenResponse{
		AccessToken:      result.AccessToken,
		RefreshToken:     result.RefreshToken,
		ExpiresIn:        result.ExpiresIn,
		RefreshExpiresIn: result.RefreshExpiresIn,
		Scope:            result.Scope,
	}, nil
}

// FetchVideoPaginated fetches all videos for a user with pagination support.
// cutoffTime determines when to stop fetching (for incremental sync).
// For immediate requests: last 90 days, max 999 videos.
// For cron jobs: last 14 days, max 999 videos.
func (c *TikTokClient) FetchVideoPaginated(ctx context.Context, accessToken string, cutoffTime time.Time, maxVideos int) ([]json.RawMessage, error) {
	collected := make([]json.RawMessage, 0, maxVideos)
	var cursor int64 = 0
	const videosPerPage = 20
	const maxPages = 200
	pagesFetched := 0

	for len(collected) < maxVideos {
		if pagesFetched >= maxPages {
			c.log.Warn().Int("pages_fetched", pagesFetched).Int("videos_collected", len(collected)).Msg("TikTokClient.FetchVideoPaginated: reached max pages limit; returning partial results")
			break
		}
		remaining := maxVideos - len(collected)
		if remaining < videosPerPage {
			remaining = videosPerPage
		}

		// Fetch a page of videos
		data, nextCursor, hasMore, err := c.FetchVideoList(ctx, accessToken, cursor, remaining)
		if err != nil {
			return collected, err
		}

		// Parse videos from the response
		var videos []json.RawMessage
		if err := json.Unmarshal(data, &videos); err != nil {
			c.log.Warn().Err(err).Msg("Failed to unmarshal videos array")
			// Try to continue with empty videos
			videos = []json.RawMessage{}
		}

		// Check cutoff time for each video
		reachedCutoff := false
		for _, videoRaw := range videos {
			var video struct {
				CreateTime int64 `json:"create_time"`
			}
			if err := json.Unmarshal(videoRaw, &video); err == nil && !cutoffTime.IsZero() {
				videoTime := time.Unix(video.CreateTime, 0)
				if videoTime.Before(cutoffTime) {
					reachedCutoff = true
					break
				}
			}
			collected = append(collected, videoRaw)
			if len(collected) >= maxVideos {
				break
			}
		}

		if reachedCutoff || !hasMore || nextCursor == 0 {
			break
		}

		pagesFetched++
		cursor = nextCursor

		// Respect rate limits
		select {
		case <-ctx.Done():
			return collected, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}

	return collected, nil
}

// IsExpectedCompetitorErrorTikTok checks if an error is an expected/known API error
// that should be logged as Warn (not captured to Sentry).
// Expected errors include invalid/expired tokens and permission issues.
func IsExpectedCompetitorErrorTikTok(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Token-related errors - invalid, expired, or not found
	if strings.Contains(errStr, "access_token_invalid") ||
		strings.Contains(errStr, "invalid or not found") ||
		strings.Contains(errStr, "token") && strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "token") && strings.Contains(errStr, "expired") ||
		strings.Contains(errStr, "token") && strings.Contains(errStr, "not found") {
		return true
	}

	// Permission/scope errors
	if strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "unauthorized") {
		return true
	}

	// Account-related errors
	if strings.Contains(errStr, "account") && strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "account") && strings.Contains(errStr, "not found") {
		return true
	}

	// API errors with 401/403 status codes
	if strings.Contains(errStr, "status 401") ||
		strings.Contains(errStr, "status 403") ||
		strings.Contains(errStr, "401") && strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "403") && strings.Contains(errStr, "forbidden") {
		return true
	}

	return false
}
