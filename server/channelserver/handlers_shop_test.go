package channelserver

import (
	"errors"
	"testing"

	"erupe-ce/common/byteframe"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfEnumerateShop_Case1_G7EarlyReturn(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.G7

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateShop{
		AckHandle: 100,
		ShopType:  1,
		ShopID:    0,
	}
	handleMsgMhfEnumerateShop(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateShop_Case1_GachaList(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.ZZ

	gachaRepo := &mockGachaRepo{
		gachas: []Gacha{
			{ID: 1, Name: "TestGacha", MinGR: 0, MinHR: 0, GachaType: 1},
		},
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateShop{
		AckHandle: 100,
		ShopType:  1,
		ShopID:    0,
	}
	handleMsgMhfEnumerateShop(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateShop_Case1_ListShopError(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.ZZ

	gachaRepo := &mockGachaRepo{
		listShopErr: errors.New("db error"),
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateShop{
		AckHandle: 100,
		ShopType:  1,
		ShopID:    0,
	}
	handleMsgMhfEnumerateShop(session, pkt)

	select {
	case <-session.sendPackets:
		// returns empty on error
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateShop_Case2_GachaDetail(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.ZZ

	gachaRepo := &mockGachaRepo{
		shopType: 1, // non-box
		allEntries: []GachaEntry{
			{ID: 10, EntryType: 1, ItemType: 1, ItemNumber: 100, ItemQuantity: 5,
				Weight: 50, Rarity: 2, Rolls: 1, FrontierPoints: 10, DailyLimit: 3, Name: "Item1"},
		},
		entryItems: map[uint32][]GachaItem{
			10: {{ItemType: 1, ItemID: 500, Quantity: 1}},
		},
		weightDivisor: 1.0,
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateShop{
		AckHandle: 100,
		ShopType:  2,
		ShopID:    1,
	}
	handleMsgMhfEnumerateShop(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateShop_Case2_AllEntriesError(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.ZZ

	gachaRepo := &mockGachaRepo{
		allEntriesErr: errors.New("db error"),
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateShop{
		AckHandle: 100,
		ShopType:  2,
		ShopID:    1,
	}
	handleMsgMhfEnumerateShop(session, pkt)

	select {
	case <-session.sendPackets:
		// returns empty on error
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateShop_Case10_ShopItems(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.ZZ

	shopRepo := &mockShopRepo{
		shopItems: []ShopItem{
			{ID: 1, ItemID: 100, Cost: 500, Quantity: 10, MinHR: 1},
			{ID: 2, ItemID: 200, Cost: 1000, Quantity: 5, MinHR: 3},
			{ID: 3, ItemID: 300, Cost: 2000, Quantity: 1, MinHR: 5},
		},
	}
	server.shopRepo = shopRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateShop{
		AckHandle: 100,
		ShopType:  10,
		ShopID:    0,
		Limit:     2, // Limit to 2 items
	}
	handleMsgMhfEnumerateShop(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateShop_Cases3to9(t *testing.T) {
	for _, shopType := range []uint8{3, 4, 5, 6, 7, 8, 9} {
		server := createMockServer()
		server.erupeConfig.RealClientMode = cfg.ZZ

		shopRepo := &mockShopRepo{
			shopItems: []ShopItem{
				{ID: 1, ItemID: 100, Cost: 500, Quantity: 10},
			},
		}
		server.shopRepo = shopRepo

		session := createMockSession(1, server)

		pkt := &mhfpacket.MsgMhfEnumerateShop{
			AckHandle: 100,
			ShopType:  shopType,
			ShopID:    0,
			Limit:     100,
		}
		handleMsgMhfEnumerateShop(session, pkt)

		select {
		case <-session.sendPackets:
			// success
		default:
			t.Errorf("No response for shop type %d", shopType)
		}
	}
}

func TestHandleMsgMhfAcquireExchangeShop_RecordsPurchases(t *testing.T) {
	server := createMockServer()
	shopRepo := &mockShopRepo{}
	server.shopRepo = shopRepo

	session := createMockSession(1, server)

	// Build payload: 2 exchanges, one with non-zero hash, one with zero hash
	payload := byteframe.NewByteFrame()
	payload.WriteUint16(2)       // count
	payload.WriteUint32(12345)   // itemHash 1
	payload.WriteUint32(3)       // buyCount 1
	payload.WriteUint32(0)       // itemHash 2 (zero, should be skipped)
	payload.WriteUint32(1)       // buyCount 2

	pkt := &mhfpacket.MsgMhfAcquireExchangeShop{
		AckHandle:      100,
		RawDataPayload: payload.Data(),
	}
	handleMsgMhfAcquireExchangeShop(session, pkt)

	if len(shopRepo.purchases) != 1 {
		t.Errorf("Expected 1 purchase recorded (skipping zero hash), got %d", len(shopRepo.purchases))
	}
	if len(shopRepo.purchases) > 0 && shopRepo.purchases[0].itemHash != 12345 {
		t.Errorf("Expected itemHash=12345, got %d", shopRepo.purchases[0].itemHash)
	}

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfExchangeFpoint2Item_Success(t *testing.T) {
	server := createMockServer()
	shopRepo := &mockShopRepo{
		fpointQuantity: 1,
		fpointValue:    100,
	}
	server.shopRepo = shopRepo

	userRepo := &mockUserRepoGacha{fpDeductBalance: 900}
	server.userRepo = userRepo

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfExchangeFpoint2Item{
		AckHandle: 100,
		TradeID:   1,
		Quantity:  1,
	}
	handleMsgMhfExchangeFpoint2Item(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfExchangeFpoint2Item_GetFpointItemError(t *testing.T) {
	server := createMockServer()
	shopRepo := &mockShopRepo{
		fpointItemErr: errors.New("not found"),
	}
	server.shopRepo = shopRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfExchangeFpoint2Item{
		AckHandle: 100,
		TradeID:   999,
		Quantity:  1,
	}
	handleMsgMhfExchangeFpoint2Item(session, pkt)

	select {
	case <-session.sendPackets:
		// returns fail
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfExchangeFpoint2Item_DeductError(t *testing.T) {
	server := createMockServer()
	shopRepo := &mockShopRepo{
		fpointQuantity: 1,
		fpointValue:    100,
	}
	server.shopRepo = shopRepo

	userRepo := &mockUserRepoGacha{fpDeductErr: errors.New("insufficient")}
	server.userRepo = userRepo

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfExchangeFpoint2Item{
		AckHandle: 100,
		TradeID:   1,
		Quantity:  1,
	}
	handleMsgMhfExchangeFpoint2Item(session, pkt)

	select {
	case <-session.sendPackets:
		// returns fail
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfExchangeItem2Fpoint_Success(t *testing.T) {
	server := createMockServer()
	shopRepo := &mockShopRepo{
		fpointQuantity: 1,
		fpointValue:    50,
	}
	server.shopRepo = shopRepo

	userRepo := &mockUserRepoGacha{fpCreditBalance: 1050}
	server.userRepo = userRepo

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfExchangeItem2Fpoint{
		AckHandle: 100,
		TradeID:   1,
		Quantity:  1,
	}
	handleMsgMhfExchangeItem2Fpoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfExchangeItem2Fpoint_GetFpointItemError(t *testing.T) {
	server := createMockServer()
	shopRepo := &mockShopRepo{
		fpointItemErr: errors.New("not found"),
	}
	server.shopRepo = shopRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfExchangeItem2Fpoint{
		AckHandle: 100,
		TradeID:   999,
		Quantity:  1,
	}
	handleMsgMhfExchangeItem2Fpoint(session, pkt)

	select {
	case <-session.sendPackets:
		// returns fail
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfExchangeItem2Fpoint_CreditError(t *testing.T) {
	server := createMockServer()
	shopRepo := &mockShopRepo{
		fpointQuantity: 1,
		fpointValue:    50,
	}
	server.shopRepo = shopRepo

	userRepo := &mockUserRepoGacha{fpCreditErr: errors.New("credit error")}
	server.userRepo = userRepo

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfExchangeItem2Fpoint{
		AckHandle: 100,
		TradeID:   1,
		Quantity:  1,
	}
	handleMsgMhfExchangeItem2Fpoint(session, pkt)

	select {
	case <-session.sendPackets:
		// returns fail
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetFpointExchangeList_Z2Mode(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.Z2

	shopRepo := &mockShopRepo{
		fpointExchanges: []FPointExchange{
			{ID: 1, ItemType: 1, ItemID: 100, Quantity: 5, FPoints: 10, Buyable: true},
			{ID: 2, ItemType: 2, ItemID: 200, Quantity: 1, FPoints: 50, Buyable: false},
		},
	}
	server.shopRepo = shopRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetFpointExchangeList{AckHandle: 100}
	handleMsgMhfGetFpointExchangeList(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetFpointExchangeList_ZZMode(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.ZZ

	shopRepo := &mockShopRepo{
		fpointExchanges: []FPointExchange{
			{ID: 1, ItemType: 1, ItemID: 100, Quantity: 5, FPoints: 10, Buyable: true},
		},
	}
	server.shopRepo = shopRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetFpointExchangeList{AckHandle: 100}
	handleMsgMhfGetFpointExchangeList(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}
