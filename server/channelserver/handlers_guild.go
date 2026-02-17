package channelserver

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"erupe-ce/common/mhfitem"
	_config "erupe-ce/config"
	"fmt"
	"sort"
	"time"

	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network/mhfpacket"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

type FestivalColor string

const (
	FestivalColorNone FestivalColor = "none"
	FestivalColorBlue FestivalColor = "blue"
	FestivalColorRed  FestivalColor = "red"
)

var FestivalColorCodes = map[FestivalColor]int16{
	FestivalColorNone: -1,
	FestivalColorBlue: 0,
	FestivalColorRed:  1,
}

type GuildApplicationType string

const (
	GuildApplicationTypeApplied GuildApplicationType = "applied"
	GuildApplicationTypeInvited GuildApplicationType = "invited"
)

type Guild struct {
	ID            uint32        `db:"id"`
	Name          string        `db:"name"`
	MainMotto     uint8         `db:"main_motto"`
	SubMotto      uint8         `db:"sub_motto"`
	CreatedAt     time.Time     `db:"created_at"`
	MemberCount   uint16        `db:"member_count"`
	RankRP        uint32        `db:"rank_rp"`
	EventRP       uint32        `db:"event_rp"`
	RoomRP        uint16        `db:"room_rp"`
	RoomExpiry    time.Time     `db:"room_expiry"`
	Comment       string        `db:"comment"`
	PugiName1     string        `db:"pugi_name_1"`
	PugiName2     string        `db:"pugi_name_2"`
	PugiName3     string        `db:"pugi_name_3"`
	PugiOutfit1   uint8         `db:"pugi_outfit_1"`
	PugiOutfit2   uint8         `db:"pugi_outfit_2"`
	PugiOutfit3   uint8         `db:"pugi_outfit_3"`
	PugiOutfits   uint32        `db:"pugi_outfits"`
	Recruiting    bool          `db:"recruiting"`
	FestivalColor FestivalColor `db:"festival_color"`
	Souls         uint32        `db:"souls"`
	AllianceID    uint32        `db:"alliance_id"`
	Icon          *GuildIcon    `db:"icon"`

	GuildLeader
}

type GuildLeader struct {
	LeaderCharID uint32 `db:"leader_id"`
	LeaderName   string `db:"leader_name"`
}

type GuildIconPart struct {
	Index    uint16
	ID       uint16
	Page     uint8
	Size     uint8
	Rotation uint8
	Red      uint8
	Green    uint8
	Blue     uint8
	PosX     uint16
	PosY     uint16
}

type GuildApplication struct {
	ID              int                  `db:"id"`
	GuildID         uint32               `db:"guild_id"`
	CharID          uint32               `db:"character_id"`
	ActorID         uint32               `db:"actor_id"`
	ApplicationType GuildApplicationType `db:"application_type"`
	CreatedAt       time.Time            `db:"created_at"`
}

type GuildIcon struct {
	Parts []GuildIconPart
}

func (gi *GuildIcon) Scan(val interface{}) (err error) {
	switch v := val.(type) {
	case []byte:
		err = json.Unmarshal(v, &gi)
	case string:
		err = json.Unmarshal([]byte(v), &gi)
	}

	return
}

func (gi *GuildIcon) Value() (valuer driver.Value, err error) {
	return json.Marshal(gi)
}

func (g *Guild) Rank() uint16 {
	rpMap := []uint32{
		24, 48, 96, 144, 192, 240, 288, 360, 432,
		504, 600, 696, 792, 888, 984, 1080, 1200,
	}
	if _config.ErupeConfig.RealClientMode <= _config.Z2 {
		rpMap = []uint32{
			3500, 6000, 8500, 11000, 13500, 16000, 20000, 24000, 28000,
			33000, 38000, 43000, 48000, 55000, 70000, 90000, 120000,
		}
	}
	for i, u := range rpMap {
		if g.RankRP < u {
			if _config.ErupeConfig.RealClientMode <= _config.S6 && i >= 12 {
				return 12
			} else if _config.ErupeConfig.RealClientMode <= _config.F5 && i >= 13 {
				return 13
			} else if _config.ErupeConfig.RealClientMode <= _config.G32 && i >= 14 {
				return 14
			}
			return uint16(i)
		}
	}
	if _config.ErupeConfig.RealClientMode <= _config.S6 {
		return 12
	} else if _config.ErupeConfig.RealClientMode <= _config.F5 {
		return 13
	} else if _config.ErupeConfig.RealClientMode <= _config.G32 {
		return 14
	}
	return 17
}

const guildInfoSelectQuery = `
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

func (guild *Guild) Save(s *Session) error {
	_, err := s.server.db.Exec(`
		UPDATE guilds SET main_motto=$2, sub_motto=$3, comment=$4, pugi_name_1=$5, pugi_name_2=$6, pugi_name_3=$7,
		pugi_outfit_1=$8, pugi_outfit_2=$9, pugi_outfit_3=$10, pugi_outfits=$11, icon=$12, leader_id=$13 WHERE id=$1
	`, guild.ID, guild.MainMotto, guild.SubMotto, guild.Comment, guild.PugiName1, guild.PugiName2, guild.PugiName3,
		guild.PugiOutfit1, guild.PugiOutfit2, guild.PugiOutfit3, guild.PugiOutfits, guild.Icon, guild.LeaderCharID)

	if err != nil {
		s.logger.Error("failed to update guild data", zap.Error(err), zap.Uint32("guildID", guild.ID))
		return err
	}

	return nil
}

func (guild *Guild) CreateApplication(s *Session, charID uint32, applicationType GuildApplicationType, transaction *sql.Tx) error {

	query := `
		INSERT INTO guild_applications (guild_id, character_id, actor_id, application_type)
		VALUES ($1, $2, $3, $4)
	`

	var err error

	if transaction == nil {
		_, err = s.server.db.Exec(query, guild.ID, charID, s.charID, applicationType)
	} else {
		_, err = transaction.Exec(query, guild.ID, charID, s.charID, applicationType)
	}

	if err != nil {
		s.logger.Error(
			"failed to add guild application",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", charID),
		)
		return err
	}

	return nil
}

func (guild *Guild) Disband(s *Session) error {
	transaction, err := s.server.db.Begin()

	if err != nil {
		s.logger.Error("failed to begin transaction", zap.Error(err))
		return err
	}

	_, err = transaction.Exec("DELETE FROM guild_characters WHERE guild_id = $1", guild.ID)

	if err != nil {
		s.logger.Error("failed to remove guild characters", zap.Error(err), zap.Uint32("guildId", guild.ID))
		rollbackTransaction(s, transaction)
		return err
	}

	_, err = transaction.Exec("DELETE FROM guilds WHERE id = $1", guild.ID)

	if err != nil {
		s.logger.Error("failed to remove guild", zap.Error(err), zap.Uint32("guildID", guild.ID))
		rollbackTransaction(s, transaction)
		return err
	}

	_, err = transaction.Exec("DELETE FROM guild_alliances WHERE parent_id=$1", guild.ID)

	if err != nil {
		s.logger.Error("failed to remove guild alliance", zap.Error(err), zap.Uint32("guildID", guild.ID))
		rollbackTransaction(s, transaction)
		return err
	}

	_, err = transaction.Exec("UPDATE guild_alliances SET sub1_id=sub2_id, sub2_id=NULL WHERE sub1_id=$1", guild.ID)

	if err != nil {
		s.logger.Error("failed to remove guild from alliance", zap.Error(err), zap.Uint32("guildID", guild.ID))
		rollbackTransaction(s, transaction)
		return err
	}

	_, err = transaction.Exec("UPDATE guild_alliances SET sub2_id=NULL WHERE sub2_id=$1", guild.ID)

	if err != nil {
		s.logger.Error("failed to remove guild from alliance", zap.Error(err), zap.Uint32("guildID", guild.ID))
		rollbackTransaction(s, transaction)
		return err
	}

	err = transaction.Commit()

	if err != nil {
		s.logger.Error("failed to commit transaction", zap.Error(err))
		return err
	}

	s.logger.Info("Character disbanded guild", zap.Uint32("charID", s.charID), zap.Uint32("guildID", guild.ID))

	return nil
}

func (guild *Guild) RemoveCharacter(s *Session, charID uint32) error {
	_, err := s.server.db.Exec("DELETE FROM guild_characters WHERE character_id=$1", charID)

	if err != nil {
		s.logger.Error(
			"failed to remove character from guild",
			zap.Error(err),
			zap.Uint32("charID", charID),
			zap.Uint32("guildID", guild.ID),
		)

		return err
	}

	return nil
}

func (guild *Guild) AcceptApplication(s *Session, charID uint32) error {
	transaction, err := s.server.db.Begin()

	if err != nil {
		s.logger.Error("failed to start db transaction", zap.Error(err))
		return err
	}

	_, err = transaction.Exec(`DELETE FROM guild_applications WHERE character_id = $1`, charID)

	if err != nil {
		s.logger.Error("failed to accept character's guild application", zap.Error(err))
		rollbackTransaction(s, transaction)
		return err
	}

	_, err = transaction.Exec(`
		INSERT INTO guild_characters (guild_id, character_id, order_index)
		VALUES ($1, $2, (SELECT MAX(order_index) + 1 FROM guild_characters WHERE guild_id = $1))
	`, guild.ID, charID)

	if err != nil {
		s.logger.Error(
			"failed to add applicant to guild",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", charID),
		)
		rollbackTransaction(s, transaction)
		return err
	}

	err = transaction.Commit()

	if err != nil {
		s.logger.Error("failed to commit db transaction", zap.Error(err))
		rollbackTransaction(s, transaction)
		return err
	}

	return nil
}

// This is relying on the fact that invitation ID is also character ID right now
// if invitation ID changes, this will break.
func (guild *Guild) CancelInvitation(s *Session, charID uint32) error {
	_, err := s.server.db.Exec(
		`DELETE FROM guild_applications WHERE character_id = $1 AND guild_id = $2 AND application_type = 'invited'`,
		charID, guild.ID,
	)

	if err != nil {
		s.logger.Error(
			"failed to cancel guild invitation",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", charID),
		)
		return err
	}

	return nil
}

func (guild *Guild) RejectApplication(s *Session, charID uint32) error {
	_, err := s.server.db.Exec(
		`DELETE FROM guild_applications WHERE character_id = $1 AND guild_id = $2 AND application_type = 'applied'`,
		charID, guild.ID,
	)

	if err != nil {
		s.logger.Error(
			"failed to reject guild application",
			zap.Error(err),
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", charID),
		)
		return err
	}

	return nil
}

func (guild *Guild) ArrangeCharacters(s *Session, charIDs []uint32) error {
	transaction, err := s.server.db.Begin()

	if err != nil {
		s.logger.Error("failed to start db transaction", zap.Error(err))
		return err
	}

	for i, id := range charIDs {
		_, err := transaction.Exec("UPDATE guild_characters SET order_index = $1 WHERE character_id = $2", 2+i, id)

		if err != nil {
			err = transaction.Rollback()

			if err != nil {
				s.logger.Error("failed to rollback db transaction", zap.Error(err))
			}

			return err
		}
	}

	err = transaction.Commit()

	if err != nil {
		s.logger.Error("failed to commit db transaction", zap.Error(err))
		return err
	}

	return nil
}

func (guild *Guild) GetApplicationForCharID(s *Session, charID uint32, applicationType GuildApplicationType) (*GuildApplication, error) {
	row := s.server.db.QueryRowx(`
		SELECT * from guild_applications WHERE character_id = $1 AND guild_id = $2 AND application_type = $3
	`, charID, guild.ID, applicationType)

	application := &GuildApplication{}

	err := row.StructScan(application)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	if err != nil {
		s.logger.Error(
			"failed to retrieve guild application for character",
			zap.Error(err),
			zap.Uint32("charID", charID),
			zap.Uint32("guildID", guild.ID),
		)
		return nil, err
	}

	return application, nil
}

func (guild *Guild) HasApplicationForCharID(s *Session, charID uint32) (bool, error) {
	row := s.server.db.QueryRowx(`
		SELECT 1 from guild_applications WHERE character_id = $1 AND guild_id = $2
	`, charID, guild.ID)

	num := 0

	err := row.Scan(&num)

	if errors.Is(err, sql.ErrNoRows) {
		return false, nil
	}

	if err != nil {
		s.logger.Error(
			"failed to retrieve guild applications for character",
			zap.Error(err),
			zap.Uint32("charID", charID),
			zap.Uint32("guildID", guild.ID),
		)
		return false, err
	}

	return true, nil
}

func CreateGuild(s *Session, guildName string) (int32, error) {
	transaction, err := s.server.db.Begin()

	if err != nil {
		s.logger.Error("failed to start db transaction", zap.Error(err))
		return 0, err
	}

	if err != nil {
		panic(err)
	}

	guildResult, err := transaction.Query(
		"INSERT INTO guilds (name, leader_id) VALUES ($1, $2) RETURNING id",
		guildName, s.charID,
	)

	if err != nil {
		s.logger.Error("failed to create guild", zap.Error(err))
		rollbackTransaction(s, transaction)
		return 0, err
	}

	var guildId int32

	guildResult.Next()

	err = guildResult.Scan(&guildId)

	if err != nil {
		s.logger.Error("failed to retrieve guild ID", zap.Error(err))
		rollbackTransaction(s, transaction)
		return 0, err
	}

	err = guildResult.Close()

	if err != nil {
		s.logger.Error("failed to finalise query", zap.Error(err))
		rollbackTransaction(s, transaction)
		return 0, err
	}

	_, err = transaction.Exec(`
		INSERT INTO guild_characters (guild_id, character_id)
		VALUES ($1, $2)
	`, guildId, s.charID)

	if err != nil {
		s.logger.Error("failed to add character to guild", zap.Error(err))
		rollbackTransaction(s, transaction)
		return 0, err
	}

	err = transaction.Commit()

	if err != nil {
		s.logger.Error("failed to commit guild creation", zap.Error(err))
		return 0, err
	}

	return guildId, nil
}

func rollbackTransaction(s *Session, transaction *sql.Tx) {
	err := transaction.Rollback()

	if err != nil {
		s.logger.Error("failed to rollback transaction", zap.Error(err))
	}
}

func GetGuildInfoByID(s *Session, guildID uint32) (*Guild, error) {
	rows, err := s.server.db.Queryx(fmt.Sprintf(`
		%s
		WHERE g.id = $1
		LIMIT 1
	`, guildInfoSelectQuery), guildID)

	if err != nil {
		s.logger.Error("failed to retrieve guild", zap.Error(err), zap.Uint32("guildID", guildID))
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	hasRow := rows.Next()

	if !hasRow {
		return nil, nil
	}

	return buildGuildObjectFromDbResult(rows, err, s)
}

func GetGuildInfoByCharacterId(s *Session, charID uint32) (*Guild, error) {
	rows, err := s.server.db.Queryx(fmt.Sprintf(`
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
	`, guildInfoSelectQuery), charID)

	if err != nil {
		s.logger.Error("failed to retrieve guild for character", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}

	defer func() { _ = rows.Close() }()

	hasRow := rows.Next()

	if !hasRow {
		return nil, nil
	}

	return buildGuildObjectFromDbResult(rows, err, s)
}

func buildGuildObjectFromDbResult(result *sqlx.Rows, _ error, s *Session) (*Guild, error) {
	guild := &Guild{}

	err := result.StructScan(guild)

	if err != nil {
		s.logger.Error("failed to retrieve guild data from database", zap.Error(err))
		return nil, err
	}

	return guild, nil
}

func handleMsgMhfCreateGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateGuild)

	guildId, err := CreateGuild(s, pkt.Name)

	if err != nil {
		bf := byteframe.NewByteFrame()

		// No reasoning behind these values other than they cause a 'failed to create'
		// style message, it's better than nothing for now.
		bf.WriteUint32(0x01010101)

		doAckSimpleFail(s, pkt.AckHandle, bf.Data())
		return
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint32(uint32(guildId))

	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfArrangeGuildMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfArrangeGuildMember)

	guild, err := GetGuildInfoByID(s, pkt.GuildID)

	if err != nil {
		s.logger.Error(
			"failed to respond to ArrangeGuildMember message",
			zap.Uint32("charID", s.charID),
		)
		return
	}

	if guild.LeaderCharID != s.charID {
		s.logger.Error("non leader attempting to rearrange guild members!",
			zap.Uint32("charID", s.charID),
			zap.Uint32("guildID", guild.ID),
		)
		return
	}

	err = guild.ArrangeCharacters(s, pkt.CharIDs)

	if err != nil {
		s.logger.Error(
			"failed to respond to ArrangeGuildMember message",
			zap.Uint32("charID", s.charID),
			zap.Uint32("guildID", guild.ID),
		)
		return
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEnumerateGuildMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildMember)

	var guild *Guild
	var err error

	if pkt.GuildID > 0 {
		guild, err = GetGuildInfoByID(s, pkt.GuildID)
	} else {
		guild, err = GetGuildInfoByCharacterId(s, s.charID)
	}

	if guild != nil {
		isApplicant, _ := guild.HasApplicationForCharID(s, s.charID)
		if isApplicant {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
			return
		}
	}

	if guild == nil && s.prevGuildID > 0 {
		guild, err = GetGuildInfoByID(s, s.prevGuildID)
	}

	if err != nil {
		s.logger.Warn("failed to retrieve guild sending no result message")
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
		return
	} else if guild == nil {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
		return
	}

	guildMembers, err := GetGuildMembers(s, guild.ID, false)

	if err != nil {
		s.logger.Error("failed to retrieve guild")
		return
	}

	alliance, err := GetAllianceData(s, guild.AllianceID)
	if err != nil {
		s.logger.Error("Failed to get alliance data")
		return
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint16(uint16(len(guildMembers)))

	sort.Slice(guildMembers[:], func(i, j int) bool {
		return guildMembers[i].OrderIndex < guildMembers[j].OrderIndex
	})

	for _, member := range guildMembers {
		bf.WriteUint32(member.CharID)
		bf.WriteUint16(member.HR)
		if s.server.erupeConfig.RealClientMode >= _config.G10 {
			bf.WriteUint16(member.GR)
		}
		if s.server.erupeConfig.RealClientMode < _config.ZZ {
			// Magnet Spike crash workaround
			bf.WriteUint16(0)
		} else {
			bf.WriteUint16(member.WeaponID)
		}
		if member.WeaponType == 1 || member.WeaponType == 5 || member.WeaponType == 10 { // If weapon is ranged
			bf.WriteUint8(7)
		} else {
			bf.WriteUint8(6)
		}
		bf.WriteUint16(member.OrderIndex)
		bf.WriteBool(member.AvoidLeadership)
		ps.Uint8(bf, member.Name, true)
	}

	for _, member := range guildMembers {
		bf.WriteUint32(member.LastLogin)
	}

	if guild.AllianceID > 0 {
		bf.WriteUint16(alliance.TotalMembers - uint16(len(guildMembers)))
		if guild.ID != alliance.ParentGuildID {
			mems, err := GetGuildMembers(s, alliance.ParentGuildID, false)
			if err != nil {
				panic(err)
			}
			for _, m := range mems {
				bf.WriteUint32(m.CharID)
			}
		}
		if guild.ID != alliance.SubGuild1ID {
			mems, err := GetGuildMembers(s, alliance.SubGuild1ID, false)
			if err != nil {
				panic(err)
			}
			for _, m := range mems {
				bf.WriteUint32(m.CharID)
			}
		}
		if guild.ID != alliance.SubGuild2ID {
			mems, err := GetGuildMembers(s, alliance.SubGuild2ID, false)
			if err != nil {
				panic(err)
			}
			for _, m := range mems {
				bf.WriteUint32(m.CharID)
			}
		}
	} else {
		bf.WriteUint16(0)
	}

	for _, member := range guildMembers {
		bf.WriteUint16(member.RPToday)
		bf.WriteUint16(member.RPYesterday)
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetGuildManageRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildManageRight)

	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	if guild == nil || s.prevGuildID != 0 {
		guild, err := GetGuildInfoByID(s, s.prevGuildID)
		s.prevGuildID = 0
		if guild == nil || err != nil {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(guild.MemberCount))
	members, _ := GetGuildMembers(s, guild.ID, false)
	for _, member := range members {
		bf.WriteUint32(member.CharID)
		bf.WriteBool(member.Recruiter)
		bf.WriteBytes(make([]byte, 3))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetUdGuildMapInfo(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetUdGuildMapInfo)
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetGuildTargetMemberNum(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildTargetMemberNum)

	var guild *Guild
	var err error

	if pkt.GuildID == 0x0 {
		guild, err = GetGuildInfoByCharacterId(s, s.charID)
	} else {
		guild, err = GetGuildInfoByID(s, pkt.GuildID)
	}

	if err != nil || guild == nil {
		doAckBufSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x02})
		return
	}

	bf := byteframe.NewByteFrame()

	bf.WriteUint16(0x0)
	bf.WriteUint16(guild.MemberCount - 1)

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func guildGetItems(s *Session, guildID uint32) []mhfitem.MHFItemStack {
	var data []byte
	var items []mhfitem.MHFItemStack
	if err := s.server.db.QueryRow(`SELECT item_box FROM guilds WHERE id=$1`, guildID).Scan(&data); err != nil && !errors.Is(err, sql.ErrNoRows) {
		s.logger.Error("Failed to get guild item box", zap.Error(err))
	}
	if len(data) > 0 {
		box := byteframe.NewByteFrameFromBytes(data)
		numStacks := box.ReadUint16()
		box.ReadUint16() // Unused
		for i := 0; i < int(numStacks); i++ {
			items = append(items, mhfitem.ReadWarehouseItem(box))
		}
	}
	return items
}

func handleMsgMhfEnumerateGuildItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildItem)
	items := guildGetItems(s, pkt.GuildID)
	bf := byteframe.NewByteFrame()
	bf.WriteBytes(mhfitem.SerializeWarehouseItems(items))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfUpdateGuildItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuildItem)
	newStacks := mhfitem.DiffItemStacks(guildGetItems(s, pkt.GuildID), pkt.UpdatedItems)
	if _, err := s.server.db.Exec(`UPDATE guilds SET item_box=$1 WHERE id=$2`, mhfitem.SerializeWarehouseItems(newStacks), pkt.GuildID); err != nil {
		s.logger.Error("Failed to update guild item box", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateGuildIcon(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuildIcon)

	guild, err := GetGuildInfoByID(s, pkt.GuildID)

	if err != nil {
		panic(err)
	}

	characterInfo, err := GetCharacterGuildData(s, s.charID)

	if err != nil {
		panic(err)
	}

	if !characterInfo.IsSubLeader() && !characterInfo.IsLeader {
		s.logger.Warn(
			"character without leadership attempting to update guild icon",
			zap.Uint32("guildID", guild.ID),
			zap.Uint32("charID", s.charID),
		)
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	icon := &GuildIcon{}

	icon.Parts = make([]GuildIconPart, len(pkt.IconParts))

	for i, p := range pkt.IconParts {
		icon.Parts[i] = GuildIconPart{
			Index:    p.Index,
			ID:       p.ID,
			Page:     p.Page,
			Size:     p.Size,
			Rotation: p.Rotation,
			Red:      p.Red,
			Green:    p.Green,
			Blue:     p.Blue,
			PosX:     p.PosX,
			PosY:     p.PosY,
		}
	}

	guild.Icon = icon

	err = guild.Save(s)

	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfReadGuildcard(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfReadGuildcard)

	resp := byteframe.NewByteFrame()
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)
	resp.WriteUint32(0)

	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEntryRookieGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryRookieGuild)
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateForceGuildRank(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGenerateUdGuildMap(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGenerateUdGuildMap)
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetGuildManageRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetGuildManageRight)
	if _, err := s.server.db.Exec("UPDATE guild_characters SET recruiter=$1 WHERE character_id=$2", pkt.Allowed, pkt.CharID); err != nil {
		s.logger.Error("Failed to update guild manage right", zap.Error(err))
	}
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfCheckMonthlyItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCheckMonthlyItem)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
	// TODO: Implement month-by-month tracker, 0 = Not claimed, 1 = Claimed
	// Also handles HLC and EXC items, IDs = 064D, 076B
}

func handleMsgMhfAcquireMonthlyItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireMonthlyItem)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEnumerateInvGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateInvGuild)
	stubEnumerateNoResults(s, pkt.AckHandle)
}

func handleMsgMhfOperationInvGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperationInvGuild)
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateGuildcard(s *Session, p mhfpacket.MHFPacket) {}
