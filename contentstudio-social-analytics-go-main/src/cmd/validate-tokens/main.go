package main

import (
	"context"
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/tokenstore"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	AppName = "validate-tokens"
)

func main() {
	// Parse CLI arguments
	platform := flag.String(
		"channel",
		"facebook",
		"Select the channel: facebook | instagram",
	)

	flag.Parse()

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		logFatal("failed to load configuration", err)
	}

	telemetry.ConfigureSentry(cfg)

	appLogger := logger.New(cfg.LogLevel)

	op := appLogger.
		Operation("validate_tokens").
		WithSentryTags(map[string]string{
			"app":     AppName,
			"channel": *platform,
		})

	defer func() {
		op.Complete(nil, "")
		logger.FlushSentry(5 * time.Second)
	}()

	appLogger.Info().
		Str("app", AppName).
		Str("channel", *platform).
		Msg("Starting token sync job")

	// Validate platform
	if !isValidPlatform(*platform) {
		appLogger.Fatal().
			Str("channel", *platform).
			Msg("Invalid channel value. Must be 'facebook' or 'instagram'")
	}

	// Initialize context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		appLogger.Warn().Msg("Received shutdown signal, canceling job...")
		cancel()
	}()

	// Initialize Redis client
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Addr,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	defer redisClient.Close()

	// Test Redis connection
	if err := redisClient.Ping(ctx).Err(); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to Redis")
	}

	appLogger.Info().Msg("Redis connection established")

	// Initialize MongoDB client
	mongoClient, err := initMongo(ctx, cfg)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to initialize MongoDB")
	}

	defer func() {
		if err := mongoClient.Disconnect(ctx); err != nil {
			appLogger.Error().Err(err).Msg("Failed to disconnect from MongoDB")
		}
	}()

	// Test MongoDB connection
	if err := mongoClient.Ping(ctx, nil); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}

	appLogger.Info().Msg("MongoDB connection established")

	mongoDB := mongoClient.Database(cfg.Mongo.Database)

	// Create TokenStore and run job
	ts := tokenstore.NewTokenStore(
		*platform,
		redisClient,
		mongoDB,
		appLogger,
	)

	if err := ts.ProcessJob(ctx); err != nil {
		if ctx.Err() != nil {
			appLogger.Warn().Msg("Job cancelled by user")
			os.Exit(0)
		}
		appLogger.Fatal().Err(err).Msg("Token sync job failed")
	}

	appLogger.Info().Msg("Token sync job completed successfully")
}

// isValidPlatform checks if the platform is supported
func isValidPlatform(platform string) bool {
	return platform == tokenstore.PlatformFacebook || platform == tokenstore.PlatformInstagram
}

// logFatal logs a fatal error and exits
func logFatal(msg string, err error) {
	log := logger.New("error")

	op := log.
		Operation("fatal_startup_error").
		WithSentryExtras(map[string]interface{}{
			"message": msg,
		})

	op.Complete(err, "")
	log.Fatal().Err(err).Msg(msg)

	logger.FlushSentry(5 * time.Second)
}

// initMongo initializes MongoDB connection
func initMongo(ctx context.Context, cfg *config.Config) (*mongo.Client, error) {
	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}

	clientOpts := options.Client().
		ApplyURI(cfg.Mongo.URI).
		SetAuth(credential)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		return nil, err
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := client.Ping(pingCtx, nil); err != nil {
		return nil, err
	}

	return client, nil
}
