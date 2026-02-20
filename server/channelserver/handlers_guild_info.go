package channelserver

import (
	"sort"
	"strings"

	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"github.com/jmoiron/sqlx"
)

// Guild sentinel and cost constants
const (
	guildNotJoinedSentinel = uint32(0xFFFFFFFF)
	guildRoomMaxRP         = uint32(55000)
)

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
		characterJoinedAt := guildNotJoinedSentinel

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
		bf.WriteUint16(guild.Rank(s.server.erupeConfig.RealClientMode))
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
			if guild.Rank(s.server.erupeConfig.RealClientMode) >= uint16(j[0]) {
				limit = j[1]
			}
		}
		if limit > 100 {
			limit = 100
		}
		bf.WriteUint8(limit)

		bf.WriteUint32(guildRoomMaxRP)
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
				bf.WriteUint16(alliance.ParentGuild.Rank(s.server.erupeConfig.RealClientMode))
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
					bf.WriteUint16(alliance.SubGuild1.Rank(s.server.erupeConfig.RealClientMode))
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
					bf.WriteUint16(alliance.SubGuild2.Rank(s.server.erupeConfig.RealClientMode))
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
			searchName, _ := stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())
			for _, guild := range tempGuilds {
				if strings.Contains(guild.Name, searchName) {
					guilds = append(guilds, guild)
				}
			}
		case mhfpacket.ENUMERATE_GUILD_TYPE_LEADER_NAME:
			searchName, _ := stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())
			for _, guild := range tempGuilds {
				if strings.Contains(guild.LeaderName, searchName) {
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
			searchName, _ := stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())
			for _, alliance := range tempAlliances {
				if strings.Contains(alliance.Name, searchName) {
					alliances = append(alliances, alliance)
				}
			}
		case mhfpacket.ENUMERATE_ALLIANCE_TYPE_LEADER_NAME:
			searchName, _ := stringsupport.SJISToUTF8(pkt.Data2.ReadNullTerminatedBytes())
			for _, alliance := range tempAlliances {
				if strings.Contains(alliance.ParentGuild.LeaderName, searchName) {
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
			bf.WriteUint16(guild.Rank(s.server.erupeConfig.RealClientMode))
			bf.WriteUint32(uint32(guild.CreatedAt.Unix()))
			ps.Uint8(bf, guild.Name, true)
			ps.Uint8(bf, guild.LeaderName, true)
			bf.WriteUint8(0x01) // Unk
			bf.WriteBool(!guild.Recruiting)
		}
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
