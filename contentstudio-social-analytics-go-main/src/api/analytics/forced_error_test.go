package analytics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"slices"
	"testing"

	"github.com/rs/zerolog"
)

func TestForcedErrorConfigFromEnv(t *testing.T) {
	t.Setenv("ANALYTICS_FORCE_ERROR_STATUSES", "400,401,500")
	t.Setenv("ANALYTICS_FORCE_ERROR_MESSAGE", "forced bad request")
	t.Setenv("ANALYTICS_FORCE_ERROR_PLATFORMS", "facebook,linkedin")

	cfg := ForcedErrorConfigFromEnv(zerolog.New(io.Discard))
	if cfg == nil {
		t.Fatal("expected forced error config")
	}
	expectedStatuses := []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError}
	if len(cfg.StatusCodes) != len(expectedStatuses) {
		t.Fatalf("expected %d statuses, got %d", len(expectedStatuses), len(cfg.StatusCodes))
	}
	for i, statusCode := range expectedStatuses {
		if cfg.StatusCodes[i] != statusCode {
			t.Fatalf("expected status %d at index %d, got %d", statusCode, i, cfg.StatusCodes[i])
		}
	}
	if cfg.Message != "forced bad request" {
		t.Fatalf("expected custom message, got %q", cfg.Message)
	}
	if len(cfg.Platforms) != 2 || cfg.Platforms[0] != "facebook" || cfg.Platforms[1] != "linkedin" {
		t.Fatalf("unexpected platforms: %#v", cfg.Platforms)
	}
}

func TestForcedErrorConfigFromEnvRejectsInvalidStatus(t *testing.T) {
	t.Setenv("ANALYTICS_FORCE_ERROR_STATUS", "200")

	cfg := ForcedErrorConfigFromEnv(zerolog.New(io.Discard))
	if cfg != nil {
		t.Fatal("expected nil config for invalid status")
	}
}

func TestWithForcedErrorInterceptsAnalyticsRoutes(t *testing.T) {
	handler := WithForcedError(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}), &ForcedErrorConfig{
		StatusCodes: []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError},
		Message:     "forced analytics error",
		Platforms:   []string{"facebook", "linkedin"},
	})

	for i := 0; i < 25; i++ {
		req := httptest.NewRequest(http.MethodGet, "/analytics/overview/facebook/summary", nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if !slices.Contains([]int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusInternalServerError}, rec.Code) {
			t.Fatalf("unexpected status %d", rec.Code)
		}

		expected := "{\"status\":false,\"message\":\"forced analytics error\"}\n"
		if rec.Body.String() != expected {
			t.Fatalf("unexpected body: %q", rec.Body.String())
		}
	}
}

func TestWithForcedErrorPassesThroughNonAnalyticsRoutes(t *testing.T) {
	handler := WithForcedError(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), &ForcedErrorConfig{
		StatusCodes: []int{http.StatusBadRequest},
		Message:     "forced analytics error",
		Platforms:   []string{"facebook", "linkedin"},
	})

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected passthrough status 202, got %d", rec.Code)
	}
}

func TestWithForcedErrorPassesThroughOtherPlatforms(t *testing.T) {
	handler := WithForcedError(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), &ForcedErrorConfig{
		StatusCodes: []int{http.StatusBadRequest},
		Message:     "forced analytics error",
		Platforms:   []string{"facebook", "linkedin"},
	})

	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/gmb/summary", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected passthrough status 202, got %d", rec.Code)
	}
}

func TestWithForcedErrorInterceptsOnlyConfiguredPaths(t *testing.T) {
	handler := WithForcedError(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}), &ForcedErrorConfig{
		StatusCodes: []int{http.StatusUnauthorized},
		Message:     "forced analytics error",
		Paths: []string{
			"/analytics/overview/facebook/summary",
			"/analytics/overview/linkedin/ai_insights",
		},
	})

	req := httptest.NewRequest(http.MethodGet, "/analytics/overview/facebook/summary", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected configured path to be intercepted, got %d", rec.Code)
	}

	req = httptest.NewRequest(http.MethodGet, "/analytics/overview/facebook/overviewImpressions", nil)
	rec = httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusAccepted {
		t.Fatalf("expected non-configured path to pass through, got %d", rec.Code)
	}
}
