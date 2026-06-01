package linkedin

import (
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/linkedin"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/linkedin"
)

// --- Mock ClickHouse infrastructure ---

type mockRow struct{}

func (m *mockRow) Err() error { return nil }
func (m *mockRow) Scan(dest ...any) error {
	for _, d := range dest {
		switch v := d.(type) {
		case *int32:
			*v = 0
		case *int64:
			*v = 0
		case *float64:
			*v = 0
		case *float32:
			*v = 0
		case *string:
			*v = ""
		case *uint8:
			*v = 0
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

var _ clickhouse.Conn = (*mockConn)(nil)

func newTestService() *LinkedInAnalyticsService {
	client := &ch.Client{
		Conn:   &mockConn{},
		Logger: zerolog.New(io.Discard),
	}
	r := repo.NewRepository(client)
	return NewLinkedInAnalyticsService(r, zerolog.New(io.Discard))
}

func validRequest() *types.LinkedInRequest {
	return &types.LinkedInRequest{
		WorkspaceID: "ws1",
		LinkedinID:  "li_123",
		StartDate:   "2025-01-01",
		EndDate:     "2025-01-31",
		Timezone:    "UTC",
	}
}

// --- Service method tests ---

func TestNewLinkedInAnalyticsService(t *testing.T) {
	svc := newTestService()
	if svc == nil {
		t.Fatal("expected non-nil service")
	}
}

func TestGetSummary(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.LinkedInRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.SummaryResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.LinkedInRequest{},
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

func TestGetAudienceGrowth(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.LinkedInRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.AudienceGrowthResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.LinkedInRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns growth and rollup",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.AudienceGrowthResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.AudienceGrowth == nil {
					t.Fatal("expected non-nil audience_growth")
				}
				if resp.AudienceGrowthRollup["current"] == nil {
					t.Fatal("expected 'current' in rollup")
				}
				if resp.AudienceGrowthRollup["previous"] == nil {
					t.Fatal("expected 'previous' in rollup")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetAudienceGrowth(context.Background(), tc.req)
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

func TestGetPageViews(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.LinkedInRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.PageViewsResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.LinkedInRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns views and rollup",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.PageViewsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.PageViews == nil {
					t.Fatal("expected non-nil page_views")
				}
				if resp.PageViewsRollup["current"] == nil {
					t.Fatal("expected 'current' in rollup")
				}
				if resp.PageViewsRollup["previous"] == nil {
					t.Fatal("expected 'previous' in rollup")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetPageViews(context.Background(), tc.req)
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

func TestGetPublishingBehaviour(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.PublishingBehaviourRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.PublishingBehaviourResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.PublishingBehaviourRequest{},
			expectErr: true,
		},
		{
			name: "valid request with default media types",
			req:  &types.PublishingBehaviourRequest{LinkedInRequest: *validRequest()},
			checkResp: func(t *testing.T, resp *types.PublishingBehaviourResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.PublishingBehaviour == nil {
					t.Fatal("expected non-nil publishing_behaviour")
				}
				if resp.PublishingBehaviourRollup["current"] == nil {
					t.Fatal("expected 'current' in rollup")
				}
				if resp.PublishingBehaviourRollup["previous"] == nil {
					t.Fatal("expected 'previous' in rollup")
				}
			},
		},
		{
			name: "valid request with custom media types",
			req: &types.PublishingBehaviourRequest{
				LinkedInRequest: *validRequest(),
				MediaType:       []string{"images", "videos"},
			},
			checkResp: func(t *testing.T, resp *types.PublishingBehaviourResponse) {
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
			resp, err := svc.GetPublishingBehaviour(context.Background(), tc.req)
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
			req:  &types.TopPostsRequest{LinkedInRequest: *validRequest()},
			checkResp: func(t *testing.T, resp *types.TopPostsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.TopPosts == nil {
					t.Fatal("expected non-nil top_posts")
				}
			},
		},
		{
			name: "valid request with custom limit and order",
			req: &types.TopPostsRequest{
				LinkedInRequest: *validRequest(),
				Limit:           5,
				OrderBy:         "impressions",
			},
			checkResp: func(t *testing.T, resp *types.TopPostsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
			},
		},
		{
			name: "valid request with hashtag filter",
			req: &types.TopPostsRequest{
				LinkedInRequest: *validRequest(),
				Hashtags:        []string{"tech", "marketing"},
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

func TestGetPostsPerDay(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.LinkedInRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.PostsPerDayResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.LinkedInRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns all days",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.PostsPerDayResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.PostsPerDays == nil {
					t.Fatal("expected non-nil posts_per_days")
				}
				days := resp.PostsPerDays.Data.Days
				for _, day := range []string{"Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"} {
					if _, ok := days[day]; !ok {
						t.Fatalf("expected %q in days", day)
					}
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetPostsPerDay(context.Background(), tc.req)
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

func TestGetHashtags(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.LinkedInRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.HashtagsResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.LinkedInRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns hashtags and rollup",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.HashtagsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.TopHashtags == nil {
					t.Fatal("expected non-nil top_hashtags")
				}
				if resp.TopHashtagsRollup["current"] == nil {
					t.Fatal("expected 'current' in rollup")
				}
				if resp.TopHashtagsRollup["previous"] == nil {
					t.Fatal("expected 'previous' in rollup")
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetHashtags(context.Background(), tc.req)
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

func TestGetFollowersDemographics(t *testing.T) {
	tests := []struct {
		name      string
		req       *types.LinkedInRequest
		expectErr bool
		checkResp func(t *testing.T, resp *types.DemographicsResponse)
	}{
		{
			name:      "invalid request",
			req:       &types.LinkedInRequest{},
			expectErr: true,
		},
		{
			name: "valid request returns demographics",
			req:  validRequest(),
			checkResp: func(t *testing.T, resp *types.DemographicsResponse) {
				if !resp.Status {
					t.Fatal("expected status true")
				}
				if resp.FollowerDemographics == nil {
					t.Fatal("expected non-nil follower_demographics")
				}
				for _, key := range []string{"seniority", "industry", "country", "city"} {
					if resp.FollowerDemographics[key] == nil {
						t.Fatalf("expected %q in demographics", key)
					}
				}
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			resp, err := svc.GetFollowersDemographics(context.Background(), tc.req)
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
	tests := []struct {
		name   string
		params *ch.QueryParams
		check  func(t *testing.T, prev *ch.QueryParams)
	}{
		{
			name: "copies account IDs and uses prev dates",
			params: &ch.QueryParams{
				AccountIDs:   []string{"li_123"},
				DateFrom:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				DateTo:       time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
				PrevDateFrom: time.Date(2024, 12, 2, 0, 0, 0, 0, time.UTC),
				PrevDateTo:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
				Timezone:     "UTC",
				DayCount:     31,
			},
			check: func(t *testing.T, prev *ch.QueryParams) {
				if prev.AccountIDs[0] != "li_123" {
					t.Fatalf("expected li_123, got %q", prev.AccountIDs[0])
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
			},
		},
	}

	svc := newTestService()
	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			prev := svc.prevPeriodParams(tc.params)
			tc.check(t, prev)
		})
	}
}

func TestParseDemographicJSON(t *testing.T) {
	tests := []struct {
		name           string
		jsonStr        string
		totalFollowers int64
		checkBuckets   func(t *testing.T, cat *types.DemographicCategory)
	}{
		{
			name:           "empty string",
			jsonStr:        "",
			totalFollowers: 100,
			checkBuckets: func(t *testing.T, cat *types.DemographicCategory) {
				if len(cat.Buckets) != 0 {
					t.Fatalf("expected empty buckets, got %d", len(cat.Buckets))
				}
			},
		},
		{
			name:           "empty object",
			jsonStr:        "{}",
			totalFollowers: 100,
			checkBuckets: func(t *testing.T, cat *types.DemographicCategory) {
				if len(cat.Buckets) != 0 {
					t.Fatalf("expected empty buckets, got %d", len(cat.Buckets))
				}
			},
		},
		{
			name:           "invalid JSON",
			jsonStr:        "not-json",
			totalFollowers: 100,
			checkBuckets: func(t *testing.T, cat *types.DemographicCategory) {
				if len(cat.Buckets) != 0 {
					t.Fatalf("expected empty buckets on invalid JSON, got %d", len(cat.Buckets))
				}
			},
		},
		{
			name:           "valid JSON with exact sum equals total",
			jsonStr:        `{"Engineering": 60, "Marketing": 40}`,
			totalFollowers: 100,
			checkBuckets: func(t *testing.T, cat *types.DemographicCategory) {
				if len(cat.Buckets) != 2 {
					t.Fatalf("expected 2 buckets, got %d", len(cat.Buckets))
				}
				// Sorted desc by value: Engineering(60), Marketing(40)
				if cat.Buckets[0] != "Engineering" {
					t.Fatalf("expected first bucket 'Engineering', got %q", cat.Buckets[0])
				}
				if cat.Values[0] != 60 {
					t.Fatalf("expected first value 60, got %d", cat.Values[0])
				}
			},
		},
		{
			name:           "valid JSON adds Others when sum < total",
			jsonStr:        `{"Engineering": 40, "Marketing": 30}`,
			totalFollowers: 100,
			checkBuckets: func(t *testing.T, cat *types.DemographicCategory) {
				if len(cat.Buckets) != 3 {
					t.Fatalf("expected 3 buckets (2 + Others), got %d", len(cat.Buckets))
				}
				lastIdx := len(cat.Buckets) - 1
				if cat.Buckets[lastIdx] != "Others" {
					t.Fatalf("expected last bucket 'Others', got %q", cat.Buckets[lastIdx])
				}
				if cat.Values[lastIdx] != 30 {
					t.Fatalf("expected Others value 30, got %d", cat.Values[lastIdx])
				}
			},
		},
		{
			name:           "no Others when totalFollowers is 0",
			jsonStr:        `{"Engineering": 40}`,
			totalFollowers: 0,
			checkBuckets: func(t *testing.T, cat *types.DemographicCategory) {
				if len(cat.Buckets) != 1 {
					t.Fatalf("expected 1 bucket, got %d", len(cat.Buckets))
				}
			},
		},
		{
			name:           "no Others when sum >= totalFollowers",
			jsonStr:        `{"Engineering": 60, "Marketing": 50}`,
			totalFollowers: 100,
			checkBuckets: func(t *testing.T, cat *types.DemographicCategory) {
				if len(cat.Buckets) != 2 {
					t.Fatalf("expected 2 buckets, got %d", len(cat.Buckets))
				}
			},
		},
		{
			name:           "sorted descending by value",
			jsonStr:        `{"A": 10, "B": 50, "C": 30}`,
			totalFollowers: 100,
			checkBuckets: func(t *testing.T, cat *types.DemographicCategory) {
				if cat.Values[0] != 50 {
					t.Fatalf("expected first value 50, got %d", cat.Values[0])
				}
				if cat.Values[1] != 30 {
					t.Fatalf("expected second value 30, got %d", cat.Values[1])
				}
				if cat.Values[2] != 10 {
					t.Fatalf("expected third value 10, got %d", cat.Values[2])
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := parseDemographicJSON(tc.jsonStr, tc.totalFollowers)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			tc.checkBuckets(t, result)
		})
	}
}

func TestEmptyInt32Slice(t *testing.T) {
	tests := []struct {
		name     string
		input    []int32
		expected int
	}{
		{name: "nil returns empty", input: nil, expected: 0},
		{name: "empty returns empty", input: []int32{}, expected: 0},
		{name: "non-empty returns same", input: []int32{1, 2, 3}, expected: 3},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := emptyInt32Slice(tc.input)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if len(result) != tc.expected {
				t.Fatalf("expected len %d, got %d", tc.expected, len(result))
			}
		})
	}
}

func TestEmptyFloat32Slice(t *testing.T) {
	tests := []struct {
		name     string
		input    []float32
		expected int
	}{
		{name: "nil returns empty", input: nil, expected: 0},
		{name: "empty returns empty", input: []float32{}, expected: 0},
		{name: "non-empty returns same", input: []float32{1.0, 2.0}, expected: 2},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := emptyFloat32Slice(tc.input)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if len(result) != tc.expected {
				t.Fatalf("expected len %d, got %d", tc.expected, len(result))
			}
		})
	}
}

func TestEmptyStringSlice(t *testing.T) {
	tests := []struct {
		name     string
		input    []string
		expected int
	}{
		{name: "nil returns empty", input: nil, expected: 0},
		{name: "empty returns empty", input: []string{}, expected: 0},
		{name: "non-empty returns same", input: []string{"a", "b"}, expected: 2},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := emptyStringSlice(tc.input)
			if result == nil {
				t.Fatal("expected non-nil result")
			}
			if len(result) != tc.expected {
				t.Fatalf("expected len %d, got %d", tc.expected, len(result))
			}
		})
	}
}

// --- Mapper tests ---

func TestMapSummary(t *testing.T) {
	tests := []struct {
		name     string
		posts    *repo.PostsSummaryResult
		insights *repo.InsightsSummaryResult
		check    func(t *testing.T, m *types.SummaryMetrics)
	}{
		{
			name:     "zero values",
			posts:    &repo.PostsSummaryResult{},
			insights: &repo.InsightsSummaryResult{},
			check: func(t *testing.T, m *types.SummaryMetrics) {
				if m.PostComments != 0 || m.TotalPosts != 0 {
					t.Fatal("expected zero values")
				}
			},
		},
		{
			name: "maps all fields correctly",
			posts: &repo.PostsSummaryResult{
				PostComments:       10,
				PostLikes:          20,
				TotalEngagement:    50,
				TotalPosts:         5,
				PostShares:         15,
				PostClicks:         25,
				PostEngagementRate: 3.2,
			},
			insights: &repo.InsightsSummaryResult{
				Followers:          1000,
				PageViews:          500,
				PageReach:          300,
				PageShares:         40,
				PageComments:       30,
				PageReactions:      60,
				PageImpressions:    800,
				PageUniqueVisitors: 200,
				EngagementRate:     5.5,
			},
			check: func(t *testing.T, m *types.SummaryMetrics) {
				if m.PostComments != 10 {
					t.Fatalf("expected 10, got %d", m.PostComments)
				}
				if m.Followers != 1000 {
					t.Fatalf("expected 1000, got %d", m.Followers)
				}
				if m.EngagementRate != 5.5 {
					t.Fatalf("expected 5.5, got %f", m.EngagementRate)
				}
				if m.PostEngagementRate != 3.2 {
					t.Fatalf("expected 3.2, got %f", m.PostEngagementRate)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapSummary(tc.posts, tc.insights)
			tc.check(t, result)
		})
	}
}

func TestMapAudienceGrowth(t *testing.T) {
	tests := []struct {
		name  string
		input *repo.AudienceResult
		check func(t *testing.T, d *types.AudienceGrowthData)
	}{
		{
			name:  "empty data",
			input: &repo.AudienceResult{},
			check: func(t *testing.T, d *types.AudienceGrowthData) {
				if d.ShowData != 0 {
					t.Fatalf("expected 0, got %d", d.ShowData)
				}
				if len(d.Buckets) != 0 {
					t.Fatalf("expected 0 buckets, got %d", len(d.Buckets))
				}
			},
		},
		{
			name: "formats buckets as YYYY-MM-DD",
			input: &repo.AudienceResult{
				ShowData:             1,
				TotalFollowerCount:   []int32{100, 110},
				OrganicFollowerCount: []int32{80, 90},
				PaidFollowerCount:    []int32{20, 20},
				TotalFollowersDaily:  []int32{0, 10},
				Buckets: []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			check: func(t *testing.T, d *types.AudienceGrowthData) {
				if d.ShowData != 1 {
					t.Fatalf("expected 1, got %d", d.ShowData)
				}
				if len(d.Buckets) != 2 {
					t.Fatalf("expected 2 buckets, got %d", len(d.Buckets))
				}
				if d.Buckets[0] != "2025-01-01" {
					t.Fatalf("expected 2025-01-01, got %q", d.Buckets[0])
				}
				if d.TotalFollowerCount[0] != 100 {
					t.Fatalf("expected 100, got %d", d.TotalFollowerCount[0])
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapAudienceGrowth(tc.input)
			tc.check(t, result)
		})
	}
}

func TestMapPageViews(t *testing.T) {
	tests := []struct {
		name  string
		input *repo.PageViewsResult
		check func(t *testing.T, d *types.PageViewsData)
	}{
		{
			name:  "empty data",
			input: &repo.PageViewsResult{},
			check: func(t *testing.T, d *types.PageViewsData) {
				if d.ShowData != 0 {
					t.Fatalf("expected 0, got %d", d.ShowData)
				}
			},
		},
		{
			name: "maps and formats buckets",
			input: &repo.PageViewsResult{
				DesktopPageViews: []int32{10, 20},
				MobilePageViews:  []int32{5, 15},
				TotalPageViews:   []int32{15, 35},
				ShowData:         35,
				Buckets: []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			check: func(t *testing.T, d *types.PageViewsData) {
				if d.ShowData != 35 {
					t.Fatalf("expected 35, got %d", d.ShowData)
				}
				if d.DesktopPageViews[0] != 10 {
					t.Fatalf("expected 10, got %d", d.DesktopPageViews[0])
				}
				if d.Buckets[1] != "2025-01-02" {
					t.Fatalf("expected 2025-01-02, got %q", d.Buckets[1])
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapPageViews(tc.input)
			tc.check(t, result)
		})
	}
}

func TestMapPublishingRollup(t *testing.T) {
	tests := []struct {
		name     string
		input    []repo.PublishingRollupRow
		expected int
	}{
		{name: "nil returns empty", input: nil, expected: 0},
		{name: "empty returns empty", input: []repo.PublishingRollupRow{}, expected: 0},
		{
			name: "maps rows correctly",
			input: []repo.PublishingRollupRow{
				{MediaType: "images", TotalPosts: 5, Likes: 10},
				{MediaType: "total", TotalPosts: 5, Likes: 10},
			},
			expected: 2,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapPublishingRollup(tc.input)
			if len(result) != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, len(result))
			}
			if tc.expected > 0 {
				if result[0].MediaType != tc.input[0].MediaType {
					t.Fatalf("expected %q, got %q", tc.input[0].MediaType, result[0].MediaType)
				}
			}
		})
	}
}

func TestMapTopPosts(t *testing.T) {
	tests := []struct {
		name     string
		input    []repo.TopPostResult
		expected int
	}{
		{name: "nil returns empty", input: nil, expected: 0},
		{name: "empty returns empty", input: []repo.TopPostResult{}, expected: 0},
		{
			name: "maps fields and formats times",
			input: []repo.TopPostResult{
				{
					LinkedinID:      "li_123",
					PostID:          "post_1",
					MediaType:       "images",
					TotalEngagement: 50.0,
					CreatedAt:       time.Date(2025, 1, 15, 10, 30, 0, 0, time.UTC),
					SavingTime:      time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC),
					PublishedAt:     time.Date(2025, 1, 15, 8, 0, 0, 0, time.UTC),
				},
			},
			expected: 1,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapTopPosts(tc.input, "America/New_York")
			if len(result) != tc.expected {
				t.Fatalf("expected %d, got %d", tc.expected, len(result))
			}
			if tc.expected > 0 {
				if result[0].PostID != "post_1" {
					t.Fatalf("expected post_1, got %q", result[0].PostID)
				}
				if result[0].CreatedAt != "2025-01-15T05:30:00-05:00" {
					t.Fatalf("expected timezone-adjusted RFC3339 format, got %q", result[0].CreatedAt)
				}
				if result[0].TotalEngagement != 50.0 {
					t.Fatalf("expected 50.0, got %f", result[0].TotalEngagement)
				}
			}
		})
	}
}

func TestMapHashtags(t *testing.T) {
	tests := []struct {
		name  string
		input *repo.HashtagsResult
		check func(t *testing.T, d *types.HashtagsData)
	}{
		{
			name:  "nil slices become empty",
			input: &repo.HashtagsResult{},
			check: func(t *testing.T, d *types.HashtagsData) {
				if d.Name == nil {
					t.Fatal("expected non-nil Name")
				}
				if len(d.Name) != 0 {
					t.Fatalf("expected 0, got %d", len(d.Name))
				}
			},
		},
		{
			name: "maps values correctly",
			input: &repo.HashtagsResult{
				Name:        []string{"tech", "ai"},
				Engagements: []int32{100, 80},
				Likes:       []int32{50, 40},
				Comments:    []int32{30, 20},
				Shares:      []int32{20, 20},
				Posts:       []int32{5, 3},
			},
			check: func(t *testing.T, d *types.HashtagsData) {
				if len(d.Name) != 2 {
					t.Fatalf("expected 2, got %d", len(d.Name))
				}
				if d.Name[0] != "tech" {
					t.Fatalf("expected tech, got %q", d.Name[0])
				}
				if d.Engagements[0] != 100 {
					t.Fatalf("expected 100, got %d", d.Engagements[0])
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapHashtags(tc.input)
			tc.check(t, result)
		})
	}
}

func TestMapHashtagsRollup(t *testing.T) {
	tests := []struct {
		name  string
		input *repo.HashtagsRollupResult
		check func(t *testing.T, r *types.HashtagsRollup)
	}{
		{
			name:  "zero values",
			input: &repo.HashtagsRollupResult{},
			check: func(t *testing.T, r *types.HashtagsRollup) {
				if r.TotalHashtags != 0 {
					t.Fatalf("expected 0, got %d", r.TotalHashtags)
				}
			},
		},
		{
			name: "maps all fields",
			input: &repo.HashtagsRollupResult{
				TotalHashtags:    15,
				TotalTimesUsed:   50,
				TotalLikes:       200,
				TotalComments:    80,
				TotalShares:      60,
				TotalEngagement:  340,
				TotalImpressions: 5000,
				TotalReach:       3000,
			},
			check: func(t *testing.T, r *types.HashtagsRollup) {
				if r.TotalHashtags != 15 {
					t.Fatalf("expected 15, got %d", r.TotalHashtags)
				}
				if r.TotalEngagement != 340 {
					t.Fatalf("expected 340, got %d", r.TotalEngagement)
				}
				if r.TotalReach != 3000 {
					t.Fatalf("expected 3000, got %d", r.TotalReach)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapHashtagsRollup(tc.input)
			tc.check(t, result)
		})
	}
}

func TestMapAudienceRollup(t *testing.T) {
	tests := []struct {
		name  string
		input *repo.AudienceRollupResult
		check func(t *testing.T, r *types.AudienceGrowthRollup)
	}{
		{
			name:  "zero values",
			input: &repo.AudienceRollupResult{},
			check: func(t *testing.T, r *types.AudienceGrowthRollup) {
				if r.TotalFollowerCount != 0 {
					t.Fatalf("expected 0, got %d", r.TotalFollowerCount)
				}
			},
		},
		{
			name: "maps all fields",
			input: &repo.AudienceRollupResult{
				OrganicFollowerCount: 800,
				PaidFollowerCount:    200,
				TotalFollowerCount:   1000,
				AvgFollowerCount:     950.5,
			},
			check: func(t *testing.T, r *types.AudienceGrowthRollup) {
				if r.OrganicFollowerCount != 800 {
					t.Fatalf("expected 800, got %d", r.OrganicFollowerCount)
				}
				if r.TotalFollowerCount != 1000 {
					t.Fatalf("expected 1000, got %d", r.TotalFollowerCount)
				}
				if r.AvgFollowerCount != 950.5 {
					t.Fatalf("expected 950.5, got %f", r.AvgFollowerCount)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapAudienceRollup(tc.input)
			tc.check(t, result)
		})
	}
}

func TestMapPageViewsRollup(t *testing.T) {
	tests := []struct {
		name  string
		input *repo.PageViewsRollupResult
		check func(t *testing.T, r *types.PageViewsRollup)
	}{
		{
			name:  "zero values",
			input: &repo.PageViewsRollupResult{},
			check: func(t *testing.T, r *types.PageViewsRollup) {
				if r.TotalPageViews != 0 {
					t.Fatalf("expected 0, got %d", r.TotalPageViews)
				}
			},
		},
		{
			name: "maps all fields",
			input: &repo.PageViewsRollupResult{
				TotalPageViews:   500,
				DesktopPageViews: 300,
				MobilePageViews:  200,
				AvgPageViews:     16.67,
			},
			check: func(t *testing.T, r *types.PageViewsRollup) {
				if r.TotalPageViews != 500 {
					t.Fatalf("expected 500, got %d", r.TotalPageViews)
				}
				if r.DesktopPageViews != 300 {
					t.Fatalf("expected 300, got %d", r.DesktopPageViews)
				}
				if r.AvgPageViews != 16.67 {
					t.Fatalf("expected 16.67, got %f", r.AvgPageViews)
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapPageViewsRollup(tc.input)
			tc.check(t, result)
		})
	}
}

func TestMapPublishingBehaviour(t *testing.T) {
	tests := []struct {
		name  string
		input *repo.PublishingResult
		check func(t *testing.T, d *types.PublishingBehaviourData)
	}{
		{
			name:  "nil slices become empty",
			input: &repo.PublishingResult{},
			check: func(t *testing.T, d *types.PublishingBehaviourData) {
				if d.Likes == nil {
					t.Fatal("expected non-nil Likes")
				}
				if len(d.Buckets) != 0 {
					t.Fatalf("expected 0 buckets, got %d", len(d.Buckets))
				}
			},
		},
		{
			name: "maps and formats correctly",
			input: &repo.PublishingResult{
				Likes:          []int32{10, 20},
				Comments:       []int32{5, 8},
				Shares:         []int32{2, 3},
				Clicks:         []int32{15, 25},
				EngagementRate: []float32{2.5, 3.1},
				Impressions:    []int32{100, 200},
				TotalPosts:     []int32{3, 4},
				Engagement:     []int32{17, 31},
				Reach:          []int32{80, 150},
				Buckets: []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				},
			},
			check: func(t *testing.T, d *types.PublishingBehaviourData) {
				if d.Likes[0] != 10 {
					t.Fatalf("expected 10, got %d", d.Likes[0])
				}
				if d.EngagementRate[0] != 2.5 {
					t.Fatalf("expected 2.5, got %f", d.EngagementRate[0])
				}
				if d.Buckets[0] != "2025-01-01" {
					t.Fatalf("expected 2025-01-01, got %q", d.Buckets[0])
				}
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := mapPublishingBehaviour(tc.input)
			tc.check(t, result)
		})
	}
}

// --- Error-returning mock infrastructure ---

type mockErrRow struct{}

func (m *mockErrRow) Err() error                { return fmt.Errorf("mock error") }
func (m *mockErrRow) Scan(dest ...any) error    { return fmt.Errorf("mock scan error") }
func (m *mockErrRow) ScanStruct(dest any) error { return fmt.Errorf("mock scan error") }

type mockErrConn struct{ mockConn }

func (m *mockErrConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	return &mockErrRow{}
}
func (m *mockErrConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	return nil, fmt.Errorf("mock query error")
}

var _ clickhouse.Conn = (*mockErrConn)(nil)

func newErrTestService() *LinkedInAnalyticsService {
	client := &ch.Client{
		Conn:   &mockErrConn{},
		Logger: zerolog.New(io.Discard),
	}
	r := repo.NewRepository(client)
	return NewLinkedInAnalyticsService(r, zerolog.New(io.Discard))
}

// --- Error-path tests ---

func TestServiceMethods_RepoErrors(t *testing.T) {
	svc := newErrTestService()

	t.Run("GetSummary", func(t *testing.T) {
		resp, err := svc.GetSummary(context.Background(), validRequest())
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !resp.Status {
			t.Fatal("expected status true")
		}
		if resp.Overview["current"].PostComments != 0 {
			t.Fatal("expected zero defaults on error")
		}
		if resp.Overview["previous"].PostComments != 0 {
			t.Fatal("expected zero defaults on error")
		}
	})

	t.Run("GetAudienceGrowth", func(t *testing.T) {
		resp, err := svc.GetAudienceGrowth(context.Background(), validRequest())
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !resp.Status {
			t.Fatal("expected status true")
		}
		if resp.AudienceGrowth == nil {
			t.Fatal("expected non-nil audience_growth")
		}
	})

	t.Run("GetPageViews", func(t *testing.T) {
		resp, err := svc.GetPageViews(context.Background(), validRequest())
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !resp.Status {
			t.Fatal("expected status true")
		}
		if resp.PageViews == nil {
			t.Fatal("expected non-nil page_views")
		}
	})

	t.Run("GetPublishingBehaviour", func(t *testing.T) {
		req := &types.PublishingBehaviourRequest{LinkedInRequest: *validRequest()}
		resp, err := svc.GetPublishingBehaviour(context.Background(), req)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !resp.Status {
			t.Fatal("expected status true")
		}
		if resp.PublishingBehaviour == nil {
			t.Fatal("expected non-nil publishing_behaviour")
		}
	})

	t.Run("GetTopPosts", func(t *testing.T) {
		req := &types.TopPostsRequest{LinkedInRequest: *validRequest()}
		resp, err := svc.GetTopPosts(context.Background(), req)
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !resp.Status {
			t.Fatal("expected status true")
		}
		if len(resp.TopPosts) != 0 {
			t.Fatalf("expected empty top posts, got %d", len(resp.TopPosts))
		}
	})

	t.Run("GetPostsPerDay", func(t *testing.T) {
		resp, err := svc.GetPostsPerDay(context.Background(), validRequest())
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !resp.Status {
			t.Fatal("expected status true")
		}
		if resp.PostsPerDays.Data.ShowData != 0 {
			t.Fatal("expected zero show_data on error")
		}
	})

	t.Run("GetHashtags", func(t *testing.T) {
		resp, err := svc.GetHashtags(context.Background(), validRequest())
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !resp.Status {
			t.Fatal("expected status true")
		}
		if resp.TopHashtags == nil {
			t.Fatal("expected non-nil top_hashtags")
		}
	})

	t.Run("GetFollowersDemographics", func(t *testing.T) {
		resp, err := svc.GetFollowersDemographics(context.Background(), validRequest())
		if err != nil {
			t.Fatalf("expected nil error, got: %v", err)
		}
		if !resp.Status {
			t.Fatal("expected status true")
		}
		if len(resp.FollowerDemographics) != 0 {
			t.Fatalf("expected empty demographics map on error, got %d entries", len(resp.FollowerDemographics))
		}
	})
}

// --- Audience growth fallback mock infrastructure ---

type mockFallbackRow struct {
	query string
}

func (m *mockFallbackRow) Err() error { return nil }
func (m *mockFallbackRow) Scan(dest ...any) error {
	if strings.Contains(m.query, "arrayFill") {
		for _, d := range dest {
			switch v := d.(type) {
			case *uint8:
				*v = 1
			case *[]int32:
				*v = []int32{0, 0}
			case *[]time.Time:
				*v = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
			}
		}
		return nil
	}
	if strings.Contains(m.query, "arrayFirst") {
		idx := 0
		values := []int32{500, 400, 100}
		for _, d := range dest {
			if v, ok := d.(*int32); ok && idx < len(values) {
				*v = values[idx]
				idx++
			}
		}
		return nil
	}
	for _, d := range dest {
		switch v := d.(type) {
		case *int32:
			*v = 0
		case *int64:
			*v = 0
		case *float64:
			*v = 0
		case *float32:
			*v = 0
		case *string:
			*v = ""
		case *uint8:
			*v = 0
		}
	}
	return nil
}
func (m *mockFallbackRow) ScanStruct(dest any) error { return nil }

type mockFallbackConn struct{ mockConn }

func (m *mockFallbackConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	return &mockFallbackRow{query: query}
}

var _ clickhouse.Conn = (*mockFallbackConn)(nil)

func newFallbackTestService() *LinkedInAnalyticsService {
	client := &ch.Client{
		Conn:   &mockFallbackConn{},
		Logger: zerolog.New(io.Discard),
	}
	r := repo.NewRepository(client)
	return NewLinkedInAnalyticsService(r, zerolog.New(io.Discard))
}

func TestGetAudienceGrowth_FallbackFillsZeros(t *testing.T) {
	svc := newFallbackTestService()
	resp, err := svc.GetAudienceGrowth(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Status {
		t.Fatal("expected status true")
	}

	growth := resp.AudienceGrowth

	if growth.TotalFollowerCount[0] != 500 {
		t.Fatalf("expected TotalFollowerCount[0]=500, got %d", growth.TotalFollowerCount[0])
	}
	if growth.TotalFollowerCount[1] != 500 {
		t.Fatalf("expected TotalFollowerCount[1]=500, got %d", growth.TotalFollowerCount[1])
	}
	if growth.OrganicFollowerCount[0] != 400 {
		t.Fatalf("expected OrganicFollowerCount[0]=400, got %d", growth.OrganicFollowerCount[0])
	}
	if growth.OrganicFollowerCount[1] != 400 {
		t.Fatalf("expected OrganicFollowerCount[1]=400, got %d", growth.OrganicFollowerCount[1])
	}
	if growth.PaidFollowerCount[0] != 100 {
		t.Fatalf("expected PaidFollowerCount[0]=100, got %d", growth.PaidFollowerCount[0])
	}
	if growth.PaidFollowerCount[1] != 100 {
		t.Fatalf("expected PaidFollowerCount[1]=100, got %d", growth.PaidFollowerCount[1])
	}
	if len(growth.Buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(growth.Buckets))
	}
	if growth.Buckets[0] != "2025-01-01" {
		t.Fatalf("expected 2025-01-01, got %q", growth.Buckets[0])
	}
}
