package quota

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

const (
	budgetKeyFormat  = "listening:budget:%s"
	defaultBudgetTTL = 5 * time.Minute
)

// DistributedQuotaTracker backs the per-owner mention budget with Redis so
// multiple fetcher replicas share a single remaining-budget counter. The
// fetcher Reserves maxPosts before a keyword fetch and Releases the UNUSED
// portion after — the emitted portion stays held as a conservative admission
// estimate against in-flight raw items that haven't reached the sink yet.
// Because raw emits include duplicates and off-language items that the sink
// never persists, Redis drifts below workspace.used_mention_credits over time;
// the short TTL forces reconciliation from MongoDB on each cache miss so the
// drift window is bounded rather than cumulative.
type DistributedQuotaTracker struct {
	redis quotaStore
	log   *logger.Logger
}

type quotaStore interface {
	Get(ctx context.Context, key string) (string, error)
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	IncrByIfExists(ctx context.Context, key string, amount int64) (int64, bool, error)
	DecrByIfExists(ctx context.Context, key string, amount int64) (int64, bool, error)
	DecrByIfPositive(ctx context.Context, key string, amount int64) (int64, bool, error)
}

func NewDistributedQuotaTracker(rdb quotaStore, log *logger.Logger) *DistributedQuotaTracker {
	return &DistributedQuotaTracker{
		redis: rdb,
		log:   log,
	}
}

// GetRemaining returns the currently available mention budget for a SuperAdmin/Workspace.
// If the budget is not in Redis, it initializes it using the provided fallback function.
// Fail-closed: returns error on Redis failures so the caller can skip the work order.
func (t *DistributedQuotaTracker) GetRemaining(ctx context.Context, quotaID string, fallback func() (int, error)) (int64, error) {
	if quotaID == "" {
		return 0, fmt.Errorf("DistributedQuotaTracker.GetRemaining: missing quotaID")
	}

	key := fmt.Sprintf(budgetKeyFormat, quotaID)
	val, err := t.redis.Get(ctx, key)
	if err != nil {
		return 0, fmt.Errorf("DistributedQuotaTracker.GetRemaining: redis get: %w", err)
	}

	if val != "" {
		remaining, err := strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("DistributedQuotaTracker.GetRemaining: parse error: %w", err)
		}
		return remaining, nil
	}

	// Key not found — initialize from fallback (DB).
	remainingInt, err := fallback()
	if err != nil {
		return 0, fmt.Errorf("DistributedQuotaTracker.GetRemaining: fallback failed: %w", err)
	}

	remaining := int64(remainingInt)
	ok, err := t.redis.SetNX(ctx, key, remaining, defaultBudgetTTL)
	if err != nil {
		return 0, fmt.Errorf("DistributedQuotaTracker.GetRemaining: SetNX failed: %w", err)
	}

	if !ok {
		// Another worker initialized the key first. Read the authoritative Redis value
		// instead of returning our potentially stale DB value.
		val, err = t.redis.Get(ctx, key)
		if err != nil {
			return 0, fmt.Errorf("DistributedQuotaTracker.GetRemaining: re-read after SetNX race: %w", err)
		}
		remaining, err = strconv.ParseInt(val, 10, 64)
		if err != nil {
			return 0, fmt.Errorf("DistributedQuotaTracker.GetRemaining: parse re-read: %w", err)
		}
	}

	return remaining, nil
}

// Reserve atomically reserves the requested amount from the shared budget.
// Returns (newBudget, true, nil) if the reservation succeeded.
// Returns (currentBudget, false, nil) if insufficient budget.
//
// If the budget key is missing (TTL expired between the work order's initial
// GetRemaining and a later Reserve in the same fan-out), Reserve will reseed
// from fallback exactly once and retry the decrement. A nil fallback preserves
// the original fail-closed behaviour for callers without a DB source.
func (t *DistributedQuotaTracker) Reserve(ctx context.Context, quotaID string, amount int64, fallback func() (int, error)) (int64, bool, error) {
	if quotaID == "" || amount <= 0 {
		return 0, false, nil
	}

	key := fmt.Sprintf(budgetKeyFormat, quotaID)
	newBudget, ok, err := t.redis.DecrByIfPositive(ctx, key, amount)
	if err != nil {
		return 0, false, fmt.Errorf("DistributedQuotaTracker.Reserve: %w", err)
	}

	if !ok && newBudget == -1 {
		// Budget key missing — reseed from fallback once, then retry.
		newBudget, ok, err = t.reseedAndRetryReserve(ctx, key, quotaID, amount, fallback)
		if err != nil {
			return 0, false, err
		}
	}

	if !ok {
		t.log.Info().
			Str("quota_id", quotaID).
			Int64("requested", amount).
			Int64("available", newBudget).
			Msg("Insufficient distributed budget for reservation")
		return newBudget, false, nil
	}

	t.log.Debug().
		Str("quota_id", quotaID).
		Int64("reserved", amount).
		Int64("new_budget", newBudget).
		Msg("Reserved distributed mention budget")

	return newBudget, true, nil
}

// reseedAndRetryReserve handles the missing-key path for Reserve. It seeds the
// budget key from the DB fallback (losing a SetNX race is fine — another
// worker won), then retries the decrement exactly once. A second missing-key
// result after reseed is pathological (TTL fired immediately or fallback
// returned a non-positive value) and is reported as an error so the caller
// can fail-closed.
func (t *DistributedQuotaTracker) reseedAndRetryReserve(ctx context.Context, key, quotaID string, amount int64, fallback func() (int, error)) (int64, bool, error) {
	if fallback == nil {
		return 0, false, fmt.Errorf("DistributedQuotaTracker.Reserve: budget key missing for quota %s (no fallback available)", quotaID)
	}

	remainingInt, err := fallback()
	if err != nil {
		return 0, false, fmt.Errorf("DistributedQuotaTracker.Reserve: reseed fallback failed: %w", err)
	}
	remaining := int64(remainingInt)

	if _, err := t.redis.SetNX(ctx, key, remaining, defaultBudgetTTL); err != nil {
		return 0, false, fmt.Errorf("DistributedQuotaTracker.Reserve: reseed SetNX failed: %w", err)
	}

	t.log.Warn().
		Str("quota_id", quotaID).
		Int64("reseeded_budget", remaining).
		Int64("requested", amount).
		Msg("Budget key missing on reserve, reseeded from MongoDB and retrying (TTL likely too short for fetch cycle)")

	newBudget, ok, err := t.redis.DecrByIfPositive(ctx, key, amount)
	if err != nil {
		return 0, false, fmt.Errorf("DistributedQuotaTracker.Reserve: retry after reseed: %w", err)
	}
	if !ok && newBudget == -1 {
		return 0, false, fmt.Errorf("DistributedQuotaTracker.Reserve: budget key still missing after reseed for quota %s", quotaID)
	}

	return newBudget, ok, nil
}

// Release returns unused reserved budget back to the shared pool.
// If the budget key has already expired (e.g. a fetcher's defer fires after the
// 5-min TTL), the release is a no-op rather than reviving the key without TTL.
// The next GetRemaining will reseed from MongoDB on cache miss anyway, so the
// dropped release is harmless: workspace.used_mention_credits is the source of
// truth and already accounts for any persisted mentions.
func (t *DistributedQuotaTracker) Release(ctx context.Context, quotaID string, amount int64) error {
	if quotaID == "" || amount <= 0 {
		return nil
	}

	key := fmt.Sprintf(budgetKeyFormat, quotaID)
	newBudget, ok, err := t.redis.IncrByIfExists(ctx, key, amount)
	if err != nil {
		return fmt.Errorf("DistributedQuotaTracker.Release: %w", err)
	}
	if !ok {
		t.log.Info().
			Str("quota_id", quotaID).
			Int64("released", amount).
			Msg("Budget key already expired, skipping release (will reseed from MongoDB)")
		return nil
	}

	t.log.Debug().
		Str("quota_id", quotaID).
		Int64("released", amount).
		Int64("new_budget", newBudget).
		Msg("Released unused distributed mention budget")

	return nil
}

// Debit atomically reduces the remaining budget in Redis. This is the
// authoritative accounting hook, called by the sink when a parsed mention is
// persisted, so the Redis counter stays in sync with workspace.used_mention_credits.
// If the key has expired, the debit is a no-op — the next GetRemaining will
// reseed from MongoDB, which already reflects the persisted mention.
func (t *DistributedQuotaTracker) Debit(ctx context.Context, quotaID string, amount int64) (int64, error) {
	if quotaID == "" || amount <= 0 {
		return 0, nil
	}

	key := fmt.Sprintf(budgetKeyFormat, quotaID)
	newBudget, ok, err := t.redis.DecrByIfExists(ctx, key, amount)
	if err != nil {
		return 0, fmt.Errorf("DistributedQuotaTracker.Debit: %w", err)
	}
	if !ok {
		t.log.Info().
			Str("quota_id", quotaID).
			Int64("debited", amount).
			Msg("Budget key already expired, skipping debit (will reseed from MongoDB)")
		return 0, nil
	}

	t.log.Debug().
		Str("quota_id", quotaID).
		Int64("debited", amount).
		Int64("new_budget", newBudget).
		Msg("Debited distributed mention budget")

	return newBudget, nil
}
