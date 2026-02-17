package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

func TestMsgSysAckRoundTrip(t *testing.T) {
	tests := []struct {
		name             string
		ackHandle        uint32
		isBufferResponse bool
		errorCode        uint8
		ackData          []byte
	}{
		{
			name:             "simple non-buffer response",
			ackHandle:        1,
			isBufferResponse: false,
			errorCode:        0,
			ackData:          []byte{0x00, 0x00, 0x00, 0x00},
		},
		{
			name:             "buffer response with small data",
			ackHandle:        0x12345678,
			isBufferResponse: true,
			errorCode:        0,
			ackData:          []byte{0x01, 0x02, 0x03, 0x04, 0x05},
		},
		{
			name:             "error response",
			ackHandle:        100,
			isBufferResponse: false,
			errorCode:        1,
			ackData:          []byte{0xDE, 0xAD, 0xBE, 0xEF},
		},
		{
			name:             "empty buffer response",
			ackHandle:        999,
			isBufferResponse: true,
			errorCode:        0,
			ackData:          []byte{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysAck{
				AckHandle:        tt.ackHandle,
				IsBufferResponse: tt.isBufferResponse,
				ErrorCode:        tt.errorCode,
				AckData:          tt.ackData,
			}
			ctx := &clientctx.ClientContext{}

			// Build
			bf := byteframe.NewByteFrame()
			err := original.Build(bf, ctx)
			if err != nil {
				t.Fatalf("Build() error = %v", err)
			}

			// Parse
			bf.Seek(0, io.SeekStart)
			parsed := &MsgSysAck{}
			err = parsed.Parse(bf, ctx)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Compare
			if parsed.AckHandle != original.AckHandle {
				t.Errorf("AckHandle = %d, want %d", parsed.AckHandle, original.AckHandle)
			}
			if parsed.IsBufferResponse != original.IsBufferResponse {
				t.Errorf("IsBufferResponse = %v, want %v", parsed.IsBufferResponse, original.IsBufferResponse)
			}
			if parsed.ErrorCode != original.ErrorCode {
				t.Errorf("ErrorCode = %d, want %d", parsed.ErrorCode, original.ErrorCode)
			}
		})
	}
}

func TestMsgSysAckLargePayload(t *testing.T) {
	// Test with payload larger than 0xFFFF to trigger extended size field
	largeData := make([]byte, 0x10000) // 65536 bytes
	for i := range largeData {
		largeData[i] = byte(i % 256)
	}

	original := &MsgSysAck{
		AckHandle:        1,
		IsBufferResponse: true,
		ErrorCode:        0,
		AckData:          largeData,
	}
	ctx := &clientctx.ClientContext{}

	// Build
	bf := byteframe.NewByteFrame()
	err := original.Build(bf, ctx)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Parse
	bf.Seek(0, io.SeekStart)
	parsed := &MsgSysAck{}
	err = parsed.Parse(bf, ctx)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	if len(parsed.AckData) != len(largeData) {
		t.Errorf("AckData len = %d, want %d", len(parsed.AckData), len(largeData))
	}
}

func TestMsgSysAckOpcode(t *testing.T) {
	pkt := &MsgSysAck{}
	if pkt.Opcode() != network.MSG_SYS_ACK {
		t.Errorf("Opcode() = %s, want MSG_SYS_ACK", pkt.Opcode())
	}
}

func TestMsgSysNopRoundTrip(t *testing.T) {
	original := &MsgSysNop{}
	ctx := &clientctx.ClientContext{}

	// Build
	bf := byteframe.NewByteFrame()
	err := original.Build(bf, ctx)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Should write no data
	if len(bf.Data()) != 0 {
		t.Errorf("MsgSysNop.Build() wrote %d bytes, want 0", len(bf.Data()))
	}

	// Parse (from empty buffer)
	parsed := &MsgSysNop{}
	err = parsed.Parse(bf, ctx)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestMsgSysNopOpcode(t *testing.T) {
	pkt := &MsgSysNop{}
	if pkt.Opcode() != network.MSG_SYS_NOP {
		t.Errorf("Opcode() = %s, want MSG_SYS_NOP", pkt.Opcode())
	}
}

func TestMsgSysEndRoundTrip(t *testing.T) {
	original := &MsgSysEnd{}
	ctx := &clientctx.ClientContext{}

	// Build
	bf := byteframe.NewByteFrame()
	err := original.Build(bf, ctx)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	// Should write no data
	if len(bf.Data()) != 0 {
		t.Errorf("MsgSysEnd.Build() wrote %d bytes, want 0", len(bf.Data()))
	}

	// Parse (from empty buffer)
	parsed := &MsgSysEnd{}
	err = parsed.Parse(bf, ctx)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}
}

func TestMsgSysEndOpcode(t *testing.T) {
	pkt := &MsgSysEnd{}
	if pkt.Opcode() != network.MSG_SYS_END {
		t.Errorf("Opcode() = %s, want MSG_SYS_END", pkt.Opcode())
	}
}

func TestMsgSysAckNonBufferResponse(t *testing.T) {
	// Non-buffer response should always read/write 4 bytes of data
	original := &MsgSysAck{
		AckHandle:        1,
		IsBufferResponse: false,
		ErrorCode:        0,
		AckData:          []byte{0xAA, 0xBB, 0xCC, 0xDD},
	}
	ctx := &clientctx.ClientContext{}

	bf := byteframe.NewByteFrame()
	err := original.Build(bf, ctx)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	bf.Seek(0, io.SeekStart)
	parsed := &MsgSysAck{}
	err = parsed.Parse(bf, ctx)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Non-buffer response should have exactly 4 bytes of data
	if len(parsed.AckData) != 4 {
		t.Errorf("Non-buffer AckData len = %d, want 4", len(parsed.AckData))
	}
}

func TestMsgSysAckNonBufferShortData(t *testing.T) {
	// Non-buffer response with short data should pad to 4 bytes
	original := &MsgSysAck{
		AckHandle:        1,
		IsBufferResponse: false,
		ErrorCode:        0,
		AckData:          []byte{0x01}, // Only 1 byte
	}
	ctx := &clientctx.ClientContext{}

	bf := byteframe.NewByteFrame()
	err := original.Build(bf, ctx)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	bf.Seek(0, io.SeekStart)
	parsed := &MsgSysAck{}
	err = parsed.Parse(bf, ctx)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Should still read 4 bytes
	if len(parsed.AckData) != 4 {
		t.Errorf("AckData len = %d, want 4", len(parsed.AckData))
	}
}

func TestMsgSysAckBuildFormat(t *testing.T) {
	pkt := &MsgSysAck{
		AckHandle:        0x12345678,
		IsBufferResponse: true,
		ErrorCode:        0x55,
		AckData:          []byte{0xAA, 0xBB},
	}
	ctx := &clientctx.ClientContext{}

	bf := byteframe.NewByteFrame()
	_ = pkt.Build(bf, ctx)

	data := bf.Data()

	// Check AckHandle (big-endian)
	if data[0] != 0x12 || data[1] != 0x34 || data[2] != 0x56 || data[3] != 0x78 {
		t.Errorf("AckHandle bytes = %X, want 12345678", data[:4])
	}

	// Check IsBufferResponse (1 = true)
	if data[4] != 1 {
		t.Errorf("IsBufferResponse byte = %d, want 1", data[4])
	}

	// Check ErrorCode
	if data[5] != 0x55 {
		t.Errorf("ErrorCode byte = %X, want 55", data[5])
	}

	// Check payload size (2 bytes, big-endian)
	if data[6] != 0x00 || data[7] != 0x02 {
		t.Errorf("PayloadSize bytes = %X %X, want 00 02", data[6], data[7])
	}

	// Check actual data
	if data[8] != 0xAA || data[9] != 0xBB {
		t.Errorf("AckData bytes = %X %X, want AA BB", data[8], data[9])
	}
}

func TestCorePacketsFromOpcode(t *testing.T) {
	coreOpcodes := []network.PacketID{
		network.MSG_SYS_NOP,
		network.MSG_SYS_END,
		network.MSG_SYS_ACK,
		network.MSG_SYS_PING,
	}

	for _, opcode := range coreOpcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Fatalf("FromOpcode(%s) returned nil", opcode)
			}
			if pkt.Opcode() != opcode {
				t.Errorf("Opcode() = %s, want %s", pkt.Opcode(), opcode)
			}
		})
	}
}
