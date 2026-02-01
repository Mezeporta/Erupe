package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

func TestAcquirePacketOpcodes(t *testing.T) {
	tests := []struct {
		name   string
		pkt    MHFPacket
		expect network.PacketID
	}{
		{"MsgMhfAcquireGuildTresure", &MsgMhfAcquireGuildTresure{}, network.MSG_MHF_ACQUIRE_GUILD_TRESURE},
		{"MsgMhfAcquireTitle", &MsgMhfAcquireTitle{}, network.MSG_MHF_ACQUIRE_TITLE},
		{"MsgMhfAcquireDistItem", &MsgMhfAcquireDistItem{}, network.MSG_MHF_ACQUIRE_DIST_ITEM},
		{"MsgMhfAcquireMonthlyItem", &MsgMhfAcquireMonthlyItem{}, network.MSG_MHF_ACQUIRE_MONTHLY_ITEM},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pkt.Opcode(); got != tt.expect {
				t.Errorf("Opcode() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestMsgMhfAcquireGuildTresureParse(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		huntID    uint32
		unk       uint8
	}{
		{"basic acquisition", 1, 12345, 0},
		{"large hunt ID", 0xABCDEF12, 0xFFFFFFFF, 1},
		{"zero values", 0, 0, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.huntID)
			bf.WriteUint8(tt.unk)
			bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfAcquireGuildTresure{}
			err := pkt.Parse(bf, &clientctx.ClientContext{})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.ackHandle)
			}
			if pkt.HuntID != tt.huntID {
				t.Errorf("HuntID = %d, want %d", pkt.HuntID, tt.huntID)
			}
			if pkt.Unk != tt.unk {
				t.Errorf("Unk = %d, want %d", pkt.Unk, tt.unk)
			}
		})
	}
}

func TestMsgMhfAcquireTitleParse(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		unk0      uint16
		unk1      uint16
		titleID   uint16
	}{
		{"acquire title 1", 1, 0, 0, 1},
		{"acquire title 100", 0x12345678, 10, 20, 100},
		{"max title ID", 0xFFFFFFFF, 0xFFFF, 0xFFFF, 0xFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint16(tt.unk0)
			bf.WriteUint16(tt.unk1)
			bf.WriteUint16(tt.titleID)
			bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfAcquireTitle{}
			err := pkt.Parse(bf, &clientctx.ClientContext{})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.ackHandle)
			}
			if pkt.Unk0 != tt.unk0 {
				t.Errorf("Unk0 = %d, want %d", pkt.Unk0, tt.unk0)
			}
			if pkt.Unk1 != tt.unk1 {
				t.Errorf("Unk1 = %d, want %d", pkt.Unk1, tt.unk1)
			}
			if pkt.TitleID != tt.titleID {
				t.Errorf("TitleID = %d, want %d", pkt.TitleID, tt.titleID)
			}
		})
	}
}

func TestMsgMhfAcquireDistItemParse(t *testing.T) {
	tests := []struct {
		name             string
		ackHandle        uint32
		distributionType uint8
		distributionID   uint32
	}{
		{"type 0", 1, 0, 12345},
		{"type 1", 0xABCD, 1, 67890},
		{"max values", 0xFFFFFFFF, 0xFF, 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint8(tt.distributionType)
			bf.WriteUint32(tt.distributionID)
			bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfAcquireDistItem{}
			err := pkt.Parse(bf, &clientctx.ClientContext{})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.ackHandle)
			}
			if pkt.DistributionType != tt.distributionType {
				t.Errorf("DistributionType = %d, want %d", pkt.DistributionType, tt.distributionType)
			}
			if pkt.DistributionID != tt.distributionID {
				t.Errorf("DistributionID = %d, want %d", pkt.DistributionID, tt.distributionID)
			}
		})
	}
}

func TestMsgMhfAcquireMonthlyItemParse(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		unk0      uint16
		unk1      uint16
		unk2      uint32
		unk3      uint32
	}{
		{"basic", 1, 0, 0, 0, 0},
		{"with values", 100, 10, 20, 30, 40},
		{"max values", 0xFFFFFFFF, 0xFFFF, 0xFFFF, 0xFFFFFFFF, 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint16(tt.unk0)
			bf.WriteUint16(tt.unk1)
			bf.WriteUint32(tt.unk2)
			bf.WriteUint32(tt.unk3)
			bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfAcquireMonthlyItem{}
			err := pkt.Parse(bf, &clientctx.ClientContext{})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.ackHandle)
			}
			if pkt.Unk0 != tt.unk0 {
				t.Errorf("Unk0 = %d, want %d", pkt.Unk0, tt.unk0)
			}
			if pkt.Unk1 != tt.unk1 {
				t.Errorf("Unk1 = %d, want %d", pkt.Unk1, tt.unk1)
			}
			if pkt.Unk2 != tt.unk2 {
				t.Errorf("Unk2 = %d, want %d", pkt.Unk2, tt.unk2)
			}
			if pkt.Unk3 != tt.unk3 {
				t.Errorf("Unk3 = %d, want %d", pkt.Unk3, tt.unk3)
			}
		})
	}
}

func TestAcquirePacketsFromOpcode(t *testing.T) {
	acquireOpcodes := []network.PacketID{
		network.MSG_MHF_ACQUIRE_GUILD_TRESURE,
		network.MSG_MHF_ACQUIRE_TITLE,
		network.MSG_MHF_ACQUIRE_DIST_ITEM,
		network.MSG_MHF_ACQUIRE_MONTHLY_ITEM,
	}

	for _, opcode := range acquireOpcodes {
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

func TestAcquirePacketEdgeCases(t *testing.T) {
	t.Run("guild tresure with max hunt ID", func(t *testing.T) {
		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)
		bf.WriteUint32(0xFFFFFFFF)
		bf.WriteUint8(255)
		bf.Seek(0, io.SeekStart)

		pkt := &MsgMhfAcquireGuildTresure{}
		err := pkt.Parse(bf, &clientctx.ClientContext{})
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if pkt.HuntID != 0xFFFFFFFF {
			t.Errorf("HuntID = %d, want %d", pkt.HuntID, 0xFFFFFFFF)
		}
	})

	t.Run("dist item with all types", func(t *testing.T) {
		for i := uint8(0); i < 5; i++ {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(1)
			bf.WriteUint8(i)
			bf.WriteUint32(12345)
			bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfAcquireDistItem{}
			err := pkt.Parse(bf, &clientctx.ClientContext{})
			if err != nil {
				t.Fatalf("Parse() error = %v for type %d", err, i)
			}

			if pkt.DistributionType != i {
				t.Errorf("DistributionType = %d, want %d", pkt.DistributionType, i)
			}
		}
	})
}
