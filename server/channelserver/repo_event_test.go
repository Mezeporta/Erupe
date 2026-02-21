package channelserver

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func setupEventRepo(t *testing.T) (*EventRepository, *sqlx.DB) {
	t.Helper()
	db := SetupTestDB(t)
	repo := NewEventRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db
}

func insertEventQuest(t *testing.T, db *sqlx.DB, questType, questID int, startTime time.Time, activeDays, inactiveDays int) uint32 {
	t.Helper()
	var id uint32
	err := db.QueryRow(
		`INSERT INTO event_quests (quest_type, quest_id, start_time, active_days, inactive_days)
		 VALUES ($1, $2, $3, $4, $5) RETURNING id`,
		questType, questID, startTime, activeDays, inactiveDays,
	).Scan(&id)
	if err != nil {
		t.Fatalf("Failed to insert event quest: %v", err)
	}
	return id
}

func TestGetEventQuestsEmpty(t *testing.T) {
	repo, _ := setupEventRepo(t)

	rows, err := repo.GetEventQuests()
	if err != nil {
		t.Fatalf("GetEventQuests failed: %v", err)
	}
	defer rows.Close()

	if rows.Next() {
		t.Error("Expected no rows for empty event_quests table")
	}
}

func TestGetEventQuestsReturnsRows(t *testing.T) {
	repo, db := setupEventRepo(t)

	now := time.Now().Truncate(time.Microsecond)
	insertEventQuest(t, db, 1, 100, now, 0, 0)
	insertEventQuest(t, db, 2, 200, now, 7, 3)

	rows, err := repo.GetEventQuests()
	if err != nil {
		t.Fatalf("GetEventQuests failed: %v", err)
	}
	defer rows.Close()

	count := 0
	for rows.Next() {
		var id, mark uint32
		var questID, flags, activeDays, inactiveDays int
		var maxPlayers, questType uint8
		var startTime time.Time
		if err := rows.Scan(&id, &maxPlayers, &questType, &questID, &mark, &flags, &startTime, &activeDays, &inactiveDays); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		count++
	}
	if count != 2 {
		t.Errorf("Expected 2 rows, got: %d", count)
	}
}

func TestGetEventQuestsOrderByQuestID(t *testing.T) {
	repo, db := setupEventRepo(t)

	now := time.Now().Truncate(time.Microsecond)
	insertEventQuest(t, db, 1, 300, now, 0, 0)
	insertEventQuest(t, db, 1, 100, now, 0, 0)
	insertEventQuest(t, db, 1, 200, now, 0, 0)

	rows, err := repo.GetEventQuests()
	if err != nil {
		t.Fatalf("GetEventQuests failed: %v", err)
	}
	defer rows.Close()

	var questIDs []int
	for rows.Next() {
		var id, mark uint32
		var questID, flags, activeDays, inactiveDays int
		var maxPlayers, questType uint8
		var startTime time.Time
		if err := rows.Scan(&id, &maxPlayers, &questType, &questID, &mark, &flags, &startTime, &activeDays, &inactiveDays); err != nil {
			t.Fatalf("Scan failed: %v", err)
		}
		questIDs = append(questIDs, questID)
	}
	if len(questIDs) != 3 || questIDs[0] != 100 || questIDs[1] != 200 || questIDs[2] != 300 {
		t.Errorf("Expected quest IDs [100, 200, 300], got: %v", questIDs)
	}
}

func TestBeginTxAndUpdateEventQuestStartTime(t *testing.T) {
	repo, db := setupEventRepo(t)

	originalTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	questID := insertEventQuest(t, db, 1, 100, originalTime, 7, 3)

	newTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tx, err := repo.BeginTx()
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	if err := repo.UpdateEventQuestStartTime(tx, questID, newTime); err != nil {
		_ = tx.Rollback()
		t.Fatalf("UpdateEventQuestStartTime failed: %v", err)
	}

	if err := tx.Commit(); err != nil {
		t.Fatalf("Commit failed: %v", err)
	}

	// Verify the update
	var got time.Time
	if err := db.QueryRow("SELECT start_time FROM event_quests WHERE id=$1", questID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !got.Equal(newTime) {
		t.Errorf("Expected start_time %v, got: %v", newTime, got)
	}
}

func TestUpdateEventQuestStartTimeRollback(t *testing.T) {
	repo, db := setupEventRepo(t)

	originalTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	questID := insertEventQuest(t, db, 1, 100, originalTime, 0, 0)

	tx, err := repo.BeginTx()
	if err != nil {
		t.Fatalf("BeginTx failed: %v", err)
	}

	newTime := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	if err := repo.UpdateEventQuestStartTime(tx, questID, newTime); err != nil {
		t.Fatalf("UpdateEventQuestStartTime failed: %v", err)
	}

	// Rollback instead of commit
	if err := tx.Rollback(); err != nil {
		t.Fatalf("Rollback failed: %v", err)
	}

	// Verify original time is preserved
	var got time.Time
	if err := db.QueryRow("SELECT start_time FROM event_quests WHERE id=$1", questID).Scan(&got); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !got.Equal(originalTime) {
		t.Errorf("Expected original start_time %v after rollback, got: %v", originalTime, got)
	}
}
