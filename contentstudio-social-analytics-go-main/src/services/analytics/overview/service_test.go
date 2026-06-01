package overview

import (
	"context"
	"testing"
	"time"

	"github.com/rs/zerolog/log"

	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/overview"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/overview"
)

// stubRepo implements only the repo methods needed for service tests.
type stubRepo struct {
	platformDataFn    func(ctx context.Context, params *repo.OverviewParams) ([]repo.PlatformDataRow, error)
	accountDataFn     func(ctx context.Context, params *repo.OverviewParams) ([]repo.AccountDataRow, error)
	accountDetailedFn func(ctx context.Context, params *repo.OverviewParams) ([]repo.AccountDataDetailedRow, error)
	accountGraphsFn   func(ctx context.Context, params *repo.OverviewParams) ([]repo.AccountDataGraphsRow, error)
	topPostsFn        func(ctx context.Context, params *repo.OverviewParams) ([]repo.TopPostRow, error)
	topGraphFn        func(ctx context.Context, params *repo.OverviewParams) (*repo.TopPerformingGraphResult, error)
}

func (s *stubRepo) GetPlatformData(ctx context.Context, params *repo.OverviewParams) ([]repo.PlatformDataRow, error) {
	if s.platformDataFn != nil {
		return s.platformDataFn(ctx, params)
	}
	return nil, nil
}
func (s *stubRepo) GetAccountData(ctx context.Context, params *repo.OverviewParams) ([]repo.AccountDataRow, error) {
	if s.accountDataFn != nil {
		return s.accountDataFn(ctx, params)
	}
	return nil, nil
}
func (s *stubRepo) GetAccountDataDetailed(ctx context.Context, params *repo.OverviewParams) ([]repo.AccountDataDetailedRow, error) {
	if s.accountDetailedFn != nil {
		return s.accountDetailedFn(ctx, params)
	}
	return nil, nil
}
func (s *stubRepo) GetAccountDataGraphs(ctx context.Context, params *repo.OverviewParams) ([]repo.AccountDataGraphsRow, error) {
	if s.accountGraphsFn != nil {
		return s.accountGraphsFn(ctx, params)
	}
	return nil, nil
}
func (s *stubRepo) GetTopPosts(ctx context.Context, params *repo.OverviewParams) ([]repo.TopPostRow, error) {
	if s.topPostsFn != nil {
		return s.topPostsFn(ctx, params)
	}
	return nil, nil
}
func (s *stubRepo) GetTopPerformingGraph(ctx context.Context, params *repo.OverviewParams) (*repo.TopPerformingGraphResult, error) {
	if s.topGraphFn != nil {
		return s.topGraphFn(ctx, params)
	}
	return &repo.TopPerformingGraphResult{}, nil
}

func newService(r repoInterface) *OverviewAnalyticsService {
	return &OverviewAnalyticsService{repo: r, logger: log.Logger}
}

var baseReq = &types.OverviewRequest{
	WorkspaceID:      "ws1",
	StartDate:        "2025-01-01",
	EndDate:          "2025-01-31",
	FacebookAccounts: []string{"fb_1"},
}

func TestRound2(t *testing.T) {
	tests := []struct {
		in   float64
		want float64
	}{
		{0.555, 0.56},
		{1.234, 1.23},
		{0.0, 0.0},
		{100.999, 101.0},
	}
	for _, tc := range tests {
		got := round2(tc.in)
		if got != tc.want {
			t.Errorf("round2(%v) = %v, want %v", tc.in, got, tc.want)
		}
	}
}

// primaryStart is the CurrentStart value that NewOverviewParams produces for baseReq.
// baseReq uses "2025-01-01" which is a full month (Jan), so secondary = Dec 2024.
// After NewSecondaryParams(), CurrentStart = "2024-12-01".
// We use this to distinguish primary from secondary calls in stubs.
const primaryStart = "2025-01-01"

func isPrimary(params *repo.OverviewParams) bool {
	return params.CurrentStart == primaryStart
}

func TestGetSummary_NoPreviousData(t *testing.T) {
	svc := newService(&stubRepo{
		platformDataFn: func(_ context.Context, params *repo.OverviewParams) ([]repo.PlatformDataRow, error) {
			if isPrimary(params) {
				return []repo.PlatformDataRow{
					{Followers: 1000, TotalPosts: 50, Engagement: 200, Impressions: 2000, Reach: 1500},
				}, nil
			}
			return nil, nil
		},
	})

	resp, err := svc.GetSummary(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Summary.Followers != 1000 {
		t.Fatalf("expected followers=1000, got %d", resp.Summary.Followers)
	}
	if resp.Summary.Posts != 50 {
		t.Fatalf("expected posts=50, got %d", resp.Summary.Posts)
	}
	// No previous data → diffs and pct changes should be zero
	if resp.Summary.DiffFollowers != 0 {
		t.Fatalf("expected diff_followers=0, got %d", resp.Summary.DiffFollowers)
	}
	if resp.Summary.FollowersChangePct != 0 {
		t.Fatalf("expected followers_change_pct=0, got %v", resp.Summary.FollowersChangePct)
	}
}

func TestGetSummary_WithPreviousData(t *testing.T) {
	svc := newService(&stubRepo{
		platformDataFn: func(_ context.Context, params *repo.OverviewParams) ([]repo.PlatformDataRow, error) {
			if isPrimary(params) {
				return []repo.PlatformDataRow{
					{Followers: 1100, TotalPosts: 55, Engagement: 220, Impressions: 2200, Reach: 1600},
				}, nil
			}
			return []repo.PlatformDataRow{
				{Followers: 1000, TotalPosts: 50, Engagement: 200, Impressions: 2000, Reach: 1500},
			}, nil
		},
	})

	resp, err := svc.GetSummary(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Summary.DiffFollowers != 100 {
		t.Fatalf("expected diff_followers=100, got %d", resp.Summary.DiffFollowers)
	}
	if resp.Summary.FollowersChangePct != 10.0 {
		t.Fatalf("expected followers_change_pct=10, got %v", resp.Summary.FollowersChangePct)
	}
	// engagement_rate = engagement/impressions*100 = 220/2200*100 = 10
	if resp.Summary.EngagementRate != 10.0 {
		t.Fatalf("expected engagement_rate=10, got %v", resp.Summary.EngagementRate)
	}
}

func TestGetSummary_EngagementRateZeroImpressions(t *testing.T) {
	svc := newService(&stubRepo{
		platformDataFn: func(_ context.Context, _ *repo.OverviewParams) ([]repo.PlatformDataRow, error) {
			return []repo.PlatformDataRow{
				{Engagement: 100, Impressions: 0},
			}, nil
		},
	})

	resp, err := svc.GetSummary(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Summary.EngagementRate != 0 {
		t.Fatalf("expected engagement_rate=0 when impressions=0, got %v", resp.Summary.EngagementRate)
	}
}

func TestGetSummary_MultiPlatformAggregation(t *testing.T) {
	svc := newService(&stubRepo{
		platformDataFn: func(_ context.Context, _ *repo.OverviewParams) ([]repo.PlatformDataRow, error) {
			return []repo.PlatformDataRow{
				{Followers: 500, TotalPosts: 10, Engagement: 50, Impressions: 500, Reach: 400},
				{Followers: 300, TotalPosts: 20, Engagement: 80, Impressions: 800, Reach: 600},
			}, nil
		},
	})

	resp, err := svc.GetSummary(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Summary.Followers != 800 {
		t.Fatalf("expected followers=800 (500+300), got %d", resp.Summary.Followers)
	}
	if resp.Summary.Posts != 30 {
		t.Fatalf("expected posts=30 (10+20), got %d", resp.Summary.Posts)
	}
	if resp.Summary.Engagement != 130 {
		t.Fatalf("expected engagement=130, got %d", resp.Summary.Engagement)
	}
}

func TestGetTopPerformingGraph_FormatsBuckets(t *testing.T) {
	t1 := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	t2 := time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC)
	svc := newService(&stubRepo{
		topGraphFn: func(_ context.Context, _ *repo.OverviewParams) (*repo.TopPerformingGraphResult, error) {
			return &repo.TopPerformingGraphResult{
				Buckets:           []time.Time{t1, t2},
				FacebookPostCount: []float64{3, 5},
			}, nil
		},
	})

	resp, err := svc.GetTopPerformingGraph(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.Buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(resp.Buckets))
	}
	if resp.Buckets[0] != "2025-01-01" || resp.Buckets[1] != "2025-01-02" {
		t.Fatalf("unexpected bucket dates: %v", resp.Buckets)
	}
	if len(resp.FacebookPostCount) != 2 || resp.FacebookPostCount[1] != 5 {
		t.Fatalf("unexpected facebook_post_count: %v", resp.FacebookPostCount)
	}
}

func TestGetPlatformDataGrouped(t *testing.T) {
	svc := newService(&stubRepo{
		platformDataFn: func(_ context.Context, _ *repo.OverviewParams) ([]repo.PlatformDataRow, error) {
			return []repo.PlatformDataRow{
				{Followers: 100, TotalPosts: 5, PlatformType: "facebook"},
				{Followers: 200, TotalPosts: 10, PlatformType: "instagram"},
			}, nil
		},
	})

	rows, err := svc.GetPlatformDataGrouped(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(rows))
	}
	if rows[0].PlatformType != "facebook" || rows[0].Followers != 100 {
		t.Fatalf("unexpected row[0]: %+v", rows[0])
	}
}

func TestGetPlatformDataIndividual(t *testing.T) {
	svc := newService(&stubRepo{
		accountDataFn: func(_ context.Context, _ *repo.OverviewParams) ([]repo.AccountDataRow, error) {
			return []repo.AccountDataRow{
				{AccountID: "acc_1", Followers: 500, PlatformType: "facebook"},
			}, nil
		},
	})

	rows, err := svc.GetPlatformDataIndividual(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].AccountID != "acc_1" {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestGetPlatformDataDetailed(t *testing.T) {
	svc := newService(&stubRepo{
		accountDetailedFn: func(_ context.Context, _ *repo.OverviewParams) ([]repo.AccountDataDetailedRow, error) {
			return []repo.AccountDataDetailedRow{
				{AccountID: "acc_1", PlatformType: "instagram", CurrentFollowers: 1000, FollowersChangePct: 5.5},
			}, nil
		},
	})

	rows, err := svc.GetPlatformDataDetailed(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 || rows[0].FollowersChangePct != 5.5 {
		t.Fatalf("unexpected rows: %+v", rows)
	}
}

func TestGetPlatformDataGraphs_FormatsBuckets(t *testing.T) {
	t1 := time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC)
	svc := newService(&stubRepo{
		accountGraphsFn: func(_ context.Context, _ *repo.OverviewParams) ([]repo.AccountDataGraphsRow, error) {
			return []repo.AccountDataGraphsRow{
				{
					AccountID:  "acc_1",
					Engagement: []float64{10, 20},
					Buckets:    []time.Time{t1},
				},
			}, nil
		},
	})

	rows, err := svc.GetPlatformDataGraphs(context.Background(), baseReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if len(rows[0].Buckets) != 1 || rows[0].Buckets[0] != "2025-01-15" {
		t.Fatalf("unexpected buckets: %v", rows[0].Buckets)
	}
}

func TestGetTopPosts(t *testing.T) {
	createdAt := time.Date(2025, 1, 10, 12, 0, 0, 0, time.UTC)
	svc := newService(&stubRepo{
		topPostsFn: func(_ context.Context, _ *repo.OverviewParams) ([]repo.TopPostRow, error) {
			return []repo.TopPostRow{
				{
					PlatformType:    "facebook",
					AccountID:       "fb_1",
					PostID:          "post_1",
					TotalEngagement: 42,
					CreatedTime:     createdAt,
				},
			}, nil
		},
	})

	topReq := &types.TopPostsRequest{
		OverviewRequest: *baseReq,
		Type:            "total_engagement",
		Limit:           10,
	}
	topReq.Timezone = "America/New_York"
	rows, err := svc.GetTopPosts(context.Background(), topReq)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if rows[0].PostID != "post_1" || rows[0].TotalEngagement != 42 {
		t.Fatalf("unexpected row: %+v", rows[0])
	}
	if rows[0].CreatedTime != "2025-01-10T07:00:00-05:00" {
		t.Fatalf("unexpected created_time format: %s", rows[0].CreatedTime)
	}
}
