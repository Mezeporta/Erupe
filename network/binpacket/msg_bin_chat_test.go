package binpacket

import (
	"bytes"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"testing"
)

func TestMsgBinChat_Opcode(t *testing.T) {
	msg := &MsgBinChat{}
	if msg.Opcode() != network.MSG_SYS_CAST_BINARY {
		t.Errorf("Opcode() = %v, want %v", msg.Opcode(), network.MSG_SYS_CAST_BINARY)
	}
}

func TestMsgBinChat_Build(t *testing.T) {
	tests := []struct {
		name     string
		msg      *MsgBinChat
		wantErr  bool
		validate func(*testing.T, []byte)
	}{
		{
			name: "basic message",
			msg: &MsgBinChat{
				Unk0:       0x01,
				Type:       ChatTypeWorld,
				Flags:      0x0000,
				Message:    "Hello",
				SenderName: "Player1",
			},
			wantErr: false,
			validate: func(t *testing.T, data []byte) {
				if len(data) == 0 {
					t.Error("Build() returned empty data")
				}
				// Verify the structure starts with Unk0, Type, Flags
				if data[0] != 0x01 {
					t.Errorf("Unk0 = 0x%X, want 0x01", data[0])
				}
				if data[1] != byte(ChatTypeWorld) {
					t.Errorf("Type = 0x%X, want 0x%X", data[1], byte(ChatTypeWorld))
				}
			},
		},
		{
			name: "all chat types",
			msg: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeStage,
				Flags:      0x1234,
				Message:    "Test",
				SenderName: "Sender",
			},
			wantErr: false,
		},
		{
			name: "empty message",
			msg: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeGuild,
				Flags:      0x0000,
				Message:    "",
				SenderName: "Player",
			},
			wantErr: false,
		},
		{
			name: "empty sender",
			msg: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeParty,
				Flags:      0x0000,
				Message:    "Hello",
				SenderName: "",
			},
			wantErr: false,
		},
		{
			name: "long message",
			msg: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeWhisper,
				Flags:      0x0000,
				Message:    "This is a very long message that contains a lot of text to test the handling of longer strings in the binary packet format.",
				SenderName: "LongNamePlayer",
			},
			wantErr: false,
		},
		{
			name: "special characters",
			msg: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeAlliance,
				Flags:      0x0000,
				Message:    "Hello!@#$%^&*()",
				SenderName: "Player_123",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			err := tt.msg.Build(bf)

			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				data := bf.Data()
				if tt.validate != nil {
					tt.validate(t, data)
				}
			}
		})
	}
}

func TestMsgBinChat_Parse(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    *MsgBinChat
		wantErr bool
	}{
		{
			name: "basic message",
			data: []byte{
				0x01,       // Unk0
				0x00,       // Type (ChatTypeWorld)
				0x00, 0x00, // Flags
				0x00, 0x08, // lenSenderName (8)
				0x00, 0x06, // lenMessage (6)
				// Message: "Hello" + null terminator (SJIS compatible ASCII)
				0x48, 0x65, 0x6C, 0x6C, 0x6F, 0x00,
				// SenderName: "Player1" + null terminator
				0x50, 0x6C, 0x61, 0x79, 0x65, 0x72, 0x31, 0x00,
			},
			want: &MsgBinChat{
				Unk0:       0x01,
				Type:       ChatTypeWorld,
				Flags:      0x0000,
				Message:    "Hello",
				SenderName: "Player1",
			},
			wantErr: false,
		},
		{
			name: "different chat type",
			data: []byte{
				0x00,       // Unk0
				0x02,       // Type (ChatTypeGuild)
				0x12, 0x34, // Flags
				0x00, 0x05, // lenSenderName
				0x00, 0x03, // lenMessage
				// Message: "Hi" + null
				0x48, 0x69, 0x00,
				// SenderName: "Bob" + null + padding
				0x42, 0x6F, 0x62, 0x00, 0x00,
			},
			want: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeGuild,
				Flags:      0x1234,
				Message:    "Hi",
				SenderName: "Bob",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrameFromBytes(tt.data)
			msg := &MsgBinChat{}

			err := msg.Parse(bf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if msg.Unk0 != tt.want.Unk0 {
					t.Errorf("Unk0 = 0x%X, want 0x%X", msg.Unk0, tt.want.Unk0)
				}
				if msg.Type != tt.want.Type {
					t.Errorf("Type = %v, want %v", msg.Type, tt.want.Type)
				}
				if msg.Flags != tt.want.Flags {
					t.Errorf("Flags = 0x%X, want 0x%X", msg.Flags, tt.want.Flags)
				}
				if msg.Message != tt.want.Message {
					t.Errorf("Message = %q, want %q", msg.Message, tt.want.Message)
				}
				if msg.SenderName != tt.want.SenderName {
					t.Errorf("SenderName = %q, want %q", msg.SenderName, tt.want.SenderName)
				}
			}
		})
	}
}

func TestMsgBinChat_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		msg  *MsgBinChat
	}{
		{
			name: "world chat",
			msg: &MsgBinChat{
				Unk0:       0x01,
				Type:       ChatTypeWorld,
				Flags:      0x0000,
				Message:    "Hello World",
				SenderName: "TestPlayer",
			},
		},
		{
			name: "stage chat",
			msg: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeStage,
				Flags:      0x1234,
				Message:    "Stage message",
				SenderName: "Player2",
			},
		},
		{
			name: "guild chat",
			msg: &MsgBinChat{
				Unk0:       0x02,
				Type:       ChatTypeGuild,
				Flags:      0xFFFF,
				Message:    "Guild announcement",
				SenderName: "GuildMaster",
			},
		},
		{
			name: "alliance chat",
			msg: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeAlliance,
				Flags:      0x0001,
				Message:    "Alliance msg",
				SenderName: "AllyLeader",
			},
		},
		{
			name: "party chat",
			msg: &MsgBinChat{
				Unk0:       0x01,
				Type:       ChatTypeParty,
				Flags:      0x0000,
				Message:    "Party up!",
				SenderName: "PartyLeader",
			},
		},
		{
			name: "whisper",
			msg: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeWhisper,
				Flags:      0x0002,
				Message:    "Secret message",
				SenderName: "Whisperer",
			},
		},
		{
			name: "empty strings",
			msg: &MsgBinChat{
				Unk0:       0x00,
				Type:       ChatTypeWorld,
				Flags:      0x0000,
				Message:    "",
				SenderName: "",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Build
			bf := byteframe.NewByteFrame()
			err := tt.msg.Build(bf)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			// Parse
			parsedMsg := &MsgBinChat{}
			parsedBf := byteframe.NewByteFrameFromBytes(bf.Data())
			err = parsedMsg.Parse(parsedBf)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Compare
			if parsedMsg.Unk0 != tt.msg.Unk0 {
				t.Errorf("Unk0 = 0x%X, want 0x%X", parsedMsg.Unk0, tt.msg.Unk0)
			}
			if parsedMsg.Type != tt.msg.Type {
				t.Errorf("Type = %v, want %v", parsedMsg.Type, tt.msg.Type)
			}
			if parsedMsg.Flags != tt.msg.Flags {
				t.Errorf("Flags = 0x%X, want 0x%X", parsedMsg.Flags, tt.msg.Flags)
			}
			if parsedMsg.Message != tt.msg.Message {
				t.Errorf("Message = %q, want %q", parsedMsg.Message, tt.msg.Message)
			}
			if parsedMsg.SenderName != tt.msg.SenderName {
				t.Errorf("SenderName = %q, want %q", parsedMsg.SenderName, tt.msg.SenderName)
			}
		})
	}
}

func TestChatType_Values(t *testing.T) {
	tests := []struct {
		chatType ChatType
		expected uint8
	}{
		{ChatTypeWorld, 0},
		{ChatTypeStage, 1},
		{ChatTypeGuild, 2},
		{ChatTypeAlliance, 3},
		{ChatTypeParty, 4},
		{ChatTypeWhisper, 5},
	}

	for _, tt := range tests {
		if uint8(tt.chatType) != tt.expected {
			t.Errorf("ChatType value = %d, want %d", uint8(tt.chatType), tt.expected)
		}
	}
}

func TestMsgBinChat_BuildParseConsistency(t *testing.T) {
	// Test that Build and Parse are consistent with each other
	// by building, parsing, building again, and comparing
	original := &MsgBinChat{
		Unk0:       0x01,
		Type:       ChatTypeWorld,
		Flags:      0x1234,
		Message:    "Test message",
		SenderName: "TestSender",
	}

	// First build
	bf1 := byteframe.NewByteFrame()
	err := original.Build(bf1)
	if err != nil {
		t.Fatalf("First Build() error = %v", err)
	}

	// Parse
	parsed := &MsgBinChat{}
	parsedBf := byteframe.NewByteFrameFromBytes(bf1.Data())
	err = parsed.Parse(parsedBf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Second build
	bf2 := byteframe.NewByteFrame()
	err = parsed.Build(bf2)
	if err != nil {
		t.Fatalf("Second Build() error = %v", err)
	}

	// Compare the two builds
	if !bytes.Equal(bf1.Data(), bf2.Data()) {
		t.Errorf("Build-Parse-Build inconsistency:\nFirst:  %v\nSecond: %v", bf1.Data(), bf2.Data())
	}
}
