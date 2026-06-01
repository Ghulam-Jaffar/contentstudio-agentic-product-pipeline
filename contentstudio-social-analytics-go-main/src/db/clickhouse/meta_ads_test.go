package clickhouse

import (
	"context"
	"io"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

type metaAdsBatch struct {
	appends [][]any
}

func (b *metaAdsBatch) Abort() error { return nil }
func (b *metaAdsBatch) Append(v ...any) error {
	b.appends = append(b.appends, append([]any(nil), v...))
	return nil
}
func (b *metaAdsBatch) AppendStruct(v any) error      { return nil }
func (b *metaAdsBatch) Column(int) driver.BatchColumn { return nil }
func (b *metaAdsBatch) Columns() []column.Interface   { return nil }
func (b *metaAdsBatch) Flush() error                  { return nil }
func (b *metaAdsBatch) Send() error                   { return nil }
func (b *metaAdsBatch) IsSent() bool                  { return false }
func (b *metaAdsBatch) Rows() int                     { return len(b.appends) }
func (b *metaAdsBatch) Close() error                  { return nil }

type metaAdsConn struct {
	batch *metaAdsBatch
}

func (c *metaAdsConn) Contributors() []string                        { return nil }
func (c *metaAdsConn) ServerVersion() (*driver.ServerVersion, error) { return nil, nil }
func (c *metaAdsConn) Select(ctx context.Context, dest any, query string, args ...any) error {
	return nil
}
func (c *metaAdsConn) Query(ctx context.Context, query string, args ...any) (driver.Rows, error) {
	return nil, nil
}
func (c *metaAdsConn) QueryRow(ctx context.Context, query string, args ...any) driver.Row {
	return nil
}
func (c *metaAdsConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	c.batch = &metaAdsBatch{}
	return c.batch, nil
}
func (c *metaAdsConn) Exec(ctx context.Context, query string, args ...any) error { return nil }
func (c *metaAdsConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return nil
}
func (c *metaAdsConn) Ping(ctx context.Context) error { return nil }
func (c *metaAdsConn) Stats() driver.Stats            { return driver.Stats{} }
func (c *metaAdsConn) Close() error                   { return nil }

func TestTruncateToHour(t *testing.T) {
	got := truncateToHour(time.Date(2025, 5, 13, 17, 42, 11, 123, time.FixedZone("PKT", 5*60*60)))
	if got.Location() != time.UTC {
		t.Fatalf("expected UTC, got %v", got.Location())
	}
	if got.Minute() != 0 || got.Second() != 0 || got.Nanosecond() != 0 {
		t.Fatalf("expected truncated hour, got %v", got)
	}
}

func TestBulkInsertMetaAdsCampaigns(t *testing.T) {
	conn := &metaAdsConn{}
	client := &Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test_db"},
		Logger: zerolog.New(io.Discard),
	}
	now := time.Date(2025, 5, 13, 17, 42, 11, 123, time.UTC)

	err := client.BulkInsertMetaAdsCampaigns(context.Background(), []*clickhousemodels.MetaAdsCampaign{
		{
			AccountID:       "act_1",
			CampaignID:      "camp-1",
			Name:            "Campaign 1",
			Status:          "ACTIVE",
			EffectiveStatus: "ACTIVE",
			Objective:       "OUTCOME_TRAFFIC",
			DailyBudget:     "10",
			LifetimeBudget:  "20",
			BudgetRemaining: "5",
			StartTime:       now,
			StopTime:        now,
			CreatedTime:     now,
			UpdatedTime:     now,
		},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if conn.batch == nil || len(conn.batch.appends) != 1 {
		t.Fatalf("expected 1 appended row, got %+v", conn.batch)
	}
	if got := conn.batch.appends[0][9].(time.Time); got.Minute() != 0 || got.Second() != 0 {
		t.Fatalf("expected truncated start time, got %v", got)
	}
}
