package channelserver

import (
	"testing"
	"time"

	"erupe-ce/network/mhfpacket"
)

func TestMailStruct(t *testing.T) {
	mail := Mail{
		ID:                   123,
		SenderID:             1000,
		RecipientID:          2000,
		Subject:              "Test Subject",
		Body:                 "Test Body Content",
		Read:                 false,
		Deleted:              false,
		Locked:               true,
		AttachedItemReceived: false,
		AttachedItemID:       500,
		AttachedItemAmount:   10,
		CreatedAt:            time.Now(),
		IsGuildInvite:        false,
		IsSystemMessage:      true,
		SenderName:           "TestSender",
	}

	if mail.ID != 123 {
		t.Errorf("ID = %d, want 123", mail.ID)
	}
	if mail.SenderID != 1000 {
		t.Errorf("SenderID = %d, want 1000", mail.SenderID)
	}
	if mail.RecipientID != 2000 {
		t.Errorf("RecipientID = %d, want 2000", mail.RecipientID)
	}
	if mail.Subject != "Test Subject" {
		t.Errorf("Subject = %s, want 'Test Subject'", mail.Subject)
	}
	if mail.Body != "Test Body Content" {
		t.Errorf("Body = %s, want 'Test Body Content'", mail.Body)
	}
	if mail.Read {
		t.Error("Read should be false")
	}
	if mail.Deleted {
		t.Error("Deleted should be false")
	}
	if !mail.Locked {
		t.Error("Locked should be true")
	}
	if mail.AttachedItemReceived {
		t.Error("AttachedItemReceived should be false")
	}
	if mail.AttachedItemID != 500 {
		t.Errorf("AttachedItemID = %d, want 500", mail.AttachedItemID)
	}
	if mail.AttachedItemAmount != 10 {
		t.Errorf("AttachedItemAmount = %d, want 10", mail.AttachedItemAmount)
	}
	if mail.IsGuildInvite {
		t.Error("IsGuildInvite should be false")
	}
	if !mail.IsSystemMessage {
		t.Error("IsSystemMessage should be true")
	}
	if mail.SenderName != "TestSender" {
		t.Errorf("SenderName = %s, want 'TestSender'", mail.SenderName)
	}
}

func TestMailStruct_DefaultValues(t *testing.T) {
	mail := Mail{}

	if mail.ID != 0 {
		t.Errorf("Default ID should be 0, got %d", mail.ID)
	}
	if mail.Subject != "" {
		t.Errorf("Default Subject should be empty, got %s", mail.Subject)
	}
	if mail.Read {
		t.Error("Default Read should be false")
	}
}

// --- Mock-based handler tests ---

func TestHandleMsgMhfListMail_Empty(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{mails: []Mail{}}
	server.mailRepo = mock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfListMail{AckHandle: 100}

	handleMsgMhfListMail(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfListMail_WithMails(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{
		mails: []Mail{
			{ID: 10, SenderID: 100, Subject: "Hello", SenderName: "Sender1", CreatedAt: time.Now()},
			{ID: 20, SenderID: 200, Subject: "World", SenderName: "Sender2", CreatedAt: time.Now(), Locked: true},
		},
	}
	server.mailRepo = mock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfListMail{AckHandle: 100}

	handleMsgMhfListMail(session, pkt)

	// Verify mailList was populated
	if session.mailList == nil {
		t.Fatal("mailList should be initialized")
	}
	if session.mailList[0] != 10 {
		t.Errorf("mailList[0] = %d, want 10", session.mailList[0])
	}
	if session.mailList[1] != 20 {
		t.Errorf("mailList[1] = %d, want 20", session.mailList[1])
	}
	if session.mailAccIndex != 2 {
		t.Errorf("mailAccIndex = %d, want 2", session.mailAccIndex)
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 10 {
			t.Errorf("Response too short: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfListMail_DBError(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{listErr: errNotFound}
	server.mailRepo = mock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfListMail{AckHandle: 100}

	handleMsgMhfListMail(session, pkt)

	select {
	case p := <-session.sendPackets:
		// Should return a fallback response with single zero byte
		if len(p.data) == 0 {
			t.Error("Should have fallback response data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadMail_Success(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{
		mailByID: map[int]*Mail{
			42: {ID: 42, Body: "Test body content"},
		},
	}
	server.mailRepo = mock
	session := createMockSession(1, server)
	session.mailList = make([]int, 256)
	session.mailList[0] = 42

	pkt := &mhfpacket.MsgMhfReadMail{
		AckHandle: 100,
		AccIndex:  0,
	}

	handleMsgMhfReadMail(session, pkt)

	if mock.markReadCalled != 42 {
		t.Errorf("MarkRead called with %d, want 42", mock.markReadCalled)
	}

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Response should have body data")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadMail_OutOfBounds(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{}
	server.mailRepo = mock
	session := createMockSession(1, server)
	// mailList is nil, so any AccIndex is out of bounds

	pkt := &mhfpacket.MsgMhfReadMail{
		AckHandle: 100,
		AccIndex:  5,
	}

	handleMsgMhfReadMail(session, pkt)

	select {
	case p := <-session.sendPackets:
		// Should get fallback single-byte response
		if len(p.data) == 0 {
			t.Error("Should have fallback response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfReadMail_ZeroMailID(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{}
	server.mailRepo = mock
	session := createMockSession(1, server)
	session.mailList = make([]int, 256)
	// mailList[0] is 0 (default)

	pkt := &mhfpacket.MsgMhfReadMail{
		AckHandle: 100,
		AccIndex:  0,
	}

	handleMsgMhfReadMail(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) == 0 {
			t.Error("Should have fallback response")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfOprtMail_Delete(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{
		mailByID: map[int]*Mail{
			42: {ID: 42},
		},
	}
	server.mailRepo = mock
	session := createMockSession(1, server)
	session.mailList = make([]int, 256)
	session.mailList[0] = 42

	pkt := &mhfpacket.MsgMhfOprtMail{
		AckHandle: 100,
		AccIndex:  0,
		Operation: mhfpacket.OperateMailDelete,
	}

	handleMsgMhfOprtMail(session, pkt)

	if mock.markDeletedID != 42 {
		t.Errorf("MarkDeleted called with %d, want 42", mock.markDeletedID)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfOprtMail_Lock(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{
		mailByID: map[int]*Mail{
			42: {ID: 42},
		},
	}
	server.mailRepo = mock
	session := createMockSession(1, server)
	session.mailList = make([]int, 256)
	session.mailList[0] = 42

	pkt := &mhfpacket.MsgMhfOprtMail{
		AckHandle: 100,
		AccIndex:  0,
		Operation: mhfpacket.OperateMailLock,
	}

	handleMsgMhfOprtMail(session, pkt)

	if mock.lockID != 42 || !mock.lockValue {
		t.Errorf("SetLocked called with ID=%d locked=%v, want ID=42 locked=true", mock.lockID, mock.lockValue)
	}
}

func TestHandleMsgMhfOprtMail_Unlock(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{
		mailByID: map[int]*Mail{
			42: {ID: 42},
		},
	}
	server.mailRepo = mock
	session := createMockSession(1, server)
	session.mailList = make([]int, 256)
	session.mailList[0] = 42

	pkt := &mhfpacket.MsgMhfOprtMail{
		AckHandle: 100,
		AccIndex:  0,
		Operation: mhfpacket.OperateMailUnlock,
	}

	handleMsgMhfOprtMail(session, pkt)

	if mock.lockID != 42 || mock.lockValue {
		t.Errorf("SetLocked called with ID=%d locked=%v, want ID=42 locked=false", mock.lockID, mock.lockValue)
	}
}

func TestHandleMsgMhfOprtMail_AcquireItem(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{
		mailByID: map[int]*Mail{
			42: {ID: 42, AttachedItemID: 100, AttachedItemAmount: 5},
		},
	}
	server.mailRepo = mock
	session := createMockSession(1, server)
	session.mailList = make([]int, 256)
	session.mailList[0] = 42

	pkt := &mhfpacket.MsgMhfOprtMail{
		AckHandle: 100,
		AccIndex:  0,
		Operation: mhfpacket.OperateMailAcquireItem,
	}

	handleMsgMhfOprtMail(session, pkt)

	if mock.itemReceivedID != 42 {
		t.Errorf("MarkItemReceived called with %d, want 42", mock.itemReceivedID)
	}
}

func TestHandleMsgMhfOprtMail_OutOfBounds(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{}
	server.mailRepo = mock
	session := createMockSession(1, server)
	// No mailList set

	pkt := &mhfpacket.MsgMhfOprtMail{
		AckHandle: 100,
		AccIndex:  5,
		Operation: mhfpacket.OperateMailDelete,
	}

	handleMsgMhfOprtMail(session, pkt)

	// Should not have called any repo methods
	if mock.markDeletedID != 0 {
		t.Error("Should not have called MarkDeleted for out-of-bounds access")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfSendMail_Direct(t *testing.T) {
	server := createMockServer()
	mock := &mockMailRepo{}
	server.mailRepo = mock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSendMail{
		AckHandle:   100,
		RecipientID: 42,
		Subject:     "Hello",
		Body:        "World",
		ItemID:      500,
		Quantity:    3,
	}

	handleMsgMhfSendMail(session, pkt)

	if len(mock.sentMails) != 1 {
		t.Fatalf("Expected 1 sent mail, got %d", len(mock.sentMails))
	}
	sent := mock.sentMails[0]
	if sent.senderID != 1 {
		t.Errorf("SenderID = %d, want 1", sent.senderID)
	}
	if sent.recipientID != 42 {
		t.Errorf("RecipientID = %d, want 42", sent.recipientID)
	}
	if sent.subject != "Hello" {
		t.Errorf("Subject = %s, want Hello", sent.subject)
	}
	if sent.itemID != 500 {
		t.Errorf("ItemID = %d, want 500", sent.itemID)
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestHandleMsgMhfSendMail_Guild(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepoForMail{
		guild: &Guild{ID: 10},
		members: []*GuildMember{
			{CharID: 100},
			{CharID: 200},
			{CharID: 300},
		},
	}
	server.mailRepo = mailMock
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSendMail{
		AckHandle:   100,
		RecipientID: 0, // 0 = guild mail
		Subject:     "Guild News",
		Body:        "Important update",
	}

	handleMsgMhfSendMail(session, pkt)

	if len(mailMock.sentMails) != 3 {
		t.Fatalf("Expected 3 sent mails (one per guild member), got %d", len(mailMock.sentMails))
	}
	for i, sent := range mailMock.sentMails {
		if sent.senderID != 1 {
			t.Errorf("Mail %d: SenderID = %d, want 1", i, sent.senderID)
		}
	}
	recipients := map[uint32]bool{}
	for _, sent := range mailMock.sentMails {
		recipients[sent.recipientID] = true
	}
	if !recipients[100] || !recipients[200] || !recipients[300] {
		t.Errorf("Expected recipients 100, 200, 300, got %v", recipients)
	}
}

func TestHandleMsgMhfSendMail_GuildNotFound(t *testing.T) {
	server := createMockServer()
	mailMock := &mockMailRepo{}
	guildMock := &mockGuildRepoForMail{getErr: errNotFound}
	server.mailRepo = mailMock
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfSendMail{
		AckHandle:   100,
		RecipientID: 0, // Guild mail
		Subject:     "Guild News",
		Body:        "Update",
	}

	handleMsgMhfSendMail(session, pkt)

	if len(mailMock.sentMails) != 0 {
		t.Errorf("No mails should be sent when guild not found, got %d", len(mailMock.sentMails))
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}
