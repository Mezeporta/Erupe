package binpacket

import (
	"bytes"
	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"testing"
)

func TestMsgBinTargeted_Opcode(t *testing.T) {
	msg := &MsgBinTargeted{}
	if msg.Opcode() != network.MSG_SYS_CAST_BINARY {
		t.Errorf("Opcode() = %v, want %v", msg.Opcode(), network.MSG_SYS_CAST_BINARY)
	}
}

func TestMsgBinTargeted_Build(t *testing.T) {
	tests := []struct {
		name     string
		msg      *MsgBinTargeted
		wantErr  bool
		validate func(*testing.T, []byte)
	}{
		{
			name: "single target with payload",
			msg: &MsgBinTargeted{
				TargetCount:    1,
				TargetCharIDs:  []uint32{12345},
				RawDataPayload: []byte{0x01, 0x02, 0x03, 0x04},
			},
			wantErr: false,
			validate: func(t *testing.T, data []byte) {
				if len(data) < 2+4+4 { // 2 bytes count + 4 bytes ID + 4 bytes payload
					t.Errorf("data length = %d, want at least %d", len(data), 2+4+4)
				}
			},
		},
		{
			name: "multiple targets",
			msg: &MsgBinTargeted{
				TargetCount:    3,
				TargetCharIDs:  []uint32{100, 200, 300},
				RawDataPayload: []byte{0xAA, 0xBB},
			},
			wantErr: false,
			validate: func(t *testing.T, data []byte) {
				expectedLen := 2 + (3 * 4) + 2 // count + 3 IDs + payload
				if len(data) != expectedLen {
					t.Errorf("data length = %d, want %d", len(data), expectedLen)
				}
			},
		},
		{
			name: "zero targets",
			msg: &MsgBinTargeted{
				TargetCount:    0,
				TargetCharIDs:  []uint32{},
				RawDataPayload: []byte{0xFF},
			},
			wantErr: false,
			validate: func(t *testing.T, data []byte) {
				if len(data) < 2+1 { // count + payload
					t.Errorf("data length = %d, want at least %d", len(data), 2+1)
				}
			},
		},
		{
			name: "empty payload",
			msg: &MsgBinTargeted{
				TargetCount:    1,
				TargetCharIDs:  []uint32{999},
				RawDataPayload: []byte{},
			},
			wantErr: false,
			validate: func(t *testing.T, data []byte) {
				expectedLen := 2 + 4 // count + 1 ID
				if len(data) != expectedLen {
					t.Errorf("data length = %d, want %d", len(data), expectedLen)
				}
			},
		},
		{
			name: "large payload",
			msg: &MsgBinTargeted{
				TargetCount:    2,
				TargetCharIDs:  []uint32{1000, 2000},
				RawDataPayload: bytes.Repeat([]byte{0xCC}, 256),
			},
			wantErr: false,
		},
		{
			name: "max uint32 target IDs",
			msg: &MsgBinTargeted{
				TargetCount:    2,
				TargetCharIDs:  []uint32{0xFFFFFFFF, 0x12345678},
				RawDataPayload: []byte{0x01},
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

func TestMsgBinTargeted_Parse(t *testing.T) {
	tests := []struct {
		name    string
		data    []byte
		want    *MsgBinTargeted
		wantErr bool
	}{
		{
			name: "single target",
			data: []byte{
				0x00, 0x01, // TargetCount = 1
				0x00, 0x00, 0x30, 0x39, // TargetCharID = 12345
				0xAA, 0xBB, 0xCC, // RawDataPayload
			},
			want: &MsgBinTargeted{
				TargetCount:    1,
				TargetCharIDs:  []uint32{12345},
				RawDataPayload: []byte{0xAA, 0xBB, 0xCC},
			},
			wantErr: false,
		},
		{
			name: "multiple targets",
			data: []byte{
				0x00, 0x03, // TargetCount = 3
				0x00, 0x00, 0x00, 0x64, // Target 1 = 100
				0x00, 0x00, 0x00, 0xC8, // Target 2 = 200
				0x00, 0x00, 0x01, 0x2C, // Target 3 = 300
				0x01, 0x02, // RawDataPayload
			},
			want: &MsgBinTargeted{
				TargetCount:    3,
				TargetCharIDs:  []uint32{100, 200, 300},
				RawDataPayload: []byte{0x01, 0x02},
			},
			wantErr: false,
		},
		{
			name: "zero targets",
			data: []byte{
				0x00, 0x00, // TargetCount = 0
				0xFF, 0xFF, // RawDataPayload
			},
			want: &MsgBinTargeted{
				TargetCount:    0,
				TargetCharIDs:  []uint32{},
				RawDataPayload: []byte{0xFF, 0xFF},
			},
			wantErr: false,
		},
		{
			name: "no payload",
			data: []byte{
				0x00, 0x01, // TargetCount = 1
				0x00, 0x00, 0x03, 0xE7, // Target = 999
			},
			want: &MsgBinTargeted{
				TargetCount:    1,
				TargetCharIDs:  []uint32{999},
				RawDataPayload: []byte{},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrameFromBytes(tt.data)
			msg := &MsgBinTargeted{}

			err := msg.Parse(bf)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if msg.TargetCount != tt.want.TargetCount {
					t.Errorf("TargetCount = %d, want %d", msg.TargetCount, tt.want.TargetCount)
				}

				if len(msg.TargetCharIDs) != len(tt.want.TargetCharIDs) {
					t.Errorf("len(TargetCharIDs) = %d, want %d", len(msg.TargetCharIDs), len(tt.want.TargetCharIDs))
				} else {
					for i, id := range msg.TargetCharIDs {
						if id != tt.want.TargetCharIDs[i] {
							t.Errorf("TargetCharIDs[%d] = %d, want %d", i, id, tt.want.TargetCharIDs[i])
						}
					}
				}

				if !bytes.Equal(msg.RawDataPayload, tt.want.RawDataPayload) {
					t.Errorf("RawDataPayload = %v, want %v", msg.RawDataPayload, tt.want.RawDataPayload)
				}
			}
		})
	}
}

func TestMsgBinTargeted_RoundTrip(t *testing.T) {
	tests := []struct {
		name string
		msg  *MsgBinTargeted
	}{
		{
			name: "single target",
			msg: &MsgBinTargeted{
				TargetCount:    1,
				TargetCharIDs:  []uint32{12345},
				RawDataPayload: []byte{0x01, 0x02, 0x03},
			},
		},
		{
			name: "multiple targets",
			msg: &MsgBinTargeted{
				TargetCount:    5,
				TargetCharIDs:  []uint32{100, 200, 300, 400, 500},
				RawDataPayload: []byte{0xAA, 0xBB, 0xCC, 0xDD},
			},
		},
		{
			name: "zero targets",
			msg: &MsgBinTargeted{
				TargetCount:    0,
				TargetCharIDs:  []uint32{},
				RawDataPayload: []byte{0xFF},
			},
		},
		{
			name: "empty payload",
			msg: &MsgBinTargeted{
				TargetCount:    2,
				TargetCharIDs:  []uint32{1000, 2000},
				RawDataPayload: []byte{},
			},
		},
		{
			name: "large IDs and payload",
			msg: &MsgBinTargeted{
				TargetCount:    3,
				TargetCharIDs:  []uint32{0xFFFFFFFF, 0x12345678, 0xABCDEF00},
				RawDataPayload: bytes.Repeat([]byte{0xDD}, 128),
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
			parsedMsg := &MsgBinTargeted{}
			parsedBf := byteframe.NewByteFrameFromBytes(bf.Data())
			err = parsedMsg.Parse(parsedBf)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Compare
			if parsedMsg.TargetCount != tt.msg.TargetCount {
				t.Errorf("TargetCount = %d, want %d", parsedMsg.TargetCount, tt.msg.TargetCount)
			}

			if len(parsedMsg.TargetCharIDs) != len(tt.msg.TargetCharIDs) {
				t.Errorf("len(TargetCharIDs) = %d, want %d", len(parsedMsg.TargetCharIDs), len(tt.msg.TargetCharIDs))
			} else {
				for i, id := range parsedMsg.TargetCharIDs {
					if id != tt.msg.TargetCharIDs[i] {
						t.Errorf("TargetCharIDs[%d] = %d, want %d", i, id, tt.msg.TargetCharIDs[i])
					}
				}
			}

			if !bytes.Equal(parsedMsg.RawDataPayload, tt.msg.RawDataPayload) {
				t.Errorf("RawDataPayload length mismatch: got %d, want %d", len(parsedMsg.RawDataPayload), len(tt.msg.RawDataPayload))
			}
		})
	}
}

func TestMsgBinTargeted_TargetCountMismatch(t *testing.T) {
	// Test that TargetCount and actual array length don't have to match
	// The Build function uses the TargetCount field
	msg := &MsgBinTargeted{
		TargetCount:    2, // Says 2
		TargetCharIDs:  []uint32{100, 200, 300}, // But has 3
		RawDataPayload: []byte{0x01},
	}

	bf := byteframe.NewByteFrame()
	err := msg.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Parse should read exactly 2 IDs as specified by TargetCount
	parsedMsg := &MsgBinTargeted{}
	parsedBf := byteframe.NewByteFrameFromBytes(bf.Data())
	err = parsedMsg.Parse(parsedBf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if parsedMsg.TargetCount != 2 {
		t.Errorf("TargetCount = %d, want 2", parsedMsg.TargetCount)
	}

	if len(parsedMsg.TargetCharIDs) != 2 {
		t.Errorf("len(TargetCharIDs) = %d, want 2", len(parsedMsg.TargetCharIDs))
	}
}

func TestMsgBinTargeted_BuildParseConsistency(t *testing.T) {
	original := &MsgBinTargeted{
		TargetCount:    3,
		TargetCharIDs:  []uint32{111, 222, 333},
		RawDataPayload: []byte{0x11, 0x22, 0x33, 0x44},
	}

	// First build
	bf1 := byteframe.NewByteFrame()
	err := original.Build(bf1)
	if err != nil {
		t.Fatalf("First Build() error = %v", err)
	}

	// Parse
	parsed := &MsgBinTargeted{}
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

func TestMsgBinTargeted_PayloadForwarding(t *testing.T) {
	// Test that RawDataPayload is correctly preserved
	// This is important as it forwards another binpacket
	originalPayload := []byte{
		0x10, 0x20, 0x30, 0x40, 0x50, 0x60, 0x70, 0x80,
		0x90, 0xA0, 0xB0, 0xC0, 0xD0, 0xE0, 0xF0, 0xFF,
	}

	msg := &MsgBinTargeted{
		TargetCount:    1,
		TargetCharIDs:  []uint32{999},
		RawDataPayload: originalPayload,
	}

	bf := byteframe.NewByteFrame()
	err := msg.Build(bf)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	parsed := &MsgBinTargeted{}
	parsedBf := byteframe.NewByteFrameFromBytes(bf.Data())
	err = parsed.Parse(parsedBf)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if !bytes.Equal(parsed.RawDataPayload, originalPayload) {
		t.Errorf("Payload not preserved:\ngot:  %v\nwant: %v", parsed.RawDataPayload, originalPayload)
	}
}
