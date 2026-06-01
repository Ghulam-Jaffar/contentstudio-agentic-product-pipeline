package twitter

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

type mockConn struct {
	queryErr  error
	row       driver.Row
	rows      driver.Rows
	lastQuery string
}

func (m *mockConn) Contributors() []string                        { return nil }
func (m *mockConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (m *mockConn) Select(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *mockConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	m.lastQuery = query
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return m.rows, nil
}
func (m *mockConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	m.lastQuery = query
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

var _ clickhouse.Conn = (*mockConn)(nil)

func newTestClient(conn *mockConn) *ch.Client {
	return &ch.Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test_db"},
		Logger: zerolog.New(io.Discard),
	}
}

func newParams() *ch.QueryParams {
	return &ch.QueryParams{
		AccountIDs: []string{"tw_123"},
		DateFrom:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Timezone:   "UTC",
		DayCount:   31,
	}
}

func TestGetSummary(t *testing.T) {
	conn := &mockConn{
		row: &mockRow{
			values: []any{
				"tw_123", "name", "img", int64(100), int64(80), int64(30), int64(2),
				int64(400), int64(120), int64(20), int64(10), int64(5), int64(70), int64(3), int64(30),
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	resp, err := r.GetSummary(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TwitterID != "tw_123" {
		t.Fatalf("expected twitter_id tw_123, got %q", resp.TwitterID)
	}
}

func TestGetSummary_Error(t *testing.T) {
	conn := &mockConn{row: &mockRow{err: errors.New("scan failed")}}
	r := NewRepository(newTestClient(conn))
	_, err := r.GetSummary(context.Background(), newParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetEngagementImpressionData(t *testing.T) {
	conn := &mockConn{
		row: &mockRow{
			values: []any{
				"tw_123",
				[]int64{1}, []int64{2}, []int64{3}, []string{"2025-01-01"},
				[]int64{4}, []int64{5}, []int64{6}, []int64{7}, []int64{8},
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	resp, err := r.GetEngagementImpressionData(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TwitterID != "tw_123" {
		t.Fatalf("expected twitter_id tw_123, got %q", resp.TwitterID)
	}
}

func TestGetFollowersTrend(t *testing.T) {
	conn := &mockConn{
		row: &mockRow{
			values: []any{
				"tw_123", "name", "username",
				[]int64{100}, []int64{5}, []int64{80}, []int64{3},
				[]string{"2025-01-01"},
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	resp, err := r.GetFollowersTrend(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PlatformID != "tw_123" {
		t.Fatalf("expected platform_id tw_123, got %q", resp.PlatformID)
	}
	if !strings.Contains(conn.lastQuery, "toDate(max(saving_time)) AS bucket_date") {
		t.Fatalf("expected bucket_date projection in query, got:\n%s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "ORDER BY bucket_date ASC") {
		t.Fatalf("expected bucket_date ordering in query, got:\n%s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "WITH FILL FROM toDate('2025-01-01') TO toDate('2025-01-31') + 1 STEP 1") {
		t.Fatalf("expected daily fill clause in query, got:\n%s", conn.lastQuery)
	}
}

func TestGetTweetsData(t *testing.T) {
	conn := &mockConn{
		rows: &mockRows{
			rows: [][]any{
				{
					"tweet_1", "2025-01-01 10:00:00", "hello", "text", "https://x.com/1",
					[]string{"https://img"}, int32(1), int32(2), int32(3), int32(4), int32(5), int32(6), int32(7), int32(8),
				},
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	rows, err := r.GetTweetsData(context.Background(), newParams(), "total_engagement", 5, "DESC")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(conn.lastQuery, "WHERE twitter_id IN ('tw_123') AND toDate(twitter_posts.tweeted_at, 'UTC')") {
		t.Fatalf("expected qualified tweeted_at filter in query, got:\n%s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "AS tweeted_at_value") {
		t.Fatalf("expected non-conflicting tweeted_at alias in query, got:\n%s", conn.lastQuery)
	}
}

func TestGetTweetsData_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	r := NewRepository(newTestClient(conn))
	_, err := r.GetTweetsData(context.Background(), newParams(), "total_engagement", 5, "DESC")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
