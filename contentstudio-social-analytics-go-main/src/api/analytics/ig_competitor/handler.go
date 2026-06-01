// Package ig_competitor provides HTTP handlers for Instagram competitor analytics.
// Migrated from PHP InstagramCompetitorController (contentstudio-backend).
package ig_competitor

import (
	"context"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/ig_competitor"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/ig_competitor"
)

// Handler handles HTTP requests for Instagram competitor analytics.
type Handler struct {
	service service.Service
	logger  zerolog.Logger
}

// NewHandler creates a new handler.
func NewHandler(svc service.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With().Str("handler", "ig-competitor").Logger(),
	}
}

// parseRequest parses query parameters into a CompetitorRequest.
func parseRequest(r *http.Request) (*types.CompetitorRequest, error) {
	q := r.URL.Query()
	req := &types.CompetitorRequest{
		ReportID:          q.Get("_id"),
		StartDate:         q.Get("start_date"),
		EndDate:           q.Get("end_date"),
		Timezone:          q.Get("timezone"),
		SortOrder:         q.Get("sort_order"),
		BusinessAccountID: q.Get("business_account_id"),
		MediaType:         q.Get("media_type"),
		MediaProductType:  q.Get("media_product_type"),
		Hashtag:           q.Get("hashtag"),
	}
	if limit := q.Get("limit"); limit != "" {
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

// handle is a shared handler pattern for analytics GET endpoints.
func (h *Handler) handle(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)) {
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

// HandleDataTableMetrics handles GET /analytics/overview/instagram/competitor/dataTableMetrics
func (h *Handler) HandleDataTableMetrics(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
		return h.service.GetDataTableMetrics(ctx, req)
	})
}

// HandlePostingActivityGraphByTypes handles GET /analytics/overview/instagram/competitor/postingActivityGraphByTypes
func (h *Handler) HandlePostingActivityGraphByTypes(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
		return h.service.GetPostingActivityGraphByTypes(ctx, req)
	})
}

// HandlePostingActivityBySpecificType handles GET /analytics/overview/instagram/competitor/postingActivityBySpecificType
func (h *Handler) HandlePostingActivityBySpecificType(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
		return h.service.GetPostingActivityBySpecificType(ctx, req)
	})
}

// HandlePostingActivityTableByType handles GET /analytics/overview/instagram/competitor/postingActivityTableByType
func (h *Handler) HandlePostingActivityTableByType(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
		return h.service.GetPostingActivityTableByType(ctx, req)
	})
}

// HandleFollowersGrowthComparison handles GET /analytics/overview/instagram/competitor/followersGrowthComparison
func (h *Handler) HandleFollowersGrowthComparison(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
		return h.service.GetFollowersGrowthComparison(ctx, req)
	})
}

// HandleTopAndLeastPerformingPosts handles GET /analytics/overview/instagram/competitor/topAndLeastPerformingPosts
func (h *Handler) HandleTopAndLeastPerformingPosts(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
		return h.service.GetTopAndLeastPerformingPosts(ctx, req)
	})
}

// HandleTopHashtags handles GET /analytics/overview/instagram/competitor/topHashtags
func (h *Handler) HandleTopHashtags(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
		return h.service.GetTopHashtags(ctx, req)
	})
}

// HandleIndividualHashtagData handles GET /analytics/overview/instagram/competitor/individualHashtagData
func (h *Handler) HandleIndividualHashtagData(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
		return h.service.GetIndividualHashtagData(ctx, req)
	})
}

// HandleBiographyData handles GET /analytics/overview/instagram/competitor/biographyData
func (h *Handler) HandleBiographyData(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
		return h.service.GetBiographyData(ctx, req)
	})
}
