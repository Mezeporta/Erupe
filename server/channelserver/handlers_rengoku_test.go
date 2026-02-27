package channelserver

import (
	"testing"

	"erupe-ce/network/mhfpacket"
)

func TestHandleMsgMhfGetRengokuRankingRank(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRengokuRankingRank{
		AckHandle: 12345,
	}

	handleMsgMhfGetRengokuRankingRank(session, pkt)

	// Verify response packet was queued
	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestRengokuScoreStruct(t *testing.T) {
	score := RengokuScore{
		Name:  "TestPlayer",
		Score: 12345,
	}

	if score.Name != "TestPlayer" {
		t.Errorf("Name = %s, want TestPlayer", score.Name)
	}
	if score.Score != 12345 {
		t.Errorf("Score = %d, want 12345", score.Score)
	}
}

func TestRengokuScoreStruct_DefaultValues(t *testing.T) {
	score := RengokuScore{}

	if score.Name != "" {
		t.Errorf("Default Name should be empty, got %s", score.Name)
	}
	if score.Score != 0 {
		t.Errorf("Default Score should be 0, got %d", score.Score)
	}
}

func TestHandleMsgMhfGetRengokuRankingRank_ResponseData(t *testing.T) {
	server := createMockServer()
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGetRengokuRankingRank{
		AckHandle: 55555,
	}

	handleMsgMhfGetRengokuRankingRank(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response packet should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestRengokuScoreStruct_Fields(t *testing.T) {
	score := RengokuScore{
		Name:  "Hunter",
		Score: 99999,
	}

	if score.Name != "Hunter" {
		t.Errorf("Name = %s, want Hunter", score.Name)
	}
	if score.Score != 99999 {
		t.Errorf("Score = %d, want 99999", score.Score)
	}
}

// TestHandleMsgMhfGetRengokuRankingRank_DifferentAck verifies rengoku ranking
// works with different ack handles.
func TestHandleMsgMhfGetRengokuRankingRank_DifferentAck(t *testing.T) {
	server := createMockServer()

	ackHandles := []uint32{0, 1, 54321, 0xDEADBEEF}
	for _, ack := range ackHandles {
		session := createMockSession(1, server)
		pkt := &mhfpacket.MsgMhfGetRengokuRankingRank{AckHandle: ack}

		handleMsgMhfGetRengokuRankingRank(session, pkt)

		select {
		case p := <-session.sendPackets:
			if len(p.data) == 0 {
				t.Errorf("AckHandle=%d: Response packet should have data", ack)
			}
		default:
			t.Errorf("AckHandle=%d: No response packet queued", ack)
		}
	}
}

// --- handleMsgMhfSaveRengokuData tests ---

func TestSaveRengokuData_TooSmall(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	server.charRepo = charMock
	server.rengokuRepo = &mockRengokuRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      100,
		RawDataPayload: make([]byte, 10), // too small
	}

	handleMsgMhfSaveRengokuData(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestSaveRengokuData_TooLarge(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	server.charRepo = charMock
	server.rengokuRepo = &mockRengokuRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      100,
		RawDataPayload: make([]byte, 5000), // too large
	}

	handleMsgMhfSaveRengokuData(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestSaveRengokuData_NormalSave(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	server.charRepo = charMock
	server.rengokuRepo = &mockRengokuRepo{}
	session := createMockSession(1, server)

	// Build valid payload (>= rengokuMinPayloadSize=91 bytes)
	payload := make([]byte, 100)
	// Set sentinel to non-zero so it's not rejected
	payload[0] = 0x00
	payload[1] = 0x00
	payload[2] = 0x00
	payload[3] = 0x01
	// Set some skill data so it's not zeroed
	payload[rengokuSkillSlotsStart] = 1

	pkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      100,
		RawDataPayload: payload,
	}

	handleMsgMhfSaveRengokuData(session, pkt)

	if charMock.columns["rengokudata"] == nil {
		t.Error("rengokudata should be saved")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestSaveRengokuData_SkillMerge(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()

	// Set up existing data with skills
	existing := make([]byte, 100)
	existing[0] = 0x00
	existing[1] = 0x00
	existing[2] = 0x00
	existing[3] = 0x01
	existing[rengokuSkillSlotsStart] = 5 // has skill
	existing[rengokuSkillValuesStart] = 3
	charMock.columns["rengokudata"] = existing

	server.charRepo = charMock
	server.rengokuRepo = &mockRengokuRepo{}
	session := createMockSession(1, server)

	// Build payload with zeroed skills but has points (triggers merge)
	payload := make([]byte, 100)
	payload[0] = 0x00
	payload[1] = 0x00
	payload[2] = 0x00
	payload[3] = 0x01
	// Skills are zeroed (default)
	// But points are set
	payload[rengokuPointsStart] = 10

	pkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      100,
		RawDataPayload: payload,
	}

	handleMsgMhfSaveRengokuData(session, pkt)

	saved := charMock.columns["rengokudata"]
	if saved == nil {
		t.Fatal("rengokudata should be saved")
	}
	// Skills should be merged from existing
	if saved[rengokuSkillSlotsStart] != 5 {
		t.Errorf("Skill slot should be merged from existing, got %d", saved[rengokuSkillSlotsStart])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestSaveRengokuData_SentinelRejection(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()

	// Set up existing data with non-zero sentinel
	existing := make([]byte, 100)
	existing[0] = 0x00
	existing[1] = 0x00
	existing[2] = 0x00
	existing[3] = 0x01
	existing[rengokuSkillSlotsStart] = 1
	charMock.columns["rengokudata"] = existing

	server.charRepo = charMock
	server.rengokuRepo = &mockRengokuRepo{}
	session := createMockSession(1, server)

	// Build payload with zero sentinel (should be rejected)
	payload := make([]byte, 100)
	// sentinel is 0 (all zeros)
	payload[rengokuSkillSlotsStart] = 1 // non-zeroed skills to skip merge path

	pkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      100,
		RawDataPayload: payload,
	}

	handleMsgMhfSaveRengokuData(session, pkt)

	// Existing data should be preserved (not overwritten)
	if charMock.columns["rengokudata"][3] != 0x01 {
		t.Error("Existing rengoku data should not be overwritten")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestSaveRengokuData_SaveError(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.saveErr = errNotFound
	server.charRepo = charMock
	server.rengokuRepo = &mockRengokuRepo{}
	session := createMockSession(1, server)

	payload := make([]byte, 100)
	payload[3] = 0x01
	payload[rengokuSkillSlotsStart] = 1

	pkt := &mhfpacket.MsgMhfSaveRengokuData{
		AckHandle:      100,
		RawDataPayload: payload,
	}

	handleMsgMhfSaveRengokuData(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfLoadRengokuData tests ---

func TestLoadRengokuData_WithData(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.columns["rengokudata"] = []byte{0x01, 0x02, 0x03, 0x04}
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadRengokuData{AckHandle: 100}

	handleMsgMhfLoadRengokuData(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestLoadRengokuData_EmptyData(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	// No rengokudata column set - returns nil/empty
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadRengokuData{AckHandle: 100}

	handleMsgMhfLoadRengokuData(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Default response should have data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestLoadRengokuData_DBError(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	charMock.loadColumnErr = errNotFound
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadRengokuData{AckHandle: 100}

	handleMsgMhfLoadRengokuData(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Default response should have data on error")
		}
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfEnumerateRengokuRanking tests ---

func TestEnumerateRengokuRanking_NoGuild_Leaderboard2(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{} // GetByCharID returns nil
	server.guildRepo = guildMock
	server.rengokuRepo = &mockRengokuRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRengokuRanking{
		AckHandle:   100,
		Leaderboard: 2, // guild leaderboard, requires guild
	}

	handleMsgMhfEnumerateRengokuRanking(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateRengokuRanking_NoGuild_Leaderboard3(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{}
	server.guildRepo = guildMock
	server.rengokuRepo = &mockRengokuRepo{}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRengokuRanking{
		AckHandle:   100,
		Leaderboard: 3,
	}

	handleMsgMhfEnumerateRengokuRanking(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateRengokuRanking_WithGuild(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 10, Name: "TestGuild"}
	guildMock := &mockGuildRepo{}
	guildMock.guild = guild
	server.guildRepo = guildMock
	server.rengokuRepo = &mockRengokuRepo{
		ranking: []RengokuScore{
			{Name: "Player1", Score: 100},
			{Name: "Player2", Score: 50},
		},
	}
	session := createMockSession(1, server)
	session.Name = "Player1"

	pkt := &mhfpacket.MsgMhfEnumerateRengokuRanking{
		AckHandle:   100,
		Leaderboard: 2,
	}

	handleMsgMhfEnumerateRengokuRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response should have ranking data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateRengokuRanking_SoloLeaderboard(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{}
	server.guildRepo = guildMock
	server.rengokuRepo = &mockRengokuRepo{
		ranking: []RengokuScore{
			{Name: "Player1", Score: 200},
		},
	}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRengokuRanking{
		AckHandle:   100,
		Leaderboard: 0, // solo, no guild required
	}

	handleMsgMhfEnumerateRengokuRanking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response should have ranking data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateRengokuRanking_QueryError(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepo{}
	server.guildRepo = guildMock
	server.rengokuRepo = &mockRengokuRepo{rankingErr: errNotFound}
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateRengokuRanking{
		AckHandle:   100,
		Leaderboard: 0,
	}

	handleMsgMhfEnumerateRengokuRanking(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateRengokuRanking_Applicant(t *testing.T) {
	server := createMockServer()
	guild := &Guild{ID: 10, Name: "TestGuild"}
	guildMock := &mockGuildRepo{hasAppResult: true}
	guildMock.guild = guild
	server.guildRepo = guildMock
	server.rengokuRepo = &mockRengokuRepo{}
	session := createMockSession(1, server)

	// Leaderboard 6 requires guild, but applicant should be treated as no guild
	pkt := &mhfpacket.MsgMhfEnumerateRengokuRanking{
		AckHandle:   100,
		Leaderboard: 6,
	}

	handleMsgMhfEnumerateRengokuRanking(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// Tests consolidated from handlers_coverage3_test.go

func TestNonTrivialHandlers_RengokuGo(t *testing.T) {
	server := createMockServer()

	t.Run("handleMsgMhfGetRengokuRankingRank", func(t *testing.T) {
		session := createMockSession(1, server)
		handleMsgMhfGetRengokuRankingRank(session, &mhfpacket.MsgMhfGetRengokuRankingRank{AckHandle: 1})
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
