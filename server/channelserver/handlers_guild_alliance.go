package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"time"

	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// GuildAlliance represents a multi-guild alliance.
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

func handleMsgMhfCreateJoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfCreateJoint)
	if err := s.server.guildRepo.CreateAlliance(pkt.Name, pkt.GuildID); err != nil {
		s.logger.Error("Failed to create guild alliance in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x01, 0x01, 0x01, 0x01})
}

func handleMsgMhfOperateJoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateJoint)

	guild, err := s.server.guildRepo.GetByID(pkt.GuildID)
	if err != nil {
		s.logger.Error("Failed to get guild info", zap.Error(err))
	}
	alliance, err := s.server.guildRepo.GetAllianceByID(pkt.AllianceID)
	if err != nil {
		s.logger.Error("Failed to get alliance info", zap.Error(err))
	}

	switch pkt.Action {
	case mhfpacket.OPERATE_JOINT_DISBAND:
		if guild.LeaderCharID == s.charID && alliance.ParentGuildID == guild.ID {
			if err := s.server.guildRepo.DeleteAlliance(alliance.ID); err != nil {
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
			if err := s.server.guildRepo.RemoveGuildFromAlliance(alliance.ID, guild.ID, alliance.SubGuild1ID, alliance.SubGuild2ID); err != nil {
				s.logger.Error("Failed to remove guild from alliance", zap.Error(err))
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
			if err := s.server.guildRepo.RemoveGuildFromAlliance(alliance.ID, kickedGuildID, alliance.SubGuild1ID, alliance.SubGuild2ID); err != nil {
				s.logger.Error("Failed to kick guild from alliance", zap.Error(err))
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
		s.logger.Error("unhandled operate joint action", zap.Uint8("action", uint8(pkt.Action)))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleMsgMhfInfoJoint(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoJoint)
	bf := byteframe.NewByteFrame()
	alliance, err := s.server.guildRepo.GetAllianceByID(pkt.AllianceID)
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
		bf.WriteUint16(alliance.ParentGuild.Rank(s.server.erupeConfig.RealClientMode))
		bf.WriteUint16(alliance.ParentGuild.MemberCount)
		ps.Uint16(bf, alliance.ParentGuild.Name, true)
		ps.Uint16(bf, alliance.ParentGuild.LeaderName, true)
		if alliance.SubGuild1ID > 0 {
			bf.WriteUint32(alliance.SubGuild1ID)
			bf.WriteUint32(alliance.SubGuild1.LeaderCharID)
			bf.WriteUint16(alliance.SubGuild1.Rank(s.server.erupeConfig.RealClientMode))
			bf.WriteUint16(alliance.SubGuild1.MemberCount)
			ps.Uint16(bf, alliance.SubGuild1.Name, true)
			ps.Uint16(bf, alliance.SubGuild1.LeaderName, true)
		}
		if alliance.SubGuild2ID > 0 {
			bf.WriteUint32(alliance.SubGuild2ID)
			bf.WriteUint32(alliance.SubGuild2.LeaderCharID)
			bf.WriteUint16(alliance.SubGuild2.Rank(s.server.erupeConfig.RealClientMode))
			bf.WriteUint16(alliance.SubGuild2.MemberCount)
			ps.Uint16(bf, alliance.SubGuild2.Name, true)
			ps.Uint16(bf, alliance.SubGuild2.LeaderName, true)
		}
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	}
}
