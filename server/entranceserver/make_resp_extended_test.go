package entranceserver

import (
	"testing"

	cfg "erupe-ce/config"

	"go.uber.org/zap"
)

// TestMakeHeader tests the makeHeader function with various inputs
func TestMakeHeader(t *testing.T) {
	tests := []struct {
		name       string
		data       []byte
		respType   string
		entryCount uint16
		key        byte
	}{
		{"empty data", []byte{}, "SV2", 0, 0x00},
		{"small data", []byte{0x01, 0x02, 0x03}, "SV2", 1, 0x00},
		{"SVR type", []byte{0xAA, 0xBB}, "SVR", 2, 0x42},
		{"USR type", []byte{0x01}, "USR", 1, 0x00},
		{"larger data", make([]byte, 100), "SV2", 5, 0xFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := makeHeader(tt.data, tt.respType, tt.entryCount, tt.key)
			if len(result) == 0 {
				t.Error("makeHeader returned empty result")
			}
			// First byte should be the key
			if result[0] != tt.key {
				t.Errorf("first byte = %x, want %x", result[0], tt.key)
			}
		})
	}
}

// helper to create a standard test config with one server entry and one channel
func newTestConfig(mode cfg.Mode) *cfg.Config {
	return &cfg.Config{
		RealClientMode: mode,
		Host:           "127.0.0.1",
		Entrance: cfg.Entrance{
			Enabled: true,
			Port:    53310,
			Entries: []cfg.EntranceServerInfo{
				{
					Name:               "TestServer",
					Description:        "Test",
					IP:                 "127.0.0.1",
					Type:               0,
					Recommended:        1,
					AllowedClientFlags: 0xFFFFFFFF,
					Channels: []cfg.EntranceChannelInfo{
						{Port: 54001, MaxPlayers: 100},
					},
				},
			},
		},
		GameplayOptions: cfg.GameplayOptions{
			ClanMemberLimits: [][]uint8{{1, 60}},
		},
	}
}

func newTestServer(config *cfg.Config) *Server {
	return &Server{
		logger:      zap.NewNop(),
		erupeConfig: config,
		serverRepo:  &mockEntranceServerRepo{currentPlayers: 10},
	}
}

// --- makeSv2Resp tests ---

func TestMakeSv2Resp_ZZ(t *testing.T) {
	config := newTestConfig(cfg.ZZ)
	s := newTestServer(config)

	result := makeSv2Resp(config, s, true)
	if len(result) == 0 {
		t.Fatal("makeSv2Resp returned empty result")
	}
	// First byte is key (0x00), then encrypted data containing "SV2"
	if result[0] != 0x00 {
		t.Errorf("first byte = %x, want 0x00", result[0])
	}
}

func TestMakeSv2Resp_G32_SVR(t *testing.T) {
	config := newTestConfig(cfg.G32)
	s := newTestServer(config)

	result := makeSv2Resp(config, s, true)
	if len(result) == 0 {
		t.Fatal("makeSv2Resp returned empty result for G3.2")
	}
	// Decrypt the response to verify it contains "SVR" instead of "SV2"
	decrypted := DecryptBin8(result[1:], result[0])
	if len(decrypted) < 3 {
		t.Fatalf("decrypted response too short: %d bytes", len(decrypted))
	}
	respType := string(decrypted[:3])
	if respType != "SVR" {
		t.Errorf("respType = %q, want SVR", respType)
	}
}

func TestMakeSv2Resp_Z1_FiltersMezFes(t *testing.T) {
	config := newTestConfig(cfg.Z1)
	// Add a MezFes entry (Type=6)
	config.Entrance.Entries = append(config.Entrance.Entries, cfg.EntranceServerInfo{
		Name:        "MezFes",
		Description: "MezFes World",
		IP:          "127.0.0.1",
		Type:        6,
		Channels:    []cfg.EntranceChannelInfo{{Port: 54002, MaxPlayers: 100}},
	})
	s := newTestServer(config)

	result := makeSv2Resp(config, s, true)
	if len(result) == 0 {
		t.Fatal("makeSv2Resp returned empty for Z1 + MezFes")
	}
	// Decrypt and check server count (should be 1, not 2)
	decrypted := DecryptBin8(result[1:], result[0])
	// Header: 3 bytes type + 2 bytes count + 2 bytes length
	serverCount := uint16(decrypted[3])<<8 | uint16(decrypted[4])
	if serverCount != 1 {
		t.Errorf("server count = %d, want 1 (MezFes should be filtered)", serverCount)
	}
}

func TestMakeSv2Resp_G6_FiltersReturn(t *testing.T) {
	config := newTestConfig(cfg.G6)
	// Add a Return entry (Type=5)
	config.Entrance.Entries = append(config.Entrance.Entries, cfg.EntranceServerInfo{
		Name:        "Return",
		Description: "Return World",
		IP:          "127.0.0.1",
		Type:        5,
		Channels:    []cfg.EntranceChannelInfo{{Port: 54002, MaxPlayers: 100}},
	})
	s := newTestServer(config)

	result := makeSv2Resp(config, s, true)
	if len(result) == 0 {
		t.Fatal("makeSv2Resp returned empty for G6 + Return")
	}
	// Decrypt and check server count
	decrypted := DecryptBin8(result[1:], result[0])
	serverCount := uint16(decrypted[3])<<8 | uint16(decrypted[4])
	if serverCount != 1 {
		t.Errorf("server count = %d, want 1 (Return should be filtered)", serverCount)
	}
}

func TestMakeSv2Resp_WithDebugLogging(t *testing.T) {
	config := newTestConfig(cfg.ZZ)
	config.DebugOptions.LogOutboundMessages = true
	s := newTestServer(config)

	result := makeSv2Resp(config, s, true)
	if len(result) == 0 {
		t.Fatal("makeSv2Resp with debug logging returned empty")
	}
}

// --- encodeServerInfo branch tests ---

func TestEncodeServerInfo_NonLocalIP(t *testing.T) {
	config := newTestConfig(cfg.ZZ)
	config.Entrance.Entries[0].IP = "192.168.1.100"
	s := newTestServer(config)

	localResult := encodeServerInfo(config, s, true)
	nonLocalResult := encodeServerInfo(config, s, false)

	if len(localResult) == 0 || len(nonLocalResult) == 0 {
		t.Fatal("encodeServerInfo returned empty result")
	}
	// First 4 bytes are the IP — they should differ between local and non-local
	if bytesEqual(localResult[:4], nonLocalResult[:4]) {
		t.Error("local and non-local should encode different IPs")
	}
	// Local IP is written as WriteUint32(0x0100007F), which is big-endian: 01 00 00 7F
	if localResult[0] != 0x01 || localResult[1] != 0x00 || localResult[2] != 0x00 || localResult[3] != 0x7F {
		t.Errorf("local IP bytes = %02x%02x%02x%02x, want 0100007F", localResult[0], localResult[1], localResult[2], localResult[3])
	}
}

func TestEncodeServerInfo_EmptyIP(t *testing.T) {
	config := newTestConfig(cfg.ZZ)
	config.Host = "10.0.0.1"
	config.Entrance.Entries[0].IP = "" // should fall back to config.Host
	s := newTestServer(config)

	result := encodeServerInfo(config, s, false)
	if len(result) == 0 {
		t.Fatal("encodeServerInfo returned empty for empty IP")
	}
	// 10.0.0.1 → LittleEndian.Uint32 → 0x0100000A, written big-endian: 01 00 00 0A
	if result[0] != 0x01 || result[1] != 0x00 || result[2] != 0x00 || result[3] != 0x0A {
		t.Errorf("fallback IP bytes = %02x%02x%02x%02x, want 0100000A", result[0], result[1], result[2], result[3])
	}
}

func TestEncodeServerInfo_GGClientMode(t *testing.T) {
	config := newTestConfig(cfg.GG)
	config.Entrance.Entries[0].AllowedClientFlags = 0xDEADBEEF
	s := newTestServer(config)

	result := encodeServerInfo(config, s, true)
	if len(result) == 0 {
		t.Fatal("encodeServerInfo returned empty for GG mode")
	}
}

func TestEncodeServerInfo_G1toG5ClientMode(t *testing.T) {
	config := newTestConfig(cfg.G3)
	s := newTestServer(config)

	result := encodeServerInfo(config, s, true)
	if len(result) == 0 {
		t.Fatal("encodeServerInfo returned empty for G3 mode")
	}
}

func TestEncodeServerInfo_ProxyPort(t *testing.T) {
	config := newTestConfig(cfg.ZZ)
	config.DebugOptions.ProxyPort = 9999
	s := newTestServer(config)

	result := encodeServerInfo(config, s, true)
	if len(result) == 0 {
		t.Fatal("encodeServerInfo returned empty for ProxyPort config")
	}
}

func TestEncodeServerInfo_Z1_SkipsMezFesType(t *testing.T) {
	configWithMezFes := newTestConfig(cfg.Z1)
	configWithMezFes.Entrance.Entries = append(configWithMezFes.Entrance.Entries, cfg.EntranceServerInfo{
		Name:        "MezFes",
		Description: "MezFes World",
		IP:          "127.0.0.1",
		Type:        6,
		Channels:    []cfg.EntranceChannelInfo{{Port: 54002, MaxPlayers: 100}},
	})

	configWithout := newTestConfig(cfg.Z1)

	sWith := newTestServer(configWithMezFes)
	sWithout := newTestServer(configWithout)

	resultWith := encodeServerInfo(configWithMezFes, sWith, true)
	resultWithout := encodeServerInfo(configWithout, sWithout, true)

	// Both should produce identical output since MezFes Type=6 is skipped on Z1
	if !bytesEqual(resultWith, resultWithout) {
		t.Errorf("MezFes entry should be skipped on Z1, but output differs: %d vs %d bytes",
			len(resultWith), len(resultWithout))
	}
}

func TestEncodeServerInfo_G6_SkipsReturnType(t *testing.T) {
	configWithReturn := newTestConfig(cfg.G6)
	configWithReturn.Entrance.Entries = append(configWithReturn.Entrance.Entries, cfg.EntranceServerInfo{
		Name:        "Return",
		Description: "Return World",
		IP:          "127.0.0.1",
		Type:        5,
		Channels:    []cfg.EntranceChannelInfo{{Port: 54002, MaxPlayers: 100}},
	})

	configWithout := newTestConfig(cfg.G6)

	sWith := newTestServer(configWithReturn)
	sWithout := newTestServer(configWithout)

	resultWith := encodeServerInfo(configWithReturn, sWith, true)
	resultWithout := encodeServerInfo(configWithout, sWithout, true)

	if !bytesEqual(resultWith, resultWithout) {
		t.Errorf("Return entry should be skipped on G6, but output differs: %d vs %d bytes",
			len(resultWith), len(resultWithout))
	}
}

// --- makeUsrResp debug logging test ---

func TestMakeUsrResp_WithDebugLogging(t *testing.T) {
	config := &cfg.Config{
		RealClientMode: cfg.ZZ,
		DebugOptions: cfg.DebugOptions{
			LogOutboundMessages: true,
		},
	}
	s := &Server{
		logger:      zap.NewNop(),
		erupeConfig: config,
		sessionRepo: &mockEntranceSessionRepo{serverID: 1234},
	}

	pkt := []byte{
		'A', 'L', 'L', '+',
		0x00,
		0x00, 0x01, // 1 entry
		0x00, 0x00, 0x00, 0x01, // char_id = 1
	}

	result := makeUsrResp(pkt, s)
	if len(result) == 0 {
		t.Error("makeUsrResp with debug logging returned empty")
	}
}
