// Package linkedin provides the business logic layer for LinkedIn analytics.
// It orchestrates repository queries, handles concurrent data fetching via errgroup,
// computes previous-period comparisons, and maps ClickHouse results to API response types.
//
// Migrated from PHP: LinkedInAnalyticsController (contentstudio-backend).
// The PHP controller combined HTTP handling and business logic; here they are separated
// into handler (api/analytics/linkedin) and service (this package).
package linkedin

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/linkedin"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/linkedin"
)

// Service defines the interface for LinkedIn analytics business logic.
// Implement this interface to create mocks for handler testing.
type Service interface {
	GetSummary(ctx context.Context, req *types.LinkedInRequest) (*types.SummaryResponse, error)
	GetAudienceGrowth(ctx context.Context, req *types.LinkedInRequest) (*types.AudienceGrowthResponse, error)
	GetPageViews(ctx context.Context, req *types.LinkedInRequest) (*types.PageViewsResponse, error)
	GetPublishingBehaviour(ctx context.Context, req *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error)
	GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error)
	GetPostsPerDay(ctx context.Context, req *types.LinkedInRequest) (*types.PostsPerDayResponse, error)
	GetHashtags(ctx context.Context, req *types.LinkedInRequest) (*types.HashtagsResponse, error)
	GetFollowersDemographics(ctx context.Context, req *types.LinkedInRequest) (*types.DemographicsResponse, error)
}

// LinkedInAnalyticsService implements LinkedIn analytics business logic.
// Each public method fetches current and previous period data concurrently
// using errgroup, then maps results to response types.
type LinkedInAnalyticsService struct {
	repo   *repo.Repository
	logger zerolog.Logger
}

var _ Service = (*LinkedInAnalyticsService)(nil)

// NewLinkedInAnalyticsService creates a new service with the given repository and logger.
func NewLinkedInAnalyticsService(r *repo.Repository, logger zerolog.Logger) *LinkedInAnalyticsService {
	return &LinkedInAnalyticsService{
		repo:   r,
		logger: logger.With().Str("service", "linkedin-analytics").Logger(),
	}
}

// GetSummary fetches summary metrics for both current and previous periods concurrently.
// Splits the linkedin_posts and linkedin_insights queries into 4 independent goroutines
// to eliminate the cross-table JOIN and maximise parallel DB execution.
func (s *LinkedInAnalyticsService) GetSummary(ctx context.Context, req *types.LinkedInRequest) (*types.SummaryResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)

	var postsCurr, postsPrev *repo.PostsSummaryResult
	var igCurr, igPrev *repo.InsightsSummaryResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetPostsSummary(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get current posts")
			r = &repo.PostsSummaryResult{}
		}
		postsCurr = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPostsSummary(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get previous posts")
			r = &repo.PostsSummaryResult{}
		}
		postsPrev = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetInsightsSummary(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get current insights")
			r = &repo.InsightsSummaryResult{}
		}
		igCurr = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetInsightsSummary(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get previous insights")
			r = &repo.InsightsSummaryResult{}
		}
		igPrev = r
		return nil
	})
	_ = eg.Wait()

	return &types.SummaryResponse{
		Status: true,
		Overview: map[string]*types.SummaryMetrics{
			"current":  mapSummary(postsCurr, igCurr),
			"previous": mapSummary(postsPrev, igPrev),
		},
	}, nil
}

// GetAudienceGrowth fetches time-series follower data, rollup stats, and handles the fallback
// case where the first follower count is 0 by looking back up to 2 years for historical data.
func (s *LinkedInAnalyticsService) GetAudienceGrowth(ctx context.Context, req *types.LinkedInRequest) (*types.AudienceGrowthResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)

	var growth *repo.AudienceResult
	var currentRollup, previousRollup *repo.AudienceRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetAudienceGrowth(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetAudienceGrowth: failed to get growth data")
			r = &repo.AudienceResult{}
		}
		growth = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetAudienceRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetAudienceGrowth: failed to get current rollup")
			r = &repo.AudienceRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetAudienceRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetAudienceGrowth: failed to get previous rollup")
			r = &repo.AudienceRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	// Fallback: if first total_follower_count is 0, query 2 years back for historical data
	if len(growth.TotalFollowerCount) > 0 && growth.TotalFollowerCount[0] == 0 {
		fallbackParams := &clickhouse.QueryParams{
			AccountIDs: params.AccountIDs,
			DateFrom:   params.DateFrom.AddDate(-2, 0, 0),
			DateTo:     params.DateFrom,
			Timezone:   params.Timezone,
		}
		lastCounts, err := s.repo.GetLastFollowerCounts(ctx, fallbackParams)
		if err == nil && lastCounts != nil {
			for i := range growth.TotalFollowerCount {
				if growth.TotalFollowerCount[i] != 0 {
					break
				}
				growth.TotalFollowerCount[i] = lastCounts.TotalFollowerCount
			}
			for i := range growth.OrganicFollowerCount {
				if growth.OrganicFollowerCount[i] != 0 {
					break
				}
				growth.OrganicFollowerCount[i] = lastCounts.OrganicFollowerCount
			}
			for i := range growth.PaidFollowerCount {
				if growth.PaidFollowerCount[i] != 0 {
					break
				}
				growth.PaidFollowerCount[i] = lastCounts.PaidFollowerCount
			}
		}
	}

	return &types.AudienceGrowthResponse{
		Status:         true,
		AudienceGrowth: mapAudienceGrowth(growth),
		AudienceGrowthRollup: map[string]*types.AudienceGrowthRollup{
			"current":  mapAudienceRollup(currentRollup),
			"previous": mapAudienceRollup(previousRollup),
		},
	}, nil
}

// GetPageViews fetches time-series page views and rollup stats for current and previous periods.
func (s *LinkedInAnalyticsService) GetPageViews(ctx context.Context, req *types.LinkedInRequest) (*types.PageViewsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)

	var views *repo.PageViewsResult
	var currentRollup, previousRollup *repo.PageViewsRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetPageViews(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPageViews: failed to get page views data")
			r = &repo.PageViewsResult{}
		}
		views = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPageViewsRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPageViews: failed to get current rollup")
			r = &repo.PageViewsRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPageViewsRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPageViews: failed to get previous rollup")
			r = &repo.PageViewsRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.PageViewsResponse{
		Status:    true,
		PageViews: mapPageViews(views),
		PageViewsRollup: map[string]*types.PageViewsRollup{
			"current":  mapPageViewsRollup(currentRollup),
			"previous": mapPageViewsRollup(previousRollup),
		},
	}, nil
}

// GetPublishingBehaviour fetches publishing metrics filtered by media types, with per-type rollups.
func (s *LinkedInAnalyticsService) GetPublishingBehaviour(ctx context.Context, req *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)
	mediaTypes := req.GetMediaTypes()

	var behaviour *repo.PublishingResult
	var currentRollup, previousRollup []repo.PublishingRollupRow
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetPublishingBehaviour(egCtx, params, mediaTypes)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPublishingBehaviour: failed to get time-series data")
			r = &repo.PublishingResult{}
		}
		behaviour = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPublishingBehaviourRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPublishingBehaviour: failed to get current rollup")
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPublishingBehaviourRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPublishingBehaviour: failed to get previous rollup")
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.PublishingBehaviourResponse{
		Status:              true,
		PublishingBehaviour: mapPublishingBehaviour(behaviour),
		PublishingBehaviourRollup: map[string][]types.PublishingBehaviourMediaType{
			"current":  mapPublishingRollup(currentRollup),
			"previous": mapPublishingRollup(previousRollup),
		},
	}, nil
}

// GetTopPosts fetches the top N posts sorted by the requested metric with optional hashtag filtering.
func (s *LinkedInAnalyticsService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	posts, err := s.repo.GetTopPosts(ctx, params, req.GetLimit(), req.GetOrderBy(), req.Hashtags)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetTopPosts: failed to get top posts")
		return &types.TopPostsResponse{Status: true, TopPosts: []types.TopPost{}}, nil
	}

	return &types.TopPostsResponse{
		Status:   true,
		TopPosts: mapTopPosts(posts, req.GetTimezone()),
	}, nil
}

// GetPostsPerDay fetches the distribution of posts across days of the week.
func (s *LinkedInAnalyticsService) GetPostsPerDay(ctx context.Context, req *types.LinkedInRequest) (*types.PostsPerDayResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	result, err := s.repo.GetPostsPerDay(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetPostsPerDay: failed to get posts per day")
		result = &repo.PostsPerDayResult{}
	}

	days := map[string]int32{
		"Monday":    result.Monday,
		"Tuesday":   result.Tuesday,
		"Wednesday": result.Wednesday,
		"Thursday":  result.Thursday,
		"Friday":    result.Friday,
		"Saturday":  result.Saturday,
		"Sunday":    result.Sunday,
	}
	showData := result.Monday + result.Tuesday + result.Wednesday +
		result.Thursday + result.Friday + result.Saturday + result.Sunday

	return &types.PostsPerDayResponse{
		Status: true,
		PostsPerDays: &types.PostsPerDayData{
			Data: types.PostsPerDayInner{
				Days:     days,
				ShowData: showData,
			},
		},
	}, nil
}

// GetHashtags fetches the top hashtags with engagement metrics and rollup comparisons.
func (s *LinkedInAnalyticsService) GetHashtags(ctx context.Context, req *types.LinkedInRequest) (*types.HashtagsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)

	var hashtags *repo.HashtagsResult
	var currentRollup, previousRollup *repo.HashtagsRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetTopHashtags(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetHashtags: failed to get top hashtags")
			r = &repo.HashtagsResult{}
		}
		hashtags = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetTopHashtagsRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetHashtags: failed to get current rollup")
			r = &repo.HashtagsRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetTopHashtagsRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetHashtags: failed to get previous rollup")
			r = &repo.HashtagsRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.HashtagsResponse{
		Status:      true,
		TopHashtags: mapHashtags(hashtags),
		TopHashtagsRollup: map[string]*types.HashtagsRollup{
			"current":  mapHashtagsRollup(currentRollup),
			"previous": mapHashtagsRollup(previousRollup),
		},
	}, nil
}

// GetFollowersDemographics fetches the latest follower breakdown by seniority, industry, country, and city.
// Parses JSON strings from ClickHouse and adds an "Others" category if counts don't sum to total.
func (s *LinkedInAnalyticsService) GetFollowersDemographics(ctx context.Context, req *types.LinkedInRequest) (*types.DemographicsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	result, err := s.repo.GetFollowersDemographics(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetFollowersDemographics: failed to get demographics")
		return &types.DemographicsResponse{
			Status:               true,
			FollowerDemographics: map[string]*types.DemographicCategory{},
		}, nil
	}

	demographics := map[string]*types.DemographicCategory{
		"seniority": parseDemographicJSON(result.FollowersBySeniority, result.TotalFollowerCount),
		"industry":  parseDemographicJSON(result.FollowersByIndustry, result.TotalFollowerCount),
		"country":   parseDemographicJSON(result.FollowersByCountry, result.TotalFollowerCount),
		"city":      parseDemographicJSON(result.FollowersByCity, result.TotalFollowerCount),
	}

	return &types.DemographicsResponse{
		Status:               true,
		FollowerDemographics: demographics,
	}, nil
}

// --- Private helpers ---

// prevPeriodParams creates query parameters for the previous comparison period.
// The previous period dates are pre-calculated in QueryParams during date parsing.
func (s *LinkedInAnalyticsService) prevPeriodParams(params *clickhouse.QueryParams) *clickhouse.QueryParams {
	return &clickhouse.QueryParams{
		AccountIDs: params.AccountIDs,
		DateFrom:   params.PrevDateFrom,
		DateTo:     params.PrevDateTo,
		Timezone:   params.Timezone,
		DayCount:   params.DayCount,
	}
}

// mapSummary converts a *repo.PostsSummaryResult and *repo.InsightsSummaryResult to *types.SummaryMetrics for the API layer.
func mapSummary(p *repo.PostsSummaryResult, i *repo.InsightsSummaryResult) *types.SummaryMetrics {
	return &types.SummaryMetrics{
		PostComments:       p.PostComments,
		PostLikes:          p.PostLikes,
		TotalEngagement:    p.TotalEngagement,
		TotalPosts:         p.TotalPosts,
		PostShares:         p.PostShares,
		PostClicks:         p.PostClicks,
		Followers:          i.Followers,
		PageViews:          i.PageViews,
		PageReach:          i.PageReach,
		PageShares:         i.PageShares,
		PageComments:       i.PageComments,
		PageReactions:      i.PageReactions,
		PageImpressions:    i.PageImpressions,
		PageUniqueVisitors: i.PageUniqueVisitors,
		EngagementRate:     i.EngagementRate,
		PostEngagementRate: p.PostEngagementRate,
	}
}

// mapAudienceGrowth converts a *repo.AudienceResult to *types.AudienceGrowthData for the API layer.
func mapAudienceGrowth(r *repo.AudienceResult) *types.AudienceGrowthData {
	buckets := make([]string, len(r.Buckets))
	for i, t := range r.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}
	return &types.AudienceGrowthData{
		ShowData:              int32(r.ShowData),
		OrganicFollowerCount:  emptyInt32Slice(r.OrganicFollowerCount),
		OrganicFollowersDaily: emptyInt32Slice(r.OrganicFollowersDaily),
		PaidFollowerCount:     emptyInt32Slice(r.PaidFollowerCount),
		PaidFollowersDaily:    emptyInt32Slice(r.PaidFollowersDaily),
		TotalFollowerCount:    emptyInt32Slice(r.TotalFollowerCount),
		TotalFollowersDaily:   emptyInt32Slice(r.TotalFollowersDaily),
		Buckets:               buckets,
	}
}

// mapAudienceRollup converts a *repo.AudienceRollupResult to *types.AudienceGrowthRollup for the API layer.
func mapAudienceRollup(r *repo.AudienceRollupResult) *types.AudienceGrowthRollup {
	return &types.AudienceGrowthRollup{
		OrganicFollowerCount: r.OrganicFollowerCount,
		PaidFollowerCount:    r.PaidFollowerCount,
		TotalFollowerCount:   r.TotalFollowerCount,
		AvgFollowerCount:     r.AvgFollowerCount,
	}
}

// mapPageViews converts a *repo.PageViewsResult to *types.PageViewsData for the API layer.
func mapPageViews(r *repo.PageViewsResult) *types.PageViewsData {
	buckets := make([]string, len(r.Buckets))
	for i, t := range r.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}
	return &types.PageViewsData{
		DesktopPageViews:      emptyInt32Slice(r.DesktopPageViews),
		MobilePageViews:       emptyInt32Slice(r.MobilePageViews),
		TotalPageViews:        emptyInt32Slice(r.TotalPageViews),
		DesktopPageViewsDaily: emptyInt32Slice(r.DesktopPageViewsDaily),
		MobilePageViewsDaily:  emptyInt32Slice(r.MobilePageViewsDaily),
		TotalPageViewsDaily:   emptyInt32Slice(r.TotalPageViewsDaily),
		ShowData:              r.ShowData,
		Buckets:               buckets,
	}
}

// mapPageViewsRollup converts a *repo.PageViewsRollupResult to *types.PageViewsRollup for the API layer.
func mapPageViewsRollup(r *repo.PageViewsRollupResult) *types.PageViewsRollup {
	return &types.PageViewsRollup{
		TotalPageViews:   r.TotalPageViews,
		DesktopPageViews: r.DesktopPageViews,
		MobilePageViews:  r.MobilePageViews,
		AvgPageViews:     r.AvgPageViews,
	}
}

// mapPublishingBehaviour converts a *repo.PublishingResult to *types.PublishingBehaviourData for the API layer.
func mapPublishingBehaviour(r *repo.PublishingResult) *types.PublishingBehaviourData {
	buckets := make([]string, len(r.Buckets))
	for i, t := range r.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}
	return &types.PublishingBehaviourData{
		Likes:          emptyInt32Slice(r.Likes),
		Comments:       emptyInt32Slice(r.Comments),
		Shares:         emptyInt32Slice(r.Shares),
		Clicks:         emptyInt32Slice(r.Clicks),
		EngagementRate: emptyFloat32Slice(r.EngagementRate),
		Impressions:    emptyInt32Slice(r.Impressions),
		TotalPosts:     emptyInt32Slice(r.TotalPosts),
		Engagement:     emptyInt32Slice(r.Engagement),
		Reach:          emptyInt32Slice(r.Reach),
		Buckets:        buckets,
	}
}

// mapPublishingRollup converts a []repo.PublishingRollupRow to []types.PublishingBehaviourMediaType for the API layer.
func mapPublishingRollup(rows []repo.PublishingRollupRow) []types.PublishingBehaviourMediaType {
	if rows == nil {
		return []types.PublishingBehaviourMediaType{}
	}
	result := make([]types.PublishingBehaviourMediaType, len(rows))
	for i, r := range rows {
		result[i] = types.PublishingBehaviourMediaType{
			MediaType:   r.MediaType,
			TotalPosts:  r.TotalPosts,
			Likes:       r.Likes,
			Comments:    r.Comments,
			Shares:      r.Shares,
			Clicks:      r.Clicks,
			Engagements: r.Engagements,
			Impressions: r.Impressions,
			Reach:       r.Reach,
		}
	}
	return result
}

// mapTopPosts converts a []repo.TopPostResult to []types.TopPost for the API layer.
func mapTopPosts(rows []repo.TopPostResult, timezone string) []types.TopPost {
	if rows == nil {
		return []types.TopPost{}
	}
	result := make([]types.TopPost, len(rows))
	for i, r := range rows {
		result[i] = types.TopPost{
			LinkedinID:      r.LinkedinID,
			PostID:          r.PostID,
			Activity:        r.Activity,
			MediaType:       r.MediaType,
			ArticleURL:      r.ArticleURL,
			ArticleTitle:    r.ArticleTitle,
			PostData:        r.PostData,
			Image:           r.Image,
			Media:           r.Media,
			Type:            r.Type,
			Hashtags:        r.Hashtags,
			Comments:        r.Comments,
			TotalEngagement: r.TotalEngagement,
			Favorites:       r.Favorites,
			Title:           r.Title,
			DayOfWeek:       r.DayOfWeek,
			HourOfDay:       r.HourOfDay,
			CreatedAt:       formatTimeInTimezone(r.CreatedAt, timezone),
			SavingTime:      formatTimeInTimezone(r.SavingTime, timezone),
			PollData:        r.PollData,
			Reach:           r.Reach,
			Repost:          r.Repost,
			PostClicks:      r.PostClicks,
			Impressions:     r.Impressions,
			PublishedAt:     formatTimeInTimezone(r.PublishedAt, timezone),
		}
	}
	return result
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

// mapHashtags converts a *repo.HashtagsResult to *types.HashtagsData for the API layer.
func mapHashtags(r *repo.HashtagsResult) *types.HashtagsData {
	return &types.HashtagsData{
		Name:        emptyStringSlice(r.Name),
		Engagements: emptyInt32Slice(r.Engagements),
		Likes:       emptyInt32Slice(r.Likes),
		Comments:    emptyInt32Slice(r.Comments),
		Shares:      emptyInt32Slice(r.Shares),
		Posts:       emptyInt32Slice(r.Posts),
	}
}

// mapHashtagsRollup converts a *repo.HashtagsRollupResult to *types.HashtagsRollup for the API layer.
func mapHashtagsRollup(r *repo.HashtagsRollupResult) *types.HashtagsRollup {
	return &types.HashtagsRollup{
		TotalHashtags:    r.TotalHashtags,
		TotalTimesUsed:   r.TotalTimesUsed,
		TotalLikes:       r.TotalLikes,
		TotalComments:    r.TotalComments,
		TotalShares:      r.TotalShares,
		TotalEngagement:  r.TotalEngagement,
		TotalImpressions: r.TotalImpressions,
		TotalReach:       r.TotalReach,
	}
}

// parseDemographicJSON parses a JSON string of key-value pairs into buckets/values arrays.
// Adds an "Others" category if the sum of values is less than totalFollowers.
func parseDemographicJSON(jsonStr string, totalFollowers int64) *types.DemographicCategory {
	if jsonStr == "" || jsonStr == "{}" {
		return &types.DemographicCategory{
			Buckets: []string{},
			Values:  []int32{},
		}
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(jsonStr), &data); err != nil {
		return &types.DemographicCategory{
			Buckets: []string{},
			Values:  []int32{},
		}
	}

	type kv struct {
		key   string
		value int32
	}
	var pairs []kv
	var sum int64

	for k, v := range data {
		var val int32
		switch tv := v.(type) {
		case float64:
			val = int32(tv)
		case int:
			val = int32(tv)
		}
		pairs = append(pairs, kv{k, val})
		sum += int64(val)
	}

	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].value > pairs[j].value
	})

	buckets := make([]string, len(pairs))
	values := make([]int32, len(pairs))
	for i, p := range pairs {
		buckets[i] = p.key
		values[i] = p.value
	}

	if totalFollowers > 0 && sum < totalFollowers {
		others := int32(totalFollowers - sum)
		buckets = append(buckets, "Others")
		values = append(values, others)
	}

	return &types.DemographicCategory{
		Buckets: buckets,
		Values:  values,
	}
}

// emptyInt32Slice returns s, or an empty slice if s is nil, ensuring JSON encodes as [] not null.
func emptyInt32Slice(s []int32) []int32 {
	if s == nil {
		return []int32{}
	}
	return s
}

// emptyFloat32Slice returns s, or an empty slice if s is nil, ensuring JSON encodes as [] not null.
func emptyFloat32Slice(s []float32) []float32 {
	if s == nil {
		return []float32{}
	}
	return s
}

// emptyStringSlice returns s, or an empty slice if s is nil, ensuring JSON encodes as [] not null.
func emptyStringSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

// Ensure unused imports are satisfied
var _ = fmt.Sprintf
