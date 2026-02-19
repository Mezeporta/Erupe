package channelserver

import (
	"erupe-ce/network/mhfpacket"
	"net"
)

// ChannelRegistry abstracts cross-channel operations behind an interface.
// The default LocalChannelRegistry wraps the in-process []*Server slice.
// Future implementations may use DB/Redis/NATS for multi-process deployments.
type ChannelRegistry interface {
	// Worldcast broadcasts a packet to all sessions across all channels.
	Worldcast(pkt mhfpacket.MHFPacket, ignoredSession *Session, ignoredChannel *Server)

	// FindSessionByCharID looks up a session by character ID across all channels.
	FindSessionByCharID(charID uint32) *Session

	// DisconnectUser disconnects all sessions belonging to the given character IDs.
	DisconnectUser(cids []uint32)

	// FindChannelForStage searches all channels for a stage whose ID has the
	// given suffix and returns the owning channel's GlobalID, or "" if not found.
	FindChannelForStage(stageSuffix string) string

	// SearchSessions searches sessions across all channels using a predicate,
	// returning up to max snapshot results.
	SearchSessions(predicate func(SessionSnapshot) bool, max int) []SessionSnapshot

	// SearchStages searches stages across all channels with a prefix filter,
	// returning up to max snapshot results.
	SearchStages(stagePrefix string, max int) []StageSnapshot

	// NotifyMailToCharID finds the session for charID and sends a mail notification.
	NotifyMailToCharID(charID uint32, sender *Session, mail *Mail)
}

// SessionSnapshot is an immutable copy of session data taken under lock.
type SessionSnapshot struct {
	CharID       uint32
	Name         string
	StageID      string
	ServerIP     net.IP
	ServerPort   uint16
	UserBinary3  []byte // Copy of userBinaryParts index 3
}

// StageSnapshot is an immutable copy of stage data taken under lock.
type StageSnapshot struct {
	ServerIP     net.IP
	ServerPort   uint16
	StageID      string
	ClientCount  int
	Reserved     int
	MaxPlayers   uint16
	RawBinData0  []byte
	RawBinData1  []byte
	RawBinData3  []byte
}
