package listening

import (
	"net/http"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	mentionsSvc "github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/mentions"
)

type AnalyticsHandler struct {
	service        *mentionsSvc.Service
	filterResolver *MentionFilterResolver
	logger         zerolog.Logger
}

func NewAnalyticsHandler(
	svc *mentionsSvc.Service,
	filterResolver *MentionFilterResolver,
	logger zerolog.Logger,
) *AnalyticsHandler {
	return &AnalyticsHandler{
		service:        svc,
		filterResolver: filterResolver,
		logger:         logger.With().Str("handler", "listening-analytics").Logger(),
	}
}

func (h *AnalyticsHandler) HandleGetAnalytics(w http.ResponseWriter, r *http.Request) {
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

	// Extract topic names from query for mapping
	topicNames := make(map[string]string)
	q := r.URL.Query()
	names := q["topic_names[]"]
	ids := q["topic_ids[]"]

	for i, id := range ids {
		if i < len(names) {
			topicNames[id] = names[i]
		}
	}

	resp, err := h.service.GetAnalytics(r.Context(), filter, topicNames)
	if err != nil {
		httputil.WriteError(w, h.logger, err)
		return
	}

	httputil.WriteJSON(w, http.StatusOK, resp)
}

func (h *AnalyticsHandler) HandleExportMentions(w http.ResponseWriter, r *http.Request) {
	format := r.URL.Query().Get("format")
	if format == "" {
		format = "csv"
	}

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

	// Extract topic names for export
	topicNames := make(map[string]string)
	q := r.URL.Query()
	names := q["topic_names[]"]
	ids := q["topic_ids[]"]
	for i, id := range ids {
		if i < len(names) {
			topicNames[id] = names[i]
		}
	}

	if format == "pdf" {
		w.Header().Set("Content-Type", "application/pdf")
		w.Header().Set("Content-Disposition", "attachment; filename=social_mentions_report.pdf")
		if err := h.service.ExportMentionsPDF(r.Context(), filter, topicNames, w); err != nil {
			h.logger.Error().Err(err).Msg("PDF Export failed")
			// Cannot write error response after headers/body started for complex exports
		}
		return
	}

	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=social_mentions_export.csv")

	if err := h.service.ExportMentionsCSV(r.Context(), filter, topicNames, w); err != nil {
		h.logger.Error().Err(err).Msg("Export failed")
	}
}
