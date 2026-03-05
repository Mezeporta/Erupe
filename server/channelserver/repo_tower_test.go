package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupTowerRepo(t *testing.T) (*TowerRepository, *sqlx.DB, uint32, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "tower_test_user")
	charID := CreateTestCharacter(t, db, userID, "TowerChar")
	leaderID := CreateTestCharacter(t, db, userID, "GuildLeader")
	guildID := CreateTestGuild(t, db, leaderID, "TowerGuild")
	// Add charID to the guild
	if _, err := db.Exec("INSERT INTO guild_characters (guild_id, character_id) VALUES ($1, $2)", guildID, charID); err != nil {
		t.Fatalf("Failed to add char to guild: %v", err)
	}
	repo := NewTowerRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID, guildID
}

func TestRepoTowerGetTowerDataAutoCreate(t *testing.T) {
	repo, _, charID, _ := setupTowerRepo(t)

	// First call should auto-create the row
	td, err := repo.GetTowerData(charID)
	if err != nil {
		t.Fatalf("GetTowerData failed: %v", err)
	}
	if td.TR != 0 || td.TRP != 0 || td.TSP != 0 {
		t.Errorf("Expected zero values, got TR=%d TRP=%d TSP=%d", td.TR, td.TRP, td.TSP)
	}
	if td.Skills == "" {
		t.Error("Expected non-empty default skills CSV")
	}
}

func TestRepoTowerGetTowerDataExisting(t *testing.T) {
	repo, db, charID, _ := setupTowerRepo(t)

	if _, err := db.Exec("INSERT INTO tower (char_id, tr, trp, tsp, block1, block2) VALUES ($1, 10, 20, 30, 40, 50)", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	td, err := repo.GetTowerData(charID)
	if err != nil {
		t.Fatalf("GetTowerData failed: %v", err)
	}
	if td.TR != 10 || td.TRP != 20 || td.TSP != 30 || td.Block1 != 40 || td.Block2 != 50 {
		t.Errorf("Expected 10/20/30/40/50, got %d/%d/%d/%d/%d", td.TR, td.TRP, td.TSP, td.Block1, td.Block2)
	}
}

func TestRepoTowerGetSkills(t *testing.T) {
	repo, db, charID, _ := setupTowerRepo(t)

	if _, err := db.Exec("INSERT INTO tower (char_id, skills) VALUES ($1, '1,2,3')", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	skills, err := repo.GetSkills(charID)
	if err != nil {
		t.Fatalf("GetSkills failed: %v", err)
	}
	if skills != "1,2,3" {
		t.Errorf("Expected '1,2,3', got: %q", skills)
	}
}

func TestRepoTowerUpdateSkills(t *testing.T) {
	repo, db, charID, _ := setupTowerRepo(t)

	if _, err := db.Exec("INSERT INTO tower (char_id, tsp) VALUES ($1, 100)", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if err := repo.UpdateSkills(charID, "5,10,15", 20); err != nil {
		t.Fatalf("UpdateSkills failed: %v", err)
	}

	var skills string
	var tsp int32
	if err := db.QueryRow("SELECT skills, tsp FROM tower WHERE char_id=$1", charID).Scan(&skills, &tsp); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if skills != "5,10,15" {
		t.Errorf("Expected skills='5,10,15', got: %q", skills)
	}
	if tsp != 80 {
		t.Errorf("Expected tsp=80 (100-20), got: %d", tsp)
	}
}

func TestRepoTowerUpdateProgress(t *testing.T) {
	repo, db, charID, _ := setupTowerRepo(t)

	if _, err := db.Exec("INSERT INTO tower (char_id) VALUES ($1)", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if err := repo.UpdateProgress(charID, 5, 10, 15, 20); err != nil {
		t.Fatalf("UpdateProgress failed: %v", err)
	}

	var tr, trp, tsp, block1 int32
	if err := db.QueryRow("SELECT tr, trp, tsp, block1 FROM tower WHERE char_id=$1", charID).Scan(&tr, &trp, &tsp, &block1); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if tr != 5 || trp != 10 || tsp != 15 || block1 != 20 {
		t.Errorf("Expected 5/10/15/20, got %d/%d/%d/%d", tr, trp, tsp, block1)
	}
}

func TestRepoTowerGetGems(t *testing.T) {
	repo, db, charID, _ := setupTowerRepo(t)

	if _, err := db.Exec("INSERT INTO tower (char_id, gems) VALUES ($1, '1,0,1')", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	gems, err := repo.GetGems(charID)
	if err != nil {
		t.Fatalf("GetGems failed: %v", err)
	}
	if gems != "1,0,1" {
		t.Errorf("Expected '1,0,1', got: %q", gems)
	}
}

func TestRepoTowerUpdateGems(t *testing.T) {
	repo, db, charID, _ := setupTowerRepo(t)

	if _, err := db.Exec("INSERT INTO tower (char_id) VALUES ($1)", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if err := repo.UpdateGems(charID, "2,3,4"); err != nil {
		t.Fatalf("UpdateGems failed: %v", err)
	}

	var gems string
	if err := db.QueryRow("SELECT gems FROM tower WHERE char_id=$1", charID).Scan(&gems); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if gems != "2,3,4" {
		t.Errorf("Expected '2,3,4', got: %q", gems)
	}
}

func TestRepoTowerGetGuildTowerRP(t *testing.T) {
	repo, _, _, guildID := setupTowerRepo(t)

	rp, err := repo.GetGuildTowerRP(guildID)
	if err != nil {
		t.Fatalf("GetGuildTowerRP failed: %v", err)
	}
	if rp != 0 {
		t.Errorf("Expected rp=0, got: %d", rp)
	}
}

func TestRepoTowerDonateGuildTowerRP(t *testing.T) {
	repo, _, _, guildID := setupTowerRepo(t)

	if err := repo.DonateGuildTowerRP(guildID, 100); err != nil {
		t.Fatalf("DonateGuildTowerRP failed: %v", err)
	}

	rp, err := repo.GetGuildTowerRP(guildID)
	if err != nil {
		t.Fatalf("GetGuildTowerRP failed: %v", err)
	}
	if rp != 100 {
		t.Errorf("Expected rp=100, got: %d", rp)
	}
}

func TestRepoTowerGetGuildTowerPageAndRP(t *testing.T) {
	repo, db, _, guildID := setupTowerRepo(t)

	if _, err := db.Exec("UPDATE guilds SET tower_mission_page=3, tower_rp=50 WHERE id=$1", guildID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	page, donated, err := repo.GetGuildTowerPageAndRP(guildID)
	if err != nil {
		t.Fatalf("GetGuildTowerPageAndRP failed: %v", err)
	}
	if page != 3 {
		t.Errorf("Expected page=3, got: %d", page)
	}
	if donated != 50 {
		t.Errorf("Expected donated=50, got: %d", donated)
	}
}

func TestRepoTowerAdvanceTenrouiraiPage(t *testing.T) {
	repo, db, charID, guildID := setupTowerRepo(t)

	// Read initial page
	var initialPage int
	if err := db.QueryRow("SELECT tower_mission_page FROM guilds WHERE id=$1", guildID).Scan(&initialPage); err != nil {
		t.Fatalf("Read initial page failed: %v", err)
	}

	// Set initial mission scores
	if _, err := db.Exec("UPDATE guild_characters SET tower_mission_1=10, tower_mission_2=20, tower_mission_3=30 WHERE character_id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if err := repo.AdvanceTenrouiraiPage(guildID); err != nil {
		t.Fatalf("AdvanceTenrouiraiPage failed: %v", err)
	}

	var page int
	if err := db.QueryRow("SELECT tower_mission_page FROM guilds WHERE id=$1", guildID).Scan(&page); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if page != initialPage+1 {
		t.Errorf("Expected page=%d (initial+1), got: %d", initialPage+1, page)
	}

	// Mission scores should be reset
	var m1, m2, m3 *int
	if err := db.QueryRow("SELECT tower_mission_1, tower_mission_2, tower_mission_3 FROM guild_characters WHERE character_id=$1", charID).Scan(&m1, &m2, &m3); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if m1 != nil || m2 != nil || m3 != nil {
		t.Errorf("Expected NULL missions after advance, got: %v/%v/%v", m1, m2, m3)
	}
}

func TestRepoTowerGetTenrouiraiProgress(t *testing.T) {
	repo, db, charID, guildID := setupTowerRepo(t)

	if _, err := db.Exec("UPDATE guilds SET tower_mission_page=2 WHERE id=$1", guildID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if _, err := db.Exec("UPDATE guild_characters SET tower_mission_1=5, tower_mission_2=10, tower_mission_3=15 WHERE character_id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	progress, err := repo.GetTenrouiraiProgress(guildID)
	if err != nil {
		t.Fatalf("GetTenrouiraiProgress failed: %v", err)
	}
	if progress.Page != 2 {
		t.Errorf("Expected page=2, got: %d", progress.Page)
	}
	if progress.Mission1 != 5 {
		t.Errorf("Expected mission1=5, got: %d", progress.Mission1)
	}
}

func TestRepoTowerGetTenrouiraiMissionScores(t *testing.T) {
	repo, db, charID, guildID := setupTowerRepo(t)

	if _, err := db.Exec("UPDATE guild_characters SET tower_mission_1=42 WHERE character_id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	scores, err := repo.GetTenrouiraiMissionScores(guildID, 1)
	if err != nil {
		t.Fatalf("GetTenrouiraiMissionScores failed: %v", err)
	}
	if len(scores) < 1 {
		t.Fatal("Expected at least 1 score entry")
	}
	if scores[0].Score != 42 {
		t.Errorf("Expected score=42, got: %d", scores[0].Score)
	}
}
