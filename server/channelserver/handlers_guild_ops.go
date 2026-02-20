package channelserver

import (
	"fmt"
	"sort"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

func handleMsgMhfOperateGuild(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuild)

	guild, err := GetGuildInfoByID(s, pkt.GuildID)
	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
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
					_ = guildMembers[0].Save(s)
					_ = guildMembers[i].Save(s)
					bf.WriteUint32(guildMembers[i].CharID)
					break
				}
			}
			_ = guild.Save(s)
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
			_ = mail.Send(s, nil)
		}
		bf.WriteUint32(uint32(response))
	case mhfpacket.OperateGuildDonateRank:
		bf.WriteBytes(handleDonateRP(s, uint16(pkt.Data1.ReadUint32()), guild, 0))
	case mhfpacket.OperateGuildSetApplicationDeny:
		if _, err := s.server.db.Exec("UPDATE guilds SET recruiting=false WHERE id=$1", guild.ID); err != nil {
			s.logger.Error("Failed to deny guild applications", zap.Error(err))
		}
	case mhfpacket.OperateGuildSetApplicationAllow:
		if _, err := s.server.db.Exec("UPDATE guilds SET recruiting=true WHERE id=$1", guild.ID); err != nil {
			s.logger.Error("Failed to allow guild applications", zap.Error(err))
		}
	case mhfpacket.OperateGuildSetAvoidLeadershipTrue:
		handleAvoidLeadershipUpdate(s, pkt, true)
	case mhfpacket.OperateGuildSetAvoidLeadershipFalse:
		handleAvoidLeadershipUpdate(s, pkt, false)
	case mhfpacket.OperateGuildUpdateComment:
		if !characterGuildInfo.IsLeader && !characterGuildInfo.IsSubLeader() {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		guild.Comment, _ = stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())
		_ = guild.Save(s)
	case mhfpacket.OperateGuildUpdateMotto:
		if !characterGuildInfo.IsLeader && !characterGuildInfo.IsSubLeader() {
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		_ = pkt.Data1.ReadUint16()
		guild.SubMotto = pkt.Data1.ReadUint8()
		guild.MainMotto = pkt.Data1.ReadUint8()
		_ = guild.Save(s)
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
		if _, err := s.server.db.Exec(`UPDATE guilds SET pugi_outfits=$1 WHERE id=$2`, pkt.Data1.ReadUint32(), guild.ID); err != nil {
			s.logger.Error("Failed to unlock guild pugi outfit", zap.Error(err))
		}
	case mhfpacket.OperateGuildDonateRoom:
		quantity := uint16(pkt.Data1.ReadUint32())
		bf.WriteBytes(handleDonateRP(s, quantity, guild, 2))
	case mhfpacket.OperateGuildDonateEvent:
		quantity := uint16(pkt.Data1.ReadUint32())
		bf.WriteBytes(handleDonateRP(s, quantity, guild, 1))
		// TODO: Move this value onto rp_yesterday and reset to 0... daily?
		if _, err := s.server.db.Exec(`UPDATE guild_characters SET rp_today=rp_today+$1 WHERE character_id=$2`, quantity, s.charID); err != nil {
			s.logger.Error("Failed to update guild character daily RP", zap.Error(err))
		}
	case mhfpacket.OperateGuildEventExchange:
		rp := uint16(pkt.Data1.ReadUint32())
		var balance uint32
		if err := s.server.db.QueryRow(`UPDATE guilds SET event_rp=event_rp-$1 WHERE id=$2 RETURNING event_rp`, rp, guild.ID).Scan(&balance); err != nil {
			s.logger.Error("Failed to exchange guild event RP", zap.Error(err))
		}
		bf.WriteUint32(balance)
	default:
		s.logger.Error("unhandled operate guild action", zap.Uint8("action", uint8(pkt.Action)))
	}

	if len(bf.Data()) > 0 {
		doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
	} else {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}

func handleRenamePugi(s *Session, bf *byteframe.ByteFrame, guild *Guild, num int) {
	name, _ := stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	switch num {
	case 1:
		guild.PugiName1 = name
	case 2:
		guild.PugiName2 = name
	default:
		guild.PugiName3 = name
	}
	_ = guild.Save(s)
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
	_ = guild.Save(s)
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
		if err := s.server.db.QueryRow(`SELECT room_rp FROM guilds WHERE id = $1`, guild.ID).Scan(&currentRP); err != nil {
			s.logger.Error("Failed to get guild room RP", zap.Error(err))
		}
		if currentRP+amount >= 30 {
			amount = 30 - currentRP
			resetRoom = true
		}
	}
	saveData.RP -= amount
	saveData.Save(s)
	switch _type {
	case 0:
		if _, err := s.server.db.Exec(`UPDATE guilds SET rank_rp = rank_rp + $1 WHERE id = $2`, amount, guild.ID); err != nil {
			s.logger.Error("Failed to update guild rank RP", zap.Error(err))
		}
	case 1:
		if _, err := s.server.db.Exec(`UPDATE guilds SET event_rp = event_rp + $1 WHERE id = $2`, amount, guild.ID); err != nil {
			s.logger.Error("Failed to update guild event RP", zap.Error(err))
		}
	case 2:
		if resetRoom {
			if _, err := s.server.db.Exec(`UPDATE guilds SET room_rp = 0 WHERE id = $1`, guild.ID); err != nil {
				s.logger.Error("Failed to reset guild room RP", zap.Error(err))
			}
			if _, err := s.server.db.Exec(`UPDATE guilds SET room_expiry = $1 WHERE id = $2`, TimeAdjusted().Add(time.Hour*24*7), guild.ID); err != nil {
				s.logger.Error("Failed to update guild room expiry", zap.Error(err))
			}
		} else {
			if _, err := s.server.db.Exec(`UPDATE guilds SET room_rp = room_rp + $1 WHERE id = $2`, amount, guild.ID); err != nil {
				s.logger.Error("Failed to update guild room RP", zap.Error(err))
			}
		}
	}
	_, _ = bf.Seek(0, 0)
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
		_ = mail.Send(s, nil)
		if s.server.Registry != nil {
			s.server.Registry.NotifyMailToCharID(pkt.CharID, s, &mail)
		} else {
			// Fallback: find the target session under lock, then notify outside the lock.
			var targetSession *Session
			for _, channel := range s.server.Channels {
				channel.Lock()
				for _, session := range channel.sessions {
					if session.charID == pkt.CharID {
						targetSession = session
						break
					}
				}
				channel.Unlock()
				if targetSession != nil {
					break
				}
			}
			if targetSession != nil {
				SendMailNotification(s, &mail, targetSession)
			}
		}
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
	}
}
