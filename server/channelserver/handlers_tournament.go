package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"time"

	"go.uber.org/zap"
)

// TournamentInfo0 represents tournament information (type 0).
type TournamentInfo0 struct {
	ID             uint32
	MaxPlayers     uint32
	CurrentPlayers uint32
	Unk1           uint16
	TextColor      uint16
	Unk2           uint32
	Time1          time.Time
	Time2          time.Time
	Time3          time.Time
	Time4          time.Time
	Time5          time.Time
	Time6          time.Time
	Unk3           uint8
	Unk4           uint8
	MinHR          uint32
	MaxHR          uint32
	Unk5           string
	Unk6           string
}

// TournamentInfo21 represents tournament information (type 21).
type TournamentInfo21 struct {
	Unk0 uint32
	Unk1 uint32
	Unk2 uint32
	Unk3 uint8
}

// TournamentInfo22 represents tournament information (type 22).
type TournamentInfo22 struct {
	Unk0 uint32
	Unk1 uint32
	Unk2 uint32
	Unk3 uint8
	Unk4 string
}

// TournamentReward represents a tournament reward entry.
type TournamentReward struct {
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
}

// tournamentState returns the state byte for the EnumerateRanking response.
// 0 = no tournament / before start, 1 = registration open, 2 = hunting active,
// 3 = ranking/reward period.
func tournamentState(now int64, t *Tournament) uint8 {
	if t == nil || now < t.StartTime {
		return 0
	}
	if now <= t.EntryEnd {
		return 1
	}
	if now <= t.RankingEnd {
		return 2
	}
	return 3
}

func handleMsgMhfEnumerateRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRanking)
	bf := byteframe.NewByteFrame()

	now := TimeAdjusted().Unix()
	tournament, err := s.server.tournamentRepo.GetActive(now)
	if err != nil {
		s.logger.Error("Failed to get active tournament for EnumerateRanking", zap.Error(err))
	}

	if tournament == nil {
		// No active tournament: write zeroed timestamps, current time, state 0, empty data.
		bf.WriteBytes(make([]byte, 16))
		bf.WriteUint32(uint32(now))
		bf.WriteUint8(0)
		ps.Uint8(bf, "", false)
		bf.WriteUint16(0) // numEvents
		bf.WriteUint8(0)  // numCups
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}

	state := tournamentState(now, tournament)

	bf.WriteUint32(uint32(tournament.StartTime))
	bf.WriteUint32(uint32(tournament.EntryEnd))
	bf.WriteUint32(uint32(tournament.RankingEnd))
	bf.WriteUint32(uint32(tournament.RewardEnd))
	bf.WriteUint32(uint32(now))
	bf.WriteUint8(state)
	ps.Uint8(bf, tournament.Name, true)

	subEvents, err := s.server.tournamentRepo.GetSubEvents()
	if err != nil {
		s.logger.Error("Failed to get tournament sub-events", zap.Error(err))
		subEvents = nil
	}
	bf.WriteUint16(uint16(len(subEvents)))
	for _, se := range subEvents {
		bf.WriteUint32(se.ID)
		bf.WriteUint16(uint16(se.CupGroup))
		bf.WriteInt16(se.EventSubType)
		bf.WriteUint32(se.QuestFileID)
		ps.Uint8(bf, se.Name, true)
	}

	cups, err := s.server.tournamentRepo.GetCups(tournament.ID)
	if err != nil {
		s.logger.Error("Failed to get tournament cups", zap.Error(err))
		cups = nil
	}
	bf.WriteUint8(uint8(len(cups)))
	for _, cup := range cups {
		bf.WriteUint32(cup.ID)
		bf.WriteUint16(uint16(cup.CupGroup))
		bf.WriteUint16(uint16(cup.CupType))
		bf.WriteUint16(uint16(cup.Unk))
		ps.Uint8(bf, cup.Name, true)
		ps.Uint16(bf, cup.Description, true)
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnumerateOrder(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateOrder)
	bf := byteframe.NewByteFrame()

	now := uint32(TimeAdjusted().Unix())
	bf.WriteUint32(pkt.EventID)
	bf.WriteUint32(now)

	entries, err := s.server.tournamentRepo.GetLeaderboard(pkt.EventID)
	if err != nil {
		s.logger.Error("Failed to get tournament leaderboard", zap.Error(err), zap.Uint32("eventID", pkt.EventID))
		entries = nil
	}

	bf.WriteUint16(uint16(len(entries)))
	bf.WriteUint16(0) // unk

	for _, e := range entries {
		bf.WriteUint32(e.CharID)
		bf.WriteUint32(e.Rank)
		bf.WriteUint16(e.Grade)
		bf.WriteUint16(0) // pad
		bf.WriteUint16(e.HR)
		if s.server.erupeConfig.RealClientMode >= cfg.G10 {
			bf.WriteUint16(e.GR)
		}
		bf.WriteUint16(0) // pad
		charNameBytes := []byte(e.CharName)
		guildNameBytes := []byte(e.GuildName)
		bf.WriteUint8(uint8(len(charNameBytes) + 1))
		bf.WriteUint8(uint8(len(guildNameBytes) + 1))
		bf.WriteNullTerminatedBytes(charNameBytes)
		bf.WriteNullTerminatedBytes(guildNameBytes)
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfInfoTournament(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoTournament)
	bf := byteframe.NewByteFrame()

	now := TimeAdjusted().Unix()

	switch pkt.QueryType {
	case 0:
		tournament, err := s.server.tournamentRepo.GetActive(now)
		if err != nil {
			s.logger.Error("Failed to get active tournament for InfoTournament type 0", zap.Error(err))
		}
		bf.WriteUint32(0) // unk header
		if tournament == nil {
			bf.WriteUint32(0) // count = 0
			break
		}
		bf.WriteUint32(1) // count
		bf.WriteUint32(tournament.ID)
		bf.WriteUint32(0) // MaxPlayers
		bf.WriteUint32(0) // CurrentPlayers
		bf.WriteUint16(0) // Unk1
		bf.WriteUint16(0) // TextColor
		bf.WriteUint32(0) // Unk2
		bf.WriteUint32(uint32(tournament.StartTime))
		bf.WriteUint32(uint32(tournament.EntryEnd))
		bf.WriteUint32(uint32(tournament.RankingEnd))
		bf.WriteUint32(uint32(tournament.RewardEnd))
		bf.WriteUint32(uint32(tournament.RewardEnd))
		bf.WriteUint32(uint32(tournament.RewardEnd))
		bf.WriteUint8(0)  // Unk3
		bf.WriteUint8(0)  // Unk4
		bf.WriteUint32(0) // MinHR
		bf.WriteUint32(0) // MaxHR
		ps.Uint8(bf, tournament.Name, true)
		ps.Uint16(bf, "", false)
	case 1:
		// Return player registration status.
		bf.WriteUint32(uint32(now))
		tournament, err := s.server.tournamentRepo.GetActive(now)
		if err != nil {
			s.logger.Error("Failed to get active tournament for InfoTournament type 1", zap.Error(err))
		}
		if tournament == nil {
			bf.WriteUint32(0) // tournamentID
			bf.WriteUint32(0) // entryID
			bf.WriteUint32(0)
			bf.WriteUint8(0) // not registered
			bf.WriteUint32(0)
			ps.Uint8(bf, "", true)
			break
		}
		entry, err := s.server.tournamentRepo.GetEntry(s.charID, tournament.ID)
		if err != nil {
			s.logger.Error("Failed to get tournament entry for InfoTournament type 1", zap.Error(err))
		}
		bf.WriteUint32(tournament.ID)
		if entry != nil {
			bf.WriteUint32(entry.ID)
			bf.WriteUint32(0)
			bf.WriteUint8(1) // registered
		} else {
			bf.WriteUint32(0)
			bf.WriteUint32(0)
			bf.WriteUint8(0) // not registered
		}
		bf.WriteUint32(0)
		ps.Uint8(bf, tournament.Name, true)
	case 2:
		// Return empty lists (reward structures unknown).
		bf.WriteUint32(0)
		bf.WriteUint32(0) // count type 21
		bf.WriteUint32(0) // count type 22
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEntryTournament(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryTournament)
	now := TimeAdjusted().Unix()

	tournament, err := s.server.tournamentRepo.GetActive(now)
	if err != nil {
		s.logger.Error("Failed to get active tournament for EntryTournament", zap.Error(err))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	if tournament == nil {
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	entryID, err := s.server.tournamentRepo.Register(s.charID, tournament.ID)
	if err != nil {
		s.logger.Error("Failed to register for tournament", zap.Error(err),
			zap.Uint32("charID", s.charID), zap.Uint32("tournamentID", tournament.ID))
		doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	bf := byteframe.NewByteFrame()
	bf.WriteUint32(entryID)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnterTournamentQuest(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnterTournamentQuest)
	s.logger.Debug("EnterTournamentQuest",
		zap.Uint32("tournamentID", pkt.TournamentID),
		zap.Uint32("entryHandle", pkt.EntryHandle),
		zap.Uint32("unk2", pkt.Unk2),
		zap.Uint32("questSlot", pkt.QuestSlot),
		zap.Uint32("stageHandle", pkt.StageHandle),
	)
	if err := s.server.tournamentRepo.SubmitResult(
		s.charID,
		pkt.TournamentID,
		pkt.Unk2,
		pkt.QuestSlot,
		pkt.StageHandle,
	); err != nil {
		s.logger.Error("Failed to submit tournament result", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireTournament(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireTournament)
	// Reward item IDs are unknown. Return an empty reward list.
	rewards := []TournamentReward{}
	bf := byteframe.NewByteFrame()
	bf.WriteUint8(uint8(len(rewards)))
	for _, reward := range rewards {
		bf.WriteUint16(reward.Unk0)
		bf.WriteUint16(reward.Unk1)
		bf.WriteUint16(reward.Unk2)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}
