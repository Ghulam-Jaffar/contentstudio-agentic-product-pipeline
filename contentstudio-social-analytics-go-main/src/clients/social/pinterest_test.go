package social

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewPinterestClient(t *testing.T) {
	client := NewPinterestClient()
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.HTTPClient == nil {
		t.Fatal("expected non-nil HTTPClient")
	}
	if client.RateLimiter == nil {
		t.Fatal("expected non-nil RateLimiter")
	}
	if client.MaxRetries != defaultPinterestMaxRetries {
		t.Fatalf("expected MaxRetries %d, got %d", defaultPinterestMaxRetries, client.MaxRetries)
	}
	if client.BaseBackoff != defaultPinterestBaseBackoff {
		t.Fatalf("expected BaseBackoff %v, got %v", defaultPinterestBaseBackoff, client.BaseBackoff)
	}
	if client.MaxBackoff != defaultPinterestMaxBackoff {
		t.Fatalf("expected MaxBackoff %v, got %v", defaultPinterestMaxBackoff, client.MaxBackoff)
	}
}

func TestNewPinterestClientWithConfig(t *testing.T) {
	cfg := PinterestClientConfig{
		RPS:         2.0,
		Burst:       5,
		MaxRetries:  5,
		BaseBackoff: 2 * time.Second,
		MaxBackoff:  20 * time.Second,
	}

	client := NewPinterestClientWithConfig(cfg)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.MaxRetries != 5 {
		t.Fatalf("expected MaxRetries 5, got %d", client.MaxRetries)
	}
	if client.BaseBackoff != 2*time.Second {
		t.Fatalf("expected BaseBackoff 2s, got %v", client.BaseBackoff)
	}
	if client.MaxBackoff != 20*time.Second {
		t.Fatalf("expected MaxBackoff 20s, got %v", client.MaxBackoff)
	}
}

func TestNewPinterestClientWithConfig_Defaults(t *testing.T) {
	cfg := PinterestClientConfig{
		RPS:         0,
		Burst:       0,
		MaxRetries:  0,
		BaseBackoff: 0,
		MaxBackoff:  0,
	}

	client := NewPinterestClientWithConfig(cfg)
	if client.MaxRetries != defaultPinterestMaxRetries {
		t.Fatalf("expected MaxRetries %d, got %d", defaultPinterestMaxRetries, client.MaxRetries)
	}
	if client.BaseBackoff != defaultPinterestBaseBackoff {
		t.Fatalf("expected BaseBackoff %v, got %v", defaultPinterestBaseBackoff, client.BaseBackoff)
	}
	if client.MaxBackoff != defaultPinterestMaxBackoff {
		t.Fatalf("expected MaxBackoff %v, got %v", defaultPinterestMaxBackoff, client.MaxBackoff)
	}
}

func TestPinterestClient_GetUserAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Error("expected Authorization header with Bearer token")
		}
		if r.URL.Path != "/user_account" {
			t.Errorf("expected path /user_account, got %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":             "user_123",
			"username":       "testuser",
			"about":          "Test bio",
			"profile_image":  "https://example.com/image.jpg",
			"website_url":    "https://example.com",
			"business_name":  "Test Business",
			"board_count":    10,
			"pin_count":      100,
			"account_type":   "BUSINESS",
			"follower_count": 5000,
			"monthly_views":  10000,
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	account, err := client.GetUserAccount(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account == nil {
		t.Fatal("expected non-nil account")
	}
	if account.ID != "user_123" {
		t.Errorf("expected ID 'user_123', got '%s'", account.ID)
	}
	if account.Username != "testuser" {
		t.Errorf("expected Username 'testuser', got '%s'", account.Username)
	}
	if account.FollowerCount != 5000 {
		t.Errorf("expected FollowerCount 5000, got %d", account.FollowerCount)
	}
}

func TestPinterestClient_GetUserAccount_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    401,
			"message": "Invalid access token",
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	_, err := client.GetUserAccount(context.Background(), "invalid-token")
	if err == nil {
		t.Fatal("expected error for unauthorized request")
	}
}

func TestPinterestClient_GetUserAccountAnalytics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if r.URL.Path != "/user_account/analytics" {
			t.Errorf("expected path /user_account/analytics, got %s", r.URL.Path)
		}

		startDate := r.URL.Query().Get("start_date")
		endDate := r.URL.Query().Get("end_date")
		if startDate == "" || endDate == "" {
			t.Error("expected start_date and end_date parameters")
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"all": map[string]interface{}{
				"daily_metrics": []map[string]interface{}{
					{
						"date":        "2024-01-01",
						"data_status": "READY",
						"metrics": map[string]interface{}{
							"IMPRESSION":     1000,
							"PIN_CLICK":      100,
							"OUTBOUND_CLICK": 50,
							"SAVE":           25,
						},
					},
					{
						"date":        "2024-01-02",
						"data_status": "READY",
						"metrics": map[string]interface{}{
							"IMPRESSION":     1200,
							"PIN_CLICK":      120,
							"OUTBOUND_CLICK": 60,
							"SAVE":           30,
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	analytics, err := client.GetUserAccountAnalytics(context.Background(), "test-token", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if analytics == nil {
		t.Fatal("expected non-nil analytics")
	}
	if len(analytics.All.DailyMetrics) != 2 {
		t.Errorf("expected 2 daily metrics, got %d", len(analytics.All.DailyMetrics))
	}
}

func TestPinterestClient_GetUserAccountAnalytics_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    400,
			"message": "Invalid date range",
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	_, err := client.GetUserAccountAnalytics(context.Background(), "test-token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestPinterestClient_GetBoards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if r.URL.Path != "/boards" {
			t.Errorf("expected path /boards, got %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"id":             "board_1",
					"name":           "Test Board 1",
					"description":    "Test description 1",
					"privacy":        "PUBLIC",
					"pin_count":      50,
					"follower_count": 100,
				},
				{
					"id":             "board_2",
					"name":           "Test Board 2",
					"description":    "Test description 2",
					"privacy":        "SECRET",
					"pin_count":      25,
					"follower_count": 50,
				},
			},
			"bookmark": "next_page_cursor",
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	boards, err := client.GetBoards(context.Background(), "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if boards == nil {
		t.Fatal("expected non-nil boards response")
	}
	if len(boards.Items) != 2 {
		t.Errorf("expected 2 boards, got %d", len(boards.Items))
	}
	if boards.Bookmark != "next_page_cursor" {
		t.Errorf("expected bookmark 'next_page_cursor', got '%s'", boards.Bookmark)
	}
	if boards.Items[0].ID != "board_1" {
		t.Errorf("expected first board ID 'board_1', got '%s'", boards.Items[0].ID)
	}
}

func TestPinterestClient_GetBoards_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	_, err := client.GetBoards(context.Background(), "test-token")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestPinterestClient_GetBoard(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if r.URL.Path != "/boards/board_123" {
			t.Errorf("expected path /boards/board_123, got %s", r.URL.Path)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":             "board_123",
			"name":           "Test Board",
			"description":    "Test description",
			"privacy":        "PUBLIC",
			"pin_count":      50,
			"follower_count": 100,
			"created_at":     "2024-01-01T00:00:00Z",
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	board, err := client.GetBoard(context.Background(), "test-token", "board_123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if board == nil {
		t.Fatal("expected non-nil board")
	}
	if board.ID != "board_123" {
		t.Errorf("expected board ID 'board_123', got '%s'", board.ID)
	}
	if board.Name != "Test Board" {
		t.Errorf("expected board name 'Test Board', got '%s'", board.Name)
	}
}

func TestPinterestClient_GetBoard_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    404,
			"message": "Board not found",
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	_, err := client.GetBoard(context.Background(), "test-token", "nonexistent_board")
	if err == nil {
		t.Fatal("expected error for not found board")
	}
}

func TestPinterestClient_GetBoardPins(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if r.URL.Path != "/boards/board_123/pins" {
			t.Errorf("expected path /boards/board_123/pins, got %s", r.URL.Path)
		}

		pageSize := r.URL.Query().Get("page_size")
		if pageSize != "25" {
			t.Errorf("expected page_size 25, got %s", pageSize)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"id":          "pin_1",
					"title":       "Pin 1",
					"description": "Test pin 1",
					"link":        "https://example.com/1",
					"board_id":    "board_123",
					"created_at":  "2024-01-01T00:00:00Z",
				},
				{
					"id":          "pin_2",
					"title":       "Pin 2",
					"description": "Test pin 2",
					"link":        "https://example.com/2",
					"board_id":    "board_123",
					"created_at":  "2024-01-02T00:00:00Z",
				},
			},
			"bookmark": "next_page_cursor",
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	pins, err := client.GetBoardPins(context.Background(), "test-token", "board_123", 25, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pins == nil {
		t.Fatal("expected non-nil pins response")
	}
	if len(pins.Items) != 2 {
		t.Errorf("expected 2 pins, got %d", len(pins.Items))
	}
	if pins.Bookmark != "next_page_cursor" {
		t.Errorf("expected bookmark 'next_page_cursor', got '%s'", pins.Bookmark)
	}
}

func TestPinterestClient_GetBoardPins_WithBookmark(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bookmark := r.URL.Query().Get("bookmark")
		if bookmark != "page_2_cursor" {
			t.Errorf("expected bookmark 'page_2_cursor', got '%s'", bookmark)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"items":    []map[string]interface{}{},
			"bookmark": "",
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	pins, err := client.GetBoardPins(context.Background(), "test-token", "board_123", 25, "page_2_cursor")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pins.Bookmark != "" {
		t.Errorf("expected empty bookmark, got '%s'", pins.Bookmark)
	}
}

func TestPinterestClient_GetUserPins(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if r.URL.Path != "/pins" {
			t.Errorf("expected path /pins, got %s", r.URL.Path)
		}

		pageSize := r.URL.Query().Get("page_size")
		if pageSize != "50" {
			t.Errorf("expected page_size 50, got %s", pageSize)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"id":          "pin_1",
					"title":       "User Pin 1",
					"description": "Test user pin 1",
					"created_at":  "2024-01-01T00:00:00Z",
				},
			},
			"bookmark": "user_pins_cursor",
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	pins, err := client.GetUserPins(context.Background(), "test-token", 50, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if pins == nil {
		t.Fatal("expected non-nil pins response")
	}
	if len(pins.Items) != 1 {
		t.Errorf("expected 1 pin, got %d", len(pins.Items))
	}
}

func TestPinterestClient_GetUserPins_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	_, err := client.GetUserPins(context.Background(), "test-token", 50, "")
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestPinterestClient_GetPinAnalytics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if r.URL.Path != "/pins/pin_123/analytics" {
			t.Errorf("expected path /pins/pin_123/analytics, got %s", r.URL.Path)
		}

		metricTypes := r.URL.Query().Get("metric_types")
		if metricTypes != "ALL" {
			t.Errorf("expected metric_types 'ALL', got '%s'", metricTypes)
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"all": map[string]interface{}{
				"daily_metrics": []map[string]interface{}{
					{
						"date":        "2024-01-01",
						"data_status": "READY",
						"metrics": map[string]interface{}{
							"IMPRESSION":     500,
							"PIN_CLICK":      50,
							"OUTBOUND_CLICK": 25,
							"SAVE":           10,
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	analytics, err := client.GetPinAnalytics(context.Background(), "test-token", "pin_123", startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if analytics == nil {
		t.Fatal("expected non-nil analytics")
	}
	if len(analytics.All.DailyMetrics) != 1 {
		t.Errorf("expected 1 daily metric, got %d", len(analytics.All.DailyMetrics))
	}
}

func TestPinterestClient_GetPinAnalytics_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	_, err := client.GetPinAnalytics(context.Background(), "test-token", "nonexistent_pin", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for API error response")
	}
}

func TestPinterestClient_GetMultiPinAnalytics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			t.Error("expected Authorization header")
		}
		if r.URL.Path != "/pins/analytics" {
			t.Errorf("expected path /pins/analytics, got %s", r.URL.Path)
		}

		pinIDs := r.URL.Query().Get("pin_ids")
		if pinIDs == "" {
			t.Error("expected pin_ids parameter")
		}

		json.NewEncoder(w).Encode(map[string]interface{}{
			"pin_1": map[string]interface{}{
				"all": map[string]interface{}{
					"daily_metrics": []map[string]interface{}{
						{
							"date":        "2024-01-01",
							"data_status": "READY",
							"metrics": map[string]interface{}{
								"IMPRESSION": 100,
								"SAVE":       10,
							},
						},
					},
				},
			},
			"pin_2": map[string]interface{}{
				"all": map[string]interface{}{
					"daily_metrics": []map[string]interface{}{
						{
							"date":        "2024-01-01",
							"data_status": "READY",
							"metrics": map[string]interface{}{
								"IMPRESSION": 200,
								"SAVE":       20,
							},
						},
					},
				},
			},
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	analytics, err := client.GetMultiPinAnalytics(context.Background(), "test-token", []string{"pin_1", "pin_2"}, startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if analytics == nil {
		t.Fatal("expected non-nil analytics")
	}
	if len(analytics) != 2 {
		t.Errorf("expected 2 pin analytics, got %d", len(analytics))
	}
	if _, ok := analytics["pin_1"]; !ok {
		t.Error("expected analytics for pin_1")
	}
	if _, ok := analytics["pin_2"]; !ok {
		t.Error("expected analytics for pin_2")
	}
}

func TestPinterestClient_GetMultiPinAnalytics_EmptyPinIDs(t *testing.T) {
	client := NewPinterestClient()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	analytics, err := client.GetMultiPinAnalytics(context.Background(), "test-token", []string{}, startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if analytics == nil {
		t.Fatal("expected non-nil analytics map")
	}
	if len(analytics) != 0 {
		t.Errorf("expected empty analytics map, got %d items", len(analytics))
	}
}

func TestPinterestClient_GetMultiPinAnalytics_ListResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(map[string]interface{}{
			"items": []map[string]interface{}{
				{
					"pin_id": "pin_1",
					"date":   "2024-01-01",
					"metrics": map[string]interface{}{
						"IMPRESSION": 100,
					},
				},
				{
					"pin_id": "pin_2",
					"date":   "2024-01-01",
					"metrics": map[string]interface{}{
						"IMPRESSION": 200,
					},
				},
			},
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	analytics, err := client.GetMultiPinAnalytics(context.Background(), "test-token", []string{"pin_1", "pin_2"}, startDate, endDate)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(analytics) != 2 {
		t.Errorf("expected 2 pin analytics, got %d", len(analytics))
	}
}

func TestPinterestClient_ContextCancelled_GetUserAccount(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetUserAccount(ctx, "test-token")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestPinterestClient_ContextCancelled_GetBoards(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetBoards(ctx, "test-token")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestPinterestClient_ContextCancelled_GetBoardPins(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetBoardPins(ctx, "test-token", "board_123", 25, "")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestPinterestClient_ContextCancelled_GetUserPins(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.GetUserPins(ctx, "test-token", 25, "")
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestPinterestClient_ContextCancelled_GetUserAccountAnalytics(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		select {}
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	startDate := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	endDate := time.Date(2024, 1, 7, 0, 0, 0, 0, time.UTC)

	_, err := client.GetUserAccountAnalytics(ctx, "test-token", startDate, endDate)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestPinterestClient_RateLimitHandling(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			w.Header().Set("x-ratelimit-reset-seconds", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":       "user_123",
			"username": "testuser",
		})
	}))
	defer server.Close()

	client := createTestPinterestClient(server.URL + "/")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	account, err := client.GetUserAccount(ctx, "test-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if account == nil {
		t.Fatal("expected non-nil account after rate limit retry")
	}
	if callCount < 2 {
		t.Errorf("expected at least 2 calls (rate limit + retry), got %d", callCount)
	}
}

func TestPinterestConstants(t *testing.T) {
	if PinterestAPIBaseURL == "" {
		t.Fatal("expected non-empty PinterestAPIBaseURL")
	}
	if defaultPinterestRPS <= 0 {
		t.Fatal("expected positive defaultPinterestRPS")
	}
	if defaultPinterestBurst <= 0 {
		t.Fatal("expected positive defaultPinterestBurst")
	}
	if PinterestFullSyncDays <= 0 {
		t.Fatal("expected positive PinterestFullSyncDays")
	}
	if PinterestIncrementalSyncDays <= 0 {
		t.Fatal("expected positive PinterestIncrementalSyncDays")
	}
	if PinterestMultiPinBatchSize <= 0 {
		t.Fatal("expected positive PinterestMultiPinBatchSize")
	}
	if PinterestMultiPinBatchSize != 25 {
		t.Fatalf("expected PinterestMultiPinBatchSize 25, got %d", PinterestMultiPinBatchSize)
	}
}

func TestGetInt64FromMetrics(t *testing.T) {
	tests := []struct {
		name     string
		metrics  map[string]interface{}
		key      string
		expected int64
	}{
		{
			name:     "float64 value",
			metrics:  map[string]interface{}{"IMPRESSION": float64(1000)},
			key:      "IMPRESSION",
			expected: 1000,
		},
		{
			name:     "int64 value",
			metrics:  map[string]interface{}{"IMPRESSION": int64(2000)},
			key:      "IMPRESSION",
			expected: 2000,
		},
		{
			name:     "int value",
			metrics:  map[string]interface{}{"IMPRESSION": int(3000)},
			key:      "IMPRESSION",
			expected: 3000,
		},
		{
			name:     "missing key",
			metrics:  map[string]interface{}{"OTHER": float64(1000)},
			key:      "IMPRESSION",
			expected: 0,
		},
		{
			name:     "nil metrics",
			metrics:  nil,
			key:      "IMPRESSION",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetInt64FromMetrics(tt.metrics, tt.key)
			if got != tt.expected {
				t.Errorf("GetInt64FromMetrics() = %d, want %d", got, tt.expected)
			}
		})
	}
}

func TestGetFloat64FromMetrics(t *testing.T) {
	tests := []struct {
		name     string
		metrics  map[string]interface{}
		key      string
		expected float64
	}{
		{
			name:     "float64 value",
			metrics:  map[string]interface{}{"RATE": float64(0.5)},
			key:      "RATE",
			expected: 0.5,
		},
		{
			name:     "int64 value",
			metrics:  map[string]interface{}{"RATE": int64(2)},
			key:      "RATE",
			expected: 2.0,
		},
		{
			name:     "int value",
			metrics:  map[string]interface{}{"RATE": int(3)},
			key:      "RATE",
			expected: 3.0,
		},
		{
			name:     "missing key",
			metrics:  map[string]interface{}{"OTHER": float64(1.0)},
			key:      "RATE",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetFloat64FromMetrics(tt.metrics, tt.key)
			if got != tt.expected {
				t.Errorf("GetFloat64FromMetrics() = %f, want %f", got, tt.expected)
			}
		})
	}
}

func TestGetStringFromMap(t *testing.T) {
	tests := []struct {
		name     string
		m        map[string]interface{}
		key      string
		expected string
	}{
		{
			name:     "string value",
			m:        map[string]interface{}{"url": "https://example.com"},
			key:      "url",
			expected: "https://example.com",
		},
		{
			name:     "missing key",
			m:        map[string]interface{}{"other": "value"},
			key:      "url",
			expected: "",
		},
		{
			name:     "non-string value",
			m:        map[string]interface{}{"url": 123},
			key:      "url",
			expected: "",
		},
		{
			name:     "nil map",
			m:        nil,
			key:      "url",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetStringFromMap(tt.m, tt.key)
			if got != tt.expected {
				t.Errorf("GetStringFromMap() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetMediaField(t *testing.T) {
	tests := []struct {
		name     string
		media    map[string]interface{}
		field    string
		expected string
	}{
		{
			name:     "string field",
			media:    map[string]interface{}{"media_type": "image"},
			field:    "media_type",
			expected: "image",
		},
		{
			name:     "missing field",
			media:    map[string]interface{}{"other": "value"},
			field:    "media_type",
			expected: "",
		},
		{
			name:     "nil media",
			media:    nil,
			field:    "media_type",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetMediaField(tt.media, tt.field)
			if got != tt.expected {
				t.Errorf("GetMediaField() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestGetPinCoverImageURL(t *testing.T) {
	tests := []struct {
		name     string
		pin      PinterestPin
		expected string
	}{
		{
			name:     "nil media",
			pin:      PinterestPin{ID: "pin_1", Media: nil},
			expected: "",
		},
		{
			name: "video with cover image",
			pin: PinterestPin{
				ID: "pin_1",
				Media: map[string]interface{}{
					"media_type":      "video",
					"cover_image_url": "https://example.com/cover.jpg",
				},
			},
			expected: "https://example.com/cover.jpg",
		},
		{
			name: "image with 150x150",
			pin: PinterestPin{
				ID: "pin_1",
				Media: map[string]interface{}{
					"media_type": "image",
					"images": map[string]interface{}{
						"150x150": map[string]interface{}{
							"url": "https://example.com/150x150.jpg",
						},
					},
				},
			},
			expected: "https://example.com/150x150.jpg",
		},
		{
			name: "multiple images with first item",
			pin: PinterestPin{
				ID: "pin_1",
				Media: map[string]interface{}{
					"media_type": "multiple_images",
					"items": []interface{}{
						map[string]interface{}{
							"images": map[string]interface{}{
								"150x150": map[string]interface{}{
									"url": "https://example.com/first.jpg",
								},
							},
						},
					},
				},
			},
			expected: "https://example.com/first.jpg",
		},
		{
			name: "multiple images empty items",
			pin: PinterestPin{
				ID: "pin_1",
				Media: map[string]interface{}{
					"media_type": "multiple_images",
					"items":      []interface{}{},
				},
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GetPinCoverImageURL(tt.pin)
			if got != tt.expected {
				t.Errorf("GetPinCoverImageURL() = %q, want %q", got, tt.expected)
			}
		})
	}
}

func TestPinterestAPIInterface(t *testing.T) {
	var _ PinterestAPI = (*PinterestClient)(nil)
}

func createTestPinterestClient(baseURL string) *PinterestClient {
	client := NewPinterestClientWithConfig(PinterestClientConfig{
		RPS:         100,
		Burst:       100,
		MaxRetries:  3,
		BaseBackoff: 10 * time.Millisecond,
		MaxBackoff:  100 * time.Millisecond,
	})

	originalMakeRequest := client.makeRequest
	_ = originalMakeRequest

	transport := &testPinterestTransport{
		baseURL:      baseURL,
		roundTripper: http.DefaultTransport,
	}
	client.HTTPClient = &http.Client{
		Transport: transport,
		Timeout:   30 * time.Second,
	}

	return client
}

type testPinterestTransport struct {
	baseURL      string
	roundTripper http.RoundTripper
}

func (t *testPinterestTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	baseURL := strings.TrimSuffix(t.baseURL, "/")
	path := req.URL.Path
	if strings.HasPrefix(path, "/v5/") {
		path = strings.TrimPrefix(path, "/v5")
	}
	newURL := baseURL + path
	if req.URL.RawQuery != "" {
		newURL += "?" + req.URL.RawQuery
	}

	newReq, err := http.NewRequestWithContext(req.Context(), req.Method, newURL, req.Body)
	if err != nil {
		return nil, err
	}
	newReq.Header = req.Header

	return t.roundTripper.RoundTrip(newReq)
}
