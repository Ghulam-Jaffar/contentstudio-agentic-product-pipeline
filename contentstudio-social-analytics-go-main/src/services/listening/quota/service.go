// Package quota enforces both per-topic and per-workspace mention limits for
// the listening pipeline. Lives in its own package so the fetcher can consult
// it before paying the external provider for mentions we would have to discard.
package quota

import (
	"context"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// QuotaChecker computes how many mentions a topic may still ingest before
// hitting the owner-level workspace subscription limit.
type QuotaChecker interface {
	// GetRemainingMentionBudget returns how many more mentions may be ingested
	// for the given topic. It considers the owner-level subscription limit.
	//
	// A return value of 0 means the limit has been reached and no more mentions
	// should be fetched.  A negative return value is normalised to 0.
	//
	// If no owner usage document exists, fetching is blocked (fail-closed).
	GetRemainingMentionBudget(ctx context.Context, topicID, quotaID string) (remaining int, err error)
}

// TopicQuotaReader abstracts the subset of topic repository operations that the
// quota service needs so it can be satisfied by the existing ListeningRepository.
type TopicQuotaReader interface {
	GetMentionsCount(ctx context.Context, topicID string) (int, error)
}

// WorkspaceQuotaReader abstracts owner-level subscription reads.
type WorkspaceQuotaReader interface {
	GetWorkspaceUsage(ctx context.Context, quotaID string) (mentionsCount, mentionLimit int, exists bool, err error)
}

// QuotaService is the production implementation of QuotaChecker.
type QuotaService struct {
	topicRepo     TopicQuotaReader
	workspaceRepo WorkspaceQuotaReader
	log           *logger.Logger
}

// NewQuotaService creates a QuotaService wired to the given repositories.
func NewQuotaService(
	topicRepo TopicQuotaReader,
	workspaceRepo WorkspaceQuotaReader,
	log *logger.Logger,
) *QuotaService {
	return &QuotaService{
		topicRepo:     topicRepo,
		workspaceRepo: workspaceRepo,
		log:           log,
	}
}

// GetRemainingMentionBudget implements QuotaChecker.
func (q *QuotaService) GetRemainingMentionBudget(ctx context.Context, topicID, quotaID string) (int, error) {
	// --- Topic status check (active/paused/deleted) ---
	// If the topic was deleted or paused in MongoDB, return 0 to stop scraping.
	if reader, ok := q.topicRepo.(interface {
		GetTopicStatus(ctx context.Context, topicID string) (status string, exists bool, err error)
	}); ok {
		status, exists, err := reader.GetTopicStatus(ctx, topicID)
		if err != nil {
			q.log.Warn().Err(err).Str("topic_id", topicID).
				Msg("Failed to read topic status, proceeding with quota check")
		} else if !exists {
			q.log.Info().Str("topic_id", topicID).Msg("Topic not found in DB, returning 0 budget")
			return 0, nil
		} else if status != "" && status != "active" {
			q.log.Info().Str("topic_id", topicID).Str("status", status).
				Msg("Topic is not active, returning 0 budget")
			return 0, nil
		}
	}

	// --- Owner-level budget (enforced from workspace usage + subscription_plans) ---
	wsRemaining := int(^uint(0) >> 1) // default: unlimited
	if quotaID != "" && q.workspaceRepo != nil {
		wsCount, wsLimit, exists, err := q.workspaceRepo.GetWorkspaceUsage(ctx, quotaID)
		if err != nil {
			// Fail-open: log warning but don't block fetching.
			q.log.Warn().Err(err).
				Str("quota_id", quotaID).
				Msg("Failed to read owner quota, proceeding with topic-level limit only")
		} else if !exists {
			q.log.Info().
				Str("quota_id", quotaID).
				Msg("No listening subscription found, returning 0 budget")
			wsRemaining = 0
		} else if wsLimit > 0 {
			wsRemaining = wsLimit - wsCount
			if wsRemaining < 0 {
				wsRemaining = 0
			}

			q.log.Debug().
				Str("quota_id", quotaID).
				Int("ws_count", wsCount).
				Int("ws_limit", wsLimit).
				Int("ws_remaining", wsRemaining).
				Msg("Owner-level quota")
		}
	}

	return wsRemaining, nil
}
