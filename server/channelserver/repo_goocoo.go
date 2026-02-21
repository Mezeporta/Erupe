package channelserver

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

// GoocooRepository centralizes all database access for the goocoo table.
type GoocooRepository struct {
	db *sqlx.DB
}

// NewGoocooRepository creates a new GoocooRepository.
func NewGoocooRepository(db *sqlx.DB) *GoocooRepository {
	return &GoocooRepository{db: db}
}

// validGoocooSlot validates the slot index to prevent SQL injection.
func validGoocooSlot(slot uint32) error {
	if slot > 4 {
		return fmt.Errorf("invalid goocoo slot index: %d", slot)
	}
	return nil
}

// EnsureExists creates a goocoo record if it doesn't already exist.
func (r *GoocooRepository) EnsureExists(charID uint32) error {
	_, err := r.db.Exec("INSERT INTO goocoo (id) VALUES ($1) ON CONFLICT DO NOTHING", charID)
	return err
}

// GetSlot reads a single goocoo slot by character ID and slot index (0-4).
func (r *GoocooRepository) GetSlot(charID uint32, slot uint32) ([]byte, error) {
	if err := validGoocooSlot(slot); err != nil {
		return nil, err
	}
	var data []byte
	err := r.db.QueryRow(fmt.Sprintf("SELECT goocoo%d FROM goocoo WHERE id=$1", slot), charID).Scan(&data)
	return data, err
}

// ClearSlot sets a goocoo slot to NULL.
func (r *GoocooRepository) ClearSlot(charID uint32, slot uint32) error {
	if err := validGoocooSlot(slot); err != nil {
		return err
	}
	_, err := r.db.Exec(fmt.Sprintf("UPDATE goocoo SET goocoo%d=NULL WHERE id=$1", slot), charID)
	return err
}

// SaveSlot writes data to a goocoo slot.
func (r *GoocooRepository) SaveSlot(charID uint32, slot uint32, data []byte) error {
	if err := validGoocooSlot(slot); err != nil {
		return err
	}
	_, err := r.db.Exec(fmt.Sprintf("UPDATE goocoo SET goocoo%d=$1 WHERE id=$2", slot), data, charID)
	return err
}
