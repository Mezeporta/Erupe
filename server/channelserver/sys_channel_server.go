// Package channelserver implements the Monster Hunter Frontier channel server.
//
// The channel server is the core gameplay component that handles actual game sessions,
// quests, player interactions, and all in-game activities. It uses a stage-based
// architecture where players move between stages (game areas/rooms) and interact
// with other players in real-time.
//
// Architecture Overview:
//
// The channel server manages three primary concepts:
//   - Sessions: Individual player connections with their state and packet queues
//   - Stages: Game rooms/areas where players interact (towns, quests, lobbies)
//   - Semaphores: Resource locks for coordinating multiplayer activities (quests, events)
//
// Multiple channel servers can run simultaneously on different ports, allowing
// horizontal scaling and separation of different world types (Newbie, Normal, etc).
//
// Thread Safety:
//
// This package extensively uses goroutines and shared state. All shared resources
// are protected by mutexes. When modifying code, always consider thread safety:
//   - Server-level: s.Lock() / s.Unlock() for session map
//   - Stage-level: s.stagesLock.RLock() / s.stagesLock.Lock() for stage map
//   - Session-level: session.Lock() / session.Unlock() for session state
//
// Use 'go test -race ./...' to detect race conditions during development.
package channelserver

import (
	"fmt"
	"net"
	"sync"

	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
	"erupe-ce/config"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/discordbot"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config holds configuration parameters for creating a new channel server.
type Config struct {
	ID          uint16                 // Channel server ID (unique identifier)
	Logger      *zap.Logger            // Logger instance for this channel server
	DB          *sqlx.DB               // Database connection pool
	DiscordBot  *discordbot.DiscordBot // Optional Discord bot for chat integration
	ErupeConfig *config.Config         // Global Erupe configuration
	Name        string                 // Display name for the server (shown in broadcasts)
	Enable      bool                   // Whether this server is enabled
}

// userBinaryPartID is a composite key for identifying a specific part of a user's binary data.
// User binary data is split into multiple indexed parts and stored separately.
type userBinaryPartID struct {
	charID uint32 // Character ID who owns this binary data
	index  uint8  // Part index (binary data is chunked into multiple parts)
}

// Server represents a Monster Hunter Frontier channel server instance.
//
// The Server manages all active player sessions, game stages, and shared resources.
// It runs two main goroutines: one for accepting connections and one for managing
// the session lifecycle.
//
// Thread Safety:
// Server embeds sync.Mutex for protecting the sessions map. Use Lock()/Unlock()
// when reading or modifying s.sessions. The stages map uses a separate RWMutex
// (stagesLock) to allow concurrent reads during normal gameplay.
type Server struct {
	sync.Mutex // Protects sessions map and isShuttingDown flag

	// Server identity and network configuration
	Channels []*Server // Reference to all channel servers (for world broadcasts)
	ID       uint16    // This server's ID
	GlobalID string    // Global identifier string
	IP       string    // Server IP address
	Port     uint16    // Server listening port

	// Core dependencies
	logger      *zap.Logger    // Logger instance
	db          *sqlx.DB       // Database connection pool
	erupeConfig *config.Config // Global configuration

	// Connection management
	acceptConns    chan net.Conn         // Channel for new accepted connections
	deleteConns    chan net.Conn         // Channel for connections to be cleaned up
	sessions       map[net.Conn]*Session // Active sessions keyed by connection
	listener       net.Listener          // TCP listener (created when Server.Start is called)
	isShuttingDown bool                  // Shutdown flag to stop goroutines gracefully

	// Stage (game room) management
	stagesLock sync.RWMutex      // Protects stages map (RWMutex for concurrent reads)
	stages     map[string]*Stage // Active stages keyed by stage ID string

	// Localization
	dict map[string]string // Language string mappings for server messages

	// User binary data storage
	// Binary data is player-specific custom data that the client stores on the server
	userBinaryPartsLock sync.RWMutex                // Protects userBinaryParts map
	userBinaryParts     map[userBinaryPartID][]byte // Chunked binary data by character

	// Semaphore (multiplayer coordination) management
	semaphoreLock  sync.RWMutex          // Protects semaphore map and semaphoreIndex
	semaphore      map[string]*Semaphore // Active semaphores keyed by semaphore ID
	semaphoreIndex uint32                // Auto-incrementing ID for new semaphores (starts at 7)

	// Optional integrations
	discordBot *discordbot.DiscordBot // Discord bot for chat relay (nil if disabled)
	name       string                 // Server display name (used in chat messages)

	// Special event system: Raviente (large-scale multiplayer raid)
	raviente *Raviente // Raviente event state and coordination
}

// Raviente manages the Raviente raid event, a large-scale multiplayer encounter.
//
// Raviente is a special monster that requires coordination between many players
// across multiple phases. This struct tracks registration, event state, and
// support/assistance data for the active Raviente encounter.
type Raviente struct {
	sync.Mutex // Protects all Raviente data during concurrent access

	register *RavienteRegister // Player registration and event timing
	state    *RavienteState    // Current state of the Raviente encounter
	support  *RavienteSupport  // Support/assistance tracking data
}

// RavienteRegister tracks player registration and timing for Raviente events.
type RavienteRegister struct {
	nextTime     uint32   // Timestamp for next Raviente event
	startTime    uint32   // Event start timestamp
	postTime     uint32   // Event post-completion timestamp
	killedTime   uint32   // Timestamp when Raviente was defeated
	ravienteType uint32   // Raviente variant (2=Berserk, 3=Extreme, 4=Extreme Limited, 5=Berserk Small)
	maxPlayers   uint32   // Maximum players allowed (determines scaling)
	carveQuest   uint32   // Quest ID for carving phase after defeat
	register     []uint32 // List of registered player IDs (up to 5 slots)
}

// RavienteState holds the dynamic state data for an active Raviente encounter.
// The state array contains 29 uint32 values tracking encounter progress.
type RavienteState struct {
	stateData []uint32 // Raviente encounter state (29 values)
}

// RavienteSupport tracks support and assistance data for Raviente encounters.
// The support array contains 25 uint32 values for coordination features.
type RavienteSupport struct {
	supportData []uint32 // Support/assistance data (25 values)
}

// NewRaviente creates and initializes a new Raviente event manager with default values.
// All state and support arrays are initialized to zero, ready for a new event.
func NewRaviente() *Raviente {
	ravienteRegister := &RavienteRegister{
		nextTime:     0,
		startTime:    0,
		killedTime:   0,
		postTime:     0,
		ravienteType: 0,
		maxPlayers:   0,
		carveQuest:   0,
	}
	ravienteState := &RavienteState{}
	ravienteSupport := &RavienteSupport{}
	ravienteRegister.register = []uint32{0, 0, 0, 0, 0}
	ravienteState.stateData = []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}
	ravienteSupport.supportData = []uint32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

	raviente := &Raviente{
		register: ravienteRegister,
		state:    ravienteState,
		support:  ravienteSupport,
	}
	return raviente
}

// GetRaviMultiplier calculates the difficulty multiplier for Raviente based on player count.
//
// Raviente scales its difficulty based on the number of active participants. If there
// are fewer players than the minimum threshold, the encounter becomes easier by returning
// a multiplier < 1. Returns 1.0 for full groups, or 0 if the semaphore doesn't exist.
//
// Minimum player thresholds:
//   - Large Raviente (maxPlayers > 8): 24 players minimum
//   - Small Raviente (maxPlayers <= 8): 4 players minimum
func (r *Raviente) GetRaviMultiplier(s *Server) float64 {
	raviSema := getRaviSemaphore(s)
	if raviSema != nil {
		var minPlayers int
		if r.register.maxPlayers > 8 {
			minPlayers = 24
		} else {
			minPlayers = 4
		}
		if len(raviSema.clients) > minPlayers {
			return 1
		}
		return float64(minPlayers / len(raviSema.clients))
	}
	return 0
}

// NewServer creates and initializes a new channel server with the given configuration.
//
// The server is initialized with default persistent stages (town areas that always exist):
//   - sl1Ns200p0a0u0: Mezeporta (main town)
//   - sl1Ns211p0a0u0: Rasta bar
//   - sl1Ns260p0a0u0: Pallone Caravan
//   - sl1Ns262p0a0u0: Pallone Guest House 1st Floor
//   - sl1Ns263p0a0u0: Pallone Guest House 2nd Floor
//   - sl2Ns379p0a0u0: Diva fountain / prayer fountain
//   - sl1Ns462p0a0u0: MezFes (festival area)
//
// Additional dynamic stages are created by players when they create quests or rooms.
// The semaphore index starts at 7 to avoid reserved IDs 0-6.
func NewServer(config *Config) *Server {
	s := &Server{
		ID:              config.ID,
		logger:          config.Logger,
		db:              config.DB,
		erupeConfig:     config.ErupeConfig,
		acceptConns:     make(chan net.Conn),
		deleteConns:     make(chan net.Conn),
		sessions:        make(map[net.Conn]*Session),
		stages:          make(map[string]*Stage),
		userBinaryParts: make(map[userBinaryPartID][]byte),
		semaphore:       make(map[string]*Semaphore),
		semaphoreIndex:  7,
		discordBot:      config.DiscordBot,
		name:            config.Name,
		raviente:        NewRaviente(),
	}

	// Mezeporta
	s.stages["sl1Ns200p0a0u0"] = NewStage("sl1Ns200p0a0u0")

	// Rasta bar stage
	s.stages["sl1Ns211p0a0u0"] = NewStage("sl1Ns211p0a0u0")

	// Pallone Carvan
	s.stages["sl1Ns260p0a0u0"] = NewStage("sl1Ns260p0a0u0")

	// Pallone Guest House 1st Floor
	s.stages["sl1Ns262p0a0u0"] = NewStage("sl1Ns262p0a0u0")

	// Pallone Guest House 2nd Floor
	s.stages["sl1Ns263p0a0u0"] = NewStage("sl1Ns263p0a0u0")

	// Diva fountain / prayer fountain.
	s.stages["sl2Ns379p0a0u0"] = NewStage("sl2Ns379p0a0u0")

	// MezFes
	s.stages["sl1Ns462p0a0u0"] = NewStage("sl1Ns462p0a0u0")

	s.dict = getLangStrings(s)

	return s
}

// Start begins listening for connections and starts the server's main goroutines.
//
// This method:
//  1. Creates a TCP listener on the configured port
//  2. Launches acceptClients() goroutine to accept new connections
//  3. Launches manageSessions() goroutine to handle session lifecycle
//  4. Optionally starts Discord chat integration
//
// Returns an error if the listener cannot be created (e.g., port in use).
// The server runs asynchronously after Start() returns successfully.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}
	s.listener = l

	go s.acceptClients()
	go s.manageSessions()

	// Start the discord bot for chat integration.
	if s.erupeConfig.Discord.Enabled && s.discordBot != nil {
		s.discordBot.Session.AddHandler(s.onDiscordMessage)
	}

	return nil
}

// Shutdown gracefully stops the server and all its goroutines.
//
// This method:
//  1. Sets the shutdown flag to stop accepting new connections
//  2. Closes the TCP listener (causes acceptClients to exit)
//  3. Closes the acceptConns channel (signals manageSessions to exit)
//
// Existing sessions are not forcibly disconnected but will eventually timeout
// or disconnect naturally. For a complete shutdown, wait for all sessions to close.
func (s *Server) Shutdown() {
	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	s.listener.Close()

	close(s.acceptConns)
}

func (s *Server) acceptClients() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.Lock()
			shutdown := s.isShuttingDown
			s.Unlock()

			if shutdown {
				break
			} else {
				s.logger.Warn("Error accepting client", zap.Error(err))
				continue
			}
		}
		s.acceptConns <- conn
	}
}

func (s *Server) manageSessions() {
	for {
		select {
		case newConn := <-s.acceptConns:
			// Gracefully handle acceptConns channel closing.
			if newConn == nil {
				s.Lock()
				shutdown := s.isShuttingDown
				s.Unlock()

				if shutdown {
					return
				}
			}

			session := NewSession(s, newConn)

			s.Lock()
			s.sessions[newConn] = session
			s.Unlock()

			session.Start()

		case delConn := <-s.deleteConns:
			s.Lock()
			delete(s.sessions, delConn)
			s.Unlock()
		}
	}
}

// BroadcastMHF sends a packet to all active sessions on this channel server.
//
// The packet is built individually for each session to handle per-session state
// (like client version differences). Packets are queued in a non-blocking manner,
// so if a session's queue is full, the packet is dropped for that session only.
//
// Parameters:
//   - pkt: The MHFPacket to broadcast to all sessions
//   - ignoredSession: Optional session to exclude from the broadcast (typically the sender)
//
// Thread Safety: This method locks the server's session map during iteration.
func (s *Server) BroadcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session) {
	// Broadcast the data.
	s.Lock()
	defer s.Unlock()
	for _, session := range s.sessions {
		if session == ignoredSession {
			continue
		}

		// Make the header
		bf := byteframe.NewByteFrame()
		bf.WriteUint16(uint16(pkt.Opcode()))

		// Build the packet onto the byteframe.
		pkt.Build(bf, session.clientContext)

		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}

// WorldcastMHF broadcasts a packet to all channel servers (world-wide broadcast).
//
// This is used for server-wide announcements like Raviente events that should be
// visible to all players across all channels. The packet is sent to every channel
// server except the one specified in ignoredChannel.
//
// Parameters:
//   - pkt: The MHFPacket to broadcast across all channels
//   - ignoredSession: Optional session to exclude from broadcasts
//   - ignoredChannel: Optional channel server to skip (typically the originating channel)
func (s *Server) WorldcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session, ignoredChannel *Server) {
	for _, c := range s.Channels {
		if c == ignoredChannel {
			continue
		}
		c.BroadcastMHF(pkt, ignoredSession)
	}
}

// BroadcastChatMessage sends a simple chat message to all sessions on this server.
//
// The message appears as a system message with the server's configured name as the sender.
// This is typically used for server announcements, maintenance notifications, or events.
//
// Parameters:
//   - message: The text message to broadcast to all players
func (s *Server) BroadcastChatMessage(message string) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	msgBinChat := &binpacket.MsgBinChat{
		Unk0:       0,
		Type:       5,
		Flags:      0x80,
		Message:    message,
		SenderName: s.name,
	}
	msgBinChat.Build(bf)

	s.BroadcastMHF(&mhfpacket.MsgSysCastedBinary{
		CharID:         0xFFFFFFFF,
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}, nil)
}

func (s *Server) BroadcastRaviente(ip uint32, port uint16, stage []byte, _type uint8) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	bf.WriteUint16(0)    // Unk
	bf.WriteUint16(0x43) // Data len
	bf.WriteUint16(3)    // Unk len
	var text string
	switch _type {
	case 2:
		text = s.dict["ravienteBerserk"]
	case 3:
		text = s.dict["ravienteExtreme"]
	case 4:
		text = s.dict["ravienteExtremeLimited"]
	case 5:
		text = s.dict["ravienteBerserkSmall"]
	default:
		s.logger.Error("Unk raviente type", zap.Uint8("_type", _type))
	}
	ps.Uint16(bf, text, true)
	bf.WriteBytes([]byte{0x5F, 0x53, 0x00})
	bf.WriteUint32(ip)   // IP address
	bf.WriteUint16(port) // Port
	bf.WriteUint16(0)    // Unk
	bf.WriteBytes(stage)
	s.WorldcastMHF(&mhfpacket.MsgSysCastedBinary{
		CharID:         0x00000000,
		BroadcastType:  BroadcastTypeServer,
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}, nil, s)
}

func (s *Server) DiscordChannelSend(charName string, content string) {
	if s.erupeConfig.Discord.Enabled && s.discordBot != nil {
		message := fmt.Sprintf("**%s**: %s", charName, content)
		s.discordBot.RealtimeChannelSend(message)
	}
}

func (s *Server) FindSessionByCharID(charID uint32) *Session {
	for _, c := range s.Channels {
		for _, session := range c.sessions {
			if session.charID == charID {
				return session
			}
		}
	}
	return nil
}

func (s *Server) FindObjectByChar(charID uint32) *Object {
	s.stagesLock.RLock()
	defer s.stagesLock.RUnlock()
	for _, stage := range s.stages {
		stage.RLock()
		for objId := range stage.objects {
			obj := stage.objects[objId]
			if obj.ownerCharID == charID {
				stage.RUnlock()
				return obj
			}
		}
		stage.RUnlock()
	}

	return nil
}

func (s *Server) NextSemaphoreID() uint32 {
	for {
		exists := false
		s.semaphoreIndex = s.semaphoreIndex + 1
		if s.semaphoreIndex == 0 {
			s.semaphoreIndex = 7 // Skip reserved indexes
		}
		for _, semaphore := range s.semaphore {
			if semaphore.id == s.semaphoreIndex {
				exists = true
			}
		}
		if exists == false {
			break
		}
	}
	return s.semaphoreIndex
}
