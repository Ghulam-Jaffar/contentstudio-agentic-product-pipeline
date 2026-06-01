package looker_studio

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/middleware"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

type mockApiKeyFinder struct {
	key *mongoModels.ApiKey
	err error
}

func (m *mockApiKeyFinder) FindActiveByUserID(_ context.Context, _ string) (*mongoModels.ApiKey, error) {
	return m.key, m.err
}

func ctxWithClaims(subject string) context.Context {
	claims := &middleware.JWTClaims{
		RegisteredClaims: jwt.RegisteredClaims{Subject: subject},
	}
	return context.WithValue(context.Background(), middleware.ClaimsContextKey, claims)
}

func newHandler(connectorID string, finder ApiKeyFinder) *Handler {
	return NewHandler(
		config.LookerStudioConfig{ConnectorID: connectorID},
		finder,
		zerolog.New(io.Discard),
	)
}

func TestHandleConnect(t *testing.T) {
	tests := []struct {
		name           string
		connectorID    string
		query          string
		ctx            context.Context
		finder         ApiKeyFinder
		expectedStatus int
		checkBody      func(t *testing.T, body map[string]interface{})
	}{
		{
			name:           "missing ConnectorID returns 500",
			connectorID:    "",
			query:          "platform=facebook&workspace_id=ws1&account_id=acc1",
			ctx:            ctxWithClaims("user1"),
			finder:         &mockApiKeyFinder{key: &mongoModels.ApiKey{Key: "k"}},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "missing platform returns 400",
			connectorID:    "connector-id",
			query:          "workspace_id=ws1&account_id=acc1",
			ctx:            ctxWithClaims("user1"),
			finder:         &mockApiKeyFinder{key: &mongoModels.ApiKey{Key: "k"}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing workspace_id returns 400",
			connectorID:    "connector-id",
			query:          "platform=facebook&account_id=acc1",
			ctx:            ctxWithClaims("user1"),
			finder:         &mockApiKeyFinder{key: &mongoModels.ApiKey{Key: "k"}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing account_id returns 400",
			connectorID:    "connector-id",
			query:          "platform=facebook&workspace_id=ws1",
			ctx:            ctxWithClaims("user1"),
			finder:         &mockApiKeyFinder{key: &mongoModels.ApiKey{Key: "k"}},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unauthenticated (no claims) returns 401",
			connectorID:    "connector-id",
			query:          "platform=facebook&workspace_id=ws1&account_id=acc1",
			ctx:            context.Background(),
			finder:         &mockApiKeyFinder{key: &mongoModels.ApiKey{Key: "k"}},
			expectedStatus: http.StatusUnauthorized,
		},
		{
			name:           "api key lookup error returns 500",
			connectorID:    "connector-id",
			query:          "platform=facebook&workspace_id=ws1&account_id=acc1",
			ctx:            ctxWithClaims("user1"),
			finder:         &mockApiKeyFinder{err: errors.New("db error")},
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "no active api key returns 400",
			connectorID:    "connector-id",
			query:          "platform=facebook&workspace_id=ws1&account_id=acc1",
			ctx:            ctxWithClaims("user1"),
			finder:         &mockApiKeyFinder{key: nil},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "fresh flow returns 200 with datasources URL",
			connectorID:    "connector-id",
			query:          "platform=facebook&workspace_id=ws1&account_id=acc1",
			ctx:            ctxWithClaims("user1"),
			finder:         &mockApiKeyFinder{key: &mongoModels.ApiKey{Key: "api-key-value"}},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				t.Helper()
				if body["status"] != true {
					t.Fatalf("expected status true, got %v", body["status"])
				}
				u, _ := body["url"].(string)
				if !strings.HasPrefix(u, lookerStudioBaseURL) {
					t.Fatalf("expected datasources URL, got %q", u)
				}
				if !strings.Contains(u, "connector-id") {
					t.Fatalf("expected connectorId in URL, got %q", u)
				}
			},
		},
		{
			name:           "template flow returns 200 with reporting URL",
			connectorID:    "connector-id",
			query:          "platform=facebook&workspace_id=ws1&account_id=acc1&template_id=tmpl-123",
			ctx:            ctxWithClaims("user1"),
			finder:         &mockApiKeyFinder{key: &mongoModels.ApiKey{Key: "api-key-value"}},
			expectedStatus: http.StatusOK,
			checkBody: func(t *testing.T, body map[string]interface{}) {
				t.Helper()
				if body["status"] != true {
					t.Fatalf("expected status true, got %v", body["status"])
				}
				u, _ := body["url"].(string)
				if !strings.HasPrefix(u, lookerStudioTemplateBaseURL) {
					t.Fatalf("expected reporting URL, got %q", u)
				}
				if !strings.Contains(u, "tmpl-123") {
					t.Fatalf("expected template ID in URL, got %q", u)
				}
				if !strings.Contains(u, "connector=community") {
					t.Fatalf("expected connector=community in URL, got %q", u)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			h := newHandler(tc.connectorID, tc.finder)
			req := httptest.NewRequest("GET", "/analytics/looker-studio/connect?"+tc.query, nil)
			req = req.WithContext(tc.ctx)
			w := httptest.NewRecorder()

			h.HandleConnect(w, req)

			if w.Code != tc.expectedStatus {
				t.Fatalf("expected %d, got %d", tc.expectedStatus, w.Code)
			}
			if tc.checkBody != nil {
				var body map[string]interface{}
				if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
					t.Fatalf("failed to decode response body: %v", err)
				}
				tc.checkBody(t, body)
			}
		})
	}
}
