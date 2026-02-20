package protocol

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
)

// BuildLoginPacket builds a MSG_SYS_LOGIN packet.
// Layout mirrors Erupe's MsgSysLogin.Parse:
//
//	uint16 opcode
//	uint32 ackHandle
//	uint32 charID
//	uint32 loginTokenNumber
//	uint16 hardcodedZero
//	uint16 requestVersion (set to 0xCAFE as dummy)
//	uint32 charID (repeated)
//	uint16 zeroed
//	uint16 always 11
//	null-terminated tokenString
//	0x00 0x10 terminator
func BuildLoginPacket(ackHandle, charID, tokenNumber uint32, tokenString string) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_SYS_LOGIN)
	bf.WriteUint32(ackHandle)
	bf.WriteUint32(charID)
	bf.WriteUint32(tokenNumber)
	bf.WriteUint16(0)      // HardcodedZero0
	bf.WriteUint16(0xCAFE) // RequestVersion (dummy)
	bf.WriteUint32(charID) // CharID1 (repeated)
	bf.WriteUint16(0)      // Zeroed
	bf.WriteUint16(11)     // Always 11
	bf.WriteNullTerminatedBytes([]byte(tokenString))
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildEnumerateStagePacket builds a MSG_SYS_ENUMERATE_STAGE packet.
// Layout mirrors Erupe's MsgSysEnumerateStage.Parse:
//
//	uint16 opcode
//	uint32 ackHandle
//	uint8  always 1
//	uint8  prefix length (including null terminator)
//	null-terminated stagePrefix
//	0x00 0x10 terminator
func BuildEnumerateStagePacket(ackHandle uint32, prefix string) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_SYS_ENUMERATE_STAGE)
	bf.WriteUint32(ackHandle)
	bf.WriteUint8(1)                         // Always 1
	bf.WriteUint8(uint8(len(prefix) + 1))    // Length including null terminator
	bf.WriteNullTerminatedBytes([]byte(prefix))
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildEnterStagePacket builds a MSG_SYS_ENTER_STAGE packet.
// Layout mirrors Erupe's MsgSysEnterStage.Parse:
//
//	uint16 opcode
//	uint32 ackHandle
//	uint8  isQuest (0=false)
//	uint8  stageID length (including null terminator)
//	null-terminated stageID
//	0x00 0x10 terminator
func BuildEnterStagePacket(ackHandle uint32, stageID string) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_SYS_ENTER_STAGE)
	bf.WriteUint32(ackHandle)
	bf.WriteUint8(0)                          // IsQuest = false
	bf.WriteUint8(uint8(len(stageID) + 1))    // Length including null terminator
	bf.WriteNullTerminatedBytes([]byte(stageID))
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildPingPacket builds a MSG_SYS_PING response packet.
//
//	uint16 opcode
//	uint32 ackHandle
//	0x00 0x10 terminator
func BuildPingPacket(ackHandle uint32) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_SYS_PING)
	bf.WriteUint32(ackHandle)
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildLogoutPacket builds a MSG_SYS_LOGOUT packet.
//
//	uint16 opcode
//	uint8  logoutType (1 = normal logout)
//	0x00 0x10 terminator
func BuildLogoutPacket() []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_SYS_LOGOUT)
	bf.WriteUint8(1) // LogoutType = normal
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildIssueLogkeyPacket builds a MSG_SYS_ISSUE_LOGKEY packet.
//
//	uint16 opcode
//	uint32 ackHandle
//	uint16 unk0
//	uint16 unk1
//	0x00 0x10 terminator
func BuildIssueLogkeyPacket(ackHandle uint32) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_SYS_ISSUE_LOGKEY)
	bf.WriteUint32(ackHandle)
	bf.WriteUint16(0)
	bf.WriteUint16(0)
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildRightsReloadPacket builds a MSG_SYS_RIGHTS_RELOAD packet.
//
//	uint16 opcode
//	uint32 ackHandle
//	uint8  count (0 = empty)
//	0x00 0x10 terminator
func BuildRightsReloadPacket(ackHandle uint32) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_SYS_RIGHTS_RELOAD)
	bf.WriteUint32(ackHandle)
	bf.WriteUint8(0) // Count = 0 (no rights entries)
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildLoaddataPacket builds a MSG_MHF_LOADDATA packet.
//
//	uint16 opcode
//	uint32 ackHandle
//	0x00 0x10 terminator
func BuildLoaddataPacket(ackHandle uint32) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_MHF_LOADDATA)
	bf.WriteUint32(ackHandle)
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildCastBinaryPacket builds a MSG_SYS_CAST_BINARY packet.
// Layout mirrors Erupe's MsgSysCastBinary.Parse:
//
//	uint16 opcode
//	uint32 unk (always 0)
//	uint8  broadcastType
//	uint8  messageType
//	uint16 dataSize
//	[]byte payload
//	0x00 0x10 terminator
func BuildCastBinaryPacket(broadcastType, messageType uint8, payload []byte) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_SYS_CAST_BINARY)
	bf.WriteUint32(0) // Unk
	bf.WriteUint8(broadcastType)
	bf.WriteUint8(messageType)
	bf.WriteUint16(uint16(len(payload)))
	bf.WriteBytes(payload)
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildChatPayload builds the inner MsgBinChat binary blob for use with BuildCastBinaryPacket.
// Layout mirrors Erupe's binpacket/msg_bin_chat.go Build:
//
//	uint8  unk0 (always 0)
//	uint8  chatType
//	uint16 flags (always 0)
//	uint16 senderNameLen (SJIS bytes + null terminator)
//	uint16 messageLen (SJIS bytes + null terminator)
//	null-terminated SJIS message
//	null-terminated SJIS senderName
func BuildChatPayload(chatType uint8, message, senderName string) []byte {
	sjisMsg := stringsupport.UTF8ToSJIS(message)
	sjisName := stringsupport.UTF8ToSJIS(senderName)
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0)                              // Unk0
	bf.WriteUint8(chatType)                       // Type
	bf.WriteUint16(0)                             // Flags
	bf.WriteUint16(uint16(len(sjisName) + 1))     // SenderName length (+ null term)
	bf.WriteUint16(uint16(len(sjisMsg) + 1))      // Message length (+ null term)
	bf.WriteNullTerminatedBytes(sjisMsg)           // Message
	bf.WriteNullTerminatedBytes(sjisName)          // SenderName
	return bf.Data()
}

// BuildEnumerateQuestPacket builds a MSG_MHF_ENUMERATE_QUEST packet.
//
//	uint16 opcode
//	uint32 ackHandle
//	uint8  unk0 (always 0)
//	uint8  world
//	uint16 counter
//	uint16 offset
//	uint8  unk1 (always 0)
//	0x00 0x10 terminator
func BuildEnumerateQuestPacket(ackHandle uint32, world uint8, counter, offset uint16) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_MHF_ENUMERATE_QUEST)
	bf.WriteUint32(ackHandle)
	bf.WriteUint8(0) // Unk0
	bf.WriteUint8(world)
	bf.WriteUint16(counter)
	bf.WriteUint16(offset)
	bf.WriteUint8(0) // Unk1
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}

// BuildGetWeeklySchedulePacket builds a MSG_MHF_GET_WEEKLY_SCHEDULE packet.
//
//	uint16 opcode
//	uint32 ackHandle
//	0x00 0x10 terminator
func BuildGetWeeklySchedulePacket(ackHandle uint32) []byte {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(MSG_MHF_GET_WEEKLY_SCHED)
	bf.WriteUint32(ackHandle)
	bf.WriteBytes([]byte{0x00, 0x10})
	return bf.Data()
}
