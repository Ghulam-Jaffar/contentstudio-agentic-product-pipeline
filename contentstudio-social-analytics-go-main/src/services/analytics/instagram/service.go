// Package instagram provides the business logic layer for Instagram analytics.
// It orchestrates repository queries, handles concurrent data fetching via errgroup,
// computes previous-period comparisons, and maps ClickHouse results to API response types.
//
// Migrated from PHP: InstagramAnalyticsController (contentstudio-backend).
package instagram

import (
	"context"
	"encoding/json"
	"math"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/instagram"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/instagram"
)

// Service defines the interface for Instagram analytics business logic.
type Service interface {
	GetSummary(ctx context.Context, req *types.InstagramRequest) (*types.SummaryResponse, error)
	GetAudienceGrowth(ctx context.Context, req *types.InstagramRequest) (*types.AudienceGrowthResponse, error)
	GetPublishingBehaviour(ctx context.Context, req *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error)
	GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error)
	GetActiveUsers(ctx context.Context, req *types.InstagramRequest) (*types.ActiveUsersResponse, error)
	GetImpressions(ctx context.Context, req *types.InstagramRequest) (*types.ImpressionsResponse, error)
	GetEngagement(ctx context.Context, req *types.InstagramRequest) (*types.EngagementResponse, error)
	GetHashtags(ctx context.Context, req *types.InstagramRequest) (*types.HashtagsResponse, error)
	GetStoriesPerformance(ctx context.Context, req *types.InstagramRequest) (*types.StoriesPerformanceResponse, error)
	GetReelsPerformance(ctx context.Context, req *types.InstagramRequest) (*types.ReelsPerformanceResponse, error)
	GetDemographicsAge(ctx context.Context, req *types.InstagramRequest) (*types.DemographicsAgeResponse, error)
	GetCountryCity(ctx context.Context, req *types.InstagramRequest) (*types.CountryCityResponse, error)
}

// InstagramAnalyticsService implements Instagram analytics business logic.
type InstagramAnalyticsService struct {
	repo   *repo.Repository
	logger zerolog.Logger
}

var _ Service = (*InstagramAnalyticsService)(nil)

// NewInstagramAnalyticsService creates a new service with the given repository and logger.
func NewInstagramAnalyticsService(r *repo.Repository, logger zerolog.Logger) *InstagramAnalyticsService {
	return &InstagramAnalyticsService{
		repo:   r,
		logger: logger.With().Str("service", "instagram-analytics").Logger(),
	}
}

// GetSummary fetches current and previous period post and insights summaries concurrently
// and returns them under "current" and "previous" keys for period-over-period comparison.
func (s *InstagramAnalyticsService) GetSummary(ctx context.Context, req *types.InstagramRequest) (*types.SummaryResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)

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

// GetAudienceGrowth fetches time-series follower data and current/previous rollups concurrently.
// When the first follower count is 0, it looks back up to 2 years for the last known count as a fallback.
func (s *InstagramAnalyticsService) GetAudienceGrowth(ctx context.Context, req *types.InstagramRequest) (*types.AudienceGrowthResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)

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

	// Fallback: if first follower count is 0, look back up to 2 years.
	if len(growth.Followers) > 0 && growth.Followers[0] == 0 {
		fallbackParams := &clickhouse.QueryParams{
			AccountIDs: params.AccountIDs,
			DateFrom:   params.DateFrom.AddDate(-2, 0, 0),
			DateTo:     params.DateFrom,
			Timezone:   params.Timezone,
		}
		last, err := s.repo.GetLastFollowerCount(ctx, fallbackParams)
		if err == nil && last != nil && last.FollowersCount > 0 {
			for i := range growth.Followers {
				if growth.Followers[i] != 0 {
					break
				}
				growth.Followers[i] = last.FollowersCount
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

// GetPublishingBehaviour fetches time-series publishing data and current/previous rollups concurrently,
// filtered by the media types supplied in the request (IMAGE, VIDEO, CAROUSEL_ALBUM, REELS).
func (s *InstagramAnalyticsService) GetPublishingBehaviour(ctx context.Context, req *types.PublishingBehaviourRequest) (*types.PublishingBehaviourResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)
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
		r, err := s.repo.GetPublishingBehaviourRollup(egCtx, params, mediaTypes)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPublishingBehaviour: failed to get current rollup")
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPublishingBehaviourRollup(egCtx, prevParams, mediaTypes)
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

// GetTopPosts fetches the highest-performing posts sorted by the requested metric and limited
// to the requested count; returns an empty slice on error rather than propagating it.
func (s *InstagramAnalyticsService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	posts, err := s.repo.GetTopPosts(ctx, params, req.GetOrderBy(), req.GetLimit(), req.Hashtags)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetTopPosts: failed to get top posts")
		return &types.TopPostsResponse{Status: true, TopPosts: []types.TopPost{}}, nil
	}

	return &types.TopPostsResponse{
		Status:   true,
		TopPosts: mapTopPosts(posts, req.GetTimezone()),
	}, nil
}

// GetActiveUsers fetches hourly and day-of-week audience activity data concurrently
// from instagram_insights and returns both in a single response.
func (s *InstagramAnalyticsService) GetActiveUsers(ctx context.Context, req *types.InstagramRequest) (*types.ActiveUsersResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	var hours *repo.ActiveUsersHoursResult
	var days *repo.ActiveUsersDaysResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetActiveUsersHours(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetActiveUsers: failed to get hourly data")
			r = &repo.ActiveUsersHoursResult{}
		}
		hours = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetActiveUsersDays(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetActiveUsers: failed to get daily data")
			r = &repo.ActiveUsersDaysResult{}
		}
		days = r
		return nil
	})
	_ = eg.Wait()

	return &types.ActiveUsersResponse{
		Status:           true,
		ActiveUsersHours: mapActiveUsersHours(hours),
		ActiveUsersDays:  mapActiveUsersDays(days),
	}, nil
}

// GetImpressions fetches time-series impression data from instagram_insights and current/previous
// period rollups concurrently, returning all three in a single response.
func (s *InstagramAnalyticsService) GetImpressions(ctx context.Context, req *types.InstagramRequest) (*types.ImpressionsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)

	var impressions *repo.ImpressionsResult
	var currentRollup, previousRollup *repo.ImpressionsRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetImpressions(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetImpressions: failed to get time-series data")
			r = &repo.ImpressionsResult{}
		}
		impressions = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetImpressionsRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetImpressions: failed to get current rollup")
			r = &repo.ImpressionsRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetImpressionsRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetImpressions: failed to get previous rollup")
			r = &repo.ImpressionsRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.ImpressionsResponse{
		Status:      true,
		Impressions: mapImpressions(impressions),
		ImpressionsRollup: map[string]*types.ImpressionsRollup{
			"current":  mapImpressionsRollup(currentRollup),
			"previous": mapImpressionsRollup(previousRollup),
		},
	}, nil
}

// GetEngagement fetches time-series engagement data from instagram_insights and current/previous
// period rollups concurrently, returning all three in a single response.
func (s *InstagramAnalyticsService) GetEngagement(ctx context.Context, req *types.InstagramRequest) (*types.EngagementResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)

	var engagement *repo.EngagementResult
	var currentRollup, previousRollup *repo.EngagementRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetEngagement(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetEngagement: failed to get time-series data")
			r = &repo.EngagementResult{}
		}
		engagement = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetEngagementRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetEngagement: failed to get current rollup")
			r = &repo.EngagementRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetEngagementRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetEngagement: failed to get previous rollup")
			r = &repo.EngagementRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.EngagementResponse{
		Status:      true,
		Engagements: mapEngagement(engagement),
		EngagementsRollup: map[string]*types.EngagementRollup{
			"current":  mapEngagementRollup(currentRollup),
			"previous": mapEngagementRollup(previousRollup),
		},
	}, nil
}

// GetHashtags fetches the top hashtags list and current/previous period rollups concurrently,
// returning aggregate engagement stats per hashtag with period-over-period comparison.
func (s *InstagramAnalyticsService) GetHashtags(ctx context.Context, req *types.InstagramRequest) (*types.HashtagsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)

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

// GetStoriesPerformance fetches time-series stories data and the previous-period rollup concurrently.
// The current-period rollup is derived from the time-series result via storiesRollupFromTimeSeries.
func (s *InstagramAnalyticsService) GetStoriesPerformance(ctx context.Context, req *types.InstagramRequest) (*types.StoriesPerformanceResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)

	var stories *repo.StoriesResult
	var previousRollup *repo.StoriesRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetStoriesPerformance(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetStoriesPerformance: failed to get time-series data")
			r = &repo.StoriesResult{}
		}
		stories = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetStoriesRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetStoriesPerformance: failed to get previous rollup")
			r = &repo.StoriesRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.StoriesPerformanceResponse{
		Status:             true,
		StoriesPerformance: mapStoriesPerformance(stories),
		StoriesRollup: map[string]*types.StoriesRollup{
			"current":  mapStoriesRollup(storiesRollupFromTimeSeries(stories)),
			"previous": mapStoriesRollup(previousRollup),
		},
	}, nil
}

// GetReelsPerformance fetches time-series reels data and current/previous period rollups concurrently,
// returning engagement, watch-time, and post-count metrics for the period.
func (s *InstagramAnalyticsService) GetReelsPerformance(ctx context.Context, req *types.InstagramRequest) (*types.ReelsPerformanceResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)

	var reels *repo.ReelsResult
	var currentRollup, previousRollup *repo.ReelsRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetReelsPerformance(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetReelsPerformance: failed to get time-series data")
			r = &repo.ReelsResult{}
		}
		reels = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetReelsRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetReelsPerformance: failed to get current rollup")
			r = &repo.ReelsRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetReelsRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetReelsPerformance: failed to get previous rollup")
			r = &repo.ReelsRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.ReelsPerformanceResponse{
		Status: true,
		Reels:  mapReelsPerformance(reels),
		ReelsRollup: map[string]*types.ReelsRollup{
			"current":  mapReelsRollup(currentRollup),
			"previous": mapReelsRollup(previousRollup),
		},
	}, nil
}

// GetDemographicsAge fetches audience demographics from instagram_insights and returns
// age, gender, and gender+age breakdowns, including the peak gender-age combination.
func (s *InstagramAnalyticsService) GetDemographicsAge(ctx context.Context, req *types.InstagramRequest) (*types.DemographicsAgeResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	result, err := s.repo.GetDemographics(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetDemographicsAge: failed to get demographics")
		return &types.DemographicsAgeResponse{
			AudienceAge:    map[string]int64{},
			AudienceGender: map[string]int64{},
		}, nil
	}

	audienceAge := parseDemographicArray(result.AudienceAge)
	audienceGender := parseDemographicArray(result.AudienceGender)
	audienceGenderAge := parseDemographicArray(result.AudienceGenderAge)

	return &types.DemographicsAgeResponse{
		AudienceAge:    audienceAge,
		AudienceGender: audienceGender,
		MaxAudienceAge: findMaxAudienceAge(audienceGenderAge),
	}, nil
}

// GetCountryCity fetches audience country and city location data from instagram_insights
// and returns both as key-value maps of location label to follower count.
func (s *InstagramAnalyticsService) GetCountryCity(ctx context.Context, req *types.InstagramRequest) (*types.CountryCityResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	result, err := s.repo.GetLocations(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetCountryCity: failed to get locations")
		return &types.CountryCityResponse{
			AudienceCity:    map[string]int64{},
			AudienceCountry: map[string]int64{},
		}, nil
	}

	return &types.CountryCityResponse{
		AudienceCity:    parseDemographicArray(result.AudienceCity),
		AudienceCountry: parseDemographicArray(result.AudienceCountry),
	}, nil
}

// --- Private helpers ---

// prevPeriodParams builds a QueryParams for the previous period by swapping DateFrom/DateTo
// with PrevDateFrom/PrevDateTo so the same repository methods can be called without modification.
func prevPeriodParams(params *clickhouse.QueryParams) *clickhouse.QueryParams {
	return &clickhouse.QueryParams{
		AccountIDs: params.AccountIDs,
		DateFrom:   params.PrevDateFrom,
		DateTo:     params.PrevDateTo,
		Timezone:   params.Timezone,
		DayCount:   params.DayCount,
	}
}

// mapSummary merges post-level and insights-level summary results into a single SummaryMetrics.
func mapSummary(p *repo.PostsSummaryResult, i *repo.InsightsSummaryResult) *types.SummaryMetrics {
	return &types.SummaryMetrics{
		TotalPosts:         p.TotalPosts,
		PostEngagement:     p.TotalEngagement,
		PostReactions:      p.Likes,
		PostComments:       p.Comments,
		PostSaves:          p.Saved,
		PostReach:          p.Reach,
		ProfileImpressions: i.Impressions,
		PostViews:          p.Views,
		TotalStories:       p.Stories,
		ProfileViews:       i.ProfileViews,
		FollowersCount:     i.FollowersCount,
		FollowsCount:       i.FollowsCount,
		AccountsEngaged:    i.AccountsEngaged,
		ProfileEngagement:  i.Engagement,
		ProfileReach:       i.Reach,
		DocCount:           p.DocCount,
		EngRate:            p.EngagementRate,
	}
}

// mapAudienceGrowth converts the repository AudienceResult to the API type, formatting
// bucket timestamps to ISO-8601 date strings and computing daily follower deltas.
func mapAudienceGrowth(r *repo.AudienceResult) *types.AudienceGrowthData {
	buckets := make([]string, len(r.Buckets))
	for i, t := range r.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}
	followers := emptyInt32Slice(r.Followers)
	return &types.AudienceGrowthData{
		ShowData:       int32(r.ShowData),
		Followers:      followers,
		FollowersDaily: computeDailyDelta(followers),
		Buckets:        buckets,
	}
}

// computeDailyDelta computes the day-over-day difference for a slice of follower counts.
func computeDailyDelta(followers []int32) []int32 {
	if len(followers) == 0 {
		return []int32{}
	}
	deltas := make([]int32, len(followers))
	for i := range followers {
		if i == 0 {
			deltas[i] = 0
		} else {
			deltas[i] = followers[i] - followers[i-1]
		}
	}
	return deltas
}

// mapAudienceRollup converts an AudienceRollupResult to the API rollup type.
func mapAudienceRollup(r *repo.AudienceRollupResult) *types.AudienceGrowthRollup {
	return &types.AudienceGrowthRollup{
		FollowerCount:  r.FollowerCount,
		FollowerGained: r.FollowerGained,
	}
}

// mapPublishingBehaviour converts the repository PublishingResult to the API type,
// formatting bucket timestamps and replacing nil slices with empty slices for consistent JSON output.
func mapPublishingBehaviour(r *repo.PublishingResult) *types.PublishingBehaviourData {
	buckets := make([]string, len(r.Buckets))
	for i, t := range r.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}
	return &types.PublishingBehaviourData{
		Likes:       emptyInt32Slice(r.Likes),
		Comments:    emptyInt32Slice(r.Comments),
		Saved:       emptyInt32Slice(r.Saved),
		Engagement:  emptyInt32Slice(r.Engagement),
		Reach:       emptyInt32Slice(r.Reach),
		Impressions: emptyInt32Slice(r.Impressions),
		Views:       emptyInt32Slice(r.Views),
		TotalPosts:  emptyInt32Slice(r.TotalPosts),
		Buckets:     buckets,
	}
}

// mapPublishingRollup converts a slice of per-media-type rollup rows to the API slice type.
func mapPublishingRollup(rows []repo.PublishingRollupRow) []types.PublishingBehaviourMediaType {
	if rows == nil {
		return []types.PublishingBehaviourMediaType{}
	}
	result := make([]types.PublishingBehaviourMediaType, len(rows))
	for i, r := range rows {
		result[i] = types.PublishingBehaviourMediaType{
			MediaType:  r.MediaType,
			TotalPosts: r.TotalPosts,
			Likes:      r.Likes,
			Comments:   r.Comments,
			Saved:      r.Saved,
			Engagement: r.Engagement,
			Reach:      r.Reach,
			Views:      r.Views,
		}
	}
	return result
}

// mapTopPosts converts repository TopPostResult rows to the API slice type,
// formatting timestamps as RFC3339 strings for consistent JSON serialisation.
func mapTopPosts(rows []repo.TopPostResult, timezone string) []types.TopPost {
	if rows == nil {
		return []types.TopPost{}
	}
	result := make([]types.TopPost, len(rows))
	for i, r := range rows {
		result[i] = types.TopPost{
			InstagramID:         r.InstagramID,
			MediaID:             r.MediaID,
			Caption:             r.Caption,
			MediaType:           r.MediaType,
			EntityType:          r.EntityType,
			MediaURL:            emptyStringSlice(r.MediaURL),
			VideoURL:            emptyStringSlice(r.VideoURL),
			Permalink:           r.Permalink,
			LikeCount:           r.LikeCount,
			CommentsCount:       r.CommentsCount,
			Saved:               r.Saved,
			Engagement:          r.Engagement,
			Reach:               r.Reach,
			Impressions:         r.Impressions,
			Views:               r.Views,
			Shares:              r.Shares,
			ReelsAvgWatchTime:   r.ReelsAvgWatchTime,
			ReelsTotalWatchTime: r.ReelsTotalWatchTime,
			Exits:               r.Exits,
			Replies:             r.Replies,
			Hashtags:            emptyStringSlice(r.Hashtags),
			DayOfWeek:           r.DayOfWeek,
			HourOfDay:           r.HourOfDay,
			PostCreatedAt:       formatTimeInTimezone(r.PostCreatedAt, timezone),
			StoredEventAt:       formatTimeInTimezone(r.StoredEventAt, timezone),
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

// mapActiveUsersHours converts hourly active-users data to the API type.
func mapActiveUsersHours(r *repo.ActiveUsersHoursResult) *types.ActiveUsersHours {
	return &types.ActiveUsersHours{
		Buckets:      emptyInt32Slice(r.Buckets),
		Values:       emptyInt32Slice(r.Values),
		HighestValue: r.HighestValue,
		HighestHour:  r.HighestHour,
	}
}

// mapActiveUsersDays converts day-of-week active-users data to the API type.
func mapActiveUsersDays(r *repo.ActiveUsersDaysResult) *types.ActiveUsersDays {
	return &types.ActiveUsersDays{
		Buckets:      emptyStringSlice(r.Buckets),
		Values:       emptyInt32Slice(r.Values),
		HighestValue: r.HighestValue,
		HighestDay:   r.HighestDay,
	}
}

// mapImpressions converts the ImpressionsResult to the API type, formatting bucket timestamps.
func mapImpressions(r *repo.ImpressionsResult) *types.ImpressionsData {
	buckets := make([]string, len(r.Buckets))
	for i, t := range r.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}
	return &types.ImpressionsData{
		ShowData:    int32(r.ShowData),
		Buckets:     buckets,
		Impressions: emptyInt32Slice(r.Impressions),
	}
}

// mapImpressionsRollup converts an ImpressionsRollupResult to the API rollup type.
func mapImpressionsRollup(r *repo.ImpressionsRollupResult) *types.ImpressionsRollup {
	return &types.ImpressionsRollup{
		TotalImpressions: r.TotalImpressions,
		AvgImpressions:   r.AvgImpressions,
	}
}

// mapEngagement converts the EngagementResult to the API type, formatting bucket timestamps.
func mapEngagement(r *repo.EngagementResult) *types.EngagementData {
	buckets := make([]string, len(r.Buckets))
	for i, t := range r.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}
	return &types.EngagementData{
		ShowData:   int32(r.ShowData),
		Buckets:    buckets,
		Engagement: emptyInt32Slice(r.Engagement),
		Comments:   emptyInt32Slice(r.Comments),
		Reactions:  emptyInt32Slice(r.Reactions),
		DocCount:   emptyInt32Slice(r.DocCount),
	}
}

// mapEngagementRollup converts an EngagementRollupResult to the API rollup type.
func mapEngagementRollup(r *repo.EngagementRollupResult) *types.EngagementRollup {
	return &types.EngagementRollup{
		Engagement:    r.Engagement,
		AvgEngagement: r.AvgEngagement,
		Comments:      r.Comments,
		Reactions:     r.Reactions,
		Saved:         r.Saved,
		Count:         r.Count,
	}
}

// mapHashtags converts a HashtagsResult to the API type with nil-safe slices.
func mapHashtags(r *repo.HashtagsResult) *types.HashtagsData {
	return &types.HashtagsData{
		Name:       emptyStringSlice(r.Name),
		Engagement: emptyInt32Slice(r.Engagement),
		Likes:      emptyInt32Slice(r.Likes),
		Comments:   emptyInt32Slice(r.Comments),
		Saved:      emptyInt32Slice(r.Saved),
		Posts:      emptyInt32Slice(r.Posts),
	}
}

// mapHashtagsRollup converts a HashtagsRollupResult to the API rollup type.
func mapHashtagsRollup(r *repo.HashtagsRollupResult) *types.HashtagsRollup {
	return &types.HashtagsRollup{
		TotalEngagement:     r.TotalEngagement,
		TotalLikes:          r.TotalLikes,
		TotalComments:       r.TotalComments,
		TotalSaves:          r.TotalSaves,
		TotalUniqueHashtags: r.TotalUniqueHashtags,
		TotalHashtagUses:    r.TotalHashtagUses,
	}
}

// mapStoriesPerformance converts the StoriesResult to the API type, formatting bucket timestamps.
func mapStoriesPerformance(r *repo.StoriesResult) *types.StoriesData {
	buckets := make([]string, len(r.Buckets))
	for i, t := range r.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}
	return &types.StoriesData{
		ShowData:            int32(r.ShowData),
		Buckets:             buckets,
		AvgStoryImpressions: emptyFloat64Slice(r.AvgStoryImpressions),
		StoryImpressions:    emptyInt32Slice(r.StoryImpressions),
		StoryReach:          emptyInt32Slice(r.StoryReach),
		StoryReply:          emptyInt32Slice(r.StoryReply),
		StoryExits:          emptyInt32Slice(r.StoryExits),
		StoryTapsForward:    emptyInt32Slice(r.StoryTapsForward),
		StoryTapsBack:       emptyInt32Slice(r.StoryTapsBack),
		PublishedStories:    emptyInt32Slice(r.PublishedStories),
	}
}

// storiesRollupFromTimeSeries derives current-period rollup metrics from the time-series result,
// eliminating a redundant DB round-trip. The rollup SQL computes the same sums/counts as the
// time-series inner query; WITH FILL zeros contribute 0 to all sums so totals are unaffected.
func storiesRollupFromTimeSeries(s *repo.StoriesResult) *repo.StoriesRollupResult {
	var storyImpressions, storyReach, storyReply, storyExits int64
	var storyTapsForward, storyTapsBack, publishedStories int64
	for i := range s.StoryImpressions {
		storyImpressions += int64(s.StoryImpressions[i])
		storyReach += int64(s.StoryReach[i])
		storyReply += int64(s.StoryReply[i])
		storyExits += int64(s.StoryExits[i])
		storyTapsForward += int64(s.StoryTapsForward[i])
		storyTapsBack += int64(s.StoryTapsBack[i])
		publishedStories += int64(s.PublishedStories[i])
	}
	var avgStoryImpressions float64
	if publishedStories > 0 {
		avgStoryImpressions = math.Round(float64(storyImpressions)/float64(publishedStories)*100) / 100
	}
	return &repo.StoriesRollupResult{
		StoryImpressions:    storyImpressions,
		AvgStoryImpressions: avgStoryImpressions,
		StoryReach:          storyReach,
		StoryReply:          storyReply,
		StoryExits:          storyExits,
		StoryTapsForward:    storyTapsForward,
		StoryTapsBack:       storyTapsBack,
		PublishedStories:    publishedStories,
	}
}

// mapStoriesRollup converts a StoriesRollupResult to the API rollup type.
func mapStoriesRollup(r *repo.StoriesRollupResult) *types.StoriesRollup {
	return &types.StoriesRollup{
		StoryImpressions:    r.StoryImpressions,
		AvgStoryImpressions: r.AvgStoryImpressions,
		StoryReach:          r.StoryReach,
		StoryReply:          r.StoryReply,
		StoryExits:          r.StoryExits,
		StoryTapsForward:    r.StoryTapsForward,
		StoryTapsBack:       r.StoryTapsBack,
		PublishedStories:    r.PublishedStories,
	}
}

// mapReelsPerformance converts the ReelsResult to the API type, formatting bucket timestamps.
func mapReelsPerformance(r *repo.ReelsResult) *types.ReelsData {
	buckets := make([]string, len(r.Buckets))
	for i, t := range r.Buckets {
		buckets[i] = t.Format("2006-01-02")
	}
	return &types.ReelsData{
		ShowData:       int32(r.ShowData),
		Buckets:        buckets,
		TotalPosts:     emptyInt32Slice(r.TotalPosts),
		Engagement:     emptyInt32Slice(r.Engagement),
		Likes:          emptyInt32Slice(r.Likes),
		Comments:       emptyInt32Slice(r.Comments),
		Saves:          emptyInt32Slice(r.Saves),
		Shares:         emptyInt32Slice(r.Shares),
		AvgWatchTime:   emptyFloat64Slice(r.AvgWatchTime),
		TotalWatchTime: emptyInt64Slice(r.TotalWatchTime),
	}
}

// mapReelsRollup converts a ReelsRollupResult to the API rollup type.
func mapReelsRollup(r *repo.ReelsRollupResult) *types.ReelsRollup {
	return &types.ReelsRollup{
		Engagement:     r.Engagement,
		Likes:          r.Likes,
		Comments:       r.Comments,
		Saves:          r.Saves,
		TotalPosts:     r.TotalPosts,
		Shares:         r.Shares,
		AvgWatchTime:   r.AvgWatchTime,
		TotalWatchTime: r.TotalWatchTime,
	}
}

// parseDemographicArray parses a ClickHouse Array(String) of demographic entries.
// Each string may be a JSON object {"key":"...", "value": N} or a "key:value" pair.
func parseDemographicArray(entries []string) map[string]int64 {
	result := make(map[string]int64)
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		// Try JSON format: {"key": "age_18_24", "value": 100}
		var obj map[string]interface{}
		if err := json.Unmarshal([]byte(entry), &obj); err == nil {
			key, _ := obj["key"].(string)
			if key == "" {
				// Some Instagram responses use "dimension_value" or "age_range"
				for _, k := range []string{"dimension_value", "age_range", "name"} {
					if v, ok := obj[k].(string); ok {
						key = v
						break
					}
				}
			}
			if key == "" {
				continue
			}
			switch v := obj["value"].(type) {
			case float64:
				result[key] += int64(v)
			}
			continue
		}
		// Try "key:value" format
		if idx := strings.LastIndex(entry, ":"); idx > 0 {
			key := entry[:idx]
			var val int64
			if _, err := parseIntFast(entry[idx+1:], &val); err == nil {
				result[key] += val
			}
		}
	}
	return result
}

// findMaxAudienceAge finds the gender+age combination with the highest count.
func findMaxAudienceAge(genderAge map[string]int64) *types.MaxAudienceAge {
	var maxKey string
	var maxVal int64
	for k, v := range genderAge {
		if v > maxVal {
			maxVal = v
			maxKey = k
		}
	}
	if maxKey == "" {
		return nil
	}
	// Keys are expected as "F.18-24" or "M.25-34"
	parts := strings.SplitN(maxKey, ".", 2)
	gender, age := "", maxKey
	if len(parts) == 2 {
		gender, age = parts[0], parts[1]
	}
	return &types.MaxAudienceAge{
		Gender: gender,
		Age:    age,
		Value:  maxVal,
	}
}

// parseIntFast parses a decimal integer string into *val.
func parseIntFast(s string, val *int64) (string, error) {
	s = strings.TrimSpace(s)
	var n int64
	if len(s) == 0 {
		return s, &parseError{}
	}
	for _, c := range s {
		if c < '0' || c > '9' {
			return s, &parseError{}
		}
		n = n*10 + int64(c-'0')
	}
	*val = n
	return s, nil
}

type parseError struct{}

func (*parseError) Error() string { return "parse error" }

// emptyInt32Slice returns an initialised empty slice instead of nil so JSON output is [] not null.
func emptyInt32Slice(s []int32) []int32 {
	if s == nil {
		return []int32{}
	}
	return s
}

// emptyInt64Slice returns an initialised empty slice instead of nil so JSON output is [] not null.
func emptyInt64Slice(s []int64) []int64 {
	if s == nil {
		return []int64{}
	}
	return s
}

// emptyFloat64Slice returns an initialised empty slice instead of nil so JSON output is [] not null.
func emptyFloat64Slice(s []float64) []float64 {
	if s == nil {
		return []float64{}
	}
	return s
}

// emptyStringSlice returns an initialised empty slice instead of nil so JSON output is [] not null.
func emptyStringSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
