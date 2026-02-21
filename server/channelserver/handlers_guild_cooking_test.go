package channelserver

import (
	"testing"
	"time"

	"erupe-ce/network/mhfpacket"
)

// --- handleMsgMhfLoadGuildCooking tests ---

func TestLoadGuildCooking_NoMeals(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		meals: []*GuildMeal{},
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadGuildCooking{AckHandle: 100}

	handleMsgMhfLoadGuildCooking(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestLoadGuildCooking_WithActiveMeals(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		meals: []*GuildMeal{
			{ID: 1, MealID: 100, Level: 3, CreatedAt: TimeAdjusted()},                        // active (within 60 min)
			{ID: 2, MealID: 200, Level: 1, CreatedAt: TimeAdjusted().Add(-2 * time.Hour)},     // expired
		},
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadGuildCooking{AckHandle: 100}

	handleMsgMhfLoadGuildCooking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 4 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestLoadGuildCooking_DBError(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		listMealsErr: errNotFound,
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfLoadGuildCooking{AckHandle: 100}

	handleMsgMhfLoadGuildCooking(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfRegistGuildCooking tests ---

func TestRegistGuildCooking_NewMeal(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		createdMealID: 42,
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfRegistGuildCooking{
		AckHandle:   100,
		OverwriteID: 0, // New meal
		MealID:      5,
		Success:     1,
	}

	handleMsgMhfRegistGuildCooking(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 8 {
			t.Errorf("Response too short: %d bytes", len(p.data))
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestRegistGuildCooking_UpdateMeal(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfRegistGuildCooking{
		AckHandle:   100,
		OverwriteID: 42, // Update existing
		MealID:      5,
		Success:     2,
	}

	handleMsgMhfRegistGuildCooking(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestRegistGuildCooking_CreateError(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		createMealErr: errNotFound,
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfRegistGuildCooking{
		AckHandle:   100,
		OverwriteID: 0,
		MealID:      5,
		Success:     1,
	}

	handleMsgMhfRegistGuildCooking(session, pkt)

	// Should return fail ack
	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfGuildHuntdata tests ---

func TestGuildHuntdata_Acquire(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGuildHuntdata{
		AckHandle: 100,
		Operation: 0, // Acquire
		GuildID:   10,
	}

	handleMsgMhfGuildHuntdata(session, pkt)

	if !guildMock.claimBoxCalled {
		t.Error("ClaimHuntBox should be called")
	}

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGuildHuntdata_Enumerate(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		guildKills: []*GuildKill{
			{ID: 1, Monster: 100},
			{ID: 2, Monster: 200},
		},
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGuildHuntdata{
		AckHandle: 100,
		Operation: 1, // Enumerate
		GuildID:   10,
	}

	handleMsgMhfGuildHuntdata(session, pkt)

	select {
	case p := <-session.sendPackets:
		if len(p.data) < 1 {
			t.Fatal("Response too short")
		}
	default:
		t.Error("No response packet queued")
	}
}

func TestGuildHuntdata_Check_HasKills(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		countKills: 5,
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGuildHuntdata{
		AckHandle: 100,
		Operation: 2, // Check
		GuildID:   10,
	}

	handleMsgMhfGuildHuntdata(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestGuildHuntdata_Check_NoKills(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{
		countKills: 0,
	}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfGuildHuntdata{
		AckHandle: 100,
		Operation: 2,
		GuildID:   10,
	}

	handleMsgMhfGuildHuntdata(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

// --- handleMsgMhfAddGuildWeeklyBonusExceptionalUser tests ---

func TestAddGuildWeeklyBonusExceptionalUser_Success(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.guild = &Guild{ID: 10}
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAddGuildWeeklyBonusExceptionalUser{
		AckHandle: 100,
		NumUsers:  3,
	}

	handleMsgMhfAddGuildWeeklyBonusExceptionalUser(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}

func TestAddGuildWeeklyBonusExceptionalUser_NoGuild(t *testing.T) {
	server := createMockServer()
	guildMock := &mockGuildRepoOps{}
	guildMock.getErr = errNotFound
	server.guildRepo = guildMock
	session := createMockSession(1, server)

	pkt := &mhfpacket.MsgMhfAddGuildWeeklyBonusExceptionalUser{
		AckHandle: 100,
		NumUsers:  3,
	}

	// Should not panic; just skips the bonus
	handleMsgMhfAddGuildWeeklyBonusExceptionalUser(session, pkt)

	select {
	case <-session.sendPackets:
	default:
		t.Error("No response packet queued")
	}
}
