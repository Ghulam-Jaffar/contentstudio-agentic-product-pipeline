package meta_ads

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

type mockRow struct {
	scanFn func(dest ...any) error
	err    error
}

func (m *mockRow) Err() error { return m.err }
func (m *mockRow) Scan(dest ...any) error {
	if m.scanFn != nil {
		return m.scanFn(dest...)
	}
	return m.err
}
func (m *mockRow) ScanStruct(dest any) error { return m.err }

type mockRows struct {
	scanFn    func(idx int, dest ...any) error
	nextCount int
	scanIndex int
	errVal    error
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
	if m.scanFn != nil {
		return m.scanFn(m.scanIndex, dest...)
	}
	return nil
}
func (m *mockRows) ScanStruct(dest any) error { return nil }
func (m *mockRows) Totals(dest ...any) error  { return nil }
func (m *mockRows) Close() error              { return nil }
func (m *mockRows) Err() error                { return m.errVal }

type mockConn struct {
	queryRow driver.Row
	rows     driver.Rows
}

func (m *mockConn) Contributors() []string                        { return nil }
func (m *mockConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (m *mockConn) Select(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (m *mockConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	if m.rows != nil {
		return m.rows, nil
	}
	return &mockRows{}, nil
}
func (m *mockConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	if m.queryRow != nil {
		return m.queryRow
	}
	return &mockRow{err: errors.New("no row configured")}
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

func testParams() *ch.QueryParams {
	return &ch.QueryParams{
		AccountIDs: []string{"act_123"},
		DateFrom:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		DateTo:     time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC),
		Timezone:   "UTC",
		DayCount:   31,
	}
}

func testRepo(conn *mockConn) *Repository {
	return NewRepository(&ch.Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test_db"},
		Logger: zerolog.New(io.Discard),
	})
}

func TestHelpers(t *testing.T) {
	params := testParams()
	if got := insightsDateFilter(params); got == "" {
		t.Fatal("expected non-empty insights filter")
	}
	if got := prevInsightsDateFilter(params); got == "" {
		t.Fatal("expected non-empty previous filter")
	}
	if got := safeDiv("sum(spend)", "sum(clicks)"); got != "if(sum(clicks) > 0, sum(spend) / sum(clicks), 0)" {
		t.Fatalf("unexpected safeDiv: %s", got)
	}
	if got := searchHaving("any(name)", "foo"); got != "positionCaseInsensitive(any(name), 'foo') > 0" {
		t.Fatalf("unexpected searchHaving: %s", got)
	}
	if got := resultsExprForObjective("OUTCOME_TRAFFIC"); got != "sum(clicks)" {
		t.Fatalf("unexpected results expr: %s", got)
	}
	if got := metricExpr("ctr"); got != "if(sum(impressions) > 0, sum(clicks) * 100 / sum(impressions), 0)" {
		t.Fatalf("unexpected metric expr: %s", got)
	}
}

func TestGetSummary(t *testing.T) {
	repo := testRepo(&mockConn{
		queryRow: &mockRow{scanFn: func(dest ...any) error {
			*dest[0].(*float64) = 123.5
			*dest[1].(*int64) = 42
			*dest[2].(*int64) = 1000
			*dest[3].(*int64) = 55
			return nil
		}},
	})

	result, err := repo.GetSummary(context.Background(), testParams())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Spend != 123.5 || result.Impressions != 1000 {
		t.Fatalf("unexpected summary: %+v", result)
	}
}

func TestGetTopCampaigns(t *testing.T) {
	rows := &mockRows{
		nextCount: 1,
		scanFn: func(idx int, dest ...any) error {
			*dest[0].(*string) = "camp-1"
			*dest[1].(*string) = "Campaign 1"
			*dest[2].(*float64) = 10.5
			*dest[3].(*int64) = 100
			*dest[4].(*float64) = 12.3
			return nil
		},
	}
	repo := testRepo(&mockConn{rows: rows})

	result, err := repo.GetTopCampaigns(context.Background(), testParams(), "spend")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0].CampaignID != "camp-1" {
		t.Fatalf("unexpected result: %+v", result)
	}
}
