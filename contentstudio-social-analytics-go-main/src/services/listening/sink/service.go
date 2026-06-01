// Package sink is the fast-path persistence stage: it consumes parsed mentions
// off listening-parsed and inserts them into ClickHouse immediately with empty
// sentiment and ai_tags so the UI can surface new mentions without waiting for
// AI enrichment. The enrichment stage consumes the same parsed topic in
// parallel, calls the AI agents API, and re-inserts the row with tags and
// sentiment — since the ClickHouse table uses ReplacingMergeTree(updated_at)
// sorted on (topic_id, platform, mention_id), the enriched row (later
// updated_at) wins on merge, so this dual-consumer pattern converges to a
// single row per mention without coordination.
//
// Retries use bounded exponential backoff because partial ClickHouse outages
// can otherwise stall the whole pipeline while retries snowball.
package sink

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	chmodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	defaultMaxRetries    = 3
	defaultRetryBaseWait = 500 * time.Millisecond
)

// MentionWriter abstracts ClickHouse batch insert for testability.
type MentionWriter interface {
	InsertMentions(ctx context.Context, mentions []chmodels.ListeningMentionRow) error
}

// TopicUpdater abstracts MongoDB topic counter updates for testability.
type TopicUpdater interface {
	GetMentionsCount(ctx context.Context, topicID string) (int, error)
	IncrementMentionsCount(ctx context.Context, topicID string, count int) error
	SetMentionsLimitReached(ctx context.Context, topicID string) error
	UpdateLastFetched(ctx context.Context, topicID string, fetchedAt time.Time, cursors map[string]string) error
	MarkFirstMentionsReceived(ctx context.Context, topicID string) (bool, error)
}

// FirstBatchNotifier broadcasts a real-time event when a topic transitions
// from "never received mentions" to "received mentions now". The frontend
// listens for this on a per-workspace Pusher channel to complete the
// "collecting your first mentions" progress animation.
type FirstBatchNotifier interface {
	Trigger(channel, event string, data interface{}) error
}

const (
	// firstBatchChannelPrefix mirrors the existing public `analytics-channel-{ws}`
	// convention used by the rest of the analytics Pusher app. No `private-`
	// prefix because the analytics Pusher app does not have a server-side auth
	// endpoint configured for private channels.
	firstBatchChannelPrefix = "listening-channel-"
	firstBatchEvent         = "listening.mentions.first_batch"
)

// TopicMentionReserver atomically reserves and releases per-topic capacity.
type TopicMentionReserver interface {
	TryReserveMentionSlot(ctx context.Context, topicID string, mentionsLimit int) (reserved bool, newCount int, err error)
	ReleaseMentionSlot(ctx context.Context, topicID string, mentionsLimit int) error
}

// WorkspaceUpdater abstracts workspace-level subscription quota operations for testability.
// The workspace document is owned by the billing system; this service only increments counts.
type WorkspaceUpdater interface {
	// IsWorkspaceMentionLimitReached returns true when the workspace quota is exhausted.
	IsWorkspaceMentionLimitReached(ctx context.Context, workspaceID string) (bool, error)
	// IncrementWorkspaceMentionsCount atomically increments the counter and returns
	// the new total and the subscription limit (both 0 when no workspace doc exists).
	IncrementWorkspaceMentionsCount(ctx context.Context, workspaceID string, count int) (newCount int, mentionLimit int, err error)
	// SetWorkspaceMentionLimitReached flags the workspace as quota-exhausted.
	SetWorkspaceMentionLimitReached(ctx context.Context, workspaceID string) error
}

// WorkspaceMentionReserver atomically reserves and releases workspace-level capacity.
type WorkspaceMentionReserver interface {
	TryReserveWorkspaceMention(ctx context.Context, workspaceID string) (reserved bool, newCount int, mentionLimit int, err error)
	ReleaseWorkspaceMentionReservation(ctx context.Context, workspaceID string) error
}

// SinkService consumes parsed mentions, persists them to ClickHouse,
// and updates MongoDB topic history plus workspace-level monthly quota counters.
type SinkService struct {
	writer             MentionWriter
	updater            TopicUpdater
	workspaceUpdater   WorkspaceUpdater // nil disables workspace-level quota enforcement
	producer           kafka.Producer
	firstBatchNotifier FirstBatchNotifier // nil disables first-batch Pusher notifications
	log                *logger.Logger
	maxRetries         int
}

// NewSinkService creates a new SinkService.
// workspaceUpdater may be nil to skip workspace-level quota enforcement (e.g. in tests).
func NewSinkService(
	writer MentionWriter,
	updater TopicUpdater,
	workspaceUpdater WorkspaceUpdater,
	producer kafka.Producer,
	log *logger.Logger,
	maxRetries int,
) *SinkService {
	if maxRetries <= 0 {
		maxRetries = defaultMaxRetries
	}
	return &SinkService{
		writer:           writer,
		updater:          updater,
		workspaceUpdater: workspaceUpdater,
		producer:         producer,
		log:              log,
		maxRetries:       maxRetries,
	}
}

// WithFirstBatchNotifier wires an optional Pusher trigger that fires once per
// topic on the first successful mention insert. Safe to omit in tests.
func (s *SinkService) WithFirstBatchNotifier(n FirstBatchNotifier) *SinkService {
	s.firstBatchNotifier = n
	return s
}

// HandleParsedMention is a kafka.MessageHandler that persists a single mention.
func (s *SinkService) HandleParsedMention(ctx context.Context, _ string, _ []byte, value []byte) error {
	var mention kafkamodels.ListeningMention
	if err := json.Unmarshal(value, &mention); err != nil {
		return fmt.Errorf("SinkService.HandleParsedMention: unmarshal: %w", err)
	}

	row := mentionToRow(mention)

	if err := s.insertWithRetry(ctx, []chmodels.ListeningMentionRow{row}); err != nil {
		s.log.Error().Err(err).
			Str("mention_id", mention.MentionID).
			Msg("Insert failed after retries, sending to DLQ")
		return s.sendToDLQ(ctx, "sink", value, err)
	}

	s.notifyFirstBatch(ctx, mention)

	if err := s.updateBookkeeping(ctx, mention); err != nil {
		if strings.Contains(err.Error(), "topic not found") {
			s.log.Info().
				Str("topic_id", mention.TopicID).
				Str("mention_id", mention.MentionID).
				Msg("Topic deleted during mention processing, discarding")
			return nil
		}
		s.log.Error().Err(err).
			Str("mention_id", mention.MentionID).
			Str("topic_id", mention.TopicID).
			Msg("Bookkeeping failed, sending to DLQ")
		return s.sendToDLQ(ctx, "sink-bookkeeping", value, err)
	}

	s.log.Debug().
		Str("mention_id", mention.MentionID).
		Str("topic_id", mention.TopicID).
		Msg("Persisted mention")

	return nil
}

// notifyFirstBatch fires the per-workspace Pusher event the first time a
// non-empty mention lands for a topic. Wins-the-race semantics come from
// MarkFirstMentionsReceived (Mongo $set on a null-only filter), so even with
// many concurrent sink workers the event triggers exactly once per topic.
// Failures are logged and swallowed — the animation has a polling fallback
// on the frontend, so a Pusher hiccup must not block the data path.
func (s *SinkService) notifyFirstBatch(ctx context.Context, mention kafkamodels.ListeningMention) {
	if mention.TopicID == "" {
		return
	}

	if s.firstBatchNotifier != nil && mention.WorkspaceID == "" {
		s.log.Warn().
			Str("topic_id", mention.TopicID).
			Msg("First-batch transition skipped: empty workspace_id cannot derive Pusher channel")
		return
	}

	won, err := s.updater.MarkFirstMentionsReceived(ctx, mention.TopicID)
	if err != nil {
		s.log.Warn().Err(err).
			Str("topic_id", mention.TopicID).
			Msg("MarkFirstMentionsReceived failed; skipping first-batch notification")
		return
	}
	if !won {
		return
	}

	if s.firstBatchNotifier == nil {
		return
	}

	channel := firstBatchChannelPrefix + mention.WorkspaceID
	payload := map[string]any{
		"topic_id":     mention.TopicID,
		"workspace_id": mention.WorkspaceID,
		"received_at":  time.Now().UTC().Format(time.RFC3339),
	}

	if err := s.firstBatchNotifier.Trigger(channel, firstBatchEvent, payload); err != nil {
		s.log.Warn().Err(err).
			Str("topic_id", mention.TopicID).
			Str("workspace_id", mention.WorkspaceID).
			Msg("Failed to trigger first-batch Pusher event")
		return
	}

	s.log.Info().
		Str("topic_id", mention.TopicID).
		Str("workspace_id", mention.WorkspaceID).
		Msg("First-batch Pusher event triggered")
}

func (s *SinkService) updateBookkeeping(ctx context.Context, mention kafkamodels.ListeningMention) error {
	quotaID := mention.SuperAdminID
	if quotaID == "" {
		quotaID = mention.WorkspaceID
	}

	if err := s.updater.IncrementMentionsCount(ctx, mention.TopicID, 1); err != nil {
		return fmt.Errorf("SinkService.updateBookkeeping: increment topic mentions count: %w", err)
	}

	if quotaID != "" && s.workspaceUpdater != nil {
		workspaceCount, workspaceLimit, err := s.workspaceUpdater.IncrementWorkspaceMentionsCount(ctx, quotaID, 1)
		if err != nil {
			return fmt.Errorf("SinkService.updateBookkeeping: increment workspace mentions count: %w", err)
		}

		if workspaceLimit > 0 && workspaceCount >= workspaceLimit {
			if err := s.workspaceUpdater.SetWorkspaceMentionLimitReached(ctx, quotaID); err != nil {
				return fmt.Errorf("SinkService.updateBookkeeping: set workspace mention limit reached: %w", err)
			}
			s.log.Info().
				Str("id", quotaID).
				Int("mentions_count", workspaceCount).
				Int("mention_limit", workspaceLimit).
				Msg("Workspace mention limit reached")
		}
	}

	return nil
}

// insertWithRetry retries ClickHouse inserts with exponential backoff.
func (s *SinkService) insertWithRetry(ctx context.Context, rows []chmodels.ListeningMentionRow) error {
	var lastErr error
	wait := defaultRetryBaseWait

	for attempt := 1; attempt <= s.maxRetries; attempt++ {
		if err := s.writer.InsertMentions(ctx, rows); err != nil {
			lastErr = err
			s.log.Warn().
				Err(err).
				Int("attempt", attempt).
				Int("max_retries", s.maxRetries).
				Msg("ClickHouse insert failed, retrying")

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(wait):
			}
			wait *= 2
			continue
		}
		return nil
	}

	return fmt.Errorf("SinkService.insertWithRetry: all %d retries exhausted: %w", s.maxRetries, lastErr)
}

// sendToDLQ produces a failed mention to the dead-letter queue.
func (s *SinkService) sendToDLQ(ctx context.Context, stage string, payload []byte, origErr error) error {
	dlq := kafkamodels.ListeningDLQMessage{
		OriginalTopic: kafkamodels.TopicListeningParsed,
		Stage:         stage,
		Error:         origErr.Error(),
		Payload:       payload,
		AttemptCount:  s.maxRetries,
		Timestamp:     time.Now().UTC(),
	}

	data, err := json.Marshal(dlq)
	if err != nil {
		return fmt.Errorf("SinkService.sendToDLQ: marshal: %w", err)
	}

	if err := s.producer.Produce(ctx, kafkamodels.TopicListeningDLQ, nil, data); err != nil {
		return fmt.Errorf("SinkService.sendToDLQ: produce: %w", err)
	}

	return nil
}

// mentionToRow converts a Kafka mention to a ClickHouse row.
func mentionToRow(m kafkamodels.ListeningMention) chmodels.ListeningMentionRow {
	return chmodels.ListeningMentionRow{
		MentionID:         m.MentionID,
		TopicID:           m.TopicID,
		Platform:          m.Platform,
		NativeID:          m.NativeID,
		ContentHash:       m.ContentHash,
		AuthorID:          m.AuthorID,
		AuthorName:        m.AuthorName,
		AuthorHandle:      m.AuthorHandle,
		AuthorImageURL:    m.AuthorImageURL,
		AuthorURL:         m.AuthorURL,
		AuthorFollowers:   m.AuthorFollowers,
		PostText:          m.PostText,
		Language:          m.Language,
		PostedAt:          m.PostedAt,
		MatchedKeywords:   m.MatchedKeywords,
		TotalEngagement:   m.TotalEngagement,
		LikesCount:        m.LikesCount,
		CommentsCount:     m.CommentsCount,
		SharesCount:       m.SharesCount,
		ContentType:       m.ContentType,
		MediaType:         m.MediaType,
		URL:               m.URL,
		MediaURLs:         m.MediaURLs,
		AITags:            m.AITags,
		SentimentLabel:    m.SentimentLabel,
		SentimentScore:    m.SentimentScore,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
		PostRead:          m.PostRead,
		PostIrrelevant:    m.PostIrrelevant,
		Bookmark:          m.Bookmark,
		SentimentOverride: m.SentimentOverride,
	}
}
