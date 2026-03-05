package channelserver

import (
	dbutil "erupe-ce/common/db"
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupDistributionRepo(t *testing.T) (*DistributionRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "dist_test_user")
	charID := CreateTestCharacter(t, db, userID, "DistChar")
	repo := NewDistributionRepository(dbutil.Wrap(db))
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func createDistribution(t *testing.T, db *sqlx.DB, charID *uint32, distType int, eventName, description string) uint32 {
	t.Helper()
	var id uint32
	err := db.QueryRow(
		`INSERT INTO distribution (character_id, type, event_name, description, times_acceptable)
		VALUES ($1, $2, $3, $4, 1) RETURNING id`,
		charID, distType, eventName, description,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to create distribution: %v", err)
	}
	return id
}

func TestRepoDistributionListEmpty(t *testing.T) {
	repo, _, charID := setupDistributionRepo(t)

	dists, err := repo.List(charID, 1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(dists) != 0 {
		t.Errorf("Expected 0 distributions, got: %d", len(dists))
	}
}

func TestRepoDistributionListCharacterSpecific(t *testing.T) {
	repo, db, charID := setupDistributionRepo(t)

	createDistribution(t, db, &charID, 1, "Personal Gift", "For you")

	dists, err := repo.List(charID, 1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(dists) != 1 {
		t.Fatalf("Expected 1 distribution, got: %d", len(dists))
	}
	if dists[0].EventName != "Personal Gift" {
		t.Errorf("Expected event_name='Personal Gift', got: %q", dists[0].EventName)
	}
}

func TestRepoDistributionListGlobal(t *testing.T) {
	repo, db, charID := setupDistributionRepo(t)

	// Global distribution (character_id=NULL)
	createDistribution(t, db, nil, 1, "Global Gift", "For everyone")

	dists, err := repo.List(charID, 1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(dists) != 1 {
		t.Fatalf("Expected 1 global distribution, got: %d", len(dists))
	}
}

func TestRepoDistributionGetItems(t *testing.T) {
	repo, db, charID := setupDistributionRepo(t)

	distID := createDistribution(t, db, &charID, 1, "Item Gift", "Has items")
	if _, err := db.Exec("INSERT INTO distribution_items (distribution_id, item_type, item_id, quantity) VALUES ($1, 1, 100, 5)", distID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if _, err := db.Exec("INSERT INTO distribution_items (distribution_id, item_type, item_id, quantity) VALUES ($1, 2, 200, 10)", distID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	items, err := repo.GetItems(distID)
	if err != nil {
		t.Fatalf("GetItems failed: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("Expected 2 items, got: %d", len(items))
	}
}

func TestRepoDistributionRecordAccepted(t *testing.T) {
	repo, db, charID := setupDistributionRepo(t)

	distID := createDistribution(t, db, &charID, 1, "Accept Test", "Test")

	if err := repo.RecordAccepted(distID, charID); err != nil {
		t.Fatalf("RecordAccepted failed: %v", err)
	}

	// Verify accepted count in list
	dists, err := repo.List(charID, 1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(dists) != 1 {
		t.Fatalf("Expected 1 distribution, got: %d", len(dists))
	}
	if dists[0].TimesAccepted != 1 {
		t.Errorf("Expected times_accepted=1, got: %d", dists[0].TimesAccepted)
	}
}

func TestRepoDistributionGetDescription(t *testing.T) {
	repo, db, charID := setupDistributionRepo(t)

	distID := createDistribution(t, db, &charID, 1, "Desc Test", "~C05Special reward!")

	desc, err := repo.GetDescription(distID)
	if err != nil {
		t.Fatalf("GetDescription failed: %v", err)
	}
	if desc != "~C05Special reward!" {
		t.Errorf("Expected description='~C05Special reward!', got: %q", desc)
	}
}

func TestRepoDistributionFiltersByType(t *testing.T) {
	repo, db, charID := setupDistributionRepo(t)

	createDistribution(t, db, &charID, 1, "Type 1", "Type 1")
	createDistribution(t, db, &charID, 2, "Type 2", "Type 2")

	dists, err := repo.List(charID, 1)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(dists) != 1 {
		t.Errorf("Expected 1 distribution of type 1, got: %d", len(dists))
	}
}
