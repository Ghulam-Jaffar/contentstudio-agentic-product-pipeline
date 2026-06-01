package facebook

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/facebook"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/facebook"
)

type Handler struct {
	service    service.Service
	aiInsights *service.AIInsightsService
	logger     zerolog.Logger
}

func NewHandler(svc service.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With().Str("handler", "facebook-analytics").Logger(),
	}
}

func (h *Handler) SetAIInsightsService(aiSvc *service.AIInsightsService) {
	h.aiInsights = aiSvc
}

func parseRequest(r *http.Request) (*types.FacebookRequest, error) {
	req := &types.FacebookRequest{
		WorkspaceID: r.URL.Query().Get("workspace_id"),
		FacebookIDs: append([]string{}, r.URL.Query()["facebook_id"]...),
		Date:        r.URL.Query().Get("date"),
		StartDate:   r.URL.Query().Get("start_date"),
		EndDate:     r.URL.Query().Get("end_date"),
		Timezone:    r.URL.Query().Get("timezone"),
		MediaType:   append([]string{}, r.URL.Query()["media_type"]...),
		OrderBy:     r.URL.Query().Get("order_by"),
	}
	if single := r.URL.Query().Get("facebook_id"); single != "" && len(req.FacebookIDs) == 0 {
		req.FacebookIDs = []string{single}
	}
	if media := r.URL.Query().Get("media_type"); media != "" && len(req.MediaType) == 0 {
		req.MediaType = []string{media}
	}
	if limit := r.URL.Query().Get("limit"); limit != "" {
		n, err := strconv.Atoi(limit)
		if err != nil {
			return nil, httputil.NewValidationError("limit must be a valid integer")
		}
		req.Limit = n
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func (h *Handler) handle(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.FacebookRequest) (interface{}, error)) {
	req, err := parseRequest(r)
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

func (h *Handler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetSummary(ctx, req)
	})
}

func (h *Handler) HandleAudienceGrowth(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetAudienceGrowth(ctx, req)
	})
}

func (h *Handler) HandlePublishingBehaviour(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetPublishingBehaviour(ctx, req)
	})
}

func (h *Handler) HandleOverviewTopPosts(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		if req.Limit <= 0 {
			req.Limit = 15
		}
		return h.service.GetTopPosts(ctx, req)
	})
}

func (h *Handler) HandleGetTopPosts(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		if req.Limit <= 0 {
			req.Limit = 15
		}
		return h.service.GetTopPosts(ctx, req)
	})
}

func (h *Handler) HandleActiveUsers(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetActiveUsers(ctx, req)
	})
}

func (h *Handler) HandleImpressions(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetImpressions(ctx, req)
	})
}

func (h *Handler) HandleEngagement(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetEngagement(ctx, req)
	})
}

func (h *Handler) HandleReelsAnalytics(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetReelsAnalytics(ctx, req)
	})
}

func (h *Handler) HandleVideoInsights(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetVideoInsights(ctx, req)
	})
}

func (h *Handler) HandleDemographics(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetDemographics(ctx, req)
	})
}

func (h *Handler) HandleOverviewDemographics(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetOverviewDemographics(ctx, req)
	})
}

func (h *Handler) HandleAudienceLocation(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.FacebookRequest) (interface{}, error) {
		return h.service.GetAudienceLocation(ctx, req)
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
		FacebookID:  q.Get("facebook_id"),
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
	if limit := q.Get("limit"); limit != "" {
		n, err := strconv.Atoi(limit)
		if err != nil {
			httputil.WriteStatusError(w, http.StatusBadRequest, "limit must be a valid integer")
			return
		}
		req.Limit = n
	}

	if req.WorkspaceID == "" || req.FacebookID == "" || req.Type == "" || req.Date == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "workspace_id, facebook_id, date, and type are required")
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
