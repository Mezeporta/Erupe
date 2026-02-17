package network

import (
	"strings"
	"testing"
)

func TestPacketIDString_KnownIDs(t *testing.T) {
	tests := []struct {
		id   PacketID
		want string
	}{
		{MSG_HEAD, "MSG_HEAD"},
		{MSG_SYS_ACK, "MSG_SYS_ACK"},
		{MSG_SYS_PING, "MSG_SYS_PING"},
		{MSG_SYS_LOGIN, "MSG_SYS_LOGIN"},
		{MSG_MHF_SAVEDATA, "MSG_MHF_SAVEDATA"},
		{MSG_MHF_CREATE_GUILD, "MSG_MHF_CREATE_GUILD"},
		{MSG_SYS_reserve1AF, "MSG_SYS_reserve1AF"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := tt.id.String()
			if got != tt.want {
				t.Errorf("PacketID(%d).String() = %q, want %q", tt.id, got, tt.want)
			}
		})
	}
}

func TestPacketIDString_OutOfRange(t *testing.T) {
	// An ID beyond the known range should return "PacketID(N)"
	id := PacketID(9999)
	got := id.String()
	if !strings.HasPrefix(got, "PacketID(") {
		t.Errorf("out-of-range PacketID String() = %q, want prefix 'PacketID('", got)
	}
}

func TestPacketIDString_AllValid(t *testing.T) {
	// Verify all valid PacketIDs produce non-empty strings
	for i := PacketID(0); i <= MSG_SYS_reserve1AF; i++ {
		got := i.String()
		if got == "" {
			t.Errorf("PacketID(%d).String() returned empty string", i)
		}
		if strings.HasPrefix(got, "PacketID(") {
			t.Errorf("PacketID(%d).String() = %q, expected named constant", i, got)
		}
	}
}
