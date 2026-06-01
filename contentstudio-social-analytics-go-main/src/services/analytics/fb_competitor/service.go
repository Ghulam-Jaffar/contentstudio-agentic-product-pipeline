// Package fb_competitor provides the business logic layer for Facebook competitor analytics.
// It resolves competitor accounts from MongoDB reports, then queries ClickHouse.
// Migrated from PHP FacebookCompetitorController (contentstudio-backend).
package fb_competitor

import (
	"context"
	"fmt"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	chRepo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/fb_competitor"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/fb_competitor"
	mongoModels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/mongo"
)

// CompetitorRepo is the interface for MongoDB competitor operations.
type CompetitorRepo interface {
	GetReportCompetitors(ctx context.Context, reportID string) (map[string]*mongoModels.Competitor, error)
}

// ClickHouseRepo is the interface for Facebook competitor ClickHouse queries.
type ClickHouseRepo interface {
	GetDataTableMetrics(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
	GetPostingActivityGraphByTypes(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
	GetPostingActivityBySpecificType(ctx context.Context, params *chRepo.CompetitorQueryParams, mediaType, sortOrder string) ([]map[string]interface{}, error)
	GetTopAndLeastPerformingPosts(ctx context.Context, params *chRepo.CompetitorQueryParams) ([]map[string]interface{}, error)
	GetTopHashtags(ctx context.Context, params *chRepo.CompetitorQueryParams, limit int) ([]map[string]interface{}, error)
	GetIndividualHashtagData(ctx context.Context, params *chRepo.CompetitorQueryParams, hashtag string) ([]map[string]interface{}, error)
	GetBiographyData(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
	GetFollowersGrowthComparison(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
	GetPostReactDistribution(ctx context.Context, params *chRepo.CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error)
	GetPostReactDistributionByCompany(ctx context.Context, params *chRepo.CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error)
	GetPostTypeDistribution(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
	GetPostEngagementOverTime(ctx context.Context, params *chRepo.CompetitorQueryParams, facebookID, sortOrder string) ([]map[string]interface{}, error)
	GetPostEngagementByCompetitor(ctx context.Context, params *chRepo.CompetitorQueryParams, sortOrder string) ([]map[string]interface{}, error)
}

// Service defines the interface for Facebook competitor analytics business logic.
type Service interface {
	GetDataTableMetrics(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostingActivityGraphByTypes(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostingActivityBySpecificType(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetTopAndLeastPerformingPosts(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetTopHashtags(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetIndividualHashtagData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetBiographyData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetFollowersGrowthComparison(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostReactDistribution(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostReactDistributionByCompany(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostTypeDistribution(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostEngagementOverTime(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
	GetPostEngagementByCompetitor(ctx context.Context, req *types.CompetitorRequest) (interface{}, error)
}

// FacebookCompetitorService implements Service.
type FacebookCompetitorService struct {
	chRepo    ClickHouseRepo
	mongoRepo CompetitorRepo
	logger    zerolog.Logger
}

// NewFacebookCompetitorService creates a new service.
func NewFacebookCompetitorService(ch ClickHouseRepo, mongo CompetitorRepo, logger zerolog.Logger) *FacebookCompetitorService {
	return &FacebookCompetitorService{
		chRepo:    ch,
		mongoRepo: mongo,
		logger:    logger.With().Str("service", "fb-competitor").Logger(),
	}
}

var _ Service = (*FacebookCompetitorService)(nil)

// resolveParams resolves a CompetitorRequest into ClickHouse query params by
// fetching competitor accounts from MongoDB and converting dates to UTC.
func (s *FacebookCompetitorService) resolveParams(ctx context.Context, req *types.CompetitorRequest) (*chRepo.CompetitorQueryParams, error) {
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
func (s *FacebookCompetitorService) GetDataTableMetrics(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
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

	// Build prev lookup by facebook_id
	prevMap := make(map[string]map[string]interface{}, len(dataPrev))
	for _, row := range dataPrev {
		if id, ok := row["facebook_id"]; ok {
			prevMap[fmt.Sprintf("%v", id)] = row
		}
	}

	// Calculate growth diffs like PHP
	metrics := make([]map[string]interface{}, 0, len(dataCurrent))
	for _, curr := range dataCurrent {
		fbID := fmt.Sprintf("%v", curr["facebook_id"])
		prev := prevMap[fbID]
		if prev == nil {
			prev = map[string]interface{}{}
		}
		curr["averagePostsPerWeekDiff"] = calculateGrowth(toFloat64(curr["averagePostsPerWeek"]), toFloat64(prev["averagePostsPerWeek"]))
		curr["engagementRateDiff"] = calculateGrowth(toFloat64(curr["engagementRate"]), toFloat64(prev["engagementRate"]))
		curr["followersCountDiff"] = calculateGrowth(toFloat64(curr["followersCount"]), toFloat64(prev["followersCount"]))
		curr["fanCountDiff"] = calculateGrowth(toFloat64(curr["fanCount"]), toFloat64(prev["fanCount"]))
		metrics = append(metrics, curr)
	}

	return map[string]interface{}{
		"data":               dataCurrent,
		"data_prev":          dataPrev,
		"data_table_metrics": metrics,
	}, nil
}

// GetPostingActivityGraphByTypes returns aggregated posting activity per media type.
func (s *FacebookCompetitorService) GetPostingActivityGraphByTypes(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
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
func (s *FacebookCompetitorService) GetPostingActivityBySpecificType(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetPostingActivityBySpecificType(ctx, params, req.MediaType, req.GetSortOrder("followersCount"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetTopAndLeastPerformingPosts returns top 5 and least 5 performing posts per competitor.
func (s *FacebookCompetitorService) GetTopAndLeastPerformingPosts(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetTopAndLeastPerformingPosts(ctx, params)
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": transformTopAndLeastPosts(data)}, nil
}

// transformTopAndLeastPosts groups flat ClickHouse rows (one per post×media asset)
// into per-competitor objects with nested top_5_posts / least_5_posts arrays,
// matching the structure the PHP controller returned.
func transformTopAndLeastPosts(rows []map[string]interface{}) []map[string]interface{} {
	if len(rows) == 0 {
		return []map[string]interface{}{}
	}

	// Group rows by facebook_id, preserving order.
	pageOrder := make([]string, 0)
	byPage := make(map[string][]map[string]interface{})
	for _, row := range rows {
		fbID := fmt.Sprintf("%v", row["facebook_id"])
		if _, exists := byPage[fbID]; !exists {
			pageOrder = append(pageOrder, fbID)
		}
		byPage[fbID] = append(byPage[fbID], row)
	}

	result := make([]map[string]interface{}, 0, len(pageOrder))
	for _, fbID := range pageOrder {
		pageRows := byPage[fbID]
		firstRow := pageRows[0]

		// Separate by category, then group by post_id.
		topOrder, topByPost := groupPostsByCategory(pageRows, "top_5_posts")
		leastOrder, leastByPost := groupPostsByCategory(pageRows, "least_5_posts")

		competitor := map[string]interface{}{
			"name":            firstRow["page_name"],
			"facebook_id":     firstRow["facebook_id"],
			"image":           firstRow["image"],
			"followers_count": firstRow["followers_count"],
			"fan_count":       firstRow["fan_count"],
			"page_name":       firstRow["page_name"],
			"page_category":   firstRow["page_category"],
			"biography":       firstRow["biography"],
			"top_5_posts":     buildPostsList(topByPost, topOrder),
			"least_5_posts":   buildPostsList(leastByPost, leastOrder),
		}
		result = append(result, competitor)
	}
	return result
}

// groupPostsByCategory filters rows matching category and groups them by post_id.
func groupPostsByCategory(rows []map[string]interface{}, category string) ([]string, map[string][]map[string]interface{}) {
	order := make([]string, 0)
	byPost := make(map[string][]map[string]interface{})
	for _, row := range rows {
		if fmt.Sprintf("%v", row["category"]) != category {
			continue
		}
		postID := fmt.Sprintf("%v", row["post_id"])
		if _, exists := byPost[postID]; !exists {
			order = append(order, postID)
		}
		byPost[postID] = append(byPost[postID], row)
	}
	return order, byPost
}

// buildPostsList converts grouped post rows into the frontend-expected format,
// collapsing multiple rows per post_id into a single post object with a media array.
func buildPostsList(byPost map[string][]map[string]interface{}, order []string) []map[string]interface{} {
	posts := make([]map[string]interface{}, 0, len(order))
	for _, postID := range order {
		rows := byPost[postID]
		first := rows[0]

		// Collect media assets (skip NULL rows from LEFT JOIN).
		media := make([]map[string]interface{}, 0)
		for _, row := range rows {
			mid := fmt.Sprintf("%v", row["media_id"])
			if mid == "" || mid == "<nil>" {
				continue
			}
			media = append(media, map[string]interface{}{
				"id":             row["media_id"],
				"caption":        row["media_assets.caption"],
				"link":           row["link"],
				"asset_type":     row["asset_type"],
				"call_to_action": row["call_to_action"],
				"created_at":     row["media_assets.created_at"],
			})
		}

		post := map[string]interface{}{
			"id":                   first["post_id"],
			"post_engagement":      first["post_engagement"],
			"like":                 first["like"],
			"haha":                 first["haha"],
			"angry":                first["angry"],
			"sad":                  first["sad"],
			"love":                 first["love"],
			"thankful":             first["thankful"],
			"total_post_reactions": first["total_post_reactions"],
			"comments":             first["comments"],
			"shares":               first["shares"],
			"caption":              first["caption"],
			"media_type":           first["media_type"],
			"status_type":          first["status_type"],
			"permalink":            first["permalink"],
			"shared_from_name":     first["shared_from_name"],
			"shared_from_id":       first["shared_from_id"],
			"shared_from_pic":      first["shared_from_pic"],
			"hashtags":             first["hashtags"],
			"day_of_week":          first["day_of_week"],
			"hour_of_day":          first["hour_of_day"],
			"created_at":           first["created_at"],
			"inserted_at":          first["inserted_at"],
			"media":                media,
		}
		posts = append(posts, post)
	}
	return posts
}

// GetTopHashtags returns top hashtags across all competitors.
func (s *FacebookCompetitorService) GetTopHashtags(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
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
func (s *FacebookCompetitorService) GetIndividualHashtagData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
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
func (s *FacebookCompetitorService) GetBiographyData(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
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

// GetFollowersGrowthComparison returns per-competitor follower growth time-series.
func (s *FacebookCompetitorService) GetFollowersGrowthComparison(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetFollowersGrowthComparison(ctx, params, req.GetSortOrder("followers_count"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetPostReactDistribution returns engagement aggregates for a single competitor.
func (s *FacebookCompetitorService) GetPostReactDistribution(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	if req.FacebookID == "" {
		return nil, fmt.Errorf("facebook_id is required for postReactDistribution")
	}
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetPostReactDistribution(ctx, params, req.FacebookID, req.GetSortOrder("followers_count"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetPostReactDistributionByCompany returns reaction breakdown for a single competitor.
func (s *FacebookCompetitorService) GetPostReactDistributionByCompany(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	if req.FacebookID == "" {
		return nil, fmt.Errorf("facebook_id is required for postReactDistributionByCompany")
	}
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetPostReactDistributionByCompany(ctx, params, req.FacebookID, req.GetSortOrder("followers_count"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetPostTypeDistribution returns post type distribution per competitor.
func (s *FacebookCompetitorService) GetPostTypeDistribution(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetPostTypeDistribution(ctx, params, req.GetSortOrder("followers_count"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetPostEngagementOverTime returns daily engagement totals for a single competitor.
func (s *FacebookCompetitorService) GetPostEngagementOverTime(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	if req.FacebookID == "" {
		return nil, fmt.Errorf("facebook_id is required for postEngagementOverTime")
	}
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetPostEngagementOverTime(ctx, params, req.FacebookID, req.GetSortOrder("followers_count"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}

// GetPostEngagementByCompetitor returns total engagement per competitor.
func (s *FacebookCompetitorService) GetPostEngagementByCompetitor(ctx context.Context, req *types.CompetitorRequest) (interface{}, error) {
	params, err := s.resolveParams(ctx, req)
	if err != nil {
		return nil, err
	}
	data, err := s.chRepo.GetPostEngagementByCompetitor(ctx, params, req.GetSortOrder("followers_count"))
	if err != nil {
		return nil, err
	}
	return map[string]interface{}{"data": data}, nil
}
