package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupCafeRepo(t *testing.T) (*CafeRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "cafe_test_user")
	charID := CreateTestCharacter(t, db, userID, "CafeChar")
	repo := NewCafeRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func createCafeBonus(t *testing.T, db *sqlx.DB, id uint32, timeReq, itemType, itemID, quantity int) {
	t.Helper()
	if _, err := db.Exec(
		"INSERT INTO cafebonus (id, time_req, item_type, item_id, quantity) VALUES ($1, $2, $3, $4, $5)",
		id, timeReq, itemType, itemID, quantity,
	); err != nil {
		t.Fatalf("Failed to create cafe bonus: %v", err)
	}
}

func TestRepoCafeGetBonusesEmpty(t *testing.T) {
	repo, _, charID := setupCafeRepo(t)

	bonuses, err := repo.GetBonuses(charID)
	if err != nil {
		t.Fatalf("GetBonuses failed: %v", err)
	}
	if len(bonuses) != 0 {
		t.Errorf("Expected 0 bonuses, got: %d", len(bonuses))
	}
}

func TestRepoCafeGetBonuses(t *testing.T) {
	repo, db, charID := setupCafeRepo(t)

	createCafeBonus(t, db, 1, 3600, 1, 100, 5)
	createCafeBonus(t, db, 2, 7200, 2, 200, 10)

	bonuses, err := repo.GetBonuses(charID)
	if err != nil {
		t.Fatalf("GetBonuses failed: %v", err)
	}
	if len(bonuses) != 2 {
		t.Fatalf("Expected 2 bonuses, got: %d", len(bonuses))
	}
	if bonuses[0].Claimed {
		t.Error("Expected first bonus unclaimed")
	}
}

func TestRepoCafeAcceptBonus(t *testing.T) {
	repo, db, charID := setupCafeRepo(t)

	createCafeBonus(t, db, 1, 3600, 1, 100, 5)

	if err := repo.AcceptBonus(1, charID); err != nil {
		t.Fatalf("AcceptBonus failed: %v", err)
	}

	bonuses, err := repo.GetBonuses(charID)
	if err != nil {
		t.Fatalf("GetBonuses failed: %v", err)
	}
	if len(bonuses) != 1 {
		t.Fatalf("Expected 1 bonus, got: %d", len(bonuses))
	}
	if !bonuses[0].Claimed {
		t.Error("Expected bonus to be claimed after AcceptBonus")
	}
}

func TestRepoCafeResetAccepted(t *testing.T) {
	repo, db, charID := setupCafeRepo(t)

	createCafeBonus(t, db, 1, 3600, 1, 100, 5)
	if err := repo.AcceptBonus(1, charID); err != nil {
		t.Fatalf("AcceptBonus failed: %v", err)
	}

	if err := repo.ResetAccepted(charID); err != nil {
		t.Fatalf("ResetAccepted failed: %v", err)
	}

	bonuses, err := repo.GetBonuses(charID)
	if err != nil {
		t.Fatalf("GetBonuses failed: %v", err)
	}
	if bonuses[0].Claimed {
		t.Error("Expected bonus unclaimed after ResetAccepted")
	}
}

func TestRepoCafeGetBonusItem(t *testing.T) {
	repo, db, _ := setupCafeRepo(t)

	createCafeBonus(t, db, 1, 3600, 7, 500, 3)

	itemType, quantity, err := repo.GetBonusItem(1)
	if err != nil {
		t.Fatalf("GetBonusItem failed: %v", err)
	}
	if itemType != 7 {
		t.Errorf("Expected itemType=7, got: %d", itemType)
	}
	if quantity != 3 {
		t.Errorf("Expected quantity=3, got: %d", quantity)
	}
}

func TestRepoCafeGetClaimable(t *testing.T) {
	repo, db, charID := setupCafeRepo(t)

	// Set character's cafe_time to 1000 seconds
	if _, err := db.Exec("UPDATE characters SET cafe_time=1000 WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	// Bonus requiring 500 seconds total (1000 + 0 elapsed >= 500) - claimable
	createCafeBonus(t, db, 1, 500, 1, 100, 1)
	// Bonus requiring 5000 seconds (1000 + 100 elapsed < 5000) - not claimable
	createCafeBonus(t, db, 2, 5000, 2, 200, 1)

	claimable, err := repo.GetClaimable(charID, 100)
	if err != nil {
		t.Fatalf("GetClaimable failed: %v", err)
	}
	if len(claimable) != 1 {
		t.Fatalf("Expected 1 claimable bonus, got: %d", len(claimable))
	}
	if claimable[0].ID != 1 {
		t.Errorf("Expected claimable bonus ID=1, got: %d", claimable[0].ID)
	}
}

func TestRepoCafeGetClaimableExcludesAccepted(t *testing.T) {
	repo, db, charID := setupCafeRepo(t)

	if _, err := db.Exec("UPDATE characters SET cafe_time=10000 WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	createCafeBonus(t, db, 1, 100, 1, 100, 1)
	if err := repo.AcceptBonus(1, charID); err != nil {
		t.Fatalf("AcceptBonus failed: %v", err)
	}

	claimable, err := repo.GetClaimable(charID, 0)
	if err != nil {
		t.Fatalf("GetClaimable failed: %v", err)
	}
	if len(claimable) != 0 {
		t.Errorf("Expected 0 claimable after accept, got: %d", len(claimable))
	}
}
