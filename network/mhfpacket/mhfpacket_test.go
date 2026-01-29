package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

func TestMHFPacketInterface(t *testing.T) {
	// Verify that packets implement the MHFPacket interface
	var _ MHFPacket = &MsgSysPing{}
	var _ MHFPacket = &MsgSysTime{}
	var _ MHFPacket = &MsgSysNop{}
	var _ MHFPacket = &MsgSysEnd{}
	var _ MHFPacket = &MsgSysLogin{}
	var _ MHFPacket = &MsgSysLogout{}
}

func TestFromOpcodeReturnsCorrectType(t *testing.T) {
	tests := []struct {
		opcode   network.PacketID
		wantType string
	}{
		{network.MSG_HEAD, "*mhfpacket.MsgHead"},
		{network.MSG_SYS_PING, "*mhfpacket.MsgSysPing"},
		{network.MSG_SYS_TIME, "*mhfpacket.MsgSysTime"},
		{network.MSG_SYS_NOP, "*mhfpacket.MsgSysNop"},
		{network.MSG_SYS_END, "*mhfpacket.MsgSysEnd"},
		{network.MSG_SYS_ACK, "*mhfpacket.MsgSysAck"},
		{network.MSG_SYS_LOGIN, "*mhfpacket.MsgSysLogin"},
		{network.MSG_SYS_LOGOUT, "*mhfpacket.MsgSysLogout"},
		{network.MSG_SYS_CREATE_STAGE, "*mhfpacket.MsgSysCreateStage"},
		{network.MSG_SYS_ENTER_STAGE, "*mhfpacket.MsgSysEnterStage"},
	}

	for _, tt := range tests {
		t.Run(tt.opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(tt.opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", tt.opcode)
				return
			}
			if pkt.Opcode() != tt.opcode {
				t.Errorf("Opcode() = %s, want %s", pkt.Opcode(), tt.opcode)
			}
		})
	}
}

func TestFromOpcodeUnknown(t *testing.T) {
	// Test with an invalid opcode
	pkt := FromOpcode(network.PacketID(0xFFFF))
	if pkt != nil {
		t.Error("FromOpcode(0xFFFF) should return nil for unknown opcode")
	}
}

func TestMsgSysPingRoundTrip(t *testing.T) {
	original := &MsgSysPing{
		AckHandle: 0x12345678,
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
	parsed := &MsgSysPing{}
	err = parsed.Parse(bf, ctx)
	if err != nil {
		t.Fatalf("Parse() error = %v", err)
	}

	// Compare
	if parsed.AckHandle != original.AckHandle {
		t.Errorf("AckHandle = %d, want %d", parsed.AckHandle, original.AckHandle)
	}
}

func TestMsgSysTimeRoundTrip(t *testing.T) {
	tests := []struct {
		name          string
		getRemoteTime bool
		timestamp     uint32
	}{
		{"no remote time", false, 1577105879},
		{"with remote time", true, 1609459200},
		{"zero timestamp", false, 0},
		{"max timestamp", true, 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			original := &MsgSysTime{
				GetRemoteTime: tt.getRemoteTime,
				Timestamp:     tt.timestamp,
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
			parsed := &MsgSysTime{}
			err = parsed.Parse(bf, ctx)
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			// Compare
			if parsed.GetRemoteTime != original.GetRemoteTime {
				t.Errorf("GetRemoteTime = %v, want %v", parsed.GetRemoteTime, original.GetRemoteTime)
			}
			if parsed.Timestamp != original.Timestamp {
				t.Errorf("Timestamp = %d, want %d", parsed.Timestamp, original.Timestamp)
			}
		})
	}
}

func TestMsgSysPingOpcode(t *testing.T) {
	pkt := &MsgSysPing{}
	if pkt.Opcode() != network.MSG_SYS_PING {
		t.Errorf("Opcode() = %s, want MSG_SYS_PING", pkt.Opcode())
	}
}

func TestMsgSysTimeOpcode(t *testing.T) {
	pkt := &MsgSysTime{}
	if pkt.Opcode() != network.MSG_SYS_TIME {
		t.Errorf("Opcode() = %s, want MSG_SYS_TIME", pkt.Opcode())
	}
}

func TestFromOpcodeSystemPackets(t *testing.T) {
	// Test all system packet opcodes return non-nil
	systemOpcodes := []network.PacketID{
		network.MSG_SYS_reserve01,
		network.MSG_SYS_reserve02,
		network.MSG_SYS_reserve03,
		network.MSG_SYS_reserve04,
		network.MSG_SYS_reserve05,
		network.MSG_SYS_reserve06,
		network.MSG_SYS_reserve07,
		network.MSG_SYS_ADD_OBJECT,
		network.MSG_SYS_DEL_OBJECT,
		network.MSG_SYS_DISP_OBJECT,
		network.MSG_SYS_HIDE_OBJECT,
		network.MSG_SYS_END,
		network.MSG_SYS_NOP,
		network.MSG_SYS_ACK,
		network.MSG_SYS_LOGIN,
		network.MSG_SYS_LOGOUT,
		network.MSG_SYS_SET_STATUS,
		network.MSG_SYS_PING,
		network.MSG_SYS_TIME,
	}

	for _, opcode := range systemOpcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", opcode)
			}
		})
	}
}

func TestFromOpcodeStagePackets(t *testing.T) {
	stageOpcodes := []network.PacketID{
		network.MSG_SYS_CREATE_STAGE,
		network.MSG_SYS_STAGE_DESTRUCT,
		network.MSG_SYS_ENTER_STAGE,
		network.MSG_SYS_BACK_STAGE,
		network.MSG_SYS_MOVE_STAGE,
		network.MSG_SYS_LEAVE_STAGE,
		network.MSG_SYS_LOCK_STAGE,
		network.MSG_SYS_UNLOCK_STAGE,
		network.MSG_SYS_RESERVE_STAGE,
		network.MSG_SYS_UNRESERVE_STAGE,
		network.MSG_SYS_SET_STAGE_PASS,
	}

	for _, opcode := range stageOpcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", opcode)
			}
		})
	}
}

func TestOpcodeMatches(t *testing.T) {
	// Verify that packets return the same opcode they were created from
	tests := []network.PacketID{
		network.MSG_HEAD,
		network.MSG_SYS_PING,
		network.MSG_SYS_TIME,
		network.MSG_SYS_END,
		network.MSG_SYS_NOP,
		network.MSG_SYS_ACK,
		network.MSG_SYS_LOGIN,
		network.MSG_SYS_CREATE_STAGE,
	}

	for _, opcode := range tests {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Skip("opcode not implemented")
			}
			if pkt.Opcode() != opcode {
				t.Errorf("Opcode() = %s, want %s", pkt.Opcode(), opcode)
			}
		})
	}
}

func TestParserInterface(t *testing.T) {
	// Verify Parser interface works
	var p Parser = &MsgSysPing{}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(123)
	bf.Seek(0, io.SeekStart)

	err := p.Parse(bf, &clientctx.ClientContext{})
	if err != nil {
		t.Errorf("Parse() error = %v", err)
	}
}

func TestBuilderInterface(t *testing.T) {
	// Verify Builder interface works
	var b Builder = &MsgSysPing{AckHandle: 456}
	bf := byteframe.NewByteFrame()

	err := b.Build(bf, &clientctx.ClientContext{})
	if err != nil {
		t.Errorf("Build() error = %v", err)
	}
	if len(bf.Data()) == 0 {
		t.Error("Build() should write data")
	}
}

func TestOpcoderInterface(t *testing.T) {
	// Verify Opcoder interface works
	var o Opcoder = &MsgSysPing{}
	opcode := o.Opcode()

	if opcode != network.MSG_SYS_PING {
		t.Errorf("Opcode() = %s, want MSG_SYS_PING", opcode)
	}
}

func TestClientContextNilSafe(t *testing.T) {
	// Some packets may need to handle nil ClientContext
	pkt := &MsgSysPing{AckHandle: 123}
	bf := byteframe.NewByteFrame()

	// This should not panic even with nil context (implementation dependent)
	// Note: The actual behavior depends on implementation
	err := pkt.Build(bf, nil)
	if err != nil {
		// Error is acceptable if nil context is not supported
		t.Logf("Build() with nil context returned error: %v", err)
	}
}

func TestMsgSysPingBuildFormat(t *testing.T) {
	pkt := &MsgSysPing{AckHandle: 0x12345678}
	bf := byteframe.NewByteFrame()
	pkt.Build(bf, &clientctx.ClientContext{})

	data := bf.Data()
	if len(data) != 4 {
		t.Errorf("Build() data len = %d, want 4", len(data))
	}

	// Verify big-endian format (default)
	if data[0] != 0x12 || data[1] != 0x34 || data[2] != 0x56 || data[3] != 0x78 {
		t.Errorf("Build() data = %x, want 12345678", data)
	}
}

func TestMsgSysTimeBuildFormat(t *testing.T) {
	pkt := &MsgSysTime{
		GetRemoteTime: true,
		Timestamp:     0xDEADBEEF,
	}
	bf := byteframe.NewByteFrame()
	pkt.Build(bf, &clientctx.ClientContext{})

	data := bf.Data()
	if len(data) != 5 {
		t.Errorf("Build() data len = %d, want 5 (1 bool + 4 uint32)", len(data))
	}

	// First byte is bool (1 = true)
	if data[0] != 1 {
		t.Errorf("GetRemoteTime byte = %d, want 1", data[0])
	}
}

func TestMsgSysNop(t *testing.T) {
	pkt := FromOpcode(network.MSG_SYS_NOP)
	if pkt == nil {
		t.Fatal("FromOpcode(MSG_SYS_NOP) returned nil")
	}
	if pkt.Opcode() != network.MSG_SYS_NOP {
		t.Errorf("Opcode() = %s, want MSG_SYS_NOP", pkt.Opcode())
	}
}

func TestMsgSysEnd(t *testing.T) {
	pkt := FromOpcode(network.MSG_SYS_END)
	if pkt == nil {
		t.Fatal("FromOpcode(MSG_SYS_END) returned nil")
	}
	if pkt.Opcode() != network.MSG_SYS_END {
		t.Errorf("Opcode() = %s, want MSG_SYS_END", pkt.Opcode())
	}
}

func TestMsgHead(t *testing.T) {
	pkt := FromOpcode(network.MSG_HEAD)
	if pkt == nil {
		t.Fatal("FromOpcode(MSG_HEAD) returned nil")
	}
	if pkt.Opcode() != network.MSG_HEAD {
		t.Errorf("Opcode() = %s, want MSG_HEAD", pkt.Opcode())
	}
}

func TestMsgSysAck(t *testing.T) {
	pkt := FromOpcode(network.MSG_SYS_ACK)
	if pkt == nil {
		t.Fatal("FromOpcode(MSG_SYS_ACK) returned nil")
	}
	if pkt.Opcode() != network.MSG_SYS_ACK {
		t.Errorf("Opcode() = %s, want MSG_SYS_ACK", pkt.Opcode())
	}
}

func TestBinaryPackets(t *testing.T) {
	binaryOpcodes := []network.PacketID{
		network.MSG_SYS_CAST_BINARY,
		network.MSG_SYS_CASTED_BINARY,
		network.MSG_SYS_SET_STAGE_BINARY,
		network.MSG_SYS_GET_STAGE_BINARY,
		network.MSG_SYS_WAIT_STAGE_BINARY,
	}

	for _, opcode := range binaryOpcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", opcode)
			}
		})
	}
}

func TestEnumeratePackets(t *testing.T) {
	enumOpcodes := []network.PacketID{
		network.MSG_SYS_ENUMERATE_CLIENT,
		network.MSG_SYS_ENUMERATE_STAGE,
	}

	for _, opcode := range enumOpcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", opcode)
			}
		})
	}
}

func TestSemaphorePackets(t *testing.T) {
	semaOpcodes := []network.PacketID{
		network.MSG_SYS_CREATE_ACQUIRE_SEMAPHORE,
		network.MSG_SYS_ACQUIRE_SEMAPHORE,
		network.MSG_SYS_RELEASE_SEMAPHORE,
		network.MSG_SYS_CHECK_SEMAPHORE,
	}

	for _, opcode := range semaOpcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", opcode)
			}
		})
	}
}

func TestObjectPackets(t *testing.T) {
	objOpcodes := []network.PacketID{
		network.MSG_SYS_ADD_OBJECT,
		network.MSG_SYS_DEL_OBJECT,
		network.MSG_SYS_DISP_OBJECT,
		network.MSG_SYS_HIDE_OBJECT,
	}

	for _, opcode := range objOpcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", opcode)
			}
		})
	}
}

func TestLogPackets(t *testing.T) {
	logOpcodes := []network.PacketID{
		network.MSG_SYS_TERMINAL_LOG,
		network.MSG_SYS_ISSUE_LOGKEY,
		network.MSG_SYS_RECORD_LOG,
	}

	for _, opcode := range logOpcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", opcode)
			}
		})
	}
}

func TestMHFSaveLoad(t *testing.T) {
	saveLoadOpcodes := []network.PacketID{
		network.MSG_MHF_SAVEDATA,
		network.MSG_MHF_LOADDATA,
	}

	for _, opcode := range saveLoadOpcodes {
		t.Run(opcode.String(), func(t *testing.T) {
			pkt := FromOpcode(opcode)
			if pkt == nil {
				t.Errorf("FromOpcode(%s) returned nil", opcode)
			}
		})
	}
}
