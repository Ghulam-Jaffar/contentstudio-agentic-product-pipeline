package twitter

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/twitter"
	service "github.com/d4interactive/contentstudio-social-analytics-go/src/services/analytics/twitter"
)

type Handler struct {
	service service.Service
	logger  zerolog.Logger
}

func NewHandler(svc service.Service, logger zerolog.Logger) *Handler {
	return &Handler{
		service: svc,
		logger:  logger.With().Str("handler", "twitter-analytics").Logger(),
	}
}

func parseBaseRequest(r *http.Request) (*types.TwitterRequest, error) {
	q := r.URL.Query()
	startDate := q.Get("start_date")
	endDate := q.Get("end_date")

	if (startDate == "" || endDate == "") && q.Get("date") != "" {
		parts := strings.SplitN(strings.TrimSpace(q.Get("date")), " - ", 2)
		if len(parts) == 2 {
			if startDate == "" {
				startDate = strings.TrimSpace(parts[0])
			}
			if endDate == "" {
				endDate = strings.TrimSpace(parts[1])
			}
		}
	}

	req := &types.TwitterRequest{
		WorkspaceID: q.Get("workspace_id"),
		TwitterID:   q.Get("twitter_id"),
		StartDate:   startDate,
		EndDate:     endDate,
		Timezone:    q.Get("timezone"),
	}
	if err := req.Validate(); err != nil {
		return nil, err
	}
	return req, nil
}

func parseTweetsRequest(r *http.Request) (*types.TweetsRequest, error) {
	baseReq, err := parseBaseRequest(r)
	if err != nil {
		return nil, err
	}

	req := &types.TweetsRequest{TwitterRequest: *baseReq}
	q := r.URL.Query()
	if limitStr := q.Get("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return nil, httputil.NewValidationError("limit must be a valid integer")
		}
		req.Limit = limit
	}
	req.OrderBy = q.Get("order_by")

	return req, nil
}

func (h *Handler) handleBase(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.TwitterRequest) (interface{}, error)) {
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

func (h *Handler) handleTweets(w http.ResponseWriter, r *http.Request, fn func(ctx context.Context, req *types.TweetsRequest) (interface{}, error)) {
	req, err := parseTweetsRequest(r)
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
	h.handleBase(w, r, func(ctx context.Context, req *types.TwitterRequest) (interface{}, error) {
		return h.service.GetPageAndPostsInsights(ctx, req)
	})
}

func (h *Handler) HandleEngagementImpressionData(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.TwitterRequest) (interface{}, error) {
		return h.service.GetEngagementImpressionData(ctx, req)
	})
}

func (h *Handler) HandleFollowersTrendData(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.TwitterRequest) (interface{}, error) {
		return h.service.GetFollowersTrendData(ctx, req)
	})
}

func (h *Handler) HandleTopTweets(w http.ResponseWriter, r *http.Request) {
	h.handleTweets(w, r, func(ctx context.Context, req *types.TweetsRequest) (interface{}, error) {
		return h.service.GetTopTweets(ctx, req)
	})
}

func (h *Handler) HandleLeastTweets(w http.ResponseWriter, r *http.Request) {
	h.handleTweets(w, r, func(ctx context.Context, req *types.TweetsRequest) (interface{}, error) {
		return h.service.GetLeastTweets(ctx, req)
	})
}

func (h *Handler) HandleCreditsUsedCount(w http.ResponseWriter, r *http.Request) {
	h.handleBase(w, r, func(ctx context.Context, req *types.TwitterRequest) (interface{}, error) {
		return h.service.GetCreditsUsedCount(ctx, req)
	})
}
