package twitter

import (
	"context"
	"fmt"
	"math"
	"time"

	"github.com/rs/zerolog"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/api/httputil"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/twitter"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/twitter"
)

// Service defines the Twitter analytics business operations exposed to the API layer.
type Service interface {
	GetPageAndPostsInsights(ctx context.Context, req *types.TwitterRequest) (*types.MetricsResponse, error)
	GetEngagementImpressionData(ctx context.Context, req *types.TwitterRequest) (*types.EngagementImpressionResponse, error)
	GetFollowersTrendData(ctx context.Context, req *types.TwitterRequest) (*types.FollowersTrendResponse, error)
	GetTopTweets(ctx context.Context, req *types.TweetsRequest) (*types.TopTweetsResponse, error)
	GetLeastTweets(ctx context.Context, req *types.TweetsRequest) (*types.LeastTweetsResponse, error)
	GetCreditsUsedCount(ctx context.Context, req *types.TwitterRequest) (*types.CreditsUsedResponse, error)
}

// TwitterAnalyticsService orchestrates ClickHouse and Mongo-backed Twitter analytics reads.
type TwitterAnalyticsService struct {
	repo         *repo.Repository
	jobsMetadata *mongo.Collection
	logger       zerolog.Logger
}

var _ Service = (*TwitterAnalyticsService)(nil)

// NewTwitterAnalyticsService constructs a Twitter analytics service with optional Mongo metadata access.
func NewTwitterAnalyticsService(r *repo.Repository, mongoDB *mongo.Database, logger zerolog.Logger) *TwitterAnalyticsService {
	var jobsCollection *mongo.Collection
	if mongoDB != nil {
		jobsCollection = mongoDB.Collection("twitter_jobs_metadata")
	}

	return &TwitterAnalyticsService{
		repo:         r,
		jobsMetadata: jobsCollection,
		logger:       logger.With().Str("service", "twitter-analytics").Logger(),
	}
}

// GetPageAndPostsInsights fetches current and previous summary windows concurrently
// and computes period-over-period growth values for the overview cards response.
func (s *TwitterAnalyticsService) GetPageAndPostsInsights(ctx context.Context, req *types.TwitterRequest) (*types.MetricsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)

	var current, previous *repo.SummaryResult
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetSummary(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPageAndPostsInsights: failed to get current summary")
			r = &repo.SummaryResult{}
		}
		current = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetSummary(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPageAndPostsInsights: failed to get previous summary")
			r = &repo.SummaryResult{}
		}
		previous = r
		return nil
	})
	_ = eg.Wait()

	data := map[string]interface{}{
		"twitter_id":              current.TwitterID,
		"tiktok_id":               current.TwitterID, // backward-compatible typo used by PHP response
		"page_name":               current.Name,
		"logo":                    current.ProfileImageURL,
		"followers_count":         current.FollowersCount,
		"followers_count_growth":  calculateGrowth(current.FollowersCount, previous.FollowersCount),
		"followers_count_diff":    calculateDiff(current.FollowersCount, previous.FollowersCount),
		"following_count":         current.FollowingCount,
		"following_count_growth":  calculateGrowth(current.FollowingCount, previous.FollowingCount),
		"following_count_diff":    calculateDiff(current.FollowingCount, previous.FollowingCount),
		"tweet_count":             current.TweetCount,
		"tweet_count_growth":      calculateGrowth(current.TweetCount, previous.TweetCount),
		"tweet_count_diff":        current.PostsTweetCount,
		"listed_count":            current.ListedCount,
		"listed_count_growth":     calculateGrowth(current.ListedCount, previous.ListedCount),
		"listed_count_diff":       calculateDiff(current.ListedCount, previous.ListedCount),
		"total_engagement":        current.TotalEngagement,
		"total_engagement_growth": calculateGrowth(current.TotalEngagement, previous.TotalEngagement),
		"total_engagement_diff":   current.TotalEngagement - previous.TotalEngagement,
		"reply_count":             current.ReplyCount,
		"reply_count_growth":      calculateGrowth(current.ReplyCount, previous.ReplyCount),
		"reply_count_diff":        current.ReplyCount - previous.ReplyCount,
		"retweet_count":           current.RetweetCount,
		"retweet_count_growth":    calculateGrowth(current.RetweetCount, previous.RetweetCount),
		"retweet_count_diff":      current.RetweetCount - previous.RetweetCount,
		"bookmark_count":          current.BookmarkCount,
		"bookmark_count_growth":   calculateGrowth(current.BookmarkCount, previous.BookmarkCount),
		"bookmark_count_diff":     current.BookmarkCount - previous.BookmarkCount,
		"quote_count":             current.QuoteCount,
		"quote_count_growth":      calculateGrowth(current.QuoteCount, previous.QuoteCount),
		"quote_count_diff":        current.QuoteCount - previous.QuoteCount,
		"like_count":              current.LikeCount,
		"like_count_growth":       calculateGrowth(current.LikeCount, previous.LikeCount),
		"like_count_diff":         current.LikeCount - previous.LikeCount,
	}

	return &types.MetricsResponse{Status: true, Data: data}, nil
}

// GetEngagementImpressionData returns the time-series dataset used by the engagement vs impressions chart.
func (s *TwitterAnalyticsService) GetEngagementImpressionData(ctx context.Context, req *types.TwitterRequest) (*types.EngagementImpressionResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	result, err := s.repo.GetEngagementImpressionData(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetEngagementImpressionData: failed to get data")
		return &types.EngagementImpressionResponse{Status: true}, nil
	}

	return &types.EngagementImpressionResponse{
		Status:          true,
		TwitterID:       result.TwitterID,
		TweetCount:      result.TweetCount,
		ImpressionCount: result.ImpressionCount,
		TotalEngagement: result.TotalEngagement,
		RetweetCount:    result.RetweetCount,
		ReplyCount:      result.ReplyCount,
		LikeCount:       result.LikeCount,
		BookmarkCount:   result.BookmarkCount,
		QuoteCount:      result.QuoteCount,
		TweetedAtDate:   result.TweetedAtDate,
	}, nil
}

// GetFollowersTrendData returns follower/following trend buckets for the selected period.
func (s *TwitterAnalyticsService) GetFollowersTrendData(ctx context.Context, req *types.TwitterRequest) (*types.FollowersTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	result, err := s.repo.GetFollowersTrend(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetFollowersTrendData: failed to get data")
		return &types.FollowersTrendResponse{Status: true}, nil
	}

	return &types.FollowersTrendResponse{
		Status:              true,
		PlatformID:          result.PlatformID,
		Name:                result.Name,
		Username:            result.Username,
		FollowerCount:       result.FollowerCount,
		FollowerCountDaily:  result.FollowerCountDaily,
		FollowingCount:      result.FollowingCount,
		FollowingCountDaily: result.FollowingCountDaily,
		Buckets:             result.Buckets,
	}, nil
}

// GetTopTweets returns the highest-performing tweets for the selected metric and limit.
func (s *TwitterAnalyticsService) GetTopTweets(ctx context.Context, req *types.TweetsRequest) (*types.TopTweetsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetTweetsData(ctx, params, req.GetOrderBy(), req.GetLimit(), "DESC")
	if err != nil {
		s.logger.Error().Err(err).Msg("GetTopTweets: failed to get data")
		return &types.TopTweetsResponse{Status: true, TopTweets: []types.Tweet{}}, nil
	}

	return &types.TopTweetsResponse{Status: true, TopTweets: mapTweets(rows)}, nil
}

// GetLeastTweets returns the lowest-performing tweets using the same sorting rules as GetTopTweets.
func (s *TwitterAnalyticsService) GetLeastTweets(ctx context.Context, req *types.TweetsRequest) (*types.LeastTweetsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetTweetsData(ctx, params, req.GetOrderBy(), req.GetLimit(), "ASC")
	if err != nil {
		s.logger.Error().Err(err).Msg("GetLeastTweets: failed to get data")
		return &types.LeastTweetsResponse{Status: true, LeastTweets: []types.Tweet{}}, nil
	}

	return &types.LeastTweetsResponse{Status: true, LeastTweets: mapTweets(rows)}, nil
}

// GetCreditsUsedCount aggregates credits consumed by Twitter jobs for the requested date range.
func (s *TwitterAnalyticsService) GetCreditsUsedCount(ctx context.Context, req *types.TwitterRequest) (*types.CreditsUsedResponse, error) {
	if s.jobsMetadata == nil {
		return nil, httputil.NewInternalError("twitter jobs metadata collection not configured")
	}

	start, end, err := types.BuildDateTimeRange(req.StartDate, req.EndDate)
	if err != nil {
		return nil, httputil.NewValidationError(err.Error())
	}

	match := bson.M{
		"workspace_id": req.WorkspaceID,
		"platform_id":  req.TwitterID,
		"job_executed_at": bson.M{
			"$gte": start,
			"$lte": end,
		},
	}

	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: match}},
		bson.D{{Key: "$group", Value: bson.M{"_id": nil, "credits_used": bson.M{"$sum": "$credits_used"}}}},
	}

	cursor, err := s.jobsMetadata.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("GetCreditsUsedCount aggregate: %w", err)
	}
	defer cursor.Close(ctx)

	creditsUsed := int64(0)
	if cursor.Next(ctx) {
		var row struct {
			CreditsUsed int64 `bson:"credits_used"`
		}
		if err := cursor.Decode(&row); err != nil {
			return nil, fmt.Errorf("GetCreditsUsedCount decode: %w", err)
		}
		creditsUsed = row.CreditsUsed
	}
	if err := cursor.Err(); err != nil {
		return nil, fmt.Errorf("GetCreditsUsedCount cursor: %w", err)
	}

	return &types.CreditsUsedResponse{
		Status: true,
		Data: types.CreditsUsedData{
			CreditsUsed: creditsUsed,
			StartDate:   start.Format(time.RFC3339),
			EndDate:     end.Format(time.RFC3339),
			WorkspaceID: req.WorkspaceID,
			TwitterID:   req.TwitterID,
		},
	}, nil
}

// prevPeriodParams rewrites the query params to point at the precomputed comparison window.
func (s *TwitterAnalyticsService) prevPeriodParams(params *clickhouse.QueryParams) *clickhouse.QueryParams {
	prev := *params
	prev.DateFrom = params.PrevDateFrom
	prev.DateTo = params.PrevDateTo
	return &prev
}

func calculateDiff(current, previous int64) interface{} {
	if previous == 0 {
		return "N/A"
	}
	return current - previous
}

func calculateGrowth(current, previous int64) interface{} {
	if previous == 0 {
		return "N/A"
	}
	growth := float64(current-previous) / float64(previous)
	return math.Round(growth*10000) / 100
}

func mapTweets(rows []repo.TweetRow) []types.Tweet {
	result := make([]types.Tweet, 0, len(rows))
	for _, row := range rows {
		result = append(result, types.Tweet{
			ID:              row.ID,
			TweetedAt:       row.TweetedAt,
			TweetText:       row.TweetText,
			TweetType:       row.TweetType,
			Permalink:       row.Permalink,
			MediaURL:        row.MediaURL,
			ListedCount:     row.ListedCount,
			RetweetCount:    row.RetweetCount,
			LikeCount:       row.LikeCount,
			ReplyCount:      row.ReplyCount,
			QuoteCount:      row.QuoteCount,
			BookmarkCount:   row.BookmarkCount,
			ImpressionCount: row.ImpressionCount,
			TotalEngagement: row.TotalEngagement,
		})
	}
	return result
}
