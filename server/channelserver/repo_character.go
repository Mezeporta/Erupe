package channelserver

import "github.com/jmoiron/sqlx"

// CharacterRepository centralizes all database access for the characters table.
type CharacterRepository struct {
	db *sqlx.DB
}

// NewCharacterRepository creates a new CharacterRepository.
func NewCharacterRepository(db *sqlx.DB) *CharacterRepository {
	return &CharacterRepository{db: db}
}

// LoadColumn reads a single []byte column by character ID.
func (r *CharacterRepository) LoadColumn(charID uint32, column string) ([]byte, error) {
	var data []byte
	err := r.db.QueryRow("SELECT "+column+" FROM characters WHERE id = $1", charID).Scan(&data)
	return data, err
}

// SaveColumn writes a single []byte column by character ID.
func (r *CharacterRepository) SaveColumn(charID uint32, column string, data []byte) error {
	_, err := r.db.Exec("UPDATE characters SET "+column+"=$1 WHERE id=$2", data, charID)
	return err
}

// ReadInt reads a single integer column (0 for NULL) by character ID.
func (r *CharacterRepository) ReadInt(charID uint32, column string) (int, error) {
	var value int
	err := r.db.QueryRow("SELECT COALESCE("+column+", 0) FROM characters WHERE id=$1", charID).Scan(&value)
	return value, err
}

// AdjustInt atomically adds delta to an integer column and returns the new value.
func (r *CharacterRepository) AdjustInt(charID uint32, column string, delta int) (int, error) {
	var value int
	err := r.db.QueryRow(
		"UPDATE characters SET "+column+"=COALESCE("+column+", 0)+$1 WHERE id=$2 RETURNING "+column,
		delta, charID,
	).Scan(&value)
	return value, err
}

// GetName returns the character name by ID.
func (r *CharacterRepository) GetName(charID uint32) (string, error) {
	var name string
	err := r.db.QueryRow("SELECT name FROM characters WHERE id=$1", charID).Scan(&name)
	return name, err
}

// GetUserID returns the owning user_id for a character.
func (r *CharacterRepository) GetUserID(charID uint32) (uint32, error) {
	var userID uint32
	err := r.db.QueryRow("SELECT user_id FROM characters WHERE id=$1", charID).Scan(&userID)
	return userID, err
}

// UpdateLastLogin sets the last_login timestamp.
func (r *CharacterRepository) UpdateLastLogin(charID uint32, timestamp int64) error {
	_, err := r.db.Exec("UPDATE characters SET last_login=$1 WHERE id=$2", timestamp, charID)
	return err
}

// UpdateTimePlayed sets the time_played value.
func (r *CharacterRepository) UpdateTimePlayed(charID uint32, timePlayed int) error {
	_, err := r.db.Exec("UPDATE characters SET time_played=$1 WHERE id=$2", timePlayed, charID)
	return err
}

// GetCharIDsByUserID returns all character IDs belonging to a user.
func (r *CharacterRepository) GetCharIDsByUserID(userID uint32) ([]uint32, error) {
	var ids []uint32
	err := r.db.Select(&ids, "SELECT id FROM characters WHERE user_id=$1", userID)
	return ids, err
}
