package channelserver

import (
	"encoding/binary"
	"erupe-ce/common/bfutil"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// User binary expected sizes and offsets (from mhfo-hd.dll RE).
// Types 4-5 are accepted by the server but never sent by the ZZ client.
const (
	userBinaryNameMaxSize = 17  // Type 1: SJIS null-terminated name
	userBinaryProfileSize = 208 // Type 2: 0xD0 — player profile
	userBinaryEquipSize   = 384 // Type 3: 0x180 — equipment/appearance

	// Type 2 profile offsets
	profileNameOff    = 0x0C // 25-byte SJIS name
	profileNameLen    = 25
	profileIntroOff   = 0x25 // 35-byte SJIS self-introduction
	profileIntroLen   = 35
	profileGuildIDOff = 0x48 // u32 guild ID

	// Type 3 equipment offsets
	equipHROff        = 0x00 // u16 HR (XOR'd with session key)
	equipWeaponOff    = 0x08 // 12-byte weapon entry
	equipHeadOff      = 0x18 // 12-byte head armor entry
	equipChestOff     = 0x24 // 12-byte chest armor entry
	equipArmsOff      = 0x30 // 12-byte arms armor entry
	equipWaistOff     = 0x3C // 12-byte waist armor entry
	equipLegsOff      = 0x48 // 12-byte legs armor entry
	equipGuildIDOff   = 0x64 // u32 guild ID
	equipGenderOff    = 0x68 // u8 gender flag
	equipSharpnessOff = 0x69 // u8 sharpness level
	equipEntrySize    = 12   // Each equipment entry: 3x u32
)

func handleMsgSysInsertUser(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgSysDeleteUser(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented

func handleMsgSysSetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysSetUserBinary)
	if pkt.BinaryType < 1 || pkt.BinaryType > 5 {
		s.logger.Warn("Invalid BinaryType", zap.Uint8("type", pkt.BinaryType))
		return
	}

	logUserBinaryFields(s, pkt.BinaryType, pkt.RawDataPayload)

	s.server.userBinary.Set(s.charID, pkt.BinaryType, pkt.RawDataPayload)

	s.server.BroadcastMHF(&mhfpacket.MsgSysNotifyUserBinary{
		CharID:     s.charID,
		BinaryType: pkt.BinaryType,
	}, s)
}

// logUserBinaryFields parses and logs the structured fields of a user binary
// payload based on its type. Logs a warning if the payload size does not match
// the expected format from the client RE.
func logUserBinaryFields(s *Session, binaryType uint8, data []byte) {
	switch binaryType {
	case 1:
		logUserBinaryName(s, data)
	case 2:
		logUserBinaryProfile(s, data)
	case 3:
		logUserBinaryEquipment(s, data)
	default:
		s.logger.Info("User binary received (unknown type)",
			zap.Uint8("type", binaryType),
			zap.Int("size", len(data)),
			zap.Uint32("charID", s.charID),
		)
	}
}

// logUserBinaryName parses type 1: character name (SJIS, null-terminated).
func logUserBinaryName(s *Session, data []byte) {
	if len(data) == 0 {
		s.logger.Warn("User binary type 1 (name): empty payload",
			zap.Uint32("charID", s.charID),
		)
		return
	}
	if len(data) > userBinaryNameMaxSize {
		s.logger.Warn("User binary type 1 (name): payload exceeds expected max",
			zap.Int("size", len(data)),
			zap.Int("expected_max", userBinaryNameMaxSize),
			zap.Uint32("charID", s.charID),
		)
	}
	name := stringsupport.SJISToUTF8Lossy(bfutil.UpToNull(data))
	s.logger.Info("User binary type 1 (name)",
		zap.String("name", name),
		zap.Int("size", len(data)),
		zap.Uint32("charID", s.charID),
	)
}

// logUserBinaryProfile parses type 2: player profile (208 bytes).
func logUserBinaryProfile(s *Session, data []byte) {
	if len(data) != userBinaryProfileSize {
		s.logger.Warn("User binary type 2 (profile): unexpected size",
			zap.Int("size", len(data)),
			zap.Int("expected", userBinaryProfileSize),
			zap.Uint32("charID", s.charID),
		)
		return
	}
	nameBytes := bfutil.UpToNull(data[profileNameOff : profileNameOff+profileNameLen])
	name := stringsupport.SJISToUTF8Lossy(nameBytes)

	introBytes := bfutil.UpToNull(data[profileIntroOff : profileIntroOff+profileIntroLen])
	intro := stringsupport.SJISToUTF8Lossy(introBytes)

	guildID := binary.BigEndian.Uint32(data[profileGuildIDOff : profileGuildIDOff+4])

	s.logger.Info("User binary type 2 (profile)",
		zap.String("name", name),
		zap.String("self_intro", intro),
		zap.Uint32("guild_id", guildID),
		zap.Int("size", len(data)),
		zap.Uint32("charID", s.charID),
	)
}

// logUserBinaryEquipment parses type 3: equipment/appearance (384 bytes).
func logUserBinaryEquipment(s *Session, data []byte) {
	if len(data) != userBinaryEquipSize {
		s.logger.Warn("User binary type 3 (equipment): unexpected size",
			zap.Int("size", len(data)),
			zap.Int("expected", userBinaryEquipSize),
			zap.Uint32("charID", s.charID),
		)
		return
	}
	hr := binary.BigEndian.Uint16(data[equipHROff : equipHROff+2])
	guildID := binary.BigEndian.Uint32(data[equipGuildIDOff : equipGuildIDOff+4])
	gender := data[equipGenderOff]
	sharpness := data[equipSharpnessOff]

	s.logger.Info("User binary type 3 (equipment)",
		zap.Uint16("hr_xored", hr),
		zap.Uint32("guild_id", guildID),
		zap.Uint8("gender", gender),
		zap.Uint8("sharpness", sharpness),
		zap.Int("size", len(data)),
		zap.Uint32("charID", s.charID),
	)
}

func handleMsgSysGetUserBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysGetUserBinary)

	data, ok := s.server.userBinary.Get(pkt.CharID, pkt.BinaryType)

	if !ok {
		doAckBufFail(s, pkt.AckHandle, make([]byte, 4))
	} else {
		doAckBufSucceed(s, pkt.AckHandle, data)
	}
}

func handleMsgSysNotifyUserBinary(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented
