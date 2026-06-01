package instagram_competitor

import (
	"context"
	"fmt"
	"strconv"
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

const tokenQueueInstagram = "instagram_valid_token_set"
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
	InstagramID       int64
	PostsFound        int
	URLsResolved      int
	ProfilePictureURL string
	ResolvedPosts     []clickhousemodels.InstagramCompetitorMinimalPost
	Err               error
	Duration          time.Duration
}

var chGetPostsFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	instagramID int64,
	limit, offset int,
) ([]clickhousemodels.InstagramCompetitorMinimalPost, error) {
	return s.ClickhouseClient.GetMinimalInstagramCompetitorOlderThan7DaysByAccount(ctx, "instagram_competitor_posts", instagramID, limit, offset)
}

var igGetURLsFn = func(
	c *social.InstagramClient,
	ctx context.Context,
	username string,
	posts []clickhousemodels.InstagramCompetitorMinimalPost,
	accessToken string,
	businessAccountID string,
) ([]clickhousemodels.InstagramCompetitorMinimalPost, string, error) {
	return c.GetCompetitorMediaURLs(ctx, username, posts, accessToken, businessAccountID)
}

var chBulkUpdateFn = func(
	s *conversions.ClickHouseSink,
	ctx context.Context,
	posts []clickhousemodels.InstagramCompetitorMinimalPost,
	profilePics map[int64]string,
) (int, error) {
	return s.ClickhouseClient.BulkUpdateInstagramCompetitorMediaURLs(ctx, "instagram_competitor_posts", posts, profilePics)
}

func Run(
	ctx context.Context,
	cfg *config.Config,
	log logger.Logger,
	repo CompetitorRepository,
	redisClient RedisClient,
	igClient *social.InstagramClient,
	chSink *conversions.ClickHouseSink,
	workerCount int,
) {
	accounts, err := repo.GetAccounts(ctx, "instagram")
	if err != nil {
		log.Error().Err(err).Str("platform", "instagram-competitor").Msg("Error fetching valid Instagram competitor accounts")
		return
	}
	if len(accounts) == 0 {
		log.Info().Str("platform", "instagram-competitor").Msg("No valid Instagram competitor accounts found to process")
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
		go accountWorker(ctx, i+1, cfg, log, redisClient, igClient, chSink, jobs, results, &wg)
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
	errCount := 0

	var allPosts []clickhousemodels.InstagramCompetitorMinimalPost
	profilePics := make(map[int64]string)

	for result := range results {
		if result.Err != nil {
			errCount++
			log.Error().
				Str("platform", "instagram-competitor").
				Int64("instagram_id", result.InstagramID).
				Err(result.Err).
				Int("posts_found", result.PostsFound).
				Int("urls_resolved", result.URLsResolved).
				Dur("duration", result.Duration).
				Msg("Competitor processing failed")
			continue
		}
		totalPosts += result.PostsFound
		totalResolved += result.URLsResolved
		allPosts = append(allPosts, result.ResolvedPosts...)
		if result.ProfilePictureURL != "" {
			profilePics[result.InstagramID] = result.ProfilePictureURL
		}
	}

	// All goroutines done — issue bulk mutations for media URLs and profile pictures.
	totalUpdated := 0
	if len(allPosts) > 0 || len(profilePics) > 0 {
		updated, err := chBulkUpdateFn(chSink, ctx, allPosts, profilePics)
		if err != nil {
			log.Error().Err(err).
				Int("posts", len(allPosts)).
				Int("profile_pics", len(profilePics)).
				Msg("Bulk update of Instagram competitor media URLs failed")
		} else {
			totalUpdated = updated
		}
	}

	log.Info().
		Str("platform", "instagram-competitor").
		Int("accounts", len(accounts)).
		Int("errors", errCount).
		Int("total_posts_found", totalPosts).
		Int("total_urls_resolved", totalResolved).
		Int("total_rows_updated", totalUpdated).
		Msg("Instagram competitor URL refresh completed")
}

func accountWorker(
	ctx context.Context,
	workerID int,
	cfg *config.Config,
	log logger.Logger,
	redisClient RedisClient,
	igClient *social.InstagramClient,
	chSink *conversions.ClickHouseSink,
	jobs <-chan AccountJob,
	results chan<- AccountResult,
	wg *sync.WaitGroup,
) {
	defer wg.Done()

	for job := range jobs {
		instagramID, _ := strconv.ParseInt(job.Account.GetCompetitorIDAsString(), 10, 64)
		accLogZ := log.With().
			Str("component", "instagram_competitor_url_refresher.account_runner").
			Int("worker_id", workerID).
			Int64("instagram_id", instagramID).
			Logger()
		accLog := &accLogZ

		start := time.Now()
		result := processAccountJob(ctx, cfg, redisClient, igClient, chSink, job.Account, accLog)
		result.Duration = time.Since(start)
		results <- result
	}
}

func processAccountJob(
	ctx context.Context,
	cfg *config.Config,
	redisClient RedisClient,
	igClient *social.InstagramClient,
	chSink *conversions.ClickHouseSink,
	account *mongomodels.Competitor,
	log *zerolog.Logger,
) AccountResult {
	instagramID, err := strconv.ParseInt(account.GetCompetitorIDAsString(), 10, 64)
	result := AccountResult{InstagramID: instagramID}
	if err != nil || instagramID == 0 {
		result.Err = fmt.Errorf("invalid instagram competitor id: %s", account.GetCompetitorIDAsString())
		return result
	}

	tokenState := newInstagramTokenState(ctx, redisClient, cfg)
	var token string
	var businessAccountID string
	tokenLoaded := false

	for offset := 0; ; offset += postBatchSize {
		posts, err := chGetPostsFn(chSink, ctx, instagramID, postBatchSize, offset)
		if err != nil {
			result.Err = err
			return result
		}
		if len(posts) == 0 {
			break
		}
		result.PostsFound += len(posts)

		if !tokenLoaded {
			token, businessAccountID, err = tokenState.Current()
			if err != nil {
				result.Err = err
				return result
			}
			tokenLoaded = true
		}

		refreshed, profilePictureURL, err := igGetURLsFn(igClient, ctx, account.Slug, posts, token, businessAccountID)
		if err != nil && competitortokens.IsInstagramTokenIssue(err) {
			if freshToken, freshBusinessID, retryErr := tokenState.Refresh(); retryErr == nil {
				token = freshToken
				businessAccountID = freshBusinessID
				refreshed, profilePictureURL, err = igGetURLsFn(igClient, ctx, account.Slug, posts, token, businessAccountID)
			}
		}
		if err != nil {
			if competitortokens.IsInstagramTokenIssue(err) {
				log.Warn().Err(err).Int64("instagram_id", instagramID).Msg("Skipping competitor after token retry failure")
				return result
			}
			result.Err = err
			return result
		}

		result.URLsResolved += len(refreshed)
		result.ResolvedPosts = append(result.ResolvedPosts, refreshed...)
		if profilePictureURL != "" && result.ProfilePictureURL == "" {
			result.ProfilePictureURL = profilePictureURL
		}

		if len(posts) < postBatchSize {
			break
		}
	}

	return result
}

type instagramTokenState struct {
	ctx         context.Context
	redisClient RedisClient
	cfg         *config.Config
	exclude     map[string]struct{}
	current     competitortokens.Candidate
	loaded      bool
}

func newInstagramTokenState(ctx context.Context, redisClient RedisClient, cfg *config.Config) *instagramTokenState {
	return &instagramTokenState{
		ctx:         ctx,
		redisClient: redisClient,
		cfg:         cfg,
		exclude:     make(map[string]struct{}, 2),
	}
}

func (s *instagramTokenState) Current() (string, string, error) {
	if s.loaded {
		return s.current.AccessToken, s.current.PlatformID, nil
	}
	return s.Refresh()
}

func (s *instagramTokenState) Refresh() (string, string, error) {
	candidate, err := competitortokens.FetchCandidate(s.ctx, s.redisClient, tokenQueueInstagram, s.cfg.DecryptionKey, true, s.exclude)
	if err != nil {
		return "", "", err
	}
	s.exclude[candidate.Key()] = struct{}{}
	s.current = candidate
	s.loaded = true
	return candidate.AccessToken, candidate.PlatformID, nil
}
