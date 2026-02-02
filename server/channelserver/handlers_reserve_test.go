package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

// Test that reserve handlers with AckHandle respond correctly

func TestHandleMsgSysReserve188(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysReserve188{
		AckHandle: 12345,
	}

	handleMsgSysReserve188(session, pkt)

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

func TestHandleMsgSysReserve18B(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysReserve18B{
		AckHandle: 12345,
	}

	handleMsgSysReserve18B(session, pkt)

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

// Test that empty reserve handlers don't panic

func TestEmptyReserveHandlers(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name    string
		handler func(s *Session, p mhfpacket.MHFPacket)
	}{
		{"handleMsgSysReserve55", handleMsgSysReserve55},
		{"handleMsgSysReserve56", handleMsgSysReserve56},
		{"handleMsgSysReserve57", handleMsgSysReserve57},
		{"handleMsgSysReserve01", handleMsgSysReserve01},
		{"handleMsgSysReserve02", handleMsgSysReserve02},
		{"handleMsgSysReserve03", handleMsgSysReserve03},
		{"handleMsgSysReserve04", handleMsgSysReserve04},
		{"handleMsgSysReserve05", handleMsgSysReserve05},
		{"handleMsgSysReserve06", handleMsgSysReserve06},
		{"handleMsgSysReserve07", handleMsgSysReserve07},
		{"handleMsgSysReserve0C", handleMsgSysReserve0C},
		{"handleMsgSysReserve0D", handleMsgSysReserve0D},
		{"handleMsgSysReserve0E", handleMsgSysReserve0E},
		{"handleMsgSysReserve4A", handleMsgSysReserve4A},
		{"handleMsgSysReserve4B", handleMsgSysReserve4B},
		{"handleMsgSysReserve4C", handleMsgSysReserve4C},
		{"handleMsgSysReserve4D", handleMsgSysReserve4D},
		{"handleMsgSysReserve4E", handleMsgSysReserve4E},
		{"handleMsgSysReserve4F", handleMsgSysReserve4F},
		{"handleMsgSysReserve5C", handleMsgSysReserve5C},
		{"handleMsgSysReserve5E", handleMsgSysReserve5E},
		{"handleMsgSysReserve5F", handleMsgSysReserve5F},
		{"handleMsgSysReserve71", handleMsgSysReserve71},
		{"handleMsgSysReserve72", handleMsgSysReserve72},
		{"handleMsgSysReserve73", handleMsgSysReserve73},
		{"handleMsgSysReserve74", handleMsgSysReserve74},
		{"handleMsgSysReserve75", handleMsgSysReserve75},
		{"handleMsgSysReserve76", handleMsgSysReserve76},
		{"handleMsgSysReserve77", handleMsgSysReserve77},
		{"handleMsgSysReserve78", handleMsgSysReserve78},
		{"handleMsgSysReserve79", handleMsgSysReserve79},
		{"handleMsgSysReserve7A", handleMsgSysReserve7A},
		{"handleMsgSysReserve7B", handleMsgSysReserve7B},
		{"handleMsgSysReserve7C", handleMsgSysReserve7C},
		{"handleMsgSysReserve7E", handleMsgSysReserve7E},
		{"handleMsgMhfReserve10F", handleMsgMhfReserve10F},
		{"handleMsgSysReserve180", handleMsgSysReserve180},
		{"handleMsgSysReserve18E", handleMsgSysReserve18E},
		{"handleMsgSysReserve18F", handleMsgSysReserve18F},
		{"handleMsgSysReserve19E", handleMsgSysReserve19E},
		{"handleMsgSysReserve19F", handleMsgSysReserve19F},
		{"handleMsgSysReserve1A4", handleMsgSysReserve1A4},
		{"handleMsgSysReserve1A6", handleMsgSysReserve1A6},
		{"handleMsgSysReserve1A7", handleMsgSysReserve1A7},
		{"handleMsgSysReserve1A8", handleMsgSysReserve1A8},
		{"handleMsgSysReserve1A9", handleMsgSysReserve1A9},
		{"handleMsgSysReserve1AA", handleMsgSysReserve1AA},
		{"handleMsgSysReserve1AB", handleMsgSysReserve1AB},
		{"handleMsgSysReserve1AC", handleMsgSysReserve1AC},
		{"handleMsgSysReserve1AD", handleMsgSysReserve1AD},
		{"handleMsgSysReserve1AE", handleMsgSysReserve1AE},
		{"handleMsgSysReserve1AF", handleMsgSysReserve1AF},
		{"handleMsgSysReserve19B", handleMsgSysReserve19B},
		{"handleMsgSysReserve192", handleMsgSysReserve192},
		{"handleMsgSysReserve193", handleMsgSysReserve193},
		{"handleMsgSysReserve194", handleMsgSysReserve194},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()

			// Call with nil packet - empty handlers should handle this
			tt.handler(session, nil)
		})
	}
}

// Test reserve handlers are registered in handler table

func TestReserveHandlersRegistered(t *testing.T) {
	if handlerTable == nil {
		t.Fatal("handlerTable should be initialized")
	}

	// Check that reserve handlers exist in the table
	reserveHandlerCount := 0
	for _, handler := range handlerTable {
		if handler != nil {
			reserveHandlerCount++
		}
	}

	if reserveHandlerCount < 50 {
		t.Errorf("Expected at least 50 handlers registered, got %d", reserveHandlerCount)
	}
}
