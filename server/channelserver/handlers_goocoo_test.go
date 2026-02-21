package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfEnumerateGuacot_Empty(t *testing.T) {
	server := createMockServer()
	mock := newMockGoocooRepo()
	server.goocooRepo = mock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuacot{AckHandle: 100}

	handleMsgMhfEnumerateGuacot(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfEnumerateGuacot_WithSlots(t *testing.T) {
	server := createMockServer()
	mock := newMockGoocooRepo()
	mock.slots[0] = []byte{0x01, 0x02, 0x03, 0x04} // slot 0 has data
	mock.slots[2] = []byte{0x05, 0x06, 0x07, 0x08} // slot 2 has data
	server.goocooRepo = mock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuacot{AckHandle: 100}

	handleMsgMhfEnumerateGuacot(session, pkt)

	select {
	case p := <-session.sendPackets:
		// Header (4 bytes) + 2 goocoo entries
		if len(p.data) < 8 {
			t.Errorf("Response too short: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUpdateGuacot_ClearSlot(t *testing.T) {
	server := createMockServer()
	mock := newMockGoocooRepo()
	mock.slots[1] = []byte{0x01, 0x02} // pre-existing data
	server.goocooRepo = mock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateGuacot{
		AckHandle: 100,
		Goocoos: []mhfpacket.Goocoo{
			{
				Index: 1,
				Data1: []int16{0, 0, 0}, // First byte 0 = clear
				Data2: []uint32{0},
				Name:  []byte("test"),
			},
		},
	}

	handleMsgMhfUpdateGuacot(session, pkt)

	if len(mock.clearCalled) != 1 || mock.clearCalled[0] != 1 {
		t.Errorf("Expected ClearSlot(1), got %v", mock.clearCalled)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUpdateGuacot_SaveSlot(t *testing.T) {
	server := createMockServer()
	mock := newMockGoocooRepo()
	server.goocooRepo = mock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateGuacot{
		AckHandle: 100,
		Goocoos: []mhfpacket.Goocoo{
			{
				Index: 2,
				Data1: []int16{1, 2, 3}, // First byte non-zero = save
				Data2: []uint32{100, 200},
				Name:  []byte("MyGoocoo"),
			},
		},
	}

	handleMsgMhfUpdateGuacot(session, pkt)

	if _, ok := mock.savedSlots[2]; !ok {
		t.Error("Expected SaveSlot to be called for slot 2")
	}
	if len(mock.clearCalled) != 0 {
		t.Error("ClearSlot should not be called for a save operation")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfUpdateGuacot_SkipInvalidIndex(t *testing.T) {
	server := createMockServer()
	mock := newMockGoocooRepo()
	server.goocooRepo = mock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateGuacot{
		AckHandle: 100,
		Goocoos: []mhfpacket.Goocoo{
			{
				Index: 5, // > 4, should be skipped
				Data1: []int16{1},
				Data2: []uint32{0},
				Name:  []byte("Bad"),
			},
		},
	}

	handleMsgMhfUpdateGuacot(session, pkt)

	if len(mock.savedSlots) != 0 {
		t.Error("SaveSlot should not be called for index > 4")
	}
	if len(mock.clearCalled) != 0 {
		t.Error("ClearSlot should not be called for index > 4")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}
