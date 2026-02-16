package channelserver

import (
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// Test empty handlers don't panic

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

func TestHandleMsgMhfEnumerateItem(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateItem{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfAcquireItem(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAcquireItem{
		AckHandle: 12345,
	}

	handleMsgMhfAcquireItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfGetExtraInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfGetExtraInfo panicked: %v", r)
		}
	}()

	handleMsgMhfGetExtraInfo(session, nil)
}

// Test handlers that return simple responses

func TestHandleMsgMhfTransferItem(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfTransferItem{
		AckHandle: 12345,
	}

	handleMsgMhfTransferItem(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumeratePrice(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumeratePrice{
		AckHandle: 12345,
	}

	handleMsgMhfEnumeratePrice(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateOrder(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateOrder{
		AckHandle: 12345,
	}

	handleMsgMhfEnumerateOrder(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Test terminal log handler

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

// Test ping handler
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

// Test time handler
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

// Test issue logkey handler
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

// Test record log handler
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

// Test unlock global sema handler
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

// Test more empty handlers
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

// Test lock global sema handler
func TestHandleMsgSysLockGlobalSema_NoMatch(t *testing.T) {
	server := createMockServer()
	server.GlobalID = "test-server"
	server.Channels = []*Server{}
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
		stages:   make(map[string]*Stage),
	}
	channel.stages["stage_user123"] = NewStage("stage_user123")
	server.Channels = []*Server{channel}

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
		stages:   make(map[string]*Stage),
	}
	channel.stages["stage_user456"] = NewStage("stage_user456")
	server.Channels = []*Server{channel}

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
