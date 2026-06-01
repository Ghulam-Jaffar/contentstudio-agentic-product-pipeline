package redis

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/rs/zerolog"
)

func newTestClient(t *testing.T) (*RedisClient, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	client, err := NewClient(mr.Addr(), "", 0, zerolog.New(io.Discard))
	if err != nil {
		t.Fatalf("failed to create redis client: %v", err)
	}
	return client, mr
}

func TestNewClient_Success(t *testing.T) {
	client, _ := newTestClient(t)
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	client.Close()
}

func TestNewClient_ConnectionFailure(t *testing.T) {
	_, err := NewClient("localhost:1", "", 0, zerolog.New(io.Discard))
	if err == nil {
		t.Fatal("expected error for unreachable redis")
	}
}

func TestSetAndGet(t *testing.T) {
	client, _ := newTestClient(t)
	defer client.Close()

	ctx := context.Background()

	err := client.Set(ctx, "test_key", "test_value", time.Minute)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	val, err := client.Get(ctx, "test_key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "test_value" {
		t.Fatalf("expected test_value, got %q", val)
	}
}

func TestGet_NonExistentKey(t *testing.T) {
	client, _ := newTestClient(t)
	defer client.Close()

	val, err := client.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("expected nil error for missing key, got: %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string for missing key, got %q", val)
	}
}

func TestSet_TTLExpiration(t *testing.T) {
	client, mr := newTestClient(t)
	defer client.Close()

	ctx := context.Background()

	err := client.Set(ctx, "ttl_key", "value", time.Second)
	if err != nil {
		t.Fatalf("Set failed: %v", err)
	}

	mr.FastForward(2 * time.Second)

	val, err := client.Get(ctx, "ttl_key")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if val != "" {
		t.Fatalf("expected empty string after TTL, got %q", val)
	}
}

func TestSet_OverwriteValue(t *testing.T) {
	client, _ := newTestClient(t)
	defer client.Close()

	ctx := context.Background()

	client.Set(ctx, "key", "first", time.Minute)
	client.Set(ctx, "key", "second", time.Minute)

	val, _ := client.Get(ctx, "key")
	if val != "second" {
		t.Fatalf("expected second, got %q", val)
	}
}

func TestClose(t *testing.T) {
	client, _ := newTestClient(t)
	err := client.Close()
	if err != nil {
		t.Fatalf("Close failed: %v", err)
	}
}
