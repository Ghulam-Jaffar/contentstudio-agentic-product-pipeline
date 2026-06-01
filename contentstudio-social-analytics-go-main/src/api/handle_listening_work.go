package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"time"

	apiModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/api"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkaModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

// ListeningTopicGetter abstracts the listening repository for testability.
type ListeningTopicGetter interface {
	GetTopicByID(ctx context.Context, topicID string) (*mongoModels.ListeningTopic, error)
}

// ListeningWorkspaceQuotaChecker abstracts workspace-level quota reads for
// pre-flight checks before dispatching work orders.
type ListeningWorkspaceQuotaChecker interface {
	IsWorkspaceMentionLimitReached(ctx context.Context, ID string) (bool, error)
	GetWorkspaceUsage(ctx context.Context, ID string) (mentionsCount, mentionLimit int, exists bool, err error)
}

const (
	initialCrawlWindow           = 30 * 24 * time.Hour // 30 days
	ErrCodeWorkspaceLimitReached = "WORKSPACE_LIMIT_REACHED"
	ErrCodeTopicInactive         = "TOPIC_INACTIVE"
	TopicStatusActive            = "active"
)

type listeningWorkAPIError struct {
	statusCode int
	code       string
	message    string
	details    string
}

func (e *listeningWorkAPIError) Error() string {
	return e.message
}

// HandleListeningWork handles incoming HTTP requests for listening work orders.
func (s *APIServer) HandleListeningWork(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.sendErrorResponse(w, http.StatusMethodNotAllowed, ErrCodeInvalidRequest, "Method not allowed", "Only POST method is supported")
		return
	}

	var req apiModels.ListeningWorkRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.Logger.Error().Err(err).Msg("Failed to decode listening work request body")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeInvalidRequest, "Invalid request body", err.Error())
		return
	}

	if req.TopicID == "" {
		s.Logger.Error().Msg("Missing topic_id in listening work request")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeMissingField, "Missing required field", "topic_id is required")
		return
	}

	if req.WorkspaceID == "" {
		s.Logger.Error().Msg("Missing workspace_id in listening work request")
		s.sendErrorResponse(w, http.StatusBadRequest, ErrCodeMissingField, "Missing required field", "workspace_id is required")
		return
	}

	s.Logger.Info().
		Str("topic_id", req.TopicID).
		Str("workspace_id", req.WorkspaceID).
		Msg("Received listening work request")

	if err := s.processListeningWork(r.Context(), req); err != nil {
		s.Logger.Error().Err(err).Str("topic_id", req.TopicID).Msg("Failed to process listening work request")

		apiErr, ok := err.(*listeningWorkAPIError)
		if ok {
			s.sendErrorResponse(w, apiErr.statusCode, apiErr.code, apiErr.message, apiErr.details)
			return
		}

		s.sendErrorResponse(w, http.StatusInternalServerError, ErrCodeInternalError, "Failed to process request", err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
	s.sendSuccessResponse(w, map[string]interface{}{
		"status":    "accepted",
		"message":   "Listening work order queued successfully",
		"topic_id":  req.TopicID,
		"timestamp": time.Now().Format(time.RFC3339),
	})
}

// processListeningWork validates the topic and produces a work order to Kafka.
func (s *APIServer) processListeningWork(ctx context.Context, req apiModels.ListeningWorkRequest) error {
	if s.ListeningRepo == nil {
		return fmt.Errorf("APIServer.processListeningWork: listening repository not configured")
	}

	topic, err := s.ListeningRepo.GetTopicByID(ctx, req.TopicID)
	if err != nil {
		return fmt.Errorf("APIServer.processListeningWork: failed to fetch topic: %w", err)
	}
	if topic == nil {
		return &listeningWorkAPIError{
			statusCode: http.StatusNotFound,
			code:       ErrCodeTopicNotFound,
			message:    "Topic not found",
			details:    "No listening topic found for provided topic_id",
		}
	}

	// Heal SuperAdminID: if it's in the request but not in DB, use the request version
	if req.SuperAdminID != "" && topic.SuperAdminID == "" {
		topic.SuperAdminID = req.SuperAdminID
	}

	// Enforce workspace/topic ownership — prevent cross-tenant access
	if topic.WorkspaceID != req.WorkspaceID {
		return &listeningWorkAPIError{
			statusCode: http.StatusForbidden,
			code:       ErrCodeForbidden,
			message:    "Forbidden",
			details:    "topic does not belong to workspace",
		}
	}

	// ── Topic status check ────────────────────────────────────────────
	// Only scrape topics that are explicitly "active". Paused, deleted,
	// or any other status stops the pipeline from wasting API credits.
	if topic.Status != "" && topic.Status != TopicStatusActive {
		s.Logger.Info().
			Str("topic_id", topic.TopicID).
			Str("status", topic.Status).
			Msg("Skipping work order for non-active topic")
		return &listeningWorkAPIError{
			statusCode: http.StatusConflict,
			code:       ErrCodeTopicInactive,
			message:    "Topic is not active",
			details:    fmt.Sprintf("Topic status is '%s', only 'active' topics are scraped", topic.Status),
		}
	}

	// ── Workspace-level quota check ────────────────────────────────────
	// Prevents dispatching work orders when the owner's monthly mention
	// budget across all topics is exhausted.
	quotaID := topic.SuperAdminID
	if quotaID == "" {
		quotaID = topic.WorkspaceID
	}

	if hasListeningWorkspaceRepo(s.ListeningWorkspaceRepo) && quotaID != "" {
		mentionsCount, mentionLimit, exists, err := s.ListeningWorkspaceRepo.GetWorkspaceUsage(ctx, quotaID)
		if err != nil {
			s.Logger.Warn().Err(err).
				Str("workspace_id", topic.WorkspaceID).
				Str("super_admin_id", topic.SuperAdminID).
				Msg("Failed to check workspace quota at dispatch, proceeding (fail-open)")
		} else if !exists {
			return &listeningWorkAPIError{
				statusCode: http.StatusConflict,
				code:       ErrCodeWorkspaceLimitReached,
				message:    "Workspace mention limit reached",
				details:    "No listening quota found for the topic owner",
			}
		} else if mentionLimit > 0 && mentionsCount >= mentionLimit {
			return &listeningWorkAPIError{
				statusCode: http.StatusConflict,
				code:       ErrCodeWorkspaceLimitReached,
				message:    "Workspace mention limit reached",
				details:    "The workspace has exhausted its mention quota across all topics",
			}
		}
	}

	s.Logger.Debug().
		Str("topic_id", topic.TopicID).
		Strs("include_keywords", topic.IncludeKeywords).
		Strs("enabled_platforms", topic.EnabledPlatforms).
		Strs("languages", topic.Languages).
		Strs("include_any", topic.IncludeAny).
		Msg("Topic fields after Normalize")

	// ── Compute scraping date window ──────────────────────────────────
	now := time.Now().UTC()
	var fromDate time.Time

	if topic.LastFetchedAt.IsZero() {
		// First-ever scrape: look back 30 days from topic creation date.
		fromDate = topic.CreatedAt.Add(-initialCrawlWindow)
		s.Logger.Info().
			Str("topic_id", topic.TopicID).
			Time("from_date", fromDate).
			Time("created_at", topic.CreatedAt).
			Msg("Initial crawl: FromDate set to CreatedAt - 30 days")
	} else {
		// Subsequent (incremental) scrape: start from last fetch time.
		fromDate = topic.LastFetchedAt
		s.Logger.Info().
			Str("topic_id", topic.TopicID).
			Time("from_date", fromDate).
			Msg("Incremental crawl: FromDate set to LastFetchedAt")
	}

	workOrder := kafkaModels.ListeningWorkOrder{
		TopicID:                  topic.TopicID,
		WorkspaceID:              topic.WorkspaceID,
		SuperAdminID:             topic.SuperAdminID,
		IncludeKeywords:          topic.IncludeKeywords,
		ExcludeKeywords:          topic.ExcludeKeywords,
		IncludeAny:               topic.IncludeAny,
		IncludeAll:               topic.IncludeAll,
		ExactMatch:               topic.ExactMatch,
		CaseSensitive:            topic.CaseSensitive,
		IncludeAuthors:           topic.IncludeAuthors,
		ExcludeAuthors:           topic.ExcludeAuthors,
		Languages:                topic.Languages,
		Regions:                  topic.Regions,
		EnabledPlatforms:         topic.EnabledPlatforms,
		GlobalExcludedSubreddits: topic.GlobalExcludedSubreddits,
		MentionsLimit:            topic.MentionsLimit,
		Cursors:                  topic.LastFetchedCursors,
		FromDate:                 fromDate,
		ToDate:                   now,
		SyncType:                 req.SyncType,
	}

	data, err := json.Marshal(workOrder)
	if err != nil {
		return fmt.Errorf("APIServer.processListeningWork: marshal failed: %w", err)
	}

	s.Logger.Info().
		Str("topic_id", req.TopicID).
		Str("workspace_id", req.WorkspaceID).
		Str("sync_type", req.SyncType).
		Str("kafka_topic", kafkaModels.TopicListeningWork).
		Int("payload_size", len(data)).
		Msg("Publishing listening work order to Kafka")

	if err := s.Producer.Produce(ctx, kafkaModels.TopicListeningWork, []byte(req.TopicID), data); err != nil {
		s.Logger.Error().Err(err).
			Str("topic_id", req.TopicID).
			Str("workspace_id", req.WorkspaceID).
			Str("sync_type", req.SyncType).
			Str("kafka_topic", kafkaModels.TopicListeningWork).
			Int("payload_size", len(data)).
			Msg("Failed to publish listening work order to Kafka")
		return fmt.Errorf("APIServer.processListeningWork: Kafka send failed: %w", err)
	}

	s.Logger.Info().
		Str("topic_id", req.TopicID).
		Str("workspace_id", req.WorkspaceID).
		Str("sync_type", req.SyncType).
		Str("kafka_topic", kafkaModels.TopicListeningWork).
		Time("from_date", fromDate).
		Time("to_date", now).
		Msg("Listening work order dispatched")

	return nil
}

func hasListeningWorkspaceRepo(repo ListeningWorkspaceQuotaChecker) bool {
	if repo == nil {
		return false
	}

	value := reflect.ValueOf(repo)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return !value.IsNil()
	default:
		return true
	}
}
