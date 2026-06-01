package tiktok

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/tiktok"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/tiktok"
)

type Handler struct {
	service    service.Service
	aiInsights *service.AIInsightsService
	logger     zerolog.Logger
}

func NewHandler(svc service.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With().Str("handler", "tiktok-analytics").Logger(),
	}
}

func (h *Handler) SetAIInsightsService(aiSvc *service.AIInsightsService) {
	h.aiInsights = aiSvc
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

func parseBaseRequest(r *http.Request) (*types.TiktokRequest, error) {
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

	req := &types.TiktokRequest{
		WorkspaceID: q.Get("workspace_id"),
		TiktokID:    q.Get("tiktok_id"),
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    q.Get("timezone"),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func parsePostsRequest(r *http.Request) (*types.PostsRequest, error) {
	baseReq, err := parseBaseRequest(r)
	if err != nil {
		return nil, err
	}
	req := &types.PostsRequest{TiktokRequest: *baseReq}
	q := r.URL.Query()
	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, httputil.NewValidationError("limit must be a valid integer")
		}
		req.Limit = limit
	}
	if offsetStr := q.Get("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return nil, httputil.NewValidationError("offset must be a valid integer")
		}
		req.Offset = offset
	}
	req.SortOrder = q.Get("sort_order")
	return req, nil
}

func (h *Handler) handleBase(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.TiktokRequest) (interface{}, error)) {
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

func (h *Handler) handlePosts(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.PostsRequest) (interface{}, error)) {
	req, err := parsePostsRequest(r)
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

func (h *Handler) HandlePageAndPostsInsights(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.TiktokRequest) (interface{}, error) {
		return h.service.GetPageAndPostsInsights(ctx, req)
	})
}

func (h *Handler) HandlePageFollowersAndViews(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.TiktokRequest) (interface{}, error) {
		return h.service.GetPageFollowersAndViews(ctx, req)
	})
}

func (h *Handler) HandlePostsAndEngagements(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.TiktokRequest) (interface{}, error) {
		return h.service.GetPostsAndEngagements(ctx, req)
	})
}

func (h *Handler) HandleDailyEngagementsData(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.TiktokRequest) (interface{}, error) {
		return h.service.GetDailyEngagementsData(ctx, req)
	})
}

func (h *Handler) HandleTopAndLeastPerformingPosts(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.TiktokRequest) (interface{}, error) {
		return h.service.GetTopAndLeastPerformingPosts(ctx, req)
	})
}

func (h *Handler) HandlePostsData(w http.ResponseWriter, r *http.Request) {
	h.handlePosts(w, r, func(ctx context.Context, req *types.PostsRequest) (interface{}, error) {
		return h.service.GetPostsData(ctx, req)
	})
}

func (h *Handler) HandleAIInsights(w http.ResponseWriter, r *http.Request) {
	if h.aiInsights == nil {
		httputil.WriteStatusError(w, http.StatusServiceUnavailable, "AI insights service not configured")
		return
	}

	q := r.URL.Query()
	req := types.AIInsightsRequest{
		WorkspaceID: q.Get("workspace_id"),
		TiktokID:    q.Get("tiktok_id"),
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
	if req.WorkspaceID == "" || req.TiktokID == "" || req.Type == "" || req.Date == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "workspace_id, tiktok_id, date, and type are required")
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
