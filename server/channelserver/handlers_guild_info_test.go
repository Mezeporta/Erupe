package channelserver

import (
	"testing"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/network/mhfpacket"
)

// guildInfoServer creates a mock server with ClanMemberLimits set,
// which handleMsgMhfInfoGuild requires.
func guildInfoServer() *Server {
	s := createMockServer()
	s.erupeConfig.GameplayOptions.ClanMemberLimits = [][]uint8{{0, 30}}
	return s
}

// --- handleMsgMhfInfoGuild tests ---

func TestInfoGuild_ByGuildID(t *testing.T) {
	server := guildInfoServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 1, IsLeader: true},
	}
	joined := time.Now()
	guildMock.membership.JoinedAt = &joined
	guildMock.guild = &Guild{
		ID:          10,
		Name:        "Test",
		Comment:     "Hello",
		MemberCount: 5,
		CreatedAt:   time.Now(),
		RoomExpiry:  time.Now().Add(time.Hour),
	}
	guildMock.guild.LeaderCharID = 1
	guildMock.guild.LeaderName = "Leader"
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoGuild{AckHandle: 100, GuildID: 10}

	handleMsgMhfInfoGuild(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 20 {
			t.Errorf("Response too short: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}

	if session.prevGuildID != 10 {
		t.Errorf("prevGuildID = %d, want 10", session.prevGuildID)
	}
}

func TestInfoGuild_ByCharID(t *testing.T) {
	server := guildInfoServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 5},
	}
	guildMock.guild = &Guild{
		ID:         10,
		Name:       "MyGuild",
		CreatedAt:  time.Now(),
		RoomExpiry: time.Now(),
	}
	guildMock.guild.LeaderCharID = 99
	guildMock.guild.LeaderName = "Boss"
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	// GuildID=0 means look up by charID
	pkt := &mhfpacket.MsgMhfInfoGuild{AckHandle: 100, GuildID: 0}

	handleMsgMhfInfoGuild(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 20 {
			t.Errorf("Response too short: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestInfoGuild_NotFound(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.getErr = errNotFound
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoGuild{AckHandle: 100, GuildID: 999}

	handleMsgMhfInfoGuild(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestInfoGuild_MembershipError(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		getMemberErr: errNotFound,
	}
	guildMock.guild = &Guild{
		ID:         10,
		Name:       "Test",
		CreatedAt:  time.Now(),
		RoomExpiry: time.Now(),
	}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoGuild{AckHandle: 100, GuildID: 10}

	handleMsgMhfInfoGuild(session, pkt)

	// Should return early with count=0 response
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestInfoGuild_WithAlliance(t *testing.T) {
	server := guildInfoServer()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 1, IsLeader: true},
		alliance: &GuildAlliance{
			ID:            5,
			Name:          "TestAlliance",
			CreatedAt:     time.Now(),
			TotalMembers:  15,
			ParentGuildID: 10,
			ParentGuild:   Guild{Name: "Test", MemberCount: 5},
		},
	}
	guildMock.guild = &Guild{
		ID:         10,
		Name:       "Test",
		CreatedAt:  time.Now(),
		RoomExpiry: time.Now(),
		AllianceID: 5,
	}
	guildMock.guild.LeaderCharID = 1
	guildMock.guild.LeaderName = "Leader"
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfInfoGuild{AckHandle: 100, GuildID: 10}

	handleMsgMhfInfoGuild(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 50 {
			t.Errorf("Alliance response too short: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfEnumerateGuild tests ---

func TestEnumerateGuild_ByName(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.guild = nil
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	// Simulate search term in Data2
	data2 := byteframe.NewByteFrame()
	data2.WriteBytes([]byte("Test\x00"))
	_, _ = data2.Seek(0, 0)

	pkt := &mhfpacket.MsgMhfEnumerateGuild{
		AckHandle: 100,
		Type:      mhfpacket.ENUMERATE_GUILD_TYPE_GUILD_NAME,
		Data2:     data2,
	}

	handleMsgMhfEnumerateGuild(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateGuild_NoResults(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.getErr = errNotFound
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	data2 := byteframe.NewByteFrame()
	data2.WriteBytes([]byte("NonExistent\x00"))
	_, _ = data2.Seek(0, 0)

	pkt := &mhfpacket.MsgMhfEnumerateGuild{
		AckHandle: 100,
		Type:      mhfpacket.ENUMERATE_GUILD_TYPE_GUILD_NAME,
		Data2:     data2,
	}

	handleMsgMhfEnumerateGuild(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}
