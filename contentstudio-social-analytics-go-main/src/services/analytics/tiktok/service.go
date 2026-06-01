package tiktok

import (
	"context"
	"math"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/tiktok"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/tiktok"
)

// Service defines the TikTok analytics operations exposed to the HTTP handlers.
type Service interface {
	GetPageAndPostsInsights(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error)
	GetPageFollowersAndViews(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error)
	GetDynamicPageFollowersAndViews(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error)
	GetPostsAndEngagements(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error)
	GetDailyEngagementsData(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error)
	GetDynamicDailyEngagementsData(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error)
	GetTopAndLeastPerformingPosts(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error)
	GetPostsData(ctx context.Context, req *types.PostsRequest) (map[string]interface{}, error)
}

// TiktokAnalyticsService coordinates repository reads and maps them to the legacy response shape.
type TiktokAnalyticsService struct {
	repo   *repo.Repository
	logger zerolog.Logger
}

var _ Service = (*TiktokAnalyticsService)(nil)

// NewTiktokAnalyticsService constructs a TikTok analytics service with ClickHouse-backed queries.
func NewTiktokAnalyticsService(r *repo.Repository, logger zerolog.Logger) *TiktokAnalyticsService {
	return &TiktokAnalyticsService{
		repo:   r,
		logger: logger.With().Str("service", "tiktok-analytics").Logger(),
	}
}

// GetPageAndPostsInsights fetches current, previous, and previous-previous summary windows concurrently
// so the overview cards can preserve PHP parity while avoiding sequential ClickHouse reads.
func (s *TiktokAnalyticsService) GetPageAndPostsInsights(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	prevParams := s.prevPeriodParams(params)
	prevPrevParams := s.prevPeriodParams(prevParams)

	var current, previous, previousPrevious *repo.SummaryResult
	eg, egCtx := errgroup.WithContext(ctx)
	eg.Go(func() error {
		r, err := s.repo.GetSummary(egCtx, params)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPageAndPostsInsights: current summary failed")
			r = &repo.SummaryResult{}
		}
		current = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetSummary(egCtx, prevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPageAndPostsInsights: previous summary failed")
			r = &repo.SummaryResult{}
		}
		previous = r
		return nil
	})
	eg.Go(func() error {
		r, err := s.repo.GetSummary(egCtx, prevPrevParams)
		if err != nil {
			s.logger.Error().Err(err).Msg("GetPageAndPostsInsights: prev-prev summary failed")
			r = &repo.SummaryResult{}
		}
		previousPrevious = r
		return nil
	})
	_ = eg.Wait()

	_ = previousPrevious // kept for parity with PHP's 3-period fetch

	data := map[string]interface{}{
		"tiktok_id":                current.TiktokID,
		"page_name":                current.PageName,
		"logo":                     current.Logo,
		"total_likes":              current.TotalLikes,
		"total_likes_growth":       calculateGrowth(current.TotalLikes, previous.TotalLikes),
		"total_likes_diff":         current.TotalLikes - previous.TotalLikes,
		"total_comments":           current.TotalComments,
		"total_comments_growth":    calculateGrowth(current.TotalComments, previous.TotalComments),
		"total_comments_diff":      current.TotalComments - previous.TotalComments,
		"total_shares":             current.TotalShares,
		"total_shares_growth":      calculateGrowth(current.TotalShares, previous.TotalShares),
		"total_shares_diff":        current.TotalShares - previous.TotalShares,
		"total_engagements":        current.TotalEngagements,
		"total_engagements_growth": calculateGrowth(current.TotalEngagements, previous.TotalEngagements),
		"total_engagements_diff":   current.TotalEngagements - previous.TotalEngagements,
		"total_posts":              current.TotalPosts,
		"total_posts_growth":       calculateGrowth(current.TotalPosts, previous.TotalPosts),
		"total_posts_diff":         current.TotalPosts - previous.TotalPosts,
		"total_followers":          current.TotalFollowerCount,
		"total_followers_growth":   calculateGrowth(current.TotalFollowerCount, previous.TotalFollowerCount),
		"total_followers_diff":     calculateDiff(current.TotalFollowerCount, previous.TotalFollowerCount),
		"total_followings":         current.TotalFollowingCount,
		"total_followings_growth":  calculateGrowth(current.TotalFollowingCount, previous.TotalFollowingCount),
		"total_followings_diff":    calculateDiff(current.TotalFollowingCount, previous.TotalFollowingCount),
		"total_video_views":        current.TotalVideoViews,
		"total_video_views_growth": calculateGrowth(current.TotalVideoViews, previous.TotalVideoViews),
		"total_video_views_diff":   current.TotalVideoViews - previous.TotalVideoViews,
	}

	return map[string]interface{}{"status": true, "data": data}, nil
}

// GetPageFollowersAndViews returns the cumulative follower/view series used by page growth charts.
func (s *TiktokAnalyticsService) GetPageFollowersAndViews(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	params.DateFrom = params.DateFrom.AddDate(0, 0, -1)
	result, err := s.repo.GetFollowersAndViews(ctx, params, false)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetPageFollowersAndViews failed")
		return map[string]interface{}{"status": true, "data": []map[string]interface{}{}}, nil
	}
	row := followersViewsToMap(result)
	row = normalizeFollowersViewsRow(row)
	return map[string]interface{}{"status": true, "data": []map[string]interface{}{row}}, nil
}

// GetDynamicPageFollowersAndViews returns daily delta-style follower/view metrics for AI and charts.
func (s *TiktokAnalyticsService) GetDynamicPageFollowersAndViews(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	params.DateFrom = params.DateFrom.AddDate(0, 0, -1)
	result, err := s.repo.GetFollowersAndViews(ctx, params, true)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetDynamicPageFollowersAndViews failed")
		return map[string]interface{}{"status": true, "data": []map[string]interface{}{}}, nil
	}
	row := followersViewsToMap(result)
	return map[string]interface{}{"status": true, "data": []map[string]interface{}{row}}, nil
}

// GetPostsAndEngagements returns publishing and aggregate engagement metrics for the selected period.
func (s *TiktokAnalyticsService) GetPostsAndEngagements(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetPostsAndEngagements(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetPostsAndEngagements failed")
		return map[string]interface{}{"status": true, "data": []map[string]interface{}{}}, nil
	}
	row := map[string]interface{}{
		"tiktok_id":            result.TiktokID,
		"page_name":            result.PageName,
		"logo":                 result.Logo,
		"days_bucket":          result.DaysBucket,
		"sum_view_count":       result.SumViewCount,
		"sum_like_count":       result.SumLikeCount,
		"sum_comments_count":   result.SumCommentsCount,
		"sum_share_count":      result.SumShareCount,
		"sum_engagement_count": result.SumEngagementCount,
		"avg_engagement_rate":  result.AvgEngagementRate,
		"post_count":           result.PostCount,
	}
	if allZeroInt64(result.PostCount) {
		for _, key := range []string{
			"post_count",
			"days_bucket",
			"sum_comments_count",
			"sum_engagement_count",
			"sum_like_count",
			"sum_share_count",
			"sum_view_count",
		} {
			row[key] = []string{}
		}
		row["post_count"] = []int64{}
		row["sum_comments_count"] = []int64{}
		row["sum_engagement_count"] = []int64{}
		row["sum_like_count"] = []int64{}
		row["sum_share_count"] = []int64{}
		row["sum_view_count"] = []int64{}
	}
	return map[string]interface{}{"status": true, "data": []map[string]interface{}{row}}, nil
}

// GetDailyEngagementsData returns cumulative and daily engagement series for the selected window.
func (s *TiktokAnalyticsService) GetDailyEngagementsData(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetDailyEngagementsData(ctx, params, false)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetDailyEngagementsData failed")
		return map[string]interface{}{"status": true, "data": []map[string]interface{}{}}, nil
	}
	row := dailyEngagementToMap(result)
	row = normalizeDailyEngagementRow(row)
	return map[string]interface{}{"status": true, "data": []map[string]interface{}{row}}, nil
}

// GetDynamicDailyEngagementsData returns the dynamic daily engagement variant used by AI insights.
func (s *TiktokAnalyticsService) GetDynamicDailyEngagementsData(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	result, err := s.repo.GetDailyEngagementsData(ctx, params, true)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetDynamicDailyEngagementsData failed")
		return map[string]interface{}{"status": true, "data": []map[string]interface{}{}}, nil
	}
	row := dailyEngagementToMap(result)
	return map[string]interface{}{"status": true, "data": []map[string]interface{}{row}}, nil
}

// GetTopAndLeastPerformingPosts returns both leader and laggard post lists in one response.
func (s *TiktokAnalyticsService) GetTopAndLeastPerformingPosts(ctx context.Context, req *types.TiktokRequest) (map[string]interface{}, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	topRows, leastRows, err := s.repo.GetTopAndLeastPerformingPosts(ctx, params)
	if err != nil {
		s.logger.Error().Err(err).Msg("GetTopAndLeastPerformingPosts failed")
		return map[string]interface{}{"status": true, "data": map[string]interface{}{"top_posts": []map[string]interface{}{}, "least_posts": []map[string]interface{}{}}}, nil
	}
	return map[string]interface{}{
		"status": true,
		"data": map[string]interface{}{
			"top_posts":   mapPosts(topRows, req.GetTimezone()),
			"least_posts": mapPosts(leastRows, req.GetTimezone()),
		},
	}, nil
}

// GetPostsData returns the paginated post table sorted by the requested metric.
func (s *TiktokAnalyticsService) GetPostsData(ctx context.Context, req *types.PostsRequest) (map[string]interface{}, error) {
	params, err := req.ToQueryParams()
	if err != nil {
		return nil, err
	}
	rows, err := s.repo.GetPostsData(ctx, params, req.GetSortOrder(), req.GetLimit(), req.GetOffset())
	if err != nil {
		s.logger.Error().Err(err).Msg("GetPostsData failed")
		return map[string]interface{}{"status": true, "data": []map[string]interface{}{}}, nil
	}
	return map[string]interface{}{"status": true, "data": mapPosts(rows, req.GetTimezone())}, nil
}

// prevPeriodParams rewrites the current query params to target the precomputed previous window.
func (s *TiktokAnalyticsService) prevPeriodParams(params *clickhouse.QueryParams) *clickhouse.QueryParams {
	prev := *params
	prev.DateFrom = params.PrevDateFrom
	prev.DateTo = params.PrevDateTo
	return &prev
}

func followersViewsToMap(result *repo.FollowersViewsResult) map[string]interface{} {
	if result == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"platform_id":          result.PlatformID,
		"display_name":         result.DisplayName,
		"logo":                 result.Logo,
		"followers_count":      result.FollowersCount,
		"views_per_day":        result.ViewsPerDay,
		"followers_count_diff": result.FollowersCountDiff,
		"views_per_day_diff":   result.ViewsPerDayDiff,
		"day_bucket":           result.DayBucket,
	}
}

func normalizeFollowersViewsRow(row map[string]interface{}) map[string]interface{} {
	dayBucket, _ := row["day_bucket"].([]string)
	followers, _ := row["followers_count"].([]int64)
	followerDiffs, _ := row["followers_count_diff"].([]int64)
	views, _ := row["views_per_day"].([]int64)
	viewDiffs, _ := row["views_per_day_diff"].([]int64)

	if len(dayBucket) > 0 {
		dayBucket = dayBucket[1:]
	}
	if len(followers) > 0 {
		followers = followers[1:]
	}
	if len(followerDiffs) > 0 {
		followerDiffs = followerDiffs[1:]
	}
	if len(views) > 0 {
		views = views[1:]
	}
	if len(viewDiffs) > 0 {
		viewDiffs = viewDiffs[1:]
	}

	normFollowers := normalizeSeries(followers, 0, false)
	row["followers_count"] = normFollowers.values
	if len(normFollowers.values) == 0 {
		row["day_bucket"] = []string{}
		row["views_per_day"] = []int64{}
		row["followers_count_diff"] = []int64{}
		row["views_per_day_diff"] = []int64{}
		return row
	}

	row["day_bucket"] = sliceStrings(dayBucket, normFollowers.count)
	row["views_per_day"] = normalizeSeries(views, normFollowers.count, true).values
	row["followers_count_diff"] = sliceInt64(followerDiffs, normFollowers.count)
	row["views_per_day_diff"] = sliceInt64(viewDiffs, normFollowers.count)
	return row
}

func normalizeDailyEngagementRow(row map[string]interface{}) map[string]interface{} {
	totalLikes, _ := row["total_video_likes"].([]int64)
	daysBucket, _ := row["days_bucket"].([]string)
	totalComments, _ := row["total_video_comments"].([]int64)
	totalShares, _ := row["total_video_shares"].([]int64)
	totalEngagement, _ := row["total_engagement"].([]int64)
	dailyLikes, _ := row["daily_video_likes"].([]int64)
	dailyComments, _ := row["daily_video_comments"].([]int64)
	dailyShares, _ := row["daily_video_shares"].([]int64)
	dailyEngagement, _ := row["daily_engagement"].([]int64)

	normLikes := normalizeSeries(totalLikes, 0, false)
	row["total_video_likes"] = normLikes.values
	if len(normLikes.values) == 0 {
		row["days_bucket"] = []string{}
		row["total_video_comments"] = []int64{}
		row["total_video_shares"] = []int64{}
		row["total_engagement"] = []int64{}
		row["daily_video_likes"] = []int64{}
		row["daily_video_comments"] = []int64{}
		row["daily_video_shares"] = []int64{}
		row["daily_engagement"] = []int64{}
		return row
	}

	row["days_bucket"] = sliceStrings(daysBucket, normLikes.count)
	row["total_video_comments"] = normalizeSeries(totalComments, normLikes.count, true).values
	row["total_video_shares"] = normalizeSeries(totalShares, normLikes.count, true).values
	row["total_engagement"] = normalizeSeries(totalEngagement, normLikes.count, true).values
	row["daily_video_likes"] = sliceInt64(dailyLikes, normLikes.count)
	row["daily_video_comments"] = sliceInt64(dailyComments, normLikes.count)
	row["daily_video_shares"] = sliceInt64(dailyShares, normLikes.count)
	row["daily_engagement"] = sliceInt64(dailyEngagement, normLikes.count)
	return row
}

func dailyEngagementToMap(result *repo.DailyEngagementResult) map[string]interface{} {
	if result == nil {
		return map[string]interface{}{}
	}
	return map[string]interface{}{
		"tiktok_id":            result.TiktokID,
		"page_name":            result.PageName,
		"logo":                 result.Logo,
		"total_video_likes":    result.TotalVideoLikes,
		"total_video_comments": result.TotalVideoComments,
		"total_video_shares":   result.TotalVideoShares,
		"daily_video_likes":    result.DailyVideoLikes,
		"daily_video_comments": result.DailyVideoComments,
		"daily_video_shares":   result.DailyVideoShares,
		"total_engagement":     result.TotalEngagement,
		"daily_engagement":     result.DailyEngagement,
		"days_bucket":          result.DaysBucket,
	}
}

func mapPosts(rows []repo.PostRow, timezone string) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(rows))
	for _, row := range rows {
		result = append(result, map[string]interface{}{
			"category":             row.Category,
			"tiktok_id":            row.TiktokID,
			"page_name":            row.PageName,
			"logo":                 row.Logo,
			"profile_link":         row.ProfileLink,
			"post_id":              row.PostID,
			"cover_image_url":      row.CoverImageURL,
			"share_url":            row.ShareURL,
			"post_description":     row.PostDescription,
			"hashtags":             row.Hashtags,
			"duration":             row.Duration,
			"height":               row.Height,
			"width":                row.Width,
			"title":                row.Title,
			"embed_html":           row.EmbedHTML,
			"embed_link":           row.EmbedLink,
			"likes_count":          row.LikesCount,
			"comments_count":       row.CommentsCount,
			"shares_count":         row.SharesCount,
			"views_count":          row.ViewsCount,
			"engagements_count":    row.EngagementsCount,
			"engagement_count":     row.EngagementsCount,
			"total_engagement":     row.TotalEngagement,
			"engagement_rate":      row.EngagementRate,
			"inserted_at":          formatTimeStringInTimezone(row.InsertedAt, timezone),
			"created_time":         formatTimeStringInTimezone(row.CreatedTime, timezone),
			"total_follower_count": row.TotalFollowerCount,
			"total":                row.Total,
		})
	}
	return result
}

func formatTimeStringInTimezone(value, timezone string) string {
	if value == "" {
		return ""
	}
	parsed, err := time.ParseInLocation("2006-01-02 15:04:05", value, time.UTC)
	if err != nil {
		return value
	}
	if timezone == "" || timezone == "UTC" {
		return parsed.UTC().Format(time.RFC3339)
	}
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		return parsed.UTC().Format(time.RFC3339)
	}
	return parsed.In(loc).Format(time.RFC3339)
}

func calculateDiff(current, previous int64) interface{} {
	if previous == 0 {
		return "N/A"
	}
	return current - previous
}

func calculateGrowth(current, previous int64) interface{} {
	if previous == 0 {
		return "N/A"
	}
	growth := float64(current-previous) / float64(previous)
	return math.Round(growth*10000) / 100
}

type normalizedSeries struct {
	values []int64
	count  int
}

func normalizeSeries(values []int64, count int, dataFlag bool) normalizedSeries {
	if count == 0 && !dataFlag {
		for _, item := range values {
			if item == -1 {
				count++
				continue
			}
			dataFlag = true
			break
		}
	}
	if !dataFlag {
		return normalizedSeries{values: []int64{}, count: 0}
	}
	result := sliceInt64(values, count)
	for i := 0; i < len(result); i++ {
		if result[i] <= 0 && i > 0 {
			result[i] = result[i-1]
		}
	}
	return normalizedSeries{values: result, count: count}
}

func sliceInt64(values []int64, start int) []int64 {
	if start >= len(values) {
		return []int64{}
	}
	if start <= 0 {
		return append([]int64{}, values...)
	}
	return append([]int64{}, values[start:]...)
}

func sliceStrings(values []string, start int) []string {
	if start >= len(values) {
		return []string{}
	}
	if start <= 0 {
		return append([]string{}, values...)
	}
	return append([]string{}, values[start:]...)
}

func allZeroInt64(values []int64) bool {
	for _, v := range values {
		if v != 0 {
			return false
		}
	}
	return true
}
