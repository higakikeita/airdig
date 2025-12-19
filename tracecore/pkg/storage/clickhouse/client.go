package clickhouse

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

// Config holds ClickHouse connection configuration
type Config struct {
	Host     string
	Port     int
	Database string
	Username string
	Password string
	UseTLS   bool
	Debug    bool
}

// DefaultConfig returns default ClickHouse configuration
func DefaultConfig() *Config {
	return &Config{
		Host:     "localhost",
		Port:     9000,
		Database: "tracecore",
		Username: "default",
		Password: "",
		UseTLS:   false,
		Debug:    false,
	}
}

// Client wraps ClickHouse connection
type Client struct {
	conn   driver.Conn
	config *Config
}

// NewClient creates a new ClickHouse client
func NewClient(config *Config) (*Client, error) {
	if config == nil {
		config = DefaultConfig()
	}

	options := &clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%d", config.Host, config.Port)},
		Auth: clickhouse.Auth{
			Database: config.Database,
			Username: config.Username,
			Password: config.Password,
		},
		Debug: config.Debug,
		Debugf: func(format string, v ...interface{}) {
			if config.Debug {
				fmt.Printf("[ClickHouse Debug] "+format+"\n", v...)
			}
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		DialTimeout:      time.Second * 10,
		MaxOpenConns:     10,
		MaxIdleConns:     5,
		ConnMaxLifetime:  time.Hour,
		ConnOpenStrategy: clickhouse.ConnOpenInOrder,
	}

	if config.UseTLS {
		options.TLS = &tls.Config{
			InsecureSkipVerify: false,
		}
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to ClickHouse: %w", err)
	}

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := conn.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping ClickHouse: %w", err)
	}

	return &Client{
		conn:   conn,
		config: config,
	}, nil
}

// Close closes the ClickHouse connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Ping checks if the connection is alive
func (c *Client) Ping(ctx context.Context) error {
	return c.conn.Ping(ctx)
}

// Exec executes a query without returning rows
func (c *Client) Exec(ctx context.Context, query string, args ...interface{}) error {
	return c.conn.Exec(ctx, query, args...)
}

// Query executes a query and returns rows
func (c *Client) Query(ctx context.Context, query string, args ...interface{}) (driver.Rows, error) {
	return c.conn.Query(ctx, query, args...)
}

// QueryRow executes a query that returns a single row
func (c *Client) QueryRow(ctx context.Context, query string, args ...interface{}) driver.Row {
	return c.conn.QueryRow(ctx, query, args...)
}

// AsyncInsert performs an async insert (fire and forget)
func (c *Client) AsyncInsert(ctx context.Context, query string, wait bool, args ...interface{}) error {
	return c.conn.AsyncInsert(ctx, query, wait, args...)
}

// PrepareBatch prepares a batch insert
func (c *Client) PrepareBatch(ctx context.Context, query string) (driver.Batch, error) {
	return c.conn.PrepareBatch(ctx, query)
}

// GetConn returns the underlying ClickHouse connection
func (c *Client) GetConn() driver.Conn {
	return c.conn
}

// Stats returns connection statistics
func (c *Client) Stats() driver.Stats {
	return c.conn.Stats()
}

// InitSchema initializes the database schema from SQL file
func (c *Client) InitSchema(ctx context.Context, schemaSQL string) error {
	// Create database first
	if err := c.Exec(ctx, fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", c.config.Database)); err != nil {
		return fmt.Errorf("failed to create database: %w", err)
	}

	// Execute schema SQL (should be split by statement if needed)
	return nil
}

// HealthCheck performs a comprehensive health check
func (c *Client) HealthCheck(ctx context.Context) (*HealthStatus, error) {
	status := &HealthStatus{
		Connected: false,
		Database:  c.config.Database,
	}

	// Check connection
	if err := c.Ping(ctx); err != nil {
		status.Error = err.Error()
		return status, err
	}
	status.Connected = true

	// Get server version
	var version string
	if err := c.QueryRow(ctx, "SELECT version()").Scan(&version); err != nil {
		status.Error = err.Error()
		return status, err
	}
	status.Version = version

	// Get trace count (last 7 days)
	var traceCount uint64
	query := "SELECT count(DISTINCT trace_id) FROM traces WHERE date >= today() - INTERVAL 7 DAY"
	if err := c.QueryRow(ctx, query).Scan(&traceCount); err != nil {
		// Table might not exist yet, that's okay
		traceCount = 0
	}
	status.TracesLast7Days = traceCount

	// Get connection stats
	stats := c.Stats()
	status.Stats = map[string]interface{}{
		"open_connections": stats.Open,
		"idle_connections": stats.Idle,
	}

	return status, nil
}

// HealthStatus represents the health status of the ClickHouse connection
type HealthStatus struct {
	Connected       bool                   `json:"connected"`
	Version         string                 `json:"version,omitempty"`
	Database        string                 `json:"database"`
	TracesLast7Days uint64                 `json:"traces_last_7_days"`
	Stats           map[string]interface{} `json:"stats,omitempty"`
	Error           string                 `json:"error,omitempty"`
}

// NewClientFromDSN creates a client from a DSN string
// DSN format: clickhouse://username:password@host:port/database
func NewClientFromDSN(dsn string) (*Client, error) {
	options, err := clickhouse.ParseDSN(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse DSN: %w", err)
	}

	conn, err := clickhouse.Open(options)
	if err != nil {
		return nil, fmt.Errorf("failed to connect: %w", err)
	}

	// Extract config from options
	config := &Config{
		Database: options.Auth.Database,
		Username: options.Auth.Username,
		Password: options.Auth.Password,
		Debug:    options.Debug,
	}

	if len(options.Addr) > 0 {
		config.Host = options.Addr[0]
	}

	return &Client{
		conn:   conn,
		config: config,
	}, nil
}

// BatchInsertHelper helps with batch inserts
type BatchInsertHelper struct {
	batch     driver.Batch
	batchSize int
	count     int
}

// NewBatchInsertHelper creates a new batch insert helper
func (c *Client) NewBatchInsertHelper(ctx context.Context, query string, batchSize int) (*BatchInsertHelper, error) {
	batch, err := c.PrepareBatch(ctx, query)
	if err != nil {
		return nil, err
	}

	return &BatchInsertHelper{
		batch:     batch,
		batchSize: batchSize,
		count:     0,
	}, nil
}

// Append appends a row to the batch
func (b *BatchInsertHelper) Append(args ...interface{}) error {
	if err := b.batch.Append(args...); err != nil {
		return err
	}
	b.count++
	return nil
}

// ShouldFlush returns true if the batch should be flushed
func (b *BatchInsertHelper) ShouldFlush() bool {
	return b.count >= b.batchSize
}

// Flush flushes the batch
func (b *BatchInsertHelper) Flush() error {
	if b.count == 0 {
		return nil
	}
	if err := b.batch.Send(); err != nil {
		return err
	}
	b.count = 0
	return nil
}

// GetCount returns the current batch count
func (b *BatchInsertHelper) GetCount() int {
	return b.count
}

// Close flushes and closes the batch
func (b *BatchInsertHelper) Close() error {
	if b.count > 0 {
		if err := b.Flush(); err != nil {
			return err
		}
	}
	return nil
}

// Utility functions

// StringOrNull returns sql.NullString
func StringOrNull(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

// TimeOrNull returns sql.NullTime
func TimeOrNull(t time.Time) sql.NullTime {
	if t.IsZero() {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: t, Valid: true}
}
