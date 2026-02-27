package signserver

import (
	"net"
	"os"
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

func TestStartSignCapture_EnabledSuccess(t *testing.T) {
	outputDir := t.TempDir()
	server := &Server{
		logger: zap.NewNop(),
		erupeConfig: &cfg.Config{
			Host: "127.0.0.1",
			Sign: cfg.Sign{Port: 53312},
			Capture: cfg.CaptureOptions{
				Enabled:     true,
				CaptureSign: true,
				OutputDir:   outputDir,
			},
		},
	}

	mc := newMockConn()
	origConn := network.NewCryptConn(mc, cfg.ZZ, nil)
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	resultConn, cleanup := startSignCapture(server, origConn, remoteAddr)
	defer cleanup()

	if resultConn == origConn {
		t.Error("startSignCapture() enabled should return a different (recording) conn")
	}
}

func TestStartSignCapture_DefaultOutputDir(t *testing.T) {
	// Use a temp dir as working directory to avoid polluting the project
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
			Host: "127.0.0.1",
			Sign: cfg.Sign{Port: 53312},
			Capture: cfg.CaptureOptions{
				Enabled:     true,
				CaptureSign: true,
				OutputDir:   "", // empty → should default to "captures"
			},
		},
	}

	mc := newMockConn()
	origConn := network.NewCryptConn(mc, cfg.ZZ, nil)
	remoteAddr := &net.TCPAddr{IP: net.ParseIP("127.0.0.1"), Port: 12345}

	resultConn, cleanup := startSignCapture(server, origConn, remoteAddr)
	defer cleanup()

	if resultConn == origConn {
		t.Error("startSignCapture() with default dir should return recording conn")
	}

	// Verify the "captures" directory was created
	info, err := os.Stat("captures")
	if err != nil {
		t.Fatalf("default 'captures' directory not created: %v", err)
	}
	if !info.IsDir() {
		t.Error("'captures' should be a directory")
	}
}
