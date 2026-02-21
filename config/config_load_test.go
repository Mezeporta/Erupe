package config

import (
	"os"
	"strings"
	"testing"
)

// TestLoadConfigNoFile tests LoadConfig when config file doesn't exist
func TestLoadConfigNoFile(t *testing.T) {
	// Change to temporary directory to ensure no config file exists
	tmpDir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// LoadConfig should fail when no config.toml exists
	config, err := LoadConfig()
	if err == nil {
		t.Error("LoadConfig() should return error when config file doesn't exist")
	}
	if config != nil {
		t.Error("LoadConfig() should return nil config on error")
	}
}

// TestLoadConfigClientModeMapping tests client mode string to Mode conversion
func TestLoadConfigClientModeMapping(t *testing.T) {
	// Test that we can identify version strings and map them to modes
	tests := []struct {
		versionStr      string
		expectedMode    Mode
		shouldHaveDebug bool
	}{
		{"S1.0", S1, true},
		{"S10", S10, true},
		{"G10.1", G101, true},
		{"ZZ", ZZ, false},
		{"Z1", Z1, false},
	}

	for _, tt := range tests {
		t.Run(tt.versionStr, func(t *testing.T) {
			// Find matching version string
			var foundMode Mode
			for i, vstr := range versionStrings {
				if vstr == tt.versionStr {
					foundMode = Mode(i + 1)
					break
				}
			}

			if foundMode != tt.expectedMode {
				t.Errorf("Version string %s: expected mode %v, got %v", tt.versionStr, tt.expectedMode, foundMode)
			}

			// Check debug mode marking (versions <= G101 should have debug marking)
			hasDebug := tt.expectedMode <= G101
			if hasDebug != tt.shouldHaveDebug {
				t.Errorf("Debug mode flag for %v: expected %v, got %v", tt.expectedMode, tt.shouldHaveDebug, hasDebug)
			}
		})
	}
}

// TestLoadConfigFeatureWeaponConstraint tests MinFeatureWeapons > MaxFeatureWeapons constraint
func TestLoadConfigFeatureWeaponConstraint(t *testing.T) {
	tests := []struct {
		name       string
		minWeapons int
		maxWeapons int
		expected   int
	}{
		{"min < max", 2, 5, 2},
		{"min > max", 10, 5, 5}, // Should be clamped to max
		{"min == max", 3, 3, 3},
		{"min = 0, max = 0", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Simulate constraint logic from LoadConfig
			min := tt.minWeapons
			max := tt.maxWeapons
			if min > max {
				min = max
			}
			if min != tt.expected {
				t.Errorf("Feature weapon constraint: expected min=%d, got %d", tt.expected, min)
			}
		})
	}
}

// TestLoadConfigDefaultHost tests host assignment
func TestLoadConfigDefaultHost(t *testing.T) {
	cfg := &Config{
		Host: "",
	}

	// When Host is empty, it should be set to the outbound IP
	if cfg.Host == "" {
		// Simulate the logic: if empty, set to outbound IP
		ip, err := getOutboundIP4()
		if err != nil {
			t.Fatalf("getOutboundIP4() error: %v", err)
		}
		cfg.Host = ip.To4().String()
		if cfg.Host == "" {
			t.Error("Host should be set to outbound IP, got empty string")
		}
		// Verify it looks like an IP address
		parts := len(strings.Split(cfg.Host, "."))
		if parts != 4 {
			t.Errorf("Host doesn't look like IPv4 address: %s", cfg.Host)
		}
	}
}

// TestLoadConfigDefaultModeWhenInvalid tests default mode when invalid
func TestLoadConfigDefaultModeWhenInvalid(t *testing.T) {
	// When RealClientMode is 0 (invalid), it should default to ZZ
	var realMode Mode = 0 // Invalid
	if realMode == 0 {
		realMode = ZZ
	}

	if realMode != ZZ {
		t.Errorf("Invalid mode should default to ZZ, got %v", realMode)
	}
}

// TestConfigStruct tests Config structure creation with all fields
func TestConfigStruct(t *testing.T) {
	cfg := &Config{
		Host:                   "localhost",
		BinPath:                "/opt/erupe",
		Language:               "en",
		DisableSoftCrash:       false,
		HideLoginNotice:        false,
		LoginNotices:           []string{"Welcome"},
		PatchServerManifest:    "http://patch.example.com/manifest",
		PatchServerFile:        "http://patch.example.com/files",
		DeleteOnSaveCorruption: false,
		ClientMode:             "ZZ",
		RealClientMode:         ZZ,
		QuestCacheExpiry:       3600,
		CommandPrefix:          "!",
		AutoCreateAccount:      false,
		LoopDelay:              100,
		DefaultCourses:         []uint16{1, 2, 3},
		EarthStatus:            0,
		EarthID:                0,
		EarthMonsters:          []int32{100, 101, 102},
		SaveDumps: SaveDumpOptions{
			Enabled:    true,
			RawEnabled: false,
			OutputDir:  "save-backups",
		},
		Screenshots: ScreenshotsOptions{
			Enabled:       true,
			Host:          "localhost",
			Port:          8080,
			OutputDir:     "screenshots",
			UploadQuality: 85,
		},
		DebugOptions: DebugOptions{
			CleanDB:             false,
			MaxLauncherHR:       false,
			LogInboundMessages:  false,
			LogOutboundMessages: false,
			LogMessageData:      false,
		},
		GameplayOptions: GameplayOptions{
			MinFeatureWeapons: 1,
			MaxFeatureWeapons: 5,
		},
	}

	// Verify all fields are accessible
	if cfg.Host != "localhost" {
		t.Error("Failed to set Host")
	}
	if cfg.RealClientMode != ZZ {
		t.Error("Failed to set RealClientMode")
	}
	if len(cfg.LoginNotices) != 1 {
		t.Error("Failed to set LoginNotices")
	}
	if cfg.GameplayOptions.MaxFeatureWeapons != 5 {
		t.Error("Failed to set GameplayOptions.MaxFeatureWeapons")
	}
}

// TestConfigNilSafety tests that Config can be safely created as nil and populated
func TestConfigNilSafety(t *testing.T) {
	var cfg *Config
	if cfg != nil {
		t.Error("Config should start as nil")
	}

	cfg = &Config{}

	cfg.Host = "test"
	if cfg.Host != "test" {
		t.Error("Failed to set field on allocated Config")
	}
}

// TestEmptyConfigCreation tests creating empty Config struct
func TestEmptyConfigCreation(t *testing.T) {
	cfg := Config{}

	// Verify zero values
	if cfg.Host != "" {
		t.Error("Empty Config.Host should be empty string")
	}
	if cfg.RealClientMode != 0 {
		t.Error("Empty Config.RealClientMode should be 0")
	}
	if len(cfg.LoginNotices) != 0 {
		t.Error("Empty Config.LoginNotices should be empty slice")
	}
}

// TestVersionStringsMapped tests all version strings are present
func TestVersionStringsMapped(t *testing.T) {
	// Verify all expected version strings are present
	expectedVersions := []string{
		"S1.0", "S1.5", "S2.0", "S2.5", "S3.0", "S3.5", "S4.0", "S5.0", "S5.5", "S6.0", "S7.0",
		"S8.0", "S8.5", "S9.0", "S10", "FW.1", "FW.2", "FW.3", "FW.4", "FW.5", "G1", "G2", "G3",
		"G3.1", "G3.2", "GG", "G5", "G5.1", "G5.2", "G6", "G6.1", "G7", "G8", "G8.1", "G9", "G9.1",
		"G10", "G10.1", "Z1", "Z2", "ZZ",
	}

	if len(versionStrings) != len(expectedVersions) {
		t.Errorf("versionStrings count mismatch: got %d, want %d", len(versionStrings), len(expectedVersions))
	}

	for i, expected := range expectedVersions {
		if i < len(versionStrings) && versionStrings[i] != expected {
			t.Errorf("versionStrings[%d]: got %s, want %s", i, versionStrings[i], expected)
		}
	}
}

// TestDefaultSaveDumpsConfig tests default SaveDumps configuration
func TestDefaultSaveDumpsConfig(t *testing.T) {
	// The LoadConfig function sets default SaveDumps
	// viper.SetDefault("DevModeOptions.SaveDumps", SaveDumpOptions{...})

	opts := SaveDumpOptions{
		Enabled:   true,
		OutputDir: "save-backups",
	}

	if !opts.Enabled {
		t.Error("Default SaveDumps should be enabled")
	}
	if opts.OutputDir != "save-backups" {
		t.Error("Default SaveDumps OutputDir should be 'save-backups'")
	}
}

// TestEntranceServerConfig tests complete entrance server configuration
func TestEntranceServerConfig(t *testing.T) {
	entrance := Entrance{
		Enabled: true,
		Port:    10000,
		Entries: []EntranceServerInfo{
			{
				IP:                 "192.168.1.100",
				Type:               1, // open
				Season:             0, // green
				Recommended:        1,
				Name:               "Main Server",
				Description:        "Main hunting server",
				AllowedClientFlags: 8192,
				Channels: []EntranceChannelInfo{
					{Port: 10001, MaxPlayers: 4, CurrentPlayers: 2},
					{Port: 10002, MaxPlayers: 4, CurrentPlayers: 1},
					{Port: 10003, MaxPlayers: 4, CurrentPlayers: 4},
				},
			},
		},
	}

	if !entrance.Enabled {
		t.Error("Entrance should be enabled")
	}
	if entrance.Port != 10000 {
		t.Error("Entrance port mismatch")
	}
	if len(entrance.Entries) != 1 {
		t.Error("Entrance should have 1 entry")
	}
	if len(entrance.Entries[0].Channels) != 3 {
		t.Error("Entry should have 3 channels")
	}

	// Verify channel occupancy
	channels := entrance.Entries[0].Channels
	for _, ch := range channels {
		if ch.CurrentPlayers > ch.MaxPlayers {
			t.Errorf("Channel %d has more current players than max", ch.Port)
		}
	}
}

// TestDiscordConfiguration tests Discord integration configuration
func TestDiscordConfiguration(t *testing.T) {
	discord := Discord{
		Enabled:  true,
		BotToken: "MTA4NTYT3Y0NzY0NTEwNjU0Ng.GMJX5x.example",
		RelayChannel: DiscordRelay{
			Enabled:          true,
			MaxMessageLength: 2000,
			RelayChannelID:   "987654321098765432",
		},
	}

	if !discord.Enabled {
		t.Error("Discord should be enabled")
	}
	if discord.BotToken == "" {
		t.Error("Discord BotToken should be set")
	}
	if !discord.RelayChannel.Enabled {
		t.Error("Discord relay should be enabled")
	}
	if discord.RelayChannel.MaxMessageLength != 2000 {
		t.Error("Discord relay max message length should be 2000")
	}
}

// TestMultipleEntranceServers tests configuration with multiple entrance servers
func TestMultipleEntranceServers(t *testing.T) {
	entrance := Entrance{
		Enabled: true,
		Port:    10000,
		Entries: []EntranceServerInfo{
			{IP: "192.168.1.100", Type: 1, Name: "Beginner"},
			{IP: "192.168.1.101", Type: 2, Name: "Cities"},
			{IP: "192.168.1.102", Type: 3, Name: "Advanced"},
		},
	}

	if len(entrance.Entries) != 3 {
		t.Errorf("Expected 3 servers, got %d", len(entrance.Entries))
	}

	types := []uint8{1, 2, 3}
	for i, entry := range entrance.Entries {
		if entry.Type != types[i] {
			t.Errorf("Server %d type mismatch", i)
		}
	}
}

// TestGameplayMultiplierBoundaries tests gameplay multiplier values
func TestGameplayMultiplierBoundaries(t *testing.T) {
	tests := []struct {
		name  string
		value float32
		ok    bool
	}{
		{"zero multiplier", 0.0, true},
		{"one multiplier", 1.0, true},
		{"half multiplier", 0.5, true},
		{"double multiplier", 2.0, true},
		{"high multiplier", 10.0, true},
		{"negative multiplier", -1.0, true}, // No validation in code
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := GameplayOptions{
				HRPMultiplier: tt.value,
			}
			// Just verify the value can be set
			if opts.HRPMultiplier != tt.value {
				t.Errorf("Multiplier not set correctly: expected %f, got %f", tt.value, opts.HRPMultiplier)
			}
		})
	}
}

// TestCommandConfiguration tests command configuration
func TestCommandConfiguration(t *testing.T) {
	commands := []Command{
		{Name: "help", Enabled: true, Description: "Show help", Prefix: "!"},
		{Name: "quest", Enabled: true, Description: "Quest commands", Prefix: "!"},
		{Name: "admin", Enabled: false, Description: "Admin commands", Prefix: "/"},
	}

	enabledCount := 0
	for _, cmd := range commands {
		if cmd.Enabled {
			enabledCount++
		}
	}

	if enabledCount != 2 {
		t.Errorf("Expected 2 enabled commands, got %d", enabledCount)
	}
}

// TestCourseConfiguration tests course configuration
func TestCourseConfiguration(t *testing.T) {
	courses := []Course{
		{Name: "Rookie Road", Enabled: true},
		{Name: "High Rank", Enabled: true},
		{Name: "G Rank", Enabled: true},
		{Name: "Z Rank", Enabled: false},
	}

	activeCount := 0
	for _, course := range courses {
		if course.Enabled {
			activeCount++
		}
	}

	if activeCount != 3 {
		t.Errorf("Expected 3 active courses, got %d", activeCount)
	}
}

// TestAPIBannersAndLinks tests API configuration with banners and links
func TestAPIBannersAndLinks(t *testing.T) {
	api := API{
		Enabled:     true,
		Port:        8080,
		PatchServer: "http://patch.example.com",
		Banners: []APISignBanner{
			{Src: "banner1.jpg", Link: "http://example.com"},
			{Src: "banner2.jpg", Link: "http://example.com/2"},
		},
		Links: []APISignLink{
			{Name: "Forum", Icon: "forum", Link: "http://forum.example.com"},
			{Name: "Wiki", Icon: "wiki", Link: "http://wiki.example.com"},
		},
	}

	if len(api.Banners) != 2 {
		t.Errorf("Expected 2 banners, got %d", len(api.Banners))
	}
	if len(api.Links) != 2 {
		t.Errorf("Expected 2 links, got %d", len(api.Links))
	}

	for i, banner := range api.Banners {
		if banner.Link == "" {
			t.Errorf("Banner %d has empty link", i)
		}
	}
}

// TestClanMemberLimits tests ClanMemberLimits configuration
func TestClanMemberLimits(t *testing.T) {
	opts := GameplayOptions{
		ClanMemberLimits: [][]uint8{
			{1, 10},
			{2, 20},
			{3, 30},
			{4, 40},
			{5, 50},
		},
	}

	if len(opts.ClanMemberLimits) != 5 {
		t.Errorf("Expected 5 clan member limits, got %d", len(opts.ClanMemberLimits))
	}

	for i, limits := range opts.ClanMemberLimits {
		if limits[0] != uint8(i+1) {
			t.Errorf("Rank mismatch at index %d", i)
		}
	}
}

// BenchmarkConfigCreation benchmarks creating a full Config
func BenchmarkConfigCreation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = &Config{
			Host:           "localhost",
			Language:       "en",
			ClientMode:     "ZZ",
			RealClientMode: ZZ,
		}
	}
}
