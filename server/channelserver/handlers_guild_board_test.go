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
	guildMock := &mockGuildRepo{
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
	guildMock := &mockGuildRepo{
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
	guildMock := &mockGuildRepo{}
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
	guildMock := &mockGuildRepo{
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
	guildMock := &mockGuildRepo{
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
	guildMock := &mockGuildRepo{
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
	guildMock := &mockGuildRepo{
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
	guildMock := &mockGuildRepo{
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

func TestHandleMsgMhfUpdateGuildMessageBoard_CreatePost(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 1,
		MessageOp: 0,
		PostType:  0,
		StampID:   1,
		Title:     "Test",
		Body:      "Hello",
	}
	handleMsgMhfUpdateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateGuildMessageBoard_CreatePostType1(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 1,
		MessageOp: 0,
		PostType:  1,
		Title:     "Notice",
		Body:      "Board",
	}
	handleMsgMhfUpdateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateGuildMessageBoard_DeletePost(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 1,
		MessageOp: 1,
		PostID:    42,
	}
	handleMsgMhfUpdateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateGuildMessageBoard_UpdatePost(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 1,
		MessageOp: 2,
		PostID:    1,
		Title:     "Updated",
		Body:      "New body",
	}
	handleMsgMhfUpdateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateGuildMessageBoard_UpdateStamp(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 1,
		MessageOp: 3,
		PostID:    1,
		StampID:   5,
	}
	handleMsgMhfUpdateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateGuildMessageBoard_LikePost(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 1,
		MessageOp: 4,
		PostID:    1,
		LikeState: true,
	}
	handleMsgMhfUpdateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateGuildMessageBoard_CheckNewPosts(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{
		AckHandle: 1,
		MessageOp: 5,
	}
	handleMsgMhfUpdateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateGuildMessageBoard_NoGuild(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{getErr: errNotFound}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{AckHandle: 1, MessageOp: 0}
	handleMsgMhfUpdateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfUpdateGuildMessageBoard_Applicant(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild, hasAppResult: true}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfUpdateGuildMessageBoard{AckHandle: 1, MessageOp: 0}
	handleMsgMhfUpdateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateGuildMessageBoard(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateGuildMessageBoard{AckHandle: 1, BoardType: 0}
	handleMsgMhfEnumerateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}

func TestHandleMsgMhfEnumerateGuildMessageBoard_Type1(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	srv.charRepo = newMockCharacterRepo()
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfEnumerateGuildMessageBoard{AckHandle: 1, BoardType: 1}
	handleMsgMhfEnumerateGuildMessageBoard(s, pkt)
	<-s.sendPackets
}
