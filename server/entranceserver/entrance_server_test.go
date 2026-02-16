package entranceserver

import (
	"net"
	"testing"
	"time"

	_config "erupe-ce/config"

	"go.uber.org/zap"
)

func TestNewServer(t *testing.T) {
	cfg := &Config{
		Logger:      nil,
		DB:          nil,
		ErupeConfig: &_config.Config{},
	}

	s := NewServer(cfg)
	if s == nil {
		t.Fatal("NewServer() returned nil")
	}
	if s.isShuttingDown {
		t.Error("New server should not be shutting down")
	}
	if s.erupeConfig == nil {
		t.Error("erupeConfig should not be nil")
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
	if s.listener != nil {
		t.Error("Zero value server should have nil listener")
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

func TestServerShutdownFlag(t *testing.T) {
	cfg := &Config{
		ErupeConfig: &_config.Config{},
	}
	s := NewServer(cfg)

	if s.isShuttingDown {
		t.Error("New server should not be shutting down")
	}

	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	if !s.isShuttingDown {
		t.Error("Server should be shutting down after flag is set")
	}
}

func TestServerConfigStorage(t *testing.T) {
	erupeConfig := &_config.Config{
		Host: "192.168.1.100",
		Entrance: _config.Entrance{
			Enabled: true,
			Port:    53310,
			Entries: []_config.EntranceServerInfo{
				{
					Name: "Test Server",
					IP:   "127.0.0.1",
					Type: 1,
				},
			},
		},
	}

	cfg := &Config{
		ErupeConfig: erupeConfig,
	}

	s := NewServer(cfg)

	if s.erupeConfig.Host != "192.168.1.100" {
		t.Errorf("Host = %s, want 192.168.1.100", s.erupeConfig.Host)
	}
	if s.erupeConfig.Entrance.Port != 53310 {
		t.Errorf("Entrance.Port = %d, want 53310", s.erupeConfig.Entrance.Port)
	}
}

func TestServerEntranceEntries(t *testing.T) {
	entries := []_config.EntranceServerInfo{
		{
			Name:        "World 1",
			IP:          "10.0.0.1",
			Type:        1,
			Recommended: 1,
			Channels: []_config.EntranceChannelInfo{
				{Port: 54001, MaxPlayers: 100},
				{Port: 54002, MaxPlayers: 100},
			},
		},
		{
			Name:        "World 2",
			IP:          "10.0.0.2",
			Type:        2,
			Recommended: 0,
			Channels: []_config.EntranceChannelInfo{
				{Port: 54003, MaxPlayers: 50},
			},
		},
	}

	erupeConfig := &_config.Config{
		Entrance: _config.Entrance{
			Enabled: true,
			Port:    53310,
			Entries: entries,
		},
	}

	cfg := &Config{ErupeConfig: erupeConfig}
	s := NewServer(cfg)

	if len(s.erupeConfig.Entrance.Entries) != 2 {
		t.Errorf("Entries count = %d, want 2", len(s.erupeConfig.Entrance.Entries))
	}

	if s.erupeConfig.Entrance.Entries[0].Name != "World 1" {
		t.Errorf("First entry name = %s, want World 1", s.erupeConfig.Entrance.Entries[0].Name)
	}

	if len(s.erupeConfig.Entrance.Entries[0].Channels) != 2 {
		t.Errorf("First entry channels = %d, want 2", len(s.erupeConfig.Entrance.Entries[0].Channels))
	}
}

func TestEncryptDecryptRoundTrip(t *testing.T) {
	tests := []struct {
		name string
		data []byte
		key  byte
	}{
		{"empty", []byte{}, 0x00},
		{"single byte", []byte{0x42}, 0x00},
		{"multiple bytes", []byte{0x01, 0x02, 0x03, 0x04}, 0x00},
		{"with key", []byte{0xDE, 0xAD, 0xBE, 0xEF}, 0x55},
		{"max key", []byte{0x01, 0x02}, 0xFF},
		{"long data", make([]byte, 100), 0x42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encrypted := EncryptBin8(tt.data, tt.key)
			decrypted := DecryptBin8(encrypted, tt.key)

			if len(decrypted) != len(tt.data) {
				t.Errorf("decrypted length = %d, want %d", len(decrypted), len(tt.data))
				return
			}

			for i := range tt.data {
				if decrypted[i] != tt.data[i] {
					t.Errorf("decrypted[%d] = 0x%X, want 0x%X", i, decrypted[i], tt.data[i])
				}
			}
		})
	}
}

func TestCalcSum32Deterministic(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}

	sum1 := CalcSum32(data)
	sum2 := CalcSum32(data)

	if sum1 != sum2 {
		t.Errorf("CalcSum32 not deterministic: got 0x%X and 0x%X", sum1, sum2)
	}
}

func TestCalcSum32DifferentInputs(t *testing.T) {
	data1 := []byte{0x01, 0x02, 0x03}
	data2 := []byte{0x01, 0x02, 0x04}

	sum1 := CalcSum32(data1)
	sum2 := CalcSum32(data2)

	if sum1 == sum2 {
		t.Error("Different inputs should produce different checksums")
	}
}

func TestEncryptBin8KeyVariation(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04}

	enc1 := EncryptBin8(data, 0x00)
	enc2 := EncryptBin8(data, 0x01)
	enc3 := EncryptBin8(data, 0xFF)

	if bytesEqual(enc1, enc2) {
		t.Error("Different keys should produce different encrypted data (0x00 vs 0x01)")
	}
	if bytesEqual(enc2, enc3) {
		t.Error("Different keys should produce different encrypted data (0x01 vs 0xFF)")
	}
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestEncryptBin8LengthPreservation(t *testing.T) {
	lengths := []int{0, 1, 7, 8, 9, 100, 1000}

	for _, length := range lengths {
		data := make([]byte, length)
		for i := range data {
			data[i] = byte(i % 256)
		}

		encrypted := EncryptBin8(data, 0x42)
		if len(encrypted) != length {
			t.Errorf("EncryptBin8 length %d changed to %d", length, len(encrypted))
		}
	}
}

func TestCalcSum32LargeInput(t *testing.T) {
	data := make([]byte, 10000)
	for i := range data {
		data[i] = byte(i % 256)
	}

	sum := CalcSum32(data)
	sum2 := CalcSum32(data)
	if sum != sum2 {
		t.Errorf("CalcSum32 inconsistent for large input: 0x%X vs 0x%X", sum, sum2)
	}
}

func TestServerMutexLocking(t *testing.T) {
	cfg := &Config{ErupeConfig: &_config.Config{}}
	s := NewServer(cfg)

	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	s.Lock()
	result := s.isShuttingDown
	s.Unlock()

	if !result {
		t.Error("Mutex should protect isShuttingDown flag")
	}
}

func TestServerStartAndShutdown(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Entrance: _config.Entrance{
			Enabled: true,
			Port:    0,
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
		Entrance: _config.Entrance{
			Port: 1,
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

func TestServerListenerAddress(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Entrance: _config.Entrance{
			Enabled: true,
			Port:    0,
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

func TestServerAcceptClientsExitsOnShutdown(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Entrance: _config.Entrance{
			Enabled: true,
			Port:    0,
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

func TestServerHandleConnectionImmediateClose(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Entrance: _config.Entrance{
			Enabled: true,
			Port:    0,
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

func TestServerHandleConnectionShortInit(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Entrance: _config.Entrance{
			Enabled: true,
			Port:    0,
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

func TestServerMultipleConnections(t *testing.T) {
	logger := zap.NewNop()
	erupeConfig := &_config.Config{
		Entrance: _config.Entrance{
			Enabled: true,
			Port:    0,
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
	}

	time.Sleep(50 * time.Millisecond)

	for _, conn := range conns {
		conn.Close()
	}
}
