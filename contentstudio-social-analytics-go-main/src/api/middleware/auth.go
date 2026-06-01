package middleware

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// ApiKeyValidator looks up an API key in the backing store.
type ApiKeyValidator interface {
	FindValidByKey(ctx context.Context, key string) (*mongoModels.ApiKey, error)
}

// ShareableLinkValidator validates X-Shareable-ID links against the backing store.
type ShareableLinkValidator interface {
	FindActiveUserIDByLinkID(ctx context.Context, linkID string) (string, error)
}

// AuthMiddleware authenticates requests using JWT (Bearer token) or an API key
// validated against the api_keys collection in MongoDB.
// Requests to /health are always allowed.
type AuthMiddleware struct {
	jwt               *JWTMiddleware
	apiKeyRepo        ApiKeyValidator
	shareableLinkRepo ShareableLinkValidator
	logger            *logger.Logger
}

func NewAuthMiddleware(jwtMw *JWTMiddleware, apiKeyRepo ApiKeyValidator, shareableLinkRepo ShareableLinkValidator, log *logger.Logger) *AuthMiddleware {
	return &AuthMiddleware{
		jwt:               jwtMw,
		apiKeyRepo:        apiKeyRepo,
		shareableLinkRepo: shareableLinkRepo,
		logger:            log,
	}
}

func (a *AuthMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			next.ServeHTTP(w, r)
			return
		}

		// Mirror PHP's shareable.optional + shareable.auth flow:
		// If X-Shareable-ID is valid, authenticate immediately and skip JWT validation.
		shareableID := strings.TrimSpace(r.Header.Get("X-Shareable-ID"))
		if strings.HasPrefix(r.URL.Path, "/analytics/") && shareableID != "" && a.shareableLinkRepo != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			userID, err := a.shareableLinkRepo.FindActiveUserIDByLinkID(ctx, shareableID)
			cancel()

			if err != nil {
				a.logger.Error().Err(err).
					Str("path", r.URL.Path).
					Msg("Shareable link lookup failed")
			} else if userID != "" {
				a.logger.Debug().
					Str("path", r.URL.Path).
					Str("user_id", userID).
					Msg("Shareable link authentication successful")
				ctx := context.WithValue(r.Context(), ClaimsContextKey, &JWTClaims{
					RegisteredClaims: jwt.RegisteredClaims{Subject: userID},
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			} else {
				a.logger.Warn().
					Str("path", r.URL.Path).
					Str("remote_addr", r.RemoteAddr).
					Msg("Invalid or disabled shareable link")
			}
		}

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" {
			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) == 2 && strings.ToLower(parts[0]) == "bearer" && a.jwt != nil {
				a.jwt.Authenticate(next).ServeHTTP(w, r)
				return
			}
		}

		apiKeyHeader := r.Header.Get("X-API-Key")
		if apiKeyHeader != "" && a.apiKeyRepo != nil {
			ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
			defer cancel()

			apiKey, err := a.apiKeyRepo.FindValidByKey(ctx, apiKeyHeader)
			if err != nil {
				a.logger.Error().Err(err).
					Str("path", r.URL.Path).
					Msg("API key lookup failed")
				httputil.WriteStatusError(w, http.StatusInternalServerError, "Authentication error")
				return
			}
			if apiKey != nil {
				a.logger.Debug().
					Str("path", r.URL.Path).
					Msg("API key authentication successful")
				ctx := context.WithValue(r.Context(), ClaimsContextKey, &JWTClaims{
					RegisteredClaims: jwt.RegisteredClaims{Subject: userIDString(apiKey.UserID)},
				})
				next.ServeHTTP(w, r.WithContext(ctx))
				return
			}
			a.logger.Warn().
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("Invalid API key")
			a.sendUnauthorized(w, "Invalid API key")
			return
		}

		a.logger.Warn().
			Str("path", r.URL.Path).
			Str("remote_addr", r.RemoteAddr).
			Msg("Missing authentication credentials")
		a.sendUnauthorized(w, "Missing authentication credentials")
	})
}

func (a *AuthMiddleware) sendUnauthorized(w http.ResponseWriter, message string) {
	httputil.WriteStatusError(w, http.StatusUnauthorized, message)
}

// userIDString converts a MongoDB user_id (ObjectID or string) to a string.
func userIDString(id interface{}) string {
	switch v := id.(type) {
	case primitive.ObjectID:
		return v.Hex()
	case string:
		return v
	default:
		return ""
	}
}
