package redis

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

// Client is the minimal cache contract: get, set, close.
// Consumers that need richer operations (locking, quota tracking, dedup)
// should define their own narrow interfaces satisfied by *RedisClient.
type Client interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Close() error
}

type RedisClient struct {
	rdb    *redis.Client
	logger zerolog.Logger
}

var compareAndDeleteScript = redis.NewScript(`
if redis.call("GET", KEYS[1]) == ARGV[1] then
	return redis.call("DEL", KEYS[1])
end
return 0
`)

var decrByIfPositiveScript = redis.NewScript(`
local current = tonumber(redis.call("GET", KEYS[1]))
if current == nil then
	return {-1, 0}
end
if current < tonumber(ARGV[1]) then
	return {current, 0}
end
local new_val = redis.call("DECRBY", KEYS[1], tonumber(ARGV[1]))
return {new_val, 1}
`)

// incrByIfExistsScript and decrByIfExistsScript refuse to mutate a missing key.
// Plain INCRBY/DECRBY auto-create the key with no TTL, which leaves a
// permanent "zombie" counter behind once the original SetNX-seeded TTL expires.
// Returning {-1, 0} lets callers treat a vanished key as a no-op and defer to
// the next SetNX-seeded reseed from the source-of-truth store.
var incrByIfExistsScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 0 then
	return {-1, 0}
end
local new_val = redis.call("INCRBY", KEYS[1], tonumber(ARGV[1]))
return {new_val, 1}
`)

var decrByIfExistsScript = redis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 0 then
	return {-1, 0}
end
local new_val = redis.call("DECRBY", KEYS[1], tonumber(ARGV[1]))
return {new_val, 1}
`)

func NewClient(addr, password string, db int, logger zerolog.Logger) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &RedisClient{
		rdb:    rdb,
		logger: logger.With().Str("component", "redis").Logger(),
	}, nil
}

func (c *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := c.rdb.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil
	}
	return val, err
}

func (c *RedisClient) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

// SetNX sets a key only if it does not already exist. Returns true if the key was set.
func (c *RedisClient) SetNX(ctx context.Context, key string, value interface{}, expiration time.Duration) (bool, error) {
	return c.rdb.SetNX(ctx, key, value, expiration).Result()
}

// Del removes one or more keys.
func (c *RedisClient) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

// DecrBy atomically decrements the key by the given amount and returns the new value.
func (c *RedisClient) DecrBy(ctx context.Context, key string, amount int64) (int64, error) {
	return c.rdb.DecrBy(ctx, key, amount).Result()
}

// DecrByIfPositive atomically decrements the key only if the current value >= amount.
// Returns (newValue, true, nil) on successful deduction.
// Returns (currentValue, false, nil) if insufficient budget or key missing.
func (c *RedisClient) DecrByIfPositive(ctx context.Context, key string, amount int64) (int64, bool, error) {
	res, err := decrByIfPositiveScript.Run(ctx, c.rdb, []string{key}, amount).Int64Slice()
	if err != nil {
		return 0, false, err
	}
	if len(res) != 2 {
		return 0, false, fmt.Errorf("DecrByIfPositive: unexpected result length %d", len(res))
	}
	return res[0], res[1] == 1, nil
}

// IncrBy atomically increments the key by the given amount and returns the new value.
func (c *RedisClient) IncrBy(ctx context.Context, key string, amount int64) (int64, error) {
	return c.rdb.IncrBy(ctx, key, amount).Result()
}

// IncrByIfExists atomically increments the key only if it already exists.
// Returns (newValue, true, nil) on successful increment.
// Returns (-1, false, nil) if the key is missing — callers should treat this
// as a no-op so a stale post-eviction call does not revive the key without TTL.
func (c *RedisClient) IncrByIfExists(ctx context.Context, key string, amount int64) (int64, bool, error) {
	res, err := incrByIfExistsScript.Run(ctx, c.rdb, []string{key}, amount).Int64Slice()
	if err != nil {
		return 0, false, err
	}
	if len(res) != 2 {
		return 0, false, fmt.Errorf("IncrByIfExists: unexpected result length %d", len(res))
	}
	return res[0], res[1] == 1, nil
}

// DecrByIfExists atomically decrements the key only if it already exists.
// Returns (newValue, true, nil) on successful decrement.
// Returns (-1, false, nil) if the key is missing.
func (c *RedisClient) DecrByIfExists(ctx context.Context, key string, amount int64) (int64, bool, error) {
	res, err := decrByIfExistsScript.Run(ctx, c.rdb, []string{key}, amount).Int64Slice()
	if err != nil {
		return 0, false, err
	}
	if len(res) != 2 {
		return 0, false, fmt.Errorf("DecrByIfExists: unexpected result length %d", len(res))
	}
	return res[0], res[1] == 1, nil
}

// Expire sets an expiration time for the given key.
func (c *RedisClient) Expire(ctx context.Context, key string, expiration time.Duration) (bool, error) {
	return c.rdb.Expire(ctx, key, expiration).Result()
}

// CompareAndDelete deletes the key iff its current value matches expected.
// It is implemented as a single Lua script to avoid GET/DEL races.
func (c *RedisClient) CompareAndDelete(ctx context.Context, key, expected string) (bool, error) {
	res, err := compareAndDeleteScript.Run(ctx, c.rdb, []string{key}, expected).Int64()
	if err != nil {
		return false, err
	}

	return res == 1, nil
}

// Close gracefully closes the Redis connection.
func (c *RedisClient) Close() error {
	return c.rdb.Close()
}
