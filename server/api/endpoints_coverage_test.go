package api

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cfg "erupe-ce/config"

	"golang.org/x/crypto/bcrypt"
)

func TestVersionEndpoint(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()
	c.ClientMode = "ZZ"

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
	}

	req := httptest.NewRequest("GET", "/version", nil)
	rec := httptest.NewRecorder()
	server.Version(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp VersionResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.ClientMode != "ZZ" {
		t.Errorf("ClientMode = %q, want ZZ", resp.ClientMode)
	}
	if resp.Name != "Erupe-CE" {
		t.Errorf("Name = %q, want Erupe-CE", resp.Name)
	}
}

func TestLandingPageEndpoint_Enabled(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()
	c.API.LandingPage = cfg.LandingPage{
		Enabled: true,
		Title:   "Test Server",
		Content: "<p>Welcome</p>",
	}

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
	}

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	server.LandingPage(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/html; charset=utf-8" {
		t.Errorf("Content-Type = %q", ct)
	}
}

func TestLandingPageEndpoint_Disabled(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()
	c.API.LandingPage = cfg.LandingPage{Enabled: false}

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
	}

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	server.LandingPage(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want 404", rec.Code)
	}
}

func TestLoginEndpoint_Success(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	hash, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.MinCost)

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		userRepo: &mockAPIUserRepo{
			credentialsID:       1,
			credentialsPassword: string(hash),
			credentialsRights:   30,
			lastLogin:           time.Now(),
			returnExpiry:        time.Now().Add(time.Hour * 24 * 30),
		},
		sessionRepo: &mockAPISessionRepo{
			createTokenID: 42,
		},
		charRepo: &mockAPICharacterRepo{
			characters: []Character{
				{ID: 1, Name: "TestHunter", HR: 100},
			},
		},
	}

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.Login(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp AuthData
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.User.TokenID != 42 {
		t.Errorf("TokenID = %d, want 42", resp.User.TokenID)
	}
	if len(resp.Characters) != 1 {
		t.Errorf("Characters count = %d, want 1", len(resp.Characters))
	}
}

func TestLoginEndpoint_UsernameNotFound(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		userRepo: &mockAPIUserRepo{
			credentialsErr: sql.ErrNoRows,
		},
	}

	body, _ := json.Marshal(map[string]string{
		"username": "nonexistent",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if rec.Body.String() != "username-error" {
		t.Errorf("body = %q, want username-error", rec.Body.String())
	}
}

func TestLoginEndpoint_WrongPassword(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	hash, _ := bcrypt.GenerateFromPassword([]byte("correct"), bcrypt.MinCost)

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		userRepo: &mockAPIUserRepo{
			credentialsID:       1,
			credentialsPassword: string(hash),
			lastLogin:           time.Now(),
			returnExpiry:        time.Now().Add(time.Hour * 24 * 30),
		},
	}

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "wrongpassword",
	})
	req := httptest.NewRequest("POST", "/login", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
	if rec.Body.String() != "password-error" {
		t.Errorf("body = %q, want password-error", rec.Body.String())
	}
}

func TestRegisterEndpoint_Success(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		userRepo: &mockAPIUserRepo{
			registerID:     1,
			registerRights: 30,
			lastLogin:      time.Now(),
			returnExpiry:   time.Now().Add(time.Hour * 24 * 30),
		},
		sessionRepo: &mockAPISessionRepo{
			createTokenID: 10,
		},
		charRepo: &mockAPICharacterRepo{},
	}

	body, _ := json.Marshal(map[string]string{
		"username": "newuser",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.Register(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp AuthData
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.User.Rights != 30 {
		t.Errorf("Rights = %d, want 30", resp.User.Rights)
	}
}

func TestCreateCharacterEndpoint_Success(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userID: 1,
		},
		charRepo: &mockAPICharacterRepo{
			newCharacter: Character{ID: 5, Name: "NewChar"},
		},
	}

	body, _ := json.Marshal(map[string]string{
		"token": "valid-token",
	})
	req := httptest.NewRequest("POST", "/character/create", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.CreateCharacter(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestCreateCharacterEndpoint_InvalidToken(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userIDErr: sql.ErrNoRows,
		},
	}

	body, _ := json.Marshal(map[string]string{
		"token": "invalid",
	})
	req := httptest.NewRequest("POST", "/character/create", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.CreateCharacter(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestDeleteCharacterEndpoint_NewChar(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userID: 1,
		},
		charRepo: &mockAPICharacterRepo{
			isNewResult: true,
		},
	}

	body, _ := json.Marshal(map[string]interface{}{
		"token":  "valid-token",
		"charId": 5,
	})
	req := httptest.NewRequest("POST", "/character/delete", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.DeleteCharacter(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestDeleteCharacterEndpoint_FinalizedChar(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userID: 1,
		},
		charRepo: &mockAPICharacterRepo{
			isNewResult: false,
		},
	}

	body, _ := json.Marshal(map[string]interface{}{
		"token":  "valid-token",
		"charId": 5,
	})
	req := httptest.NewRequest("POST", "/character/delete", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.DeleteCharacter(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
}

func TestExportSaveEndpoint_Success(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userID: 1,
		},
		charRepo: &mockAPICharacterRepo{
			exportResult: map[string]interface{}{
				"name": "TestHunter",
				"hr":   100,
			},
		},
	}

	body, _ := json.Marshal(map[string]interface{}{
		"token":  "valid-token",
		"charId": 1,
	})
	req := httptest.NewRequest("POST", "/character/export", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.ExportSave(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}

	var resp ExportData
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp.Character["name"] != "TestHunter" {
		t.Errorf("character name = %v, want TestHunter", resp.Character["name"])
	}
}
