package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupShopRepo(t *testing.T) (*ShopRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "shop_test_user")
	charID := CreateTestCharacter(t, db, userID, "ShopChar")
	repo := NewShopRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func TestRepoShopGetShopItemsEmpty(t *testing.T) {
	repo, _, charID := setupShopRepo(t)

	items, err := repo.GetShopItems(1, 1, charID)
	if err != nil {
		t.Fatalf("GetShopItems failed: %v", err)
	}
	if len(items) != 0 {
		t.Errorf("Expected 0 items, got: %d", len(items))
	}
}

func TestRepoShopGetShopItems(t *testing.T) {
	repo, db, charID := setupShopRepo(t)

	// Insert shop items
	if _, err := db.Exec(
		`INSERT INTO shop_items (id, shop_type, shop_id, item_id, cost, quantity, min_hr, min_sr, min_gr, store_level, max_quantity, road_floors, road_fatalis)
		VALUES (1, 1, 100, 500, 1000, 1, 0, 0, 0, 0, 99, 0, 0)`,
	); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	items, err := repo.GetShopItems(1, 100, charID)
	if err != nil {
		t.Fatalf("GetShopItems failed: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got: %d", len(items))
	}
	if items[0].ItemID != 500 {
		t.Errorf("Expected item_id=500, got: %d", items[0].ItemID)
	}
	if items[0].Cost != 1000 {
		t.Errorf("Expected cost=1000, got: %d", items[0].Cost)
	}
	if items[0].UsedQuantity != 0 {
		t.Errorf("Expected used_quantity=0, got: %d", items[0].UsedQuantity)
	}
}

func TestRepoShopRecordPurchaseInsertAndUpdate(t *testing.T) {
	repo, db, charID := setupShopRepo(t)

	// First purchase inserts a new row
	if err := repo.RecordPurchase(charID, 1, 3); err != nil {
		t.Fatalf("RecordPurchase (insert) failed: %v", err)
	}

	var bought int
	if err := db.QueryRow("SELECT bought FROM shop_items_bought WHERE character_id=$1 AND shop_item_id=$2", charID, 1).Scan(&bought); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if bought != 3 {
		t.Errorf("Expected bought=3, got: %d", bought)
	}

	// Second purchase updates (adds to) the existing row
	if err := repo.RecordPurchase(charID, 1, 2); err != nil {
		t.Fatalf("RecordPurchase (update) failed: %v", err)
	}

	if err := db.QueryRow("SELECT bought FROM shop_items_bought WHERE character_id=$1 AND shop_item_id=$2", charID, 1).Scan(&bought); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if bought != 5 {
		t.Errorf("Expected bought=5 (3+2), got: %d", bought)
	}
}

func TestRepoShopGetFpointItem(t *testing.T) {
	repo, db, _ := setupShopRepo(t)

	if _, err := db.Exec("INSERT INTO fpoint_items (id, item_type, item_id, quantity, fpoints, buyable) VALUES (1, 1, 100, 5, 200, true)"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	quantity, fpoints, err := repo.GetFpointItem(1)
	if err != nil {
		t.Fatalf("GetFpointItem failed: %v", err)
	}
	if quantity != 5 {
		t.Errorf("Expected quantity=5, got: %d", quantity)
	}
	if fpoints != 200 {
		t.Errorf("Expected fpoints=200, got: %d", fpoints)
	}
}

func TestRepoShopGetFpointExchangeList(t *testing.T) {
	repo, db, _ := setupShopRepo(t)

	if _, err := db.Exec("INSERT INTO fpoint_items (id, item_type, item_id, quantity, fpoints, buyable) VALUES (1, 1, 100, 5, 200, true)"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if _, err := db.Exec("INSERT INTO fpoint_items (id, item_type, item_id, quantity, fpoints, buyable) VALUES (2, 2, 200, 10, 500, false)"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	exchanges, err := repo.GetFpointExchangeList()
	if err != nil {
		t.Fatalf("GetFpointExchangeList failed: %v", err)
	}
	if len(exchanges) != 2 {
		t.Fatalf("Expected 2 exchange items, got: %d", len(exchanges))
	}
	// Ordered by buyable DESC, so buyable=true first
	if !exchanges[0].Buyable {
		t.Error("Expected first item to have buyable=true")
	}
}

func TestRepoShopGetFpointExchangeListEmpty(t *testing.T) {
	repo, _, _ := setupShopRepo(t)

	exchanges, err := repo.GetFpointExchangeList()
	if err != nil {
		t.Fatalf("GetFpointExchangeList failed: %v", err)
	}
	if len(exchanges) != 0 {
		t.Errorf("Expected 0 exchange items, got: %d", len(exchanges))
	}
}
