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
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	metaadsmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	metaAdsAPIVersion = "v21.0"
	metaAdsBaseURL    = "https://graph.facebook.com/"

	// metaAdsMaxPages is a safety cap on cursor-based pagination loops.
	metaAdsMaxPages = 100

	// Rate limiting: Meta Marketing API allows 200 calls per hour per token (conservative side).
	// We default to 2 RPS per token and 6 RPS globally.
	metaAdsDefaultPerTokenRPS   = 2.0
	metaAdsDefaultPerTokenBurst = 4
	metaAdsDefaultGlobalRPS     = 6.0
	metaAdsDefaultGlobalBurst   = 10

	// Insight fields returned by all three insight levels.
	insightCommonFields = "spend,impressions,reach,clicks,unique_clicks,ctr,unique_ctr,cpc,cpm,cpp,frequency,date_start,date_stop,actions"

	// Action types we want from the insights endpoints.
	insightActionFilter = `[{"field":"action_type","operator":"IN","value":["purchase","offsite_conversion.fb_pixel_purchase","link_click","lead","offsite_conversion.fb_pixel_lead","post_engagement","mobile_app_install"]}]`
)

// ─────────────────────────────────────────────────────────────────────────────
// Rate limiter for Meta Ads
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsRateManager manages per-token and global rate limiting for the Meta
// Ads API. The zero value is not usable; use NewMetaAdsRateManager.
type MetaAdsRateManager struct {
	global   *rate.Limiter
	perToken map[string]*rate.Limiter
	limits   MetaAdsRateLimits
	mu       sync.Mutex
}

type MetaAdsRateLimits struct {
	PerTokenRPS   float64
	PerTokenBurst int
	GlobalRPS     float64
	GlobalBurst   int
}

// NewMetaAdsRateManager creates a new rate manager with the provided limits.
// Zero values fall back to safe defaults.
func NewMetaAdsRateManager(lims MetaAdsRateLimits) *MetaAdsRateManager {
	if lims.PerTokenRPS <= 0 {
		lims.PerTokenRPS = metaAdsDefaultPerTokenRPS
	}
	if lims.PerTokenBurst <= 0 {
		lims.PerTokenBurst = metaAdsDefaultPerTokenBurst
	}
	if lims.GlobalRPS <= 0 {
		lims.GlobalRPS = metaAdsDefaultGlobalRPS
	}
	if lims.GlobalBurst <= 0 {
		lims.GlobalBurst = metaAdsDefaultGlobalBurst
	}
	return &MetaAdsRateManager{
		global:   rate.NewLimiter(rate.Limit(lims.GlobalRPS), lims.GlobalBurst),
		perToken: make(map[string]*rate.Limiter),
		limits:   lims,
	}
}

func (rm *MetaAdsRateManager) tokenLimiter(token string) *rate.Limiter {
	rm.mu.Lock()
	defer rm.mu.Unlock()
	if lim, ok := rm.perToken[token]; ok {
		return lim
	}
	lim := rate.NewLimiter(rate.Limit(rm.limits.PerTokenRPS), rm.limits.PerTokenBurst)
	rm.perToken[token] = lim
	return lim
}

// Wait blocks until both the global and per-token rate limiters allow a request.
func (rm *MetaAdsRateManager) Wait(ctx context.Context, token string) error {
	if err := rm.global.Wait(ctx); err != nil {
		return err
	}
	return rm.tokenLimiter(token).Wait(ctx)
}

// ─────────────────────────────────────────────────────────────────────────────
// Client
// ─────────────────────────────────────────────────────────────────────────────

// MetaAdsClient is an HTTP client for the Meta Ads (Marketing) Graph API.
type MetaAdsClient struct {
	httpClient *http.Client
	appSecret  string
	rate       *MetaAdsRateManager
	log        *logger.Logger
}

// NewMetaAdsClient creates a client with default rate limits.
func NewMetaAdsClient(appSecret string, log *logger.Logger) *MetaAdsClient {
	return NewMetaAdsClientWithRates(appSecret, NewMetaAdsRateManager(MetaAdsRateLimits{}), log)
}

// NewMetaAdsClientWithRates creates a client with a custom rate manager.
func NewMetaAdsClientWithRates(appSecret string, rm *MetaAdsRateManager, log *logger.Logger) *MetaAdsClient {
	tr := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}
	return &MetaAdsClient{
		httpClient: &http.Client{Timeout: 30 * time.Second, Transport: tr},
		appSecret:  appSecret,
		rate:       rm,
		log:        log,
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Low-level helpers
// ─────────────────────────────────────────────────────────────────────────────

// appsecretProof computes the appsecret_proof HMAC-SHA256 signature.
func (c *MetaAdsClient) appsecretProof(accessToken string) string {
	h := hmac.New(sha256.New, []byte(c.appSecret))
	h.Write([]byte(accessToken))
	return hex.EncodeToString(h.Sum(nil))
}

// get performs a rate-limited GET and decodes the JSON response into v.
// Retry policy:
//   - HTTP 429 (Too Many Requests): up to 3 retries with exponential backoff (~5s, ~10s, ~30s).
//   - Network error or other non-200 status: retry once immediately.
func (c *MetaAdsClient) get(ctx context.Context, accessToken, rawURL string, v interface{}) error {
	const maxRetries429 = 3
	backoffs429 := [maxRetries429]time.Duration{5 * time.Second, 10 * time.Second, 30 * time.Second}

	var retries429, retriesOther int

	for {
		if err := c.rate.Wait(ctx, accessToken); err != nil {
			return fmt.Errorf("MetaAdsClient.get: rate limit wait: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return fmt.Errorf("MetaAdsClient.get: new request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if retriesOther < 1 {
				retriesOther++
				c.log.Warn().Err(err).Msg("Meta Ads API request failed, retrying once")
				continue
			}
			return fmt.Errorf("MetaAdsClient.get: do request: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("MetaAdsClient.get: read body: %w", readErr)
		}

		switch {
		case resp.StatusCode == http.StatusTooManyRequests:
			if retries429 < maxRetries429 {
				backoff := backoffs429[retries429]
				retries429++
				c.log.Warn().
					Int("retry", retries429).
					Dur("backoff", backoff).
					Msg("Meta Ads API rate limited (429), backing off before retry")
				select {
				case <-time.After(backoff):
				case <-ctx.Done():
					return ctx.Err()
				}
				continue
			}
			return fmt.Errorf("MetaAdsClient.get: too many requests (429), all %d retries exhausted", maxRetries429)

		case resp.StatusCode != http.StatusOK:
			var fbErr struct {
				Error struct {
					Message string `json:"message"`
					Type    string `json:"type"`
					Code    int    `json:"code"`
				} `json:"error"`
			}
			_ = json.Unmarshal(body, &fbErr)
			apiErr := fmt.Errorf("MetaAdsClient.get: status %d: %s (code %d)",
				resp.StatusCode, fbErr.Error.Message, fbErr.Error.Code)
			if retriesOther < 1 {
				retriesOther++
				c.log.Warn().Err(apiErr).Msg("Meta Ads API returned non-200, retrying once")
				continue
			}
			return apiErr

		default:
			if err := json.Unmarshal(body, v); err != nil {
				return fmt.Errorf("MetaAdsClient.get: unmarshal: %w", err)
			}
			return nil
		}
	}
}

// buildURL constructs a Graph API URL with the base path and query params.
func (c *MetaAdsClient) buildURL(path, accessToken string, params url.Values) string {
	if params == nil {
		params = url.Values{}
	}
	params.Set("access_token", accessToken)
	params.Set("appsecret_proof", c.appsecretProof(accessToken))
	return metaAdsBaseURL + metaAdsAPIVersion + "/" + path + "?" + params.Encode()
}

// ─────────────────────────────────────────────────────────────────────────────
// Debug Token
// ─────────────────────────────────────────────────────────────────────────────

// DebugTokenResult holds the response from the debug_token endpoint.
type DebugTokenResult struct {
	Data struct {
		AppID               string `json:"app_id"`
		Type                string `json:"type"`
		Application         string `json:"application"`
		DataAccessExpiresAt int64  `json:"data_access_expires_at"`
		ExpiresAt           int64  `json:"expires_at"`
		IsValid             bool   `json:"is_valid"`
		Error               *struct {
			Code    int    `json:"code"`
			Message string `json:"message"`
		} `json:"error,omitempty"`
	} `json:"data"`
}

// DebugToken calls the /debug_token endpoint and returns the parsed result.
// appAccessToken must be "{app_id}|{app_secret}".
// Retry policy: any non-200 response or network error is retried once.
func (c *MetaAdsClient) DebugToken(ctx context.Context, inputToken, appAccessToken string) (*DebugTokenResult, error) {
	params := url.Values{
		"input_token":  []string{inputToken},
		"access_token": []string{appAccessToken},
	}
	rawURL := metaAdsBaseURL + "debug_token?" + params.Encode()

	for attempt := 0; attempt <= 1; attempt++ {
		if err := c.rate.Wait(ctx, appAccessToken); err != nil {
			return nil, fmt.Errorf("MetaAdsClient.DebugToken: rate limit: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
		if err != nil {
			return nil, fmt.Errorf("MetaAdsClient.DebugToken: new request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			if attempt == 0 {
				c.log.Warn().Err(err).Msg("Meta Ads debug_token request failed, retrying once")
				continue
			}
			return nil, fmt.Errorf("MetaAdsClient.DebugToken: do request: %w", err)
		}

		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			if attempt == 0 {
				c.log.Warn().Int("status", resp.StatusCode).Msg("Meta Ads debug_token returned non-200, retrying once")
				continue
			}
			return nil, fmt.Errorf("MetaAdsClient.DebugToken: status %d after retry", resp.StatusCode)
		}

		var result DebugTokenResult
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("MetaAdsClient.DebugToken: unmarshal: %w", err)
		}
		return &result, nil
	}
	return nil, fmt.Errorf("MetaAdsClient.DebugToken: all attempts failed")
}

// ─────────────────────────────────────────────────────────────────────────────
// Account Info
// ─────────────────────────────────────────────────────────────────────────────

// FetchAccountInfo retrieves account-level metadata for the given ad account ID.
// accountID should be in the "act_XXXX" format.
func (c *MetaAdsClient) FetchAccountInfo(ctx context.Context, accountID, accessToken string) (*metaadsmodels.RawMetaAdsAccountInfo, error) {
	fields := "id,name,currency,account_status,timezone_name,business,amount_spent,balance,spend_cap,created_time"
	params := url.Values{"fields": []string{fields}}
	rawURL := c.buildURL(accountID, accessToken, params)

	var result metaadsmodels.RawMetaAdsAccountInfo
	if err := c.get(ctx, accessToken, rawURL, &result); err != nil {
		return nil, fmt.Errorf("MetaAdsClient.FetchAccountInfo(%s): %w", accountID, err)
	}
	return &result, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Campaigns
// ─────────────────────────────────────────────────────────────────────────────

// FetchCampaigns returns all campaigns for the ad account within the given date
// range, following cursor-based pagination.
func (c *MetaAdsClient) FetchCampaigns(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]metaadsmodels.RawMetaAdsCampaign, error) {
	fields := "id,name,status,effective_status,objective,daily_budget,lifetime_budget,budget_remaining,start_time,stop_time,created_time,updated_time"
	params := url.Values{
		"fields":     []string{fields},
		"time_range": []string{fmt.Sprintf(`{"since":"%s","until":"%s"}`, since.Format("2006-01-02"), until.Format("2006-01-02"))},
		"limit":      []string{"200"},
	}
	rawURL := c.buildURL(accountID+"/campaigns", accessToken, params)

	var all []metaadsmodels.RawMetaAdsCampaign
	for page := 0; page < metaAdsMaxPages; page++ {
		var resp metaadsmodels.PaginatedResponse[metaadsmodels.RawMetaAdsCampaign]
		if err := c.get(ctx, accessToken, rawURL, &resp); err != nil {
			return nil, fmt.Errorf("MetaAdsClient.FetchCampaigns page %d: %w", page, err)
		}
		all = append(all, resp.Data...)
		if resp.Paging.Next == "" {
			break
		}
		rawURL = resp.Paging.Next
	}
	return all, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Ad Sets
// ─────────────────────────────────────────────────────────────────────────────

// FetchAdsets returns all ad sets for the ad account within the given date range.
func (c *MetaAdsClient) FetchAdsets(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]metaadsmodels.RawMetaAdsAdset, error) {
	fields := "id,name,campaign_id,status,effective_status,daily_budget,lifetime_budget,budget_remaining,billing_event,optimization_goal,bid_strategy,targeting,start_time,stop_time,end_time,created_time"
	params := url.Values{
		"fields":     []string{fields},
		"time_range": []string{fmt.Sprintf(`{"since":"%s","until":"%s"}`, since.Format("2006-01-02"), until.Format("2006-01-02"))},
		"limit":      []string{"200"},
	}
	rawURL := c.buildURL(accountID+"/adsets", accessToken, params)

	var all []metaadsmodels.RawMetaAdsAdset
	for page := 0; page < metaAdsMaxPages; page++ {
		var resp metaadsmodels.PaginatedResponse[metaadsmodels.RawMetaAdsAdset]
		if err := c.get(ctx, accessToken, rawURL, &resp); err != nil {
			return nil, fmt.Errorf("MetaAdsClient.FetchAdsets page %d: %w", page, err)
		}
		all = append(all, resp.Data...)
		if resp.Paging.Next == "" {
			break
		}
		rawURL = resp.Paging.Next
	}
	return all, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Ads
// ─────────────────────────────────────────────────────────────────────────────

// FetchAds returns all ads for the ad account within the given date range.
func (c *MetaAdsClient) FetchAds(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]metaadsmodels.RawMetaAdsAd, error) {
	fields := "id,name,adset_id,adset{name},campaign_id,campaign{name,objective},status,effective_status,creative{id,name,title,body,image_url,thumbnail_url,object_type,effective_object_story_id},daily_budget,lifetime_budget,budget_remaining,created_time,updated_time"
	params := url.Values{
		"fields":     []string{fields},
		"time_range": []string{fmt.Sprintf(`{"since":"%s","until":"%s"}`, since.Format("2006-01-02"), until.Format("2006-01-02"))},
		"limit":      []string{"200"},
	}
	rawURL := c.buildURL(accountID+"/ads", accessToken, params)

	var all []metaadsmodels.RawMetaAdsAd
	for page := 0; page < metaAdsMaxPages; page++ {
		var resp metaadsmodels.PaginatedResponse[metaadsmodels.RawMetaAdsAd]
		if err := c.get(ctx, accessToken, rawURL, &resp); err != nil {
			return nil, fmt.Errorf("MetaAdsClient.FetchAds page %d: %w", page, err)
		}
		all = append(all, resp.Data...)
		if resp.Paging.Next == "" {
			break
		}
		rawURL = resp.Paging.Next
	}
	return all, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Insights (Campaign / Adset / Ad)
// ─────────────────────────────────────────────────────────────────────────────

// fetchInsights is the shared paginated insights fetcher.
// level must be "campaign", "adset", or "ad".
// levelFields are the level-specific fields (e.g. "campaign_id,campaign_name").
func (c *MetaAdsClient) fetchInsights(
	ctx context.Context,
	accountID, accessToken string,
	level, levelFields string,
	since, until time.Time,
) ([]metaadsmodels.RawMetaAdsInsightRow, error) {
	allFields := levelFields + "," + insightCommonFields
	params := url.Values{
		"fields":         []string{allFields},
		"level":          []string{level},
		"time_increment": []string{"1"},
		"time_range":     []string{fmt.Sprintf(`{"since":"%s","until":"%s"}`, since.Format("2006-01-02"), until.Format("2006-01-02"))},
		"filtering":      []string{insightActionFilter},
		"limit":          []string{"200"},
	}
	rawURL := c.buildURL(accountID+"/insights", accessToken, params)

	var all []metaadsmodels.RawMetaAdsInsightRow
	for page := 0; page < metaAdsMaxPages; page++ {
		var resp metaadsmodels.PaginatedResponse[metaadsmodels.RawMetaAdsInsightRow]
		if err := c.get(ctx, accessToken, rawURL, &resp); err != nil {
			return nil, fmt.Errorf("MetaAdsClient.fetchInsights[%s] page %d: %w", level, page, err)
		}
		all = append(all, resp.Data...)
		if resp.Paging.Next == "" {
			break
		}
		rawURL = resp.Paging.Next
	}
	return all, nil
}

// FetchCampaignInsights returns daily campaign-level insights.
func (c *MetaAdsClient) FetchCampaignInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]metaadsmodels.RawMetaAdsInsightRow, error) {
	return c.fetchInsights(ctx, accountID, accessToken, "campaign", "campaign_id,campaign_name,objective", since, until)
}

// FetchAdsetInsights returns daily ad-set-level insights.
func (c *MetaAdsClient) FetchAdsetInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]metaadsmodels.RawMetaAdsInsightRow, error) {
	return c.fetchInsights(ctx, accountID, accessToken, "adset", "adset_id,adset_name,campaign_id,campaign_name", since, until)
}

// FetchAdInsights returns daily ad-level insights.
func (c *MetaAdsClient) FetchAdInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]metaadsmodels.RawMetaAdsInsightRow, error) {
	return c.fetchInsights(ctx, accountID, accessToken, "ad", "ad_id,ad_name,adset_id,campaign_id,campaign_name", since, until)
}

// ─────────────────────────────────────────────────────────────────────────────
// Demographics
// ─────────────────────────────────────────────────────────────────────────────

// fetchDemographicsInsights fetches insights with a specific breakdown.
func (c *MetaAdsClient) fetchDemographicsInsights(
	ctx context.Context,
	accountID, accessToken string,
	breakdown string,
	since, until time.Time,
) ([]metaadsmodels.RawMetaAdsDemographicsRow, error) {
	demFields := "impressions,reach,clicks,spend,ctr,cpm,cpc,cpp,frequency,date_start,date_stop"
	params := url.Values{
		"fields":         []string{demFields},
		"level":          []string{"account"},
		"time_increment": []string{"1"},
		"breakdowns":     []string{breakdown},
		"time_range":     []string{fmt.Sprintf(`{"since":"%s","until":"%s"}`, since.Format("2006-01-02"), until.Format("2006-01-02"))},
		"limit":          []string{"200"},
	}
	rawURL := c.buildURL(accountID+"/insights", accessToken, params)

	var all []metaadsmodels.RawMetaAdsDemographicsRow
	for page := 0; page < metaAdsMaxPages; page++ {
		var resp metaadsmodels.PaginatedResponse[metaadsmodels.RawMetaAdsDemographicsRow]
		if err := c.get(ctx, accessToken, rawURL, &resp); err != nil {
			return nil, fmt.Errorf("MetaAdsClient.fetchDemographicsInsights[%s] page %d: %w", breakdown, page, err)
		}
		all = append(all, resp.Data...)
		if resp.Paging.Next == "" {
			break
		}
		rawURL = resp.Paging.Next
	}
	return all, nil
}

// FetchAgeGenderInsights returns age/gender breakdown insights.
func (c *MetaAdsClient) FetchAgeGenderInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]metaadsmodels.RawMetaAdsDemographicsRow, error) {
	return c.fetchDemographicsInsights(ctx, accountID, accessToken, "age,gender", since, until)
}

// FetchDevicePlatformInsights returns impression_device/publisher_platform/platform_position breakdown insights.
func (c *MetaAdsClient) FetchDevicePlatformInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]metaadsmodels.RawMetaAdsDemographicsRow, error) {
	return c.fetchDemographicsInsights(ctx, accountID, accessToken, "impression_device,publisher_platform,platform_position", since, until)
}

// FetchRegionCountryInsights returns country/region breakdown insights.
func (c *MetaAdsClient) FetchRegionCountryInsights(ctx context.Context, accountID, accessToken string, since, until time.Time) ([]metaadsmodels.RawMetaAdsDemographicsRow, error) {
	return c.fetchDemographicsInsights(ctx, accountID, accessToken, "country,region", since, until)
}

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// ParseMetaAdsFloat64 safely parses a string as float64; returns 0 on failure.
func ParseMetaAdsFloat64(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseFloat(s, 64)
	return v
}

// ParseMetaAdsInt64 safely parses a string as int64; returns 0 on failure.
func ParseMetaAdsInt64(s string) int64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}
	v, _ := strconv.ParseInt(s, 10, 64)
	return v
}

// MetaAdsActionValue extracts the "value" field for a given action_type from
// an actions array. Returns 0 if the type is not found.
func MetaAdsActionValue(actions []metaadsmodels.RawMetaAdsAction, actionType string) int64 {
	for _, a := range actions {
		if a.ActionType == actionType {
			return ParseMetaAdsInt64(a.Value)
		}
	}
	return 0
}

// TruncateToHour truncates a time.Time value to the start of the hour.
// Used to standardise all timestamps before ClickHouse insertion.
func TruncateToHour(t time.Time) time.Time {
	return t.Truncate(time.Hour)
}
