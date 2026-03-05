package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupAchievementRepo(t *testing.T) (*AchievementRepository, *sqlx.DB, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "ach_test_user")
	charID := CreateTestCharacter(t, db, userID, "AchChar")
	repo := NewAchievementRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, charID
}

func TestRepoAchievementEnsureExists(t *testing.T) {
	repo, db, charID := setupAchievementRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM achievements WHERE id=$1", charID).Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row, got: %d", count)
	}
}

func TestRepoAchievementEnsureExistsIdempotent(t *testing.T) {
	repo, db, charID := setupAchievementRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("First EnsureExists failed: %v", err)
	}
	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("Second EnsureExists failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM achievements WHERE id=$1", charID).Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 row after idempotent calls, got: %d", count)
	}
}

func TestRepoAchievementGetAllScores(t *testing.T) {
	repo, db, charID := setupAchievementRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	// Set some scores directly
	if _, err := db.Exec("UPDATE achievements SET ach0=10, ach5=42, ach32=99 WHERE id=$1", charID); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	scores, err := repo.GetAllScores(charID)
	if err != nil {
		t.Fatalf("GetAllScores failed: %v", err)
	}
	if scores[0] != 10 {
		t.Errorf("Expected ach0=10, got: %d", scores[0])
	}
	if scores[5] != 42 {
		t.Errorf("Expected ach5=42, got: %d", scores[5])
	}
	if scores[32] != 99 {
		t.Errorf("Expected ach32=99, got: %d", scores[32])
	}
}

func TestRepoAchievementGetAllScoresDefault(t *testing.T) {
	repo, _, charID := setupAchievementRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	scores, err := repo.GetAllScores(charID)
	if err != nil {
		t.Fatalf("GetAllScores failed: %v", err)
	}
	for i, s := range scores {
		if s != 0 {
			t.Errorf("Expected ach%d=0 by default, got: %d", i, s)
		}
	}
}

func TestRepoAchievementIncrementScore(t *testing.T) {
	repo, db, charID := setupAchievementRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	if err := repo.IncrementScore(charID, 5); err != nil {
		t.Fatalf("First IncrementScore failed: %v", err)
	}
	if err := repo.IncrementScore(charID, 5); err != nil {
		t.Fatalf("Second IncrementScore failed: %v", err)
	}

	var val int32
	if err := db.QueryRow("SELECT ach5 FROM achievements WHERE id=$1", charID).Scan(&val); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if val != 2 {
		t.Errorf("Expected ach5=2 after two increments, got: %d", val)
	}
}

func TestRepoAchievementIncrementScoreOutOfRange(t *testing.T) {
	repo, _, charID := setupAchievementRepo(t)

	if err := repo.EnsureExists(charID); err != nil {
		t.Fatalf("EnsureExists failed: %v", err)
	}

	err := repo.IncrementScore(charID, 33)
	if err == nil {
		t.Fatal("Expected error for achievementID=33, got nil")
	}
}
