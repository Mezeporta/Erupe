package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

type TreasureHunt struct {
	HuntID      uint32 `db:"id"`
	HostID      uint32 `db:"host_id"`
	Destination uint32 `db:"destination"`
	Level       uint32 `db:"level"`
	Return      uint32 `db:"return"`
	Acquired    bool   `db:"acquired"`
	Claimed     bool   `db:"claimed"`
	Hunters     string `db:"hunters"`
	Treasure    string `db:"treasure"`
	HuntData    []byte `db:"hunt_data"`
}

func handleMsgMhfEnumerateGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateGuildTresure)
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil {
		s.logger.Error("failed to get guild info", zap.Error(err), zap.Uint32("charID", s.charID))
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	bf := byteframe.NewByteFrame()
	hunts := 0
	rows, _ := s.server.db.Queryx("SELECT id, host_id, destination, level, return, acquired, claimed, hunters, treasure, hunt_data FROM guild_hunts WHERE guild_id=$1 AND $2 < return+604800", guild.ID, TimeAdjusted().Unix())
	for rows.Next() {
		hunt := &TreasureHunt{}
		err = rows.StructScan(&hunt)
		// Remove self from other hunter count
		hunt.Hunters = stringsupport.CSVRemove(hunt.Hunters, int(s.charID))
		if err != nil {
			s.logger.Error("failed to scan treasure hunt row", zap.Error(err))
			continue
		}
		if pkt.MaxHunts == 1 {
			if hunt.HostID != s.charID || hunt.Acquired {
				continue
			}
			hunts++
			bf.WriteUint32(hunt.HuntID)
			bf.WriteUint32(hunt.Destination)
			bf.WriteUint32(hunt.Level)
			bf.WriteUint32(uint32(stringsupport.CSVLength(hunt.Hunters)))
			bf.WriteUint32(hunt.Return)
			bf.WriteBool(false)
			bf.WriteBool(false)
			bf.WriteBytes(hunt.HuntData)
			break
		} else if pkt.MaxHunts == 30 && hunt.Acquired && hunt.Level == 2 {
			if hunts == 30 {
				break
			}
			hunts++
			bf.WriteUint32(hunt.HuntID)
			bf.WriteUint32(hunt.Destination)
			bf.WriteUint32(hunt.Level)
			bf.WriteUint32(uint32(stringsupport.CSVLength(hunt.Hunters)))
			bf.WriteUint32(hunt.Return)
			bf.WriteBool(hunt.Claimed)
			bf.WriteBool(stringsupport.CSVContains(hunt.Treasure, int(s.charID)))
			bf.WriteBytes(hunt.HuntData)
		}
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint16(uint16(hunts))
	resp.WriteUint16(uint16(hunts))
	resp.WriteBytes(bf.Data())
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfRegistGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfRegistGuildTresure)
	bf := byteframe.NewByteFrameFromBytes(pkt.Data)
	huntData := byteframe.NewByteFrame()
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil {
		s.logger.Error("failed to get guild info for treasure registration", zap.Error(err), zap.Uint32("charID", s.charID))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	guildCats := getGuildAirouList(s)
	destination := bf.ReadUint32()
	level := bf.ReadUint32()
	huntData.WriteUint32(s.charID)
	huntData.WriteBytes(stringsupport.PaddedString(s.Name, 18, true))
	catsUsed := ""
	for i := 0; i < 5; i++ {
		catID := bf.ReadUint32()
		huntData.WriteUint32(catID)
		if catID > 0 {
			catsUsed = stringsupport.CSVAdd(catsUsed, int(catID))
			for _, cat := range guildCats {
				if cat.CatID == catID {
					huntData.WriteBytes(cat.CatName)
					break
				}
			}
			huntData.WriteBytes(bf.ReadBytes(9))
		}
	}
	_, err = s.server.db.Exec("INSERT INTO guild_hunts (guild_id, host_id, destination, level, return, hunt_data, cats_used) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		guild.ID, s.charID, destination, level, TimeAdjusted().Unix(), huntData.Data(), catsUsed)
	if err != nil {
		s.logger.Error("failed to insert guild hunt", zap.Error(err), zap.Uint32("guildID", guild.ID))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireGuildTresure(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildTresure)
	_, err := s.server.db.Exec("UPDATE guild_hunts SET acquired=true WHERE id=$1", pkt.HuntID)
	if err != nil {
		s.logger.Error("failed to acquire guild treasure", zap.Error(err), zap.Uint32("huntID", pkt.HuntID))
		doAckSimpleFail(s, pkt.AckHandle, nil)
		return
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func treasureHuntUnregister(s *Session) {
	guild, err := GetGuildInfoByCharacterId(s, s.charID)
	if err != nil || guild == nil {
		return
	}
	var huntID int
	var hunters string
	rows, err := s.server.db.Queryx("SELECT id, hunters FROM guild_hunts WHERE guild_id=$1", guild.ID)
	if err != nil {
		return
	}
	for rows.Next() {
		rows.Scan(&huntID, &hunters)
		hunters = stringsupport.CSVRemove(hunters, int(s.charID))
		s.server.db.Exec("UPDATE guild_hunts SET hunters=$1 WHERE id=$2", hunters, huntID)
	}
}

func handleMsgMhfOperateGuildTresureReport(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfOperateGuildTresureReport)
	var csv string
	if pkt.State == 0 { // Report registration
		// Unregister from all other hunts
		treasureHuntUnregister(s)
		if pkt.HuntID != 0 {
			// Register to selected hunt
			err := s.server.db.QueryRow("SELECT hunters FROM guild_hunts WHERE id=$1", pkt.HuntID).Scan(&csv)
			if err != nil {
				s.logger.Error("failed to get hunters for guild hunt", zap.Error(err), zap.Uint32("huntID", pkt.HuntID))
				doAckSimpleFail(s, pkt.AckHandle, nil)
				return
			}
			csv = stringsupport.CSVAdd(csv, int(s.charID))
			_, err = s.server.db.Exec("UPDATE guild_hunts SET hunters=$1 WHERE id=$2", csv, pkt.HuntID)
			if err != nil {
				s.logger.Error("failed to update hunters for guild hunt", zap.Error(err), zap.Uint32("huntID", pkt.HuntID))
				doAckSimpleFail(s, pkt.AckHandle, nil)
				return
			}
		}
	} else if pkt.State == 1 { // Collected by hunter
		s.server.db.Exec("UPDATE guild_hunts SET hunters='', claimed=true WHERE id=$1", pkt.HuntID)
	} else if pkt.State == 2 { // Claim treasure
		err := s.server.db.QueryRow("SELECT treasure FROM guild_hunts WHERE id=$1", pkt.HuntID).Scan(&csv)
		if err != nil {
			s.logger.Error("failed to get treasure for guild hunt", zap.Error(err), zap.Uint32("huntID", pkt.HuntID))
			doAckSimpleFail(s, pkt.AckHandle, nil)
			return
		}
		csv = stringsupport.CSVAdd(csv, int(s.charID))
		_, err = s.server.db.Exec("UPDATE guild_hunts SET treasure=$1 WHERE id=$2", csv, pkt.HuntID)
		if err != nil {
			s.logger.Error("failed to update treasure for guild hunt", zap.Error(err), zap.Uint32("huntID", pkt.HuntID))
			doAckSimpleFail(s, pkt.AckHandle, nil)
			return
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfGetGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetGuildTresureSouvenir)

	doAckBufSucceed(s, pkt.AckHandle, make([]byte, 6))
}

func handleMsgMhfAcquireGuildTresureSouvenir(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireGuildTresureSouvenir)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
