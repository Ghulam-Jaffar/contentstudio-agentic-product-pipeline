// Package youtube provides the business logic layer for YouTube analytics.
// It orchestrates repository queries, handles concurrent data fetching via errgroup,
// computes previous-period comparisons, and maps ClickHouse results to API response types.
//
// Migrated from PHP: YouTubeAnalyticsController (contentstudio-backend).
package youtube

import (
	"context"
	"math"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/youtube"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/youtube"
)

// Service defines the interface for YouTube analytics business logic.
type Service interface {
	GetSummary(ctx context.Context, req *types.YoutubeRequest) (*types.SummaryResponse, error)
	GetSubscriberTrend(ctx context.Context, req *types.YoutubeRequest) (*types.SubscriberTrendResponse, error)
	GetDynamicSubscriberTrend(ctx context.Context, req *types.YoutubeRequest) (*types.SubscriberTrendResponse, error)
	GetEngagementTrend(ctx context.Context, req *types.YoutubeRequest) (*types.EngagementTrendResponse, error)
	GetDynamicEngagementTrend(ctx context.Context, req *types.YoutubeRequest) (*types.EngagementTrendResponse, error)
	GetViewsTrend(ctx context.Context, req *types.YoutubeRequest) (*types.ViewsTrendResponse, error)
	GetDynamicViewsTrend(ctx context.Context, req *types.YoutubeRequest) (*types.ViewsTrendResponse, error)
	GetWatchTimeTrend(ctx context.Context, req *types.YoutubeRequest) (*types.WatchTimeTrendResponse, error)
	GetDynamicWatchTimeTrend(ctx context.Context, req *types.YoutubeRequest) (*types.WatchTimeTrendResponse, error)
	GetFindVideo(ctx context.Context, req *types.YoutubeRequest) (*types.FindVideoResponse, error)
	GetVideoSharing(ctx context.Context, req *types.YoutubeRequest) (*types.VideoSharingResponse, error)
	GetTopVideos(ctx context.Context, req *types.YoutubeRequest) (*types.TopVideosResponse, error)
	GetLeastVideos(ctx context.Context, req *types.YoutubeRequest) (*types.LeastVideosResponse, error)
	GetSortedTopVideos(ctx context.Context, req *types.TopVideosRequest) (*types.SortedTopVideosResponse, error)
	GetPerformanceAndSchedule(ctx context.Context, req *types.YoutubeRequest) (*types.PerformanceScheduleResponse, error)
}

// YoutubeAnalyticsService implements YouTube analytics business logic.
type YoutubeAnalyticsService struct {
	repo   *repo.Repository
	logger zerolog.Logger
}

var _ Service = (*YoutubeAnalyticsService)(nil)

// NewYoutubeAnalyticsService creates a new service with the given repository and logger.
func NewYoutubeAnalyticsService(r *repo.Repository, logger zerolog.Logger) *YoutubeAnalyticsService {
	return &YoutubeAnalyticsService{
		repo:   r,
		logger: logger.With().Str("service", "youtube-analytics").Logger(),
	}
}

// GetSummary fetches current and previous period metrics concurrently from youtube_activity_insights,
// youtube_channels, and youtube_videos, and returns them under "current" and "previous" keys.
func (s *YoutubeAnalyticsService) GetSummary(ctx context.Context, req *types.YoutubeRequest) (*types.SummaryResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)

	var actCurr, actPrev *repo.ActivitySummaryResult
	var subsCurr, subsPrev *repo.SubscriberSummaryResult
	var vidCurr, vidPrev *repo.VideoCountResult

	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetActivitySummary(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get current activity")
			r = &repo.ActivitySummaryResult{}
		}
		actCurr = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetActivitySummary(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get previous activity")
			r = &repo.ActivitySummaryResult{}
		}
		actPrev = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetSubscriberSummary(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get current subscribers")
			r = &repo.SubscriberSummaryResult{}
		}
		subsCurr = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetSubscriberSummary(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get previous subscribers")
			r = &repo.SubscriberSummaryResult{}
		}
		subsPrev = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetVideoCount(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get current video count")
			r = &repo.VideoCountResult{}
		}
		vidCurr = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetVideoCount(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetSummary: failed to get previous video count")
			r = &repo.VideoCountResult{}
		}
		vidPrev = r
		return nil
	})
	_ = eg.Wait()

	curr := mapSummary(actCurr, subsCurr, vidCurr)
	prev := mapSummary(actPrev, subsPrev, vidPrev)

	return &types.SummaryResponse{
		Status: true,
		Overview: &types.SummaryOverview{
			Current:    curr,
			Previous:   prev,
			Percentage: mapSummaryPercentage(curr, prev),
			Difference: mapSummaryDifference(curr, prev),
		},
	}, nil
}

// GetSubscriberTrend returns daily subscriber trend data.
// When the first SubscribersTotal element is zero the latest known count is used as a fallback.
func (s *YoutubeAnalyticsService) GetSubscriberTrend(ctx context.Context, req *types.YoutubeRequest) (*types.SubscriberTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetSubscriberTrend(ctx, params, true)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetSubscriberTrend: failed")
		result = &repo.SubscriberTrendResult{}
	}
	result = s.fillLeadingSubscriberZeros(ctx, params, result)
	return mapSubscriberTrend(result, ""), nil
}

// GetDynamicSubscriberTrend returns daily or monthly subscriber trend based on the date range.
func (s *YoutubeAnalyticsService) GetDynamicSubscriberTrend(ctx context.Context, req *types.YoutubeRequest) (*types.SubscriberTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	daily := isYoutubeDailyGranularity(params)
	result, err := s.repo.GetSubscriberTrend(ctx, params, daily)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetDynamicSubscriberTrend: failed")
		result = &repo.SubscriberTrendResult{}
	}
	result = s.fillLeadingSubscriberZeros(ctx, params, result)
	level := aggregationLevelFromDaily(daily)
	return mapSubscriberTrend(result, level), nil
}

// GetEngagementTrend returns daily engagement trend data.
func (s *YoutubeAnalyticsService) GetEngagementTrend(ctx context.Context, req *types.YoutubeRequest) (*types.EngagementTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetEngagementTrend(ctx, params, true)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetEngagementTrend: failed")
		result = &repo.EngagementTrendResult{}
	}
	return mapEngagementTrend(result, ""), nil
}

// GetDynamicEngagementTrend returns daily or monthly engagement trend based on the date range.
func (s *YoutubeAnalyticsService) GetDynamicEngagementTrend(ctx context.Context, req *types.YoutubeRequest) (*types.EngagementTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	daily := isYoutubeDailyGranularity(params)
	result, err := s.repo.GetEngagementTrend(ctx, params, daily)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetDynamicEngagementTrend: failed")
		result = &repo.EngagementTrendResult{}
	}
	return mapEngagementTrend(result, aggregationLevelFromDaily(daily)), nil
}

// GetViewsTrend returns daily views trend data.
func (s *YoutubeAnalyticsService) GetViewsTrend(ctx context.Context, req *types.YoutubeRequest) (*types.ViewsTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetViewsTrend(ctx, params, true)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetViewsTrend: failed")
		result = &repo.ViewsTrendResult{}
	}
	return mapViewsTrend(result, ""), nil
}

// GetDynamicViewsTrend returns daily or monthly views trend based on the date range.
func (s *YoutubeAnalyticsService) GetDynamicViewsTrend(ctx context.Context, req *types.YoutubeRequest) (*types.ViewsTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	daily := isYoutubeDailyGranularity(params)
	result, err := s.repo.GetViewsTrend(ctx, params, daily)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetDynamicViewsTrend: failed")
		result = &repo.ViewsTrendResult{}
	}
	return mapViewsTrend(result, aggregationLevelFromDaily(daily)), nil
}

// GetWatchTimeTrend returns daily watch time trend data.
func (s *YoutubeAnalyticsService) GetWatchTimeTrend(ctx context.Context, req *types.YoutubeRequest) (*types.WatchTimeTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetWatchTimeTrend(ctx, params, true)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetWatchTimeTrend: failed")
		result = &repo.WatchTimeTrendResult{}
	}
	return mapWatchTimeTrend(result, ""), nil
}

// GetDynamicWatchTimeTrend returns daily or monthly watch time trend based on the date range.
func (s *YoutubeAnalyticsService) GetDynamicWatchTimeTrend(ctx context.Context, req *types.YoutubeRequest) (*types.WatchTimeTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	daily := isYoutubeDailyGranularity(params)
	result, err := s.repo.GetWatchTimeTrend(ctx, params, daily)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetDynamicWatchTimeTrend: failed")
		result = &repo.WatchTimeTrendResult{}
	}
	return mapWatchTimeTrend(result, aggregationLevelFromDaily(daily)), nil
}

// GetFindVideo returns traffic source breakdown data.
func (s *YoutubeAnalyticsService) GetFindVideo(ctx context.Context, req *types.YoutubeRequest) (*types.FindVideoResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetFindVideo(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetFindVideo: failed")
		return &types.FindVideoResponse{Status: true, Data: []types.TrafficSourceItem{}}, nil
	}
	items := make([]types.TrafficSourceItem, len(rows))
	for i, r := range rows {
		items[i] = types.TrafficSourceItem{Name: r.Name, Value: r.Value, PercValue: r.PercValue}
	}
	return &types.FindVideoResponse{Status: true, Data: items}, nil
}

// GetVideoSharing returns sharing platform breakdown data.
func (s *YoutubeAnalyticsService) GetVideoSharing(ctx context.Context, req *types.YoutubeRequest) (*types.VideoSharingResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetVideoSharing(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetVideoSharing: failed")
		return &types.VideoSharingResponse{Status: true, Data: []types.SharingItem{}}, nil
	}
	items := make([]types.SharingItem, len(rows))
	for i, r := range rows {
		items[i] = types.SharingItem{Name: r.Name, Value: r.Value, PercValue: r.PercValue}
	}
	return &types.VideoSharingResponse{Status: true, Data: items}, nil
}

// GetTopVideos fetches the top 5 videos by views and the top 5 by engagement concurrently.
func (s *YoutubeAnalyticsService) GetTopVideos(ctx context.Context, req *types.YoutubeRequest) (*types.TopVideosResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	var byViews, byEngagement []repo.VideoRow
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetTopVideos(egCtx, params, "views", 5, false)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetTopVideos: failed to get top by views")
		}
		byViews = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetTopVideos(egCtx, params, "engagement", 5, false)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetTopVideos: failed to get top by engagement")
		}
		byEngagement = r
		return nil
	})
	_ = eg.Wait()

	return &types.TopVideosResponse{
		Status:                      true,
		TopPostsOrderedByViews:      mapVideoRows(byViews, req.GetTimezone()),
		TopPostsOrderedByEngagement: mapVideoRows(byEngagement, req.GetTimezone()),
	}, nil
}

// GetLeastVideos fetches the bottom 5 videos by views and the bottom 5 by engagement concurrently.
func (s *YoutubeAnalyticsService) GetLeastVideos(ctx context.Context, req *types.YoutubeRequest) (*types.LeastVideosResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	var byViews, byEngagement []repo.VideoRow
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetTopVideos(egCtx, params, "views", 5, true)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetLeastVideos: failed to get least by views")
		}
		byViews = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetTopVideos(egCtx, params, "engagement", 5, true)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetLeastVideos: failed to get least by engagement")
		}
		byEngagement = r
		return nil
	})
	_ = eg.Wait()

	return &types.LeastVideosResponse{
		Status:                        true,
		LeastPostsOrderedByViews:      mapVideoRows(byViews, req.GetTimezone()),
		LeastPostsOrderedByEngagement: mapVideoRows(byEngagement, req.GetTimezone()),
	}, nil
}

// GetSortedTopVideos returns videos sorted by a configurable metric with a configurable limit.
func (s *YoutubeAnalyticsService) GetSortedTopVideos(ctx context.Context, req *types.TopVideosRequest) (*types.SortedTopVideosResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetTopVideos(ctx, params, req.GetOrderBy(), req.GetLimit(), false)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetSortedTopVideos: failed")
		return &types.SortedTopVideosResponse{Status: true, Data: []types.VideoItem{}}, nil
	}
	return &types.SortedTopVideosResponse{
		Status: true,
		Data:   mapVideoRows(rows, req.GetTimezone()),
	}, nil
}

// GetPerformanceAndSchedule fetches video engagement and view performance data concurrently.
func (s *YoutubeAnalyticsService) GetPerformanceAndSchedule(ctx context.Context, req *types.YoutubeRequest) (*types.PerformanceScheduleResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	var eng *repo.PerformanceEngagementResult
	var views *repo.PerformanceViewsResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		r, err := s.repo.GetPerformanceEngagement(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPerformanceAndSchedule: failed to get engagement")
			r = &repo.PerformanceEngagementResult{}
		}
		eng = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetPerformanceViews(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPerformanceAndSchedule: failed to get views")
			r = &repo.PerformanceViewsResult{}
		}
		views = r
		return nil
	})
	_ = eg.Wait()

	return &types.PerformanceScheduleResponse{
		Status:     true,
		Engagement: mapPerformanceEngagement(eng),
		VideoViews: mapPerformanceViews(views),
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

// isYoutubeDailyGranularity returns true when the date range is 180 days or fewer.
func isYoutubeDailyGranularity(params *clickhouse.QueryParams) bool {
	return params.DayCount <= 180
}

// aggregationLevelFromDaily maps the daily boolean to the aggregation level string.
func aggregationLevelFromDaily(daily bool) string {
	if daily {
		return "daily"
	}
	return "monthly"
}

// fillLeadingSubscriberZeros replaces leading zeros in SubscribersTotal with the latest
// known subscriber count fetched from youtube_channels before the period start.
func (s *YoutubeAnalyticsService) fillLeadingSubscriberZeros(ctx context.Context, params *clickhouse.QueryParams, result *repo.SubscriberTrendResult) *repo.SubscriberTrendResult {
	if len(result.SubscribersTotal) == 0 || result.SubscribersTotal[0] != 0 {
		return result
	}
	last, err := s.repo.GetLatestSubscriberCount(ctx, params)
	if err != nil || last == nil || last.SubscriberCount == 0 {
		return result
	}
	for i := range result.SubscribersTotal {
		if result.SubscribersTotal[i] != 0 {
			break
		}
		result.SubscribersTotal[i] = last.SubscriberCount
	}
	return result
}

// mapSummary combines activity, subscriber and video count results into SummaryMetrics.
func mapSummary(a *repo.ActivitySummaryResult, s *repo.SubscriberSummaryResult, v *repo.VideoCountResult) *types.SummaryMetrics {
	return &types.SummaryMetrics{
		WatchTime:       a.WatchTime,
		AvgViewDuration: sanitizeFloat(a.AvgViewDuration),
		Likes:           a.Likes,
		Dislikes:        a.Dislikes,
		Comments:        a.Comments,
		Shares:          a.Shares,
		Engagement:      a.Engagement,
		Views:           a.Views,
		Subscribers:     s.Subscribers,
		Videos:          v.VideoCount,
	}
}

// sanitizeFloat returns 0 for NaN or Inf values.
func sanitizeFloat(v float64) float64 {
	if math.IsNaN(v) || math.IsInf(v, 0) {
		return 0
	}
	return v
}

// pctChange computes the rounded percentage change from prev to curr.
func pctChange(curr, prev float64) float64 {
	if math.IsNaN(curr) || math.IsNaN(prev) || math.IsInf(curr, 0) || math.IsInf(prev, 0) || prev == 0 {
		return 0
	}
	return math.Round((curr-prev)/prev*100*100) / 100
}

// mapSummaryPercentage computes percentage change for each metric.
func mapSummaryPercentage(curr, prev *types.SummaryMetrics) *types.SummaryChangeMetrics {
	return &types.SummaryChangeMetrics{
		WatchTime:       pctChange(float64(curr.WatchTime), float64(prev.WatchTime)),
		AvgViewDuration: pctChange(curr.AvgViewDuration, prev.AvgViewDuration),
		Likes:           pctChange(float64(curr.Likes), float64(prev.Likes)),
		Dislikes:        pctChange(float64(curr.Dislikes), float64(prev.Dislikes)),
		Comments:        pctChange(float64(curr.Comments), float64(prev.Comments)),
		Shares:          pctChange(float64(curr.Shares), float64(prev.Shares)),
		Engagement:      pctChange(float64(curr.Engagement), float64(prev.Engagement)),
		Subscribers:     pctChange(float64(curr.Subscribers), float64(prev.Subscribers)),
		Views:           pctChange(float64(curr.Views), float64(prev.Views)),
		Videos:          pctChange(float64(curr.Videos), float64(prev.Videos)),
	}
}

// mapSummaryDifference computes the absolute difference for each metric.
func mapSummaryDifference(curr, prev *types.SummaryMetrics) *types.SummaryMetrics {
	return &types.SummaryMetrics{
		WatchTime:       curr.WatchTime - prev.WatchTime,
		AvgViewDuration: sanitizeFloat(curr.AvgViewDuration - prev.AvgViewDuration),
		Likes:           curr.Likes - prev.Likes,
		Dislikes:        curr.Dislikes - prev.Dislikes,
		Comments:        curr.Comments - prev.Comments,
		Shares:          curr.Shares - prev.Shares,
		Engagement:      curr.Engagement - prev.Engagement,
		Subscribers:     curr.Subscribers - prev.Subscribers,
		Views:           curr.Views - prev.Views,
		Videos:          curr.Videos - prev.Videos,
	}
}

// mapSubscriberTrend converts SubscriberTrendResult to the API response type.
func mapSubscriberTrend(r *repo.SubscriberTrendResult, level string) *types.SubscriberTrendResponse {
	buckets := bucketsToStrings(r.Buckets)
	return &types.SubscriberTrendResponse{
		Status:                 true,
		ShowData:               int32(r.ShowData),
		SubscribersGainedDaily: emptyInt32Slice(r.SubscribersGainedDaily),
		SubscribersTotal:       emptyInt32Slice(r.SubscribersTotal),
		Buckets:                buckets,
		AggregationLevel:       level,
	}
}

// mapEngagementTrend converts EngagementTrendResult to the API response type.
func mapEngagementTrend(r *repo.EngagementTrendResult, level string) *types.EngagementTrendResponse {
	buckets := bucketsToStrings(r.Buckets)
	return &types.EngagementTrendResponse{
		Status:           true,
		ShowData:         int32(r.ShowData),
		LikeDaily:        emptyInt32Slice(r.LikeDaily),
		LikeTotal:        emptyInt32Slice(r.LikeTotal),
		DislikeDaily:     emptyInt32Slice(r.DislikeDaily),
		DislikeTotal:     emptyInt32Slice(r.DislikeTotal),
		ShareDaily:       emptyInt32Slice(r.ShareDaily),
		ShareTotal:       emptyInt32Slice(r.ShareTotal),
		CommentDaily:     emptyInt32Slice(r.CommentDaily),
		CommentTotal:     emptyInt32Slice(r.CommentTotal),
		EngagementDaily:  emptyInt32Slice(r.EngagementDaily),
		EngagementTotal:  emptyInt32Slice(r.EngagementTotal),
		Buckets:          buckets,
		AggregationLevel: level,
	}
}

// mapViewsTrend converts ViewsTrendResult to the API response type.
func mapViewsTrend(r *repo.ViewsTrendResult, level string) *types.ViewsTrendResponse {
	buckets := bucketsToStrings(r.Buckets)
	return &types.ViewsTrendResponse{
		Status:                  true,
		ShowData:                int32(r.ShowData),
		SubscriberViewsDaily:    emptyInt32Slice(r.SubscriberViewsDaily),
		SubscriberViewsTotal:    emptyInt32Slice(r.SubscriberViewsTotal),
		NonSubscriberViewsDaily: emptyInt32Slice(r.NonSubscriberViewsDaily),
		NonSubscriberViewsTotal: emptyInt32Slice(r.NonSubscriberViewsTotal),
		VideoViewsDaily:         emptyInt32Slice(r.VideoViewsDaily),
		VideoViewsTotal:         emptyInt32Slice(r.VideoViewsTotal),
		Buckets:                 buckets,
		AggregationLevel:        level,
	}
}

// mapWatchTimeTrend converts WatchTimeTrendResult to the API response type.
func mapWatchTimeTrend(r *repo.WatchTimeTrendResult, level string) *types.WatchTimeTrendResponse {
	buckets := bucketsToStrings(r.Buckets)
	return &types.WatchTimeTrendResponse{
		Status:                      true,
		ShowData:                    int32(r.ShowData),
		SubscriberWatchTimeDaily:    emptyInt32Slice(r.SubscriberWatchTimeDaily),
		SubscriberWatchTimeTotal:    emptyInt32Slice(r.SubscriberWatchTimeTotal),
		NonSubscriberWatchTimeDaily: emptyInt32Slice(r.NonSubscriberWatchTimeDaily),
		NonSubscriberWatchTimeTotal: emptyInt32Slice(r.NonSubscriberWatchTimeTotal),
		AverageWatchTime:            emptyFloat64Slice(r.AverageWatchTime),
		Buckets:                     buckets,
		AggregationLevel:            level,
	}
}

// mapVideoRows converts a slice of VideoRow to the API VideoItem slice.
func mapVideoRows(rows []repo.VideoRow, timezone string) []types.VideoItem {
	if rows == nil {
		return []types.VideoItem{}
	}
	result := make([]types.VideoItem, len(rows))
	for i, r := range rows {
		result[i] = types.VideoItem{
			VideoID:           r.VideoID,
			Title:             r.Title,
			Description:       r.Description,
			Duration:          r.Duration,
			ThumbnailURL:      r.ThumbnailURL,
			MediaType:         r.MediaType,
			IframeEmbedURL:    r.IframeEmbedURL,
			ShareURL:          r.ShareURL,
			Engagement:        r.Engagement,
			Likes:             r.Likes,
			Dislikes:          r.Dislikes,
			Views:             r.Views,
			RedViews:          r.RedViews,
			Favorites:         r.Favorites,
			Comments:          r.Comments,
			SubscribersGained: r.SubscribersGained,
			Shares:            r.Shares,
			MinutesWatched:    r.MinutesWatched,
			RedMinutesWatched: r.RedMinutesWatched,
			AvgViewDuration:   r.AvgViewDuration,
			AvgViewPercentage: r.AvgViewPercentage,
			EngagementRate:    r.EngagementRate,
			PublishedAt:       formatTimeInTimezone(r.PublishedAt, timezone),
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

// mapPerformanceEngagement converts PerformanceEngagementResult to the API type.
func mapPerformanceEngagement(r *repo.PerformanceEngagementResult) *types.PerformanceEngagementData {
	buckets := bucketsToStrings(r.Buckets)
	return &types.PerformanceEngagementData{
		ShowData:   int32(r.ShowData),
		Buckets:    buckets,
		Count:      emptyInt32Slice(r.Count),
		Likes:      emptyInt32Slice(r.Likes),
		Dislikes:   emptyInt32Slice(r.Dislikes),
		Shares:     emptyInt32Slice(r.Shares),
		Comments:   emptyInt32Slice(r.Comments),
		Engagement: emptyInt32Slice(r.Engagement),
	}
}

// mapPerformanceViews converts PerformanceViewsResult to the API type.
func mapPerformanceViews(r *repo.PerformanceViewsResult) *types.PerformanceViewsData {
	buckets := bucketsToStrings(r.Buckets)
	return &types.PerformanceViewsData{
		ShowData:           int32(r.ShowData),
		Buckets:            buckets,
		Count:              emptyInt32Slice(r.Count),
		SubscriberViews:    emptyInt32Slice(r.SubscriberViews),
		NonSubscriberViews: emptyInt32Slice(r.NonSubscriberViews),
	}
}

// bucketsToStrings formats time.Time bucket slices to "YYYY-MM-DD" strings.
func bucketsToStrings(buckets []time.Time) []string {
	if buckets == nil {
		return []string{}
	}
	result := make([]string, len(buckets))
	for i, t := range buckets {
		result[i] = t.Format("2006-01-02")
	}
	return result
}

// emptyInt32Slice returns an initialised empty slice instead of nil so JSON output is [] not null.
func emptyInt32Slice(s []int32) []int32 {
	if s == nil {
		return []int32{}
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

// sumInt32 returns the sum of all values in the slice; used to detect empty time-series data.
func sumInt32(values []int32) int32 {
	var total int32
	for _, v := range values {
		total += v
	}
	return total
}
