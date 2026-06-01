package gmb

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/gmb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/ai"
)

// --- Mock cache ---

type mockCache struct {
	store map[string]string
}

func newMockCache() *mockCache {
	return &mockCache{store: make(map[string]string)}
}

func (m *mockCache) Get(_ context.Context, key string) (string, error) {
	return m.store[key], nil
}

func (m *mockCache) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	m.store[key] = value.(string)
	return nil
}

func (m *mockCache) SetNX(_ context.Context, key string, value interface{}, _ time.Duration) (bool, error) {
	if _, ok := m.store[key]; ok {
		return false, nil
	}
	m.store[key] = value.(string)
	return true, nil
}

func (m *mockCache) Del(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(m.store, k)
	}
	return nil
}

func (m *mockCache) CompareAndDelete(_ context.Context, key, expected string) (bool, error) {
	if val, ok := m.store[key]; ok && val == expected {
		delete(m.store, key)
		return true, nil
	}
	return false, nil
}

func (m *mockCache) DecrBy(_ context.Context, _ string, _ int64) (int64, error) { return 0, nil }
func (m *mockCache) DecrByIfPositive(_ context.Context, _ string, _ int64) (int64, bool, error) {
	return 0, false, nil
}
func (m *mockCache) IncrBy(_ context.Context, _ string, _ int64) (int64, error) { return 0, nil }
func (m *mockCache) Expire(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}
func (m *mockCache) Close() error { return nil }

// --- Helpers ---

func newTestAIService(serverURL string) *AIInsightsService {
	svc := newTestService()
	cfg := &config.AIAgentsConfig{BaseURL: serverURL, Timeout: 10}
	agentClient := ai.NewAgentClient(cfg, zerolog.New(io.Discard))
	return NewAIInsightsService(svc, agentClient, newMockCache())
}

func newTestAIServiceNoCache(serverURL string) *AIInsightsService {
	svc := newTestService()
	cfg := &config.AIAgentsConfig{BaseURL: serverURL, Timeout: 10}
	agentClient := ai.NewAgentClient(cfg, zerolog.New(io.Discard))
	return NewAIInsightsService(svc, agentClient, nil)
}

func aiAgentServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"insights":        map[string]interface{}{"key": "value"},
			"processing_time": 1.23,
			"model_used":      "gemini-2.5-flash",
		})
	}))
}

// --- Tests ---

func TestNewAIInsightsService(t *testing.T) {
	server := aiAgentServer(t)
	defer server.Close()

	svc := newTestAIService(server.URL)
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestGetAIInsights_InvalidType(t *testing.T) {
	server := aiAgentServer(t)
	defer server.Close()

	svc := newTestAIService(server.URL)

	_, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		Type:        "invalid_type",
		Date:        []interface{}{"2025-01-01", "2025-01-31"},
	})
	if err == nil {
		t.Fatal("expected error for invalid insight type")
	}
}

func TestGetAIInsights_ValidTypes(t *testing.T) {
	server := aiAgentServer(t)
	defer server.Close()

	validTypes := []string{
		"impressions_overview",
		"actions_overview",
		"search_keywords",
		"publishing_behavior",
		"insights_summary",
	}

	for _, insightType := range validTypes {
		t.Run(insightType, func(t *testing.T) {
			svc := newTestAIService(server.URL)

			result, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
				WorkspaceID: "ws1",
				GmbID:       "loc_123",
				Type:        insightType,
				Date:        []interface{}{"2025-01-01", "2025-01-31"},
				Language:    "en",
			})
			if err != nil {
				t.Fatalf("unexpected error for type %s: %v", insightType, err)
			}
			// Mock ClickHouse returns empty data, so service returns insufficient data
			// This is correct behavior — no data = no AI call
			if result == nil {
				t.Fatalf("expected non-nil result for type %s", insightType)
			}
		})
	}
}

func TestGetAIInsights_CacheHit(t *testing.T) {
	server := aiAgentServer(t)
	defer server.Close()

	svc := newTestAIService(server.URL)

	req := &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		Type:        "impressions_overview",
		Date:        []interface{}{"2025-01-01", "2025-01-31"},
		Language:    "en",
	}

	// Pre-populate cache with a valid response
	cacheKey := svc.buildCacheKey(req)
	cachedData, _ := json.Marshal(map[string]interface{}{
		"insights":        map[string]interface{}{"key": "cached_value"},
		"processing_time": 0.5,
	})
	svc.cache.Set(context.Background(), cacheKey, string(cachedData), 0)

	// Call should return cached result
	result, err := svc.GetAIInsights(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["success"] != true {
		t.Fatal("expected success=true from cache")
	}
	insights := result["insights"].(map[string]interface{})
	if insights["key"] != "cached_value" {
		t.Fatal("expected cached_value from cache hit")
	}
}

func TestGetAIInsights_NilCache(t *testing.T) {
	server := aiAgentServer(t)
	defer server.Close()

	svc := newTestAIServiceNoCache(server.URL)

	result, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		Type:        "impressions_overview",
		Date:        []interface{}{"2025-01-01", "2025-01-31"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With nil cache and empty mock data, returns insufficient data without panic
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if result["message"] != "insufficient data" {
		t.Fatalf("expected insufficient data message, got %v", result["message"])
	}
}

func TestGetAIInsights_DefaultLanguage(t *testing.T) {
	// Default language is applied in two places: cache key and AI agent payload
	// Test via cache key since mock ClickHouse returns empty data (agent not called)
	svc := newTestAIService("http://localhost")

	req := &types.AIInsightsRequest{
		Type:  "impressions_overview",
		GmbID: "loc_123",
		Date:  "2025-01-01 - 2025-01-31",
	}
	key := svc.buildCacheKey(req)
	if !strings.HasSuffix(key, ":en") {
		t.Fatalf("expected cache key to end with :en (default language), got %q", key)
	}

	// Also verify via cache: pre-populate and confirm language default flows through GetAIInsights
	server := aiAgentServer(t)
	defer server.Close()
	svc = newTestAIService(server.URL)

	cacheKey := svc.buildCacheKey(&types.AIInsightsRequest{
		Type:     "impressions_overview",
		GmbID:    "loc_123",
		Date:     []interface{}{"2025-01-01", "2025-01-31"},
		Language: "", // empty = should default to "en"
	})
	cachedData, _ := json.Marshal(map[string]interface{}{"from": "cache"})
	svc.cache.Set(context.Background(), cacheKey, string(cachedData), 0)

	result, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		Type:        "impressions_overview",
		Date:        []interface{}{"2025-01-01", "2025-01-31"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["success"] != true {
		t.Fatal("expected cache hit with default language key")
	}
}

func TestGetAIInsights_InsufficientData(t *testing.T) {
	// Mock ClickHouse returns empty data, service should return insufficient data
	// without calling the AI agent at all
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("AI agent should not be called when there is no data")
	}))
	defer server.Close()

	svc := newTestAIService(server.URL)

	result, err := svc.GetAIInsights(context.Background(), &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		Type:        "impressions_overview",
		Date:        []interface{}{"2025-01-01", "2025-01-31"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["success"] != false {
		t.Fatal("expected success=false for insufficient data")
	}
	if result["message"] != "insufficient data" {
		t.Fatalf("expected 'insufficient data' message, got %v", result["message"])
	}
}

func TestToGMBRequest_ParsesDashSeparatedDate(t *testing.T) {
	svc := newTestAIService("http://localhost")

	req := svc.toGMBRequest(&types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		Date:        "2025-01-01 - 2025-01-31",
		Timezone:    "UTC",
	})

	if req.StartDate != "2025-01-01" || req.EndDate != "2025-01-31" {
		t.Fatalf("expected dash-separated date to parse, got start=%q end=%q", req.StartDate, req.EndDate)
	}
}

func TestToGMBRequest_PrefersDateOverExplicitStartEndDate(t *testing.T) {
	svc := newTestAIService("http://localhost")

	req := svc.toGMBRequest(&types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		StartDate:   "2025-02-01",
		EndDate:     "2025-02-28",
		Date:        "2025-01-01 - 2025-01-31",
		Timezone:    "UTC",
	})

	if req.StartDate != "2025-01-01" || req.EndDate != "2025-01-31" {
		t.Fatalf("expected date to win, got start=%q end=%q", req.StartDate, req.EndDate)
	}
}

// --- toGMBRequest tests ---

func TestToGMBRequest_ArrayDate(t *testing.T) {
	svc := newTestAIService("http://localhost")

	req := &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		Date:        []interface{}{"2025-01-01", "2025-01-31"},
		Timezone:    "America/New_York",
	}

	gmbReq := svc.toGMBRequest(req)

	if gmbReq.StartDate != "2025-01-01" {
		t.Fatalf("expected 2025-01-01, got %q", gmbReq.StartDate)
	}
	if gmbReq.EndDate != "2025-01-31" {
		t.Fatalf("expected 2025-01-31, got %q", gmbReq.EndDate)
	}
	if gmbReq.WorkspaceID != "ws1" {
		t.Fatalf("expected ws1, got %q", gmbReq.WorkspaceID)
	}
	if gmbReq.Timezone != "America/New_York" {
		t.Fatalf("expected America/New_York, got %q", gmbReq.Timezone)
	}
}

func TestToGMBRequest_StringDate(t *testing.T) {
	svc := newTestAIService("http://localhost")

	req := &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		Date:        "2025-01-01 - 2025-01-31",
	}

	gmbReq := svc.toGMBRequest(req)

	if gmbReq.StartDate != "2025-01-01" {
		t.Fatalf("expected 2025-01-01, got %q", gmbReq.StartDate)
	}
	if gmbReq.EndDate != "2025-01-31" {
		t.Fatalf("expected 2025-01-31, got %q", gmbReq.EndDate)
	}
}

func TestToGMBRequest_UnknownDateType(t *testing.T) {
	svc := newTestAIService("http://localhost")

	req := &types.AIInsightsRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		Date:        12345,
	}

	gmbReq := svc.toGMBRequest(req)

	if gmbReq.StartDate != "" {
		t.Fatalf("expected empty start_date, got %q", gmbReq.StartDate)
	}
	if gmbReq.EndDate != "" {
		t.Fatalf("expected empty end_date, got %q", gmbReq.EndDate)
	}
}

// --- buildCacheKey tests ---

func TestBuildCacheKey(t *testing.T) {
	svc := newTestAIService("http://localhost")

	req := &types.AIInsightsRequest{
		Type:     "impressions_overview",
		GmbID:    "loc_123",
		Date:     "2025-01-01,2025-01-31",
		Language: "fr",
	}

	key := svc.buildCacheKey(req)
	expected := "gmb_AI:impressions_overview:loc_123:2025-01-01,2025-01-31:fr"
	if key != expected {
		t.Fatalf("expected %q, got %q", expected, key)
	}
}

func TestBuildCacheKey_DefaultLanguage(t *testing.T) {
	svc := newTestAIService("http://localhost")

	req := &types.AIInsightsRequest{
		Type:  "actions_overview",
		GmbID: "loc_456",
		Date:  []interface{}{"2025-02-01", "2025-02-28"},
	}

	key := svc.buildCacheKey(req)
	if key == "" {
		t.Fatal("expected non-empty cache key")
	}
	// Should contain "en" as default language
	expected := "gmb_AI:actions_overview:loc_456:[2025-02-01 2025-02-28]:en"
	if key != expected {
		t.Fatalf("expected %q, got %q", expected, key)
	}
}

// --- Helper map tests ---

func TestImpressionsToMap_Nil(t *testing.T) {
	svc := newTestAIService("http://localhost")
	result := svc.impressionsToMap(&types.ImpressionsResponse{})
	if result == nil {
		t.Fatal("expected non-nil map")
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(result))
	}
}

func TestImpressionsToMap_WithData(t *testing.T) {
	svc := newTestAIService("http://localhost")
	result := svc.impressionsToMap(&types.ImpressionsResponse{
		Impressions: &types.ImpressionsData{
			Buckets:            []string{"2025-01-01"},
			DesktopMapsDaily:   []int64{10},
			DesktopSearchDaily: []int64{20},
			MobileMapsDaily:    []int64{5},
			MobileSearchDaily:  []int64{15},
		},
	})

	if result["buckets"] == nil {
		t.Fatal("expected buckets in map")
	}
}

func TestActionsToMap_Nil(t *testing.T) {
	svc := newTestAIService("http://localhost")
	result := svc.actionsToMap(&types.ActionsResponse{})
	if result == nil {
		t.Fatal("expected non-nil map")
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %d keys", len(result))
	}
}

func TestActionsToMap_WithData(t *testing.T) {
	svc := newTestAIService("http://localhost")
	result := svc.actionsToMap(&types.ActionsResponse{
		Actions: &types.ActionsData{
			Buckets:                []string{"2025-01-01"},
			CallClicksDaily:        []int64{5},
			WebsiteClicksDaily:     []int64{10},
			DirectionRequestsDaily: []int64{3},
			OtherActionsDaily:      []int64{1},
		},
	})

	if result["buckets"] == nil {
		t.Fatal("expected buckets in map")
	}
	if result["call_clicks"] == nil {
		t.Fatal("expected call_clicks in map")
	}
}
