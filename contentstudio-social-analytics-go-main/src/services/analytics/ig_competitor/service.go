// Package ig_competitor provides the business logic layer for Instagram competitor analytics.
// It resolves competitor accounts from MongoDB reports, then queries ClickHouse.
// Migrated from PHP InstagramCompetitorController (contentstudio-backend).
package ig_competitor

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	chRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/ig_competitor"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/ig_competitor"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// CompetitorRepo is the interface for MongoDB competitor operations.
type CompetitorRepo interface {
	GetReportCompetitors(ctx context.Context, reportID string) (map[string]*mongoModels.Competitor, error)
}

// ClickHouseRepo is the interface for Instagram competitor ClickHouse queries.
type ClickHouseRepo interface {
	GetDataTableMetrics(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
	GetPostingActivityGraphByTypes(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
	GetPostingActivityBySpecificType(ctx context.Context, params *chRepo.CompetitorQueryParams, mediaType, mediaProductType, sortOrder string) ([]map[string]interface{}, error)
	GetPostingActivityTableByType(ctx context.Context, params *chRepo.CompetitorQueryParams, mediaType, mediaProductType, sortOrder string) ([]map[string]interface{}, error)
	GetFollowersGrowthComparison(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
	GetTopAndLeastPerformingPosts(ctx context.Context, params *chRepo.CompetitorQueryParams) ([]map[string]interface{}, error)
	GetTopHashtags(ctx context.Context, params *chRepo.CompetitorQueryParams, limit int) ([]map[string]interface{}, error)
	GetIndividualHashtagData(ctx context.Context, params *chRepo.CompetitorQueryParams, hashtag string) ([]map[string]interface{}, error)
	GetBiographyData(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
}

// Service defines the interface for Instagram competitor analytics business logic.
type Service interface {
	GetDataTableMetrics(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostingActivityGraphByTypes(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostingActivityBySpecificType(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostingActivityTableByType(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetFollowersGrowthComparison(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetTopAndLeastPerformingPosts(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetTopHashtags(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetIndividualHashtagData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetBiographyData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
}

// InstagramCompetitorService implements Service.
type InstagramCompetitorService struct {
	chRepo    ClickHouseRepo
	mongoRepo CompetitorRepo
	logger    zerolog.Logger
}

// NewInstagramCompetitorService creates a new service.
func NewInstagramCompetitorService(ch ClickHouseRepo, mongo CompetitorRepo, logger zerolog.Logger) *InstagramCompetitorService {
	return &InstagramCompetitorService{
		chRepo:    ch,
		mongoRepo: mongo,
		logger:    logger.With().Str("service", "ig-competitor").Logger(),
	}
}

var _ Service = (*InstagramCompetitorService)(nil)

// resolveParams resolves a CompetitorRequest into ClickHouse query params.
func (s *InstagramCompetitorService) resolveParams(ctx context.Context, req *types.CompetitorRequest) (*chRepo.CompetitorQueryParams, error) {
	startUTC, endUTC, daysDiff, err := req.ToUTCDateRange()
	if err != nil {
		return nil, err
	}

	accounts := map[string]chRepo.AccountInfo{}
	var pageIDs []string

	if req.ReportID != "" {
		competitors, err := s.mongoRepo.GetReportCompetitors(ctx, req.ReportID)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve report competitors: %w", err)
		}
		for id, comp := range competitors {
			pageIDs = append(pageIDs, id)
			accounts[id] = chRepo.AccountInfo{
				Image: comp.Image,
				Name:  comp.Name,
				State: comp.State,
				Slug:  comp.Slug,
			}
		}
	}

	return &chRepo.CompetitorQueryParams{
		PageIDs:   pageIDs,
		Accounts:  accounts,
		StartDate: startUTC,
		EndDate:   endUTC,
		DaysDiff:  daysDiff,
	}, nil
}

func calculateGrowth(current, previous float64) interface{} {
	if previous == 0 {
		return "N/A"
	}
	growth := (current - previous) / previous * 100
	return fmt.Sprintf("%.2f", growth)
}

func toFloat64(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int64:
		return float64(val)
	case int32:
		return float64(val)
	case int:
		return float64(val)
	case uint64:
		return float64(val)
	default:
		return 0
	}
}

// GetDataTableMetrics fetches current and previous period metrics concurrently.
func (s *InstagramCompetitorService) GetDataTableMetrics(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	prevParams := params.PrevPeriod()
	sortOrder := req.GetSortOrder("followersCount")

	var dataCurrent, dataPrev []map[string]interface{}
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.chRepo.GetDataTableMetrics(egCtx, params, sortOrder)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetDataTableMetrics: current query failed")
			return err
		}
		dataCurrent = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.chRepo.GetDataTableMetrics(egCtx, prevParams, sortOrder)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetDataTableMetrics: previous query failed")
			return err
		}
		dataPrev = r
		return nil
	})
	if err := eg.Wait(); err != nil {
		return nil, err
	}

	// Build prev lookup by business_account_id
	prevMap := make(map[string]map[string]interface{}, len(dataPrev))
	for _, row := range dataPrev {
		if id, ok := row["business_account_id"]; ok {
			prevMap[fmt.Sprintf("%v", id)] = row
		}
	}

	// Calculate growth diffs like PHP
	metrics := make([]map[string]interface{}, 0, len(dataCurrent))
	for _, curr := range dataCurrent {
		baID := fmt.Sprintf("%v", curr["business_account_id"])
		prev := prevMap[baID]
		if prev == nil {
			prev = map[string]interface{}{}
		}
		curr["averagePostsPerWeekDiff"] = calculateGrowth(toFloat64(curr["averagePostsPerWeek"]), toFloat64(prev["averagePostsPerWeek"]))
		curr["engagementRateDiff"] = calculateGrowth(toFloat64(curr["engagementRate"]), toFloat64(prev["engagementRate"]))
		curr["followersCountDiff"] = calculateGrowth(toFloat64(curr["followersCount"]), toFloat64(prev["followersCount"]))
		curr["followingCountDiff"] = calculateGrowth(toFloat64(curr["followingCount"]), toFloat64(prev["followingCount"]))
		metrics = append(metrics, curr)
	}

	return map[string]interface{}{
		"data":               dataCurrent,
		"data_prev":          dataPrev,
		"data_table_metrics": metrics,
	}, nil
}

// GetPostingActivityGraphByTypes returns aggregated posting activity per media type.
func (s *InstagramCompetitorService) GetPostingActivityGraphByTypes(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetPostingActivityGraphByTypes(ctx, params, req.GetSortOrder("avgTotalEngagements"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetPostingActivityBySpecificType returns per-competitor metrics for a specific media type.
func (s *InstagramCompetitorService) GetPostingActivityBySpecificType(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetPostingActivityBySpecificType(ctx, params, req.MediaType, req.MediaProductType, req.GetSortOrder("followersCount"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetPostingActivityTableByType returns table view metrics for a specific media type.
func (s *InstagramCompetitorService) GetPostingActivityTableByType(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetPostingActivityTableByType(ctx, params, req.MediaType, req.MediaProductType, req.GetSortOrder("followersCount"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetFollowersGrowthComparison returns per-competitor follower growth time-series.
func (s *InstagramCompetitorService) GetFollowersGrowthComparison(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetFollowersGrowthComparison(ctx, params, req.GetSortOrder("total_followed_by_count"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetTopAndLeastPerformingPosts returns top 5 and least 5 performing posts per competitor.
func (s *InstagramCompetitorService) GetTopAndLeastPerformingPosts(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetTopAndLeastPerformingPosts(ctx, params)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetTopHashtags returns top hashtags across all competitors.
func (s *InstagramCompetitorService) GetTopHashtags(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetTopHashtags(ctx, params, req.GetLimit())
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetIndividualHashtagData returns per-competitor data for a specific hashtag.
func (s *InstagramCompetitorService) GetIndividualHashtagData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetIndividualHashtagData(ctx, params, req.Hashtag)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetBiographyData returns biography data per competitor.
func (s *InstagramCompetitorService) GetBiographyData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetBiographyData(ctx, params, req.GetSortOrder("biography_length"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}
