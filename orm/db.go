package orm

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct {
	pool   *pgxpool.Pool
	config *DBConfig
	mu     sync.RWMutex
}

type DBConfig struct {
	Host              string
	Port              int
	Name              string
	User              string
	Password          string
	SSLMode           string
	MaxConns          int32
	MinConns          int32
	MaxConnLifetime   time.Duration
	MaxConnIdleTime   time.Duration
	HealthCheckPeriod time.Duration
}

func DefaultDBConfig() *DBConfig {
	return &DBConfig{
		Host:              "localhost",
		Port:              5432,
		Name:              "",
		User:              "",
		Password:          "",
		SSLMode:           "prefer",
		MaxConns:          10,
		MinConns:          2,
		MaxConnLifetime:   30 * time.Minute,
		MaxConnIdleTime:   5 * time.Minute,
		HealthCheckPeriod: 30 * time.Second,
	}
}

func DBConfigFromSettings(host string, port int, name, user, password string) *DBConfig {
	cfg := DefaultDBConfig()
	cfg.Host = host
	cfg.Port = port
	cfg.Name = name
	cfg.User = user
	cfg.Password = password
	return cfg
}

func (c *DBConfig) DSN() string {
	sslMode := c.SSLMode
	if sslMode == "" {
		sslMode = "prefer"
	}
	return fmt.Sprintf("host=%s port=%d dbname=%s user=%s password=%s sslmode=%s",
		c.Host, c.Port, c.Name, c.User, c.Password, sslMode)
}

func OpenDB(ctx context.Context, config *DBConfig) (*DB, error) {
	poolConfig, err := pgxpool.ParseConfig(config.DSN())
	if err != nil {
		return nil, fmt.Errorf("orm: failed to parse db config: %w", err)
	}

	poolConfig.MaxConns = config.MaxConns
	poolConfig.MinConns = config.MinConns
	poolConfig.MaxConnLifetime = config.MaxConnLifetime
	poolConfig.MaxConnIdleTime = config.MaxConnIdleTime
	poolConfig.HealthCheckPeriod = config.HealthCheckPeriod

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("orm: failed to create connection pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("orm: failed to ping database: %w", err)
	}

	slog.Info("orm: database connection pool established", "host", config.Host, "db", config.Name)

	return &DB{
		pool:   pool,
		config: config,
	}, nil
}

func (db *DB) Close() {
	if db.pool != nil {
		db.pool.Close()
	}
}

func (db *DB) Ping(ctx context.Context) error {
	return db.pool.Ping(ctx)
}

func (db *DB) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	rows, err := db.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("orm: query error: %w", err)
	}
	return &pgxRows{rows: rows}, nil
}

func (db *DB) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	return &pgxRow{row: db.pool.QueryRow(ctx, sql, args...)}
}

func (db *DB) Exec(ctx context.Context, sql string, args ...interface{}) (Result, error) {
	tag, err := db.pool.Exec(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("orm: exec error: %w", err)
	}
	return &pgxResult{tag: tag}, nil
}

func (db *DB) Begin(ctx context.Context) (Tx, error) {
	tx, err := db.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("orm: begin transaction: %w", err)
	}
	return &pgxTx{tx: tx}, nil
}

func (db *DB) Pool() *pgxpool.Pool {
	return db.pool
}

type Rows interface {
	Next() bool
	Scan(dest ...interface{}) error
	Columns() ([]string, error)
	ColumnTypes() ([]ColumnType, error)
	Close() error
	Err() error
}

type ColumnType interface {
	Name() string
}

type Row interface {
	Scan(dest ...interface{}) error
	Columns() ([]string, error)
}

type Result interface {
	RowsAffected() int64
	LastInsertId() (int64, bool)
}

type Tx interface {
	Commit() error
	Rollback() error
	Exec(ctx context.Context, sql string, args ...interface{}) (Result, error)
	Query(ctx context.Context, sql string, args ...interface{}) (Rows, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) Row
	Savepoint(ctx context.Context, name string) error
	RollbackToSavepoint(ctx context.Context, name string) error
}

type pgxRows struct {
	rows pgx.Rows
}

func (r *pgxRows) Next() bool                 { return r.rows.Next() }
func (r *pgxRows) Scan(dest ...interface{}) error { return r.rows.Scan(dest...) }
func (r *pgxRows) Columns() ([]string, error) {
	fd := r.rows.FieldDescriptions()
	cols := make([]string, len(fd))
	for i, f := range fd {
		cols[i] = string(f.Name)
	}
	return cols, nil
}
func (r *pgxRows) ColumnTypes() ([]ColumnType, error) { return nil, nil }
func (r *pgxRows) Close() error                        { r.rows.Close(); return nil }
func (r *pgxRows) Err() error                          { return r.rows.Err() }

type pgxRow struct {
	row pgx.Row
}

func (r *pgxRow) Scan(dest ...interface{}) error { return r.row.Scan(dest...) }
func (r *pgxRow) Columns() ([]string, error)     { return nil, fmt.Errorf("orm: Columns not available on single row") }

type pgxResult struct {
	tag pgconn.CommandTag
}

func (r *pgxResult) RowsAffected() int64     { return r.tag.RowsAffected() }
func (r *pgxResult) LastInsertId() (int64, bool) { return 0, false }

type pgxTx struct {
	tx pgx.Tx
}

func (t *pgxTx) Commit() error   { return t.tx.Commit(context.Background()) }
func (t *pgxTx) Rollback() error { return t.tx.Rollback(context.Background()) }

func (t *pgxTx) Exec(ctx context.Context, sql string, args ...interface{}) (Result, error) {
	tag, err := t.tx.Exec(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgxResult{tag: tag}, nil
}

func (t *pgxTx) Query(ctx context.Context, sql string, args ...interface{}) (Rows, error) {
	rows, err := t.tx.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	return &pgxRows{rows: rows}, nil
}

func (t *pgxTx) QueryRow(ctx context.Context, sql string, args ...interface{}) Row {
	row := t.tx.QueryRow(ctx, sql, args...)
	return &pgxRow{row: row}
}

func (t *pgxTx) Savepoint(ctx context.Context, name string) error {
	_, err := t.tx.Exec(ctx, "SAVEPOINT "+quote(name))
	return err
}

func (t *pgxTx) RollbackToSavepoint(ctx context.Context, name string) error {
	_, err := t.tx.Exec(ctx, "ROLLBACK TO SAVEPOINT "+quote(name))
	return err
}