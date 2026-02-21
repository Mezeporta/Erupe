package mhfpacket

import (
	"bytes"
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	cfg "erupe-ce/config"
	"erupe-ce/network/clientctx"
)

// --- 5-stmt packets (medium complexity) ---

// TestParseMediumVoteFesta verifies Parse for MsgMhfVoteFesta.
// Fields: AckHandle(u32), FestaID(u32), GuildID(u32), TrialID(u32)
func TestParseMediumVoteFesta(t *testing.T) {
	tests := []struct {
		name    string
		ack     uint32
		festaID uint32
		guildID uint32
		trialID uint32
	}{
		{"typical", 0x11223344, 1, 500, 42},
		{"zero", 0, 0, 0, 0},
		{"max", 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ack)
			bf.WriteUint32(tt.festaID)
			bf.WriteUint32(tt.guildID)
			bf.WriteUint32(tt.trialID)

			_, _ = bf.Seek(0, io.SeekStart)
			pkt := &MsgMhfVoteFesta{}
			if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ack {
				t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ack)
			}
			if pkt.FestaID != tt.festaID {
				t.Errorf("FestaID = %d, want %d", pkt.FestaID, tt.festaID)
			}
			if pkt.GuildID != tt.guildID {
				t.Errorf("GuildID = %d, want %d", pkt.GuildID, tt.guildID)
			}
			if pkt.TrialID != tt.trialID {
				t.Errorf("TrialID = %d, want %d", pkt.TrialID, tt.trialID)
			}
		})
	}
}

// TestParseMediumAcquireSemaphore verifies Parse for MsgSysAcquireSemaphore.
// Fields: AckHandle(u32), SemaphoreIDLength(u8), SemaphoreID(string via bfutil.UpToNull)
func TestParseMediumAcquireSemaphore(t *testing.T) {
	tests := []struct {
		name        string
		ack         uint32
		semaphoreID string
	}{
		{"typical", 0xAABBCCDD, "quest_semaphore"},
		{"short", 1, "s"},
		{"empty", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ack)
			// SemaphoreIDLength includes the null terminator in the read
			idBytes := []byte(tt.semaphoreID)
			idBytes = append(idBytes, 0x00) // null terminator
			bf.WriteUint8(uint8(len(idBytes)))
			bf.WriteBytes(idBytes)

			_, _ = bf.Seek(0, io.SeekStart)
			pkt := &MsgSysAcquireSemaphore{}
			if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ack {
				t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ack)
			}
			if pkt.SemaphoreID != tt.semaphoreID {
				t.Errorf("SemaphoreID = %q, want %q", pkt.SemaphoreID, tt.semaphoreID)
			}
		})
	}
}

// TestParseMediumCheckSemaphore verifies Parse for MsgSysCheckSemaphore.
// Fields: AckHandle(u32), semaphoreIDLength(u8), SemaphoreID(string via bfutil.UpToNull)
func TestParseMediumCheckSemaphore(t *testing.T) {
	tests := []struct {
		name        string
		ack         uint32
		semaphoreID string
	}{
		{"typical", 0x12345678, "global_semaphore"},
		{"short id", 42, "x"},
		{"empty id", 0, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ack)
			idBytes := []byte(tt.semaphoreID)
			idBytes = append(idBytes, 0x00)
			bf.WriteUint8(uint8(len(idBytes)))
			bf.WriteBytes(idBytes)

			_, _ = bf.Seek(0, io.SeekStart)
			pkt := &MsgSysCheckSemaphore{}
			if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ack {
				t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ack)
			}
			if pkt.SemaphoreID != tt.semaphoreID {
				t.Errorf("SemaphoreID = %q, want %q", pkt.SemaphoreID, tt.semaphoreID)
			}
		})
	}
}

// TestParseMediumGetUserBinary verifies Parse for MsgSysGetUserBinary.
// Fields: AckHandle(u32), CharID(u32), BinaryType(u8)
func TestParseMediumGetUserBinary(t *testing.T) {
	tests := []struct {
		name       string
		ack        uint32
		charID     uint32
		binaryType uint8
	}{
		{"typical", 0xDEADBEEF, 12345, 1},
		{"zero", 0, 0, 0},
		{"max", 0xFFFFFFFF, 0xFFFFFFFF, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ack)
			bf.WriteUint32(tt.charID)
			bf.WriteUint8(tt.binaryType)

			_, _ = bf.Seek(0, io.SeekStart)
			pkt := &MsgSysGetUserBinary{}
			if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ack {
				t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ack)
			}
			if pkt.CharID != tt.charID {
				t.Errorf("CharID = %d, want %d", pkt.CharID, tt.charID)
			}
			if pkt.BinaryType != tt.binaryType {
				t.Errorf("BinaryType = %d, want %d", pkt.BinaryType, tt.binaryType)
			}
		})
	}
}

// TestParseMediumSetObjectBinary verifies Parse for MsgSysSetObjectBinary.
// Fields: ObjID(u32), DataSize(u16), RawDataPayload([]byte of DataSize)
func TestParseMediumSetObjectBinary(t *testing.T) {
	tests := []struct {
		name    string
		objID   uint32
		payload []byte
	}{
		{"typical", 42, []byte{0x01, 0x02, 0x03, 0x04}},
		{"empty", 0, []byte{}},
		{"large", 0xCAFEBABE, []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF, 0x11, 0x22, 0x33, 0x44}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.objID)
			bf.WriteUint16(uint16(len(tt.payload)))
			bf.WriteBytes(tt.payload)

			_, _ = bf.Seek(0, io.SeekStart)
			pkt := &MsgSysSetObjectBinary{}
			if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.ObjID != tt.objID {
				t.Errorf("ObjID = %d, want %d", pkt.ObjID, tt.objID)
			}
			if pkt.DataSize != uint16(len(tt.payload)) {
				t.Errorf("DataSize = %d, want %d", pkt.DataSize, len(tt.payload))
			}
			if !bytes.Equal(pkt.RawDataPayload, tt.payload) {
				t.Errorf("RawDataPayload = %v, want %v", pkt.RawDataPayload, tt.payload)
			}
		})
	}
}

// TestParseMediumSetUserBinary verifies Parse for MsgSysSetUserBinary.
// Fields: BinaryType(u8), DataSize(u16), RawDataPayload([]byte of DataSize)
func TestParseMediumSetUserBinary(t *testing.T) {
	tests := []struct {
		name       string
		binaryType uint8
		payload    []byte
	}{
		{"typical", 1, []byte{0xDE, 0xAD, 0xBE, 0xEF}},
		{"empty", 0, []byte{}},
		{"max type", 255, []byte{0x01}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(tt.binaryType)
			bf.WriteUint16(uint16(len(tt.payload)))
			bf.WriteBytes(tt.payload)

			_, _ = bf.Seek(0, io.SeekStart)
			pkt := &MsgSysSetUserBinary{}
			if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.BinaryType != tt.binaryType {
				t.Errorf("BinaryType = %d, want %d", pkt.BinaryType, tt.binaryType)
			}
			if pkt.DataSize != uint16(len(tt.payload)) {
				t.Errorf("DataSize = %d, want %d", pkt.DataSize, len(tt.payload))
			}
			if !bytes.Equal(pkt.RawDataPayload, tt.payload) {
				t.Errorf("RawDataPayload = %v, want %v", pkt.RawDataPayload, tt.payload)
			}
		})
	}
}

// --- 4-stmt packets ---

// TestParseMediumGetUdRanking verifies Parse for MsgMhfGetUdRanking.
// Fields: AckHandle(u32), Unk0(u8)
func TestParseMediumGetUdRanking(t *testing.T) {
	tests := []struct {
		name string
		ack  uint32
		unk0 uint8
	}{
		{"typical", 0x11223344, 5},
		{"zero", 0, 0},
		{"max", 0xFFFFFFFF, 255},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ack)
			bf.WriteUint8(tt.unk0)

			_, _ = bf.Seek(0, io.SeekStart)
			pkt := &MsgMhfGetUdRanking{}
			if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ack {
				t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ack)
			}
			if pkt.Unk0 != tt.unk0 {
				t.Errorf("Unk0 = %d, want %d", pkt.Unk0, tt.unk0)
			}
		})
	}
}

// TestParseMediumGetUdTacticsRanking verifies Parse for MsgMhfGetUdTacticsRanking.
// Fields: AckHandle(u32), GuildID(u32)
func TestParseMediumGetUdTacticsRanking(t *testing.T) {
	tests := []struct {
		name    string
		ack     uint32
		guildID uint32
	}{
		{"typical", 0xAABBCCDD, 500},
		{"zero", 0, 0},
		{"max", 0xFFFFFFFF, 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ack)
			bf.WriteUint32(tt.guildID)

			_, _ = bf.Seek(0, io.SeekStart)
			pkt := &MsgMhfGetUdTacticsRanking{}
			if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ack {
				t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ack)
			}
			if pkt.GuildID != tt.guildID {
				t.Errorf("GuildID = %d, want %d", pkt.GuildID, tt.guildID)
			}
		})
	}
}

// TestParseMediumRegistGuildTresure verifies Parse for MsgMhfRegistGuildTresure.
// Fields: AckHandle(u32), DataLen(u16), Data([]byte), trailing u32 (discarded)
func TestParseMediumRegistGuildTresure(t *testing.T) {
	tests := []struct {
		name string
		ack  uint32
		data []byte
	}{
		{"typical", 0x12345678, []byte{0x01, 0x02, 0x03}},
		{"empty data", 1, []byte{}},
		{"larger data", 0xDEADBEEF, []byte{0xAA, 0xBB, 0xCC, 0xDD, 0xEE, 0xFF}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ack)
			bf.WriteUint16(uint16(len(tt.data)))
			bf.WriteBytes(tt.data)
			bf.WriteUint32(0) // trailing uint32 that is read and discarded

			_, _ = bf.Seek(0, io.SeekStart)
			pkt := &MsgMhfRegistGuildTresure{}
			if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ack {
				t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, tt.ack)
			}
			if !bytes.Equal(pkt.Data, tt.data) {
				t.Errorf("Data = %v, want %v", pkt.Data, tt.data)
			}
		})
	}
}

// TestParseMediumUpdateMyhouseInfo verifies Parse for MsgMhfUpdateMyhouseInfo.
// Fields: AckHandle(u32), Unk0([]byte of 0x16A bytes)
func TestParseMediumUpdateMyhouseInfo(t *testing.T) {
	t.Run("typical", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		ack := uint32(0xCAFEBABE)
		bf.WriteUint32(ack)

		// 0x16A = 362 bytes
		payload := make([]byte, 0x16A)
		for i := range payload {
			payload[i] = byte(i % 256)
		}
		bf.WriteBytes(payload)

		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfUpdateMyhouseInfo{}
		if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
		if len(pkt.Data) != 0x16A {
			t.Errorf("Unk0 length = %d, want %d", len(pkt.Data), 0x16A)
		}
		if !bytes.Equal(pkt.Data, payload) {
			t.Error("Unk0 content mismatch")
		}
	})

	t.Run("zero values", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(0)
		bf.WriteBytes(make([]byte, 0x16A))

		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfUpdateMyhouseInfo{}
		if err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ}); err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if pkt.AckHandle != 0 {
			t.Errorf("AckHandle = 0x%X, want 0", pkt.AckHandle)
		}
		if len(pkt.Data) != 0x16A {
			t.Errorf("Unk0 length = %d, want %d", len(pkt.Data), 0x16A)
		}
	})
}


// --- 3-stmt packets (AckHandle-only Parse) ---

// TestParseMediumAckHandleOnlyBatch tests Parse for all 3-stmt packets that only
// read a single AckHandle uint32. These are verified to parse correctly and
// return the expected AckHandle value.
func TestParseMediumAckHandleOnlyBatch(t *testing.T) {
	packets := []struct {
		name string
		pkt  MHFPacket
		// getAck extracts the AckHandle from the parsed packet
		getAck func() uint32
	}{
		{
			"MsgMhfGetUdBonusQuestInfo",
			&MsgMhfGetUdBonusQuestInfo{},
			nil,
		},
		{
			"MsgMhfGetUdDailyPresentList",
			&MsgMhfGetUdDailyPresentList{},
			nil,
		},
		{
			"MsgMhfGetUdGuildMapInfo",
			&MsgMhfGetUdGuildMapInfo{},
			nil,
		},
		{
			"MsgMhfGetUdMonsterPoint",
			&MsgMhfGetUdMonsterPoint{},
			nil,
		},
		{
			"MsgMhfGetUdMyRanking",
			&MsgMhfGetUdMyRanking{},
			nil,
		},
		{
			"MsgMhfGetUdNormaPresentList",
			&MsgMhfGetUdNormaPresentList{},
			nil,
		},
		{
			"MsgMhfGetUdRankingRewardList",
			&MsgMhfGetUdRankingRewardList{},
			nil,
		},
		{
			"MsgMhfGetUdSelectedColorInfo",
			&MsgMhfGetUdSelectedColorInfo{},
			nil,
		},
		{
			"MsgMhfGetUdShopCoin",
			&MsgMhfGetUdShopCoin{},
			nil,
		},
		{
			"MsgMhfGetUdTacticsBonusQuest",
			&MsgMhfGetUdTacticsBonusQuest{},
			nil,
		},
		{
			"MsgMhfGetUdTacticsFirstQuestBonus",
			&MsgMhfGetUdTacticsFirstQuestBonus{},
			nil,
		},
		{
			"MsgMhfGetUdTacticsFollower",
			&MsgMhfGetUdTacticsFollower{},
			nil,
		},
		{
			"MsgMhfGetUdTacticsLog",
			&MsgMhfGetUdTacticsLog{},
			nil,
		},
		{
			"MsgMhfGetUdTacticsPoint",
			&MsgMhfGetUdTacticsPoint{},
			nil,
		},
		{
			"MsgMhfGetUdTacticsRewardList",
			&MsgMhfGetUdTacticsRewardList{},
			nil,
		},
		{
			"MsgMhfReceiveCafeDurationBonus",
			&MsgMhfReceiveCafeDurationBonus{},
			nil,
		},
		{
			"MsgSysDeleteSemaphore",
			&MsgSysDeleteSemaphore{},
			nil,
		},
		{
			"MsgSysReleaseSemaphore",
			&MsgSysReleaseSemaphore{},
			nil,
		},
	}

	ctx := &clientctx.ClientContext{RealClientMode: cfg.ZZ}
	ackValues := []uint32{0x12345678, 0, 0xFFFFFFFF, 0xDEADBEEF}

	for _, tc := range packets {
		for _, ackVal := range ackValues {
			t.Run(tc.name+"/ack_"+ackHex(ackVal), func(t *testing.T) {
				bf := byteframe.NewByteFrame()
				bf.WriteUint32(ackVal)
				_, _ = bf.Seek(0, io.SeekStart)

				err := tc.pkt.Parse(bf, ctx)
				if err != nil {
					t.Fatalf("Parse() error = %v", err)
				}
			})
		}
	}
}

// TestParseMediumAckHandleOnlyVerifyValues tests each 3-stmt AckHandle-only
// packet individually, verifying that the AckHandle field is correctly populated.
func TestParseMediumAckHandleOnlyVerifyValues(t *testing.T) {
	ctx := &clientctx.ClientContext{RealClientMode: cfg.ZZ}
	ack := uint32(0xCAFEBABE)

	makeFrame := func() *byteframe.ByteFrame {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(ack)
		_, _ = bf.Seek(0, io.SeekStart)
		return bf
	}

	t.Run("MsgMhfGetUdBonusQuestInfo", func(t *testing.T) {
		pkt := &MsgMhfGetUdBonusQuestInfo{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdDailyPresentList", func(t *testing.T) {
		pkt := &MsgMhfGetUdDailyPresentList{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdGuildMapInfo", func(t *testing.T) {
		pkt := &MsgMhfGetUdGuildMapInfo{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdMonsterPoint", func(t *testing.T) {
		pkt := &MsgMhfGetUdMonsterPoint{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdMyRanking", func(t *testing.T) {
		pkt := &MsgMhfGetUdMyRanking{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdNormaPresentList", func(t *testing.T) {
		pkt := &MsgMhfGetUdNormaPresentList{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdRankingRewardList", func(t *testing.T) {
		pkt := &MsgMhfGetUdRankingRewardList{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdSelectedColorInfo", func(t *testing.T) {
		pkt := &MsgMhfGetUdSelectedColorInfo{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdShopCoin", func(t *testing.T) {
		pkt := &MsgMhfGetUdShopCoin{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdTacticsBonusQuest", func(t *testing.T) {
		pkt := &MsgMhfGetUdTacticsBonusQuest{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdTacticsFirstQuestBonus", func(t *testing.T) {
		pkt := &MsgMhfGetUdTacticsFirstQuestBonus{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdTacticsFollower", func(t *testing.T) {
		pkt := &MsgMhfGetUdTacticsFollower{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdTacticsLog", func(t *testing.T) {
		pkt := &MsgMhfGetUdTacticsLog{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdTacticsPoint", func(t *testing.T) {
		pkt := &MsgMhfGetUdTacticsPoint{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfGetUdTacticsRewardList", func(t *testing.T) {
		pkt := &MsgMhfGetUdTacticsRewardList{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgMhfReceiveCafeDurationBonus", func(t *testing.T) {
		pkt := &MsgMhfReceiveCafeDurationBonus{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})

	t.Run("MsgSysDeleteSemaphore", func(t *testing.T) {
		pkt := &MsgSysDeleteSemaphore{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.SemaphoreID != ack {
			t.Errorf("SemaphoreID = 0x%X, want 0x%X", pkt.SemaphoreID, ack)
		}
	})

	t.Run("MsgSysReleaseSemaphore", func(t *testing.T) {
		pkt := &MsgSysReleaseSemaphore{}
		if err := pkt.Parse(makeFrame(), ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != ack {
			t.Errorf("AckHandle = 0x%X, want 0x%X", pkt.AckHandle, ack)
		}
	})
}

// TestParseMediumDeleteUser verifies that MsgSysDeleteUser.Parse returns
// NOT IMPLEMENTED error (Parse is not implemented, only Build is).
func TestParseMediumDeleteUser(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(12345)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysDeleteUser{}
	err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ})
	if err == nil {
		t.Fatal("Parse() should return error for NOT IMPLEMENTED")
	}
	if err.Error() != "NOT IMPLEMENTED" {
		t.Errorf("Parse() error = %q, want %q", err.Error(), "NOT IMPLEMENTED")
	}
}

// TestParseMediumInsertUser verifies that MsgSysInsertUser.Parse returns
// NOT IMPLEMENTED error (Parse is not implemented, only Build is).
func TestParseMediumInsertUser(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(12345)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysInsertUser{}
	err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: cfg.ZZ})
	if err == nil {
		t.Fatal("Parse() should return error for NOT IMPLEMENTED")
	}
	if err.Error() != "NOT IMPLEMENTED" {
		t.Errorf("Parse() error = %q, want %q", err.Error(), "NOT IMPLEMENTED")
	}
}

// ackHex returns a hex string for a uint32 ack value, used for test naming.
func ackHex(v uint32) string {
	const hex = "0123456789ABCDEF"
	buf := make([]byte, 8)
	for i := 7; i >= 0; i-- {
		buf[i] = hex[v&0xF]
		v >>= 4
	}
	return string(buf)
}
