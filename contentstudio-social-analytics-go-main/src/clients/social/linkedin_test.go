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

	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func TestNewLinkedInClient(t *testing.T) {
	client := NewLinkedInClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.HTTPClient == nil {
		t.Fatal("expected non-nil HTTPClient")
	}
	if len(client.BaseURL) != 2 {
		t.Fatalf("expected 2 base URLs, got %d", len(client.BaseURL))
	}
	if client.BaseURL["v1"] != "https://api.linkedin.com/v2/" {
		t.Fatalf("expected v1 URL, got '%s'", client.BaseURL["v1"])
	}
	if client.BaseURL["v2"] != "https://api.linkedin.com/rest/" {
		t.Fatalf("expected v2 URL, got '%s'", client.BaseURL["v2"])
	}
}

func TestLinkedInClient_FetchShares(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if r.Header.Get("LinkedIn-Version") != defaultAPIVersion {
			t.Errorf("expected LinkedIn-Version %s", defaultAPIVersion)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{"id": "share_1"},
				{"id": "share_2"},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	shares, err := client.FetchShares(context.Background(), "org_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(shares) != 2 {
		t.Fatalf("expected 2 shares, got %d", len(shares))
	}
}

func TestLinkedInClient_FetchShares_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad Request"))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchShares(context.Background(), "org_123", "token")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestLinkedInClient_FetchPostsPaginated(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		posts := []map[string]interface{}{
			{"id": "post_" + string(rune('0'+callCount)), "createdAt": time.Now().UnixMilli()},
		}
		resp := map[string]interface{}{
			"elements": posts,
			"paging": map[string]interface{}{
				"start": (callCount - 1) * 100,
				"total": 150,
				"count": 100,
			},
		}
		// Only return full page on first call
		if callCount >= 2 {
			resp["elements"] = []map[string]interface{}{}
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	posts, err := client.FetchPostsPaginated(context.Background(), "123", "organization", "token", time.Time{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(posts) != 1 {
		t.Fatalf("expected 1 post, got %d", len(posts))
	}
}

func TestLinkedInClient_FetchPostsPaginated_WithCutoff(t *testing.T) {
	cutoff := time.Now().Add(-7 * 24 * time.Hour)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return one recent post and one old post
		posts := []map[string]interface{}{
			{"id": "post_new", "createdAt": time.Now().UnixMilli()},
			{"id": "post_old", "createdAt": cutoff.Add(-24 * time.Hour).UnixMilli()},
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": posts,
			"paging":   map[string]interface{}{"start": 0, "total": 2, "count": 100},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	posts, err := client.FetchPostsPaginated(context.Background(), "123", "organization", "token", cutoff)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Should only get the new post (old one is before cutoff)
	if len(posts) != 1 {
		t.Fatalf("expected 1 post (cutoff applied), got %d", len(posts))
	}
}

func TestLinkedInClient_FetchPostsPaginated_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return posts that trigger pagination
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": make([]map[string]interface{}, 100),
			"paging":   map[string]interface{}{"start": 0, "total": 200, "count": 100},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.FetchPostsPaginated(ctx, "123", "organization", "token", time.Time{})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestLinkedInClient_joinIDs(t *testing.T) {
	client := NewLinkedInClient()

	result := client.joinIDs("ids", []string{"id1", "id2", "id3"})
	expected := "?ids=id1&ids=id2&ids=id3"
	if result != expected {
		t.Fatalf("expected '%s', got '%s'", expected, result)
	}
}

func TestLinkedInClient_FetchImagesRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"img_1": map[string]interface{}{"downloadUrl": "https://example.com/img1.jpg"},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	body, err := client.FetchImagesRaw(context.Background(), []string{"img_1"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestLinkedInClient_FetchImagesRaw_EmptyIDs(t *testing.T) {
	client := NewLinkedInClient()

	body, err := client.FetchImagesRaw(context.Background(), []string{}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != nil {
		t.Fatal("expected nil body for empty IDs")
	}
}

func TestLinkedInClient_FetchVideosRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"vid_1": map[string]interface{}{"downloadUrl": "https://example.com/vid1.mp4"},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	body, err := client.FetchVideosRaw(context.Background(), []string{"vid_1"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestLinkedInClient_FetchVideosRaw_EmptyIDs(t *testing.T) {
	client := NewLinkedInClient()

	body, err := client.FetchVideosRaw(context.Background(), []string{}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != nil {
		t.Fatal("expected nil body for empty IDs")
	}
}

func TestLinkedInClient_makeRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify headers
		if r.Header.Get("Authorization") != "Bearer test_token" {
			t.Error("expected Authorization header")
		}
		if r.Header.Get("LinkedIn-Version") != defaultAPIVersion {
			t.Errorf("expected LinkedIn-Version %s", defaultAPIVersion)
		}
		if r.Header.Get("Custom-Header") != "custom_value" {
			t.Error("expected Custom-Header")
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success": true}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()

	body, status, err := client.makeRequest(
		context.Background(),
		server.URL,
		"test_token",
		map[string]string{"Custom-Header": "custom_value"},
		false,
	)

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

func TestLinkedInClient_FetchFollowerData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"firstDegreeSize": 10000,
			"elements": []map[string]interface{}{
				{
					"followerCountsBySeniority": []map[string]interface{}{
						{"seniority": "SENIOR", "followerCounts": map[string]int{"organicFollowerCount": 500}},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	body, err := client.FetchFollowerData(context.Background(), "org_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestLinkedInClient_FetchPageStatisticsRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"totalPageStatistics": map[string]interface{}{
						"views": map[string]interface{}{
							"allPageViews": map[string]int{"pageViews": 1000},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	startMs := time.Now().Add(-7 * 24 * time.Hour).UnixMilli()
	endMs := time.Now().UnixMilli()

	body, err := client.FetchPageStatisticsRaw(context.Background(), "org_123", "token", startMs, endMs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestLinkedInClient_FetchShareStatisticsRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"totalShareStatistics": map[string]interface{}{
						"shareCount":             100,
						"uniqueImpressionsCount": 5000,
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	startMs := time.Now().Add(-7 * 24 * time.Hour).UnixMilli()
	endMs := time.Now().UnixMilli()

	body, err := client.FetchShareStatisticsRaw(context.Background(), "org_123", "token", startMs, endMs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestLinkedInClient_FetchOrganizationDetailsRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":              12345,
			"localizedName":   "Test Company",
			"vanityName":      "testcompany",
			"staffCountRange": "SIZE_51_200",
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	body, err := client.FetchOrganizationDetailsRaw(context.Background(), "org_123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestLinkedInClient_ResolveGeoIDs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"urn:li:geo:100": map[string]interface{}{
					"defaultLocalizedName": map[string]interface{}{
						"value": "United States",
					},
				},
				"urn:li:geo:101": map[string]interface{}{
					"defaultLocalizedName": map[string]interface{}{
						"value": "Canada",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	// ResolveGeoIDs uses BaseURL["v1"] for the geo API
	client.BaseURL["v1"] = server.URL + "/"

	result, err := client.ResolveGeoIDs(context.Background(), []string{"100", "101"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestLinkedInClient_ResolveGeoIDs_EmptyIDs(t *testing.T) {
	client := NewLinkedInClient()

	result, err := client.ResolveGeoIDs(context.Background(), []string{}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result (empty map)")
	}
}

func TestLinkedInClient_FetchStatsRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"share": "urn:li:share:123",
					"totalShareStatistics": map[string]interface{}{
						"shareCount": 50,
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	body, err := client.FetchStatsRaw(context.Background(), "org_123", []string{"ugc_123"}, []string{"share_123"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestLinkedInClient_FetchDocumentsRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"doc_1": map[string]interface{}{"downloadUrl": "https://example.com/doc1.pdf"},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	body, err := client.FetchDocumentsRaw(context.Background(), []string{"doc_1"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(body) == 0 {
		t.Fatal("expected non-empty body")
	}
}

func TestLinkedInClient_FetchDocumentsRaw_EmptyIDs(t *testing.T) {
	client := NewLinkedInClient()

	body, err := client.FetchDocumentsRaw(context.Background(), []string{}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if body != nil {
		t.Fatal("expected nil body for empty IDs")
	}
}

func TestLinkedInClient_GetPostURLs(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/posts"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"elements": []map[string]any{
					{
						"id": "urn:li:ugcPost:post_1",
						"content": map[string]any{
							"multiImage": map[string]any{
								"images": []map[string]any{
									{"id": "urn:li:image:img_1"},
								},
							},
						},
					},
					{
						"id": "urn:li:ugcPost:post_2",
						"content": map[string]any{
							"media": map[string]any{
								"id": "urn:li:video:vid_1",
							},
						},
					},
					{
						"id": "urn:li:ugcPost:post_3",
						"content": map[string]any{
							"media": map[string]any{
								"id": "urn:li:document:doc_1",
							},
						},
					},
				},
				"paging": map[string]any{"start": 0, "total": 3, "count": 100},
			})
		case strings.Contains(r.URL.Path, "/images"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": map[string]any{
					"urn:li:image:img_1": map[string]any{
						"id":          "urn:li:image:img_1",
						"downloadUrl": "https://example.com/img_1.jpg",
					},
				},
			})
		case strings.Contains(r.URL.Path, "/videos"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": map[string]any{
					"urn:li:video:vid_1": map[string]any{
						"id":        "urn:li:video:vid_1",
						"thumbnail": "https://example.com/vid_1.jpg",
					},
				},
			})
		case strings.Contains(r.URL.Path, "/documents"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"results": map[string]any{
					"urn:li:document:doc_1": map[string]any{
						"id":          "urn:li:document:doc_1",
						"downloadUrl": "https://example.com/doc_1.pdf",
					},
				},
			})
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	posts := []clickhousemodels.LinkedInMinimalPost{
		{LinkedinID: "li_123", PostID: "post_1", Activity: "urn:li:ugcPost:post_1"},
		{LinkedinID: "li_123", PostID: "post_2", Activity: "urn:li:ugcPost:post_2"},
		{LinkedinID: "li_123", PostID: "post_3", Activity: "urn:li:ugcPost:post_3"},
	}

	refreshed, err := client.GetPostURLs(context.Background(), "li_123", "organization", "token", posts)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(refreshed) != 3 {
		t.Fatalf("expected 3 refreshed posts, got %d", len(refreshed))
	}
	if refreshed[0].Image != "https://example.com/img_1.jpg" || len(refreshed[0].Media) != 1 {
		t.Fatalf("unexpected image refresh result: %+v", refreshed[0])
	}
	if refreshed[1].Image != "https://example.com/vid_1.jpg" || refreshed[1].Media[0] != "https://example.com/vid_1.jpg" {
		t.Fatalf("unexpected video refresh result: %+v", refreshed[1])
	}
	if refreshed[2].Image != "https://example.com/doc_1.pdf" || refreshed[2].Media[0] != "https://example.com/doc_1.pdf" {
		t.Fatalf("unexpected document refresh result: %+v", refreshed[2])
	}
}

func TestLinkedInClient_GetPostURLs_FetchAssetError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/posts"):
			_ = json.NewEncoder(w).Encode(map[string]any{
				"elements": []map[string]any{
					{
						"id": "urn:li:ugcPost:post_1",
						"content": map[string]any{
							"media": map[string]any{
								"id": "urn:li:video:vid_1",
							},
						},
					},
				},
				"paging": map[string]any{"start": 0, "total": 1, "count": 100},
			})
		case strings.Contains(r.URL.Path, "/videos"):
			http.Error(w, "bad upstream", http.StatusBadGateway)
		default:
			_ = json.NewEncoder(w).Encode(map[string]any{"results": map[string]any{}})
		}
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.GetPostURLs(context.Background(), "li_123", "organization", "token", []clickhousemodels.LinkedInMinimalPost{
		{LinkedinID: "li_123", PostID: "post_1", Activity: "urn:li:ugcPost:post_1"},
	})
	if err == nil {
		t.Fatal("expected error when asset fetch fails")
	}
}

func TestConstants(t *testing.T) {
	if defaultAPIVersion == "" {
		t.Fatal("expected non-empty defaultAPIVersion")
	}
	if restliHeaderVersion == "" {
		t.Fatal("expected non-empty restliHeaderVersion")
	}
}

func TestLinkedInClient_GetGeoIDsFromFollowerStatsRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"followerCountsByGeoCountry": []map[string]interface{}{
						{"geo": "urn:li:geo:103644278", "count": 100},
						{"geo": "urn:li:geo:101174742", "count": 50},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	ids, err := client.GetGeoIDsFromFollowerStatsRaw(context.Background(), "org123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("expected non-empty geo IDs")
	}
}

func TestLinkedInClient_GetGeoIDsFromFollowerStatsRaw_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Invalid token",
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.GetGeoIDsFromFollowerStatsRaw(context.Background(), "org123", "token")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestLinkedInClient_GetGeoIDsWithTypeFromFollowerStatsRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"followerCountsByGeoCountry": []map[string]interface{}{
						{"geo": "urn:li:geo:103644278", "count": 100},
					},
					"followerCountsByGeo": []map[string]interface{}{
						{"geo": "urn:li:geo:101174742", "count": 50},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	ids, err := client.GetGeoIDsWithTypeFromFollowerStatsRaw(context.Background(), "org123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(ids) == 0 {
		t.Fatal("expected non-empty geo IDs with type")
	}
}

func TestLinkedInClient_FetchMemberCreatorPostAnalyticsRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{"impressionCount": 1000, "shareCount": 50},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchMemberCreatorPostAnalyticsRaw(context.Background(), "token", "IMPRESSION", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestLinkedInClient_FetchMemberCreatorPostAnalyticsRaw_WithDates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify date range is included in request
		if !strings.Contains(r.URL.RawQuery, "dateRange") {
			t.Error("expected dateRange in request")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	startDate := time.Now().Add(-7 * 24 * time.Hour)
	endDate := time.Now()

	_, err := client.FetchMemberCreatorPostAnalyticsRaw(context.Background(), "token", "IMPRESSION", &startDate, &endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLinkedInClient_FetchMemberCreatorPostAnalyticsRaw_MembersReached(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// MEMBERS_REACHED should NOT have DAILY aggregation
		if strings.Contains(r.URL.RawQuery, "aggregation=DAILY") {
			t.Error("MEMBERS_REACHED should not have DAILY aggregation")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	startDate := time.Now().Add(-7 * 24 * time.Hour)
	endDate := time.Now()

	_, err := client.FetchMemberCreatorPostAnalyticsRaw(context.Background(), "token", "MEMBERS_REACHED", &startDate, &endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLinkedInClient_FetchMemberFollowersCountRaw(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"firstDegreeSize": 5000,
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchMemberFollowersCountRaw(context.Background(), "token", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestLinkedInClient_FetchMemberFollowersCountRaw_WithDateRange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify date range is included
		if !strings.Contains(r.URL.RawQuery, "dateRange") {
			t.Error("expected dateRange in request")
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	startDate := time.Now().Add(-30 * 24 * time.Hour)
	endDate := time.Now()

	_, err := client.FetchMemberFollowersCountRaw(context.Background(), "token", &startDate, &endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLinkedInClient_BuildFollowerDataWithGeoNames(t *testing.T) {
	followerData := []byte(`{
		"paging": {},
		"elements": [{
			"followerCountsByGeoCountry": [
				{"geo": "urn:li:geo:103644278", "count": 100}
			]
		}]
	}`)

	client := NewLinkedInClient()

	stats := &FollowerStatsWithGeoIDs{
		RawStats:   followerData,
		GeoIDs:     []GeoIDWithType{{ID: "103644278", Type: "country"}},
		TotalCount: 100,
	}

	geoNames := map[string]string{
		"103644278": "United States",
	}

	result, err := client.BuildFollowerDataWithGeoNames(stats, geoNames)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty result")
	}
}

func TestLinkedInClient_BuildFollowerDataWithGeoNames_InvalidJSON(t *testing.T) {
	client := NewLinkedInClient()

	stats := &FollowerStatsWithGeoIDs{
		RawStats: []byte(`not json`),
	}

	_, err := client.BuildFollowerDataWithGeoNames(stats, map[string]string{})
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestLinkedInClient_FetchFollowerStatsWithGeoIDs(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if strings.Contains(r.URL.Path, "networkSizes") {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"firstDegreeSize": 1000,
			})
		} else {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"paging": map[string]interface{}{},
				"elements": []map[string]interface{}{
					{
						"followerCountsByGeoCountry": []map[string]interface{}{
							{"geo": "urn:li:geo:100", "count": 500},
						},
					},
				},
			})
		}
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v1"] = server.URL + "/"
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchFollowerStatsWithGeoIDs(context.Background(), "org123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_makeRequest_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		json.NewEncoder(w).Encode(map[string]interface{}{})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, _, err := client.makeRequest(ctx, server.URL, "token", nil, false)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

// Note: extractGeoIDsFromElements and extractGeoIDsWithTypeFromElements are unexported
// functions and are tested indirectly through GetGeoIDsFromFollowerStatsRaw and
// GetGeoIDsWithTypeFromFollowerStatsRaw

func TestLinkedInClient_FetchFollowerDataWithGeoNames_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid token"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v1"] = server.URL + "/"
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchFollowerDataWithGeoNames(context.Background(), "org123", "token", nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchOrganizationDetailsRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "Forbidden"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchOrganizationDetailsRaw(context.Background(), "org123", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchPageStatisticsRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Bad request"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchPageStatisticsRaw(context.Background(), "org123", "token", 0, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchShareStatisticsRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		w.Write([]byte(`{"message": "Rate limited"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchShareStatisticsRaw(context.Background(), "org123", "token", 0, 0)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchMemberCreatorPostAnalyticsRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"message": "Service unavailable"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchMemberCreatorPostAnalyticsRaw(context.Background(), "token", "IMPRESSION", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchMemberFollowersCountRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte(`{"message": "Bad gateway"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchMemberFollowersCountRaw(context.Background(), "token", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchImagesRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid token"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchImagesRaw(context.Background(), []string{"img1", "img2"}, "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchVideosRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "Forbidden"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchVideosRaw(context.Background(), []string{"vid1"}, "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchDocumentsRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"message": "Bad request"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchDocumentsRaw(context.Background(), []string{"doc1"}, "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchStatsRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
		w.Write([]byte(`{"message": "Service unavailable"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchStatsRaw(context.Background(), "org123", []string{"ugc1"}, []string{"share1"}, "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchShares_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"message": "Invalid token"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v1"] = server.URL + "/"
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchShares(context.Background(), "org123", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchPostsPaginated_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(`{"message": "Forbidden"}`))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchPostsPaginated(context.Background(), "org123", "ORGANIZATION", "token", time.Now())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchFollowerDataWithGeoNames_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"followerCountsByGeoCountry": []map[string]interface{}{
						{"geo": "urn:li:geo:100", "count": 500},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v1"] = server.URL + "/"
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchFollowerDataWithGeoNames(context.Background(), "org123", "token", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchShares_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{"id": "share1", "text": "Test share"},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v1"] = server.URL + "/"
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchShares(context.Background(), "org123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 share, got %d", len(result))
	}
}

func TestLinkedInClient_FetchPostsPaginated_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{"id": "post1", "commentary": "Test post"},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchPostsPaginated(context.Background(), "org123", "ORGANIZATION", "token", time.Now().Add(-24*time.Hour))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
}

func TestLinkedInClient_FetchImagesRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"img1": map[string]interface{}{"downloadUrl": "http://example.com/img.jpg"},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchImagesRaw(context.Background(), []string{"img1"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchVideosRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"vid1": map[string]interface{}{"downloadUrl": "http://example.com/video.mp4"},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchVideosRaw(context.Background(), []string{"vid1"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchDocumentsRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"doc1": map[string]interface{}{"downloadUrl": "http://example.com/doc.pdf"},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchDocumentsRaw(context.Background(), []string{"doc1"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchStatsRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{"totalShareStatistics": map[string]interface{}{"impressionCount": 100}},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchStatsRaw(context.Background(), "org123", []string{"ugc1"}, []string{"share1"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchOrganizationDetailsRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":            "org123",
			"localizedName": "Test Organization",
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchOrganizationDetailsRaw(context.Background(), "org123", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchPageStatisticsRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{"totalPageStatistics": map[string]interface{}{"views": 1000}},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchPageStatisticsRaw(context.Background(), "org123", "token", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchShareStatisticsRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{"totalShareStatistics": map[string]interface{}{"shareCount": 50}},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchShareStatisticsRaw(context.Background(), "org123", "token", 0, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchMemberCreatorPostAnalyticsRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{"impressions": 500},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchMemberCreatorPostAnalyticsRaw(context.Background(), "token", "IMPRESSION", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchMemberFollowersCountRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{"followerCount": 1000},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchMemberFollowersCountRaw(context.Background(), "token", nil, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_fetchFollowers_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"firstDegreeSize": 5000,
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.fetchFollowers(context.Background(), "123456", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_fetchFollowers_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.fetchFollowers(context.Background(), "123456", "token")
	if err == nil {
		t.Fatal("expected error for unauthorized")
	}
}

func TestLinkedInClient_fetchFollowerStatsRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"followerCountsByAssociationType": []map[string]interface{}{
						{"associationType": "COMPANY", "followerCounts": map[string]interface{}{"organicFollowerCount": 100}},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.fetchFollowerStatsRaw(context.Background(), "123456", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_fetchFollowerStatsRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte("Internal error"))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.fetchFollowerStatsRaw(context.Background(), "123456", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_ResolveGeoIDs_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"100": map[string]interface{}{
					"defaultLocalizedName": map[string]interface{}{
						"value": "United States",
					},
				},
				"200": map[string]interface{}{
					"defaultLocalizedName": map[string]interface{}{
						"value": "Canada",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v1"] = server.URL + "/"

	result, err := client.ResolveGeoIDs(context.Background(), []string{"100", "200"}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 geo mappings, got %d", len(result))
	}
	if result["100"] != "United States" {
		t.Fatalf("expected 'United States', got '%s'", result["100"])
	}
}

func TestLinkedInClient_ResolveGeoIDs_Empty(t *testing.T) {
	client := NewLinkedInClient()

	result, err := client.ResolveGeoIDs(context.Background(), []string{}, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 geo mappings, got %d", len(result))
	}
}

func TestLinkedInClient_ResolveGeoIDs_LargeBatch(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		json.NewEncoder(w).Encode(map[string]interface{}{
			"results": map[string]interface{}{
				"1": map[string]interface{}{
					"defaultLocalizedName": map[string]interface{}{
						"value": "Location",
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v1"] = server.URL + "/"

	// Create 150 geo IDs to trigger 2 batches (batch size is 100)
	geoIDs := make([]string, 150)
	for i := 0; i < 150; i++ {
		geoIDs[i] = "1"
	}

	_, err := client.ResolveGeoIDs(context.Background(), geoIDs, "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if callCount != 2 {
		t.Fatalf("expected 2 API calls for 150 IDs, got %d", callCount)
	}
}

func TestLinkedInClient_resolveGeoIDsBatch_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("Bad request"))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v1"] = server.URL + "/"

	_, err := client.resolveGeoIDsBatch(context.Background(), []string{"100"}, "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_FetchFollowerStatsWithGeoIDs_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if strings.Contains(r.URL.Path, "networkSizes") {
			// Return follower count
			json.NewEncoder(w).Encode(map[string]interface{}{
				"firstDegreeSize": 5000,
			})
		} else {
			// Return follower stats
			json.NewEncoder(w).Encode(map[string]interface{}{
				"elements": []map[string]interface{}{
					{
						"followerCountsByGeoCountry": []map[string]interface{}{
							{"geo": "urn:li:geo:100", "followerCounts": map[string]interface{}{"organicFollowerCount": 500}},
						},
					},
				},
			})
		}
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.FetchFollowerStatsWithGeoIDs(context.Background(), "123456", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
}

func TestLinkedInClient_FetchFollowerStatsWithGeoIDs_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte("Unauthorized"))
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.FetchFollowerStatsWithGeoIDs(context.Background(), "123456", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_GetGeoIDsWithTypeFromFollowerStatsRaw_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"elements": []map[string]interface{}{
				{
					"followerCountsByGeoCountry": []map[string]interface{}{
						{"geo": "urn:li:geo:100"},
					},
					"followerCountsByGeo": []map[string]interface{}{
						{"geo": "urn:li:geo:200"},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	result, err := client.GetGeoIDsWithTypeFromFollowerStatsRaw(context.Background(), "123456", "token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 geo IDs, got %d", len(result))
	}
}

func TestLinkedInClient_GetGeoIDsWithTypeFromFollowerStatsRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.GetGeoIDsWithTypeFromFollowerStatsRaw(context.Background(), "123456", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestLinkedInClient_GetGeoIDsFromFollowerStatsRaw_Error(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewLinkedInClient()
	client.BaseURL["v2"] = server.URL + "/"

	_, err := client.GetGeoIDsFromFollowerStatsRaw(context.Background(), "123456", "token")
	if err == nil {
		t.Fatal("expected error")
	}
}

// Tests for IsExpectedCompetitorErrorLI function
func TestIsExpectedCompetitorErrorLI(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{"nil error", nil, false},
		{"EXPIRED_ACCESS_TOKEN", errors.New("EXPIRED_ACCESS_TOKEN"), true},
		{"INVALID_POST_FINDER_AUTHOR_ENTITY_TYPE", errors.New("INVALID_POST_FINDER_AUTHOR_ENTITY_TYPE"), true},
		{"The token used in the request has expired", errors.New("The token used in the request has expired"), true},
		{"token invalid or expired", errors.New("token invalid or expired"), true},
		{"status 401", errors.New("status 401 unauthorized"), true},
		{"status 403", errors.New("status 403 forbidden"), true},
		{"status 400", errors.New("status 400 bad request"), true},
		{"unauthorized lowercase", errors.New("unauthorized access attempt"), true},
		{"permission lowercase", errors.New("permission denied"), true},
		{"not authorized", errors.New("not authorized to access this resource"), true},
		{"network error", errors.New("network timeout"), false},
		{"parse error", errors.New("failed to parse json"), false},
		{"status 500", errors.New("internal server error status 500"), false},
		{"generic error", errors.New("unknown error"), false},
		{"rate limit", errors.New("rate limited - try again later"), false},
		{"access denied no auth", errors.New("access denied"), false},
		{"Expired Token uppercase", errors.New("Expired Token"), false},
		{"permission word separate", errors.New("access_forbidden_details"), false},
		{"Case sensitive match", errors.New("EXPIRED_ACCESS_TOKEN is invalid"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsExpectedCompetitorErrorLI(tt.err)
			if got != tt.expected {
				t.Errorf("IsExpectedCompetitorErrorLI() = %v, want %v for error: %v", got, tt.expected, tt.err)
			}
		})
	}
}
