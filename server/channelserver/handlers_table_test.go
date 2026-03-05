package channelserver

import (
	"testing"

	"erupe-ce/network"
)

func TestBuildHandlerTable_EntryCount(t *testing.T) {
	table := buildHandlerTable()
	// handlers_table.go has exactly 432 entries (one per packet ID).
	// This test catches accidental deletions or duplicates.
	const expectedCount = 432
	if len(table) != expectedCount {
		t.Errorf("handler table has %d entries, want %d", len(table), expectedCount)
	}
}

func TestBuildHandlerTable_CriticalOpcodes(t *testing.T) {
	table := buildHandlerTable()

	critical := []struct {
		name string
		id   network.PacketID
	}{
		{"MSG_SYS_LOGIN", network.MSG_SYS_LOGIN},
		{"MSG_SYS_LOGOUT", network.MSG_SYS_LOGOUT},
		{"MSG_SYS_PING", network.MSG_SYS_PING},
		{"MSG_SYS_ACK", network.MSG_SYS_ACK},
		{"MSG_SYS_CAST_BINARY", network.MSG_SYS_CAST_BINARY},
		{"MSG_SYS_ENTER_STAGE", network.MSG_SYS_ENTER_STAGE},
		{"MSG_SYS_LEAVE_STAGE", network.MSG_SYS_LEAVE_STAGE},
		{"MSG_MHF_SAVEDATA", network.MSG_MHF_SAVEDATA},
		{"MSG_MHF_LOADDATA", network.MSG_MHF_LOADDATA},
		{"MSG_MHF_ENUMERATE_QUEST", network.MSG_MHF_ENUMERATE_QUEST},
		{"MSG_MHF_CREATE_GUILD", network.MSG_MHF_CREATE_GUILD},
		{"MSG_MHF_INFO_GUILD", network.MSG_MHF_INFO_GUILD},
		{"MSG_MHF_GET_ACHIEVEMENT", network.MSG_MHF_GET_ACHIEVEMENT},
		{"MSG_MHF_PLAY_NORMAL_GACHA", network.MSG_MHF_PLAY_NORMAL_GACHA},
		{"MSG_MHF_SEND_MAIL", network.MSG_MHF_SEND_MAIL},
		{"MSG_MHF_SAVE_RENGOKU_DATA", network.MSG_MHF_SAVE_RENGOKU_DATA},
		{"MSG_MHF_LOAD_RENGOKU_DATA", network.MSG_MHF_LOAD_RENGOKU_DATA},
	}

	for _, tc := range critical {
		if _, ok := table[tc.id]; !ok {
			t.Errorf("critical opcode %s (0x%04X) is not mapped in handler table", tc.name, uint16(tc.id))
		}
	}
}

func TestBuildHandlerTable_NoNilHandlers(t *testing.T) {
	table := buildHandlerTable()
	for id, handler := range table {
		if handler == nil {
			t.Errorf("handler for opcode 0x%04X is nil", uint16(id))
		}
	}
}
