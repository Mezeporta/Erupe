package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupMiscRepo(t *testing.T) (*MiscRepository, *sqlx.DB) {
	t.Helper()
	db := SetupTestDB(t)
	repo := NewMiscRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db
}

func TestRepoMiscUpsertTrendWeapon(t *testing.T) {
	repo, db := setupMiscRepo(t)

	if err := repo.UpsertTrendWeapon(100, 1); err != nil {
		t.Fatalf("UpsertTrendWeapon failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT count FROM trend_weapons WHERE weapon_id=100").Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count=1, got: %d", count)
	}
}

func TestRepoMiscUpsertTrendWeaponIncrement(t *testing.T) {
	repo, db := setupMiscRepo(t)

	if err := repo.UpsertTrendWeapon(100, 1); err != nil {
		t.Fatalf("First UpsertTrendWeapon failed: %v", err)
	}
	if err := repo.UpsertTrendWeapon(100, 1); err != nil {
		t.Fatalf("Second UpsertTrendWeapon failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT count FROM trend_weapons WHERE weapon_id=100").Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count=2 after upsert, got: %d", count)
	}
}

func TestRepoMiscGetTrendWeaponsEmpty(t *testing.T) {
	repo, _ := setupMiscRepo(t)

	weapons, err := repo.GetTrendWeapons(1)
	if err != nil {
		t.Fatalf("GetTrendWeapons failed: %v", err)
	}
	if len(weapons) != 0 {
		t.Errorf("Expected 0 weapons, got: %d", len(weapons))
	}
}

func TestRepoMiscGetTrendWeaponsOrdering(t *testing.T) {
	repo, _ := setupMiscRepo(t)

	// Insert weapons with different counts
	for i := 0; i < 3; i++ {
		if err := repo.UpsertTrendWeapon(uint16(100+i), 1); err != nil {
			t.Fatalf("UpsertTrendWeapon failed: %v", err)
		}
	}
	// Give weapon 101 more uses
	if err := repo.UpsertTrendWeapon(101, 1); err != nil {
		t.Fatalf("UpsertTrendWeapon failed: %v", err)
	}
	if err := repo.UpsertTrendWeapon(101, 1); err != nil {
		t.Fatalf("UpsertTrendWeapon failed: %v", err)
	}

	weapons, err := repo.GetTrendWeapons(1)
	if err != nil {
		t.Fatalf("GetTrendWeapons failed: %v", err)
	}
	if len(weapons) != 3 {
		t.Fatalf("Expected 3 weapons, got: %d", len(weapons))
	}
	// First should be the one with highest count (101 with count=3)
	if weapons[0] != 101 {
		t.Errorf("Expected first weapon=101 (highest count), got: %d", weapons[0])
	}
}

func TestRepoMiscGetTrendWeaponsLimit3(t *testing.T) {
	repo, _ := setupMiscRepo(t)

	for i := 0; i < 5; i++ {
		if err := repo.UpsertTrendWeapon(uint16(100+i), 1); err != nil {
			t.Fatalf("UpsertTrendWeapon failed: %v", err)
		}
	}

	weapons, err := repo.GetTrendWeapons(1)
	if err != nil {
		t.Fatalf("GetTrendWeapons failed: %v", err)
	}
	if len(weapons) != 3 {
		t.Errorf("Expected max 3 weapons, got: %d", len(weapons))
	}
}
