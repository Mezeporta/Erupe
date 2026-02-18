package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgSysInsertUser(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysInsertUser panicked: %v", r)
		}
	}()

	handleMsgSysInsertUser(session, nil)
}

func TestHandleMsgSysDeleteUser(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysDeleteUser panicked: %v", r)
		}
	}()

	handleMsgSysDeleteUser(session, nil)
}

func TestHandleMsgSysNotifyUserBinary(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgSysNotifyUserBinary panicked: %v", r)
		}
	}()

	handleMsgSysNotifyUserBinary(session, nil)
}

func TestHandleMsgSysGetUserBinary_FromCache(t *testing.T) {
	server := createMockServer()
	server.userBinaryParts = make(map[userBinaryPartID][]byte)
	session := createMockSession(1, server)

	// Pre-populate cache
	key := userBinaryPartID{charID: 100, index: 1}
	server.userBinaryParts[key] = []byte{0x01, 0x02, 0x03, 0x04}

	pkt := &mhfpacket.MsgSysGetUserBinary{
		AckHandle:  12345,
		CharID:     100,
		BinaryType: 1,
	}

	handleMsgSysGetUserBinary(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgSysGetUserBinary_NotInCache(t *testing.T) {
	server := createMockServer()
	server.userBinaryParts = make(map[userBinaryPartID][]byte)
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysGetUserBinary{
		AckHandle:  12345,
		CharID:     100,
		BinaryType: 1,
	}

	handleMsgSysGetUserBinary(session, pkt)

	// Should return a fail ACK (no DB fallback, just cache miss)
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestUserBinaryPartID_AsMapKey(t *testing.T) {
	// Test that userBinaryPartID works as map key
	parts := make(map[userBinaryPartID][]byte)

	key1 := userBinaryPartID{charID: 1, index: 0}
	key2 := userBinaryPartID{charID: 1, index: 1}
	key3 := userBinaryPartID{charID: 2, index: 0}

	parts[key1] = []byte{0x01}
	parts[key2] = []byte{0x02}
	parts[key3] = []byte{0x03}

	if len(parts) != 3 {
		t.Errorf("Expected 3 parts, got %d", len(parts))
	}

	if parts[key1][0] != 0x01 {
		t.Error("Key1 data mismatch")
	}
	if parts[key2][0] != 0x02 {
		t.Error("Key2 data mismatch")
	}
	if parts[key3][0] != 0x03 {
		t.Error("Key3 data mismatch")
	}
}
