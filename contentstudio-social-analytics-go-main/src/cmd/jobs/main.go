package main

import (
	"context"
	"flag"
	"strings"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/cmd/jobs/fetcher"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"

	stdlog "log" // Standard logger for initial errors

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// parseAccountTypes parses the accountType flag value into a slice of strings.
// Supports both single value ("page") and comma-separated values ("profile,page").
func parseAccountTypes(accountType string) []string {
	if accountType == "" {
		return nil
	}
	types := strings.Split(accountType, ",")
	result := make([]string, 0, len(types))
	for _, t := range types {
		t = strings.TrimSpace(t)
		if t != "" {
			result = append(result, t)
		}
	}
	return result
}

func main() {
	// 1. Define and parse command-line flags
	socialNetwork := flag.String("socialNetwork", "", "Social network to process (e.g., facebook)")
	accountType := flag.String("accountType", "", "Type of account to fetch (e.g., page, group). Supports comma-separated values (e.g., profile,page)")
	syncType := flag.String("syncType", "incremental", "Type of sync operation (incremental or full_sync)")
	flag.Parse()

	if *socialNetwork == "" {
		stdlog.Fatal("Error: -socialNetwork flag is required")
	}

	// Parse account types into slice
	accountTypes := parseAccountTypes(*accountType)

	if len(accountTypes) == 0 && *socialNetwork == "facebook" {
		stdlog.Fatal("Error: -accountType flag is required for Facebook processing")
	}

	// 2. Load Configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		stdlog.Fatalf("Failed to load configuration: %v", err)
	}

	// 3. Initialize Logger
	appLogger := logger.New(cfg.LogLevel)
	appLogger.Info().
		Str("socialNetwork", *socialNetwork).
		Strs("accountTypes", accountTypes).
		Str("syncType", *syncType).
		Msg("Scheduler starting")

	// 4. Initialize MongoDB
	appLogger.Info().Msg("Connecting to MongoDB...")

	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	clientOpts := options.Client().ApplyURI(cfg.Mongo.URI).SetAuth(credential)

	mongoClient, err := mongo.Connect(context.TODO(), clientOpts)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(context.TODO())

	ctxPing, cancelPing := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelPing()
	if err := mongoClient.Ping(ctxPing, readpref.Primary()); err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}
	appLogger.Info().Msg("Successfully connected and pinged MongoDB.")

	// 5. Initialize Kafka Producer
	appLogger.Info().Msg("Initializing Kafka producer...")
	kafkaProducer, err := kafka.NewProducer(cfg.Kafka, appLogger.Logger)
	if err != nil {
		appLogger.Fatal().Err(err).Msg("Failed to initialize Kafka producer")
	}
	defer kafkaProducer.Close()
	appLogger.Info().Msg("Kafka producer initialized successfully.")

	db := mongoClient.Database(cfg.Mongo.Database)

	ctx := context.Background()
	appLogger.Info().Msg("Running scheduled sync...")

	switch *socialNetwork {
	case "facebook":
		fetcher.ProcessFacebookAccounts(db, kafkaProducer, appLogger.Logger, accountTypes, *syncType)
	case "instagram":
		fetcher.ProcessInstagramAccounts(ctx, db, kafkaProducer, appLogger.Logger, accountTypes, *syncType)
	case "linkedin":
		fetcher.ProcessLinkedinAccounts(ctx, db, kafkaProducer, appLogger.Logger, accountTypes, *syncType)
	case "tiktok":
		fetcher.ProcessTikTokAccounts(ctx, db, kafkaProducer, appLogger.Logger, accountTypes, *syncType)
	case "listening":
		fetcher.ProcessListeningTopics(ctx, db, kafkaProducer, appLogger.Logger)
	default:
		appLogger.Warn().Str("socialNetwork", *socialNetwork).Msg("No handler implemented for this social network")
	}
}
