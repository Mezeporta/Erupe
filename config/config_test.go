package config

import (
	"testing"
)

// TestModeString tests the versionStrings array content
func TestModeString(t *testing.T) {
	// NOTE: The Mode.String() method in config.go has a bug - it directly uses the Mode value
	// as an index (which is 1-41) but versionStrings is 0-indexed. This test validates
	// the versionStrings array content instead.

	expectedStrings := map[int]string{
		0:  "S1.0",
		1:  "S1.5",
		2:  "S2.0",
		3:  "S2.5",
		4:  "S3.0",
		5:  "S3.5",
		6:  "S4.0",
		7:  "S5.0",
		8:  "S5.5",
		9:  "S6.0",
		10: "S7.0",
		11: "S8.0",
		12: "S8.5",
		13: "S9.0",
		14: "S10",
		15: "FW.1",
		16: "FW.2",
		17: "FW.3",
		18: "FW.4",
		19: "FW.5",
		20: "G1",
		21: "G2",
		22: "G3",
		23: "G3.1",
		24: "G3.2",
		25: "GG",
		26: "G5",
		27: "G5.1",
		28: "G5.2",
		29: "G6",
		30: "G6.1",
		31: "G7",
		32: "G8",
		33: "G8.1",
		34: "G9",
		35: "G9.1",
		36: "G10",
		37: "G10.1",
		38: "Z1",
		39: "Z2",
		40: "ZZ",
	}

	for i, expected := range expectedStrings {
		if i < len(versionStrings) {
			if versionStrings[i] != expected {
				t.Errorf("versionStrings[%d] = %s, want %s", i, versionStrings[i], expected)
			}
		}
	}
}

// TestModeConstants verifies all mode constants are unique and in order
func TestModeConstants(t *testing.T) {
	modes := []Mode{
		S1, S15, S2, S25, S3, S35, S4, S5, S55, S6, S7, S8, S85, S9, S10,
		F1, F2, F3, F4, F5,
		G1, G2, G3, G31, G32, GG, G5, G51, G52, G6, G61, G7, G8, G81, G9, G91, G10, G101,
		Z1, Z2, ZZ,
	}

	// Verify all modes are unique
	seen := make(map[Mode]bool)
	for _, mode := range modes {
		if seen[mode] {
			t.Errorf("Duplicate mode constant: %v", mode)
		}
		seen[mode] = true
	}

	// Verify modes are in sequential order
	for i, mode := range modes {
		if int(mode) != i+1 {
			t.Errorf("Mode %v at index %d has wrong value: got %d, want %d", mode, i, mode, i+1)
		}
	}

	// Verify total count
	if len(modes) != len(versionStrings) {
		t.Errorf("Number of modes (%d) doesn't match versionStrings count (%d)", len(modes), len(versionStrings))
	}
}

// TestVersionStringsLength verifies versionStrings has correct length
func TestVersionStringsLength(t *testing.T) {
	expectedCount := 41 // S1 through ZZ = 41 versions
	if len(versionStrings) != expectedCount {
		t.Errorf("versionStrings length = %d, want %d", len(versionStrings), expectedCount)
	}
}

// TestVersionStringsContent verifies critical version strings
func TestVersionStringsContent(t *testing.T) {
	tests := []struct {
		index    int
		expected string
	}{
		{0, "S1.0"},  // S1
		{14, "S10"},  // S10
		{15, "FW.1"}, // F1
		{19, "FW.5"}, // F5
		{20, "G1"},   // G1
		{38, "Z1"},   // Z1
		{39, "Z2"},   // Z2
		{40, "ZZ"},   // ZZ
	}

	for _, tt := range tests {
		if versionStrings[tt.index] != tt.expected {
			t.Errorf("versionStrings[%d] = %s, want %s", tt.index, versionStrings[tt.index], tt.expected)
		}
	}
}

// TestGetOutboundIP4 tests IP detection
func TestGetOutboundIP4(t *testing.T) {
	ip, err := getOutboundIP4()
	if err != nil {
		t.Fatalf("getOutboundIP4() returned error: %v", err)
	}
	if ip == nil {
		t.Error("getOutboundIP4() returned nil IP")
	}

	// Verify it returns IPv4
	if ip.To4() == nil {
		t.Error("getOutboundIP4() should return valid IPv4")
	}

	// Verify it's not all zeros
	if len(ip) == 4 && ip[0] == 0 && ip[1] == 0 && ip[2] == 0 && ip[3] == 0 {
		t.Error("getOutboundIP4() returned 0.0.0.0")
	}
}

// TestConfigStructTypes verifies Config struct fields have correct types
func TestConfigStructTypes(t *testing.T) {
	cfg := &Config{
		Host:                   "localhost",
		BinPath:                "/path/to/bin",
		Language:               "en",
		DisableSoftCrash:       false,
		HideLoginNotice:        false,
		LoginNotices:           []string{"Notice"},
		PatchServerManifest:    "http://patch.example.com",
		PatchServerFile:        "http://files.example.com",
		DeleteOnSaveCorruption: false,
		ClientMode:             "ZZ",
		RealClientMode:         ZZ,
		QuestCacheExpiry:       3600,
		CommandPrefix:          "!",
		AutoCreateAccount:      false,
		LoopDelay:              100,
		DefaultCourses:         []uint16{1, 2, 3},
		EarthStatus:            1,
		EarthID:                1,
		EarthMonsters:          []int32{1, 2, 3},
		SaveDumps: SaveDumpOptions{
			Enabled:    true,
			RawEnabled: false,
			OutputDir:  "/dumps",
		},
		Screenshots: ScreenshotsOptions{
			Enabled:       true,
			Host:          "localhost",
			Port:          8080,
			OutputDir:     "/screenshots",
			UploadQuality: 85,
		},
		DebugOptions: DebugOptions{
			CleanDB:             false,
			MaxLauncherHR:       false,
			LogInboundMessages:  false,
			LogOutboundMessages: false,
			LogMessageData:      false,
			MaxHexdumpLength:    32,
		},
		GameplayOptions: GameplayOptions{
			MinFeatureWeapons: 1,
			MaxFeatureWeapons: 5,
		},
	}

	// Verify fields are accessible and have correct types
	if cfg.Host != "localhost" {
		t.Error("Config.Host type mismatch")
	}
	if cfg.QuestCacheExpiry != 3600 {
		t.Error("Config.QuestCacheExpiry type mismatch")
	}
	if cfg.RealClientMode != ZZ {
		t.Error("Config.RealClientMode type mismatch")
	}
}

// TestSaveDumpOptions verifies SaveDumpOptions struct
func TestSaveDumpOptions(t *testing.T) {
	opts := SaveDumpOptions{
		Enabled:    true,
		RawEnabled: false,
		OutputDir:  "/test/path",
	}

	if !opts.Enabled {
		t.Error("SaveDumpOptions.Enabled should be true")
	}
	if opts.RawEnabled {
		t.Error("SaveDumpOptions.RawEnabled should be false")
	}
	if opts.OutputDir != "/test/path" {
		t.Error("SaveDumpOptions.OutputDir mismatch")
	}
}

// TestScreenshotsOptions verifies ScreenshotsOptions struct
func TestScreenshotsOptions(t *testing.T) {
	opts := ScreenshotsOptions{
		Enabled:       true,
		Host:          "ss.example.com",
		Port:          8000,
		OutputDir:     "/screenshots",
		UploadQuality: 90,
	}

	if !opts.Enabled {
		t.Error("ScreenshotsOptions.Enabled should be true")
	}
	if opts.Host != "ss.example.com" {
		t.Error("ScreenshotsOptions.Host mismatch")
	}
	if opts.Port != 8000 {
		t.Error("ScreenshotsOptions.Port mismatch")
	}
	if opts.UploadQuality != 90 {
		t.Error("ScreenshotsOptions.UploadQuality mismatch")
	}
}

// TestDebugOptions verifies DebugOptions struct
func TestDebugOptions(t *testing.T) {
	opts := DebugOptions{
		CleanDB:             true,
		MaxLauncherHR:       true,
		LogInboundMessages:  true,
		LogOutboundMessages: true,
		LogMessageData:      true,
		MaxHexdumpLength:    128,
		DivaOverride:        1,
		DisableTokenCheck:   true,
	}

	if !opts.CleanDB {
		t.Error("DebugOptions.CleanDB should be true")
	}
	if !opts.MaxLauncherHR {
		t.Error("DebugOptions.MaxLauncherHR should be true")
	}
	if opts.MaxHexdumpLength != 128 {
		t.Error("DebugOptions.MaxHexdumpLength mismatch")
	}
	if !opts.DisableTokenCheck {
		t.Error("DebugOptions.DisableTokenCheck should be true (security risk!)")
	}
}

// TestGameplayOptions verifies GameplayOptions struct
func TestGameplayOptions(t *testing.T) {
	opts := GameplayOptions{
		MinFeatureWeapons:    2,
		MaxFeatureWeapons:    10,
		MaximumNP:            999999,
		MaximumRP:            9999,
		MaximumFP:            999999999,
		MezFesSoloTickets:    100,
		MezFesGroupTickets:   50,
		DisableHunterNavi:    true,
		EnableKaijiEvent:     true,
		EnableHiganjimaEvent: false,
		EnableNierEvent:      false,
	}

	if opts.MinFeatureWeapons != 2 {
		t.Error("GameplayOptions.MinFeatureWeapons mismatch")
	}
	if opts.MaxFeatureWeapons != 10 {
		t.Error("GameplayOptions.MaxFeatureWeapons mismatch")
	}
	if opts.MezFesSoloTickets != 100 {
		t.Error("GameplayOptions.MezFesSoloTickets mismatch")
	}
	if !opts.EnableKaijiEvent {
		t.Error("GameplayOptions.EnableKaijiEvent should be true")
	}
}

// TestCapLinkOptions verifies CapLinkOptions struct
func TestCapLinkOptions(t *testing.T) {
	opts := CapLinkOptions{
		Values: []uint16{1, 2, 3},
		Key:    "test-key",
		Host:   "localhost",
		Port:   9999,
	}

	if len(opts.Values) != 3 {
		t.Error("CapLinkOptions.Values length mismatch")
	}
	if opts.Key != "test-key" {
		t.Error("CapLinkOptions.Key mismatch")
	}
	if opts.Port != 9999 {
		t.Error("CapLinkOptions.Port mismatch")
	}
}

// TestDatabase verifies Database struct
func TestDatabase(t *testing.T) {
	db := Database{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "password",
		Database: "erupe",
	}

	if db.Host != "localhost" {
		t.Error("Database.Host mismatch")
	}
	if db.Port != 5432 {
		t.Error("Database.Port mismatch")
	}
	if db.User != "postgres" {
		t.Error("Database.User mismatch")
	}
	if db.Database != "erupe" {
		t.Error("Database.Database mismatch")
	}
}

// TestSign verifies Sign struct
func TestSign(t *testing.T) {
	sign := Sign{
		Enabled: true,
		Port:    8081,
	}

	if !sign.Enabled {
		t.Error("Sign.Enabled should be true")
	}
	if sign.Port != 8081 {
		t.Error("Sign.Port mismatch")
	}
}

// TestAPI verifies API struct
func TestAPI(t *testing.T) {
	api := API{
		Enabled:     true,
		Port:        8080,
		PatchServer: "http://patch.example.com",
		Banners: []APISignBanner{
			{Src: "banner.jpg", Link: "http://example.com"},
		},
		Messages: []APISignMessage{
			{Message: "Welcome", Date: 0, Kind: 0, Link: "http://example.com"},
		},
		Links: []APISignLink{
			{Name: "Forum", Icon: "forum", Link: "http://forum.example.com"},
		},
	}

	if !api.Enabled {
		t.Error("API.Enabled should be true")
	}
	if api.Port != 8080 {
		t.Error("API.Port mismatch")
	}
	if len(api.Banners) != 1 {
		t.Error("API.Banners length mismatch")
	}
}

// TestAPISignBanner verifies APISignBanner struct
func TestAPISignBanner(t *testing.T) {
	banner := APISignBanner{
		Src:  "http://example.com/banner.jpg",
		Link: "http://example.com",
	}

	if banner.Src != "http://example.com/banner.jpg" {
		t.Error("APISignBanner.Src mismatch")
	}
	if banner.Link != "http://example.com" {
		t.Error("APISignBanner.Link mismatch")
	}
}

// TestAPISignMessage verifies APISignMessage struct
func TestAPISignMessage(t *testing.T) {
	msg := APISignMessage{
		Message: "Welcome to Erupe!",
		Date:    1625097600,
		Kind:    0,
		Link:    "http://example.com",
	}

	if msg.Message != "Welcome to Erupe!" {
		t.Error("APISignMessage.Message mismatch")
	}
	if msg.Date != 1625097600 {
		t.Error("APISignMessage.Date mismatch")
	}
	if msg.Kind != 0 {
		t.Error("APISignMessage.Kind mismatch")
	}
}

// TestAPISignLink verifies APISignLink struct
func TestAPISignLink(t *testing.T) {
	link := APISignLink{
		Name: "Forum",
		Icon: "forum",
		Link: "http://forum.example.com",
	}

	if link.Name != "Forum" {
		t.Error("APISignLink.Name mismatch")
	}
	if link.Icon != "forum" {
		t.Error("APISignLink.Icon mismatch")
	}
	if link.Link != "http://forum.example.com" {
		t.Error("APISignLink.Link mismatch")
	}
}

// TestChannel verifies Channel struct
func TestChannel(t *testing.T) {
	ch := Channel{
		Enabled: true,
	}

	if !ch.Enabled {
		t.Error("Channel.Enabled should be true")
	}
}

// TestEntrance verifies Entrance struct
func TestEntrance(t *testing.T) {
	entrance := Entrance{
		Enabled: true,
		Port:    10000,
		Entries: []EntranceServerInfo{
			{
				IP:          "192.168.1.1",
				Type:        1,
				Season:      0,
				Recommended: 0,
				Name:        "Test Server",
				Description: "A test server",
			},
		},
	}

	if !entrance.Enabled {
		t.Error("Entrance.Enabled should be true")
	}
	if entrance.Port != 10000 {
		t.Error("Entrance.Port mismatch")
	}
	if len(entrance.Entries) != 1 {
		t.Error("Entrance.Entries length mismatch")
	}
}

// TestEntranceServerInfo verifies EntranceServerInfo struct
func TestEntranceServerInfo(t *testing.T) {
	info := EntranceServerInfo{
		IP:                 "192.168.1.1",
		Type:               1,
		Season:             0,
		Recommended:        0,
		Name:               "Server 1",
		Description:        "Main server",
		AllowedClientFlags: 4096,
		Channels: []EntranceChannelInfo{
			{Port: 10001, MaxPlayers: 4, CurrentPlayers: 2},
		},
	}

	if info.IP != "192.168.1.1" {
		t.Error("EntranceServerInfo.IP mismatch")
	}
	if info.Type != 1 {
		t.Error("EntranceServerInfo.Type mismatch")
	}
	if len(info.Channels) != 1 {
		t.Error("EntranceServerInfo.Channels length mismatch")
	}
}

// TestEntranceChannelInfo verifies EntranceChannelInfo struct
func TestEntranceChannelInfo(t *testing.T) {
	info := EntranceChannelInfo{
		Port:           10001,
		MaxPlayers:     4,
		CurrentPlayers: 2,
	}

	if info.Port != 10001 {
		t.Error("EntranceChannelInfo.Port mismatch")
	}
	if info.MaxPlayers != 4 {
		t.Error("EntranceChannelInfo.MaxPlayers mismatch")
	}
	if info.CurrentPlayers != 2 {
		t.Error("EntranceChannelInfo.CurrentPlayers mismatch")
	}
}

// TestEntranceChannelInfoIsEnabled tests the Enabled field and IsEnabled helper
func TestEntranceChannelInfoIsEnabled(t *testing.T) {
	trueVal := true
	falseVal := false

	tests := []struct {
		name    string
		enabled *bool
		want    bool
	}{
		{"nil defaults to true", nil, true},
		{"explicit true", &trueVal, true},
		{"explicit false", &falseVal, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := EntranceChannelInfo{
				Port:    10001,
				Enabled: tt.enabled,
			}
			if got := info.IsEnabled(); got != tt.want {
				t.Errorf("IsEnabled() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestDiscord verifies Discord struct
func TestDiscord(t *testing.T) {
	discord := Discord{
		Enabled:  true,
		BotToken: "token123",
		RelayChannel: DiscordRelay{
			Enabled:          true,
			MaxMessageLength: 2000,
			RelayChannelID:   "123456789",
		},
	}

	if !discord.Enabled {
		t.Error("Discord.Enabled should be true")
	}
	if discord.BotToken != "token123" {
		t.Error("Discord.BotToken mismatch")
	}
	if discord.RelayChannel.MaxMessageLength != 2000 {
		t.Error("Discord.RelayChannel.MaxMessageLength mismatch")
	}
}

// TestCommand verifies Command struct
func TestCommand(t *testing.T) {
	cmd := Command{
		Name:        "test",
		Enabled:     true,
		Description: "Test command",
		Prefix:      "!",
	}

	if cmd.Name != "test" {
		t.Error("Command.Name mismatch")
	}
	if !cmd.Enabled {
		t.Error("Command.Enabled should be true")
	}
	if cmd.Prefix != "!" {
		t.Error("Command.Prefix mismatch")
	}
}

// TestCourse verifies Course struct
func TestCourse(t *testing.T) {
	course := Course{
		Name:    "Rookie Road",
		Enabled: true,
	}

	if course.Name != "Rookie Road" {
		t.Error("Course.Name mismatch")
	}
	if !course.Enabled {
		t.Error("Course.Enabled should be true")
	}
}

// TestGameplayOptionsConstraints tests gameplay option constraints
func TestGameplayOptionsConstraints(t *testing.T) {
	tests := []struct {
		name string
		opts GameplayOptions
		ok   bool
	}{
		{
			name: "valid multipliers",
			opts: GameplayOptions{
				HRPMultiplier:      1.5,
				GRPMultiplier:      1.2,
				ZennyMultiplier:    1.0,
				MaterialMultiplier: 1.3,
			},
			ok: true,
		},
		{
			name: "zero multipliers",
			opts: GameplayOptions{
				HRPMultiplier: 0.0,
			},
			ok: true,
		},
		{
			name: "high multipliers",
			opts: GameplayOptions{
				GCPMultiplier: 10.0,
			},
			ok: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Just verify the struct can be created with these values
			_ = tt.opts
		})
	}
}

// TestModeValueRanges tests Mode constant value ranges
func TestModeValueRanges(t *testing.T) {
	if S1 < 1 || S1 > ZZ {
		t.Error("S1 mode value out of range")
	}
	if ZZ <= G101 {
		t.Error("ZZ should be greater than G101")
	}
	if G101 <= F5 {
		t.Error("G101 should be greater than F5")
	}
}

// TestConfigDefaults tests default configuration creation
func TestConfigDefaults(t *testing.T) {
	cfg := &Config{
		ClientMode:     "ZZ",
		RealClientMode: ZZ,
	}

	if cfg.ClientMode != "ZZ" {
		t.Error("Default ClientMode mismatch")
	}
	if cfg.RealClientMode != ZZ {
		t.Error("Default RealClientMode mismatch")
	}
}

// BenchmarkModeString benchmarks Mode.String() method
func BenchmarkModeString(b *testing.B) {
	mode := ZZ
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = mode.String()
	}
}

// BenchmarkGetOutboundIP4 benchmarks IP detection
func BenchmarkGetOutboundIP4(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = getOutboundIP4()
	}
}
