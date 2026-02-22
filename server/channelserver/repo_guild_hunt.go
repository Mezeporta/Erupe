package channelserver

import (
	"database/sql"
	"errors"
	"time"
)

// GuildKill represents a kill log entry for guild hunt data.
type GuildKill struct {
	ID      uint32 `db:"id"`
	Monster uint32 `db:"monster"`
}

// GetPendingHunt returns the pending (unacquired) hunt for a character, or nil if none.
func (r *GuildRepository) GetPendingHunt(charID uint32) (*TreasureHunt, error) {
	hunt := &TreasureHunt{}
	err := r.db.QueryRowx(
		`SELECT id, host_id, destination, level, start, hunt_data FROM guild_hunts WHERE host_id=$1 AND acquired=FALSE`,
		charID).StructScan(hunt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return hunt, nil
}

// ListGuildHunts returns acquired level-2 hunts for a guild, with hunter counts and claim status.
func (r *GuildRepository) ListGuildHunts(guildID, charID uint32) ([]*TreasureHunt, error) {
	rows, err := r.db.Queryx(`SELECT gh.id, gh.host_id, gh.destination, gh.level, gh.start, gh.collected, gh.hunt_data,
		(SELECT COUNT(*) FROM guild_characters gc WHERE gc.treasure_hunt = gh.id AND gc.character_id <> $1) AS hunters,
		CASE
			WHEN ghc.character_id IS NOT NULL THEN true
			ELSE false
		END AS claimed
		FROM guild_hunts gh
		LEFT JOIN guild_hunts_claimed ghc ON gh.id = ghc.hunt_id AND ghc.character_id = $1
		WHERE gh.guild_id=$2 AND gh.level=2 AND gh.acquired=TRUE
	`, charID, guildID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var hunts []*TreasureHunt
	for rows.Next() {
		hunt := &TreasureHunt{}
		if err := rows.StructScan(hunt); err != nil {
			continue
		}
		hunts = append(hunts, hunt)
	}
	return hunts, nil
}

// CreateHunt inserts a new guild treasure hunt.
func (r *GuildRepository) CreateHunt(guildID, hostID, destination, level uint32, huntData []byte, catsUsed string) error {
	_, err := r.db.Exec(
		`INSERT INTO guild_hunts (guild_id, host_id, destination, level, hunt_data, cats_used) VALUES ($1, $2, $3, $4, $5, $6)`,
		guildID, hostID, destination, level, huntData, catsUsed)
	return err
}

// AcquireHunt marks a treasure hunt as acquired.
func (r *GuildRepository) AcquireHunt(huntID uint32) error {
	_, err := r.db.Exec(`UPDATE guild_hunts SET acquired=true WHERE id=$1`, huntID)
	return err
}

// RegisterHuntReport sets a character's active treasure hunt.
func (r *GuildRepository) RegisterHuntReport(huntID, charID uint32) error {
	_, err := r.db.Exec(`UPDATE guild_characters SET treasure_hunt=$1 WHERE character_id=$2`, huntID, charID)
	return err
}

// CollectHunt marks a hunt as collected and clears all characters' treasure_hunt references.
func (r *GuildRepository) CollectHunt(huntID uint32) error {
	if _, err := r.db.Exec(`UPDATE guild_hunts SET collected=true WHERE id=$1`, huntID); err != nil {
		return err
	}
	_, err := r.db.Exec(`UPDATE guild_characters SET treasure_hunt=NULL WHERE treasure_hunt=$1`, huntID)
	return err
}

// ClaimHuntReward records that a character has claimed a treasure hunt reward.
func (r *GuildRepository) ClaimHuntReward(huntID, charID uint32) error {
	_, err := r.db.Exec(`INSERT INTO guild_hunts_claimed VALUES ($1, $2)`, huntID, charID)
	return err
}

// ListGuildKills returns kill log entries for guild members since the character's last box claim.
func (r *GuildRepository) ListGuildKills(guildID, charID uint32) ([]*GuildKill, error) {
	rows, err := r.db.Queryx(`SELECT kl.id, kl.monster FROM kill_logs kl
		INNER JOIN guild_characters gc ON kl.character_id = gc.character_id
		WHERE gc.guild_id=$1
		AND kl.timestamp >= (SELECT box_claimed FROM guild_characters WHERE character_id=$2)
	`, guildID, charID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var kills []*GuildKill
	for rows.Next() {
		kill := &GuildKill{}
		if err := rows.StructScan(kill); err != nil {
			continue
		}
		kills = append(kills, kill)
	}
	return kills, nil
}

// CountGuildKills returns the count of kill log entries for guild members since the character's last box claim.
func (r *GuildRepository) CountGuildKills(guildID, charID uint32) (int, error) {
	var count int
	err := r.db.QueryRow(`SELECT COUNT(*) FROM kill_logs kl
		INNER JOIN guild_characters gc ON kl.character_id = gc.character_id
		WHERE gc.guild_id=$1
		AND kl.timestamp >= (SELECT box_claimed FROM guild_characters WHERE character_id=$2)
	`, guildID, charID).Scan(&count)
	return count, err
}

// ClearTreasureHunt clears the treasure_hunt field for a character on logout.
func (r *GuildRepository) ClearTreasureHunt(charID uint32) error {
	_, err := r.db.Exec(`UPDATE guild_characters SET treasure_hunt=NULL WHERE character_id=$1`, charID)
	return err
}

// InsertKillLog records a monster kill log entry for a character.
func (r *GuildRepository) InsertKillLog(charID uint32, monster int, quantity uint8, timestamp time.Time) error {
	_, err := r.db.Exec(`INSERT INTO kill_logs (character_id, monster, quantity, timestamp) VALUES ($1, $2, $3, $4)`, charID, monster, quantity, timestamp)
	return err
}
