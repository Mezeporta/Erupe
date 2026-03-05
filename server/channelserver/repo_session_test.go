package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupSessionRepo(t *testing.T) (*SessionRepository, *sqlx.DB, uint32, uint32, uint32, string) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "session_test_user")
	charID := CreateTestCharacter(t, db, userID, "SessionChar")
	token := "test_token_12345"
	sessionID := CreateTestSignSession(t, db, userID, token)
	repo := NewSessionRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, userID, charID, sessionID, token
}

func TestRepoSessionValidateLoginToken(t *testing.T) {
	repo, _, _, charID, sessionID, token := setupSessionRepo(t)

	err := repo.ValidateLoginToken(token, sessionID, charID)
	if err != nil {
		t.Fatalf("ValidateLoginToken failed: %v", err)
	}
}

func TestRepoSessionValidateLoginTokenInvalidToken(t *testing.T) {
	repo, _, _, charID, sessionID, _ := setupSessionRepo(t)

	err := repo.ValidateLoginToken("wrong_token", sessionID, charID)
	if err == nil {
		t.Fatal("Expected error for invalid token, got nil")
	}
}

func TestRepoSessionValidateLoginTokenWrongChar(t *testing.T) {
	repo, _, _, _, sessionID, token := setupSessionRepo(t)

	err := repo.ValidateLoginToken(token, sessionID, 999999)
	if err == nil {
		t.Fatal("Expected error for wrong char ID, got nil")
	}
}

func TestRepoSessionValidateLoginTokenWrongSession(t *testing.T) {
	repo, _, _, charID, _, token := setupSessionRepo(t)

	err := repo.ValidateLoginToken(token, 999999, charID)
	if err == nil {
		t.Fatal("Expected error for wrong session ID, got nil")
	}
}

func TestRepoSessionBindSession(t *testing.T) {
	repo, db, _, charID, _, token := setupSessionRepo(t)

	CreateTestServer(t, db, 1)

	if err := repo.BindSession(token, 1, charID); err != nil {
		t.Fatalf("BindSession failed: %v", err)
	}

	var serverID *uint16
	var boundCharID *uint32
	if err := db.QueryRow("SELECT server_id, char_id FROM sign_sessions WHERE token=$1", token).Scan(&serverID, &boundCharID); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if serverID == nil || *serverID != 1 {
		t.Errorf("Expected server_id=1, got: %v", serverID)
	}
	if boundCharID == nil || *boundCharID != charID {
		t.Errorf("Expected char_id=%d, got: %v", charID, boundCharID)
	}
}

func TestRepoSessionClearSession(t *testing.T) {
	repo, db, _, charID, _, token := setupSessionRepo(t)

	CreateTestServer(t, db, 1)

	if err := repo.BindSession(token, 1, charID); err != nil {
		t.Fatalf("BindSession failed: %v", err)
	}

	if err := repo.ClearSession(token); err != nil {
		t.Fatalf("ClearSession failed: %v", err)
	}

	var serverID, boundCharID *int
	if err := db.QueryRow("SELECT server_id, char_id FROM sign_sessions WHERE token=$1", token).Scan(&serverID, &boundCharID); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if serverID != nil {
		t.Errorf("Expected server_id=NULL, got: %v", *serverID)
	}
	if boundCharID != nil {
		t.Errorf("Expected char_id=NULL, got: %v", *boundCharID)
	}
}

func TestRepoSessionUpdatePlayerCount(t *testing.T) {
	repo, db, _, _, _, _ := setupSessionRepo(t)

	CreateTestServer(t, db, 1)

	if err := repo.UpdatePlayerCount(1, 42); err != nil {
		t.Fatalf("UpdatePlayerCount failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT current_players FROM servers WHERE server_id=1").Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 42 {
		t.Errorf("Expected current_players=42, got: %d", count)
	}
}

func TestRepoSessionUpdatePlayerCountTwice(t *testing.T) {
	repo, db, _, _, _, _ := setupSessionRepo(t)

	CreateTestServer(t, db, 1)

	if err := repo.UpdatePlayerCount(1, 10); err != nil {
		t.Fatalf("First UpdatePlayerCount failed: %v", err)
	}
	if err := repo.UpdatePlayerCount(1, 25); err != nil {
		t.Fatalf("Second UpdatePlayerCount failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT current_players FROM servers WHERE server_id=1").Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 25 {
		t.Errorf("Expected current_players=25, got: %d", count)
	}
}
