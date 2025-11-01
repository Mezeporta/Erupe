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
	var data []byte
	err := s.server.db.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Error("Failed to load platedata", zap.Error(err))
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSavePlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateData)
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
			data = make([]byte, 140000)
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
	var data []byte
	err := s.server.db.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Error("Failed to load platebox", zap.Error(err))
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSavePlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateBox)

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
			data = make([]byte, 4800)
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
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateMyset)
	var data []byte
	err := s.server.db.QueryRow("SELECT platemyset FROM characters WHERE id = $1", s.charID).Scan(&data)
	if len(data) == 0 {
		s.logger.Error("Failed to load platemyset", zap.Error(err))
		data = make([]byte, 1920)
	}
	doAckBufSucceed(s, pkt.AckHandle, data)
}

func handleMsgMhfSavePlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateMyset)
	// looks to always return the full thing, simply update database, no extra processing
	dumpSaveData(s, pkt.RawDataPayload, "platemyset")
	_, err := s.server.db.Exec("UPDATE characters SET platemyset=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Error("Failed to save platemyset", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}
