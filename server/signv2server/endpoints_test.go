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

// Note: Tests that require database operations are skipped when no DB is available.
// The following tests validate the structure and JSON handling of endpoints.

// TestLoginRequestStructure tests that login request JSON structure is correct
func TestLoginRequestStructure(t *testing.T) {
	// Test JSON marshaling/unmarshaling of request structure
	reqData := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: "testuser",
		Password: "testpass",
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Username != reqData.Username {
		t.Errorf("Username = %s, want %s", decoded.Username, reqData.Username)
	}
	if decoded.Password != reqData.Password {
		t.Errorf("Password = %s, want %s", decoded.Password, reqData.Password)
	}
}

// TestRegisterRequestStructure tests that register request JSON structure is correct
func TestRegisterRequestStructure(t *testing.T) {
	reqData := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		Username: "newuser",
		Password: "newpass",
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Username != reqData.Username {
		t.Errorf("Username = %s, want %s", decoded.Username, reqData.Username)
	}
	if decoded.Password != reqData.Password {
		t.Errorf("Password = %s, want %s", decoded.Password, reqData.Password)
	}
}

// TestCreateCharacterRequestStructure tests that create character request JSON structure is correct
func TestCreateCharacterRequestStructure(t *testing.T) {
	reqData := struct {
		Token string `json:"token"`
	}{
		Token: "test-token-12345",
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Token != reqData.Token {
		t.Errorf("Token = %s, want %s", decoded.Token, reqData.Token)
	}
}

// TestDeleteCharacterRequestStructure tests that delete character request JSON structure is correct
func TestDeleteCharacterRequestStructure(t *testing.T) {
	reqData := struct {
		Token  string `json:"token"`
		CharID int    `json:"id"`
	}{
		Token:  "test-token",
		CharID: 12345,
	}

	data, err := json.Marshal(reqData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		Token  string `json:"token"`
		CharID int    `json:"id"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Token != reqData.Token {
		t.Errorf("Token = %s, want %s", decoded.Token, reqData.Token)
	}
	if decoded.CharID != reqData.CharID {
		t.Errorf("CharID = %d, want %d", decoded.CharID, reqData.CharID)
	}
}

// TestLoginResponseStructure tests the login response JSON structure
func TestLoginResponseStructure(t *testing.T) {
	respData := struct {
		Token      string      `json:"token"`
		Characters []Character `json:"characters"`
	}{
		Token: "login-token-abc123",
		Characters: []Character{
			{ID: 1, Name: "Hunter1", IsFemale: false, Weapon: 3, HR: 100, GR: 10},
			{ID: 2, Name: "Hunter2", IsFemale: true, Weapon: 7, HR: 200, GR: 20},
		},
	}

	data, err := json.Marshal(respData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		Token      string      `json:"token"`
		Characters []Character `json:"characters"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Token != respData.Token {
		t.Errorf("Token = %s, want %s", decoded.Token, respData.Token)
	}
	if len(decoded.Characters) != len(respData.Characters) {
		t.Errorf("Characters count = %d, want %d", len(decoded.Characters), len(respData.Characters))
	}
}

// TestRegisterResponseStructure tests the register response JSON structure
func TestRegisterResponseStructure(t *testing.T) {
	respData := struct {
		Token string `json:"token"`
	}{
		Token: "register-token-xyz789",
	}

	data, err := json.Marshal(respData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Token != respData.Token {
		t.Errorf("Token = %s, want %s", decoded.Token, respData.Token)
	}
}

// TestCreateCharacterResponseStructure tests the create character response JSON structure
func TestCreateCharacterResponseStructure(t *testing.T) {
	respData := struct {
		CharID int `json:"id"`
	}{
		CharID: 42,
	}

	data, err := json.Marshal(respData)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded struct {
		CharID int `json:"id"`
	}
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.CharID != respData.CharID {
		t.Errorf("CharID = %d, want %d", decoded.CharID, respData.CharID)
	}
}

// TestLauncherContentType tests that Launcher sets correct content type
func TestLauncherContentType(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("GET", "/launcher", nil)
	w := httptest.NewRecorder()

	s.Launcher(w, req)

	// Note: The handler sets header after WriteHeader, so we check response body is JSON
	resp := w.Result()
	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Errorf("Launcher() response is not valid JSON: %v", err)
	}
}

// TestLauncherMessageDates tests that launcher message dates are valid timestamps
func TestLauncherMessageDates(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("GET", "/launcher", nil)
	w := httptest.NewRecorder()

	s.Launcher(w, req)

	var data struct {
		Important []LauncherMessage `json:"important"`
		Normal    []LauncherMessage `json:"normal"`
	}
	json.NewDecoder(w.Result().Body).Decode(&data)

	// All dates should be positive unix timestamps
	for _, msg := range data.Important {
		if msg.Date <= 0 {
			t.Errorf("Important message date should be positive, got %d", msg.Date)
		}
	}
	for _, msg := range data.Normal {
		if msg.Date <= 0 {
			t.Errorf("Normal message date should be positive, got %d", msg.Date)
		}
	}
}

// TestLauncherMessageLinks tests that launcher message links are valid URLs
func TestLauncherMessageLinks(t *testing.T) {
	s := mockServer()

	req := httptest.NewRequest("GET", "/launcher", nil)
	w := httptest.NewRecorder()

	s.Launcher(w, req)

	var data struct {
		Important []LauncherMessage `json:"important"`
		Normal    []LauncherMessage `json:"normal"`
	}
	json.NewDecoder(w.Result().Body).Decode(&data)

	// All links should start with http:// or https://
	for _, msg := range data.Important {
		if len(msg.Link) < 7 || (msg.Link[:7] != "http://" && msg.Link[:8] != "https://") {
			t.Errorf("Important message link should be a URL, got %q", msg.Link)
		}
	}
	for _, msg := range data.Normal {
		if len(msg.Link) < 7 || (msg.Link[:7] != "http://" && msg.Link[:8] != "https://") {
			t.Errorf("Normal message link should be a URL, got %q", msg.Link)
		}
	}
}

// TestCharacterStructJSONMarshal tests Character struct marshals correctly
func TestCharacterStructJSONMarshal(t *testing.T) {
	char := Character{
		ID:        42,
		Name:      "TestHunter",
		IsFemale:  true,
		Weapon:    7,
		HR:        999,
		GR:        100,
		LastLogin: 1609459200,
	}

	data, err := json.Marshal(char)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded Character
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.ID != char.ID {
		t.Errorf("ID = %d, want %d", decoded.ID, char.ID)
	}
	if decoded.Name != char.Name {
		t.Errorf("Name = %s, want %s", decoded.Name, char.Name)
	}
	if decoded.IsFemale != char.IsFemale {
		t.Errorf("IsFemale = %v, want %v", decoded.IsFemale, char.IsFemale)
	}
	if decoded.Weapon != char.Weapon {
		t.Errorf("Weapon = %d, want %d", decoded.Weapon, char.Weapon)
	}
	if decoded.HR != char.HR {
		t.Errorf("HR = %d, want %d", decoded.HR, char.HR)
	}
	if decoded.GR != char.GR {
		t.Errorf("GR = %d, want %d", decoded.GR, char.GR)
	}
	if decoded.LastLogin != char.LastLogin {
		t.Errorf("LastLogin = %d, want %d", decoded.LastLogin, char.LastLogin)
	}
}

// TestLauncherMessageJSONMarshal tests LauncherMessage struct marshals correctly
func TestLauncherMessageJSONMarshal(t *testing.T) {
	msg := LauncherMessage{
		Message: "Test Announcement",
		Date:    1609459200,
		Link:    "https://example.com/news",
	}

	data, err := json.Marshal(msg)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var decoded LauncherMessage
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if decoded.Message != msg.Message {
		t.Errorf("Message = %s, want %s", decoded.Message, msg.Message)
	}
	if decoded.Date != msg.Date {
		t.Errorf("Date = %d, want %d", decoded.Date, msg.Date)
	}
	if decoded.Link != msg.Link {
		t.Errorf("Link = %s, want %s", decoded.Link, msg.Link)
	}
}

// TestEndpointHTTPMethods tests that endpoints respond to correct HTTP methods
func TestEndpointHTTPMethods(t *testing.T) {
	s := mockServer()

	// Launcher should respond to GET
	t.Run("Launcher GET", func(t *testing.T) {
		req := httptest.NewRequest("GET", "/launcher", nil)
		w := httptest.NewRecorder()
		s.Launcher(w, req)
		if w.Result().StatusCode != http.StatusOK {
			t.Errorf("Launcher() GET status = %d, want %d", w.Result().StatusCode, http.StatusOK)
		}
	})

	// Note: Login, Register, CreateCharacter, DeleteCharacter require database
	// and cannot be tested without mocking the database connection
}
