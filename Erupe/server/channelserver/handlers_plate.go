package channelserver

import (
	"github.com/Solenataris/Erupe/network/mhfpacket"
	"github.com/Solenataris/Erupe/server/channelserver/compression/deltacomp"
	"github.com/Solenataris/Erupe/server/channelserver/compression/nullcomp"
	"go.uber.org/zap"
)

func handleMsgMhfLoadPlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateData)
	var data []byte
	err := s.server.db.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get plate data savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		doAckBufSucceed(s, pkt.AckHandle, []byte{})
	}
}

func handleMsgMhfSavePlateData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateData)

	dumpSaveData(s, pkt.RawDataPayload, "_platedata")

	if pkt.IsDataDiff {
		var data []byte

		// Load existing save
		err := s.server.db.QueryRow("SELECT platedata FROM characters WHERE id = $1", s.charID).Scan(&data)
		if err != nil {
			s.logger.Fatal("Failed to get platedata savedata from db", zap.Error(err))
		}

		if len(data) > 0 {
			// Decompress
			s.logger.Info("Decompressing...")
			data, err = nullcomp.Decompress(data)
			if err != nil {
				s.logger.Fatal("Failed to decompress savedata from db", zap.Error(err))
			}
		} else {
			// create empty save if absent
			data = make([]byte, 0x1AF20)
		}

		// Perform diff and compress it to write back to db
		s.logger.Info("Diffing...")
		saveOutput, err := nullcomp.Compress(deltacomp.ApplyDataDiff(pkt.RawDataPayload, data))
		if err != nil {
			s.logger.Fatal("Failed to diff and compress platedata savedata", zap.Error(err))
		}

		_, err = s.server.db.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", saveOutput, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update platedata savedata in db", zap.Error(err))
		}

		s.logger.Info("Wrote recompressed platedata back to DB.")
	} else {
		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET platedata=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update platedata savedata in db", zap.Error(err))
		}
	}

	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateBox)
	var data []byte
	err := s.server.db.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get sigil box savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		doAckBufSucceed(s, pkt.AckHandle, []byte{})
	}
}

func handleMsgMhfSavePlateBox(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateBox)

	dumpSaveData(s, pkt.RawDataPayload, "_platebox")

	if pkt.IsDataDiff {
		var data []byte

		// Load existing save
		err := s.server.db.QueryRow("SELECT platebox FROM characters WHERE id = $1", s.charID).Scan(&data)
		if err != nil {
			s.logger.Fatal("Failed to get sigil box savedata from db", zap.Error(err))
		}

		// Decompress
		if len(data) > 0 {
			// Decompress
			s.logger.Info("Decompressing...")
			data, err = nullcomp.Decompress(data)
			if err != nil {
				s.logger.Fatal("Failed to decompress savedata from db", zap.Error(err))
			}
		} else {
			// create empty save if absent
			data = make([]byte, 0x820)
		}

		// Perform diff and compress it to write back to db
		s.logger.Info("Diffing...")
		saveOutput, err := nullcomp.Compress(deltacomp.ApplyDataDiff(pkt.RawDataPayload, data))
		if err != nil {
			s.logger.Fatal("Failed to diff and compress savedata", zap.Error(err))
		}

		_, err = s.server.db.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", saveOutput, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update platebox savedata in db", zap.Error(err))
		}

		s.logger.Info("Wrote recompressed platebox back to DB.")
	} else {
		// simply update database, no extra processing
		_, err := s.server.db.Exec("UPDATE characters SET platebox=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
		if err != nil {
			s.logger.Fatal("Failed to update platedata savedata in db", zap.Error(err))
		}
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfLoadPlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadPlateMyset)
	var data []byte
	err := s.server.db.QueryRow("SELECT platemyset FROM characters WHERE id = $1", s.charID).Scan(&data)
	if err != nil {
		s.logger.Fatal("Failed to get presets sigil savedata from db", zap.Error(err))
	}

	if len(data) > 0 {
		doAckBufSucceed(s, pkt.AckHandle, data)
	} else {
		blankData := make([]byte, 0x780)
		doAckBufSucceed(s, pkt.AckHandle, blankData)
	}
}

func handleMsgMhfSavePlateMyset(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSavePlateMyset)
	// looks to always return the full thing, simply update database, no extra processing

	_, err := s.server.db.Exec("UPDATE characters SET platemyset=$1 WHERE id=$2", pkt.RawDataPayload, s.charID)
	if err != nil {
		s.logger.Fatal("Failed to update platemyset savedata in db", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, []byte{0x00, 0x00, 0x00, 0x00})
}
