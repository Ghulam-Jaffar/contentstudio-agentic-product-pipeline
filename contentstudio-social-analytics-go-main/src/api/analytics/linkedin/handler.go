// Package linkedin provides HTTP handlers for LinkedIn analytics API endpoints.
// Each handler parses query parameters, delegates to the service layer, and writes JSON responses.
//
// This is the HTTP layer of the 3-layer architecture:
//
//	Handler (this package) → Service (services/analytics/linkedin) → Repository (db/clickhouse/analytics-get-queries/linkedin)
//
// Migrated from PHP: LinkedInAnalyticsController (contentstudio-backend).
package linkedin

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/linkedin"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/linkedin"
)

// LinkedInHandler handles HTTP requests for LinkedIn analytics endpoints.
type LinkedInHandler struct {
	service    service.Service
	aiInsights *service.AIInsightsService
	logger     zerolog.Logger
}

// NewLinkedInHandler creates a new handler with the given service and logger.
func NewLinkedInHandler(svc service.Service, logger zerolog.Logger) *LinkedInHandler {
	return &LinkedInHandler{
		service: svc,
		logger:  logger.With().Str("handler", "linkedin-analytics").Logger(),
	}
}

func (h *LinkedInHandler) SetAIInsightsService(aiSvc *service.AIInsightsService) {
	h.aiInsights = aiSvc
}

// parseBaseRequest extracts and validates common query parameters from the HTTP request.
func parseBaseRequest(r *http.Request) (*types.LinkedInRequest, error) {
	q := r.URL.Query()
	startDate := q.Get("start_date")
	endDate := q.Get("end_date")
	if startDate == "" || endDate == "" {
		if parsedStart, parsedEnd, ok := parseDateRangeQuery(q.Get("date")); ok {
			if startDate == "" {
				startDate = parsedStart
			}
			if endDate == "" {
				endDate = parsedEnd
			}
		}
	}
	req := &types.LinkedInRequest{
		WorkspaceID: q.Get("workspace_id"),
		LinkedinID:  q.Get("linkedin_id"),
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

	startDate := strings.TrimSpace(parts[0])
	endDate := strings.TrimSpace(parts[1])
	if startDate == "" || endDate == "" {
		return "", "", false
	}

	return startDate, endDate, true
}

// handleBase is a helper that parses the request, calls the service, and writes the response.
func (h *LinkedInHandler) handleBase(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.LinkedInRequest) (interface{}, error)) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	resp, err := fn(r.Context(), req)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleSummary handles GET /analytics/overview/linkedin/summary
func (h *LinkedInHandler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.LinkedInRequest) (interface{}, error) {
		return h.service.GetSummary(ctx, req)
	})
}

// HandleAudienceGrowth handles GET /analytics/overview/linkedin/audienceGrowth
func (h *LinkedInHandler) HandleAudienceGrowth(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.LinkedInRequest) (interface{}, error) {
		return h.service.GetAudienceGrowth(ctx, req)
	})
}

// HandlePageViews handles GET /analytics/overview/linkedin/pageViews
func (h *LinkedInHandler) HandlePageViews(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.LinkedInRequest) (interface{}, error) {
		return h.service.GetPageViews(ctx, req)
	})
}

// HandlePublishingBehaviour handles GET /analytics/overview/linkedin/publishingBehaviour
// Accepts an additional media_type query parameter (comma-separated) for filtering.
func (h *LinkedInHandler) HandlePublishingBehaviour(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	pubReq := &types.PublishingBehaviourRequest{LinkedInRequest: *req}
	mediaType := r.URL.Query().Get("media_type")
	if mediaType != "" {
		pubReq.MediaType = strings.Split(mediaType, ",")
	}
	resp, err := h.service.GetPublishingBehaviour(r.Context(), pubReq)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleTopPosts handles GET /analytics/overview/linkedin/topPosts (default limit: 3)
func (h *LinkedInHandler) HandleTopPosts(w http.ResponseWriter, r *http.Request) {
	h.handleTopPostsWithDefault(w, r, 3)
}

// HandleGetTopPosts handles GET /analytics/overview/linkedin/getTopPosts (default limit: 15)
func (h *LinkedInHandler) HandleGetTopPosts(w http.ResponseWriter, r *http.Request) {
	h.handleTopPostsWithDefault(w, r, 15)
}

// handleTopPostsWithDefault is a shared handler for topPosts and getTopPosts with configurable default limit.
// Accepts optional query params: limit, order_by, hashtags (comma-separated).
func (h *LinkedInHandler) handleTopPostsWithDefault(w http.ResponseWriter, r *http.Request, defaultLimit int) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	topReq := &types.TopPostsRequest{LinkedInRequest: *req}
	q := r.URL.Query()
	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			httputil.WriteError(w, h.logger, httputil.NewValidationError("limit must be a valid integer"))
			return
		}
		topReq.Limit = limit
	}
	if topReq.Limit <= 0 {
		topReq.Limit = defaultLimit
	}
	topReq.OrderBy = q.Get("order_by")
	if hashtags := q.Get("hashtags"); hashtags != "" {
		topReq.Hashtags = strings.Split(hashtags, ",")
	}
	resp, err := h.service.GetTopPosts(r.Context(), topReq)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandlePostsPerDay handles GET /analytics/overview/linkedin/postsPerDays
func (h *LinkedInHandler) HandlePostsPerDay(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.LinkedInRequest) (interface{}, error) {
		return h.service.GetPostsPerDay(ctx, req)
	})
}

// HandleHashtags handles GET /analytics/overview/linkedin/hashtags
func (h *LinkedInHandler) HandleHashtags(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.LinkedInRequest) (interface{}, error) {
		return h.service.GetHashtags(ctx, req)
	})
}

// HandleFollowersDemographics handles GET /analytics/overview/linkedin/followersDemographics
func (h *LinkedInHandler) HandleFollowersDemographics(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.LinkedInRequest) (interface{}, error) {
		return h.service.GetFollowersDemographics(ctx, req)
	})
}

func (h *LinkedInHandler) HandleAIInsights(w http.ResponseWriter, r *http.Request) {
	if h.aiInsights == nil {
		httputil.WriteStatusError(w, http.StatusServiceUnavailable, "AI insights service not configured")
		return
	}

	q := r.URL.Query()
	req := types.AIInsightsRequest{
		WorkspaceID: q.Get("workspace_id"),
		LinkedinID:  q.Get("linkedin_id"),
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

	if req.WorkspaceID == "" || req.LinkedinID == "" || req.Type == "" || req.Date == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "workspace_id, linkedin_id, date, and type are required")
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
