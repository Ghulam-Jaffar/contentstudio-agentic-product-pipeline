package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

type stubApiKeyValidator struct {
	validKeys map[string]bool
}

type stubShareableLinkValidator struct {
	validLinks map[string]string
}

func (s *stubApiKeyValidator) FindValidByKey(_ context.Context, key string) (*mongoModels.ApiKey, error) {
	if s.validKeys[key] {
		return &mongoModels.ApiKey{Key: key}, nil
	}
	return nil, nil
}

func (s *stubShareableLinkValidator) FindActiveUserIDByLinkID(_ context.Context, linkID string) (string, error) {
	return s.validLinks[linkID], nil
}

func newStubValidator(keys ...string) ApiKeyValidator {
	m := make(map[string]bool, len(keys))
	for _, k := range keys {
		m[k] = true
	}
	return &stubApiKeyValidator{validKeys: m}
}

func newStubShareableValidator(validLinks map[string]string) ShareableLinkValidator {
	if validLinks == nil {
		validLinks = map[string]string{}
	}
	return &stubShareableLinkValidator{validLinks: validLinks}
}

func okHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}

func newTestAuthMiddleware(jwtSecret, apiKey string, validShareableLinks map[string]string) *AuthMiddleware {
	log, _ := logger.NewTestLogger()
	var jwtMw *JWTMiddleware
	if jwtSecret != "" {
		jwtMw = NewJWTMiddleware(&config.JWTConfig{Secret: jwtSecret}, log)
	}
	var validator ApiKeyValidator
	if apiKey != "" {
		validator = newStubValidator(apiKey)
	}
	shareableValidator := newStubShareableValidator(validShareableLinks)
	return NewAuthMiddleware(jwtMw, validator, shareableValidator, log)
}

func TestAuthMiddleware_HealthExempt(t *testing.T) {
	m := newTestAuthMiddleware("", "my-key", nil)
	handler := m.Authenticate(okHandler())

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for /health, got %d", w.Code)
	}
}

func TestAuthMiddleware_APIKeyOnly(t *testing.T) {
	m := newTestAuthMiddleware("", "my-secret-key", nil)
	handler := m.Authenticate(okHandler())

	tests := []struct {
		name           string
		apiKey         string
		expectedStatus int
	}{
		{"valid key", "my-secret-key", http.StatusOK},
		{"invalid key", "wrong-key", http.StatusUnauthorized},
		{"missing key", "", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tc.apiKey != "" {
				req.Header.Set("X-API-Key", tc.apiKey)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthMiddleware_JWTOnly(t *testing.T) {
	m := newTestAuthMiddleware(testSecret, "", nil)
	handler := m.Authenticate(okHandler())

	validToken := generateToken(testSecret, "user1", "", time.Hour)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
	}{
		{"valid JWT", "Bearer " + validToken, http.StatusOK},
		{"invalid JWT", "Bearer bad.token.here", http.StatusUnauthorized},
		{"missing auth", "", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthMiddleware_BothConfigured(t *testing.T) {
	m := newTestAuthMiddleware(testSecret, "my-api-key", nil)
	handler := m.Authenticate(okHandler())

	validToken := generateToken(testSecret, "user1", "", time.Hour)

	tests := []struct {
		name           string
		authHeader     string
		apiKey         string
		expectedStatus int
	}{
		{"JWT takes priority", "Bearer " + validToken, "", http.StatusOK},
		{"API key works without Bearer", "", "my-api-key", http.StatusOK},
		{"invalid JWT fails even if API key present", "Bearer bad.token", "my-api-key", http.StatusUnauthorized},
		{"no credentials", "", "", http.StatusUnauthorized},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			if tc.apiKey != "" {
				req.Header.Set("X-API-Key", tc.apiKey)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestAuthMiddleware_ShareableOnly(t *testing.T) {
	m := newTestAuthMiddleware("", "", map[string]string{"share-id-1": "user123"})
	handler := m.Authenticate(okHandler())

	req := httptest.NewRequest("GET", "/analytics/overview/facebook/summary", nil)
	req.Header.Set("X-Shareable-ID", "share-id-1")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 for valid shareable link, got %d", w.Code)
	}
}

func TestAuthMiddleware_ShareableOverridesInvalidJWT(t *testing.T) {
	m := newTestAuthMiddleware(testSecret, "", map[string]string{"share-id-2": "user456"})
	handler := m.Authenticate(okHandler())

	req := httptest.NewRequest("GET", "/analytics/overview/instagram/summary", nil)
	req.Header.Set("Authorization", "Bearer invalid.token")
	req.Header.Set("X-Shareable-ID", "share-id-2")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when valid shareable link exists (even with invalid JWT), got %d", w.Code)
	}
}

func TestAuthMiddleware_InvalidShareableFallsBackToJWT(t *testing.T) {
	m := newTestAuthMiddleware(testSecret, "", map[string]string{"share-id-3": "user789"})
	handler := m.Authenticate(okHandler())
	validToken := generateToken(testSecret, "user1", "", time.Hour)

	req := httptest.NewRequest("GET", "/analytics/overview/linkedin/summary", nil)
	req.Header.Set("Authorization", "Bearer "+validToken)
	req.Header.Set("X-Shareable-ID", "invalid-share-id")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200 when JWT is valid and shareable link is invalid, got %d", w.Code)
	}
}

func TestAuthMiddleware_ShareableIgnoredOnNonAnalyticsPath(t *testing.T) {
	m := newTestAuthMiddleware("", "", map[string]string{"share-id-4": "user999"})
	handler := m.Authenticate(okHandler())

	req := httptest.NewRequest("GET", "/api/v1/immediate-work", nil)
	req.Header.Set("X-Shareable-ID", "share-id-4")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 when shareable header is used on non-analytics path, got %d", w.Code)
	}
}
