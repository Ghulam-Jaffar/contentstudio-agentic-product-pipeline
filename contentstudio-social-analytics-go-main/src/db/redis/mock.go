package redis

import (
	"context"
	"time"
)

// MockRedisClient is a mock implementation of Redis client for testing.
type MockRedisClient struct {
	GetFunc              func(ctx context.Context, key string) (string, error)
	SetFunc              func(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	SetNXFunc            func(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error)
	DelFunc              func(ctx context.Context, keys ...string) error
	CompareAndDeleteFunc func(ctx context.Context, key, expected string) (bool, error)
	DecrByFunc           func(ctx context.Context, key string, amount int64) (int64, error)
	DecrByIfPositiveFunc func(ctx context.Context, key string, amount int64) (int64, bool, error)
	IncrByFunc           func(ctx context.Context, key string, amount int64) (int64, error)
	IncrByIfExistsFunc   func(ctx context.Context, key string, amount int64) (int64, bool, error)
	DecrByIfExistsFunc   func(ctx context.Context, key string, amount int64) (int64, bool, error)
	ExpireFunc           func(ctx context.Context, key string, expiration time.Duration) (bool, error)
	ExistsFunc           func(ctx context.Context, keys ...string) (int64, error)
	SRandMemberFunc      func(ctx context.Context, key string) (string, error)
	SAddFunc             func(ctx context.Context, key string, members ...interface{}) error
	SMembersFunc         func(ctx context.Context, key string) ([]string, error)
	PingFunc             func(ctx context.Context) error
	CloseFunc            func() error
}

func (m *MockRedisClient) Get(ctx context.Context, key string) (string, error) {
	if m.GetFunc != nil {
		return m.GetFunc(ctx, key)
	}
	return "", nil
}

func (m *MockRedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	if m.SetFunc != nil {
		return m.SetFunc(ctx, key, value, expiration)
	}
	return nil
}

func (m *MockRedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	if m.SetNXFunc != nil {
		return m.SetNXFunc(ctx, key, value, expiration)
	}
	return true, nil
}

func (m *MockRedisClient) Del(ctx context.Context, keys ...string) error {
	if m.DelFunc != nil {
		return m.DelFunc(ctx, keys...)
	}
	return nil
}

func (m *MockRedisClient) CompareAndDelete(ctx context.Context, key, expected string) (bool, error) {
	if m.CompareAndDeleteFunc != nil {
		return m.CompareAndDeleteFunc(ctx, key, expected)
	}
	return true, nil
}

func (m *MockRedisClient) DecrBy(ctx context.Context, key string, amount int64) (int64, error) {
	if m.DecrByFunc != nil {
		return m.DecrByFunc(ctx, key, amount)
	}
	return 0, nil
}

func (m *MockRedisClient) DecrByIfPositive(ctx context.Context, key string, amount int64) (int64, bool, error) {
	if m.DecrByIfPositiveFunc != nil {
		return m.DecrByIfPositiveFunc(ctx, key, amount)
	}
	return 0, false, nil
}

func (m *MockRedisClient) IncrBy(ctx context.Context, key string, amount int64) (int64, error) {
	if m.IncrByFunc != nil {
		return m.IncrByFunc(ctx, key, amount)
	}
	return 0, nil
}

func (m *MockRedisClient) IncrByIfExists(ctx context.Context, key string, amount int64) (int64, bool, error) {
	if m.IncrByIfExistsFunc != nil {
		return m.IncrByIfExistsFunc(ctx, key, amount)
	}
	return 0, false, nil
}

func (m *MockRedisClient) DecrByIfExists(ctx context.Context, key string, amount int64) (int64, bool, error) {
	if m.DecrByIfExistsFunc != nil {
		return m.DecrByIfExistsFunc(ctx, key, amount)
	}
	return 0, false, nil
}

func (m *MockRedisClient) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	if m.ExpireFunc != nil {
		return m.ExpireFunc(ctx, key, expiration)
	}
	return true, nil
}

func (m *MockRedisClient) Exists(ctx context.Context, keys ...string) (int64, error) {
	if m.ExistsFunc != nil {
		return m.ExistsFunc(ctx, keys...)
	}
	return 0, nil
}

func (m *MockRedisClient) SRandMember(ctx context.Context, key string) (string, error) {
	if m.SRandMemberFunc != nil {
		return m.SRandMemberFunc(ctx, key)
	}
	return "", nil
}

func (m *MockRedisClient) SAdd(ctx context.Context, key string, members ...interface{}) error {
	if m.SAddFunc != nil {
		return m.SAddFunc(ctx, key, members...)
	}
	return nil
}

func (m *MockRedisClient) SMembers(ctx context.Context, key string) ([]string, error) {
	if m.SMembersFunc != nil {
		return m.SMembersFunc(ctx, key)
	}
	return nil, nil
}

func (m *MockRedisClient) Ping(ctx context.Context) error {
	if m.PingFunc != nil {
		return m.PingFunc(ctx)
	}
	return nil
}

func (m *MockRedisClient) Close() error {
	if m.CloseFunc != nil {
		return m.CloseFunc()
	}
	return nil
}
