package channelserver

import (
	"testing"
	"time"

	"erupe-ce/network/mhfpacket"
)

// --- handleMsgMhfUpdateGuildMessageBoard tests ---

func TestUpdateGuildMessageBoard_CreatePost(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 100,
		MessageOp: 0, // Create
		PostType:  0,
		StampID:   5,
		Title:     "Test Title",
		Body:      "Test Body",
	}

	handleMsgMhfUpdateGuildMessageBoard(session, pkt)

	if guildMock.createdPost == nil {
		t.Fatal("CreatePost should be called")
	}
	if guildMock.createdPost[0].(uint32) != 10 {
		t.Errorf("CreatePost guildID = %d, want 10", guildMock.createdPost[0])
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestUpdateGuildMessageBoard_DeletePost(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	guildMock := &mockGuildRepoOps{
		membership: &GuildMember{GuildID: 10, CharID: 1, OrderIndex: 1},
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 100,
		MessageOp: 1, // Delete
		PostID:    42,
	}

	handleMsgMhfUpdateGuildMessageBoard(session, pkt)

	if guildMock.deletedPostID != 42 {
		t.Errorf("DeletePost postID = %d, want 42", guildMock.deletedPostID)
	}
}

func TestUpdateGuildMessageBoard_NoGuild(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	guildMock := &mockGuildRepoOps{}
	guildMock.getErr = errNotFound
	server.guildRepo = guildMock
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 100,
		MessageOp: 0,
	}

	handleMsgMhfUpdateGuildMessageBoard(session, pkt)

	// Returns early with empty success
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestUpdateGuildMessageBoard_Applicant(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	guildMock := &mockGuildRepoOps{
		hasAppResult: true, // is an applicant
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 999
	server.guildRepo = guildMock
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 100,
		MessageOp: 0,
	}

	handleMsgMhfUpdateGuildMessageBoard(session, pkt)

	if guildMock.createdPost != nil {
		t.Error("Applicant should not be able to create posts")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestUpdateGuildMessageBoard_HasAppError(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	guildMock := &mockGuildRepoOps{
		hasAppErr: errNotFound, // error checking app status
	}
	guildMock.guild = &Guild{ID: 10}
	guildMock.guild.LeaderCharID = 1
	server.guildRepo = guildMock
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 100,
		MessageOp: 0,
		Title:     "Test",
		Body:      "Body",
	}

	// Should log warning and treat as non-applicant (applicant=false on error)
	handleMsgMhfUpdateGuildMessageBoard(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfEnumerateGuildMessageBoard tests ---

func TestEnumerateGuildMessageBoard_NoPosts(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	guildMock := &mockGuildRepoOps{
		posts: []*MessageBoardPost{},
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildMessageBoard{
		AckHandle: 100,
		BoardType: 0,
		MaxPosts:  100,
	}

	handleMsgMhfEnumerateGuildMessageBoard(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateGuildMessageBoard_WithPosts(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	guildMock := &mockGuildRepoOps{
		posts: []*MessageBoardPost{
			{ID: 1, AuthorID: 100, StampID: 5, Title: "Hello", Body: "World", Timestamp: time.Now()},
			{ID: 2, AuthorID: 200, StampID: 0, Title: "Test", Body: "Post", Timestamp: time.Now()},
		},
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildMessageBoard{
		AckHandle: 100,
		BoardType: 0,
		MaxPosts:  100,
	}

	handleMsgMhfEnumerateGuildMessageBoard(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 8 {
			t.Errorf("Response too short for 2 posts: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestEnumerateGuildMessageBoard_DBError(t *testing.T) {
	server := createMockServer()
	charMock := newMockCharacterRepo()
	guildMock := &mockGuildRepoOps{
		listPostsErr: errNotFound,
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	server.charRepo = charMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfEnumerateGuildMessageBoard{
		AckHandle: 100,
		BoardType: 0,
		MaxPosts:  100,
	}

	handleMsgMhfEnumerateGuildMessageBoard(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}
