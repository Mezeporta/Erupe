package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupCharRepo(t *testing.T) (*CharacterRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "repo_test_user")
	charID := CreateTestCharacter(t, db, userID, "RepoChar")
	repo := NewCharacterRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func TestLoadColumn(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	// Write a known blob to a column
	blob := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	_, err := db.Exec("UPDATE characters SET kouryou_point=$1 WHERE id=$2", blob, charID)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	data, err := repo.LoadColumn(charID, "kouryou_point")
	if err != nil {
		t.Fatalf("LoadColumn failed: %v", err)
	}
	if len(data) != 4 || data[0] != 0xDE || data[3] != 0xEF {
		t.Errorf("LoadColumn returned unexpected data: %x", data)
	}
}

func TestLoadColumnNil(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	// Column should be NULL by default
	data, err := repo.LoadColumn(charID, "kouryou_point")
	if err != nil {
		t.Fatalf("LoadColumn failed: %v", err)
	}
	if data != nil {
		t.Errorf("Expected nil for NULL column, got: %x", data)
	}
}

func TestSaveColumn(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	blob := []byte{0x01, 0x02, 0x03}
	if err := repo.SaveColumn(charID, "kouryou_point", blob); err != nil {
		t.Fatalf("SaveColumn failed: %v", err)
	}

	// Verify via direct SELECT
	var got []byte
	if err := db.QueryRow("SELECT kouryou_point FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if len(got) != 3 || got[0] != 0x01 || got[2] != 0x03 {
		t.Errorf("SaveColumn wrote unexpected data: %x", got)
	}
}

func TestReadInt(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	// time_played defaults to 0 via COALESCE
	val, err := repo.ReadInt(charID, "time_played")
	if err != nil {
		t.Fatalf("ReadInt failed: %v", err)
	}
	if val != 0 {
		t.Errorf("Expected 0 for default time_played, got: %d", val)
	}
}

func TestReadIntWithValue(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	_, err := db.Exec("UPDATE characters SET time_played=42 WHERE id=$1", charID)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	val, err := repo.ReadInt(charID, "time_played")
	if err != nil {
		t.Fatalf("ReadInt failed: %v", err)
	}
	if val != 42 {
		t.Errorf("Expected 42, got: %d", val)
	}
}

func TestAdjustInt(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	// First adjustment from NULL (COALESCE makes it 0 + 10 = 10)
	val, err := repo.AdjustInt(charID, "time_played", 10)
	if err != nil {
		t.Fatalf("AdjustInt failed: %v", err)
	}
	if val != 10 {
		t.Errorf("Expected 10 after first adjust, got: %d", val)
	}

	// Second adjustment: 10 + 5 = 15
	val, err = repo.AdjustInt(charID, "time_played", 5)
	if err != nil {
		t.Fatalf("AdjustInt failed: %v", err)
	}
	if val != 15 {
		t.Errorf("Expected 15 after second adjust, got: %d", val)
	}
}

func TestGetName(t *testing.T) {
	repo, _, charID := setupCharRepo(t)

	name, err := repo.GetName(charID)
	if err != nil {
		t.Fatalf("GetName failed: %v", err)
	}
	if name != "RepoChar" {
		t.Errorf("Expected 'RepoChar', got: %q", name)
	}
}

func TestGetUserID(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	// Look up the expected user_id
	var expectedUID uint32
	if err := db.QueryRow("SELECT user_id FROM characters WHERE id=$1", charID).Scan(&expectedUID); err != nil {
		t.Fatalf("Setup query failed: %v", err)
	}

	uid, err := repo.GetUserID(charID)
	if err != nil {
		t.Fatalf("GetUserID failed: %v", err)
	}
	if uid != expectedUID {
		t.Errorf("Expected user_id %d, got: %d", expectedUID, uid)
	}
}

func TestUpdateLastLogin(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	ts := int64(1700000000)
	if err := repo.UpdateLastLogin(charID, ts); err != nil {
		t.Fatalf("UpdateLastLogin failed: %v", err)
	}

	var got int64
	if err := db.QueryRow("SELECT last_login FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if got != ts {
		t.Errorf("Expected last_login %d, got: %d", ts, got)
	}
}

func TestUpdateTimePlayed(t *testing.T) {
	repo, db, charID := setupCharRepo(t)

	if err := repo.UpdateTimePlayed(charID, 999); err != nil {
		t.Fatalf("UpdateTimePlayed failed: %v", err)
	}

	var got int
	if err := db.QueryRow("SELECT time_played FROM characters WHERE id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if got != 999 {
		t.Errorf("Expected time_played 999, got: %d", got)
	}
}

func TestGetCharIDsByUserID(t *testing.T) {
	repo, db, _ := setupCharRepo(t)

	// Create a second user with multiple characters
	uid2 := CreateTestUser(t, db, "multi_char_user")
	cid1 := CreateTestCharacter(t, db, uid2, "Char1")
	cid2 := CreateTestCharacter(t, db, uid2, "Char2")

	ids, err := repo.GetCharIDsByUserID(uid2)
	if err != nil {
		t.Fatalf("GetCharIDsByUserID failed: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("Expected 2 character IDs, got: %d", len(ids))
	}

	// Check both IDs are present (order may vary)
	found := map[uint32]bool{cid1: false, cid2: false}
	for _, id := range ids {
		found[id] = true
	}
	if !found[cid1] || !found[cid2] {
		t.Errorf("Expected IDs %d and %d, got: %v", cid1, cid2, ids)
	}
}

func TestGetCharIDsByUserIDEmpty(t *testing.T) {
	repo, db, _ := setupCharRepo(t)

	// Create a user with no characters
	uid := CreateTestUser(t, db, "no_chars_user")

	ids, err := repo.GetCharIDsByUserID(uid)
	if err != nil {
		t.Fatalf("GetCharIDsByUserID failed: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("Expected 0 character IDs for user with no chars, got: %d", len(ids))
	}
}
