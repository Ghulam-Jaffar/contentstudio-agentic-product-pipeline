package youtube

import (
	"context"
	"errors"
	"io"
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
		AccountIDs: []string{"yt_123"},
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

// --- GetActivitySummary ---

func TestGetActivitySummary_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 50000  // WatchTime
				*dest[1].(*float64) = 4.5  // AvgViewDuration
				*dest[2].(*int64) = 1200   // Likes
				*dest[3].(*int64) = 30     // Dislikes
				*dest[4].(*int64) = 250    // Comments
				*dest[5].(*int64) = 180    // Shares
				*dest[6].(*int64) = 1660   // Engagement
				*dest[7].(*int64) = 25000  // Views
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetActivitySummary(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.WatchTime != 50000 {
		t.Fatalf("expected WatchTime 50000, got %d", result.WatchTime)
	}
	if result.AvgViewDuration != 4.5 {
		t.Fatalf("expected AvgViewDuration 4.5, got %f", result.AvgViewDuration)
	}
	if result.Likes != 1200 {
		t.Fatalf("expected Likes 1200, got %d", result.Likes)
	}
	if result.Engagement != 1660 {
		t.Fatalf("expected Engagement 1660, got %d", result.Engagement)
	}
	if result.Views != 25000 {
		t.Fatalf("expected Views 25000, got %d", result.Views)
	}
}

func TestGetActivitySummary_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetActivitySummary(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetSubscriberSummary ---

func TestGetSubscriberSummary_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 15000
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetSubscriberSummary(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Subscribers != 15000 {
		t.Fatalf("expected Subscribers 15000, got %d", result.Subscribers)
	}
}

func TestGetSubscriberSummary_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetSubscriberSummary(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetVideoCount ---

func TestGetVideoCount_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 42
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetVideoCount(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.VideoCount != 42 {
		t.Fatalf("expected VideoCount 42, got %d", result.VideoCount)
	}
}

func TestGetVideoCount_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetVideoCount(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetSubscriberTrend ---

func TestGetSubscriberTrend_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{10, 15, 20}
				*dest[2].(*[]int32) = []int32{100, 115, 135}
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
	result, err := repo.GetSubscriberTrend(context.Background(), newTestParams(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.SubscribersGainedDaily) != 3 {
		t.Fatalf("expected 3 daily entries, got %d", len(result.SubscribersGainedDaily))
	}
	if len(result.Buckets) != 3 {
		t.Fatalf("expected 3 buckets, got %d", len(result.Buckets))
	}
}

func TestGetSubscriberTrend_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetSubscriberTrend(context.Background(), newTestParams(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetLatestSubscriberCount ---

func TestGetLatestSubscriberCount_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int32) = 14500
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetLatestSubscriberCount(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.SubscriberCount != 14500 {
		t.Fatalf("expected SubscriberCount 14500, got %d", result.SubscriberCount)
	}
}

func TestGetLatestSubscriberCount_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetLatestSubscriberCount(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetEngagementTrend ---

func TestGetEngagementTrend_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{10, 20}   // LikesDaily
				*dest[2].(*[]int32) = []int32{10, 30}   // LikesTotal
				*dest[3].(*[]int32) = []int32{1, 2}     // DislikesDaily
				*dest[4].(*[]int32) = []int32{1, 3}     // DislikesTotal
				*dest[5].(*[]int32) = []int32{5, 8}     // SharesDaily
				*dest[6].(*[]int32) = []int32{5, 13}    // SharesTotal
				*dest[7].(*[]int32) = []int32{3, 4}     // CommentsDaily
				*dest[8].(*[]int32) = []int32{3, 7}     // CommentsTotal
				*dest[9].(*[]int32) = []int32{19, 34}   // EngagementDaily
				*dest[10].(*[]int32) = []int32{19, 53}  // EngagementTotal
				*dest[11].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetEngagementTrend(context.Background(), newTestParams(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.LikeDaily) != 2 {
		t.Fatalf("expected 2 daily entries, got %d", len(result.LikeDaily))
	}
}

func TestGetEngagementTrend_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetEngagementTrend(context.Background(), newTestParams(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetViewsTrend ---

func TestGetViewsTrend_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{100, 200}  // SubscriberViewsDaily
				*dest[2].(*[]int32) = []int32{100, 300}  // SubscriberViewsTotal
				*dest[3].(*[]int32) = []int32{50, 80}    // NonSubscriberViewsDaily
				*dest[4].(*[]int32) = []int32{50, 130}   // NonSubscriberViewsTotal
				*dest[5].(*[]int32) = []int32{150, 280}  // VideoViewsDaily
				*dest[6].(*[]int32) = []int32{150, 430}  // VideoViewsTotal
				*dest[7].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetViewsTrend(context.Background(), newTestParams(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.VideoViewsDaily) != 2 {
		t.Fatalf("expected 2 daily entries, got %d", len(result.VideoViewsDaily))
	}
}

func TestGetViewsTrend_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetViewsTrend(context.Background(), newTestParams(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetWatchTimeTrend ---

func TestGetWatchTimeTrend_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{500, 600}    // WatchTimeDaily
				*dest[2].(*[]int32) = []int32{500, 1100}   // WatchTimeTotal
				*dest[3].(*[]int32) = []int32{100, 120}    // RedWatchTimeDaily
				*dest[4].(*[]int32) = []int32{100, 220}    // RedWatchTimeTotal
				*dest[5].(*[]float64) = []float64{4.5, 5.0} // AverageWatchTime
				*dest[6].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetWatchTimeTrend(context.Background(), newTestParams(), true)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.SubscriberWatchTimeDaily) != 2 {
		t.Fatalf("expected 2 daily entries, got %d", len(result.SubscriberWatchTimeDaily))
	}
	if len(result.AverageWatchTime) != 2 {
		t.Fatalf("expected 2 avg watch time entries, got %d", len(result.AverageWatchTime))
	}
}

func TestGetWatchTimeTrend_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetWatchTimeTrend(context.Background(), newTestParams(), true)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetFindVideo ---

func TestGetFindVideo_Success(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 2,
			scanFn: func(idx int, dest ...any) error {
				switch idx {
				case 0:
					*dest[0].(*string) = "YouTube search"
					*dest[1].(*int64) = 5000
					*dest[2].(*float64) = 62.5
				case 1:
					*dest[0].(*string) = "External"
					*dest[1].(*int64) = 3000
					*dest[2].(*float64) = 37.5
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetFindVideo(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 traffic sources, got %d", len(result))
	}
	if result[0].Name != "YouTube search" {
		t.Fatalf("expected 'YouTube search', got %q", result[0].Name)
	}
	if result[0].Value != 5000 {
		t.Fatalf("expected Value 5000, got %d", result[0].Value)
	}
}

func TestGetFindVideo_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetFindVideo(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetFindVideo_ScanError(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{nextCount: 1, scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetFindVideo(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetFindVideo_EmptyResult(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{nextCount: 0},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetFindVideo(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 results, got %d", len(result))
	}
}

// --- GetVideoSharing ---

func TestGetVideoSharing_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				// 31 int64 fields in SharedInsightsRow
				for i := 0; i < 31; i++ {
					*dest[i].(*int64) = int64(i + 1)
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetVideoSharing(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) == 0 {
		t.Fatal("expected non-empty sharing results")
	}
}

func TestGetVideoSharing_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetVideoSharing(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetTopVideos ---

func TestGetTopVideos_Success(t *testing.T) {
	publishedAt := time.Date(2025, 1, 10, 14, 0, 0, 0, time.UTC)
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 1,
			scanFn: func(idx int, dest ...any) error {
				*dest[0].(*string) = "video_1"
				*dest[1].(*string) = "My Video"
				*dest[2].(*string) = "Video description"
				*dest[3].(*int64) = 300 // Duration
				*dest[4].(*string) = "https://thumb.url"
				*dest[5].(*string) = "video"
				*dest[6].(*string) = "https://iframe.url"
				*dest[7].(*string) = "https://share.url"
				*dest[8].(*int64) = 1500  // Engagement
				*dest[9].(*int64) = 1000  // Likes
				*dest[10].(*int64) = 20   // Dislikes
				*dest[11].(*int64) = 8000 // Views
				*dest[12].(*int64) = 200  // RedViews
				*dest[13].(*int64) = 50   // Favorites
				*dest[14].(*int64) = 180  // Comments
				*dest[15].(*int64) = 250  // SubscribersGained
				*dest[16].(*int64) = 300  // Shares
				*dest[17].(*int64) = 5000 // MinutesWatched
				*dest[18].(*int64) = 800  // RedMinutesWatched
				*dest[19].(*float64) = 3.75       // AvgViewDuration
				*dest[20].(*float64) = 45.5       // AvgViewPercentage
				*dest[21].(*float64) = 18.75      // EngagementRate
				*dest[22].(*time.Time) = publishedAt
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetTopVideos(context.Background(), newTestParams(), "views", 10, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 video, got %d", len(result))
	}
	if result[0].VideoID != "video_1" {
		t.Fatalf("expected video_1, got %q", result[0].VideoID)
	}
	if result[0].Views != 8000 {
		t.Fatalf("expected Views 8000, got %d", result[0].Views)
	}
	if result[0].EngagementRate != 18.75 {
		t.Fatalf("expected EngagementRate 18.75, got %f", result[0].EngagementRate)
	}
}

func TestGetTopVideos_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopVideos(context.Background(), newTestParams(), "views", 10, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTopVideos_ScanError(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{nextCount: 1, scanErr: errors.New("scan failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopVideos(context.Background(), newTestParams(), "views", 10, false)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTopVideos_EmptyResult(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{nextCount: 0},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetTopVideos(context.Background(), newTestParams(), "views", 10, false)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 videos, got %d", len(result))
	}
}

// --- GetPerformanceEngagement ---

func TestGetPerformanceEngagement_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				*dest[2].(*[]int32) = []int32{5, 8}    // Count
				*dest[3].(*[]int32) = []int32{10, 20}  // Likes
				*dest[4].(*[]int32) = []int32{1, 2}    // Dislikes
				*dest[5].(*[]int32) = []int32{3, 4}    // Shares
				*dest[6].(*[]int32) = []int32{5, 8}    // Comments
				*dest[7].(*[]int32) = []int32{24, 42}  // Engagement
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPerformanceEngagement(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.Buckets) != 2 {
		t.Fatalf("expected 2 buckets, got %d", len(result.Buckets))
	}
}

func TestGetPerformanceEngagement_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPerformanceEngagement(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetPerformanceViews ---

func TestGetPerformanceViews_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*uint8) = 1
				*dest[1].(*[]int32) = []int32{10}    // Count
				*dest[2].(*[]int32) = []int32{1000}  // SubscriberViews
				*dest[3].(*[]int32) = []int32{500}   // NonSubscriberViews
				*dest[4].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPerformanceViews(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData 1, got %d", result.ShowData)
	}
	if len(result.Count) != 1 {
		t.Fatalf("expected 1 count entry, got %d", len(result.Count))
	}
	if len(result.Buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(result.Buckets))
	}
}

func TestGetPerformanceViews_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPerformanceViews(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
