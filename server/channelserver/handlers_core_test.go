package channelserver

import (
	"testing"

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

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfEnumerateItem panicked: %v", r)
		}
	}()

	handleMsgMhfEnumerateItem(session, nil)
}

func TestHandleMsgMhfAcquireItem(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfAcquireItem panicked: %v", r)
		}
	}()

	handleMsgMhfAcquireItem(session, nil)
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
		Entries:   []*mhfpacket.TerminalLogEntry{},
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
		Entries: []*mhfpacket.TerminalLogEntry{
			{Type1: 1, Type2: 2, Data: []int16{1, 2, 3}},
			{Type1: 3, Type2: 4, Data: []int16{4, 5, 6}},
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
