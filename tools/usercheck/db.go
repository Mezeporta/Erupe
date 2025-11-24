package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/lib/pq"
)

// DBConfig holds database connection configuration.
type DBConfig struct {
	Host       string
	Port       int
	User       string
	Password   string
	DBName     string
	ConfigPath string
}

// ErupeConfig represents the relevant parts of config.json.
type ErupeConfig struct {
	Database struct {
		Host     string `json:"Host"`
		Port     int    `json:"Port"`
		User     string `json:"User"`
		Password string `json:"Password"`
		Database string `json:"Database"`
	} `json:"Database"`
}

// loadConfigFile loads database settings from config.json.
func loadConfigFile(path string) (*ErupeConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg ErupeConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// findConfigFile searches for config.json in common locations.
func findConfigFile() string {
	// Check paths relative to current directory and up to project root
	paths := []string{
		"config.json",
		"../../config.json", // From tools/usercheck/
		"../../../config.json",
	}

	// Also check if we can find it via executable path
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		paths = append(paths,
			filepath.Join(dir, "config.json"),
			filepath.Join(dir, "../../config.json"),
		)
	}

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p
		}
	}

	return ""
}

// addDBFlags adds common database flags to a FlagSet.
func addDBFlags(fs *flag.FlagSet, cfg *DBConfig) {
	fs.StringVar(&cfg.ConfigPath, "config", "", "Path to config.json (auto-detected if not specified)")
	fs.StringVar(&cfg.Host, "host", "", "Database host (overrides config.json)")
	fs.IntVar(&cfg.Port, "port", 0, "Database port (overrides config.json)")
	fs.StringVar(&cfg.User, "user", "", "Database user (overrides config.json)")
	fs.StringVar(&cfg.Password, "password", "", "Database password (overrides config.json)")
	fs.StringVar(&cfg.DBName, "dbname", "", "Database name (overrides config.json)")
}

// resolveDBConfig resolves the final database configuration.
// Priority: CLI flags > environment variables > config.json > defaults
func resolveDBConfig(cfg *DBConfig) error {
	// Try to load from config.json
	configPath := cfg.ConfigPath
	if configPath == "" {
		configPath = findConfigFile()
	}

	var fileCfg *ErupeConfig
	if configPath != "" {
		var err error
		fileCfg, err = loadConfigFile(configPath)
		if err != nil {
			// Only error if user explicitly specified a config path
			if cfg.ConfigPath != "" {
				return fmt.Errorf("failed to load config file: %w", err)
			}
			// Otherwise just ignore and use defaults/flags
		}
	}

	// Apply config.json values as base
	if fileCfg != nil {
		if cfg.Host == "" {
			cfg.Host = fileCfg.Database.Host
		}
		if cfg.Port == 0 {
			cfg.Port = fileCfg.Database.Port
		}
		if cfg.User == "" {
			cfg.User = fileCfg.Database.User
		}
		if cfg.Password == "" {
			cfg.Password = fileCfg.Database.Password
		}
		if cfg.DBName == "" {
			cfg.DBName = fileCfg.Database.Database
		}
	}

	// Apply environment variables
	if cfg.Host == "" {
		cfg.Host = os.Getenv("ERUPE_DB_HOST")
	}
	if cfg.User == "" {
		cfg.User = os.Getenv("ERUPE_DB_USER")
	}
	if cfg.Password == "" {
		cfg.Password = os.Getenv("ERUPE_DB_PASSWORD")
	}
	if cfg.DBName == "" {
		cfg.DBName = os.Getenv("ERUPE_DB_NAME")
	}

	// Apply defaults
	if cfg.Host == "" {
		cfg.Host = "localhost"
	}
	if cfg.Port == 0 {
		cfg.Port = 5432
	}
	if cfg.User == "" {
		cfg.User = "postgres"
	}
	if cfg.DBName == "" {
		cfg.DBName = "erupe"
	}

	// Password is required
	if cfg.Password == "" {
		return fmt.Errorf("database password is required (set in config.json, use -password flag, or ERUPE_DB_PASSWORD env var)")
	}

	return nil
}

// connectDB establishes a connection to the PostgreSQL database.
func connectDB(cfg *DBConfig) (*sql.DB, error) {
	if err := resolveDBConfig(cfg); err != nil {
		return nil, err
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
