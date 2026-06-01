package facebook_competitor

import (
	"context"
	"sync"
	"time"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/clients/social"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse/conversions"
	mongomodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/utils/competitortokens"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
)

const tokenQueueFacebook = "facebook_valid_token_set"
const postBatchSize = 500

type CompetitorRepository interface {
	GetAccounts(ctx context.Context, platformType string) ([]*mongomodels.Competitor, error)
}

type RedisClient interface {
	Do(ctx context.Context, args ...interface{}) *redis.Cmd
}

type AccountJob struct {
	Account *mongomodels.Competitor
}

type AccountResult struct {
	FacebookID          string
	PostsFound          int
	SharedPostsFound    int
	URLsResolved        int
	SharedPicsResolved  int
	ResolvedAssets      []clickhousemodels.FacebookCompetitorMinimalMediaAsset
	ResolvedSharedPics  []clickhousemodels.FacebookCompetitorMinimalSharedPost
	Err                 error
	Duration            time.Duration
}

var chGetPostsFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	facebookID string,
	limit, offset int,
) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
	return s.ClickhouseClient.GetMinimalFacebookCompetitorMediaAssetsOlderThan7DaysByAccount(ctx, "facebook_competitor_media_assets", facebookID, limit, offset)
}

var fbGetURLsFn = func(
	c *social.FacebookClient,
	ctx context.Context,
	facebookID string,
	accessToken string,
	assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset,
) ([]clickhousemodels.FacebookCompetitorMinimalMediaAsset, error) {
	return c.GetCompetitorMediaAssetURLs(ctx, facebookID, accessToken, assets)
}

var chGetSharedPostsFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	facebookID string,
	limit, offset int,
) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
	return s.ClickhouseClient.GetMinimalFacebookCompetitorSharedPostsOlderThan7DaysByAccount(ctx, "facebook_competitor_posts", facebookID, limit, offset)
}

var fbGetSharedPicsFn = func(
	c *social.FacebookClient,
	ctx context.Context,
	facebookID string,
	accessToken string,
	posts []clickhousemodels.FacebookCompetitorMinimalSharedPost,
) ([]clickhousemodels.FacebookCompetitorMinimalSharedPost, error) {
	return c.GetCompetitorSharedFromPictures(ctx, facebookID, accessToken, posts)
}

var chBulkUpdateAssetsFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	assets []clickhousemodels.FacebookCompetitorMinimalMediaAsset,
) (int, error) {
	return s.ClickhouseClient.BulkUpdateFacebookCompetitorMediaAssetURLs(ctx, "facebook_competitor_media_assets", assets)
}

var chBulkUpdateSharedPicsFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	posts []clickhousemodels.FacebookCompetitorMinimalSharedPost,
) (int, error) {
	return s.ClickhouseClient.BulkUpdateFacebookCompetitorSharedPictures(ctx, "facebook_competitor_posts", posts)
}

func Run(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	repo CompetitorRepository,
	redisClient RedisClient,
	fbClient *social.FacebookClient,
	chSink *conversions.ClickHouseSink,
	workerCount int,
) {
	accounts, err := repo.GetAccounts(ctx, "facebook")
	if err != nil {
		log.Error().Err(err).Str("platform", "facebook-competitor").Msg("Error fetching valid Facebook competitor accounts")
		return
	}
	if len(accounts) == 0 {
		log.Info().Str("platform", "facebook-competitor").Msg("No valid Facebook competitor accounts found to process")
		return
	}

	if workerCount <= 0 {
		workerCount = 10
	}
	if len(accounts) < workerCount {
		workerCount = len(accounts)
	}

	jobs := make(chan AccountJob)
	results := make(chan AccountResult, workerCount*2)

	var wg sync.WaitGroup
	wg.Add(workerCount)
	for i := 0; i < workerCount; i++ {
		go accountWorker(ctx, i+1, cfg, log, redisClient, fbClient, chSink, jobs, results, &wg)
	}

	go func() {
		for _, acc := range accounts {
			jobs <- AccountJob{Account: acc}
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	totalPosts := 0
	totalResolved := 0
	totalSharedPosts := 0
	totalSharedResolved := 0
	errCount := 0

	var allAssets []clickhousemodels.FacebookCompetitorMinimalMediaAsset
	var allSharedPics []clickhousemodels.FacebookCompetitorMinimalSharedPost

	for result := range results {
		if result.Err != nil {
			errCount++
			log.Error().
				Str("platform", "facebook-competitor").
				Str("facebook_id", result.FacebookID).
				Err(result.Err).
				Int("posts_found", result.PostsFound).
				Int("shared_posts_found", result.SharedPostsFound).
				Int("urls_resolved", result.URLsResolved).
				Int("shared_pics_resolved", result.SharedPicsResolved).
				Dur("duration", result.Duration).
				Msg("Competitor processing failed")
			continue
		}
		totalPosts += result.PostsFound
		totalSharedPosts += result.SharedPostsFound
		totalResolved += result.URLsResolved
		totalSharedResolved += result.SharedPicsResolved
		allAssets = append(allAssets, result.ResolvedAssets...)
		allSharedPics = append(allSharedPics, result.ResolvedSharedPics...)
	}

	// All goroutines done — issue one bulk mutation for each asset type.
	totalUpdated := 0
	totalSharedUpdated := 0

	if len(allAssets) > 0 {
		updated, err := chBulkUpdateAssetsFn(chSink, ctx, allAssets)
		if err != nil {
			log.Error().Err(err).Int("assets", len(allAssets)).Msg("Bulk update of Facebook competitor media asset URLs failed")
		} else {
			totalUpdated = updated
		}
	}

	if len(allSharedPics) > 0 {
		updated, err := chBulkUpdateSharedPicsFn(chSink, ctx, allSharedPics)
		if err != nil {
			log.Error().Err(err).Int("shared_pics", len(allSharedPics)).Msg("Bulk update of Facebook competitor shared pictures failed")
		} else {
			totalSharedUpdated = updated
		}
	}

	log.Info().
		Str("platform", "facebook-competitor").
		Int("accounts", len(accounts)).
		Int("errors", errCount).
		Int("total_posts_found", totalPosts).
		Int("total_shared_posts_found", totalSharedPosts).
		Int("total_urls_resolved", totalResolved).
		Int("total_shared_pics_resolved", totalSharedResolved).
		Int("total_rows_updated", totalUpdated).
		Int("total_shared_pics_updated", totalSharedUpdated).
		Msg("Facebook competitor URL refresh completed")
}

func accountWorker(
	ctx context.Context,
	workerID int,
	cfg *config.Config,
	log logger.Logger,
	redisClient RedisClient,
	fbClient *social.FacebookClient,
	chSink *conversions.ClickHouseSink,
	jobs <-chan AccountJob,
	results chan<- AccountResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for job := range jobs {
		facebookID := job.Account.GetCompetitorIDAsString()
		accLogZ := log.With().
			Str("component", "facebook_competitor_url_refresher.account_runner").
			Int("worker_id", workerID).
			Str("facebook_id", facebookID).
			Logger()
		accLog := &accLogZ

		start := time.Now()
		result := processAccountJob(ctx, cfg, redisClient, fbClient, chSink, job.Account, accLog)
		result.Duration = time.Since(start)
		results <- result
	}
}

func processAccountJob(
	ctx context.Context,
	cfg *config.Config,
	redisClient RedisClient,
	fbClient *social.FacebookClient,
	chSink *conversions.ClickHouseSink,
	account *mongomodels.Competitor,
	log *zerolog.Logger,
) AccountResult {
	facebookID := account.GetCompetitorIDAsString()
	result := AccountResult{FacebookID: facebookID}

	tokenState := newFacebookTokenState(ctx, redisClient, cfg)
	var token string
	var tokenErr error
	tokenLoaded := false

	loadToken := func() error {
		if tokenLoaded {
			return nil
		}
		token, tokenErr = tokenState.Current()
		if tokenErr != nil {
			return tokenErr
		}
		tokenLoaded = true
		return nil
	}

	for offset := 0; ; offset += postBatchSize {
		posts, err := chGetPostsFn(chSink, ctx, facebookID, postBatchSize, offset)
		if err != nil {
			result.Err = err
			return result
		}
		if len(posts) == 0 {
			break
		}
		result.PostsFound += len(posts)

		if err := loadToken(); err != nil {
			result.Err = err
			return result
		}

		refreshed, err := fbGetURLsFn(fbClient, ctx, facebookID, token, posts)
		if err != nil && competitortokens.IsFacebookTokenIssue(err) {
			if freshToken, retryErr := tokenState.Refresh(); retryErr == nil {
				token = freshToken
				refreshed, err = fbGetURLsFn(fbClient, ctx, facebookID, token, posts)
			}
		}
		if err != nil {
			if competitortokens.IsFacebookTokenIssue(err) {
				log.Warn().Err(err).Str("facebook_id", facebookID).Msg("Skipping competitor after token retry failure")
				return result
			}
			result.Err = err
			return result
		}
		result.URLsResolved += len(refreshed)
		result.ResolvedAssets = append(result.ResolvedAssets, refreshed...)

		if len(posts) < postBatchSize {
			break
		}
	}

	for offset := 0; ; offset += postBatchSize {
		sharedPosts, err := chGetSharedPostsFn(chSink, ctx, facebookID, postBatchSize, offset)
		if err != nil {
			result.Err = err
			return result
		}
		if len(sharedPosts) == 0 {
			break
		}
		result.SharedPostsFound += len(sharedPosts)

		if err := loadToken(); err != nil {
			result.Err = err
			return result
		}

		refreshedSharedPics, err := fbGetSharedPicsFn(fbClient, ctx, facebookID, token, sharedPosts)
		if err != nil && competitortokens.IsFacebookTokenIssue(err) {
			if freshToken, retryErr := tokenState.Refresh(); retryErr == nil {
				token = freshToken
				refreshedSharedPics, err = fbGetSharedPicsFn(fbClient, ctx, facebookID, token, sharedPosts)
			}
		}
		if err != nil {
			if competitortokens.IsFacebookTokenIssue(err) {
				log.Warn().Err(err).Str("facebook_id", facebookID).Msg("Skipping competitor shared picture refresh after token retry failure")
				return result
			}
			result.Err = err
			return result
		}
		result.SharedPicsResolved += len(refreshedSharedPics)
		result.ResolvedSharedPics = append(result.ResolvedSharedPics, refreshedSharedPics...)

		if len(sharedPosts) < postBatchSize {
			break
		}
	}

	return result
}

type facebookTokenState struct {
	ctx         context.Context
	redisClient RedisClient
	cfg         *config.Config
	exclude     map[string]struct{}
	current     competitortokens.Candidate
	loaded      bool
}

func newFacebookTokenState(ctx context.Context, redisClient RedisClient, cfg *config.Config) *facebookTokenState {
	return &facebookTokenState{
		ctx:         ctx,
		redisClient: redisClient,
		cfg:         cfg,
		exclude:     make(map[string]struct{}, 2),
	}
}

func (s *facebookTokenState) Current() (string, error) {
	if s.loaded {
		return s.current.AccessToken, nil
	}
	return s.Refresh()
}

func (s *facebookTokenState) Refresh() (string, error) {
	candidate, err := competitortokens.FetchCandidate(s.ctx, s.redisClient, tokenQueueFacebook, s.cfg.DecryptionKey, false, s.exclude)
	if err != nil {
		return "", err
	}
	s.exclude[candidate.Key()] = struct{}{}
	s.current = candidate
	s.loaded = true
	return candidate.AccessToken, nil
}
