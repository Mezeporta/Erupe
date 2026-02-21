package channelserver

import (
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
func (r *ScenarioRepository) GetCounters() (*sqlx.Rows, error) {
	return r.db.Queryx("SELECT scenario_id, category_id FROM scenario_counter")
}
