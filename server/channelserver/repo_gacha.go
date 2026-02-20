package channelserver

import (
	"github.com/jmoiron/sqlx"
)

// GachaRepository centralizes all database access for gacha-related tables
// (gacha_shop, gacha_entries, gacha_items, gacha_stepup, gacha_box).
type GachaRepository struct {
	db *sqlx.DB
}

// NewGachaRepository creates a new GachaRepository.
func NewGachaRepository(db *sqlx.DB) *GachaRepository {
	return &GachaRepository{db: db}
}

// GetEntryForTransaction reads the cost type/amount and roll count for a gacha transaction.
func (r *GachaRepository) GetEntryForTransaction(gachaID uint32, rollID uint8) (itemType uint8, itemNumber uint16, rolls int, err error) {
	err = r.db.QueryRowx(
		`SELECT item_type, item_number, rolls FROM gacha_entries WHERE gacha_id = $1 AND entry_type = $2`,
		gachaID, rollID,
	).Scan(&itemType, &itemNumber, &rolls)
	return
}

// GetRewardPool returns the entry_type=100 reward pool for a gacha, ordered by weight descending.
func (r *GachaRepository) GetRewardPool(gachaID uint32) ([]GachaEntry, error) {
	var entries []GachaEntry
	rows, err := r.db.Queryx(
		`SELECT id, weight, rarity FROM gacha_entries WHERE gacha_id = $1 AND entry_type = 100 ORDER BY weight DESC`,
		gachaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry GachaEntry
		if err := rows.StructScan(&entry); err == nil {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// GetItemsForEntry returns the items associated with a gacha entry ID.
func (r *GachaRepository) GetItemsForEntry(entryID uint32) ([]GachaItem, error) {
	var items []GachaItem
	rows, err := r.db.Queryx(
		`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = $1`,
		entryID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item GachaItem
		if err := rows.StructScan(&item); err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// GetGuaranteedItems returns items for the entry matching a roll type and gacha ID.
func (r *GachaRepository) GetGuaranteedItems(rollType uint8, gachaID uint32) ([]GachaItem, error) {
	var items []GachaItem
	rows, err := r.db.Queryx(
		`SELECT item_type, item_id, quantity FROM gacha_items WHERE entry_id = (SELECT id FROM gacha_entries WHERE entry_type = $1 AND gacha_id = $2)`,
		rollType, gachaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var item GachaItem
		if err := rows.StructScan(&item); err == nil {
			items = append(items, item)
		}
	}
	return items, nil
}

// Stepup methods

// GetStepupStep returns the current stepup step for a character on a gacha.
func (r *GachaRepository) GetStepupStep(gachaID uint32, charID uint32) (uint8, error) {
	var step uint8
	err := r.db.QueryRow(
		`SELECT step FROM gacha_stepup WHERE gacha_id = $1 AND character_id = $2`,
		gachaID, charID,
	).Scan(&step)
	return step, err
}

// HasEntryType returns whether a gacha has any entries of the given type.
func (r *GachaRepository) HasEntryType(gachaID uint32, entryType uint8) (bool, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(1) FROM gacha_entries WHERE gacha_id = $1 AND entry_type = $2`,
		gachaID, entryType,
	).Scan(&count)
	return count > 0, err
}

// DeleteStepup removes the stepup state for a character on a gacha.
func (r *GachaRepository) DeleteStepup(gachaID uint32, charID uint32) error {
	_, err := r.db.Exec(
		`DELETE FROM gacha_stepup WHERE gacha_id = $1 AND character_id = $2`,
		gachaID, charID,
	)
	return err
}

// InsertStepup records a new stepup step for a character on a gacha.
func (r *GachaRepository) InsertStepup(gachaID uint32, step uint8, charID uint32) error {
	_, err := r.db.Exec(
		`INSERT INTO gacha_stepup (gacha_id, step, character_id) VALUES ($1, $2, $3)`,
		gachaID, step, charID,
	)
	return err
}

// Box gacha methods

// GetBoxEntryIDs returns the entry IDs already drawn for a box gacha.
func (r *GachaRepository) GetBoxEntryIDs(gachaID uint32, charID uint32) ([]uint32, error) {
	var ids []uint32
	rows, err := r.db.Queryx(
		`SELECT entry_id FROM gacha_box WHERE gacha_id = $1 AND character_id = $2`,
		gachaID, charID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id uint32
		if err := rows.Scan(&id); err == nil {
			ids = append(ids, id)
		}
	}
	return ids, nil
}

// InsertBoxEntry records a drawn entry in a box gacha.
func (r *GachaRepository) InsertBoxEntry(gachaID uint32, entryID uint32, charID uint32) error {
	_, err := r.db.Exec(
		`INSERT INTO gacha_box (gacha_id, entry_id, character_id) VALUES ($1, $2, $3)`,
		gachaID, entryID, charID,
	)
	return err
}

// DeleteBoxEntries resets all drawn entries for a box gacha.
func (r *GachaRepository) DeleteBoxEntries(gachaID uint32, charID uint32) error {
	_, err := r.db.Exec(
		`DELETE FROM gacha_box WHERE gacha_id = $1 AND character_id = $2`,
		gachaID, charID,
	)
	return err
}

// Shop listing methods

// ListShop returns all gacha shop definitions.
func (r *GachaRepository) ListShop() ([]Gacha, error) {
	var gachas []Gacha
	rows, err := r.db.Queryx(
		`SELECT id, min_gr, min_hr, name, url_banner, url_feature, url_thumbnail, wide, recommended, gacha_type, hidden FROM gacha_shop`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var g Gacha
		if err := rows.StructScan(&g); err == nil {
			gachas = append(gachas, g)
		}
	}
	return gachas, nil
}

// GetShopType returns the gacha_type for a gacha shop ID.
func (r *GachaRepository) GetShopType(shopID uint32) (int, error) {
	var gachaType int
	err := r.db.QueryRow(
		`SELECT gacha_type FROM gacha_shop WHERE id = $1`,
		shopID,
	).Scan(&gachaType)
	return gachaType, err
}

// GetAllEntries returns all entries for a gacha, ordered by weight descending.
func (r *GachaRepository) GetAllEntries(gachaID uint32) ([]GachaEntry, error) {
	var entries []GachaEntry
	rows, err := r.db.Queryx(
		`SELECT entry_type, id, item_type, item_number, item_quantity, weight, rarity, rolls, daily_limit, frontier_points, COALESCE(name, '') AS name FROM gacha_entries WHERE gacha_id = $1 ORDER BY weight DESC`,
		gachaID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var entry GachaEntry
		if err := rows.StructScan(&entry); err == nil {
			entries = append(entries, entry)
		}
	}
	return entries, nil
}

// GetWeightDivisor returns the total weight / 100000 for probability display.
func (r *GachaRepository) GetWeightDivisor(gachaID uint32) (float64, error) {
	var divisor float64
	err := r.db.QueryRow(
		`SELECT COALESCE(SUM(weight) / 100000.0, 0) AS chance FROM gacha_entries WHERE gacha_id = $1`,
		gachaID,
	).Scan(&divisor)
	return divisor, err
}
