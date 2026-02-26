package channelserver

import (
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"testing"
)

func TestHandleMsgMhfEnumerateHouse_Method3_SearchByName(t *testing.T) {
	srv := createMockServer()
	srv.erupeConfig.RealClientMode = cfg.ZZ
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateHouse{AckHandle: 1, Method: 3, Name: "TestHouse"}
	handleMsgMhfEnumerateHouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateHouse_Method4_ByCharID(t *testing.T) {
	srv := createMockServer()
	srv.erupeConfig.RealClientMode = cfg.ZZ
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateHouse{AckHandle: 1, Method: 4, CharID: 200}
	handleMsgMhfEnumerateHouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateHouse_Method5_RecentVisitors(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateHouse{AckHandle: 1, Method: 5}
	handleMsgMhfEnumerateHouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateHouse_Method1_Friends(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	charRepo := newMockCharacterRepo()
	charRepo.strings["friends"] = ""
	srv.charRepo = charRepo
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateHouse{AckHandle: 1, Method: 1}
	handleMsgMhfEnumerateHouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateHouse_Method2_GuildMembers(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateHouse{AckHandle: 1, Method: 2}
	handleMsgMhfEnumerateHouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateHouse_Method2_NoGuild(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	srv.guildRepo = &mockGuildRepo{getErr: errNotFound}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateHouse{AckHandle: 1, Method: 2}
	handleMsgMhfEnumerateHouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfSaveDecoMyset_ShortPayload(t *testing.T) {
	srv := createMockServer()
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfSaveDecoMyset{AckHandle: 1, RawDataPayload: []byte{0x00, 0x01}}
	handleMsgMhfSaveDecoMyset(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfSaveDecoMyset_WithData(t *testing.T) {
	srv := createMockServer()
	charRepo := newMockCharacterRepo()
	// Pre-populate with version byte + 0 sets
	charRepo.columns["decomyset"] = []byte{0x01, 0x00}
	srv.charRepo = charRepo
	srv.erupeConfig.RealClientMode = cfg.ZZ

	s := createMockSession(100, srv)

	// Build payload: version byte + 1 set with index 0 + 76 bytes of data
	payload := make([]byte, 3+2+76)
	payload[0] = 0x01 // version
	payload[1] = 0x01 // count
	payload[2] = 0x00 // padding

	pkt := &mhfpacket.MsgMhfSaveDecoMyset{AckHandle: 1, RawDataPayload: payload}
	handleMsgMhfSaveDecoMyset(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfInfoTournament_Type2(t *testing.T) {
	srv := createMockServer()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfInfoTournament{AckHandle: 1, QueryType: 2}
	handleMsgMhfInfoTournament(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateInterior_Normal(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateInterior{AckHandle: 1, InteriorData: make([]byte, 20)}
	handleMsgMhfUpdateInterior(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateInterior_TooLarge(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateInterior{AckHandle: 1, InteriorData: make([]byte, 100)}
	handleMsgMhfUpdateInterior(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateMyhouseInfo_Normal(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateMyhouseInfo{AckHandle: 1, Data: make([]byte, 9)}
	handleMsgMhfUpdateMyhouseInfo(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateMyhouseInfo_TooLarge(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateMyhouseInfo{AckHandle: 1, Data: make([]byte, 600)}
	handleMsgMhfUpdateMyhouseInfo(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfGetMyhouseInfo(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfGetMyhouseInfo{AckHandle: 1}
	handleMsgMhfGetMyhouseInfo(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateTitle(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateTitle{AckHandle: 1}
	handleMsgMhfEnumerateTitle(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfAcquireTitle(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfAcquireTitle{AckHandle: 1, TitleIDs: []uint16{1, 2, 3}}
	handleMsgMhfAcquireTitle(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateHouse(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateHouse{AckHandle: 1, State: 2, Password: "1234"}
	handleMsgMhfUpdateHouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfOperateWarehouse_Op0(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOperateWarehouse{AckHandle: 1, Operation: 0}
	handleMsgMhfOperateWarehouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfOperateWarehouse_Op1(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOperateWarehouse{AckHandle: 1, Operation: 1}
	handleMsgMhfOperateWarehouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfOperateWarehouse_Op2_Rename(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOperateWarehouse{AckHandle: 1, Operation: 2, BoxType: 0, BoxIndex: 1, Name: "MyBox"}
	handleMsgMhfOperateWarehouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfOperateWarehouse_Op3(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOperateWarehouse{AckHandle: 1, Operation: 3}
	handleMsgMhfOperateWarehouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfOperateWarehouse_Op4(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOperateWarehouse{AckHandle: 1, Operation: 4}
	handleMsgMhfOperateWarehouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateWarehouse_Items(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateWarehouse{AckHandle: 1, BoxType: 0, BoxIndex: 0}
	handleMsgMhfEnumerateWarehouse(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateWarehouse_Equipment(t *testing.T) {
	srv := createMockServer()
	srv.houseRepo = newMockHouseRepoForItems()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateWarehouse{AckHandle: 1, BoxType: 1, BoxIndex: 0}
	handleMsgMhfEnumerateWarehouse(s, pkt)
	<-s.sendPackets
}
