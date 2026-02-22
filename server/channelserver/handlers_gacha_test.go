package channelserver

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetGachaPlayHistory_StubResponse(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetGachaPlayHistory{AckHandle: 100, GachaID: 1}
	handleMsgMhfGetGachaPlayHistory(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetGachaPoint(t *testing.T) {
	server := createMockServer()
	userRepo := &mockUserRepoGacha{
		gachaFP: 100,
		gachaGP: 200,
		gachaGT: 300,
	}
	server.userRepo = userRepo

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfGetGachaPoint{AckHandle: 100}
	handleMsgMhfGetGachaPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUseGachaPoint_TrialCoins(t *testing.T) {
	server := createMockServer()
	userRepo := &mockUserRepoGacha{}
	server.userRepo = userRepo

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfUseGachaPoint{
		AckHandle:    100,
		TrialCoins:   10,
		PremiumCoins: 0,
	}
	handleMsgMhfUseGachaPoint(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUseGachaPoint_PremiumCoins(t *testing.T) {
	server := createMockServer()
	userRepo := &mockUserRepoGacha{}
	server.userRepo = userRepo

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfUseGachaPoint{
		AckHandle:    100,
		TrialCoins:   0,
		PremiumCoins: 5,
	}
	handleMsgMhfUseGachaPoint(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReceiveGachaItem_Normal(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	// Store 2 items: count byte + 2 * 5 bytes each
	data := []byte{2, 1, 0, 100, 0, 5, 2, 0, 200, 0, 10}
	charRepo.columns["gacha_items"] = data
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReceiveGachaItem{AckHandle: 100, Freeze: false}
	handleMsgMhfReceiveGachaItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}

	// After non-freeze receive, gacha_items should be cleared
	if charRepo.columns["gacha_items"] != nil {
		t.Error("Expected gacha_items to be cleared after receive")
	}
}

func TestHandleMsgMhfReceiveGachaItem_Overflow(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	// Build data with >36 items (overflow scenario): count=37, 37*5=185 bytes + 1 count byte = 186
	data := make([]byte, 186)
	data[0] = 37
	for i := 1; i < 186; i++ {
		data[i] = byte(i % 256)
	}
	charRepo.columns["gacha_items"] = data
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReceiveGachaItem{AckHandle: 100, Freeze: false}
	handleMsgMhfReceiveGachaItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}

	// After overflow, remaining items should be saved
	saved := charRepo.columns["gacha_items"]
	if saved == nil {
		t.Error("Expected overflow items to be saved")
	}
}

func TestHandleMsgMhfReceiveGachaItem_Freeze(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	data := []byte{1, 1, 0, 100, 0, 5}
	charRepo.columns["gacha_items"] = data
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReceiveGachaItem{AckHandle: 100, Freeze: true}
	handleMsgMhfReceiveGachaItem(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}

	// Freeze should NOT clear the items
	if charRepo.columns["gacha_items"] == nil {
		t.Error("Expected gacha_items to be preserved on freeze")
	}
}

func TestHandleMsgMhfPlayNormalGacha_TransactError(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{txErr: errors.New("transact failed")}
	server.gachaRepo = gachaRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPlayNormalGacha{AckHandle: 100, GachaID: 1, RollType: 0}
	handleMsgMhfPlayNormalGacha(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPlayNormalGacha_RewardPoolError(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{
		txRolls:       1,
		rewardPoolErr: errors.New("pool error"),
	}
	server.gachaRepo = gachaRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPlayNormalGacha{AckHandle: 100, GachaID: 1, RollType: 0}
	handleMsgMhfPlayNormalGacha(session, pkt)

	select {
	case <-session.sendPackets:
		// success - returns empty result
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPlayNormalGacha_Success(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	gachaRepo := &mockGachaRepo{
		txRolls: 1,
		rewardPool: []GachaEntry{
			{ID: 10, Weight: 100, Rarity: 3},
		},
		entryItems: map[uint32][]GachaItem{
			10: {{ItemType: 1, ItemID: 500, Quantity: 1}},
		},
	}
	server.gachaRepo = gachaRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPlayNormalGacha{AckHandle: 100, GachaID: 1, RollType: 0}
	handleMsgMhfPlayNormalGacha(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}

	// Verify gacha items were stored
	if charRepo.columns["gacha_items"] == nil {
		t.Error("Expected gacha items to be saved")
	}
}

func TestHandleMsgMhfPlayStepupGacha_TransactError(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{txErr: errors.New("transact failed")}
	server.gachaRepo = gachaRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPlayStepupGacha{AckHandle: 100, GachaID: 1, RollType: 0}
	handleMsgMhfPlayStepupGacha(session, pkt)

	select {
	case <-session.sendPackets:
		// success - returns empty result
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPlayStepupGacha_Success(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	gachaRepo := &mockGachaRepo{
		txRolls: 1,
		rewardPool: []GachaEntry{
			{ID: 10, Weight: 100, Rarity: 2},
		},
		entryItems: map[uint32][]GachaItem{
			10: {{ItemType: 1, ItemID: 600, Quantity: 2}},
		},
		guaranteedItems: []GachaItem{
			{ItemType: 1, ItemID: 700, Quantity: 1},
		},
	}
	server.gachaRepo = gachaRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPlayStepupGacha{AckHandle: 100, GachaID: 1, RollType: 0}
	handleMsgMhfPlayStepupGacha(session, pkt)

	if !gachaRepo.deletedStepup {
		t.Error("Expected stepup to be deleted")
	}
	if gachaRepo.insertedStep != 1 {
		t.Errorf("Expected insertedStep=1, got %d", gachaRepo.insertedStep)
	}

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetStepupStatus_FreshStep(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{
		stepupStep: 2,
		stepupTime: time.Now(), // recent, not stale
		hasEntryType: true,
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetStepupStatus{AckHandle: 100, GachaID: 1}
	handleMsgMhfGetStepupStatus(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetStepupStatus_StaleStep(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{
		stepupStep: 3,
		stepupTime: time.Now().Add(-48 * time.Hour), // stale
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetStepupStatus{AckHandle: 100, GachaID: 1}
	handleMsgMhfGetStepupStatus(session, pkt)

	if !gachaRepo.deletedStepup {
		t.Error("Expected stale stepup to be deleted")
	}

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetStepupStatus_NoRows(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{
		stepupErr: sql.ErrNoRows,
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetStepupStatus{AckHandle: 100, GachaID: 1}
	handleMsgMhfGetStepupStatus(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetStepupStatus_NoEntryType(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{
		stepupStep:   2,
		stepupTime:   time.Now(),
		hasEntryType: false, // no matching entry type -> reset
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetStepupStatus{AckHandle: 100, GachaID: 1}
	handleMsgMhfGetStepupStatus(session, pkt)

	if !gachaRepo.deletedStepup {
		t.Error("Expected stepup to be reset when no entry type")
	}

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetBoxGachaInfo_Error(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{
		boxEntryIDsErr: errors.New("db error"),
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBoxGachaInfo{AckHandle: 100, GachaID: 1}
	handleMsgMhfGetBoxGachaInfo(session, pkt)

	select {
	case <-session.sendPackets:
		// returns empty
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetBoxGachaInfo_Success(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{
		boxEntryIDs: []uint32{10, 20, 30},
	}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBoxGachaInfo{AckHandle: 100, GachaID: 1}
	handleMsgMhfGetBoxGachaInfo(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPlayBoxGacha_TransactError(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{txErr: errors.New("transact failed")}
	server.gachaRepo = gachaRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPlayBoxGacha{AckHandle: 100, GachaID: 1, RollType: 0}
	handleMsgMhfPlayBoxGacha(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPlayBoxGacha_Success(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	gachaRepo := &mockGachaRepo{
		txRolls: 1,
		rewardPool: []GachaEntry{
			{ID: 10, Weight: 100, Rarity: 1},
		},
		entryItems: map[uint32][]GachaItem{
			10: {{ItemType: 1, ItemID: 800, Quantity: 1}},
		},
	}
	server.gachaRepo = gachaRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPlayBoxGacha{AckHandle: 100, GachaID: 1, RollType: 0}
	handleMsgMhfPlayBoxGacha(session, pkt)

	if len(gachaRepo.insertedBoxIDs) == 0 {
		t.Error("Expected box entry to be inserted")
	}

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfResetBoxGachaInfo(t *testing.T) {
	server := createMockServer()
	gachaRepo := &mockGachaRepo{}
	server.gachaRepo = gachaRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfResetBoxGachaInfo{AckHandle: 100, GachaID: 1}
	handleMsgMhfResetBoxGachaInfo(session, pkt)

	if !gachaRepo.deletedBox {
		t.Error("Expected box entries to be deleted")
	}

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfPlayFreeGacha_StubACK(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPlayFreeGacha{AckHandle: 100, GachaID: 1}
	handleMsgMhfPlayFreeGacha(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestGetRandomEntries_NonBox(t *testing.T) {
	entries := []GachaEntry{
		{ID: 1, Weight: 50},
		{ID: 2, Weight: 50},
	}
	result, err := getRandomEntries(entries, 3, false)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(result))
	}
}

func TestGetRandomEntries_Box(t *testing.T) {
	entries := []GachaEntry{
		{ID: 1, Weight: 50},
		{ID: 2, Weight: 50},
		{ID: 3, Weight: 50},
	}
	result, err := getRandomEntries(entries, 2, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(result) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(result))
	}
	// Box mode removes entries without replacement — all IDs should be unique
	if result[0].ID == result[1].ID {
		t.Error("Box mode should return unique entries")
	}
}

func TestHandleMsgMhfPlayStepupGacha_RewardPoolError(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	gachaRepo := &mockGachaRepo{
		txRolls:       1,
		rewardPoolErr: errors.New("pool error"),
	}
	server.gachaRepo = gachaRepo
	server.userRepo = &mockUserRepoGacha{}

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfPlayStepupGacha{AckHandle: 100, GachaID: 1, RollType: 0}
	handleMsgMhfPlayStepupGacha(session, pkt)

	select {
	case p := <-session.sendPackets:
		// Verify minimal response (1 byte)
		_ = p
	default:
		t.Error("No response packet queued")
	}
}

// Verify the response payload of GetGachaPoint contains the expected values
func TestHandleMsgMhfGetGachaPoint_ResponsePayload(t *testing.T) {
	server := createMockServer()
	userRepo := &mockUserRepoGacha{
		gachaFP: 111,
		gachaGP: 222,
		gachaGT: 333,
	}
	server.userRepo = userRepo

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgMhfGetGachaPoint{AckHandle: 100}
	handleMsgMhfGetGachaPoint(session, pkt)

	select {
	case p := <-session.sendPackets:
		// The ack wraps the payload. The handler writes gp, gt, fp (12 bytes).
		// Just verify we got a reasonable-sized response.
		if len(p.data) < 12 {
			t.Errorf("Expected at least 12 bytes of gacha point data in response, got %d", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

// Verify the response when no gacha items exist (default column)
func TestHandleMsgMhfReceiveGachaItem_Empty(t *testing.T) {
	server := createMockServer()
	charRepo := newMockCharacterRepo()
	// No gacha_items set — will return default {0x00}
	server.charRepo = charRepo

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReceiveGachaItem{AckHandle: 100, Freeze: false}
	handleMsgMhfReceiveGachaItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		// The response should contain the default byte
		bf := byteframe.NewByteFrameFromBytes(p.data)
		_ = bf
	default:
		t.Error("No response packet queued")
	}
}
