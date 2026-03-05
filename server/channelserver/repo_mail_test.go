package channelserver

import (
	"testing"

	"github.com/jmoiron/sqlx"
)

func setupMailRepo(t *testing.T) (*MailRepository, *sqlx.DB, uint32, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "mail_sender")
	senderID := CreateTestCharacter(t, db, userID, "Sender")
	userID2 := CreateTestUser(t, db, "mail_recipient")
	recipientID := CreateTestCharacter(t, db, userID2, "Recipient")
	repo := NewMailRepository(db)
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, senderID, recipientID
}

func TestRepoMailSendMail(t *testing.T) {
	repo, db, senderID, recipientID := setupMailRepo(t)

	if err := repo.SendMail(senderID, recipientID, "Hello", "World", 0, 0, false, false); err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM mail WHERE sender_id=$1 AND recipient_id=$2", senderID, recipientID).Scan(&count); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 mail, got: %d", count)
	}
}

func TestRepoMailSendMailWithItem(t *testing.T) {
	repo, db, senderID, recipientID := setupMailRepo(t)

	if err := repo.SendMail(senderID, recipientID, "Gift", "Item for you", 100, 5, false, false); err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	var itemID, itemAmount int
	if err := db.QueryRow("SELECT attached_item, attached_item_amount FROM mail WHERE sender_id=$1", senderID).Scan(&itemID, &itemAmount); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if itemID != 100 || itemAmount != 5 {
		t.Errorf("Expected item=100 amount=5, got item=%d amount=%d", itemID, itemAmount)
	}
}

func TestRepoMailGetListForCharacter(t *testing.T) {
	repo, _, senderID, recipientID := setupMailRepo(t)

	if err := repo.SendMail(senderID, recipientID, "Mail1", "Body1", 0, 0, false, false); err != nil {
		t.Fatalf("SendMail 1 failed: %v", err)
	}
	if err := repo.SendMail(senderID, recipientID, "Mail2", "Body2", 0, 0, false, false); err != nil {
		t.Fatalf("SendMail 2 failed: %v", err)
	}

	mails, err := repo.GetListForCharacter(recipientID)
	if err != nil {
		t.Fatalf("GetListForCharacter failed: %v", err)
	}
	if len(mails) != 2 {
		t.Fatalf("Expected 2 mails, got: %d", len(mails))
	}
	// Should include sender name
	if mails[0].SenderName != "Sender" {
		t.Errorf("Expected sender_name='Sender', got: %q", mails[0].SenderName)
	}
}

func TestRepoMailGetListExcludesDeleted(t *testing.T) {
	repo, _, senderID, recipientID := setupMailRepo(t)

	if err := repo.SendMail(senderID, recipientID, "Visible", "", 0, 0, false, false); err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}
	if err := repo.SendMail(senderID, recipientID, "Deleted", "", 0, 0, false, false); err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	// Get the list and delete the second mail
	mails, _ := repo.GetListForCharacter(recipientID)
	if err := repo.MarkDeleted(mails[0].ID); err != nil {
		t.Fatalf("MarkDeleted failed: %v", err)
	}

	mails, err := repo.GetListForCharacter(recipientID)
	if err != nil {
		t.Fatalf("GetListForCharacter failed: %v", err)
	}
	if len(mails) != 1 {
		t.Fatalf("Expected 1 mail after deletion, got: %d", len(mails))
	}
}

func TestRepoMailGetByID(t *testing.T) {
	repo, db, senderID, recipientID := setupMailRepo(t)

	if err := repo.SendMail(senderID, recipientID, "Detail", "Full body text", 50, 2, true, false); err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	var mailID int
	if err := db.QueryRow("SELECT id FROM mail WHERE sender_id=$1", senderID).Scan(&mailID); err != nil {
		t.Fatalf("Setup query failed: %v", err)
	}

	mail, err := repo.GetByID(mailID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if mail.Subject != "Detail" {
		t.Errorf("Expected subject='Detail', got: %q", mail.Subject)
	}
	if mail.Body != "Full body text" {
		t.Errorf("Expected body='Full body text', got: %q", mail.Body)
	}
	if !mail.IsGuildInvite {
		t.Error("Expected is_guild_invite=true")
	}
	if mail.SenderName != "Sender" {
		t.Errorf("Expected sender_name='Sender', got: %q", mail.SenderName)
	}
}

func TestRepoMailMarkRead(t *testing.T) {
	repo, db, senderID, recipientID := setupMailRepo(t)

	if err := repo.SendMail(senderID, recipientID, "Unread", "", 0, 0, false, false); err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	var mailID int
	if err := db.QueryRow("SELECT id FROM mail WHERE sender_id=$1", senderID).Scan(&mailID); err != nil {
		t.Fatalf("Setup query failed: %v", err)
	}

	if err := repo.MarkRead(mailID); err != nil {
		t.Fatalf("MarkRead failed: %v", err)
	}

	var read bool
	if err := db.QueryRow("SELECT read FROM mail WHERE id=$1", mailID).Scan(&read); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !read {
		t.Error("Expected read=true")
	}
}

func TestRepoMailSetLocked(t *testing.T) {
	repo, db, senderID, recipientID := setupMailRepo(t)

	if err := repo.SendMail(senderID, recipientID, "Lock Test", "", 0, 0, false, false); err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	var mailID int
	if err := db.QueryRow("SELECT id FROM mail WHERE sender_id=$1", senderID).Scan(&mailID); err != nil {
		t.Fatalf("Setup query failed: %v", err)
	}

	if err := repo.SetLocked(mailID, true); err != nil {
		t.Fatalf("SetLocked failed: %v", err)
	}

	var locked bool
	if err := db.QueryRow("SELECT locked FROM mail WHERE id=$1", mailID).Scan(&locked); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !locked {
		t.Error("Expected locked=true")
	}

	// Unlock
	if err := repo.SetLocked(mailID, false); err != nil {
		t.Fatalf("SetLocked(false) failed: %v", err)
	}
	if err := db.QueryRow("SELECT locked FROM mail WHERE id=$1", mailID).Scan(&locked); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if locked {
		t.Error("Expected locked=false after unlock")
	}
}

func TestRepoMailMarkItemReceived(t *testing.T) {
	repo, db, senderID, recipientID := setupMailRepo(t)

	if err := repo.SendMail(senderID, recipientID, "Item Mail", "", 100, 1, false, false); err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	var mailID int
	if err := db.QueryRow("SELECT id FROM mail WHERE sender_id=$1", senderID).Scan(&mailID); err != nil {
		t.Fatalf("Setup query failed: %v", err)
	}

	if err := repo.MarkItemReceived(mailID); err != nil {
		t.Fatalf("MarkItemReceived failed: %v", err)
	}

	var received bool
	if err := db.QueryRow("SELECT attached_item_received FROM mail WHERE id=$1", mailID).Scan(&received); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !received {
		t.Error("Expected attached_item_received=true")
	}
}

func TestRepoMailSystemMessage(t *testing.T) {
	repo, db, senderID, recipientID := setupMailRepo(t)

	if err := repo.SendMail(senderID, recipientID, "System", "System alert", 0, 0, false, true); err != nil {
		t.Fatalf("SendMail failed: %v", err)
	}

	var isSys bool
	if err := db.QueryRow("SELECT is_sys_message FROM mail WHERE sender_id=$1", senderID).Scan(&isSys); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if !isSys {
		t.Error("Expected is_sys_message=true")
	}
}
