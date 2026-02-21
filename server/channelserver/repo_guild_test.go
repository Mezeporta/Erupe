package channelserver

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
)

func setupGuildRepo(t *testing.T) (*GuildRepository, *sqlx.DB, uint32, uint32) {
	t.Helper()
	db := SetupTestDB(t)
	userID := CreateTestUser(t, db, "guild_test_user")
	charID := CreateTestCharacter(t, db, userID, "GuildLeader")
	repo := NewGuildRepository(db)
	guildID := CreateTestGuild(t, db, charID, "TestGuild")
	t.Cleanup(func() { TeardownTestDB(t, db) })
	return repo, db, guildID, charID
}

func TestGetByID(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	guild, err := repo.GetByID(guildID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if guild == nil {
		t.Fatal("Expected guild, got nil")
	}
	if guild.ID != guildID {
		t.Errorf("Expected guild ID %d, got %d", guildID, guild.ID)
	}
	if guild.Name != "TestGuild" {
		t.Errorf("Expected name 'TestGuild', got %q", guild.Name)
	}
	if guild.LeaderCharID != charID {
		t.Errorf("Expected leader %d, got %d", charID, guild.LeaderCharID)
	}
}

func TestGetByIDNotFound(t *testing.T) {
	repo, _, _, _ := setupGuildRepo(t)

	guild, err := repo.GetByID(999999)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}
	if guild != nil {
		t.Errorf("Expected nil for non-existent guild, got: %+v", guild)
	}
}

func TestGetByCharID(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	guild, err := repo.GetByCharID(charID)
	if err != nil {
		t.Fatalf("GetByCharID failed: %v", err)
	}
	if guild == nil {
		t.Fatal("Expected guild, got nil")
	}
	if guild.ID != guildID {
		t.Errorf("Expected guild ID %d, got %d", guildID, guild.ID)
	}
}

func TestGetByCharIDNotFound(t *testing.T) {
	repo, _, _, _ := setupGuildRepo(t)

	guild, err := repo.GetByCharID(999999)
	if err != nil {
		t.Fatalf("GetByCharID failed: %v", err)
	}
	if guild != nil {
		t.Errorf("Expected nil for non-member, got: %+v", guild)
	}
}

func TestCreate(t *testing.T) {
	db := SetupTestDB(t)
	defer TeardownTestDB(t, db)
	repo := NewGuildRepository(db)
	userID := CreateTestUser(t, db, "create_guild_user")
	charID := CreateTestCharacter(t, db, userID, "CreateLeader")

	guildID, err := repo.Create(charID, "NewGuild")
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if guildID <= 0 {
		t.Errorf("Expected positive guild ID, got %d", guildID)
	}

	// Verify guild exists
	guild, err := repo.GetByID(uint32(guildID))
	if err != nil {
		t.Fatalf("GetByID after Create failed: %v", err)
	}
	if guild == nil {
		t.Fatal("Created guild not found")
	}
	if guild.Name != "NewGuild" {
		t.Errorf("Expected name 'NewGuild', got %q", guild.Name)
	}

	// Verify leader is a member
	member, err := repo.GetCharacterMembership(charID)
	if err != nil {
		t.Fatalf("GetCharacterMembership failed: %v", err)
	}
	if member == nil {
		t.Fatal("Leader not found as guild member")
	}
}

func TestSaveGuild(t *testing.T) {
	repo, _, guildID, _ := setupGuildRepo(t)

	guild, err := repo.GetByID(guildID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	guild.Comment = "Updated comment"
	guild.MainMotto = 5
	guild.SubMotto = 3

	if err := repo.Save(guild); err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	updated, err := repo.GetByID(guildID)
	if err != nil {
		t.Fatalf("GetByID after Save failed: %v", err)
	}
	if updated.Comment != "Updated comment" {
		t.Errorf("Expected comment 'Updated comment', got %q", updated.Comment)
	}
	if updated.MainMotto != 5 || updated.SubMotto != 3 {
		t.Errorf("Expected mottos 5/3, got %d/%d", updated.MainMotto, updated.SubMotto)
	}
}

func TestDisband(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	if err := repo.Disband(guildID); err != nil {
		t.Fatalf("Disband failed: %v", err)
	}

	guild, err := repo.GetByID(guildID)
	if err != nil {
		t.Fatalf("GetByID after Disband failed: %v", err)
	}
	if guild != nil {
		t.Errorf("Expected nil after disband, got: %+v", guild)
	}

	member, err := repo.GetCharacterMembership(charID)
	if err != nil {
		t.Fatalf("GetCharacterMembership after Disband failed: %v", err)
	}
	if member != nil {
		t.Errorf("Expected nil membership after disband, got: %+v", member)
	}
}

func TestGetMembers(t *testing.T) {
	repo, db, guildID, leaderID := setupGuildRepo(t)

	// Add a second member
	user2 := CreateTestUser(t, db, "member_user")
	member2 := CreateTestCharacter(t, db, user2, "Member2")
	if _, err := db.Exec("INSERT INTO guild_characters (guild_id, character_id, order_index) VALUES ($1, $2, 2)", guildID, member2); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	members, err := repo.GetMembers(guildID, false)
	if err != nil {
		t.Fatalf("GetMembers failed: %v", err)
	}
	if len(members) != 2 {
		t.Fatalf("Expected 2 members, got %d", len(members))
	}

	ids := map[uint32]bool{leaderID: false, member2: false}
	for _, m := range members {
		ids[m.CharID] = true
	}
	if !ids[leaderID] || !ids[member2] {
		t.Errorf("Expected members %d and %d, got: %v", leaderID, member2, members)
	}
}

func TestGetCharacterMembership(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	member, err := repo.GetCharacterMembership(charID)
	if err != nil {
		t.Fatalf("GetCharacterMembership failed: %v", err)
	}
	if member == nil {
		t.Fatal("Expected membership, got nil")
	}
	if member.GuildID != guildID {
		t.Errorf("Expected guild ID %d, got %d", guildID, member.GuildID)
	}
	if !member.IsLeader {
		t.Error("Expected leader flag to be true")
	}
}

func TestSaveMember(t *testing.T) {
	repo, _, _, charID := setupGuildRepo(t)

	member, err := repo.GetCharacterMembership(charID)
	if err != nil {
		t.Fatalf("GetCharacterMembership failed: %v", err)
	}

	member.AvoidLeadership = true
	member.OrderIndex = 5

	if err := repo.SaveMember(member); err != nil {
		t.Fatalf("SaveMember failed: %v", err)
	}

	updated, err := repo.GetCharacterMembership(charID)
	if err != nil {
		t.Fatalf("GetCharacterMembership after Save failed: %v", err)
	}
	if !updated.AvoidLeadership {
		t.Error("Expected avoid_leadership=true")
	}
	if updated.OrderIndex != 5 {
		t.Errorf("Expected order_index=5, got %d", updated.OrderIndex)
	}
}

func TestRemoveCharacter(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	// Add and remove a member
	user2 := CreateTestUser(t, db, "remove_user")
	char2 := CreateTestCharacter(t, db, user2, "RemoveMe")
	if _, err := db.Exec("INSERT INTO guild_characters (guild_id, character_id, order_index) VALUES ($1, $2, 2)", guildID, char2); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	if err := repo.RemoveCharacter(char2); err != nil {
		t.Fatalf("RemoveCharacter failed: %v", err)
	}

	member, err := repo.GetCharacterMembership(char2)
	if err != nil {
		t.Fatalf("GetCharacterMembership after remove failed: %v", err)
	}
	if member != nil {
		t.Errorf("Expected nil membership after remove, got: %+v", member)
	}
}

func TestApplicationWorkflow(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	user2 := CreateTestUser(t, db, "applicant_user")
	applicantID := CreateTestCharacter(t, db, user2, "Applicant")

	// Create application
	err := repo.CreateApplication(guildID, applicantID, applicantID, GuildApplicationTypeApplied, nil)
	if err != nil {
		t.Fatalf("CreateApplication failed: %v", err)
	}

	// Check HasApplication
	has, err := repo.HasApplication(guildID, applicantID)
	if err != nil {
		t.Fatalf("HasApplication failed: %v", err)
	}
	if !has {
		t.Error("Expected application to exist")
	}

	// Get application
	app, err := repo.GetApplication(guildID, applicantID, GuildApplicationTypeApplied)
	if err != nil {
		t.Fatalf("GetApplication failed: %v", err)
	}
	if app == nil {
		t.Fatal("Expected application, got nil")
	}

	// Accept
	err = repo.AcceptApplication(guildID, applicantID)
	if err != nil {
		t.Fatalf("AcceptApplication failed: %v", err)
	}

	// Verify membership
	member, err := repo.GetCharacterMembership(applicantID)
	if err != nil {
		t.Fatalf("GetCharacterMembership after accept failed: %v", err)
	}
	if member == nil {
		t.Fatal("Expected membership after accept")
	}

	// Verify application removed
	has, err = repo.HasApplication(guildID, applicantID)
	if err != nil {
		t.Fatalf("HasApplication after accept failed: %v", err)
	}
	if has {
		t.Error("Expected no application after accept")
	}
}

func TestRejectApplication(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	user2 := CreateTestUser(t, db, "reject_user")
	applicantID := CreateTestCharacter(t, db, user2, "Rejected")

	err := repo.CreateApplication(guildID, applicantID, applicantID, GuildApplicationTypeApplied, nil)
	if err != nil {
		t.Fatalf("CreateApplication failed: %v", err)
	}

	err = repo.RejectApplication(guildID, applicantID)
	if err != nil {
		t.Fatalf("RejectApplication failed: %v", err)
	}

	has, err := repo.HasApplication(guildID, applicantID)
	if err != nil {
		t.Fatalf("HasApplication after reject failed: %v", err)
	}
	if has {
		t.Error("Expected no application after reject")
	}
}

func TestSetRecruiting(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	if err := repo.SetRecruiting(guildID, false); err != nil {
		t.Fatalf("SetRecruiting failed: %v", err)
	}

	var recruiting bool
	if err := db.QueryRow("SELECT recruiting FROM guilds WHERE id=$1", guildID).Scan(&recruiting); err != nil {
		t.Fatalf("Verification query failed: %v", err)
	}
	if recruiting {
		t.Error("Expected recruiting=false")
	}
}

func TestRPOperations(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	// AddRankRP
	if err := repo.AddRankRP(guildID, 100); err != nil {
		t.Fatalf("AddRankRP failed: %v", err)
	}
	var rankRP uint16
	if err := db.QueryRow("SELECT rank_rp FROM guilds WHERE id=$1", guildID).Scan(&rankRP); err != nil {
		t.Fatalf("Verification failed: %v", err)
	}
	if rankRP != 100 {
		t.Errorf("Expected rank_rp=100, got %d", rankRP)
	}

	// AddEventRP
	if err := repo.AddEventRP(guildID, 50); err != nil {
		t.Fatalf("AddEventRP failed: %v", err)
	}

	// ExchangeEventRP
	balance, err := repo.ExchangeEventRP(guildID, 20)
	if err != nil {
		t.Fatalf("ExchangeEventRP failed: %v", err)
	}
	if balance != 30 {
		t.Errorf("Expected event_rp balance=30, got %d", balance)
	}

	// Room RP operations
	if err := repo.AddRoomRP(guildID, 10); err != nil {
		t.Fatalf("AddRoomRP failed: %v", err)
	}
	roomRP, err := repo.GetRoomRP(guildID)
	if err != nil {
		t.Fatalf("GetRoomRP failed: %v", err)
	}
	if roomRP != 10 {
		t.Errorf("Expected room_rp=10, got %d", roomRP)
	}

	if err := repo.SetRoomRP(guildID, 0); err != nil {
		t.Fatalf("SetRoomRP failed: %v", err)
	}
	roomRP, err = repo.GetRoomRP(guildID)
	if err != nil {
		t.Fatalf("GetRoomRP after reset failed: %v", err)
	}
	if roomRP != 0 {
		t.Errorf("Expected room_rp=0, got %d", roomRP)
	}

	// SetRoomExpiry
	expiry := time.Date(2025, 7, 1, 0, 0, 0, 0, time.UTC)
	if err := repo.SetRoomExpiry(guildID, expiry); err != nil {
		t.Fatalf("SetRoomExpiry failed: %v", err)
	}
	var gotExpiry time.Time
	if err := db.QueryRow("SELECT room_expiry FROM guilds WHERE id=$1", guildID).Scan(&gotExpiry); err != nil {
		t.Fatalf("Verification failed: %v", err)
	}
	if !gotExpiry.Equal(expiry) {
		t.Errorf("Expected expiry %v, got %v", expiry, gotExpiry)
	}
}

func TestItemBox(t *testing.T) {
	repo, _, guildID, _ := setupGuildRepo(t)

	// Initially nil
	data, err := repo.GetItemBox(guildID)
	if err != nil {
		t.Fatalf("GetItemBox failed: %v", err)
	}
	if data != nil {
		t.Errorf("Expected nil item box initially, got %x", data)
	}

	// Save and retrieve
	blob := []byte{0x01, 0x02, 0x03}
	if err := repo.SaveItemBox(guildID, blob); err != nil {
		t.Fatalf("SaveItemBox failed: %v", err)
	}

	data, err = repo.GetItemBox(guildID)
	if err != nil {
		t.Fatalf("GetItemBox after save failed: %v", err)
	}
	if len(data) != 3 || data[0] != 0x01 || data[2] != 0x03 {
		t.Errorf("Expected %x, got %x", blob, data)
	}
}

func TestListAll(t *testing.T) {
	repo, db, _, _ := setupGuildRepo(t)

	// Create a second guild
	user2 := CreateTestUser(t, db, "list_user")
	char2 := CreateTestCharacter(t, db, user2, "ListLeader")
	CreateTestGuild(t, db, char2, "SecondGuild")

	guilds, err := repo.ListAll()
	if err != nil {
		t.Fatalf("ListAll failed: %v", err)
	}
	if len(guilds) < 2 {
		t.Errorf("Expected at least 2 guilds, got %d", len(guilds))
	}
}

func TestArrangeCharacters(t *testing.T) {
	repo, db, guildID, leaderID := setupGuildRepo(t)

	// Add two more members
	user2 := CreateTestUser(t, db, "arrange_user2")
	char2 := CreateTestCharacter(t, db, user2, "Char2")
	user3 := CreateTestUser(t, db, "arrange_user3")
	char3 := CreateTestCharacter(t, db, user3, "Char3")
	if _, err := db.Exec("INSERT INTO guild_characters (guild_id, character_id, order_index) VALUES ($1, $2, 2)", guildID, char2); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}
	if _, err := db.Exec("INSERT INTO guild_characters (guild_id, character_id, order_index) VALUES ($1, $2, 3)", guildID, char3); err != nil {
		t.Fatalf("Failed to add member: %v", err)
	}

	// Rearrange (excludes leader, sets order_index starting at 2)
	if err := repo.ArrangeCharacters([]uint32{char3, char2}); err != nil {
		t.Fatalf("ArrangeCharacters failed: %v", err)
	}

	// Verify order changed
	var order2, order3 uint16
	_ = db.QueryRow("SELECT order_index FROM guild_characters WHERE character_id=$1", char2).Scan(&order2)
	_ = db.QueryRow("SELECT order_index FROM guild_characters WHERE character_id=$1", char3).Scan(&order3)
	if order3 != 2 || order2 != 3 {
		t.Errorf("Expected char3=2, char2=3 but got char3=%d, char2=%d", order3, order2)
	}
	_ = leaderID
}

func TestSetRecruiter(t *testing.T) {
	repo, db, _, charID := setupGuildRepo(t)

	if err := repo.SetRecruiter(charID, true); err != nil {
		t.Fatalf("SetRecruiter failed: %v", err)
	}

	var recruiter bool
	if err := db.QueryRow("SELECT recruiter FROM guild_characters WHERE character_id=$1", charID).Scan(&recruiter); err != nil {
		t.Fatalf("Verification failed: %v", err)
	}
	if !recruiter {
		t.Error("Expected recruiter=true")
	}
}

func TestAddMemberDailyRP(t *testing.T) {
	repo, db, _, charID := setupGuildRepo(t)

	if err := repo.AddMemberDailyRP(charID, 25); err != nil {
		t.Fatalf("AddMemberDailyRP failed: %v", err)
	}

	var rp uint16
	if err := db.QueryRow("SELECT rp_today FROM guild_characters WHERE character_id=$1", charID).Scan(&rp); err != nil {
		t.Fatalf("Verification failed: %v", err)
	}
	if rp != 25 {
		t.Errorf("Expected rp_today=25, got %d", rp)
	}
}

// --- Invitation / Scout tests ---

func TestCancelInvitation(t *testing.T) {
	repo, db, guildID, leaderID := setupGuildRepo(t)

	user2 := CreateTestUser(t, db, "invite_user")
	char2 := CreateTestCharacter(t, db, user2, "Invited")

	if err := repo.CreateApplication(guildID, char2, leaderID, GuildApplicationTypeInvited, nil); err != nil {
		t.Fatalf("CreateApplication (invited) failed: %v", err)
	}

	if err := repo.CancelInvitation(guildID, char2); err != nil {
		t.Fatalf("CancelInvitation failed: %v", err)
	}

	has, err := repo.HasApplication(guildID, char2)
	if err != nil {
		t.Fatalf("HasApplication failed: %v", err)
	}
	if has {
		t.Error("Expected no application after cancellation")
	}
}

func TestListInvitedCharacters(t *testing.T) {
	repo, db, guildID, leaderID := setupGuildRepo(t)

	user2 := CreateTestUser(t, db, "scout_user")
	char2 := CreateTestCharacter(t, db, user2, "Scouted")

	if err := repo.CreateApplication(guildID, char2, leaderID, GuildApplicationTypeInvited, nil); err != nil {
		t.Fatalf("CreateApplication failed: %v", err)
	}

	chars, err := repo.ListInvitedCharacters(guildID)
	if err != nil {
		t.Fatalf("ListInvitedCharacters failed: %v", err)
	}
	if len(chars) != 1 {
		t.Fatalf("Expected 1 invited character, got %d", len(chars))
	}
	if chars[0].CharID != char2 {
		t.Errorf("Expected char ID %d, got %d", char2, chars[0].CharID)
	}
	if chars[0].Name != "Scouted" {
		t.Errorf("Expected name 'Scouted', got %q", chars[0].Name)
	}
	if chars[0].ActorID != leaderID {
		t.Errorf("Expected actor ID %d, got %d", leaderID, chars[0].ActorID)
	}
}

func TestListInvitedCharactersEmpty(t *testing.T) {
	repo, _, guildID, _ := setupGuildRepo(t)

	chars, err := repo.ListInvitedCharacters(guildID)
	if err != nil {
		t.Fatalf("ListInvitedCharacters failed: %v", err)
	}
	if len(chars) != 0 {
		t.Errorf("Expected 0 invited characters, got %d", len(chars))
	}
}

func TestGetByCharIDWithApplication(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	user2 := CreateTestUser(t, db, "app_char_user")
	char2 := CreateTestCharacter(t, db, user2, "Applicant2")

	if err := repo.CreateApplication(guildID, char2, char2, GuildApplicationTypeApplied, nil); err != nil {
		t.Fatalf("CreateApplication failed: %v", err)
	}

	guild, err := repo.GetByCharID(char2)
	if err != nil {
		t.Fatalf("GetByCharID failed: %v", err)
	}
	if guild == nil {
		t.Fatal("Expected guild via application, got nil")
	}
	if guild.ID != guildID {
		t.Errorf("Expected guild ID %d, got %d", guildID, guild.ID)
	}
}

func TestGetMembersApplicants(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	user2 := CreateTestUser(t, db, "applicant_member_user")
	char2 := CreateTestCharacter(t, db, user2, "AppMember")

	if err := repo.CreateApplication(guildID, char2, char2, GuildApplicationTypeApplied, nil); err != nil {
		t.Fatalf("CreateApplication failed: %v", err)
	}

	applicants, err := repo.GetMembers(guildID, true)
	if err != nil {
		t.Fatalf("GetMembers(applicants=true) failed: %v", err)
	}
	if len(applicants) != 1 {
		t.Fatalf("Expected 1 applicant, got %d", len(applicants))
	}
	if applicants[0].CharID != char2 {
		t.Errorf("Expected applicant char ID %d, got %d", char2, applicants[0].CharID)
	}
	if !applicants[0].IsApplicant {
		t.Error("Expected IsApplicant=true")
	}
}

// --- SetPugiOutfits ---

func TestSetPugiOutfits(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	if err := repo.SetPugiOutfits(guildID, 0xFF); err != nil {
		t.Fatalf("SetPugiOutfits failed: %v", err)
	}

	var outfits uint32
	if err := db.QueryRow("SELECT pugi_outfits FROM guilds WHERE id=$1", guildID).Scan(&outfits); err != nil {
		t.Fatalf("Verification failed: %v", err)
	}
	if outfits != 0xFF {
		t.Errorf("Expected pugi_outfits=0xFF, got %d", outfits)
	}
}

// --- Guild Posts ---

func TestCreateAndListPosts(t *testing.T) {
	repo, db, guildID, charID := setupGuildRepo(t)
	_ = db

	if err := repo.CreatePost(guildID, charID, 1, 0, "Hello", "World", 10); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}
	if err := repo.CreatePost(guildID, charID, 2, 0, "Second", "Post", 10); err != nil {
		t.Fatalf("CreatePost 2 failed: %v", err)
	}

	posts, err := repo.ListPosts(guildID, 0)
	if err != nil {
		t.Fatalf("ListPosts failed: %v", err)
	}
	if len(posts) != 2 {
		t.Fatalf("Expected 2 posts, got %d", len(posts))
	}
	// Newest first
	if posts[0].Title != "Second" {
		t.Errorf("Expected newest first, got %q", posts[0].Title)
	}
}

func TestCreatePostMaxPosts(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	// Create 3 posts with maxPosts=2 — the oldest should be soft-deleted
	for i := 0; i < 3; i++ {
		if err := repo.CreatePost(guildID, charID, 0, 0, fmt.Sprintf("Post%d", i), "body", 2); err != nil {
			t.Fatalf("CreatePost %d failed: %v", i, err)
		}
	}

	posts, err := repo.ListPosts(guildID, 0)
	if err != nil {
		t.Fatalf("ListPosts failed: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("Expected 2 posts after max enforcement, got %d", len(posts))
	}
}

func TestDeletePost(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	if err := repo.CreatePost(guildID, charID, 0, 0, "ToDelete", "body", 10); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}
	posts, _ := repo.ListPosts(guildID, 0)
	if len(posts) == 0 {
		t.Fatal("Expected post to exist")
	}

	if err := repo.DeletePost(posts[0].ID); err != nil {
		t.Fatalf("DeletePost failed: %v", err)
	}

	posts, _ = repo.ListPosts(guildID, 0)
	if len(posts) != 0 {
		t.Errorf("Expected 0 posts after delete, got %d", len(posts))
	}
}

func TestUpdatePost(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	if err := repo.CreatePost(guildID, charID, 0, 0, "Original", "body", 10); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}
	posts, _ := repo.ListPosts(guildID, 0)

	if err := repo.UpdatePost(posts[0].ID, "Updated", "new body"); err != nil {
		t.Fatalf("UpdatePost failed: %v", err)
	}

	posts, _ = repo.ListPosts(guildID, 0)
	if posts[0].Title != "Updated" || posts[0].Body != "new body" {
		t.Errorf("Expected 'Updated'/'new body', got %q/%q", posts[0].Title, posts[0].Body)
	}
}

func TestUpdatePostStamp(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	if err := repo.CreatePost(guildID, charID, 0, 0, "Stamp", "body", 10); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}
	posts, _ := repo.ListPosts(guildID, 0)

	if err := repo.UpdatePostStamp(posts[0].ID, 42); err != nil {
		t.Fatalf("UpdatePostStamp failed: %v", err)
	}

	posts, _ = repo.ListPosts(guildID, 0)
	if posts[0].StampID != 42 {
		t.Errorf("Expected stamp_id=42, got %d", posts[0].StampID)
	}
}

func TestPostLikedBy(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	if err := repo.CreatePost(guildID, charID, 0, 0, "Like", "body", 10); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}
	posts, _ := repo.ListPosts(guildID, 0)

	if err := repo.SetPostLikedBy(posts[0].ID, "100,200"); err != nil {
		t.Fatalf("SetPostLikedBy failed: %v", err)
	}

	liked, err := repo.GetPostLikedBy(posts[0].ID)
	if err != nil {
		t.Fatalf("GetPostLikedBy failed: %v", err)
	}
	if liked != "100,200" {
		t.Errorf("Expected '100,200', got %q", liked)
	}
}

func TestCountNewPosts(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	since := time.Now().Add(-1 * time.Hour)

	if err := repo.CreatePost(guildID, charID, 0, 0, "New", "body", 10); err != nil {
		t.Fatalf("CreatePost failed: %v", err)
	}

	count, err := repo.CountNewPosts(guildID, since)
	if err != nil {
		t.Fatalf("CountNewPosts failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 new post, got %d", count)
	}

	// Future time should yield 0
	count, err = repo.CountNewPosts(guildID, time.Now().Add(1*time.Hour))
	if err != nil {
		t.Fatalf("CountNewPosts (future) failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected 0 new posts with future time, got %d", count)
	}
}

func TestListPostsByType(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	if err := repo.CreatePost(guildID, charID, 0, 0, "TypeA", "body", 10); err != nil {
		t.Fatalf("CreatePost type 0 failed: %v", err)
	}
	if err := repo.CreatePost(guildID, charID, 0, 1, "TypeB", "body", 10); err != nil {
		t.Fatalf("CreatePost type 1 failed: %v", err)
	}

	posts0, _ := repo.ListPosts(guildID, 0)
	posts1, _ := repo.ListPosts(guildID, 1)
	if len(posts0) != 1 {
		t.Errorf("Expected 1 type-0 post, got %d", len(posts0))
	}
	if len(posts1) != 1 {
		t.Errorf("Expected 1 type-1 post, got %d", len(posts1))
	}
}

// --- Guild Alliances ---

func TestCreateAndGetAlliance(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	if err := repo.CreateAlliance("TestAlliance", guildID); err != nil {
		t.Fatalf("CreateAlliance failed: %v", err)
	}

	var allianceID uint32
	if err := db.QueryRow("SELECT id FROM guild_alliances WHERE parent_id=$1", guildID).Scan(&allianceID); err != nil {
		t.Fatalf("Alliance not found in DB: %v", err)
	}

	alliance, err := repo.GetAllianceByID(allianceID)
	if err != nil {
		t.Fatalf("GetAllianceByID failed: %v", err)
	}
	if alliance == nil {
		t.Fatal("Expected alliance, got nil")
	}
	if alliance.Name != "TestAlliance" {
		t.Errorf("Expected name 'TestAlliance', got %q", alliance.Name)
	}
	if alliance.ParentGuildID != guildID {
		t.Errorf("Expected parent guild %d, got %d", guildID, alliance.ParentGuildID)
	}
	if alliance.ParentGuild.ID != guildID {
		t.Errorf("Expected populated ParentGuild.ID=%d, got %d", guildID, alliance.ParentGuild.ID)
	}
}

func TestGetAllianceByIDNotFound(t *testing.T) {
	repo, _, _, _ := setupGuildRepo(t)

	alliance, err := repo.GetAllianceByID(999999)
	if err != nil {
		t.Fatalf("GetAllianceByID failed: %v", err)
	}
	if alliance != nil {
		t.Errorf("Expected nil for non-existent alliance, got: %+v", alliance)
	}
}

func TestListAlliances(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	if err := repo.CreateAlliance("Alliance1", guildID); err != nil {
		t.Fatalf("CreateAlliance failed: %v", err)
	}

	// Create a second guild and alliance
	user2 := CreateTestUser(t, db, "alliance_user2")
	char2 := CreateTestCharacter(t, db, user2, "AlliLeader2")
	guild2 := CreateTestGuild(t, db, char2, "AlliGuild2")
	if err := repo.CreateAlliance("Alliance2", guild2); err != nil {
		t.Fatalf("CreateAlliance 2 failed: %v", err)
	}

	alliances, err := repo.ListAlliances()
	if err != nil {
		t.Fatalf("ListAlliances failed: %v", err)
	}
	if len(alliances) < 2 {
		t.Errorf("Expected at least 2 alliances, got %d", len(alliances))
	}
}

func TestDeleteAlliance(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	if err := repo.CreateAlliance("ToDelete", guildID); err != nil {
		t.Fatalf("CreateAlliance failed: %v", err)
	}

	var allianceID uint32
	if err := db.QueryRow("SELECT id FROM guild_alliances WHERE parent_id=$1", guildID).Scan(&allianceID); err != nil {
		t.Fatalf("Alliance not found: %v", err)
	}

	if err := repo.DeleteAlliance(allianceID); err != nil {
		t.Fatalf("DeleteAlliance failed: %v", err)
	}

	alliance, err := repo.GetAllianceByID(allianceID)
	if err != nil {
		t.Fatalf("GetAllianceByID after delete failed: %v", err)
	}
	if alliance != nil {
		t.Errorf("Expected nil after delete, got: %+v", alliance)
	}
}

func TestRemoveGuildFromAllianceSub1(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	user2 := CreateTestUser(t, db, "alli_sub1_user")
	char2 := CreateTestCharacter(t, db, user2, "Sub1Leader")
	guild2 := CreateTestGuild(t, db, char2, "SubGuild1")

	if err := repo.CreateAlliance("AlliSub", guildID); err != nil {
		t.Fatalf("CreateAlliance failed: %v", err)
	}
	var allianceID uint32
	if err := db.QueryRow("SELECT id FROM guild_alliances WHERE parent_id=$1", guildID).Scan(&allianceID); err != nil {
		t.Fatalf("Failed to get alliance ID: %v", err)
	}

	// Add sub1
	if _, err := db.Exec("UPDATE guild_alliances SET sub1_id=$1 WHERE id=$2", guild2, allianceID); err != nil {
		t.Fatalf("Failed to set sub1: %v", err)
	}

	// Remove sub1
	if err := repo.RemoveGuildFromAlliance(allianceID, guild2, guild2, 0); err != nil {
		t.Fatalf("RemoveGuildFromAlliance failed: %v", err)
	}

	alliance, err := repo.GetAllianceByID(allianceID)
	if err != nil {
		t.Fatalf("GetAllianceByID failed: %v", err)
	}
	if alliance == nil {
		t.Fatal("Expected alliance to still exist")
	}
	if alliance.SubGuild1ID != 0 {
		t.Errorf("Expected sub1_id=0, got %d", alliance.SubGuild1ID)
	}
}

func TestRemoveGuildFromAllianceSub1ShiftsSub2(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	user2 := CreateTestUser(t, db, "alli_shift_user2")
	char2 := CreateTestCharacter(t, db, user2, "Shift2Leader")
	guild2 := CreateTestGuild(t, db, char2, "ShiftGuild2")

	user3 := CreateTestUser(t, db, "alli_shift_user3")
	char3 := CreateTestCharacter(t, db, user3, "Shift3Leader")
	guild3 := CreateTestGuild(t, db, char3, "ShiftGuild3")

	if err := repo.CreateAlliance("AlliShift", guildID); err != nil {
		t.Fatalf("CreateAlliance failed: %v", err)
	}
	var allianceID uint32
	if err := db.QueryRow("SELECT id FROM guild_alliances WHERE parent_id=$1", guildID).Scan(&allianceID); err != nil {
		t.Fatalf("Failed to get alliance ID: %v", err)
	}
	if _, err := db.Exec("UPDATE guild_alliances SET sub1_id=$1, sub2_id=$2 WHERE id=$3", guild2, guild3, allianceID); err != nil {
		t.Fatalf("Failed to set sub guilds: %v", err)
	}

	// Remove sub1 — sub2 should shift into sub1's slot
	if err := repo.RemoveGuildFromAlliance(allianceID, guild2, guild2, guild3); err != nil {
		t.Fatalf("RemoveGuildFromAlliance failed: %v", err)
	}

	alliance, err := repo.GetAllianceByID(allianceID)
	if err != nil {
		t.Fatalf("GetAllianceByID failed: %v", err)
	}
	if alliance == nil {
		t.Fatal("Expected alliance to still exist")
	}
	if alliance.SubGuild1ID != guild3 {
		t.Errorf("Expected sub1_id=%d (shifted from sub2), got %d", guild3, alliance.SubGuild1ID)
	}
	if alliance.SubGuild2ID != 0 {
		t.Errorf("Expected sub2_id=0, got %d", alliance.SubGuild2ID)
	}
}

func TestRemoveGuildFromAllianceSub2(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	user2 := CreateTestUser(t, db, "alli_s2_user2")
	char2 := CreateTestCharacter(t, db, user2, "S2Leader2")
	guild2 := CreateTestGuild(t, db, char2, "S2Guild2")

	user3 := CreateTestUser(t, db, "alli_s2_user3")
	char3 := CreateTestCharacter(t, db, user3, "S2Leader3")
	guild3 := CreateTestGuild(t, db, char3, "S2Guild3")

	if err := repo.CreateAlliance("AlliS2", guildID); err != nil {
		t.Fatalf("CreateAlliance failed: %v", err)
	}
	var allianceID uint32
	if err := db.QueryRow("SELECT id FROM guild_alliances WHERE parent_id=$1", guildID).Scan(&allianceID); err != nil {
		t.Fatalf("Failed to get alliance ID: %v", err)
	}
	if _, err := db.Exec("UPDATE guild_alliances SET sub1_id=$1, sub2_id=$2 WHERE id=$3", guild2, guild3, allianceID); err != nil {
		t.Fatalf("Failed to set sub guilds: %v", err)
	}

	// Remove sub2 directly
	if err := repo.RemoveGuildFromAlliance(allianceID, guild3, guild2, guild3); err != nil {
		t.Fatalf("RemoveGuildFromAlliance failed: %v", err)
	}

	alliance, err := repo.GetAllianceByID(allianceID)
	if err != nil {
		t.Fatalf("GetAllianceByID failed: %v", err)
	}
	if alliance == nil {
		t.Fatal("Expected alliance to still exist")
	}
	if alliance.SubGuild1ID != guild2 {
		t.Errorf("Expected sub1_id=%d unchanged, got %d", guild2, alliance.SubGuild1ID)
	}
	if alliance.SubGuild2ID != 0 {
		t.Errorf("Expected sub2_id=0, got %d", alliance.SubGuild2ID)
	}
}

// --- Guild Adventures ---

func TestCreateAndListAdventures(t *testing.T) {
	repo, _, guildID, _ := setupGuildRepo(t)

	if err := repo.CreateAdventure(guildID, 5, 1000, 2000); err != nil {
		t.Fatalf("CreateAdventure failed: %v", err)
	}

	adventures, err := repo.ListAdventures(guildID)
	if err != nil {
		t.Fatalf("ListAdventures failed: %v", err)
	}
	if len(adventures) != 1 {
		t.Fatalf("Expected 1 adventure, got %d", len(adventures))
	}
	if adventures[0].Destination != 5 {
		t.Errorf("Expected destination=5, got %d", adventures[0].Destination)
	}
	if adventures[0].Depart != 1000 {
		t.Errorf("Expected depart=1000, got %d", adventures[0].Depart)
	}
	if adventures[0].Return != 2000 {
		t.Errorf("Expected return=2000, got %d", adventures[0].Return)
	}
}

func TestCreateAdventureWithCharge(t *testing.T) {
	repo, _, guildID, _ := setupGuildRepo(t)

	if err := repo.CreateAdventureWithCharge(guildID, 3, 50, 1000, 2000); err != nil {
		t.Fatalf("CreateAdventureWithCharge failed: %v", err)
	}

	adventures, err := repo.ListAdventures(guildID)
	if err != nil {
		t.Fatalf("ListAdventures failed: %v", err)
	}
	if len(adventures) != 1 {
		t.Fatalf("Expected 1 adventure, got %d", len(adventures))
	}
	if adventures[0].Charge != 50 {
		t.Errorf("Expected charge=50, got %d", adventures[0].Charge)
	}
}

func TestChargeAdventure(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	if err := repo.CreateAdventure(guildID, 1, 1000, 2000); err != nil {
		t.Fatalf("CreateAdventure failed: %v", err)
	}
	adventures, _ := repo.ListAdventures(guildID)
	advID := adventures[0].ID

	if err := repo.ChargeAdventure(advID, 25); err != nil {
		t.Fatalf("ChargeAdventure failed: %v", err)
	}

	var charge uint32
	if err := db.QueryRow("SELECT charge FROM guild_adventures WHERE id=$1", advID).Scan(&charge); err != nil {
		t.Fatalf("Failed to get charge: %v", err)
	}
	if charge != 25 {
		t.Errorf("Expected charge=25, got %d", charge)
	}
}

func TestCollectAdventure(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	if err := repo.CreateAdventure(guildID, 1, 1000, 2000); err != nil {
		t.Fatalf("CreateAdventure failed: %v", err)
	}
	adventures, _ := repo.ListAdventures(guildID)
	advID := adventures[0].ID

	if err := repo.CollectAdventure(advID, charID); err != nil {
		t.Fatalf("CollectAdventure failed: %v", err)
	}

	// Verify collected_by updated
	adventures, _ = repo.ListAdventures(guildID)
	if adventures[0].CollectedBy == "" {
		t.Error("Expected collected_by to be non-empty")
	}
}

func TestListAdventuresEmpty(t *testing.T) {
	repo, _, guildID, _ := setupGuildRepo(t)

	adventures, err := repo.ListAdventures(guildID)
	if err != nil {
		t.Fatalf("ListAdventures failed: %v", err)
	}
	if len(adventures) != 0 {
		t.Errorf("Expected 0 adventures, got %d", len(adventures))
	}
}

// --- Guild Treasure Hunts ---

func TestCreateAndGetPendingHunt(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	huntData := []byte{0xAA, 0xBB, 0xCC}
	if err := repo.CreateHunt(guildID, charID, 10, 1, huntData, ""); err != nil {
		t.Fatalf("CreateHunt failed: %v", err)
	}

	hunt, err := repo.GetPendingHunt(charID)
	if err != nil {
		t.Fatalf("GetPendingHunt failed: %v", err)
	}
	if hunt == nil {
		t.Fatal("Expected pending hunt, got nil")
	}
	if hunt.HostID != charID {
		t.Errorf("Expected host_id=%d, got %d", charID, hunt.HostID)
	}
	if hunt.Destination != 10 {
		t.Errorf("Expected destination=10, got %d", hunt.Destination)
	}
	if hunt.Level != 1 {
		t.Errorf("Expected level=1, got %d", hunt.Level)
	}
	if len(hunt.HuntData) != 3 || hunt.HuntData[0] != 0xAA {
		t.Errorf("Expected hunt_data [AA BB CC], got %x", hunt.HuntData)
	}
}

func TestGetPendingHuntNone(t *testing.T) {
	repo, _, _, charID := setupGuildRepo(t)

	hunt, err := repo.GetPendingHunt(charID)
	if err != nil {
		t.Fatalf("GetPendingHunt failed: %v", err)
	}
	if hunt != nil {
		t.Errorf("Expected nil when no pending hunt, got: %+v", hunt)
	}
}

func TestAcquireHunt(t *testing.T) {
	repo, db, guildID, charID := setupGuildRepo(t)

	if err := repo.CreateHunt(guildID, charID, 10, 2, nil, ""); err != nil {
		t.Fatalf("CreateHunt failed: %v", err)
	}
	hunt, _ := repo.GetPendingHunt(charID)

	if err := repo.AcquireHunt(hunt.HuntID); err != nil {
		t.Fatalf("AcquireHunt failed: %v", err)
	}

	// After acquiring, it should no longer appear as pending
	pending, _ := repo.GetPendingHunt(charID)
	if pending != nil {
		t.Error("Expected no pending hunt after acquire")
	}

	// Verify in DB
	var acquired bool
	if err := db.QueryRow("SELECT acquired FROM guild_hunts WHERE id=$1", hunt.HuntID).Scan(&acquired); err != nil {
		t.Fatalf("Failed to get acquired: %v", err)
	}
	if !acquired {
		t.Error("Expected acquired=true in DB")
	}
}

func TestListGuildHunts(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	// Create a level-2 hunt and acquire it
	if err := repo.CreateHunt(guildID, charID, 10, 2, []byte{0x01}, ""); err != nil {
		t.Fatalf("CreateHunt failed: %v", err)
	}
	hunt, _ := repo.GetPendingHunt(charID)
	if err := repo.AcquireHunt(hunt.HuntID); err != nil {
		t.Fatalf("AcquireHunt failed: %v", err)
	}

	// Create a level-1 hunt (should not appear)
	if err := repo.CreateHunt(guildID, charID, 20, 1, nil, ""); err != nil {
		t.Fatalf("CreateHunt level-1 failed: %v", err)
	}

	hunts, err := repo.ListGuildHunts(guildID, charID)
	if err != nil {
		t.Fatalf("ListGuildHunts failed: %v", err)
	}
	if len(hunts) != 1 {
		t.Fatalf("Expected 1 acquired level-2 hunt, got %d", len(hunts))
	}
	if hunts[0].Destination != 10 {
		t.Errorf("Expected destination=10, got %d", hunts[0].Destination)
	}
}

func TestRegisterHuntReport(t *testing.T) {
	repo, db, guildID, charID := setupGuildRepo(t)

	if err := repo.CreateHunt(guildID, charID, 10, 2, nil, ""); err != nil {
		t.Fatalf("CreateHunt failed: %v", err)
	}
	hunt, _ := repo.GetPendingHunt(charID)

	if err := repo.RegisterHuntReport(hunt.HuntID, charID); err != nil {
		t.Fatalf("RegisterHuntReport failed: %v", err)
	}

	var treasureHunt *uint32
	if err := db.QueryRow("SELECT treasure_hunt FROM guild_characters WHERE character_id=$1", charID).Scan(&treasureHunt); err != nil {
		t.Fatalf("Failed to get treasure_hunt: %v", err)
	}
	if treasureHunt == nil || *treasureHunt != hunt.HuntID {
		t.Errorf("Expected treasure_hunt=%d, got %v", hunt.HuntID, treasureHunt)
	}
}

func TestCollectHunt(t *testing.T) {
	repo, db, guildID, charID := setupGuildRepo(t)

	if err := repo.CreateHunt(guildID, charID, 10, 2, nil, ""); err != nil {
		t.Fatalf("CreateHunt failed: %v", err)
	}
	hunt, _ := repo.GetPendingHunt(charID)
	if err := repo.RegisterHuntReport(hunt.HuntID, charID); err != nil {
		t.Fatalf("RegisterHuntReport failed: %v", err)
	}

	if err := repo.CollectHunt(hunt.HuntID); err != nil {
		t.Fatalf("CollectHunt failed: %v", err)
	}

	// Hunt should be marked collected
	var collected bool
	if err := db.QueryRow("SELECT collected FROM guild_hunts WHERE id=$1", hunt.HuntID).Scan(&collected); err != nil {
		t.Fatalf("Failed to scan collected: %v", err)
	}
	if !collected {
		t.Error("Expected collected=true")
	}

	// Character's treasure_hunt should be cleared
	var treasureHunt *uint32
	if err := db.QueryRow("SELECT treasure_hunt FROM guild_characters WHERE character_id=$1", charID).Scan(&treasureHunt); err != nil {
		t.Fatalf("Failed to get treasure_hunt: %v", err)
	}
	if treasureHunt != nil {
		t.Errorf("Expected treasure_hunt=NULL, got %v", *treasureHunt)
	}
}

func TestClaimHuntReward(t *testing.T) {
	repo, db, guildID, charID := setupGuildRepo(t)

	if err := repo.CreateHunt(guildID, charID, 10, 2, nil, ""); err != nil {
		t.Fatalf("CreateHunt failed: %v", err)
	}
	hunt, _ := repo.GetPendingHunt(charID)

	if err := repo.ClaimHuntReward(hunt.HuntID, charID); err != nil {
		t.Fatalf("ClaimHuntReward failed: %v", err)
	}

	var count int
	if err := db.QueryRow("SELECT COUNT(*) FROM guild_hunts_claimed WHERE hunt_id=$1 AND character_id=$2", hunt.HuntID, charID).Scan(&count); err != nil {
		t.Fatalf("Failed to scan claimed count: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected 1 claimed entry, got %d", count)
	}
}

// --- Guild Meals ---

func TestCreateAndListMeals(t *testing.T) {
	repo, _, guildID, _ := setupGuildRepo(t)

	now := time.Now().UTC().Truncate(time.Second)
	id, err := repo.CreateMeal(guildID, 5, 3, now)
	if err != nil {
		t.Fatalf("CreateMeal failed: %v", err)
	}
	if id == 0 {
		t.Error("Expected non-zero meal ID")
	}

	meals, err := repo.ListMeals(guildID)
	if err != nil {
		t.Fatalf("ListMeals failed: %v", err)
	}
	if len(meals) != 1 {
		t.Fatalf("Expected 1 meal, got %d", len(meals))
	}
	if meals[0].MealID != 5 {
		t.Errorf("Expected meal_id=5, got %d", meals[0].MealID)
	}
	if meals[0].Level != 3 {
		t.Errorf("Expected level=3, got %d", meals[0].Level)
	}
}

func TestUpdateMeal(t *testing.T) {
	repo, _, guildID, _ := setupGuildRepo(t)

	now := time.Now().UTC().Truncate(time.Second)
	id, _ := repo.CreateMeal(guildID, 5, 3, now)

	later := now.Add(30 * time.Minute)
	if err := repo.UpdateMeal(id, 10, 5, later); err != nil {
		t.Fatalf("UpdateMeal failed: %v", err)
	}

	meals, _ := repo.ListMeals(guildID)
	if meals[0].MealID != 10 {
		t.Errorf("Expected meal_id=10, got %d", meals[0].MealID)
	}
	if meals[0].Level != 5 {
		t.Errorf("Expected level=5, got %d", meals[0].Level)
	}
}

func TestListMealsEmpty(t *testing.T) {
	repo, _, guildID, _ := setupGuildRepo(t)

	meals, err := repo.ListMeals(guildID)
	if err != nil {
		t.Fatalf("ListMeals failed: %v", err)
	}
	if len(meals) != 0 {
		t.Errorf("Expected 0 meals, got %d", len(meals))
	}
}

// --- Kill tracking ---

func TestClaimHuntBox(t *testing.T) {
	repo, db, _, charID := setupGuildRepo(t)

	claimedAt := time.Now().UTC().Truncate(time.Second)
	if err := repo.ClaimHuntBox(charID, claimedAt); err != nil {
		t.Fatalf("ClaimHuntBox failed: %v", err)
	}

	var got time.Time
	if err := db.QueryRow("SELECT box_claimed FROM guild_characters WHERE character_id=$1", charID).Scan(&got); err != nil {
		t.Fatalf("Failed to scan box_claimed: %v", err)
	}
	if !got.Equal(claimedAt) {
		t.Errorf("Expected box_claimed=%v, got %v", claimedAt, got)
	}
}

func TestListAndCountGuildKills(t *testing.T) {
	repo, db, guildID, charID := setupGuildRepo(t)

	// Set box_claimed to the past so kills after it are visible
	past := time.Now().Add(-1 * time.Hour).UTC().Truncate(time.Second)
	if err := repo.ClaimHuntBox(charID, past); err != nil {
		t.Fatalf("ClaimHuntBox failed: %v", err)
	}

	// Insert kill logs for this character
	if _, err := db.Exec("INSERT INTO kill_logs (character_id, monster, quantity, timestamp) VALUES ($1, 100, 1, NOW())", charID); err != nil {
		t.Fatalf("Failed to insert kill log: %v", err)
	}
	if _, err := db.Exec("INSERT INTO kill_logs (character_id, monster, quantity, timestamp) VALUES ($1, 200, 1, NOW())", charID); err != nil {
		t.Fatalf("Failed to insert kill log: %v", err)
	}

	kills, err := repo.ListGuildKills(guildID, charID)
	if err != nil {
		t.Fatalf("ListGuildKills failed: %v", err)
	}
	if len(kills) != 2 {
		t.Fatalf("Expected 2 kills, got %d", len(kills))
	}

	count, err := repo.CountGuildKills(guildID, charID)
	if err != nil {
		t.Fatalf("CountGuildKills failed: %v", err)
	}
	if count != 2 {
		t.Errorf("Expected count=2, got %d", count)
	}
}

func TestListGuildKillsEmpty(t *testing.T) {
	repo, _, guildID, charID := setupGuildRepo(t)

	// Set box_claimed to now — no kills after it
	if err := repo.ClaimHuntBox(charID, time.Now().UTC()); err != nil {
		t.Fatalf("ClaimHuntBox failed: %v", err)
	}

	kills, err := repo.ListGuildKills(guildID, charID)
	if err != nil {
		t.Fatalf("ListGuildKills failed: %v", err)
	}
	if len(kills) != 0 {
		t.Errorf("Expected 0 kills, got %d", len(kills))
	}

	count, err := repo.CountGuildKills(guildID, charID)
	if err != nil {
		t.Fatalf("CountGuildKills failed: %v", err)
	}
	if count != 0 {
		t.Errorf("Expected count=0, got %d", count)
	}
}

// --- Disband with alliance cleanup ---

func TestDisbandCleansUpAlliance(t *testing.T) {
	repo, db, guildID, _ := setupGuildRepo(t)

	// Create alliance with this guild as parent
	if err := repo.CreateAlliance("DisbandAlliance", guildID); err != nil {
		t.Fatalf("CreateAlliance failed: %v", err)
	}

	var allianceID uint32
	if err := db.QueryRow("SELECT id FROM guild_alliances WHERE parent_id=$1", guildID).Scan(&allianceID); err != nil {
		t.Fatalf("Failed to scan alliance ID: %v", err)
	}

	if err := repo.Disband(guildID); err != nil {
		t.Fatalf("Disband failed: %v", err)
	}

	// Alliance should be deleted too (parent_id match in Disband)
	alliance, _ := repo.GetAllianceByID(allianceID)
	if alliance != nil {
		t.Errorf("Expected alliance to be deleted after parent guild disband, got: %+v", alliance)
	}
}
