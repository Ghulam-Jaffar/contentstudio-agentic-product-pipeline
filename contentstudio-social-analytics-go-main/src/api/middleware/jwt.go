// Package middleware provides HTTP middleware for the analytics API server.
// Currently includes JWT authentication middleware using HMAC-SHA256 signed tokens.
package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/golang-jwt/jwt/v5"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

type contextKey string

// ClaimsContextKey is the context key used to store JWT claims after authentication.
const (
	ClaimsContextKey contextKey = "jwt_claims"
)

// JWTMiddleware validates Bearer tokens on incoming requests.
type JWTMiddleware struct {
	Config *config.JWTConfig
	Logger *logger.Logger
}

// JWTClaims extends RegisteredClaims for ContentStudio tokens.
// The "sub" claim is handled by RegisteredClaims.Subject.
type JWTClaims struct {
	jwt.RegisteredClaims
}

// NewJWTMiddleware creates a new JWT middleware with the given configuration and logger.
func NewJWTMiddleware(cfg *config.JWTConfig, log *logger.Logger) *JWTMiddleware {
	return &JWTMiddleware{
		Config: cfg,
		Logger: log,
	}
}

// Authenticate returns an http.Handler that validates the Bearer token and injects claims into context.
// Returns 401 for missing, malformed, or invalid tokens. Supports optional issuer validation.
func (m *JWTMiddleware) Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			m.Logger.Warn().
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("Missing Authorization header")
			m.sendUnauthorized(w, "Missing Authorization header")
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			m.Logger.Warn().
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("Invalid Authorization header format")
			m.sendUnauthorized(w, "Invalid Authorization header format")
			return
		}

		tokenString := parts[1]

		isAdmin := false
		token, err := jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(m.Config.Secret), nil
		})

		if err != nil && m.Config.AdminSecret != "" {
			token, err = jwt.ParseWithClaims(tokenString, &JWTClaims{}, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(m.Config.AdminSecret), nil
			})
			if err == nil {
				isAdmin = true
			}
		}

		if err != nil {
			m.Logger.Warn().
				Err(err).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("Invalid JWT token")
			m.sendUnauthorized(w, "Invalid token")
			return
		}

		claims, ok := token.Claims.(*JWTClaims)
		if !ok || !token.Valid {
			m.Logger.Warn().
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("Invalid JWT claims")
			m.sendUnauthorized(w, "Invalid token claims")
			return
		}

		if !isAdmin && m.Config.Issuer != "" && claims.Issuer != m.Config.Issuer {
			m.Logger.Warn().
				Str("expected_issuer", m.Config.Issuer).
				Str("actual_issuer", claims.Issuer).
				Str("path", r.URL.Path).
				Str("remote_addr", r.RemoteAddr).
				Msg("Invalid JWT issuer")
			m.sendUnauthorized(w, "Invalid token issuer")
			return
		}

		m.Logger.Debug().
			Str("subject", claims.Subject).
			Str("path", r.URL.Path).
			Msg("JWT authentication successful")

		ctx := context.WithValue(r.Context(), ClaimsContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// sendUnauthorized writes a 401 JSON error response.
func (m *JWTMiddleware) sendUnauthorized(w http.ResponseWriter, message string) {
	httputil.WriteStatusError(w, http.StatusUnauthorized, message)
}

// GetClaims extracts JWT claims from the request context. Returns nil if not present.
func GetClaims(ctx context.Context) *JWTClaims {
	claims, ok := ctx.Value(ClaimsContextKey).(*JWTClaims)
	if !ok {
		return nil
	}
	return claims
}
