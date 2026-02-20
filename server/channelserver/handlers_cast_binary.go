package channelserver

import (
	"erupe-ce/common/byteframe"
	"erupe-ce/common/token"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"math"
	"strings"

	"go.uber.org/zap"
)

// MSG_SYS_CAST[ED]_BINARY types enum
const (
	BinaryMessageTypeState      = 0
	BinaryMessageTypeChat       = 1
	BinaryMessageTypeQuest      = 2
	BinaryMessageTypeData       = 3
	BinaryMessageTypeMailNotify = 4
	BinaryMessageTypeEmote      = 6
)

// MSG_SYS_CAST[ED]_BINARY broadcast types enum
const (
	BroadcastTypeTargeted = 0x01
	BroadcastTypeStage    = 0x03
	BroadcastTypeServer   = 0x06
	BroadcastTypeWorld    = 0x0a
)

func handleMsgSysCastBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCastBinary)
	tmp := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)

	if pkt.BroadcastType == BroadcastTypeStage && pkt.MessageType == BinaryMessageTypeData && len(pkt.RawDataPayload) == 0x10 {
		if tmp.ReadUint16() == 0x0002 && tmp.ReadUint8() == 0x18 {
			var timer bool
			if err := s.server.db.QueryRow(`SELECT COALESCE(timer, false) FROM users WHERE id=$1`, s.userID).Scan(&timer); err != nil {
				s.logger.Error("Failed to get timer setting", zap.Error(err))
			}
			if timer {
				_ = tmp.ReadBytes(9)
				tmp.SetLE()
				frame := tmp.ReadUint32()
				sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.timer, frame/30/60/60, frame/30/60, frame/30%60, int(math.Round(float64(frame%30*100)/3)), frame))
			}
		}
	}

	if s.server.erupeConfig.DebugOptions.QuestTools {
		if pkt.BroadcastType == BroadcastTypeStage && pkt.MessageType == BinaryMessageTypeQuest && len(pkt.RawDataPayload) > 32 {
			// This is only correct most of the time
			tmp.ReadBytes(20)
			tmp.SetLE()
			x := tmp.ReadFloat32()
			y := tmp.ReadFloat32()
			z := tmp.ReadFloat32()
			s.logger.Debug("Coord", zap.Float32s("XYZ", []float32{x, y, z}))
		}
	}

	// Parse out the real casted binary payload
	var msgBinTargeted *binpacket.MsgBinTargeted
	var message, author string
	var returnToSender bool
	if pkt.MessageType == BinaryMessageTypeChat {
		tmp.SetLE()
		_, _ = tmp.Seek(8, 0)
		message = string(tmp.ReadNullTerminatedBytes())
		author = string(tmp.ReadNullTerminatedBytes())
	}

	// Customise payload
	realPayload := pkt.RawDataPayload
	if pkt.BroadcastType == BroadcastTypeTargeted {
		tmp.SetBE()
		_, _ = tmp.Seek(0, 0)
		msgBinTargeted = &binpacket.MsgBinTargeted{}
		err := msgBinTargeted.Parse(tmp)
		if err != nil {
			s.logger.Warn("Failed to parse targeted cast binary")
			return
		}
		realPayload = msgBinTargeted.RawDataPayload
	} else if pkt.MessageType == BinaryMessageTypeChat {
		if message == "@dice" {
			returnToSender = true
			m := binpacket.MsgBinChat{
				Type:       BinaryMessageTypeChat,
				Flags:      4,
				Message:    fmt.Sprintf(`%d`, token.RNG.Intn(100)+1),
				SenderName: author,
			}
			bf := byteframe.NewByteFrame()
			bf.SetLE()
			_ = m.Build(bf)
			realPayload = bf.Data()
		} else {
			bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
			bf.SetLE()
			chatMessage := &binpacket.MsgBinChat{}
			_ = chatMessage.Parse(bf)
			if strings.HasPrefix(chatMessage.Message, s.server.erupeConfig.CommandPrefix) {
				parseChatCommand(s, chatMessage.Message)
				return
			}
			if (pkt.BroadcastType == BroadcastTypeStage && s.stage.id == "sl1Ns200p0a0u0") || pkt.BroadcastType == BroadcastTypeWorld {
				s.server.DiscordChannelSend(chatMessage.SenderName, chatMessage.Message)
			}
		}
	}

	// Make the response to forward to the other client(s).
	resp := &mhfpacket.MsgSysCastedBinary{
		CharID:         s.charID,
		BroadcastType:  pkt.BroadcastType, // (The client never uses Type0 upon receiving)
		MessageType:    pkt.MessageType,
		RawDataPayload: realPayload,
	}

	// Send to the proper recipients.
	switch pkt.BroadcastType {
	case BroadcastTypeWorld:
		s.server.WorldcastMHF(resp, s, nil)
	case BroadcastTypeStage:
		if returnToSender {
			s.stage.BroadcastMHF(resp, nil)
		} else {
			s.stage.BroadcastMHF(resp, s)
		}
	case BroadcastTypeServer:
		if pkt.MessageType == BinaryMessageTypeChat {
			raviSema := s.server.getRaviSemaphore()
			if raviSema != nil {
				raviSema.BroadcastMHF(resp, s)
			}
		} else {
			s.server.BroadcastMHF(resp, s)
		}
	case BroadcastTypeTargeted:
		for _, targetID := range (*msgBinTargeted).TargetCharIDs {
			char := s.server.FindSessionByCharID(targetID)

			if char != nil {
				char.QueueSendMHFNonBlocking(resp)
			}
		}
	default:
		s.Lock()
		haveStage := s.stage != nil
		if haveStage {
			s.stage.BroadcastMHF(resp, s)
		}
		s.Unlock()
	}
}

func handleMsgSysCastedBinary(s *Session, p mhfpacket.MHFPacket) {}
