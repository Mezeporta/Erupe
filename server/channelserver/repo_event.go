package channelserver

import (
	"time"

	"github.com/jmoiron/sqlx"
)

// EventRepository centralizes all database access for event-related tables.
type EventRepository struct {
	db *sqlx.DB
}

// NewEventRepository creates a new EventRepository.
func NewEventRepository(db *sqlx.DB) *EventRepository {
	return &EventRepository{db: db}
}

// GetFeatureWeapon returns the featured weapon bitfield for a given start time.
func (r *EventRepository) GetFeatureWeapon(startTime time.Time) (activeFeature, error) {
	var af activeFeature
	err := r.db.QueryRowx(`SELECT start_time, featured FROM feature_weapon WHERE start_time=$1`, startTime).StructScan(&af)
	return af, err
}

// InsertFeatureWeapon stores a new featured weapon entry.
func (r *EventRepository) InsertFeatureWeapon(startTime time.Time, features uint32) error {
	_, err := r.db.Exec(`INSERT INTO feature_weapon VALUES ($1, $2)`, startTime, features)
	return err
}

// GetLoginBoosts returns all login boost rows for a character, ordered by week_req.
func (r *EventRepository) GetLoginBoosts(charID uint32) (*sqlx.Rows, error) {
	return r.db.Queryx("SELECT week_req, expiration, reset FROM login_boost WHERE char_id=$1 ORDER BY week_req", charID)
}

// InsertLoginBoost creates a new login boost entry.
func (r *EventRepository) InsertLoginBoost(charID uint32, weekReq uint8, expiration, reset time.Time) error {
	_, err := r.db.Exec(`INSERT INTO login_boost VALUES ($1, $2, $3, $4)`, charID, weekReq, expiration, reset)
	return err
}

// UpdateLoginBoost updates expiration and reset for a login boost entry.
func (r *EventRepository) UpdateLoginBoost(charID uint32, weekReq uint8, expiration, reset time.Time) error {
	_, err := r.db.Exec(`UPDATE login_boost SET expiration=$1, reset=$2 WHERE char_id=$3 AND week_req=$4`, expiration, reset, charID, weekReq)
	return err
}
