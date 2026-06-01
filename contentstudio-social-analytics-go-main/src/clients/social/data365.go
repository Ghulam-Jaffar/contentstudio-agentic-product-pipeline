package social

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// Data365SearchStatus represents the status of a Data365 async search task.
// The API wraps the task status inside the `data` envelope:
// {"data": {"status": "finished"}, "status": "ok", "error": null}
type Data365SearchStatus struct {
	Data struct {
		Status string `json:"status"`
		Error  string `json:"error,omitempty"`
	} `json:"data"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

// Data365SearchResult holds paginated search results from Data365.
// The API response shape is: {"data": {"items": [...], "page_info": {"cursor": "...", "has_next_page": true}}}
type Data365SearchResult struct {
	Data json.RawMessage `json:"data"`
	// Cursor is extracted from data.page_info.cursor after decode.
	Cursor string
}

// UnsupportedSearchError means Data365 rejected a search/update request for
// this platform+query combination as unsupported for the current endpoint.
// The fetcher treats this as a non-retryable skip for that specific pair.
type UnsupportedSearchError struct {
	Platform   string
	Keyword    string
	StatusCode int
}

func (e *UnsupportedSearchError) Error() string {
	return fmt.Sprintf("Data365Client.TriggerSearch: HTTP %d", e.StatusCode)
}

func IsUnsupportedSearchError(err error) bool {
	var target *UnsupportedSearchError
	return errors.As(err, &target)
}

// data365ResultEnvelope is used internally to extract items and pagination from the data field.
type data365ResultEnvelope struct {
	Items    []json.RawMessage `json:"items"`
	PageInfo struct {
		Cursor      string `json:"cursor"`
		HasNextPage bool   `json:"has_next_page"`
	} `json:"page_info"`
}

// Data365Client handles communication with the Data365 social listening API.
type Data365Client struct {
	httpClient   *http.Client
	baseURL      string
	accessToken  string
	pollInterval time.Duration
	pollTimeout  time.Duration
	log          *logger.Logger
}

// NewData365Client creates a Data365Client from configuration.
func NewData365Client(cfg config.Data365Config, log *logger.Logger) *Data365Client {
	pollInterval := time.Duration(cfg.PollInterval) * time.Second
	if pollInterval == 0 {
		pollInterval = 5 * time.Second
	}
	pollTimeout := time.Duration(cfg.PollTimeout) * time.Second
	if pollTimeout == 0 {
		// 15 min default leaves headroom for Data365 to finish high-volume
		// searches (e.g., max_posts=5000 on initial sync). Override via
		// APP_DATA365_POLL_TIMEOUT (seconds) if Data365 latency drifts.
		pollTimeout = 15 * time.Minute
	}
	return &Data365Client{
		httpClient:   &http.Client{Timeout: 30 * time.Second},
		baseURL:      cfg.BaseURL,
		accessToken:  cfg.AccessToken,
		pollInterval: pollInterval,
		pollTimeout:  pollTimeout,
		log:          log,
	}
}

// searchUpdatePath returns the POST/GET update URL per platform.
func searchUpdatePath(platform, keyword string) string {
	switch platform {
	case "facebook":
		return fmt.Sprintf("/facebook/search/%s/posts/latest/update", url.PathEscape(keyword))
	case "instagram":
		return "/instagram/search/post/update"
	case "tiktok":
		return "/tiktok/search/post/update"
	case "twitter":
		return "/twitter/search/post/update"
	case "reddit":
		return "/reddit/search/post/update"
	case "threads":
		return "/threads/search/post/update"
	default:
		return fmt.Sprintf("/%s/search/post/update", platform)
	}
}

// searchResultsPath returns the GET results URL per platform.
func searchResultsPath(platform, keyword string) string {
	switch platform {
	case "facebook":
		return fmt.Sprintf("/facebook/search/%s/posts/latest/posts", url.PathEscape(keyword))
	case "instagram":
		return "/instagram/search/post/items"
	case "tiktok":
		return "/tiktok/search/post/items"
	case "twitter":
		return "/twitter/search/post/posts"
	case "reddit":
		return "/reddit/search/post/items"
	case "threads":
		return "/threads/search/post/items"
	default:
		return fmt.Sprintf("/%s/search/post/items", platform)
	}
}

// buildSearchParams builds query parameters for a search update/status request.
// Each platform has different supported params — this function applies only what
// Data365 actually accepts per platform to avoid silently-ignored parameters.
func buildSearchParams(platform, keyword, accessToken string, maxPosts int, fromDate, toDate time.Time, languages []string) url.Values {
	params := url.Values{}
	params.Set("access_token", accessToken)
	if maxPosts > 0 {
		params.Set("max_posts", fmt.Sprintf("%d", maxPosts))
	}

	switch platform {
	case "facebook":
		// Keyword is a URL path parameter for Facebook — not added here.
		// Language filtering is only available on the results endpoint, not the trigger.
		if !fromDate.IsZero() {
			params.Set("from_date", fromDate.UTC().Format("2006-01-02"))
		}
		if !toDate.IsZero() {
			params.Set("to_date", toDate.UTC().Format("2006-01-02"))
		}

	case "instagram":
		// Instagram trigger supports keywords and max_posts only.
		// Date range and language are not available in the trigger endpoint.
		params.Set("keywords", keyword)

	case "twitter":
		// Twitter supports keywords, search type, date range, and language in the trigger.
		params.Set("keywords", keyword)
		params.Set("search_type", "latest")
		if !fromDate.IsZero() {
			params.Set("from_date", fromDate.UTC().Format("2006-01-02"))
		}
		if !toDate.IsZero() {
			params.Set("to_date", toDate.UTC().Format("2006-01-02"))
		}
		if len(languages) > 0 {
			params.Set("lang", strings.Join(languages, ","))
		}

	case "tiktok":
		// TikTok requires load_posts=true to actually fetch posts alongside search results.
		// Date range is not supported in the keyword trigger (only a date_posted bucket enum).
		params.Set("keywords", keyword)
		params.Set("load_posts", "true")
		params.Set("sort_type", "relevance")

	case "reddit":
		// sort_type=new enables from_date filtering; to_date is only in the results endpoint.
		params.Set("keywords", keyword)
		params.Set("sort_type", "new")
		if !fromDate.IsZero() {
			params.Set("from_date", fromDate.UTC().Format("2006-01-02"))
		}

	case "threads":
		// Threads supports keywords, sort order, and date range in the trigger.
		params.Set("keywords", keyword)
		params.Set("sort_type", "recent")
		if !fromDate.IsZero() {
			params.Set("from_date", fromDate.UTC().Format("2006-01-02"))
		}
		if !toDate.IsZero() {
			params.Set("to_date", toDate.UTC().Format("2006-01-02"))
		}

	default:
		params.Set("keywords", keyword)
		if !fromDate.IsZero() {
			params.Set("from_date", fromDate.UTC().Format("2006-01-02"))
		}
		if !toDate.IsZero() {
			params.Set("to_date", toDate.UTC().Format("2006-01-02"))
		}
	}

	return params
}

// TriggerSearch initiates an async search task on Data365 for the given platform and keyword.
// fromDate, when non-zero, restricts results to posts on or after that date.
// toDate, when non-zero, restricts results to posts before that date.
// languages, when non-empty, restricts results to those ISO 639-1 language codes.
func (c *Data365Client) TriggerSearch(ctx context.Context, platform, keyword string, maxPosts int, fromDate, toDate time.Time, languages []string) error {
	path := searchUpdatePath(platform, keyword)
	params := buildSearchParams(platform, keyword, c.accessToken, maxPosts, fromDate, toDate, languages)
	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, reqURL, nil)
	if err != nil {
		return fmt.Errorf("Data365Client.TriggerSearch: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("Data365Client.TriggerSearch: HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden {
		c.log.Warn().
			Str("platform", platform).
			Str("keyword", keyword).
			Int("status", resp.StatusCode).
			Msg("Data365 search not available for this query")
		return &UnsupportedSearchError{
			Platform:   platform,
			Keyword:    keyword,
			StatusCode: resp.StatusCode,
		}
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("Data365Client.TriggerSearch: HTTP %d: %s", resp.StatusCode, string(body))
	}

	return nil
}

// PollUntilFinished polls the search task status until it reaches "finished" or times out.
// maxPosts, fromDate, toDate, and languages must match the values used in TriggerSearch so Data365 resolves the same task.
func (c *Data365Client) PollUntilFinished(ctx context.Context, platform, keyword string, maxPosts int, fromDate, toDate time.Time, languages []string) error {
	path := searchUpdatePath(platform, keyword)
	params := buildSearchParams(platform, keyword, c.accessToken, maxPosts, fromDate, toDate, languages)
	statusURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	deadline := time.Now().Add(c.pollTimeout)
	interval := c.pollInterval

	for {
		if time.Now().After(deadline) {
			return fmt.Errorf("Data365Client.PollUntilFinished: timeout after %v", c.pollTimeout)
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(interval):
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
		if err != nil {
			return fmt.Errorf("Data365Client.PollUntilFinished: build request: %w", err)
		}

		resp, err := c.httpClient.Do(req)
		if err != nil {
			return fmt.Errorf("Data365Client.PollUntilFinished: HTTP request: %w", err)
		}

		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return fmt.Errorf("Data365Client.PollUntilFinished: read body: %w", readErr)
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return fmt.Errorf("Data365Client.PollUntilFinished: HTTP %d: %s", resp.StatusCode, string(body))
		}

		var status Data365SearchStatus
		if decErr := json.Unmarshal(body, &status); decErr != nil {
			return fmt.Errorf("Data365Client.PollUntilFinished: decode status: %w", decErr)
		}

		c.log.Debug().
			Str("platform", platform).
			Str("keyword", keyword).
			Str("status", status.Data.Status).
			Msg("Poll status")

		switch status.Data.Status {
		case "finished":
			return nil
		case "error", "failed":
			return fmt.Errorf("Data365Client.PollUntilFinished: task failed: %s", status.Data.Error)
		}

		// Exponential backoff up to 30s
		interval = time.Duration(float64(interval) * 1.5)
		if interval > 30*time.Second {
			interval = 30 * time.Second
		}
	}
}

// FetchResults retrieves paginated search results. Pass empty cursor for first page.
// fromDate/toDate apply date filters on Data365's cached results — critical for
// Instagram and TikTok where the trigger endpoint does not support date ranges.
// languages applies lang= filtering on results for Facebook and Twitter (the only
// platforms Data365 supports language filtering on the results endpoint).
func (c *Data365Client) FetchResults(ctx context.Context, platform, keyword, cursor string, fromDate, toDate time.Time, languages []string) (*Data365SearchResult, error) {
	path := searchResultsPath(platform, keyword)
	params := url.Values{}
	params.Set("access_token", c.accessToken)
	params.Set("max_page_size", "100")
	if platform != "facebook" {
		params.Set("keywords", keyword)
	}
	if cursor != "" {
		params.Set("cursor", cursor)
	}
	// Date range filters on cached results — supported by all platforms in the results endpoint.
	if !fromDate.IsZero() {
		params.Set("from_date", fromDate.UTC().Format("2006-01-02"))
	}
	if !toDate.IsZero() {
		params.Set("to_date", toDate.UTC().Format("2006-01-02"))
	}
	// Language filter is available in results only for Facebook and Twitter.
	if len(languages) > 0 && (platform == "facebook" || platform == "twitter") {
		params.Set("lang", strings.Join(languages, ","))
	}

	reqURL := fmt.Sprintf("%s%s?%s", c.baseURL, path, params.Encode())

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("Data365Client.FetchResults: build request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("Data365Client.FetchResults: HTTP request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("Data365Client.FetchResults: HTTP %d: %s", resp.StatusCode, string(body))
	}

	var result Data365SearchResult
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("Data365Client.FetchResults: decode response: %w", err)
	}

	// Extract cursor from data.page_info.cursor
	if len(result.Data) > 0 {
		var envelope data365ResultEnvelope
		if err := json.Unmarshal(result.Data, &envelope); err == nil && envelope.PageInfo.HasNextPage {
			result.Cursor = envelope.PageInfo.Cursor
		}
	}

	return &result, nil
}
