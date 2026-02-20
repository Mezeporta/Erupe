package channelserver

import (
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/deltacomp"
	"erupe-ce/server/channelserver/compression/nullcomp"

	"go.uber.org/zap"
)

func handleMsgMhfSavedata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavedata)
	characterSaveData, err := GetCharacterSaveData(s, s.charID)
	if err != nil {
		s.logger.Error("failed to retrieve character save data from db", zap.Error(err), zap.Uint32("charID", s.charID))
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	// Snapshot current house tier before applying the update so we can
	// restore it if the incoming data is corrupted (issue #92).
	prevHouseTier := make([]byte, len(characterSaveData.HouseTier))
	copy(prevHouseTier, characterSaveData.HouseTier)

	// Var to hold the decompressed savedata for updating the launcher response fields.
	if pkt.SaveType == 1 {
		// Diff-based update.
		// diffs themselves are also potentially compressed
		diff, err := nullcomp.Decompress(pkt.RawDataPayload)
		if err != nil {
			s.logger.Error("Failed to decompress diff", zap.Error(err))
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		// Perform diff.
		s.logger.Info("Diffing...")
		characterSaveData.decompSave = deltacomp.ApplyDataDiff(diff, characterSaveData.decompSave)
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "savedata")
		// Regular blob update.
		saveData, err := nullcomp.Decompress(pkt.RawDataPayload)
		if err != nil {
			s.logger.Error("Failed to decompress savedata from packet", zap.Error(err))
			doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		if s.server.erupeConfig.SaveDumps.RawEnabled {
			dumpSaveData(s, saveData, "raw-savedata")
		}
		s.logger.Info("Updating save with blob")
		characterSaveData.decompSave = saveData
	}
	characterSaveData.updateStructWithSaveData()

	// Mitigate house theme corruption (issue #92): the game client
	// sometimes sends house_tier as -1 (all 0xFF bytes), which causes
	// the house theme to vanish on next login. If the new value looks
	// corrupted, restore the previous value in both the struct and the
	// decompressed blob so Save() persists consistent data.
	if len(prevHouseTier) > 0 && characterSaveData.isHouseTierCorrupted() {
		s.logger.Warn("Detected corrupted house_tier in save data, restoring previous value",
			zap.Binary("corrupted", characterSaveData.HouseTier),
			zap.Binary("restored", prevHouseTier),
			zap.Uint32("charID", s.charID),
		)
		characterSaveData.restoreHouseTier(prevHouseTier)
	}

	s.playtime = characterSaveData.Playtime
	s.playtimeTime = time.Now()

	// Bypass name-checker if new
	if characterSaveData.IsNewCharacter {
		s.Name = characterSaveData.Name
	}

	// Force name to match session to prevent corruption detection false positives
	// This handles SJIS/UTF-8 encoding differences and ensures saves succeed across all game versions
	if characterSaveData.Name != s.Name && !characterSaveData.IsNewCharacter {
		s.logger.Info("Correcting name mismatch in savedata", zap.String("savedata_name", characterSaveData.Name), zap.String("session_name", s.Name))
		characterSaveData.Name = s.Name
		characterSaveData.updateSaveDataWithStruct()
	}

	if characterSaveData.Name == s.Name || s.server.erupeConfig.RealClientMode <= _config.S10 {
		characterSaveData.Save(s)
		s.logger.Info("Wrote recompressed savedata back to DB.")
	} else {
		_ = s.rawConn.Close()
		s.logger.Warn("Save cancelled due to corruption.")
		if s.server.erupeConfig.DeleteOnSaveCorruption {
			if err := s.server.charRepo.SetDeleted(s.charID); err != nil {
				s.logger.Error("Failed to mark character as deleted", zap.Error(err))
			}
		}
		return
	}
	if err := s.server.charRepo.SaveString(s.charID, "name", characterSaveData.Name); err != nil {
		s.logger.Error("Failed to update character name in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func grpToGR(n int) uint16 {
	var gr int
	a := []int{208750, 593400, 993400, 1400900, 2315900, 3340900, 4505900, 5850900, 7415900, 9230900, 11345900, 100000000}
	b := []int{7850, 8000, 8150, 9150, 10250, 11650, 13450, 15650, 18150, 21150, 23950}
	c := []int{51, 100, 150, 200, 300, 400, 500, 600, 700, 800, 900}

	for i := 0; i < len(a); i++ {
		if n < a[i] {
			if i == 0 {
				for {
					n -= 500
					if n <= 500 {
						if n < 0 {
							i--
						}
						break
					} else {
						i++
						for j := 0; j < i; j++ {
							n -= 150
						}
					}
				}
				gr = i + 2
			} else {
				n -= a[i-1]
				gr = c[i-1]
				gr += n / b[i-1]
			}
			break
		}
	}
	return uint16(gr)
}

func dumpSaveData(s *Session, data []byte, suffix string) {
	if !s.server.erupeConfig.SaveDumps.Enabled {
		return
	} else {
		dir := filepath.Join(s.server.erupeConfig.SaveDumps.OutputDir, fmt.Sprintf("%d", s.charID))
		path := filepath.Join(s.server.erupeConfig.SaveDumps.OutputDir, fmt.Sprintf("%d", s.charID), fmt.Sprintf("%d_%s.bin", s.charID, suffix))
		_, err := os.Stat(dir)
		if err != nil {
			if os.IsNotExist(err) {
				err = os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					s.logger.Error("Error dumping savedata, could not create folder")
					return
				}
			} else {
				s.logger.Error("Error dumping savedata")
				return
			}
		}
		err = os.WriteFile(path, data, 0644)
		if err != nil {
			s.logger.Error("Error dumping savedata, could not write file", zap.Error(err))
		}
	}
}

func handleMsgMhfLoaddata(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoaddata)
	if _, err := os.Stat(filepath.Join(s.server.erupeConfig.BinPath, "save_override.bin")); err == nil {
		data, _ := os.ReadFile(filepath.Join(s.server.erupeConfig.BinPath, "save_override.bin"))
		doAckBufSucceed(s, pkt.AckHandle, data)
		return
	}

	data, err := s.server.charRepo.LoadColumn(s.charID, "savedata")
	if err != nil || len(data) == 0 {
		s.logger.Warn(fmt.Sprintf("Failed to load savedata (CID: %d)", s.charID), zap.Error(err))
		_ = s.rawConn.Close() // Terminate the connection
		return
	}
	doAckBufSucceed(s, pkt.AckHandle, data)

	decompSaveData, err := nullcomp.Decompress(data)
	if err != nil {
		s.logger.Error("Failed to decompress savedata", zap.Error(err))
	}
	bf := byteframe.NewByteFrameFromBytes(decompSaveData)
	_, _ = bf.Seek(88, io.SeekStart)
	name := bf.ReadNullTerminatedBytes()
	s.server.userBinaryPartsLock.Lock()
	s.server.userBinaryParts[userBinaryPartID{charID: s.charID, index: 1}] = append(name, []byte{0x00}...)
	s.server.userBinaryPartsLock.Unlock()
	s.Name, _ = stringsupport.SJISToUTF8(name)
}

func handleMsgMhfSaveScenarioData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveScenarioData)
	saveCharacterData(s, pkt.AckHandle, "scenariodata", pkt.RawDataPayload, 65536)
}

func handleMsgMhfLoadScenarioData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadScenarioData)
	loadCharacterData(s, pkt.AckHandle, "scenariodata", make([]byte, 10))
}

func handleMsgSysAuthData(s *Session, p mhfpacket.MHFPacket) {}
