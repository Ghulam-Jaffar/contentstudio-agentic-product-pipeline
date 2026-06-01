// Package social provides API clients for social networks.
package social

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

const (
	gmbTokenURL              = "https://oauth2.googleapis.com/token"
	gmbVoiceOfMerchantURL    = "https://mybusinessverifications.googleapis.com/v1/locations/"
	gmbPerformanceMetricsURL = "https://businessprofileperformance.googleapis.com/v1/locations/"
	gmbMyBusinessURL         = "https://mybusiness.googleapis.com/v4/"

	// GMB retry/backoff constants (aligned with Facebook/Instagram patterns).
	gmbMaxAttempts = 5
	gmbBaseBackoff = 500 * time.Millisecond
	gmbMaxBackoff  = 10 * time.Second
)

// GMBClient is a client for Google My Business / Google Business Profile APIs.
// It uses the shared RateManager for global + per-token rate limiting, and
// exponential backoff with jitter on retries (same pattern as FacebookClient).
//
// Google Business Profile API quotas (per GCP project, each API has its own quota):
//   - Business Profile Performance API:      300 QPM (fetchMultiDailyMetrics, searchKeywords)
//   - My Business Business Information API:   300 QPM (locations, localPosts, media)
//   - My Business Account Management API:     300 QPM (accounts)
//   - My Business Verifications API:          300 QPM (VoiceOfMerchant)
//   - My Business v4 legacy (reviews, media): 300 QPM
//   - OAuth2 token endpoint:                  separate quota (not GBP-specific)
//
// The quota is per-API, NOT shared across APIs. However, endpoints within the same
// API share a single 300 QPM bucket. With 10 workers each making ~8-12 calls,
// we use a conservative 4 RPS global limit (240/min) to stay safely under quota.
//
// Rate limiting is applied to ALL 7 endpoints:
//   - RefreshToken           → waitRate() in its own retry loop
//   - FetchVoiceOfMerchant   → doJSONGet() → waitRate()
//   - FetchPerformanceMetrics→ doJSONGet() → waitRate()
//   - FetchSearchKeywords    → doJSONGet() → waitRate()
//   - FetchLocalPosts        → doJSONGet() → waitRate()
//   - FetchReviews           → doJSONGet() → waitRate()
//   - FetchMediaAssets       → doJSONGet() → waitRate()
type GMBClient struct {
	httpClient   *http.Client
	clientID     string
	clientSecret string
	log          *logger.Logger
	rate         *RateManager
}

// NewGMBClient returns a new client with default rate limits (backward compatible).
// Default: 4 RPS global (240/min, safely under 300/min project quota), 2 RPS per token.
func NewGMBClient(clientID, clientSecret string) *GMBClient {
	return NewGMBClientWithRates(clientID, clientSecret, NewRateManager(RateLimits{
		GlobalRPS:     4.0,
		GlobalBurst:   5,
		PerTokenRPS:   2.0,
		PerTokenBurst: 3,
	}))
}

// NewGMBClientWithRates creates a client wired to a shared RateManager.
// Use this when you want to configure rate limits from service config.
func NewGMBClientWithRates(clientID, clientSecret string, rm *RateManager) *GMBClient {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		TLSHandshakeTimeout:   8 * time.Second,
		ResponseHeaderTimeout: 40 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		MaxIdleConns:          50,
		MaxIdleConnsPerHost:   20,
		IdleConnTimeout:       90 * time.Second,
		ForceAttemptHTTP2:     true,
	}
	if rm == nil {
		rm = NewRateManager(RateLimits{
			GlobalRPS:     4.0,
			GlobalBurst:   5,
			PerTokenRPS:   2.0,
			PerTokenBurst: 3,
		})
	}
	return &GMBClient{
		httpClient: &http.Client{
			Transport: tr,
			Timeout:   30 * time.Second,
		},
		clientID:     clientID,
		clientSecret: clientSecret,
		log:          logger.New("info"),
		rate:         rm,
	}
}

// waitRate applies global + per-token throttling before each GMB API call.
func (c *GMBClient) waitRate(ctx context.Context, token string) error {
	if c.rate == nil {
		c.rate = NewRateManager(RateLimits{
			GlobalRPS:     4.0,
			GlobalBurst:   5,
			PerTokenRPS:   2.0,
			PerTokenBurst: 3,
		})
	}
	return c.rate.Wait(ctx, token)
}

// gmbComputeBackoff calculates exponential backoff with ±25% jitter.
func gmbComputeBackoff(attempt int) time.Duration {
	d := gmbBaseBackoff << (attempt - 1)
	if d > gmbMaxBackoff {
		d = gmbMaxBackoff
	}
	j := time.Duration(int64(d) / 4) // ±25% jitter window
	return d - j/2 + time.Duration(rand.Int63n(int64(j)+1))
}

// doJSONGet executes a GET request with Bearer auth, decodes JSON into result,
// using rate limiting, exponential backoff with jitter, and Retry-After header
// support (same pattern as FacebookClient.doWithRetry).
func (c *GMBClient) doJSONGet(ctx context.Context, fullURL, accessToken, methodName string, result interface{}) error {
	for attempt := 1; attempt <= gmbMaxAttempts; attempt++ {
		// Rate-limit wait before each attempt
		if err := c.waitRate(ctx, accessToken); err != nil {
			c.log.Warn().Err(err).Str("method", methodName).Msg("Rate limit wait failed")
			return fmt.Errorf("%s: rate limit wait failed: %w", methodName, err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, fullURL, nil)
		if err != nil {
			return fmt.Errorf("%s: failed to create request: %w", methodName, err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		req.Header.Set("Content-Type", "application/json")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == gmbMaxAttempts {
				c.log.Warn().Err(err).Str("method", methodName).Int("attempt", attempt).Msg("HTTP request failed; max attempts reached")
				return fmt.Errorf("%s: request failed: %w", methodName, err)
			}
			delay := gmbComputeBackoff(attempt)
			c.log.Warn().Err(err).Str("method", methodName).Int("attempt", attempt).Dur("backoff", delay).Msg("HTTP request failed; retrying with backoff")
			time.Sleep(delay)
			continue
		}

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			if err := json.Unmarshal(body, result); err != nil {
				return fmt.Errorf("%s: failed to decode response: %w", methodName, err)
			}
			return nil
		}

		// Non-200: build error message
		currentErr := fmt.Errorf("%s: API error (status %d): %s", methodName, resp.StatusCode, truncateBody(body, 512))

		// For expected auth/permission errors, don't retry
		if IsExpectedCompetitorErrorGMB(currentErr) && attempt == 1 {
			c.log.Warn().Int("status", resp.StatusCode).Str("method", methodName).Msg("Expected client error (auth/permission); not retrying")
			return currentErr
		}

		// Honor Retry-After header from Google
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if s, _ := strconv.Atoi(strings.TrimSpace(ra)); s > 0 {
				c.log.Warn().Str("method", methodName).Int("attempt", attempt).Int("retry_after_sec", s).Int("status", resp.StatusCode).Msg("Respecting Retry-After header")
				time.Sleep(time.Duration(s) * time.Second)
				continue
			}
		}

		if attempt == gmbMaxAttempts {
			c.log.Warn().Int("status", resp.StatusCode).Str("method", methodName).Int("attempt", attempt).Msg("Giving up after max attempts")
			return currentErr
		}

		delay := gmbComputeBackoff(attempt)
		c.log.Warn().Int("status", resp.StatusCode).Str("method", methodName).Int("attempt", attempt).Dur("backoff", delay).Msg("Non-200 response; retrying with backoff")
		time.Sleep(delay)
	}

	return fmt.Errorf("%s: unreachable", methodName)
}

// truncateBody truncates a response body for logging.
func truncateBody(body []byte, maxLen int) string {
	if len(body) <= maxLen {
		return string(body)
	}
	return string(body[:maxLen])
}

// RefreshToken refreshes an expired access token using the refresh token.
// Uses the Google OAuth2 token endpoint with client_id, client_secret, and refresh_token.
// Includes rate limiting and exponential backoff. Errors are logged as warnings.
func (c *GMBClient) RefreshToken(ctx context.Context, refreshToken string) (*RefreshTokenResponse, error) {
	if c.clientID == "" || c.clientSecret == "" {
		return nil, fmt.Errorf("GMBClient.RefreshToken: client credentials not configured")
	}

	for attempt := 1; attempt <= gmbMaxAttempts; attempt++ {
		// Rate-limit wait (token refresh uses a different Google endpoint but still counts)
		if err := c.waitRate(ctx, refreshToken); err != nil {
			c.log.Warn().Err(err).Msg("GMBClient.RefreshToken: rate limit wait failed")
			return nil, fmt.Errorf("GMBClient.RefreshToken: rate limit wait failed: %w", err)
		}

		values := url.Values{}
		values.Set("client_id", c.clientID)
		values.Set("client_secret", c.clientSecret)
		values.Set("refresh_token", refreshToken)
		values.Set("grant_type", "refresh_token")

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, gmbTokenURL,
			strings.NewReader(values.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == gmbMaxAttempts {
				c.log.Warn().Err(err).Int("attempt", attempt).Msg("GMBClient.RefreshToken: HTTP request failed; max attempts reached")
				return nil, fmt.Errorf("GMBClient.RefreshToken: request failed: %w", err)
			}
			delay := gmbComputeBackoff(attempt)
			c.log.Warn().Err(err).Int("attempt", attempt).Dur("backoff", delay).Msg("GMBClient.RefreshToken: HTTP request failed; retrying with backoff")
			time.Sleep(delay)
			continue
		}

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 4096))
		resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			var result struct {
				AccessToken  string `json:"access_token"`
				ExpiresIn    int    `json:"expires_in"`
				Scope        string `json:"scope"`
				TokenType    string `json:"token_type"`
				RefreshToken string `json:"refresh_token"`
			}
			if err := json.Unmarshal(body, &result); err != nil {
				return nil, err
			}
			if result.AccessToken == "" {
				return nil, fmt.Errorf("GMBClient.RefreshToken: no access token in response")
			}
			return &RefreshTokenResponse{
				AccessToken:  result.AccessToken,
				RefreshToken: result.RefreshToken,
				ExpiresIn:    result.ExpiresIn,
				Scope:        result.Scope,
			}, nil
		}

		currentErr := fmt.Errorf("GMBClient.RefreshToken: token refresh failed (status %d): %s", resp.StatusCode, truncateBody(body, 512))

		// For expected auth errors, don't retry
		if IsExpectedCompetitorErrorGMB(currentErr) && attempt == 1 {
			c.log.Warn().Int("status", resp.StatusCode).Msg("GMBClient.RefreshToken: expected auth error; not retrying")
			return nil, currentErr
		}

		// Honor Retry-After header
		if ra := resp.Header.Get("Retry-After"); ra != "" {
			if s, _ := strconv.Atoi(strings.TrimSpace(ra)); s > 0 {
				c.log.Warn().Int("attempt", attempt).Int("retry_after_sec", s).Int("status", resp.StatusCode).Msg("GMBClient.RefreshToken: respecting Retry-After header")
				time.Sleep(time.Duration(s) * time.Second)
				continue
			}
		}

		if attempt == gmbMaxAttempts {
			c.log.Warn().Int("status", resp.StatusCode).Int("attempt", attempt).Msg("GMBClient.RefreshToken: giving up after max attempts")
			return nil, currentErr
		}

		delay := gmbComputeBackoff(attempt)
		c.log.Warn().Int("status", resp.StatusCode).Int("attempt", attempt).Dur("backoff", delay).Msg("GMBClient.RefreshToken: retrying with backoff")
		time.Sleep(delay)
	}

	return nil, fmt.Errorf("GMBClient.RefreshToken: unreachable")
}

// VoiceOfMerchantResponse represents the response from the Voice of Merchant API.
type VoiceOfMerchantResponse struct {
	HasVoiceOfMerchant   bool `json:"hasVoiceOfMerchant"`
	HasBusinessAuthority bool `json:"hasBusinessAuthority"`
}

// FetchVoiceOfMerchant checks whether the account has access to Business Performance APIs.
func (c *GMBClient) FetchVoiceOfMerchant(ctx context.Context, locationID, accessToken string) (*VoiceOfMerchantResponse, error) {
	fullURL := gmbVoiceOfMerchantURL + locationID + "/VoiceOfMerchantState"
	var result VoiceOfMerchantResponse
	if err := c.doJSONGet(ctx, fullURL, accessToken, "GMBClient.FetchVoiceOfMerchant", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GMBPerformanceResponse represents the response from the daily performance metrics API.
type GMBPerformanceResponse struct {
	MultiDailyMetricTimeSeries []struct {
		DailyMetricTimeSeries []struct {
			DailyMetric string `json:"dailyMetric"`
			TimeSeries  struct {
				DatedValues []struct {
					Date struct {
						Year  int `json:"year"`
						Month int `json:"month"`
						Day   int `json:"day"`
					} `json:"date"`
					Value string `json:"value"`
				} `json:"datedValues"`
			} `json:"timeSeries"`
		} `json:"dailyMetricTimeSeries"`
	} `json:"multiDailyMetricTimeSeries"`
}

// FetchPerformanceMetrics fetches daily performance metrics for a location.
// The startDate and endDate define the date range for the query.
func (c *GMBClient) FetchPerformanceMetrics(ctx context.Context, locationID, accessToken string, startDate, endDate time.Time) (*GMBPerformanceResponse, error) {
	endpoint := gmbPerformanceMetricsURL + locationID + ":fetchMultiDailyMetricsTimeSeries"

	metrics := []string{
		"BUSINESS_IMPRESSIONS_DESKTOP_MAPS",
		"BUSINESS_IMPRESSIONS_DESKTOP_SEARCH",
		"BUSINESS_IMPRESSIONS_MOBILE_MAPS",
		"BUSINESS_IMPRESSIONS_MOBILE_SEARCH",
		"CALL_CLICKS",
		"WEBSITE_CLICKS",
		"BUSINESS_DIRECTION_REQUESTS",
		"BUSINESS_CONVERSATIONS",
		"BUSINESS_BOOKINGS",
		"BUSINESS_FOOD_ORDERS",
		"BUSINESS_FOOD_MENU_CLICKS",
	}

	params := url.Values{}
	for _, m := range metrics {
		params.Add("dailyMetrics", m)
	}
	params.Set("dailyRange.startDate.year", fmt.Sprintf("%d", startDate.Year()))
	params.Set("dailyRange.startDate.month", fmt.Sprintf("%d", int(startDate.Month())))
	params.Set("dailyRange.startDate.day", fmt.Sprintf("%d", startDate.Day()))
	params.Set("dailyRange.endDate.year", fmt.Sprintf("%d", endDate.Year()))
	params.Set("dailyRange.endDate.month", fmt.Sprintf("%d", int(endDate.Month())))
	params.Set("dailyRange.endDate.day", fmt.Sprintf("%d", endDate.Day()))

	fullURL := endpoint + "?" + params.Encode()
	var result GMBPerformanceResponse
	if err := c.doJSONGet(ctx, fullURL, accessToken, "GMBClient.FetchPerformanceMetrics", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GMBSearchKeywordsResponse represents the response from the monthly search keywords API.
type GMBSearchKeywordsResponse struct {
	SearchKeywordsCounts []struct {
		SearchKeyword string `json:"searchKeyword"`
		InsightsValue struct {
			Value     string `json:"value"`
			Threshold string `json:"threshold"`
		} `json:"insightsValue"`
	} `json:"searchKeywordsCounts"`
}

// FetchSearchKeywords fetches monthly search keywords that triggered business impressions.
func (c *GMBClient) FetchSearchKeywords(ctx context.Context, locationID, accessToken string, startMonth, endMonth time.Time) (*GMBSearchKeywordsResponse, error) {
	endpoint := gmbPerformanceMetricsURL + locationID + "/searchkeywords/impressions/monthly"

	params := url.Values{}
	params.Set("monthlyRange.startMonth.year", fmt.Sprintf("%d", startMonth.Year()))
	params.Set("monthlyRange.startMonth.month", fmt.Sprintf("%d", int(startMonth.Month())))
	params.Set("monthlyRange.endMonth.year", fmt.Sprintf("%d", endMonth.Year()))
	params.Set("monthlyRange.endMonth.month", fmt.Sprintf("%d", int(endMonth.Month())))

	fullURL := endpoint + "?" + params.Encode()
	var result GMBSearchKeywordsResponse
	if err := c.doJSONGet(ctx, fullURL, accessToken, "GMBClient.FetchSearchKeywords", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GMBLocalPostsResponse represents the response from the local posts API.
type GMBLocalPostsResponse struct {
	LocalPosts []struct {
		Name         string `json:"name"`
		LanguageCode string `json:"languageCode"`
		Summary      string `json:"summary"`
		State        string `json:"state"`
		UpdateTime   string `json:"updateTime"`
		CreateTime   string `json:"createTime"`
		SearchURL    string `json:"searchUrl"`
		Media        []struct {
			Name        string `json:"name"`
			MediaFormat string `json:"mediaFormat"`
			GoogleURL   string `json:"googleUrl"`
		} `json:"media"`
		TopicType string `json:"topicType"`
	} `json:"localPosts"`
	NextPageToken string `json:"nextPageToken"`
}

// FetchLocalPosts fetches all local posts for a business location.
// accountID is the GMB account ID, locationID is the GMB location ID.
func (c *GMBClient) FetchLocalPosts(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBLocalPostsResponse, error) {
	endpoint := fmt.Sprintf("%saccounts/%s/locations/%s/localPosts", gmbMyBusinessURL, accountID, locationID)

	params := url.Values{}
	params.Set("pageSize", "100")
	if pageToken != "" {
		params.Set("pageToken", pageToken)
	}

	fullURL := endpoint + "?" + params.Encode()
	var result GMBLocalPostsResponse
	if err := c.doJSONGet(ctx, fullURL, accessToken, "GMBClient.FetchLocalPosts", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GMBReviewsResponse represents the response from the reviews API.
type GMBReviewsResponse struct {
	Reviews []struct {
		ReviewID string `json:"reviewId"`
		Reviewer struct {
			ProfilePhotoURL string `json:"profilePhotoUrl"`
			DisplayName     string `json:"displayName"`
		} `json:"reviewer"`
		StarRating  string `json:"starRating"`
		Comment     string `json:"comment"`
		CreateTime  string `json:"createTime"`
		UpdateTime  string `json:"updateTime"`
		Name        string `json:"name"`
		ReviewReply *struct {
			Comment    string `json:"comment"`
			UpdateTime string `json:"updateTime"`
		} `json:"reviewReply,omitempty"`
	} `json:"reviews"`
	NextPageToken    string `json:"nextPageToken"`
	TotalReviewCount int    `json:"totalReviewCount"`
}

// FetchReviews fetches customer reviews for a business location.
// Supports pagination via pageToken. Pass empty string for first page.
func (c *GMBClient) FetchReviews(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBReviewsResponse, error) {
	endpoint := fmt.Sprintf("%saccounts/%s/locations/%s/reviews", gmbMyBusinessURL, accountID, locationID)

	params := url.Values{}
	params.Set("orderBy", "updateTime desc")
	params.Set("pageSize", "50")
	if pageToken != "" {
		params.Set("pageToken", pageToken)
	}

	fullURL := endpoint + "?" + params.Encode()
	var result GMBReviewsResponse
	if err := c.doJSONGet(ctx, fullURL, accessToken, "GMBClient.FetchReviews", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GMBMediaAssetsResponse represents the response from the media assets API.
type GMBMediaAssetsResponse struct {
	MediaItems []struct {
		Name                string `json:"name"`
		MediaFormat         string `json:"mediaFormat"`
		LocationAssociation struct {
			Category string `json:"category"`
		} `json:"locationAssociation"`
		GoogleURL    string `json:"googleUrl"`
		ThumbnailURL string `json:"thumbnailUrl"`
		CreateTime   string `json:"createTime"`
		Dimensions   struct {
			WidthPixels  int `json:"widthPixels"`
			HeightPixels int `json:"heightPixels"`
		} `json:"dimensions"`
		SourceURL string `json:"sourceUrl"`
	} `json:"mediaItems"`
	NextPageToken       string `json:"nextPageToken"`
	TotalMediaItemCount int    `json:"totalMediaItemCount"`
}

// FetchMediaAssets fetches media assets for a business location.
func (c *GMBClient) FetchMediaAssets(ctx context.Context, accountID, locationID, accessToken, pageToken string) (*GMBMediaAssetsResponse, error) {
	fullURL := fmt.Sprintf("%saccounts/%s/locations/%s/media", gmbMyBusinessURL, accountID, locationID)
	if pageToken != "" {
		fullURL += "?pageToken=" + url.QueryEscape(pageToken)
	}
	var result GMBMediaAssetsResponse
	if err := c.doJSONGet(ctx, fullURL, accessToken, "GMBClient.FetchMediaAssets", &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// IsExpectedCompetitorErrorGMB checks if an error is an expected/known API error
// that should be logged as Warn (not captured to Sentry).
// Expected errors include invalid/expired tokens and permission issues.
func IsExpectedCompetitorErrorGMB(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Token-related errors
	if strings.Contains(errStr, "invalid_grant") ||
		strings.Contains(errStr, "invalid_client") ||
		strings.Contains(errStr, "token") && strings.Contains(errStr, "invalid") ||
		strings.Contains(errStr, "token") && strings.Contains(errStr, "expired") ||
		strings.Contains(errStr, "token") && strings.Contains(errStr, "revoked") {
		return true
	}

	// Permission/scope errors
	if strings.Contains(errStr, "permission") ||
		strings.Contains(errStr, "forbidden") ||
		strings.Contains(errStr, "unauthorized") ||
		strings.Contains(errStr, "PERMISSION_DENIED") {
		return true
	}

	// Account/location not found errors
	if strings.Contains(errStr, "NOT_FOUND") ||
		strings.Contains(errStr, "not found") {
		return true
	}

	// API errors with 401/403 status codes
	if strings.Contains(errStr, "status 401") ||
		strings.Contains(errStr, "status 403") ||
		strings.Contains(errStr, "status 404") {
		return true
	}

	return false
}
