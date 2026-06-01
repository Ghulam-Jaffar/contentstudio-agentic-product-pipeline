// Package instagram provides HTTP handlers for Instagram analytics API endpoints.
// Each handler parses query parameters, delegates to the service layer, and writes JSON responses.
//
// This is the HTTP layer of the 3-layer architecture:
//
//	Handler (this package) → Service (services/analytics/instagram) → Repository (db/clickhouse/analytics-get-queries/instagram)
//
// Migrated from PHP: InstagramAnalyticsController (contentstudio-backend).
package instagram

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/instagram"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/instagram"
)

// InstagramHandler handles HTTP requests for Instagram analytics endpoints.
type InstagramHandler struct {
	service    service.Service
	aiInsights *service.AIInsightsService
	logger     zerolog.Logger
}

// NewInstagramHandler creates a new handler with the given service and logger.
func NewInstagramHandler(svc service.Service, logger zerolog.Logger) *InstagramHandler {
	return &InstagramHandler{
		service: svc,
		logger:  logger.With().Str("handler", "instagram-analytics").Logger(),
	}
}

// SetAIInsightsService attaches the AI insights service to this handler.
func (h *InstagramHandler) SetAIInsightsService(aiSvc *service.AIInsightsService) {
	h.aiInsights = aiSvc
}

// parseBaseRequest extracts and validates common query parameters.
func parseBaseRequest(r *http.Request) (*types.InstagramRequest, error) {
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
	req := &types.InstagramRequest{
		WorkspaceID: q.Get("workspace_id"),
		InstagramID: q.Get("instagram_id"),
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

// handleBase is a helper that parses the request, calls the service, and writes the response.
func (h *InstagramHandler) handleBase(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.InstagramRequest) (interface{}, error)) {
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

// HandleSummary handles GET /analytics/overview/instagram/summary
func (h *InstagramHandler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetSummary(ctx, req)
	})
}

// HandleAudienceGrowth handles GET /analytics/overview/instagram/audienceGrowth
func (h *InstagramHandler) HandleAudienceGrowth(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetAudienceGrowth(ctx, req)
	})
}

// HandlePublishingBehaviour handles GET /analytics/overview/instagram/publishingBehaviour
func (h *InstagramHandler) HandlePublishingBehaviour(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	pubReq := &types.PublishingBehaviourRequest{InstagramRequest: *req}
	if mediaType := r.URL.Query().Get("media_type"); mediaType != "" {
		pubReq.MediaType = strings.Split(mediaType, ",")
	}
	resp, err := h.service.GetPublishingBehaviour(r.Context(), pubReq)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandleTopPosts handles GET /analytics/overview/instagram/topPosts (default limit: 5)
func (h *InstagramHandler) HandleTopPosts(w http.ResponseWriter, r *http.Request) {
	h.handleTopPostsWithDefault(w, r, 5)
}

// HandleGetTopPosts handles GET /analytics/overview/instagram/getTopPosts (default limit: 15)
func (h *InstagramHandler) HandleGetTopPosts(w http.ResponseWriter, r *http.Request) {
	h.handleTopPostsWithDefault(w, r, 15)
}

// handleTopPostsWithDefault is shared by topPosts and getTopPosts with a configurable default limit.
func (h *InstagramHandler) handleTopPostsWithDefault(w http.ResponseWriter, r *http.Request, defaultLimit int) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	topReq := &types.TopPostsRequest{InstagramRequest: *req}
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

// HandleActiveUsers handles GET /analytics/overview/instagram/activeUsers
func (h *InstagramHandler) HandleActiveUsers(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetActiveUsers(ctx, req)
	})
}

// HandleImpressions handles GET /analytics/overview/instagram/impressions
func (h *InstagramHandler) HandleImpressions(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetImpressions(ctx, req)
	})
}

// HandleEngagement handles GET /analytics/overview/instagram/engagement
func (h *InstagramHandler) HandleEngagement(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetEngagement(ctx, req)
	})
}

// HandleHashtags handles GET /analytics/overview/instagram/hashtags
func (h *InstagramHandler) HandleHashtags(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetHashtags(ctx, req)
	})
}

// HandleStoriesPerformance handles GET /analytics/overview/instagram/storiesPerformance
func (h *InstagramHandler) HandleStoriesPerformance(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetStoriesPerformance(ctx, req)
	})
}

// HandleReelsPerformance handles GET /analytics/overview/instagram/reelsPerformance
func (h *InstagramHandler) HandleReelsPerformance(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetReelsPerformance(ctx, req)
	})
}

// HandleDemographicsAge handles GET /analytics/overview/instagram/demographicsAge
func (h *InstagramHandler) HandleDemographicsAge(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetDemographicsAge(ctx, req)
	})
}

// HandleCountryCity handles GET /analytics/overview/instagram/countryCity
func (h *InstagramHandler) HandleCountryCity(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.InstagramRequest) (interface{}, error) {
		return h.service.GetCountryCity(ctx, req)
	})
}

// HandleAIInsights handles GET /analytics/overview/instagram/ai_insights
func (h *InstagramHandler) HandleAIInsights(w http.ResponseWriter, r *http.Request) {
	if h.aiInsights == nil {
		httputil.WriteStatusError(w, http.StatusServiceUnavailable, "AI insights service not configured")
		return
	}

	q := r.URL.Query()
	req := types.AIInsightsRequest{
		WorkspaceID: q.Get("workspace_id"),
		InstagramID: q.Get("instagram_id"),
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

	if req.WorkspaceID == "" || req.InstagramID == "" || req.Type == "" || req.Date == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "workspace_id, instagram_id, date, and type are required")
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
