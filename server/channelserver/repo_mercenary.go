package channelserver

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

// MercenaryRepository centralizes database access for mercenary/rasta/airou sequences and queries.
type MercenaryRepository struct {
	db *sqlx.DB
}

// NewMercenaryRepository creates a new MercenaryRepository.
func NewMercenaryRepository(db *sqlx.DB) *MercenaryRepository {
	return &MercenaryRepository{db: db}
}

// NextRastaID returns the next value from the rasta_id_seq sequence.
func (r *MercenaryRepository) NextRastaID() (uint32, error) {
	var id uint32
	err := r.db.QueryRow("SELECT nextval('rasta_id_seq')").Scan(&id)
	return id, err
}

// NextAirouID returns the next value from the airou_id_seq sequence.
func (r *MercenaryRepository) NextAirouID() (uint32, error) {
	var id uint32
	err := r.db.QueryRow("SELECT nextval('airou_id_seq')").Scan(&id)
	return id, err
}

// GetMercenaryLoans returns characters that have a pact with the given character's rasta_id.
func (r *MercenaryRepository) GetMercenaryLoans(charID uint32) (*sql.Rows, error) {
	return r.db.Query("SELECT name, id, pact_id FROM characters WHERE pact_id=(SELECT rasta_id FROM characters WHERE id=$1)", charID)
}

// GetGuildHuntCatsUsed returns cats_used and start from guild_hunts for a given character.
func (r *MercenaryRepository) GetGuildHuntCatsUsed(charID uint32) (*sql.Rows, error) {
	return r.db.Query(`SELECT cats_used, start FROM guild_hunts gh
		INNER JOIN characters c ON gh.host_id = c.id WHERE c.id=$1`, charID)
}

// GetGuildAirou returns otomoairou data for all characters in a guild.
func (r *MercenaryRepository) GetGuildAirou(guildID uint32) (*sql.Rows, error) {
	return r.db.Query(`SELECT c.otomoairou FROM characters c
	INNER JOIN guild_characters gc ON gc.character_id = c.id
	WHERE gc.guild_id = $1 AND c.otomoairou IS NOT NULL
	ORDER BY c.id LIMIT 60`, guildID)
}
