package main

import (
	"database/sql"
	"flag"
	"fmt"
	"os"

	_ "github.com/lib/pq"
)

// DBConfig holds database connection configuration.
type DBConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	DBName   string
}

// addDBFlags adds common database flags to a FlagSet.
func addDBFlags(fs *flag.FlagSet, cfg *DBConfig) {
	fs.StringVar(&cfg.Host, "host", getEnvOrDefault("ERUPE_DB_HOST", "localhost"), "Database host")
	fs.IntVar(&cfg.Port, "port", 5432, "Database port")
	fs.StringVar(&cfg.User, "user", getEnvOrDefault("ERUPE_DB_USER", "postgres"), "Database user")
	fs.StringVar(&cfg.Password, "password", os.Getenv("ERUPE_DB_PASSWORD"), "Database password")
	fs.StringVar(&cfg.DBName, "dbname", getEnvOrDefault("ERUPE_DB_NAME", "erupe"), "Database name")
}

// getEnvOrDefault returns the environment variable value or a default.
func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// connectDB establishes a connection to the PostgreSQL database.
func connectDB(cfg *DBConfig) (*sql.DB, error) {
	if cfg.Password == "" {
		return nil, fmt.Errorf("database password is required (use -password flag or ERUPE_DB_PASSWORD env var)")
	}

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.DBName,
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return db, nil
}

// ConnectedUser represents a user currently connected to the server.
type ConnectedUser struct {
	CharID     uint32
	CharName   string
	ServerID   uint16
	ServerName string
	UserID     int
	Username   string
	LastLogin  sql.NullTime
	HR         int
	GR         int
}

// ServerStatus represents the status of a channel server.
type ServerStatus struct {
	ServerID       uint16
	WorldName      string
	WorldDesc      string
	Land           int
	CurrentPlayers int
	Season         int
}

// LoginHistory represents a player's login history entry.
type LoginHistory struct {
	CharID    uint32
	CharName  string
	LastLogin sql.NullTime
	HR        int
	GR        int
	Username  string
}
