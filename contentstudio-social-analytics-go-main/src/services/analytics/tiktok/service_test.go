package tiktok

import (
	"context"
	"io"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/tiktok"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/tiktok"
)

type mockRow struct {
	values []any
	err    error
}

func (m *mockRow) Err() error { return m.err }
func (m *mockRow) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	for i, v := range m.values {
		if i >= len(dest) {
			break
		}
		switch d := dest[i].(type) {
		case *string:
			if x, ok := v.(string); ok {
				*d = x
			}
		case *int64:
			if x, ok := v.(int64); ok {
				*d = x
			}
		case *float64:
			if x, ok := v.(float64); ok {
				*d = x
			}
		case *[]int64:
			if x, ok := v.([]int64); ok {
				*d = x
			}
		case *[]string:
			if x, ok := v.([]string); ok {
				*d = x
			}
		}
	}
	return nil
}
func (m *mockRow) ScanStruct(dest any) error { return m.err }

type mockRows struct {
	rows      [][]any
	idx       int
	scanErr   error
	returnErr error
}

func (m *mockRows) Columns() []string                { return nil }
func (m *mockRows) ColumnTypes() []driver.ColumnType { return nil }
func (m *mockRows) Next() bool {
	if m.idx >= len(m.rows) {
		return false
	}
	m.idx++
	return true
}
func (m *mockRows) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	vals := m.rows[m.idx-1]
	for i, v := range vals {
		if i >= len(dest) {
			break
		}
		switch d := dest[i].(type) {
		case *string:
			if x, ok := v.(string); ok {
				*d = x
			}
		case *int64:
			if x, ok := v.(int64); ok {
				*d = x
			}
		case *float64:
			if x, ok := v.(float64); ok {
				*d = x
			}
		case *[]string:
			if x, ok := v.([]string); ok {
				*d = x
			}
		}
	}
	return nil
}
func (m *mockRows) ScanStruct(dest any) error { return m.scanErr }
func (m *mockRows) Totals(dest ...any) error  { return nil }
func (m *mockRows) Close() error              { return nil }
func (m *mockRows) Err() error                { return m.returnErr }

type mockConn struct {
	row      driver.Row
	rows     driver.Rows
	queryErr error
}

func (m *mockConn) Contributors() []string                        { return nil }
func (m *mockConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (m *mockConn) Select(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *mockConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return m.rows, nil
}
func (m *mockConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	return m.row
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

func validRequest() *types.TiktokRequest {
	return &types.TiktokRequest{
		WorkspaceID: "ws1",
		TiktokID:    "tt_123",
		StartDate:   "2025-01-01",
		EndDate:     "2025-01-31",
		Timezone:    "UTC",
	}
}

func newTestService() *TiktokAnalyticsService {
	conn := &mockConn{
		row: &mockRow{
			values: []any{
				"tt_123", "name", "",
				int64(100), int64(20), int64(10), int64(130), int64(5),
				int64(1000), int64(100), int64(5000),
			},
		},
		rows: &mockRows{
			rows: [][]any{
				{
					"top_posts", "tt_123", "name", "", "https://profile", "post_1", "https://img", "https://share", "desc",
					[]string{"#a"}, int64(10), int64(720), int64(1280), "title", "<iframe/>", "https://embed",
					int64(100), int64(20), int64(10), int64(500), int64(130), int64(130), float64(13.5),
					"2025-01-01 00:00:00", "2025-01-01 00:00:00", int64(1000), int64(1),
				},
			},
		},
	}
	client := &ch.Client{Conn: conn, Logger: zerolog.New(io.Discard)}
	r := repo.NewRepository(client)
	return &TiktokAnalyticsService{repo: r, logger: zerolog.New(io.Discard)}
}

func TestGetPageAndPostsInsights(t *testing.T) {
	svc := newTestService()

	if _, err := svc.GetPageAndPostsInsights(context.Background(), &types.TiktokRequest{}); err == nil {
		t.Fatal("expected validation error for invalid request")
	}

	resp, err := svc.GetPageAndPostsInsights(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data := resp["data"].(map[string]interface{})
	if data["tiktok_id"] != "tt_123" {
		t.Fatalf("expected tiktok_id tt_123, got %v", data["tiktok_id"])
	}
}

func TestGetPageFollowersAndViews(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetPageFollowersAndViews(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestGetPostsAndEngagements(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetPostsAndEngagements(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestGetDailyEngagementsData(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetDailyEngagementsData(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestGetTopAndLeastPerformingPosts(t *testing.T) {
	svc := newTestService()
	req := validRequest()
	req.Timezone = "America/New_York"
	resp, err := svc.GetTopAndLeastPerformingPosts(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	topPosts := resp["data"].(map[string]interface{})["top_posts"].([]map[string]interface{})
	if topPosts[0]["created_time"] != "2024-12-31T19:00:00-05:00" {
		t.Fatalf("unexpected timezone-adjusted created_time: %v", topPosts[0]["created_time"])
	}
}

func TestGetPostsData(t *testing.T) {
	svc := newTestService()
	req := &types.PostsRequest{
		TiktokRequest: *validRequest(),
		Limit:         5,
		Offset:        0,
		SortOrder:     "total_engagement",
	}
	resp, err := svc.GetPostsData(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	posts := resp["data"].([]map[string]interface{})
	if posts[0]["created_time"] != "2025-01-01T00:00:00Z" {
		t.Fatalf("unexpected created_time: %v", posts[0]["created_time"])
	}
}

func TestGrowthAndDiffHelpers(t *testing.T) {
	if got := calculateDiff(10, 0); got != "N/A" {
		t.Fatalf("expected N/A diff for previous=0, got %v", got)
	}
	if got := calculateGrowth(10, 0); got != "N/A" {
		t.Fatalf("expected N/A growth for previous=0, got %v", got)
	}
	if got := calculateDiff(15, 10); got != int64(5) {
		t.Fatalf("expected diff=5, got %v", got)
	}
}
