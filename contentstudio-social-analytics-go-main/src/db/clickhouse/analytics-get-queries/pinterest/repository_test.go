package pinterest

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
	if m.queryRowResult != nil {
		return m.queryRowResult
	}
	return &mockRow{scanErr: errors.New("no mock row configured")}
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

func newTestParams() *ch.QueryParams {
	return &ch.QueryParams{
		AccountIDs: []string{"pin_123"},
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

// --- GetSummaryForUser ---

func TestGetSummaryForUser_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 5000  // FollowerCount
				*dest[1].(*int64) = 10000 // Impressions
				*dest[2].(*int64) = 300   // PinClicks
				*dest[3].(*int64) = 150   // OutboundClicks
				*dest[4].(*int64) = 200   // Saves
				*dest[5].(*int64) = 650   // TotalEngagement
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetSummaryForUser(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FollowerCount != 5000 {
		t.Fatalf("expected FollowerCount 5000, got %d", result.FollowerCount)
	}
	if result.Impressions != 10000 {
		t.Fatalf("expected Impressions 10000, got %d", result.Impressions)
	}
	if result.TotalEngagement != 650 {
		t.Fatalf("expected TotalEngagement 650, got %d", result.TotalEngagement)
	}
}

func TestGetSummaryForUser_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetSummaryForUser(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetSummaryForBoard ---

func TestGetSummaryForBoard_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 3000
				*dest[1].(*int64) = 8000
				*dest[2].(*int64) = 200
				*dest[3].(*int64) = 100
				*dest[4].(*int64) = 150
				*dest[5].(*int64) = 450
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetSummaryForBoard(context.Background(), newTestParams(), "('board_1','board_2')")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FollowerCount != 3000 {
		t.Fatalf("expected FollowerCount 3000, got %d", result.FollowerCount)
	}
}

func TestGetSummaryForBoard_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetSummaryForBoard(context.Background(), newTestParams(), "('board_1')")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetFollowerTrendForUser ---

func TestGetFollowerTrendForUser_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{100, 105, 110}
				*dest[2].(*[]int32) = []int32{5, 5, 5}
				*dest[3].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetFollowerTrendForUser(context.Background(), newTestParams(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.FollowersDaily) != 3 {
		t.Fatalf("expected 3 daily entries, got %d", len(result.FollowersDaily))
	}
	if len(result.Buckets) != 3 {
		t.Fatalf("expected 3 buckets, got %d", len(result.Buckets))
	}
}

func TestGetFollowerTrendForUser_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetFollowerTrendForUser(context.Background(), newTestParams(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetFollowerTrendForBoard ---

func TestGetFollowerTrendForBoard_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{50, 55}
				*dest[2].(*[]int32) = []int32{5, 5}
				*dest[3].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetFollowerTrendForBoard(context.Background(), newTestParams(), "('board_1')", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
}

func TestGetFollowerTrendForBoard_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetFollowerTrendForBoard(context.Background(), newTestParams(), "('board_1')", false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetImpressionsTrendForUser ---

func TestGetImpressionsTrendForUser_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{500, 600}
				*dest[2].(*[]int32) = []int32{500, 1100}
				*dest[3].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetImpressionsTrendForUser(context.Background(), newTestParams(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.ImpressionsDaily) != 2 {
		t.Fatalf("expected 2 daily entries, got %d", len(result.ImpressionsDaily))
	}
}

func TestGetImpressionsTrendForUser_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetImpressionsTrendForUser(context.Background(), newTestParams(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetImpressionsTrendForBoard ---

func TestGetImpressionsTrendForBoard_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetImpressionsTrendForBoard(context.Background(), newTestParams(), "('board_1')", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetEngagementTrendForUser ---

func TestGetEngagementTrendForUser_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{10, 20} // SavesDaily
				*dest[2].(*[]int32) = []int32{10, 30} // SavesTotal
				*dest[3].(*[]int32) = []int32{5, 8}   // OutboundClicksDaily
				*dest[4].(*[]int32) = []int32{5, 13}  // OutboundClicksTotal
				*dest[5].(*[]int32) = []int32{20, 30} // PinClicksDaily
				*dest[6].(*[]int32) = []int32{20, 50} // PinClicksTotal
				*dest[7].(*[]int32) = []int32{35, 58} // EngagementDaily
				*dest[8].(*[]int32) = []int32{35, 93} // EngagementTotal
				*dest[9].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetEngagementTrendForUser(context.Background(), newTestParams(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.SavesDaily) != 2 {
		t.Fatalf("expected 2 saves daily, got %d", len(result.SavesDaily))
	}
}

func TestGetEngagementTrendForUser_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetEngagementTrendForUser(context.Background(), newTestParams(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetEngagementTrendForBoard ---

func TestGetEngagementTrendForBoard_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetEngagementTrendForBoard(context.Background(), newTestParams(), "('board_1')", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetPinPostingForUser ---

func TestGetPinPostingForUser_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{3, 5, 2}
				*dest[2].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPinPostingForUser(context.Background(), newTestParams(), "all", true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.PinsCount) != 3 {
		t.Fatalf("expected 3 pins counts, got %d", len(result.PinsCount))
	}
}

func TestGetPinPostingForUser_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPinPostingForUser(context.Background(), newTestParams(), "all", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPinPostingForUser_DedupesByPinAndBucket(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, _ = repo.GetPinPostingForUser(context.Background(), newTestParams(), "all", true)
	if !strings.Contains(conn.lastQuery, "GROUP BY pin_id, bucket") {
		t.Fatalf("expected query to dedupe by pin_id and bucket, got: %s", conn.lastQuery)
	}
}

// --- GetPinPostingForBoard ---

func TestGetPinPostingForBoard_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPinPostingForBoard(context.Background(), newTestParams(), "('board_1')", "all", true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPinPostingForBoard_DedupesByPinAndBucket(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, _ = repo.GetPinPostingForBoard(context.Background(), newTestParams(), "('board_1')", "all", true)
	if !strings.Contains(conn.lastQuery, "GROUP BY pin_id, bucket") {
		t.Fatalf("expected query to dedupe by pin_id and bucket, got: %s", conn.lastQuery)
	}
}

// --- GetPinRollupForUser ---

func TestGetPinRollupForUser_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 100    // TotalPins
				*dest[1].(*int64) = 5000   // Impressions
				*dest[2].(*int64) = 200    // PinClicks
				*dest[3].(*int64) = 80     // OutboundClicks
				*dest[4].(*int64) = 150    // Saves
				*dest[5].(*float64) = 75.5 // QuartilePercView
				*dest[6].(*int64) = 1000   // VideoViews
				*dest[7].(*int64) = 500    // Video10sViews
				*dest[8].(*float64) = 12.3 // AvgWatchTime
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPinRollupForUser(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalPins != 100 {
		t.Fatalf("expected TotalPins 100, got %d", result.TotalPins)
	}
	if result.Impressions != 5000 {
		t.Fatalf("expected Impressions 5000, got %d", result.Impressions)
	}
	if result.AvgWatchTime != 12.3 {
		t.Fatalf("expected AvgWatchTime 12.3, got %f", result.AvgWatchTime)
	}
}

func TestGetPinRollupForUser_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPinRollupForUser(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetPinRollupForBoard ---

func TestGetPinRollupForBoard_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPinRollupForBoard(context.Background(), newTestParams(), "('board_1')")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetPinsForUser ---

func TestGetPinsForUser_Success(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 1,
			scanFn: func(idx int, dest ...any) error {
				*dest[0].(*string) = "pin_1"
				*dest[1].(*string) = "My Board"
				*dest[2].(*string) = "https://pinterest.com/pin/1"
				*dest[3].(*string) = "https://embed.link"
				*dest[4].(*string) = "Pin Title"
				*dest[5].(*string) = "Pin description"
				*dest[6].(*string) = "user_1"
				*dest[7].(*string) = "image"
				*dest[8].(*string) = "https://img.url"
				*dest[9].(*string) = "#FF0000"
				*dest[10].(*string) = "standard"
				*dest[11].(*[]string) = []string{"tag1"}
				*dest[12].(*int64) = 600
				*dest[13].(*int64) = 400
				*dest[14].(*time.Time) = createdAt
				*dest[15].(*int64) = 5000
				*dest[16].(*int64) = 200
				*dest[17].(*int64) = 80
				*dest[18].(*int64) = 150
				*dest[19].(*int64) = 430
				*dest[20].(*float64) = 8.6
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPinsForUser(context.Background(), newTestParams(), "impressions", 10, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 pin, got %d", len(result))
	}
	if result[0].PinID != "pin_1" {
		t.Fatalf("expected pin_1, got %q", result[0].PinID)
	}
	if result[0].Impressions != 5000 {
		t.Fatalf("expected Impressions 5000, got %d", result[0].Impressions)
	}
}

func TestGetPinsForUser_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPinsForUser(context.Background(), newTestParams(), "impressions", 10, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPinsForUser_ScanError(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{nextCount: 1, scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPinsForUser(context.Background(), newTestParams(), "impressions", 10, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPinsForUser_EmptyResult(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{nextCount: 0},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPinsForUser(context.Background(), newTestParams(), "impressions", 10, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 pins, got %d", len(result))
	}
}

// --- GetPinsForBoard ---

func TestGetPinsForBoard_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPinsForBoard(context.Background(), newTestParams(), "('board_1')", "impressions", 10, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetPinPerformanceForUser ---

func TestGetPinPerformanceForUser_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{3, 5}     // PinsCount
				*dest[2].(*[]int32) = []int32{100, 200} // PinClicks
				*dest[3].(*[]int32) = []int32{50, 80}   // OutboundClicks
				*dest[4].(*[]int32) = []int32{30, 60}   // Saves
				*dest[5].(*[]int32) = []int32{180, 340} // Engagements
				*dest[6].(*[]int32) = []int32{500, 800} // Impressions
				*dest[7].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPinPerformanceForUser(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.PinsCount) != 2 {
		t.Fatalf("expected 2 pins counts, got %d", len(result.PinsCount))
	}
	if len(result.Impressions) != 2 {
		t.Fatalf("expected 2 impressions, got %d", len(result.Impressions))
	}
}

func TestGetPinPerformanceForUser_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPinPerformanceForUser(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPinPerformanceForUser_DedupesPinData(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, _ = repo.GetPinPerformanceForUser(context.Background(), newTestParams())
	if !strings.Contains(conn.lastQuery, "GROUP BY pp.pin_id, pin_date") {
		t.Fatalf("expected query to dedupe pin_data by pin_id and pin_date, got: %s", conn.lastQuery)
	}
}

// --- GetPinPerformanceForBoard ---

func TestGetPinPerformanceForBoard_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPinPerformanceForBoard(context.Background(), newTestParams(), "('board_1')")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPinPerformanceForBoard_DedupesPinData(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, _ = repo.GetPinPerformanceForBoard(context.Background(), newTestParams(), "('board_1')")
	if !strings.Contains(conn.lastQuery, "GROUP BY pp.pin_id, pin_date") {
		t.Fatalf("expected query to dedupe pin_data by pin_id and pin_date, got: %s", conn.lastQuery)
	}
}
