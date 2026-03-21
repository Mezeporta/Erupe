package channelserver

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// SaveAtomicParams bundles all fields needed for an atomic save transaction.
type SaveAtomicParams struct {
	CharID     uint32
	CompSave   []byte
	Hash       []byte // SHA-256 of decompressed savedata
	HR         uint16
	GR         uint16
	IsFemale   bool
	WeaponType uint8
	WeaponID   uint16

	// House data (written to user_binary)
	HouseTier     []byte
	HouseData     []byte
	BookshelfData []byte
	GalleryData   []byte
	ToreData      []byte
	GardenData    []byte

	// Optional backup (nil means skip)
	BackupSlot int
	BackupData []byte
}

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

// ErrCharacterNotFound is returned by write methods when no character row is matched.
var ErrCharacterNotFound = errors.New("character not found")

// SaveColumn writes a single []byte column by character ID.
// Returns ErrCharacterNotFound if no row was updated (character does not exist).
func (r *CharacterRepository) SaveColumn(charID uint32, column string, data []byte) error {
	result, err := r.db.Exec("UPDATE characters SET "+column+"=$1 WHERE id=$2", data, charID)
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("SaveColumn %s for char %d: %w", column, charID, ErrCharacterNotFound)
	}
	return nil
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

// SaveMercenary updates savemercenary and optionally rasta_id.
// When rastaID is 0, only the mercenary blob is saved — the existing rasta_id
// (typically NULL for characters without a mercenary) is preserved. Writing 0
// would pollute GetMercenaryLoans queries that match on pact_id.
// Returns ErrCharacterNotFound if no row was updated.
func (r *CharacterRepository) SaveMercenary(charID uint32, data []byte, rastaID uint32) error {
	var result sql.Result
	var err error
	if rastaID == 0 {
		result, err = r.db.Exec("UPDATE characters SET savemercenary=$1 WHERE id=$2", data, charID)
	} else {
		result, err = r.db.Exec("UPDATE characters SET savemercenary=$1, rasta_id=$2 WHERE id=$3", data, rastaID, charID)
	}
	if err != nil {
		return err
	}
	n, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return fmt.Errorf("SaveMercenary for char %d: %w", charID, ErrCharacterNotFound)
	}
	return nil
}

// UpdateGCPAndPact updates gcp and pact_id atomically.
func (r *CharacterRepository) UpdateGCPAndPact(charID uint32, gcp uint32, pactID uint32) error {
	_, err := r.db.Exec("UPDATE characters SET gcp=$1, pact_id=$2 WHERE id=$3", gcp, pactID, charID)
	return err
}

// SavedataBackup holds one row from the savedata_backups table.
type SavedataBackup struct {
	Slot    int
	Data    []byte
	SavedAt time.Time
}

// LoadBackupsByRecency returns all backup slots for a character, ordered
// most-recent first. Returns an empty (non-nil) slice if no backups exist.
func (r *CharacterRepository) LoadBackupsByRecency(charID uint32) ([]SavedataBackup, error) {
	rows, err := r.db.Query(
		`SELECT slot, savedata, saved_at FROM savedata_backups
		 WHERE char_id = $1
		 ORDER BY saved_at DESC`,
		charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close() //nolint:errcheck // rows.Close error is non-actionable here

	backups := make([]SavedataBackup, 0)
	for rows.Next() {
		var b SavedataBackup
		if err := rows.Scan(&b.Slot, &b.Data, &b.SavedAt); err != nil {
			return nil, err
		}
		backups = append(backups, b)
	}
	return backups, rows.Err()
}

// SaveBackup upserts a savedata snapshot into the rotating backup table.
func (r *CharacterRepository) SaveBackup(charID uint32, slot int, data []byte) error {
	_, err := r.db.Exec(`
		INSERT INTO savedata_backups (char_id, slot, savedata, saved_at)
		VALUES ($1, $2, $3, now())
		ON CONFLICT (char_id, slot) DO UPDATE SET savedata = $3, saved_at = now()
	`, charID, slot, data)
	return err
}

// GetLastBackupTime returns the most recent backup timestamp for a character.
// Returns the zero time if no backups exist.
func (r *CharacterRepository) GetLastBackupTime(charID uint32) (time.Time, error) {
	var t sql.NullTime
	err := r.db.QueryRow(
		"SELECT MAX(saved_at) FROM savedata_backups WHERE char_id = $1", charID,
	).Scan(&t)
	if err != nil {
		return time.Time{}, err
	}
	if !t.Valid {
		return time.Time{}, nil
	}
	return t.Time, nil
}

// SaveCharacterDataAtomic performs all save-related writes in a single
// database transaction. If any step fails, everything is rolled back.
func (r *CharacterRepository) SaveCharacterDataAtomic(params SaveAtomicParams) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback() //nolint:errcheck // rollback is no-op after commit

	// 1. Save character data + hash
	if _, err := tx.Exec(
		`UPDATE characters SET savedata=$1, savedata_hash=$2, is_new_character=false, hr=$3, gr=$4, is_female=$5, weapon_type=$6, weapon_id=$7 WHERE id=$8`,
		params.CompSave, params.Hash, params.HR, params.GR, params.IsFemale, params.WeaponType, params.WeaponID, params.CharID,
	); err != nil {
		return fmt.Errorf("save character data: %w", err)
	}

	// 2. Save house data
	if _, err := tx.Exec(
		`UPDATE user_binary SET house_tier=$1, house_data=$2, bookshelf=$3, gallery=$4, tore=$5, garden=$6 WHERE id=$7`,
		params.HouseTier, params.HouseData, params.BookshelfData, params.GalleryData, params.ToreData, params.GardenData, params.CharID,
	); err != nil {
		return fmt.Errorf("save house data: %w", err)
	}

	// 3. Optional backup
	if params.BackupData != nil {
		if _, err := tx.Exec(
			`INSERT INTO savedata_backups (char_id, slot, savedata, saved_at)
			 VALUES ($1, $2, $3, now())
			 ON CONFLICT (char_id, slot) DO UPDATE SET savedata = $3, saved_at = now()`,
			params.CharID, params.BackupSlot, params.BackupData,
		); err != nil {
			return fmt.Errorf("save backup: %w", err)
		}
	}

	return tx.Commit()
}

// LoadSaveDataWithHash reads the core save columns plus the integrity hash.
// The hash may be nil for characters saved before checksums were introduced.
func (r *CharacterRepository) LoadSaveDataWithHash(charID uint32) (uint32, []byte, bool, string, []byte, error) {
	var id uint32
	var savedata []byte
	var isNew bool
	var name string
	var hash []byte
	err := r.db.QueryRow(
		"SELECT id, savedata, is_new_character, name, savedata_hash FROM characters WHERE id = $1", charID,
	).Scan(&id, &savedata, &isNew, &name, &hash)
	return id, savedata, isNew, name, hash, err
}

// FindByRastaID looks up name and id by rasta_id.
func (r *CharacterRepository) FindByRastaID(rastaID int) (charID uint32, name string, err error) {
	err = r.db.QueryRow("SELECT name, id FROM characters WHERE rasta_id=$1", rastaID).Scan(&name, &charID)
	return
}

// SaveCharacterData updates the core save fields on a character.
func (r *CharacterRepository) SaveCharacterData(charID uint32, compSave []byte, hr, gr uint16, isFemale bool, weaponType uint8, weaponID uint16) error {
	_, err := r.db.Exec(`UPDATE characters SET savedata=$1, is_new_character=false, hr=$2, gr=$3, is_female=$4, weapon_type=$5, weapon_id=$6 WHERE id=$7`,
		compSave, hr, gr, isFemale, weaponType, weaponID, charID)
	return err
}

// SaveHouseData updates house-related fields in user_binary.
func (r *CharacterRepository) SaveHouseData(charID uint32, houseTier []byte, houseData, bookshelf, gallery, tore, garden []byte) error {
	_, err := r.db.Exec(`UPDATE user_binary SET house_tier=$1, house_data=$2, bookshelf=$3, gallery=$4, tore=$5, garden=$6 WHERE id=$7`,
		houseTier, houseData, bookshelf, gallery, tore, garden, charID)
	return err
}

// LoadSaveData reads the core save columns for a character.
// Returns charID, savedata, isNewCharacter, name, and any error.
func (r *CharacterRepository) LoadSaveData(charID uint32) (uint32, []byte, bool, string, error) {
	var id uint32
	var savedata []byte
	var isNew bool
	var name string
	err := r.db.QueryRow("SELECT id, savedata, is_new_character, name FROM characters WHERE id = $1", charID).
		Scan(&id, &savedata, &isNew, &name)
	return id, savedata, isNew, name, err
}
