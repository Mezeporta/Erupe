package channelserver

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// TournamentRepository centralizes all database access for tournament tables.
type TournamentRepository struct {
	db *sqlx.DB
}

// NewTournamentRepository creates a new TournamentRepository.
func NewTournamentRepository(db *sqlx.DB) *TournamentRepository {
	return &TournamentRepository{db: db}
}

// GetActive returns the most recently started tournament that is still within its
// reward window (reward_end >= now), or nil if no active tournament exists.
func (r *TournamentRepository) GetActive(now int64) (*Tournament, error) {
	var t Tournament
	err := r.db.QueryRowx(
		`SELECT id, name, start_time, entry_end, ranking_end, reward_end
		 FROM tournaments
		 WHERE start_time <= $1 AND reward_end >= $1
		 ORDER BY start_time DESC
		 LIMIT 1`,
		now,
	).StructScan(&t)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active tournament: %w", err)
	}
	return &t, nil
}

// GetCups returns all cups belonging to the given tournament, ordered by ID.
func (r *TournamentRepository) GetCups(tournamentID uint32) ([]TournamentCup, error) {
	var cups []TournamentCup
	err := r.db.Select(&cups,
		`SELECT id, cup_group, cup_type, unk, name, description
		 FROM tournament_cups
		 WHERE tournament_id = $1
		 ORDER BY id`,
		tournamentID,
	)
	return cups, err
}

// GetSubEvents returns all sub-events ordered by cup group and event sub type.
func (r *TournamentRepository) GetSubEvents() ([]TournamentSubEvent, error) {
	var events []TournamentSubEvent
	err := r.db.Select(&events,
		`SELECT id, cup_group, event_sub_type, quest_file_id, name
		 FROM tournament_sub_events
		 ORDER BY cup_group, event_sub_type`,
	)
	return events, err
}

// Register registers a character for a tournament. If the character is already
// registered the existing entry ID is returned (ON CONFLICT DO NOTHING, then re-SELECT).
func (r *TournamentRepository) Register(charID, tournamentID uint32) (uint32, error) {
	_, err := r.db.Exec(
		`INSERT INTO tournament_entries (char_id, tournament_id)
		 VALUES ($1, $2)
		 ON CONFLICT (char_id, tournament_id) DO NOTHING`,
		charID, tournamentID,
	)
	if err != nil {
		return 0, fmt.Errorf("insert tournament entry: %w", err)
	}
	var id uint32
	err = r.db.QueryRow(
		`SELECT id FROM tournament_entries WHERE char_id = $1 AND tournament_id = $2`,
		charID, tournamentID,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("fetch tournament entry id: %w", err)
	}
	return id, nil
}

// GetEntry returns the registration record for a character/tournament pair, or nil if not found.
func (r *TournamentRepository) GetEntry(charID, tournamentID uint32) (*TournamentEntry, error) {
	var e TournamentEntry
	err := r.db.QueryRowx(
		`SELECT id, char_id, tournament_id
		 FROM tournament_entries
		 WHERE char_id = $1 AND tournament_id = $2`,
		charID, tournamentID,
	).StructScan(&e)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get tournament entry: %w", err)
	}
	return &e, nil
}

// SubmitResult records a completed tournament run for a character.
func (r *TournamentRepository) SubmitResult(charID, tournamentID, eventID, questSlot, stageHandle uint32) error {
	_, err := r.db.Exec(
		`INSERT INTO tournament_results (char_id, tournament_id, event_id, quest_slot, stage_handle)
		 VALUES ($1, $2, $3, $4, $5)`,
		charID, tournamentID, eventID, questSlot, stageHandle,
	)
	if err != nil {
		return fmt.Errorf("insert tournament result: %w", err)
	}
	return nil
}

// GetLeaderboard returns the ranked leaderboard for an event ID.
// Rank is assigned by submission order (first submitted = rank 1).
// Returns at most 100 entries.
func (r *TournamentRepository) GetLeaderboard(eventID uint32) ([]TournamentRankEntry, error) {
	type row struct {
		CharID    uint32 `db:"char_id"`
		Rank      int64  `db:"rank"`
		Grade     int    `db:"grade"`
		HR        int    `db:"hr"`
		GR        int    `db:"gr"`
		CharName  string `db:"char_name"`
		GuildName string `db:"guild_name"`
	}
	var rows []row
	err := r.db.Select(&rows, `
		SELECT
		    r.char_id,
		    ROW_NUMBER() OVER (ORDER BY r.submitted_at ASC)::int AS rank,
		    c.gr::int AS grade,
		    c.hr::int AS hr,
		    c.gr::int AS gr,
		    c.name AS char_name,
		    COALESCE(g.name, '') AS guild_name
		FROM tournament_results r
		JOIN characters c ON c.id = r.char_id
		LEFT JOIN guild_characters gc ON gc.character_id = r.char_id
		LEFT JOIN guilds g ON g.id = gc.guild_id
		WHERE r.event_id = $1
		ORDER BY r.submitted_at ASC
		LIMIT 100`,
		eventID,
	)
	if err != nil {
		return nil, fmt.Errorf("get tournament leaderboard: %w", err)
	}
	entries := make([]TournamentRankEntry, len(rows))
	for i, row := range rows {
		entries[i] = TournamentRankEntry{
			CharID:    row.CharID,
			Rank:      uint32(row.Rank),
			Grade:     uint16(row.Grade),
			HR:        uint16(row.HR),
			GR:        uint16(row.GR),
			CharName:  row.CharName,
			GuildName: row.GuildName,
		}
	}
	return entries, nil
}
