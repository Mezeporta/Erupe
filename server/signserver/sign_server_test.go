package signserver

import (
	"fmt"
	"testing"
)

func TestRespIDConstants(t *testing.T) {
	tests := []struct {
		respID RespID
		value  uint16
	}{
		{SIGN_UNKNOWN, 0},
		{SIGN_SUCCESS, 1},
		{SIGN_EFAILED, 2},
		{SIGN_EILLEGAL, 3},
		{SIGN_EALERT, 4},
		{SIGN_EABORT, 5},
		{SIGN_ERESPONSE, 6},
		{SIGN_EDATABASE, 7},
		{SIGN_EABSENCE, 8},
		{SIGN_ERESIGN, 9},
		{SIGN_ESUSPEND_D, 10},
		{SIGN_ELOCK, 11},
		{SIGN_EPASS, 12},
		{SIGN_ERIGHT, 13},
		{SIGN_EAUTH, 14},
		{SIGN_ESUSPEND, 15},
		{SIGN_EELIMINATE, 16},
		{SIGN_ECLOSE, 17},
		{SIGN_ECLOSE_EX, 18},
		{SIGN_EINTERVAL, 19},
		{SIGN_EMOVED, 20},
		{SIGN_ENOTREADY, 21},
		{SIGN_EALREADY, 22},
		{SIGN_EIPADDR, 23},
		{SIGN_EHANGAME, 24},
		{SIGN_UPD_ONLY, 25},
		{SIGN_EMBID, 26},
		{SIGN_ECOGCODE, 27},
		{SIGN_ETOKEN, 28},
		{SIGN_ECOGLINK, 29},
		{SIGN_EMAINTE, 30},
		{SIGN_EMAINTE_NOUPDATE, 31},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("RespID_%d", tt.value), func(t *testing.T) {
			if uint16(tt.respID) != tt.value {
				t.Errorf("RespID = %d, want %d", uint16(tt.respID), tt.value)
			}
		})
	}
}

func TestRespIDType(t *testing.T) {
	// Verify RespID is based on uint16
	var r RespID = 0xFFFF
	if uint16(r) != 0xFFFF {
		t.Errorf("RespID max value = %d, want %d", uint16(r), 0xFFFF)
	}
}

func TestMakeSignInFailureResp(t *testing.T) {
	tests := []RespID{
		SIGN_UNKNOWN,
		SIGN_EFAILED,
		SIGN_EILLEGAL,
		SIGN_ESUSPEND,
		SIGN_EELIMINATE,
		SIGN_EIPADDR,
	}

	for _, respID := range tests {
		t.Run(fmt.Sprintf("RespID_%d", respID), func(t *testing.T) {
			resp := makeSignInFailureResp(respID)

			if len(resp) != 1 {
				t.Errorf("makeSignInFailureResp() len = %d, want 1", len(resp))
			}
			if resp[0] != uint8(respID) {
				t.Errorf("makeSignInFailureResp() = %d, want %d", resp[0], uint8(respID))
			}
		})
	}
}

func TestMakeSignInFailureRespAllCodes(t *testing.T) {
	// Test all possible RespID values 0-39
	for i := uint16(0); i <= 40; i++ {
		resp := makeSignInFailureResp(RespID(i))
		if len(resp) != 1 {
			t.Errorf("makeSignInFailureResp(%d) len = %d, want 1", i, len(resp))
		}
		if resp[0] != uint8(i) {
			t.Errorf("makeSignInFailureResp(%d) = %d", i, resp[0])
		}
	}
}

func TestSignSuccessIsOne(t *testing.T) {
	// SIGN_SUCCESS must be 1 for the protocol to work correctly
	if SIGN_SUCCESS != 1 {
		t.Errorf("SIGN_SUCCESS = %d, must be 1", SIGN_SUCCESS)
	}
}

func TestSignUnknownIsZero(t *testing.T) {
	// SIGN_UNKNOWN must be 0 as the zero value
	if SIGN_UNKNOWN != 0 {
		t.Errorf("SIGN_UNKNOWN = %d, must be 0", SIGN_UNKNOWN)
	}
}

func TestRespIDValues(t *testing.T) {
	// Test specific RespID values are correct
	tests := []struct {
		name   string
		respID RespID
		value  uint16
	}{
		{"SIGN_UNKNOWN", SIGN_UNKNOWN, 0},
		{"SIGN_SUCCESS", SIGN_SUCCESS, 1},
		{"SIGN_EFAILED", SIGN_EFAILED, 2},
		{"SIGN_EILLEGAL", SIGN_EILLEGAL, 3},
		{"SIGN_ESUSPEND", SIGN_ESUSPEND, 15},
		{"SIGN_EELIMINATE", SIGN_EELIMINATE, 16},
		{"SIGN_EIPADDR", SIGN_EIPADDR, 23},
		{"SIGN_EMAINTE", SIGN_EMAINTE, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint16(tt.respID) != tt.value {
				t.Errorf("%s = %d, want %d", tt.name, uint16(tt.respID), tt.value)
			}
		})
	}
}

func TestUnknownRespIDRange(t *testing.T) {
	// Test the unknown IDs 32-35
	unknownIDs := []RespID{UNK_32, UNK_33, UNK_34, UNK_35}
	expectedValues := []uint16{32, 33, 34, 35}

	for i, id := range unknownIDs {
		if uint16(id) != expectedValues[i] {
			t.Errorf("Unknown ID %d = %d, want %d", i, uint16(id), expectedValues[i])
		}
	}
}

func TestSpecialRespIDs(t *testing.T) {
	// Test platform-specific IDs
	if SIGN_XBRESPONSE != 36 {
		t.Errorf("SIGN_XBRESPONSE = %d, want 36", SIGN_XBRESPONSE)
	}
	if SIGN_EPSI != 37 {
		t.Errorf("SIGN_EPSI = %d, want 37", SIGN_EPSI)
	}
	if SIGN_EMBID_PSI != 38 {
		t.Errorf("SIGN_EMBID_PSI = %d, want 38", SIGN_EMBID_PSI)
	}
}

func TestMakeSignInFailureRespBoundary(t *testing.T) {
	// Test boundary values
	resp := makeSignInFailureResp(RespID(0))
	if resp[0] != 0 {
		t.Errorf("makeSignInFailureResp(0) = %d, want 0", resp[0])
	}

	resp = makeSignInFailureResp(RespID(255))
	if resp[0] != 255 {
		t.Errorf("makeSignInFailureResp(255) = %d, want 255", resp[0])
	}
}

func TestErrorRespIDsAreDifferent(t *testing.T) {
	// Ensure all error codes are unique
	seen := make(map[RespID]bool)
	errorCodes := []RespID{
		SIGN_UNKNOWN, SIGN_SUCCESS, SIGN_EFAILED, SIGN_EILLEGAL,
		SIGN_EALERT, SIGN_EABORT, SIGN_ERESPONSE, SIGN_EDATABASE,
		SIGN_EABSENCE, SIGN_ERESIGN, SIGN_ESUSPEND_D, SIGN_ELOCK,
		SIGN_EPASS, SIGN_ERIGHT, SIGN_EAUTH, SIGN_ESUSPEND,
		SIGN_EELIMINATE, SIGN_ECLOSE, SIGN_ECLOSE_EX, SIGN_EINTERVAL,
		SIGN_EMOVED, SIGN_ENOTREADY, SIGN_EALREADY, SIGN_EIPADDR,
		SIGN_EHANGAME, SIGN_UPD_ONLY, SIGN_EMBID, SIGN_ECOGCODE,
		SIGN_ETOKEN, SIGN_ECOGLINK, SIGN_EMAINTE, SIGN_EMAINTE_NOUPDATE,
	}

	for _, code := range errorCodes {
		if seen[code] {
			t.Errorf("Duplicate RespID value: %d", code)
		}
		seen[code] = true
	}
}

func TestFailureRespIsMinimal(t *testing.T) {
	// Failure response should be exactly 1 byte for efficiency
	for i := RespID(0); i <= SIGN_EMBID_PSI; i++ {
		if i == SIGN_SUCCESS {
			continue // Success has different format
		}
		resp := makeSignInFailureResp(i)
		if len(resp) != 1 {
			t.Errorf("makeSignInFailureResp(%d) should be 1 byte, got %d", i, len(resp))
		}
	}
}

func TestNewServer(t *testing.T) {
	// Test that NewServer creates a valid server
	cfg := &Config{
		Logger:      nil,
		DB:          nil,
		ErupeConfig: nil,
	}

	s := NewServer(cfg)
	if s == nil {
		t.Fatal("NewServer() returned nil")
	}
	if s.isShuttingDown {
		t.Error("New server should not be shutting down")
	}
}

func TestNewServerWithNilConfig(t *testing.T) {
	// Testing with nil fields in config
	cfg := &Config{}
	s := NewServer(cfg)
	if s == nil {
		t.Fatal("NewServer() returned nil for empty config")
	}
}

func TestServerType(t *testing.T) {
	// Test Server struct fields
	s := &Server{}
	if s.isShuttingDown {
		t.Error("Zero value server should not be shutting down")
	}
	if s.sessions != nil {
		t.Error("Zero value server should have nil sessions map")
	}
}

func TestConfigFields(t *testing.T) {
	// Test Config struct fields
	cfg := &Config{
		Logger:      nil,
		DB:          nil,
		ErupeConfig: nil,
	}

	if cfg.Logger != nil {
		t.Error("Config Logger should be nil")
	}
	if cfg.DB != nil {
		t.Error("Config DB should be nil")
	}
	if cfg.ErupeConfig != nil {
		t.Error("Config ErupeConfig should be nil")
	}
}
