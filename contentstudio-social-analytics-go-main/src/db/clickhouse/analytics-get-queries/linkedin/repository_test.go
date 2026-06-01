package linkedin

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

type mockConn struct {
	pingErr         error
	closeErr        error
	queryErr        error
	execErr         error
	prepareBatchErr error
	batchAppendErr  error
	batchSendErr    error
	queryRowResult  driver.Row
	queryRows       driver.Rows
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
	return m.queryRows, nil
}
func (m *mockConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	if m.queryRowResult != nil {
		return m.queryRowResult
	}
	return nil
}
func (m *mockConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	if m.prepareBatchErr != nil {
		return nil, m.prepareBatchErr
	}
	return &mockBatch{appendErr: m.batchAppendErr, sendErr: m.batchSendErr}, nil
}
func (m *mockConn) Exec(ctx context.Context, query string, args ...any) error {
	return m.execErr
}
func (m *mockConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return nil
}
func (m *mockConn) Ping(ctx context.Context) error { return m.pingErr }
func (m *mockConn) Stats() driver.Stats             { return driver.Stats{} }
func (m *mockConn) Close() error                    { return m.closeErr }

type mockBatch struct {
	appendErr   error
	sendErr     error
	appendCount int
}

func (m *mockBatch) Abort() error                  { return nil }
func (m *mockBatch) Append(v ...any) error         { m.appendCount++; return m.appendErr }
func (m *mockBatch) AppendStruct(v any) error      { return m.appendErr }
func (m *mockBatch) Column(int) driver.BatchColumn { return nil }
func (m *mockBatch) Columns() []column.Interface   { return nil }
func (m *mockBatch) Flush() error                  { return nil }
func (m *mockBatch) Send() error                   { return m.sendErr }
func (m *mockBatch) IsSent() bool                  { return false }
func (m *mockBatch) Rows() int                     { return m.appendCount }
func (m *mockBatch) Close() error                  { return nil }

type mockRow struct {
	scanErr error
	values  []any
}

func (m *mockRow) Err() error { return m.scanErr }
func (m *mockRow) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	if m.values != nil {
		for i, v := range m.values {
			if i < len(dest) {
				switch d := dest[i].(type) {
				case *int32:
					if val, ok := v.(int32); ok {
						*d = val
					}
				case *int64:
					if val, ok := v.(int64); ok {
						*d = val
					}
				case *float64:
					if val, ok := v.(float64); ok {
						*d = val
					}
				case *float32:
					if val, ok := v.(float32); ok {
						*d = val
					}
				case *string:
					if val, ok := v.(string); ok {
						*d = val
					}
				case *uint8:
					if val, ok := v.(uint8); ok {
						*d = val
					}
				case *[]int32:
					if val, ok := v.([]int32); ok {
						*d = val
					}
				case *[]float32:
					if val, ok := v.([]float32); ok {
						*d = val
					}
				case *[]string:
					if val, ok := v.([]string); ok {
						*d = val
					}
				case *[]time.Time:
					if val, ok := v.([]time.Time); ok {
						*d = val
					}
				}
			}
		}
	}
	return nil
}
func (m *mockRow) ScanStruct(dest any) error { return m.scanErr }

type mockRows struct {
	scanErr    error
	nextCount  int
	closeErr   error
	columns    []string
	errVal     error
	scanValues [][]any
	scanIndex  int
}

func (m *mockRows) Columns() []string                { return m.columns }
func (m *mockRows) ColumnTypes() []driver.ColumnType { return nil }
func (m *mockRows) Next() bool {
	if m.nextCount > 0 {
		m.nextCount--
		return true
	}
	return false
}
func (m *mockRows) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	if m.scanValues != nil && m.scanIndex < len(m.scanValues) {
		vals := m.scanValues[m.scanIndex]
		m.scanIndex++
		for i, v := range vals {
			if i < len(dest) {
				switch d := dest[i].(type) {
				case *string:
					if s, ok := v.(string); ok {
						*d = s
					}
				case *int32:
					if n, ok := v.(int32); ok {
						*d = n
					}
				case *int64:
					if n, ok := v.(int64); ok {
						*d = n
					}
				case *float64:
					if f, ok := v.(float64); ok {
						*d = f
					}
				}
			}
		}
	}
	return nil
}
func (m *mockRows) ScanStruct(dest any) error { return m.scanErr }
func (m *mockRows) Totals(dest ...any) error  { return nil }
func (m *mockRows) Close() error              { return m.closeErr }
func (m *mockRows) Err() error                { return m.errVal }

var _ clickhouse.Conn = (*mockConn)(nil)

func newTestClient(conn *mockConn) *ch.Client {
	return &ch.Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test_db"},
		Logger: zerolog.New(io.Discard),
	}
}

func newTestParams() *ch.QueryParams {
	return &ch.QueryParams{
		AccountIDs: []string{"li_123"},
		DateFrom:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Timezone:   "UTC",
		DayCount:   31,
	}
}

func Test_GetSummary_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				int32(100), int32(200), int32(300), int32(50), int32(80), int32(150),
				int32(10000), int32(5000), int32(8000), int32(40), int32(60), int32(500),
				int32(20000), int32(3000), float64(5.5), float64(3.2),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetSummary(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.PostComments != 100 {
		t.Fatalf("expected PostComments=100, got %d", result.PostComments)
	}
	if result.Followers != 10000 {
		t.Fatalf("expected Followers=10000, got %d", result.Followers)
	}
}

func Test_GetSummary_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetSummary(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetAudienceGrowth_Success(t *testing.T) {
	buckets := []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				uint8(1),
				[]int32{100}, []int32{5},
				[]int32{20}, []int32{1},
				[]int32{120}, []int32{6},
				buckets,
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetAudienceGrowth(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData=1, got %d", result.ShowData)
	}
}

func Test_GetAudienceGrowth_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetAudienceGrowth(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetLastFollowerCounts_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int32(12000), int32(10000), int32(2000)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetLastFollowerCounts(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFollowerCount != 12000 {
		t.Fatalf("expected TotalFollowerCount=12000, got %d", result.TotalFollowerCount)
	}
}

func Test_GetLastFollowerCounts_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetLastFollowerCounts(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetAudienceRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int32(10000), int32(2000), int32(12000), float64(11500.5)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetAudienceRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFollowerCount != 12000 {
		t.Fatalf("expected TotalFollowerCount=12000, got %d", result.TotalFollowerCount)
	}
	if result.AvgFollowerCount != 11500.5 {
		t.Fatalf("expected AvgFollowerCount=11500.5, got %f", result.AvgFollowerCount)
	}
}

func Test_GetAudienceRollup_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetAudienceRollup(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetPageViews_Success(t *testing.T) {
	buckets := []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]int32{100}, []int32{200}, []int32{300},
				[]int32{10}, []int32{20}, []int32{30},
				int32(1), buckets,
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPageViews(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData=1, got %d", result.ShowData)
	}
}

func Test_GetPageViews_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPageViews(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetPageViewsRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int32(5000), int32(3000), int32(2000), float64(250.5)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPageViewsRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalPageViews != 5000 {
		t.Fatalf("expected TotalPageViews=5000, got %d", result.TotalPageViews)
	}
}

func Test_GetPageViewsRollup_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPageViewsRollup(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetPublishingBehaviour_Success(t *testing.T) {
	buckets := []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]int32{100}, []int32{50}, []int32{30}, []int32{80},
				[]float32{5.5}, []int32{20000}, []int32{10},
				[]int32{180}, []int32{8000}, buckets,
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPublishingBehaviour(context.Background(), newTestParams(), []string{"images", "videos"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Likes) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(result.Likes))
	}
}

func Test_GetPublishingBehaviour_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPublishingBehaviour(context.Background(), newTestParams(), []string{"images"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetPublishingBehaviourRollup_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPublishingBehaviourRollup(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetPublishingBehaviourRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 2,
			scanValues: [][]any{
				{"images", int32(10), int32(100), int32(50), int32(30), int32(80), int32(180), int32(5000), int32(3000)},
				{"total", int32(10), int32(100), int32(50), int32(30), int32(80), int32(180), int32(5000), int32(3000)},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	results, err := repo.GetPublishingBehaviourRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(results))
	}
	if results[0].MediaType != "images" {
		t.Fatalf("expected first row media_type='images', got %q", results[0].MediaType)
	}
}

func Test_GetTopPosts_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopPosts(context.Background(), newTestParams(), 10, "total_engagement", nil)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetTopPosts_EmptyResult(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{nextCount: 0},
	}
	repo := NewRepository(newTestClient(conn))
	results, err := repo.GetTopPosts(context.Background(), newTestParams(), 10, "total_engagement", nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func Test_GetTopPosts_WithHashtags(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{nextCount: 0},
	}
	repo := NewRepository(newTestClient(conn))
	results, err := repo.GetTopPosts(context.Background(), newTestParams(), 10, "total_engagement", []string{"#linkedin", "#test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func Test_GetPostsPerDay_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int32(5), int32(8), int32(10), int32(7), int32(6), int32(3), int32(2)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPostsPerDay(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Monday != 5 {
		t.Fatalf("expected Monday=5, got %d", result.Monday)
	}
	if result.Wednesday != 10 {
		t.Fatalf("expected Wednesday=10, got %d", result.Wednesday)
	}
}

func Test_GetPostsPerDay_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostsPerDay(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetTopHashtags_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]string{"#linkedin", "#test"},
				[]int32{500, 200}, []int32{300, 100}, []int32{100, 50},
				[]int32{100, 50}, []int32{10, 5},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetTopHashtags(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Name) != 2 {
		t.Fatalf("expected 2 hashtags, got %d", len(result.Name))
	}
}

func Test_GetTopHashtags_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopHashtags(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetTopHashtagsRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				int32(15), int32(50), int32(500), int32(200), int32(100),
				int32(800), int32(20000), int32(10000),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetTopHashtagsRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalHashtags != 15 {
		t.Fatalf("expected TotalHashtags=15, got %d", result.TotalHashtags)
	}
	if result.TotalEngagement != 800 {
		t.Fatalf("expected TotalEngagement=800, got %d", result.TotalEngagement)
	}
}

func Test_GetTopHashtagsRollup_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopHashtagsRollup(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}

func Test_GetFollowersDemographics_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				"{\"senior\":100}", "{\"tech\":500}", "{\"US\":3000}",
				"{\"SF\":500}", int64(12000),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetFollowersDemographics(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalFollowerCount != 12000 {
		t.Fatalf("expected TotalFollowerCount=12000, got %d", result.TotalFollowerCount)
	}
}

func Test_GetFollowersDemographics_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetFollowersDemographics(context.Background(), newTestParams())
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
