package channelserver

import (
	"fmt"

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
func (r *MiscRepository) GetTrendWeapons(weaponType uint8) ([]uint16, error) {
	rows, err := r.db.Query("SELECT weapon_id FROM trend_weapons WHERE weapon_type=$1 ORDER BY count DESC LIMIT 3", weaponType)
	if err != nil {
		return nil, fmt.Errorf("query trend_weapons: %w", err)
	}
	defer rows.Close()
	var result []uint16
	for rows.Next() {
		var id uint16
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan trend_weapons: %w", err)
		}
		result = append(result, id)
	}
	return result, rows.Err()
}

// UpsertTrendWeapon increments the count for a weapon, inserting it if it doesn't exist.
func (r *MiscRepository) UpsertTrendWeapon(weaponID uint16, weaponType uint8) error {
	_, err := r.db.Exec(`INSERT INTO trend_weapons (weapon_id, weapon_type, count) VALUES ($1, $2, 1) ON CONFLICT (weapon_id) DO
		UPDATE SET count = trend_weapons.count+1`, weaponID, weaponType)
	return err
}
