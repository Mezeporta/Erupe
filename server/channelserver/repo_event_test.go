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

func TestUpdateEventQuestStartTimes(t *testing.T) {
	repo, db := setupEventRepo(t)

	originalTime := time.Date(2025, 1, 1, 12, 0, 0, 0, time.UTC)
	id1 := insertEventQuest(t, db, 1, 100, originalTime, 7, 3)
	id2 := insertEventQuest(t, db, 2, 200, originalTime, 5, 2)

	newTime1 := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)
	newTime2 := time.Date(2025, 7, 20, 12, 0, 0, 0, time.UTC)

	err := repo.UpdateEventQuestStartTimes([]EventQuestUpdate{
		{ID: id1, StartTime: newTime1},
		{ID: id2, StartTime: newTime2},
	})
	if err != nil {
		t.Fatalf("UpdateEventQuestStartTimes failed: %v", err)
	}

	// Verify both updates
	var got1, got2 time.Time
	if err := db.QueryRow("SELECT start_time FROM event_quests WHERE id=$1", id1).Scan(&got1); err != nil {
		t.Fatalf("Verification query failed for id1: %v", err)
	}
	if !got1.Equal(newTime1) {
		t.Errorf("Expected start_time %v for id1, got: %v", newTime1, got1)
	}
	if err := db.QueryRow("SELECT start_time FROM event_quests WHERE id=$1", id2).Scan(&got2); err != nil {
		t.Fatalf("Verification query failed for id2: %v", err)
	}
	if !got2.Equal(newTime2) {
		t.Errorf("Expected start_time %v for id2, got: %v", newTime2, got2)
	}
}

func TestUpdateEventQuestStartTimesEmpty(t *testing.T) {
	repo, _ := setupEventRepo(t)

	// Empty slice should be a no-op
	err := repo.UpdateEventQuestStartTimes(nil)
	if err != nil {
		t.Fatalf("UpdateEventQuestStartTimes with nil should not error, got: %v", err)
	}

	err = repo.UpdateEventQuestStartTimes([]EventQuestUpdate{})
	if err != nil {
		t.Fatalf("UpdateEventQuestStartTimes with empty slice should not error, got: %v", err)
	}
}
