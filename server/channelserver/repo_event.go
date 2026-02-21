package channelserver

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

// EventQuest represents a row from the event_quests table.
type EventQuest struct {
	ID           uint32    `db:"id"`
	MaxPlayers   uint8     `db:"max_players"`
	QuestType    uint8     `db:"quest_type"`
	QuestID      int       `db:"quest_id"`
	Mark         uint32    `db:"mark"`
	Flags        int       `db:"flags"`
	StartTime    time.Time `db:"start_time"`
	ActiveDays   int       `db:"active_days"`
	InactiveDays int       `db:"inactive_days"`
}

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
func (r *EventRepository) GetLoginBoosts(charID uint32) ([]loginBoost, error) {
	var result []loginBoost
	err := r.db.Select(&result, "SELECT week_req, expiration, reset FROM login_boost WHERE char_id=$1 ORDER BY week_req", charID)
	return result, err
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

// GetEventQuests returns all event quest rows ordered by quest_id.
func (r *EventRepository) GetEventQuests() ([]EventQuest, error) {
	var result []EventQuest
	err := r.db.Select(&result, "SELECT id, COALESCE(max_players, 4) AS max_players, quest_type, quest_id, COALESCE(mark, 0) AS mark, COALESCE(flags, -1) AS flags, start_time, COALESCE(active_days, 0) AS active_days, COALESCE(inactive_days, 0) AS inactive_days FROM event_quests ORDER BY quest_id")
	return result, err
}

// UpdateEventQuestStartTime updates the start_time for an event quest within a transaction.
func (r *EventRepository) UpdateEventQuestStartTime(tx *sql.Tx, id uint32, startTime time.Time) error {
	_, err := tx.Exec("UPDATE event_quests SET start_time = $1 WHERE id = $2", startTime, id)
	return err
}

// BeginTx starts a new database transaction.
func (r *EventRepository) BeginTx() (*sql.Tx, error) {
	return r.db.Begin()
}
