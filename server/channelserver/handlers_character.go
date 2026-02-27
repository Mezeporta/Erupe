package channelserver

import (
	"database/sql"
	"errors"
	"fmt"

	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

// GetCharacterSaveData loads a character's save data from the database.
func GetCharacterSaveData(s *Session, charID uint32) (*CharacterSaveData, error) {
	id, savedata, isNew, name, err := s.server.charRepo.LoadSaveData(charID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("No savedata found", zap.Uint32("charID", charID))
			return nil, errors.New("no savedata found")
		}
		s.logger.Error("Failed to get savedata", zap.Error(err), zap.Uint32("charID", charID))
		return nil, err
	}

	saveData := &CharacterSaveData{
		CharID:         id,
		compSave:       savedata,
		IsNewCharacter: isNew,
		Name:           name,
		Mode:           s.server.erupeConfig.RealClientMode,
		Pointers:       getPointers(s.server.erupeConfig.RealClientMode),
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

func (save *CharacterSaveData) Save(s *Session) error {
	if save.decompSave == nil {
		s.logger.Warn("No decompressed save data, skipping save",
			zap.Uint32("charID", save.CharID),
		)
		return errors.New("no decompressed save data")
	}

	if !s.kqfOverride {
		s.kqf = save.KQF
	} else {
		save.KQF = s.kqf
	}

	save.updateSaveDataWithStruct()

	if s.server.erupeConfig.RealClientMode >= cfg.G1 {
		err := save.Compress()
		if err != nil {
			s.logger.Error("Failed to compress savedata", zap.Error(err))
			return fmt.Errorf("compress savedata: %w", err)
		}
	} else {
		// Saves were not compressed
		save.compSave = save.decompSave
	}

	if err := s.server.charRepo.SaveCharacterData(save.CharID, save.compSave, save.HR, save.GR, save.Gender, save.WeaponType, save.WeaponID); err != nil {
		s.logger.Error("Failed to update savedata", zap.Error(err), zap.Uint32("charID", save.CharID))
		return fmt.Errorf("save character data: %w", err)
	}

	if err := s.server.charRepo.SaveHouseData(s.charID, save.HouseTier, save.HouseData, save.BookshelfData, save.GalleryData, save.ToreData, save.GardenData); err != nil {
		s.logger.Error("Failed to update user binary house data", zap.Error(err))
		return fmt.Errorf("save house data: %w", err)
	}

	return nil
}

func handleMsgMhfSexChanger(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSexChanger)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
