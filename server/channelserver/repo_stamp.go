package channelserver

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// StampRepository centralizes all database access for the stamps table.
type StampRepository struct {
	db *sqlx.DB
}

// NewStampRepository creates a new StampRepository.
func NewStampRepository(db *sqlx.DB) *StampRepository {
	return &StampRepository{db: db}
}

// GetChecked returns the last check time for the given stamp type ("hl" or "ex").
func (r *StampRepository) GetChecked(charID uint32, stampType string) (time.Time, error) {
	var lastCheck time.Time
	err := r.db.QueryRow(fmt.Sprintf("SELECT %s_checked FROM stamps WHERE character_id=$1", stampType), charID).Scan(&lastCheck)
	return lastCheck, err
}

// Init inserts a new stamps record for a character with both check times set to now.
func (r *StampRepository) Init(charID uint32, now time.Time) error {
	_, err := r.db.Exec("INSERT INTO stamps (character_id, hl_checked, ex_checked) VALUES ($1, $2, $2)", charID, now)
	return err
}

// SetChecked updates the check time for a given stamp type.
func (r *StampRepository) SetChecked(charID uint32, stampType string, now time.Time) error {
	_, err := r.db.Exec(fmt.Sprintf(`UPDATE stamps SET %s_checked=$1 WHERE character_id=$2`, stampType), now, charID)
	return err
}

// IncrementTotal increments the total stamp count for a given stamp type.
func (r *StampRepository) IncrementTotal(charID uint32, stampType string) error {
	_, err := r.db.Exec(fmt.Sprintf("UPDATE stamps SET %s_total=%s_total+1 WHERE character_id=$1", stampType, stampType), charID)
	return err
}

// GetTotals returns the total and redeemed counts for a given stamp type.
func (r *StampRepository) GetTotals(charID uint32, stampType string) (total, redeemed uint16, err error) {
	err = r.db.QueryRow(fmt.Sprintf("SELECT %s_total, %s_redeemed FROM stamps WHERE character_id=$1", stampType, stampType), charID).Scan(&total, &redeemed)
	return
}

// ExchangeYearly performs a yearly stamp exchange, subtracting 48 from both hl_total and hl_redeemed.
func (r *StampRepository) ExchangeYearly(charID uint32) (total, redeemed uint16, err error) {
	err = r.db.QueryRow("UPDATE stamps SET hl_total=hl_total-48, hl_redeemed=hl_redeemed-48 WHERE character_id=$1 RETURNING hl_total, hl_redeemed", charID).Scan(&total, &redeemed)
	return
}

// Exchange performs a stamp exchange, adding 8 to the redeemed count for a given stamp type.
func (r *StampRepository) Exchange(charID uint32, stampType string) (total, redeemed uint16, err error) {
	err = r.db.QueryRow(fmt.Sprintf("UPDATE stamps SET %s_redeemed=%s_redeemed+8 WHERE character_id=$1 RETURNING %s_total, %s_redeemed", stampType, stampType, stampType, stampType), charID).Scan(&total, &redeemed)
	return
}

// GetMonthlyClaimed returns the last monthly item claim time for the given type.
func (r *StampRepository) GetMonthlyClaimed(charID uint32, monthlyType string) (time.Time, error) {
	var claimed time.Time
	err := r.db.QueryRow(
		fmt.Sprintf("SELECT %s_claimed FROM stamps WHERE character_id=$1", monthlyType), charID,
	).Scan(&claimed)
	return claimed, err
}

// SetMonthlyClaimed updates the monthly item claim time for the given type.
func (r *StampRepository) SetMonthlyClaimed(charID uint32, monthlyType string, now time.Time) error {
	_, err := r.db.Exec(
		fmt.Sprintf("UPDATE stamps SET %s_claimed=$1 WHERE character_id=$2", monthlyType), now, charID,
	)
	return err
}
