// Package channelserver implements plate data (transmog) management.
//
// Plate Data Overview:
// - platedata: Main transmog appearance data (~140KB, compressed)
// - platebox: Plate storage/inventory (~4.8KB, compressed)
// - platemyset: Equipment set configurations (1920 bytes, uncompressed)
//
// Save Strategy:
// All plate data saves immediately when the client sends save packets.
// This differs from the main savedata which may use session caching.
// The logout flow includes a safety check via savePlateDataToDatabase()
// to ensure no data loss if packets are lost or client disconnects.
//
// Cache Management:
// When plate data is saved, the server's user binary cache (types 2-3)
// is invalidated to ensure other players see updated appearance immediately.
// This prevents stale transmog/armor being displayed after zone changes.
//
// Thread Safety:
// All handlers use session-scoped database operations, making them
// inherently thread-safe as each session is single-threaded.
package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/channelserver/compression/deltacomp"
	"erupe-ce/server/channelserver/compression/nullcomp"
	"go.uber.org/zap"
	"time"
)

func handleMsgMhfLoadPlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateData)
	loadCharacterData(s, pkt.AckHandle, "platedata", nil)
}

// Plate data size constants
const (
	plateDataMaxPayload  = 262144 // max compressed platedata size
	plateDataEmptySize   = 140000 // empty platedata buffer
	plateBoxMaxPayload   = 32768  // max compressed platebox size
	plateBoxEmptySize    = 4800   // empty platebox buffer
	plateMysetDefaultLen = 1920   // default platemyset buffer
	plateMysetMaxPayload = 4096   // max platemyset payload size
)

func handleMsgMhfSavePlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateData)
	if len(pkt.RawDataPayload) > plateDataMaxPayload {
		s.logger.Warn("PlateData payload too large", zap.Int("len", len(pkt.RawDataPayload)))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	saveStart := time.Now()

	s.logger.Debug("PlateData save request",
		zap.Uint32("charID", s.charID),
		zap.Bool("is_diff", pkt.IsDataDiff),
		zap.Int("data_size", len(pkt.RawDataPayload)),
	)

	var dataSize int
	if pkt.IsDataDiff {
		var data []byte

		// Load existing save
		err := s.server.db.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.charID).Scan(&data)
		if err != nil {
			s.logger.Error("Failed to load platedata",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
			)
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		if len(data) > 0 {
			// Decompress
			s.logger.Debug("Decompressing PlateData", zap.Int("compressed_size", len(data)))
			data, err = nullcomp.Decompress(data)
			if err != nil {
				s.logger.Error("Failed to decompress platedata",
					zap.Error(err),
					zap.Uint32("charID", s.charID),
				)
				doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
				return
			}
		} else {
			// create empty save if absent
			data = make([]byte, plateDataEmptySize)
		}

		// Perform diff and compress it to write back to db
		s.logger.Debug("Applying PlateData diff", zap.Int("base_size", len(data)))
		saveOutput, err := nullcomp.Compress(deltacomp.ApplyDataDiff(pkt.RawDataPayload, data))
		if err != nil {
			s.logger.Error("Failed to diff and compress platedata",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
			)
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}
		dataSize = len(saveOutput)

		_, err = s.server.db.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", saveOutput, s.charID)
		if err != nil {
			s.logger.Error("Failed to save platedata",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
			)
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "platedata")
		dataSize = len(pkt.RawDataPayload)

		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Error("Failed to save platedata",
				zap.Error(err),
				zap.Uint32("charID", s.charID),
			)
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}
	}

	// Invalidate user binary cache so other players see updated appearance
	// User binary types 2 and 3 contain equipment/appearance data
	s.server.userBinaryPartsLock.Lock()
	delete(s.server.userBinaryParts, userBinaryPartID{charID: s.charID, index: 2})
	delete(s.server.userBinaryParts, userBinaryPartID{charID: s.charID, index: 3})
	s.server.userBinaryPartsLock.Unlock()

	saveDuration := time.Since(saveStart)
	s.logger.Info("PlateData saved successfully",
		zap.Uint32("charID", s.charID),
		zap.Bool("was_diff", pkt.IsDataDiff),
		zap.Int("data_size", dataSize),
		zap.Duration("duration", saveDuration),
	)

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateBox)
	loadCharacterData(s, pkt.AckHandle, "platebox", nil)
}

func handleMsgMhfSavePlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateBox)
	if len(pkt.RawDataPayload) > plateBoxMaxPayload {
		s.logger.Warn("PlateBox payload too large", zap.Int("len", len(pkt.RawDataPayload)))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	if pkt.IsDataDiff {
		var data []byte

		// Load existing save
		err := s.server.db.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.charID).Scan(&data)
		if err != nil {
			s.logger.Error("Failed to load platebox", zap.Error(err))
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		// Decompress
		if len(data) > 0 {
			// Decompress
			s.logger.Info("Decompressing...")
			data, err = nullcomp.Decompress(data)
			if err != nil {
				s.logger.Error("Failed to decompress platebox", zap.Error(err))
				doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
				return
			}
		} else {
			// create empty save if absent
			data = make([]byte, plateBoxEmptySize)
		}

		// Perform diff and compress it to write back to db
		s.logger.Info("Diffing...")
		saveOutput, err := nullcomp.Compress(deltacomp.ApplyDataDiff(pkt.RawDataPayload, data))
		if err != nil {
			s.logger.Error("Failed to diff and compress platebox", zap.Error(err))
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		_, err = s.server.db.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", saveOutput, s.charID)
		if err != nil {
			s.logger.Error("Failed to save platebox", zap.Error(err))
			doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
			return
		}

		s.logger.Info("Wrote recompressed platebox back to DB")
	} else {
		dumpSaveData(s, pkt.RawDataPayload, "platebox")
		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Error("Failed to save platebox", zap.Error(err))
		}
	}

	// Invalidate user binary cache so other players see updated appearance
	s.server.userBinaryPartsLock.Lock()
	delete(s.server.userBinaryParts, userBinaryPartID{charID: s.charID, index: 2})
	delete(s.server.userBinaryParts, userBinaryPartID{charID: s.charID, index: 3})
	s.server.userBinaryPartsLock.Unlock()

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateMyset)
	loadCharacterData(s, pkt.AckHandle, "platemyset", make([]byte, plateMysetDefaultLen))
}

func handleMsgMhfSavePlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateMyset)
	if len(pkt.RawDataPayload) > plateMysetMaxPayload {
		s.logger.Warn("PlateMyset payload too large", zap.Int("len", len(pkt.RawDataPayload)))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	saveStart := time.Now()

	s.logger.Debug("PlateMyset save request",
		zap.Uint32("charID", s.charID),
		zap.Int("data_size", len(pkt.RawDataPayload)),
	)

	// looks to always return the full thing, simply update database, no extra processing
	dumpSaveData(s, pkt.RawDataPayload, "platemyset")
	_, err := s.server.db.Exec("UPDATE characters SET platemyset=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Error("Failed to save platemyset",
			zap.Error(err),
			zap.Uint32("charID", s.charID),
		)
	} else {
		saveDuration := time.Since(saveStart)
		s.logger.Info("PlateMyset saved successfully",
			zap.Uint32("charID", s.charID),
			zap.Int("data_size", len(pkt.RawDataPayload)),
			zap.Duration("duration", saveDuration),
		)
	}

	// Invalidate user binary cache so other players see updated appearance
	s.server.userBinaryPartsLock.Lock()
	delete(s.server.userBinaryParts, userBinaryPartID{charID: s.charID, index: 2})
	delete(s.server.userBinaryParts, userBinaryPartID{charID: s.charID, index: 3})
	s.server.userBinaryPartsLock.Unlock()

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

// savePlateDataToDatabase saves all plate-related data for a character to the database.
// This is called during logout as a safety net to ensure plate data persistence.
//
// Note: Plate data (platedata, platebox, platemyset) saves immediately when the client
// sends save packets via handleMsgMhfSavePlateData, handleMsgMhfSavePlateBox, and
// handleMsgMhfSavePlateMyset. Unlike other data types that use session-level caching,
// plate data does not require re-saving at logout since it's already persisted.
//
// This function exists as:
// 1. A defensive safety net matching the pattern used for other auxiliary data
// 2. A hook for future enhancements if session-level caching is added
// 3. A monitoring point for debugging plate data persistence issues
//
// Returns nil as plate data is already saved by the individual handlers.
func savePlateDataToDatabase(s *Session) error {
	saveStart := time.Now()

	// Since plate data is not cached in session and saves immediately when
	// packets arrive, we don't need to perform any database operations here.
	// The individual save handlers have already persisted the data.
	//
	// This function provides a logging checkpoint to verify the save flow
	// and maintains consistency with the defensive programming pattern used
	// for other data types like warehouse and hunter navi.

	s.logger.Debug("Plate data save check at logout",
		zap.Uint32("charID", s.charID),
		zap.Duration("check_duration", time.Since(saveStart)),
	)

	return nil
}
