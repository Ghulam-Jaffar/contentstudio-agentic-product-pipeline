// Package fetcher consumes scheduler work orders and calls the external
// listening provider API for each topic. It applies per-topic and per-owner
// quota before fetching to avoid paying for mentions we already cannot store,
// then emits raw mention payloads onto the parser topic.
package fetcher

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	kafkamodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/services/listening/quota"
)

// QuotaChecker computes how many mentions a topic may still ingest.
type QuotaChecker interface {
	GetRemainingMentionBudget(ctx context.Context, topicID, quotaID string) (remaining int, err error)
}

// SuperAdminResolver looks up the super_admin_id for a workspace.
type SuperAdminResolver interface {
	GetSuperAdminID(ctx context.Context, workspaceID string) (string, error)
}

// Data365API abstracts the Data365 client for testability.
type Data365API interface {
	TriggerSearch(ctx context.Context, platform, keyword string, maxPosts int, fromDate, toDate time.Time, languages []string) error
	PollUntilFinished(ctx context.Context, platform, keyword string, maxPosts int, fromDate, toDate time.Time, languages []string) error
	FetchResults(ctx context.Context, platform, keyword, cursor string, fromDate, toDate time.Time, languages []string) (*social.Data365SearchResult, error)
}

// FetchProgressTracker persists successful fetch progress for incremental runs.
type FetchProgressTracker interface {
	UpdateLastFetched(ctx context.Context, topicID string, fetchedAt time.Time, cursors map[string]string) error
}

// TopicSyncMarker persists the initial-sync-done flag directly in MongoDB.
// Replaces the previous HTTP callback to the Laravel backend: the Go pipeline
// now owns this write since it is the source of truth for fetch completion
// and Laravel had no side-effects on the update beyond persisting the flag.
type TopicSyncMarker interface {
	MarkInitialSyncDone(ctx context.Context, topicID, workspaceID string, eventAt time.Time) (bool, error)
}

// EmptyBatchNotifier broadcasts a real-time event when an initial sync
// completes without emitting any mentions. The frontend listens for this
// alongside the sink's success event to exit the awaiting state — without
// it, a topic that yields zero results from Data365 leaves the loader
// asymptoting at 95% with no terminal signal.
type EmptyBatchNotifier interface {
	Trigger(channel, event string, data interface{}) error
}

// QuotaTopicCounter returns how many active topics share a given quota
// namespace (super-admin or workspace). The fetcher uses this to clamp the
// per-keyword max_posts cap to a topic's fair share of the shared mention
// budget — without it, a single topic's initial sync (effectiveMaxInitial
// = 5000) can drain the entire budget on the first WO, starving sibling
// topics created in the same onboarding burst. Nil disables the dynamic
// clamp and the configured cap is used as-is.
type QuotaTopicCounter interface {
	CountActiveTopicsForQuota(ctx context.Context, quotaID string) (int, error)
}

const (
	emptyBatchChannelPrefix = "listening-channel-"
	emptyBatchEvent         = "listening.mentions.first_batch_empty"
)

// Default per-keyword max_posts caps when the corresponding env override is unset.
// Incremental default of 500 covers most steady-state daily volumes without
// dropping mentions on moderately-active topics; initial-sync default is
// higher because the first crawl needs more headroom to complete in 1-2 cycles
// for high-volume topics. Both are overridable via APP_LISTENING_BATCH_SIZE
// and APP_LISTENING_BATCH_SIZE_INITIAL — set lower if Data365 cost is a
// constraint, higher if topics consistently hit the cap and lose mentions.
const (
	defaultMaxPostsIncremental = 500
	defaultMaxPostsInitial     = 5000
)

// FetcherService consumes listening work orders from Kafka, fetches raw data
// from Data365 for each enabled platform, and produces raw payloads downstream.
type FetcherService struct {
	data365             Data365API
	producer            kafka.Producer
	lock                *redis.DistributedLock
	quota               QuotaChecker // nil disables quota enforcement at the fetcher level
	tracker             *quota.DistributedQuotaTracker
	superAdminResolver  SuperAdminResolver
	progress            FetchProgressTracker
	syncMarker          TopicSyncMarker
	emptyNotifier       EmptyBatchNotifier // nil disables empty-batch Pusher notifications
	quotaTopicCounter   QuotaTopicCounter  // nil disables fair-share clamp on per-keyword cap
	log                 *logger.Logger
	lockTTL             time.Duration
	maxPostsIncremental int
	maxPostsInitial     int
}

// NewFetcherService creates a new FetcherService.
//
// maxPostsIncremental caps per-keyword fetches for steady-state work orders
// (SyncType != "initial"). maxPostsInitial caps the same for the first crawl
// of a topic. Pass 0 for either to fall back to the package defaults.
func NewFetcherService(
	data365 Data365API,
	producer kafka.Producer,
	lock *redis.DistributedLock,
	log *logger.Logger,
	lockTTLMin int,
	maxPostsIncremental int,
	maxPostsInitial int,
	quota QuotaChecker,
) *FetcherService {
	ttl := time.Duration(lockTTLMin) * time.Minute
	if ttl == 0 {
		ttl = 10 * time.Minute
	}
	if maxPostsIncremental == 0 {
		maxPostsIncremental = defaultMaxPostsIncremental
	}
	if maxPostsInitial == 0 {
		maxPostsInitial = defaultMaxPostsInitial
	}
	return &FetcherService{
		data365:             data365,
		producer:            producer,
		lock:                lock,
		quota:               quota,
		log:                 log,
		lockTTL:             ttl,
		maxPostsIncremental: maxPostsIncremental,
		maxPostsInitial:     maxPostsInitial,
	}
}

// effectiveMaxPosts returns the per-keyword cap for a given work order based on
// whether it is an initial sync or an incremental run.
func (s *FetcherService) effectiveMaxPosts(syncType string) int {
	if syncType == "initial" {
		return s.maxPostsInitial
	}
	return s.maxPostsIncremental
}

// cursorKey is the composite key used in WorkOrder.Cursors and
// LastFetchedCursors to disambiguate per-platform-per-keyword resume tokens.
// Format: "<platform>:<keyword>". Keep this stable — it's persisted in MongoDB.
func cursorKey(platform, keyword string) string {
	return platform + ":" + keyword
}

// collectCursors snapshots a sync.Map of cursor accumulator entries into the
// plain map[string]string shape that FetchProgressTracker.UpdateLastFetched
// expects. Returns nil when no entries were stored — callers should treat
// that as "no cursor change" rather than "clear all cursors."
func collectCursors(m *sync.Map) map[string]string {
	if m == nil {
		return nil
	}
	out := map[string]string{}
	m.Range(func(k, v any) bool {
		key, kok := k.(string)
		val, vok := v.(string)
		if kok && vok {
			out[key] = val
		}
		return true
	})
	if len(out) == 0 {
		return nil
	}
	return out
}

// WithTopicSyncMarker attaches a Mongo-backed marker for flipping
// is_initial_sync_done after the first crawl finishes.
func (s *FetcherService) WithTopicSyncMarker(marker TopicSyncMarker) *FetcherService {
	s.syncMarker = marker
	return s
}

// WithEmptyBatchNotifier wires an optional Pusher trigger that fires once
// per topic when an initial sync completes without emitting any mentions.
// Safe to omit in tests and when Pusher config is incomplete — the FE has
// a timeout fallback so a missing notifier degrades gracefully.
func (s *FetcherService) WithEmptyBatchNotifier(n EmptyBatchNotifier) *FetcherService {
	s.emptyNotifier = n
	return s
}

// WithQuotaTopicCounter wires a counter used to clamp the per-keyword
// max_posts cap to this topic's fair share of the shared mention budget.
// Required for sane onboarding behaviour when several topics under one
// super-admin go through initial sync simultaneously; without it, the
// first WO to process can drain the entire budget on its 5000-cap initial
// crawl, leaving sibling topics permanently stuck at is_initial_sync_done
// = false until the billing cycle resets. Nil leaves the configured cap
// untouched (preserves prior behaviour).
func (s *FetcherService) WithQuotaTopicCounter(c QuotaTopicCounter) *FetcherService {
	s.quotaTopicCounter = c
	return s
}

// WithProgressTracker attaches a progress tracker for advancing LastFetchedAt after a successful run.
func (s *FetcherService) WithProgressTracker(progress FetchProgressTracker) *FetcherService {
	s.progress = progress
	return s
}

// WithDistributedQuota attaches a distributed quota tracker for cross-worker coordination.
func (s *FetcherService) WithDistributedQuota(tracker *quota.DistributedQuotaTracker) *FetcherService {
	s.tracker = tracker
	return s
}

// WithSuperAdminResolver attaches a resolver for looking up super_admin_id
// from the workspace owner when the work order doesn't have it.
func (s *FetcherService) WithSuperAdminResolver(resolver SuperAdminResolver) *FetcherService {
	s.superAdminResolver = resolver
	return s
}

// HandleWorkOrder is a kafka.MessageHandler that processes a single listening work order.
func (s *FetcherService) HandleWorkOrder(ctx context.Context, _ string, _ []byte, value []byte) error {
	s.log.Debug().RawJSON("raw_payload", value).Msg("Received work order payload")

	var order kafkamodels.ListeningWorkOrder
	if err := json.Unmarshal(value, &order); err != nil {
		return fmt.Errorf("FetcherService.HandleWorkOrder: unmarshal: %w", err)
	}

	log := s.log.With().
		Str("topic_id", order.TopicID).
		Str("workspace_id", order.WorkspaceID).
		Logger()

	log.Debug().
		Strs("include_keywords", order.IncludeKeywords).
		Strs("enabled_platforms", order.EnabledPlatforms).
		Strs("languages", order.Languages).
		Msg("Decoded work order fields")

	quotaID := order.SuperAdminID
	if quotaID == "" && order.WorkspaceID != "" && s.superAdminResolver != nil {
		saID, err := s.superAdminResolver.GetSuperAdminID(ctx, order.WorkspaceID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to resolve super_admin_id from workspace owner, skipping work order (fail-closed)")
			return nil
		} else if saID != "" {
			quotaID = saID
		}
	}
	if quotaID == "" {
		quotaID = order.WorkspaceID
	}

	// ── Pre-flight quota check ──────────────────────────────────────────
	// Fail-closed: if we cannot verify budget, skip the work order to
	// avoid wasting expensive Data365 API credits.
	if s.tracker != nil && s.quota != nil {
		remaining, err := s.tracker.GetRemaining(ctx, quotaID, func() (int, error) {
			return s.quota.GetRemainingMentionBudget(ctx, order.TopicID, quotaID)
		})
		if err != nil {
			log.Error().Err(err).Msg("Distributed quota pre-check failed, skipping work order (fail-closed)")
			return nil
		}
		if remaining <= 0 {
			log.Info().Msg("Distributed mention budget exhausted, skipping entire work order")
			return nil
		}
		log.Info().Int64("remaining_budget", remaining).Msg("Distributed quota pre-check passed")
	} else if s.quota != nil {
		remaining, err := s.quota.GetRemainingMentionBudget(ctx, order.TopicID, quotaID)
		if err != nil {
			log.Error().Err(err).Msg("Quota pre-check failed, skipping work order (fail-closed)")
			return nil
		}
		if remaining <= 0 {
			log.Info().Msg("Mention budget exhausted, skipping entire work order")
			return nil
		}
		log.Info().Int("remaining_budget", remaining).Msg("Quota pre-check passed")
	}

	lockKey := fmt.Sprintf("listening:fetch:%s", order.TopicID)
	token, err := s.lock.Acquire(ctx, lockKey, s.lockTTL)
	if err != nil {
		log.Error().Err(err).Msg("Failed to acquire lock")
		return fmt.Errorf("FetcherService.HandleWorkOrder: lock acquire: %w", err)
	}
	if token == "" {
		log.Info().Msg("Another fetcher is processing this topic, skipping")
		return nil
	}
	defer s.releaseLock(lockKey, token)

	log.Info().
		Strs("platforms", order.EnabledPlatforms).
		Strs("keywords", order.IncludeKeywords).
		Msg("Processing work order")

	cursors := &sync.Map{}
	// emittedTotal aggregates raw items emitted to the parser topic across
	// all (platform, keyword) goroutines. Used after a successful initial
	// sync to detect "Data365 returned nothing" and fire the empty-batch
	// Pusher event — race-free because we know zero items were produced,
	// so the sink's success event cannot win over it.
	var emittedTotal atomic.Int64
	var fetchAttempted atomic.Bool
	var coverageIncomplete atomic.Bool
	fetchErr := s.fetchAllPlatforms(ctx, order, quotaID, cursors, &emittedTotal, &fetchAttempted, &coverageIncomplete)
	if fetchErr != nil {
		log.Error().Err(fetchErr).Msg("Fetch run failed; leaving progress and sync status untouched so the work order retries")
		return fmt.Errorf("FetcherService.HandleWorkOrder: %w", fetchErr)
	}
	if order.SyncType == "initial" && coverageIncomplete.Load() {
		err := fmt.Errorf("initial sync coverage incomplete")
		log.Warn().Err(err).Msg("Initial sync did not cover every platform/keyword pair; leaving progress and sync status untouched so the work order retries")
		return fmt.Errorf("FetcherService.HandleWorkOrder: %w", err)
	}

	if s.progress != nil {
		fetchedAt := order.ToDate
		if fetchedAt.IsZero() {
			fetchedAt = time.Now().UTC()
		}
		if err := s.progress.UpdateLastFetched(ctx, order.TopicID, fetchedAt, collectCursors(cursors)); err != nil {
			log.Warn().Err(err).Time("fetched_at", fetchedAt).Msg("Failed to advance fetch progress after fetch run")
		}
	}

	if s.syncMarker != nil && order.SyncType == "initial" {
		eventAt := order.ToDate
		if eventAt.IsZero() {
			eventAt = time.Now().UTC()
		}
		applied, err := s.syncMarker.MarkInitialSyncDone(ctx, order.TopicID, order.WorkspaceID, eventAt)
		switch {
		case err != nil:
			log.Error().Err(err).Msg("Failed to mark initial sync done in MongoDB")
		case !applied:
			log.Info().Msg("Initial sync mark skipped: a newer event already landed")
		}

		// Fire the empty-batch Pusher event when this initial sync settled
		// without emitting any items to the parser. We gate on applied=true
		// so a stale work order (newer event already landed) doesn't fire
		// a misleading empty signal after another cycle has succeeded.
		if err == nil && applied && fetchAttempted.Load() && emittedTotal.Load() == 0 {
			s.notifyFirstBatchEmpty(order)
		}
	}

	return nil
}

// notifyFirstBatchEmpty fires the per-workspace Pusher event when an
// initial sync completed without any mentions reaching Kafka. Mirrors
// the sink's notifyFirstBatch swallow-and-log error handling so a Pusher
// hiccup never blocks the data path — the FE has a timeout fallback.
func (s *FetcherService) notifyFirstBatchEmpty(order kafkamodels.ListeningWorkOrder) {
	if s.emptyNotifier == nil {
		return
	}
	if order.WorkspaceID == "" {
		s.log.Warn().
			Str("topic_id", order.TopicID).
			Msg("First-batch-empty notification skipped: empty workspace_id cannot derive Pusher channel")
		return
	}

	channel := emptyBatchChannelPrefix + order.WorkspaceID
	payload := map[string]any{
		"topic_id":     order.TopicID,
		"workspace_id": order.WorkspaceID,
		"settled_at":   time.Now().UTC().Format(time.RFC3339),
		"reason":       "no_matches",
	}

	if err := s.emptyNotifier.Trigger(channel, emptyBatchEvent, payload); err != nil {
		s.log.Warn().Err(err).
			Str("topic_id", order.TopicID).
			Str("workspace_id", order.WorkspaceID).
			Msg("Failed to trigger first-batch-empty Pusher event")
		return
	}

	s.log.Info().
		Str("topic_id", order.TopicID).
		Str("workspace_id", order.WorkspaceID).
		Msg("First-batch-empty Pusher event triggered")
}

func (s *FetcherService) releaseLock(lockKey, token string) {
	releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := s.lock.Release(releaseCtx, lockKey, token); err != nil {
		s.log.Warn().
			Err(err).
			Str("lock_key", lockKey).
			Msg("Failed to release fetch lock")
	}
}

// fetchAllPlatforms fans out fetching across all enabled platforms concurrently.
//
// cursors is a concurrent accumulator: each platform goroutine writes its
// per-keyword resume cursors into it via fetchAndEmitResults. The caller
// (HandleWorkOrder) reads this on success to persist progress.
//
// emittedTotal accumulates the count of raw items emitted to the parser
// topic across all platform goroutines. May be nil; when supplied, callers
// use it after a successful initial sync to detect zero-result runs and
// emit the empty-batch Pusher event.
//
// fetchAttempted flips true once at least one keyword search is actually
// attempted. This lets callers distinguish "provider returned zero items"
// from "we no-op'd before calling Data365" (for example due to exhausted
// quota) so they do not emit a misleading empty-batch event.
//
// coverageIncomplete flips true when a goroutine could not reserve budget from
// the distributed tracker (cross-work-order contention drained Redis between
// pre-flight and this goroutine's reservation). Initial syncs use this to
// retry instead of being falsely marked complete. Consumption of a platform's
// pre-partitioned private slice is normal completion under best-effort and
// does NOT flip this flag — the slice was sized as that platform's fair share
// for the cycle.
func (s *FetcherService) fetchAllPlatforms(ctx context.Context, order kafkamodels.ListeningWorkOrder, quotaID string, cursors *sync.Map, emittedTotal *atomic.Int64, fetchAttempted *atomic.Bool, coverageIncomplete *atomic.Bool) error {
	effectiveMax := s.effectiveMaxPosts(order.SyncType)

	var remaining int64
	hasQuota := s.quota != nil
	switch {
	case s.tracker != nil && s.quota != nil:
		r, err := s.tracker.GetRemaining(ctx, quotaID, func() (int, error) {
			return s.quota.GetRemainingMentionBudget(ctx, order.TopicID, quotaID)
		})
		if err != nil {
			s.log.Error().Err(err).Str("topic_id", order.TopicID).
				Msg("Quota check at platform fan-out failed, skipping (fail-closed)")
			return fmt.Errorf("FetcherService.fetchAllPlatforms: quota check: %w", err)
		}
		remaining = int64(r)
	case s.quota != nil:
		r, err := s.quota.GetRemainingMentionBudget(ctx, order.TopicID, quotaID)
		if err != nil {
			s.log.Error().Err(err).Str("topic_id", order.TopicID).
				Msg("Quota check at platform fan-out failed, skipping (fail-closed)")
			return fmt.Errorf("FetcherService.fetchAllPlatforms: quota check: %w", err)
		}
		remaining = int64(r)
	}

	if hasQuota && remaining <= 0 {
		s.log.Info().Str("topic_id", order.TopicID).
			Msg("Budget exhausted before platform fan-out, skipping")
		return nil
	}

	nPlatforms := int64(len(order.EnabledPlatforms))
	if nPlatforms == 0 {
		return nil
	}

	// Clamp the per-keyword cap to this topic's fair share of the shared
	// mention budget. Without this, a single topic's initial sync (cap=5000)
	// can drain the full super-admin budget on the first WO, leaving sibling
	// topics emitted in the same scheduler tick permanently stuck at
	// is_initial_sync_done=false. Fair-share math:
	//   per-keyword cap = remaining / topicsSharingQuota / nPlatforms / nKeywords
	// Nil counter, empty keywords, missing budget, or 0/1 topics all skip
	// the clamp so the configured cap stays in force.
	nKeywords := int64(len(order.IncludeKeywords))
	if s.quotaTopicCounter != nil && hasQuota && remaining > 0 && nKeywords > 0 {
		topicCount, err := s.quotaTopicCounter.CountActiveTopicsForQuota(ctx, quotaID)
		if err != nil {
			s.log.Warn().Err(err).
				Str("topic_id", order.TopicID).
				Str("quota_id", quotaID).
				Msg("Topic count for fair-share cap failed; using configured cap")
		} else if topicCount > 1 {
			fairSharePerKeyword := remaining / int64(topicCount) / nPlatforms / nKeywords
			if fairSharePerKeyword > 0 && fairSharePerKeyword < int64(effectiveMax) {
				s.log.Info().
					Str("topic_id", order.TopicID).
					Str("quota_id", quotaID).
					Int("topic_count", topicCount).
					Int("configured_cap", effectiveMax).
					Int64("fair_share_cap", fairSharePerKeyword).
					Msg("Clamping per-keyword cap to fair share of shared mention budget")
				effectiveMax = int(fairSharePerKeyword)
			}
		}
	}

	// Pre-partition the workspace mention budget across platforms (best-effort)
	// so goroutines hold private slices and don't race for a shared counter
	// mid-fan-out. Each platform receives floor(remaining/nPlatforms) and the
	// modulo remainder is handed out one-per-platform from the head of
	// EnabledPlatforms so nothing is stranded. When remaining < nPlatforms the
	// trailing platforms get a zero slice and skip this cycle — the scheduler's
	// next pass reseeds budget so coverage round-robins over time. The
	// distributed tracker is still consulted per reservation to coordinate with
	// concurrent work orders under the same quotaID; only intra-work-order
	// contention is removed by the pre-split. When no quota is wired
	// (s.quota == nil), goroutines receive nil localBudgets and reservations
	// pass through unbounded (preserves prior dev/test behaviour).
	var perPlatform, leftover int64
	if hasQuota {
		perPlatform = remaining / nPlatforms
		leftover = remaining - perPlatform*nPlatforms
	}

	var wg sync.WaitGroup
	errCh := make(chan error, len(order.EnabledPlatforms))
	for i, platform := range order.EnabledPlatforms {
		wg.Add(1)
		var lb *int64
		if hasQuota {
			b := perPlatform
			if int64(i) < leftover {
				b++
			}
			lb = &b
		}
		go func(p string, localBudget *int64) {
			defer wg.Done()
			if err := s.fetchPlatform(ctx, order, p, quotaID, localBudget, effectiveMax, cursors, emittedTotal, fetchAttempted, coverageIncomplete); err != nil {
				errCh <- err
			}
		}(platform, lb)
	}
	wg.Wait()
	close(errCh)

	for err := range errCh {
		if err != nil {
			return err
		}
	}

	return nil
}

// fetchPlatform fetches data for all keywords on a single platform.
//
// effectiveMax is the per-keyword max_posts cap for this work order (already
// resolved to the initial-vs-incremental value by the caller). cursors is a
// concurrent accumulator into which fetchAndEmitResults stores the final
// resume cursor per (platform, keyword); it is read by HandleWorkOrder after
// all platforms finish to persist progress for the next cycle.
func (s *FetcherService) fetchPlatform(
	ctx context.Context,
	order kafkamodels.ListeningWorkOrder,
	platform string,
	quotaID string,
	localBudget *int64,
	effectiveMax int,
	cursors *sync.Map,
	emittedTotal *atomic.Int64,
	fetchAttempted *atomic.Bool,
	coverageIncomplete *atomic.Bool,
) error {
	log := s.log.With().
		Str("topic_id", order.TopicID).
		Str("platform", platform).
		Logger()

	for _, keyword := range order.IncludeKeywords {
		kw := prepareKeyword(platform, keyword)
		reserved, ok, err := s.reserveKeywordBudget(ctx, order.TopicID, quotaID, localBudget, effectiveMax)
		if err != nil {
			log.Error().Err(err).Msg("Failed to reserve budget, stopping keyword iteration (fail-closed)")
			return nil
		}
		if !ok {
			// A private-slice exhaustion is normal completion under best-effort
			// partitioning; only a distributed-tracker miss (slice still has
			// budget but Redis said no) signals real cross-work-order contention
			// worth retrying. claimLocalBudget left the slice at 0 in the local
			// case and adjustLocalBudget restored it in the distributed case, so
			// the post-state of localBudget disambiguates the two cleanly.
			exhaustedLocally := localBudget != nil && atomic.LoadInt64(localBudget) <= 0
			if !exhaustedLocally && coverageIncomplete != nil {
				coverageIncomplete.Store(true)
			}
			log.Info().Bool("local_exhausted", exhaustedLocally).Msg("Budget exhausted, stopping keyword iteration")
			return nil
		}
		postsToRequest := int(reserved)
		if fetchAttempted != nil {
			fetchAttempted.Store(true)
		}

		event := log.Info().
			Str("keyword", kw).
			Int("max_posts", postsToRequest).
			Int64("reserved", reserved)
		if localBudget != nil {
			event = event.Int64("local_budget_remaining", atomic.LoadInt64(localBudget))
		}
		event.Msg("Triggering search")

		if err := s.data365.TriggerSearch(ctx, platform, kw, postsToRequest, order.FromDate, order.ToDate, order.Languages); err != nil {
			if relErr := s.releaseReservedBudget(ctx, quotaID, localBudget, reserved); relErr != nil {
				log.Warn().Err(relErr).Msg("Failed to release budget after trigger failure")
			}
			var unsupported *social.UnsupportedSearchError
			if errors.As(err, &unsupported) {
				log.Warn().Err(err).Str("keyword", kw).Msg("Skipping unsupported Data365 search query")
				continue
			}
			log.Error().Err(err).Str("keyword", kw).Msg("Failed to trigger search")
			return fmt.Errorf("FetcherService.fetchPlatform trigger %s/%s: %w", platform, kw, err)
		}

		if err := s.data365.PollUntilFinished(ctx, platform, kw, postsToRequest, order.FromDate, order.ToDate, order.Languages); err != nil {
			log.Error().Err(err).Str("keyword", kw).Msg("Poll failed")
			if relErr := s.releaseReservedBudget(ctx, quotaID, localBudget, reserved); relErr != nil {
				log.Warn().Err(relErr).Msg("Failed to release budget after poll failure")
			}
			return fmt.Errorf("FetcherService.fetchPlatform poll %s/%s: %w", platform, kw, err)
		}

		if err := s.fetchAndEmitResults(ctx, order, platform, kw, quotaID, localBudget, reserved, cursors, emittedTotal); err != nil {
			log.Error().Err(err).Str("keyword", kw).Msg("Failed to fetch/emit results")
			return err
		}
	}

	return nil
}

// fetchAndEmitResults paginates through Data365 results and emits raw payloads to Kafka.
// reserved is the number of mentions pre-reserved for this keyword as an admission-
// control hold. Only the UNUSED portion is released at the end; the emitted portion
// stays held in Redis as a conservative in-flight estimate against raw items that
// haven't yet been dedup-filtered and persisted by the sink. The 5-minute budget
// TTL reconciles against MongoDB truth on each cache miss, so the raw-vs-parsed
// drift is bounded by the TTL rather than accumulating forever.
//
// Cursor resumption: the loop starts from order.Cursors[cursorKey(platform,keyword)]
// (empty string = fresh fetch) and writes the *next* unfetched cursor into the
// cursors accumulator on success. Empty stored value means pagination is
// exhausted; a non-empty value means the next cycle should resume from there.
// Cursors are only persisted by HandleWorkOrder if the entire work order
// succeeds — partial-failure runs leave the prior cycle's cursors untouched.
func (s *FetcherService) fetchAndEmitResults(
	ctx context.Context,
	order kafkamodels.ListeningWorkOrder,
	platform, keyword string,
	quotaID string,
	localBudget *int64,
	reserved int64,
	cursors *sync.Map,
	emittedTotal *atomic.Int64,
) (err error) {
	ckey := cursorKey(platform, keyword)
	cursor := ""
	if order.Cursors != nil {
		cursor = order.Cursors[ckey]
	}
	if cursor != "" {
		s.log.Info().
			Str("topic_id", order.TopicID).
			Str("platform", platform).
			Str("keyword", keyword).
			Msg("Resuming pagination from stored cursor")
	}
	// nextCursor is what we'll persist for the next cycle: the cursor that
	// FetchResults would use next. Empty means "exhausted, no resume."
	nextCursor := cursor

	page := 0
	totalEmitted := int64(0)

	defer func() {
		if emittedTotal != nil {
			emittedTotal.Add(totalEmitted)
		}
		unused := reserved - totalEmitted
		if unused <= 0 {
			return
		}
		if relErr := s.releaseReservedBudget(ctx, quotaID, localBudget, unused); relErr != nil {
			s.log.Warn().Err(relErr).Int64("unused", unused).Msg("Failed to release unused budget back to pool")
		}
	}()

	for {
		remainingReserved := reserved - totalEmitted
		if remainingReserved <= 0 {
			s.log.Info().
				Str("topic_id", order.TopicID).
				Str("platform", platform).
				Int64("reserved", reserved).
				Msg("Reserved budget exhausted during pagination, stopping")
			break
		}

		result, fetchErr := s.data365.FetchResults(ctx, platform, keyword, cursor, order.FromDate, order.ToDate, order.Languages)
		if fetchErr != nil {
			return fmt.Errorf("FetcherService.fetchAndEmitResults: page %d: %w", page, fetchErr)
		}

		if len(result.Data) == 0 || string(result.Data) == "null" {
			nextCursor = ""
			break
		}

		itemCount, empty := countResultItems(result.Data)
		if empty {
			nextCursor = ""
			break
		}

		payload := kafkamodels.ListeningRawPayload{
			TopicID:   order.TopicID,
			Platform:  platform,
			Keyword:   keyword,
			RawData:   result.Data,
			WorkOrder: order,
		}

		data, marshalErr := json.Marshal(payload)
		if marshalErr != nil {
			return fmt.Errorf("FetcherService.fetchAndEmitResults: marshal: %w", marshalErr)
		}

		key := fmt.Sprintf("%s:%s:%s", order.TopicID, platform, keyword)
		if produceErr := s.producer.Produce(ctx, kafkamodels.TopicListeningRaw, []byte(key), data); produceErr != nil {
			return fmt.Errorf("FetcherService.fetchAndEmitResults: produce: %w", produceErr)
		}

		totalEmitted += itemCount
		page++
		// Track the cursor Data365 says is next. If we break here, this is
		// what the next cycle should resume from (or "" if exhausted).
		nextCursor = result.Cursor

		s.log.Debug().
			Str("topic_id", order.TopicID).
			Str("platform", platform).
			Int("page", page).
			Int64("items_on_page", itemCount).
			Int64("reserved_remaining", reserved-totalEmitted).
			Msg("Emitted raw payload page")

		if result.Cursor == "" {
			break
		}
		cursor = result.Cursor
	}

	if cursors != nil {
		cursors.Store(ckey, nextCursor)
	}

	return nil
}

func (s *FetcherService) reserveKeywordBudget(ctx context.Context, topicID, quotaID string, localBudget *int64, maxPosts int) (int64, bool, error) {
	reserved := claimLocalBudget(localBudget, int64(maxPosts))
	if reserved <= 0 {
		return 0, false, nil
	}
	if s.tracker == nil {
		return reserved, true, nil
	}

	actualReserved, ok, err := s.reserveDistributedBudget(ctx, topicID, quotaID, reserved)
	if err != nil {
		_, _ = s.adjustLocalBudget(localBudget, reserved)
		return 0, false, err
	}
	if !ok {
		_, _ = s.adjustLocalBudget(localBudget, reserved)
		return 0, false, nil
	}
	if actualReserved < reserved {
		_, _ = s.adjustLocalBudget(localBudget, reserved-actualReserved)
	}

	return actualReserved, true, nil
}

func (s *FetcherService) reserveDistributedBudget(ctx context.Context, topicID, quotaID string, requested int64) (int64, bool, error) {
	reseed := s.budgetReseedFallback(ctx, topicID, quotaID)

	newBudget, ok, err := s.tracker.Reserve(ctx, quotaID, requested, reseed)
	if err != nil {
		return 0, false, err
	}
	if ok {
		return requested, true, nil
	}
	if newBudget <= 0 {
		return 0, false, nil
	}

	fallback := newBudget
	_, ok, err = s.tracker.Reserve(ctx, quotaID, fallback, reseed)
	if err != nil {
		return 0, false, err
	}
	if !ok {
		return 0, false, nil
	}

	return fallback, true, nil
}

// budgetReseedFallback builds the closure DistributedQuotaTracker.Reserve uses
// to repopulate Redis from MongoDB when the budget key has expired mid-fan-out.
// Returns nil when the QuotaChecker is not wired up so Reserve falls back to
// the original fail-closed behaviour rather than panicking on a nil deref.
func (s *FetcherService) budgetReseedFallback(ctx context.Context, topicID, quotaID string) func() (int, error) {
	if s.quota == nil {
		return nil
	}
	return func() (int, error) {
		return s.quota.GetRemainingMentionBudget(ctx, topicID, quotaID)
	}
}

func (s *FetcherService) releaseReservedBudget(ctx context.Context, quotaID string, localBudget *int64, amount int64) error {
	if amount <= 0 {
		return nil
	}

	// Release the shared Redis hold first — it's the cross-worker authoritative
	// state. A local counter that's only meaningful within this work order
	// must not block the shared release.
	if s.tracker != nil {
		if err := s.tracker.Release(ctx, quotaID, amount); err != nil {
			return err
		}
	}
	s.adjustLocalBudget(localBudget, amount)
	return nil
}

func claimLocalBudget(localBudget *int64, requested int64) int64 {
	if requested <= 0 {
		return 0
	}
	if localBudget == nil {
		return requested
	}

	for {
		current := atomic.LoadInt64(localBudget)
		if current <= 0 {
			return 0
		}

		claim := requested
		if claim > current {
			claim = current
		}
		if atomic.CompareAndSwapInt64(localBudget, current, current-claim) {
			return claim
		}
	}
}

func (s *FetcherService) adjustLocalBudget(localBudget *int64, delta int64) (int64, bool) {
	if localBudget == nil || delta == 0 {
		return 0, true
	}
	newBudget := atomic.AddInt64(localBudget, delta)
	return newBudget, true
}

func countResultItems(data json.RawMessage) (int64, bool) {
	var envelope struct {
		Items []json.RawMessage `json:"items"`
	}
	if err := json.Unmarshal(data, &envelope); err == nil && envelope.Items != nil {
		return int64(len(envelope.Items)), len(envelope.Items) == 0
	}

	var items []json.RawMessage
	if err := json.Unmarshal(data, &items); err == nil {
		return int64(len(items)), len(items) == 0
	}

	return 1, false
}

// prepareKeyword adjusts keywords per platform rules.
// Instagram only supports hashtag search, so auto-prepend # if missing.
func prepareKeyword(platform, keyword string) string {
	if platform == "instagram" && !strings.HasPrefix(keyword, "#") {
		return "#" + keyword
	}
	return keyword
}
