package facebook

import (
	"context"
	"math"
	"sort"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/facebook"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/facebook"
)

// Package facebook provides the business logic layer for Facebook analytics.
// It orchestrates repository queries, handles concurrent data fetching via errgroup,
// computes previous-period comparisons, and maps ClickHouse results to API response types.
//
// Migrated from PHP: FacebookAnalyticsController (contentstudio-backend).

// Repository is the interface for Facebook analytics ClickHouse queries.
// Defined here so the service is testable with mock implementations.
type Repository interface {
	GetSummary(ctx context.Context, params *clickhouse.QueryParams) (*repo.SummaryResult, error)
	GetPostsSummary(ctx context.Context, params *clickhouse.QueryParams) (*repo.PostsSummaryResult, error)
	GetInsightsSummary(ctx context.Context, params *clickhouse.QueryParams) (*repo.InsightsSummaryResult, error)
	GetAudienceGrowth(ctx context.Context, params *clickhouse.QueryParams) (*repo.AudienceGrowthResult, error)
	GetLastFollowerCounts(ctx context.Context, params *clickhouse.QueryParams) (*repo.LastFollowerCounts, error)
	GetAudienceGrowthRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.AudienceGrowthRollupResult, error)
	GetPublishingBehaviour(ctx context.Context, params *clickhouse.QueryParams, mediaTypes []string) (*repo.PublishingBehaviourResult, error)
	GetPublishingBehaviourRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.PublishingRollupResult, error)
	GetTopPosts(ctx context.Context, params *clickhouse.QueryParams, mediaTypes []string, limit int, orderBy string) ([]repo.TopPostRow, error)
	GetActiveUsersHours(ctx context.Context, params *clickhouse.QueryParams) (*repo.ActiveUsersHoursResult, error)
	GetActiveUsersDays(ctx context.Context, params *clickhouse.QueryParams) (*repo.ActiveUsersDaysResult, error)
	GetImpressions(ctx context.Context, params *clickhouse.QueryParams) (*repo.ImpressionsResult, error)
	GetImpressionsRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.ImpressionsRollupResult, error)
	GetEngagement(ctx context.Context, params *clickhouse.QueryParams) (*repo.EngagementResult, error)
	GetEngagementRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.EngagementRollupResult, error)
	GetReelsAnalytics(ctx context.Context, params *clickhouse.QueryParams) (*repo.ReelsAnalyticsResult, error)
	GetReelsRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.ReelsRollupResult, error)
	GetVideoInsights(ctx context.Context, params *clickhouse.QueryParams) (*repo.VideoInsightsResult, error)
	GetVideoRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.VideoRollupResult, error)
	GetAudienceGender(ctx context.Context, params *clickhouse.QueryParams) (*repo.AudienceGenderResult, error)
	GetMaxGenderAge(ctx context.Context, params *clickhouse.QueryParams) (*repo.MaxGenderAgeResult, error)
	GetAudienceAge(ctx context.Context, params *clickhouse.QueryParams) (*repo.AudienceAgeResult, error)
	GetAudienceCountry(ctx context.Context, params *clickhouse.QueryParams) (map[string]int32, error)
	GetAudienceCity(ctx context.Context, params *clickhouse.QueryParams) (map[string]int32, error)
}

// Service defines the interface for Facebook analytics business logic.
// Implement this interface to create mocks for handler testing.
type Service interface {
	GetSummary(ctx context.Context, req *types.FacebookRequest) (*types.SummaryResponse, error)
	GetAudienceGrowth(ctx context.Context, req *types.FacebookRequest) (*types.AudienceGrowthResponse, error)
	GetPublishingBehaviour(ctx context.Context, req *types.FacebookRequest) (*types.PublishingBehaviourResponse, error)
	GetTopPosts(ctx context.Context, req *types.FacebookRequest) (*types.TopPostsResponse, error)
	GetActiveUsers(ctx context.Context, req *types.FacebookRequest) (*types.ActiveUsersResponse, error)
	GetImpressions(ctx context.Context, req *types.FacebookRequest) (*types.ImpressionsResponse, error)
	GetEngagement(ctx context.Context, req *types.FacebookRequest) (*types.EngagementResponse, error)
	GetReelsAnalytics(ctx context.Context, req *types.FacebookRequest) (*types.ReelsAnalyticsResponse, error)
	GetVideoInsights(ctx context.Context, req *types.FacebookRequest) (*types.VideoInsightsResponse, error)
	GetDemographics(ctx context.Context, req *types.FacebookRequest) (*types.DemographicsResponse, error)
	GetOverviewDemographics(ctx context.Context, req *types.FacebookRequest) (*types.DemographicsResponse, error)
	GetAudienceLocation(ctx context.Context, req *types.FacebookRequest) (*types.DemographicsResponse, error)
}

// FacebookAnalyticsService implements Facebook analytics business logic.
// Each public method fetches current and previous period data concurrently
// using errgroup, then maps results to response types.
type FacebookAnalyticsService struct {
	repo   Repository
	logger zerolog.Logger
}

// NewFacebookAnalyticsService creates a new service with the given repository and logger.
func NewFacebookAnalyticsService(r Repository, logger zerolog.Logger) *FacebookAnalyticsService {
	return &FacebookAnalyticsService{
		repo:   r,
		logger: logger.With().Str("service", "facebook-analytics").Logger(),
	}
}

var _ Service = (*FacebookAnalyticsService)(nil)

// GetSummary fetches summary metrics for both current and previous periods concurrently.
// Runs 4 goroutines (posts/insights × current/previous) to maximise parallel DB execution.
func (s *FacebookAnalyticsService) GetSummary(ctx context.Context, req *types.FacebookRequest) (*types.SummaryResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prev := prevPeriodParams(params)

	var postsCurr, postsPrev *repo.PostsSummaryResult
	var igCurr, igPrev *repo.InsightsSummaryResult
	eg, _ := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetPostsSummary(ctx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: current posts query failed")
			return err
		}
		postsCurr = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPostsSummary(ctx, prev)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: previous posts query failed")
			return err
		}
		postsPrev = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetInsightsSummary(ctx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: current insights query failed")
			return err
		}
		igCurr = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetInsightsSummary(ctx, prev)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: previous insights query failed")
			return err
		}
		igPrev = r
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	return &types.SummaryResponse{
		Status: true,
		Overview: map[string]*types.SummaryMetrics{
			"current":  mapSummary(postsCurr, igCurr),
			"previous": mapSummary(postsPrev, igPrev),
		},
	}, nil
}

// GetAudienceGrowth fetches time-series fan data, rollup stats, and handles the fallback
// case where the first fan count is 0 by looking back up to 2 years for historical data.
func (s *FacebookAnalyticsService) GetAudienceGrowth(ctx context.Context, req *types.FacebookRequest) (*types.AudienceGrowthResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prev := prevPeriodParams(params)

	var growth *repo.AudienceGrowthResult
	var currentRollup, previousRollup *repo.AudienceGrowthRollupResult
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetAudienceGrowth(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetAudienceGrowth: growth query failed")
			r = &repo.AudienceGrowthResult{}
		}
		growth = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetAudienceGrowthRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetAudienceGrowth: current rollup query failed")
			r = &repo.AudienceGrowthRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetAudienceGrowthRollup(egCtx, prev)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetAudienceGrowth: previous rollup query failed")
			r = &repo.AudienceGrowthRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	if len(growth.FanCount) > 0 && growth.FanCount[0] == 0 {
		fallbackParams := &clickhouse.QueryParams{
			AccountIDs: params.AccountIDs,
			DateFrom:   params.DateFrom.AddDate(-2, 0, 0),
			DateTo:     params.DateFrom,
			Timezone:   params.Timezone,
			DayCount:   params.DayCount,
		}
		if last, err := s.repo.GetLastFollowerCounts(ctx, fallbackParams); err == nil && last != nil {
			for i := range growth.FanCount {
				if growth.FanCount[i] != 0 {
					break
				}
				growth.FanCount[i] = last.PageFans
				if i < len(growth.PageFansByLike) && growth.PageFansByLike[i] == 0 {
					growth.PageFansByLike[i] = last.PageFansByLike
				}
				if i < len(growth.PageFansByUnlike) && growth.PageFansByUnlike[i] == 0 {
					growth.PageFansByUnlike[i] = last.PageFansByUnlike
				}
			}
		}
	}

	return &types.AudienceGrowthResponse{
		Status:         true,
		AudienceGrowth: mapAudienceGrowth(growth),
		AudienceGrowthRollup: map[string]*types.AudienceGrowthRollup{
			"current":  mapAudienceGrowthRollup(currentRollup),
			"previous": mapAudienceGrowthRollup(previousRollup),
		},
	}, nil
}

// GetPublishingBehaviour fetches time-series publishing metrics filtered by media types,
// with rollup comparisons for current and previous periods.
func (s *FacebookAnalyticsService) GetPublishingBehaviour(ctx context.Context, req *types.FacebookRequest) (*types.PublishingBehaviourResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prev := prevPeriodParams(params)
	mediaTypes := req.GetMediaTypes()

	var current *repo.PublishingBehaviourResult
	var currentRollup, previousRollup *repo.PublishingRollupResult
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetPublishingBehaviour(egCtx, params, mediaTypes)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPublishingBehaviour: current query failed")
			r = &repo.PublishingBehaviourResult{}
		}
		current = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPublishingBehaviourRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPublishingBehaviour: current rollup query failed")
			r = &repo.PublishingRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPublishingBehaviourRollup(egCtx, prev)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPublishingBehaviour: previous rollup query failed")
			r = &repo.PublishingRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.PublishingBehaviourResponse{
		Status:              true,
		PublishingBehaviour: mapPublishingBehaviour(current),
		PublishingBehaviourRollup: map[string]*types.PublishingRollup{
			"current":  mapPublishingRollup(currentRollup),
			"previous": mapPublishingRollup(previousRollup),
		},
	}, nil
}

// GetTopPosts fetches the top N posts sorted by the requested metric with optional media type filtering.
func (s *FacebookAnalyticsService) GetTopPosts(ctx context.Context, req *types.FacebookRequest) (*types.TopPostsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetTopPosts(ctx, params, req.GetMediaTypes(), req.GetLimit(15), req.GetOrderBy())
	if err != nil {
		s.logger.Error().Err(err).Msg("GetTopPosts: query failed")
		return nil, err
	}

	return &types.TopPostsResponse{
		Status:   true,
		TopPosts: mapTopPosts(rows, req.GetTimezone()),
	}, nil
}

// GetActiveUsers fetches hourly and daily fan activity distributions concurrently.
// After retrieval, adjustActiveUserHours shifts the hour buckets from UTC+8 to the
// requested timezone to match the PHP implementation's offset convention.
func (s *FacebookAnalyticsService) GetActiveUsers(ctx context.Context, req *types.FacebookRequest) (*types.ActiveUsersResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	var hours *repo.ActiveUsersHoursResult
	var days *repo.ActiveUsersDaysResult
	var hoursErr, daysErr error
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetActiveUsersHours(egCtx, params)
		if err != nil {
			hoursErr = err
			return nil
		}
		hours = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetActiveUsersDays(egCtx, params)
		if err != nil {
			daysErr = err
			return nil
		}
		days = r
		return nil
	})
	_ = eg.Wait()
	if hoursErr != nil {
		s.logger.Error().Err(hoursErr).Msg("GetActiveUsers: hourly query failed")
		return nil, hoursErr
	}
	if daysErr != nil {
		s.logger.Error().Err(daysErr).Msg("GetActiveUsers: daily query failed")
		return nil, daysErr
	}

	adjustActiveUserHours(hours, req.GetTimezone())

	return &types.ActiveUsersResponse{
		Status: true,
		ActiveUsers: &types.ActiveUsersData{
			ActiveUsersHours: mapActiveUsersHours(hours),
			ActiveUsersDays:  mapActiveUsersDays(days),
		},
	}, nil
}

// GetImpressions fetches time-series page impressions and rollup stats for current and previous periods.
func (s *FacebookAnalyticsService) GetImpressions(ctx context.Context, req *types.FacebookRequest) (*types.ImpressionsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prev := prevPeriodParams(params)

	var current *repo.ImpressionsResult
	var currentRollup, previousRollup *repo.ImpressionsRollupResult
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetImpressions(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetImpressions: current query failed")
			r = &repo.ImpressionsResult{}
		}
		current = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetImpressionsRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetImpressions: current rollup query failed")
			r = &repo.ImpressionsRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetImpressionsRollup(egCtx, prev)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetImpressions: previous rollup query failed")
			r = &repo.ImpressionsRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.ImpressionsResponse{
		Status:      true,
		Impressions: mapImpressions(current),
		ImpressionsRollup: map[string]*types.ImpressionsRollup{
			"current":  mapImpressionsRollup(currentRollup),
			"previous": mapImpressionsRollup(previousRollup),
		},
	}, nil
}

// GetEngagement fetches time-series page engagement and rollup stats for current and previous periods.
func (s *FacebookAnalyticsService) GetEngagement(ctx context.Context, req *types.FacebookRequest) (*types.EngagementResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prev := prevPeriodParams(params)

	var current *repo.EngagementResult
	var currentRollup, previousRollup *repo.EngagementRollupResult
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetEngagement(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetEngagement: current query failed")
			r = &repo.EngagementResult{}
		}
		current = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetEngagementRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetEngagement: current rollup query failed")
			r = &repo.EngagementRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetEngagementRollup(egCtx, prev)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetEngagement: previous rollup query failed")
			r = &repo.EngagementRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.EngagementResponse{
		Status: true,
		Engagement: &types.EngagementContainer{
			Engagement: mapEngagement(current),
			EngagementRollup: map[string]*types.EngagementRollup{
				"current":  mapEngagementRollup(currentRollup),
				"previous": mapEngagementRollup(previousRollup),
			},
		},
	}, nil
}

// GetReelsAnalytics fetches time-series reels data and rollup stats for current and previous periods.
// Errors from all three goroutines are captured separately and checked after eg.Wait
// so that only one error is propagated in a well-defined order.
func (s *FacebookAnalyticsService) GetReelsAnalytics(ctx context.Context, req *types.FacebookRequest) (*types.ReelsAnalyticsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prev := prevPeriodParams(params)

	var current *repo.ReelsAnalyticsResult
	var currentRollup, previousRollup *repo.ReelsRollupResult
	var currentErr, currentRollupErr, previousRollupErr error
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetReelsAnalytics(egCtx, params)
		if err != nil {
			currentErr = err
			return nil
		}
		current = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetReelsRollup(egCtx, params)
		if err != nil {
			currentRollupErr = err
			return nil
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetReelsRollup(egCtx, prev)
		if err != nil {
			previousRollupErr = err
			return nil
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()
	if currentErr != nil {
		s.logger.Error().Err(currentErr).Msg("GetReelsAnalytics: current query failed")
		return nil, currentErr
	}
	if currentRollupErr != nil {
		s.logger.Error().Err(currentRollupErr).Msg("GetReelsAnalytics: current rollup query failed")
		return nil, currentRollupErr
	}
	if previousRollupErr != nil {
		s.logger.Error().Err(previousRollupErr).Msg("GetReelsAnalytics: previous rollup query failed")
		return nil, previousRollupErr
	}

	return &types.ReelsAnalyticsResponse{
		Status: true,
		Reels:  mapReels(current),
		ReelsRollup: map[string]*types.ReelsRollup{
			"current":  mapReelsRollup(currentRollup),
			"previous": mapReelsRollup(previousRollup),
		},
	}, nil
}

// GetVideoInsights fetches time-series video data and previous-period rollup.
// The current-period rollup is derived from the time-series result in videoRollupFromTimeSeries
// to avoid a redundant DB round-trip.
func (s *FacebookAnalyticsService) GetVideoInsights(ctx context.Context, req *types.FacebookRequest) (*types.VideoInsightsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prev := prevPeriodParams(params)

	var current *repo.VideoInsightsResult
	var previousRollup *repo.VideoRollupResult
	var currentErr, previousRollupErr error
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetVideoInsights(egCtx, params)
		if err != nil {
			currentErr = err
			return nil
		}
		current = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetVideoRollup(egCtx, prev)
		if err != nil {
			previousRollupErr = err
			return nil
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()
	if currentErr != nil {
		s.logger.Error().Err(currentErr).Msg("GetVideoInsights: current query failed")
		return nil, currentErr
	}
	if previousRollupErr != nil {
		s.logger.Error().Err(previousRollupErr).Msg("GetVideoInsights: previous rollup query failed")
		return nil, previousRollupErr
	}

	return &types.VideoInsightsResponse{
		Status:        true,
		VideoInsights: mapVideoInsights(current),
		VideoRollup: map[string]*types.VideoRollup{
			"current":  mapVideoRollup(videoRollupFromTimeSeries(current)),
			"previous": mapVideoRollup(previousRollup),
		},
	}, nil
}

// GetDemographics combines overview demographics (gender, age) with audience location (country, city).
func (s *FacebookAnalyticsService) GetDemographics(ctx context.Context, req *types.FacebookRequest) (*types.DemographicsResponse, error) {
	resp, err := s.GetOverviewDemographics(ctx, req)
	if err != nil {
		return nil, err
	}
	location, err := s.GetAudienceLocation(ctx, req)
	if err != nil {
		return nil, err
	}
	resp.AudienceCountry = location.AudienceCountry
	resp.AudienceCity = location.AudienceCity
	return resp, nil
}

// GetOverviewDemographics fetches gender, age, and peak-age-group data concurrently from facebook_insights.
func (s *FacebookAnalyticsService) GetOverviewDemographics(ctx context.Context, req *types.FacebookRequest) (*types.DemographicsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	var gender *repo.AudienceGenderResult
	var age *repo.AudienceAgeResult
	var maxGenderAge *repo.MaxGenderAgeResult
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetAudienceGender(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetOverviewDemographics: gender query failed")
			r = &repo.AudienceGenderResult{}
		}
		gender = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetAudienceAge(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetOverviewDemographics: age query failed")
			r = &repo.AudienceAgeResult{}
		}
		age = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetMaxGenderAge(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetOverviewDemographics: max gender age query failed")
			r = &repo.MaxGenderAgeResult{}
		}
		maxGenderAge = r
		return nil
	})
	_ = eg.Wait()

	return &types.DemographicsResponse{
		Status:         true,
		AudienceGender: mapAudienceGender(gender),
		Fans:           gender.Fans,
		AudienceAge:    mapAudienceAge(age),
		MaxGenderAge:   mapMaxGenderAge(maxGenderAge),
	}, nil
}

// GetAudienceLocation fetches country and city fan distributions concurrently from facebook_insights.
func (s *FacebookAnalyticsService) GetAudienceLocation(ctx context.Context, req *types.FacebookRequest) (*types.DemographicsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	var country, city map[string]int32
	var countryErr, cityErr error
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetAudienceCountry(egCtx, params)
		if err != nil {
			countryErr = err
			return nil
		}
		country = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetAudienceCity(egCtx, params)
		if err != nil {
			cityErr = err
			return nil
		}
		city = r
		return nil
	})
	_ = eg.Wait()
	if countryErr != nil {
		s.logger.Error().Err(countryErr).Msg("GetAudienceLocation: country query failed")
		return nil, countryErr
	}
	if cityErr != nil {
		s.logger.Error().Err(cityErr).Msg("GetAudienceLocation: city query failed")
		return nil, cityErr
	}

	return &types.DemographicsResponse{
		Status:          true,
		AudienceCountry: country,
		AudienceCity:    city,
	}, nil
}

// prevPeriodParams creates query parameters for the previous comparison period.
// The previous period dates are pre-calculated in QueryParams during date parsing.
func prevPeriodParams(params *clickhouse.QueryParams) *clickhouse.QueryParams {
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
	if p == nil {
		p = &repo.PostsSummaryResult{}
	}
	if i == nil {
		i = &repo.InsightsSummaryResult{}
	}
	return &types.SummaryMetrics{
		DocCount:               p.DocCount,
		TotalEngagement:        p.TotalEngagement,
		Reactions:              p.Reactions,
		Comments:               p.Comments,
		PostsClicks:            p.PostsClicks,
		Impressions:            p.Impressions,
		Reach:                  p.Reach,
		Repost:                 p.Repost,
		PositiveSentiment:      i.PositiveSentiment,
		NegativeSentiment:      i.NegativeSentiment,
		PageImpressions:        i.PageImpressions,
		PageImpressionsPaid:    i.PageImpressionsPaid,
		PageImpressionsOrganic: i.PageImpressionsOrganic,
		PageEngagements:        i.PageEngagements,
		PagePositiveFeedback:   i.PagePositiveFeedback,
		PageNegativeFeedback:   i.PageNegativeFeedback,
		FanCount:               i.FanCount,
		TalkingAboutCount:      i.TalkingAboutCount,
		PageFollows:            i.PageFollows,
	}
}

// mapAudienceGrowth converts a *repo.AudienceGrowthResult to *types.AudienceGrowthData for the API layer.
func mapAudienceGrowth(r *repo.AudienceGrowthResult) *types.AudienceGrowthData {
	if r == nil {
		r = &repo.AudienceGrowthResult{}
	}
	return &types.AudienceGrowthData{
		ShowData:         int32(r.ShowData),
		FanCount:         emptyInt32Slice(r.FanCount),
		PageFansDaily:    emptyInt32Slice(r.PageFansDaily),
		PageFansByLike:   emptyInt32Slice(r.PageFansByLike),
		PageFansByUnlike: emptyInt32Slice(r.PageFansByUnlike),
		PageImpressions:  emptyInt32Slice(r.PageImpressions),
		PageEngagements:  emptyInt32Slice(r.PageEngagements),
		Buckets:          formatDates(r.Buckets),
	}
}

// mapAudienceGrowthRollup converts a *repo.AudienceGrowthRollupResult to *types.AudienceGrowthRollup for the API layer.
func mapAudienceGrowthRollup(r *repo.AudienceGrowthRollupResult) *types.AudienceGrowthRollup {
	if r == nil {
		r = &repo.AudienceGrowthRollupResult{}
	}
	return &types.AudienceGrowthRollup{
		AvgPageFansByLike:   r.AvgPageFansByLike,
		AvgPageFansByUnlike: r.AvgPageFansByUnlike,
		FanCount:            r.FanCount,
		TalkingAboutCount:   r.TalkingAboutCount,
		DocCount:            r.DocCount,
	}
}

// mapPublishingBehaviour converts a *repo.PublishingBehaviourResult to *types.PublishingBehaviourData for the API layer.
func mapPublishingBehaviour(r *repo.PublishingBehaviourResult) *types.PublishingBehaviourData {
	if r == nil {
		r = &repo.PublishingBehaviourResult{}
	}
	return &types.PublishingBehaviourData{
		ReactionsEngagement: emptyInt32Slice(r.ReactionsEngagement),
		CommentsEngagement:  emptyInt32Slice(r.CommentsEngagement),
		SharesEngagement:    emptyInt32Slice(r.SharesEngagement),
		PaidImpressions:     emptyInt32Slice(r.PaidImpressions),
		OrganicImpressions:  emptyInt32Slice(r.OrganicImpressions),
		ViralImpressions:    emptyInt32Slice(r.ViralImpressions),
		PaidReach:           emptyInt32Slice(r.PaidReach),
		OrganicReach:        emptyInt32Slice(r.OrganicReach),
		ViralReach:          emptyInt32Slice(r.ViralReach),
		Buckets:             formatDates(r.Buckets),
		PostCount:           emptyInt32Slice(r.PostCount),
	}
}

// mapPublishingRollup converts a *repo.PublishingRollupResult to *types.PublishingRollup for the API layer.
func mapPublishingRollup(r *repo.PublishingRollupResult) *types.PublishingRollup {
	if r == nil {
		r = &repo.PublishingRollupResult{}
	}
	return &types.PublishingRollup{
		DocCount:        r.DocCount,
		TotalEngagement: r.TotalEngagement,
		Reactions:       r.Reactions,
		Comments:        r.Comments,
		PostClicks:      r.PostClicks,
		Impressions:     r.Impressions,
		Shares:          r.Shares,
	}
}

// mapTopPosts converts TopPostRow rows (which may have multiple rows per post due to media assets)
// into a deduplicated []TopPost with MediaAssets collected per post.
func mapTopPosts(rows []repo.TopPostRow, timezone string) []types.TopPost {
	type indexedPost struct {
		index int
		post  *types.TopPost
	}

	// seen tracks the first occurrence of each post_id to collect media assets into it.
	seen := make(map[string]*indexedPost, len(rows))
	out := make([]types.TopPost, 0, len(rows))

	for _, row := range rows {
		entry, ok := seen[row.PostID]
		if !ok {
			post := types.TopPost{
				PageName:                     row.PageName,
				PageID:                       row.PageID,
				PostID:                       row.PostID,
				Permalink:                    row.Permalink,
				StatusType:                   row.StatusType,
				MediaType:                    row.MediaType,
				VideoID:                      row.VideoID,
				Category:                     row.Category,
				PublishedBy:                  row.PublishedBy,
				PublishedByURL:               row.PublishedByURL,
				SharedFromName:               row.SharedFromName,
				SharedFromID:                 row.SharedFromID,
				SharedFromLink:               row.SharedFromLink,
				Like:                         row.Like,
				Love:                         row.Love,
				Haha:                         row.Haha,
				Wow:                          row.Wow,
				Sad:                          row.Sad,
				Angry:                        row.Angry,
				Total:                        row.Total,
				Shares:                       row.Shares,
				Comments:                     row.Comments,
				PostClicks:                   row.PostClicks,
				TotalEngagement:              row.TotalEngagement,
				PostEngagedUsers:             row.PostEngagedUsers,
				DayOfWeek:                    row.DayOfWeek,
				HourOfDay:                    row.HourOfDay,
				CreatedTime:                  formatTimeInTimezone(row.CreatedTime, timezone),
				UpdatedTime:                  formatTimeInTimezone(row.UpdatedTime, timezone),
				SavingTime:                   formatTimeInTimezone(row.SavingTime, timezone),
				MessageTags:                  row.MessageTags,
				PostMetadata:                 row.PostMetadata,
				Caption:                      row.Caption,
				Description:                  row.Description,
				FullPicture:                  row.FullPicture,
				Link:                         row.Link,
				PostImpressions:              row.PostImpressions,
				PostImpressionsUnique:        row.PostImpressionsUnique,
				PostImpressionsPaid:          row.PostImpressionsPaid,
				PostImpressionsPaidUnique:    row.PostImpressionsPaidUnique,
				PostImpressionsOrganic:       row.PostImpressionsOrganic,
				PostImpressionsOrganicUnique: row.PostImpressionsOrganicUnique,
				PostImpressionsViral:         row.PostImpressionsViral,
				PostImpressionsViralUnique:   row.PostImpressionsViralUnique,
				PostVideoViews:               row.PostVideoViews,
				TotalImpressions:             row.TotalImpressions,
				MediaAssets:                  []types.MediaAsset{},
			}
			out = append(out, post)
			entry = &indexedPost{index: len(out) - 1, post: &out[len(out)-1]}
			seen[row.PostID] = entry
		}

		if row.MediaID != "" || row.MediaCaption != "" || row.MediaLink != "" || row.AssetType != "" {
			entry.post.MediaAssets = append(entry.post.MediaAssets, types.MediaAsset{
				MediaID:      row.MediaID,
				Caption:      row.MediaCaption,
				Link:         row.MediaLink,
				AssetType:    row.AssetType,
				CallToAction: row.CallToAction,
				CreatedAt:    formatTimeInTimezone(row.AssetCreatedAt, timezone),
			})
		}
	}

	return out
}

// mapActiveUsersHours converts a *repo.ActiveUsersHoursResult to *types.ActiveUsersHours for the API layer.
func mapActiveUsersHours(r *repo.ActiveUsersHoursResult) *types.ActiveUsersHours {
	if r == nil {
		r = &repo.ActiveUsersHoursResult{}
	}
	return &types.ActiveUsersHours{
		Buckets:      emptyInt32Slice(r.Buckets),
		Values:       emptyInt32Slice(r.Values),
		HighestValue: r.HighestValue,
		HighestHour:  r.HighestHour,
	}
}

// mapActiveUsersDays converts a *repo.ActiveUsersDaysResult to *types.ActiveUsersDays for the API layer.
func mapActiveUsersDays(r *repo.ActiveUsersDaysResult) *types.ActiveUsersDays {
	if r == nil {
		r = &repo.ActiveUsersDaysResult{}
	}
	return &types.ActiveUsersDays{
		Buckets:      emptyStringSlice(r.Buckets),
		Values:       emptyInt32Slice(r.Values),
		HighestValue: r.HighestValue,
		HighestDay:   r.HighestDay,
	}
}

// mapImpressions converts a *repo.ImpressionsResult to *types.ImpressionsData for the API layer.
func mapImpressions(r *repo.ImpressionsResult) *types.ImpressionsData {
	if r == nil {
		r = &repo.ImpressionsResult{}
	}
	return &types.ImpressionsData{
		PageImpressions: emptyInt32Slice(r.PageImpressions),
		Buckets:         formatDates(r.Buckets),
	}
}

// mapImpressionsRollup converts a *repo.ImpressionsRollupResult to *types.ImpressionsRollup for the API layer.
func mapImpressionsRollup(r *repo.ImpressionsRollupResult) *types.ImpressionsRollup {
	if r == nil {
		r = &repo.ImpressionsRollupResult{}
	}
	return &types.ImpressionsRollup{
		TotalImpressions:      r.TotalImpressions,
		AvgImpressionsPerDay:  r.AvgImpressionsPerDay,
		AvgImpressionsPerWeek: r.AvgImpressionsPerWeek,
	}
}

// mapEngagement converts a *repo.EngagementResult to *types.EngagementData for the API layer.
func mapEngagement(r *repo.EngagementResult) *types.EngagementData {
	if r == nil {
		r = &repo.EngagementResult{}
	}
	return &types.EngagementData{
		PageEngagements: emptyInt32Slice(r.PageEngagements),
		Buckets:         formatDates(r.Buckets),
	}
}

// mapEngagementRollup converts a *repo.EngagementRollupResult to *types.EngagementRollup for the API layer.
func mapEngagementRollup(r *repo.EngagementRollupResult) *types.EngagementRollup {
	if r == nil {
		r = &repo.EngagementRollupResult{}
	}
	return &types.EngagementRollup{
		PageEngagements:       r.PageEngagements,
		AvgEngagementsPerDay:  r.AvgEngagementsPerDay,
		AvgEngagementsPerWeek: r.AvgEngagementsPerWeek,
	}
}

// mapReels converts a *repo.ReelsAnalyticsResult to *types.ReelsData for the API layer.
func mapReels(r *repo.ReelsAnalyticsResult) *types.ReelsData {
	if r == nil {
		r = &repo.ReelsAnalyticsResult{}
	}
	return &types.ReelsData{
		Buckets:             formatDates(r.Buckets),
		TotalReels:          emptyInt32Slice(r.TotalReels),
		TotalSecondsWatched: emptyFloat64Slice(r.TotalSecondsWatched),
		InitialPlays:        emptyInt32Slice(r.InitialPlays),
		Engagement:          emptyInt32Slice(r.Engagement),
		Reactions:           emptyInt32Slice(r.Reactions),
		Comments:            emptyInt32Slice(r.Comments),
		Shares:              emptyInt32Slice(r.Shares),
		ShowData:            r.ShowData,
	}
}

// mapReelsRollup converts a *repo.ReelsRollupResult to *types.ReelsRollup for the API layer.
func mapReelsRollup(r *repo.ReelsRollupResult) *types.ReelsRollup {
	if r == nil {
		r = &repo.ReelsRollupResult{}
	}
	return &types.ReelsRollup{
		TotalReels:            r.TotalReels,
		AverageSecondsWatched: r.AverageSecondsWatched,
		TotalSecondsWatched:   r.TotalSecondsWatched,
		InitialPlays:          r.InitialPlays,
		Reach:                 r.Reach,
		Engagement:            r.Engagement,
		Reactions:             r.Reactions,
		Comments:              r.Comments,
		Shares:                r.Shares,
	}
}

// mapVideoInsights converts a *repo.VideoInsightsResult to *types.VideoInsightsData for the API layer.
func mapVideoInsights(r *repo.VideoInsightsResult) *types.VideoInsightsData {
	if r == nil {
		r = &repo.VideoInsightsResult{}
	}
	return &types.VideoInsightsData{
		Buckets:         formatDates(r.Buckets),
		TotalViewTime:   emptyFloat64Slice(r.TotalViewTime),
		OrganicViewTime: emptyFloat64Slice(r.OrganicViewTime),
		PaidViewTime:    emptyFloat64Slice(r.PaidViewTime),
		TotalViews:      emptyInt32Slice(r.TotalViews),
		OrganicViews:    emptyInt32Slice(r.OrganicViews),
		PaidViews:       emptyInt32Slice(r.PaidViews),
		Comments:        emptyInt32Slice(r.Comments),
		Reactions:       emptyInt32Slice(r.Reactions),
		Shares:          emptyInt32Slice(r.Shares),
		TotalPosts:      emptyInt32Slice(r.TotalPosts),
	}
}

// videoRollupFromTimeSeries derives the current-period VideoRollupResult from the
// already-fetched time-series data, eliminating a redundant DB round-trip.
// All rollup fields are simple sums of daily values; WITH FILL zeros contribute 0
// to all sums and do not affect totals.
func videoRollupFromTimeSeries(v *repo.VideoInsightsResult) *repo.VideoRollupResult {
	if v == nil {
		return &repo.VideoRollupResult{}
	}
	r := &repo.VideoRollupResult{}
	for i := range v.TotalViewTime {
		r.TotalViewTime += v.TotalViewTime[i]
		r.OrganicViewTime += v.OrganicViewTime[i]
		r.PaidViewTime += v.PaidViewTime[i]
		r.TotalViews += v.TotalViews[i]
		r.OrganicViews += v.OrganicViews[i]
		r.PaidViews += v.PaidViews[i]
		r.Comments += v.Comments[i]
		r.Reactions += v.Reactions[i]
		r.Shares += v.Shares[i]
		r.TotalPosts += v.TotalPosts[i]
	}
	return r
}

// mapVideoRollup converts a *repo.VideoRollupResult to *types.VideoRollup for the API layer.
func mapVideoRollup(r *repo.VideoRollupResult) *types.VideoRollup {
	if r == nil {
		r = &repo.VideoRollupResult{}
	}
	return &types.VideoRollup{
		TotalViewTime:   r.TotalViewTime,
		OrganicViewTime: r.OrganicViewTime,
		PaidViewTime:    r.PaidViewTime,
		TotalViews:      r.TotalViews,
		OrganicViews:    r.OrganicViews,
		PaidViews:       r.PaidViews,
		Comments:        r.Comments,
		Reactions:       r.Reactions,
		Shares:          r.Shares,
		TotalPosts:      r.TotalPosts,
	}
}

// mapAudienceGender converts a *repo.AudienceGenderResult to a gender-keyed map[string]int32 for the API layer.
func mapAudienceGender(r *repo.AudienceGenderResult) map[string]int32 {
	if r == nil {
		r = &repo.AudienceGenderResult{}
	}
	return map[string]int32{
		"M": r.M,
		"F": r.F,
		"U": r.U,
	}
}

// mapAudienceAge converts a *repo.AudienceAgeResult to *types.AudienceAgeData for the API layer.
func mapAudienceAge(r *repo.AudienceAgeResult) *types.AudienceAgeData {
	if r == nil {
		r = &repo.AudienceAgeResult{}
	}
	breakdown := &types.AgeBreakdown{
		Age65Plus: r.Age65Plus,
		Age55To64: r.Age55To64,
		Age45To54: r.Age45To54,
		Age35To44: r.Age35To44,
		Age25To34: r.Age25To34,
		Age18To34: r.Age18To34,
		Age13To17: r.Age13To17,
	}
	return &types.AudienceAgeData{
		FansAge: breakdown,
		MaxAge:  maxAgeValue(breakdown),
	}
}

// mapMaxGenderAge converts a *repo.MaxGenderAgeResult to *types.MaxGenderAge for the API layer.
func mapMaxGenderAge(r *repo.MaxGenderAgeResult) *types.MaxGenderAge {
	if r == nil {
		r = &repo.MaxGenderAgeResult{}
	}
	return &types.MaxGenderAge{
		MaxValue: r.MaxValue,
		Age:      r.Age,
		Gender:   r.Gender,
	}
}

// adjustActiveUserHours shifts hour buckets from the UTC+8 base stored in ClickHouse
// to the user's requested timezone. The +8 offset in the formula undoes the stored base;
// the timezone offset is then applied to produce local hours in [0, 23].
func adjustActiveUserHours(r *repo.ActiveUsersHoursResult, timezone string) {
	if r == nil || len(r.Buckets) == 0 {
		return
	}

	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return
	}

	_, offsetSeconds := time.Now().In(loc).Zone()
	interval := int32(math.Round(float64(offsetSeconds)/3600.0)) + 8

	type pair struct {
		bucket int32
		value  int32
	}

	pairs := make([]pair, 0, len(r.Buckets))
	for i, bucket := range r.Buckets {
		adjusted := bucket + interval
		if adjusted >= 24 {
			adjusted -= 24
		}
		if adjusted < 0 {
			adjusted += 24
		}
		value := int32(0)
		if i < len(r.Values) {
			value = r.Values[i]
		}
		pairs = append(pairs, pair{bucket: adjusted, value: value})
	}

	sort.Slice(pairs, func(i, j int) bool { return pairs[i].bucket < pairs[j].bucket })
	r.Buckets = r.Buckets[:0]
	r.Values = r.Values[:0]
	for _, p := range pairs {
		r.Buckets = append(r.Buckets, p.bucket)
		r.Values = append(r.Values, p.value)
	}

	r.HighestHour += interval
	if r.HighestHour >= 24 {
		r.HighestHour -= 24
	}
	if r.HighestHour < 0 {
		r.HighestHour += 24
	}
}

// formatDates converts a []time.Time to []string in "YYYY-MM-DD" format for JSON responses.
func formatDates(values []time.Time) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		out = append(out, value.Format("2006-01-02"))
	}
	return out
}

// formatTime formats a time.Time as RFC3339, returning "" for zero/epoch times.
func formatTime(value time.Time) string {
	if value.IsZero() || value.Year() == 1970 {
		return ""
	}
	return value.Format(time.RFC3339)
}

// formatTimeInTimezone relocates timestamps to the requested timezone before serializing.
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

// emptyInt32Slice returns s, or an empty slice if s is nil, ensuring JSON encodes as [] not null.
func emptyInt32Slice(values []int32) []int32 {
	if len(values) == 0 {
		return []int32{}
	}
	return values
}

// emptyFloat64Slice returns s, or an empty slice if s is nil, ensuring JSON encodes as [] not null.
func emptyFloat64Slice(values []float64) []float64 {
	if len(values) == 0 {
		return []float64{}
	}
	return values
}

// emptyStringSlice returns s, or an empty slice if s is nil, ensuring JSON encodes as [] not null.
func emptyStringSlice(values []string) []string {
	if len(values) == 0 {
		return []string{}
	}
	return values
}

// maxAgeValue returns the largest fan count across all age brackets.
// Used to drive the highlighted bar in the age demographics chart.
func maxAgeValue(age *types.AgeBreakdown) int32 {
	values := []int32{
		age.Age65Plus,
		age.Age55To64,
		age.Age45To54,
		age.Age35To44,
		age.Age25To34,
		age.Age18To34,
		age.Age13To17,
	}
	var max int32
	for _, value := range values {
		if value > max {
			max = value
		}
	}
	return max
}
