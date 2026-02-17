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
		unk       bool
	}{
		{"basic acquisition", 1, 12345, false},
		{"large hunt ID", 0xABCDEF12, 0xFFFFFFFF, true},
		{"zero values", 0, 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint32(tt.huntID)
			bf.WriteBool(tt.unk)
			_, _ = bf.Seek(0, io.SeekStart)

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
				t.Errorf("Unk = %v, want %v", pkt.Unk, tt.unk)
			}
		})
	}
}

func TestMsgMhfAcquireTitleParse(t *testing.T) {
	tests := []struct {
		name      string
		ackHandle uint32
		titleIDs  []uint16
	}{
		{"acquire title 1", 1, []uint16{1}},
		{"acquire titles 100 200", 0x12345678, []uint16{100, 200}},
		{"no titles", 0xFFFFFFFF, []uint16{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint16(uint16(len(tt.titleIDs))) // count
			bf.WriteUint16(0)                         // zeroed
			for _, id := range tt.titleIDs {
				bf.WriteUint16(id)
			}
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgMhfAcquireTitle{}
			err := pkt.Parse(bf, &clientctx.ClientContext{})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.ackHandle)
			}
			if len(pkt.TitleIDs) != len(tt.titleIDs) {
				t.Errorf("TitleIDs len = %d, want %d", len(pkt.TitleIDs), len(tt.titleIDs))
			}
			for i, id := range tt.titleIDs {
				if i < len(pkt.TitleIDs) && pkt.TitleIDs[i] != id {
					t.Errorf("TitleIDs[%d] = %d, want %d", i, pkt.TitleIDs[i], id)
				}
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
			_, _ = bf.Seek(0, io.SeekStart)

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
		unk0      uint8
		unk1      uint8
		unk2      uint16
		unk3      uint32
	}{
		{"basic", 1, 0, 0, 0, 0},
		{"with values", 100, 10, 20, 30, 40},
		{"max values", 0xFFFFFFFF, 0xFF, 0xFF, 0xFFFF, 0xFFFFFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint8(tt.unk0)
			bf.WriteUint8(tt.unk1)
			bf.WriteUint16(tt.unk2)
			bf.WriteUint32(tt.unk3)
			bf.WriteUint32(0) // Zeroed (consumed by Parse)
			_, _ = bf.Seek(0, io.SeekStart)

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
		bf.WriteBool(true)
		_, _ = bf.Seek(0, io.SeekStart)

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
			_, _ = bf.Seek(0, io.SeekStart)

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
