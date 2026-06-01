package gmb

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/gmb"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/gmb"
)

type GMBHandler struct {
	service    service.Service
	aiInsights *service.AIInsightsService
	logger     zerolog.Logger
}

func NewGMBHandler(svc service.Service, logger zerolog.Logger) *GMBHandler {
	return &GMBHandler{
		service: svc,
		logger:  logger.With().Str("handler", "gmb-analytics").Logger(),
	}
}

func (h *GMBHandler) SetAIInsightsService(aiSvc *service.AIInsightsService) {
	h.aiInsights = aiSvc
}

func parseBaseRequest(r *http.Request) (*types.GMBRequest, error) {
	q := r.URL.Query()
	req := &types.GMBRequest{
		WorkspaceID: q.Get("workspace_id"),
		GmbID:       q.Get("gmb_id"),
		StartDate:   q.Get("start_date"),
		EndDate:     q.Get("end_date"),
		Timezone:    q.Get("timezone"),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (h *GMBHandler) handleBase(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.GMBRequest) (interface{}, error)) {
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

func (h *GMBHandler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.GMBRequest) (interface{}, error) {
		return h.service.GetSummary(ctx, req)
	})
}

func (h *GMBHandler) HandleImpressions(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.GMBRequest) (interface{}, error) {
		return h.service.GetImpressions(ctx, req)
	})
}

func (h *GMBHandler) HandleActions(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.GMBRequest) (interface{}, error) {
		return h.service.GetActions(ctx, req)
	})
}

func (h *GMBHandler) HandleSearchKeywords(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	kwReq := &types.SearchKeywordsRequest{GMBRequest: *req}
	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			httputil.WriteError(w, h.logger, httputil.NewValidationError("limit must be a valid integer"))
			return
		}
		kwReq.Limit = limit
	}
	resp, err := h.service.GetSearchKeywords(r.Context(), kwReq)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *GMBHandler) HandleTopPosts(w http.ResponseWriter, r *http.Request) {
	req, err := parseBaseRequest(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	topReq := &types.TopPostsRequest{GMBRequest: *req}
	q := r.URL.Query()
	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			httputil.WriteError(w, h.logger, httputil.NewValidationError("limit must be a valid integer"))
			return
		}
		topReq.Limit = limit
	}
	topReq.OrderBy = q.Get("order_by")
	resp, err := h.service.GetTopPosts(r.Context(), topReq)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *GMBHandler) HandlePublishingBehavior(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.GMBRequest) (interface{}, error) {
		return h.service.GetPublishingBehavior(ctx, req)
	})
}

func (h *GMBHandler) HandleReviews(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.GMBRequest) (interface{}, error) {
		return h.service.GetReviews(ctx, req)
	})
}

func (h *GMBHandler) HandleMediaActivity(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.GMBRequest) (interface{}, error) {
		return h.service.GetMediaActivity(ctx, req)
	})
}

func (h *GMBHandler) HandleAIInsights(w http.ResponseWriter, r *http.Request) {
	if h.aiInsights == nil {
		httputil.WriteStatusError(w, http.StatusServiceUnavailable, "AI insights service not configured")
		return
	}

	q := r.URL.Query()
	req := types.AIInsightsRequest{
		WorkspaceID: q.Get("workspace_id"),
		GmbID:       q.Get("gmb_id"),
		StartDate:   q.Get("start_date"),
		EndDate:     q.Get("end_date"),
		Date:        q.Get("date"),
		Timezone:    q.Get("timezone"),
		Type:        q.Get("type"),
		Language:    q.Get("language"),
	}
	if req.Date == "" && req.StartDate != "" && req.EndDate != "" {
		req.Date = strings.TrimSpace(req.StartDate) + " - " + strings.TrimSpace(req.EndDate)
	}
	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			httputil.WriteStatusError(w, http.StatusBadRequest, "limit must be a valid integer")
			return
		}
		req.Limit = limit
	}

	if req.WorkspaceID == "" || req.GmbID == "" || req.Type == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "workspace_id, gmb_id, and type are required")
		return
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
