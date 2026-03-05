package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupScenarioRepo(t *testing.T) (*ScenarioRepository, *sqlx.DB) {
	t.Helper()
	db := SetupTestDB(t)
	repo := NewScenarioRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db
}

func TestRepoScenarioGetCountersEmpty(t *testing.T) {
	repo, _ := setupScenarioRepo(t)

	counters, err := repo.GetCounters()
	if err != nil {
		t.Fatalf("GetCounters failed: %v", err)
	}
	if len(counters) != 0 {
		t.Errorf("Expected 0 counters, got: %d", len(counters))
	}
}

func TestRepoScenarioGetCounters(t *testing.T) {
	repo, db := setupScenarioRepo(t)

	if _, err := db.Exec("INSERT INTO scenario_counter (id, scenario_id, category_id) VALUES (1, 100, 0)"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}
	if _, err := db.Exec("INSERT INTO scenario_counter (id, scenario_id, category_id) VALUES (2, 200, 1)"); err != nil {
		t.Fatalf("Setup failed: %v", err)
	}

	counters, err := repo.GetCounters()
	if err != nil {
		t.Fatalf("GetCounters failed: %v", err)
	}
	if len(counters) != 2 {
		t.Fatalf("Expected 2 counters, got: %d", len(counters))
	}

	// Check both values exist (order may vary)
	found100, found200 := false, false
	for _, c := range counters {
		if c.MainID == 100 {
			found100 = true
		}
		if c.MainID == 200 {
			found200 = true
		}
	}
	if !found100 || !found200 {
		t.Errorf("Expected scenario_ids 100 and 200, got: %+v", counters)
	}
}
