package channelserver

import (
	"errors"

	_config "erupe-ce/config"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

func GetCharacterSaveData(s *Session, charID uint32) (*CharacterSaveData, error) {
	result, err := s.server.db.Query("SELECT id, savedata, is_new_character, name FROM characters WHERE id = $1", charID)
	if err != nil {
		s.logger.Error("Failed to get savedata", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}
	defer func() { _ = result.Close() }()
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

	if _, err := s.server.db.Exec(`UPDATE user_binary SET house_tier=$1, house_data=$2, bookshelf=$3, gallery=$4, tore=$5, garden=$6 WHERE id=$7
	`, save.HouseTier, save.HouseData, save.BookshelfData, save.GalleryData, save.ToreData, save.GardenData, s.charID); err != nil {
		s.logger.Error("Failed to update user binary house data", zap.Error(err))
	}
}

func handleMsgMhfSexChanger(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSexChanger)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
