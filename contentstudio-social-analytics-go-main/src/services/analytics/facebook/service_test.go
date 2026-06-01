package facebook

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/facebook"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/facebook"
)

type stubRepo struct {
	getSummaryFn                   func(context.Context, *clickhouse.QueryParams) (*repo.SummaryResult, error)
	getAudienceGrowthFn            func(context.Context, *clickhouse.QueryParams) (*repo.AudienceGrowthResult, error)
	getLastFollowerCountsFn        func(context.Context, *clickhouse.QueryParams) (*repo.LastFollowerCounts, error)
	getAudienceGrowthRollupFn      func(context.Context, *clickhouse.QueryParams) (*repo.AudienceGrowthRollupResult, error)
	getPublishingBehaviourFn       func(context.Context, *clickhouse.QueryParams, []string) (*repo.PublishingBehaviourResult, error)
	getPublishingBehaviourRollupFn func(context.Context, *clickhouse.QueryParams) (*repo.PublishingRollupResult, error)
	getTopPostsFn                  func(context.Context, *clickhouse.QueryParams, []string, int, string) ([]repo.TopPostRow, error)
	getActiveUsersHoursFn          func(context.Context, *clickhouse.QueryParams) (*repo.ActiveUsersHoursResult, error)
	getActiveUsersDaysFn           func(context.Context, *clickhouse.QueryParams) (*repo.ActiveUsersDaysResult, error)
	getImpressionsFn               func(context.Context, *clickhouse.QueryParams) (*repo.ImpressionsResult, error)
	getImpressionsRollupFn         func(context.Context, *clickhouse.QueryParams) (*repo.ImpressionsRollupResult, error)
	getEngagementFn                func(context.Context, *clickhouse.QueryParams) (*repo.EngagementResult, error)
	getEngagementRollupFn          func(context.Context, *clickhouse.QueryParams) (*repo.EngagementRollupResult, error)
	getReelsAnalyticsFn            func(context.Context, *clickhouse.QueryParams) (*repo.ReelsAnalyticsResult, error)
	getReelsRollupFn               func(context.Context, *clickhouse.QueryParams) (*repo.ReelsRollupResult, error)
	getVideoInsightsFn             func(context.Context, *clickhouse.QueryParams) (*repo.VideoInsightsResult, error)
	getVideoRollupFn               func(context.Context, *clickhouse.QueryParams) (*repo.VideoRollupResult, error)
	getAudienceGenderFn            func(context.Context, *clickhouse.QueryParams) (*repo.AudienceGenderResult, error)
	getMaxGenderAgeFn              func(context.Context, *clickhouse.QueryParams) (*repo.MaxGenderAgeResult, error)
	getAudienceAgeFn               func(context.Context, *clickhouse.QueryParams) (*repo.AudienceAgeResult, error)
	getAudienceCountryFn           func(context.Context, *clickhouse.QueryParams) (map[string]int32, error)
	getAudienceCityFn              func(context.Context, *clickhouse.QueryParams) (map[string]int32, error)
}

func (s *stubRepo) GetSummary(ctx context.Context, params *clickhouse.QueryParams) (*repo.SummaryResult, error) {
	if s.getSummaryFn != nil {
		return s.getSummaryFn(ctx, params)
	}
	return &repo.SummaryResult{}, nil
}
func (s *stubRepo) GetPostsSummary(_ context.Context, _ *clickhouse.QueryParams) (*repo.PostsSummaryResult, error) {
	return &repo.PostsSummaryResult{}, nil
}
func (s *stubRepo) GetInsightsSummary(_ context.Context, _ *clickhouse.QueryParams) (*repo.InsightsSummaryResult, error) {
	return &repo.InsightsSummaryResult{}, nil
}
func (s *stubRepo) GetAudienceGrowth(ctx context.Context, params *clickhouse.QueryParams) (*repo.AudienceGrowthResult, error) {
	if s.getAudienceGrowthFn != nil {
		return s.getAudienceGrowthFn(ctx, params)
	}
	return &repo.AudienceGrowthResult{}, nil
}
func (s *stubRepo) GetLastFollowerCounts(ctx context.Context, params *clickhouse.QueryParams) (*repo.LastFollowerCounts, error) {
	if s.getLastFollowerCountsFn != nil {
		return s.getLastFollowerCountsFn(ctx, params)
	}
	return &repo.LastFollowerCounts{}, nil
}
func (s *stubRepo) GetAudienceGrowthRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.AudienceGrowthRollupResult, error) {
	if s.getAudienceGrowthRollupFn != nil {
		return s.getAudienceGrowthRollupFn(ctx, params)
	}
	return &repo.AudienceGrowthRollupResult{}, nil
}
func (s *stubRepo) GetPublishingBehaviour(ctx context.Context, params *clickhouse.QueryParams, mediaTypes []string) (*repo.PublishingBehaviourResult, error) {
	if s.getPublishingBehaviourFn != nil {
		return s.getPublishingBehaviourFn(ctx, params, mediaTypes)
	}
	return &repo.PublishingBehaviourResult{}, nil
}
func (s *stubRepo) GetPublishingBehaviourRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.PublishingRollupResult, error) {
	if s.getPublishingBehaviourRollupFn != nil {
		return s.getPublishingBehaviourRollupFn(ctx, params)
	}
	return &repo.PublishingRollupResult{}, nil
}
func (s *stubRepo) GetTopPosts(ctx context.Context, params *clickhouse.QueryParams, mediaTypes []string, limit int, orderBy string) ([]repo.TopPostRow, error) {
	if s.getTopPostsFn != nil {
		return s.getTopPostsFn(ctx, params, mediaTypes, limit, orderBy)
	}
	return nil, nil
}
func (s *stubRepo) GetActiveUsersHours(ctx context.Context, params *clickhouse.QueryParams) (*repo.ActiveUsersHoursResult, error) {
	if s.getActiveUsersHoursFn != nil {
		return s.getActiveUsersHoursFn(ctx, params)
	}
	return &repo.ActiveUsersHoursResult{}, nil
}
func (s *stubRepo) GetActiveUsersDays(ctx context.Context, params *clickhouse.QueryParams) (*repo.ActiveUsersDaysResult, error) {
	if s.getActiveUsersDaysFn != nil {
		return s.getActiveUsersDaysFn(ctx, params)
	}
	return &repo.ActiveUsersDaysResult{}, nil
}
func (s *stubRepo) GetImpressions(ctx context.Context, params *clickhouse.QueryParams) (*repo.ImpressionsResult, error) {
	if s.getImpressionsFn != nil {
		return s.getImpressionsFn(ctx, params)
	}
	return &repo.ImpressionsResult{}, nil
}
func (s *stubRepo) GetImpressionsRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.ImpressionsRollupResult, error) {
	if s.getImpressionsRollupFn != nil {
		return s.getImpressionsRollupFn(ctx, params)
	}
	return &repo.ImpressionsRollupResult{}, nil
}
func (s *stubRepo) GetEngagement(ctx context.Context, params *clickhouse.QueryParams) (*repo.EngagementResult, error) {
	if s.getEngagementFn != nil {
		return s.getEngagementFn(ctx, params)
	}
	return &repo.EngagementResult{}, nil
}
func (s *stubRepo) GetEngagementRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.EngagementRollupResult, error) {
	if s.getEngagementRollupFn != nil {
		return s.getEngagementRollupFn(ctx, params)
	}
	return &repo.EngagementRollupResult{}, nil
}
func (s *stubRepo) GetReelsAnalytics(ctx context.Context, params *clickhouse.QueryParams) (*repo.ReelsAnalyticsResult, error) {
	if s.getReelsAnalyticsFn != nil {
		return s.getReelsAnalyticsFn(ctx, params)
	}
	return &repo.ReelsAnalyticsResult{}, nil
}
func (s *stubRepo) GetReelsRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.ReelsRollupResult, error) {
	if s.getReelsRollupFn != nil {
		return s.getReelsRollupFn(ctx, params)
	}
	return &repo.ReelsRollupResult{}, nil
}
func (s *stubRepo) GetVideoInsights(ctx context.Context, params *clickhouse.QueryParams) (*repo.VideoInsightsResult, error) {
	if s.getVideoInsightsFn != nil {
		return s.getVideoInsightsFn(ctx, params)
	}
	return &repo.VideoInsightsResult{}, nil
}
func (s *stubRepo) GetVideoRollup(ctx context.Context, params *clickhouse.QueryParams) (*repo.VideoRollupResult, error) {
	if s.getVideoRollupFn != nil {
		return s.getVideoRollupFn(ctx, params)
	}
	return &repo.VideoRollupResult{}, nil
}
func (s *stubRepo) GetAudienceGender(ctx context.Context, params *clickhouse.QueryParams) (*repo.AudienceGenderResult, error) {
	if s.getAudienceGenderFn != nil {
		return s.getAudienceGenderFn(ctx, params)
	}
	return &repo.AudienceGenderResult{}, nil
}
func (s *stubRepo) GetMaxGenderAge(ctx context.Context, params *clickhouse.QueryParams) (*repo.MaxGenderAgeResult, error) {
	if s.getMaxGenderAgeFn != nil {
		return s.getMaxGenderAgeFn(ctx, params)
	}
	return &repo.MaxGenderAgeResult{}, nil
}
func (s *stubRepo) GetAudienceAge(ctx context.Context, params *clickhouse.QueryParams) (*repo.AudienceAgeResult, error) {
	if s.getAudienceAgeFn != nil {
		return s.getAudienceAgeFn(ctx, params)
	}
	return &repo.AudienceAgeResult{}, nil
}
func (s *stubRepo) GetAudienceCountry(ctx context.Context, params *clickhouse.QueryParams) (map[string]int32, error) {
	if s.getAudienceCountryFn != nil {
		return s.getAudienceCountryFn(ctx, params)
	}
	return map[string]int32{}, nil
}
func (s *stubRepo) GetAudienceCity(ctx context.Context, params *clickhouse.QueryParams) (map[string]int32, error) {
	if s.getAudienceCityFn != nil {
		return s.getAudienceCityFn(ctx, params)
	}
	return map[string]int32{}, nil
}

func newTestService(repo Repository) *FacebookAnalyticsService {
	return NewFacebookAnalyticsService(repo, zerolog.New(io.Discard))
}

func validRequest() *types.FacebookRequest {
	return &types.FacebookRequest{
		WorkspaceID: "ws1",
		FacebookIDs: []string{"fb_123"},
		StartDate:   "2025-01-01",
		EndDate:     "2025-01-31",
		Timezone:    "UTC",
	}
}

func TestGetAudienceGrowth_FillsInitialZerosFromHistoricalCounts(t *testing.T) {
	svc := newTestService(&stubRepo{
		getAudienceGrowthFn: func(context.Context, *clickhouse.QueryParams) (*repo.AudienceGrowthResult, error) {
			return &repo.AudienceGrowthResult{
				ShowData:         1,
				FanCount:         []int32{0, 0, 25},
				PageFansDaily:    []int32{0, 0, 5},
				PageFansByLike:   []int32{0, 0, 2},
				PageFansByUnlike: []int32{0, 0, 1},
				PageImpressions:  []int32{10, 20, 30},
				PageEngagements:  []int32{1, 2, 3},
				Buckets:          []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
			}, nil
		},
		getLastFollowerCountsFn: func(context.Context, *clickhouse.QueryParams) (*repo.LastFollowerCounts, error) {
			return &repo.LastFollowerCounts{
				PageFans:         99,
				PageFansByLike:   7,
				PageFansByUnlike: 3,
			}, nil
		},
	})

	resp, err := svc.GetAudienceGrowth(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := resp.AudienceGrowth.FanCount[0]; got != 99 {
		t.Fatalf("expected first fan_count to be backfilled, got %d", got)
	}
	if got := resp.AudienceGrowth.PageFansByLike[0]; got != 7 {
		t.Fatalf("expected first page_fans_by_like to be backfilled, got %d", got)
	}
}

func TestGetTopPosts_GroupsMediaAssetsByPostID(t *testing.T) {
	svc := newTestService(&stubRepo{
		getTopPostsFn: func(context.Context, *clickhouse.QueryParams, []string, int, string) ([]repo.TopPostRow, error) {
			return []repo.TopPostRow{
				{
					PageID:       "fb_123",
					PostID:       "post_1",
					MediaID:      "asset_1",
					AssetType:    "image",
					MediaCaption: "caption-1",
					CreatedTime:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				},
				{
					PageID:       "fb_123",
					PostID:       "post_1",
					MediaID:      "asset_2",
					AssetType:    "image",
					MediaCaption: "caption-2",
					CreatedTime:  time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC),
				},
			}, nil
		},
	})

	req := validRequest()
	req.Timezone = "America/New_York"

	resp, err := svc.GetTopPosts(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.TopPosts) != 1 {
		t.Fatalf("expected one grouped post, got %d", len(resp.TopPosts))
	}
	if len(resp.TopPosts[0].MediaAssets) != 2 {
		t.Fatalf("expected two grouped media assets, got %d", len(resp.TopPosts[0].MediaAssets))
	}
	if resp.TopPosts[0].CreatedTime != "2025-01-01T07:00:00-05:00" {
		t.Fatalf("expected timezone-adjusted created_time, got %q", resp.TopPosts[0].CreatedTime)
	}
}

func TestGetActiveUsers_AdjustsBucketsForTimezone(t *testing.T) {
	svc := newTestService(&stubRepo{
		getActiveUsersHoursFn: func(context.Context, *clickhouse.QueryParams) (*repo.ActiveUsersHoursResult, error) {
			return &repo.ActiveUsersHoursResult{
				Buckets:      []int32{0, 23},
				Values:       []int32{10, 20},
				HighestValue: 20,
				HighestHour:  23,
			}, nil
		},
		getActiveUsersDaysFn: func(context.Context, *clickhouse.QueryParams) (*repo.ActiveUsersDaysResult, error) {
			return &repo.ActiveUsersDaysResult{
				Buckets: []string{"Monday"},
				Values:  []int32{5},
			}, nil
		},
	})

	resp, err := svc.GetActiveUsers(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.ActiveUsers.ActiveUsersHours.Buckets) != 2 {
		t.Fatalf("expected hourly buckets, got %d", len(resp.ActiveUsers.ActiveUsersHours.Buckets))
	}
	if got := resp.ActiveUsers.ActiveUsersHours.HighestHour; got != 7 {
		t.Fatalf("expected highest hour to shift to 7 for UTC (+8), got %d", got)
	}
}

func TestGetActiveUsers_ReturnsErrorWhenHourlyQueryFails(t *testing.T) {
	expectedErr := errors.New("hourly query failed")
	svc := newTestService(&stubRepo{
		getActiveUsersHoursFn: func(context.Context, *clickhouse.QueryParams) (*repo.ActiveUsersHoursResult, error) {
			return nil, expectedErr
		},
		getActiveUsersDaysFn: func(context.Context, *clickhouse.QueryParams) (*repo.ActiveUsersDaysResult, error) {
			return &repo.ActiveUsersDaysResult{}, nil
		},
	})

	_, err := svc.GetActiveUsers(context.Background(), validRequest())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected hourly query error, got %v", err)
	}
}

func TestGetAudienceLocation_ReturnsErrorWhenCountryQueryFails(t *testing.T) {
	expectedErr := errors.New("country query failed")
	svc := newTestService(&stubRepo{
		getAudienceCountryFn: func(context.Context, *clickhouse.QueryParams) (map[string]int32, error) {
			return nil, expectedErr
		},
		getAudienceCityFn: func(context.Context, *clickhouse.QueryParams) (map[string]int32, error) {
			return map[string]int32{"Lahore": 3}, nil
		},
	})

	_, err := svc.GetAudienceLocation(context.Background(), validRequest())
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected country query error, got %v", err)
	}
}
