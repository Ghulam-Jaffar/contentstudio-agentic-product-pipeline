package social

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// YouTube API endpoints.
const (
	YouTubeDataAPIURL      = "https://www.googleapis.com/youtube/v3/"
	YouTubeAnalyticsAPIURL = "https://youtubeanalytics.googleapis.com/v2/"
	YouTubeOAuthTokenURL   = "https://oauth2.googleapis.com/token"
	YouTubeShortsURL       = "https://www.youtube.com/shorts/"

	// YouTube Shorts are videos ≤60 seconds
	maxShortDurationSeconds = 60

	// Rate limiting defaults
	defaultYouTubeRPS   = 5.0 // 5 requests per second
	defaultYouTubeBurst = 10  // Allow burst of 10 requests

	// Retry settings with exponential backoff
	defaultMaxRetries  = 4
	defaultBaseBackoff = 500 * time.Millisecond
	defaultMaxBackoff  = 8 * time.Second

	// Parallel short detection settings
	maxParallelShortChecks = 15
)

// YouTubeAPI defines the interface for YouTube client operations.
// This interface allows for mocking in tests.
type YouTubeAPI interface {
	RefreshToken(ctx context.Context, refreshToken string) (*YouTubeTokenResponse, error)
	FetchChannels(ctx context.Context, accessToken string) (*YouTubeChannelResponse, error)
	FetchVideos(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]YouTubeActivityItem, error)
	FetchVideoDetails(ctx context.Context, accessToken string, videoIDs []string) ([]YouTubeVideoItem, error)
	FetchActivityInsights(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error)
	FetchTrafficInsights(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error)
	FetchSharedInsights(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error)
	FetchVideoInsights(ctx context.Context, accessToken, videoID string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error)
	FetchAllVideosAnalytics(ctx context.Context, accessToken string, startDate, endDate time.Time) (map[string]*VideoAnalytics, error)
	DetectMediaTypes(ctx context.Context, videos []YouTubeVideoItem) map[string]string
	IsYouTubeShort(ctx context.Context, videoID string) bool
}

// YouTubeClient handles communication with YouTube Data API v3 and Analytics API v2.
type YouTubeClient struct {
	HTTPClient      *http.Client
	ShortHTTPClient *http.Client // Dedicated client for short detection (no redirects)
	ClientID        string
	ClientSecret    string
	RateLimiter     *rate.Limiter
	MaxRetries      int
	BaseBackoff     time.Duration
	MaxBackoff      time.Duration
}

// Verify YouTubeClient implements YouTubeAPI
var _ YouTubeAPI = (*YouTubeClient)(nil)

// YouTubeClientConfig holds configuration for the YouTube client.
type YouTubeClientConfig struct {
	ClientID     string
	ClientSecret string
	RPS          float64 // Requests per second
	Burst        int     // Burst size
	MaxRetries   int
	BaseBackoff  time.Duration
	MaxBackoff   time.Duration
}

// NewYouTubeClient returns a new YouTube client with default settings.
func NewYouTubeClient(clientID, clientSecret string) *YouTubeClient {
	return NewYouTubeClientWithConfig(YouTubeClientConfig{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RPS:          defaultYouTubeRPS,
		Burst:        defaultYouTubeBurst,
		MaxRetries:   defaultMaxRetries,
		BaseBackoff:  defaultBaseBackoff,
		MaxBackoff:   defaultMaxBackoff,
	})
}

// NewYouTubeClientWithConfig returns a new YouTube client with custom configuration.
func NewYouTubeClientWithConfig(cfg YouTubeClientConfig) *YouTubeClient {
	// Set defaults if not provided
	if cfg.RPS <= 0 {
		cfg.RPS = defaultYouTubeRPS
	}
	if cfg.Burst <= 0 {
		cfg.Burst = defaultYouTubeBurst
	}
	if cfg.MaxRetries <= 0 {
		cfg.MaxRetries = defaultMaxRetries
	}
	if cfg.BaseBackoff <= 0 {
		cfg.BaseBackoff = defaultBaseBackoff
	}
	if cfg.MaxBackoff <= 0 {
		cfg.MaxBackoff = defaultMaxBackoff
	}

	// Create HTTP client with connection pooling
	transport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 20,
		IdleConnTimeout:     90 * time.Second,
	}

	// Create dedicated client for short detection (doesn't follow redirects)
	shortTransport := &http.Transport{
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     60 * time.Second,
	}

	return &YouTubeClient{
		HTTPClient: &http.Client{
			Timeout:   30 * time.Second,
			Transport: transport,
		},
		ShortHTTPClient: &http.Client{
			Timeout:   5 * time.Second,
			Transport: shortTransport,
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		RateLimiter:  rate.NewLimiter(rate.Limit(cfg.RPS), cfg.Burst),
		MaxRetries:   cfg.MaxRetries,
		BaseBackoff:  cfg.BaseBackoff,
		MaxBackoff:   cfg.MaxBackoff,
	}
}

// YouTubeTokenResponse represents an OAuth token refresh response.
type YouTubeTokenResponse struct {
	AccessToken  string `json:"access_token"`
	ExpiresIn    int    `json:"expires_in"`
	Scope        string `json:"scope"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token,omitempty"`
}

// YouTubeChannelResponse represents the channels.list API response.
type YouTubeChannelResponse struct {
	Kind     string `json:"kind"`
	Etag     string `json:"etag"`
	PageInfo struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []YouTubeChannelItem `json:"items"`
}

// YouTubeChannelItem represents a single channel from the API.
type YouTubeChannelItem struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet struct {
		Title       string `json:"title"`
		Description string `json:"description"`
		CustomURL   string `json:"customUrl"`
		PublishedAt string `json:"publishedAt"`
		Thumbnails  struct {
			Default struct {
				URL string `json:"url"`
			} `json:"default"`
			High struct {
				URL string `json:"url"`
			} `json:"high"`
		} `json:"thumbnails"`
		Country string `json:"country"`
	} `json:"snippet"`
	Statistics struct {
		ViewCount             string `json:"viewCount"`
		SubscriberCount       string `json:"subscriberCount"`
		HiddenSubscriberCount bool   `json:"hiddenSubscriberCount"`
		VideoCount            string `json:"videoCount"`
	} `json:"statistics"`
	BrandingSettings struct {
		Image struct {
			BannerExternalURL string `json:"bannerExternalUrl"`
		} `json:"image"`
	} `json:"brandingSettings"`
	ContentDetails struct {
		RelatedPlaylists struct {
			Uploads string `json:"uploads"`
		} `json:"relatedPlaylists"`
	} `json:"contentDetails"`
}

// YouTubeActivitiesResponse represents the activities.list API response.
type YouTubeActivitiesResponse struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	PageInfo      struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []YouTubeActivityItem `json:"items"`
}

// YouTubeActivityItem represents a single activity (video upload) from the API.
type YouTubeActivityItem struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet struct {
		PublishedAt string `json:"publishedAt"`
		ChannelID   string `json:"channelId"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Thumbnails  struct {
			Default struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"default"`
			High struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"high"`
		} `json:"thumbnails"`
		Type string `json:"type"`
	} `json:"snippet"`
	ContentDetails struct {
		Upload struct {
			VideoID string `json:"videoId"`
		} `json:"upload"`
	} `json:"contentDetails"`
}

// YouTubeAnalyticsResponse represents the Analytics API reports response.
type YouTubeAnalyticsResponse struct {
	Kind          string `json:"kind"`
	ColumnHeaders []struct {
		Name       string `json:"name"`
		ColumnType string `json:"columnType"`
		DataType   string `json:"dataType"`
	} `json:"columnHeaders"`
	Rows [][]interface{} `json:"rows"`
}

// RefreshToken refreshes the OAuth access token using the refresh token.
func (c *YouTubeClient) RefreshToken(ctx context.Context, refreshToken string) (*YouTubeTokenResponse, error) {
	data := url.Values{}
	data.Set("client_id", c.ClientID)
	data.Set("client_secret", c.ClientSecret)
	data.Set("refresh_token", refreshToken)
	data.Set("grant_type", "refresh_token")

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, YouTubeOAuthTokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("YouTubeClient.RefreshToken: failed to create refresh token request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("YouTubeClient.RefreshToken: refresh token request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("YouTubeClient.RefreshToken: failed to read refresh token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("YouTubeClient.RefreshToken: refresh token failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp YouTubeTokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("YouTubeClient.RefreshToken: failed to parse refresh token response: %w", err)
	}

	return &tokenResp, nil
}

// makeRequest performs a GET request with rate limiting and exponential backoff retry.
func (c *YouTubeClient) makeRequest(ctx context.Context, reqURL string, accessToken string) ([]byte, int, error) {
	var lastErr error
	var lastStatus int

	for attempt := 0; attempt < c.MaxRetries; attempt++ {
		// Apply rate limiting
		if err := c.RateLimiter.Wait(ctx); err != nil {
			return nil, 0, fmt.Errorf("YouTubeClient.makeRequest: rate limiter wait failed: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
		if err != nil {
			return nil, 0, fmt.Errorf("YouTubeClient.makeRequest: failed to create request: %w", err)
		}
		req.Header.Set("Authorization", "Bearer "+accessToken)

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

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			lastErr = err
			continue
		}

		lastStatus = resp.StatusCode

		// Success
		if resp.StatusCode == http.StatusOK {
			return body, resp.StatusCode, nil
		}

		// Auth errors - don't retry, return immediately
		if resp.StatusCode == http.StatusUnauthorized {
			return body, resp.StatusCode, fmt.Errorf("YouTubeClient.makeRequest: request failed with status 401: unauthorized")
		}

		// Quota/rate limit errors - don't retry
		if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
			return body, resp.StatusCode, nil
		}

		// Other errors - retry with exponential backoff
		lastErr = fmt.Errorf("YouTubeClient.makeRequest: request failed with status %d", resp.StatusCode)
		if attempt < c.MaxRetries-1 {
			if sleepErr := c.exponentialBackoff(ctx, attempt); sleepErr != nil {
				return nil, 0, sleepErr
			}
		}
	}

	return nil, lastStatus, lastErr
}

// exponentialBackoff sleeps for an exponentially increasing duration.
func (c *YouTubeClient) exponentialBackoff(ctx context.Context, attempt int) error {
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

// FetchChannels fetches channel metadata using the YouTube Data API.
func (c *YouTubeClient) FetchChannels(ctx context.Context, accessToken string) (*YouTubeChannelResponse, error) {
	params := url.Values{}
	params.Set("part", "id,snippet,statistics,brandingSettings,contentDetails")
	params.Set("mine", "true")
	params.Set("maxResults", "50")

	reqURL := YouTubeDataAPIURL + "channels?" + params.Encode()

	body, status, err := c.makeRequest(ctx, reqURL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchChannels: failed to fetch channels: %w", err)
	}

	if status != http.StatusOK {
		if status == http.StatusForbidden || status == http.StatusTooManyRequests {
			return nil, fmt.Errorf("YouTubeClient.FetchChannels: youtube API quota/rate limit exceeded (status %d)", status)
		}
		return nil, fmt.Errorf("YouTubeClient.FetchChannels: youtube API error (status %d): %s", status, string(body))
	}

	var resp YouTubeChannelResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchChannels: failed to parse channels response: %w", err)
	}

	return &resp, nil
}

// YouTubePlaylistItemsResponse represents the playlistItems.list API response.
type YouTubePlaylistItemsResponse struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	PageInfo      struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []YouTubePlaylistItem `json:"items"`
}

// YouTubePlaylistItem represents a single item from the playlistItems.list API.
type YouTubePlaylistItem struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet struct {
		PublishedAt string `json:"publishedAt"`
		ChannelID   string `json:"channelId"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Thumbnails  struct {
			Default struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"default"`
			High struct {
				URL    string `json:"url"`
				Width  int    `json:"width"`
				Height int    `json:"height"`
			} `json:"high"`
		} `json:"thumbnails"`
		ResourceID struct {
			Kind    string `json:"kind"`
			VideoID string `json:"videoId"`
		} `json:"resourceId"`
	} `json:"snippet"`
	ContentDetails struct {
		VideoID          string `json:"videoId"`
		VideoPublishedAt string `json:"videoPublishedAt"`
	} `json:"contentDetails"`
}

// FetchVideos fetches uploaded videos using the channel's uploads playlist.
// Uses playlistItems.list instead of the deprecated activities.list API.
// Results are ordered newest-first; pagination stops when items are older than since.
func (c *YouTubeClient) FetchVideos(ctx context.Context, accessToken string, uploadsPlaylistID string, since time.Time) ([]YouTubeActivityItem, error) {
	if uploadsPlaylistID == "" {
		return nil, fmt.Errorf("YouTubeClient.FetchVideos: uploads playlist ID is empty")
	}

	var allItems []YouTubeActivityItem
	pageToken := ""

	for {
		params := url.Values{}
		params.Set("part", "snippet,contentDetails")
		params.Set("playlistId", uploadsPlaylistID)
		params.Set("maxResults", "50")
		if pageToken != "" {
			params.Set("pageToken", pageToken)
		}

		reqURL := YouTubeDataAPIURL + "playlistItems?" + params.Encode()
		body, status, err := c.makeRequest(ctx, reqURL, accessToken)
		if err != nil {
			return allItems, fmt.Errorf("YouTubeClient.FetchVideos: failed to fetch playlist items: %w", err)
		}

		if status != http.StatusOK {
			if status == http.StatusForbidden || status == http.StatusTooManyRequests {
				return allItems, fmt.Errorf("YouTubeClient.FetchVideos: youtube API quota/rate limit exceeded (status %d)", status)
			}
			return allItems, fmt.Errorf("YouTubeClient.FetchVideos: youtube API error (status %d): %s", status, string(body))
		}

		var resp YouTubePlaylistItemsResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return allItems, fmt.Errorf("YouTubeClient.FetchVideos: failed to parse playlist items response: %w", err)
		}

		reachedOlderItems := false
		for _, item := range resp.Items {
			videoID := item.ContentDetails.VideoID
			if videoID == "" {
				continue
			}

			publishedStr := item.ContentDetails.VideoPublishedAt
			if publishedStr == "" {
				publishedStr = item.Snippet.PublishedAt
			}

			publishedAt, err := time.Parse(time.RFC3339, publishedStr)
			if err != nil {
				continue
			}

			if publishedAt.Before(since) {
				reachedOlderItems = true
				break
			}

			// Map to YouTubeActivityItem for backward compatibility
			activity := YouTubeActivityItem{
				ID: item.ID,
			}
			activity.Snippet.PublishedAt = publishedStr
			activity.Snippet.ChannelID = item.Snippet.ChannelID
			activity.Snippet.Title = item.Snippet.Title
			activity.Snippet.Description = item.Snippet.Description
			activity.Snippet.Thumbnails.Default.URL = item.Snippet.Thumbnails.Default.URL
			activity.Snippet.Thumbnails.Default.Width = item.Snippet.Thumbnails.Default.Width
			activity.Snippet.Thumbnails.Default.Height = item.Snippet.Thumbnails.Default.Height
			activity.Snippet.Thumbnails.High.URL = item.Snippet.Thumbnails.High.URL
			activity.Snippet.Thumbnails.High.Width = item.Snippet.Thumbnails.High.Width
			activity.Snippet.Thumbnails.High.Height = item.Snippet.Thumbnails.High.Height
			activity.Snippet.Type = "upload"
			activity.ContentDetails.Upload.VideoID = videoID
			allItems = append(allItems, activity)
		}

		if reachedOlderItems || resp.NextPageToken == "" {
			break
		}
		pageToken = resp.NextPageToken
	}

	return allItems, nil
}

// FetchActivityInsights fetches daily activity insights from the Analytics API.
func (c *YouTubeClient) FetchActivityInsights(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error) {
	metrics := []string{
		"redViews",
		"views",
		"likes",
		"dislikes",
		"comments",
		"shares",
		"subscribersGained",
		"estimatedMinutesWatched",
		"estimatedRedMinutesWatched",
		"averageViewDuration",
		"averageViewPercentage",
	}

	params := url.Values{}
	params.Set("ids", "channel==MINE")
	params.Set("startDate", startDate.Format("2006-01-02"))
	params.Set("endDate", endDate.Format("2006-01-02"))
	params.Set("metrics", strings.Join(metrics, ","))
	params.Set("dimensions", "day")

	reqURL := YouTubeAnalyticsAPIURL + "reports?" + params.Encode()
	body, status, err := c.makeRequest(ctx, reqURL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchActivityInsights: failed to fetch activity insights: %w", err)
	}

	if status != http.StatusOK {
		if status == http.StatusForbidden || status == http.StatusTooManyRequests {
			return nil, fmt.Errorf("YouTubeClient.FetchActivityInsights: youtube API quota/rate limit exceeded (status %d)", status)
		}
		return nil, fmt.Errorf("YouTubeClient.FetchActivityInsights: youtube API error (status %d): %s", status, string(body))
	}

	var resp YouTubeAnalyticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchActivityInsights: failed to parse activity insights response: %w", err)
	}

	return &resp, nil
}

// FetchTrafficInsights fetches traffic source insights from the Analytics API.
func (c *YouTubeClient) FetchTrafficInsights(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error) {
	params := url.Values{}
	params.Set("ids", "channel==MINE")
	params.Set("startDate", startDate.Format("2006-01-02"))
	params.Set("endDate", endDate.Format("2006-01-02"))
	params.Set("metrics", "views,estimatedMinutesWatched")
	params.Set("dimensions", "day,insightTrafficSourceType")

	reqURL := YouTubeAnalyticsAPIURL + "reports?" + params.Encode()
	body, status, err := c.makeRequest(ctx, reqURL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchTrafficInsights: failed to fetch traffic insights: %w", err)
	}

	if status != http.StatusOK {
		if status == http.StatusForbidden || status == http.StatusTooManyRequests {
			return nil, fmt.Errorf("YouTubeClient.FetchTrafficInsights: youtube API quota/rate limit exceeded (status %d)", status)
		}
		return nil, fmt.Errorf("YouTubeClient.FetchTrafficInsights: youtube API error (status %d): %s", status, string(body))
	}

	var resp YouTubeAnalyticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchTrafficInsights: failed to parse traffic insights response: %w", err)
	}

	return &resp, nil
}

// FetchSharedInsights fetches sharing service insights from the Analytics API.
func (c *YouTubeClient) FetchSharedInsights(ctx context.Context, accessToken string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error) {
	params := url.Values{}
	params.Set("ids", "channel==MINE")
	params.Set("startDate", startDate.Format("2006-01-02"))
	params.Set("endDate", endDate.Format("2006-01-02"))
	params.Set("metrics", "shares")
	params.Set("dimensions", "sharingService")

	reqURL := YouTubeAnalyticsAPIURL + "reports?" + params.Encode()
	body, status, err := c.makeRequest(ctx, reqURL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchSharedInsights: failed to fetch shared insights: %w", err)
	}

	if status != http.StatusOK {
		if status == http.StatusForbidden || status == http.StatusTooManyRequests {
			return nil, fmt.Errorf("YouTubeClient.FetchSharedInsights: youtube API quota/rate limit exceeded (status %d)", status)
		}
		return nil, fmt.Errorf("YouTubeClient.FetchSharedInsights: youtube API error (status %d): %s", status, string(body))
	}

	var resp YouTubeAnalyticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchSharedInsights: failed to parse shared insights response: %w", err)
	}

	return &resp, nil
}

// FetchVideoInsights fetches insights for a specific video from the Analytics API.
func (c *YouTubeClient) FetchVideoInsights(ctx context.Context, accessToken, videoID string, startDate, endDate time.Time) (*YouTubeAnalyticsResponse, error) {
	metrics := []string{
		"redViews",
		"views",
		"likes",
		"dislikes",
		"comments",
		"shares",
		"subscribersGained",
		"estimatedMinutesWatched",
		"estimatedRedMinutesWatched",
		"averageViewDuration",
		"averageViewPercentage",
	}

	params := url.Values{}
	params.Set("ids", "channel==MINE")
	params.Set("startDate", startDate.Format("2006-01-02"))
	params.Set("endDate", endDate.Format("2006-01-02"))
	params.Set("metrics", strings.Join(metrics, ","))
	params.Set("filters", "video=="+videoID)

	reqURL := YouTubeAnalyticsAPIURL + "reports?" + params.Encode()
	body, status, err := c.makeRequest(ctx, reqURL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchVideoInsights: failed to fetch video insights: %w", err)
	}

	if status != http.StatusOK {
		if status == http.StatusForbidden || status == http.StatusTooManyRequests {
			return nil, fmt.Errorf("YouTubeClient.FetchVideoInsights: youtube API quota/rate limit exceeded (status %d)", status)
		}
		return nil, fmt.Errorf("YouTubeClient.FetchVideoInsights: youtube API error (status %d): %s", status, string(body))
	}

	var resp YouTubeAnalyticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchVideoInsights: failed to parse video insights response: %w", err)
	}

	return &resp, nil
}

// VideoAnalytics holds aggregated analytics for a single video.
type VideoAnalytics struct {
	VideoID                     string
	Views                       int64
	Likes                       int64
	Dislikes                    int64
	Comments                    int64
	Saved                       int64 // videosAddedToPlaylists
	SubscribersGained           int64
	EstimatedMinutesWatched     int64
	AverageViewDuration         int64
	AverageViewPercentage       float64
	Impressions                 int64
	ImpressionsClickThroughRate float64
}

// FetchAllVideosAnalytics fetches analytics for all videos in a single API call.
// Returns a map of videoID -> VideoAnalytics.
func (c *YouTubeClient) FetchAllVideosAnalytics(ctx context.Context, accessToken string, startDate, endDate time.Time) (map[string]*VideoAnalytics, error) {
	metrics := []string{
		"views",
		"likes",
		"dislikes",
		"comments",
		"videosAddedToPlaylists",
		"subscribersGained",
		"estimatedMinutesWatched",
		"averageViewDuration",
		"averageViewPercentage",
		"annotationImpressions",
		"annotationClickThroughRate",
	}

	params := url.Values{}
	params.Set("ids", "channel==MINE")
	params.Set("startDate", startDate.Format("2006-01-02"))
	params.Set("endDate", endDate.Format("2006-01-02"))
	params.Set("metrics", strings.Join(metrics, ","))
	params.Set("dimensions", "video")
	params.Set("maxResults", "200")
	params.Set("sort", "-views")

	reqURL := YouTubeAnalyticsAPIURL + "reports?" + params.Encode()
	body, status, err := c.makeRequest(ctx, reqURL, accessToken)
	if err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchAllVideosAnalytics: failed to fetch all videos analytics: %w", err)
	}

	if status != http.StatusOK {
		if status == http.StatusForbidden || status == http.StatusTooManyRequests {
			return nil, fmt.Errorf("YouTubeClient.FetchAllVideosAnalytics: youtube API quota/rate limit exceeded (status %d)", status)
		}
		return nil, fmt.Errorf("YouTubeClient.FetchAllVideosAnalytics: youtube API error (status %d): %s", status, string(body))
	}

	var resp YouTubeAnalyticsResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("YouTubeClient.FetchAllVideosAnalytics: failed to parse all videos analytics response: %w", err)
	}

	// Build column index map
	colIndex := make(map[string]int)
	for i, col := range resp.ColumnHeaders {
		colIndex[col.Name] = i
	}

	// Parse rows into map
	result := make(map[string]*VideoAnalytics)
	for _, row := range resp.Rows {
		if len(row) == 0 {
			continue
		}
		videoID, ok := row[colIndex["video"]].(string)
		if !ok || videoID == "" {
			continue
		}

		analytics := &VideoAnalytics{
			VideoID:                     videoID,
			Views:                       getInt64FromAnalyticsRow(row, colIndex, "views"),
			Likes:                       getInt64FromAnalyticsRow(row, colIndex, "likes"),
			Dislikes:                    getInt64FromAnalyticsRow(row, colIndex, "dislikes"),
			Comments:                    getInt64FromAnalyticsRow(row, colIndex, "comments"),
			Saved:                       getInt64FromAnalyticsRow(row, colIndex, "videosAddedToPlaylists"),
			SubscribersGained:           getInt64FromAnalyticsRow(row, colIndex, "subscribersGained"),
			EstimatedMinutesWatched:     getInt64FromAnalyticsRow(row, colIndex, "estimatedMinutesWatched"),
			AverageViewDuration:         getInt64FromAnalyticsRow(row, colIndex, "averageViewDuration"),
			AverageViewPercentage:       getFloat64FromAnalyticsRow(row, colIndex, "averageViewPercentage"),
			Impressions:                 getInt64FromAnalyticsRow(row, colIndex, "annotationImpressions"),
			ImpressionsClickThroughRate: getFloat64FromAnalyticsRow(row, colIndex, "annotationClickThroughRate"),
		}
		result[videoID] = analytics
	}

	return result, nil
}

// getInt64FromAnalyticsRow safely extracts an int64 value from an analytics row.
func getInt64FromAnalyticsRow(row []interface{}, colIndex map[string]int, colName string) int64 {
	idx, ok := colIndex[colName]
	if !ok || idx >= len(row) {
		return 0
	}
	switch v := row[idx].(type) {
	case float64:
		return int64(v)
	case int64:
		return v
	case int:
		return int64(v)
	}
	return 0
}

// getFloat64FromAnalyticsRow safely extracts a float64 value from an analytics row.
func getFloat64FromAnalyticsRow(row []interface{}, colIndex map[string]int, colName string) float64 {
	idx, ok := colIndex[colName]
	if !ok || idx >= len(row) {
		return 0
	}
	switch v := row[idx].(type) {
	case float64:
		return v
	case int64:
		return float64(v)
	case int:
		return float64(v)
	}
	return 0
}

// IsYouTubeShort checks if a video is a YouTube Short by making a HEAD request.
// Use IsShortByDuration as the primary detection method for better performance.
func (c *YouTubeClient) IsYouTubeShort(ctx context.Context, videoID string) bool {
	req, err := http.NewRequestWithContext(ctx, http.MethodHead, YouTubeShortsURL+videoID, nil)
	if err != nil {
		return false
	}

	// Use the shared client with connection pooling (no redirects)
	resp, err := c.ShortHTTPClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	// If we get a 200 OK, it's a Short
	return resp.StatusCode == http.StatusOK
}

// IsShortByDuration determines if a video is a Short based on its duration.
// YouTube Shorts are videos that are 60 seconds or less.
// This is much faster than making HTTP requests.
func IsShortByDuration(duration string) bool {
	seconds := ParseISO8601Duration(duration)
	return seconds > 0 && seconds <= maxShortDurationSeconds
}

// ParseISO8601Duration parses an ISO 8601 duration string (e.g., "PT1M30S", "PT45S")
// and returns the total duration in seconds.
func ParseISO8601Duration(duration string) int {
	if duration == "" {
		return 0
	}

	// Match ISO 8601 duration format: PT#H#M#S
	re := regexp.MustCompile(`PT(?:(\d+)H)?(?:(\d+)M)?(?:(\d+)S)?`)
	matches := re.FindStringSubmatch(duration)
	if matches == nil {
		return 0
	}

	var hours, minutes, seconds int
	if matches[1] != "" {
		hours, _ = strconv.Atoi(matches[1])
	}
	if matches[2] != "" {
		minutes, _ = strconv.Atoi(matches[2])
	}
	if matches[3] != "" {
		seconds, _ = strconv.Atoi(matches[3])
	}

	return hours*3600 + minutes*60 + seconds
}

// DetectMediaTypes determines the media type (video/short) for multiple videos in parallel.
// It uses duration-based detection as primary method, falling back to HTTP detection if duration is unavailable.
// Returns a map of videoID -> mediaType ("video" or "short").
func (c *YouTubeClient) DetectMediaTypes(ctx context.Context, videos []YouTubeVideoItem) map[string]string {
	result := make(map[string]string)
	var mu sync.Mutex

	// First pass: use duration-based detection (fast)
	var needHTTPCheck []string
	for _, video := range videos {
		if video.ContentDetails.Duration != "" {
			if IsShortByDuration(video.ContentDetails.Duration) {
				mu.Lock()
				result[video.ID] = "short"
				mu.Unlock()
			} else {
				mu.Lock()
				result[video.ID] = "video"
				mu.Unlock()
			}
		} else {
			// No duration available, need HTTP check
			needHTTPCheck = append(needHTTPCheck, video.ID)
		}
	}

	// Second pass: parallel HTTP detection for videos without duration
	if len(needHTTPCheck) > 0 {
		httpResults := c.DetectShortsParallel(ctx, needHTTPCheck)
		mu.Lock()
		for videoID, isShort := range httpResults {
			if isShort {
				result[videoID] = "short"
			} else {
				result[videoID] = "video"
			}
		}
		mu.Unlock()
	}

	return result
}

// DetectShortsParallel checks multiple videos for Short status in parallel.
// Uses a semaphore to limit concurrent HTTP requests.
func (c *YouTubeClient) DetectShortsParallel(ctx context.Context, videoIDs []string) map[string]bool {
	results := make(map[string]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Semaphore to limit concurrent requests
	sem := make(chan struct{}, maxParallelShortChecks)

	for _, id := range videoIDs {
		wg.Add(1)
		go func(videoID string) {
			defer wg.Done()

			// Acquire semaphore
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			case <-ctx.Done():
				return
			}

			isShort := c.IsYouTubeShort(ctx, videoID)

			mu.Lock()
			results[videoID] = isShort
			mu.Unlock()
		}(id)
	}

	wg.Wait()
	return results
}

// GenerateEmbedHTML generates the iframe embed HTML for a video.
func GenerateEmbedHTML(videoID string) string {
	return fmt.Sprintf(
		`<iframe width="560" height="315" src="https://www.youtube.com/embed/%s" frameborder="0" allowfullscreen></iframe>`,
		videoID,
	)
}

// YouTubeVideoListResponse represents the videos.list API response.
type YouTubeVideoListResponse struct {
	Kind          string `json:"kind"`
	Etag          string `json:"etag"`
	NextPageToken string `json:"nextPageToken,omitempty"`
	PageInfo      struct {
		TotalResults   int `json:"totalResults"`
		ResultsPerPage int `json:"resultsPerPage"`
	} `json:"pageInfo"`
	Items []YouTubeVideoItem `json:"items"`
}

// YouTubeVideoItem represents a single video from the videos.list API.
type YouTubeVideoItem struct {
	Kind    string `json:"kind"`
	Etag    string `json:"etag"`
	ID      string `json:"id"`
	Snippet struct {
		PublishedAt string `json:"publishedAt"`
		ChannelID   string `json:"channelId"`
		Title       string `json:"title"`
		Description string `json:"description"`
		Thumbnails  struct {
			Default struct {
				URL string `json:"url"`
			} `json:"default"`
			High struct {
				URL string `json:"url"`
			} `json:"high"`
		} `json:"thumbnails"`
	} `json:"snippet"`
	ContentDetails struct {
		Duration string `json:"duration"`
	} `json:"contentDetails"`
	Statistics struct {
		ViewCount     string `json:"viewCount"`
		LikeCount     string `json:"likeCount"`
		DislikeCount  string `json:"dislikeCount"`
		FavoriteCount string `json:"favoriteCount"`
		CommentCount  string `json:"commentCount"`
	} `json:"statistics"`
}

// FetchVideoDetails fetches video details including statistics for a list of video IDs.
// The YouTube API allows up to 50 video IDs per request.
func (c *YouTubeClient) FetchVideoDetails(ctx context.Context, accessToken string, videoIDs []string) ([]YouTubeVideoItem, error) {
	if len(videoIDs) == 0 {
		return nil, nil
	}

	var allItems []YouTubeVideoItem

	// Process in batches of 50 (API limit)
	for i := 0; i < len(videoIDs); i += 50 {
		end := i + 50
		if end > len(videoIDs) {
			end = len(videoIDs)
		}
		batch := videoIDs[i:end]

		params := url.Values{}
		params.Set("part", "id,snippet,contentDetails,statistics")
		params.Set("id", strings.Join(batch, ","))

		reqURL := YouTubeDataAPIURL + "videos?" + params.Encode()
		body, status, err := c.makeRequest(ctx, reqURL, accessToken)
		if err != nil {
			return allItems, fmt.Errorf("YouTubeClient.FetchVideoDetails: failed to fetch video details: %w", err)
		}

		if status != http.StatusOK {
			if status == http.StatusForbidden || status == http.StatusTooManyRequests {
				return allItems, fmt.Errorf("YouTubeClient.FetchVideoDetails: youtube API quota/rate limit exceeded (status %d)", status)
			}
			return allItems, fmt.Errorf("YouTubeClient.FetchVideoDetails: youtube API error (status %d): %s", status, string(body))
		}

		var resp YouTubeVideoListResponse
		if err := json.Unmarshal(body, &resp); err != nil {
			return allItems, fmt.Errorf("YouTubeClient.FetchVideoDetails: failed to parse video details response: %w", err)
		}

		allItems = append(allItems, resp.Items...)
	}

	return allItems, nil
}
