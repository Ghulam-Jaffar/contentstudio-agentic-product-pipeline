// Package scheduler drives the listening pipeline entrypoint: it periodically
// scans active topics in Mongo and enqueues fetch work orders onto Kafka, using
// a Redis lock so only one scheduler instance runs the recurring sweep at a time.
package scheduler

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	redisdb "github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

const (
	recurringSchedulerLockKey = "listening:scheduler:recurring"
	recurringCrawlWindow      = 30 * 24 * time.Hour
)

// RecurringTopicRepository defines the topic reads needed by the recurring scheduler.
type RecurringTopicRepository interface {
	GetActiveTopics(ctx context.Context) ([]*mongomodels.ListeningTopic, error)
}

// WorkOrderProducer defines the minimal producer contract used by the scheduler.
type WorkOrderProducer interface {
	Produce(ctx context.Context, topic string, key, value []byte) error
}

// SuperAdminResolver looks up the super_admin_id for a workspace from the
// workspace owner. Used to populate work orders when the
// topic document (created by Laravel) doesn't have super_admin_id.
type SuperAdminResolver interface {
	GetSuperAdminID(ctx context.Context, workspaceID string) (string, error)
}

// OwnerQuotaChecker returns the owner-level listening usage and limit.
type OwnerQuotaChecker interface {
	GetWorkspaceUsage(ctx context.Context, quotaID string) (mentionsCount, mentionLimit int, exists bool, err error)
}

// RecurringSchedulerStats captures one recurring scheduler pass for observability and tests.
type RecurringSchedulerStats struct {
	Total                     int
	Produced                  int
	Failed                    int
	SkippedInactive           int
	SkippedInitialSyncPending int
	SkippedInvalid            int
	SkippedLimitReached       int
	SkippedQuotaReached       int
}

// RecurringScheduler scans listening topics and emits incremental work orders for topics
// that are active, have completed initial sync, and whose owner still has mention budget left.
type RecurringScheduler struct {
	repo               RecurringTopicRepository
	producer           WorkOrderProducer
	superAdminResolver SuperAdminResolver
	quotaChecker       OwnerQuotaChecker
	lock               *redisdb.DistributedLock
	log                *logger.Logger
	lockTTL            time.Duration
	now                func() time.Time
}

// NewRecurringScheduler creates a recurring listening scheduler.
func NewRecurringScheduler(
	repo RecurringTopicRepository,
	producer WorkOrderProducer,
	lock *redisdb.DistributedLock,
	log *logger.Logger,
	lockTTL time.Duration,
) *RecurringScheduler {
	if lockTTL <= 0 {
		lockTTL = 15 * time.Minute
	}

	return &RecurringScheduler{
		repo:     repo,
		producer: producer,
		lock:     lock,
		log:      log,
		lockTTL:  lockTTL,
		now:      func() time.Time { return time.Now().UTC() },
	}
}

// WithSuperAdminResolver attaches a resolver that looks up super_admin_id
// from the workspace owner when it's missing on the topic document.
func (s *RecurringScheduler) WithSuperAdminResolver(resolver SuperAdminResolver) *RecurringScheduler {
	s.superAdminResolver = resolver
	return s
}

// WithOwnerQuotaChecker attaches owner-level quota checks so the scheduler can
// skip work orders that would be dropped later by quota enforcement.
func (s *RecurringScheduler) WithOwnerQuotaChecker(checker OwnerQuotaChecker) *RecurringScheduler {
	s.quotaChecker = checker
	return s
}

// BuildRecurringWorkOrder constructs an incremental work order from a listening topic.
func BuildRecurringWorkOrder(topic *mongomodels.ListeningTopic, now time.Time) kafkamodels.ListeningWorkOrder {
	fromDate := topic.LastFetchedAt
	if fromDate.IsZero() {
		base := topic.CreatedAt
		if base.IsZero() {
			base = now.UTC()
		}
		fromDate = base.Add(-recurringCrawlWindow)
	}

	return kafkamodels.ListeningWorkOrder{
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
		ToDate:                   now.UTC(),
		SyncType:                 "incremental",
	}
}

// RunOnce executes one recurring scheduling pass.
func (s *RecurringScheduler) RunOnce(ctx context.Context) (RecurringSchedulerStats, error) {
	var stats RecurringSchedulerStats

	if s.repo == nil {
		return stats, fmt.Errorf("RecurringScheduler.RunOnce: repository not configured")
	}
	if s.producer == nil {
		return stats, fmt.Errorf("RecurringScheduler.RunOnce: producer not configured")
	}

	if s.lock != nil {
		token, err := s.lock.Acquire(ctx, recurringSchedulerLockKey, s.lockTTL)
		if err != nil {
			return stats, fmt.Errorf("RecurringScheduler.RunOnce: acquire lock: %w", err)
		}
		if token == "" {
			s.log.Info().Str("lock_key", recurringSchedulerLockKey).Msg("Recurring listening scheduler already running, skipping pass")
			return stats, nil
		}
		defer s.lock.Release(ctx, recurringSchedulerLockKey, token)
	}

	topics, err := s.repo.GetActiveTopics(ctx)
	if err != nil {
		return stats, fmt.Errorf("RecurringScheduler.RunOnce: get topics: %w", err)
	}

	now := s.now().UTC()
	s.log.Info().Int("candidate_topics", len(topics)).Time("run_at", now).Msg("Recurring listening scheduler pass started")

	for _, topic := range topics {
		stats.Total++
		topic.Normalize()

		topicLog := s.log.With().
			Str("topic_id", topic.TopicID).
			Str("workspace_id", topic.WorkspaceID).
			Str("status", topic.Status).
			Logger()

		switch {
		case topic.TopicID == "" || topic.WorkspaceID == "":
			stats.SkippedInvalid++
			topicLog.Warn().Msg("Skipping recurring sync: missing topic or workspace identifier")
			continue
		case topic.Status != "" && topic.Status != "active":
			stats.SkippedInactive++
			topicLog.Debug().Msg("Skipping recurring sync: topic is not active")
			continue
		case !topic.IsInitialSyncDone:
			stats.SkippedInitialSyncPending++
			topicLog.Debug().Msg("Skipping recurring sync: initial sync not completed")
			continue
		case len(topic.IncludeKeywords) == 0:
			stats.SkippedInvalid++
			topicLog.Warn().Msg("Skipping recurring sync: topic has no include keywords")
			continue
		case len(topic.EnabledPlatforms) == 0:
			stats.SkippedInvalid++
			topicLog.Warn().Msg("Skipping recurring sync: topic has no enabled platforms")
			continue
		}

		// Resolve SuperAdminID from the workspace owner when missing on the topic.
		// Laravel-created topics store created_by (the creating user) instead of
		// super_admin_id (the billing owner), so we look it up from the workspace.
		if topic.SuperAdminID == "" && topic.WorkspaceID != "" && s.superAdminResolver != nil {
			saID, err := s.superAdminResolver.GetSuperAdminID(ctx, topic.WorkspaceID)
			if err != nil {
				topicLog.Warn().Err(err).Msg("Failed to resolve super_admin_id from workspace owner")
			} else if saID != "" {
				topic.SuperAdminID = saID
			}
		}

		quotaID := topic.WorkspaceID
		if s.quotaChecker != nil && quotaID != "" {
			mentionsCount, mentionLimit, exists, err := s.quotaChecker.GetWorkspaceUsage(ctx, quotaID)
			if err != nil {
				topicLog.Warn().Err(err).Str("quota_id", quotaID).
					Msg("Failed to read workspace quota during recurring scheduling")
			} else if !exists {
				stats.SkippedQuotaReached++
				topicLog.Info().Str("quota_id", quotaID).
					Msg("Skipping recurring sync: workspace quota document not found")
				continue
			} else if mentionLimit > 0 && mentionsCount >= mentionLimit {
				stats.SkippedQuotaReached++
				topicLog.Info().
					Str("quota_id", quotaID).
					Int("workspace_mentions_count", mentionsCount).
					Int("workspace_mentions_limit", mentionLimit).
					Msg("Skipping recurring sync: workspace mention quota reached")
				continue
			}
		}

		workOrder := BuildRecurringWorkOrder(topic, now)
		if topic.LastFetchedAt.IsZero() {
			topicLog.Warn().
				Time("created_at", topic.CreatedAt).
				Time("fallback_from_date", workOrder.FromDate).
				Msg("Recurring sync topic is missing last_fetched_at, falling back to created_at - 30 days")
		}

		data, err := json.Marshal(workOrder)
		if err != nil {
			stats.Failed++
			topicLog.Error().Err(err).Msg("Failed to marshal recurring listening work order")
			continue
		}

		if err := s.producer.Produce(ctx, kafkamodels.TopicListeningWork, []byte(topic.TopicID), data); err != nil {
			stats.Failed++
			topicLog.Error().Err(err).Msg("Failed to produce recurring listening work order")
			continue
		}

		stats.Produced++
		topicLog.Info().
			Str("sync_type", workOrder.SyncType).
			Time("from_date", workOrder.FromDate).
			Time("to_date", workOrder.ToDate).
			Int("mentions_count", topic.Usage.MentionsCount).
			Int("mentions_limit", topic.MentionsLimit).
			Int("keyword_count", len(topic.IncludeKeywords)).
			Int("platform_count", len(topic.EnabledPlatforms)).
			Msg("Produced recurring listening work order")
	}

	s.log.Info().
		Int("total", stats.Total).
		Int("produced", stats.Produced).
		Int("failed", stats.Failed).
		Int("skipped_inactive", stats.SkippedInactive).
		Int("skipped_initial_sync_pending", stats.SkippedInitialSyncPending).
		Int("skipped_invalid", stats.SkippedInvalid).
		Int("skipped_limit_reached", stats.SkippedLimitReached).
		Int("skipped_quota_reached", stats.SkippedQuotaReached).
		Msg("Recurring listening scheduler pass completed")

	return stats, nil
}
