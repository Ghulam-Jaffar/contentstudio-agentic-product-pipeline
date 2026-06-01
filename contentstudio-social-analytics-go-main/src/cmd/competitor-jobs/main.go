package main

import (
	"context"
	"errors"
	"flag"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/common/telemetry"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/cmd/competitor-jobs/fetcher"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

//
// -------------------- Constants & Defaults --------------------
//

const (
	// Supported social networks
	SocialFacebook  = "facebook"
	SocialInstagram = "instagram"

	// Supported sync types
	SyncIncremental = "incremental"
	SyncFullRefresh = "full_refresh"

	// MongoDB
	MongoPingTimeout = 5 * time.Second

	// Application
	AppName = "competitor-jobs-fetcher"
)

//
// -------------------- Main --------------------
//

func main() {
	// --------------------
	// Parse CLI arguments
	// --------------------
	socialNetwork := flag.String(
		"socialNetwork",
		SocialFacebook,
		"Target social network: facebook | instagram",
	)

	syncType := flag.String(
		"syncType",
		SyncIncremental,
		"Sync type: incremental | full_refresh",
	)

	flag.Parse()

	// --------------------
	// Load configuration
	// --------------------
	cfg, err := config.LoadConfig()
	if err != nil {
		logFatal("failed to load configuration", err)
	}

	telemetry.ConfigureSentry(cfg)

	appLogger := logger.New(cfg.LogLevel)

	rootOp := appLogger.
		Operation("competitor_jobs_fetcher").
		WithSentryTags(map[string]string{
			"app":            AppName,
			"social_network": *socialNetwork,
			"sync_type":      *syncType,
		})

	defer func() {
		rootOp.Complete(nil, "")
		logger.FlushSentry(5 * time.Second)
	}()

	appLogger.Info().
		Str("app", AppName).
		Str("social_network", *socialNetwork).
		Str("sync_type", *syncType).
		Msg("Starting job fetcher")

	// --------------------
	// Validate CLI input
	// --------------------
	if !isValidSocialNetwork(*socialNetwork) {
		err := errors.New("invalid socialNetwork value")
		rootOp.Complete(err, "invalid_cli")
		appLogger.Fatal().
			Str("socialNetwork", *socialNetwork).
			Msg("Invalid socialNetwork value")
	}

	if !isValidSyncType(*syncType) {
		err := errors.New("invalid syncType value")
		rootOp.Complete(err, "invalid_cli")
		appLogger.Fatal().
			Str("syncType", *syncType).
			Msg("Invalid syncType value")
	}

	// --------------------
	// Initialize Kafka producer
	// --------------------
	producer, err := kafka.NewProducer(cfg.Kafka, appLogger.Logger)
	if err != nil {
		rootOp.Complete(err, "kafka_init_failed")
		appLogger.Fatal().Err(err).Msg("Failed to initialize Kafka producer")
	}
	defer producer.Close()

	// --------------------
	// Initialize MongoDB
	// --------------------
	ctx := context.Background()

	mongoClient, err := initMongo(ctx, cfg)
	if err != nil {
		rootOp.Complete(err, "mongo_init_failed")
		appLogger.Fatal().Err(err).Msg("Failed to initialize MongoDB")
	}
	defer mongoClient.Disconnect(ctx)

	db := mongoClient.Database(cfg.Mongo.Database)

	appLogger.Info().
		Str("database", cfg.Mongo.Database).
		Msg("MongoDB connection established")

	// --------------------
	// Dispatch by network
	// --------------------
	switch *socialNetwork {

	case SocialFacebook:
		appLogger.Info().Msg("Processing Facebook accounts")
		fetcher.ProcessFacebookAccounts(
			ctx,
			db,
			producer,
			appLogger,
			*syncType,
		)

	case SocialInstagram:
		appLogger.Info().Msg("Processing Instagram accounts")
		fetcher.ProcessInstagramAccounts(
			ctx,
			db,
			producer,
			appLogger,
			*syncType,
		)
	}

	appLogger.Info().Msg("Job fetcher completed successfully")
}

//
// -------------------- Helpers --------------------
//

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

// isValidSocialNetwork validates supported social networks.
func isValidSocialNetwork(value string) bool {
	switch value {
	case SocialFacebook, SocialInstagram:
		return true
	default:
		return false
	}
}

// isValidSyncType validates supported sync modes.
func isValidSyncType(value string) bool {
	switch value {
	case SyncIncremental, SyncFullRefresh:
		return true
	default:
		return false
	}
}

// logFatal is a minimal fallback logger for very early failures.
func logFatal(msg string, err error) {
	_, _ = os.Stderr.WriteString(msg + ": " + err.Error() + "\n")
	os.Exit(1)
}
