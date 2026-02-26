package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"testing"
)

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
