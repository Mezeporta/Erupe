package channelserver

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"erupe-ce/server/channelserver/compression/nullcomp"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// TestDBConfig holds the configuration for the test database
type TestDBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
}

// DefaultTestDBConfig returns the default test database configuration
// that matches docker-compose.test.yml
func DefaultTestDBConfig() *TestDBConfig {
	return &TestDBConfig{
		Host:     getEnv("TEST_DB_HOST", "localhost"),
		Port:     getEnv("TEST_DB_PORT", "5433"),
		User:     getEnv("TEST_DB_USER", "test"),
		Password: getEnv("TEST_DB_PASSWORD", "test"),
		DBName:   getEnv("TEST_DB_NAME", "erupe_test"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// SetupTestDB creates a connection to the test database and applies the schema
func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	config := DefaultTestDBConfig()
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.DBName,
	)

	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		t.Skipf("Failed to connect to test database: %v. Run: docker compose -f docker/docker-compose.test.yml up -d", err)
		return nil
	}

	// Test connection
	if err := db.Ping(); err != nil {
		_ = db.Close()
		t.Skipf("Test database not available: %v. Run: docker compose -f docker/docker-compose.test.yml up -d", err)
		return nil
	}

	// Clean the database before tests
	CleanTestDB(t, db)

	// Apply schema
	ApplyTestSchema(t, db)

	return db
}

// CleanTestDB drops all tables to ensure a clean state
func CleanTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// Drop all tables in the public schema
	_, err := db.Exec(`
		DO $$ DECLARE
			r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'DROP TABLE IF EXISTS ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)
	if err != nil {
		t.Logf("Warning: Failed to clean database: %v", err)
	}
}

// ApplyTestSchema applies the database schema from init.sql using pg_restore
func ApplyTestSchema(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// Find the project root (where schemas/ directory is located)
	projectRoot := findProjectRoot(t)
	schemaPath := filepath.Join(projectRoot, "schemas", "init.sql")

	// Get the connection config
	config := DefaultTestDBConfig()

	// Use pg_restore to load the schema dump
	// The init.sql file is a pg_dump custom format, so we need pg_restore
	cmd := exec.Command("pg_restore",
		"-h", config.Host,
		"-p", config.Port,
		"-U", config.User,
		"-d", config.DBName,
		"--no-owner",
		"--no-acl",
		"-c", // clean (drop) before recreating
		schemaPath,
	)
	cmd.Env = append(os.Environ(), fmt.Sprintf("PGPASSWORD=%s", config.Password))

	output, err := cmd.CombinedOutput()
	if err != nil {
		// pg_restore may error on first run (no tables to drop), that's usually ok
		t.Logf("pg_restore output: %s", string(output))
		// Check if it's a fatal error
		if !strings.Contains(string(output), "does not exist") {
			t.Logf("pg_restore error (may be non-fatal): %v", err)
		}
	}

	// Apply patch schemas in order
	applyPatchSchemas(t, db, projectRoot)
}

// applyPatchSchemas applies all patch schema files in numeric order
func applyPatchSchemas(t *testing.T, db *sqlx.DB, projectRoot string) {
	t.Helper()

	patchDir := filepath.Join(projectRoot, "schemas", "patch-schema")
	entries, err := os.ReadDir(patchDir)
	if err != nil {
		t.Logf("Warning: Could not read patch-schema directory: %v", err)
		return
	}

	// Sort patch files numerically
	var patchFiles []string
	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".sql") {
			patchFiles = append(patchFiles, entry.Name())
		}
	}
	sort.Strings(patchFiles)

	// Apply each patch in its own transaction
	for _, filename := range patchFiles {
		patchPath := filepath.Join(patchDir, filename)
		patchSQL, err := os.ReadFile(patchPath)
		if err != nil {
			t.Logf("Warning: Failed to read patch file %s: %v", filename, err)
			continue
		}

		// Start a new transaction for each patch
		tx, err := db.Begin()
		if err != nil {
			t.Logf("Warning: Failed to start transaction for patch %s: %v", filename, err)
			continue
		}

		_, err = tx.Exec(string(patchSQL))
		if err != nil {
			_ = tx.Rollback()
			t.Logf("Warning: Failed to apply patch %s: %v", filename, err)
			// Continue with other patches even if one fails
		} else {
			_ = tx.Commit()
		}
	}
}

// findProjectRoot finds the project root directory by looking for the schemas directory
func findProjectRoot(t *testing.T) string {
	t.Helper()

	// Start from current directory and walk up
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}

	for {
		schemasPath := filepath.Join(dir, "schemas")
		if stat, err := os.Stat(schemasPath); err == nil && stat.IsDir() {
			return dir
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("Could not find project root (schemas directory not found)")
		}
		dir = parent
	}
}

// TeardownTestDB closes the database connection
func TeardownTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
	if db != nil {
		_ = db.Close()
	}
}

// CreateTestUser creates a test user and returns the user ID
func CreateTestUser(t *testing.T, db *sqlx.DB, username string) uint32 {
	t.Helper()

	var userID uint32
	err := db.QueryRow(`
		INSERT INTO users (username, password, rights)
		VALUES ($1, 'test_password_hash', 0)
		RETURNING id
	`, username).Scan(&userID)

	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	return userID
}

// CreateTestCharacter creates a test character and returns the character ID
func CreateTestCharacter(t *testing.T, db *sqlx.DB, userID uint32, name string) uint32 {
	t.Helper()

	// Create minimal valid savedata (needs to be large enough for the game to parse)
	// The name is at offset 88, and various game mode pointers extend up to ~147KB for ZZ mode
	// We need at least 150KB to accommodate all possible pointer offsets
	saveData := make([]byte, 150000) // Large enough for all game modes
	copy(saveData[88:], append([]byte(name), 0x00)) // Name at offset 88 with null terminator

	// Import the nullcomp package for compression
	compressed, err := nullcomp.Compress(saveData)
	if err != nil {
		t.Fatalf("Failed to compress savedata: %v", err)
	}

	var charID uint32
	err = db.QueryRow(`
		INSERT INTO characters (user_id, is_female, is_new_character, name, unk_desc_string, gr, hr, weapon_type, last_login, savedata, decomyset, savemercenary)
		VALUES ($1, false, false, $2, '', 0, 0, 0, 0, $3, '', '')
		RETURNING id
	`, userID, name, compressed).Scan(&charID)

	if err != nil {
		t.Fatalf("Failed to create test character: %v", err)
	}

	return charID
}
