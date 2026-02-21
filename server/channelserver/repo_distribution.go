package channelserver

import (
	"github.com/jmoiron/sqlx"
)

// DistributionRepository centralizes all database access for the distribution,
// distribution_items, and distributions_accepted tables.
type DistributionRepository struct {
	db *sqlx.DB
}

// NewDistributionRepository creates a new DistributionRepository.
func NewDistributionRepository(db *sqlx.DB) *DistributionRepository {
	return &DistributionRepository{db: db}
}

// List returns all distributions matching the given character and type.
func (r *DistributionRepository) List(charID uint32, distType uint8) ([]Distribution, error) {
	rows, err := r.db.Queryx(`
		SELECT d.id, event_name, description, COALESCE(rights, 0) AS rights, COALESCE(selection, false) AS selection, times_acceptable,
		COALESCE(min_hr, -1) AS min_hr, COALESCE(max_hr, -1) AS max_hr,
		COALESCE(min_sr, -1) AS min_sr, COALESCE(max_sr, -1) AS max_sr,
		COALESCE(min_gr, -1) AS min_gr, COALESCE(max_gr, -1) AS max_gr,
		(
    		SELECT count(*) FROM distributions_accepted da
    		WHERE d.id = da.distribution_id AND da.character_id = $1
		) AS times_accepted,
		COALESCE(deadline, TO_TIMESTAMP(0)) AS deadline
		FROM distribution d
		WHERE character_id = $1 AND type = $2 OR character_id IS NULL AND type = $2 ORDER BY id DESC
	`, charID, distType)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var dists []Distribution
	for rows.Next() {
		var d Distribution
		if err := rows.StructScan(&d); err != nil {
			continue
		}
		dists = append(dists, d)
	}
	return dists, nil
}

// GetItems returns all items for a given distribution.
func (r *DistributionRepository) GetItems(distributionID uint32) ([]DistributionItem, error) {
	rows, err := r.db.Queryx(`SELECT id, item_type, COALESCE(item_id, 0) AS item_id, COALESCE(quantity, 0) AS quantity FROM distribution_items WHERE distribution_id=$1`, distributionID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var items []DistributionItem
	for rows.Next() {
		var item DistributionItem
		if err := rows.StructScan(&item); err != nil {
			continue
		}
		items = append(items, item)
	}
	return items, nil
}

// RecordAccepted records that a character has accepted a distribution.
func (r *DistributionRepository) RecordAccepted(distributionID, charID uint32) error {
	_, err := r.db.Exec(`INSERT INTO public.distributions_accepted VALUES ($1, $2)`, distributionID, charID)
	return err
}

// GetDescription returns the description text for a distribution.
func (r *DistributionRepository) GetDescription(distributionID uint32) (string, error) {
	var desc string
	err := r.db.QueryRow("SELECT description FROM distribution WHERE id = $1", distributionID).Scan(&desc)
	return desc, err
}
