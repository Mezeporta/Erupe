// Package db provides a transparent database wrapper that rewrites
// PostgreSQL-style SQL for SQLite when needed.
package db

import (
	"context"
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// DB wraps *sqlx.DB and transparently adapts PostgreSQL queries for SQLite.
// For PostgreSQL, all methods are simple pass-throughs with zero overhead.
type DB struct {
	*sqlx.DB
	sqlite bool
}

// Wrap creates a DB wrapper around an existing *sqlx.DB connection.
// Returns nil if db is nil.
func Wrap(db *sqlx.DB) *DB {
	if db == nil {
		return nil
	}
	return &DB{
		DB:     db,
		sqlite: IsSQLite(db),
	}
}

func (d *DB) adapt(query string) string {
	if !d.sqlite {
		return query
	}
	return AdaptSQL(query)
}

// Exec executes a query without returning any rows.
func (d *DB) Exec(query string, args ...interface{}) (sql.Result, error) {
	return d.DB.Exec(d.adapt(query), args...)
}

// ExecContext executes a query without returning any rows.
func (d *DB) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return d.DB.ExecContext(ctx, d.adapt(query), args...)
}

// Query executes a query that returns rows.
func (d *DB) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.Query(d.adapt(query), args...)
}

// QueryContext executes a query that returns rows.
func (d *DB) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return d.DB.QueryContext(ctx, d.adapt(query), args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (d *DB) QueryRow(query string, args ...interface{}) *sql.Row {
	return d.DB.QueryRow(d.adapt(query), args...)
}

// QueryRowContext executes a query that is expected to return at most one row.
func (d *DB) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return d.DB.QueryRowContext(ctx, d.adapt(query), args...)
}

// Get queries a single row and scans it into dest.
func (d *DB) Get(dest interface{}, query string, args ...interface{}) error {
	return d.DB.Get(dest, d.adapt(query), args...)
}

// Select queries multiple rows and scans them into dest.
func (d *DB) Select(dest interface{}, query string, args ...interface{}) error {
	return d.DB.Select(dest, d.adapt(query), args...)
}

// Queryx executes a query that returns sqlx.Rows.
func (d *DB) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return d.DB.Queryx(d.adapt(query), args...)
}

// QueryRowx executes a query that returns a sqlx.Row.
func (d *DB) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return d.DB.QueryRowx(d.adapt(query), args...)
}

// QueryRowxContext executes a query that returns a sqlx.Row.
func (d *DB) QueryRowxContext(ctx context.Context, query string, args ...interface{}) *sqlx.Row {
	return d.DB.QueryRowxContext(ctx, d.adapt(query), args...)
}

// BeginTxx starts a new transaction with context and options.
// The returned Tx wrapper adapts queries the same way as DB.
func (d *DB) BeginTxx(ctx context.Context, opts *sql.TxOptions) (*Tx, error) {
	tx, err := d.DB.BeginTxx(ctx, opts)
	if err != nil {
		return nil, err
	}
	return &Tx{Tx: tx, sqlite: d.sqlite}, nil
}

// Tx wraps *sqlx.Tx and transparently adapts PostgreSQL queries for SQLite.
type Tx struct {
	*sqlx.Tx
	sqlite bool
}

func (t *Tx) adapt(query string) string {
	if !t.sqlite {
		return query
	}
	return AdaptSQL(query)
}

// Exec executes a query without returning any rows.
func (t *Tx) Exec(query string, args ...interface{}) (sql.Result, error) {
	return t.Tx.Exec(t.adapt(query), args...)
}

// ExecContext executes a query without returning any rows.
func (t *Tx) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return t.Tx.ExecContext(ctx, t.adapt(query), args...)
}

// Query executes a query that returns rows.
func (t *Tx) Query(query string, args ...interface{}) (*sql.Rows, error) {
	return t.Tx.Query(t.adapt(query), args...)
}

// QueryRow executes a query that is expected to return at most one row.
func (t *Tx) QueryRow(query string, args ...interface{}) *sql.Row {
	return t.Tx.QueryRow(t.adapt(query), args...)
}

// Queryx executes a query that returns sqlx.Rows.
func (t *Tx) Queryx(query string, args ...interface{}) (*sqlx.Rows, error) {
	return t.Tx.Queryx(t.adapt(query), args...)
}

// QueryRowx executes a query that returns a sqlx.Row.
func (t *Tx) QueryRowx(query string, args ...interface{}) *sqlx.Row {
	return t.Tx.QueryRowx(t.adapt(query), args...)
}

// Get queries a single row and scans it into dest.
func (t *Tx) Get(dest interface{}, query string, args ...interface{}) error {
	return t.Tx.Get(dest, t.adapt(query), args...)
}
