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

// DivaEvent represents a diva event row with ID and start_time epoch.
type DivaEvent struct {
	ID        uint32 `db:"id"`
	StartTime uint32 `db:"start_time"`
}

// GetEvents returns all diva events with their ID and start_time epoch.
func (r *DivaRepository) GetEvents() ([]DivaEvent, error) {
	var result []DivaEvent
	err := r.db.Select(&result, "SELECT id, (EXTRACT(epoch FROM start_time)::int) as start_time FROM events WHERE event_type='diva'")
	return result, err
}

// AddPoints atomically adds quest and bonus points for a character in a diva event.
func (r *DivaRepository) AddPoints(charID, eventID, questPoints, bonusPoints uint32) error {
	_, err := r.db.Exec(`
		INSERT INTO diva_points (char_id, event_id, quest_points, bonus_points, updated_at)
		VALUES ($1, $2, $3, $4, now())
		ON CONFLICT (char_id, event_id) DO UPDATE
		SET quest_points = diva_points.quest_points + EXCLUDED.quest_points,
		    bonus_points = diva_points.bonus_points + EXCLUDED.bonus_points,
		    updated_at = now()`,
		charID, eventID, questPoints, bonusPoints)
	return err
}

// GetPoints returns the accumulated quest and bonus points for a character in an event.
func (r *DivaRepository) GetPoints(charID, eventID uint32) (int64, int64, error) {
	var qp, bp int64
	err := r.db.QueryRow(
		"SELECT quest_points, bonus_points FROM diva_points WHERE char_id=$1 AND event_id=$2",
		charID, eventID).Scan(&qp, &bp)
	if err != nil {
		return 0, 0, err
	}
	return qp, bp, nil
}

// GetTotalPoints returns the sum of all players' quest and bonus points for an event.
func (r *DivaRepository) GetTotalPoints(eventID uint32) (int64, int64, error) {
	var qp, bp int64
	err := r.db.QueryRow(
		"SELECT COALESCE(SUM(quest_points),0), COALESCE(SUM(bonus_points),0) FROM diva_points WHERE event_id=$1",
		eventID).Scan(&qp, &bp)
	if err != nil {
		return 0, 0, err
	}
	return qp, bp, nil
}
