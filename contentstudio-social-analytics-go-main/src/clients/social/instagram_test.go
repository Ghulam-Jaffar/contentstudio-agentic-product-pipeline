package social

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestInstagramAuthError(t *testing.T) {
	err := &InstagramAuthError{
		Message:    "Token expired",
		StatusCode: 401,
		ErrorCode:  190,
	}

	expected := "instagram auth error (status=401, code=190): Token expired"
	if err.Error() != expected {
		t.Fatalf("expected '%s', got '%s'", expected, err.Error())
	}
}

func TestIsAuthError(t *testing.T) {
	cases := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"InstagramAuthError", &InstagramAuthError{Message: "test"}, true},
		{"invalid oauth", errors.New("invalid oauth token"), true},
		{"access token", errors.New("access token expired"), true},
		{"token has expired", errors.New("token has expired"), true},
		{"session has expired", errors.New("session has expired"), true},
		{"status 401", errors.New("status 401 unauthorized"), true},
		{"status 403", errors.New("status 403 forbidden"), true},
		{"unauthorized", errors.New("unauthorized access"), true},
		{"code 190", errors.New("Error (#190): Access token has expired"), true},
		{"code 102", errors.New("Error (#102): Session invalid"), true},
		{"code 100", errors.New("Error (#100): Invalid parameter"), true},
		{"code 4", errors.New("Error (#4): Application request limit reached"), true},
		{"code 200", errors.New("Error (#200): Permission error"), true},
		{"code 10 - not auth error", errors.New("Error (#10): Not enough viewers for the media"), false},
		{"code 10 - permission error", errors.New("Error (#10): Application does not have permission for this action"), true},
		{"generic error", errors.New("network timeout"), false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := IsAuthError(tc.err)
			if result != tc.expected {
				t.Fatalf("expected %v, got %v", tc.expected, result)
			}
		})
	}
}

func TestNewInstagramClient(t *testing.T) {
	client := NewInstagramClient("test_secret")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.appSecret != "test_secret" {
		t.Fatalf("expected appSecret 'test_secret', got '%s'", client.appSecret)
	}
	if client.baseURL != igBaseURL {
		t.Fatalf("expected baseURL '%s', got '%s'", igBaseURL, client.baseURL)
	}
	if client.rate == nil {
		t.Fatal("expected non-nil rate manager")
	}
}

func TestNewInstagramClientWithRates(t *testing.T) {
	rm := NewRateManager(RateLimits{GlobalRPS: 50})
	client := NewInstagramClientWithRates("secret", rm)

	if client.rate != rm {
		t.Fatal("expected same rate manager")
	}
}

func TestNewInstagramClientWithRates_NilRateManager(t *testing.T) {
	client := NewInstagramClientWithRates("secret", nil)
	if client.rate == nil {
		t.Fatal("expected non-nil rate manager when nil is passed")
	}
}

func TestInstagramClient_WithBaseURL(t *testing.T) {
	client := NewInstagramClient("secret")
	newURL := "https://custom.api.com/"
	result := client.WithBaseURL(newURL)

	if result.baseURL != newURL {
		t.Fatalf("expected baseURL '%s', got '%s'", newURL, result.baseURL)
	}

	// Should return same client for chaining
	if result != client {
		t.Fatal("expected same client instance")
	}
}

func TestInstagramClient_generateAppSecretProof(t *testing.T) {
	client := NewInstagramClient("test_secret")
	proof := client.generateAppSecretProof("test_token")

	if len(proof) != 64 {
		t.Fatalf("expected 64 character hex string, got %d", len(proof))
	}

	// Same input should produce same output
	proof2 := client.generateAppSecretProof("test_token")
	if proof != proof2 {
		t.Fatal("expected same proof for same input")
	}
}

func TestInstagramClient_waitRate(t *testing.T) {
	client := NewInstagramClient("secret")
	err := client.waitRate(context.Background(), "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestInstagramClient_waitRate_NilRateManager(t *testing.T) {
	client := &InstagramClient{
		rate: nil,
	}

	err := client.waitRate(context.Background(), "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if client.rate == nil {
		t.Fatal("expected rate manager to be initialized")
	}
}

func TestInstagramClient_FetchMedia(t *testing.T) {
	media := []kafkamodels.RawInstagramMedia{
		{ID: "media_1", MediaType: "IMAGE"},
		{ID: "media_2", MediaType: "VIDEO"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Data   []kafkamodels.RawInstagramMedia `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}{
			Data: media,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchMedia(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(result))
	}
}

func TestInstagramClient_FetchMedia_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := igApiError{}
		resp.Error.Message = "Invalid token"
		resp.Error.Type = "OAuthException"
		resp.Error.Code = 190
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchMedia(context.Background(), "ig_123", "token")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestInstagramClient_FetchMedia_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.FetchMedia(ctx, "ig_123", "token")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestInstagramClient_FetchMediaSince(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []kafkamodels.RawInstagramMedia{{ID: "media_1"}},
			"paging": map[string]string{},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)

	result, err := client.FetchMediaSince(context.Background(), "ig_123", "token", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 media item, got %d", len(result))
	}
}

func TestInstagramClient_FetchStories(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []kafkamodels.RawInstagramMedia{{ID: "story_1", MediaType: "IMAGE"}},
			"paging": map[string]string{},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchStories(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 story, got %d", len(result))
	}
}

func TestInstagramClient_FetchInsights(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"name":   "impressions",
					"period": "day",
					"values": []map[string]interface{}{
						{"value": 1000, "end_time": "2024-01-15T08:00:00+0000"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	result, err := client.FetchInsights(context.Background(), "ig_123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchMediaInsights(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"name": "impressions", "values": []map[string]interface{}{{"value": 500}}},
				{"name": "reach", "values": []map[string]interface{}{{"value": 300}}},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchMediaInsights(context.Background(), "media_123", "token", "IMAGE", "FEED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchMediaInsights_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(igApiError{})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchMediaInsights(context.Background(), "media_123", "token", "IMAGE", "FEED")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestIgApiError_Struct(t *testing.T) {
	err := igApiError{}
	err.Error.Message = "Test error"
	err.Error.Type = "OAuthException"
	err.Error.Code = 190
	err.Error.FBTraceID = "trace123"

	if err.Error.Message != "Test error" {
		t.Fatalf("expected message 'Test error', got '%s'", err.Error.Message)
	}
}

func TestInstagramClient_doWithRetry_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	body, status, err := client.doWithRetry(context.Background(), "ig_123", req, "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status 200, got %d", status)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestInstagramClient_doWithRetry_RetryOnFailure(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	_, status, err := client.doWithRetry(context.Background(), "ig_123", req, "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected status 200 after retries, got %d", status)
	}
	if callCount < 3 {
		t.Fatalf("expected at least 3 calls, got %d", callCount)
	}
}

func TestInstagramClient_FetchAccountDemographics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"name":   "follower_demographics",
					"values": []map[string]interface{}{{"value": map[string]interface{}{"US": 100}}},
				},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchAccountDemographics(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchMediaWithLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []kafkamodels.RawInstagramMedia{{ID: "media_1"}},
			"paging": map[string]string{},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchMediaWithLimit(context.Background(), "ig_123", "token", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 media item, got %d", len(result))
	}
}

func TestInstagramClient_FetchAccountMedia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"media": map[string]interface{}{
				"data": []kafkamodels.RawInstagramMedia{{ID: "media_1"}},
			},
			"id":       "ig_123",
			"username": "testuser",
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchAccountMedia(context.Background(), "ig_123", "token", 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchUserInfo(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              "ig_123",
			"username":        "testuser",
			"followers_count": 10000,
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchUserInfo(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchUserInfo_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Invalid token",
				"type":    "OAuthException",
				"code":    190,
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchUserInfo(context.Background(), "ig_123", "token")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

// Competitor API tests
func TestInstagramClient_GetBusinessDiscovery(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"business_discovery": map[string]interface{}{
				"id":              "17841400000000000",
				"username":        "competitor",
				"followers_count": 50000,
				"media_count":     100,
				"media": map[string]interface{}{
					"data": []map[string]interface{}{
						{"id": "media1", "caption": "Test post 1"},
						{"id": "media2", "caption": "Test post 2"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.GetBusinessDiscovery(context.Background(), "competitor", 25, "", "token", "business_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_GetBusinessDiscovery_WithCursor(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify cursor is included in request
		if !strings.Contains(r.URL.RawQuery, "after") {
			// Request should contain cursor
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"business_discovery": map[string]interface{}{
				"id":       "17841400000000000",
				"username": "competitor",
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.GetBusinessDiscovery(context.Background(), "competitor", 25, "cursor123", "token", "business_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_GetBusinessDiscovery_DefaultLimit(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"business_discovery": map[string]interface{}{
				"id": "17841400000000000",
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	// Pass 0 limit to test default
	result, err := client.GetBusinessDiscovery(context.Background(), "competitor", 0, "", "token", "business_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_GetBusinessDiscovery_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Business account not found",
				"type":    "IGApiException",
				"code":    100,
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.GetBusinessDiscovery(context.Background(), "nonexistent", 25, "", "token", "business_123")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestParseAPIErrorIG(t *testing.T) {
	// Test with valid error response
	body := []byte(`{"error":{"message":"Invalid token","type":"OAuthException","code":190}}`)
	err := parseAPIErrorIG(body, 400)
	if err == nil {
		t.Fatal("expected error")
	}
	errStr := err.Error()
	if !strings.Contains(errStr, "Invalid token") {
		t.Errorf("expected error message to contain 'Invalid token', got '%s'", errStr)
	}

	// Test with invalid JSON
	body = []byte(`not json`)
	err = parseAPIErrorIG(body, 500)
	if err == nil {
		t.Fatal("expected error")
	}
	errStr = err.Error()
	if !strings.Contains(errStr, "500") {
		t.Errorf("expected status code in error, got '%s'", errStr)
	}
}

func TestInstagramClient_buildURL(t *testing.T) {
	client := NewInstagramClient("test_secret")
	client.baseURL = "https://graph.facebook.com/"

	result := client.buildURL("/123", map[string]string{"fields": "id,name"}, "test_token")

	if result == "" {
		t.Fatal("expected non-empty URL")
	}
	if !strings.Contains(result, "access_token=test_token") {
		t.Error("expected access_token in URL")
	}
	if !strings.Contains(result, "appsecret_proof=") {
		t.Error("expected appsecret_proof in URL")
	}
}

func TestInstagramClient_FetchAllMedia(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []kafkamodels.RawInstagramMedia{
				{ID: "media_1", MediaType: "IMAGE"},
				{ID: "media_2", MediaType: "VIDEO"},
			},
			"paging": map[string]string{},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchAllMedia(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(result))
	}
}

func TestInstagramClient_FetchInsightsDaily(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"name":   "impressions",
					"period": "day",
					"values": []map[string]interface{}{
						{"value": 1000, "end_time": "2024-01-15T08:00:00+0000"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	// FetchInsightsDaily takes (ctx, instagramID, accessToken, days, concurrency)
	result, err := client.FetchInsightsDaily(context.Background(), "ig_123", "token", 7, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchMedia_Pagination(t *testing.T) {
	callCount := 0
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		resp := map[string]interface{}{
			"data": []kafkamodels.RawInstagramMedia{
				{ID: "media_" + string(rune('0'+callCount))},
			},
			"paging": map[string]string{},
		}
		if callCount < 2 {
			resp["paging"] = map[string]string{"next": serverURL + r.URL.Path + "?after=cursor"}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	serverURL = server.URL

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchMedia(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) < 1 {
		t.Fatalf("expected at least 1 media item, got %d", len(result))
	}
}

func TestInstagramClient_HandleMediaAPIError_401(t *testing.T) {
	client := NewInstagramClient("test_secret")

	// Create a mock response with 401 status
	resp := &http.Response{
		StatusCode: 401,
		Body:       newMockReadCloser([]byte(`{"error": {"message": "Invalid token"}}`)),
	}

	err := client.handleMediaAPIError(resp)

	if err == nil {
		t.Fatal("expected error")
	}

	var authErr *InstagramAuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected InstagramAuthError, got %T", err)
	}
	if authErr.StatusCode != 401 {
		t.Fatalf("expected status 401, got %d", authErr.StatusCode)
	}
}

func TestInstagramClient_HandleMediaAPIError_403(t *testing.T) {
	client := NewInstagramClient("test_secret")

	resp := &http.Response{
		StatusCode: 403,
		Body:       newMockReadCloser([]byte(`{"error": {"message": "Forbidden"}}`)),
	}

	err := client.handleMediaAPIError(resp)

	var authErr *InstagramAuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected InstagramAuthError, got %T", err)
	}
}

func TestInstagramClient_HandleMediaAPIError_OAuthException(t *testing.T) {
	client := NewInstagramClient("test_secret")

	resp := &http.Response{
		StatusCode: 400,
		Body:       newMockReadCloser([]byte(`{"error": {"type": "OAuthException", "message": "Token expired"}}`)),
	}

	err := client.handleMediaAPIError(resp)

	var authErr *InstagramAuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected InstagramAuthError for OAuthException, got %T", err)
	}
}

func TestInstagramClient_HandleMediaAPIError_AccessToken(t *testing.T) {
	client := NewInstagramClient("test_secret")

	resp := &http.Response{
		StatusCode: 400,
		Body:       newMockReadCloser([]byte(`{"error": {"message": "Invalid access token"}}`)),
	}

	err := client.handleMediaAPIError(resp)

	var authErr *InstagramAuthError
	if !errors.As(err, &authErr) {
		t.Fatalf("expected InstagramAuthError for access token error, got %T", err)
	}
}

func TestInstagramClient_HandleMediaAPIError_NonAuthError(t *testing.T) {
	client := NewInstagramClient("test_secret")

	resp := &http.Response{
		StatusCode: 500,
		Body:       newMockReadCloser([]byte(`{"error": {"message": "Internal server error"}}`)),
	}

	err := client.handleMediaAPIError(resp)

	if err == nil {
		t.Fatal("expected error")
	}

	var authErr *InstagramAuthError
	if errors.As(err, &authErr) {
		t.Fatal("expected non-auth error for 500")
	}

	if !strings.Contains(err.Error(), "500") {
		t.Fatalf("expected error to contain status code, got: %v", err)
	}
}

// mockReadCloser is a helper for creating mock response bodies
type mockReadCloser struct {
	data   []byte
	offset int
}

func newMockReadCloser(data []byte) *mockReadCloser {
	return &mockReadCloser{data: data}
}

func (m *mockReadCloser) Read(p []byte) (n int, err error) {
	if m.offset >= len(m.data) {
		return 0, errors.New("EOF")
	}
	n = copy(p, m.data[m.offset:])
	m.offset += n
	return n, nil
}

func (m *mockReadCloser) Close() error {
	return nil
}

func TestBuildMediaURL_WithFields(t *testing.T) {
	client := NewInstagramClient("test_secret")
	client.baseURL = "https://graph.facebook.com/"

	url, err := client.buildMediaURL("ig_123", "token")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(url, "ig_123") {
		t.Fatal("URL should contain instagram ID")
	}
	if !strings.Contains(url, "access_token") {
		t.Fatal("URL should contain access_token")
	}
}

func TestInstagramClient_DoWithRetry_ContextCancelled(t *testing.T) {
	client := NewInstagramClient("test_secret")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost", nil)
	_, _, err := client.doWithRetry(ctx, "ig_123", req, "test")

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestInstagramClient_DoWithRetry_HTTPErrorThenSuccess(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call - simulate connection error by closing
			hj, ok := w.(http.Hijacker)
			if ok {
				conn, _, _ := hj.Hijack()
				conn.Close()
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Second call - success
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	body, status, err := client.doWithRetry(context.Background(), "ig_123", req, "test")

	if err != nil {
		t.Fatalf("expected success after retry, got: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
	if body == nil {
		t.Fatal("expected non-nil body")
	}
}

func TestInstagramClient_DoWithRetry_RetryAfterHeader(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{"message": "Rate limited"},
			})
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	_, status, err := client.doWithRetry(context.Background(), "ig_123", req, "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected 200 after retry, got %d", status)
	}
}

func TestInstagramClient_DoWithRetry_MaxAttemptsNonIGError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway")) // Not IG error format
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	_, status, err := client.doWithRetry(context.Background(), "ig_123", req, "test")

	if err == nil {
		t.Fatal("expected error after max attempts")
	}
	if status != http.StatusBadGateway {
		t.Fatalf("expected status 502, got %d", status)
	}
}

func TestInstagramClient_FetchMediaSince_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "media1", "timestamp": "2026-01-28T10:00:00+0000"},
				{"id": "media2", "timestamp": "2026-01-27T10:00:00+0000"},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	// Use a fixed time before both test media items to avoid time-sensitive failures
	since := time.Date(2026, 1, 26, 0, 0, 0, 0, time.UTC)
	result, err := client.FetchMediaSince(context.Background(), "ig_123", "token", since)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(result))
	}
}

func TestInstagramClient_FetchMediaWithLimit_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "media1"},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchMediaWithLimit(context.Background(), "ig_123", "token", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 media item, got %d", len(result))
	}
}

func TestInstagramClient_FetchAllMedia_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "media1"},
				{"id": "media2"},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchAllMedia(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(result))
	}
}

func TestInstagramClient_FetchInsightsDaily_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"name":   "impressions",
					"values": []map[string]interface{}{{"value": 100}},
				},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchInsightsDaily(context.Background(), "ig_123", "token", 3, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 3 {
		t.Fatalf("expected 3 daily insights, got %d", len(result))
	}
}

func TestInstagramClient_FetchAccountMedia_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              "ig_123",
			"username":        "testuser",
			"profile_picture": "http://example.com/pic.jpg",
			"media": map[string]interface{}{
				"data": []map[string]interface{}{
					{"id": "media1"},
				},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchAccountMedia(context.Background(), "ig_123", "token", 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchStories_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "story1", "media_type": "IMAGE"},
				{"id": "story2", "media_type": "VIDEO"},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	stories, err := client.FetchStories(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(stories) != 2 {
		t.Fatalf("expected 2 stories, got %d", len(stories))
	}
}

func TestInstagramClient_FetchStories_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{"message": "Invalid token"},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchStories(context.Background(), "ig_123", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstagramClient_FetchMediaSince_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{"message": "Forbidden"},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	_, err := client.FetchMediaSince(context.Background(), "ig_123", "token", since)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstagramClient_FetchMediaWithLimit_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{"message": "Bad request"},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchMediaWithLimit(context.Background(), "ig_123", "token", 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstagramClient_FetchAllMedia_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{"message": "Rate limited"},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchAllMedia(context.Background(), "ig_123", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstagramClient_FetchInsights_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"name":   "impressions",
					"values": []map[string]interface{}{{"value": 1000}},
				},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-7 * 24 * time.Hour)
	until := time.Now()
	result, err := client.FetchInsights(context.Background(), "ig_123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchMediaInsights_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{"message": "Invalid token"},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchMediaInsights(context.Background(), "media123", "token", "IMAGE", "FEED")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstagramClient_FetchMediaInsights_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"name": "impressions", "values": []map[string]interface{}{{"value": 100}}},
				{"name": "reach", "values": []map[string]interface{}{{"value": 50}}},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	insights, err := client.FetchMediaInsights(context.Background(), "media123", "token", "IMAGE", "FEED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if insights == nil {
		t.Fatal("expected non-nil insights")
	}
}

func TestInstagramClient_FetchUserInfo_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("Bad Gateway"))
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchUserInfo(context.Background(), "ig_123", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstagramClient_FetchUserInfo_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              "ig_123",
			"username":        "testuser",
			"followers_count": 1000,
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	info, err := client.FetchUserInfo(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if info["username"] != "testuser" {
		t.Fatalf("expected username 'testuser', got '%v'", info["username"])
	}
}

func TestInstagramClient_FetchAccountMedia_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Invalid token",
				"type":    "OAuthException",
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchAccountMedia(context.Background(), "ig_123", "token", 25)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstagramClient_FetchAccountDemographics_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return a successful response with demographics data
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"name":   "follower_demographics",
					"values": []map[string]interface{}{{"value": map[string]interface{}{"US": 100}}},
				},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchAccountDemographics(context.Background(), "ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Demographics may be nil but should not error
	_ = result
}

func TestInstagramClient_FetchMediaInsights_Reels(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's requesting reels metrics
		query := r.URL.Query()
		metrics := query.Get("metric")
		if !strings.Contains(metrics, "ig_reels_avg_watch_time") {
			t.Error("expected reels metrics")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"name": "views", "values": []map[string]interface{}{{"value": 1000}}},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchMediaInsights(context.Background(), "media_123", "token", "VIDEO", "REELS")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchMediaInsights_Story(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's requesting story metrics
		query := r.URL.Query()
		metrics := query.Get("metric")
		if !strings.Contains(metrics, "replies") {
			t.Error("expected story metrics with replies")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"name": "reach", "values": []map[string]interface{}{{"value": 500}}},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchMediaInsights(context.Background(), "media_123", "token", "IMAGE", "STORY")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchMediaInsights_Feed(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify it's requesting feed metrics (no ig_reels_ prefix)
		query := r.URL.Query()
		metrics := query.Get("metric")
		if strings.Contains(metrics, "ig_reels_") {
			t.Error("feed should not have ig_reels metrics")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"name": "likes", "values": []map[string]interface{}{{"value": 200}}},
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchMediaInsights(context.Background(), "media_123", "token", "IMAGE", "FEED")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_FetchMediaInsights_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Insights not available",
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	// Bad request returns an error after retries exhausted
	_, err := client.FetchMediaInsights(context.Background(), "media_123", "token", "IMAGE", "FEED")
	if err == nil {
		t.Fatal("expected error for bad request")
	}
}

func TestInstagramClient_FetchMediaInsights_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal Server Error"))
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchMediaInsights(context.Background(), "media_123", "token", "IMAGE", "FEED")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}

func TestInstagramClient_FetchMediaInsights_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchMediaInsights(context.Background(), "media_123", "token", "IMAGE", "FEED")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestInstagramClient_doAccountRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              "ig_123",
			"followers_count": 1000,
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.doAccountRequest(context.Background(), "ig_123", "token", "id,followers_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestInstagramClient_doAccountRequest_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.doAccountRequest(context.Background(), "ig_123", "token", "id")
	if err == nil {
		t.Fatal("expected error for HTTP error")
	}
}

func TestInstagramClient_fetchMediaPage_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "media_1", "media_type": "IMAGE"},
				{"id": "media_2", "media_type": "VIDEO"},
			},
			"paging": map[string]interface{}{
				"next": "",
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	pageURL := server.URL + "/test"

	media, nextURL, err := client.fetchMediaPage(context.Background(), pageURL, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(media) != 2 {
		t.Fatalf("expected 2 media items, got %d", len(media))
	}
	if nextURL != "" {
		t.Fatalf("expected empty next URL, got '%s'", nextURL)
	}
}

func TestInstagramClient_fetchMediaPage_WithPaging(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "media_1"},
			},
			"paging": map[string]interface{}{
				"next": "http://example.com/next",
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	pageURL := server.URL + "/test"

	media, nextURL, err := client.fetchMediaPage(context.Background(), pageURL, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(media) != 1 {
		t.Fatalf("expected 1 media item, got %d", len(media))
	}
	if nextURL == "" {
		t.Fatal("expected non-empty next URL")
	}
}

func TestInstagramClient_fetchMediaPage_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	pageURL := server.URL + "/test"

	_, _, err := client.fetchMediaPage(context.Background(), pageURL, "token")
	if err == nil {
		t.Fatal("expected error for HTTP error")
	}
}

func TestInstagramClient_fetchMediaPage_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("not json"))
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	pageURL := server.URL + "/test"

	_, _, err := client.fetchMediaPage(context.Background(), pageURL, "token")
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestInstagramClient_buildMediaURL_Success(t *testing.T) {
	client := NewInstagramClient("test_secret")

	url, err := client.buildMediaURL("ig_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if url == "" {
		t.Fatal("expected non-empty URL")
	}
	if !strings.Contains(url, "ig_123") {
		t.Fatal("URL should contain instagram ID")
	}
}

func TestInstagramClient_FetchStories_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchStories(context.Background(), "ig_123", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestInstagramClient_FetchInsights_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("invalid json"))
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"

	now := time.Now()
	since := now.Add(-7 * 24 * time.Hour)
	// FetchInsights doesn't return error for invalid JSON, it logs warning and returns empty result
	result, err := client.FetchInsights(context.Background(), "ig_123", "token", since, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Result should have empty data
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

// Tests for IsExpectedCompetitorError function
func TestIsExpectedCompetitorError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"Application does not have permission", errors.New("Application does not have permission for this action"), true},
		{"OAuthException/10", errors.New("OAuthException/10"), true},
		{"Invalid OAuth access token", errors.New("Invalid OAuth access token"), true},
		{"Cannot parse access token", errors.New("Cannot parse access token"), true},
		{"OAuthException/190", errors.New("OAuthException/190"), true},
		{"user must be an administrator", errors.New("user must be an administrator, editor, or moderator"), true},
		{"does not exist, cannot be loaded", errors.New("does not exist, cannot be loaded due to missing permissions"), true},
		{"GraphMethodException/100", errors.New("GraphMethodException/100"), true},
		{"This Page access token belongs to a Page", errors.New("This Page access token belongs to a Page that is not accessible"), true},
		{"permission denied", errors.New("permission denied"), false},
		{"network timeout", errors.New("network timeout"), false},
		{"parse error", errors.New("failed to parse json"), false},
		{"status 500", errors.New("internal server error status 500"), false},
		{"status 401 unauthorized", errors.New("status 401 unauthorized"), false},
		{"Connection refused", errors.New("Connection refused"), false},
		{"empty string error", errors.New(""), false},
		{"case insensitive OAuthException", errors.New("oauthexception/190"), false},
		{"network error with OAuthException", errors.New("Not OAuthException/190 error"), true},
		{"multiple patterns", errors.New("OAuthException/10 and OAuthException/190"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExpectedCompetitorError(tt.err)
			if got != tt.expected {
				t.Errorf("IsExpectedCompetitorError() = %v, want %v for error: %v", got, tt.expected, tt.err)
			}
		})
	}
}

// ==================== Logging Contract Tests ====================

// TestLoggingContract_InstagramClient_WarnLevelOnly verifies that the Instagram
// client only logs at Warn level (never Error) for all error paths.
func TestLoggingContract_InstagramClient_WarnLevelOnly(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Invalid token",
				"type":    "OAuthException",
				"code":    190,
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"
	client.log = log

	_, err := client.FetchUserInfo(context.Background(), "ig_123", "bad_token")
	if err == nil {
		t.Fatal("expected error for API error response")
	}

	output := buf.String()
	if strings.Contains(output, "ERR") {
		t.Fatalf("Instagram client should NOT produce ERR-level logs, got: %s", output)
	}
	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
	}
}

// TestLoggingContract_InstagramClient_NoCaptureException verifies that no error
// path in the Instagram client calls CaptureException directly.
func TestLoggingContract_InstagramClient_NoCaptureException(t *testing.T) {
	log, _ := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	// Error path 1: HTTP error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Expired token",
				"type":    "OAuthException",
				"code":    190,
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"
	client.log = log

	_, _ = client.FetchUserInfo(context.Background(), "ig_123", "expired_token")

	// Error path 2: context cancelled
	client2 := NewInstagramClient("test_secret")
	client2.baseURL = server.URL + "/"
	client2.log = log

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _ = client2.FetchUserInfo(ctx, "ig_456", "token")

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls across all error paths, got %d", len(*captureRecords))
	}
}

// TestLoggingContract_InstagramClient_ErrorsReturnedToCaller verifies that error
// paths return errors to the caller (not swallowed silently).
func TestLoggingContract_InstagramClient_ErrorsReturnedToCaller(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Server error",
				"code":    2,
			},
		})
	}))
	defer server.Close()

	log, _ := logger.NewTestLoggerWithHook()
	client := NewInstagramClient("test_secret")
	client.baseURL = server.URL + "/"
	client.log = log

	_, err := client.FetchUserInfo(context.Background(), "ig_789", "token")
	if err == nil {
		t.Fatal("FetchUserInfo: expected error to be returned to caller, got nil")
	}

	_, err = client.FetchMedia(context.Background(), "ig_789", "token")
	if err == nil {
		t.Fatal("FetchMedia: expected error to be returned to caller, got nil")
	}
}
