package channelserver

import (
	"erupe-ce/common/mhfmon"
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
		return
	}
	// Var to hold the decompressed savedata for updating the launcher response fields.
	if pkt.SaveType == 1 {
		// Diff-based update.
		// diffs themselves are also potentially compressed
		diff, err := nullcomp.Decompress(pkt.RawDataPayload)
		if err != nil {
			s.logger.Error("Failed to decompress diff", zap.Error(err))
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
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
			doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		if s.server.erupeConfig.SaveDumps.RawEnabled {
			dumpSaveData(s, saveData, "raw-savedata")
		}
		s.logger.Info("Updating save with blob")
		characterSaveData.decompSave = saveData
	}
	characterSaveData.updateStructWithSaveData()

	// Bypass name-checker if new
	if characterSaveData.IsNewCharacter == true {
		s.Name = characterSaveData.Name
	}

	if characterSaveData.Name == s.Name || _config.ErupeConfig.RealClientMode <= _config.S10 {
		characterSaveData.Save(s)
		s.logger.Info("Wrote recompressed savedata back to DB.")
	} else {
		s.rawConn.Close()
		s.logger.Warn("Save cancelled due to corruption.")
		if s.server.erupeConfig.DeleteOnSaveCorruption {
			s.server.db.Exec("UPDATE characters SET deleted=true WHERE id=$1", s.charID)
		}
		return
	}
	_, err = s.server.db.Exec("UPDATE characters SET name=$1 WHERE id=$2", characterSaveData.Name, s.charID)
	if err != nil {
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

	var data []byte
	err := s.server.db.QueryRow("SELECT savedata FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil || len(data) == 0 {
		s.logger.Warn(fmt.Sprintf("Failed to load savedata (CID: %d)", s.charID), zap.Error(err))
		s.rawConn.Close() // Terminate the connection
		return
	}
	doAckBufSucceed(s, pkt.AckHandle, data)

	decompSaveData, err := nullcomp.Decompress(data)
	if err != nil {
		s.logger.Error("Failed to decompress savedata", zap.Error(err))
	}
	bf := byteframe.NewByteFrameFromBytes(decompSaveData)
	bf.Seek(88, io.SeekStart)
	name := bf.ReadNullTerminatedBytes()
	s.server.userBinaryPartsLock.Lock()
	s.server.userBinaryParts[userBinaryPartID{charID: s.charID, index: 1}] = append(name, []byte{0x00}...)
	s.server.userBinaryPartsLock.Unlock()
	s.Name = stringsupport.SJISToUTF8(name)
}

func handleMsgMhfSaveScenarioData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveScenarioData)
	dumpSaveData(s, pkt.RawDataPayload, "scenario")
	_, err := s.server.db.Exec("UPDATE characters SET scenariodata = $1 WHERE id = $2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Error("Failed to update scenario data in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfLoadScenarioData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadScenarioData)
	var scenarioData []byte
	bf := byteframe.NewByteFrame()
	err := s.server.db.QueryRow("SELECT scenariodata FROM characters WHERE id = $1", s.charID).Scan(&scenarioData)
	if err != nil || len(scenarioData) < 10 {
		s.logger.Error("Failed to load scenariodata", zap.Error(err))
		bf.WriteBytes(make([]byte, 10))
	} else {
		bf.WriteBytes(scenarioData)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

type PaperMissionTimetable struct {
	Start time.Time
	End   time.Time
}

type PaperMissionData struct {
	Unk0            uint8
	Unk1            uint8
	Target          int16
	Reward1ID       uint16
	Reward1Quantity uint8
	Reward2ID       uint16
	Reward2Quantity uint8
}

type PaperMission struct {
	Timetables []PaperMissionTimetable
	Data       []PaperMissionData
}

type PaperData struct {
	ID      uint16
	Ward    int16
	Option1 int16
	Option2 int16
	Option3 int16
	Option4 int16
	Option5 int16
}

type PaperGift struct {
	ItemID uint16
	Unk1   uint8
	Unk2   uint8
	Chance uint16
}

func handleMsgMhfGetPaperData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfGetPaperData)
	var data []*byteframe.ByteFrame
	var paperData []PaperData
	var paperMissions PaperMission

	// pkt.Type
	// if pkt.Type 3 then Unk2==0  PaperMissionData
	// if pkt.Type 0 then unk2 4, 5 or 6  PaperData
	// if pkt.Type 2 then unk2 6001 6011  PaperGiftData
	// is pkt.Unk2 a index?

	switch pkt.ID {
	case 0:
		//PaperMissionData Target
		// 1: Total Floors
		// 2: TRP Acquired
		// 3: Treasure Chests
		// 4: Old Tresure Chests
		// 5: Defeat Large Monster
		// 6: Dist 1 Dure Slays
		// 7: Dist 2 Dure Slays
		// 8: Dist 3 Dure Slays
		// 9: Dist 4 Dure Slays

		paperMissions = PaperMission{
			[]PaperMissionTimetable{{TimeMidnight(), TimeMidnight().Add(24 * time.Hour)}},
			[]PaperMissionData{{1, 1, 50, 7, 10, 8, 11},
				{1, 2, 100, 7, 12, 8, 13},
				{1, 3, 150, 7, 14, 8, 15},
				{1, 4, 200, 7, 16, 8, 17},
				{1, 5, 250, 7, 18, 8, 19},
				{1, 6, 300, 7, 21, 8, 21}},
		}

	case 4:
		//Triggers on Tower Menu Load and on Tower Quest Load
		paperData = []PaperData{

			//Seen Monsters (id,on off, 0, 0, 0, 0, 0)
			//Value is based on 2001 for monsters
			{1011, 1, 0, 0, 0, 0, 0},
			{1011, 2, 0, 0, 0, 0, 0},
			//Seen Items (id,on off, 0, 0, 0, 0, 0)
			//Value is based in 6001 for items
			{1012, 1, 0, 0, 0, 0, 0},
			{1012, 2, 0, 0, 0, 0, 0},

			//Its possible that these also controll the annoucement banners and chat messages ...
			// Functions to look at...
			//tower_announce_move() -> disp_tower_announce()
			// sendTowerVenomChatMsg()
		}
	case 5:
		//On load into MezePorta
		paperData = []PaperData{
			// getTowerQuestTowerLevel
			{1001, 1, 1, 0, 0, 0, 0},
			{1001, 2, 0, 0, 0, 0, 0},
			// iniTQT
			{1003, 1, 100, 100, 200, 100, 0},
			{1003, 2, 150, 100, 240, 100, 0},
			{1004, 10, 9999, 40, 0, 0, 0},
			{1005, 10, 500, 0, 0, 0, 0},
			// getPaperDataSetFromProp
			{1007, 1, 0, 0, 0, 0, 0},
			{1008, 200, 400, 3000, 400, 3000, 0},
			// getPaperDataSetParam1 / Dure Goal
			{1010, 1, 100, 0, 0, 0, 0},
			{1010, 2, 100, 0, 0, 0, 0},
			// update_disp_flag / getPaperDataSetParam1
			{1011, 1, 6000, 15000, 20000, 25000, 30000},
			{1011, 2, 6000, 15000, 20000, 25000, 30000},
			{1012, 1, 8000, 17500, 22500, 27500, 31000},
			{1012, 2, 8000, 17500, 22500, 27500, 31000},
			// setServerZako
			{1015, 1, mhfmon.Velociprey, mhfmon.Velociprey, mhfmon.Velociprey, 0, 0},
			{1015, 2, mhfmon.Velociprey, mhfmon.Velociprey, mhfmon.Velociprey, 0, 0},
			// createTowerFloorRandomNumberArray
			{1101, 1, 2016, 500, 0, 0, 0},
			{1101, 2, 2016, 500, 0, 0, 0},
			// HRP/SRP/GRP/GSRP/TRP reward
			{1103, 1, 0, 0, 3000, 0, 3000},
			{1103, 2, 0, 0, 3000, 0, 3000},
			// getTowerNextVenomLevel {ID, Block, MinFloor, MaxFloor, Frequency, Unk, Unk}
			{1104, 1, 10, 9999, 40, 0, 0},
			{1104, 2, 10, 9999, 40, 0, 0},
			{1105, 1, 10, 500, 0, 0, 0},
			{1105, 2, 10, 500, 0, 0, 0},
			// setServerBoss {ID, Block, Monster, Unk, Unk, Index, Points}
			{2001, 1, mhfmon.Velocidrome, 58, 0, 1, 100},
			{2001, 1, mhfmon.Bulldrome, 58, 0, 2, 150},
			{2001, 1, mhfmon.Gypceros, 58, 0, 3, 200},
			{2001, 1, mhfmon.Hypnocatrice, 58, 0, 4, 200},
			{2001, 1, mhfmon.Lavasioth, 58, 0, 5, 500},
			{2001, 1, mhfmon.Gravios, 58, 0, 6, 700},
			{2001, 1, mhfmon.Basarios, 58, 0, 7, 250},
			{2001, 1, mhfmon.Rajang, 58, 0, 8, 1000},
			{2001, 1, mhfmon.ShogunCeanataur, 58, 0, 9, 500},
			{2001, 1, mhfmon.Tigrex, 58, 0, 10, 800},
			{2001, 1, mhfmon.Espinas, 58, 0, 11, 900},
			{2001, 1, mhfmon.Pariapuria, 58, 0, 12, 600},

			{2001, 2, mhfmon.Velocidrome, 60, 0, 1, 100},
			{2001, 2, mhfmon.ShogunCeanataur, 60, 0, 2, 500},
			{2001, 2, mhfmon.Gypceros, 60, 0, 3, 200},
			{2001, 2, mhfmon.Hypnocatrice, 60, 0, 4, 200},
			{2001, 2, mhfmon.Lavasioth, 60, 0, 5, 500},

			{2001, 2, mhfmon.Gravios, 60, 0, 6, 700},
			{2001, 2, mhfmon.Basarios, 60, 0, 7, 350},
			{2001, 2, mhfmon.Rajang, 60, 0, 8, 1000},
			{2001, 2, mhfmon.Bulldrome, 60, 0, 9, 150},
			{2001, 2, mhfmon.Tigrex, 60, 0, 10, 800},
			{2001, 2, mhfmon.Espinas, 60, 0, 11, 900},
			{2001, 2, mhfmon.Pariapuria, 60, 0, 12, 600},
			{2001, 2, mhfmon.PurpleGypceros, 60, 0, 13, 200},
			{2001, 2, mhfmon.BurningEspinas, 60, 0, 14, 900},
			{2001, 2, mhfmon.YianGaruga, 60, 0, 15, 600},
			{2001, 2, mhfmon.Dyuragaua, 60, 0, 16, 1000},
		}
	case 6:
		//Loads on Tower Quest load
		paperData = []PaperData{
			// updateClearTowerFloor
			{1002, 100, 0, 0, 0, 0, 0},
			// give_gem_func
			{1006, 1, 10000, 10000, 0, 0, 0},
			{1006, 2, 10000, 20000, 0, 0, 0},
			{1009, 20, 0, 0, 0, 0, 0},
			// ttcStageInitDRP
			{1013, 1, 1, 1, 100, 200, 300},
			{1013, 1, 1, 2, 100, 200, 300},
			{1013, 1, 2, 1, 300, 100, 200},
			{1013, 1, 2, 2, 300, 100, 200},
			{1013, 1, 3, 1, 200, 300, 100},
			{1013, 1, 3, 2, 200, 300, 100},
			{1013, 2, 1, 1, 300, 100, 200},
			{1013, 2, 1, 2, 300, 100, 200},
			{1013, 2, 2, 1, 200, 300, 100},
			{1013, 2, 2, 2, 200, 300, 100},
			{1013, 2, 3, 1, 100, 200, 300},
			{1013, 2, 3, 2, 100, 200, 300},
			{1013, 3, 1, 1, 200, 300, 100},
			{1013, 3, 1, 2, 200, 300, 100},
			{1013, 3, 2, 1, 100, 200, 300},
			{1013, 3, 2, 2, 100, 200, 300},
			{1013, 3, 3, 1, 300, 100, 200},
			{1013, 3, 3, 2, 300, 100, 200},
			{1016, 1, 1, 80, 0, 0, 0},
			{1016, 1, 2, 80, 0, 0, 0},
			{1016, 1, 3, 80, 0, 0, 0},
			{1016, 2, 1, 80, 0, 0, 0},
			{1016, 2, 2, 80, 0, 0, 0},
			{1016, 2, 3, 80, 0, 0, 0},
			{1201, 1, 60, 50, 0, 0, 0},
			{1201, 2, 60, 50, 0, 0, 0},
			// Gimmick Damage {ID, Block, StartFloor, EndFloor, Multiplier*100, Unk, Unk}
			{1202, 1, 0, 5, 50, 0, 0},
			{1202, 1, 6, 20, 60, 0, 0},
			{1202, 1, 21, 40, 70, 0, 0},
			{1202, 1, 41, 120, 80, 0, 0},
			{1202, 1, 121, 160, 90, 0, 0},
			{1202, 1, 161, 250, 100, 0, 0},
			{1202, 1, 251, 500, 100, 0, 0},
			{1202, 1, 501, 9999, 100, 0, 0},
			{1202, 2, 0, 100, 100, 0, 0},
			{1202, 2, 101, 200, 100, 0, 0},
			{1202, 2, 201, 500, 150, 0, 0},
			{1202, 2, 501, 9999, 150, 0, 0},
			// Mon Damage {ID, Block, StartFloor, EndFloor, Multiplier*100, Unk, Unk}
			{1203, 1, 0, 5, 10, 0, 0},
			{1203, 1, 6, 10, 20, 0, 0},
			{1203, 1, 11, 30, 30, 0, 0},
			{1203, 1, 31, 60, 40, 0, 0},
			{1203, 1, 61, 120, 50, 0, 0},
			{1203, 1, 121, 130, 60, 0, 0},
			{1203, 1, 131, 140, 70, 0, 0},
			{1203, 1, 141, 150, 80, 0, 0},
			{1203, 1, 151, 160, 85, 0, 0},
			{1203, 1, 161, 200, 100, 0, 0},
			{1203, 1, 201, 500, 100, 0, 0},
			{1203, 1, 501, 9999, 100, 0, 0},
			{1203, 2, 0, 120, 70, 0, 0},
			{1203, 2, 121, 500, 120, 0, 0},
			{1203, 2, 501, 9999, 120, 0, 0},
			// Mon HP {ID, Block, StartFloor, EndFloor, Multiplier*100, Unk, Unk}
			{1204, 1, 0, 5, 15, 0, 0},
			{1204, 1, 6, 10, 20, 0, 0},
			{1204, 1, 11, 15, 25, 0, 0},
			{1204, 1, 16, 20, 27, 0, 0},
			{1204, 1, 21, 25, 30, 0, 0},
			{1204, 1, 26, 30, 32, 0, 0},
			{1204, 1, 31, 40, 35, 0, 0},
			{1204, 1, 41, 50, 37, 0, 0},
			{1204, 1, 51, 60, 40, 0, 0},
			{1204, 1, 61, 70, 43, 0, 0},
			{1204, 1, 71, 80, 45, 0, 0},
			{1204, 1, 81, 90, 47, 0, 0},
			{1204, 1, 91, 100, 50, 0, 0},
			{1204, 1, 101, 110, 60, 0, 0},
			{1204, 1, 111, 120, 70, 0, 0},
			{1204, 1, 121, 130, 75, 0, 0},
			{1204, 1, 131, 140, 82, 0, 0},
			{1204, 1, 141, 160, 85, 0, 0},
			{1204, 1, 161, 200, 100, 0, 0},
			{1204, 1, 201, 500, 100, 0, 0},
			{1204, 1, 501, 9999, 100, 0, 0},
			{1204, 2, 0, 120, 70, 0, 0},
			{1204, 2, 121, 500, 120, 0, 0},
			{1204, 2, 501, 9999, 120, 0, 0},
			// Supply Items {ID, Block, Unk, ItemID, Quantity, Unk, Unk}
			{4001, 1, 0, 0, 0, 0, 0},
			{4001, 2, 0, 10667, 5, 0, 1},
			{4001, 2, 0, 10667, 5, 0, 1},
			{4001, 2, 0, 10667, 5, 0, 1},
			{4001, 2, 0, 10667, 5, 0, 1},
			{4001, 2, 0, 10668, 2, 0, 1},
			{4001, 2, 0, 10668, 2, 0, 1},
			{4001, 2, 0, 10668, 2, 0, 1},
			{4001, 2, 0, 10668, 2, 0, 1},
			{4001, 2, 0, 10669, 1, 0, 1},
			{4001, 2, 0, 10669, 1, 0, 1},
			{4001, 2, 0, 10669, 1, 0, 1},
			{4001, 2, 0, 10669, 1, 0, 1},
			{4001, 2, 0, 10671, 3, 0, 1},
			{4001, 2, 0, 10671, 3, 0, 1},
			{4001, 2, 0, 10671, 3, 0, 1},
			{4001, 2, 0, 10671, 3, 0, 1},
			{4001, 2, 0, 10384, 1, 0, 1},
			{4001, 2, 0, 10384, 1, 0, 1},
			{4001, 2, 0, 10670, 2, 0, 1},
			{4001, 2, 0, 10670, 2, 0, 1},
			{4001, 2, 0, 10682, 2, 0, 1},
			{4001, 2, 0, 10683, 2, 0, 1},
			{4001, 2, 0, 10678, 1, 0, 1},
			{4001, 2, 0, 10678, 1, 0, 1},
			// Item Rewards {ID, Block, Unk, ItemID, Quantity?, Chance*100, Unk}
			{4005, 1, 0, 11159, 1, 5000, 1},
			{4005, 1, 0, 11160, 1, 3350, 1},
			{4005, 1, 0, 11161, 1, 1500, 1},
			{4005, 1, 0, 11162, 1, 100, 1},
			{4005, 1, 0, 11163, 1, 50, 1},
			{4005, 2, 0, 11159, 2, 1800, 1},
			{4005, 2, 0, 11160, 2, 1200, 1},
			{4005, 2, 0, 11161, 2, 500, 1},
			{4005, 2, 0, 11162, 1, 50, 1},
			{4005, 2, 0, 11037, 1, 150, 1},
			{4005, 2, 0, 11038, 1, 150, 1},
			{4005, 2, 0, 11044, 1, 150, 1},
			{4005, 2, 0, 11057, 1, 150, 1},
			{4005, 2, 0, 11059, 1, 150, 1},
			{4005, 2, 0, 11079, 1, 150, 1},
			{4005, 2, 0, 11098, 1, 150, 1},
			{4005, 2, 0, 11104, 1, 150, 1},
			{4005, 2, 0, 11117, 1, 150, 1},
			{4005, 2, 0, 11128, 1, 150, 1},
			{4005, 2, 0, 11133, 1, 150, 1},
			{4005, 2, 0, 11137, 1, 150, 1},
			{4005, 2, 0, 11143, 1, 150, 1},
			{4005, 2, 0, 11132, 1, 150, 1},
			{4005, 2, 0, 11039, 1, 150, 1},
			{4005, 2, 0, 11040, 1, 150, 1},
			{4005, 2, 0, 11049, 1, 150, 1},
			{4005, 2, 0, 11061, 1, 150, 1},
			{4005, 2, 0, 11063, 1, 150, 1},
			{4005, 2, 0, 11077, 1, 150, 1},
			{4005, 2, 0, 11099, 1, 150, 1},
			{4005, 2, 0, 11105, 1, 150, 1},
			{4005, 2, 0, 11129, 1, 150, 1},
			{4005, 2, 0, 11130, 1, 150, 1},
			{4005, 2, 0, 11131, 1, 150, 1},
			{4005, 2, 0, 11139, 1, 150, 1},
			{4005, 2, 0, 11145, 1, 150, 1},
			{4005, 2, 0, 11096, 1, 150, 1},
			{4005, 2, 0, 11041, 1, 150, 1},
			{4005, 2, 0, 11047, 1, 150, 1},
			{4005, 2, 0, 11054, 1, 150, 1},
			{4005, 2, 0, 11065, 1, 150, 1},
			{4005, 2, 0, 11068, 1, 150, 1},
			{4005, 2, 0, 11075, 1, 150, 1},
			{4005, 2, 0, 11100, 1, 150, 1},
			{4005, 2, 0, 11106, 1, 150, 1},
			{4005, 2, 0, 11119, 1, 150, 1},
			{4005, 2, 0, 11135, 1, 150, 1},
			{4005, 2, 0, 11136, 1, 150, 1},
			{4005, 2, 0, 11138, 1, 150, 1},
			{4005, 2, 0, 11088, 1, 150, 1},
			{4005, 2, 0, 10370, 1, 150, 1},
			{4005, 2, 0, 10368, 1, 150, 1},
			{4006, 1, 0, 11159, 1, 5000, 1},
			{4006, 1, 0, 11160, 1, 3350, 1},
			{4006, 1, 0, 11161, 1, 1500, 1},
			{4006, 1, 0, 11162, 1, 100, 1},
			{4006, 1, 0, 11163, 1, 50, 1},
			{4006, 2, 0, 11159, 2, 1800, 1},
			{4006, 2, 0, 11160, 2, 1200, 1},
			{4006, 2, 0, 11161, 2, 500, 1},
			{4006, 2, 0, 11162, 1, 50, 1},
			{4006, 2, 0, 11037, 1, 150, 1},
			{4006, 2, 0, 11038, 1, 150, 1},
			{4006, 2, 0, 11044, 1, 150, 1},
			{4006, 2, 0, 11057, 1, 150, 1},
			{4006, 2, 0, 11059, 1, 150, 1},
			{4006, 2, 0, 11079, 1, 150, 1},
			{4006, 2, 0, 11098, 1, 150, 1},
			{4006, 2, 0, 11104, 1, 150, 1},
			{4006, 2, 0, 11117, 1, 150, 1},
			{4006, 2, 0, 11128, 1, 150, 1},
			{4006, 2, 0, 11133, 1, 150, 1},
			{4006, 2, 0, 11137, 1, 150, 1},
			{4006, 2, 0, 11143, 1, 150, 1},
			{4006, 2, 0, 11132, 1, 150, 1},
			{4006, 2, 0, 11039, 1, 150, 1},
			{4006, 2, 0, 11040, 1, 150, 1},
			{4006, 2, 0, 11049, 1, 150, 1},
			{4006, 2, 0, 11061, 1, 150, 1},
			{4006, 2, 0, 11063, 1, 150, 1},
			{4006, 2, 0, 11077, 1, 150, 1},
			{4006, 2, 0, 11099, 1, 150, 1},
			{4006, 2, 0, 11105, 1, 150, 1},
			{4006, 2, 0, 11129, 1, 150, 1},
			{4006, 2, 0, 11130, 1, 150, 1},
			{4006, 2, 0, 11131, 1, 150, 1},
			{4006, 2, 0, 11139, 1, 150, 1},
			{4006, 2, 0, 11145, 1, 150, 1},
			{4006, 2, 0, 11096, 1, 150, 1},
			{4006, 2, 0, 11041, 1, 150, 1},
			{4006, 2, 0, 11047, 1, 150, 1},
			{4006, 2, 0, 11054, 1, 150, 1},
			{4006, 2, 0, 11065, 1, 150, 1},
			{4006, 2, 0, 11068, 1, 150, 1},
			{4006, 2, 0, 11075, 1, 150, 1},
			{4006, 2, 0, 11100, 1, 150, 1},
			{4006, 2, 0, 11106, 1, 150, 1},
			{4006, 2, 0, 11119, 1, 150, 1},
			{4006, 2, 0, 11135, 1, 150, 1},
			{4006, 2, 0, 11136, 1, 150, 1},
			{4006, 2, 0, 11138, 1, 150, 1},
			{4006, 2, 0, 11088, 1, 150, 1},
			{4006, 2, 0, 10370, 1, 150, 1},
			{4006, 2, 0, 10368, 1, 150, 1},
			{4007, 1, 0, 11058, 1, 70, 1},
			{4007, 1, 0, 11060, 1, 70, 1},
			{4007, 1, 0, 11062, 1, 70, 1},
			{4007, 1, 0, 11064, 1, 70, 1},
			{4007, 1, 0, 11066, 1, 70, 1},
			{4007, 1, 0, 11118, 1, 70, 1},
			{4007, 1, 0, 11120, 1, 70, 1},
			{4007, 1, 0, 11110, 1, 70, 1},
			{4007, 1, 0, 11112, 1, 70, 1},
			{4007, 1, 0, 11114, 1, 70, 1},
			{4007, 1, 0, 11042, 1, 70, 1},
			{4007, 1, 0, 11043, 1, 70, 1},
			{4007, 1, 0, 11074, 1, 70, 1},
			{4007, 1, 0, 11140, 1, 70, 1},
			{4007, 1, 0, 11067, 1, 70, 1},
			{4007, 1, 0, 11048, 1, 70, 1},
			{4007, 1, 0, 11046, 1, 70, 1},
			{4007, 1, 0, 11103, 1, 70, 1},
			{4007, 1, 0, 11107, 1, 70, 1},
			{4007, 1, 0, 11108, 1, 70, 1},
			{4007, 1, 0, 11121, 1, 70, 1},
			{4007, 1, 0, 11134, 1, 70, 1},
			{4007, 1, 0, 11084, 1, 70, 1},
			{4007, 1, 0, 11085, 1, 70, 1},
			{4007, 1, 0, 11086, 1, 70, 1},
			{4007, 1, 0, 11087, 1, 70, 1},
			{4007, 1, 0, 11094, 1, 70, 1},
			{4007, 1, 0, 11095, 1, 70, 1},
			{4007, 1, 0, 10374, 1, 70, 1},
			{4007, 1, 0, 10375, 1, 70, 1},
			{4007, 1, 0, 10376, 1, 70, 1},
			{4007, 1, 0, 10377, 1, 70, 1},
			{4007, 1, 0, 10378, 1, 70, 1},
			{4007, 1, 0, 11069, 1, 45, 1},
			{4007, 1, 0, 11071, 1, 45, 1},
			{4007, 1, 0, 11073, 1, 45, 1},
			{4007, 1, 0, 11076, 1, 45, 1},
			{4007, 1, 0, 11078, 1, 45, 1},
			{4007, 1, 0, 11116, 1, 45, 1},
			{4007, 1, 0, 11123, 1, 45, 1},
			{4007, 1, 0, 11127, 1, 45, 1},
			{4007, 1, 0, 11142, 1, 45, 1},
			{4007, 1, 0, 11056, 1, 45, 1},
			{4007, 1, 0, 11090, 1, 45, 1},
			{4007, 1, 0, 11097, 1, 45, 1},
			{4007, 1, 0, 10367, 1, 45, 1},
			{4007, 1, 0, 10371, 1, 45, 1},
			{4007, 1, 0, 10373, 1, 45, 1},
			{4007, 1, 0, 11080, 1, 15, 1},
			{4007, 1, 0, 11081, 1, 15, 1},
			{4007, 1, 0, 11083, 1, 15, 1},
			{4007, 1, 0, 11125, 1, 15, 1},
			{4007, 1, 0, 11093, 1, 14, 1},
			{4007, 1, 0, 11053, 1, 10, 1},
			{4007, 1, 0, 11147, 1, 10, 1},
			{4007, 1, 0, 10372, 1, 5, 1},
			{4007, 1, 0, 10369, 1, 1, 1},
			{4007, 1, 0, 11163, 1, 150, 1},
			{4007, 1, 0, 11465, 1, 50, 1},
			{4007, 1, 0, 11466, 1, 25, 1},
			{4007, 1, 0, 11467, 1, 200, 1},
			{4007, 1, 0, 11468, 1, 400, 1},
			{4007, 1, 0, 11469, 1, 150, 1},
			{4007, 1, 0, 11037, 1, 92, 1},
			{4007, 1, 0, 11038, 1, 92, 1},
			{4007, 1, 0, 11044, 1, 92, 1},
			{4007, 1, 0, 11057, 1, 92, 1},
			{4007, 1, 0, 11059, 1, 92, 1},
			{4007, 1, 0, 11079, 1, 92, 1},
			{4007, 1, 0, 11098, 1, 92, 1},
			{4007, 1, 0, 11104, 1, 92, 1},
			{4007, 1, 0, 11117, 1, 92, 1},
			{4007, 1, 0, 11133, 1, 92, 1},
			{4007, 1, 0, 11137, 1, 92, 1},
			{4007, 1, 0, 11143, 1, 92, 1},
			{4007, 1, 0, 11132, 1, 92, 1},
			{4007, 1, 0, 11039, 1, 92, 1},
			{4007, 1, 0, 11040, 1, 92, 1},
			{4007, 1, 0, 11049, 1, 92, 1},
			{4007, 1, 0, 11061, 1, 92, 1},
			{4007, 1, 0, 11063, 1, 92, 1},
			{4007, 1, 0, 11077, 1, 92, 1},
			{4007, 1, 0, 11099, 1, 92, 1},
			{4007, 1, 0, 11105, 1, 92, 1},
			{4007, 1, 0, 11129, 1, 92, 1},
			{4007, 1, 0, 11130, 1, 92, 1},
			{4007, 1, 0, 11131, 1, 92, 1},
			{4007, 1, 0, 11139, 1, 92, 1},
			{4007, 1, 0, 11145, 1, 91, 1},
			{4007, 1, 0, 11096, 1, 91, 1},
			{4007, 1, 0, 11041, 1, 91, 1},
			{4007, 1, 0, 11047, 1, 91, 1},
			{4007, 1, 0, 11054, 1, 91, 1},
			{4007, 1, 0, 11065, 1, 91, 1},
			{4007, 1, 0, 11068, 1, 91, 1},
			{4007, 1, 0, 11075, 1, 91, 1},
			{4007, 1, 0, 11100, 1, 91, 1},
			{4007, 1, 0, 11106, 1, 91, 1},
			{4007, 1, 0, 11119, 1, 91, 1},
			{4007, 1, 0, 11135, 1, 91, 1},
			{4007, 1, 0, 11136, 1, 91, 1},
			{4007, 1, 0, 11138, 1, 91, 1},
			{4007, 1, 0, 11088, 1, 91, 1},
			{4007, 1, 0, 10370, 1, 91, 1},
			{4007, 1, 0, 10368, 1, 91, 1},
			{4007, 1, 0, 11045, 1, 91, 1},
			{4007, 1, 0, 11070, 1, 91, 1},
			{4007, 1, 0, 11101, 1, 91, 1},
			{4007, 1, 0, 11109, 1, 91, 1},
			{4007, 1, 0, 11122, 1, 91, 1},
			{4007, 1, 0, 11141, 1, 91, 1},
			{4007, 1, 0, 11051, 1, 91, 1},
			{4007, 1, 0, 11102, 1, 91, 1},
			{4007, 1, 0, 11124, 1, 91, 1},
			{4007, 1, 0, 11072, 1, 91, 1},
			{4007, 1, 0, 11082, 1, 91, 1},
			{4007, 1, 0, 11115, 1, 91, 1},
			{4007, 1, 0, 11144, 1, 91, 1},
			{4007, 1, 0, 11089, 1, 91, 1},
			{4007, 1, 0, 11091, 1, 91, 1},
			{4007, 1, 0, 11092, 1, 91, 1},
			{4007, 1, 0, 11050, 1, 91, 1},
			{4007, 1, 0, 11111, 1, 91, 1},
			{4007, 1, 0, 11113, 1, 91, 1},
			{4007, 1, 0, 11126, 1, 91, 1},
			{4007, 1, 0, 11055, 1, 91, 1},
			{4007, 1, 0, 11052, 1, 91, 1},
			{4007, 1, 0, 11146, 1, 91, 1},
			{4007, 2, 0, 11058, 1, 90, 1},
			{4007, 2, 0, 11060, 1, 90, 1},
			{4007, 2, 0, 11062, 1, 90, 1},
			{4007, 2, 0, 11064, 1, 90, 1},
			{4007, 2, 0, 11066, 1, 90, 1},
			{4007, 2, 0, 11118, 1, 90, 1},
			{4007, 2, 0, 11120, 1, 90, 1},
			{4007, 2, 0, 11110, 1, 90, 1},
			{4007, 2, 0, 11112, 1, 90, 1},
			{4007, 2, 0, 11114, 1, 90, 1},
			{4007, 2, 0, 11042, 1, 90, 1},
			{4007, 2, 0, 11043, 1, 90, 1},
			{4007, 2, 0, 11074, 1, 90, 1},
			{4007, 2, 0, 11140, 1, 90, 1},
			{4007, 2, 0, 11067, 1, 90, 1},
			{4007, 2, 0, 11048, 1, 90, 1},
			{4007, 2, 0, 11046, 1, 90, 1},
			{4007, 2, 0, 11103, 1, 90, 1},
			{4007, 2, 0, 11107, 1, 90, 1},
			{4007, 2, 0, 11108, 1, 90, 1},
			{4007, 2, 0, 11121, 1, 90, 1},
			{4007, 2, 0, 11134, 1, 90, 1},
			{4007, 2, 0, 11084, 1, 90, 1},
			{4007, 2, 0, 11085, 1, 90, 1},
			{4007, 2, 0, 11086, 1, 90, 1},
			{4007, 2, 0, 11087, 1, 90, 1},
			{4007, 2, 0, 11094, 1, 90, 1},
			{4007, 2, 0, 11095, 1, 90, 1},
			{4007, 2, 0, 10374, 1, 90, 1},
			{4007, 2, 0, 10375, 1, 90, 1},
			{4007, 2, 0, 10376, 1, 90, 1},
			{4007, 2, 0, 10377, 1, 90, 1},
			{4007, 2, 0, 10378, 1, 90, 1},
			{4007, 2, 0, 11069, 1, 80, 1},
			{4007, 2, 0, 11071, 1, 80, 1},
			{4007, 2, 0, 11073, 1, 80, 1},
			{4007, 2, 0, 11076, 1, 80, 1},
			{4007, 2, 0, 11078, 1, 80, 1},
			{4007, 2, 0, 11116, 1, 80, 1},
			{4007, 2, 0, 11123, 1, 80, 1},
			{4007, 2, 0, 11127, 1, 80, 1},
			{4007, 2, 0, 11142, 1, 80, 1},
			{4007, 2, 0, 11056, 1, 80, 1},
			{4007, 2, 0, 11090, 1, 80, 1},
			{4007, 2, 0, 11097, 1, 80, 1},
			{4007, 2, 0, 10367, 1, 80, 1},
			{4007, 2, 0, 10371, 1, 80, 1},
			{4007, 2, 0, 10373, 1, 80, 1},
			{4007, 2, 0, 11080, 1, 22, 1},
			{4007, 2, 0, 11081, 1, 22, 1},
			{4007, 2, 0, 11083, 1, 22, 1},
			{4007, 2, 0, 11125, 1, 22, 1},
			{4007, 2, 0, 11093, 1, 22, 1},
			{4007, 2, 0, 11053, 1, 15, 1},
			{4007, 2, 0, 11147, 1, 15, 1},
			{4007, 2, 0, 10372, 1, 8, 1},
			{4007, 2, 0, 10369, 1, 2, 1},
			{4007, 2, 0, 11159, 3, 1220, 1},
			{4007, 2, 0, 11160, 3, 650, 1},
			{4007, 2, 0, 11161, 3, 160, 1},
			{4007, 2, 0, 11661, 1, 800, 1},
			{4007, 2, 0, 11662, 1, 800, 1},
			{4007, 2, 0, 11163, 1, 500, 1},
			{4007, 2, 0, 11162, 1, 550, 1},
			{4007, 2, 0, 11465, 1, 50, 1},
			{4007, 2, 0, 11466, 1, 25, 1},
			{4007, 2, 0, 11467, 1, 250, 1},
			{4007, 2, 0, 11468, 1, 500, 1},
			{4007, 2, 0, 11469, 1, 175, 1},
			// Probably treasure chest rewards
			{4202, 1, 0, 11163, 1, 6000, 1},
			{4202, 1, 0, 11465, 1, 200, 1},
			{4202, 1, 0, 11466, 1, 100, 1},
			{4202, 1, 0, 11467, 1, 1000, 1},
			{4202, 1, 0, 11468, 1, 2000, 1},
			{4202, 1, 0, 11469, 1, 700, 1},
			{4202, 2, 0, 11661, 1, 800, 1},
			{4202, 2, 0, 11662, 1, 800, 1},
			{4202, 2, 0, 11163, 1, 400, 1},
			{4202, 2, 0, 11465, 1, 400, 1},
			{4202, 2, 0, 11466, 1, 200, 1},
			{4202, 2, 0, 11467, 1, 2000, 1},
			{4202, 2, 0, 11468, 1, 4000, 1},
			{4202, 2, 0, 11469, 1, 1400, 1},
		}
	default:
		if pkt.ID < 1000 {
			s.logger.Info("PaperData request for unknown type", zap.Uint32("Unk2", pkt.ID))
		}

	}

	switch pkt.Type {

	case 0:
		for _, pdata := range paperData {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(pdata.ID)
			bf.WriteInt16(pdata.Ward)
			bf.WriteInt16(pdata.Option1)
			bf.WriteInt16(pdata.Option2)
			bf.WriteInt16(pdata.Option3)
			bf.WriteInt16(pdata.Option4)
			bf.WriteInt16(pdata.Option5)
			data = append(data, bf)
		}
		doAckEarthSucceed(s, pkt.AckHandle, data)
	case 2:
		var paperGifts []PaperGift
		var paperGift PaperGift
		paperGiftData, err := s.server.db.Queryx("SELECT  item_id, unk0, unk1, chance FROM paper_data_gifts WHERE gift_id = $1", pkt.ID)
		if err != nil {
			paperGiftData.Close()
			s.logger.Info("PaperGift request for unknown type", zap.Uint32("ID", pkt.ID))

		}

		for paperGiftData.Next() {
			err = paperGiftData.Scan(&paperGift.ItemID, &paperGift.Unk1, &paperGift.Unk2, &paperGift.Chance)
			if err != nil {
				continue
			}
			paperGifts = append(paperGifts, paperGift)
		}
		for _, gift := range paperGifts {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(gift.ItemID)
			bf.WriteUint8(gift.Unk1)
			bf.WriteUint8(gift.Unk2)
			bf.WriteUint16(gift.Chance)
			data = append(data, bf)
		}
		doAckEarthSucceed(s, pkt.AckHandle, data)

	case 3:
		bf := byteframe.NewByteFrame()
		bf.WriteUint16(uint16(len(paperMissions.Timetables)))
		bf.WriteUint16(uint16(len(paperMissions.Data)))
		for _, timetable := range paperMissions.Timetables {
			bf.WriteUint32(uint32(timetable.Start.Unix()))
			bf.WriteUint32(uint32(timetable.End.Unix()))
		}
		for _, mdata := range paperMissions.Data {
			bf.WriteUint8(mdata.Unk0)
			bf.WriteUint8(mdata.Unk1)
			bf.WriteInt16(mdata.Target)
			bf.WriteUint16(mdata.Reward1ID)
			bf.WriteUint8(mdata.Reward1Quantity)
			bf.WriteUint16(mdata.Reward2ID)
			bf.WriteUint8(mdata.Reward2Quantity)
		}
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
	default:
		s.logger.Info("PaperData request for unknown type", zap.Uint32("Type", pkt.Type))

	}
}

func handleMsgSysAuthData(s *Session, p mhfpacket.MHFPacket) {}
