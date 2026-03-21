package channelserver

// Tests for guild subsystem methods not covered by repo_guild_test.go:
//   - SetAllianceRecruiting (repo_guild_alliance.go)
//   - RolloverDailyRP       (repo_guild_rp.go)
//   - AddWeeklyBonusUsers   (repo_guild_rp.go)
//   - InsertKillLog         (repo_guild_hunt.go)
//   - ClearTreasureHunt     (repo_guild_hunt.go)

import (
	"testing"
	"time"
)

func TestSetAllianceRecruiting(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "sar_user")
	charID := CreateTestCharacter(t, db, userID, "SAR_Leader")
	guildID := CreateTestGuild(t, db, charID, "SAR_Guild")
	repo := NewGuildRepository(db)

	if err := repo.CreateAlliance("SAR_Alliance", guildID); err != nil {
		t.Fatalf("CreateAlliance failed: %v", err)
	}
	alliances, err := repo.ListAlliances()
	if err != nil {
		t.Fatalf("ListAlliances failed: %v", err)
	}
	if len(alliances) == 0 {
		t.Fatal("Expected at least 1 alliance")
	}
	allianceID := alliances[0].ID

	// Default should be false.
	if alliances[0].Recruiting {
		t.Error("Expected initial Recruiting=false")
	}

	if err := repo.SetAllianceRecruiting(allianceID, true); err != nil {
		t.Fatalf("SetAllianceRecruiting(true) failed: %v", err)
	}
	alliance, err := repo.GetAllianceByID(allianceID)
	if err != nil {
		t.Fatalf("GetAllianceByID after set true failed: %v", err)
	}
	if !alliance.Recruiting {
		t.Error("Expected Recruiting=true after SetAllianceRecruiting(true)")
	}

	if err := repo.SetAllianceRecruiting(allianceID, false); err != nil {
		t.Fatalf("SetAllianceRecruiting(false) failed: %v", err)
	}
	alliance, err = repo.GetAllianceByID(allianceID)
	if err != nil {
		t.Fatalf("GetAllianceByID after set false failed: %v", err)
	}
	if alliance.Recruiting {
		t.Error("Expected Recruiting=false after SetAllianceRecruiting(false)")
	}
}

func TestRolloverDailyRP(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "rollover_user")
	charID := CreateTestCharacter(t, db, userID, "Rollover_Leader")
	guildID := CreateTestGuild(t, db, charID, "Rollover_Guild")
	repo := NewGuildRepository(db)

	// Set rp_today for the member so we can verify the rollover.
	if _, err := db.Exec("UPDATE guild_characters SET rp_today = 50 WHERE character_id = $1", charID); err != nil {
		t.Fatalf("Failed to set rp_today: %v", err)
	}

	noon := time.Now().UTC()
	if err := repo.RolloverDailyRP(guildID, noon); err != nil {
		t.Fatalf("RolloverDailyRP failed: %v", err)
	}

	var rpToday, rpYesterday int
	if err := db.QueryRow("SELECT rp_today, rp_yesterday FROM guild_characters WHERE character_id = $1", charID).
		Scan(&rpToday, &rpYesterday); err != nil {
		t.Fatalf("Failed to read rp values: %v", err)
	}
	if rpToday != 0 {
		t.Errorf("Expected rp_today=0 after rollover, got %d", rpToday)
	}
	if rpYesterday != 50 {
		t.Errorf("Expected rp_yesterday=50 after rollover, got %d", rpYesterday)
	}
}

func TestRolloverDailyRP_Idempotent(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "idem_rollover_user")
	charID := CreateTestCharacter(t, db, userID, "Idem_Rollover_Leader")
	guildID := CreateTestGuild(t, db, charID, "Idem_Rollover_Guild")
	repo := NewGuildRepository(db)

	if _, err := db.Exec("UPDATE guild_characters SET rp_today = 100 WHERE character_id = $1", charID); err != nil {
		t.Fatalf("Failed to set rp_today: %v", err)
	}

	noon := time.Now().UTC()
	if err := repo.RolloverDailyRP(guildID, noon); err != nil {
		t.Fatalf("First RolloverDailyRP failed: %v", err)
	}
	// Second call with same noon should be a no-op (rp_reset_at >= noon).
	if err := repo.RolloverDailyRP(guildID, noon); err != nil {
		t.Fatalf("Second RolloverDailyRP (idempotent) failed: %v", err)
	}

	var rpToday int
	_ = db.QueryRow("SELECT rp_today FROM guild_characters WHERE character_id = $1", charID).Scan(&rpToday)
	if rpToday != 0 {
		t.Errorf("Expected rp_today=0 after idempotent rollover, got %d", rpToday)
	}
}

func TestAddWeeklyBonusUsers(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "wbu_user")
	charID := CreateTestCharacter(t, db, userID, "WBU_Leader")
	guildID := CreateTestGuild(t, db, charID, "WBU_Guild")
	repo := NewGuildRepository(db)

	if err := repo.AddWeeklyBonusUsers(guildID, 3); err != nil {
		t.Fatalf("AddWeeklyBonusUsers failed: %v", err)
	}

	// Verify the column incremented.
	var wbu int
	if err := db.QueryRow("SELECT weekly_bonus_users FROM guilds WHERE id = $1", guildID).Scan(&wbu); err != nil {
		t.Fatalf("Failed to read weekly_bonus_users: %v", err)
	}
	if wbu != 3 {
		t.Errorf("Expected weekly_bonus_users=3, got %d", wbu)
	}

	// Add again and verify accumulation.
	if err := repo.AddWeeklyBonusUsers(guildID, 2); err != nil {
		t.Fatalf("Second AddWeeklyBonusUsers failed: %v", err)
	}
	if err := db.QueryRow("SELECT weekly_bonus_users FROM guilds WHERE id = $1", guildID).Scan(&wbu); err != nil {
		t.Fatalf("Failed to read weekly_bonus_users after second add: %v", err)
	}
	if wbu != 5 {
		t.Errorf("Expected weekly_bonus_users=5 after second add, got %d", wbu)
	}
}

func TestInsertKillLogAndCount(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "kill_log_user")
	charID := CreateTestCharacter(t, db, userID, "Kill_Logger")
	guildID := CreateTestGuild(t, db, charID, "Kill_Guild")
	repo := NewGuildRepository(db)

	// Set box_claimed to 1 hour ago so kills inserted now are within the window.
	if _, err := db.Exec("UPDATE guild_characters SET box_claimed = now() - interval '1 hour' WHERE character_id = $1", charID); err != nil {
		t.Fatalf("Failed to set box_claimed: %v", err)
	}

	if err := repo.InsertKillLog(charID, 42, 2, time.Now()); err != nil {
		t.Fatalf("InsertKillLog failed: %v", err)
	}

	count, err := repo.CountGuildKills(guildID, charID)
	if err != nil {
		t.Fatalf("CountGuildKills failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 kill log entry, got %d", count)
	}
}

func TestClearTreasureHunt(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)

	userID := CreateTestUser(t, db, "cth_user")
	charID := CreateTestCharacter(t, db, userID, "CTH_Leader")
	guildID := CreateTestGuild(t, db, charID, "CTH_Guild")
	repo := NewGuildRepository(db)

	// Create and register a hunt.
	if err := repo.CreateHunt(guildID, charID, 7, 1, []byte{}, ""); err != nil {
		t.Fatalf("CreateHunt failed: %v", err)
	}
	hunt, err := repo.GetPendingHunt(charID)
	if err != nil || hunt == nil {
		t.Fatalf("GetPendingHunt failed or nil: %v", err)
	}
	if err := repo.RegisterHuntReport(hunt.HuntID, charID); err != nil {
		t.Fatalf("RegisterHuntReport failed: %v", err)
	}

	// Verify treasure_hunt is set.
	var th interface{}
	if err := db.QueryRow("SELECT treasure_hunt FROM guild_characters WHERE character_id = $1", charID).Scan(&th); err != nil {
		t.Fatalf("Failed to read treasure_hunt: %v", err)
	}
	if th == nil {
		t.Error("Expected treasure_hunt to be set after RegisterHuntReport")
	}

	// Clear it.
	if err := repo.ClearTreasureHunt(charID); err != nil {
		t.Fatalf("ClearTreasureHunt failed: %v", err)
	}

	if err := db.QueryRow("SELECT treasure_hunt FROM guild_characters WHERE character_id = $1", charID).Scan(&th); err != nil {
		t.Fatalf("Failed to read treasure_hunt after clear: %v", err)
	}
	if th != nil {
		t.Errorf("Expected treasure_hunt=nil after ClearTreasureHunt, got %v", th)
	}
}
