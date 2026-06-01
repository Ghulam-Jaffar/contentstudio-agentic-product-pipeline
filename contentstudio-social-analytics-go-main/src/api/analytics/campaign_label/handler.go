// Package campaign_label provides HTTP handlers for campaign & label analytics API endpoints.
// All endpoints use POST because their payloads contain variable-length arrays of campaigns,
// labels, and per-platform account IDs that would be impractical as query parameters.
//
// This is the HTTP layer of the 3-layer architecture:
//
//	Handler (this package) → Service (services/analytics/campaign_label) → Repository (db/clickhouse/analytics-get-queries/campaign_label)
//
// Migrated from PHP: CampaignLabelAnalyticsController (contentstudio-backend).
package campaign_label

import (
	"encoding/json"
	"net/http"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/campaign_label"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/campaign_label"
)

// Handler handles HTTP requests for campaign/label analytics endpoints.
type Handler struct {
	service service.Service
	logger  zerolog.Logger
}

// NewHandler creates a new handler with the given service and logger.
func NewHandler(svc service.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With().Str("handler", "campaign-label-analytics").Logger(),
	}
}

// parseCampaignLabelBody decodes and validates a CampaignLabelRequest from the POST body.
func parseCampaignLabelBody(r *http.Request) (*types.CampaignLabelRequest, error) {
	var body types.CampaignLabelRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		return nil, httputil.NewValidationError("invalid request body")
	}
	if err := body.Validate(); err != nil {
		return nil, err
	}
	return &body, nil
}

// HandleSetPostIds handles POST /analytics/campaignLabelAnalytics/setPostIdsForCampaignsAndLabels.
// Resolves campaign/label IDs to their aggregated social platform post IDs via MongoDB.
// Falls back to plans → postings collections for campaigns/labels not yet cached.
func (h *Handler) HandleSetPostIds(w http.ResponseWriter, r *http.Request) {
	body, err := parseCampaignLabelBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	resp, err := h.service.SetPostIds(r.Context(), body)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleSummaryAnalytics handles POST /analytics/campaignLabelAnalytics/getSummaryAnalytics.
// Returns cross-platform aggregated metrics (posts, engagement, impressions) for current
// and previous periods with computed differences and percentage changes.
func (h *Handler) HandleSummaryAnalytics(w http.ResponseWriter, r *http.Request) {
	body, err := parseCampaignLabelBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	resp, err := h.service.GetSummaryAnalytics(r.Context(), body)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleBreakdownData handles POST /analytics/campaignLabelAnalytics/getCampaignLabelBreakdownData.
// Returns per-campaign/label breakdown of posts, engagement, and impressions for
// current and previous periods, grouped by campaign/label ID and era.
func (h *Handler) HandleBreakdownData(w http.ResponseWriter, r *http.Request) {
	body, err := parseCampaignLabelBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	resp, err := h.service.GetBreakdownData(r.Context(), body)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleInsightsBreakdown handles POST /analytics/campaignLabelAnalytics/getCampaignLabelInsightsBreakdown.
// Returns time-series insights data (daily engagement, impressions, post counts)
// grouped by campaign/label ID for charting.
func (h *Handler) HandleInsightsBreakdown(w http.ResponseWriter, r *http.Request) {
	body, err := parseCampaignLabelBody(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	resp, err := h.service.GetInsightsBreakdown(r.Context(), body)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandlePlannerAnalytics handles POST /analytics/campaignLabelAnalytics/getPlannerAnalytics.
// Returns detailed per-post analytics for a single platform, used by the planner view.
// Uses POST because it accepts an array of post IDs in the payload.
func (h *Handler) HandlePlannerAnalytics(w http.ResponseWriter, r *http.Request) {
	var body types.PlannerAnalyticsRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, h.logger, httputil.NewValidationError("invalid request body"))
		return
	}
	if err := body.Validate(); err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	resp, err := h.service.GetPlannerAnalytics(r.Context(), &body)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}
