package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/clientctx"
)

// TestMsgMhfGetAchievementParse tests MsgMhfGetAchievement parsing
func TestMsgMhfGetAchievementDetailedParse(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x12345678) // AckHandle
	bf.WriteUint32(54321)      // CharID
	bf.WriteUint32(99999)      // Unk1
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfGetAchievement{}
	err := pkt.Parse(bf, &clientctx.ClientContext{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x12345678 {
		t.Errorf("AckHandle = 0x%X, want 0x12345678", pkt.AckHandle)
	}
	if pkt.CharID != 54321 {
		t.Errorf("CharID = %d, want 54321", pkt.CharID)
	}
}

// TestMsgMhfAddAchievementDetailedParse tests MsgMhfAddAchievement parsing
func TestMsgMhfAddAchievementDetailedParse(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(42)      // AchievementID
	bf.WriteUint16(12345)  // Unk1
	bf.WriteUint16(0xFFFF) // Unk2 - max value
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfAddAchievement{}
	err := pkt.Parse(bf, &clientctx.ClientContext{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AchievementID != 42 {
		t.Errorf("AchievementID = %d, want 42", pkt.AchievementID)
	}
	if pkt.Unk1 != 12345 {
		t.Errorf("Unk1 = %d, want 12345", pkt.Unk1)
	}
	if pkt.Unk2 != 0xFFFF {
		t.Errorf("Unk2 = %d, want 65535", pkt.Unk2)
	}
}

// TestMsgSysCastBinaryDetailedParse tests MsgSysCastBinary parsing with various payloads
func TestMsgSysCastBinaryDetailedParse(t *testing.T) {
	tests := []struct {
		name          string
		unk           uint32
		broadcastType uint8
		messageType   uint8
		payload       []byte
	}{
		{"empty payload", 0, 1, 2, []byte{}},
		{"typical payload", 0x006400C8, 0x10, 0x20, []byte{0x01, 0x02, 0x03}},
		{"chat message", 0, 0x01, 0x01, []byte("Hello, World!")},
		{"binary data", 0xFFFFFFFF, 0xFF, 0xFF, []byte{0xDE, 0xAD, 0xBE, 0xEF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.unk)
			bf.WriteUint8(tt.broadcastType)
			bf.WriteUint8(tt.messageType)
			bf.WriteUint16(uint16(len(tt.payload)))
			bf.WriteBytes(tt.payload)
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgSysCastBinary{}
			err := pkt.Parse(bf, &clientctx.ClientContext{})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.Unk != tt.unk {
				t.Errorf("Unk = %d, want %d", pkt.Unk, tt.unk)
			}
			if pkt.BroadcastType != tt.broadcastType {
				t.Errorf("BroadcastType = %d, want %d", pkt.BroadcastType, tt.broadcastType)
			}
			if pkt.MessageType != tt.messageType {
				t.Errorf("MessageType = %d, want %d", pkt.MessageType, tt.messageType)
			}
			if len(pkt.RawDataPayload) != len(tt.payload) {
				t.Errorf("RawDataPayload len = %d, want %d", len(pkt.RawDataPayload), len(tt.payload))
			}
		})
	}
}

// TestMsgSysLogoutParse tests MsgSysLogout parsing
func TestMsgSysLogoutDetailedParse(t *testing.T) {
	tests := []struct {
		unk0 uint8
	}{
		{0},
		{1},
		{100},
		{255},
	}

	for _, tt := range tests {
		bf := byteframe.NewByteFrame()
		bf.WriteUint8(tt.unk0)
		_, _ = bf.Seek(0, io.SeekStart)

		pkt := &MsgSysLogout{}
		err := pkt.Parse(bf, &clientctx.ClientContext{})
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if pkt.Unk0 != tt.unk0 {
			t.Errorf("Unk0 = %d, want %d", pkt.Unk0, tt.unk0)
		}
	}
}

// TestMsgSysBackStageParse tests MsgSysBackStage parsing
func TestMsgSysBackStageDetailedParse(t *testing.T) {
	tests := []struct {
		ackHandle uint32
	}{
		{0},
		{1},
		{0x12345678},
		{0xFFFFFFFF},
	}

	for _, tt := range tests {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(tt.ackHandle)
		_, _ = bf.Seek(0, io.SeekStart)

		pkt := &MsgSysBackStage{}
		err := pkt.Parse(bf, &clientctx.ClientContext{})
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if pkt.AckHandle != tt.ackHandle {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ackHandle)
		}
	}
}

// TestMsgSysPingParse tests MsgSysPing parsing
func TestMsgSysPingDetailedParse(t *testing.T) {
	tests := []struct {
		ackHandle uint32
	}{
		{0},
		{0xABCDEF12},
		{0xFFFFFFFF},
	}

	for _, tt := range tests {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(tt.ackHandle)
		_, _ = bf.Seek(0, io.SeekStart)

		pkt := &MsgSysPing{}
		err := pkt.Parse(bf, &clientctx.ClientContext{})
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if pkt.AckHandle != tt.ackHandle {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ackHandle)
		}
	}
}

// TestMsgSysTimeParse tests MsgSysTime parsing
func TestMsgSysTimeDetailedParse(t *testing.T) {
	tests := []struct {
		getRemoteTime bool
		timestamp     uint32
	}{
		{false, 0},
		{true, 1577836800}, // 2020-01-01 00:00:00
		{false, 0xFFFFFFFF},
	}

	for _, tt := range tests {
		bf := byteframe.NewByteFrame()
		bf.WriteBool(tt.getRemoteTime)
		bf.WriteUint32(tt.timestamp)
		_, _ = bf.Seek(0, io.SeekStart)

		pkt := &MsgSysTime{}
		err := pkt.Parse(bf, &clientctx.ClientContext{})
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if pkt.GetRemoteTime != tt.getRemoteTime {
			t.Errorf("GetRemoteTime = %v, want %v", pkt.GetRemoteTime, tt.getRemoteTime)
		}
		if pkt.Timestamp != tt.timestamp {
			t.Errorf("Timestamp = %d, want %d", pkt.Timestamp, tt.timestamp)
		}
	}
}
