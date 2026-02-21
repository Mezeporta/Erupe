package channelserver

import (
	"github.com/jmoiron/sqlx"
)

// CafeRepository centralizes all database access for cafe-related tables.
type CafeRepository struct {
	db *sqlx.DB
}

// NewCafeRepository creates a new CafeRepository.
func NewCafeRepository(db *sqlx.DB) *CafeRepository {
	return &CafeRepository{db: db}
}

// ResetAccepted deletes all accepted cafe bonuses for a character.
func (r *CafeRepository) ResetAccepted(charID uint32) error {
	_, err := r.db.Exec(`DELETE FROM cafe_accepted WHERE character_id=$1`, charID)
	return err
}

// GetBonuses returns all cafe bonuses with their claimed status for a character.
func (r *CafeRepository) GetBonuses(charID uint32) (*sqlx.Rows, error) {
	return r.db.Queryx(`
	SELECT cb.id, time_req, item_type, item_id, quantity,
	(
		SELECT count(*)
		FROM cafe_accepted ca
		WHERE cb.id = ca.cafe_id AND ca.character_id = $1
	)::int::bool AS claimed
	FROM cafebonus cb ORDER BY id ASC;`, charID)
}

// GetClaimable returns unclaimed cafe bonuses where the character has enough accumulated time.
func (r *CafeRepository) GetClaimable(charID uint32, elapsedSec int64) (*sqlx.Rows, error) {
	return r.db.Queryx(`
	SELECT c.id, time_req, item_type, item_id, quantity
	FROM cafebonus c
	WHERE (
		SELECT count(*)
		FROM cafe_accepted ca
		WHERE c.id = ca.cafe_id AND ca.character_id = $1
	) < 1 AND (
		SELECT ch.cafe_time + $2
		FROM characters ch
		WHERE ch.id = $1
	) >= time_req`, charID, elapsedSec)
}

// GetBonusItem returns the item type and quantity for a specific cafe bonus.
func (r *CafeRepository) GetBonusItem(bonusID uint32) (itemType, quantity uint32, err error) {
	err = r.db.QueryRow(`SELECT cb.id, item_type, quantity FROM cafebonus cb WHERE cb.id=$1`, bonusID).Scan(&bonusID, &itemType, &quantity)
	return
}

// AcceptBonus records that a character has accepted a cafe bonus.
func (r *CafeRepository) AcceptBonus(bonusID, charID uint32) error {
	_, err := r.db.Exec("INSERT INTO cafe_accepted VALUES ($1, $2)", bonusID, charID)
	return err
}
