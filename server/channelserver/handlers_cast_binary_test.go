package channelserver

import (
	"testing"
)

func TestBinaryMessageTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"BinaryMessageTypeState", BinaryMessageTypeState, 0},
		{"BinaryMessageTypeChat", BinaryMessageTypeChat, 1},
		{"BinaryMessageTypeQuest", BinaryMessageTypeQuest, 2},
		{"BinaryMessageTypeData", BinaryMessageTypeData, 3},
		{"BinaryMessageTypeMailNotify", BinaryMessageTypeMailNotify, 4},
		{"BinaryMessageTypeEmote", BinaryMessageTypeEmote, 6},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestBroadcastTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		constant int
		expected int
	}{
		{"BroadcastTypeTargeted", BroadcastTypeTargeted, 0x01},
		{"BroadcastTypeStage", BroadcastTypeStage, 0x03},
		{"BroadcastTypeServer", BroadcastTypeServer, 0x06},
		{"BroadcastTypeWorld", BroadcastTypeWorld, 0x0a},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.constant != tt.expected {
				t.Errorf("%s = %d, want %d", tt.name, tt.constant, tt.expected)
			}
		})
	}
}

func TestCommandsMapInitialized(t *testing.T) {
	// commands map should be initialized by init()
	if commands == nil {
		t.Error("commands map should be initialized")
	}
}

func TestSendServerChatMessage(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Should not panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("sendServerChatMessage panicked: %v", r)
		}
	}()

	sendServerChatMessage(session, "Test message")

	// Should queue a packet
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No packet queued")
	}
}
