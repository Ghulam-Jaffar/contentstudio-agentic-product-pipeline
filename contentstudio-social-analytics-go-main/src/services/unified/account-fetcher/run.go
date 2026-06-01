package main

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/cmd/jobs/fetcher"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/kafka"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// PlatformProcessor defines the interface for platform-specific account processing.
// This abstraction enables dependency injection for testing with mock implementations.
//
// Performance: Each platform processor runs in its own goroutine for parallel execution.
type PlatformProcessor interface {
	// ProcessFacebook fetches and queues Facebook accounts for analytics processing.
	ProcessFacebook(
		db *mongo.Database,
		producer kafka.Producer,
		log zerolog.Logger,
		accountTypes []string,
		syncType string,
	)

	// ProcessInstagram fetches and queues Instagram accounts for analytics processing.
	ProcessInstagram(
		ctx context.Context,
		db *mongo.Database,
		producer kafka.Producer,
		log zerolog.Logger,
		accountTypes []string,
		syncType string,
	)

	// ProcessLinkedin fetches and queues LinkedIn accounts for analytics processing.
	ProcessLinkedin(
		ctx context.Context,
		db *mongo.Database,
		producer kafka.Producer,
		log zerolog.Logger,
		accountTypes []string,
		syncType string,
	)

	// ProcessYouTube fetches and queues YouTube accounts for analytics processing.
	ProcessYouTube(
		ctx context.Context,
		db *mongo.Database,
		producer kafka.Producer,
		log zerolog.Logger,
		syncType string,
	)

	// ProcessTikTok fetches and queues TikTok accounts for analytics processing.
	ProcessTikTok(
		ctx context.Context,
		db *mongo.Database,
		producer kafka.Producer,
		log zerolog.Logger,
		accountTypes []string,
		syncType string,
	)

	// ProcessTwitter fetches and queues Twitter accounts for analytics processing.
	ProcessTwitter(
		ctx context.Context,
		db *mongo.Database,
		producer kafka.Producer,
		log zerolog.Logger,
		accountTypes []string,
		syncType string,
	)

	// ProcessPinterest fetches and queues Pinterest accounts for analytics processing.
	ProcessPinterest(
		ctx context.Context,
		db *mongo.Database,
		producer kafka.Producer,
		log zerolog.Logger,
		accountTypes []string,
		syncType string,
	)

	// ProcessGMB fetches and queues Google My Business accounts for analytics processing.
	ProcessGMB(
		ctx context.Context,
		db *mongo.Database,
		producer kafka.Producer,
		log zerolog.Logger,
		syncType string,
	)

	// ProcessMetaAds fetches and queues Meta Ads accounts for analytics processing.
	ProcessMetaAds(
		db *mongo.Database,
		producer kafka.Producer,
		log zerolog.Logger,
		syncType string,
	)
}

// DefaultPlatformProcessor is the production implementation using the fetcher package.
// It delegates to platform-specific fetcher functions for actual processing.
type DefaultPlatformProcessor struct{}

// ProcessFacebook delegates to the fetcher package for Facebook account processing.
func (p *DefaultPlatformProcessor) ProcessFacebook(
	db *mongo.Database,
	producer kafka.Producer,
	log zerolog.Logger,
	accountTypes []string,
	syncType string,
) {
	fetcher.ProcessFacebookAccounts(db, producer, log, accountTypes, syncType)
}

// ProcessInstagram delegates to the fetcher package for Instagram account processing.
func (p *DefaultPlatformProcessor) ProcessInstagram(
	ctx context.Context,
	db *mongo.Database,
	producer kafka.Producer,
	log zerolog.Logger,
	accountTypes []string,
	syncType string,
) {
	fetcher.ProcessInstagramAccounts(ctx, db, producer, log, accountTypes, syncType)
}

// ProcessLinkedin delegates to the fetcher package for LinkedIn account processing.
func (p *DefaultPlatformProcessor) ProcessLinkedin(
	ctx context.Context,
	db *mongo.Database,
	producer kafka.Producer,
	log zerolog.Logger,
	accountTypes []string,
	syncType string,
) {
	fetcher.ProcessLinkedinAccounts(ctx, db, producer, log, accountTypes, syncType)
}

// ProcessYouTube delegates to the fetcher package for YouTube account processing.
func (p *DefaultPlatformProcessor) ProcessYouTube(
	ctx context.Context,
	db *mongo.Database,
	producer kafka.Producer,
	log zerolog.Logger,
	syncType string,
) {
	fetcher.ProcessYouTubeAccounts(ctx, db, producer, log, syncType)
}

// ProcessTikTok delegates to the fetcher package for TikTok account processing.
func (p *DefaultPlatformProcessor) ProcessTikTok(
	ctx context.Context,
	db *mongo.Database,
	producer kafka.Producer,
	log zerolog.Logger,
	accountTypes []string,
	syncType string,
) {
	fetcher.ProcessTikTokAccounts(ctx, db, producer, log, accountTypes, syncType)
}

// ProcessTwitter delegates to the fetcher package for Twitter account processing.
func (p *DefaultPlatformProcessor) ProcessTwitter(
	ctx context.Context,
	db *mongo.Database,
	producer kafka.Producer,
	log zerolog.Logger,
	accountTypes []string,
	syncType string,
) {
	fetcher.ProcessTwitterAccounts(ctx, db, producer, log, accountTypes, syncType)
}

// ProcessPinterest delegates to the fetcher package for Pinterest account processing.
func (p *DefaultPlatformProcessor) ProcessPinterest(
	ctx context.Context,
	db *mongo.Database,
	producer kafka.Producer,
	log zerolog.Logger,
	accountTypes []string,
	syncType string,
) {
	fetcher.ProcessPinterestAccounts(ctx, db, producer, log, accountTypes, syncType)
}

// ProcessGMB delegates to the fetcher package for Google My Business account processing.
func (p *DefaultPlatformProcessor) ProcessGMB(
	ctx context.Context,
	db *mongo.Database,
	producer kafka.Producer,
	log zerolog.Logger,
	syncType string,
) {
	fetcher.ProcessGMBAccounts(ctx, db, producer, log, syncType)
}

// ProcessMetaAds delegates to the fetcher package for Meta Ads account processing.
func (p *DefaultPlatformProcessor) ProcessMetaAds(
	db *mongo.Database,
	producer kafka.Producer,
	log zerolog.Logger,
	syncType string,
) {
	fetcher.ProcessMetaAdsAccounts(db, producer, log, syncType)
}

// ServiceDependencies holds all injectable dependencies for the unified account fetcher.
// This struct enables dependency injection for testing and flexible configuration.
type ServiceDependencies struct {
	Database          *mongo.Database   // MongoDB database connection
	Producer          kafka.Producer    // Kafka producer for work order messages
	Processor         PlatformProcessor // Platform-specific processor implementation
	Logger            *logger.Logger    // Structured logger instance
	Platforms         []string          // List of platforms to process
	SyncType          string            // Sync type: "incremental" or "full_sync"
	FacebookAcctTypes []string          // Facebook account types: "page", "group"
}

// RunService starts the unified account fetcher with the given dependencies.
// It processes all configured platforms in parallel using goroutines.
//
// Performance: O(n) where n is number of platforms. Each platform runs concurrently.
// Memory: Minimal overhead as processing is delegated to fetcher package.
func RunService(ctx context.Context, deps *ServiceDependencies) error {
	log := deps.Logger
	startTime := time.Now()

	log.Info().
		Strs("platforms", deps.Platforms).
		Str("sync_type", deps.SyncType).
		Msg("Processing platforms in parallel")

	var wg sync.WaitGroup

	// Launch parallel processors for each platform
	for _, platform := range deps.Platforms {
		wg.Add(1)
		go func(p string) {
			defer wg.Done()
			processPlatform(ctx, p, deps)
		}(platform)
	}

	wg.Wait()

	log.Info().
		Dur("total_duration", time.Since(startTime)).
		Strs("platforms", deps.Platforms).
		Msg("Unified Account Fetcher completed")

	return nil
}

// processPlatform handles processing for a single platform.
// Routes to the appropriate processor method based on platform name.
func processPlatform(ctx context.Context, platform string, deps *ServiceDependencies) {
	platformLog := deps.Logger.Logger.With().Str("platform", platform).Logger()
	platformStart := time.Now()

	switch platform {
	case "facebook":
		platformLog.Info().
			Strs("account_types", deps.FacebookAcctTypes).
			Msg("Starting Facebook account fetch")
		deps.Processor.ProcessFacebook(
			deps.Database,
			deps.Producer,
			platformLog,
			deps.FacebookAcctTypes,
			deps.SyncType,
		)

	case "instagram":
		platformLog.Info().Msg("Starting Instagram account fetch")
		deps.Processor.ProcessInstagram(
			ctx,
			deps.Database,
			deps.Producer,
			platformLog,
			nil,
			deps.SyncType,
		)

	case "linkedin":
		platformLog.Info().Msg("Starting LinkedIn account fetch")
		deps.Processor.ProcessLinkedin(
			ctx,
			deps.Database,
			deps.Producer,
			platformLog,
			nil,
			deps.SyncType,
		)

	case "youtube":
		platformLog.Info().Msg("Starting YouTube account fetch")
		deps.Processor.ProcessYouTube(
			ctx,
			deps.Database,
			deps.Producer,
			platformLog,
			deps.SyncType,
		)

	case "tiktok":
		platformLog.Info().Msg("Starting TikTok account fetch")
		deps.Processor.ProcessTikTok(
			ctx,
			deps.Database,
			deps.Producer,
			platformLog,
			nil,
			deps.SyncType,
		)

	case "twitter":
		platformLog.Info().Msg("Starting Twitter account fetch")
		deps.Processor.ProcessTwitter(
			ctx,
			deps.Database,
			deps.Producer,
			platformLog,
			nil,
			deps.SyncType,
		)

	case "pinterest":
		platformLog.Info().Msg("Starting Pinterest account fetch")
		deps.Processor.ProcessPinterest(
			ctx,
			deps.Database,
			deps.Producer,
			platformLog,
			nil,
			deps.SyncType,
		)

	case "gmb":
		platformLog.Info().Msg("Starting GMB account fetch")
		deps.Processor.ProcessGMB(
			ctx,
			deps.Database,
			deps.Producer,
			platformLog,
			deps.SyncType,
		)

	case "meta_ads":
		platformLog.Info().Msg("Starting Meta Ads account fetch")
		deps.Processor.ProcessMetaAds(
			deps.Database,
			deps.Producer,
			platformLog,
			deps.SyncType,
		)

	default:
		platformLog.Warn().Msg("No handler implemented for platform")
		return
	}

	platformLog.Info().
		Dur("duration", time.Since(platformStart)).
		Msg("Completed platform account fetch")
}

// BuildPlatformList builds the list of platforms to process from a flag string.
// Returns default platforms if input is empty.
func BuildPlatformList(platformsFlag string) []string {
	return determinePlatforms(platformsFlag)
}

// BuildAccountTypeList builds the list of account types from a flag string.
// Returns nil if input is empty.
func BuildAccountTypeList(accountTypesFlag string) []string {
	return parseAccountTypes(accountTypesFlag)
}

// NormalizeSyncType normalizes the sync type to a valid value.
// Maps "full" to "full_sync", defaults to "incremental".
func NormalizeSyncType(syncType string) string {
	return ParseSyncType(syncType)
}

// GetPlatformScaleInfo returns expected account scale for a platform.
// Used for monitoring and capacity planning.
func GetPlatformScaleInfo(platform string) (int, bool) {
	scale, ok := PlatformScale[platform]
	return scale, ok
}

// GetTotalAccountScale returns the total expected account count across all platforms.
// Performance: O(n) where n is number of platforms in PlatformScale map.
func GetTotalAccountScale() int {
	total := 0
	for _, scale := range PlatformScale {
		total += scale
	}
	return total
}

// IsPlatformSupported checks if a platform has an implemented handler.
func IsPlatformSupported(platform string) bool {
	switch platform {
	case "facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest":
		return true
	default:
		return false
	}
}

// GetSupportedPlatformsForFetcher returns platforms with implemented handlers.
func GetSupportedPlatformsForFetcher() []string {
	return []string{"facebook", "instagram", "linkedin", "youtube", "tiktok", "twitter", "pinterest"}
}

// CalculatePlatformProgress calculates progress as a percentage.
// Returns 0.0 if total is zero to avoid division by zero.
func CalculatePlatformProgress(processed, total int) float64 {
	if total == 0 {
		return 0.0
	}
	return float64(processed) / float64(total) * 100.0
}

// SplitPlatformString splits a comma-separated platform string into a slice.
func SplitPlatformString(platforms string) []string {
	return determinePlatforms(platforms)
}

// ValidatePlatformList validates all platforms in a list.
// Returns two slices: valid platforms and invalid platforms.
//
// Performance: O(n) where n is number of platforms.
func ValidatePlatformList(platforms []string) (valid, invalid []string) {
	for _, p := range platforms {
		if IsPlatformSupported(p) {
			valid = append(valid, p)
		} else {
			invalid = append(invalid, p)
		}
	}
	return valid, invalid
}

// CreateServiceConfig creates a ServiceDependencies with parsed configuration.
// Does not set Database, Producer, Processor, or Logger - those must be set separately.
func CreateServiceConfig(
	platforms, syncType, fbAccountTypes string,
) *ServiceDependencies {
	return &ServiceDependencies{
		Platforms:         BuildPlatformList(platforms),
		SyncType:          NormalizeSyncType(syncType),
		FacebookAcctTypes: BuildAccountTypeList(fbAccountTypes),
	}
}

// EstimateProcessingLoad calculates estimated work based on platform scales.
// Useful for capacity planning and load balancing.
//
// Performance: O(n) where n is number of platforms.
func EstimateProcessingLoad(platforms []string) int {
	total := 0
	for _, p := range platforms {
		if scale, ok := PlatformScale[p]; ok {
			total += scale
		}
	}
	return total
}

// FilterPlatformsWithScale returns only platforms that have scale information.
// Pre-allocates slice capacity for performance.
func FilterPlatformsWithScale(platforms []string) []string {
	result := make([]string, 0, len(platforms))
	for _, p := range platforms {
		if _, ok := PlatformScale[p]; ok {
			result = append(result, p)
		}
	}
	return result
}

// CompareSyncTypes compares two sync type strings for equivalence after normalization.
func CompareSyncTypes(a, b string) bool {
	return NormalizeSyncType(a) == NormalizeSyncType(b)
}

// MergePlatformLists merges two platform lists removing duplicates.
// Preserves order: elements from 'a' come before elements from 'b'.
//
// Performance: O(n+m) where n and m are lengths of input slices.
// Memory: Uses map for O(1) duplicate detection.
func MergePlatformLists(a, b []string) []string {
	seen := make(map[string]bool, len(a)+len(b))
	result := make([]string, 0, len(a)+len(b))

	for _, p := range a {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	for _, p := range b {
		if !seen[p] {
			seen[p] = true
			result = append(result, p)
		}
	}
	return result
}

// GetPlatformCount returns the number of platforms to process.
func GetPlatformCount(platforms string) int {
	list := BuildPlatformList(platforms)
	return len(list)
}

// IsFullSync returns true if the sync type is a full sync.
func IsFullSync(syncType string) bool {
	return NormalizeSyncType(syncType) == "full_sync"
}

// GetFacebookAccountTypeCount returns the number of Facebook account types.
func GetFacebookAccountTypeCount(accountTypes string) int {
	list := BuildAccountTypeList(accountTypes)
	return len(list)
}
