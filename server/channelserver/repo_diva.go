package channelserver

import (
	"github.com/jmoiron/sqlx"
)

// DivaRepository centralizes all database access for diva defense events.
type DivaRepository struct {
	db *sqlx.DB
}

// NewDivaRepository creates a new DivaRepository.
func NewDivaRepository(db *sqlx.DB) *DivaRepository {
	return &DivaRepository{db: db}
}

// DeleteEvents removes all diva events.
func (r *DivaRepository) DeleteEvents() error {
	_, err := r.db.Exec("DELETE FROM events WHERE event_type='diva'")
	return err
}

// InsertEvent creates a new diva event with the given start epoch.
func (r *DivaRepository) InsertEvent(startEpoch uint32) error {
	_, err := r.db.Exec("INSERT INTO events (event_type, start_time) VALUES ('diva', to_timestamp($1)::timestamp without time zone)", startEpoch)
	return err
}

// GetEvents returns all diva events with their ID and start_time epoch.
func (r *DivaRepository) GetEvents() (*sqlx.Rows, error) {
	return r.db.Queryx("SELECT id, (EXTRACT(epoch FROM start_time)::int) as start_time FROM events WHERE event_type='diva'")
}
