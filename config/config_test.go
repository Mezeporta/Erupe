package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestConfigStructDefaults(t *testing.T) {
	// Test that Config struct has expected zero values
	c := &Config{}

	if c.DevMode != false {
		t.Error("DevMode default should be false")
	}
	if c.Host != "" {
		t.Error("Host default should be empty")
	}
	if c.HideLoginNotice != false {
		t.Error("HideLoginNotice default should be false")
	}
}

func TestDevModeOptionsDefaults(t *testing.T) {
	d := DevModeOptions{}

	if d.AutoCreateAccount != false {
		t.Error("AutoCreateAccount default should be false")
	}
	if d.CleanDB != false {
		t.Error("CleanDB default should be false")
	}
	if d.MaxLauncherHR != false {
		t.Error("MaxLauncherHR default should be false")
	}
	if d.LogInboundMessages != false {
		t.Error("LogInboundMessages default should be false")
	}
	if d.LogOutboundMessages != false {
		t.Error("LogOutboundMessages default should be false")
	}
	if d.MaxHexdumpLength != 0 {
		t.Error("MaxHexdumpLength default should be 0")
	}
}

func TestGameplayOptionsDefaults(t *testing.T) {
	g := GameplayOptions{}

	if g.FeaturedWeapons != 0 {
		t.Error("FeaturedWeapons default should be 0")
	}
	if g.MaximumNP != 0 {
		t.Error("MaximumNP default should be 0")
	}
	if g.MaximumRP != 0 {
		t.Error("MaximumRP default should be 0")
	}
	if g.DisableLoginBoost != false {
		t.Error("DisableLoginBoost default should be false")
	}
}

func TestLoggingDefaults(t *testing.T) {
	l := Logging{}

	if l.LogToFile != false {
		t.Error("LogToFile default should be false")
	}
	if l.LogFilePath != "" {
		t.Error("LogFilePath default should be empty")
	}
	if l.LogMaxSize != 0 {
		t.Error("LogMaxSize default should be 0")
	}
}

func TestDatabaseStruct(t *testing.T) {
	d := Database{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "secret",
		Database: "erupe",
	}

	if d.Host != "localhost" {
		t.Errorf("Host = %s, want localhost", d.Host)
	}
	if d.Port != 5432 {
		t.Errorf("Port = %d, want 5432", d.Port)
	}
	if d.User != "postgres" {
		t.Errorf("User = %s, want postgres", d.User)
	}
	if d.Password != "secret" {
		t.Errorf("Password = %s, want secret", d.Password)
	}
	if d.Database != "erupe" {
		t.Errorf("Database = %s, want erupe", d.Database)
	}
}

func TestSignStruct(t *testing.T) {
	s := Sign{
		Enabled: true,
		Port:    53312,
	}

	if s.Enabled != true {
		t.Error("Enabled should be true")
	}
	if s.Port != 53312 {
		t.Errorf("Port = %d, want 53312", s.Port)
	}
}

func TestSignV2Struct(t *testing.T) {
	s := SignV2{
		Enabled: true,
		Port:    8080,
	}

	if s.Enabled != true {
		t.Error("Enabled should be true")
	}
	if s.Port != 8080 {
		t.Errorf("Port = %d, want 8080", s.Port)
	}
}

func TestEntranceStruct(t *testing.T) {
	e := Entrance{
		Enabled: true,
		Port:    53310,
		Entries: []EntranceServerInfo{
			{
				IP:   "127.0.0.1",
				Type: 1,
				Name: "Test Server",
			},
		},
	}

	if e.Enabled != true {
		t.Error("Enabled should be true")
	}
	if e.Port != 53310 {
		t.Errorf("Port = %d, want 53310", e.Port)
	}
	if len(e.Entries) != 1 {
		t.Errorf("Entries len = %d, want 1", len(e.Entries))
	}
}

func TestEntranceServerInfoStruct(t *testing.T) {
	info := EntranceServerInfo{
		IP:                 "192.168.1.1",
		Type:               2,
		Season:             1,
		Recommended:        3,
		Name:               "Test World",
		Description:        "A test server",
		AllowedClientFlags: 4096,
		Channels: []EntranceChannelInfo{
			{Port: 54001, MaxPlayers: 100, CurrentPlayers: 50},
		},
	}

	if info.IP != "192.168.1.1" {
		t.Errorf("IP = %s, want 192.168.1.1", info.IP)
	}
	if info.Type != 2 {
		t.Errorf("Type = %d, want 2", info.Type)
	}
	if info.Season != 1 {
		t.Errorf("Season = %d, want 1", info.Season)
	}
	if info.Recommended != 3 {
		t.Errorf("Recommended = %d, want 3", info.Recommended)
	}
	if info.Name != "Test World" {
		t.Errorf("Name = %s, want Test World", info.Name)
	}
	if info.Description != "A test server" {
		t.Errorf("Description = %s, want A test server", info.Description)
	}
	if info.AllowedClientFlags != 4096 {
		t.Errorf("AllowedClientFlags = %d, want 4096", info.AllowedClientFlags)
	}
	if len(info.Channels) != 1 {
		t.Errorf("Channels len = %d, want 1", len(info.Channels))
	}
}

func TestEntranceChannelInfoStruct(t *testing.T) {
	ch := EntranceChannelInfo{
		Port:           54001,
		MaxPlayers:     100,
		CurrentPlayers: 25,
	}

	if ch.Port != 54001 {
		t.Errorf("Port = %d, want 54001", ch.Port)
	}
	if ch.MaxPlayers != 100 {
		t.Errorf("MaxPlayers = %d, want 100", ch.MaxPlayers)
	}
	if ch.CurrentPlayers != 25 {
		t.Errorf("CurrentPlayers = %d, want 25", ch.CurrentPlayers)
	}
}

func TestDiscordStruct(t *testing.T) {
	d := Discord{
		Enabled:           true,
		BotToken:          "test-token",
		RealtimeChannelID: "123456789",
	}

	if d.Enabled != true {
		t.Error("Enabled should be true")
	}
	if d.BotToken != "test-token" {
		t.Errorf("BotToken = %s, want test-token", d.BotToken)
	}
	if d.RealtimeChannelID != "123456789" {
		t.Errorf("RealtimeChannelID = %s, want 123456789", d.RealtimeChannelID)
	}
}

func TestCommandStruct(t *testing.T) {
	cmd := Command{
		Name:    "teleport",
		Enabled: true,
		Prefix:  "!",
	}

	if cmd.Name != "teleport" {
		t.Errorf("Name = %s, want teleport", cmd.Name)
	}
	if cmd.Enabled != true {
		t.Error("Enabled should be true")
	}
	if cmd.Prefix != "!" {
		t.Errorf("Prefix = %s, want !", cmd.Prefix)
	}
}

func TestCourseStruct(t *testing.T) {
	course := Course{
		Name:    "Premium",
		Enabled: true,
	}

	if course.Name != "Premium" {
		t.Errorf("Name = %s, want Premium", course.Name)
	}
	if course.Enabled != true {
		t.Error("Enabled should be true")
	}
}

func TestSaveDumpOptionsStruct(t *testing.T) {
	s := SaveDumpOptions{
		Enabled:   true,
		OutputDir: "/tmp/dumps",
	}

	if s.Enabled != true {
		t.Error("Enabled should be true")
	}
	if s.OutputDir != "/tmp/dumps" {
		t.Errorf("OutputDir = %s, want /tmp/dumps", s.OutputDir)
	}
}

func TestIsTestMode(t *testing.T) {
	// When running tests, isTestMode should return true
	if !isTestMode() {
		t.Error("isTestMode() should return true when running tests")
	}
}

func TestLoadConfigMissingFile(t *testing.T) {
	// Create a temporary directory without a config file
	tmpDir, err := os.MkdirTemp("", "erupe-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Save current directory and change to temp
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(origDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// LoadConfig should fail without config.json
	_, err = LoadConfig()
	if err == nil {
		t.Error("LoadConfig() should return error when config file is missing")
	}
}

func TestLoadConfigValidFile(t *testing.T) {
	// Create a temporary directory with a valid config file
	tmpDir, err := os.MkdirTemp("", "erupe-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create minimal config.json
	configContent := `{
		"Host": "127.0.0.1",
		"DevMode": true,
		"Database": {
			"Host": "localhost",
			"Port": 5432,
			"User": "postgres",
			"Password": "password",
			"Database": "erupe"
		}
	}`

	configPath := filepath.Join(tmpDir, "config.json")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Save current directory and change to temp
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(origDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	// LoadConfig should succeed
	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	if cfg.Host != "127.0.0.1" {
		t.Errorf("Host = %s, want 127.0.0.1", cfg.Host)
	}
	if cfg.DevMode != true {
		t.Error("DevMode should be true")
	}
	if cfg.Database.Host != "localhost" {
		t.Errorf("Database.Host = %s, want localhost", cfg.Database.Host)
	}
}

func TestLoadConfigDefaults(t *testing.T) {
	// Create a temporary directory with minimal config
	tmpDir, err := os.MkdirTemp("", "erupe-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create minimal config.json (just enough to pass)
	configContent := `{
		"Host": "192.168.1.1"
	}`

	configPath := filepath.Join(tmpDir, "config.json")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Save current directory and change to temp
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(origDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	cfg, err := LoadConfig()
	if err != nil {
		t.Fatalf("LoadConfig() error = %v", err)
	}

	// Check Logging defaults are applied
	if cfg.Logging.LogToFile != true {
		t.Error("Logging.LogToFile should default to true")
	}
	if cfg.Logging.LogFilePath != "logs/erupe.log" {
		t.Errorf("Logging.LogFilePath = %s, want logs/erupe.log", cfg.Logging.LogFilePath)
	}
	if cfg.Logging.LogMaxSize != 100 {
		t.Errorf("Logging.LogMaxSize = %d, want 100", cfg.Logging.LogMaxSize)
	}
	if cfg.Logging.LogMaxBackups != 3 {
		t.Errorf("Logging.LogMaxBackups = %d, want 3", cfg.Logging.LogMaxBackups)
	}
	if cfg.Logging.LogMaxAge != 28 {
		t.Errorf("Logging.LogMaxAge = %d, want 28", cfg.Logging.LogMaxAge)
	}
	if cfg.Logging.LogCompress != true {
		t.Error("Logging.LogCompress should default to true")
	}

	// Check SaveDumps defaults
	if cfg.DevModeOptions.SaveDumps.Enabled != false {
		t.Error("SaveDumps.Enabled should default to false")
	}
	if cfg.DevModeOptions.SaveDumps.OutputDir != "savedata" {
		t.Errorf("SaveDumps.OutputDir = %s, want savedata", cfg.DevModeOptions.SaveDumps.OutputDir)
	}
}

func TestLoadConfigInvalidJSON(t *testing.T) {
	// Create a temporary directory with invalid JSON config
	tmpDir, err := os.MkdirTemp("", "erupe-config-test-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create invalid JSON
	configContent := `{ this is not valid json }`

	configPath := filepath.Join(tmpDir, "config.json")
	err = os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write config file: %v", err)
	}

	// Save current directory and change to temp
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current dir: %v", err)
	}
	defer os.Chdir(origDir)

	err = os.Chdir(tmpDir)
	if err != nil {
		t.Fatalf("Failed to change to temp dir: %v", err)
	}

	_, err = LoadConfig()
	if err == nil {
		t.Error("LoadConfig() should return error for invalid JSON")
	}
}

func TestChannelStruct(t *testing.T) {
	ch := Channel{
		Enabled: true,
	}

	if ch.Enabled != true {
		t.Error("Enabled should be true")
	}
}

func TestConfigCompleteStructure(t *testing.T) {
	// Test building a complete config structure
	cfg := &Config{
		Host:                "192.168.1.100",
		BinPath:             "/bin",
		Language:            "JP",
		DisableSoftCrash:    true,
		HideLoginNotice:     false,
		LoginNotices:        []string{"Notice 1", "Notice 2"},
		PatchServerManifest: "http://patch.example.com/manifest",
		PatchServerFile:     "http://patch.example.com/files",
		DevMode:             true,
		DevModeOptions: DevModeOptions{
			AutoCreateAccount:   true,
			CleanDB:             false,
			MaxLauncherHR:       true,
			LogInboundMessages:  true,
			LogOutboundMessages: true,
			MaxHexdumpLength:    256,
		},
		GameplayOptions: GameplayOptions{
			FeaturedWeapons:     5,
			MaximumNP:           99999,
			MaximumRP:           65535,
			DisableLoginBoost:   false,
			BoostTimeDuration:   60,
			GuildMealDuration:   30,
			BonusQuestAllowance: 10,
			DailyQuestAllowance: 5,
		},
		Database: Database{
			Host:     "db.example.com",
			Port:     5432,
			User:     "erupe",
			Password: "secret",
			Database: "erupe_db",
		},
		Sign: Sign{
			Enabled: true,
			Port:    53312,
		},
		SignV2: SignV2{
			Enabled: true,
			Port:    8080,
		},
		Channel: Channel{
			Enabled: true,
		},
		Entrance: Entrance{
			Enabled: true,
			Port:    53310,
		},
	}

	// Verify values are set correctly
	if cfg.Host != "192.168.1.100" {
		t.Errorf("Host = %s, want 192.168.1.100", cfg.Host)
	}
	if cfg.GameplayOptions.MaximumNP != 99999 {
		t.Errorf("MaximumNP = %d, want 99999", cfg.GameplayOptions.MaximumNP)
	}
	if len(cfg.LoginNotices) != 2 {
		t.Errorf("LoginNotices len = %d, want 2", len(cfg.LoginNotices))
	}
}
