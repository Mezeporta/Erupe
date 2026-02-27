package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
	}

	handler := server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
	var errResp ErrorResponse
	if err := json.NewDecoder(rec.Body).Decode(&errResp); err != nil {
		t.Fatalf("decode error: %v", err)
	}
	if errResp.Error != "unauthorized" {
		t.Errorf("error = %q, want unauthorized", errResp.Error)
	}
}

func TestAuthMiddleware_MalformedHeader(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
	}

	handler := server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Basic dXNlcjpwYXNz")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		sessionRepo: &mockAPISessionRepo{
			userIDErr: sql.ErrNoRows,
		},
	}

	handler := server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want 401", rec.Code)
	}
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	logger := NewTestLogger(t)
	server := &APIServer{
		logger:      logger,
		erupeConfig: NewTestConfig(),
		sessionRepo: &mockAPISessionRepo{
			userID: 42,
		},
	}

	var gotUserID uint32
	var gotOK bool
	handler := server.AuthMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUserID, gotOK = UserIDFromContext(r.Context())
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("POST", "/test", nil)
	req.Header.Set("Authorization", "Bearer valid-token")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want 200", rec.Code)
	}
	if !gotOK {
		t.Fatal("userID not found in context")
	}
	if gotUserID != 42 {
		t.Errorf("userID = %d, want 42", gotUserID)
	}
}

func TestUserIDFromContext_Missing(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)
	uid, ok := UserIDFromContext(req.Context())
	if ok {
		t.Errorf("expected ok=false, got uid=%d", uid)
	}
}
