package channelserver

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// MiscRepository centralizes database access for miscellaneous game tables.
type MiscRepository struct {
	db *sqlx.DB
}

// NewMiscRepository creates a new MiscRepository.
func NewMiscRepository(db *sqlx.DB) *MiscRepository {
	return &MiscRepository{db: db}
}

// GetTrendWeapons returns the top 3 weapon IDs for a given weapon type, ordered by count descending.
func (r *MiscRepository) GetTrendWeapons(weaponType uint8) (*sql.Rows, error) {
	return r.db.Query("SELECT weapon_id FROM trend_weapons WHERE weapon_type=$1 ORDER BY count DESC LIMIT 3", weaponType)
}

// UpsertTrendWeapon increments the count for a weapon, inserting it if it doesn't exist.
func (r *MiscRepository) UpsertTrendWeapon(weaponID uint16, weaponType uint8) error {
	_, err := r.db.Exec(`INSERT INTO trend_weapons (weapon_id, weapon_type, count) VALUES ($1, $2, 1) ON CONFLICT (weapon_id) DO
		UPDATE SET count = trend_weapons.count+1`, weaponID, weaponType)
	return err
}
