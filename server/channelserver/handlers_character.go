package channelserver

import (
	"bytes"
	"crypto/sha256"
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

// GetCharacterSaveData loads a character's save data from the database and
// verifies its integrity checksum when one is stored.
func GetCharacterSaveData(s *Session, charID uint32) (*CharacterSaveData, error) {
	id, savedata, isNew, name, storedHash, err := s.server.charRepo.LoadSaveDataWithHash(charID)
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

	// Verify integrity checksum if one was stored with this save.
	// A nil hash means the character was saved before checksums were introduced,
	// so we skip verification (the next save will compute and store the hash).
	if storedHash != nil {
		computedHash := sha256.Sum256(saveData.decompSave)
		if !bytes.Equal(storedHash, computedHash[:]) {
			s.logger.Error("Savedata integrity check failed: hash mismatch",
				zap.Uint32("charID", charID),
				zap.Binary("stored_hash", storedHash),
				zap.Binary("computed_hash", computedHash[:]),
			)
			// TODO: attempt recovery from savedata_backups here
			return nil, errors.New("savedata integrity check failed")
		}
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

	// Compute integrity hash over the decompressed save.
	hash := sha256.Sum256(save.decompSave)

	// Build the atomic save params — character data, house data, hash, and
	// optionally a backup snapshot, all in one transaction.
	params := SaveAtomicParams{
		CharID:        save.CharID,
		CompSave:      save.compSave,
		Hash:          hash[:],
		HR:            save.HR,
		GR:            save.GR,
		IsFemale:      save.Gender,
		WeaponType:    save.WeaponType,
		WeaponID:      save.WeaponID,
		HouseTier:     save.HouseTier,
		HouseData:     save.HouseData,
		BookshelfData: save.BookshelfData,
		GalleryData:   save.GalleryData,
		ToreData:      save.ToreData,
		GardenData:    save.GardenData,
	}

	// Time-gated rotating backup: include the previous compressed savedata
	// in the transaction if enough time has elapsed since the last backup.
	if len(prevCompSave) > 0 {
		if slot, ok := shouldBackup(s, save.CharID); ok {
			params.BackupSlot = slot
			params.BackupData = prevCompSave
		}
	}

	if err := s.server.charRepo.SaveCharacterDataAtomic(params); err != nil {
		s.logger.Error("Failed to save character data atomically",
			zap.Error(err), zap.Uint32("charID", save.CharID))
		return fmt.Errorf("atomic save: %w", err)
	}

	return nil
}

// shouldBackup checks whether enough time has elapsed since the last backup
// and returns the target slot if a backup should be included in the save
// transaction. Returns (slot, true) if a backup is due, (0, false) otherwise.
func shouldBackup(s *Session, charID uint32) (int, bool) {
	lastBackup, err := s.server.charRepo.GetLastBackupTime(charID)
	if err != nil {
		s.logger.Warn("Failed to query last backup time, skipping backup",
			zap.Error(err), zap.Uint32("charID", charID))
		return 0, false
	}

	if time.Since(lastBackup) < saveBackupInterval {
		return 0, false
	}

	slot := int(lastBackup.Unix()/int64(saveBackupInterval.Seconds())) % saveBackupSlots
	return slot, true
}

func handleMsgMhfSexChanger(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSexChanger)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
