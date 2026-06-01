package social

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

func newTestData365Client(serverURL string) *Data365Client {
	log, _ := logger.NewTestLogger()
	return NewData365Client(config.Data365Config{
		BaseURL:      serverURL,
		AccessToken:  "test-token",
		PollInterval: 0, // will default to 5s but tests use short-lived servers
		PollTimeout:  5,
	}, log)
}

func TestSearchUpdatePath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		platform string
		keyword  string
		want     string
	}{
		{"facebook", "golang", "/facebook/search/golang/posts/latest/update"},
		{"instagram", "travel", "/instagram/search/post/update"},
		{"tiktok", "dance", "/tiktok/search/post/update"},
		{"twitter", "news", "/twitter/search/post/update"},
		{"reddit", "tesla", "/reddit/search/post/update"},
		{"threads", "tech", "/threads/search/post/update"},
	}
	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			t.Parallel()
			got := searchUpdatePath(tt.platform, tt.keyword)
			if got != tt.want {
				t.Errorf("searchUpdatePath(%q, %q) = %q, want %q", tt.platform, tt.keyword, got, tt.want)
			}
		})
	}
}

func TestSearchResultsPath(t *testing.T) {
	t.Parallel()
	tests := []struct {
		platform string
		keyword  string
		want     string
	}{
		{"facebook", "golang", "/facebook/search/golang/posts/latest/posts"},
		{"instagram", "travel", "/instagram/search/post/items"},
		{"tiktok", "dance", "/tiktok/search/post/items"},
		{"twitter", "news", "/twitter/search/post/posts"},
		{"reddit", "tesla", "/reddit/search/post/items"},
		{"threads", "tech", "/threads/search/post/items"},
	}
	for _, tt := range tests {
		t.Run(tt.platform, func(t *testing.T) {
			t.Parallel()
			got := searchResultsPath(tt.platform, tt.keyword)
			if got != tt.want {
				t.Errorf("searchResultsPath(%q, %q) = %q, want %q", tt.platform, tt.keyword, got, tt.want)
			}
		})
	}
}

func TestBuildSearchParams_Twitter(t *testing.T) {
	t.Parallel()
	params := buildSearchParams("twitter", "news", "tok", 50, time.Time{}, time.Time{}, nil)
	if params.Get("search_type") != "latest" {
		t.Error("expected search_type=latest for twitter")
	}
	if params.Get("keywords") != "news" {
		t.Error("expected keywords=news")
	}
	if params.Get("max_posts") != "50" {
		t.Error("expected max_posts=50")
	}
}

func TestBuildSearchParams_Facebook(t *testing.T) {
	t.Parallel()
	params := buildSearchParams("facebook", "golang", "tok", 50, time.Time{}, time.Time{}, nil)
	if params.Get("keywords") != "" {
		t.Error("facebook should not have keywords query param")
	}
}

func TestBuildSearchParams_DateFrom(t *testing.T) {
	t.Parallel()
	from := time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC)
	params := buildSearchParams("twitter", "news", "tok", 50, from, time.Time{}, nil)
	if params.Get("from_date") != "2024-01-15" {
		t.Errorf("expected from_date=2024-01-15, got %q", params.Get("from_date"))
	}
	paramsNoDate := buildSearchParams("twitter", "news", "tok", 50, time.Time{}, time.Time{}, nil)
	if paramsNoDate.Get("from_date") != "" {
		t.Error("zero fromDate should not set from_date param")
	}
}

func TestTriggerSearch_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Query().Get("access_token") != "test-token" {
			t.Error("missing access_token")
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	err := client.TriggerSearch(context.Background(), "twitter", "golang", 50, time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestTriggerSearch_404(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	err := client.TriggerSearch(context.Background(), "twitter", "golang", 50, time.Time{}, time.Time{}, nil)
	if err == nil {
		t.Fatal("expected error for 404")
	}
	if !IsUnsupportedSearchError(err) {
		t.Fatalf("expected unsupported search error, got %T", err)
	}
}

func TestTriggerSearch_403(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	err := client.TriggerSearch(context.Background(), "reddit", "golang", 50, time.Time{}, time.Time{}, nil)
	if err == nil {
		t.Fatal("expected error for 403")
	}
	if !IsUnsupportedSearchError(err) {
		t.Fatalf("expected unsupported search error, got %T", err)
	}
}

func TestTriggerSearch_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	err := client.TriggerSearch(context.Background(), "reddit", "test", 50, time.Time{}, time.Time{}, nil)
	if err == nil {
		t.Fatal("expected error for 500")
	}
}

// pollStatus builds the Data365 API envelope for poll responses.
func pollStatus(taskStatus string, taskError string) map[string]interface{} {
	return map[string]interface{}{
		"data":   map[string]interface{}{"status": taskStatus, "error": taskError},
		"status": "ok",
		"error":  nil,
	}
}

func TestPollUntilFinished_ImmediateSuccess(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(pollStatus("finished", ""))
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	client.pollInterval = 1 // 1ns for fast test
	err := client.PollUntilFinished(context.Background(), "twitter", "golang", 50, time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestPollUntilFinished_EventualSuccess(t *testing.T) {
	t.Parallel()
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		n := atomic.AddInt32(&calls, 1)
		if n < 3 {
			json.NewEncoder(w).Encode(pollStatus("pending", ""))
			return
		}
		json.NewEncoder(w).Encode(pollStatus("finished", ""))
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	client.pollInterval = 1 // 1ns
	err := client.PollUntilFinished(context.Background(), "tiktok", "dance", 50, time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if atomic.LoadInt32(&calls) < 3 {
		t.Error("expected at least 3 poll calls")
	}
}

func TestPollUntilFinished_TaskFailed(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(pollStatus("error", "rate limited"))
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	client.pollInterval = 1
	err := client.PollUntilFinished(context.Background(), "reddit", "test", 50, time.Time{}, time.Time{}, nil)
	if err == nil {
		t.Fatal("expected error for failed task")
	}
}

func TestPollUntilFinished_ContextCancelled(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		json.NewEncoder(w).Encode(pollStatus("pending", ""))
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	client.pollInterval = 1

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately
	err := client.PollUntilFinished(ctx, "twitter", "test", 50, time.Time{}, time.Time{}, nil)
	if err == nil {
		t.Fatal("expected context cancelled error")
	}
}

func TestFetchResults_Success(t *testing.T) {
	t.Parallel()
	// Data365 returns results wrapped in {"data": {"items": [...], "page_info": {...}}}.
	// The client extracts Cursor from data.page_info.cursor after decoding.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Errorf("expected GET, got %s", r.Method)
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"data":{"items":[{"id":"1"},{"id":"2"}],"page_info":{"cursor":"next-page","has_next_page":true}}}`))
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	result, err := client.FetchResults(context.Background(), "twitter", "golang", "", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Cursor != "next-page" {
		t.Errorf("expected cursor 'next-page', got %q", result.Cursor)
	}
}

func TestFetchResults_WithCursor(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("cursor") != "page2" {
			t.Error("expected cursor=page2")
		}
		json.NewEncoder(w).Encode(Data365SearchResult{
			Data: json.RawMessage(`[]`),
		})
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	_, err := client.FetchResults(context.Background(), "reddit", "test", "page2", time.Time{}, time.Time{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestFetchResults_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("bad gateway"))
	}))
	defer srv.Close()

	client := newTestData365Client(srv.URL)
	_, err := client.FetchResults(context.Background(), "threads", "test", "", time.Time{}, time.Time{}, nil)
	if err == nil {
		t.Fatal("expected error for 502")
	}
}

func TestAllPlatformSearchPaths(t *testing.T) {
	t.Parallel()
	platforms := []string{"facebook", "instagram", "tiktok", "twitter", "reddit", "threads"}
	for _, p := range platforms {
		t.Run(p, func(t *testing.T) {
			t.Parallel()
			updatePath := searchUpdatePath(p, "test")
			resultsPath := searchResultsPath(p, "test")
			if updatePath == "" {
				t.Error("empty update path")
			}
			if resultsPath == "" {
				t.Error("empty results path")
			}
			// Update path should end with /update
			if updatePath[len(updatePath)-7:] != "/update" {
				t.Errorf("update path should end with /update: %s", updatePath)
			}
		})
	}
}
