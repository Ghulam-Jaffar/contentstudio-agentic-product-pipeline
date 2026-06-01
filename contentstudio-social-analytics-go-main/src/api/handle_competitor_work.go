package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// HandleCompetitorWork handles incoming HTTP requests for competitor work orders
func (s *APIServer) HandleCompetitorWork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendErrorResponse(w, http.StatusMethodNotAllowed, ErrCodeInvalidRequest, "Method not allowed", "Only POST method is supported")
		return
	}

	var req apiModels.CompetitorWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to decode competitor work request body")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}

	if req.PageID == "" {
		s.Logger.Error().Msg("Missing page_id in competitor work request")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeMissingField, "Missing required field", "page_id is required")
		return
	}

	if req.Channel == "" {
		s.Logger.Error().Msg("Missing channel in competitor work request")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeMissingField, "Missing required field", "channel is required")
		return
	}

	s.Logger.Info().
		Str("report_id", req.ReportID).
		Str("page_id", req.PageID).
		Str("channel", req.Channel).
		Msg("Received competitor work request")

	if err := s.processCompetitorWork(r.Context(), req); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to process competitor work request")
		s.sendErrorResponse(w, http.StatusInternalServerError, ErrCodeInternalError, "Failed to process request", err.Error())
		return
	}

	s.sendSuccessResponse(w, map[string]interface{}{
		"status":    "success",
		"message":   "Competitor work order dispatched successfully",
		"report_id": req.ReportID,
		"page_id":   req.PageID,
		"channel":   req.Channel,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// processCompetitorWork creates and sends a competitor work order to Kafka
func (s *APIServer) processCompetitorWork(ctx context.Context, req apiModels.CompetitorWorkRequest) error {
	// Build Kafka work order - use new field names
	workOrder := kafkaModels.CompetitorWorkOrder{
		ReportID:  req.ReportID,
		PageID:    req.PageID,
		Channel:   req.Channel,
		Mode:      "full_refresh",
		StartDate: req.StartDate,
		EndDate:   req.EndDate,
	}

	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.processCompetitorWork: marshal failed: %w", err)
	}

	// Kafka topic
	topic := fmt.Sprintf("competitor-work-order-%s", req.Channel)

	if err := s.Producer.Produce(ctx, topic, []byte(req.PageID), data); err != nil {
		return fmt.Errorf("APIServer.processCompetitorWork: Kafka send failed: %w", err)
	}

	s.Logger.Info().
		Str("report_id", req.ReportID).
		Str("page_id", req.PageID).
		Str("topic", topic).
		Msg("Competitor work dispatched")

	return nil
}
