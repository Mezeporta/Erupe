package signserver

import (
	"net"
	"testing"

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

func TestStartSignCapture_Disabled(t *testing.T) {
	server := &Server{
		logger: zap.NewNop(),
		erupeConfig: &cfg.Config{
			Capture: cfg.CaptureOptions{
				Enabled:     false,
				CaptureSign: false,
			},
		},
	}

	mc := newMockConn()
	origConn := network.NewCryptConn(mc, cfg.ZZ, nil)
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	resultConn, cleanup := startSignCapture(server, origConn, remoteAddr)

	if resultConn != origConn {
		t.Error("startSignCapture() disabled should return original conn")
	}

	// cleanup should be a no-op, just verify it doesn't panic
	cleanup()
}

func TestStartSignCapture_EnabledButSignDisabled(t *testing.T) {
	server := &Server{
		logger: zap.NewNop(),
		erupeConfig: &cfg.Config{
			Capture: cfg.CaptureOptions{
				Enabled:     true,
				CaptureSign: false,
			},
		},
	}

	mc := newMockConn()
	origConn := network.NewCryptConn(mc, cfg.ZZ, nil)
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	resultConn, cleanup := startSignCapture(server, origConn, remoteAddr)

	if resultConn != origConn {
		t.Error("startSignCapture() with sign disabled should return original conn")
	}
	cleanup()
}
