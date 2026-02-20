package binpacket

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"testing"
)

func TestMsgBinMailNotify_Opcode(t *testing.T) {
	msg := MsgBinMailNotify{}
	if msg.Opcode() != network.MSG_SYS_CASTED_BINARY {
		t.Errorf("Opcode() = %v, want %v", msg.Opcode(), network.MSG_SYS_CASTED_BINARY)
	}
}

func TestMsgBinMailNotify_Build(t *testing.T) {
	tests := []struct {
		name       string
		senderName string
		wantErr    bool
		validate   func(*testing.T, []byte)
	}{
		{
			name:       "basic sender name",
			senderName: "Player1",
			wantErr:    false,
			validate: func(t *testing.T, data []byte) {
				if len(data) == 0 {
					t.Error("Build() returned empty data")
				}
				// First byte should be 0x01 (Unk)
				if data[0] != 0x01 {
					t.Errorf("First byte = 0x%X, want 0x01", data[0])
				}
				// Total length should be 1 (Unk) + 21 (padded string)
				expectedLen := 1 + 21
				if len(data) != expectedLen {
					t.Errorf("data length = %d, want %d", len(data), expectedLen)
				}
			},
		},
		{
			name:       "empty sender name",
			senderName: "",
			wantErr:    false,
			validate: func(t *testing.T, data []byte) {
				if len(data) != 22 { // 1 + 21
					t.Errorf("data length = %d, want 22", len(data))
				}
			},
		},
		{
			name:       "long sender name",
			senderName: "VeryLongPlayerNameThatExceeds21Characters",
			wantErr:    false,
			validate: func(t *testing.T, data []byte) {
				if len(data) != 22 { // 1 + 21 (truncated/padded)
					t.Errorf("data length = %d, want 22", len(data))
				}
			},
		},
		{
			name:       "exactly 21 characters",
			senderName: "ExactlyTwentyOneChar1",
			wantErr:    false,
			validate: func(t *testing.T, data []byte) {
				if len(data) != 22 {
					t.Errorf("data length = %d, want 22", len(data))
				}
			},
		},
		{
			name:       "special characters",
			senderName: "Player_123",
			wantErr:    false,
			validate: func(t *testing.T, data []byte) {
				if len(data) != 22 {
					t.Errorf("data length = %d, want 22", len(data))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := MsgBinMailNotify{
				SenderName: tt.senderName,
			}

			bf := byteframe.NewByteFrame()
			err := msg.Build(bf)

			if (err != nil) != tt.wantErr {
				t.Errorf("Build() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && tt.validate != nil {
				tt.validate(t, bf.Data())
			}
		})
	}
}

func TestMsgBinMailNotify_Parse_ReturnsError(t *testing.T) {
	// Document that Parse() is not implemented and returns an error
	msg := MsgBinMailNotify{}
	bf := byteframe.NewByteFrame()

	err := msg.Parse(bf)
	if err == nil {
		t.Error("Parse() should return an error (not implemented)")
	}
}

func TestMsgBinMailNotify_BuildMultiple(t *testing.T) {
	// Test building multiple messages to ensure no state pollution
	names := []string{"Player1", "Player2", "Player3"}

	for _, name := range names {
		msg := MsgBinMailNotify{SenderName: name}
		bf := byteframe.NewByteFrame()
		err := msg.Build(bf)
		if err != nil {
			t.Errorf("Build(%s) error = %v", name, err)
		}

		data := bf.Data()
		if len(data) != 22 {
			t.Errorf("Build(%s) length = %d, want 22", name, len(data))
		}
	}
}

func TestMsgBinMailNotify_PaddingBehavior(t *testing.T) {
	// Test that the padded string is always 21 bytes
	tests := []struct {
		name       string
		senderName string
	}{
		{"short", "A"},
		{"medium", "PlayerName"},
		{"long", "VeryVeryLongPlayerName"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			msg := MsgBinMailNotify{SenderName: tt.senderName}
			bf := byteframe.NewByteFrame()
			err := msg.Build(bf)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			data := bf.Data()
			// Skip first byte (Unk), check remaining 21 bytes
			if len(data) < 22 {
				t.Fatalf("data too short: %d bytes", len(data))
			}

			paddedString := data[1:22]
			if len(paddedString) != 21 {
				t.Errorf("padded string length = %d, want 21", len(paddedString))
			}
		})
	}
}

func TestMsgBinMailNotify_BuildStructure(t *testing.T) {
	// Test the structure of the built data
	msg := MsgBinMailNotify{SenderName: "Test"}
	bf := byteframe.NewByteFrame()
	err := msg.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	data := bf.Data()

	// Check structure: 1 byte Unk + 21 bytes padded string = 22 bytes total
	if len(data) != 22 {
		t.Errorf("data length = %d, want 22", len(data))
	}

	// First byte should be 0x01
	if data[0] != 0x01 {
		t.Errorf("Unk byte = 0x%X, want 0x01", data[0])
	}

	// The rest (21 bytes) should contain the sender name (SJIS encoded) and padding
	// We can't verify exact content without knowing SJIS encoding details,
	// but we can verify length
	paddedPortion := data[1:]
	if len(paddedPortion) != 21 {
		t.Errorf("padded portion length = %d, want 21", len(paddedPortion))
	}
}

func TestMsgBinMailNotify_ValueSemantics(t *testing.T) {
	// Test that MsgBinMailNotify uses value semantics (not pointer receiver for Opcode)
	msg := MsgBinMailNotify{SenderName: "Test"}

	// Should work with value
	opcode := msg.Opcode()
	if opcode != network.MSG_SYS_CASTED_BINARY {
		t.Errorf("Opcode() = %v, want %v", opcode, network.MSG_SYS_CASTED_BINARY)
	}

	// Should also work with pointer (Go allows this)
	msgPtr := &MsgBinMailNotify{SenderName: "Test"}
	opcode2 := msgPtr.Opcode()
	if opcode2 != network.MSG_SYS_CASTED_BINARY {
		t.Errorf("Opcode() on pointer = %v, want %v", opcode2, network.MSG_SYS_CASTED_BINARY)
	}
}
