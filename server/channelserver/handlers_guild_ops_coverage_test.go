package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringsupport"
	"erupe-ce/network/mhfpacket"
	"testing"
)

func TestHandleRenamePugi_Pugi1(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	s := createMockSession(100, srv)

	bf := byteframe.NewByteFrame()
	nameBytes := stringsupport.UTF8ToSJIS("TestPugi")
	bf.WriteBytes(nameBytes)
	bf.WriteUint8(0) // null terminator
	bf.Seek(0, 0)

	handleRenamePugi(s, bf, guild, 1)
	if guild.PugiName1 != "TestPugi" {
		t.Errorf("PugiName1 = %q, want TestPugi", guild.PugiName1)
	}
}

func TestHandleRenamePugi_Pugi2(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	s := createMockSession(100, srv)

	bf := byteframe.NewByteFrame()
	nameBytes := stringsupport.UTF8ToSJIS("Pugi2")
	bf.WriteBytes(nameBytes)
	bf.WriteUint8(0)
	bf.Seek(0, 0)

	handleRenamePugi(s, bf, guild, 2)
	if guild.PugiName2 != "Pugi2" {
		t.Errorf("PugiName2 = %q, want Pugi2", guild.PugiName2)
	}
}

func TestHandleRenamePugi_Pugi3Default(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	s := createMockSession(100, srv)

	bf := byteframe.NewByteFrame()
	nameBytes := stringsupport.UTF8ToSJIS("Pugi3")
	bf.WriteBytes(nameBytes)
	bf.WriteUint8(0)
	bf.Seek(0, 0)

	handleRenamePugi(s, bf, guild, 3)
	if guild.PugiName3 != "Pugi3" {
		t.Errorf("PugiName3 = %q, want Pugi3", guild.PugiName3)
	}
}

func TestHandleChangePugi_AllNums(t *testing.T) {
	srv := createMockServer()
	guild := &Guild{ID: 1}
	srv.guildRepo = &mockGuildRepo{guild: guild}
	s := createMockSession(100, srv)

	handleChangePugi(s, 5, guild, 1)
	if guild.PugiOutfit1 != 5 {
		t.Errorf("PugiOutfit1 = %d, want 5", guild.PugiOutfit1)
	}

	handleChangePugi(s, 10, guild, 2)
	if guild.PugiOutfit2 != 10 {
		t.Errorf("PugiOutfit2 = %d, want 10", guild.PugiOutfit2)
	}

	handleChangePugi(s, 15, guild, 3)
	if guild.PugiOutfit3 != 15 {
		t.Errorf("PugiOutfit3 = %d, want 15", guild.PugiOutfit3)
	}
}

func TestHandleAvoidLeadershipUpdate_Success(t *testing.T) {
	srv := createMockServer()
	membership := &GuildMember{CharID: 100, AvoidLeadership: false}
	srv.guildRepo = &mockGuildRepo{membership: membership}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOperateGuild{AckHandle: 1}
	handleAvoidLeadershipUpdate(s, pkt, true)
	<-s.sendPackets

	if !membership.AvoidLeadership {
		t.Error("AvoidLeadership should be true")
	}
}

func TestHandleAvoidLeadershipUpdate_GetMembershipError(t *testing.T) {
	srv := createMockServer()
	srv.guildRepo = &mockGuildRepo{getMemberErr: errNotFound}
	s := createMockSession(100, srv)

	pkt := &mhfpacket.MsgMhfOperateGuild{AckHandle: 1}
	handleAvoidLeadershipUpdate(s, pkt, true)
	<-s.sendPackets
}
