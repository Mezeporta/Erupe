package setup

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"go.uber.org/zap"
)

func TestBuildDefaultConfig(t *testing.T) {
	req := FinishRequest{
		DBHost:            "myhost",
		DBPort:            5433,
		DBUser:            "myuser",
		DBPassword:        "secret",
		DBName:            "mydb",
		Host:              "10.0.0.1",
		ClientMode:        "ZZ",
		AutoCreateAccount: true,
	}
	cfg := buildDefaultConfig(req)

	// Check top-level keys from user input
	if cfg["Host"] != "10.0.0.1" {
		t.Errorf("Host = %v, want 10.0.0.1", cfg["Host"])
	}
	if cfg["ClientMode"] != "ZZ" {
		t.Errorf("ClientMode = %v, want ZZ", cfg["ClientMode"])
	}
	if cfg["AutoCreateAccount"] != true {
		t.Errorf("AutoCreateAccount = %v, want true", cfg["AutoCreateAccount"])
	}

	// Check database section
	db, ok := cfg["Database"].(map[string]interface{})
	if !ok {
		t.Fatal("Database section not a map")
	}
	if db["Host"] != "myhost" {
		t.Errorf("Database.Host = %v, want myhost", db["Host"])
	}
	if db["Port"] != 5433 {
		t.Errorf("Database.Port = %v, want 5433", db["Port"])
	}
	if db["User"] != "myuser" {
		t.Errorf("Database.User = %v, want myuser", db["User"])
	}
	if db["Password"] != "secret" {
		t.Errorf("Database.Password = %v, want secret", db["Password"])
	}
	if db["Database"] != "mydb" {
		t.Errorf("Database.Database = %v, want mydb", db["Database"])
	}

	// Wizard config is now minimal — only user-provided values.
	// Viper defaults fill the rest at load time.
	requiredKeys := []string{"Host", "ClientMode", "AutoCreateAccount", "Database"}
	for _, key := range requiredKeys {
		if _, ok := cfg[key]; !ok {
			t.Errorf("missing required key %q", key)
		}
	}

	// Verify it marshals to valid JSON
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		t.Fatalf("failed to marshal config: %v", err)
	}
	if len(data) < 50 {
		t.Errorf("config JSON unexpectedly short: %d bytes", len(data))
	}
}

func TestDetectIP(t *testing.T) {
	ws := &wizardServer{
		logger: zap.NewNop(),
		done:   make(chan struct{}),
	}
	req := httptest.NewRequest("GET", "/api/setup/detect-ip", nil)
	w := httptest.NewRecorder()
	ws.handleDetectIP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	ip, ok := resp["ip"]
	if !ok || ip == "" {
		t.Error("expected non-empty IP in response")
	}
}

func TestClientModes(t *testing.T) {
	ws := &wizardServer{
		logger: zap.NewNop(),
		done:   make(chan struct{}),
	}
	req := httptest.NewRequest("GET", "/api/setup/client-modes", nil)
	w := httptest.NewRecorder()
	ws.handleClientModes(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp map[string][]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	modes := resp["modes"]
	if len(modes) != 41 {
		t.Errorf("got %d modes, want 41", len(modes))
	}
	// First should be S1.0, last should be ZZ
	if modes[0] != "S1.0" {
		t.Errorf("first mode = %q, want S1.0", modes[0])
	}
	if modes[len(modes)-1] != "ZZ" {
		t.Errorf("last mode = %q, want ZZ", modes[len(modes)-1])
	}
}

func TestWriteConfig(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	cfg := buildDefaultConfig(FinishRequest{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "postgres",
		DBPassword: "pass",
		DBName:     "erupe",
		Host:       "127.0.0.1",
		ClientMode: "ZZ",
	})

	if err := writeConfig(cfg); err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("reading config.json: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("config.json is not valid JSON: %v", err)
	}
	if parsed["Host"] != "127.0.0.1" {
		t.Errorf("Host = %v, want 127.0.0.1", parsed["Host"])
	}
}

func TestHandleIndex(t *testing.T) {
	ws := &wizardServer{
		logger: zap.NewNop(),
		done:   make(chan struct{}),
	}
	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()
	ws.handleIndex(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if ct := w.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
	body := w.Body.String()
	if !contains(body, "Erupe Setup Wizard") {
		t.Error("response body missing wizard title")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func TestBuildDefaultConfig_EmptyLanguage(t *testing.T) {
	req := FinishRequest{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "postgres",
		DBPassword: "pass",
		DBName:     "erupe",
		Host:       "127.0.0.1",
		ClientMode: "ZZ",
		Language:   "", // empty — should default to "jp"
	}
	cfg := buildDefaultConfig(req)

	lang, ok := cfg["Language"].(string)
	if !ok {
		t.Fatal("Language is not a string")
	}
	if lang != "jp" {
		t.Errorf("Language = %q, want %q", lang, "jp")
	}
}

func TestHandleTestDB_InvalidJSON(t *testing.T) {
	ws := &wizardServer{
		logger: zap.NewNop(),
		done:   make(chan struct{}),
	}
	req := httptest.NewRequest("POST", "/api/setup/test-db", strings.NewReader("{invalid"))
	w := httptest.NewRecorder()
	ws.handleTestDB(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["error"] != "invalid JSON" {
		t.Errorf("error = %q, want %q", resp["error"], "invalid JSON")
	}
}

func TestHandleInitDB_InvalidJSON(t *testing.T) {
	ws := &wizardServer{
		logger: zap.NewNop(),
		done:   make(chan struct{}),
	}
	req := httptest.NewRequest("POST", "/api/setup/init-db", strings.NewReader("not json"))
	w := httptest.NewRecorder()
	ws.handleInitDB(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["error"] != "invalid JSON" {
		t.Errorf("error = %q, want %q", resp["error"], "invalid JSON")
	}
}

func TestHandleFinish_InvalidJSON(t *testing.T) {
	ws := &wizardServer{
		logger: zap.NewNop(),
		done:   make(chan struct{}),
	}
	req := httptest.NewRequest("POST", "/api/setup/finish", strings.NewReader("%%%"))
	w := httptest.NewRecorder()
	ws.handleFinish(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusBadRequest)
	}
	var resp map[string]string
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["error"] != "invalid JSON" {
		t.Errorf("error = %q, want %q", resp["error"], "invalid JSON")
	}
}

func TestHandleFinish_Success(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	done := make(chan struct{})
	ws := &wizardServer{
		logger: zap.NewNop(),
		done:   done,
	}

	body := `{"dbHost":"localhost","dbPort":5432,"dbUser":"postgres","dbPassword":"pw","dbName":"erupe","host":"10.0.0.5","clientMode":"G10","autoCreateAccount":false}`
	req := httptest.NewRequest("POST", "/api/setup/finish", strings.NewReader(body))
	w := httptest.NewRecorder()
	ws.handleFinish(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", w.Code, http.StatusOK)
	}

	// Verify config.json was written
	data, err := os.ReadFile(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("config.json not written: %v", err)
	}
	var parsed map[string]interface{}
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("config.json is not valid JSON: %v", err)
	}
	if parsed["Host"] != "10.0.0.5" {
		t.Errorf("Host = %v, want 10.0.0.5", parsed["Host"])
	}
	if parsed["ClientMode"] != "G10" {
		t.Errorf("ClientMode = %v, want G10", parsed["ClientMode"])
	}

	// Verify done channel was closed
	select {
	case <-done:
		// expected
	default:
		t.Error("done channel was not closed after successful finish")
	}
}

func TestWriteJSON(t *testing.T) {
	tests := []struct {
		name       string
		status     int
		payload    interface{}
		wantStatus int
	}{
		{
			name:       "OK with string map",
			status:     http.StatusOK,
			payload:    map[string]string{"key": "value"},
			wantStatus: http.StatusOK,
		},
		{
			name:       "BadRequest with error",
			status:     http.StatusBadRequest,
			payload:    map[string]string{"error": "bad input"},
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "InternalServerError",
			status:     http.StatusInternalServerError,
			payload:    map[string]string{"error": "something broke"},
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:       "OK with nested payload",
			status:     http.StatusOK,
			payload:    map[string]interface{}{"count": 42, "items": []string{"a", "b"}},
			wantStatus: http.StatusOK,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			writeJSON(w, tc.status, tc.payload)

			if w.Code != tc.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tc.wantStatus)
			}
			ct := w.Header().Get("Content-Type")
			if ct != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", ct)
			}
			// Verify body is valid JSON
			var decoded interface{}
			if err := json.NewDecoder(w.Body).Decode(&decoded); err != nil {
				t.Errorf("response body is not valid JSON: %v", err)
			}
		})
	}
}

func TestClientModesContainsExpected(t *testing.T) {
	modes := clientModes()
	expected := []string{"ZZ", "G10", "FW.4", "S1.0", "Z2", "GG"}
	modeSet := make(map[string]bool, len(modes))
	for _, m := range modes {
		modeSet[m] = true
	}
	for _, exp := range expected {
		if !modeSet[exp] {
			t.Errorf("clientModes() missing expected mode %q", exp)
		}
	}
}

func TestHandleInitDB_NoOps(t *testing.T) {
	ws := &wizardServer{
		logger: zap.NewNop(),
		done:   make(chan struct{}),
	}
	// All flags false — no DB operations, should succeed immediately.
	body := `{"host":"localhost","port":5432,"user":"test","password":"test","dbName":"test","createDB":false,"applySchema":false,"applyBundled":false}`
	req := httptest.NewRequest("POST", "/api/setup/init-db", strings.NewReader(body))
	w := httptest.NewRecorder()
	ws.handleInitDB(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var resp map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["success"] != true {
		t.Errorf("success = %v, want true", resp["success"])
	}
	log, ok := resp["log"].([]interface{})
	if !ok {
		t.Fatal("log should be an array")
	}
	// Should contain the "complete" message
	found := false
	for _, entry := range log {
		if s, ok := entry.(string); ok && strings.Contains(s, "complete") {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected completion message in log")
	}
}

func TestBuildDefaultConfig_WithLanguage(t *testing.T) {
	req := FinishRequest{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "postgres",
		DBPassword: "pass",
		DBName:     "erupe",
		Host:       "127.0.0.1",
		ClientMode: "ZZ",
		Language:   "en",
	}
	cfg := buildDefaultConfig(req)
	if cfg["Language"] != "en" {
		t.Errorf("Language = %v, want en", cfg["Language"])
	}
}

func TestWriteConfig_Permissions(t *testing.T) {
	dir := t.TempDir()
	origDir, _ := os.Getwd()
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(origDir) }()

	cfg := buildDefaultConfig(FinishRequest{
		DBHost:     "localhost",
		DBPort:     5432,
		DBUser:     "postgres",
		DBPassword: "pass",
		DBName:     "erupe",
		Host:       "127.0.0.1",
		ClientMode: "ZZ",
	})

	if err := writeConfig(cfg); err != nil {
		t.Fatalf("writeConfig failed: %v", err)
	}

	info, err := os.Stat(filepath.Join(dir, "config.json"))
	if err != nil {
		t.Fatalf("stat config.json: %v", err)
	}
	// File should be 0600 (owner read/write only)
	if perm := info.Mode().Perm(); perm != 0600 {
		t.Errorf("config.json permissions = %o, want 0600", perm)
	}
}
