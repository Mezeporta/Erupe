package channelserver

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupGachaRepo(t *testing.T) (*GachaRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "gacha_test_user")
	charID := CreateTestCharacter(t, db, userID, "GachaChar")
	repo := NewGachaRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func TestRepoGachaListShopEmpty(t *testing.T) {
	repo, _, _ := setupGachaRepo(t)

	shops, err := repo.ListShop()
	if err != nil {
		t.Fatalf("ListShop failed: %v", err)
	}
	if len(shops) != 0 {
		t.Errorf("Expected empty shop list, got: %d", len(shops))
	}
}

func TestRepoGachaListShop(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	CreateTestGachaShop(t, db, "Test Gacha", 1)
	CreateTestGachaShop(t, db, "Premium Gacha", 2)

	shops, err := repo.ListShop()
	if err != nil {
		t.Fatalf("ListShop failed: %v", err)
	}
	if len(shops) != 2 {
		t.Fatalf("Expected 2 shops, got: %d", len(shops))
	}
}

func TestRepoGachaGetShopType(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Type Test", 3)

	gachaType, err := repo.GetShopType(shopID)
	if err != nil {
		t.Fatalf("GetShopType failed: %v", err)
	}
	if gachaType != 3 {
		t.Errorf("Expected gacha_type=3, got: %d", gachaType)
	}
}

func TestRepoGachaGetEntryForTransaction(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Entry Test", 1)
	_, err := db.Exec(
		`INSERT INTO gacha_entries (gacha_id, entry_type, weight, rarity, item_type, item_number, item_quantity, rolls, frontier_points, daily_limit)
		VALUES ($1, 5, 100, 1, 7, 500, 10, 3, 0, 0)`, shopID,
	)
	if err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	itemType, itemNumber, rolls, err := repo.GetEntryForTransaction(shopID, 5)
	if err != nil {
		t.Fatalf("GetEntryForTransaction failed: %v", err)
	}
	if itemType != 7 {
		t.Errorf("Expected itemType=7, got: %d", itemType)
	}
	if itemNumber != 500 {
		t.Errorf("Expected itemNumber=500, got: %d", itemNumber)
	}
	if rolls != 3 {
		t.Errorf("Expected rolls=3, got: %d", rolls)
	}
}

func TestRepoGachaGetRewardPoolEmpty(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Empty Pool", 1)

	entries, err := repo.GetRewardPool(shopID)
	if err != nil {
		t.Fatalf("GetRewardPool failed: %v", err)
	}
	if len(entries) != 0 {
		t.Errorf("Expected empty reward pool, got: %d", len(entries))
	}
}

func TestRepoGachaGetRewardPoolOrdering(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Pool Test", 1)
	// entry_type=100 is the reward pool
	CreateTestGachaEntry(t, db, shopID, 100, 50)
	CreateTestGachaEntry(t, db, shopID, 100, 200)
	CreateTestGachaEntry(t, db, shopID, 100, 100)
	// entry_type=5 should NOT appear in reward pool
	CreateTestGachaEntry(t, db, shopID, 5, 999)

	entries, err := repo.GetRewardPool(shopID)
	if err != nil {
		t.Fatalf("GetRewardPool failed: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("Expected 3 reward entries, got: %d", len(entries))
	}
	// Should be ordered by weight DESC
	if entries[0].Weight < entries[1].Weight || entries[1].Weight < entries[2].Weight {
		t.Errorf("Expected descending weight order, got: %v, %v, %v", entries[0].Weight, entries[1].Weight, entries[2].Weight)
	}
}

func TestRepoGachaGetItemsForEntry(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Items Test", 1)
	entryID := CreateTestGachaEntry(t, db, shopID, 100, 100)
	CreateTestGachaItem(t, db, entryID, 1, 100, 5)
	CreateTestGachaItem(t, db, entryID, 2, 200, 10)

	items, err := repo.GetItemsForEntry(entryID)
	if err != nil {
		t.Fatalf("GetItemsForEntry failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got: %d", len(items))
	}
}

func TestRepoGachaGetGuaranteedItems(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Guaranteed Test", 1)
	entryID := CreateTestGachaEntry(t, db, shopID, 10, 0)
	CreateTestGachaItem(t, db, entryID, 3, 300, 1)

	items, err := repo.GetGuaranteedItems(10, shopID)
	if err != nil {
		t.Fatalf("GetGuaranteedItems failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 guaranteed item, got: %d", len(items))
	}
	if items[0].ItemID != 300 {
		t.Errorf("Expected item_id=300, got: %d", items[0].ItemID)
	}
}

func TestRepoGachaGetAllEntries(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "All Entries", 1)
	CreateTestGachaEntry(t, db, shopID, 100, 50)
	CreateTestGachaEntry(t, db, shopID, 5, 200)

	entries, err := repo.GetAllEntries(shopID)
	if err != nil {
		t.Fatalf("GetAllEntries failed: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("Expected 2 entries, got: %d", len(entries))
	}
}

func TestRepoGachaGetWeightDivisorZero(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Zero Weight", 1)

	divisor, err := repo.GetWeightDivisor(shopID)
	if err != nil {
		t.Fatalf("GetWeightDivisor failed: %v", err)
	}
	if divisor != 0 {
		t.Errorf("Expected divisor=0 for empty, got: %f", divisor)
	}
}

func TestRepoGachaGetWeightDivisor(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Weight Test", 1)
	CreateTestGachaEntry(t, db, shopID, 100, 50000)
	CreateTestGachaEntry(t, db, shopID, 100, 50000)

	divisor, err := repo.GetWeightDivisor(shopID)
	if err != nil {
		t.Fatalf("GetWeightDivisor failed: %v", err)
	}
	// (50000 + 50000) / 100000 = 1.0
	if divisor != 1.0 {
		t.Errorf("Expected divisor=1.0, got: %f", divisor)
	}
}

func TestRepoGachaHasEntryTypeTrue(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "HasType Test", 1)
	CreateTestGachaEntry(t, db, shopID, 100, 50)

	has, err := repo.HasEntryType(shopID, 100)
	if err != nil {
		t.Fatalf("HasEntryType failed: %v", err)
	}
	if !has {
		t.Error("Expected HasEntryType=true for entry_type=100")
	}
}

func TestRepoGachaHasEntryTypeFalse(t *testing.T) {
	repo, db, _ := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "HasType False", 1)

	has, err := repo.HasEntryType(shopID, 100)
	if err != nil {
		t.Fatalf("HasEntryType failed: %v", err)
	}
	if has {
		t.Error("Expected HasEntryType=false for empty gacha")
	}
}

// Stepup tests

func TestRepoGachaStepupLifecycle(t *testing.T) {
	repo, db, charID := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Stepup Test", 1)

	// Insert stepup
	if err := repo.InsertStepup(shopID, 1, charID); err != nil {
		t.Fatalf("InsertStepup failed: %v", err)
	}

	// Get step
	step, err := repo.GetStepupStep(shopID, charID)
	if err != nil {
		t.Fatalf("GetStepupStep failed: %v", err)
	}
	if step != 1 {
		t.Errorf("Expected step=1, got: %d", step)
	}

	// Delete stepup
	if err := repo.DeleteStepup(shopID, charID); err != nil {
		t.Fatalf("DeleteStepup failed: %v", err)
	}

	// Get step should fail
	_, err = repo.GetStepupStep(shopID, charID)
	if err == nil {
		t.Fatal("Expected error after DeleteStepup, got nil")
	}
}

func TestRepoGachaGetStepupWithTime(t *testing.T) {
	repo, db, charID := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Stepup Time", 1)

	if err := repo.InsertStepup(shopID, 2, charID); err != nil {
		t.Fatalf("InsertStepup failed: %v", err)
	}

	step, createdAt, err := repo.GetStepupWithTime(shopID, charID)
	if err != nil {
		t.Fatalf("GetStepupWithTime failed: %v", err)
	}
	if step != 2 {
		t.Errorf("Expected step=2, got: %d", step)
	}
	if createdAt.IsZero() {
		t.Error("Expected non-zero created_at")
	}
}

func TestRepoGachaGetStepupWithTimeNotFound(t *testing.T) {
	repo, db, charID := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Stepup NF", 1)

	_, _, err := repo.GetStepupWithTime(shopID, charID)
	if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("Expected sql.ErrNoRows, got: %v", err)
	}
}

// Box gacha tests

func TestRepoGachaBoxLifecycle(t *testing.T) {
	repo, db, charID := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Box Test", 1)
	entryID1 := CreateTestGachaEntry(t, db, shopID, 100, 50)
	entryID2 := CreateTestGachaEntry(t, db, shopID, 100, 100)

	// Initially empty
	ids, err := repo.GetBoxEntryIDs(shopID, charID)
	if err != nil {
		t.Fatalf("GetBoxEntryIDs failed: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("Expected empty box, got: %d entries", len(ids))
	}

	// Insert drawn entries
	if err := repo.InsertBoxEntry(shopID, entryID1, charID); err != nil {
		t.Fatalf("InsertBoxEntry failed: %v", err)
	}
	if err := repo.InsertBoxEntry(shopID, entryID2, charID); err != nil {
		t.Fatalf("InsertBoxEntry failed: %v", err)
	}

	ids, err = repo.GetBoxEntryIDs(shopID, charID)
	if err != nil {
		t.Fatalf("GetBoxEntryIDs failed: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("Expected 2 box entries, got: %d", len(ids))
	}

	// Delete all box entries (reset)
	if err := repo.DeleteBoxEntries(shopID, charID); err != nil {
		t.Fatalf("DeleteBoxEntries failed: %v", err)
	}

	ids, err = repo.GetBoxEntryIDs(shopID, charID)
	if err != nil {
		t.Fatalf("GetBoxEntryIDs after delete failed: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("Expected empty box after delete, got: %d", len(ids))
	}
}

func TestRepoGachaBoxIsolation(t *testing.T) {
	repo, db, charID := setupGachaRepo(t)

	shopID := CreateTestGachaShop(t, db, "Box Iso", 1)
	entryID := CreateTestGachaEntry(t, db, shopID, 100, 50)

	// Create another character
	userID2 := CreateTestUser(t, db, "gacha_other_user")
	charID2 := CreateTestCharacter(t, db, userID2, "GachaChar2")

	// Char1 draws
	if err := repo.InsertBoxEntry(shopID, entryID, charID); err != nil {
		t.Fatalf("InsertBoxEntry failed: %v", err)
	}

	// Char2 should have empty box
	ids, err := repo.GetBoxEntryIDs(shopID, charID2)
	if err != nil {
		t.Fatalf("GetBoxEntryIDs for char2 failed: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("Expected empty box for char2, got: %d entries", len(ids))
	}
}
