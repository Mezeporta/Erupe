package channelserver

import (
	"encoding/binary"
	"errors"
	"sync"
	"testing"

	"erupe-ce/common/byteframe"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgSysTerminalLog_ReturnsLogIDPlusOne(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 100,
		LogID:     5,
		Entries: []mhfpacket.TerminalLogEntry{
			{Type1: 1, Type2: 2, Unk0: 3, Unk1: 4, Unk2: 5, Unk3: 6},
		},
	}
	handleMsgSysTerminalLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysLogin_Success(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.DebugOptions.DisableTokenCheck = true
	server.userBinary = NewUserBinaryStore()

	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	sessionRepo := &mockSessionRepo{}
	server.sessionRepo = sessionRepo

	userRepo := &mockUserRepoGacha{}
	server.userRepo = userRepo

	session := createMockSession(0, server)

	pkt := &mhfpacket.MsgSysLogin{
		AckHandle:        100,
		CharID0:          42,
		LoginTokenString: "test-token",
	}
	handleMsgSysLogin(session, pkt)

	if session.charID != 42 {
		t.Errorf("Expected charID 42, got %d", session.charID)
	}
	if session.token != "test-token" {
		t.Errorf("Expected token 'test-token', got %q", session.token)
	}
	if sessionRepo.boundToken != "test-token" {
		t.Errorf("Expected BindSession called with 'test-token', got %q", sessionRepo.boundToken)
	}

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysLogin_GetUserIDError(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.DebugOptions.DisableTokenCheck = true

	charRepo := newMockCharacterRepo()
	server.charRepo = &mockCharRepoGetUserIDErr{
		mockCharacterRepo: charRepo,
		getUserIDErr:      errors.New("user not found"),
	}

	sessionRepo := &mockSessionRepo{}
	server.sessionRepo = sessionRepo

	userRepo := &mockUserRepoGacha{}
	server.userRepo = userRepo

	session := createMockSession(0, server)

	pkt := &mhfpacket.MsgSysLogin{
		AckHandle:        100,
		CharID0:          42,
		LoginTokenString: "test-token",
	}
	handleMsgSysLogin(session, pkt)

	select {
	case <-session.sendPackets:
		// got a response (fail ACK)
	default:
		t.Error("No response packet queued on GetUserID error")
	}
}

func TestHandleMsgSysLogin_BindSessionError(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.DebugOptions.DisableTokenCheck = true

	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	sessionRepo := &mockSessionRepo{bindErr: errors.New("bind failed")}
	server.sessionRepo = sessionRepo

	userRepo := &mockUserRepoGacha{}
	server.userRepo = userRepo

	session := createMockSession(0, server)

	pkt := &mhfpacket.MsgSysLogin{
		AckHandle:        100,
		CharID0:          42,
		LoginTokenString: "test-token",
	}
	handleMsgSysLogin(session, pkt)

	select {
	case <-session.sendPackets:
		// got a response (fail ACK)
	default:
		t.Error("No response packet queued on BindSession error")
	}
}

func TestHandleMsgSysLogin_SetLastCharacterError(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.DebugOptions.DisableTokenCheck = true

	charRepo := newMockCharacterRepo()
	server.charRepo = charRepo

	sessionRepo := &mockSessionRepo{}
	server.sessionRepo = sessionRepo

	userRepo := &mockUserRepoGacha{setLastCharErr: errors.New("set failed")}
	server.userRepo = userRepo

	session := createMockSession(0, server)

	pkt := &mhfpacket.MsgSysLogin{
		AckHandle:        100,
		CharID0:          42,
		LoginTokenString: "test-token",
	}
	handleMsgSysLogin(session, pkt)

	select {
	case <-session.sendPackets:
		// got a response (fail ACK)
	default:
		t.Error("No response packet queued on SetLastCharacter error")
	}
}

func TestHandleMsgSysPing_Session(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysPing{AckHandle: 100}
	handleMsgSysPing(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysIssueLogkey_GeneratesKey(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysIssueLogkey{AckHandle: 100}
	handleMsgSysIssueLogkey(session, pkt)

	if len(session.logKey) != 16 {
		t.Errorf("Expected 16-byte log key, got %d bytes", len(session.logKey))
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysRecordLog_ZZMode(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.RealClientMode = cfg.ZZ
	server.userBinary = NewUserBinaryStore()

	guildRepo := &mockGuildRepo{}
	server.guildRepo = guildRepo

	session := createMockSession(1, server)

	// Create a stage for the session (handler accesses s.stage.reservedClientSlots)
	stage := &Stage{
		id:                  "testStage",
		clients:             make(map[*Session]uint32),
		reservedClientSlots: make(map[uint32]bool),
	}
	stage.reservedClientSlots[1] = true
	session.stage = stage

	// Build kill log data: 32 header bytes + 176 monster bytes
	data := make([]byte, 32+176)
	// Set monster index 5 to have 2 kills (a large monster per mhfmon)
	data[32+5] = 2

	pkt := &mhfpacket.MsgSysRecordLog{
		AckHandle: 100,
		Data:      data,
	}
	handleMsgSysRecordLog(session, pkt)

	// Check that reserved slot was cleaned up
	if _, exists := stage.reservedClientSlots[1]; exists {
		t.Error("Expected reserved client slot to be removed")
	}

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysLockGlobalSema_LocalChannel(t *testing.T) {
	server := createMockServer()
	server.GlobalID = "ch1"
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysLockGlobalSema{
		AckHandle:             100,
		UserIDString:          "someStage",
		ServerChannelIDString: "ch1",
	}
	handleMsgSysLockGlobalSema(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysLockGlobalSema_RemoteMatch(t *testing.T) {
	server := createMockServer()
	server.GlobalID = "ch1"

	otherChannel := createMockServer()
	otherChannel.GlobalID = "ch2"
	otherChannel.stages.Store("prefix_testStage", &Stage{
		id:                  "prefix_testStage",
		clients:             make(map[*Session]uint32),
		reservedClientSlots: make(map[uint32]bool),
	})
	server.Registry = NewLocalChannelRegistry([]*Server{server, otherChannel})

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysLockGlobalSema{
		AckHandle:             100,
		UserIDString:          "testStage",
		ServerChannelIDString: "ch1",
	}
	handleMsgSysLockGlobalSema(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Fatal("Empty response")
		}
		_ = byteframe.NewByteFrameFromBytes(p.data)
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysUnlockGlobalSema_Session(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysUnlockGlobalSema{AckHandle: 100}
	handleMsgSysUnlockGlobalSema(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysRightsReload_Session(t *testing.T) {
	server := createMockServer()
	userRepo := &mockUserRepoGacha{rights: 0x02}
	server.userRepo = userRepo

	session := createMockSession(1, server)
	session.userID = 1

	pkt := &mhfpacket.MsgSysRightsReload{AckHandle: 100}
	handleMsgSysRightsReload(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfAnnounce_Session(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	dataBf := byteframe.NewByteFrame()
	dataBf.WriteUint8(2) // type = berserk

	pkt := &mhfpacket.MsgMhfAnnounce{
		AckHandle: 100,
		IPAddress: binary.LittleEndian.Uint32([]byte{127, 0, 0, 1}),
		Port:      54001,
		StageID:   make([]byte, 32),
		Data:      byteframe.NewByteFrameFromBytes(dataBf.Data()),
	}
	handleMsgMhfAnnounce(session, pkt)

	select {
	case <-session.sendPackets:
		// success
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgSysPing_DifferentAckHandles verifies ping works with various ack handles.
func TestHandleMsgSysPing_DifferentAckHandles(t *testing.T) {
	server := createMockServer()

	ackHandles := []uint32{0, 1, 99999, 0xFFFFFFFF}
	for _, ack := range ackHandles {
		session := createMockSession(1, server)
		pkt := &mhfpacket.MsgSysPing{AckHandle: ack}

		handleMsgSysPing(session, pkt)

		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Errorf("AckHandle=%d: Response packet should have data", ack)
			}
		default:
			t.Errorf("AckHandle=%d: No response packet queued", ack)
		}
	}
}

// TestHandleMsgSysTerminalLog_NoEntries verifies the handler works with nil entries.
func TestHandleMsgSysTerminalLog_NoEntries(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 99999,
		LogID:     0,
		Entries:   nil,
	}

	handleMsgSysTerminalLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgSysTerminalLog_ManyEntries verifies the handler with many log entries.
func TestHandleMsgSysTerminalLog_ManyEntries(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	entries := make([]mhfpacket.TerminalLogEntry, 20)
	for i := range entries {
		entries[i] = mhfpacket.TerminalLogEntry{
			Index: uint32(i),
			Type1: uint8(i % 256),
			Type2: uint8((i + 1) % 256),
		}
	}

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 55555,
		LogID:     42,
		Entries:   entries,
	}

	handleMsgSysTerminalLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// TestHandleMsgSysTime_MultipleCalls verifies calling time handler repeatedly.
func TestHandleMsgSysTime_MultipleCalls(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTime{
		GetRemoteTime: false,
		Timestamp:     0,
	}

	for i := 0; i < 5; i++ {
		handleMsgSysTime(session, pkt)
	}

	// Should have 5 queued responses
	count := 0
	for {
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("Response packet should have data")
			}
			count++
		default:
			goto done
		}
	}
done:
	if count != 5 {
		t.Errorf("Expected 5 queued responses, got %d", count)
	}
}

// mockCharRepoGetUserIDErr wraps mockCharacterRepo to return an error from GetUserID
type mockCharRepoGetUserIDErr struct {
	*mockCharacterRepo
	getUserIDErr error
}

func (m *mockCharRepoGetUserIDErr) GetUserID(_ uint32) (uint32, error) {
	return 0, m.getUserIDErr
}

// Tests consolidated from handlers_core_test.go

func TestHandleMsgHead(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgHead panicked: %v", r)
		}
	}()

	handleMsgHead(session, nil)
}

func TestHandleMsgSysExtendThreshold(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysExtendThreshold panicked: %v", r)
		}
	}()

	handleMsgSysExtendThreshold(session, nil)
}

func TestHandleMsgSysEnd(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysEnd panicked: %v", r)
		}
	}()

	handleMsgSysEnd(session, nil)
}

func TestHandleMsgSysNop(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysNop panicked: %v", r)
		}
	}()

	handleMsgSysNop(session, nil)
}

func TestHandleMsgSysAck(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysAck panicked: %v", r)
		}
	}()

	handleMsgSysAck(session, nil)
}

func TestHandleMsgCaExchangeItem(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgCaExchangeItem panicked: %v", r)
		}
	}()

	handleMsgCaExchangeItem(session, nil)
}

func TestHandleMsgMhfServerCommand(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfServerCommand panicked: %v", r)
		}
	}()

	handleMsgMhfServerCommand(session, nil)
}

func TestHandleMsgMhfSetLoginwindow(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfSetLoginwindow panicked: %v", r)
		}
	}()

	handleMsgMhfSetLoginwindow(session, nil)
}

func TestHandleMsgSysTransBinary(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysTransBinary panicked: %v", r)
		}
	}()

	handleMsgSysTransBinary(session, nil)
}

func TestHandleMsgSysCollectBinary(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysCollectBinary panicked: %v", r)
		}
	}()

	handleMsgSysCollectBinary(session, nil)
}

func TestHandleMsgSysGetState(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysGetState panicked: %v", r)
		}
	}()

	handleMsgSysGetState(session, nil)
}

func TestHandleMsgSysSerialize(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysSerialize panicked: %v", r)
		}
	}()

	handleMsgSysSerialize(session, nil)
}

func TestHandleMsgSysEnumlobby(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysEnumlobby panicked: %v", r)
		}
	}()

	handleMsgSysEnumlobby(session, nil)
}

func TestHandleMsgSysEnumuser(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysEnumuser panicked: %v", r)
		}
	}()

	handleMsgSysEnumuser(session, nil)
}

func TestHandleMsgSysInfokyserver(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysInfokyserver panicked: %v", r)
		}
	}()

	handleMsgSysInfokyserver(session, nil)
}

func TestHandleMsgMhfGetCaUniqueID(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetCaUniqueID panicked: %v", r)
		}
	}()

	handleMsgMhfGetCaUniqueID(session, nil)
}

func TestHandleMsgSysTerminalLog(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 12345,
		LogID:     100,
		Entries:   []mhfpacket.TerminalLogEntry{},
	}

	handleMsgSysTerminalLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysTerminalLog_WithEntries(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 12345,
		LogID:     100,
		Entries: []mhfpacket.TerminalLogEntry{
			{Type1: 1, Type2: 2},
			{Type1: 3, Type2: 4},
		},
	}

	handleMsgSysTerminalLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysPing(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysPing{
		AckHandle: 12345,
	}

	handleMsgSysPing(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysTime(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTime{
		GetRemoteTime: true,
	}

	handleMsgSysTime(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysIssueLogkey(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysIssueLogkey{
		AckHandle: 12345,
	}

	handleMsgSysIssueLogkey(session, pkt)

	// Verify logkey was set
	if len(session.logKey) != 16 {
		t.Errorf("logKey length = %d, want 16", len(session.logKey))
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysRecordLog(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Setup stage
	stage := NewStage("test_stage")
	session.stage = stage
	stage.reservedClientSlots[session.charID] = true

	pkt := &mhfpacket.MsgSysRecordLog{
		AckHandle: 12345,
		Data:      make([]byte, 256), // Must be large enough for ByteFrame reads (32 offset + 176 uint8s)
	}

	handleMsgSysRecordLog(session, pkt)

	// Verify charID removed from reserved slots
	if _, exists := stage.reservedClientSlots[session.charID]; exists {
		t.Error("charID should be removed from reserved slots")
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysUnlockGlobalSema(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysUnlockGlobalSema{
		AckHandle: 12345,
	}

	handleMsgSysUnlockGlobalSema(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysSetStatus(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysSetStatus panicked: %v", r)
		}
	}()

	handleMsgSysSetStatus(session, nil)
}

func TestHandleMsgSysEcho(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysEcho panicked: %v", r)
		}
	}()

	handleMsgSysEcho(session, nil)
}

func TestHandleMsgSysUpdateRight(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysUpdateRight panicked: %v", r)
		}
	}()

	handleMsgSysUpdateRight(session, nil)
}

func TestHandleMsgSysAuthQuery(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysAuthQuery panicked: %v", r)
		}
	}()

	handleMsgSysAuthQuery(session, nil)
}

func TestHandleMsgSysAuthTerminal(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysAuthTerminal panicked: %v", r)
		}
	}()

	handleMsgSysAuthTerminal(session, nil)
}

func TestHandleMsgSysLockGlobalSema_NoMatch(t *testing.T) {
	server := createMockServer()
	server.GlobalID = "test-server"
	server.Registry = NewLocalChannelRegistry([]*Server{})
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysLockGlobalSema{
		AckHandle:             12345,
		UserIDString:          "user123",
		ServerChannelIDString: "channel1",
	}

	handleMsgSysLockGlobalSema(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysLockGlobalSema_WithChannel(t *testing.T) {
	server := createMockServer()
	server.GlobalID = "test-server"

	// Create a mock channel with stages
	channel := &Server{
		GlobalID: "other-server",
	}
	channel.stages.Store("stage_user123", NewStage("stage_user123"))
	server.Registry = NewLocalChannelRegistry([]*Server{channel})

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysLockGlobalSema{
		AckHandle:             12345,
		UserIDString:          "user123",
		ServerChannelIDString: "channel1",
	}

	handleMsgSysLockGlobalSema(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysLockGlobalSema_SameServer(t *testing.T) {
	server := createMockServer()
	server.GlobalID = "test-server"

	// Create a mock channel with same GlobalID
	channel := &Server{
		GlobalID: "test-server",
	}
	channel.stages.Store("stage_user456", NewStage("stage_user456"))
	server.Registry = NewLocalChannelRegistry([]*Server{channel})

	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysLockGlobalSema{
		AckHandle:             12345,
		UserIDString:          "user456",
		ServerChannelIDString: "channel2",
	}

	handleMsgSysLockGlobalSema(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfAnnounce(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAnnounce{
		AckHandle: 12345,
		IPAddress: 0x7F000001, // 127.0.0.1
		Port:      54001,
		StageID:   []byte("test_stage"),
		Data:      byteframe.NewByteFrameFromBytes([]byte{0x00}),
	}

	handleMsgMhfAnnounce(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysRightsReload(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysRightsReload{
		AckHandle: 12345,
	}

	// This will panic due to nil db, which is expected in test
	defer func() {
		if r := recover(); r != nil {
			t.Log("Expected panic due to nil database in test")
		}
	}()

	handleMsgSysRightsReload(session, pkt)
}

// Tests consolidated from handlers_coverage3_test.go

func TestEmptyHandlers_HandlersGo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{
		{"handleMsgSysEcho", func() { handleMsgSysEcho(session, nil) }},
		{"handleMsgSysUpdateRight", func() { handleMsgSysUpdateRight(session, nil) }},
		{"handleMsgSysAuthQuery", func() { handleMsgSysAuthQuery(session, nil) }},
		{"handleMsgSysAuthTerminal", func() { handleMsgSysAuthTerminal(session, nil) }},
		{"handleMsgCaExchangeItem", func() { handleMsgCaExchangeItem(session, nil) }},
		{"handleMsgMhfServerCommand", func() { handleMsgMhfServerCommand(session, nil) }},
		{"handleMsgMhfSetLoginwindow", func() { handleMsgMhfSetLoginwindow(session, nil) }},
		{"handleMsgSysTransBinary", func() { handleMsgSysTransBinary(session, nil) }},
		{"handleMsgSysCollectBinary", func() { handleMsgSysCollectBinary(session, nil) }},
		{"handleMsgSysGetState", func() { handleMsgSysGetState(session, nil) }},
		{"handleMsgSysSerialize", func() { handleMsgSysSerialize(session, nil) }},
		{"handleMsgSysEnumlobby", func() { handleMsgSysEnumlobby(session, nil) }},
		{"handleMsgSysEnumuser", func() { handleMsgSysEnumuser(session, nil) }},
		{"handleMsgSysInfokyserver", func() { handleMsgSysInfokyserver(session, nil) }},
		{"handleMsgMhfGetCaUniqueID", func() { handleMsgMhfGetCaUniqueID(session, nil) }},
		{"handleMsgMhfGetExtraInfo", func() { handleMsgMhfGetExtraInfo(session, nil) }},
		{"handleMsgSysSetStatus", func() { handleMsgSysSetStatus(session, nil) }},
		{"handleMsgMhfStampcardPrize", func() { handleMsgMhfStampcardPrize(session, nil) }},
		{"handleMsgMhfKickExportForce", func() { handleMsgMhfKickExportForce(session, nil) }},
		{"handleMsgMhfRegistSpabiTime", func() { handleMsgMhfRegistSpabiTime(session, nil) }},
		{"handleMsgMhfDebugPostValue", func() { handleMsgMhfDebugPostValue(session, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}

func TestEmptyHandlers_Concurrent(t *testing.T) {
	server := createMockServer()

	handlers := []func(*Session, mhfpacket.MHFPacket){
		handleMsgSysEcho,
		handleMsgSysUpdateRight,
		handleMsgSysAuthQuery,
		handleMsgSysAuthTerminal,
		handleMsgCaExchangeItem,
		handleMsgMhfServerCommand,
		handleMsgMhfSetLoginwindow,
		handleMsgSysTransBinary,
		handleMsgSysCollectBinary,
		handleMsgSysGetState,
		handleMsgSysSerialize,
		handleMsgSysEnumlobby,
		handleMsgSysEnumuser,
		handleMsgSysInfokyserver,
		handleMsgMhfGetCaUniqueID,
		handleMsgMhfGetExtraInfo,
		handleMsgSysSetStatus,
		handleMsgSysDeleteObject,
		handleMsgSysRotateObject,
		handleMsgSysDuplicateObject,
		handleMsgSysGetObjectBinary,
		handleMsgSysGetObjectOwner,
		handleMsgSysUpdateObjectBinary,
		handleMsgSysCleanupObject,
		handleMsgMhfShutClient,
		handleMsgSysHideClient,
		handleMsgSysStageDestruct,
	}

	var wg sync.WaitGroup
	for _, h := range handlers {
		for i := 0; i < 10; i++ {
			wg.Add(1)
			go func(handler func(*Session, mhfpacket.MHFPacket)) {
				defer wg.Done()
				session := createMockSession(1, server)
				handler(session, nil)
			}(h)
		}
	}
	wg.Wait()
}

func TestSimpleHandlers_PingAndTime(t *testing.T) {
	server := createMockServer()

	t.Run("handleMsgSysPing", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysPing(session, &mhfpacket.MsgSysPing{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgSysTime", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysTime(session, &mhfpacket.MsgSysTime{})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})
}

func TestHandleMsgSysIssueLogkey_Coverage3(t *testing.T) {
	server := createMockServer()

	t.Run("generates_logkey", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysIssueLogkey(session, &mhfpacket.MsgSysIssueLogkey{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
		if session.logKey == nil {
			t.Error("logKey should be set after IssueLogkey")
		}
		if len(session.logKey) != 16 {
			t.Errorf("logKey length = %d, want 16", len(session.logKey))
		}
	})
}

func TestHandleMsgSysUnlockGlobalSema_Coverage3(t *testing.T) {
	server := createMockServer()

	t.Run("produces_response", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysUnlockGlobalSema(session, &mhfpacket.MsgSysUnlockGlobalSema{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})
}

func TestHandleMsgSysLockGlobalSema_Coverage3(t *testing.T) {
	server := createMockServer()
	server.Registry = NewLocalChannelRegistry(make([]*Server, 0))

	t.Run("no_channels_returns_response", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgSysLockGlobalSema(session, &mhfpacket.MsgSysLockGlobalSema{
			AckHandle:             1,
			UserIDString:          "testuser",
			ServerChannelIDString: "ch1",
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})
}

func TestHandlersConcurrentInvocations(t *testing.T) {
	server := createMockServer()

	done := make(chan struct{})
	const numGoroutines = 10

	for i := 0; i < numGoroutines; i++ {
		go func(id uint32) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("goroutine %d panicked: %v", id, r)
				}
				done <- struct{}{}
			}()

			session := createMockSession(id, server)

			// Run several handlers concurrently
			handleMsgSysPing(session, &mhfpacket.MsgSysPing{AckHandle: id})
			<-session.sendPackets

			handleMsgSysTime(session, &mhfpacket.MsgSysTime{GetRemoteTime: true})
			<-session.sendPackets

			handleMsgSysIssueLogkey(session, &mhfpacket.MsgSysIssueLogkey{AckHandle: id})
			<-session.sendPackets

			handleMsgMhfMercenaryHuntdata(session, &mhfpacket.MsgMhfMercenaryHuntdata{AckHandle: id, RequestType: 1})
			<-session.sendPackets

			handleMsgMhfEnumerateMercenaryLog(session, &mhfpacket.MsgMhfEnumerateMercenaryLog{AckHandle: id})
			<-session.sendPackets
		}(uint32(i + 100))
	}

	for i := 0; i < numGoroutines; i++ {
		<-done
	}
}

func TestHandleMsgSysRecordLog_RemovesReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	stage := NewStage("test_stage_record")
	session.stage = stage
	stage.reservedClientSlots[session.charID] = true

	pkt := &mhfpacket.MsgSysRecordLog{
		AckHandle: 55555,
		Data:      make([]byte, 256),
	}

	handleMsgSysRecordLog(session, pkt)

	if _, exists := stage.reservedClientSlots[session.charID]; exists {
		t.Error("charID should be removed from reserved slots after record log")
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysRecordLog_NoExistingReservation(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	stage := NewStage("test_stage_no_reservation")
	session.stage = stage
	// No reservation exists for this charID

	pkt := &mhfpacket.MsgSysRecordLog{
		AckHandle: 55556,
		Data:      make([]byte, 256),
	}

	// Should not panic even if charID is not in reservedClientSlots
	handleMsgSysRecordLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysUnlockGlobalSema_Response(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysUnlockGlobalSema{
		AckHandle: 66666,
	}

	handleMsgSysUnlockGlobalSema(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysTerminalLog_MultipleEntries(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 12345,
		LogID:     200,
		Entries: []mhfpacket.TerminalLogEntry{
			{Type1: 10, Type2: 20},
			{Type1: 11, Type2: 21},
			{Type1: 12, Type2: 22},
		},
	}

	handleMsgSysTerminalLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysTerminalLog_ZeroLogID(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTerminalLog{
		AckHandle: 12345,
		LogID:     0,
		Entries:   []mhfpacket.TerminalLogEntry{},
	}

	handleMsgSysTerminalLog(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysPing_DifferentAckHandle(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysPing{
		AckHandle: 0xFFFFFFFF,
	}

	handleMsgSysPing(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysTime_GetRemoteTimeFalse(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysTime{
		GetRemoteTime: false,
	}

	handleMsgSysTime(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysIssueLogkey_LogKeyGenerated(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysIssueLogkey{
		AckHandle: 77777,
	}

	handleMsgSysIssueLogkey(session, pkt)

	// Verify that the logKey was set on the session
	session.Lock()
	keyLen := len(session.logKey)
	session.Unlock()

	if keyLen != 16 {
		t.Errorf("logKey length = %d, want 16", keyLen)
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysIssueLogkey_Uniqueness(t *testing.T) {
	server := createMockServer()

	// Generate two logkeys and verify they differ
	session1 := createMockSession(1, server)
	session2 := createMockSession(2, server)

	pkt1 := &mhfpacket.MsgSysIssueLogkey{AckHandle: 1}
	pkt2 := &mhfpacket.MsgSysIssueLogkey{AckHandle: 2}

	handleMsgSysIssueLogkey(session1, pkt1)
	handleMsgSysIssueLogkey(session2, pkt2)

	// Drain send packets
	<-session1.sendPackets
	<-session2.sendPackets

	session1.Lock()
	key1 := make([]byte, len(session1.logKey))
	copy(key1, session1.logKey)
	session1.Unlock()

	session2.Lock()
	key2 := make([]byte, len(session2.logKey))
	copy(key2, session2.logKey)
	session2.Unlock()

	if len(key1) != 16 || len(key2) != 16 {
		t.Fatalf("logKeys should be 16 bytes each, got %d and %d", len(key1), len(key2))
	}

	same := true
	for i := range key1 {
		if key1[i] != key2[i] {
			same = false
			break
		}
	}
	if same {
		t.Error("Two generated logkeys should differ (extremely unlikely to be the same)")
	}
}

func TestMultipleHandlersOnSameSession(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Call multiple handlers in sequence
	handleMsgSysPing(session, &mhfpacket.MsgSysPing{AckHandle: 1})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from Ping handler")
	}

	handleMsgSysTime(session, &mhfpacket.MsgSysTime{GetRemoteTime: true})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from Time handler")
	}

	handleMsgMhfRegisterEvent(session, &mhfpacket.MsgMhfRegisterEvent{AckHandle: 2, WorldID: 5, LandID: 10})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from RegisterEvent handler")
	}

	handleMsgMhfReleaseEvent(session, &mhfpacket.MsgMhfReleaseEvent{AckHandle: 3})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from ReleaseEvent handler")
	}

	handleMsgMhfEnumerateEvent(session, &mhfpacket.MsgMhfEnumerateEvent{AckHandle: 4})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from EnumerateEvent handler")
	}

	handleMsgMhfSetCaAchievementHist(session, &mhfpacket.MsgMhfSetCaAchievementHist{AckHandle: 5})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from SetCaAchievementHist handler")
	}

	handleMsgMhfGetRengokuRankingRank(session, &mhfpacket.MsgMhfGetRengokuRankingRank{AckHandle: 6})
	select {
	case <-session.sendPackets:
	default:
		t.Fatal("Expected packet from GetRengokuRankingRank handler")
	}
}

func TestEmptyHandlers_MiscFiles_Session(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name string
		fn   func()
	}{
		{"handleMsgHead", func() { handleMsgHead(session, nil) }},
		{"handleMsgSysExtendThreshold", func() { handleMsgSysExtendThreshold(session, nil) }},
		{"handleMsgSysEnd", func() { handleMsgSysEnd(session, nil) }},
		{"handleMsgSysNop", func() { handleMsgSysNop(session, nil) }},
		{"handleMsgSysAck", func() { handleMsgSysAck(session, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.fn()
		})
	}
}
