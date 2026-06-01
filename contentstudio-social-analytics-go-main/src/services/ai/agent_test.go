package ai

import (
	"context"
	"encoding/json"
	"io"
	"math"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

func TestNewAgentClient(t *testing.T) {
	cfg := &config.AIAgentsConfig{
		BaseURL: "http://localhost:8000",
		APIKey:  "test-key",
		Timeout: 60,
	}
	client := NewAgentClient(cfg, zerolog.New(io.Discard))

	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.baseURL != "http://localhost:8000/" {
		t.Fatalf("expected trailing slash, got %q", client.baseURL)
	}
	if client.apiKey != "test-key" {
		t.Fatalf("expected test-key, got %q", client.apiKey)
	}
}

func TestNewAgentClient_DefaultTimeout(t *testing.T) {
	cfg := &config.AIAgentsConfig{
		BaseURL: "http://localhost:8000/",
		Timeout: 0,
	}
	client := NewAgentClient(cfg, zerolog.New(io.Discard))

	if client.httpClient.Timeout.Seconds() != 300 {
		t.Fatalf("expected 300s timeout, got %v", client.httpClient.Timeout)
	}
}

func TestNewAgentClient_TrailingSlashNormalization(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://localhost:8000", "http://localhost:8000/"},
		{"http://localhost:8000/", "http://localhost:8000/"},
		{"http://localhost:8000///", "http://localhost:8000/"},
	}

	for _, tc := range tests {
		cfg := &config.AIAgentsConfig{BaseURL: tc.input}
		client := NewAgentClient(cfg, zerolog.New(io.Discard))
		if client.baseURL != tc.expected {
			t.Fatalf("input %q: expected %q, got %q", tc.input, tc.expected, client.baseURL)
		}
	}
}

func TestRequest_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Fatalf("expected application/json content-type")
		}
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Fatalf("expected Bearer test-key, got %q", r.Header.Get("Authorization"))
		}
		if r.URL.Path != "/api/v1/gmb/impressions-overview" {
			t.Fatalf("expected /api/v1/gmb/impressions-overview, got %q", r.URL.Path)
		}

		var body map[string]interface{}
		json.NewDecoder(r.Body).Decode(&body)
		if body["language"] != "en" {
			t.Fatalf("expected language=en, got %v", body["language"])
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"insights":        map[string]interface{}{"key": "value"},
			"processing_time": 1.5,
		})
	}))
	defer server.Close()

	cfg := &config.AIAgentsConfig{BaseURL: server.URL, APIKey: "test-key", Timeout: 10}
	client := NewAgentClient(cfg, zerolog.New(io.Discard))

	result, err := client.Request(context.Background(), "gmb/impressions-overview", map[string]interface{}{
		"language": "en",
		"dataset":  map[string]interface{}{"buckets": []string{"2025-01-01"}},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["processing_time"] == nil {
		t.Fatal("expected processing_time in response")
	}
}

func TestRequest_NoAPIKey(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "" {
			t.Fatalf("expected no Authorization header, got %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	cfg := &config.AIAgentsConfig{BaseURL: server.URL, Timeout: 10}
	client := NewAgentClient(cfg, zerolog.New(io.Discard))

	_, err := client.Request(context.Background(), "gmb/test", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequest_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error":"internal"}`))
	}))
	defer server.Close()

	cfg := &config.AIAgentsConfig{BaseURL: server.URL, Timeout: 10}
	client := NewAgentClient(cfg, zerolog.New(io.Discard))

	_, err := client.Request(context.Background(), "gmb/test", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for 500 response")
	}
}

func TestRequest_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`not json`))
	}))
	defer server.Close()

	cfg := &config.AIAgentsConfig{BaseURL: server.URL, Timeout: 10}
	client := NewAgentClient(cfg, zerolog.New(io.Discard))

	_, err := client.Request(context.Background(), "gmb/test", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for invalid JSON response")
	}
}

func TestRequest_ContextCancelled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	cfg := &config.AIAgentsConfig{BaseURL: server.URL, Timeout: 10}
	client := NewAgentClient(cfg, zerolog.New(io.Discard))

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	_, err := client.Request(ctx, "gmb/test", map[string]interface{}{})
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}

func TestRequest_EndpointTrimming(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/gmb/test" {
			t.Fatalf("expected /api/v1/gmb/test, got %q", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	cfg := &config.AIAgentsConfig{BaseURL: server.URL, Timeout: 10}
	client := NewAgentClient(cfg, zerolog.New(io.Discard))

	_, err := client.Request(context.Background(), "/gmb/test", map[string]interface{}{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRequest_SanitizesNonFiniteFloatsInPayload(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}

		dataset, ok := body["dataset"].(map[string]interface{})
		if !ok {
			t.Fatalf("expected dataset object, got %#v", body["dataset"])
		}
		if got := dataset["rate"].(float64); got != 0 {
			t.Fatalf("expected rate to be sanitized to 0, got %v", got)
		}

		series, ok := dataset["series"].([]interface{})
		if !ok {
			t.Fatalf("expected series array, got %#v", dataset["series"])
		}
		if got := series[0].(float64); got != 0 {
			t.Fatalf("expected series[0] to be sanitized to 0, got %v", got)
		}
		if got := series[1].(float64); got != 12.5 {
			t.Fatalf("expected series[1] to remain 12.5, got %v", got)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"ok": true})
	}))
	defer server.Close()

	cfg := &config.AIAgentsConfig{BaseURL: server.URL, Timeout: 10}
	client := NewAgentClient(cfg, zerolog.New(io.Discard))

	_, err := client.Request(context.Background(), "gmb/test", map[string]interface{}{
		"language": "en",
		"dataset": map[string]interface{}{
			"rate":   math.NaN(),
			"series": []float64{math.Inf(1), 12.5},
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
