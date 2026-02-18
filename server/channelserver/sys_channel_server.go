package channelserver

import (
	"fmt"
	"net"
	"sync"
	"time"

	"erupe-ce/common/byteframe"
	_config "erupe-ce/config"
	"erupe-ce/network/binpacket"
	"erupe-ce/network/mhfpacket"
	"erupe-ce/server/discordbot"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"
)

// Config struct allows configuring the server.
type Config struct {
	ID          uint16
	Logger      *zap.Logger
	DB          *sqlx.DB
	DiscordBot  *discordbot.DiscordBot
	ErupeConfig *_config.Config
	Name        string
	Enable      bool
}

// Map key type for a user binary part.
type userBinaryPartID struct {
	charID uint32
	index  uint8
}

// Server is a MHF channel server.
type Server struct {
	sync.Mutex
	Channels       []*Server
	ID             uint16
	GlobalID       string
	IP             string
	Port           uint16
	logger         *zap.Logger
	db             *sqlx.DB
	erupeConfig    *_config.Config
	acceptConns    chan net.Conn
	deleteConns    chan net.Conn
	sessions       map[net.Conn]*Session
	listener       net.Listener // Listener that is created when Server.Start is called.
	isShuttingDown bool

	stagesLock sync.RWMutex
	stages     map[string]*Stage

	// Used to map different languages
	i18n i18n

	// UserBinary
	userBinaryPartsLock sync.RWMutex
	userBinaryParts     map[userBinaryPartID][]byte

	// Semaphore
	semaphoreLock  sync.RWMutex
	semaphore      map[string]*Semaphore
	semaphoreIndex uint32

	// Discord chat integration
	discordBot *discordbot.DiscordBot

	name string

	raviente *Raviente

	questCacheLock sync.RWMutex
	questCacheData map[int][]byte
	questCacheTime map[int]time.Time
}

// NewServer creates a new Server type.
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
		raviente: &Raviente{
			id:       1,
			register: make([]uint32, 30),
			state:    make([]uint32, 30),
			support:  make([]uint32, 30),
		},
		questCacheData: make(map[int][]byte),
		questCacheTime: make(map[int]time.Time),
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

	s.i18n = getLangStrings(s)

	return s
}

// Start starts the server in a new goroutine.
func (s *Server) Start() error {
	l, err := net.Listen("tcp", fmt.Sprintf(":%d", s.Port))
	if err != nil {
		return err
	}
	s.listener = l

	go s.acceptClients()
	go s.manageSessions()
	go s.invalidateSessions()

	// Start the discord bot for chat integration.
	if s.erupeConfig.Discord.Enabled && s.discordBot != nil {
		s.discordBot.Session.AddHandler(s.onDiscordMessage)
		s.discordBot.Session.AddHandler(s.onInteraction)
	}

	return nil
}

// Shutdown tries to shut down the server gracefully.
func (s *Server) Shutdown() {
	s.Lock()
	s.isShuttingDown = true
	s.Unlock()

	if s.listener != nil {
		_ = s.listener.Close()
	}

	if s.acceptConns != nil {
		close(s.acceptConns)
	}
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

func (s *Server) getObjectId() uint16 {
	ids := make(map[uint16]struct{})
	for _, sess := range s.sessions {
		ids[sess.objectID] = struct{}{}
	}
	for i := uint16(1); i < 100; i++ {
		if _, ok := ids[i]; !ok {
			return i
		}
	}
	s.logger.Warn("object ids overflowed", zap.Int("sessions", len(s.sessions)))
	return 0
}

func (s *Server) invalidateSessions() {
	for !s.isShuttingDown {

		for _, sess := range s.sessions {
			if time.Since(sess.lastPacket) > time.Second*time.Duration(30) {
				s.logger.Info("session timeout", zap.String("Name", sess.Name))
				logoutPlayer(sess)
			}
		}
		time.Sleep(time.Second * 10)
	}
}

// BroadcastMHF queues a MHFPacket to be sent to all sessions.
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
		_ = pkt.Build(bf, session.clientContext)

		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}

// WorldcastMHF broadcasts a packet to all sessions across all channel servers.
func (s *Server) WorldcastMHF(pkt mhfpacket.MHFPacket, ignoredSession *Session, ignoredChannel *Server) {
	for _, c := range s.Channels {
		if c == ignoredChannel {
			continue
		}
		c.BroadcastMHF(pkt, ignoredSession)
	}
}

// BroadcastChatMessage broadcasts a simple chat message to all the sessions.
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
	_ = msgBinChat.Build(bf)

	s.BroadcastMHF(&mhfpacket.MsgSysCastedBinary{
		MessageType:    BinaryMessageTypeChat,
		RawDataPayload: bf.Data(),
	}, nil)
}

// DiscordChannelSend sends a chat message to the configured Discord channel.
func (s *Server) DiscordChannelSend(charName string, content string) {
	if s.erupeConfig.Discord.Enabled && s.discordBot != nil {
		message := fmt.Sprintf("**%s**: %s", charName, content)
		_ = s.discordBot.RealtimeChannelSend(message)
	}
}

// DiscordScreenShotSend sends a screenshot link to the configured Discord channel.
func (s *Server) DiscordScreenShotSend(charName string, title string, description string, articleToken string) {
	if s.erupeConfig.Discord.Enabled && s.discordBot != nil {
		imageUrl := fmt.Sprintf("%s:%d/api/ss/bbs/%s", s.erupeConfig.Screenshots.Host, s.erupeConfig.Screenshots.Port, articleToken)
		message := fmt.Sprintf("**%s**: %s - %s %s", charName, title, description, imageUrl)
		_ = s.discordBot.RealtimeChannelSend(message)
	}
}

// FindSessionByCharID looks up a session by character ID across all channels.
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

// DisconnectUser disconnects all sessions belonging to the given user ID.
func (s *Server) DisconnectUser(uid uint32) {
	var cid uint32
	var cids []uint32
	rows, err := s.db.Query(`SELECT id FROM characters WHERE user_id=$1`, uid)
	if err != nil {
		s.logger.Error("Failed to query characters for disconnect", zap.Error(err))
	} else {
		defer func() { _ = rows.Close() }()
		for rows.Next() {
			_ = rows.Scan(&cid)
			cids = append(cids, cid)
		}
	}
	for _, c := range s.Channels {
		for _, session := range c.sessions {
			for _, cid := range cids {
				if session.charID == cid {
					_ = session.rawConn.Close()
					break
				}
			}
		}
	}
}

// FindObjectByChar finds a stage object owned by the given character ID.
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

// HasSemaphore checks if the given session is hosting any semaphore.
func (s *Server) HasSemaphore(ses *Session) bool {
	for _, semaphore := range s.semaphore {
		if semaphore.host == ses {
			return true
		}
	}
	return false
}

// Season returns the current in-game season (0-2) based on server ID and time.
func (s *Server) Season() uint8 {
	sid := int64(((s.ID & 0xFF00) - 4096) / 256)
	return uint8(((TimeAdjusted().Unix() / 86400) + sid) % 3)
}
