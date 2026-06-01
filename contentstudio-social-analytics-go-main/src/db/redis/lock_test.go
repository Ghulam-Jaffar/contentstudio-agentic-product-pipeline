package redis

import (
	"context"
	"fmt"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockRedisClient implements the Client interface for testing.
type mockRedisClient struct {
	store map[string]string
}

func newMockRedisClient() *mockRedisClient {
	return &mockRedisClient{store: make(map[string]string)}
}

func (m *mockRedisClient) Get(_ context.Context, key string) (string, error) {
	val, ok := m.store[key]
	if !ok {
		return "", nil
	}
	return val, nil
}

func (m *mockRedisClient) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	m.store[key] = fmt.Sprintf("%v", value)
	return nil
}

func (m *mockRedisClient) SetNX(_ context.Context, key string, value interface{}, _ time.Duration) (bool, error) {
	if _, exists := m.store[key]; exists {
		return false, nil
	}
	m.store[key] = fmt.Sprintf("%v", value)
	return true, nil
}

func (m *mockRedisClient) Del(_ context.Context, keys ...string) error {
	for _, k := range keys {
		delete(m.store, k)
	}
	return nil
}

func (m *mockRedisClient) DecrBy(_ context.Context, key string, amount int64) (int64, error) {
	var current int64
	if val, ok := m.store[key]; ok {
		fmt.Sscanf(val, "%d", &current)
	}
	newVal := current - amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, nil
}

func (m *mockRedisClient) DecrByIfPositive(_ context.Context, key string, amount int64) (int64, bool, error) {
	var current int64
	val, ok := m.store[key]
	if !ok {
		return -1, false, nil
	}
	fmt.Sscanf(val, "%d", &current)
	if current < amount {
		return current, false, nil
	}
	newVal := current - amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, true, nil
}

func (m *mockRedisClient) IncrBy(_ context.Context, key string, amount int64) (int64, error) {
	var current int64
	if val, ok := m.store[key]; ok {
		fmt.Sscanf(val, "%d", &current)
	}
	newVal := current + amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, nil
}

func (m *mockRedisClient) Expire(_ context.Context, key string, _ time.Duration) (bool, error) {
	return true, nil
}

func (m *mockRedisClient) CompareAndDelete(_ context.Context, key, expected string) (bool, error) {
	current, ok := m.store[key]
	if !ok || current != expected {
		return false, nil
	}

	delete(m.store, key)
	return true, nil
}

func (m *mockRedisClient) Close() error {
	return nil
}

func newTestLogger() zerolog.Logger {
	return zerolog.New(io.Discard)
}

func TestDistributedLock_AcquireAndRelease(t *testing.T) {
	t.Parallel()

	mock := newMockRedisClient()
	lock := NewDistributedLock(mock, newTestLogger())
	ctx := context.Background()
	key := "test:lock:topic1"
	ttl := 30 * time.Minute

	// First acquire should succeed and return a token.
	token, err := lock.Acquire(ctx, key, ttl)
	require.NoError(t, err)
	assert.NotEmpty(t, token)

	// Second acquire on same key should fail.
	token2, err := lock.Acquire(ctx, key, ttl)
	require.NoError(t, err)
	assert.Empty(t, token2)

	// Release with correct token should succeed.
	err = lock.Release(ctx, key, token)
	require.NoError(t, err)

	// After release, acquire should succeed again.
	token3, err := lock.Acquire(ctx, key, ttl)
	require.NoError(t, err)
	assert.NotEmpty(t, token3)
}

func TestDistributedLock_IndependentKeys(t *testing.T) {
	t.Parallel()

	mock := newMockRedisClient()
	lock := NewDistributedLock(mock, newTestLogger())
	ctx := context.Background()
	ttl := 30 * time.Minute

	// Acquiring different keys should both succeed.
	token1, err := lock.Acquire(ctx, "lock:a", ttl)
	require.NoError(t, err)
	assert.NotEmpty(t, token1)

	token2, err := lock.Acquire(ctx, "lock:b", ttl)
	require.NoError(t, err)
	assert.NotEmpty(t, token2)

	// Release one should not affect the other.
	err = lock.Release(ctx, "lock:a", token1)
	require.NoError(t, err)

	tokenB, err := lock.Acquire(ctx, "lock:b", ttl)
	require.NoError(t, err)
	assert.Empty(t, tokenB) // still held
}

func TestDistributedLock_WrongTokenCannotRelease(t *testing.T) {
	t.Parallel()

	mock := newMockRedisClient()
	lock := NewDistributedLock(mock, newTestLogger())
	ctx := context.Background()
	key := "test:lock:safety"
	ttl := 30 * time.Minute

	// Worker A acquires the lock.
	tokenA, err := lock.Acquire(ctx, key, ttl)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenA)

	// Worker B tries to release with a wrong token — should NOT delete.
	err = lock.Release(ctx, key, "wrong-token")
	require.NoError(t, err)

	// Lock should still be held — acquire fails.
	tokenRetry, err := lock.Acquire(ctx, key, ttl)
	require.NoError(t, err)
	assert.Empty(t, tokenRetry, "lock should still be held after wrong-token release")

	// Original owner can still release.
	err = lock.Release(ctx, key, tokenA)
	require.NoError(t, err)

	tokenFinal, err := lock.Acquire(ctx, key, ttl)
	require.NoError(t, err)
	assert.NotEmpty(t, tokenFinal, "lock should be acquirable after correct release")
}
