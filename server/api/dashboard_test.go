package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestDashboardStatsJSON_NoDB verifies the stats endpoint returns valid JSON
// with safe zero values when no database is configured.
func TestDashboardStatsJSON_NoDB(t *testing.T) {
	logger := NewTestLogger(t)
	defer func() { _ = logger.Sync() }()

	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		startTime:   time.Now().Add(-5 * time.Minute),
		// db intentionally nil
	}

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats", nil)
	rec := httptest.NewRecorder()

	server.DashboardStatsJSON(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	ct := rec.Header().Get("Content-Type")
	if !strings.HasPrefix(ct, "application/json") {
		t.Errorf("Expected Content-Type application/json, got %q", ct)
	}

	var stats DashboardStats
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Verify required fields are present and have expected zero-DB values.
	if stats.ServerVersion == "" {
		t.Error("Expected non-empty ServerVersion")
	}
	if stats.Uptime == "" || stats.Uptime == "unknown" {
		// startTime is set so uptime should be computed, not "unknown".
		t.Errorf("Expected computed uptime, got %q", stats.Uptime)
	}
	if stats.TotalAccounts != 0 {
		t.Errorf("Expected TotalAccounts=0 without DB, got %d", stats.TotalAccounts)
	}
	if stats.TotalCharacters != 0 {
		t.Errorf("Expected TotalCharacters=0 without DB, got %d", stats.TotalCharacters)
	}
	if stats.OnlinePlayers != 0 {
		t.Errorf("Expected OnlinePlayers=0 without DB, got %d", stats.OnlinePlayers)
	}
	if stats.DatabaseOK {
		t.Error("Expected DatabaseOK=false without DB")
	}
	if stats.Channels != nil {
		t.Errorf("Expected nil Channels without DB, got %v", stats.Channels)
	}
}

// TestDashboardStatsJSON_UptimeUnknown verifies "unknown" uptime when startTime is zero.
func TestDashboardStatsJSON_UptimeUnknown(t *testing.T) {
	logger := NewTestLogger(t)
	defer func() { _ = logger.Sync() }()

	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		// startTime is zero value
	}

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats", nil)
	rec := httptest.NewRecorder()

	server.DashboardStatsJSON(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	var stats DashboardStats
	if err := json.NewDecoder(rec.Body).Decode(&stats); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if stats.Uptime != "unknown" {
		t.Errorf("Expected Uptime='unknown' for zero startTime, got %q", stats.Uptime)
	}
}

// TestDashboardStatsJSON_JSONShape validates every field of the DashboardStats payload.
func TestDashboardStatsJSON_JSONShape(t *testing.T) {
	logger := NewTestLogger(t)
	defer func() { _ = logger.Sync() }()

	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		startTime:   time.Now(),
	}

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/stats", nil)
	rec := httptest.NewRecorder()

	server.DashboardStatsJSON(rec, req)

	// Decode into a raw map so we can check key presence independent of type.
	var raw map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("Failed to decode response as raw map: %v", err)
	}

	requiredKeys := []string{
		"uptime", "serverVersion", "clientMode",
		"onlinePlayers", "totalAccounts", "totalCharacters",
		"databaseOK",
	}
	for _, key := range requiredKeys {
		if _, ok := raw[key]; !ok {
			t.Errorf("Missing required JSON key %q", key)
		}
	}
}

// TestFormatDuration covers the human-readable duration formatter.
func TestFormatDuration(t *testing.T) {
	tests := []struct {
		d    time.Duration
		want string
	}{
		{10 * time.Second, "10s"},
		{90 * time.Second, "1m 30s"},
		{2*time.Hour + 15*time.Minute + 5*time.Second, "2h 15m 5s"},
		{25*time.Hour + 3*time.Minute + 0*time.Second, "1d 1h 3m 0s"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := formatDuration(tt.d)
			if got != tt.want {
				t.Errorf("formatDuration(%v) = %q, want %q", tt.d, got, tt.want)
			}
		})
	}
}
