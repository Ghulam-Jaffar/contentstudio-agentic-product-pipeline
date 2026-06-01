package sentiment

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type mockProducerRecorder struct {
	messages []producedMsg
	mu       sync.Mutex
	err      error
}

type producedMsg struct {
	Topic string
	Key   string
	Value []byte
}

func (m *mockProducerRecorder) Produce(_ context.Context, topic string, key, value []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.err != nil {
		return m.err
	}
	m.messages = append(m.messages, producedMsg{Topic: topic, Key: string(key), Value: value})
	return nil
}

func (m *mockProducerRecorder) Close() error { return nil }

type mockRedisForLock struct {
	store map[string]string
	mu    sync.Mutex
}

func newMockRedisForLock() *mockRedisForLock {
	return &mockRedisForLock{store: make(map[string]string)}
}

func (m *mockRedisForLock) Get(_ context.Context, key string) (string, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.store[key], nil
}
func (m *mockRedisForLock) Set(_ context.Context, key string, value interface{}, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.store[key] = fmt.Sprintf("%v", value)
	return nil
}
func (m *mockRedisForLock) Del(_ context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, k := range keys {
		delete(m.store, k)
	}
	return nil
}

func (m *mockRedisForLock) SetNX(_ context.Context, key string, value interface{}, _ time.Duration) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if _, ok := m.store[key]; ok {
		return false, nil
	}
	m.store[key] = fmt.Sprintf("%v", value)
	return true, nil
}

func (m *mockRedisForLock) DecrBy(_ context.Context, key string, amount int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	valStr := m.store[key]
	var current int64
	if valStr != "" {
		fmt.Sscanf(valStr, "%d", &current)
	}
	newVal := current - amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, nil
}

func (m *mockRedisForLock) DecrByIfPositive(_ context.Context, key string, amount int64) (int64, bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	valStr, exists := m.store[key]
	if !exists || valStr == "" {
		return -1, false, nil
	}
	var current int64
	fmt.Sscanf(valStr, "%d", &current)
	if current < amount {
		return current, false, nil
	}
	newVal := current - amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, true, nil
}

func (m *mockRedisForLock) IncrBy(_ context.Context, key string, amount int64) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	valStr := m.store[key]
	var current int64
	if valStr != "" {
		fmt.Sscanf(valStr, "%d", &current)
	}
	newVal := current + amount
	m.store[key] = fmt.Sprintf("%d", newVal)
	return newVal, nil
}

func (m *mockRedisForLock) Expire(_ context.Context, _ string, _ time.Duration) (bool, error) {
	return true, nil
}

func (m *mockRedisForLock) CompareAndDelete(_ context.Context, key, expected string) (bool, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.store[key] == expected {
		delete(m.store, key)
		return true, nil
	}
	return false, nil
}

func (m *mockRedisForLock) Close() error { return nil }
