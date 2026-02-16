package channelserver

import (
	"testing"
)

func TestStubEnumerateNoResults(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Call stubEnumerateNoResults - it queues a packet
	stubEnumerateNoResults(session, 12345)

	// Verify packet was queued
	select {
	case pkt := <-session.sendPackets:
		if len(pkt.data) == 0 {
			t.Error("Packet data should not be empty")
		}
	default:
		t.Error("No packet was queued")
	}
}

func TestDoAckBufSucceed(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	testData := []byte{0x01, 0x02, 0x03, 0x04}
	doAckBufSucceed(session, 12345, testData)

	// Verify packet was queued
	select {
	case pkt := <-session.sendPackets:
		if len(pkt.data) == 0 {
			t.Error("Packet data should not be empty")
		}
	default:
		t.Error("No packet was queued")
	}
}

func TestDoAckBufFail(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	testData := []byte{0x01, 0x02, 0x03, 0x04}
	doAckBufFail(session, 12345, testData)

	// Verify packet was queued
	select {
	case pkt := <-session.sendPackets:
		if len(pkt.data) == 0 {
			t.Error("Packet data should not be empty")
		}
	default:
		t.Error("No packet was queued")
	}
}

func TestDoAckSimpleSucceed(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	testData := []byte{0x00, 0x00, 0x00, 0x00}
	doAckSimpleSucceed(session, 12345, testData)

	// Verify packet was queued
	select {
	case pkt := <-session.sendPackets:
		if len(pkt.data) == 0 {
			t.Error("Packet data should not be empty")
		}
	default:
		t.Error("No packet was queued")
	}
}

func TestDoAckSimpleFail(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	testData := []byte{0x00, 0x00, 0x00, 0x00}
	doAckSimpleFail(session, 12345, testData)

	// Verify packet was queued
	select {
	case pkt := <-session.sendPackets:
		if len(pkt.data) == 0 {
			t.Error("Packet data should not be empty")
		}
	default:
		t.Error("No packet was queued")
	}
}

func TestDoAck_EmptyData(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should work with empty data
	doAckBufSucceed(session, 0, []byte{})

	select {
	case pkt := <-session.sendPackets:
		// Empty data is valid
		_ = pkt
	default:
		t.Error("No packet was queued with empty data")
	}
}

func TestDoAck_NilData(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should work with nil data
	doAckBufSucceed(session, 0, nil)

	select {
	case pkt := <-session.sendPackets:
		// Nil data is valid
		_ = pkt
	default:
		t.Error("No packet was queued with nil data")
	}
}

func TestDoAck_LargeData(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Test with large data
	largeData := make([]byte, 65536)
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	doAckBufSucceed(session, 99999, largeData)

	select {
	case pkt := <-session.sendPackets:
		if len(pkt.data) == 0 {
			t.Error("Packet data should not be empty for large data")
		}
	default:
		t.Error("No packet was queued with large data")
	}
}

func TestDoAck_AckHandleZero(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Test with ack handle 0
	doAckSimpleSucceed(session, 0, []byte{0x00})

	select {
	case pkt := <-session.sendPackets:
		_ = pkt
	default:
		t.Error("No packet was queued with zero ack handle")
	}
}

func TestDoAck_AckHandleMax(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Test with max uint32 ack handle
	doAckSimpleSucceed(session, 0xFFFFFFFF, []byte{0x00})

	select {
	case pkt := <-session.sendPackets:
		_ = pkt
	default:
		t.Error("No packet was queued with max ack handle")
	}
}

// Test that handlers don't panic with empty packets
func TestEmptyHandlers(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name    string
		handler func(s *Session, p interface{})
	}{
		{"handleMsgHead", func(s *Session, p interface{}) { handleMsgHead(s, nil) }},
		{"handleMsgSysExtendThreshold", func(s *Session, p interface{}) { handleMsgSysExtendThreshold(s, nil) }},
		{"handleMsgSysEnd", func(s *Session, p interface{}) { handleMsgSysEnd(s, nil) }},
		{"handleMsgSysNop", func(s *Session, p interface{}) { handleMsgSysNop(s, nil) }},
		{"handleMsgSysAck", func(s *Session, p interface{}) { handleMsgSysAck(s, nil) }},
		{"handleMsgSysAuthData", func(s *Session, p interface{}) { handleMsgSysAuthData(s, nil) }},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.handler(session, nil)
		})
	}
}
