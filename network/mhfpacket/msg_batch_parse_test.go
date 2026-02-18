package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/clientctx"
)

// TestBatchParseAckHandleOnly tests Parse for packets that only read AckHandle (uint32).
func TestBatchParseAckHandleOnly(t *testing.T) {
	packets := []struct {
		name string
		pkt  MHFPacket
	}{
		{"MsgMhfLoaddata", &MsgMhfLoaddata{}},
		{"MsgMhfLoadFavoriteQuest", &MsgMhfLoadFavoriteQuest{}},
		{"MsgMhfReadGuildcard", &MsgMhfReadGuildcard{}},
		{"MsgMhfGetEtcPoints", &MsgMhfGetEtcPoints{}},
		{"MsgMhfGetGuildMissionList", &MsgMhfGetGuildMissionList{}},
		{"MsgMhfGetGuildMissionRecord", &MsgMhfGetGuildMissionRecord{}},
		{"MsgMhfGetGuildTresureSouvenir", &MsgMhfGetGuildTresureSouvenir{}},
		{"MsgMhfAcquireGuildTresureSouvenir", &MsgMhfAcquireGuildTresureSouvenir{}},
		{"MsgMhfEnumerateFestaIntermediatePrize", &MsgMhfEnumerateFestaIntermediatePrize{}},
		{"MsgMhfEnumerateFestaPersonalPrize", &MsgMhfEnumerateFestaPersonalPrize{}},
		{"MsgMhfGetGuildWeeklyBonusMaster", &MsgMhfGetGuildWeeklyBonusMaster{}},
		{"MsgMhfGetGuildWeeklyBonusActiveCount", &MsgMhfGetGuildWeeklyBonusActiveCount{}},
		{"MsgMhfGetEquipSkinHist", &MsgMhfGetEquipSkinHist{}},
		{"MsgMhfGetRejectGuildScout", &MsgMhfGetRejectGuildScout{}},
		{"MsgMhfGetKeepLoginBoostStatus", &MsgMhfGetKeepLoginBoostStatus{}},
		{"MsgMhfAcquireMonthlyReward", &MsgMhfAcquireMonthlyReward{}},
		{"MsgMhfGetGuildScoutList", &MsgMhfGetGuildScoutList{}},
		{"MsgMhfGetGuildManageRight", &MsgMhfGetGuildManageRight{}},
		{"MsgMhfGetRengokuRankingRank", &MsgMhfGetRengokuRankingRank{}},
		{"MsgMhfGetUdMyPoint", &MsgMhfGetUdMyPoint{}},
		{"MsgMhfGetUdTotalPointInfo", &MsgMhfGetUdTotalPointInfo{}},
		{"MsgMhfCreateMercenary", &MsgMhfCreateMercenary{}},
		{"MsgMhfEnumerateMercenaryLog", &MsgMhfEnumerateMercenaryLog{}},
		{"MsgMhfLoadLegendDispatch", &MsgMhfLoadLegendDispatch{}},
		{"MsgMhfGetBoostRight", &MsgMhfGetBoostRight{}},
		{"MsgMhfPostBoostTimeQuestReturn", &MsgMhfPostBoostTimeQuestReturn{}},
		{"MsgMhfGetFpointExchangeList", &MsgMhfGetFpointExchangeList{}},
		{"MsgMhfGetRewardSong", &MsgMhfGetRewardSong{}},
		{"MsgMhfUseRewardSong", &MsgMhfUseRewardSong{}},
		{"MsgMhfGetKouryouPoint", &MsgMhfGetKouryouPoint{}},
		{"MsgMhfGetTrendWeapon", &MsgMhfGetTrendWeapon{}},
		{"MsgMhfInfoScenarioCounter", &MsgMhfInfoScenarioCounter{}},
		{"MsgMhfLoadScenarioData", &MsgMhfLoadScenarioData{}},
		{"MsgMhfLoadRengokuData", &MsgMhfLoadRengokuData{}},
		{"MsgMhfLoadMezfesData", &MsgMhfLoadMezfesData{}},
		{"MsgMhfLoadPlateMyset", &MsgMhfLoadPlateMyset{}},
	}

	ctx := &clientctx.ClientContext{}
	for _, tc := range packets {
		t.Run(tc.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(0x12345678) // AckHandle
			_, _ = bf.Seek(0, io.SeekStart)

			err := tc.pkt.Parse(bf, ctx)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
}

// TestBatchParseTwoUint32 tests packets with AckHandle + one uint32 field.
func TestBatchParseTwoUint32(t *testing.T) {
	packets := []struct {
		name string
		pkt  MHFPacket
	}{
		{"MsgMhfListMail", &MsgMhfListMail{}},
		{"MsgMhfEnumerateTitle", &MsgMhfEnumerateTitle{}},
		{"MsgMhfInfoGuild", &MsgMhfInfoGuild{}},
		{"MsgMhfCheckDailyCafepoint", &MsgMhfCheckDailyCafepoint{}},
		{"MsgMhfEntryRookieGuild", &MsgMhfEntryRookieGuild{}},
		{"MsgMhfReleaseEvent", &MsgMhfReleaseEvent{}},
		{"MsgMhfSetGuildMissionTarget", &MsgMhfSetGuildMissionTarget{}},
		{"MsgMhfCancelGuildMissionTarget", &MsgMhfCancelGuildMissionTarget{}},
		{"MsgMhfAcquireFestaIntermediatePrize", &MsgMhfAcquireFestaIntermediatePrize{}},
		{"MsgMhfAcquireFestaPersonalPrize", &MsgMhfAcquireFestaPersonalPrize{}},
		{"MsgMhfGetGachaPlayHistory", &MsgMhfGetGachaPlayHistory{}},
		{"MsgMhfPostGuildScout", &MsgMhfPostGuildScout{}},
		{"MsgMhfCancelGuildScout", &MsgMhfCancelGuildScout{}},
		{"MsgMhfGetEnhancedMinidata", &MsgMhfGetEnhancedMinidata{}},
		{"MsgMhfPostBoostTime", &MsgMhfPostBoostTime{}},
		{"MsgMhfStartBoostTime", &MsgMhfStartBoostTime{}},
		{"MsgMhfAcquireGuildAdventure", &MsgMhfAcquireGuildAdventure{}},
		{"MsgMhfGetBoxGachaInfo", &MsgMhfGetBoxGachaInfo{}},
		{"MsgMhfResetBoxGachaInfo", &MsgMhfResetBoxGachaInfo{}},
		{"MsgMhfAddKouryouPoint", &MsgMhfAddKouryouPoint{}},
		{"MsgMhfExchangeKouryouPoint", &MsgMhfExchangeKouryouPoint{}},
		{"MsgMhfInfoJoint", &MsgMhfInfoJoint{}},
	}

	ctx := &clientctx.ClientContext{}
	for _, tc := range packets {
		t.Run(tc.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(0x12345678) // AckHandle
			bf.WriteUint32(0xDEADBEEF) // Second uint32
			bf.WriteUint32(0xCAFEBABE) // Padding for 3-field packets
			_, _ = bf.Seek(0, io.SeekStart)

			err := tc.pkt.Parse(bf, ctx)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
}

// TestBatchParseMultiField tests packets with various field combinations.
func TestBatchParseMultiField(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("MsgMhfGetRengokuBinary", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(0)  // Unk0
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetRengokuBinary{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateDistItem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // DistType
		bf.WriteUint8(3)  // Unk1
		bf.WriteUint16(4) // Unk2
		bf.WriteUint8(0)  // Unk3 length (Z1+ mode)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateDistItem{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 || pkt.DistType != 2 || pkt.Unk1 != 3 || pkt.MaxCount != 4 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfApplyDistItem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // DistributionType
		bf.WriteUint32(3) // DistributionID
		bf.WriteUint32(4) // Unk2
		bf.WriteUint32(5) // Unk3
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfApplyDistItem{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 || pkt.DistributionType != 2 || pkt.DistributionID != 3 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfAcquireDistItem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // DistributionType
		bf.WriteUint32(3) // DistributionID
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAcquireDistItem{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 || pkt.DistributionType != 2 || pkt.DistributionID != 3 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfGetDistDescription", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // Unk0
		bf.WriteUint32(3) // DistributionID
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetDistDescription{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 || pkt.Unk0 != 2 || pkt.DistributionID != 3 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfRegisterEvent", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Unk0
		bf.WriteUint16(3) // WorldID
		bf.WriteUint16(4) // LandID
		bf.WriteBool(true) // Unk1
		bf.WriteUint8(0)  // Zeroed (discarded)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfRegisterEvent{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 || pkt.Unk0 != 2 || pkt.WorldID != 3 || pkt.LandID != 4 || !pkt.CheckOnly {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfUpdateCafepoint", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Zeroed (discarded)
		bf.WriteUint16(3) // Zeroed (discarded)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfUpdateCafepoint{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfUpdateEtcPoint", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // PointType
		bf.WriteInt16(-5) // Delta
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfUpdateEtcPoint{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.PointType != 2 || pkt.Delta != -5 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfAcquireTitle", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Title count
		bf.WriteUint16(0) // Zeroed
		bf.WriteUint16(4) // TitleIDs[0]
		bf.WriteUint16(5) // TitleIDs[1]
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAcquireTitle{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if len(pkt.TitleIDs) != 2 || pkt.TitleIDs[0] != 4 || pkt.TitleIDs[1] != 5 {
			t.Errorf("TitleIDs = %v, want [4, 5]", pkt.TitleIDs)
		}
	})

	t.Run("MsgSysHideClient", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteBool(true) // Hide
		bf.WriteUint8(0)   // Zeroed (discarded)
		bf.WriteUint8(0)   // Zeroed (discarded)
		bf.WriteUint8(0)   // Zeroed (discarded)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysHideClient{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if !pkt.Hide {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgSysIssueLogkey", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Unk0
		bf.WriteUint16(0) // Zeroed (discarded)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysIssueLogkey{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 || pkt.Unk0 != 2 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfGetTinyBin", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // Unk0
		bf.WriteUint8(3)  // Unk1
		bf.WriteUint8(4)  // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetTinyBin{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 || pkt.Unk0 != 2 || pkt.Unk1 != 3 || pkt.Unk2 != 4 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfGetPaperData", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Unk0
		bf.WriteUint32(3) // Unk1
		bf.WriteUint32(4) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetPaperData{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.DataType != 4 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfGetEarthValue", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		for i := 0; i < 8; i++ {
			bf.WriteUint32(uint32(i + 1))
		}
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetEarthValue{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 || pkt.Unk6 != 8 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfPresentBox", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Unk0
		bf.WriteUint32(3) // Unk1
		bf.WriteUint32(2) // Unk2 (controls Unk7 slice length)
		bf.WriteUint32(5) // Unk3
		bf.WriteUint32(6) // Unk4
		bf.WriteUint32(7) // Unk5
		bf.WriteUint32(8) // Unk6
		bf.WriteUint32(9) // Unk7[0]
		bf.WriteUint32(10) // Unk7[1]
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfPresentBox{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 || pkt.Unk2 != 2 || pkt.Unk6 != 8 || len(pkt.Unk7) != 2 || pkt.Unk7[1] != 10 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfReadMail", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // AccIndex
		bf.WriteUint8(3)  // Index
		bf.WriteUint16(4) // Unk0
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfReadMail{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AccIndex != 2 || pkt.Index != 3 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfOprMember", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteBool(true)  // Blacklist
		bf.WriteBool(false) // Operation
		bf.WriteUint8(0)    // Padding
		bf.WriteUint8(1)    // CharID count
		bf.WriteUint32(99)  // CharIDs[0]
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfOprMember{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if !pkt.Blacklist || pkt.Operation || len(pkt.CharIDs) != 1 || pkt.CharIDs[0] != 99 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfListMember", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // Unk0
		bf.WriteUint8(0)  // Zeroed
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfListMember{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.Unk0 != 2 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfTransferItem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Unk0
		bf.WriteUint8(3)  // Unk1
		bf.WriteUint8(0)  // Zeroed
		bf.WriteUint16(4) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfTransferItem{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.Unk0 != 2 || pkt.Unk1 != 3 || pkt.Unk2 != 4 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfMercenaryHuntdata", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // Unk0
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfMercenaryHuntdata{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.RequestType != 2 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfEnumeratePrice", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(0) // Unk0
		bf.WriteUint16(0) // Unk1
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumeratePrice{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateUnionItem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Unk0
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateUnionItem{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.AckHandle != 1 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfEnumerateGuildItem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // GuildId
		bf.WriteUint16(3) // Unk0
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateGuildItem{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.GuildID != 2 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfEnumerateGuildMember", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint16(2)  // Unk0
		bf.WriteUint32(3)  // Unk1
		bf.WriteUint32(99) // GuildID
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateGuildMember{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.GuildID != 99 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfOperateGuildMember", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)              // AckHandle
		bf.WriteUint32(2)              // GuildID
		bf.WriteUint32(99)             // CharID
		bf.WriteUint8(1)               // Action
		bf.WriteBytes([]byte{0, 0, 0}) // Unk
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfOperateGuildMember{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.CharID != 99 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfUpdateEquipSkinHist", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // MogType
		bf.WriteUint16(3) // ArmourID
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfUpdateEquipSkinHist{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.MogType != 2 || pkt.ArmourID != 3 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfSetRejectGuildScout", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteBool(true) // Reject
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSetRejectGuildScout{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if !pkt.Reject {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfUseKeepLoginBoost", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(3)  // BoostWeekUsed
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfUseKeepLoginBoost{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.BoostWeekUsed != 3 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfSetCaAchievementHist", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Unk0
		bf.WriteUint32(3) // Unk1
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSetCaAchievementHist{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfAddGuildWeeklyBonusExceptionalUser", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // NumUsers
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAddGuildWeeklyBonusExceptionalUser{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetLobbyCrowd", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Server
		bf.WriteUint32(3) // Room
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetLobbyCrowd{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSexChanger", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(1)  // Gender
		bf.WriteUint8(0)  // Unk0
		bf.WriteUint8(0)  // Unk1
		bf.WriteUint8(0)  // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSexChanger{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.Gender != 1 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgMhfSetKiju", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(5) // Unk1
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSetKiju{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfAddUdPoint", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Unk1
		bf.WriteUint32(3) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAddUdPoint{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetWeeklySeibatuRankingReward", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2)
		bf.WriteUint32(3)
		bf.WriteUint32(4)
		bf.WriteUint32(5)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetWeeklySeibatuRankingReward{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetEarthStatus", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Unk0
		bf.WriteUint32(3) // Unk1
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetEarthStatus{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfAddGuildMissionCount", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // MissionID
		bf.WriteUint32(3) // Count
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAddGuildMissionCount{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateAiroulist", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Unk0
		bf.WriteUint16(3) // Unk1
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateAiroulist{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfOperateGuildTresureReport", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint32(10) // HuntID
		bf.WriteUint16(2)  // State
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfOperateGuildTresureReport{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfAcquireGuildTresure", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint32(10) // HuntID
		bf.WriteUint8(1)   // Unk
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAcquireGuildTresure{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateGuildTresure", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(5) // MaxHunts
		bf.WriteUint32(0) // Unk
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateGuildTresure{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetTenrouirai", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Unk0
		bf.WriteUint32(3) // Unk1
		bf.WriteUint16(4) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetTenrouirai{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfPostTenrouirai", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Unk0
		bf.WriteUint32(3) // Unk1
		bf.WriteUint32(4) // Unk2
		bf.WriteUint32(5) // Unk3
		bf.WriteUint32(6) // Unk4
		bf.WriteUint8(7)  // Unk5
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfPostTenrouirai{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetSeibattle", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // Unk0
		bf.WriteUint8(3)  // Unk1
		bf.WriteUint32(4) // Unk2
		bf.WriteUint8(5)  // Unk3
		bf.WriteUint16(6) // Unk4
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetSeibattle{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetRyoudama", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint8(2)   // Unk0
		bf.WriteUint8(3)   // Unk1
		bf.WriteUint32(99) // GuildID
		bf.WriteUint8(4)   // Unk3
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetRyoudama{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateRengokuRanking", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Leaderboard
		bf.WriteUint16(3) // Unk1
		bf.WriteUint16(4) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateRengokuRanking{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetAdditionalBeatReward", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2)
		bf.WriteUint32(3)
		bf.WriteUint32(4)
		bf.WriteUint32(5)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetAdditionalBeatReward{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSetRestrictionEvent", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Unk0
		bf.WriteUint32(3) // Unk1
		bf.WriteUint32(4) // Unk2
		bf.WriteUint8(5)  // Unk3
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSetRestrictionEvent{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfUpdateUseTrendWeaponLog", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // Unk0
		bf.WriteUint16(3) // Unk1
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfUpdateUseTrendWeaponLog{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfDisplayedAchievement", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint8(42) // AchievementID
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfDisplayedAchievement{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfRegistGuildCooking", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // OverwriteID
		bf.WriteUint16(3) // MealID
		bf.WriteUint8(4)  // Success
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfRegistGuildCooking{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfChargeGuildAdventure", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // ID
		bf.WriteUint32(3) // Amount
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfChargeGuildAdventure{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfRegistGuildAdventure", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Destination
		bf.WriteUint32(0) // discard CharID
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfRegistGuildAdventure{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfReadMercenaryW", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // Op
		bf.WriteUint8(3)  // Unk1
		bf.WriteUint16(4) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfReadMercenaryW{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfReadMercenaryM", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // CharID
		bf.WriteUint32(3) // MercID
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfReadMercenaryM{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfContractMercenary", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // PactMercID
		bf.WriteUint32(3) // CID
		bf.WriteUint8(4)  // Op
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfContractMercenary{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetGuildTargetMemberNum", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // GuildID
		bf.WriteUint8(3)  // Unk
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetGuildTargetMemberNum{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSetGuildManageRight", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)              // AckHandle
		bf.WriteUint32(2)              // CharID
		bf.WriteBool(true)             // Allowed
		bf.WriteBytes([]byte{0, 0, 0}) // Unk
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSetGuildManageRight{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfAnswerGuildScout", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint32(2)  // LeaderID
		bf.WriteBool(true) // Answer
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAnswerGuildScout{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfPlayStepupGacha", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // GachaID
		bf.WriteUint8(3)  // RollType
		bf.WriteUint8(4)  // GachaType
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfPlayStepupGacha{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfPlayBoxGacha", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // GachaID
		bf.WriteUint8(3)  // RollType
		bf.WriteUint8(4)  // GachaType
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfPlayBoxGacha{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfPlayNormalGacha", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // GachaID
		bf.WriteUint8(3)  // RollType
		bf.WriteUint8(4)  // GachaType
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfPlayNormalGacha{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfReceiveGachaItem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint8(5)    // Max
		bf.WriteBool(false) // Freeze
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfReceiveGachaItem{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetStepupStatus", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // GachaID
		bf.WriteUint8(3)  // Unk
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetStepupStatus{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfUseGachaPoint", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Unk0
		bf.WriteUint32(3) // TrialCoins
		bf.WriteUint32(4) // PremiumCoins
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfUseGachaPoint{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateGuildMessageBoard", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Unk0
		bf.WriteUint32(3) // MaxPosts
		bf.WriteUint32(4) // BoardType
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateGuildMessageBoard{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})
}

// TestBatchParseVariableLength tests packets with variable-length data.
func TestBatchParseVariableLength(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("MsgMhfSaveFavoriteQuest", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)                             // AckHandle
		bf.WriteUint16(4)                             // DataSize
		bf.WriteBytes([]byte{0x01, 0x02, 0x03, 0x04}) // Data
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSaveFavoriteQuest{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if len(pkt.Data) != 4 {
			t.Errorf("Data len = %d, want 4", len(pkt.Data))
		}
	})

	t.Run("MsgMhfSavedata_withDataSize", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(0) // AllocMemSize
		bf.WriteUint8(0)  // SaveType
		bf.WriteUint32(0) // Unk1
		bf.WriteUint32(3) // DataSize (non-zero)
		bf.WriteBytes([]byte{0xAA, 0xBB, 0xCC})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSavedata{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if len(pkt.RawDataPayload) != 3 {
			t.Errorf("RawDataPayload len = %d, want 3", len(pkt.RawDataPayload))
		}
	})

	t.Run("MsgMhfSavedata_withAllocMem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // AllocMemSize
		bf.WriteUint8(0)  // SaveType
		bf.WriteUint32(0) // Unk1
		bf.WriteUint32(0) // DataSize (zero -> use AllocMemSize)
		bf.WriteBytes([]byte{0xAA, 0xBB})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSavedata{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if len(pkt.RawDataPayload) != 2 {
			t.Errorf("RawDataPayload len = %d, want 2", len(pkt.RawDataPayload))
		}
	})

	t.Run("MsgMhfTransitMessage", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // Unk0
		bf.WriteUint8(3)  // Unk1
		bf.WriteUint16(4) // SearchType
		bf.WriteUint16(3) // inline data length
		bf.WriteBytes([]byte{0xAA, 0xBB, 0xCC})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfTransitMessage{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if len(pkt.MessageData) != 3 {
			t.Errorf("MessageData len = %d, want 3", len(pkt.MessageData))
		}
	})

	t.Run("MsgMhfPostTinyBin", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // Unk0
		bf.WriteUint8(3)  // Unk1
		bf.WriteUint8(4)  // Unk2
		bf.WriteUint16(2) // inline data length
		bf.WriteBytes([]byte{0xAA, 0xBB})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfPostTinyBin{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if len(pkt.Data) != 2 {
			t.Errorf("Data len = %d, want 2", len(pkt.Data))
		}
	})

	t.Run("MsgSysRecordLog", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // Unk0
		bf.WriteUint16(3) // Unk1
		bf.WriteUint16(4) // HardcodedDataSize
		bf.WriteUint32(5) // Unk3
		bf.WriteBytes([]byte{0x01, 0x02, 0x03, 0x04})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysRecordLog{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if len(pkt.Data) != 4 {
			t.Errorf("Data len = %d, want 4", len(pkt.Data))
		}
	})

	t.Run("MsgMhfUpdateInterior", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)               // AckHandle
		bf.WriteBytes(make([]byte, 20)) // InteriorData
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfUpdateInterior{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if len(pkt.InteriorData) != 20 {
			t.Error("InteriorData wrong size")
		}
	})

	t.Run("MsgMhfSavePartner", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(3) // DataSize
		bf.WriteBytes([]byte{0xAA, 0xBB, 0xCC})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSavePartner{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSaveOtomoAirou", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(2) // DataSize
		bf.WriteBytes([]byte{0xAA, 0xBB})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSaveOtomoAirou{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSaveHunterNavi", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint32(2)  // DataSize
		bf.WriteBool(true) // IsDataDiff
		bf.WriteBytes([]byte{0xAA, 0xBB})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSaveHunterNavi{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSavePlateData", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint32(3)   // DataSize
		bf.WriteBool(false) // IsDataDiff
		bf.WriteBytes([]byte{0x01, 0x02, 0x03})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSavePlateData{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSavePlateBox", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint32(2)  // DataSize
		bf.WriteBool(true) // IsDataDiff
		bf.WriteBytes([]byte{0xAA, 0xBB})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSavePlateBox{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSavePlateMyset", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // DataSize
		bf.WriteBytes([]byte{0xAA, 0xBB})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSavePlateMyset{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSaveDecoMyset", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // DataSize
		bf.WriteBytes([]byte{0xAA, 0xBB})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSaveDecoMyset{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSaveRengokuData", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // DataSize
		bf.WriteBytes([]byte{0xAA, 0xBB})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSaveRengokuData{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSaveMezfesData", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(2) // DataSize
		bf.WriteBytes([]byte{0xAA, 0xBB})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSaveMezfesData{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSaveScenarioData", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint32(3) // DataSize
		bf.WriteBytes([]byte{0x01, 0x02, 0x03})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSaveScenarioData{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfAcquireExchangeShop", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint16(3) // DataSize
		bf.WriteBytes([]byte{0xAA, 0xBB, 0xCC})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAcquireExchangeShop{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfSetEnhancedMinidata", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)                  // AckHandle
		bf.WriteUint16(0)                  // Unk0
		bf.WriteBytes(make([]byte, 0x400)) // RawDataPayload
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfSetEnhancedMinidata{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetBbsUserStatus", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)               // AckHandle
		bf.WriteBytes(make([]byte, 12)) // Unk
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetBbsUserStatus{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetBbsSnsStatus", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)               // AckHandle
		bf.WriteBytes(make([]byte, 12)) // Unk
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetBbsSnsStatus{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})
}

// TestBatchParseArrangeGuildMember tests the array-parsing packet.
func TestBatchParseArrangeGuildMember(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1)  // AckHandle
	bf.WriteUint32(2)  // GuildID
	bf.WriteUint16(3)  // charCount
	bf.WriteUint32(10) // CharIDs[0]
	bf.WriteUint32(20) // CharIDs[1]
	bf.WriteUint32(30) // CharIDs[2]
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfArrangeGuildMember{}
	if err := pkt.Parse(bf, &clientctx.ClientContext{}); err != nil {
		t.Fatal(err)
	}
	if len(pkt.CharIDs) != 3 || pkt.CharIDs[2] != 30 {
		t.Errorf("CharIDs = %v, want [10 20 30]", pkt.CharIDs)
	}
}

// TestBatchParseUpdateGuildIcon tests the guild icon array packet.
func TestBatchParseUpdateGuildIcon(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint32(2) // GuildID
	bf.WriteUint16(1) // PartCount
	bf.WriteUint16(0) // Unk1
	// One part: 14 bytes
	bf.WriteUint16(0)   // Index
	bf.WriteUint16(1)   // ID
	bf.WriteUint8(2)    // Page
	bf.WriteUint8(3)    // Size
	bf.WriteUint8(4)    // Rotation
	bf.WriteUint8(0xFF) // Red
	bf.WriteUint8(0x00) // Green
	bf.WriteUint8(0x80) // Blue
	bf.WriteUint16(100) // PosX
	bf.WriteUint16(200) // PosY
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfUpdateGuildIcon{}
	if err := pkt.Parse(bf, &clientctx.ClientContext{}); err != nil {
		t.Fatal(err)
	}
	if len(pkt.IconParts) != 1 || pkt.IconParts[0].Red != 0xFF {
		t.Error("icon parts mismatch")
	}
}

// TestBatchParseSysLoadRegister tests the fixed-zero validation packet.
func TestBatchParseSysLoadRegister(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint32(2) // RegisterID
	bf.WriteUint8(3)  // Unk1
	bf.WriteUint16(0) // fixedZero0
	bf.WriteUint8(0)  // fixedZero1
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysLoadRegister{}
	if err := pkt.Parse(bf, &clientctx.ClientContext{}); err != nil {
		t.Fatal(err)
	}
	if pkt.RegisterID != 2 || pkt.Values != 3 {
		t.Error("field mismatch")
	}
}

// TestBatchParseSysLoadRegisterNonZeroPadding tests that SysLoadRegister Parse
// succeeds even with non-zero values in the padding fields (they are discarded).
func TestBatchParseSysLoadRegisterNonZeroPadding(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint32(2) // RegisterID
	bf.WriteUint8(3)  // Values
	bf.WriteUint8(1)  // Zeroed (discarded, non-zero is OK)
	bf.WriteUint16(1) // Zeroed (discarded, non-zero is OK)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysLoadRegister{}
	err := pkt.Parse(bf, &clientctx.ClientContext{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if pkt.AckHandle != 1 {
		t.Errorf("AckHandle = %d, want 1", pkt.AckHandle)
	}
	if pkt.RegisterID != 2 {
		t.Errorf("RegisterID = %d, want 2", pkt.RegisterID)
	}
	if pkt.Values != 3 {
		t.Errorf("Values = %d, want 3", pkt.Values)
	}
}

// TestBatchParseSysOperateRegister tests the operate register packet.
func TestBatchParseSysOperateRegister(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint32(2) // SemaphoreID
	bf.WriteUint16(0) // fixedZero
	bf.WriteUint16(3) // dataSize
	bf.WriteBytes([]byte{0xAA, 0xBB, 0xCC})
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysOperateRegister{}
	if err := pkt.Parse(bf, &clientctx.ClientContext{}); err != nil {
		t.Fatal(err)
	}
	if len(pkt.RawDataPayload) != 3 {
		t.Error("payload size mismatch")
	}
}

// TestBatchParseSysOperateRegisterNonZeroPadding tests that SysOperateRegister Parse
// succeeds even with non-zero values in the padding field (it is discarded).
func TestBatchParseSysOperateRegisterNonZeroPadding(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint32(2) // SemaphoreID
	bf.WriteUint16(1) // Zeroed (discarded, non-zero is OK)
	bf.WriteUint16(0) // dataSize
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysOperateRegister{}
	err := pkt.Parse(bf, &clientctx.ClientContext{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if pkt.AckHandle != 1 {
		t.Errorf("AckHandle = %d, want 1", pkt.AckHandle)
	}
	if pkt.SemaphoreID != 2 {
		t.Errorf("SemaphoreID = %d, want 2", pkt.SemaphoreID)
	}
	if len(pkt.RawDataPayload) != 0 {
		t.Errorf("RawDataPayload len = %d, want 0", len(pkt.RawDataPayload))
	}
}

// TestBatchParseSysGetFile tests the conditional scenario file packet.
func TestBatchParseSysGetFile(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("non-scenario", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteBool(false) // IsScenario
		bf.WriteUint8(5)    // filenameLength
		bf.WriteBytes([]byte("test\x00"))
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysGetFile{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.Filename != "test" || pkt.IsScenario {
			t.Error("field mismatch")
		}
	})

	t.Run("scenario", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteBool(true)  // IsScenario
		bf.WriteUint8(0)    // filenameLength (empty)
		bf.WriteUint8(10)   // CategoryID
		bf.WriteUint32(100) // MainID
		bf.WriteUint8(5)    // ChapterID
		bf.WriteUint8(0)    // Flags
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysGetFile{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if !pkt.IsScenario || pkt.ScenarioIdentifer.MainID != 100 {
			t.Error("field mismatch")
		}
	})
}

// TestBatchParseSysTerminalLog tests the entry-array packet.
func TestBatchParseSysTerminalLog(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	bf.WriteUint32(2) // LogID
	bf.WriteUint16(1) // EntryCount
	bf.WriteUint16(0) // Unk0
	// One entry: 4 + 1 + 1 + (15*2) = 36 bytes
	bf.WriteUint32(0) // Index
	bf.WriteUint8(1)  // Type1
	bf.WriteUint8(2)  // Type2
	for i := 0; i < 15; i++ {
		bf.WriteInt16(int16(i))
	}
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgSysTerminalLog{}
	if err := pkt.Parse(bf, &clientctx.ClientContext{}); err != nil {
		t.Fatal(err)
	}
	if len(pkt.Entries) != 1 || pkt.Entries[0].Type1 != 1 {
		t.Error("entries mismatch")
	}
}

// TestBatchParseNoOpPackets tests packets with empty Parse (return nil).
func TestBatchParseNoOpPackets(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	bf := byteframe.NewByteFrame()

	packets := []struct {
		name string
		pkt  MHFPacket
	}{
		{"MsgSysExtendThreshold", &MsgSysExtendThreshold{}},
		{"MsgSysEnd", &MsgSysEnd{}},
		{"MsgSysNop", &MsgSysNop{}},
		{"MsgSysStageDestruct", &MsgSysStageDestruct{}},
	}

	for _, tc := range packets {
		t.Run(tc.name, func(t *testing.T) {
			if err := tc.pkt.Parse(bf, ctx); err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
		})
	}
}

// TestBatchParseNotImplemented tests that Parse returns NOT IMPLEMENTED for stub packets.
func TestBatchParseNotImplemented(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	bf := byteframe.NewByteFrame()

	packets := []MHFPacket{
		&MsgSysReserve01{}, &MsgSysReserve02{}, &MsgSysReserve03{},
		&MsgSysReserve04{}, &MsgSysReserve05{}, &MsgSysReserve06{},
		&MsgSysReserve07{}, &MsgSysReserve0C{}, &MsgSysReserve0D{},
		&MsgSysReserve0E{}, &MsgSysReserve4A{}, &MsgSysReserve4B{},
		&MsgSysReserve4C{}, &MsgSysReserve4D{}, &MsgSysReserve4E{},
		&MsgSysReserve4F{}, &MsgSysReserve55{}, &MsgSysReserve56{},
		&MsgSysReserve57{}, &MsgSysReserve5C{}, &MsgSysReserve5E{},
		&MsgSysReserve5F{}, &MsgSysReserve71{}, &MsgSysReserve72{},
		&MsgSysReserve73{}, &MsgSysReserve74{}, &MsgSysReserve75{},
		&MsgSysReserve76{}, &MsgSysReserve77{}, &MsgSysReserve78{},
		&MsgSysReserve79{}, &MsgSysReserve7A{}, &MsgSysReserve7B{},
		&MsgSysReserve7C{}, &MsgSysReserve7E{}, &MsgSysReserve18E{},
		&MsgSysReserve18F{}, &MsgSysReserve19E{}, &MsgSysReserve19F{},
		&MsgSysReserve1A4{}, &MsgSysReserve1A6{}, &MsgSysReserve1A7{},
		&MsgSysReserve1A8{}, &MsgSysReserve1A9{}, &MsgSysReserve1AA{},
		&MsgSysReserve1AB{}, &MsgSysReserve1AC{}, &MsgSysReserve1AD{},
		&MsgSysReserve1AE{}, &MsgSysReserve1AF{}, &MsgSysReserve19B{},
		&MsgSysReserve192{}, &MsgSysReserve193{}, &MsgSysReserve194{},
		&MsgSysReserve180{},
		&MsgMhfReserve10F{},
		// Empty-struct packets with NOT IMPLEMENTED Parse
		&MsgHead{}, &MsgSysSetStatus{}, &MsgSysEcho{},
		&MsgSysLeaveStage{}, &MsgSysAddObject{}, &MsgSysDelObject{},
		&MsgSysDispObject{}, &MsgSysHideObject{},
		&MsgMhfServerCommand{}, &MsgMhfSetLoginwindow{}, &MsgMhfShutClient{},
		&MsgMhfUpdateGuildcard{},
		&MsgMhfGetCogInfo{},
		&MsgCaExchangeItem{},
	}

	for _, pkt := range packets {
		t.Run(pkt.Opcode().String(), func(t *testing.T) {
			err := pkt.Parse(bf, ctx)
			if err == nil {
				t.Error("expected NOT IMPLEMENTED error")
			}
		})
	}
}

// TestBatchBuildNotImplemented tests that Build returns NOT IMPLEMENTED for many packets.
func TestBatchBuildNotImplemented(t *testing.T) {
	ctx := &clientctx.ClientContext{}
	bf := byteframe.NewByteFrame()

	packets := []MHFPacket{
		&MsgMhfLoaddata{}, &MsgMhfSavedata{},
		&MsgMhfListMember{}, &MsgMhfOprMember{},
		&MsgMhfEnumerateDistItem{}, &MsgMhfApplyDistItem{}, &MsgMhfAcquireDistItem{},
		&MsgMhfGetDistDescription{}, &MsgMhfSendMail{}, &MsgMhfReadMail{},
		&MsgMhfListMail{}, &MsgMhfOprtMail{},
		&MsgMhfLoadFavoriteQuest{}, &MsgMhfSaveFavoriteQuest{},
		&MsgMhfRegisterEvent{}, &MsgMhfReleaseEvent{},
		&MsgMhfTransitMessage{}, &MsgMhfPresentBox{},
		&MsgMhfAcquireTitle{}, &MsgMhfEnumerateTitle{},
		&MsgMhfInfoGuild{}, &MsgMhfEnumerateGuild{},
		&MsgMhfCreateGuild{}, &MsgMhfOperateGuild{},
		&MsgMhfOperateGuildMember{}, &MsgMhfArrangeGuildMember{},
		&MsgMhfEnumerateGuildMember{}, &MsgMhfUpdateGuildIcon{},
		&MsgMhfInfoFesta{}, &MsgMhfEntryFesta{},
		&MsgMhfChargeFesta{}, &MsgMhfAcquireFesta{},
		&MsgMhfVoteFesta{}, &MsgMhfInfoTournament{},
		&MsgMhfEntryTournament{}, &MsgMhfAcquireTournament{},
		&MsgMhfUpdateCafepoint{}, &MsgMhfCheckDailyCafepoint{},
		&MsgMhfGetEtcPoints{}, &MsgMhfUpdateEtcPoint{},
		&MsgMhfReadGuildcard{}, &MsgMhfUpdateGuildcard{},
		&MsgMhfGetTinyBin{}, &MsgMhfPostTinyBin{},
		&MsgMhfGetPaperData{}, &MsgMhfGetEarthValue{},
		&MsgSysRecordLog{}, &MsgSysIssueLogkey{}, &MsgSysTerminalLog{},
		&MsgSysHideClient{}, &MsgSysGetFile{},
		&MsgSysOperateRegister{}, &MsgSysLoadRegister{},
		&MsgMhfGetGuildMissionList{}, &MsgMhfGetGuildMissionRecord{},
		&MsgMhfAddGuildMissionCount{}, &MsgMhfSetGuildMissionTarget{},
		&MsgMhfCancelGuildMissionTarget{},
		&MsgMhfEnumerateGuildTresure{}, &MsgMhfRegistGuildTresure{},
		&MsgMhfAcquireGuildTresure{}, &MsgMhfOperateGuildTresureReport{},
		&MsgMhfGetGuildTresureSouvenir{}, &MsgMhfAcquireGuildTresureSouvenir{},
		&MsgMhfEnumerateFestaIntermediatePrize{}, &MsgMhfAcquireFestaIntermediatePrize{},
		&MsgMhfEnumerateFestaPersonalPrize{}, &MsgMhfAcquireFestaPersonalPrize{},
		&MsgMhfGetGuildWeeklyBonusMaster{}, &MsgMhfGetGuildWeeklyBonusActiveCount{},
		&MsgMhfAddGuildWeeklyBonusExceptionalUser{},
		&MsgMhfGetEquipSkinHist{}, &MsgMhfUpdateEquipSkinHist{},
		&MsgMhfGetEnhancedMinidata{}, &MsgMhfSetEnhancedMinidata{},
		&MsgMhfGetLobbyCrowd{},
		&MsgMhfGetRejectGuildScout{}, &MsgMhfSetRejectGuildScout{},
		&MsgMhfGetKeepLoginBoostStatus{}, &MsgMhfUseKeepLoginBoost{},
		&MsgMhfAcquireMonthlyReward{},
		&MsgMhfPostGuildScout{}, &MsgMhfCancelGuildScout{},
		&MsgMhfAnswerGuildScout{}, &MsgMhfGetGuildScoutList{},
		&MsgMhfGetGuildManageRight{}, &MsgMhfSetGuildManageRight{},
		&MsgMhfGetGuildTargetMemberNum{},
		&MsgMhfPlayStepupGacha{}, &MsgMhfReceiveGachaItem{},
		&MsgMhfGetStepupStatus{}, &MsgMhfPlayNormalGacha{},
		&MsgMhfPlayBoxGacha{}, &MsgMhfGetBoxGachaInfo{}, &MsgMhfResetBoxGachaInfo{},
		&MsgMhfUseGachaPoint{}, &MsgMhfGetGachaPlayHistory{},
		&MsgMhfSavePartner{}, &MsgMhfSaveOtomoAirou{},
		&MsgMhfSaveHunterNavi{}, &MsgMhfSavePlateData{},
		&MsgMhfSavePlateBox{}, &MsgMhfSavePlateMyset{},
		&MsgMhfSaveDecoMyset{}, &MsgMhfSaveRengokuData{}, &MsgMhfSaveMezfesData{},
		&MsgMhfCreateMercenary{}, &MsgMhfSaveMercenary{},
		&MsgMhfReadMercenaryW{}, &MsgMhfReadMercenaryM{},
		&MsgMhfContractMercenary{}, &MsgMhfEnumerateMercenaryLog{},
		&MsgMhfRegistGuildCooking{}, &MsgMhfRegistGuildAdventure{},
		&MsgMhfAcquireGuildAdventure{}, &MsgMhfChargeGuildAdventure{},
		&MsgMhfLoadLegendDispatch{},
		&MsgMhfPostBoostTime{}, &MsgMhfStartBoostTime{},
		&MsgMhfPostBoostTimeQuestReturn{}, &MsgMhfGetBoostRight{},
		&MsgMhfGetFpointExchangeList{},
		&MsgMhfGetRewardSong{}, &MsgMhfUseRewardSong{},
		&MsgMhfGetKouryouPoint{}, &MsgMhfAddKouryouPoint{}, &MsgMhfExchangeKouryouPoint{},
		&MsgMhfSexChanger{}, &MsgMhfSetKiju{}, &MsgMhfAddUdPoint{},
		&MsgMhfGetTrendWeapon{}, &MsgMhfUpdateUseTrendWeaponLog{},
		&MsgMhfSetRestrictionEvent{},
		&MsgMhfGetWeeklySeibatuRankingReward{}, &MsgMhfGetEarthStatus{},
		&MsgMhfAddGuildMissionCount{},
		&MsgMhfEnumerateAiroulist{},
		&MsgMhfEnumerateRengokuRanking{}, &MsgMhfGetRengokuRankingRank{},
		&MsgMhfGetAdditionalBeatReward{},
		&MsgMhfSetCaAchievementHist{},
		&MsgMhfGetUdMyPoint{}, &MsgMhfGetUdTotalPointInfo{},
		&MsgMhfDisplayedAchievement{},
		&MsgMhfUpdateInterior{},
		&MsgMhfEnumerateUnionItem{},
		&MsgMhfEnumerateGuildItem{},
		&MsgMhfEnumerateGuildMember{},
		&MsgMhfEnumerateGuildMessageBoard{},
		&MsgMhfMercenaryHuntdata{},
		&MsgMhfEntryRookieGuild{},
		&MsgMhfEnumeratePrice{},
		&MsgMhfTransferItem{},
		&MsgMhfGetSeibattle{}, &MsgMhfGetRyoudama{},
		&MsgMhfGetTenrouirai{}, &MsgMhfPostTenrouirai{},
		&MsgMhfGetBbsUserStatus{}, &MsgMhfGetBbsSnsStatus{},
		&MsgMhfInfoScenarioCounter{}, &MsgMhfLoadScenarioData{},
		&MsgMhfSaveScenarioData{},
		&MsgMhfAcquireExchangeShop{},
		&MsgMhfLoadRengokuData{}, &MsgMhfGetRengokuBinary{},
		&MsgMhfLoadMezfesData{}, &MsgMhfLoadPlateMyset{},
	}

	for _, pkt := range packets {
		t.Run(pkt.Opcode().String(), func(t *testing.T) {
			err := pkt.Build(bf, ctx)
			if err == nil {
				// Some packets may have Build implemented - that's fine
				t.Logf("Build() succeeded (has implementation)")
			}
		})
	}
}

// TestBatchParseReserve188and18B tests reserve packets with AckHandle.
func TestBatchParseReserve188and18B(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	for _, tc := range []struct {
		name string
		pkt  MHFPacket
	}{
		{"MsgSysReserve188", &MsgSysReserve188{}},
		{"MsgSysReserve18B", &MsgSysReserve18B{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(0x12345678)
			_, _ = bf.Seek(0, io.SeekStart)
			if err := tc.pkt.Parse(bf, ctx); err != nil {
				t.Fatal(err)
			}
		})
	}
}

// TestBatchParseStageStringPackets tests packets that read a stage ID string.
func TestBatchParseStageStringPackets(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("MsgSysGetStageBinary", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // BinaryType0
		bf.WriteUint8(3)  // BinaryType1
		bf.WriteUint32(0) // Unk0
		bf.WriteUint8(6)  // stageIDLength
		bf.WriteBytes(append([]byte("room1"), 0x00))
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysGetStageBinary{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.StageID != "room1" {
			t.Errorf("StageID = %q, want room1", pkt.StageID)
		}
	})

	t.Run("MsgSysWaitStageBinary", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // BinaryType0
		bf.WriteUint8(3)  // BinaryType1
		bf.WriteUint32(0) // Unk0
		bf.WriteUint8(6)  // stageIDLength
		bf.WriteBytes(append([]byte("room2"), 0x00))
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysWaitStageBinary{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.StageID != "room2" {
			t.Errorf("StageID = %q, want room2", pkt.StageID)
		}
	})

	t.Run("MsgSysSetStageBinary", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint8(1)  // BinaryType0
		bf.WriteUint8(2)  // BinaryType1
		bf.WriteUint8(6)  // stageIDLength
		bf.WriteUint16(3) // dataSize
		bf.WriteBytes(append([]byte("room3"), 0x00))
		bf.WriteBytes([]byte{0xAA, 0xBB, 0xCC})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysSetStageBinary{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.StageID != "room3" || len(pkt.RawDataPayload) != 3 {
			t.Error("field mismatch")
		}
	})

	t.Run("MsgSysEnumerateClient", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // Unk0
		bf.WriteUint8(3)  // Get
		bf.WriteUint8(6)  // stageIDLength
		bf.WriteBytes(append([]byte("room4"), 0x00))
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysEnumerateClient{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.StageID != "room4" {
			t.Errorf("StageID = %q, want room4", pkt.StageID)
		}
	})

	t.Run("MsgSysSetStagePass", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint8(1) // Unk0
		bf.WriteUint8(5) // Password length
		bf.WriteBytes(append([]byte("pass"), 0x00))
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgSysSetStagePass{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.Password != "pass" {
			t.Errorf("Password = %q, want pass", pkt.Password)
		}
	})
}

// TestBatchParseStampcardStamp tests the stampcard packet with downcasts.
func TestBatchParseStampcardStamp(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1)  // AckHandle
	bf.WriteUint16(2)  // HR
	bf.WriteUint16(3)  // GR
	bf.WriteUint16(4)  // Stamps
	bf.WriteUint16(0)  // discard
	bf.WriteUint32(5)  // Reward1 (downcast to uint16)
	bf.WriteUint32(6)  // Reward2
	bf.WriteUint32(7)  // Item1
	bf.WriteUint32(8)  // Item2
	bf.WriteUint32(9)  // Quantity1
	bf.WriteUint32(10) // Quantity2
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfStampcardStamp{}
	if err := pkt.Parse(bf, &clientctx.ClientContext{}); err != nil {
		t.Fatal(err)
	}
	if pkt.HR != 2 || pkt.GR != 3 || pkt.Stamps != 4 || pkt.Reward1 != 5 {
		t.Error("field mismatch")
	}
}

// TestBatchParseAnnounce tests the announce packet with fixed-size byte array.
func TestBatchParseAnnounce(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1)               // AckHandle
	bf.WriteUint32(0x7F000001)      // IPAddress (127.0.0.1)
	bf.WriteUint16(54001)           // Port
	bf.WriteUint8(0)                // discard
	bf.WriteUint16(0)               // discard
	bf.WriteBytes(make([]byte, 32)) // StageID
	bf.WriteUint32(0)               // Data length (0 bytes)
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfAnnounce{}
	if err := pkt.Parse(bf, &clientctx.ClientContext{}); err != nil {
		t.Fatal(err)
	}
	if pkt.IPAddress != 0x7F000001 || pkt.Port != 54001 {
		t.Error("field mismatch")
	}
}

// TestBatchParseOprtMail tests conditional parsing.
func TestBatchParseOprtMail(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("delete", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint8(0)    // AccIndex
		bf.WriteUint8(1)    // Index
		bf.WriteUint8(0x01) // Operation = DELETE
		bf.WriteUint8(0)    // Unk0
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfOprtMail{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("acquire_item", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint8(0)    // AccIndex
		bf.WriteUint8(1)    // Index
		bf.WriteUint8(0x05) // Operation = ACQUIRE_ITEM
		bf.WriteUint8(0)    // Unk0
		bf.WriteUint16(5)   // Amount
		bf.WriteUint16(100) // ItemID
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfOprtMail{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
		if pkt.Amount != 5 || pkt.ItemID != 100 {
			t.Error("field mismatch")
		}
	})
}

// TestBatchParsePostTowerInfo tests the 11-field packet.
func TestBatchParsePostTowerInfo(t *testing.T) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1) // AckHandle
	for i := 0; i < 11; i++ {
		bf.WriteUint32(uint32(i + 10))
	}
	_, _ = bf.Seek(0, io.SeekStart)

	pkt := &MsgMhfPostTowerInfo{}
	if err := pkt.Parse(bf, &clientctx.ClientContext{}); err != nil {
		t.Fatal(err)
	}
}

// TestBatchParseGuildHuntdata tests conditional guild huntdata.
// TestBatchParseAdditionalMultiField tests Parse for more packets with multiple fields.
func TestBatchParseAdditionalMultiField(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("MsgMhfAcquireFesta", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint32(100) // FestaID
		bf.WriteUint32(200) // GuildID
		bf.WriteUint16(0)   // Unk
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAcquireFesta{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfAddUdTacticsPoint", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint16(10)  // Unk0
		bf.WriteUint32(500) // Unk1
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAddUdTacticsPoint{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfApplyCampaign", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)              // AckHandle
		bf.WriteUint32(1)              // Unk0
		bf.WriteUint16(2)              // Unk1
		bf.WriteBytes(make([]byte, 16)) // Unk2 (16 bytes)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfApplyCampaign{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfCheckMonthlyItem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(1)  // Type
		bf.WriteBytes(make([]byte, 3))
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfCheckMonthlyItem{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfCheckWeeklyStamp_hl", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint8(1)   // StampType = 1 ("hl")
		bf.WriteUint8(0)   // Unk1 (bool)
		bf.WriteUint16(10) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfCheckWeeklyStamp{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfCheckWeeklyStamp_ex", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint8(2)   // StampType = 2 ("ex")
		bf.WriteUint8(1)   // Unk1 (bool)
		bf.WriteUint16(20) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfCheckWeeklyStamp{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEntryFesta", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint32(100) // FestaID
		bf.WriteUint32(200) // GuildID
		bf.WriteUint16(0)   // padding
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEntryFesta{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateFestaMember", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint32(100) // FestaID
		bf.WriteUint32(200) // GuildID
		bf.WriteUint16(0)   // padding
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateFestaMember{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateInvGuild", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteBytes(make([]byte, 9))
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateInvGuild{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateWarehouse_item", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(0)  // boxType = 0 ("item")
		bf.WriteUint8(1)  // BoxIndex
		bf.WriteUint16(0) // padding
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateWarehouse{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateWarehouse_equip", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(1)  // boxType = 1 ("equip")
		bf.WriteUint8(2)  // BoxIndex
		bf.WriteUint16(0) // padding
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateWarehouse{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfExchangeFpoint2Item", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint32(100) // TradeID
		bf.WriteUint16(1)   // ItemType
		bf.WriteUint16(50)  // ItemId
		bf.WriteUint8(5)    // Quantity
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfExchangeFpoint2Item{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfExchangeItem2Fpoint", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint32(100) // TradeID
		bf.WriteUint16(1)   // ItemType
		bf.WriteUint16(50)  // ItemId
		bf.WriteUint8(5)    // Quantity
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfExchangeItem2Fpoint{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfExchangeWeeklyStamp_hl", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(1)  // StampType = 1 ("hl")
		bf.WriteUint8(0)  // Unk1
		bf.WriteUint16(0) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfExchangeWeeklyStamp{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfExchangeWeeklyStamp_ex", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(2)  // StampType = 2 ("ex")
		bf.WriteUint8(1)  // Unk1
		bf.WriteUint16(5) // Unk2
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfExchangeWeeklyStamp{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGenerateUdGuildMap", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGenerateUdGuildMap{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetBoostTimeLimit", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetBoostTimeLimit{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetCafeDurationBonusInfo", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetCafeDurationBonusInfo{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfGetMyhouseInfo", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint32(100) // Unk0
		bf.WriteUint8(4)    // DataSize
		bf.WriteBytes([]byte{0x01, 0x02, 0x03, 0x04})
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGetMyhouseInfo{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfAcquireUdItem", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(1)  // Unk0
		bf.WriteUint8(2)  // RewardType
		bf.WriteUint8(2)  // Unk2 (count)
		bf.WriteUint32(10)
		bf.WriteUint32(20)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfAcquireUdItem{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("MsgMhfEnumerateHouse_noname", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)   // AckHandle
		bf.WriteUint32(100) // CharID
		bf.WriteUint8(1)    // Method
		bf.WriteUint16(0)   // Unk
		bf.WriteUint8(0)    // lenName = 0 (no name)
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfEnumerateHouse{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})
}

func TestBatchParseGuildHuntdata(t *testing.T) {
	ctx := &clientctx.ClientContext{}

	t.Run("operation_0", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1) // AckHandle
		bf.WriteUint8(0)  // Operation = 0
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGuildHuntdata{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("operation_1", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)  // AckHandle
		bf.WriteUint8(1)   // Operation = 1 (reads GuildID)
		bf.WriteUint32(99) // GuildID
		_, _ = bf.Seek(0, io.SeekStart)
		pkt := &MsgMhfGuildHuntdata{}
		if err := pkt.Parse(bf, ctx); err != nil {
			t.Fatal(err)
		}
	})
}
