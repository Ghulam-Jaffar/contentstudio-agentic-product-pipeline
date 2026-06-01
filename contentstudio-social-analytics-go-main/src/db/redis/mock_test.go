package redis

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestMockRedisClient_Get(t *testing.T) {
	mock := &MockRedisClient{}

	// Test with nil function
	result, err := mock.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Fatalf("expected empty string, got '%s'", result)
	}

	// Test with custom function
	mock.GetFunc = func(ctx context.Context, key string) (string, error) {
		return "value", nil
	}
	result, err = mock.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "value" {
		t.Fatalf("expected 'value', got '%s'", result)
	}

	// Test with error
	mock.GetFunc = func(ctx context.Context, key string) (string, error) {
		return "", errors.New("redis error")
	}
	_, err = mock.Get(context.Background(), "key")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockRedisClient_Set(t *testing.T) {
	mock := &MockRedisClient{}

	// Test with nil function
	err := mock.Set(context.Background(), "key", "value", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	called := false
	mock.SetFunc = func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
		called = true
		if key != "test_key" {
			t.Errorf("expected key 'test_key', got '%s'", key)
		}
		return nil
	}
	err = mock.Set(context.Background(), "test_key", "value", time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected SetFunc to be called")
	}

	// Test with error
	mock.SetFunc = func(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
		return errors.New("redis error")
	}
	err = mock.Set(context.Background(), "key", "value", time.Hour)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockRedisClient_Del(t *testing.T) {
	mock := &MockRedisClient{}

	// Test with nil function
	err := mock.Del(context.Background(), "key1", "key2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	var deletedKeys []string
	mock.DelFunc = func(ctx context.Context, keys ...string) error {
		deletedKeys = keys
		return nil
	}
	err = mock.Del(context.Background(), "key1", "key2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(deletedKeys) != 2 {
		t.Fatalf("expected 2 keys, got %d", len(deletedKeys))
	}

	// Test with error
	mock.DelFunc = func(ctx context.Context, keys ...string) error {
		return errors.New("redis error")
	}
	err = mock.Del(context.Background(), "key1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockRedisClient_Exists(t *testing.T) {
	mock := &MockRedisClient{}

	// Test with nil function
	count, err := mock.Exists(context.Background(), "key1", "key2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 0 {
		t.Fatalf("expected 0, got %d", count)
	}

	// Test with custom function
	mock.ExistsFunc = func(ctx context.Context, keys ...string) (int64, error) {
		return int64(len(keys)), nil
	}
	count, err = mock.Exists(context.Background(), "key1", "key2", "key3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if count != 3 {
		t.Fatalf("expected 3, got %d", count)
	}

	// Test with error
	mock.ExistsFunc = func(ctx context.Context, keys ...string) (int64, error) {
		return 0, errors.New("redis error")
	}
	_, err = mock.Exists(context.Background(), "key1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockRedisClient_SRandMember(t *testing.T) {
	mock := &MockRedisClient{}

	// Test with nil function
	result, err := mock.SRandMember(context.Background(), "set_key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "" {
		t.Fatalf("expected empty string, got '%s'", result)
	}

	// Test with custom function
	mock.SRandMemberFunc = func(ctx context.Context, key string) (string, error) {
		return "random_member", nil
	}
	result, err = mock.SRandMember(context.Background(), "set_key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != "random_member" {
		t.Fatalf("expected 'random_member', got '%s'", result)
	}

	// Test with error
	mock.SRandMemberFunc = func(ctx context.Context, key string) (string, error) {
		return "", errors.New("redis error")
	}
	_, err = mock.SRandMember(context.Background(), "set_key")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockRedisClient_SAdd(t *testing.T) {
	mock := &MockRedisClient{}

	// Test with nil function
	err := mock.SAdd(context.Background(), "set_key", "member1", "member2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	var addedMembers []interface{}
	mock.SAddFunc = func(ctx context.Context, key string, members ...interface{}) error {
		addedMembers = members
		return nil
	}
	err = mock.SAdd(context.Background(), "set_key", "member1", "member2")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(addedMembers) != 2 {
		t.Fatalf("expected 2 members, got %d", len(addedMembers))
	}

	// Test with error
	mock.SAddFunc = func(ctx context.Context, key string, members ...interface{}) error {
		return errors.New("redis error")
	}
	err = mock.SAdd(context.Background(), "set_key", "member1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockRedisClient_SMembers(t *testing.T) {
	mock := &MockRedisClient{}

	// Test with nil function
	result, err := mock.SMembers(context.Background(), "set_key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != nil {
		t.Fatalf("expected nil, got %v", result)
	}

	// Test with custom function
	mock.SMembersFunc = func(ctx context.Context, key string) ([]string, error) {
		return []string{"member1", "member2"}, nil
	}
	result, err = mock.SMembers(context.Background(), "set_key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 members, got %d", len(result))
	}

	// Test with error
	mock.SMembersFunc = func(ctx context.Context, key string) ([]string, error) {
		return nil, errors.New("redis error")
	}
	_, err = mock.SMembers(context.Background(), "set_key")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockRedisClient_Ping(t *testing.T) {
	mock := &MockRedisClient{}

	// Test with nil function
	err := mock.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	called := false
	mock.PingFunc = func(ctx context.Context) error {
		called = true
		return nil
	}
	err = mock.Ping(context.Background())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected PingFunc to be called")
	}

	// Test with error
	mock.PingFunc = func(ctx context.Context) error {
		return errors.New("connection error")
	}
	err = mock.Ping(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockRedisClient_Close(t *testing.T) {
	mock := &MockRedisClient{}

	// Test with nil function
	err := mock.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test with custom function
	called := false
	mock.CloseFunc = func() error {
		called = true
		return nil
	}
	err = mock.Close()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("expected CloseFunc to be called")
	}

	// Test with error
	mock.CloseFunc = func() error {
		return errors.New("close error")
	}
	err = mock.Close()
	if err == nil {
		t.Fatal("expected error")
	}
}
