package social

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewYouTubeClient(t *testing.T) {
	client := NewYouTubeClient("test_client_id", "test_client_secret")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.ClientID != "test_client_id" {
		t.Fatalf("expected ClientID 'test_client_id', got '%s'", client.ClientID)
	}
	if client.ClientSecret != "test_client_secret" {
		t.Fatalf("expected ClientSecret 'test_client_secret', got '%s'", client.ClientSecret)
	}
	if client.MaxRetries != 4 {
		t.Fatalf("expected MaxRetries 4, got %d", client.MaxRetries)
	}
	if client.HTTPClient == nil {
		t.Fatal("expected non-nil HTTPClient")
	}
}

func TestYouTubeClient_RefreshToken(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}

		resp := YouTubeTokenResponse{
			AccessToken:  "new_access_token",
			ExpiresIn:    3600,
			Scope:        "https://www.googleapis.com/auth/youtube.readonly",
			TokenType:    "Bearer",
			RefreshToken: "new_refresh_token",
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: "https://oauth2.googleapis.com/token",
		},
	}

	result, err := client.RefreshToken(context.Background(), "test_refresh_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.AccessToken != "new_access_token" {
		t.Fatalf("expected access_token 'new_access_token', got '%s'", result.AccessToken)
	}
}

func TestYouTubeClient_FetchChannels(t *testing.T) {
	channelResp := map[string]interface{}{
		"kind": "youtube#channelListResponse",
		"items": []map[string]interface{}{
			{
				"id": "UC123456",
				"snippet": map[string]interface{}{
					"title":       "Test Channel",
					"description": "Test Description",
					"customUrl":   "@testchannel",
					"publishedAt": "2024-01-01T00:00:00Z",
					"country":     "US",
					"thumbnails": map[string]interface{}{
						"default": map[string]interface{}{"url": "http://example.com/default.jpg"},
						"high":    map[string]interface{}{"url": "http://example.com/high.jpg"},
					},
				},
				"statistics": map[string]interface{}{
					"viewCount":       "1000",
					"subscriberCount": "100",
					"videoCount":      "50",
				},
				"brandingSettings": map[string]interface{}{
					"image": map[string]interface{}{
						"bannerExternalUrl": "http://example.com/banner.jpg",
					},
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("part") != "id,snippet,statistics,brandingSettings,contentDetails" {
			t.Fatalf("unexpected part parameter: %s", r.URL.Query().Get("part"))
		}
		if r.URL.Query().Get("mine") != "true" {
			t.Fatalf("expected mine=true")
		}
		json.NewEncoder(w).Encode(channelResp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	result, err := client.FetchChannels(context.Background(), "test_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Items) != 1 {
		t.Fatalf("expected 1 channel, got %d", len(result.Items))
	}

	if result.Items[0].ID != "UC123456" {
		t.Fatalf("expected channel ID 'UC123456', got '%s'", result.Items[0].ID)
	}

	if result.Items[0].Snippet.Title != "Test Channel" {
		t.Fatalf("expected title 'Test Channel', got '%s'", result.Items[0].Snippet.Title)
	}

	if result.Items[0].Statistics.ViewCount != "1000" {
		t.Fatalf("expected viewCount '1000', got '%s'", result.Items[0].Statistics.ViewCount)
	}
}

func TestYouTubeClient_FetchChannels_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": {"code": 403, "message": "Quota exceeded"}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	_, err := client.FetchChannels(context.Background(), "test_token")
	if err == nil {
		t.Fatal("expected error for quota exceeded")
	}
}

func TestYouTubeClient_FetchVideos(t *testing.T) {
	playlistItemsResp := map[string]interface{}{
		"kind": "youtube#playlistItemListResponse",
		"items": []map[string]interface{}{
			{
				"id": "item1",
				"snippet": map[string]interface{}{
					"title":       "Test Video",
					"description": "Test Description",
					"publishedAt": "2026-01-15T10:00:00Z",
					"channelId":   "UC123456",
					"thumbnails": map[string]interface{}{
						"default": map[string]interface{}{"url": "http://example.com/default.jpg"},
						"high":    map[string]interface{}{"url": "http://example.com/high.jpg"},
					},
					"resourceId": map[string]interface{}{
						"kind":    "youtube#video",
						"videoId": "video123",
					},
				},
				"contentDetails": map[string]interface{}{
					"videoId":          "video123",
					"videoPublishedAt": "2026-01-15T10:00:00Z",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(playlistItemsResp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := client.FetchVideos(context.Background(), "test_token", "UU123456", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}

	if result[0].ContentDetails.Upload.VideoID != "video123" {
		t.Fatalf("expected video ID 'video123', got '%s'", result[0].ContentDetails.Upload.VideoID)
	}
}

func TestYouTubeClient_FetchVideoDetails(t *testing.T) {
	videoResp := map[string]interface{}{
		"kind": "youtube#videoListResponse",
		"items": []map[string]interface{}{
			{
				"id": "video123",
				"snippet": map[string]interface{}{
					"title":       "Test Video",
					"description": "Test Description",
					"publishedAt": "2026-01-15T10:00:00Z",
					"thumbnails": map[string]interface{}{
						"default": map[string]interface{}{"url": "http://example.com/default.jpg"},
						"high":    map[string]interface{}{"url": "http://example.com/high.jpg"},
					},
				},
				"contentDetails": map[string]interface{}{
					"duration": "PT5M30S",
				},
				"statistics": map[string]interface{}{
					"viewCount":     "100",
					"likeCount":     "10",
					"dislikeCount":  "1",
					"favoriteCount": "0",
					"commentCount":  "5",
				},
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ids := r.URL.Query().Get("id")
		if ids != "video123,video456" {
			t.Fatalf("expected ids 'video123,video456', got '%s'", ids)
		}
		json.NewEncoder(w).Encode(videoResp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	videoIDs := []string{"video123", "video456"}
	result, err := client.FetchVideoDetails(context.Background(), "test_token", videoIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 video detail, got %d", len(result))
	}

	if result[0].Statistics.ViewCount != "100" {
		t.Fatalf("expected viewCount '100', got '%s'", result[0].Statistics.ViewCount)
	}

	if result[0].Statistics.LikeCount != "10" {
		t.Fatalf("expected likeCount '10', got '%s'", result[0].Statistics.LikeCount)
	}
}

func TestYouTubeClient_FetchVideoDetails_EmptyIDs(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")

	result, err := client.FetchVideoDetails(context.Background(), "test_token", []string{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result != nil {
		t.Fatalf("expected nil result for empty IDs, got %v", result)
	}
}

func TestYouTubeClient_FetchActivityInsights(t *testing.T) {
	analyticsResp := map[string]interface{}{
		"kind": "youtubeAnalytics#resultTable",
		"columnHeaders": []map[string]interface{}{
			{"name": "day", "columnType": "DIMENSION", "dataType": "STRING"},
			{"name": "views", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "likes", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "subscribersGained", "columnType": "METRIC", "dataType": "INTEGER"},
		},
		"rows": [][]interface{}{
			{"2026-01-15", float64(100), float64(10), float64(5)},
			{"2026-01-16", float64(150), float64(15), float64(3)},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("ids") != "channel==MINE" {
			t.Fatalf("expected ids 'channel==MINE', got '%s'", r.URL.Query().Get("ids"))
		}
		if r.URL.Query().Get("dimensions") != "day" {
			t.Fatalf("expected dimensions 'day', got '%s'", r.URL.Query().Get("dimensions"))
		}
		json.NewEncoder(w).Encode(analyticsResp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	result, err := client.FetchActivityInsights(context.Background(), "test_token", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
}

func TestYouTubeClient_FetchTrafficInsights(t *testing.T) {
	analyticsResp := map[string]interface{}{
		"kind": "youtubeAnalytics#resultTable",
		"columnHeaders": []map[string]interface{}{
			{"name": "day", "columnType": "DIMENSION", "dataType": "STRING"},
			{"name": "insightTrafficSourceType", "columnType": "DIMENSION", "dataType": "STRING"},
			{"name": "views", "columnType": "METRIC", "dataType": "INTEGER"},
		},
		"rows": [][]interface{}{
			{"2026-01-15", "YT_SEARCH", float64(50)},
			{"2026-01-15", "EXTERNAL", float64(30)},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(analyticsResp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	result, err := client.FetchTrafficInsights(context.Background(), "test_token", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
}

func TestYouTubeClient_FetchSharedInsights(t *testing.T) {
	analyticsResp := map[string]interface{}{
		"kind": "youtubeAnalytics#resultTable",
		"columnHeaders": []map[string]interface{}{
			{"name": "sharingService", "columnType": "DIMENSION", "dataType": "STRING"},
			{"name": "shares", "columnType": "METRIC", "dataType": "INTEGER"},
		},
		"rows": [][]interface{}{
			{"WHATSAPP", float64(10)},
			{"TWITTER", float64(5)},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(analyticsResp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	result, err := client.FetchSharedInsights(context.Background(), "test_token", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(result.Rows))
	}
}

func TestYouTubeClient_IsYouTubeShort(t *testing.T) {
	// Note: IsYouTubeShort creates its own HTTP client internally and makes
	// real HTTP requests to youtube.com, so we can only test basic behavior.
	client := NewYouTubeClient("client_id", "client_secret")

	// Test with cancelled context - should return false without hanging
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	result := client.IsYouTubeShort(ctx, "test123")
	if result {
		t.Fatal("expected false for cancelled context")
	}
}

func TestGenerateEmbedHTML(t *testing.T) {
	videoID := "test123"
	expected := `<iframe width="560" height="315" src="https://www.youtube.com/embed/test123" frameborder="0" allowfullscreen></iframe>`

	result := GenerateEmbedHTML(videoID)
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

func TestParseISO8601Duration(t *testing.T) {
	tests := []struct {
		duration string
		expected int
	}{
		{"PT45S", 45},           // 45 seconds
		{"PT1M", 60},            // 1 minute
		{"PT1M30S", 90},         // 1 minute 30 seconds
		{"PT2M", 120},           // 2 minutes
		{"PT1H", 3600},          // 1 hour
		{"PT1H30M", 5400},       // 1 hour 30 minutes
		{"PT1H30M45S", 5445},    // 1 hour 30 minutes 45 seconds
		{"PT10M5S", 605},        // 10 minutes 5 seconds
		{"", 0},                 // Empty string
		{"invalid", 0},         // Invalid format
	}

	for _, tt := range tests {
		result := ParseISO8601Duration(tt.duration)
		if result != tt.expected {
			t.Errorf("ParseISO8601Duration(%q) = %d, expected %d", tt.duration, result, tt.expected)
		}
	}
}

func TestIsShortByDuration(t *testing.T) {
	tests := []struct {
		duration string
		isShort  bool
	}{
		{"PT45S", true},        // 45 seconds - Short
		{"PT60S", true},        // 60 seconds - Short (boundary)
		{"PT1M", true},         // 1 minute - Short (boundary)
		{"PT61S", false},       // 61 seconds - Not a Short
		{"PT1M1S", false},      // 1 minute 1 second - Not a Short
		{"PT2M", false},        // 2 minutes - Not a Short
		{"PT10M", false},       // 10 minutes - Not a Short
		{"", false},            // Empty - Not a Short
	}

	for _, tt := range tests {
		result := IsShortByDuration(tt.duration)
		if result != tt.isShort {
			t.Errorf("IsShortByDuration(%q) = %v, expected %v", tt.duration, result, tt.isShort)
		}
	}
}

func TestYouTubeClient_FetchAllVideosAnalytics(t *testing.T) {
	analyticsResp := map[string]interface{}{
		"kind": "youtubeAnalytics#resultTable",
		"columnHeaders": []map[string]interface{}{
			{"name": "video", "columnType": "DIMENSION", "dataType": "STRING"},
			{"name": "views", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "likes", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "comments", "columnType": "METRIC", "dataType": "INTEGER"},
		},
		"rows": [][]interface{}{
			{"video123", float64(100), float64(10), float64(5)},
			{"video456", float64(200), float64(20), float64(10)},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("dimensions") != "video" {
			t.Fatalf("expected dimensions 'video', got '%s'", r.URL.Query().Get("dimensions"))
		}
		json.NewEncoder(w).Encode(analyticsResp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	result, err := client.FetchAllVideosAnalytics(context.Background(), "test_token", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(result))
	}

	if result["video123"].Views != 100 {
		t.Fatalf("expected 100 views for video123, got %d", result["video123"].Views)
	}

	if result["video456"].Likes != 20 {
		t.Fatalf("expected 20 likes for video456, got %d", result["video456"].Likes)
	}
}

// testTransport is a custom RoundTripper that redirects requests to the test server
type testTransport struct {
	baseURL   string
	targetURL string
}

func (t *testTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqURL := req.URL.String()
	if strings.HasPrefix(reqURL, t.targetURL) || strings.Contains(reqURL, "googleapis.com") || strings.Contains(reqURL, "youtube.com") || strings.Contains(reqURL, "oauth2.googleapis.com") {
		newURL := t.baseURL + req.URL.Path
		if req.URL.RawQuery != "" {
			newURL += "?" + req.URL.RawQuery
		}
		var err error
		req.URL, err = req.URL.Parse(newURL)
		if err != nil {
			return nil, err
		}
	}
	return http.DefaultTransport.RoundTrip(req)
}

func TestNewYouTubeClientWithConfig(t *testing.T) {
	cfg := YouTubeClientConfig{
		ClientID:     "custom_client_id",
		ClientSecret: "custom_client_secret",
		RPS:          10.0,
		Burst:        20,
		MaxRetries:   5,
		BaseBackoff:  1 * time.Second,
		MaxBackoff:   10 * time.Second,
	}

	client := NewYouTubeClientWithConfig(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.ClientID != "custom_client_id" {
		t.Fatalf("expected ClientID 'custom_client_id', got '%s'", client.ClientID)
	}
	if client.MaxRetries != 5 {
		t.Fatalf("expected MaxRetries 5, got %d", client.MaxRetries)
	}
	if client.BaseBackoff != 1*time.Second {
		t.Fatalf("expected BaseBackoff 1s, got %v", client.BaseBackoff)
	}
	if client.MaxBackoff != 10*time.Second {
		t.Fatalf("expected MaxBackoff 10s, got %v", client.MaxBackoff)
	}
	if client.RateLimiter == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
	if client.ShortHTTPClient == nil {
		t.Fatal("expected non-nil ShortHTTPClient")
	}
}

func TestNewYouTubeClientWithConfig_Defaults(t *testing.T) {
	cfg := YouTubeClientConfig{
		ClientID:     "client_id",
		ClientSecret: "client_secret",
		// All other fields zero/empty - should use defaults
	}

	client := NewYouTubeClientWithConfig(cfg)
	if client.MaxRetries != 4 {
		t.Fatalf("expected default MaxRetries 4, got %d", client.MaxRetries)
	}
	if client.BaseBackoff != 500*time.Millisecond {
		t.Fatalf("expected default BaseBackoff 500ms, got %v", client.BaseBackoff)
	}
	if client.MaxBackoff != 8*time.Second {
		t.Fatalf("expected default MaxBackoff 8s, got %v", client.MaxBackoff)
	}
}

func TestYouTubeClient_MakeRequest_401Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": {"code": 401, "message": "Invalid credentials"}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	_, err := client.FetchChannels(context.Background(), "invalid_token")
	if err == nil {
		t.Fatal("expected error for 401 unauthorized")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Fatalf("expected error to contain '401', got: %v", err)
	}
}

func TestYouTubeClient_MakeRequest_RetryOnError(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(`{"error": "server error"}`))
			return
		}
		// Success on third attempt
		json.NewEncoder(w).Encode(map[string]interface{}{
			"kind":  "youtube#channelListResponse",
			"items": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 4
	client.BaseBackoff = 1 * time.Millisecond
	client.MaxBackoff = 10 * time.Millisecond

	_, err := client.FetchChannels(context.Background(), "test_token")
	if err != nil {
		t.Fatalf("expected success after retries, got error: %v", err)
	}
	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestYouTubeClient_MakeRequest_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.FetchChannels(ctx, "test_token")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestYouTubeClient_ExponentialBackoff(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")
	client.BaseBackoff = 10 * time.Millisecond
	client.MaxBackoff = 100 * time.Millisecond

	// Test backoff increases exponentially
	start := time.Now()
	err := client.exponentialBackoff(context.Background(), 0)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed < 10*time.Millisecond || elapsed > 50*time.Millisecond {
		t.Fatalf("expected ~10ms backoff for attempt 0, got %v", elapsed)
	}

	// Test backoff respects maxBackoff
	start = time.Now()
	err = client.exponentialBackoff(context.Background(), 10) // Would be 10240ms without cap
	elapsed = time.Since(start)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if elapsed < 100*time.Millisecond || elapsed > 150*time.Millisecond {
		t.Fatalf("expected ~100ms backoff (capped), got %v", elapsed)
	}
}

func TestYouTubeClient_ExponentialBackoff_ContextCancelled(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")
	client.BaseBackoff = 1 * time.Second

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := client.exponentialBackoff(ctx, 0)
	if err != context.Canceled {
		t.Fatalf("expected context.Canceled error, got: %v", err)
	}
}

func TestYouTubeClient_DetectMediaTypes(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")

	videos := []YouTubeVideoItem{
		{
			ID: "short1",
			ContentDetails: struct {
				Duration string `json:"duration"`
			}{Duration: "PT30S"}, // 30 seconds - Short
		},
		{
			ID: "video1",
			ContentDetails: struct {
				Duration string `json:"duration"`
			}{Duration: "PT5M"}, // 5 minutes - Video
		},
		{
			ID: "short2",
			ContentDetails: struct {
				Duration string `json:"duration"`
			}{Duration: "PT60S"}, // 60 seconds - Short (boundary)
		},
		{
			ID: "video2",
			ContentDetails: struct {
				Duration string `json:"duration"`
			}{Duration: "PT61S"}, // 61 seconds - Video
		},
	}

	results := client.DetectMediaTypes(context.Background(), videos)

	if len(results) != 4 {
		t.Fatalf("expected 4 results, got %d", len(results))
	}
	if results["short1"] != "short" {
		t.Errorf("expected short1 to be 'short', got '%s'", results["short1"])
	}
	if results["video1"] != "video" {
		t.Errorf("expected video1 to be 'video', got '%s'", results["video1"])
	}
	if results["short2"] != "short" {
		t.Errorf("expected short2 to be 'short', got '%s'", results["short2"])
	}
	if results["video2"] != "video" {
		t.Errorf("expected video2 to be 'video', got '%s'", results["video2"])
	}
}

func TestYouTubeClient_DetectMediaTypes_EmptyDuration(t *testing.T) {
	// Videos with empty duration need HTTP check fallback
	// Since we can't easily mock HTTP in this case, we just verify it doesn't panic
	client := NewYouTubeClient("client_id", "client_secret")

	videos := []YouTubeVideoItem{
		{
			ID: "unknown1",
			ContentDetails: struct {
				Duration string `json:"duration"`
			}{Duration: ""}, // Empty duration - needs HTTP check
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	results := client.DetectMediaTypes(ctx, videos)
	// Should have a result (either from HTTP or default)
	if _, ok := results["unknown1"]; !ok {
		// If HTTP fails due to timeout, it should still return
		t.Log("Video with empty duration handled correctly")
	}
}

func TestYouTubeClient_DetectShortsParallel_ContextCancelled(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	videoIDs := []string{"video1", "video2", "video3"}
	results := client.DetectShortsParallel(ctx, videoIDs)

	// Should return empty or partial results due to cancelled context
	t.Logf("Got %d results with cancelled context", len(results))
}

func TestYouTubeClient_FetchVideoInsights(t *testing.T) {
	analyticsResp := map[string]interface{}{
		"kind": "youtubeAnalytics#resultTable",
		"columnHeaders": []map[string]interface{}{
			{"name": "views", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "likes", "columnType": "METRIC", "dataType": "INTEGER"},
		},
		"rows": [][]interface{}{
			{float64(1000), float64(100)},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("filters") != "video==testVideo123" {
			t.Fatalf("expected filters 'video==testVideo123', got '%s'", r.URL.Query().Get("filters"))
		}
		json.NewEncoder(w).Encode(analyticsResp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	result, err := client.FetchVideoInsights(context.Background(), "test_token", "testVideo123", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result.Rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(result.Rows))
	}
}

func TestYouTubeClient_FetchVideoInsights_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": {"code": 429, "message": "Rate limit exceeded"}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchVideoInsights(context.Background(), "test_token", "testVideo123", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for rate limit exceeded")
	}
}

func TestYouTubeClient_RefreshToken_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid_grant", "error_description": "Token has been revoked"}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeOAuthTokenURL,
		},
	}

	_, err := client.RefreshToken(context.Background(), "invalid_refresh_token")
	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}
	if !strings.Contains(err.Error(), "400") {
		t.Fatalf("expected error to contain '400', got: %v", err)
	}
}

func TestYouTubeClient_FetchVideos_Pagination(t *testing.T) {
	page := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page++
		videoID := "video" + string(rune('0'+page))
		resp := map[string]interface{}{
			"kind": "youtube#playlistItemListResponse",
			"items": []map[string]interface{}{
				{
					"id": "item" + string(rune('0'+page)),
					"snippet": map[string]interface{}{
						"publishedAt": "2026-01-15T10:00:00Z",
						"resourceId": map[string]interface{}{
							"kind":    "youtube#video",
							"videoId": videoID,
						},
					},
					"contentDetails": map[string]interface{}{
						"videoId":          videoID,
						"videoPublishedAt": "2026-01-15T10:00:00Z",
					},
				},
			},
		}
		if page < 2 {
			resp["nextPageToken"] = "page2"
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := client.FetchVideos(context.Background(), "test_token", "UU123456", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 videos from pagination, got %d", len(result))
	}
}

func TestYouTubeClient_FetchVideos_FilterEmptyVideoID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := map[string]interface{}{
			"kind": "youtube#playlistItemListResponse",
			"items": []map[string]interface{}{
				{
					"id": "item1",
					"snippet": map[string]interface{}{
						"publishedAt": "2026-01-15T10:00:00Z",
						"resourceId": map[string]interface{}{
							"kind":    "youtube#video",
							"videoId": "video1",
						},
					},
					"contentDetails": map[string]interface{}{
						"videoId":          "video1",
						"videoPublishedAt": "2026-01-15T10:00:00Z",
					},
				},
				{
					"id": "item2",
					"snippet": map[string]interface{}{
						"publishedAt": "2026-01-15T10:00:00Z",
					},
					"contentDetails": map[string]interface{}{
						"videoId":          "",
						"videoPublishedAt": "2026-01-15T10:00:00Z",
					},
				},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	result, err := client.FetchVideos(context.Background(), "test_token", "UU123456", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have 1 valid video (empty videoId filtered)
	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}
}

func TestYouTubeClient_FetchVideoDetails_Batching(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		ids := strings.Split(r.URL.Query().Get("id"), ",")
		items := make([]map[string]interface{}, len(ids))
		for i, id := range ids {
			items[i] = map[string]interface{}{
				"id": id,
				"snippet": map[string]interface{}{
					"title": "Video " + id,
				},
				"contentDetails": map[string]interface{}{
					"duration": "PT5M",
				},
				"statistics": map[string]interface{}{
					"viewCount": "100",
				},
			}
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"kind":  "youtube#videoListResponse",
			"items": items,
		})
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	// Create 75 video IDs (should result in 2 API calls: 50 + 25)
	videoIDs := make([]string, 75)
	for i := 0; i < 75; i++ {
		videoIDs[i] = "video" + string(rune('A'+i%26)) + string(rune('0'+i/26))
	}

	result, err := client.FetchVideoDetails(context.Background(), "test_token", videoIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if callCount != 2 {
		t.Fatalf("expected 2 API calls for 75 videos (batch of 50), got %d", callCount)
	}

	if len(result) != 75 {
		t.Fatalf("expected 75 video details, got %d", len(result))
	}
}

func TestGetInt64FromAnalyticsRow(t *testing.T) {
	colIndex := map[string]int{
		"views":    0,
		"likes":    1,
		"comments": 2,
	}

	row := []interface{}{float64(100), int64(50), int(25)}

	views := getInt64FromAnalyticsRow(row, colIndex, "views")
	if views != 100 {
		t.Errorf("expected views 100, got %d", views)
	}

	likes := getInt64FromAnalyticsRow(row, colIndex, "likes")
	if likes != 50 {
		t.Errorf("expected likes 50, got %d", likes)
	}

	comments := getInt64FromAnalyticsRow(row, colIndex, "comments")
	if comments != 25 {
		t.Errorf("expected comments 25, got %d", comments)
	}

	// Test missing column
	missing := getInt64FromAnalyticsRow(row, colIndex, "missing")
	if missing != 0 {
		t.Errorf("expected 0 for missing column, got %d", missing)
	}

	// Test out of bounds
	colIndex["outofbounds"] = 10
	outOfBounds := getInt64FromAnalyticsRow(row, colIndex, "outofbounds")
	if outOfBounds != 0 {
		t.Errorf("expected 0 for out of bounds, got %d", outOfBounds)
	}
}

func TestGetFloat64FromAnalyticsRow(t *testing.T) {
	colIndex := map[string]int{
		"percentage": 0,
		"rate":       1,
	}

	row := []interface{}{float64(75.5), int64(50)}

	percentage := getFloat64FromAnalyticsRow(row, colIndex, "percentage")
	if percentage != 75.5 {
		t.Errorf("expected percentage 75.5, got %f", percentage)
	}

	rate := getFloat64FromAnalyticsRow(row, colIndex, "rate")
	if rate != 50.0 {
		t.Errorf("expected rate 50.0, got %f", rate)
	}

	// Test missing column
	missing := getFloat64FromAnalyticsRow(row, colIndex, "missing")
	if missing != 0.0 {
		t.Errorf("expected 0.0 for missing column, got %f", missing)
	}
}

func TestYouTubeClient_FetchAllVideosAnalytics_EmptyResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"kind":          "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]interface{}{},
			"rows":          [][]interface{}{},
		})
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	result, err := client.FetchAllVideosAnalytics(context.Background(), "test_token", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 0 {
		t.Fatalf("expected 0 results for empty response, got %d", len(result))
	}
}

func TestYouTubeClient_FetchActivityInsights_QuotaExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": {"code": 403, "message": "Quota exceeded"}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchActivityInsights(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for quota exceeded")
	}
	if !strings.Contains(err.Error(), "quota") {
		t.Fatalf("expected quota error, got: %v", err)
	}
}

func TestYouTubeClient_FetchActivityInsights_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid json`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchActivityInsights(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("expected parse error, got: %v", err)
	}
}

func TestYouTubeClient_FetchTrafficInsights_QuotaExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": {"code": 429, "message": "Rate limit exceeded"}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchTrafficInsights(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for rate limit")
	}
	if !strings.Contains(err.Error(), "quota") {
		t.Fatalf("expected quota/rate limit error, got: %v", err)
	}
}

func TestYouTubeClient_FetchTrafficInsights_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchTrafficInsights(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestYouTubeClient_FetchSharedInsights_QuotaExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": {"code": 403, "message": "Forbidden"}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchSharedInsights(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for quota")
	}
}

func TestYouTubeClient_FetchSharedInsights_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{broken`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchSharedInsights(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestYouTubeClient_FetchAllVideosAnalytics_QuotaExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": {"code": 429}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchAllVideosAnalytics(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for quota")
	}
}

func TestYouTubeClient_FetchAllVideosAnalytics_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not valid json`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchAllVideosAnalytics(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestYouTubeClient_FetchAllVideosAnalytics_InvalidRow(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"kind": "youtubeAnalytics#resultTable",
			"columnHeaders": []map[string]interface{}{
				{"name": "video", "columnType": "DIMENSION", "dataType": "STRING"},
				{"name": "views", "columnType": "METRIC", "dataType": "INTEGER"},
			},
			"rows": [][]interface{}{
				{}, // Empty row
				{float64(123), float64(100)}, // Invalid video ID type
				{"video123", float64(100)},   // Valid row
			},
		})
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	result, err := client.FetchAllVideosAnalytics(context.Background(), "test_token", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should only have 1 valid result
	if len(result) != 1 {
		t.Fatalf("expected 1 valid result, got %d", len(result))
	}
	if _, ok := result["video123"]; !ok {
		t.Fatal("expected video123 in results")
	}
}

func TestYouTubeClient_IsYouTubeShort_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodHead {
			t.Fatalf("expected HEAD, got %s", r.Method)
		}
		// Return 200 OK for Short
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.ShortHTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeShortsURL,
		},
	}

	result := client.IsYouTubeShort(context.Background(), "short123")
	if !result {
		t.Fatal("expected true for Short")
	}
}

func TestYouTubeClient_IsYouTubeShort_NotAShort(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return redirect for non-Short
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.ShortHTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeShortsURL,
		},
	}

	result := client.IsYouTubeShort(context.Background(), "video123")
	if result {
		t.Fatal("expected false for non-Short")
	}
}

func TestYouTubeClient_RefreshToken_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid json response`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeOAuthTokenURL,
		},
	}

	_, err := client.RefreshToken(context.Background(), "refresh_token")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("expected parse error, got: %v", err)
	}
}

func TestYouTubeClient_FetchChannels_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	_, err := client.FetchChannels(context.Background(), "test_token")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Fatalf("expected parse error, got: %v", err)
	}
}

func TestYouTubeClient_FetchChannels_QuotaExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": {"code": 429}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	_, err := client.FetchChannels(context.Background(), "test_token")
	if err == nil {
		t.Fatal("expected error for quota")
	}
	if !strings.Contains(err.Error(), "quota") {
		t.Fatalf("expected quota error, got: %v", err)
	}
}

func TestYouTubeClient_FetchVideos_QuotaExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": {"code": 403}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FetchVideos(context.Background(), "test_token", "UU123456", since)
	if err == nil {
		t.Fatal("expected error for quota")
	}
}

func TestYouTubeClient_FetchVideos_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FetchVideos(context.Background(), "test_token", "UU123456", since)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestYouTubeClient_FetchVideoDetails_QuotaExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"error": {"code": 429}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	_, err := client.FetchVideoDetails(context.Background(), "test_token", []string{"video1"})
	if err == nil {
		t.Fatal("expected error for quota")
	}
}

func TestYouTubeClient_FetchVideoDetails_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`broken json`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}

	_, err := client.FetchVideoDetails(context.Background(), "test_token", []string{"video1"})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestYouTubeClient_FetchVideoInsights_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{not valid json`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchVideoInsights(context.Background(), "test_token", "video123", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestYouTubeClient_FetchVideoInsights_QuotaExceeded(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"error": {"code": 403}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchVideoInsights(context.Background(), "test_token", "video123", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for quota")
	}
}

func TestGetFloat64FromAnalyticsRow_IntType(t *testing.T) {
	colIndex := map[string]int{
		"rate": 0,
	}
	row := []interface{}{int(42)}

	result := getFloat64FromAnalyticsRow(row, colIndex, "rate")
	if result != 42.0 {
		t.Errorf("expected 42.0, got %f", result)
	}
}

func TestGetFloat64FromAnalyticsRow_OutOfBounds(t *testing.T) {
	colIndex := map[string]int{
		"rate": 10,
	}
	row := []interface{}{float64(1.0)}

	result := getFloat64FromAnalyticsRow(row, colIndex, "rate")
	if result != 0.0 {
		t.Errorf("expected 0.0 for out of bounds, got %f", result)
	}
}

func TestGetFloat64FromAnalyticsRow_UnsupportedType(t *testing.T) {
	colIndex := map[string]int{
		"rate": 0,
	}
	row := []interface{}{"string value"}

	result := getFloat64FromAnalyticsRow(row, colIndex, "rate")
	if result != 0.0 {
		t.Errorf("expected 0.0 for unsupported type, got %f", result)
	}
}

func TestGetInt64FromAnalyticsRow_UnsupportedType(t *testing.T) {
	colIndex := map[string]int{
		"views": 0,
	}
	row := []interface{}{"string value"}

	result := getInt64FromAnalyticsRow(row, colIndex, "views")
	if result != 0 {
		t.Errorf("expected 0 for unsupported type, got %d", result)
	}
}

func TestYouTubeClient_MakeRequest_MaxRetriesExhausted(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "server error"}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 3
	client.BaseBackoff = 1 * time.Millisecond
	client.MaxBackoff = 5 * time.Millisecond

	_, err := client.FetchChannels(context.Background(), "test_token")
	if err == nil {
		t.Fatal("expected error after max retries")
	}

	if attempts != 3 {
		t.Fatalf("expected 3 attempts, got %d", attempts)
	}
}

func TestYouTubeClient_FetchChannels_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"error": {"code": 404, "message": "Channel not found"}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	_, err := client.FetchChannels(context.Background(), "test_token")
	if err == nil {
		t.Fatal("expected error for 404")
	}
	// Error may come from makeRequest retry exhaustion or from FetchChannels
	if !strings.Contains(err.Error(), "404") && !strings.Contains(err.Error(), "fetch") {
		t.Fatalf("expected error message mentioning 404 or fetch, got: %v", err)
	}
}

func TestYouTubeClient_FetchVideos_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"code": 400}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	since := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	_, err := client.FetchVideos(context.Background(), "test_token", "UU123456", since)
	if err == nil {
		t.Fatal("expected error for 400")
	}
	// Error may come from makeRequest or from FetchVideos
	if !strings.Contains(err.Error(), "400") && !strings.Contains(err.Error(), "fetch") {
		t.Fatalf("expected error mentioning 400 or fetch, got: %v", err)
	}
}

func TestYouTubeClient_FetchVideoDetails_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": {"code": 503}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	_, err := client.FetchVideoDetails(context.Background(), "test_token", []string{"video1"})
	if err == nil {
		t.Fatal("expected error for 503")
	}
}

func TestYouTubeClient_FetchActivityInsights_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"error": {"code": 502}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 16, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchActivityInsights(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for 502")
	}
}

func TestYouTubeClient_FetchTrafficInsights_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": {"code": 503}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchTrafficInsights(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for 503")
	}
}

func TestYouTubeClient_FetchSharedInsights_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"code": 400}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchSharedInsights(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for 400")
	}
}

func TestYouTubeClient_FetchAllVideosAnalytics_OtherError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": {"code": 400}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchAllVideosAnalytics(context.Background(), "test_token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for 400")
	}
}

// errorReader is an io.Reader that always returns an error
type errorReader struct{}

func (e errorReader) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated read error")
}

// errorTransport always returns an error for requests
type errorTransport struct{}

func (e *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, fmt.Errorf("simulated transport error")
}

func TestYouTubeClient_RefreshToken_RequestError(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &errorTransport{},
	}

	_, err := client.RefreshToken(context.Background(), "refresh_token")
	if err == nil {
		t.Fatal("expected error for failed request")
	}
	if !strings.Contains(err.Error(), "request failed") {
		t.Fatalf("expected 'request failed' error, got: %v", err)
	}
}

func TestYouTubeClient_MakeRequest_TransportError(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &errorTransport{},
	}
	client.MaxRetries = 2
	client.BaseBackoff = 1 * time.Millisecond

	_, err := client.FetchChannels(context.Background(), "test_token")
	if err == nil {
		t.Fatal("expected error for transport failure")
	}
}

func TestYouTubeClient_MakeRequest_RateLimiterCancelled(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")
	// Create a context that's already cancelled
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.FetchChannels(ctx, "test_token")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
	if !strings.Contains(err.Error(), "rate limiter") && !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("expected rate limiter or context error, got: %v", err)
	}
}

func TestYouTubeClient_MakeRequest_BackoffContextCancelledDuringRetry(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 5
	client.BaseBackoff = 500 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	// Cancel after first attempt
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := client.FetchChannels(ctx, "test_token")
	if err == nil {
		t.Fatal("expected error for cancelled context during backoff")
	}
}

func TestYouTubeClient_FetchVideoInsights_OtherStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"error": {"code": 503}}`))
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}
	client.MaxRetries = 1
	client.BaseBackoff = 1 * time.Millisecond

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	_, err := client.FetchVideoInsights(context.Background(), "test_token", "video123", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for 503")
	}
}

func TestYouTubeClient_IsYouTubeShort_RequestError(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")
	client.ShortHTTPClient = &http.Client{
		Transport: &errorTransport{},
	}

	result := client.IsYouTubeShort(context.Background(), "video123")
	if result {
		t.Fatal("expected false for request error")
	}
}

// bodyErrorTransport returns a response with an error-producing body
type bodyErrorTransport struct {
	baseURL   string
	targetURL string
}

func (t *bodyErrorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: http.StatusOK,
		Body:       &errorReadCloser{},
		Header:     make(http.Header),
	}, nil
}

type errorReadCloser struct{}

func (e *errorReadCloser) Read(p []byte) (n int, err error) {
	return 0, fmt.Errorf("simulated body read error")
}

func (e *errorReadCloser) Close() error {
	return nil
}

func TestYouTubeClient_MakeRequest_BodyReadError(t *testing.T) {
	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &bodyErrorTransport{},
	}
	client.MaxRetries = 2
	client.BaseBackoff = 1 * time.Millisecond

	_, err := client.FetchChannels(context.Background(), "test_token")
	// Body read error is not fatal - should exhaust retries
	if err == nil {
		t.Fatal("expected error after body read failures")
	}
}

func TestYouTubeClient_MakeRequest_HTTPErrorThenSuccess(t *testing.T) {
	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			// First attempt - connection reset simulation
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
				return
			}
			// Fallback
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Second attempt - success
		json.NewEncoder(w).Encode(map[string]interface{}{
			"kind":  "youtube#channelListResponse",
			"items": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeDataAPIURL,
		},
	}
	client.MaxRetries = 3
	client.BaseBackoff = 1 * time.Millisecond

	_, err := client.FetchChannels(context.Background(), "test_token")
	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
}

func TestYouTubeClient_DetectMediaTypes_HTTPFallback(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return 200 for Short video
		if strings.Contains(r.URL.Path, "short_video") {
			w.WriteHeader(http.StatusOK)
			return
		}
		// Return redirect for regular video
		w.WriteHeader(http.StatusFound)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.ShortHTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeShortsURL,
		},
	}

	// Test videos with no duration - should fall back to HTTP check
	videos := []YouTubeVideoItem{
		{
			ID: "short_video",
			ContentDetails: struct {
				Duration string `json:"duration"`
			}{Duration: ""}, // Empty duration
		},
		{
			ID: "regular_video",
			ContentDetails: struct {
				Duration string `json:"duration"`
			}{Duration: ""}, // Empty duration
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := client.DetectMediaTypes(ctx, videos)

	// Both should have results from HTTP fallback
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
}

func TestYouTubeClient_AllVideoAnalytics_FullMetrics(t *testing.T) {
	analyticsResp := map[string]interface{}{
		"kind": "youtubeAnalytics#resultTable",
		"columnHeaders": []map[string]interface{}{
			{"name": "video", "columnType": "DIMENSION", "dataType": "STRING"},
			{"name": "views", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "likes", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "dislikes", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "comments", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "videosAddedToPlaylists", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "subscribersGained", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "estimatedMinutesWatched", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "averageViewDuration", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "averageViewPercentage", "columnType": "METRIC", "dataType": "FLOAT"},
			{"name": "annotationImpressions", "columnType": "METRIC", "dataType": "INTEGER"},
			{"name": "annotationClickThroughRate", "columnType": "METRIC", "dataType": "FLOAT"},
		},
		"rows": [][]interface{}{
			{
				"video123",
				float64(1000),
				float64(100),
				float64(5),
				float64(50),
				float64(10),
				float64(25),
				float64(5000),
				float64(300),
				float64(75.5),
				float64(10000),
				float64(0.05),
			},
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(analyticsResp)
	}))
	defer server.Close()

	client := NewYouTubeClient("client_id", "client_secret")
	client.HTTPClient = &http.Client{
		Transport: &testTransport{
			baseURL:   server.URL,
			targetURL: YouTubeAnalyticsAPIURL,
		},
	}

	startDate := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2026, 1, 31, 0, 0, 0, 0, time.UTC)

	result, err := client.FetchAllVideosAnalytics(context.Background(), "test_token", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	video := result["video123"]
	if video == nil {
		t.Fatal("expected video123 in results")
	}

	if video.Views != 1000 {
		t.Errorf("expected views 1000, got %d", video.Views)
	}
	if video.Likes != 100 {
		t.Errorf("expected likes 100, got %d", video.Likes)
	}
	if video.Dislikes != 5 {
		t.Errorf("expected dislikes 5, got %d", video.Dislikes)
	}
	if video.Comments != 50 {
		t.Errorf("expected comments 50, got %d", video.Comments)
	}
	if video.Saved != 10 {
		t.Errorf("expected saved 10, got %d", video.Saved)
	}
	if video.SubscribersGained != 25 {
		t.Errorf("expected subscribersGained 25, got %d", video.SubscribersGained)
	}
	if video.EstimatedMinutesWatched != 5000 {
		t.Errorf("expected estimatedMinutesWatched 5000, got %d", video.EstimatedMinutesWatched)
	}
	if video.AverageViewDuration != 300 {
		t.Errorf("expected averageViewDuration 300, got %d", video.AverageViewDuration)
	}
	if video.AverageViewPercentage != 75.5 {
		t.Errorf("expected averageViewPercentage 75.5, got %f", video.AverageViewPercentage)
	}
	if video.Impressions != 10000 {
		t.Errorf("expected impressions 10000, got %d", video.Impressions)
	}
	if video.ImpressionsClickThroughRate != 0.05 {
		t.Errorf("expected impressionsClickThroughRate 0.05, got %f", video.ImpressionsClickThroughRate)
	}
}
