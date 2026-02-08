package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestReserveHandlersWithAck(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	// Test handleMsgSysReserve188
	handleMsgSysReserve188(session, &mhfpacket.MsgSysReserve188{AckHandle: 12345})
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Reserve188: response should have data")
		}
	default:
		t.Error("Reserve188: no response queued")
	}

	// Test handleMsgSysReserve18B
	handleMsgSysReserve18B(session, &mhfpacket.MsgSysReserve18B{AckHandle: 12345})
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Reserve18B: response should have data")
		}
	default:
		t.Error("Reserve18B: no response queued")
	}
}

func TestReserveEmptyHandlers(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	tests := []struct {
		name    string
		handler func(s *Session, p mhfpacket.MHFPacket)
		pkt     mhfpacket.MHFPacket
	}{
		{"Reserve55", handleMsgSysReserve55, &mhfpacket.MsgSysReserve55{}},
		{"Reserve56", handleMsgSysReserve56, &mhfpacket.MsgSysReserve56{}},
		{"Reserve57", handleMsgSysReserve57, &mhfpacket.MsgSysReserve57{}},
		{"Reserve01", handleMsgSysReserve01, &mhfpacket.MsgSysReserve01{}},
		{"Reserve02", handleMsgSysReserve02, &mhfpacket.MsgSysReserve02{}},
		{"Reserve03", handleMsgSysReserve03, &mhfpacket.MsgSysReserve03{}},
		{"Reserve04", handleMsgSysReserve04, &mhfpacket.MsgSysReserve04{}},
		{"Reserve05", handleMsgSysReserve05, &mhfpacket.MsgSysReserve05{}},
		{"Reserve06", handleMsgSysReserve06, &mhfpacket.MsgSysReserve06{}},
		{"Reserve07", handleMsgSysReserve07, &mhfpacket.MsgSysReserve07{}},
		{"Reserve0C", handleMsgSysReserve0C, &mhfpacket.MsgSysReserve0C{}},
		{"Reserve0D", handleMsgSysReserve0D, &mhfpacket.MsgSysReserve0D{}},
		{"Reserve0E", handleMsgSysReserve0E, &mhfpacket.MsgSysReserve0E{}},
		{"Reserve4A", handleMsgSysReserve4A, &mhfpacket.MsgSysReserve4A{}},
		{"Reserve4B", handleMsgSysReserve4B, &mhfpacket.MsgSysReserve4B{}},
		{"Reserve4C", handleMsgSysReserve4C, &mhfpacket.MsgSysReserve4C{}},
		{"Reserve4D", handleMsgSysReserve4D, &mhfpacket.MsgSysReserve4D{}},
		{"Reserve4E", handleMsgSysReserve4E, &mhfpacket.MsgSysReserve4E{}},
		{"Reserve4F", handleMsgSysReserve4F, &mhfpacket.MsgSysReserve4F{}},
		{"Reserve5C", handleMsgSysReserve5C, &mhfpacket.MsgSysReserve5C{}},
		{"Reserve5E", handleMsgSysReserve5E, &mhfpacket.MsgSysReserve5E{}},
		{"Reserve5F", handleMsgSysReserve5F, &mhfpacket.MsgSysReserve5F{}},
		{"Reserve71", handleMsgSysReserve71, &mhfpacket.MsgSysReserve71{}},
		{"Reserve72", handleMsgSysReserve72, &mhfpacket.MsgSysReserve72{}},
		{"Reserve73", handleMsgSysReserve73, &mhfpacket.MsgSysReserve73{}},
		{"Reserve74", handleMsgSysReserve74, &mhfpacket.MsgSysReserve74{}},
		{"Reserve75", handleMsgSysReserve75, &mhfpacket.MsgSysReserve75{}},
		{"Reserve76", handleMsgSysReserve76, &mhfpacket.MsgSysReserve76{}},
		{"Reserve77", handleMsgSysReserve77, &mhfpacket.MsgSysReserve77{}},
		{"Reserve78", handleMsgSysReserve78, &mhfpacket.MsgSysReserve78{}},
		{"Reserve79", handleMsgSysReserve79, &mhfpacket.MsgSysReserve79{}},
		{"Reserve7A", handleMsgSysReserve7A, &mhfpacket.MsgSysReserve7A{}},
		{"Reserve7B", handleMsgSysReserve7B, &mhfpacket.MsgSysReserve7B{}},
		{"Reserve7C", handleMsgSysReserve7C, &mhfpacket.MsgSysReserve7C{}},
		{"Reserve7E", handleMsgSysReserve7E, &mhfpacket.MsgSysReserve7E{}},
		{"Reserve10F", handleMsgMhfReserve10F, &mhfpacket.MsgMhfReserve10F{}},
		{"Reserve180", handleMsgSysReserve180, &mhfpacket.MsgSysReserve180{}},
		{"Reserve18E", handleMsgSysReserve18E, &mhfpacket.MsgSysReserve18E{}},
		{"Reserve18F", handleMsgSysReserve18F, &mhfpacket.MsgSysReserve18F{}},
		{"Reserve19E", handleMsgSysReserve19E, &mhfpacket.MsgSysReserve19E{}},
		{"Reserve19F", handleMsgSysReserve19F, &mhfpacket.MsgSysReserve19F{}},
		{"Reserve1A4", handleMsgSysReserve1A4, &mhfpacket.MsgSysReserve1A4{}},
		{"Reserve1A6", handleMsgSysReserve1A6, &mhfpacket.MsgSysReserve1A6{}},
		{"Reserve1A7", handleMsgSysReserve1A7, &mhfpacket.MsgSysReserve1A7{}},
		{"Reserve1A8", handleMsgSysReserve1A8, &mhfpacket.MsgSysReserve1A8{}},
		{"Reserve1A9", handleMsgSysReserve1A9, &mhfpacket.MsgSysReserve1A9{}},
		{"Reserve1AA", handleMsgSysReserve1AA, &mhfpacket.MsgSysReserve1AA{}},
		{"Reserve1AB", handleMsgSysReserve1AB, &mhfpacket.MsgSysReserve1AB{}},
		{"Reserve1AC", handleMsgSysReserve1AC, &mhfpacket.MsgSysReserve1AC{}},
		{"Reserve1AD", handleMsgSysReserve1AD, &mhfpacket.MsgSysReserve1AD{}},
		{"Reserve1AE", handleMsgSysReserve1AE, &mhfpacket.MsgSysReserve1AE{}},
		{"Reserve1AF", handleMsgSysReserve1AF, &mhfpacket.MsgSysReserve1AF{}},
		{"Reserve19B", handleMsgSysReserve19B, &mhfpacket.MsgSysReserve19B{}},
		{"Reserve192", handleMsgSysReserve192, &mhfpacket.MsgSysReserve192{}},
		{"Reserve193", handleMsgSysReserve193, &mhfpacket.MsgSysReserve193{}},
		{"Reserve194", handleMsgSysReserve194, &mhfpacket.MsgSysReserve194{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("%s panicked: %v", tt.name, r)
				}
			}()
			tt.handler(session, tt.pkt)
		})
	}
}
