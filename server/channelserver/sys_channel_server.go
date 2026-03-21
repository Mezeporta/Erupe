package channelserver

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"time"

	"erupe-ce/common/byteframe"
	"erupe-ce/common/decryption"
	cfg "erupe-ce/config"
	"erupe-ce/network"
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
	ErupeConfig *cfg.Config
	Name        string
	Enable      bool
}

// Server is a MHF channel server.
//
// Lock ordering (acquire in this order to avoid deadlocks):
//  1. Server.Mutex          – protects sessions map
//  2. Stage.RWMutex         – protects per-stage state (clients, objects)
//  3. Server.semaphoreLock  – protects semaphore map
//
// Note: Server.stages is a StageMap (sync.Map-backed), so it requires no
// external lock for reads or writes.
//
// Self-contained stores (userBinary, minidata, questCache) manage their
// own locks internally and may be acquired at any point.
type Server struct {
	sync.Mutex
	Registry           ChannelRegistry
	ID                 uint16
	GlobalID           string
	IP                 string
	Port               uint16
	logger             *zap.Logger
	db                 *sqlx.DB
	charRepo           CharacterRepo
	guildRepo          GuildRepo
	userRepo           UserRepo
	gachaRepo          GachaRepo
	houseRepo          HouseRepo
	festaRepo          FestaRepo
	towerRepo          TowerRepo
	rengokuRepo        RengokuRepo
	mailRepo           MailRepo
	stampRepo          StampRepo
	distRepo           DistributionRepo
	sessionRepo        SessionRepo
	eventRepo          EventRepo
	achievementRepo    AchievementRepo
	shopRepo           ShopRepo
	cafeRepo           CafeRepo
	goocooRepo         GoocooRepo
	divaRepo           DivaRepo
	miscRepo           MiscRepo
	scenarioRepo       ScenarioRepo
	mercenaryRepo      MercenaryRepo
	mailService        *MailService
	guildService       *GuildService
	achievementService *AchievementService
	gachaService       *GachaService
	towerService       *TowerService
	festaService       *FestaService
	erupeConfig        *cfg.Config
	acceptConns        chan net.Conn
	deleteConns        chan net.Conn
	sessions           map[net.Conn]*Session
	listener           net.Listener // Listener that is created when Server.Start is called.
	isShuttingDown     bool
	done               chan struct{} // Closed on Shutdown to wake background goroutines.

	stages StageMap

	// Used to map different languages
	i18n i18n

	userBinary *UserBinaryStore
	minidata   *MinidataStore

	// Per-character save locks prevent concurrent save operations for the
	// same character from racing and defeating corruption detection.
	charSaveLocks CharacterLocks

	// Semaphore
	semaphoreLock  sync.RWMutex
	semaphore      map[string]*Semaphore
	semaphoreIndex uint32

	// Discord chat integration
	discordBot *discordbot.DiscordBot

	name string

	raviente *Raviente

	questCache *QuestCache

	rengokuBin []byte // Cached rengoku_data.bin (ECD-encrypted, served to clients as-is)

	handlerTable map[network.PacketID]handlerFunc
}

// NewServer creates a new Server type.
func NewServer(config *Config) *Server {
	s := &Server{
		ID:             config.ID,
		logger:         config.Logger,
		db:             config.DB,
		erupeConfig:    config.ErupeConfig,
		acceptConns:    make(chan net.Conn),
		deleteConns:    make(chan net.Conn),
		done:           make(chan struct{}),
		sessions:       make(map[net.Conn]*Session),
		userBinary:     NewUserBinaryStore(),
		minidata:       NewMinidataStore(),
		semaphore:      make(map[string]*Semaphore),
		semaphoreIndex: 7,
		discordBot:     config.DiscordBot,
		name:           config.Name,
		raviente: &Raviente{
			id:       1,
			register: make([]uint32, 30),
			state:    make([]uint32, 30),
			support:  make([]uint32, 30),
		},
		questCache:   NewQuestCache(config.ErupeConfig.QuestCacheExpiry),
		handlerTable: buildHandlerTable(),
	}

	s.charRepo = NewCharacterRepository(config.DB)
	s.guildRepo = NewGuildRepository(config.DB)
	s.userRepo = NewUserRepository(config.DB)
	s.gachaRepo = NewGachaRepository(config.DB)
	s.houseRepo = NewHouseRepository(config.DB)
	s.festaRepo = NewFestaRepository(config.DB)
	s.towerRepo = NewTowerRepository(config.DB)
	s.rengokuRepo = NewRengokuRepository(config.DB)
	s.mailRepo = NewMailRepository(config.DB)
	s.stampRepo = NewStampRepository(config.DB)
	s.distRepo = NewDistributionRepository(config.DB)
	s.sessionRepo = NewSessionRepository(config.DB)
	s.eventRepo = NewEventRepository(config.DB)
	s.achievementRepo = NewAchievementRepository(config.DB)
	s.shopRepo = NewShopRepository(config.DB)
	s.cafeRepo = NewCafeRepository(config.DB)
	s.goocooRepo = NewGoocooRepository(config.DB)
	s.divaRepo = NewDivaRepository(config.DB)
	s.miscRepo = NewMiscRepository(config.DB)
	s.scenarioRepo = NewScenarioRepository(config.DB)
	s.mercenaryRepo = NewMercenaryRepository(config.DB)

	s.mailService = NewMailService(s.mailRepo, s.guildRepo, s.logger)
	s.guildService = NewGuildService(s.guildRepo, s.mailService, s.charRepo, s.logger)
	s.achievementService = NewAchievementService(s.achievementRepo, s.logger)
	s.gachaService = NewGachaService(s.gachaRepo, s.userRepo, s.charRepo, s.logger, config.ErupeConfig.GameplayOptions.MaximumNP)
	s.towerService = NewTowerService(s.towerRepo, s.logger)
	s.festaService = NewFestaService(s.festaRepo, s.logger)

	// Mezeporta
	s.stages.Store("sl1Ns200p0a0u0", NewStage("sl1Ns200p0a0u0"))

	// Rasta bar stage
	s.stages.Store("sl1Ns211p0a0u0", NewStage("sl1Ns211p0a0u0"))

	// Pallone Carvan
	s.stages.Store("sl1Ns260p0a0u0", NewStage("sl1Ns260p0a0u0"))

	// Pallone Guest House 1st Floor
	s.stages.Store("sl1Ns262p0a0u0", NewStage("sl1Ns262p0a0u0"))

	// Pallone Guest House 2nd Floor
	s.stages.Store("sl1Ns263p0a0u0", NewStage("sl1Ns263p0a0u0"))

	// Diva fountain / prayer fountain.
	s.stages.Store("sl2Ns379p0a0u0", NewStage("sl2Ns379p0a0u0"))

	// MezFes
	s.stages.Store("sl1Ns462p0a0u0", NewStage("sl1Ns462p0a0u0"))

	s.rengokuBin = loadRengokuBinary(config.ErupeConfig.BinPath, s.logger)

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

	initCommands(s.erupeConfig.Commands, s.logger)

	go s.acceptClients()
	go s.manageSessions()
	go s.invalidateSessions()

	// Start the discord bot for chat integration.
	if s.erupeConfig.Discord.Enabled && s.discordBot != nil {
		s.discordBot.AddHandler(s.onDiscordMessage)
		s.discordBot.AddHandler(s.onInteraction)
	}

	return nil
}

// Shutdown tries to shut down the server gracefully. Safe to call multiple times.
func (s *Server) Shutdown() {
	s.Lock()
	alreadyShutDown := s.isShuttingDown
	s.isShuttingDown = true
	s.Unlock()

	if alreadyShutDown {
		return
	}

	close(s.done)

	if s.listener != nil {
		_ = s.listener.Close()
	}

}

// ShutdownAndDrain stops accepting new connections, force-closes every active
// session so that their logoutPlayer cleanup runs (saves character data, removes
// from stages, etc.), then waits until all sessions have been removed from the
// sessions map or ctx is cancelled.  It is safe to call multiple times.
func (s *Server) ShutdownAndDrain(ctx context.Context) {
	s.Shutdown()

	// Snapshot all active connections while holding the lock, then close them
	// outside the lock so we don't hold it during I/O.  Closing a connection
	// causes the session's recvLoop to see io.EOF and call logoutPlayer(), which
	// in turn deletes the entry from s.sessions under the server mutex.
	s.Lock()
	conns := make([]net.Conn, 0, len(s.sessions))
	for conn := range s.sessions {
		conns = append(conns, conn)
	}
	s.Unlock()

	for _, conn := range conns {
		_ = conn.Close()
	}

	// Poll until logoutPlayer has removed every session or the deadline passes.
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			s.Lock()
			remaining := len(s.sessions)
			s.Unlock()
			s.logger.Warn("Shutdown drain timed out", zap.Int("remaining_sessions", remaining))
			return
		case <-ticker.C:
			s.Lock()
			n := len(s.sessions)
			s.Unlock()
			if n == 0 {
				s.logger.Info("Shutdown drain complete")
				return
			}
		}
	}
}

func (s *Server) acceptClients() {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			s.Lock()
			shutdown := s.isShuttingDown
			s.Unlock()

			if shutdown || errors.Is(err, net.ErrClosed) {
				break
			} else {
				s.logger.Warn("Error accepting client", zap.Error(err))
				continue
			}
		}
		select {
		case s.acceptConns <- conn:
		case <-s.done:
			_ = conn.Close()
			return
		}
	}
}

func (s *Server) manageSessions() {
	for {
		select {
		case <-s.done:
			return
		case newConn := <-s.acceptConns:
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
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-s.done:
			return
		case <-ticker.C:
		}

		s.Lock()
		var timedOut []*Session
		for _, sess := range s.sessions {
			if time.Since(sess.lastPacket) > time.Second*time.Duration(30) {
				timedOut = append(timedOut, sess)
			}
		}
		s.Unlock()

		for _, sess := range timedOut {
			s.logger.Info("session timeout", zap.String("Name", sess.Name))
			logoutPlayer(sess)
		}
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
	s.Registry.Worldcast(pkt, ignoredSession, ignoredChannel)
}

// BroadcastChatMessage broadcasts a simple chat message to all the sessions.
func (s *Server) BroadcastChatMessage(message string) {
	bf := byteframe.NewByteFrame()
	bf.SetLE()
	msgBinChat := &binpacket.MsgBinChat{
		Unk0:       0,
		Type:       5,
		Flags:      chatFlagServer,
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
	return s.Registry.FindSessionByCharID(charID)
}

// DisconnectUser disconnects all sessions belonging to the given user ID.
func (s *Server) DisconnectUser(uid uint32) {
	cids, err := s.charRepo.GetCharIDsByUserID(uid)
	if err != nil {
		s.logger.Error("Failed to query characters for disconnect", zap.Error(err))
	}
	s.Registry.DisconnectUser(cids)
}

// FindObjectByChar finds a stage object owned by the given character ID.
func (s *Server) FindObjectByChar(charID uint32) *Object {
	var found *Object
	s.stages.Range(func(_ string, stage *Stage) bool {
		stage.RLock()
		for _, obj := range stage.objects {
			if obj.ownerCharID == charID {
				found = obj
				stage.RUnlock()
				return false // stop iteration
			}
		}
		stage.RUnlock()
		return true
	})
	return found
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

// Server ID arithmetic constants
const (
	serverIDHighMask = uint16(0xFF00)
	serverIDBase     = 0x1000 // first server ID offset
	serverIDStride   = 0x100  // spacing between server IDs
)

// Season returns the current in-game season (0-2) based on server ID and time.
func (s *Server) Season() uint8 {
	sid := int64(((s.ID & serverIDHighMask) - serverIDBase) / serverIDStride)
	return uint8(((TimeAdjusted().Unix() / secsPerDay) + sid) % 3)
}

// loadRengokuBinary loads and caches Hunting Road config. It tries
// rengoku_data.bin first and falls back to rengoku_data.json (built on the
// fly). Returns ECD-encrypted bytes ready to serve, or nil if no valid source
// is found.
func loadRengokuBinary(binPath string, logger *zap.Logger) []byte {
	path := filepath.Join(binPath, "rengoku_data.bin")
	data, err := os.ReadFile(path)
	if err == nil {
		if len(data) < 4 {
			logger.Warn("rengoku_data.bin too small, ignoring",
				zap.Int("bytes", len(data)))
		} else if magic := binary.LittleEndian.Uint32(data[:4]); magic != decryption.ECDMagic {
			logger.Warn("rengoku_data.bin has invalid ECD magic, ignoring",
				zap.String("expected", fmt.Sprintf("0x%08x", decryption.ECDMagic)),
				zap.String("got", fmt.Sprintf("0x%08x", magic)))
		} else {
			// Decrypt and decompress to validate the internal structure and emit a
			// human-readable summary at startup. Failures here are non-fatal: the
			// encrypted blob is still served to clients unchanged.
			if plain, decErr := decryption.DecodeECD(data); decErr != nil {
				logger.Warn("rengoku_data.bin ECD decryption failed — serving anyway",
					zap.Error(decErr))
			} else {
				raw := decryption.UnpackSimple(plain)
				if info, parseErr := parseRengokuBinary(raw); parseErr != nil {
					logger.Warn("rengoku_data.bin structural validation failed",
						zap.Error(parseErr))
				} else {
					logger.Info("Hunting Road config",
						zap.Int("multi_floors", info.MultiFloors),
						zap.Int("multi_spawn_tables", info.MultiSpawnTables),
						zap.Int("solo_floors", info.SoloFloors),
						zap.Int("solo_spawn_tables", info.SoloSpawnTables),
						zap.Int("unique_monsters", info.UniqueMonsters),
					)
				}
			}
			logger.Info("Loaded rengoku_data.bin", zap.Int("bytes", len(data)))
			return data
		}
	}

	if enc := loadRengokuFromJSON(binPath, logger); enc != nil {
		return enc
	}

	logger.Warn("No Hunting Road config found (rengoku_data.bin or rengoku_data.json), Hunting Road will be unavailable")
	return nil
}
