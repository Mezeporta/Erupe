package channelserver

import (
	"encoding/binary"
	_config "erupe-ce/config"
	"erupe-ce/network"
	"sync"
	"testing"
	"time"
)

// IntegrationTest_PacketQueueFlow verifies the complete packet flow
// from queueing to sending, ensuring packets are sent individually
func IntegrationTest_PacketQueueFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name         string
		packetCount  int
		queueDelay   time.Duration
		wantPackets  int
	}{
		{
			name:         "sequential_packets",
			packetCount:  10,
			queueDelay:   10 * time.Millisecond,
			wantPackets:  10,
		},
		{
			name:         "rapid_fire_packets",
			packetCount:  50,
			queueDelay:   1 * time.Millisecond,
			wantPackets:  50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

			s := &Session{
				sendPackets: make(chan packet, 100),
				closed:      false,
				server: &Server{
					erupeConfig: &_config.Config{
						DebugOptions: _config.DebugOptions{
							LogOutboundMessages: false,
						},
					},
				},
			}
			s.cryptConn = mock

			// Start send loop
			go s.sendLoop()

			// Queue packets with delay
			go func() {
				for i := 0; i < tt.packetCount; i++ {
					testData := []byte{0x00, byte(i), 0xAA, 0xBB}
					s.QueueSend(testData)
					time.Sleep(tt.queueDelay)
				}
			}()

			// Wait for all packets to be processed
			timeout := time.After(5 * time.Second)
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					t.Fatal("timeout waiting for packets")
				case <-ticker.C:
					if mock.PacketCount() >= tt.wantPackets {
						goto done
					}
				}
			}

		done:
			s.closed = true
			time.Sleep(50 * time.Millisecond)

			sentPackets := mock.GetSentPackets()
			if len(sentPackets) != tt.wantPackets {
				t.Errorf("got %d packets, want %d", len(sentPackets), tt.wantPackets)
			}

			// Verify each packet has terminator
			for i, pkt := range sentPackets {
				if len(pkt) < 2 {
					t.Errorf("packet %d too short", i)
					continue
				}
				if pkt[len(pkt)-2] != 0x00 || pkt[len(pkt)-1] != 0x10 {
					t.Errorf("packet %d missing terminator", i)
				}
			}
		})
	}
}

// IntegrationTest_ConcurrentQueueing verifies thread-safe packet queueing
func IntegrationTest_ConcurrentQueueing(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	s := &Session{
		sendPackets: make(chan packet, 200),
		closed:      false,
		server: &Server{
			erupeConfig: &_config.Config{
				DebugOptions: _config.DebugOptions{
					LogOutboundMessages: false,
				},
			},
		},
	}
		s.cryptConn = mock

	go s.sendLoop()

	// Number of concurrent goroutines
	goroutineCount := 10
	packetsPerGoroutine := 10
	expectedTotal := goroutineCount * packetsPerGoroutine

	var wg sync.WaitGroup
	wg.Add(goroutineCount)

	// Launch concurrent packet senders
	for g := 0; g < goroutineCount; g++ {
		go func(goroutineID int) {
			defer wg.Done()
			for i := 0; i < packetsPerGoroutine; i++ {
				testData := []byte{
					byte(goroutineID),
					byte(i),
					0xAA,
					0xBB,
				}
				s.QueueSend(testData)
			}
		}(g)
	}

	// Wait for all goroutines to finish queueing
	wg.Wait()

	// Wait for packets to be sent
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for packets")
		case <-ticker.C:
			if mock.PacketCount() >= expectedTotal {
				goto done
			}
		}
	}

done:
	s.closed = true
	time.Sleep(50 * time.Millisecond)

	sentPackets := mock.GetSentPackets()
	if len(sentPackets) != expectedTotal {
		t.Errorf("got %d packets, want %d", len(sentPackets), expectedTotal)
	}

	// Verify no packet concatenation occurred
	for i, pkt := range sentPackets {
		if len(pkt) < 2 {
			t.Errorf("packet %d too short", i)
			continue
		}

		// Each packet should have exactly one terminator at the end
		terminatorCount := 0
		for j := 0; j < len(pkt)-1; j++ {
			if pkt[j] == 0x00 && pkt[j+1] == 0x10 {
				terminatorCount++
			}
		}

		if terminatorCount != 1 {
			t.Errorf("packet %d has %d terminators, want 1", i, terminatorCount)
		}
	}
}

// IntegrationTest_AckPacketFlow verifies ACK packet generation and sending
func IntegrationTest_AckPacketFlow(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	s := &Session{
		sendPackets: make(chan packet, 100),
		closed:      false,
		server: &Server{
			erupeConfig: &_config.Config{
				DebugOptions: _config.DebugOptions{
					LogOutboundMessages: false,
				},
			},
		},
	}
		s.cryptConn = mock

	go s.sendLoop()

	// Queue multiple ACKs
	ackCount := 5
	for i := 0; i < ackCount; i++ {
		ackHandle := uint32(0x1000 + i)
		ackData := []byte{0xAA, 0xBB, byte(i), 0xDD}
		s.QueueAck(ackHandle, ackData)
	}

	// Wait for ACKs to be sent
	time.Sleep(200 * time.Millisecond)
	s.closed = true
	time.Sleep(50 * time.Millisecond)

	sentPackets := mock.GetSentPackets()
	if len(sentPackets) != ackCount {
		t.Fatalf("got %d ACK packets, want %d", len(sentPackets), ackCount)
	}

	// Verify each ACK packet structure
	for i, pkt := range sentPackets {
		// Check minimum length: opcode(2) + handle(4) + data(4) + terminator(2) = 12
		if len(pkt) < 12 {
			t.Errorf("ACK packet %d too short: %d bytes", i, len(pkt))
			continue
		}

		// Verify opcode
		opcode := binary.BigEndian.Uint16(pkt[0:2])
		if opcode != uint16(network.MSG_SYS_ACK) {
			t.Errorf("ACK packet %d wrong opcode: got 0x%04X, want 0x%04X",
				i, opcode, network.MSG_SYS_ACK)
		}

		// Verify terminator
		if pkt[len(pkt)-2] != 0x00 || pkt[len(pkt)-1] != 0x10 {
			t.Errorf("ACK packet %d missing terminator", i)
		}
	}
}

// IntegrationTest_MixedPacketTypes verifies different packet types don't interfere
func IntegrationTest_MixedPacketTypes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	s := &Session{
		sendPackets: make(chan packet, 100),
		closed:      false,
		server: &Server{
			erupeConfig: &_config.Config{
				DebugOptions: _config.DebugOptions{
					LogOutboundMessages: false,
				},
			},
		},
	}
		s.cryptConn = mock

	go s.sendLoop()

	// Mix different packet types
	// Regular packet
	s.QueueSend([]byte{0x00, 0x01, 0xAA})

	// ACK packet
	s.QueueAck(0x12345678, []byte{0xBB, 0xCC})

	// Another regular packet
	s.QueueSend([]byte{0x00, 0x02, 0xDD})

	// Non-blocking packet
	s.QueueSendNonBlocking([]byte{0x00, 0x03, 0xEE})

	// Wait for all packets
	time.Sleep(200 * time.Millisecond)
	s.closed = true
	time.Sleep(50 * time.Millisecond)

	sentPackets := mock.GetSentPackets()
	if len(sentPackets) != 4 {
		t.Fatalf("got %d packets, want 4", len(sentPackets))
	}

	// Verify each packet has its own terminator
	for i, pkt := range sentPackets {
		if pkt[len(pkt)-2] != 0x00 || pkt[len(pkt)-1] != 0x10 {
			t.Errorf("packet %d missing terminator", i)
		}
	}
}

// IntegrationTest_PacketOrderPreservation verifies packets are sent in order
func IntegrationTest_PacketOrderPreservation(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	s := &Session{
		sendPackets: make(chan packet, 100),
		closed:      false,
		server: &Server{
			erupeConfig: &_config.Config{
				DebugOptions: _config.DebugOptions{
					LogOutboundMessages: false,
				},
			},
		},
	}
		s.cryptConn = mock

	go s.sendLoop()

	// Queue packets with sequential identifiers
	packetCount := 20
	for i := 0; i < packetCount; i++ {
		testData := []byte{0x00, byte(i), 0xAA}
		s.QueueSend(testData)
	}

	// Wait for packets
	time.Sleep(300 * time.Millisecond)
	s.closed = true
	time.Sleep(50 * time.Millisecond)

	sentPackets := mock.GetSentPackets()
	if len(sentPackets) != packetCount {
		t.Fatalf("got %d packets, want %d", len(sentPackets), packetCount)
	}

	// Verify order is preserved
	for i, pkt := range sentPackets {
		if len(pkt) < 2 {
			t.Errorf("packet %d too short", i)
			continue
		}

		// Check the sequential byte we added
		if pkt[1] != byte(i) {
			t.Errorf("packet order violated: position %d has sequence byte %d", i, pkt[1])
		}
	}
}

// IntegrationTest_QueueBackpressure verifies behavior under queue pressure
func IntegrationTest_QueueBackpressure(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	// Small queue to test backpressure
	s := &Session{
		sendPackets: make(chan packet, 5),
		closed:      false,
		server: &Server{
			erupeConfig: &_config.Config{
				DebugOptions: _config.DebugOptions{
					LogOutboundMessages: false,
				},
				LoopDelay: 50, // Slower processing to create backpressure
			},
		},
	}
		s.cryptConn = mock

	go s.sendLoop()

	// Try to queue more than capacity using non-blocking
	attemptCount := 10
	successCount := 0

	for i := 0; i < attemptCount; i++ {
		testData := []byte{0x00, byte(i), 0xAA}
		select {
		case s.sendPackets <- packet{testData, true}:
			successCount++
		default:
			// Queue full, packet dropped
		}
		time.Sleep(5 * time.Millisecond)
	}

	// Wait for processing
	time.Sleep(1 * time.Second)
	s.closed = true
	time.Sleep(50 * time.Millisecond)

	// Some packets should have been sent
	sentCount := mock.PacketCount()
	if sentCount == 0 {
		t.Error("no packets sent despite queueing attempts")
	}

	t.Logf("Successfully queued %d/%d packets, sent %d", successCount, attemptCount, sentCount)
}
