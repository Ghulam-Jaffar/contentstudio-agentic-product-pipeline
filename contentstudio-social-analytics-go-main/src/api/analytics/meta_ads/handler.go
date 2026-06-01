package meta_ads

import (
	"context"
	"net/http"
	"strconv"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/meta_ads"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/meta_ads"
)

// Handler handles HTTP requests for Meta Ads analytics endpoints.
type Handler struct {
	service service.Service
	logger  zerolog.Logger
}

// NewHandler returns a new Handler.
func NewHandler(svc service.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With().Str("handler", "meta-ads-analytics").Logger(),
	}
}

// parseRequest reads all supported query parameters from the request.
func parseRequest(r *http.Request) (*types.MetaAdsRequest, error) {
	q := r.URL.Query()
	req := &types.MetaAdsRequest{
		WorkspaceID: q.Get("workspace_id"),
		AccountID:   q.Get("account_id"),
		StartDate:   q.Get("start_date"),
		EndDate:     q.Get("end_date"),
		Timezone:    q.Get("timezone"),
		Language:    q.Get("language"),
		Status:      q.Get("status"),
		Objective:   q.Get("objective"),
		Search:      q.Get("search"),
		OrderBy:     q.Get("order_by"),
		OrderDir:    q.Get("order_dir"),
		Metric:      q.Get("metric"),
		Breakdown:   q.Get("breakdown"),
		Country:     q.Get("country"),
		Level:       q.Get("level"),
		SortBy:      q.Get("sort_by"),
	}

	if pageStr := q.Get("page"); pageStr != "" {
		n, err := strconv.Atoi(pageStr)
		if err != nil {
			return nil, httputil.NewValidationError("page must be a valid integer")
		}
		req.Page = n
	}
	if perPageStr := q.Get("per_page"); perPageStr != "" {
		n, err := strconv.Atoi(perPageStr)
		if err != nil {
			return nil, httputil.NewValidationError("per_page must be a valid integer")
		}
		req.PerPage = n
	}

	if err := req.Validate(); err != nil {
		return nil, err
	}
	if req.Language == "" {
		req.Language = r.Header.Get("X-LOCALE")
	}
	return req, nil
}

// handle is the common request/response wrapper.
func (h *Handler) handle(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error)) {
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

// HandleSummary returns the 7 overview metric cards.
func (h *Handler) HandleSummary(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetSummary(ctx, req)
	})
}

// HandleResultsByObjective returns aggregated results grouped by objective.
func (h *Handler) HandleResultsByObjective(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetResultsByObjective(ctx, req)
	})
}

// HandleImpressionsVsSpend returns daily impressions and spend time-series data.
func (h *Handler) HandleImpressionsVsSpend(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetImpressionsVsSpend(ctx, req)
	})
}

// HandleClicksVsCTR returns daily clicks and CTR time-series data.
func (h *Handler) HandleClicksVsCTR(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetClicksVsCTR(ctx, req)
	})
}

// HandleTopCampaigns returns the top 5 campaigns by spend.
func (h *Handler) HandleTopCampaigns(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetTopCampaigns(ctx, req)
	})
}

// HandlePerformanceTrend returns the performance trend chart data for a given metric.
func (h *Handler) HandlePerformanceTrend(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetPerformanceTrend(ctx, req)
	})
}

// HandlePerformanceByLevel returns performance breakdown by campaign, adset, or ad.
func (h *Handler) HandlePerformanceByLevel(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetPerformanceByLevel(ctx, req)
	})
}

// HandlePerformanceByPlatform returns performance breakdown by publisher platform.
func (h *Handler) HandlePerformanceByPlatform(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetPerformanceByPlatform(ctx, req)
	})
}

// HandleCampaignsList returns a paginated, filterable, sortable list of campaigns.
func (h *Handler) HandleCampaignsList(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetCampaignsList(ctx, req)
	})
}

// HandleAdSetsList returns a paginated, filterable, sortable list of ad sets.
func (h *Handler) HandleAdSetsList(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetAdSetsList(ctx, req)
	})
}

// HandleAdsList returns a paginated, filterable, sortable list of ads.
func (h *Handler) HandleAdsList(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetAdsList(ctx, req)
	})
}

// HandleDemographicsAgeGender returns demographic data broken down by age and gender.
func (h *Handler) HandleDemographicsAgeGender(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetDemographicsAgeGender(ctx, req)
	})
}

// HandleDemographicsRegionCountry returns demographic data broken down by region or country.
func (h *Handler) HandleDemographicsRegionCountry(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetDemographicsRegionCountry(ctx, req)
	})
}

// HandleAIInsightsSummary returns AI-powered insights summary for Meta Ads.
func (h *Handler) HandleAIInsightsSummary(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetAIInsightsSummary(ctx, req)
	})
}

// HandleAIInsightsDetailed returns detailed AI-powered insights for Meta Ads.
func (h *Handler) HandleAIInsightsDetailed(w http.ResponseWriter, r *http.Request) {
	h.handle(w, r, func(ctx context.Context, req *types.MetaAdsRequest) (interface{}, error) {
		return h.service.GetAIInsightsDetailed(ctx, req)
	})
}
