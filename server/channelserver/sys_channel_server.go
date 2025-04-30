package channelserver

import (
	"fmt"
	"net"
	"strings"
	"sync"
	"time"

	"erupe-ce/common/byteframe"
	ps "erupe-ce/common/pascalstring"
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
	objectIDs      map[*Session]uint16
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

type Raviente struct {
	sync.Mutex
	id       uint16
	register []uint32
	state    []uint32
	support  []uint32
}

func (s *Server) resetRaviente() {
	for _, semaphore := range s.semaphore {
		if strings.HasPrefix(semaphore.name, "hs_l0") {
			return
		}
	}
	s.logger.Debug("All Raviente Semaphores empty, resetting")
	s.raviente.id = s.raviente.id + 1
	s.raviente.register = make([]uint32, 30)
	s.raviente.state = make([]uint32, 30)
	s.raviente.support = make([]uint32, 30)
}

func (s *Server) GetRaviMultiplier() float64 {
	raviSema := s.getRaviSemaphore()
	if raviSema != nil {
		var minPlayers int
		if s.raviente.register[9] > 8 {
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

func (s *Server) UpdateRavi(semaID uint32, index uint8, value uint32, update bool) (uint32, uint32) {
	var prev uint32
	var dest *[]uint32
	switch semaID {
	case 0x40000:
		switch index {
		case 17, 28: // Ignore res and poison
			break
		default:
			value = uint32(float64(value) * s.GetRaviMultiplier())
		}
		dest = &s.raviente.state
	case 0x50000:
		dest = &s.raviente.support
	case 0x60000:
		dest = &s.raviente.register
	default:
		return 0, 0
	}
	if update {
		(*dest)[index] += value
	} else {
		(*dest)[index] = value
	}
	return prev, (*dest)[index]
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
		objectIDs:       make(map[*Session]uint16),
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

func (s *Server) invalidateSessions() {
	for {
		if s.isShuttingDown {
			break
		}
		for _, sess := range s.sessions {
			if time.Now().Sub(sess.lastPacket) > time.Second*time.Duration(30) {
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
		pkt.Build(bf, session.clientContext)

		// Enqueue in a non-blocking way that drops the packet if the connections send buffer channel is full.
		session.QueueSendNonBlocking(bf.Data())
	}
}

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
	msgBinChat.Build(bf)

	s.BroadcastMHF(&mhfpacket.MsgSysCastedBinary{
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
		text = s.i18n.raviente.berserk
	case 3:
		text = s.i18n.raviente.extreme
	case 4:
		text = s.i18n.raviente.extremeLimited
	case 5:
		text = s.i18n.raviente.berserkSmall
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

func (s *Server) DiscordScreenShotSend(charName string, title string, description string, articleToken string) {
	if s.erupeConfig.Discord.Enabled && s.discordBot != nil {
		imageUrl := fmt.Sprintf("%s:%d/api/ss/bbs/%s", s.erupeConfig.Screenshots.Host, s.erupeConfig.Screenshots.Port, articleToken)
		message := fmt.Sprintf("**%s**: %s - %s %s", charName, title, description, imageUrl)
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

func (s *Server) DisconnectUser(uid uint32) {
	var cid uint32
	var cids []uint32
	rows, _ := s.db.Query(`SELECT id FROM characters WHERE user_id=$1`, uid)
	for rows.Next() {
		rows.Scan(&cid)
		cids = append(cids, cid)
	}
	for _, c := range s.Channels {
		for _, session := range c.sessions {
			for _, cid := range cids {
				if session.charID == cid {
					session.rawConn.Close()
					break
				}
			}
		}
	}
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

func (s *Server) HasSemaphore(ses *Session) bool {
	for _, semaphore := range s.semaphore {
		if semaphore.host == ses {
			return true
		}
	}
	return false
}

func (s *Server) Season() uint8 {
	sid := int64(((s.ID & 0xFF00) - 4096) / 256)
	return uint8(((TimeAdjusted().Unix() / 86400) + sid) % 3)
}
