package channelserver

import (
	"encoding/hex"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcourse"
	"erupe-ce/common/token"
	"erupe-ce/config"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"golang.org/x/exp/slices"
	"math"
	"strings"
	"time"

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

var commands map[string]config.Command

func init() {
	commands = make(map[string]config.Command)

	// Skip initialization if config is not loaded (e.g., during tests)
	if config.ErupeConfig == nil {
		return
	}

	zapConfig := zap.NewDevelopmentConfig()
	zapConfig.DisableCaller = true
	zapLogger, _ := zapConfig.Build()
	defer zapLogger.Sync()
	logger := zapLogger.Named("commands")
	cmds := config.ErupeConfig.Commands
	for _, cmd := range cmds {
		commands[cmd.Name] = cmd
		if cmd.Enabled {
			logger.Info(fmt.Sprintf("Command %s: Enabled, prefix: %s", cmd.Name, cmd.Prefix))
		} else {
			logger.Info(fmt.Sprintf("Command %s: Disabled", cmd.Name))
		}
	}
}

func sendDisabledCommandMessage(s *Session, cmd config.Command) {
	sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandDisabled"], cmd.Name))
}

func sendServerChatMessage(s *Session, message string) {
	// Make the inside of the casted binary
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	msgBinChat := &binpacket.MsgBinChat{
		Unk0:       0,
		Type:       5,
		Flags:      0x80,
		Message:    message,
		SenderName: "Erupe",
	}
	msgBinChat.Build(bf)

	castedBin := &mhfpacket.MsgSysCastedBinary{
		CharID:         s.charID,
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}

	s.QueueSendMHF(castedBin)
}

func parseChatCommand(s *Session, command string) {
	if strings.HasPrefix(command, commands["Reload"].Prefix) {
		// Flush all objects and users and reload
		if commands["Reload"].Enabled {
			sendServerChatMessage(s, s.server.dict["commandReload"])
			var temp mhfpacket.MHFPacket
			deleteNotif := byteframe.NewByteFrame()
			for _, object := range s.stage.objects {
				if object.ownerCharID == s.charID {
					continue
				}
				temp = &mhfpacket.MsgSysDeleteObject{ObjID: object.id}
				deleteNotif.WriteUint16(uint16(temp.Opcode()))
				temp.Build(deleteNotif, s.clientContext)
			}
			for _, session := range s.server.sessions {
				if s == session {
					continue
				}
				temp = &mhfpacket.MsgSysDeleteUser{CharID: session.charID}
				deleteNotif.WriteUint16(uint16(temp.Opcode()))
				temp.Build(deleteNotif, s.clientContext)
			}
			deleteNotif.WriteUint16(0x0010)
			s.QueueSend(deleteNotif.Data())
			time.Sleep(500 * time.Millisecond)
			reloadNotif := byteframe.NewByteFrame()
			for _, session := range s.server.sessions {
				if s == session {
					continue
				}
				temp = &mhfpacket.MsgSysInsertUser{CharID: session.charID}
				reloadNotif.WriteUint16(uint16(temp.Opcode()))
				temp.Build(reloadNotif, s.clientContext)
				for i := 0; i < 3; i++ {
					temp = &mhfpacket.MsgSysNotifyUserBinary{
						CharID:     session.charID,
						BinaryType: uint8(i + 1),
					}
					reloadNotif.WriteUint16(uint16(temp.Opcode()))
					temp.Build(reloadNotif, s.clientContext)
				}
			}
			for _, obj := range s.stage.objects {
				if obj.ownerCharID == s.charID {
					continue
				}
				temp = &mhfpacket.MsgSysDuplicateObject{
					ObjID:       obj.id,
					X:           obj.x,
					Y:           obj.y,
					Z:           obj.z,
					Unk0:        0,
					OwnerCharID: obj.ownerCharID,
				}
				reloadNotif.WriteUint16(uint16(temp.Opcode()))
				temp.Build(reloadNotif, s.clientContext)
			}
			reloadNotif.WriteUint16(0x0010)
			s.QueueSend(reloadNotif.Data())
		} else {
			sendDisabledCommandMessage(s, commands["Reload"])
		}
	}

	if strings.HasPrefix(command, commands["KeyQuest"].Prefix) {
		if commands["KeyQuest"].Enabled {
			if strings.HasPrefix(command, fmt.Sprintf("%s get", commands["KeyQuest"].Prefix)) {
				sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandKqfGet"], s.kqf))
			} else if strings.HasPrefix(command, fmt.Sprintf("%s set", commands["KeyQuest"].Prefix)) {
				var hexs string
				n, numerr := fmt.Sscanf(command, fmt.Sprintf("%s set %%s", commands["KeyQuest"].Prefix), &hexs)
				if numerr != nil || n != 1 || len(hexs) != 16 {
					sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandKqfSetError"], commands["KeyQuest"].Prefix))
				} else {
					hexd, _ := hex.DecodeString(hexs)
					s.kqf = hexd
					s.kqfOverride = true
					sendServerChatMessage(s, s.server.dict["commandKqfSetSuccess"])
				}
			}
		} else {
			sendDisabledCommandMessage(s, commands["KeyQuest"])
		}
	}

	if strings.HasPrefix(command, commands["Rights"].Prefix) {
		// Set account rights
		if commands["Rights"].Enabled {
			var v uint32
			n, err := fmt.Sscanf(command, fmt.Sprintf("%s %%d", commands["Rights"].Prefix), &v)
			if err != nil || n != 1 {
				sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandRightsError"], commands["Rights"].Prefix))
			} else {
				_, err = s.server.db.Exec("UPDATE users u SET rights=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", v, s.charID)
				if err == nil {
					sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandRightsSuccess"], v))
				}
			}
		} else {
			sendDisabledCommandMessage(s, commands["Rights"])
		}
	}

	if strings.HasPrefix(command, commands["Course"].Prefix) {
		if commands["Course"].Enabled {
			var name string
			n, err := fmt.Sscanf(command, fmt.Sprintf("%s %%s", commands["Course"].Prefix), &name)
			if err != nil || n != 1 {
				sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandCourseError"], commands["Course"].Prefix))
			} else {
				name = strings.ToLower(name)
				for _, course := range mhfcourse.Courses() {
					for _, alias := range course.Aliases() {
						if strings.EqualFold(name, alias) {
							if slices.Contains(s.server.erupeConfig.Courses, config.Course{Name: course.Aliases()[0], Enabled: true}) {
								var delta, rightsInt uint32
								if mhfcourse.CourseExists(course.ID, s.courses) {
									ei := slices.IndexFunc(s.courses, func(c mhfcourse.Course) bool {
										for _, alias := range c.Aliases() {
											if strings.EqualFold(name, alias) {
												return true
											}
										}
										return false
									})
									if ei != -1 {
										delta = uint32(-1 * math.Pow(2, float64(course.ID)))
										sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandCourseDisabled"], course.Aliases()[0]))
									}
								} else {
									delta = uint32(math.Pow(2, float64(course.ID)))
									sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandCourseEnabled"], course.Aliases()[0]))
								}
								err = s.server.db.QueryRow("SELECT rights FROM users u INNER JOIN characters c ON u.id = c.user_id WHERE c.id = $1", s.charID).Scan(&rightsInt)
								if err == nil {
									s.server.db.Exec("UPDATE users u SET rights=$1 WHERE u.id=(SELECT c.user_id FROM characters c WHERE c.id=$2)", rightsInt+delta, s.charID)
								}
								updateRights(s)
							} else {
								sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandCourseLocked"], course.Aliases()[0]))
							}
							return
						}
					}
				}
				sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandCourseError"], commands["Course"].Prefix))
			}
		} else {
			sendDisabledCommandMessage(s, commands["Course"])
		}
	}

	if strings.HasPrefix(command, commands["Raviente"].Prefix) {
		if commands["Raviente"].Enabled {
			if getRaviSemaphore(s.server) != nil {
				s.server.raviente.Lock()
				if !strings.HasPrefix(command, "!ravi ") {
					sendServerChatMessage(s, s.server.dict["commandRaviNoCommand"])
				} else {
					if strings.HasPrefix(command, "!ravi start") {
						if s.server.raviente.register.startTime == 0 {
							s.server.raviente.register.startTime = s.server.raviente.register.postTime
							sendServerChatMessage(s, s.server.dict["commandRaviStartSuccess"])
							s.notifyRavi()
						} else {
							sendServerChatMessage(s, s.server.dict["commandRaviStartError"])
						}
					} else if strings.HasPrefix(command, "!ravi cm") || strings.HasPrefix(command, "!ravi checkmultiplier") {
						sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandRaviMultiplier"], s.server.raviente.GetRaviMultiplier(s.server)))
					} else if strings.HasPrefix(command, "!ravi sr") || strings.HasPrefix(command, "!ravi sendres") {
						if s.server.raviente.state.stateData[28] > 0 {
							sendServerChatMessage(s, s.server.dict["commandRaviResSuccess"])
							s.server.raviente.state.stateData[28] = 0
						} else {
							sendServerChatMessage(s, s.server.dict["commandRaviResError"])
						}
					} else if strings.HasPrefix(command, "!ravi ss") || strings.HasPrefix(command, "!ravi sendsed") {
						sendServerChatMessage(s, s.server.dict["commandRaviSedSuccess"])
						// Total BerRavi HP
						HP := s.server.raviente.state.stateData[0] + s.server.raviente.state.stateData[1] + s.server.raviente.state.stateData[2] + s.server.raviente.state.stateData[3] + s.server.raviente.state.stateData[4]
						s.server.raviente.support.supportData[1] = HP
					} else if strings.HasPrefix(command, "!ravi rs") || strings.HasPrefix(command, "!ravi reqsed") {
						sendServerChatMessage(s, s.server.dict["commandRaviRequest"])
						// Total BerRavi HP
						HP := s.server.raviente.state.stateData[0] + s.server.raviente.state.stateData[1] + s.server.raviente.state.stateData[2] + s.server.raviente.state.stateData[3] + s.server.raviente.state.stateData[4]
						s.server.raviente.support.supportData[1] = HP + 12
					} else {
						sendServerChatMessage(s, s.server.dict["commandRaviError"])
					}
				}
				s.server.raviente.Unlock()
			} else {
				sendServerChatMessage(s, s.server.dict["commandRaviNoPlayers"])
			}
		} else {
			sendDisabledCommandMessage(s, commands["Raviente"])
		}
	}

	if strings.HasPrefix(command, commands["Teleport"].Prefix) {
		if commands["Teleport"].Enabled {
			var x, y int16
			n, err := fmt.Sscanf(command, fmt.Sprintf("%s %%d %%d", commands["Teleport"].Prefix), &x, &y)
			if err != nil || n != 2 {
				sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandTeleportError"], commands["Teleport"].Prefix))
			} else {
				sendServerChatMessage(s, fmt.Sprintf(s.server.dict["commandTeleportSuccess"], x, y))

				// Make the inside of the casted binary
				payload := byteframe.NewByteFrame()
				payload.SetLE()
				payload.WriteUint8(2) // SetState type(position == 2)
				payload.WriteInt16(x) // X
				payload.WriteInt16(y) // Y
				payloadBytes := payload.Data()

				s.QueueSendMHF(&mhfpacket.MsgSysCastedBinary{
					CharID:         s.charID,
					MessageType:    BinaryMessageTypeState,
					RawDataPayload: payloadBytes,
				})
			}
		} else {
			sendDisabledCommandMessage(s, commands["Teleport"])
		}
	}
}

func handleMsgSysCastBinary(s *Session, p mhfpacket.MHFPacket) {
	pkt := p.(*mhfpacket.MsgSysCastBinary)
	tmp := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)

	if pkt.BroadcastType == 0x03 && pkt.MessageType == 0x03 && len(pkt.RawDataPayload) == 0x10 {
		if tmp.ReadUint16() == 0x0002 && tmp.ReadUint8() == 0x18 {
			_ = tmp.ReadBytes(9)
			tmp.SetLE()
			frame := tmp.ReadUint32()
			sendServerChatMessage(s, fmt.Sprintf("TIME : %d'%d.%03d (%dframe)", frame/30/60, frame/30%60, int(math.Round(float64(frame%30*100)/3)), frame))
		}
	}

	if s.server.erupeConfig.DevModeOptions.QuestDebugTools == true && s.server.erupeConfig.DevMode {
		if pkt.BroadcastType == 0x03 && pkt.MessageType == 0x02 && len(pkt.RawDataPayload) > 32 {
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
	var authorLen, msgLen uint16
	var msg []byte

	isDiceCommand := false
	if pkt.MessageType == BinaryMessageTypeChat {
		tmp.SetLE()
		tmp.Seek(int64(0), 0)
		_ = tmp.ReadUint32()
		authorLen = tmp.ReadUint16()
		msgLen = tmp.ReadUint16()
		msg = tmp.ReadNullTerminatedBytes()
	}

	// Customise payload
	realPayload := pkt.RawDataPayload
	if pkt.BroadcastType == BroadcastTypeTargeted {
		tmp.SetBE()
		tmp.Seek(int64(0), 0)
		msgBinTargeted = &binpacket.MsgBinTargeted{}
		err := msgBinTargeted.Parse(tmp)
		if err != nil {
			s.logger.Warn("Failed to parse targeted cast binary")
			return
		}
		realPayload = msgBinTargeted.RawDataPayload
	} else if pkt.MessageType == BinaryMessageTypeChat {
		if msgLen == 6 && string(msg) == "@dice" {
			isDiceCommand = true
			roll := byteframe.NewByteFrame()
			roll.WriteInt16(1) // Unk
			roll.SetLE()
			roll.WriteUint16(4) // Unk
			roll.WriteUint16(authorLen)
			dice := fmt.Sprintf("%d", token.RNG().Intn(100)+1)
			roll.WriteUint16(uint16(len(dice) + 1))
			roll.WriteNullTerminatedBytes([]byte(dice))
			roll.WriteNullTerminatedBytes(tmp.ReadNullTerminatedBytes())
			realPayload = roll.Data()
		} else {
			bf := byteframe.NewByteFrameFromBytes(pkt.RawDataPayload)
			bf.SetLE()
			chatMessage := &binpacket.MsgBinChat{}
			chatMessage.Parse(bf)
			if strings.HasPrefix(chatMessage.Message, "!") {
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
		if isDiceCommand {
			s.stage.BroadcastMHF(resp, nil) // send dice result back to caller
		} else {
			s.stage.BroadcastMHF(resp, s)
		}
	case BroadcastTypeServer:
		if pkt.MessageType == 1 {
			if getRaviSemaphore(s.server) != nil {
				s.server.BroadcastMHF(resp, s)
			}
		} else {
			s.server.BroadcastMHF(resp, s)
		}
	case BroadcastTypeTargeted:
		for _, targetID := range (*msgBinTargeted).TargetCharIDs {
			char := s.server.FindSessionByCharID(targetID)

			if char != nil {
				char.QueueSendMHF(resp)
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
