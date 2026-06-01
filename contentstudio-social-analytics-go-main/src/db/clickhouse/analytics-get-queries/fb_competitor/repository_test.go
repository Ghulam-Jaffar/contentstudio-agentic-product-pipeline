package fb_competitor

import (
	"context"
	"errors"
	"io"
	"reflect"
	"strings"
	"testing"

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

type mockColumnType struct {
	name     string
	scanType reflect.Type
}

func (m *mockColumnType) Name() string             { return m.name }
func (m *mockColumnType) Nullable() bool           { return false }
func (m *mockColumnType) ScanType() reflect.Type   { return m.scanType }
func (m *mockColumnType) DatabaseTypeName() string { return "" }
func (m *mockColumnType) ColumnType() string       { return "" }

type mockRows struct {
	values      [][]any
	index       int
	err         error
	columnNames []string
	columnTypes []driver.ColumnType
}

func (m *mockRows) Columns() []string                { return m.columnNames }
func (m *mockRows) ColumnTypes() []driver.ColumnType { return m.columnTypes }
func (m *mockRows) Next() bool                       { return m.index < len(m.values) }
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
		reflect.ValueOf(dest[i]).Elem().Set(reflect.ValueOf(v))
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

func newParams() *CompetitorQueryParams {
	return &CompetitorQueryParams{
		PageIDs: []string{"fb_123", "fb_456"},
		Accounts: map[string]AccountInfo{
			"fb_123": {Image: "img1.jpg", Name: "Page One", State: "active", Slug: "page-one"},
			"fb_456": {Image: "img2.jpg", Name: "Page Two", State: "active", Slug: "page-two"},
		},
		StartDate: "2025-01-01 00:00:01",
		EndDate:   "2025-01-31 23:59:59",
		DaysDiff:  30,
	}
}

func emptyMockRows() *mockRows {
	return &mockRows{
		columnNames: []string{"facebook_id"},
		columnTypes: []driver.ColumnType{&mockColumnType{name: "facebook_id", scanType: reflect.TypeOf("")}},
	}
}

// --- Helper function tests ---

func TestAccountFilter(t *testing.T) {
	result := accountFilter("facebook_id", []string{"fb_1", "fb_2"})
	if !strings.Contains(result, "facebook_id IN") {
		t.Fatalf("expected facebook_id IN clause, got %q", result)
	}
	if !strings.Contains(result, "fb_1") || !strings.Contains(result, "fb_2") {
		t.Fatalf("expected both IDs in result, got %q", result)
	}
}

func TestAccountValues(t *testing.T) {
	tests := []struct {
		name string
		ids  []string
		want string
	}{
		{name: "empty", ids: nil, want: "((''))"},
		{name: "single", ids: []string{"fb_1"}, want: "(('fb_1'))"},
		{name: "multiple", ids: []string{"a", "b"}, want: "(('a'),('b'))"},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := accountValues(tc.ids)
			if got != tc.want {
				t.Fatalf("expected %q, got %q", tc.want, got)
			}
		})
	}
}

func TestDateFilter(t *testing.T) {
	result := dateFilter("2025-01-01", "2025-01-31", false)
	if !strings.Contains(result, "created_at BETWEEN") {
		t.Fatalf("expected created_at BETWEEN for posts, got %q", result)
	}

	result = dateFilter("2025-01-01", "2025-01-31", true)
	if !strings.Contains(result, "inserted_at BETWEEN") {
		t.Fatalf("expected inserted_at BETWEEN for insights, got %q", result)
	}
}

func TestDateFilterField(t *testing.T) {
	result := dateFilterField("my_field", "2025-01-01", "2025-01-31")
	if !strings.Contains(result, "my_field BETWEEN") {
		t.Fatalf("expected my_field BETWEEN, got %q", result)
	}
}

func TestConstantConditions(t *testing.T) {
	accounts := map[string]AccountInfo{
		"fb_1": {Name: "Test Page", State: "active"},
	}
	result := constantConditions(accounts, "name", "facebook_id")
	if !strings.Contains(result, "multiIf(") {
		t.Fatalf("expected multiIf, got %q", result)
	}
	if !strings.Contains(result, "Test Page") {
		t.Fatalf("expected account name in result, got %q", result)
	}
}

func TestDayMapping(t *testing.T) {
	result := dayMapping("dow")
	if !strings.Contains(result, "Monday") || !strings.Contains(result, "Sunday") {
		t.Fatalf("expected day names in result, got %q", result)
	}
}

func TestHourMapping(t *testing.T) {
	result := hourMapping("hour")
	if !strings.Contains(result, "12:00 AM") || !strings.Contains(result, "11:00 PM") {
		t.Fatalf("expected hour labels in result, got %q", result)
	}
}

func TestAccountInfoFieldValue(t *testing.T) {
	info := AccountInfo{Image: "img.jpg", Name: "Test", State: "active", Slug: "test-slug"}
	tests := map[string]string{
		"image": "img.jpg",
		"name":  "Test",
		"state": "active",
		"slug":  "test-slug",
		"other": "",
	}
	for field, expected := range tests {
		got := info.fieldValue(field)
		if got != expected {
			t.Fatalf("fieldValue(%q) = %q, want %q", field, got, expected)
		}
	}
}

func TestPrevPeriod(t *testing.T) {
	p := newParams()
	prev := p.PrevPeriod()
	if prev.StartDate == p.StartDate {
		t.Fatal("expected prev StartDate to differ from current")
	}
	if prev.EndDate == p.EndDate {
		t.Fatal("expected prev EndDate to differ from current")
	}
	if len(prev.PageIDs) != len(p.PageIDs) {
		t.Fatal("expected same PageIDs in prev period")
	}
}

func TestSanitizeString(t *testing.T) {
	got := sanitizeString("it's a test")
	if strings.Contains(got, "'") && !strings.Contains(got, "\\'") {
		t.Fatalf("expected escaped single quotes, got %q", got)
	}
}

func TestSanitizeOrderBy(t *testing.T) {
	got := sanitizeOrderBy("`field`\"name\"'val'")
	if strings.ContainsAny(got, "`\"'") {
		t.Fatalf("expected no backtick/quote chars, got %q", got)
	}
}

// --- Repository method tests ---

func TestNewRepository(t *testing.T) {
	repo := NewRepository(newTestClient(&mockConn{}))
	if repo == nil {
		t.Fatal("expected non-nil repository")
	}
}

func TestGetDataTableMetrics_QueryContainsExpectedClauses(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetDataTableMetrics(context.Background(), newParams(), "followersCount")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "facebook_competitor_posts") {
		t.Fatalf("expected posts table reference, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "facebook_competitor_insights") {
		t.Fatalf("expected insights table reference, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "ORDER BY") {
		t.Fatalf("expected ORDER BY clause, got: %s", conn.lastQuery)
	}
}

func TestGetDataTableMetrics_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("connection failed")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetDataTableMetrics(context.Background(), newParams(), "followersCount")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPostingActivityGraphByTypes_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostingActivityGraphByTypes(context.Background(), newParams(), "avgTotalEngagements")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "media_type") {
		t.Fatalf("expected media_type in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "facebook_competitor_posts") {
		t.Fatalf("expected posts table, got: %s", conn.lastQuery)
	}
}

func TestGetPostingActivityGraphByTypes_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostingActivityGraphByTypes(context.Background(), newParams(), "avgTotalEngagements")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPostingActivityBySpecificType_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostingActivityBySpecificType(context.Background(), newParams(), "video", "followersCount")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "video") {
		t.Fatalf("expected media_type in query, got: %s", conn.lastQuery)
	}
}

func TestGetPostingActivityBySpecificType_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostingActivityBySpecificType(context.Background(), newParams(), "video", "followersCount")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTopAndLeastPerformingPosts_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopAndLeastPerformingPosts(context.Background(), newParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "top_5_posts") {
		t.Fatalf("expected top_5_posts category, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "least_5_posts") {
		t.Fatalf("expected least_5_posts category, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "facebook_competitor_media_assets") {
		t.Fatalf("expected media_assets table, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "LIMIT 5 BY facebook_id") {
		t.Fatalf("expected LIMIT 5 BY facebook_id, got: %s", conn.lastQuery)
	}
}

func TestGetTopAndLeastPerformingPosts_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopAndLeastPerformingPosts(context.Background(), newParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTopHashtags_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopHashtags(context.Background(), newParams(), 7)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "hashtags") {
		t.Fatalf("expected hashtags in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "LIMIT 7") {
		t.Fatalf("expected LIMIT 7, got: %s", conn.lastQuery)
	}
}

func TestGetTopHashtags_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopHashtags(context.Background(), newParams(), 7)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetIndividualHashtagData_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetIndividualHashtagData(context.Background(), newParams(), "marketing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "marketing") {
		t.Fatalf("expected hashtag in query, got: %s", conn.lastQuery)
	}
}

func TestGetIndividualHashtagData_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetIndividualHashtagData(context.Background(), newParams(), "marketing")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetBiographyData_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetBiographyData(context.Background(), newParams(), "biography_length")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "biography") {
		t.Fatalf("expected biography in query, got: %s", conn.lastQuery)
	}
	if !strings.Contains(conn.lastQuery, "lengthUTF8(biography)") {
		t.Fatalf("expected lengthUTF8(biography) not nested in aggregate, got: %s", conn.lastQuery)
	}
}

func TestGetBiographyData_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetBiographyData(context.Background(), newParams(), "biography_length")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetFollowersGrowthComparison_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetFollowersGrowthComparison(context.Background(), newParams(), "followers_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "facebook_competitor_insights") {
		t.Fatalf("expected insights table, got: %s", conn.lastQuery)
	}
}

func TestGetFollowersGrowthComparison_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetFollowersGrowthComparison(context.Background(), newParams(), "followers_count")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPostReactDistribution_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostReactDistribution(context.Background(), newParams(), "fb_123", "followers_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "fb_123") {
		t.Fatalf("expected facebook_id in query, got: %s", conn.lastQuery)
	}
}

func TestGetPostReactDistribution_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostReactDistribution(context.Background(), newParams(), "fb_123", "followers_count")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPostReactDistributionByCompany_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostReactDistributionByCompany(context.Background(), newParams(), "fb_123", "followers_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "fb_123") {
		t.Fatalf("expected facebook_id in query, got: %s", conn.lastQuery)
	}
}

func TestGetPostReactDistributionByCompany_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostReactDistributionByCompany(context.Background(), newParams(), "fb_123", "followers_count")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPostTypeDistribution_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostTypeDistribution(context.Background(), newParams(), "followers_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "media_type") {
		t.Fatalf("expected media_type in query, got: %s", conn.lastQuery)
	}
}

func TestGetPostTypeDistribution_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostTypeDistribution(context.Background(), newParams(), "followers_count")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPostEngagementOverTime_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostEngagementOverTime(context.Background(), newParams(), "fb_123", "followers_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "fb_123") {
		t.Fatalf("expected facebook_id in query, got: %s", conn.lastQuery)
	}
}

func TestGetPostEngagementOverTime_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostEngagementOverTime(context.Background(), newParams(), "fb_123", "followers_count")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPostEngagementByCompetitor_QueryContains(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostEngagementByCompetitor(context.Background(), newParams(), "followers_count")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(conn.lastQuery, "post_engagement") {
		t.Fatalf("expected post_engagement in query, got: %s", conn.lastQuery)
	}
}

func TestGetPostEngagementByCompetitor_Error(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("fail")}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPostEngagementByCompetitor(context.Background(), newParams(), "followers_count")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQueryRows_ReturnsEmptySliceForNoRows(t *testing.T) {
	conn := &mockConn{queryRows: emptyMockRows()}
	repo := NewRepository(newTestClient(conn))
	results, err := repo.queryRows(context.Background(), "SELECT 1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if results == nil {
		t.Fatal("expected non-nil empty slice")
	}
	if len(results) != 0 {
		t.Fatalf("expected 0 results, got %d", len(results))
	}
}

func TestQueryRows_ScanError(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			values:      [][]any{{"value1"}},
			columnNames: []string{"col"},
			columnTypes: []driver.ColumnType{&mockColumnType{name: "col", scanType: reflect.TypeOf("")}},
			err:         errors.New("scan failed"),
		},
	}
	repo := NewRepository(newTestClient(conn))
	_, err := repo.queryRows(context.Background(), "SELECT 1")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestQueryRows_Success(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			values:      [][]any{{"fb_123"}, {"fb_456"}},
			columnNames: []string{"facebook_id"},
			columnTypes: []driver.ColumnType{&mockColumnType{name: "facebook_id", scanType: reflect.TypeOf("")}},
		},
	}
	repo := NewRepository(newTestClient(conn))
	results, err := repo.queryRows(context.Background(), "SELECT facebook_id")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Fatalf("expected 2 rows, got %d", len(results))
	}
	if results[0]["facebook_id"] != "fb_123" {
		t.Fatalf("expected fb_123, got %v", results[0]["facebook_id"])
	}
}
