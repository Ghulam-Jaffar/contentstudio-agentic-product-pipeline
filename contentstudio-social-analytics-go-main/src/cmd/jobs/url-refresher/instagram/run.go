package instagram

import (
	"context"
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

const (
	mongoBatchSize int64 = 50
	postBatchSize        = 500
)

type AccountResult struct {
	InstagramID  string
	PostsFound   int
	URLsResolved int
	Refreshed    []clickhousemodels.InstagramMinimalPost
	Err          error
}

var mongoGetValidInstagramIDsFn = func(
	repo mongodb.UnifiedSocialRepository,
	ctx context.Context,
) ([]string, error) {
	accounts, err := repo.GetValidAccounts(ctx, mongomodels.PlatformInstagram, nil)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		id := acc.PlatformIdentifier
		if id == "" {
			id = acc.InstagramID
		}
		if id != "" {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

var chGetDistinctInstagramIDsFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	tableName string,
	validIDs []string,
) ([]string, error) {
	return s.ClickhouseClient.GetDistinctInstagramIDsWithStaleURLs(ctx, tableName, validIDs)
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
	instagramID string,
	limit, offset int,
) ([]clickhousemodels.InstagramMinimalPost, error) {
	return s.ClickhouseClient.GetMinimalInstagramOlderThan20DaysByAccount(ctx, "instagram_posts", instagramID, limit, offset)
}

var igGetURLsFn = func(
	c *social.InstagramClient,
	ctx context.Context,
	instagramID string,
	account mongomodels.SocialIntegration,
	decryptionKey string,
	posts []clickhousemodels.InstagramMinimalPost,
) ([]clickhousemodels.InstagramMinimalPost, error) {
	token := resolveToken(account, decryptionKey)
	return c.GetMediaURLs(ctx, instagramID, token, posts)
}

var chBulkUpdateFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	posts []clickhousemodels.InstagramMinimalPost,
) (int, error) {
	return s.ClickhouseClient.BulkUpdateInstagramMediaURLs(ctx, "instagram_posts", posts)
}

var chMarkInstagramRefreshedFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	instagramID string,
) error {
	return s.ClickhouseClient.MarkInstagramPostsRefreshed(ctx, "instagram_posts", instagramID)
}

var chBulkMarkInstagramRefreshedFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	instagramIDs []string,
) error {
	return s.ClickhouseClient.BulkMarkInstagramPostsRefreshed(ctx, "instagram_posts", instagramIDs)
}

func Run(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	repo mongodb.UnifiedSocialRepository,
	igClientFB *social.InstagramClient,
	igClientIG *social.InstagramClient,
	chSink *conversions.ClickHouseSink,
	concurrency int,
	accountType string,
) {
	validIDs, err := mongoGetValidInstagramIDsFn(repo, ctx)
	if err != nil {
		log.Error().Err(err).Str("platform", "instagram").Msg("Failed to get valid Instagram account IDs from MongoDB")
		return
	}
	if len(validIDs) == 0 {
		log.Info().Str("platform", "instagram").Msg("No valid Instagram accounts found in MongoDB")
		return
	}
	log.Info().Str("platform", "instagram").Int("valid_accounts", len(validIDs)).Msg("Fetched valid Instagram account IDs from MongoDB")

	igIDs, err := chGetDistinctInstagramIDsFn(chSink, ctx, "instagram_posts", validIDs)
	if err != nil {
		log.Error().Err(err).Str("platform", "instagram").Msg("Failed to get distinct Instagram IDs with stale URLs")
		return
	}
	if len(igIDs) == 0 {
		log.Info().Str("platform", "instagram").Msg("No Instagram posts with stale URLs found")
		return
	}

	if concurrency <= 0 {
		concurrency = 30
	}

	total := int64(len(igIDs))
	log.Info().
		Str("platform", "instagram").
		Int64("total_accounts", total).
		Int("concurrency", concurrency).
		Msg("Starting Instagram URL refresh")

	sem := semaphore.NewWeighted(int64(concurrency))
	var wg sync.WaitGroup

	var (
		processed      atomic.Int64
		totalPosts     atomic.Int64
		totalResolved  atomic.Int64
		totalUpdated   atomic.Int64
		errCount       atomic.Int64
		allRefreshed   []clickhousemodels.InstagramMinimalPost
		successfulIDs  []string
		mu             sync.Mutex
	)

	batchSize := int(mongoBatchSize)
	for i := 0; i < len(igIDs); i += batchSize {
		if ctx.Err() != nil {
			break
		}

		end := i + batchSize
		if end > len(igIDs) {
			end = len(igIDs)
		}
		batch := igIDs[i:end]

		accounts, err := mongoGetAccountsByIDsFn(repo, ctx, mongomodels.PlatformInstagram, batch)
		if err != nil {
			log.Error().Err(err).Str("platform", "instagram").Msg("Failed to fetch accounts batch from MongoDB")
			break
		}

		accountByID := make(map[string]mongomodels.SocialIntegration, len(accounts))
		for _, acc := range accounts {
			id := acc.PlatformIdentifier
			if id == "" {
				id = acc.InstagramID
			}
			accountByID[id] = acc
		}

		log.Info().
			Str("platform", "instagram").
			Int64("total_accounts", total).
			Int64("accounts_processed", processed.Load()).
			Int64("errors", errCount.Load()).
			Int64("total_posts_found", totalPosts.Load()).
			Int64("total_urls_resolved", totalResolved.Load()).
			Int64("total_rows_updated", totalUpdated.Load()).
			Msg("Instagram URL refresh progress")

		for _, igID := range batch {
			acc, ok := accountByID[igID]
			if !ok {
				log.Warn().Str("platform", "instagram").Str("instagram_id", igID).Msg("Account not found in MongoDB, skipping")
				if err := chMarkInstagramRefreshedFn(chSink, ctx, igID); err != nil {
					log.Error().Err(err).Str("instagram_id", igID).Msg("Failed to mark Instagram posts refreshed after skip")
				}
				processed.Add(1)
				continue
			}

			token := resolveToken(acc, cfg.DecryptionKey)
			if strings.TrimSpace(token) == "" {
				log.Warn().Str("platform", "instagram").Str("instagram_id", igID).Msg("No token found, skipping")
				if err := chMarkInstagramRefreshedFn(chSink, ctx, igID); err != nil {
					log.Error().Err(err).Str("instagram_id", igID).Msg("Failed to mark Instagram posts refreshed after skip")
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

				accLogZ := log.With().
					Str("platform", "instagram").
					Str("instagram_id", igID).
					Logger()
				accLog := &accLogZ

				connectedViaIG := getBoolFromExtraData(acc.ExtraData, "connected_via_instagram")
				igClient := igClientFB
				if connectedViaIG {
					igClient = igClientIG
				}

				start := time.Now()
				result := processAccountJob(ctx, cfg, igClient, chSink, acc, accLog)
				dur := time.Since(start)

				processed.Add(1)
				if result.Err != nil {
					errCount.Add(1)
					accLog.Error().Err(result.Err).
						Int("posts_found", result.PostsFound).
						Dur("duration", dur).
						Msg("Account URL refresh failed")
					if social.IsAuthError(result.Err) {
						if markErr := chMarkInstagramRefreshedFn(chSink, ctx, igID); markErr != nil {
							accLog.Error().Err(markErr).Msg("Failed to mark Instagram posts refreshed after auth error")
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
				}
				successfulIDs = append(successfulIDs, igID)
				mu.Unlock()
			}()
		}

		wg.Wait()

		if len(allRefreshed) > 0 {
			n, err := chBulkUpdateFn(chSink, ctx, allRefreshed)
			if err != nil {
				log.Error().Err(err).Int("attempted", len(allRefreshed)).Msg("Failed to bulk update Instagram media URLs in ClickHouse")
			} else {
				totalUpdated.Add(int64(n))
			}
			allRefreshed = nil
		}
	}

	if len(successfulIDs) > 0 {
		if markErr := chBulkMarkInstagramRefreshedFn(chSink, ctx, successfulIDs); markErr != nil {
			log.Error().Err(markErr).Int("count", len(successfulIDs)).Msg("Failed to bulk mark Instagram posts refreshed")
		} else {
			log.Info().Int("count", len(successfulIDs)).Msg("Bulk marked Instagram posts refreshed")
		}
	}

	log.Info().
		Str("platform", "instagram").
		Int64("total_accounts", total).
		Int64("accounts_processed", processed.Load()).
		Int64("errors", errCount.Load()).
		Int64("total_posts_found", totalPosts.Load()).
		Int64("total_urls_resolved", totalResolved.Load()).
		Int64("total_rows_updated", totalUpdated.Load()).
		Msg("Instagram URL refresh completed")
}

func processAccountJob(
	ctx context.Context,
	cfg *config.Config,
	igClient *social.InstagramClient,
	chSink *conversions.ClickHouseSink,
	account mongomodels.SocialIntegration,
	log *zerolog.Logger,
) AccountResult {
	accountID := account.PlatformIdentifier
	if accountID == "" {
		accountID = account.InstagramID
	}
	res := AccountResult{InstagramID: accountID}

	for offset := 0; ; offset += postBatchSize {
		posts, err := chGetPostsFn(chSink, ctx, accountID, postBatchSize, offset)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get Instagram posts eligible for URL refresh")
			res.Err = err
			return res
		}
		if len(posts) == 0 {
			break
		}
		res.PostsFound += len(posts)

		refreshed, err := igGetURLsFn(igClient, ctx, accountID, account, cfg.DecryptionKey, posts)
		if err != nil {
			log.Error().Err(err).Msg("Failed to refresh Instagram media URLs")
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

func resolveToken(account mongomodels.SocialIntegration, decryptionKey string) string {
	connectedViaIG := getBoolFromExtraData(account.ExtraData, "connected_via_instagram")

	var raw string
	if connectedViaIG {
		raw = account.AccessToken
		if raw == "" {
			raw = getStringFromExtraData(account.ExtraData, "access_token")
		}
	} else {
		if account.UserDetails != nil {
			if details, ok := account.UserDetails.(map[string]interface{}); ok {
				if token, exists := details["access_token"].(string); exists && strings.TrimSpace(token) != "" {
					raw = token
				}
			}
		}
		if raw == "" && strings.TrimSpace(account.LongAccessToken) != "" {
			raw = account.LongAccessToken
		}
		if raw == "" {
			raw = account.AccessToken
		}
	}

	return decryptIfNeeded(raw, decryptionKey)
}

func decryptIfNeeded(token, decryptionKey string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	if decryptionKey != "" {
		if decrypted, err := crypto.DecryptToken(token, decryptionKey); err == nil && strings.TrimSpace(decrypted) != "" {
			return decrypted
		}
	}
	return token
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

func getBoolFromExtraData(extraData map[string]interface{}, key string) bool {
	if extraData == nil {
		return false
	}
	if value, ok := extraData[key].(bool); ok {
		return value
	}
	return false
}
