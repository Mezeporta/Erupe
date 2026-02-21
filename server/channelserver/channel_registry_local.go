package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"net"
	"strings"
)

// LocalChannelRegistry is the in-process ChannelRegistry backed by []*Server.
type LocalChannelRegistry struct {
	channels []*Server
}

// NewLocalChannelRegistry creates a LocalChannelRegistry wrapping the given channels.
func NewLocalChannelRegistry(channels []*Server) *LocalChannelRegistry {
	return &LocalChannelRegistry{channels: channels}
}

func (r *LocalChannelRegistry) Worldcast(pkt mhfpacket.MHFPacket, ignoredSession *Session, ignoredChannel *Server) {
	for _, c := range r.channels {
		if c == ignoredChannel {
			continue
		}
		c.BroadcastMHF(pkt, ignoredSession)
	}
}

func (r *LocalChannelRegistry) FindSessionByCharID(charID uint32) *Session {
	for _, c := range r.channels {
		c.Lock()
		for _, session := range c.sessions {
			if session.charID == charID {
				c.Unlock()
				return session
			}
		}
		c.Unlock()
	}
	return nil
}

func (r *LocalChannelRegistry) DisconnectUser(cids []uint32) {
	for _, c := range r.channels {
		c.Lock()
		for _, session := range c.sessions {
			for _, cid := range cids {
				if session.charID == cid {
					_ = session.rawConn.Close()
					break
				}
			}
		}
		c.Unlock()
	}
}

func (r *LocalChannelRegistry) FindChannelForStage(stageSuffix string) string {
	for _, channel := range r.channels {
		channel.stagesLock.RLock()
		for id := range channel.stages {
			if strings.HasSuffix(id, stageSuffix) {
				gid := channel.GlobalID
				channel.stagesLock.RUnlock()
				return gid
			}
		}
		channel.stagesLock.RUnlock()
	}
	return ""
}

func (r *LocalChannelRegistry) SearchSessions(predicate func(SessionSnapshot) bool, max int) []SessionSnapshot {
	var results []SessionSnapshot
	for _, c := range r.channels {
		if len(results) >= max {
			break
		}
		c.Lock()
		for _, session := range c.sessions {
			if len(results) >= max {
				break
			}
			snap := SessionSnapshot{
				CharID:     session.charID,
				Name:       session.Name,
				ServerIP:   net.ParseIP(c.IP).To4(),
				ServerPort: c.Port,
			}
			if session.stage != nil {
				snap.StageID = session.stage.id
			}
			snap.UserBinary3 = c.userBinary.GetCopy(session.charID, 3)
			if predicate(snap) {
				results = append(results, snap)
			}
		}
		c.Unlock()
	}
	return results
}

func (r *LocalChannelRegistry) SearchStages(stagePrefix string, max int) []StageSnapshot {
	var results []StageSnapshot
	for _, c := range r.channels {
		if len(results) >= max {
			break
		}
		c.stagesLock.RLock()
		for _, stage := range c.stages {
			if len(results) >= max {
				break
			}
			if !strings.HasPrefix(stage.id, stagePrefix) {
				continue
			}
			stage.RLock()
			bin0 := stage.rawBinaryData[stageBinaryKey{1, 0}]
			bin0Copy := make([]byte, len(bin0))
			copy(bin0Copy, bin0)
			bin1 := stage.rawBinaryData[stageBinaryKey{1, 1}]
			bin1Copy := make([]byte, len(bin1))
			copy(bin1Copy, bin1)
			bin3 := stage.rawBinaryData[stageBinaryKey{1, 3}]
			bin3Copy := make([]byte, len(bin3))
			copy(bin3Copy, bin3)

			results = append(results, StageSnapshot{
				ServerIP:    net.ParseIP(c.IP).To4(),
				ServerPort:  c.Port,
				StageID:     stage.id,
				ClientCount: len(stage.clients) + len(stage.reservedClientSlots),
				Reserved:    len(stage.reservedClientSlots),
				MaxPlayers:  stage.maxPlayers,
				RawBinData0: bin0Copy,
				RawBinData1: bin1Copy,
				RawBinData3: bin3Copy,
			})
			stage.RUnlock()
		}
		c.stagesLock.RUnlock()
	}
	return results
}

func (r *LocalChannelRegistry) NotifyMailToCharID(charID uint32, sender *Session, mail *Mail) {
	session := r.FindSessionByCharID(charID)
	if session != nil {
		SendMailNotification(sender, mail, session)
	}
}
