package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

const releaseTimeout = 5 * time.Second

// DistributedLock provides a Redis-based distributed lock using SetNX with
// unique token ownership. Only the holder of the token can release the lock,
// preventing races where an expired lease is released by a stale worker.
type DistributedLock struct {
	client lockClient
	logger zerolog.Logger
}

type lockClient interface {
	SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	CompareAndDelete(ctx context.Context, key, expected string) (bool, error)
}

// NewDistributedLock creates a new distributed lock backed by the given Redis client.
func NewDistributedLock(client lockClient, logger zerolog.Logger) *DistributedLock {
	return &DistributedLock{
		client: client,
		logger: logger.With().Str("component", "distributed_lock").Logger(),
	}
}

// Acquire attempts to acquire a lock for the given key with the specified TTL.
// Returns a non-empty token if the lock was acquired, or an empty string if it is already held.
// The token must be passed to Release to prove ownership.
func (l *DistributedLock) Acquire(ctx context.Context, key string, ttl time.Duration) (string, error) {
	token := uuid.New().String()

	acquired, err := l.client.SetNX(ctx, key, token, ttl)
	if err != nil {
		return "", fmt.Errorf("DistributedLock.Acquire: failed to set lock for key %s: %w", key, err)
	}

	if acquired {
		l.logger.Debug().Str("key", key).Str("token", token).Dur("ttl", ttl).Msg("Lock acquired")
		return token, nil
	}

	l.logger.Debug().Str("key", key).Msg("Lock already held")
	return "", nil
}

// Release removes the lock for the given key only if the caller holds the matching token.
// This prevents a stale worker from releasing a lock that was re-acquired by another worker.
//
// The Redis call is detached from the caller ctx and given a fresh timeout, because
// Release is typically invoked via defer on shutdown paths where the caller ctx has
// already been cancelled — propagating that cancellation would leave the lock stuck
// in Redis until its TTL expired, blocking the next worker for that key.
func (l *DistributedLock) Release(_ context.Context, key, token string) error {
	ctx, cancel := context.WithTimeout(context.Background(), releaseTimeout)
	defer cancel()

	released, err := l.client.CompareAndDelete(ctx, key, token)
	if err != nil {
		return fmt.Errorf("DistributedLock.Release: compare-and-delete failed for key %s: %w", key, err)
	}

	if !released {
		l.logger.Warn().Str("key", key).Msg("Lock token mismatch, not releasing (expired or re-acquired)")
		return nil
	}

	l.logger.Debug().Str("key", key).Msg("Lock released")
	return nil
}
