package api

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthEndpoint_NilDB(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		db:          nil,
	}

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	server.Health(rec, req)

	if rec.Code != http.StatusServiceUnavailable {
		t.Errorf("status = %d, want 503", rec.Code)
	}
	var resp map[string]string
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if resp["status"] != "unhealthy" {
		t.Errorf("status = %q, want unhealthy", resp["status"])
	}
}

func TestRegisterEndpoint_EmptyPassword(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
	}

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "",
	})
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestRegisterEndpoint_InvalidJSON(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
	}

	req := httptest.NewRequest("POST", "/register", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	server.Register(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestLoginEndpoint_InvalidJSON(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
	}

	req := httptest.NewRequest("POST", "/login", bytes.NewReader([]byte("not json")))
	rec := httptest.NewRecorder()
	server.Login(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestCreateCharacterEndpoint_InvalidJSON(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
	}

	req := httptest.NewRequest("POST", "/character/create", bytes.NewReader([]byte("bad")))
	rec := httptest.NewRecorder()
	server.CreateCharacter(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestDeleteCharacterEndpoint_InvalidJSON(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
	}

	req := httptest.NewRequest("POST", "/character/delete", bytes.NewReader([]byte("bad")))
	rec := httptest.NewRecorder()
	server.DeleteCharacter(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestExportSaveEndpoint_InvalidJSON(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
	}

	req := httptest.NewRequest("POST", "/character/export", bytes.NewReader([]byte("bad")))
	rec := httptest.NewRecorder()
	server.ExportSave(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want 400", rec.Code)
	}
}

func TestRegisterEndpoint_CreateUserError(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		userRepo: &mockAPIUserRepo{
			registerErr: errors.New("db connection failed"),
		},
	}

	body, _ := json.Marshal(map[string]string{
		"username": "testuser",
		"password": "password123",
	})
	req := httptest.NewRequest("POST", "/register", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.Register(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want 500", rec.Code)
	}
}

func TestDeleteCharacterEndpoint_InvalidToken(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userIDErr: errors.New("bad token"),
		},
	}

	body, _ := json.Marshal(map[string]interface{}{
		"token":  "invalid",
		"charId": 5,
	})
	req := httptest.NewRequest("POST", "/character/delete", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.DeleteCharacter(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestExportSaveEndpoint_InvalidToken(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userIDErr: errors.New("bad token"),
		},
	}

	body, _ := json.Marshal(map[string]interface{}{
		"token":  "invalid",
		"charId": 1,
	})
	req := httptest.NewRequest("POST", "/character/export", bytes.NewReader(body))
	rec := httptest.NewRecorder()
	server.ExportSave(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}
