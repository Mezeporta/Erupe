package api

import (
	"testing"

	cfg "erupe-ce/config"
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
func NewTestConfig() *cfg.Config {
	return &cfg.Config{
		API: cfg.API{
			Port:        8000,
			PatchServer: "http://localhost:8080",
			Banners:     []cfg.APISignBanner{},
			Messages:    []cfg.APISignMessage{},
			Links:       []cfg.APISignLink{},
		},
		Screenshots: cfg.ScreenshotsOptions{
			Enabled:       true,
			OutputDir:     "/tmp/screenshots",
			UploadQuality: 85,
		},
		DebugOptions: cfg.DebugOptions{
			MaxLauncherHR: false,
		},
		GameplayOptions: cfg.GameplayOptions{
			MezFesSoloTickets:    100,
			MezFesGroupTickets:   50,
			MezFesDuration:       604800, // 1 week
			MezFesSwitchMinigame: false,
		},
		LoginNotices:    []string{"Welcome to Erupe!"},
		HideLoginNotice: false,
	}
}
