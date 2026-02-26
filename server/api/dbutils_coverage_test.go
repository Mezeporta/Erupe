package api

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"
)

func TestCreateNewUser_Success(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		userRepo: &mockAPIUserRepo{
			registerID:     1,
			registerRights: 30,
		},
	}

	uid, rights, err := server.createNewUser(context.Background(), "testuser", "password123")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != 1 {
		t.Errorf("uid = %d, want 1", uid)
	}
	if rights != 30 {
		t.Errorf("rights = %d, want 30", rights)
	}
}

func TestCreateLoginToken_Success(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			createTokenID: 42,
		},
	}

	tid, token, err := server.createLoginToken(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if tid != 42 {
		t.Errorf("tid = %d, want 42", tid)
	}
	if token == "" {
		t.Error("token should not be empty")
	}
}

func TestCreateLoginToken_Error(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			createTokenErr: errors.New("db error"),
		},
	}

	_, _, err := server.createLoginToken(context.Background(), 1)
	if err == nil {
		t.Error("expected error")
	}
}

func TestUserIDFromToken_Valid(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userID: 42,
		},
	}

	uid, err := server.userIDFromToken(context.Background(), "valid-token")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if uid != 42 {
		t.Errorf("uid = %d, want 42", uid)
	}
}

func TestUserIDFromToken_ErrNoRows(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userIDErr: sql.ErrNoRows,
		},
	}

	_, err := server.userIDFromToken(context.Background(), "invalid-token")
	if err == nil {
		t.Error("expected error for invalid token")
	}
}

func TestUserIDFromToken_OtherError(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		sessionRepo: &mockAPISessionRepo{
			userIDErr: errors.New("connection refused"),
		},
	}

	_, err := server.userIDFromToken(context.Background(), "some-token")
	if err == nil {
		t.Error("expected error")
	}
}

func TestCreateCharacter_ExistingNew(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		charRepo: &mockAPICharacterRepo{
			newCharacter: Character{ID: 5, Name: "NewChar"},
		},
	}

	char, err := server.createCharacter(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if char.ID != 5 {
		t.Errorf("char ID = %d, want 5", char.ID)
	}
}

func TestCreateCharacter_CreateNew(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		charRepo: &mockAPICharacterRepo{
			newCharacterErr: sql.ErrNoRows,
			countForUser:    2,
			createChar:      Character{ID: 10, Name: "Created"},
		},
	}

	char, err := server.createCharacter(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if char.ID != 10 {
		t.Errorf("char ID = %d, want 10", char.ID)
	}
}

func TestCreateCharacter_MaxExceeded(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		charRepo: &mockAPICharacterRepo{
			newCharacterErr: sql.ErrNoRows,
			countForUser:    16,
			createCharErr:   errors.New("cannot have more than 16 characters"),
		},
	}

	_, err := server.createCharacter(context.Background(), 1)
	if err == nil {
		t.Error("expected error for max chars exceeded")
	}
}

func TestDeleteCharacter_IsNewHardDelete(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	charRepo := &mockAPICharacterRepo{isNewResult: true}
	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		charRepo:    charRepo,
	}

	err := server.deleteCharacter(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeleteCharacter_FinalizedSoftDelete(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	charRepo := &mockAPICharacterRepo{isNewResult: false}
	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		charRepo:    charRepo,
	}

	err := server.deleteCharacter(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGetCharactersForUser(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		charRepo: &mockAPICharacterRepo{
			characters: []Character{
				{ID: 1, Name: "Char1"},
				{ID: 2, Name: "Char2"},
			},
		},
	}

	chars, err := server.getCharactersForUser(context.Background(), 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(chars) != 2 {
		t.Errorf("count = %d, want 2", len(chars))
	}
}

func TestExportSave(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		charRepo: &mockAPICharacterRepo{
			exportResult: map[string]interface{}{"name": "Hunter"},
		},
	}

	result, err := server.exportSave(context.Background(), 1, 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["name"] != "Hunter" {
		t.Errorf("name = %v, want Hunter", result["name"])
	}
}

func TestGetReturnExpiry_RecentLogin(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		userRepo: &mockAPIUserRepo{
			lastLogin:    time.Now(),
			returnExpiry: time.Now().Add(time.Hour * 24 * 15),
		},
	}

	expiry := server.getReturnExpiry(1)
	if expiry.IsZero() {
		t.Error("expiry should not be zero")
	}
}

func TestGetReturnExpiry_OldLogin(t *testing.T) {
	logger := NewTestLogger(t)
	c := NewTestConfig()

	server := &APIServer{
		logger:      logger,
		erupeConfig: c,
		userRepo: &mockAPIUserRepo{
			lastLogin:    time.Now().Add(-time.Hour * 24 * 100), // 100 days ago
			returnExpiry: time.Now().Add(time.Hour * 24 * 30),
		},
	}

	expiry := server.getReturnExpiry(1)
	if expiry.Before(time.Now()) {
		t.Error("expiry should be in the future for returning player")
	}
}
