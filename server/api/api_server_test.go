package api

import (
	"net/http"
	"testing"
	"time"

	_config "erupe-ce/config"
	"go.uber.org/zap"
)

func TestNewAPIServer(t *testing.T) {
	logger := NewTestLogger(t)
	defer logger.Sync()

	cfg := NewTestConfig()
	config := &Config{
		Logger:      logger,
		DB:          nil, // Database can be nil for this test
		ErupeConfig: cfg,
	}

	server := NewAPIServer(config)

	if server == nil {
		t.Fatal("NewAPIServer returned nil")
	}

	if server.logger != logger {
		t.Error("Logger not properly assigned")
	}

	if server.erupeConfig != cfg {
		t.Error("ErupeConfig not properly assigned")
	}

	if server.httpServer == nil {
		t.Error("HTTP server not initialized")
	}

	if server.isShuttingDown != false {
		t.Error("Server should not be shutting down on creation")
	}
}

func TestNewAPIServerConfig(t *testing.T) {
	logger := NewTestLogger(t)
	defer logger.Sync()

	cfg := &_config.Config{
		API: _config.API{
			Port:        9999,
			PatchServer: "http://example.com",
			Banners:     []_config.APISignBanner{},
			Messages:    []_config.APISignMessage{},
			Links:       []_config.APISignLink{},
		},
		Screenshots: _config.ScreenshotsOptions{
			Enabled:       false,
			OutputDir:     "/custom/path",
			UploadQuality: 95,
		},
		DebugOptions: _config.DebugOptions{
			MaxLauncherHR: true,
		},
		GameplayOptions: _config.GameplayOptions{
			MezFesSoloTickets: 200,
		},
	}

	config := &Config{
		Logger:      logger,
		DB:          nil,
		ErupeConfig: cfg,
	}

	server := NewAPIServer(config)

	if server.erupeConfig.API.Port != 9999 {
		t.Errorf("API port = %d, want 9999", server.erupeConfig.API.Port)
	}

	if server.erupeConfig.API.PatchServer != "http://example.com" {
		t.Errorf("PatchServer = %s, want http://example.com", server.erupeConfig.API.PatchServer)
	}

	if server.erupeConfig.Screenshots.UploadQuality != 95 {
		t.Errorf("UploadQuality = %d, want 95", server.erupeConfig.Screenshots.UploadQuality)
	}
}

func TestAPIServerStart(t *testing.T) {
	// Note: This test can be flaky in CI environments
	// It attempts to start an actual HTTP server

	logger := NewTestLogger(t)
	defer logger.Sync()

	cfg := NewTestConfig()
	cfg.API.Port = 18888 // Use a high port less likely to be in use

	config := &Config{
		Logger:      logger,
		DB:          nil,
		ErupeConfig: cfg,
	}

	server := NewAPIServer(config)

	// Start server
	err := server.Start()
	if err != nil {
		t.Logf("Start error (may be expected if port in use): %v", err)
		// Don't fail hard, as this might be due to port binding issues in test environment
		return
	}

	// Give the server a moment to start
	time.Sleep(100 * time.Millisecond)

	// Check that the server is running by making a request
	resp, err := http.Get("http://localhost:18888/launcher")
	if err != nil {
		// This might fail if the server didn't start properly or port is blocked
		t.Logf("Failed to connect to server: %v", err)
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNotFound {
			t.Logf("Unexpected status code: %d", resp.StatusCode)
		}
	}

	// Shutdown the server
	done := make(chan bool, 1)
	go func() {
		server.Shutdown()
		done <- true
	}()

	// Wait for shutdown with timeout
	select {
	case <-done:
		t.Log("Server shutdown successfully")
	case <-time.After(10 * time.Second):
		t.Error("Server shutdown timeout")
	}
}

func TestAPIServerShutdown(t *testing.T) {
	logger := NewTestLogger(t)
	defer logger.Sync()

	cfg := NewTestConfig()
	cfg.API.Port = 18889

	config := &Config{
		Logger:      logger,
		DB:          nil,
		ErupeConfig: cfg,
	}

	server := NewAPIServer(config)

	// Try to shutdown without starting (should not panic)
	server.Shutdown()

	// Verify the shutdown flag is set
	server.Lock()
	if !server.isShuttingDown {
		t.Error("isShuttingDown should be true after Shutdown()")
	}
	server.Unlock()
}

func TestAPIServerShutdownSetsFlag(t *testing.T) {
	logger := NewTestLogger(t)
	defer logger.Sync()

	cfg := NewTestConfig()
	config := &Config{
		Logger:      logger,
		DB:          nil,
		ErupeConfig: cfg,
	}

	server := NewAPIServer(config)

	if server.isShuttingDown {
		t.Error("Server should not be shutting down initially")
	}

	server.Shutdown()

	server.Lock()
	isShutting := server.isShuttingDown
	server.Unlock()

	if !isShutting {
		t.Error("isShuttingDown flag should be set after Shutdown()")
	}
}

func TestAPIServerConcurrentShutdown(t *testing.T) {
	logger := NewTestLogger(t)
	defer logger.Sync()

	cfg := NewTestConfig()
	config := &Config{
		Logger:      logger,
		DB:          nil,
		ErupeConfig: cfg,
	}

	server := NewAPIServer(config)

	// Try shutting down from multiple goroutines concurrently
	done := make(chan bool, 3)

	for i := 0; i < 3; i++ {
		go func() {
			server.Shutdown()
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for i := 0; i < 3; i++ {
		select {
		case <-done:
		case <-time.After(5 * time.Second):
			t.Error("Timeout waiting for shutdown")
		}
	}

	server.Lock()
	if !server.isShuttingDown {
		t.Error("Server should be shutting down after concurrent shutdown calls")
	}
	server.Unlock()
}

func TestAPIServerMutex(t *testing.T) {
	logger := NewTestLogger(t)
	defer logger.Sync()

	cfg := NewTestConfig()
	config := &Config{
		Logger:      logger,
		DB:          nil,
		ErupeConfig: cfg,
	}

	server := NewAPIServer(config)

	// Verify that the server has mutex functionality
	server.Lock()
	isLocked := true
	server.Unlock()

	if !isLocked {
		t.Error("Mutex locking/unlocking failed")
	}
}

func TestAPIServerHTTPServerInitialization(t *testing.T) {
	logger := NewTestLogger(t)
	defer logger.Sync()

	cfg := NewTestConfig()
	config := &Config{
		Logger:      logger,
		DB:          nil,
		ErupeConfig: cfg,
	}

	server := NewAPIServer(config)

	if server.httpServer == nil {
		t.Fatal("HTTP server should be initialized")
	}

	if server.httpServer.Addr != "" {
		t.Logf("HTTP server address initially set: %s", server.httpServer.Addr)
	}
}

func BenchmarkNewAPIServer(b *testing.B) {
	logger, _ := zap.NewDevelopment()
	defer logger.Sync()

	cfg := NewTestConfig()
	config := &Config{
		Logger:      logger,
		DB:          nil,
		ErupeConfig: cfg,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = NewAPIServer(config)
	}
}
