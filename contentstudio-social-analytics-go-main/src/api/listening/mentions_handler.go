package listening

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	mentionsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/mentions"
)

type MentionsHandler struct {
	service        *mentionsSvc.Service
	filterResolver *MentionFilterResolver
	logger         zerolog.Logger
}

func NewMentionsHandler(
	svc *mentionsSvc.Service,
	filterResolver *MentionFilterResolver,
	logger zerolog.Logger,
) *MentionsHandler {
	return &MentionsHandler{
		service:        svc,
		filterResolver: filterResolver,
		logger:         logger.With().Str("handler", "listening-mentions").Logger(),
	}
}

func (h *MentionsHandler) HandleListMentions(w http.ResponseWriter, r *http.Request) {
	filter, err := parseMentionFilter(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	filter, err = h.filterResolver.Resolve(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	resp, err := h.service.ListMentions(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *MentionsHandler) HandlePatchMention(w http.ResponseWriter, r *http.Request) {
	mentionID := r.PathValue("id")
	if mentionID == "" {
		httputil.WriteStatusError(w, http.StatusBadRequest, "mention id is required")
		return
	}

	var patch apiModels.MentionPatchRequest
	if err := json.NewDecoder(r.Body).Decode(&patch); err != nil {
		httputil.WriteStatusError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if patch.SentimentOverride != "" {
		valid := map[string]bool{"positive": true, "neutral": true, "negative": true}
		if !valid[patch.SentimentOverride] {
			httputil.WriteStatusError(w, http.StatusBadRequest, "sentiment_override must be positive, neutral, or negative")
			return
		}
	}

	if err := h.service.PatchMention(r.Context(), mentionID, &patch); err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":  true,
		"message": "mention updated",
	})
}

func (h *MentionsHandler) HandleMarkAllRead(w http.ResponseWriter, r *http.Request) {
	var req apiModels.MarkAllReadRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteStatusError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	filter := &apiModels.MentionFilter{
		TopicIDs:   req.TopicIDs,
		Platforms:  req.Platforms,
		Sentiments: req.Sentiments,
	}

	count, err := h.service.MarkAllRead(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status":       true,
		"message":      "mentions marked as read",
		"marked_count": count,
	})
}

func (h *MentionsHandler) HandleUnreadCount(w http.ResponseWriter, r *http.Request) {
	filter, err := parseMentionFilter(r)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	filter, err = h.filterResolver.Resolve(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	count, err := h.service.GetUnreadCount(r.Context(), filter)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, apiModels.UnreadCountResponse{
		Status:      true,
		UnreadCount: count,
	})
}

func parseMentionFilter(r *http.Request) (*apiModels.MentionFilter, error) {
	q := r.URL.Query()

	filter := &apiModels.MentionFilter{
		WorkspaceID:       q.Get("workspace_id"),
		TopicIDs:          queryArray(q, "topic_ids[]"),
		Platforms:         queryArray(q, "platforms[]"),
		Sentiments:        queryArray(q, "sentiments[]"),
		AITags:            queryArray(q, "ai_tags[]"),
		ExcludeAITags:     queryArray(q, "exclude_ai_tags[]"),
		Language:          queryArray(q, "language[]"),
		DateFrom:          q.Get("date_from"),
		DateTo:            q.Get("date_to"),
		Sort:              q.Get("sort"),
		Cursor:            q.Get("cursor"),
		ViewID:            q.Get("view_id"),
		IncludeIrrelevant: q.Get("include_irrelevant") == "true" || q.Get("include_irrelevant") == "1",
		Search:            q.Get("search"),
	}

	if limit := q.Get("limit"); limit != "" {
		n, err := strconv.Atoi(limit)
		if err != nil {
			return nil, httputil.NewValidationError("limit must be a valid integer")
		}
		filter.Limit = n
	}

	if minFollowers := q.Get("min_followers"); minFollowers != "" {
		n, err := strconv.Atoi(minFollowers)
		if err != nil {
			return nil, httputil.NewValidationError("min_followers must be a valid integer")
		}
		filter.MinFollowers = n
	}

	if minTotalEngagement := q.Get("min_total_engagement"); minTotalEngagement != "" {
		n, err := strconv.Atoi(minTotalEngagement)
		if err != nil {
			return nil, httputil.NewValidationError("min_total_engagement must be a valid integer")
		}
		filter.MinTotalEngagement = n
	}

	if isBookmarked := q.Get("is_bookmarked"); isBookmarked != "" {
		v := isBookmarked == "true"
		filter.IsBookmarked = &v
	}

	if isRead := q.Get("is_read"); isRead != "" {
		v := isRead == "true"
		filter.IsRead = &v
	}
	return filter, nil
}

func queryArray(q map[string][]string, key string) []string {
	values := q[key]
	if len(values) == 0 {
		return nil
	}
	var result []string
	for _, v := range values {
		for _, part := range strings.Split(v, ",") {
			trimmed := strings.TrimSpace(part)
			if trimmed != "" {
				result = append(result, trimmed)
			}
		}
	}
	return result
}
