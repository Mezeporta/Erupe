package api

import (
	"database/sql"
	"testing"

	_config "erupe-ce/config"
	"go.uber.org/zap"

	"github.com/jmoiron/sqlx"
)

// MockDB provides a mock database for testing
type MockDB struct {
	QueryRowFunc    func(query string, args ...interface{}) *sql.Row
	QueryFunc       func(query string, args ...interface{}) (*sql.Rows, error)
	ExecFunc        func(query string, args ...interface{}) (sql.Result, error)
	QueryRowContext func(ctx interface{}, query string, args ...interface{}) *sql.Row
	GetContext      func(ctx interface{}, dest interface{}, query string, args ...interface{}) error
	SelectContext   func(ctx interface{}, dest interface{}, query string, args ...interface{}) error
}

// NewTestLogger creates a logger for testing
func NewTestLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}
	return logger
}

// NewTestConfig creates a default test configuration
func NewTestConfig() *_config.Config {
	return &_config.Config{
		API: _config.API{
			Port:        8000,
			PatchServer: "http://localhost:8080",
			Banners:     []_config.APISignBanner{},
			Messages:    []_config.APISignMessage{},
			Links:       []_config.APISignLink{},
		},
		Screenshots: _config.ScreenshotsOptions{
			Enabled:       true,
			OutputDir:     "/tmp/screenshots",
			UploadQuality: 85,
		},
		DebugOptions: _config.DebugOptions{
			MaxLauncherHR: false,
		},
		GameplayOptions: _config.GameplayOptions{
			MezFesSoloTickets:     100,
			MezFesGroupTickets:    50,
			MezFesDuration:        604800, // 1 week
			MezFesSwitchMinigame:  false,
		},
		LoginNotices:   []string{"Welcome to Erupe!"},
		HideLoginNotice: false,
	}
}

// NewTestAPIServer creates an API server for testing with a real database
func NewTestAPIServer(t *testing.T, db *sqlx.DB) *APIServer {
	logger := NewTestLogger(t)
	cfg := NewTestConfig()
	config := &Config{
		Logger:      logger,
		DB:          db,
		ErupeConfig: cfg,
	}
	return NewAPIServer(config)
}

// CleanupTestData removes test data from the database
func CleanupTestData(t *testing.T, db *sqlx.DB, userID uint32) {
	// Delete characters associated with the user
	_, err := db.Exec("DELETE FROM characters WHERE user_id = $1", userID)
	if err != nil {
		t.Logf("Error cleaning up characters: %v", err)
	}

	// Delete sign sessions for the user
	_, err = db.Exec("DELETE FROM sign_sessions WHERE user_id = $1", userID)
	if err != nil {
		t.Logf("Error cleaning up sign_sessions: %v", err)
	}

	// Delete the user
	_, err = db.Exec("DELETE FROM users WHERE id = $1", userID)
	if err != nil {
		t.Logf("Error cleaning up users: %v", err)
	}
}

// GetTestDBConnection returns a test database connection (requires database to be running)
func GetTestDBConnection(t *testing.T) *sqlx.DB {
	// This function would need to connect to a test database
	// For now, it's a placeholder that returns nil
	// In practice, you'd use a test database container or mock
	return nil
}
