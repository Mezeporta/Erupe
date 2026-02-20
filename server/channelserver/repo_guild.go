package channelserver

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
)

// GuildRepository centralizes all database access for guild-related tables
// (guilds, guild_characters, guild_applications).
type GuildRepository struct {
	db *sqlx.DB
}

// NewGuildRepository creates a new GuildRepository.
func NewGuildRepository(db *sqlx.DB) *GuildRepository {
	return &GuildRepository{db: db}
}

const guildInfoSelectSQL = `
SELECT
	g.id,
	g.name,
	rank_rp,
	event_rp,
	room_rp,
	COALESCE(room_expiry, '1970-01-01') AS room_expiry,
	main_motto,
	sub_motto,
	created_at,
	leader_id,
	c.name AS leader_name,
	comment,
	COALESCE(pugi_name_1, '') AS pugi_name_1,
	COALESCE(pugi_name_2, '') AS pugi_name_2,
	COALESCE(pugi_name_3, '') AS pugi_name_3,
	pugi_outfit_1,
	pugi_outfit_2,
	pugi_outfit_3,
	pugi_outfits,
	recruiting,
	COALESCE((SELECT team FROM festa_registrations fr WHERE fr.guild_id = g.id), 'none') AS festival_color,
	COALESCE((SELECT SUM(fs.souls) FROM festa_submissions fs WHERE fs.guild_id=g.id), 0) AS souls,
	COALESCE((
		SELECT id FROM guild_alliances ga WHERE
	 	ga.parent_id = g.id OR
	 	ga.sub1_id = g.id OR
	 	ga.sub2_id = g.id
	), 0) AS alliance_id,
	icon,
	(SELECT count(1) FROM guild_characters gc WHERE gc.guild_id = g.id) AS member_count
	FROM guilds g
	JOIN guild_characters gc ON gc.character_id = leader_id
	JOIN characters c on leader_id = c.id
`

const guildMembersSelectSQL = `
SELECT
	COALESCE(g.id, 0) AS guild_id,
	joined_at,
	COALESCE((SELECT SUM(souls) FROM festa_submissions fs WHERE fs.character_id=c.id), 0) AS souls,
	COALESCE(rp_today, 0) AS rp_today,
	COALESCE(rp_yesterday, 0) AS rp_yesterday,
	c.name,
	c.id AS character_id,
	COALESCE(order_index, 0) AS order_index,
	c.last_login,
	COALESCE(recruiter, false) AS recruiter,
	COALESCE(avoid_leadership, false) AS avoid_leadership,
	c.hr,
	c.gr,
	c.weapon_id,
	c.weapon_type,
	CASE WHEN g.leader_id = c.id THEN true ELSE false END AS is_leader,
	character.is_applicant
	FROM (
		SELECT character_id, true as is_applicant, guild_id
		FROM guild_applications ga
		WHERE ga.application_type = 'applied'
		UNION
		SELECT character_id, false as is_applicant, guild_id
		FROM guild_characters gc
	) character
	JOIN characters c on character.character_id = c.id
	LEFT JOIN guild_characters gc ON gc.character_id = character.character_id
	LEFT JOIN guilds g ON g.id = gc.guild_id
`

func scanGuild(rows *sqlx.Rows) (*Guild, error) {
	guild := &Guild{}
	if err := rows.StructScan(guild); err != nil {
		return nil, err
	}
	return guild, nil
}

func scanGuildMember(rows *sqlx.Rows) (*GuildMember, error) {
	member := &GuildMember{}
	if err := rows.StructScan(member); err != nil {
		return nil, err
	}
	return member, nil
}

// GetByID retrieves guild info by guild ID, returning nil if not found.
func (r *GuildRepository) GetByID(guildID uint32) (*Guild, error) {
	rows, err := r.db.Queryx(fmt.Sprintf(`%s WHERE g.id = $1 LIMIT 1`, guildInfoSelectSQL), guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	return scanGuild(rows)
}

// GetByCharID retrieves guild info for a character, including applied guilds.
func (r *GuildRepository) GetByCharID(charID uint32) (*Guild, error) {
	rows, err := r.db.Queryx(fmt.Sprintf(`
		%s
		WHERE EXISTS(
				SELECT 1
				FROM guild_characters gc1
				WHERE gc1.character_id = $1
				  AND gc1.guild_id = g.id
			)
		   OR EXISTS(
				SELECT 1
				FROM guild_applications ga
				WHERE ga.character_id = $1
				  AND ga.guild_id = g.id
				  AND ga.application_type = 'applied'
			)
		LIMIT 1
	`, guildInfoSelectSQL), charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if !rows.Next() {
		return nil, nil
	}
	return scanGuild(rows)
}

// ListAll returns all guilds. Used for guild enumeration/search.
func (r *GuildRepository) ListAll() ([]*Guild, error) {
	rows, err := r.db.Queryx(guildInfoSelectSQL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var guilds []*Guild
	for rows.Next() {
		guild, err := scanGuild(rows)
		if err != nil {
			continue
		}
		guilds = append(guilds, guild)
	}
	return guilds, nil
}

// Create creates a new guild and adds the leader as its first member.
func (r *GuildRepository) Create(leaderCharID uint32, guildName string) (int32, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, err
	}

	var guildID int32
	err = tx.QueryRow(
		"INSERT INTO guilds (name, leader_id) VALUES ($1, $2) RETURNING id",
		guildName, leaderCharID,
	).Scan(&guildID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	_, err = tx.Exec(`INSERT INTO guild_characters (guild_id, character_id) VALUES ($1, $2)`, guildID, leaderCharID)
	if err != nil {
		_ = tx.Rollback()
		return 0, err
	}

	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return guildID, nil
}

// Save persists guild metadata changes.
func (r *GuildRepository) Save(guild *Guild) error {
	_, err := r.db.Exec(`
		UPDATE guilds SET main_motto=$2, sub_motto=$3, comment=$4, pugi_name_1=$5, pugi_name_2=$6, pugi_name_3=$7,
		pugi_outfit_1=$8, pugi_outfit_2=$9, pugi_outfit_3=$10, pugi_outfits=$11, icon=$12, leader_id=$13 WHERE id=$1
	`, guild.ID, guild.MainMotto, guild.SubMotto, guild.Comment, guild.PugiName1, guild.PugiName2, guild.PugiName3,
		guild.PugiOutfit1, guild.PugiOutfit2, guild.PugiOutfit3, guild.PugiOutfits, guild.Icon, guild.LeaderCharID)
	return err
}

// Disband removes a guild, its members, and cleans up alliance references.
func (r *GuildRepository) Disband(guildID uint32) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	stmts := []string{
		"DELETE FROM guild_characters WHERE guild_id = $1",
		"DELETE FROM guilds WHERE id = $1",
		"DELETE FROM guild_alliances WHERE parent_id=$1",
	}
	for _, stmt := range stmts {
		if _, err := tx.Exec(stmt, guildID); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	if _, err := tx.Exec("UPDATE guild_alliances SET sub1_id=sub2_id, sub2_id=NULL WHERE sub1_id=$1", guildID); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.Exec("UPDATE guild_alliances SET sub2_id=NULL WHERE sub2_id=$1", guildID); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// RemoveCharacter removes a character from their guild.
func (r *GuildRepository) RemoveCharacter(charID uint32) error {
	_, err := r.db.Exec("DELETE FROM guild_characters WHERE character_id=$1", charID)
	return err
}

// AcceptApplication deletes the application and adds the character to the guild.
func (r *GuildRepository) AcceptApplication(guildID, charID uint32) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	if _, err := tx.Exec(`DELETE FROM guild_applications WHERE character_id = $1`, charID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.Exec(`
		INSERT INTO guild_characters (guild_id, character_id, order_index)
		VALUES ($1, $2, (SELECT MAX(order_index) + 1 FROM guild_characters WHERE guild_id = $1))
	`, guildID, charID); err != nil {
		_ = tx.Rollback()
		return err
	}

	return tx.Commit()
}

// CreateApplication inserts a guild application or invitation.
// If tx is non-nil, the operation participates in the given transaction.
func (r *GuildRepository) CreateApplication(guildID, charID, actorID uint32, appType GuildApplicationType, tx *sql.Tx) error {
	query := `INSERT INTO guild_applications (guild_id, character_id, actor_id, application_type) VALUES ($1, $2, $3, $4)`
	if tx != nil {
		_, err := tx.Exec(query, guildID, charID, actorID, appType)
		return err
	}
	_, err := r.db.Exec(query, guildID, charID, actorID, appType)
	return err
}

// CancelInvitation removes an invitation for a character.
func (r *GuildRepository) CancelInvitation(guildID, charID uint32) error {
	_, err := r.db.Exec(
		`DELETE FROM guild_applications WHERE character_id = $1 AND guild_id = $2 AND application_type = 'invited'`,
		charID, guildID,
	)
	return err
}

// RejectApplication removes an applied application for a character.
func (r *GuildRepository) RejectApplication(guildID, charID uint32) error {
	_, err := r.db.Exec(
		`DELETE FROM guild_applications WHERE character_id = $1 AND guild_id = $2 AND application_type = 'applied'`,
		charID, guildID,
	)
	return err
}

// ArrangeCharacters reorders guild members by updating their order_index values.
func (r *GuildRepository) ArrangeCharacters(charIDs []uint32) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	for i, id := range charIDs {
		if _, err := tx.Exec("UPDATE guild_characters SET order_index = $1 WHERE character_id = $2", 2+i, id); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

// GetApplication retrieves a specific application by character, guild, and type.
// Returns nil, nil if not found.
func (r *GuildRepository) GetApplication(guildID, charID uint32, appType GuildApplicationType) (*GuildApplication, error) {
	app := &GuildApplication{}
	err := r.db.QueryRowx(`
		SELECT * from guild_applications WHERE character_id = $1 AND guild_id = $2 AND application_type = $3
	`, charID, guildID, appType).StructScan(app)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return app, nil
}

// HasApplication checks whether any application exists for the character in the guild.
func (r *GuildRepository) HasApplication(guildID, charID uint32) (bool, error) {
	var n int
	err := r.db.QueryRow(`SELECT 1 from guild_applications WHERE character_id = $1 AND guild_id = $2`, charID, guildID).Scan(&n)
	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// GetItemBox returns the raw item_box bytes for a guild.
func (r *GuildRepository) GetItemBox(guildID uint32) ([]byte, error) {
	var data []byte
	err := r.db.QueryRow(`SELECT item_box FROM guilds WHERE id=$1`, guildID).Scan(&data)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	return data, err
}

// SaveItemBox writes the serialized item box data for a guild.
func (r *GuildRepository) SaveItemBox(guildID uint32, data []byte) error {
	_, err := r.db.Exec(`UPDATE guilds SET item_box=$1 WHERE id=$2`, data, guildID)
	return err
}

// GetMembers loads all members (or applicants) of a guild.
func (r *GuildRepository) GetMembers(guildID uint32, applicants bool) ([]*GuildMember, error) {
	rows, err := r.db.Queryx(fmt.Sprintf(`
		%s
		WHERE character.guild_id = $1 AND is_applicant = $2
	`, guildMembersSelectSQL), guildID, applicants)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	members := make([]*GuildMember, 0)
	for rows.Next() {
		member, err := scanGuildMember(rows)
		if err != nil {
			return nil, err
		}
		members = append(members, member)
	}
	return members, nil
}

// GetCharacterMembership loads a character's guild membership data.
// Returns nil, nil if the character is not in any guild.
func (r *GuildRepository) GetCharacterMembership(charID uint32) (*GuildMember, error) {
	rows, err := r.db.Queryx(fmt.Sprintf("%s	WHERE character.character_id=$1", guildMembersSelectSQL), charID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, nil
	}
	return scanGuildMember(rows)
}

// SaveMember persists guild member changes (avoid_leadership and order_index).
func (r *GuildRepository) SaveMember(member *GuildMember) error {
	_, err := r.db.Exec(
		"UPDATE guild_characters SET avoid_leadership=$1, order_index=$2 WHERE character_id=$3",
		member.AvoidLeadership, member.OrderIndex, member.CharID,
	)
	return err
}

// SetRecruiting updates whether a guild is accepting applications.
func (r *GuildRepository) SetRecruiting(guildID uint32, recruiting bool) error {
	_, err := r.db.Exec("UPDATE guilds SET recruiting=$1 WHERE id=$2", recruiting, guildID)
	return err
}

// SetPugiOutfits updates the unlocked pugi outfit bitmask.
func (r *GuildRepository) SetPugiOutfits(guildID uint32, outfits uint32) error {
	_, err := r.db.Exec(`UPDATE guilds SET pugi_outfits=$1 WHERE id=$2`, outfits, guildID)
	return err
}

// SetRecruiter updates whether a character has recruiter rights.
func (r *GuildRepository) SetRecruiter(charID uint32, allowed bool) error {
	_, err := r.db.Exec("UPDATE guild_characters SET recruiter=$1 WHERE character_id=$2", allowed, charID)
	return err
}

// AddMemberDailyRP adds RP to a member's daily total.
func (r *GuildRepository) AddMemberDailyRP(charID uint32, amount uint16) error {
	_, err := r.db.Exec(`UPDATE guild_characters SET rp_today=rp_today+$1 WHERE character_id=$2`, amount, charID)
	return err
}

// ExchangeEventRP subtracts RP from a guild's event pool and returns the new balance.
func (r *GuildRepository) ExchangeEventRP(guildID uint32, amount uint16) (uint32, error) {
	var balance uint32
	err := r.db.QueryRow(`UPDATE guilds SET event_rp=event_rp-$1 WHERE id=$2 RETURNING event_rp`, amount, guildID).Scan(&balance)
	return balance, err
}

// AddRankRP adds RP to a guild's rank total.
func (r *GuildRepository) AddRankRP(guildID uint32, amount uint16) error {
	_, err := r.db.Exec(`UPDATE guilds SET rank_rp = rank_rp + $1 WHERE id = $2`, amount, guildID)
	return err
}

// AddEventRP adds RP to a guild's event total.
func (r *GuildRepository) AddEventRP(guildID uint32, amount uint16) error {
	_, err := r.db.Exec(`UPDATE guilds SET event_rp = event_rp + $1 WHERE id = $2`, amount, guildID)
	return err
}

// GetRoomRP returns the current room RP for a guild.
func (r *GuildRepository) GetRoomRP(guildID uint32) (uint16, error) {
	var rp uint16
	err := r.db.QueryRow(`SELECT room_rp FROM guilds WHERE id = $1`, guildID).Scan(&rp)
	return rp, err
}

// SetRoomRP sets the room RP for a guild.
func (r *GuildRepository) SetRoomRP(guildID uint32, rp uint16) error {
	_, err := r.db.Exec(`UPDATE guilds SET room_rp = $1 WHERE id = $2`, rp, guildID)
	return err
}

// AddRoomRP atomically adds RP to a guild's room total.
func (r *GuildRepository) AddRoomRP(guildID uint32, amount uint16) error {
	_, err := r.db.Exec(`UPDATE guilds SET room_rp = room_rp + $1 WHERE id = $2`, amount, guildID)
	return err
}

// SetRoomExpiry sets the room expiry time for a guild.
func (r *GuildRepository) SetRoomExpiry(guildID uint32, expiry time.Time) error {
	_, err := r.db.Exec(`UPDATE guilds SET room_expiry = $1 WHERE id = $2`, expiry, guildID)
	return err
}
