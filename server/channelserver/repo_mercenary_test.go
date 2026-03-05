package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupMercenaryRepo(t *testing.T) (*MercenaryRepository, *sqlx.DB, uint32, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "merc_test_user")
	charID := CreateTestCharacter(t, db, userID, "MercChar")
	guildID := CreateTestGuild(t, db, charID, "MercGuild")
	repo := NewMercenaryRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID, guildID
}

func TestRepoMercenaryNextRastaID(t *testing.T) {
	repo, _, _, _ := setupMercenaryRepo(t)

	id1, err := repo.NextRastaID()
	if err != nil {
		t.Fatalf("NextRastaID failed: %v", err)
	}
	id2, err := repo.NextRastaID()
	if err != nil {
		t.Fatalf("NextRastaID second call failed: %v", err)
	}
	if id2 <= id1 {
		t.Errorf("Expected increasing IDs, got: %d then %d", id1, id2)
	}
}

func TestRepoMercenaryNextAirouID(t *testing.T) {
	repo, _, _, _ := setupMercenaryRepo(t)

	id1, err := repo.NextAirouID()
	if err != nil {
		t.Fatalf("NextAirouID failed: %v", err)
	}
	id2, err := repo.NextAirouID()
	if err != nil {
		t.Fatalf("NextAirouID second call failed: %v", err)
	}
	if id2 <= id1 {
		t.Errorf("Expected increasing IDs, got: %d then %d", id1, id2)
	}
}

func TestRepoMercenaryGetMercenaryLoansEmpty(t *testing.T) {
	repo, _, charID, _ := setupMercenaryRepo(t)

	loans, err := repo.GetMercenaryLoans(charID)
	if err != nil {
		t.Fatalf("GetMercenaryLoans failed: %v", err)
	}
	if len(loans) != 0 {
		t.Errorf("Expected 0 loans, got: %d", len(loans))
	}
}

func TestRepoMercenaryGetMercenaryLoans(t *testing.T) {
	repo, db, charID, _ := setupMercenaryRepo(t)

	// Set rasta_id on charID
	if _, err := db.Exec("UPDATE characters SET rasta_id=999 WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup rasta_id failed: %v", err)
	}

	// Create another character that has a pact with charID's rasta
	user2 := CreateTestUser(t, db, "merc_user2")
	char2 := CreateTestCharacter(t, db, user2, "PactHolder")
	if _, err := db.Exec("UPDATE characters SET pact_id=999 WHERE id=$1", char2); err != nil {
		t.Fatalf("Setup pact_id failed: %v", err)
	}

	loans, err := repo.GetMercenaryLoans(charID)
	if err != nil {
		t.Fatalf("GetMercenaryLoans failed: %v", err)
	}
	if len(loans) != 1 {
		t.Fatalf("Expected 1 loan, got: %d", len(loans))
	}
	if loans[0].Name != "PactHolder" {
		t.Errorf("Expected name='PactHolder', got: %q", loans[0].Name)
	}
	if loans[0].CharID != char2 {
		t.Errorf("Expected charID=%d, got: %d", char2, loans[0].CharID)
	}
}

func TestRepoMercenaryGetGuildHuntCatsUsedEmpty(t *testing.T) {
	repo, _, charID, _ := setupMercenaryRepo(t)

	cats, err := repo.GetGuildHuntCatsUsed(charID)
	if err != nil {
		t.Fatalf("GetGuildHuntCatsUsed failed: %v", err)
	}
	if len(cats) != 0 {
		t.Errorf("Expected 0 cat usages, got: %d", len(cats))
	}
}

func TestRepoMercenaryGetGuildHuntCatsUsed(t *testing.T) {
	repo, db, charID, guildID := setupMercenaryRepo(t)

	// Insert a guild hunt with cats_used
	if _, err := db.Exec(
		`INSERT INTO guild_hunts (guild_id, host_id, destination, level, hunt_data, cats_used, acquired, collected, start)
		VALUES ($1, $2, 1, 1, $3, '1,2,3', false, false, now())`,
		guildID, charID, []byte{0x00},
	); err != nil {
		t.Fatalf("Setup guild_hunts failed: %v", err)
	}

	cats, err := repo.GetGuildHuntCatsUsed(charID)
	if err != nil {
		t.Fatalf("GetGuildHuntCatsUsed failed: %v", err)
	}
	if len(cats) != 1 {
		t.Fatalf("Expected 1 cat usage, got: %d", len(cats))
	}
	if cats[0].CatsUsed != "1,2,3" {
		t.Errorf("Expected cats_used='1,2,3', got: %q", cats[0].CatsUsed)
	}
}

func TestRepoMercenaryGetGuildAirouEmpty(t *testing.T) {
	repo, _, _, guildID := setupMercenaryRepo(t)

	airou, err := repo.GetGuildAirou(guildID)
	if err != nil {
		t.Fatalf("GetGuildAirou failed: %v", err)
	}
	if len(airou) != 0 {
		t.Errorf("Expected 0 airou, got: %d", len(airou))
	}
}

func TestRepoMercenaryGetGuildAirou(t *testing.T) {
	repo, db, charID, guildID := setupMercenaryRepo(t)

	// Set otomoairou on the character
	airouData := []byte{0xAA, 0xBB, 0xCC}
	if _, err := db.Exec("UPDATE characters SET otomoairou=$1 WHERE id=$2", airouData, charID); err != nil {
		t.Fatalf("Setup otomoairou failed: %v", err)
	}

	airou, err := repo.GetGuildAirou(guildID)
	if err != nil {
		t.Fatalf("GetGuildAirou failed: %v", err)
	}
	if len(airou) != 1 {
		t.Fatalf("Expected 1 airou, got: %d", len(airou))
	}
	if len(airou[0]) != 3 || airou[0][0] != 0xAA {
		t.Errorf("Expected airou data, got: %x", airou[0])
	}
}
