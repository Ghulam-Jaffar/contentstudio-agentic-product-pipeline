package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

const testSecret = "test-secret-key-for-jwt-testing"

func newTestMiddleware(secret, issuer string) *JWTMiddleware {
	log, _ := logger.NewTestLogger()
	return NewJWTMiddleware(&config.JWTConfig{
		Secret: secret,
		Issuer: issuer,
	}, log)
}

func generateToken(secret, subject, issuer string, expiry time.Duration) string {
	claims := &JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   subject,
			Issuer:    issuer,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(expiry)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(secret))
	return signed
}

func TestNewJWTMiddleware(t *testing.T) {
	m := newTestMiddleware(testSecret, "")
	if m == nil {
		t.Fatal("expected non-nil middleware")
	}
	if m.Config.Secret != testSecret {
		t.Fatalf("expected secret %q, got %q", testSecret, m.Config.Secret)
	}
}

func TestAuthenticate(t *testing.T) {
	validToken := generateToken(testSecret, "user123", "", time.Hour)
	wrongSecretToken := generateToken("wrong-secret", "user123", "", time.Hour)
	tokenWithIssuer := generateToken(testSecret, "user123", "my-issuer", time.Hour)
	tokenWithWrongIssuer := generateToken(testSecret, "user123", "wrong-issuer", time.Hour)

	tests := []struct {
		name            string
		secret          string
		issuer          string
		authHeader      string
		expectedStatus  int
		expectNextCall  bool
		expectedMessage string
	}{
		{
			name:            "missing Authorization header",
			secret:          testSecret,
			authHeader:      "",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Missing Authorization header",
		},
		{
			name:            "Basic auth instead of Bearer",
			secret:          testSecret,
			authHeader:      "Basic sometoken",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Invalid Authorization header format",
		},
		{
			name:            "no space in header",
			secret:          testSecret,
			authHeader:      "Bearertoken",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Invalid Authorization header format",
		},
		{
			name:            "malformed token",
			secret:          testSecret,
			authHeader:      "Bearer invalid.token.here",
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Invalid token",
		},
		{
			name:            "wrong signing secret",
			secret:          testSecret,
			authHeader:      "Bearer " + wrongSecretToken,
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Invalid token",
		},
		{
			name:           "valid token",
			secret:         testSecret,
			authHeader:     "Bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "lowercase bearer prefix",
			secret:         testSecret,
			authHeader:     "bearer " + validToken,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
		},
		{
			name:            "issuer mismatch",
			secret:          testSecret,
			issuer:          "expected-issuer",
			authHeader:      "Bearer " + tokenWithWrongIssuer,
			expectedStatus:  http.StatusUnauthorized,
			expectedMessage: "Invalid token issuer",
		},
		{
			name:           "issuer match",
			secret:         testSecret,
			issuer:         "my-issuer",
			authHeader:     "Bearer " + tokenWithIssuer,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
		},
		{
			name:           "empty issuer config allows any issuer",
			secret:         testSecret,
			issuer:         "",
			authHeader:     "Bearer " + tokenWithIssuer,
			expectedStatus: http.StatusOK,
			expectNextCall: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := newTestMiddleware(tc.secret, tc.issuer)

			var nextCalled bool
			handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			}))

			req := httptest.NewRequest("GET", "/test", nil)
			if tc.authHeader != "" {
				req.Header.Set("Authorization", tc.authHeader)
			}
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
			if tc.expectNextCall && !nextCalled {
				t.Fatal("expected next handler to be called")
			}
			if !tc.expectNextCall && nextCalled {
				t.Fatal("next handler should not have been called")
			}
			if tc.expectedMessage != "" {
				var resp map[string]interface{}
				json.NewDecoder(w.Body).Decode(&resp)
				if resp["message"] != tc.expectedMessage {
					t.Fatalf("expected message %q, got %v", tc.expectedMessage, resp["message"])
				}
			}
		})
	}
}

func TestAuthenticate_ClaimsInContext(t *testing.T) {
	m := newTestMiddleware(testSecret, "")
	tokenStr := generateToken(testSecret, "user123", "", time.Hour)

	handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims := GetClaims(r.Context())
		if claims == nil {
			t.Fatal("expected claims in context")
		}
		if claims.Subject != "user123" {
			t.Fatalf("expected sub 'user123', got %q", claims.Subject)
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenStr)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestGetClaims(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		expectNil bool
		expectSub string
	}{
		{
			name:      "no claims in context",
			ctx:       context.Background(),
			expectNil: true,
		},
		{
			name:      "wrong type in context",
			ctx:       context.WithValue(context.Background(), ClaimsContextKey, "not-claims"),
			expectNil: true,
		},
		{
			name: "valid claims in context",
			ctx: context.WithValue(context.Background(), ClaimsContextKey, &JWTClaims{
				RegisteredClaims: jwt.RegisteredClaims{Subject: "user456"},
			}),
			expectNil: false,
			expectSub: "user456",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			claims := GetClaims(tc.ctx)
			if tc.expectNil {
				if claims != nil {
					t.Fatal("expected nil claims")
				}
				return
			}
			if claims == nil {
				t.Fatal("expected non-nil claims")
			}
			if claims.Subject != tc.expectSub {
				t.Fatalf("expected sub %q, got %q", tc.expectSub, claims.Subject)
			}
		})
	}
}

func TestAuthenticate_AdminSecretFallback(t *testing.T) {
	const adminSecret = "admin-secret-key"
	log, _ := logger.NewTestLogger()
	m := NewJWTMiddleware(&config.JWTConfig{
		Secret:      testSecret,
		AdminSecret: adminSecret,
		Issuer:      "lumen-jwt",
	}, log)

	adminToken := generateToken(adminSecret, "admin-user", "", time.Hour)
	wrongToken := generateToken("completely-wrong-secret", "user", "", time.Hour)

	tests := []struct {
		name           string
		token          string
		expectedStatus int
	}{
		{"primary secret still works", generateToken(testSecret, "user", "lumen-jwt", time.Hour), http.StatusOK},
		{"admin secret accepted without issuer", adminToken, http.StatusOK},
		{"wrong secret rejected", wrongToken, http.StatusUnauthorized},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			handler := m.Authenticate(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			req.Header.Set("Authorization", "Bearer "+tc.token)
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
		})
	}
}

func TestSendUnauthorized(t *testing.T) {
	tests := []struct {
		name    string
		message string
	}{
		{name: "simple message", message: "test error"},
		{name: "missing header", message: "Missing Authorization header"},
		{name: "invalid token", message: "Invalid token"},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			m := newTestMiddleware(testSecret, "")
			w := httptest.NewRecorder()
			m.sendUnauthorized(w, tc.message)

			if w.Code != http.StatusUnauthorized {
				t.Fatalf("expected 401, got %d", w.Code)
			}
			if ct := w.Header().Get("Content-Type"); ct != "application/json" {
				t.Fatalf("expected application/json, got %q", ct)
			}
			var resp map[string]interface{}
			json.NewDecoder(w.Body).Decode(&resp)
			if resp["status"] != false {
				t.Fatalf("expected status false, got %v", resp["status"])
			}
			if resp["message"] != tc.message {
				t.Fatalf("expected %q, got %v", tc.message, resp["message"])
			}
		})
	}
}
