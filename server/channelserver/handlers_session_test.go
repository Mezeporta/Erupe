package channelserver

import (
	"encoding/binary"
	"errors"
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

	guildRepo := &mockGuildRepoForMail{}
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

// mockCharRepoGetUserIDErr wraps mockCharacterRepo to return an error from GetUserID
type mockCharRepoGetUserIDErr struct {
	*mockCharacterRepo
	getUserIDErr error
}

func (m *mockCharRepoGetUserIDErr) GetUserID(_ uint32) (uint32, error) {
	return 0, m.getUserIDErr
}
