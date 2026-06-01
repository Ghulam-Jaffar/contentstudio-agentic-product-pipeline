// Package overview provides HTTP handlers for the cross-platform Overview analytics API endpoints.
// Each handler parses query parameters, delegates to the service layer, and writes JSON responses.
//
// This is the HTTP layer of the 3-layer architecture:
//
//	Handler (this package) → Service (services/analytics/overview) → Repository (db/clickhouse/analytics-get-queries/overview)
//
// Migrated from PHP: OverviewV2Controller (contentstudio-backend).
package overview

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/overview"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/overview"
)

// Handler handles HTTP requests for Overview analytics endpoints.
type Handler struct {
	service    service.Service
	aiInsights *service.AIInsightsService
	logger     zerolog.Logger
}

// NewHandler creates a new handler with the given service and logger.
func NewHandler(svc service.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With().Str("handler", "overview-analytics").Logger(),
	}
}

// SetAIInsightsService attaches the AI insights service to this handler.
func (h *Handler) SetAIInsightsService(aiSvc *service.AIInsightsService) {
	h.aiInsights = aiSvc
}

type overviewBody struct {
	types.OverviewRequest
	Type  string `json:"type"`
	Limit int    `json:"limit"`
}

func parseOverviewBody(r *http.Request) (*overviewBody, error) {
	var body overviewBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, httputil.NewValidationError("invalid request body")
	}
	if body.StartDate == "" || body.EndDate == "" {
		return nil, httputil.NewValidationError("start_date and end_date are required")
	}
	return &body, nil
}

// HandleSummary handles POST /analytics/overview/overviewV2/getSummary.
// Returns cross-platform aggregated metrics for current and secondary periods with pct changes.
func (h *Handler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	body, err := parseOverviewBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	resp, err := h.service.GetSummary(r.Context(), &body.OverviewRequest)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleTopPerformingGraph handles POST /analytics/overview/overviewV2/getTopPerformingGraph.
// Returns per-platform daily time-series arrays from mv_social_daily_metrics.
func (h *Handler) HandleTopPerformingGraph(w http.ResponseWriter, r *http.Request) {
	body, err := parseOverviewBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	resp, err := h.service.GetTopPerformingGraph(r.Context(), &body.OverviewRequest)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandlePlatformData handles POST /analytics/overview/overviewV2/getPlatformData.
// When type=grouped: returns per-platform aggregated rows.
// When type=individual (or any other value): returns per-account aggregated rows.
func (h *Handler) HandlePlatformData(w http.ResponseWriter, r *http.Request) {
	body, err := parseOverviewBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	if body.Type == "grouped" {
		rows, err := h.service.GetPlatformDataGrouped(r.Context(), &body.OverviewRequest)
		if err != nil {
			httputil.WriteError(w, h.logger, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, rows)
	} else {
		rows, err := h.service.GetPlatformDataIndividual(r.Context(), &body.OverviewRequest)
		if err != nil {
			httputil.WriteError(w, h.logger, err)
			return
		}
		httputil.WriteJSON(w, http.StatusOK, rows)
	}
}

// HandlePlatformDataDetailed handles POST /analytics/overview/overviewV2/getPlatformDataDetailed.
// Returns current/previous period metrics with pct changes per account.
func (h *Handler) HandlePlatformDataDetailed(w http.ResponseWriter, r *http.Request) {
	body, err := parseOverviewBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	rows, err := h.service.GetPlatformDataDetailed(r.Context(), &body.OverviewRequest)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// HandlePlatformDataGraphs handles POST /analytics/overview/overviewV2/getPlatformDataGraphs.
// Returns per-account time-series engagement/reach/impressions/posts arrays.
func (h *Handler) HandlePlatformDataGraphs(w http.ResponseWriter, r *http.Request) {
	body, err := parseOverviewBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	rows, err := h.service.GetPlatformDataGraphs(r.Context(), &body.OverviewRequest)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// HandleTopPosts handles POST /analytics/overview/overviewV2/getTopPosts.
// Returns up to N posts per selected platform, globally ordered by the specified metric.
func (h *Handler) HandleTopPosts(w http.ResponseWriter, r *http.Request) {
	body, err := parseOverviewBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	limit := 20
	if body.Limit > 0 {
		limit = body.Limit
	}

	req := &types.TopPostsRequest{
		OverviewRequest: body.OverviewRequest,
		Type:            body.Type,
		Limit:           limit,
	}

	rows, err := h.service.GetTopPosts(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, rows)
}

// HandleAIInsights handles POST /analytics/overview/overviewV2/ai_insights.
// Dispatches to the AI agent based on the "type" field in the request body.
func (h *Handler) HandleAIInsights(w http.ResponseWriter, r *http.Request) {
	if h.aiInsights == nil {
		httputil.WriteStatusError(w, http.StatusServiceUnavailable, "AI insights service not configured")
		return
	}

	body, err := parseOverviewBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	limit := body.Limit
	if limit <= 0 {
		limit = 5
	}

	req := &types.AIInsightsRequest{
		OverviewRequest: body.OverviewRequest,
		Type:            body.Type,
		Limit:           limit,
	}
	if locale := r.Header.Get("X-LOCALE"); locale != "" {
		req.Language = locale
	}

	result, err := h.aiInsights.GetAIInsights(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}
