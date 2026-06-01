package main

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

func TestCorsMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		origin     string
		wantStatus int
	}{
		{
			name:       "OPTIONS returns 204 without calling next",
			method:     http.MethodOptions,
			origin:     "https://example.com",
			wantStatus: http.StatusNoContent,
		},
		{
			name:       "GET passes through with CORS headers",
			method:     http.MethodGet,
			origin:     "http://127.0.0.1:5173",
			wantStatus: http.StatusOK,
		},
		{
			name:       "POST passes through with CORS headers",
			method:     http.MethodPost,
			origin:     "https://other.example.com",
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			})
			handler := corsMiddleware(next)

			req := httptest.NewRequest(tc.method, "/test", nil)
			req.Header.Set("Origin", tc.origin)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d", tc.wantStatus, rec.Code)
			}
			if got := rec.Header().Get("Access-Control-Allow-Origin"); got != tc.origin {
				t.Errorf("Access-Control-Allow-Origin: want %q, got %q", tc.origin, got)
			}
		})
	}
}

func TestMethodOverrideMiddleware(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		method     string
		override   string
		wantMethod string
	}{
		{"GET ignores override header", http.MethodGet, "DELETE", http.MethodGet},
		{"POST + DELETE override uses DELETE", http.MethodPost, "DELETE", http.MethodDelete},
		{"POST + PUT override uses PUT", http.MethodPost, "PUT", http.MethodPut},
		{"POST + PATCH override uses PATCH", http.MethodPost, "PATCH", http.MethodPatch},
		{"POST + GET override uses GET", http.MethodPost, "GET", http.MethodGet},
		{"POST without header stays POST", http.MethodPost, "", http.MethodPost},
		{"POST + CONNECT stays POST (not in allowed set)", http.MethodPost, "CONNECT", http.MethodPost},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			var gotMethod string
			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotMethod = r.Method
			})
			handler := methodOverrideMiddleware(next)

			req := httptest.NewRequest(tc.method, "/test", nil)
			if tc.override != "" {
				req.Header.Set("X-Http-Method-Override", tc.override)
			}
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if gotMethod != tc.wantMethod {
				t.Errorf("method: want %q, got %q", tc.wantMethod, gotMethod)
			}
		})
	}
}

func TestLoggingMiddleware(t *testing.T) {
	t.Parallel()

	log, _ := logger.NewTestLogger()

	tests := []struct {
		name       string
		respStatus int
	}{
		{"passes through 200", http.StatusOK},
		{"passes through 201", http.StatusCreated},
		{"passes through 404", http.StatusNotFound},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tc.respStatus)
			})
			handler := loggingMiddleware(log, next)

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			rec := httptest.NewRecorder()
			handler.ServeHTTP(rec, req)

			if rec.Code != tc.respStatus {
				t.Errorf("status: want %d, got %d", tc.respStatus, rec.Code)
			}
		})
	}
}
