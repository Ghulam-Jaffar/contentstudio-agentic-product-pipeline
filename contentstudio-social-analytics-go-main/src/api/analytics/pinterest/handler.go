// Package pinterest provides HTTP handlers for Pinterest analytics API endpoints.
// Each handler parses query parameters, delegates to the service layer, and writes JSON responses.
//
// This is the HTTP layer of the 3-layer architecture:
//
//	Handler (this package) → Service (services/analytics/pinterest) → Repository (db/clickhouse/analytics-get-queries/pinterest)
package pinterest

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/pinterest"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/pinterest"
)

// Handler handles HTTP requests for Pinterest analytics endpoints.
type Handler struct {
	service    service.Service
	aiInsights *service.AIInsightsService
	logger     zerolog.Logger
}

// NewHandler creates a new handler with the given service and logger.
func NewHandler(svc service.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With().Str("handler", "pinterest-analytics").Logger(),
	}
}

// SetAIInsightsService attaches the AI insights service to this handler.
func (h *Handler) SetAIInsightsService(aiSvc *service.AIInsightsService) {
	h.aiInsights = aiSvc
}

// parseBaseRequest extracts and validates common Pinterest query parameters.
// board_id is optional; when absent the request runs in user mode.
func parseBaseRequest(r *http.Request) (*types.PinterestRequest, error) {
	q := r.URL.Query()
	startDate := q.Get("start_date")
	endDate := q.Get("end_date")
	if startDate == "" || endDate == "" {
		if s, e, ok := parseDateRangeQuery(q.Get("date")); ok {
			if startDate == "" {
				startDate = s
			}
			if endDate == "" {
				endDate = e
			}
		}
	}
	req := &types.PinterestRequest{
		WorkspaceID: q.Get("workspace_id"),
		PinterestID: q.Get("pinterest_id"),
		BoardID:     q.Get("board_id"),
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    q.Get("timezone"),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func parseDateRangeQuery(raw string) (string, string, bool) {
	parts := strings.SplitN(strings.TrimSpace(raw), " - ", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	start := strings.TrimSpace(parts[0])
	end := strings.TrimSpace(parts[1])
	if start == "" || end == "" {
		return "", "", false
	}
	return start, end, true
}

// HandleSummary handles GET /analytics/overview/pinterest/summary
func (h *Handler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := h.service.GetSummary(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleFollowerTrend handles GET /analytics/overview/pinterest/followerTrend
func (h *Handler) HandleFollowerTrend(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := h.service.GetFollowerTrend(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleDynamicFollowerTrend handles GET /analytics/overview/pinterest/dynamicFollowerTrend
func (h *Handler) HandleDynamicFollowerTrend(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := h.service.GetDynamicFollowerTrend(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleImpressionsTrend handles GET /analytics/overview/pinterest/impressionsTrend
func (h *Handler) HandleImpressionsTrend(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := h.service.GetImpressionsTrend(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleDynamicImpressionsTrend handles GET /analytics/overview/pinterest/dynamicImpressionsTrend
func (h *Handler) HandleDynamicImpressionsTrend(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := h.service.GetDynamicImpressionsTrend(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleEngagementTrend handles GET /analytics/overview/pinterest/engagementTrend
func (h *Handler) HandleEngagementTrend(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := h.service.GetEngagementTrend(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleDynamicEngagementTrend handles GET /analytics/overview/pinterest/dynamicEngagementTrend
func (h *Handler) HandleDynamicEngagementTrend(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := h.service.GetDynamicEngagementTrend(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandlePinPosting handles GET /analytics/overview/pinterest/pinPosting
// Parses optional filter_by param ('video' or 'image').
func (h *Handler) HandlePinPosting(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	filteredReq := &types.FilteredPinRequest{
		PinterestRequest: *req,
		FilterBy:         r.URL.Query().Get("filter_by"),
	}
	resp, err := h.service.GetPinPosting(r.Context(), filteredReq)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleDynamicPinPosting handles GET /analytics/overview/pinterest/dynamicPinPosting
func (h *Handler) HandleDynamicPinPosting(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	filteredReq := &types.FilteredPinRequest{
		PinterestRequest: *req,
		FilterBy:         r.URL.Query().Get("filter_by"),
	}
	resp, err := h.service.GetDynamicPinPosting(r.Context(), filteredReq)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandlePinRollup handles GET /analytics/overview/pinterest/pinRollup
func (h *Handler) HandlePinRollup(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := h.service.GetPinRollup(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleTopPins handles GET /analytics/overview/pinterest/topPins
// Parses optional order_by and limit params.
func (h *Handler) HandleTopPins(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	q := r.URL.Query()
	topReq := &types.TopPinsRequest{
		PinterestRequest: *req,
		OrderBy:          q.Get("order_by"),
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			httputil.WriteError(w, h.logger, httputil.NewValidationError("limit must be a valid integer"))
			return
		}
		topReq.Limit = limit
	}
	if topReq.Limit <= 0 {
		topReq.Limit = 5
	}
	resp, err := h.service.GetTopPins(r.Context(), topReq)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandlePinPerformance handles GET /analytics/overview/pinterest/pinPerformance
func (h *Handler) HandlePinPerformance(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := h.service.GetPinPerformance(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleAIInsights handles GET /analytics/overview/pinterest/ai_insights
func (h *Handler) HandleAIInsights(w http.ResponseWriter, r *http.Request) {
	if h.aiInsights == nil {
		httputil.WriteStatusError(w, http.StatusServiceUnavailable, "AI insights service not configured")
		return
	}

	q := r.URL.Query()
	req := types.AIInsightsRequest{
		WorkspaceID: q.Get("workspace_id"),
		PinterestID: q.Get("pinterest_id"),
		BoardID:     q.Get("board_id"),
		Date:        q.Get("date"),
		Timezone:    q.Get("timezone"),
		Type:        q.Get("type"),
		Language:    q.Get("language"),
	}
	if req.Date == "" {
		startDate := q.Get("start_date")
		endDate := q.Get("end_date")
		if startDate != "" && endDate != "" {
			req.Date = strings.TrimSpace(startDate) + " - " + strings.TrimSpace(endDate)
		}
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			httputil.WriteStatusError(w, http.StatusBadRequest, "limit must be a valid integer")
			return
		}
		req.Limit = limit
	}

	if req.WorkspaceID == "" || req.PinterestID == "" || req.Type == "" || req.Date == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "workspace_id, pinterest_id, date, and type are required")
		return
	}
	if req.Limit <= 0 {
		req.Limit = 15
	}

	locale := r.Header.Get("X-LOCALE")
	if locale != "" && req.Language == "" {
		req.Language = locale
	}

	result, err := h.aiInsights.GetAIInsights(r.Context(), &req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}
