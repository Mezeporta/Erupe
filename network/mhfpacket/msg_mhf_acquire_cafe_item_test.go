package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

func init() {
	// Initialize ErupeConfig for tests that access it
	_config.ErupeConfig = &_config.Config{
		RealClientMode: _config.ZZ, // Default to ZZ for tests
	}
}

func TestMsgMhfAcquireCafeItemOpcode(t *testing.T) {
	pkt := &MsgMhfAcquireCafeItem{}
	if pkt.Opcode() != network.MSG_MHF_ACQUIRE_CAFE_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_ACQUIRE_CAFE_ITEM", pkt.Opcode())
	}
}

func TestMsgMhfAcquireCafeItemParse(t *testing.T) {
	// Test basic parsing with current implementation (always reads uint32 for PointCost)
	// Current code: m.PointCost = bf.ReadUint32() (no client mode check)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x12345678) // AckHandle
	bf.WriteUint16(1)          // ItemType
	bf.WriteUint16(100)        // ItemID
	bf.WriteUint16(5)          // Quant
	bf.WriteUint32(1000)       // PointCost (uint32)
	bf.WriteUint16(0)          // Unk0

	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfAcquireCafeItem{}
	ctx := &clientctx.ClientContext{}

	err := pkt.Parse(bf, ctx)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x12345678 {
		t.Errorf("AckHandle = 0x%X, want 0x12345678", pkt.AckHandle)
	}
	if pkt.ItemType != 1 {
		t.Errorf("ItemType = %d, want 1", pkt.ItemType)
	}
	if pkt.ItemID != 100 {
		t.Errorf("ItemID = %d, want 100", pkt.ItemID)
	}
	if pkt.Quant != 5 {
		t.Errorf("Quant = %d, want 5", pkt.Quant)
	}
	if pkt.PointCost != 1000 {
		t.Errorf("PointCost = %d, want 1000", pkt.PointCost)
	}
}

// TestMsgMhfAcquireCafeItemParseUint32PointCost documents the current behavior.
//
// CURRENT BEHAVIOR: Always reads PointCost as uint32.
//
// EXPECTED BEHAVIOR AFTER FIX (commit 3d0114c):
// - G6+: Read PointCost as uint32
// - G1-G5.2: Read PointCost as uint16
//
// This test verifies current uint32 parsing works correctly.
// After the fix is applied, this test should still pass for G6+ clients.
func TestMsgMhfAcquireCafeItemParseUint32PointCost(t *testing.T) {
	tests := []struct {
		name      string
		pointCost uint32
		wantCost  uint32
	}{
		{"small cost", 100, 100},
		{"medium cost", 5000, 5000},
		{"large cost exceeding uint16", 70000, 70000},
		{"max uint32", 0xFFFFFFFF, 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(0xAAAABBBB) // AckHandle
			bf.WriteUint16(1)          // ItemType
			bf.WriteUint16(200)        // ItemID
			bf.WriteUint16(10)         // Quant
			bf.WriteUint32(tt.pointCost)
			bf.WriteUint16(0) // Unk0

			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfAcquireCafeItem{}
			ctx := &clientctx.ClientContext{}

			err := pkt.Parse(bf, ctx)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.PointCost != tt.wantCost {
				t.Errorf("PointCost = %d, want %d", pkt.PointCost, tt.wantCost)
			}
		})
	}
}

// TestMsgMhfAcquireCafeItemParseFieldOrder verifies the exact field order in parsing.
// This is important because the fix changes when PointCost is read (uint16 vs uint32).
func TestMsgMhfAcquireCafeItemParseFieldOrder(t *testing.T) {
	// Build a packet with known values
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(0x11223344) // AckHandle (offset 0-3)
	bf.WriteUint16(0x5566)     // ItemType (offset 4-5)
	bf.WriteUint16(0x7788)     // ItemID (offset 6-7)
	bf.WriteUint16(0x99AA)     // Quant (offset 8-9)
	bf.WriteUint32(0xBBCCDDEE) // PointCost (offset 10-13)
	bf.WriteUint16(0xFF00)     // Unk0 (offset 14-15)

	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfAcquireCafeItem{}
	err := pkt.Parse(bf, &clientctx.ClientContext{})
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if pkt.AckHandle != 0x11223344 {
		t.Errorf("AckHandle = 0x%X, want 0x11223344", pkt.AckHandle)
	}
	if pkt.ItemType != 0x5566 {
		t.Errorf("ItemType = 0x%X, want 0x5566", pkt.ItemType)
	}
	if pkt.ItemID != 0x7788 {
		t.Errorf("ItemID = 0x%X, want 0x7788", pkt.ItemID)
	}
	if pkt.Quant != 0x99AA {
		t.Errorf("Quant = 0x%X, want 0x99AA", pkt.Quant)
	}
	if pkt.PointCost != 0xBBCCDDEE {
		t.Errorf("PointCost = 0x%X, want 0xBBCCDDEE", pkt.PointCost)
	}
	if pkt.Unk0 != 0xFF00 {
		t.Errorf("Unk0 = 0x%X, want 0xFF00", pkt.Unk0)
	}
}

func TestMsgMhfAcquireCafeItemBuildNotImplemented(t *testing.T) {
	pkt := &MsgMhfAcquireCafeItem{
		AckHandle: 123,
		ItemType:  1,
		ItemID:    100,
		Quant:     5,
		PointCost: 1000,
	}

	bf := byteframe.NewByteFrame()
	ctx := &clientctx.ClientContext{}

	err := pkt.Build(bf, ctx)
	if err == nil {
		t.Error("Build() should return error (NOT IMPLEMENTED)")
	}
}

func TestMsgMhfAcquireCafeItemFromOpcode(t *testing.T) {
	pkt := FromOpcode(network.MSG_MHF_ACQUIRE_CAFE_ITEM)
	if pkt == nil {
		t.Fatal("FromOpcode(MSG_MHF_ACQUIRE_CAFE_ITEM) returned nil")
	}
	if pkt.Opcode() != network.MSG_MHF_ACQUIRE_CAFE_ITEM {
		t.Errorf("Opcode() = %s, want MSG_MHF_ACQUIRE_CAFE_ITEM", pkt.Opcode())
	}
}
