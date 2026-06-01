package tiktok

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
	queryErr error
	row      driver.Row
	rows     driver.Rows
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
		AccountIDs: []string{"tt_123"},
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
				"tt_123", "name", "",
				int64(100), int64(20), int64(10), int64(130), int64(5),
				int64(1000), int64(100), int64(5000),
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	resp, err := r.GetSummary(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TiktokID != "tt_123" {
		t.Fatalf("expected tiktok_id tt_123, got %q", resp.TiktokID)
	}
}

func TestGetSummaryError(t *testing.T) {
	conn := &mockConn{row: &mockRow{err: errors.New("scan failed")}}
	r := NewRepository(newTestClient(conn))
	if _, err := r.GetSummary(context.Background(), newParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetFollowersAndViews(t *testing.T) {
	conn := &mockConn{
		row: &mockRow{
			values: []any{
				"tt_123", "name", "",
				[]int64{100, 105}, []int64{1000, 1010}, []int64{0, 5}, []int64{0, 10},
				[]string{"2025-01-01", "2025-01-02"},
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	resp, err := r.GetFollowersAndViews(context.Background(), newParams(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.PlatformID != "tt_123" {
		t.Fatalf("expected platform_id tt_123, got %q", resp.PlatformID)
	}
}

func TestGetPostsAndEngagements(t *testing.T) {
	conn := &mockConn{
		row: &mockRow{
			values: []any{
				"tt_123", "name", "",
				[]string{"2025-01-01"}, []int64{100}, []int64{20}, []int64{10}, []int64{5},
				[]int64{35}, []int64{7}, []int64{2},
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	resp, err := r.GetPostsAndEngagements(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TiktokID != "tt_123" {
		t.Fatalf("expected tiktok_id tt_123, got %q", resp.TiktokID)
	}
	if !strings.Contains(conn.lastQuery, "groupArray(formatDateTime(posting_day, '%Y-%m-%d')) AS days_bucket") {
		t.Fatalf("expected formatted days_bucket in outer query, got:\n%s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "WITH FILL FROM toDate('2025-01-01') TO toDate('2025-02-01') STEP 1") {
		t.Fatalf("expected date-based fill clause in query, got:\n%s", conn.lastQuery)
	}
}

func TestGetDailyEngagementsData(t *testing.T) {
	conn := &mockConn{
		row: &mockRow{
			values: []any{
				"tt_123", "name", "",
				[]int64{10}, []int64{5}, []int64{2},
				[]int64{10}, []int64{5}, []int64{2},
				[]int64{17}, []int64{17}, []string{"2025-01-01"},
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	resp, err := r.GetDailyEngagementsData(context.Background(), newParams(), false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TiktokID != "tt_123" {
		t.Fatalf("expected tiktok_id tt_123, got %q", resp.TiktokID)
	}
	if !strings.Contains(conn.lastQuery, "platform_id AS tiktok_id") {
		t.Fatalf("expected direct platform_id projection in query, got:\n%s", conn.lastQuery)
	}
}

func TestGetTopAndLeastPerformingPosts(t *testing.T) {
	conn := &mockConn{
		rows: &mockRows{
			rows: [][]any{
				{
					"top_posts", "tt_123", "name", "", "https://profile", "post_1", "https://img", "https://share", "desc",
					[]string{"#a"}, int64(10), int64(720), int64(1280), "title", "<iframe/>", "https://embed",
					int64(100), int64(20), int64(10), int64(500), int64(130), int64(130), float64(13.5),
					"2025-01-01 00:00:00", "2025-01-01 00:00:00", int64(1000),
				},
				{
					"least_posts", "tt_123", "name", "", "https://profile", "post_2", "https://img", "https://share", "desc",
					[]string{"#b"}, int64(8), int64(720), int64(1280), "title2", "<iframe/>", "https://embed",
					int64(1), int64(1), int64(0), int64(20), int64(2), int64(2), float64(0.2),
					"2025-01-02 00:00:00", "2025-01-02 00:00:00", int64(1000),
				},
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	top, least, err := r.GetTopAndLeastPerformingPosts(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(top) != 1 || len(least) != 1 {
		t.Fatalf("expected 1 top and 1 least row, got top=%d least=%d", len(top), len(least))
	}
	if !strings.Contains(conn.lastQuery, "AS inserted_at_value") {
		t.Fatalf("expected non-conflicting inserted_at alias in query, got:\n%s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "AS created_time_value") {
		t.Fatalf("expected non-conflicting created_time alias in query, got:\n%s", conn.lastQuery)
	}
}

func TestGetPostsData(t *testing.T) {
	conn := &mockConn{
		rows: &mockRows{
			rows: [][]any{
				{
					"tt_123", "name", "", "https://profile", "post_1", "https://img", "https://share", "desc",
					[]string{"#a"}, int64(10), int64(720), int64(1280), "title", "<iframe/>", "https://embed",
					int64(100), int64(20), int64(10), int64(500), int64(130), int64(130), float64(13.5),
					"2025-01-01 00:00:00", "2025-01-01 00:00:00", int64(1000), int64(1),
				},
			},
		},
	}
	r := NewRepository(newTestClient(conn))
	rows, err := r.GetPostsData(context.Background(), newParams(), "total_engagement", 5, 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}
	if !strings.Contains(conn.lastQuery, "AS inserted_at_value") {
		t.Fatalf("expected non-conflicting inserted_at alias in query, got:\n%s", conn.lastQuery)
	}
}

func TestGetPostsDataQueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	r := NewRepository(newTestClient(conn))
	if _, err := r.GetPostsData(context.Background(), newParams(), "total_engagement", 5, 0); err == nil {
		t.Fatal("expected error, got nil")
	}
}
