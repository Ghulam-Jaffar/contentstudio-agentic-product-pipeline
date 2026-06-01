package main

import (
	"context"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

var newRateManager = social.NewRateManager
var newRedisClient = func(opts *redis.Options) redisClient {
	return redis.NewClient(opts)
}

type redisClient interface {
	Do(ctx context.Context, args ...interface{}) *redis.Cmd
	Close() error
}

func mustLoadConfig() *config.Config {
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}
	return cfg
}

func initLogger(cfg *config.Config) *logger.Logger {
	return logger.New(cfg.LogLevel)
}

func mustConnectMongo(ctx context.Context, cfg *config.Config, log logger.Logger) (*mongo.Database, func()) {
	mLog := log.With().Str("component", "bootstrap.mongo").Logger()
	mLog.Info().Msg("Connecting to MongoDB...")

	cred := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	opts := options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(cred)

	start := time.Now()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		mLog.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}

	ctxPing, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := client.Ping(ctxPing, readpref.Primary()); err != nil {
		mLog.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}

	mLog.Info().
		Dur("connect_and_ping_ms", time.Since(start)).
		Str("database", cfg.Mongo.Database).
		Msg("Successfully connected and pinged MongoDB")

	db := client.Database(cfg.Mongo.Database)
	cleanup := func() {
		if derr := client.Disconnect(context.Background()); derr != nil {
			mLog.Error().Err(derr).Msg("Error on MongoDB disconnect")
		}
	}

	return db, cleanup
}

func buildRateManager(cfg *config.Config, log logger.Logger) *social.RateManager {
	rl := log.With().Str("component", "ratelimit").Logger()

	perTokenRPS := cfg.URLRefresher.PerTokenRPS
	if perTokenRPS <= 0 {
		perTokenRPS = cfg.Facebook.PerTokenRPS
	}
	if perTokenRPS <= 0 {
		perTokenRPS = 3.0
	}
	perTokenBurst := cfg.URLRefresher.PerTokenBurst
	if perTokenBurst <= 0 {
		perTokenBurst = cfg.Facebook.PerTokenBurst
	}
	if perTokenBurst <= 0 {
		perTokenBurst = 3
	}
	globalRPS := cfg.URLRefresher.GlobalRPS
	if globalRPS <= 0 {
		globalRPS = cfg.Facebook.GlobalRPS
	}
	if globalRPS <= 0 {
		globalRPS = 10.0
	}
	globalBurst := cfg.URLRefresher.GlobalBurst
	if globalBurst <= 0 {
		globalBurst = cfg.Facebook.GlobalBurst
	}
	if globalBurst <= 0 {
		globalBurst = 10
	}

	rl.Info().
		Float64("per_token_rps", perTokenRPS).
		Int("per_token_burst", perTokenBurst).
		Float64("global_rps", globalRPS).
		Int("global_burst", globalBurst).
		Msg("Configured rate limits")

	return newRateManager(social.RateLimits{
		PerTokenRPS:   perTokenRPS,
		PerTokenBurst: perTokenBurst,
		GlobalRPS:     globalRPS,
		GlobalBurst:   globalBurst,
	})
}

func mustConnectRedis(cfg *config.Config) redisClient {
	maxRetries := cfg.Redis.MaxRetries
	if maxRetries < 3 {
		maxRetries = 3
	}

	poolSize := cfg.Redis.PoolSize
	if poolSize < 50 {
		poolSize = 50
	}

	return newRedisClient(&redis.Options{
		Addr:         cfg.Redis.Addr,
		Password:     cfg.Redis.Password,
		DB:           cfg.Redis.DB,
		MaxRetries:   maxRetries,
		PoolSize:     poolSize,
		DialTimeout:  20 * time.Second,
		ReadTimeout:  20 * time.Second,
		WriteTimeout: 20 * time.Second,
		PoolTimeout:  20 * time.Second,
	})
}
