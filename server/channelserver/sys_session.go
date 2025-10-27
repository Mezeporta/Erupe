package channelserver

import (
	"encoding/binary"
	"encoding/hex"
	"erupe-ce/common/mhfcourse"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/stringstack"
	"erupe-ce/network"
	"erupe-ce/network/clientctx"
	"erupe-ce/network/mhfpacket"
	"go.uber.org/zap"
)

// packet is an internal wrapper for queued outbound packets.
type packet struct {
	data        []byte // Raw packet bytes to send
	nonBlocking bool   // If true, drop packet if queue is full instead of blocking
}

// Session represents an active player connection to the channel server.
//
// Each Session manages a single player's connection lifecycle, including:
//   - Packet send/receive loops running in separate goroutines
//   - Current stage (game area) and stage movement history
//   - Character state (ID, courses, guild, etc.)
//   - Mail system state
//   - Quest/semaphore participation
//
// Lifecycle:
//  1. Created by NewSession() when a player connects
//  2. Started with Start() which launches send/recv goroutines
//  3. Processes packets through handlePacketGroup() -> handler functions
//  4. Cleaned up when connection closes or times out (30 second inactivity)
//
// Thread Safety:
// Session embeds sync.Mutex to protect mutable state. Most handler functions
// acquire the session lock when modifying session fields. The packet queue
// (sendPackets channel) is safe for concurrent access.
type Session struct {
	sync.Mutex // Protects session state during concurrent handler execution

	// Core connection and logging
	logger        *zap.Logger                // Logger with connection address
	server        *Server                    // Parent server reference
	rawConn       net.Conn                   // Underlying TCP connection
	cryptConn     *network.CryptConn         // Encrypted connection wrapper
	sendPackets   chan packet                // Outbound packet queue (buffered, size 20)
	clientContext *clientctx.ClientContext   // Client version and capabilities
	lastPacket    time.Time                  // Timestamp of last received packet (for timeout detection)

	// Stage (game area) state
	userEnteredStage bool    // Whether player has entered any stage during this session
	stageID          string  // Current stage ID string (e.g., "sl1Ns200p0a0u0")
	stage            *Stage  // Pointer to current stage object
	reservationStage *Stage  // Stage reserved for quest (used by unreserve packet)
	stagePass        string  // Temporary password storage for password-protected stages
	stageMoveStack   *stringstack.StringStack // Navigation history for "back" functionality

	// Player identity and state
	charID       uint32               // Character ID for this session
	Name         string               // Character name (for debugging/logging)
	prevGuildID  uint32               // Last guild ID queried (cached for InfoGuild)
	token        string               // Authentication token from sign server
	logKey       []byte               // Logging encryption key
	sessionStart int64                // Session start timestamp (Unix time)
	courses      []mhfcourse.Course   // Active Monster Hunter courses (buffs/subscriptions)
	kqf          []byte               // Key Quest Flags (quest progress tracking)
	kqfOverride  bool                 // Whether KQF is being overridden

	// Quest/event coordination
	semaphore *Semaphore // Semaphore for quest/event participation (if in a coordinated activity)

	// Mail system state
	// The mail system uses an accumulated index system where the client tracks
	// mail by incrementing indices rather than direct mail IDs
	mailAccIndex uint8 // Current accumulated mail index for this session
	mailList     []int // Maps accumulated indices to actual mail IDs

	// Connection state
	closed bool // Whether connection has been closed (prevents double-cleanup)
}

// NewSession creates and initializes a new Session for an incoming connection.
//
// The session is created with:
//   - A logger tagged with the connection's remote address
//   - An encrypted connection wrapper
//   - A buffered packet send queue (size 20)
//   - Initialized stage movement stack for navigation
//   - Session start time set to current time
//
// After creation, call Start() to begin processing packets.
func NewSession(server *Server, conn net.Conn) *Session {
	s := &Session{
		logger:         server.logger.Named(conn.RemoteAddr().String()),
		server:         server,
		rawConn:        conn,
		cryptConn:      network.NewCryptConn(conn),
		sendPackets:    make(chan packet, 20),
		clientContext:  &clientctx.ClientContext{}, // Unused
		lastPacket:     time.Now(),
		sessionStart:   TimeAdjusted().Unix(),
		stageMoveStack: stringstack.New(),
	}
	return s
}

// Start begins the session's packet processing by launching send and receive goroutines.
//
// This method spawns two long-running goroutines:
//  1. sendLoop(): Continuously sends queued packets to the client
//  2. recvLoop(): Continuously receives and processes packets from the client
//
// The receive loop handles packet parsing, routing to handlers, and recursive
// packet group processing (when multiple packets arrive in one read).
//
// Both loops run until the connection closes or times out. Unlike the sign and
// entrance servers, the channel server does NOT expect an 8-byte NULL initialization.
func (s *Session) Start() {
	go func() {
		s.logger.Debug("New connection", zap.String("RemoteAddr", s.rawConn.RemoteAddr().String()))
		// Unlike the sign and entrance server,
		// the client DOES NOT initalize the channel connection with 8 NULL bytes.
		go s.sendLoop()
		s.recvLoop()
	}()
}

// QueueSend queues a packet for transmission to the client (blocking).
//
// This method:
//  1. Logs the outbound packet (if dev mode is enabled)
//  2. Attempts to enqueue the packet to the send channel
//  3. If the queue is full, flushes non-blocking packets and retries
//
// Blocking vs Non-blocking:
// This is a blocking send - if the queue fills, it will flush non-blocking
// packets (broadcasts, non-critical messages) to make room for this packet.
// Use QueueSendNonBlocking() for packets that can be safely dropped.
//
// Thread Safety: Safe for concurrent calls from multiple goroutines.
func (s *Session) QueueSend(data []byte) {
	// FIX: Check data length before reading opcode to prevent crash on empty packets
	if len(data) >= 2 {
		s.logMessage(binary.BigEndian.Uint16(data[0:2]), data, "Server", s.Name)
	}
	select {
	case s.sendPackets <- packet{data, false}:
		// Enqueued data
	default:
		s.logger.Warn("Packet queue too full, flushing!")
		var tempPackets []packet
		for len(s.sendPackets) > 0 {
			tempPacket := <-s.sendPackets
			if !tempPacket.nonBlocking {
				tempPackets = append(tempPackets, tempPacket)
			}
		}
		for _, tempPacket := range tempPackets {
			s.sendPackets <- tempPacket
		}
		s.sendPackets <- packet{data, false}
	}
}

// QueueSendNonBlocking queues a packet for transmission (non-blocking, lossy).
//
// Unlike QueueSend(), this method drops the packet immediately if the send queue
// is full. This is used for broadcast messages, stage updates, and other packets
// where occasional packet loss is acceptable (client will re-sync or request again).
//
// Use cases:
//   - Stage broadcasts (player movement, chat)
//   - Server-wide announcements
//   - Non-critical status updates
//
// Thread Safety: Safe for concurrent calls from multiple goroutines.
func (s *Session) QueueSendNonBlocking(data []byte) {
	select {
	case s.sendPackets <- packet{data, true}:
		if len(data) >= 2 {
			s.logMessage(binary.BigEndian.Uint16(data[0:2]), data, "Server", s.Name)
		}
	default:
		s.logger.Warn("Packet queue too full, dropping!")
	}
}

// QueueSendMHF queues a structured MHFPacket for transmission to the client.
//
// This is a convenience method that:
//  1. Creates a byteframe and writes the packet opcode
//  2. Calls the packet's Build() method to serialize its data
//  3. Queues the resulting bytes using QueueSend()
//
// The packet is built with the session's clientContext, allowing version-specific
// packet formatting when needed.
func (s *Session) QueueSendMHF(pkt mhfpacket.MHFPacket) {
	// Make the header
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(pkt.Opcode()))

	// Build the packet onto the byteframe.
	pkt.Build(bf, s.clientContext)

	// Queue it.
	s.QueueSend(bf.Data())
}

// QueueAck sends an acknowledgment packet with optional response data.
//
// Many client packets include an "ack handle" field - a unique identifier the client
// uses to match responses to requests. This method constructs and queues a MSG_SYS_ACK
// packet containing the ack handle and response data.
//
// Parameters:
//   - ackHandle: The ack handle from the original client packet
//   - data: Response payload bytes (can be empty for simple acks)
func (s *Session) QueueAck(ackHandle uint32, data []byte) {
	bf := byteframe.NewByteFrame()
	bf.WriteUint16(uint16(network.MSG_SYS_ACK))
	bf.WriteUint32(ackHandle)
	bf.WriteBytes(data)
	s.QueueSend(bf.Data())
}

func (s *Session) sendLoop() {
	for {
		if s.closed {
			return
		}
		// Send each packet individually with its own terminator
		for len(s.sendPackets) > 0 {
			pkt := <-s.sendPackets
			err := s.cryptConn.SendPacket(append(pkt.data, []byte{0x00, 0x10}...))
			if err != nil {
				s.logger.Warn("Failed to send packet", zap.Error(err))
			}
		}
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *Session) recvLoop() {
	for {
		if time.Now().Add(-30 * time.Second).After(s.lastPacket) {
			logoutPlayer(s)
			return
		}
		if s.closed {
			logoutPlayer(s)
			return
		}
		pkt, err := s.cryptConn.ReadPacket()
		if err == io.EOF {
			s.logger.Info(fmt.Sprintf("[%s] Disconnected", s.Name))
			logoutPlayer(s)
			return
		}
		if err != nil {
			s.logger.Warn("Error on ReadPacket, exiting recv loop", zap.Error(err))
			logoutPlayer(s)
			return
		}
		s.handlePacketGroup(pkt)
		time.Sleep(10 * time.Millisecond)
	}
}

func (s *Session) handlePacketGroup(pktGroup []byte) {
	s.lastPacket = time.Now()
	bf := byteframe.NewByteFrameFromBytes(pktGroup)
	opcodeUint16 := bf.ReadUint16()
	opcode := network.PacketID(opcodeUint16)

	// This shouldn't be needed, but it's better to recover and let the connection die than to panic the server.
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[%s]", s.Name)
			fmt.Println("Recovered from panic", r)
		}
	}()

	s.logMessage(opcodeUint16, pktGroup, s.Name, "Server")

	if opcode == network.MSG_SYS_LOGOUT {
		s.closed = true
		return
	}
	// Get the packet parser and handler for this opcode.
	mhfPkt := mhfpacket.FromOpcode(opcode)
	if mhfPkt == nil {
		fmt.Println("Got opcode which we don't know how to parse, can't parse anymore for this group")
		return
	}
	// Parse the packet.
	err := mhfPkt.Parse(bf, s.clientContext)
	if err != nil {
		fmt.Printf("\n!!! [%s] %s NOT IMPLEMENTED !!! \n\n\n", s.Name, opcode)
		return
	}
	// Handle the packet.
	handlerTable[opcode](s, mhfPkt)
	// If there is more data on the stream that the .Parse method didn't read, then read another packet off it.
	remainingData := bf.DataFromCurrent()
	if len(remainingData) >= 2 {
		s.handlePacketGroup(remainingData)
	}
}

func ignored(opcode network.PacketID) bool {
	ignoreList := []network.PacketID{
		network.MSG_SYS_END,
		network.MSG_SYS_PING,
		network.MSG_SYS_NOP,
		network.MSG_SYS_TIME,
		network.MSG_SYS_EXTEND_THRESHOLD,
		network.MSG_SYS_POSITION_OBJECT,
		network.MSG_MHF_SAVEDATA,
	}
	set := make(map[network.PacketID]struct{}, len(ignoreList))
	for _, s := range ignoreList {
		set[s] = struct{}{}
	}
	_, r := set[opcode]
	return r
}

func (s *Session) logMessage(opcode uint16, data []byte, sender string, recipient string) {
	if !s.server.erupeConfig.DevMode {
		return
	}

	if sender == "Server" && !s.server.erupeConfig.DevModeOptions.LogOutboundMessages {
		return
	} else if !s.server.erupeConfig.DevModeOptions.LogInboundMessages {
		return
	}

	opcodePID := network.PacketID(opcode)
	if ignored(opcodePID) {
		return
	}
	fmt.Printf("[%s] -> [%s]\n", sender, recipient)
	fmt.Printf("Opcode: %s\n", opcodePID)
	if len(data) <= s.server.erupeConfig.DevModeOptions.MaxHexdumpLength {
		fmt.Printf("Data [%d bytes]:\n%s\n", len(data), hex.Dump(data))
	} else {
		fmt.Printf("Data [%d bytes]:\n(Too long!)\n\n", len(data))
	}
}
