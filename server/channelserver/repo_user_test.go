package channelserver

import (
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func setupUserRepo(t *testing.T) (*UserRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "user_repo_test")
	repo := NewUserRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, userID
}

func TestBanUserPermanent(t *testing.T) {
	repo, db, userID := setupUserRepo(t)

	if err := repo.BanUser(userID, nil); err != nil {
		t.Fatalf("BanUser (permanent) failed: %v", err)
	}

	// Verify ban exists with NULL expires
	var gotUserID uint32
	var expires sql.NullTime
	err := db.QueryRow("SELECT user_id, expires FROM bans WHERE user_id=$1", userID).Scan(&gotUserID, &expires)
	if err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("Expected user_id %d, got: %d", userID, gotUserID)
	}
	if expires.Valid {
		t.Errorf("Expected NULL expires for permanent ban, got: %v", expires.Time)
	}
}

func TestBanUserTemporary(t *testing.T) {
	repo, db, userID := setupUserRepo(t)

	expiry := time.Now().Add(24 * time.Hour).Truncate(time.Microsecond)
	if err := repo.BanUser(userID, &expiry); err != nil {
		t.Fatalf("BanUser (temporary) failed: %v", err)
	}

	var gotUserID uint32
	var got time.Time
	err := db.QueryRow("SELECT user_id, expires FROM bans WHERE user_id=$1", userID).Scan(&gotUserID, &got)
	if err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if gotUserID != userID {
		t.Errorf("Expected user_id %d, got: %d", userID, gotUserID)
	}
	if !got.Equal(expiry) {
		t.Errorf("Expected expires %v, got: %v", expiry, got)
	}
}

func TestBanUserUpsertPermanentToTemporary(t *testing.T) {
	repo, db, userID := setupUserRepo(t)

	// First: permanent ban
	if err := repo.BanUser(userID, nil); err != nil {
		t.Fatalf("BanUser (permanent) failed: %v", err)
	}

	// Upsert: change to temporary
	expiry := time.Now().Add(1 * time.Hour).Truncate(time.Microsecond)
	if err := repo.BanUser(userID, &expiry); err != nil {
		t.Fatalf("BanUser (upsert to temporary) failed: %v", err)
	}

	var got time.Time
	err := db.QueryRow("SELECT expires FROM bans WHERE user_id=$1", userID).Scan(&got)
	if err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !got.Equal(expiry) {
		t.Errorf("Expected expires %v after upsert, got: %v", expiry, got)
	}
}

func TestBanUserUpsertTemporaryToPermanent(t *testing.T) {
	repo, db, userID := setupUserRepo(t)

	// First: temporary ban
	expiry := time.Now().Add(1 * time.Hour).Truncate(time.Microsecond)
	if err := repo.BanUser(userID, &expiry); err != nil {
		t.Fatalf("BanUser (temporary) failed: %v", err)
	}

	// Upsert: change to permanent
	if err := repo.BanUser(userID, nil); err != nil {
		t.Fatalf("BanUser (upsert to permanent) failed: %v", err)
	}

	var expires sql.NullTime
	err := db.QueryRow("SELECT expires FROM bans WHERE user_id=$1", userID).Scan(&expires)
	if err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if expires.Valid {
		t.Errorf("Expected NULL expires after upsert to permanent, got: %v", expires.Time)
	}
}
