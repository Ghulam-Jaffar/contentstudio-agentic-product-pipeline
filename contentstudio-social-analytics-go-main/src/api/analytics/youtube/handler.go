// Package youtube provides HTTP handlers for YouTube analytics API endpoints.
// Each handler parses query parameters, delegates to the service layer, and writes JSON responses.
//
// This is the HTTP layer of the 3-layer architecture:
//
//	Handler (this package) → Service (services/analytics/youtube) → Repository (db/clickhouse/analytics-get-queries/youtube)
//
// Migrated from PHP: YouTubeAnalyticsController (contentstudio-backend).
package youtube

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/youtube"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/youtube"
)

// Handler handles HTTP requests for YouTube analytics endpoints.
type Handler struct {
	service    service.Service
	aiInsights *service.AIInsightsService
	logger     zerolog.Logger
}

// NewHandler creates a new handler with the given service and logger.
func NewHandler(svc service.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With().Str("handler", "youtube-analytics").Logger(),
	}
}

// SetAIInsightsService attaches the AI insights service to this handler.
func (h *Handler) SetAIInsightsService(aiSvc *service.AIInsightsService) {
	h.aiInsights = aiSvc
}

// parseBaseRequest extracts and validates common query parameters.
func parseBaseRequest(r *http.Request) (*types.YoutubeRequest, error) {
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
	req := &types.YoutubeRequest{
		WorkspaceID: q.Get("workspace_id"),
		YoutubeID:   q.Get("youtube_id"),
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    q.Get("timezone"),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

// parseDateRangeQuery splits a "YYYY-MM-DD - YYYY-MM-DD" query param into start and end strings.
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
func (h *Handler) handleBase(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error)) {
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

// HandleSummary handles GET /analytics/overview/youtube/summary
func (h *Handler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetSummary(ctx, req)
	})
}

// HandleSubscriberTrend handles GET /analytics/overview/youtube/subscriberTrend
func (h *Handler) HandleSubscriberTrend(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetSubscriberTrend(ctx, req)
	})
}

// HandleDynamicSubscriberTrend handles GET /analytics/overview/youtube/dynamicSubscriberTrend
func (h *Handler) HandleDynamicSubscriberTrend(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetDynamicSubscriberTrend(ctx, req)
	})
}

// HandleEngagementTrend handles GET /analytics/overview/youtube/engagementTrend
func (h *Handler) HandleEngagementTrend(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetEngagementTrend(ctx, req)
	})
}

// HandleDynamicEngagementTrend handles GET /analytics/overview/youtube/dynamicEngagementTrend
func (h *Handler) HandleDynamicEngagementTrend(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetDynamicEngagementTrend(ctx, req)
	})
}

// HandleViewsTrend handles GET /analytics/overview/youtube/viewsTrend
func (h *Handler) HandleViewsTrend(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetViewsTrend(ctx, req)
	})
}

// HandleDynamicViewsTrend handles GET /analytics/overview/youtube/dynamicViewsTrend
func (h *Handler) HandleDynamicViewsTrend(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetDynamicViewsTrend(ctx, req)
	})
}

// HandleWatchTimeTrend handles GET /analytics/overview/youtube/watchTimeTrend
func (h *Handler) HandleWatchTimeTrend(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetWatchTimeTrend(ctx, req)
	})
}

// HandleDynamicWatchTimeTrend handles GET /analytics/overview/youtube/dynamicWatchTimeTrend
func (h *Handler) HandleDynamicWatchTimeTrend(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetDynamicWatchTimeTrend(ctx, req)
	})
}

// HandleFindVideo handles GET /analytics/overview/youtube/findVideo
func (h *Handler) HandleFindVideo(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetFindVideo(ctx, req)
	})
}

// HandleVideoSharing handles GET /analytics/overview/youtube/videoSharing
func (h *Handler) HandleVideoSharing(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetVideoSharing(ctx, req)
	})
}

// HandleTopPosts handles GET /analytics/overview/youtube/topPosts
func (h *Handler) HandleTopPosts(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetTopVideos(ctx, req)
	})
}

// HandleLeastPosts handles GET /analytics/overview/youtube/leastPosts
func (h *Handler) HandleLeastPosts(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetLeastVideos(ctx, req)
	})
}

// HandleGetTopPosts handles GET /analytics/overview/youtube/getTopPosts
// Accepts optional order_by and limit query parameters.
func (h *Handler) HandleGetTopPosts(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	topReq := &types.TopVideosRequest{YoutubeRequest: *req}
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
		topReq.Limit = 15
	}
	topReq.OrderBy = q.Get("order_by")

	resp, err := h.service.GetSortedTopVideos(r.Context(), topReq)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

// HandlePerformanceAndSchedule handles GET /analytics/overview/youtube/performanceAndSchedule
func (h *Handler) HandlePerformanceAndSchedule(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.YoutubeRequest) (interface{}, error) {
		return h.service.GetPerformanceAndSchedule(ctx, req)
	})
}

// HandleAIInsights handles GET /analytics/overview/youtube/ai_insights
func (h *Handler) HandleAIInsights(w http.ResponseWriter, r *http.Request) {
	if h.aiInsights == nil {
		httputil.WriteStatusError(w, http.StatusServiceUnavailable, "AI insights service not configured")
		return
	}

	q := r.URL.Query()
	req := types.AIInsightsRequest{
		WorkspaceID: q.Get("workspace_id"),
		YoutubeID:   q.Get("youtube_id"),
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

	if req.WorkspaceID == "" || req.YoutubeID == "" || req.Type == "" || req.Date == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "workspace_id, youtube_id, date, and type are required")
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
