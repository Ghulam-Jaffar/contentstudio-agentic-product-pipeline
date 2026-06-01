package gmb

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/gmb"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/gmb"
)

type Service interface {
	GetSummary(ctx context.Context, req *types.GMBRequest) (*types.SummaryResponse, error)
	GetImpressions(ctx context.Context, req *types.GMBRequest) (*types.ImpressionsResponse, error)
	GetActions(ctx context.Context, req *types.GMBRequest) (*types.ActionsResponse, error)
	GetSearchKeywords(ctx context.Context, req *types.SearchKeywordsRequest) (*types.SearchKeywordsResponse, error)
	GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error)
	GetPublishingBehavior(ctx context.Context, req *types.GMBRequest) (*types.PublishingBehaviorResponse, error)
	GetReviews(ctx context.Context, req *types.GMBRequest) (*types.ReviewsResponse, error)
	GetMediaActivity(ctx context.Context, req *types.GMBRequest) (*types.MediaActivityResponse, error)
}

type GMBAnalyticsService struct {
	repo   *repo.Repository
	logger zerolog.Logger
}

var _ Service = (*GMBAnalyticsService)(nil)

func NewGMBAnalyticsService(r *repo.Repository, logger zerolog.Logger) *GMBAnalyticsService {
	return &GMBAnalyticsService{
		repo:   r,
		logger: logger.With().Str("service", "gmb-analytics").Logger(),
	}
}

func (s *GMBAnalyticsService) GetSummary(ctx context.Context, req *types.GMBRequest) (*types.SummaryResponse, error) {
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
			s.logger.Error().Err(err).Msg("GetSummary: failed to get current period")
			r = &repo.SummaryResult{}
		}
		current = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetSummary(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get previous period")
			r = &repo.SummaryResult{}
		}
		previous = r
		return nil
	})
	_ = eg.Wait()

	return &types.SummaryResponse{
		Status: true,
		Overview: map[string]*types.SummaryMetrics{
			"current":  mapSummary(current),
			"previous": mapSummary(previous),
		},
	}, nil
}

func (s *GMBAnalyticsService) GetImpressions(ctx context.Context, req *types.GMBRequest) (*types.ImpressionsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)

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
		ImpressionsRolup: map[string]*types.ImpressionsRollupData{
			"current":  mapImpressionsRollup(currentRollup),
			"previous": mapImpressionsRollup(previousRollup),
		},
	}, nil
}

func (s *GMBAnalyticsService) GetActions(ctx context.Context, req *types.GMBRequest) (*types.ActionsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)

	var actions *repo.ActionsResult
	var currentRollup, previousRollup *repo.ActionsRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetActions(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetActions: failed to get time-series data")
			r = &repo.ActionsResult{}
		}
		actions = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetActionsRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetActions: failed to get current rollup")
			r = &repo.ActionsRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetActionsRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetActions: failed to get previous rollup")
			r = &repo.ActionsRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.ActionsResponse{
		Status:  true,
		Actions: mapActions(actions),
		ActionsRollup: map[string]*types.ActionsRollupData{
			"current":  mapActionsRollup(currentRollup),
			"previous": mapActionsRollup(previousRollup),
		},
	}, nil
}

func (s *GMBAnalyticsService) GetSearchKeywords(ctx context.Context, req *types.SearchKeywordsRequest) (*types.SearchKeywordsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	rows, err := s.repo.GetSearchKeywords(ctx, params, req.GetLimit())
	if err != nil {
		s.logger.Error().Err(err).Msg("GetSearchKeywords: failed to get keywords")
		return &types.SearchKeywordsResponse{Status: true, Keywords: []types.SearchKeyword{}}, nil
	}

	return &types.SearchKeywordsResponse{
		Status:   true,
		Keywords: mapSearchKeywords(rows),
	}, nil
}

func (s *GMBAnalyticsService) GetTopPosts(ctx context.Context, req *types.TopPostsRequest) (*types.TopPostsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	posts, err := s.repo.GetTopPosts(ctx, params, req.GetLimit(), req.GetOrderBy())
	if err != nil {
		s.logger.Error().Err(err).Msg("GetTopPosts: failed to get top posts")
		return &types.TopPostsResponse{Status: true, Posts: []types.TopPost{}}, nil
	}

	return &types.TopPostsResponse{
		Status: true,
		Posts:  mapTopPosts(posts, req.GetTimezone()),
	}, nil
}

func (s *GMBAnalyticsService) GetPublishingBehavior(ctx context.Context, req *types.GMBRequest) (*types.PublishingBehaviorResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	var publishing *repo.PublishingResult
	var topicTypes []repo.TopicTypeRow
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetPublishingBehavior(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPublishingBehavior: failed to get time-series data")
			r = &repo.PublishingResult{}
		}
		publishing = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetTopicTypes(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPublishingBehavior: failed to get topic types")
		}
		topicTypes = r
		return nil
	})
	_ = eg.Wait()

	return &types.PublishingBehaviorResponse{
		Status:              true,
		PublishingBehaviour: mapPublishingBehavior(publishing, topicTypes),
	}, nil
}

func (s *GMBAnalyticsService) GetReviews(ctx context.Context, req *types.GMBRequest) (*types.ReviewsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)

	var summary *repo.ReviewsSummaryResult
	var timeSeries *repo.ReviewsTimeSeriesResult
	var reviewsList []repo.ReviewRow
	var currentRollup, previousRollup *repo.ReviewsRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetReviewsSummary(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetReviews: failed to get summary")
			r = &repo.ReviewsSummaryResult{}
		}
		summary = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetReviewsTimeSeries(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetReviews: failed to get time-series")
			r = &repo.ReviewsTimeSeriesResult{}
		}
		timeSeries = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetReviewsList(egCtx, params, 50)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetReviews: failed to get reviews list")
		}
		reviewsList = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetReviewsRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetReviews: failed to get current rollup")
			r = &repo.ReviewsRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetReviewsRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetReviews: failed to get previous rollup")
			r = &repo.ReviewsRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.ReviewsResponse{
		Status:  true,
		Reviews: mapReviews(summary, timeSeries, reviewsList),
		ReviewsRollup: map[string]*types.ReviewsRollupData{
			"current":  mapReviewsRollup(currentRollup),
			"previous": mapReviewsRollup(previousRollup),
		},
	}, nil
}

func (s *GMBAnalyticsService) GetMediaActivity(ctx context.Context, req *types.GMBRequest) (*types.MediaActivityResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)

	var media *repo.MediaResult
	var currentRollup, previousRollup *repo.MediaRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetMediaActivity(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetMediaActivity: failed to get time-series data")
			r = &repo.MediaResult{}
		}
		media = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetMediaActivityRollup(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetMediaActivity: failed to get current rollup")
			r = &repo.MediaRollupResult{}
		}
		currentRollup = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetMediaActivityRollup(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetMediaActivity: failed to get previous rollup")
			r = &repo.MediaRollupResult{}
		}
		previousRollup = r
		return nil
	})
	_ = eg.Wait()

	return &types.MediaActivityResponse{
		Status:        true,
		MediaActivity: mapMediaActivity(media),
		MediaActivityRollup: map[string]*types.MediaActivityRollupData{
			"current":  mapMediaActivityRollup(currentRollup),
			"previous": mapMediaActivityRollup(previousRollup),
		},
	}, nil
}

// --- Private helpers ---

func (s *GMBAnalyticsService) prevPeriodParams(params *clickhouse.QueryParams) *clickhouse.QueryParams {
	return &clickhouse.QueryParams{
		AccountIDs: params.AccountIDs,
		DateFrom:   params.PrevDateFrom,
		DateTo:     params.PrevDateTo,
		Timezone:   params.Timezone,
		DayCount:   params.DayCount,
	}
}

func mapSummary(r *repo.SummaryResult) *types.SummaryMetrics {
	return &types.SummaryMetrics{
		TotalImpressions:  r.TotalImpressions,
		SearchImpressions: r.SearchImpressions,
		MapsImpressions:   r.MapsImpressions,
		WebsiteClicks:     r.WebsiteClicks,
		CallClicks:        r.CallClicks,
		DirectionRequests: r.DirectionRequests,
		OtherActions:      r.OtherActions,
		TotalReviews:      r.TotalReviews,
		AverageRating:     r.AverageRating,
		TotalPosts:        r.TotalPosts,
	}
}

func mapImpressions(r *repo.ImpressionsResult) *types.ImpressionsData {
	buckets := formatBuckets(r.Buckets)
	return &types.ImpressionsData{
		DesktopMaps:           emptyInt64Slice(cumSum(r.DesktopMapsDaily)),
		DesktopSearch:         emptyInt64Slice(cumSum(r.DesktopSearchDaily)),
		MobileMaps:            emptyInt64Slice(cumSum(r.MobileMapsDaily)),
		MobileSearch:          emptyInt64Slice(cumSum(r.MobileSearchDaily)),
		TotalImpressions:      emptyInt64Slice(cumSum(r.TotalImpressionsDaily)),
		DesktopMapsDaily:      emptyInt64Slice(r.DesktopMapsDaily),
		DesktopSearchDaily:    emptyInt64Slice(r.DesktopSearchDaily),
		MobileMapsDaily:       emptyInt64Slice(r.MobileMapsDaily),
		MobileSearchDaily:     emptyInt64Slice(r.MobileSearchDaily),
		TotalImpressionsDaily: emptyInt64Slice(r.TotalImpressionsDaily),
		ShowData:              r.ShowData,
		Buckets:               buckets,
	}
}

func mapImpressionsRollup(r *repo.ImpressionsRollupResult) *types.ImpressionsRollupData {
	return &types.ImpressionsRollupData{
		TotalImpressions: r.TotalImpressions,
		DesktopMaps:      r.DesktopMaps,
		DesktopSearch:    r.DesktopSearch,
		MobileMaps:       r.MobileMaps,
		MobileSearch:     r.MobileSearch,
		AvgImpressions:   r.AvgImpressions,
	}
}

func mapActions(r *repo.ActionsResult) *types.ActionsData {
	buckets := formatBuckets(r.Buckets)
	return &types.ActionsData{
		CallClicks:             emptyInt64Slice(cumSum(r.CallClicksDaily)),
		WebsiteClicks:          emptyInt64Slice(cumSum(r.WebsiteClicksDaily)),
		DirectionRequests:      emptyInt64Slice(cumSum(r.DirectionRequestsDaily)),
		OtherActions:           emptyInt64Slice(cumSum(r.OtherActionsDaily)),
		CallClicksDaily:        emptyInt64Slice(r.CallClicksDaily),
		WebsiteClicksDaily:     emptyInt64Slice(r.WebsiteClicksDaily),
		DirectionRequestsDaily: emptyInt64Slice(r.DirectionRequestsDaily),
		OtherActionsDaily:      emptyInt64Slice(r.OtherActionsDaily),
		ShowData:               r.ShowData,
		Buckets:                buckets,
	}
}

func mapActionsRollup(r *repo.ActionsRollupResult) *types.ActionsRollupData {
	return &types.ActionsRollupData{
		TotalCallClicks:        r.TotalCallClicks,
		TotalWebsiteClicks:     r.TotalWebsiteClicks,
		TotalDirectionRequests: r.TotalDirectionRequests,
		TotalOtherActions:      r.TotalOtherActions,
		AvgActions:             r.AvgActions,
	}
}

func mapSearchKeywords(rows []repo.SearchKeywordRow) []types.SearchKeyword {
	if rows == nil {
		return []types.SearchKeyword{}
	}
	result := make([]types.SearchKeyword, len(rows))
	for i, r := range rows {
		result[i] = types.SearchKeyword{
			Keyword:              r.Keyword,
			ImpressionsValue:     r.ImpressionsValue,
			ImpressionsThreshold: r.ImpressionsThreshold,
			KeywordMonth:         r.KeywordMonth.Format("2006-01"),
		}
	}
	return result
}

func mapTopPosts(rows []repo.TopPostRow, timezone string) []types.TopPost {
	if rows == nil {
		return []types.TopPost{}
	}
	result := make([]types.TopPost, len(rows))
	for i, r := range rows {
		result[i] = types.TopPost{
			PostName:        r.PostName,
			Summary:         r.Summary,
			State:           r.State,
			TopicType:       r.TopicType,
			SearchURL:       r.SearchURL,
			MediaNames:      emptyStringSlice(r.MediaNames),
			MediaFormats:    emptyStringSlice(r.MediaFormats),
			MediaGoogleURLs: emptyStringSlice(r.MediaGoogleURLs),
			CreatedAt:       formatTimeInTimezone(r.CreatedAt, timezone),
		}
	}
	return result
}

func mapPublishingBehavior(pub *repo.PublishingResult, topicTypes []repo.TopicTypeRow) *types.PublishingBehaviorData {
	buckets := formatBuckets(pub.Buckets)
	tt := make([]types.TopicTypeCount, len(topicTypes))
	for i, t := range topicTypes {
		tt[i] = types.TopicTypeCount{Name: t.TopicType, Count: t.Count}
	}
	if tt == nil {
		tt = []types.TopicTypeCount{}
	}
	return &types.PublishingBehaviorData{
		Buckets:    buckets,
		PostCount:  emptyInt64Slice(pub.PostCount),
		TopicTypes: tt,
	}
}

func mapReviews(summary *repo.ReviewsSummaryResult, ts *repo.ReviewsTimeSeriesResult, list []repo.ReviewRow) *types.ReviewsData {
	buckets := formatBuckets(ts.Buckets)
	starDist := map[string]int64{
		"1": summary.Star1,
		"2": summary.Star2,
		"3": summary.Star3,
		"4": summary.Star4,
		"5": summary.Star5,
	}
	reviewItems := make([]types.ReviewItem, len(list))
	for i, r := range list {
		reviewItems[i] = types.ReviewItem{
			ReviewID:                r.ReviewID,
			ReviewerDisplayName:     r.ReviewerDisplayName,
			ReviewerProfilePhotoURL: r.ReviewerProfilePhotoURL,
			StarRating:              r.StarRating,
			Comment:                 r.Comment,
			ReplyComment:            r.ReplyComment,
			CreatedAt:               formatTimeInTimezone(r.CreatedAt, "UTC"),
		}
	}
	if reviewItems == nil {
		reviewItems = []types.ReviewItem{}
	}
	return &types.ReviewsData{
		AvgRating:        summary.AvgRating,
		TotalReviews:     summary.TotalReviews,
		StarDistribution: starDist,
		Buckets:          buckets,
		DailyReviews:     emptyInt64Slice(ts.DailyReviews),
		ReviewsList:      reviewItems,
	}
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

func mapReviewsRollup(r *repo.ReviewsRollupResult) *types.ReviewsRollupData {
	return &types.ReviewsRollupData{
		TotalReviews: r.TotalReviews,
		AvgRating:    r.AvgRating,
	}
}

func mapMediaActivity(r *repo.MediaResult) *types.MediaActivityData {
	buckets := formatBuckets(r.Buckets)
	return &types.MediaActivityData{
		PhotoCount:      emptyInt64Slice(cumSum(r.PhotoCountDaily)),
		VideoCount:      emptyInt64Slice(cumSum(r.VideoCountDaily)),
		PhotoCountDaily: emptyInt64Slice(r.PhotoCountDaily),
		VideoCountDaily: emptyInt64Slice(r.VideoCountDaily),
		ShowData:        r.ShowData,
		Buckets:         buckets,
	}
}

func mapMediaActivityRollup(r *repo.MediaRollupResult) *types.MediaActivityRollupData {
	return &types.MediaActivityRollupData{
		TotalPhotos: r.TotalPhotos,
		TotalVideos: r.TotalVideos,
		AvgMedia:    r.AvgMedia,
	}
}

// --- Utility functions ---

func formatBuckets(times []time.Time) []string {
	if times == nil {
		return []string{}
	}
	buckets := make([]string, len(times))
	for i, t := range times {
		buckets[i] = t.Format("2006-01-02")
	}
	return buckets
}

func cumSum(daily []int64) []int64 {
	if daily == nil {
		return nil
	}
	result := make([]int64, len(daily))
	var sum int64
	for i, v := range daily {
		sum += v
		result[i] = sum
	}
	return result
}

func emptyInt64Slice(s []int64) []int64 {
	if s == nil {
		return []int64{}
	}
	return s
}

func emptyStringSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
