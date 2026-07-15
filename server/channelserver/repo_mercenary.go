package channelserver

import (
	"database/sql" // ADD THIS
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
	Name         string
	CharID       uint32
	PactID       int
	ContractDate sql.NullTime
}

// GetMercenaryLoans returns characters that have a pact with the given character's rasta_id.
func (r *MercenaryRepository) GetMercenaryLoans(charID uint32) ([]MercenaryLoan, error) {
	rows, err := r.db.Query("SELECT name, id, pact_id, mercenary_contract_date FROM characters WHERE pact_id=(SELECT rasta_id FROM characters WHERE id=$1)", charID)
	if err != nil {
		return nil, fmt.Errorf("query mercenary loans: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var result []MercenaryLoan
	for rows.Next() {
		var l MercenaryLoan
		if err := rows.Scan(&l.Name, &l.CharID, &l.PactID, &l.ContractDate); err != nil {
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
	defer func() { _ = rows.Close() }()
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
	defer func() { _ = rows.Close() }()
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

// LogMercenaryEvent inserts a row into mercenary_logs recording a contract state change.
func (r *MercenaryRepository) LogMercenaryEvent(playerID, mercenaryID uint32, logTime time.Time, isInitiator bool, state int) error {
	_, err := r.db.Exec(`
		INSERT INTO mercenary_logs(player_id, mercenary_id, log_time, is_initiator, state)
		VALUES ($1, $2, $3, $4, $5)
	`, playerID, mercenaryID, logTime, isInitiator, state)
	return err
}

// MercenaryKillLog represents an aggregated kill-log entry for a mercenary reward claim.
type MercenaryKillLog struct {
	ID      uint32 `db:"id"`
	Monster uint32 `db:"monster"`
}

// GetLastMercenaryReward returns the last claimed kill_log_id and claim time for a mercenary.
// If no reward row exists, sql.ErrNoRows is returned so callers can distinguish "first claim".
func (r *MercenaryRepository) GetLastMercenaryReward(mercenaryID uint32, defaultClaim time.Time) (lastKillLogID uint32, lastClaimAt time.Time, err error) {
	err = r.db.QueryRow(`
		SELECT COALESCE(last_kill_log_id, 0), COALESCE(last_claim_at, $2)
		FROM mercenary_rewards
		WHERE mercenary_id = $1
		ORDER BY last_claim_at DESC
		LIMIT 1
	`, mercenaryID, defaultClaim).Scan(&lastKillLogID, &lastClaimAt)
	return
}

// GetRecentKillLogs returns up to `limit` most recent distinct-monster kill logs for
// characters on loan to the given mercenary (rasta), with id greater than afterID.
func (r *MercenaryRepository) GetRecentKillLogs(rastaID uint32, afterID uint32, limit int) ([]MercenaryKillLog, error) {
	rows, err := r.db.Queryx(`
		SELECT MAX(id) AS id, monster
		FROM kill_logs
		WHERE character_id IN (
			SELECT id FROM characters WHERE pact_id = $1
		)
		AND id > $2
		GROUP BY monster
		ORDER BY MAX(timestamp) DESC
		LIMIT $3
	`, rastaID, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("query recent kill logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logs []MercenaryKillLog
	for rows.Next() {
		var log MercenaryKillLog
		if err := rows.StructScan(&log); err != nil {
			return nil, fmt.Errorf("scan kill log row: %w", err)
		}
		logs = append(logs, log)
	}
	return logs, rows.Err()
}

// InsertMercenaryReward records a new reward claim checkpoint for a mercenary.
func (r *MercenaryRepository) InsertMercenaryReward(mercenaryID, lastKillLogID uint32, claimAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO mercenary_rewards (mercenary_id, last_kill_log_id, last_claim_at)
		VALUES ($1, $2, $3)
	`, mercenaryID, lastKillLogID, claimAt)
	return err
}

// MercenaryLogEntry represents one contract-history entry for a character.
type MercenaryLogEntry struct {
	Name        string
	Date        time.Time
	IsInitiator bool
	State       uint16
}

// GetMercenaryLogs returns the mercenary contract history for a character, most recent first.
func (r *MercenaryRepository) GetMercenaryLogs(charID uint32) ([]MercenaryLogEntry, error) {
	rows, err := r.db.Query(`
		SELECT COALESCE(c.name, ''), ml.log_time, ml.is_initiator, ml.state
		FROM mercenary_logs ml
		LEFT JOIN characters c ON c.id = ml.mercenary_id
		WHERE ml.player_id = $1
		ORDER BY ml.log_time DESC
	`, charID)
	if err != nil {
		return nil, fmt.Errorf("query mercenary logs: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var logs []MercenaryLogEntry
	for rows.Next() {
		var l MercenaryLogEntry
		if err := rows.Scan(&l.Name, &l.Date, &l.IsInitiator, &l.State); err != nil {
			return nil, fmt.Errorf("scan mercenary log row: %w", err)
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
