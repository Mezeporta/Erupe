package entranceserver

import (
	"bytes"
	"io"
	"net"
	"os"
	"sync"
	"testing"
	"time"

	cfg "erupe-ce/config"
	"erupe-ce/network"

	"go.uber.org/zap"
)

func TestSanitizeAddr(t *testing.T) {
	tests := []struct {
		name string
		addr string
		want string
	}{
		{"ip_port", "127.0.0.1:8080", "127.0.0.1_8080"},
		{"no_colon", "127.0.0.1", "127.0.0.1"},
		{"empty", "", ""},
		{"multiple_colons", "::1:8080", "__1_8080"},
		{"ipv6", "[::1]:8080", "[__1]_8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := sanitizeAddr(tt.addr)
			if got != tt.want {
				t.Errorf("sanitizeAddr(%q) = %q, want %q", tt.addr, got, tt.want)
			}
		})
	}
}

// mockConn implements net.Conn for testing capture functions.
type mockConn struct {
	readBuf  *bytes.Buffer
	writeBuf *bytes.Buffer
	mu       sync.Mutex
	closed   bool
}

func newMockConn() *mockConn {
	return &mockConn{
		readBuf:  new(bytes.Buffer),
		writeBuf: new(bytes.Buffer),
	}
}

func (m *mockConn) Read(b []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.EOF
	}
	return m.readBuf.Read(b)
}

func (m *mockConn) Write(b []byte) (int, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.closed {
		return 0, io.ErrClosedPipe
	}
	return m.writeBuf.Write(b)
}

func (m *mockConn) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockConn) LocalAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 53310}
}

func (m *mockConn) RemoteAddr() net.Addr {
	return &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}
}

func (m *mockConn) SetDeadline(time.Time) error      { return nil }
func (m *mockConn) SetReadDeadline(time.Time) error  { return nil }
func (m *mockConn) SetWriteDeadline(time.Time) error { return nil }

func TestStartEntranceCapture_Disabled(t *testing.T) {
	server := &Server{
		logger: zap.NewNop(),
		erupeConfig: &cfg.Config{
			Capture: cfg.CaptureOptions{
				Enabled:         false,
				CaptureEntrance: false,
			},
		},
	}

	mc := newMockConn()
	origConn := network.NewCryptConn(mc, cfg.ZZ, nil)
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	resultConn, cleanup := startEntranceCapture(server, origConn, remoteAddr)

	if resultConn != origConn {
		t.Error("disabled capture should return original conn")
	}
	cleanup()
}

func TestStartEntranceCapture_EnabledButEntranceDisabled(t *testing.T) {
	server := &Server{
		logger: zap.NewNop(),
		erupeConfig: &cfg.Config{
			Capture: cfg.CaptureOptions{
				Enabled:         true,
				CaptureEntrance: false,
			},
		},
	}

	mc := newMockConn()
	origConn := network.NewCryptConn(mc, cfg.ZZ, nil)
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	resultConn, cleanup := startEntranceCapture(server, origConn, remoteAddr)

	if resultConn != origConn {
		t.Error("capture with entrance disabled should return original conn")
	}
	cleanup()
}

func TestStartEntranceCapture_EnabledSuccess(t *testing.T) {
	outputDir := t.TempDir()
	server := &Server{
		logger: zap.NewNop(),
		erupeConfig: &cfg.Config{
			Host:     "127.0.0.1",
			Entrance: cfg.Entrance{Port: 53310},
			Capture: cfg.CaptureOptions{
				Enabled:         true,
				CaptureEntrance: true,
				OutputDir:       outputDir,
			},
		},
	}

	mc := newMockConn()
	origConn := network.NewCryptConn(mc, cfg.ZZ, nil)
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	resultConn, cleanup := startEntranceCapture(server, origConn, remoteAddr)
	defer cleanup()

	if resultConn == origConn {
		t.Error("enabled capture should return a different (recording) conn")
	}
}

func TestStartEntranceCapture_DefaultOutputDir(t *testing.T) {
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	server := &Server{
		logger: zap.NewNop(),
		erupeConfig: &cfg.Config{
			Host:     "127.0.0.1",
			Entrance: cfg.Entrance{Port: 53310},
			Capture: cfg.CaptureOptions{
				Enabled:         true,
				CaptureEntrance: true,
				OutputDir:       "", // empty → should default to "captures"
			},
		},
	}

	mc := newMockConn()
	origConn := network.NewCryptConn(mc, cfg.ZZ, nil)
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	resultConn, cleanup := startEntranceCapture(server, origConn, remoteAddr)
	defer cleanup()

	if resultConn == origConn {
		t.Error("capture with default dir should return recording conn")
	}

	info, err := os.Stat("captures")
	if err != nil {
		t.Fatalf("default 'captures' directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("'captures' should be a directory")
	}
}
