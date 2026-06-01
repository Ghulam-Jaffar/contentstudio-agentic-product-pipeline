package social

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

func TestNewTwitterClient(t *testing.T) {
	client := NewTwitterClient("consumer_key", "consumer_secret")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.httpClient == nil {
		t.Fatal("expected non-nil httpClient")
	}
	if client.consumerKey != "consumer_key" {
		t.Fatalf("expected consumerKey 'consumer_key', got '%s'", client.consumerKey)
	}
	if client.consumerSecret != "consumer_secret" {
		t.Fatalf("expected consumerSecret 'consumer_secret', got '%s'", client.consumerSecret)
	}
	if client.baseURL != twitterBaseURL {
		t.Fatalf("expected baseURL '%s', got '%s'", twitterBaseURL, client.baseURL)
	}
}

func TestTwitterClient_FetchUserTweets(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify OAuth header present
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":        "tweet_1",
					"text":      "Hello World",
					"author_id": "user_123",
					"public_metrics": map[string]interface{}{
						"retweet_count":    10,
						"reply_count":      5,
						"like_count":       100,
						"impression_count": 1000,
					},
				},
			},
			"meta": map[string]interface{}{
				"result_count": 1,
				"next_token":   "next_page_token",
			},
		})
	}))
	defer server.Close()

	client := NewTwitterClient("consumer_key", "consumer_secret")
	client.baseURL = server.URL + "/"

	resp, err := client.FetchUserTweets(context.Background(), "user_123", "oauth_token", "oauth_secret", 40, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 tweet, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "tweet_1" {
		t.Fatalf("expected tweet ID 'tweet_1', got '%s'", resp.Data[0].ID)
	}
	if resp.Meta == nil {
		t.Fatal("expected meta to be non-nil")
	}
	if resp.Meta.NextToken != "next_page_token" {
		t.Fatalf("expected next_token 'next_page_token', got '%s'", resp.Meta.NextToken)
	}
}

func TestTwitterClient_FetchUserTweets_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check pagination_token is present in query
		paginationToken := r.URL.Query().Get("pagination_token")
		if paginationToken != "page2" {
			t.Errorf("expected pagination_token 'page2', got '%s'", paginationToken)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{},
			"meta": map[string]interface{}{
				"result_count": 0,
			},
		})
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"

	resp, err := client.FetchUserTweets(context.Background(), "user_123", "token", "secret", 40, "page2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 0 {
		t.Fatalf("expected 0 tweets, got %d", len(resp.Data))
	}
}

func TestTwitterClient_FetchUserTweets_DefaultMaxResults(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		maxResults := r.URL.Query().Get("max_results")
		if maxResults != "40" {
			t.Errorf("expected max_results '40' (default), got '%s'", maxResults)
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{},
			"meta": map[string]interface{}{"result_count": 0},
		})
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchUserTweets(context.Background(), "user_123", "token", "secret", 0, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTwitterClient_FetchUserTweets_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"errors":[{"detail":"unauthorized"}]}`))
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchUserTweets(context.Background(), "user_123", "token", "secret", 40, "")
	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	if !IsExpectedTwitterError(err) {
		t.Fatalf("expected twitter auth error, got: %v", err)
	}
}

func TestTwitterClient_FetchUserTweets_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchUserTweets(context.Background(), "user_123", "token", "secret", 40, "")
	if err == nil {
		t.Fatal("expected error for 429 response")
	}
	if !IsExpectedTwitterError(err) {
		t.Fatalf("expected rate limit error, got: %v", err)
	}
}

func TestTwitterClient_FetchUserTweets_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"errors":[{"detail":"bad request"}]}`))
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchUserTweets(context.Background(), "user_123", "token", "secret", 40, "")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestTwitterClient_FetchUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		// Verify IDs parameter
		ids := r.URL.Query().Get("ids")
		if ids != "user_123" {
			t.Errorf("expected ids 'user_123', got '%s'", ids)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"id":                "user_123",
					"name":              "Test User",
					"username":          "testuser",
					"profile_image_url": "https://example.com/img.jpg",
					"public_metrics": map[string]interface{}{
						"followers_count": 10000,
						"following_count": 500,
						"tweet_count":     5000,
						"listed_count":    100,
						"like_count":      2000,
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"

	resp, err := client.FetchUserInfo(context.Background(), []string{"user_123"}, "token", "secret")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Fatalf("expected 1 user, got %d", len(resp.Data))
	}
	if resp.Data[0].ID != "user_123" {
		t.Fatalf("expected user ID 'user_123', got '%s'", resp.Data[0].ID)
	}
	if resp.Data[0].PublicMetrics.FollowersCount != 10000 {
		t.Fatalf("expected followers_count 10000, got %d", resp.Data[0].PublicMetrics.FollowersCount)
	}
}

func TestTwitterClient_FetchUserInfo_EmptyIDs(t *testing.T) {
	client := NewTwitterClient("key", "secret")
	_, err := client.FetchUserInfo(context.Background(), []string{}, "token", "secret")
	if err == nil {
		t.Fatal("expected error for empty user IDs")
	}
}

func TestTwitterClient_FetchUserInfo_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchUserInfo(context.Background(), []string{"user_123"}, "token", "secret")
	if err == nil {
		t.Fatal("expected error for unauthorized response")
	}
}

func TestTwitterClient_FetchUserInfo_RateLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchUserInfo(context.Background(), []string{"user_123"}, "token", "secret")
	if err == nil {
		t.Fatal("expected error for rate limit response")
	}
}

func TestSignRequest(t *testing.T) {
	client := NewTwitterClient("consumer_key", "consumer_secret")

	req, _ := http.NewRequest("GET", "https://api.twitter.com/2/users/123/tweets?max_results=10", nil)
	client.signRequest(req, "oauth_token", "oauth_token_secret")

	auth := req.Header.Get("Authorization")
	if auth == "" {
		t.Fatal("expected Authorization header to be set")
	}
	if !containsSubstring(auth, "OAuth") {
		t.Fatal("expected Authorization header to start with 'OAuth'")
	}
	if !containsSubstring(auth, "oauth_consumer_key") {
		t.Fatal("expected oauth_consumer_key in header")
	}
	if !containsSubstring(auth, "oauth_signature") {
		t.Fatal("expected oauth_signature in header")
	}
	if !containsSubstring(auth, "oauth_token") {
		t.Fatal("expected oauth_token in header")
	}
}

func TestGenerateNonce(t *testing.T) {
	n1 := generateNonce()
	n2 := generateNonce()
	if n1 == "" {
		t.Fatal("nonce should not be empty")
	}
	if n1 == n2 {
		t.Fatal("nonces should be unique")
	}
}

func TestIsExpectedTwitterError(t *testing.T) {
	tests := []struct {
		err      error
		expected bool
	}{
		{nil, false},
		{fmt.Errorf("twitter api rate limited (429)"), true},
		{fmt.Errorf("twitter api unauthorized (401)"), true},
		{fmt.Errorf("some other error"), false},
		{fmt.Errorf("contains 429 rate limited"), true},
	}

	for _, tt := range tests {
		result := IsExpectedTwitterError(tt.err)
		if result != tt.expected {
			t.Errorf("IsExpectedTwitterError(%v) = %v, want %v", tt.err, result, tt.expected)
		}
	}
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// ==================== Logging Contract Tests ====================

// TestLoggingContract_TwitterClient_WarnLevelOnly verifies the Twitter client
// only logs at Warn level (never Error) for all error paths.
func TestLoggingContract_TwitterClient_WarnLevelOnly(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Unauthorized", "code": 401},
			},
		})
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"
	client.log = log

	_, err := client.FetchUserInfo(context.Background(), []string{"testuser"}, "token", "token_secret")
	if err == nil {
		t.Fatal("expected error for API error response")
	}

	output := buf.String()
	if strings.Contains(output, "ERR") {
		t.Fatalf("Twitter client should NOT produce ERR-level logs, got: %s", output)
	}
	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
	}
}

// TestLoggingContract_TwitterClient_NoCaptureException verifies no error path
// in the Twitter client calls CaptureException.
func TestLoggingContract_TwitterClient_NoCaptureException(t *testing.T) {
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	log, _ := logger.NewTestLoggerWithHook()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"errors":[{"message":"Forbidden"}]}`))
	}))
	defer server.Close()

	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"
	client.log = log

	_, _ = client.FetchUserInfo(context.Background(), []string{"testuser"}, "token", "token_secret")

	// Context cancelled error path
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _ = client.FetchUserInfo(ctx, []string{"testuser2"}, "token", "token_secret")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls across all error paths, got %d", len(*captureRecords))
	}
}

// TestLoggingContract_TwitterClient_ErrorsReturnedToCaller verifies the Twitter
// client returns errors to callers rather than swallowing them.
func TestLoggingContract_TwitterClient_ErrorsReturnedToCaller(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"errors":[{"message":"Internal Server Error"}]}`))
	}))
	defer server.Close()

	log, _ := logger.NewTestLoggerWithHook()
	client := NewTwitterClient("key", "secret")
	client.baseURL = server.URL + "/"
	client.log = log

	_, err := client.FetchUserInfo(context.Background(), []string{"testuser"}, "token", "token_secret")
	if err == nil {
		t.Fatal("FetchUserInfo: expected error to be returned to caller, got nil")
	}
}
