package binpacket

import (
	"bytes"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
)

func TestMsgBinTargetedOpcode(t *testing.T) {
	m := &MsgBinTargeted{}
	if m.Opcode() != network.MSG_SYS_CAST_BINARY {
		t.Errorf("MsgBinTargeted.Opcode() = %v, want MSG_SYS_CAST_BINARY", m.Opcode())
	}
}

func TestMsgBinTargetedParseEmpty(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0) // TargetCount = 0

	_, _ = bf.Seek(0, 0)

	m := &MsgBinTargeted{}
	err := m.Parse(bf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if m.TargetCount != 0 {
		t.Errorf("TargetCount = %d, want 0", m.TargetCount)
	}
	if len(m.TargetCharIDs) != 0 {
		t.Errorf("TargetCharIDs len = %d, want 0", len(m.TargetCharIDs))
	}
}

func TestMsgBinTargetedParseSingleTarget(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(1)          // TargetCount = 1
	bf.WriteUint32(0x12345678) // TargetCharID
	bf.WriteBytes([]byte{0xDE, 0xAD, 0xBE, 0xEF})

	_, _ = bf.Seek(0, 0)

	m := &MsgBinTargeted{}
	err := m.Parse(bf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if m.TargetCount != 1 {
		t.Errorf("TargetCount = %d, want 1", m.TargetCount)
	}
	if len(m.TargetCharIDs) != 1 {
		t.Errorf("TargetCharIDs len = %d, want 1", len(m.TargetCharIDs))
	}
	if m.TargetCharIDs[0] != 0x12345678 {
		t.Errorf("TargetCharIDs[0] = %x, want 0x12345678", m.TargetCharIDs[0])
	}
	if !bytes.Equal(m.RawDataPayload, []byte{0xDE, 0xAD, 0xBE, 0xEF}) {
		t.Errorf("RawDataPayload = %v, want [0xDE, 0xAD, 0xBE, 0xEF]", m.RawDataPayload)
	}
}

func TestMsgBinTargetedParseMultipleTargets(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(3) // TargetCount = 3
	bf.WriteUint32(100)
	bf.WriteUint32(200)
	bf.WriteUint32(300)
	bf.WriteBytes([]byte{0x01, 0x02, 0x03})

	_, _ = bf.Seek(0, 0)

	m := &MsgBinTargeted{}
	err := m.Parse(bf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if m.TargetCount != 3 {
		t.Errorf("TargetCount = %d, want 3", m.TargetCount)
	}
	if len(m.TargetCharIDs) != 3 {
		t.Errorf("TargetCharIDs len = %d, want 3", len(m.TargetCharIDs))
	}
	if m.TargetCharIDs[0] != 100 || m.TargetCharIDs[1] != 200 || m.TargetCharIDs[2] != 300 {
		t.Errorf("TargetCharIDs = %v, want [100, 200, 300]", m.TargetCharIDs)
	}
}

func TestMsgBinTargetedBuild(t *testing.T) {
	m := &MsgBinTargeted{
		TargetCount:    2,
		TargetCharIDs:  []uint32{0x11111111, 0x22222222},
		RawDataPayload: []byte{0xAA, 0xBB},
	}

	bf := byteframe.NewByteFrame()
	err := m.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	expected := []byte{
		0x00, 0x02, // TargetCount
		0x11, 0x11, 0x11, 0x11, // TargetCharIDs[0]
		0x22, 0x22, 0x22, 0x22, // TargetCharIDs[1]
		0xAA, 0xBB, // RawDataPayload
	}

	if !bytes.Equal(bf.Data(), expected) {
		t.Errorf("Build() = %v, want %v", bf.Data(), expected)
	}
}

func TestMsgBinTargetedRoundTrip(t *testing.T) {
	original := &MsgBinTargeted{
		TargetCount:    3,
		TargetCharIDs:  []uint32{1000, 2000, 3000},
		RawDataPayload: []byte{0x01, 0x02, 0x03, 0x04, 0x05},
	}

	// Build
	bf := byteframe.NewByteFrame()
	err := original.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Parse
	_, _ = bf.Seek(0, 0)
	parsed := &MsgBinTargeted{}
	err = parsed.Parse(bf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Compare
	if parsed.TargetCount != original.TargetCount {
		t.Errorf("TargetCount = %d, want %d", parsed.TargetCount, original.TargetCount)
	}
	if len(parsed.TargetCharIDs) != len(original.TargetCharIDs) {
		t.Errorf("TargetCharIDs len = %d, want %d", len(parsed.TargetCharIDs), len(original.TargetCharIDs))
	}
	for i := range original.TargetCharIDs {
		if parsed.TargetCharIDs[i] != original.TargetCharIDs[i] {
			t.Errorf("TargetCharIDs[%d] = %d, want %d", i, parsed.TargetCharIDs[i], original.TargetCharIDs[i])
		}
	}
	if !bytes.Equal(parsed.RawDataPayload, original.RawDataPayload) {
		t.Errorf("RawDataPayload = %v, want %v", parsed.RawDataPayload, original.RawDataPayload)
	}
}

func TestMsgBinMailNotifyOpcode(t *testing.T) {
	m := MsgBinMailNotify{}
	if m.Opcode() != network.MSG_SYS_CASTED_BINARY {
		t.Errorf("MsgBinMailNotify.Opcode() = %v, want MSG_SYS_CASTED_BINARY", m.Opcode())
	}
}

func TestMsgBinMailNotifyBuild(t *testing.T) {
	m := MsgBinMailNotify{
		SenderName: "TestPlayer",
	}

	bf := byteframe.NewByteFrame()
	err := m.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	data := bf.Data()

	// First byte should be 0x01 (Unk)
	if data[0] != 0x01 {
		t.Errorf("First byte = %x, want 0x01", data[0])
	}

	// Total length should be 1 (Unk) + 21 (padded name) = 22
	if len(data) != 22 {
		t.Errorf("Data len = %d, want 22", len(data))
	}
}

func TestMsgBinMailNotifyBuildEmptyName(t *testing.T) {
	m := MsgBinMailNotify{
		SenderName: "",
	}

	bf := byteframe.NewByteFrame()
	err := m.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(bf.Data()) != 22 {
		t.Errorf("Data len = %d, want 22", len(bf.Data()))
	}
}

func TestMsgBinChatOpcode(t *testing.T) {
	m := &MsgBinChat{}
	if m.Opcode() != network.MSG_SYS_CAST_BINARY {
		t.Errorf("MsgBinChat.Opcode() = %v, want MSG_SYS_CAST_BINARY", m.Opcode())
	}
}

func TestMsgBinChatTypes(t *testing.T) {
	tests := []struct {
		chatType ChatType
		value    uint8
	}{
		{ChatTypeStage, 1},
		{ChatTypeGuild, 2},
		{ChatTypeAlliance, 3},
		{ChatTypeParty, 4},
		{ChatTypeWhisper, 5},
	}

	for _, tt := range tests {
		if uint8(tt.chatType) != tt.value {
			t.Errorf("ChatType %v = %d, want %d", tt.chatType, uint8(tt.chatType), tt.value)
		}
	}
}

func TestMsgBinChatBuildParse(t *testing.T) {
	original := &MsgBinChat{
		Unk0:       0x00,
		Type:       ChatTypeStage,
		Flags:      0x0000,
		Message:    "Hello",
		SenderName: "Player",
	}

	// Build
	bf := byteframe.NewByteFrame()
	err := original.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Parse
	_, _ = bf.Seek(0, 0)
	parsed := &MsgBinChat{}
	err = parsed.Parse(bf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Compare
	if parsed.Unk0 != original.Unk0 {
		t.Errorf("Unk0 = %d, want %d", parsed.Unk0, original.Unk0)
	}
	if parsed.Type != original.Type {
		t.Errorf("Type = %d, want %d", parsed.Type, original.Type)
	}
	if parsed.Flags != original.Flags {
		t.Errorf("Flags = %d, want %d", parsed.Flags, original.Flags)
	}
	if parsed.Message != original.Message {
		t.Errorf("Message = %q, want %q", parsed.Message, original.Message)
	}
	if parsed.SenderName != original.SenderName {
		t.Errorf("SenderName = %q, want %q", parsed.SenderName, original.SenderName)
	}
}

func TestMsgBinChatBuildParseJapanese(t *testing.T) {
	original := &MsgBinChat{
		Unk0:       0x00,
		Type:       ChatTypeGuild,
		Flags:      0x0001,
		Message:    "こんにちは",
		SenderName: "テスト",
	}

	// Build
	bf := byteframe.NewByteFrame()
	err := original.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Parse
	_, _ = bf.Seek(0, 0)
	parsed := &MsgBinChat{}
	err = parsed.Parse(bf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsed.Message != original.Message {
		t.Errorf("Message = %q, want %q", parsed.Message, original.Message)
	}
	if parsed.SenderName != original.SenderName {
		t.Errorf("SenderName = %q, want %q", parsed.SenderName, original.SenderName)
	}
}

func TestMsgBinChatBuildParseEmpty(t *testing.T) {
	original := &MsgBinChat{
		Unk0:       0x00,
		Type:       ChatTypeParty,
		Flags:      0x0000,
		Message:    "",
		SenderName: "",
	}

	// Build
	bf := byteframe.NewByteFrame()
	err := original.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Parse
	_, _ = bf.Seek(0, 0)
	parsed := &MsgBinChat{}
	err = parsed.Parse(bf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsed.Message != "" {
		t.Errorf("Message = %q, want empty", parsed.Message)
	}
	if parsed.SenderName != "" {
		t.Errorf("SenderName = %q, want empty", parsed.SenderName)
	}
}

func TestMsgBinChatBuildFormat(t *testing.T) {
	m := &MsgBinChat{
		Unk0:       0x12,
		Type:       ChatTypeWhisper,
		Flags:      0x3456,
		Message:    "Hi",
		SenderName: "A",
	}

	bf := byteframe.NewByteFrame()
	err := m.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	data := bf.Data()

	// Verify header structure
	if data[0] != 0x12 {
		t.Errorf("Unk0 = %x, want 0x12", data[0])
	}
	if data[1] != uint8(ChatTypeWhisper) {
		t.Errorf("Type = %x, want %x", data[1], uint8(ChatTypeWhisper))
	}
	// Flags at bytes 2-3 (big endian)
	if data[2] != 0x34 || data[3] != 0x56 {
		t.Errorf("Flags = %x%x, want 3456", data[2], data[3])
	}
}

func TestMsgBinChatAllTypes(t *testing.T) {
	types := []ChatType{
		ChatTypeStage,
		ChatTypeGuild,
		ChatTypeAlliance,
		ChatTypeParty,
		ChatTypeWhisper,
	}

	for _, chatType := range types {
		t.Run("", func(t *testing.T) {
			original := &MsgBinChat{
				Type:       chatType,
				Message:    "Test",
				SenderName: "Player",
			}

			bf := byteframe.NewByteFrame()
			err := original.Build(bf)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			_, _ = bf.Seek(0, 0)
			parsed := &MsgBinChat{}
			err = parsed.Parse(bf)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.Type != chatType {
				t.Errorf("Type = %d, want %d", parsed.Type, chatType)
			}
		})
	}
}

func TestMsgBinMailNotifyParseReturnsError(t *testing.T) {
	m := MsgBinMailNotify{}
	bf := byteframe.NewByteFrame()
	err := m.Parse(bf)
	if err == nil {
		t.Error("Parse() should return an error (not implemented)")
	}
}

func TestMsgBinMailNotifyBuildLongName(t *testing.T) {
	m := MsgBinMailNotify{
		SenderName: "ThisIsAVeryLongPlayerNameThatExceeds21Characters",
	}

	bf := byteframe.NewByteFrame()
	err := m.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Data should still be 22 bytes (1 + 21)
	if len(bf.Data()) != 22 {
		t.Errorf("Data len = %d, want 22", len(bf.Data()))
	}
}
