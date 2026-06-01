package listening

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
)

// newViewsHandlerNoService creates a ViewsHandler with a nil service.
// Only use it in test cases that return before any service method is called
// (i.e. validation failures).
func newViewsHandlerNoService() *ViewsHandler {
	return NewViewsHandler(nil, zerolog.Nop())
}

func TestHandleListViews(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		url        string
		wantStatus int
	}{
		{
			name:       "returns 400 when workspace_id is missing",
			url:        "/api/listening/views",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 when workspace_id is empty string",
			url:        "/api/listening/views?workspace_id=",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newViewsHandlerNoService()
			req := httptest.NewRequest(http.MethodGet, tc.url, nil)
			rec := httptest.NewRecorder()
			h.HandleListViews(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleCreateView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "returns 400 on invalid JSON body",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on empty body",
			body:       nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newViewsHandlerNoService()
			req := httptest.NewRequest(http.MethodPost, "/api/listening/views", bodyOf(tc.body))
			rec := httptest.NewRecorder()
			h.HandleCreateView(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleUpdateView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		viewID     string
		body       interface{}
		wantStatus int
	}{
		{
			name:       "returns 400 when view id path value is absent",
			viewID:     "",
			body:       apiModels.ViewRequest{Name: "My View"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on invalid JSON body",
			viewID:     "v1",
			body:       "not-json",
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "returns 400 on empty body with valid id",
			viewID:     "v1",
			body:       nil,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newViewsHandlerNoService()
			req := httptest.NewRequest(http.MethodPut, "/api/listening/views/"+tc.viewID, bodyOf(tc.body))
			if tc.viewID != "" {
				req.SetPathValue("id", tc.viewID)
			}
			rec := httptest.NewRecorder()
			h.HandleUpdateView(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestHandleDeleteView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		viewID     string
		wantStatus int
	}{
		{
			name:       "returns 400 when view id path value is absent",
			viewID:     "",
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			h := newViewsHandlerNoService()
			req := httptest.NewRequest(http.MethodDelete, "/api/listening/views/"+tc.viewID, nil)
			rec := httptest.NewRecorder()
			h.HandleDeleteView(rec, req)

			if rec.Code != tc.wantStatus {
				t.Errorf("status: want %d, got %d (body: %s)", tc.wantStatus, rec.Code, rec.Body.String())
			}
		})
	}
}
