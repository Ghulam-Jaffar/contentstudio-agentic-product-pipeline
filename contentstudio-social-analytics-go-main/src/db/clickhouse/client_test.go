package clickhouse

import (
	"context"
	"errors"
	"io"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/column"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	clickhousemodels "github.com/d4interactive/contentstudio-social-analytics-go/src/models/db/clickhouse"
)

func testLogger() zerolog.Logger {
	return zerolog.New(io.Discard)
}

type mockConn struct {
	pingErr          error
	closeErr         error
	queryErr         error
	execErr          error
	prepareBatchErr  error
	batchAppendErr   error
	batchSendErr     error
	queryRows        driver.Rows
	prepareBatchMock driver.Batch
	queryRowResult   driver.Row
}

func (m *mockConn) Contributors() []string                                  { return nil }
func (m *mockConn) ServerVersion() (*driver.ServerVersion, error)           { return nil, nil }
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
	return nil
}

type mockRow struct {
	scanErr error
	values  []any
}

func (m *mockRow) Err() error { return m.scanErr }
func (m *mockRow) Scan(dest ...any) error {
	if m.scanErr != nil {
		return m.scanErr
	}
	if m.values != nil {
		for i, v := range m.values {
			if i < len(dest) {
				switch d := dest[i].(type) {
				case *int64:
					if val, ok := v.(int64); ok {
						*d = val
					}
				case *string:
					if val, ok := v.(string); ok {
						*d = val
					}
				}
			}
		}
	}
	return nil
}
func (m *mockRow) ScanStruct(dest any) error { return m.scanErr }
func (m *mockConn) PrepareBatch(ctx context.Context, query string, opts ...driver.PrepareBatchOption) (driver.Batch, error) {
	if m.prepareBatchErr != nil {
		return nil, m.prepareBatchErr
	}
	if m.prepareBatchMock != nil {
		return m.prepareBatchMock, nil
	}
	return &mockBatch{appendErr: m.batchAppendErr, sendErr: m.batchSendErr}, nil
}
func (m *mockConn) Exec(ctx context.Context, query string, args ...any) error {
	return m.execErr
}
func (m *mockConn) AsyncInsert(ctx context.Context, query string, wait bool, args ...any) error {
	return nil
}
func (m *mockConn) Ping(ctx context.Context) error {
	return m.pingErr
}
func (m *mockConn) Stats() driver.Stats { return driver.Stats{} }
func (m *mockConn) Close() error        { return m.closeErr }

type mockBatch struct {
	appendErr   error
	sendErr     error
	appendCount int
}

func (m *mockBatch) Abort() error                               { return nil }
func (m *mockBatch) Append(v ...any) error                      { m.appendCount++; return m.appendErr }
func (m *mockBatch) AppendStruct(v any) error                   { return m.appendErr }
func (m *mockBatch) Column(int) driver.BatchColumn              { return nil }
func (m *mockBatch) Columns() []column.Interface                { return nil }
func (m *mockBatch) Flush() error                               { return nil }
func (m *mockBatch) Send() error                                { return m.sendErr }
func (m *mockBatch) IsSent() bool                               { return false }
func (m *mockBatch) Rows() int                                  { return m.appendCount }
func (m *mockBatch) Close() error                               { return nil }

type mockRows struct {
	scanErr    error
	nextCount  int
	closeErr   error
	columns    []string
	errVal     error
	scanValues [][]any
	scanIndex  int
}

func (m *mockRows) Columns() []string                { return m.columns }
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
	if m.scanValues != nil && m.scanIndex < len(m.scanValues) {
		vals := m.scanValues[m.scanIndex]
		m.scanIndex++
		for i, v := range vals {
			if i < len(dest) {
				switch d := dest[i].(type) {
				case *string:
					if s, ok := v.(string); ok {
						*d = s
					}
				case *uint64:
					if u, ok := v.(uint64); ok {
						*d = u
					}
				case *int64:
					if i, ok := v.(int64); ok {
						*d = i
					}
				}
			}
		}
	}
	return nil
}
func (m *mockRows) ScanStruct(dest any) error { return m.scanErr }
func (m *mockRows) Totals(dest ...any) error  { return nil }
func (m *mockRows) Close() error              { return m.closeErr }
func (m *mockRows) Err() error                { return m.errVal }

func newTestClient(conn *mockConn) *Client {
	return &Client{
		Conn:   conn,
		Config: config.ClickHouseConfig{Database: "test_db"},
		Logger: testLogger(),
	}
}

func Test_NewClient_Table(t *testing.T) {
	originalOpenConn := openConn
	defer func() { openConn = originalOpenConn }()

	cases := []struct {
		name      string
		cfg       config.ClickHouseConfig
		mockConn  *mockConn
		mockErr   error
		pingErr   error
		expectErr bool
	}{
		{
			name: "success",
			cfg: config.ClickHouseConfig{
				Host:     "localhost",
				Port:     9000,
				Database: "test",
				Username: "default",
				Password: "",
			},
			mockConn:  &mockConn{},
			expectErr: false,
		},
		{
			name: "success with secure and compression disabled",
			cfg: config.ClickHouseConfig{
				Host:        "localhost",
				Port:        9000,
				Database:    "test",
				Username:    "default",
				Secure:      true,
				Compression: false,
			},
			mockConn:  &mockConn{},
			expectErr: false,
		},
		{
			name: "success with max execution time",
			cfg: config.ClickHouseConfig{
				Host:                  "localhost",
				Port:                  9000,
				Database:              "test",
				MaxExecutionTimeInSec: 120,
			},
			mockConn:  &mockConn{},
			expectErr: false,
		},
		{
			name: "open connection error",
			cfg: config.ClickHouseConfig{
				Host:     "localhost",
				Port:     9000,
				Database: "test",
			},
			mockErr:   errors.New("connection refused"),
			expectErr: true,
		},
		{
			name: "ping error",
			cfg: config.ClickHouseConfig{
				Host:     "localhost",
				Port:     9000,
				Database: "test",
			},
			mockConn:  &mockConn{pingErr: errors.New("ping failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			openConn = func(opt *clickhouse.Options) (clickhouse.Conn, error) {
				if tc.mockErr != nil {
					return nil, tc.mockErr
				}
				return tc.mockConn, nil
			}

			client, err := NewClient(tc.cfg, testLogger())
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if client == nil {
					t.Fatal("expected client, got nil")
				}
			}
		})
	}
}

func Test_Client_Health_Table(t *testing.T) {
	cases := []struct {
		name      string
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "healthy connection",
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name:      "ping fails",
			conn:      &mockConn{pingErr: errors.New("connection refused")},
			expectErr: true,
		},
		{
			name:      "nil connection",
			conn:      nil,
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var client *Client
			if tc.conn == nil {
				client = &Client{Conn: nil, Logger: testLogger()}
			} else {
				client = newTestClient(tc.conn)
			}

			err := client.Health()
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_Client_Close_Table(t *testing.T) {
	cases := []struct {
		name      string
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "close success",
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name:      "close error",
			conn:      &mockConn{closeErr: errors.New("close failed")},
			expectErr: true,
		},
		{
			name:      "nil connection",
			conn:      nil,
			expectErr: false,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			var client *Client
			if tc.conn == nil {
				client = &Client{Conn: nil, Logger: testLogger()}
			} else {
				client = newTestClient(tc.conn)
			}

			err := client.Close()
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertPosts_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		posts     []*clickhousemodels.FacebookPosts
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty posts",
			posts:     []*clickhousemodels.FacebookPosts{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single post success",
			posts: []*clickhousemodels.FacebookPosts{
				{
					PageID:      "page_1",
					PostID:      "post_1",
					PageName:    "Test Page",
					CreatedTime: now,
					UpdatedTime: now,
					SavingTime:  now,
				},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "multiple posts success",
			posts: []*clickhousemodels.FacebookPosts{
				{PageID: "page_1", PostID: "post_1", PageName: "Test Page 1", CreatedTime: now, UpdatedTime: now, SavingTime: now},
				{PageID: "page_2", PostID: "post_2", PageName: "Test Page 2", CreatedTime: now, UpdatedTime: now, SavingTime: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			posts: []*clickhousemodels.FacebookPosts{
				{PageID: "page_1", PostID: "post_1", CreatedTime: now, UpdatedTime: now, SavingTime: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			posts: []*clickhousemodels.FacebookPosts{
				{PageID: "page_1", PostID: "post_1", CreatedTime: now, UpdatedTime: now, SavingTime: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			posts: []*clickhousemodels.FacebookPosts{
				{PageID: "page_1", PostID: "post_1", CreatedTime: now, UpdatedTime: now, SavingTime: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertPosts(context.Background(), tc.posts)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertMediaAssets_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		assets    []*clickhousemodels.FacebookMediaAssets
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty assets",
			assets:    []*clickhousemodels.FacebookMediaAssets{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single asset success",
			assets: []*clickhousemodels.FacebookMediaAssets{
				{PageID: "page_1", MediaID: "media_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			assets: []*clickhousemodels.FacebookMediaAssets{
				{PageID: "page_1", MediaID: "media_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			assets: []*clickhousemodels.FacebookMediaAssets{
				{PageID: "page_1", MediaID: "media_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			assets: []*clickhousemodels.FacebookMediaAssets{
				{PageID: "page_1", MediaID: "media_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertMediaAssets(context.Background(), tc.assets)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertVideoInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.FacebookVideoInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.FacebookVideoInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.FacebookVideoInsights{
				{PageID: "page_1", PostID: "post_1", VideoID: "video_1", CreatedTime: now, UpdatedTime: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.FacebookVideoInsights{
				{PageID: "page_1", PostID: "post_1", VideoID: "video_1", CreatedTime: now, UpdatedTime: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.FacebookVideoInsights{
				{PageID: "page_1", PostID: "post_1", VideoID: "video_1", CreatedTime: now, UpdatedTime: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.FacebookVideoInsights{
				{PageID: "page_1", PostID: "post_1", VideoID: "video_1", CreatedTime: now, UpdatedTime: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertVideoInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertReelsInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.FacebookReelsInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.FacebookReelsInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.FacebookReelsInsights{
				{PageID: "page_1", PostID: "post_1", CreatedAt: now, SavingTime: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.FacebookReelsInsights{
				{PageID: "page_1", PostID: "post_1", CreatedAt: now, SavingTime: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.FacebookReelsInsights{
				{PageID: "page_1", PostID: "post_1", CreatedAt: now, SavingTime: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.FacebookReelsInsights{
				{PageID: "page_1", PostID: "post_1", CreatedAt: now, SavingTime: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertReelsInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.FacebookInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.FacebookInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.FacebookInsights{
				{HashID: "hash_1", PageID: "page_1", CreatedTime: now, SavingTime: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.FacebookInsights{
				{HashID: "hash_1", PageID: "page_1", CreatedTime: now, SavingTime: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.FacebookInsights{
				{HashID: "hash_1", PageID: "page_1", CreatedTime: now, SavingTime: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.FacebookInsights{
				{HashID: "hash_1", PageID: "page_1", CreatedTime: now, SavingTime: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertInstagramPosts_Table(t *testing.T) {
	now := time.Now()
	timestamp := now.Unix()
	cases := []struct {
		name      string
		posts     []*clickhousemodels.InstagramPost
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty posts",
			posts:     []*clickhousemodels.InstagramPost{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single post success",
			posts: []*clickhousemodels.InstagramPost{
				{InstagramID: "ig_1", MediaID: "media_1", Timestamp: timestamp, StoredEventAt: now, PostCreatedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			posts: []*clickhousemodels.InstagramPost{
				{InstagramID: "ig_1", MediaID: "media_1", Timestamp: timestamp, StoredEventAt: now, PostCreatedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			posts: []*clickhousemodels.InstagramPost{
				{InstagramID: "ig_1", MediaID: "media_1", Timestamp: timestamp, StoredEventAt: now, PostCreatedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			posts: []*clickhousemodels.InstagramPost{
				{InstagramID: "ig_1", MediaID: "media_1", Timestamp: timestamp, StoredEventAt: now, PostCreatedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertInstagramPosts(context.Background(), tc.posts)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertInstagramInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.InstagramInsight
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.InstagramInsight{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.InstagramInsight{
				{InstagramID: "ig_1", RecordID: "record_1", CreatedTime: now, UpdatedTime: now, StoredEventAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.InstagramInsight{
				{InstagramID: "ig_1", RecordID: "record_1", CreatedTime: now, UpdatedTime: now, StoredEventAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.InstagramInsight{
				{InstagramID: "ig_1", RecordID: "record_1", CreatedTime: now, UpdatedTime: now, StoredEventAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.InstagramInsight{
				{InstagramID: "ig_1", RecordID: "record_1", CreatedTime: now, UpdatedTime: now, StoredEventAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertInstagramInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertLinkedInPosts_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		posts     []*clickhousemodels.LinkedInPosts
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty posts",
			posts:     []*clickhousemodels.LinkedInPosts{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single post success",
			posts: []*clickhousemodels.LinkedInPosts{
				{LinkedinID: "ln_1", PostID: "post_1", CreatedAt: now, PublishedAt: now, LastModifiedAt: now, SavingTime: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			posts: []*clickhousemodels.LinkedInPosts{
				{LinkedinID: "ln_1", PostID: "post_1", CreatedAt: now, PublishedAt: now, LastModifiedAt: now, SavingTime: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			posts: []*clickhousemodels.LinkedInPosts{
				{LinkedinID: "ln_1", PostID: "post_1", CreatedAt: now, PublishedAt: now, LastModifiedAt: now, SavingTime: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			posts: []*clickhousemodels.LinkedInPosts{
				{LinkedinID: "ln_1", PostID: "post_1", CreatedAt: now, PublishedAt: now, LastModifiedAt: now, SavingTime: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertLinkedInPosts(context.Background(), tc.posts)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertLinkedInInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.LinkedInInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.LinkedInInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.LinkedInInsights{
				{LinkedinID: "ln_1", RecordID: "record_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.LinkedInInsights{
				{LinkedinID: "ln_1", RecordID: "record_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.LinkedInInsights{
				{LinkedinID: "ln_1", RecordID: "record_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.LinkedInInsights{
				{LinkedinID: "ln_1", RecordID: "record_1", InsertedAt: now, CreatedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertLinkedInInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertTikTokPosts_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		posts     []*clickhousemodels.TikTokPosts
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty posts",
			posts:     []*clickhousemodels.TikTokPosts{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single post success",
			posts: []*clickhousemodels.TikTokPosts{
				{TikTokID: "tt_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			posts: []*clickhousemodels.TikTokPosts{
				{TikTokID: "tt_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			posts: []*clickhousemodels.TikTokPosts{
				{TikTokID: "tt_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			posts: []*clickhousemodels.TikTokPosts{
				{TikTokID: "tt_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertTikTokPosts(context.Background(), tc.posts)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_BulkInsertTikTokInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.TikTokInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.TikTokInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.TikTokInsights{
				{TikTokID: "tt_1", RecordID: "record_1", InsertedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.TikTokInsights{
				{TikTokID: "tt_1", RecordID: "record_1", InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.TikTokInsights{
				{TikTokID: "tt_1", RecordID: "record_1", InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.TikTokInsights{
				{TikTokID: "tt_1", RecordID: "record_1", InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.BulkInsertTikTokInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_GetMinimalOlderThan20DaysByPage_Table(t *testing.T) {
	cases := []struct {
		name      string
		tableName string
		pageID    string
		conn      *mockConn
		expectErr bool
		wantCount int
	}{
		{
			name:      "empty table name uses default",
			tableName: "",
			pageID:    "page_1",
			conn:      &mockConn{queryRows: &mockRows{nextCount: 0}},
			expectErr: false,
			wantCount: 0,
		},
		{
			name:      "empty pageID returns error",
			tableName: "facebook_posts",
			pageID:    "",
			conn:      &mockConn{},
			expectErr: true,
		},
		{
			name:      "query error",
			tableName: "facebook_posts",
			pageID:    "page_1",
			conn:      &mockConn{queryErr: errors.New("query failed")},
			expectErr: true,
		},
		{
			name:      "success with empty results",
			tableName: "facebook_posts",
			pageID:    "page_1",
			conn:      &mockConn{queryRows: &mockRows{nextCount: 0}},
			expectErr: false,
			wantCount: 0,
		},
		{
			name:      "scan struct error",
			tableName: "facebook_posts",
			pageID:    "page_1",
			conn:      &mockConn{queryRows: &mockRows{nextCount: 1, scanErr: errors.New("scan failed")}},
			expectErr: true,
		},
		{
			name:      "rows err error",
			tableName: "facebook_posts",
			pageID:    "page_1",
			conn:      &mockConn{queryRows: &mockRows{nextCount: 0, errVal: errors.New("rows error")}},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			results, err := client.GetMinimalOlderThan20DaysByPage(context.Background(), tc.tableName, tc.pageID, 500, 0)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(results) != tc.wantCount {
					t.Fatalf("expected %d results, got %d", tc.wantCount, len(results))
				}
			}
		})
	}
}

func Test_UpdateFullPictures_Table(t *testing.T) {
	cases := []struct {
		name        string
		tableName   string
		pageID      string
		rows        []clickhousemodels.MinimalPost
		conn        *mockConn
		expectErr   bool
		expectCount int
	}{
		{
			name:        "empty table name uses default",
			tableName:   "",
			pageID:      "page_1",
			rows:        []clickhousemodels.MinimalPost{{PageID: "page_1", PostID: "post_1", FullPicture: "http://example.com/pic.jpg"}},
			conn:        &mockConn{},
			expectErr:   false,
			expectCount: 1,
		},
		{
			name:      "empty pageID returns error",
			tableName: "facebook_posts",
			pageID:    "",
			rows:      []clickhousemodels.MinimalPost{},
			conn:      &mockConn{},
			expectErr: true,
		},
		{
			name:        "empty rows returns 0",
			tableName:   "facebook_posts",
			pageID:      "page_1",
			rows:        []clickhousemodels.MinimalPost{},
			conn:        &mockConn{},
			expectErr:   false,
			expectCount: 0,
		},
		{
			name:      "rows with different pageID are filtered",
			tableName: "facebook_posts",
			pageID:    "page_1",
			rows: []clickhousemodels.MinimalPost{
				{PageID: "page_2", PostID: "post_1", FullPicture: "http://example.com/pic.jpg"},
			},
			conn:        &mockConn{},
			expectErr:   false,
			expectCount: 0,
		},
		{
			name:      "rows with empty post_id are filtered",
			tableName: "facebook_posts",
			pageID:    "page_1",
			rows: []clickhousemodels.MinimalPost{
				{PageID: "page_1", PostID: "", FullPicture: "http://example.com/pic.jpg"},
			},
			conn:        &mockConn{},
			expectErr:   false,
			expectCount: 0,
		},
		{
			name:      "rows with empty full_picture are filtered",
			tableName: "facebook_posts",
			pageID:    "page_1",
			rows: []clickhousemodels.MinimalPost{
				{PageID: "page_1", PostID: "post_1", FullPicture: ""},
			},
			conn:        &mockConn{},
			expectErr:   false,
			expectCount: 0,
		},
		{
			name:      "exec error",
			tableName: "facebook_posts",
			pageID:    "page_1",
			rows: []clickhousemodels.MinimalPost{
				{PageID: "page_1", PostID: "post_1", FullPicture: "http://example.com/pic.jpg"},
			},
			conn:      &mockConn{execErr: errors.New("exec failed")},
			expectErr: true,
		},
		{
			name:      "duplicate post_ids are deduplicated",
			tableName: "facebook_posts",
			pageID:    "page_1",
			rows: []clickhousemodels.MinimalPost{
				{PageID: "page_1", PostID: "post_1", FullPicture: "http://example.com/pic1.jpg"},
				{PageID: "page_1", PostID: "post_1", FullPicture: "http://example.com/pic2.jpg"},
			},
			conn:        &mockConn{},
			expectErr:   false,
			expectCount: 1,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			count, err := client.UpdateFullPictures(context.Background(), tc.tableName, tc.pageID, tc.rows)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if count != tc.expectCount {
					t.Fatalf("expected count %d, got %d", tc.expectCount, count)
				}
			}
		})
	}
}

func Test_GetGeoMappings_Table(t *testing.T) {
	cases := []struct {
		name      string
		geoIDs    []string
		conn      *mockConn
		expectErr bool
		wantCount int
	}{
		{
			name:      "empty geoIDs",
			geoIDs:    []string{},
			conn:      &mockConn{},
			expectErr: false,
			wantCount: 0,
		},
		{
			name:      "query error",
			geoIDs:    []string{"geo_1"},
			conn:      &mockConn{queryErr: errors.New("query failed")},
			expectErr: true,
		},
		{
			name:      "success with empty results",
			geoIDs:    []string{"geo_1"},
			conn:      &mockConn{queryRows: &mockRows{nextCount: 0}},
			expectErr: false,
			wantCount: 0,
		},
		{
			name:      "scan error",
			geoIDs:    []string{"geo_1"},
			conn:      &mockConn{queryRows: &mockRows{nextCount: 1, scanErr: errors.New("scan failed")}},
			expectErr: true,
		},
		{
			name:      "rows err error",
			geoIDs:    []string{"geo_1"},
			conn:      &mockConn{queryRows: &mockRows{nextCount: 0, errVal: errors.New("rows error")}},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			results, err := client.GetGeoMappings(context.Background(), tc.geoIDs)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				if len(results) != tc.wantCount {
					t.Fatalf("expected %d results, got %d", tc.wantCount, len(results))
				}
			}
		})
	}
}

func Test_InsertGeoMappings_Table(t *testing.T) {
	cases := []struct {
		name      string
		mappings  map[string]string
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty mappings",
			mappings:  map[string]string{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name:      "single mapping success",
			mappings:  map[string]string{"geo_1": "Location 1"},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name:      "prepare batch error",
			mappings:  map[string]string{"geo_1": "Location 1"},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name:      "append error",
			mappings:  map[string]string{"geo_1": "Location 1"},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name:      "send error",
			mappings:  map[string]string{"geo_1": "Location 1"},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.InsertGeoMappings(context.Background(), tc.mappings)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_InsertGeoMappingsWithType_Table(t *testing.T) {
	cases := []struct {
		name      string
		mappings  []GeoMappingWithType
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty mappings",
			mappings:  []GeoMappingWithType{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name:      "single mapping success",
			mappings:  []GeoMappingWithType{{ID: "geo_1", Name: "Location 1", Type: "city"}},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name:      "prepare batch error",
			mappings:  []GeoMappingWithType{{ID: "geo_1", Name: "Location 1", Type: "city"}},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name:      "append error",
			mappings:  []GeoMappingWithType{{ID: "geo_1", Name: "Location 1", Type: "city"}},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name:      "send error",
			mappings:  []GeoMappingWithType{{ID: "geo_1", Name: "Location 1", Type: "city"}},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.InsertGeoMappingsWithType(context.Background(), tc.mappings)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_InsertCompetitorInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.FacebookCompetitorInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.FacebookCompetitorInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.FacebookCompetitorInsights{
				{RecordID: "record_1", PageID: "page_1", InsertedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.FacebookCompetitorInsights{
				{RecordID: "record_1", PageID: "page_1", InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.FacebookCompetitorInsights{
				{RecordID: "record_1", PageID: "page_1", InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.FacebookCompetitorInsights{
				{RecordID: "record_1", PageID: "page_1", InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.InsertCompetitorInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_InsertCompetitorPosts_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		posts     []*clickhousemodels.FacebookCompetitorPosts
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty posts",
			posts:     []*clickhousemodels.FacebookCompetitorPosts{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single post success",
			posts: []*clickhousemodels.FacebookCompetitorPosts{
				{FacebookID: "fb_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			posts: []*clickhousemodels.FacebookCompetitorPosts{
				{FacebookID: "fb_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			posts: []*clickhousemodels.FacebookCompetitorPosts{
				{FacebookID: "fb_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			posts: []*clickhousemodels.FacebookCompetitorPosts{
				{FacebookID: "fb_1", PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.InsertCompetitorPosts(context.Background(), tc.posts)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_InsertCompetitorMediaAssets_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		assets    []*clickhousemodels.FacebookCompetitorMediaAssets
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty assets",
			assets:    []*clickhousemodels.FacebookCompetitorMediaAssets{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single asset success",
			assets: []*clickhousemodels.FacebookCompetitorMediaAssets{
				{MediaID: "media_1", PostID: "post_1", PageID: "page_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			assets: []*clickhousemodels.FacebookCompetitorMediaAssets{
				{MediaID: "media_1", PostID: "post_1", PageID: "page_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			assets: []*clickhousemodels.FacebookCompetitorMediaAssets{
				{MediaID: "media_1", PostID: "post_1", PageID: "page_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			assets: []*clickhousemodels.FacebookCompetitorMediaAssets{
				{MediaID: "media_1", PostID: "post_1", PageID: "page_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.InsertCompetitorMediaAssets(context.Background(), tc.assets)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_InsertInstagramCompetitorInsights_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		insights  []*clickhousemodels.InstagramCompetitorInsights
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty insights",
			insights:  []*clickhousemodels.InstagramCompetitorInsights{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single insight success",
			insights: []*clickhousemodels.InstagramCompetitorInsights{
				{RecordID: "record_1", InstagramAccountID: "ig_1", InsertedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			insights: []*clickhousemodels.InstagramCompetitorInsights{
				{RecordID: "record_1", InstagramAccountID: "ig_1", InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			insights: []*clickhousemodels.InstagramCompetitorInsights{
				{RecordID: "record_1", InstagramAccountID: "ig_1", InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			insights: []*clickhousemodels.InstagramCompetitorInsights{
				{RecordID: "record_1", InstagramAccountID: "ig_1", InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.InsertInstagramCompetitorInsights(context.Background(), tc.insights)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_InsertInstagramCompetitorPosts_Table(t *testing.T) {
	now := time.Now()
	cases := []struct {
		name      string
		posts     []*clickhousemodels.InstagramCompetitorPosts
		conn      *mockConn
		expectErr bool
	}{
		{
			name:      "empty posts",
			posts:     []*clickhousemodels.InstagramCompetitorPosts{},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "single post success",
			posts: []*clickhousemodels.InstagramCompetitorPosts{
				{InstagramID: 12345, PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{},
			expectErr: false,
		},
		{
			name: "prepare batch error",
			posts: []*clickhousemodels.InstagramCompetitorPosts{
				{InstagramID: 12345, PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{prepareBatchErr: errors.New("prepare failed")},
			expectErr: true,
		},
		{
			name: "append error",
			posts: []*clickhousemodels.InstagramCompetitorPosts{
				{InstagramID: 12345, PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchAppendErr: errors.New("append failed")},
			expectErr: true,
		},
		{
			name: "send error",
			posts: []*clickhousemodels.InstagramCompetitorPosts{
				{InstagramID: 12345, PostID: "post_1", CreatedAt: now, InsertedAt: now},
			},
			conn:      &mockConn{batchSendErr: errors.New("send failed")},
			expectErr: true,
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			err := client.InsertInstagramCompetitorPosts(context.Background(), tc.posts)
			if tc.expectErr {
				if err == nil {
					t.Fatalf("expected error, got nil")
				}
			} else if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func Test_LogAsyncInsertSettings_Table(t *testing.T) {
	cases := []struct {
		name string
		conn *mockConn
	}{
		{
			name: "query success no rows",
			conn: &mockConn{queryRows: &mockRows{nextCount: 0}},
		},
		{
			name: "query success with rows",
			conn: &mockConn{queryRows: &mockRows{nextCount: 2}},
		},
		{
			name: "query success with scan error",
			conn: &mockConn{queryRows: &mockRows{nextCount: 1, scanErr: errors.New("scan error")}},
		},
		{
			name: "query error",
			conn: &mockConn{queryErr: errors.New("query failed")},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			client.LogAsyncInsertSettings(context.Background())
		})
	}
}

func Test_LogPartsStats_Table(t *testing.T) {
	cases := []struct {
		name        string
		tablePrefix string
		conn        *mockConn
	}{
		{
			name:        "with prefix no rows",
			tablePrefix: "facebook_",
			conn:        &mockConn{queryRows: &mockRows{nextCount: 0}},
		},
		{
			name:        "without prefix no rows",
			tablePrefix: "",
			conn:        &mockConn{queryRows: &mockRows{nextCount: 0}},
		},
		{
			name:        "with prefix and rows OK status",
			tablePrefix: "facebook_",
			conn: &mockConn{queryRows: &mockRows{
				nextCount: 1,
				scanValues: [][]any{
					{"test_table", uint64(100), uint64(1000)},
				},
			}},
		},
		{
			name:        "with rows WARNING status (parts > 10000)",
			tablePrefix: "facebook_",
			conn: &mockConn{queryRows: &mockRows{
				nextCount: 1,
				scanValues: [][]any{
					{"test_table", uint64(15000), uint64(1000000)},
				},
			}},
		},
		{
			name:        "with rows CRITICAL status (parts > 50000)",
			tablePrefix: "facebook_",
			conn: &mockConn{queryRows: &mockRows{
				nextCount: 1,
				scanValues: [][]any{
					{"test_table", uint64(60000), uint64(5000000)},
				},
			}},
		},
		{
			name:        "with scan error",
			tablePrefix: "facebook_",
			conn:        &mockConn{queryRows: &mockRows{nextCount: 1, scanErr: errors.New("scan error")}},
		},
		{
			name:        "query error",
			tablePrefix: "facebook_",
			conn:        &mockConn{queryErr: errors.New("query failed")},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			client.LogPartsStats(context.Background(), tc.tablePrefix)
		})
	}
}

func Test_LogAsyncInsertQueue_Table(t *testing.T) {
	cases := []struct {
		name string
		conn *mockConn
	}{
		{
			name: "query success no rows",
			conn: &mockConn{queryRows: &mockRows{nextCount: 0}},
		},
		{
			name: "query success with rows",
			conn: &mockConn{queryRows: &mockRows{nextCount: 2}},
		},
		{
			name: "query success with scan error",
			conn: &mockConn{queryRows: &mockRows{nextCount: 1, scanErr: errors.New("scan error")}},
		},
		{
			name: "query error",
			conn: &mockConn{queryErr: errors.New("query failed")},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			client := newTestClient(tc.conn)
			client.LogAsyncInsertQueue(context.Background())
		})
	}
}

func Test_StartMonitoring_Table(t *testing.T) {
	conn := &mockConn{queryRows: &mockRows{nextCount: 0}}
	client := newTestClient(conn)

	ctx, cancel := context.WithCancel(context.Background())
	client.StartMonitoring(ctx, 100*time.Millisecond, "facebook_")

	time.Sleep(150 * time.Millisecond)
	cancel()

	time.Sleep(50 * time.Millisecond)
}

func Test_BulkInsert_MultipleItems_Table(t *testing.T) {
	now := time.Now()

	t.Run("multiple Facebook posts", func(t *testing.T) {
		posts := make([]*clickhousemodels.FacebookPosts, 10)
		for i := 0; i < 10; i++ {
			posts[i] = &clickhousemodels.FacebookPosts{
				PageID:      "page_1",
				PostID:      "post_" + string(rune('a'+i)),
				CreatedTime: now,
				UpdatedTime: now,
				SavingTime:  now,
			}
		}
		client := newTestClient(&mockConn{})
		err := client.BulkInsertPosts(context.Background(), posts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("multiple Instagram posts", func(t *testing.T) {
		posts := make([]*clickhousemodels.InstagramPost, 10)
		timestamp := now.Unix()
		for i := 0; i < 10; i++ {
			posts[i] = &clickhousemodels.InstagramPost{
				InstagramID:   "ig_1",
				MediaID:       "media_" + string(rune('a'+i)),
				Timestamp:     timestamp,
				StoredEventAt: now,
				PostCreatedAt: now,
			}
		}
		client := newTestClient(&mockConn{})
		err := client.BulkInsertInstagramPosts(context.Background(), posts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("multiple LinkedIn posts", func(t *testing.T) {
		posts := make([]*clickhousemodels.LinkedInPosts, 10)
		for i := 0; i < 10; i++ {
			posts[i] = &clickhousemodels.LinkedInPosts{
				LinkedinID:     "ln_1",
				PostID:         "post_" + string(rune('a'+i)),
				CreatedAt:      now,
				PublishedAt:    now,
				LastModifiedAt: now,
				SavingTime:     now,
			}
		}
		client := newTestClient(&mockConn{})
		err := client.BulkInsertLinkedInPosts(context.Background(), posts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("multiple TikTok posts", func(t *testing.T) {
		posts := make([]*clickhousemodels.TikTokPosts, 10)
		for i := 0; i < 10; i++ {
			posts[i] = &clickhousemodels.TikTokPosts{
				TikTokID:   "tt_1",
				PostID:     "post_" + string(rune('a'+i)),
				CreatedAt:  now,
				InsertedAt: now,
			}
		}
		client := newTestClient(&mockConn{})
		err := client.BulkInsertTikTokPosts(context.Background(), posts)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
