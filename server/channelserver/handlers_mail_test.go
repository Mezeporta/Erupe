package channelserver

import (
	"testing"
	"time"
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
