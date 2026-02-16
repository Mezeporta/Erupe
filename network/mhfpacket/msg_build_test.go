package mhfpacket

import (
	"bytes"
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/clientctx"
)

// TestBuildParseDuplicateObject verifies Build/Parse round-trip for MsgSysDuplicateObject.
// This packet carries object ID, 3D position (float32 x/y/z), and owner character ID.
func TestBuildParseDuplicateObject(t *testing.T) {
	tests := []struct {
		name        string
		objID       uint32
		x, y, z     float32
		unk0        uint32
		ownerCharID uint32
	}{
		{"typical values", 42, 1.5, 2.5, 3.5, 0, 12345},
		{"zero values", 0, 0, 0, 0, 0, 0},
		{"large values", 0xFFFFFFFF, -100.25, 200.75, -300.125, 0xDEADBEEF, 0xCAFEBABE},
		{"negative coords", 1, -1.0, -2.0, -3.0, 100, 200},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysDuplicateObject{
				ObjID:       tt.objID,
				X:           tt.x,
				Y:           tt.y,
				Z:           tt.z,
				Unk0:        tt.unk0,
				OwnerCharID: tt.ownerCharID,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysDuplicateObject{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.ObjID != original.ObjID {
				t.Errorf("ObjID = %d, want %d", parsed.ObjID, original.ObjID)
			}
			if parsed.X != original.X {
				t.Errorf("X = %f, want %f", parsed.X, original.X)
			}
			if parsed.Y != original.Y {
				t.Errorf("Y = %f, want %f", parsed.Y, original.Y)
			}
			if parsed.Z != original.Z {
				t.Errorf("Z = %f, want %f", parsed.Z, original.Z)
			}
			if parsed.Unk0 != original.Unk0 {
				t.Errorf("Unk0 = %d, want %d", parsed.Unk0, original.Unk0)
			}
			if parsed.OwnerCharID != original.OwnerCharID {
				t.Errorf("OwnerCharID = %d, want %d", parsed.OwnerCharID, original.OwnerCharID)
			}
		})
	}
}

// TestBuildParsePositionObject verifies Build/Parse round-trip for MsgSysPositionObject.
// This packet updates an object's 3D position (float32 x/y/z).
func TestBuildParsePositionObject(t *testing.T) {
	tests := []struct {
		name    string
		objID   uint32
		x, y, z float32
	}{
		{"origin", 1, 0, 0, 0},
		{"typical position", 100, 50.5, 75.25, -10.125},
		{"max object id", 0xFFFFFFFF, 999.999, -999.999, 0.001},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysPositionObject{
				ObjID: tt.objID,
				X:     tt.x,
				Y:     tt.y,
				Z:     tt.z,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysPositionObject{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.ObjID != original.ObjID {
				t.Errorf("ObjID = %d, want %d", parsed.ObjID, original.ObjID)
			}
			if parsed.X != original.X {
				t.Errorf("X = %f, want %f", parsed.X, original.X)
			}
			if parsed.Y != original.Y {
				t.Errorf("Y = %f, want %f", parsed.Y, original.Y)
			}
			if parsed.Z != original.Z {
				t.Errorf("Z = %f, want %f", parsed.Z, original.Z)
			}
		})
	}
}

// TestBuildParseCastedBinary verifies Build/Parse round-trip for MsgSysCastedBinary.
// This packet carries broadcast data with a length-prefixed payload.
func TestBuildParseCastedBinary(t *testing.T) {
	tests := []struct {
		name           string
		charID         uint32
		broadcastType  uint8
		messageType    uint8
		rawDataPayload []byte
	}{
		{"small payload", 12345, 1, 2, []byte{0xAA, 0xBB, 0xCC}},
		{"empty payload", 0, 0, 0, []byte{}},
		{"single byte payload", 0xDEADBEEF, 255, 128, []byte{0xFF}},
		{"larger payload", 42, 3, 4, []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A}},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysCastedBinary{
				CharID:         tt.charID,
				BroadcastType:  tt.broadcastType,
				MessageType:    tt.messageType,
				RawDataPayload: tt.rawDataPayload,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysCastedBinary{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.CharID != original.CharID {
				t.Errorf("CharID = %d, want %d", parsed.CharID, original.CharID)
			}
			if parsed.BroadcastType != original.BroadcastType {
				t.Errorf("BroadcastType = %d, want %d", parsed.BroadcastType, original.BroadcastType)
			}
			if parsed.MessageType != original.MessageType {
				t.Errorf("MessageType = %d, want %d", parsed.MessageType, original.MessageType)
			}
			if !bytes.Equal(parsed.RawDataPayload, original.RawDataPayload) {
				t.Errorf("RawDataPayload = %v, want %v", parsed.RawDataPayload, original.RawDataPayload)
			}
		})
	}
}

// TestBuildParseLoadRegister verifies manual-build/Parse round-trip for MsgSysLoadRegister.
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
func TestBuildParseLoadRegister(t *testing.T) {
	tests := []struct {
		name       string
		ackHandle  uint32
		registerID uint32
		values     uint8
	}{
		{"typical", 0x11223344, 100, 1},
		{"zero values", 0, 0, 0},
		{"max values", 0xFFFFFFFF, 0xFFFFFFFF, 255},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.registerID)
			bf.WriteUint8(tt.values)
			bf.WriteUint8(0)  // Zeroed
			bf.WriteUint16(0) // Zeroed

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysLoadRegister{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, tt.ackHandle)
			}
			if parsed.RegisterID != tt.registerID {
				t.Errorf("RegisterID = %d, want %d", parsed.RegisterID, tt.registerID)
			}
			if parsed.Values != tt.values {
				t.Errorf("Values = %d, want %d", parsed.Values, tt.values)
			}
		})
	}
}

// TestBuildParseOperateRegister verifies manual-build/Parse round-trip for MsgSysOperateRegister.
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
func TestBuildParseOperateRegister(t *testing.T) {
	tests := []struct {
		name        string
		ackHandle   uint32
		semaphoreID uint32
		payload     []byte
	}{
		{"typical", 1, 42, []byte{0x01, 0x02, 0x03}},
		{"empty payload", 0, 0, []byte{}},
		{"large payload", 0xFFFFFFFF, 0xDEADBEEF, make([]byte, 256)},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.semaphoreID)
			bf.WriteUint16(0) // Zeroed
			bf.WriteUint16(uint16(len(tt.payload)))
			bf.WriteBytes(tt.payload)

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysOperateRegister{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, tt.ackHandle)
			}
			if parsed.SemaphoreID != tt.semaphoreID {
				t.Errorf("SemaphoreID = %d, want %d", parsed.SemaphoreID, tt.semaphoreID)
			}
			if !bytes.Equal(parsed.RawDataPayload, tt.payload) {
				t.Errorf("RawDataPayload length = %d, want %d", len(parsed.RawDataPayload), len(tt.payload))
			}
		})
	}
}

// TestBuildParseNotifyUserBinary verifies Build/Parse round-trip for MsgSysNotifyUserBinary.
func TestBuildParseNotifyUserBinary(t *testing.T) {
	tests := []struct {
		name       string
		charID     uint32
		binaryType uint8
	}{
		{"typical", 12345, 1},
		{"zero", 0, 0},
		{"max", 0xFFFFFFFF, 255},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysNotifyUserBinary{
				CharID:     tt.charID,
				BinaryType: tt.binaryType,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysNotifyUserBinary{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.CharID != original.CharID {
				t.Errorf("CharID = %d, want %d", parsed.CharID, original.CharID)
			}
			if parsed.BinaryType != original.BinaryType {
				t.Errorf("BinaryType = %d, want %d", parsed.BinaryType, original.BinaryType)
			}
		})
	}
}

// TestBuildParseTime verifies Build/Parse round-trip for MsgSysTime.
// This packet carries a boolean flag and a Unix timestamp.
func TestBuildParseTime(t *testing.T) {
	tests := []struct {
		name          string
		getRemoteTime bool
		timestamp     uint32
	}{
		{"request remote time", true, 1577105879},
		{"no request", false, 0},
		{"max timestamp", true, 0xFFFFFFFF},
		{"typical timestamp", false, 1700000000},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysTime{
				GetRemoteTime: tt.getRemoteTime,
				Timestamp:     tt.timestamp,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysTime{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.GetRemoteTime != original.GetRemoteTime {
				t.Errorf("GetRemoteTime = %v, want %v", parsed.GetRemoteTime, original.GetRemoteTime)
			}
			if parsed.Timestamp != original.Timestamp {
				t.Errorf("Timestamp = %d, want %d", parsed.Timestamp, original.Timestamp)
			}
		})
	}
}

// TestBuildParseUpdateObjectBinary verifies Build/Parse round-trip for MsgSysUpdateObjectBinary.
func TestBuildParseUpdateObjectBinary(t *testing.T) {
	tests := []struct {
		name string
		unk0 uint32
		unk1 uint32
	}{
		{"typical", 42, 100},
		{"zero", 0, 0},
		{"max", 0xFFFFFFFF, 0xFFFFFFFF},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysUpdateObjectBinary{
				Unk0: tt.unk0,
				Unk1: tt.unk1,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysUpdateObjectBinary{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.Unk0 != original.Unk0 {
				t.Errorf("Unk0 = %d, want %d", parsed.Unk0, original.Unk0)
			}
			if parsed.Unk1 != original.Unk1 {
				t.Errorf("Unk1 = %d, want %d", parsed.Unk1, original.Unk1)
			}
		})
	}
}

// TestBuildParseArrangeGuildMember verifies manual-build/Parse round-trip for MsgMhfArrangeGuildMember.
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
// Parse reads: uint32 AckHandle, uint32 GuildID, uint8 zeroed, uint8 charCount, then charCount * uint32.
func TestBuildParseArrangeGuildMember(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		guildID   uint32
		charIDs   []uint32
	}{
		{"single member", 1, 100, []uint32{12345}},
		{"multiple members", 0x12345678, 200, []uint32{111, 222, 333, 444}},
		{"no members", 42, 300, []uint32{}},
		{"many members", 999, 400, []uint32{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.guildID)
			bf.WriteUint8(0) // Zeroed
			bf.WriteUint8(uint8(len(tt.charIDs)))
			for _, id := range tt.charIDs {
				bf.WriteUint32(id)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfArrangeGuildMember{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, tt.ackHandle)
			}
			if parsed.GuildID != tt.guildID {
				t.Errorf("GuildID = %d, want %d", parsed.GuildID, tt.guildID)
			}
			if len(parsed.CharIDs) != len(tt.charIDs) {
				t.Fatalf("CharIDs length = %d, want %d", len(parsed.CharIDs), len(tt.charIDs))
			}
			for i, id := range parsed.CharIDs {
				if id != tt.charIDs[i] {
					t.Errorf("CharIDs[%d] = %d, want %d", i, id, tt.charIDs[i])
				}
			}
		})
	}
}

// TestBuildParseEnumerateGuildMember verifies manual-build/Parse round-trip for MsgMhfEnumerateGuildMember.
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
// Parse reads: uint32 AckHandle, uint8 zeroed, uint8 always1, uint32 AllianceID, uint32 GuildID.
func TestBuildParseEnumerateGuildMember(t *testing.T) {
	tests := []struct {
		name       string
		ackHandle  uint32
		allianceID uint32
		guildID    uint32
	}{
		{"typical", 1, 0, 100},
		{"zero", 0, 0, 0},
		{"large values", 0xFFFFFFFF, 0xDEADBEEF, 0xCAFEBABE},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint8(0) // Zeroed
			bf.WriteUint8(1) // Always 1
			bf.WriteUint32(tt.allianceID)
			bf.WriteUint32(tt.guildID)

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfEnumerateGuildMember{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, tt.ackHandle)
			}
			if parsed.AllianceID != tt.allianceID {
				t.Errorf("AllianceID = %d, want %d", parsed.AllianceID, tt.allianceID)
			}
			if parsed.GuildID != tt.guildID {
				t.Errorf("GuildID = %d, want %d", parsed.GuildID, tt.guildID)
			}
		})
	}
}

// TestBuildParseStateCampaign verifies manual-build/Parse round-trip for MsgMhfStateCampaign.
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
func TestBuildParseStateCampaign(t *testing.T) {
	tests := []struct {
		name       string
		ackHandle  uint32
		campaignID uint32
		unk1       uint16
	}{
		{"typical", 1, 10, 300},
		{"zero", 0, 0, 0},
		{"max", 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFF},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.campaignID)
			bf.WriteUint16(tt.unk1)

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfStateCampaign{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, tt.ackHandle)
			}
			if parsed.CampaignID != tt.campaignID {
				t.Errorf("CampaignID = %d, want %d", parsed.CampaignID, tt.campaignID)
			}
			if parsed.Unk1 != tt.unk1 {
				t.Errorf("Unk1 = %d, want %d", parsed.Unk1, tt.unk1)
			}
		})
	}
}

// TestBuildParseApplyCampaign verifies manual-build/Parse round-trip for MsgMhfApplyCampaign.
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
func TestBuildParseApplyCampaign(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		unk0      uint32
		unk1      uint16
		unk2      []byte
	}{
		{"typical", 0x55667788, 5, 10, make([]byte, 16)},
		{"zero", 0, 0, 0, make([]byte, 16)},
		{"max", 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFF, make([]byte, 16)},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.unk0)
			bf.WriteUint16(tt.unk1)
			bf.WriteBytes(tt.unk2)

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfApplyCampaign{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, tt.ackHandle)
			}
			if parsed.Unk0 != tt.unk0 {
				t.Errorf("Unk0 = %d, want %d", parsed.Unk0, tt.unk0)
			}
			if parsed.Unk1 != tt.unk1 {
				t.Errorf("Unk1 = %d, want %d", parsed.Unk1, tt.unk1)
			}
			if len(parsed.Unk2) != len(tt.unk2) {
				t.Errorf("Unk2 len = %d, want %d", len(parsed.Unk2), len(tt.unk2))
			}
		})
	}
}

// TestBuildParseEnumerateCampaign verifies Build/Parse round-trip for MsgMhfEnumerateCampaign.
func TestBuildParseEnumerateCampaign(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		unk0      uint16
		unk1      uint16
	}{
		{"typical", 42, 1, 2},
		{"zero", 0, 0, 0},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgMhfEnumerateCampaign{
				AckHandle: tt.ackHandle,
				Unk0:      tt.unk0,
				Unk1:      tt.unk1,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfEnumerateCampaign{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != original.AckHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, original.AckHandle)
			}
			if parsed.Unk0 != original.Unk0 {
				t.Errorf("Unk0 = %d, want %d", parsed.Unk0, original.Unk0)
			}
			if parsed.Unk1 != original.Unk1 {
				t.Errorf("Unk1 = %d, want %d", parsed.Unk1, original.Unk1)
			}
		})
	}
}

// TestBuildParseEnumerateEvent verifies Build/Parse round-trip for MsgMhfEnumerateEvent.
func TestBuildParseEnumerateEvent(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
	}{
		{"typical", 0x11223344},
		{"nonzero", 42},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgMhfEnumerateEvent{
				AckHandle: tt.ackHandle,
			}

			bf := byteframe.NewByteFrame()
			// Build is NOT IMPLEMENTED; manually write the binary representation
			bf.WriteUint32(original.AckHandle)
			bf.WriteUint16(0) // Zeroed (discarded by Parse)
			bf.WriteUint16(0) // Zeroed (discarded by Parse)

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfEnumerateEvent{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != original.AckHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, original.AckHandle)
			}
		})
	}
}

// TestBuildParseAddUdTacticsPoint verifies Build/Parse round-trip for MsgMhfAddUdTacticsPoint.
func TestBuildParseAddUdTacticsPoint(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		unk0      uint16
		unk1      uint32
	}{
		{"typical", 1, 100, 50000},
		{"zero", 0, 0, 0},
		{"max", 0xFFFFFFFF, 0xFFFF, 0xFFFFFFFF},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgMhfAddUdTacticsPoint{
				AckHandle: tt.ackHandle,
				Unk0:      tt.unk0,
				Unk1:      tt.unk1,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfAddUdTacticsPoint{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != original.AckHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, original.AckHandle)
			}
			if parsed.Unk0 != original.Unk0 {
				t.Errorf("Unk0 = %d, want %d", parsed.Unk0, original.Unk0)
			}
			if parsed.Unk1 != original.Unk1 {
				t.Errorf("Unk1 = %d, want %d", parsed.Unk1, original.Unk1)
			}
		})
	}
}

// TestBuildParseApplyDistItem verifies manual-build/Parse round-trip for MsgMhfApplyDistItem.
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
// Note: Unk2 and Unk3 are conditionally parsed based on RealClientMode (G8+ and G10+).
// Default test config is ZZ, so both Unk2 and Unk3 are read.
func TestBuildParseApplyDistItem(t *testing.T) {
	tests := []struct {
		name             string
		ackHandle        uint32
		distributionType uint8
		distributionID   uint32
		unk2             uint32
		unk3             uint32
	}{
		{"typical", 0x12345678, 1, 42, 100, 200},
		{"zero", 0, 0, 0, 0, 0},
		{"max", 0xFFFFFFFF, 255, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint8(tt.distributionType)
			bf.WriteUint32(tt.distributionID)
			bf.WriteUint32(tt.unk2) // Read when RealClientMode >= G8
			bf.WriteUint32(tt.unk3) // Read when RealClientMode >= G10

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfApplyDistItem{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, tt.ackHandle)
			}
			if parsed.DistributionType != tt.distributionType {
				t.Errorf("DistributionType = %d, want %d", parsed.DistributionType, tt.distributionType)
			}
			if parsed.DistributionID != tt.distributionID {
				t.Errorf("DistributionID = %d, want %d", parsed.DistributionID, tt.distributionID)
			}
			if parsed.Unk2 != tt.unk2 {
				t.Errorf("Unk2 = %d, want %d", parsed.Unk2, tt.unk2)
			}
			if parsed.Unk3 != tt.unk3 {
				t.Errorf("Unk3 = %d, want %d", parsed.Unk3, tt.unk3)
			}
		})
	}
}

// TestBuildParseEnumerateDistItem verifies Build/Parse round-trip for MsgMhfEnumerateDistItem.
func TestBuildParseEnumerateDistItem(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		distType  uint8
		unk1      uint8
		unk2      uint16
	}{
		{"typical", 0xAABBCCDD, 5, 100, 200},
		{"zero", 0, 0, 0, 0},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgMhfEnumerateDistItem{
				AckHandle: tt.ackHandle,
				DistType:  tt.distType,
				Unk1:      tt.unk1,
				Unk2:      tt.unk2,
			}

			bf := byteframe.NewByteFrame()
			// Build is NOT IMPLEMENTED; manually write the binary representation
			bf.WriteUint32(original.AckHandle)
			bf.WriteUint8(original.DistType)
			bf.WriteUint8(original.Unk1)
			bf.WriteUint16(original.Unk2)
			bf.WriteUint8(0) // Unk3 length (for Z1+ client mode)

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfEnumerateDistItem{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != original.AckHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, original.AckHandle)
			}
			if parsed.DistType != original.DistType {
				t.Errorf("DistType = %d, want %d", parsed.DistType, original.DistType)
			}
			if parsed.Unk1 != original.Unk1 {
				t.Errorf("Unk1 = %d, want %d", parsed.Unk1, original.Unk1)
			}
			if parsed.Unk2 != original.Unk2 {
				t.Errorf("Unk2 = %d, want %d", parsed.Unk2, original.Unk2)
			}
		})
	}
}

// TestBuildParseAcquireExchangeShop verifies Build/Parse round-trip for MsgMhfAcquireExchangeShop.
// This packet has a separate DataSize field and a length-prefixed raw data payload.
func TestBuildParseAcquireExchangeShop(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		payload   []byte
	}{
		{"small payload", 1, []byte{0x01, 0x02, 0x03, 0x04}},
		{"empty payload", 0, []byte{}},
		{"larger payload", 0xDEADBEEF, []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x22}},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgMhfAcquireExchangeShop{
				AckHandle:      tt.ackHandle,
				DataSize:       uint16(len(tt.payload)),
				RawDataPayload: tt.payload,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfAcquireExchangeShop{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != original.AckHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, original.AckHandle)
			}
			if parsed.DataSize != original.DataSize {
				t.Errorf("DataSize = %d, want %d", parsed.DataSize, original.DataSize)
			}
			if !bytes.Equal(parsed.RawDataPayload, original.RawDataPayload) {
				t.Errorf("RawDataPayload = %v, want %v", parsed.RawDataPayload, original.RawDataPayload)
			}
		})
	}
}

// TestBuildParseDisplayedAchievement verifies Parse for MsgMhfDisplayedAchievement.
// This struct has no exported fields; Parse only discards a single zeroed byte.
func TestBuildParseDisplayedAchievement(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(0) // Zeroed (discarded by Parse)
	bf.Seek(0, io.SeekStart)

	parsed := &MsgMhfDisplayedAchievement{}
	if err := parsed.Parse(bf, ctx); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

// TestBuildParseAddKouryouPoint verifies Build/Parse round-trip for MsgMhfAddKouryouPoint.
func TestBuildParseAddKouryouPoint(t *testing.T) {
	tests := []struct {
		name          string
		ackHandle     uint32
		kouryouPoints uint32
	}{
		{"typical", 1, 5000},
		{"zero", 0, 0},
		{"max", 0xFFFFFFFF, 0xFFFFFFFF},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgMhfAddKouryouPoint{
				AckHandle:     tt.ackHandle,
				KouryouPoints: tt.kouryouPoints,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfAddKouryouPoint{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != original.AckHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, original.AckHandle)
			}
			if parsed.KouryouPoints != original.KouryouPoints {
				t.Errorf("KouryouPoints = %d, want %d", parsed.KouryouPoints, original.KouryouPoints)
			}
		})
	}
}

// TestBuildParseCheckDailyCafepoint verifies manual-build/Parse round-trip for MsgMhfCheckDailyCafepoint.
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
func TestBuildParseCheckDailyCafepoint(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		unk       uint32
	}{
		{"typical", 0x11223344, 100},
		{"zero", 0, 0},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.unk)

			bf.Seek(0, io.SeekStart)
			parsed := &MsgMhfCheckDailyCafepoint{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, tt.ackHandle)
			}
			if parsed.Unk != tt.unk {
				t.Errorf("Unk = %d, want %d", parsed.Unk, tt.unk)
			}
		})
	}
}

// TestBuildParsePing verifies Build/Parse round-trip for MsgSysPing.
func TestBuildParsePing(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
	}{
		{"typical", 0x12345678},
		{"zero", 0},
		{"max", 0xFFFFFFFF},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysPing{
				AckHandle: tt.ackHandle,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysPing{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != original.AckHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, original.AckHandle)
			}
		})
	}
}

// TestBuildParseDeleteObject verifies Build/Parse round-trip for MsgSysDeleteObject.
func TestBuildParseDeleteObject(t *testing.T) {
	tests := []struct {
		name  string
		objID uint32
	}{
		{"typical", 42},
		{"zero", 0},
		{"max", 0xFFFFFFFF},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysDeleteObject{
				ObjID: tt.objID,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysDeleteObject{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.ObjID != original.ObjID {
				t.Errorf("ObjID = %d, want %d", parsed.ObjID, original.ObjID)
			}
		})
	}
}

// TestBuildParseNotifyRegister verifies Build/Parse round-trip for MsgSysNotifyRegister.
func TestBuildParseNotifyRegister(t *testing.T) {
	tests := []struct {
		name       string
		registerID uint32
	}{
		{"typical", 100},
		{"zero", 0},
		{"max", 0xFFFFFFFF},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysNotifyRegister{
				RegisterID: tt.registerID,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysNotifyRegister{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.RegisterID != original.RegisterID {
				t.Errorf("RegisterID = %d, want %d", parsed.RegisterID, original.RegisterID)
			}
		})
	}
}

// TestBuildParseUnlockStage verifies Parse for MsgSysUnlockStage.
// This struct has no exported fields; Parse only discards a single zeroed uint16.
func TestBuildParseUnlockStage(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(0) // Zeroed (discarded by Parse)
	bf.Seek(0, io.SeekStart)

	parsed := &MsgSysUnlockStage{}
	if err := parsed.Parse(bf, ctx); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

// TestBuildParseUnlockGlobalSema verifies Build/Parse round-trip for MsgSysUnlockGlobalSema.
func TestBuildParseUnlockGlobalSema(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
	}{
		{"typical", 0xAABBCCDD},
		{"zero", 0},
	}

	ctx := &clientctx.ClientContext{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysUnlockGlobalSema{
				AckHandle: tt.ackHandle,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysUnlockGlobalSema{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.AckHandle != original.AckHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, original.AckHandle)
			}
		})
	}
}

// TestBuildParseStageDestruct verifies Build/Parse round-trip for MsgSysStageDestruct.
// This packet has no fields at all.
func TestBuildParseStageDestruct(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	original := &MsgSysStageDestruct{}

	bf := byteframe.NewByteFrame()
	if err := original.Build(bf, ctx); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(bf.Data()) != 0 {
		t.Errorf("Build() wrote %d bytes, want 0", len(bf.Data()))
	}

	parsed := &MsgSysStageDestruct{}
	if err := parsed.Parse(bf, ctx); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

// TestBuildParseCastedBinaryPayloadIntegrity verifies that a large payload is preserved
// exactly through Build/Parse for MsgSysCastedBinary.
func TestBuildParseCastedBinaryPayloadIntegrity(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	// Build a payload with recognizable pattern
	payload := make([]byte, 1024)
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	original := &MsgSysCastedBinary{
		CharID:         0x12345678,
		BroadcastType:  0x03,
		MessageType:    0x07,
		RawDataPayload: payload,
	}

	bf := byteframe.NewByteFrame()
	if err := original.Build(bf, ctx); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	bf.Seek(0, io.SeekStart)
	parsed := &MsgSysCastedBinary{}
	if err := parsed.Parse(bf, ctx); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(parsed.RawDataPayload) != len(payload) {
		t.Fatalf("Payload length = %d, want %d", len(parsed.RawDataPayload), len(payload))
	}

	for i, b := range parsed.RawDataPayload {
		if b != payload[i] {
			t.Errorf("Payload byte %d = 0x%02X, want 0x%02X", i, b, payload[i])
			break // Only report first mismatch
		}
	}
}

// TestBuildParseOperateRegisterPayloadIntegrity verifies payload integrity through
// manual-build/Parse for MsgSysOperateRegister.
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
func TestBuildParseOperateRegisterPayloadIntegrity(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	payload := make([]byte, 512)
	for i := range payload {
		payload[i] = byte((i * 7) % 256) // Non-trivial pattern
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xAABBCCDD)           // AckHandle
	bf.WriteUint32(42)                   // SemaphoreID
	bf.WriteUint16(0)                    // Zeroed
	bf.WriteUint16(uint16(len(payload))) // dataSize
	bf.WriteBytes(payload)

	bf.Seek(0, io.SeekStart)
	parsed := &MsgSysOperateRegister{}
	if err := parsed.Parse(bf, ctx); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !bytes.Equal(parsed.RawDataPayload, payload) {
		t.Errorf("Payload mismatch: got %d bytes, want %d bytes", len(parsed.RawDataPayload), len(payload))
	}
}

// TestBuildParseArrangeGuildMemberEmptySlice ensures that an empty CharIDs slice
// round-trips correctly (the uint8 count field should be 0).
// Build is NOT IMPLEMENTED, so we manually write the binary representation.
// Parse reads: uint32 AckHandle, uint32 GuildID, uint8 zeroed, uint8 charCount.
func TestBuildParseArrangeGuildMemberEmptySlice(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1)   // AckHandle
	bf.WriteUint32(100) // GuildID
	bf.WriteUint8(0)    // Zeroed
	bf.WriteUint8(0)    // charCount = 0

	// Verify the wire size: uint32 + uint32 + uint8 + uint8 = 10 bytes
	if len(bf.Data()) != 10 {
		t.Errorf("wrote %d bytes, want 10 for empty CharIDs", len(bf.Data()))
	}

	bf.Seek(0, io.SeekStart)
	parsed := &MsgMhfArrangeGuildMember{}
	if err := parsed.Parse(bf, ctx); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(parsed.CharIDs) != 0 {
		t.Errorf("CharIDs length = %d, want 0", len(parsed.CharIDs))
	}
}

// TestBuildBinaryFormat verifies the exact binary output format of a Build call
// for MsgSysDuplicateObject to ensure correct endianness and field ordering.
func TestBuildBinaryFormat(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	pkt := &MsgSysDuplicateObject{
		ObjID:       0x00000001,
		X:           0,
		Y:           0,
		Z:           0,
		Unk0:        0x00000002,
		OwnerCharID: 0x00000003,
	}

	bf := byteframe.NewByteFrame()
	if err := pkt.Build(bf, ctx); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	data := bf.Data()
	// Expected: 6 fields * 4 bytes = 24 bytes
	if len(data) != 24 {
		t.Fatalf("Build() wrote %d bytes, want 24", len(data))
	}

	// ObjID = 0x00000001 in big-endian
	if data[0] != 0x00 || data[1] != 0x00 || data[2] != 0x00 || data[3] != 0x01 {
		t.Errorf("ObjID bytes = %X, want 00000001", data[0:4])
	}

	// Unk0 = 0x00000002 at offset 16 (after ObjID + 3 floats)
	if data[16] != 0x00 || data[17] != 0x00 || data[18] != 0x00 || data[19] != 0x02 {
		t.Errorf("Unk0 bytes = %X, want 00000002", data[16:20])
	}

	// OwnerCharID = 0x00000003 at offset 20
	if data[20] != 0x00 || data[21] != 0x00 || data[22] != 0x00 || data[23] != 0x03 {
		t.Errorf("OwnerCharID bytes = %X, want 00000003", data[20:24])
	}
}

// TestBuildParseTimeBooleanEncoding verifies that the boolean field in MsgSysTime
// is encoded/decoded correctly for both true and false.
func TestBuildParseTimeBooleanEncoding(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	for _, val := range []bool{true, false} {
		t.Run("GetRemoteTime="+boolStr(val), func(t *testing.T) {
			original := &MsgSysTime{
				GetRemoteTime: val,
				Timestamp:     1234567890,
			}

			bf := byteframe.NewByteFrame()
			if err := original.Build(bf, ctx); err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			// Check raw byte: true=1, false=0
			data := bf.Data()
			if val && data[0] != 1 {
				t.Errorf("Boolean true encoded as %d, want 1", data[0])
			}
			if !val && data[0] != 0 {
				t.Errorf("Boolean false encoded as %d, want 0", data[0])
			}

			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysTime{}
			if err := parsed.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if parsed.GetRemoteTime != val {
				t.Errorf("GetRemoteTime = %v, want %v", parsed.GetRemoteTime, val)
			}
		})
	}
}

func boolStr(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// TestBuildParseSysAckBufferSmall verifies MsgSysAck round-trip with buffer response
// using the normal (non-extended) size field.
func TestBuildParseSysAckBufferSmall(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	payload := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	original := &MsgSysAck{
		AckHandle:        0xDEADBEEF,
		IsBufferResponse: true,
		ErrorCode:        0,
		AckData:          payload,
	}

	bf := byteframe.NewByteFrame()
	if err := original.Build(bf, ctx); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	bf.Seek(0, io.SeekStart)
	parsed := &MsgSysAck{}
	if err := parsed.Parse(bf, ctx); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsed.AckHandle != original.AckHandle {
		t.Errorf("AckHandle = 0x%X, want 0x%X", parsed.AckHandle, original.AckHandle)
	}
	if parsed.IsBufferResponse != original.IsBufferResponse {
		t.Errorf("IsBufferResponse = %v, want %v", parsed.IsBufferResponse, original.IsBufferResponse)
	}
	if parsed.ErrorCode != original.ErrorCode {
		t.Errorf("ErrorCode = %d, want %d", parsed.ErrorCode, original.ErrorCode)
	}
	if !bytes.Equal(parsed.AckData, original.AckData) {
		t.Errorf("AckData = %v, want %v", parsed.AckData, original.AckData)
	}
}

// TestBuildParseSysAckExtendedSize verifies MsgSysAck round-trip with a payload
// large enough to trigger the extended size field (>= 0xFFFF bytes).
func TestBuildParseSysAckExtendedSize(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	payload := make([]byte, 0x10000) // 65536 bytes, triggers extended size
	for i := range payload {
		payload[i] = byte(i % 256)
	}

	original := &MsgSysAck{
		AckHandle:        42,
		IsBufferResponse: true,
		ErrorCode:        0,
		AckData:          payload,
	}

	bf := byteframe.NewByteFrame()
	if err := original.Build(bf, ctx); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	bf.Seek(0, io.SeekStart)
	parsed := &MsgSysAck{}
	if err := parsed.Parse(bf, ctx); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(parsed.AckData) != len(payload) {
		t.Fatalf("AckData length = %d, want %d", len(parsed.AckData), len(payload))
	}
	if !bytes.Equal(parsed.AckData, payload) {
		t.Error("AckData content mismatch after extended size round-trip")
	}
}

// TestBuildParseSysAckNonBuffer verifies MsgSysAck round-trip with non-buffer response
// (exactly 4 bytes of data always read in Parse).
func TestBuildParseSysAckNonBuffer(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	original := &MsgSysAck{
		AckHandle:        100,
		IsBufferResponse: false,
		ErrorCode:        5,
		AckData:          []byte{0xAA, 0xBB, 0xCC, 0xDD},
	}

	bf := byteframe.NewByteFrame()
	if err := original.Build(bf, ctx); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	bf.Seek(0, io.SeekStart)
	parsed := &MsgSysAck{}
	if err := parsed.Parse(bf, ctx); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsed.AckHandle != original.AckHandle {
		t.Errorf("AckHandle = %d, want %d", parsed.AckHandle, original.AckHandle)
	}
	if parsed.IsBufferResponse != false {
		t.Errorf("IsBufferResponse = %v, want false", parsed.IsBufferResponse)
	}
	if parsed.ErrorCode != 5 {
		t.Errorf("ErrorCode = %d, want 5", parsed.ErrorCode)
	}
	// Non-buffer always reads exactly 4 bytes
	if len(parsed.AckData) != 4 {
		t.Errorf("AckData length = %d, want 4", len(parsed.AckData))
	}
	if !bytes.Equal(parsed.AckData, []byte{0xAA, 0xBB, 0xCC, 0xDD}) {
		t.Errorf("AckData = %v, want [AA BB CC DD]", parsed.AckData)
	}
}
