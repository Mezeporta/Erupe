package network

import (
	"testing"
)

func TestPacketIDType(t *testing.T) {
	// PacketID is based on uint16
	var p PacketID = 0xFFFF
	if uint16(p) != 0xFFFF {
		t.Errorf("PacketID max value = %d, want %d", uint16(p), 0xFFFF)
	}
}

func TestPacketIDConstants(t *testing.T) {
	// Test critical packet IDs are correct
	tests := []struct {
		name   string
		id     PacketID
		expect uint16
	}{
		{"MSG_HEAD", MSG_HEAD, 0},
		{"MSG_SYS_END", MSG_SYS_END, 0x10},
		{"MSG_SYS_NOP", MSG_SYS_NOP, 0x11},
		{"MSG_SYS_ACK", MSG_SYS_ACK, 0x12},
		{"MSG_SYS_LOGIN", MSG_SYS_LOGIN, 0x14},
		{"MSG_SYS_LOGOUT", MSG_SYS_LOGOUT, 0x15},
		{"MSG_SYS_PING", MSG_SYS_PING, 0x17},
		{"MSG_SYS_TIME", MSG_SYS_TIME, 0x1A},
		{"MSG_SYS_CREATE_STAGE", MSG_SYS_CREATE_STAGE, 0x20},
		{"MSG_SYS_ENTER_STAGE", MSG_SYS_ENTER_STAGE, 0x22},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if uint16(tt.id) != tt.expect {
				t.Errorf("%s = 0x%X, want 0x%X", tt.name, uint16(tt.id), tt.expect)
			}
		})
	}
}

func TestPacketIDString(t *testing.T) {
	// Test that String() method works for known packet IDs
	tests := []struct {
		id       PacketID
		contains string
	}{
		{MSG_HEAD, "MSG_HEAD"},
		{MSG_SYS_PING, "MSG_SYS_PING"},
		{MSG_SYS_END, "MSG_SYS_END"},
		{MSG_SYS_NOP, "MSG_SYS_NOP"},
		{MSG_SYS_ACK, "MSG_SYS_ACK"},
		{MSG_SYS_LOGIN, "MSG_SYS_LOGIN"},
		{MSG_SYS_LOGOUT, "MSG_SYS_LOGOUT"},
	}

	for _, tt := range tests {
		t.Run(tt.contains, func(t *testing.T) {
			got := tt.id.String()
			if got != tt.contains {
				t.Errorf("String() = %q, want %q", got, tt.contains)
			}
		})
	}
}

func TestPacketIDUnknown(t *testing.T) {
	// Unknown packet ID should still have a valid string representation
	unknown := PacketID(0xFFFF)
	str := unknown.String()
	if str == "" {
		t.Error("String() for unknown PacketID should not be empty")
	}
}

func TestPacketIDZero(t *testing.T) {
	// MSG_HEAD should be 0
	if MSG_HEAD != 0 {
		t.Errorf("MSG_HEAD = %d, want 0", MSG_HEAD)
	}
}

func TestSystemPacketIDRange(t *testing.T) {
	// System packets should be in a specific range
	systemPackets := []PacketID{
		MSG_SYS_reserve01,
		MSG_SYS_reserve02,
		MSG_SYS_reserve03,
		MSG_SYS_ADD_OBJECT,
		MSG_SYS_DEL_OBJECT,
		MSG_SYS_END,
		MSG_SYS_NOP,
		MSG_SYS_ACK,
		MSG_SYS_LOGIN,
		MSG_SYS_LOGOUT,
		MSG_SYS_PING,
		MSG_SYS_TIME,
	}

	for _, pkt := range systemPackets {
		// System packets should have IDs > 0 (MSG_HEAD is 0)
		if pkt < MSG_SYS_reserve01 {
			t.Errorf("System packet %s has ID %d, should be >= MSG_SYS_reserve01", pkt, pkt)
		}
	}
}

func TestMHFPacketIDRange(t *testing.T) {
	// MHF packets start at MSG_MHF_SAVEDATA (0x60)
	mhfPackets := []PacketID{
		MSG_MHF_SAVEDATA,
		MSG_MHF_LOADDATA,
		MSG_MHF_ENUMERATE_QUEST,
		MSG_MHF_ACQUIRE_TITLE,
		MSG_MHF_ACQUIRE_DIST_ITEM,
		MSG_MHF_ACQUIRE_MONTHLY_ITEM,
	}

	for _, pkt := range mhfPackets {
		// MHF packets should be >= MSG_MHF_SAVEDATA
		if pkt < MSG_MHF_SAVEDATA {
			t.Errorf("MHF packet %s has ID %d, should be >= MSG_MHF_SAVEDATA (%d)", pkt, pkt, MSG_MHF_SAVEDATA)
		}
	}
}

func TestStagePacketIDsSequential(t *testing.T) {
	// Stage-related packets should be sequential
	stagePackets := []PacketID{
		MSG_SYS_CREATE_STAGE,
		MSG_SYS_STAGE_DESTRUCT,
		MSG_SYS_ENTER_STAGE,
		MSG_SYS_BACK_STAGE,
		MSG_SYS_MOVE_STAGE,
		MSG_SYS_LEAVE_STAGE,
		MSG_SYS_LOCK_STAGE,
		MSG_SYS_UNLOCK_STAGE,
	}

	for i := 1; i < len(stagePackets); i++ {
		if stagePackets[i] != stagePackets[i-1]+1 {
			t.Errorf("Stage packets not sequential: %s (%d) should follow %s (%d)",
				stagePackets[i], stagePackets[i], stagePackets[i-1], stagePackets[i-1])
		}
	}
}

func TestPacketIDUniqueness(t *testing.T) {
	// Sample of important packet IDs should be unique
	packets := []PacketID{
		MSG_HEAD,
		MSG_SYS_END,
		MSG_SYS_NOP,
		MSG_SYS_ACK,
		MSG_SYS_LOGIN,
		MSG_SYS_LOGOUT,
		MSG_SYS_PING,
		MSG_SYS_TIME,
		MSG_SYS_CREATE_STAGE,
		MSG_SYS_ENTER_STAGE,
		MSG_MHF_SAVEDATA,
		MSG_MHF_LOADDATA,
	}

	seen := make(map[PacketID]bool)
	for _, pkt := range packets {
		if seen[pkt] {
			t.Errorf("Duplicate PacketID: %s (%d)", pkt, pkt)
		}
		seen[pkt] = true
	}
}

func TestAcquirePacketIDs(t *testing.T) {
	// Verify acquire-related packet IDs exist and are correct type
	acquirePackets := []PacketID{
		MSG_MHF_ACQUIRE_DIST_ITEM,
		MSG_MHF_ACQUIRE_TITLE,
		MSG_MHF_ACQUIRE_ITEM,
		MSG_MHF_ACQUIRE_MONTHLY_ITEM,
		MSG_MHF_ACQUIRE_CAFE_ITEM,
		MSG_MHF_ACQUIRE_GUILD_TRESURE,
	}

	for _, pkt := range acquirePackets {
		str := pkt.String()
		if str == "" {
			t.Errorf("PacketID %d should have a string representation", pkt)
		}
	}
}

func TestGuildPacketIDs(t *testing.T) {
	// Verify guild-related packet IDs
	guildPackets := []PacketID{
		MSG_MHF_CREATE_GUILD,
		MSG_MHF_OPERATE_GUILD,
		MSG_MHF_OPERATE_GUILD_MEMBER,
		MSG_MHF_INFO_GUILD,
		MSG_MHF_ENUMERATE_GUILD,
		MSG_MHF_UPDATE_GUILD,
	}

	for _, pkt := range guildPackets {
		// All guild packets should be MHF packets
		if pkt < MSG_MHF_SAVEDATA {
			t.Errorf("Guild packet %s should be an MHF packet (>= 0x60)", pkt)
		}
	}
}
