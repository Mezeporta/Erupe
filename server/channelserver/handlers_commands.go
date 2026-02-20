package channelserver

import (
	"crypto/rand"
	"encoding/hex"
	"erupe-ce/common/byteframe"
	"erupe-ce/common/mhfcid"
	"erupe-ce/common/mhfcourse"
	"erupe-ce/config"
	"erupe-ce/network"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"fmt"
	"math"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	commands     map[string]_config.Command
	commandsOnce sync.Once
)

func initCommands(cmds []_config.Command, logger *zap.Logger) {
	commandsOnce.Do(func() {
		commands = make(map[string]_config.Command)
		for _, cmd := range cmds {
			commands[cmd.Name] = cmd
			if cmd.Enabled {
				logger.Info(fmt.Sprintf("Command %s: Enabled, prefix: %s", cmd.Name, cmd.Prefix))
			} else {
				logger.Info(fmt.Sprintf("Command %s: Disabled", cmd.Name))
			}
		}
	})
}

func sendDisabledCommandMessage(s *Session, cmd _config.Command) {
	sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.disabled, cmd.Name))
}

const chatFlagServer = 0x80 // marks a message as server-originated

func sendServerChatMessage(s *Session, message string) {
	// Make the inside of the casted binary
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	msgBinChat := &binpacket.MsgBinChat{
		Unk0:       0,
		Type:       5,
		Flags:      chatFlagServer,
		Message:    message,
		SenderName: "Erupe",
	}
	_ = msgBinChat.Build(bf)

	castedBin := &mhfpacket.MsgSysCastedBinary{
		CharID:         0,
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}

	s.QueueSendMHFNonBlocking(castedBin)
}

func parseChatCommand(s *Session, command string) {
	args := strings.Split(command[len(s.server.erupeConfig.CommandPrefix):], " ")
	switch args[0] {
	case commands["Ban"].Prefix:
		if s.isOp() {
			if len(args) > 1 {
				var expiry time.Time
				if len(args) > 2 {
					var length int
					var unit string
					n, err := fmt.Sscanf(args[2], `%d%s`, &length, &unit)
					if err == nil && n == 2 {
						switch unit {
						case "s", "second", "seconds":
							expiry = time.Now().Add(time.Duration(length) * time.Second)
						case "m", "mi", "minute", "minutes":
							expiry = time.Now().Add(time.Duration(length) * time.Minute)
						case "h", "hour", "hours":
							expiry = time.Now().Add(time.Duration(length) * time.Hour)
						case "d", "day", "days":
							expiry = time.Now().Add(time.Duration(length) * time.Hour * 24)
						case "mo", "month", "months":
							expiry = time.Now().Add(time.Duration(length) * time.Hour * 24 * 30)
						case "y", "year", "years":
							expiry = time.Now().Add(time.Duration(length) * time.Hour * 24 * 365)
						}
					} else {
						sendServerChatMessage(s, s.server.i18n.commands.ban.error)
						return
					}
				}
				cid := mhfcid.ConvertCID(args[1])
				if cid > 0 {
					uid, uname, err := s.server.userRepo.GetByIDAndUsername(cid)
					if err == nil {
						if expiry.IsZero() {
							if _, err := s.server.db.Exec(`INSERT INTO bans VALUES ($1)
                 				ON CONFLICT (user_id) DO UPDATE SET expires=NULL`, uid); err != nil {
								s.logger.Error("Failed to ban user", zap.Error(err))
							}
							sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.ban.success, uname))
						} else {
							if _, err := s.server.db.Exec(`INSERT INTO bans VALUES ($1, $2)
                 				ON CONFLICT (user_id) DO UPDATE SET expires=$2`, uid, expiry); err != nil {
								s.logger.Error("Failed to ban user with expiry", zap.Error(err))
							}
							sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.ban.success, uname)+fmt.Sprintf(s.server.i18n.commands.ban.length, expiry.Format(time.DateTime)))
						}
						s.server.DisconnectUser(uid)
					} else {
						sendServerChatMessage(s, s.server.i18n.commands.ban.noUser)
					}
				} else {
					sendServerChatMessage(s, s.server.i18n.commands.ban.invalid)
				}
			} else {
				sendServerChatMessage(s, s.server.i18n.commands.ban.error)
			}
		} else {
			sendServerChatMessage(s, s.server.i18n.commands.noOp)
		}
	case commands["Timer"].Prefix:
		if commands["Timer"].Enabled || s.isOp() {
			state, err := s.server.userRepo.GetTimer(s.userID)
			if err != nil {
				s.logger.Error("Failed to get timer state", zap.Error(err))
			}
			if err := s.server.userRepo.SetTimer(s.userID, !state); err != nil {
				s.logger.Error("Failed to update timer setting", zap.Error(err))
			}
			if state {
				sendServerChatMessage(s, s.server.i18n.commands.timer.disabled)
			} else {
				sendServerChatMessage(s, s.server.i18n.commands.timer.enabled)
			}
		} else {
			sendDisabledCommandMessage(s, commands["Timer"])
		}
	case commands["PSN"].Prefix:
		if commands["PSN"].Enabled || s.isOp() {
			if len(args) > 1 {
				exists, err := s.server.userRepo.CountByPSNID(args[1])
				if err != nil {
					s.logger.Error("Failed to check PSN ID existence", zap.Error(err))
				}
				if exists == 0 {
					err := s.server.userRepo.SetPSNID(s.userID, args[1])
					if err == nil {
						sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.psn.success, args[1]))
					}
				} else {
					sendServerChatMessage(s, s.server.i18n.commands.psn.exists)
				}
			} else {
				sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.psn.error, commands["PSN"].Prefix))
			}
		} else {
			sendDisabledCommandMessage(s, commands["PSN"])
		}
	case commands["Reload"].Prefix:
		if commands["Reload"].Enabled || s.isOp() {
			sendServerChatMessage(s, s.server.i18n.commands.reload)
			var temp mhfpacket.MHFPacket
			deleteNotif := byteframe.NewByteFrame()
			for _, object := range s.stage.objects {
				if object.ownerCharID == s.charID {
					continue
				}
				temp = &mhfpacket.MsgSysDeleteObject{ObjID: object.id}
				deleteNotif.WriteUint16(uint16(temp.Opcode()))
				_ = temp.Build(deleteNotif, s.clientContext)
			}
			for _, session := range s.server.sessions {
				if s == session {
					continue
				}
				temp = &mhfpacket.MsgSysDeleteUser{CharID: session.charID}
				deleteNotif.WriteUint16(uint16(temp.Opcode()))
				_ = temp.Build(deleteNotif, s.clientContext)
			}
			deleteNotif.WriteUint16(uint16(network.MSG_SYS_END))
			s.QueueSendNonBlocking(deleteNotif.Data())
			time.Sleep(500 * time.Millisecond)
			reloadNotif := byteframe.NewByteFrame()
			for _, session := range s.server.sessions {
				if s == session {
					continue
				}
				temp = &mhfpacket.MsgSysInsertUser{CharID: session.charID}
				reloadNotif.WriteUint16(uint16(temp.Opcode()))
				_ = temp.Build(reloadNotif, s.clientContext)
				for i := 0; i < 3; i++ {
					temp = &mhfpacket.MsgSysNotifyUserBinary{
						CharID:     session.charID,
						BinaryType: uint8(i + 1),
					}
					reloadNotif.WriteUint16(uint16(temp.Opcode()))
					_ = temp.Build(reloadNotif, s.clientContext)
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
				_ = temp.Build(reloadNotif, s.clientContext)
			}
			reloadNotif.WriteUint16(uint16(network.MSG_SYS_END))
			s.QueueSendNonBlocking(reloadNotif.Data())
		} else {
			sendDisabledCommandMessage(s, commands["Reload"])
		}
	case commands["KeyQuest"].Prefix:
		if commands["KeyQuest"].Enabled || s.isOp() {
			if s.server.erupeConfig.RealClientMode < _config.G10 {
				sendServerChatMessage(s, s.server.i18n.commands.kqf.version)
			} else {
				if len(args) > 1 {
					switch args[1] {
					case "get":
						sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.kqf.get, s.kqf))
					case "set":
						if len(args) > 2 && len(args[2]) == 16 {
							hexd, _ := hex.DecodeString(args[2])
							s.kqf = hexd
							s.kqfOverride = true
							sendServerChatMessage(s, s.server.i18n.commands.kqf.set.success)
						} else {
							sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.kqf.set.error, commands["KeyQuest"].Prefix))
						}
					}
				}
			}
		} else {
			sendDisabledCommandMessage(s, commands["KeyQuest"])
		}
	case commands["Rights"].Prefix:
		if commands["Rights"].Enabled || s.isOp() {
			if len(args) > 1 {
				v, _ := strconv.Atoi(args[1])
				err := s.server.userRepo.SetRights(s.userID, uint32(v))
				if err == nil {
					sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.rights.success, v))
				} else {
					sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.rights.error, commands["Rights"].Prefix))
				}
			} else {
				sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.rights.error, commands["Rights"].Prefix))
			}
		} else {
			sendDisabledCommandMessage(s, commands["Rights"])
		}
	case commands["Course"].Prefix:
		if commands["Course"].Enabled || s.isOp() {
			if len(args) > 1 {
				for _, course := range mhfcourse.Courses() {
					for _, alias := range course.Aliases() {
						if strings.EqualFold(args[1], alias) {
							if slices.Contains(s.server.erupeConfig.Courses, _config.Course{Name: course.Aliases()[0], Enabled: true}) {
								var delta uint32
								if mhfcourse.CourseExists(course.ID, s.courses) {
									ei := slices.IndexFunc(s.courses, func(c mhfcourse.Course) bool {
										for _, alias := range c.Aliases() {
											if strings.EqualFold(args[1], alias) {
												return true
											}
										}
										return false
									})
									if ei != -1 {
										delta = uint32(-1 * math.Pow(2, float64(course.ID)))
										sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.course.disabled, course.Aliases()[0]))
									}
								} else {
									delta = uint32(math.Pow(2, float64(course.ID)))
									sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.course.enabled, course.Aliases()[0]))
								}
								rightsInt, err := s.server.userRepo.GetRights(s.userID)
								if err == nil {
									if err := s.server.userRepo.SetRights(s.userID, rightsInt+delta); err != nil {
										s.logger.Error("Failed to update user rights", zap.Error(err))
									}
								}
								updateRights(s)
							} else {
								sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.course.locked, course.Aliases()[0]))
							}
							return
						}
					}
				}
			} else {
				sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.course.error, commands["Course"].Prefix))
			}
		} else {
			sendDisabledCommandMessage(s, commands["Course"])
		}
	case commands["Raviente"].Prefix:
		if commands["Raviente"].Enabled || s.isOp() {
			if len(args) > 1 {
				if s.server.getRaviSemaphore() != nil {
					switch args[1] {
					case "start":
						if s.server.raviente.register[1] == 0 {
							s.server.raviente.register[1] = s.server.raviente.register[3]
							sendServerChatMessage(s, s.server.i18n.commands.ravi.start.success)
							s.notifyRavi()
						} else {
							sendServerChatMessage(s, s.server.i18n.commands.ravi.start.error)
						}
					case "cm", "check", "checkmultiplier", "multiplier":
						sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.ravi.multiplier, s.server.GetRaviMultiplier()))
					case "sr", "sendres", "resurrection", "ss", "sendsed", "rs", "reqsed":
						if s.server.erupeConfig.RealClientMode == _config.ZZ {
							switch args[1] {
							case "sr", "sendres", "resurrection":
								if s.server.raviente.state[28] > 0 {
									sendServerChatMessage(s, s.server.i18n.commands.ravi.res.success)
									s.server.raviente.state[28] = 0
								} else {
									sendServerChatMessage(s, s.server.i18n.commands.ravi.res.error)
								}
							case "ss", "sendsed":
								sendServerChatMessage(s, s.server.i18n.commands.ravi.sed.success)
								// Total BerRavi HP
								HP := s.server.raviente.state[0] + s.server.raviente.state[1] + s.server.raviente.state[2] + s.server.raviente.state[3] + s.server.raviente.state[4]
								s.server.raviente.support[1] = HP
							case "rs", "reqsed":
								sendServerChatMessage(s, s.server.i18n.commands.ravi.request)
								// Total BerRavi HP
								HP := s.server.raviente.state[0] + s.server.raviente.state[1] + s.server.raviente.state[2] + s.server.raviente.state[3] + s.server.raviente.state[4]
								s.server.raviente.support[1] = HP + 1
							}
						} else {
							sendServerChatMessage(s, s.server.i18n.commands.ravi.version)
						}
					default:
						sendServerChatMessage(s, s.server.i18n.commands.ravi.error)
					}
				} else {
					sendServerChatMessage(s, s.server.i18n.commands.ravi.noPlayers)
				}
			} else {
				sendServerChatMessage(s, s.server.i18n.commands.ravi.error)
			}
		} else {
			sendDisabledCommandMessage(s, commands["Raviente"])
		}
	case commands["Teleport"].Prefix:
		if commands["Teleport"].Enabled || s.isOp() {
			if len(args) > 2 {
				x, _ := strconv.ParseInt(args[1], 10, 16)
				y, _ := strconv.ParseInt(args[2], 10, 16)
				payload := byteframe.NewByteFrame()
				payload.SetLE()
				payload.WriteUint8(2)        // SetState type(position == 2)
				payload.WriteInt16(int16(x)) // X
				payload.WriteInt16(int16(y)) // Y
				payloadBytes := payload.Data()
				s.QueueSendMHFNonBlocking(&mhfpacket.MsgSysCastedBinary{
					CharID:         s.charID,
					MessageType:    BinaryMessageTypeState,
					RawDataPayload: payloadBytes,
				})
				sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.teleport.success, x, y))
			} else {
				sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.teleport.error, commands["Teleport"].Prefix))
			}
		} else {
			sendDisabledCommandMessage(s, commands["Teleport"])
		}
	case commands["Discord"].Prefix:
		if commands["Discord"].Enabled || s.isOp() {
			_token, err := s.server.userRepo.GetDiscordToken(s.userID)
			if err != nil {
				randToken := make([]byte, 4)
				_, _ = rand.Read(randToken)
				_token = fmt.Sprintf("%x-%x", randToken[:2], randToken[2:])
				if err := s.server.userRepo.SetDiscordToken(s.userID, _token); err != nil {
					s.logger.Error("Failed to update discord token", zap.Error(err))
				}
			}
			sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.discord.success, _token))
		} else {
			sendDisabledCommandMessage(s, commands["Discord"])
		}
	case commands["Playtime"].Prefix:
		if commands["Playtime"].Enabled || s.isOp() {
			playtime := s.playtime + uint32(time.Since(s.playtimeTime).Seconds())
			sendServerChatMessage(s, fmt.Sprintf(s.server.i18n.commands.playtime, playtime/60/60, playtime/60%60, playtime%60))
		} else {
			sendDisabledCommandMessage(s, commands["Playtime"])
		}
	case commands["Help"].Prefix:
		if commands["Help"].Enabled || s.isOp() {
			for _, command := range commands {
				if command.Enabled || s.isOp() {
					sendServerChatMessage(s, fmt.Sprintf("%s%s: %s", s.server.erupeConfig.CommandPrefix, command.Prefix, command.Description))
				}
			}
		} else {
			sendDisabledCommandMessage(s, commands["Help"])
		}
	}
}
