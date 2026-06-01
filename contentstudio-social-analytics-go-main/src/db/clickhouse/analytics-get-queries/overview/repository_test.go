package overview

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
)

// --- Mock ClickHouse connection ---

type mockConn struct {
	queryErr       error
	queryRows      driver.Rows
	queryRowResult driver.Row
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
	return nil, nil
}
func (m *mockConn) Exec(context.Context, string, ...any) error              { return nil }
func (m *mockConn) AsyncInsert(context.Context, string, bool, ...any) error { return nil }
func (m *mockConn) Ping(context.Context) error                              { return nil }
func (m *mockConn) Stats() driver.Stats                                     { return driver.Stats{} }
func (m *mockConn) Close() error                                            { return nil }

type mockRow struct {
	scanErr error
	scanFn  func(dest ...any) error
}

func (m *mockRow) Err() error { return m.scanErr }
func (m *mockRow) Scan(dest ...any) error {
	if m.scanFn != nil {
		return m.scanFn(dest...)
	}
	return m.scanErr
}
func (m *mockRow) ScanStruct(dest any) error { return m.scanErr }

type mockRows struct {
	nextCount int
	scanErr   error
	errVal    error
	scanFn    func(idx int, dest ...any) error
	scanIndex int
}

func (m *mockRows) Columns() []string                { return nil }
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
	if m.scanFn != nil {
		idx := m.scanIndex
		m.scanIndex++
		return m.scanFn(idx, dest...)
	}
	return nil
}
func (m *mockRows) ScanStruct(dest any) error { return m.scanErr }
func (m *mockRows) Totals(dest ...any) error  { return nil }
func (m *mockRows) Close() error              { return nil }
func (m *mockRows) Err() error                { return m.errVal }

func newTestClient(conn *mockConn) *ch.Client {
	return &ch.Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test_db"},
		Logger: zerolog.New(io.Discard),
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

func TestNewOverviewParams_BasicDateRange(t *testing.T) {
	params, err := NewOverviewParams(
		"2025-01-01", "2025-01-31",
		[]string{"fb_1"}, []string{"ig_1"}, nil, nil, nil, nil,
		"UTC",
		"total_engagement", 20,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.CurrentStart != "2025-01-01" {
		t.Fatalf("expected CurrentStart 2025-01-01, got %q", params.CurrentStart)
	}
	// PHP adds 1 day to end date
	if params.CurrentEnd != "2025-02-01" {
		t.Fatalf("expected CurrentEnd 2025-02-01, got %q", params.CurrentEnd)
	}
	if !params.IncludeFacebook {
		t.Fatal("expected IncludeFacebook true")
	}
	if !params.IncludeInstagram {
		t.Fatal("expected IncludeInstagram true")
	}
	if params.IncludeLinkedIn {
		t.Fatal("expected IncludeLinkedIn false")
	}
}

func TestNewOverviewParams_FullMonthSecondaryPeriod(t *testing.T) {
	// Full month: Jan 1–Jan 31 → secondary should be Dec 1–Dec 31 (+1 day = Jan 1)
	params, err := NewOverviewParams(
		"2025-01-01", "2025-01-31",
		[]string{"fb_1"}, nil, nil, nil, nil, nil,
		"UTC",
		"", 0,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.SecondaryStart != "2024-12-01" {
		t.Fatalf("expected SecondaryStart 2024-12-01, got %q", params.SecondaryStart)
	}
	if params.SecondaryEnd != "2025-01-01" {
		t.Fatalf("expected SecondaryEnd 2025-01-01, got %q", params.SecondaryEnd)
	}
}

func TestNewOverviewParams_NonFullMonthSecondaryPeriod(t *testing.T) {
	// Non-full month: Jan 10–Jan 20 → 11 days → secondary is Dec 30–Jan 10
	params, err := NewOverviewParams(
		"2025-01-10", "2025-01-20",
		[]string{"fb_1"}, nil, nil, nil, nil, nil,
		"UTC",
		"", 0,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.SecondaryEnd != "2025-01-10" {
		t.Fatalf("expected SecondaryEnd 2025-01-10, got %q", params.SecondaryEnd)
	}
	if params.SecondaryStart != "2024-12-30" {
		t.Fatalf("expected SecondaryStart 2024-12-30, got %q", params.SecondaryStart)
	}
}

func TestNewOverviewParams_InvalidStartDate(t *testing.T) {
	_, err := NewOverviewParams("invalid", "2025-01-31", nil, nil, nil, nil, nil, nil, "UTC", "", 0)
	if err == nil {
		t.Fatal("expected error for invalid start date")
	}
}

func TestNewOverviewParams_InvalidEndDate(t *testing.T) {
	_, err := NewOverviewParams("2025-01-01", "invalid", nil, nil, nil, nil, nil, nil, "UTC", "", 0)
	if err == nil {
		t.Fatal("expected error for invalid end date")
	}
}

func TestNewOverviewParams_DefaultSortAndLimit(t *testing.T) {
	params, err := NewOverviewParams("2025-01-01", "2025-01-15", nil, nil, nil, nil, nil, nil, "UTC", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Type != "total_engagement" {
		t.Fatalf("expected default sort 'total_engagement', got %q", params.Type)
	}
	if params.Limit != 20 {
		t.Fatalf("expected default limit 20, got %d", params.Limit)
	}
}

func TestNewOverviewParams_OverviewSortDefaultsToEngagement(t *testing.T) {
	params, err := NewOverviewParams("2025-01-01", "2025-01-15", nil, nil, nil, nil, nil, nil, "UTC", "overview", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Type != "total_engagement" {
		t.Fatalf("expected sort 'total_engagement', got %q", params.Type)
	}
}

func TestNewOverviewParams_AllPlatforms(t *testing.T) {
	params, err := NewOverviewParams(
		"2025-01-01", "2025-01-31",
		[]string{"fb_1"}, []string{"ig_1"}, []string{"li_1"},
		[]string{"tk_1"}, []string{"pt_1"}, []string{"yt_1"},
		"UTC",
		"impressions", 50,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !params.IncludeFacebook || !params.IncludeInstagram || !params.IncludeLinkedIn ||
		!params.IncludeTiktok || !params.IncludePinterest || !params.IncludeYouTube {
		t.Fatal("expected all platforms to be included")
	}
	if params.Type != "impressions" {
		t.Fatalf("expected sort 'impressions', got %q", params.Type)
	}
	if params.Limit != 50 {
		t.Fatalf("expected limit 50, got %d", params.Limit)
	}
}

func TestOverviewParams_DateFilter(t *testing.T) {
	params, _ := NewOverviewParams("2025-01-01", "2025-01-31", nil, nil, nil, nil, nil, nil, "UTC", "", 0)
	filter := params.dateFilter("created_at")
	expected := "created_at >= toDateTime('2025-01-01',0) AND created_at < toDateTime('2025-02-01',0)"
	if filter != expected {
		t.Fatalf("expected %q, got %q", expected, filter)
	}
}

func TestOverviewParams_SecondaryDateFilter(t *testing.T) {
	params, _ := NewOverviewParams("2025-01-01", "2025-01-31", nil, nil, nil, nil, nil, nil, "UTC", "", 0)
	filter := params.secondaryDateFilter("created_at")
	expected := "created_at >= toDateTime('2024-12-01',0) AND created_at < toDateTime('2025-01-01',0)"
	if filter != expected {
		t.Fatalf("expected %q, got %q", expected, filter)
	}
}

func TestOverviewParams_NewSecondaryParams(t *testing.T) {
	params, _ := NewOverviewParams("2025-01-01", "2025-01-31",
		[]string{"fb_1"}, nil, nil, nil, nil, nil, "UTC", "", 0)
	sec := params.NewSecondaryParams()
	if sec.CurrentStart != params.SecondaryStart {
		t.Fatalf("expected CurrentStart=%q, got %q", params.SecondaryStart, sec.CurrentStart)
	}
	if sec.CurrentEnd != params.SecondaryEnd {
		t.Fatalf("expected CurrentEnd=%q, got %q", params.SecondaryEnd, sec.CurrentEnd)
	}
	// Original should be unchanged
	if params.CurrentStart != "2025-01-01" {
		t.Fatal("original params were modified")
	}
}

func TestGetTopPerformingGraph_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				// 24 float64 arrays (6 platforms × 4 metrics)
				for i := 1; i <= 24; i++ {
					*dest[i].(*[]float64) = []float64{1.0, 2.0}
				}
				return nil
			},
		},
	}
	params, _ := NewOverviewParams("2025-01-01", "2025-01-31",
		[]string{"fb_1"}, []string{"ig_1"}, nil, nil, nil, nil, "UTC", "", 0)

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetTopPerformingGraph(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(result.Buckets))
	}
	if len(result.FacebookPostCount) != 2 {
		t.Fatalf("expected 2 fb post counts, got %d", len(result.FacebookPostCount))
	}
}

func TestGetTopPerformingGraph_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	params, _ := NewOverviewParams("2025-01-01", "2025-01-31",
		[]string{"fb_1"}, nil, nil, nil, nil, nil, "UTC", "", 0)

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopPerformingGraph(context.Background(), params)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPlatformData_NoPlatforms(t *testing.T) {
	conn := &mockConn{}
	params, _ := NewOverviewParams("2025-01-01", "2025-01-31",
		nil, nil, nil, nil, nil, nil, "UTC", "", 0)

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPlatformData(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 results for no platforms, got %d", len(result))
	}
}

func TestGetAccountData_NoPlatforms(t *testing.T) {
	conn := &mockConn{}
	params, _ := NewOverviewParams("2025-01-01", "2025-01-31",
		nil, nil, nil, nil, nil, nil, "UTC", "", 0)

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetAccountData(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 results for no platforms, got %d", len(result))
	}
}

func TestGetTopPosts_NoPlatforms(t *testing.T) {
	conn := &mockConn{}
	params, _ := NewOverviewParams("2025-01-01", "2025-01-31",
		nil, nil, nil, nil, nil, nil, "UTC", "", 0)

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetTopPosts(context.Background(), params)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 results for no platforms, got %d", len(result))
	}
}

func TestGetTopPosts_AppliesPerPlatformLimit(t *testing.T) {
	conn := &mockConn{queryRows: &mockRows{}}
	repo := NewRepository(newTestClient(conn))
	params, err := NewOverviewParams(
		"2025-01-01", "2025-01-31",
		[]string{"fb_1"}, []string{"ig_1"},
		nil, nil, nil, nil,
		"UTC",
		"total_engagement", 10,
	)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.GetTopPosts(context.Background(), params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	query := strings.TrimSpace(conn.lastQuery)
	if got := strings.Count(query, "LIMIT 10"); got != 2 {
		t.Fatalf("expected per-platform LIMIT 10 to appear twice, got %d in query: %s", got, query)
	}
	if !strings.HasSuffix(query, "ORDER BY total_engagement DESC") {
		t.Fatalf("expected final query to end with global ordering only, got: %s", query)
	}
}

func TestQueryInstagramTopPosts_UsesStoredEventAtForLatestSnapshot(t *testing.T) {
	conn := &mockConn{queryRows: &mockRows{}}
	repo := NewRepository(newTestClient(conn))
	params, err := NewOverviewParams("2025-01-01", "2025-01-31",
		nil, []string{"ig_1"}, nil, nil, nil, nil, "UTC", "total_engagement", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.GetTopPosts(context.Background(), params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "argMax(media_url, stored_event_at)") {
		t.Fatalf("expected latest media_url by stored_event_at, got: %s", conn.lastQuery)
	}
}

func TestQueryLinkedInTopPosts_UsesSavingTimeForLatestSnapshot(t *testing.T) {
	conn := &mockConn{queryRows: &mockRows{}}
	repo := NewRepository(newTestClient(conn))
	params, err := NewOverviewParams("2025-01-01", "2025-01-31",
		nil, nil, []string{"li_1"}, nil, nil, nil, "UTC", "total_engagement", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.GetTopPosts(context.Background(), params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "argMax(image, saving_time)") {
		t.Fatalf("expected latest image by saving_time, got: %s", conn.lastQuery)
	}
}

func TestQueryTiktokTopPosts_UsesInsertedAtForLatestSnapshot(t *testing.T) {
	conn := &mockConn{queryRows: &mockRows{}}
	repo := NewRepository(newTestClient(conn))
	params, err := NewOverviewParams("2025-01-01", "2025-01-31",
		nil, nil, nil, []string{"tt_1"}, nil, nil, "UTC", "total_engagement", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.GetTopPosts(context.Background(), params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "argMax(embed_link, inserted_at)") {
		t.Fatalf("expected latest embed_link by inserted_at, got: %s", conn.lastQuery)
	}
}

func TestQueryYouTubeTopPosts_UsesInsertedAtForLatestSnapshot(t *testing.T) {
	conn := &mockConn{queryRows: &mockRows{}}
	repo := NewRepository(newTestClient(conn))
	params, err := NewOverviewParams("2025-01-01", "2025-01-31",
		nil, nil, nil, nil, nil, []string{"yt_1"}, "UTC", "total_engagement", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := repo.GetTopPosts(context.Background(), params); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "argMax(thumbnail_url, inserted_at)") {
		t.Fatalf("expected latest thumbnail_url by inserted_at, got: %s", conn.lastQuery)
	}
}

func TestNewOverviewParams_NormalizesKyivTimezone(t *testing.T) {
	params, err := NewOverviewParams("2025-01-01", "2025-01-31", nil, nil, nil, nil, nil, nil, "Europe/Kyiv", "", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if params.Timezone != "Europe/Riga" {
		t.Fatalf("expected Europe/Riga, got %q", params.Timezone)
	}
}

func TestTopPostsGlobalSortCol_Mappings(t *testing.T) {
	tests := []struct {
		sortType string
		want     string
	}{
		{sortType: "total_impressions", want: "greatest(views, reach)"},
		{sortType: "impressions", want: "greatest(views, reach)"},
		{sortType: "reach", want: "reach"},
		{sortType: "likes", want: "likes"},
		{sortType: "comments", want: "comments"},
		{sortType: "shares", want: "shares"},
		{sortType: "views", want: "views"},
		{sortType: "total_engagement", want: "total_engagement"},
		{sortType: "unknown", want: "total_engagement"},
	}

	for _, tt := range tests {
		got := topPostsGlobalSortCol(tt.sortType)
		if got != tt.want {
			t.Fatalf("topPostsGlobalSortCol(%q) = %q, want %q", tt.sortType, got, tt.want)
		}
	}
}
