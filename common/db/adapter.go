// Package db provides a database adapter that transparently translates
// PostgreSQL-style SQL to the active driver's dialect.
//
// When the driver is "sqlite", queries are rewritten on the fly:
//   - $1, $2, ... → ?, ?, ...
//   - now()        → CURRENT_TIMESTAMP
//   - ::type casts → removed
//   - ILIKE        → LIKE  (SQLite LIKE is case-insensitive for ASCII)
package db

import (
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
)

// IsSQLite reports whether the given sqlx.DB is backed by a SQLite driver.
func IsSQLite(db *sqlx.DB) bool {
	return db.DriverName() == "sqlite" || db.DriverName() == "sqlite3"
}

// Adapt rewrites a PostgreSQL query for the active driver.
// For Postgres it's a no-op. For SQLite it translates placeholders and
// Postgres-specific syntax.
func Adapt(db *sqlx.DB, query string) string {
	if !IsSQLite(db) {
		return query
	}
	return AdaptSQL(query)
}

// castRe matches Postgres type casts like ::int, ::text, ::timestamptz,
// ::character varying, etc.
// castRe matches Postgres type casts: ::int, ::text, ::timestamptz,
// ::character varying(N), etc. The space is allowed only when followed
// by a word char (e.g. "character varying") to avoid eating trailing spaces.
var castRe = regexp.MustCompile(`::[a-zA-Z_]\w*(?:\s+\w+)*(?:\([^)]*\))?`)

// dollarParamRe matches Postgres-style positional parameters: $1, $2, etc.
var dollarParamRe = regexp.MustCompile(`\$\d+`)

// AdaptSQL translates a PostgreSQL query to SQLite-compatible SQL.
// Exported so it can be tested without a real DB connection.
func AdaptSQL(query string) string {
	// 1. Replace now() with CURRENT_TIMESTAMP
	query = strings.ReplaceAll(query, "now()", "CURRENT_TIMESTAMP")
	query = strings.ReplaceAll(query, "NOW()", "CURRENT_TIMESTAMP")

	// 2. Strip Postgres type casts (::int, ::text, ::timestamptz, etc.)
	query = castRe.ReplaceAllString(query, "")

	// 3. ILIKE → LIKE (SQLite LIKE is case-insensitive for ASCII by default)
	query = strings.ReplaceAll(query, " ILIKE ", " LIKE ")
	query = strings.ReplaceAll(query, " ilike ", " LIKE ")

	// 4. Strip "public." schema prefix (SQLite has no schemas)
	query = strings.ReplaceAll(query, "public.", "")

	// 5. TRUNCATE → DELETE FROM (SQLite has no TRUNCATE)
	query = strings.ReplaceAll(query, "TRUNCATE ", "DELETE FROM ")
	query = strings.ReplaceAll(query, "truncate ", "DELETE FROM ")

	// 6. Replace $1,$2,... → ?,?,...
	query = dollarParamRe.ReplaceAllString(query, "?")

	return query
}
