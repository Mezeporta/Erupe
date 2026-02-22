package channelserver

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const allianceInfoSelectSQL = `
SELECT
ga.id,
ga.name,
created_at,
parent_id,
CASE
	WHEN sub1_id IS NULL THEN 0
	ELSE sub1_id
END,
CASE
	WHEN sub2_id IS NULL THEN 0
	ELSE sub2_id
END
FROM guild_alliances ga
`

// GetAllianceByID loads alliance data including parent and sub guilds.
func (r *GuildRepository) GetAllianceByID(allianceID uint32) (*GuildAlliance, error) {
	rows, err := r.db.Queryx(fmt.Sprintf(`%s WHERE ga.id = $1`, allianceInfoSelectSQL), allianceID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	if !rows.Next() {
		return nil, nil
	}
	return r.scanAllianceWithGuilds(rows)
}

// ListAlliances returns all alliances with their guild data populated.
func (r *GuildRepository) ListAlliances() ([]*GuildAlliance, error) {
	rows, err := r.db.Queryx(allianceInfoSelectSQL)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var alliances []*GuildAlliance
	for rows.Next() {
		alliance, err := r.scanAllianceWithGuilds(rows)
		if err != nil {
			continue
		}
		alliances = append(alliances, alliance)
	}
	return alliances, nil
}

// CreateAlliance creates a new guild alliance with the given parent guild.
func (r *GuildRepository) CreateAlliance(name string, parentGuildID uint32) error {
	_, err := r.db.Exec("INSERT INTO guild_alliances (name, parent_id) VALUES ($1, $2)", name, parentGuildID)
	return err
}

// DeleteAlliance removes an alliance by ID.
func (r *GuildRepository) DeleteAlliance(allianceID uint32) error {
	_, err := r.db.Exec("DELETE FROM guild_alliances WHERE id=$1", allianceID)
	return err
}

// RemoveGuildFromAlliance removes a guild from its alliance, shifting sub2 into sub1's slot if needed.
func (r *GuildRepository) RemoveGuildFromAlliance(allianceID, guildID, subGuild1ID, subGuild2ID uint32) error {
	if guildID == subGuild1ID && subGuild2ID > 0 {
		_, err := r.db.Exec(`UPDATE guild_alliances SET sub1_id = sub2_id, sub2_id = NULL WHERE id = $1`, allianceID)
		return err
	} else if guildID == subGuild1ID {
		_, err := r.db.Exec(`UPDATE guild_alliances SET sub1_id = NULL WHERE id = $1`, allianceID)
		return err
	}
	_, err := r.db.Exec(`UPDATE guild_alliances SET sub2_id = NULL WHERE id = $1`, allianceID)
	return err
}

// scanAllianceWithGuilds scans an alliance row and populates its guild data.
func (r *GuildRepository) scanAllianceWithGuilds(rows *sqlx.Rows) (*GuildAlliance, error) {
	alliance := &GuildAlliance{}
	if err := rows.StructScan(alliance); err != nil {
		return nil, err
	}

	parentGuild, err := r.GetByID(alliance.ParentGuildID)
	if err != nil {
		return nil, err
	}
	alliance.ParentGuild = *parentGuild
	alliance.TotalMembers += parentGuild.MemberCount

	if alliance.SubGuild1ID > 0 {
		subGuild1, err := r.GetByID(alliance.SubGuild1ID)
		if err != nil {
			return nil, err
		}
		alliance.SubGuild1 = *subGuild1
		alliance.TotalMembers += subGuild1.MemberCount
	}

	if alliance.SubGuild2ID > 0 {
		subGuild2, err := r.GetByID(alliance.SubGuild2ID)
		if err != nil {
			return nil, err
		}
		alliance.SubGuild2 = *subGuild2
		alliance.TotalMembers += subGuild2.MemberCount
	}

	return alliance, nil
}
