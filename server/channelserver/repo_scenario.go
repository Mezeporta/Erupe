package channelserver

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// ScenarioRepository centralizes all database access for the scenario_counter table.
type ScenarioRepository struct {
	db *sqlx.DB
}

// NewScenarioRepository creates a new ScenarioRepository.
func NewScenarioRepository(db *sqlx.DB) *ScenarioRepository {
	return &ScenarioRepository{db: db}
}

// GetCounters returns all scenario counters.
func (r *ScenarioRepository) GetCounters() ([]Scenario, error) {
	rows, err := r.db.Query("SELECT scenario_id, category_id FROM scenario_counter")
	if err != nil {
		return nil, fmt.Errorf("query scenario_counter: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var result []Scenario
	for rows.Next() {
		var s Scenario
		if err := rows.Scan(&s.MainID, &s.CategoryID); err != nil {
			return nil, fmt.Errorf("scan scenario_counter: %w", err)
		}
		result = append(result, s)
	}
	return result, rows.Err()
}
