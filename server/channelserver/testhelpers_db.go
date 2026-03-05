package channelserver

import (
	"fmt"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"erupe-ce/server/channelserver/compression/nullcomp"
	"erupe-ce/server/migrations"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

var (
	testDBOnce        sync.Once
	testDB            *sqlx.DB
	testDBSetupFailed bool
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

// SetupTestDB creates a connection to the test database and applies the schema.
// The schema is applied only once per test binary via sync.Once. Subsequent calls
// only TRUNCATE data for test isolation, avoiding expensive pg_restore + patch cycles.
func SetupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()

	testDBOnce.Do(func() {
		config := DefaultTestDBConfig()
		connStr := fmt.Sprintf(
			"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
			config.Host, config.Port, config.User, config.Password, config.DBName,
		)

		db, err := sqlx.Open("postgres", connStr)
		if err != nil {
			testDBSetupFailed = true
			return
		}

		if err := db.Ping(); err != nil {
			_ = db.Close()
			testDBSetupFailed = true
			return
		}

		// Clean the database and apply schema once
		CleanTestDB(t, db)
		ApplyTestSchema(t, db)

		testDB = db
	})

	if testDBSetupFailed || testDB == nil {
		t.Skipf("Test database not available. Run: docker compose -f docker/docker-compose.test.yml up -d")
		return nil
	}

	// Truncate all data for test isolation (schema stays intact)
	truncateAllTables(t, testDB)

	return testDB
}

// CleanTestDB drops all objects in the public schema to ensure a clean state
func CleanTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()

	// Drop and recreate the public schema to remove all objects (tables, types, sequences, etc.)
	_, err := db.Exec(`DROP SCHEMA public CASCADE; CREATE SCHEMA public;`)
	if err != nil {
		t.Logf("Warning: Failed to clean database: %v", err)
	}
}

// ApplyTestSchema applies the database schema using the embedded migration system.
func ApplyTestSchema(t *testing.T, db *sqlx.DB) {
	t.Helper()

	logger, _ := zap.NewDevelopment()
	_, err := migrations.Migrate(db, logger.Named("test-migrations"))
	if err != nil {
		t.Fatalf("Failed to apply schema migrations: %v", err)
	}
}

// truncateAllTables truncates all tables in the public schema for test isolation.
// It retries on deadlock, which can occur when a previous test's goroutines still
// hold connections with in-flight DB operations.
func truncateAllTables(t *testing.T, db *sqlx.DB) {
	t.Helper()

	rows, err := db.Query("SELECT tablename FROM pg_tables WHERE schemaname = 'public'")
	if err != nil {
		t.Fatalf("Failed to list tables for truncation: %v", err)
	}
	defer func() { _ = rows.Close() }()

	var tables []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			t.Fatalf("Failed to scan table name: %v", err)
		}
		tables = append(tables, name)
	}

	if len(tables) == 0 {
		return
	}

	stmt := "TRUNCATE " + strings.Join(tables, ", ") + " CASCADE"
	const maxRetries = 3
	for attempt := 1; attempt <= maxRetries; attempt++ {
		_, err := db.Exec(stmt)
		if err == nil {
			return
		}
		if attempt < maxRetries {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		t.Fatalf("Failed to truncate tables after %d attempts: %v", maxRetries, err)
	}
}

// TeardownTestDB is a no-op. The shared DB connection is reused across tests
// and closed automatically at process exit.
func TeardownTestDB(t *testing.T, db *sqlx.DB) {
	t.Helper()
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
	saveData := make([]byte, 150000)                // Large enough for all game modes
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

// CreateTestGuild creates a test guild with the given leader and returns the guild ID
func CreateTestGuild(t *testing.T, db *sqlx.DB, leaderCharID uint32, name string) uint32 {
	t.Helper()

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	var guildID uint32
	err = tx.QueryRow(
		"INSERT INTO guilds (name, leader_id) VALUES ($1, $2) RETURNING id",
		name, leaderCharID,
	).Scan(&guildID)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("Failed to create test guild: %v", err)
	}

	_, err = tx.Exec(
		"INSERT INTO guild_characters (guild_id, character_id) VALUES ($1, $2)",
		guildID, leaderCharID,
	)
	if err != nil {
		_ = tx.Rollback()
		t.Fatalf("Failed to add leader to guild: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Failed to commit guild creation: %v", err)
	}

	return guildID
}

// CreateTestSignSession creates a sign session and returns the session ID.
func CreateTestSignSession(t *testing.T, db *sqlx.DB, userID uint32, token string) uint32 {
	t.Helper()

	var id uint32
	err := db.QueryRow(
		`INSERT INTO sign_sessions (user_id, token) VALUES ($1, $2) RETURNING id`,
		userID, token,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to create test sign session: %v", err)
	}
	return id
}

// CreateTestServer creates a server entry for testing.
func CreateTestServer(t *testing.T, db *sqlx.DB, serverID uint16) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO servers (server_id, current_players) VALUES ($1, 0)`,
		serverID,
	)
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}
}

// CreateTestUserBinary creates a user_binary row for the given character ID.
func CreateTestUserBinary(t *testing.T, db *sqlx.DB, charID uint32) {
	t.Helper()

	_, err := db.Exec(`INSERT INTO user_binary (id) VALUES ($1)`, charID)
	if err != nil {
		t.Fatalf("Failed to create test user_binary: %v", err)
	}
}

// CreateTestGachaShop creates a gacha shop entry and returns its ID.
func CreateTestGachaShop(t *testing.T, db *sqlx.DB, name string, gachaType int) uint32 {
	t.Helper()

	var id uint32
	err := db.QueryRow(
		`INSERT INTO gacha_shop (name, gacha_type, min_gr, min_hr, url_banner, url_feature, url_thumbnail, wide, recommended, hidden)
		VALUES ($1, $2, 0, 0, '', '', '', false, false, false) RETURNING id`,
		name, gachaType,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to create test gacha shop: %v", err)
	}
	return id
}

// CreateTestGachaEntry creates a gacha entry and returns its ID.
func CreateTestGachaEntry(t *testing.T, db *sqlx.DB, gachaID uint32, entryType int, weight int) uint32 {
	t.Helper()

	var id uint32
	err := db.QueryRow(
		`INSERT INTO gacha_entries (gacha_id, entry_type, weight, rarity, item_type, item_number, item_quantity, rolls, frontier_points, daily_limit)
		VALUES ($1, $2, $3, 1, 0, 0, 0, 1, 0, 0) RETURNING id`,
		gachaID, entryType, weight,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to create test gacha entry: %v", err)
	}
	return id
}

// CreateTestGachaItem creates a gacha item for an entry.
func CreateTestGachaItem(t *testing.T, db *sqlx.DB, entryID uint32, itemType uint8, itemID uint16, quantity uint16) {
	t.Helper()

	_, err := db.Exec(
		`INSERT INTO gacha_items (entry_id, item_type, item_id, quantity) VALUES ($1, $2, $3, $4)`,
		entryID, itemType, itemID, quantity,
	)
	if err != nil {
		t.Fatalf("Failed to create test gacha item: %v", err)
	}
}

// SetTestDB assigns a database to a Server and initializes all repositories.
// Use this in integration tests instead of setting s.server.db directly.
func SetTestDB(s *Server, db *sqlx.DB) {
	s.db = db
	s.charRepo = NewCharacterRepository(db)
	s.guildRepo = NewGuildRepository(db)
	s.userRepo = NewUserRepository(db)
	s.gachaRepo = NewGachaRepository(db)
	s.houseRepo = NewHouseRepository(db)
	s.festaRepo = NewFestaRepository(db)
	s.towerRepo = NewTowerRepository(db)
	s.rengokuRepo = NewRengokuRepository(db)
	s.mailRepo = NewMailRepository(db)
	s.stampRepo = NewStampRepository(db)
	s.distRepo = NewDistributionRepository(db)
	s.sessionRepo = NewSessionRepository(db)
	s.eventRepo = NewEventRepository(db)
	s.achievementRepo = NewAchievementRepository(db)
	s.shopRepo = NewShopRepository(db)
	s.cafeRepo = NewCafeRepository(db)
	s.goocooRepo = NewGoocooRepository(db)
	s.divaRepo = NewDivaRepository(db)
	s.miscRepo = NewMiscRepository(db)
	s.scenarioRepo = NewScenarioRepository(db)
	s.mercenaryRepo = NewMercenaryRepository(db)
}
