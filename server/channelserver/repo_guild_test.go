package channelserver

import (
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
