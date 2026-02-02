package channelserver

import (
	"testing"
	"time"

	"erupe-ce/network/mhfpacket"
)

// Test simple handler patterns that don't require database

func TestHandlerMsgMhfSexChanger(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSexChanger{
		AckHandle: 12345,
	}

	// Should not panic
	handleMsgMhfSexChanger(session, pkt)

	// Should queue a response
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandlerMsgMhfEnterTournamentQuest(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic with nil packet (empty handler)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("handleMsgMhfEnterTournamentQuest panicked: %v", r)
		}
	}()

	handleMsgMhfEnterTournamentQuest(session, nil)
}

func TestHandlerMsgMhfGetUdBonusQuestInfo(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetUdBonusQuestInfo{
		AckHandle: 12345,
	}

	handleMsgMhfGetUdBonusQuestInfo(session, pkt)

	// Should queue a response
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

// Test that acknowledge handlers work correctly

func TestAckResponseFormats(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name    string
		handler func(s *Session, ackHandle uint32, data []byte)
	}{
		{"doAckBufSucceed", doAckBufSucceed},
		{"doAckBufFail", doAckBufFail},
		{"doAckSimpleSucceed", doAckSimpleSucceed},
		{"doAckSimpleFail", doAckSimpleFail},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)
			testData := []byte{0x01, 0x02, 0x03, 0x04}

			tt.handler(session, 99999, testData)

			select {
			case pkt := <-session.sendPackets:
				if pkt.data == nil {
					t.Error("Packet data should not be nil")
				}
			default:
				t.Error("Handler should queue a packet")
			}
		})
	}
}

func TestStubHandlers(t *testing.T) {
	server := createMockServer()

	tests := []struct {
		name    string
		handler func(s *Session, ackHandle uint32)
	}{
		{"stubEnumerateNoResults", stubEnumerateNoResults},
		{"stubGetNoResults", stubGetNoResults},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			session := createMockSession(1, server)

			tt.handler(session, 12345)

			select {
			case pkt := <-session.sendPackets:
				if pkt.data == nil {
					t.Error("Packet data should not be nil")
				}
			default:
				t.Error("Stub handler should queue a packet")
			}
		})
	}
}

// Test packet queueing

func TestSessionQueueSendMHF(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgSysAck{
		AckHandle:        12345,
		IsBufferResponse: false,
		ErrorCode:        0,
		AckData:          []byte{0x00},
	}

	session.QueueSendMHF(pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Queued packet should have data")
		}
	default:
		t.Error("QueueSendMHF should queue a packet")
	}
}

func TestSessionQueueSendNonBlocking(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	data := []byte{0x01, 0x02, 0x03, 0x04}
	session.QueueSendNonBlocking(data)

	select {
	case p := <-session.sendPackets:
		if len(p.data) != 4 {
			t.Errorf("Queued data len = %d, want 4", len(p.data))
		}
		if p.nonBlocking != true {
			t.Error("Packet should be marked as non-blocking")
		}
	default:
		t.Error("QueueSendNonBlocking should queue data")
	}
}

func TestSessionQueueSendNonBlocking_FullQueue(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Fill the queue
	for i := 0; i < 20; i++ {
		session.sendPackets <- packet{data: []byte{byte(i)}, nonBlocking: true}
	}

	// Non-blocking send should not block when queue is full
	// It should drop the packet instead
	done := make(chan bool, 1)
	go func() {
		session.QueueSendNonBlocking([]byte{0xFF})
		done <- true
	}()

	// Wait for completion with a reasonable timeout
	// The function should return immediately (dropping the packet)
	select {
	case <-done:
		// Good - didn't block, function completed
	case <-time.After(100 * time.Millisecond):
		t.Error("QueueSendNonBlocking blocked on full queue")
	}
}
