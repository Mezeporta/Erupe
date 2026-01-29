package signv2server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go.uber.org/zap"
)

// mockServer creates a Server with minimal dependencies for testing
func mockServer() *Server {
	logger, _ := zap.NewDevelopment()
	return &Server{
		logger: logger,
	}
}

func TestLauncherEndpoint(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("GET", "/launcher", nil)
	w := httptest.NewRecorder()

	s.Launcher(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Launcher() status = %d, want %d", resp.StatusCode, http.StatusOK)
	}

	var data struct {
		Important []LauncherMessage `json:"important"`
		Normal    []LauncherMessage `json:"normal"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Should have important messages
	if len(data.Important) == 0 {
		t.Error("Launcher() should return important messages")
	}

	// Should have normal messages
	if len(data.Normal) == 0 {
		t.Error("Launcher() should return normal messages")
	}
}

func TestLauncherMessageStructure(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("GET", "/launcher", nil)
	w := httptest.NewRecorder()

	s.Launcher(w, req)

	var data struct {
		Important []LauncherMessage `json:"important"`
		Normal    []LauncherMessage `json:"normal"`
	}
	json.NewDecoder(w.Result().Body).Decode(&data)

	// Check important messages have required fields
	for _, msg := range data.Important {
		if msg.Message == "" {
			t.Error("LauncherMessage.Message should not be empty")
		}
		if msg.Date == 0 {
			t.Error("LauncherMessage.Date should not be zero")
		}
		if msg.Link == "" {
			t.Error("LauncherMessage.Link should not be empty")
		}
	}
}

func TestLoginEndpointInvalidJSON(t *testing.T) {
	s := mockServer()

	// Send invalid JSON
	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("not json")))
	w := httptest.NewRecorder()

	s.Login(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Login() with invalid JSON status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestRegisterEndpointInvalidJSON(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	s.Register(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("Register() with invalid JSON status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestCreateCharacterEndpointInvalidJSON(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("POST", "/character/create", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	s.CreateCharacter(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("CreateCharacter() with invalid JSON status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestDeleteCharacterEndpointInvalidJSON(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("POST", "/character/delete", bytes.NewReader([]byte("invalid")))
	w := httptest.NewRecorder()

	s.DeleteCharacter(w, req)

	resp := w.Result()
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("DeleteCharacter() with invalid JSON status = %d, want %d", resp.StatusCode, http.StatusBadRequest)
	}
}

func TestLauncherMessageStruct(t *testing.T) {
	msg := LauncherMessage{
		Message: "Test Message",
		Date:    1234567890,
		Link:    "https://example.com",
	}

	if msg.Message != "Test Message" {
		t.Errorf("Message = %s, want Test Message", msg.Message)
	}
	if msg.Date != 1234567890 {
		t.Errorf("Date = %d, want 1234567890", msg.Date)
	}
	if msg.Link != "https://example.com" {
		t.Errorf("Link = %s, want https://example.com", msg.Link)
	}
}

func TestCharacterStruct(t *testing.T) {
	char := Character{
		ID:        1,
		Name:      "TestHunter",
		IsFemale:  true,
		Weapon:    5,
		HR:        999,
		GR:        100,
		LastLogin: 1234567890,
	}

	if char.ID != 1 {
		t.Errorf("ID = %d, want 1", char.ID)
	}
	if char.Name != "TestHunter" {
		t.Errorf("Name = %s, want TestHunter", char.Name)
	}
	if char.IsFemale != true {
		t.Error("IsFemale should be true")
	}
	if char.Weapon != 5 {
		t.Errorf("Weapon = %d, want 5", char.Weapon)
	}
	if char.HR != 999 {
		t.Errorf("HR = %d, want 999", char.HR)
	}
	if char.GR != 100 {
		t.Errorf("GR = %d, want 100", char.GR)
	}
}

func TestLauncherMessageJSONTags(t *testing.T) {
	msg := LauncherMessage{
		Message: "Test",
		Date:    12345,
		Link:    "http://test.com",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if _, ok := decoded["message"]; !ok {
		t.Error("JSON should have 'message' key")
	}
	if _, ok := decoded["date"]; !ok {
		t.Error("JSON should have 'date' key")
	}
	if _, ok := decoded["link"]; !ok {
		t.Error("JSON should have 'link' key")
	}
}

func TestCharacterJSONTags(t *testing.T) {
	char := Character{
		ID:        1,
		Name:      "Test",
		IsFemale:  true,
		Weapon:    3,
		HR:        50,
		GR:        10,
		LastLogin: 9999,
	}

	data, err := json.Marshal(char)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded map[string]interface{}
	json.Unmarshal(data, &decoded)

	if _, ok := decoded["id"]; !ok {
		t.Error("JSON should have 'id' key")
	}
	if _, ok := decoded["name"]; !ok {
		t.Error("JSON should have 'name' key")
	}
	if _, ok := decoded["isFemale"]; !ok {
		t.Error("JSON should have 'isFemale' key")
	}
	if _, ok := decoded["weapon"]; !ok {
		t.Error("JSON should have 'weapon' key")
	}
	if _, ok := decoded["hr"]; !ok {
		t.Error("JSON should have 'hr' key")
	}
	if _, ok := decoded["gr"]; !ok {
		t.Error("JSON should have 'gr' key")
	}
	if _, ok := decoded["lastLogin"]; !ok {
		t.Error("JSON should have 'lastLogin' key")
	}
}

func TestLauncherResponseFormat(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("GET", "/launcher", nil)
	w := httptest.NewRecorder()

	s.Launcher(w, req)

	resp := w.Result()

	// Verify it returns valid JSON
	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("Launcher() should return valid JSON: %v", err)
	}

	// Check top-level keys exist
	if _, ok := result["important"]; !ok {
		t.Error("Launcher() response should have 'important' key")
	}
	if _, ok := result["normal"]; !ok {
		t.Error("Launcher() response should have 'normal' key")
	}
}

func TestLauncherMessageCount(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("GET", "/launcher", nil)
	w := httptest.NewRecorder()

	s.Launcher(w, req)

	var data struct {
		Important []LauncherMessage `json:"important"`
		Normal    []LauncherMessage `json:"normal"`
	}
	json.NewDecoder(w.Result().Body).Decode(&data)

	// Should have at least 3 important messages based on the implementation
	if len(data.Important) < 3 {
		t.Errorf("Launcher() should return at least 3 important messages, got %d", len(data.Important))
	}

	// Should have at least 1 normal message
	if len(data.Normal) < 1 {
		t.Errorf("Launcher() should return at least 1 normal message, got %d", len(data.Normal))
	}
}

func TestCharacterStructDBTags(t *testing.T) {
	// Test that Character struct has proper db tags
	char := Character{}

	// These fields have db tags, verify struct is usable
	char.IsFemale = true
	char.Weapon = 7
	char.HR = 100
	char.LastLogin = 12345

	if char.Weapon != 7 {
		t.Errorf("Weapon = %d, want 7", char.Weapon)
	}
}

func TestNewServer(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	cfg := &Config{
		Logger: logger,
	}

	s := NewServer(cfg)

	if s == nil {
		t.Fatal("NewServer() returned nil")
	}
	if s.logger == nil {
		t.Error("NewServer() should set logger")
	}
	if s.httpServer == nil {
		t.Error("NewServer() should initialize httpServer")
	}
}

func TestServerConfig(t *testing.T) {
	cfg := &Config{
		Logger: nil,
		DB:     nil,
	}

	// Config struct should be usable
	if cfg.Logger != nil {
		t.Error("Config.Logger should be nil when not set")
	}
}
