package channelserver

import (
	"fmt"
	"time"

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

// MercenaryLoan represents a character that has a pact with a rasta.
type MercenaryLoan struct {
	Name   string
	CharID uint32
	PactID int
}

// GetMercenaryLoans returns characters that have a pact with the given character's rasta_id.
func (r *MercenaryRepository) GetMercenaryLoans(charID uint32) ([]MercenaryLoan, error) {
	rows, err := r.db.Query("SELECT name, id, pact_id FROM characters WHERE pact_id=(SELECT rasta_id FROM characters WHERE id=$1)", charID)
	if err != nil {
		return nil, fmt.Errorf("query mercenary loans: %w", err)
	}
	defer rows.Close()
	var result []MercenaryLoan
	for rows.Next() {
		var l MercenaryLoan
		if err := rows.Scan(&l.Name, &l.CharID, &l.PactID); err != nil {
			return nil, fmt.Errorf("scan mercenary loan: %w", err)
		}
		result = append(result, l)
	}
	return result, rows.Err()
}

// GuildHuntCatUsage represents cats_used and start time from a guild hunt.
type GuildHuntCatUsage struct {
	CatsUsed string
	Start    time.Time
}

// GetGuildHuntCatsUsed returns cats_used and start from guild_hunts for a given character.
func (r *MercenaryRepository) GetGuildHuntCatsUsed(charID uint32) ([]GuildHuntCatUsage, error) {
	rows, err := r.db.Query(`SELECT cats_used, start FROM guild_hunts gh
		INNER JOIN characters c ON gh.host_id = c.id WHERE c.id=$1`, charID)
	if err != nil {
		return nil, fmt.Errorf("query guild hunt cats: %w", err)
	}
	defer rows.Close()
	var result []GuildHuntCatUsage
	for rows.Next() {
		var u GuildHuntCatUsage
		if err := rows.Scan(&u.CatsUsed, &u.Start); err != nil {
			return nil, fmt.Errorf("scan guild hunt cat: %w", err)
		}
		result = append(result, u)
	}
	return result, rows.Err()
}

// GetGuildAirou returns otomoairou data for all characters in a guild.
func (r *MercenaryRepository) GetGuildAirou(guildID uint32) ([][]byte, error) {
	rows, err := r.db.Query(`SELECT c.otomoairou FROM characters c
	INNER JOIN guild_characters gc ON gc.character_id = c.id
	WHERE gc.guild_id = $1 AND c.otomoairou IS NOT NULL
	ORDER BY c.id LIMIT 60`, guildID)
	if err != nil {
		return nil, fmt.Errorf("query guild airou: %w", err)
	}
	defer rows.Close()
	var result [][]byte
	for rows.Next() {
		var data []byte
		if err := rows.Scan(&data); err != nil {
			return nil, fmt.Errorf("scan guild airou: %w", err)
		}
		result = append(result, data)
	}
	return result, rows.Err()
}
