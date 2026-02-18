package api

import (
	"testing"

	_config "erupe-ce/config"
	"go.uber.org/zap"
)

// NewTestLogger creates a logger for testing
func NewTestLogger(t *testing.T) *zap.Logger {
	logger, err := zap.NewDevelopment()
	if err != nil {
		t.Fatalf("Failed to create test logger: %v", err)
	}
	return logger
}

// NewTestConfig creates a default test configuration
func NewTestConfig() *_config.Config {
	return &_config.Config{
		API: _config.API{
			Port:        8000,
			PatchServer: "http://localhost:8080",
			Banners:     []_config.APISignBanner{},
			Messages:    []_config.APISignMessage{},
			Links:       []_config.APISignLink{},
		},
		Screenshots: _config.ScreenshotsOptions{
			Enabled:       true,
			OutputDir:     "/tmp/screenshots",
			UploadQuality: 85,
		},
		DebugOptions: _config.DebugOptions{
			MaxLauncherHR: false,
		},
		GameplayOptions: _config.GameplayOptions{
			MezFesSoloTickets:     100,
			MezFesGroupTickets:    50,
			MezFesDuration:        604800, // 1 week
			MezFesSwitchMinigame:  false,
		},
		LoginNotices:   []string{"Welcome to Erupe!"},
		HideLoginNotice: false,
	}
}

