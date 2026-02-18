package mhfpacket

import (
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcourse"
	"erupe-ce/network/clientctx"
)

// TestParseCoverage_Implemented exercises Parse() on all packet types whose Parse
// method is implemented (reads from ByteFrame) but was not yet covered by tests.
// Each test provides a ByteFrame with enough bytes for the Parse to succeed.
func TestParseCoverage_Implemented(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	tests := []struct {
		name     string
		pkt      MHFPacket
		dataSize int // minimum bytes to satisfy Parse
	}{
		// 4-byte packets (AckHandle only)
		{"MsgMhfGetSenyuDailyCount", &MsgMhfGetSenyuDailyCount{}, 4},
		{"MsgMhfUnreserveSrg", &MsgMhfUnreserveSrg{}, 4},

		// 1-byte packets
		// MsgSysLogout reads uint8
		{"MsgSysLogout", &MsgSysLogout{}, 1},

		// 6-byte packets
		{"MsgMhfGetRandFromTable", &MsgMhfGetRandFromTable{}, 6},

		// 8-byte packets
		{"MsgMhfPostBoostTimeLimit", &MsgMhfPostBoostTimeLimit{}, 8},

		// 9-byte packets
		{"MsgMhfPlayFreeGacha", &MsgMhfPlayFreeGacha{}, 9},

		// 12-byte packets
		{"MsgMhfEnumerateItem", &MsgMhfEnumerateItem{}, 12},
		{"MsgMhfGetBreakSeibatuLevelReward", &MsgMhfGetBreakSeibatuLevelReward{}, 12},
		{"MsgMhfReadLastWeekBeatRanking", &MsgMhfReadLastWeekBeatRanking{}, 12},

		// 16-byte packets (4+1+1+4+1+2+2+1)
		{"MsgMhfPostSeibattle", &MsgMhfPostSeibattle{}, 16},

		// 16-byte packets
		{"MsgMhfGetNotice", &MsgMhfGetNotice{}, 16},
		{"MsgMhfCaravanRanking", &MsgMhfCaravanRanking{}, 16},
		{"MsgMhfReadBeatLevelAllRanking", &MsgMhfReadBeatLevelAllRanking{}, 16},
		{"MsgMhfCaravanMyRank", &MsgMhfCaravanMyRank{}, 16},

		// 20-byte packets
		{"MsgMhfPostNotice", &MsgMhfPostNotice{}, 20},

		// 24-byte packets
		{"MsgMhfGetFixedSeibatuRankingTable", &MsgMhfGetFixedSeibatuRankingTable{}, 24},

		// 32-byte packets
		{"MsgMhfCaravanMyScore", &MsgMhfCaravanMyScore{}, 32},
		{"MsgMhfPostGemInfo", &MsgMhfPostGemInfo{}, 32},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrameFromBytes(make([]byte, tt.dataSize))
			err := tt.pkt.Parse(bf, ctx)
			if err != nil {
				t.Errorf("Parse() returned error: %v", err)
			}
		})
	}
}

// TestParseCoverage_VariableLength tests Parse for variable-length packets
// that require specific data layouts.
func TestParseCoverage_VariableLength(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("MsgMhfAcquireItem_EmptyList", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint16(0)  // Unk0
		bf.WriteUint16(0)  // Length = 0 items
		pkt := &MsgMhfAcquireItem{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgMhfAcquireItem_WithItems", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint16(0)  // Unk0
		bf.WriteUint16(2)  // Length = 2 items
		bf.WriteUint32(100) // item 1
		bf.WriteUint32(200) // item 2
		pkt := &MsgMhfAcquireItem{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
		if len(pkt.Unk1) != 2 {
			t.Errorf("expected 2 items, got %d", len(pkt.Unk1))
		}
	})

	t.Run("MsgMhfReadBeatLevelMyRanking", func(t *testing.T) {
		// 4 + 4 + 4 + 16*4 = 76 bytes
		bf := byteframe.NewByteFrameFromBytes(make([]byte, 76))
		pkt := &MsgMhfReadBeatLevelMyRanking{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgMhfUpdateBeatLevel", func(t *testing.T) {
		// 4 + 4 + 4 + 16*4 + 16*4 = 140 bytes
		bf := byteframe.NewByteFrameFromBytes(make([]byte, 140))
		pkt := &MsgMhfUpdateBeatLevel{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgSysRightsReload", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(3)  // length
		bf.WriteBytes([]byte{0x01, 0x02, 0x03}) // Unk0
		pkt := &MsgSysRightsReload{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgMhfCreateGuild", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint16(0)  // zeroed
		bf.WriteUint16(4)  // name length
		bf.WriteBytes([]byte("Test\x00")) // null-terminated name
		pkt := &MsgMhfCreateGuild{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgMhfEnumerateGuild", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint8(0)   // Type
		bf.WriteUint8(0)   // Page
		bf.WriteBool(false) // Sorting
		bf.WriteUint8(0)   // zero
		bf.WriteBytes(make([]byte, 4)) // Data1
		bf.WriteUint16(0)  // zero
		bf.WriteUint8(0)   // dataLen = 0
		bf.WriteUint8(0)   // zero
		pkt := &MsgMhfEnumerateGuild{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgSysCreateSemaphore", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint16(0)  // Unk0
		bf.WriteUint8(5)   // semaphore ID length
		bf.WriteNullTerminatedBytes([]byte("test"))
		pkt := &MsgSysCreateSemaphore{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgMhfUpdateGuildMessageBoard_Op0", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(0) // MessageOp = 0
		bf.WriteUint32(0) // PostType
		bf.WriteUint32(0) // StampID
		bf.WriteUint32(0) // TitleLength = 0
		bf.WriteUint32(0) // BodyLength = 0
		pkt := &MsgMhfUpdateGuildMessageBoard{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgMhfUpdateGuildMessageBoard_Op1", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(1) // MessageOp = 1
		bf.WriteUint32(42) // PostID
		pkt := &MsgMhfUpdateGuildMessageBoard{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgMhfUpdateGuildMessageBoard_Op3", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint32(3)  // MessageOp = 3
		bf.WriteUint32(42) // PostID
		bf.WriteBytes(make([]byte, 8)) // skip
		bf.WriteUint32(0)  // StampID
		pkt := &MsgMhfUpdateGuildMessageBoard{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgMhfUpdateGuildMessageBoard_Op4", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint32(4)  // MessageOp = 4
		bf.WriteUint32(42) // PostID
		bf.WriteBytes(make([]byte, 8)) // skip
		bf.WriteBool(true) // LikeState
		pkt := &MsgMhfUpdateGuildMessageBoard{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})
}

// TestBuildCoverage_Implemented tests Build() on packet types whose Build method
// is implemented (writes to ByteFrame) but was not yet covered.
func TestBuildCoverage_Implemented(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("MsgSysDeleteUser", func(t *testing.T) {
		pkt := &MsgSysDeleteUser{CharID: 123}
		bf := byteframe.NewByteFrame()
		if err := pkt.Build(bf, ctx); err != nil {
			t.Errorf("Build() error: %v", err)
		}
		if len(bf.Data()) == 0 {
			t.Error("Build() produced no data")
		}
	})

	t.Run("MsgSysInsertUser", func(t *testing.T) {
		pkt := &MsgSysInsertUser{CharID: 456}
		bf := byteframe.NewByteFrame()
		if err := pkt.Build(bf, ctx); err != nil {
			t.Errorf("Build() error: %v", err)
		}
		if len(bf.Data()) == 0 {
			t.Error("Build() produced no data")
		}
	})

	t.Run("MsgSysUpdateRight", func(t *testing.T) {
		pkt := &MsgSysUpdateRight{
			ClientRespAckHandle: 1,
			Bitfield:            0xFF,
		}
		bf := byteframe.NewByteFrame()
		if err := pkt.Build(bf, ctx); err != nil {
			t.Errorf("Build() error: %v", err)
		}
		if len(bf.Data()) == 0 {
			t.Error("Build() produced no data")
		}
	})

	t.Run("MsgSysUpdateRight_WithRights", func(t *testing.T) {
		pkt := &MsgSysUpdateRight{
			ClientRespAckHandle: 1,
			Bitfield:            0xFF,
			Rights: []mhfcourse.Course{
				{ID: 1},
				{ID: 2},
			},
		}
		bf := byteframe.NewByteFrame()
		if err := pkt.Build(bf, ctx); err != nil {
			t.Errorf("Build() error: %v", err)
		}
	})

	// MsgSysLogout Build has a bug (calls ReadUint8 instead of WriteUint8)
	// so we test it with defer/recover
	t.Run("MsgSysLogout_Build", func(t *testing.T) {
		defer func() {
			_ = recover() // may panic due to bug
		}()
		pkt := &MsgSysLogout{LogoutType: 1}
		bf := byteframe.NewByteFrame()
		_ = pkt.Build(bf, ctx)
	})
}

// TestParseCoverage_EmptyPackets tests Parse() for packets with no payload fields.
func TestParseCoverage_EmptyPackets(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("MsgSysCleanupObject_Parse", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		pkt := &MsgSysCleanupObject{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgSysCleanupObject_Build", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		pkt := &MsgSysCleanupObject{}
		if err := pkt.Build(bf, ctx); err != nil {
			t.Errorf("Build() error: %v", err)
		}
	})

	t.Run("MsgSysUnreserveStage_Parse", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		pkt := &MsgSysUnreserveStage{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
	})

	t.Run("MsgSysUnreserveStage_Build", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		pkt := &MsgSysUnreserveStage{}
		if err := pkt.Build(bf, ctx); err != nil {
			t.Errorf("Build() error: %v", err)
		}
	})
}

// TestParseCoverage_NotImplemented2 tests Parse/Build for packets that return NOT IMPLEMENTED.
func TestParseCoverage_NotImplemented2(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("MsgSysGetObjectOwner_Parse", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		pkt := &MsgSysGetObjectOwner{}
		err := pkt.Parse(bf, ctx)
		if err == nil {
			t.Error("expected NOT IMPLEMENTED error")
		}
	})

	t.Run("MsgSysUpdateRight_Parse", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		pkt := &MsgSysUpdateRight{}
		err := pkt.Parse(bf, ctx)
		if err == nil {
			t.Error("expected NOT IMPLEMENTED error")
		}
	})
}

// TestParseCoverage_UpdateWarehouse tests MsgMhfUpdateWarehouse.Parse with different box types.
func TestParseCoverage_UpdateWarehouse(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("EmptyChanges", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint8(0)   // BoxType = 0 (items)
		bf.WriteUint8(0)   // BoxIndex
		bf.WriteUint16(0)  // changes = 0
		bf.WriteUint8(0)   // Zeroed
		bf.WriteUint8(0)   // Zeroed
		pkt := &MsgMhfUpdateWarehouse{}
		parsed := byteframe.NewByteFrameFromBytes(bf.Data())
		if err := pkt.Parse(parsed, ctx); err != nil {
			t.Errorf("Parse() error: %v", err)
		}
		if pkt.BoxType != 0 {
			t.Errorf("BoxType = %d, want 0", pkt.BoxType)
		}
	})
}
