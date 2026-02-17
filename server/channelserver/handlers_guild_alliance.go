package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"fmt"
	"time"

	"erupe-ce/network/mhfpacket"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

const allianceInfoSelectQuery = `
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

type GuildAlliance struct {
	ID           uint32    `db:"id"`
	Name         string    `db:"name"`
	CreatedAt    time.Time `db:"created_at"`
	TotalMembers uint16

	ParentGuildID uint32 `db:"parent_id"`
	SubGuild1ID   uint32 `db:"sub1_id"`
	SubGuild2ID   uint32 `db:"sub2_id"`

	ParentGuild Guild
	SubGuild1   Guild
	SubGuild2   Guild
}

func GetAllianceData(s *Session, AllianceID uint32) (*GuildAlliance, error) {
	rows, err := s.server.db.Queryx(fmt.Sprintf(`
		%s
		WHERE ga.id = $1
	`, allianceInfoSelectQuery), AllianceID)
	if err != nil {
		s.logger.Error("Failed to retrieve alliance data from database", zap.Error(err))
		return nil, err
	}
	defer rows.Close()
	hasRow := rows.Next()
	if !hasRow {
		return nil, nil
	}

	return buildAllianceObjectFromDbResult(rows, err, s)
}

func buildAllianceObjectFromDbResult(result *sqlx.Rows, _ error, s *Session) (*GuildAlliance, error) {
	alliance := &GuildAlliance{}

	err := result.StructScan(alliance)

	if err != nil {
		s.logger.Error("failed to retrieve alliance from database", zap.Error(err))
		return nil, err
	}

	parentGuild, err := GetGuildInfoByID(s, alliance.ParentGuildID)
	if err != nil {
		s.logger.Error("Failed to get parent guild info", zap.Error(err))
		return nil, err
	} else {
		alliance.ParentGuild = *parentGuild
		alliance.TotalMembers += parentGuild.MemberCount
	}

	if alliance.SubGuild1ID > 0 {
		subGuild1, err := GetGuildInfoByID(s, alliance.SubGuild1ID)
		if err != nil {
			s.logger.Error("Failed to get sub guild 1 info", zap.Error(err))
			return nil, err
		} else {
			alliance.SubGuild1 = *subGuild1
			alliance.TotalMembers += subGuild1.MemberCount
		}
	}

	if alliance.SubGuild2ID > 0 {
		subGuild2, err := GetGuildInfoByID(s, alliance.SubGuild2ID)
		if err != nil {
			s.logger.Error("Failed to get sub guild 2 info", zap.Error(err))
			return nil, err
		} else {
			alliance.SubGuild2 = *subGuild2
			alliance.TotalMembers += subGuild2.MemberCount
		}
	}

	return alliance, nil
}

func handleMsgMhfCreateJoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateJoint)
	_, err := s.server.db.Exec("INSERT INTO guild_alliances (name, parent_id) VALUES ($1, $2)", pkt.Name, pkt.GuildID)
	if err != nil {
		s.logger.Error("Failed to create guild alliance in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x01, 0x01, 0x01, 0x01})
}

func handleMsgMhfOperateJoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateJoint)

	guild, err := GetGuildInfoByID(s, pkt.GuildID)
	if err != nil {
		s.logger.Error("Failed to get guild info", zap.Error(err))
	}
	alliance, err := GetAllianceData(s, pkt.AllianceID)
	if err != nil {
		s.logger.Error("Failed to get alliance info", zap.Error(err))
	}

	switch pkt.Action {
	case mhfpacket.OPERATE_JOINT_DISBAND:
		if guild.LeaderCharID == s.charID && alliance.ParentGuildID == guild.ID {
			_, err = s.server.db.Exec("DELETE FROM guild_alliances WHERE id=$1", alliance.ID)
			if err != nil {
				s.logger.Error("Failed to disband alliance", zap.Error(err))
			}
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		} else {
			s.logger.Warn(
				"Non-owner of alliance attempted disband",
				zap.Uint32("CharID", s.charID),
				zap.Uint32("AllyID", alliance.ID),
			)
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		}
	case mhfpacket.OPERATE_JOINT_LEAVE:
		if guild.LeaderCharID == s.charID {
			if guild.ID == alliance.SubGuild1ID && alliance.SubGuild2ID > 0 {
				_, _ = s.server.db.Exec(`UPDATE guild_alliances SET sub1_id = sub2_id, sub2_id = NULL WHERE id = $1`, alliance.ID)
			} else if guild.ID == alliance.SubGuild1ID && alliance.SubGuild2ID == 0 {
				_, _ = s.server.db.Exec(`UPDATE guild_alliances SET sub1_id = NULL WHERE id = $1`, alliance.ID)
			} else {
				_, _ = s.server.db.Exec(`UPDATE guild_alliances SET sub2_id = NULL WHERE id = $1`, alliance.ID)
			}
			// NOTE: Alliance join requests are not yet implemented (no DB table exists),
			// so there are no pending applications to clean up on leave.
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		} else {
			s.logger.Warn(
				"Non-owner of guild attempted alliance leave",
				zap.Uint32("CharID", s.charID),
			)
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		}
	case mhfpacket.OPERATE_JOINT_KICK:
		if alliance.ParentGuild.LeaderCharID == s.charID {
			kickedGuildID := pkt.Data1.ReadUint32()
			if kickedGuildID == alliance.SubGuild1ID && alliance.SubGuild2ID > 0 {
				_, _ = s.server.db.Exec(`UPDATE guild_alliances SET sub1_id = sub2_id, sub2_id = NULL WHERE id = $1`, alliance.ID)
			} else if kickedGuildID == alliance.SubGuild1ID && alliance.SubGuild2ID == 0 {
				_, _ = s.server.db.Exec(`UPDATE guild_alliances SET sub1_id = NULL WHERE id = $1`, alliance.ID)
			} else {
				_, _ = s.server.db.Exec(`UPDATE guild_alliances SET sub2_id = NULL WHERE id = $1`, alliance.ID)
			}
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		} else {
			s.logger.Warn(
				"Non-owner of alliance attempted kick",
				zap.Uint32("CharID", s.charID),
				zap.Uint32("AllyID", alliance.ID),
			)
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		}
	default:
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		panic(fmt.Sprintf("Unhandled operate joint action '%d'", pkt.Action))
	}
}

func handleMsgMhfInfoJoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoJoint)
	bf := byteframe.NewByteFrame()
	alliance, err := GetAllianceData(s, pkt.AllianceID)
	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
	} else {
		bf.WriteUint32(alliance.ID)
		bf.WriteUint32(uint32(alliance.CreatedAt.Unix()))
		bf.WriteUint16(alliance.TotalMembers)
		bf.WriteUint16(0x0000) // Unk
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
		bf.WriteUint32(alliance.ParentGuild.LeaderCharID)
		bf.WriteUint16(alliance.ParentGuild.Rank())
		bf.WriteUint16(alliance.ParentGuild.MemberCount)
		ps.Uint16(bf, alliance.ParentGuild.Name, true)
		ps.Uint16(bf, alliance.ParentGuild.LeaderName, true)
		if alliance.SubGuild1ID > 0 {
			bf.WriteUint32(alliance.SubGuild1ID)
			bf.WriteUint32(alliance.SubGuild1.LeaderCharID)
			bf.WriteUint16(alliance.SubGuild1.Rank())
			bf.WriteUint16(alliance.SubGuild1.MemberCount)
			ps.Uint16(bf, alliance.SubGuild1.Name, true)
			ps.Uint16(bf, alliance.SubGuild1.LeaderName, true)
		}
		if alliance.SubGuild2ID > 0 {
			bf.WriteUint32(alliance.SubGuild2ID)
			bf.WriteUint32(alliance.SubGuild2.LeaderCharID)
			bf.WriteUint16(alliance.SubGuild2.Rank())
			bf.WriteUint16(alliance.SubGuild2.MemberCount)
			ps.Uint16(bf, alliance.SubGuild2.Name, true)
			ps.Uint16(bf, alliance.SubGuild2.LeaderName, true)
		}
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	}
}
