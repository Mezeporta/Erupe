package channelserver

import (
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"errors"
	"erupe-ce/common/mhfitem"
	_config "erupe-ce/config"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
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
		guild.PugiOutfit1, guild.PugiOutfit2, guild.PugiOutfit3, guild.PugiOutfits, guild.Icon, guild.GuildLeader.LeaderCharID)

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

	defer rows.Close()

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

	defer rows.Close()

	hasRow := rows.Next()

	if !hasRow {
		return nil, nil
	}

	return buildGuildObjectFromDbResult(rows, err, s)
}

func buildGuildObjectFromDbResult(result *sqlx.Rows, err error, s *Session) (*Guild, error) {
	guild := &Guild{}

	err = result.StructScan(guild)

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

func handleMsgMhfOperateGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuild)

	guild, err := GetGuildInfoByID(s, pkt.GuildID)
	characterGuildInfo, err := GetCharacterGuildData(s, s.charID)
	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	bf := byteframe.NewByteFrame()

	switch pkt.Action {
	case mhfpacket.OperateGuildDisband:
		response := 1
		if guild.LeaderCharID != s.charID {
			s.logger.Warn(fmt.Sprintf("character '%d' is attempting to manage guild '%d' without permission", s.charID, guild.ID))
			response = 0
		} else {
			err = guild.Disband(s)
			if err != nil {
				response = 0
			}
		}
		bf.WriteUint32(uint32(response))
	case mhfpacket.OperateGuildResign:
		guildMembers, err := GetGuildMembers(s, guild.ID, false)
		if err == nil {
			sort.Slice(guildMembers[:], func(i, j int) bool {
				return guildMembers[i].OrderIndex < guildMembers[j].OrderIndex
			})
			for i := 1; i < len(guildMembers); i++ {
				if !guildMembers[i].AvoidLeadership {
					guild.LeaderCharID = guildMembers[i].CharID
					guildMembers[0].OrderIndex = guildMembers[i].OrderIndex
					guildMembers[i].OrderIndex = 1
					guildMembers[0].Save(s)
					guildMembers[i].Save(s)
					bf.WriteUint32(guildMembers[i].CharID)
					break
				}
			}
			guild.Save(s)
		}
	case mhfpacket.OperateGuildApply:
		err = guild.CreateApplication(s, s.charID, GuildApplicationTypeApplied, nil)
		if err == nil {
			bf.WriteUint32(guild.LeaderCharID)
		} else {
			bf.WriteUint32(0)
		}
	case mhfpacket.OperateGuildLeave:
		if characterGuildInfo.IsApplicant {
			err = guild.RejectApplication(s, s.charID)
		} else {
			err = guild.RemoveCharacter(s, s.charID)
		}
		response := 1
		if err != nil {
			response = 0
		} else {
			mail := Mail{
				RecipientID:     s.charID,
				Subject:         "Withdrawal",
				Body:            fmt.Sprintf("You have withdrawn from 「%s」.", guild.Name),
				IsSystemMessage: true,
			}
			mail.Send(s, nil)
		}
		bf.WriteUint32(uint32(response))
	case mhfpacket.OperateGuildDonateRank:
		bf.WriteBytes(handleDonateRP(s, uint16(pkt.Data1.ReadUint32()), guild, 0))
	case mhfpacket.OperateGuildSetApplicationDeny:
		s.server.db.Exec("UPDATE guilds SET recruiting=false WHERE id=$1", guild.ID)
	case mhfpacket.OperateGuildSetApplicationAllow:
		s.server.db.Exec("UPDATE guilds SET recruiting=true WHERE id=$1", guild.ID)
	case mhfpacket.OperateGuildSetAvoidLeadershipTrue:
		handleAvoidLeadershipUpdate(s, pkt, true)
	case mhfpacket.OperateGuildSetAvoidLeadershipFalse:
		handleAvoidLeadershipUpdate(s, pkt, false)
	case mhfpacket.OperateGuildUpdateComment:
		if !characterGuildInfo.IsLeader && !characterGuildInfo.IsSubLeader() {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		guild.Comment = stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())
		guild.Save(s)
	case mhfpacket.OperateGuildUpdateMotto:
		if !characterGuildInfo.IsLeader && !characterGuildInfo.IsSubLeader() {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		_ = pkt.Data1.ReadUint16()
		guild.SubMotto = pkt.Data1.ReadUint8()
		guild.MainMotto = pkt.Data1.ReadUint8()
		guild.Save(s)
	case mhfpacket.OperateGuildRenamePugi1:
		handleRenamePugi(s, pkt.Data2, guild, 1)
	case mhfpacket.OperateGuildRenamePugi2:
		handleRenamePugi(s, pkt.Data2, guild, 2)
	case mhfpacket.OperateGuildRenamePugi3:
		handleRenamePugi(s, pkt.Data2, guild, 3)
	case mhfpacket.OperateGuildChangePugi1:
		handleChangePugi(s, uint8(pkt.Data1.ReadUint32()), guild, 1)
	case mhfpacket.OperateGuildChangePugi2:
		handleChangePugi(s, uint8(pkt.Data1.ReadUint32()), guild, 2)
	case mhfpacket.OperateGuildChangePugi3:
		handleChangePugi(s, uint8(pkt.Data1.ReadUint32()), guild, 3)
	case mhfpacket.OperateGuildUnlockOutfit:
		// TODO: This doesn't implement blocking, if someone unlocked the same outfit at the same time
		s.server.db.Exec(`UPDATE guilds SET pugi_outfits=pugi_outfits+$1 WHERE id=$2`, int(math.Pow(float64(pkt.Data1.ReadUint32()), 2)), guild.ID)
	case mhfpacket.OperateGuildDonateRoom:
		quantity := uint16(pkt.Data1.ReadUint32())
		bf.WriteBytes(handleDonateRP(s, quantity, guild, 2))
	case mhfpacket.OperateGuildDonateEvent:
		quantity := uint16(pkt.Data1.ReadUint32())
		bf.WriteBytes(handleDonateRP(s, quantity, guild, 1))
		// TODO: Move this value onto rp_yesterday and reset to 0... daily?
		s.server.db.Exec(`UPDATE guild_characters SET rp_today=rp_today+$1 WHERE character_id=$2`, quantity, s.charID)
	case mhfpacket.OperateGuildEventExchange:
		rp := uint16(pkt.Data1.ReadUint32())
		var balance uint32
		s.server.db.QueryRow(`UPDATE guilds SET event_rp=event_rp-$1 WHERE id=$2 RETURNING event_rp`, rp, guild.ID).Scan(&balance)
		bf.WriteUint32(balance)
	default:
		panic(fmt.Sprintf("unhandled operate guild action '%d'", pkt.Action))
	}

	if len(bf.Data()) > 0 {
		doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleRenamePugi(s *Session, bf *byteframe.ByteFrame, guild *Guild, num int) {
	name := stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	switch num {
	case 1:
		guild.PugiName1 = name
	case 2:
		guild.PugiName2 = name
	default:
		guild.PugiName3 = name
	}
	guild.Save(s)
}

func handleChangePugi(s *Session, outfit uint8, guild *Guild, num int) {
	switch num {
	case 1:
		guild.PugiOutfit1 = outfit
	case 2:
		guild.PugiOutfit2 = outfit
	case 3:
		guild.PugiOutfit3 = outfit
	}
	guild.Save(s)
}

func handleDonateRP(s *Session, amount uint16, guild *Guild, _type int) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0)
	saveData, err := GetCharacterSaveData(s, s.charID)
	if err != nil {
		return bf.Data()
	}
	var resetRoom bool
	if _type == 2 {
		var currentRP uint16
		s.server.db.QueryRow(`SELECT room_rp FROM guilds WHERE id = $1`, guild.ID).Scan(&currentRP)
		if currentRP+amount >= 30 {
			amount = 30 - currentRP
			resetRoom = true
		}
	}
	saveData.RP -= amount
	saveData.Save(s)
	switch _type {
	case 0:
		s.server.db.Exec(`UPDATE guilds SET rank_rp = rank_rp + $1 WHERE id = $2`, amount, guild.ID)
	case 1:
		s.server.db.Exec(`UPDATE guilds SET event_rp = event_rp + $1 WHERE id = $2`, amount, guild.ID)
	case 2:
		if resetRoom {
			s.server.db.Exec(`UPDATE guilds SET room_rp = 0 WHERE id = $1`, guild.ID)
			s.server.db.Exec(`UPDATE guilds SET room_expiry = $1 WHERE id = $2`, TimeAdjusted().Add(time.Hour*24*7), guild.ID)
		} else {
			s.server.db.Exec(`UPDATE guilds SET room_rp = room_rp + $1 WHERE id = $2`, amount, guild.ID)
		}
	}
	bf.Seek(0, 0)
	bf.WriteUint32(uint32(saveData.RP))
	return bf.Data()
}

func handleAvoidLeadershipUpdate(s *Session, pkt *mhfpacket.MsgMhfOperateGuild, avoidLeadership bool) {
	characterGuildData, err := GetCharacterGuildData(s, s.charID)

	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	characterGuildData.AvoidLeadership = avoidLeadership

	err = characterGuildData.Save(s)

	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfOperateGuildMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuildMember)

	guild, err := GetGuildInfoByCharacterId(s, pkt.CharID)

	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	actorCharacter, err := GetCharacterGuildData(s, s.charID)

	if err != nil || (!actorCharacter.IsSubLeader() && guild.LeaderCharID != s.charID) {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	var mail Mail
	switch pkt.Action {
	case mhfpacket.OPERATE_GUILD_MEMBER_ACTION_ACCEPT:
		err = guild.AcceptApplication(s, pkt.CharID)
		mail = Mail{
			RecipientID:     pkt.CharID,
			Subject:         "Accepted!",
			Body:            fmt.Sprintf("Your application to join 「%s」 was accepted.", guild.Name),
			IsSystemMessage: true,
		}
	case mhfpacket.OPERATE_GUILD_MEMBER_ACTION_REJECT:
		err = guild.RejectApplication(s, pkt.CharID)
		mail = Mail{
			RecipientID:     pkt.CharID,
			Subject:         "Rejected",
			Body:            fmt.Sprintf("Your application to join 「%s」 was rejected.", guild.Name),
			IsSystemMessage: true,
		}
	case mhfpacket.OPERATE_GUILD_MEMBER_ACTION_KICK:
		err = guild.RemoveCharacter(s, pkt.CharID)
		mail = Mail{
			RecipientID:     pkt.CharID,
			Subject:         "Kicked",
			Body:            fmt.Sprintf("You were kicked from 「%s」.", guild.Name),
			IsSystemMessage: true,
		}
	default:
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		s.logger.Warn(fmt.Sprintf("unhandled operateGuildMember action '%d'", pkt.Action))
	}

	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
	} else {
		mail.Send(s, nil)
		for _, channel := range s.server.Channels {
			for _, session := range channel.sessions {
				if session.charID == pkt.CharID {
					SendMailNotification(s, &mail, session)
				}
			}
		}
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgMhfInfoGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoGuild)

	var guild *Guild
	var err error

	if pkt.GuildID > 0 {
		guild, err = GetGuildInfoByID(s, pkt.GuildID)
	} else {
		guild, err = GetGuildInfoByCharacterId(s, s.charID)
	}

	if err == nil && guild != nil {
		s.prevGuildID = guild.ID

		guildName := stringsupport.UTF8ToSJIS(guild.Name)
		guildComment := stringsupport.UTF8ToSJIS(guild.Comment)
		guildLeaderName := stringsupport.UTF8ToSJIS(guild.LeaderName)

		characterGuildData, err := GetCharacterGuildData(s, s.charID)
		characterJoinedAt := uint32(0xFFFFFFFF)

		if characterGuildData != nil && characterGuildData.JoinedAt != nil {
			characterJoinedAt = uint32(characterGuildData.JoinedAt.Unix())
		}

		if err != nil {
			resp := byteframe.NewByteFrame()
			resp.WriteUint32(0) // Count
			resp.WriteUint8(0)  // Unk, read if count == 0.

			doAckBufSucceed(s, pkt.AckHandle, resp.Data())
			return
		}

		bf := byteframe.NewByteFrame()

		bf.WriteUint32(guild.ID)
		bf.WriteUint32(guild.LeaderCharID)
		bf.WriteUint16(guild.Rank())
		bf.WriteUint16(guild.MemberCount)

		bf.WriteUint8(guild.MainMotto)
		bf.WriteUint8(guild.SubMotto)

		// Unk appears to be static
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		bf.WriteUint8(0)

		flags := uint8(0)
		if !guild.Recruiting {
			flags |= 0x01
		}
		//if guild.Suspended {
		//	flags |= 0x02
		//}
		bf.WriteUint8(flags)

		if characterGuildData == nil || characterGuildData.IsApplicant {
			bf.WriteUint16(0)
		} else if guild.LeaderCharID == s.charID {
			bf.WriteUint16(1)
		} else {
			bf.WriteUint16(2)
		}

		bf.WriteUint32(uint32(guild.CreatedAt.Unix()))
		bf.WriteUint32(characterJoinedAt)
		bf.WriteUint8(uint8(len(guildName)))
		bf.WriteUint8(uint8(len(guildComment)))
		bf.WriteUint8(uint8(5)) // Length of unknown string below
		bf.WriteUint8(uint8(len(guildLeaderName)))
		bf.WriteBytes(guildName)
		bf.WriteBytes(guildComment)
		bf.WriteInt8(int8(FestivalColorCodes[guild.FestivalColor]))
		bf.WriteUint32(guild.RankRP)
		bf.WriteBytes(guildLeaderName)
		bf.WriteUint32(0)   // Unk
		bf.WriteBool(false) // isReturnGuild
		bf.WriteBool(false) // earnedSpecialHall
		bf.WriteUint8(2)
		bf.WriteUint8(2)
		bf.WriteUint32(guild.EventRP) // Skipped if last byte is <2?
		ps.Uint8(bf, guild.PugiName1, true)
		ps.Uint8(bf, guild.PugiName2, true)
		ps.Uint8(bf, guild.PugiName3, true)
		bf.WriteUint8(guild.PugiOutfit1)
		bf.WriteUint8(guild.PugiOutfit2)
		bf.WriteUint8(guild.PugiOutfit3)
		if s.server.erupeConfig.RealClientMode >= _config.Z1 {
			bf.WriteUint8(guild.PugiOutfit1)
			bf.WriteUint8(guild.PugiOutfit2)
			bf.WriteUint8(guild.PugiOutfit3)
		}
		bf.WriteUint32(guild.PugiOutfits)

		limit := s.server.erupeConfig.GameplayOptions.ClanMemberLimits[0][1]
		for _, j := range s.server.erupeConfig.GameplayOptions.ClanMemberLimits {
			if guild.Rank() >= uint16(j[0]) {
				limit = j[1]
			}
		}
		if limit > 100 {
			limit = 100
		}
		bf.WriteUint8(limit)

		bf.WriteUint32(55000)
		bf.WriteUint32(uint32(guild.RoomExpiry.Unix()))
		bf.WriteUint16(guild.RoomRP)
		bf.WriteUint16(0) // Ignored

		if guild.AllianceID > 0 {
			alliance, err := GetAllianceData(s, guild.AllianceID)
			if err != nil {
				bf.WriteUint32(0) // Error, no alliance
			} else {
				bf.WriteUint32(alliance.ID)
				bf.WriteUint32(uint32(alliance.CreatedAt.Unix()))
				bf.WriteUint16(alliance.TotalMembers)
				bf.WriteUint8(0) // Ignored
				bf.WriteUint8(0)
				ps.Uint16(bf, alliance.Name, true)
				if alliance.SubGuild1ID > 0 {
					if alliance.SubGuild2ID > 0 {
						bf.WriteUint8(3)
					} else {
						bf.WriteUint8(2)
					}
				} else {
					bf.WriteUint8(1)
				}
				bf.WriteUint32(alliance.ParentGuildID)
				bf.WriteUint32(0) // Unk1
				if alliance.ParentGuildID == guild.ID {
					bf.WriteUint16(1)
				} else {
					bf.WriteUint16(0)
				}
				bf.WriteUint16(alliance.ParentGuild.Rank())
				bf.WriteUint16(alliance.ParentGuild.MemberCount)
				ps.Uint16(bf, alliance.ParentGuild.Name, true)
				ps.Uint16(bf, alliance.ParentGuild.LeaderName, true)
				if alliance.SubGuild1ID > 0 {
					bf.WriteUint32(alliance.SubGuild1ID)
					bf.WriteUint32(0) // Unk1
					if alliance.SubGuild1ID == guild.ID {
						bf.WriteUint16(1)
					} else {
						bf.WriteUint16(0)
					}
					bf.WriteUint16(alliance.SubGuild1.Rank())
					bf.WriteUint16(alliance.SubGuild1.MemberCount)
					ps.Uint16(bf, alliance.SubGuild1.Name, true)
					ps.Uint16(bf, alliance.SubGuild1.LeaderName, true)
				}
				if alliance.SubGuild2ID > 0 {
					bf.WriteUint32(alliance.SubGuild2ID)
					bf.WriteUint32(0) // Unk1
					if alliance.SubGuild2ID == guild.ID {
						bf.WriteUint16(1)
					} else {
						bf.WriteUint16(0)
					}
					bf.WriteUint16(alliance.SubGuild2.Rank())
					bf.WriteUint16(alliance.SubGuild2.MemberCount)
					ps.Uint16(bf, alliance.SubGuild2.Name, true)
					ps.Uint16(bf, alliance.SubGuild2.LeaderName, true)
				}
			}
		} else {
			bf.WriteUint32(0) // No alliance
		}

		applicants, err := GetGuildMembers(s, guild.ID, true)
		if err != nil || (characterGuildData != nil && !characterGuildData.CanRecruit()) {
			bf.WriteUint16(0)
		} else {
			bf.WriteUint16(uint16(len(applicants)))
			for _, applicant := range applicants {
				bf.WriteUint32(applicant.CharID)
				bf.WriteUint32(0)
				bf.WriteUint16(applicant.HR)
				if s.server.erupeConfig.RealClientMode >= _config.G10 {
					bf.WriteUint16(applicant.GR)
				}
				ps.Uint8(bf, applicant.Name, true)
			}
		}

		type Activity struct {
			Pass uint8
			Unk1 uint8
			Unk2 uint8
		}
		activity := []Activity{
			// 1,0,0 = ok
			// 0,0,0 = ng
		}
		bf.WriteUint8(uint8(len(activity)))
		for _, info := range activity {
			bf.WriteUint8(info.Pass)
			bf.WriteUint8(info.Unk1)
			bf.WriteUint8(info.Unk2)
		}

		type AllianceInvite struct {
			GuildID    uint32
			LeaderID   uint32
			Unk0       uint16
			Unk1       uint16
			Members    uint16
			GuildName  string
			LeaderName string
		}
		allianceInvites := []AllianceInvite{}
		bf.WriteUint8(uint8(len(allianceInvites)))
		for _, invite := range allianceInvites {
			bf.WriteUint32(invite.GuildID)
			bf.WriteUint32(invite.LeaderID)
			bf.WriteUint16(invite.Unk0)
			bf.WriteUint16(invite.Unk1)
			bf.WriteUint16(invite.Members)
			ps.Uint16(bf, invite.GuildName, true)
			ps.Uint16(bf, invite.LeaderName, true)
		}

		if guild.Icon != nil {
			bf.WriteUint8(uint8(len(guild.Icon.Parts)))

			for _, p := range guild.Icon.Parts {
				bf.WriteUint16(p.Index)
				bf.WriteUint16(p.ID)
				bf.WriteUint8(p.Page)
				bf.WriteUint8(p.Size)
				bf.WriteUint8(p.Rotation)
				bf.WriteUint8(p.Red)
				bf.WriteUint8(p.Green)
				bf.WriteUint8(p.Blue)
				bf.WriteUint16(p.PosX)
				bf.WriteUint16(p.PosY)
			}
		} else {
			bf.WriteUint8(0)
		}
		bf.WriteUint8(0) // Unk

		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgMhfEnumerateGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuild)

	var guilds []*Guild
	var alliances []*GuildAlliance
	var rows *sqlx.Rows
	var err error

	if pkt.Type <= 8 {
		var tempGuilds []*Guild
		rows, err = s.server.db.Queryx(guildInfoSelectQuery)
		if err == nil {
			for rows.Next() {
				guild, err := buildGuildObjectFromDbResult(rows, err, s)
				if err != nil {
					continue
				}
				tempGuilds = append(tempGuilds, guild)
			}
		}
		switch pkt.Type {
		case mhfpacket.ENUMERATE_GUILD_TYPE_GUILD_NAME:
			for _, guild := range tempGuilds {
				if strings.Contains(guild.Name, stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())) {
					guilds = append(guilds, guild)
				}
			}
		case mhfpacket.ENUMERATE_GUILD_TYPE_LEADER_NAME:
			for _, guild := range tempGuilds {
				if strings.Contains(guild.LeaderName, stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())) {
					guilds = append(guilds, guild)
				}
			}
		case mhfpacket.ENUMERATE_GUILD_TYPE_LEADER_ID:
			CID := pkt.Data1.ReadUint32()
			for _, guild := range tempGuilds {
				if guild.LeaderCharID == CID {
					guilds = append(guilds, guild)
				}
			}
		case mhfpacket.ENUMERATE_GUILD_TYPE_ORDER_MEMBERS:
			if pkt.Sorting {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].MemberCount > tempGuilds[j].MemberCount
				})
			} else {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].MemberCount < tempGuilds[j].MemberCount
				})
			}
			guilds = tempGuilds
		case mhfpacket.ENUMERATE_GUILD_TYPE_ORDER_REGISTRATION:
			if pkt.Sorting {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].CreatedAt.Unix() > tempGuilds[j].CreatedAt.Unix()
				})
			} else {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].CreatedAt.Unix() < tempGuilds[j].CreatedAt.Unix()
				})
			}
			guilds = tempGuilds
		case mhfpacket.ENUMERATE_GUILD_TYPE_ORDER_RANK:
			if pkt.Sorting {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].RankRP > tempGuilds[j].RankRP
				})
			} else {
				sort.Slice(tempGuilds, func(i, j int) bool {
					return tempGuilds[i].RankRP < tempGuilds[j].RankRP
				})
			}
			guilds = tempGuilds
		case mhfpacket.ENUMERATE_GUILD_TYPE_MOTTO:
			mainMotto := uint8(pkt.Data1.ReadUint16())
			subMotto := uint8(pkt.Data1.ReadUint16())
			for _, guild := range tempGuilds {
				if guild.MainMotto == mainMotto && guild.SubMotto == subMotto {
					guilds = append(guilds, guild)
				}
			}
		case mhfpacket.ENUMERATE_GUILD_TYPE_RECRUITING:
			recruitingMotto := uint8(pkt.Data1.ReadUint16())
			for _, guild := range tempGuilds {
				if guild.MainMotto == recruitingMotto {
					guilds = append(guilds, guild)
				}
			}
		}
	}

	if pkt.Type > 8 {
		var tempAlliances []*GuildAlliance
		rows, err = s.server.db.Queryx(allianceInfoSelectQuery)
		if err == nil {
			for rows.Next() {
				alliance, _ := buildAllianceObjectFromDbResult(rows, err, s)
				tempAlliances = append(tempAlliances, alliance)
			}
		}
		switch pkt.Type {
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_ALLIANCE_NAME:
			for _, alliance := range tempAlliances {
				if strings.Contains(alliance.Name, stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())) {
					alliances = append(alliances, alliance)
				}
			}
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_LEADER_NAME:
			for _, alliance := range tempAlliances {
				if strings.Contains(alliance.ParentGuild.LeaderName, stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())) {
					alliances = append(alliances, alliance)
				}
			}
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_LEADER_ID:
			CID := pkt.Data1.ReadUint32()
			for _, alliance := range tempAlliances {
				if alliance.ParentGuild.LeaderCharID == CID {
					alliances = append(alliances, alliance)
				}
			}
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_ORDER_MEMBERS:
			if pkt.Sorting {
				sort.Slice(tempAlliances, func(i, j int) bool {
					return tempAlliances[i].TotalMembers > tempAlliances[j].TotalMembers
				})
			} else {
				sort.Slice(tempAlliances, func(i, j int) bool {
					return tempAlliances[i].TotalMembers < tempAlliances[j].TotalMembers
				})
			}
			alliances = tempAlliances
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_ORDER_REGISTRATION:
			if pkt.Sorting {
				sort.Slice(tempAlliances, func(i, j int) bool {
					return tempAlliances[i].CreatedAt.Unix() > tempAlliances[j].CreatedAt.Unix()
				})
			} else {
				sort.Slice(tempAlliances, func(i, j int) bool {
					return tempAlliances[i].CreatedAt.Unix() < tempAlliances[j].CreatedAt.Unix()
				})
			}
			alliances = tempAlliances
		}
	}

	if err != nil || (guilds == nil && alliances == nil) {
		stubEnumerateNoResults(s, pkt.AckHandle)
		return
	}

	bf := byteframe.NewByteFrame()

	if pkt.Type > 8 {
		hasNextPage := false
		if len(alliances) > 10 {
			hasNextPage = true
			alliances = alliances[:10]
		}
		bf.WriteUint16(uint16(len(alliances)))
		bf.WriteBool(hasNextPage)
		for _, alliance := range alliances {
			bf.WriteUint32(alliance.ID)
			bf.WriteUint32(alliance.ParentGuild.LeaderCharID)
			bf.WriteUint16(alliance.TotalMembers)
			bf.WriteUint16(0x0000)
			if alliance.SubGuild1ID == 0 && alliance.SubGuild2ID == 0 {
				bf.WriteUint16(1)
			} else if alliance.SubGuild1ID > 0 && alliance.SubGuild2ID == 0 || alliance.SubGuild1ID == 0 && alliance.SubGuild2ID > 0 {
				bf.WriteUint16(2)
			} else {
				bf.WriteUint16(3)
			}
			bf.WriteUint32(uint32(alliance.CreatedAt.Unix()))
			ps.Uint8(bf, alliance.Name, true)
			ps.Uint8(bf, alliance.ParentGuild.LeaderName, true)
			bf.WriteUint8(0x01) // Unk
			bf.WriteBool(true)  // TODO: Enable GuildAlliance applications
		}
	} else {
		hasNextPage := false
		if len(guilds) > 10 {
			hasNextPage = true
			guilds = guilds[:10]
		}
		bf.WriteUint16(uint16(len(guilds)))
		bf.WriteBool(hasNextPage)
		for _, guild := range guilds {
			bf.WriteUint32(guild.ID)
			bf.WriteUint32(guild.LeaderCharID)
			bf.WriteUint16(guild.MemberCount)
			bf.WriteUint16(0x0000) // Unk
			bf.WriteUint16(guild.Rank())
			bf.WriteUint32(uint32(guild.CreatedAt.Unix()))
			ps.Uint8(bf, guild.Name, true)
			ps.Uint8(bf, guild.LeaderName, true)
			bf.WriteUint8(0x01) // Unk
			bf.WriteBool(!guild.Recruiting)
		}
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
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

	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if guild == nil || s.prevGuildID != 0 {
		guild, err = GetGuildInfoByID(s, s.prevGuildID)
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
	s.server.db.QueryRow(`SELECT item_box FROM guilds WHERE id=$1`, guildID).Scan(&data)
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
	s.server.db.Exec(`UPDATE guilds SET item_box=$1 WHERE id=$2`, mhfitem.SerializeWarehouseItems(newStacks), pkt.GuildID)
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

type GuildMission struct {
	ID          uint32
	Unk         uint32
	Type        uint16
	Goal        uint16
	Quantity    uint16
	SkipTickets uint16
	GR          bool
	RewardType  uint16
	RewardLevel uint16
}

func handleMsgMhfGetGuildMissionList(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildMissionList)
	bf := byteframe.NewByteFrame()
	missions := []GuildMission{
		{431201, 574, 1, 4761, 35, 1, false, 2, 1},
		{431202, 755, 0, 95, 12, 2, false, 3, 2},
		{431203, 746, 0, 95, 6, 1, false, 1, 1},
		{431204, 581, 0, 83, 16, 2, false, 4, 2},
		{431205, 694, 1, 4763, 25, 1, false, 2, 1},
		{431206, 988, 0, 27, 16, 1, false, 6, 1},
		{431207, 730, 1, 4768, 25, 1, false, 4, 1},
		{431208, 680, 1, 3567, 50, 2, false, 2, 2},
		{431209, 1109, 0, 34, 60, 2, false, 6, 2},
		{431210, 128, 1, 8921, 70, 2, false, 3, 2},
		{431211, 406, 0, 59, 10, 1, false, 1, 1},
		{431212, 1170, 0, 70, 90, 3, false, 6, 3},
		{431213, 164, 0, 38, 24, 2, false, 6, 2},
		{431214, 378, 1, 3556, 150, 3, false, 1, 3},
		{431215, 446, 0, 94, 20, 2, false, 4, 2},
	}
	for _, mission := range missions {
		bf.WriteUint32(mission.ID)
		bf.WriteUint32(mission.Unk)
		bf.WriteUint16(mission.Type)
		bf.WriteUint16(mission.Goal)
		bf.WriteUint16(mission.Quantity)
		bf.WriteUint16(mission.SkipTickets)
		bf.WriteBool(mission.GR)
		bf.WriteUint16(mission.RewardType)
		bf.WriteUint16(mission.RewardLevel)
		bf.WriteUint32(uint32(TimeAdjusted().Unix()))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetGuildMissionRecord(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildMissionRecord)

	// No guild mission records = 0x190 empty bytes
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 0x190))
}

func handleMsgMhfAddGuildMissionCount(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddGuildMissionCount)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfSetGuildMissionTarget(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetGuildMissionTarget)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfCancelGuildMissionTarget(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCancelGuildMissionTarget)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

type GuildMeal struct {
	ID        uint32    `db:"id"`
	MealID    uint32    `db:"meal_id"`
	Level     uint32    `db:"level"`
	CreatedAt time.Time `db:"created_at"`
}

func handleMsgMhfLoadGuildCooking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadGuildCooking)
	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	data, err := s.server.db.Queryx("SELECT id, meal_id, level, created_at FROM guild_meals WHERE guild_id = $1", guild.ID)
	if err != nil {
		s.logger.Error("Failed to get guild meals from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 2))
		return
	}
	var meals []GuildMeal
	var temp GuildMeal
	for data.Next() {
		err = data.StructScan(&temp)
		if err != nil {
			continue
		}
		if temp.CreatedAt.Add(60 * time.Minute).After(TimeAdjusted()) {
			meals = append(meals, temp)
		}
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(meals)))
	for _, meal := range meals {
		bf.WriteUint32(meal.ID)
		bf.WriteUint32(meal.MealID)
		bf.WriteUint32(meal.Level)
		bf.WriteUint32(uint32(meal.CreatedAt.Unix()))
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfRegistGuildCooking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildCooking)
	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	startTime := TimeAdjusted().Add(time.Duration(s.server.erupeConfig.GameplayOptions.ClanMealDuration-3600) * time.Second)
	if pkt.OverwriteID != 0 {
		s.server.db.Exec("UPDATE guild_meals SET meal_id = $1, level = $2, created_at = $3 WHERE id = $4", pkt.MealID, pkt.Success, startTime, pkt.OverwriteID)
	} else {
		s.server.db.QueryRow("INSERT INTO guild_meals (guild_id, meal_id, level, created_at) VALUES ($1, $2, $3, $4) RETURNING id", guild.ID, pkt.MealID, pkt.Success, startTime).Scan(&pkt.OverwriteID)
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(1)
	bf.WriteUint32(pkt.OverwriteID)
	bf.WriteUint32(uint32(pkt.MealID))
	bf.WriteUint32(uint32(pkt.Success))
	bf.WriteUint32(uint32(startTime.Unix()))
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGetGuildWeeklyBonusMaster(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildWeeklyBonusMaster)

	// Values taken from brand new guild capture
	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 40))
}
func handleMsgMhfGetGuildWeeklyBonusActiveCount(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildWeeklyBonusActiveCount)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(60) // Active count
	bf.WriteUint8(60) // Current active count
	bf.WriteUint8(0)  // New active count
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfGuildHuntdata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGuildHuntdata)
	bf := byteframe.NewByteFrame()
	switch pkt.Operation {
	case 0: // Acquire
		s.server.db.Exec(`UPDATE guild_characters SET box_claimed=$1 WHERE character_id=$2`, TimeAdjusted(), s.charID)
	case 1: // Enumerate
		bf.WriteUint8(0) // Entries
		rows, err := s.server.db.Query(`SELECT kl.id, kl.monster FROM kill_logs kl
			INNER JOIN guild_characters gc ON kl.character_id = gc.character_id
			WHERE gc.guild_id=$1
			AND kl.timestamp >= (SELECT box_claimed FROM guild_characters WHERE character_id=$2)
		`, pkt.GuildID, s.charID)
		if err == nil {
			var count uint8
			var huntID, monID uint32
			for rows.Next() {
				err = rows.Scan(&huntID, &monID)
				if err != nil {
					continue
				}
				count++
				if count > 255 {
					count = 255
					rows.Close()
					break
				}
				bf.WriteUint32(huntID)
				bf.WriteUint32(monID)
			}
			bf.Seek(0, 0)
			bf.WriteUint8(count)
		}
	case 2: // Check
		guild, err := GetGuildInfoByCharacterId(s, s.charID)
		if err == nil {
			var count uint8
			err = s.server.db.QueryRow(`SELECT COUNT(*) FROM kill_logs kl
				INNER JOIN guild_characters gc ON kl.character_id = gc.character_id
				WHERE gc.guild_id=$1
				AND kl.timestamp >= (SELECT box_claimed FROM guild_characters WHERE character_id=$2)
			`, guild.ID, s.charID).Scan(&count)
			if err == nil && count > 0 {
				bf.WriteBool(true)
			} else {
				bf.WriteBool(false)
			}
		} else {
			bf.WriteBool(false)
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

type MessageBoardPost struct {
	ID        uint32    `db:"id"`
	StampID   uint32    `db:"stamp_id"`
	Title     string    `db:"title"`
	Body      string    `db:"body"`
	AuthorID  uint32    `db:"author_id"`
	Timestamp time.Time `db:"created_at"`
	LikedBy   string    `db:"liked_by"`
}

func handleMsgMhfEnumerateGuildMessageBoard(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildMessageBoard)
	guild, _ := GetGuildInfoByCharacterId(s, s.charID)
	if pkt.BoardType == 1 {
		pkt.MaxPosts = 4
	}
	msgs, err := s.server.db.Queryx("SELECT id, stamp_id, title, body, author_id, created_at, liked_by FROM guild_posts WHERE guild_id = $1 AND post_type = $2 ORDER BY created_at DESC", guild.ID, int(pkt.BoardType))
	if err != nil {
		s.logger.Error("Failed to get guild messages from db", zap.Error(err))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	s.server.db.Exec("UPDATE characters SET guild_post_checked = now() WHERE id = $1", s.charID)
	bf := byteframe.NewByteFrame()
	var postCount uint32
	for msgs.Next() {
		postData := &MessageBoardPost{}
		err = msgs.StructScan(&postData)
		if err != nil {
			continue
		}
		postCount++
		bf.WriteUint32(postData.ID)
		bf.WriteUint32(postData.AuthorID)
		bf.WriteUint32(0)
		bf.WriteUint32(uint32(postData.Timestamp.Unix()))
		bf.WriteUint32(uint32(stringsupport.CSVLength(postData.LikedBy)))
		bf.WriteBool(stringsupport.CSVContains(postData.LikedBy, int(s.charID)))
		bf.WriteUint32(postData.StampID)
		ps.Uint32(bf, postData.Title, true)
		ps.Uint32(bf, postData.Body, true)
	}
	data := byteframe.NewByteFrame()
	data.WriteUint32(postCount)
	data.WriteBytes(bf.Data())
	doAckBufSucceed(s, pkt.AckHandle, data.Data())
}

func handleMsgMhfUpdateGuildMessageBoard(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfUpdateGuildMessageBoard)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	applicant := false
	if guild != nil {
		applicant, _ = guild.HasApplicationForCharID(s, s.charID)
	}
	if err != nil || guild == nil || applicant {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	switch pkt.MessageOp {
	case 0: // Create message
		s.server.db.Exec("INSERT INTO guild_posts (guild_id, author_id, stamp_id, post_type, title, body) VALUES ($1, $2, $3, $4, $5, $6)", guild.ID, s.charID, pkt.StampID, pkt.PostType, pkt.Title, pkt.Body)
		// TODO: if there are too many messages, purge excess
	case 1: // Delete message
		s.server.db.Exec("DELETE FROM guild_posts WHERE id = $1", pkt.PostID)
	case 2: // Update message
		s.server.db.Exec("UPDATE guild_posts SET title = $1, body = $2 WHERE id = $3", pkt.Title, pkt.Body, pkt.PostID)
	case 3: // Update stamp
		s.server.db.Exec("UPDATE guild_posts SET stamp_id = $1 WHERE id = $2", pkt.StampID, pkt.PostID)
	case 4: // Like message
		var likedBy string
		err := s.server.db.QueryRow("SELECT liked_by FROM guild_posts WHERE id = $1", pkt.PostID).Scan(&likedBy)
		if err != nil {
			s.logger.Error("Failed to get guild message like data from db", zap.Error(err))
		} else {
			if pkt.LikeState {
				likedBy = stringsupport.CSVAdd(likedBy, int(s.charID))
				s.server.db.Exec("UPDATE guild_posts SET liked_by = $1 WHERE id = $2", likedBy, pkt.PostID)
			} else {
				likedBy = stringsupport.CSVRemove(likedBy, int(s.charID))
				s.server.db.Exec("UPDATE guild_posts SET liked_by = $1 WHERE id = $2", likedBy, pkt.PostID)
			}
		}
	case 5: // Check for new messages
		var timeChecked time.Time
		var newPosts int
		err := s.server.db.QueryRow("SELECT guild_post_checked FROM characters WHERE id = $1", s.charID).Scan(&timeChecked)
		if err == nil {
			s.server.db.QueryRow("SELECT COUNT(*) FROM guild_posts WHERE guild_id = $1 AND (EXTRACT(epoch FROM created_at)::int) > $2", guild.ID, timeChecked.Unix()).Scan(&newPosts)
			if newPosts > 0 {
				doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x01})
				return
			}
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEntryRookieGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryRookieGuild)
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateForceGuildRank(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAddGuildWeeklyBonusExceptionalUser(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAddGuildWeeklyBonusExceptionalUser)
	// TODO: record pkt.NumUsers to DB
	// must use addition
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfGenerateUdGuildMap(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGenerateUdGuildMap)
	doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfUpdateGuild(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfSetGuildManageRight(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSetGuildManageRight)
	s.server.db.Exec("UPDATE guild_characters SET recruiter=$1 WHERE character_id=$2", pkt.Allowed, pkt.CharID)
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
