package quota

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/redis"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// TestDistributedQuotaTrackerReserve covers every Reserve outcome on the
// missing-key code path: nil-fallback fail-closed, reseed+retry happy path,
// reseed yielding exhausted budget (insufficient, no error), pathological
// still-missing-after-reseed (fail-loud), SetNX race-lost (succeed against
// winner's value), and wrapped fallback errors.
//
// Each row programs a per-call DecrByIfPositive sequence plus a SetNX result.
// "want" fields capture both the API contract (ok/budget/err) and observable
// mock-call counts so regressions in the retry semantics surface as test
// failures rather than silent behaviour drift.
func TestDistributedQuotaTrackerReserve(t *testing.T) {
	t.Parallel()

	dbErr := errors.New("mongo timeout")

	type decrResult struct {
		budget int64
		ok     bool
	}

	cases := []struct {
		name            string
		decrSequence    []decrResult
		setNXOK         bool // result returned by SetNX (only consulted when SetNX is hit)
		fallback        func() (int, error)
		amount          int64
		wantOK          bool
		wantBudget      int64
		wantErrContains string
		wantErrIs       error
		wantDecrCalls   int64
		wantSetNXCalls  int64
		wantSetNXValue  int64 // 0 = don't assert
		assertSetNXTTL  bool
	}{
		{
			name:            "nil fallback fails closed when budget key missing",
			decrSequence:    []decrResult{{-1, false}},
			fallback:        nil,
			amount:          10,
			wantErrContains: "budget key missing",
			wantDecrCalls:   1,
		},
		{
			name:           "reseeds via SetNX and retries when key missing",
			decrSequence:   []decrResult{{-1, false}, {450, true}},
			setNXOK:        true,
			fallback:       func() (int, error) { return 500, nil },
			amount:         50,
			wantOK:         true,
			wantBudget:     450,
			wantDecrCalls:  2,
			wantSetNXCalls: 1,
			wantSetNXValue: 500,
			assertSetNXTTL: true,
		},
		{
			name:           "returns insufficient when reseeded budget is exhausted",
			decrSequence:   []decrResult{{-1, false}, {0, false}},
			setNXOK:        true,
			fallback:       func() (int, error) { return 0, nil },
			amount:         50,
			wantBudget:     0,
			wantDecrCalls:  2,
			wantSetNXCalls: 1,
		},
		{
			name:            "fails closed when key still missing after reseed",
			decrSequence:    []decrResult{{-1, false}, {-1, false}},
			setNXOK:         true,
			fallback:        func() (int, error) { return 500, nil },
			amount:          50,
			wantErrContains: "still missing after reseed",
			wantDecrCalls:   2,
			wantSetNXCalls:  1,
		},
		{
			name:           "succeeds against winner's value when SetNX race lost",
			decrSequence:   []decrResult{{-1, false}, {750, true}},
			setNXOK:        false,
			fallback:       func() (int, error) { return 500, nil },
			amount:         50,
			wantOK:         true,
			wantBudget:     750,
			wantDecrCalls:  2,
			wantSetNXCalls: 1,
		},
		{
			name:          "wraps fallback errors via errors.Is",
			decrSequence:  []decrResult{{-1, false}},
			fallback:      func() (int, error) { return 0, dbErr },
			amount:        50,
			wantErrIs:     dbErr,
			wantDecrCalls: 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			log := logger.NewNop()
			var decrCalls, setNXCalls, setNXValue, setNXTTL atomic.Int64

			rdb := &redis.MockRedisClient{
				DecrByIfPositiveFunc: func(_ context.Context, _ string, _ int64) (int64, bool, error) {
					n := decrCalls.Add(1)
					idx := int(n - 1)
					if idx >= len(tc.decrSequence) {
						idx = len(tc.decrSequence) - 1 // saturate at last entry for safety
					}
					r := tc.decrSequence[idx]
					return r.budget, r.ok, nil
				},
				SetNXFunc: func(_ context.Context, _ string, value interface{}, ttl time.Duration) (bool, error) {
					setNXCalls.Add(1)
					if v, ok := value.(int64); ok {
						setNXValue.Store(v)
					}
					setNXTTL.Store(int64(ttl))
					return tc.setNXOK, nil
				},
			}

			tracker := NewDistributedQuotaTracker(rdb, log)
			newBudget, ok, err := tracker.Reserve(context.Background(), "sa-1", tc.amount, tc.fallback)

			switch {
			case tc.wantErrContains != "":
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tc.wantErrContains)
				}
				if !strings.Contains(err.Error(), tc.wantErrContains) {
					t.Fatalf("expected error containing %q, got %v", tc.wantErrContains, err)
				}
			case tc.wantErrIs != nil:
				if !errors.Is(err, tc.wantErrIs) {
					t.Fatalf("expected error to wrap %v, got %v", tc.wantErrIs, err)
				}
			default:
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
			}

			if ok != tc.wantOK {
				t.Fatalf("expected ok=%v, got %v", tc.wantOK, ok)
			}
			if newBudget != tc.wantBudget {
				t.Fatalf("expected newBudget=%d, got %d", tc.wantBudget, newBudget)
			}

			if got := decrCalls.Load(); got != tc.wantDecrCalls {
				t.Fatalf("expected %d DecrByIfPositive calls, got %d", tc.wantDecrCalls, got)
			}
			if got := setNXCalls.Load(); got != tc.wantSetNXCalls {
				t.Fatalf("expected %d SetNX calls, got %d", tc.wantSetNXCalls, got)
			}
			if tc.wantSetNXValue != 0 {
				if got := setNXValue.Load(); got != tc.wantSetNXValue {
					t.Fatalf("expected SetNX value=%d, got %d", tc.wantSetNXValue, got)
				}
			}
			if tc.assertSetNXTTL {
				if got := setNXTTL.Load(); got != int64(defaultBudgetTTL) {
					t.Fatalf("expected SetNX TTL=%v, got %v", defaultBudgetTTL, time.Duration(got))
				}
			}
		})
	}
}

// Regression: a late Release firing after the 5-min budget TTL has expired
// must not revive the key without a TTL. Pre-fix the IncrBy path created a
// permanent value=0/TTL=-1 zombie; post-fix Release no-ops and the next
// GetRemaining reseeds the key from MongoDB on cache miss.
func TestDistributedQuotaTrackerReleaseIsNoOpWhenBudgetKeyExpired(t *testing.T) {
	t.Parallel()

	log := logger.NewNop()
	var incrCalls, incrIfExistsCalls atomic.Int64
	rdb := &redis.MockRedisClient{
		IncrByFunc: func(_ context.Context, _ string, _ int64) (int64, error) {
			incrCalls.Add(1)
			return 0, nil
		},
		IncrByIfExistsFunc: func(_ context.Context, _ string, _ int64) (int64, bool, error) {
			incrIfExistsCalls.Add(1)
			return -1, false, nil
		},
	}

	tracker := NewDistributedQuotaTracker(rdb, log)
	if err := tracker.Release(context.Background(), "sa-1", 25); err != nil {
		t.Fatalf("Release returned unexpected error on missing key: %v", err)
	}
	if got := incrCalls.Load(); got != 0 {
		t.Fatalf("Release must not call plain IncrBy (would revive key without TTL); got %d calls", got)
	}
	if got := incrIfExistsCalls.Load(); got != 1 {
		t.Fatalf("expected exactly one IncrByIfExists call, got %d", got)
	}
}

// Happy path: when the budget key still exists, Release returns nil and the
// IncrByIfExists wrapper carries the new budget through without error.
func TestDistributedQuotaTrackerReleaseSucceedsWhenBudgetKeyExists(t *testing.T) {
	t.Parallel()

	log := logger.NewNop()
	rdb := &redis.MockRedisClient{
		IncrByIfExistsFunc: func(_ context.Context, _ string, amount int64) (int64, bool, error) {
			return 50 + amount, true, nil
		},
	}

	tracker := NewDistributedQuotaTracker(rdb, log)
	if err := tracker.Release(context.Background(), "sa-1", 25); err != nil {
		t.Fatalf("Release returned unexpected error on present key: %v", err)
	}
}

// Regression: a Debit firing after TTL expiry must also no-op rather than
// auto-create the key with no TTL via plain DECRBY.
func TestDistributedQuotaTrackerDebitIsNoOpWhenBudgetKeyExpired(t *testing.T) {
	t.Parallel()

	log := logger.NewNop()
	var decrCalls, decrIfExistsCalls atomic.Int64
	rdb := &redis.MockRedisClient{
		DecrByFunc: func(_ context.Context, _ string, _ int64) (int64, error) {
			decrCalls.Add(1)
			return 0, nil
		},
		DecrByIfExistsFunc: func(_ context.Context, _ string, _ int64) (int64, bool, error) {
			decrIfExistsCalls.Add(1)
			return -1, false, nil
		},
	}

	tracker := NewDistributedQuotaTracker(rdb, log)
	newBudget, err := tracker.Debit(context.Background(), "sa-1", 1)
	if err != nil {
		t.Fatalf("Debit returned unexpected error on missing key: %v", err)
	}
	if newBudget != 0 {
		t.Fatalf("expected newBudget=0 on missing-key no-op, got %d", newBudget)
	}
	if got := decrCalls.Load(); got != 0 {
		t.Fatalf("Debit must not call plain DecrBy; got %d calls", got)
	}
	if got := decrIfExistsCalls.Load(); got != 1 {
		t.Fatalf("expected exactly one DecrByIfExists call, got %d", got)
	}
}

// Happy path: when the budget key still exists, Debit returns the new budget
// from the DecrByIfExists wrapper.
func TestDistributedQuotaTrackerDebitSucceedsWhenBudgetKeyExists(t *testing.T) {
	t.Parallel()

	log := logger.NewNop()
	rdb := &redis.MockRedisClient{
		DecrByIfExistsFunc: func(_ context.Context, _ string, amount int64) (int64, bool, error) {
			return 100 - amount, true, nil
		},
	}

	tracker := NewDistributedQuotaTracker(rdb, log)
	newBudget, err := tracker.Debit(context.Background(), "sa-1", 7)
	if err != nil {
		t.Fatalf("Debit returned unexpected error on present key: %v", err)
	}
	if newBudget != 93 {
		t.Fatalf("expected newBudget=93 on successful debit, got %d", newBudget)
	}
}
