package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupGoocooRepo(t *testing.T) (*GoocooRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "goocoo_test_user")
	charID := CreateTestCharacter(t, db, userID, "GoocooChar")
	repo := NewGoocooRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func TestRepoGoocooEnsureExists(t *testing.T) {
	repo, db, charID := setupGoocooRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM goocoo WHERE id=$1", charID).Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 goocoo row, got: %d", count)
	}
}

func TestRepoGoocooEnsureExistsIdempotent(t *testing.T) {
	repo, _, charID := setupGoocooRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("First EnsureExists failed: %v", err)
	}
	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("Second EnsureExists failed: %v", err)
	}
}

func TestRepoGoocooSaveAndGetSlot(t *testing.T) {
	repo, _, charID := setupGoocooRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	data := []byte{0xAA, 0xBB, 0xCC}
	if err := repo.SaveSlot(charID, 0, data); err != nil {
		t.Fatalf("SaveSlot failed: %v", err)
	}

	got, err := repo.GetSlot(charID, 0)
	if err != nil {
		t.Fatalf("GetSlot failed: %v", err)
	}
	if len(got) != 3 || got[0] != 0xAA {
		t.Errorf("Expected saved data, got: %x", got)
	}
}

func TestRepoGoocooGetSlotNull(t *testing.T) {
	repo, _, charID := setupGoocooRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	got, err := repo.GetSlot(charID, 0)
	if err != nil {
		t.Fatalf("GetSlot failed: %v", err)
	}
	if got != nil {
		t.Errorf("Expected nil for NULL slot, got: %x", got)
	}
}

func TestRepoGoocooSaveMultipleSlots(t *testing.T) {
	repo, _, charID := setupGoocooRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	if err := repo.SaveSlot(charID, 0, []byte{0x01}); err != nil {
		t.Fatalf("SaveSlot(0) failed: %v", err)
	}
	if err := repo.SaveSlot(charID, 3, []byte{0x04}); err != nil {
		t.Fatalf("SaveSlot(3) failed: %v", err)
	}

	got0, _ := repo.GetSlot(charID, 0)
	got3, _ := repo.GetSlot(charID, 3)
	if len(got0) != 1 || got0[0] != 0x01 {
		t.Errorf("Slot 0 unexpected: %x", got0)
	}
	if len(got3) != 1 || got3[0] != 0x04 {
		t.Errorf("Slot 3 unexpected: %x", got3)
	}
}

func TestRepoGoococClearSlot(t *testing.T) {
	repo, _, charID := setupGoocooRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	if err := repo.SaveSlot(charID, 2, []byte{0xFF}); err != nil {
		t.Fatalf("SaveSlot failed: %v", err)
	}

	if err := repo.ClearSlot(charID, 2); err != nil {
		t.Fatalf("ClearSlot failed: %v", err)
	}

	got, err := repo.GetSlot(charID, 2)
	if err != nil {
		t.Fatalf("GetSlot failed: %v", err)
	}
	if got != nil {
		t.Errorf("Expected nil after ClearSlot, got: %x", got)
	}
}

func TestRepoGoocooInvalidSlot(t *testing.T) {
	repo, _, charID := setupGoocooRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	_, err := repo.GetSlot(charID, 5)
	if err == nil {
		t.Fatal("Expected error for invalid slot index 5")
	}

	err = repo.SaveSlot(charID, 5, []byte{0x00})
	if err == nil {
		t.Fatal("Expected error for SaveSlot with invalid slot index 5")
	}

	err = repo.ClearSlot(charID, 5)
	if err == nil {
		t.Fatal("Expected error for ClearSlot with invalid slot index 5")
	}
}
