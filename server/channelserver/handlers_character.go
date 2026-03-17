package channelserver

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

// Backup configuration constants.
const (
	saveBackupSlots    = 3                // number of rotating backup slots per character
	saveBackupInterval = 30 * time.Minute // minimum time between backups
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

	// Capture the previous compressed savedata before it's overwritten by
	// Compress(). This is what gets backed up — the last known-good state.
	prevCompSave := save.compSave

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

	// Time-gated rotating backup: snapshot the previous compressed savedata
	// before overwriting, but only if enough time has elapsed since the last
	// backup. This keeps storage bounded (3 slots × blob size per character)
	// while providing recovery points.
	if len(prevCompSave) > 0 {
		maybeSaveBackup(s, save.CharID, prevCompSave)
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

// maybeSaveBackup checks whether enough time has elapsed since the last backup
// and, if so, writes the given compressed savedata into the next rotating slot.
// Errors are logged but do not block the save — backups are best-effort.
func maybeSaveBackup(s *Session, charID uint32, compSave []byte) {
	lastBackup, err := s.server.charRepo.GetLastBackupTime(charID)
	if err != nil {
		s.logger.Warn("Failed to query last backup time, skipping backup",
			zap.Error(err), zap.Uint32("charID", charID))
		return
	}

	if time.Since(lastBackup) < saveBackupInterval {
		return
	}

	// Pick the next slot using a simple counter derived from the backup times.
	// We rotate through slots 0, 1, 2 based on how many backups exist modulo
	// the slot count. In practice this fills slots in order and then overwrites
	// the oldest.
	slot := int(lastBackup.Unix()/int64(saveBackupInterval.Seconds())) % saveBackupSlots

	if err := s.server.charRepo.SaveBackup(charID, slot, compSave); err != nil {
		s.logger.Warn("Failed to save backup",
			zap.Error(err), zap.Uint32("charID", charID), zap.Int("slot", slot))
		return
	}

	s.logger.Info("Savedata backup created",
		zap.Uint32("charID", charID), zap.Int("slot", slot))
}

func handleMsgMhfSexChanger(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSexChanger)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
