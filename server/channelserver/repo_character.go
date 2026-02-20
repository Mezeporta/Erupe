package channelserver

import (
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

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

// ReadTime reads a single time.Time column by character ID.
// Returns the provided default if the column is NULL.
func (r *CharacterRepository) ReadTime(charID uint32, column string, defaultVal time.Time) (time.Time, error) {
	var t sql.NullTime
	err := r.db.QueryRow("SELECT "+column+" FROM characters WHERE id=$1", charID).Scan(&t)
	if err != nil {
		return defaultVal, err
	}
	if !t.Valid {
		return defaultVal, nil
	}
	return t.Time, nil
}

// SaveTime writes a single time.Time column by character ID.
func (r *CharacterRepository) SaveTime(charID uint32, column string, value time.Time) error {
	_, err := r.db.Exec("UPDATE characters SET "+column+"=$1 WHERE id=$2", value, charID)
	return err
}

// SaveInt writes a single integer column by character ID.
func (r *CharacterRepository) SaveInt(charID uint32, column string, value int) error {
	_, err := r.db.Exec("UPDATE characters SET "+column+"=$1 WHERE id=$2", value, charID)
	return err
}

// SaveBool writes a single boolean column by character ID.
func (r *CharacterRepository) SaveBool(charID uint32, column string, value bool) error {
	_, err := r.db.Exec("UPDATE characters SET "+column+"=$1 WHERE id=$2", value, charID)
	return err
}

// SaveString writes a single string column by character ID.
func (r *CharacterRepository) SaveString(charID uint32, column string, value string) error {
	_, err := r.db.Exec("UPDATE characters SET "+column+"=$1 WHERE id=$2", value, charID)
	return err
}

// ReadBool reads a single boolean column by character ID.
func (r *CharacterRepository) ReadBool(charID uint32, column string) (bool, error) {
	var value bool
	err := r.db.QueryRow("SELECT "+column+" FROM characters WHERE id=$1", charID).Scan(&value)
	return value, err
}

// ReadString reads a single string column by character ID (empty string for NULL).
func (r *CharacterRepository) ReadString(charID uint32, column string) (string, error) {
	var value sql.NullString
	err := r.db.QueryRow("SELECT "+column+" FROM characters WHERE id=$1", charID).Scan(&value)
	if err != nil {
		return "", err
	}
	return value.String, nil
}

// LoadColumnWithDefault reads a []byte column, returning defaultVal if NULL.
func (r *CharacterRepository) LoadColumnWithDefault(charID uint32, column string, defaultVal []byte) ([]byte, error) {
	var data []byte
	err := r.db.QueryRow("SELECT "+column+" FROM characters WHERE id=$1", charID).Scan(&data)
	if err != nil {
		return defaultVal, err
	}
	if data == nil {
		return defaultVal, nil
	}
	return data, nil
}

// SetDeleted marks a character as deleted.
func (r *CharacterRepository) SetDeleted(charID uint32) error {
	_, err := r.db.Exec("UPDATE characters SET deleted=true WHERE id=$1", charID)
	return err
}

// UpdateDailyCafe sets daily_time, bonus_quests, and daily_quests atomically.
func (r *CharacterRepository) UpdateDailyCafe(charID uint32, dailyTime time.Time, bonusQuests, dailyQuests uint32) error {
	_, err := r.db.Exec("UPDATE characters SET daily_time=$1, bonus_quests=$2, daily_quests=$3 WHERE id=$4",
		dailyTime, bonusQuests, dailyQuests, charID)
	return err
}

// ResetDailyQuests zeroes bonus_quests and daily_quests.
func (r *CharacterRepository) ResetDailyQuests(charID uint32) error {
	_, err := r.db.Exec("UPDATE characters SET bonus_quests=0, daily_quests=0 WHERE id=$1", charID)
	return err
}

// ReadEtcPoints reads bonus_quests, daily_quests, and promo_points.
func (r *CharacterRepository) ReadEtcPoints(charID uint32) (bonusQuests, dailyQuests, promoPoints uint32, err error) {
	err = r.db.QueryRow("SELECT bonus_quests, daily_quests, promo_points FROM characters WHERE id=$1", charID).
		Scan(&bonusQuests, &dailyQuests, &promoPoints)
	return
}

// ResetCafeTime zeroes cafe_time and sets cafe_reset.
func (r *CharacterRepository) ResetCafeTime(charID uint32, cafeReset time.Time) error {
	_, err := r.db.Exec("UPDATE characters SET cafe_time=0, cafe_reset=$1 WHERE id=$2", cafeReset, charID)
	return err
}

// UpdateGuildPostChecked sets guild_post_checked to now().
func (r *CharacterRepository) UpdateGuildPostChecked(charID uint32) error {
	_, err := r.db.Exec("UPDATE characters SET guild_post_checked=now() WHERE id=$1", charID)
	return err
}

// ReadGuildPostChecked reads guild_post_checked timestamp.
func (r *CharacterRepository) ReadGuildPostChecked(charID uint32) (time.Time, error) {
	var t time.Time
	err := r.db.QueryRow("SELECT guild_post_checked FROM characters WHERE id=$1", charID).Scan(&t)
	return t, err
}

// SaveMercenary updates savemercenary and rasta_id atomically.
func (r *CharacterRepository) SaveMercenary(charID uint32, data []byte, rastaID uint32) error {
	_, err := r.db.Exec("UPDATE characters SET savemercenary=$1, rasta_id=$2 WHERE id=$3", data, rastaID, charID)
	return err
}

// UpdateGCPAndPact updates gcp and pact_id atomically.
func (r *CharacterRepository) UpdateGCPAndPact(charID uint32, gcp uint32, pactID uint32) error {
	_, err := r.db.Exec("UPDATE characters SET gcp=$1, pact_id=$2 WHERE id=$3", gcp, pactID, charID)
	return err
}

// FindByRastaID looks up name and id by rasta_id.
func (r *CharacterRepository) FindByRastaID(rastaID int) (charID uint32, name string, err error) {
	err = r.db.QueryRow("SELECT name, id FROM characters WHERE rasta_id=$1", rastaID).Scan(&name, &charID)
	return
}
