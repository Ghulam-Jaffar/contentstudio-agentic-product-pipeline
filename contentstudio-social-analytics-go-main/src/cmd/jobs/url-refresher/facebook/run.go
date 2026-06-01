package facebook

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
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/rs/zerolog"
)

const mongoBatchSize int64 = 50
const postBatchSize = 500

var mongoGetValidPageIDsFn = func(
	repo mongodb.UnifiedSocialRepository,
	ctx context.Context,
) ([]string, error) {
	accounts, err := repo.GetValidAccounts(ctx, mongomodels.PlatformFacebook, nil)
	if err != nil {
		return nil, err
	}
	ids := make([]string, 0, len(accounts))
	for _, acc := range accounts {
		if acc.PlatformIdentifier != "" {
			ids = append(ids, acc.PlatformIdentifier)
		}
	}
	return ids, nil
}

var chGetDistinctPageIDsFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	tableName string,
	validPageIDs []string,
) ([]string, error) {
	return s.ClickhouseClient.GetDistinctFacebookPageIDsWithStaleURLs(ctx, tableName, validPageIDs)
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
	pageID string,
	limit, offset int,
) ([]clickhouse.MinimalPost, error) {
	return s.ClickhouseClient.GetMinimalOlderThan20DaysByPage(ctx, "facebook_posts", pageID, limit, offset)
}

var fbGetThumbsFn = func(
	c *social.FacebookClient,
	ctx context.Context,
	pageID, accessToken, longAccessToken, decryptionKey string,
	posts []clickhouse.MinimalPost,
) ([]clickhouse.MinimalPost, error) {
	return c.GetPostThumbnails(ctx, pageID, accessToken, longAccessToken, decryptionKey, posts)
}

var chBulkUpdateFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	thumbs []clickhouse.MinimalPost,
) (int, error) {
	return s.ClickhouseClient.BulkUpdateFullPictures(ctx, "facebook_posts", thumbs)
}

var chMarkPageRefreshedFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	pageID string,
) error {
	return s.ClickhouseClient.MarkFacebookPostsRefreshed(ctx, "facebook_posts", pageID)
}

var chBulkMarkPageRefreshedFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	pageIDs []string,
) error {
	return s.ClickhouseClient.BulkMarkFacebookPostsRefreshed(ctx, "facebook_posts", pageIDs)
}

func Run(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	repo mongodb.UnifiedSocialRepository,
	fbClient *social.FacebookClient,
	chSink *conversions.ClickHouseSink,
	concurrency int,
	accountType string,
) {
	validPageIDs, err := mongoGetValidPageIDsFn(repo, ctx)
	if err != nil {
		log.Error().Err(err).Str("platform", "facebook").Msg("Failed to get valid Facebook account IDs from MongoDB")
		return
	}
	if len(validPageIDs) == 0 {
		log.Info().Str("platform", "facebook").Msg("No valid Facebook accounts found in MongoDB")
		return
	}
	log.Info().Str("platform", "facebook").Int("valid_accounts", len(validPageIDs)).Msg("Fetched valid Facebook account IDs from MongoDB")

	pageIDs, err := chGetDistinctPageIDsFn(chSink, ctx, "facebook_posts", validPageIDs)
	if err != nil {
		log.Error().Err(err).Str("platform", "facebook").Msg("Failed to get distinct page IDs with stale URLs")
		return
	}
	if len(pageIDs) == 0 {
		log.Info().Str("platform", "facebook").Msg("No Facebook posts with stale URLs found")
		return
	}

	if concurrency <= 0 {
		concurrency = 30
	}

	total := int64(len(pageIDs))
	log.Info().
		Str("platform", "facebook").
		Int64("total_accounts", total).
		Int("concurrency", concurrency).
		Msg("Starting Facebook URL refresh")

	sem := semaphore.NewWeighted(int64(concurrency))
	var wg sync.WaitGroup

	var (
		processed      atomic.Int64
		totalPosts     atomic.Int64
		totalThumbs    atomic.Int64
		totalUpdated   atomic.Int64
		errCount       atomic.Int64
		allThumbs      []clickhouse.MinimalPost
		zeroResultIDs  []string
		mu             sync.Mutex
	)

	batchSize := int(mongoBatchSize)
	for i := 0; i < len(pageIDs); i += batchSize {
		if ctx.Err() != nil {
			break
		}

		end := i + batchSize
		if end > len(pageIDs) {
			end = len(pageIDs)
		}
		batch := pageIDs[i:end]

		accounts, err := mongoGetAccountsByIDsFn(repo, ctx, mongomodels.PlatformFacebook, batch)
		if err != nil {
			log.Error().Err(err).Str("platform", "facebook").Msg("Failed to fetch accounts batch from MongoDB")
			break
		}

		accountByID := make(map[string]mongomodels.SocialIntegration, len(accounts))
		for _, acc := range accounts {
			accountByID[acc.PlatformIdentifier] = acc
		}

		log.Info().
			Str("platform", "facebook").
			Int64("total_accounts", total).
			Int64("accounts_processed", processed.Load()).
			Int64("errors", errCount.Load()).
			Int64("total_posts_found", totalPosts.Load()).
			Int64("total_thumbs_resolved", totalThumbs.Load()).
			Msg("Facebook URL refresh progress")

		for _, pageID := range batch {
			acc, ok := accountByID[pageID]
			if !ok {
				log.Warn().Str("platform", "facebook").Str("page_id", pageID).Msg("Account not found in MongoDB, skipping")
				if err := chMarkPageRefreshedFn(chSink, ctx, pageID); err != nil {
					log.Error().Err(err).Str("page_id", pageID).Msg("Failed to mark Facebook posts refreshed after skip")
				}
				processed.Add(1)
				continue
			}

			accessToken := getStringFromExtraData(acc.ExtraData, "access_token")
			longAccessToken := acc.GetAccessToken()
			if accessToken == "" && longAccessToken == "" {
				accessToken = acc.AccessToken
			}
			if strings.TrimSpace(accessToken) == "" && strings.TrimSpace(longAccessToken) == "" {
				log.Warn().Str("platform", "facebook").Str("page_id", pageID).Msg("No token found, skipping")
				if err := chMarkPageRefreshedFn(chSink, ctx, pageID); err != nil {
					log.Error().Err(err).Str("page_id", pageID).Msg("Failed to mark Facebook posts refreshed after skip")
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
					Str("platform", "facebook").
					Str("page_id", pageID).
					Logger()
				accLog := &accLogZ

				start := time.Now()
				result := processAccountJob(ctx, cfg, fbClient, chSink, acc, accLog)
				dur := time.Since(start)

				processed.Add(1)
				if result.Err != nil {
					errCount.Add(1)
					accLog.Error().Err(result.Err).
						Int("posts_found", result.PostsFound).
						Dur("duration", dur).
						Msg("Account URL refresh failed")
					if social.IsFacebookAuthError(result.Err) {
						if markErr := chMarkPageRefreshedFn(chSink, ctx, pageID); markErr != nil {
							accLog.Error().Err(markErr).Msg("Failed to mark Facebook posts refreshed after auth error")
						}
					}
					return
				}

				totalPosts.Add(int64(result.PostsFound))
				totalThumbs.Add(int64(result.ThumbsResolved))
				accLog.Debug().
					Int("posts_found", result.PostsFound).
					Int("thumbs_resolved", result.ThumbsResolved).
					Dur("duration", dur).
					Msg("Account URL refresh done")

				mu.Lock()
				if len(result.Thumbs) > 0 {
					allThumbs = append(allThumbs, result.Thumbs...)
				} else {
					zeroResultIDs = append(zeroResultIDs, pageID)
				}
				mu.Unlock()
			}()
		}

		wg.Wait()

		if len(allThumbs) > 0 {
			n, err := chBulkUpdateFn(chSink, ctx, allThumbs)
			if err != nil {
				log.Error().Err(err).Int("attempted", len(allThumbs)).Msg("Failed to bulk update Facebook post thumbnails in ClickHouse")
			} else {
				totalUpdated.Add(int64(n))
			}
			allThumbs = nil
		}
	}

	if len(zeroResultIDs) > 0 {
		if markErr := chBulkMarkPageRefreshedFn(chSink, ctx, zeroResultIDs); markErr != nil {
			log.Error().Err(markErr).Int("count", len(zeroResultIDs)).Msg("Failed to bulk mark Facebook posts refreshed")
		} else {
			log.Info().Int("count", len(zeroResultIDs)).Msg("Bulk marked Facebook posts refreshed")
		}
	}

	log.Info().
		Str("platform", "facebook").
		Int64("total_accounts", total).
		Int64("accounts_processed", processed.Load()).
		Int64("errors", errCount.Load()).
		Int64("total_posts_found", totalPosts.Load()).
		Int64("total_thumbs_resolved", totalThumbs.Load()).
		Int64("total_rows_updated", totalUpdated.Load()).
		Msg("Facebook URL refresh completed")
}

type AccountResult struct {
	PageID         string
	PostsFound     int
	ThumbsResolved int
	Thumbs         []clickhouse.MinimalPost
	Err            error
}

func processAccountJob(
	ctx context.Context,
	cfg *config.Config,
	fbClient *social.FacebookClient,
	chSink *conversions.ClickHouseSink,
	account mongomodels.SocialIntegration,
	log *zerolog.Logger,
) AccountResult {
	res := AccountResult{PageID: account.PlatformIdentifier}

	accessToken := getStringFromExtraData(account.ExtraData, "access_token")
	longAccessToken := account.GetAccessToken()
	if accessToken == "" && longAccessToken == "" {
		accessToken = account.AccessToken
	}

	for offset := 0; ; offset += postBatchSize {
		posts, err := chGetPostsFn(chSink, ctx, account.PlatformIdentifier, postBatchSize, offset)
		if err != nil {
			log.Error().Err(err).Msg("Failed to get Facebook posts eligible for URL refresh")
			res.Err = err
			return res
		}
		if len(posts) == 0 {
			break
		}
		res.PostsFound += len(posts)

		thumbs, err := fbGetThumbsFn(
			fbClient,
			ctx,
			account.PlatformIdentifier,
			accessToken,
			longAccessToken,
			cfg.DecryptionKey,
			posts,
		)
		if err != nil {
			log.Error().Err(err).Msg("Failed to refresh Facebook post thumbnails")
			res.Err = err
			return res
		}
		res.ThumbsResolved += len(thumbs)
		res.Thumbs = append(res.Thumbs, thumbs...)

		if len(posts) < postBatchSize {
			break
		}
	}

	return res
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
