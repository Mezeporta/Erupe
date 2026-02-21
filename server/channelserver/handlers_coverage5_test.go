package channelserver

import (
	"testing"

	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

// =============================================================================
// equipSkinHistSize: pure function, tests all 3 config branches
// =============================================================================

func TestEquipSkinHistSize_Default(t *testing.T) {
	got := equipSkinHistSize(cfg.ZZ)
	if got != 3200 {
		t.Errorf("equipSkinHistSize(ZZ) = %d, want 3200", got)
	}
}

func TestEquipSkinHistSize_Z2(t *testing.T) {
	got := equipSkinHistSize(cfg.Z2)
	if got != 2560 {
		t.Errorf("equipSkinHistSize(Z2) = %d, want 2560", got)
	}
}

func TestEquipSkinHistSize_Z1(t *testing.T) {
	got := equipSkinHistSize(cfg.Z1)
	if got != 1280 {
		t.Errorf("equipSkinHistSize(Z1) = %d, want 1280", got)
	}
}

func TestEquipSkinHistSize_OlderMode(t *testing.T) {
	got := equipSkinHistSize(cfg.G1)
	if got != 1280 {
		t.Errorf("equipSkinHistSize(G1) = %d, want 1280", got)
	}
}

// =============================================================================
// DB-free guild handlers: simple ack stubs
// =============================================================================

func TestHandleMsgMhfAddGuildMissionCount(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfAddGuildMissionCount(session, &mhfpacket.MsgMhfAddGuildMissionCount{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfSetGuildMissionTarget(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfSetGuildMissionTarget(session, &mhfpacket.MsgMhfSetGuildMissionTarget{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfCancelGuildMissionTarget(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfCancelGuildMissionTarget(session, &mhfpacket.MsgMhfCancelGuildMissionTarget{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfGetGuildMissionRecord(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetGuildMissionRecord(session, &mhfpacket.MsgMhfGetGuildMissionRecord{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfAcquireGuildTresureSouvenir(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfAcquireGuildTresureSouvenir(session, &mhfpacket.MsgMhfAcquireGuildTresureSouvenir{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

func TestHandleMsgMhfGetUdGuildMapInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetUdGuildMapInfo(session, &mhfpacket.MsgMhfGetUdGuildMapInfo{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

// =============================================================================
// DB-free guild mission list handler (large static data)
// =============================================================================

func TestHandleMsgMhfGetGuildMissionList(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	handleMsgMhfGetGuildMissionList(session, &mhfpacket.MsgMhfGetGuildMissionList{
		AckHandle: 1,
	})

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("response should have data")
		}
	default:
		t.Error("no response queued")
	}
}

// handleMsgMhfEnumerateUnionItem requires DB (calls userGetItems)

// handleMsgMhfRegistSpabiTime, handleMsgMhfKickExportForce, handleMsgMhfUseUdShopCoin
// are tested in handlers_misc_test.go

// handleMsgMhfGetUdShopCoin and handleMsgMhfGetLobbyCrowd are tested in handlers_misc_test.go

// handleMsgMhfEnumerateGuacot requires DB (calls getGoocooData)

// handleMsgMhfPostRyoudama is tested in handlers_caravan_test.go
// handleMsgMhfResetTitle is tested in handlers_coverage2_test.go
