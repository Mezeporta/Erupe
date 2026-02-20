package channelserver

import (
	"encoding/binary"
	_config "erupe-ce/config"
	"erupe-ce/network"
	"sync"
	"testing"
	"time"
)

const skipIntegrationTestMsg = "skipping integration test in short mode"

// IntegrationTest_PacketQueueFlow verifies the complete packet flow
// from queueing to sending, ensuring packets are sent individually
func IntegrationTest_PacketQueueFlow(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationTestMsg)
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
			s.closed.Store(true)
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
		t.Skip(skipIntegrationTestMsg)
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	s := &Session{
		sendPackets: make(chan packet, 200),
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
	s.closed.Store(true)
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
		t.Skip(skipIntegrationTestMsg)
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	s := &Session{
		sendPackets: make(chan packet, 100),
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
	s.closed.Store(true)
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
		t.Skip(skipIntegrationTestMsg)
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	s := &Session{
		sendPackets: make(chan packet, 100),
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
	s.closed.Store(true)
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
		t.Skip(skipIntegrationTestMsg)
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	s := &Session{
		sendPackets: make(chan packet, 100),
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
	s.closed.Store(true)
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
		t.Skip(skipIntegrationTestMsg)
	}

	// Fixed with network.Conn interface
	// Mock implementation available

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}

	// Small queue to test backpressure
	s := &Session{
		sendPackets: make(chan packet, 5),
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
	s.closed.Store(true)
	time.Sleep(50 * time.Millisecond)

	// Some packets should have been sent
	sentCount := mock.PacketCount()
	if sentCount == 0 {
		t.Error("no packets sent despite queueing attempts")
	}

	t.Logf("Successfully queued %d/%d packets, sent %d", successCount, attemptCount, sentCount)
}

// IntegrationTest_GuildEnumerationFlow tests end-to-end guild enumeration
func IntegrationTest_GuildEnumerationFlow(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationTestMsg)
	}

	tests := []struct {
		name         string
		guildCount   int
		membersPerGuild int
		wantValid    bool
	}{
		{
			name:            "single_guild",
			guildCount:      1,
			membersPerGuild: 1,
			wantValid:       true,
		},
		{
			name:            "multiple_guilds",
			guildCount:      10,
			membersPerGuild: 5,
			wantValid:       true,
		},
		{
			name:            "large_guilds",
			guildCount:      100,
			membersPerGuild: 50,
			wantValid:       true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)

			go s.sendLoop()

			// Simulate guild enumeration request
			for i := 0; i < tt.guildCount; i++ {
				guildData := make([]byte, 100) // Simplified guild data
				for j := 0; j < len(guildData); j++ {
					guildData[j] = byte((i*256 + j) % 256)
				}
				s.QueueSend(guildData)
			}

			// Wait for processing
			timeout := time.After(3 * time.Second)
			ticker := time.NewTicker(50 * time.Millisecond)
			defer ticker.Stop()

			for {
				select {
				case <-timeout:
					t.Fatal("timeout waiting for guild enumeration")
				case <-ticker.C:
					if mock.PacketCount() >= tt.guildCount {
						goto done
					}
				}
			}

		done:
			s.closed.Store(true)
			time.Sleep(50 * time.Millisecond)

			sentPackets := mock.GetSentPackets()
			if len(sentPackets) != tt.guildCount {
				t.Errorf("guild enumeration: got %d packets, want %d", len(sentPackets), tt.guildCount)
			}

			// Verify each guild packet has terminator
			for i, pkt := range sentPackets {
				if len(pkt) < 2 {
					t.Errorf("guild packet %d too short", i)
					continue
				}
				if pkt[len(pkt)-2] != 0x00 || pkt[len(pkt)-1] != 0x10 {
					t.Errorf("guild packet %d missing terminator", i)
				}
			}
		})
	}
}

// IntegrationTest_ConcurrentClientAccess tests concurrent client access scenarios
func IntegrationTest_ConcurrentClientAccess(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationTestMsg)
	}

	tests := []struct {
		name            string
		concurrentClients int
		packetsPerClient  int
		wantTotalPackets  int
	}{
		{
			name:             "two_concurrent_clients",
			concurrentClients: 2,
			packetsPerClient:  5,
			wantTotalPackets: 10,
		},
		{
			name:             "five_concurrent_clients",
			concurrentClients: 5,
			packetsPerClient:  10,
			wantTotalPackets: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var wg sync.WaitGroup
			totalPackets := 0
			var mu sync.Mutex

			wg.Add(tt.concurrentClients)

			for clientID := 0; clientID < tt.concurrentClients; clientID++ {
				go func(cid int) {
					defer wg.Done()

					mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
					s := createTestSession(mock)
					go s.sendLoop()

					// Client sends packets
					for i := 0; i < tt.packetsPerClient; i++ {
						testData := []byte{byte(cid), byte(i), 0xAA, 0xBB}
						s.QueueSend(testData)
					}

					time.Sleep(100 * time.Millisecond)
					s.closed.Store(true)
					time.Sleep(50 * time.Millisecond)

					sentCount := mock.PacketCount()
					mu.Lock()
					totalPackets += sentCount
					mu.Unlock()
				}(clientID)
			}

			wg.Wait()

			if totalPackets != tt.wantTotalPackets {
				t.Errorf("concurrent access: got %d packets, want %d", totalPackets, tt.wantTotalPackets)
			}
		})
	}
}

// IntegrationTest_ClientVersionCompatibility tests version-specific packet handling
func IntegrationTest_ClientVersionCompatibility(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationTestMsg)
	}

	tests := []struct {
		name          string
		clientVersion _config.Mode
		shouldSucceed bool
	}{
		{
			name:          "version_z2",
			clientVersion: _config.Z2,
			shouldSucceed: true,
		},
		{
			name:          "version_s6",
			clientVersion: _config.S6,
			shouldSucceed: true,
		},
		{
			name:          "version_g32",
			clientVersion: _config.G32,
			shouldSucceed: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := &Session{
				sendPackets: make(chan packet, 100),
				server: &Server{
					erupeConfig: &_config.Config{
						RealClientMode: tt.clientVersion,
					},
				},
			}
			s.cryptConn = mock

			go s.sendLoop()

			// Send version-specific packet
			testData := []byte{0x00, 0x01, 0xAA, 0xBB}
			s.QueueSend(testData)

			time.Sleep(100 * time.Millisecond)
			s.closed.Store(true)
			time.Sleep(50 * time.Millisecond)

			sentCount := mock.PacketCount()
			if (sentCount > 0) != tt.shouldSucceed {
				t.Errorf("version compatibility: got %d packets, shouldSucceed %v", sentCount, tt.shouldSucceed)
			}
		})
	}
}

// IntegrationTest_PacketPrioritization tests handling of priority packets
func IntegrationTest_PacketPrioritization(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationTestMsg)
	}

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)

	go s.sendLoop()

	// Queue normal priority packets
	for i := 0; i < 5; i++ {
		s.QueueSend([]byte{0x00, byte(i), 0xAA})
	}

	// Queue high priority ACK packet
	s.QueueAck(0x12345678, []byte{0xBB, 0xCC})

	// Queue more normal packets
	for i := 5; i < 10; i++ {
		s.QueueSend([]byte{0x00, byte(i), 0xDD})
	}

	time.Sleep(200 * time.Millisecond)
	s.closed.Store(true)
	time.Sleep(50 * time.Millisecond)

	sentPackets := mock.GetSentPackets()
	if len(sentPackets) < 10 {
		t.Errorf("expected at least 10 packets, got %d", len(sentPackets))
	}

	// Verify all packets have terminators
	for i, pkt := range sentPackets {
		if len(pkt) < 2 || pkt[len(pkt)-2] != 0x00 || pkt[len(pkt)-1] != 0x10 {
			t.Errorf("packet %d missing or invalid terminator", i)
		}
	}
}

// IntegrationTest_DataIntegrityUnderLoad tests data integrity under load
func IntegrationTest_DataIntegrityUnderLoad(t *testing.T) {
	if testing.Short() {
		t.Skip(skipIntegrationTestMsg)
	}

	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)

	go s.sendLoop()

	// Send large number of packets with unique identifiers
	packetCount := 100
	for i := range packetCount {
		// Each packet contains a unique identifier
		testData := make([]byte, 10)
		binary.LittleEndian.PutUint32(testData[0:4], uint32(i))
		binary.LittleEndian.PutUint32(testData[4:8], uint32(i*2))
		testData[8] = 0xAA
		testData[9] = 0xBB
		s.QueueSend(testData)
	}

	// Wait for processing
	timeout := time.After(5 * time.Second)
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("timeout waiting for packets under load")
		case <-ticker.C:
			if mock.PacketCount() >= packetCount {
				goto done
			}
		}
	}

done:
	s.closed.Store(true)
	time.Sleep(50 * time.Millisecond)

	sentPackets := mock.GetSentPackets()
	if len(sentPackets) != packetCount {
		t.Errorf("data integrity: got %d packets, want %d", len(sentPackets), packetCount)
	}

	// Verify no duplicate packets
	seen := make(map[string]bool)
	for i, pkt := range sentPackets {
		packetStr := string(pkt)
		if seen[packetStr] && len(pkt) > 2 {
			t.Errorf("duplicate packet detected at index %d", i)
		}
		seen[packetStr] = true
	}
}
