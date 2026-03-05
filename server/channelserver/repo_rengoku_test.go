package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupRengokuRepo(t *testing.T) (*RengokuRepository, *sqlx.DB, uint32, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "rengoku_test_user")
	charID := CreateTestCharacter(t, db, userID, "RengokuChar")
	guildID := CreateTestGuild(t, db, charID, "RengokuGuild")
	repo := NewRengokuRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID, guildID
}

func TestRepoRengokuUpsertScoreNew(t *testing.T) {
	repo, db, charID, _ := setupRengokuRepo(t)

	if err := repo.UpsertScore(charID, 10, 500, 5, 200); err != nil {
		t.Fatalf("UpsertScore failed: %v", err)
	}

	var stagesMp, pointsMp, stagesSp, pointsSp uint32
	if err := db.QueryRow("SELECT max_stages_mp, max_points_mp, max_stages_sp, max_points_sp FROM rengoku_score WHERE character_id=$1", charID).Scan(&stagesMp, &pointsMp, &stagesSp, &pointsSp); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if stagesMp != 10 || pointsMp != 500 || stagesSp != 5 || pointsSp != 200 {
		t.Errorf("Expected 10/500/5/200, got %d/%d/%d/%d", stagesMp, pointsMp, stagesSp, pointsSp)
	}
}

func TestRepoRengokuUpsertScoreUpdate(t *testing.T) {
	repo, db, charID, _ := setupRengokuRepo(t)

	if err := repo.UpsertScore(charID, 10, 500, 5, 200); err != nil {
		t.Fatalf("First UpsertScore failed: %v", err)
	}
	if err := repo.UpsertScore(charID, 20, 1000, 15, 800); err != nil {
		t.Fatalf("Second UpsertScore failed: %v", err)
	}

	var stagesMp, pointsMp uint32
	if err := db.QueryRow("SELECT max_stages_mp, max_points_mp FROM rengoku_score WHERE character_id=$1", charID).Scan(&stagesMp, &pointsMp); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if stagesMp != 20 || pointsMp != 1000 {
		t.Errorf("Expected 20/1000 after update, got %d/%d", stagesMp, pointsMp)
	}
}

func TestRepoRengokuGetRankingGlobal(t *testing.T) {
	repo, _, charID, _ := setupRengokuRepo(t)

	if err := repo.UpsertScore(charID, 10, 500, 5, 200); err != nil {
		t.Fatalf("UpsertScore failed: %v", err)
	}

	// Leaderboard 0 = max_stages_mp (global)
	scores, err := repo.GetRanking(0, 0)
	if err != nil {
		t.Fatalf("GetRanking failed: %v", err)
	}
	if len(scores) != 1 {
		t.Fatalf("Expected 1 score, got: %d", len(scores))
	}
	if scores[0].Score != 10 {
		t.Errorf("Expected score=10, got: %d", scores[0].Score)
	}
	if scores[0].Name != "RengokuChar" {
		t.Errorf("Expected name='RengokuChar', got: %q", scores[0].Name)
	}
}

func TestRepoRengokuGetRankingGuildFiltered(t *testing.T) {
	repo, db, charID, guildID := setupRengokuRepo(t)

	if err := repo.UpsertScore(charID, 10, 500, 5, 200); err != nil {
		t.Fatalf("UpsertScore failed: %v", err)
	}

	// Create another character in a different guild
	user2 := CreateTestUser(t, db, "rengoku_user2")
	char2 := CreateTestCharacter(t, db, user2, "RengokuChar2")
	CreateTestGuild(t, db, char2, "OtherGuild")
	if err := repo.UpsertScore(char2, 20, 1000, 15, 800); err != nil {
		t.Fatalf("UpsertScore char2 failed: %v", err)
	}

	// Leaderboard 2 = max_stages_mp (guild-filtered)
	scores, err := repo.GetRanking(2, guildID)
	if err != nil {
		t.Fatalf("GetRanking failed: %v", err)
	}
	if len(scores) != 1 {
		t.Fatalf("Expected 1 guild-filtered score, got: %d", len(scores))
	}
	if scores[0].Name != "RengokuChar" {
		t.Errorf("Expected 'RengokuChar' in guild ranking, got: %q", scores[0].Name)
	}
}

func TestRepoRengokuGetRankingPointsLeaderboard(t *testing.T) {
	repo, _, charID, _ := setupRengokuRepo(t)

	if err := repo.UpsertScore(charID, 10, 500, 5, 200); err != nil {
		t.Fatalf("UpsertScore failed: %v", err)
	}

	// Leaderboard 1 = max_points_mp (global)
	scores, err := repo.GetRanking(1, 0)
	if err != nil {
		t.Fatalf("GetRanking failed: %v", err)
	}
	if len(scores) != 1 {
		t.Fatalf("Expected 1 score, got: %d", len(scores))
	}
	if scores[0].Score != 500 {
		t.Errorf("Expected score=500 for points leaderboard, got: %d", scores[0].Score)
	}
}

func TestRepoRengokuGetRankingSPLeaderboard(t *testing.T) {
	repo, _, charID, _ := setupRengokuRepo(t)

	if err := repo.UpsertScore(charID, 10, 500, 5, 200); err != nil {
		t.Fatalf("UpsertScore failed: %v", err)
	}

	// Leaderboard 4 = max_stages_sp (global)
	scores, err := repo.GetRanking(4, 0)
	if err != nil {
		t.Fatalf("GetRanking failed: %v", err)
	}
	if len(scores) != 1 {
		t.Fatalf("Expected 1 score, got: %d", len(scores))
	}
	if scores[0].Score != 5 {
		t.Errorf("Expected score=5 for SP stages leaderboard, got: %d", scores[0].Score)
	}
}
