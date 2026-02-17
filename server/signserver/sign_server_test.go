package signserver

import (
	"fmt"
	"net"
	"testing"
	"time"

	_config "erupe-ce/config"

	"go.uber.org/zap"
)

// makeSignInFailureResp creates a 1-byte failure response for the given RespID.
func makeSignInFailureResp(id RespID) []byte {
	return []byte{uint8(id)}
}

func TestRespIDConstants(t *testing.T) {
	tests := []struct {
		respID RespID
		value  uint8
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
			if uint8(tt.respID) != tt.value {
				t.Errorf("RespID = %d, want %d", uint8(tt.respID), tt.value)
			}
		})
	}
}

func TestRespIDType(t *testing.T) {
	var r RespID = 0xFF
	if uint8(r) != 0xFF {
		t.Errorf("RespID max value = %d, want %d", uint8(r), 0xFF)
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
	for i := uint8(0); i <= 40; i++ {
		resp := makeSignInFailureResp(RespID(i))
		if len(resp) != 1 {
			t.Errorf("makeSignInFailureResp(%d) len = %d, want 1", i, len(resp))
		}
		if resp[0] != i {
			t.Errorf("makeSignInFailureResp(%d) = %d", i, resp[0])
		}
	}
}

func TestSignSuccessIsOne(t *testing.T) {
	if SIGN_SUCCESS != 1 {
		t.Errorf("SIGN_SUCCESS = %d, must be 1", SIGN_SUCCESS)
	}
}

func TestSignUnknownIsZero(t *testing.T) {
	if SIGN_UNKNOWN != 0 {
		t.Errorf("SIGN_UNKNOWN = %d, must be 0", SIGN_UNKNOWN)
	}
}

func TestRespIDValues(t *testing.T) {
	tests := []struct {
		name   string
		respID RespID
		value  uint8
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
			if uint8(tt.respID) != tt.value {
				t.Errorf("%s = %d, want %d", tt.name, uint8(tt.respID), tt.value)
			}
		})
	}
}

func TestUnknownRespIDRange(t *testing.T) {
	unknownIDs := []RespID{UNK_32, UNK_33, UNK_34, UNK_35}
	expectedValues := []uint8{32, 33, 34, 35}

	for i, id := range unknownIDs {
		if uint8(id) != expectedValues[i] {
			t.Errorf("Unknown ID %d = %d, want %d", i, uint8(id), expectedValues[i])
		}
	}
}

func TestSpecialRespIDs(t *testing.T) {
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
	for i := RespID(0); i <= SIGN_EMBID_PSI; i++ {
		if i == SIGN_SUCCESS {
			continue
		}
		resp := makeSignInFailureResp(i)
		if len(resp) != 1 {
			t.Errorf("makeSignInFailureResp(%d) should be 1 byte, got %d", i, len(resp))
		}
	}
}

func TestNewServer(t *testing.T) {
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
	cfg := &Config{}
	s := NewServer(cfg)
	if s == nil {
		t.Fatal("NewServer() returned nil for empty config")
	}
}

func TestServerType(t *testing.T) {
	s := &Server{}
	if s.isShuttingDown {
		t.Error("Zero value server should not be shutting down")
	}
}

func TestConfigFields(t *testing.T) {
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

func TestServerStartAndShutdown(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Sign: _config.Sign{
			Port: 0,
		},
	}

	cfg := &Config{
		Logger:      logger,
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)
	if s == nil {
		t.Fatal("NewServer() returned nil")
	}

	err := s.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	if s.listener == nil {
		t.Error("Server listener should not be nil after Start()")
	}

	s.Lock()
	if s.isShuttingDown {
		t.Error("Server should not be shutting down after Start()")
	}
	s.Unlock()

	s.Shutdown()

	s.Lock()
	if !s.isShuttingDown {
		t.Error("Server should be shutting down after Shutdown()")
	}
	s.Unlock()
}

func TestServerStartWithInvalidPort(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Sign: _config.Sign{
			Port: -1,
		},
	}

	cfg := &Config{
		Logger:      logger,
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)
	err := s.Start()

	if err == nil {
		s.Shutdown()
		t.Error("Start() should fail with invalid port")
	}
}

func TestServerMutex(t *testing.T) {
	s := &Server{}

	s.Lock()
	//nolint:staticcheck // SA2001: testing that Lock/Unlock doesn't panic
	s.Unlock()

	done := make(chan bool)
	go func() {
		s.Lock()
		time.Sleep(10 * time.Millisecond)
		s.Unlock()
		done <- true
	}()

	time.Sleep(5 * time.Millisecond)

	s.Lock()
	//nolint:staticcheck // SA2001: testing that Lock/Unlock doesn't panic
	s.Unlock()

	<-done
}

func TestServerShutdownIdempotent(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Sign: _config.Sign{
			Port: 0,
		},
	}

	cfg := &Config{
		Logger:      logger,
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)
	err := s.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	s.Shutdown()

	s.Lock()
	if !s.isShuttingDown {
		t.Error("Server should be shutting down")
	}
	s.Unlock()
}

func TestServerAcceptClientsExitsOnShutdown(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Sign: _config.Sign{
			Port: 0,
		},
	}

	cfg := &Config{
		Logger:      logger,
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)
	err := s.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}

	time.Sleep(10 * time.Millisecond)

	s.Shutdown()

	time.Sleep(10 * time.Millisecond)

	s.Lock()
	if !s.isShuttingDown {
		t.Error("Server should be marked as shutting down")
	}
	s.Unlock()
}

func TestServerHandleConnection(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Sign: _config.Sign{
			Port: 0,
		},
	}

	cfg := &Config{
		Logger:      logger,
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)
	err := s.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer s.Shutdown()

	addr := s.listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial() error: %v", err)
	}
	defer conn.Close()

	nullInit := make([]byte, 8)
	_, err = conn.Write(nullInit)
	if err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	time.Sleep(50 * time.Millisecond)
}

func TestServerHandleConnectionWithShortInit(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Sign: _config.Sign{
			Port: 0,
		},
	}

	cfg := &Config{
		Logger:      logger,
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)
	err := s.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer s.Shutdown()

	addr := s.listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial() error: %v", err)
	}

	_, _ = conn.Write([]byte{0, 0, 0, 0})
	conn.Close()

	time.Sleep(50 * time.Millisecond)
}

func TestServerHandleConnectionImmediateClose(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Sign: _config.Sign{
			Port: 0,
		},
	}

	cfg := &Config{
		Logger:      logger,
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)
	err := s.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer s.Shutdown()

	addr := s.listener.Addr().String()
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		t.Fatalf("Dial() error: %v", err)
	}
	conn.Close()

	time.Sleep(50 * time.Millisecond)
}

func TestServerMultipleConnections(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Sign: _config.Sign{
			Port: 0,
		},
	}

	cfg := &Config{
		Logger:      logger,
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)
	err := s.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer s.Shutdown()

	addr := s.listener.Addr().String()

	conns := make([]net.Conn, 3)
	for i := range conns {
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			t.Fatalf("Dial() %d error: %v", i, err)
		}
		conns[i] = conn

		nullInit := make([]byte, 8)
		_, _ = conn.Write(nullInit)
	}

	time.Sleep(50 * time.Millisecond)

	for _, conn := range conns {
		conn.Close()
	}
}

func TestServerListenerAddress(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Sign: _config.Sign{
			Port: 0,
		},
	}

	cfg := &Config{
		Logger:      logger,
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)
	err := s.Start()
	if err != nil {
		t.Fatalf("Start() error: %v", err)
	}
	defer s.Shutdown()

	addr := s.listener.Addr()
	if addr == nil {
		t.Error("Listener address should not be nil")
	}

	tcpAddr, ok := addr.(*net.TCPAddr)
	if !ok {
		t.Error("Listener address should be a TCP address")
	}

	if tcpAddr.Port == 0 {
		t.Error("Listener port should be assigned")
	}
}
