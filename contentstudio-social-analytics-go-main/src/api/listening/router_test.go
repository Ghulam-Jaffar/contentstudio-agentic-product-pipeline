package listening

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	mentionsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/mentions"
)

func TestRegisterRoutes(t *testing.T) {
	t.Parallel()

	stub := &stubMentionReader{}
	svc := mentionsSvc.NewService(stub, zerolog.Nop())
	resolver := NewMentionFilterResolver(stubViewResolver{}, stubTopicWorkspaceResolver{})

	mentions := NewMentionsHandler(svc, resolver, zerolog.Nop())
	analytics := NewAnalyticsHandler(svc, resolver, zerolog.Nop())
	views := NewViewsHandler(nil, zerolog.Nop())

	mux := http.NewServeMux()
	RegisterRoutes(mux, mentions, views, analytics)

	// serveAndCapture runs a request through the mux, recovering from any
	// nil-service panics that happen once a registered handler is invoked.
	// A 404 means no route matched; any other status means a route was found.
	serveAndCapture := func(method, path string) int {
		req := httptest.NewRequest(method, path, nil)
		rec := httptest.NewRecorder()
		func() {
			defer func() { recover() }()
			mux.ServeHTTP(rec, req)
		}()
		return rec.Code
	}

	tests := []struct {
		name    string
		method  string
		path    string
		want404 bool
	}{
		// Registered routes — all should be reachable (non-404).
		// Requests are crafted to trigger validation failures before any service
		// call, so they return real HTTP status codes rather than panicking.
		{"GET mentions", "GET", "/api/listening/mentions?limit=bad", false},
		{"PATCH mention by id", "PATCH", "/api/listening/mentions/m1", false},
		{"POST mark-all-read", "POST", "/api/listening/mentions/mark-all-read", false},
		{"GET unread-count", "GET", "/api/listening/mentions/unread-count?limit=bad", false},
		{"GET analytics", "GET", "/api/listening/analytics?limit=bad", false},
		{"GET analytics export", "GET", "/api/listening/analytics/export?limit=bad", false},
		{"GET views missing workspace_id", "GET", "/api/listening/views", false},
		{"POST views empty body", "POST", "/api/listening/views", false},
		{"PUT views invalid body", "PUT", "/api/listening/views/v1", false},
		{"DELETE views by id", "DELETE", "/api/listening/views/v1", false},

		// Completely unknown paths should 404.
		{"unknown path", "GET", "/api/listening/does-not-exist", true},
		{"unknown nested path", "GET", "/api/listening/mentions/unknown/nested", true},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			code := serveAndCapture(tc.method, tc.path)
			got404 := code == http.StatusNotFound
			if got404 != tc.want404 {
				t.Errorf("%s %s: want 404=%v, got status=%d", tc.method, tc.path, tc.want404, code)
			}
		})
	}
}
