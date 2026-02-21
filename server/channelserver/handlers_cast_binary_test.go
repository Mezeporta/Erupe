package channelserver

import (
	"net"
	"slices"
	"strings"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcourse"
	cfg "erupe-ce/config"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
)

// TestSendServerChatMessage verifies that server chat messages are correctly formatted and queued
func TestSendServerChatMessage(t *testing.T) {
	tests := []struct {
		name    string
		message string
		wantErr bool
	}{
		{
			name:    "simple_message",
			message: "Hello, World!",
			wantErr: false,
		},
		{
			name:    "empty_message",
			message: "",
			wantErr: false,
		},
		{
			name:    "special_characters",
			message: "Test @#$%^&*()",
			wantErr: false,
		},
		{
			name:    "unicode_message",
			message: "テスト メッセージ",
			wantErr: false,
		},
		{
			name:    "long_message",
			message: strings.Repeat("A", 1000),
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)

			// Send the chat message
			sendServerChatMessage(s, tt.message)

			// Verify the message was queued
			if len(s.sendPackets) == 0 {
				t.Error("no packets were queued")
				return
			}

			// Read from the channel with timeout to avoid hanging
			select {
			case pkt := <-s.sendPackets:
				if pkt.data == nil {
					t.Error("packet data is nil")
				}
				// Verify it's an MHFPacket (contains opcode)
				if len(pkt.data) < 2 {
					t.Error("packet too short to contain opcode")
				}
			default:
				t.Error("no packet available in channel")
			}
		})
	}
}

// TestHandleMsgSysCastBinary_SimpleData verifies basic data message handling
func TestHandleMsgSysCastBinary_SimpleData(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = 54321
	s.stage = NewStage("test_stage")
	s.stage.clients[s] = s.charID
	s.server.sessions = make(map[net.Conn]*Session)

	// Create a data message payload
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	bf.WriteUint32(0xDEADBEEF)

	pkt := &mhfpacket.MsgSysCastBinary{
		Unk:            0,
		BroadcastType:  BroadcastTypeStage,
		MessageType:    BinaryMessageTypeData,
		RawDataPayload: bf.Data(),
	}

	// Should not panic
	handleMsgSysCastBinary(s, pkt)
}

// TestHandleMsgSysCastBinary_DiceCommand verifies the @dice command
func TestHandleMsgSysCastBinary_DiceCommand(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = 99999
	s.stage = NewStage("test_stage")
	s.stage.clients[s] = s.charID
	s.server.sessions = make(map[net.Conn]*Session)

	// Build a chat message with @dice command
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	msg := &binpacket.MsgBinChat{
		Unk0:       0,
		Type:       5,
		Flags:      0x80,
		Message:    "@dice",
		SenderName: "TestPlayer",
	}
	_ = msg.Build(bf)

	pkt := &mhfpacket.MsgSysCastBinary{
		Unk:            0,
		BroadcastType:  BroadcastTypeStage,
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}

	// Should execute dice command and return
	handleMsgSysCastBinary(s, pkt)

	// Verify a response was queued (dice result)
	if len(s.sendPackets) == 0 {
		t.Error("dice command did not queue a response")
	}
}

// TestBroadcastTypes verifies different broadcast types are handled
func TestBroadcastTypes(t *testing.T) {
	tests := []struct {
		name          string
		broadcastType uint8
		buildPayload  func() []byte
	}{
		{
			name:          "broadcast_targeted",
			broadcastType: BroadcastTypeTargeted,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetBE() // Targeted uses BE
				msg := &binpacket.MsgBinTargeted{
					TargetCharIDs: []uint32{1, 2, 3},
					RawDataPayload: []byte{0xDE, 0xAD, 0xBE, 0xEF},
				}
				_ = msg.Build(bf)
				return bf.Data()
			},
		},
		{
			name:          "broadcast_stage",
			broadcastType: BroadcastTypeStage,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetLE()
				bf.WriteUint32(0x12345678)
				return bf.Data()
			},
		},
		{
			name:          "broadcast_server",
			broadcastType: BroadcastTypeServer,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetLE()
				bf.WriteUint32(0x12345678)
				return bf.Data()
			},
		},
		{
			name:          "broadcast_world",
			broadcastType: BroadcastTypeWorld,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetLE()
				bf.WriteUint32(0x12345678)
				return bf.Data()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)
			s.charID = 22222
			s.stage = NewStage("test_stage")
			s.stage.clients[s] = s.charID
			s.server.sessions = make(map[net.Conn]*Session)

			pkt := &mhfpacket.MsgSysCastBinary{
				Unk:            0,
				BroadcastType:  tt.broadcastType,
				MessageType:    BinaryMessageTypeState,
				RawDataPayload: tt.buildPayload(),
			}

			// Should handle without panic
			handleMsgSysCastBinary(s, pkt)
		})
	}
}

// TestBinaryMessageTypes verifies different message types are handled
func TestBinaryMessageTypes(t *testing.T) {
	tests := []struct {
		name        string
		messageType uint8
		buildPayload func() []byte
	}{
		{
			name:        "msg_type_state",
			messageType: BinaryMessageTypeState,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetLE()
				bf.WriteUint32(0xDEADBEEF)
				return bf.Data()
			},
		},
		{
			name:        "msg_type_chat",
			messageType: BinaryMessageTypeChat,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetLE()
				msg := &binpacket.MsgBinChat{
					Unk0:       0,
					Type:       5,
					Flags:      0x80,
					Message:    "test",
					SenderName: "Player",
				}
				_ = msg.Build(bf)
				return bf.Data()
			},
		},
		{
			name:        "msg_type_quest",
			messageType: BinaryMessageTypeQuest,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetLE()
				bf.WriteUint32(0xDEADBEEF)
				return bf.Data()
			},
		},
		{
			name:        "msg_type_data",
			messageType: BinaryMessageTypeData,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetLE()
				bf.WriteUint32(0xDEADBEEF)
				return bf.Data()
			},
		},
		{
			name:        "msg_type_mail_notify",
			messageType: BinaryMessageTypeMailNotify,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetLE()
				bf.WriteUint32(0xDEADBEEF)
				return bf.Data()
			},
		},
		{
			name:        "msg_type_emote",
			messageType: BinaryMessageTypeEmote,
			buildPayload: func() []byte {
				bf := byteframe.NewByteFrame()
				bf.SetLE()
				bf.WriteUint32(0xDEADBEEF)
				return bf.Data()
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)
			s.charID = 33333
			s.stage = NewStage("test_stage")
			s.stage.clients[s] = s.charID
			s.server.sessions = make(map[net.Conn]*Session)

			pkt := &mhfpacket.MsgSysCastBinary{
				Unk:            0,
				BroadcastType:  BroadcastTypeStage,
				MessageType:    tt.messageType,
				RawDataPayload: tt.buildPayload(),
			}

			// Should handle without panic
			handleMsgSysCastBinary(s, pkt)
		})
	}
}

// TestSlicesContainsUsage verifies the slices.Contains function works correctly
func TestSlicesContainsUsage(t *testing.T) {
	tests := []struct {
		name     string
		items    []cfg.Course
		target   cfg.Course
		expected bool
	}{
		{
			name: "item_exists",
			items: []cfg.Course{
				{Name: "Course1", Enabled: true},
				{Name: "Course2", Enabled: false},
			},
			target:   cfg.Course{Name: "Course1", Enabled: true},
			expected: true,
		},
		{
			name: "item_not_found",
			items: []cfg.Course{
				{Name: "Course1", Enabled: true},
				{Name: "Course2", Enabled: false},
			},
			target:   cfg.Course{Name: "Course3", Enabled: true},
			expected: false,
		},
		{
			name:     "empty_slice",
			items:    []cfg.Course{},
			target:   cfg.Course{Name: "Course1", Enabled: true},
			expected: false,
		},
		{
			name: "enabled_mismatch",
			items: []cfg.Course{
				{Name: "Course1", Enabled: true},
			},
			target:   cfg.Course{Name: "Course1", Enabled: false},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slices.Contains(tt.items, tt.target)
			if result != tt.expected {
				t.Errorf("slices.Contains() = %v, want %v", result, tt.expected)
			}
		})
	}
}

// TestSlicesIndexFuncUsage verifies the slices.IndexFunc function works correctly
func TestSlicesIndexFuncUsage(t *testing.T) {
	tests := []struct {
		name      string
		courses   []mhfcourse.Course
		predicate func(mhfcourse.Course) bool
		expected  int
	}{
		{
			name:    "empty_slice",
			courses: []mhfcourse.Course{},
			predicate: func(c mhfcourse.Course) bool {
				return true
			},
			expected: -1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := slices.IndexFunc(tt.courses, tt.predicate)
			if result != tt.expected {
				t.Errorf("slices.IndexFunc() = %d, want %d", result, tt.expected)
			}
		})
	}
}

// TestChatMessageParsing verifies chat message extraction from binary payload
func TestChatMessageParsing(t *testing.T) {
	tests := []struct {
		name           string
		messageContent string
		authorName     string
	}{
		{
			name:           "standard_message",
			messageContent: "Hello World",
			authorName:     "Player123",
		},
		{
			name:           "special_chars_message",
			messageContent: "Test@#$%^&*()",
			authorName:     "SpecialUser",
		},
		{
			name:           "empty_message",
			messageContent: "",
			authorName:     "Silent",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build a binary chat message
			bf := byteframe.NewByteFrame()
			bf.SetLE()
			msg := &binpacket.MsgBinChat{
				Unk0:       0,
				Type:       5,
				Flags:      0x80,
				Message:    tt.messageContent,
				SenderName: tt.authorName,
			}
			_ = msg.Build(bf)

			// Parse it back
			parseBf := byteframe.NewByteFrameFromBytes(bf.Data())
			parseBf.SetLE()
			_, _ = parseBf.Seek(8, 0) // Skip initial bytes

			message := string(parseBf.ReadNullTerminatedBytes())
			author := string(parseBf.ReadNullTerminatedBytes())

			if message != tt.messageContent {
				t.Errorf("message mismatch: got %q, want %q", message, tt.messageContent)
			}
			if author != tt.authorName {
				t.Errorf("author mismatch: got %q, want %q", author, tt.authorName)
			}
		})
	}
}

// TestBinaryMessageTypeEnums verifies message type constants
func TestBinaryMessageTypeEnums(t *testing.T) {
	tests := []struct {
		name    string
		typeVal uint8
		typeID  uint8
	}{
		{
			name:    "state_type",
			typeVal: BinaryMessageTypeState,
			typeID:  0,
		},
		{
			name:    "chat_type",
			typeVal: BinaryMessageTypeChat,
			typeID:  1,
		},
		{
			name:    "quest_type",
			typeVal: BinaryMessageTypeQuest,
			typeID:  2,
		},
		{
			name:    "data_type",
			typeVal: BinaryMessageTypeData,
			typeID:  3,
		},
		{
			name:    "mail_notify_type",
			typeVal: BinaryMessageTypeMailNotify,
			typeID:  4,
		},
		{
			name:    "emote_type",
			typeVal: BinaryMessageTypeEmote,
			typeID:  6,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.typeVal != tt.typeID {
				t.Errorf("type mismatch: got %d, want %d", tt.typeVal, tt.typeID)
			}
		})
	}
}

// TestBroadcastTypeEnums verifies broadcast type constants
func TestBroadcastTypeEnums(t *testing.T) {
	tests := []struct {
		name    string
		typeVal uint8
		typeID  uint8
	}{
		{
			name:    "targeted_type",
			typeVal: BroadcastTypeTargeted,
			typeID:  0x01,
		},
		{
			name:    "stage_type",
			typeVal: BroadcastTypeStage,
			typeID:  0x03,
		},
		{
			name:    "server_type",
			typeVal: BroadcastTypeServer,
			typeID:  0x06,
		},
		{
			name:    "world_type",
			typeVal: BroadcastTypeWorld,
			typeID:  0x0a,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.typeVal != tt.typeID {
				t.Errorf("type mismatch: got %d, want %d", tt.typeVal, tt.typeID)
			}
		})
	}
}

// TestPayloadHandling verifies raw payload handling in different scenarios
func TestPayloadHandling(t *testing.T) {
	tests := []struct {
		name          string
		payloadSize   int
		broadcastType uint8
		messageType   uint8
	}{
		{
			name:          "empty_payload",
			payloadSize:   0,
			broadcastType: BroadcastTypeStage,
			messageType:   BinaryMessageTypeData,
		},
		{
			name:          "small_payload",
			payloadSize:   4,
			broadcastType: BroadcastTypeStage,
			messageType:   BinaryMessageTypeData,
		},
		{
			name:          "large_payload",
			payloadSize:   10000,
			broadcastType: BroadcastTypeStage,
			messageType:   BinaryMessageTypeData,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
			s := createTestSession(mock)
			s.charID = 44444
			s.stage = NewStage("test_stage")
			s.stage.clients[s] = s.charID
			s.server.sessions = make(map[net.Conn]*Session)

			// Create payload of specified size
			payload := make([]byte, tt.payloadSize)
			for i := 0; i < len(payload); i++ {
				payload[i] = byte(i % 256)
			}

			pkt := &mhfpacket.MsgSysCastBinary{
				Unk:            0,
				BroadcastType:  tt.broadcastType,
				MessageType:    tt.messageType,
				RawDataPayload: payload,
			}

			// Should handle without panic
			handleMsgSysCastBinary(s, pkt)
		})
	}
}

// TestCastedBinaryPacketConstruction verifies correct packet construction
func TestCastedBinaryPacketConstruction(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = 77777

	message := "Test message"

	sendServerChatMessage(s, message)

	// Verify a packet was queued
	if len(s.sendPackets) == 0 {
		t.Fatal("no packets queued")
	}

	// Extract packet from channel
	pkt := <-s.sendPackets

	if pkt.data == nil {
		t.Error("packet data is nil")
	}

	// The packet should be at least a valid MHF packet with opcode
	if len(pkt.data) < 2 {
		t.Error("packet too short")
	}
}

// TestNilPayloadHandling verifies safe handling of nil payloads
func TestNilPayloadHandling(t *testing.T) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = 55555
	s.stage = NewStage("test_stage")
	s.stage.clients[s] = s.charID
	s.server.sessions = make(map[net.Conn]*Session)

	pkt := &mhfpacket.MsgSysCastBinary{
		Unk:            0,
		BroadcastType:  BroadcastTypeStage,
		MessageType:    BinaryMessageTypeData,
		RawDataPayload: nil,
	}

	// Should handle nil payload without panic
	handleMsgSysCastBinary(s, pkt)
}

// BenchmarkSendServerChatMessage benchmarks the chat message sending
func BenchmarkSendServerChatMessage(b *testing.B) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)

	message := "This is a benchmark message"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sendServerChatMessage(s, message)
	}
}

// BenchmarkHandleMsgSysCastBinary benchmarks the packet handling
func BenchmarkHandleMsgSysCastBinary(b *testing.B) {
	mock := &MockCryptConn{sentPackets: make([][]byte, 0)}
	s := createTestSession(mock)
	s.charID = 99999
	s.stage = NewStage("test_stage")
	s.stage.clients[s] = s.charID
	s.server.sessions = make(map[net.Conn]*Session)

	// Prepare packet
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	bf.WriteUint32(0x12345678)

	pkt := &mhfpacket.MsgSysCastBinary{
		Unk:            0,
		BroadcastType:  BroadcastTypeStage,
		MessageType:    BinaryMessageTypeData,
		RawDataPayload: bf.Data(),
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		handleMsgSysCastBinary(s, pkt)
	}
}

// BenchmarkSlicesContains benchmarks the slices.Contains function
func BenchmarkSlicesContains(b *testing.B) {
	courses := []cfg.Course{
		{Name: "Course1", Enabled: true},
		{Name: "Course2", Enabled: false},
		{Name: "Course3", Enabled: true},
		{Name: "Course4", Enabled: false},
		{Name: "Course5", Enabled: true},
	}

	target := cfg.Course{Name: "Course3", Enabled: true}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = slices.Contains(courses, target)
	}
}

// BenchmarkSlicesIndexFunc benchmarks the slices.IndexFunc function
func BenchmarkSlicesIndexFunc(b *testing.B) {
	// Create mock courses (empty as real data not needed for benchmark)
	courses := make([]mhfcourse.Course, 100)

	predicate := func(c mhfcourse.Course) bool {
		return false // Worst case - always iterate to end
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = slices.IndexFunc(courses, predicate)
	}
}
