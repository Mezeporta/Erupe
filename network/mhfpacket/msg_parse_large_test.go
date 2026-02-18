package mhfpacket

import (
	"bytes"
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/clientctx"
)

// TestParseLargeMsgSysUpdateRightBuild tests Build for MsgSysUpdateRight (no Parse implementation).
func TestParseLargeMsgSysUpdateRightBuild(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	original := &MsgSysUpdateRight{
		ClientRespAckHandle: 0x12345678,
		Bitfield:            0xDEADBEEF,
		Rights:              nil,
		TokenLength:             0,
	}

	bf := byteframe.NewByteFrame()
	if err := original.Build(bf, ctx); err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Verify binary output manually:
	// uint32 ClientRespAckHandle + uint32 Bitfield + uint16 Rights count(0) + uint16 padding(0) + ps.Uint16 empty string(uint16(1) + 0x00)
	data := bf.Data()
	if len(data) < 12 {
		t.Fatalf("Build() wrote %d bytes, want at least 12", len(data))
	}

	_, _ = bf.Seek(0, io.SeekStart)
	if bf.ReadUint32() != 0x12345678 {
		t.Error("ClientRespAckHandle mismatch")
	}
	if bf.ReadUint32() != 0xDEADBEEF {
		t.Error("Bitfield mismatch")
	}
	if bf.ReadUint16() != 0 {
		t.Error("Rights count should be 0")
	}
}

// TestParseLargeMsgMhfOperateWarehouse tests Parse for MsgMhfOperateWarehouse.
func TestParseLargeMsgMhfOperateWarehouse(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xAABBCCDD) // AckHandle
	bf.WriteUint8(1)           // Operation
	bf.WriteUint8(0)           // BoxType = item
	bf.WriteUint8(2)           // BoxIndex
	bf.WriteUint8(8)           // lenName (unused but read)
	bf.WriteUint16(0)          // Unk
	bf.WriteBytes([]byte("TestBox"))
	bf.WriteUint8(0) // null terminator
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfOperateWarehouse{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xAABBCCDD {
		t.Errorf("AckHandle = 0x%X, want 0xAABBCCDD", pkt.AckHandle)
	}
	if pkt.Operation != 1 {
		t.Errorf("Operation = %d, want 1", pkt.Operation)
	}
	if pkt.BoxType != 0 {
		t.Errorf("BoxType = %d, want 0", pkt.BoxType)
	}
	if pkt.BoxIndex != 2 {
		t.Errorf("BoxIndex = %d, want 2", pkt.BoxIndex)
	}
	if pkt.Name != "TestBox" {
		t.Errorf("Name = %q, want %q", pkt.Name, "TestBox")
	}
}

// TestParseLargeMsgMhfOperateWarehouseEquip tests Parse for MsgMhfOperateWarehouse with equip box type.
func TestParseLargeMsgMhfOperateWarehouseEquip(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(42)  // AckHandle
	bf.WriteUint8(2)    // Operation
	bf.WriteUint8(1)    // BoxType = equip
	bf.WriteUint8(0)    // BoxIndex
	bf.WriteUint8(5)    // lenName
	bf.WriteUint16(0)   // Unk
	bf.WriteBytes([]byte("Arms"))
	bf.WriteUint8(0) // null terminator
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfOperateWarehouse{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.BoxType != 1 {
		t.Errorf("BoxType = %d, want 1", pkt.BoxType)
	}
	if pkt.Name != "Arms" {
		t.Errorf("Name = %q, want %q", pkt.Name, "Arms")
	}
}

// TestParseLargeMsgMhfLoadHouse tests Parse for MsgMhfLoadHouse.
func TestParseLargeMsgMhfLoadHouse(t *testing.T) {
	tests := []struct {
		name        string
		ackHandle   uint32
		charID      uint32
		destination uint8
		checkPass   bool
		password    string
	}{
		{"with password", 0xAABBCCDD, 12345, 1, true, "pass123"},
		{"no password", 0x11111111, 0, 0, false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.charID)
			bf.WriteUint8(tt.destination)
			bf.WriteBool(tt.checkPass)
			bf.WriteUint16(0) // Unk (hardcoded 0)
			bf.WriteUint8(uint8(len(tt.password) + 1)) // Password length
			bf.WriteBytes([]byte(tt.password))
			bf.WriteUint8(0) // null terminator
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfLoadHouse{}
			if err := pkt.Parse(bf, nil); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ackHandle)
			}
			if pkt.CharID != tt.charID {
				t.Errorf("CharID = %d, want %d", pkt.CharID, tt.charID)
			}
			if pkt.Destination != tt.destination {
				t.Errorf("Destination = %d, want %d", pkt.Destination, tt.destination)
			}
			if pkt.CheckPass != tt.checkPass {
				t.Errorf("CheckPass = %v, want %v", pkt.CheckPass, tt.checkPass)
			}
			if pkt.Password != tt.password {
				t.Errorf("Password = %q, want %q", pkt.Password, tt.password)
			}
		})
	}
}

// TestParseLargeMsgMhfSendMail tests Parse for MsgMhfSendMail.
func TestParseLargeMsgMhfSendMail(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x12345678) // AckHandle
	bf.WriteUint32(99999)      // RecipientID
	bf.WriteUint16(6)          // SubjectLength
	bf.WriteUint16(12)         // BodyLength
	bf.WriteUint32(5)          // Quantity
	bf.WriteUint16(1001)       // ItemID
	bf.WriteBytes([]byte("Hello"))
	bf.WriteUint8(0) // null terminator for Subject
	bf.WriteBytes([]byte("Hello World"))
	bf.WriteUint8(0) // null terminator for Body
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfSendMail{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x12345678 {
		t.Errorf("AckHandle = 0x%X, want 0x12345678", pkt.AckHandle)
	}
	if pkt.RecipientID != 99999 {
		t.Errorf("RecipientID = %d, want 99999", pkt.RecipientID)
	}
	if pkt.SubjectLength != 6 {
		t.Errorf("SubjectLength = %d, want 6", pkt.SubjectLength)
	}
	if pkt.BodyLength != 12 {
		t.Errorf("BodyLength = %d, want 12", pkt.BodyLength)
	}
	if pkt.Quantity != 5 {
		t.Errorf("Quantity = %d, want 5", pkt.Quantity)
	}
	if pkt.ItemID != 1001 {
		t.Errorf("ItemID = %d, want 1001", pkt.ItemID)
	}
	if pkt.Subject != "Hello" {
		t.Errorf("Subject = %q, want %q", pkt.Subject, "Hello")
	}
	if pkt.Body != "Hello World" {
		t.Errorf("Body = %q, want %q", pkt.Body, "Hello World")
	}
}

// TestParseLargeMsgMhfApplyBbsArticle tests Parse for MsgMhfApplyBbsArticle.
func TestParseLargeMsgMhfApplyBbsArticle(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xCAFEBABE) // AckHandle
	bf.WriteUint32(42)         // Unk0

	// Unk1: 16 bytes
	unk1 := make([]byte, 16)
	for i := range unk1 {
		unk1[i] = byte(i + 1)
	}
	bf.WriteBytes(unk1)

	// Name: 32 bytes (padded with nulls) - uses bfutil.UpToNull
	nameBytes := make([]byte, 32)
	copy(nameBytes, "Hunter")
	bf.WriteBytes(nameBytes)

	// Title: 128 bytes (padded with nulls)
	titleBytes := make([]byte, 128)
	copy(titleBytes, "My Post Title")
	bf.WriteBytes(titleBytes)

	// Description: 256 bytes (padded with nulls)
	descBytes := make([]byte, 256)
	copy(descBytes, "This is a description")
	bf.WriteBytes(descBytes)

	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfApplyBbsArticle{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xCAFEBABE {
		t.Errorf("AckHandle = 0x%X, want 0xCAFEBABE", pkt.AckHandle)
	}
	if pkt.Unk0 != 42 {
		t.Errorf("Unk0 = %d, want 42", pkt.Unk0)
	}
	if !bytes.Equal(pkt.Unk1, unk1) {
		t.Error("Unk1 mismatch")
	}
	if pkt.Name != "Hunter" {
		t.Errorf("Name = %q, want %q", pkt.Name, "Hunter")
	}
	if pkt.Title != "My Post Title" {
		t.Errorf("Title = %q, want %q", pkt.Title, "My Post Title")
	}
	if pkt.Description != "This is a description" {
		t.Errorf("Description = %q, want %q", pkt.Description, "This is a description")
	}
}

// TestParseLargeMsgMhfChargeFesta tests Parse for MsgMhfChargeFesta.
func TestParseLargeMsgMhfChargeFesta(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x11223344) // AckHandle
	bf.WriteUint32(100)        // FestaID
	bf.WriteUint32(200)        // GuildID
	bf.WriteUint16(3)          // soul count
	bf.WriteUint16(10)         // soul value 1
	bf.WriteUint16(20)         // soul value 2
	bf.WriteUint16(30)         // soul value 3
	bf.WriteUint8(0)           // Unk
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfChargeFesta{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x11223344 {
		t.Errorf("AckHandle = 0x%X, want 0x11223344", pkt.AckHandle)
	}
	if pkt.FestaID != 100 {
		t.Errorf("FestaID = %d, want 100", pkt.FestaID)
	}
	if pkt.GuildID != 200 {
		t.Errorf("GuildID = %d, want 200", pkt.GuildID)
	}
	if len(pkt.Souls) != 3 {
		t.Fatalf("Souls len = %d, want 3", len(pkt.Souls))
	}
	expectedSouls := []uint16{10, 20, 30}
	for i, v := range expectedSouls {
		if pkt.Souls[i] != v {
			t.Errorf("Souls[%d] = %d, want %d", i, pkt.Souls[i], v)
		}
	}
}

// TestParseLargeMsgMhfChargeFestaZeroSouls tests Parse for MsgMhfChargeFesta with zero soul entries.
func TestParseLargeMsgMhfChargeFestaZeroSouls(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1)   // AckHandle
	bf.WriteUint32(0)   // FestaID
	bf.WriteUint32(0)   // GuildID
	bf.WriteUint16(0)   // soul count = 0
	bf.WriteUint8(0)    // Unk
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfChargeFesta{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(pkt.Souls) != 0 {
		t.Errorf("Souls len = %d, want 0", len(pkt.Souls))
	}
}

// TestParseLargeMsgMhfOperateJoint tests Parse for MsgMhfOperateJoint.
// Parse reads: uint32 AckHandle, uint32 AllianceID, uint32 GuildID, uint8 Action,
// uint8 dataLen, 4 bytes Data1, dataLen bytes Data2.
func TestParseLargeMsgMhfOperateJoint(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x12345678)              // AckHandle
	bf.WriteUint32(100)                     // AllianceID
	bf.WriteUint32(200)                     // GuildID
	bf.WriteUint8(0x01)                     // Action = OPERATE_JOINT_DISBAND
	bf.WriteUint8(3)                        // dataLen = 3
	bf.WriteBytes([]byte{0xAA, 0xBB, 0xCC, 0xDD}) // Data1 (always 4 bytes)
	bf.WriteBytes([]byte{0x01, 0x02, 0x03})        // Data2 (dataLen bytes)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfOperateJoint{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x12345678 {
		t.Errorf("AckHandle = 0x%X, want 0x12345678", pkt.AckHandle)
	}
	if pkt.AllianceID != 100 {
		t.Errorf("AllianceID = %d, want 100", pkt.AllianceID)
	}
	if pkt.GuildID != 200 {
		t.Errorf("GuildID = %d, want 200", pkt.GuildID)
	}
	if pkt.Action != OPERATE_JOINT_DISBAND {
		t.Errorf("Action = %d, want %d", pkt.Action, OPERATE_JOINT_DISBAND)
	}
	if pkt.Data1 == nil {
		t.Fatal("Data1 is nil")
	}
	if pkt.Data2 == nil {
		t.Fatal("Data2 is nil")
	}
}

// TestParseLargeMsgMhfOperationInvGuild tests Parse for MsgMhfOperationInvGuild.
func TestParseLargeMsgMhfOperationInvGuild(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xAABBCCDD) // AckHandle
	bf.WriteUint8(1)           // Operation
	bf.WriteUint8(5)           // ActiveHours
	bf.WriteUint8(7)           // DaysActive
	bf.WriteUint8(3)           // PlayStyle
	bf.WriteUint8(2)           // GuildRequest
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfOperationInvGuild{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xAABBCCDD {
		t.Errorf("AckHandle = 0x%X, want 0xAABBCCDD", pkt.AckHandle)
	}
	if pkt.Operation != 1 {
		t.Errorf("Operation = %d, want 1", pkt.Operation)
	}
	if pkt.ActiveHours != 5 {
		t.Errorf("ActiveHours = %d, want 5", pkt.ActiveHours)
	}
	if pkt.DaysActive != 7 {
		t.Errorf("DaysActive = %d, want 7", pkt.DaysActive)
	}
	if pkt.PlayStyle != 3 {
		t.Errorf("PlayStyle = %d, want 3", pkt.PlayStyle)
	}
	if pkt.GuildRequest != 2 {
		t.Errorf("GuildRequest = %d, want 2", pkt.GuildRequest)
	}
}

// TestParseLargeMsgMhfSaveMercenary tests Parse for MsgMhfSaveMercenary.
func TestParseLargeMsgMhfSaveMercenary(t *testing.T) {
	mercData := []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xCAFEBABE) // AckHandle
	bf.WriteUint32(0)          // lenData (skipped)
	bf.WriteUint32(5000)       // GCP
	bf.WriteUint32(42)         // PactMercID
	bf.WriteUint32(uint32(len(mercData))) // dataSize
	bf.WriteUint32(0)          // Merc index (skipped)
	bf.WriteBytes(mercData)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfSaveMercenary{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xCAFEBABE {
		t.Errorf("AckHandle = 0x%X, want 0xCAFEBABE", pkt.AckHandle)
	}
	if pkt.GCP != 5000 {
		t.Errorf("GCP = %d, want 5000", pkt.GCP)
	}
	if pkt.PactMercID != 42 {
		t.Errorf("PactMercID = %d, want 42", pkt.PactMercID)
	}
	if !bytes.Equal(pkt.MercData, mercData) {
		t.Errorf("MercData = %v, want %v", pkt.MercData, mercData)
	}
}

// TestParseLargeMsgMhfUpdateHouse tests Parse for MsgMhfUpdateHouse.
func TestParseLargeMsgMhfUpdateHouse(t *testing.T) {
	tests := []struct {
		name     string
		state    uint8
		password string
	}{
		{"with password", 1, "secret"},
		{"empty password", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(0x12345678) // AckHandle
			bf.WriteUint8(tt.state)    // State
			bf.WriteUint8(1)           // Unk1
			bf.WriteUint16(0)          // Unk2
			bf.WriteUint8(uint8(len(tt.password) + 1)) // Password length
			bf.WriteBytes([]byte(tt.password))
			bf.WriteUint8(0) // null terminator
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfUpdateHouse{}
			if err := pkt.Parse(bf, nil); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != 0x12345678 {
				t.Errorf("AckHandle = 0x%X, want 0x12345678", pkt.AckHandle)
			}
			if pkt.State != tt.state {
				t.Errorf("State = %d, want %d", pkt.State, tt.state)
			}
			if pkt.Unk1 != 1 {
				t.Errorf("Unk1 = %d, want 1", pkt.Unk1)
			}
			if pkt.Password != tt.password {
				t.Errorf("Password = %q, want %q", pkt.Password, tt.password)
			}
		})
	}
}

// TestParseLargeMsgSysCreateAcquireSemaphore tests Parse for MsgSysCreateAcquireSemaphore.
func TestParseLargeMsgSysCreateAcquireSemaphore(t *testing.T) {
	semID := "stage_001"
	semBytes := make([]byte, len(semID)+1) // include space for null if needed
	copy(semBytes, semID)

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xDEADBEEF) // AckHandle
	bf.WriteUint16(100)        // Unk0
	bf.WriteUint8(4)           // PlayerCount
	bf.WriteUint8(uint8(len(semBytes))) // SemaphoreIDLength
	bf.WriteBytes(semBytes)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysCreateAcquireSemaphore{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xDEADBEEF {
		t.Errorf("AckHandle = 0x%X, want 0xDEADBEEF", pkt.AckHandle)
	}
	if pkt.Unk0 != 100 {
		t.Errorf("Unk0 = %d, want 100", pkt.Unk0)
	}
	if pkt.PlayerCount != 4 {
		t.Errorf("PlayerCount = %d, want 4", pkt.PlayerCount)
	}
	if pkt.SemaphoreID != semID {
		t.Errorf("SemaphoreID = %q, want %q", pkt.SemaphoreID, semID)
	}
}

// TestParseLargeMsgMhfOperateGuild tests Parse for MsgMhfOperateGuild.
func TestParseLargeMsgMhfOperateGuild(t *testing.T) {
	dataPayload := []byte{0x10, 0x20, 0x30, 0x40, 0x50}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xAABBCCDD)          // AckHandle
	bf.WriteUint32(999)                 // GuildID
	bf.WriteUint8(0x09)                 // Action = OperateGuildUpdateComment
	bf.WriteUint8(uint8(len(dataPayload))) // dataLen
	bf.WriteBytes([]byte{0x01, 0x02, 0x03, 0x04}) // Data1 (always 4 bytes)
	bf.WriteBytes(dataPayload) // Data2 (dataLen bytes)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfOperateGuild{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xAABBCCDD {
		t.Errorf("AckHandle = 0x%X, want 0xAABBCCDD", pkt.AckHandle)
	}
	if pkt.GuildID != 999 {
		t.Errorf("GuildID = %d, want 999", pkt.GuildID)
	}
	if pkt.Action != OperateGuildUpdateComment {
		t.Errorf("Action = %d, want %d", pkt.Action, OperateGuildUpdateComment)
	}
	if pkt.Data1 == nil {
		t.Fatal("Data1 is nil")
	}
	if pkt.Data2 == nil {
		t.Fatal("Data2 is nil")
	}
	data2Bytes := pkt.Data2.Data()
	if !bytes.Equal(data2Bytes, dataPayload) {
		t.Errorf("Data2 = %v, want %v", data2Bytes, dataPayload)
	}
}

// TestParseLargeMsgMhfReadBeatLevel tests Parse for MsgMhfReadBeatLevel.
func TestParseLargeMsgMhfReadBeatLevel(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x12345678) // AckHandle
	bf.WriteUint32(1)          // Unk0
	bf.WriteUint32(4)          // ValidIDCount

	// Write 16 uint32 IDs
	ids := [16]uint32{0x74, 0x6B, 0x02, 0x24, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	for _, id := range ids {
		bf.WriteUint32(id)
	}
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfReadBeatLevel{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x12345678 {
		t.Errorf("AckHandle = 0x%X, want 0x12345678", pkt.AckHandle)
	}
	if pkt.Unk0 != 1 {
		t.Errorf("Unk0 = %d, want 1", pkt.Unk0)
	}
	if pkt.ValidIDCount != 4 {
		t.Errorf("ValidIDCount = %d, want 4", pkt.ValidIDCount)
	}
	for i, id := range ids {
		if pkt.IDs[i] != id {
			t.Errorf("IDs[%d] = 0x%X, want 0x%X", i, pkt.IDs[i], id)
		}
	}
}

// TestParseLargeMsgSysCreateObject tests Parse for MsgSysCreateObject.
func TestParseLargeMsgSysCreateObject(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		x, y, z   float32
		unk0      uint32
	}{
		{"origin", 1, 0.0, 0.0, 0.0, 0},
		{"typical", 0x12345678, 1.5, 2.5, 3.5, 42},
		{"negative coords", 0xFFFFFFFF, -100.25, 200.75, -300.125, 0xDEADBEEF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteFloat32(tt.x)
			bf.WriteFloat32(tt.y)
			bf.WriteFloat32(tt.z)
			bf.WriteUint32(tt.unk0)
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgSysCreateObject{}
			if err := pkt.Parse(bf, nil); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ackHandle)
			}
			if pkt.X != tt.x {
				t.Errorf("X = %f, want %f", pkt.X, tt.x)
			}
			if pkt.Y != tt.y {
				t.Errorf("Y = %f, want %f", pkt.Y, tt.y)
			}
			if pkt.Z != tt.z {
				t.Errorf("Z = %f, want %f", pkt.Z, tt.z)
			}
			if pkt.Unk0 != tt.unk0 {
				t.Errorf("Unk0 = %d, want %d", pkt.Unk0, tt.unk0)
			}
		})
	}
}

// TestParseLargeMsgSysLockGlobalSema tests Parse for MsgSysLockGlobalSema.
func TestParseLargeMsgSysLockGlobalSema(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xDEADBEEF) // AckHandle
	bf.WriteUint16(8)          // UserIDLength
	bf.WriteUint16(11)         // ServerChannelIDLength
	bf.WriteBytes([]byte("user123"))
	bf.WriteUint8(0) // null terminator
	bf.WriteBytes([]byte("channel_01"))
	bf.WriteUint8(0) // null terminator
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysLockGlobalSema{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xDEADBEEF {
		t.Errorf("AckHandle = 0x%X, want 0xDEADBEEF", pkt.AckHandle)
	}
	if pkt.UserIDLength != 8 {
		t.Errorf("UserIDLength = %d, want 8", pkt.UserIDLength)
	}
	if pkt.ServerChannelIDLength != 11 {
		t.Errorf("ServerChannelIDLength = %d, want 11", pkt.ServerChannelIDLength)
	}
	if pkt.UserIDString != "user123" {
		t.Errorf("UserIDString = %q, want %q", pkt.UserIDString, "user123")
	}
	if pkt.ServerChannelIDString != "channel_01" {
		t.Errorf("ServerChannelIDString = %q, want %q", pkt.ServerChannelIDString, "channel_01")
	}
}

// TestParseLargeMsgMhfCreateJoint tests Parse for MsgMhfCreateJoint.
func TestParseLargeMsgMhfCreateJoint(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xCAFEBABE)     // AckHandle
	bf.WriteUint32(500)            // GuildID
	bf.WriteUint32(15)             // len (unused)
	bf.WriteBytes([]byte("Alliance01"))
	bf.WriteUint8(0) // null terminator
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfCreateJoint{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xCAFEBABE {
		t.Errorf("AckHandle = 0x%X, want 0xCAFEBABE", pkt.AckHandle)
	}
	if pkt.GuildID != 500 {
		t.Errorf("GuildID = %d, want 500", pkt.GuildID)
	}
	if pkt.Name != "Alliance01" {
		t.Errorf("Name = %q, want %q", pkt.Name, "Alliance01")
	}
}

// TestParseLargeMsgMhfGetUdTacticsRemainingPoint tests Parse for MsgMhfGetUdTacticsRemainingPoint.
func TestParseLargeMsgMhfGetUdTacticsRemainingPoint(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x12345678) // AckHandle
	bf.WriteUint32(100)        // Unk0
	bf.WriteUint32(200)        // Unk1
	bf.WriteUint32(300)        // Unk2
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfGetUdTacticsRemainingPoint{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x12345678 {
		t.Errorf("AckHandle = 0x%X, want 0x12345678", pkt.AckHandle)
	}
	if pkt.Unk0 != 100 {
		t.Errorf("Unk0 = %d, want 100", pkt.Unk0)
	}
	if pkt.Unk1 != 200 {
		t.Errorf("Unk1 = %d, want 200", pkt.Unk1)
	}
	if pkt.Unk2 != 300 {
		t.Errorf("Unk2 = %d, want 300", pkt.Unk2)
	}
}

// TestParseLargeMsgMhfPostCafeDurationBonusReceived tests Parse for MsgMhfPostCafeDurationBonusReceived.
func TestParseLargeMsgMhfPostCafeDurationBonusReceived(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xAABBCCDD) // AckHandle
	bf.WriteUint32(3)          // count
	bf.WriteUint32(1001)       // CafeBonusID[0]
	bf.WriteUint32(1002)       // CafeBonusID[1]
	bf.WriteUint32(1003)       // CafeBonusID[2]
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfPostCafeDurationBonusReceived{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xAABBCCDD {
		t.Errorf("AckHandle = 0x%X, want 0xAABBCCDD", pkt.AckHandle)
	}
	if len(pkt.CafeBonusID) != 3 {
		t.Fatalf("CafeBonusID len = %d, want 3", len(pkt.CafeBonusID))
	}
	expected := []uint32{1001, 1002, 1003}
	for i, v := range expected {
		if pkt.CafeBonusID[i] != v {
			t.Errorf("CafeBonusID[%d] = %d, want %d", i, pkt.CafeBonusID[i], v)
		}
	}
}

// TestParseLargeMsgMhfPostCafeDurationBonusReceivedEmpty tests Parse with zero IDs.
func TestParseLargeMsgMhfPostCafeDurationBonusReceivedEmpty(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint32(0) // count = 0
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfPostCafeDurationBonusReceived{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
	if len(pkt.CafeBonusID) != 0 {
		t.Errorf("CafeBonusID len = %d, want 0", len(pkt.CafeBonusID))
	}
}

// TestParseLargeMsgMhfRegistGuildAdventureDiva tests Parse for MsgMhfRegistGuildAdventureDiva.
func TestParseLargeMsgMhfRegistGuildAdventureDiva(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x12345678) // AckHandle
	bf.WriteUint32(5)          // Destination
	bf.WriteUint32(1000)       // Charge
	bf.WriteUint32(42)         // CharID (skipped)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfRegistGuildAdventureDiva{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x12345678 {
		t.Errorf("AckHandle = 0x%X, want 0x12345678", pkt.AckHandle)
	}
	if pkt.Destination != 5 {
		t.Errorf("Destination = %d, want 5", pkt.Destination)
	}
	if pkt.Charge != 1000 {
		t.Errorf("Charge = %d, want 1000", pkt.Charge)
	}
}

// TestParseLargeMsgMhfStateFestaG tests Parse for MsgMhfStateFestaG.
func TestParseLargeMsgMhfStateFestaG(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xDEADBEEF) // AckHandle
	bf.WriteUint32(100)        // FestaID
	bf.WriteUint32(200)        // GuildID
	bf.WriteUint16(0)          // Hardcoded 0
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfStateFestaG{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xDEADBEEF {
		t.Errorf("AckHandle = 0x%X, want 0xDEADBEEF", pkt.AckHandle)
	}
	if pkt.FestaID != 100 {
		t.Errorf("FestaID = %d, want 100", pkt.FestaID)
	}
	if pkt.GuildID != 200 {
		t.Errorf("GuildID = %d, want 200", pkt.GuildID)
	}
}

// TestParseLargeMsgMhfStateFestaU tests Parse for MsgMhfStateFestaU.
func TestParseLargeMsgMhfStateFestaU(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xCAFEBABE) // AckHandle
	bf.WriteUint32(300)        // FestaID
	bf.WriteUint32(400)        // GuildID
	bf.WriteUint16(0)          // Hardcoded 0
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfStateFestaU{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xCAFEBABE {
		t.Errorf("AckHandle = 0x%X, want 0xCAFEBABE", pkt.AckHandle)
	}
	if pkt.FestaID != 300 {
		t.Errorf("FestaID = %d, want 300", pkt.FestaID)
	}
	if pkt.GuildID != 400 {
		t.Errorf("GuildID = %d, want 400", pkt.GuildID)
	}
}

// TestParseLargeMsgSysEnumerateStage tests Parse for MsgSysEnumerateStage.
func TestParseLargeMsgSysEnumerateStage(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x11223344) // AckHandle
	bf.WriteUint8(1)           // Unk0
	bf.WriteUint8(0)           // skipped byte
	bf.WriteBytes([]byte("quest_"))
	bf.WriteUint8(0) // null terminator
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysEnumerateStage{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x11223344 {
		t.Errorf("AckHandle = 0x%X, want 0x11223344", pkt.AckHandle)
	}
	if pkt.StagePrefix != "quest_" {
		t.Errorf("StagePrefix = %q, want %q", pkt.StagePrefix, "quest_")
	}
}

// TestParseLargeMsgSysReserveStage tests Parse for MsgSysReserveStage.
func TestParseLargeMsgSysReserveStage(t *testing.T) {
	stageID := "stage_42"
	stageBytes := make([]byte, len(stageID)+1) // padded with null at end
	copy(stageBytes, stageID)

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0xAABBCCDD) // AckHandle
	bf.WriteUint8(0x11)        // Ready
	bf.WriteUint8(uint8(len(stageBytes))) // stageIDLength
	bf.WriteBytes(stageBytes)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysReserveStage{}
	if err := pkt.Parse(bf, nil); err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0xAABBCCDD {
		t.Errorf("AckHandle = 0x%X, want 0xAABBCCDD", pkt.AckHandle)
	}
	if pkt.Ready != 0x11 {
		t.Errorf("Ready = 0x%X, want 0x11", pkt.Ready)
	}
	if pkt.StageID != stageID {
		t.Errorf("StageID = %q, want %q", pkt.StageID, stageID)
	}
}
