package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupDivaRepo(t *testing.T) (*DivaRepository, *sqlx.DB) {
	t.Helper()
	db := SetupTestDB(t)
	repo := NewDivaRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db
}

func TestRepoDivaInsertAndGetEvents(t *testing.T) {
	repo, _ := setupDivaRepo(t)

	if err := repo.InsertEvent(1700000000); err != nil {
		t.Fatalf("InsertEvent failed: %v", err)
	}

	events, err := repo.GetEvents()
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 1 {
		t.Fatalf("Expected 1 event, got: %d", len(events))
	}
	if events[0].StartTime != 1700000000 {
		t.Errorf("Expected start_time=1700000000, got: %d", events[0].StartTime)
	}
}

func TestRepoDivaGetEventsEmpty(t *testing.T) {
	repo, _ := setupDivaRepo(t)

	events, err := repo.GetEvents()
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events, got: %d", len(events))
	}
}

func TestRepoDivaDeleteEvents(t *testing.T) {
	repo, _ := setupDivaRepo(t)

	if err := repo.InsertEvent(1700000000); err != nil {
		t.Fatalf("InsertEvent failed: %v", err)
	}
	if err := repo.InsertEvent(1700100000); err != nil {
		t.Fatalf("InsertEvent failed: %v", err)
	}

	if err := repo.DeleteEvents(); err != nil {
		t.Fatalf("DeleteEvents failed: %v", err)
	}

	events, err := repo.GetEvents()
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 0 {
		t.Errorf("Expected 0 events after delete, got: %d", len(events))
	}
}

func TestRepoDivaMultipleEvents(t *testing.T) {
	repo, _ := setupDivaRepo(t)

	if err := repo.InsertEvent(1700000000); err != nil {
		t.Fatalf("InsertEvent 1 failed: %v", err)
	}
	if err := repo.InsertEvent(1700100000); err != nil {
		t.Fatalf("InsertEvent 2 failed: %v", err)
	}

	events, err := repo.GetEvents()
	if err != nil {
		t.Fatalf("GetEvents failed: %v", err)
	}
	if len(events) != 2 {
		t.Errorf("Expected 2 events, got: %d", len(events))
	}
}

func TestRepoDivaDeleteOnlyDivaEvents(t *testing.T) {
	repo, db := setupDivaRepo(t)

	// Insert a diva event
	if err := repo.InsertEvent(1700000000); err != nil {
		t.Fatalf("InsertEvent failed: %v", err)
	}
	// Insert a festa event (should not be deleted)
	if _, err := db.Exec("INSERT INTO events (event_type, start_time) VALUES ('festa', now())"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	if err := repo.DeleteEvents(); err != nil {
		t.Fatalf("DeleteEvents failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM events WHERE event_type='festa'").Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected festa event to survive, got count=%d", count)
	}
}
