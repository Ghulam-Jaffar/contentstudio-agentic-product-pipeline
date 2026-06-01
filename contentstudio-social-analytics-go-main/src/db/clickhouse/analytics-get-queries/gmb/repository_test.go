package gmb

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	ch "github.com/d4interactive/contentstudio-social-analytics-go/src/db/clickhouse"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
)

// --- Mock ClickHouse connection ---

type mockConn struct {
	queryErr       error
	queryRows      driver.Rows
	queryRowResult driver.Row
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
	return &mockRow{scanErr: errors.New("no mock row configured")}
}
func (m *mockConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	return nil, nil
}
func (m *mockConn) Exec(ctx context.Context, query string, args ...any) error { return nil }
func (m *mockConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return nil
}
func (m *mockConn) Ping(ctx context.Context) error { return nil }
func (m *mockConn) Stats() driver.Stats            { return driver.Stats{} }
func (m *mockConn) Close() error                   { return nil }

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
	nextCount  int
	scanErr    error
	errVal     error
	scanFn     func(idx int, dest ...any) error
	scanIndex  int
}

func (m *mockRows) Columns() []string                { return nil }
func (m *mockRows) ColumnTypes() []driver.ColumnType  { return nil }
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
		AccountIDs: []string{"loc_123"},
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

func TestGmbDateFilter(t *testing.T) {
	tests := []struct {
		name       string
		ids        string
		dateFilter string
		expected   string
	}{
		{
			name:       "single id",
			ids:        "('accounts/123/locations/456')",
			dateFilter: "toDateTime(created_at, 0, 'UTC') BETWEEN toDateTime('2025-01-01', 0) AND toDateTime('2025-01-31', 0)",
			expected:   "gmb_id IN ('accounts/123/locations/456') AND toDateTime(created_at, 0, 'UTC') BETWEEN toDateTime('2025-01-01', 0) AND toDateTime('2025-01-31', 0)",
		},
		{
			name:       "multiple ids",
			ids:        "('accounts/1/locations/1','accounts/2/locations/2')",
			dateFilter: "created_at >= '2025-01-01'",
			expected:   "gmb_id IN ('accounts/1/locations/1','accounts/2/locations/2') AND created_at >= '2025-01-01'",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			result := gmbDateFilter(tc.ids, tc.dateFilter)
			if result != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, result)
			}
		})
	}
}

func TestGetSummary_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				if len(dest) != 10 {
					return errors.New("expected 10 scan destinations")
				}
				*dest[0].(*int64) = 1000  // total_impressions
				*dest[1].(*int64) = 400   // search_impressions
				*dest[2].(*int64) = 600   // maps_impressions
				*dest[3].(*int64) = 150   // website_clicks
				*dest[4].(*int64) = 80    // call_clicks
				*dest[5].(*int64) = 40    // direction_requests
				*dest[6].(*int64) = 20    // other_actions
				*dest[7].(*int64) = 25    // total_reviews
				*dest[8].(*float64) = 4.5 // average_rating
				*dest[9].(*int64) = 12    // total_posts
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetSummary(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalImpressions != 1000 {
		t.Fatalf("expected TotalImpressions 1000, got %d", result.TotalImpressions)
	}
	if result.SearchImpressions != 400 {
		t.Fatalf("expected SearchImpressions 400, got %d", result.SearchImpressions)
	}
	if result.AverageRating != 4.5 {
		t.Fatalf("expected AverageRating 4.5, got %f", result.AverageRating)
	}
	if result.TotalPosts != 12 {
		t.Fatalf("expected TotalPosts 12, got %d", result.TotalPosts)
	}
}

func TestGetSummary_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetSummary(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetImpressions_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				if len(dest) != 7 {
					return errors.New("expected 7 scan destinations")
				}
				*dest[0].(*[]int64) = []int64{10, 20, 30}     // desktop_maps_daily
				*dest[1].(*[]int64) = []int64{15, 25, 35}     // desktop_search_daily
				*dest[2].(*[]int64) = []int64{5, 10, 15}      // mobile_maps_daily
				*dest[3].(*[]int64) = []int64{8, 12, 18}      // mobile_search_daily
				*dest[4].(*[]int64) = []int64{38, 67, 98}     // total_impressions_daily
				*dest[5].(*int64) = 203                        // show_data
				*dest[6].(*[]time.Time) = []time.Time{         // buckets
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetImpressions(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.DesktopMapsDaily) != 3 {
		t.Fatalf("expected 3 daily values, got %d", len(result.DesktopMapsDaily))
	}
	if result.ShowData != 203 {
		t.Fatalf("expected ShowData 203, got %d", result.ShowData)
	}
	if len(result.Buckets) != 3 {
		t.Fatalf("expected 3 buckets, got %d", len(result.Buckets))
	}
}

func TestGetImpressions_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetImpressions(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetImpressionsRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 500    // total_impressions
				*dest[1].(*int64) = 100    // desktop_maps
				*dest[2].(*int64) = 150    // desktop_search
				*dest[3].(*int64) = 120    // mobile_maps
				*dest[4].(*int64) = 130    // mobile_search
				*dest[5].(*float64) = 16.1 // avg_impressions
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetImpressionsRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalImpressions != 500 {
		t.Fatalf("expected TotalImpressions 500, got %d", result.TotalImpressions)
	}
}

func TestGetActions_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*[]int64) = []int64{5, 10}   // call_clicks_daily
				*dest[1].(*[]int64) = []int64{20, 30}  // website_clicks_daily
				*dest[2].(*[]int64) = []int64{3, 7}    // direction_requests_daily
				*dest[3].(*[]int64) = []int64{1, 2}    // other_actions_daily
				*dest[4].(*int64) = 78                  // show_data
				*dest[5].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetActions(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.CallClicksDaily) != 2 {
		t.Fatalf("expected 2 daily values, got %d", len(result.CallClicksDaily))
	}
	if result.ShowData != 78 {
		t.Fatalf("expected ShowData 78, got %d", result.ShowData)
	}
}

func TestGetActions_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetActions(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetActionsRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 50     // total_call_clicks
				*dest[1].(*int64) = 200    // total_website_clicks
				*dest[2].(*int64) = 30     // total_direction_requests
				*dest[3].(*int64) = 10     // total_other_actions
				*dest[4].(*float64) = 9.35 // avg_actions
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetActionsRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalCallClicks != 50 {
		t.Fatalf("expected TotalCallClicks 50, got %d", result.TotalCallClicks)
	}
	if result.TotalWebsiteClicks != 200 {
		t.Fatalf("expected TotalWebsiteClicks 200, got %d", result.TotalWebsiteClicks)
	}
}

func TestGetSearchKeywords_Success(t *testing.T) {
	keywordMonth := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 2,
			scanFn: func(idx int, dest ...any) error {
				switch idx {
				case 0:
					*dest[0].(*string) = "pizza near me"
					*dest[1].(*int64) = 1500
					*dest[2].(*int64) = 0
					*dest[3].(*time.Time) = keywordMonth
				case 1:
					*dest[0].(*string) = "best restaurant"
					*dest[1].(*int64) = 800
					*dest[2].(*int64) = 1
					*dest[3].(*time.Time) = keywordMonth
				}
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetSearchKeywords(context.Background(), newTestParams(), 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 keywords, got %d", len(result))
	}
	if result[0].Keyword != "pizza near me" {
		t.Fatalf("expected 'pizza near me', got %q", result[0].Keyword)
	}
	if result[0].ImpressionsValue != 1500 {
		t.Fatalf("expected ImpressionsValue 1500, got %d", result[0].ImpressionsValue)
	}
}

func TestGetSearchKeywords_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetSearchKeywords(context.Background(), newTestParams(), 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSearchKeywords_ScanError(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 1,
			scanErr:   errors.New("scan failed"),
		},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetSearchKeywords(context.Background(), newTestParams(), 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetSearchKeywords_EmptyResult(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{nextCount: 0},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetSearchKeywords(context.Background(), newTestParams(), 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Fatalf("expected 0 keywords, got %d", len(result))
	}
}

func TestGetTopPosts_Success(t *testing.T) {
	createdAt := time.Date(2025, 1, 15, 12, 0, 0, 0, time.UTC)
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 1,
			scanFn: func(idx int, dest ...any) error {
				*dest[0].(*string) = "locations/loc_1/localPosts/post_1"
				*dest[1].(*string) = "Check out our new menu!"
				*dest[2].(*string) = "LIVE"
				*dest[3].(*string) = "STANDARD"
				*dest[4].(*string) = "https://search.google.com/local/posts"
				*dest[5].(*[]string) = []string{"media_1"}
				*dest[6].(*[]string) = []string{"PHOTO"}
				*dest[7].(*[]string) = []string{"https://lh3.google.com/media_1"}
				*dest[8].(*time.Time) = createdAt
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetTopPosts(context.Background(), newTestParams(), 15, "created_at")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 post, got %d", len(result))
	}
	if result[0].State != "LIVE" {
		t.Fatalf("expected state LIVE, got %q", result[0].State)
	}
	if result[0].TopicType != "STANDARD" {
		t.Fatalf("expected topic_type STANDARD, got %q", result[0].TopicType)
	}
}

func TestGetTopPosts_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopPosts(context.Background(), newTestParams(), 15, "created_at")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTopPosts_ScanError(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 1,
			scanErr:   errors.New("scan failed"),
		},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopPosts(context.Background(), newTestParams(), 15, "created_at")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetPublishingBehavior_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*[]int64) = []int64{1, 0, 2}
				*dest[1].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetPublishingBehavior(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.PostCount) != 3 {
		t.Fatalf("expected 3 post counts, got %d", len(result.PostCount))
	}
}

func TestGetPublishingBehavior_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetPublishingBehavior(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetTopicTypes_Success(t *testing.T) {
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 2,
			scanFn: func(idx int, dest ...any) error {
				switch idx {
				case 0:
					*dest[0].(*string) = "STANDARD"
					*dest[1].(*int64) = 8
				case 1:
					*dest[0].(*string) = "EVENT"
					*dest[1].(*int64) = 3
				}
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetTopicTypes(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 2 {
		t.Fatalf("expected 2 topic types, got %d", len(result))
	}
	if result[0].TopicType != "STANDARD" {
		t.Fatalf("expected STANDARD, got %q", result[0].TopicType)
	}
	if result[0].Count != 8 {
		t.Fatalf("expected count 8, got %d", result[0].Count)
	}
}

func TestGetTopicTypes_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetTopicTypes(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetReviewsSummary_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*float64) = 4.2 // avg_rating
				*dest[1].(*int64) = 25    // total_reviews
				*dest[2].(*int64) = 1     // star_1
				*dest[3].(*int64) = 2     // star_2
				*dest[4].(*int64) = 3     // star_3
				*dest[5].(*int64) = 8     // star_4
				*dest[6].(*int64) = 11    // star_5
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetReviewsSummary(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AvgRating != 4.2 {
		t.Fatalf("expected AvgRating 4.2, got %f", result.AvgRating)
	}
	if result.TotalReviews != 25 {
		t.Fatalf("expected TotalReviews 25, got %d", result.TotalReviews)
	}
	if result.Star5 != 11 {
		t.Fatalf("expected Star5 11, got %d", result.Star5)
	}
}

func TestGetReviewsSummary_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetReviewsSummary(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetReviewsTimeSeries_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*[]int64) = []int64{2, 0, 1}
				*dest[1].(*[]time.Time) = []time.Time{
					time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 2, 0, 0, 0, 0, time.UTC),
					time.Date(2025, 1, 3, 0, 0, 0, 0, time.UTC),
				}
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetReviewsTimeSeries(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.DailyReviews) != 3 {
		t.Fatalf("expected 3 daily reviews, got %d", len(result.DailyReviews))
	}
}

func TestGetReviewsList_Success(t *testing.T) {
	createdAt := time.Date(2025, 1, 10, 14, 30, 0, 0, time.UTC)
	conn := &mockConn{
		queryRows: &mockRows{
			nextCount: 1,
			scanFn: func(idx int, dest ...any) error {
				*dest[0].(*string) = "review_1"
				*dest[1].(*string) = "John Doe"
				*dest[2].(*string) = "https://lh3.google.com/photo"
				*dest[3].(*int64) = 5
				*dest[4].(*string) = "Great place!"
				*dest[5].(*string) = "Thank you!"
				*dest[6].(*time.Time) = createdAt
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetReviewsList(context.Background(), newTestParams(), 50)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 {
		t.Fatalf("expected 1 review, got %d", len(result))
	}
	if result[0].ReviewID != "review_1" {
		t.Fatalf("expected review_1, got %q", result[0].ReviewID)
	}
	if result[0].StarRating != 5 {
		t.Fatalf("expected star_rating 5, got %d", result[0].StarRating)
	}
	if result[0].Comment != "Great place!" {
		t.Fatalf("expected 'Great place!', got %q", result[0].Comment)
	}
}

func TestGetReviewsList_QueryError(t *testing.T) {
	conn := &mockConn{queryErr: errors.New("query failed")}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetReviewsList(context.Background(), newTestParams(), 50)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetReviewsRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 25
				*dest[1].(*float64) = 4.2
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetReviewsRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalReviews != 25 {
		t.Fatalf("expected TotalReviews 25, got %d", result.TotalReviews)
	}
	if result.AvgRating != 4.2 {
		t.Fatalf("expected AvgRating 4.2, got %f", result.AvgRating)
	}
}

func TestGetReviewsRollup_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetReviewsRollup(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetMediaActivity_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*[]int64) = []int64{3, 5, 2}   // photo_count_daily
				*dest[1].(*[]int64) = []int64{1, 0, 1}   // video_count_daily
				*dest[2].(*int64) = 12                     // show_data
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
	result, err := repo.GetMediaActivity(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result.PhotoCountDaily) != 3 {
		t.Fatalf("expected 3 photo counts, got %d", len(result.PhotoCountDaily))
	}
	if len(result.VideoCountDaily) != 3 {
		t.Fatalf("expected 3 video counts, got %d", len(result.VideoCountDaily))
	}
	if result.ShowData != 12 {
		t.Fatalf("expected ShowData 12, got %d", result.ShowData)
	}
}

func TestGetMediaActivity_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetMediaActivity(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestGetMediaActivityRollup_Success(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{
			scanFn: func(dest ...any) error {
				*dest[0].(*int64) = 45     // total_photos
				*dest[1].(*int64) = 8      // total_videos
				*dest[2].(*float64) = 1.71 // avg_media
				return nil
			},
		},
	}

	repo := NewRepository(newTestClient(conn))
	result, err := repo.GetMediaActivityRollup(context.Background(), newTestParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.TotalPhotos != 45 {
		t.Fatalf("expected TotalPhotos 45, got %d", result.TotalPhotos)
	}
	if result.TotalVideos != 8 {
		t.Fatalf("expected TotalVideos 8, got %d", result.TotalVideos)
	}
}

func TestGetMediaActivityRollup_Error(t *testing.T) {
	conn := &mockConn{
		queryRowResult: &mockRow{scanErr: errors.New("query failed")},
	}

	repo := NewRepository(newTestClient(conn))
	_, err := repo.GetMediaActivityRollup(context.Background(), newTestParams())
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
