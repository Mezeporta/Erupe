package channelserver

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"erupe-ce/common/stringsupport"

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
	COALESCE(rp_reset_at, '2000-01-01'::timestamptz) AS rp_reset_at,
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

// --- Guild Posts ---

// ListPosts returns active guild posts of the given type, ordered by newest first.
func (r *GuildRepository) ListPosts(guildID uint32, postType int) ([]*MessageBoardPost, error) {
	rows, err := r.db.Queryx(
		`SELECT id, stamp_id, title, body, author_id, created_at, liked_by
		 FROM guild_posts WHERE guild_id = $1 AND post_type = $2 AND deleted = false
		 ORDER BY created_at DESC`, guildID, postType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var posts []*MessageBoardPost
	for rows.Next() {
		post := &MessageBoardPost{}
		if err := rows.StructScan(post); err != nil {
			continue
		}
		posts = append(posts, post)
	}
	return posts, nil
}

// CreatePost inserts a new guild post and soft-deletes excess posts beyond maxPosts.
func (r *GuildRepository) CreatePost(guildID, authorID, stampID uint32, postType int, title, body string, maxPosts int) error {
	tx, err := r.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec(
		`INSERT INTO guild_posts (guild_id, author_id, stamp_id, post_type, title, body) VALUES ($1, $2, $3, $4, $5, $6)`,
		guildID, authorID, stampID, postType, title, body); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE guild_posts SET deleted = true WHERE id IN (
		SELECT id FROM guild_posts WHERE guild_id = $1 AND post_type = $2 AND deleted = false
		ORDER BY created_at DESC OFFSET $3
	)`, guildID, postType, maxPosts); err != nil {
		return err
	}
	return tx.Commit()
}

// DeletePost soft-deletes a guild post by ID.
func (r *GuildRepository) DeletePost(postID uint32) error {
	_, err := r.db.Exec("UPDATE guild_posts SET deleted = true WHERE id = $1", postID)
	return err
}

// UpdatePost updates the title and body of a guild post.
func (r *GuildRepository) UpdatePost(postID uint32, title, body string) error {
	_, err := r.db.Exec("UPDATE guild_posts SET title = $1, body = $2 WHERE id = $3", title, body, postID)
	return err
}

// UpdatePostStamp updates the stamp of a guild post.
func (r *GuildRepository) UpdatePostStamp(postID, stampID uint32) error {
	_, err := r.db.Exec("UPDATE guild_posts SET stamp_id = $1 WHERE id = $2", stampID, postID)
	return err
}

// GetPostLikedBy returns the liked_by CSV string for a guild post.
func (r *GuildRepository) GetPostLikedBy(postID uint32) (string, error) {
	var likedBy string
	err := r.db.QueryRow("SELECT liked_by FROM guild_posts WHERE id = $1", postID).Scan(&likedBy)
	return likedBy, err
}

// SetPostLikedBy updates the liked_by CSV string for a guild post.
func (r *GuildRepository) SetPostLikedBy(postID uint32, likedBy string) error {
	_, err := r.db.Exec("UPDATE guild_posts SET liked_by = $1 WHERE id = $2", likedBy, postID)
	return err
}

// CountNewPosts returns the count of non-deleted posts created after the given time.
func (r *GuildRepository) CountNewPosts(guildID uint32, since time.Time) (int, error) {
	var count int
	err := r.db.QueryRow(
		`SELECT COUNT(*) FROM guild_posts WHERE guild_id = $1 AND deleted = false AND (EXTRACT(epoch FROM created_at)::int) > $2`,
		guildID, since.Unix()).Scan(&count)
	return count, err
}

// --- Guild Alliances ---

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
	defer rows.Close()
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
	defer rows.Close()
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

// --- Guild Adventures ---

// ListAdventures returns all adventures for a guild.
func (r *GuildRepository) ListAdventures(guildID uint32) ([]*GuildAdventure, error) {
	rows, err := r.db.Queryx(
		"SELECT id, destination, charge, depart, return, collected_by FROM guild_adventures WHERE guild_id = $1", guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var adventures []*GuildAdventure
	for rows.Next() {
		adv := &GuildAdventure{}
		if err := rows.StructScan(adv); err != nil {
			continue
		}
		adventures = append(adventures, adv)
	}
	return adventures, nil
}

// CreateAdventure inserts a new guild adventure.
func (r *GuildRepository) CreateAdventure(guildID, destination uint32, depart, returnTime int64) error {
	_, err := r.db.Exec(
		"INSERT INTO guild_adventures (guild_id, destination, depart, return) VALUES ($1, $2, $3, $4)",
		guildID, destination, depart, returnTime)
	return err
}

// CreateAdventureWithCharge inserts a new guild adventure with an initial charge (Diva variant).
func (r *GuildRepository) CreateAdventureWithCharge(guildID, destination, charge uint32, depart, returnTime int64) error {
	_, err := r.db.Exec(
		"INSERT INTO guild_adventures (guild_id, destination, charge, depart, return) VALUES ($1, $2, $3, $4, $5)",
		guildID, destination, charge, depart, returnTime)
	return err
}

// CollectAdventure marks an adventure as collected by the given character (CSV append).
// Uses SELECT FOR UPDATE to prevent concurrent double-collect.
func (r *GuildRepository) CollectAdventure(adventureID uint32, charID uint32) error {
	tx, err := r.db.BeginTxx(context.Background(), nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var collectedBy string
	err = tx.QueryRow("SELECT collected_by FROM guild_adventures WHERE id = $1 FOR UPDATE", adventureID).Scan(&collectedBy)
	if err != nil {
		return err
	}
	collectedBy = stringsupport.CSVAdd(collectedBy, int(charID))
	if _, err = tx.Exec("UPDATE guild_adventures SET collected_by = $1 WHERE id = $2", collectedBy, adventureID); err != nil {
		return err
	}
	return tx.Commit()
}

// ChargeAdventure adds charge to a guild adventure.
func (r *GuildRepository) ChargeAdventure(adventureID uint32, amount uint32) error {
	_, err := r.db.Exec("UPDATE guild_adventures SET charge = charge + $1 WHERE id = $2", amount, adventureID)
	return err
}

// --- Guild Treasure Hunts ---

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
	defer rows.Close()
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

// --- Guild Cooking/Meals ---

// ListMeals returns all meals for a guild.
func (r *GuildRepository) ListMeals(guildID uint32) ([]*GuildMeal, error) {
	rows, err := r.db.Queryx("SELECT id, meal_id, level, created_at FROM guild_meals WHERE guild_id = $1", guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var meals []*GuildMeal
	for rows.Next() {
		meal := &GuildMeal{}
		if err := rows.StructScan(meal); err != nil {
			continue
		}
		meals = append(meals, meal)
	}
	return meals, nil
}

// CreateMeal inserts a new guild meal and returns the new ID.
func (r *GuildRepository) CreateMeal(guildID, mealID, level uint32, createdAt time.Time) (uint32, error) {
	var id uint32
	err := r.db.QueryRow(
		"INSERT INTO guild_meals (guild_id, meal_id, level, created_at) VALUES ($1, $2, $3, $4) RETURNING id",
		guildID, mealID, level, createdAt).Scan(&id)
	return id, err
}

// UpdateMeal updates an existing guild meal's fields.
func (r *GuildRepository) UpdateMeal(mealID, newMealID, level uint32, createdAt time.Time) error {
	_, err := r.db.Exec("UPDATE guild_meals SET meal_id = $1, level = $2, created_at = $3 WHERE id = $4",
		newMealID, level, createdAt, mealID)
	return err
}

// ClaimHuntBox updates the box_claimed timestamp for a guild character.
func (r *GuildRepository) ClaimHuntBox(charID uint32, claimedAt time.Time) error {
	_, err := r.db.Exec(`UPDATE guild_characters SET box_claimed=$1 WHERE character_id=$2`, claimedAt, charID)
	return err
}

// GuildKill represents a kill log entry for guild hunt data.
type GuildKill struct {
	ID      uint32 `db:"id"`
	Monster uint32 `db:"monster"`
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
	defer rows.Close()
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

// --- Guild Scouts ---

// ScoutedCharacter represents an invited character in the scout list.
type ScoutedCharacter struct {
	CharID  uint32 `db:"id"`
	Name    string `db:"name"`
	HR      uint16 `db:"hr"`
	GR      uint16 `db:"gr"`
	ActorID uint32 `db:"actor_id"`
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

// ListInvitedCharacters returns all characters with pending guild invitations.
func (r *GuildRepository) ListInvitedCharacters(guildID uint32) ([]*ScoutedCharacter, error) {
	rows, err := r.db.Queryx(`
		SELECT c.id, c.name, c.hr, c.gr, ga.actor_id
			FROM guild_applications ga
			JOIN characters c ON c.id = ga.character_id
		WHERE ga.guild_id = $1 AND ga.application_type = 'invited'
	`, guildID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var chars []*ScoutedCharacter
	for rows.Next() {
		sc := &ScoutedCharacter{}
		if err := rows.StructScan(sc); err != nil {
			continue
		}
		chars = append(chars, sc)
	}
	return chars, nil
}

// RolloverDailyRP moves rp_today into rp_yesterday for all members of a guild,
// then updates the guild's rp_reset_at timestamp.
func (r *GuildRepository) RolloverDailyRP(guildID uint32, noon time.Time) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(
		`UPDATE guild_characters SET rp_yesterday = rp_today, rp_today = 0 WHERE guild_id = $1`,
		guildID,
	); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.Exec(
		`UPDATE guilds SET rp_reset_at = $1 WHERE id = $2`,
		noon, guildID,
	); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// AddWeeklyBonusUsers atomically adds numUsers to the guild's weekly bonus exceptional user count.
func (r *GuildRepository) AddWeeklyBonusUsers(guildID uint32, numUsers uint8) error {
	_, err := r.db.Exec(
		"UPDATE guilds SET weekly_bonus_users = weekly_bonus_users + $1 WHERE id = $2",
		numUsers, guildID,
	)
	return err
}
