package channelserver

import (
	"bytes"
	"encoding/binary"
	"io"
	
	cfg "erupe-ce/config"
	"erupe-ce/network"
	"sync"
	"testing"
	"time"

	"go.uber.org/zap"
)

// MockCryptConn simulates the encrypted connection for testing
type MockCryptConn struct {
	sentPackets [][]byte
	mu          sync.Mutex
}

func (m *MockCryptConn) SendPacket(data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	// Make a copy to avoid race conditions
	packetCopy := make([]byte, len(data))
	copy(packetCopy, data)
	m.sentPackets = append(m.sentPackets, packetCopy)
	return nil
}

func (m *MockCryptConn) ReadPacket() ([]byte, error) {
	// Return EOF to simulate graceful disconnect
	// This makes recvLoop() exit and call logoutPlayer()
	return nil, io.EOF
}

func (m *MockCryptConn) GetSentPackets() [][]byte {
	m.mu.Lock()
	defer m.mu.Unlock()
	packets := make([][]byte, len(m.sentPackets))
	copy(packets, m.sentPackets)
	return packets
}

func (m *MockCryptConn) PacketCount() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.sentPackets)
}

// createTestSession creates a properly initialized session for testing
func createTestSession(mock network.Conn) *Session {
	// Create a production logger for testing (will output to stderr)
	logger, _ := zap.NewProduction()

	server := &Server{
		erupeConfig: &cfg.Config{
			DebugOptions: cfg.DebugOptions{
				LogOutboundMessages: false,
			},
		},
	}
	server.Registry = NewLocalChannelRegistry([]*Server{server})
	s := &Session{
		logger:      logger,
		sendPackets: make(chan packet, 20),
		cryptConn:   mock,
		server:      server,
	}
	return s
}

// TestPacketQueueIndividualSending verifies that packets are sent individually
// with their own terminators instead of being concatenated
func TestPacketQueueIndividualSending(t *testing.T) {
	tests := []struct {
		name           string
		packetCount    int
		wantPackets    int
		wantTerminators int
	}{
		{
			name:           "single_packet",
			packetCount:    1,
			wantPackets:    1,
			wantTerminators: 1,
		},
		{
			name:           "multiple_packets",
			packetCount:    5,
			wantPackets:    5,
			wantTerminators: 5,
		},
		{
			name:           "many_packets",
			packetCount:    20,
			wantPackets:    20,
			wantTerminators: 20,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)

			// Start the send loop in a goroutine
			go s.sendLoop()

			// Queue multiple packets
			for i := 0; i < tt.packetCount; i++ {
				testData := []byte{0x00, byte(i), 0xAA, 0xBB}
				s.sendPackets <- packet{testData, true}
			}

			// Wait for packets to be processed
			time.Sleep(100 * time.Millisecond)

			// Stop the session
			s.closed.Store(true)
			time.Sleep(50 * time.Millisecond)

			// Verify packet count
			sentPackets := mock.GetSentPackets()
			if len(sentPackets) != tt.wantPackets {
				t.Errorf("got %d packets, want %d", len(sentPackets), tt.wantPackets)
			}

			// Verify each packet has its own terminator (0x00 0x10)
			terminatorCount := 0
			for _, pkt := range sentPackets {
				if len(pkt) < 2 {
					t.Errorf("packet too short: %d bytes", len(pkt))
					continue
				}
				// Check for terminator at the end
				if pkt[len(pkt)-2] == 0x00 && pkt[len(pkt)-1] == 0x10 {
					terminatorCount++
				}
			}

			if terminatorCount != tt.wantTerminators {
				t.Errorf("got %d terminators, want %d", terminatorCount, tt.wantTerminators)
			}
		})
	}
}

// TestPacketQueueNoConcatenation verifies that packets are NOT concatenated
// This test specifically checks the bug that was fixed
func TestPacketQueueNoConcatenation(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)

	go s.sendLoop()

	// Send 3 different packets with distinct data
	packet1 := []byte{0x00, 0x01, 0xAA}
	packet2 := []byte{0x00, 0x02, 0xBB}
	packet3 := []byte{0x00, 0x03, 0xCC}

	s.sendPackets <- packet{packet1, true}
	s.sendPackets <- packet{packet2, true}
	s.sendPackets <- packet{packet3, true}

	time.Sleep(100 * time.Millisecond)
	s.closed.Store(true)
	time.Sleep(50 * time.Millisecond)

	sentPackets := mock.GetSentPackets()

	// Should have 3 separate packets
	if len(sentPackets) != 3 {
		t.Fatalf("got %d packets, want 3", len(sentPackets))
	}

	// Each packet should NOT contain data from other packets
	// Verify packet 1 doesn't contain 0xBB or 0xCC
	if bytes.Contains(sentPackets[0], []byte{0xBB}) {
		t.Error("packet 1 contains data from packet 2 (concatenation detected)")
	}
	if bytes.Contains(sentPackets[0], []byte{0xCC}) {
		t.Error("packet 1 contains data from packet 3 (concatenation detected)")
	}

	// Verify packet 2 doesn't contain 0xCC
	if bytes.Contains(sentPackets[1], []byte{0xCC}) {
		t.Error("packet 2 contains data from packet 3 (concatenation detected)")
	}
}

// TestQueueSendUsesQueue verifies that QueueSend actually queues packets
// instead of sending them directly (the bug we fixed)
func TestQueueSendUsesQueue(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)

	// Don't start sendLoop yet - we want to verify packets are queued

	// Call QueueSend
	testData := []byte{0x00, 0x01, 0xAA, 0xBB}
	s.QueueSend(testData)

	// Give it a moment
	time.Sleep(10 * time.Millisecond)

	// WITHOUT sendLoop running, packets should NOT be sent yet
	if mock.PacketCount() > 0 {
		t.Error("QueueSend sent packet directly instead of queueing it")
	}

	// Verify packet is in the queue
	if len(s.sendPackets) != 1 {
		t.Errorf("expected 1 packet in queue, got %d", len(s.sendPackets))
	}

	// Now start sendLoop and verify it gets sent
	go s.sendLoop()
	time.Sleep(100 * time.Millisecond)

	if mock.PacketCount() != 1 {
		t.Errorf("expected 1 packet sent after sendLoop, got %d", mock.PacketCount())
	}

	s.closed.Store(true)
}

// TestPacketTerminatorFormat verifies the exact terminator format
func TestPacketTerminatorFormat(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)

	go s.sendLoop()

	testData := []byte{0x00, 0x01, 0xAA, 0xBB}
	s.sendPackets <- packet{testData, true}

	time.Sleep(100 * time.Millisecond)
	s.closed.Store(true)
	time.Sleep(50 * time.Millisecond)

	sentPackets := mock.GetSentPackets()
	if len(sentPackets) != 1 {
		t.Fatalf("expected 1 packet, got %d", len(sentPackets))
	}

	pkt := sentPackets[0]

	// Packet should be: original data + 0x00 + 0x10
	expectedLen := len(testData) + 2
	if len(pkt) != expectedLen {
		t.Errorf("expected packet length %d, got %d", expectedLen, len(pkt))
	}

	// Verify terminator bytes
	if pkt[len(pkt)-2] != 0x00 {
		t.Errorf("expected terminator byte 1 to be 0x00, got 0x%02X", pkt[len(pkt)-2])
	}
	if pkt[len(pkt)-1] != 0x10 {
		t.Errorf("expected terminator byte 2 to be 0x10, got 0x%02X", pkt[len(pkt)-1])
	}

	// Verify original data is intact
	for i := 0; i < len(testData); i++ {
		if pkt[i] != testData[i] {
			t.Errorf("original data corrupted at byte %d: got 0x%02X, want 0x%02X", i, pkt[i], testData[i])
		}
	}
}

// TestQueueSendNonBlockingDropsOnFull verifies non-blocking queue behavior
func TestQueueSendNonBlockingDropsOnFull(t *testing.T) {
	// Create a mock logger to avoid nil pointer in QueueSendNonBlocking
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	// Create session with small queue
	s := createTestSession(mock)
	s.sendPackets = make(chan packet, 2) // Override with smaller queue

	// Don't start sendLoop - let queue fill up

	// Fill the queue
	testData1 := []byte{0x00, 0x01}
	testData2 := []byte{0x00, 0x02}
	testData3 := []byte{0x00, 0x03}

	s.QueueSendNonBlocking(testData1)
	s.QueueSendNonBlocking(testData2)

	// Queue is now full (capacity 2)
	// This should be dropped
	s.QueueSendNonBlocking(testData3)

	// Verify only 2 packets in queue
	if len(s.sendPackets) != 2 {
		t.Errorf("expected 2 packets in queue, got %d", len(s.sendPackets))
	}

	s.closed.Store(true)
}

// TestPacketQueueAckFormat verifies ACK packet format
func TestPacketQueueAckFormat(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)

	go s.sendLoop()

	// Queue an ACK
	ackHandle := uint32(0x12345678)
	ackData := []byte{0xAA, 0xBB, 0xCC, 0xDD}
	s.QueueAck(ackHandle, ackData)

	time.Sleep(100 * time.Millisecond)
	s.closed.Store(true)
	time.Sleep(50 * time.Millisecond)

	sentPackets := mock.GetSentPackets()
	if len(sentPackets) != 1 {
		t.Fatalf("expected 1 ACK packet, got %d", len(sentPackets))
	}

	pkt := sentPackets[0]

	// Verify ACK packet structure:
	// 2 bytes: MSG_SYS_ACK opcode
	// 4 bytes: ack handle
	// N bytes: data
	// 2 bytes: terminator

	if len(pkt) < 8 {
		t.Fatalf("ACK packet too short: %d bytes", len(pkt))
	}

	// Check opcode
	opcode := binary.BigEndian.Uint16(pkt[0:2])
	if opcode != uint16(network.MSG_SYS_ACK) {
		t.Errorf("expected MSG_SYS_ACK opcode 0x%04X, got 0x%04X", network.MSG_SYS_ACK, opcode)
	}

	// Check ack handle
	receivedHandle := binary.BigEndian.Uint32(pkt[2:6])
	if receivedHandle != ackHandle {
		t.Errorf("expected ack handle 0x%08X, got 0x%08X", ackHandle, receivedHandle)
	}

	// Check data
	receivedData := pkt[6 : len(pkt)-2]
	if !bytes.Equal(receivedData, ackData) {
		t.Errorf("ACK data mismatch: got %v, want %v", receivedData, ackData)
	}

	// Check terminator
	if pkt[len(pkt)-2] != 0x00 || pkt[len(pkt)-1] != 0x10 {
		t.Error("ACK packet missing proper terminator")
	}
}
