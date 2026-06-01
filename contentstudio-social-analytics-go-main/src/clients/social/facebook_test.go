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
	fbmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestNewRateManager(t *testing.T) {
	// Test with zero values (should use defaults)
	rm := NewRateManager(RateLimits{})
	if rm == nil {
		t.Fatal("expected non-nil RateManager")
	}
	if rm.limits.PerTokenRPS != 4.0 {
		t.Fatalf("expected PerTokenRPS 4.0, got %f", rm.limits.PerTokenRPS)
	}
	if rm.limits.PerTokenBurst != 4 {
		t.Fatalf("expected PerTokenBurst 4, got %d", rm.limits.PerTokenBurst)
	}
	if rm.limits.GlobalRPS != 12.0 {
		t.Fatalf("expected GlobalRPS 12.0, got %f", rm.limits.GlobalRPS)
	}
	if rm.limits.GlobalBurst != 12 {
		t.Fatalf("expected GlobalBurst 12, got %d", rm.limits.GlobalBurst)
	}
}

func TestNewRateManager_CustomLimits(t *testing.T) {
	rm := NewRateManager(RateLimits{
		PerTokenRPS:   10.0,
		PerTokenBurst: 5,
		GlobalRPS:     20.0,
		GlobalBurst:   10,
	})
	if rm.limits.PerTokenRPS != 10.0 {
		t.Fatalf("expected PerTokenRPS 10.0, got %f", rm.limits.PerTokenRPS)
	}
	if rm.limits.GlobalBurst != 10 {
		t.Fatalf("expected GlobalBurst 10, got %d", rm.limits.GlobalBurst)
	}
}

func TestRateManager_Wait(t *testing.T) {
	rm := NewRateManager(RateLimits{
		PerTokenRPS:   100.0,
		PerTokenBurst: 10,
		GlobalRPS:     100.0,
		GlobalBurst:   10,
	})

	ctx := context.Background()
	err := rm.Wait(ctx, "test_token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRateManager_Wait_ContextCancelled(t *testing.T) {
	rm := NewRateManager(RateLimits{
		PerTokenRPS:   0.001, // Very slow
		PerTokenBurst: 1,
		GlobalRPS:     0.001,
		GlobalBurst:   1,
	})

	// Use up the burst
	ctx := context.Background()
	_ = rm.Wait(ctx, "test_token")

	// Now cancel context
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	err := rm.Wait(ctx, "test_token")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestRateManager_tokenLimiter(t *testing.T) {
	rm := NewRateManager(RateLimits{})

	// Get limiter for token
	lim1 := rm.tokenLimiter("token1")
	if lim1 == nil {
		t.Fatal("expected non-nil limiter")
	}

	// Get same limiter again
	lim2 := rm.tokenLimiter("token1")
	if lim1 != lim2 {
		t.Fatal("expected same limiter for same token")
	}

	// Get different limiter for different token
	lim3 := rm.tokenLimiter("token2")
	if lim1 == lim3 {
		t.Fatal("expected different limiter for different token")
	}
}

func TestNewFacebookClient(t *testing.T) {
	client := NewFacebookClient("test_secret")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.appSecret != "test_secret" {
		t.Fatalf("expected appSecret 'test_secret', got '%s'", client.appSecret)
	}
	if client.baseURL != fbBaseURL {
		t.Fatalf("expected baseURL '%s', got '%s'", fbBaseURL, client.baseURL)
	}
	if client.rate == nil {
		t.Fatal("expected non-nil rate manager")
	}
}

func TestNewFacebookClientWithRates(t *testing.T) {
	rm := NewRateManager(RateLimits{GlobalRPS: 50})
	client := NewFacebookClientWithRates("secret", rm)

	if client.rate != rm {
		t.Fatal("expected same rate manager")
	}
}

func TestNewFacebookClientWithRates_NilRateManager(t *testing.T) {
	client := NewFacebookClientWithRates("secret", nil)
	if client.rate == nil {
		t.Fatal("expected non-nil rate manager when nil is passed")
	}
}

func TestFacebookClient_generateAppSecretProof(t *testing.T) {
	client := NewFacebookClient("test_secret")
	proof := client.generateAppSecretProof("test_token")

	if len(proof) != 64 {
		t.Fatalf("expected 64 character hex string, got %d", len(proof))
	}

	// Same input should produce same output
	proof2 := client.generateAppSecretProof("test_token")
	if proof != proof2 {
		t.Fatal("expected same proof for same input")
	}

	// Different input should produce different output
	proof3 := client.generateAppSecretProof("different_token")
	if proof == proof3 {
		t.Fatal("expected different proof for different input")
	}
}

func TestFacebookClient_FetchPostsWithLimit(t *testing.T) {
	posts := []kafkamodels.RawFacebookPost{
		{ID: "123_456", Message: "Test post 1"},
		{ID: "123_457", Message: "Test post 2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Data   []kafkamodels.RawFacebookPost `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}{
			Data: posts,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchPostsWithLimit(context.Background(), "123", "token", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(result))
	}
}

func TestFacebookClient_FetchPostsWithLimit_Pagination(t *testing.T) {
	callCount := 0
	var serverURL string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		posts := []kafkamodels.RawFacebookPost{
			{ID: "123_" + string(rune('0'+callCount))},
		}
		resp := struct {
			Data   []kafkamodels.RawFacebookPost `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}{
			Data: posts,
		}
		if callCount < 3 {
			// Return absolute URL with server host for pagination
			resp.Paging.Next = serverURL + r.URL.String()
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()
	serverURL = server.URL

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchPostsWithLimit(context.Background(), "123", "token", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 posts (3 pages), got %d", len(result))
	}
}

func TestFacebookClient_FetchPostsWithLimit_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := apiError{}
		resp.Error.Message = "Invalid token"
		resp.Error.Type = "OAuthException"
		resp.Error.Code = 190
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchPostsWithLimit(context.Background(), "123", "token", 1)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestFacebookClient_FetchPostsWithLimit_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.FetchPostsWithLimit(ctx, "123", "token", 1)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestFacebookClient_FetchPosts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []kafkamodels.RawFacebookPost{{ID: "123_456"}},
			"paging": map[string]string{},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchPosts(context.Background(), "123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
}

func TestFacebookClient_FetchVideosWithLimit(t *testing.T) {
	videos := []kafkamodels.RawFacebookVideo{
		{ID: "video_1"},
		{ID: "video_2"},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		resp := struct {
			Data   []kafkamodels.RawFacebookVideo `json:"data"`
			Paging struct {
				Next string `json:"next"`
			} `json:"paging"`
		}{
			Data: videos,
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchVideosWithLimit(context.Background(), "123", "token", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 videos, got %d", len(result))
	}
}

func TestFacebookClient_FetchVideos(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []kafkamodels.RawFacebookVideo{{ID: "video_1"}},
			"paging": map[string]string{},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchVideos(context.Background(), "123", "token", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}
}

func TestFacebookClient_FetchVideosSince(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []kafkamodels.RawFacebookVideo{{ID: "video_1"}},
			"paging": map[string]string{},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	result, err := client.FetchVideosSince(context.Background(), "123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}
}

func TestFacebookClient_FetchPostsSince(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data":   []kafkamodels.RawFacebookPost{{ID: "123_456"}},
			"paging": map[string]string{},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	result, err := client.FetchPostsSince(context.Background(), "123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
}

func TestFacebookClient_FetchInsights(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is insights, demographic, or page info request
		if r.URL.Path == "/v20.0/123/insights" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []kafkamodels.FacebookInsightData{
					{
						Name:   "page_follows",
						Period: "day",
						Values: []kafkamodels.FacebookInsightValue{
							{Value: 100, EndTime: "2024-01-15T08:00:00+0000"},
						},
					},
				},
			})
		} else if r.URL.Path == "/v20.0/123" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"talking_about_count": 50,
				"category":            "Test",
				"fan_count":           1000,
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{"data": []interface{}{}})
		}
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	result, err := client.FetchInsights(context.Background(), "123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result.PageID != "123" {
		t.Fatalf("expected PageID '123', got '%s'", result.PageID)
	}
}

func TestFacebookClient_FetchInsights_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiError{})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	_, err := client.FetchInsights(context.Background(), "123", "token", since, until)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestFacebookClient_waitRate_NilRateManager(t *testing.T) {
	client := &FacebookClient{
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

func TestNormalizePostID(t *testing.T) {
	cases := []struct {
		name     string
		postID   string
		pageID   string
		expected string
	}{
		{"empty postID", "", "123", ""},
		{"with underscore", "123_456", "123", "123_456"},
		{"without underscore", "456", "123", "123_456"},
		{"whitespace", "  456  ", "123", "123_456"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizePostID(tc.postID, tc.pageID)
			if result != tc.expected {
				t.Fatalf("expected '%s', got '%s'", tc.expected, result)
			}
		})
	}
}

func TestBestImage(t *testing.T) {
	cases := []struct {
		name     string
		a        fbImage
		b        fbImage
		expected fbImage
	}{
		{
			name:     "a larger",
			a:        fbImage{Src: "a.jpg", W: 100, H: 100},
			b:        fbImage{Src: "b.jpg", W: 50, H: 50},
			expected: fbImage{Src: "a.jpg", W: 100, H: 100},
		},
		{
			name:     "b larger",
			a:        fbImage{Src: "a.jpg", W: 50, H: 50},
			b:        fbImage{Src: "b.jpg", W: 100, H: 100},
			expected: fbImage{Src: "b.jpg", W: 100, H: 100},
		},
		{
			name:     "equal size returns a",
			a:        fbImage{Src: "a.jpg", W: 100, H: 100},
			b:        fbImage{Src: "b.jpg", W: 100, H: 100},
			expected: fbImage{Src: "a.jpg", W: 100, H: 100},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			result := bestImage(tc.a, tc.b)
			if result.Src != tc.expected.Src {
				t.Fatalf("expected Src '%s', got '%s'", tc.expected.Src, result.Src)
			}
		})
	}
}

func TestComputeBackoff(t *testing.T) {
	for i := 1; i <= 5; i++ {
		backoff := computeBackoff(i)
		if backoff < 0 {
			t.Fatalf("backoff should be positive, got %v", backoff)
		}
		if i > 1 && backoff > maxBackoff+time.Second {
			t.Fatalf("backoff exceeded max: %v", backoff)
		}
	}
}

func TestMin(t *testing.T) {
	if min(5, 10) != 5 {
		t.Fatal("expected 5")
	}
	if min(10, 5) != 5 {
		t.Fatal("expected 5")
	}
	if min(5, 5) != 5 {
		t.Fatal("expected 5")
	}
}

func TestWithMinBudget(t *testing.T) {
	// Test with parent that has sufficient deadline
	parent, parentCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer parentCancel()

	ctx, cancel := withMinBudget(parent, 1*time.Second)
	defer cancel()

	if ctx == nil {
		t.Fatal("expected non-nil context")
	}

	// Test with parent that has no deadline
	ctx2, cancel2 := withMinBudget(context.Background(), 1*time.Second)
	defer cancel2()

	if ctx2 == nil {
		t.Fatal("expected non-nil context")
	}
}

func TestTernaryID(t *testing.T) {
	// Test nil input
	result := ternaryID(nil)
	if result != "" {
		t.Fatalf("expected empty string for nil input, got '%s'", result)
	}

	// Test non-nil input
	input := &struct {
		ID string `json:"id"`
	}{ID: "test_id"}
	result = ternaryID(input)
	if result != "test_id" {
		t.Fatalf("expected 'test_id', got '%s'", result)
	}
}

func TestApiError_Struct(t *testing.T) {
	err := apiError{}
	err.Error.Message = "Test error"
	err.Error.Type = "OAuthException"
	err.Error.Code = 190
	err.Error.FBTraceID = "trace123"

	if err.Error.Message != "Test error" {
		t.Fatalf("expected message 'Test error', got '%s'", err.Error.Message)
	}
}

func TestRateLimits_Struct(t *testing.T) {
	limits := RateLimits{
		PerTokenRPS:   5.0,
		PerTokenBurst: 10,
		GlobalRPS:     15.0,
		GlobalBurst:   20,
	}

	if limits.PerTokenRPS != 5.0 {
		t.Fatalf("expected PerTokenRPS 5.0, got %f", limits.PerTokenRPS)
	}
}

func TestBuildUniqueNormalizedIDs(t *testing.T) {
	posts := []struct {
		PostID string
		PageID string
	}{
		{PostID: "456", PageID: "123"},
		{PostID: "456", PageID: "123"}, // Duplicate
		{PostID: "789", PageID: "123"},
		{PostID: "", PageID: "123"}, // Empty
	}

	// Convert to MinimalPost format for testing
	type MinimalPost struct {
		PostID string
	}
	var minPosts []MinimalPost
	for _, p := range posts {
		minPosts = append(minPosts, MinimalPost{PostID: p.PostID})
	}

	// Manual test of deduplication logic
	seen := make(map[string]struct{})
	var result []string
	for _, p := range minPosts {
		np := normalizePostID(p.PostID, "123")
		if np == "" {
			continue
		}
		if _, ok := seen[np]; ok {
			continue
		}
		seen[np] = struct{}{}
		result = append(result, np)
	}

	if len(result) != 2 {
		t.Fatalf("expected 2 unique IDs, got %d", len(result))
	}
}

// Competitor API tests
func TestFacebookClient_GetCompetitorPageDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/v20.0/page123" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"id":              "page123",
				"name":            "Test Page",
				"fan_count":       1000,
				"about":           "Test about",
				"followers_count": 500,
			})
		} else if r.URL.Path == "/v20.0/page123/picture" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": map[string]interface{}{
					"url": "https://example.com/pic.jpg",
				},
			})
		}
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	page, pic, err := client.GetCompetitorPageDetails(context.Background(), "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if page == nil {
		t.Fatal("expected non-nil page")
	}
	if page.Name != "Test Page" {
		t.Fatalf("expected name 'Test Page', got '%s'", page.Name)
	}
	if pic != nil && pic.Data != nil && pic.Data.URL != "https://example.com/pic.jpg" {
		t.Fatalf("expected picture URL, got '%s'", pic.Data.URL)
	}
}

func TestFacebookClient_GetCompetitorPageDetails_APIError(t *testing.T) {
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

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	_, _, err := client.GetCompetitorPageDetails(context.Background(), "page123", "token")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestFacebookClient_GetCompetitorPosts(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "post1", "message": "Test post 1"},
				{"id": "post2", "message": "Test post 2"},
			},
			"paging": map[string]interface{}{
				"next": "",
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	posts, nextURL, err := client.GetCompetitorPosts(context.Background(), "page123", "token", since, until, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("expected 2 posts, got %d", len(posts))
	}
	if nextURL != "" {
		t.Fatalf("expected empty next URL, got '%s'", nextURL)
	}
}

func TestFacebookClient_GetCompetitorPosts_WithPagination(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "post1", "message": "Test post 1"},
			},
			"paging": map[string]interface{}{
				"next": "https://graph.facebook.com/next_page",
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()

	posts, nextURL, err := client.GetCompetitorPosts(context.Background(), "page123", "token", since, until, 25)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if nextURL != "https://graph.facebook.com/next_page" {
		t.Fatalf("expected next URL, got '%s'", nextURL)
	}
}

func TestFacebookClient_GetCompetitorPostsFromURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "post2", "message": "Test post 2"},
			},
			"paging": nil,
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	posts, nextURL, err := client.GetCompetitorPostsFromURL(context.Background(), server.URL+"/next_page", "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
	if nextURL != "" {
		t.Fatalf("expected empty next URL, got '%s'", nextURL)
	}
}

func TestFacebookClient_GetCompetitorPostsFromURL_AddSecretProof(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify that appsecret_proof is added
		if r.URL.Query().Get("appsecret_proof") == "" {
			t.Error("expected appsecret_proof in request")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	// URL without appsecret_proof
	_, _, err := client.GetCompetitorPostsFromURL(context.Background(), server.URL+"/next_page?param=value", "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFacebookClient_GetCompetitorSharedPostDetails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "parent_123",
			"message": "Original shared post",
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	post, err := client.GetCompetitorSharedPostDetails(context.Background(), "parent_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if post == nil {
		t.Fatal("expected non-nil post")
	}
	if post.ID != "parent_123" {
		t.Fatalf("expected post ID 'parent_123', got '%s'", post.ID)
	}
}

func TestFacebookClient_GetCompetitorPagePicture(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": map[string]interface{}{
				"url": "https://example.com/profile.jpg",
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	pic, err := client.GetCompetitorPagePicture(context.Background(), "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pic == nil {
		t.Fatal("expected non-nil picture")
	}
	if pic.Data == nil || pic.Data.URL != "https://example.com/profile.jpg" {
		t.Fatal("expected picture URL")
	}
}

func TestFacebookClient_GetCompetitorPagePicture_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Access denied",
				"type":    "OAuthException",
				"code":    200,
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.GetCompetitorPagePicture(context.Background(), "page123", "token")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestBuildCompURL(t *testing.T) {
	result := buildCompURL(
		"https://graph.facebook.com/v20.0/123",
		map[string]string{"fields": "id,name"},
		"test_token",
		"proof123",
	)

	if result == "" {
		t.Fatal("expected non-empty URL")
	}
	// Verify key components are present
	if !contains(result, "access_token=test_token") {
		t.Error("expected access_token in URL")
	}
	if !contains(result, "appsecret_proof=proof123") {
		t.Error("expected appsecret_proof in URL")
	}
	if !contains(result, "fields=") {
		t.Error("expected fields in URL")
	}
}

func TestParseAPIErrorFB(t *testing.T) {
	// Test with valid error response
	body := []byte(`{"error":{"message":"Invalid token","type":"OAuthException","code":190}}`)
	err := parseAPIErrorFB(body, 400)
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "Invalid token") {
		t.Errorf("expected error message to contain 'Invalid token', got '%s'", err.Error())
	}

	// Test with invalid JSON
	body = []byte(`not json`)
	err = parseAPIErrorFB(body, 500)
	if err == nil {
		t.Fatal("expected error")
	}
	if !contains(err.Error(), "500") {
		t.Errorf("expected status code in error, got '%s'", err.Error())
	}
}

func TestFacebookClient_doWithRetry(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	body, status, err := client.doWithRetry(context.Background(), "123", req, "test")

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

func TestFacebookClient_doWithRetry_RetryOnServerError(t *testing.T) {
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

	client := NewFacebookClient("test_secret")

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	_, status, err := client.doWithRetry(context.Background(), "123", req, "test")

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

func TestDedupeChildren(t *testing.T) {
	// Test the deduplication logic manually
	seen := make(map[string]bool)
	var result []string

	children := []string{"a", "b", "a", "c", "b"}
	for _, child := range children {
		if !seen[child] {
			seen[child] = true
			result = append(result, child)
		}
	}

	if len(result) != 3 {
		t.Fatalf("expected 3 unique children, got %d", len(result))
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Tests for uncovered functions

func TestBuildUniqueNormalizedIDs_WithClickhouseType(t *testing.T) {
	posts := []clickhouse.MinimalPost{
		{PostID: "456", PageID: "123"},
		{PostID: "456", PageID: "123"}, // Duplicate
		{PostID: "789", PageID: "123"},
		{PostID: "", PageID: "123"},        // Empty - should be skipped
		{PostID: "123_999", PageID: "123"}, // Already has page ID prefix
	}

	result := buildUniqueNormalizedIDs(posts, "123")

	if len(result) != 3 {
		t.Fatalf("expected 3 unique IDs, got %d: %v", len(result), result)
	}
}

func TestBuildUniqueNormalizedIDs_Empty(t *testing.T) {
	result := buildUniqueNormalizedIDs([]clickhouse.MinimalPost{}, "123")
	if len(result) != 0 {
		t.Fatalf("expected 0 IDs for empty input, got %d", len(result))
	}
}

func TestDecodeByID_Valid(t *testing.T) {
	input := []byte(`{
		"123_456": {"id": "123_456", "full_picture": "http://example.com/pic.jpg"},
		"123_789": {"id": "123_789", "full_picture": "http://example.com/pic2.jpg"}
	}`)

	result, err := decodeByID(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 items, got %d", len(result))
	}
	if result["123_456"].ID != "123_456" {
		t.Fatal("expected ID 123_456")
	}
}

func TestDecodeByID_InvalidJSON(t *testing.T) {
	input := []byte(`{invalid json}`)
	_, err := decodeByID(input)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestDecodeByID_Empty(t *testing.T) {
	input := []byte(`{}`)
	result, err := decodeByID(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 items, got %d", len(result))
	}
}

func TestExtractBestPerPost(t *testing.T) {
	batch := []string{"123_456", "123_789", "123_999"}
	byID := map[string]clickhouse.FbItem{
		"123_456": {
			ID:          "123_456",
			FullPicture: "http://example.com/fallback.jpg",
			Attachments: &struct {
				Data []struct {
					MediaType string `json:"media_type"`
					Type      string `json:"type"`
					Target    *struct {
						ID string `json:"id"`
					} `json:"target"`
					Media *struct {
						Image *struct {
							Src           string `json:"src"`
							Width, Height int
						} `json:"image"`
					} `json:"media"`
					Subattachments *struct {
						Data []struct {
							MediaType string `json:"media_type"`
							Type      string `json:"type"`
							Target    *struct {
								ID string `json:"id"`
							} `json:"target"`
							Media *struct {
								Image *struct {
									Src           string `json:"src"`
									Width, Height int
								} `json:"image"`
							} `json:"media"`
						} `json:"data"`
					} `json:"subattachments"`
				} `json:"data"`
			}{
				Data: []struct {
					MediaType string `json:"media_type"`
					Type      string `json:"type"`
					Target    *struct {
						ID string `json:"id"`
					} `json:"target"`
					Media *struct {
						Image *struct {
							Src           string `json:"src"`
							Width, Height int
						} `json:"image"`
					} `json:"media"`
					Subattachments *struct {
						Data []struct {
							MediaType string `json:"media_type"`
							Type      string `json:"type"`
							Target    *struct {
								ID string `json:"id"`
							} `json:"target"`
							Media *struct {
								Image *struct {
									Src           string `json:"src"`
									Width, Height int
								} `json:"image"`
							} `json:"media"`
						} `json:"data"`
					} `json:"subattachments"`
				}{
					{
						Media: &struct {
							Image *struct {
								Src           string `json:"src"`
								Width, Height int
							} `json:"image"`
						}{
							Image: &struct {
								Src           string `json:"src"`
								Width, Height int
							}{
								Src:    "http://example.com/best.jpg",
								Width:  800,
								Height: 600,
							},
						},
					},
				},
			},
		},
		"123_789": {
			ID:          "123_789",
			FullPicture: "http://example.com/only_full_picture.jpg",
		},
		// 123_999 not in byID - should be skipped
	}

	outThumbs := make(map[string]string)
	extracted := extractBestPerPost(batch, byID, outThumbs, make(map[string]struct{}))

	if extracted != 2 {
		t.Fatalf("expected 2 extracted, got %d", extracted)
	}
	if outThumbs["123_456"] != "http://example.com/best.jpg" {
		t.Fatalf("expected best.jpg for 123_456, got %s", outThumbs["123_456"])
	}
	if outThumbs["123_789"] != "http://example.com/only_full_picture.jpg" {
		t.Fatalf("expected only_full_picture.jpg for 123_789, got %s", outThumbs["123_789"])
	}
}

func TestExtractBestPerPost_EmptyID(t *testing.T) {
	batch := []string{"123_456"}
	byID := map[string]clickhouse.FbItem{
		"123_456": {ID: ""}, // Empty ID - should be skipped
	}

	outThumbs := make(map[string]string)
	extracted := extractBestPerPost(batch, byID, outThumbs, make(map[string]struct{}))

	if extracted != 0 {
		t.Fatalf("expected 0 extracted for empty ID, got %d", extracted)
	}
}

func TestExtractBestPerPost_NoAttachmentsNoFullPicture(t *testing.T) {
	batch := []string{"123_456"}
	byID := map[string]clickhouse.FbItem{
		"123_456": {ID: "123_456"}, // No attachments, no full_picture
	}

	outThumbs := make(map[string]string)
	extracted := extractBestPerPost(batch, byID, outThumbs, make(map[string]struct{}))

	if extracted != 0 {
		t.Fatalf("expected 0 extracted for no image, got %d", extracted)
	}
}

func TestAssembleResults(t *testing.T) {
	client := NewFacebookClient("secret")
	posts := []clickhouse.MinimalPost{
		{PostID: "456", PageID: "123"},
		{PostID: "789", PageID: "123"},
		{PostID: "999", PageID: "123"}, // This one has no thumbnail
	}
	outThumbs := map[string]string{
		"123_456": "http://example.com/pic1.jpg",
		"123_789": "http://example.com/pic2.jpg",
		// 123_999 not in outThumbs
	}

	result := client.assembleResults("123", posts, outThumbs)

	if len(result) != 2 {
		t.Fatalf("expected 2 results, got %d", len(result))
	}
	if result[0].FullPicture != "http://example.com/pic1.jpg" {
		t.Fatalf("expected pic1.jpg, got %s", result[0].FullPicture)
	}
}

func TestAssembleResults_Empty(t *testing.T) {
	client := NewFacebookClient("secret")
	result := client.assembleResults("123", []clickhouse.MinimalPost{}, map[string]string{})

	if len(result) != 0 {
		t.Fatalf("expected 0 results for empty input, got %d", len(result))
	}
}

func TestResolveToken_WithLongToken(t *testing.T) {
	// This test requires the crypto package - we'll test the fallback behavior
	client := NewFacebookClient("secret")

	// Test with empty tokens - should return error
	_, err := client.resolveToken("123", "", "", "key")
	if err == nil {
		t.Fatal("expected error for empty tokens")
	}

	// Test with valid access token
	token, err := client.resolveToken("123", "valid_token", "", "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "valid_token" {
		t.Fatalf("expected 'valid_token', got '%s'", token)
	}
}

func TestResolveToken_EmptyAfterTrim(t *testing.T) {
	client := NewFacebookClient("secret")

	// Test with whitespace-only token
	_, err := client.resolveToken("123", "   ", "", "key")
	if err == nil {
		t.Fatal("expected error for whitespace-only token")
	}
}

func TestDedupeChildren_WithFbModels(t *testing.T) {
	children := []fbmodels.ChildThumb{
		{MediaID: "1", ThumbURL: "http://a.com"},
		{MediaID: "2", ThumbURL: "http://b.com"},
		{MediaID: "1", ThumbURL: "http://a.com"}, // Duplicate
		{MediaID: "3", ThumbURL: "http://c.com"},
		{MediaID: "2", ThumbURL: "http://b.com"}, // Duplicate
	}

	result := dedupeChildren(children)

	if len(result) != 3 {
		t.Fatalf("expected 3 unique children, got %d", len(result))
	}
}

func TestDedupeChildren_Empty(t *testing.T) {
	result := dedupeChildren([]fbmodels.ChildThumb{})
	if len(result) != 0 {
		t.Fatalf("expected 0 for empty input, got %d", len(result))
	}
}

func TestDedupeChildren_SameMediaDifferentURL(t *testing.T) {
	children := []fbmodels.ChildThumb{
		{MediaID: "1", ThumbURL: "http://a.com"},
		{MediaID: "1", ThumbURL: "http://b.com"}, // Same media, different URL - should NOT be deduped
	}

	result := dedupeChildren(children)

	if len(result) != 2 {
		t.Fatalf("expected 2 (different URLs), got %d", len(result))
	}
}

func TestComputeBackoff_LargeTry(t *testing.T) {
	// Test that large try values don't cause overflow and are capped
	backoff := computeBackoff(10)

	// Should be capped at maxBackoff
	if backoff > 30*time.Second {
		t.Fatalf("backoff should be capped, got %v", backoff)
	}
	if backoff < 0 {
		t.Fatal("backoff should not be negative")
	}
}

func TestWithMinBudget_ParentHasEnoughBudget(t *testing.T) {
	parent, parentCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer parentCancel()

	ctx, cancel := withMinBudget(parent, 1*time.Second)
	defer cancel()

	// Context should work
	select {
	case <-ctx.Done():
		t.Fatal("context should not be done immediately")
	default:
		// OK
	}
}

func TestWithMinBudget_ParentHasNoBudget(t *testing.T) {
	parent, parentCancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer parentCancel()

	ctx, cancel := withMinBudget(parent, 10*time.Second)
	defer cancel()

	// Context should have its own timeout
	select {
	case <-ctx.Done():
		t.Fatal("context should not be done immediately")
	default:
		// OK
	}
}

func TestWithMinBudget_ParentCancelled(t *testing.T) {
	parent, parentCancel := context.WithCancel(context.Background())

	ctx, cancel := withMinBudget(parent, 1*time.Second)
	defer cancel()

	// Cancel parent
	parentCancel()

	// Wait a bit for the goroutine to propagate cancellation
	time.Sleep(50 * time.Millisecond)

	select {
	case <-ctx.Done():
		// OK - parent cancellation propagated
	case <-time.After(200 * time.Millisecond):
		t.Fatal("expected context to be cancelled after parent cancellation")
	}
}

func TestTernaryID_Nil(t *testing.T) {
	result := ternaryID(nil)
	if result != "" {
		t.Fatalf("expected empty string for nil, got '%s'", result)
	}
}

func TestTernaryID_WithValue(t *testing.T) {
	input := &struct {
		ID string `json:"id"`
	}{ID: "test123"}

	result := ternaryID(input)
	if result != "test123" {
		t.Fatalf("expected 'test123', got '%s'", result)
	}
}

func TestFacebookClient_GetPostThumbnails_EmptyPosts(t *testing.T) {
	client := NewFacebookClient("secret")

	result, err := client.GetPostThumbnails(
		context.Background(),
		"123",
		"token",
		"",
		"",
		[]clickhouse.MinimalPost{},
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatal("expected nil result for empty posts")
	}
}

func TestFacebookClient_GetPostThumbnails_NoToken(t *testing.T) {
	client := NewFacebookClient("secret")

	_, err := client.GetPostThumbnails(
		context.Background(),
		"123",
		"", // Empty token
		"",
		"",
		[]clickhouse.MinimalPost{{PostID: "456"}},
	)

	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestFacebookClient_GetPostThumbnails_ContextCancelled(t *testing.T) {
	client := NewFacebookClient("secret")

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := client.GetPostThumbnails(
		ctx,
		"123",
		"token",
		"",
		"",
		[]clickhouse.MinimalPost{{PostID: "456"}},
	)

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestFacebookClient_ProcessWindow_EmptyIDs(t *testing.T) {
	client := NewFacebookClient("secret")

	// Window with only empty/whitespace post IDs
	window := []clickhouse.MinimalPost{
		{PostID: "", PageID: "123"},
		{PostID: "   ", PageID: "123"},
	}

	outThumbs := make(map[string]string)
	err := client.processWindow(context.Background(), "123", "token", window, outThumbs, make(map[string]struct{}), 0, 2)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should have no results since all IDs are invalid
	if len(outThumbs) != 0 {
		t.Fatalf("expected 0 results for invalid IDs, got %d", len(outThumbs))
	}
}

func TestFacebookClient_ProcessWindow_ContextCancelled(t *testing.T) {
	client := NewFacebookClient("secret")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	window := []clickhouse.MinimalPost{{PostID: "456", PageID: "123"}}
	outThumbs := make(map[string]string)

	err := client.processWindow(ctx, "123", "token", window, outThumbs, make(map[string]struct{}), 0, 1)

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestFacebookClient_ProcessBatch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"123_456": map[string]interface{}{
				"id":           "123_456",
				"full_picture": "http://example.com/pic.jpg",
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("secret")
	client.baseURL = server.URL + "/"

	batch := []string{"123_456"}
	outThumbs := make(map[string]string)

	err := client.processBatch(context.Background(), "123", "token", batch, outThumbs, make(map[string]struct{}))

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outThumbs["123_456"] != "http://example.com/pic.jpg" {
		t.Fatalf("expected pic.jpg, got %s", outThumbs["123_456"])
	}
}

func TestFacebookClient_ProcessBatch_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := NewFacebookClient("secret")
	client.baseURL = server.URL + "/"

	batch := []string{"123_456"}
	outThumbs := make(map[string]string)

	err := client.processBatch(context.Background(), "123", "token", batch, outThumbs, make(map[string]struct{}))

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFacebookClient_DoWithRetry_RateLimitError(t *testing.T) {
	client := NewFacebookClient("secret")

	// Cancel context to trigger rate limit wait failure
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	req, _ := http.NewRequestWithContext(ctx, "GET", "http://localhost", nil)
	req.URL.RawQuery = "access_token=test"

	_, _, err := client.doWithRetry(ctx, "123", req, "test")

	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestFacebookClient_DoWithRetry_RetryAfterHeader(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			w.Write([]byte(`{"error": {"message": "Rate limited"}}`))
			return
		}
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewFacebookClient("secret")

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	_, status, err := client.doWithRetry(context.Background(), "123", req, "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if status != http.StatusOK {
		t.Fatalf("expected 200, got %d", status)
	}
}

func TestFacebookClient_DoWithRetry_MaxAttemptsExhausted(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": {"message": "Server error", "type": "OAuthException", "code": 1}}`))
	}))
	defer server.Close()

	client := NewFacebookClient("secret")

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	_, _, err := client.doWithRetry(context.Background(), "123", req, "test")

	if err == nil {
		t.Fatal("expected error after max attempts")
	}
	// Should have tried maxAttempts times (5)
	if callCount != 5 {
		t.Fatalf("expected 5 attempts, got %d", callCount)
	}
}

func TestFacebookClient_DoWithRetry_NonFBError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`Bad Gateway`)) // Not FB error format
	}))
	defer server.Close()

	client := NewFacebookClient("secret")

	req, _ := http.NewRequestWithContext(context.Background(), "GET", server.URL, nil)
	_, _, err := client.doWithRetry(context.Background(), "123", req, "test")

	if err == nil {
		t.Fatal("expected error")
	}
}

// Tests for makeAPICall
func TestFacebookClient_MakeAPICall_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"id": "123", "name": "Test"})
	}))
	defer server.Close()

	client := NewFacebookClient("secret")
	client.baseURL = server.URL + "/"

	var result map[string]string
	err := client.makeAPICall(context.Background(), "", "/test", nil, &result, "token", "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["id"] != "123" {
		t.Fatalf("expected id '123', got '%s'", result["id"])
	}
}

func TestFacebookClient_MakeAPICall_WithFullURL(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check that appsecret_proof was added
		if !strings.Contains(r.URL.RawQuery, "appsecret_proof") {
			t.Fatal("expected appsecret_proof in URL")
		}
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewFacebookClient("secret")

	var result map[string]string
	err := client.makeAPICall(context.Background(), server.URL+"?param=value", "", nil, &result, "token", "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFacebookClient_MakeAPICall_WithFullURLNoQueryString(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
	}))
	defer server.Close()

	client := NewFacebookClient("secret")

	var result map[string]string
	err := client.makeAPICall(context.Background(), server.URL, "", nil, &result, "token", "test")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFacebookClient_MakeAPICall_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{invalid json}`))
	}))
	defer server.Close()

	client := NewFacebookClient("secret")
	client.baseURL = server.URL + "/"

	var result map[string]string
	err := client.makeAPICall(context.Background(), "", "/test", nil, &result, "token", "test")

	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFacebookClient_MakeAPICall_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Invalid parameter",
				"code":    100,
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("secret")
	client.baseURL = server.URL + "/"

	var result map[string]string
	err := client.makeAPICall(context.Background(), "", "/test", nil, &result, "token", "test")

	if err == nil {
		t.Fatal("expected error for API error")
	}
}

// Tests for FetchVideosWithLimit edge cases
func TestFacebookClient_FetchVideosWithLimit_Error(t *testing.T) {
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

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	_, err := client.FetchVideosWithLimit(context.Background(), "123", "token", 1)
	if err == nil {
		t.Fatal("expected error for invalid token")
	}
}

func TestFacebookClient_FetchVideosWithLimit_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]interface{}{"data": []map[string]string{}})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.FetchVideosWithLimit(ctx, "123", "token", 1)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestFacebookClient_FetchVideosSince_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{"message": "Bad request"},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()
	_, err := client.FetchVideosSince(context.Background(), "123", "token", since, until)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFacebookClient_FetchPostsSince_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{"message": "Forbidden"},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()
	_, err := client.FetchPostsSince(context.Background(), "123", "token", since, until)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestFacebookClient_FetchInsights_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{"message": "Rate limited"},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-24 * time.Hour)
	until := time.Now()
	_, err := client.FetchInsights(context.Background(), "123", "token", since, until)
	if err == nil {
		t.Fatal("expected error")
	}
}

// Test for GetPostThumbnails with successful HTTP call
func TestFacebookClient_GetPostThumbnails_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"123_456": map[string]interface{}{
				"id":           "123_456",
				"full_picture": "http://example.com/pic.jpg",
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("secret")
	client.baseURL = server.URL + "/"

	posts := []clickhouse.MinimalPost{
		{PostID: "456", PageID: "123"},
	}

	result, err := client.GetPostThumbnails(
		context.Background(),
		"123",
		"token",
		"",
		"",
		posts,
	)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 result, got %d", len(result))
	}
	if result[0].FullPicture != "http://example.com/pic.jpg" {
		t.Fatalf("expected pic.jpg, got %s", result[0].FullPicture)
	}
}

// Test ProcessWindow with successful batch
func TestFacebookClient_ProcessWindow_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"123_456": map[string]interface{}{
				"id":           "123_456",
				"full_picture": "http://example.com/pic.jpg",
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("secret")
	client.baseURL = server.URL + "/"

	window := []clickhouse.MinimalPost{
		{PostID: "456", PageID: "123"},
	}
	outThumbs := make(map[string]string)

	err := client.processWindow(context.Background(), "123", "token", window, outThumbs, make(map[string]struct{}), 0, 1)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if outThumbs["123_456"] != "http://example.com/pic.jpg" {
		t.Fatalf("expected pic.jpg, got %s", outThumbs["123_456"])
	}
}

func TestFacebookClient_FetchVideosWithLimit_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "video1", "title": "Test Video"},
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.FetchVideosWithLimit(context.Background(), "123", "token", 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}
}

func TestFacebookClient_FetchVideosSince_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "video1", "created_time": "2026-01-28T10:00:00+0000"},
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-48 * time.Hour)
	until := time.Now()
	result, err := client.FetchVideosSince(context.Background(), "123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}
}

func TestFacebookClient_FetchPostsSince_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "123_456", "message": "Test post"},
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-48 * time.Hour)
	until := time.Now()
	result, err := client.FetchPostsSince(context.Background(), "123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
}

func TestFacebookClient_FetchInsights_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{
					"name":   "page_impressions",
					"values": []map[string]interface{}{{"value": 1000}},
				},
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	since := time.Now().Add(-7 * 24 * time.Hour)
	until := time.Now()
	result, err := client.FetchInsights(context.Background(), "123", "token", since, until)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestFacebookClient_DoWithRetry_HTTPRequestError(t *testing.T) {
	client := NewFacebookClient("secret")
	client.httpClient = &http.Client{
		Transport: &errorTransportFB{},
	}

	req, _ := http.NewRequestWithContext(context.Background(), "GET", "http://localhost", nil)
	_, _, err := client.doWithRetry(context.Background(), "123", req, "test")

	if err == nil {
		t.Fatal("expected error for transport failure")
	}
}

type errorTransportFB struct{}

func (e *errorTransportFB) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, errors.New("simulated transport error")
}

func TestFacebookClient_GetCompetitorSharedPostDetails_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":      "parent123",
			"message": "Shared post content",
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	result, err := client.GetCompetitorSharedPostDetails(context.Background(), "parent123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestFacebookClient_GetCompetitorPostsFromURL_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": []map[string]interface{}{
				{"id": "post1", "message": "Test post"},
			},
			"paging": map[string]interface{}{
				"next": "",
			},
		})
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	result, next, err := client.GetCompetitorPostsFromURL(context.Background(), server.URL, "page123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
	if next != "" {
		t.Fatalf("expected empty next URL, got %s", next)
	}
}

func TestFacebookClient_resolveToken_AccessToken(t *testing.T) {
	client := NewFacebookClient("test_secret")

	// With access token but no long token
	token, err := client.resolveToken("page_123", "access_token", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "access_token" {
		t.Fatalf("expected 'access_token', got '%s'", token)
	}
}

func TestFacebookClient_resolveToken_EmptyToken(t *testing.T) {
	client := NewFacebookClient("test_secret")

	// With empty tokens
	_, err := client.resolveToken("page_123", "", "", "")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestFacebookClient_resolveToken_WhitespaceToken(t *testing.T) {
	client := NewFacebookClient("test_secret")

	// With whitespace-only token
	_, err := client.resolveToken("page_123", "   ", "", "")
	if err == nil {
		t.Fatal("expected error for whitespace token")
	}
}

func TestFacebookClient_FetchInsights_PartialSuccess(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{
					{"name": "page_impressions", "values": []map[string]interface{}{{"value": 1000}}},
				},
			})
		} else {
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"data": []map[string]interface{}{},
			})
		}
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"

	now := time.Now()
	since := now.Add(-7 * 24 * time.Hour)
	result, err := client.FetchInsights(context.Background(), "page_123", "token", since, now)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should return partial results
	_ = result
}

// Tests for IsExpectedCompetitorErrorFB function
func TestIsExpectedCompetitorErrorFB(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"Error validating access token", errors.New("Error validating access token"), true},
		{"The session has been invalidated", errors.New("The session has been invalidated"), true},
		{"does not exist permission", errors.New("does not exist, cannot be loaded due to missing permissions"), true},
		{"GraphMethodException/100", errors.New("GraphMethodException/100"), true},
		{"does not support this operation", errors.New("does not support this operation"), true},
		{"Tried accessing nonexisting", errors.New("Tried accessing nonexisting field"), true},
		{"OAuthException/100", errors.New("OAuthException/100"), true},
		{"OAuthException/2", errors.New("OAuthException/2"), true},
		{"Not enough viewers", errors.New("Not enough viewers for the media to show insights"), true},
		{"OAuthException/10", errors.New("OAuthException/10"), true},
		{"status 401", errors.New("status 401 unauthorized"), true},
		{"status 403", errors.New("status 403 forbidden"), true},
		{"status 404", errors.New("status 404 not found"), true},
		{"network error", errors.New("network timeout"), false},
		{"parse error", errors.New("failed to parse json"), false},
		{"status 500", errors.New("status 500 internal server error"), false},
		{"generic error", errors.New("unknown error"), false},
		{"status 400", errors.New("status 400 bad request"), false},
		{"permission denied no status", errors.New("permission denied"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExpectedCompetitorErrorFB(tt.err)
			if got != tt.expected {
				t.Errorf("IsExpectedCompetitorErrorFB() = %v, want %v for error: %v", got, tt.expected, tt.err)
			}
		})
	}
}

// ==================== Logging Contract Tests ====================

func TestLoggingContract_FacebookClient_WarnLevelOnly(t *testing.T) {
	log, buf := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	// Mock server that returns a non-200 status to trigger error logging
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := apiError{}
		resp.Error.Message = "Invalid token"
		resp.Error.Type = "OAuthException"
		resp.Error.Code = 190
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"
	client.log = log // inject test logger

	_, err := client.FetchPostsWithLimit(context.Background(), "123", "token", 1)
	if err == nil {
		t.Fatal("expected error for API error response")
	}

	output := buf.String()
	// Facebook client logs at Warn level for errors
	if strings.Contains(output, "ERR") {
		t.Fatalf("unexpected ERR-level log in output: %s", output)
	}

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls, got %d", len(*captureRecords))
	}
}

func TestLoggingContract_FacebookClient_NoCaptureException(t *testing.T) {
	log, _ := logger.NewTestLoggerWithHook()
	captureRecords, cleanup := logger.InstallCaptureSpy()
	defer cleanup()

	// Error path 1: API error response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		resp := apiError{}
		resp.Error.Message = "Bad request"
		resp.Error.Code = 100
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewFacebookClient("test_secret")
	client.baseURL = server.URL + "/"
	client.log = log

	_, _ = client.FetchPostsWithLimit(context.Background(), "123", "token", 1)

	// Error path 2: Context cancelled
	client2 := NewFacebookClient("test_secret")
	client2.baseURL = server.URL + "/"
	client2.log = log

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _ = client2.FetchPostsWithLimit(ctx, "123", "token", 1)

	if len(*captureRecords) != 0 {
		t.Fatalf("expected 0 CaptureException calls across error paths, got %d", len(*captureRecords))
	}
}
