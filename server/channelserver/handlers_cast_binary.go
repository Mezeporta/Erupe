package channelserver

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
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

	const (
		timerPayloadSize = 0x10           // expected payload length for timer packets
		timerSubtype     = uint16(0x0002) // timer data subtype identifier
		timerFlag        = uint8(0x18)    // timer flag byte
	)
	if pkt.BroadcastType == BroadcastTypeStage && pkt.MessageType == BinaryMessageTypeData && len(pkt.RawDataPayload) == timerPayloadSize {
		if tmp.ReadUint16() == timerSubtype && tmp.ReadUint8() == timerFlag {
			timer, err := s.server.userRepo.GetTimer(s.userID)
			if err != nil {
				s.logger.Error("Failed to get timer setting", zap.Error(err))
			}
			if timer {
				_ = tmp.ReadBytes(9)
				tmp.SetLE()
				frame := tmp.ReadUint32()
				sendServerChatMessage(s, fmt.Sprintf(s.I18n().timer, frame/30/60/60, frame/30/60, frame/30%60, int(math.Round(float64(frame%30*100)/3)), frame))
			}
		}
	}

	if s.server.erupeConfig.DebugOptions.QuestTools {
		if pkt.BroadcastType == BroadcastTypeStage && pkt.MessageType == BinaryMessageTypeQuest && len(pkt.RawDataPayload) > 32 {
			// Temporary raw dump to find the real struct layout -- ground
			// truth for the decode below.
			s.logger.Debug("QuestBinaryRaw",
				zap.Uint32("charID", uint32(s.charID)),
				zap.Int("len", len(pkt.RawDataPayload)),
				zap.String("hex", hex.EncodeToString(pkt.RawDataPayload)),
			)
			// Party-slot records are tagged 0x0N020030 (LE) at their own
			// offset 0, N = party slot 1-4, with XYZ at record-relative
			// offset 8-19. Locate slot 1's own record explicitly instead of
			// blindly reading a fixed offset 20: a fixed read there aliases
			// a different entity's own XYZ (e.g. the 0x7c000108 marker's,
			// which starts at offset 12) whenever this payload isn't a
			// plain party-slot message, silently logging the wrong entity's
			// position as "Coord".
			slot1ID := []byte{0x30, 0x00, 0x02, 0x01}
			if idx := bytes.Index(pkt.RawDataPayload, slot1ID); idx != -1 && len(pkt.RawDataPayload) >= idx+20 {
				x := math.Float32frombits(binary.LittleEndian.Uint32(pkt.RawDataPayload[idx+8:]))
				y := math.Float32frombits(binary.LittleEndian.Uint32(pkt.RawDataPayload[idx+12:]))
				z := math.Float32frombits(binary.LittleEndian.Uint32(pkt.RawDataPayload[idx+16:]))
				s.logger.Debug("Coord", zap.Float32s("XYZ", []float32{x, y, z}))
			}
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

func handleMsgSysCastedBinary(s *Session, p mhfpacket.MHFPacket) {} // stub: unimplemented
