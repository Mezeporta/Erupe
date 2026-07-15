package migrations

import (
	"embed"
	"fmt"
	"io/fs"
	"sort"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

//go:embed sql/*.sql
var migrationFS embed.FS

//go:embed seed/*.sql seed/*/*.json
var seedFS embed.FS

// Migrate creates the schema_version table if needed, detects existing databases
// (auto-marks baseline as applied), then runs all pending migrations in order.
// Each migration runs in its own transaction.
func Migrate(db *sqlx.DB, logger *zap.Logger) (int, error) {
	if err := ensureVersionTable(db); err != nil {
		return 0, fmt.Errorf("creating schema_version table: %w", err)
	}

	if err := detectExistingDB(db, logger); err != nil {
		return 0, fmt.Errorf("detecting existing database: %w", err)
	}

	migrations, err := readMigrations()
	if err != nil {
		return 0, fmt.Errorf("reading migration files: %w", err)
	}

	applied, err := appliedVersions(db)
	if err != nil {
		return 0, fmt.Errorf("querying applied versions: %w", err)
	}

	count := 0
	for _, m := range migrations {
		if applied[m.version] {
			continue
		}
		logger.Info(fmt.Sprintf("Applying migration %04d: %s", m.version, m.filename))
		if err := applyMigration(db, m); err != nil {
			return count, fmt.Errorf("applying %s: %w", m.filename, err)
		}
		count++
	}

	return count, nil
}

// ApplySeedData runs all seed/*.sql files and seed/<table>/*.json files. Not
// tracked in schema_version. Safe to run multiple times if seed files use ON
// CONFLICT DO NOTHING (SQL) or "onConflict" (JSON).
//
// JSON seed files (see seed_json.go) are a hand-editing-friendly alternative
// to SQL for plain tabular data; seed data needing real SQL logic (subqueries,
// idempotency guards, NOW()-relative rows) stays as .sql. Each JSON file's
// table comes from its enclosing seed/<table>/ directory rather than a field
// inside the file, so a directory can't disagree with its own contents and
// a file can never straddle two tables.
func ApplySeedData(db *sqlx.DB, logger *zap.Logger) (int, error) {
	seedDir, err := fs.Sub(seedFS, "seed")
	if err != nil {
		return 0, fmt.Errorf("opening seed directory: %w", err)
	}

	entries, err := fs.ReadDir(seedDir, ".")
	if err != nil {
		return 0, fmt.Errorf("reading seed directory: %w", err)
	}

	var sqlNames []string
	var tables []string
	for _, e := range entries {
		if e.IsDir() {
			tables = append(tables, e.Name())
		} else if strings.HasSuffix(e.Name(), ".sql") {
			sqlNames = append(sqlNames, e.Name())
		}
	}
	sort.Strings(sqlNames)
	sort.Strings(tables)

	count := 0
	for _, name := range sqlNames {
		data, err := fs.ReadFile(seedDir, name)
		if err != nil {
			return count, fmt.Errorf("reading seed file %s: %w", name, err)
		}
		logger.Info(fmt.Sprintf("Applying seed data: %s", name))
		if _, err := db.Exec(string(data)); err != nil {
			return count, fmt.Errorf("executing seed file %s: %w", name, err)
		}
		count++
	}

	jsonCount, err := applySeedJSONTables(db, logger, seedDir, tables)
	return count + jsonCount, err
}

// seedJSONFile pairs a JSON seed file with the table its enclosing directory
// names.
type seedJSONFile struct {
	table string
	path  string // "<table>/<name>.json", relative to seedDir
}

// applySeedJSONTables applies every <table>/*.json file under seedDir. Table
// directories are attempted in alphabetical order, but that order isn't
// guaranteed to satisfy foreign-key dependencies between seeded tables (e.g.
// "campaigns" sorts after "campaign_rewards", which references it) — rather
// than hardcode dependency order (reintroducing the coupling this format
// avoids elsewhere), a file that fails on a foreign-key violation is retried
// after the rest of the batch, until a full pass makes no progress.
func applySeedJSONTables(db *sqlx.DB, logger *zap.Logger, seedDir fs.FS, tables []string) (int, error) {
	var pending []seedJSONFile
	for _, table := range tables {
		if !identifierPattern.MatchString(table) {
			return 0, fmt.Errorf("invalid seed table directory name %q", table)
		}
		names, err := fs.ReadDir(seedDir, table)
		if err != nil {
			return 0, fmt.Errorf("reading seed/%s: %w", table, err)
		}
		var jsonNames []string
		for _, n := range names {
			if !n.IsDir() && strings.HasSuffix(n.Name(), ".json") {
				jsonNames = append(jsonNames, n.Name())
			}
		}
		sort.Strings(jsonNames)
		for _, name := range jsonNames {
			pending = append(pending, seedJSONFile{table: table, path: table + "/" + name})
		}
	}

	count := 0
	for len(pending) > 0 {
		var retry []seedJSONFile
		var lastErr error
		progressed := false
		for _, f := range pending {
			data, err := fs.ReadFile(seedDir, f.path)
			if err != nil {
				return count, fmt.Errorf("reading seed file %s: %w", f.path, err)
			}
			logger.Info(fmt.Sprintf("Applying seed data: seed/%s", f.path))
			if err := applySeedJSON(db, f.path, f.table, data); err != nil {
				if isForeignKeyViolation(err) {
					retry = append(retry, f)
					lastErr = err
					continue
				}
				return count, err
			}
			count++
			progressed = true
		}
		if !progressed {
			return count, fmt.Errorf("could not resolve seed apply order (circular or missing foreign key target?): %w", lastErr)
		}
		pending = retry
	}
	return count, nil
}

// Version returns the highest applied migration number, or 0 if none.
func Version(db *sqlx.DB) (int, error) {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(
		SELECT 1 FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name = 'schema_version'
	)`).Scan(&exists)
	if err != nil {
		return 0, err
	}
	if !exists {
		return 0, nil
	}

	var version int
	err = db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version").Scan(&version)
	return version, err
}

type migration struct {
	version  int
	filename string
	sql      string
}

func ensureVersionTable(db *sqlx.DB) error {
	_, err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_version (
		version    INTEGER PRIMARY KEY,
		filename   TEXT NOT NULL,
		applied_at TIMESTAMPTZ DEFAULT now()
	)`)
	return err
}

// detectExistingDB checks if the database has tables but no schema_version rows.
// If so, it marks the baseline migration (version 1) as already applied.
func detectExistingDB(db *sqlx.DB, logger *zap.Logger) error {
	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM schema_version").Scan(&count); err != nil {
		return err
	}
	if count > 0 {
		return nil // Already tracked
	}

	// Check if the database has any user tables (beyond schema_version itself)
	var tableCount int
	err := db.QueryRow(`SELECT COUNT(*) FROM information_schema.tables
		WHERE table_schema = 'public' AND table_name != 'schema_version'`).Scan(&tableCount)
	if err != nil {
		return err
	}
	if tableCount == 0 {
		return nil // Fresh database
	}

	// Existing database without migration tracking — mark baseline as applied
	logger.Info("Detected existing database without schema_version tracking, marking baseline as applied")
	_, err = db.Exec("INSERT INTO schema_version (version, filename) VALUES (1, '0001_init.sql')")
	return err
}

func readMigrations() ([]migration, error) {
	files, err := fs.ReadDir(migrationFS, "sql")
	if err != nil {
		return nil, err
	}

	var migrations []migration
	for _, f := range files {
		if f.IsDir() || !strings.HasSuffix(f.Name(), ".sql") {
			continue
		}
		version, err := parseVersion(f.Name())
		if err != nil {
			return nil, fmt.Errorf("parsing version from %s: %w", f.Name(), err)
		}
		data, err := migrationFS.ReadFile("sql/" + f.Name())
		if err != nil {
			return nil, err
		}
		migrations = append(migrations, migration{
			version:  version,
			filename: f.Name(),
			sql:      string(data),
		})
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].version < migrations[j].version
	})
	return migrations, nil
}

func parseVersion(filename string) (int, error) {
	parts := strings.SplitN(filename, "_", 2)
	if len(parts) < 2 {
		return 0, fmt.Errorf("invalid migration filename: %s (expected NNNN_description.sql)", filename)
	}
	return strconv.Atoi(parts[0])
}

func appliedVersions(db *sqlx.DB) (map[int]bool, error) {
	rows, err := db.Query("SELECT version FROM schema_version")
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	applied := make(map[int]bool)
	for rows.Next() {
		var v int
		if err := rows.Scan(&v); err != nil {
			return nil, err
		}
		applied[v] = true
	}
	return applied, rows.Err()
}

func applyMigration(db *sqlx.DB, m migration) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(m.sql); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.Exec(
		"INSERT INTO schema_version (version, filename) VALUES ($1, $2)",
		m.version, m.filename,
	); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}
