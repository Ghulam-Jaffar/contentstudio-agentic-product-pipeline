package scheduler

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	redisdb "github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
)

type mockRecurringTopicRepository struct {
	topics                  []*mongomodels.ListeningTopic
	getActiveTopicsCalls    int
	setLimitReachedTopicIDs []string
	getErr                  error
	setErr                  error
}

func (m *mockRecurringTopicRepository) GetActiveTopics(_ context.Context) ([]*mongomodels.ListeningTopic, error) {
	m.getActiveTopicsCalls++
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.topics, nil
}

func (m *mockRecurringTopicRepository) SetMentionsLimitReached(_ context.Context, topicID string) error {
	m.setLimitReachedTopicIDs = append(m.setLimitReachedTopicIDs, topicID)
	return m.setErr
}

type mockOwnerQuotaChecker struct {
	mentionsCount int
	mentionLimit  int
	exists        bool
	err           error
	quotaIDs      []string
}

func (m *mockOwnerQuotaChecker) GetWorkspaceUsage(_ context.Context, quotaID string) (int, int, bool, error) {
	m.quotaIDs = append(m.quotaIDs, quotaID)
	return m.mentionsCount, m.mentionLimit, m.exists, m.err
}

func TestRecurringSchedulerRunOnceProducesIncrementalWorkOrder(t *testing.T) {
	t.Parallel()

	log, _ := logger.NewTestLogger()
	producer := &mockProducerRecorder{}
	repo := &mockRecurringTopicRepository{
		topics: []*mongomodels.ListeningTopic{
			{
				TopicID:           "topic-1",
				WorkspaceID:       "ws-1",
				SuperAdminID:      "sa-1",
				Status:            "active",
				IsInitialSyncDone: true,
				IncludeKeywords:   []string{"iran"},
				EnabledPlatforms:  []string{"twitter", "facebook"},
				MentionsLimit:     100,
				Usage: mongomodels.ListeningTopicUsage{
					MentionsCount: 12,
					MentionsLimit: 100,
				},
				LastFetchedAt: time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC),
				LastFetchedCursors: mongomodels.FlexStringMap{
					"twitter": "cursor-1",
				},
				CreatedAt: time.Date(2026, 4, 1, 8, 0, 0, 0, time.UTC),
			},
		},
	}

	scheduler := NewRecurringScheduler(repo, producer, nil, log, time.Minute)
	scheduler.now = func() time.Time {
		return time.Date(2026, 4, 11, 9, 0, 0, 0, time.UTC)
	}

	stats, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if stats.Total != 1 || stats.Produced != 1 || stats.Failed != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if len(producer.messages) != 1 {
		t.Fatalf("expected 1 produced message, got %d", len(producer.messages))
	}

	var workOrder kafkamodels.ListeningWorkOrder
	if err := json.Unmarshal(producer.messages[0].Value, &workOrder); err != nil {
		t.Fatalf("failed to decode work order: %v", err)
	}

	if workOrder.SyncType != "incremental" {
		t.Fatalf("expected incremental sync type, got %q", workOrder.SyncType)
	}
	if !workOrder.FromDate.Equal(time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected from_date: %s", workOrder.FromDate)
	}
	if !workOrder.ToDate.Equal(time.Date(2026, 4, 11, 9, 0, 0, 0, time.UTC)) {
		t.Fatalf("unexpected to_date: %s", workOrder.ToDate)
	}
	if workOrder.Cursors["twitter"] != "cursor-1" {
		t.Fatalf("expected twitter cursor to be preserved, got %+v", workOrder.Cursors)
	}
}

func TestRecurringSchedulerRunOnceSkipsPendingInvalidAndLimitReachedTopics(t *testing.T) {
	t.Parallel()

	log, _ := logger.NewTestLogger()
	producer := &mockProducerRecorder{}
	repo := &mockRecurringTopicRepository{
		topics: []*mongomodels.ListeningTopic{
			{
				TopicID:           "pending-topic",
				WorkspaceID:       "ws-1",
				Status:            "active",
				IsInitialSyncDone: false,
				IncludeKeywords:   []string{"iran"},
				EnabledPlatforms:  []string{"twitter"},
			},
			{
				TopicID:           "paused-topic",
				WorkspaceID:       "ws-1",
				Status:            "paused",
				IsInitialSyncDone: true,
				IncludeKeywords:   []string{"iran"},
				EnabledPlatforms:  []string{"twitter"},
			},
			{
				TopicID:           "invalid-topic",
				WorkspaceID:       "ws-1",
				Status:            "active",
				IsInitialSyncDone: true,
				EnabledPlatforms:  []string{"twitter"},
			},
			{
				TopicID:           "limited-topic",
				WorkspaceID:       "ws-1",
				Status:            "active",
				IsInitialSyncDone: true,
				IncludeKeywords:   []string{"iran"},
				EnabledPlatforms:  []string{"twitter"},
				MentionsLimit:     5,
				Usage: mongomodels.ListeningTopicUsage{
					MentionsCount: 5,
					MentionsLimit: 5,
				},
			},
		},
	}

	scheduler := NewRecurringScheduler(repo, producer, nil, log, time.Minute)
	stats, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Total != 4 {
		t.Fatalf("expected 4 total topics, got %+v", stats)
	}
	if stats.Produced != 1 {
		t.Fatalf("expected one work order for the historical-count topic, got %+v", stats)
	}
	if stats.SkippedInitialSyncPending != 1 || stats.SkippedInactive != 1 || stats.SkippedInvalid != 1 || stats.SkippedLimitReached != 0 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if len(repo.setLimitReachedTopicIDs) != 0 {
		t.Fatalf("expected no topic limit flags to be written, got %+v", repo.setLimitReachedTopicIDs)
	}
}

func TestRecurringSchedulerRunOnceSkipsWhenOwnerQuotaReached(t *testing.T) {
	t.Parallel()

	log, _ := logger.NewTestLogger()
	producer := &mockProducerRecorder{}
	repo := &mockRecurringTopicRepository{
		topics: []*mongomodels.ListeningTopic{
			{
				TopicID:           "topic-1",
				WorkspaceID:       "ws-1",
				SuperAdminID:      "sa-1",
				Status:            "active",
				IsInitialSyncDone: true,
				IncludeKeywords:   []string{"iran"},
				EnabledPlatforms:  []string{"twitter"},
				MentionsLimit:     100,
				Usage: mongomodels.ListeningTopicUsage{
					MentionsCount: 12,
					MentionsLimit: 100,
				},
			},
		},
	}
	quotaChecker := &mockOwnerQuotaChecker{
		mentionsCount: 1000,
		mentionLimit:  1000,
		exists:        true,
	}

	scheduler := NewRecurringScheduler(repo, producer, nil, log, time.Minute).
		WithOwnerQuotaChecker(quotaChecker)

	stats, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if stats.Total != 1 || stats.Produced != 0 || stats.SkippedQuotaReached != 1 {
		t.Fatalf("unexpected stats: %+v", stats)
	}
	if len(producer.messages) != 0 {
		t.Fatalf("expected no produced messages, got %d", len(producer.messages))
	}
	if len(quotaChecker.quotaIDs) != 1 || quotaChecker.quotaIDs[0] != "ws-1" {
		t.Fatalf("expected quota check for ws-1, got %+v", quotaChecker.quotaIDs)
	}
}

func TestRecurringSchedulerRunOnceSkipsWhenLockAlreadyHeld(t *testing.T) {
	t.Parallel()

	log, _ := logger.NewTestLogger()
	producer := &mockProducerRecorder{}
	repo := &mockRecurringTopicRepository{
		topics: []*mongomodels.ListeningTopic{
			{
				TopicID:           "topic-1",
				WorkspaceID:       "ws-1",
				Status:            "active",
				IsInitialSyncDone: true,
				IncludeKeywords:   []string{"iran"},
				EnabledPlatforms:  []string{"twitter"},
			},
		},
	}

	lockClient := newMockRedisForLock()
	lock := redisdb.NewDistributedLock(lockClient, log.Logger)
	token, err := lock.Acquire(context.Background(), recurringSchedulerLockKey, time.Minute)
	if err != nil {
		t.Fatalf("failed to pre-acquire lock: %v", err)
	}
	if token == "" {
		t.Fatal("expected pre-acquired token")
	}

	scheduler := NewRecurringScheduler(repo, producer, lock, log, time.Minute)
	stats, err := scheduler.RunOnce(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if repo.getActiveTopicsCalls != 0 {
		t.Fatalf("expected scheduler to skip before reading topics, got %d calls", repo.getActiveTopicsCalls)
	}
	if len(producer.messages) != 0 {
		t.Fatalf("expected no produced messages, got %d", len(producer.messages))
	}
	if stats.Total != 0 || stats.Produced != 0 {
		t.Fatalf("unexpected stats when lock is held: %+v", stats)
	}
}
