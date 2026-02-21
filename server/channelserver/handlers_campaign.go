package channelserver

import (
	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/common/stringsupport"
	cfg "erupe-ce/config"
	"erupe-ce/network/mhfpacket"
	"time"
)

// CampaignEvent represents a promotional campaign event.
type CampaignEvent struct {
	ID         uint32
	Unk0       uint32
	MinHR      int16
	MaxHR      int16
	MinSR      int16
	MaxSR      int16
	MinGR      int16
	MaxGR      int16
	Unk1       uint16
	Unk2       uint8
	Unk3       uint8
	Unk4       uint16
	Unk5       uint16
	Start      time.Time
	End        time.Time
	Unk6       uint8
	String0    string
	String1    string
	String2    string
	String3    string
	Link       string
	Prefix     string
	Categories []uint16
}

// CampaignCategory represents a category grouping for campaign events.
type CampaignCategory struct {
	ID          uint16
	Type        uint8
	Title       string
	Description string
}

// CampaignLink links a campaign event to its items/rewards.
type CampaignLink struct {
	CategoryID uint16
	CampaignID uint32
}

func handleMsgMhfEnumerateCampaign(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateCampaign)
	bf := byteframe.NewByteFrame()

	events := []CampaignEvent{}
	categories := []CampaignCategory{}
	var campaignLinks []CampaignLink

	if len(events) > 255 {
		bf.WriteUint8(255)
		bf.WriteUint16(uint16(len(events)))
	} else {
		bf.WriteUint8(uint8(len(events)))
	}
	for _, event := range events {
		bf.WriteUint32(event.ID)
		bf.WriteUint32(event.Unk0)
		bf.WriteInt16(event.MinHR)
		bf.WriteInt16(event.MaxHR)
		bf.WriteInt16(event.MinSR)
		bf.WriteInt16(event.MaxSR)
		if s.server.erupeConfig.RealClientMode >= cfg.G3 {
			bf.WriteInt16(event.MinGR)
			bf.WriteInt16(event.MaxGR)
		}
		bf.WriteUint16(event.Unk1)
		bf.WriteUint8(event.Unk2)
		bf.WriteUint8(event.Unk3)
		bf.WriteUint16(event.Unk4)
		bf.WriteUint16(event.Unk5)
		bf.WriteUint32(uint32(event.Start.Unix()))
		bf.WriteUint32(uint32(event.End.Unix()))
		bf.WriteUint8(event.Unk6)
		ps.Uint8(bf, event.String0, true)
		ps.Uint8(bf, event.String1, true)
		ps.Uint8(bf, event.String2, true)
		ps.Uint8(bf, event.String3, true)
		ps.Uint8(bf, event.Link, true)
		for i := range event.Categories {
			campaignLinks = append(campaignLinks, CampaignLink{event.Categories[i], event.ID})
		}
	}

	if len(events) > 255 {
		bf.WriteUint8(255)
		bf.WriteUint16(uint16(len(events)))
	} else {
		bf.WriteUint8(uint8(len(events)))
	}
	for _, event := range events {
		bf.WriteUint32(event.ID)
		bf.WriteUint8(1) // Always 1?
		bf.WriteBytes([]byte(event.Prefix))
	}

	if len(categories) > 255 {
		bf.WriteUint8(255)
		bf.WriteUint16(uint16(len(categories)))
	} else {
		bf.WriteUint8(uint8(len(categories)))
	}
	for _, category := range categories {
		bf.WriteUint16(category.ID)
		bf.WriteUint8(category.Type)
		xTitle := stringsupport.UTF8ToSJIS(category.Title)
		xDescription := stringsupport.UTF8ToSJIS(category.Description)
		bf.WriteUint8(uint8(len(xTitle)))
		bf.WriteUint8(uint8(len(xDescription)))
		bf.WriteBytes(xTitle)
		bf.WriteBytes(xDescription)
	}

	if len(campaignLinks) > 255 {
		bf.WriteUint8(255)
		bf.WriteUint16(uint16(len(campaignLinks)))
	} else {
		bf.WriteUint8(uint8(len(campaignLinks)))
	}
	for _, link := range campaignLinks {
		bf.WriteUint16(link.CategoryID)
		bf.WriteUint32(link.CampaignID)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfStateCampaign(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfStateCampaign)
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(1)
	bf.WriteUint16(0)
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfApplyCampaign(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfApplyCampaign)
	bf := byteframe.NewByteFrame()
	bf.WriteUint32(1)
	doAckSimpleSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfEnumerateItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfEnumerateItem)
	items := []struct {
		Unk0 uint32
		Unk1 uint16
		Unk2 uint16
		Unk3 uint16
		Unk4 uint32
		Unk5 uint32
	}{}
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(len(items)))
	for _, item := range items {
		bf.WriteUint32(item.Unk0)
		bf.WriteUint16(item.Unk1)
		bf.WriteUint16(item.Unk2)
		bf.WriteUint16(item.Unk3)
		bf.WriteUint32(item.Unk4)
		bf.WriteUint32(item.Unk5)
	}
	doAckBufSucceed(s, pkt.AckHandle, bf.Data())
}

func handleMsgMhfAcquireItem(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgMhfAcquireItem)
	doAckSimpleSucceed(s, pkt.AckHandle, make([]byte, 4))
}
