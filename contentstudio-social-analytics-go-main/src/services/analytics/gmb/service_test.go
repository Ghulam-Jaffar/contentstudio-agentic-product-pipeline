package gmb

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/gmb"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/gmb"
)

// --- Mock ClickHouse infrastructure ---

type mockRow struct{}

func (m *mockRow) Err() error { return nil }
func (m *mockRow) Scan(dest ...any) error {
	for _, d := range dest {
		switch v := d.(type) {
		case *int64:
			*v = 0
		case *float64:
			*v = 0
		case *string:
			*v = ""
		case *[]int64:
			*v = []int64{}
		case *[]time.Time:
			*v = []time.Time{}
		}
	}
	return nil
}
func (m *mockRow) ScanStruct(dest any) error { return nil }

type mockRows struct{}

func (m *mockRows) Next() bool                       { return false }
func (m *mockRows) Scan(dest ...any) error           { return nil }
func (m *mockRows) ScanStruct(dest any) error        { return nil }
func (m *mockRows) ColumnTypes() []driver.ColumnType { return nil }
func (m *mockRows) Totals(dest ...any) error         { return nil }
func (m *mockRows) Columns() []string                { return nil }
func (m *mockRows) Close() error                     { return nil }
func (m *mockRows) Err() error                       { return nil }

type mockConn struct{}

func (m *mockConn) Contributors() []string                        { return nil }
func (m *mockConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (m *mockConn) Select(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *mockConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	return &mockRows{}, nil
}
func (m *mockConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	return &mockRow{}
}
func (m *mockConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return &mockBatch{}, nil
}
func (m *mockConn) Exec(ctx context.Context, query string, args ...any) error { return nil }
func (m *mockConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return nil
}
func (m *mockConn) Ping(ctx context.Context) error { return nil }
func (m *mockConn) Stats() driver.Stats            { return driver.Stats{} }
func (m *mockConn) Close() error                   { return nil }

type mockBatch struct{}

func (m *mockBatch) Abort() error                  { return nil }
func (m *mockBatch) Append(v ...any) error         { return nil }
func (m *mockBatch) AppendStruct(v any) error      { return nil }
func (m *mockBatch) Column(int) driver.BatchColumn { return nil }
func (m *mockBatch) Columns() []column.Interface   { return nil }
func (m *mockBatch) Flush() error                  { return nil }
func (m *mockBatch) Send() error                   { return nil }
func (m *mockBatch) IsSent() bool                  { return false }
func (m *mockBatch) Rows() int                     { return 0 }
func (m *mockBatch) Close() error                  { return nil }

func newTestService() *GMBAnalyticsService {
	client := &ch.Client{
		Conn:   &mockConn{},
		Config: config.ClickHouseConfig{Database: "test_db"},
		Logger: zerolog.New(io.Discard),
	}
	r := repo.NewRepository(client)
	return NewGMBAnalyticsService(r, zerolog.New(io.Discard))
}

func validRequest() *types.GMBRequest {
	return &types.GMBRequest{
		WorkspaceID: "ws1",
		GmbID:       "loc_123",
		StartDate:   "2025-01-01",
		EndDate:     "2025-01-31",
		Timezone:    "UTC",
	}
}

// --- Service method tests ---

func TestNewGMBAnalyticsService(t *testing.T) {
	svc := newTestService()
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestGetSummary(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.GMBRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.SummaryResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.GMBRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns current and previous",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.SummaryResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.Overview["current"] == nil {
					t.Fatal("expected 'current' in overview")
				}
				if resp.Overview["previous"] == nil {
					t.Fatal("expected 'previous' in overview")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetSummary(context.Background(), tc.req)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkResp != nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestGetImpressions(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.GMBRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.ImpressionsResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.GMBRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns impressions and rollup",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.ImpressionsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.Impressions == nil {
					t.Fatal("expected non-nil impressions")
				}
				if resp.ImpressionsRolup["current"] == nil {
					t.Fatal("expected 'current' in rollup")
				}
				if resp.ImpressionsRolup["previous"] == nil {
					t.Fatal("expected 'previous' in rollup")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetImpressions(context.Background(), tc.req)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkResp != nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestGetActions(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.GMBRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.ActionsResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.GMBRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns actions and rollup",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.ActionsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.Actions == nil {
					t.Fatal("expected non-nil actions")
				}
				if resp.ActionsRollup["current"] == nil {
					t.Fatal("expected 'current' in rollup")
				}
				if resp.ActionsRollup["previous"] == nil {
					t.Fatal("expected 'previous' in rollup")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetActions(context.Background(), tc.req)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkResp != nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestGetSearchKeywords(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.SearchKeywordsRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.SearchKeywordsResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.SearchKeywordsRequest{},
			expectErr: true,
		},
		{
			name: "valid request with defaults",
			req:  &types.SearchKeywordsRequest{GMBRequest: *validRequest()},
			checkResp: func(t *testing.T, resp *types.SearchKeywordsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.Keywords == nil {
					t.Fatal("expected non-nil keywords")
				}
			},
		},
		{
			name: "valid request with custom limit",
			req:  &types.SearchKeywordsRequest{GMBRequest: *validRequest(), Limit: 10},
			checkResp: func(t *testing.T, resp *types.SearchKeywordsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetSearchKeywords(context.Background(), tc.req)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkResp != nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestGetTopPosts(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.TopPostsRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.TopPostsResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.TopPostsRequest{},
			expectErr: true,
		},
		{
			name: "valid request with defaults",
			req:  &types.TopPostsRequest{GMBRequest: *validRequest()},
			checkResp: func(t *testing.T, resp *types.TopPostsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.Posts == nil {
					t.Fatal("expected non-nil posts")
				}
			},
		},
		{
			name: "valid request with custom limit and order",
			req: &types.TopPostsRequest{
				GMBRequest: *validRequest(),
				Limit:      5,
				OrderBy:    "created_at",
			},
			checkResp: func(t *testing.T, resp *types.TopPostsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
			},
		},
		{
			name: "invalid order_by defaults to created_at",
			req: &types.TopPostsRequest{
				GMBRequest: *validRequest(),
				OrderBy:    "invalid_field",
			},
			checkResp: func(t *testing.T, resp *types.TopPostsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetTopPosts(context.Background(), tc.req)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkResp != nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestGetPublishingBehavior(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.GMBRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.PublishingBehaviorResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.GMBRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns publishing data",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.PublishingBehaviorResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.PublishingBehaviour == nil {
					t.Fatal("expected non-nil publishing_behaviour")
				}
				if resp.PublishingBehaviour.Buckets == nil {
					t.Fatal("expected non-nil buckets")
				}
				if resp.PublishingBehaviour.PostCount == nil {
					t.Fatal("expected non-nil post_count")
				}
				if resp.PublishingBehaviour.TopicTypes == nil {
					t.Fatal("expected non-nil topic_types")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetPublishingBehavior(context.Background(), tc.req)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkResp != nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestGetReviews(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.GMBRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.ReviewsResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.GMBRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns reviews and rollup",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.ReviewsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.Reviews == nil {
					t.Fatal("expected non-nil reviews")
				}
				if resp.Reviews.StarDistribution == nil {
					t.Fatal("expected non-nil star_distribution")
				}
				for _, star := range []string{"1", "2", "3", "4", "5"} {
					if _, ok := resp.Reviews.StarDistribution[star]; !ok {
						t.Fatalf("expected star %q in distribution", star)
					}
				}
				if resp.ReviewsRollup["current"] == nil {
					t.Fatal("expected 'current' in rollup")
				}
				if resp.ReviewsRollup["previous"] == nil {
					t.Fatal("expected 'previous' in rollup")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetReviews(context.Background(), tc.req)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkResp != nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

func TestGetMediaActivity(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.GMBRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.MediaActivityResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.GMBRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns media activity and rollup",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.MediaActivityResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.MediaActivity == nil {
					t.Fatal("expected non-nil media_activity")
				}
				if resp.MediaActivityRollup["current"] == nil {
					t.Fatal("expected 'current' in rollup")
				}
				if resp.MediaActivityRollup["previous"] == nil {
					t.Fatal("expected 'previous' in rollup")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetMediaActivity(context.Background(), tc.req)
			if tc.expectErr {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.checkResp != nil {
				tc.checkResp(t, resp)
			}
		})
	}
}

// --- Private helper tests ---

func TestPrevPeriodParams(t *testing.T) {
	svc := newTestService()
	params := &ch.QueryParams{
		AccountIDs:   []string{"loc_123"},
		DateFrom:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:       time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		PrevDateFrom: time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC),
		PrevDateTo:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		Timezone:     "UTC",
		DayCount:     31,
	}

	prev := svc.prevPeriodParams(params)

	if prev.AccountIDs[0] != "loc_123" {
		t.Fatalf("expected loc_123, got %q", prev.AccountIDs[0])
	}
	if prev.DateFrom != time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("expected prev date from 2024-12-02, got %v", prev.DateFrom)
	}
	if prev.DateTo != time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC) {
		t.Fatalf("expected prev date to 2024-12-31, got %v", prev.DateTo)
	}
	if prev.Timezone != "UTC" {
		t.Fatalf("expected UTC, got %q", prev.Timezone)
	}
	if prev.DayCount != 31 {
		t.Fatalf("expected 31, got %d", prev.DayCount)
	}
}

func TestMapSummary(t *testing.T) {
	r := &repo.SummaryResult{
		TotalImpressions: 1000, SearchImpressions: 400, MapsImpressions: 600,
		WebsiteClicks: 150, CallClicks: 80, DirectionRequests: 40,
		OtherActions: 20, TotalReviews: 25, AverageRating: 4.5, TotalPosts: 12,
	}
	result := mapSummary(r)

	if result.TotalImpressions != 1000 {
		t.Fatalf("expected 1000, got %d", result.TotalImpressions)
	}
	if result.AverageRating != 4.5 {
		t.Fatalf("expected 4.5, got %f", result.AverageRating)
	}
	if result.TotalPosts != 12 {
		t.Fatalf("expected 12, got %d", result.TotalPosts)
	}
}

func TestMapImpressions(t *testing.T) {
	r := &repo.ImpressionsResult{
		DesktopMapsDaily:      []int64{10, 20, 30},
		DesktopSearchDaily:    []int64{5, 10, 15},
		MobileMapsDaily:       []int64{3, 6, 9},
		MobileSearchDaily:     []int64{2, 4, 6},
		TotalImpressionsDaily: []int64{20, 40, 60},
		ShowData:              120,
		Buckets: []time.Time{
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
		},
	}
	result := mapImpressions(r)

	// cumulative sums
	if result.DesktopMaps[2] != 60 {
		t.Fatalf("expected cumulative desktop_maps[2]=60, got %d", result.DesktopMaps[2])
	}
	// daily values preserved
	if result.DesktopMapsDaily[1] != 20 {
		t.Fatalf("expected daily desktop_maps[1]=20, got %d", result.DesktopMapsDaily[1])
	}
	if result.ShowData != 120 {
		t.Fatalf("expected show_data 120, got %d", result.ShowData)
	}
	if len(result.Buckets) != 3 {
		t.Fatalf("expected 3 buckets, got %d", len(result.Buckets))
	}
	if result.Buckets[0] != "2025-01-01" {
		t.Fatalf("expected 2025-01-01, got %q", result.Buckets[0])
	}
}

func TestMapImpressions_NilSlices(t *testing.T) {
	r := &repo.ImpressionsResult{}
	result := mapImpressions(r)

	if result.DesktopMaps == nil {
		t.Fatal("expected non-nil desktop_maps (should be empty slice)")
	}
	if len(result.DesktopMaps) != 0 {
		t.Fatalf("expected 0 length, got %d", len(result.DesktopMaps))
	}
	if len(result.Buckets) != 0 {
		t.Fatalf("expected 0 buckets, got %d", len(result.Buckets))
	}
}

func TestMapActions(t *testing.T) {
	r := &repo.ActionsResult{
		CallClicksDaily:        []int64{5, 10},
		WebsiteClicksDaily:     []int64{20, 30},
		DirectionRequestsDaily: []int64{3, 7},
		OtherActionsDaily:      []int64{1, 2},
		ShowData:               78,
		Buckets: []time.Time{
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
		},
	}
	result := mapActions(r)

	// cumulative sums
	if result.CallClicks[1] != 15 {
		t.Fatalf("expected cumulative call_clicks[1]=15, got %d", result.CallClicks[1])
	}
	if result.WebsiteClicks[1] != 50 {
		t.Fatalf("expected cumulative website_clicks[1]=50, got %d", result.WebsiteClicks[1])
	}
	// daily preserved
	if result.CallClicksDaily[0] != 5 {
		t.Fatalf("expected daily call_clicks[0]=5, got %d", result.CallClicksDaily[0])
	}
	if result.ShowData != 78 {
		t.Fatalf("expected show_data 78, got %d", result.ShowData)
	}
}

func TestMapImpressionsRollup(t *testing.T) {
	r := &repo.ImpressionsRollupResult{
		TotalImpressions: 500, DesktopMaps: 100, DesktopSearch: 150,
		MobileMaps: 120, MobileSearch: 130, AvgImpressions: 16.1,
	}
	result := mapImpressionsRollup(r)

	if result.TotalImpressions != 500 {
		t.Fatalf("expected 500, got %d", result.TotalImpressions)
	}
	if result.AvgImpressions != 16.1 {
		t.Fatalf("expected 16.1, got %f", result.AvgImpressions)
	}
}

func TestMapActionsRollup(t *testing.T) {
	r := &repo.ActionsRollupResult{
		TotalCallClicks: 50, TotalWebsiteClicks: 200,
		TotalDirectionRequests: 30, TotalOtherActions: 10, AvgActions: 9.35,
	}
	result := mapActionsRollup(r)

	if result.TotalCallClicks != 50 {
		t.Fatalf("expected 50, got %d", result.TotalCallClicks)
	}
	if result.AvgActions != 9.35 {
		t.Fatalf("expected 9.35, got %f", result.AvgActions)
	}
}

func TestMapSearchKeywords(t *testing.T) {
	rows := []repo.SearchKeywordRow{
		{Keyword: "pizza", ImpressionsValue: 1500, ImpressionsThreshold: 0, KeywordMonth: time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)},
		{Keyword: "restaurant", ImpressionsValue: 800, ImpressionsThreshold: 1, KeywordMonth: time.Date(2025, 2, 1, 0, 0, 0, 0, time.UTC)},
	}
	result := mapSearchKeywords(rows)

	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[0].Keyword != "pizza" {
		t.Fatalf("expected pizza, got %q", result[0].Keyword)
	}
	if result[0].KeywordMonth != "2025-01" {
		t.Fatalf("expected 2025-01, got %q", result[0].KeywordMonth)
	}
	if result[1].KeywordMonth != "2025-02" {
		t.Fatalf("expected 2025-02, got %q", result[1].KeywordMonth)
	}
}

func TestMapSearchKeywords_Nil(t *testing.T) {
	result := mapSearchKeywords(nil)
	if result == nil {
		t.Fatal("expected non-nil (empty slice)")
	}
	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestMapTopPosts(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 12, 30, 0, 0, time.UTC)
	rows := []repo.TopPostRow{
		{
			PostName: "post_1", State: "LIVE", TopicType: "STANDARD",
			SearchURL:  "https://search.google.com/local/posts",
			MediaNames: []string{"m1"}, MediaFormats: []string{"PHOTO"},
			MediaGoogleURLs: []string{"https://lh3.google.com/m1"},
			CreatedAt:       createdAt,
		},
	}
	result := mapTopPosts(rows, "America/New_York")

	if len(result) != 1 {
		t.Fatalf("expected 1, got %d", len(result))
	}
	if result[0].PostName != "post_1" {
		t.Fatalf("expected post_1, got %q", result[0].PostName)
	}
	if result[0].CreatedAt != "2025-01-15T07:30:00-05:00" {
		t.Fatalf("expected timezone-adjusted RFC3339 format, got %q", result[0].CreatedAt)
	}
	if len(result[0].MediaNames) != 1 {
		t.Fatalf("expected 1 media name, got %d", len(result[0].MediaNames))
	}
}

func TestMapTopPosts_Nil(t *testing.T) {
	result := mapTopPosts(nil, "UTC")
	if result == nil {
		t.Fatal("expected non-nil (empty slice)")
	}
	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestMapTopPosts_NilMediaSlices(t *testing.T) {
	rows := []repo.TopPostRow{
		{PostName: "post_1", CreatedAt: time.Now()},
	}
	result := mapTopPosts(rows, "UTC")

	if result[0].MediaNames == nil {
		t.Fatal("expected non-nil media_names")
	}
	if result[0].MediaFormats == nil {
		t.Fatal("expected non-nil media_formats")
	}
	if result[0].MediaGoogleURLs == nil {
		t.Fatal("expected non-nil media_google_urls")
	}
}

func TestMapPublishingBehavior(t *testing.T) {
	pub := &repo.PublishingResult{
		PostCount: []int64{1, 0, 2},
		Buckets: []time.Time{
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
		},
	}
	topicTypes := []repo.TopicTypeRow{
		{TopicType: "STANDARD", Count: 8},
		{TopicType: "EVENT", Count: 3},
	}

	result := mapPublishingBehavior(pub, topicTypes)

	if len(result.PostCount) != 3 {
		t.Fatalf("expected 3 post counts, got %d", len(result.PostCount))
	}
	if len(result.TopicTypes) != 2 {
		t.Fatalf("expected 2 topic types, got %d", len(result.TopicTypes))
	}
	if result.TopicTypes[0].Name != "STANDARD" {
		t.Fatalf("expected STANDARD, got %q", result.TopicTypes[0].Name)
	}
	if result.TopicTypes[0].Count != 8 {
		t.Fatalf("expected 8, got %d", result.TopicTypes[0].Count)
	}
}

func TestMapPublishingBehavior_NilTopicTypes(t *testing.T) {
	pub := &repo.PublishingResult{}
	result := mapPublishingBehavior(pub, nil)

	if result.TopicTypes == nil {
		t.Fatal("expected non-nil topic_types")
	}
	if len(result.TopicTypes) != 0 {
		t.Fatalf("expected 0, got %d", len(result.TopicTypes))
	}
}

func TestMapReviews(t *testing.T) {
	summary := &repo.ReviewsSummaryResult{
		AvgRating: 4.2, TotalReviews: 25,
		Star1: 1, Star2: 2, Star3: 3, Star4: 8, Star5: 11,
	}
	ts := &repo.ReviewsTimeSeriesResult{
		DailyReviews: []int64{2, 0, 1},
		Buckets: []time.Time{
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
		},
	}
	list := []repo.ReviewRow{
		{
			ReviewID: "r1", ReviewerDisplayName: "John", ReviewerProfilePhotoURL: "url",
			StarRating: 5, Comment: "Great!", ReplyComment: "Thanks!",
			CreatedAt: time.Date(2025, 1, 10, 14, 0, 0, 0, time.UTC),
		},
	}

	result := mapReviews(summary, ts, list)

	if result.AvgRating != 4.2 {
		t.Fatalf("expected 4.2, got %f", result.AvgRating)
	}
	if result.TotalReviews != 25 {
		t.Fatalf("expected 25, got %d", result.TotalReviews)
	}
	if result.StarDistribution["5"] != 11 {
		t.Fatalf("expected star 5 = 11, got %d", result.StarDistribution["5"])
	}
	if result.StarDistribution["1"] != 1 {
		t.Fatalf("expected star 1 = 1, got %d", result.StarDistribution["1"])
	}
	if len(result.DailyReviews) != 3 {
		t.Fatalf("expected 3 daily reviews, got %d", len(result.DailyReviews))
	}
	if len(result.ReviewsList) != 1 {
		t.Fatalf("expected 1 review item, got %d", len(result.ReviewsList))
	}
	if result.ReviewsList[0].ReviewID != "r1" {
		t.Fatalf("expected r1, got %q", result.ReviewsList[0].ReviewID)
	}
	if result.ReviewsList[0].CreatedAt != "2025-01-10T14:00:00Z" {
		t.Fatalf("expected RFC3339 format, got %q", result.ReviewsList[0].CreatedAt)
	}
}

func TestMapReviews_EmptyList(t *testing.T) {
	summary := &repo.ReviewsSummaryResult{}
	ts := &repo.ReviewsTimeSeriesResult{}
	result := mapReviews(summary, ts, nil)

	if result.ReviewsList == nil {
		t.Fatal("expected non-nil reviews_list")
	}
	if len(result.ReviewsList) != 0 {
		t.Fatalf("expected 0, got %d", len(result.ReviewsList))
	}
}

func TestMapReviewsRollup(t *testing.T) {
	r := &repo.ReviewsRollupResult{TotalReviews: 25, AvgRating: 4.2}
	result := mapReviewsRollup(r)

	if result.TotalReviews != 25 {
		t.Fatalf("expected 25, got %d", result.TotalReviews)
	}
	if result.AvgRating != 4.2 {
		t.Fatalf("expected 4.2, got %f", result.AvgRating)
	}
}

func TestMapMediaActivity(t *testing.T) {
	r := &repo.MediaResult{
		PhotoCountDaily: []int64{3, 5, 2},
		VideoCountDaily: []int64{1, 0, 1},
		ShowData:        12,
		Buckets: []time.Time{
			time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
			time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
		},
	}
	result := mapMediaActivity(r)

	// cumulative photos: 3, 8, 10
	if result.PhotoCount[2] != 10 {
		t.Fatalf("expected cumulative photo_count[2]=10, got %d", result.PhotoCount[2])
	}
	// cumulative videos: 1, 1, 2
	if result.VideoCount[2] != 2 {
		t.Fatalf("expected cumulative video_count[2]=2, got %d", result.VideoCount[2])
	}
	// daily preserved
	if result.PhotoCountDaily[1] != 5 {
		t.Fatalf("expected daily photo_count[1]=5, got %d", result.PhotoCountDaily[1])
	}
	if result.ShowData != 12 {
		t.Fatalf("expected 12, got %d", result.ShowData)
	}
}

func TestMapMediaActivityRollup(t *testing.T) {
	r := &repo.MediaRollupResult{TotalPhotos: 45, TotalVideos: 8, AvgMedia: 1.71}
	result := mapMediaActivityRollup(r)

	if result.TotalPhotos != 45 {
		t.Fatalf("expected 45, got %d", result.TotalPhotos)
	}
	if result.TotalVideos != 8 {
		t.Fatalf("expected 8, got %d", result.TotalVideos)
	}
	if result.AvgMedia != 1.71 {
		t.Fatalf("expected 1.71, got %f", result.AvgMedia)
	}
}

// --- Utility function tests ---

func TestFormatBuckets(t *testing.T) {
	times := []time.Time{
		time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
		time.Date(2025, 12, 31, 0, 0, 0, 0, time.UTC),
	}
	result := formatBuckets(times)

	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	if result[0] != "2025-01-01" {
		t.Fatalf("expected 2025-01-01, got %q", result[0])
	}
	if result[2] != "2025-12-31" {
		t.Fatalf("expected 2025-12-31, got %q", result[2])
	}
}

func TestFormatBuckets_Nil(t *testing.T) {
	result := formatBuckets(nil)
	if result == nil {
		t.Fatal("expected non-nil (empty slice)")
	}
	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestFormatBuckets_Empty(t *testing.T) {
	result := formatBuckets([]time.Time{})
	if len(result) != 0 {
		t.Fatalf("expected 0, got %d", len(result))
	}
}

func TestCumSum(t *testing.T) {
	tests := []struct {
		name     string
		input    []int64
		expected []int64
	}{
		{name: "nil", input: nil, expected: nil},
		{name: "empty", input: []int64{}, expected: []int64{}},
		{name: "single", input: []int64{5}, expected: []int64{5}},
		{name: "multiple", input: []int64{10, 20, 30}, expected: []int64{10, 30, 60}},
		{name: "with zeros", input: []int64{5, 0, 0, 10}, expected: []int64{5, 5, 5, 15}},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := cumSum(tc.input)
			if tc.expected == nil {
				if result != nil {
					t.Fatalf("expected nil, got %v", result)
				}
				return
			}
			if len(result) != len(tc.expected) {
				t.Fatalf("expected len %d, got %d", len(tc.expected), len(result))
			}
			for i, v := range tc.expected {
				if result[i] != v {
					t.Fatalf("at index %d: expected %d, got %d", i, v, result[i])
				}
			}
		})
	}
}

func TestEmptyInt64Slice(t *testing.T) {
	if emptyInt64Slice(nil) == nil {
		t.Fatal("expected non-nil for nil input")
	}
	if len(emptyInt64Slice(nil)) != 0 {
		t.Fatal("expected empty slice for nil input")
	}

	input := []int64{1, 2, 3}
	result := emptyInt64Slice(input)
	if len(result) != 3 {
		t.Fatalf("expected 3, got %d", len(result))
	}
	if result[0] != 1 {
		t.Fatalf("expected 1, got %d", result[0])
	}
}

func TestEmptyStringSlice(t *testing.T) {
	if emptyStringSlice(nil) == nil {
		t.Fatal("expected non-nil for nil input")
	}
	if len(emptyStringSlice(nil)) != 0 {
		t.Fatal("expected empty slice for nil input")
	}

	input := []string{"a", "b"}
	result := emptyStringSlice(input)
	if len(result) != 2 {
		t.Fatalf("expected 2, got %d", len(result))
	}
	if result[0] != "a" {
		t.Fatalf("expected a, got %q", result[0])
	}
}
