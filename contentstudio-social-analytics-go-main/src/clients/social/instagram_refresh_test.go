package social

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	apimodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

func TestInstagramClient_GetMediaURLs(t *testing.T) {
	var requestCount int
	var seenFields string
	var seenToken string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++
		seenFields = r.URL.Query().Get("fields")
		seenToken = r.URL.Query().Get("access_token")

		switch path.Base(r.URL.Path) {
		case "media_1":
			_ = json.NewEncoder(w).Encode(kafkamodels.RawInstagramMedia{
				ID:           "media_1",
				MediaType:    "IMAGE",
				MediaURL:     "https://example.com/media_1.jpg",
				ThumbnailURL: "https://example.com/media_1-thumb.jpg",
			})
		case "media_2":
			_ = json.NewEncoder(w).Encode(kafkamodels.RawInstagramMedia{
				ID:           "media_2",
				MediaType:    "VIDEO",
				MediaURL:     "https://example.com/media_2.mp4",
				ThumbnailURL: "https://example.com/media_2-thumb.jpg",
				Children: struct {
					Data []kafkamodels.InstagramChild `json:"data"`
				}{
					Data: []kafkamodels.InstagramChild{
						{MediaType: "IMAGE", MediaURL: "https://example.com/media_2-child-1.jpg"},
						{MediaType: "VIDEO", MediaURL: "https://example.com/media_2-child-2.mp4", ThumbnailURL: "https://example.com/media_2-child-2-thumb.jpg"},
					},
				},
			})
		default:
			t.Fatalf("unexpected media id: %s", path.Base(r.URL.Path))
		}
	}))
	defer server.Close()

	client := NewInstagramClient("app-secret")
	client.WithBaseURL(server.URL + "/")

	posts := []clickhousemodels.InstagramMinimalPost{
		{MediaID: "media_1"},
		{MediaID: ""},
		{MediaID: "media_1"},
		{MediaID: "media_2"},
	}

	got, err := client.GetMediaURLs(context.Background(), "ig_123", "access-token", posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if requestCount != 2 {
		t.Fatalf("expected 2 unique requests, got %d", requestCount)
	}
	if seenFields != igRefreshFields {
		t.Fatalf("unexpected fields query: %s", seenFields)
	}
	if seenToken != "access-token" {
		t.Fatalf("unexpected access token query: %s", seenToken)
	}

	if len(got) != 2 {
		t.Fatalf("expected 2 refreshed posts, got %d", len(got))
	}

	if got[0].MediaID != "media_1" || !reflect.DeepEqual(got[0].MediaURL, []string{"https://example.com/media_1.jpg"}) || len(got[0].VideoURL) != 0 {
		t.Fatalf("unexpected first result: %+v", got[0])
	}

	wantSecond := clickhousemodels.InstagramMinimalPost{
		InstagramID: "ig_123",
		MediaID:     "media_2",
		MediaURL: []string{
			"https://example.com/media_2-thumb.jpg",
			"https://example.com/media_2-child-1.jpg",
			"https://example.com/media_2-child-2-thumb.jpg",
		},
		VideoURL: []string{
			"https://example.com/media_2.mp4",
			"https://example.com/media_2-child-2.mp4",
		},
	}

	if got[1].InstagramID != wantSecond.InstagramID || got[1].MediaID != wantSecond.MediaID || !reflect.DeepEqual(got[1].MediaURL, wantSecond.MediaURL) || !reflect.DeepEqual(got[1].VideoURL, wantSecond.VideoURL) {
		t.Fatalf("unexpected second result: got %+v want %+v", got[1], wantSecond)
	}
}

func TestInstagramClient_GetMediaURLs_EmptyAccessToken(t *testing.T) {
	client := NewInstagramClient("app-secret")

	got, err := client.GetMediaURLs(context.Background(), "ig_123", "", nil)
	if err == nil {
		t.Fatal("expected error for empty access token")
	}
	if got != nil {
		t.Fatalf("expected nil result, got %+v", got)
	}
}

func TestInstagramClient_GetMediaURLs_SkipsInaccessibleMediaIDs(t *testing.T) {
	var requestCount int

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount++

		switch path.Base(r.URL.Path) {
		case "bad_media":
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]interface{}{
					"message": "Unsupported get request. Object with ID 'bad_media' does not exist, cannot be loaded due to missing permissions, or does not support this operation.",
					"type":    "GraphMethodException",
					"code":    100,
				},
			})
		case "good_media":
			_ = json.NewEncoder(w).Encode(kafkamodels.RawInstagramMedia{
				ID:        "good_media",
				MediaType: "IMAGE",
				MediaURL:  "https://example.com/good.jpg",
			})
		default:
			t.Fatalf("unexpected media id: %s", path.Base(r.URL.Path))
		}
	}))
	defer server.Close()

	client := NewInstagramClient("app-secret")
	client.WithBaseURL(server.URL + "/")

	posts := []clickhousemodels.InstagramMinimalPost{
		{MediaID: "bad_media"},
		{MediaID: "good_media"},
	}

	got, err := client.GetMediaURLs(context.Background(), "ig_123", "access-token", posts)
	if err != nil {
		t.Fatalf("expected skip behavior, got error: %v", err)
	}

	if requestCount != 2 {
		t.Fatalf("expected 2 requests, got %d", requestCount)
	}
	// Both media IDs are returned: good_media with real URLs, bad_media with empty URLs (to clear
	// it from ClickHouse so it is never retried again).
	if len(got) != 2 {
		t.Fatalf("expected 2 results (good + cleared bad), got %d", len(got))
	}
	var good, bad *clickhousemodels.InstagramMinimalPost
	for i := range got {
		switch got[i].MediaID {
		case "good_media":
			good = &got[i]
		case "bad_media":
			bad = &got[i]
		}
	}
	if good == nil || len(good.MediaURL) == 0 {
		t.Fatalf("expected good_media to have a URL, got %+v", good)
	}
	if bad == nil || len(bad.MediaURL) != 0 || len(bad.VideoURL) != 0 {
		t.Fatalf("expected bad_media to have empty URLs, got %+v", bad)
	}
}

func TestInstagramClient_GetMediaURLs_FastFailOnTokenError(t *testing.T) {
	var requestCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestCount.Add(1)
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"error": map[string]interface{}{
				"message": "Invalid OAuth access token - Cannot parse access token",
				"type":    "OAuthException",
				"code":    190,
			},
		})
	}))
	defer server.Close()

	client := NewInstagramClient("app-secret")
	client.WithBaseURL(server.URL + "/")

	// 20 posts — without fast-fail all 20 would be attempted
	posts := make([]clickhousemodels.InstagramMinimalPost, 20)
	for i := range posts {
		posts[i].MediaID = fmt.Sprintf("media_%d", i)
	}

	_, err := client.GetMediaURLs(context.Background(), "ig_123", "bad-token", posts)
	if err == nil {
		t.Fatal("expected an error for invalid token, got nil")
	}
	if !strings.Contains(err.Error(), "OAuthException") {
		t.Fatalf("expected OAuthException error, got: %v", err)
	}
	// With fast-fail via context cancellation, far fewer than 20 requests should be made.
	// Without fast-fail all 20 would run; with it only the first concurrent batch (~5) fires.
	if got := requestCount.Load(); got >= 20 {
		t.Fatalf("expected fast-fail (fewer than 20 requests), got %d", got)
	}
}

// TestGetCompetitorMediaURLs_StopsAtOldestTargetDate verifies that pagination stops as soon as the
// business discovery feed returns a post whose timestamp is older than the oldest stale target.
//
// Setup:
//   - targets: [post_target_ghost] — a page we're looking for that never appears in the API feed
//   - Page 1: [post_new (2d), another_post (3d)]  — neither matches; cursor to page 2
//   - Page 2: [post_old (10d)] — older than oldest target (5d) → pagination must stop here
//
// Expected: 2 pages served, empty result (target was never in feed; we stop before exhausting cursor).
func TestGetCompetitorMediaURLs_StopsAtOldestTargetDate(t *testing.T) {
	now := time.Now().UTC()
	layout := "2006-01-02T15:04:05-0700"

	oldestCreatedAt := now.Add(-5 * 24 * time.Hour)

	page1 := apimodels.InstagramBusinessDiscoveryResponse{
		BusinessDiscovery: apimodels.BusinessDiscovery{
			ProfilePictureURL: "https://example.com/profile.jpg",
			Media: &apimodels.MediaPaging{
				Data: []apimodels.InstagramMedia{
					{ID: "post_new", MediaURL: "https://example.com/new.jpg", MediaType: "IMAGE", Timestamp: now.Add(-2 * 24 * time.Hour).Format(layout)},
					{ID: "another_post", MediaURL: "https://example.com/another.jpg", MediaType: "IMAGE", Timestamp: now.Add(-3 * 24 * time.Hour).Format(layout)},
				},
				Paging: &apimodels.InstagramPaging{
					Cursors: &apimodels.Cursors{After: "cursor_page_2"},
				},
			},
		},
	}

	page2 := apimodels.InstagramBusinessDiscoveryResponse{
		BusinessDiscovery: apimodels.BusinessDiscovery{
			Media: &apimodels.MediaPaging{
				Data: []apimodels.InstagramMedia{
					// 10 days old — older than oldest target (5d) → triggers date-stop
					{ID: "post_old", MediaURL: "https://example.com/old.jpg", MediaType: "IMAGE", Timestamp: now.Add(-10 * 24 * time.Hour).Format(layout)},
				},
			},
		},
	}

	var pagesServed int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pagesServed++
		var resp apimodels.InstagramBusinessDiscoveryResponse
		if pagesServed == 1 {
			resp = page1
		} else {
			resp = page2
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewInstagramClient("secret")
	client.WithBaseURL(server.URL + "/")

	posts := []clickhousemodels.InstagramCompetitorMinimalPost{
		{InstagramID: 42, PostID: "post_target_ghost", MediaURL: "https://example.com/stale.jpg", CreatedAt: oldestCreatedAt},
	}

	got, _, err := client.GetCompetitorMediaURLs(
		context.Background(), "competitor_username", posts, "token", "biz_123",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Page 1 has no target; page 2 triggers the date-stop. Exactly 2 pages should be served.
	if pagesServed != 2 {
		t.Fatalf("expected 2 pages served (date-stop on page 2), got %d", pagesServed)
	}
	// The target post was never in the feed, so the result is empty.
	if len(got) != 0 {
		t.Fatalf("expected 0 refreshed posts (target not in feed), got %d: %+v", len(got), got)
	}
}

// TestGetCompetitorMediaURLs_StopsWhenAllFound verifies that when all targets are found before
// the date boundary, pagination stops immediately without requesting further pages.
func TestGetCompetitorMediaURLs_StopsWhenAllFound(t *testing.T) {
	now := time.Now().UTC()
	layout := "2006-01-02T15:04:05-0700"

	// Both targets are recent (1 and 2 days old). oldestTarget → 2 days.
	// Page 1 returns both targets (timestamps 1d and 2d), neither is before oldestTarget (2d sec).
	// After processing page 1, len(seen)==len(targets) → break before page 2 is ever fetched.
	page1 := apimodels.InstagramBusinessDiscoveryResponse{
		BusinessDiscovery: apimodels.BusinessDiscovery{
			ProfilePictureURL: "https://example.com/profile.jpg",
			Media: &apimodels.MediaPaging{
				Data: []apimodels.InstagramMedia{
					{ID: "post_a", MediaURL: "https://example.com/a.jpg", MediaType: "IMAGE", Timestamp: now.Add(-1 * 24 * time.Hour).Format(layout)},
					{ID: "post_b", MediaURL: "https://example.com/b.jpg", MediaType: "IMAGE", Timestamp: now.Add(-2 * 24 * time.Hour).Format(layout)},
				},
				Paging: &apimodels.InstagramPaging{
					Cursors: &apimodels.Cursors{After: "cursor_never_used"},
				},
			},
		},
	}

	var pagesServed int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pagesServed++
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(page1)
	}))
	defer server.Close()

	client := NewInstagramClient("secret")
	client.WithBaseURL(server.URL + "/")

	posts := []clickhousemodels.InstagramCompetitorMinimalPost{
		{InstagramID: 1, PostID: "post_a", CreatedAt: now.Add(-1 * 24 * time.Hour)},
		{InstagramID: 1, PostID: "post_b", CreatedAt: now.Add(-2 * 24 * time.Hour)},
	}

	got, _, err := client.GetCompetitorMediaURLs(
		context.Background(), "competitor", posts, "token", "biz",
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pagesServed != 1 {
		t.Fatalf("expected exactly 1 page (all targets found), got %d", pagesServed)
	}
	if len(got) != 2 {
		t.Fatalf("expected 2 refreshed posts, got %d", len(got))
	}
}

// TestGetCompetitorMediaURLs_EmptyPosts verifies that an empty input slice returns immediately
// with nil results and no HTTP calls.
func TestGetCompetitorMediaURLs_EmptyPosts(t *testing.T) {
	var pagesServed int
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pagesServed++
	}))
	defer server.Close()

	client := NewInstagramClient("secret")
	client.WithBaseURL(server.URL + "/")

	got, pic, err := client.GetCompetitorMediaURLs(context.Background(), "user", nil, "token", "biz")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != nil || pic != "" {
		t.Fatalf("expected nil/empty results for empty posts, got %v / %q", got, pic)
	}
	if pagesServed != 0 {
		t.Fatalf("expected no HTTP calls for empty posts, got %d", pagesServed)
	}
}
