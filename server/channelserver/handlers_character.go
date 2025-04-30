package channelserver

import (
	"encoding/binary"
	"errors"
	"erupe-ce/common/bfutil"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"

	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"go.uber.org/zap"
)

type SavePointer int

const (
	pGender        = iota // +1
	pRP                   // +2
	pHouseTier            // +5
	pHouseData            // +195
	pBookshelfData        // +lBookshelfData
	pGalleryData          // +1748
	pToreData             // +240
	pGardenData           // +68
	pPlaytime             // +4
	pWeaponType           // +1
	pWeaponID             // +2
	pHR                   // +2
	pGRP                  // +4
	pKQF                  // +8
	lBookshelfData
)

type CharacterSaveData struct {
	CharID         uint32
	Name           string
	IsNewCharacter bool
	Pointers       map[SavePointer]int

	Gender        bool
	RP            uint16
	HouseTier     []byte
	HouseData     []byte
	BookshelfData []byte
	GalleryData   []byte
	ToreData      []byte
	GardenData    []byte
	Playtime      uint32
	WeaponType    uint8
	WeaponID      uint16
	HR            uint16
	GR            uint16
	KQF           []byte

	compSave   []byte
	decompSave []byte
}

func getPointers() map[SavePointer]int {
	pointers := map[SavePointer]int{pGender: 81, lBookshelfData: 5576}
	switch _config.ErupeConfig.RealClientMode {
	case _config.ZZ:
		pointers[pPlaytime] = 128356
		pointers[pWeaponID] = 128522
		pointers[pWeaponType] = 128789
		pointers[pHouseTier] = 129900
		pointers[pToreData] = 130228
		pointers[pHR] = 130550
		pointers[pGRP] = 130556
		pointers[pHouseData] = 130561
		pointers[pBookshelfData] = 139928
		pointers[pGalleryData] = 140064
		pointers[pGardenData] = 142424
		pointers[pRP] = 142614
		pointers[pKQF] = 146720
	case _config.Z2, _config.Z1, _config.G101, _config.G10, _config.G91, _config.G9, _config.G81, _config.G8,
		_config.G7, _config.G61, _config.G6, _config.G52, _config.G51, _config.G5, _config.GG, _config.G32, _config.G31,
		_config.G3, _config.G2, _config.G1:
		pointers[pPlaytime] = 92356
		pointers[pWeaponID] = 92522
		pointers[pWeaponType] = 92789
		pointers[pHouseTier] = 93900
		pointers[pToreData] = 94228
		pointers[pHR] = 94550
		pointers[pGRP] = 94556
		pointers[pHouseData] = 94561
		pointers[pBookshelfData] = 89118 // TODO: fix bookshelf data pointer
		pointers[pGalleryData] = 104064
		pointers[pGardenData] = 106424
		pointers[pRP] = 106614
		pointers[pKQF] = 110720
	case _config.F5, _config.F4:
		pointers[pPlaytime] = 60356
		pointers[pWeaponID] = 60522
		pointers[pWeaponType] = 60789
		pointers[pHouseTier] = 61900
		pointers[pToreData] = 62228
		pointers[pHR] = 62550
		pointers[pHouseData] = 62561
		pointers[pBookshelfData] = 57118 // TODO: fix bookshelf data pointer
		pointers[pGalleryData] = 72064
		pointers[pGardenData] = 74424
		pointers[pRP] = 74614
	case _config.S6:
		pointers[pPlaytime] = 12356
		pointers[pWeaponID] = 12522
		pointers[pWeaponType] = 12789
		pointers[pHouseTier] = 13900
		pointers[pToreData] = 14228
		pointers[pHR] = 14550
		pointers[pHouseData] = 14561
		pointers[pBookshelfData] = 9118 // TODO: fix bookshelf data pointer
		pointers[pGalleryData] = 24064
		pointers[pGardenData] = 26424
		pointers[pRP] = 26614
	}
	if _config.ErupeConfig.RealClientMode == _config.G5 {
		pointers[lBookshelfData] = 5548
	} else if _config.ErupeConfig.RealClientMode <= _config.GG {
		pointers[lBookshelfData] = 4520
	}
	return pointers
}

func GetCharacterSaveData(s *Session, charID uint32) (*CharacterSaveData, error) {
	result, err := s.server.db.Query("SELECT id, savedata, is_new_character, name FROM characters WHERE id = $1", charID)
	if err != nil {
		s.logger.Error("Failed to get savedata", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}
	defer result.Close()
	if !result.Next() {
		err = errors.New("no savedata found")
		s.logger.Error("No savedata found", zap.Uint32("charID", charID))
		return nil, err
	}

	saveData := &CharacterSaveData{
		Pointers: getPointers(),
	}
	err = result.Scan(&saveData.CharID, &saveData.compSave, &saveData.IsNewCharacter, &saveData.Name)
	if err != nil {
		s.logger.Error("Failed to scan savedata", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}

	if saveData.compSave == nil {
		return saveData, nil
	}

	err = saveData.Decompress()
	if err != nil {
		s.logger.Error("Failed to decompress savedata", zap.Error(err))
		return nil, err
	}

	saveData.updateStructWithSaveData()

	return saveData, nil
}

func (save *CharacterSaveData) Save(s *Session) {
	if !s.kqfOverride {
		s.kqf = save.KQF
	} else {
		save.KQF = s.kqf
	}

	save.updateSaveDataWithStruct()

	if _config.ErupeConfig.RealClientMode >= _config.G1 {
		err := save.Compress()
		if err != nil {
			s.logger.Error("Failed to compress savedata", zap.Error(err))
			return
		}
	} else {
		// Saves were not compressed
		save.compSave = save.decompSave
	}

	_, err := s.server.db.Exec(`UPDATE characters SET savedata=$1, is_new_character=false, hr=$2, gr=$3, is_female=$4, weapon_type=$5, weapon_id=$6 WHERE id=$7
	`, save.compSave, save.HR, save.GR, save.Gender, save.WeaponType, save.WeaponID, save.CharID)
	if err != nil {
		s.logger.Error("Failed to update savedata", zap.Error(err), zap.Uint32("charID", save.CharID))
	}

	s.server.db.Exec(`UPDATE user_binary SET house_tier=$1, house_data=$2, bookshelf=$3, gallery=$4, tore=$5, garden=$6 WHERE id=$7
	`, save.HouseTier, save.HouseData, save.BookshelfData, save.GalleryData, save.ToreData, save.GardenData, s.charID)
}

func (save *CharacterSaveData) Compress() error {
	var err error
	save.compSave, err = nullcomp.Compress(save.decompSave)
	if err != nil {
		return err
	}
	return nil
}

func (save *CharacterSaveData) Decompress() error {
	var err error
	save.decompSave, err = nullcomp.Decompress(save.compSave)
	if err != nil {
		return err
	}
	return nil
}

// This will update the character save with the values stored in the save struct
func (save *CharacterSaveData) updateSaveDataWithStruct() {
	rpBytes := make([]byte, 2)
	binary.LittleEndian.PutUint16(rpBytes, save.RP)
	if _config.ErupeConfig.RealClientMode >= _config.F4 {
		copy(save.decompSave[save.Pointers[pRP]:save.Pointers[pRP]+2], rpBytes)
	}
	if _config.ErupeConfig.RealClientMode >= _config.G10 {
		copy(save.decompSave[save.Pointers[pKQF]:save.Pointers[pKQF]+8], save.KQF)
	}
}

// This will update the save struct with the values stored in the character save
func (save *CharacterSaveData) updateStructWithSaveData() {
	save.Name = stringsupport.SJISToUTF8(bfutil.UpToNull(save.decompSave[88:100]))
	if save.decompSave[save.Pointers[pGender]] == 1 {
		save.Gender = true
	} else {
		save.Gender = false
	}
	if !save.IsNewCharacter {
		if _config.ErupeConfig.RealClientMode >= _config.S6 {
			save.RP = binary.LittleEndian.Uint16(save.decompSave[save.Pointers[pRP] : save.Pointers[pRP]+2])
			save.HouseTier = save.decompSave[save.Pointers[pHouseTier] : save.Pointers[pHouseTier]+5]
			save.HouseData = save.decompSave[save.Pointers[pHouseData] : save.Pointers[pHouseData]+195]
			save.BookshelfData = save.decompSave[save.Pointers[pBookshelfData] : save.Pointers[pBookshelfData]+save.Pointers[lBookshelfData]]
			save.GalleryData = save.decompSave[save.Pointers[pGalleryData] : save.Pointers[pGalleryData]+1748]
			save.ToreData = save.decompSave[save.Pointers[pToreData] : save.Pointers[pToreData]+240]
			save.GardenData = save.decompSave[save.Pointers[pGardenData] : save.Pointers[pGardenData]+68]
			save.Playtime = binary.LittleEndian.Uint32(save.decompSave[save.Pointers[pPlaytime] : save.Pointers[pPlaytime]+4])
			save.WeaponType = save.decompSave[save.Pointers[pWeaponType]]
			save.WeaponID = binary.LittleEndian.Uint16(save.decompSave[save.Pointers[pWeaponID] : save.Pointers[pWeaponID]+2])
			save.HR = binary.LittleEndian.Uint16(save.decompSave[save.Pointers[pHR] : save.Pointers[pHR]+2])
			if _config.ErupeConfig.RealClientMode >= _config.G1 {
				if save.HR == uint16(999) {
					save.GR = grpToGR(int(binary.LittleEndian.Uint32(save.decompSave[save.Pointers[pGRP] : save.Pointers[pGRP]+4])))
				}
			}
			if _config.ErupeConfig.RealClientMode >= _config.G10 {
				save.KQF = save.decompSave[save.Pointers[pKQF] : save.Pointers[pKQF]+8]
			}
		}
	}
	return
}

func handleMsgMhfSexChanger(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSexChanger)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
