package mhfpacket

import (
	"io"
	"testing"

	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
)

func TestStagePacketOpcodes(t *testing.T) {
	tests := []struct {
		name   string
		pkt    MHFPacket
		expect network.PacketID
	}{
		{"MsgSysCreateStage", &MsgSysCreateStage{}, network.MSG_SYS_CREATE_STAGE},
		{"MsgSysEnterStage", &MsgSysEnterStage{}, network.MSG_SYS_ENTER_STAGE},
		{"MsgSysMoveStage", &MsgSysMoveStage{}, network.MSG_SYS_MOVE_STAGE},
		{"MsgSysBackStage", &MsgSysBackStage{}, network.MSG_SYS_BACK_STAGE},
		{"MsgSysLockStage", &MsgSysLockStage{}, network.MSG_SYS_LOCK_STAGE},
		{"MsgSysUnlockStage", &MsgSysUnlockStage{}, network.MSG_SYS_UNLOCK_STAGE},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.pkt.Opcode(); got != tt.expect {
				t.Errorf("Opcode() = %v, want %v", got, tt.expect)
			}
		})
	}
}

func TestMsgSysCreateStageFields(t *testing.T) {
	tests := []struct {
		name        string
		ackHandle   uint32
		unk0        uint8
		playerCount uint8
		stageID     string
	}{
		{"empty stage", 1, 1, 4, ""},
		{"mezeporta", 0x12345678, 2, 8, "sl1Ns200p0a0u0"},
		{"quest room", 100, 1, 4, "q1234"},
		{"max players", 0xFFFFFFFF, 2, 16, "max_stage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.ackHandle)
			bf.WriteUint8(tt.unk0)
			bf.WriteUint8(tt.playerCount)
			stageIDBytes := []byte(tt.stageID)
			bf.WriteUint8(uint8(len(stageIDBytes)))
			bf.WriteBytes(append(stageIDBytes, 0x00))
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgSysCreateStage{}
			err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.ackHandle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.ackHandle)
			}
			if pkt.CreateType != tt.unk0 {
				t.Errorf("CreateType = %d, want %d", pkt.CreateType, tt.unk0)
			}
			if pkt.PlayerCount != tt.playerCount {
				t.Errorf("PlayerCount = %d, want %d", pkt.PlayerCount, tt.playerCount)
			}
			if pkt.StageID != tt.stageID {
				t.Errorf("StageID = %q, want %q", pkt.StageID, tt.stageID)
			}
		})
	}
}

func TestMsgSysEnterStageFields(t *testing.T) {
	tests := []struct {
		name    string
		handle  uint32
		unk     bool
		stageID string
	}{
		{"enter town", 1, false, "town01"},
		{"force enter", 2, true, "quest_stage"},
		{"rasta bar", 999, false, "sl1Ns211p0a0u0"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.handle)
			bf.WriteBool(tt.unk)
			stageIDBytes := []byte(tt.stageID)
			bf.WriteUint8(uint8(len(stageIDBytes)))
			bf.WriteBytes(append(stageIDBytes, 0x00))
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgSysEnterStage{}
			err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.handle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.handle)
			}
			if pkt.IsQuest != tt.unk {
				t.Errorf("Unk = %v, want %v", pkt.IsQuest, tt.unk)
			}
			if pkt.StageID != tt.stageID {
				t.Errorf("StageID = %q, want %q", pkt.StageID, tt.stageID)
			}
		})
	}
}

func TestMsgSysMoveStageFields(t *testing.T) {
	tests := []struct {
		name    string
		handle  uint32
		unkBool uint8
		stageID string
	}{
		{"move to area", 1, 0, "area01"},
		{"move to quest", 0xABCD, 1, "quest12345"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.handle)
			bf.WriteUint8(tt.unkBool)
			stageIDBytes := []byte(tt.stageID)
			bf.WriteUint8(uint8(len(stageIDBytes)))
			bf.WriteBytes(stageIDBytes)
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgSysMoveStage{}
			err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.handle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.handle)
			}
			if pkt.UnkBool != tt.unkBool {
				t.Errorf("UnkBool = %d, want %d", pkt.UnkBool, tt.unkBool)
			}
			if pkt.StageID != tt.stageID {
				t.Errorf("StageID = %q, want %q", pkt.StageID, tt.stageID)
			}
		})
	}
}

func TestMsgSysLockStageFields(t *testing.T) {
	tests := []struct {
		name    string
		handle  uint32
		unk0    uint8
		unk1    uint8
		stageID string
	}{
		{"lock room", 1, 1, 1, "room01"},
		{"private party", 0x1234, 1, 1, "party_stage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.handle)
			bf.WriteUint8(tt.unk0)
			bf.WriteUint8(tt.unk1)
			stageIDBytes := []byte(tt.stageID)
			bf.WriteUint8(uint8(len(stageIDBytes)))
			bf.WriteBytes(append(stageIDBytes, 0x00))
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgSysLockStage{}
			err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.handle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.handle)
			}
			// Unk0 and Unk1 are read but discarded by Parse, so we only verify
			// that Parse consumed the bytes without error
			if pkt.StageID != tt.stageID {
				t.Errorf("StageID = %q, want %q", pkt.StageID, tt.stageID)
			}
		})
	}
}

func TestMsgSysUnlockStageFields(t *testing.T) {
	tests := []struct {
		name string
		unk0 uint16
	}{
		{"zero", 0},
		{"typical", 1},
		{"max", 0xFFFF},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint16(tt.unk0)
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgSysUnlockStage{}
			err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}
			// MsgSysUnlockStage is an empty struct; Parse reads and discards a uint16.
			// We just verify Parse doesn't error.
		})
	}
}

func TestMsgSysBackStageFields(t *testing.T) {
	tests := []struct {
		name   string
		handle uint32
	}{
		{"small handle", 1},
		{"large handle", 0xDEADBEEF},
		{"zero", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			bf := byteframe.NewByteFrame()
			bf.WriteUint32(tt.handle)
			_, _ = bf.Seek(0, io.SeekStart)

			pkt := &MsgSysBackStage{}
			err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
			if err != nil {
				t.Fatalf("Parse() error = %v", err)
			}

			if pkt.AckHandle != tt.handle {
				t.Errorf("AckHandle = %d, want %d", pkt.AckHandle, tt.handle)
			}
		})
	}
}

func TestStageIDEdgeCases(t *testing.T) {
	t.Run("long stage ID", func(t *testing.T) {
		// Stage ID with max length (255 bytes)
		longID := make([]byte, 200)
		for i := range longID {
			longID[i] = 'a' + byte(i%26)
		}

		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)
		bf.WriteUint8(1)
		bf.WriteUint8(4)
		bf.WriteUint8(uint8(len(longID)))
		bf.WriteBytes(append(longID, 0x00))
		_, _ = bf.Seek(0, io.SeekStart)

		pkt := &MsgSysCreateStage{}
		err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		if pkt.StageID != string(longID) {
			t.Errorf("StageID length = %d, want %d", len(pkt.StageID), len(longID))
		}
	})

	t.Run("stage ID with null terminator", func(t *testing.T) {
		// String terminated with null byte
		stageID := "test\x00extra"

		bf := byteframe.NewByteFrame()
		bf.WriteUint32(1)
		bf.WriteUint8(0)
		bf.WriteUint8(uint8(len(stageID)))
		bf.WriteBytes([]byte(stageID))
		_, _ = bf.Seek(0, io.SeekStart)

		pkt := &MsgSysEnterStage{}
		err := pkt.Parse(bf, &clientctx.ClientContext{RealClientMode: _config.ZZ})
		if err != nil {
			t.Fatalf("Parse() error = %v", err)
		}

		// Should truncate at null
		if pkt.StageID != "test" {
			t.Errorf("StageID = %q, want %q (should truncate at null)", pkt.StageID, "test")
		}
	})
}

func TestStagePacketFromOpcode(t *testing.T) {
	stageOpcodes := []network.PacketID{
		network.MSG_SYS_CREATE_STAGE,
		network.MSG_SYS_ENTER_STAGE,
		network.MSG_SYS_BACK_STAGE,
		network.MSG_SYS_MOVE_STAGE,
		network.MSG_SYS_LOCK_STAGE,
		network.MSG_SYS_UNLOCK_STAGE,
	}

	for _, opcode := range stageOpcodes {
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
