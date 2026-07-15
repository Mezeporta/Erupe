package migrations

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"github.com/jmoiron/sqlx"
)

// seedJSONFile is the JSON seed format: a sequence of table blocks applied in
// order, for operators who find hand-editing raw SQL INSERT statements
// unapproachable. Sits alongside seed/*.sql, which remains the format for
// seed data that needs real SQL logic (subqueries, NOW()-relative rows,
// idempotency guards) rather than plain tabular data.
type seedJSONFile struct {
	Blocks []seedJSONBlock `json:"blocks"`
}

type seedJSONBlock struct {
	Table string `json:"table"`
	// Comment is free-text documentation carried over from the original SQL
	// file header; it has no functional effect on seeding.
	Comment string `json:"comment,omitempty"`
	// Truncate empties the table before inserting, matching a handful of
	// seed files that reset their table on every run instead of relying on
	// ON CONFLICT DO NOTHING.
	Truncate   bool            `json:"truncate,omitempty"`
	OnConflict string          `json:"onConflict,omitempty"`
	Columns    []string        `json:"columns"`
	Rows       [][]interface{} `json:"rows"`
}

// identifierPattern restricts table/column names to safe SQL identifiers,
// optionally schema-qualified (e.g. "public.shop_items"). These names are
// interpolated directly into the query text since they can't be bound as
// placeholders, so validation is mandatory even though the source is a
// compile-time embedded file rather than user input.
var identifierPattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*(\.[a-zA-Z_][a-zA-Z0-9_]*)?$`)

func applySeedJSON(db *sqlx.DB, name string, data []byte) error {
	var f seedJSONFile
	if err := json.Unmarshal(data, &f); err != nil {
		return fmt.Errorf("parsing %s: %w", name, err)
	}
	for i, block := range f.Blocks {
		if err := applySeedJSONBlock(db, block); err != nil {
			return fmt.Errorf("%s block %d (%s): %w", name, i, block.Table, err)
		}
	}
	return nil
}

func applySeedJSONBlock(db *sqlx.DB, block seedJSONBlock) error {
	if !identifierPattern.MatchString(block.Table) {
		return fmt.Errorf("invalid table name %q", block.Table)
	}
	for _, col := range block.Columns {
		if !identifierPattern.MatchString(col) {
			return fmt.Errorf("invalid column name %q", col)
		}
	}
	if len(block.Rows) == 0 {
		return nil
	}

	if block.Truncate {
		if _, err := db.Exec(fmt.Sprintf("TRUNCATE %s", block.Table)); err != nil {
			return fmt.Errorf("truncating: %w", err)
		}
	}

	query, args, err := buildSeedInsert(block)
	if err != nil {
		return err
	}
	if _, err := db.Exec(query, args...); err != nil {
		return fmt.Errorf("inserting: %w", err)
	}
	return nil
}

func buildSeedInsert(block seedJSONBlock) (string, []interface{}, error) {
	var sb strings.Builder
	fmt.Fprintf(&sb, "INSERT INTO %s (%s) VALUES ", block.Table, strings.Join(block.Columns, ", "))

	var args []interface{}
	for r, row := range block.Rows {
		if len(row) != len(block.Columns) {
			return "", nil, fmt.Errorf("row %d has %d values, want %d", r, len(row), len(block.Columns))
		}
		if r > 0 {
			sb.WriteString(", ")
		}
		sb.WriteByte('(')
		for c, val := range row {
			if c > 0 {
				sb.WriteString(", ")
			}
			raw, isRaw, err := rawSQLValue(val)
			if err != nil {
				return "", nil, fmt.Errorf("row %d column %s: %w", r, block.Columns[c], err)
			}
			if isRaw {
				sb.WriteString(string(raw))
				continue
			}
			args = append(args, normalizeSeedValue(val))
			fmt.Fprintf(&sb, "$%d", len(args))
		}
		sb.WriteByte(')')
	}

	if block.OnConflict != "" {
		fmt.Fprintf(&sb, " ON CONFLICT %s", block.OnConflict)
	}
	return sb.String(), args, nil
}

// rawSQLValue detects the {"raw": "..."} escape hatch for values JSON has no
// literal syntax for, namely computed SQL expressions like NOW().
func rawSQLValue(val interface{}) (string, bool, error) {
	m, ok := val.(map[string]interface{})
	if !ok {
		return "", false, nil
	}
	rawVal, ok := m["raw"]
	if !ok || len(m) != 1 {
		return "", false, fmt.Errorf(`object value must be {"raw": "<sql>"}, got %v`, m)
	}
	s, ok := rawVal.(string)
	if !ok {
		return "", false, fmt.Errorf(`"raw" value must be a string, got %v`, rawVal)
	}
	return s, true, nil
}

// normalizeSeedValue converts JSON's float64 number decoding back to an
// integer when the source literal had no fractional part, so integer
// columns receive "500000" rather than "500000.0"-shaped parameters.
func normalizeSeedValue(val interface{}) interface{} {
	f, ok := val.(float64)
	if !ok {
		return val
	}
	if f == math.Trunc(f) {
		return int64(f)
	}
	return f
}
