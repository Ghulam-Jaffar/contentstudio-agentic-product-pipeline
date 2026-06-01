package clickhouse

import (
	"context"
	"crypto/tls"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/rs/zerolog"

	"github.com/d4interactive/contentstudio-social-analytics-go/src/config"
	"github.com/d4interactive/contentstudio-social-analytics-go/src/logger"
)

// connOpener is a function type for opening ClickHouse connections (allows mocking in tests)
type connOpener func(opt *clickhouse.Options) (clickhouse.Conn, error)

// openConn is the default connection opener, can be overridden in tests
var openConn connOpener = clickhouse.Open

// Client represents a ClickHouse client
type Client struct {
	Conn   clickhouse.Conn
	Config config.ClickHouseConfig
	Logger zerolog.Logger
}

// NewClient creates a new ClickHouse client with connection
func NewClient(cfg config.ClickHouseConfig, logger zerolog.Logger) (*Client, error) {
	client := &Client{
		Config: cfg,
		Logger: logger.With().Str("component", "schema").Logger(),
	}

	if err := client.connect(); err != nil {
		return nil, fmt.Errorf("NewClient: failed to connect to ClickHouse: %w", err)
	}

	return client, nil
}

// connect establishes connection to ClickHouse
func (c *Client) connect() error {
	maxExecTime := c.Config.MaxExecutionTimeInSec
	if maxExecTime <= 0 {
		maxExecTime = 60
	}

	options := &clickhouse.Options{
		Protocol: clickhouse.Native, // Use native protocol for port 9000
		Addr:     []string{fmt.Sprintf("%s:%d", c.Config.Host, c.Config.Port)},
		Auth: clickhouse.Auth{
			Database: c.Config.Database,
			Username: c.Config.Username,
			Password: c.Config.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 180, // 180 seconds timeout for queries
			"insert_deduplicate": 0,   // Disable insert deduplication - ReplacingMergeTree handles dedup at query time with FINAL
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		MaxOpenConns:    c.Config.MaxOpenConns,
		MaxIdleConns:    c.Config.MaxIdleConns,
		ConnMaxLifetime: time.Hour,
	}

	// Only enable TLS if secure is configured
	if c.Config.Secure {
		options.TLS = &tls.Config{InsecureSkipVerify: true}
	}

	// Disable compression if configured
	if !c.Config.Compression {
		options.Compression = &clickhouse.Compression{
			Method: clickhouse.CompressionNone,
		}
	}

	conn, err := openConn(options)
	if err != nil {
		return fmt.Errorf("Client.connect: failed to open ClickHouse connection: %w", err)
	}

	c.Conn = conn

	// Test the connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Conn.Ping(ctx); err != nil {
		c.Conn.Close()
		return fmt.Errorf("Client.connect: failed to ping ClickHouse: %w", err)
	}

	c.Logger.Info().
		Str("host", c.Config.Host).
		Int("port", c.Config.Port).
		Str("database", c.Config.Database).
		Msg("Successfully connected to ClickHouse")

	return nil
}

// Health checks the health of ClickHouse connection
func (c *Client) Health() error {
	if c.Conn == nil {
		return fmt.Errorf("Client.Health: ClickHouse connection is nil")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.Conn.Ping(ctx); err != nil {
		return fmt.Errorf("Client.Health: ClickHouse ping failed: %w", err)
	}

	return nil
}

// Close closes the ClickHouse connection
func (c *Client) Close() error {
	if c.Conn != nil {
		return c.Conn.Close()
	}
	return nil
}

// TablePartsInfo holds parts count info for a table
type TablePartsInfo struct {
	Table      string
	PartsCount uint64
	TotalRows  uint64
}

// LogAsyncInsertSettings logs the current async insert settings
func (c *Client) LogAsyncInsertSettings(ctx context.Context) {
	rows, err := c.Conn.Query(ctx, `
		SELECT name, value 
		FROM system.settings 
		WHERE name LIKE '%async_insert%'
	`)
	if err != nil {
		c.Logger.Warn().Err(err).Msg("Failed to query async insert settings")
		return
	}
	defer rows.Close()

	c.Logger.Info().Msg("=== Async Insert Settings ===")
	for rows.Next() {
		var name, value string
		if err := rows.Scan(&name, &value); err != nil {
			continue
		}
		c.Logger.Info().
			Str("setting", name).
			Str("value", value).
			Msg("Async setting")
	}
}

// LogPartsStats logs the current parts count per table
// tablePrefix filters tables (e.g., "facebook_", "instagram_"). Empty means all tables.
func (c *Client) LogPartsStats(ctx context.Context, tablePrefix string) {
	var query string
	var args []interface{}

	if tablePrefix != "" {
		query = `
			SELECT 
				table,
				count() as parts_count,
				sum(rows) as total_rows
			FROM system.parts 
			WHERE active AND database = ? AND (table LIKE ? OR table LIKE ?)
			GROUP BY table
			ORDER BY parts_count DESC
		`
		args = []interface{}{c.Config.Database, tablePrefix + "%", "mv_%"}
	} else {
		query = `
			SELECT 
				table,
				count() as parts_count,
				sum(rows) as total_rows
			FROM system.parts 
			WHERE active AND database = ?
			GROUP BY table
			ORDER BY parts_count DESC
		`
		args = []interface{}{c.Config.Database}
	}

	rows, err := c.Conn.Query(ctx, query, args...)
	if err != nil {
		c.Logger.Warn().Err(err).Msg("Failed to query parts stats")
		logger.CaptureException(err, map[string]string{
			"component": "clickhouse",
			"operation": "query_parts_stats",
		}, nil)
		return
	}
	defer rows.Close()

	c.Logger.Info().Str("database", c.Config.Database).Str("filter", tablePrefix).Msg("=== Table Parts Stats ===")
	for rows.Next() {
		var info TablePartsInfo
		if err := rows.Scan(&info.Table, &info.PartsCount, &info.TotalRows); err != nil {
			continue
		}

		logEvent := c.Logger.Info().
			Str("table", info.Table).
			Uint64("parts_count", info.PartsCount).
			Uint64("total_rows", info.TotalRows)

		if info.PartsCount > 50000 {
			logEvent.Str("status", "CRITICAL")
			logger.CaptureException(
				fmt.Errorf("Client.LogPartsStats: ClickHouse table %s has CRITICAL parts count: %d", info.Table, info.PartsCount),
				map[string]string{
					"component": "clickhouse",
					"table":     info.Table,
					"database":  c.Config.Database,
					"status":    "CRITICAL",
				},
				map[string]interface{}{
					"parts_count": info.PartsCount,
					"total_rows":  info.TotalRows,
				},
			)
		} else if info.PartsCount > 10000 {
			logEvent.Str("status", "WARNING")
			logger.CaptureException(
				fmt.Errorf("Client.LogPartsStats: ClickHouse table %s has WARNING parts count: %d", info.Table, info.PartsCount),
				map[string]string{
					"component": "clickhouse",
					"table":     info.Table,
					"database":  c.Config.Database,
					"status":    "WARNING",
				},
				map[string]interface{}{
					"parts_count": info.PartsCount,
					"total_rows":  info.TotalRows,
				},
			)
		} else {
			logEvent.Str("status", "OK")
		}
		logEvent.Msg("Table parts info")
	}
}

// LogAsyncInsertQueue logs pending async inserts
func (c *Client) LogAsyncInsertQueue(ctx context.Context) {
	rows, err := c.Conn.Query(ctx, `
		SELECT 
			database,
			table,
			count() as pending_queries
		FROM system.asynchronous_inserts
		GROUP BY database, table
	`)
	if err != nil {
		c.Logger.Warn().Err(err).Msg("Failed to query async insert queue (table may not exist in this ClickHouse version)")
		return
	}
	defer rows.Close()

	c.Logger.Info().Msg("=== Async Insert Queue ===")
	hasRows := false
	for rows.Next() {
		hasRows = true
		var database, table string
		var pendingQueries uint64
		if err := rows.Scan(&database, &table, &pendingQueries); err != nil {
			continue
		}
		c.Logger.Info().
			Str("database", database).
			Str("table", table).
			Uint64("pending_queries", pendingQueries).
			Msg("Async insert queue")
	}
	if !hasRows {
		c.Logger.Info().Msg("No pending async inserts")
	}
}

// StartMonitoring starts a background goroutine to log stats periodically
// tablePrefix filters tables (e.g., "facebook_", "instagram_"). Empty means all tables.
func (c *Client) StartMonitoring(ctx context.Context, interval time.Duration, tablePrefix string) {
	go func() {
		// Log settings once at startup
		c.LogAsyncInsertSettings(ctx)

		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				c.LogPartsStats(ctx, tablePrefix)
				c.LogAsyncInsertQueue(ctx)
			}
		}
	}()
	c.Logger.Info().Dur("interval", interval).Str("table_prefix", tablePrefix).Msg("Started ClickHouse monitoring")
}
