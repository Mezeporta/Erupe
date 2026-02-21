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

	quests, err := repo.GetEventQuests()
	if err != nil {
		t.Fatalf("GetEventQuests failed: %v", err)
	}

	if len(quests) != 0 {
		t.Errorf("Expected no quests for empty event_quests table, got: %d", len(quests))
	}
}

func TestGetEventQuestsReturnsRows(t *testing.T) {
	repo, db := setupEventRepo(t)

	now := time.Now().Truncate(time.Microsecond)
	insertEventQuest(t, db, 1, 100, now, 0, 0)
	insertEventQuest(t, db, 2, 200, now, 7, 3)

	quests, err := repo.GetEventQuests()
	if err != nil {
		t.Fatalf("GetEventQuests failed: %v", err)
	}

	if len(quests) != 2 {
		t.Errorf("Expected 2 quests, got: %d", len(quests))
	}
	if quests[0].QuestID != 100 {
		t.Errorf("Expected first quest ID 100, got: %d", quests[0].QuestID)
	}
	if quests[1].QuestID != 200 {
		t.Errorf("Expected second quest ID 200, got: %d", quests[1].QuestID)
	}
	if quests[0].QuestType != 1 {
		t.Errorf("Expected first quest type 1, got: %d", quests[0].QuestType)
	}
	if quests[1].ActiveDays != 7 {
		t.Errorf("Expected second quest active_days 7, got: %d", quests[1].ActiveDays)
	}
	if quests[1].InactiveDays != 3 {
		t.Errorf("Expected second quest inactive_days 3, got: %d", quests[1].InactiveDays)
	}
}

func TestGetEventQuestsOrderByQuestID(t *testing.T) {
	repo, db := setupEventRepo(t)

	now := time.Now().Truncate(time.Microsecond)
	insertEventQuest(t, db, 1, 300, now, 0, 0)
	insertEventQuest(t, db, 1, 100, now, 0, 0)
	insertEventQuest(t, db, 1, 200, now, 0, 0)

	quests, err := repo.GetEventQuests()
	if err != nil {
		t.Fatalf("GetEventQuests failed: %v", err)
	}

	if len(quests) != 3 || quests[0].QuestID != 100 || quests[1].QuestID != 200 || quests[2].QuestID != 300 {
		ids := make([]int, len(quests))
		for i, q := range quests {
			ids[i] = q.QuestID
		}
		t.Errorf("Expected quest IDs [100, 200, 300], got: %v", ids)
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
