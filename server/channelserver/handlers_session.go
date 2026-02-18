package channelserver

import (
	"crypto/rand"
	"encoding/binary"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcourse"
	"erupe-ce/common/mhfmon"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"io"
	"net"
	"strings"
	"time"

	"go.uber.org/zap"
)

func handleMsgHead(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysExtendThreshold(s *Session, p mhfpacket.MHFPacket) {
	// No data aside from header, no resp required.
}

func handleMsgSysEnd(s *Session, p mhfpacket.MHFPacket) {
	// No data aside from header, no resp required.
}

func handleMsgSysNop(s *Session, p mhfpacket.MHFPacket) {
	// No data aside from header, no resp required.
}

func handleMsgSysAck(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysTerminalLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysTerminalLog)
	for i := range pkt.Entries {
		s.server.logger.Info("SysTerminalLog",
			zap.Uint8("Type1", pkt.Entries[i].Type1),
			zap.Uint8("Type2", pkt.Entries[i].Type2),
			zap.Int16("Unk0", pkt.Entries[i].Unk0),
			zap.Int32("Unk1", pkt.Entries[i].Unk1),
			zap.Int32("Unk2", pkt.Entries[i].Unk2),
			zap.Int32("Unk3", pkt.Entries[i].Unk3),
			zap.Int32s("Unk4", pkt.Entries[i].Unk4),
		)
	}
	resp := byteframe.NewByteFrame()
	resp.WriteUint32(pkt.LogID + 1) // LogID to use for requests after this.
	doAckSimpleSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgSysLogin(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLogin)

	if !s.server.erupeConfig.DebugOptions.DisableTokenCheck {
		var token string
		err := s.server.db.QueryRow("SELECT token FROM sign_sessions ss INNER JOIN public.users u on ss.user_id = u.id WHERE token=$1 AND ss.id=$2 AND u.id=(SELECT c.user_id FROM characters c WHERE c.id=$3)", pkt.LoginTokenString, pkt.LoginTokenNumber, pkt.CharID0).Scan(&token)
		if err != nil {
			_ = s.rawConn.Close()
			s.logger.Warn(fmt.Sprintf("Invalid login token, offending CID: (%d)", pkt.CharID0))
			return
		}
	}

	s.Lock()
	s.charID = pkt.CharID0
	s.token = pkt.LoginTokenString
	s.Unlock()

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(uint32(TimeAdjusted().Unix())) // Unix timestamp

	_, err := s.server.db.Exec("UPDATE servers SET current_players=$1 WHERE server_id=$2", len(s.server.sessions), s.server.ID)
	if err != nil {
		panic(err)
	}

	_, err = s.server.db.Exec("UPDATE sign_sessions SET server_id=$1, char_id=$2 WHERE token=$3", s.server.ID, s.charID, s.token)
	if err != nil {
		panic(err)
	}

	_, err = s.server.db.Exec("UPDATE characters SET last_login=$1 WHERE id=$2", TimeAdjusted().Unix(), s.charID)
	if err != nil {
		panic(err)
	}

	_, err = s.server.db.Exec("UPDATE users u SET last_character=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$1)", s.charID)
	if err != nil {
		panic(err)
	}

	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())

	updateRights(s)

	s.server.BroadcastMHF(&mhfpacket.MsgSysInsertUser{CharID: s.charID}, s)
}

func handleMsgSysLogout(s *Session, p mhfpacket.MHFPacket) {
	logoutPlayer(s)
}

// saveAllCharacterData saves all character data to the database with proper error handling.
// This function ensures data persistence even if the client disconnects unexpectedly.
// It handles:
// - Main savedata blob (compressed)
// - User binary data (house, gallery, etc.)
// - Plate data (transmog appearance, storage, equipment sets)
// - Playtime updates
// - RP updates
// - Name corruption prevention
func saveAllCharacterData(s *Session, rpToAdd int) error {
	saveStart := time.Now()

	// Get current savedata from database
	characterSaveData, err := GetCharacterSaveData(s, s.charID)
	if err != nil {
		s.logger.Error("Failed to retrieve character save data",
			zap.Error(err),
			zap.Uint32("charID", s.charID),
			zap.String("name", s.Name),
		)
		return err
	}

	if characterSaveData == nil {
		s.logger.Warn("Character save data is nil, skipping save",
			zap.Uint32("charID", s.charID),
			zap.String("name", s.Name),
		)
		return nil
	}

	// Force name to match to prevent corruption detection issues
	// This handles SJIS/UTF-8 encoding differences across game versions
	if characterSaveData.Name != s.Name {
		s.logger.Debug("Correcting name mismatch before save",
			zap.String("savedata_name", characterSaveData.Name),
			zap.String("session_name", s.Name),
			zap.Uint32("charID", s.charID),
		)
		characterSaveData.Name = s.Name
		characterSaveData.updateSaveDataWithStruct()
	}

	// Update playtime from session
	if !s.playtimeTime.IsZero() {
		sessionPlaytime := uint32(time.Since(s.playtimeTime).Seconds())
		s.playtime += sessionPlaytime
		s.logger.Debug("Updated playtime",
			zap.Uint32("session_playtime_seconds", sessionPlaytime),
			zap.Uint32("total_playtime", s.playtime),
			zap.Uint32("charID", s.charID),
		)
	}
	characterSaveData.Playtime = s.playtime

	// Update RP if any gained during session
	if rpToAdd > 0 {
		characterSaveData.RP += uint16(rpToAdd)
		if characterSaveData.RP >= s.server.erupeConfig.GameplayOptions.MaximumRP {
			characterSaveData.RP = s.server.erupeConfig.GameplayOptions.MaximumRP
			s.logger.Debug("RP capped at maximum",
				zap.Uint16("max_rp", s.server.erupeConfig.GameplayOptions.MaximumRP),
				zap.Uint32("charID", s.charID),
			)
		}
		s.logger.Debug("Added RP",
			zap.Int("rp_gained", rpToAdd),
			zap.Uint16("new_rp", characterSaveData.RP),
			zap.Uint32("charID", s.charID),
		)
	}

	// Save to database (main savedata + user_binary)
	characterSaveData.Save(s)

	// Save auxiliary data types
	// Note: Plate data saves immediately when client sends save packets,
	// so this is primarily a safety net for monitoring and consistency
	if err := savePlateDataToDatabase(s); err != nil {
		s.logger.Error("Failed to save plate data during logout",
			zap.Error(err),
			zap.Uint32("charID", s.charID),
		)
		// Don't return error - continue with logout even if plate save fails
	}

	saveDuration := time.Since(saveStart)
	s.logger.Info("Saved character data successfully",
		zap.Uint32("charID", s.charID),
		zap.String("name", s.Name),
		zap.Duration("duration", saveDuration),
		zap.Int("rp_added", rpToAdd),
		zap.Uint32("playtime", s.playtime),
	)

	return nil
}

func logoutPlayer(s *Session) {
	logoutStart := time.Now()

	// Log logout initiation with session details
	sessionDuration := time.Duration(0)
	if s.sessionStart > 0 {
		sessionDuration = time.Since(time.Unix(s.sessionStart, 0))
	}

	s.logger.Info("Player logout initiated",
		zap.Uint32("charID", s.charID),
		zap.String("name", s.Name),
		zap.Duration("session_duration", sessionDuration),
	)

	// Calculate session metrics FIRST (before cleanup)
	var timePlayed int
	var sessionTime int
	var rpGained int

	if s.charID != 0 {
		_ = s.server.db.QueryRow("SELECT time_played FROM characters WHERE id = $1", s.charID).Scan(&timePlayed)
		sessionTime = int(TimeAdjusted().Unix()) - int(s.sessionStart)
		timePlayed += sessionTime

		if mhfcourse.CourseExists(30, s.courses) {
			rpGained = timePlayed / 900
			timePlayed = timePlayed % 900
			if _, err := s.server.db.Exec("UPDATE characters SET cafe_time=cafe_time+$1 WHERE id=$2", sessionTime, s.charID); err != nil {
				s.logger.Error("Failed to update cafe time", zap.Error(err))
			}
		} else {
			rpGained = timePlayed / 1800
			timePlayed = timePlayed % 1800
		}

		s.logger.Debug("Session metrics calculated",
			zap.Uint32("charID", s.charID),
			zap.Int("session_time_seconds", sessionTime),
			zap.Int("rp_gained", rpGained),
			zap.Int("time_played_remainder", timePlayed),
		)

		// Save all character data ONCE with all updates
		// This is the safety net that ensures data persistence even if client
		// didn't send save packets before disconnecting
		if err := saveAllCharacterData(s, rpGained); err != nil {
			s.logger.Error("Failed to save character data during logout",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
				zap.String("name", s.Name),
			)
			// Continue with logout even if save fails
		}

		// Update time_played and guild treasure hunt
		if _, err := s.server.db.Exec("UPDATE characters SET time_played = $1 WHERE id = $2", timePlayed, s.charID); err != nil {
			s.logger.Error("Failed to update time played", zap.Error(err))
		}
		if _, err := s.server.db.Exec(`UPDATE guild_characters SET treasure_hunt=NULL WHERE character_id=$1`, s.charID); err != nil {
			s.logger.Error("Failed to clear treasure hunt", zap.Error(err))
		}
	}

	// NOW do cleanup (after save is complete)
	s.server.Lock()
	delete(s.server.sessions, s.rawConn)
	_ = s.rawConn.Close()
	s.server.Unlock()

	// Stage cleanup
	for _, stage := range s.server.stages {
		// Tell sessions registered to disconnecting players quest to unregister
		if stage.host != nil && stage.host.charID == s.charID {
			for _, sess := range s.server.sessions {
				for rSlot := range stage.reservedClientSlots {
					if sess.charID == rSlot && sess.stage != nil && sess.stage.id[3:5] != "Qs" {
						sess.QueueSendMHFNonBlocking(&mhfpacket.MsgSysStageDestruct{})
					}
				}
			}
		}
		for session := range stage.clients {
			if session.charID == s.charID {
				delete(stage.clients, session)
			}
		}
	}

	// Update sign sessions and server player count
	_, err := s.server.db.Exec("UPDATE sign_sessions SET server_id=NULL, char_id=NULL WHERE token=$1", s.token)
	if err != nil {
		panic(err)
	}

	_, err = s.server.db.Exec("UPDATE servers SET current_players=$1 WHERE server_id=$2", len(s.server.sessions), s.server.ID)
	if err != nil {
		panic(err)
	}

	if s.stage == nil {
		logoutDuration := time.Since(logoutStart)
		s.logger.Info("Player logout completed",
			zap.Uint32("charID", s.charID),
			zap.String("name", s.Name),
			zap.Duration("logout_duration", logoutDuration),
		)
		return
	}

	// Broadcast user deletion and final cleanup
	s.server.BroadcastMHF(&mhfpacket.MsgSysDeleteUser{
		CharID: s.charID,
	}, s)

	s.server.Lock()
	for _, stage := range s.server.stages {
		delete(stage.reservedClientSlots, s.charID)
	}
	s.server.Unlock()

	removeSessionFromSemaphore(s)
	removeSessionFromStage(s)

	logoutDuration := time.Since(logoutStart)
	s.logger.Info("Player logout completed",
		zap.Uint32("charID", s.charID),
		zap.String("name", s.Name),
		zap.Duration("logout_duration", logoutDuration),
		zap.Int("rp_gained", rpGained),
	)
}

func handleMsgSysSetStatus(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysPing(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysPing)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgSysTime(s *Session, p mhfpacket.MHFPacket) {
	resp := &mhfpacket.MsgSysTime{
		GetRemoteTime: false,
		Timestamp:     uint32(TimeAdjusted().Unix()), // JP timezone
	}
	s.QueueSendMHF(resp)
	s.notifyRavi()
}

func handleMsgSysIssueLogkey(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysIssueLogkey)

	// Make a random log key for this session.
	logKey := make([]byte, 16)
	_, err := rand.Read(logKey)
	if err != nil {
		panic(err)
	}

	// TODO(Andoryuuta): In the offical client, the log key index is off by one,
	// cutting off the last byte in _most uses_. Find and document these accordingly.
	s.Lock()
	s.logKey = logKey
	s.Unlock()

	// Issue it.
	resp := byteframe.NewByteFrame()
	resp.WriteBytes(logKey)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgSysRecordLog(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysRecordLog)
	if _config.ErupeConfig.RealClientMode == _config.ZZ {
		bf := byteframe.NewByteFrameFromBytes(pkt.Data)
		_, _ = bf.Seek(32, 0)
		var val uint8
		for i := 0; i < 176; i++ {
			val = bf.ReadUint8()
			if val > 0 && mhfmon.Monsters[i].Large {
				if _, err := s.server.db.Exec(`INSERT INTO kill_logs (character_id, monster, quantity, timestamp) VALUES ($1, $2, $3, $4)`, s.charID, i, val, TimeAdjusted()); err != nil {
					s.logger.Error("Failed to insert kill log", zap.Error(err))
				}
			}
		}
	}
	// remove a client returning to town from reserved slots to make sure the stage is hidden from board
	delete(s.stage.reservedClientSlots, s.charID)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgSysEcho(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysLockGlobalSema(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysLockGlobalSema)
	var sgid string
	for _, channel := range s.server.Channels {
		for id := range channel.stages {
			if strings.HasSuffix(id, pkt.UserIDString) {
				sgid = channel.GlobalID
			}
		}
	}
	bf := byteframe.NewByteFrame()
	if len(sgid) > 0 && sgid != s.server.GlobalID {
		bf.WriteUint8(0)
		bf.WriteUint8(0)
		ps.Uint16(bf, sgid, false)
	} else {
		bf.WriteUint8(2)
		bf.WriteUint8(0)
		ps.Uint16(bf, pkt.ServerChannelIDString, false)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgSysUnlockGlobalSema(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysUnlockGlobalSema)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgSysUpdateRight(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAuthQuery(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysAuthTerminal(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysRightsReload(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysRightsReload)
	updateRights(s)
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfTransitMessage(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfTransitMessage)

	local := strings.Split(s.rawConn.RemoteAddr().String(), ":")[0] == "127.0.0.1"

	var maxResults, port, count uint16
	var cid uint32
	var term, ip string
	bf := byteframe.NewByteFrameFromBytes(pkt.MessageData)
	switch pkt.SearchType {
	case 1:
		maxResults = 1
		cid = bf.ReadUint32()
	case 2:
		bf.ReadUint16() // term length
		maxResults = bf.ReadUint16()
		bf.ReadUint8() // Unk
		term = stringsupport.SJISToUTF8(bf.ReadNullTerminatedBytes())
	case 3:
		_ip := bf.ReadBytes(4)
		ip = fmt.Sprintf("%d.%d.%d.%d", _ip[3], _ip[2], _ip[1], _ip[0])
		port = bf.ReadUint16()
		bf.ReadUint16() // term length
		maxResults = bf.ReadUint16()
		bf.ReadUint8()
		term = string(bf.ReadNullTerminatedBytes())
	}

	resp := byteframe.NewByteFrame()
	resp.WriteUint16(0)
	switch pkt.SearchType {
	case 1, 2, 3: // usersearchidx, usersearchname, lobbysearchname
		for _, c := range s.server.Channels {
			for _, session := range c.sessions {
				if count == maxResults {
					break
				}
				if pkt.SearchType == 1 && session.charID != cid {
					continue
				}
				if pkt.SearchType == 2 && !strings.Contains(session.Name, term) {
					continue
				}
				if pkt.SearchType == 3 && session.server.IP != ip && session.server.Port != port && session.stage.id != term {
					continue
				}
				count++
				sessionName := stringsupport.UTF8ToSJIS(session.Name)
				sessionStage := stringsupport.UTF8ToSJIS(session.stage.id)
				if !local {
					resp.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(c.IP).To4()))
				} else {
					resp.WriteUint32(0x0100007F)
				}
				resp.WriteUint16(c.Port)
				resp.WriteUint32(session.charID)
				resp.WriteUint8(uint8(len(sessionStage) + 1))
				resp.WriteUint8(uint8(len(sessionName) + 1))
				resp.WriteUint16(uint16(len(c.userBinaryParts[userBinaryPartID{charID: session.charID, index: 3}])))

				// TODO: This case might be <=G2
				if _config.ErupeConfig.RealClientMode <= _config.G1 {
					resp.WriteBytes(make([]byte, 8))
				} else {
					resp.WriteBytes(make([]byte, 40))
				}
				resp.WriteBytes(make([]byte, 8))

				resp.WriteNullTerminatedBytes(sessionStage)
				resp.WriteNullTerminatedBytes(sessionName)
				resp.WriteBytes(c.userBinaryParts[userBinaryPartID{session.charID, 3}])
			}
		}
	case 4: // lobbysearch
		type FindPartyParams struct {
			StagePrefix     string
			RankRestriction int16
			Targets         []int16
			Unk0            []int16
			Unk1            []int16
			QuestID         []int16
		}
		findPartyParams := FindPartyParams{
			StagePrefix: "sl2Ls210",
		}
		numParams := bf.ReadUint8()
		maxResults = bf.ReadUint16()
		for i := uint8(0); i < numParams; i++ {
			switch bf.ReadUint8() {
			case 0:
				values := bf.ReadUint8()
				for i := uint8(0); i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.RankRestriction = bf.ReadInt16()
					} else {
						findPartyParams.RankRestriction = int16(bf.ReadInt8())
					}
				}
			case 1:
				values := bf.ReadUint8()
				for i := uint8(0); i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.Targets = append(findPartyParams.Targets, bf.ReadInt16())
					} else {
						findPartyParams.Targets = append(findPartyParams.Targets, int16(bf.ReadInt8()))
					}
				}
			case 2:
				values := bf.ReadUint8()
				for i := uint8(0); i < values; i++ {
					var value int16
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						value = bf.ReadInt16()
					} else {
						value = int16(bf.ReadInt8())
					}
					switch value {
					case 0: // Public Bar
						findPartyParams.StagePrefix = "sl2Ls210"
					case 1: // Tokotoko Partnya
						findPartyParams.StagePrefix = "sl2Ls463"
					case 2: // Hunting Prowess Match
						findPartyParams.StagePrefix = "sl2Ls286"
					case 3: // Volpakkun Together
						findPartyParams.StagePrefix = "sl2Ls465"
					case 5: // Quick Party
						// Unk
					}
				}
			case 3: // Unknown
				values := bf.ReadUint8()
				for i := uint8(0); i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.Unk0 = append(findPartyParams.Unk0, bf.ReadInt16())
					} else {
						findPartyParams.Unk0 = append(findPartyParams.Unk0, int16(bf.ReadInt8()))
					}
				}
			case 4: // Looking for n or already have n
				values := bf.ReadUint8()
				for i := uint8(0); i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.Unk1 = append(findPartyParams.Unk1, bf.ReadInt16())
					} else {
						findPartyParams.Unk1 = append(findPartyParams.Unk1, int16(bf.ReadInt8()))
					}
				}
			case 5:
				values := bf.ReadUint8()
				for i := uint8(0); i < values; i++ {
					if _config.ErupeConfig.RealClientMode >= _config.Z1 {
						findPartyParams.QuestID = append(findPartyParams.QuestID, bf.ReadInt16())
					} else {
						findPartyParams.QuestID = append(findPartyParams.QuestID, int16(bf.ReadInt8()))
					}
				}
			}
		}
		for _, c := range s.server.Channels {
			for _, stage := range c.stages {
				if count == maxResults {
					break
				}
				if strings.HasPrefix(stage.id, findPartyParams.StagePrefix) {
					sb3 := byteframe.NewByteFrameFromBytes(stage.rawBinaryData[stageBinaryKey{1, 3}])
					_, _ = sb3.Seek(4, 0)

					stageDataParams := 7
					if _config.ErupeConfig.RealClientMode <= _config.G10 {
						stageDataParams = 4
					} else if _config.ErupeConfig.RealClientMode <= _config.Z1 {
						stageDataParams = 6
					}

					var stageData []int16
					for i := 0; i < stageDataParams; i++ {
						if _config.ErupeConfig.RealClientMode >= _config.Z1 {
							stageData = append(stageData, sb3.ReadInt16())
						} else {
							stageData = append(stageData, int16(sb3.ReadInt8()))
						}
					}

					if findPartyParams.RankRestriction >= 0 {
						if stageData[0] > findPartyParams.RankRestriction {
							continue
						}
					}

					var hasTarget bool
					if len(findPartyParams.Targets) > 0 {
						for _, target := range findPartyParams.Targets {
							if target == stageData[1] {
								hasTarget = true
								break
							}
						}
						if !hasTarget {
							continue
						}
					}

					count++
					if !local {
						resp.WriteUint32(binary.LittleEndian.Uint32(net.ParseIP(c.IP).To4()))
					} else {
						resp.WriteUint32(0x0100007F)
					}
					resp.WriteUint16(c.Port)

					resp.WriteUint16(0) // Static?
					resp.WriteUint16(0) // Unk, [0 1 2]
					resp.WriteUint16(uint16(len(stage.clients) + len(stage.reservedClientSlots)))
					resp.WriteUint16(stage.maxPlayers)
					// TODO: Retail returned the number of clients in quests, not workshop/my series
					resp.WriteUint16(uint16(len(stage.reservedClientSlots)))

					resp.WriteUint8(0) // Static?
					resp.WriteUint8(uint8(stage.maxPlayers))
					resp.WriteUint8(1) // Static?
					resp.WriteUint8(uint8(len(stage.id) + 1))
					resp.WriteUint8(uint8(len(stage.rawBinaryData[stageBinaryKey{1, 0}])))
					resp.WriteUint8(uint8(len(stage.rawBinaryData[stageBinaryKey{1, 1}])))

					for i := range stageData {
						if _config.ErupeConfig.RealClientMode >= _config.Z1 {
							resp.WriteInt16(stageData[i])
						} else {
							resp.WriteInt8(int8(stageData[i]))
						}
					}
					resp.WriteUint8(0) // Unk
					resp.WriteUint8(0) // Unk

					resp.WriteNullTerminatedBytes([]byte(stage.id))
					resp.WriteBytes(stage.rawBinaryData[stageBinaryKey{1, 0}])
					resp.WriteBytes(stage.rawBinaryData[stageBinaryKey{1, 1}])
				}
			}
		}
	}
	_, _ = resp.Seek(0, io.SeekStart)
	resp.WriteUint16(count)
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgCaExchangeItem(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfServerCommand(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfAnnounce(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAnnounce)
	s.server.BroadcastRaviente(pkt.IPAddress, pkt.Port, pkt.StageID, pkt.Data.ReadUint8())
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfSetLoginwindow(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysTransBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysCollectBinary(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysGetState(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysSerialize(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysEnumlobby(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysEnumuser(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgSysInfokyserver(s *Session, p mhfpacket.MHFPacket) {}

func handleMsgMhfGetCaUniqueID(s *Session, p mhfpacket.MHFPacket) {}
