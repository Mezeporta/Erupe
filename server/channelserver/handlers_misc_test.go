package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

// Test handlers with simple responses

func TestHandleMsgMhfGetEarthStatus(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthStatus{
		AckHandle: 12345,
	}

	handleMsgMhfGetEarthStatus(session, pkt)

	select {
	case p := <-session.sendPackets:
		if p.data == nil {
			t.Error("Response packet data should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetEarthValue_Type1(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthValue{
		AckHandle: 12345,
		ReqType:   1,
	}

	handleMsgMhfGetEarthValue(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetEarthValue_Type2(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthValue{
		AckHandle: 12345,
		ReqType:   2,
	}

	handleMsgMhfGetEarthValue(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetEarthValue_Type3(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthValue{
		AckHandle: 12345,
		ReqType:   3,
	}

	handleMsgMhfGetEarthValue(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetEarthValue_UnknownType(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetEarthValue{
		AckHandle: 12345,
		ReqType:   99, // Unknown type
	}

	handleMsgMhfGetEarthValue(session, pkt)

	select {
	case p := <-session.sendPackets:
		// Should still return a response (empty values)
		if p.data == nil {
			t.Error("Response packet data should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadBeatLevel(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevel{
		AckHandle:    12345,
		ValidIDCount: 2,
		IDs:          [16]uint32{1, 2},
	}

	handleMsgMhfReadBeatLevel(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadBeatLevel_NoIDs(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevel{
		AckHandle:    12345,
		ValidIDCount: 0,
		IDs:          [16]uint32{},
	}

	handleMsgMhfReadBeatLevel(session, pkt)

	select {
	case p := <-session.sendPackets:
		if p.data == nil {
			t.Error("Response packet data should not be nil")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUpdateBeatLevel(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateBeatLevel{
		AckHandle: 12345,
	}

	handleMsgMhfUpdateBeatLevel(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Test empty handlers don't panic

func TestHandleMsgMhfStampcardPrize(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfStampcardPrize panicked: %v", r)
		}
	}()

	handleMsgMhfStampcardPrize(session, nil)
}

func TestHandleMsgMhfUnreserveSrg(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUnreserveSrg{
		AckHandle: 12345,
	}

	handleMsgMhfUnreserveSrg(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadBeatLevelAllRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevelAllRanking{
		AckHandle: 12345,
	}

	handleMsgMhfReadBeatLevelAllRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadBeatLevelMyRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevelMyRanking{
		AckHandle: 12345,
	}

	handleMsgMhfReadBeatLevelMyRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadLastWeekBeatRanking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadLastWeekBeatRanking{
		AckHandle: 12345,
	}

	handleMsgMhfReadLastWeekBeatRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetFixedSeibatuRankingTable(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetFixedSeibatuRankingTable{
		AckHandle: 12345,
	}

	handleMsgMhfGetFixedSeibatuRankingTable(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfKickExportForce(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfKickExportForce panicked: %v", r)
		}
	}()

	handleMsgMhfKickExportForce(session, nil)
}

func TestHandleMsgMhfRegistSpabiTime(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfRegistSpabiTime panicked: %v", r)
		}
	}()

	handleMsgMhfRegistSpabiTime(session, nil)
}

func TestHandleMsgMhfDebugPostValue(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfDebugPostValue panicked: %v", r)
		}
	}()

	handleMsgMhfDebugPostValue(session, nil)
}

func TestHandleMsgMhfGetCogInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetCogInfo panicked: %v", r)
		}
	}()

	handleMsgMhfGetCogInfo(session, nil)
}

// Additional handler tests for coverage

func TestHandleMsgMhfGetNotice(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetNotice{
		AckHandle: 12345,
	}

	handleMsgMhfGetNotice(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostNotice(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostNotice{
		AckHandle: 12345,
	}

	handleMsgMhfPostNotice(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetRandFromTable(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRandFromTable{
		AckHandle: 12345,
		Results:   3,
	}

	handleMsgMhfGetRandFromTable(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetSenyuDailyCount(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetSenyuDailyCount{
		AckHandle: 12345,
	}

	handleMsgMhfGetSenyuDailyCount(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetSeibattle(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetSeibattle{
		AckHandle: 12345,
	}

	handleMsgMhfGetSeibattle(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPostSeibattle(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPostSeibattle{
		AckHandle: 12345,
	}

	handleMsgMhfPostSeibattle(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetDailyMissionMaster(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetDailyMissionMaster panicked: %v", r)
		}
	}()

	handleMsgMhfGetDailyMissionMaster(session, nil)
}

func TestHandleMsgMhfGetDailyMissionPersonal(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetDailyMissionPersonal panicked: %v", r)
		}
	}()

	handleMsgMhfGetDailyMissionPersonal(session, nil)
}

func TestHandleMsgMhfSetDailyMissionPersonal(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfSetDailyMissionPersonal panicked: %v", r)
		}
	}()

	handleMsgMhfSetDailyMissionPersonal(session, nil)
}

func TestHandleMsgMhfGetUdShopCoin(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdShopCoin{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdShopCoin(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUseUdShopCoin(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfUseUdShopCoin panicked: %v", r)
		}
	}()

	handleMsgMhfUseUdShopCoin(session, nil)
}

func TestHandleMsgMhfGetLobbyCrowd(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetLobbyCrowd{
		AckHandle: 12345,
	}

	handleMsgMhfGetLobbyCrowd(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Distribution struct tests
func TestDistributionStruct(t *testing.T) {
	dist := Distribution{
		ID:              1,
		MinHR:           1,
		MaxHR:           999,
		MinSR:           0,
		MaxSR:           999,
		MinGR:           0,
		MaxGR:           999,
		TimesAcceptable: 1,
		TimesAccepted:   0,
		EventName:       "Test Event",
		Description:     "Test Description",
		Selection:       false,
	}

	if dist.ID != 1 {
		t.Errorf("ID = %d, want 1", dist.ID)
	}
	if dist.EventName != "Test Event" {
		t.Errorf("EventName = %s, want Test Event", dist.EventName)
	}
}

func TestDistributionItemStruct(t *testing.T) {
	item := DistributionItem{
		ItemType: 1,
		ID:       100,
		ItemID:   1234,
		Quantity: 10,
	}

	if item.ItemType != 1 {
		t.Errorf("ItemType = %d, want 1", item.ItemType)
	}
	if item.ItemID != 1234 {
		t.Errorf("ItemID = %d, want 1234", item.ItemID)
	}
}
