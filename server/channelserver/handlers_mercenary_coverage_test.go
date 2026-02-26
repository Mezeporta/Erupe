package channelserver

import (
	"testing"

	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfLoadPartner(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfLoadPartner{AckHandle: 1}
	handleMsgMhfLoadPartner(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSavePartner(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfSavePartner{AckHandle: 1, RawDataPayload: []byte{1, 2, 3, 4}}
	handleMsgMhfSavePartner(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfLoadHunterNavi_G8(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	server.erupeConfig.RealClientMode = cfg.G10
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfLoadHunterNavi{AckHandle: 1}
	handleMsgMhfLoadHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfLoadHunterNavi_G7(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	server.erupeConfig.RealClientMode = cfg.G7
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfLoadHunterNavi{AckHandle: 1}
	handleMsgMhfLoadHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveHunterNavi_NoDiff(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	data := make([]byte, 100)
	pkt := &mhfpacket.MsgMhfSaveHunterNavi{
		AckHandle:      1,
		IsDataDiff:     false,
		RawDataPayload: data,
	}
	handleMsgMhfSaveHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveHunterNavi_Diff(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.columns["hunternavi"] = make([]byte, 552)
	server.charRepo = charRepo
	server.erupeConfig.RealClientMode = cfg.G10
	session := createMockSession(100, server)

	// Create a valid diff payload (deltacomp format: pairs of offset+data)
	// A simple diff: zero length means no changes
	diffData := make([]byte, 4) // minimal diff
	pkt := &mhfpacket.MsgMhfSaveHunterNavi{
		AckHandle:      1,
		IsDataDiff:     true,
		RawDataPayload: diffData,
	}
	handleMsgMhfSaveHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveHunterNavi_OversizedPayload(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	data := make([]byte, 5000) // > 4096
	pkt := &mhfpacket.MsgMhfSaveHunterNavi{
		AckHandle:      1,
		IsDataDiff:     false,
		RawDataPayload: data,
	}
	handleMsgMhfSaveHunterNavi(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfCreateMercenary_Success(t *testing.T) {
	server := createMockServer()
	server.mercenaryRepo = &mockMercenaryRepo{nextRastaID: 42}
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfCreateMercenary{AckHandle: 1}
	handleMsgMhfCreateMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfCreateMercenary_Error(t *testing.T) {
	server := createMockServer()
	server.mercenaryRepo = &mockMercenaryRepo{rastaIDErr: errNotFound}
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfCreateMercenary{AckHandle: 1}
	handleMsgMhfCreateMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveMercenary_Normal(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	mercData := make([]byte, 100)
	// Write a uint32 index at the start
	mercData[0] = 0
	mercData[1] = 0
	mercData[2] = 0
	mercData[3] = 1
	pkt := &mhfpacket.MsgMhfSaveMercenary{
		AckHandle:  1,
		GCP:        500,
		PactMercID: 10,
		MercData:   mercData,
	}
	handleMsgMhfSaveMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfSaveMercenary_Oversized(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfSaveMercenary{
		AckHandle: 1,
		MercData:  make([]byte, 70000),
	}
	handleMsgMhfSaveMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryM_EmptyData(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfReadMercenaryM{AckHandle: 1, CharID: 200}
	handleMsgMhfReadMercenaryM(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryM_WithData(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.columns["savemercenary"] = []byte{0x01, 0x02, 0x03, 0x04}
	server.charRepo = charRepo
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfReadMercenaryM{AckHandle: 1, CharID: 100}
	handleMsgMhfReadMercenaryM(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfContractMercenary_Op0(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfContractMercenary{AckHandle: 1, Op: 0, CID: 200, PactMercID: 42}
	handleMsgMhfContractMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfContractMercenary_Op1(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfContractMercenary{AckHandle: 1, Op: 1}
	handleMsgMhfContractMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfContractMercenary_Op2(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfContractMercenary{AckHandle: 1, Op: 2, CID: 200}
	handleMsgMhfContractMercenary(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryW_NoPact(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo
	server.mercenaryRepo = &mockMercenaryRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfReadMercenaryW{AckHandle: 1, Op: 0}
	handleMsgMhfReadMercenaryW(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryW_WithPact(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	charRepo.ints["pact_id"] = 42
	server.charRepo = charRepo
	server.mercenaryRepo = &mockMercenaryRepo{}
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfReadMercenaryW{AckHandle: 1, Op: 0}
	handleMsgMhfReadMercenaryW(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfReadMercenaryW_Op2(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo
	server.mercenaryRepo = &mockMercenaryRepo{}
	session := createMockSession(100, server)

	// Op 2 skips loan enumeration
	pkt := &mhfpacket.MsgMhfReadMercenaryW{AckHandle: 1, Op: 2}
	handleMsgMhfReadMercenaryW(session, pkt)
	<-session.sendPackets
}

func TestHandleMsgMhfLoadOtomoAirou(t *testing.T) {
	server := createMockServer()
	server.charRepo = newMockCharacterRepo()
	session := createMockSession(100, server)

	pkt := &mhfpacket.MsgMhfLoadOtomoAirou{AckHandle: 1}
	handleMsgMhfLoadOtomoAirou(session, pkt)
	<-session.sendPackets
}
