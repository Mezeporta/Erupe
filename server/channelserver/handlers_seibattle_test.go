package channelserver

import (
	"encoding/binary"
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetSeibattle_AllTypes(t *testing.T) {
	tests := []struct {
		name    string
		pktType uint8
	}{
		{"Timetable", 1},
		{"KeyScore", 3},
		{"Career", 4},
		{"Opponent", 5},
		{"ConventionResult", 6},
		{"CharScore", 7},
		{"CurResult", 8},
		{"UnknownType", 99},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := createMockServer()
			server.erupeConfig.EarthID = 1
			session := createMockSession(1, server)

			pkt := &mhfpacket.MsgMhfGetSeibattle{
				AckHandle: 100,
				Type:      tt.pktType,
			}
			handleMsgMhfGetSeibattle(session, pkt)

			select {
			case p := <-session.sendPackets:
				_, errCode, ackData := parseAckBufData(t, p.data)
				if errCode != 0 {
					t.Errorf("ErrorCode = %d, want 0", errCode)
				}
				// Earth header: EarthID(4) + 0(4) + 0(4) + count(4) = 16 bytes minimum
				if len(ackData) < 16 {
					t.Errorf("AckData too short: %d bytes", len(ackData))
				}
				earthID := binary.BigEndian.Uint32(ackData[:4])
				if earthID != 1 {
					t.Errorf("EarthID = %d, want 1", earthID)
				}
			default:
				t.Fatal("No response queued")
			}
		})
	}
}

func TestHandleMsgMhfGetSeibattle_TimetableEntryCount(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.EarthID = 1
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetSeibattle{
		AckHandle: 100,
		Type:      1, // Timetable
	}
	handleMsgMhfGetSeibattle(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		count := binary.BigEndian.Uint32(ackData[12:16])
		if count != 3 {
			t.Errorf("timetable count = %d, want 3", count)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfGetBreakSeibatuLevelReward_DataSize(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetBreakSeibatuLevelReward{AckHandle: 100}
	handleMsgMhfGetBreakSeibatuLevelReward(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		// 4 × int32 = 16 bytes
		if len(ackData) != 16 {
			t.Errorf("AckData len = %d, want 16", len(ackData))
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfGetWeeklySeibatuRankingReward_EarthFormat(t *testing.T) {
	server := createMockServer()
	server.erupeConfig.EarthID = 42
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetWeeklySeibatuRankingReward{AckHandle: 100}
	handleMsgMhfGetWeeklySeibatuRankingReward(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		earthID := binary.BigEndian.Uint32(ackData[:4])
		if earthID != 42 {
			t.Errorf("EarthID = %d, want 42", earthID)
		}
		count := binary.BigEndian.Uint32(ackData[12:16])
		if count != 1 {
			t.Errorf("reward count = %d, want 1", count)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfGetFixedSeibatuRankingTable_DataSize(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetFixedSeibatuRankingTable{AckHandle: 100}
	handleMsgMhfGetFixedSeibatuRankingTable(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		// 4 + 4 + 32 = 40 bytes
		if len(ackData) != 40 {
			t.Errorf("AckData len = %d, want 40", len(ackData))
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfReadBeatLevel_VerifyIDEcho(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevel{
		AckHandle:    100,
		ValidIDCount: 2,
		IDs:          [16]uint32{0x74, 0x6B},
	}
	handleMsgMhfReadBeatLevel(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		// 2 entries × (4+4+4+4) = 32 bytes
		if len(ackData) != 32 {
			t.Errorf("AckData len = %d, want 32", len(ackData))
		}
		firstID := binary.BigEndian.Uint32(ackData[:4])
		if firstID != 0x74 {
			t.Errorf("first ID = 0x%x, want 0x74", firstID)
		}
		secondID := binary.BigEndian.Uint32(ackData[16:20])
		if secondID != 0x6B {
			t.Errorf("second ID = 0x%x, want 0x6B", secondID)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfReadBeatLevelAllRanking_DataSize(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevelAllRanking{AckHandle: 100}
	handleMsgMhfReadBeatLevelAllRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		// 4+4+4 + 100*(4+4+32) = 4012 bytes
		expectedLen := 12 + 100*40
		if len(ackData) != expectedLen {
			t.Errorf("AckData len = %d, want %d", len(ackData), expectedLen)
		}
	default:
		t.Fatal("No response queued")
	}
}

func TestHandleMsgMhfReadBeatLevelMyRanking_EmptyResponse(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadBeatLevelMyRanking{AckHandle: 100}
	handleMsgMhfReadBeatLevelMyRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		if len(ackData) != 0 {
			t.Errorf("AckData len = %d, want 0", len(ackData))
		}
	default:
		t.Fatal("No response queued")
	}
}

// Tests consolidated from handlers_coverage3_test.go

func TestNonTrivialHandlers_NoDB_Seibattle(t *testing.T) {
	server := createMockServer()

	t.Run("handleMsgMhfGetSeibattle", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetSeibattle(session, &mhfpacket.MsgMhfGetSeibattle{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfUpdateBeatLevel", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfUpdateBeatLevel(session, &mhfpacket.MsgMhfUpdateBeatLevel{AckHandle: 1})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})

	t.Run("handleMsgMhfReadBeatLevel", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfReadBeatLevel(session, &mhfpacket.MsgMhfReadBeatLevel{
			AckHandle:    1,
			ValidIDCount: 2,
			IDs:          [16]uint32{100, 200},
		})
		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Error("response should have data")
			}
		default:
			t.Error("no response queued")
		}
	})
}

func TestHandleMsgMhfReadLastWeekBeatRanking_DataSize(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfReadLastWeekBeatRanking{AckHandle: 100}
	handleMsgMhfReadLastWeekBeatRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		_, _, ackData := parseAckBufData(t, p.data)
		if len(ackData) != 16 {
			t.Errorf("AckData len = %d, want 16", len(ackData))
		}
	default:
		t.Fatal("No response queued")
	}
}
