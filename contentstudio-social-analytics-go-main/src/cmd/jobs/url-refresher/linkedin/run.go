package linkedin

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"golang.org/x/sync/semaphore"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/mongodb"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/crypto"
	"github.com/rs/zerolog"
)

const mongoBatchSize int64 = 50
const postBatchSize = 500

type AccountResult struct {
	LinkedinID   string
	EntityType   string
	PostsFound   int
	URLsResolved int
	Refreshed    []clickhousemodels.LinkedInMinimalPost
	Err          error
}

var mongoGetValidLinkedInIDsFn = func(
	repo mongodb.UnifiedSocialRepository,
	ctx context.Context,
) ([]string, error) {
	accounts, err := repo.GetValidAccounts(ctx, mongomodels.PlatformLinkedIn, nil)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		if id := resolveLinkedInID(acc); id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

var chGetDistinctLinkedInIDsFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	tableName string,
	validIDs []string,
) ([]string, error) {
	return s.ClickhouseClient.GetDistinctLinkedInIDsWithStaleURLs(ctx, tableName, validIDs)
}

var mongoGetAccountsByIDsFn = func(
	repo mongodb.UnifiedSocialRepository,
	ctx context.Context,
	platformType string,
	ids []string,
) ([]mongomodels.SocialIntegration, error) {
	return repo.GetAccountsByPlatformIDs(ctx, platformType, ids)
}

var chGetPostsFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	linkedinID string,
	limit, offset int,
) ([]clickhousemodels.LinkedInMinimalPost, error) {
	return s.ClickhouseClient.GetMinimalLinkedInOlderThan7DaysByAccount(ctx, "linkedin_posts", linkedinID, limit, offset)
}

var liGetURLsFn = func(
	c *social.LinkedInClient,
	ctx context.Context,
	linkedinID, entityType, accessToken string,
	posts []clickhousemodels.LinkedInMinimalPost,
) ([]clickhousemodels.LinkedInMinimalPost, error) {
	return c.GetPostURLs(ctx, linkedinID, entityType, accessToken, posts)
}

var chBulkUpdateFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	posts []clickhousemodels.LinkedInMinimalPost,
) (int, error) {
	return s.ClickhouseClient.BulkUpdateLinkedInPostURLs(ctx, "linkedin_posts", posts)
}

var chMarkLinkedInRefreshedFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	linkedinID string,
) error {
	return s.ClickhouseClient.MarkLinkedInPostsRefreshed(ctx, "linkedin_posts", linkedinID)
}

var chBulkMarkLinkedInRefreshedFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	linkedinIDs []string,
) error {
	return s.ClickhouseClient.BulkMarkLinkedInPostsRefreshed(ctx, "linkedin_posts", linkedinIDs)
}

func Run(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	repo mongodb.UnifiedSocialRepository,
	liClient *social.LinkedInClient,
	chSink *conversions.ClickHouseSink,
	concurrency int,
	accountType string,
) {
	validIDs, err := mongoGetValidLinkedInIDsFn(repo, ctx)
	if err != nil {
		log.Error().Err(err).Str("platform", "linkedin").Msg("Failed to get valid LinkedIn account IDs from MongoDB")
		return
	}
	if len(validIDs) == 0 {
		log.Info().Str("platform", "linkedin").Msg("No valid LinkedIn accounts found in MongoDB")
		return
	}
	log.Info().Str("platform", "linkedin").Int("valid_accounts", len(validIDs)).Msg("Fetched valid LinkedIn account IDs from MongoDB")

	linkedinIDs, err := chGetDistinctLinkedInIDsFn(chSink, ctx, "linkedin_posts", validIDs)
	if err != nil {
		log.Error().Err(err).Str("platform", "linkedin").Msg("Failed to get distinct LinkedIn IDs with stale URLs")
		return
	}
	if len(linkedinIDs) == 0 {
		log.Info().Str("platform", "linkedin").Msg("No LinkedIn posts with stale URLs found")
		return
	}

	if concurrency <= 0 {
		concurrency = 15
	}

	total := int64(len(linkedinIDs))
	log.Info().
		Str("platform", "linkedin").
		Int64("total_accounts", total).
		Int("concurrency", concurrency).
		Msg("Starting LinkedIn URL refresh")

	sem := semaphore.NewWeighted(int64(concurrency))
	var wg sync.WaitGroup

	var (
		processed      atomic.Int64
		totalPosts     atomic.Int64
		totalResolved  atomic.Int64
		totalUpdated   atomic.Int64
		errCount       atomic.Int64
		allRefreshed   []clickhousemodels.LinkedInMinimalPost
		zeroResultIDs  []string
		mu             sync.Mutex
	)

	batchSize := int(mongoBatchSize)
	for i := 0; i < len(linkedinIDs); i += batchSize {
		if ctx.Err() != nil {
			break
		}

		end := i + batchSize
		if end > len(linkedinIDs) {
			end = len(linkedinIDs)
		}
		batch := linkedinIDs[i:end]

		accounts, err := mongoGetAccountsByIDsFn(repo, ctx, mongomodels.PlatformLinkedIn, batch)
		if err != nil {
			log.Error().Err(err).Str("platform", "linkedin").Msg("Failed to fetch accounts batch from MongoDB")
			break
		}

		accountByID := make(map[string]mongomodels.SocialIntegration, len(accounts))
		for _, acc := range accounts {
			id := resolveLinkedInID(acc)
			accountByID[id] = acc
		}

		log.Info().
			Str("platform", "linkedin").
			Int64("total_accounts", total).
			Int64("accounts_processed", processed.Load()).
			Int64("errors", errCount.Load()).
			Int64("total_posts_found", totalPosts.Load()).
			Int64("total_urls_resolved", totalResolved.Load()).
			Int64("total_rows_updated", totalUpdated.Load()).
			Msg("LinkedIn URL refresh progress")

		for _, linkedinID := range batch {
			acc, ok := accountByID[linkedinID]
			if !ok {
				log.Warn().Str("platform", "linkedin").Str("linkedin_id", linkedinID).Msg("Account not found in MongoDB, skipping")
				if err := chMarkLinkedInRefreshedFn(chSink, ctx, linkedinID); err != nil {
					log.Error().Err(err).Str("linkedin_id", linkedinID).Msg("Failed to mark LinkedIn posts refreshed after skip")
				}
				processed.Add(1)
				continue
			}

			token := resolveToken(acc, cfg.DecryptionKey)
			if strings.TrimSpace(token) == "" {
				log.Warn().Str("platform", "linkedin").Str("linkedin_id", linkedinID).Msg("No token found, skipping")
				if err := chMarkLinkedInRefreshedFn(chSink, ctx, linkedinID); err != nil {
					log.Error().Err(err).Str("linkedin_id", linkedinID).Msg("Failed to mark LinkedIn posts refreshed after skip")
				}
				processed.Add(1)
				continue
			}

			if err := sem.Acquire(ctx, 1); err != nil {
				break
			}
			wg.Add(1)
			go func() {
				defer wg.Done()
				defer sem.Release(1)

				entityType := resolveEntityType(acc)
				accLogZ := log.With().
					Str("platform", "linkedin").
					Str("linkedin_id", linkedinID).
					Str("entity_type", entityType).
					Logger()
				accLog := &accLogZ

				start := time.Now()
				result := processAccountJob(ctx, cfg, liClient, chSink, acc, accLog)
				dur := time.Since(start)

				processed.Add(1)
				if result.Err != nil {
					errCount.Add(1)
					accLog.Error().Err(result.Err).
						Int("posts_found", result.PostsFound).
						Dur("duration", dur).
						Msg("Account URL refresh failed")
					if social.IsLinkedInAuthError(result.Err) {
						if markErr := chMarkLinkedInRefreshedFn(chSink, ctx, linkedinID); markErr != nil {
							accLog.Error().Err(markErr).Msg("Failed to mark LinkedIn posts refreshed after auth error")
						}
					}
					return
				}

				totalPosts.Add(int64(result.PostsFound))
				totalResolved.Add(int64(result.URLsResolved))
				accLog.Debug().
					Int("posts_found", result.PostsFound).
					Int("urls_resolved", result.URLsResolved).
					Dur("duration", dur).
					Msg("Account URL refresh done")

				mu.Lock()
				if len(result.Refreshed) > 0 {
					allRefreshed = append(allRefreshed, result.Refreshed...)
				} else {
					zeroResultIDs = append(zeroResultIDs, linkedinID)
				}
				mu.Unlock()
			}()
		}

		wg.Wait()

		if len(allRefreshed) > 0 {
			n, err := chBulkUpdateFn(chSink, ctx, allRefreshed)
			if err != nil {
				log.Error().Err(err).Int("attempted", len(allRefreshed)).Msg("Failed to bulk update LinkedIn post URLs in ClickHouse")
			} else {
				totalUpdated.Add(int64(n))
			}
			allRefreshed = nil
		}
	}

	if len(zeroResultIDs) > 0 {
		if markErr := chBulkMarkLinkedInRefreshedFn(chSink, ctx, zeroResultIDs); markErr != nil {
			log.Error().Err(markErr).Int("count", len(zeroResultIDs)).Msg("Failed to bulk mark LinkedIn posts refreshed")
		} else {
			log.Info().Int("count", len(zeroResultIDs)).Msg("Bulk marked LinkedIn posts refreshed")
		}
	}

	log.Info().
		Str("platform", "linkedin").
		Int64("total_accounts", total).
		Int64("accounts_processed", processed.Load()).
		Int64("errors", errCount.Load()).
		Int64("total_posts_found", totalPosts.Load()).
		Int64("total_urls_resolved", totalResolved.Load()).
		Int64("total_rows_updated", totalUpdated.Load()).
		Msg("LinkedIn URL refresh completed")
}

func processAccountJob(
	ctx context.Context,
	cfg *config.Config,
	liClient *social.LinkedInClient,
	chSink *conversions.ClickHouseSink,
	account mongomodels.SocialIntegration,
	log *zerolog.Logger,
) AccountResult {
	linkedinID := resolveLinkedInID(account)
	entityType := resolveEntityType(account)
	res := AccountResult{LinkedinID: linkedinID, EntityType: entityType}

	token := resolveToken(account, cfg.DecryptionKey)
	if strings.TrimSpace(token) == "" {
		res.Err = errors.New("linkedin access token is empty")
		log.Error().Msg("LinkedIn access token is empty")
		return res
	}

	for offset := 0; ; offset += postBatchSize {
		posts, err := chGetPostsFn(chSink, ctx, linkedinID, postBatchSize, offset)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get LinkedIn posts eligible for URL refresh")
			res.Err = err
			return res
		}
		if len(posts) == 0 {
			break
		}
		res.PostsFound += len(posts)

		refreshed, err := liGetURLsFn(liClient, ctx, linkedinID, entityType, token, posts)
		if err != nil {
			log.Error().Err(err).Msg("Failed to refresh LinkedIn post URLs")
			res.Err = err
			return res
		}
		res.URLsResolved += len(refreshed)
		res.Refreshed = append(res.Refreshed, refreshed...)

		if len(posts) < postBatchSize {
			break
		}
	}

	return res
}

func normalizeAccountTypes(accountType string) []string {
	trimmed := strings.TrimSpace(accountType)
	if trimmed == "" {
		return []string{mongomodels.TypePage}
	}
	switch strings.ToLower(trimmed) {
	case "page":
		return []string{mongomodels.TypePage}
	case "profile":
		return []string{mongomodels.TypeProfile}
	default:
		return []string{trimmed}
	}
}

func resolveLinkedInID(account mongomodels.SocialIntegration) string {
	if strings.TrimSpace(account.PlatformIdentifier) != "" {
		return account.PlatformIdentifier
	}
	return account.LinkedinID
}

func resolveEntityType(account mongomodels.SocialIntegration) string {
	if strings.EqualFold(strings.TrimSpace(account.Type), mongomodels.TypeProfile) {
		return "profile"
	}
	return "organization"
}

func resolveToken(account mongomodels.SocialIntegration, decryptionKey string) string {
	accessToken := strings.TrimSpace(account.GetAccessToken())
	if accessToken == "" {
		accessToken = getStringFromExtraData(account.ExtraData, "access_token")
	}
	if strings.TrimSpace(accessToken) == "" {
		return ""
	}
	if decrypted, err := crypto.DecryptToken(accessToken, decryptionKey); err == nil && strings.TrimSpace(decrypted) != "" {
		return decrypted
	}
	return accessToken
}

func getStringFromExtraData(extraData map[string]interface{}, key string) string {
	if extraData == nil {
		return ""
	}
	if value, ok := extraData[key].(string); ok {
		return value
	}
	return ""
}
