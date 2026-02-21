package channelserver

import (
	"database/sql"
	"errors"
	"sort"
	"time"

	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/token"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"

	"go.uber.org/zap"
)

func handleMsgMhfSaveMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfSaveMezfesData)
	saveCharacterData(s, pkt.AckHandle, "mezfes", pkt.RawDataPayload, 4096)
}

func handleMsgMhfLoadMezfesData(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfLoadMezfesData)
	loadCharacterData(s, pkt.AckHandle, "mezfes",
		[]byte{0x00, 0x00, 0x00, 0x00, 0x02, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

func handleMsgMhfEnumerateRanking(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateRanking)
	bf := byteframe.NewByteFrame()
	state := s.server.erupeConfig.DebugOptions.TournamentOverride
	// Unk
	// Unk
	// Start?
	// End?
	midnight := TimeMidnight()
	switch state {
	case 1:
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(3 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(13 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(20 * 24 * time.Hour).Unix()))
	case 2:
		bf.WriteUint32(uint32(midnight.Add(-3 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(10 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(17 * 24 * time.Hour).Unix()))
	case 3:
		bf.WriteUint32(uint32(midnight.Add(-13 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Add(-10 * 24 * time.Hour).Unix()))
		bf.WriteUint32(uint32(midnight.Unix()))
		bf.WriteUint32(uint32(midnight.Add(7 * 24 * time.Hour).Unix()))
	default:
		bf.WriteBytes(make([]byte, 16))
		bf.WriteUint32(uint32(TimeAdjusted().Unix())) // TS Current Time
		bf.WriteUint8(3)
		bf.WriteBytes(make([]byte, 4))
		doAckBufSucceed(s, pkt.AckHandle, bf.Data())
		return
	}
	bf.WriteUint32(uint32(TimeAdjusted().Unix())) // TS Current Time
	bf.WriteUint8(3)
	ps.Uint8(bf, "", false)
	bf.WriteUint16(0) // numEvents
	bf.WriteUint8(0)  // numCups

	/*
		struct event
		uint32 eventID
		uint16 unk
		uint16 unk
		uint32 unk
		psUint8 name

		struct cup
		uint32 cupID
		uint16 unk
		uint16 unk
		uint16 unk
		psUint8 name
		psUint16 desc
	*/

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func cleanupFesta(s *Session) {
	if err := s.server.festaRepo.CleanupAll(); err != nil {
		s.logger.Error("Failed to cleanup festa", zap.Error(err))
	}
}

// Festa timing constants (all values in seconds)
const (
	festaVotingDuration = 9000    // 150 min voting window
	festaRewardDuration = 1240200 // ~14.35 days reward period
	festaEventLifespan  = 2977200 // ~34.5 days total event window
)

func generateFestaTimestamps(s *Session, start uint32, debug bool) []uint32 {
	timestamps := make([]uint32, 5)
	midnight := TimeMidnight()
	if debug && start <= 3 {
		midnight := uint32(midnight.Unix())
		switch start {
		case 1:
			timestamps[0] = midnight
			timestamps[1] = timestamps[0] + secsPerWeek
			timestamps[2] = timestamps[1] + secsPerWeek
			timestamps[3] = timestamps[2] + festaVotingDuration
			timestamps[4] = timestamps[3] + festaRewardDuration
		case 2:
			timestamps[0] = midnight - secsPerWeek
			timestamps[1] = midnight
			timestamps[2] = timestamps[1] + secsPerWeek
			timestamps[3] = timestamps[2] + festaVotingDuration
			timestamps[4] = timestamps[3] + festaRewardDuration
		case 3:
			timestamps[0] = midnight - 2*secsPerWeek
			timestamps[1] = midnight - secsPerWeek
			timestamps[2] = midnight
			timestamps[3] = timestamps[2] + festaVotingDuration
			timestamps[4] = timestamps[3] + festaRewardDuration
		}
		return timestamps
	}
	if start == 0 || TimeAdjusted().Unix() > int64(start)+festaEventLifespan {
		cleanupFesta(s)
		// Generate a new festa, starting midnight tomorrow
		start = uint32(midnight.Add(24 * time.Hour).Unix())
		if err := s.server.festaRepo.InsertEvent(start); err != nil {
			s.logger.Error("Failed to insert festa event", zap.Error(err))
		}
	}
	timestamps[0] = start
	timestamps[1] = timestamps[0] + secsPerWeek
	timestamps[2] = timestamps[1] + secsPerWeek
	timestamps[3] = timestamps[2] + festaVotingDuration
	timestamps[4] = timestamps[3] + festaRewardDuration
	return timestamps
}

// FestaTrial represents a festa trial/challenge entry.
type FestaTrial struct {
	ID        uint32        `db:"id"`
	Objective uint16        `db:"objective"`
	GoalID    uint32        `db:"goal_id"`
	TimesReq  uint16        `db:"times_req"`
	Locale    uint16        `db:"locale_req"`
	Reward    uint16        `db:"reward"`
	Monopoly  FestivalColor `db:"monopoly"`
	Unk       uint16
}

// FestaReward represents a festa reward entry.
type FestaReward struct {
	Unk0     uint8
	Unk1     uint8
	ItemType uint16
	Quantity uint16
	ItemID   uint16
	MinHR    uint16 // Minimum Hunter Rank to receive this reward
	MinSR    uint16 // Minimum Skill Rank (max across weapon types) to receive this reward
	MinGR    uint8  // Minimum G Rank to receive this reward
}

func handleMsgMhfInfoFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoFesta)
	bf := byteframe.NewByteFrame()

	const festaIDSentinel = uint32(0xDEADBEEF)
	id, start := festaIDSentinel, uint32(0)
	events, err := s.server.festaRepo.GetFestaEvents()
	if err != nil {
		s.logger.Error("Failed to query festa schedule", zap.Error(err))
	} else {
		for _, e := range events {
			id = e.ID
			start = e.StartTime
		}
	}

	var timestamps []uint32
	if s.server.erupeConfig.DebugOptions.FestaOverride >= 0 {
		if s.server.erupeConfig.DebugOptions.FestaOverride == 0 {
			doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
			return
		}
		timestamps = generateFestaTimestamps(s, uint32(s.server.erupeConfig.DebugOptions.FestaOverride), true)
	} else {
		timestamps = generateFestaTimestamps(s, start, false)
	}

	if timestamps[0] > uint32(TimeAdjusted().Unix()) {
		doAckBufSucceed(s, pkt.AckHandle, make([]byte, 4))
		return
	}

	blueSouls, err := s.server.festaRepo.GetTeamSouls("blue")
	if err != nil {
		s.logger.Error("Failed to get blue souls", zap.Error(err))
	}
	redSouls, err := s.server.festaRepo.GetTeamSouls("red")
	if err != nil {
		s.logger.Error("Failed to get red souls", zap.Error(err))
	}

	bf.WriteUint32(id)
	for _, timestamp := range timestamps {
		bf.WriteUint32(timestamp)
	}
	bf.WriteUint32(uint32(TimeAdjusted().Unix()))
	bf.WriteUint8(4)
	ps.Uint8(bf, "", false)
	bf.WriteUint32(0)
	bf.WriteUint32(blueSouls)
	bf.WriteUint32(redSouls)

	trials, err := s.server.festaRepo.GetTrialsWithMonopoly()
	if err != nil {
		s.logger.Error("Failed to query festa trials", zap.Error(err))
	}
	bf.WriteUint16(uint16(len(trials)))
	for _, trial := range trials {
		bf.WriteUint32(trial.ID)
		bf.WriteUint16(trial.Objective)
		bf.WriteUint32(trial.GoalID)
		bf.WriteUint16(trial.TimesReq)
		bf.WriteUint16(trial.Locale)
		bf.WriteUint16(trial.Reward)
		bf.WriteInt16(FestivalColorCodes[trial.Monopoly])
		if s.server.erupeConfig.RealClientMode >= cfg.F4 { // Not in S6.0
			bf.WriteUint16(trial.Unk)
		}
	}

	// The Winner and Loser Armor IDs are missing
	// Item 7011 may not exist in older versions, remove to prevent crashes
	// Fields: {Unk0, Unk1, ItemType, Quantity, ItemID, MinHR, MinSR, MinGR}
	rewards := []FestaReward{
		{1, 0, 7, 350, 1520, 0, 0, 0},
		{1, 0, 7, 1000, 7011, 0, 0, 1},
		{1, 0, 12, 1000, 0, 0, 0, 0},
		{1, 0, 13, 0, 0, 0, 0, 0},
		//{1, 0, 1, 0, 0, 0, 0, 0},
		{2, 0, 7, 350, 1520, 0, 0, 0},
		{2, 0, 7, 1000, 7011, 0, 0, 1},
		{2, 0, 12, 1000, 0, 0, 0, 0},
		{2, 0, 13, 0, 0, 0, 0, 0},
		//{2, 0, 4, 0, 0, 0, 0, 0},
		{3, 0, 7, 350, 1520, 0, 0, 0},
		{3, 0, 7, 1000, 7011, 0, 0, 1},
		{3, 0, 12, 1000, 0, 0, 0, 0},
		{3, 0, 13, 0, 0, 0, 0, 0},
		//{3, 0, 1, 0, 0, 0, 0, 0},
		{4, 0, 7, 350, 1520, 0, 0, 0},
		{4, 0, 7, 1000, 7011, 0, 0, 1},
		{4, 0, 12, 1000, 0, 0, 0, 0},
		{4, 0, 13, 0, 0, 0, 0, 0},
		//{4, 0, 4, 0, 0, 0, 0, 0},
		{5, 0, 7, 350, 1520, 0, 0, 0},
		{5, 0, 7, 1000, 7011, 0, 0, 1},
		{5, 0, 12, 1000, 0, 0, 0, 0},
		{5, 0, 13, 0, 0, 0, 0, 0},
		//{5, 0, 1, 0, 0, 0, 0, 0},
	}

	bf.WriteUint16(uint16(len(rewards)))
	for _, reward := range rewards {
		bf.WriteUint8(reward.Unk0)
		bf.WriteUint8(reward.Unk1)
		bf.WriteUint16(reward.ItemType)
		bf.WriteUint16(reward.Quantity)
		bf.WriteUint16(reward.ItemID)
		// Confirmed present in G3 via Wii U disassembly of import_festa_info
		if s.server.erupeConfig.RealClientMode >= cfg.G3 {
			bf.WriteUint16(reward.MinHR)
			bf.WriteUint16(reward.MinSR)
			bf.WriteUint8(reward.MinGR)
		}
	}
	if s.server.erupeConfig.RealClientMode <= cfg.G61 {
		if s.server.erupeConfig.GameplayOptions.MaximumFP > 0xFFFF {
			s.server.erupeConfig.GameplayOptions.MaximumFP = 0xFFFF
		}
		bf.WriteUint16(uint16(s.server.erupeConfig.GameplayOptions.MaximumFP))
	} else {
		bf.WriteUint32(s.server.erupeConfig.GameplayOptions.MaximumFP)
	}
	bf.WriteUint16(100) // Reward multiplier (%)

	bf.WriteUint16(4)
	for i := uint16(0); i < 4; i++ {
		ranking, err := s.server.festaRepo.GetTopGuildForTrial(i + 1)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("Failed to get festa trial ranking", zap.Error(err))
		}
		bf.WriteUint32(ranking.GuildID)
		bf.WriteUint16(i + 1)
		bf.WriteInt16(FestivalColorCodes[ranking.Team])
		ps.Uint8(bf, ranking.GuildName, true)
	}
	bf.WriteUint16(7)
	for i := uint16(0); i < 7; i++ {
		offset := secsPerDay * uint32(i)
		ranking, err := s.server.festaRepo.GetTopGuildInWindow(timestamps[1]+offset, timestamps[1]+offset+secsPerDay)
		if err != nil && !errors.Is(err, sql.ErrNoRows) {
			s.logger.Error("Failed to get festa daily ranking", zap.Error(err))
		}
		bf.WriteUint32(ranking.GuildID)
		bf.WriteUint16(i + 1)
		bf.WriteInt16(FestivalColorCodes[ranking.Team])
		ps.Uint8(bf, ranking.GuildName, true)
	}

	bf.WriteUint32(0) // Clan goal
	// Final bonus rates
	bf.WriteUint32(5000) // 5000+ souls
	bf.WriteUint32(2000) // 2000-4999 souls
	bf.WriteUint32(1000) // 1000-1999 souls
	bf.WriteUint32(100)  // 100-999 souls
	bf.WriteUint16(300)  // 300% bonus
	bf.WriteUint16(200)  // 200% bonus
	bf.WriteUint16(150)  // 150% bonus
	bf.WriteUint16(100)  // Normal rate
	bf.WriteUint16(50)   // 50% penalty

	if s.server.erupeConfig.RealClientMode >= cfg.G52 {
		ps.Uint16(bf, "", false)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// state festa (U)ser
func handleMsgMhfStateFestaU(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateFestaU)
	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	applicant := false
	if guild != nil {
		var appErr error
		applicant, appErr = s.server.guildRepo.HasApplication(guild.ID, s.charID)
		if appErr != nil {
			s.logger.Warn("Failed to check guild application status", zap.Error(appErr))
		}
	}
	if err != nil || guild == nil || applicant {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	souls, err := s.server.festaRepo.GetCharSouls(s.charID)
	if err != nil {
		s.logger.Error("Failed to get festa user souls", zap.Error(err))
	}
	claimed := s.server.festaRepo.HasClaimedMainPrize(s.charID)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(souls)
	if !claimed {
		bf.WriteBool(true)
		bf.WriteBool(false)
	} else {
		bf.WriteBool(false)
		bf.WriteBool(true)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

// state festa (G)uild
func handleMsgMhfStateFestaG(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateFestaG)
	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	applicant := false
	if guild != nil {
		var appErr error
		applicant, appErr = s.server.guildRepo.HasApplication(guild.ID, s.charID)
		if appErr != nil {
			s.logger.Warn("Failed to check guild application status", zap.Error(appErr))
		}
	}
	resp := byteframe.NewByteFrame()
	if err != nil || guild == nil || applicant {
		resp.WriteUint32(0)
		resp.WriteInt32(0)
		resp.WriteInt32(-1)
		resp.WriteInt32(0)
		resp.WriteInt32(0)
		doAckBufSucceed(s, pkt.AckHandle, resp.Data())
		return
	}
	resp.WriteUint32(guild.Souls)
	resp.WriteInt32(1) // unk
	resp.WriteInt32(1) // unk, rank?
	resp.WriteInt32(1) // unk
	resp.WriteInt32(1) // unk
	doAckBufSucceed(s, pkt.AckHandle, resp.Data())
}

func handleMsgMhfEnumerateFestaMember(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateFestaMember)
	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	members, err := s.server.guildRepo.GetMembers(guild.ID, false)
	if err != nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	sort.Slice(members, func(i, j int) bool {
		return members[i].Souls > members[j].Souls
	})
	var validMembers []*GuildMember
	for _, member := range members {
		if member.Souls > 0 {
			validMembers = append(validMembers, member)
		}
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(validMembers)))
	bf.WriteUint16(0) // Unk
	for _, member := range validMembers {
		bf.WriteUint32(member.CharID)
		if s.server.erupeConfig.RealClientMode <= cfg.Z1 {
			bf.WriteUint16(uint16(member.Souls))
			bf.WriteUint16(0)
		} else {
			bf.WriteUint32(member.Souls)
		}
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfVoteFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfVoteFesta)
	if err := s.server.festaRepo.VoteTrial(s.charID, pkt.TrialID); err != nil {
		s.logger.Error("Failed to update festa trial vote", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfEntryFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryFesta)
	guild, err := s.server.guildRepo.GetByCharID(s.charID)
	if err != nil || guild == nil {
		doAckSimpleFail(s, pkt.AckHandle, make([]byte, 4))
		return
	}
	team := uint32(token.RNG.Intn(2))
	teamName := "blue"
	if team == 1 {
		teamName = "red"
	}
	if err := s.server.festaRepo.RegisterGuild(guild.ID, teamName); err != nil {
		s.logger.Error("Failed to register guild for festa", zap.Error(err))
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(team)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfChargeFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfChargeFesta)
	if err := s.server.festaRepo.SubmitSouls(s.charID, pkt.GuildID, pkt.Souls); err != nil {
		s.logger.Error("Failed to submit festa souls", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireFesta(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireFesta)
	if err := s.server.festaRepo.ClaimPrize(0, s.charID); err != nil {
		s.logger.Error("Failed to accept festa prize", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireFestaPersonalPrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireFestaPersonalPrize)
	if err := s.server.festaRepo.ClaimPrize(pkt.PrizeID, s.charID); err != nil {
		s.logger.Error("Failed to accept festa personal prize", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

func handleMsgMhfAcquireFestaIntermediatePrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireFestaIntermediatePrize)
	if err := s.server.festaRepo.ClaimPrize(pkt.PrizeID, s.charID); err != nil {
		s.logger.Error("Failed to accept festa intermediate prize", zap.Error(err))
	}
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

// Prize represents a festa prize entry.
type Prize struct {
	ID       uint32 `db:"id"`
	Tier     uint32 `db:"tier"`
	SoulsReq uint32 `db:"souls_req"`
	ItemID   uint32 `db:"item_id"`
	NumItem  uint32 `db:"num_item"`
	Claimed  int    `db:"claimed"`
}

func writePrizeList(s *Session, pkt mhfpacket.MHFPacket, ackHandle uint32, prizeType string) {
	prizes, err := s.server.festaRepo.ListPrizes(s.charID, prizeType)
	var count uint32
	prizeData := byteframe.NewByteFrame()
	if err != nil {
		s.logger.Error("Failed to query festa prizes", zap.Error(err), zap.String("type", prizeType))
	} else {
		for _, prize := range prizes {
			count++
			prizeData.WriteUint32(prize.ID)
			prizeData.WriteUint32(prize.Tier)
			prizeData.WriteUint32(prize.SoulsReq)
			prizeData.WriteUint32(7) // Unk
			prizeData.WriteUint32(prize.ItemID)
			prizeData.WriteUint32(prize.NumItem)
			prizeData.WriteBool(prize.Claimed > 0)
		}
	}
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(count)
	bf.WriteBytes(prizeData.Data())
	doAckBufSucceed(s, ackHandle, bf.Data())
}

func handleMsgMhfEnumerateFestaPersonalPrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateFestaPersonalPrize)
	writePrizeList(s, p, pkt.AckHandle, "personal")
}

func handleMsgMhfEnumerateFestaIntermediatePrize(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateFestaIntermediatePrize)
	writePrizeList(s, p, pkt.AckHandle, "guild")
}
