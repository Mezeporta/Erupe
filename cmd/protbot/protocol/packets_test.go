package protocol

import (
	"encoding/binary"
	"testing"

	"erupe-ce/common/byteframe"
)

// TestBuildLoginPacket verifies that the binary layout matches Erupe's Parse.
func TestBuildLoginPacket(t *testing.T) {
	ackHandle := uint32(1)
	charID := uint32(100)
	tokenNumber := uint32(42)
	tokenString := "0123456789ABCDEF"

	pkt := BuildLoginPacket(ackHandle, charID, tokenNumber, tokenString)

	bf := byteframe.NewByteFrameFromBytes(pkt)

	opcode := bf.ReadUint16()
	if opcode != MSG_SYS_LOGIN {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", opcode, MSG_SYS_LOGIN)
	}

	gotAck := bf.ReadUint32()
	if gotAck != ackHandle {
		t.Fatalf("ackHandle: got %d, want %d", gotAck, ackHandle)
	}

	gotCharID0 := bf.ReadUint32()
	if gotCharID0 != charID {
		t.Fatalf("charID0: got %d, want %d", gotCharID0, charID)
	}

	gotTokenNum := bf.ReadUint32()
	if gotTokenNum != tokenNumber {
		t.Fatalf("tokenNumber: got %d, want %d", gotTokenNum, tokenNumber)
	}

	gotZero := bf.ReadUint16()
	if gotZero != 0 {
		t.Fatalf("hardcodedZero: got %d, want 0", gotZero)
	}

	gotVersion := bf.ReadUint16()
	if gotVersion != 0xCAFE {
		t.Fatalf("requestVersion: got 0x%04X, want 0xCAFE", gotVersion)
	}

	gotCharID1 := bf.ReadUint32()
	if gotCharID1 != charID {
		t.Fatalf("charID1: got %d, want %d", gotCharID1, charID)
	}

	gotZeroed := bf.ReadUint16()
	if gotZeroed != 0 {
		t.Fatalf("zeroed: got %d, want 0", gotZeroed)
	}

	gotEleven := bf.ReadUint16()
	if gotEleven != 11 {
		t.Fatalf("always11: got %d, want 11", gotEleven)
	}

	gotToken := string(bf.ReadNullTerminatedBytes())
	if gotToken != tokenString {
		t.Fatalf("tokenString: got %q, want %q", gotToken, tokenString)
	}

	// Verify terminator.
	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildEnumerateStagePacket verifies binary layout matches Erupe's Parse.
func TestBuildEnumerateStagePacket(t *testing.T) {
	ackHandle := uint32(5)
	prefix := "sl1Ns"

	pkt := BuildEnumerateStagePacket(ackHandle, prefix)
	bf := byteframe.NewByteFrameFromBytes(pkt)

	opcode := bf.ReadUint16()
	if opcode != MSG_SYS_ENUMERATE_STAGE {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", opcode, MSG_SYS_ENUMERATE_STAGE)
	}

	gotAck := bf.ReadUint32()
	if gotAck != ackHandle {
		t.Fatalf("ackHandle: got %d, want %d", gotAck, ackHandle)
	}

	alwaysOne := bf.ReadUint8()
	if alwaysOne != 1 {
		t.Fatalf("alwaysOne: got %d, want 1", alwaysOne)
	}

	prefixLen := bf.ReadUint8()
	if prefixLen != uint8(len(prefix)+1) {
		t.Fatalf("prefixLen: got %d, want %d", prefixLen, len(prefix)+1)
	}

	gotPrefix := string(bf.ReadNullTerminatedBytes())
	if gotPrefix != prefix {
		t.Fatalf("prefix: got %q, want %q", gotPrefix, prefix)
	}

	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildEnterStagePacket verifies binary layout matches Erupe's Parse.
func TestBuildEnterStagePacket(t *testing.T) {
	ackHandle := uint32(7)
	stageID := "sl1Ns200p0a0u0"

	pkt := BuildEnterStagePacket(ackHandle, stageID)
	bf := byteframe.NewByteFrameFromBytes(pkt)

	opcode := bf.ReadUint16()
	if opcode != MSG_SYS_ENTER_STAGE {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", opcode, MSG_SYS_ENTER_STAGE)
	}

	gotAck := bf.ReadUint32()
	if gotAck != ackHandle {
		t.Fatalf("ackHandle: got %d, want %d", gotAck, ackHandle)
	}

	isQuest := bf.ReadUint8()
	if isQuest != 0 {
		t.Fatalf("isQuest: got %d, want 0", isQuest)
	}

	stageLen := bf.ReadUint8()
	if stageLen != uint8(len(stageID)+1) {
		t.Fatalf("stageLen: got %d, want %d", stageLen, len(stageID)+1)
	}

	gotStage := string(bf.ReadNullTerminatedBytes())
	if gotStage != stageID {
		t.Fatalf("stageID: got %q, want %q", gotStage, stageID)
	}

	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildPingPacket verifies MSG_SYS_PING binary layout.
func TestBuildPingPacket(t *testing.T) {
	ackHandle := uint32(99)
	pkt := BuildPingPacket(ackHandle)
	bf := byteframe.NewByteFrameFromBytes(pkt)

	if op := bf.ReadUint16(); op != MSG_SYS_PING {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", op, MSG_SYS_PING)
	}
	if ack := bf.ReadUint32(); ack != ackHandle {
		t.Fatalf("ackHandle: got %d, want %d", ack, ackHandle)
	}
	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildLogoutPacket verifies MSG_SYS_LOGOUT binary layout.
func TestBuildLogoutPacket(t *testing.T) {
	pkt := BuildLogoutPacket()
	bf := byteframe.NewByteFrameFromBytes(pkt)

	if op := bf.ReadUint16(); op != MSG_SYS_LOGOUT {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", op, MSG_SYS_LOGOUT)
	}
	if lt := bf.ReadUint8(); lt != 1 {
		t.Fatalf("logoutType: got %d, want 1", lt)
	}
	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildIssueLogkeyPacket verifies MSG_SYS_ISSUE_LOGKEY binary layout.
func TestBuildIssueLogkeyPacket(t *testing.T) {
	ackHandle := uint32(10)
	pkt := BuildIssueLogkeyPacket(ackHandle)
	bf := byteframe.NewByteFrameFromBytes(pkt)

	if op := bf.ReadUint16(); op != MSG_SYS_ISSUE_LOGKEY {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", op, MSG_SYS_ISSUE_LOGKEY)
	}
	if ack := bf.ReadUint32(); ack != ackHandle {
		t.Fatalf("ackHandle: got %d, want %d", ack, ackHandle)
	}
	if v := bf.ReadUint16(); v != 0 {
		t.Fatalf("unk0: got %d, want 0", v)
	}
	if v := bf.ReadUint16(); v != 0 {
		t.Fatalf("unk1: got %d, want 0", v)
	}
	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildRightsReloadPacket verifies MSG_SYS_RIGHTS_RELOAD binary layout.
func TestBuildRightsReloadPacket(t *testing.T) {
	ackHandle := uint32(20)
	pkt := BuildRightsReloadPacket(ackHandle)
	bf := byteframe.NewByteFrameFromBytes(pkt)

	if op := bf.ReadUint16(); op != MSG_SYS_RIGHTS_RELOAD {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", op, MSG_SYS_RIGHTS_RELOAD)
	}
	if ack := bf.ReadUint32(); ack != ackHandle {
		t.Fatalf("ackHandle: got %d, want %d", ack, ackHandle)
	}
	if c := bf.ReadUint8(); c != 0 {
		t.Fatalf("count: got %d, want 0", c)
	}
	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildLoaddataPacket verifies MSG_MHF_LOADDATA binary layout.
func TestBuildLoaddataPacket(t *testing.T) {
	ackHandle := uint32(30)
	pkt := BuildLoaddataPacket(ackHandle)
	bf := byteframe.NewByteFrameFromBytes(pkt)

	if op := bf.ReadUint16(); op != MSG_MHF_LOADDATA {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", op, MSG_MHF_LOADDATA)
	}
	if ack := bf.ReadUint32(); ack != ackHandle {
		t.Fatalf("ackHandle: got %d, want %d", ack, ackHandle)
	}
	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildCastBinaryPacket verifies MSG_SYS_CAST_BINARY binary layout.
func TestBuildCastBinaryPacket(t *testing.T) {
	payload := []byte{0xDE, 0xAD, 0xBE, 0xEF}
	pkt := BuildCastBinaryPacket(0x03, 1, payload)
	bf := byteframe.NewByteFrameFromBytes(pkt)

	if op := bf.ReadUint16(); op != MSG_SYS_CAST_BINARY {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", op, MSG_SYS_CAST_BINARY)
	}
	if unk := bf.ReadUint32(); unk != 0 {
		t.Fatalf("unk: got %d, want 0", unk)
	}
	if bt := bf.ReadUint8(); bt != 0x03 {
		t.Fatalf("broadcastType: got %d, want 3", bt)
	}
	if mt := bf.ReadUint8(); mt != 1 {
		t.Fatalf("messageType: got %d, want 1", mt)
	}
	if ds := bf.ReadUint16(); ds != uint16(len(payload)) {
		t.Fatalf("dataSize: got %d, want %d", ds, len(payload))
	}
	gotPayload := bf.ReadBytes(uint(len(payload)))
	for i, b := range payload {
		if gotPayload[i] != b {
			t.Fatalf("payload[%d]: got 0x%02X, want 0x%02X", i, gotPayload[i], b)
		}
	}
	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildChatPayload verifies the MsgBinChat inner binary layout and SJIS encoding.
func TestBuildChatPayload(t *testing.T) {
	chatType := uint8(1)
	message := "Hello"
	senderName := "TestUser"

	payload := BuildChatPayload(chatType, message, senderName)
	bf := byteframe.NewByteFrameFromBytes(payload)

	if unk := bf.ReadUint8(); unk != 0 {
		t.Fatalf("unk0: got %d, want 0", unk)
	}
	if ct := bf.ReadUint8(); ct != chatType {
		t.Fatalf("chatType: got %d, want %d", ct, chatType)
	}
	if flags := bf.ReadUint16(); flags != 0 {
		t.Fatalf("flags: got %d, want 0", flags)
	}
	nameLen := bf.ReadUint16()
	msgLen := bf.ReadUint16()
	// "Hello" in ASCII/SJIS = 5 bytes + 1 null = 6
	if msgLen != 6 {
		t.Fatalf("messageLen: got %d, want 6", msgLen)
	}
	// "TestUser" in ASCII/SJIS = 8 bytes + 1 null = 9
	if nameLen != 9 {
		t.Fatalf("senderNameLen: got %d, want 9", nameLen)
	}

	gotMsg := string(bf.ReadNullTerminatedBytes())
	if gotMsg != message {
		t.Fatalf("message: got %q, want %q", gotMsg, message)
	}
	gotName := string(bf.ReadNullTerminatedBytes())
	if gotName != senderName {
		t.Fatalf("senderName: got %q, want %q", gotName, senderName)
	}
}

// TestBuildEnumerateQuestPacket verifies MSG_MHF_ENUMERATE_QUEST binary layout.
func TestBuildEnumerateQuestPacket(t *testing.T) {
	ackHandle := uint32(40)
	world := uint8(2)
	counter := uint16(100)
	offset := uint16(50)

	pkt := BuildEnumerateQuestPacket(ackHandle, world, counter, offset)
	bf := byteframe.NewByteFrameFromBytes(pkt)

	if op := bf.ReadUint16(); op != MSG_MHF_ENUMERATE_QUEST {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", op, MSG_MHF_ENUMERATE_QUEST)
	}
	if ack := bf.ReadUint32(); ack != ackHandle {
		t.Fatalf("ackHandle: got %d, want %d", ack, ackHandle)
	}
	if u0 := bf.ReadUint8(); u0 != 0 {
		t.Fatalf("unk0: got %d, want 0", u0)
	}
	if w := bf.ReadUint8(); w != world {
		t.Fatalf("world: got %d, want %d", w, world)
	}
	if c := bf.ReadUint16(); c != counter {
		t.Fatalf("counter: got %d, want %d", c, counter)
	}
	if o := bf.ReadUint16(); o != offset {
		t.Fatalf("offset: got %d, want %d", o, offset)
	}
	if u1 := bf.ReadUint8(); u1 != 0 {
		t.Fatalf("unk1: got %d, want 0", u1)
	}
	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestBuildGetWeeklySchedulePacket verifies MSG_MHF_GET_WEEKLY_SCHEDULE binary layout.
func TestBuildGetWeeklySchedulePacket(t *testing.T) {
	ackHandle := uint32(50)
	pkt := BuildGetWeeklySchedulePacket(ackHandle)
	bf := byteframe.NewByteFrameFromBytes(pkt)

	if op := bf.ReadUint16(); op != MSG_MHF_GET_WEEKLY_SCHED {
		t.Fatalf("opcode: got 0x%04X, want 0x%04X", op, MSG_MHF_GET_WEEKLY_SCHED)
	}
	if ack := bf.ReadUint32(); ack != ackHandle {
		t.Fatalf("ackHandle: got %d, want %d", ack, ackHandle)
	}
	term := bf.ReadBytes(2)
	if term[0] != 0x00 || term[1] != 0x10 {
		t.Fatalf("terminator: got %02X %02X, want 00 10", term[0], term[1])
	}
}

// TestOpcodeValues verifies opcode constants match Erupe's iota-based enum.
func TestOpcodeValues(t *testing.T) {
	_ = binary.BigEndian // ensure import used
	tests := []struct {
		name string
		got  uint16
		want uint16
	}{
		{"MSG_SYS_ACK", MSG_SYS_ACK, 0x0012},
		{"MSG_SYS_LOGIN", MSG_SYS_LOGIN, 0x0014},
		{"MSG_SYS_LOGOUT", MSG_SYS_LOGOUT, 0x0015},
		{"MSG_SYS_PING", MSG_SYS_PING, 0x0017},
		{"MSG_SYS_CAST_BINARY", MSG_SYS_CAST_BINARY, 0x0018},
		{"MSG_SYS_TIME", MSG_SYS_TIME, 0x001A},
		{"MSG_SYS_CASTED_BINARY", MSG_SYS_CASTED_BINARY, 0x001B},
		{"MSG_SYS_ISSUE_LOGKEY", MSG_SYS_ISSUE_LOGKEY, 0x001D},
		{"MSG_SYS_ENTER_STAGE", MSG_SYS_ENTER_STAGE, 0x0022},
		{"MSG_SYS_ENUMERATE_STAGE", MSG_SYS_ENUMERATE_STAGE, 0x002F},
		{"MSG_SYS_INSERT_USER", MSG_SYS_INSERT_USER, 0x0050},
		{"MSG_SYS_DELETE_USER", MSG_SYS_DELETE_USER, 0x0051},
		{"MSG_SYS_UPDATE_RIGHT", MSG_SYS_UPDATE_RIGHT, 0x0058},
		{"MSG_SYS_RIGHTS_RELOAD", MSG_SYS_RIGHTS_RELOAD, 0x005D},
		{"MSG_MHF_LOADDATA", MSG_MHF_LOADDATA, 0x0061},
		{"MSG_MHF_ENUMERATE_QUEST", MSG_MHF_ENUMERATE_QUEST, 0x009F},
		{"MSG_MHF_GET_WEEKLY_SCHED", MSG_MHF_GET_WEEKLY_SCHED, 0x00E1},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s: got 0x%04X, want 0x%04X", tt.name, tt.got, tt.want)
		}
	}
}
