package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
	"erupe-ce/network/clientctx"
)

// TestParseSmallNotImplemented tests Parse for packets whose Parse method returns
// "NOT IMPLEMENTED". We verify that Parse returns a non-nil error and does not panic.
func TestParseSmallNotImplemented(t *testing.T) {
	packets := []struct {
		name string
		pkt  MHFPacket
	}{
		// MHF packets - NOT IMPLEMENTED
		{"MsgMhfAcceptReadReward", &MsgMhfAcceptReadReward{}},
		{"MsgMhfAddRewardSongCount", &MsgMhfAddRewardSongCount{}},
		{"MsgMhfDebugPostValue", &MsgMhfDebugPostValue{}},
		{"MsgMhfEnterTournamentQuest", &MsgMhfEnterTournamentQuest{}},
		{"MsgMhfGetCaAchievementHist", &MsgMhfGetCaAchievementHist{}},
		{"MsgMhfGetCaUniqueID", &MsgMhfGetCaUniqueID{}},
		{"MsgMhfGetDailyMissionMaster", &MsgMhfGetDailyMissionMaster{}},
		{"MsgMhfGetDailyMissionPersonal", &MsgMhfGetDailyMissionPersonal{}},
		{"MsgMhfGetExtraInfo", &MsgMhfGetExtraInfo{}},
		{"MsgMhfGetRestrictionEvent", &MsgMhfGetRestrictionEvent{}},
		{"MsgMhfKickExportForce", &MsgMhfKickExportForce{}},
		{"MsgMhfPaymentAchievement", &MsgMhfPaymentAchievement{}},
		{"MsgMhfPostRyoudama", &MsgMhfPostRyoudama{}},
		{"MsgMhfRegistSpabiTime", &MsgMhfRegistSpabiTime{}},
		{"MsgMhfResetAchievement", &MsgMhfResetAchievement{}},
		{"MsgMhfResetTitle", &MsgMhfResetTitle{}},
		{"MsgMhfSetCaAchievement", &MsgMhfSetCaAchievement{}},
		{"MsgMhfSetDailyMissionPersonal", &MsgMhfSetDailyMissionPersonal{}},
		{"MsgMhfSetUdTacticsFollower", &MsgMhfSetUdTacticsFollower{}},
		{"MsgMhfStampcardPrize", &MsgMhfStampcardPrize{}},
		{"MsgMhfUpdateForceGuildRank", &MsgMhfUpdateForceGuildRank{}},
		{"MsgMhfUseUdShopCoin", &MsgMhfUseUdShopCoin{}},

		// SYS packets - NOT IMPLEMENTED
		{"MsgSysAuthData", &MsgSysAuthData{}},
		{"MsgSysAuthQuery", &MsgSysAuthQuery{}},
		{"MsgSysAuthTerminal", &MsgSysAuthTerminal{}},
		{"MsgSysCloseMutex", &MsgSysCloseMutex{}},
		{"MsgSysCollectBinary", &MsgSysCollectBinary{}},
		{"MsgSysCreateMutex", &MsgSysCreateMutex{}},
		{"MsgSysCreateOpenMutex", &MsgSysCreateOpenMutex{}},
		{"MsgSysDeleteMutex", &MsgSysDeleteMutex{}},
		{"MsgSysEnumlobby", &MsgSysEnumlobby{}},
		{"MsgSysEnumuser", &MsgSysEnumuser{}},
		{"MsgSysGetObjectBinary", &MsgSysGetObjectBinary{}},
		{"MsgSysGetState", &MsgSysGetState{}},
		{"MsgSysInfokyserver", &MsgSysInfokyserver{}},
		{"MsgSysOpenMutex", &MsgSysOpenMutex{}},
		{"MsgSysRotateObject", &MsgSysRotateObject{}},
		{"MsgSysSerialize", &MsgSysSerialize{}},
		{"MsgSysTransBinary", &MsgSysTransBinary{}},
	}

	ctx := &clientctx.ClientContext{RealClientMode: _config.ZZ}
	for _, tc := range packets {
		t.Run(tc.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			// Write some padding bytes so Parse has data available if it tries to read.
			bf.WriteUint32(0)
			_, _ = bf.Seek(0, io.SeekStart)

			err := tc.pkt.Parse(bf, ctx)
			if err == nil {
				t.Fatalf("Parse() expected error for NOT IMPLEMENTED packet, got nil")
			}
			if err.Error() != "NOT IMPLEMENTED" {
				t.Fatalf("Parse() error = %q, want %q", err.Error(), "NOT IMPLEMENTED")
			}
		})
	}
}

// TestParseSmallNoData tests Parse for packets with no fields that return nil.
func TestParseSmallNoData(t *testing.T) {
	packets := []struct {
		name string
		pkt  MHFPacket
	}{
		{"MsgSysCleanupObject", &MsgSysCleanupObject{}},
		{"MsgSysUnreserveStage", &MsgSysUnreserveStage{}},
	}

	ctx := &clientctx.ClientContext{RealClientMode: _config.ZZ}
	for _, tc := range packets {
		t.Run(tc.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			err := tc.pkt.Parse(bf, ctx)
			if err != nil {
				t.Fatalf("Parse() error = %v, want nil", err)
			}
		})
	}
}

// TestParseSmallLogout tests Parse for MsgSysLogout which reads a single uint8 field.
func TestParseSmallLogout(t *testing.T) {
	tests := []struct {
		name string
		unk0 uint8
	}{
		{"hardcoded 1", 1},
		{"zero", 0},
		{"max", 255},
	}

	ctx := &clientctx.ClientContext{RealClientMode: _config.ZZ}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint8(tt.unk0)
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgSysLogout{}
			err := pkt.Parse(bf, ctx)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			if pkt.LogoutType != tt.unk0 {
				t.Errorf("Unk0 = %d, want %d", pkt.LogoutType, tt.unk0)
			}
		})
	}
}

// TestParseSmallEnumerateHouse tests Parse for MsgMhfEnumerateHouse which reads
// AckHandle, CharID, Method, Unk, lenName, and optional Name.
func TestParseSmallEnumerateHouse(t *testing.T) {
	ctx := &clientctx.ClientContext{RealClientMode: _config.ZZ}

	t.Run("no name", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(0x11223344) // AckHandle
		bf.WriteUint32(0xDEADBEEF) // CharID
		bf.WriteUint8(2)           // Method
		bf.WriteUint16(100)        // Unk
		bf.WriteUint8(0)           // lenName = 0 (no name)
		_, _ = bf.Seek(0, io.SeekStart)

		pkt := &MsgMhfEnumerateHouse{}
		err := pkt.Parse(bf, ctx)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if pkt.AckHandle != 0x11223344 {
			t.Errorf("AckHandle = 0x%X, want 0x11223344", pkt.AckHandle)
		}
		if pkt.CharID != 0xDEADBEEF {
			t.Errorf("CharID = 0x%X, want 0xDEADBEEF", pkt.CharID)
		}
		if pkt.Method != 2 {
			t.Errorf("Method = %d, want 2", pkt.Method)
		}
		if pkt.Name != "" {
			t.Errorf("Name = %q, want empty", pkt.Name)
		}
	})

	t.Run("with name", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint32(42)  // CharID
		bf.WriteUint8(1)    // Method
		bf.WriteUint16(200) // Unk
		// The name is SJIS null-terminated bytes. Use ASCII-compatible bytes.
		nameBytes := []byte("Test\x00")
		bf.WriteUint8(uint8(len(nameBytes))) // lenName > 0
		bf.WriteBytes(nameBytes)             // null-terminated name
		_, _ = bf.Seek(0, io.SeekStart)

		pkt := &MsgMhfEnumerateHouse{}
		err := pkt.Parse(bf, ctx)
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}
		if pkt.AckHandle != 1 {
			t.Errorf("AckHandle = %d, want 1", pkt.AckHandle)
		}
		if pkt.CharID != 42 {
			t.Errorf("CharID = %d, want 42", pkt.CharID)
		}
		if pkt.Method != 1 {
			t.Errorf("Method = %d, want 1", pkt.Method)
		}
		if pkt.Name != "Test" {
			t.Errorf("Name = %q, want %q", pkt.Name, "Test")
		}
	})
}

// TestParseSmallNotImplementedDoesNotPanic ensures that calling Parse on NOT IMPLEMENTED
// packets returns an error and does not panic.
func TestParseSmallNotImplementedDoesNotPanic(t *testing.T) {
	packets := []MHFPacket{
		&MsgMhfAcceptReadReward{},
		&MsgSysAuthData{},
		&MsgSysSerialize{},
	}

	ctx := &clientctx.ClientContext{RealClientMode: _config.ZZ}
	for _, pkt := range packets {
		t.Run("not_implemented", func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			err := pkt.Parse(bf, ctx)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}
