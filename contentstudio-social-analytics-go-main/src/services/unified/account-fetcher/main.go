package main

import (
	"context"
	"flag"
	"strings"
	"sync"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/cmd/jobs/fetcher"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// PlatformScale defines expected scale per platform for logging/monitoring
// Total: ~52K accounts
var PlatformScale = map[string]int{
	"facebook":  24000,
	"instagram": 16000,
	"linkedin":  8000,
	"youtube":   4000,
	"tiktok":    2000,
	"twitter":   1000,
	"pinterest": 1000,
	"gmb":       500,
	"meta_ads":  1000,
}

func main() {
	// Parse flags
	syncType := flag.String("syncType", "incremental", "Type of sync (incremental or full_sync)")
	platforms := flag.String("platforms", "", "Comma-separated platforms to process (empty = all supported)")
	facebookAccountTypes := flag.String("facebookAccountTypes", "page,group", "Facebook account types (comma-separated)")
	flag.Parse()

	// Load config
	cfg, err := config.LoadConfig()
	if err != nil {
		panic("Failed to load configuration: " + err.Error())
	}

	// Initialize logger
	log := logger.New(cfg.LogLevel)
	log.Info().
		Str("sync_type", *syncType).
		Str("platforms", *platforms).
		Msg("Starting Unified Account Fetcher")

	// Connect to MongoDB with connection pool configuration
	credential := options.Credential{
		Username:   cfg.Mongo.Username,
		Password:   cfg.Mongo.Password,
		AuthSource: cfg.Mongo.Database,
	}
	clientOpts := options.Client().
		ApplyURI(cfg.Mongo.URI).
		SetAuth(credential).
		SetMaxPoolSize(50).                          // Max connections in pool
		SetMinPoolSize(10).                          // Keep minimum connections ready
		SetSocketTimeout(60 * time.Second).          // Timeout for socket operations
		SetServerSelectionTimeout(10 * time.Second). // Timeout to select a server
		SetConnectTimeout(10 * time.Second)          // Timeout to establish connection

	mongoClient, err := mongo.Connect(context.Background(), clientOpts)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to MongoDB")
	}
	defer mongoClient.Disconnect(context.Background())

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := mongoClient.Ping(ctx, readpref.Primary()); err != nil {
		log.Fatal().Err(err).Msg("Failed to ping MongoDB")
	}
	log.Info().Msg("Connected to MongoDB")

	// Initialize Kafka producer
	producer, err := kafka.NewProducer(cfg.Kafka, log.Logger)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create Kafka producer")
	}
	defer producer.Close()
	log.Info().Msg("Kafka producer initialized")

	db := mongoClient.Database(cfg.Mongo.Database)
	bgCtx := context.Background()

	// Determine which platforms to process
	platformsToProcess := determinePlatforms(*platforms)

	log.Info().
		Strs("platforms", platformsToProcess).
		Msg("Processing platforms in parallel")

	// Process all platforms concurrently
	var wg sync.WaitGroup
	startTime := time.Now()

	for _, platform := range platformsToProcess {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			platformLog := log.Logger.With().Str("platform", p).Logger()
			platformStart := time.Now()

			switch p {
			case "facebook":
				accountTypes := parseAccountTypes(*facebookAccountTypes)
				platformLog.Info().Strs("account_types", accountTypes).Msg("Starting Facebook account fetch")
				fetcher.ProcessFacebookAccounts(db, producer, platformLog, accountTypes, *syncType)

			case "instagram":
				platformLog.Info().Msg("Starting Instagram account fetch")
				fetcher.ProcessInstagramAccounts(bgCtx, db, producer, platformLog, nil, *syncType)

			case "linkedin":
				platformLog.Info().Msg("Starting LinkedIn account fetch")
				fetcher.ProcessLinkedinAccounts(bgCtx, db, producer, platformLog, []string{"page", "profile"}, *syncType)

			case "youtube":
				platformLog.Info().Msg("Starting YouTube account fetch")
				fetcher.ProcessYouTubeAccounts(bgCtx, db, producer, platformLog, *syncType)

			case "tiktok":
				platformLog.Info().Msg("Starting TikTok account fetch")
				fetcher.ProcessTikTokAccounts(bgCtx, db, producer, platformLog, nil, *syncType)

			case "twitter":
				platformLog.Info().Msg("Starting Twitter account fetch")
				fetcher.ProcessTwitterAccounts(bgCtx, db, producer, platformLog, nil, *syncType)

			case "pinterest":
				platformLog.Info().Msg("Starting Pinterest account fetch")
				fetcher.ProcessPinterestAccounts(bgCtx, db, producer, platformLog, nil, *syncType)

			case "gmb":
				platformLog.Info().Msg("Starting GMB account fetch")
				fetcher.ProcessGMBAccounts(bgCtx, db, producer, platformLog, *syncType)

			case "meta_ads":
				platformLog.Info().Msg("Starting Meta Ads account fetch")
				fetcher.ProcessMetaAdsAccounts(db, producer, platformLog, *syncType)

			}

			platformLog.Info().
				Dur("duration", time.Since(platformStart)).
				Msg("Completed platform account fetch")
		}(platform)
	}

	wg.Wait()

	log.Info().
		Dur("total_duration", time.Since(startTime)).
		Strs("platforms", platformsToProcess).
		Msg("Unified Account Fetcher completed")
}

func determinePlatforms(platformsFlag string) []string {
	if platformsFlag == "" {
		// Default: all supported platforms
		return []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb", "meta_ads"}
	}

	var result []string
	for _, p := range strings.Split(platformsFlag, ",") {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

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

// ProcessPlatformConfig holds configuration for platform processing
type ProcessPlatformConfig struct {
	Platform             string
	SyncType             string
	FacebookAccountTypes []string
}

// ValidatePlatform checks if a platform is supported
func ValidatePlatform(platform string) bool {
	switch platform {
	case "facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb", "meta_ads":
		return true
	default:
		return false
	}
}

// GetDefaultPlatforms returns the default platforms to process
func GetDefaultPlatforms() []string {
	return []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest", "gmb", "meta_ads"}
}

// FilterValidPlatforms filters a list of platforms to only valid ones
func FilterValidPlatforms(platforms []string) []string {
	result := make([]string, 0, len(platforms))
	for _, p := range platforms {
		if ValidatePlatform(p) {
			result = append(result, p)
		}
	}
	return result
}

// ParseSyncType validates and normalizes the sync type
func ParseSyncType(syncType string) string {
	switch syncType {
	case "full_sync", "full":
		return "full_sync"
	case "immediate":
		return "immediate"
	case "incremental", "":
		return "incremental"
	default:
		return "incremental"
	}
}
