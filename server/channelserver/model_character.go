package channelserver

import (
	"encoding/binary"

	"erupe-ce/common/bfutil"
	"erupe-ce/common/stringsupport"
	_config "erupe-ce/config"
	"erupe-ce/server/channelserver/compression/nullcomp"
)

// SavePointer identifies a section within the character save data blob.
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

// CharacterSaveData holds a character's save data and its parsed fields.
type CharacterSaveData struct {
	CharID         uint32
	Name           string
	IsNewCharacter bool
	Mode           _config.Mode
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

func getPointers(mode _config.Mode) map[SavePointer]int {
	pointers := map[SavePointer]int{pGender: 81, lBookshelfData: 5576}
	switch mode {
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
	if mode == _config.G5 {
		pointers[lBookshelfData] = 5548
	} else if mode <= _config.GG {
		pointers[lBookshelfData] = 4520
	}
	return pointers
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
	if save.Mode >= _config.F4 {
		copy(save.decompSave[save.Pointers[pRP]:save.Pointers[pRP]+2], rpBytes)
	}
	if save.Mode >= _config.G10 {
		copy(save.decompSave[save.Pointers[pKQF]:save.Pointers[pKQF]+8], save.KQF)
	}
}

// This will update the save struct with the values stored in the character save
func (save *CharacterSaveData) updateStructWithSaveData() {
	save.Name, _ = stringsupport.SJISToUTF8(bfutil.UpToNull(save.decompSave[88:100]))
	if save.decompSave[save.Pointers[pGender]] == 1 {
		save.Gender = true
	} else {
		save.Gender = false
	}
	if !save.IsNewCharacter {
		if save.Mode >= _config.S6 {
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
			if save.Mode >= _config.G1 {
				if save.HR == uint16(999) {
					save.GR = grpToGR(int(binary.LittleEndian.Uint32(save.decompSave[save.Pointers[pGRP] : save.Pointers[pGRP]+4])))
				}
			}
			if save.Mode >= _config.G10 {
				save.KQF = save.decompSave[save.Pointers[pKQF] : save.Pointers[pKQF]+8]
			}
		}
	}
}
