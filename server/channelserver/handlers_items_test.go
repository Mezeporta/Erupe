package channelserver

import (
	"testing"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfitem"
	"erupe-ce/network/mhfpacket"
)

// --- userGetItems tests ---

func TestUserGetItems_NilData(t *testing.T) {
	server := createMockServer()
	userMock := &mockUserRepoForItems{itemBoxData: nil}
	server.userRepo = userMock
	session := createMockSession(1, server)
	session.userID = 1

	items := userGetItems(session)

	if len(items) != 0 {
		t.Errorf("Expected empty items, got %d", len(items))
	}
}

func TestUserGetItems_DBError(t *testing.T) {
	server := createMockServer()
	userMock := &mockUserRepoForItems{itemBoxErr: errNotFound}
	server.userRepo = userMock
	session := createMockSession(1, server)
	session.userID = 1

	items := userGetItems(session)

	if len(items) != 0 {
		t.Errorf("Expected empty items on error, got %d", len(items))
	}
}

func TestUserGetItems_ParsesData(t *testing.T) {
	// Build serialized item box with 1 item
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(1) // numStacks
	bf.WriteUint16(0) // unused
	// Item stack: warehouseID(4) + itemID(2) + quantity(2) + unk0(4) = 12 bytes
	bf.WriteUint32(100)  // warehouseID
	bf.WriteUint16(500)  // itemID
	bf.WriteUint16(3)    // quantity
	bf.WriteUint32(0)    // unk0

	server := createMockServer()
	userMock := &mockUserRepoForItems{itemBoxData: bf.Data()}
	server.userRepo = userMock
	session := createMockSession(1, server)
	session.userID = 1

	items := userGetItems(session)

	if len(items) != 1 {
		t.Fatalf("Expected 1 item, got %d", len(items))
	}
	if items[0].Item.ItemID != 500 {
		t.Errorf("ItemID = %d, want 500", items[0].Item.ItemID)
	}
	if items[0].Quantity != 3 {
		t.Errorf("Quantity = %d, want 3", items[0].Quantity)
	}
}

// --- handleMsgMhfCheckWeeklyStamp tests ---

func TestCheckWeeklyStamp_InvalidType(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCheckWeeklyStamp{
		AckHandle: 100,
		StampType: "invalid",
	}

	handleMsgMhfCheckWeeklyStamp(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestCheckWeeklyStamp_FirstCheck(t *testing.T) {
	server := createMockServer()
	stampMock := &mockStampRepoForItems{
		checkedErr: errNotFound, // no existing record
		totals:     [2]uint16{0, 0},
	}
	server.stampRepo = stampMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCheckWeeklyStamp{
		AckHandle: 100,
		StampType: "hl",
	}

	handleMsgMhfCheckWeeklyStamp(session, pkt)

	if !stampMock.initCalled {
		t.Error("Init should be called on first check")
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 14 {
			t.Errorf("Response too short: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestCheckWeeklyStamp_WithinWeek(t *testing.T) {
	server := createMockServer()
	stampMock := &mockStampRepoForItems{
		checkedTime: TimeAdjusted(), // checked right now (within this week)
		totals:      [2]uint16{3, 1},
	}
	server.stampRepo = stampMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCheckWeeklyStamp{
		AckHandle: 100,
		StampType: "hl",
	}

	handleMsgMhfCheckWeeklyStamp(session, pkt)

	if stampMock.incrementCalled {
		t.Error("IncrementTotal should not be called within same week")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestCheckWeeklyStamp_WeekRollover(t *testing.T) {
	server := createMockServer()
	stampMock := &mockStampRepoForItems{
		checkedTime: TimeWeekStart().Add(-24 * time.Hour), // before this week
		totals:      [2]uint16{5, 2},
	}
	server.stampRepo = stampMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCheckWeeklyStamp{
		AckHandle: 100,
		StampType: "ex",
	}

	handleMsgMhfCheckWeeklyStamp(session, pkt)

	if !stampMock.incrementCalled {
		t.Error("IncrementTotal should be called after week rollover")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestCheckWeeklyStamp_GetTotalsError(t *testing.T) {
	server := createMockServer()
	stampMock := &mockStampRepoForItems{
		checkedTime: TimeAdjusted(),
		totalsErr:   errNotFound,
	}
	server.stampRepo = stampMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfCheckWeeklyStamp{
		AckHandle: 100,
		StampType: "hl",
	}

	// Should not panic; logs warning, returns zeros
	handleMsgMhfCheckWeeklyStamp(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfExchangeWeeklyStamp tests ---

func TestExchangeWeeklyStamp_InvalidType(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfExchangeWeeklyStamp{
		AckHandle: 100,
		StampType: "invalid",
	}

	handleMsgMhfExchangeWeeklyStamp(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestExchangeWeeklyStamp_HL(t *testing.T) {
	server := createMockServer()
	stampMock := &mockStampRepoForItems{
		exchangeResult: [2]uint16{10, 5},
	}
	houseMock := newMockHouseRepoForItems()
	server.stampRepo = stampMock
	server.houseRepo = houseMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfExchangeWeeklyStamp{
		AckHandle: 100,
		StampType: "hl",
	}

	handleMsgMhfExchangeWeeklyStamp(session, pkt)

	// Verify warehouse gift box was updated (index 10)
	if houseMock.setData[10] == nil {
		t.Error("Gift box should be updated with ticket item")
	}
	// Parse the gift box to verify the item
	if len(houseMock.setData[10]) > 0 {
		bf := byteframe.NewByteFrameFromBytes(houseMock.setData[10])
		count := bf.ReadUint16()
		if count != 1 {
			t.Errorf("Expected 1 item in gift box, got %d", count)
		}
		bf.ReadUint16() // unused
		item := mhfitem.ReadWarehouseItem(bf)
		if item.Item.ItemID != 1630 {
			t.Errorf("ItemID = %d, want 1630 (HL ticket)", item.Item.ItemID)
		}
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestExchangeWeeklyStamp_EX(t *testing.T) {
	server := createMockServer()
	stampMock := &mockStampRepoForItems{
		exchangeResult: [2]uint16{10, 5},
	}
	houseMock := newMockHouseRepoForItems()
	server.stampRepo = stampMock
	server.houseRepo = houseMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfExchangeWeeklyStamp{
		AckHandle: 100,
		StampType: "ex",
	}

	handleMsgMhfExchangeWeeklyStamp(session, pkt)

	if houseMock.setData[10] == nil {
		t.Error("Gift box should be updated with ticket item")
	}
	if len(houseMock.setData[10]) > 0 {
		bf := byteframe.NewByteFrameFromBytes(houseMock.setData[10])
		count := bf.ReadUint16()
		if count != 1 {
			t.Errorf("Expected 1 item in gift box, got %d", count)
		}
		bf.ReadUint16() // unused
		item := mhfitem.ReadWarehouseItem(bf)
		if item.Item.ItemID != 1631 {
			t.Errorf("ItemID = %d, want 1631 (EX ticket)", item.Item.ItemID)
		}
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestExchangeWeeklyStamp_ExchangeError(t *testing.T) {
	server := createMockServer()
	stampMock := &mockStampRepoForItems{
		exchangeErr: errNotFound,
	}
	server.stampRepo = stampMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfExchangeWeeklyStamp{
		AckHandle: 100,
		StampType: "hl",
	}

	handleMsgMhfExchangeWeeklyStamp(session, pkt)

	// Should return fail ack
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestExchangeWeeklyStamp_Yearly(t *testing.T) {
	server := createMockServer()
	stampMock := &mockStampRepoForItems{
		yearlyResult: [2]uint16{20, 10},
	}
	houseMock := newMockHouseRepoForItems()
	server.stampRepo = stampMock
	server.houseRepo = houseMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfExchangeWeeklyStamp{
		AckHandle:    100,
		StampType:    "ex",
		ExchangeType: 10, // Yearly
	}

	handleMsgMhfExchangeWeeklyStamp(session, pkt)

	if houseMock.setData[10] == nil {
		t.Error("Gift box should be updated with yearly ticket")
	}
	if len(houseMock.setData[10]) > 0 {
		bf := byteframe.NewByteFrameFromBytes(houseMock.setData[10])
		count := bf.ReadUint16()
		if count != 1 {
			t.Errorf("Expected 1 item in gift box, got %d", count)
		}
		bf.ReadUint16() // unused
		item := mhfitem.ReadWarehouseItem(bf)
		if item.Item.ItemID != 2210 {
			t.Errorf("ItemID = %d, want 2210 (yearly ticket)", item.Item.ItemID)
		}
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}
