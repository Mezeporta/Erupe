package main

import (
	"database/sql"
	"flag"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// escapeConnStringValue
// ---------------------------------------------------------------------------

func TestEscapeConnStringValue_Empty(t *testing.T) {
	got := escapeConnStringValue("")
	if got != "" {
		t.Errorf("expected empty string, got %q", got)
	}
}

func TestEscapeConnStringValue_NoSpecialChars(t *testing.T) {
	got := escapeConnStringValue("hello world")
	if got != "hello world" {
		t.Errorf("expected %q, got %q", "hello world", got)
	}
}

func TestEscapeConnStringValue_SingleQuote(t *testing.T) {
	got := escapeConnStringValue("it's")
	want := "it''s"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEscapeConnStringValue_Backslash(t *testing.T) {
	got := escapeConnStringValue(`path\to\file`)
	want := `path\\to\\file`
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEscapeConnStringValue_BothQuoteAndBackslash(t *testing.T) {
	got := escapeConnStringValue(`it's\path`)
	want := `it''s\\path`
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEscapeConnStringValue_MultipleConsecutiveQuotes(t *testing.T) {
	got := escapeConnStringValue("a'''b")
	want := "a''''''b"
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEscapeConnStringValue_OnlySpecialChars(t *testing.T) {
	got := escapeConnStringValue(`'\`)
	want := `''\\`
	if got != want {
		t.Errorf("expected %q, got %q", want, got)
	}
}

func TestEscapeConnStringValue_Table(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  string
	}{
		{"unicode", "p@$$w\u00f6rd", "p@$$w\u00f6rd"},
		{"spaces only", "   ", "   "},
		{"leading quote", "'start", "''start"},
		{"trailing quote", "end'", "end''"},
		{"trailing backslash", `end\`, `end\\`},
		{"multiple backslashes", `a\\b`, `a\\\\b`},
		{"mixed complex", `x'y\z'w\`, `x''y\\z''w\\`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := escapeConnStringValue(tt.input)
			if got != tt.want {
				t.Errorf("escapeConnStringValue(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// loadConfigFile
// ---------------------------------------------------------------------------

func TestLoadConfigFile_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{
		"Database": {
			"Host": "myhost",
			"Port": 1234,
			"User": "myuser",
			"Password": "mypass",
			"Database": "mydb"
		}
	}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfigFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Host != "myhost" {
		t.Errorf("Host = %q, want %q", cfg.Database.Host, "myhost")
	}
	if cfg.Database.Port != 1234 {
		t.Errorf("Port = %d, want %d", cfg.Database.Port, 1234)
	}
	if cfg.Database.User != "myuser" {
		t.Errorf("User = %q, want %q", cfg.Database.User, "myuser")
	}
	if cfg.Database.Password != "mypass" {
		t.Errorf("Password = %q, want %q", cfg.Database.Password, "mypass")
	}
	if cfg.Database.Database != "mydb" {
		t.Errorf("Database = %q, want %q", cfg.Database.Database, "mydb")
	}
}

func TestLoadConfigFile_NonExistent(t *testing.T) {
	_, err := loadConfigFile("/tmp/nonexistent_config_test_12345.json")
	if err == nil {
		t.Fatal("expected error for non-existent file, got nil")
	}
}

func TestLoadConfigFile_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte("not valid json {{{"), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadConfigFile(path)
	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
}

func TestLoadConfigFile_EmptyFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(""), 0644); err != nil {
		t.Fatal(err)
	}

	_, err := loadConfigFile(path)
	if err == nil {
		t.Fatal("expected error for empty file, got nil")
	}
}

func TestLoadConfigFile_NoDatabaseField(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	if err := os.WriteFile(path, []byte(`{"SomeOther": "field"}`), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfigFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Database fields should be zero-values
	if cfg.Database.Host != "" {
		t.Errorf("expected empty Host, got %q", cfg.Database.Host)
	}
	if cfg.Database.Port != 0 {
		t.Errorf("expected Port 0, got %d", cfg.Database.Port)
	}
	if cfg.Database.Password != "" {
		t.Errorf("expected empty Password, got %q", cfg.Database.Password)
	}
}

func TestLoadConfigFile_PartialDatabase(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	content := `{"Database": {"Host": "partial", "Port": 9999}}`
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg, err := loadConfigFile(path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Database.Host != "partial" {
		t.Errorf("Host = %q, want %q", cfg.Database.Host, "partial")
	}
	if cfg.Database.Port != 9999 {
		t.Errorf("Port = %d, want %d", cfg.Database.Port, 9999)
	}
	if cfg.Database.User != "" {
		t.Errorf("User = %q, want empty", cfg.Database.User)
	}
}

// ---------------------------------------------------------------------------
// findConfigFile
// ---------------------------------------------------------------------------

func TestFindConfigFile_NotFound(t *testing.T) {
	// Run in a temp directory where no config.json exists.
	// We save and restore the working directory so other tests are not affected.
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	dir := t.TempDir()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	result := findConfigFile()
	if result != "" {
		t.Errorf("expected empty string when no config.json exists, got %q", result)
	}
}

func TestFindConfigFile_InCurrentDir(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	result := findConfigFile()
	if result == "" {
		t.Error("expected findConfigFile to find config.json in current directory")
	}
}

func TestFindConfigFile_TwoLevelsUp(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	// Simulate tools/usercheck/ structure: config.json is ../../config.json
	dir := t.TempDir()
	subDir := filepath.Join(dir, "tools", "usercheck")
	if err := os.MkdirAll(subDir, 0755); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(dir, "config.json")
	if err := os.WriteFile(configPath, []byte(`{}`), 0644); err != nil {
		t.Fatal(err)
	}

	if err := os.Chdir(subDir); err != nil {
		t.Fatal(err)
	}

	result := findConfigFile()
	if result == "" {
		t.Error("expected findConfigFile to find config.json two levels up")
	}
}

// ---------------------------------------------------------------------------
// addDBFlags
// ---------------------------------------------------------------------------

func TestAddDBFlags_RegistersAllFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := &DBConfig{}
	addDBFlags(fs, cfg)

	expectedFlags := []string{"config", "host", "port", "user", "password", "dbname"}
	for _, name := range expectedFlags {
		f := fs.Lookup(name)
		if f == nil {
			t.Errorf("expected flag %q to be registered, but it was not found", name)
		}
	}
}

func TestAddDBFlags_ParseFlags(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := &DBConfig{}
	addDBFlags(fs, cfg)

	args := []string{
		"-host", "dbhost",
		"-port", "6543",
		"-user", "dbuser",
		"-password", "dbpass",
		"-dbname", "testdb",
		"-config", "/some/path.json",
	}
	if err := fs.Parse(args); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	if cfg.Host != "dbhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "dbhost")
	}
	if cfg.Port != 6543 {
		t.Errorf("Port = %d, want %d", cfg.Port, 6543)
	}
	if cfg.User != "dbuser" {
		t.Errorf("User = %q, want %q", cfg.User, "dbuser")
	}
	if cfg.Password != "dbpass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "dbpass")
	}
	if cfg.DBName != "testdb" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "testdb")
	}
	if cfg.ConfigPath != "/some/path.json" {
		t.Errorf("ConfigPath = %q, want %q", cfg.ConfigPath, "/some/path.json")
	}
}

func TestAddDBFlags_DefaultValues(t *testing.T) {
	fs := flag.NewFlagSet("test", flag.ContinueOnError)
	cfg := &DBConfig{}
	addDBFlags(fs, cfg)

	// Parse with no arguments
	if err := fs.Parse(nil); err != nil {
		t.Fatalf("unexpected parse error: %v", err)
	}

	// All fields should be zero values (defaults from flag package)
	if cfg.Host != "" {
		t.Errorf("Host = %q, want empty", cfg.Host)
	}
	if cfg.Port != 0 {
		t.Errorf("Port = %d, want 0", cfg.Port)
	}
	if cfg.User != "" {
		t.Errorf("User = %q, want empty", cfg.User)
	}
	if cfg.Password != "" {
		t.Errorf("Password = %q, want empty", cfg.Password)
	}
}

// ---------------------------------------------------------------------------
// resolveDBConfig
// ---------------------------------------------------------------------------

func TestResolveDBConfig_AllPreset(t *testing.T) {
	// Clear environment variables that could interfere
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_USER")
	os.Unsetenv("ERUPE_DB_PASSWORD")
	os.Unsetenv("ERUPE_DB_NAME")

	cfg := &DBConfig{
		Host:     "myhost",
		Port:     5555,
		User:     "myuser",
		Password: "mypass",
		DBName:   "mydb",
	}

	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "myhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "myhost")
	}
	if cfg.Port != 5555 {
		t.Errorf("Port = %d, want %d", cfg.Port, 5555)
	}
	if cfg.User != "myuser" {
		t.Errorf("User = %q, want %q", cfg.User, "myuser")
	}
	if cfg.Password != "mypass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "mypass")
	}
	if cfg.DBName != "mydb" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "mydb")
	}
}

func TestResolveDBConfig_MissingPassword(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_USER")
	os.Unsetenv("ERUPE_DB_PASSWORD")
	os.Unsetenv("ERUPE_DB_NAME")

	// Change to a temp dir so no config.json is found
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	cfg := &DBConfig{}
	err = resolveDBConfig(cfg)
	if err == nil {
		t.Fatal("expected error for missing password, got nil")
	}
	if got := err.Error(); got != "database password is required (set in config.json, use -password flag, or ERUPE_DB_PASSWORD env var)" {
		t.Errorf("unexpected error message: %q", got)
	}
}

func TestResolveDBConfig_PasswordFromEnv(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_USER")
	os.Unsetenv("ERUPE_DB_NAME")

	os.Setenv("ERUPE_DB_PASSWORD", "envpass")

	// Change to a temp dir so no config.json is found
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	cfg := &DBConfig{}
	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Password != "envpass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "envpass")
	}
}

func TestResolveDBConfig_HostFromEnv(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_USER")
	os.Unsetenv("ERUPE_DB_NAME")

	os.Setenv("ERUPE_DB_HOST", "envhost")
	os.Setenv("ERUPE_DB_PASSWORD", "envpass")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	cfg := &DBConfig{}
	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "envhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "envhost")
	}
}

func TestResolveDBConfig_UserFromEnv(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_NAME")

	os.Setenv("ERUPE_DB_USER", "envuser")
	os.Setenv("ERUPE_DB_PASSWORD", "envpass")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	cfg := &DBConfig{}
	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.User != "envuser" {
		t.Errorf("User = %q, want %q", cfg.User, "envuser")
	}
}

func TestResolveDBConfig_DBNameFromEnv(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_USER")

	os.Setenv("ERUPE_DB_PASSWORD", "envpass")
	os.Setenv("ERUPE_DB_NAME", "envdb")

	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	cfg := &DBConfig{}
	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.DBName != "envdb" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "envdb")
	}
}

func TestResolveDBConfig_DefaultsApplied(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_USER")
	os.Unsetenv("ERUPE_DB_PASSWORD")
	os.Unsetenv("ERUPE_DB_NAME")

	// Change to a temp dir so no config.json is found
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatal(err)
	}

	cfg := &DBConfig{Password: "provided"}
	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "localhost")
	}
	if cfg.Port != 5432 {
		t.Errorf("Port = %d, want %d", cfg.Port, 5432)
	}
	if cfg.User != "postgres" {
		t.Errorf("User = %q, want %q", cfg.User, "postgres")
	}
	if cfg.DBName != "erupe" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "erupe")
	}
}

func TestResolveDBConfig_ConfigFileOverrides(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_USER")
	os.Unsetenv("ERUPE_DB_PASSWORD")
	os.Unsetenv("ERUPE_DB_NAME")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := `{
		"Database": {
			"Host": "filehost",
			"Port": 7777,
			"User": "fileuser",
			"Password": "filepass",
			"Database": "filedb"
		}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	cfg := &DBConfig{ConfigPath: configPath}
	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "filehost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "filehost")
	}
	if cfg.Port != 7777 {
		t.Errorf("Port = %d, want %d", cfg.Port, 7777)
	}
	if cfg.User != "fileuser" {
		t.Errorf("User = %q, want %q", cfg.User, "fileuser")
	}
	if cfg.Password != "filepass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "filepass")
	}
	if cfg.DBName != "filedb" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "filedb")
	}
}

func TestResolveDBConfig_FlagsOverrideConfigFile(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_USER")
	os.Unsetenv("ERUPE_DB_PASSWORD")
	os.Unsetenv("ERUPE_DB_NAME")

	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := `{
		"Database": {
			"Host": "filehost",
			"Port": 7777,
			"User": "fileuser",
			"Password": "filepass",
			"Database": "filedb"
		}
	}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// CLI flags take priority since they're already set in cfg
	cfg := &DBConfig{
		ConfigPath: configPath,
		Host:       "clihost",
		Port:       8888,
		User:       "cliuser",
		Password:   "clipass",
		DBName:     "clidb",
	}
	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if cfg.Host != "clihost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "clihost")
	}
	if cfg.Port != 8888 {
		t.Errorf("Port = %d, want %d", cfg.Port, 8888)
	}
	if cfg.User != "cliuser" {
		t.Errorf("User = %q, want %q", cfg.User, "cliuser")
	}
	if cfg.Password != "clipass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "clipass")
	}
	if cfg.DBName != "clidb" {
		t.Errorf("DBName = %q, want %q", cfg.DBName, "clidb")
	}
}

func TestResolveDBConfig_ExplicitConfigPathInvalid(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_USER")
	os.Unsetenv("ERUPE_DB_PASSWORD")
	os.Unsetenv("ERUPE_DB_NAME")

	cfg := &DBConfig{
		ConfigPath: "/nonexistent/path/config.json",
		Password:   "pass",
	}
	err := resolveDBConfig(cfg)
	if err == nil {
		t.Fatal("expected error when explicitly specifying non-existent config path")
	}
}

func TestResolveDBConfig_AutoDetectedConfigPathInvalid(t *testing.T) {
	// When config.json is found by findConfigFile but is invalid JSON,
	// resolveDBConfig should silently ignore it (because user didn't explicitly specify it)
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_USER")
	os.Unsetenv("ERUPE_DB_PASSWORD")
	os.Unsetenv("ERUPE_DB_NAME")

	dir := t.TempDir()
	// Write a broken config.json
	if err := os.WriteFile(filepath.Join(dir, "config.json"), []byte("BROKEN"), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}

	cfg := &DBConfig{Password: "pass"}
	// Should not error -- broken auto-detected config is silently ignored
	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Defaults should be applied
	if cfg.Host != "localhost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "localhost")
	}
	if cfg.Port != 5432 {
		t.Errorf("Port = %d, want %d", cfg.Port, 5432)
	}
}

func TestResolveDBConfig_EnvOverridesConfig(t *testing.T) {
	origHost := os.Getenv("ERUPE_DB_HOST")
	origUser := os.Getenv("ERUPE_DB_USER")
	origPass := os.Getenv("ERUPE_DB_PASSWORD")
	origName := os.Getenv("ERUPE_DB_NAME")
	defer func() {
		os.Setenv("ERUPE_DB_HOST", origHost)
		os.Setenv("ERUPE_DB_USER", origUser)
		os.Setenv("ERUPE_DB_PASSWORD", origPass)
		os.Setenv("ERUPE_DB_NAME", origName)
	}()

	// Config file provides some values, env provides others
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config.json")
	content := `{"Database": {"Host": "filehost", "Port": 7777}}`
	if err := os.WriteFile(configPath, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}

	// Env provides password and user, which config file doesn't set
	os.Setenv("ERUPE_DB_PASSWORD", "envpass")
	os.Setenv("ERUPE_DB_USER", "envuser")
	os.Unsetenv("ERUPE_DB_HOST")
	os.Unsetenv("ERUPE_DB_NAME")

	cfg := &DBConfig{ConfigPath: configPath}
	if err := resolveDBConfig(cfg); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Host comes from config file
	if cfg.Host != "filehost" {
		t.Errorf("Host = %q, want %q", cfg.Host, "filehost")
	}
	// User comes from env (config file didn't set it)
	if cfg.User != "envuser" {
		t.Errorf("User = %q, want %q", cfg.User, "envuser")
	}
	// Password comes from env
	if cfg.Password != "envpass" {
		t.Errorf("Password = %q, want %q", cfg.Password, "envpass")
	}
}

// ---------------------------------------------------------------------------
// Struct types construction
// ---------------------------------------------------------------------------

func TestConnectedUser_Construction(t *testing.T) {
	now := time.Now()
	u := ConnectedUser{
		CharID:     42,
		CharName:   "Hunter",
		ServerID:   1,
		ServerName: "World1",
		UserID:     10,
		Username:   "player1",
		LastLogin:  sql.NullTime{Time: now, Valid: true},
		HR:         999,
		GR:         50,
	}

	if u.CharID != 42 {
		t.Errorf("CharID = %d, want 42", u.CharID)
	}
	if u.CharName != "Hunter" {
		t.Errorf("CharName = %q, want %q", u.CharName, "Hunter")
	}
	if u.ServerID != 1 {
		t.Errorf("ServerID = %d, want 1", u.ServerID)
	}
	if u.ServerName != "World1" {
		t.Errorf("ServerName = %q, want %q", u.ServerName, "World1")
	}
	if u.UserID != 10 {
		t.Errorf("UserID = %d, want 10", u.UserID)
	}
	if u.Username != "player1" {
		t.Errorf("Username = %q, want %q", u.Username, "player1")
	}
	if !u.LastLogin.Valid {
		t.Error("LastLogin.Valid = false, want true")
	}
	if u.HR != 999 {
		t.Errorf("HR = %d, want 999", u.HR)
	}
	if u.GR != 50 {
		t.Errorf("GR = %d, want 50", u.GR)
	}
}

func TestConnectedUser_NullLastLogin(t *testing.T) {
	u := ConnectedUser{
		CharID:   1,
		CharName: "Test",
	}
	if u.LastLogin.Valid {
		t.Error("LastLogin.Valid = true, want false for zero value")
	}
}

func TestServerStatus_Construction(t *testing.T) {
	s := ServerStatus{
		ServerID:       5,
		WorldName:      "Frontier",
		WorldDesc:      "A great server",
		Land:           2,
		CurrentPlayers: 100,
		Season:         1,
	}

	if s.ServerID != 5 {
		t.Errorf("ServerID = %d, want 5", s.ServerID)
	}
	if s.WorldName != "Frontier" {
		t.Errorf("WorldName = %q, want %q", s.WorldName, "Frontier")
	}
	if s.WorldDesc != "A great server" {
		t.Errorf("WorldDesc = %q, want %q", s.WorldDesc, "A great server")
	}
	if s.Land != 2 {
		t.Errorf("Land = %d, want 2", s.Land)
	}
	if s.CurrentPlayers != 100 {
		t.Errorf("CurrentPlayers = %d, want 100", s.CurrentPlayers)
	}
	if s.Season != 1 {
		t.Errorf("Season = %d, want 1", s.Season)
	}
}

func TestLoginHistory_Construction(t *testing.T) {
	now := time.Now()
	h := LoginHistory{
		CharID:    7,
		CharName:  "Veteran",
		LastLogin: sql.NullTime{Time: now, Valid: true},
		HR:        500,
		GR:        25,
		Username:  "vet_player",
	}

	if h.CharID != 7 {
		t.Errorf("CharID = %d, want 7", h.CharID)
	}
	if h.CharName != "Veteran" {
		t.Errorf("CharName = %q, want %q", h.CharName, "Veteran")
	}
	if !h.LastLogin.Valid {
		t.Error("LastLogin.Valid = false, want true")
	}
	if h.HR != 500 {
		t.Errorf("HR = %d, want 500", h.HR)
	}
	if h.GR != 25 {
		t.Errorf("GR = %d, want 25", h.GR)
	}
	if h.Username != "vet_player" {
		t.Errorf("Username = %q, want %q", h.Username, "vet_player")
	}
}

func TestLoginHistory_NullLastLogin(t *testing.T) {
	h := LoginHistory{
		CharID:   1,
		CharName: "NewPlayer",
	}
	if h.LastLogin.Valid {
		t.Error("LastLogin.Valid = true, want false for zero value")
	}
}

// ---------------------------------------------------------------------------
// DBConfig and ErupeConfig struct construction
// ---------------------------------------------------------------------------

func TestDBConfig_ZeroValue(t *testing.T) {
	cfg := DBConfig{}
	if cfg.Host != "" {
		t.Errorf("Host = %q, want empty", cfg.Host)
	}
	if cfg.Port != 0 {
		t.Errorf("Port = %d, want 0", cfg.Port)
	}
	if cfg.User != "" {
		t.Errorf("User = %q, want empty", cfg.User)
	}
	if cfg.Password != "" {
		t.Errorf("Password = %q, want empty", cfg.Password)
	}
	if cfg.DBName != "" {
		t.Errorf("DBName = %q, want empty", cfg.DBName)
	}
	if cfg.ConfigPath != "" {
		t.Errorf("ConfigPath = %q, want empty", cfg.ConfigPath)
	}
}

func TestErupeConfig_ZeroValue(t *testing.T) {
	cfg := ErupeConfig{}
	if cfg.Database.Host != "" {
		t.Errorf("Database.Host = %q, want empty", cfg.Database.Host)
	}
	if cfg.Database.Port != 0 {
		t.Errorf("Database.Port = %d, want 0", cfg.Database.Port)
	}
}
