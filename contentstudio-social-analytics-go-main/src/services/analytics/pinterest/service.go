// Package pinterest provides the business logic layer for Pinterest analytics.
// It orchestrates repository queries, handles concurrent data fetching via errgroup,
// computes previous-period comparisons, and maps ClickHouse results to API response types.
//
// Pinterest supports two modes controlled by PinterestRequest.HasBoard():
//   - User mode: queries pinterest_user_insights and pinterest_users
//   - Board mode: queries pinterest_pin_insights filtered by board_id
package pinterest

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/pinterest"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/pinterest"
)

// logIfErr logs err at error level unless it is a context cancellation/timeout,
// which are expected when the HTTP client disconnects before the query finishes.
func logIfErr(log zerolog.Logger, err error, msg string) {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return
	}
	log.Error().Err(err).Msg(msg)
}

// Service defines the interface for Pinterest analytics business logic.
type Service interface {
	GetSummary(ctx context.Context, req *types.PinterestRequest) (*types.SummaryResponse, error)
	GetFollowerTrend(ctx context.Context, req *types.PinterestRequest) (*types.FollowerTrendResponse, error)
	GetDynamicFollowerTrend(ctx context.Context, req *types.PinterestRequest) (*types.FollowerTrendResponse, error)
	GetImpressionsTrend(ctx context.Context, req *types.PinterestRequest) (*types.ImpressionsTrendResponse, error)
	GetDynamicImpressionsTrend(ctx context.Context, req *types.PinterestRequest) (*types.ImpressionsTrendResponse, error)
	GetEngagementTrend(ctx context.Context, req *types.PinterestRequest) (*types.EngagementTrendResponse, error)
	GetDynamicEngagementTrend(ctx context.Context, req *types.PinterestRequest) (*types.EngagementTrendResponse, error)
	GetPinPosting(ctx context.Context, req *types.FilteredPinRequest) (*types.PinPostingResponse, error)
	GetDynamicPinPosting(ctx context.Context, req *types.FilteredPinRequest) (*types.PinPostingResponse, error)
	GetPinRollup(ctx context.Context, req *types.PinterestRequest) (*types.PinRollupResponse, error)
	GetTopPins(ctx context.Context, req *types.TopPinsRequest) (*types.TopPinsResponse, error)
	GetPinPerformance(ctx context.Context, req *types.PinterestRequest) (*types.PinPerformanceResponse, error)
}

// PinterestAnalyticsService implements Pinterest analytics business logic.
type PinterestAnalyticsService struct {
	repo   *repo.Repository
	logger zerolog.Logger
}

var _ Service = (*PinterestAnalyticsService)(nil)

// NewPinterestAnalyticsService creates a new service with the given repository and logger.
func NewPinterestAnalyticsService(r *repo.Repository, logger zerolog.Logger) *PinterestAnalyticsService {
	return &PinterestAnalyticsService{
		repo:   r,
		logger: logger.With().Str("service", "pinterest-analytics").Logger(),
	}
}

// GetSummary fetches current and previous period summary metrics concurrently and returns them
// under "current" and "previous" keys for period-over-period comparison.
func (s *PinterestAnalyticsService) GetSummary(ctx context.Context, req *types.PinterestRequest) (*types.SummaryResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)
	boardIDs := req.FormatBoardIDs()
	hasBoard := req.HasBoard()

	var curr, prev *repo.SummaryResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var r *repo.SummaryResult
		var e error
		if hasBoard {
			r, e = s.repo.GetSummaryForBoard(egCtx, params, boardIDs)
		} else {
			r, e = s.repo.GetSummaryForUser(egCtx, params)
		}
		if e != nil {
			logIfErr(s.logger, e, "GetSummary: failed to get current summary")
			r = &repo.SummaryResult{}
		}
		curr = r
		return nil
	})
	eg.Go(func() error {
		var r *repo.SummaryResult
		var e error
		if hasBoard {
			r, e = s.repo.GetSummaryForBoard(egCtx, prevParams, boardIDs)
		} else {
			r, e = s.repo.GetSummaryForUser(egCtx, prevParams)
		}
		if e != nil {
			logIfErr(s.logger, e, "GetSummary: failed to get previous summary")
			r = &repo.SummaryResult{}
		}
		prev = r
		return nil
	})
	_ = eg.Wait()

	currMetrics := mapSummary(curr)
	prevMetrics := mapSummary(prev)
	return &types.SummaryResponse{
		Status: true,
		Overview: &types.SummaryOverview{
			Current:    currMetrics,
			Previous:   prevMetrics,
			Percentage: mapSummaryPercentage(currMetrics, prevMetrics),
			Difference: mapSummaryDifference(currMetrics, prevMetrics),
		},
	}, nil
}

// GetFollowerTrend returns daily follower trend data for the requested period.
func (s *PinterestAnalyticsService) GetFollowerTrend(ctx context.Context, req *types.PinterestRequest) (*types.FollowerTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	return s.followerTrend(ctx, req, params, true, "")
}

// GetDynamicFollowerTrend returns follower trend data with granularity auto-selected based on date range.
// Ranges of 180 days or fewer use daily buckets; longer ranges use monthly buckets.
func (s *PinterestAnalyticsService) GetDynamicFollowerTrend(ctx context.Context, req *types.PinterestRequest) (*types.FollowerTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	daily := isPinterestDailyGranularity(params)
	aggLevel := aggregationLevelFromDaily(daily)
	return s.followerTrend(ctx, req, params, daily, aggLevel)
}

func (s *PinterestAnalyticsService) followerTrend(ctx context.Context, req *types.PinterestRequest, params *clickhouse.QueryParams, daily bool, aggLevel string) (*types.FollowerTrendResponse, error) {
	boardIDs := req.FormatBoardIDs()
	var result *repo.FollowerTrendResult
	var err error
	if req.HasBoard() {
		result, err = s.repo.GetFollowerTrendForBoard(ctx, params, boardIDs, daily)
	} else {
		result, err = s.repo.GetFollowerTrendForUser(ctx, params, daily)
	}
	if err != nil {
		logIfErr(s.logger, err, "GetFollowerTrend: failed to get data")
		result = &repo.FollowerTrendResult{}
	}
	return mapFollowerTrend(result, aggLevel), nil
}

// GetImpressionsTrend returns daily impressions trend data for the requested period.
func (s *PinterestAnalyticsService) GetImpressionsTrend(ctx context.Context, req *types.PinterestRequest) (*types.ImpressionsTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	return s.impressionsTrend(ctx, req, params, true, "")
}

// GetDynamicImpressionsTrend returns impressions trend data with auto-selected granularity.
func (s *PinterestAnalyticsService) GetDynamicImpressionsTrend(ctx context.Context, req *types.PinterestRequest) (*types.ImpressionsTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	daily := isPinterestDailyGranularity(params)
	aggLevel := aggregationLevelFromDaily(daily)
	return s.impressionsTrend(ctx, req, params, daily, aggLevel)
}

func (s *PinterestAnalyticsService) impressionsTrend(ctx context.Context, req *types.PinterestRequest, params *clickhouse.QueryParams, daily bool, aggLevel string) (*types.ImpressionsTrendResponse, error) {
	boardIDs := req.FormatBoardIDs()
	var result *repo.ImpressionsTrendResult
	var err error
	if req.HasBoard() {
		result, err = s.repo.GetImpressionsTrendForBoard(ctx, params, boardIDs, daily)
	} else {
		result, err = s.repo.GetImpressionsTrendForUser(ctx, params, daily)
	}
	if err != nil {
		logIfErr(s.logger, err, "GetImpressionsTrend: failed to get data")
		result = &repo.ImpressionsTrendResult{}
	}
	return mapImpressionsTrend(result, aggLevel), nil
}

// GetEngagementTrend returns daily engagement trend data for the requested period.
func (s *PinterestAnalyticsService) GetEngagementTrend(ctx context.Context, req *types.PinterestRequest) (*types.EngagementTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	return s.engagementTrend(ctx, req, params, true, "")
}

// GetDynamicEngagementTrend returns engagement trend data with auto-selected granularity.
func (s *PinterestAnalyticsService) GetDynamicEngagementTrend(ctx context.Context, req *types.PinterestRequest) (*types.EngagementTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	daily := isPinterestDailyGranularity(params)
	aggLevel := aggregationLevelFromDaily(daily)
	return s.engagementTrend(ctx, req, params, daily, aggLevel)
}

func (s *PinterestAnalyticsService) engagementTrend(ctx context.Context, req *types.PinterestRequest, params *clickhouse.QueryParams, daily bool, aggLevel string) (*types.EngagementTrendResponse, error) {
	boardIDs := req.FormatBoardIDs()
	var result *repo.EngagementTrendResult
	var err error
	if req.HasBoard() {
		result, err = s.repo.GetEngagementTrendForBoard(ctx, params, boardIDs, daily)
	} else {
		result, err = s.repo.GetEngagementTrendForUser(ctx, params, daily)
	}
	if err != nil {
		logIfErr(s.logger, err, "GetEngagementTrend: failed to get data")
		result = &repo.EngagementTrendResult{}
	}
	return mapEngagementTrend(result, aggLevel), nil
}

// GetPinPosting returns daily pin posting frequency data, optionally filtered by media type.
func (s *PinterestAnalyticsService) GetPinPosting(ctx context.Context, req *types.FilteredPinRequest) (*types.PinPostingResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	return s.pinPosting(ctx, req, params, true, "")
}

// GetDynamicPinPosting returns pin posting frequency data with auto-selected granularity.
func (s *PinterestAnalyticsService) GetDynamicPinPosting(ctx context.Context, req *types.FilteredPinRequest) (*types.PinPostingResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	daily := isPinterestDailyGranularity(params)
	aggLevel := aggregationLevelFromDaily(daily)
	return s.pinPosting(ctx, req, params, daily, aggLevel)
}

func (s *PinterestAnalyticsService) pinPosting(ctx context.Context, req *types.FilteredPinRequest, params *clickhouse.QueryParams, daily bool, aggLevel string) (*types.PinPostingResponse, error) {
	boardIDs := req.FormatBoardIDs()
	var result *repo.PinPostingResult
	var err error
	if req.HasBoard() {
		result, err = s.repo.GetPinPostingForBoard(ctx, params, boardIDs, req.FilterBy, daily)
	} else {
		result, err = s.repo.GetPinPostingForUser(ctx, params, req.FilterBy, daily)
	}
	if err != nil {
		logIfErr(s.logger, err, "GetPinPosting: failed to get data")
		result = &repo.PinPostingResult{}
	}
	return mapPinPosting(result, aggLevel), nil
}

// GetPinRollup fetches current and previous period pin rollup metrics concurrently.
func (s *PinterestAnalyticsService) GetPinRollup(ctx context.Context, req *types.PinterestRequest) (*types.PinRollupResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := prevPeriodParams(params)
	boardIDs := req.FormatBoardIDs()
	hasBoard := req.HasBoard()

	var curr, prev *repo.PinRollupResult
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var r *repo.PinRollupResult
		var e error
		if hasBoard {
			r, e = s.repo.GetPinRollupForBoard(egCtx, params, boardIDs)
		} else {
			r, e = s.repo.GetPinRollupForUser(egCtx, params)
		}
		if e != nil {
			logIfErr(s.logger, e, "GetPinRollup: failed to get current rollup")
			r = &repo.PinRollupResult{}
		}
		curr = r
		return nil
	})
	eg.Go(func() error {
		var r *repo.PinRollupResult
		var e error
		if hasBoard {
			r, e = s.repo.GetPinRollupForBoard(egCtx, prevParams, boardIDs)
		} else {
			r, e = s.repo.GetPinRollupForUser(egCtx, prevParams)
		}
		if e != nil {
			logIfErr(s.logger, e, "GetPinRollup: failed to get previous rollup")
			r = &repo.PinRollupResult{}
		}
		prev = r
		return nil
	})
	_ = eg.Wait()

	currRollup := mapPinRollup(curr)
	prevRollup := mapPinRollup(prev)
	return &types.PinRollupResponse{
		Status: true,
		Overview: &types.PinRollupOverview{
			Current:    currRollup,
			Previous:   prevRollup,
			Percentage: mapPinRollupPercentage(currRollup, prevRollup),
			Difference: mapPinRollupDifference(currRollup, prevRollup),
		},
	}, nil
}

// GetTopPins fetches the top and least performing pins concurrently sorted by the requested metric.
// Default orderBy is "impressions"; default limit is 5.
func (s *PinterestAnalyticsService) GetTopPins(ctx context.Context, req *types.TopPinsRequest) (*types.TopPinsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	boardIDs := req.FormatBoardIDs()
	hasBoard := req.HasBoard()

	orderBy := req.OrderBy
	if orderBy == "" {
		orderBy = "impressions"
	}
	limit := req.Limit
	if limit <= 0 {
		limit = 5
	}

	var topRows, leastRows []repo.PinRow
	eg, egCtx := errgroup.WithContext(ctx)

	eg.Go(func() error {
		var rows []repo.PinRow
		var e error
		if hasBoard {
			rows, e = s.repo.GetPinsForBoard(egCtx, params, boardIDs, orderBy, limit, false)
		} else {
			rows, e = s.repo.GetPinsForUser(egCtx, params, orderBy, limit, false)
		}
		if e != nil {
			logIfErr(s.logger, e, "GetTopPins: failed to get top pins")
		}
		topRows = rows
		return nil
	})
	eg.Go(func() error {
		var rows []repo.PinRow
		var e error
		if hasBoard {
			rows, e = s.repo.GetPinsForBoard(egCtx, params, boardIDs, orderBy, limit, true)
		} else {
			rows, e = s.repo.GetPinsForUser(egCtx, params, orderBy, limit, true)
		}
		if e != nil {
			logIfErr(s.logger, e, "GetTopPins: failed to get least pins")
		}
		leastRows = rows
		return nil
	})
	_ = eg.Wait()

	return &types.TopPinsResponse{
		Status: true,
		Top:    mapPinRows(topRows, req.GetTimezone()),
		Least:  mapPinRows(leastRows, req.GetTimezone()),
	}, nil
}

// GetPinPerformance returns daily time-series performance metrics grouped by pin creation date.
func (s *PinterestAnalyticsService) GetPinPerformance(ctx context.Context, req *types.PinterestRequest) (*types.PinPerformanceResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	boardIDs := req.FormatBoardIDs()

	var result *repo.PinPerformanceResult
	var fetchErr error
	if req.HasBoard() {
		result, fetchErr = s.repo.GetPinPerformanceForBoard(ctx, params, boardIDs)
	} else {
		result, fetchErr = s.repo.GetPinPerformanceForUser(ctx, params)
	}
	if fetchErr != nil {
		logIfErr(s.logger, fetchErr, "GetPinPerformance: failed to get data")
		result = &repo.PinPerformanceResult{}
	}
	return mapPinPerformance(result), nil
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

// isPinterestDailyGranularity returns true when the period is 180 days or fewer,
// triggering daily bucket aggregation instead of monthly.
func isPinterestDailyGranularity(params *clickhouse.QueryParams) bool {
	return params.DayCount <= 180
}

// aggregationLevelFromDaily returns the string aggregation level for the response.
func aggregationLevelFromDaily(daily bool) string {
	if daily {
		return "daily"
	}
	return "monthly"
}

// pctChange returns the rounded percentage change from prev to curr. Returns 0 if prev is 0.
func pctChange(curr, prev float64) float64 {
	if prev == 0 {
		return 0
	}
	return math.Round((curr-prev)/prev*100*100) / 100
}

// mapSummary converts a SummaryResult to the API SummaryMetrics type.
func mapSummary(r *repo.SummaryResult) *types.SummaryMetrics {
	return &types.SummaryMetrics{
		FollowerCount:   r.FollowerCount,
		Impressions:     r.Impressions,
		PinClicks:       r.PinClicks,
		OutboundClicks:  r.OutboundClicks,
		Saves:           r.Saves,
		TotalEngagement: r.TotalEngagement,
	}
}

// mapSummaryPercentage computes percentage changes between current and previous SummaryMetrics.
func mapSummaryPercentage(curr, prev *types.SummaryMetrics) *types.SummaryChangeMetrics {
	return &types.SummaryChangeMetrics{
		FollowerCount:   pctChange(float64(curr.FollowerCount), float64(prev.FollowerCount)),
		Impressions:     pctChange(float64(curr.Impressions), float64(prev.Impressions)),
		PinClicks:       pctChange(float64(curr.PinClicks), float64(prev.PinClicks)),
		OutboundClicks:  pctChange(float64(curr.OutboundClicks), float64(prev.OutboundClicks)),
		Saves:           pctChange(float64(curr.Saves), float64(prev.Saves)),
		TotalEngagement: pctChange(float64(curr.TotalEngagement), float64(prev.TotalEngagement)),
	}
}

// mapSummaryDifference computes absolute differences between current and previous SummaryMetrics.
func mapSummaryDifference(curr, prev *types.SummaryMetrics) *types.SummaryMetrics {
	return &types.SummaryMetrics{
		FollowerCount:   curr.FollowerCount - prev.FollowerCount,
		Impressions:     curr.Impressions - prev.Impressions,
		PinClicks:       curr.PinClicks - prev.PinClicks,
		OutboundClicks:  curr.OutboundClicks - prev.OutboundClicks,
		Saves:           curr.Saves - prev.Saves,
		TotalEngagement: curr.TotalEngagement - prev.TotalEngagement,
	}
}

// mapFollowerTrend converts a FollowerTrendResult to the API response type,
// formatting bucket timestamps to ISO-8601 date strings.
func mapFollowerTrend(r *repo.FollowerTrendResult, aggLevel string) *types.FollowerTrendResponse {
	buckets := makeBuckets(r.Buckets)
	resp := &types.FollowerTrendResponse{
		Status:          true,
		ShowData:        int32(r.ShowData),
		FollowersDaily:  emptyInt32Slice(r.FollowersDaily),
		FollowersGained: emptyInt32Slice(r.FollowersGained),
		Buckets:         buckets,
	}
	if aggLevel != "" {
		resp.AggregationLevel = aggLevel
	}
	return resp
}

// mapImpressionsTrend converts an ImpressionsTrendResult to the API response type.
func mapImpressionsTrend(r *repo.ImpressionsTrendResult, aggLevel string) *types.ImpressionsTrendResponse {
	buckets := makeBuckets(r.Buckets)
	resp := &types.ImpressionsTrendResponse{
		Status:           true,
		ShowData:         int32(r.ShowData),
		ImpressionsDaily: emptyInt32Slice(r.ImpressionsDaily),
		ImpressionsTotal: emptyInt32Slice(r.ImpressionsTotal),
		Buckets:          buckets,
	}
	if aggLevel != "" {
		resp.AggregationLevel = aggLevel
	}
	return resp
}

// mapEngagementTrend converts an EngagementTrendResult to the API response type.
func mapEngagementTrend(r *repo.EngagementTrendResult, aggLevel string) *types.EngagementTrendResponse {
	buckets := makeBuckets(r.Buckets)
	resp := &types.EngagementTrendResponse{
		Status:              true,
		ShowData:            int32(r.ShowData),
		SavesDaily:          emptyInt32Slice(r.SavesDaily),
		SavesTotal:          emptyInt32Slice(r.SavesTotal),
		OutboundClicksDaily: emptyInt32Slice(r.OutboundClicksDaily),
		OutboundClicksTotal: emptyInt32Slice(r.OutboundClicksTotal),
		PinClicksDaily:      emptyInt32Slice(r.PinClicksDaily),
		PinClicksTotal:      emptyInt32Slice(r.PinClicksTotal),
		EngagementDaily:     emptyInt32Slice(r.EngagementDaily),
		EngagementTotal:     emptyInt32Slice(r.EngagementTotal),
		Buckets:             buckets,
	}
	if aggLevel != "" {
		resp.AggregationLevel = aggLevel
	}
	return resp
}

// mapPinPosting converts a PinPostingResult to the API response type.
func mapPinPosting(r *repo.PinPostingResult, aggLevel string) *types.PinPostingResponse {
	buckets := makeBuckets(r.Buckets)
	resp := &types.PinPostingResponse{
		Status:    true,
		ShowData:  int32(r.ShowData),
		PinsCount: emptyInt32Slice(r.PinsCount),
		Buckets:   buckets,
	}
	if aggLevel != "" {
		resp.AggregationLevel = aggLevel
	}
	return resp
}

// mapPinRollup converts a PinRollupResult to the API PinRollupMetrics type.
func mapPinRollup(r *repo.PinRollupResult) *types.PinRollupMetrics {
	return &types.PinRollupMetrics{
		TotalPins:        r.TotalPins,
		Impressions:      r.Impressions,
		PinClicks:        r.PinClicks,
		OutboundClicks:   r.OutboundClicks,
		Saves:            r.Saves,
		QuartilePercView: r.QuartilePercView,
		VideoViews:       r.VideoViews,
		Video10sViews:    r.Video10sViews,
		AvgWatchTime:     r.AvgWatchTime,
	}
}

// mapPinRollupPercentage computes percentage changes between current and previous PinRollupMetrics.
func mapPinRollupPercentage(curr, prev *types.PinRollupMetrics) *types.PinRollupChangeMetrics {
	return &types.PinRollupChangeMetrics{
		TotalPins:        pctChange(float64(curr.TotalPins), float64(prev.TotalPins)),
		Impressions:      pctChange(float64(curr.Impressions), float64(prev.Impressions)),
		PinClicks:        pctChange(float64(curr.PinClicks), float64(prev.PinClicks)),
		OutboundClicks:   pctChange(float64(curr.OutboundClicks), float64(prev.OutboundClicks)),
		Saves:            pctChange(float64(curr.Saves), float64(prev.Saves)),
		QuartilePercView: pctChange(curr.QuartilePercView, prev.QuartilePercView),
		VideoViews:       pctChange(float64(curr.VideoViews), float64(prev.VideoViews)),
		Video10sViews:    pctChange(float64(curr.Video10sViews), float64(prev.Video10sViews)),
		AvgWatchTime:     pctChange(curr.AvgWatchTime, prev.AvgWatchTime),
	}
}

// mapPinRollupDifference computes absolute differences between current and previous PinRollupMetrics.
func mapPinRollupDifference(curr, prev *types.PinRollupMetrics) *types.PinRollupMetrics {
	return &types.PinRollupMetrics{
		TotalPins:        curr.TotalPins - prev.TotalPins,
		Impressions:      curr.Impressions - prev.Impressions,
		PinClicks:        curr.PinClicks - prev.PinClicks,
		OutboundClicks:   curr.OutboundClicks - prev.OutboundClicks,
		Saves:            curr.Saves - prev.Saves,
		QuartilePercView: curr.QuartilePercView - prev.QuartilePercView,
		VideoViews:       curr.VideoViews - prev.VideoViews,
		Video10sViews:    curr.Video10sViews - prev.Video10sViews,
		AvgWatchTime:     curr.AvgWatchTime - prev.AvgWatchTime,
	}
}

// mapPinRows converts a slice of PinRow to the API PinItem slice type.
func mapPinRows(rows []repo.PinRow, timezone string) []types.PinItem {
	if rows == nil {
		return []types.PinItem{}
	}
	result := make([]types.PinItem, len(rows))
	for i, r := range rows {
		result[i] = types.PinItem{
			PinID:           r.PinID,
			BoardName:       r.BoardName,
			Permalink:       r.Permalink,
			EmbedLink:       r.EmbedLink,
			Title:           r.Title,
			Description:     r.Description,
			BoardOwner:      r.BoardOwner,
			MediaType:       r.MediaType,
			CoverImageURL:   r.CoverImageURL,
			DominantColor:   r.DominantColor,
			CreativeType:    r.CreativeType,
			ProductTags:     r.ProductTags,
			Height:          r.Height,
			Width:           r.Width,
			CreatedAt:       formatTimeInTimezone(r.CreatedAt, timezone),
			Impressions:     r.Impressions,
			PinClicks:       r.PinClicks,
			OutboundClicks:  r.OutboundClicks,
			Saves:           r.Saves,
			TotalEngagement: r.TotalEngagement,
			EngagementRate:  r.EngagementRate,
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

// mapPinPerformance converts a PinPerformanceResult to the API response type.
func mapPinPerformance(r *repo.PinPerformanceResult) *types.PinPerformanceResponse {
	buckets := makeBuckets(r.Buckets)
	return &types.PinPerformanceResponse{
		Status:         true,
		ShowData:       int32(r.ShowData),
		PinsCount:      emptyInt32Slice(r.PinsCount),
		PinClicks:      emptyInt32Slice(r.PinClicks),
		OutboundClicks: emptyInt32Slice(r.OutboundClicks),
		Saves:          emptyInt32Slice(r.Saves),
		Engagements:    emptyInt32Slice(r.Engagements),
		Impressions:    emptyInt32Slice(r.Impressions),
		Buckets:        buckets,
	}
}

// makeBuckets formats a slice of time.Time bucket values to ISO-8601 date strings.
func makeBuckets(ts []time.Time) []string {
	buckets := make([]string, len(ts))
	for i, t := range ts {
		buckets[i] = t.Format("2006-01-02")
	}
	return buckets
}

// emptyInt32Slice returns an initialised empty slice instead of nil so JSON output is [] not null.
func emptyInt32Slice(s []int32) []int32 {
	if s == nil {
		return []int32{}
	}
	return s
}
