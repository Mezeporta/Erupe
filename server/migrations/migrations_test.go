package migrations

import (
	"fmt"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func testDB(t *testing.T) *sqlx.DB {
	t.Helper()

	host := getEnv("TEST_DB_HOST", "localhost")
	port := getEnv("TEST_DB_PORT", "5433")
	user := getEnv("TEST_DB_USER", "test")
	password := getEnv("TEST_DB_PASSWORD", "test")
	dbName := getEnv("TEST_DB_NAME", "erupe_test")

	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbName,
	)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Test database not available: %v", err)
		return nil
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Skipf("Test database not available: %v", err)
		return nil
	}

	// Clean slate
	_, err = db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public;")
	if err != nil {
		t.Fatalf("Failed to clean database: %v", err)
	}

	return db
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestMigrateEmptyDB(t *testing.T) {
	db := testDB(t)
	defer func() { _ = db.Close() }()

	logger, _ := zap.NewDevelopment()

	allMigrations, err := readMigrations()
	if err != nil {
		t.Fatalf("readMigrations failed: %v", err)
	}
	wantCount := len(allMigrations)
	wantVersion := allMigrations[wantCount-1].version

	applied, err := Migrate(db, logger)
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	if applied != wantCount {
		t.Errorf("expected %d migrations applied, got %d", wantCount, applied)
	}

	ver, err := Version(db)
	if err != nil {
		t.Fatalf("Version failed: %v", err)
	}
	if ver != wantVersion {
		t.Errorf("expected version %d, got %d", wantVersion, ver)
	}
}

func TestMigrateAlreadyMigrated(t *testing.T) {
	db := testDB(t)
	defer func() { _ = db.Close() }()

	logger, _ := zap.NewDevelopment()

	// First run
	_, err := Migrate(db, logger)
	if err != nil {
		t.Fatalf("First Migrate failed: %v", err)
	}

	// Second run should apply 0
	applied, err := Migrate(db, logger)
	if err != nil {
		t.Fatalf("Second Migrate failed: %v", err)
	}
	if applied != 0 {
		t.Errorf("expected 0 migrations on second run, got %d", applied)
	}
}

func TestMigrateExistingDBWithoutSchemaVersion(t *testing.T) {
	db := testDB(t)
	defer func() { _ = db.Close() }()

	logger, _ := zap.NewDevelopment()

	allMigrations, err := readMigrations()
	if err != nil {
		t.Fatalf("readMigrations failed: %v", err)
	}
	// Baseline (0001) is auto-marked, remaining are applied
	wantApplied := len(allMigrations) - 1
	wantVersion := allMigrations[len(allMigrations)-1].version

	// Simulate an existing database that has the full 0001 schema applied
	// but no schema_version tracking yet (pre-migration-system installs).
	// First, run all migrations normally to get the real schema...
	_, err = Migrate(db, logger)
	if err != nil {
		t.Fatalf("Initial Migrate failed: %v", err)
	}
	// ...then drop schema_version to simulate the pre-tracking state.
	_, err = db.Exec("DROP TABLE schema_version")
	if err != nil {
		t.Fatalf("Failed to drop schema_version: %v", err)
	}

	// Migrate should detect existing DB and auto-mark baseline,
	// then apply remaining migrations.
	applied, err := Migrate(db, logger)
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	if applied != wantApplied {
		t.Errorf("expected %d migrations applied (baseline auto-marked, rest applied), got %d", wantApplied, applied)
	}

	ver, err := Version(db)
	if err != nil {
		t.Fatalf("Version failed: %v", err)
	}
	if ver != wantVersion {
		t.Errorf("expected version %d, got %d", wantVersion, ver)
	}
}

func TestVersionEmptyDB(t *testing.T) {
	db := testDB(t)
	defer func() { _ = db.Close() }()

	ver, err := Version(db)
	if err != nil {
		t.Fatalf("Version failed: %v", err)
	}
	if ver != 0 {
		t.Errorf("expected version 0 on empty DB, got %d", ver)
	}
}

func TestApplySeedData(t *testing.T) {
	db := testDB(t)
	defer func() { _ = db.Close() }()

	logger, _ := zap.NewDevelopment()

	// Apply schema first
	_, err := Migrate(db, logger)
	if err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}

	count, err := ApplySeedData(db, logger)
	if err != nil {
		t.Fatalf("ApplySeedData failed: %v", err)
	}
	if count == 0 {
		t.Error("expected at least 1 seed file applied, got 0")
	}
}

func TestParseVersion(t *testing.T) {
	tests := []struct {
		filename string
		want     int
		wantErr  bool
	}{
		{"0001_init.sql", 1, false},
		{"0002_add_users.sql", 2, false},
		{"0100_big_change.sql", 100, false},
		{"bad.sql", 0, true},
	}
	for _, tt := range tests {
		got, err := parseVersion(tt.filename)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseVersion(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("parseVersion(%q) = %d, want %d", tt.filename, got, tt.want)
		}
	}
}

func TestReadMigrations(t *testing.T) {
	migrations, err := readMigrations()
	if err != nil {
		t.Fatalf("readMigrations failed: %v", err)
	}
	if len(migrations) == 0 {
		t.Fatal("expected at least 1 migration, got 0")
	}
	if migrations[0].version != 1 {
		t.Errorf("first migration version = %d, want 1", migrations[0].version)
	}
	if migrations[0].filename != "0001_init.sql" {
		t.Errorf("first migration filename = %q, want 0001_init.sql", migrations[0].filename)
	}
}

func TestReadMigrations_Sorted(t *testing.T) {
	migrations, err := readMigrations()
	if err != nil {
		t.Fatalf("readMigrations failed: %v", err)
	}
	for i := 1; i < len(migrations); i++ {
		if migrations[i].version <= migrations[i-1].version {
			t.Errorf("migrations not sorted: version %d at index %d follows version %d at index %d",
				migrations[i].version, i, migrations[i-1].version, i-1)
		}
	}
}

func TestReadMigrations_AllHaveSQL(t *testing.T) {
	migrations, err := readMigrations()
	if err != nil {
		t.Fatalf("readMigrations failed: %v", err)
	}
	for _, m := range migrations {
		if m.sql == "" {
			t.Errorf("migration %s has empty SQL", m.filename)
		}
	}
}

func TestReadMigrations_BaselineIsLargest(t *testing.T) {
	migrations, err := readMigrations()
	if err != nil {
		t.Fatalf("readMigrations failed: %v", err)
	}
	if len(migrations) < 2 {
		t.Skip("not enough migrations to compare sizes")
	}
	// The baseline (0001_init.sql) should be the largest migration.
	baselineLen := len(migrations[0].sql)
	for _, m := range migrations[1:] {
		if len(m.sql) > baselineLen {
			t.Errorf("migration %s (%d bytes) is larger than baseline (%d bytes)",
				m.filename, len(m.sql), baselineLen)
		}
	}
}

func TestParseVersion_Comprehensive(t *testing.T) {
	tests := []struct {
		filename string
		want     int
		wantErr  bool
	}{
		{"0001_init.sql", 1, false},
		{"0002_add_users.sql", 2, false},
		{"0100_big_change.sql", 100, false},
		{"9999_final.sql", 9999, false},
		{"bad.sql", 0, true},
		{"noseparator", 0, true},
		{"abc_description.sql", 0, true},
	}
	for _, tt := range tests {
		got, err := parseVersion(tt.filename)
		if (err != nil) != tt.wantErr {
			t.Errorf("parseVersion(%q) error = %v, wantErr %v", tt.filename, err, tt.wantErr)
			continue
		}
		if got != tt.want {
			t.Errorf("parseVersion(%q) = %d, want %d", tt.filename, got, tt.want)
		}
	}
}
