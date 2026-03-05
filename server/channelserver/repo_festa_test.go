package channelserver

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func setupFestaRepo(t *testing.T) (*FestaRepository, *sqlx.DB, uint32, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "festa_test_user")
	charID := CreateTestCharacter(t, db, userID, "FestaChar")
	guildID := CreateTestGuild(t, db, charID, "FestaGuild")
	repo := NewFestaRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID, guildID
}

func TestRepoFestaInsertAndGetEvents(t *testing.T) {
	repo, _, _, _ := setupFestaRepo(t)

	startTime := uint32(time.Date(2025, 6, 1, 0, 0, 0, 0, time.UTC).Unix())
	if err := repo.InsertEvent(startTime); err != nil {
		t.Fatalf("InsertEvent failed: %v", err)
	}

	events, err := repo.GetFestaEvents()
	if err != nil {
		t.Fatalf("GetFestaEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got: %d", len(events))
	}
	if events[0].StartTime != startTime {
		t.Errorf("Expected start_time=%d, got: %d", startTime, events[0].StartTime)
	}
}

func TestRepoFestaCleanupAll(t *testing.T) {
	repo, _, _, _ := setupFestaRepo(t)

	if err := repo.InsertEvent(1000000); err != nil {
		t.Fatalf("InsertEvent failed: %v", err)
	}

	if err := repo.CleanupAll(); err != nil {
		t.Fatalf("CleanupAll failed: %v", err)
	}

	events, err := repo.GetFestaEvents()
	if err != nil {
		t.Fatalf("GetFestaEvents failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events after cleanup, got: %d", len(events))
	}
}

func TestRepoFestaRegisterGuild(t *testing.T) {
	repo, db, _, guildID := setupFestaRepo(t)

	if err := repo.RegisterGuild(guildID, "blue"); err != nil {
		t.Fatalf("RegisterGuild failed: %v", err)
	}

	var team string
	if err := db.QueryRow("SELECT team FROM festa_registrations WHERE guild_id=$1", guildID).Scan(&team); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if team != "blue" {
		t.Errorf("Expected team='blue', got: %q", team)
	}
}

func TestRepoFestaGetTeamSouls(t *testing.T) {
	repo, _, _, guildID := setupFestaRepo(t)

	if err := repo.RegisterGuild(guildID, "red"); err != nil {
		t.Fatalf("RegisterGuild failed: %v", err)
	}

	souls, err := repo.GetTeamSouls("red")
	if err != nil {
		t.Fatalf("GetTeamSouls failed: %v", err)
	}
	// No submissions yet, should be 0
	if souls != 0 {
		t.Errorf("Expected souls=0, got: %d", souls)
	}
}

func TestRepoFestaSubmitSouls(t *testing.T) {
	repo, _, charID, guildID := setupFestaRepo(t)

	if err := repo.RegisterGuild(guildID, "blue"); err != nil {
		t.Fatalf("RegisterGuild failed: %v", err)
	}

	souls := []uint16{10, 20, 30}
	if err := repo.SubmitSouls(charID, guildID, souls); err != nil {
		t.Fatalf("SubmitSouls failed: %v", err)
	}

	charSouls, err := repo.GetCharSouls(charID)
	if err != nil {
		t.Fatalf("GetCharSouls failed: %v", err)
	}
	// 10 + 20 + 30 = 60
	if charSouls != 60 {
		t.Errorf("Expected charSouls=60, got: %d", charSouls)
	}
}

func TestRepoFestaGetCharSoulsEmpty(t *testing.T) {
	repo, _, charID, _ := setupFestaRepo(t)

	souls, err := repo.GetCharSouls(charID)
	if err != nil {
		t.Fatalf("GetCharSouls failed: %v", err)
	}
	if souls != 0 {
		t.Errorf("Expected souls=0, got: %d", souls)
	}
}

func TestRepoFestaVoteTrial(t *testing.T) {
	repo, db, charID, _ := setupFestaRepo(t)

	if err := repo.VoteTrial(charID, 42); err != nil {
		t.Fatalf("VoteTrial failed: %v", err)
	}

	var trialVote *uint32
	if err := db.QueryRow("SELECT trial_vote FROM guild_characters WHERE character_id=$1", charID).Scan(&trialVote); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if trialVote == nil || *trialVote != 42 {
		t.Errorf("Expected trial_vote=42, got: %v", trialVote)
	}
}

func TestRepoFestaClaimPrize(t *testing.T) {
	repo, db, charID, _ := setupFestaRepo(t)

	if err := repo.ClaimPrize(5, charID); err != nil {
		t.Fatalf("ClaimPrize failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM festa_prizes_accepted WHERE prize_id=5 AND character_id=$1", charID).Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 accepted prize, got: %d", count)
	}
}

func TestRepoFestaHasClaimedMainPrize(t *testing.T) {
	repo, _, charID, _ := setupFestaRepo(t)

	// Not claimed yet
	if repo.HasClaimedMainPrize(charID) {
		t.Error("Expected HasClaimedMainPrize=false before claiming")
	}

	// Claim main prize (ID=0)
	if err := repo.ClaimPrize(0, charID); err != nil {
		t.Fatalf("ClaimPrize failed: %v", err)
	}

	if !repo.HasClaimedMainPrize(charID) {
		t.Error("Expected HasClaimedMainPrize=true after claiming")
	}
}

func TestRepoFestaListPrizes(t *testing.T) {
	repo, db, charID, _ := setupFestaRepo(t)

	if _, err := db.Exec("INSERT INTO festa_prizes (id, type, tier, souls_req, item_id, num_item) VALUES (1, 'personal', 1, 100, 500, 1)"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if _, err := db.Exec("INSERT INTO festa_prizes (id, type, tier, souls_req, item_id, num_item) VALUES (2, 'personal', 2, 200, 600, 2)"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if _, err := db.Exec("INSERT INTO festa_prizes (id, type, tier, souls_req, item_id, num_item) VALUES (3, 'guild', 1, 300, 700, 3)"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	prizes, err := repo.ListPrizes(charID, "personal")
	if err != nil {
		t.Fatalf("ListPrizes failed: %v", err)
	}
	if len(prizes) != 2 {
		t.Fatalf("Expected 2 personal prizes, got: %d", len(prizes))
	}
}

func TestRepoFestaListPrizesWithClaimed(t *testing.T) {
	repo, db, charID, _ := setupFestaRepo(t)

	if _, err := db.Exec("INSERT INTO festa_prizes (id, type, tier, souls_req, item_id, num_item) VALUES (1, 'personal', 1, 100, 500, 1)"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if err := repo.ClaimPrize(1, charID); err != nil {
		t.Fatalf("ClaimPrize failed: %v", err)
	}

	prizes, err := repo.ListPrizes(charID, "personal")
	if err != nil {
		t.Fatalf("ListPrizes failed: %v", err)
	}
	if len(prizes) != 1 {
		t.Fatalf("Expected 1 prize, got: %d", len(prizes))
	}
	if prizes[0].Claimed != 1 {
		t.Errorf("Expected claimed=1, got: %d", prizes[0].Claimed)
	}
}

func TestRepoFestaGetTeamSoulsWithSubmissions(t *testing.T) {
	repo, db, charID, guildID := setupFestaRepo(t)

	if err := repo.RegisterGuild(guildID, "blue"); err != nil {
		t.Fatalf("RegisterGuild failed: %v", err)
	}

	// Create second guild on red team
	user2 := CreateTestUser(t, db, "festa_user2")
	char2 := CreateTestCharacter(t, db, user2, "FestaChar2")
	guild2 := CreateTestGuild(t, db, char2, "RedGuild")
	if err := repo.RegisterGuild(guild2, "red"); err != nil {
		t.Fatalf("RegisterGuild failed: %v", err)
	}

	// Submit souls
	if err := repo.SubmitSouls(charID, guildID, []uint16{50}); err != nil {
		t.Fatalf("SubmitSouls blue failed: %v", err)
	}
	if err := repo.SubmitSouls(char2, guild2, []uint16{30}); err != nil {
		t.Fatalf("SubmitSouls red failed: %v", err)
	}

	blueSouls, err := repo.GetTeamSouls("blue")
	if err != nil {
		t.Fatalf("GetTeamSouls(blue) failed: %v", err)
	}
	if blueSouls != 50 {
		t.Errorf("Expected blue souls=50, got: %d", blueSouls)
	}

	redSouls, err := repo.GetTeamSouls("red")
	if err != nil {
		t.Fatalf("GetTeamSouls(red) failed: %v", err)
	}
	if redSouls != 30 {
		t.Errorf("Expected red souls=30, got: %d", redSouls)
	}
}
