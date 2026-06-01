package facebook

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
	queryRowResult driver.Row
	queryRows      driver.Rows
	queryErr       error
	lastQuery      string
}

func (m *mockConn) Contributors() []string                        { return nil }
func (m *mockConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (m *mockConn) Select(context.Context, any, string, ...any) error {
	return nil
}
func (m *mockConn) Query(_ context.Context, query string, _ ...any) (driver.Rows, error) {
	m.lastQuery = query
	if m.queryErr != nil {
		return nil, m.queryErr
	}
	return m.queryRows, nil
}
func (m *mockConn) QueryRow(_ context.Context, query string, _ ...any) driver.Row {
	m.lastQuery = query
	return m.queryRowResult
}
func (m *mockConn) PrepareBatch(context.Context, string, ...driver.PrepareBatchOption) (driver.Batch, error) {
	return &mockBatch{}, nil
}
func (m *mockConn) Exec(context.Context, string, ...any) error              { return nil }
func (m *mockConn) AsyncInsert(context.Context, string, bool, ...any) error { return nil }
func (m *mockConn) Ping(context.Context) error                              { return nil }
func (m *mockConn) Stats() driver.Stats                                     { return driver.Stats{} }
func (m *mockConn) Close() error                                            { return nil }

type mockBatch struct{}

func (m *mockBatch) Abort() error                  { return nil }
func (m *mockBatch) Append(v ...any) error         { return nil }
func (m *mockBatch) AppendStruct(any) error        { return nil }
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
	for i, value := range m.values {
		switch d := dest[i].(type) {
		case *int32:
			*d = value.(int32)
		case *float64:
			*d = value.(float64)
		case *[]int32:
			*d = value.([]int32)
		case *[]float64:
			*d = value.([]float64)
		case *[]time.Time:
			*d = value.([]time.Time)
		case *[]string:
			*d = value.([]string)
		case *uint8:
			*d = value.(uint8)
		case *string:
			*d = value.(string)
		}
	}
	return nil
}
func (m *mockRow) ScanStruct(any) error { return m.err }

type mockRows struct {
	values [][]any
	index  int
	err    error
}

func (m *mockRows) Columns() []string                { return nil }
func (m *mockRows) ColumnTypes() []driver.ColumnType { return nil }
func (m *mockRows) Next() bool {
	return m.index < len(m.values)
}
func (m *mockRows) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	row := m.values[m.index]
	m.index++
	for i, value := range row {
		switch d := dest[i].(type) {
		case *string:
			*d = value.(string)
		case *int32:
			*d = value.(int32)
		}
	}
	return nil
}
func (m *mockRows) ScanStruct(any) error { return m.err }
func (m *mockRows) Totals(...any) error  { return nil }
func (m *mockRows) Close() error         { return nil }
func (m *mockRows) Err() error           { return m.err }

var _ clickhouse.Conn = (*mockConn)(nil)

func newTestClient(conn *mockConn) *ch.Client {
	return &ch.Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test"},
		Logger: zerolog.New(io.Discard),
	}
}

func newParams() *ch.QueryParams {
	return &ch.QueryParams{
		AccountIDs: []string{"fb_123"},
		DateFrom:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Timezone:   "UTC",
		DayCount:   31,
	}
}

func TestGetSummary(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				int32(1), int32(2), int32(3), int32(4), int32(5), int32(6), int32(7), int32(8),
				int32(9), int32(10), int32(11), int32(12), int32(13), int32(14), int32(15),
				int32(16), int32(17), int32(18), int32(19),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetSummary(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DocCount != 1 || result.PageFollows != 19 {
		t.Fatalf("unexpected summary result: %+v", result)
	}
	if !strings.Contains(conn.lastQuery, "WHERE (post_id, saving_time) IN (SELECT post_id, saving_time FROM posts)") {
		t.Fatalf("expected summary query to use explicit posts subquery, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "argMin(page_fans, saving_time) AS fan_count") {
		t.Fatalf("expected summary query to use PHP-style fan count dedupe, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "argMin(page_follows, saving_time) AS page_follows") {
		t.Fatalf("expected summary query to use PHP-style page_follows dedupe, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "max(page_post_engagements) AS page_engagements") {
		t.Fatalf("expected summary query to use PHP-style engagement aggregation, got: %s", conn.lastQuery)
	}
}

func TestGetAudienceCountry(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			values: [][]any{
				{"US", int32(10)},
				{"UK", int32(5)},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetAudienceCountry(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["US"] != 10 || result["UK"] != 5 {
		t.Fatalf("unexpected country map: %+v", result)
	}
	if strings.Contains(conn.lastQuery, "a.countries") {
		t.Fatalf("expected country query not to reference invalid alias a.countries, got: %s", conn.lastQuery)
	}
}

func TestGetSummary_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetSummary(context.Background(), newParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPostsSummary(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int32(5), int32(50), int32(20), int32(10), int32(8), int32(1000), int32(800), int32(12)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPostsSummary(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DocCount != 5 {
		t.Fatalf("expected DocCount=5, got %d", result.DocCount)
	}
	if result.TotalEngagement != 50 {
		t.Fatalf("expected TotalEngagement=50, got %d", result.TotalEngagement)
	}
	if !strings.Contains(conn.lastQuery, "AND (post_id, saving_time) IN (SELECT post_id, saving_time FROM posts)") {
		t.Fatalf("expected posts query to use dedup subquery with partition filter, got: %s", conn.lastQuery)
	}
}

func TestGetPostsSummary_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetPostsSummary(context.Background(), newParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetInsightsSummary(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				int32(100), int32(5), int32(5000), int32(2000), int32(3000),
				int32(800), int32(60), int32(10), int32(12000), int32(400), int32(11500),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetInsightsSummary(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FanCount != 12000 {
		t.Fatalf("expected FanCount=12000, got %d", result.FanCount)
	}
	if result.PageImpressions != 5000 {
		t.Fatalf("expected PageImpressions=5000, got %d", result.PageImpressions)
	}
	if !strings.Contains(conn.lastQuery, "argMin(page_fans, saving_time) AS fan_count") {
		t.Fatalf("expected insights query to use PHP-style fan count dedupe, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "argMin(page_follows, saving_time) AS page_follows") {
		t.Fatalf("expected insights query to use PHP-style page_follows dedupe, got: %s", conn.lastQuery)
	}
}

func TestGetInsightsSummary_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetInsightsSummary(context.Background(), newParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTopPosts_UsesSnakeCaseAssetColumns(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{},
	}
	repo := NewRepository(newTestClient(conn))

	_, err := repo.GetTopPosts(context.Background(), newParams(), []string{"images"}, 10, "total_engagement")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(conn.lastQuery, "toString(asset_type) AS asset_type") {
		t.Fatalf("expected asset_type column in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "toString(call_to_action) AS call_to_action") {
		t.Fatalf("expected call_to_action column in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "created_at AS asset_created_at") {
		t.Fatalf("expected created_at column in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "WHERE (post_id, saving_time) IN (SELECT post_id, saving_time FROM posts)") {
		t.Fatalf("expected top posts query to use explicit posts subquery, got: %s", conn.lastQuery)
	}
}

func TestGetReelsAnalytics_UsesCreatedTimeForFacebookPosts(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]time.Time{}, []int32{}, []float64{}, []int32{}, []int32{}, []int32{}, []int32{}, []int32{}, int32(0),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))

	_, err := repo.GetReelsAnalytics(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(conn.lastQuery, "FROM facebook_posts\n    WHERE toString(page_id)") {
		t.Fatalf("expected facebook_posts CTE in reels query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "toDate(created_time, 'UTC')") {
		t.Fatalf("expected reels query to filter facebook_posts by created_time, got: %s", conn.lastQuery)
	}
}

func TestGetVideoInsights_UsesCreatedTimeForFacebookPosts(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]time.Time{}, []float64{}, []float64{}, []float64{}, []int32{}, []int32{}, []int32{}, []int32{}, []int32{}, []int32{}, []int32{},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))

	_, err := repo.GetVideoInsights(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(conn.lastQuery, "FROM facebook_posts\n    WHERE toString(page_id)") {
		t.Fatalf("expected facebook_posts CTE in video query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "toDate(created_time, 'UTC')") {
		t.Fatalf("expected video query to filter facebook_posts by created_time, got: %s", conn.lastQuery)
	}
	if strings.Contains(conn.lastQuery, "media_type = 'videos'") || strings.Contains(conn.lastQuery, "media_type='videos'") {
		t.Fatalf("expected video query not to filter facebook_posts by media_type, got: %s", conn.lastQuery)
	}
}

func TestGetVideoRollup_DoesNotRequireVideoMediaTypeOnFacebookPosts(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				float64(0), float64(0), float64(0), int32(0), int32(0), int32(0), int32(0), int32(0), int32(0), int32(0),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))

	_, err := repo.GetVideoRollup(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(conn.lastQuery, "FROM facebook_posts\n    WHERE toString(page_id)") {
		t.Fatalf("expected facebook_posts CTE in video rollup query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "toDate(created_time, 'UTC')") {
		t.Fatalf("expected video rollup query to filter facebook_posts by created_time, got: %s", conn.lastQuery)
	}
	if strings.Contains(conn.lastQuery, "media_type = 'videos'") || strings.Contains(conn.lastQuery, "media_type='videos'") {
		t.Fatalf("expected video rollup query not to filter facebook_posts by media_type, got: %s", conn.lastQuery)
	}
}

func TestGetAudienceCity_UsesDirectArrayJoinColumns(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			values: [][]any{
				{"Lahore", int32(3)},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))

	result, err := repo.GetAudienceCity(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["Lahore"] != 3 {
		t.Fatalf("unexpected city map: %+v", result)
	}
	if strings.Contains(conn.lastQuery, "a.cities") {
		t.Fatalf("expected city query not to reference invalid alias a.cities, got: %s", conn.lastQuery)
	}
}

func TestGetActiveUsersHours_CastsHighestValueToInt32(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]int32{1, 2, 3},
				[]int32{4, 5, 6},
				int32(6),
				int32(3),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))

	result, err := repo.GetActiveUsersHours(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HighestValue != 6 {
		t.Fatalf("expected highest value 6, got %d", result.HighestValue)
	}
	if !strings.Contains(conn.lastQuery, "toInt32(arrayMax(values)) AS highest_value") {
		t.Fatalf("expected active users hours query to cast highest_value to Int32, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "arrayMap(x -> toInt32(x / count()), sumForEach(values)) AS value") {
		t.Fatalf("expected active users hours query to cast averaged values to Int32, got: %s", conn.lastQuery)
	}
}
