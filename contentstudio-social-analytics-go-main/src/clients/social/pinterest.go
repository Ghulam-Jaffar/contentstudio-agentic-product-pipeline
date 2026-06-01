package social

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/time/rate"
)

const (
	PinterestAPIBaseURL = "https://api.pinterest.com/v5/"

	defaultPinterestRPS   = 1.0
	defaultPinterestBurst = 1

	defaultPinterestMaxRetries  = 3
	defaultPinterestBaseBackoff = 1 * time.Second
	defaultPinterestMaxBackoff  = 10 * time.Second

	PinterestFullSyncDays        = 86
	PinterestIncrementalSyncDays = 3

	PinterestMultiPinBatchSize = 25
)

type PinterestAPI interface {
	GetUserAccount(ctx context.Context, accessToken string) (*PinterestUserAccount, error)
	GetUserAccountAnalytics(ctx context.Context, accessToken string, startDate, endDate time.Time) (*PinterestUserAnalyticsResponse, error)
	GetBoards(ctx context.Context, accessToken string) (*PinterestBoardsResponse, error)
	GetBoard(ctx context.Context, accessToken, boardID string) (*PinterestBoard, error)
	GetBoardPins(ctx context.Context, accessToken, boardID string, pageSize int, bookmark string) (*PinterestPinsResponse, error)
	GetUserPins(ctx context.Context, accessToken string, pageSize int, bookmark string) (*PinterestPinsResponse, error)
	GetPinAnalytics(ctx context.Context, accessToken, pinID string, startDate, endDate time.Time) (*PinterestPinAnalyticsResponse, error)
	GetMultiPinAnalytics(ctx context.Context, accessToken string, pinIDs []string, startDate, endDate time.Time) (map[string]*PinterestPinAnalyticsResponse, error)
}

type PinterestClient struct {
	HTTPClient  *http.Client
	RateLimiter *rate.Limiter
	MaxRetries  int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
}

var _ PinterestAPI = (*PinterestClient)(nil)

type PinterestClientConfig struct {
	RPS         float64
	Burst       int
	MaxRetries  int
	BaseBackoff time.Duration
	MaxBackoff  time.Duration
}

func NewPinterestClient() *PinterestClient {
	return NewPinterestClientWithConfig(PinterestClientConfig{
		RPS:         defaultPinterestRPS,
		Burst:       defaultPinterestBurst,
		MaxRetries:  defaultPinterestMaxRetries,
		BaseBackoff: defaultPinterestBaseBackoff,
		MaxBackoff:  defaultPinterestMaxBackoff,
	})
}

func NewPinterestClientWithConfig(cfg PinterestClientConfig) *PinterestClient {
	if cfg.RPS <= 0 {
		cfg.RPS = defaultPinterestRPS
	}
	if cfg.Burst <= 0 {
		cfg.Burst = defaultPinterestBurst
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = defaultPinterestMaxRetries
	}
	if cfg.BaseBackoff <= 0 {
		cfg.BaseBackoff = defaultPinterestBaseBackoff
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = defaultPinterestMaxBackoff
	}

	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	return &PinterestClient{
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		RateLimiter: rate.NewLimiter(rate.Limit(cfg.RPS), cfg.Burst),
		MaxRetries:  cfg.MaxRetries,
		BaseBackoff: cfg.BaseBackoff,
		MaxBackoff:  cfg.MaxBackoff,
	}
}

// Pinterest API Response Types

type PinterestUserAccount struct {
	ID             string `json:"id"`
	Username       string `json:"username"`
	About          string `json:"about"`
	ProfileImage   string `json:"profile_image"`
	WebsiteURL     string `json:"website_url"`
	BusinessName   string `json:"business_name"`
	BoardCount     int    `json:"board_count"`
	PinCount       int    `json:"pin_count"`
	AccountType    string `json:"account_type"`
	FollowerCount  int64  `json:"follower_count"`
	FollowingCount int64  `json:"following_count"`
	MonthlyViews   int64  `json:"monthly_views"`
}

type PinterestUserAnalyticsResponse struct {
	All struct {
		DailyMetrics []PinterestDailyMetric `json:"daily_metrics"`
	} `json:"all"`
}

type PinterestDailyMetric struct {
	Date       string                 `json:"date"`
	DataStatus string                 `json:"data_status"`
	Metrics    map[string]interface{} `json:"metrics"`
}

type PinterestBoardsResponse struct {
	Items    []PinterestBoard `json:"items"`
	Bookmark string           `json:"bookmark,omitempty"`
}

type PinterestBoard struct {
	ID                string                 `json:"id"`
	Name              string                 `json:"name"`
	Description       string                 `json:"description"`
	Privacy           string                 `json:"privacy"`
	PinCount          int                    `json:"pin_count"`
	FollowerCount     int                    `json:"follower_count"`
	CollaboratorCount int                    `json:"collaborator_count"`
	CreatedAt         string                 `json:"created_at"`
	Owner             map[string]interface{} `json:"owner"`
	Media             map[string]interface{} `json:"media"`
	PinThumbnailURLs  []string               `json:"pin_thumbnail_urls"`
}

type PinterestPinsResponse struct {
	Items    []PinterestPin `json:"items"`
	Bookmark string         `json:"bookmark,omitempty"`
}

type PinterestPin struct {
	ID              string                 `json:"id"`
	Title           string                 `json:"title"`
	Description     string                 `json:"description"`
	Note            string                 `json:"note"`
	Link            string                 `json:"link"`
	DominantColor   string                 `json:"dominant_color"`
	BoardID         string                 `json:"board_id"`
	BoardSectionID  string                 `json:"board_section_id"`
	ParentPinID     string                 `json:"parent_pin_id"`
	CreatedAt       string                 `json:"created_at"`
	CreativeType    string                 `json:"creative_type"`
	IsStandard      bool                   `json:"is_standard"`
	IsOwner         bool                   `json:"is_owner"`
	HasBeenPromoted bool                   `json:"has_been_promoted"`
	Media           map[string]interface{} `json:"media"`
	BoardOwner      map[string]interface{} `json:"board_owner"`
	ProductTags     []interface{}          `json:"product_tags"`
}

type PinterestPinAnalyticsResponse struct {
	All struct {
		DailyMetrics []PinterestDailyMetric `json:"daily_metrics"`
	} `json:"all"`
}

func (c *PinterestClient) makeRequest(ctx context.Context, method, reqURL string, accessToken string, body io.Reader) ([]byte, int, error) {
	var lastErr error
	var lastStatus int

	for attempt := 0; attempt < c.MaxRetries; attempt++ {
		if err := c.RateLimiter.Wait(ctx); err != nil {
			return nil, 0, fmt.Errorf("PinterestClient.makeRequest: rate limiter wait failed: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, method, reqURL, body)
		if err != nil {
			return nil, 0, fmt.Errorf("PinterestClient.makeRequest: failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)
		if body != nil {
			req.Header.Set("Content-Type", "application/json")
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err
			if attempt < c.MaxRetries-1 {
				if sleepErr := c.exponentialBackoff(ctx, attempt); sleepErr != nil {
					return nil, 0, sleepErr
				}
			}
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		lastStatus = resp.StatusCode

		if resp.StatusCode == http.StatusOK {
			return respBody, resp.StatusCode, nil
		}

		if resp.StatusCode == http.StatusUnauthorized {
			return respBody, resp.StatusCode, fmt.Errorf("PinterestClient.makeRequest: pinterest API unauthorized (status 401): %s", string(respBody))
		}

		if resp.StatusCode == http.StatusTooManyRequests {
			resetSeconds := 60
			if reset := resp.Header.Get("x-ratelimit-reset-seconds"); reset != "" {
				fmt.Sscanf(reset, "%d", &resetSeconds)
			}
			select {
			case <-ctx.Done():
				return nil, 0, ctx.Err()
			case <-time.After(time.Duration(resetSeconds+1) * time.Second):
			}
			continue
		}

		if resp.StatusCode == http.StatusInternalServerError {
			return respBody, resp.StatusCode, nil
		}

		lastErr = fmt.Errorf("PinterestClient.makeRequest: request failed with status %d: %s", resp.StatusCode, string(respBody))
		if attempt < c.MaxRetries-1 {
			if sleepErr := c.exponentialBackoff(ctx, attempt); sleepErr != nil {
				return nil, 0, sleepErr
			}
		}
	}

	return nil, lastStatus, lastErr
}

func (c *PinterestClient) exponentialBackoff(ctx context.Context, attempt int) error {
	backoff := c.BaseBackoff * time.Duration(1<<attempt)
	if backoff > c.MaxBackoff {
		backoff = c.MaxBackoff
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(backoff):
		return nil
	}
}

func (c *PinterestClient) GetUserAccount(ctx context.Context, accessToken string) (*PinterestUserAccount, error) {
	reqURL := PinterestAPIBaseURL + "user_account"

	body, status, err := c.makeRequest(ctx, http.MethodGet, reqURL, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("PinterestClient.GetUserAccount: failed to fetch user account: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("PinterestClient.GetUserAccount: pinterest API error (status %d): %s", status, string(body))
	}

	var resp PinterestUserAccount
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("PinterestClient.GetUserAccount: failed to parse user account response: %w", err)
	}

	return &resp, nil
}

func (c *PinterestClient) GetUserAccountAnalytics(ctx context.Context, accessToken string, startDate, endDate time.Time) (*PinterestUserAnalyticsResponse, error) {
	params := url.Values{}
	params.Set("start_date", startDate.Format("2006-01-02"))
	params.Set("end_date", endDate.Format("2006-01-02"))

	reqURL := PinterestAPIBaseURL + "user_account/analytics?" + params.Encode()

	body, status, err := c.makeRequest(ctx, http.MethodGet, reqURL, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("PinterestClient.GetUserAccountAnalytics: failed to fetch user analytics: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("PinterestClient.GetUserAccountAnalytics: pinterest API error (status %d): %s", status, string(body))
	}

	var resp PinterestUserAnalyticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("PinterestClient.GetUserAccountAnalytics: failed to parse user analytics response: %w", err)
	}

	return &resp, nil
}

func (c *PinterestClient) GetBoards(ctx context.Context, accessToken string) (*PinterestBoardsResponse, error) {
	reqURL := PinterestAPIBaseURL + "boards"

	body, status, err := c.makeRequest(ctx, http.MethodGet, reqURL, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("PinterestClient.GetBoards: failed to fetch boards: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("PinterestClient.GetBoards: pinterest API error (status %d): %s", status, string(body))
	}

	var resp PinterestBoardsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("PinterestClient.GetBoards: failed to parse boards response: %w", err)
	}

	return &resp, nil
}

func (c *PinterestClient) GetBoard(ctx context.Context, accessToken, boardID string) (*PinterestBoard, error) {
	reqURL := PinterestAPIBaseURL + "boards/" + boardID

	body, status, err := c.makeRequest(ctx, http.MethodGet, reqURL, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("PinterestClient.GetBoard: failed to fetch board: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("PinterestClient.GetBoard: pinterest API error (status %d): %s", status, string(body))
	}

	var resp PinterestBoard
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("PinterestClient.GetBoard: failed to parse board response: %w", err)
	}

	return &resp, nil
}

func (c *PinterestClient) GetBoardPins(ctx context.Context, accessToken, boardID string, pageSize int, bookmark string) (*PinterestPinsResponse, error) {
	params := url.Values{}
	params.Set("page_size", fmt.Sprintf("%d", pageSize))
	params.Set("include_protected_pins", "true")
	if bookmark != "" {
		params.Set("bookmark", bookmark)
	}

	reqURL := PinterestAPIBaseURL + "boards/" + boardID + "/pins?" + params.Encode()

	body, status, err := c.makeRequest(ctx, http.MethodGet, reqURL, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("PinterestClient.GetBoardPins: failed to fetch board pins: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("PinterestClient.GetBoardPins: pinterest API error (status %d): %s", status, string(body))
	}

	var resp PinterestPinsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("PinterestClient.GetBoardPins: failed to parse board pins response: %w", err)
	}

	return &resp, nil
}

func (c *PinterestClient) GetUserPins(ctx context.Context, accessToken string, pageSize int, bookmark string) (*PinterestPinsResponse, error) {
	params := url.Values{}
	params.Set("page_size", fmt.Sprintf("%d", pageSize))
	if bookmark != "" {
		params.Set("bookmark", bookmark)
	}

	reqURL := PinterestAPIBaseURL + "pins?" + params.Encode()

	body, status, err := c.makeRequest(ctx, http.MethodGet, reqURL, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("PinterestClient.GetUserPins: failed to fetch user pins: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("PinterestClient.GetUserPins: pinterest API error (status %d): %s", status, string(body))
	}

	var resp PinterestPinsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("PinterestClient.GetUserPins: failed to parse user pins response: %w", err)
	}

	return &resp, nil
}

func (c *PinterestClient) GetPinAnalytics(ctx context.Context, accessToken, pinID string, startDate, endDate time.Time) (*PinterestPinAnalyticsResponse, error) {
	params := url.Values{}
	params.Set("start_date", startDate.Format("2006-01-02"))
	params.Set("end_date", endDate.Format("2006-01-02"))
	params.Set("metric_types", "ALL")

	reqURL := PinterestAPIBaseURL + "pins/" + pinID + "/analytics?" + params.Encode()

	body, status, err := c.makeRequest(ctx, http.MethodGet, reqURL, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("PinterestClient.GetPinAnalytics: failed to fetch pin analytics: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("PinterestClient.GetPinAnalytics: pinterest API error (status %d): %s", status, string(body))
	}

	var resp PinterestPinAnalyticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("PinterestClient.GetPinAnalytics: failed to parse pin analytics response: %w", err)
	}

	return &resp, nil
}

func (c *PinterestClient) GetMultiPinAnalytics(ctx context.Context, accessToken string, pinIDs []string, startDate, endDate time.Time) (map[string]*PinterestPinAnalyticsResponse, error) {
	if len(pinIDs) == 0 {
		return make(map[string]*PinterestPinAnalyticsResponse), nil
	}

	result := make(map[string]*PinterestPinAnalyticsResponse)

	for i := 0; i < len(pinIDs); i += PinterestMultiPinBatchSize {
		end := i + PinterestMultiPinBatchSize
		if end > len(pinIDs) {
			end = len(pinIDs)
		}
		batch := pinIDs[i:end]

		batchResult, err := c.fetchMultiPinAnalyticsBatch(ctx, accessToken, batch, startDate, endDate)
		if err != nil {
			for _, pinID := range batch {
				singleResult, singleErr := c.GetPinAnalytics(ctx, accessToken, pinID, startDate, endDate)
				if singleErr == nil {
					result[pinID] = singleResult
				}
			}
			continue
		}

		for pinID, analytics := range batchResult {
			result[pinID] = analytics
		}
	}

	return result, nil
}

func (c *PinterestClient) fetchMultiPinAnalyticsBatch(ctx context.Context, accessToken string, pinIDs []string, startDate, endDate time.Time) (map[string]*PinterestPinAnalyticsResponse, error) {
	params := url.Values{}
	params.Set("pin_ids", strings.Join(pinIDs, ","))
	params.Set("start_date", startDate.Format("2006-01-02"))
	params.Set("end_date", endDate.Format("2006-01-02"))
	params.Set("metric_types", "ALL")

	reqURL := PinterestAPIBaseURL + "pins/analytics?" + params.Encode()

	body, status, err := c.makeRequest(ctx, http.MethodGet, reqURL, accessToken, nil)
	if err != nil {
		return nil, fmt.Errorf("PinterestClient.fetchMultiPinAnalyticsBatch: failed to fetch multi-pin analytics: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("PinterestClient.fetchMultiPinAnalyticsBatch: pinterest API error (status %d): %s", status, string(body))
	}

	return c.normalizeMultiPinAnalyticsResponse(body)
}

func (c *PinterestClient) normalizeMultiPinAnalyticsResponse(body []byte) (map[string]*PinterestPinAnalyticsResponse, error) {
	var rawResp map[string]interface{}
	if err := json.Unmarshal(body, &rawResp); err != nil {
		return nil, err
	}

	result := make(map[string]*PinterestPinAnalyticsResponse)

	for key, value := range rawResp {
		if valueMap, ok := value.(map[string]interface{}); ok {
			if _, hasAll := valueMap["all"]; hasAll {
				pinBytes, err := json.Marshal(valueMap)
				if err != nil {
					continue
				}
				var pinAnalytics PinterestPinAnalyticsResponse
				if err := json.Unmarshal(pinBytes, &pinAnalytics); err != nil {
					continue
				}
				result[key] = &pinAnalytics
			}
		}
	}

	if len(result) > 0 {
		return result, nil
	}

	var listKeys = []string{"items", "data", "results"}
	for _, listKey := range listKeys {
		if items, ok := rawResp[listKey].([]interface{}); ok {
			return c.normalizeMultiPinAnalyticsList(items)
		}
	}

	return result, nil
}

func (c *PinterestClient) normalizeMultiPinAnalyticsList(items []interface{}) (map[string]*PinterestPinAnalyticsResponse, error) {
	grouped := make(map[string][]PinterestDailyMetric)

	for _, item := range items {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			continue
		}

		var pinID string
		for _, key := range []string{"pin_id", "pinId", "id"} {
			if id, ok := itemMap[key].(string); ok && id != "" {
				pinID = id
				break
			}
		}
		if pinID == "" {
			continue
		}

		if metrics, ok := itemMap["metrics"].(map[string]interface{}); ok {
			if date, ok := itemMap["date"].(string); ok {
				dataStatus := "READY"
				if ds, ok := itemMap["data_status"].(string); ok {
					dataStatus = ds
				}
				grouped[pinID] = append(grouped[pinID], PinterestDailyMetric{
					Date:       date,
					DataStatus: dataStatus,
					Metrics:    metrics,
				})
			}
		}

		if dailyMetrics, ok := itemMap["daily_metrics"].([]interface{}); ok {
			for _, dm := range dailyMetrics {
				if dmMap, ok := dm.(map[string]interface{}); ok {
					date, _ := dmMap["date"].(string)
					dataStatus, _ := dmMap["data_status"].(string)
					metrics, _ := dmMap["metrics"].(map[string]interface{})
					if date != "" && metrics != nil {
						grouped[pinID] = append(grouped[pinID], PinterestDailyMetric{
							Date:       date,
							DataStatus: dataStatus,
							Metrics:    metrics,
						})
					}
				}
			}
		}
	}

	result := make(map[string]*PinterestPinAnalyticsResponse)
	for pinID, dailyMetrics := range grouped {
		result[pinID] = &PinterestPinAnalyticsResponse{
			All: struct {
				DailyMetrics []PinterestDailyMetric `json:"daily_metrics"`
			}{
				DailyMetrics: dailyMetrics,
			},
		}
	}

	return result, nil
}

func GetInt64FromMetrics(metrics map[string]interface{}, key string) int64 {
	if val, ok := metrics[key]; ok {
		switch v := val.(type) {
		case float64:
			return int64(v)
		case int64:
			return v
		case int:
			return int64(v)
		}
	}
	return 0
}

func GetFloat64FromMetrics(metrics map[string]interface{}, key string) float64 {
	if val, ok := metrics[key]; ok {
		switch v := val.(type) {
		case float64:
			return v
		case int64:
			return float64(v)
		case int:
			return float64(v)
		}
	}
	return 0
}

func GetStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func GetMediaField(media map[string]interface{}, field string) string {
	if val, ok := media[field]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

func GetPinCoverImageURL(pin PinterestPin) string {
	if pin.Media == nil {
		return ""
	}

	mediaType := GetMediaField(pin.Media, "media_type")

	if strings.Contains(mediaType, "video") {
		return GetMediaField(pin.Media, "cover_image_url")
	}

	if strings.Contains(mediaType, "multiple_images") {
		if items, ok := pin.Media["items"].([]interface{}); ok && len(items) > 0 {
			if firstItem, ok := items[0].(map[string]interface{}); ok {
				if images, ok := firstItem["images"].(map[string]interface{}); ok {
					if img150, ok := images["150x150"].(map[string]interface{}); ok {
						return GetStringFromMap(img150, "url")
					}
				}
			}
		}
		return ""
	}

	if images, ok := pin.Media["images"].(map[string]interface{}); ok {
		if img150, ok := images["150x150"].(map[string]interface{}); ok {
			return GetStringFromMap(img150, "url")
		}
	}

	return ""
}
