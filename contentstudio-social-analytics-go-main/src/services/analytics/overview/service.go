// Package overview provides the business logic layer for cross-platform Overview analytics.
// It orchestrates ClickHouse repository queries and maps results to API response types.
//
// Migrated from PHP: OverviewV2Controller + OverviewV2Builder (contentstudio-backend).
package overview

import (
	"context"
	"math"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/overview"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/overview"
)

// Service defines the interface for Overview analytics business logic.
type Service interface {
	GetSummary(ctx context.Context, req *types.OverviewRequest) (*types.SummaryResponse, error)
	GetTopPerformingGraph(ctx context.Context, req *types.OverviewRequest) (*types.TopPerformingGraphResponse, error)
	GetPlatformDataGrouped(ctx context.Context, req *types.OverviewRequest) ([]*types.PlatformDataRow, error)
	GetPlatformDataIndividual(ctx context.Context, req *types.OverviewRequest) ([]*types.AccountDataRow, error)
	GetPlatformDataDetailed(ctx context.Context, req *types.OverviewRequest) ([]*types.AccountDataDetailedRow, error)
	GetPlatformDataGraphs(ctx context.Context, req *types.OverviewRequest) ([]*types.AccountDataGraphsRow, error)
	GetTopPosts(ctx context.Context, req *types.TopPostsRequest) ([]*types.TopPostRow, error)
}

// repoInterface abstracts the ClickHouse repository for testability.
type repoInterface interface {
	GetPlatformData(ctx context.Context, params *repo.OverviewParams) ([]repo.PlatformDataRow, error)
	GetAccountData(ctx context.Context, params *repo.OverviewParams) ([]repo.AccountDataRow, error)
	GetAccountDataDetailed(ctx context.Context, params *repo.OverviewParams) ([]repo.AccountDataDetailedRow, error)
	GetAccountDataGraphs(ctx context.Context, params *repo.OverviewParams) ([]repo.AccountDataGraphsRow, error)
	GetTopPosts(ctx context.Context, params *repo.OverviewParams) ([]repo.TopPostRow, error)
	GetTopPerformingGraph(ctx context.Context, params *repo.OverviewParams) (*repo.TopPerformingGraphResult, error)
}

// OverviewAnalyticsService implements Overview analytics business logic.
type OverviewAnalyticsService struct {
	repo   repoInterface
	logger zerolog.Logger
}

var _ Service = (*OverviewAnalyticsService)(nil)

// NewOverviewAnalyticsService creates a new service with the given repository and logger.
func NewOverviewAnalyticsService(r *repo.Repository, logger zerolog.Logger) *OverviewAnalyticsService {
	return &OverviewAnalyticsService{
		repo:   r,
		logger: logger.With().Str("service", "overview-analytics").Logger(),
	}
}

// buildParams constructs OverviewParams from the base request.
func buildParams(req *types.OverviewRequest) (*repo.OverviewParams, error) {
	return repo.NewOverviewParams(
		req.StartDate,
		req.EndDate,
		req.FacebookAccounts,
		req.InstagramAccounts,
		req.LinkedInAccounts,
		req.TiktokAccounts,
		req.PinterestAccounts,
		req.YouTubeAccounts,
		req.Timezone,
		"",
		0,
	)
}

// round2 rounds a float64 to 2 decimal places (matching PHP round($x, 2)).
func round2(x float64) float64 {
	return math.Round(x*100) / 100
}

// GetSummary fetches platform data for current and secondary periods concurrently,
// sums the metrics across platforms, then computes engagement rates, diffs, and pct changes.
// Mirrors PHP OverviewV2Controller.getSummary() which calls getPlatformDataQuery() twice.
func (s *OverviewAnalyticsService) GetSummary(ctx context.Context, req *types.OverviewRequest) (*types.SummaryResponse, error) {
	params, err := buildParams(req)
	if err != nil {
		return nil, err
	}

	secParams := params.NewSecondaryParams()

	var currentRows, secRows []repo.PlatformDataRow
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		rows, err := s.repo.GetPlatformData(egCtx, params)
		if err != nil {
			return err
		}
		currentRows = rows
		return nil
	})
	eg.Go(func() error {
		rows, err := s.repo.GetPlatformData(egCtx, secParams)
		if err != nil {
			return err
		}
		secRows = rows
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	var followers, posts, engagement, impressions, reach int64
	for _, row := range currentRows {
		followers += int64(row.Followers)
		posts += int64(row.TotalPosts)
		engagement += int64(row.Engagement)
		impressions += int64(row.Impressions)
		reach += int64(row.Reach)
	}

	var secFollowers, secPosts, secEngagement, secImpressions, secReach int64
	for _, row := range secRows {
		secFollowers += int64(row.Followers)
		secPosts += int64(row.TotalPosts)
		secEngagement += int64(row.Engagement)
		secImpressions += int64(row.Impressions)
		secReach += int64(row.Reach)
	}

	engagementRate := 0.0
	if impressions > 0 {
		engagementRate = round2(float64(engagement) / float64(impressions) * 100)
	}
	secEngagementRate := 0.0
	if secImpressions > 0 {
		secEngagementRate = round2(float64(secEngagement) / float64(secImpressions) * 100)
	}

	diffFollowers := int64(0)
	if secFollowers > 0 {
		diffFollowers = followers - secFollowers
	}
	diffPosts := int64(0)
	if secPosts > 0 {
		diffPosts = posts - secPosts
	}
	diffEngagement := int64(0)
	if secEngagement > 0 {
		diffEngagement = engagement - secEngagement
	}
	diffImpressions := int64(0)
	if secImpressions > 0 {
		diffImpressions = impressions - secImpressions
	}
	diffReach := int64(0)
	if secReach > 0 {
		diffReach = reach - secReach
	}
	diffEngagementRate := 0.0
	if secEngagementRate > 0 {
		diffEngagementRate = round2(engagementRate - secEngagementRate)
	}

	followersChangePct := 0.0
	if secFollowers > 0 {
		followersChangePct = round2(float64(followers-secFollowers) / float64(secFollowers) * 100)
	}
	postsChangePct := 0.0
	if secPosts > 0 {
		postsChangePct = round2(float64(posts-secPosts) / float64(secPosts) * 100)
	}
	engagementChangePct := 0.0
	if secEngagement > 0 {
		engagementChangePct = round2(float64(engagement-secEngagement) / float64(secEngagement) * 100)
	}
	impressionsChangePct := 0.0
	if secImpressions > 0 {
		impressionsChangePct = round2(float64(impressions-secImpressions) / float64(secImpressions) * 100)
	}
	reachChangePct := 0.0
	if secReach > 0 {
		reachChangePct = round2(float64(reach-secReach) / float64(secReach) * 100)
	}
	engagementRateChangePct := 0.0
	if secEngagementRate > 0 {
		engagementRateChangePct = round2(float64(engagementRate-secEngagementRate) / float64(secEngagementRate) * 100)
	}

	return &types.SummaryResponse{
		Summary: &types.SummaryData{
			Followers:               followers,
			Posts:                   posts,
			Engagement:              engagement,
			Impressions:             impressions,
			Reach:                   reach,
			EngagementRate:          engagementRate,
			SecondaryFollowers:      secFollowers,
			SecondaryPosts:          secPosts,
			SecondaryEngagement:     secEngagement,
			SecondaryImpressions:    secImpressions,
			SecondaryReach:          secReach,
			SecondaryEngagementRate: secEngagementRate,
			DiffFollowers:           diffFollowers,
			DiffPosts:               diffPosts,
			DiffEngagement:          diffEngagement,
			DiffImpressions:         diffImpressions,
			DiffReach:               diffReach,
			DiffEngagementRate:      diffEngagementRate,
			FollowersChangePct:      followersChangePct,
			PostsChangePct:          postsChangePct,
			EngagementChangePct:     engagementChangePct,
			ImpressionsChangePct:    impressionsChangePct,
			ReachChangePct:          reachChangePct,
			EngagementRateChangePct: engagementRateChangePct,
		},
	}, nil
}

// GetTopPerformingGraph fetches time-series data from mv_social_daily_metrics and formats bucket dates.
func (s *OverviewAnalyticsService) GetTopPerformingGraph(ctx context.Context, req *types.OverviewRequest) (*types.TopPerformingGraphResponse, error) {
	params, err := buildParams(req)
	if err != nil {
		return nil, err
	}

	result, err := s.repo.GetTopPerformingGraph(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetTopPerformingGraph: query failed")
		return &types.TopPerformingGraphResponse{}, nil
	}

	buckets := make([]string, len(result.Buckets))
	for i, t := range result.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}

	return &types.TopPerformingGraphResponse{
		Buckets:                  buckets,
		FacebookPostCount:        result.FacebookPostCount,
		InstagramPostCount:       result.InstagramPostCount,
		LinkedInPostCount:        result.LinkedInPostCount,
		TiktokPostCount:          result.TiktokPostCount,
		YouTubePostCount:         result.YouTubePostCount,
		PinterestPostCount:       result.PinterestPostCount,
		FacebookEngagementCount:  result.FacebookEngagementCount,
		InstagramEngagementCount: result.InstagramEngagementCount,
		LinkedInEngagementCount:  result.LinkedInEngagementCount,
		TiktokEngagementCount:    result.TiktokEngagementCount,
		YouTubeEngagementCount:   result.YouTubeEngagementCount,
		PinterestEngagementCount: result.PinterestEngagementCount,
		FacebookImpressionCount:  result.FacebookImpressionCount,
		InstagramImpressionCount: result.InstagramImpressionCount,
		LinkedInImpressionCount:  result.LinkedInImpressionCount,
		TiktokImpressionCount:    result.TiktokImpressionCount,
		YouTubeImpressionCount:   result.YouTubeImpressionCount,
		PinterestImpressionCount: result.PinterestImpressionCount,
		FacebookReachCount:       result.FacebookReachCount,
		InstagramReachCount:      result.InstagramReachCount,
		LinkedInReachCount:       result.LinkedInReachCount,
		TiktokReachCount:         result.TiktokReachCount,
		YouTubeReachCount:        result.YouTubeReachCount,
		PinterestReachCount:      result.PinterestReachCount,
	}, nil
}

// GetPlatformDataGrouped returns aggregated metrics per platform (type="grouped").
func (s *OverviewAnalyticsService) GetPlatformDataGrouped(ctx context.Context, req *types.OverviewRequest) ([]*types.PlatformDataRow, error) {
	params, err := buildParams(req)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetPlatformData(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*types.PlatformDataRow, len(rows))
	for i, row := range rows {
		result[i] = &types.PlatformDataRow{
			Followers:    row.Followers,
			TotalPosts:   row.TotalPosts,
			Engagement:   row.Engagement,
			Impressions:  row.Impressions,
			Reach:        row.Reach,
			Reactions:    row.Reactions,
			Comments:     row.Comments,
			Shares:       row.Shares,
			PlatformType: row.PlatformType,
		}
	}
	return result, nil
}

// GetPlatformDataIndividual returns aggregated metrics per account (type="individual" or any non-"grouped").
func (s *OverviewAnalyticsService) GetPlatformDataIndividual(ctx context.Context, req *types.OverviewRequest) ([]*types.AccountDataRow, error) {
	params, err := buildParams(req)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetAccountData(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*types.AccountDataRow, len(rows))
	for i, row := range rows {
		result[i] = &types.AccountDataRow{
			Followers:    row.Followers,
			TotalPosts:   row.TotalPosts,
			Engagement:   row.Engagement,
			Impressions:  row.Impressions,
			Reach:        row.Reach,
			Reactions:    row.Reactions,
			Comments:     row.Comments,
			Shares:       row.Shares,
			PlatformType: row.PlatformType,
			AccountID:    row.AccountID,
		}
	}
	return result, nil
}

// GetPlatformDataDetailed returns current/previous period metrics with pct changes per account.
func (s *OverviewAnalyticsService) GetPlatformDataDetailed(ctx context.Context, req *types.OverviewRequest) ([]*types.AccountDataDetailedRow, error) {
	params, err := buildParams(req)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetAccountDataDetailed(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*types.AccountDataDetailedRow, len(rows))
	for i, row := range rows {
		result[i] = &types.AccountDataDetailedRow{
			PlatformType:         row.PlatformType,
			AccountID:            row.AccountID,
			AccountName:          row.AccountName,
			CurrentFollowers:     row.CurrentFollowers,
			OldFollowers:         row.OldFollowers,
			CurrentPosts:         row.CurrentPosts,
			OldPosts:             row.OldPosts,
			CurrentEngagement:    row.CurrentEngagement,
			OldEngagement:        row.OldEngagement,
			CurrentImpressions:   row.CurrentImpressions,
			OldImpressions:       row.OldImpressions,
			CurrentReach:         row.CurrentReach,
			OldReach:             row.OldReach,
			FollowersChangePct:   row.FollowersChangePct,
			PostsChangePct:       row.PostsChangePct,
			EngagementChangePct:  row.EngagementChangePct,
			ImpressionsChangePct: row.ImpressionsChangePct,
			ReachChangePct:       row.ReachChangePct,
		}
	}
	return result, nil
}

// GetPlatformDataGraphs returns per-account time-series engagement/reach/impressions/posts arrays.
func (s *OverviewAnalyticsService) GetPlatformDataGraphs(ctx context.Context, req *types.OverviewRequest) ([]*types.AccountDataGraphsRow, error) {
	params, err := buildParams(req)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetAccountDataGraphs(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*types.AccountDataGraphsRow, len(rows))
	for i, row := range rows {
		buckets := make([]string, len(row.Buckets))
		for j, t := range row.Buckets {
			buckets[j] = t.Format("2006-01-02")
		}
		result[i] = &types.AccountDataGraphsRow{
			AccountID:   row.AccountID,
			Engagement:  row.Engagement,
			Reach:       row.Reach,
			Impressions: row.Impressions,
			Posts:       row.Posts,
			Buckets:     buckets,
		}
	}
	return result, nil
}

// GetTopPosts returns up to N posts per selected platform, then globally orders the merged set.
func (s *OverviewAnalyticsService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) ([]*types.TopPostRow, error) {
	params, err := repo.NewOverviewParams(
		req.StartDate,
		req.EndDate,
		req.FacebookAccounts,
		req.InstagramAccounts,
		req.LinkedInAccounts,
		req.TiktokAccounts,
		req.PinterestAccounts,
		req.YouTubeAccounts,
		req.Timezone,
		req.Type,
		req.Limit,
	)
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetTopPosts(ctx, params)
	if err != nil {
		return nil, err
	}

	result := make([]*types.TopPostRow, len(rows))
	for i, row := range rows {
		result[i] = &types.TopPostRow{
			PlatformType:    row.PlatformType,
			AccountID:       row.AccountID,
			PostID:          row.PostID,
			Likes:           row.Likes,
			Comments:        row.Comments,
			Shares:          row.Shares,
			Saves:           row.Saves,
			PinClicks:       row.PinClicks,
			OutboundClicks:  row.OutboundClicks,
			DislikesCount:   row.DislikesCount,
			Permalink:       row.Permalink,
			MediaType:       row.MediaType,
			Thumbnail:       row.Thumbnail,
			Category:        row.Category,
			CreatedTime:     formatTimeInTimezone(row.CreatedTime, params.Timezone),
			TotalEngagement: row.TotalEngagement,
			Views:           row.Views,
			Reach:           row.Reach,
		}
	}
	return result, nil
}

func formatTimeInTimezone(value time.Time, timezone string) string {
	if value.IsZero() || value.Year() == 1970 {
		return ""
	}
	if timezone == "" || timezone == "UTC" {
		return value.UTC().Format(time.RFC3339)
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return value.UTC().Format(time.RFC3339)
	}
	return value.In(loc).Format(time.RFC3339)
}
