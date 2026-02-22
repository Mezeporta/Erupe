package channelserver

import (
	"encoding/binary"

	"erupe-ce/common/bfutil"
	"erupe-ce/common/stringsupport"
	cfg "erupe-ce/config"
	"erupe-ce/server/channelserver/compression/nullcomp"
)

// SavePointer identifies a section within the character save data blob.
type SavePointer int

const (
	pGender = iota
	pRP
	pHouseTier
	pHouseData
	pBookshelfData
	pGalleryData
	pToreData
	pGardenData
	pPlaytime
	pWeaponType
	pWeaponID
	pHR
	pGRP
	pKQF
	lBookshelfData
)

// CharacterSaveData holds a character's save data and its parsed fields.
type CharacterSaveData struct {
	CharID         uint32
	Name           string
	IsNewCharacter bool
	Mode           cfg.Mode
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

func getPointers(mode cfg.Mode) map[SavePointer]int {
	pointers := map[SavePointer]int{pGender: 81, lBookshelfData: 5576}
	switch mode {
	case cfg.ZZ:
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
	case cfg.Z2, cfg.Z1, cfg.G101, cfg.G10, cfg.G91, cfg.G9, cfg.G81, cfg.G8,
		cfg.G7, cfg.G61, cfg.G6, cfg.G52, cfg.G51, cfg.G5, cfg.GG, cfg.G32, cfg.G31,
		cfg.G3, cfg.G2, cfg.G1:
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
	case cfg.F5, cfg.F4:
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
	case cfg.S6:
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
	if mode == cfg.G5 {
		pointers[lBookshelfData] = 5548
	} else if mode <= cfg.GG {
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
	if save.Mode >= cfg.F4 {
		copy(save.decompSave[save.Pointers[pRP]:save.Pointers[pRP]+saveFieldRP], rpBytes)
	}
	if save.Mode >= cfg.G10 {
		copy(save.decompSave[save.Pointers[pKQF]:save.Pointers[pKQF]+saveFieldKQF], save.KQF)
	}
}

// This will update the save struct with the values stored in the character save
// Save data field sizes
const (
	saveFieldRP         = 2
	saveFieldHouseTier  = 5
	saveFieldHouseData  = 195
	saveFieldGallery    = 1748
	saveFieldTore       = 240
	saveFieldGarden     = 68
	saveFieldPlaytime   = 4
	saveFieldWeaponID   = 2
	saveFieldHR         = 2
	saveFieldGRP        = 4
	saveFieldKQF        = 8
	saveFieldNameOffset = 88
	saveFieldNameLen    = 12
)

func (save *CharacterSaveData) updateStructWithSaveData() {
	save.Name = stringsupport.SJISToUTF8Lossy(bfutil.UpToNull(save.decompSave[saveFieldNameOffset : saveFieldNameOffset+saveFieldNameLen]))
	if save.decompSave[save.Pointers[pGender]] == 1 {
		save.Gender = true
	} else {
		save.Gender = false
	}
	if !save.IsNewCharacter {
		if save.Mode >= cfg.S6 {
			save.RP = binary.LittleEndian.Uint16(save.decompSave[save.Pointers[pRP] : save.Pointers[pRP]+saveFieldRP])
			save.HouseTier = save.decompSave[save.Pointers[pHouseTier] : save.Pointers[pHouseTier]+saveFieldHouseTier]
			save.HouseData = save.decompSave[save.Pointers[pHouseData] : save.Pointers[pHouseData]+saveFieldHouseData]
			save.BookshelfData = save.decompSave[save.Pointers[pBookshelfData] : save.Pointers[pBookshelfData]+save.Pointers[lBookshelfData]]
			save.GalleryData = save.decompSave[save.Pointers[pGalleryData] : save.Pointers[pGalleryData]+saveFieldGallery]
			save.ToreData = save.decompSave[save.Pointers[pToreData] : save.Pointers[pToreData]+saveFieldTore]
			save.GardenData = save.decompSave[save.Pointers[pGardenData] : save.Pointers[pGardenData]+saveFieldGarden]
			save.Playtime = binary.LittleEndian.Uint32(save.decompSave[save.Pointers[pPlaytime] : save.Pointers[pPlaytime]+saveFieldPlaytime])
			save.WeaponType = save.decompSave[save.Pointers[pWeaponType]]
			save.WeaponID = binary.LittleEndian.Uint16(save.decompSave[save.Pointers[pWeaponID] : save.Pointers[pWeaponID]+saveFieldWeaponID])
			save.HR = binary.LittleEndian.Uint16(save.decompSave[save.Pointers[pHR] : save.Pointers[pHR]+saveFieldHR])
			if save.Mode >= cfg.G1 {
				if save.HR == uint16(999) {
					save.GR = grpToGR(int(binary.LittleEndian.Uint32(save.decompSave[save.Pointers[pGRP] : save.Pointers[pGRP]+saveFieldGRP])))
				}
			}
			if save.Mode >= cfg.G10 {
				save.KQF = save.decompSave[save.Pointers[pKQF] : save.Pointers[pKQF]+saveFieldKQF]
			}
		}
	}
}

// isHouseTierCorrupted checks whether the house tier field contains 0xFF
// bytes, which indicates an uninitialized or -1 value from the game client.
// The game uses small positive integers for theme IDs; 0xFF is never valid.
func (save *CharacterSaveData) isHouseTierCorrupted() bool {
	for _, b := range save.HouseTier {
		if b == 0xFF {
			return true
		}
	}
	return false
}

// restoreHouseTier replaces the current house tier with the given value in
// both the struct field and the underlying decompressed save blob, keeping
// them consistent for Save().
func (save *CharacterSaveData) restoreHouseTier(valid []byte) {
	save.HouseTier = make([]byte, len(valid))
	copy(save.HouseTier, valid)
	offset, ok := save.Pointers[pHouseTier]
	if ok && offset+len(valid) <= len(save.decompSave) {
		copy(save.decompSave[offset:offset+len(valid)], valid)
	}
}
