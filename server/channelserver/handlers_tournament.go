package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/network/mhfpacket"
	"time"
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

func handleMsgMhfInfoTournament(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfInfoTournament)
	bf := byteframe.NewByteFrame()

	tournamentInfo0 := []TournamentInfo0{}
	tournamentInfo21 := []TournamentInfo21{}
	tournamentInfo22 := []TournamentInfo22{}

	switch pkt.QueryType {
	case 0:
		bf.WriteUint32(0)
		bf.WriteUint32(uint32(len(tournamentInfo0)))
		for _, tinfo := range tournamentInfo0 {
			bf.WriteUint32(tinfo.ID)
			bf.WriteUint32(tinfo.MaxPlayers)
			bf.WriteUint32(tinfo.CurrentPlayers)
			bf.WriteUint16(tinfo.Unk1)
			bf.WriteUint16(tinfo.TextColor)
			bf.WriteUint32(tinfo.Unk2)
			bf.WriteUint32(uint32(tinfo.Time1.Unix()))
			bf.WriteUint32(uint32(tinfo.Time2.Unix()))
			bf.WriteUint32(uint32(tinfo.Time3.Unix()))
			bf.WriteUint32(uint32(tinfo.Time4.Unix()))
			bf.WriteUint32(uint32(tinfo.Time5.Unix()))
			bf.WriteUint32(uint32(tinfo.Time6.Unix()))
			bf.WriteUint8(tinfo.Unk3)
			bf.WriteUint8(tinfo.Unk4)
			bf.WriteUint32(tinfo.MinHR)
			bf.WriteUint32(tinfo.MaxHR)
			ps.Uint8(bf, tinfo.Unk5, true)
			ps.Uint16(bf, tinfo.Unk6, true)
		}
	case 1:
		bf.WriteUint32(uint32(TimeAdjusted().Unix()))
		bf.WriteUint32(0) // Registered ID
		bf.WriteUint32(0)
		bf.WriteUint32(0)
		bf.WriteUint8(0)
		bf.WriteUint32(0)
		ps.Uint8(bf, "", true)
	case 2:
		bf.WriteUint32(0)
		bf.WriteUint32(uint32(len(tournamentInfo21)))
		for _, info := range tournamentInfo21 {
			bf.WriteUint32(info.Unk0)
			bf.WriteUint32(info.Unk1)
			bf.WriteUint32(info.Unk2)
			bf.WriteUint8(info.Unk3)
		}
		bf.WriteUint32(uint32(len(tournamentInfo22)))
		for _, info := range tournamentInfo22 {
			bf.WriteUint32(info.Unk0)
			bf.WriteUint32(info.Unk1)
			bf.WriteUint32(info.Unk2)
			bf.WriteUint8(info.Unk3)
			ps.Uint8(bf, info.Unk4, true)
		}
	}

	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEntryTournament(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEntryTournament)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}

// TournamentReward represents a tournament reward entry.
type TournamentReward struct {
	Unk0 uint16
	Unk1 uint16
	Unk2 uint16
}

func handleMsgMhfAcquireTournament(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireTournament)
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
