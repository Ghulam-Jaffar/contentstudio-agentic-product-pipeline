package twitter

import (
	"context"
	"io"
	"testing"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	repo "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse/analytics-get-queries/twitter"
	types "github.com/d4interactive/contentstudio-social-analytics-go/src/models/analytics/twitter"
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
		case *[]int64:
			if x, ok := v.([]int64); ok {
				*d = x
			}
		case *[]string:
			if x, ok := v.([]string); ok {
				*d = x
			}
		case *int32:
			if x, ok := v.(int32); ok {
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
		case *[]string:
			if x, ok := v.([]string); ok {
				*d = x
			}
		case *int32:
			if x, ok := v.(int32); ok {
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

func validRequest() *types.TwitterRequest {
	return &types.TwitterRequest{
		WorkspaceID: "ws1",
		TwitterID:   "tw_123",
		StartDate:   "2025-01-01",
		EndDate:     "2025-01-31",
		Timezone:    "UTC",
	}
}

func newTestService() *TwitterAnalyticsService {
	conn := &mockConn{
		row: &mockRow{
			values: []any{
				"tw_123", "name", "img", int64(100), int64(80), int64(30), int64(2),
				int64(400), int64(120), int64(20), int64(10), int64(5), int64(70), int64(3), int64(30),
			},
		},
		rows: &mockRows{
			rows: [][]any{
				{
					"tweet_1", "2025-01-01 10:00:00", "hello", "text", "https://x.com/1",
					[]string{"https://img"}, int32(1), int32(2), int32(3), int32(4), int32(5), int32(6), int32(7), int32(8),
				},
			},
		},
	}
	client := &ch.Client{Conn: conn, Logger: zerolog.New(io.Discard)}
	r := repo.NewRepository(client)
	return &TwitterAnalyticsService{repo: r, logger: zerolog.New(io.Discard)}
}

func TestGetPageAndPostsInsights(t *testing.T) {
	svc := newTestService()

	_, err := svc.GetPageAndPostsInsights(context.Background(), &types.TwitterRequest{})
	if err == nil {
		t.Fatal("expected validation error for invalid request")
	}

	resp, err := svc.GetPageAndPostsInsights(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Data["twitter_id"] != "tw_123" {
		t.Fatalf("expected twitter_id tw_123, got %v", resp.Data["twitter_id"])
	}
}

func TestGetEngagementImpressionData(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetEngagementImpressionData(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestGetFollowersTrendData(t *testing.T) {
	svc := newTestService()
	resp, err := svc.GetFollowersTrendData(context.Background(), validRequest())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
}

func TestGetTopTweets(t *testing.T) {
	svc := newTestService()
	req := &types.TweetsRequest{TwitterRequest: *validRequest(), Limit: 5, OrderBy: "total_engagement"}
	resp, err := svc.GetTopTweets(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.TopTweets) != 1 {
		t.Fatalf("expected 1 row, got %d", len(resp.TopTweets))
	}
}

func TestGetLeastTweets(t *testing.T) {
	svc := newTestService()
	req := &types.TweetsRequest{TwitterRequest: *validRequest(), Limit: 5, OrderBy: "total_engagement"}
	resp, err := svc.GetLeastTweets(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resp.LeastTweets) != 1 {
		t.Fatalf("expected 1 row, got %d", len(resp.LeastTweets))
	}
}

func TestGetCreditsUsedCount_NoCollection(t *testing.T) {
	svc := newTestService()
	_, err := svc.GetCreditsUsedCount(context.Background(), validRequest())
	if err == nil {
		t.Fatal("expected error when jobs metadata collection is not configured")
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
