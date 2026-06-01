package listening

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	viewsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/views"
)

type ViewsHandler struct {
	service *viewsSvc.Service
	logger  zerolog.Logger
}

func NewViewsHandler(svc *viewsSvc.Service, logger zerolog.Logger) *ViewsHandler {
	return &ViewsHandler{
		service: svc,
		logger:  logger.With().Str("handler", "listening-views").Logger(),
	}
}

func (h *ViewsHandler) HandleListViews(w http.ResponseWriter, r *http.Request) {
	workspaceID := r.URL.Query().Get("workspace_id")
	if workspaceID == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "workspace_id is required")
		return
	}

	views, err := h.service.ListViews(r.Context(), workspaceID)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   views,
	})
}

func (h *ViewsHandler) HandleCreateView(w http.ResponseWriter, r *http.Request) {
	var req apiModels.ViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteStatusError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	view, err := h.service.CreateView(r.Context(), &req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, map[string]interface{}{
		"status": true,
		"data":   view,
	})
}

func (h *ViewsHandler) HandleUpdateView(w http.ResponseWriter, r *http.Request) {
	viewID := r.PathValue("id")
	if viewID == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "view id is required")
		return
	}

	var req apiModels.ViewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteStatusError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	view, err := h.service.UpdateView(r.Context(), viewID, &req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status": true,
		"data":   view,
	})
}

func (h *ViewsHandler) HandleDeleteView(w http.ResponseWriter, r *http.Request) {
	viewID := r.PathValue("id")
	if viewID == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "view id is required")
		return
	}

	if err := h.service.DeleteView(r.Context(), viewID); err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "view deleted",
	})
}
