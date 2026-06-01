// Package meta_ads implements the business logic for Meta Ads analytics.
// It orchestrates concurrent ClickHouse queries and maps raw results to API responses.
package meta_ads

import (
	"context"
	"fmt"
	"math"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/meta_ads"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/meta_ads"
)

// Repository defines the data access methods used by the service.
type Repository interface {
	GetSummary(ctx context.Context, params *ch.QueryParams) (*repo.SummaryResult, error)
	GetResultsByObjective(ctx context.Context, params *ch.QueryParams) ([]repo.ObjectiveResultRow, error)
	GetResultsByObjectivePrevBatch(ctx context.Context, params *ch.QueryParams) (map[string]int64, error)
	GetDailyMetrics(ctx context.Context, params *ch.QueryParams) ([]repo.DailyMetricsRow, error)
	GetTopCampaigns(ctx context.Context, params *ch.QueryParams, sortBy string) ([]repo.TopCampaignRow, error)
	GetPerformanceTrend(ctx context.Context, params *ch.QueryParams, metric string) ([]repo.TrendRow, error)
	GetPerformanceByCampaign(ctx context.Context, params *ch.QueryParams, metric string) ([]repo.LevelBreakdownRow, error)
	GetPerformanceByAdSet(ctx context.Context, params *ch.QueryParams, metric string) ([]repo.LevelBreakdownRow, error)
	GetPerformanceByAd(ctx context.Context, params *ch.QueryParams, metric string) ([]repo.LevelBreakdownRow, error)
	GetPerformanceByPlatform(ctx context.Context, params *ch.QueryParams, metric string) ([]repo.PlatformBreakdownRow, error)
	GetCampaignsList(ctx context.Context, params *ch.QueryParams, status, objective, search, orderBy, orderDir string, page, perPage int) ([]repo.CampaignTableRow, int64, error)
	GetCampaignFilterOptions(ctx context.Context, params *ch.QueryParams) ([]string, []string, error)
	GetAdSetsList(ctx context.Context, params *ch.QueryParams, status, objective, search, orderBy, orderDir string, page, perPage int) ([]repo.AdSetTableRow, int64, error)
	GetAdSetFilterOptions(ctx context.Context, params *ch.QueryParams) ([]string, error)
	GetAdsList(ctx context.Context, params *ch.QueryParams, status, objective, search, orderBy, orderDir string, page, perPage int) ([]repo.AdTableRow, int64, error)
	GetAdFilterOptions(ctx context.Context, params *ch.QueryParams) ([]string, error)
	GetDemographicsAgeGender(ctx context.Context, params *ch.QueryParams) ([]repo.AgeGenderRow, error)
	GetDemographicsRegionCountry(ctx context.Context, params *ch.QueryParams, breakdown, countryFilter, orderBy, orderDir string) ([]repo.RegionCountryRow, error)
	GetAvailableCountries(ctx context.Context, params *ch.QueryParams) ([]string, error)
	GetAccountCurrency(ctx context.Context, accountID string) (string, error)
}

// Service defines the business logic interface for Meta Ads analytics.
type Service interface {
	GetSummary(ctx context.Context, req *types.MetaAdsRequest) (*types.SummaryResponse, error)
	GetResultsByObjective(ctx context.Context, req *types.MetaAdsRequest) (*types.ResultsByObjectiveResponse, error)
	GetImpressionsVsSpend(ctx context.Context, req *types.MetaAdsRequest) (*types.ImpressionsVsSpendResponse, error)
	GetClicksVsCTR(ctx context.Context, req *types.MetaAdsRequest) (*types.ClicksVsCTRResponse, error)
	GetTopCampaigns(ctx context.Context, req *types.MetaAdsRequest) (*types.TopCampaignsResponse, error)
	GetPerformanceTrend(ctx context.Context, req *types.MetaAdsRequest) (*types.PerformanceTrendResponse, error)
	GetPerformanceByLevel(ctx context.Context, req *types.MetaAdsRequest) (*types.PerformanceByLevelResponse, error)
	GetPerformanceByPlatform(ctx context.Context, req *types.MetaAdsRequest) (*types.PerformanceByPlatformResponse, error)
	GetCampaignsList(ctx context.Context, req *types.MetaAdsRequest) (*types.TableResponse, error)
	GetAdSetsList(ctx context.Context, req *types.MetaAdsRequest) (*types.TableResponse, error)
	GetAdsList(ctx context.Context, req *types.MetaAdsRequest) (*types.TableResponse, error)
	GetDemographicsAgeGender(ctx context.Context, req *types.MetaAdsRequest) (*types.DemographicsAgeGenderResponse, error)
	GetDemographicsRegionCountry(ctx context.Context, req *types.MetaAdsRequest) (*types.DemographicsRegionCountryResponse, error)
	// AI Insights
	GetAIInsightsSummary(ctx context.Context, req *types.MetaAdsRequest) (map[string]interface{}, error)
	GetAIInsightsDetailed(ctx context.Context, req *types.MetaAdsRequest) (map[string]interface{}, error)
}

// MetaAdsService implements Service.
type MetaAdsService struct {
	repo      Repository
	logger    zerolog.Logger
	aiService *AIInsightsService
}

// NewMetaAdsService returns a new MetaAdsService.
func NewMetaAdsService(r *repo.Repository, logger zerolog.Logger) *MetaAdsService {
	return &MetaAdsService{
		repo:   r,
		logger: logger.With().Str("service", "meta-ads-analytics").Logger(),
	}
}

// SetAIInsightsService sets the AI insights dependency.
func (s *MetaAdsService) SetAIInsightsService(aiSvc *AIInsightsService) {
	s.aiService = aiSvc
}

// GetAccountCurrency returns the stored currency for the requested Meta Ads account.
func (s *MetaAdsService) GetAccountCurrency(ctx context.Context, req *types.MetaAdsRequest) (string, error) {
	if s.repo == nil || req == nil || req.AccountID == "" {
		return "USD", nil
	}

	currency, err := s.repo.GetAccountCurrency(ctx, req.AccountID)
	if err != nil {
		return "", err
	}
	if currency == "" {
		return "USD", nil
	}
	return currency, nil
}

// ─────────────────────────────────────────────
// Helper
// ─────────────────────────────────────────────

func percentChange(current, previous float64) float64 {
	if previous == 0 {
		if current == 0 {
			return 0
		}
		return 100
	}
	v := (current - previous) / math.Abs(previous) * 100
	return math.Round(v*100) / 100
}

func metricValue(current, previous float64) types.MetricValue {
	return types.MetricValue{
		Current:  current,
		Previous: previous,
		Change:   percentChange(current, previous),
	}
}

func safeRate(num, denom float64) float64 {
	if denom == 0 {
		return 0
	}
	return num / denom
}

// objectiveLabel maps internal objective keys to display labels.
var objectiveLabel = map[string]string{
	"OUTCOME_AWARENESS":     "Awareness",
	"OUTCOME_TRAFFIC":       "Traffic",
	"OUTCOME_ENGAGEMENT":    "Engagement",
	"OUTCOME_LEADS":         "Leads",
	"OUTCOME_APP_PROMOTION": "App Promotion",
	"OUTCOME_SALES":         "Sales",
}

// ─────────────────────────────────────────────
// GetSummary
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetSummary(ctx context.Context, req *types.MetaAdsRequest) (*types.SummaryResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prev := *params
	prev.DateFrom = params.PrevDateFrom
	prev.DateTo = params.PrevDateTo

	var curr, prevResult *repo.SummaryResult
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var e error
		curr, e = s.repo.GetSummary(ctx, params)
		return e
	})
	eg.Go(func() error {
		var e error
		prevResult, e = s.repo.GetSummary(ctx, &prev)
		return e
	})
	if err := eg.Wait(); err != nil {
		s.logger.Error().Err(err).Msg("GetSummary failed")
		return nil, err
	}

	currCPM := safeRate(float64(curr.Spend)*1000, float64(curr.Impressions))
	prevCPM := safeRate(float64(prevResult.Spend)*1000, float64(prevResult.Impressions))
	currCPC := safeRate(curr.Spend, float64(curr.Clicks))
	prevCPC := safeRate(prevResult.Spend, float64(prevResult.Clicks))
	currCTR := safeRate(float64(curr.Clicks)*100, float64(curr.Impressions))
	prevCTR := safeRate(float64(prevResult.Clicks)*100, float64(prevResult.Impressions))

	return &types.SummaryResponse{
		Status:      true,
		Spend:       metricValue(curr.Spend, prevResult.Spend),
		Reach:       metricValue(float64(curr.Reach), float64(prevResult.Reach)),
		Impressions: metricValue(float64(curr.Impressions), float64(prevResult.Impressions)),
		Clicks:      metricValue(float64(curr.Clicks), float64(prevResult.Clicks)),
		CPM:         metricValue(currCPM, prevCPM),
		CPC:         metricValue(currCPC, prevCPC),
		CTR:         metricValue(currCTR, prevCTR),
	}, nil
}

// ─────────────────────────────────────────────
// GetResultsByObjective
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetResultsByObjective(ctx context.Context, req *types.MetaAdsRequest) (*types.ResultsByObjectiveResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prev := *params
	prev.DateFrom = params.PrevDateFrom
	prev.DateTo = params.PrevDateTo

	var curr []repo.ObjectiveResultRow
	var prevMap map[string]int64
	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var e error
		curr, e = s.repo.GetResultsByObjective(ctx, params)
		return e
	})
	eg.Go(func() error {
		var e error
		prevMap, e = s.repo.GetResultsByObjectivePrevBatch(ctx, &prev)
		return e
	})
	if err := eg.Wait(); err != nil {
		s.logger.Error().Err(err).Msg("GetResultsByObjective failed")
		return nil, err
	}

	result := make([]types.ObjectiveResult, 0, len(curr))
	for _, row := range curr {
		prevResults := prevMap[row.Objective]
		costPerResult := safeRate(row.Spend, float64(row.Results))
		result = append(result, types.ObjectiveResult{
			Objective:     row.Objective,
			Label:         objectiveLabel[row.Objective],
			Results:       row.Results,
			ResultsPrev:   prevResults,
			ResultsChange: percentChange(float64(row.Results), float64(prevResults)),
			CostPerResult: math.Round(costPerResult*100) / 100,
			CampaignCount: row.CampaignCount,
		})
	}

	return &types.ResultsByObjectiveResponse{Status: true, Data: result}, nil
}

// ─────────────────────────────────────────────
// GetImpressionsVsSpend
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetImpressionsVsSpend(ctx context.Context, req *types.MetaAdsRequest) (*types.ImpressionsVsSpendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetDailyMetrics(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetImpressionsVsSpend failed")
		return nil, err
	}

	dates := make([]string, len(rows))
	impressions := make([]float64, len(rows))
	spend := make([]float64, len(rows))
	for i, r := range rows {
		dates[i] = r.InsightsDate
		impressions[i] = float64(r.Impressions)
		spend[i] = r.Spend
	}
	return &types.ImpressionsVsSpendResponse{
		Status:      true,
		Dates:       dates,
		Impressions: impressions,
		Spend:       spend,
	}, nil
}

// ─────────────────────────────────────────────
// GetClicksVsCTR
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetClicksVsCTR(ctx context.Context, req *types.MetaAdsRequest) (*types.ClicksVsCTRResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetDailyMetrics(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetClicksVsCTR failed")
		return nil, err
	}

	dates := make([]string, len(rows))
	clicks := make([]float64, len(rows))
	ctr := make([]float64, len(rows))
	for i, r := range rows {
		dates[i] = r.InsightsDate
		clicks[i] = float64(r.Clicks)
		ctr[i] = r.CTR
	}
	return &types.ClicksVsCTRResponse{
		Status: true,
		Dates:  dates,
		Clicks: clicks,
		CTR:    ctr,
	}, nil
}

// ─────────────────────────────────────────────
// GetTopCampaigns
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetTopCampaigns(ctx context.Context, req *types.MetaAdsRequest) (*types.TopCampaignsResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	sortBy := req.SortBy
	if sortBy != "impressions" && sortBy != "ctr" {
		sortBy = "spend"
	}
	rows, err := s.repo.GetTopCampaigns(ctx, params, sortBy)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetTopCampaigns failed")
		return nil, err
	}

	data := make([]types.TopCampaignRow, len(rows))
	for i, r := range rows {
		data[i] = types.TopCampaignRow{
			CampaignID:   r.CampaignID,
			CampaignName: r.CampaignName,
			Spend:        r.Spend,
			Impressions:  r.Impressions,
			CTR:          r.CTR,
		}
	}
	return &types.TopCampaignsResponse{Status: true, Data: data}, nil
}

// ─────────────────────────────────────────────
// GetPerformanceTrend
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetPerformanceTrend(ctx context.Context, req *types.MetaAdsRequest) (*types.PerformanceTrendResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	metric := normalizeMetric(req.Metric)
	rows, err := s.repo.GetPerformanceTrend(ctx, params, metric)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetPerformanceTrend failed")
		return nil, err
	}

	dates := make([]string, len(rows))
	values := make([]float64, len(rows))
	var total float64
	for i, r := range rows {
		dates[i] = r.InsightsDate
		values[i] = r.Value
		total += r.Value
	}
	return &types.PerformanceTrendResponse{
		Status: true,
		Metric: metric,
		Dates:  dates,
		Values: values,
		Total:  total,
	}, nil
}

// ─────────────────────────────────────────────
// GetPerformanceByLevel
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetPerformanceByLevel(ctx context.Context, req *types.MetaAdsRequest) (*types.PerformanceByLevelResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	metric := normalizeMetric(req.Metric)
	level := req.Level
	if level == "" {
		level = "campaign"
	}

	var rows []repo.LevelBreakdownRow
	switch level {
	case "adset":
		rows, err = s.repo.GetPerformanceByAdSet(ctx, params, metric)
	case "ad":
		rows, err = s.repo.GetPerformanceByAd(ctx, params, metric)
	default:
		rows, err = s.repo.GetPerformanceByCampaign(ctx, params, metric)
		level = "campaign"
	}
	if err != nil {
		s.logger.Error().Err(err).Str("level", level).Msg("GetPerformanceByLevel failed")
		return nil, err
	}

	data := make([]types.PerformanceLevelRow, len(rows))
	for i, r := range rows {
		data[i] = types.PerformanceLevelRow{ID: r.ID, Name: r.Name, Value: r.Value}
	}
	return &types.PerformanceByLevelResponse{
		Status:  true,
		Level:   level,
		Metric:  metric,
		Data:    data,
		HasMore: len(rows) == 20,
	}, nil
}

// ─────────────────────────────────────────────
// GetPerformanceByPlatform
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetPerformanceByPlatform(ctx context.Context, req *types.MetaAdsRequest) (*types.PerformanceByPlatformResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	metric := normalizeMetric(req.Metric)
	rows, err := s.repo.GetPerformanceByPlatform(ctx, params, metric)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetPerformanceByPlatform failed")
		return nil, err
	}

	var total float64
	for _, r := range rows {
		total += r.Value
	}

	data := make([]types.PlatformBreakdownRow, len(rows))
	for i, r := range rows {
		pct := safeRate(r.Value*100, total)
		data[i] = types.PlatformBreakdownRow{
			Platform: r.Platform,
			Value:    r.Value,
			Percent:  math.Round(pct*100) / 100,
		}
	}
	return &types.PerformanceByPlatformResponse{
		Status: true,
		Metric: metric,
		Total:  total,
		Data:   data,
	}, nil
}

// ─────────────────────────────────────────────
// GetCampaignsList
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetCampaignsList(ctx context.Context, req *types.MetaAdsRequest) (*types.TableResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	page, perPage := normalizePagination(req.Page, req.PerPage)

	var rows []repo.CampaignTableRow
	var total int64
	var statuses, objectives []string

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var e error
		rows, total, e = s.repo.GetCampaignsList(ctx, params, req.Status, req.Objective, req.Search, req.OrderBy, req.OrderDir, page, perPage)
		return e
	})
	eg.Go(func() error {
		var e error
		statuses, objectives, e = s.repo.GetCampaignFilterOptions(ctx, params)
		return e
	})
	if err := eg.Wait(); err != nil {
		s.logger.Error().Err(err).Msg("GetCampaignsList failed")
		return nil, err
	}

	data := make([]types.TableRow, len(rows))
	for i, r := range rows {
		data[i] = types.TableRow{
			ID:          r.CampaignID,
			Name:        r.CampaignName,
			Status:      r.Status,
			Objective:   r.Objective,
			Results:     r.Results,
			Spend:       r.Spend,
			Reach:       r.Reach,
			Impressions: r.Impressions,
			Frequency:   r.Frequency,
			Clicks:      r.Clicks,
			CPM:         r.CPM,
			CPC:         r.CPC,
			CTR:         r.CTR,
		}
	}
	return &types.TableResponse{
		Status:              true,
		Data:                data,
		Total:               total,
		Page:                page,
		PerPage:             perPage,
		AvailableStatuses:   statuses,
		AvailableObjectives: objectives,
	}, nil
}

// ─────────────────────────────────────────────
// GetAdSetsList
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetAdSetsList(ctx context.Context, req *types.MetaAdsRequest) (*types.TableResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	page, perPage := normalizePagination(req.Page, req.PerPage)

	var rows []repo.AdSetTableRow
	var total int64
	var statuses []string

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var e error
		rows, total, e = s.repo.GetAdSetsList(ctx, params, req.Status, req.Objective, req.Search, req.OrderBy, req.OrderDir, page, perPage)
		return e
	})
	eg.Go(func() error {
		var e error
		statuses, e = s.repo.GetAdSetFilterOptions(ctx, params)
		return e
	})
	if err := eg.Wait(); err != nil {
		s.logger.Error().Err(err).Msg("GetAdSetsList failed")
		return nil, err
	}

	// Campaign objectives (reuse the campaign filter options)
	_, objectives, err := s.repo.GetCampaignFilterOptions(ctx, params)
	if err != nil {
		objectives = nil
	}

	data := make([]types.TableRow, len(rows))
	for i, r := range rows {
		data[i] = types.TableRow{
			ID:          r.AdSetID,
			Name:        r.AdSetName,
			ParentID:    r.CampaignID,
			ParentName:  r.CampaignName,
			Status:      r.Status,
			Objective:   r.Objective,
			Results:     r.Results,
			Spend:       r.Spend,
			Reach:       r.Reach,
			Impressions: r.Impressions,
			Frequency:   r.Frequency,
			Clicks:      r.Clicks,
			CPM:         r.CPM,
			CPC:         r.CPC,
			CTR:         r.CTR,
		}
	}
	return &types.TableResponse{
		Status:              true,
		Data:                data,
		Total:               total,
		Page:                page,
		PerPage:             perPage,
		AvailableStatuses:   statuses,
		AvailableObjectives: objectives,
	}, nil
}

// ─────────────────────────────────────────────
// GetAdsList
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetAdsList(ctx context.Context, req *types.MetaAdsRequest) (*types.TableResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	page, perPage := normalizePagination(req.Page, req.PerPage)

	var rows []repo.AdTableRow
	var total int64
	var statuses []string

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var e error
		rows, total, e = s.repo.GetAdsList(ctx, params, req.Status, req.Objective, req.Search, req.OrderBy, req.OrderDir, page, perPage)
		return e
	})
	eg.Go(func() error {
		var e error
		statuses, e = s.repo.GetAdFilterOptions(ctx, params)
		return e
	})
	if err := eg.Wait(); err != nil {
		s.logger.Error().Err(err).Msg("GetAdsList failed")
		return nil, err
	}

	_, objectives, err := s.repo.GetCampaignFilterOptions(ctx, params)
	if err != nil {
		objectives = nil
	}

	data := make([]types.TableRow, len(rows))
	for i, r := range rows {
		data[i] = types.TableRow{
			ID:                             r.AdID,
			Name:                           r.AdName,
			ParentID:                       r.AdSetID,
			ParentName:                     r.AdSetName,
			GrandParentID:                  r.CampaignID,
			GrandParentName:                r.CampaignName,
			Status:                         r.Status,
			Objective:                      r.Objective,
			CreativeName:                   r.CreativeName,
			CreativeTitle:                  r.CreativeTitle,
			CreativeBody:                   r.CreativeBody,
			CreativeThumbnailURL:           r.CreativeThumbnailURL,
			CreativeEffectiveObjectStoryID: r.CreativeEffectiveObjectStoryID,
			Results:                        r.Results,
			Spend:                          r.Spend,
			Reach:                          r.Reach,
			Impressions:                    r.Impressions,
			Frequency:                      r.Frequency,
			Clicks:                         r.Clicks,
			CPM:                            r.CPM,
			CPC:                            r.CPC,
			CTR:                            r.CTR,
		}
	}
	return &types.TableResponse{
		Status:              true,
		Data:                data,
		Total:               total,
		Page:                page,
		PerPage:             perPage,
		AvailableStatuses:   statuses,
		AvailableObjectives: objectives,
	}, nil
}

// ─────────────────────────────────────────────
// GetDemographicsAgeGender
// ─────────────────────────────────────────────

var ageOrder = []string{"13-17", "18-24", "25-34", "35-44", "45-54", "55-64", "65+"}

func (s *MetaAdsService) GetDemographicsAgeGender(ctx context.Context, req *types.MetaAdsRequest) (*types.DemographicsAgeGenderResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	metric := normalizeMetric(req.Metric)
	rows, err := s.repo.GetDemographicsAgeGender(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetDemographicsAgeGender failed")
		return nil, err
	}

	// Aggregate by age and by gender
	ageMap := make(map[string]float64)
	genderMap := make(map[string]float64)
	genderCount := make(map[string]int64)

	for _, r := range rows {
		val := metricValueForRow(r, metric)
		ageMap[r.Age] += val
		genderMap[r.Gender] += val
		genderCount[r.Gender] += r.Impressions
	}

	// Age totals
	var ageTotal float64
	for _, v := range ageMap {
		ageTotal += v
	}

	byAge := make([]types.AgeBreakdownRow, 0, len(ageOrder))
	for _, age := range ageOrder {
		if val, ok := ageMap[age]; ok {
			pct := safeRate(val*100, ageTotal)
			byAge = append(byAge, types.AgeBreakdownRow{
				AgeRange: age,
				Value:    val,
				Percent:  math.Round(pct*100) / 100,
			})
		}
	}

	// Gender totals
	var genderTotal float64
	for _, v := range genderMap {
		genderTotal += v
	}
	genders := []string{"male", "female", "unknown"}
	byGender := make([]types.GenderBreakdownRow, 0, 3)
	for _, g := range genders {
		if val, ok := genderMap[g]; ok {
			pct := safeRate(val*100, genderTotal)
			byGender = append(byGender, types.GenderBreakdownRow{
				Gender:  g,
				Value:   val,
				Percent: math.Round(pct*100) / 100,
				Count:   genderCount[g],
			})
		}
	}

	return &types.DemographicsAgeGenderResponse{
		Status:   true,
		Metric:   metric,
		ByAge:    byAge,
		ByGender: byGender,
	}, nil
}

func metricValueForRow(r repo.AgeGenderRow, metric string) float64 {
	switch metric {
	case "impressions":
		return float64(r.Impressions)
	case "reach":
		return float64(r.Reach)
	case "clicks":
		return float64(r.Clicks)
	case "cpm":
		return r.CPM
	case "cpc":
		return r.CPC
	case "ctr":
		return r.CTR
	case "frequency":
		return r.Frequency
	default:
		return r.Spend
	}
}

// ─────────────────────────────────────────────
// GetDemographicsRegionCountry
// ─────────────────────────────────────────────

func (s *MetaAdsService) GetDemographicsRegionCountry(ctx context.Context, req *types.MetaAdsRequest) (*types.DemographicsRegionCountryResponse, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}

	breakdown := req.Breakdown
	if breakdown != "country" {
		breakdown = "region"
	}

	var rows []repo.RegionCountryRow
	var countries []string

	eg, ctx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		var e error
		rows, e = s.repo.GetDemographicsRegionCountry(ctx, params, breakdown, req.Country, req.OrderBy, req.OrderDir)
		return e
	})
	if breakdown == "region" {
		eg.Go(func() error {
			var e error
			countries, e = s.repo.GetAvailableCountries(ctx, params)
			return e
		})
	}
	if err := eg.Wait(); err != nil {
		s.logger.Error().Err(err).Msg("GetDemographicsRegionCountry failed")
		return nil, err
	}

	data := make([]types.RegionCountryRow, len(rows))
	for i, r := range rows {
		data[i] = types.RegionCountryRow{
			Country:     r.Country,
			Region:      r.Region,
			Spend:       r.Spend,
			Impressions: r.Impressions,
			Clicks:      r.Clicks,
			CTR:         r.CTR,
		}
	}
	return &types.DemographicsRegionCountryResponse{
		Status:             true,
		Breakdown:          breakdown,
		Data:               data,
		AvailableCountries: countries,
	}, nil
}

// GetAIInsightsSummary returns AI-powered insights summary for Meta Ads.
func (s *MetaAdsService) GetAIInsightsSummary(ctx context.Context, req *types.MetaAdsRequest) (map[string]interface{}, error) {
	if s.aiService == nil {
		return nil, fmt.Errorf("AI insights service not initialized")
	}
	return s.aiService.GetAIInsights(ctx, "aiInsightsSummary", req)
}

// GetAIInsightsDetailed returns detailed AI-powered insights for Meta Ads.
func (s *MetaAdsService) GetAIInsightsDetailed(ctx context.Context, req *types.MetaAdsRequest) (map[string]interface{}, error) {
	if s.aiService == nil {
		return nil, fmt.Errorf("AI insights service not initialized")
	}
	return s.aiService.GetAIInsights(ctx, "aiInsightsDetailed", req)
}

// ─────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────

func normalizeMetric(m string) string {
	valid := map[string]bool{
		"spend": true, "impressions": true, "reach": true, "clicks": true,
		"cpm": true, "cpc": true, "ctr": true, "frequency": true,
	}
	if valid[m] {
		return m
	}
	return "spend"
}

func normalizePagination(page, perPage int) (int, int) {
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 10
	}
	return page, perPage
}
