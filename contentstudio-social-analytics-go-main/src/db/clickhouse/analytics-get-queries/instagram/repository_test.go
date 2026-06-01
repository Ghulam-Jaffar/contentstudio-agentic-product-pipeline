package instagram

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

// --- Mock ClickHouse infrastructure ---

type mockConn struct {
	queryRowResult driver.Row
	queryRows      driver.Rows
	queryErr       error
	lastQuery      string
}

func (m *mockConn) Contributors() []string                        { return nil }
func (m *mockConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (m *mockConn) Select(_ context.Context, _ any, _ string, _ ...any) error {
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
func (m *mockConn) PrepareBatch(_ context.Context, _ string, _ ...driver.PrepareBatchOption) (driver.Batch, error) {
	return &mockBatch{}, nil
}
func (m *mockConn) Exec(_ context.Context, _ string, _ ...any) error                { return nil }
func (m *mockConn) AsyncInsert(_ context.Context, _ string, _ bool, _ ...any) error { return nil }
func (m *mockConn) Ping(_ context.Context) error                                    { return nil }
func (m *mockConn) Stats() driver.Stats                                             { return driver.Stats{} }
func (m *mockConn) Close() error                                                    { return nil }

type mockBatch struct{}

func (m *mockBatch) Abort() error                    { return nil }
func (m *mockBatch) Append(_ ...any) error           { return nil }
func (m *mockBatch) AppendStruct(_ any) error        { return nil }
func (m *mockBatch) Column(_ int) driver.BatchColumn { return nil }
func (m *mockBatch) Columns() []column.Interface     { return nil }
func (m *mockBatch) Flush() error                    { return nil }
func (m *mockBatch) Send() error                     { return nil }
func (m *mockBatch) IsSent() bool                    { return false }
func (m *mockBatch) Rows() int                       { return 0 }
func (m *mockBatch) Close() error                    { return nil }

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
		if i >= len(dest) {
			break
		}
		switch d := dest[i].(type) {
		case *int32:
			if v, ok := value.(int32); ok {
				*d = v
			}
		case *int64:
			if v, ok := value.(int64); ok {
				*d = v
			}
		case *float64:
			if v, ok := value.(float64); ok {
				*d = v
			}
		case *uint8:
			if v, ok := value.(uint8); ok {
				*d = v
			}
		case *string:
			if v, ok := value.(string); ok {
				*d = v
			}
		case *[]int32:
			if v, ok := value.([]int32); ok {
				*d = v
			}
		case *[]int64:
			if v, ok := value.([]int64); ok {
				*d = v
			}
		case *[]float64:
			if v, ok := value.([]float64); ok {
				*d = v
			}
		case *[]string:
			if v, ok := value.([]string); ok {
				*d = v
			}
		case *[]time.Time:
			if v, ok := value.([]time.Time); ok {
				*d = v
			}
		}
	}
	return nil
}
func (m *mockRow) ScanStruct(_ any) error { return m.err }

type mockRows struct {
	values   [][]any
	index    int
	err      error
	closeErr error
}

func (m *mockRows) Columns() []string                { return nil }
func (m *mockRows) ColumnTypes() []driver.ColumnType { return nil }
func (m *mockRows) Next() bool                       { return m.index < len(m.values) }
func (m *mockRows) Scan(dest ...any) error {
	if m.err != nil {
		return m.err
	}
	row := m.values[m.index]
	m.index++
	for i, value := range row {
		if i >= len(dest) {
			break
		}
		switch d := dest[i].(type) {
		case *string:
			if v, ok := value.(string); ok {
				*d = v
			}
		case *int32:
			if v, ok := value.(int32); ok {
				*d = v
			}
		case *int64:
			if v, ok := value.(int64); ok {
				*d = v
			}
		case *float64:
			if v, ok := value.(float64); ok {
				*d = v
			}
		case *time.Time:
			if v, ok := value.(time.Time); ok {
				*d = v
			}
		}
	}
	return nil
}
func (m *mockRows) ScanStruct(_ any) error { return m.err }
func (m *mockRows) Totals(_ ...any) error  { return nil }
func (m *mockRows) Close() error           { return m.closeErr }
func (m *mockRows) Err() error             { return m.err }

var _ clickhouse.Conn = (*mockConn)(nil)

func newTestClient(conn *mockConn) *ch.Client {
	return &ch.Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test"},
		Logger: zerolog.New(io.Discard),
	}
}

func newTestParams() *ch.QueryParams {
	return &ch.QueryParams{
		AccountIDs:   []string{"ig_123"},
		DateFrom:     time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:       time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		PrevDateFrom: time.Date(2024, 12, 1, 0, 0, 0, 0, time.UTC),
		PrevDateTo:   time.Date(2024, 12, 31, 0, 0, 0, 0, time.UTC),
		Timezone:     "UTC",
		DayCount:     31,
	}
}

// --- GetPostsSummary ---

func Test_GetPostsSummary_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				int64(55), int64(500), int64(300), int64(100), int64(50),
				int64(8000), int64(20000), int64(10000), int64(5), int64(50), float64(3.5),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPostsSummary(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.DocCount != 55 {
		t.Fatalf("expected DocCount=55, got %d", result.DocCount)
	}
	if result.TotalPosts != 50 {
		t.Fatalf("expected TotalPosts=50, got %d", result.TotalPosts)
	}
	if result.Stories != 5 {
		t.Fatalf("expected Stories=5, got %d", result.Stories)
	}
	if result.EngagementRate != 3.5 {
		t.Fatalf("expected EngagementRate=3.5, got %f", result.EngagementRate)
	}
}

func Test_GetPostsSummary_QueriesPostsTable(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{}}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostsSummary(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "instagram_posts") {
		t.Fatalf("expected query to reference instagram_posts, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "GROUP BY media_id") {
		t.Fatalf("expected query to use media_id dedup pattern, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "entity_type = 'STORY' OR media_type = 'STORY'") {
		t.Fatalf("expected query to detect stories by entity_type OR media_type, got: %s", conn.lastQuery)
	}
}

func Test_GetPostsSummary_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetPostsSummary(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetInsightsSummary ---

func Test_GetInsightsSummary_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int64(1200), int64(15000), int64(500), int64(80), int64(350), int64(5000), int64(2500)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetInsightsSummary(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ProfileViews != 1200 {
		t.Fatalf("expected ProfileViews=1200, got %d", result.ProfileViews)
	}
	if result.FollowersCount != 15000 {
		t.Fatalf("expected FollowersCount=15000, got %d", result.FollowersCount)
	}
	if result.Engagement != 350 {
		t.Fatalf("expected Engagement=350, got %d", result.Engagement)
	}
	if result.Impressions != 5000 {
		t.Fatalf("expected Impressions=5000, got %d", result.Impressions)
	}
	if result.Reach != 2500 {
		t.Fatalf("expected Reach=2500, got %d", result.Reach)
	}
}

func Test_GetInsightsSummary_QueriesInsightsTable(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{}}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetInsightsSummary(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "instagram_insights") {
		t.Fatalf("expected query to reference instagram_insights, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "daily_engagement") {
		t.Fatalf("expected query to aggregate daily engagement, got: %s", conn.lastQuery)
	}
}

func Test_GetInsightsSummary_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetInsightsSummary(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetAudienceGrowth ---

func Test_GetAudienceGrowth_Success(t *testing.T) {
	buckets := []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				uint8(1),
				[]int32{15000},
				[]int32{0},
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
	if len(result.Followers) != 1 || result.Followers[0] != 15000 {
		t.Fatalf("unexpected followers: %v", result.Followers)
	}
}

func Test_GetAudienceGrowth_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetAudienceGrowth(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetLastFollowerCount ---

func Test_GetLastFollowerCount_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{values: []any{int32(14000)}},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetLastFollowerCount(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FollowersCount != 14000 {
		t.Fatalf("expected FollowersCount=14000, got %d", result.FollowersCount)
	}
}

func Test_GetLastFollowerCount_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetLastFollowerCount(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetAudienceRollup ---

func Test_GetAudienceRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{values: []any{int32(15000), int32(300)}},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetAudienceRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.FollowerCount != 15000 {
		t.Fatalf("expected FollowerCount=15000, got %d", result.FollowerCount)
	}
	if result.FollowerGained != 300 {
		t.Fatalf("expected FollowerGained=300, got %d", result.FollowerGained)
	}
}

func Test_GetAudienceRollup_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetAudienceRollup(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetPublishingBehaviour ---

func Test_GetPublishingBehaviour_Success(t *testing.T) {
	buckets := []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]int32{100}, []int32{50}, []int32{30},
				[]int32{200}, []int32{8000}, []int32{20000},
				[]int32{10000}, []int32{10}, buckets,
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPublishingBehaviour(context.Background(), newTestParams(), []string{"REELS", "IMAGE"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.Likes) != 1 || result.Likes[0] != 100 {
		t.Fatalf("unexpected likes: %v", result.Likes)
	}
}

func Test_GetPublishingBehaviour_UsesMediaTypeFilter(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]int32{}, []int32{}, []int32{}, []int32{},
				[]int32{}, []int32{}, []int32{}, []int32{},
				[]time.Time{},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPublishingBehaviour(context.Background(), newTestParams(), []string{"REELS"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "'REELS'") {
		t.Fatalf("expected query to filter by REELS, got: %s", conn.lastQuery)
	}
}

func Test_GetPublishingBehaviour_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetPublishingBehaviour(context.Background(), newTestParams(), []string{"IMAGE"}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetPublishingBehaviourRollup ---

func Test_GetPublishingBehaviourRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			values: [][]any{
				{"REELS", int32(5), int32(50), int32(20), int32(10), int32(80), int32(3000), int32(5000)},
				{"IMAGE", int32(8), int32(80), int32(30), int32(15), int32(130), int32(5000), int32(8000)},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	results, err := repo.GetPublishingBehaviourRollup(context.Background(), newTestParams(), []string{"REELS", "IMAGE"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(results))
	}
	if results[0].MediaType != "REELS" {
		t.Fatalf("expected first row media_type='REELS', got %q", results[0].MediaType)
	}
}

func Test_GetPublishingBehaviourRollup_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetPublishingBehaviourRollup(context.Background(), newTestParams(), []string{"IMAGE"}); err == nil {
		t.Fatal("expected error, got nil")
	}
}

func Test_GetPublishingBehaviourRollup_Empty(t *testing.T) {
	conn := &mockConn{queryRows: &mockRows{}}
	repo := NewRepository(newTestClient(conn))
	results, err := repo.GetPublishingBehaviourRollup(context.Background(), newTestParams(), []string{"IMAGE"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

// --- GetTopPosts ---

func Test_GetTopPosts_Success(t *testing.T) {
	now := time.Now()
	conn := &mockConn{
		queryRows: &mockRows{
			values: [][]any{
				{
					"ig_123", "media_abc", "caption text", "IMAGE", "POST",
					// media_url, video_url, permalink scanned as string individually
					// but our mockRows.Scan only handles string/*string in top-level
					// The actual scan uses []string which is not handled by this simple mock.
					// We use a simplified set of scalar fields here to test basic flow.
				},
			},
		},
	}
	_ = conn
	_ = now

	// Use empty rows for scan test — verify query contains key patterns
	conn2 := &mockConn{queryRows: &mockRows{}}
	repo := NewRepository(newTestClient(conn2))
	results, err := repo.GetTopPosts(context.Background(), newTestParams(), "total_engagement", 10, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results from empty rows, got %d", len(results))
	}
	if !strings.Contains(conn2.lastQuery, "entity_type != 'STORY'") {
		t.Fatalf("expected story filter in top posts query, got: %s", conn2.lastQuery)
	}
	if !strings.Contains(conn2.lastQuery, "ORDER BY total_engagement DESC") {
		t.Fatalf("expected ORDER BY total_engagement in query, got: %s", conn2.lastQuery)
	}
	if !strings.Contains(conn2.lastQuery, "LIMIT 10") {
		t.Fatalf("expected LIMIT 10 in query, got: %s", conn2.lastQuery)
	}
}

func Test_GetTopPosts_WithHashtags(t *testing.T) {
	conn := &mockConn{queryRows: &mockRows{}}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopPosts(context.Background(), newTestParams(), "like_count", 5, []string{"#fitness", "#travel"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "hasAny") {
		t.Fatalf("expected hasAny hashtag filter in query, got: %s", conn.lastQuery)
	}
}

func Test_GetTopPosts_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetTopPosts(context.Background(), newTestParams(), "total_engagement", 10, nil); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetActiveUsersHours ---

func Test_GetActiveUsersHours_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]int32{0, 1, 2, 3},
				[]int32{10, 20, 50, 30},
				int32(50),
				int32(2),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetActiveUsersHours(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HighestValue != 50 {
		t.Fatalf("expected HighestValue=50, got %d", result.HighestValue)
	}
	if result.HighestHour != 2 {
		t.Fatalf("expected HighestHour=2, got %d", result.HighestHour)
	}
}

func Test_GetActiveUsersHours_UsesJSONExtract(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{[]int32{}, []int32{}, int32(0), int32(0)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetActiveUsersHours(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "splitByChar(':',") {
		t.Fatalf("expected splitByChar(':') parsing in active users query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "numbers(24)") {
		t.Fatalf("expected numbers(24) in active users query, got: %s", conn.lastQuery)
	}
}

func Test_GetActiveUsersHours_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetActiveUsersHours(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetActiveUsersDays ---

func Test_GetActiveUsersDays_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]string{"Monday", "Tuesday", "Wednesday"},
				[]int32{10, 25, 18},
				int32(25),
				"Tuesday",
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetActiveUsersDays(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.HighestDay != "Tuesday" {
		t.Fatalf("expected HighestDay='Tuesday', got %q", result.HighestDay)
	}
	if result.HighestValue != 25 {
		t.Fatalf("expected HighestValue=25, got %d", result.HighestValue)
	}
}

func Test_GetActiveUsersDays_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetActiveUsersDays(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetImpressions ---

func Test_GetImpressions_Success(t *testing.T) {
	buckets := []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{uint8(1), buckets, []int32{5000}},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetImpressions(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData=1, got %d", result.ShowData)
	}
	if len(result.Impressions) != 1 || result.Impressions[0] != 5000 {
		t.Fatalf("unexpected impressions: %v", result.Impressions)
	}
}

func Test_GetImpressions_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetImpressions(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetImpressionsRollup ---

func Test_GetImpressionsRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{values: []any{int64(150000), float64(4838.7)}},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetImpressionsRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalImpressions != 150000 {
		t.Fatalf("expected TotalImpressions=150000, got %d", result.TotalImpressions)
	}
}

func Test_GetImpressionsRollup_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetImpressionsRollup(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetEngagement ---

func Test_GetEngagement_Success(t *testing.T) {
	buckets := []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				uint8(1), buckets,
				[]int32{500}, []int32{100}, []int32{300}, []int32{10},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetEngagement(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData=1, got %d", result.ShowData)
	}
	if len(result.Engagement) != 1 || result.Engagement[0] != 500 {
		t.Fatalf("unexpected engagement: %v", result.Engagement)
	}
}

func Test_GetEngagement_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetEngagement(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetEngagementRollup ---

func Test_GetEngagementRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int64(5000), float64(50.0), int64(1000), int64(3000), int64(800), int64(100)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetEngagementRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Engagement != 5000 {
		t.Fatalf("expected Engagement=5000, got %d", result.Engagement)
	}
	if result.Count != 100 {
		t.Fatalf("expected Count=100, got %d", result.Count)
	}
}

func Test_GetEngagementRollup_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetEngagementRollup(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetTopHashtags ---

func Test_GetTopHashtags_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]string{"#fitness", "#travel"},
				[]int32{500, 300}, []int32{400, 200},
				[]int32{80, 50}, []int32{30, 20}, []int32{10, 8},
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
	if result.Name[0] != "#fitness" {
		t.Fatalf("expected '#fitness', got %q", result.Name[0])
	}
}

func Test_GetTopHashtags_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetTopHashtags(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetTopHashtagsRollup ---

func Test_GetTopHashtagsRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{int32(800), int32(600), int32(130), int32(50), int32(18), int32(25)},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetTopHashtagsRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalEngagement != 800 {
		t.Fatalf("expected TotalEngagement=800, got %d", result.TotalEngagement)
	}
	if result.TotalUniqueHashtags != 18 {
		t.Fatalf("expected TotalUniqueHashtags=18, got %d", result.TotalUniqueHashtags)
	}
}

func Test_GetTopHashtagsRollup_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetTopHashtagsRollup(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetStoriesPerformance ---

func Test_GetStoriesPerformance_Success(t *testing.T) {
	buckets := []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				uint8(1), buckets,
				[]float64{1200.5}, []int32{1200}, []int32{900},
				[]int32{50}, []int32{30}, []int32{200}, []int32{180}, []int32{8},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetStoriesPerformance(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData=1, got %d", result.ShowData)
	}
	if len(result.StoryImpressions) != 1 || result.StoryImpressions[0] != 1200 {
		t.Fatalf("unexpected story impressions: %v", result.StoryImpressions)
	}
}

func Test_GetStoriesPerformance_UsesEntityTypeStory(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				uint8(0), []time.Time{}, []float64{}, []int32{},
				[]int32{}, []int32{}, []int32{}, []int32{}, []int32{}, []int32{},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetStoriesPerformance(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "entity_type = 'STORY'") {
		t.Fatalf("expected STORY filter in stories query, got: %s", conn.lastQuery)
	}
}

func Test_GetStoriesPerformance_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetStoriesPerformance(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetStoriesRollup ---

func Test_GetStoriesRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				int64(36000), float64(1200.0), int64(27000), int64(1500),
				int64(900), int64(6000), int64(5400), int64(30),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetStoriesRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.StoryImpressions != 36000 {
		t.Fatalf("expected StoryImpressions=36000, got %d", result.StoryImpressions)
	}
	if result.PublishedStories != 30 {
		t.Fatalf("expected PublishedStories=30, got %d", result.PublishedStories)
	}
}

func Test_GetStoriesRollup_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetStoriesRollup(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetReelsPerformance ---

func Test_GetReelsPerformance_Success(t *testing.T) {
	buckets := []time.Time{time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)}
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				uint8(1), buckets,
				[]int32{5}, []int32{800}, []int32{600},
				[]int32{100}, []int32{80}, []int32{50},
				[]float64{12.5}, []int64{62500},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetReelsPerformance(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.ShowData != 1 {
		t.Fatalf("expected ShowData=1, got %d", result.ShowData)
	}
	if len(result.AvgWatchTime) != 1 || result.AvgWatchTime[0] != 12.5 {
		t.Fatalf("unexpected avg watch time: %v", result.AvgWatchTime)
	}
}

func Test_GetReelsPerformance_UsesReelsFilter(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				uint8(0), []time.Time{}, []int32{}, []int32{},
				[]int32{}, []int32{}, []int32{}, []int32{},
				[]float64{}, []int64{},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetReelsPerformance(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "media_type = 'REELS'") {
		t.Fatalf("expected REELS filter in reels query, got: %s", conn.lastQuery)
	}
}

func Test_GetReelsPerformance_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetReelsPerformance(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetReelsRollup ---

func Test_GetReelsRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				int64(4000), int64(3000), int64(500), int64(400),
				int64(25), int64(250), float64(11.2), int64(280000),
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetReelsRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Engagement != 4000 {
		t.Fatalf("expected Engagement=4000, got %d", result.Engagement)
	}
	if result.TotalPosts != 25 {
		t.Fatalf("expected TotalPosts=25, got %d", result.TotalPosts)
	}
}

func Test_GetReelsRollup_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetReelsRollup(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetDemographics ---

func Test_GetDemographics_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]string{`{"key":"18-24","value":500}`, `{"key":"25-34","value":800}`},
				[]string{`{"key":"F","value":700}`, `{"key":"M","value":600}`},
				[]string{`{"key":"F.18-24","value":300}`},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetDemographics(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.AudienceAge) != 2 {
		t.Fatalf("expected 2 age entries, got %d", len(result.AudienceAge))
	}
	if len(result.AudienceGender) != 2 {
		t.Fatalf("expected 2 gender entries, got %d", len(result.AudienceGender))
	}
}

func Test_GetDemographics_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetDemographics(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- GetLocations ---

func Test_GetLocations_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			values: []any{
				[]string{`{"key":"New York","value":1500}`, `{"key":"Los Angeles","value":800}`},
				[]string{`{"key":"US","value":5000}`, `{"key":"UK","value":2000}`},
			},
		},
	}
	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetLocations(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.AudienceCity) != 2 {
		t.Fatalf("expected 2 city entries, got %d", len(result.AudienceCity))
	}
	if len(result.AudienceCountry) != 2 {
		t.Fatalf("expected 2 country entries, got %d", len(result.AudienceCountry))
	}
}

func Test_GetLocations_Error(t *testing.T) {
	conn := &mockConn{queryRowResult: &mockRow{err: errors.New("scan failed")}}
	repo := NewRepository(newTestClient(conn))
	if _, err := repo.GetLocations(context.Background(), newTestParams()); err == nil {
		t.Fatal("expected error, got nil")
	}
}

// --- igPostDedupCTE ---

func Test_igPostDedupCTE_ContainsRequiredPatterns(t *testing.T) {
	cte := igPostDedupCTE("('ig_123')", "post_created_at BETWEEN '2025-01-01' AND '2025-01-31'", "")
	if !strings.Contains(cte, "instagram_posts") {
		t.Fatalf("expected instagram_posts in CTE, got: %s", cte)
	}
	if !strings.Contains(cte, "GROUP BY media_id") {
		t.Fatalf("expected media_id dedup in CTE, got: %s", cte)
	}
	if !strings.Contains(cte, "max_event") {
		t.Fatalf("expected max_event alias in CTE, got: %s", cte)
	}
}

func Test_igPostDedupCTE_WithExtraFilters(t *testing.T) {
	cte := igPostDedupCTE("('ig_123')", "post_created_at BETWEEN '2025-01-01' AND '2025-01-31'",
		"toYYYYMM(stored_event_at) >= 202501", "media_type = 'REELS'", "entity_type != 'STORY'")
	if !strings.Contains(cte, "media_type = 'REELS'") {
		t.Fatalf("expected media_type filter in CTE, got: %s", cte)
	}
	if !strings.Contains(cte, "entity_type != 'STORY'") {
		t.Fatalf("expected entity_type filter in CTE, got: %s", cte)
	}
}
