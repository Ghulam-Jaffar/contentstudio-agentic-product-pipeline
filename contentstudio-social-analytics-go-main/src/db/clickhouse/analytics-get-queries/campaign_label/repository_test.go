package campaign_label

import (
	"context"
	"errors"
	"io"
	"reflect"
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

// --- Mock ClickHouse connection ---

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
	for i, v := range m.values {
		if i >= len(dest) {
			break
		}
		switch d := dest[i].(type) {
		case *int32:
			*d = v.(int32)
		case *float64:
			*d = v.(float64)
		case *string:
			*d = v.(string)
		case *[]int32:
			*d = v.([]int32)
		case *[]string:
			*d = v.([]string)
		}
	}
	return nil
}
func (m *mockRow) ScanStruct(any) error { return m.err }

type mockColumnType struct {
	name string
}

func (m *mockColumnType) Name() string             { return m.name }
func (m *mockColumnType) Nullable() bool           { return false }
func (m *mockColumnType) ScanType() reflect.Type   { return nil }
func (m *mockColumnType) DatabaseTypeName() string { return "" }
func (m *mockColumnType) ColumnType() string       { return "" }

type mockRows struct {
	values      [][]any
	index       int
	err         error
	columnNames []string
	columnTypes []driver.ColumnType
}

func (m *mockRows) Columns() []string { return m.columnNames }
func (m *mockRows) ColumnTypes() []driver.ColumnType {
	return m.columnTypes
}
func (m *mockRows) Next() bool {
	return m.index < len(m.values)
}
func (m *mockRows) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	row := m.values[m.index]
	m.index++
	for i, v := range row {
		if i >= len(dest) {
			break
		}
		switch d := dest[i].(type) {
		case *string:
			*d = v.(string)
		case *int32:
			*d = v.(int32)
		case *float64:
			*d = v.(float64)
		case *[]int32:
			*d = v.([]int32)
		case *[]string:
			*d = v.([]string)
		case *interface{}:
			*d = v
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
		Config: config.ClickHouseConfig{Database: "test_db"},
		Logger: zerolog.New(io.Discard),
	}
}

func newTestParams() *ch.QueryParams {
	return &ch.QueryParams{
		AccountIDs: []string{"fb_123"},
		DateFrom:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Timezone:   "UTC",
		DayCount:   31,
	}
}

// --- Tests ---

func TestNewRepository(t *testing.T) {
	client := newTestClient(&mockConn{})
	repo := NewRepository(client)
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}

func TestDateFilter(t *testing.T) {
	start := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	result := dateFilter("created_time", start, end)
	if !strings.Contains(result, "2025-01-01") || !strings.Contains(result, "2025-01-31") {
		t.Fatalf("expected date range in filter, got %q", result)
	}
	if !strings.Contains(result, "created_time") {
		t.Fatalf("expected field name in filter, got %q", result)
	}
}

func TestFormatPostIDs(t *testing.T) {
	tests := []struct {
		name     string
		ids      []string
		expected string
	}{
		{name: "empty", ids: nil, expected: "['']"},
		{name: "single", ids: []string{"post1"}, expected: "['post1']"},
		{name: "multiple", ids: []string{"p1", "p2", "p3"}, expected: "['p1','p2','p3']"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := formatPostIDs(tc.ids)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestGetSummary_EmptyPostIDs(t *testing.T) {
	repo := NewRepository(newTestClient(&mockConn{}))
	result, err := repo.GetSummary(context.Background(), nil, newTestParams(), map[string]bool{"facebook": true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalPosts != 0 {
		t.Fatalf("expected 0 total_posts for empty IDs, got %d", result.TotalPosts)
	}
}

func TestGetSummary_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int32(10), int32(500), int32(2000), float64(0.25)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetSummary(context.Background(), []string{"post1", "post2"}, newTestParams(), map[string]bool{
		"facebook":  true,
		"instagram": true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalPosts != 10 {
		t.Fatalf("expected TotalPosts 10, got %d", result.TotalPosts)
	}
	if result.TotalEngagement != 500 {
		t.Fatalf("expected TotalEngagement 500, got %d", result.TotalEngagement)
	}
	if result.TotalImpressions != 2000 {
		t.Fatalf("expected TotalImpressions 2000, got %d", result.TotalImpressions)
	}
	if result.TotalEngagementRatePerImpression != 0.25 {
		t.Fatalf("expected rate 0.25, got %f", result.TotalEngagementRatePerImpression)
	}
}

func TestGetSummary_NoPlatformsEnabled(t *testing.T) {
	repo := NewRepository(newTestClient(&mockConn{}))
	result, err := repo.GetSummary(context.Background(), []string{"post1"}, newTestParams(), map[string]bool{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalPosts != 0 {
		t.Fatalf("expected 0 total_posts when no platforms enabled, got %d", result.TotalPosts)
	}
}

func TestGetSummary_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{err: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetSummary(context.Background(), []string{"post1"}, newTestParams(), map[string]bool{"facebook": true})
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSummary_QueryContainsPlatformCTEs(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int32(0), int32(0), int32(0), float64(0)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	_, _ = repo.GetSummary(context.Background(), []string{"post1"}, newTestParams(), map[string]bool{
		"facebook":  true,
		"instagram": true,
		"youtube":   true,
	})

	if !strings.Contains(conn.lastQuery, "facebook_posts") {
		t.Fatalf("expected facebook_posts in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "instagram_posts") {
		t.Fatalf("expected instagram_posts in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "youtube_videos") {
		t.Fatalf("expected youtube_videos in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "UNION ALL") {
		t.Fatalf("expected UNION ALL in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "facebook_post_ids") {
		t.Fatalf("expected facebook_post_ids fallback CTE in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "fp.video_id IN (SELECT raw_post_id FROM facebook_input_ids)") {
		t.Fatalf("expected video_id fallback in facebook query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "arrayExists(id -> endsWith(toString(fp.permalink), id), ids)") {
		t.Fatalf("expected permalink fallback in facebook query, got: %s", conn.lastQuery)
	}
}

func TestGetBreakdownData_EmptyObjects(t *testing.T) {
	repo := NewRepository(newTestClient(&mockConn{}))
	results, err := repo.GetBreakdownData(context.Background(), nil, newTestParams(), "current")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results for empty objects, got %d", len(results))
	}
}

func TestGetBreakdownData_Success(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			values: [][]any{
				{"campaign1", "current", int32(5), int32(100), int32(500)},
				{"campaign2", "current", int32(3), int32(80), int32(300)},
			},
		},
	}
	objects := map[string][]string{
		"campaign1": {"post1", "post2"},
		"campaign2": {"post3"},
	}
	repo := NewRepository(newTestClient(conn))
	results, err := repo.GetBreakdownData(context.Background(), objects, newTestParams(), "current")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 results, got %d", len(results))
	}
	if results[0].ID != "campaign1" {
		t.Fatalf("expected campaign1, got %q", results[0].ID)
	}
	if results[0].TotalPosts != 5 {
		t.Fatalf("expected 5 posts, got %d", results[0].TotalPosts)
	}
	if !strings.Contains(conn.lastQuery, "matched_post_ids") {
		t.Fatalf("expected matched_post_ids dedupe CTE in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "fp.video_id = fi.raw_post_id") {
		t.Fatalf("expected raw mongo id for facebook video_id match, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "ARRAY JOIN arrayFilter(") {
		t.Fatalf("expected permalink-only normalization in query, got: %s", conn.lastQuery)
	}
}

func TestGetBreakdownData_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	objects := map[string][]string{"c1": {"p1"}}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetBreakdownData(context.Background(), objects, newTestParams(), "current")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetBreakdownData_ScanError(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			values: [][]any{{"x", "y", int32(1), int32(2), int32(3)}},
			err:    errors.New("scan failed"),
		},
	}
	objects := map[string][]string{"c1": {"p1"}}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetBreakdownData(context.Background(), objects, newTestParams(), "current")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetInsightsData_EmptyObjects(t *testing.T) {
	repo := NewRepository(newTestClient(&mockConn{}))
	results, err := repo.GetInsightsData(context.Background(), nil, newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestGetInsightsData_Success(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			values: [][]any{
				{"campaign1", []int32{10, 20}, []int32{100, 200}, []int32{1, 2}, []string{"2025-01-01", "2025-01-02"}},
			},
		},
	}
	objects := map[string][]string{"campaign1": {"post1"}}
	repo := NewRepository(newTestClient(conn))
	results, err := repo.GetInsightsData(context.Background(), objects, newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 1 {
		t.Fatalf("expected 1 result, got %d", len(results))
	}
	if results[0].ID != "campaign1" {
		t.Fatalf("expected campaign1, got %q", results[0].ID)
	}
	if len(results[0].TotalEngagement) != 2 {
		t.Fatalf("expected 2 engagement entries, got %d", len(results[0].TotalEngagement))
	}
}

func TestGetInsightsData_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	objects := map[string][]string{"c1": {"p1"}}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetInsightsData(context.Background(), objects, newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPlannerAnalytics_EmptyPostIDs(t *testing.T) {
	repo := NewRepository(newTestClient(&mockConn{}))
	result, err := repo.GetPlannerAnalytics(context.Background(), nil, "facebook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map, got %d entries", len(result))
	}
}

func TestGetPlannerAnalytics_UnknownPlatform(t *testing.T) {
	repo := NewRepository(newTestClient(&mockConn{}))
	result, err := repo.GetPlannerAnalytics(context.Background(), []string{"post1"}, "unknown_platform")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected empty map for unknown platform, got %d entries", len(result))
	}
}

func TestGetPlannerAnalytics_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPlannerAnalytics(context.Background(), []string{"post1"}, "facebook")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPlannerAnalytics_FacebookSuccess(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			columnNames: []string{"engagement", "engagement_tooltip"},
			columnTypes: []driver.ColumnType{
				&mockColumnType{name: "engagement"},
				&mockColumnType{name: "engagement_tooltip"},
			},
			values: [][]any{
				{interface{}(int64(500)), interface{}("tooltip text")},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPlannerAnalytics(context.Background(), []string{"post1"}, "facebook")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, ok := result["engagement"]; !ok {
		t.Fatal("expected 'engagement' key in result")
	}
	if !strings.Contains(conn.lastQuery, "facebook_posts") {
		t.Fatalf("expected facebook_posts in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "facebook_post_ids") {
		t.Fatalf("expected facebook_post_ids CTE for facebook platform, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "arrayExists(id -> endsWith(toString(fp.permalink), id), ids)") {
		t.Fatalf("expected permalink fallback in facebook planner query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "engagement_rate") {
		t.Fatalf("expected engagement_rate in facebook planner query, got: %s", conn.lastQuery)
	}
}

func TestGetPlannerAnalytics_PlatformQueries(t *testing.T) {
	platforms := map[string]string{
		"instagram": "instagram_posts",
		"linkedin":  "linkedin_posts",
		"tiktok":    "tiktok_posts",
		"youtube":   "youtube_videos",
		"pinterest": "pinterest_pins",
	}

	for platform, expectedTable := range platforms {
		t.Run(platform, func(t *testing.T) {
			conn := &mockConn{
				queryRows: &mockRows{
					columnNames: []string{"engagement"},
					columnTypes: []driver.ColumnType{&mockColumnType{name: "engagement"}},
				},
			}
			repo := NewRepository(newTestClient(conn))
			_, err := repo.GetPlannerAnalytics(context.Background(), []string{"post1"}, platform)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if !strings.Contains(conn.lastQuery, expectedTable) {
				t.Fatalf("expected %s in %s query, got: %s", expectedTable, platform, conn.lastQuery)
			}
			if !strings.Contains(conn.lastQuery, "engagement_rate") {
				t.Fatalf("expected engagement_rate in %s planner query, got: %s", platform, conn.lastQuery)
			}
		})
	}
}

func TestBuildPostIdPairsCTE(t *testing.T) {
	objects := map[string][]string{
		"campaign1": {"post1", "post2"},
		"label1":    {"post3"},
	}
	cte, allIDs := buildPostIdPairsCTE(objects)
	if !strings.Contains(cte, "WITH pairs AS") {
		t.Fatal("expected WITH pairs AS in CTE")
	}
	if !strings.Contains(cte, "postIds AS") {
		t.Fatal("expected postIds AS in CTE")
	}
	if len(allIDs) != 3 {
		t.Fatalf("expected 3 post IDs, got %d", len(allIDs))
	}
}

func TestBuildPostIdPairsCTE_Empty(t *testing.T) {
	cte, allIDs := buildPostIdPairsCTE(map[string][]string{})
	if !strings.Contains(cte, "('','')") {
		t.Fatal("expected empty pair fallback in CTE")
	}
	if len(allIDs) != 0 {
		t.Fatalf("expected 0 post IDs, got %d", len(allIDs))
	}
}
